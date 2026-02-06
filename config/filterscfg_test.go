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

func TestFilterSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONS := &FilterSJsonCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaResources: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:  {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaTrends:    {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRankings:  {{Values: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	expected := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaResources: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"}}},
			utils.MetaAccounts:  {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"}}},
			utils.MetaTrends:    {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends), "*conn1"}}},
			utils.MetaRankings:  {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings), "*conn1"}}},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.filterSCfg.loadFromJSONCfg(cfgJSONS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.filterSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.filterSCfg))
	}
	cfgJSONS = nil
	if err := jsnCfg.filterSCfg.loadFromJSONCfg(cfgJSONS); err != nil {
		t.Error(err)
	}
}

func TestFilterSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"filters": {
			"conns": {
				"*stats": [{"Values": ["*internal:*stats", "*conn1"]}],
				"*resources": [{"Values": ["*internal:*resources", "*conn1"]}],
				"*accounts": [{"Values": ["*internal:*accounts", "*conn1"]}],
				"*trends": [{"Values": ["*internal:*trends", "*conn1"]}],
				"*rankings": [{"Values": ["*internal:*rankings", "*conn1"]}]
			}
	},
}`
	eMap := map[string]any{
		utils.ConnsCfg: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaResources: {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:  {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaTrends:    {{Values: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRankings:  {{Values: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFilterSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
      "filters": {}
}`
	eMap := map[string]any{
		utils.ConnsCfg: map[string][]*DynamicStringSliceOpt{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFilterSCfgClone(t *testing.T) {
	ban := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaResources: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"}}},
			utils.MetaAccounts:  {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), "*conn1"}}},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaStats][0].Values[1] = ""; ban.Conns[utils.MetaStats][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaResources][0].Values[1] = ""; ban.Conns[utils.MetaResources][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaAccounts][0].Values[1] = ""; ban.Conns[utils.MetaAccounts][0].Values[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffFilterSJsonCfg(t *testing.T) {
	var d *FilterSJsonCfg

	v1 := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{},
	}

	v2 := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{"*localhost"}}},
			utils.MetaResources: {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:  {{Values: []string{"*localhost"}}},
		},
	}

	expected := &FilterSJsonCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{"*localhost"}}},
			utils.MetaResources: {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:  {{Values: []string{"*localhost"}}},
		},
	}

	rcv := diffFilterSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected2 := &FilterSJsonCfg{}

	rcv = diffFilterSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestFilterSCloneSection(t *testing.T) {
	fltrSCfg := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{"*localhost"}}},
			utils.MetaResources: {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:  {{Values: []string{"*localhost"}}},
		},
	}

	exp := &FilterSCfg{
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaStats:     {{Values: []string{"*localhost"}}},
			utils.MetaResources: {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:  {{Values: []string{"*localhost"}}},
		},
	}

	rcv := fltrSCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
