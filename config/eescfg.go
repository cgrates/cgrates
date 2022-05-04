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

package config

import (
	"reflect"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// EEsCfg the config for Event Exporters
type EEsCfg struct {
	Enabled         bool
	AttributeSConns []string
	Cache           map[string]*CacheParamCfg
	Exporters       []*EventExporterCfg
}

// GetDefaultExporter returns the exporter with the *default id
func (eeS *EEsCfg) GetDefaultExporter() *EventExporterCfg {
	for _, es := range eeS.Exporters {
		if es.ID == utils.MetaDefault {
			return es
		}
	}
	return nil
}

// loadEesCfg loads the Ees section of the configuration
func (eeS *EEsCfg) Load(ctx *context.Context, jsnCfg ConfigDB, cfg *CGRConfig) (err error) {
	jsnEEsCfg := new(EEsJsonCfg)
	if err = jsnCfg.GetSection(ctx, EEsJSON, jsnEEsCfg); err != nil {
		return
	}
	return eeS.loadFromJSONCfg(jsnEEsCfg, cfg.templates, cfg.generalCfg.RSRSep)
}

func (eeS *EEsCfg) loadFromJSONCfg(jsnCfg *EEsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		eeS.Enabled = *jsnCfg.Enabled
	}
	for kJsn, vJsn := range jsnCfg.Cache {
		val := new(CacheParamCfg)
		if err := val.loadFromJSONCfg(vJsn); err != nil {
			return err
		}
		eeS.Cache[kJsn] = val
	}
	if jsnCfg.Attributes_conns != nil {
		eeS.AttributeSConns = updateInternalConns(*jsnCfg.Attributes_conns, utils.MetaAttributes)
	}
	return eeS.appendEEsExporters(jsnCfg.Exporters, msgTemplates, sep)
}

func (eeS *EEsCfg) appendEEsExporters(exporters *[]*EventExporterJsonCfg, msgTemplates map[string][]*FCTemplate, separator string) (err error) {
	if exporters == nil {
		return
	}
	for _, jsnExp := range *exporters {
		var exp *EventExporterCfg
		if jsnExp.Id != nil {
			for _, exporter := range eeS.Exporters {
				if exporter.ID == *jsnExp.Id {
					exp = exporter
					break
				}
			}
		}
		if exp == nil {
			exp = getDftEvExpCfg()
			eeS.Exporters = append(eeS.Exporters, exp)
		}
		if err = exp.loadFromJSONCfg(jsnExp, msgTemplates, separator); err != nil {
			return
		}
	}
	return
}

func (EEsCfg) SName() string             { return EEsJSON }
func (eeS EEsCfg) CloneSection() Section { return eeS.Clone() }

