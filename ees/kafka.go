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
	"errors"
	"os"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/segmentio/kafka-go"
)

// NewKafkaEE creates a kafka poster
func NewKafkaEE(cfg *config.EventExporterCfg, dc *utils.SafeMapStorage) (*KafkaEE, error) {
	pstr := &KafkaEE{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}

	topic := utils.DefaultQueueID
	if cfg.Opts.KafkaTopic != nil {
		topic = *cfg.Opts.KafkaTopic
	}

	// Configure TLS if enabled.
	var tlsCfg *tls.Config
	if cfg.Opts.KafkaTLS != nil && *cfg.Opts.KafkaTLS {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Load additional CA certificates if a path is provided.
		if cfg.Opts.KafkaCAPath != nil && *cfg.Opts.KafkaCAPath != "" {
			ca, err := os.ReadFile(*cfg.Opts.KafkaCAPath)
			if err != nil {
				return nil, err
			}
			if !rootCAs.AppendCertsFromPEM(ca) {
				return nil, errors.New("failed to append certificates from PEM file")
			}
		}

		tlsCfg = &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: cfg.Opts.KafkaSkipTLSVerify != nil && *cfg.Opts.KafkaSkipTLSVerify,
		}
	}

	pstr.writer = &kafka.Writer{
		Addr:  kafka.TCP(pstr.Cfg().ExportPath),
		Topic: topic,

		// Leave it to the ExportWithAttempts function
		// to handle the connect attempts.
		MaxAttempts: 1,

		// To handle both TLS and non-TLS connections consistently in the Close() function,
		// we always specify Transport, even if empty. This allows us to call
		// CloseIdleConnections on our Transport instance, avoiding the need to differentiate
		// between TLS and non-TLS connections.
		Transport: &kafka.Transport{
			TLS: tlsCfg,
		},
	}

	if cfg.Opts.KafkaBatchSize != nil {
		pstr.writer.BatchSize = *cfg.Opts.KafkaBatchSize
	}

	return pstr, nil
}

// KafkaEE is a kafka poster
type KafkaEE struct {
	writer *kafka.Writer
	cfg    *config.EventExporterCfg
	dc     *utils.SafeMapStorage
	reqs   *concReq
	bytePreparing
}

func (k *KafkaEE) Cfg() *config.EventExporterCfg { return k.cfg }

func (k *KafkaEE) Connect() error { return nil }

func (k *KafkaEE) ExportEvent(_ *context.Context, content any, key any) error {
	k.reqs.get()
	defer k.reqs.done()
	return k.writer.WriteMessages(context.TODO(), kafka.Message{
		Key:   []byte(key.(string)),
		Value: content.([]byte),
	})
}

func (k *KafkaEE) Close() error {

	// Manually close idle connections to prevent them from running indefinitely
	// after the Kafka writer is purged. Without this, goroutines will accumulate
	// over time with each new Kafka writer.
	tsp, ok := k.writer.Transport.(*kafka.Transport)
	if ok {
		tsp.CloseIdleConnections()
	}

	return k.writer.Close()
}

func (k *KafkaEE) GetMetrics() *utils.SafeMapStorage { return k.dc }
func (k *KafkaEE) ExtraData(ev *utils.CGREvent) any {
	return utils.ConcatenatedKey(
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaOriginID), utils.GenUUID()),
		utils.FirstNonEmpty(engine.MapEvent(ev.APIOpts).GetStringIgnoreErrors(utils.MetaRunID), utils.MetaDefault),
	)
}
