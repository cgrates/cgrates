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
	"encoding/json"
)

// General config section
type GeneralJsonCfg struct {
	Node_id                *string
	Logger                 *string
	Log_level              *int
	Rounding_decimals      *int
	Dbdata_encoding        *string
	Tpexport_dir           *string
	Poster_attempts        *int
	Failed_posts_dir       *string
	Failed_posts_ttl       *string
	Default_request_type   *string
	Default_category       *string
	Default_tenant         *string
	Default_timezone       *string
	Default_caching        *string
	Connect_attempts       *int
	Reconnects             *int
	Max_reconnect_interval *string
	Connect_timeout        *string
	Reply_timeout          *string
	Locking_timeout        *string
	Digest_separator       *string
	Digest_equal           *string
	Rsr_separator          *string
	Max_parallel_conns     *int
}

// Listen config section
type ListenJsonCfg struct {
	Rpc_json     *string
	Rpc_gob      *string
	Http         *string
	Rpc_json_tls *string
	Rpc_gob_tls  *string
	Http_tls     *string
}

type HTTPClientOptsJson struct {
	SkipTLSVerify         *bool   `json:"skipTlsVerify"`
	TLSHandshakeTimeout   *string `json:"tlsHandshakeTimeout"`
	DisableKeepAlives     *bool   `json:"disableKeepAlives"`
	DisableCompression    *bool   `json:"disableCompression"`
	MaxIdleConns          *int    `json:"maxIdleConns"`
	MaxIdleConnsPerHost   *int    `json:"maxIdleConnsPerHost"`
	MaxConnsPerHost       *int    `json:"maxConnsPerHost"`
	IdleConnTimeout       *string `json:"IdleConnTimeout"`
	ResponseHeaderTimeout *string `json:"responseHeaderTimeout"`
	ExpectContinueTimeout *string `json:"expectContinueTimeout"`
	ForceAttemptHTTP2     *bool   `json:"forceAttemptHttp2"`
	DialTimeout           *string `json:"dialTimeout"`
	DialFallbackDelay     *string `json:"dialFallbackDelay"`
	DialKeepAlive         *string `json:"dialKeepAlive"`
}

// HTTP config section
type HTTPJsonCfg struct {
	Json_rpc_url        *string
	Registrars_url      *string
	Ws_url              *string
	Freeswitch_cdrs_url *string
	Http_Cdrs           *string
	Use_basic_auth      *bool
	Auth_users          *map[string]string
	Client_opts         *HTTPClientOptsJson
}

type TlsJsonCfg struct {
	Server_certificate *string
	Server_key         *string
	Server_policy      *int
	Server_name        *string
	Client_certificate *string
	Client_key         *string
	Ca_certificate     *string
}

type DBOptsJson struct {
	RedisMaxConns           *int              `json:"redisMaxConns"`
	RedisConnectAttempts    *int              `json:"redisConnectAttempts"`
	RedisSentinel           *string           `json:"redisSentinel"`
	RedisCluster            *bool             `json:"redisCluster"`
	RedisClusterSync        *string           `json:"redisClusterSync"`
	RedisClusterOndownDelay *string           `json:"redisClusterOndownDelay"`
	RedisConnectTimeout     *string           `json:"redisConnectTimeout"`
	RedisReadTimeout        *string           `json:"redisReadTimeout"`
	RedisWriteTimeout       *string           `json:"redisWriteTimeout"`
	MongoQueryTimeout       *string           `json:"mongoQueryTimeout"`
	RedisTLS                *bool             `json:"redisTLS"`
	RedisClientCertificate  *string           `json:"redisClientCertificate"`
	RedisClientKey          *string           `json:"redisClientKey"`
	RedisCACertificate      *string           `json:"redisCACertificate"`
	SQLMaxOpenConns         *int              `json:"sqlMaxOpenConns"`
	SQLMaxIdleConns         *int              `json:"sqlMaxIdleConns"`
	SQLConnMaxLifetime      *string           `json:"sqlConnMaxLifetime"`
	MySQLDSNParams          map[string]string `json:"mysqlDSNParams"`
	PgSSLMode               *string           `json:"pgSSLMode"`
	MySQLLocation           *string           `json:"mysqlLocation"`
}

