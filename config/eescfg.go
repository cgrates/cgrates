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
	RPCConnTimeout           *time.Duration
	RPCReplyTimeout          *time.Duration
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
	return &EventExporterOpts{
		CSVFieldSeparator:        eeOpts.CSVFieldSeparator,
		ElsIndex:                 eeOpts.ElsIndex,
		ElsIfPrimaryTerm:         eeOpts.ElsIfPrimaryTerm,
		ElsIfSeqNo:               eeOpts.ElsIfSeqNo,
		ElsOpType:                eeOpts.ElsOpType,
		ElsPipeline:              eeOpts.ElsPipeline,
		ElsRouting:               eeOpts.ElsRouting,
		ElsTimeout:               eeOpts.ElsTimeout,
		ElsVersion:               eeOpts.ElsVersion,
		ElsVersionType:           eeOpts.ElsVersionType,
		ElsWaitForActiveShards:   eeOpts.ElsWaitForActiveShards,
		SQLMaxIdleConns:          eeOpts.SQLMaxIdleConns,
		SQLMaxOpenConns:          eeOpts.SQLMaxOpenConns,
		SQLConnMaxLifetime:       eeOpts.SQLConnMaxLifetime,
		SQLTableName:             eeOpts.SQLTableName,
		SQLDBName:                eeOpts.SQLDBName,
		SSLMode:                  eeOpts.SSLMode,
		KafkaTopic:               eeOpts.KafkaTopic,
		AMQPQueueID:              eeOpts.AMQPQueueID,
		AMQPRoutingKey:           eeOpts.AMQPRoutingKey,
		AMQPExchange:             eeOpts.AMQPExchange,
		AMQPExchangeType:         eeOpts.AMQPExchangeType,
		AWSRegion:                eeOpts.AWSRegion,
		AWSKey:                   eeOpts.AWSKey,
		AWSSecret:                eeOpts.AWSSecret,
		AWSToken:                 eeOpts.AWSToken,
		SQSQueueID:               eeOpts.SQSQueueID,
		S3BucketID:               eeOpts.S3BucketID,
		S3FolderPath:             eeOpts.S3FolderPath,
		NATSJetStream:            eeOpts.NATSJetStream,
		NATSSubject:              eeOpts.NATSSubject,
		NATSJWTFile:              eeOpts.NATSJWTFile,
		NATSSeedFile:             eeOpts.NATSSeedFile,
		NATSCertificateAuthority: eeOpts.NATSCertificateAuthority,
		NATSClientCertificate:    eeOpts.NATSClientCertificate,
		NATSClientKey:            eeOpts.NATSClientKey,
		NATSJetStreamMaxWait:     eeOpts.NATSJetStreamMaxWait,
		RPCCodec:                 eeOpts.RPCCodec,
		ServiceMethod:            eeOpts.ServiceMethod,
		KeyPath:                  eeOpts.KeyPath,
		CertPath:                 eeOpts.CertPath,
		CAPath:                   eeOpts.CAPath,
		TLS:                      eeOpts.TLS,
		RPCConnTimeout:           eeOpts.RPCConnTimeout,
		RPCReplyTimeout:          eeOpts.RPCReplyTimeout,
	}
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
	opts := map[string]interface{}{
		utils.CSVFieldSepOpt:           eeC.Opts.CSVFieldSeparator,
		utils.ElsIndex:                 eeC.Opts.ElsIndex,
		utils.ElsIfPrimaryTerm:         eeC.Opts.ElsIfPrimaryTerm,
		utils.ElsIfSeqNo:               eeC.Opts.ElsIfSeqNo,
		utils.ElsOpType:                eeC.Opts.ElsOpType,
		utils.ElsPipeline:              eeC.Opts.ElsPipeline,
		utils.ElsRouting:               eeC.Opts.ElsRouting,
		utils.ElsTimeout:               eeC.Opts.ElsTimeout,
		utils.ElsVersionLow:            eeC.Opts.ElsVersion,
		utils.ElsVersionType:           eeC.Opts.ElsVersionType,
		utils.ElsWaitForActiveShards:   eeC.Opts.ElsWaitForActiveShards,
		utils.SQLMaxIdleConnsCfg:       eeC.Opts.SQLMaxIdleConns,
		utils.SQLMaxOpenConns:          eeC.Opts.SQLMaxOpenConns,
		utils.SQLConnMaxLifetime:       eeC.Opts.SQLConnMaxLifetime,
		utils.SQLTableNameOpt:          eeC.Opts.SQLTableName,
		utils.SQLDBNameOpt:             eeC.Opts.SQLDBName,
		utils.SSLModeCfg:               eeC.Opts.SSLMode,
		utils.KafkaTopic:               eeC.Opts.KafkaTopic,
		utils.AMQPQueueID:              eeC.Opts.AMQPQueueID,
		utils.AMQPRoutingKey:           eeC.Opts.AMQPRoutingKey,
		utils.AMQPExchange:             eeC.Opts.AMQPExchange,
		utils.AMQPExchangeType:         eeC.Opts.AMQPExchangeType,
		utils.AWSRegion:                eeC.Opts.AWSRegion,
		utils.AWSKey:                   eeC.Opts.AWSKey,
		utils.AWSSecret:                eeC.Opts.AWSSecret,
		utils.AWSToken:                 eeC.Opts.AWSToken,
		utils.SQSQueueID:               eeC.Opts.SQSQueueID,
		utils.S3Bucket:                 eeC.Opts.S3BucketID,
		utils.S3FolderPath:             eeC.Opts.S3FolderPath,
		utils.NatsJetStream:            eeC.Opts.NATSJetStream,
		utils.NatsSubject:              eeC.Opts.NATSSubject,
		utils.NatsJWTFile:              eeC.Opts.NATSJWTFile,
		utils.NatsSeedFile:             eeC.Opts.NATSSeedFile,
		utils.NatsCertificateAuthority: eeC.Opts.NATSCertificateAuthority,
		utils.NatsClientCertificate:    eeC.Opts.NATSClientCertificate,
		utils.NatsClientKey:            eeC.Opts.NATSClientKey,
		utils.NatsJetStreamMaxWait:     eeC.Opts.NATSJetStreamMaxWait,
		utils.RpcCodec:                 eeC.Opts.RPCCodec,
		utils.ServiceMethod:            eeC.Opts.ServiceMethod,
		utils.KeyPath:                  eeC.Opts.KeyPath,
		utils.CertPath:                 eeC.Opts.CertPath,
		utils.CaPath:                   eeC.Opts.CAPath,
		utils.TLS:                      eeC.Opts.TLS,
		utils.RpcConnTimeout:           eeC.Opts.RPCConnTimeout,
		utils.RpcReplyTimeout:          eeC.Opts.RPCReplyTimeout,
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
	CSVFieldSeparator        *string `json:"csvFieldSeparator"`
	ElsIndex                 *string `json:"elsIndex"`
	ElsIfPrimaryTerm         *int    `json:"elsIfPrimaryTerm"`
	ElsIfSeqNo               *int    `json:"elsIfSeqNo"`
	ElsOpType                *string `json:"elsOpType"`
	ElsPipeline              *string `json:"elsPipeline"`
	ElsRouting               *string `json:"elsRouting"`
	ElsTimeout               *string `json:"elsTimeout"`
	ElsVersion               *int    `json:"elsVersion"`
	ElsVersionType           *string `json:"elsVersionType"`
	ElsWaitForActiveShards   *string `json:"elsWaitForActiveShards"`
	SQLMaxIdleConns          *int    `json:"sqlMaxIdleConns"`
	SQLMaxOpenConns          *int    `json:"sqlMaxOpenConns"`
	SQLConnMaxLifetime       *string `json:"sqlConnMaxLifetime"`
	SQLTableName             *string `json:"sqlTableName"`
	SQLDBName                *string `json:"sqlDBName"`
	SSLMode                  *string `json:"sslMode"`
	KafkaTopic               *string `json:"kafkaTopic"`
	AMQPQueueID              *string `json:"amqpQueueID"`
	AMQPRoutingKey           *string `json:"amqpRoutingKey"`
	AMQPExchange             *string `json:"amqpExchange"`
	AMQPExchangeType         *string `json:"amqpExchangeType"`
	AWSRegion                *string `json:"awsRegion"`
	AWSKey                   *string `json:"awsKey"`
	AWSSecret                *string `json:"awsSecret"`
	AWSToken                 *string `json:"awsToken"`
	SQSQueueID               *string `json:"sqsQueueID"`
	S3BucketID               *string `json:"s3BucketID"`
	S3FolderPath             *string `json:"s3FolderPath"`
	NATSJetStream            *bool   `json:"natsJetStream"`
	NATSSubject              *string `json:"natsSubject"`
	NATSJWTFile              *string `json:"natsJWTFile"`
	NATSSeedFile             *string `json:"natsSeedFile"`
	NATSCertificateAuthority *string `json:"natsCertificateAuthority"`
	NATSClientCertificate    *string `json:"natsClientCertificate"`
	NATSClientKey            *string `json:"natsClientKey"`
	NATSJetStreamMaxWait     *string `json:"natsJetStreamMaxWait"`
	RPCCodec                 *string `json:"rpcCodec"`
	ServiceMethod            *string `json:"serviceMethod"`
	KeyPath                  *string `json:"keyPath"`
	CertPath                 *string `json:"certPath"`
	CAPath                   *string `json:"caPath"`
	TLS                      *bool   `json:"tls"`
	RPCConnTimeout           *string `json:"rpcConnTimeout"`
	RPCReplyTimeout          *string `json:"rpcReplyTimeout"`
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
	Attempts            *int
	Concurrent_requests *int
	Failed_posts_dir    *string
	Fields              *[]*FcTemplateJsonCfg
}

