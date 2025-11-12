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
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NewNatsEE creates a kafka poster
func NewNatsEE(cfg *config.EventExporterCfg, nodeID string, connTimeout time.Duration, em *utils.ExporterMetrics) (natsPstr *NatsEE, err error) {
	natsPstr = &NatsEE{
		cfg:     cfg,
		em:      em,
		subject: utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	err = natsPstr.parseOpts(cfg.Opts, nodeID, connTimeout)
	return
}

// NatsEE is a kafka poster
type NatsEE struct {
	subject   string // identifier of the CDR queue where we publish
	jetStream bool
	opts      []nats.Option

	poster   *nats.Conn
	posterJS jetstream.JetStream

	cfg          *config.EventExporterCfg
	em           *utils.ExporterMetrics
	reqs         *concReq
	sync.RWMutex // protect writer
	bytePreparing
}

func (pstr *NatsEE) parseOpts(opts *config.EventExporterOpts, nodeID string, connTimeout time.Duration) error {
	if opts.NATSJetStream != nil {
		pstr.jetStream = *opts.NATSJetStream
	}
	if opts.NATSSubject != nil {
		pstr.subject = *opts.NATSSubject
	}
	var err error
	pstr.opts, err = GetNatsOpts(opts, nodeID, connTimeout)
	return err
}

func (pstr *NatsEE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *NatsEE) Connect() error {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.poster != nil {
		return nil
	}

	var err error
	pstr.poster, err = nats.Connect(pstr.Cfg().ExportPath, pstr.opts...)
	if err != nil {
		return err
	}
	if pstr.jetStream {
		pstr.posterJS, err = jetstream.New(pstr.poster)
	}
	return err
}

func (pstr *NatsEE) ExportEvent(ctx *context.Context, content, _ any) error {
	pstr.reqs.get()
	defer pstr.reqs.done()
	pstr.RLock()
	defer pstr.RUnlock()
	if pstr.poster == nil {
		return utils.ErrDisconnected
	}
	var err error
	if pstr.jetStream {
		ctx := context.TODO()
		if pstr.cfg.Opts.NATSJetStreamMaxWait != nil {
			ctx, _ = context.WithTimeout(ctx, *pstr.cfg.Opts.NATSJetStreamMaxWait)
		}
		_, err = pstr.posterJS.Publish(ctx, pstr.subject, content.([]byte))
	} else {
		err = pstr.poster.Publish(pstr.subject, content.([]byte))
	}
	return err
}

func (pstr *NatsEE) Close() error {
	pstr.Lock()
	defer pstr.Unlock()

	if pstr.poster == nil {
		return nil
	}

	err := pstr.poster.Drain()
	pstr.poster = nil
	return err
}

func (pstr *NatsEE) GetMetrics() *utils.ExporterMetrics { return pstr.em }

func (pstr *NatsEE) ExtraData(ev *utils.CGREvent) any { return nil }

func GetNatsOpts(opts *config.EventExporterOpts, nodeID string, connTimeout time.Duration) ([]nats.Option, error) {
	natsOpts := make([]nats.Option, 0, 7)
	natsOpts = append(natsOpts, nats.Name(utils.CGRateSLwr+nodeID),
		nats.Timeout(connTimeout),
		nats.DrainTimeout(time.Second))
	if opts.NATSJWTFile != nil {
		keys := make([]string, 0, 1)
		if opts.NATSSeedFile != nil {
			keys = append(keys, *opts.NATSSeedFile)
		}
		natsOpts = append(natsOpts, nats.UserCredentials(*opts.NATSJWTFile, keys...))
	}
	if opts.NATSSeedFile != nil {
		opt, err := nats.NkeyOptionFromSeed(*opts.NATSSeedFile)
		if err != nil {
			return nil, err
		}
		natsOpts = append(natsOpts, opt)
	}

	switch {
	case opts.NATSClientCertificate != nil && opts.NATSClientKey != nil:
		natsOpts = append(natsOpts, nats.ClientCert(*opts.NATSClientCertificate, *opts.NATSClientKey))
	case opts.NATSClientCertificate != nil:
		return nil, fmt.Errorf("has certificate but no key")
	case opts.NATSClientKey != nil:
		return nil, fmt.Errorf("has key but no certificate")
	}

	if opts.NATSCertificateAuthority != nil {
		natsOpts = append(natsOpts,
			func(o *nats.Options) error {
				pool, err := x509.SystemCertPool()
				if err != nil {
					return err
				}
				rootPEM, err := os.ReadFile(*opts.NATSCertificateAuthority)
				if err != nil || rootPEM == nil {
					return fmt.Errorf("nats: error loading or parsing rootCA file: %v", err)
				}
				ok := pool.AppendCertsFromPEM(rootPEM)
				if !ok {
					return fmt.Errorf("nats: failed to parse root certificate from %q",
						*opts.NATSCertificateAuthority)
				}
				if o.TLSConfig == nil {
					o.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
				}
				o.TLSConfig.RootCAs = pool
				o.Secure = true
				return nil
			})
	}
	return natsOpts, nil
}
