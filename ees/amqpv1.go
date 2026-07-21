// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"sync"

	amqpv1 "github.com/Azure/go-amqp"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// NewAMQPv1EE creates a poster for amqpv1
func NewAMQPv1EE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) *AMQPv1EE {
	pstr := &AMQPv1EE{
		cfg:     cfg,
		em:      em,
		queueID: "/" + utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	if cfg.Opts.AMQPQueueID != nil {
		pstr.queueID = "/" + *cfg.Opts.AMQPQueueID
	}
	if cfg.Opts.AMQPUsername != nil && cfg.Opts.AMQPPassword != nil {
		pstr.connOpts = &amqpv1.ConnOptions{
			SASLType: amqpv1.SASLTypePlain(*cfg.Opts.AMQPUsername, *cfg.Opts.AMQPPassword),
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
	em           *utils.ExporterMetrics
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

func (pstr *AMQPv1EE) ExportEvent(ctx *context.Context, content, _ any) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	defer func() {
		pstr.RUnlock()
		pstr.reqs.done()
	}()
	if pstr.session == nil {
		return utils.ErrDisconnected
	}
	sender, err := pstr.session.NewSender(ctx, pstr.queueID, nil)
	if err != nil {
		return
	}
	// Send message
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

func (pstr *AMQPv1EE) GetMetrics() *utils.ExporterMetrics { return pstr.em }

func (pstr *AMQPv1EE) ExtraData(*utils.CGREvent) any { return nil }