func diffEventExporterOptsJsonCfg(d *EventExporterOptsJson, v1, v2 *EventExporterOpts) *EventExporterOptsJson {
	if d == nil {
		d = new(EventExporterOptsJson)
	}
	if *v1.CSVFieldSeparator != *v2.CSVFieldSeparator {
		d.CSVFieldSeparator = v2.CSVFieldSeparator
	}
	if *v1.ElsIndex != *v2.ElsIndex {
		d.ElsIndex = v2.ElsIndex
	}
	if *v1.ElsIfPrimaryTerm != *v2.ElsIfPrimaryTerm {
		d.ElsIfPrimaryTerm = v2.ElsIfPrimaryTerm
	}
	if *v1.ElsIfSeqNo != *v2.ElsIfSeqNo {
		d.ElsIfSeqNo = v2.ElsIfSeqNo
	}
	if *v1.ElsOpType != *v2.ElsOpType {
		d.ElsOpType = v2.ElsOpType
	}
	if *v1.ElsPipeline != *v2.ElsPipeline {
		d.ElsPipeline = v2.ElsPipeline
	}
	if *v1.ElsRouting != *v2.ElsRouting {
		d.ElsRouting = v2.ElsRouting
	}
	if *v1.ElsTimeout != *v2.ElsTimeout {
		d.ElsTimeout = utils.StringPointer(v2.ElsTimeout.String())
	}
	if *v1.ElsVersion != *v2.ElsVersion {
		d.ElsVersion = v2.ElsVersion
	}
	if *v1.ElsVersionType != *v2.ElsVersionType {
		d.ElsVersionType = v2.ElsVersionType
	}
	if *v1.ElsWaitForActiveShards != *v2.ElsWaitForActiveShards {
		d.ElsWaitForActiveShards = v2.ElsWaitForActiveShards
	}
	if *v1.SQLMaxIdleConns != *v2.SQLMaxIdleConns {
		d.SQLMaxIdleConns = v2.SQLMaxIdleConns
	}
	if *v1.SQLMaxOpenConns != *v2.SQLMaxOpenConns {
		d.SQLMaxOpenConns = v2.SQLMaxOpenConns
	}
	if *v1.SQLConnMaxLifetime != *v2.SQLConnMaxLifetime {
		d.SQLConnMaxLifetime = utils.StringPointer(v2.SQLConnMaxLifetime.String())
	}
	if *v1.SQLTableName != *v2.SQLTableName {
		d.SQLTableName = v2.SQLTableName
	}
	if *v1.SQLDBName != *v2.SQLDBName {
		d.SQLDBName = v2.SQLDBName
	}
	if *v1.SSLMode != *v2.SSLMode {
		d.SSLMode = v2.SSLMode
	}
	if *v1.KafkaTopic != *v2.KafkaTopic {
		d.KafkaTopic = v2.KafkaTopic
	}
	if *v1.AMQPQueueID != *v2.AMQPQueueID {
		d.AMQPQueueID = v2.AMQPQueueID
	}
	if *v1.AMQPRoutingKey != *v2.AMQPRoutingKey {
		d.AMQPRoutingKey = v2.AMQPRoutingKey
	}
	if *v1.AMQPExchange != *v2.AMQPExchange {
		d.AMQPExchange = v2.AMQPExchange
	}
	if *v1.AMQPExchangeType != *v2.AMQPExchangeType {
		d.AMQPExchangeType = v2.AMQPExchangeType
	}
	if *v1.AWSRegion != *v2.AWSRegion {
		d.AWSRegion = v2.AWSRegion
	}
	if *v1.AWSKey != *v2.AWSKey {
		d.AWSKey = v2.AWSKey
	}
	if *v1.AWSSecret != *v2.AWSSecret {
		d.AWSSecret = v2.AWSSecret
	}
	if *v1.AWSToken != *v2.AWSToken {
		d.AWSToken = v2.AWSToken
	}
	if *v1.SQSQueueID != *v2.SQSQueueID {
		d.SQSQueueID = v2.SQSQueueID
	}
	if *v1.S3BucketID != *v2.S3BucketID {
		d.S3BucketID = v2.S3BucketID
	}
	if *v1.S3FolderPath != *v2.S3FolderPath {
		d.S3FolderPath = v2.S3FolderPath
	}
	if *v1.NATSJetStream != *v2.NATSJetStream {
		d.NATSJetStream = v2.NATSJetStream
	}
	if *v1.NATSSubject != *v2.NATSSubject {
		d.NATSSubject = v2.NATSSubject
	}
	if *v1.NATSJWTFile != *v2.NATSJWTFile {
		d.NATSJWTFile = v2.NATSJWTFile
	}
	if *v1.NATSSeedFile != *v2.NATSSeedFile {
		d.NATSSeedFile = v2.NATSSeedFile
	}
	if *v1.NATSCertificateAuthority != *v2.NATSCertificateAuthority {
		d.NATSCertificateAuthority = v2.NATSCertificateAuthority
	}
	if *v1.NATSClientCertificate != *v2.NATSClientCertificate {
		d.NATSClientCertificate = v2.NATSClientCertificate
	}
	if *v1.NATSClientKey != *v2.NATSClientKey {
		d.NATSClientKey = v2.NATSClientKey
	}
	if *v1.NATSJetStreamMaxWait != *v2.NATSJetStreamMaxWait {
		d.NATSJetStreamMaxWait = utils.StringPointer(v2.NATSJetStreamMaxWait.String())
	}
	if *v1.RPCCodec != *v2.RPCCodec {
		d.RPCCodec = v2.RPCCodec
	}
	if *v1.ServiceMethod != *v2.ServiceMethod {
		d.ServiceMethod = v2.ServiceMethod
	}
	if *v1.KeyPath != *v2.KeyPath {
		d.KeyPath = v2.KeyPath
	}
	if *v1.CertPath != *v2.CertPath {
		d.CertPath = v2.CertPath
	}
	if *v1.CAPath != *v2.CAPath {
		d.CAPath = v2.CAPath
	}
	if *v1.TLS != *v2.TLS {
		d.TLS = v2.TLS
	}
	if *v1.RPCConnTimeout != *v2.RPCConnTimeout {
		d.RPCConnTimeout = utils.StringPointer(v2.RPCConnTimeout.String())
	}
	if *v1.RPCReplyTimeout != *v2.RPCReplyTimeout {
		d.RPCReplyTimeout = utils.StringPointer(v2.RPCReplyTimeout.String())
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
	return new(EventExporterCfg)
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