// Database config
type DbJsonCfg struct {
	Db_type               *string
	Db_host               *string
	Db_port               *int
	Db_name               *string
	Db_user               *string
	Db_password           *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Remote_conns          *[]string
	Remote_conn_id        *string
	Replication_conns     *[]string
	Replication_filtered  *bool
	Replication_cache     *string
	Items                 *map[string]*ItemOptJson
	Opts                  *DBOptsJson
}

type ItemOptJson struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Remote     *bool
	Replicate  *bool
	// used for ArgDispatcher in case we send this to a dispatcher engine
	Route_id *string
	Api_key  *string
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Apiers_conns    *[]string
}

// Rater config section
type RalsJsonCfg struct {
	Enabled                    *bool
	Thresholds_conns           *[]string
	Stats_conns                *[]string
	Rp_subject_prefix_matching *bool
	Remove_expired             *bool
	Max_computed_usage         *map[string]string
	Max_increments             *int
	Balance_rating_subject     *map[string]string
}

// Scheduler config section
type SchedulerJsonCfg struct {
	Enabled                 *bool
	Cdrs_conns              *[]string
	Thresholds_conns        *[]string
	Stats_conns             *[]string
	Filters                 *[]string
	Dynaprepaid_actionplans *[]string
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled              *bool
	Extra_fields         *[]string
	Store_cdrs           *bool
	Session_cost_retries *int
	Chargers_conns       *[]string
	Rals_conns           *[]string
	Attributes_conns     *[]string
	Thresholds_conns     *[]string
	Stats_conns          *[]string
	Online_cdr_exports   *[]string
	Scheduler_conns      *[]string
	Ees_conns            *[]string
}

// EventReaderSJsonCfg contains the configuration of EventReaderService
type ERsJsonCfg struct {
	Enabled           *bool
	Sessions_conns    *[]string
	Readers           *[]*EventReaderJsonCfg
	Partial_cache_ttl *string
}

