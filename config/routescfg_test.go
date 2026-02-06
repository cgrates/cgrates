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
	"github.com/ericlagergren/decimal"
)

func TestRouteSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &RouteSJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(true),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Exists_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Notexists_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaResources:  {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.MetaInternal, "conn1"}}},
		},
		Default_ratio: utils.IntPointer(10),
		Nested_fields: utils.BoolPointer(true),
		Opts:          &RoutesOptsJson{},
	}
	expected := &RouteSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		ExistsIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NotExistsIndexedFields: &[]string{"*req.index1", "*req.index2"},
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "conn1"}}},
			utils.MetaResources:  {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), "conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "conn1"}}},
		},
		DefaultRatio: 10,
		NestedFields: true,
		Opts: &RoutesOpts{
			Context:      []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			ProfileCount: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			IgnoreErrors: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			MaxCost:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			Limit:        []*DynamicIntPointerOpt{},
			Offset:       []*DynamicIntPointerOpt{},
			MaxItems:     []*DynamicIntPointerOpt{},
			Usage:        []*DynamicDecimalOpt{{value: RatesUsageDftOpt}},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.routeSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.routeSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.routeSCfg))
	}

	cfgJSON.Opts.Usage = []*DynamicInterfaceOpt{
		{
			Tenant: "cgrates.org",
			Value:  "error",
		},
	}
	errExpect := "can't convert <error> to decimal"
	if err := jsonCfg.routeSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but received \n %v", errExpect, err.Error())
	}

	cfgJSON = nil
	if err := jsonCfg.routeSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}
}

