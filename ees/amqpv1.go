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
	"context"
	"sync"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAMQPv1EE creates a poster for amqpv1
func NewAMQPv1EE(cfg *config.EventExporterCfg, dc *utils.ExporterMetrics) *AMQPv1EE {
	pstr := &AMQPv1EE{
		cfg:     cfg,
		dc:      dc,
		queueID: "/" + utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	if amqp := cfg.Opts.AMQP; amqp != nil {
		if amqp.QueueID != nil {
			pstr.queueID = "/" + *amqp.QueueID
		}
		if amqp.Username != nil && amqp.Password != nil {
			pstr.connOpts = &amqpv1.ConnOptions{
				SASLType: amqpv1.SASLTypePlain(*amqp.Username, *amqp.Password),
			}
		}
	}
	return pstr
}

// AMQPv1EE a poster for amqpv1
type AMQPv1EE struct {
	queueID  string // identifier of the CDR queue where we publish
	conn     *amqpv1.Conn
	connOpts *amqpv1.ConnOptions
	session  *amqpv1.Session

	cfg          *config.EventExporterCfg
	dc           *utils.ExporterMetrics
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *AMQPv1EE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *AMQPv1EE) Connect() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.conn == nil {
		if pstr.conn, err = amqpv1.Dial(context.TODO(), pstr.Cfg().ExportPath, pstr.connOpts); err != nil {
			return
		}
	}
	if pstr.session == nil {
		pstr.session, err = pstr.conn.NewSession(context.TODO(), nil)
		if err != nil {
			// reset client and try again
			// used in case of closed connection because of idle time
			if pstr.conn != nil {
				pstr.conn.Close() // Make shure the connection is closed before reseting it
				pstr.conn = nil
			}
		}
	}
	return
}

func (pstr *AMQPv1EE) ExportEvent(content any, _ string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	defer func() {
		pstr.RUnlock()
		pstr.reqs.done()
	}()
	if pstr.session == nil {
		return utils.ErrDisconnected
	}
	sender, err := pstr.session.NewSender(context.TODO(), pstr.queueID, nil)
	if err != nil {
		return
	}
	// Send message
	ctx := context.Background()
	err = sender.Send(ctx, amqpv1.NewMessage(content.([]byte)), nil)
	sender.Close(ctx)
	return
}

func (pstr *AMQPv1EE) Close() (err error) {
	pstr.Lock()
	if pstr.session != nil {
		pstr.session.Close(context.Background())
		pstr.session = nil
	}
	if pstr.conn != nil {
		err = pstr.conn.Close()
		pstr.conn = nil
	}
	pstr.Unlock()
	return
}

func (pstr *AMQPv1EE) GetMetrics() *utils.ExporterMetrics { return pstr.dc }
