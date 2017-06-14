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
	Instance_id          *string
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
	Connect_attempts     *int
	Reconnects           *int
	Connect_timeout      *string
	Reply_timeout        *string
	Response_cache_ttl   *string
	Internal_ttl         *string
	Locking_timeout      *string
}

// Listen config section
type ListenJsonCfg struct {
	Rpc_json *string
	Rpc_gob  *string
	Http     *string
}

// HTTP config section
type HTTPJsonCfg struct {
	Json_rpc_url   *string
	Ws_url         *string
	Use_basic_auth *bool
	Auth_users     *map[string]string
}

// Database config
type DbJsonCfg struct {
	Db_type           *string
	Db_host           *string
	Db_port           *int
	Db_name           *string
	Db_user           *string
	Db_password       *string
	Max_open_conns    *int // Used only in case of storDb
	Max_idle_conns    *int
	Load_history_size *int // Used in case of dataDb to limit the length of the loads history
	Cdrs_indexes      *[]string
}

// Rater config section
type RalsJsonCfg struct {
	Enabled                     *bool
	Cdrstats_conns              *[]*HaPoolJsonCfg
	Historys_conns              *[]*HaPoolJsonCfg
	Pubsubs_conns               *[]*HaPoolJsonCfg
	Aliases_conns               *[]*HaPoolJsonCfg
	Users_conns                 *[]*HaPoolJsonCfg
	Rp_subject_prefix_matching  *bool
	Lcr_subject_prefix_matching *bool
}

// Scheduler config section
type SchedulerJsonCfg struct {
	Enabled *bool
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled             *bool
	Extra_fields        *[]string
	Store_cdrs          *bool
	Cdr_account_summary *bool
	Sm_cost_retries     *int
	Rals_conns          *[]*HaPoolJsonCfg
	Pubsubs_conns       *[]*HaPoolJsonCfg
	Users_conns         *[]*HaPoolJsonCfg
	Aliases_conns       *[]*HaPoolJsonCfg
	Cdrstats_conns      *[]*HaPoolJsonCfg
	Online_cdr_exports  *[]string
}

type CdrReplicationJsonCfg struct {
	Transport      *string
	Address        *string
	Synchronous    *bool
	Attempts       *int
	Cdr_filter     *string
	Content_fields *[]*CdrFieldJsonCfg
}

// Cdrstats config section
type CdrStatsJsonCfg struct {
	Enabled       *bool
	Save_Interval *string
}

// One cdr field config, used in cdre and cdrc
type CdrFieldJsonCfg struct {
	Tag                  *string
	Type                 *string
	Field_id             *string
	Handler_id           *string
	Value                *string
	Append               *bool
	Width                *int
	Strip                *string
	Padding              *string
	Layout               *string
	Field_filter         *string
	Mandatory            *bool
	Cost_shift_digits    *int
	Rounding_decimals    *int
	Timezone             *string
	Mask_destinationd_id *string
	Mask_length          *int
	Break_on_success     *bool
}

// Cdre config section
type CdreJsonCfg struct {
	Export_format         *string
	Export_path           *string
	Cdr_filter            *string
	Synchronous           *bool
	Attempts              *int
	Field_separator       *string
	Usage_multiply_factor *map[string]float64
	Cost_multiply_factor  *float64
	Header_fields         *[]*CdrFieldJsonCfg
	Content_fields        *[]*CdrFieldJsonCfg
	Trailer_fields        *[]*CdrFieldJsonCfg
}

// Cdrc config section
type CdrcJsonCfg struct {
	Id                          *string
	Enabled                     *bool
	Dry_run                     *bool
	Cdrs_conns                  *[]*HaPoolJsonCfg
	Cdr_format                  *string
	Field_separator             *string
	Timezone                    *string
	Run_delay                   *int
	Data_usage_multiply_factor  *float64
	Cdr_in_dir                  *string
	Cdr_out_dir                 *string
	Failed_calls_prefix         *string
	Cdr_path                    *string
	Cdr_source_id               *string
	Cdr_filter                  *string
	Continue_on_success         *bool
	Max_open_files              *int
	Partial_record_cache        *string
	Partial_cache_expiry_action *string
	Header_fields               *[]*CdrFieldJsonCfg
	Content_fields              *[]*CdrFieldJsonCfg
	Trailer_fields              *[]*CdrFieldJsonCfg
	Cache_dump_fields           *[]*CdrFieldJsonCfg
}

// SM-Generic config section
type SmGenericJsonCfg struct {
	Enabled               *bool
	Listen_bijson         *string
	Rals_conns            *[]*HaPoolJsonCfg
	Cdrs_conns            *[]*HaPoolJsonCfg
	Smg_replication_conns *[]*HaPoolJsonCfg
	Debit_interval        *string
	Min_call_duration     *string
	Max_call_duration     *string
	Session_ttl           *string
	Session_ttl_max_delay *string
	Session_ttl_last_used *string
	Session_ttl_usage     *string
	Session_indexes       *[]string
}

// SM-FreeSWITCH config section
type SmFsJsonCfg struct {
	Enabled                *bool
	Rals_conns             *[]*HaPoolJsonCfg
	Cdrs_conns             *[]*HaPoolJsonCfg
	Rls_conns              *[]*HaPoolJsonCfg
	Create_cdr             *bool
	Extra_fields           *[]string
	Debit_interval         *string
	Min_call_duration      *string
	Max_call_duration      *string
	Min_dur_low_balance    *string
	Low_balance_ann_file   *string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
	Subscribe_park         *bool
	Channel_sync_interval  *string
	Max_wait_connection    *string
	Event_socket_conns     *[]*FsConnJsonCfg
}