func TestRouteSCfgloadFromJsonCfgOpts(t *testing.T) {
	routeSOpt := &RoutesOpts{}
	if err := routeSOpt.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestRouteSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"routes": {},
}`
	eMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.ExistsIndexedFieldsCfg:    []string{},
		utils.NotExistsIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:           false,
		utils.ConnsCfg:                  map[string][]*DynamicStringSliceOpt{},
		utils.DefaultRatioCfg:           1,
		utils.OptsCfg: map[string]any{
			utils.OptsContext:         []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			utils.MetaLimitCfg:        []*DynamicIntPointerOpt{},
			utils.MetaOffsetCfg:       []*DynamicIntPointerOpt{},
			utils.MetaMaxItemsCfg:     []*DynamicIntPointerOpt{},
			utils.MetaProfileCountCfg: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			utils.MetaIgnoreErrorsCfg: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			utils.MetaMaxCostCfg:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			utils.MetaUsage:           []*DynamicDecimalOpt{{value: RoutesUsageDftOpt}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRouteSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"routes": {
			"enabled": true,
			"indexed_selects":false,
			"string_indexed_fields": ["*req.string"],
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
			"suffix_indexed_fields": ["*req.prefix","*req.indexed"],
			"exists_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
			"notexists_indexed_fields": ["*req.prefix","*req.indexed"],
			"nested_fields": true,
			"conns": {
				"*attributes": [{"Values": ["*internal:*attributes", "conn1"]}],
				"*resources": [{"Values": ["*internal:*resources", "conn1"]}],
				"*stats": [{"Values": ["*internal:*stats", "conn1"]}],
				"*rates": [{"Values": ["*internal:*rates", "conn1"]}],
				"*accounts": [{"Values": ["*internal:*accounts", "conn1"]}]
			},
			"default_ratio":2,
		},
	}`
	eMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.IndexedSelectsCfg:         false,
		utils.StringIndexedFieldsCfg:    []string{"*req.string"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.prefix", "*req.indexed"},
		utils.ExistsIndexedFieldsCfg:    []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.NotExistsIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed"},
		utils.NestedFieldsCfg:           true,
		utils.ConnsCfg: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaResources:  {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaRates:      {{Values: []string{utils.MetaInternal, "conn1"}}},
			utils.MetaAccounts:   {{Values: []string{utils.MetaInternal, "conn1"}}},
		},
		utils.DefaultRatioCfg: 2,
		utils.OptsCfg: map[string]any{
			utils.OptsContext:         []*DynamicStringOpt{{value: RoutesContextDftOpt}},
			utils.MetaLimitCfg:        []*DynamicIntPointerOpt{},
			utils.MetaOffsetCfg:       []*DynamicIntPointerOpt{},
			utils.MetaMaxItemsCfg:     []*DynamicIntPointerOpt{},
			utils.MetaProfileCountCfg: []*DynamicIntPointerOpt{{value: RoutesProfileCountDftOpt}},
			utils.MetaIgnoreErrorsCfg: []*DynamicBoolOpt{{value: RoutesIgnoreErrorsDftOpt}},
			utils.MetaMaxCostCfg:      []*DynamicInterfaceOpt{{Value: RoutesMaxCostDftOpt}},
			utils.MetaUsage:           []*DynamicDecimalOpt{{value: RoutesUsageDftOpt}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRouteSCfgClone(t *testing.T) {
	ban := &RouteSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		ExistsIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NotExistsIndexedFields: &[]string{"*req.index1", "*req.index2"},
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "conn1"}}},
			utils.MetaResources:  {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "conn1"}}},
			utils.MetaStats:      {{Values: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "conn1"}}},
		},
		DefaultRatio: 10,
		NestedFields: true,
		Opts:         &RoutesOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.Conns[utils.MetaAttributes][0].Values[1] = ""; ban.Conns[utils.MetaAttributes][0].Values[1] != "conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaResources][0].Values[1] = ""; ban.Conns[utils.MetaResources][0].Values[1] != "conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.Conns[utils.MetaStats][0].Values[1] = ""; ban.Conns[utils.MetaStats][0].Values[1] != "conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = ""; (*ban.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = ""; (*ban.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = ""; (*ban.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffRouteSJsonCfg(t *testing.T) {
	var d *RouteSJsonCfg

	v1 := &RouteSCfg{
		Enabled:             false,
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaResources:  {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
			utils.MetaRates:      {{Values: []string{"*localhost"}}},
			utils.MetaAccounts:   {{Values: []string{"*localhost"}}},
		},
		DefaultRatio: 2,
		Opts: &RoutesOpts{
			Context: []*DynamicStringOpt{
				{
					value: utils.MetaAny,
				},
			},
			IgnoreErrors: []*DynamicBoolOpt{
				{
					value: true,
				},
			},
			MaxCost: []*DynamicInterfaceOpt{
				{
					Value: 5,
				},
			},
			Limit: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(2),
				},
			},
			Offset: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(2),
				},
			},
			ProfileCount: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt(nil, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(3), nil),
			},
			MaxItems: []*DynamicIntPointerOpt{
				{
					FilterIDs: []string{"id1"},
					Tenant:    "cgrates.net",
					value:     utils.IntPointer(1),
				},
			},
		},
	}

	v2 := &RouteSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.index11"},
		PrefixIndexedFields: &[]string{"*req.index22"},
		SuffixIndexedFields: &[]string{"*req.index33"},
		NestedFields:        true,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{"*birpc"}}},
			utils.MetaResources:  {{Values: []string{"*birpc"}}},
			utils.MetaStats:      {{Values: []string{"*birpc"}}},
			utils.MetaRates:      {{Values: []string{"*birpc"}}},
			utils.MetaAccounts:   {{Values: []string{"*birpc"}}},
		},
		DefaultRatio: 3,
		Opts: &RoutesOpts{
			Context: []*DynamicStringOpt{
				{
					value: utils.MetaSessionS,
				},
			},
			IgnoreErrors: []*DynamicBoolOpt{
				{
					value: false,
				},
			},
			MaxCost: []*DynamicInterfaceOpt{
				{
					Value: 6,
				},
			},
			Limit: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(3),
				},
			},
			Offset: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(3),
				},
			},
			ProfileCount: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(2),
				},
			},
			Usage: []*DynamicDecimalOpt{
				NewDynamicDecimalOpt(nil, "cgrates.org", decimal.WithContext(utils.DecimalContext).SetUint64(2), nil),
			},
			MaxItems: []*DynamicIntPointerOpt{
				{
					FilterIDs: []string{"id2"},
					Tenant:    "cgrates.org",
					value:     utils.IntPointer(2),
				},
			},
		},
	}

	expected := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{"*birpc"}}},
			utils.MetaResources:  {{Values: []string{"*birpc"}}},
			utils.MetaStats:      {{Values: []string{"*birpc"}}},
			utils.MetaRates:      {{Values: []string{"*birpc"}}},
			utils.MetaAccounts:   {{Values: []string{"*birpc"}}},
		},
		Default_ratio: utils.IntPointer(3),
		Opts: &RoutesOptsJson{
			Context: []*DynamicInterfaceOpt{
				{
					Value: utils.MetaSessionS,
				},
			},
			IgnoreErrors: []*DynamicInterfaceOpt{
				{
					Value: false,
				},
			},
			MaxCost: []*DynamicInterfaceOpt{
				{
					Value: 6,
				},
			},
			Limit: []*DynamicInterfaceOpt{
				{
					Value: utils.IntPointer(3),
				},
			},
			Offset: []*DynamicInterfaceOpt{
				{
					Value: utils.IntPointer(3),
				},
			},
			ProfileCount: []*DynamicInterfaceOpt{
				{
					Value: utils.IntPointer(2),
				},
			},
			Usage: []*DynamicInterfaceOpt{
				{
					Tenant: "cgrates.org",
					Value:  "2",
				},
			},
			MaxItems: []*DynamicInterfaceOpt{
				{
					FilterIDs: []string{"id2"},
					Tenant:    "cgrates.org",
					Value:     utils.IntPointer(2),
				},
			},
		},
	}

	rcv := diffRouteSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &RouteSJsonCfg{
		Opts: &RoutesOptsJson{},
	}
	rcv = diffRouteSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestRouteSCloneSection(t *testing.T) {
	routeScfg := &RouteSCfg{
		Enabled:             false,
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaResources:  {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
		},
		DefaultRatio: 2,

		Opts: &RoutesOpts{
			Context: []*DynamicStringOpt{
				{
					value: utils.MetaAny,
				},
			},
			IgnoreErrors: []*DynamicBoolOpt{
				{
					value: true,
				},
			},
			MaxCost: []*DynamicInterfaceOpt{
				{
					Value: 5,
				},
			},
			Limit: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
			Offset: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
			ProfileCount: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
		},
	}

	exp := &RouteSCfg{
		Enabled:             false,
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Conns: map[string][]*DynamicStringSliceOpt{
			utils.MetaAttributes: {{Values: []string{"*localhost"}}},
			utils.MetaResources:  {{Values: []string{"*localhost"}}},
			utils.MetaStats:      {{Values: []string{"*localhost"}}},
		},
		DefaultRatio: 2,
		Opts: &RoutesOpts{
			Context: []*DynamicStringOpt{
				{
					value: utils.MetaAny,
				},
			},
			IgnoreErrors: []*DynamicBoolOpt{
				{
					value: true,
				},
			},
			MaxCost: []*DynamicInterfaceOpt{
				{
					Value: 5,
				},
			},
			Limit: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
			Offset: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
			ProfileCount: []*DynamicIntPointerOpt{
				{
					value: utils.IntPointer(1),
				},
			},
		},
	}

	rcv := routeScfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
