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
	Http_skip_tls_veify *bool
	Rounding_decimals   *int
	Dbdata_encoding     *string
	Tpexport_dir        *string
	Default_reqtype     *string
	Default_category    *string
	Default_tenant      *string
	Default_subject     *string
}

// Listen config section
type ListenJsonCfg struct {
	Rpc_json *string
	Rpc_gob  *string
	Http     *string
}

// Database config
type DbJsonCfg struct {
	Db_type        *string
	Db_host        *string
	Db_port        *int
	Db_name        *string
	Db_user        *string
	Db_passwd      *string
	Max_open_conns *int // Used only in case of storDb
	Max_idle_conns *int
}

// Balancer config section
type BalancerJsonCfg struct {
	Enabled *bool
}

// Rater config section
type RaterJsonCfg struct {
	Enabled  *bool
	Balancer *string
}

// Scheduler config section
type SchedulerJsonCfg struct {
	Enabled *bool
}

// Cdrs config section
type CdrsJsonCfg struct {
	Enabled       *bool
	Extra_fields  *[]string
	Mediator      *string
	Cdrstats      *string
	Store_disable *bool
}

// Mediator config section
type MediatorJsonCfg struct {
	Enabled       *bool
	Reconnects    *int
	Rater         *string
	Cdrstats      *string
	Store_disable *bool
}

// Cdrstats config section
type CdrStatsJsonCfg struct {
	Enabled              *bool
	Queue_length         *int
	Time_window          *string
	Metrics              *[]string
	Setup_interval       *[]string
	Tors                 *[]string
	Cdr_hosts            *[]string
	Cdr_sources          *[]string
	Req_types            *[]string
	Directions           *[]string
	Tenants              *[]string
	Categories           *[]string
	Accounts             *[]string
	Subjects             *[]string
	Destination_prefixes *[]string
	Usage_interval       *[]string
	Mediation_run_ids    *[]string
	Rated_accounts       *[]string
	Rated_subjects       *[]string
	Cost_interval        *[]float64
}

// One cdr field config, used in cdre and cdrc
type CdrFieldJsonCfg struct {
	Tag          *string
	Type         *string
	Cdr_field_id *string
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
	Cdr_format                 *string
	Field_separator            *string
	Data_usage_multiply_factor *float64
	Sms_usage_multiply_factor  *float64
	Cost_multiply_factor       *float64
	Cost_rounding_decimals     *int
	Cost_shift_digits          *int
	Mask_destination_id        *string
	Mask_length                *int
	Export_dir                 *string
	Header_fields              *[]*CdrFieldJsonCfg
	Content_fields             *[]*CdrFieldJsonCfg
	Trailer_fields             *[]*CdrFieldJsonCfg
}

// Cdrc config section
type CdrcJsonCfg struct {
	Enabled                    *bool
	Cdrs_address               *string
	Cdr_format                 *string
	Field_separator            *string
	Run_delay                  *int
	Data_usage_multiply_factor *float64
	Cdr_in_dir                 *string
	Cdr_out_dir                *string
	Cdr_source_id              *string
	Cdr_filter                 *string
	Cdr_fields                 *[]*CdrFieldJsonCfg
}

// Session manager config section
type SessionManagerJsonCfg struct {
	Enabled           *bool
	Switch_type       *string
	Rater             *string
	Cdrs              *string
	Reconnects        *int
	Debit_interval    *int
	Min_call_duration *string
	Max_call_duration *string
}

// FreeSWITCH config section
type FSJsonCfg struct {
	Server                 *string
	Password               *string
	Reconnects             *int
	Min_dur_low_balance    *string
	Low_balance_ann_file   *string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
	Cdr_extra_fields       *[]string
}

// Kamailio config section
type KamailioJsonCfg struct {
	Evapi_addr *string
	Reconnects *int
}

// Opensips config section
type OsipsJsonCfg struct {
	Listen_udp                *string
	Mi_addr                   *string
	Events_subscribe_interval *string
	Reconnects                *int
}

// SM-FreeSWITCH config section
type SmFsJsonCfg struct {
	Enabled                *bool
	Rater                  *string
	Cdrs                   *string
	Cdr_extra_fields       *[]string
	Debit_interval         *string
	Min_call_duration      *string
	Max_call_duration      *string
	Min_dur_low_balance    *string
	Low_balance_ann_file   *string
	Empty_balance_context  *string
	Empty_balance_ann_file *string
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
	Debit_interval            *string
	Min_call_duration         *string
	Max_call_duration         *string
	Events_subscribe_interval *string
	Mi_addr                   *string
	Reconnects                *int
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

// History agent config section
type HistAgentJsonCfg struct {
	Enabled *bool
	Server  *string
}

// Mailer config section
type MailerJsonCfg struct {
	Server       *string
	Auth_user    *string
	Auth_passwd  *string
	From_address *string
}
