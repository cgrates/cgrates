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
	tests := []struct {
		name     string
		jsonCfg  *CdrsJsonCfg
		expected *CdrsCfg
		expErr   bool
	}{
		{
			name: "Complete CdrsJsonCfg",
			jsonCfg: &CdrsJsonCfg{
				Enabled:              utils.BoolPointer(true),
				Session_cost_retries: utils.IntPointer(1),
				Conns: map[string][]*DynamicConns{
					utils.MetaChargers:   {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaStats:      {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaActions:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaEEs:        {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaRates:      {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
					utils.MetaAccounts:   {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
				},
				Opts: &CdrsOptsJson{
					Accounts: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Attributes: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Chargers: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Export: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Rates: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Stats: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Thresholds: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Refund: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Rerate: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
					Store: []*DynamicInterfaceOpt{
						{
							Tenant: "cgrates.org",
							Value:  true,
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       true,
				SMCostRetries: 1,
				Conns: map[string][]*DynamicConns{
					utils.MetaChargers:   {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"}}},
					utils.MetaAttributes: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
					utils.MetaThresholds: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
					utils.MetaStats:      {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
					utils.MetaActions:    {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
					utils.MetaEEs:        {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
					utils.MetaRates:      {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), "*conn1"}}},
					utils.MetaAccounts:   {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"}}},
				},
				OnlineCDRExports: nil,
				ExtraFields:      utils.RSRParsers{},
				Opts: &CdrsOpts{
					Accounts: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Attributes: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Chargers: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Export: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Rates: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Stats: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Thresholds: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Refund: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Rerate: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
					Store: []*DynamicBoolOpt{
						{
							Tenant: "cgrates.org",
							value:  true,
						},
						{},
					},
				},
			},
		},
		{
			name: "Accounts",
			jsonCfg: &CdrsJsonCfg{
				Enabled:              utils.BoolPointer(false),
				Extra_fields:         utils.SliceStringPointer([]string{}),
				Session_cost_retries: nil,
				Conns:                map[string][]*DynamicConns{},
				Opts: &CdrsOptsJson{
					Accounts: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Attributes",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Attributes: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Chargers",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Chargers: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Export",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Export: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Rates",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Rates: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Stats",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Stats: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Thresholds",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Thresholds: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Refund",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Refund: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Rerate",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Rerate: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name: "Store",
			jsonCfg: &CdrsJsonCfg{
				Opts: &CdrsOptsJson{
					Store: []*DynamicInterfaceOpt{
						{
							Value: "2",
						},
					},
				},
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
			expErr: true,
		},
		{
			name:    "nil CdrsJsonCfg",
			jsonCfg: nil,
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
		},
		{
			name: "Nil Opts",
			jsonCfg: &CdrsJsonCfg{
				Opts: nil,
			},
			expected: &CdrsCfg{
				Enabled:       false,
				ExtraFields:   utils.RSRParsers{},
				Conns:         map[string][]*DynamicConns{},
				SMCostRetries: 5,
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
					Store:      []*DynamicBoolOpt{{}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()
			if err := jsnCfg.cdrsCfg.loadFromJSONCfg(tt.jsonCfg); err != nil && !tt.expErr {
				t.Error(err)
			} else if !reflect.DeepEqual(utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.cdrsCfg)) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.cdrsCfg))
			}
		})
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
		"extraFields": ["~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"],
		"sessionCostRetries": 5,
		"conns": {
			"*chargers": [{"connIDs": ["*internal:*chargers","*conn1"]}],
			"*attributes": [{"connIDs": ["*internal:*attributes","*conn1"]}],
			"*thresholds": [{"connIDs": ["*internal:*thresholds","*conn1"]}],
			"*stats": [{"connIDs": ["*internal:*stats","*conn1"]}],
			"*actions": [{"connIDs": ["*internal:*actions","*conn1"]}],
			"*ees": [{"connIDs": ["*internal:*ees","*conn1"]}],
			"*rates": [{"connIDs": ["*internal:*rates","*conn1"]}],
			"*accounts": [{"connIDs": ["*internal:*accounts","*conn1"]}]
		},
		"onlineCDRExports":["http_localhost", "amqp_localhost", "http_test_file"],
	},
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.ExtraFieldsCfg:     []string{"~*req.PayPalAccount", "~*req.LCRProfile", "~*req.ResourceID"},
		utils.SessionCostRetires: 5,
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAttributes: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaStats:      {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRates:      {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:   {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
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
              "*chargers": [{"connIDs": ["conn1", "conn2"]}],
              "*attributes": [{"connIDs": ["*internal"]}],
              "*ees": [{"connIDs": ["conn1"]}]
          },
       },
}`
	eMap := map[string]any{
		utils.EnabledCfg:         true,
		utils.ExtraFieldsCfg:     []string{},
		utils.SessionCostRetires: 5,
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"conn1", "conn2"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*internal"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"conn1"}}},
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), "*conn1"}}},
			utils.MetaAttributes: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
			utils.MetaStats:      {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaActions:    {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), "*conn1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"}}},
		},
		OnlineCDRExports: []string{"randomVal"},
		ExtraFields:      utils.RSRParsers{},
		Opts:             &CdrsOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaChargers][0].ConnIDs[1] = ""; ban.Conns[utils.MetaChargers][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaAttributes][0].ConnIDs[1] = ""; ban.Conns[utils.MetaAttributes][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaThresholds][0].ConnIDs[1] = ""; ban.Conns[utils.MetaThresholds][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaStats][0].ConnIDs[1] = ""; ban.Conns[utils.MetaStats][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaActions][0].ConnIDs[1] = ""; ban.Conns[utils.MetaActions][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaEEs][0].ConnIDs[1] = ""; ban.Conns[utils.MetaEEs][0].ConnIDs[1] != "*conn1" {
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaStats:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAccounts:   {{ConnIDs: []string{"*localhost"}}},
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
			Refund: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Rerate: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					value:  false,
				},
			},
			Store: []*DynamicBoolOpt{
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaStats:      {{ConnIDs: []string{"*birpc"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*birpc"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*birpc"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAccounts:   {{ConnIDs: []string{"*birpc"}}},
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
			Refund: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Rerate: []*DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					value:  true,
				},
			},
			Store: []*DynamicBoolOpt{
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*birpc"}}},
			utils.MetaStats:      {{ConnIDs: []string{"*birpc"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*birpc"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*birpc"}}},
			utils.MetaRates:      {{ConnIDs: []string{"*birpc"}}},
			utils.MetaAccounts:   {{ConnIDs: []string{"*birpc"}}},
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
			Refund: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Rerate: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
			Store: []*DynamicInterfaceOpt{
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaStats:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
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
		Conns: map[string][]*DynamicConns{
			utils.MetaChargers:   {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAttributes: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaStats:      {{ConnIDs: []string{"*localhost"}}},
			utils.MetaActions:    {{ConnIDs: []string{"*localhost"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"*localhost"}}},
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
