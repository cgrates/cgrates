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

func TestRouteSCfgloadFromJsonCfg(t *testing.T) {
	cfgJSON := &RouteSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Attributes_conns:      &[]string{utils.MetaInternal, "conn1"},
		Resources_conns:       &[]string{utils.MetaInternal, "conn1"},
		Stats_conns:           &[]string{utils.MetaInternal, "conn1"},
		Rates_conns:           &[]string{utils.MetaInternal, "conn1"},
		Accounts_conns:        &[]string{utils.MetaInternal, "conn1"},
		Default_ratio:         utils.IntPointer(10),
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &RouteSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "conn1"},
		RateSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS), "conn1"},
		AccountSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "conn1"},
		DefaultRatio:        10,
		NestedFields:        true,
		Opts: &RoutesOpts{
			Context: []*utils.DynamicStringOpt{
				{
					Value: utils.MetaRoutes,
				},
			},
			ProfileCount: []*utils.DynamicFloat64Opt{
				{
					Value: 1,
				},
			},
			IgnoreErrors: []*utils.DynamicBoolOpt{
				{
					Value: false,
				},
			},
			MaxCost: []*utils.DynamicInterfaceOpt{
				{
					Value: utils.EmptyString,
				},
			},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.routeSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.routeSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.routeSCfg))
	}
}

func TestRouteSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"routes": {},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.StatSConnsCfg:          []string{},
		utils.RateSConnsCfg:          []string{},
		utils.AccountSConnsCfg:       []string{},
		utils.DefaultRatioCfg:        1,
		utils.OptsCfg: map[string]interface{}{
			utils.OptsContext:         utils.MetaRoutes,
			utils.MetaProfileCountCfg: float64(1),
			utils.MetaIgnoreErrorsCfg: false,
			utils.MetaMaxCostCfg:      utils.EmptyString,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
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
			"nested_fields": true,
			"attributes_conns": ["*internal:*attributes", "conn1"],
			"resources_conns": ["*internal:*resources", "conn1"],
			"stats_conns": ["*internal:*stats", "conn1"],
			"rals_conns": ["*internal:*responder", "conn1"],
			"rates_conns": ["*internal:*rates", "conn1"],
			"accounts_conns": ["*internal:*accounts", "conn1"],
			"default_ratio":2,
		},
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.IndexedSelectsCfg:      false,
		utils.StringIndexedFieldsCfg: []string{"*req.string"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed"},
		utils.NestedFieldsCfg:        true,
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal, "conn1"},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal, "conn1"},
		utils.StatSConnsCfg:          []string{utils.MetaInternal, "conn1"},
		utils.RateSConnsCfg:          []string{utils.MetaInternal, "conn1"},
		utils.AccountSConnsCfg:       []string{utils.MetaInternal, "conn1"},
		utils.DefaultRatioCfg:        2,
		utils.OptsCfg: map[string]interface{}{
			utils.OptsContext:         utils.MetaRoutes,
			utils.MetaProfileCountCfg: float64(1),
			utils.MetaIgnoreErrorsCfg: false,
			utils.MetaMaxCostCfg:      utils.EmptyString,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.routeSCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRouteSCfgClone(t *testing.T) {
	ban := &RouteSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "conn1"},
		DefaultRatio:        10,
		NestedFields:        true,
		Opts:                &RoutesOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.AttributeSConns[1] = ""; ban.AttributeSConns[1] != "conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ResourceSConns[1] = ""; ban.ResourceSConns[1] != "conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "conn1" {
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
		AttributeSConns:     []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		StatSConns:          []string{"*localhost"},
		DefaultRatio:        2,
		Opts: &RoutesOpts{
			Context: []*utils.DynamicStringOpt{
				{
					Value: utils.MetaAny,
				},
			},
			IgnoreErrors: []*utils.DynamicBoolOpt{
				{
					Value: true,
				},
			},
			MaxCost: []*utils.DynamicInterfaceOpt{
				{
					Value: 5,
				},
			},
			Limit:  utils.IntPointer(1),
			Offset: utils.IntPointer(1),
			ProfileCount: []*utils.DynamicFloat64Opt{
				{
					Value: 1,
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
		AttributeSConns:     []string{"*birpc"},
		ResourceSConns:      []string{"*birpc"},
		StatSConns:          []string{"*birpc"},
		DefaultRatio:        3,
		Opts: &RoutesOpts{
			Context: []*utils.DynamicStringOpt{
				{
					Value: utils.MetaRoutes,
				},
			},
			IgnoreErrors: []*utils.DynamicBoolOpt{
				{
					Value: false,
				},
			},
			MaxCost: []*utils.DynamicInterfaceOpt{
				{
					Value: utils.MetaEventCost,
				},
			},
			Limit:  utils.IntPointer(2),
			Offset: utils.IntPointer(2),
			ProfileCount: []*utils.DynamicFloat64Opt{
				{
					Value: 2,
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
		Attributes_conns:      &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Stats_conns:           &[]string{"*birpc"},
		Default_ratio:         utils.IntPointer(3),
		Opts: &RoutesOptsJson{
			Context: map[string]string{
				utils.EmptyString: utils.MetaRoutes,
			},
			IgnoreErrors: map[string]bool{
				utils.EmptyString: false,
			},
			MaxCost: map[string]string{
				utils.EmptyString: utils.EmptyString,
			},
			Limit:  utils.IntPointer(2),
			Offset: utils.IntPointer(2),
			ProfileCount: map[string]float64{
				utils.EmptyString: 2,
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