type EventReaderOptsJson struct {
	PartialPath                       *string `json:"partialPath"`
	PartialCacheAction                *string `json:"partialCacheAction"`
	PartialOrderField                 *string `json:"partialOrderField"`
	PartialCSVFieldSeparator          *string `json:"partialcsvFieldSeparator"`
	CSVRowLength                      *int    `json:"csvRowLength"`
	CSVFieldSeparator                 *string `json:"csvFieldSeparator"`
	CSVHeaderDefineChar               *string `json:"csvHeaderDefineChar"`
	CSVLazyQuotes                     *bool   `json:"csvLazyQuotes"`
	XMLRootPath                       *string `json:"xmlRootPath"`
	AMQPQueueID                       *string `json:"amqpQueueID"`
	AMQPQueueIDProcessed              *string `json:"amqpQueueIDProcessed"`
	AMQPConsumerTag                   *string `json:"amqpConsumerTag"`
	AMQPExchange                      *string `json:"amqpExchange"`
	AMQPExchangeType                  *string `json:"amqpExchangeType"`
	AMQPRoutingKey                    *string `json:"amqpRoutingKey"`
	AMQPExchangeProcessed             *string `json:"amqpExchangeProcessed"`
	AMQPExchangeTypeProcessed         *string `json:"amqpExchangeTypeProcessed"`
	AMQPRoutingKeyProcessed           *string `json:"amqpRoutingKeyProcessed"`
	KafkaTopic                        *string `json:"kafkaTopic"`
	KafkaGroupID                      *string `json:"kafkaGroupID"`
	KafkaMaxWait                      *string `json:"kafkaMaxWait"`
	KafkaTopicProcessed               *string `json:"kafkaTopicProcessed"`
	SQLDBName                         *string `json:"sqlDBName"`
	SQLTableName                      *string `json:"sqlTableName"`
	PgSSLMode                         *string `json:"pgSSLMode"`
	SQLDBNameProcessed                *string `json:"sqlDBNameProcessed"`
	SQLTableNameProcessed             *string `json:"sqlTableNameProcessed"`
	PgSSLModeProcessed                *string `json:"pgSSLModeProcessed"`
	AWSRegion                         *string `json:"awsRegion"`
	AWSKey                            *string `json:"awsKey"`
	AWSSecret                         *string `json:"awsSecret"`
	AWSToken                          *string `json:"awsToken"`
	AWSRegionProcessed                *string `json:"awsRegionProcessed"`
	AWSKeyProcessed                   *string `json:"awsKeyProcessed"`
	AWSSecretProcessed                *string `json:"awsSecretProcessed"`
	AWSTokenProcessed                 *string `json:"awsTokenProcessed"`
	SQSQueueID                        *string `json:"sqsQueueID"`
	SQSQueueIDProcessed               *string `json:"sqsQueueIDProcessed"`
	S3BucketID                        *string `json:"s3BucketID"`
	S3FolderPathProcessed             *string `json:"s3FolderPathProcessed"`
	S3BucketIDProcessed               *string `json:"s3BucketIDProcessed"`
	NATSJetStream                     *bool   `json:"natsJetStream"`
	NATSConsumerName                  *string `json:"natsConsumerName"`
	NATSSubject                       *string `json:"natsSubject"`
	NATSQueueID                       *string `json:"natsQueueID"`
	NATSJWTFile                       *string `json:"natsJWTFile"`
	NATSSeedFile                      *string `json:"natsSeedFile"`
	NATSCertificateAuthority          *string `json:"natsCertificateAuthority"`
	NATSClientCertificate             *string `json:"natsClientCertificate"`
	NATSClientKey                     *string `json:"natsClientKey"`
	NATSJetStreamMaxWait              *string `json:"natsJetStreamMaxWait"`
	NATSJetStreamProcessed            *bool   `json:"natsJetStreamProcessed"`
	NATSSubjectProcessed              *string `json:"natsSubjectProcessed"`
	NATSJWTFileProcessed              *string `json:"natsJWTFileProcessed"`
	NATSSeedFileProcessed             *string `json:"natsSeedFileProcessed"`
	NATSCertificateAuthorityProcessed *string `json:"natsCertificateAuthorityProcessed"`
	NATSClientCertificateProcessed    *string `json:"natsClientCertificateProcessed"`
	NATSClientKeyProcessed            *string `json:"natsClientKeyProcessed"`
	NATSJetStreamMaxWaitProcessed     *string `json:"natsJetStreamMaxWaitProcessed"`
}

// EventReaderSJsonCfg is the configuration of a single EventReader
type EventReaderJsonCfg struct {
	Id                    *string
	Type                  *string
	Run_delay             *string
	Concurrent_requests   *int
	Source_path           *string
	Processed_path        *string
	Opts                  *EventReaderOptsJson
	Tenant                *string
	Timezone              *string
	Filters               *[]string
	Flags                 *[]string
	Fields                *[]*FcTemplateJsonCfg
	Partial_commit_fields *[]*FcTemplateJsonCfg
	Cache_dump_fields     *[]*FcTemplateJsonCfg
}

// EEsJsonCfg contains the configuration of EventExporterService
type EEsJsonCfg struct {
	Enabled          *bool
	Attributes_conns *[]string
	Cache            *map[string]*CacheParamJsonCfg
	Exporters        *[]*EventExporterJsonCfg
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
	PgSSLMode                *string                `json:"pgSSLMode"`
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
	Attempts            *int
	Failed_posts_dir    *string
	Concurrent_requests *int
	Fields              *[]*FcTemplateJsonCfg
}

