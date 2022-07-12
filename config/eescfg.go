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

func (eeS *EEsCfg) loadFromJSONCfg(jsnCfg *EEsJsonCfg, msgTemplates map[string][]*FCTemplate, sep string, dfltExpCfg *EventExporterCfg) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.Enabled != nil {
		eeS.Enabled = *jsnCfg.Enabled
	}
	if jsnCfg.Cache != nil {
		for kJsn, vJsn := range *jsnCfg.Cache {
			val := new(CacheParamCfg)
			if err := val.loadFromJSONCfg(vJsn); err != nil {
				return err
			}
			eeS.Cache[kJsn] = val
		}
	}
	if jsnCfg.Attributes_conns != nil {
		eeS.AttributeSConns = make([]string, len(*jsnCfg.Attributes_conns))
		for i, fID := range *jsnCfg.Attributes_conns {
			// if we have the connection internal we change the name so we can have internal rpc for each subsystem
			eeS.AttributeSConns[i] = fID
			if fID == utils.MetaInternal {
				eeS.AttributeSConns[i] = utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes)
			}
		}
	}
	return eeS.appendEEsExporters(jsnCfg.Exporters, msgTemplates, sep, dfltExpCfg)
}

func (eeS *EEsCfg) appendEEsExporters(exporters *[]*EventExporterJsonCfg, msgTemplates map[string][]*FCTemplate, separator string, dfltExpCfg *EventExporterCfg) (err error) {
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
			if dfltExpCfg != nil {
				exp = dfltExpCfg.Clone()
			} else {
				exp = new(EventExporterCfg)
				exp.Opts = &EventExporterOpts{}
			}
			eeS.Exporters = append(eeS.Exporters, exp)
		}
		if err = exp.loadFromJSONCfg(jsnExp, msgTemplates, separator); err != nil {
			return
		}
	}
	return
}

// Clone returns a deep copy of EEsCfg
func (eeS *EEsCfg) Clone() (cln *EEsCfg) {
	cln = &EEsCfg{
		Enabled:         eeS.Enabled,
		AttributeSConns: make([]string, len(eeS.AttributeSConns)),
		Cache:           make(map[string]*CacheParamCfg),
		Exporters:       make([]*EventExporterCfg, len(eeS.Exporters)),
	}
	for idx, sConn := range eeS.AttributeSConns {
		cln.AttributeSConns[idx] = sConn
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
func (eeS *EEsCfg) AsMapInterface(separator string) (initialMP map[string]interface{}) {
	initialMP = map[string]interface{}{
		utils.EnabledCfg: eeS.Enabled,
	}
	if eeS.AttributeSConns != nil {
		attributeSConns := make([]string, len(eeS.AttributeSConns))
		for i, item := range eeS.AttributeSConns {
			attributeSConns[i] = item
			if item == utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes) {
				attributeSConns[i] = utils.MetaInternal
			}
		}
		initialMP[utils.AttributeSConnsCfg] = attributeSConns
	}
	if eeS.Cache != nil {
		cache := make(map[string]interface{}, len(eeS.Cache))
		for key, value := range eeS.Cache {
			cache[key] = value.AsMapInterface()
		}
		initialMP[utils.CacheCfg] = cache
	}
	if eeS.Exporters != nil {
		exporters := make([]map[string]interface{}, len(eeS.Exporters))
		for i, item := range eeS.Exporters {
			exporters[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ExportersCfg] = exporters
	}
	return
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
	PgSSLMode                *string
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
	if jsnCfg.PgSSLMode != nil {
		eeOpts.PgSSLMode = jsnCfg.PgSSLMode
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
		eeC.Filters = make([]string, len(*jsnEec.Filters))
		for i, fltr := range *jsnEec.Filters {
			eeC.Filters[i] = fltr
		}
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = make([]string, len(*jsnEec.Attribute_ids))
		for i, fltr := range *jsnEec.Attribute_ids {
			eeC.AttributeSIDs[i] = fltr
		}
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
		cln.MYSQLDSNParams = make(map[string]string)
		cln.MYSQLDSNParams = eeOpts.MYSQLDSNParams
	}
	if eeOpts.SQLTableName != nil {
		cln.SQLTableName = utils.StringPointer(*eeOpts.SQLTableName)
	}
	if eeOpts.SQLDBName != nil {
		cln.SQLDBName = utils.StringPointer(*eeOpts.SQLDBName)
	}
	if eeOpts.PgSSLMode != nil {
		cln.PgSSLMode = utils.StringPointer(*eeOpts.PgSSLMode)
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
		cln.Filters = make([]string, len(eeC.Filters))
		for idx, val := range eeC.Filters {
			cln.Filters[idx] = val
		}
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = make([]string, len(eeC.AttributeSIDs))
		for idx, val := range eeC.AttributeSIDs {
			cln.AttributeSIDs[idx] = val
		}
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
	if eeC.Opts.PgSSLMode != nil {
		opts[utils.PgSSLModeCfg] = *eeC.Opts.PgSSLMode
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
