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
func NewNatsEE(cfg *config.EventExporterCfg, nodeID string, connTimeout time.Duration, dc *utils.ExporterMetrics) (natsPstr *NatsEE, err error) {
	natsPstr = &NatsEE{
		cfg:     cfg,
		dc:      dc,
		subject: utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	err = natsPstr.parseOpts(cfg.Opts.NATS, nodeID, connTimeout)
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
	dc           *utils.ExporterMetrics
	reqs         *concReq
	sync.RWMutex // protect writer
	bytePreparing
}

func (pstr *NatsEE) parseOpts(opts *config.NATSOpts, nodeID string, connTimeout time.Duration) error {
	if opts == nil {
		return nil
	}

	if opts.JetStream != nil {
		pstr.jetStream = *opts.JetStream
	}
	if opts.Subject != nil {
		pstr.subject = *opts.Subject
	}

	var err error
	pstr.opts, err = GetNatsOpts(opts, nodeID, connTimeout)
	if err != nil {
		return err
	}

	return nil
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

func (pstr *NatsEE) ExportEvent(content any, _ string) error {
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
		if pstr.cfg.Opts.NATS.JetStreamMaxWait != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, *pstr.cfg.Opts.NATS.JetStreamMaxWait)
			defer cancel()
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

func (pstr *NatsEE) GetMetrics() *utils.ExporterMetrics { return pstr.dc }

func GetNatsOpts(opts *config.NATSOpts, nodeID string, connTimeout time.Duration) ([]nats.Option, error) {
	natsOpts := make([]nats.Option, 0, 7)
	natsOpts = append(natsOpts, nats.Name(utils.CGRateSLwr+nodeID),
		nats.Timeout(connTimeout),
		nats.DrainTimeout(time.Second))
	if opts.JWTFile != nil {
		keys := make([]string, 0, 1)
		if opts.SeedFile != nil {
			keys = append(keys, *opts.SeedFile)
		}
		natsOpts = append(natsOpts, nats.UserCredentials(*opts.JWTFile, keys...))
	}
	if opts.SeedFile != nil {
		opt, err := nats.NkeyOptionFromSeed(*opts.SeedFile)
		if err != nil {
			return nil, err
		}
		natsOpts = append(natsOpts, opt)
	}

	switch {
	case opts.ClientCertificate != nil && opts.ClientKey != nil:
		natsOpts = append(natsOpts, nats.ClientCert(*opts.ClientCertificate, *opts.ClientKey))
	case opts.ClientCertificate != nil:
		return nil, fmt.Errorf("has certificate but no key")
	case opts.ClientKey != nil:
		return nil, fmt.Errorf("has key but no certificate")
	}

	if opts.CertificateAuthority != nil {
		natsOpts = append(natsOpts,
			func(o *nats.Options) error {
				pool, err := x509.SystemCertPool()
				if err != nil {
					return err
				}
				rootPEM, err := os.ReadFile(*opts.CertificateAuthority)
				if err != nil || rootPEM == nil {
					return fmt.Errorf("nats: error loading or parsing rootCA file: %v", err)
				}
				ok := pool.AppendCertsFromPEM(rootPEM)
				if !ok {
					return fmt.Errorf("nats: failed to parse root certificate from %q",
						*opts.CertificateAuthority)
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
