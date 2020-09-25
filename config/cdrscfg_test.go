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
		Chargers_conns:       &[]string{"*internal"},
		Rals_conns:           &[]string{"*internal"},
		Attributes_conns:     &[]string{"*internal"},
		Thresholds_conns:     &[]string{"*internal"},
		Stats_conns:          &[]string{"*conn1", "*conn2"},
		Online_cdr_exports:   &[]string{"randomVal"},
		Scheduler_conns:      &[]string{"*internal"},
		Ees_conns:            &[]string{"*internal"},
	}
	expected := &CdrsCfg{
		Enabled:          true,
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{"*internal:*chargers"},
		RaterConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder)},
		AttributeSConns:  []string{"*internal:*attributes"},
		ThresholdSConns:  []string{"*internal:*thresholds"},
		StatSConns:       []string{"*conn1", "*conn2"},
		OnlineCDRExports: []string{"randomVal"},
		SchedulerConns:   []string{"*internal:*scheduler"},
		EEsConns:         []string{"*internal:*ees"},
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.cdrsCfg.loadFromJsonCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.cdrsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.cdrsCfg))
	}
}

func TestExtraFieldsinAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {
		"enabled": true,
		"extra_fields": ["~effective_caller_id_number:s/(\\d+)/+$1/","~Custom_Val:s/(\\d+)/+$1/"],
		"chargers_conns":["*localhost"],
		"store_cdrs": true,
		"online_cdr_exports": []
	},
	}`
	expectedExtra := []string{`~effective_caller_id_number:s/(\d+)/+$1/`, "~Custom_Val:s/(\\d+)/+$1/"}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv[utils.ExtraFieldsCfg], expectedExtra) {
		t.Errorf("Expected %+v \n, recieved %+v \n", expectedExtra, rcv[utils.ExtraFieldsCfg])
	}
}

func TestCdrsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {
		"enabled": false,						
		"extra_fields": ["~Custom_Val:s/(\\d+)/+$1/"],
		"store_cdrs": true,						
		"session_cost_retries": 5,				
		"chargers_conns":["*localhost"],				
		"rals_conns": ["*internal"],
		"attributes_conns": [],					
		"thresholds_conns": [],					
		"stats_conns": [],						
		"online_cdr_exports":[],
		"scheduler_conns": [],		
        "ees_conns": [],
	},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          false,
		utils.ExtraFieldsCfg:      []string{"~Custom_Val:s/(\\d+)/+$1/"},
		utils.StoreCdrsCfg:        true,
		utils.SessionCostRetires:  5,
		utils.ChargerSConnsCfg:    []string{"*localhost"},
		utils.RALsConnsCfg:        []string{"*internal"},
		utils.AttributeSConnsCfg:  []string{},
		utils.ThresholdSConnsCfg:  []string{},
		utils.StatSConnsCfg:       []string{},
		utils.OnlineCDRExportsCfg: []string{},
		utils.SchedulerConnsCfg:   []string{},
		utils.EEsConnsCfg:         []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v ", eMap, rcv)
	}
}

func TestCdrsCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"cdrs": {
			"enabled": true,						
			"extra_fields": ["PayPalAccount", "LCRProfile", "ResourceID"],
			"store_cdrs": true,						
			"session_cost_retries": 9,				
			"chargers_conns":["*internal"],				
			"rals_conns": ["*internal"],
			"attributes_conns": ["*internal"],					
			"thresholds_conns": ["*internal"],					
			"stats_conns": ["*internal"],						
			"online_cdr_exports":["http_localhost", "amqp_localhost", "http_test_file", "amqp_test_file","aws_test_file","sqs_test_file","kafka_localhost","s3_test_file"],
			"scheduler_conns": ["*internal"],	
            "ees_conns": [],
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.ExtraFieldsCfg:      []string{"PayPalAccount", "LCRProfile", "ResourceID"},
		utils.StoreCdrsCfg:        true,
		utils.SessionCostRetires:  9,
		utils.ChargerSConnsCfg:    []string{"*internal"},
		utils.RALsConnsCfg:        []string{"*internal"},
		utils.AttributeSConnsCfg:  []string{"*internal"},
		utils.ThresholdSConnsCfg:  []string{"*internal"},
		utils.StatSConnsCfg:       []string{"*internal"},
		utils.OnlineCDRExportsCfg: []string{"http_localhost", "amqp_localhost", "http_test_file", "amqp_test_file", "aws_test_file", "sqs_test_file", "kafka_localhost", "s3_test_file"},
		utils.SchedulerConnsCfg:   []string{"*internal"},
		utils.EEsConnsCfg:         []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v", eMap, rcv)
	}
}

func TestCdrsCfgAsMapInterface2(t *testing.T) {
	cfgJsonStr := `{
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJsonStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v", eMap, rcv)
	}
}
