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

	copy(cln.AttributeSConns, eeS.AttributeSConns)
	for key, value := range eeS.Cache {
		cln.Cache[key] = value.Clone()
	}
	for idx, exp := range eeS.Exporters {
		cln.Exporters[idx] = exp.Clone()
	}
	return
}

// AsMapInterface returns the config as a map[string]any
func (eeS *EEsCfg) AsMapInterface(separator string) (initialMP map[string]any) {
	initialMP = map[string]any{
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
		cache := make(map[string]any, len(eeS.Cache))
		for key, value := range eeS.Cache {
			cache[key] = value.AsMapInterface()
		}
		initialMP[utils.CacheCfg] = cache
	}
	if eeS.Exporters != nil {
		exporters := make([]map[string]any, len(eeS.Exporters))
		for i, item := range eeS.Exporters {
			exporters[i] = item.AsMapInterface(separator)
		}
		initialMP[utils.ExportersCfg] = exporters
	}
	return
}

func (eeS *EEsCfg) exporterIDs() []string {
	ids := make([]string, 0, len(eeS.Exporters))
	for _, exporter := range eeS.Exporters {
		ids = append(ids, exporter.ID)
	}
	return ids
}

type ElsOpts struct {
	// index request opts
	Index               *string
	Refresh             *string
	OpType              *string
	Pipeline            *string
	Routing             *string
	Timeout             *time.Duration
	WaitForActiveShards *string

	// elasticsearch client opts
	CAPath                   *string
	DiscoverNodesOnStart     *bool
	DiscoverNodeInterval     *time.Duration
	Cloud                    *bool
	APIKey                   *string
	CertificateFingerprint   *string
	ServiceToken             *string
	Username                 *string // Username for HTTP Basic Authentication.
	Password                 *string
	EnableDebugLogger        *bool
	Logger                   *string
	CompressRequestBody      *bool
	CompressRequestBodyLevel *int
	RetryOnStatus            *[]int
	MaxRetries               *int
	DisableRetry             *bool
}

type SQLOpts struct {
	MaxIdleConns        *int
	MaxOpenConns        *int
	ConnMaxLifetime     *time.Duration
	MYSQLDSNParams      map[string]string
	TableName           *string
	DBName              *string
	UpdateIndexedFields *[]string
	PgSSLMode           *string
}

type AMQPOpts struct {
	RoutingKey   *string
	QueueID      *string
	Exchange     *string
	ExchangeType *string
	Username     *string
	Password     *string
}

type AWSOpts struct {
	Region       *string
	Key          *string
	Secret       *string
	Token        *string
	SQSQueueID   *string
	S3BucketID   *string
	S3FolderPath *string
}

type NATSOpts struct {
	JetStream            *bool
	Subject              *string
	JWTFile              *string
	SeedFile             *string
	CertificateAuthority *string
	ClientCertificate    *string
	ClientKey            *string
	JetStreamMaxWait     *time.Duration
}

type RPCOpts struct {
	RPCCodec        *string
	ServiceMethod   *string
	KeyPath         *string
	CertPath        *string
	CAPath          *string
	TLS             *bool
	ConnIDs         *[]string
	RPCConnTimeout  *time.Duration
	RPCReplyTimeout *time.Duration
	RPCAPIOpts      map[string]any
}

type KafkaOpts struct {
	Topic         *string
	BatchSize     *int
	TLS           *bool
	CAPath        *string
	SkipTLSVerify *bool
}

type EventExporterOpts struct {
	CSVFieldSeparator *string
	Els               *ElsOpts
	SQL               *SQLOpts
	AMQP              *AMQPOpts
	AWS               *AWSOpts
	NATS              *NATSOpts
	RPC               *RPCOpts
	Kafka             *KafkaOpts
}

