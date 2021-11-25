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
	"io/ioutil"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

// NewNatsEE creates a kafka poster
func NewNatsEE(cfg *config.EventExporterCfg, nodeID string, connTimeout time.Duration, dc *utils.SafeMapStorage) (natsPstr *NatsEE, err error) {
	natsPstr = &NatsEE{
		cfg:     cfg,
		dc:      dc,
		subject: utils.DefaultQueueID,
		reqs:    newConcReq(cfg.ConcurrentRequests),
	}
	err = natsPstr.parseOpt(cfg.Opts, nodeID, connTimeout)
	return
}

// NatsEE is a kafka poster
type NatsEE struct {
	subject   string // identifier of the CDR queue where we publish
	jetStream bool
	opts      []nats.Option
	jsOpts    []nats.JSOpt

	poster   *nats.Conn
	posterJS nats.JetStreamContext

	cfg          *config.EventExporterCfg
	dc           *utils.SafeMapStorage
	reqs         *concReq
	sync.RWMutex // protect writer
	bytePreparing
}

func (pstr *NatsEE) parseOpt(opts *config.EventExporterOpts, nodeID string, connTimeout time.Duration) (err error) {
	if opts.NATSJetStream != nil {
		pstr.jetStream = *opts.NATSJetStream
	}
	pstr.subject = utils.DefaultQueueID
	if opts.NATSSubject != nil {
		pstr.subject = *opts.NATSSubject
	}
	pstr.opts, err = GetNatsOpts(opts, nodeID, connTimeout)
	if pstr.jetStream {
		if opts.NATSJetStreamMaxWait != nil {
			pstr.jsOpts = []nats.JSOpt{nats.MaxWait(*opts.NATSJetStreamMaxWait)}
		}
	}
	return
}

func (pstr *NatsEE) Cfg() *config.EventExporterCfg { return pstr.cfg }

func (pstr *NatsEE) Connect() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.poster == nil {
		if pstr.poster, err = nats.Connect(pstr.Cfg().ExportPath, pstr.opts...); err != nil {
			return
		}
		if pstr.jetStream {
			pstr.posterJS, err = pstr.poster.JetStream(pstr.jsOpts...)
		}
	}
	return
}

func (pstr *NatsEE) ExportEvent(content interface{}, _ string) (err error) {
	pstr.reqs.get()
	pstr.RLock()
	if pstr.poster == nil {
		pstr.RUnlock()
		pstr.reqs.done()
		return utils.ErrDisconnected
	}
	if pstr.jetStream {
		_, err = pstr.posterJS.Publish(pstr.subject, content.([]byte))
	} else {
		err = pstr.poster.Publish(pstr.subject, content.([]byte))
	}
	pstr.RUnlock()
	pstr.reqs.done()
	return
}

func (pstr *NatsEE) Close() (err error) {
	pstr.Lock()
	if pstr.poster != nil {
		err = pstr.poster.Drain()
		pstr.poster = nil
	}
	pstr.Unlock()
	return
}

func (pstr *NatsEE) GetMetrics() *utils.SafeMapStorage { return pstr.dc }

func GetNatsOpts(opts *config.EventExporterOpts, nodeID string, connTimeout time.Duration) (nop []nats.Option, err error) {
	nop = make([]nats.Option, 0, 7)
	nop = append(nop, nats.Name(utils.CGRateSLwr+nodeID),
		nats.Timeout(connTimeout),
		nats.DrainTimeout(time.Second))
	if opts.NATSJWTFile != nil {
		keys := make([]string, 0, 1)
		if opts.NATSSeedFile != nil {
			keys = append(keys, *opts.NATSSeedFile)
		}
		nop = append(nop, nats.UserCredentials(*opts.NATSJWTFile, keys...))
	}
	if opts.NATSSeedFile != nil {
		opt, err := nats.NkeyOptionFromSeed(*opts.NATSSeedFile)
		if err != nil {
			return nil, err
		}
		nop = append(nop, opt)
	}
	if opts.NATSClientCertificate != nil {
		if opts.NATSClientKey == nil {
			err = fmt.Errorf("has certificate but no key")
			return
		}
		nop = append(nop, nats.ClientCert(*opts.NATSClientCertificate, *opts.NATSClientKey))
	} else if opts.NATSClientKey != nil {
		err = fmt.Errorf("has key but no certificate")
		return
	}

	if opts.NATSCertificateAuthority != nil {
		nop = append(nop,
			func(o *nats.Options) error {
				pool, err := x509.SystemCertPool()
				if err != nil {
					return err
				}
				rootPEM, err := ioutil.ReadFile(*opts.NATSCertificateAuthority)
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
	return
}
