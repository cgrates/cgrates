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
	"slices"
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

// ExporterCfg iterates over the Exporters slice and returns the exporter
// configuration associated with the specified "id". If none were found, the
// method will return nil.
func (eeS *EEsCfg) ExporterCfg(id string) *EventExporterCfg {
	for _, eeCfg := range eeS.Exporters {
		if eeCfg.ID == id {
			return eeCfg
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
		cln.AttributeSConns = slices.Clone(eeS.AttributeSConns)
	}
	for key, value := range eeS.Cache {
		cln.Cache[key] = value.Clone()
	}
	for idx, exp := range eeS.Exporters {
		cln.Exporters[idx] = exp.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (eeS EEsCfg) AsMapInterface(separator string) any {
	mp := map[string]any{
		utils.EnabledCfg: eeS.Enabled,
	}
	if eeS.AttributeSConns != nil {
		mp[utils.AttributeSConnsCfg] = getInternalJSONConns(eeS.AttributeSConns)
	}
	if eeS.Cache != nil {
		cache := make(map[string]any, len(eeS.Cache))
		for key, value := range eeS.Cache {
			cache[key] = value.AsMapInterface()
		}
		mp[utils.CacheCfg] = cache
	}
	if eeS.Exporters != nil {
		exporters := make([]map[string]any, len(eeS.Exporters))
		for i, item := range eeS.Exporters {
			exporters[i] = item.AsMapInterface(separator)
		}
		mp[utils.ExportersCfg] = exporters
	}
	return mp
}

func (eeS *EEsCfg) exporterIDs() []string {
	ids := make([]string, 0, len(eeS.Exporters))
	for _, exporter := range eeS.Exporters {
		ids = append(ids, exporter.ID)
	}
	return ids
}

type EventExporterOpts struct {
	CSVFieldSeparator *string

	// elasticsearch index request opts
	ElsIndex               *string
	ElsRefresh             *string
	ElsOpType              *string
	ElsPipeline            *string
	ElsRouting             *string
	ElsTimeout             *time.Duration
	ElsWaitForActiveShards *string

	// elasticsearch client opts
	ElsCAPath                   *string
	ElsDiscoverNodesOnStart     *bool
	ElsDiscoverNodeInterval     *time.Duration
	ElsCloud                    *bool
	ElsAPIKey                   *string
	ElsCertificateFingerprint   *string
	ElsServiceToken             *string
	ElsUsername                 *string // Username for HTTP Basic Authentication.
	ElsPassword                 *string
	ElsEnableDebugLogger        *bool
	ElsLogger                   *string
	ElsCompressRequestBody      *bool
	ElsCompressRequestBodyLevel *int
	ElsRetryOnStatus            *[]int
	ElsMaxRetries               *int
	ElsDisableRetry             *bool

	SQLMaxIdleConns          *int
	SQLMaxOpenConns          *int
	SQLConnMaxLifetime       *time.Duration
	MYSQLDSNParams           map[string]string
	SQLTableName             *string
	SQLDBName                *string
	PgSSLMode                *string
	KafkaTopic               *string
	KafkaBatchSize           *int
	KafkaTLS                 *bool
	KafkaCAPath              *string
	KafkaSkipTLSVerify       *bool
	AMQPRoutingKey           *string
	AMQPQueueID              *string
	AMQPExchange             *string
	AMQPExchangeType         *string
	AMQPUsername             *string
	AMQPPassword             *string
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
	RPCAPIOpts               map[string]any
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
	EFsConns           []string // connection to EFService
	ConcurrentRequests int
	Fields             []*FCTemplate
	headerFields       []*FCTemplate
	contentFields      []*FCTemplate
	trailerFields      []*FCTemplate
}

// NewEventExporterCfg is a constructor for the EventExporterCfg, that is needed to initialize posters that are used by the
// readers and HTTP exporter actions
func NewEventExporterCfg(ID, exportType, exportPath, failedPostsDir string, attempts int, opts *EventExporterOpts) *EventExporterCfg {
	if opts == nil {
		opts = new(EventExporterOpts)
	}
	return &EventExporterCfg{
		ID:             ID,
		Type:           exportType,
		ExportPath:     exportPath,
		FailedPostsDir: failedPostsDir,
		Attempts:       attempts,
		Opts:           opts,
	}
}

func (eeOpts *EventExporterOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.CSVFieldSeparator != nil {
		eeOpts.CSVFieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if jsnCfg.ElsCloud != nil {
		eeOpts.ElsCloud = jsnCfg.ElsCloud
	}
	if jsnCfg.ElsAPIKey != nil {
		eeOpts.ElsAPIKey = jsnCfg.ElsAPIKey
	}
	if jsnCfg.ElsServiceToken != nil {
		eeOpts.ElsServiceToken = jsnCfg.ElsServiceToken
	}
	if jsnCfg.ElsCertificateFingerprint != nil {
		eeOpts.ElsCertificateFingerprint = jsnCfg.ElsCertificateFingerprint
	}
	if jsnCfg.ElsEnableDebugLogger != nil {
		eeOpts.ElsEnableDebugLogger = jsnCfg.ElsEnableDebugLogger
	}
	if jsnCfg.ElsLogger != nil {
		eeOpts.ElsLogger = jsnCfg.ElsLogger
	}
	if jsnCfg.ElsCompressRequestBody != nil {
		eeOpts.ElsCompressRequestBody = jsnCfg.ElsCompressRequestBody
	}
	if jsnCfg.ElsCompressRequestBodyLevel != nil {
		eeOpts.ElsCompressRequestBodyLevel = jsnCfg.ElsCompressRequestBodyLevel
	}
	if jsnCfg.ElsUsername != nil {
		eeOpts.ElsUsername = jsnCfg.ElsUsername
	}
	if jsnCfg.ElsPassword != nil {
		eeOpts.ElsPassword = jsnCfg.ElsPassword
	}
	if jsnCfg.ElsCAPath != nil {
		eeOpts.ElsCAPath = jsnCfg.ElsCAPath
	}
	if jsnCfg.ElsDiscoverNodesOnStart != nil {
		eeOpts.ElsDiscoverNodesOnStart = jsnCfg.ElsDiscoverNodesOnStart
	}
	if jsnCfg.ElsDiscoverNodesInterval != nil {
		var nodesInterval time.Duration
		if nodesInterval, err = utils.ParseDurationWithSecs(*jsnCfg.ElsDiscoverNodesInterval); err != nil {
			return
		}
		eeOpts.ElsDiscoverNodeInterval = utils.DurationPointer(nodesInterval)
	}
	if jsnCfg.ElsRetryOnStatus != nil {
		eeOpts.ElsRetryOnStatus = jsnCfg.ElsRetryOnStatus
	}
	if jsnCfg.ElsMaxRetries != nil {
		eeOpts.ElsMaxRetries = jsnCfg.ElsMaxRetries
	}
	if jsnCfg.ElsDisableRetry != nil {
		eeOpts.ElsDisableRetry = jsnCfg.ElsDisableRetry
	}
	if jsnCfg.ElsIndex != nil {
		eeOpts.ElsIndex = jsnCfg.ElsIndex
	}
	if jsnCfg.ElsRefresh != nil {
		eeOpts.ElsRefresh = jsnCfg.ElsRefresh
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
	if jsnCfg.KafkaBatchSize != nil {
		eeOpts.KafkaBatchSize = jsnCfg.KafkaBatchSize
	}
	if jsnCfg.KafkaTLS != nil {
		eeOpts.KafkaTLS = jsnCfg.KafkaTLS
	}
	if jsnCfg.KafkaCAPath != nil {
		eeOpts.KafkaCAPath = jsnCfg.KafkaCAPath
	}
	if jsnCfg.KafkaSkipTLSVerify != nil {
		eeOpts.KafkaSkipTLSVerify = jsnCfg.KafkaSkipTLSVerify
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
	if jsnCfg.AMQPUsername != nil {
		eeOpts.AMQPUsername = jsnCfg.AMQPUsername
	}
	if jsnCfg.AMQPPassword != nil {
		eeOpts.AMQPPassword = jsnCfg.AMQPPassword
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
		eeOpts.RPCAPIOpts = make(map[string]any)
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
		eeC.Filters = slices.Clone(*jsnEec.Filters)
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = slices.Clone(*jsnEec.Attribute_ids)
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
	if jsnEec.Efs_conns != nil {
		eeC.EFsConns = updateInternalConns(*jsnEec.Efs_conns, utils.MetaEFs)
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
		cln.CSVFieldSeparator = new(string)
		*cln.CSVFieldSeparator = *eeOpts.CSVFieldSeparator
	}
	if eeOpts.ElsIndex != nil {
		cln.ElsIndex = new(string)
		*cln.ElsIndex = *eeOpts.ElsIndex
	}
	if eeOpts.ElsRefresh != nil {
		cln.ElsRefresh = new(string)
		*cln.ElsRefresh = *eeOpts.ElsRefresh
	}
	if eeOpts.ElsOpType != nil {
		cln.ElsOpType = new(string)
		*cln.ElsOpType = *eeOpts.ElsOpType
	}
	if eeOpts.ElsPipeline != nil {
		cln.ElsPipeline = new(string)
		*cln.ElsPipeline = *eeOpts.ElsPipeline
	}
	if eeOpts.ElsRouting != nil {
		cln.ElsRouting = new(string)
		*cln.ElsRouting = *eeOpts.ElsRouting
	}
	if eeOpts.ElsTimeout != nil {
		cln.ElsTimeout = new(time.Duration)
		*cln.ElsTimeout = *eeOpts.ElsTimeout
	}
	if eeOpts.ElsWaitForActiveShards != nil {
		cln.ElsWaitForActiveShards = new(string)
		*cln.ElsWaitForActiveShards = *eeOpts.ElsWaitForActiveShards
	}
	if eeOpts.SQLMaxIdleConns != nil {
		cln.SQLMaxIdleConns = new(int)
		*cln.SQLMaxIdleConns = *eeOpts.SQLMaxIdleConns
	}
	if eeOpts.SQLMaxOpenConns != nil {
		cln.SQLMaxOpenConns = new(int)
		*cln.SQLMaxOpenConns = *eeOpts.SQLMaxOpenConns
	}
	if eeOpts.SQLConnMaxLifetime != nil {
		cln.SQLConnMaxLifetime = new(time.Duration)
		*cln.SQLConnMaxLifetime = *eeOpts.SQLConnMaxLifetime
	}
	if eeOpts.MYSQLDSNParams != nil {
		cln.MYSQLDSNParams = eeOpts.MYSQLDSNParams
	}
	if eeOpts.SQLTableName != nil {
		cln.SQLTableName = new(string)
		*cln.SQLTableName = *eeOpts.SQLTableName
	}
	if eeOpts.SQLDBName != nil {
		cln.SQLDBName = new(string)
		*cln.SQLDBName = *eeOpts.SQLDBName
	}
	if eeOpts.PgSSLMode != nil {
		cln.PgSSLMode = new(string)
		*cln.PgSSLMode = *eeOpts.PgSSLMode
	}
	if eeOpts.KafkaTopic != nil {
		cln.KafkaTopic = new(string)
		*cln.KafkaTopic = *eeOpts.KafkaTopic
	}
	if eeOpts.KafkaBatchSize != nil {
		cln.KafkaBatchSize = new(int)
		*cln.KafkaBatchSize = *eeOpts.KafkaBatchSize
	}
	if eeOpts.KafkaTLS != nil {
		cln.KafkaTLS = new(bool)
		*cln.KafkaTLS = *eeOpts.KafkaTLS
	}
	if eeOpts.KafkaCAPath != nil {
		cln.KafkaCAPath = new(string)
		*cln.KafkaCAPath = *eeOpts.KafkaCAPath
	}
	if eeOpts.KafkaSkipTLSVerify != nil {
		cln.KafkaSkipTLSVerify = new(bool)
		*cln.KafkaSkipTLSVerify = *eeOpts.KafkaSkipTLSVerify
	}
	if eeOpts.AMQPQueueID != nil {
		cln.AMQPQueueID = new(string)
		*cln.AMQPQueueID = *eeOpts.AMQPQueueID
	}
	if eeOpts.AMQPRoutingKey != nil {
		cln.AMQPRoutingKey = new(string)
		*cln.AMQPRoutingKey = *eeOpts.AMQPRoutingKey
	}
	if eeOpts.AMQPExchange != nil {
		cln.AMQPExchange = new(string)
		*cln.AMQPExchange = *eeOpts.AMQPExchange
	}
	if eeOpts.AMQPExchangeType != nil {
		cln.AMQPExchangeType = new(string)
		*cln.AMQPExchangeType = *eeOpts.AMQPExchangeType
	}
	if eeOpts.AMQPUsername != nil {
		cln.AMQPUsername = new(string)
		*cln.AMQPUsername = *eeOpts.AMQPUsername
	}
	if eeOpts.AMQPPassword != nil {
		cln.AMQPPassword = new(string)
		*cln.AMQPPassword = *eeOpts.AMQPPassword
	}
	if eeOpts.AWSRegion != nil {
		cln.AWSRegion = new(string)
		*cln.AWSRegion = *eeOpts.AWSRegion
	}
	if eeOpts.AWSKey != nil {
		cln.AWSKey = new(string)
		*cln.AWSKey = *eeOpts.AWSKey
	}
	if eeOpts.AWSSecret != nil {
		cln.AWSSecret = new(string)
		*cln.AWSSecret = *eeOpts.AWSSecret
	}
	if eeOpts.AWSToken != nil {
		cln.AWSToken = new(string)
		*cln.AWSToken = *eeOpts.AWSToken
	}
	if eeOpts.SQSQueueID != nil {
		cln.SQSQueueID = new(string)
		*cln.SQSQueueID = *eeOpts.SQSQueueID
	}
	if eeOpts.S3BucketID != nil {
		cln.S3BucketID = new(string)
		*cln.S3BucketID = *eeOpts.S3BucketID
	}
	if eeOpts.S3FolderPath != nil {
		cln.S3FolderPath = new(string)
		*cln.S3FolderPath = *eeOpts.S3FolderPath
	}
	if eeOpts.NATSJetStream != nil {
		cln.NATSJetStream = new(bool)
		*cln.NATSJetStream = *eeOpts.NATSJetStream
	}
	if eeOpts.NATSSubject != nil {
		cln.NATSSubject = new(string)
		*cln.NATSSubject = *eeOpts.NATSSubject
	}
	if eeOpts.NATSJWTFile != nil {
		cln.NATSJWTFile = new(string)
		*cln.NATSJWTFile = *eeOpts.NATSJWTFile
	}
	if eeOpts.NATSSeedFile != nil {
		cln.NATSSeedFile = new(string)
		*cln.NATSSeedFile = *eeOpts.NATSSeedFile
	}
	if eeOpts.NATSCertificateAuthority != nil {
		cln.NATSCertificateAuthority = new(string)
		*cln.NATSCertificateAuthority = *eeOpts.NATSCertificateAuthority
	}
	if eeOpts.NATSClientCertificate != nil {
		cln.NATSClientCertificate = new(string)
		cln.NATSClientCertificate = utils.StringPointer(*eeOpts.NATSClientCertificate)
	}
	if eeOpts.NATSClientKey != nil {
		cln.NATSClientKey = new(string)
		*cln.NATSClientKey = *eeOpts.NATSClientKey
	}
	if eeOpts.NATSJetStreamMaxWait != nil {
		cln.NATSJetStreamMaxWait = new(time.Duration)
		*cln.NATSJetStreamMaxWait = *eeOpts.NATSJetStreamMaxWait
	}
	if eeOpts.RPCCodec != nil {
		cln.RPCCodec = new(string)
		*cln.RPCCodec = *eeOpts.RPCCodec
	}
	if eeOpts.ServiceMethod != nil {
		cln.ServiceMethod = new(string)
		*cln.ServiceMethod = *eeOpts.ServiceMethod
	}
	if eeOpts.KeyPath != nil {
		cln.KeyPath = new(string)
		*cln.KeyPath = *eeOpts.KeyPath
	}
	if eeOpts.CertPath != nil {
		cln.CertPath = new(string)
		*cln.CertPath = *eeOpts.CertPath
	}
	if eeOpts.CAPath != nil {
		cln.CAPath = new(string)
		*cln.CAPath = *eeOpts.CAPath
	}
	if eeOpts.TLS != nil {
		cln.TLS = new(bool)
		*cln.TLS = *eeOpts.TLS
	}
	if eeOpts.ConnIDs != nil {
		cln.ConnIDs = new([]string)
		*cln.ConnIDs = *eeOpts.ConnIDs
	}
	if eeOpts.RPCConnTimeout != nil {
		cln.RPCConnTimeout = new(time.Duration)
		*cln.RPCConnTimeout = *eeOpts.RPCConnTimeout
	}
	if eeOpts.RPCReplyTimeout != nil {
		cln.RPCReplyTimeout = new(time.Duration)
		*cln.RPCReplyTimeout = *eeOpts.RPCReplyTimeout
	}
	if eeOpts.RPCAPIOpts != nil {
		cln.RPCAPIOpts = make(map[string]any)
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
		cln.Filters = slices.Clone(eeC.Filters)
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = slices.Clone(eeC.AttributeSIDs)
	}
	if eeC.EFsConns != nil {
		cln.EFsConns = slices.Clone(eeC.EFsConns)
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

// AsMapInterface returns the config as a map[string]any
func (eeC *EventExporterCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	flgs := eeC.Flags.SliceFlags()
	if flgs == nil {
		flgs = []string{}
	}
	initialMP = map[string]any{
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
		utils.OptsCfg:               eeC.Opts.AsMapInterface(),
	}
	if eeC.EFsConns != nil {
		initialMP[utils.EFsConnsCfg] = getInternalJSONConns(eeC.EFsConns)
	}
	if eeC.Fields != nil {
		fields := make([]map[string]any, 0, len(eeC.Fields))
		for _, fld := range eeC.Fields {
			fields = append(fields, fld.AsMapInterface(separator))
		}
		initialMP[utils.FieldsCfg] = fields
	}
	return
}

func (optsEes *EventExporterOpts) AsMapInterface() map[string]any {
	opts := map[string]any{}
	if optsEes.CSVFieldSeparator != nil {
		opts[utils.CSVFieldSepOpt] = *optsEes.CSVFieldSeparator
	}

	if optsEes.ElsIndex != nil {
		opts[utils.ElsIndex] = *optsEes.ElsIndex
	}
	if optsEes.ElsRefresh != nil {
		opts[utils.ElsRefresh] = *optsEes.ElsRefresh
	}
	if optsEes.ElsOpType != nil {
		opts[utils.ElsOpType] = *optsEes.ElsOpType
	}
	if optsEes.ElsPipeline != nil {
		opts[utils.ElsPipeline] = *optsEes.ElsPipeline
	}
	if optsEes.ElsRouting != nil {
		opts[utils.ElsRouting] = *optsEes.ElsRouting
	}
	if optsEes.ElsTimeout != nil {
		opts[utils.ElsTimeout] = optsEes.ElsTimeout.String()
	}
	if optsEes.ElsWaitForActiveShards != nil {
		opts[utils.ElsWaitForActiveShards] = *optsEes.ElsWaitForActiveShards
	}
	if optsEes.SQLMaxIdleConns != nil {
		opts[utils.SQLMaxIdleConnsCfg] = *optsEes.SQLMaxIdleConns
	}
	if optsEes.SQLMaxOpenConns != nil {
		opts[utils.SQLMaxOpenConns] = *optsEes.SQLMaxOpenConns
	}
	if optsEes.MYSQLDSNParams != nil {
		opts[utils.MYSQLDSNParams] = optsEes.MYSQLDSNParams
	}
	if optsEes.SQLConnMaxLifetime != nil {
		opts[utils.SQLConnMaxLifetime] = optsEes.SQLConnMaxLifetime.String()
	}
	if optsEes.PgSSLMode != nil {
		opts[utils.PgSSLModeCfg] = *optsEes.PgSSLMode
	}
	if optsEes.SQLTableName != nil {
		opts[utils.SQLTableNameOpt] = *optsEes.SQLTableName
	}
	if optsEes.SQLDBName != nil {
		opts[utils.SQLDBNameOpt] = *optsEes.SQLDBName
	}
	if optsEes.KafkaTopic != nil {
		opts[utils.KafkaTopic] = *optsEes.KafkaTopic
	}
	if optsEes.KafkaBatchSize != nil {
		opts[utils.KafkaBatchSize] = *optsEes.KafkaBatchSize
	}
	if optsEes.KafkaTLS != nil {
		opts[utils.KafkaTLS] = *optsEes.KafkaTLS
	}
	if optsEes.KafkaCAPath != nil {
		opts[utils.KafkaCAPath] = *optsEes.KafkaCAPath
	}
	if optsEes.KafkaSkipTLSVerify != nil {
		opts[utils.KafkaSkipTLSVerify] = *optsEes.KafkaSkipTLSVerify
	}
	if optsEes.AMQPQueueID != nil {
		opts[utils.AMQPQueueID] = *optsEes.AMQPQueueID
	}
	if optsEes.AMQPRoutingKey != nil {
		opts[utils.AMQPRoutingKey] = *optsEes.AMQPRoutingKey
	}
	if optsEes.AMQPExchange != nil {
		opts[utils.AMQPExchange] = *optsEes.AMQPExchange
	}
	if optsEes.AMQPExchangeType != nil {
		opts[utils.AMQPExchangeType] = *optsEes.AMQPExchangeType
	}
	if optsEes.AMQPUsername != nil {
		opts[utils.AMQPUsername] = *optsEes.AMQPUsername
	}
	if optsEes.AMQPPassword != nil {
		opts[utils.AMQPPassword] = *optsEes.AMQPPassword
	}
	if optsEes.AWSRegion != nil {
		opts[utils.AWSRegion] = *optsEes.AWSRegion
	}
	if optsEes.AWSKey != nil {
		opts[utils.AWSKey] = *optsEes.AWSKey
	}
	if optsEes.AWSSecret != nil {
		opts[utils.AWSSecret] = *optsEes.AWSSecret
	}
	if optsEes.AWSToken != nil {
		opts[utils.AWSToken] = *optsEes.AWSToken
	}
	if optsEes.SQSQueueID != nil {
		opts[utils.SQSQueueID] = *optsEes.SQSQueueID
	}
	if optsEes.S3BucketID != nil {
		opts[utils.S3Bucket] = *optsEes.S3BucketID
	}
	if optsEes.S3FolderPath != nil {
		opts[utils.S3FolderPath] = *optsEes.S3FolderPath
	}
	if optsEes.NATSJetStream != nil {
		opts[utils.NatsJetStream] = *optsEes.NATSJetStream
	}
	if optsEes.NATSSubject != nil {
		opts[utils.NatsSubject] = *optsEes.NATSSubject
	}
	if optsEes.NATSJWTFile != nil {
		opts[utils.NatsJWTFile] = *optsEes.NATSJWTFile
	}
	if optsEes.NATSSeedFile != nil {
		opts[utils.NatsSeedFile] = *optsEes.NATSSeedFile
	}
	if optsEes.NATSCertificateAuthority != nil {
		opts[utils.NatsCertificateAuthority] = *optsEes.NATSCertificateAuthority
	}
	if optsEes.NATSClientCertificate != nil {
		opts[utils.NatsClientCertificate] = *optsEes.NATSClientCertificate
	}
	if optsEes.NATSClientKey != nil {
		opts[utils.NatsClientKey] = *optsEes.NATSClientKey
	}
	if optsEes.NATSJetStreamMaxWait != nil {
		opts[utils.NatsJetStreamMaxWait] = optsEes.NATSJetStreamMaxWait.String()
	}
	if optsEes.RPCCodec != nil {
		opts[utils.RpcCodec] = *optsEes.RPCCodec
	}
	if optsEes.ServiceMethod != nil {
		opts[utils.ServiceMethod] = *optsEes.ServiceMethod
	}
	if optsEes.KeyPath != nil {
		opts[utils.KeyPath] = *optsEes.KeyPath
	}
	if optsEes.CertPath != nil {
		opts[utils.CertPath] = *optsEes.CertPath
	}
	if optsEes.CAPath != nil {
		opts[utils.CaPath] = *optsEes.CAPath
	}
	if optsEes.TLS != nil {
		opts[utils.Tls] = *optsEes.TLS
	}
	if optsEes.ConnIDs != nil {
		opts[utils.ConnIDs] = *optsEes.ConnIDs
	}
	if optsEes.RPCConnTimeout != nil {
		opts[utils.RpcConnTimeout] = optsEes.RPCConnTimeout.String()
	}
	if optsEes.RPCReplyTimeout != nil {
		opts[utils.RpcReplyTimeout] = optsEes.RPCReplyTimeout.String()
	}
	if optsEes.RPCAPIOpts != nil {
		opts[utils.RPCAPIOpts] = optsEes.RPCAPIOpts
	}
	return opts
}

type EventExporterOptsJson struct {
	CSVFieldSeparator           *string           `json:"csvFieldSeparator"`
	ElsCloud                    *bool             `json:"elsCloud"`
	ElsAPIKey                   *string           `json:"elsApiKey"`
	ElsServiceToken             *string           `json:"elsServiceToken"`
	ElsCertificateFingerprint   *string           `json:"elsCertificateFingerPrint"`
	ElsUsername                 *string           `json:"elsUsername"`
	ElsPassword                 *string           `json:"elsPassword"`
	ElsCAPath                   *string           `json:"elsCAPath"`
	ElsDiscoverNodesOnStart     *bool             `json:"elsDiscoverNodesOnStart"`
	ElsDiscoverNodesInterval    *string           `json:"elsDiscoverNodesInterval"`
	ElsEnableDebugLogger        *bool             `json:"elsEnableDebugLogger"`
	ElsLogger                   *string           `json:"elsLogger"`
	ElsCompressRequestBody      *bool             `json:"elsCompressRequestBody"`
	ElsCompressRequestBodyLevel *int              `json:"elsCompressRequestBodyLevel"`
	ElsRetryOnStatus            *[]int            `json:"elsRetryOnStatus"`
	ElsMaxRetries               *int              `json:"elsMaxRetries"`
	ElsDisableRetry             *bool             `json:"elsDisableRetry"`
	ElsIndex                    *string           `json:"elsIndex"`
	ElsRefresh                  *string           `json:"elsRefresh"`
	ElsOpType                   *string           `json:"elsOpType"`
	ElsPipeline                 *string           `json:"elsPipeline"`
	ElsRouting                  *string           `json:"elsRouting"`
	ElsTimeout                  *string           `json:"elsTimeout"`
	ElsWaitForActiveShards      *string           `json:"elsWaitForActiveShards"`
	SQLMaxIdleConns             *int              `json:"sqlMaxIdleConns"`
	SQLMaxOpenConns             *int              `json:"sqlMaxOpenConns"`
	SQLConnMaxLifetime          *string           `json:"sqlConnMaxLifetime"`
	MYSQLDSNParams              map[string]string `json:"mysqlDSNParams"`
	SQLTableName                *string           `json:"sqlTableName"`
	SQLDBName                   *string           `json:"sqlDBName"`
	PgSSLMode                   *string           `json:"pgSSLMode"`
	KafkaTopic                  *string           `json:"kafkaTopic"`
	KafkaBatchSize              *int              `json:"kafkaBatchSize"`
	KafkaTLS                    *bool             `json:"kafkaTLS"`
	KafkaCAPath                 *string           `json:"kafkaCAPath"`
	KafkaSkipTLSVerify          *bool             `json:"kafkaSkipTLSVerify"`
	AMQPQueueID                 *string           `json:"amqpQueueID"`
	AMQPRoutingKey              *string           `json:"amqpRoutingKey"`
	AMQPExchange                *string           `json:"amqpExchange"`
	AMQPExchangeType            *string           `json:"amqpExchangeType"`
	AMQPUsername                *string           `json:"amqpUsername"`
	AMQPPassword                *string           `json:"amqpPassword"`
	AWSRegion                   *string           `json:"awsRegion"`
	AWSKey                      *string           `json:"awsKey"`
	AWSSecret                   *string           `json:"awsSecret"`
	AWSToken                    *string           `json:"awsToken"`
	SQSQueueID                  *string           `json:"sqsQueueID"`
	S3BucketID                  *string           `json:"s3BucketID"`
	S3FolderPath                *string           `json:"s3FolderPath"`
	NATSJetStream               *bool             `json:"natsJetStream"`
	NATSSubject                 *string           `json:"natsSubject"`
	NATSJWTFile                 *string           `json:"natsJWTFile"`
	NATSSeedFile                *string           `json:"natsSeedFile"`
	NATSCertificateAuthority    *string           `json:"natsCertificateAuthority"`
	NATSClientCertificate       *string           `json:"natsClientCertificate"`
	NATSClientKey               *string           `json:"natsClientKey"`
	NATSJetStreamMaxWait        *string           `json:"natsJetStreamMaxWait"`
	RPCCodec                    *string           `json:"rpcCodec"`
	ServiceMethod               *string           `json:"serviceMethod"`
	KeyPath                     *string           `json:"keyPath"`
	CertPath                    *string           `json:"certPath"`
	CAPath                      *string           `json:"caPath"`
	ConnIDs                     *[]string         `json:"connIDs"`
	TLS                         *bool             `json:"tls"`
	RPCConnTimeout              *string           `json:"rpcConnTimeout"`
	RPCReplyTimeout             *string           `json:"rpcReplyTimeout"`
	RPCAPIOpts                  map[string]any    `json:"rpcAPIOpts"`
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
	Efs_conns           *[]string
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
	if v2.ElsRefresh != nil {
		if v1.ElsRefresh == nil ||
			*v1.ElsRefresh != *v2.ElsRefresh {
			d.ElsRefresh = v2.ElsRefresh
		}
	} else {
		d.ElsRefresh = nil
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
	if v2.PgSSLMode != nil {
		if v1.PgSSLMode == nil ||
			*v1.PgSSLMode != *v2.PgSSLMode {
			d.PgSSLMode = v2.PgSSLMode
		}
	} else {
		d.PgSSLMode = nil
	}
	if v2.KafkaTopic != nil {
		if v1.KafkaTopic == nil ||
			*v1.KafkaTopic != *v2.KafkaTopic {
			d.KafkaTopic = v2.KafkaTopic
		}
	} else {
		d.KafkaTopic = nil
	}
	if v2.KafkaBatchSize != nil {
		if v1.KafkaBatchSize == nil ||
			*v1.KafkaBatchSize != *v2.KafkaBatchSize {
			d.KafkaBatchSize = v2.KafkaBatchSize
		}
	} else {
		d.KafkaBatchSize = nil
	}
	if v2.KafkaTLS != nil {
		if v1.KafkaTLS == nil ||
			*v1.KafkaTLS != *v2.KafkaTLS {
			d.KafkaTLS = v2.KafkaTLS
		}
	} else {
		d.KafkaTLS = nil
	}
	if v2.KafkaCAPath != nil {
		if v1.KafkaCAPath == nil ||
			*v1.KafkaCAPath != *v2.KafkaCAPath {
			d.KafkaCAPath = v2.KafkaCAPath
		}
	} else {
		d.KafkaCAPath = nil
	}
	if v2.KafkaSkipTLSVerify != nil {
		if v1.KafkaSkipTLSVerify == nil ||
			*v1.KafkaSkipTLSVerify != *v2.KafkaSkipTLSVerify {
			d.KafkaSkipTLSVerify = v2.KafkaSkipTLSVerify
		}
	} else {
		d.KafkaSkipTLSVerify = nil
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
	if v2.AMQPUsername != nil {
		if v1.AMQPUsername == nil ||
			*v1.AMQPUsername != *v2.AMQPUsername {
			d.AMQPUsername = v2.AMQPUsername
		}
	} else {
		d.AMQPUsername = nil
	}
	if v2.AMQPPassword != nil {
		if v1.AMQPPassword == nil ||
			*v1.AMQPPassword != *v2.AMQPPassword {
			d.AMQPPassword = v2.AMQPPassword
		}
	} else {
		d.AMQPPassword = nil
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
	if !slices.Equal(v1.Filters, v2.Filters) {
		d.Filters = &v2.Filters
	}
	flgs1 := v1.Flags.SliceFlags()
	flgs2 := v2.Flags.SliceFlags()
	if !slices.Equal(flgs1, flgs2) {
		d.Flags = &flgs2
	}
	if !slices.Equal(v1.AttributeSIDs, v2.AttributeSIDs) {
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
	if !slices.Equal(v1.EFsConns, v2.EFsConns) {
		d.Efs_conns = &v2.EFsConns
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
	if !slices.Equal(v1.AttributeSConns, v2.AttributeSConns) {
		d.Attributes_conns = utils.SliceStringPointer(getInternalJSONConns(v2.AttributeSConns))
	}
	d.Cache = diffCacheParamsJsonCfg(d.Cache, v2.Cache)
	d.Exporters = diffEventExportersJsonCfg(d.Exporters, v1.Exporters, v2.Exporters, separator)
	return d
}