// EventExporterCfg the config for a Event Exporter
type EventExporterCfg struct {
	ID                   string
	Type                 string
	ExportPath           string
	Opts                 *EventExporterOpts
	Timezone             string
	Filters              []string
	Flags                utils.FlagsWithParams
	AttributeSIDs        []string // selective AttributeS profiles
	AttributeSCtx        string   // context to use when querying AttributeS
	Synchronous          bool
	Attempts             int
	FailedPostsDir       string
	ConcurrentRequests   int
	MetricsResetSchedule string
	Fields               []*FCTemplate
	headerFields         []*FCTemplate
	contentFields        []*FCTemplate
	trailerFields        []*FCTemplate
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
func (elsOpts *ElsOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.ElsCloud != nil {
		elsOpts.Cloud = jsnCfg.ElsCloud
	}
	if jsnCfg.ElsAPIKey != nil {
		elsOpts.APIKey = jsnCfg.ElsAPIKey
	}
	if jsnCfg.ElsServiceToken != nil {
		elsOpts.ServiceToken = jsnCfg.ElsServiceToken
	}
	if jsnCfg.ElsCertificateFingerprint != nil {
		elsOpts.CertificateFingerprint = jsnCfg.ElsCertificateFingerprint
	}
	if jsnCfg.ElsEnableDebugLogger != nil {
		elsOpts.EnableDebugLogger = jsnCfg.ElsEnableDebugLogger
	}
	if jsnCfg.ElsLogger != nil {
		elsOpts.Logger = jsnCfg.ElsLogger
	}
	if jsnCfg.ElsCompressRequestBody != nil {
		elsOpts.CompressRequestBody = jsnCfg.ElsCompressRequestBody
	}
	if jsnCfg.ElsCompressRequestBodyLevel != nil {
		elsOpts.CompressRequestBodyLevel = jsnCfg.ElsCompressRequestBodyLevel
	}
	if jsnCfg.ElsUsername != nil {
		elsOpts.Username = jsnCfg.ElsUsername
	}
	if jsnCfg.ElsPassword != nil {
		elsOpts.Password = jsnCfg.ElsPassword
	}
	if jsnCfg.ElsCAPath != nil {
		elsOpts.CAPath = jsnCfg.ElsCAPath
	}
	if jsnCfg.ElsDiscoverNodesOnStart != nil {
		elsOpts.DiscoverNodesOnStart = jsnCfg.ElsDiscoverNodesOnStart
	}
	if jsnCfg.ElsDiscoverNodesInterval != nil {
		var nodesInterval time.Duration
		if nodesInterval, err = utils.ParseDurationWithSecs(*jsnCfg.ElsDiscoverNodesInterval); err != nil {
			return
		}
		elsOpts.DiscoverNodeInterval = utils.DurationPointer(nodesInterval)
	}
	if jsnCfg.ElsRetryOnStatus != nil {
		elsOpts.RetryOnStatus = jsnCfg.ElsRetryOnStatus
	}
	if jsnCfg.ElsMaxRetries != nil {
		elsOpts.MaxRetries = jsnCfg.ElsMaxRetries
	}
	if jsnCfg.ElsDisableRetry != nil {
		elsOpts.DisableRetry = jsnCfg.ElsDisableRetry
	}
	if jsnCfg.ElsIndex != nil {
		elsOpts.Index = jsnCfg.ElsIndex
	}
	if jsnCfg.ElsRefresh != nil {
		elsOpts.Refresh = jsnCfg.ElsRefresh
	}
	if jsnCfg.ElsOpType != nil {
		elsOpts.OpType = jsnCfg.ElsOpType
	}
	if jsnCfg.ElsPipeline != nil {
		elsOpts.Pipeline = jsnCfg.ElsPipeline
	}
	if jsnCfg.ElsRouting != nil {
		elsOpts.Routing = jsnCfg.ElsRouting
	}
	if jsnCfg.ElsTimeout != nil {
		var elsTimeout time.Duration
		if elsTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.ElsTimeout); err != nil {
			return
		}
		elsOpts.Timeout = utils.DurationPointer(elsTimeout)
	}
	if jsnCfg.ElsWaitForActiveShards != nil {
		elsOpts.WaitForActiveShards = jsnCfg.ElsWaitForActiveShards
	}
	return
}

func (kafkaOpts *KafkaOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.KafkaTopic != nil {
		kafkaOpts.Topic = jsnCfg.KafkaTopic
	}
	if jsnCfg.KafkaBatchSize != nil {
		kafkaOpts.BatchSize = jsnCfg.KafkaBatchSize
	}
	if jsnCfg.KafkaTLS != nil {
		kafkaOpts.TLS = jsnCfg.KafkaTLS
	}
	if jsnCfg.KafkaCAPath != nil {
		kafkaOpts.CAPath = jsnCfg.KafkaCAPath
	}
	if jsnCfg.KafkaSkipTLSVerify != nil {
		kafkaOpts.SkipTLSVerify = jsnCfg.KafkaSkipTLSVerify
	}
	return
}

func (sqlOpts *SQLOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.SQLMaxIdleConns != nil {
		sqlOpts.MaxIdleConns = jsnCfg.SQLMaxIdleConns
	}
	if jsnCfg.SQLMaxOpenConns != nil {
		sqlOpts.MaxOpenConns = jsnCfg.SQLMaxOpenConns
	}
	if jsnCfg.SQLConnMaxLifetime != nil {
		var sqlConnMaxLifetime time.Duration
		if sqlConnMaxLifetime, err = utils.ParseDurationWithNanosecs(*jsnCfg.SQLConnMaxLifetime); err != nil {
			return
		}
		sqlOpts.ConnMaxLifetime = utils.DurationPointer(sqlConnMaxLifetime)
	}
	if jsnCfg.MYSQLDSNParams != nil {
		sqlOpts.MYSQLDSNParams = make(map[string]string)
		sqlOpts.MYSQLDSNParams = jsnCfg.MYSQLDSNParams
	}
	if jsnCfg.SQLTableName != nil {
		sqlOpts.TableName = jsnCfg.SQLTableName
	}
	if jsnCfg.SQLDBName != nil {
		sqlOpts.DBName = jsnCfg.SQLDBName
	}
	if jsnCfg.SQLUpdateIndexedFields != nil {
		uif := make([]string, len(*jsnCfg.SQLUpdateIndexedFields))
		copy(uif, *jsnCfg.SQLUpdateIndexedFields)
		sqlOpts.UpdateIndexedFields = &uif
	}
	if jsnCfg.PgSSLMode != nil {
		sqlOpts.PgSSLMode = jsnCfg.PgSSLMode
	}
	return
}

