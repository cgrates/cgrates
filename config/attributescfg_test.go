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

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestAttributeSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AttributeSJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(false),
		Resources_conns:          &[]string{"*internal", "*conn1"},
		Stats_conns:              &[]string{"*internal", "*conn1"},
		Accounts_conns:           &[]string{"*internal", "*conn1"},
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index1"},
		Exists_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Notexists_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:            utils.BoolPointer(true),
	}
	expected := &AttributeSCfg{
		Enabled:                true,
		AccountSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"},
		StatSConns:             []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ResourceSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		IndexedSelects:         false,
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1"},
		ExistsIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NotExistsIndexedFields: &[]string{"*req.index1"},
		NestedFields:           true,
		Opts: &AttributesOpts{
			ProfileIDs:           []*utils.DynamicStringSliceOpt{},
			ProcessRuns:          []*utils.DynamicIntOpt{},
			ProfileRuns:          []*utils.DynamicIntOpt{},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.attributeSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.attributeSCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.attributeSCfg))
	}

	jsonCfg = nil
	if err := jsnCfg.attributeSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	}
}

func TestAttributeSLoadFromJsonCfgOpts(t *testing.T) {
	attrOpt := &AttributesOpts{
		ProfileIDs: []*utils.DynamicStringSliceOpt{
			{
				Values: []string{},
			},
		},
		ProcessRuns: []*utils.DynamicIntOpt{
			{
				Value: 1,
			},
		},
		ProfileRuns: []*utils.DynamicIntOpt{
			{
				Value: 0,
			},
		},
		ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
			{
				Value: false,
			},
		},
	}

	exp := &AttributesOpts{
		ProfileIDs: []*utils.DynamicStringSliceOpt{
			{
				Values: []string{},
			},
		},
		ProcessRuns: []*utils.DynamicIntOpt{
			{
				Value: 1,
			},
		},
		ProfileRuns: []*utils.DynamicIntOpt{
			{
				Value: 0,
			},
		},
		ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
			{
				Value: false,
			},
		},
	}

	attrOpt.loadFromJSONCfg(nil)
	if !reflect.DeepEqual(attrOpt, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(attrOpt))
	}
}

func TestAttributeSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"attributes": {								
	"enabled": true,	
	"stats_conns": ["*internal"],			
	"resources_conns": ["*internal"],		
	"accounts_conns": ["*internal"],			
	"prefix_indexed_fields": ["*req.index1","*req.index2"],		
    "string_indexed_fields": ["*req.index1"],
	"exists_indexed_fields": ["*req.index1","*req.index2"],		
    "notexists_indexed_fields": ["*req.index1"],
	"opts": {
		"*processRuns": [
				{
					"Value": 3,
				},
			],
	},					
	},		
}`
	eMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.StatSConnsCfg:             []string{utils.MetaInternal},
		utils.ResourceSConnsCfg:         []string{utils.MetaInternal},
		utils.AccountSConnsCfg:          []string{utils.MetaInternal},
		utils.StringIndexedFieldsCfg:    []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.ExistsIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.NotExistsIndexedFieldsCfg: []string{"*req.index1"},
		utils.IndexedSelectsCfg:         true,
		utils.NestedFieldsCfg:           false,
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs: []*utils.DynamicStringSliceOpt{},
			utils.MetaProcessRunsCfg: []*utils.DynamicIntOpt{
				{
					Value: 3,
				},
			},
			utils.MetaProfileRunsCfg:       []*utils.DynamicIntOpt{},
			utils.MetaProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.attributeSCfg.AsMapInterface(""); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAttributeSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
     "attributes": {
           "suffix_indexed_fields": ["*req.index1","*req.index2"],
		   "notexists_indexed_fields": ["*req.index1","*req.index2"],
           "nested_fields": true,
           "enabled": true,
           "opts": {
			"*processRuns": [
				{
					"Value": 7,
				},
			],
		},	
     },
}`
	expectedMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.StatSConnsCfg:             []string{},
		utils.ResourceSConnsCfg:         []string{},
		utils.AccountSConnsCfg:          []string{},
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.ExistsIndexedFieldsCfg:    []string{},
		utils.NotExistsIndexedFieldsCfg: []string{"*req.index1", "*req.index2"},
		utils.NestedFieldsCfg:           true,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs: []*utils.DynamicStringSliceOpt{},
			utils.MetaProcessRunsCfg: []*utils.DynamicIntOpt{
				{
					Value: 7,
				},
			},
			utils.MetaProfileRunsCfg:       []*utils.DynamicIntOpt{},
			utils.MetaProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.attributeSCfg.AsMapInterface(""); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v \n, receieved %+v", utils.ToJSON(expectedMap), utils.ToJSON(newMap))
	}
}

