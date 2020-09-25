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

func TestDispatcherSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &DispatcherSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Prefix_indexed_fields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		Suffix_indexed_fields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		Attributes_conns:      &[]string{"*internal"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &DispatcherSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		SuffixIndexedFields: &[]string{"*req.prefix", "*req.indexed", "*req.fields"},
		AttributeSConns:     []string{"*internal:*attributes"},
		NestedFields:        true,
	}
	if jsnCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsnCfg.dispatcherSCfg.loadFromJsonCfg(jsonCfg); err != nil {
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
			"nested_fields": false,
			"attributes_conns": [],
		},
		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
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
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
            "suffix_indexed_fields": [],
			"nested_fields": false,
			"attributes_conns": ["*internal"],
		},
		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{"*internal"},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}

func TestDispatcherSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
		"dispatchers":{},
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.AttributeSConnsCfg:     []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.dispatcherSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
