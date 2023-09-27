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

package ers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NewNatsER return a new amqp event reader
func NewNatsER(cfg *config.CGRConfig, cfgIdx int,
	rdrEvents, partialEvents chan *erEvent, rdrErr chan error,
	fltrS *engine.FilterS, rdrExit chan struct{}) (EventReader, error) {
	rdr := &NatsER{
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
		for i := 0; i < concReq; i++ {
			rdr.cap <- struct{}{}
		}
	}
	err := rdr.processOpts()
	if err != nil {
		return nil, err
	}
	err = rdr.createPoster()
	if err != nil {
		return nil, err
	}
	return rdr, nil
}

// NatsER implements EventReader interface for amqp message
type NatsER struct {
	// sync.RWMutex
	cgrCfg *config.CGRConfig
	cfgIdx int // index of config instance within ERsCfg.Readers
	fltrS  *engine.FilterS

	rdrEvents     chan *erEvent // channel to dispatch the events created to
	partialEvents chan *erEvent // channel to dispatch the partial events created to
	rdrExit       chan struct{}
	rdrErr        chan error
	cap           chan struct{}

	subject      string
	queueID      string
	jetStream    bool
	consumerName string
	streamName   string
	opts         []nats.Option

	poster *ees.NatsEE
}

// Config returns the curent configuration
func (rdr *NatsER) Config() *config.EventReaderCfg {
	return rdr.cgrCfg.ERsCfg().Readers[rdr.cfgIdx]
}

// Serve will subscribe to a NATS subject and process incoming messages until the rdrExit channel
// will be closed.
func (rdr *NatsER) Serve() error {

	// Establish a connection to the nats server.
	nc, err := nats.Connect(rdr.Config().SourcePath, rdr.opts...)
	if err != nil {
		return err
	}

	// Define the message handler. Its content will get executed for every received message.
	handleMessage := func(msgData []byte) {

		// If the rdr.cap channel buffer is empty, block until a resource is available. Otherwise
		// allocate one resource and start processing the message.
		if rdr.Config().ConcurrentReqs != -1 {
			<-rdr.cap
		}
		go func() {
			handlerErr := rdr.processMessage(msgData)
			if handlerErr != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> processing message %s error: %s",
						utils.ERs, string(msgData), handlerErr.Error()))
			}

			// Export the received message if a poster has been defined.
			if rdr.poster != nil {
				handlerErr = ees.ExportWithAttempts(rdr.poster, msgData, utils.EmptyString)
				if handlerErr != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<%s> writing message %s error: %s",
							utils.ERs, string(msgData), handlerErr.Error()))
				}
			}

			// Release the resource back to rdr.cap channel.
			if rdr.Config().ConcurrentReqs != -1 {
				rdr.cap <- struct{}{}
			}

		}()
	}

	// Subscribe to the appropriate NATS subject.
	if !rdr.jetStream {
		_, err = nc.QueueSubscribe(rdr.subject, rdr.queueID, func(msg *nats.Msg) {
			handleMessage(msg.Data)
		})
		if err != nil {
			nc.Drain()
			return err
		}
	} else {
		var js jetstream.JetStream
		js, err = jetstream.New(nc)
		if err != nil {
			nc.Drain()
			return err
		}
		ctx := context.TODO()
		if jsMaxWait := rdr.Config().Opts.NATS.JetStreamMaxWait; jsMaxWait != nil {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, *jsMaxWait)
			defer cancel()
		}

		var cons jetstream.Consumer
		cons, err = js.CreateOrUpdateConsumer(ctx, rdr.streamName, jetstream.ConsumerConfig{
			FilterSubject: rdr.subject,
			Durable:       rdr.consumerName,
			AckPolicy:     jetstream.AckAllPolicy,
		})
		if err != nil {
			nc.Drain()
			return err
		}

		_, err = cons.Consume(func(msg jetstream.Msg) {
			handleMessage(msg.Data())
		})
		if err != nil {
			nc.Drain()
			return err
		}
	}

	go func() {

		// Wait for exit signal.
		<-rdr.rdrExit
		utils.Logger.Info(
			fmt.Sprintf("<%s> stop monitoring nats path <%s>",
				utils.ERs, rdr.Config().SourcePath))
		nc.Drain()
		if rdr.poster != nil {
			rdr.poster.Close()
		}
	}()

	return nil
}

func (rdr *NatsER) processMessage(msg []byte) (err error) {
	var decodedMessage map[string]any
	if err = json.Unmarshal(msg, &decodedMessage); err != nil {
		return
	}
	agReq := agents.NewAgentRequest(
		utils.MapStorage(decodedMessage), nil,
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

func (rdr *NatsER) createPoster() (err error) {
	processedOpt := getProcessedOptions(rdr.Config().Opts)
	if processedOpt == nil && len(rdr.Config().ProcessedPath) == 0 {
		return
	}
	eeCfg := config.NewEventExporterCfg(rdr.Config().ID, "", utils.FirstNonEmpty(rdr.Config().ProcessedPath, rdr.Config().SourcePath),
		rdr.cgrCfg.GeneralCfg().FailedPostsDir, rdr.cgrCfg.GeneralCfg().PosterAttempts, processedOpt)
	rdr.poster, err = ees.NewNatsEE(eeCfg, rdr.cgrCfg.GeneralCfg().NodeID,
		rdr.cgrCfg.GeneralCfg().ConnectTimeout, nil)
	return
}

func (rdr *NatsER) processOpts() (err error) {
	if rdr.Config().Opts.NATS.Subject != nil {
		rdr.subject = *rdr.Config().Opts.NATS.Subject
	}

	rdr.queueID = rdr.cgrCfg.GeneralCfg().NodeID
	if rdr.Config().Opts.NATS.QueueID != nil {
		rdr.queueID = *rdr.Config().Opts.NATS.QueueID
	}

	rdr.consumerName = utils.CGRateSLwr
	if rdr.Config().Opts.NATS.ConsumerName != nil {
		rdr.consumerName = *rdr.Config().Opts.NATS.ConsumerName
	}

	if rdr.Config().Opts.NATS.StreamName != nil {
		rdr.streamName = *rdr.Config().Opts.NATS.StreamName
	}

	if rdr.Config().Opts.NATS.JetStream != nil {
		rdr.jetStream = *rdr.Config().Opts.NATS.JetStream
	}
	rdr.opts, err = GetNatsOpts(rdr.Config().Opts.NATS,
		rdr.cgrCfg.GeneralCfg().NodeID,
		rdr.cgrCfg.GeneralCfg().ConnectTimeout)
	return
}

func GetNatsOpts(opts *config.NATSROpts, nodeID string, connTimeout time.Duration) (nop []nats.Option, err error) {
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
