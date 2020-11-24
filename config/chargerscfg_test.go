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

func TestChargerSCfgloadFromJsonCfg(t *testing.T) {
	jsonCfg := &ChargerSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Attributes_conns:      &[]string{utils.MetaInternal, "*conn1"},
		String_indexed_fields: &[]string{"*req.Field1"},
		Prefix_indexed_fields: &[]string{"*req.Field1", "*req.Field2"},
		Suffix_indexed_fields: &[]string{"*req.Field1", "*req.Field2"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &ChargerSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		StringIndexedFields: &[]string{"*req.Field1"},
		PrefixIndexedFields: &[]string{"*req.Field1", "*req.Field2"},
		SuffixIndexedFields: &[]string{"*req.Field1", "*req.Field2"},
		NestedFields:        true,
	}
	if jsncfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsncfg.chargerSCfg.loadFromJSONCfg(jsonCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsncfg.chargerSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsncfg.chargerSCfg))
	}
}

func TestChargerSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"chargers": {								
		"enabled": false,						
		"attributes_conns": [],					
		"indexed_selects":true,
		"prefix_indexed_fields": [],
		"nested_fields": false,					
	},	
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.AttributeSConnsCfg:     []string{},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.SuffixIndexedFieldsCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.chargerSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}

func TestChargerSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"chargers": {								
			"enabled": false,						
			"attributes_conns": ["*internal:*attributes", "*conn1"],					
			"indexed_selects":true,			
            "string_indexed_fields": ["*req.Field1","*req.Field2","*req.Field3"],
			 "prefix_indexed_fields": ["*req.DestinationPrefix"],
             "suffix_indexed_fields": ["*req.Field1","*req.Field2","*req.Field3"],		
			"nested_fields": false,					
		},	
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.AttributeSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.IndexedSelectsCfg:      true,
		utils.StringIndexedFieldsCfg: []string{"*req.Field1", "*req.Field2", "*req.Field3"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.DestinationPrefix"},
		utils.NestedFieldsCfg:        false,
		utils.SuffixIndexedFieldsCfg: []string{"*req.Field1", "*req.Field2", "*req.Field3"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.chargerSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}

func TestChargerSCfgClone(t *testing.T) {
	ban := &ChargerSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		AttributeSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), "*conn1"},
		StringIndexedFields: &[]string{"*req.Field1"},
		PrefixIndexedFields: &[]string{"*req.Field1", "*req.Field2"},
		SuffixIndexedFields: &[]string{"*req.Field1", "*req.Field2"},
		NestedFields:        true,
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.AttributeSConns[1] = ""; ban.AttributeSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = ""; (*ban.StringIndexedFields)[0] != "*req.Field1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = ""; (*ban.PrefixIndexedFields)[0] != "*req.Field1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = ""; (*ban.SuffixIndexedFields)[0] != "*req.Field1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
