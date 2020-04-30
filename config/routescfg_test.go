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

func TestRouteSCfgloadFromJsonCfg(t *testing.T) {
	var supscfg, expected RouteSCfg
	if err := supscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, supscfg)
	}
	if err := supscfg.loadFromJsonCfg(new(RouteSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, supscfg)
	}
	cfgJSONStr := `{
"routes": {								// Route service 
	"enabled": false,						// starts RouteS service: <true|false>.
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
	"attributes_conns": [],					// address where to reach the AttributeS <""|127.0.0.1:2013>
	"resources_conns": [],					// address where to reach the Resource service, empty to disable functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the Stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"default_ratio":1,
},
}`
	expected = RouteSCfg{
		PrefixIndexedFields: &[]string{"index1", "index2"},
		AttributeSConns:     []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		DefaultRatio:        1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.RouteSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, supscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, supscfg)
	}
}

func TestRouteSCfgAsMapInterface(t *testing.T) {
	var supscfg RouteSCfg
	cfgJSONStr := `{
	"routes": {
		"enabled": false,
		"indexed_selects":true,
		"prefix_indexed_fields": [],
		"nested_fields": false,
		"attributes_conns": [],
		"resources_conns": [],
		"stats_conns": [],
		"rals_conns": [],
		"default_ratio":1
	},
}`
	eMap := map[string]interface{}{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{},
		"string_indexed_fields": []string{},
		"nested_fields":         false,
		"attributes_conns":      []string{},
		"resources_conns":       []string{},
		"stats_conns":           []string{},
		"rals_conns":            []string{},
		"default_ratio":         1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.RouteSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if rcv := supscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"routes": {
			"enabled": false,
			"indexed_selects":true,
			"prefix_indexed_fields": ["prefix","indexed","fields"],
			"nested_fields": false,
			"attributes_conns": ["*internal"],
			"resources_conns": ["*internal"],
			"stats_conns": ["*internal"],
			"rals_conns": ["*internal"],
			"default_ratio":1
		},
	}`
	eMap = map[string]interface{}{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{"prefix", "indexed", "fields"},
		"string_indexed_fields": []string{},
		"nested_fields":         false,
		"attributes_conns":      []string{"*internal"},
		"resources_conns":       []string{"*internal"},
		"stats_conns":           []string{"*internal"},
		"rals_conns":            []string{"*internal"},
		"default_ratio":         1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.RouteSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if rcv := supscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
