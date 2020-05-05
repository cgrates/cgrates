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
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
},	
}`
	expected = ChargerSCfg{
		Enabled:             true,
		AttributeSConns:     []string{},
		PrefixIndexedFields: &[]string{"index1", "index2"},
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
	var chgscfg ChargerSCfg
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
		"enabled":               false,
		"attributes_conns":      []string{},
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{},
		"nested_fields":         false,
		"string_indexed_fields": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnChgCfg, err := jsnCfg.ChargerServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = chgscfg.loadFromJsonCfg(jsnChgCfg); err != nil {
		t.Error(err)
	} else if rcv := chgscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"chargers": {								
			"enabled": false,						
			"attributes_conns": ["*internal"],					
			"indexed_selects":true,					
			"prefix_indexed_fields": [],			
			"nested_fields": false,					
		},	
	}`
	eMap = map[string]interface{}{
		"enabled":               false,
		"attributes_conns":      []string{"*internal"},
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{},
		"nested_fields":         false,
		"string_indexed_fields": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnChgCfg, err := jsnCfg.ChargerServJsonCfg(); err != nil {
		t.Error(err)
	} else if err = chgscfg.loadFromJsonCfg(jsnChgCfg); err != nil {
		t.Error(err)
	} else if rcv := chgscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
