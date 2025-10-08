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

package efs

import (
	"sync"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// FailedExportersEEs used to save the failed post to file
type FailedExportersEEs struct {
	lk             sync.RWMutex
	Path           string
	Opts           *config.EventExporterOpts
	Format         string
	Events         []any
	failedPostsDir string
	module         string

	connMngr *engine.ConnManager
}

func AsOptsEESConfig(opts map[string]any) (*config.EventExporterOpts, error) {
	optsCfg := new(config.EventExporterOpts)
	if len(opts) == 0 {
		return optsCfg, nil
	}
	if _, has := opts[utils.CSVFieldSepOpt]; has {
		optsCfg.CSVFieldSeparator = utils.StringPointer(utils.IfaceAsString(utils.CSVFieldSepOpt))
	}
	if _, has := opts[utils.ElsIndex]; has {
		optsCfg.ElsIndex = utils.StringPointer(utils.IfaceAsString(utils.ElsIndex))
	}
	if _, has := opts[utils.ElsRefresh]; has {
		optsCfg.ElsRefresh = utils.StringPointer(utils.IfaceAsString(utils.ElsRefresh))
	}
	if _, has := opts[utils.ElsOpType]; has {
		optsCfg.ElsOpType = utils.StringPointer(utils.IfaceAsString(utils.ElsOpType))
	}
	if _, has := opts[utils.ElsPipeline]; has {
		optsCfg.ElsPipeline = utils.StringPointer(utils.IfaceAsString(utils.ElsPipeline))
	}
	if _, has := opts[utils.ElsRouting]; has {
		optsCfg.ElsRouting = utils.StringPointer(utils.IfaceAsString(utils.ElsRouting))
	}
	if _, has := opts[utils.ElsTimeout]; has {
		t, err := utils.IfaceAsDuration(utils.ElsTimeout)
		if err != nil {
			return nil, err
		}
		optsCfg.ElsTimeout = &t
	}
	if _, has := opts[utils.ElsWaitForActiveShards]; has {
		optsCfg.ElsWaitForActiveShards = utils.StringPointer(utils.IfaceAsString(utils.ElsWaitForActiveShards))
	}
	if _, has := opts[utils.SQLMaxIdleConnsCfg]; has {
		x, err := utils.IfaceAsInt(utils.SQLMaxIdleConnsCfg)
		if err != nil {
			return nil, err
		}
		optsCfg.SQLMaxIdleConns = utils.IntPointer(x)
	}
	if _, has := opts[utils.SQLMaxOpenConns]; has {
		x, err := utils.IfaceAsInt(utils.SQLMaxOpenConns)
		if err != nil {
			return nil, err
		}
		optsCfg.SQLMaxOpenConns = utils.IntPointer(x)
	}
	if _, has := opts[utils.SQLConnMaxLifetime]; has {
		t, err := utils.IfaceAsDuration(utils.SQLConnMaxLifetime)
		if err != nil {
			return nil, err
		}
		optsCfg.SQLConnMaxLifetime = &t
	}
	if _, has := opts[utils.MYSQLDSNParams]; has {
		optsCfg.MYSQLDSNParams = opts[utils.SQLConnMaxLifetime].(map[string]string)
	}
	if _, has := opts[utils.SQLTableNameOpt]; has {
		optsCfg.SQLTableName = utils.StringPointer(utils.IfaceAsString(utils.SQLTableNameOpt))
	}
	if _, has := opts[utils.SQLDBNameOpt]; has {
		optsCfg.SQLDBName = utils.StringPointer(utils.IfaceAsString(utils.SQLDBNameOpt))
	}
	if _, has := opts[utils.PgSSLModeCfg]; has {
		optsCfg.PgSSLMode = utils.StringPointer(utils.IfaceAsString(utils.PgSSLModeCfg))
	}
	if _, has := opts[utils.KafkaTopic]; has {
		optsCfg.KafkaTopic = utils.StringPointer(utils.IfaceAsString(utils.KafkaTopic))
	}
	if _, has := opts[utils.AMQPQueueID]; has {
		optsCfg.AMQPQueueID = utils.StringPointer(utils.IfaceAsString(utils.AMQPQueueID))
	}
	if _, has := opts[utils.AMQPRoutingKey]; has {
		optsCfg.AMQPRoutingKey = utils.StringPointer(utils.IfaceAsString(utils.AMQPRoutingKey))
	}
	if _, has := opts[utils.AMQPExchange]; has {
		optsCfg.AMQPExchange = utils.StringPointer(utils.IfaceAsString(utils.AMQPExchange))
	}
	if _, has := opts[utils.AMQPExchangeType]; has {
		optsCfg.AMQPExchangeType = utils.StringPointer(utils.IfaceAsString(utils.AMQPExchangeType))
	}
	if _, has := opts[utils.AWSRegion]; has {
		optsCfg.AWSRegion = utils.StringPointer(utils.IfaceAsString(utils.AWSRegion))
	}
	if _, has := opts[utils.AWSKey]; has {
		optsCfg.AWSKey = utils.StringPointer(utils.IfaceAsString(utils.AWSKey))
	}
	if _, has := opts[utils.AWSSecret]; has {
		optsCfg.AWSSecret = utils.StringPointer(utils.IfaceAsString(utils.AWSSecret))
	}
	if _, has := opts[utils.AWSToken]; has {
		optsCfg.AWSToken = utils.StringPointer(utils.IfaceAsString(utils.AWSToken))
	}
	if _, has := opts[utils.SQSQueueID]; has {
		optsCfg.SQSQueueID = utils.StringPointer(utils.IfaceAsString(utils.SQSQueueID))
	}
	if _, has := opts[utils.S3Bucket]; has {
		optsCfg.S3BucketID = utils.StringPointer(utils.IfaceAsString(utils.S3Bucket))
	}
	if _, has := opts[utils.S3FolderPath]; has {
		optsCfg.S3FolderPath = utils.StringPointer(utils.IfaceAsString(utils.S3FolderPath))
	}
	if _, has := opts[utils.NatsJetStream]; has {
		x, err := utils.IfaceAsBool(utils.NatsJetStream)
		if err != nil {
			return nil, err
		}
		optsCfg.NATSJetStream = utils.BoolPointer(x)
	}
	if _, has := opts[utils.NatsSubject]; has {
		optsCfg.NATSSubject = utils.StringPointer(utils.IfaceAsString(utils.NatsSubject))
	}
	if _, has := opts[utils.NatsJWTFile]; has {
		optsCfg.NATSJWTFile = utils.StringPointer(utils.IfaceAsString(utils.NatsJWTFile))
	}
	if _, has := opts[utils.NatsSeedFile]; has {
		optsCfg.NATSSeedFile = utils.StringPointer(utils.IfaceAsString(utils.NatsSeedFile))
	}
	if _, has := opts[utils.NatsCertificateAuthority]; has {
		optsCfg.NATSCertificateAuthority = utils.StringPointer(utils.IfaceAsString(utils.NatsCertificateAuthority))
	}
	if _, has := opts[utils.NatsClientCertificate]; has {
		optsCfg.NATSClientCertificate = utils.StringPointer(utils.IfaceAsString(utils.NatsClientCertificate))
	}
	if _, has := opts[utils.NatsClientKey]; has {
		optsCfg.NATSClientKey = utils.StringPointer(utils.IfaceAsString(utils.NatsClientKey))
	}
	if _, has := opts[utils.NatsJetStreamMaxWait]; has {
		t, err := utils.IfaceAsDuration(utils.NatsJetStreamMaxWait)
		if err != nil {
			return nil, err
		}
		optsCfg.NATSJetStreamMaxWait = &t
	}
	if _, has := opts[utils.RpcCodec]; has {
		optsCfg.RPCCodec = utils.StringPointer(utils.IfaceAsString(utils.RpcCodec))
	}
	if _, has := opts[utils.ServiceMethod]; has {
		optsCfg.ServiceMethod = utils.StringPointer(utils.IfaceAsString(utils.ServiceMethod))
	}
	if _, has := opts[utils.KeyPath]; has {
		optsCfg.KeyPath = utils.StringPointer(utils.IfaceAsString(utils.KeyPath))
	}
	if _, has := opts[utils.CertPath]; has {
		optsCfg.CertPath = utils.StringPointer(utils.IfaceAsString(utils.CertPath))
	}
	if _, has := opts[utils.CaPath]; has {
		optsCfg.CAPath = utils.StringPointer(utils.IfaceAsString(utils.CaPath))
	}
	if _, has := opts[utils.Tls]; has {
		x, err := utils.IfaceAsBool(utils.Tls)
		if err != nil {
			return nil, err
		}
		optsCfg.TLS = utils.BoolPointer(x)
	}
	if _, has := opts[utils.ConnIDs]; has {
		optsCfg.ConnIDs = opts[utils.ConnIDs].(*[]string)
	}
	if _, has := opts[utils.RpcConnTimeout]; has {
		t, err := utils.IfaceAsDuration(utils.RpcConnTimeout)
		if err != nil {
			return nil, err
		}
		optsCfg.RPCConnTimeout = &t
	}
	if _, has := opts[utils.RpcReplyTimeout]; has {
		t, err := utils.IfaceAsDuration(utils.RpcReplyTimeout)
		if err != nil {
			return nil, err
		}
		optsCfg.RPCReplyTimeout = &t
	}
	if _, has := opts[utils.RPCAPIOpts]; has {
		optsCfg.RPCAPIOpts = opts[utils.RPCAPIOpts].(map[string]any)
	}
	return optsCfg, nil
}

// AddEvent adds one event
func (expEv *FailedExportersEEs) AddEvent(ev any) {
	expEv.lk.Lock()
	expEv.Events = append(expEv.Events, ev)
	expEv.lk.Unlock()
}

// ReplayFailedPosts tryies to post cdrs again
func (expEv *FailedExportersEEs) ReplayFailedPosts(ctx *context.Context, attempts int, tnt string) (err error) {
	failedEvents := &FailedExportersEEs{
		Path:   expEv.Path,
		Opts:   expEv.Opts,
		Format: expEv.Format,
	}

	eeCfg := config.NewEventExporterCfg("ReplayFailedPosts", expEv.Format, expEv.Path, utils.MetaNone,
		attempts, expEv.Opts)
	var ee ees.EventExporter
	if ee, err = ees.NewEventExporter(eeCfg, config.CgrConfig(), nil, nil); err != nil {
		return
	}
	keyFunc := func() string { return utils.EmptyString }
	if expEv.Format == utils.MetaKafkajsonMap || expEv.Format == utils.MetaS3jsonMap {
		keyFunc = utils.UUIDSha1Prefix
	}
	for _, ev := range expEv.Events {
		if err = ees.ExportWithAttempts(context.Background(), ee, ev, keyFunc(), expEv.connMngr, tnt); err != nil {
			failedEvents.AddEvent(ev)
		}
	}
	ee.Close()
	switch len(failedEvents.Events) {
	case 0: // none failed to be replayed
		return nil
	case len(expEv.Events): // all failed, return last encountered error
		return err
	default:
		return utils.ErrPartiallyExecuted
	}
}
