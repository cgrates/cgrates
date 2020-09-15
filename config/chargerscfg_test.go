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
	var chgscfg, expected ChargerSCfg
	if err := chgscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chgscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, chgscfg)
	}
	if err := chgscfg.loadFromJsonCfg(new(ChargerSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(chgscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, chgscfg)
	}
	cfgJSONStr := `{
"chargers": {								// Charger service
	"enabled": true,						// starts charger service: <true|false>.
	"attributes_conns": [],					// address where to reach the AttributeS <""|127.0.0.1:2013>
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["*req.index1", "*req.index2"],			// query indexes based on these fields for faster processing
},	
}`
	expected = ChargerSCfg{
		Enabled:             true,
		AttributeSConns:     []string{},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnChgCfg, err := jsnCfg.ChargerServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = chgscfg.loadFromJsonCfg(jsnChgCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, chgscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, chgscfg)
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
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.chargerSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}

func TestChargerSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"chargers": {								
			"enabled": false,						
			"attributes_conns": ["*internal"],					
			"indexed_selects":true,					
			"prefix_indexed_fields": [],			
			"nested_fields": false,					
		},	
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.AttributeSConnsCfg:     []string{"*internal"},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
		utils.SuffixIndexedFieldsCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.chargerSCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("Expected %+v, recieved %+v", eMap, rcv)
	}
}

func TestChargerSCfgAsMapInterface2(t *testing.T) {
	cfgJSONStr := `{
      "chargers": {
          "prefix_indexed_fields": ["*req.DestinationPrefix"],
          "suffix_indexed_fields": ["*req.Field1","*req.Field2","*req.Field3"],
      },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.AttributeSConnsCfg:     []string{},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{"*req.DestinationPrefix"},
		utils.NestedFieldsCfg:        false,
		utils.SuffixIndexedFieldsCfg: []string{"*req.Field1", "*req.Field2", "*req.Field3"},
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.chargerSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v, received %+v", eMap, rcv)
	}
}
