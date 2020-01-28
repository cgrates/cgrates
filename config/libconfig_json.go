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

// General config section
type GeneralJsonCfg struct {
	Node_id              *string
	Logger               *string
	Log_level            *int
	Http_skip_tls_verify *bool
	Rounding_decimals    *int
	Dbdata_encoding      *string
	Tpexport_dir         *string
	Poster_attempts      *int
	Failed_posts_dir     *string
	Default_request_type *string
	Default_category     *string
	Default_tenant       *string
	Default_timezone     *string
	Default_caching      *string
	Connect_attempts     *int
	Reconnects           *int
	Connect_timeout      *string
	Reply_timeout        *string
	Locking_timeout      *string
	Digest_separator     *string
	Digest_equal         *string
	Rsr_separator        *string
	Max_parralel_conns   *int
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

// HTTP config section
type HTTPJsonCfg struct {
	Json_rpc_url        *string
	Ws_url              *string
	Freeswitch_cdrs_url *string
	Http_Cdrs           *string
	Use_basic_auth      *bool
	Auth_users          *map[string]string
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

// Database config
type DbJsonCfg struct {
	Db_type               *string
	Db_host               *string
	Db_port               *int
	Db_name               *string
	Db_user               *string
	Db_password           *string
	Max_open_conns        *int // Used only in case of storDb
	Max_idle_conns        *int
	Conn_max_lifetime     *int // Used only in case of storDb
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Redis_sentinel        *string
	Query_timeout         *string
	Sslmode               *string // Used only in case of storDb
	Remote_conns          *[]string
	Replication_conns     *[]string
	Items                 *map[string]*ItemOptJson
}

type ItemOptJson struct {
	Remote    *bool
	Replicate *bool
	Ttl       *string
}

// Filters config
type FilterSJsonCfg struct {
	Stats_conns     *[]string
	Resources_conns *[]string
	Rals_conns      *[]string
}

// Rater config section
type RalsJsonCfg struct {
	Enabled                    *bool
	Thresholds_conns           *[]string
	Stats_conns                *[]string
	CacheS_conns               *[]string
	Rp_subject_prefix_matching *bool
	Remove_expired             *bool
	Max_computed_usage         *map[string]string
	Max_increments             *int
	Balance_rating_subject     *map[string]string
}

// Scheduler config section
type SchedulerJsonCfg struct {
	Enabled    *bool
	Cdrs_conns *[]string
	Filters    *[]string
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
}

// Cdre config section
type CdreJsonCfg struct {
	Export_format      *string
	Export_path        *string
	Filters            *[]string
	Tenant             *string
	Attributes_context *string
	Synchronous        *bool
	Attempts           *int
	Field_separator    *string
	Header_fields      *[]*FcTemplateJsonCfg
	Content_fields     *[]*FcTemplateJsonCfg
	Trailer_fields     *[]*FcTemplateJsonCfg
}

// EventReaderSJsonCfg contains the configuration of EventReaderService
type ERsJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Readers        *[]*EventReaderJsonCfg
}

// EventReaderSJsonCfg is the configuration of a single EventReader
type EventReaderJsonCfg struct {
	Id                          *string
	Type                        *string
	Field_separator             *string
	Run_delay                   *int
	Concurrent_requests         *int
	Source_path                 *string
	Processed_path              *string
	Xml_root_path               *string
	Tenant                      *string
	Timezone                    *string
	Filters                     *[]string
	Flags                       *[]string
	Failed_calls_prefix         *string
	Partial_record_cache        *string
	Partial_cache_expiry_action *string
	Header_fields               *[]*FcTemplateJsonCfg
	Fields                      *[]*FcTemplateJsonCfg
	Content_fields              *[]*FcTemplateJsonCfg
	Trailer_fields              *[]*FcTemplateJsonCfg
	Cache_dump_fields           *[]*FcTemplateJsonCfg
}

// SM-Generic config section
type SessionSJsonCfg struct {
	Enabled               *bool
	Listen_bijson         *string
	Chargers_conns        *[]string
	Rals_conns            *[]string
	Resources_conns       *[]string
	Thresholds_conns      *[]string
	Stats_conns           *[]string
	Suppliers_conns       *[]string
	Cdrs_conns            *[]string
	Replication_conns     *[]string
	Attributes_conns      *[]string
	Debit_interval        *string
	Store_session_costs   *bool
	Min_call_duration     *string
	Max_call_duration     *string
	Session_ttl           *string
	Session_ttl_max_delay *string
	Session_ttl_last_used *string
	Session_ttl_usage     *string
	Session_indexes       *[]string
	Client_protocol       *float64
	Channel_sync_interval *string
	Terminate_attempts    *int
	Alterable_fields      *[]string
}

// FreeSWITCHAgent config section
type FreeswitchAgentJsonCfg struct {
	Enabled                *bool
	Sessions_conns         *[]string
	Subscribe_park         *bool
	Create_cdr             *bool
	Extra_fields           *[]string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
	Max_wait_connection    *string
	Event_socket_conns     *[]*FsConnJsonCfg
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Address    *string
	Password   *string
	Reconnects *int
	Alias      *string
}

type RPCConnsJson struct {
	Strategy *string
	PoolSize *int
	Conns    *[]*RemoteHostJson
}

// Represents one connection instance towards a rater/cdrs server
type RemoteHostJson struct {
	Address     *string
	Transport   *string
	Synchronous *bool
	Tls         *bool
}

type AstConnJsonCfg struct {
	Alias            *string
	Address          *string
	User             *string
	Password         *string
	Connect_attempts *int
	Reconnects       *int
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
}

type CacheJsonCfg map[string]*CacheParamJsonCfg

// SM-Kamailio config section
type KamAgentJsonCfg struct {
	Enabled        *bool
	Sessions_conns *[]string
	Create_cdr     *bool
	Evapi_conns    *[]*KamConnJsonCfg
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Alias      *string
	Address    *string
	Reconnects *int
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
	Templates            map[string][]*FcTemplateJsonCfg
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
	Continue       *bool
}

// Attribute service config section
type AttributeSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Process_runs          *int
}

