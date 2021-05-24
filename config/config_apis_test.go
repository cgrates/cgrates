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