func (amqpOpts *AMQPOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {

	if jsnCfg.AMQPQueueID != nil {
		amqpOpts.QueueID = jsnCfg.AMQPQueueID
	}
	if jsnCfg.AMQPRoutingKey != nil {
		amqpOpts.RoutingKey = jsnCfg.AMQPRoutingKey
	}
	if jsnCfg.AMQPExchange != nil {
		amqpOpts.Exchange = jsnCfg.AMQPExchange
	}
	if jsnCfg.AMQPExchangeType != nil {
		amqpOpts.ExchangeType = jsnCfg.AMQPExchangeType
	}
	if jsnCfg.AMQPUsername != nil {
		amqpOpts.Username = jsnCfg.AMQPUsername
	}
	if jsnCfg.AMQPPassword != nil {
		amqpOpts.Password = jsnCfg.AMQPPassword
	}
	return
}

func (awsOpts *AWSOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.AWSRegion != nil {
		awsOpts.Region = jsnCfg.AWSRegion
	}
	if jsnCfg.AWSKey != nil {
		awsOpts.Key = jsnCfg.AWSKey
	}
	if jsnCfg.AWSSecret != nil {
		awsOpts.Secret = jsnCfg.AWSSecret
	}
	if jsnCfg.AWSToken != nil {
		awsOpts.Token = jsnCfg.AWSToken
	}
	if jsnCfg.SQSQueueID != nil {
		awsOpts.SQSQueueID = jsnCfg.SQSQueueID
	}
	if jsnCfg.S3BucketID != nil {
		awsOpts.S3BucketID = jsnCfg.S3BucketID
	}
	if jsnCfg.S3FolderPath != nil {
		awsOpts.S3FolderPath = jsnCfg.S3FolderPath
	}
	return
}
func (natsOpts *NATSOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.NATSJetStream != nil {
		natsOpts.JetStream = jsnCfg.NATSJetStream
	}
	if jsnCfg.NATSSubject != nil {
		natsOpts.Subject = jsnCfg.NATSSubject
	}
	if jsnCfg.NATSJWTFile != nil {
		natsOpts.JWTFile = jsnCfg.NATSJWTFile
	}
	if jsnCfg.NATSSeedFile != nil {
		natsOpts.SeedFile = jsnCfg.NATSSeedFile
	}
	if jsnCfg.NATSCertificateAuthority != nil {
		natsOpts.CertificateAuthority = jsnCfg.NATSCertificateAuthority
	}
	if jsnCfg.NATSClientCertificate != nil {
		natsOpts.ClientCertificate = jsnCfg.NATSClientCertificate
	}
	if jsnCfg.NATSClientKey != nil {
		natsOpts.ClientKey = jsnCfg.NATSClientKey
	}
	if jsnCfg.NATSJetStreamMaxWait != nil {
		var natsJetStreamMaxWait time.Duration
		if natsJetStreamMaxWait, err = utils.ParseDurationWithNanosecs(*jsnCfg.NATSJetStreamMaxWait); err != nil {
			return
		}
		natsOpts.JetStreamMaxWait = utils.DurationPointer(natsJetStreamMaxWait)
	}
	return
}
func (rpcOpts *RPCOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg.RPCCodec != nil {
		rpcOpts.RPCCodec = jsnCfg.RPCCodec
	}
	if jsnCfg.ServiceMethod != nil {
		rpcOpts.ServiceMethod = jsnCfg.ServiceMethod
	}
	if jsnCfg.KeyPath != nil {
		rpcOpts.KeyPath = jsnCfg.KeyPath
	}
	if jsnCfg.CertPath != nil {
		rpcOpts.CertPath = jsnCfg.CertPath
	}
	if jsnCfg.CAPath != nil {
		rpcOpts.CAPath = jsnCfg.CAPath
	}
	if jsnCfg.TLS != nil {
		rpcOpts.TLS = jsnCfg.TLS
	}
	if jsnCfg.ConnIDs != nil {
		rpcOpts.ConnIDs = jsnCfg.ConnIDs
	}
	if jsnCfg.RPCConnTimeout != nil {
		var rpcConnTimeout time.Duration
		if rpcConnTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RPCConnTimeout); err != nil {
			return
		}
		rpcOpts.RPCConnTimeout = utils.DurationPointer(rpcConnTimeout)
	}
	if jsnCfg.RPCReplyTimeout != nil {
		var rpcReplyTimeout time.Duration
		if rpcReplyTimeout, err = utils.ParseDurationWithNanosecs(*jsnCfg.RPCReplyTimeout); err != nil {
			return
		}
		rpcOpts.RPCReplyTimeout = utils.DurationPointer(rpcReplyTimeout)
	}
	if jsnCfg.RPCAPIOpts != nil {
		rpcOpts.RPCAPIOpts = make(map[string]any)
		rpcOpts.RPCAPIOpts = jsnCfg.RPCAPIOpts
	}

	return
}

