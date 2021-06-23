/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

// NewNatsER return a new amqp event reader
func NewNatsER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {
	rdr := &NatsER{
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
	return rdr, nil
}

// NatsER implements EventReader interface for amqp message
type NatsER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	dialURL string
	queueID string
	subject string

	jetStream bool

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}

	poster   *nats.Conn
	posterJS nats.JetStream

	posterSubject string
}

// Config returns the curent configuration
func (rdr *NatsER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the nats subject
func (rdr *NatsER) Serve() (err error) {
	// Connect to a server
	var nc *nats.Conn
	if nc, err = nats.Connect(rdr.dialURL); err != nil {
		return
	}

	ch := make(chan *nats.Msg)
	if !rdr.jetStream {
		if _, err = nc.ChanQueueSubscribe(rdr.subject, rdr.queueID, ch); err != nil {
			return
		}
	} else {
		// Create JetStream Context
		var js nats.JetStreamContext
		if js, err = nc.JetStream(); err != nil {
			return
		}
		if _, err = js.QueueSubscribe(rdr.subject, rdr.queueID, func(msg *nats.Msg) {
			ch <- msg
		}); err != nil {
			return
		}
	}
	for {
		if rdr.Config().ConcurrentReqs != -1 {
			<-rdr.cap // do not try to read if the limit is reached
		}
		select {
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring nats path <%s>",
					utils.ERs, rdr.dialURL))
			nc.Drain()
			return
		case msg := <-ch:
			if err = rdr.processMessage(msg.Data); err != nil {
				nc.Drain() // ignore this error(if any) in favor of the error processMessage
				return
			}

			go func(msg *nats.Msg) {
				if err := rdr.processMessage(msg.Data); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, string(msg.Data), err.Error()))
				}
				if rdr.poster != nil { // post it
					if err := rdr.postMessage(msg.Data); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> writing message %s error: %s",
								utils.ERs, string(msg.Data), err.Error()))
					}
				}
				if rdr.Config().ConcurrentReqs != -1 {
					rdr.cap <- struct{}{}
				}
			}(msg)
		}
	}
}

func (rdr *NatsER) processMessage(msg []byte) (err error) {
	var decodedMessage map[string]interface{}
	if err = json.Unmarshal(msg, &decodedMessage); err != nil {
		return
	}
	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), nil,
		nil, nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
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

func (rdr *NatsER) postMessage(msg []byte) (err error) {
	if rdr.posterJS != nil {
		_, err = rdr.posterJS.Publish(rdr.posterSubject, msg)
		return
	}
	return rdr.poster.Publish(rdr.posterSubject, msg)
}

func (rdr *NatsER) createPoster(opts map[string]interface{}) (err error) {
	if rdr.poster, err = nats.Connect(utils.FirstNonEmpty(rdr.Config().ProcessedPath, rdr.Config().SourcePath)); err != nil {
		return
	}
	jsOpt := rdr.jetStream
	if jsOptVal, has := opts["natsJetStreamProcessed"]; has {
		if jsOpt, err = utils.IfaceAsBool(jsOptVal); err != nil {
			return
		}
	}

	if jsOpt {
		// Create JetStream Context
		rdr.posterJS, err = rdr.poster.JetStream()
	}
	return
}
