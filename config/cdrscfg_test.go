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

func TestCdrsCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Store_cdrs:           utils.BoolPointer(true),
		Session_cost_retries: utils.IntPointer(1),
		Chargers_conns:       &[]string{utils.MetaInternal, "*conn1"},
		Rals_conns:           &[]string{utils.MetaInternal, "*conn1"},
		Attributes_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Thresholds_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:          &[]string{utils.MetaInternal, "*conn1"},
		Online_cdr_exports:   &[]string{"randomVal"},
		Scheduler_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Ees_conns:            &[]string{utils.MetaInternal, "*conn1"},
	}
	expected := &CdrsCfg{
		Enabled:          true,
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		RaterConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder), "*conn1"},
		AttributeSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		ThresholdSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		OnlineCDRExports: []string{"randomVal"},
		SchedulerConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler), "*conn1"},
		EEsConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
		ExtraFields:      RSRParsers{},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.cdrsCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.cdrsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.cdrsCfg))
	}
}

func TestExtraFieldsinloadFromJsonCfg(t *testing.T) {
	cfgJSON := &CdrsJsonCfg{
		Extra_fields: &[]string{utils.EmptyString},
	}
	expectedErrMessage := "emtpy RSRParser in rule: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.cdrsCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expectedErrMessage {
		t.Errorf("Expected %+v, received %+v", expectedErrMessage, err)
	}
}

func TestCdrsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {
		"enabled": true,						
		"extra_fields": ["~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"],
		"store_cdrs": true,						
		"session_cost_retries": 5,				
		"chargers_conns":["*internal:*chargers","*conn1"],			
		"rals_conns": ["*internal:*responder","*conn1"],
		"attributes_conns": ["*internal:*attributes","*conn1"],					
		"thresholds_conns": ["*internal:*thresholds","*conn1"],					
		"stats_conns": ["*internal:*stats","*conn1"],						
		"online_cdr_exports":["http_localhost", "amqp_localhost", "http_test_file"],
		"scheduler_conns": ["*internal:*scheduler","*conn1"],		
        "ees_conns": ["*internal:*ees","*conn1"],
	},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.ExtraFieldsCfg:      []string{"~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"},
		utils.StoreCdrsCfg:        true,
		utils.SessionCostRetires:  5,
		utils.ChargerSConnsCfg:    []string{utils.MetaInternal, "*conn1"},
		utils.RALsConnsCfg:        []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg:  []string{utils.MetaInternal, "*conn1"},
		utils.ThresholdSConnsCfg:  []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.OnlineCDRExportsCfg: []string{"http_localhost", "amqp_localhost", "http_test_file"},
		utils.SchedulerConnsCfg:   []string{utils.MetaInternal, "*conn1"},
		utils.EEsConnsCfg:         []string{utils.MetaInternal, "*conn1"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v ", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestCdrsCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
       "cdrs": {
          "enabled":true,
          "chargers_conns": ["conn1", "conn2"],
          "attributes_conns": ["*internal"],
          "ees_conns": ["conn1"],
       },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.ExtraFieldsCfg:      []string{},
		utils.StoreCdrsCfg:        true,
		utils.SessionCostRetires:  5,
		utils.ChargerSConnsCfg:    []string{"conn1", "conn2"},
		utils.RALsConnsCfg:        []string{},
		utils.AttributeSConnsCfg:  []string{"*internal"},
		utils.ThresholdSConnsCfg:  []string{},
		utils.StatSConnsCfg:       []string{},
		utils.OnlineCDRExportsCfg: []string{},
		utils.SchedulerConnsCfg:   []string{},
		utils.EEsConnsCfg:         []string{"conn1"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v", eMap, rcv)
	}
}

func TestCdrsCfgClone(t *testing.T) {
	ban := &CdrsCfg{
		Enabled:          true,
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		RaterConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder), "*conn1"},
		AttributeSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		ThresholdSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		SchedulerConns:   []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler), "*conn1"},
		EEsConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
		OnlineCDRExports: []string{"randomVal"},
		ExtraFields:      RSRParsers{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ChargerSConns[1] = ""; ban.ChargerSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.RaterConns[1] = ""; ban.RaterConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.AttributeSConns[1] = ""; ban.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ThresholdSConns[1] = ""; ban.ThresholdSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.SchedulerConns[1] = ""; ban.SchedulerConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EEsConns[1] = ""; ban.EEsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.OnlineCDRExports[0] = ""; ban.OnlineCDRExports[0] != "randomVal" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
