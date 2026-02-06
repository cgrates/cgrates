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

	"github.com/cgrates/cgrates/utils"
)

func TestCdrsCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Session_cost_retries: utils.IntPointer(1),
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAttributes: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaThresholds: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaActions:    {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaEEs:        {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	expected := &CdrsCfg{
		Enabled: true,

		SMCostRetries: 1,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"}}},
			utils.MetaAttributes: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
			utils.MetaThresholds: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaActions:    {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
			utils.MetaEEs:        {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), "*conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"}}},
		},
		OnlineCDRExports: nil,
		ExtraFields:      utils.RSRParsers{},
		Opts: &CdrsOpts{
			Accounts:   []*DynamicBoolOpt{{}},
			Attributes: []*DynamicBoolOpt{{}},
			Chargers:   []*DynamicBoolOpt{{}},
			Export:     []*DynamicBoolOpt{{}},
			Rates:      []*DynamicBoolOpt{{}},
			Stats:      []*DynamicBoolOpt{{}},
			Thresholds: []*DynamicBoolOpt{{}},
			Refund:     []*DynamicBoolOpt{{}},
			Rerate:     []*DynamicBoolOpt{{}},
			Store:      []*DynamicBoolOpt{{value: true}},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.cdrsCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.cdrsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.cdrsCfg))
	}

	jsonCfg = nil
	if err := jsnCfg.cdrsCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestCdrsCfgloadFromJsonCfgOpt(t *testing.T) {
	cdrsOpt := &CdrsOpts{
		Accounts: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Attributes: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Chargers: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Export: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Rates: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Stats: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Thresholds: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
	}

	exp := &CdrsOpts{
		Accounts: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Attributes: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Chargers: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Export: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Rates: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Stats: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
		Thresholds: []*DynamicBoolOpt{
			{
				value: false,
			},
		},
	}

	cdrsOpt.loadFromJSONCfg(nil)
	if !reflect.DeepEqual(exp, cdrsOpt) {
		t.Errorf("Expected %+v, received %+v", exp, cdrsOpt)
	}
}

func TestExtraFieldsinloadFromJsonCfg(t *testing.T) {
	cfgJSON := &CdrsJsonCfg{
		Extra_fields: &[]string{utils.EmptyString},
	}
	expectedErrMessage := "empty RSRParser in rule: <>"
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.cdrsCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expectedErrMessage {
		t.Errorf("expected %q, received %q", expectedErrMessage, err)
	}
}

func TestCdrsCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"cdrs": {
		"enabled": true,
		"extra_fields": ["~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"],
		"session_cost_retries": 5,
		"conns": {
			"*chargers": [{"values": ["*internal:*chargers","*conn1"]}],
			"*attributes": [{"values": ["*internal:*attributes","*conn1"]}],
			"*thresholds": [{"values": ["*internal:*thresholds","*conn1"]}],
			"*stats": [{"values": ["*internal:*stats","*conn1"]}],
			"*actions": [{"values": ["*internal:*actions","*conn1"]}],
			"*ees": [{"values": ["*internal:*ees","*conn1"]}],
			"*rates": [{"values": ["*internal:*rates","*conn1"]}],
			"*accounts": [{"values": ["*internal:*accounts","*conn1"]}]
		},
		"online_cdr_exports":["http_localhost", "amqp_localhost", "http_test_file"],
	},
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.ExtraFieldsCfg:     []string{"~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"},
		utils.SessionCostRetires: 5,
		utils.ConnsCfg: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAttributes: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaThresholds: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaActions:    {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaEEs:        {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.MetaInternal, "*conn1"}}},
		},
		utils.OnlineCDRExportsCfg: []string{"http_localhost", "amqp_localhost", "http_test_file"},
		utils.OptsCfg: map[string]any{
			utils.MetaAccounts:   []*DynamicBoolOpt{{}},
			utils.MetaAttributes: []*DynamicBoolOpt{{}},
			utils.MetaChargers:   []*DynamicBoolOpt{{}},
			utils.MetaEEs:        []*DynamicBoolOpt{{}},
			utils.MetaRates:      []*DynamicBoolOpt{{}},
			utils.MetaStats:      []*DynamicBoolOpt{{}},
			utils.MetaThresholds: []*DynamicBoolOpt{{}},
			utils.MetaRefund:     []*DynamicBoolOpt{{}},
			utils.MetaRerate:     []*DynamicBoolOpt{{}},
			utils.MetaStore:      []*DynamicBoolOpt{{value: true}},
		},
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
          "conns": {
              "*chargers": [{"values": ["conn1", "conn2"]}],
              "*attributes": [{"values": ["*internal"]}],
              "*ees": [{"values": ["conn1"]}]
          },
       },
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.ExtraFieldsCfg:     []string{},
		utils.SessionCostRetires: 5,
		utils.ConnsCfg: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"conn1", "conn2"}}},
			utils.MetaAttributes: {{Values: []string{"*internal"}}},
			utils.MetaEEs:        {{Values: []string{"conn1"}}},
		},
		utils.OnlineCDRExportsCfg: []string(nil),
		utils.OptsCfg: map[string]any{
			utils.MetaAccounts:   []*DynamicBoolOpt{{}},
			utils.MetaAttributes: []*DynamicBoolOpt{{}},
			utils.MetaChargers:   []*DynamicBoolOpt{{}},
			utils.MetaEEs:        []*DynamicBoolOpt{{}},
			utils.MetaRates:      []*DynamicBoolOpt{{}},
			utils.MetaStats:      []*DynamicBoolOpt{{}},
			utils.MetaThresholds: []*DynamicBoolOpt{{}},
			utils.MetaRefund:     []*DynamicBoolOpt{{}},
			utils.MetaRerate:     []*DynamicBoolOpt{{}},
			utils.MetaStore:      []*DynamicBoolOpt{{value: true}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.cdrsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, recieved %+v", eMap, rcv)
	}
}

