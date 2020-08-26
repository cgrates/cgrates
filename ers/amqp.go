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
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/streadway/amqp"
)

const (
	defaultConsumerTag = "cgrates"
	consumerTag        = "consumer_tag"
)

// NewAMQPER return a new kafka event reader
func NewAMQPER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &AMQPER{
		cgrCfg:    cfg,
		cfgIdx:    cfgIdx,
		fltrS:     fltrS,
		rdrEvents: rdrEvents,
		rdrExit:   rdrExit,
		rdrErr:    rdrErr,
	}
	if concReq := rdr.Config().ConcurrentReqs; concReq != -1 {
		rdr.cap = make(chan struct{}, concReq)
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	er = rdr
	err = rdr.setURL(rdr.Config().SourcePath)
	return
}

// AMQPER implements EventReader interface for kafka message
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

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
	cap       chan struct{}

	conn    *amqp.Connection
	channel *amqp.Channel
}

// Config returns the curent configuration
func (rdr *AMQPER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the kafka topic
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
				fmt.Sprintf("<%s> stop monitoring kafka path <%s>",
					utils.ERs, rdr.dialURL))
			rdr.close()
			return
		case msg := <-msgChan:
			if len(msg.Body) == 0 {
				continue
			}
			go func(msg amqp.Delivery) {
				if err := rdr.processMessage(msg.Body); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, msg.MessageId, err.Error()))
				}
				if rdr.Config().ProcessedPath != utils.EmptyString { // post it
					if err := engine.PostersCache.PostAMQP(rdr.Config().ProcessedPath,
						rdr.cgrCfg.GeneralCfg().PosterAttempts, msg.Body); err != nil {
						utils.Logger.Warning(
							fmt.Sprintf("<%s> writing message %s error: %s",
								utils.ERs, msg.MessageId, err.Error()))
					}
				}
				if rdr.Config().ConcurrentReqs != -1 {
					rdr.cap <- struct{}{}
				}
			}(msg)
		}
	}
}

func (rdr *AMQPER) processMessage(msg []byte) (err error) {
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
		rdr.fltrS, nil, nil) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	rdr.rdrEvents <- &erEvent{
		cgrEvent: config.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, utils.NestingSep),
		rdrCfg:   rdr.Config(),
		opts:     config.NMAsMapInterface(agReq.Opts, utils.NestingSep),
	}
	return
}

func (rdr *AMQPER) setURL(dialURL string) (err error) {
	var u *url.URL
	if u, err = url.Parse(dialURL); err != nil {
		return
	}
	qry := u.Query()
	q := url.Values{}
	for _, key := range engine.AMQPPosibleQuery {
		if vals, has := qry[key]; has && len(vals) != 0 {
			q.Add(key, vals[0])
		}
	}
	rdr.dialURL = strings.Split(dialURL, "?")[0]
	if params := q.Encode(); params != utils.EmptyString {
		rdr.dialURL += "?" + params

	}
	rdr.queueID = engine.DefaultQueueID
	if vals, has := qry[engine.QueueID]; has && len(vals) != 0 {
		rdr.queueID = vals[0]
	}
	rdr.tag = defaultConsumerTag
	if vals, has := qry[consumerTag]; has && len(vals) != 0 {
		rdr.tag = vals[0]
	}

	if vals, has := qry[engine.RoutingKey]; has && len(vals) != 0 {
		rdr.routingKey = vals[0]
	}
	if vals, has := qry[engine.Exchange]; has && len(vals) != 0 {
		rdr.exchange = vals[0]
		rdr.exchangeType = engine.DefaultExchangeType
	}
	if vals, has := qry[engine.ExchangeType]; has && len(vals) != 0 {
		rdr.exchangeType = vals[0]
	}

	return nil
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
