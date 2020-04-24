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

func TestResourceSConfigloadFromJsonCfg(t *testing.T) {
	var rlcfg, expected ResourceSConfig
	if err := rlcfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rlcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rlcfg)
	}
	if err := rlcfg.loadFromJsonCfg(new(ResourceSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rlcfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, rlcfg)
	}
	cfgJSONStr := `{
"resources": {								// Resource service (*new)
	"enabled": true,						// starts ResourceLimiter service: <true|false>.
	"store_interval": "1s",					// dump cache regularly to dataDB, 0 - dump at start/shutdown: <""|$dur>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
},	
}`
	expected = ResourceSConfig{
		Enabled:             true,
		StoreInterval:       time.Duration(time.Second),
		ThresholdSConns:     []string{},
		PrefixIndexedFields: &[]string{"index1", "index2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRlcCfg, err := jsnCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = rlcfg.loadFromJsonCfg(jsnRlcCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rlcfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, rlcfg)
	}
}

func TestResourceSConfigAsMapInterface(t *testing.T) {
	var rlcfg ResourceSConfig

	cfgJSONStr := `{
	"resources": {								
		"enabled": false,						
		"store_interval": "",					
		"thresholds_conns": [],					
		"indexed_selects":true,					
		"prefix_indexed_fields": [],			
		"nested_fields": false,					
	},	
}`
	eMap := map[string]interface{}{
		"enabled":               false,
		"store_interval":        "",
		"thresholds_conns":      []string{},
		"indexed_selects":       true,
		"string_indexed_fields": []string{},
		"prefix_indexed_fields": []string{},
		"nested_fields":         false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRlcCfg, err := jsnCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = rlcfg.loadFromJsonCfg(jsnRlcCfg); err != nil {
		t.Error(err)
	} else if rcv := rlcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"resources": {								
			"enabled": false,						
			"store_interval": "7m",					
			"thresholds_conns": ["*internal"],					
			"indexed_selects":true,					
			"prefix_indexed_fields": ["prefix_indexed_fields1","prefix_indexed_fields2"],			
			"nested_fields": false,					
		},	
	}`
	eMap = map[string]interface{}{
		"enabled":               false,
		"store_interval":        "7m0s",
		"thresholds_conns":      []string{"*internal"},
		"indexed_selects":       true,
		"string_indexed_fields": []string{},
		"prefix_indexed_fields": []string{"prefix_indexed_fields1", "prefix_indexed_fields2"},
		"nested_fields":         false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnRlcCfg, err := jsnCfg.ResourceSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = rlcfg.loadFromJsonCfg(jsnRlcCfg); err != nil {
		t.Error(err)
	} else if rcv := rlcfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
