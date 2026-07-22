// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestAnalyzerSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AnalyzerSJsonCfg{
		Enabled: utils.BoolPointer(false),
	}
	expected := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		TTL:             24 * time.Hour,
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(jsnCfg.analyzerSCfg, expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, jsnCfg.analyzerSCfg)
	}
}

func TestAnalyzerSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{},
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:         false,
		utils.CleanupIntervalCfg: "1h0m0s",
		utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
		utils.IndexTypeCfg:       utils.MetaScorch,
		utils.TTLCfg:             "24h0m0s",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , received: %+v", eMap, rcv)
	}
}

func TestAnalyzerSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"analyzers":{
            "enabled": true,  
        },
    }
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.CleanupIntervalCfg: "1h0m0s",
		utils.DBPathCfg:          "/var/spool/cgrates/analyzers",
		utils.IndexTypeCfg:       utils.MetaScorch,
		utils.TTLCfg:             "24h0m0s",
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.analyzerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected: %+v , received: %+v", eMap, rcv)
	}
}

func TestAnalyzerSCfgloadFromJsonCfgErr(t *testing.T) {
	jsonCfg := &AnalyzerSJsonCfg{
		Cleanup_interval: utils.StringPointer("24ha"),
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
		t.Errorf("Expected error received nil")
	}
	jsonCfg = &AnalyzerSJsonCfg{
		Ttl: utils.StringPointer("24ha"),
	}
	jsnCfg = NewDefaultCGRConfig()
	if err := jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
		t.Errorf("Expected error received nil")
	}
}

func TestAnalyzerSCfgClone(t *testing.T) {
	cS := &AnalyzerSCfg{
		Enabled:         false,
		CleanupInterval: time.Hour,
		DBPath:          "/var/spool/cgrates/analyzers",
		IndexType:       utils.MetaScorch,
		TTL:             24 * time.Hour,
	}
	rcv := cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.DBPath = ""; cS.DBPath != "/var/spool/cgrates/analyzers" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	cS = nil
	rcv = cS.Clone()
	if !reflect.DeepEqual(cS, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}

}
