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

func TestAttributeSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &AttributeSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(false),
		Resources_conns:       &[]string{"*internal", "*conn1"},
		Stats_conns:           &[]string{"*internal", "*conn1"},
		Accounts_conns:        &[]string{"*internal", "*conn1"},
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &AttributeSCfg{
		Enabled:             true,
		AccountSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), "*conn1"},
		StatSConns:          []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "*conn1"},
		ResourceSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), "*conn1"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		Opts: &AttributesOpts{
			ProcessRuns: 1,
		},
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.attributeSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.attributeSCfg) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.attributeSCfg))
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
	"opts": {
		"*processRuns": 3,
	},					
	},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StatSConnsCfg:          []string{utils.MetaInternal},
		utils.ResourceSConnsCfg:      []string{utils.MetaInternal},
		utils.AccountSConnsCfg:       []string{utils.MetaInternal},
		utils.StringIndexedFieldsCfg: []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.index1", "*req.index2"},
		utils.IndexedSelectsCfg:      true,
		utils.NestedFieldsCfg:        false,
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAttributeIDsCfg: []string(nil),
			utils.MetaProcessRunsCfg:  3,
			utils.MetaProfileRunsCfg:  0,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.attributeSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestAttributeSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
     "attributes": {
           "suffix_indexed_fields": ["*req.index1","*req.index2"],
           "nested_fields": true,
           "enabled": true,
           "opts": {
			"*processRuns": 7,
		},	
     },
}`
	expectedMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StatSConnsCfg:          []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.AccountSConnsCfg:       []string{},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{"*req.index1", "*req.index2"},
		utils.NestedFieldsCfg:        true,
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAttributeIDsCfg: []string(nil),
			utils.MetaProcessRunsCfg:  7,
			utils.MetaProfileRunsCfg:  0,
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if newMap := cgrCfg.attributeSCfg.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
		t.Errorf("Expected %+v \n, receieved %+v", utils.ToJSON(expectedMap), utils.ToJSON(newMap))
	}
}

func TestAttributeSCfgAsMapInterface3(t *testing.T) {
	myJSONStr := `
{
    "attributes": {}
}
`
	expectedMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.StatSConnsCfg:          []string{},
		utils.ResourceSConnsCfg:      []string{},
		utils.AccountSConnsCfg:       []string{},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.OptsCfg: map[string]interface{}{
			utils.MetaAttributeIDsCfg: []string(nil),
			utils.MetaProcessRunsCfg:  1,
			utils.MetaProfileRunsCfg:  0,
		},
	}
	if conv, err := NewCGRConfigFromJSONStringWithDefaults(myJSONStr); err != nil {
		t.Error(err)
	} else if newMap := conv.attributeSCfg.AsMapInterface(); !reflect.DeepEqual(expectedMap, newMap) {
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
			ProcessRuns: 1,
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
			ProcessRuns: 1,
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
		Opts:                  &AttributesOptsJson{},
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