func TestCdrsCfgClone(t *testing.T) {
	ban := &CdrsCfg{
		Enabled: true,

		SMCostRetries: 1,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"}}},
			utils.MetaAttributes: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
			utils.MetaThresholds: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaActions:    {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
			utils.MetaEEs:        {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
		},
		OnlineCDRExports: []string{"randomVal"},
		ExtraFields:      utils.RSRParsers{},
		Opts:             &CdrsOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaChargers][0].Values[1] = ""; ban.Conns[utils.MetaChargers][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaAttributes][0].Values[1] = ""; ban.Conns[utils.MetaAttributes][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaThresholds][0].Values[1] = ""; ban.Conns[utils.MetaThresholds][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaStats][0].Values[1] = ""; ban.Conns[utils.MetaStats][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaActions][0].Values[1] = ""; ban.Conns[utils.MetaActions][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaEEs][0].Values[1] = ""; ban.Conns[utils.MetaEEs][0].Values[1] != "*conn1" {
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
		ExtraFields: utils.RSRParsers{
			{
				Rules: "Rule1",
			},
		},

		SMCostRetries: 2,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"*localhost"}}},
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaThresholds: {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
			utils.MetaActions:    {{Values: []string{"*localhost"}}},
			utils.MetaEEs:        {{Values: []string{"*localhost"}}},
			utils.MetaRates:      {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:   {{Values: []string{"*localhost"}}},
		},
		OnlineCDRExports: []string{},
		Opts: &CdrsOpts{
			Accounts: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Attributes: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Chargers: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Export: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Rates: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Stats: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
		},
	}

	v2 := &CdrsCfg{
		Enabled: true,
		ExtraFields: utils.RSRParsers{
			{
				Rules: "Rule2",
			},
		},

		SMCostRetries: 1,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"*birpc"}}},
			utils.MetaAttributes: {{Values: []string{"*birpc"}}},
			utils.MetaThresholds: {{Values: []string{"*birpc"}}},
			utils.MetaStats:      {{Values: []string{"*birpc"}}},
			utils.MetaActions:    {{Values: []string{"*birpc"}}},
			utils.MetaEEs:        {{Values: []string{"*birpc"}}},
			utils.MetaRates:      {{Values: []string{"*birpc"}}},
			utils.MetaAccounts:   {{Values: []string{"*birpc"}}},
		},
		OnlineCDRExports: []string{"val1"},
		Opts: &CdrsOpts{
			Accounts: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Attributes: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Chargers: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Export: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Rates: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Stats: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Thresholds: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
		},
	}

	expected := &CdrsJsonCfg{
		Enabled:              utils.BoolPointer(true),
		Extra_fields:         &[]string{"Rule2"},
		Session_cost_retries: utils.IntPointer(1),
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"*birpc"}}},
			utils.MetaAttributes: {{Values: []string{"*birpc"}}},
			utils.MetaThresholds: {{Values: []string{"*birpc"}}},
			utils.MetaStats:      {{Values: []string{"*birpc"}}},
			utils.MetaActions:    {{Values: []string{"*birpc"}}},
			utils.MetaEEs:        {{Values: []string{"*birpc"}}},
			utils.MetaRates:      {{Values: []string{"*birpc"}}},
			utils.MetaAccounts:   {{Values: []string{"*birpc"}}},
		},
		Online_cdr_exports: &[]string{"val1"},
		Opts: &CdrsOptsJson{
			Accounts: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Attributes: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Chargers: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Export: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Rates: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Stats: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Thresholds: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
		},
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

func TestCdrsCfgCloneSection(t *testing.T) {
	cdrsCfg := &CdrsCfg{
		Enabled: false,
		ExtraFields: utils.RSRParsers{
			{
				Rules: "Rule1",
			},
		},
		SMCostRetries: 2,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"*localhost"}}},
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaThresholds: {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
			utils.MetaActions:    {{Values: []string{"*localhost"}}},
			utils.MetaEEs:        {{Values: []string{"*localhost"}}},
		},
		OnlineCDRExports: []string{},
		Opts:             &CdrsOpts{},
	}

	exp := &CdrsCfg{
		Enabled: false,
		ExtraFields: utils.RSRParsers{
			{
				Rules: "Rule1",
			},
		},
		SMCostRetries: 2,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaChargers:   {{Values: []string{"*localhost"}}},
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaThresholds: {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
			utils.MetaActions:    {{Values: []string{"*localhost"}}},
			utils.MetaEEs:        {{Values: []string{"*localhost"}}},
		},
		OnlineCDRExports: []string{},
		Opts:             &CdrsOpts{},
	}

	rcv := cdrsCfg.CloneSection()
	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestCdrsCfgdiffCdrsJsonCfg(t *testing.T) {
	coj := &CdrsOptsJson{}
	co := &CdrsOpts{}
	bl := true
	d := &CdrsJsonCfg{}
	v2 := &CdrsCfg{
		Enabled: true,
		Opts:    co,
	}
	v1 := &CdrsCfg{
		Enabled: false,
		Opts:    co,
	}
	exp := &CdrsJsonCfg{
		Enabled: &bl,
		Opts:    coj,
	}
	rcv := diffCdrsJsonCfg(d, v1, v2)

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