// Clone returns a deep copy of EEsCfg
func (eeS EEsCfg) Clone() (cln *EEsCfg) {
	cln = &EEsCfg{
		Enabled:   eeS.Enabled,
		Cache:     make(map[string]*CacheParamCfg),
		Exporters: make([]*EventExporterCfg, len(eeS.Exporters)),
	}
	if eeS.AttributeSConns != nil {
		cln.AttributeSConns = utils.CloneStringSlice(eeS.AttributeSConns)
	}
	for key, value := range eeS.Cache {
		cln.Cache[key] = value.Clone()
	}
	for idx, exp := range eeS.Exporters {
		cln.Exporters[idx] = exp.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (eeS EEsCfg) AsMapInterface(separator string) interface{} {
	mp := map[string]interface{}{
		utils.EnabledCfg: eeS.Enabled,
	}
	if eeS.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(eeS.AttributeSConns)
	}
	if eeS.Cache != nil {
		cache := make(map[string]interface{}, len(eeS.Cache))
		for key, value := range eeS.Cache {
			cache[key] = value.AsMapInterface()
		}
		mp[utils.CacheCfg] = cache
	}
	if eeS.Exporters != nil {
		exporters := make([]map[string]interface{}, len(eeS.Exporters))
		for i, item := range eeS.Exporters {
			exporters[i] = item.AsMapInterface(separator)
		}
		mp[utils.ExportersCfg] = exporters
	}
	return mp
}

type EventExporterOpts struct {
	CSVFieldSeparator        *string
	ElsIndex                 *string
	ElsIfPrimaryTerm         *int
	ElsIfSeqNo               *int
	ElsOpType                *string
	ElsPipeline              *string
	ElsRouting               *string
	ElsTimeout               *time.Duration
	ElsVersion               *int
	ElsVersionType           *string
	ElsWaitForActiveShards   *string
	SQLMaxIdleConns          *int
	SQLMaxOpenConns          *int
	SQLConnMaxLifetime       *time.Duration
	MYSQLDSNParams           map[string]string
	SQLTableName             *string
	SQLDBName                *string
	SSLMode                  *string
	KafkaTopic               *string
	AMQPRoutingKey           *string
	AMQPQueueID              *string
	AMQPExchange             *string
	AMQPExchangeType         *string
	AWSRegion                *string
	AWSKey                   *string
	AWSSecret                *string
	AWSToken                 *string
	SQSQueueID               *string
	S3BucketID               *string
	S3FolderPath             *string
	NATSJetStream            *bool
	NATSSubject              *string
	NATSJWTFile              *string
	NATSSeedFile             *string
	NATSCertificateAuthority *string
	NATSClientCertificate    *string
	NATSClientKey            *string
	NATSJetStreamMaxWait     *time.Duration
	RPCCodec                 *string
	ServiceMethod            *string
	KeyPath                  *string
	CertPath                 *string
	CAPath                   *string
	TLS                      *bool
	ConnIDs                  *[]string
	RPCConnTimeout           *time.Duration
	RPCReplyTimeout          *time.Duration
	RPCAPIOpts               map[string]interface{}
}

// EventExporterCfg the config for a Event Exporter
type EventExporterCfg struct {
	ID                 string
	Type               string
	ExportPath         string
	Opts               *EventExporterOpts
	Timezone           string
	Filters            []string
	Flags              utils.FlagsWithParams
	AttributeSIDs      []string // selective AttributeS profiles
	AttributeSCtx      string   // context to use when querying AttributeS
	Synchronous        bool
	Blocker            bool
	Attempts           int
	FailedPostsDir     string
	ConcurrentRequests int
	Fields             []*FCTemplate
	headerFields       []*FCTemplate
	contentFields      []*FCTemplate
	trailerFields      []*FCTemplate
}

func (eeOpts *EventExporterOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.CSVFieldSeparator != nil {
		eeOpts.CSVFieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if jsnCfg.ElsIndex != nil {
		eeOpts.ElsIndex = jsnCfg.ElsIndex
	}
	if jsnCfg.ElsIfPrimaryTerm != nil {
		eeOpts.ElsIfPrimaryTerm = jsnCfg.ElsIfPrimaryTerm
	}
	if jsnCfg.ElsIfSeqNo != nil {
		eeOpts.ElsIfSeqNo = jsnCfg.ElsIfSeqNo
	}
	if jsnCfg.ElsOpType != nil {
		eeOpts.ElsOpType = jsnCfg.ElsOpType
	}
	if jsnCfg.ElsPipeline != nil {
		eeOpts.ElsPipeline = jsnCfg.ElsPipeline
	}
	if jsnCfg.ElsRouting != nil {
		eeOpts.ElsRouting = jsnCfg.ElsRouting
	}
	if jsnCfg.ElsTimeout != nil {
		var elsTimeout time.Duration
		if elsTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ElsTimeout); err != nil {
			return
		}
		eeOpts.ElsTimeout = utils.DurationPointer(elsTimeout)
	}
	if jsnCfg.ElsVersion != nil {
		eeOpts.ElsVersion = jsnCfg.ElsVersion
	}
	if jsnCfg.ElsVersionType != nil {
		eeOpts.ElsVersionType = jsnCfg.ElsVersionType
	}
	if jsnCfg.ElsWaitForActiveShards != nil {
		eeOpts.ElsWaitForActiveShards = jsnCfg.ElsWaitForActiveShards
	}
	if jsnCfg.SQLMaxIdleConns != nil {
		eeOpts.SQLMaxIdleConns = jsnCfg.SQLMaxIdleConns
	}
	if jsnCfg.SQLMaxOpenConns != nil {
		eeOpts.SQLMaxOpenConns = jsnCfg.SQLMaxOpenConns
	}
	if jsnCfg.SQLConnMaxLifetime != nil {
		var sqlConnMaxLifetime time.Duration
		if sqlConnMaxLifetime, err = utils.ParseDurationWithNanosecs(*jsnCfg.SQLConnMaxLifetime); err != nil {
			return
		}
		eeOpts.SQLConnMaxLifetime = utils.DurationPointer(sqlConnMaxLifetime)
	}
	if jsnCfg.MYSQLDSNParams != nil {
		eeOpts.MYSQLDSNParams = make(map[string]string)
		eeOpts.MYSQLDSNParams = jsnCfg.MYSQLDSNParams
	}
	if jsnCfg.SQLTableName != nil {
		eeOpts.SQLTableName = jsnCfg.SQLTableName
	}
	if jsnCfg.SQLDBName != nil {
		eeOpts.SQLDBName = jsnCfg.SQLDBName
	}
	if jsnCfg.SSLMode != nil {
		eeOpts.SSLMode = jsnCfg.SSLMode
	}
	if jsnCfg.KafkaTopic != nil {
		eeOpts.KafkaTopic = jsnCfg.KafkaTopic
	}
	if jsnCfg.AMQPQueueID != nil {
		eeOpts.AMQPQueueID = jsnCfg.AMQPQueueID
	}
	if jsnCfg.AMQPRoutingKey != nil {
		eeOpts.AMQPRoutingKey = jsnCfg.AMQPRoutingKey
	}
	if jsnCfg.AMQPExchange != nil {
		eeOpts.AMQPExchange = jsnCfg.AMQPExchange
	}
	if jsnCfg.AMQPExchangeType != nil {
		eeOpts.AMQPExchangeType = jsnCfg.AMQPExchangeType
	}
	if jsnCfg.AWSRegion != nil {
		eeOpts.AWSRegion = jsnCfg.AWSRegion
	}
	if jsnCfg.AWSKey != nil {
		eeOpts.AWSKey = jsnCfg.AWSKey
	}
	if jsnCfg.AWSSecret != nil {
		eeOpts.AWSSecret = jsnCfg.AWSSecret
	}
	if jsnCfg.AWSToken != nil {
		eeOpts.AWSToken = jsnCfg.AWSToken
	}
	if jsnCfg.SQSQueueID != nil {
		eeOpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.S3BucketID != nil {
		eeOpts.S3BucketID = jsnCfg.S3BucketID
	}
	if jsnCfg.S3FolderPath != nil {
		eeOpts.S3FolderPath = jsnCfg.S3FolderPath
	}
	if jsnCfg.NATSJetStream != nil {
		eeOpts.NATSJetStream = jsnCfg.NATSJetStream
	}
	if jsnCfg.NATSSubject != nil {
		eeOpts.NATSSubject = jsnCfg.NATSSubject
	}
	if jsnCfg.NATSJWTFile != nil {
		eeOpts.NATSJWTFile = jsnCfg.NATSJWTFile
	}
	if jsnCfg.NATSSeedFile != nil {
		eeOpts.NATSSeedFile = jsnCfg.NATSSeedFile
	}
	if jsnCfg.NATSCertificateAuthority != nil {
		eeOpts.NATSCertificateAuthority = jsnCfg.NATSCertificateAuthority
	}
	if jsnCfg.NATSClientCertificate != nil {
		eeOpts.NATSClientCertificate = jsnCfg.NATSClientCertificate
	}
	if jsnCfg.NATSClientKey != nil {
		eeOpts.NATSClientKey = jsnCfg.NATSClientKey
	}
	if jsnCfg.NATSJetStreamMaxWait != nil {
		var natsJetStreamMaxWait time.Duration
		if natsJetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWait); err != nil {
			return
		}
		eeOpts.NATSJetStreamMaxWait = utils.DurationPointer(natsJetStreamMaxWait)
	}
	if jsnCfg.RPCCodec != nil {
		eeOpts.RPCCodec = jsnCfg.RPCCodec
	}
	if jsnCfg.ServiceMethod != nil {
		eeOpts.ServiceMethod = jsnCfg.ServiceMethod
	}
	if jsnCfg.KeyPath != nil {
		eeOpts.KeyPath = jsnCfg.KeyPath
	}
	if jsnCfg.CertPath != nil {
		eeOpts.CertPath = jsnCfg.CertPath
	}
	if jsnCfg.CAPath != nil {
		eeOpts.CAPath = jsnCfg.CAPath
	}
	if jsnCfg.TLS != nil {
		eeOpts.TLS = jsnCfg.TLS
	}
	if jsnCfg.ConnIDs != nil {
		eeOpts.ConnIDs = jsnCfg.ConnIDs
	}
	if jsnCfg.RPCConnTimeout != nil {
		var rpcConnTimeout time.Duration
		if rpcConnTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RPCConnTimeout); err != nil {
			return
		}
		eeOpts.RPCConnTimeout = utils.DurationPointer(rpcConnTimeout)
	}
	if jsnCfg.RPCReplyTimeout != nil {
		var rpcReplyTimeout time.Duration
		if rpcReplyTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RPCReplyTimeout); err != nil {
			return
		}
		eeOpts.RPCReplyTimeout = utils.DurationPointer(rpcReplyTimeout)
	}
	if jsnCfg.RPCAPIOpts != nil {
		eeOpts.RPCAPIOpts = make(map[string]interface{})
		eeOpts.RPCAPIOpts = jsnCfg.RPCAPIOpts
	}
	return
}

