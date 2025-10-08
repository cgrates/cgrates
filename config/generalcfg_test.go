/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestGeneralCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
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
		Opts:                 &GeneralOptsJson{},
	}

	expected := &GeneralCfg{
		NodeID:           "randomID",
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
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
		DefaultCaching:   utils.MetaReload,
		Opts: &GeneralOpts{
			ExporterIDs: []*DynamicStringSliceOpt{},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.generalCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.generalCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.generalCfg))
	}
	cfgJSON = nil
	if err := jsnCfg.generalCfg.loadFromJSONCfg(cfgJSON); err != nil {
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

	cfgJSON3 := &GeneralJsonCfg{
		Locking_timeout: utils.StringPointer("1ss"),
	}
	jsonCfg = NewDefaultCGRConfig()
	if err := jsonCfg.generalCfg.loadFromJSONCfg(cfgJSON3); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %v", expected, err)
	}

}

func TestGeneralCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"general": {
			"node_id": "cgrates",																				
			"rounding_decimals": 5,									
			"dbdata_encoding": "*msgpack",							
			"tpexport_dir": "/var/spool/cgrates/tpe",											
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
			"max_parallel_conns": 100,								
		},
	}`
	eMap := map[string]any{
		utils.NodeIDCfg:               "cgrates",
		utils.RoundingDecimalsCfg:     5,
		utils.DBDataEncodingCfg:       "*msgpack",
		utils.TpExportPathCfg:         "/var/spool/cgrates/tpe",
		utils.DefaultReqTypeCfg:       "*rated",
		utils.DefaultCategoryCfg:      "call",
		utils.DefaultTenantCfg:        "cgrates.org",
		utils.DefaultTimezoneCfg:      "Local",
		utils.DefaultCachingCfg:       "*reload",
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      5,
		utils.MaxReconnectIntervalCfg: "0",
		utils.ReconnectsCfg:           -1,
		utils.ConnectTimeoutCfg:       "1s",
		utils.ReplyTimeoutCfg:         "2s",
		utils.LockingTimeoutCfg:       "1s",
		utils.DigestSeparatorCfg:      ",",
		utils.DigestEqualCfg:          ":",
		utils.MaxParallelConnsCfg:     100,
		utils.DecimalMaxScaleCfg:      0,
		utils.DecimalMinScaleCfg:      0,
		utils.DecimalPrecisionCfg:     0,
		utils.DecimalRoundingModeCfg:  "*toNearestEven",
		utils.OptsCfg: map[string]any{
			utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
		},
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
			"max_reconnect_interval": "5m",
            "connect_timeout": "0s",
            "reply_timeout": "0s",
            "max_call_duration": "0"
        }
}`
	eMap := map[string]any{
		utils.NodeIDCfg:               "ENGINE1",
		utils.RoundingDecimalsCfg:     5,
		utils.DBDataEncodingCfg:       "*msgpack",
		utils.TpExportPathCfg:         "/var/spool/cgrates/tpe",
		utils.DefaultReqTypeCfg:       "*rated",
		utils.DefaultCategoryCfg:      "call",
		utils.DefaultTenantCfg:        "cgrates.org",
		utils.DefaultTimezoneCfg:      "Local",
		utils.DefaultCachingCfg:       "*reload",
		utils.CachingDlayCfg:          "0",
		utils.ConnectAttemptsCfg:      5,
		utils.ReconnectsCfg:           -1,
		utils.MaxReconnectIntervalCfg: "5m0s",
		utils.ConnectTimeoutCfg:       "0",
		utils.ReplyTimeoutCfg:         "0",
		utils.LockingTimeoutCfg:       "0",
		utils.DigestSeparatorCfg:      ",",
		utils.DigestEqualCfg:          ":",
		utils.MaxParallelConnsCfg:     100,
		utils.DecimalMaxScaleCfg:      0,
		utils.DecimalMinScaleCfg:      0,
		utils.DecimalPrecisionCfg:     0,
		utils.DecimalRoundingModeCfg:  "*toNearestEven",
		utils.OptsCfg: map[string]any{
			utils.MetaExporterIDs: []*DynamicStringSliceOpt{},
		},
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
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
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
		DefaultCaching:   utils.MetaReload,
		Opts: &GeneralOpts{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					Values: []string{"*ees"},
				},
			},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.NodeID = ""; ban.NodeID != "randomID" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffGeneralJsonCfg(t *testing.T) {
	var d *GeneralJsonCfg

	v1 := &GeneralCfg{
		NodeID:           "randomID2",
		RoundingDecimals: 1,
		DBDataEncoding:   "msgpack2",
		TpExportPath:     "/var/spool/cgrates/tpe/test",
		DefaultReqType:   utils.MetaPrepaid,
		DefaultCategory:  utils.ForcedDisconnectCfg,
		DefaultTenant:    "itsyscom.com",
		DefaultTimezone:  "UTC",
		ConnectAttempts:  5,
		Reconnects:       2,
		ConnectTimeout:   5 * time.Second,
		ReplyTimeout:     1 * time.Second,
		DigestSeparator:  "",
		DigestEqual:      "",
		MaxParallelConns: 50,
		DefaultCaching:   utils.MetaClear,
		Opts: &GeneralOpts{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					Values: []string{"*ees"},
				},
			},
		},
		MaxReconnectInterval: time.Duration(5),
		DecimalMaxScale:      5,
		DecimalMinScale:      5,
		DecimalPrecision:     5,
		DecimalRoundingMode:  decimal.ToNearestAway,
	}

	v2 := &GeneralCfg{
		NodeID:           "randomID",
		RoundingDecimals: 5,
		DBDataEncoding:   "msgpack",
		TpExportPath:     "/var/spool/cgrates/tpe",
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
		DefaultCaching:   utils.MetaReload,
		LockingTimeout:   2 * time.Second,
		Opts: &GeneralOpts{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					Values: []string{"*syslog"},
				},
			},
		},
		MaxReconnectInterval: time.Duration(2),
		DecimalMaxScale:      2,
		DecimalMinScale:      2,
		DecimalPrecision:     2,
		DecimalRoundingMode:  decimal.ToNearestEven,
	}

	expected := &GeneralJsonCfg{
		Node_id:              utils.StringPointer("randomID"),
		Rounding_decimals:    utils.IntPointer(5),
		Dbdata_encoding:      utils.StringPointer("msgpack"),
		Tpexport_dir:         utils.StringPointer("/var/spool/cgrates/tpe"),
		Default_request_type: utils.StringPointer(utils.MetaRated),
		Default_category:     utils.StringPointer(utils.Call),
		Default_tenant:       utils.StringPointer("cgrates.org"),
		Default_timezone:     utils.StringPointer("Local"),
		Default_caching:      utils.StringPointer(utils.MetaReload),
		Connect_attempts:     utils.IntPointer(3),
		Reconnects:           utils.IntPointer(-1),
		Connect_timeout:      utils.StringPointer("1s"),
		Reply_timeout:        utils.StringPointer("2s"),
		Locking_timeout:      utils.StringPointer("2s"),
		Digest_separator:     utils.StringPointer(","),
		Digest_equal:         utils.StringPointer(":"),
		Max_parallel_conns:   utils.IntPointer(100),
		Opts: &GeneralOptsJson{
			ExporterIDs: []*DynamicStringSliceOpt{
				{
					Values: []string{"*syslog"},
				},
			},
		},
		Max_reconnect_interval: utils.StringPointer("2ns"),
		Decimal_max_scale:      utils.IntPointer(2),
		Decimal_min_scale:      utils.IntPointer(2),
		Decimal_precision:      utils.IntPointer(2),
		Decimal_rounding_mode:  utils.StringPointer("ToNearestEven"),
	}

	rcv := diffGeneralJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &GeneralJsonCfg{
		Opts: &GeneralOptsJson{},
	}

	rcv = diffGeneralJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestGeneralCfgCloneSection(t *testing.T) {
	gnrCfg := &GeneralCfg{
		NodeID:           "randomID2",
		RoundingDecimals: 1,
		DBDataEncoding:   "msgpack2",
		TpExportPath:     "/var/spool/cgrates/tpe/test",
		DefaultReqType:   utils.MetaPrepaid,
		DefaultCategory:  utils.ForcedDisconnectCfg,
		DefaultTenant:    "itsyscom.com",
		DefaultTimezone:  "UTC",
		ConnectAttempts:  5,
		Reconnects:       2,
		ConnectTimeout:   5 * time.Second,
		ReplyTimeout:     1 * time.Second,
		DigestSeparator:  "",
		DigestEqual:      "",
		MaxParallelConns: 50,
		DefaultCaching:   utils.MetaClear,
		Opts:             &GeneralOpts{},
	}

	exp := &GeneralCfg{
		NodeID:           "randomID2",
		RoundingDecimals: 1,
		DBDataEncoding:   "msgpack2",
		TpExportPath:     "/var/spool/cgrates/tpe/test",
		DefaultReqType:   utils.MetaPrepaid,
		DefaultCategory:  utils.ForcedDisconnectCfg,
		DefaultTenant:    "itsyscom.com",
		DefaultTimezone:  "UTC",
		ConnectAttempts:  5,
		Reconnects:       2,
		ConnectTimeout:   5 * time.Second,
		ReplyTimeout:     1 * time.Second,
		DigestSeparator:  "",
		DigestEqual:      "",
		MaxParallelConns: 50,
		DefaultCaching:   utils.MetaClear,
		Opts:             &GeneralOpts{},
	}

	rcv := gnrCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestGeneralOptsLoadFromJSONCfgNilJson(t *testing.T) {
	generalOpts := &GeneralOpts{}
	var jsnCfg *GeneralOptsJson
	generalOptsClone := &GeneralOpts{}
	generalOpts.loadFromJSONCfg(jsnCfg)
	if !reflect.DeepEqual(generalOptsClone, generalOpts) {
		t.Errorf("Expected GeneralOpts to not change, Was <%+v>,\nNow is <%+v>",
			generalOptsClone, generalOpts)
	}
}
func TestGeneralCfgloadFromJsonCfgMaxReconnInterval(t *testing.T) {
	cfgJSON := &GeneralJsonCfg{Max_reconnect_interval: utils.StringPointer("invalid time")}

	expected := `time: invalid duration "invalid time"`
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.generalCfg.loadFromJSONCfg(cfgJSON); err.Error() != expected {
		t.Errorf("Expected error <%v>, Received error <%v>", expected, err.Error())
	}
}

func TestGeneralOptsCloneNil(t *testing.T) {

	var generalOpts *GeneralOpts
	generalOptsClone := generalOpts.Clone()
	if !reflect.DeepEqual(generalOptsClone, generalOpts) {
		t.Errorf("Expected GeneralOpts to not change, Was <%+v>,\nNow is <%+v>",
			generalOptsClone, generalOpts)
	}
}
