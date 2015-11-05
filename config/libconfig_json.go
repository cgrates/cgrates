/*
Real-time Charging System for Telecom & ISP environments
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
	Http_skip_tls_verify *bool
	Rounding_decimals    *int
	Dbdata_encoding      *string
	Tpexport_dir         *string
	Http_failed_dir      *string
	Default_reqtype      *string
	Default_category     *string
	Default_tenant       *string
	Default_subject      *string
	Default_timezone     *string
	Connect_attempts     *int
	Reconnects           *int
	Response_cache_ttl   *string
	Internal_ttl         *string
}

// Listen config section
type ListenJsonCfg struct {
	Rpc_json *string
	Rpc_gob  *string
	Http     *string
}

// Database config
type DbJsonCfg struct {
	Db_type           *string
	Db_host           *string
	Db_port           *int
	Db_name           *string
	Db_user           *string
	Db_passwd         *string
	Max_open_conns    *int // Used only in case of storDb
	Max_idle_conns    *int
	Load_history_size *int // Used in case of dataDb to limit the length of the loads history
}

// Balancer config section
type BalancerJsonCfg struct {
	Enabled *bool
}

// Rater config section
type RaterJsonCfg struct {
	Enabled  *bool
	Balancer *string
	Cdrstats *string
	Historys *string
	Pubsubs  *string
	Aliases  *string
	Users    *string
}

// Scheduler config section
type SchedulerJsonCfg struct {
	Enabled *bool
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled         *bool
	Extra_fields    *[]string
	Store_cdrs      *bool
	Rater           *string
	Pubsubs         *string
	Users           *string
	Aliases         *string
	Cdrstats        *string
	Cdr_replication *[]*CdrReplicationJsonCfg
}

type CdrReplicationJsonCfg struct {
	Transport   *string
	Server      *string
	Synchronous *bool
	Attempts    *int
	Cdr_filter  *string
}

// Cdrstats config section
type CdrStatsJsonCfg struct {
	Enabled       *bool
	Save_Interval *string
}

// One cdr field config, used in cdre and cdrc
type CdrFieldJsonCfg struct {
	Tag          *string
	Type         *string
	Cdr_field_id *string
	Metatag_id   *string
	Value        *string
	Width        *int
	Strip        *string
	Padding      *string
	Layout       *string
	Field_filter *string
	Mandatory    *bool
}

// Cdre config section
type CdreJsonCfg struct {
	Cdr_format                    *string
	Field_separator               *string
	Data_usage_multiply_factor    *float64
	Sms_usage_multiply_factor     *float64
	Generic_usage_multiply_factor *float64
	Cost_multiply_factor          *float64
	Cost_rounding_decimals        *int
	Cost_shift_digits             *int
	Mask_destination_id           *string
	Mask_length                   *int
	Export_dir                    *string
	Header_fields                 *[]*CdrFieldJsonCfg
	Content_fields                *[]*CdrFieldJsonCfg
	Trailer_fields                *[]*CdrFieldJsonCfg
}

// Cdrc config section
type CdrcJsonCfg struct {
	Enabled                    *bool
	Dry_run                    *bool
	Cdrs                       *string
	Cdr_format                 *string
	Field_separator            *string
	Timezone                   *string
	Run_delay                  *int
	Data_usage_multiply_factor *float64
	Cdr_in_dir                 *string
	Cdr_out_dir                *string
	Failed_calls_prefix        *string
	Cdr_source_id              *string
	Cdr_filter                 *string
	Max_open_files             *int
	Partial_record_cache       *string
	Header_fields              *[]*CdrFieldJsonCfg
	Content_fields             *[]*CdrFieldJsonCfg
	Trailer_fields             *[]*CdrFieldJsonCfg
}

// SM-Generic config section
type SmGenericJsonCfg struct {
	Enabled           *bool
	Listen_bijson     *string
	Rater             *string
	Cdrs              *string
	Debit_interval    *string
	Min_call_duration *string
	Max_call_duration *string
}

// SM-FreeSWITCH config section
type SmFsJsonCfg struct {
	Enabled                *bool
	Rater                  *string
	Cdrs                   *string
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
	Connections            *[]*FsConnJsonCfg
}

// Represents one connection instance towards FreeSWITCH
type FsConnJsonCfg struct {
	Server     *string
	Password   *string
	Reconnects *int
}

// SM-Kamailio config section
type SmKamJsonCfg struct {
	Enabled           *bool
	Rater             *string
	Cdrs              *string
	Create_cdr        *bool
	Debit_interval    *string
	Min_call_duration *string
	Max_call_duration *string
	Connections       *[]*KamConnJsonCfg
}

// Represents one connection instance towards Kamailio
type KamConnJsonCfg struct {
	Evapi_addr *string
	Reconnects *int
}

// SM-OpenSIPS config section
type SmOsipsJsonCfg struct {
	Enabled                   *bool
	Listen_udp                *string
	Rater                     *string
	Cdrs                      *string
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

// Mailer config section
type MailerJsonCfg struct {
	Server       *string
	Auth_user    *string
	Auth_passwd  *string
	From_address *string
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
