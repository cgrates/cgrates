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
	"strings"

	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

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
	Default_ratio         *int
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
	Suffix_indexed_fields *[]string
	Nested_fields         *bool // applies when indexed fields is not defined
	Attributes_conns      *[]string
}

type RegistrarCJsonCfg struct {
	Enabled          *bool
	Registrars_conns *[]string
	Hosts            map[string][]*RemoteHostJson
	Refresh_interval *string
}

type RegistrarCJsonCfgs struct {
	RPC        *RegistrarCJsonCfg
	Dispatcher *RegistrarCJsonCfg
}

type LoaderCfgJson struct {
	Tpid             *string
	Data_path        *string
	Disable_reverse  *bool
	Field_separator  *string
	Caches_conns     *[]string
	Actions_conns    *[]string
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
	Out_dataDB_opts     map[string]interface{}
	Out_storDB_opts     map[string]interface{}
}

// Analyzer service json config section
type AnalyzerSJsonCfg struct {
	Enabled          *bool
	Db_path          *string
	Index_type       *string
	Ttl              *string
	Cleanup_interval *string
}

type RateSJsonCfg struct {
	Enabled                    *bool
	Indexed_selects            *bool
	String_indexed_fields      *[]string
	Prefix_indexed_fields      *[]string
	Suffix_indexed_fields      *[]string
	Nested_fields              *bool // applies when indexed fields is not defined
	Rate_indexed_selects       *bool
	Rate_string_indexed_fields *[]string
	Rate_prefix_indexed_fields *[]string
	Rate_suffix_indexed_fields *[]string
	Rate_nested_fields         *bool // applies when indexed fields is not defined
	Verbosity                  *int
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

// updateInternalConns updates the connection list by specifying the subsystem for internal connections
func updateInternalConns(conns []string, subsystem string) (c []string) {
	subsystem = utils.MetaInternal + utils.ConcatenatedKeySep + subsystem
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		// if we have the connection internal we change the name so we can have internal rpc for each subsystem
		if conn == utils.MetaInternal {
			c[i] = subsystem
		}
	}
	return
}

// updateInternalConns updates the connection list by specifying the subsystem for internal connections
func updateBiRPCInternalConns(conns []string, subsystem string) (c []string) {
	subsystem = utils.ConcatenatedKeySep + subsystem
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		// if we have the connection internal we change the name so we can have internal rpc for each subsystem
		if conn == utils.MetaInternal ||
			conn == rpcclient.BiRPCInternal {
			c[i] += subsystem
		}
	}
	return
}

func getInternalJSONConns(conns []string) (c []string) {
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		if strings.HasPrefix(conn, utils.MetaInternal) {
			c[i] = utils.MetaInternal
		}
	}
	return
}

func getBiRPCInternalJSONConns(conns []string) (c []string) {
	c = make([]string, len(conns))
	for i, conn := range conns {
		c[i] = conn
		if strings.HasPrefix(conn, utils.MetaInternal) {
			c[i] = utils.MetaInternal
		} else if strings.HasPrefix(conn, rpcclient.BiRPCInternal) {
			c[i] = rpcclient.BiRPCInternal
		}
	}
	return
}
