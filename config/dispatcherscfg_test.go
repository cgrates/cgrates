// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestDispatcherSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		String_indexed_fields: &[]string{"*req.prefix", "*req.indexed"},
		Prefix_indexed_fields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		Suffix_indexed_fields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		ExistsIndexedFields:   &[]string{"*req.exists", "*req.indexed", "*req.fields"},
		Attributes_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Nested_fields:         utils.BoolPointer(true),
		Any_subsystem:         utils.BoolPointer(true),
	}
	expected := &DispatcherSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.prefix", "*req.indexed"},
		PrefixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		SuffixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		ExistsIndexedFields: &[]string{"*req.exists", "*req.indexed", "*req.fields"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		NestedFields:        true,
		AnySubsystem:        true,
	}
	jsnCfg := NewDefaultCGRConfig()
	if err := jsnCfg.dispatcherSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.dispatcherSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsnCfg.dispatcherSCfg))
	}
}

func TestDispatcherSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
			"prefix_indexed_fields": [],
            "suffix_indexed_fields": [],
            "exists_indexed_fields": [],
			"nested_fields": false,
			"attributes_conns": [],
		},
		
}`
	eMap := map[string]any{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.ExistsIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
		utils.AnySubsystemCfg:        true,
		utils.PreventLoopCfg:         false,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDispatcherSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
            "string_indexed_fields": ["*req.prefix"],
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
            "suffix_indexed_fields": ["*req.prefix"],
            "exists_indexed_fields": ["*req.exists"],
			"nested_fields": false,
			"attributes_conns": ["*internal:*attributes", "*conn1"],
			"prevent_loop": true
		},
		
}`
	eMap := map[string]any{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.StringIndexedFieldsCfg: []string{"*req.prefix"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix"},
		utils.ExistsIndexedFieldsCfg: []string{"*req.exists"},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{"*internal", "*conn1"},
		utils.AnySubsystemCfg:        true,
		utils.PreventLoopCfg:         true,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDispatcherSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
		"dispatchers":{},
}`
	eMap := map[string]any{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.ExistsIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
		utils.AnySubsystemCfg:        true,
		utils.PreventLoopCfg:         false,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
func TestDispatcherSCfgClone(t *testing.T) {
	ban := &DispatcherSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &[]string{"*req.prefix", "*req.indexed"},
		PrefixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		SuffixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		NestedFields:        true,
		AnySubsystem:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.AttributeSConns[1] = ""; ban.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if (*rcv.StringIndexedFields)[0] = ""; (*ban.StringIndexedFields)[0] != "*req.prefix" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	if (*rcv.PrefixIndexedFields)[0] = ""; (*ban.PrefixIndexedFields)[0] != "*req.prefix" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = ""; (*ban.SuffixIndexedFields)[0] != "*req.prefix" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	ban = nil
	rcv = ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
}
