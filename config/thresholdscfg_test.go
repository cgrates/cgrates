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

func TestThresholdSCfgloadFromJsonCfg(t *testing.T) {
	var thscfg, expected ThresholdSCfg
	if err := thscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, thscfg)
	}
	if err := thscfg.loadFromJsonCfg(new(ThresholdSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, thscfg)
	}
	cfgJSONStr := `{
"thresholds": {								// Threshold service (*new)
	"enabled": false,						// starts ThresholdS service: <true|false>.
	"store_interval": "2h",					// dump cache regularly to dataDB, 0 - dump at start/shutdown: <""|$dur>
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
	},		
}`
	expected = ThresholdSCfg{
		StoreInterval:       time.Duration(time.Hour * 2),
		PrefixIndexedFields: &[]string{"index1", "index2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.ThresholdSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = thscfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, thscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, thscfg)
	}
}

func TestThresholdSCfgAsMapInterface(t *testing.T) {
	var thscfg ThresholdSCfg

	cfgJSONStr := `{
		"thresholds": {								
			"enabled": false,						
			"store_interval": "",					
			"indexed_selects":true,					
			"prefix_indexed_fields": [],			
			"nested_fields": false,					
		},		
}`
	eMap := map[string]interface{}{
		"enabled":               false,
		"store_interval":        "",
		"indexed_selects":       true,
		"string_indexed_fields": []string{},
		"prefix_indexed_fields": []string{},
		"nested_fields":         false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.ThresholdSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = thscfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if rcv := thscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"thresholds": {								
			"enabled": true,						
			"store_interval": "96h",					
			"indexed_selects":true,
			"string_indexed_fields": ["string","indexed","fields"],					
			"prefix_indexed_fields": ["prefix_indexed_fields1","prefix_indexed_fields2"],			
			"nested_fields": true,					
		},		
}`
	eMap = map[string]interface{}{
		"enabled":               true,
		"store_interval":        "96h0m0s",
		"indexed_selects":       true,
		"string_indexed_fields": []string{"string", "indexed", "fields"},
		"prefix_indexed_fields": []string{"prefix_indexed_fields1", "prefix_indexed_fields2"},
		"nested_fields":         true,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnThSCfg, err := jsnCfg.ThresholdSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = thscfg.loadFromJsonCfg(jsnThSCfg); err != nil {
		t.Error(err)
	} else if rcv := thscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
