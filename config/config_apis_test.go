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
