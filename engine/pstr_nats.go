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
package engine

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
)

// NewNatsPoster creates a kafka poster
func NewNatsPoster(dialURL string, attempts int, opts map[string]interface{}, nodeID string, connTimeout time.Duration) (natsPstr *NatsPoster, err error) {
	natsPstr = &NatsPoster{
		dialURL:  dialURL,
		subject:  utils.DefaultQueueID,
		attempts: attempts,
	}
	err = natsPstr.parseOpt(opts, nodeID, connTimeout)
	return
}

// NatsPoster is a kafka poster
type NatsPoster struct {
	dialURL    string
	subject    string // identifier of the CDR queue where we publish
	attempts   int
	jetStream  bool
	opts       []nats.Option
	sync.Mutex // protect writer

	poster   *nats.Conn
	posterJS nats.JetStreamContext
}

// Post is the method being called when we need to post anything in the queue
// the optional chn will permits channel caching
func (pstr *NatsPoster) Post(content []byte, _ string) (err error) {
	fib := utils.Fib()
	for i := 0; i < pstr.attempts; i++ {
		if err = pstr.newPostWriter(); err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	if err != nil {
		utils.Logger.Warning(fmt.Sprintf("<NatsPoster> connecting to nats server, err: %s", err.Error()))
		return
	}
	for i := 0; i < pstr.attempts; i++ {
		pstr.Lock()

		if pstr.jetStream {
			_, err = pstr.posterJS.Publish(pstr.subject, content)
		} else {
			err = pstr.poster.Publish(pstr.subject, content)
		}
		pstr.Unlock()

		if err == nil {
			break
		}
		if i+1 < pstr.attempts {
			time.Sleep(time.Duration(fib()) * time.Second)
		}
	}
	return
}

// Close closes the kafka writer
func (pstr *NatsPoster) Close() {
	pstr.Lock()
	if pstr.poster != nil {
		pstr.poster.Drain()
	}
	pstr.poster = nil
	pstr.Unlock()
}

func (pstr *NatsPoster) parseOpt(opts map[string]interface{}, nodeID string, connTimeout time.Duration) (err error) {
	if useJetStreamVal, has := opts[utils.NatsJetStream]; has {
		if pstr.jetStream, err = utils.IfaceAsBool(useJetStreamVal); err != nil {
			return
		}
	}
	pstr.subject = utils.DefaultQueueID
	if vals, has := opts[utils.NatsSubject]; has {
		pstr.subject = utils.IfaceAsString(vals)
	}
	pstr.opts, err = GetNatsOpts(opts, nodeID, connTimeout)
	return
}

func (pstr *NatsPoster) newPostWriter() (err error) {
	pstr.Lock()
	defer pstr.Unlock()
	if pstr.poster == nil {
		if pstr.poster, err = nats.Connect(pstr.dialURL, pstr.opts...); err != nil {
			return
		}
		if pstr.jetStream {
			pstr.posterJS, err = pstr.poster.JetStream()
		}
	}
	return
}

func GetNatsOpts(opts map[string]interface{}, nodeID string, connTimeout time.Duration) (nop []nats.Option, err error) {
	nop = make([]nats.Option, 0, 7)
	nop = append(nop, nats.Name(utils.CGRateSLwr+nodeID),
		nats.Timeout(connTimeout),
		nats.DrainTimeout(time.Second))
	if userFile, has := opts[utils.NatsJWTFile]; has {
		keys := make([]string, 0, 1)
		if keyFile, has := opts[utils.NatsSeedFile]; has {
			keys = append(keys, utils.IfaceAsString(keyFile))
		}
		nop = append(nop, nats.UserCredentials(utils.IfaceAsString(userFile), keys...))
	}
	if nkeyFile, has := opts[utils.NatsSeedFile]; has {
		opt, err := nats.NkeyOptionFromSeed(utils.IfaceAsString(nkeyFile))
		if err != nil {
			return nil, err
		}
		nop = append(nop, opt)
	}
	if certFile, has := opts[utils.NatsClientCertificate]; has {
		clientFile, has := opts[utils.NatsClientKey]
		if !has {
			err = fmt.Errorf("has certificate but no key")
			return
		}
		nop = append(nop, nats.ClientCert(utils.IfaceAsString(certFile), utils.IfaceAsString(clientFile)))
	} else if _, has := opts[utils.NatsClientKey]; has {
		err = fmt.Errorf("has key but no certificate")
		return
	}

	if caFile, has := opts[utils.NatsCertificateAuthority]; has {
		nop = append(nop,
			func(o *nats.Options) error {
				pool, err := x509.SystemCertPool()
				if err != nil {
					return err
				}
				rootPEM, err := ioutil.ReadFile(utils.IfaceAsString(caFile))
				if err != nil || rootPEM == nil {
					return fmt.Errorf("nats: error loading or parsing rootCA file: %v", err)
				}
				ok := pool.AppendCertsFromPEM(rootPEM)
				if !ok {
					return fmt.Errorf("nats: failed to parse root certificate from %q", caFile)
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