func (eeOpts *EventExporterOpts) loadFromJSONCfg(jsnCfg *EventExporterOptsJson) (err error) {
	if jsnCfg == nil {
		return
	}
	if jsnCfg.CSVFieldSeparator != nil {
		eeOpts.CSVFieldSeparator = jsnCfg.CSVFieldSeparator
	}
	if err = eeOpts.Els.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.Kafka.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.SQL.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.AMQP.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.AWS.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.NATS.loadFromJSONCfg(jsnCfg); err != nil {
		return
	}
	if err = eeOpts.RPC.loadFromJSONCfg(jsnCfg); err != nil {
		return
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
		copy(eeC.Filters, *jsnEec.Filters)
	}
	if jsnEec.Flags != nil {
		eeC.Flags = utils.FlagsWithParamsFromSlice(*jsnEec.Flags)
	}
	if jsnEec.Attribute_context != nil {
		eeC.AttributeSCtx = *jsnEec.Attribute_context
	}
	if jsnEec.Attribute_ids != nil {
		eeC.AttributeSIDs = make([]string, len(*jsnEec.Attribute_ids))
		copy(eeC.AttributeSIDs, *jsnEec.Attribute_ids)

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
	if jsnEec.MetricsResetSchedule != nil {
		eeC.MetricsResetSchedule = *jsnEec.MetricsResetSchedule
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

func (elsOpts *ElsOpts) Clone() *ElsOpts {
	cln := &ElsOpts{}
	if elsOpts.Index != nil {
		cln.Index = new(string)
		*cln.Index = *elsOpts.Index
	}
	if elsOpts.Refresh != nil {
		cln.Refresh = new(string)
		*cln.Refresh = *elsOpts.Refresh
	}
	if elsOpts.OpType != nil {
		cln.OpType = new(string)
		*cln.OpType = *elsOpts.OpType
	}
	if elsOpts.Pipeline != nil {
		cln.Pipeline = new(string)
		*cln.Pipeline = *elsOpts.Pipeline
	}
	if elsOpts.Routing != nil {
		cln.Routing = new(string)
		*cln.Routing = *elsOpts.Routing
	}
	if elsOpts.Timeout != nil {
		cln.Timeout = new(time.Duration)
		*cln.Timeout = *elsOpts.Timeout
	}
	if elsOpts.WaitForActiveShards != nil {
		cln.WaitForActiveShards = new(string)
		*cln.WaitForActiveShards = *elsOpts.WaitForActiveShards
	}
	if elsOpts.CAPath != nil {
		cln.CAPath = new(string)
		*cln.CAPath = *elsOpts.CAPath
	}
	if elsOpts.DiscoverNodesOnStart != nil {
		cln.DiscoverNodesOnStart = new(bool)
		*cln.DiscoverNodesOnStart = *elsOpts.DiscoverNodesOnStart
	}
	if elsOpts.DiscoverNodeInterval != nil {
		cln.DiscoverNodeInterval = new(time.Duration)
		*cln.DiscoverNodeInterval = *elsOpts.DiscoverNodeInterval
	}
	if elsOpts.Cloud != nil {
		cln.Cloud = new(bool)
		*cln.Cloud = *elsOpts.Cloud
	}
	if elsOpts.APIKey != nil {
		cln.APIKey = new(string)
		*cln.APIKey = *elsOpts.APIKey
	}
	if elsOpts.CertificateFingerprint != nil {
		cln.CertificateFingerprint = new(string)
		*cln.CertificateFingerprint = *elsOpts.CertificateFingerprint
	}
	if elsOpts.ServiceToken != nil {
		cln.ServiceToken = new(string)
		*cln.ServiceToken = *elsOpts.ServiceToken
	}
	if elsOpts.Username != nil {
		cln.Username = new(string)
		*cln.Username = *elsOpts.Username
	}
	if elsOpts.Password != nil {
		cln.Password = new(string)
		*cln.Password = *elsOpts.Password
	}
	if elsOpts.EnableDebugLogger != nil {
		cln.EnableDebugLogger = new(bool)
		*cln.EnableDebugLogger = *elsOpts.EnableDebugLogger
	}
	if elsOpts.Logger != nil {
		cln.Logger = new(string)
		*cln.Logger = *elsOpts.Logger
	}
	if elsOpts.CompressRequestBody != nil {
		cln.CompressRequestBody = new(bool)
		*cln.CompressRequestBody = *elsOpts.CompressRequestBody
	}
	if elsOpts.CompressRequestBodyLevel != nil {
		cln.CompressRequestBodyLevel = new(int)
		*cln.CompressRequestBodyLevel = *elsOpts.CompressRequestBodyLevel
	}
	if elsOpts.RetryOnStatus != nil {
		cln.RetryOnStatus = new([]int)
		*cln.RetryOnStatus = *elsOpts.RetryOnStatus
	}
	if elsOpts.MaxRetries != nil {
		cln.MaxRetries = new(int)
		*cln.MaxRetries = *elsOpts.MaxRetries
	}
	if elsOpts.DisableRetry != nil {
		cln.DisableRetry = new(bool)
		*cln.DisableRetry = *elsOpts.DisableRetry
	}
	return cln
}

func (kafkaOpts *KafkaOpts) Clone() *KafkaOpts {
	cln := &KafkaOpts{}
	if kafkaOpts.Topic != nil {
		cln.Topic = new(string)
		*cln.Topic = *kafkaOpts.Topic
	}
	if kafkaOpts.BatchSize != nil {
		cln.BatchSize = new(int)
		*cln.BatchSize = *kafkaOpts.BatchSize
	}
	if kafkaOpts.TLS != nil {
		cln.TLS = new(bool)
		*cln.TLS = *kafkaOpts.TLS
	}
	if kafkaOpts.CAPath != nil {
		cln.CAPath = new(string)
		*cln.CAPath = *kafkaOpts.CAPath
	}
	if kafkaOpts.SkipTLSVerify != nil {
		cln.SkipTLSVerify = new(bool)
		*cln.SkipTLSVerify = *kafkaOpts.SkipTLSVerify
	}
	return cln
}

func (sqlOpts *SQLOpts) Clone() *SQLOpts {
	cln := &SQLOpts{}
	if sqlOpts.MaxIdleConns != nil {
		cln.MaxIdleConns = new(int)
		*cln.MaxIdleConns = *sqlOpts.MaxIdleConns
	}
	if sqlOpts.MaxOpenConns != nil {
		cln.MaxOpenConns = new(int)
		*cln.MaxOpenConns = *sqlOpts.MaxOpenConns
	}
	if sqlOpts.ConnMaxLifetime != nil {
		cln.ConnMaxLifetime = new(time.Duration)
		*cln.ConnMaxLifetime = *sqlOpts.ConnMaxLifetime
	}
	if sqlOpts.MYSQLDSNParams != nil {
		cln.MYSQLDSNParams = make(map[string]string)
		cln.MYSQLDSNParams = sqlOpts.MYSQLDSNParams
	}
	if sqlOpts.TableName != nil {
		cln.TableName = new(string)
		*cln.TableName = *sqlOpts.TableName
	}
	if sqlOpts.DBName != nil {
		cln.DBName = new(string)
		*cln.DBName = *sqlOpts.DBName
	}
	if sqlOpts.UpdateIndexedFields != nil {
		idx := make([]string, len(*sqlOpts.UpdateIndexedFields))
		copy(idx, *sqlOpts.UpdateIndexedFields)
		cln.UpdateIndexedFields = &idx
	}
	if sqlOpts.PgSSLMode != nil {
		cln.PgSSLMode = new(string)
		*cln.PgSSLMode = *sqlOpts.PgSSLMode
	}
	return cln
}

func (amqpOpts *AMQPOpts) Clone() *AMQPOpts {
	cln := &AMQPOpts{}
	if amqpOpts.QueueID != nil {
		cln.QueueID = new(string)
		*cln.QueueID = *amqpOpts.QueueID
	}
	if amqpOpts.RoutingKey != nil {
		cln.RoutingKey = new(string)
		*cln.RoutingKey = *amqpOpts.RoutingKey
	}
	if amqpOpts.Exchange != nil {
		cln.Exchange = new(string)
		*cln.Exchange = *amqpOpts.Exchange
	}
	if amqpOpts.ExchangeType != nil {
		cln.ExchangeType = new(string)
		*cln.ExchangeType = *amqpOpts.ExchangeType
	}
	if amqpOpts.Username != nil {
		cln.Username = new(string)
		*cln.Username = *amqpOpts.Username
	}
	if amqpOpts.Password != nil {
		cln.Password = new(string)
		*cln.Password = *amqpOpts.Password
	}
	return cln
}

func (awsOpts *AWSOpts) Clone() *AWSOpts {
	cln := &AWSOpts{}
	if awsOpts.Region != nil {
		cln.Region = new(string)
		*cln.Region = *awsOpts.Region
	}
	if awsOpts.Key != nil {
		cln.Key = new(string)
		*cln.Key = *awsOpts.Key
	}
	if awsOpts.Secret != nil {
		cln.Secret = new(string)
		*cln.Secret = *awsOpts.Secret
	}
	if awsOpts.Token != nil {
		cln.Token = new(string)
		*cln.Token = *awsOpts.Token
	}
	if awsOpts.SQSQueueID != nil {
		cln.SQSQueueID = new(string)
		*cln.SQSQueueID = *awsOpts.SQSQueueID
	}
	if awsOpts.S3BucketID != nil {
		cln.S3BucketID = new(string)
		*cln.S3BucketID = *awsOpts.S3BucketID
	}
	if awsOpts.S3FolderPath != nil {
		cln.S3FolderPath = new(string)
		*cln.S3FolderPath = *awsOpts.S3FolderPath
	}
	return cln
}

func (natsOpts *NATSOpts) Clone() *NATSOpts {
	cln := &NATSOpts{}
	if natsOpts.JetStream != nil {
		cln.JetStream = new(bool)
		*cln.JetStream = *natsOpts.JetStream
	}
	if natsOpts.Subject != nil {
		cln.Subject = new(string)
		*cln.Subject = *natsOpts.Subject
	}
	if natsOpts.JWTFile != nil {
		cln.JWTFile = new(string)
		*cln.JWTFile = *natsOpts.JWTFile
	}
	if natsOpts.SeedFile != nil {
		cln.SeedFile = new(string)
		*cln.SeedFile = *natsOpts.SeedFile
	}
	if natsOpts.CertificateAuthority != nil {
		cln.CertificateAuthority = new(string)
		*cln.CertificateAuthority = *natsOpts.CertificateAuthority
	}
	if natsOpts.ClientCertificate != nil {
		cln.ClientCertificate = new(string)
		*cln.ClientCertificate = *natsOpts.ClientCertificate
	}
	if natsOpts.ClientKey != nil {
		cln.ClientKey = new(string)
		*cln.ClientKey = *natsOpts.ClientKey
	}
	if natsOpts.JetStreamMaxWait != nil {
		cln.JetStreamMaxWait = new(time.Duration)
		*cln.JetStreamMaxWait = *natsOpts.JetStreamMaxWait
	}
	return cln
}

func (rpcOpts *RPCOpts) Clone() *RPCOpts {
	cln := &RPCOpts{}
	if rpcOpts.RPCCodec != nil {
		cln.RPCCodec = new(string)
		*cln.RPCCodec = *rpcOpts.RPCCodec
	}
	if rpcOpts.ServiceMethod != nil {
		cln.ServiceMethod = new(string)
		*cln.ServiceMethod = *rpcOpts.ServiceMethod
	}
	if rpcOpts.KeyPath != nil {
		cln.KeyPath = new(string)
		*cln.KeyPath = *rpcOpts.KeyPath
	}
	if rpcOpts.CertPath != nil {
		cln.CertPath = new(string)
		*cln.CertPath = *rpcOpts.CertPath
	}
	if rpcOpts.CAPath != nil {
		cln.CAPath = new(string)
		*cln.CAPath = *rpcOpts.CAPath
	}
	if rpcOpts.TLS != nil {
		cln.TLS = new(bool)
		*cln.TLS = *rpcOpts.TLS
	}
	if rpcOpts.ConnIDs != nil {
		cln.ConnIDs = new([]string)
		*cln.ConnIDs = *rpcOpts.ConnIDs
	}
	if rpcOpts.RPCConnTimeout != nil {
		cln.RPCConnTimeout = new(time.Duration)
		*cln.RPCConnTimeout = *rpcOpts.RPCConnTimeout
	}
	if rpcOpts.RPCReplyTimeout != nil {
		cln.RPCReplyTimeout = new(time.Duration)
		*cln.RPCReplyTimeout = *rpcOpts.RPCReplyTimeout
	}
	if rpcOpts.RPCAPIOpts != nil {
		cln.RPCAPIOpts = make(map[string]any)
		cln.RPCAPIOpts = rpcOpts.RPCAPIOpts
	}

	return cln
}
func (eeOpts *EventExporterOpts) Clone() *EventExporterOpts {
	cln := &EventExporterOpts{}
	if eeOpts.CSVFieldSeparator != nil {
		cln.CSVFieldSeparator = new(string)
		*cln.CSVFieldSeparator = *eeOpts.CSVFieldSeparator
	}
	if eeOpts.Els != nil {
		cln.Els = eeOpts.Els.Clone()
	}
	if eeOpts.SQL != nil {
		cln.SQL = eeOpts.SQL.Clone()
	}
	if eeOpts.Kafka != nil {
		cln.Kafka = eeOpts.Kafka.Clone()
	}
	if eeOpts.AMQP != nil {
		cln.AMQP = eeOpts.AMQP.Clone()
	}
	if eeOpts.AWS != nil {
		cln.AWS = eeOpts.AWS.Clone()
	}
	if eeOpts.NATS != nil {
		cln.NATS = eeOpts.NATS.Clone()
	}
	if eeOpts.RPC != nil {
		cln.RPC = eeOpts.RPC.Clone()
	}
	return cln
}

// Clone returns a deep copy of EventExporterCfg
func (eeC EventExporterCfg) Clone() (cln *EventExporterCfg) {
	cln = &EventExporterCfg{
		ID:                   eeC.ID,
		Type:                 eeC.Type,
		ExportPath:           eeC.ExportPath,
		Timezone:             eeC.Timezone,
		Flags:                eeC.Flags.Clone(),
		AttributeSCtx:        eeC.AttributeSCtx,
		Synchronous:          eeC.Synchronous,
		Attempts:             eeC.Attempts,
		ConcurrentRequests:   eeC.ConcurrentRequests,
		MetricsResetSchedule: eeC.MetricsResetSchedule,
		Fields:               make([]*FCTemplate, len(eeC.Fields)),
		headerFields:         make([]*FCTemplate, len(eeC.headerFields)),
		contentFields:        make([]*FCTemplate, len(eeC.contentFields)),
		trailerFields:        make([]*FCTemplate, len(eeC.trailerFields)),
		Opts:                 eeC.Opts.Clone(),
		FailedPostsDir:       eeC.FailedPostsDir,
	}

	if eeC.Filters != nil {
		cln.Filters = make([]string, len(eeC.Filters))
		copy(cln.Filters, eeC.Filters)
	}
	if eeC.AttributeSIDs != nil {
		cln.AttributeSIDs = make([]string, len(eeC.AttributeSIDs))
		copy(cln.AttributeSIDs, eeC.AttributeSIDs)
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
	opts := map[string]any{}
	if eeC.Opts.CSVFieldSeparator != nil {
		opts[utils.CSVFieldSepOpt] = *eeC.Opts.CSVFieldSeparator
	}
	if elsOpts := eeC.Opts.Els; elsOpts != nil {
		if elsOpts.Index != nil {
			opts[utils.ElsIndex] = *elsOpts.Index
		}
		if elsOpts.Refresh != nil {
			opts[utils.ElsRefresh] = *elsOpts.Refresh
		}
		if elsOpts.OpType != nil {
			opts[utils.ElsOpType] = *elsOpts.OpType
		}
		if elsOpts.Pipeline != nil {
			opts[utils.ElsPipeline] = *elsOpts.Pipeline
		}
		if elsOpts.Routing != nil {
			opts[utils.ElsRouting] = *elsOpts.Routing
		}
		if elsOpts.Timeout != nil {
			opts[utils.ElsTimeout] = elsOpts.Timeout.String()
		}
		if elsOpts.WaitForActiveShards != nil {
			opts[utils.ElsWaitForActiveShards] = *elsOpts.WaitForActiveShards
		}
		if elsOpts.CAPath != nil {
			opts[utils.ElsCAPath] = *elsOpts.CAPath
		}
		if elsOpts.DiscoverNodesOnStart != nil {
			opts[utils.ElsDiscoverNodesOnStart] = *elsOpts.DiscoverNodesOnStart
		}
		if elsOpts.DiscoverNodeInterval != nil {
			opts[utils.ElsDiscoverNodeInterval] = *elsOpts.DiscoverNodeInterval
		}
		if elsOpts.Cloud != nil {
			opts[utils.ElsCloud] = *elsOpts.Cloud
		}
		if elsOpts.APIKey != nil {
			opts[utils.ElsAPIKey] = *elsOpts.APIKey
		}
		if elsOpts.CertificateFingerprint != nil {
			opts[utils.ElsCertificateFingerprint] = *elsOpts.CertificateFingerprint
		}
		if elsOpts.ServiceToken != nil {
			opts[utils.ElsServiceToken] = *elsOpts.ServiceToken
		}
		if elsOpts.Username != nil {
			opts[utils.ElsUsername] = *elsOpts.Username
		}
		if elsOpts.Password != nil {
			opts[utils.ElsPassword] = *elsOpts.Password
		}
		if elsOpts.EnableDebugLogger != nil {
			opts[utils.ElsEnableDebugLogger] = *elsOpts.EnableDebugLogger
		}
		if elsOpts.Logger != nil {
			opts[utils.ElsLogger] = *elsOpts.Logger
		}
		if elsOpts.CompressRequestBody != nil {
			opts[utils.ElsCompressRequestBody] = *elsOpts.CompressRequestBody
		}
		if elsOpts.CompressRequestBodyLevel != nil {
			opts[utils.ElsCompressRequestBodyLevel] = *elsOpts.CompressRequestBodyLevel
		}
		if elsOpts.RetryOnStatus != nil {
			opts[utils.ElsRetryOnStatus] = *elsOpts.RetryOnStatus
		}
		if elsOpts.MaxRetries != nil {
			opts[utils.ElsMaxRetries] = *elsOpts.MaxRetries
		}
		if elsOpts.DisableRetry != nil {
			opts[utils.ElsDisableRetry] = *elsOpts.DisableRetry
		}
	}
	if sqlOpts := eeC.Opts.SQL; sqlOpts != nil {
		if sqlOpts.MaxIdleConns != nil {
			opts[utils.SQLMaxIdleConnsCfg] = *sqlOpts.MaxIdleConns
		}
		if sqlOpts.MaxOpenConns != nil {
			opts[utils.SQLMaxOpenConns] = *sqlOpts.MaxOpenConns
		}
		if sqlOpts.ConnMaxLifetime != nil {
			opts[utils.SQLConnMaxLifetime] = sqlOpts.ConnMaxLifetime.String()
		}
		if sqlOpts.MYSQLDSNParams != nil {
			opts[utils.MYSQLDSNParams] = sqlOpts.MYSQLDSNParams
		}
		if sqlOpts.TableName != nil {
			opts[utils.SQLTableNameOpt] = *sqlOpts.TableName
		}
		if sqlOpts.DBName != nil {
			opts[utils.SQLDBNameOpt] = *sqlOpts.DBName
		}
		if sqlOpts.UpdateIndexedFields != nil {
			updateIndexedFields := make([]string, len(*sqlOpts.UpdateIndexedFields))
			copy(updateIndexedFields, *sqlOpts.UpdateIndexedFields)
			opts[utils.SQLUpdateIndexedFieldsOpt] = updateIndexedFields
		}
		if sqlOpts.PgSSLMode != nil {
			opts[utils.PgSSLModeCfg] = *sqlOpts.PgSSLMode
		}
	}
	if kafkaOpts := eeC.Opts.Kafka; kafkaOpts != nil {
		if kafkaOpts.Topic != nil {
			opts[utils.KafkaTopic] = *kafkaOpts.Topic
		}
		if kafkaOpts.BatchSize != nil {
			opts[utils.KafkaBatchSize] = *kafkaOpts.BatchSize
		}
		if kafkaOpts.TLS != nil {
			opts[utils.KafkaTLS] = *kafkaOpts.TLS
		}
		if kafkaOpts.CAPath != nil {
			opts[utils.KafkaCAPath] = *kafkaOpts.CAPath
		}
		if kafkaOpts.SkipTLSVerify != nil {
			opts[utils.KafkaSkipTLSVerify] = *kafkaOpts.SkipTLSVerify
		}
	}
	if amOpts := eeC.Opts.AMQP; amOpts != nil {
		if amOpts.QueueID != nil {
			opts[utils.AMQPQueueID] = *amOpts.QueueID
		}
		if amOpts.RoutingKey != nil {
			opts[utils.AMQPRoutingKey] = *amOpts.RoutingKey
		}
		if amOpts.Exchange != nil {
			opts[utils.AMQPExchange] = *amOpts.Exchange
		}
		if amOpts.ExchangeType != nil {
			opts[utils.AMQPExchangeType] = *amOpts.ExchangeType
		}
		if amOpts.Username != nil {
			opts[utils.AMQPUsername] = *amOpts.Username
		}
		if amOpts.Password != nil {
			opts[utils.AMQPPassword] = *amOpts.Password
		}
	}
	if awsOpts := eeC.Opts.AWS; awsOpts != nil {
		if awsOpts.Region != nil {
			opts[utils.AWSRegion] = *awsOpts.Region
		}
		if awsOpts.Key != nil {
			opts[utils.AWSKey] = *awsOpts.Key
		}
		if awsOpts.Secret != nil {
			opts[utils.AWSSecret] = *awsOpts.Secret
		}
		if awsOpts.Token != nil {
			opts[utils.AWSToken] = *awsOpts.Token
		}
		if awsOpts.SQSQueueID != nil {
			opts[utils.SQSQueueID] = *awsOpts.SQSQueueID
		}
		if awsOpts.S3BucketID != nil {
			opts[utils.S3Bucket] = *awsOpts.S3BucketID
		}
		if awsOpts.S3FolderPath != nil {
			opts[utils.S3FolderPath] = *awsOpts.S3FolderPath
		}
	}
	if natOpts := eeC.Opts.NATS; natOpts != nil {
		if natOpts.JetStream != nil {
			opts[utils.NatsJetStream] = *natOpts.JetStream
		}
		if natOpts.Subject != nil {
			opts[utils.NatsSubject] = *natOpts.Subject
		}
		if natOpts.JWTFile != nil {
			opts[utils.NatsJWTFile] = *natOpts.JWTFile
		}
		if natOpts.SeedFile != nil {
			opts[utils.NatsSeedFile] = *natOpts.SeedFile
		}
		if natOpts.CertificateAuthority != nil {
			opts[utils.NatsCertificateAuthority] = *natOpts.CertificateAuthority
		}
		if natOpts.ClientCertificate != nil {
			opts[utils.NatsClientCertificate] = *natOpts.ClientCertificate
		}
		if natOpts.ClientKey != nil {
			opts[utils.NatsClientKey] = *natOpts.ClientKey
		}
		if natOpts.JetStreamMaxWait != nil {
			opts[utils.NatsJetStreamMaxWait] = natOpts.JetStreamMaxWait.String()
		}
	}
	if rpcOpts := eeC.Opts.RPC; rpcOpts != nil {
		if rpcOpts.RPCCodec != nil {
			opts[utils.RpcCodec] = *rpcOpts.RPCCodec
		}
		if rpcOpts.ServiceMethod != nil {
			opts[utils.ServiceMethod] = *rpcOpts.ServiceMethod
		}
		if rpcOpts.KeyPath != nil {
			opts[utils.KeyPath] = *rpcOpts.KeyPath
		}
		if rpcOpts.CertPath != nil {
			opts[utils.CertPath] = *rpcOpts.CertPath
		}
		if rpcOpts.CAPath != nil {
			opts[utils.CaPath] = *rpcOpts.CAPath
		}
		if rpcOpts.TLS != nil {
			opts[utils.Tls] = *rpcOpts.TLS
		}
		if rpcOpts.ConnIDs != nil {
			opts[utils.ConnIDs] = *rpcOpts.ConnIDs
		}
		if rpcOpts.RPCConnTimeout != nil {
			opts[utils.RpcConnTimeout] = rpcOpts.RPCConnTimeout.String()
		}
		if rpcOpts.RPCReplyTimeout != nil {
			opts[utils.RpcReplyTimeout] = rpcOpts.RPCReplyTimeout.String()
		}
		if rpcOpts.RPCAPIOpts != nil {
			opts[utils.RPCAPIOpts] = rpcOpts.RPCAPIOpts
		}
	}

	flgs := eeC.Flags.SliceFlags()
	if flgs == nil {
		flgs = []string{}
	}
	initialMP = map[string]any{
		utils.IDCfg:                   eeC.ID,
		utils.TypeCfg:                 eeC.Type,
		utils.ExportPathCfg:           eeC.ExportPath,
		utils.TimezoneCfg:             eeC.Timezone,
		utils.FiltersCfg:              eeC.Filters,
		utils.FlagsCfg:                flgs,
		utils.AttributeContextCfg:     eeC.AttributeSCtx,
		utils.AttributeIDsCfg:         eeC.AttributeSIDs,
		utils.SynchronousCfg:          eeC.Synchronous,
		utils.AttemptsCfg:             eeC.Attempts,
		utils.ConcurrentRequestsCfg:   eeC.ConcurrentRequests,
		utils.MetricsResetScheduleCfg: eeC.MetricsResetSchedule,
		utils.FailedPostsDirCfg:       eeC.FailedPostsDir,
		utils.OptsCfg:                 opts,
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
