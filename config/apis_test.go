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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

//Section from map
func TestPrepareGeneralSectionFromMap(t *testing.T) {
	section := GeneralJSON
	mp := map[string]interface{}{
		"Node_id":              "",
		"Logger":               "*syslog",
		"Log_level":            7,
		"Rounding_decimals":    5,
		"Dbdata_encoding":      "*msgpack",
		"Tpexport_dir":         "/var/spool/cgrates/tpe",
		"Poster_attempts":      3,
		"Failed_posts_dir":     "/var/spool/cgrates/failed_posts",
		"Failed_posts_ttl":     "5s",
		"Default_request_type": "*rated",
		"Default_category":     "call",
		"Default_tenant":       "cgrates.org",
		"Default_timezone":     "Local",
		"Default_caching":      "*reload",
		"Min_call_duration":    "0s",
		"Max_call_duration":    "3h",
		"Connect_attempts":     5,
		"Reconnects":           -1,
		"Connect_timeout":      "1s",
		"Reply_timeout":        "2s",
		"Locking_timeout":      "0",
		"Digest_separator":     ",",
		"Digest_equal":         ":",
		"Rsr_separator":        ";",
		"Max_parallel_conns":   100,
	}
	expected := &GeneralJsonCfg{
		Node_id:              utils.StringPointer(""),
		Logger:               utils.StringPointer("*syslog"),
		Log_level:            utils.IntPointer(7),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("*msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Poster_attempts:      utils.IntPointer(3),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Failed_posts_ttl:     utils.StringPointer("5s"),
		Default_request_type: utils.StringPointer("*rated"),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Default_caching:      utils.StringPointer("*reload"),
		Connect_attempts:     utils.IntPointer(5),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("0"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Rsr_separator:        utils.StringPointer(";"),
		Max_parallel_conns:   utils.IntPointer(100),
	}
	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRPCConnsSectionFromMap(t *testing.T) {
	section := RPCConnsJSON
	mp := &RPCConnsJson{
		"CONN_1": {
			Strategy: utils.StringPointer("*disconnect"),
			PoolSize: utils.IntPointer(2),
			Conns: &[]*RemoteHostJson{
				{
					Id:      utils.StringPointer("conn1_id"),
					Address: utils.StringPointer("127.0.0.1:8080"),
				},
			},
		},
	}

	expected := &RPCConnsJson{
		"CONN_1": {
			Strategy: utils.StringPointer("*disconnect"),
			PoolSize: utils.IntPointer(2),
			Conns: &[]*RemoteHostJson{
				{
					Id:      utils.StringPointer("conn1_id"),
					Address: utils.StringPointer("127.0.0.1:8080"),
				},
			},
		},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCacheSectionFromMap(t *testing.T) {
	section := CacheJSON
	mp := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			"*resource_profiles": {
				Limit:      utils.IntPointer(-1),
				Ttl:        utils.StringPointer(""),
				Static_ttl: utils.BoolPointer(false),
				Precache:   utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
	}

	expected := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			"*resource_profiles": {
				Limit:      utils.IntPointer(-1),
				Ttl:        utils.StringPointer(""),
				Static_ttl: utils.BoolPointer(false),
				Precache:   utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareListenSectionFromMap(t *testing.T) {
	section := ListenJSON
	mp := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}

	expected := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareHTTPSectionFromMap(t *testing.T) {
	section := HTTPJSON
	mp := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Registrars_url:      utils.StringPointer("/registrar"),
		Ws_url:              utils.StringPointer("/ws"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users: &map[string]string{
			"user1": "pass1",
		},
		Client_opts: map[string]interface{}{},
	}

	expected := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Registrars_url:      utils.StringPointer("/registrar"),
		Ws_url:              utils.StringPointer("/ws"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users: &map[string]string{
			"user1": "pass1",
		},
		Client_opts: map[string]interface{}{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareStorDBSectionFromMap(t *testing.T) {
	section := StorDBJSON
	mp := map[string]interface{}{
		"db_type":               "*mysql",
		"db_host":               "127.0.0.1",
		"db_port":               3306,
		"db_name":               "cgrates",
		"db_user":               "cgrates",
		"db_password":           "",
		"string_indexed_fields": []string{},
		"prefix_indexed_fields": []string{},
		"opts":                  map[string]interface{}{},
		"items":                 map[string]*ItemOptJson{},
	}

	expected := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Opts:                  map[string]interface{}{},
		Items:                 map[string]*ItemOptJson{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDataDBSectionFromMap(t *testing.T) {
	section := DataDBJSON
	mp := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Opts:                  map[string]interface{}{},
		Items:                 map[string]*ItemOptJson{},
	}

	expected := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Opts:                  map[string]interface{}{},
		Items:                 map[string]*ItemOptJson{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareFilterSSectionFromMap(t *testing.T) {
	section := FilterSJSON
	mp := &FilterSJsonCfg{

		Stats_conns:     &[]string{"*birpc"},
		Resources_conns: &[]string{"*birpc"},
		Admins_conns:    &[]string{"*birpc"},
	}

	expected := &FilterSJsonCfg{
		Stats_conns:     &[]string{"*birpc"},
		Resources_conns: &[]string{"*birpc"},
		Admins_conns:    &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCDRsSectionFromMap(t *testing.T) {
	section := CDRsJSON
	mp := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Extra_fields:         &[]string{"extra_field"},
		Session_cost_retries: utils.IntPointer(2),
		Store_cdrs:           utils.BoolPointer(true),
		Stats_conns:          &[]string{"*birpc"},
		Chargers_conns:       &[]string{"*birpc"},
		Thresholds_conns:     &[]string{"*birpc"},
		Attributes_conns:     &[]string{"*birpc"},
		Ees_conns:            &[]string{"*birpc"},
		Online_cdr_exports:   &[]string{"online_cdr_export"},
		Actions_conns:        &[]string{"*birpc"},
	}

	expected := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Extra_fields:         &[]string{"extra_field"},
		Session_cost_retries: utils.IntPointer(2),
		Store_cdrs:           utils.BoolPointer(true),
		Stats_conns:          &[]string{"*birpc"},
		Chargers_conns:       &[]string{"*birpc"},
		Thresholds_conns:     &[]string{"*birpc"},
		Attributes_conns:     &[]string{"*birpc"},
		Ees_conns:            &[]string{"*birpc"},
		Online_cdr_exports:   &[]string{"online_cdr_export"},
		Actions_conns:        &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareERsSectionFromMap(t *testing.T) {
	section := ERsJSON
	mp := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id: utils.StringPointer("ERsID"),
			},
		},
		Partial_cache_ttl:    utils.StringPointer("partial_cache_ttl"),
		Partial_cache_action: utils.StringPointer("partial_cache_action"),
		Partial_path:         utils.StringPointer("/partial/path"),
	}

	expected := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id: utils.StringPointer("ERsID"),
			},
		},
		Partial_cache_ttl:    utils.StringPointer("partial_cache_ttl"),
		Partial_cache_action: utils.StringPointer("partial_cache_action"),
		Partial_path:         utils.StringPointer("/partial/path"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareEEsSectionFromMap(t *testing.T) {
	section := EEsJSON
	mp := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*birpc"},
		Cache: map[string]*CacheParamJsonCfg{
			"CACHE1": {
				Limit: utils.IntPointer(2),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id: utils.StringPointer("EEsID"),
			},
		},
	}

	expected := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*birpc"},
		Cache: map[string]*CacheParamJsonCfg{
			"CACHE1": {
				Limit: utils.IntPointer(2),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id: utils.StringPointer("EEsID"),
			},
		},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSessionSSectionFromMap(t *testing.T) {
	section := SessionSJSON
	mp := &SessionSJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Listen_bijson:          utils.StringPointer("*tcp"),
		Listen_bigob:           utils.StringPointer("*tcp"),
		Stats_conns:            &[]string{"*birpc"},
		Chargers_conns:         &[]string{"*birpc"},
		Thresholds_conns:       &[]string{"*birpc"},
		Attributes_conns:       &[]string{"*birpc"},
		Cdrs_conns:             &[]string{"*birpc"},
		Resources_conns:        &[]string{"*birpc"},
		Routes_conns:           &[]string{"*birpc"},
		Replication_conns:      &[]string{"*birpc"},
		Debit_interval:         utils.StringPointer("2s"),
		Store_session_costs:    utils.BoolPointer(true),
		Session_ttl:            utils.StringPointer("session_ttl"),
		Session_ttl_max_delay:  utils.StringPointer("session_ttl_max_delay"),
		Session_ttl_last_used:  utils.StringPointer("session_ttl_last_used"),
		Session_ttl_usage:      utils.StringPointer("session_ttl_usage"),
		Session_ttl_last_usage: utils.StringPointer("session_ttl_last_usage"),
		Session_indexes:        &[]string{"session_index"},
		Client_protocol:        utils.Float64Pointer(12.2),
		Channel_sync_interval:  utils.StringPointer("channel_sync_interval"),
		Terminate_attempts:     utils.IntPointer(2),
		Alterable_fields:       &[]string{"alterable_field"},
		Min_dur_low_balance:    utils.StringPointer("min_dur_low_balance"),
		Actions_conns:          &[]string{"*birpc"},
		Stir: &STIRJsonCfg{
			Payload_maxduration: utils.StringPointer("2s"),
		},
		Default_usage: map[string]string{},
	}

	expected := &SessionSJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Listen_bijson:          utils.StringPointer("*tcp"),
		Listen_bigob:           utils.StringPointer("*tcp"),
		Stats_conns:            &[]string{"*birpc"},
		Chargers_conns:         &[]string{"*birpc"},
		Thresholds_conns:       &[]string{"*birpc"},
		Attributes_conns:       &[]string{"*birpc"},
		Cdrs_conns:             &[]string{"*birpc"},
		Resources_conns:        &[]string{"*birpc"},
		Routes_conns:           &[]string{"*birpc"},
		Replication_conns:      &[]string{"*birpc"},
		Debit_interval:         utils.StringPointer("2s"),
		Store_session_costs:    utils.BoolPointer(true),
		Session_ttl:            utils.StringPointer("session_ttl"),
		Session_ttl_max_delay:  utils.StringPointer("session_ttl_max_delay"),
		Session_ttl_last_used:  utils.StringPointer("session_ttl_last_used"),
		Session_ttl_usage:      utils.StringPointer("session_ttl_usage"),
		Session_ttl_last_usage: utils.StringPointer("session_ttl_last_usage"),
		Session_indexes:        &[]string{"session_index"},
		Client_protocol:        utils.Float64Pointer(12.2),
		Channel_sync_interval:  utils.StringPointer("channel_sync_interval"),
		Terminate_attempts:     utils.IntPointer(2),
		Alterable_fields:       &[]string{"alterable_field"},
		Min_dur_low_balance:    utils.StringPointer("min_dur_low_balance"),
		Actions_conns:          &[]string{"*birpc"},
		Stir: &STIRJsonCfg{
			Payload_maxduration: utils.StringPointer("2s"),
		},
		Default_usage: map[string]string{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareFreeSWITCHAgentSectionFromMap(t *testing.T) {
	section := FreeSWITCHAgentJSON
	mp := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Sessions_conns:         &[]string{"*birpc"},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(true),
		Extra_fields:           &[]string{"extra_field"},
		Low_balance_ann_file:   utils.StringPointer("low_balance_ann_file"),
		Empty_balance_context:  utils.StringPointer("empty_balance_context"),
		Empty_balance_ann_file: utils.StringPointer("empty_balance_ann_file"),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns:     &[]*FsConnJsonCfg{},
	}

	expected := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Sessions_conns:         &[]string{"*birpc"},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(true),
		Extra_fields:           &[]string{"extra_field"},
		Low_balance_ann_file:   utils.StringPointer("low_balance_ann_file"),
		Empty_balance_context:  utils.StringPointer("empty_balance_context"),
		Empty_balance_ann_file: utils.StringPointer("empty_balance_ann_file"),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns:     &[]*FsConnJsonCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareKamAgentSectionFromMap(t *testing.T) {
	section := KamailioAgentJSON
	mp := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Evapi_conns:    &[]*KamConnJsonCfg{},
		Timezone:       utils.StringPointer("UTC"),
	}

	expected := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Evapi_conns:    &[]*KamConnJsonCfg{},
		Timezone:       utils.StringPointer("UTC"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAsteriskAgentSectionFromMap(t *testing.T) {
	section := AsteriskAgentJSON
	mp := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Asterisk_conns: &[]*AstConnJsonCfg{},
	}

	expected := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Create_cdr:     utils.BoolPointer(true),
		Asterisk_conns: &[]*AstConnJsonCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDiameterAgentSectionFromMap(t *testing.T) {
	section := DiameterAgentJSON
	mp := &DiameterAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:8080"),
		Listen_net:           utils.StringPointer("tcp"),
		Dictionaries_path:    utils.StringPointer("dictionaries/path"),
		Sessions_conns:       &[]string{"*localhost"},
		Origin_host:          utils.StringPointer("origin_host"),
		Origin_realm:         utils.StringPointer("origin_realm"),
		Vendor_id:            utils.IntPointer(2),
		Product_name:         utils.StringPointer("prod_name"),
		Concurrent_requests:  utils.IntPointer(3),
		Synced_conn_requests: utils.BoolPointer(false),
		Asr_template:         utils.StringPointer("asr_template"),
		Rar_template:         utils.StringPointer("rar_template"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	expected := &DiameterAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:8080"),
		Listen_net:           utils.StringPointer("tcp"),
		Dictionaries_path:    utils.StringPointer("dictionaries/path"),
		Sessions_conns:       &[]string{"*localhost"},
		Origin_host:          utils.StringPointer("origin_host"),
		Origin_realm:         utils.StringPointer("origin_realm"),
		Vendor_id:            utils.IntPointer(2),
		Product_name:         utils.StringPointer("prod_name"),
		Concurrent_requests:  utils.IntPointer(3),
		Synced_conn_requests: utils.BoolPointer(false),
		Asr_template:         utils.StringPointer("asr_template"),
		Rar_template:         utils.StringPointer("rar_template"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRadiusAgentSectionFromMap(t *testing.T) {
	section := RadiusAgentJSON
	mp := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(true),
		Listen_net:  utils.StringPointer("tcp"),
		Listen_auth: utils.StringPointer("listen_auth"),
		Listen_acct: utils.StringPointer("listen_acct"),
		Client_secrets: map[string]string{
			"user1": "pass1",
		},
		Client_dictionaries: map[string]string{
			"user1": "dictionary1",
		},
		Sessions_conns:     &[]string{"*birpc"},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	expected := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(true),
		Listen_net:  utils.StringPointer("tcp"),
		Listen_auth: utils.StringPointer("listen_auth"),
		Listen_acct: utils.StringPointer("listen_acct"),
		Client_secrets: map[string]string{
			"user1": "pass1",
		},
		Client_dictionaries: map[string]string{
			"user1": "dictionary1",
		},
		Sessions_conns:     &[]string{"*birpc"},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDNSAgentSectionFromMap(t *testing.T) {
	section := DNSAgentJSON
	mp := &DNSAgentJsonCfg{
		Enabled:            utils.BoolPointer(true),
		Listen:             utils.StringPointer("127.0.0.1:8080"),
		Listen_net:         utils.StringPointer("tcp"),
		Sessions_conns:     &[]string{"*birpc"},
		Timezone:           utils.StringPointer("UTC"),
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	expected := &DNSAgentJsonCfg{
		Enabled:            utils.BoolPointer(true),
		Listen:             utils.StringPointer("127.0.0.1:8080"),
		Listen_net:         utils.StringPointer("tcp"),
		Sessions_conns:     &[]string{"*birpc"},
		Timezone:           utils.StringPointer("UTC"),
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAttributeSSectionFromMap(t *testing.T) {
	section := AttributeSJSON
	mp := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Admins_conns:          &[]string{"*birpc"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Process_runs:          utils.IntPointer(2),
		Any_context:           utils.BoolPointer(false),
	}

	expected := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Admins_conns:          &[]string{"*birpc"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Process_runs:          utils.IntPointer(2),
		Any_context:           utils.BoolPointer(false),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareChargerSSectionFromMap(t *testing.T) {
	section := ChargerSJSON
	mp := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Attributes_conns:      &[]string{"*birpc"},
		Nested_fields:         utils.BoolPointer(true),
	}

	expected := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Attributes_conns:      &[]string{"*birpc"},
		Nested_fields:         utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareResourceSSectionFromMap(t *testing.T) {
	section := ResourceSJSON
	mp := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
	}

	expected := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareStatSSectionFromMap(t *testing.T) {
	section := StatSJSON
	mp := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(true),
		Thresholds_conns:         &[]string{"*birpc"},
		Store_uncompressed_limit: utils.IntPointer(2),
		Store_interval:           utils.StringPointer("1s"),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index3"},
		Nested_fields:            utils.BoolPointer(true),
	}

	expected := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(true),
		Thresholds_conns:         &[]string{"*birpc"},
		Store_uncompressed_limit: utils.IntPointer(2),
		Store_interval:           utils.StringPointer("1s"),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index3"},
		Nested_fields:            utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareThresholdSSectionFromMap(t *testing.T) {
	section := ThresholdSJSON
	mp := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
	}

	expected := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRouteSSectionFromMap(t *testing.T) {
	section := RouteSJSON
	mp := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Stats_conns:           &[]string{"*birpc"},
		Default_ratio:         utils.IntPointer(2),
	}

	expected := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Stats_conns:           &[]string{"*birpc"},
		Default_ratio:         utils.IntPointer(2),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSureTaxSectionFromMap(t *testing.T) {
	section := SureTaxJSON
	mp := &SureTaxJsonCfg{
		Url:                utils.StringPointer("sure_tax_url"),
		Client_number:      utils.StringPointer("client_number"),
		Validation_key:     utils.StringPointer("validation_key"),
		Business_unit:      utils.StringPointer("business_unit"),
		Timezone:           utils.StringPointer("UTC"),
		Include_local_cost: utils.BoolPointer(false),
		Return_file_code:   utils.StringPointer("return_file_code"),
		Response_group:     utils.StringPointer("response_group"),
		Response_type:      utils.StringPointer("response_type"),
		Regulatory_code:    utils.StringPointer("regulatory_code"),
		Client_tracking:    utils.StringPointer("client_tracking"),
		Customer_number:    utils.StringPointer("custom_number"),
		Orig_number:        utils.StringPointer("orig_number"),
	}

	expected := &SureTaxJsonCfg{
		Url:                utils.StringPointer("sure_tax_url"),
		Client_number:      utils.StringPointer("client_number"),
		Validation_key:     utils.StringPointer("validation_key"),
		Business_unit:      utils.StringPointer("business_unit"),
		Timezone:           utils.StringPointer("UTC"),
		Include_local_cost: utils.BoolPointer(false),
		Return_file_code:   utils.StringPointer("return_file_code"),
		Response_group:     utils.StringPointer("response_group"),
		Response_type:      utils.StringPointer("response_type"),
		Regulatory_code:    utils.StringPointer("regulatory_code"),
		Client_tracking:    utils.StringPointer("client_tracking"),
		Customer_number:    utils.StringPointer("custom_number"),
		Orig_number:        utils.StringPointer("orig_number"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDispatcherSSectionFromMap(t *testing.T) {
	section := DispatcherSJSON
	mp := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*birpc"},
		Any_subsystem:         utils.BoolPointer(false),
	}

	expected := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*birpc"},
		Any_subsystem:         utils.BoolPointer(false),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRegistrarCSectionFromMap(t *testing.T) {
	section := RegistrarCJSON
	mp := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Enabled: utils.BoolPointer(false),
		},
		Dispatcher: &RegistrarCJsonCfg{
			Enabled: utils.BoolPointer(false),
		},
	}

	expected := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Enabled: utils.BoolPointer(false),
		},
		Dispatcher: &RegistrarCJsonCfg{
			Enabled: utils.BoolPointer(false),
		},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareLoaderSectionFromMap(t *testing.T) {
	section := LoaderJSON
	mp := &LoaderCfgJson{
		Tpid:            utils.StringPointer("tpid"),
		Data_path:       utils.StringPointer("data_path"),
		Disable_reverse: utils.BoolPointer(true),
		Field_separator: utils.StringPointer("fld_separator"),
		Caches_conns:    &[]string{"*birpc"},
		Actions_conns:   &[]string{"*birpc"},
	}

	expected := &LoaderCfgJson{
		Tpid:            utils.StringPointer("tpid"),
		Data_path:       utils.StringPointer("data_path"),
		Disable_reverse: utils.BoolPointer(true),
		Field_separator: utils.StringPointer("fld_separator"),
		Caches_conns:    &[]string{"*birpc"},
		Actions_conns:   &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareMigratorSectionFromMap(t *testing.T) {
	section := MigratorJSON
	mp := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("8080"),
		Out_dataDB_name:     utils.StringPointer("cgrates"),
		Out_dataDB_user:     utils.StringPointer("cgrates"),
		Out_dataDB_password: utils.StringPointer("CGRateS.org"),
		Out_dataDB_encoding: utils.StringPointer("utf-8"),
		Out_storDB_type:     utils.StringPointer("postgres"),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("8037"),
		Out_storDB_name:     utils.StringPointer("cgrates"),
		Out_storDB_user:     utils.StringPointer("cgrates"),
		Out_storDB_password: utils.StringPointer("CGRateS.org"),
		Out_dataDB_opts:     map[string]interface{}{},
		Out_storDB_opts:     map[string]interface{}{},
	}

	expected := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("8080"),
		Out_dataDB_name:     utils.StringPointer("cgrates"),
		Out_dataDB_user:     utils.StringPointer("cgrates"),
		Out_dataDB_password: utils.StringPointer("CGRateS.org"),
		Out_dataDB_encoding: utils.StringPointer("utf-8"),
		Out_storDB_type:     utils.StringPointer("postgres"),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("8037"),
		Out_storDB_name:     utils.StringPointer("cgrates"),
		Out_storDB_user:     utils.StringPointer("cgrates"),
		Out_storDB_password: utils.StringPointer("CGRateS.org"),
		Out_dataDB_opts:     map[string]interface{}{},
		Out_storDB_opts:     map[string]interface{}{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareTlsSectionFromMap(t *testing.T) {
	section := TlsJSON
	mp := &TlsJsonCfg{
		Server_certificate: utils.StringPointer("server_certificate"),
		Server_key:         utils.StringPointer("server_key"),
		Server_policy:      utils.IntPointer(2),
		Server_name:        utils.StringPointer("server_name"),
		Client_certificate: utils.StringPointer("client_certificate"),
		Client_key:         utils.StringPointer("client_key"),
		Ca_certificate:     utils.StringPointer("ca_certificate"),
	}

	expected := &TlsJsonCfg{
		Server_certificate: utils.StringPointer("server_certificate"),
		Server_key:         utils.StringPointer("server_key"),
		Server_policy:      utils.IntPointer(2),
		Server_name:        utils.StringPointer("server_name"),
		Client_certificate: utils.StringPointer("client_certificate"),
		Client_key:         utils.StringPointer("client_key"),
		Ca_certificate:     utils.StringPointer("ca_certificate"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAnalyzerSSectionFromMap(t *testing.T) {
	section := AnalyzerSJSON
	mp := &AnalyzerSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Db_path:          utils.StringPointer("/db/path"),
		Index_type:       utils.StringPointer("index_type"),
		Ttl:              utils.StringPointer("Ttl"),
		Cleanup_interval: utils.StringPointer("2s"),
	}

	expected := &AnalyzerSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Db_path:          utils.StringPointer("/db/path"),
		Index_type:       utils.StringPointer("index_type"),
		Ttl:              utils.StringPointer("Ttl"),
		Cleanup_interval: utils.StringPointer("2s"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAdminSSectionFromMap(t *testing.T) {
	section := AdminSJSON
	mp := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Caches_conns:     &[]string{"*birpc"},
		Actions_conns:    &[]string{"*birpc"},
		Attributes_conns: &[]string{"*birpc"},
		Ees_conns:        &[]string{"*birpc"},
	}

	expected := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Caches_conns:     &[]string{"*birpc"},
		Actions_conns:    &[]string{"*birpc"},
		Attributes_conns: &[]string{"*birpc"},
		Ees_conns:        &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRateSSectionFromMap(t *testing.T) {
	section := RateSJSON
	mp := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Indexed_selects:            utils.BoolPointer(true),
		String_indexed_fields:      &[]string{"*req.index1"},
		Prefix_indexed_fields:      &[]string{"*req.index2"},
		Suffix_indexed_fields:      &[]string{"*req.index3"},
		Nested_fields:              utils.BoolPointer(true),
		Rate_indexed_selects:       utils.BoolPointer(false),
		Rate_string_indexed_fields: &[]string{"*req.index1"},
		Rate_prefix_indexed_fields: &[]string{"*req.index2"},
		Rate_suffix_indexed_fields: &[]string{"*req.index3"},
		Rate_nested_fields:         utils.BoolPointer(false),
		Verbosity:                  utils.IntPointer(2),
	}

	expected := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Indexed_selects:            utils.BoolPointer(true),
		String_indexed_fields:      &[]string{"*req.index1"},
		Prefix_indexed_fields:      &[]string{"*req.index2"},
		Suffix_indexed_fields:      &[]string{"*req.index3"},
		Nested_fields:              utils.BoolPointer(true),
		Rate_indexed_selects:       utils.BoolPointer(false),
		Rate_string_indexed_fields: &[]string{"*req.index1"},
		Rate_prefix_indexed_fields: &[]string{"*req.index2"},
		Rate_suffix_indexed_fields: &[]string{"*req.index3"},
		Rate_nested_fields:         utils.BoolPointer(false),
		Verbosity:                  utils.IntPointer(2),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSIPAgentSectionFromMap(t *testing.T) {
	section := SIPAgentJSON
	mp := &SIPAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:8080"),
		Listen_net:           utils.StringPointer("tcp"),
		Sessions_conns:       &[]string{"*birpc"},
		Timezone:             utils.StringPointer("UTC"),
		Retransmission_timer: utils.StringPointer("1s"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	expected := &SIPAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:8080"),
		Listen_net:           utils.StringPointer("tcp"),
		Sessions_conns:       &[]string{"*birpc"},
		Timezone:             utils.StringPointer("UTC"),
		Retransmission_timer: utils.StringPointer("1s"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPreapreTemplatesSectionFromMap(t *testing.T) {
	section := TemplatesJSON
	mp := &FcTemplatesJsonCfg{
		"TEMPLATE_1": {
			{
				Type:  utils.StringPointer("template_type"),
				Value: utils.StringPointer("template_value"),
				Tag:   utils.StringPointer("template_tag"),
			},
		},
	}

	expected := &FcTemplatesJsonCfg{
		"TEMPLATE_1": {
			{
				Type:  utils.StringPointer("template_type"),
				Value: utils.StringPointer("template_value"),
				Tag:   utils.StringPointer("template_tag"),
			},
		},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareConfigSSectionFromMap(t *testing.T) {
	section := ConfigSJSON
	mp := &ConfigSCfgJson{
		Enabled:  utils.BoolPointer(true),
		Url:      utils.StringPointer("/config/url"),
		Root_dir: utils.StringPointer("/root/dir"),
	}

	expected := &ConfigSCfgJson{
		Enabled:  utils.BoolPointer(true),
		Url:      utils.StringPointer("/config/url"),
		Root_dir: utils.StringPointer("/root/dir"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAPIBanSectionFromMap(t *testing.T) {
	section := APIBanJSON
	mp := &APIBanJsonCfg{
		Enabled: utils.BoolPointer(true),
		Keys:    &[]string{"key1", "key2"},
	}

	expected := &APIBanJsonCfg{
		Enabled: utils.BoolPointer(true),
		Keys:    &[]string{"key1", "key2"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCoreSSectionFromMap(t *testing.T) {
	section := CoreSJSON
	mp := &CoreSJsonCfg{
		Caps:                utils.IntPointer(3),
		Caps_strategy:       utils.StringPointer("caps_strategy"),
		Caps_stats_interval: utils.StringPointer("caps_stats_interval"),
		Shutdown_timeout:    utils.StringPointer("shutdown_timeout"),
	}

	expected := &CoreSJsonCfg{
		Caps:                utils.IntPointer(3),
		Caps_strategy:       utils.StringPointer("caps_strategy"),
		Caps_stats_interval: utils.StringPointer("caps_stats_interval"),
		Shutdown_timeout:    utils.StringPointer("shutdown_timeout"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareActionSSectionFromMap(t *testing.T) {
	section := ActionSJSON
	mp := &ActionSJsonCfg{
		Enabled:                   utils.BoolPointer(true),
		Cdrs_conns:                &[]string{"*birpc"},
		Ees_conns:                 &[]string{"*birpc"},
		Thresholds_conns:          &[]string{"*birpc"},
		Stats_conns:               &[]string{"*birpc"},
		Accounts_conns:            &[]string{"*birpc"},
		Tenants:                   &[]string{"cgrates"},
		Indexed_selects:           utils.BoolPointer(false),
		String_indexed_fields:     &[]string{"*req.index1"},
		Prefix_indexed_fields:     &[]string{"*req.index2"},
		Suffix_indexed_fields:     &[]string{"*req.index3"},
		Nested_fields:             utils.BoolPointer(true),
		Dynaprepaid_actionprofile: &[]string{"action_profile_1"},
	}

	expected := &ActionSJsonCfg{
		Enabled:                   utils.BoolPointer(true),
		Cdrs_conns:                &[]string{"*birpc"},
		Ees_conns:                 &[]string{"*birpc"},
		Thresholds_conns:          &[]string{"*birpc"},
		Stats_conns:               &[]string{"*birpc"},
		Accounts_conns:            &[]string{"*birpc"},
		Tenants:                   &[]string{"cgrates"},
		Indexed_selects:           utils.BoolPointer(false),
		String_indexed_fields:     &[]string{"*req.index1"},
		Prefix_indexed_fields:     &[]string{"*req.index2"},
		Suffix_indexed_fields:     &[]string{"*req.index3"},
		Nested_fields:             utils.BoolPointer(true),
		Dynaprepaid_actionprofile: &[]string{"action_profile_1"},
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAccountSSectionFromMap(t *testing.T) {
	section := AccountSJSON
	mp := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*localhost"},
		Rates_conns:           &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Max_iterations:        utils.IntPointer(2),
		Max_usage:             utils.StringPointer("3s"),
	}

	expected := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*localhost"},
		Rates_conns:           &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
		Max_iterations:        utils.IntPointer(2),
		Max_usage:             utils.StringPointer("3s"),
	}

	if cfgSec, err := prepareSectionFromMap(section, mp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(mp), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSectionFromMapError(t *testing.T) {
	section := LoaderSJSON
	// mp := map[string]interface{}{
	// 	"enabled":  false,
	// 	"url":      "/configs/",
	// 	"root_dir": "/var/spool/cgrates/configs",
	// }
	cfgJSONStr := `{
		"loaders": {
			"id": "*default",									
			"enabled": false,									
			"tenant": "",										
			"dry_run": false,									
			"run_delay": "0",									
			"lock_filename": ".cgr.lck",						
			"caches_conns": ["*internal"],
			"field_separator": ",",								
			"tp_in_dir": "/var/spool/cgrates/loader/in",		
			"tp_out_dir": "/var/spool/cgrates/loader/out",				
			"data":[
				{
					"type": "*attributes",						
					"file_name": "Attributes.csv",				
					"fields": [
						{"tag": "TenantID", "path": "Tenant", "type": "*variable", "value": "~*req.0", "mandatory": true},
						{"tag": "ProfileID", "path": "ID", "type": "*variable", "value": "~*req.1", "mandatory": true},
						{"tag": "FilterIDs", "path": "FilterIDs", "type": "*variable", "value": "~*req.2"},
						{"tag": "Weight", "path": "Weight", "type": "*variable", "value": "~*req.3"},
						{"tag": "AttributeFilterIDs", "path": "AttributeFilterIDs", "type": "*variable", "value": "~*req.4"},
						{"tag": "Path", "path": "Path", "type": "*variable", "value": "~*req.5"},
						{"tag": "Type", "path": "Type", "type": "*variable", "value": "~*req.6"},
						{"tag": "Value", "path": "Value", "type": "*variable", "value": "~*req.7"},
						{"tag": "Blocker", "path": "Blocker", "type": "*variable", "value": "~*req.8"},
					],
				},
			],
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	errExpect := "json: Unmarshal(non-pointer []*config.LoaderJsonCfg)"
	if _, err := prepareSectionFromMap(section, cgrCfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected error: %v", err)
	}
}

//Section from DB
func TestPrepareGeneralSectionFromDB(t *testing.T) {
	section := GeneralJSON
	// var db ConfigDB
	cfgJSONStr := `{
		"general": {
			"Node_id":              "",
			"Logger":               "*syslog",
			"Log_level":            7,
			"Rounding_decimals":    5,
			"Dbdata_encoding":      "*msgpack",
			"Tpexport_dir":         "/var/spool/cgrates/tpe",
			"Poster_attempts":      3,
			"Failed_posts_dir":     "/var/spool/cgrates/failed_posts",
			"Failed_posts_ttl":     "5s",
			"Default_request_type": "*rated",
			"Default_category":     "call",
			"Default_tenant":       "cgrates.org",
			"Default_timezone":     "Local",
			"Default_caching":      "*reload",
			"Min_call_duration":    "0s",
			"Max_call_duration":    "3h",
			"Connect_attempts":     5,
			"Reconnects":           -1,
			"Connect_timeout":      "1s",
			"Reply_timeout":        "2s",
			"Locking_timeout":      "0",
			"Digest_separator":     ",",
			"Digest_equal":         ":",
			"Rsr_separator":        ";",
			"Max_parallel_conns":   100,
		},
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &GeneralJsonCfg{
		Node_id:              utils.StringPointer(""),
		Logger:               utils.StringPointer("*syslog"),
		Log_level:            utils.IntPointer(7),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("*msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Poster_attempts:      utils.IntPointer(3),
		Failed_posts_dir:     utils.StringPointer("/var/spool/cgrates/failed_posts"),
		Failed_posts_ttl:     utils.StringPointer("5s"),
		Default_request_type: utils.StringPointer("*rated"),
		Default_category:     utils.StringPointer("call"),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Default_caching:      utils.StringPointer("*reload"),
		Connect_attempts:     utils.IntPointer(5),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("0"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Rsr_separator:        utils.StringPointer(";"),
		Max_parallel_conns:   utils.IntPointer(100),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCacheSectionFromDB(t *testing.T) {
	section := CacheJSON
	cfgJSONStr := `{
		"caches":{
			"partitions": {
				"*resource_profiles": {"limit": -1, "ttl": "", "static_ttl": false, "precache": false, "replicate": false},
			},
		},
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &CacheJsonCfg{
		Partitions: map[string]*CacheParamJsonCfg{
			"*resource_profiles": {
				Limit:      utils.IntPointer(-1),
				Ttl:        utils.StringPointer(""),
				Static_ttl: utils.BoolPointer(false),
				Precache:   utils.BoolPointer(false),
				Replicate:  utils.BoolPointer(false),
			},
		},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareListenSectionFromDB(t *testing.T) {
	section := ListenJSON
	cfgJSONStr := `{
		"listen": {
			"rpc_json": "127.0.0.1:2012",			
			"rpc_gob": "127.0.0.1:2013",			
			"http": "127.0.0.1:2080",				
			"rpc_json_tls" : "127.0.0.1:2022",		
			"rpc_gob_tls": "127.0.0.1:2023",		
			"http_tls": "127.0.0.1:2280",			
		},
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ListenJsonCfg{
		Rpc_json:     utils.StringPointer("127.0.0.1:2012"),
		Rpc_gob:      utils.StringPointer("127.0.0.1:2013"),
		Http:         utils.StringPointer("127.0.0.1:2080"),
		Rpc_json_tls: utils.StringPointer("127.0.0.1:2022"),
		Rpc_gob_tls:  utils.StringPointer("127.0.0.1:2023"),
		Http_tls:     utils.StringPointer("127.0.0.1:2280"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareHTTPSectionFromDB(t *testing.T) {
	section := HTTPJSON
	cfgJSONStr := `{
		"http": {
			"json_rpc_url": "/jsonrpc",								
			"registrars_url": "/registrar",							
			"ws_url": "/ws",										
			"freeswitch_cdrs_url": "/freeswitch_json",				
			"http_cdrs": "/cdr_http",								
			"use_basic_auth": false,								
			"auth_users": {},
		},										
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &HTTPJsonCfg{
		Json_rpc_url:        utils.StringPointer("/jsonrpc"),
		Registrars_url:      utils.StringPointer("/registrar"),
		Ws_url:              utils.StringPointer("/ws"),
		Freeswitch_cdrs_url: utils.StringPointer("/freeswitch_json"),
		Http_Cdrs:           utils.StringPointer("/cdr_http"),
		Use_basic_auth:      utils.BoolPointer(false),
		Auth_users:          &map[string]string{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareStorDBSectionFromDB(t *testing.T) {
	section := StorDBJSON
	cfgJSONStr := `{
		"stor_db": {								
			"db_type": "*mysql",					
			"db_host": "127.0.0.1",					
			"db_port": 3306,						
			"db_name": "cgrates",					
			"db_user": "cgrates",					
			"db_password": "",						
			"string_indexed_fields": [],			
			"prefix_indexed_fields":[],				
			"opts": {},
			"items":{},
		},									
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &DbJsonCfg{
		Db_type:               utils.StringPointer("*mysql"),
		Db_host:               utils.StringPointer("127.0.0.1"),
		Db_port:               utils.IntPointer(3306),
		Db_name:               utils.StringPointer("cgrates"),
		Db_user:               utils.StringPointer("cgrates"),
		Db_password:           utils.StringPointer(""),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Opts:                  map[string]interface{}{},
		Items:                 map[string]*ItemOptJson{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDataDBSectionFromDB(t *testing.T) {
	section := DataDBJSON
	cfgJSONStr := `{
		"data_db": {								
			"db_type": "*redis",					
			"db_host": "127.0.0.1",					
			"db_port": 6379, 						
			"db_name": "10", 						
			"db_user": "cgrates", 					
			"db_password": "", 						
			"remote_conns":[],						 
			"remote_conn_id": "",					
			"replication_conns":[],					
			"replication_filtered": false, 			
			"replication_cache": "", 				
			"items":{},
			"opts":{}
		},								
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &DbJsonCfg{
		Db_type:              utils.StringPointer("*redis"),
		Db_host:              utils.StringPointer("127.0.0.1"),
		Db_port:              utils.IntPointer(6379),
		Db_name:              utils.StringPointer("10"),
		Db_user:              utils.StringPointer("cgrates"),
		Db_password:          utils.StringPointer(""),
		Remote_conns:         &[]string{},
		Remote_conn_id:       utils.StringPointer(""),
		Replication_conns:    &[]string{},
		Replication_filtered: utils.BoolPointer(false),
		Replication_cache:    utils.StringPointer(""),
		Opts:                 map[string]interface{}{},
		Items:                map[string]*ItemOptJson{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareFilterSSectionFromDb(t *testing.T) {
	section := FilterSJSON
	cfgJSONStr := `{
		"filters": {								
			"stats_conns": ["*birpc"],						
			"resources_conns": ["*birpc"],					
			"admins_conns": ["*birpc"],						
		},								
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &FilterSJsonCfg{
		Stats_conns:     &[]string{"*birpc"},
		Resources_conns: &[]string{"*birpc"},
		Admins_conns:    &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCDRsSectionFromDB(t *testing.T) {
	section := CDRsJSON
	cfgJSONStr := `{
		"cdrs": {									
			"enabled": true,						
			"extra_fields": ["extra_field"],						
			"store_cdrs": true,						
			"session_cost_retries": 2,				
			"chargers_conns": ["*birpc"],					
			"attributes_conns": ["*birpc"],					
			"thresholds_conns": ["*birpc"],					
			"stats_conns": ["*birpc"],						
			"online_cdr_exports":["online_cdr_export"],				
			"actions_conns": ["*birpc"],					
			"ees_conns": ["*birpc"],						
		},								
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Extra_fields:         &[]string{"extra_field"},
		Session_cost_retries: utils.IntPointer(2),
		Store_cdrs:           utils.BoolPointer(true),
		Stats_conns:          &[]string{"*birpc"},
		Chargers_conns:       &[]string{"*birpc"},
		Thresholds_conns:     &[]string{"*birpc"},
		Attributes_conns:     &[]string{"*birpc"},
		Ees_conns:            &[]string{"*birpc"},
		Online_cdr_exports:   &[]string{"online_cdr_export"},
		Actions_conns:        &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareERsSectionFromDB(t *testing.T) {
	section := ERsJSON
	cfgJSONStr := `{
		"ers": {														
			"enabled": true,											
			"sessions_conns":["*birpc"],								
			"partial_cache_ttl": "1s",										
			"partial_cache_action": "*none",							
			"partial_path": "/var/spool/cgrates/ers/partial",		
			"readers": [
				{
					"id": "*default",									
				},
			],
		},								
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ERsJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc"},
		Readers: &[]*EventReaderJsonCfg{
			{
				Id: utils.StringPointer("*default"),
			},
		},
		Partial_cache_ttl:    utils.StringPointer("1s"),
		Partial_cache_action: utils.StringPointer("*none"),
		Partial_path:         utils.StringPointer("/var/spool/cgrates/ers/partial"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareEEsSectionFromDB(t *testing.T) {
	section := EEsJSON
	cfgJSONStr := `{
		"ees": {									
			"enabled": true,						
			"attributes_conns":["*birpc"],					
			"cache": {
				"*file_csv": {"limit": -1, "ttl": "5s", "static_ttl": false},
			},
			"exporters": [
				{
					"id": "*default",																	
				},
			],
		},
									
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &EEsJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Attributes_conns: &[]string{"*birpc"},
		Cache: map[string]*CacheParamJsonCfg{
			"*file_csv": {
				Limit:      utils.IntPointer(-1),
				Ttl:        utils.StringPointer("5s"),
				Static_ttl: utils.BoolPointer(false),
			},
		},
		Exporters: &[]*EventExporterJsonCfg{
			{
				Id: utils.StringPointer("*default"),
			},
		},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSessionSSectionFromDB(t *testing.T) {
	section := SessionSJSON
	cfgJSONStr := `{
		"sessions": {
			"enabled": true,						
			"listen_bijson": "127.0.0.1:2014",		
			"listen_bigob": "",						
			"chargers_conns": ["*birpc"],					
			"cdrs_conns": ["*birpc"],						
			"resources_conns": ["*birpc"],					
			"thresholds_conns": ["*birpc"],					
			"stats_conns": ["*birpc"],						
			"routes_conns": ["*birpc"],						
			"attributes_conns": ["*birpc"],					
			"replication_conns": [],				
			"debit_interval": "2s",					
			"store_session_costs": false,			
			"default_usage":{},
			"session_ttl": "0s",					
			"session_ttl_max_delay": "",			
			"session_ttl_last_used": "",			
			"session_ttl_usage": "",				
			"session_last_usage": "",				
			"session_indexes": ["session_index"],					
			"client_protocol": 1.0,					
			"channel_sync_interval": "0",			
			"terminate_attempts": 2,				
			"alterable_fields": [],					
			"min_dur_low_balance": "5s",			
			"stir": {},					
		},						
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &SessionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Listen_bijson:         utils.StringPointer("127.0.0.1:2014"),
		Listen_bigob:          utils.StringPointer(""),
		Stats_conns:           &[]string{"*birpc"},
		Chargers_conns:        &[]string{"*birpc"},
		Thresholds_conns:      &[]string{"*birpc"},
		Attributes_conns:      &[]string{"*birpc"},
		Cdrs_conns:            &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Routes_conns:          &[]string{"*birpc"},
		Replication_conns:     &[]string{},
		Debit_interval:        utils.StringPointer("2s"),
		Store_session_costs:   utils.BoolPointer(false),
		Session_ttl:           utils.StringPointer("0s"),
		Session_ttl_max_delay: utils.StringPointer(""),
		Session_ttl_last_used: utils.StringPointer(""),
		Session_ttl_usage:     utils.StringPointer(""),
		Session_indexes:       &[]string{"session_index"},
		Client_protocol:       utils.Float64Pointer(1.0),
		Channel_sync_interval: utils.StringPointer("0"),
		Terminate_attempts:    utils.IntPointer(2),
		Alterable_fields:      &[]string{},
		Min_dur_low_balance:   utils.StringPointer("5s"),
		Stir:                  &STIRJsonCfg{},
		Default_usage:         map[string]string{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareFreeSWITCHAgentSectionFromDB(t *testing.T) {
	section := FreeSWITCHAgentJSON
	cfgJSONStr := `{
		"freeswitch_agent": {
			"enabled": true,						
			"sessions_conns": ["*birpc_internal"],
			"subscribe_park": true,					
			"create_cdr": false,					
			"extra_fields": ["extra_field"],						
			"low_balance_ann_file": "low_balance_ann_file",				
			"empty_balance_context": "",			
			"empty_balance_ann_file": "",			
			"max_wait_connection": "2s",			
			"event_socket_conns":[],
		},				
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &FreeswitchAgentJsonCfg{
		Enabled:                utils.BoolPointer(true),
		Sessions_conns:         &[]string{"*birpc_internal"},
		Subscribe_park:         utils.BoolPointer(true),
		Create_cdr:             utils.BoolPointer(false),
		Extra_fields:           &[]string{"extra_field"},
		Low_balance_ann_file:   utils.StringPointer("low_balance_ann_file"),
		Empty_balance_context:  utils.StringPointer(""),
		Empty_balance_ann_file: utils.StringPointer(""),
		Max_wait_connection:    utils.StringPointer("2s"),
		Event_socket_conns:     &[]*FsConnJsonCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareKamAgentSectionFromDB(t *testing.T) {
	section := KamailioAgentJSON
	cfgJSONStr := `{
		"kamailio_agent": {
			"enabled": true,						
			"sessions_conns": ["*birpc_internal"],
			"create_cdr": false,					
			"timezone": "UTC",							
			"evapi_conns":[							
				{"address": "127.0.0.1:8448", "reconnects": 5}
			],
		},			
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &KamAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc_internal"},
		Create_cdr:     utils.BoolPointer(false),
		Evapi_conns: &[]*KamConnJsonCfg{
			{
				Address:    utils.StringPointer("127.0.0.1:8448"),
				Reconnects: utils.IntPointer(5),
			},
		},
		Timezone: utils.StringPointer("UTC"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAsteriskAgentSectionFromDB(t *testing.T) {
	section := AsteriskAgentJSON
	cfgJSONStr := `{
		"asterisk_agent": {
			"enabled": true,						
			"sessions_conns": ["*birpc_internal"],
			"create_cdr": false,					
			"asterisk_conns":[],
		},			
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &AsteriskAgentJsonCfg{
		Enabled:        utils.BoolPointer(true),
		Sessions_conns: &[]string{"*birpc_internal"},
		Create_cdr:     utils.BoolPointer(false),
		Asterisk_conns: &[]*AstConnJsonCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDiameterAgentSectionFromDB(t *testing.T) {
	section := DiameterAgentJSON
	cfgJSONStr := `{
		"diameter_agent": {
			"enabled": true,											
			"listen": "127.0.0.1:3868",									
			"listen_net": "tcp",										
			"dictionaries_path": "/usr/share/cgrates/diameter/dict/",	
			"sessions_conns": ["*localhost"],
			"origin_host": "CGR-DA",									
			"origin_realm": "cgrates.org",								
			"vendor_id": 2,												
			"product_name": "CGRateS",									
			"concurrent_requests": 3,									
			"synced_conn_requests": false,								
			"asr_template": "",											
			"rar_template": "",											
			"forced_disconnect": "*none",								
			"request_processors": [],
		},			
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &DiameterAgentJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Listen:               utils.StringPointer("127.0.0.1:3868"),
		Listen_net:           utils.StringPointer("tcp"),
		Dictionaries_path:    utils.StringPointer("/usr/share/cgrates/diameter/dict/"),
		Sessions_conns:       &[]string{"*localhost"},
		Origin_host:          utils.StringPointer("CGR-DA"),
		Origin_realm:         utils.StringPointer("cgrates.org"),
		Vendor_id:            utils.IntPointer(2),
		Product_name:         utils.StringPointer("CGRateS"),
		Concurrent_requests:  utils.IntPointer(3),
		Synced_conn_requests: utils.BoolPointer(false),
		Asr_template:         utils.StringPointer(""),
		Rar_template:         utils.StringPointer(""),
		Forced_disconnect:    utils.StringPointer("*none"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRadiusAgentSectionFromDB(t *testing.T) {
	section := RadiusAgentJSON
	cfgJSONStr := `{
		"radius_agent": {
			"enabled": true,											
			"listen_net": "udp",										
			"listen_auth": "127.0.0.1:1812",							
			"listen_acct": "127.0.0.1:1813",							
			"client_secrets": {											
				"*default": "CGRateS.org"
			},
			"client_dictionaries": {									
				"*default": "/usr/share/cgrates/radius/dict/",			
			},
			"sessions_conns": ["*birpc"],
			"request_processors": [],
		},		
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &RadiusAgentJsonCfg{
		Enabled:     utils.BoolPointer(true),
		Listen_net:  utils.StringPointer("udp"),
		Listen_auth: utils.StringPointer("127.0.0.1:1812"),
		Listen_acct: utils.StringPointer("127.0.0.1:1813"),
		Client_secrets: map[string]string{
			"*default": "CGRateS.org",
		},
		Client_dictionaries: map[string]string{
			"*default": "/usr/share/cgrates/radius/dict/",
		},
		Sessions_conns:     &[]string{"*birpc"},
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDNSAgentSectionFromDB(t *testing.T) {
	section := DNSAgentJSON
	cfgJSONStr := `{
		"dns_agent": {
			"enabled": true,											// enables the DNS agent: <true|false>
			"listen": "127.0.0.1:2053",									// address where to listen for DNS requests <x.y.z.y:1234>
			"listen_net": "udp",										// network to listen on <udp|tcp|tcp-tls>
			"sessions_conns": ["*birpc_internal"],
			"timezone": "UTC",												// timezone of the events if not specified  <UTC|Local|$IANA_TZ_DB>
			"request_processors": [										// request processors to be applied to DNS messages
			],
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &DNSAgentJsonCfg{
		Enabled:            utils.BoolPointer(true),
		Listen:             utils.StringPointer("127.0.0.1:2053"),
		Listen_net:         utils.StringPointer("udp"),
		Sessions_conns:     &[]string{"*birpc_internal"},
		Timezone:           utils.StringPointer("UTC"),
		Request_processors: &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAttributeSSectionFromDB(t *testing.T) {
	section := AttributeSJSON
	cfgJSONStr := `{
		"attributes": {								
			"enabled": false,						
			"stats_conns": ["*birpc"],						
			"resources_conns": ["*birpc"],					
			"admins_conns": ["*birpc"],						
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": false,					
			"process_runs": 1,						
			"any_context": true,					
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(false),
		Stats_conns:           &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Admins_conns:          &[]string{"*birpc"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(false),
		Process_runs:          utils.IntPointer(1),
		Any_context:           utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareChargerSSectionFromDB(t *testing.T) {
	section := ChargerSJSON
	cfgJSONStr := `{
		"chargers": {								
			"enabled": true,						
			"attributes_conns": ["*birpc"],					
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": true,					
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Attributes_conns:      &[]string{"*birpc"},
		Nested_fields:         utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareResourceSSectionFromDB(t *testing.T) {
	section := ResourceSJSON
	cfgJSONStr := `{
		"resources": {								
			"enabled": true,						
			"store_interval": "1s",					
			"thresholds_conns": ["*birpc"],					
			"indexed_selects": true,				
			"string_indexed_fields": [],			 
			"prefix_indexed_fields": [],			 
			"suffix_indexed_fields": [],			 
			"nested_fields": false,					
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareStatSSectionFromDB(t *testing.T) {
	section := StatSJSON
	cfgJSONStr := `{
		"stats": {									
			"enabled": true,						
			"store_interval": "1s",					
			"store_uncompressed_limit": 2,			
			"thresholds_conns": ["*birpc"],					
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": false,					
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(true),
		Thresholds_conns:         &[]string{"*birpc"},
		Store_uncompressed_limit: utils.IntPointer(2),
		Store_interval:           utils.StringPointer("1s"),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index3"},
		Nested_fields:            utils.BoolPointer(false),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareThresholdSSectionFromDB(t *testing.T) {
	section := ThresholdSJSON
	cfgJSONStr := `{
		"thresholds": {								
			"enabled": true,						
			"store_interval": "1s",					
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": true,					
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("1s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(true),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRouteSSectionFromDB(t *testing.T) {
	section := RouteSJSON
	cfgJSONStr := `{
		"routes": {									
			"enabled": true,						
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": false,					
			"attributes_conns": ["*birpc"],					
			"resources_conns": ["*birpc"],					
			"stats_conns": ["*birpc"],						
			"default_ratio":2						
		},	
	}`

	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(false),
		Attributes_conns:      &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Stats_conns:           &[]string{"*birpc"},
		Default_ratio:         utils.IntPointer(2),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSureTaxSectionFromDB(t *testing.T) {
	section := SureTaxJSON
	cfgJSONStr := `{
		"suretax": {
			"url": "sure_tax_url",								
			"client_number": "client_number",					
			"validation_key": "validation_key",					
			"business_unit": "business_unit",					
			"timezone": "Local",					
			"include_local_cost": false,			
			"return_file_code": "0",				
			"response_group": "03",					
			"response_type": "D4",					
			"regulatory_code": "03",				
			"client_tracking": "~*req.CGRID",		
			"customer_number": "~*req.Subject",		
			"orig_number":  "~*req.Subject", 					
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &SureTaxJsonCfg{
		Url:                utils.StringPointer("sure_tax_url"),
		Client_number:      utils.StringPointer("client_number"),
		Validation_key:     utils.StringPointer("validation_key"),
		Business_unit:      utils.StringPointer("business_unit"),
		Timezone:           utils.StringPointer("Local"),
		Include_local_cost: utils.BoolPointer(false),
		Return_file_code:   utils.StringPointer("0"),
		Response_group:     utils.StringPointer("03"),
		Response_type:      utils.StringPointer("D4"),
		Regulatory_code:    utils.StringPointer("03"),
		Client_tracking:    utils.StringPointer("~*req.CGRID"),
		Customer_number:    utils.StringPointer("~*req.Subject"),
		Orig_number:        utils.StringPointer("~*req.Subject"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareDispatcherSSectionFromDB(t *testing.T) {
	section := DispatcherSJSON
	cfgJSONStr := `{
		"dispatchers":{								
			"enabled": true,						
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": false,					
			"attributes_conns": [],					
			"any_subsystem": false,					
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index3"},
		Nested_fields:         utils.BoolPointer(false),
		Attributes_conns:      &[]string{},
		Any_subsystem:         utils.BoolPointer(false),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRegistrarCSectionFromDB(t *testing.T) {
	section := RegistrarCJSON
	cfgJSONStr := `{
		"registrarc":{
			"rpc":{
				"enabled": false,
				"registrars_conns": [],
				"hosts": {},  
				"refresh_interval": "5m",
			},
			"dispatcher":{
				"enabled": false,
				"registrars_conns": [],
				"hosts": {},  
				"refresh_interval": "5m",
			},
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &RegistrarCJsonCfgs{
		RPC: &RegistrarCJsonCfg{
			Enabled:          utils.BoolPointer(false),
			Registrars_conns: &[]string{},
			Hosts:            map[string][]*RemoteHostJson{},
			Refresh_interval: utils.StringPointer("5m"),
		},
		Dispatcher: &RegistrarCJsonCfg{
			Enabled:          utils.BoolPointer(false),
			Registrars_conns: &[]string{},
			Hosts:            map[string][]*RemoteHostJson{},
			Refresh_interval: utils.StringPointer("5m"),
		},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareLoaderSectionFromDB(t *testing.T) {
	section := LoaderJSON
	cfgJSONStr := `{
		"loader": {											
			"tpid": "tpid",										
			"data_path": "./",								
			"disable_reverse": false,						
			"field_separator": ",",							
			"caches_conns":["*localhost"],
			"actions_conns": ["*localhost"],				
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &LoaderCfgJson{
		Tpid:            utils.StringPointer("tpid"),
		Data_path:       utils.StringPointer("./"),
		Disable_reverse: utils.BoolPointer(false),
		Field_separator: utils.StringPointer(","),
		Caches_conns:    &[]string{"*localhost"},
		Actions_conns:   &[]string{"*localhost"},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareMigratorSectionFromDB(t *testing.T) {
	section := MigratorJSON
	cfgJSONStr := `{
		"migrator": {
			"out_datadb_type": "redis",
			"out_datadb_host": "127.0.0.1",
			"out_datadb_port": "6379",
			"out_datadb_name": "10",
			"out_datadb_user": "cgrates",
			"out_datadb_password": "",
			"out_datadb_encoding" : "msgpack",		
			"out_stordb_type": "mysql",
			"out_stordb_host": "127.0.0.1",
			"out_stordb_port": "3306",
			"out_stordb_name": "cgrates",
			"out_stordb_user": "cgrates",
			"out_stordb_password": "",
			"out_datadb_opts":{},
			"out_stordb_opts":{},
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &MigratorCfgJson{
		Out_dataDB_type:     utils.StringPointer("redis"),
		Out_dataDB_host:     utils.StringPointer("127.0.0.1"),
		Out_dataDB_port:     utils.StringPointer("6379"),
		Out_dataDB_name:     utils.StringPointer("10"),
		Out_dataDB_user:     utils.StringPointer("cgrates"),
		Out_dataDB_password: utils.StringPointer(""),
		Out_dataDB_encoding: utils.StringPointer("msgpack"),
		Out_storDB_type:     utils.StringPointer("mysql"),
		Out_storDB_host:     utils.StringPointer("127.0.0.1"),
		Out_storDB_port:     utils.StringPointer("3306"),
		Out_storDB_name:     utils.StringPointer("cgrates"),
		Out_storDB_user:     utils.StringPointer("cgrates"),
		Out_storDB_password: utils.StringPointer(""),
		Out_dataDB_opts:     map[string]interface{}{},
		Out_storDB_opts:     map[string]interface{}{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareTlsSectionFromDB(t *testing.T) {
	section := TlsJSON
	cfgJSONStr := `{
		"tls": {
			"server_certificate" : "server_certificate",			
			"server_key":"server_key",					
			"client_certificate" : "client_certificate",			
			"client_key":"client_key",					
			"ca_certificate":"ca_certificate",				
			"server_policy":4,					
			"server_name":"server_name",
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &TlsJsonCfg{
		Server_certificate: utils.StringPointer("server_certificate"),
		Server_key:         utils.StringPointer("server_key"),
		Server_policy:      utils.IntPointer(4),
		Server_name:        utils.StringPointer("server_name"),
		Client_certificate: utils.StringPointer("client_certificate"),
		Client_key:         utils.StringPointer("client_key"),
		Ca_certificate:     utils.StringPointer("ca_certificate"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAnalyzerSSectionFromDB(t *testing.T) {
	section := AnalyzerSJSON
	cfgJSONStr := `{
		"analyzers":{									
			"enabled": true,							
			 "db_path": "/var/spool/cgrates/analyzers",	
			"index_type": "*scorch",					
			"ttl": "24h",								
			"cleanup_interval": "1h",					
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &AnalyzerSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Db_path:          utils.StringPointer("/var/spool/cgrates/analyzers"),
		Index_type:       utils.StringPointer("*scorch"),
		Ttl:              utils.StringPointer("24h"),
		Cleanup_interval: utils.StringPointer("1h"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAdminSSectionFromDB(t *testing.T) {
	section := AdminSJSON
	cfgJSONStr := `{
		"admins": {
			"enabled": true,
			"caches_conns":["*internal"],
			"actions_conns": ["*birpc"],					
			"attributes_conns": ["*birpc"],					
			"ees_conns": ["*birpc"],						
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &AdminSJsonCfg{
		Enabled:          utils.BoolPointer(true),
		Caches_conns:     &[]string{"*internal"},
		Actions_conns:    &[]string{"*birpc"},
		Attributes_conns: &[]string{"*birpc"},
		Ees_conns:        &[]string{"*birpc"},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareRateSSectionFromDB(t *testing.T) {
	section := RateSJSON
	cfgJSONStr := `{
		"rates": {
			"enabled": true,
			"indexed_selects": true,				// enable profile matching exclusively on indexes
			"string_indexed_fields": ["*req.index1"],			// query indexes based on these fields for faster processing
			"prefix_indexed_fields": ["*req.index2"],			// query indexes based on these fields for faster processing
			"suffix_indexed_fields": ["*req.index3"],			// query indexes based on these fields for faster processing
			"nested_fields": true,					// determines which field is checked when matching indexed filters(true: all; false: only the one on the first level)
			"rate_indexed_selects": false,			// enable profile matching exclusively on indexes
			"rate_string_indexed_fields": ["*req.index1"],		// query indexes based on these fields for faster processing
			"rate_prefix_indexed_fields": ["*req.index2"],		// query indexes based on these fields for faster processing
			"rate_suffix_indexed_fields": ["*req.index3"],		// query indexes based on these fields for faster processing
			"rate_nested_fields": false,			// determines which field is checked when matching indexed filters(true: all; false: only the one on the first level)
			"verbosity": 1000,                      // number of increment iterations allowed
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Indexed_selects:            utils.BoolPointer(true),
		String_indexed_fields:      &[]string{"*req.index1"},
		Prefix_indexed_fields:      &[]string{"*req.index2"},
		Suffix_indexed_fields:      &[]string{"*req.index3"},
		Nested_fields:              utils.BoolPointer(true),
		Rate_indexed_selects:       utils.BoolPointer(false),
		Rate_string_indexed_fields: &[]string{"*req.index1"},
		Rate_prefix_indexed_fields: &[]string{"*req.index2"},
		Rate_suffix_indexed_fields: &[]string{"*req.index3"},
		Rate_nested_fields:         utils.BoolPointer(false),
		Verbosity:                  utils.IntPointer(1000),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareSIPAgentSectionFromDB(t *testing.T) {
	section := SIPAgentJSON
	cfgJSONStr := `{
		"sip_agent": {							// SIP Agents, only used for redirections
			"enabled": false,					// enables the SIP agent: <true|false>
			"listen": "127.0.0.1:5060",			// address where to listen for SIP requests <x.y.z.y:1234>
			"listen_net": "tcp",				// network to listen on <udp|tcp|tcp-tls>
			"sessions_conns": ["*internal"],
			"timezone": "UTC",						// timezone of the events if not specified  <UTC|Local|$IANA_TZ_DB>
			"retransmission_timer": "1s",		// the duration to wait to receive an ACK before resending the reply
			"request_processors": [				// request processors to be applied to SIP messages
			],
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &SIPAgentJsonCfg{
		Enabled:              utils.BoolPointer(false),
		Listen:               utils.StringPointer("127.0.0.1:5060"),
		Listen_net:           utils.StringPointer("tcp"),
		Sessions_conns:       &[]string{"*internal"},
		Timezone:             utils.StringPointer("UTC"),
		Retransmission_timer: utils.StringPointer("1s"),
		Request_processors:   &[]*ReqProcessorJsnCfg{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareConfigSSectionFromDB(t *testing.T) {
	section := ConfigSJSON
	cfgJSONStr := `{
		"configs": {
			"enabled": true,
			"url": "/configs/",										
			"root_dir": "/var/spool/cgrates/configs",				
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ConfigSCfgJson{
		Enabled:  utils.BoolPointer(true),
		Url:      utils.StringPointer("/configs/"),
		Root_dir: utils.StringPointer("/var/spool/cgrates/configs"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAPIBanSectionFromDB(t *testing.T) {
	section := APIBanJSON
	cfgJSONStr := `{
		"apiban": {
			"enabled": false,
			"keys": ["key1", "key2"],
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &APIBanJsonCfg{
		Enabled: utils.BoolPointer(false),
		Keys:    &[]string{"key1", "key2"},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareActionSSectionFromDB(t *testing.T) {
	section := ActionSJSON
	cfgJSONStr := `{
		"actions": {								
			"enabled": true,						
			"cdrs_conns": ["*birpc"],						
			"ees_conns": ["*birpc"],						
			"thresholds_conns": ["*birpc"],					
			"stats_conns": ["*birpc"],						
			"accounts_conns": ["*birpc"],					
			"tenants":["cgrates"],							
			"indexed_selects": true,				
			"string_indexed_fields": ["*req.index1"],			
			"prefix_indexed_fields": ["*req.index2"],			
			"suffix_indexed_fields": ["*req.index3"],			
			"nested_fields": false,					
			"dynaprepaid_actionprofile": [],	
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &ActionSJsonCfg{
		Enabled:                   utils.BoolPointer(true),
		Cdrs_conns:                &[]string{"*birpc"},
		Ees_conns:                 &[]string{"*birpc"},
		Thresholds_conns:          &[]string{"*birpc"},
		Stats_conns:               &[]string{"*birpc"},
		Accounts_conns:            &[]string{"*birpc"},
		Tenants:                   &[]string{"cgrates"},
		Indexed_selects:           utils.BoolPointer(true),
		String_indexed_fields:     &[]string{"*req.index1"},
		Prefix_indexed_fields:     &[]string{"*req.index2"},
		Suffix_indexed_fields:     &[]string{"*req.index3"},
		Nested_fields:             utils.BoolPointer(false),
		Dynaprepaid_actionprofile: &[]string{},
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareAccountSSectionFromDB(t *testing.T) {
	section := AccountSJSON
	cfgJSONStr := `{
		"accounts": {								
			"enabled": true,						
			"indexed_selects": true,				
			"attributes_conns": ["*localhost"],					
			"rates_conns": ["*localhost"],						
			"thresholds_conns": ["*localhost"],					
			"string_indexed_fields": [],			
			"prefix_indexed_fields": [],			
			"suffix_indexed_fields": [],			
			"nested_fields": false,					
			"max_iterations": 1000,                
			"max_usage": "72h",                     
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &AccountSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{"*localhost"},
		Rates_conns:           &[]string{"*localhost"},
		Thresholds_conns:      &[]string{"*localhost"},
		String_indexed_fields: &[]string{},
		Prefix_indexed_fields: &[]string{},
		Suffix_indexed_fields: &[]string{},
		Nested_fields:         utils.BoolPointer(false),
		Max_iterations:        utils.IntPointer(1000),
		Max_usage:             utils.StringPointer("72h"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}

func TestPrepareCoreSSectionFromDB(t *testing.T) {
	section := CoreSJSON
	cfgJSONStr := `{
		"cores": {
			"caps": 3,							 
			"caps_strategy": "*busy",				
			"caps_stats_interval": "0",			
			"shutdown_timeout": "1s"			
		},	
	}`
	cgrCfgJSON, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr))
	if err != nil {
		t.Error(err)
	}

	expected := &CoreSJsonCfg{
		Caps:                utils.IntPointer(3),
		Caps_strategy:       utils.StringPointer("*busy"),
		Caps_stats_interval: utils.StringPointer("0"),
		Shutdown_timeout:    utils.StringPointer("1s"),
	}

	if cfgSec, err := prepareSectionFromDB(section, cgrCfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cfgSec, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(cfgSec))
	}
}
