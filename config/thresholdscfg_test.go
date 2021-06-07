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
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestThresholdSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("2"),
		String_indexed_fields: &[]string{"*req.prefix"},
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
		Actions_conns:         &[]string{utils.MetaInternal},
	}
	expected := &ThresholdSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2,
		StringIndexedFields: &[]string{"*req.prefix"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		ActionSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions)},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.thresholdSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.thresholdSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.thresholdSCfg))
	}
}

func TestThresholdSCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &ThresholdSJsonCfg{
		Store_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.thresholdSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestThresholdSCfgAsMapInterfaceCase1(t *testing.T) {
	cfgJSONStr := `{
		"thresholds": {},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.StoreIntervalCfg:       "",
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.ActionSConnsCfg:        []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.thresholdSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expextec %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestThresholdSCfgAsMapInterfaceCase2(t *testing.T) {
	cfgJSONStr := `{
		"thresholds": {								
			"enabled": true,						
			"store_interval": "96h",					
			"indexed_selects": false,	
            "string_indexed_fields": ["*req.string"],
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],	
            "suffix_indexed_fields": ["*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"],		
			"nested_fields": true,					
			"actions_conns": ["*internal"],
		},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StoreIntervalCfg:       "96h0m0s",
		utils.IndexedSelectsCfg:      false,
		utils.StringIndexedFieldsCfg: []string{"*req.string"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"},
		utils.NestedFieldsCfg:        true,
		utils.ActionSConnsCfg:        []string{utils.MetaInternal},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.thresholdSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expexted %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
func TestThresholdSCfgClone(t *testing.T) {
	ban := &ThresholdSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
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

func TestDiffThresholdSJsonCfg(t *testing.T) {
	var d *ThresholdSJsonCfg

	v1 := &ThresholdSCfg{
		Enabled:             false,
		IndexedSelects:      false,
		StoreInterval:       1 * time.Second,
		StringIndexedFields: &[]string{"req.index1"},
		PrefixIndexedFields: &[]string{"req.index2"},
		SuffixIndexedFields: &[]string{"req.index3"},
		ActionSConns:        []string{},
		NestedFields:        false,
	}

	v2 := &ThresholdSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2 * time.Second,
		StringIndexedFields: &[]string{"req.index11"},
		PrefixIndexedFields: &[]string{"req.index22"},
		SuffixIndexedFields: &[]string{"req.index33"},
		ActionSConns:        []string{"*internal"},
		NestedFields:        true,
	}

	expected := &ThresholdSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"req.index11"},
		Prefix_indexed_fields: &[]string{"req.index22"},
		Suffix_indexed_fields: &[]string{"req.index33"},
		Actions_conns:         &[]string{"*internal"},
		Nested_fields:         utils.BoolPointer(true),
	}

	rcv := diffThresholdSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &ThresholdSJsonCfg{}
	rcv = diffThresholdSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