func TestAttributeSCfgAsMapInterface3(t *testing.T) {
	myJSONStr := `
{
    "attributes": {}
}
`
	expectedMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.StatSConnsCfg:             []string{},
		utils.ResourceSConnsCfg:         []string{},
		utils.AccountSConnsCfg:          []string{},
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.ExistsIndexedFieldsCfg:    []string{},
		utils.NotExistsIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:           false,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs:           []*utils.DynamicStringSliceOpt{},
			utils.MetaProcessRunsCfg:       []*utils.DynamicIntOpt{},
			utils.MetaProfileRunsCfg:       []*utils.DynamicIntOpt{},
			utils.MetaProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	if conv, err := NewCGRConfigFromJSONStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := conv.attributeSCfg.AsMapInterface(""); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v, receieved %+v", expectedMap, newMap)
	}
}

func TestAttributeSCfgClone(t *testing.T) {
	ban := &AttributeSCfg{
		Enabled:             true,
		AccountSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		Opts:                &AttributesOpts{},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.AccountSConns[1] = ""; ban.AccountSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.StatSConns[1] = ""; ban.StatSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if rcv.ResourceSConns[1] = ""; ban.ResourceSConns[1] != "*conn1" {
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

func TestDiffAttributeSJsonCfg(t *testing.T) {
	var d *AttributeSJsonCfg

	v1 := &AttributeSCfg{
		Enabled:             false,
		StatSConns:          []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		AccountSConns:       []string{"*localhost"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		Opts: &AttributesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Values: []string{"prf1"},
				},
			},
			ProcessRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     1,
				},
			},
			ProfileRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     1,
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					Value:  false,
				},
			},
		},
	}

	v2 := &AttributeSCfg{
		Enabled:             true,
		StatSConns:          []string{"*birpc"},
		ResourceSConns:      []string{"*birpc"},
		AccountSConns:       []string{"*birpc"},
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.Field1"},
		PrefixIndexedFields: nil,
		SuffixIndexedFields: nil,
		NestedFields:        false,
		Opts: &AttributesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Values: []string{"prf2"},
				},
			},
			ProcessRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     2,
				},
			},
			ProfileRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     2,
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
		},
	}

	expected := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Stats_conns:           &[]string{"*birpc"},
		Resources_conns:       &[]string{"*birpc"},
		Accounts_conns:        &[]string{"*birpc"},
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.Field1"},
		Prefix_indexed_fields: nil,
		Suffix_indexed_fields: nil,
		Nested_fields:         utils.BoolPointer(false),
		Opts: &AttributesOptsJson{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Values: []string{"prf2"},
				},
			},
			ProcessRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     2,
				},
			},
			ProfileRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     2,
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
		},
	}

	rcv := diffAttributeSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v2_2 := v1
	expected2 := &AttributeSJsonCfg{
		Opts: &AttributesOptsJson{},
	}
	rcv = diffAttributeSJsonCfg(d, v1, v2_2)
	if !reflect.DeepEqual(rcv, expected2) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected2), utils.ToJSON(rcv))
	}
}

func TestAttributeSCloneSection(t *testing.T) {
	attrCfg := &AttributeSCfg{
		Enabled:             false,
		StatSConns:          []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		AccountSConns:       []string{"*localhost"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		Opts: &AttributesOpts{
			ProcessRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     1,
				},
			},
		},
	}

	exp := &AttributeSCfg{
		Enabled:             false,
		StatSConns:          []string{"*localhost"},
		ResourceSConns:      []string{"*localhost"},
		AccountSConns:       []string{"*localhost"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{},
		PrefixIndexedFields: &[]string{},
		SuffixIndexedFields: &[]string{},
		NestedFields:        true,
		Opts: &AttributesOpts{
			ProcessRuns: []*utils.DynamicIntOpt{
				{
					FilterIDs: []string{},
					Value:     1,
				},
			},
		},
	}

	rcv := attrCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestLoadAttributeSCfg(t *testing.T) {
	alS := &AttributeSCfg{}
	ctx := &context.Context{}
	jsnCfg := new(mockDb)
	cgrcfg := &CGRConfig{}
	if err := alS.Load(ctx, jsnCfg, cgrcfg); err == nil || err != utils.ErrNotImplemented {
		t.Errorf("Expected error <%v>, Received error <%v>", utils.ErrNotImplemented, err)
	}

}