func (eeC *EventExporterCfg) loadFromJSONCfg(jsnEec *EventExporterJsonCfg, msgTemplates map[string][]*FCTemplate, separator string) (err error) {
	if jsnEec == nil {
		return
	}
	if jsnEec.Id != nil {
		eeC.ID = *jsnEec.Id
	}
	if jsnEec.Type != nil {
		eeC.Type = *jsnEec.Type
	}
	if jsnEec.Export_path != nil {
		eeC.ExportPath = *jsnEec.Export_path
	}
	if jsnEec.Timezone != nil {
		eeC.Timezone = *jsnEec.Timezone
	}
	if jsnEec.Filters != nil {
		eeC.Filters = utils.CloneStringSlice(*jsnEec.Filters)
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = utils.CloneStringSlice(*jsnEec.Attribute_ids)
	}
	if jsnEec.Synchronous != nil {
		eeC.Synchronous = *jsnEec.Synchronous
	}
	if jsnEec.Blocker != nil {
		eeC.Blocker = *jsnEec.Blocker
	}
	if jsnEec.Attempts != nil {
		eeC.Attempts = *jsnEec.Attempts
	}
	if jsnEec.Concurrent_requests != nil {
		eeC.ConcurrentRequests = *jsnEec.Concurrent_requests
	}
	if jsnEec.Fields != nil {
		eeC.Fields, err = FCTemplatesFromFCTemplatesJSONCfg(*jsnEec.Fields, separator)
		if err != nil {
			return
		}
		if tpls, err := InflateTemplates(eeC.Fields, msgTemplates); err != nil {
			return err
		} else if tpls != nil {
			eeC.Fields = tpls
		}
		eeC.ComputeFields()
	}
	if jsnEec.Failed_posts_dir != nil {
		eeC.FailedPostsDir = *jsnEec.Failed_posts_dir
	}
	if jsnEec.Opts != nil {
		err = eeC.Opts.loadFromJSONCfg(jsnEec.Opts)
	}
	return
}

// ComputeFields will split the fields in header trailer or content
// exported for ees testing
func (eeC *EventExporterCfg) ComputeFields() {
	eeC.headerFields = make([]*FCTemplate, 0)
	eeC.contentFields = make([]*FCTemplate, 0)
	eeC.trailerFields = make([]*FCTemplate, 0)
	for _, field := range eeC.Fields {
		switch field.GetPathSlice()[0] {
		case utils.MetaHdr:
			eeC.headerFields = append(eeC.headerFields, field)
		case utils.MetaExp, utils.MetaUCH:
			eeC.contentFields = append(eeC.contentFields, field)
		case utils.MetaTrl:
			eeC.trailerFields = append(eeC.trailerFields, field)
		}
	}
}

// HeaderFields returns the fields that have *hdr prefix
func (eeC *EventExporterCfg) HeaderFields() []*FCTemplate {
	return eeC.headerFields
}

// ContentFields returns the fields that do not have *hdr or *trl prefix
func (eeC *EventExporterCfg) ContentFields() []*FCTemplate {
	return eeC.contentFields
}

// TrailerFields returns the fields that have *trl prefix
func (eeC *EventExporterCfg) TrailerFields() []*FCTemplate {
	return eeC.trailerFields
}

func (eeOpts *EventExporterOpts) Clone() *EventExporterOpts {
	cln := &EventExporterOpts{}
	if eeOpts.CSVFieldSeparator != nil {
		cln.CSVFieldSeparator = utils.StringPointer(*eeOpts.CSVFieldSeparator)
	}
	if eeOpts.ElsIndex != nil {
		cln.ElsIndex = utils.StringPointer(*eeOpts.ElsIndex)
	}
	if eeOpts.ElsIfPrimaryTerm != nil {
		cln.ElsIfPrimaryTerm = utils.IntPointer(*eeOpts.ElsIfPrimaryTerm)
	}
	if eeOpts.ElsIfSeqNo != nil {
		cln.ElsIfSeqNo = utils.IntPointer(*eeOpts.ElsIfSeqNo)
	}
	if eeOpts.ElsOpType != nil {
		cln.ElsOpType = utils.StringPointer(*eeOpts.ElsOpType)
	}
	if eeOpts.ElsPipeline != nil {
		cln.ElsPipeline = utils.StringPointer(*eeOpts.ElsPipeline)
	}
	if eeOpts.ElsRouting != nil {
		cln.ElsRouting = utils.StringPointer(*eeOpts.ElsRouting)
	}
	if eeOpts.ElsTimeout != nil {
		cln.ElsTimeout = utils.DurationPointer(*eeOpts.ElsTimeout)
	}
	if eeOpts.ElsVersion != nil {
		cln.ElsVersion = utils.IntPointer(*eeOpts.ElsVersion)
	}
	if eeOpts.ElsVersionType != nil {
		cln.ElsVersionType = utils.StringPointer(*eeOpts.ElsVersionType)
	}
	if eeOpts.ElsWaitForActiveShards != nil {
		cln.ElsWaitForActiveShards = utils.StringPointer(*eeOpts.ElsWaitForActiveShards)
	}
	if eeOpts.SQLMaxIdleConns != nil {
		cln.SQLMaxIdleConns = utils.IntPointer(*eeOpts.SQLMaxIdleConns)
	}
	if eeOpts.SQLMaxOpenConns != nil {
		cln.SQLMaxOpenConns = utils.IntPointer(*eeOpts.SQLMaxOpenConns)
	}
	if eeOpts.SQLConnMaxLifetime != nil {
		cln.SQLConnMaxLifetime = utils.DurationPointer(*eeOpts.SQLConnMaxLifetime)
	}
	if eeOpts.MYSQLDSNParams != nil {
		cln.MYSQLDSNParams = eeOpts.MYSQLDSNParams
	}
	if eeOpts.SQLTableName != nil {
		cln.SQLTableName = utils.StringPointer(*eeOpts.SQLTableName)
	}
	if eeOpts.SQLDBName != nil {
		cln.SQLDBName = utils.StringPointer(*eeOpts.SQLDBName)
	}
	if eeOpts.SSLMode != nil {
		cln.SSLMode = utils.StringPointer(*eeOpts.SSLMode)
	}
	if eeOpts.KafkaTopic != nil {
		cln.KafkaTopic = utils.StringPointer(*eeOpts.KafkaTopic)
	}
	if eeOpts.AMQPQueueID != nil {
		cln.AMQPQueueID = utils.StringPointer(*eeOpts.AMQPQueueID)
	}
	if eeOpts.AMQPRoutingKey != nil {
		cln.AMQPRoutingKey = utils.StringPointer(*eeOpts.AMQPRoutingKey)
	}
	if eeOpts.AMQPExchange != nil {
		cln.AMQPExchange = utils.StringPointer(*eeOpts.AMQPExchange)
	}
	if eeOpts.AMQPExchangeType != nil {
		cln.AMQPExchangeType = utils.StringPointer(*eeOpts.AMQPExchangeType)
	}
	if eeOpts.AWSRegion != nil {
		cln.AWSRegion = utils.StringPointer(*eeOpts.AWSRegion)
	}
	if eeOpts.AWSKey != nil {
		cln.AWSKey = utils.StringPointer(*eeOpts.AWSKey)
	}
	if eeOpts.AWSSecret != nil {
		cln.AWSSecret = utils.StringPointer(*eeOpts.AWSSecret)
	}
	if eeOpts.AWSToken != nil {
		cln.AWSToken = utils.StringPointer(*eeOpts.AWSToken)
	}
	if eeOpts.SQSQueueID != nil {
		cln.SQSQueueID = utils.StringPointer(*eeOpts.SQSQueueID)
	}
	if eeOpts.S3BucketID != nil {
		cln.S3BucketID = utils.StringPointer(*eeOpts.S3BucketID)
	}
	if eeOpts.S3FolderPath != nil {
		cln.S3FolderPath = utils.StringPointer(*eeOpts.S3FolderPath)
	}
	if eeOpts.NATSJetStream != nil {
		cln.NATSJetStream = utils.BoolPointer(*eeOpts.NATSJetStream)
	}
	if eeOpts.NATSSubject != nil {
		cln.NATSSubject = utils.StringPointer(*eeOpts.NATSSubject)
	}
	if eeOpts.NATSJWTFile != nil {
		cln.NATSJWTFile = utils.StringPointer(*eeOpts.NATSJWTFile)
	}
	if eeOpts.NATSSeedFile != nil {
		cln.NATSSeedFile = utils.StringPointer(*eeOpts.NATSSeedFile)
	}
	if eeOpts.NATSCertificateAuthority != nil {
		cln.NATSCertificateAuthority = utils.StringPointer(*eeOpts.NATSCertificateAuthority)
	}
	if eeOpts.NATSClientCertificate != nil {
		cln.NATSClientCertificate = utils.StringPointer(*eeOpts.NATSClientCertificate)
	}
	if eeOpts.NATSClientKey != nil {
		cln.NATSClientKey = utils.StringPointer(*eeOpts.NATSClientKey)
	}
	if eeOpts.NATSJetStreamMaxWait != nil {
		cln.NATSJetStreamMaxWait = utils.DurationPointer(*eeOpts.NATSJetStreamMaxWait)
	}
	if eeOpts.RPCCodec != nil {
		cln.RPCCodec = utils.StringPointer(*eeOpts.RPCCodec)
	}
	if eeOpts.ServiceMethod != nil {
		cln.ServiceMethod = utils.StringPointer(*eeOpts.ServiceMethod)
	}
	if eeOpts.KeyPath != nil {
		cln.KeyPath = utils.StringPointer(*eeOpts.KeyPath)
	}
	if eeOpts.CertPath != nil {
		cln.CertPath = utils.StringPointer(*eeOpts.CertPath)
	}
	if eeOpts.CAPath != nil {
		cln.CAPath = utils.StringPointer(*eeOpts.CAPath)
	}
	if eeOpts.TLS != nil {
		cln.TLS = utils.BoolPointer(*eeOpts.TLS)
	}
	if eeOpts.ConnIDs != nil {
		cln.ConnIDs = utils.SliceStringPointer(*eeOpts.ConnIDs)
	}
	if eeOpts.RPCConnTimeout != nil {
		cln.RPCConnTimeout = utils.DurationPointer(*eeOpts.RPCConnTimeout)
	}
	if eeOpts.RPCReplyTimeout != nil {
		cln.RPCReplyTimeout = utils.DurationPointer(*eeOpts.RPCReplyTimeout)
	}
	if eeOpts.RPCAPIOpts != nil {
		cln.RPCAPIOpts = make(map[string]interface{})
		cln.RPCAPIOpts = eeOpts.RPCAPIOpts
	}
	return cln
}

