/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package ers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaER return a new kafka event reader
func NewKafkaER(cfg *config.CGRConfig, cfgIdx int, rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (EventReader, error) {
	rdr := &KafkaER{
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
	}
	rdr.dialURL = rdr.Config().SourcePath
	if err := rdr.setOpts(rdr.Config().Opts); err != nil {
		return nil, err
	}
	return rdr, nil

}

// KafkaER implements EventReader interface for kafka message
type KafkaER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	dialURL       string
	topic         string
	groupID       string
	maxWait       time.Duration
	TLS           bool   // if true, it will attempt to authentica the server it connects to
	caPath        string // path to CA pem file
	skipTLSVerify bool   // if true, it skips certificate validation

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}
}

// Config returns the curent configuration
func (rdr *KafkaER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will start the gorutines needed to watch the kafka topic
func (rdr *KafkaER) Serve() (err error) {
	readerCfg := kafka.ReaderConfig{
		Brokers: []string{rdr.dialURL},
		GroupID: rdr.groupID,
		Topic:   rdr.topic,
		MaxWait: rdr.maxWait,
	}
	if rdr.TLS {
		var rootCAs *x509.CertPool
		if rootCAs, err = x509.SystemCertPool(); err != nil {
			return
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if rdr.caPath != "" {
			var ca []byte
			if ca, err = os.ReadFile(rdr.caPath); err != nil {
				return
			}
			if !rootCAs.AppendCertsFromPEM(ca) {
				return
			}
		}
		readerCfg.Dialer = &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
			TLS: &tls.Config{
				RootCAs:            rootCAs,
				InsecureSkipVerify: rdr.skipTLSVerify,
			},
		}
	}

	r := kafka.NewReader(readerCfg)

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
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			return
		}
	}
	for {
		if rdr.Config().ConcurrentReqs != -1 {
			rdr.cap <- struct{}{}
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
			if rdr.Config().ConcurrentReqs != -1 {
				<-rdr.cap
			}
		}(msg)
	}
}

func (rdr *KafkaER) processMessage(msg []byte) (err error) {
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

func (rdr *KafkaER) setOpts(opts *config.EventReaderOpts) (err error) {
	rdr.topic = utils.KafkaDefaultTopic
	rdr.groupID = utils.KafkaDefaultGroupID
	rdr.maxWait = utils.KafkaDefaultMaxWait
	if opts.KafkaTopic != nil {
		rdr.topic = *opts.KafkaTopic
	}
	if opts.KafkaGroupID != nil {
		rdr.groupID = *opts.KafkaGroupID
	}
	if opts.KafkaMaxWait != nil {
		rdr.maxWait = *opts.KafkaMaxWait
	}
	if opts.KafkaTLS != nil && *opts.KafkaTLS {
		rdr.TLS = true
	}
	if opts.KafkaCAPath != nil {
		rdr.caPath = *opts.KafkaCAPath
	}
	if opts.KafkaSkipTLSVerify != nil && *opts.KafkaSkipTLSVerify {
		rdr.skipTLSVerify = true
	}
	return
}
