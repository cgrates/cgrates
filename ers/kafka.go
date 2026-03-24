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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/twmb/franz-go/pkg/kgo"
)

// NewKafkaER return a new kafka event reader
func NewKafkaER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
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
	tls           bool   // if true it will attempt to authenticate the server it connects to
	caPath        string // path to CA pem file
	skipTLSVerify bool   // if true it skips certificate validation

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
func (rdr *KafkaER) Serve() error {
	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(rdr.dialURL),
		kgo.ConsumeTopics(rdr.topic),
	}
	if rdr.maxWait >= 10*time.Millisecond {
		kgoOpts = append(kgoOpts, kgo.FetchMaxWait(rdr.maxWait))
	}
	if rdr.groupID != "" {
		kgoOpts = append(kgoOpts, kgo.ConsumerGroup(rdr.groupID))
	}

	if rdr.tls {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			return err
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if rdr.caPath != "" {
			ca, err := os.ReadFile(rdr.caPath)
			if err != nil {
				return err
			}
			if !rootCAs.AppendCertsFromPEM(ca) {
				return errors.New("failed to append certificates from PEM file")
			}
		}
		kgoOpts = append(kgoOpts, kgo.DialTLSConfig(&tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: rdr.skipTLSVerify,
		}))
	}

	cl, err := kgo.NewClient(kgoOpts...)
	if err != nil {
		return err
	}

	if rdr.Config().RunDelay == time.Duration(0) { // 0 disables the automatic read, maybe done per API
		cl.Close()
		return nil
	}

	go func() {
		<-rdr.rdrExit
		utils.Logger.Info(
			fmt.Sprintf("<%s> stop monitoring kafka path <%s>",
				utils.ERs, rdr.dialURL))
		cl.Close()
	}()
	go rdr.readLoop(cl) // read until the client is closed
	return nil
}

func (rdr *KafkaER) readLoop(cl *kgo.Client) {
	if rdr.Config().StartDelay > 0 {
		select {
		case <-time.After(rdr.Config().StartDelay):
		case <-rdr.rdrExit:
			return
		}
	}
	for {
		fetches := cl.PollFetches(context.Background())
		fetches.EachRecord(func(r *kgo.Record) {
			if rdr.Config().ConcurrentReqs != -1 {
				rdr.cap <- struct{}{}
			}
			go func() {
				if err := rdr.processMessage(r.Value); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> processing message %s error: %s",
							utils.ERs, string(r.Key), err.Error()))
				}
				if rdr.Config().ConcurrentReqs != -1 {
					<-rdr.cap
				}
			}()
		})
		for _, fe := range fetches.Errors() {
			if errors.Is(fe.Err, kgo.ErrClientClosed) || errors.Is(fe.Err, context.Canceled) {
				return
			}
			utils.Logger.Warning(
				fmt.Sprintf("<%s> fetch error on %s/%d: %s",
					utils.ERs, fe.Topic, fe.Partition, fe.Err))
		}
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
	if pass, err = rdr.fltrS.Pass(agReq.Tenant, rdr.Config().Filters,
		agReq); err != nil || !pass {
		return
	}
	if err = agReq.SetFields(rdr.Config().Fields); err != nil {
		return
	}
	cgrEv := utils.NMAsCGREvent(agReq.CGRRequest, agReq.Tenant, agReq.Opts)
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
	if kfkOpts := opts.Kafka; kfkOpts != nil {
		if kfkOpts.Topic != nil {
			rdr.topic = *kfkOpts.Topic
		}
		if kfkOpts.GroupID != nil {
			rdr.groupID = *kfkOpts.GroupID
		}
		if kfkOpts.MaxWait != nil {
			rdr.maxWait = *kfkOpts.MaxWait
		}
		if kfkOpts.TLS != nil && *kfkOpts.TLS {
			rdr.tls = true
		}
		if kfkOpts.CAPath != nil {
			rdr.caPath = *kfkOpts.CAPath
		}
		if kfkOpts.SkipTLSVerify != nil && *kfkOpts.SkipTLSVerify {
			rdr.skipTLSVerify = true
		}
	}
	return
}
