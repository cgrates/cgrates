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
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err != nil {
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
	eMap := map[string]interface{}{
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
	eMap := map[string]interface{}{
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
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
		t.Errorf("Expected error received nil")
	}
	jsonCfg = &AnalyzerSJsonCfg{
		Ttl: utils.StringPointer("24ha"),
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.analyzerSCfg.loadFromJSONCfg(jsonCfg); err == nil {
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(cS), utils.ToJSON(rcv))
	}
	if rcv.DBPath = ""; cS.DBPath != "/var/spool/cgrates/analyzers" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
