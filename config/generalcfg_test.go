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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestGeneralCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Logger:               utils.StringPointer(utils.MetaSysLog),
		Log_level:            utils.IntPointer(6),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Default_request_type: utils.StringPointer(utils.MetaRated),
		Default_category:     utils.StringPointer(utils.Call),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
	}

	expected := &GeneralCfg{
		NodeID:           "randomID",
		Logger:           utils.MetaSysLog,
		LogLevel:         6,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		PosterAttempts:   3,
		DefaultReqType:   utils.MetaRated,
		DefaultCategory:  utils.Call,
		DefaultTenant:    "cgrates.org",
		DefaultTimezone:  "Local",
		ConnectAttempts:  3,
		Reconnects:       -1,
		ConnectTimeout:   time.Second,
		ReplyTimeout:     2 * time.Second,
		DigestSeparator:  ",",
		DigestEqual:      ":",
		MaxParallelConns: 100,
		RSRSep:           ";",
		DefaultCaching:   utils.MetaReload,
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.generalCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.generalCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.generalCfg))
	}

	cfgJSON.Max_reconnect_interval = utils.StringPointer("test1")

	if err := jsnCfg.generalCfg.loadFromJSONCfg(cfgJSON); err == nil {
		t.Error(err)
	}
}

func TestGeneralParseDurationCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &GeneralJsonCfg{
		Connect_timeout: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.generalCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %v", expected, err)
	}

	cfgJSON1 := &GeneralJsonCfg{
		Reply_timeout: utils.StringPointer("1ss"),
	}
	jsonCfg = NewDefaultCGRConfig()
	if err := jsonCfg.generalCfg.loadFromJSONCfg(cfgJSON1); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %v", expected, err)
	}

	cfgJSON2 := &GeneralJsonCfg{
		Locking_timeout: utils.StringPointer("1ss"),
	}
	jsonCfg = NewDefaultCGRConfig()
	if err := jsonCfg.generalCfg.loadFromJSONCfg(cfgJSON2); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %v", expected, err)
	}

}

func TestGeneralCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"general": {
			"node_id": "cgrates",											
			"logger":"*syslog",										
			"log_level": 6,											
			"rounding_decimals": 5,									
			"dbdata_encoding": "*msgpack",							
			"tpexport_dir": "/var/spool/cgrates/tpe",				
			"poster_attempts": 3,									
			"default_request_type": "*rated",						
			"default_category": "call",								
			"default_tenant": "cgrates.org",						
			"default_timezone": "Local",							
			"default_caching":"*reload",							
			"connect_attempts": 5,									
			"reconnects": -1,										
			"connect_timeout": "1s",								
			"reply_timeout": "2s",									
			"locking_timeout": "1s",									
			"digest_separator": ",",								
			"digest_equal": ":",									
			"rsr_separator": ";",									
			"max_parallel_conns": 100,
			"max_reconnect_interval":"1s"
		},
	}`

	eMap := map[string]any{
		utils.NodeIDCfg:               "cgrates",
		utils.LoggerCfg:               "*syslog",
		utils.LogLevelCfg:             6,
		utils.RoundingDecimalsCfg:     5,
		utils.DBDataEncodingCfg:       "*msgpack",
		utils.TpExportPathCfg:         "/var/spool/cgrates/tpe",
		utils.PosterAttemptsCfg:       3,
		utils.DefaultReqTypeCfg:       "*rated",
		utils.DefaultCategoryCfg:      "call",
		utils.DefaultTenantCfg:        "cgrates.org",
		utils.DefaultTimezoneCfg:      "Local",
		utils.DefaultCachingCfg:       "*reload",
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      5,
		utils.ReconnectsCfg:           -1,
		utils.MaxReconnectIntervalCfg: "1s",
		utils.ConnectTimeoutCfg:       "1s",
		utils.ReplyTimeoutCfg:         "2s",
		utils.LockingTimeoutCfg:       "1s",
		utils.DigestSeparatorCfg:      ",",
		utils.DigestEqualCfg:          ":",
		utils.RSRSepCfg:               ";",
		utils.MaxParallelConnsCfg:     100,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.generalCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestGeneralCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
      "general": {
            "node_id": "ENGINE1",
            "locking_timeout": "0",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
            "max_call_duration": "0"
        }
}`
	eMap := map[string]any{
		utils.NodeIDCfg:               "ENGINE1",
		utils.LoggerCfg:               "*syslog",
		utils.LogLevelCfg:             6,
		utils.RoundingDecimalsCfg:     5,
		utils.DBDataEncodingCfg:       "*msgpack",
		utils.TpExportPathCfg:         "/var/spool/cgrates/tpe",
		utils.PosterAttemptsCfg:       3,
		utils.DefaultReqTypeCfg:       "*rated",
		utils.DefaultCategoryCfg:      "call",
		utils.DefaultTenantCfg:        "cgrates.org",
		utils.DefaultTimezoneCfg:      "Local",
		utils.DefaultCachingCfg:       "*reload",
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      5,
		utils.ReconnectsCfg:           -1,
		utils.MaxReconnectIntervalCfg: "0",
		utils.ConnectTimeoutCfg:       "0",
		utils.ReplyTimeoutCfg:         "0",
		utils.LockingTimeoutCfg:       "0",
		utils.DigestSeparatorCfg:      ",",
		utils.DigestEqualCfg:          ":",
		utils.RSRSepCfg:               ";",
		utils.MaxParallelConnsCfg:     100,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.generalCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestGeneralCfgClone(t *testing.T) {
	ban := &GeneralCfg{
		NodeID:           "randomID",
		Logger:           utils.MetaSysLog,
		LogLevel:         6,
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
		PosterAttempts:   3,
		DefaultReqType:   utils.MetaRated,
		DefaultCategory:  utils.Call,
		DefaultTenant:    "cgrates.org",
		DefaultTimezone:  "Local",
		ConnectAttempts:  3,
		Reconnects:       -1,
		ConnectTimeout:   time.Second,
		ReplyTimeout:     2 * time.Second,
		DigestSeparator:  ",",
		DigestEqual:      ":",
		MaxParallelConns: 100,
		RSRSep:           ";",
		DefaultCaching:   utils.MetaReload,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.NodeID = ""; ban.NodeID != "randomID" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestCachingDelay(t *testing.T) {

	gencfg := &GeneralCfg{
		CachingDelay: 5 * time.Second,
	}

	expectedMap := map[string]any{
		utils.CachingDlayCfg: "5s",
	}
	resultMap := gencfg.AsMapInterface()
	if resultMap[utils.CachingDlayCfg] != expectedMap[utils.CachingDlayCfg] {
		t.Errorf("Non-zero CachingDelay test failed. Expected %v, but got %v", expectedMap[utils.CachingDlayCfg], resultMap[utils.CachingDlayCfg])
	}
	gencfg = &GeneralCfg{
		CachingDelay: 0,
	}
	expectedMap = map[string]any{
		utils.CachingDlayCfg: "0",
	}
	resultMap = gencfg.AsMapInterface()
	if resultMap[utils.CachingDlayCfg] != expectedMap[utils.CachingDlayCfg] {
		t.Errorf("Zero CachingDelay test failed. Expected %v, but got %v", expectedMap[utils.CachingDlayCfg], resultMap[utils.CachingDlayCfg])
	}
}
