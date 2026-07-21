// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestFilterSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSONS := &FilterSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaResources: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaTrends:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRankings:  {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
		},
	}
	expected := &FilterSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaResources: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"}}},
			utils.MetaTrends:    {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends), "*conn1"}}},
			utils.MetaRankings:  {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings), "*conn1"}}},
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
				"*stats": [{"connIDs": ["*internal:*stats", "*conn1"]}],
				"*resources": [{"connIDs": ["*internal:*resources", "*conn1"]}],
				"*accounts": [{"connIDs": ["*internal:*accounts", "*conn1"]}],
				"*trends": [{"connIDs": ["*internal:*trends", "*conn1"]}],
				"*rankings": [{"connIDs": ["*internal:*rankings", "*conn1"]}]
			}
	},
}`
	eMap := map[string]any{
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaResources: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaTrends:    {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
			utils.MetaRankings:  {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
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
		utils.ConnsCfg: map[string][]*DynamicConns{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.filterSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestFilterSCfgClone(t *testing.T) {
	ban := &FilterSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"}}},
			utils.MetaResources: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), "*conn1"}}},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaStats][0].ConnIDs[1] = ""; ban.Conns[utils.MetaStats][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaResources][0].ConnIDs[1] = ""; ban.Conns[utils.MetaResources][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaAccounts][0].ConnIDs[1] = ""; ban.Conns[utils.MetaAccounts][0].ConnIDs[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffFilterSJsonCfg(t *testing.T) {
	var d *FilterSJsonCfg

	v1 := &FilterSCfg{
		Conns: map[string][]*DynamicConns{},
	}

	v2 := &FilterSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaResources: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{"*localhost"}}},
		},
	}

	expected := &FilterSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaResources: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{"*localhost"}}},
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
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaResources: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{"*localhost"}}},
		},
	}

	exp := &FilterSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:     {{ConnIDs: []string{"*localhost"}}},
			utils.MetaResources: {{ConnIDs: []string{"*localhost"}}},
			utils.MetaAccounts:  {{ConnIDs: []string{"*localhost"}}},
		},
	}

	rcv := fltrSCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
