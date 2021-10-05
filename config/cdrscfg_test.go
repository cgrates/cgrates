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
		Attributes_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Thresholds_conns:     &[]string{utils.MetaInternal, "*conn1"},
		Stats_conns:          &[]string{utils.MetaInternal, "*conn1"},
		Online_cdr_exports:   &[]string{"randomVal"},
		Actions_conns:        &[]string{utils.MetaInternal, "*conn1"},
		Ees_conns:            &[]string{utils.MetaInternal, "*conn1"},
		Rates_conns:          &[]string{utils.MetaInternal, "*conn1"},
		Accounts_conns:       &[]string{utils.MetaInternal, "*conn1"},
	}
	expected := &CdrsCfg{
		Enabled:          true,
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		AttributeSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		ThresholdSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		OnlineCDRExports: []string{"randomVal"},
		ActionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		EEsConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
		RateSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS), "*conn1"},
		AccountSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"},
		ExtraFields:      RSRParsers{},
		Opts: &CdrsOpts{
			Accounts: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Attributes: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Chargers: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Export: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Rates: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Stats: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
			Thresholds: []*utils.DynamicBoolOpt{
				{
					FilterIDs: []string{utils.MetaDefault},
					Value:     false,
				},
			},
		},
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
		"actions_conns": ["*internal:*actions","*conn1"],		
        "ees_conns": ["*internal:*ees","*conn1"],
        "rates_conns": ["*internal:*rates","*conn1"],
        "accounts_conns": ["*internal:*accounts","*conn1"],
	},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:          true,
		utils.ExtraFieldsCfg:      []string{"~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"},
		utils.StoreCdrsCfg:        true,
		utils.SessionCostRetires:  5,
		utils.ChargerSConnsCfg:    []string{utils.MetaInternal, "*conn1"},
		utils.AttributeSConnsCfg:  []string{utils.MetaInternal, "*conn1"},
		utils.ThresholdSConnsCfg:  []string{utils.MetaInternal, "*conn1"},
		utils.StatSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.OnlineCDRExportsCfg: []string{"http_localhost", "amqp_localhost", "http_test_file"},
		utils.ActionSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.EEsConnsCfg:         []string{utils.MetaInternal, "*conn1"},
		utils.RateSConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.AccountSConnsCfg:    []string{utils.MetaInternal, "*conn1"},
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAccounts: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaAttributes: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaChargers: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaEEs: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaRateS: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaStats: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaThresholds: map[string]bool{
				utils.MetaDefault: false,
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
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
		utils.AttributeSConnsCfg:  []string{"*internal"},
		utils.ThresholdSConnsCfg:  []string{},
		utils.StatSConnsCfg:       []string{},
		utils.OnlineCDRExportsCfg: []string(nil),
		utils.ActionSConnsCfg:     []string{},
		utils.EEsConnsCfg:         []string{"conn1"},
		utils.RateSConnsCfg:       []string{},
		utils.AccountSConnsCfg:    []string{},
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAccounts: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaAttributes: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaChargers: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaEEs: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaRateS: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaStats: map[string]bool{
				utils.MetaDefault: false,
			},
			utils.MetaThresholds: map[string]bool{
				utils.MetaDefault: false,
			},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v", eMap, rcv)
	}
}

func TestCdrsCfgClone(t *testing.T) {
	ban := &CdrsCfg{
		Enabled:          true,
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"},
		AttributeSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		ThresholdSConns:  []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StatSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ActionSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"},
		EEsConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
		OnlineCDRExports: []string{"randomVal"},
		ExtraFields:      RSRParsers{},
		Opts:             &CdrsOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ChargerSConns[1] = ""; ban.ChargerSConns[1] != "*conn1" {
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
	if rcv.ActionSConns[1] = ""; ban.ActionSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.EEsConns[1] = ""; ban.EEsConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if rcv.OnlineCDRExports[0] = ""; ban.OnlineCDRExports[0] != "randomVal" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffCdrsJsonCfg(t *testing.T) {
	var d *CdrsJsonCfg

	v1 := &CdrsCfg{
		Enabled: false,
		ExtraFields: RSRParsers{
			{
				Rules: "Rule1",
			},
		},
		StoreCdrs:        false,
		SMCostRetries:    2,
		ChargerSConns:    []string{"*localhost"},
		AttributeSConns:  []string{"*localhost"},
		ThresholdSConns:  []string{"*localhost"},
		StatSConns:       []string{"*localhost"},
		OnlineCDRExports: []string{},
		ActionSConns:     []string{"*localhost"},
		EEsConns:         []string{"*localhost"},
		Opts:             &CdrsOpts{},
	}

	v2 := &CdrsCfg{
		Enabled: true,
		ExtraFields: RSRParsers{
			{
				Rules: "Rule2",
			},
		},
		StoreCdrs:        true,
		SMCostRetries:    1,
		ChargerSConns:    []string{"*birpc"},
		AttributeSConns:  []string{"*birpc"},
		ThresholdSConns:  []string{"*birpc"},
		StatSConns:       []string{"*birpc"},
		OnlineCDRExports: []string{"val1"},
		ActionSConns:     []string{"*birpc"},
		EEsConns:         []string{"*birpc"},
		Opts:             &CdrsOpts{},
	}

	expected := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Extra_fields:         &[]string{"Rule2"},
		Store_cdrs:           utils.BoolPointer(true),
		Session_cost_retries: utils.IntPointer(1),
		Chargers_conns:       &[]string{"*birpc"},
		Attributes_conns:     &[]string{"*birpc"},
		Thresholds_conns:     &[]string{"*birpc"},
		Stats_conns:          &[]string{"*birpc"},
		Online_cdr_exports:   &[]string{"val1"},
		Actions_conns:        &[]string{"*birpc"},
		Ees_conns:            &[]string{"*birpc"},
		Opts:                 &CdrsOptsJson{},
	}

	rcv := diffCdrsJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &CdrsJsonCfg{
		Opts: &CdrsOptsJson{},
	}

	rcv = diffCdrsJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}
