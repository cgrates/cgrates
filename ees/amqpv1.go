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
	"sync"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAMQPv1EE creates a poster for amqpv1
func NewAMQPv1EE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *AMQPv1EE {
	pstr := &AMQPv1EE{
		cfg:     cfg,
		dc:      dc,
		queueID: "/" + utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	if vals, has := cfg.Opts[utils.AMQPQueueID]; has {
		pstr.queueID = "/" + utils.IfaceAsString(vals)
	}
	return pstr
}

// AMQPv1EE a poster for amqpv1
type AMQPv1EE struct {
	queueID string // identifier of the CDR queue where we publish
	client  *amqpv1.Client
	session *amqpv1.Session

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *AMQPv1EE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *AMQPv1EE) Connect() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.client == nil {
		if pstr.client, err = amqpv1.Dial(pstr.Cfg().ExportPath); err != nil {
			return
		}
	}
	if pstr.session == nil {
		pstr.session, err = pstr.client.NewSession()
		if err != nil {
			// reset client and try again
			// used in case of closed connection because of idle time
			if pstr.client != nil {
				pstr.client.Close() // Make shure the connection is closed before reseting it
				pstr.client = nil
			}
		}
	}
	return
}

func (pstr *AMQPv1EE) ExportEvent(ctx *context.Context, content interface{}, _ string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	defer func() {
		pstr.RUnlock()
		pstr.reqs.done()
	}()
	if pstr.session == nil {
		return utils.ErrDisconnected
	}
	sender, err := pstr.session.NewSender(
		amqpv1.LinkTargetAddress(pstr.queueID),
	)
	if err != nil {
		return
	}
	// Send message
	err = sender.Send(ctx, amqpv1.NewMessage(content.([]byte)))
	sender.Close(ctx)
	return
}

func (pstr *AMQPv1EE) Close() (err error) {
	pstr.Lock()
	if pstr.session != nil {
		pstr.session.Close(context.Background())
		pstr.session = nil
	}
	if pstr.client != nil {
		err = pstr.client.Close()
		pstr.client = nil
	}
	pstr.Unlock()
	return
}

func (pstr *AMQPv1EE) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