// SessionSJsonCfg config section
type SessionSJsonCfg struct {
	Enabled                *bool
	Listen_bijson          *string
	Listen_bigob           *string
	Chargers_conns         *[]string
	Rals_conns             *[]string
	Resources_conns        *[]string
	Thresholds_conns       *[]string
	Stats_conns            *[]string
	Routes_conns           *[]string
	Cdrs_conns             *[]string
	Replication_conns      *[]string
	Attributes_conns       *[]string
	Debit_interval         *string
	Store_session_costs    *bool
	Session_ttl            *string
	Session_ttl_max_delay  *string
	Session_ttl_last_used  *string
	Session_ttl_usage      *string
	Session_ttl_last_usage *string
	Session_indexes        *[]string
	Client_protocol        *float64
	Channel_sync_interval  *string
	Terminate_attempts     *int
	Alterable_fields       *[]string
	Min_dur_low_balance    *string
	Scheduler_conns        *[]string
	Stir                   *STIRJsonCfg
	Default_usage          *map[string]string
}

// FreeSWITCHAgent config section
type FreeswitchAgentJsonCfg struct {
	Enabled                *bool
	Sessions_conns         *[]string
	Subscribe_park         *bool
	Create_cdr             *bool
	Extra_fields           *[]string
	Low_balance_ann_file   *string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
	Max_wait_connection    *string
	Event_socket_conns     *[]*FsConnJsonCfg
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Address                *string
	Password               *string
	Reconnects             *int
	Max_reconnect_interval *string
	Alias                  *string
}

type RPCConnsJson struct {
	Strategy *string
	PoolSize *int
	Conns    *[]*RemoteHostJson
}

// Represents one connection instance towards a rater/cdrs server
type RemoteHostJson struct {
	Id                     *string
	Address                *string
	Transport              *string
	Synchronous            *bool
	Tls                    *bool
	Key_path               *string
	Cert_path              *string
	Ca_path                *string
	Conn_attempts          *int
	Reconnects             *int
	Max_reconnect_interval *string
	Connect_timeout        *string
	Reply_timeout          *string
}

type AstConnJsonCfg struct {
	Alias                  *string
	Address                *string
	User                   *string
	Password               *string
	Connect_attempts       *int
	Reconnects             *int
	Max_reconnect_interval *string
}

type AsteriskAgentJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Create_cdr     *bool
	Asterisk_conns *[]*AstConnJsonCfg
}

type CacheParamJsonCfg struct {
	Limit      *int
	Ttl        *string
	Static_ttl *bool
	Precache   *bool
	Remote     *bool
	Replicate  *bool
}

type CacheJsonCfg struct {
	Partitions        *map[string]*CacheParamJsonCfg
	Replication_conns *[]string
	Remote_conns      *[]string
}

// SM-Kamailio config section
type KamAgentJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Create_cdr     *bool
	Evapi_conns    *[]*KamConnJsonCfg
	Timezone       *string
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Alias                  *string
	Address                *string
	Reconnects             *int
	Max_reconnect_interval *string
}

// Represents one connection instance towards OpenSIPS
type OsipsConnJsonCfg struct {
	Mi_addr    *string
	Reconnects *int
}

// DiameterAgent configuration
type DiameterAgentJsonCfg struct {
	Enabled              *bool
	Listen               *string
	Listen_net           *string
	Dictionaries_path    *string
	Sessions_conns       *[]string
	Origin_host          *string
	Origin_realm         *string
	Vendor_id            *int
	Product_name         *string
	Concurrent_requests  *int
	Synced_conn_requests *bool
	Asr_template         *string
	Rar_template         *string
	Forced_disconnect    *string
	Request_processors   *[]*ReqProcessorJsnCfg
}

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled             *bool
	Listen_net          *string
	Listen_auth         *string
	Listen_acct         *string
	Client_secrets      *map[string]string
	Client_dictionaries *map[string]string
	Sessions_conns      *[]string
	Timezone            *string
	Request_processors  *[]*ReqProcessorJsnCfg
}