// ChargerSJsonCfg service config section
type ChargerSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Attributes_conns      *[]string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
}

// ResourceLimiter service config section
type ResourceSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Thresholds_conns      *[]string
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
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
	Nested_fields            *bool // applies when indexed fields is not defined
}

// Threshold service config section
type ThresholdSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	Store_interval        *string
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
}

// Supplier service config section
type SupplierSJsonCfg struct {
	Enabled               *bool
	Indexed_selects       *bool
	String_indexed_fields *[]string
	Prefix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
	Resources_conns       *[]string
	Stats_conns           *[]string
	Default_ratio         *int
}

type LoaderJsonDataType struct {
	Type      *string
	File_name *string
	Fields    *[]*FcTemplateJsonCfg
}

type LoaderJsonCfg struct {
	ID              *string
	Enabled         *bool
	Tenant          *string
	Dry_run         *bool
	Run_delay       *int
	Lock_filename   *string
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
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
}

type LoaderCfgJson struct {
	Tpid            *string
	Data_path       *string
	Disable_reverse *bool
	Field_separator *string
	Caches_conns    *[]string
	Scheduler_conns *[]string
}

type MigratorCfgJson struct {
	Out_dataDB_type           *string
	Out_dataDB_host           *string
	Out_dataDB_port           *string
	Out_dataDB_name           *string
	Out_dataDB_user           *string
	Out_dataDB_password       *string
	Out_dataDB_encoding       *string
	Out_dataDB_redis_sentinel *string
	Out_storDB_type           *string
	Out_storDB_host           *string
	Out_storDB_port           *string
	Out_storDB_name           *string
	Out_storDB_user           *string
	Out_storDB_password       *string
	Users_filters             *[]string
}

type FcTemplateJsonCfg struct {
	Tag                  *string
	Type                 *string
	Field_id             *string
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
	Break_on_success     *bool
	Handler_id           *string
	Layout               *string
	Cost_shift_digits    *int
	Rounding_decimals    *int
	Mask_destinationd_id *string
	Mask_length          *int
}

// Analyzer service json config section
type AnalyzerSJsonCfg struct {
	Enabled *bool
}

type ApierJsonCfg struct {
	Caches_conns     *[]string
	Scheduler_conns  *[]string
	Attributes_conns *[]string
}