// Represents one connection instance towards a rater/cdrs server
type HaPoolJsonCfg struct {
	Address     *string
	Transport   *string
	Synchronous *bool
}

type AstConnJsonCfg struct {
	Address          *string
	User             *string
	Password         *string
	Connect_attempts *int
	Reconnects       *int
}

type SMAsteriskJsonCfg struct {
	Enabled          *bool
	Sm_generic_conns *[]*HaPoolJsonCfg // Connections towards generic SMf
	Create_cdr       *bool
	Asterisk_conns   *[]*AstConnJsonCfg
}

type CacheParamJsonCfg struct {
	Limit    *int
	Ttl      *string
	Precache *bool
}

type CacheJsonCfg struct {
	Destinations         *CacheParamJsonCfg
	Reverse_destinations *CacheParamJsonCfg
	Rating_plans         *CacheParamJsonCfg
	Rating_profiles      *CacheParamJsonCfg
	Lcr                  *CacheParamJsonCfg
	Cdr_stats            *CacheParamJsonCfg
	Actions              *CacheParamJsonCfg
	Action_plans         *CacheParamJsonCfg
	Account_action_plans *CacheParamJsonCfg
	Action_triggers      *CacheParamJsonCfg
	Shared_groups        *CacheParamJsonCfg
	Aliases              *CacheParamJsonCfg
	Reverse_aliases      *CacheParamJsonCfg
	Derived_chargers     *CacheParamJsonCfg
	Resource_limits      *CacheParamJsonCfg
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Address    *string
	Password   *string
	Reconnects *int
}

// SM-Kamailio config section
type SmKamJsonCfg struct {
	Enabled           *bool
	Rals_conns        *[]*HaPoolJsonCfg
	Cdrs_conns        *[]*HaPoolJsonCfg
	Rls_conns         *[]*HaPoolJsonCfg
	Create_cdr        *bool
	Debit_interval    *string
	Min_call_duration *string
	Max_call_duration *string
	Evapi_conns       *[]*KamConnJsonCfg
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Address    *string
	Reconnects *int
}

// SM-OpenSIPS config section
type SmOsipsJsonCfg struct {
	Enabled                   *bool
	Listen_udp                *string
	Rals_conns                *[]*HaPoolJsonCfg
	Cdrs_conns                *[]*HaPoolJsonCfg
	Create_cdr                *bool
	Debit_interval            *string
	Min_call_duration         *string
	Max_call_duration         *string
	Events_subscribe_interval *string
	Mi_addr                   *string
}

// Represents one connection instance towards OpenSIPS
type OsipsConnJsonCfg struct {
	Mi_addr    *string
	Reconnects *int
}

// DiameterAgent configuration
type DiameterAgentJsonCfg struct {
	Enabled              *bool             // enables the diameter agent: <true|false>
	Listen               *string           // address where to listen for diameter requests <x.y.z.y:1234>
	Dictionaries_dir     *string           // path towards additional dictionaries
	Sm_generic_conns     *[]*HaPoolJsonCfg // Connections towards generic SM
	Pubsubs_conns        *[]*HaPoolJsonCfg // connection towards pubsubs
	Create_cdr           *bool
	Cdr_requires_session *bool
	Debit_interval       *string
	Timezone             *string // timezone for timestamps where not specified <""|UTC|Local|$IANA_TZ_DB>
	Origin_host          *string
	Origin_realm         *string
	Vendor_id            *int
	Product_name         *string
	Request_processors   *[]*DARequestProcessorJsnCfg
}

// One Diameter request processor configuration
type DARequestProcessorJsnCfg struct {
	Id                  *string
	Dry_run             *bool
	Publish_event       *bool
	Request_filter      *string
	Flags               *[]string
	Continue_on_success *bool
	Append_cca          *bool
	CCR_fields          *[]*CdrFieldJsonCfg
	CCA_fields          *[]*CdrFieldJsonCfg
}

// Radius Agent configuration section
type RadiusAgentJsonCfg struct {
	Enabled              *bool
	Listen_net           *string
	Listen_auth          *string
	Listen_acct          *string
	Client_secrets       *map[string]string
	Client_dictionaries  *map[string]string
	Sm_generic_conns     *[]*HaPoolJsonCfg
	Create_cdr           *bool
	Cdr_requires_session *bool
	Timezone             *string
	Request_processors   *[]*RAReqProcessorJsnCfg
}

type RAReqProcessorJsnCfg struct {
	Id                  *string
	Dry_run             *bool
	Request_filter      *string
	Flags               *[]string
	Continue_on_success *bool
	Append_reply        *bool
	Request_fields      *[]*CdrFieldJsonCfg
	Reply_fields        *[]*CdrFieldJsonCfg
}

// History server config section
type HistServJsonCfg struct {
	Enabled       *bool
	History_dir   *string
	Save_interval *string
}

// PubSub server config section
type PubSubServJsonCfg struct {
	Enabled *bool
}

// Aliases server config section
type AliasesServJsonCfg struct {
	Enabled *bool
}

// Users server config section
type UserServJsonCfg struct {
	Enabled *bool
	Indexes *[]string
}

// ResourceLimiter service config section
type ResourceLimiterServJsonCfg struct {
	Enabled             *bool
	Cdrstats_conns      *[]*HaPoolJsonCfg
	Cache_dump_interval *string
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
