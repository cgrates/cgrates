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

package ees

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/twmb/franz-go/pkg/kgo"
)

const defaultKafkaTimeout = 30 * time.Second

// NewKafkaEE creates a kafka poster
func NewKafkaEE(cfg *config.EventExporterCfg, em *utils.ExporterMetrics) (*KafkaEE, error) {
	pstr := &KafkaEE{
		cfg:  cfg,
		em:   em,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}

	topic := utils.DefaultQueueID
	if cfg.Opts.KafkaTopic != nil {
		topic = *cfg.Opts.KafkaTopic
	}

	kgoOpts := []kgo.Opt{
		kgo.SeedBrokers(cfg.ExportPath),
		kgo.DefaultProduceTopic(topic),
		kgo.DisableIdempotentWrite(),
		kgo.RecordRetries(0),
	}
	if cfg.Opts.KafkaLinger != nil {
		kgoOpts = append(kgoOpts, kgo.ProducerLinger(*cfg.Opts.KafkaLinger))
	}

	// Configure TLS if enabled.
	if cfg.Opts.KafkaTLS != nil && *cfg.Opts.KafkaTLS {
		tlsCfg, err := buildTLSConfig(cfg.Opts.KafkaCAPath, cfg.Opts.KafkaSkipTLSVerify)
		if err != nil {
			return nil, err
		}
		kgoOpts = append(kgoOpts, kgo.DialTLSConfig(tlsCfg))
	}

	pstr.timeout = defaultKafkaTimeout
	if cfg.Opts.KafkaDeliveryTimeout != nil {
		pstr.timeout = *cfg.Opts.KafkaDeliveryTimeout
	}

	var err error
	pstr.client, err = kgo.NewClient(kgoOpts...)
	if err != nil {
		return nil, err
	}

	return pstr, nil
}

// KafkaEE is a kafka poster
type KafkaEE struct {
	client  *kgo.Client
	timeout time.Duration
	cfg     *config.EventExporterCfg
	em      *utils.ExporterMetrics
	reqs    *concReq
	bytePreparing
}

func (k *KafkaEE) Cfg() *config.EventExporterCfg { return k.cfg }

func (k *KafkaEE) Connect() error { return nil }

func (k *KafkaEE) ExportEvent(_ *context.Context, content any, key any) error {
	k.reqs.get()
	defer k.reqs.done()
	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()
	rec := &kgo.Record{Key: []byte(key.(string)), Value: content.([]byte)}
	ch := make(chan error, 1)
	k.client.Produce(ctx, rec, func(_ *kgo.Record, err error) { ch <- err })
	return <-ch
}

func (k *KafkaEE) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()
	err := k.client.Flush(ctx)
	k.client.Close()
	return err
}

func (k *KafkaEE) GetMetrics() *utils.ExporterMetrics { return k.em }
func (k *KafkaEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}

func buildTLSConfig(caPath *string, skipVerify *bool) (*tls.Config, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if caPath != nil && *caPath != "" {
		ca, err := os.ReadFile(*caPath)
		if err != nil {
			return nil, err
		}
		if !rootCAs.AppendCertsFromPEM(ca) {
			return nil, errors.New("failed to append certificates from PEM file")
		}
	}
	return &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: skipVerify != nil && *skipVerify,
	}, nil
}