// Conecto Agent configuration section
type HttpAgentJsonCfg struct {
	Id                 *string
	Url                *string
	Sessions_conns     *[]string
	Request_payload    *string
	Reply_payload      *string
	Request_processors *[]*ReqProcessorJsnCfg
}

// DNSAgentJsonCfg
type DNSAgentJsonCfg struct {
	Enabled            *bool
	Listen             *string
	Listen_net         *string
	Sessions_conns     *[]string
	Timezone           *string
	Request_processors *[]*ReqProcessorJsnCfg
}

type ReqProcessorJsnCfg struct {
	ID             *string
	Filters        *[]string
	Tenant         *string
	Timezone       *string
	Flags          *[]string
	Request_fields *[]*FcTemplateJsonCfg
	Reply_fields   *[]*FcTemplateJsonCfg
}

type AttributesOptsJson struct {
	ProfileIDs           *[]string `json:"*profileIDs"`
	ProfileRuns          *int      `json:"*profileRuns"`
	ProfileIgnoreFilters *bool     `json:"*profileIgnoreFilters"`
	ProcessRuns          *int      `json:"*processRuns"`
	Context              *string   `json:"*context"`
}

// Attribute service config section
type AttributeSJsonCfg struct {
	Enabled               *bool
	Stats_conns           *[]string
	Resources_conns       *[]string
	Apiers_conns          *[]string
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Any_context           *bool
	Opts                  *AttributesOptsJson
}

// ChargerSJsonCfg service config section
type ChargerSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Attributes_conns      *[]string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
}

type ResourcesOptsJson struct {
	UsageID  *string  `json:"*usageID"`
	UsageTTL *string  `json:"*usageTTL"`
	Units    *float64 `json:"*units"`
}

// ResourceLimiter service config section
type ResourceSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Thresholds_conns      *[]string
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Opts                  *ResourcesOptsJson
}

type StatsOptsJson struct {
	ProfileIDs           *[]string `json:"*profileIDs"`
	ProfileIgnoreFilters *bool     `json:"*profileIgnoreFilters"`
}

// Stat service config section
type StatServJsonCfg struct {
	Enabled                  *bool
	Indexed_selects          *bool
	Store_interval           *string
	Store_uncompressed_limit *int
	Thresholds_conns         *[]string
	String_indexed_fields    *[]string
	Prefix_indexed_fields    *[]string
	Suffix_indexed_fields    *[]string
	Nested_fields            *bool // applies when indexed fields is not defined
	Opts                     *StatsOptsJson
}

type ThresholdsOptsJson struct {
	ProfileIDs           *[]string `json:"*profileIDs"`
	ProfileIgnoreFilters *bool     `json:"*profileIgnoreFilters"`
}

// Threshold service config section
type ThresholdSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Opts                  *ThresholdsOptsJson
}

type RoutesOptsJson struct {
	Context      *string     `json:"*context"`
	IgnoreErrors *bool       `json:"*ignoreErrors"`
	MaxCost      interface{} `json:"*maxCost"`
	Limit        *int        `json:"*limit"`
	Offset       *int        `json:"*offset"`
	ProfileCount *int        `json:"*profileCount"`
}

// Route service config section
type RouteSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
	Resources_conns       *[]string
	Stats_conns           *[]string
	Rals_conns            *[]string
	Default_ratio         *int
	Opts                  *RoutesOptsJson
}

type LoaderJsonDataType struct {
	Type      *string
	File_name *string
	Flags     *[]string
	Fields    *[]*FcTemplateJsonCfg
}

type LoaderJsonCfg struct {
	ID              *string
	Enabled         *bool
	Tenant          *string
	Dry_run         *bool
	Run_delay       *string
	Lockfile_path   *string
	Caches_conns    *[]string
	Field_separator *string
	Tp_in_dir       *string
	Tp_out_dir      *string
	Data            *[]*LoaderJsonDataType
}

