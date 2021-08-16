/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package ees

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

// NewAMQPee creates a new amqp poster
// "amqp://guest:guest@localhost:5672/?queueID=cgrates_cdrs"
func NewAMQPee(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *AMQPee {
	amqp := &AMQPee{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}
	amqp.parseOpts(cfg.Opts)
	return amqp
}

// AMQPee used to post cdrs to amqp
type AMQPee struct {
	queueID      string // identifier of the CDR queue where we publish
	exchange     string
	exchangeType string
	routingKey   string
	conn         *amqp.Connection
	postChan     *amqp.Channel

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *AMQPee) parseOpts(dialURL map[string]interface{}) {
	pstr.queueID = utils.DefaultQueueID
	pstr.routingKey = utils.DefaultQueueID
	if vals, has := dialURL[utils.AMQPQueueID]; has {
		pstr.queueID = utils.IfaceAsString(vals)
	}
	if vals, has := dialURL[utils.AMQPRoutingKey]; has {
		pstr.routingKey = utils.IfaceAsString(vals)
	}
	if vals, has := dialURL[utils.AMQPExchange]; has {
		pstr.exchange = utils.IfaceAsString(vals)
		pstr.exchangeType = utils.DefaultExchangeType
	}
	if vals, has := dialURL[utils.AMQPExchangeType]; has {
		pstr.exchangeType = utils.IfaceAsString(vals)
	}
}

func (pstr *AMQPee) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *AMQPee) Connect() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.conn == nil {
		if pstr.conn, err = amqp.Dial(pstr.Cfg().ExportPath); err != nil {
			return
		}
		go func() { // monitor connection errors so we can restart
			if err := <-pstr.conn.NotifyClose(make(chan *amqp.Error)); err != nil {
				utils.Logger.Err(fmt.Sprintf("Connection error received: %s", err.Error()))
				pstr.Close()
			}
		}()
	}
	if pstr.postChan != nil {
		return
	}

	if pstr.postChan, err = pstr.conn.Channel(); err != nil {
		return
	}

	if pstr.exchange != "" {
		if err = pstr.postChan.ExchangeDeclare(
			pstr.exchange,     // name
			pstr.exchangeType, // type
			true,              // durable
			false,             // audo-delete
			false,             // internal
			false,             // no-wait
			nil,               // args
		); err != nil {
			return
		}
	}

	if _, err = pstr.postChan.QueueDeclare(
		pstr.queueID, // name
		true,         // durable
		false,        // auto-delete
		false,        // exclusive
		false,        // no-wait
		nil,          // args
	); err != nil {
		return
	}

	if pstr.exchange != "" {
		if err = pstr.postChan.QueueBind(
			pstr.queueID,    // queue
			pstr.routingKey, // key
			pstr.exchange,   // exchange
			false,           // no-wait
			nil,             // args
		); err != nil {
			return
		}
	}
	return
}

func (pstr *AMQPee) ExportEvent(content interface{}, _ string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	err = pstr.postChan.Publish(
		pstr.exchange,   // exchange
		pstr.routingKey, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  utils.ContentJSON,
			Body:         content.([]byte),
		})
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *AMQPee) Close() (err error) {
	pstr.Lock()
	if pstr.postChan != nil {
		pstr.postChan.Close()
		pstr.postChan = nil
	}
	if pstr.conn != nil {
		err = pstr.conn.Close()
		pstr.conn = nil
	}
	pstr.Unlock()
	return
}

func (pstr *AMQPee) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
