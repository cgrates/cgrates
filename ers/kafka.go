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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	kafka "github.com/segmentio/kafka-go"
)

const (
	defaultTopic   = "cgrates"
	defaultGroupID = "cgrates"
	defaultMaxWait = time.Millisecond
)

// NewKafkaER return a new kafka event reader
func NewKafkaER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (er EventReader, err error) {

	rdr := &KafkaER{
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

// KafkaER implements EventReader interface for kafka message
type KafkaER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	dialURL string
	topic   string
	groupID string
	maxWait time.Duration

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
	cap       chan struct{}
}

// Config returns the curent configuration
func (rdr *KafkaER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the kafka topic
func (rdr *KafkaER) Serve() (err error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{rdr.dialURL},
		GroupID: rdr.groupID,
		Topic:   rdr.topic,
		MaxWait: rdr.maxWait,
	})

	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		return
	}

	go func(r *kafka.Reader) { // use a secondary gorutine because the ReadMessage is blocking function
		select {
		case <-rdr.rdrExit:
			utils.Logger.Info(
				fmt.Sprintf("<%s> stop monitoring kafka path <%s>",
					utils.ERs, rdr.dialURL))
			r.Close() // already locked in library
			return
		}
	}(r)
	go rdr.readLoop(r) // read until the connection is closed
	return
}

func (rdr *KafkaER) readLoop(r *kafka.Reader) {
	for {
		if rdr.Config().ConcurrentReqs != -1 {
			<-rdr.cap // do not try to read if the limit is reached
		}
		msg, err := r.ReadMessage(context.Background())
		if err != nil {
			if err == io.EOF {
				// ignore io.EOF received from closing the connection from our side
				// this is happening when we stop the reader
				return
			}
			//  send it to the error channel
			rdr.rdrErr <- err
			return
		}
		go func(msg kafka.Message) {
			if err := rdr.processMessage(msg.Value); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> processing message %s error: %s",
						utils.ERs, string(msg.Key), err.Error()))
			}
			if rdr.Config().ProcessedPath != utils.EmptyString { // post it
				if err := engine.PostersCache.PostKafka(rdr.Config().ProcessedPath,
					rdr.cgrCfg.GeneralCfg().PosterAttempts, msg.Value, string(msg.Key)); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> writing message %s error: %s",
							utils.ERs, string(msg.Key), err.Error()))
				}
			}
			if rdr.Config().ConcurrentReqs != -1 {
				rdr.cap <- struct{}{}
			}
		}(msg)
	}
}

func (rdr *KafkaER) processMessage(msg []byte) (err error) {
	var decodedMessage map[string]interface{}
	if err = json.Unmarshal(msg, &decodedMessage); err != nil {
		return
	}

	reqVars := make(map[string]interface{})
	agReq := agents.NewAgentRequest(
		utils.NavigableMap(decodedMessage), reqVars,
		nil, nil, rdr.Config().Tenant,
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
	rdr.rdrEvents <- &erEvent{cgrEvent: config.NMAsCGREvent(agReq.CGRRequest,
		agReq.Tenant, utils.NestingSep),
		rdrCfg: rdr.Config()}
	return
}

func (rdr *KafkaER) setURL(dialURL string) (err error) {
	var u *url.URL
	if u, err = url.Parse(dialURL); err != nil {
		return
	}
	qry := u.Query()

	rdr.dialURL = strings.Split(dialURL, "?")[0]
	rdr.topic = defaultTopic
	if vals, has := qry[utils.KafkaTopic]; has && len(vals) != 0 {
		rdr.topic = vals[0]
	}
	rdr.groupID = defaultGroupID
	if vals, has := qry[utils.KafkaGroupID]; has && len(vals) != 0 {
		rdr.groupID = vals[0]
	}
	rdr.maxWait = defaultMaxWait
	if vals, has := qry[utils.KafkaMaxWait]; has && len(vals) != 0 {
		rdr.maxWait, err = time.ParseDuration(vals[0])
	}
	return
}
