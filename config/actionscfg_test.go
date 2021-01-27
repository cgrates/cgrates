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

func TestActionSCfgLoadFromJSONCfg(t *testing.T) {
	jsonCfg := &ActionSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(false),
		Tenants:               &[]string{"itsyscom.com"},
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &ActionSCfg{
		Enabled:             true,
		EEsConns:            []string{},
		CDRsConns:           []string{},
		ThresholdSConns:     []string{},
		StatSConns:          []string{},
		IndexedSelects:      false,
		Tenants:             &[]string{"itsyscom.com"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	jsnCfg := NewDefaultCGRConfig()
	if err = jsnCfg.actionSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsnCfg.actionSCfg) {
		t.Errorf("\nExpecting <%+v>,\n Received <%+v>", utils.ToJSON(expected), utils.ToJSON(jsnCfg.actionSCfg))
	}
}

func TestActionSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
"actions": {								
	"enabled": true,
	"tenants": ["itsyscom.com"],
	"indexed_selects": false,
	"string_indexed_fields": ["*req.index1"],			
	"prefix_indexed_fields": ["*req.index1","*req.index2"],		
    "suffix_indexed_fields": ["*req.index1"],
	"nested_fields": true,						
	},		
}`

	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.EEsConnsCfg:            []string{},
		utils.ThresholdSConnsCfg:     []string{},
		utils.StatSConnsCfg:          []string{},
		utils.CDRsConnsCfg:           []string{},
		utils.Tenants:                []string{"itsyscom.com"},
		utils.IndexedSelectsCfg:      false,
		utils.StringIndexedFieldsCfg: []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.index1", "*req.index2"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.index1"},
		utils.NestedFieldsCfg:        true,
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.actionSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected: %+v\n Received: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestActionSCfgClone(t *testing.T) {
	ban := &ActionSCfg{
		Enabled:             true,
		Tenants:             &[]string{"itsyscom.com"},
		IndexedSelects:      false,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if (*rcv.Tenants)[0] = ""; (*ban.Tenants)[0] != "itsyscom.com" {
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
