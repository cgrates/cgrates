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
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaEE creates a kafka poster
func NewKafkaEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) *KafkaEE {
	kfkPstr := &KafkaEE{
		cfg:   cfg,
		dc:    dc,
		topic: utils.DefaultQueueID,
		reqs:  newConcReq(cfg.ConcurrentRequests),
	}
	if cfg.Opts.KafkaTopic != nil {
		kfkPstr.topic = *cfg.Opts.KafkaTopic
	}
	if cfg.Opts.KafkaTLS != nil && *cfg.Opts.KafkaTLS {
		kfkPstr.TLS = true
	}
	if cfg.Opts.KafkaCAPath != nil {
		kfkPstr.caPath = *cfg.Opts.KafkaCAPath
	}
	if cfg.Opts.KafkaSkipTLSVerify != nil && *cfg.Opts.KafkaSkipTLSVerify {
		kfkPstr.skipTLSVerify = true
	}
	return kfkPstr
}

// KafkaEE is a kafka poster
type KafkaEE struct {
	topic         string // identifier of the CDR queue where we publish
	TLS           bool   // if true, it will attempt to authenticate the server
	caPath        string // path to CA pem file
	skipTLSVerify bool   // if true, it skips certificate validation
	writer        *kafka.Writer

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect connection
	bytePreparing
}

func (pstr *KafkaEE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *KafkaEE) Connect() (err error) {
	pstr.Lock()
	if pstr.writer == nil {
		pstr.writer = &kafka.Writer{
			Addr:        kafka.TCP(pstr.Cfg().ExportPath),
			Topic:       pstr.topic,
			MaxAttempts: pstr.Cfg().Attempts,
		}
	}
	if pstr.TLS {
		var rootCAs *x509.CertPool
		if rootCAs, err = x509.SystemCertPool(); err != nil {
			return
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if pstr.caPath != "" {
			var ca []byte
			if ca, err = os.ReadFile(pstr.caPath); err != nil {
				return
			}
			if !rootCAs.AppendCertsFromPEM(ca) {
				return
			}
		}
		pstr.writer.Transport = &kafka.Transport{
			Dial: (&net.Dialer{
				Timeout:   3 * time.Second,
				DualStack: true,
			}).DialContext,
			TLS: &tls.Config{
				RootCAs:            rootCAs,
				InsecureSkipVerify: pstr.skipTLSVerify,
			},
		}
	}
	pstr.Unlock()
	return
}

func (pstr *KafkaEE) ExportEvent(ctx *context.Context, content any, extraData any) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	if pstr.writer == nil {
		pstr.RUnlock()
		pstr.reqs.done()
		return utils.ErrDisconnected
	}
	kafkaKey := extraData.(string)
	err = pstr.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(kafkaKey),
		Value: content.([]byte),
	})
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *KafkaEE) Close() (err error) {
	pstr.Lock()
	if pstr.writer != nil {
		err = pstr.writer.Close()
		pstr.writer = nil
	}
	pstr.Unlock()
	return
}

func (pstr *KafkaEE) GetMetrics() *utils.SafeMapStorage { return pstr.dc }
func (pstr *KafkaEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
