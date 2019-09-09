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
	defaultTopic   = "cgrates_cdrc"
	defaultGroupID = "cgrates_consumer"
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

	rdrEvents chan *erEvent // channel to dispatch the events created to
	rdrExit   chan struct{}
	rdrErr    chan error
}

func (rdr *KafkaER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

func (rdr *KafkaER) Serve() (err error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          []string{rdr.dialURL},
		GroupID:          rdr.groupID,
		Topic:            rdr.topic,
		MinBytes:         10e3, // 10KB
		MaxBytes:         10e6, // 10MB
		RebalanceTimeout: time.Second,
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
	go func(r *kafka.Reader) { // read until the conection is closed
		for {
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
			if err := rdr.processMessage(msg.Value); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> processing message %s error: %s",
						utils.ERs, string(msg.Key), err.Error()))
			}
		}
	}(r)
	return
}

func (rdr *KafkaER) processMessage(msg []byte) (err error) {
	var decodedMessage map[string]interface{}
	if err = json.Unmarshal(msg, &decodedMessage); err != nil {
		return
	}

	reqVars := make(map[string]interface{})
	agReq := agents.NewAgentRequest(
		config.NewNavigableMap(decodedMessage), reqVars,
		nil, nil, rdr.Config().Tenant,
		rdr.cgrCfg.GeneralCfg().DefaultTenant,
		utils.FirstNonEmpty(rdr.Config().Timezone,
			rdr.cgrCfg.GeneralCfg().DefaultTimezone),
		rdr.fltrS) // create an AgentRequest
	var pass bool
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	var navMp *config.NavigableMap
	if navMp, err = agReq.AsNavigableMap(rdr.Config().ContentFields); err != nil {
		return
	}
	rdr.rdrEvents <- &erEvent{cgrEvent: navMp.AsCGREvent(
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
	return
}
