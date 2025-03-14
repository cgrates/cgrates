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
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	kafka "github.com/segmentio/kafka-go"
)

// NewKafkaEE creates a kafka poster
func NewKafkaEE(cfg *config.EventExporterCfg, dc *utils.ExporterMetrics) (*KafkaEE, error) {
	pstr := &KafkaEE{
		cfg:  cfg,
		dc:   dc,
		reqs: newConcReq(cfg.ConcurrentRequests),
	}

	opts := cfg.Opts.Kafka
	topic := utils.DefaultQueueID
	if opts.Topic != nil {
		topic = *opts.Topic
	}

	// Configure TLS if enabled.
	var tlsCfg *tls.Config
	if opts.TLS != nil && *opts.TLS {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}

		// Load additional CA certificates if a path is provided.
		if opts.CAPath != nil && *opts.CAPath != "" {
			ca, err := os.ReadFile(*opts.CAPath)
			if err != nil {
				return nil, err
			}
			if !rootCAs.AppendCertsFromPEM(ca) {
				return nil, errors.New("failed to append certificates from PEM file")
			}
		}

		tlsCfg = &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: opts.SkipTLSVerify != nil && *opts.SkipTLSVerify,
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

	if opts.BatchSize != nil {
		pstr.writer.BatchSize = *opts.BatchSize
	}

	return pstr, nil
}

// KafkaEE is a kafka poster
type KafkaEE struct {
	writer *kafka.Writer
	cfg    *config.EventExporterCfg
	dc     *utils.ExporterMetrics
	reqs   *concReq
	bytePreparing
}

func (k *KafkaEE) Cfg() *config.EventExporterCfg { return k.cfg }

func (k *KafkaEE) Connect() error { return nil }

func (k *KafkaEE) ExportEvent(content any, key string) (err error) {
	k.reqs.get()
	defer k.reqs.done()
	return k.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
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

func (k *KafkaEE) GetMetrics() *utils.ExporterMetrics { return k.dc }
