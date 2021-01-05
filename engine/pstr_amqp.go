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

package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

// NewAMQPPoster creates a new amqp poster
// "amqp://guest:guest@localhost:5672/?queueID=cgrates_cdrs"
func NewAMQPPoster(dialURL string, attempts int, opts map[string]interface{}) *AMQPPoster {
	amqp := &AMQPPoster{
		attempts: attempts,
		dialURL:  dialURL,
	}
	amqp.parseOpts(opts)
	return amqp
}

// AMQPPoster used to post cdrs to amqp
type AMQPPoster struct {
	dialURL      string
	queueID      string // identifier of the CDR queue where we publish
	exchange     string
	exchangeType string
	routingKey   string
	attempts     int
	sync.Mutex   // protect connection
	conn         *amqp.Connection
}

func (pstr *AMQPPoster) parseOpts(dialURL map[string]interface{}) {
	pstr.queueID = utils.DefaultQueueID
	pstr.routingKey = utils.DefaultQueueID
	if vals, has := dialURL[utils.QueueID]; has {
		pstr.queueID = utils.IfaceAsString(vals)
	}
	if vals, has := dialURL[utils.RoutingKey]; has {
		pstr.routingKey = utils.IfaceAsString(vals)
	}
	if vals, has := dialURL[utils.Exchange]; has {
		pstr.exchange = utils.IfaceAsString(vals)
		pstr.exchangeType = utils.DefaultExchangeType
	}
	if vals, has := dialURL[utils.ExchangeType]; has {
		pstr.exchangeType = utils.IfaceAsString(vals)
	}
	return
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *AMQPPoster) Post(content []byte, _ string) (err error) {
	var chn *amqp.Channel
	fib := utils.Fib()

	for i := 0; i < pstr.attempts; i++ {
		if chn, err = pstr.newPostChannel(); err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<AMQPPoster> creating new post channel, err: %s", err.Error()))
		return
	}
	for i := 0; i < pstr.attempts; i++ {
		if err = chn.Publish(
			pstr.exchange,   // exchange
			pstr.routingKey, // routing key
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  utils.ContentJSON,
				Body:         content,
			}); err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		return
	}
	if chn != nil {
		chn.Close()
	}
	return
}

// Close closes the connections
func (pstr *AMQPPoster) Close() {
	pstr.Lock()
	if pstr.conn != nil {
		pstr.conn.Close()
	}
	pstr.conn = nil
	pstr.Unlock()
}

func (pstr *AMQPPoster) newPostChannel() (postChan *amqp.Channel, err error) {
	pstr.Lock()
	if pstr.conn == nil {
		var conn *amqp.Connection
		conn, err = amqp.Dial(pstr.dialURL)
		if err == nil {
			pstr.conn = conn
			go func() { // monitor connection errors so we can restart
				if err := <-pstr.conn.NotifyClose(make(chan *amqp.Error)); err != nil {
					utils.Logger.Err(fmt.Sprintf("Connection error received: %s", err.Error()))
					pstr.Close()
				}
			}()
		}
	}
	pstr.Unlock()
	if err != nil {
		return nil, err
	}
	if postChan, err = pstr.conn.Channel(); err != nil {
		return
	}

	if pstr.exchange != "" {
		if err = postChan.ExchangeDeclare(
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

	if _, err = postChan.QueueDeclare(
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
		if err = postChan.QueueBind(
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