// Clone returns a deep copy of EventExporterCfg
func (eeC EventExporterCfg) Clone() (cln *EventExporterCfg) {
	cln = &EventExporterCfg{
		ID:                 eeC.ID,
		Type:               eeC.Type,
		ExportPath:         eeC.ExportPath,
		Timezone:           eeC.Timezone,
		Flags:              eeC.Flags.Clone(),
		AttributeSCtx:      eeC.AttributeSCtx,
		Synchronous:        eeC.Synchronous,
		Blocker:            eeC.Blocker,
		Attempts:           eeC.Attempts,
		ConcurrentRequests: eeC.ConcurrentRequests,
		Fields:             make([]*FCTemplate, len(eeC.Fields)),
		headerFields:       make([]*FCTemplate, len(eeC.headerFields)),
		contentFields:      make([]*FCTemplate, len(eeC.contentFields)),
		trailerFields:      make([]*FCTemplate, len(eeC.trailerFields)),
		Opts:               eeC.Opts.Clone(),
		FailedPostsDir:     eeC.FailedPostsDir,
	}

	if eeC.Filters != nil {
		cln.Filters = utils.CloneStringSlice(eeC.Filters)
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = utils.CloneStringSlice(eeC.AttributeSIDs)
	}

	for idx, fld := range eeC.Fields {
		cln.Fields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.headerFields {
		cln.headerFields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.contentFields {
		cln.contentFields[idx] = fld.Clone()
	}
	for idx, fld := range eeC.trailerFields {
		cln.trailerFields[idx] = fld.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]interface{}
func (eeC *EventExporterCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	opts := map[string]interface{}{}
	if eeC.Opts.CSVFieldSeparator != nil {
		opts[utils.CSVFieldSepOpt] = *eeC.Opts.CSVFieldSeparator
	}
	if eeC.Opts.ElsIndex != nil {
		opts[utils.ElsIndex] = *eeC.Opts.ElsIndex
	}
	if eeC.Opts.ElsIfPrimaryTerm != nil {
		opts[utils.ElsIfPrimaryTerm] = *eeC.Opts.ElsIfPrimaryTerm
	}
	if eeC.Opts.ElsIfSeqNo != nil {
		opts[utils.ElsIfSeqNo] = *eeC.Opts.ElsIfSeqNo
	}
	if eeC.Opts.ElsOpType != nil {
		opts[utils.ElsOpType] = *eeC.Opts.ElsOpType
	}
	if eeC.Opts.ElsPipeline != nil {
		opts[utils.ElsPipeline] = *eeC.Opts.ElsPipeline
	}
	if eeC.Opts.ElsRouting != nil {
		opts[utils.ElsRouting] = *eeC.Opts.ElsRouting
	}
	if eeC.Opts.ElsTimeout != nil {
		opts[utils.ElsTimeout] = eeC.Opts.ElsTimeout.String()
	}
	if eeC.Opts.ElsVersion != nil {
		opts[utils.ElsVersionLow] = *eeC.Opts.ElsVersion
	}
	if eeC.Opts.ElsVersionType != nil {
		opts[utils.ElsVersionType] = *eeC.Opts.ElsVersionType
	}
	if eeC.Opts.ElsWaitForActiveShards != nil {
		opts[utils.ElsWaitForActiveShards] = *eeC.Opts.ElsWaitForActiveShards
	}
	if eeC.Opts.SQLMaxIdleConns != nil {
		opts[utils.SQLMaxIdleConnsCfg] = *eeC.Opts.SQLMaxIdleConns
	}
	if eeC.Opts.SQLMaxOpenConns != nil {
		opts[utils.SQLMaxOpenConns] = *eeC.Opts.SQLMaxOpenConns
	}
	if eeC.Opts.SQLConnMaxLifetime != nil {
		opts[utils.SQLConnMaxLifetime] = eeC.Opts.SQLConnMaxLifetime.String()
	}
	if eeC.Opts.MYSQLDSNParams != nil {
		opts[utils.MYSQLDSNParams] = eeC.Opts.MYSQLDSNParams
	}
	if eeC.Opts.SQLTableName != nil {
		opts[utils.SQLTableNameOpt] = *eeC.Opts.SQLTableName
	}
	if eeC.Opts.SQLDBName != nil {
		opts[utils.SQLDBNameOpt] = *eeC.Opts.SQLDBName
	}
	if eeC.Opts.SSLMode != nil {
		opts[utils.SSLModeCfg] = *eeC.Opts.SSLMode
	}
	if eeC.Opts.KafkaTopic != nil {
		opts[utils.KafkaTopic] = *eeC.Opts.KafkaTopic
	}
	if eeC.Opts.AMQPQueueID != nil {
		opts[utils.AMQPQueueID] = *eeC.Opts.AMQPQueueID
	}
	if eeC.Opts.AMQPRoutingKey != nil {
		opts[utils.AMQPRoutingKey] = *eeC.Opts.AMQPRoutingKey
	}
	if eeC.Opts.AMQPExchange != nil {
		opts[utils.AMQPExchange] = *eeC.Opts.AMQPExchange
	}
	if eeC.Opts.AMQPExchangeType != nil {
		opts[utils.AMQPExchangeType] = *eeC.Opts.AMQPExchangeType
	}
	if eeC.Opts.AWSRegion != nil {
		opts[utils.AWSRegion] = *eeC.Opts.AWSRegion
	}
	if eeC.Opts.AWSKey != nil {
		opts[utils.AWSKey] = *eeC.Opts.AWSKey
	}
	if eeC.Opts.AWSSecret != nil {
		opts[utils.AWSSecret] = *eeC.Opts.AWSSecret
	}
	if eeC.Opts.AWSToken != nil {
		opts[utils.AWSToken] = *eeC.Opts.AWSToken
	}
	if eeC.Opts.SQSQueueID != nil {
		opts[utils.SQSQueueID] = *eeC.Opts.SQSQueueID
	}
	if eeC.Opts.S3BucketID != nil {
		opts[utils.S3Bucket] = *eeC.Opts.S3BucketID
	}
	if eeC.Opts.S3FolderPath != nil {
		opts[utils.S3FolderPath] = *eeC.Opts.S3FolderPath
	}
	if eeC.Opts.NATSJetStream != nil {
		opts[utils.NatsJetStream] = *eeC.Opts.NATSJetStream
	}
	if eeC.Opts.NATSSubject != nil {
		opts[utils.NatsSubject] = *eeC.Opts.NATSSubject
	}
	if eeC.Opts.NATSJWTFile != nil {
		opts[utils.NatsJWTFile] = *eeC.Opts.NATSJWTFile
	}
	if eeC.Opts.NATSSeedFile != nil {
		opts[utils.NatsSeedFile] = *eeC.Opts.NATSSeedFile
	}
	if eeC.Opts.NATSCertificateAuthority != nil {
		opts[utils.NatsCertificateAuthority] = *eeC.Opts.NATSCertificateAuthority
	}
	if eeC.Opts.NATSClientCertificate != nil {
		opts[utils.NatsClientCertificate] = *eeC.Opts.NATSClientCertificate
	}
	if eeC.Opts.NATSClientKey != nil {
		opts[utils.NatsClientKey] = *eeC.Opts.NATSClientKey
	}
	if eeC.Opts.NATSJetStreamMaxWait != nil {
		opts[utils.NatsJetStreamMaxWait] = eeC.Opts.NATSJetStreamMaxWait.String()
	}
	if eeC.Opts.RPCCodec != nil {
		opts[utils.RpcCodec] = *eeC.Opts.RPCCodec
	}
	if eeC.Opts.ServiceMethod != nil {
		opts[utils.ServiceMethod] = *eeC.Opts.ServiceMethod
	}
	if eeC.Opts.KeyPath != nil {
		opts[utils.KeyPath] = *eeC.Opts.KeyPath
	}
	if eeC.Opts.CertPath != nil {
		opts[utils.CertPath] = *eeC.Opts.CertPath
	}
	if eeC.Opts.CAPath != nil {
		opts[utils.CaPath] = *eeC.Opts.CAPath
	}
	if eeC.Opts.TLS != nil {
		opts[utils.Tls] = *eeC.Opts.TLS
	}
	if eeC.Opts.ConnIDs != nil {
		opts[utils.ConnIDs] = *eeC.Opts.ConnIDs
	}
	if eeC.Opts.RPCConnTimeout != nil {
		opts[utils.RpcConnTimeout] = eeC.Opts.RPCConnTimeout.String()
	}
	if eeC.Opts.RPCReplyTimeout != nil {
		opts[utils.RpcReplyTimeout] = eeC.Opts.RPCReplyTimeout.String()
	}
	if eeC.Opts.RPCAPIOpts != nil {
		opts[utils.RPCAPIOpts] = eeC.Opts.RPCAPIOpts
	}

	flgs := eeC.Flags.SliceFlags()
	if flgs == nil {
		flgs = []string{}
	}
	initialMP = map[string]interface{}{
		utils.IDCfg:                 eeC.ID,
		utils.TypeCfg:               eeC.Type,
		utils.ExportPathCfg:         eeC.ExportPath,
		utils.TimezoneCfg:           eeC.Timezone,
		utils.FiltersCfg:            eeC.Filters,
		utils.FlagsCfg:              flgs,
		utils.AttributeContextCfg:   eeC.AttributeSCtx,
		utils.AttributeIDsCfg:       eeC.AttributeSIDs,
		utils.SynchronousCfg:        eeC.Synchronous,
		utils.BlockerCfg:            eeC.Blocker,
		utils.AttemptsCfg:           eeC.Attempts,
		utils.ConcurrentRequestsCfg: eeC.ConcurrentRequests,
		utils.FailedPostsDirCfg:     eeC.FailedPostsDir,
		utils.OptsCfg:               opts,
	}
	if eeC.Fields != nil {
		fields := make([]map[string]interface{}, 0, len(eeC.Fields))
		for _, fld := range eeC.Fields {
			fields = append(fields, fld.AsMapInterface(separator))
		}
		initialMP[utils.FieldsCfg] = fields
	}
	return
}

type EventExporterOptsJson struct {
	CSVFieldSeparator        *string                `json:"csvFieldSeparator"`
	ElsIndex                 *string                `json:"elsIndex"`
	ElsIfPrimaryTerm         *int                   `json:"elsIfPrimaryTerm"`
	ElsIfSeqNo               *int                   `json:"elsIfSeqNo"`
	ElsOpType                *string                `json:"elsOpType"`
	ElsPipeline              *string                `json:"elsPipeline"`
	ElsRouting               *string                `json:"elsRouting"`
	ElsTimeout               *string                `json:"elsTimeout"`
	ElsVersion               *int                   `json:"elsVersion"`
	ElsVersionType           *string                `json:"elsVersionType"`
	ElsWaitForActiveShards   *string                `json:"elsWaitForActiveShards"`
	SQLMaxIdleConns          *int                   `json:"sqlMaxIdleConns"`
	SQLMaxOpenConns          *int                   `json:"sqlMaxOpenConns"`
	SQLConnMaxLifetime       *string                `json:"sqlConnMaxLifetime"`
	MYSQLDSNParams           map[string]string      `json:"mysqlDSNParams"`
	SQLTableName             *string                `json:"sqlTableName"`
	SQLDBName                *string                `json:"sqlDBName"`
	SSLMode                  *string                `json:"sslMode"`
	KafkaTopic               *string                `json:"kafkaTopic"`
	AMQPQueueID              *string                `json:"amqpQueueID"`
	AMQPRoutingKey           *string                `json:"amqpRoutingKey"`
	AMQPExchange             *string                `json:"amqpExchange"`
	AMQPExchangeType         *string                `json:"amqpExchangeType"`
	AWSRegion                *string                `json:"awsRegion"`
	AWSKey                   *string                `json:"awsKey"`
	AWSSecret                *string                `json:"awsSecret"`
	AWSToken                 *string                `json:"awsToken"`
	SQSQueueID               *string                `json:"sqsQueueID"`
	S3BucketID               *string                `json:"s3BucketID"`
	S3FolderPath             *string                `json:"s3FolderPath"`
	NATSJetStream            *bool                  `json:"natsJetStream"`
	NATSSubject              *string                `json:"natsSubject"`
	NATSJWTFile              *string                `json:"natsJWTFile"`
	NATSSeedFile             *string                `json:"natsSeedFile"`
	NATSCertificateAuthority *string                `json:"natsCertificateAuthority"`
	NATSClientCertificate    *string                `json:"natsClientCertificate"`
	NATSClientKey            *string                `json:"natsClientKey"`
	NATSJetStreamMaxWait     *string                `json:"natsJetStreamMaxWait"`
	RPCCodec                 *string                `json:"rpcCodec"`
	ServiceMethod            *string                `json:"serviceMethod"`
	KeyPath                  *string                `json:"keyPath"`
	CertPath                 *string                `json:"certPath"`
	CAPath                   *string                `json:"caPath"`
	ConnIDs                  *[]string              `json:"connIDs"`
	TLS                      *bool                  `json:"tls"`
	RPCConnTimeout           *string                `json:"rpcConnTimeout"`
	RPCReplyTimeout          *string                `json:"rpcReplyTimeout"`
	RPCAPIOpts               map[string]interface{} `json:"rpcAPIOpts"`
}

// EventExporterJsonCfg is the configuration of a single EventExporter
type EventExporterJsonCfg struct {
	Id                  *string
	Type                *string
	Export_path         *string
	Opts                *EventExporterOptsJson
	Timezone            *string
	Filters             *[]string
	Flags               *[]string
	Attribute_ids       *[]string
	Attribute_context   *string
	Synchronous         *bool
	Blocker             *bool
	Attempts            *int
	Concurrent_requests *int
	Failed_posts_dir    *string
	Fields              *[]*FcTemplateJsonCfg
}

func diffEventExporterOptsJsonCfg(d *EventExporterOptsJson, v1, v2 *EventExporterOpts) *EventExporterOptsJson {
	if d == nil {
		d = new(EventExporterOptsJson)
	}
	if v2.CSVFieldSeparator != nil {
		if v1.CSVFieldSeparator == nil ||
			*v1.CSVFieldSeparator != *v2.CSVFieldSeparator {
			d.CSVFieldSeparator = v2.CSVFieldSeparator
		}
	} else {
		d.CSVFieldSeparator = nil
	}
	if v2.ElsIndex != nil {
		if v1.ElsIndex == nil ||
			*v1.ElsIndex != *v2.ElsIndex {
			d.ElsIndex = v2.ElsIndex
		}
	} else {
		d.ElsIndex = nil
	}
	if v2.ElsIfPrimaryTerm != nil {
		if v1.ElsIfPrimaryTerm == nil ||
			*v1.ElsIfPrimaryTerm != *v2.ElsIfPrimaryTerm {
			d.ElsIfPrimaryTerm = v2.ElsIfPrimaryTerm
		}
	} else {
		d.ElsIfPrimaryTerm = nil
	}
	if v2.ElsIfSeqNo != nil {
		if v1.ElsIfSeqNo == nil ||
			*v1.ElsIfSeqNo != *v2.ElsIfSeqNo {
			d.ElsIfSeqNo = v2.ElsIfSeqNo
		}
	} else {
		d.ElsIfSeqNo = nil
	}
	if v2.ElsOpType != nil {
		if v1.ElsOpType == nil ||
			*v1.ElsOpType != *v2.ElsOpType {
			d.ElsOpType = v2.ElsOpType
		}
	} else {
		d.ElsOpType = nil
	}
	if v2.ElsPipeline != nil {
		if v1.ElsPipeline == nil ||
			*v1.ElsPipeline != *v2.ElsPipeline {
			d.ElsPipeline = v2.ElsPipeline
		}
	} else {
		d.ElsPipeline = nil
	}
	if v2.ElsRouting != nil {
		if v1.ElsRouting == nil ||
			*v1.ElsRouting != *v2.ElsRouting {
			d.ElsRouting = v2.ElsRouting
		}
	} else {
		d.ElsRouting = nil
	}
	if v2.ElsTimeout != nil {
		if v1.ElsTimeout == nil ||
			*v1.ElsTimeout != *v2.ElsTimeout {
			d.ElsTimeout = utils.StringPointer(v2.ElsTimeout.String())
		}
	} else {
		d.ElsTimeout = nil
	}
	if v2.ElsVersion != nil {
		if v1.ElsVersion == nil ||
			*v1.ElsVersion != *v2.ElsVersion {
			d.ElsVersion = v2.ElsVersion
		}
	} else {
		d.ElsVersion = nil
	}
	if v2.ElsVersionType != nil {
		if v1.ElsVersionType == nil ||
			*v1.ElsVersionType != *v2.ElsVersionType {
			d.ElsVersionType = v2.ElsVersionType
		}
	} else {
		d.ElsVersionType = nil
	}
	if v2.ElsWaitForActiveShards != nil {
		if v1.ElsWaitForActiveShards == nil ||
			*v1.ElsWaitForActiveShards != *v2.ElsWaitForActiveShards {
			d.ElsWaitForActiveShards = v2.ElsWaitForActiveShards
		}
	} else {
		d.ElsWaitForActiveShards = nil
	}
	if v2.SQLMaxIdleConns != nil {
		if v1.SQLMaxIdleConns == nil ||
			*v1.SQLMaxIdleConns != *v2.SQLMaxIdleConns {
			d.SQLMaxIdleConns = v2.SQLMaxIdleConns
		}
	} else {
		d.SQLMaxIdleConns = nil
	}
	if v2.SQLMaxOpenConns != nil {
		if v1.SQLMaxOpenConns == nil ||
			*v1.SQLMaxOpenConns != *v2.SQLMaxOpenConns {
			d.SQLMaxOpenConns = v2.SQLMaxOpenConns
		}
	} else {
		d.SQLMaxOpenConns = nil
	}
	if v2.SQLConnMaxLifetime != nil {
		if v1.SQLConnMaxLifetime == nil ||
			*v1.SQLConnMaxLifetime != *v2.SQLConnMaxLifetime {
			d.SQLConnMaxLifetime = utils.StringPointer(v2.SQLConnMaxLifetime.String())
		}
	} else {
		d.SQLConnMaxLifetime = nil
	}
	if v2.MYSQLDSNParams != nil {
		if v1.MYSQLDSNParams == nil || !reflect.DeepEqual(v1.MYSQLDSNParams, v2.MYSQLDSNParams) {
			d.MYSQLDSNParams = v2.MYSQLDSNParams
		}
	} else {
		d.MYSQLDSNParams = nil
	}
	if v2.SQLTableName != nil {
		if v1.SQLTableName == nil ||
			*v1.SQLTableName != *v2.SQLTableName {
			d.SQLTableName = v2.SQLTableName
		}
	} else {
		d.SQLTableName = nil
	}
	if v2.SQLDBName != nil {
		if v1.SQLDBName == nil ||
			*v1.SQLDBName != *v2.SQLDBName {
			d.SQLDBName = v2.SQLDBName
		}
	} else {
		d.SQLDBName = nil
	}
	if v2.SSLMode != nil {
		if v1.SSLMode == nil ||
			*v1.SSLMode != *v2.SSLMode {
			d.SSLMode = v2.SSLMode
		}
	} else {
		d.SSLMode = nil
	}
	if v2.KafkaTopic != nil {
		if v1.KafkaTopic == nil ||
			*v1.KafkaTopic != *v2.KafkaTopic {
			d.KafkaTopic = v2.KafkaTopic
		}
	} else {
		d.KafkaTopic = nil
	}
	if v2.AMQPQueueID != nil {
		if v1.AMQPQueueID == nil ||
			*v1.AMQPQueueID != *v2.AMQPQueueID {
			d.AMQPQueueID = v2.AMQPQueueID
		}
	} else {
		d.AMQPQueueID = nil
	}
	if v2.AMQPRoutingKey != nil {
		if v1.AMQPRoutingKey == nil ||
			*v1.AMQPRoutingKey != *v2.AMQPRoutingKey {
			d.AMQPRoutingKey = v2.AMQPRoutingKey
		}
	} else {
		d.AMQPRoutingKey = nil
	}
	if v2.AMQPExchange != nil {
		if v1.AMQPExchange == nil ||
			*v1.AMQPExchange != *v2.AMQPExchange {
			d.AMQPExchange = v2.AMQPExchange
		}
	} else {
		d.AMQPExchange = nil
	}
	if v2.AMQPExchangeType != nil {
		if v1.AMQPExchangeType == nil ||
			*v1.AMQPExchangeType != *v2.AMQPExchangeType {
			d.AMQPExchangeType = v2.AMQPExchangeType
		}
	} else {
		d.AMQPExchangeType = nil
	}
	if v2.AWSRegion != nil {
		if v1.AWSRegion == nil ||
			*v1.AWSRegion != *v2.AWSRegion {
			d.AWSRegion = v2.AWSRegion
		}
	} else {
		d.AWSRegion = nil
	}
	if v2.AWSKey != nil {
		if v1.AWSKey == nil ||
			*v1.AWSKey != *v2.AWSKey {
			d.AWSKey = v2.AWSKey
		}
	} else {
		d.AWSKey = nil
	}
	if v2.AWSSecret != nil {
		if v1.AWSSecret == nil ||
			*v1.AWSSecret != *v2.AWSSecret {
			d.AWSSecret = v2.AWSSecret
		}
	} else {
		d.AWSSecret = nil
	}
	if v2.AWSToken != nil {
		if v1.AWSToken == nil ||
			*v1.AWSToken != *v2.AWSToken {
			d.AWSToken = v2.AWSToken
		}
	} else {
		d.AWSToken = nil
	}
	if v2.SQSQueueID != nil {
		if v1.SQSQueueID == nil ||
			*v1.SQSQueueID != *v2.SQSQueueID {
			d.SQSQueueID = v2.SQSQueueID
		}
	} else {
		d.SQSQueueID = nil
	}
	if v2.S3BucketID != nil {
		if v1.S3BucketID == nil ||
			*v1.S3BucketID != *v2.S3BucketID {
			d.S3BucketID = v2.S3BucketID
		}
	} else {
		d.S3BucketID = nil
	}
	if v2.S3FolderPath != nil {
		if v1.S3FolderPath == nil ||
			*v1.S3FolderPath != *v2.S3FolderPath {
			d.S3FolderPath = v2.S3FolderPath
		}
	} else {
		d.S3FolderPath = nil
	}
	if v2.NATSJetStream != nil {
		if v1.NATSJetStream == nil ||
			*v1.NATSJetStream != *v2.NATSJetStream {
			d.NATSJetStream = v2.NATSJetStream
		}
	} else {
		d.NATSJetStream = nil
	}
	if v2.NATSSubject != nil {
		if v1.NATSSubject == nil ||
			*v1.NATSSubject != *v2.NATSSubject {
			d.NATSSubject = v2.NATSSubject
		}
	} else {
		d.NATSSubject = nil
	}
	if v2.NATSJWTFile != nil {
		if v1.NATSJWTFile == nil ||
			*v1.NATSJWTFile != *v2.NATSJWTFile {
			d.NATSJWTFile = v2.NATSJWTFile
		}
	} else {
		d.NATSJWTFile = nil
	}
	if v2.NATSSeedFile != nil {
		if v1.NATSSeedFile == nil ||
			*v1.NATSSeedFile != *v2.NATSSeedFile {
			d.NATSSeedFile = v2.NATSSeedFile
		}
	} else {
		d.NATSSeedFile = nil
	}
	if v2.NATSCertificateAuthority != nil {
		if v1.NATSCertificateAuthority == nil ||
			*v1.NATSCertificateAuthority != *v2.NATSCertificateAuthority {
			d.NATSCertificateAuthority = v2.NATSCertificateAuthority
		}
	} else {
		d.NATSCertificateAuthority = nil
	}
	if v2.NATSClientCertificate != nil {
		if v1.NATSClientCertificate == nil ||
			*v1.NATSClientCertificate != *v2.NATSClientCertificate {
			d.NATSClientCertificate = v2.NATSClientCertificate
		}
	} else {
		d.NATSClientCertificate = nil
	}
	if v2.NATSClientKey != nil {
		if v1.NATSClientKey == nil ||
			*v1.NATSClientKey != *v2.NATSClientKey {
			d.NATSClientKey = v2.NATSClientKey
		}
	} else {
		d.NATSClientKey = nil
	}
	if v2.NATSJetStreamMaxWait != nil {
		if v1.NATSJetStreamMaxWait == nil ||
			*v1.NATSJetStreamMaxWait != *v2.NATSJetStreamMaxWait {
			d.NATSJetStreamMaxWait = utils.StringPointer(v2.NATSJetStreamMaxWait.String())
		}
	} else {
		d.NATSJetStreamMaxWait = nil
	}
	if v2.RPCCodec != nil {
		if v1.RPCCodec == nil ||
			*v1.RPCCodec != *v2.RPCCodec {
			d.RPCCodec = v2.RPCCodec
		}
	} else {
		d.RPCCodec = nil
	}
	if v2.ServiceMethod != nil {
		if v1.ServiceMethod == nil ||
			*v1.ServiceMethod != *v2.ServiceMethod {
			d.ServiceMethod = v2.ServiceMethod
		}
	} else {
		d.ServiceMethod = nil
	}
	if v2.KeyPath != nil {
		if v1.KeyPath == nil ||
			*v1.KeyPath != *v2.KeyPath {
			d.KeyPath = v2.KeyPath
		}
	} else {
		d.KeyPath = nil
	}
	if v2.CertPath != nil {
		if v1.CertPath == nil ||
			*v1.CertPath != *v2.CertPath {
			d.CertPath = v2.CertPath
		}
	} else {
		d.CertPath = nil
	}
	if v2.CAPath != nil {
		if v1.CAPath == nil ||
			*v1.CAPath != *v2.CAPath {
			d.CAPath = v2.CAPath
		}
	} else {
		d.CAPath = nil
	}
	if v2.TLS != nil {
		if v1.TLS == nil ||
			*v1.TLS != *v2.TLS {
			d.TLS = v2.TLS
		}
	} else {
		d.TLS = nil
	}
	if v2.ConnIDs != nil {
		equal := true
		for i, val := range *v2.ConnIDs {
			if (*v1.ConnIDs)[i] != val {
				equal = false
				break
			}
		}
		if v1.ConnIDs == nil || !equal {
			d.ConnIDs = v2.ConnIDs
		} else {
			d.ConnIDs = nil
		}
	}
	if v2.RPCConnTimeout != nil {
		if v1.RPCConnTimeout == nil ||
			*v1.RPCConnTimeout != *v2.RPCConnTimeout {
			d.RPCConnTimeout = utils.StringPointer(v2.RPCConnTimeout.String())
		}
	} else {
		d.RPCConnTimeout = nil
	}
	if v2.RPCReplyTimeout != nil {
		if v1.RPCReplyTimeout == nil ||
			*v1.RPCReplyTimeout != *v2.RPCReplyTimeout {
			d.RPCReplyTimeout = utils.StringPointer(v2.RPCReplyTimeout.String())
		}
	} else {
		d.RPCReplyTimeout = nil
	}
	return d
}

func diffEventExporterJsonCfg(d *EventExporterJsonCfg, v1, v2 *EventExporterCfg, separator string) *EventExporterJsonCfg {
	if d == nil {
		d = new(EventExporterJsonCfg)
	}
	if v1.ID != v2.ID {
		d.Id = utils.StringPointer(v2.ID)
	}
	if v1.Type != v2.Type {
		d.Type = utils.StringPointer(v2.Type)
	}
	if v1.ExportPath != v2.ExportPath {
		d.Export_path = utils.StringPointer(v2.ExportPath)
	}
	d.Opts = diffEventExporterOptsJsonCfg(d.Opts, v1.Opts, v2.Opts)
	if v1.Timezone != v2.Timezone {
		d.Timezone = utils.StringPointer(v2.Timezone)
	}
	if !utils.SliceStringEqual(v1.Filters, v2.Filters) {
		d.Filters = &v2.Filters
	}
	flgs1 := v1.Flags.SliceFlags()
	flgs2 := v2.Flags.SliceFlags()
	if !utils.SliceStringEqual(flgs1, flgs2) {
		d.Flags = &flgs2
	}
	if !utils.SliceStringEqual(v1.AttributeSIDs, v2.AttributeSIDs) {
		d.Attribute_ids = &v2.AttributeSIDs
	}
	if v1.AttributeSCtx != v2.AttributeSCtx {
		d.Attribute_context = utils.StringPointer(v2.AttributeSCtx)
	}
	if v1.Synchronous != v2.Synchronous {
		d.Synchronous = utils.BoolPointer(v2.Synchronous)
	}
	if v1.Blocker != v2.Blocker {
		d.Blocker = utils.BoolPointer(v2.Blocker)
	}
	if v1.Attempts != v2.Attempts {
		d.Attempts = utils.IntPointer(v2.Attempts)
	}
	if v1.ConcurrentRequests != v2.ConcurrentRequests {
		d.Concurrent_requests = utils.IntPointer(v2.ConcurrentRequests)
	}
	var flds []*FcTemplateJsonCfg
	if d.Fields != nil {
		flds = *d.Fields
	}
	flds = diffFcTemplateJsonCfg(flds, v1.Fields, v2.Fields, separator)
	if flds != nil {
		d.Fields = &flds
	}
	if v1.FailedPostsDir != v2.FailedPostsDir {
		d.Failed_posts_dir = utils.StringPointer(v2.FailedPostsDir)
	}
	return d
}

func getEventExporterJsonCfg(d []*EventExporterJsonCfg, id string) (*EventExporterJsonCfg, int) {
	for i, v := range d {
		if v.Id != nil && *v.Id == id {
			return v, i
		}
	}
	return nil, -1
}

func getEventExporterCfg(d []*EventExporterCfg, id string) *EventExporterCfg {
	for _, v := range d {
		if v.ID == id {
			return v
		}
	}
	return &EventExporterCfg{
		Opts: &EventExporterOpts{},
	}
}

func diffEventExportersJsonCfg(d *[]*EventExporterJsonCfg, v1, v2 []*EventExporterCfg, separator string) *[]*EventExporterJsonCfg {
	if d == nil || *d == nil {
		d = &[]*EventExporterJsonCfg{}
	}
	for _, val := range v2 {
		dv, i := getEventExporterJsonCfg(*d, val.ID)
		dv = diffEventExporterJsonCfg(dv, getEventExporterCfg(v1, val.ID), val, separator)
		if i == -1 {
			*d = append(*d, dv)
		} else {
			(*d)[i] = dv
		}
	}

	return d
}

// EEsJsonCfg contains the configuration of EventExporterService
type EEsJsonCfg struct {
	Enabled          *bool
	Attributes_conns *[]string
	Cache            map[string]*CacheParamJsonCfg
	Exporters        *[]*EventExporterJsonCfg
}

func diffEEsJsonCfg(d *EEsJsonCfg, v1, v2 *EEsCfg, separator string) *EEsJsonCfg {
	if d == nil {
		d = new(EEsJsonCfg)
	}
	if v1.Enabled != v2.Enabled {
		d.Enabled = utils.BoolPointer(v2.Enabled)
	}
	if !utils.SliceStringEqual(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	d.Cache = diffCacheParamsJsonCfg(d.Cache, v2.Cache)
	d.Exporters = diffEventExportersJsonCfg(d.Exporters, v1.Exporters, v2.Exporters, separator)
	return d
}
