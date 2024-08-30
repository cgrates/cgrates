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

package ers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NewAMQPER return a new amqp event reader
func NewAMQPER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	rdr := &AMQPER{
		cgrCfg:        cfg,
		cfgIdx:        cfgIdx,
		fltrS:         fltrS,
		rdrEvents:     rdrEvents,
		partialEvents: partialEvents,
		rdrExit:       rdrExit,
		rdrErr:        rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	rdr.dialURL = rdr.Config().SourcePath
	rdr.setOpts(rdr.Config().Opts)
	return rdr, nil
}

// AMQPER implements EventReader interface for amqp message
type AMQPER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	dialURL      string
	queueID      string
	tag          string
	exchange     string
	exchangeType string
	routingKey   string

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}

	conn    *amqp.Connection
	channel *amqp.Channel
}

// Config returns the curent configuration
func (rdr *AMQPER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the amqp topic
func (rdr *AMQPER) Serve() (err error) {
	if rdr.conn, err = amqp.Dial(rdr.dialURL); err != nil {
		return
	}
	if rdr.channel, err = rdr.conn.Channel(); err != nil {
		rdr.close()
		return
	}
	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}

	if rdr.exchange != "" {
		if err = rdr.channel.ExchangeDeclare(
			rdr.exchange,     // name
			rdr.exchangeType, // type
			true,             // durable
			false,            // audo-delete
			false,            // internal
			false,            // no-wait
			nil,              // args
		); err != nil {
			return
		}
	}

	if _, err = rdr.channel.QueueDeclare(
		rdr.queueID, // name
		true,        // durable
		false,       // auto-delete
		false,       // exclusive
		false,       // no-wait
		nil,         // args
	); err != nil {
		return
	}

	if rdr.exchange != "" {
		if err = rdr.channel.QueueBind(
			rdr.queueID,    // queue
			rdr.routingKey, // key
			rdr.exchange,   // exchange
			false,          // no-wait
			nil,            // args
		); err != nil {
			return
		}
	}

	var msgChan <-chan amqp.Delivery
	if msgChan, err = rdr.channel.Consume(rdr.queueID, rdr.tag,
		false, false, false, true, nil); err != nil {
		return
	}
	go rdr.readLoop(msgChan) // read until the connection is closed
	return
}

func (rdr *AMQPER) readLoop(msgChan <-chan amqp.Delivery) {
	for {
		if rdr.Config().ConcurrentReqs != -1 {
			<-rdr.cap // do not try to read if the limit is reached
		}
		select {
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring amqp path <%s>",
					utils.ERs, rdr.dialURL))
			rdr.close()
			return
		case msg, ok := <-msgChan:
			if !ok {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> lost connection to AMQP server at %s, closing reader...",
						utils.ERs, rdr.dialURL))
				rdr.close()
				return
			}
			go func(msg amqp.Delivery) {
				err := rdr.processMessage(msg.Body)
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, msg.MessageId, err.Error()))

					err = msg.Reject(true)
					if err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> error negatively acknowledging message %s: %s",
								utils.ERs, msg.MessageId, err.Error()))
					}
					return
				}

				err = msg.Ack(false)
				if err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> error acknowledging message %s: %s",
							utils.ERs, msg.MessageId, err.Error()))
				}
				if rdr.Config().ConcurrentReqs != -1 {
					rdr.cap <- struct{}{}
				}
			}(msg)
		}
	}
}

func (rdr *AMQPER) processMessage(msg []byte) (err error) {
	var decodedMessage map[string]any
	if err = json.Unmarshal(msg, &decodedMessage); err != nil {
		return
	}
	reqVars := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{utils.MetaReaderID: utils.NewLeafNode(rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx].ID)}}
	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), reqVars,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(context.TODO(), agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep, agReq.Opts)
	rdrEv := rdr.rdrEvents
	if _, isPartial := cgrEv.APIOpts[utils.PartialOpt]; isPartial {
		rdrEv = rdr.partialEvents
	}
	rdrEv <- &erEvent{
		cgrEvent: cgrEv,
		rdrCfg:   rdr.Config(),
	}
	return
}

func (rdr *AMQPER) setOpts(opts *config.EventReaderOpts) {
	rdr.queueID = utils.DefaultQueueID
	if opts.AMQPQueueID != nil {
		rdr.queueID = *opts.AMQPQueueID
	}
	rdr.tag = utils.AMQPDefaultConsumerTag
	if opts.AMQPConsumerTag != nil {
		rdr.tag = *opts.AMQPConsumerTag
	}
	if opts.AMQPRoutingKey != nil {
		rdr.routingKey = *opts.AMQPRoutingKey
	}
	if opts.AMQPExchange != nil {
		rdr.exchange = *opts.AMQPExchange
		rdr.exchangeType = utils.DefaultExchangeType
	}
	if opts.AMQPExchangeType != nil {
		rdr.exchangeType = *opts.AMQPExchangeType
	}
}

func (rdr *AMQPER) close() (err error) {
	if rdr.channel != nil {
		if err = rdr.channel.Cancel(rdr.tag, true); err != nil {
			return
		}
		if err = rdr.channel.Close(); err != nil {
			return
		}
	}
	return rdr.conn.Close()
}