// Mailer config section
type MailerJsonCfg struct {
	Server        *string
	Auth_user     *string
	Auth_password *string
	From_address  *string
}

// SureTax config section
type SureTaxJsonCfg struct {
	Url                     *string
	Client_number           *string
	Validation_key          *string
	Business_unit           *string
	Timezone                *string
	Include_local_cost      *bool
	Return_file_code        *string
	Response_group          *string
	Response_type           *string
	Regulatory_code         *string
	Client_tracking         *string
	Customer_number         *string
	Orig_number             *string
	Term_number             *string
	Bill_to_number          *string
	Zipcode                 *string
	Plus4                   *string
	P2PZipcode              *string
	P2PPlus4                *string
	Units                   *string
	Unit_type               *string
	Tax_included            *string
	Tax_situs_rule          *string
	Trans_type_code         *string
	Sales_type_code         *string
	Tax_exemption_code_list *string
}

type DispatcherSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
	Any_subsystem         *bool
	Prevent_loop          *bool
}

type RegistrarCJsonCfg struct {
	Registrars_conns *[]string
	Hosts            []*RemoteHostJsonWithTenant
	Refresh_interval *string
}

type RegistrarCJsonCfgs struct {
	RPC         *RegistrarCJsonCfg
	Dispatchers *RegistrarCJsonCfg
}

type LoaderCfgJson struct {
	Tpid             *string
	Data_path        *string
	Disable_reverse  *bool
	Field_separator  *string
	Caches_conns     *[]string
	Scheduler_conns  *[]string
	Gapi_credentials *json.RawMessage
	Gapi_token       *json.RawMessage
}

type MigratorCfgJson struct {
	Out_dataDB_type     *string
	Out_dataDB_host     *string
	Out_dataDB_port     *string
	Out_dataDB_name     *string
	Out_dataDB_user     *string
	Out_dataDB_password *string
	Out_dataDB_encoding *string
	Out_storDB_type     *string
	Out_storDB_host     *string
	Out_storDB_port     *string
	Out_storDB_name     *string
	Out_storDB_user     *string
	Out_storDB_password *string
	Users_filters       *[]string
	Out_dataDB_opts     *DBOptsJson
	Out_storDB_opts     *DBOptsJson
}

type FcTemplateJsonCfg struct {
	Tag                  *string
	Type                 *string
	Path                 *string
	Attribute_id         *string
	Filters              *[]string
	Value                *string
	Width                *int
	Strip                *string
	Padding              *string
	Mandatory            *bool
	New_branch           *bool
	Timezone             *string
	Blocker              *bool
	Layout               *string
	Cost_shift_digits    *int
	Rounding_decimals    *int
	Mask_destinationd_id *string
	Mask_length          *int
}

// Analyzer service json config section
type AnalyzerSJsonCfg struct {
	Enabled          *bool
	Db_path          *string
	Index_type       *string
	Ttl              *string
	Cleanup_interval *string
}

type ApierJsonCfg struct {
	Enabled          *bool
	Caches_conns     *[]string
	Scheduler_conns  *[]string
	Attributes_conns *[]string
	Ees_conns        *[]string
}

type STIRJsonCfg struct {
	Allowed_attest      *[]string
	Payload_maxduration *string
	Default_attest      *string
	Publickey_path      *string
	Privatekey_path     *string
}

// SIPAgentJsonCfg
type SIPAgentJsonCfg struct {
	Enabled              *bool
	Listen               *string
	Listen_net           *string
	Sessions_conns       *[]string
	Timezone             *string
	Retransmission_timer *string
	Request_processors   *[]*ReqProcessorJsnCfg
}

type ConfigSCfgJson struct {
	Enabled  *bool
	Url      *string
	Root_dir *string
}

type APIBanJsonCfg struct {
	Enabled *bool
	Keys    *[]string
}

type CoreSJsonCfg struct {
	Caps                *int
	Caps_strategy       *string
	Caps_stats_interval *string
	Shutdown_timeout    *string
}
