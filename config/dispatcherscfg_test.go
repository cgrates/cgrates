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
	var daCfg, expected DispatcherSCfg
	if err := daCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	if err := daCfg.loadFromJsonCfg(new(DispatcherSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	cfgJSONStr := `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
			//"string_indexed_fields": [],
			"prefix_indexed_fields": [],
			"nested_fields": false,
			"attributes_conns": [],
		},
		
}`
	expected = DispatcherSCfg{
		Enabled:             false,
		IndexedSelects:      true,
		PrefixIndexedFields: &[]string{},
		AttributeSConns:     []string{},
		NestedFields:        false,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, daCfg) {
		t.Errorf("Expected: %+v,\nRecived: %+v", utils.ToJSON(expected), utils.ToJSON(daCfg))
	}
}

func TestDispatcherSCfgAsMapInterface(t *testing.T) {
	var daCfg, expected DispatcherSCfg
	if err := daCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	if err := daCfg.loadFromJsonCfg(new(DispatcherSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, daCfg)
	}
	cfgJSONStr := `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
			//"string_indexed_fields": [],
			"prefix_indexed_fields": [],
			"nested_fields": false,
			"attributes_conns": [],
		},
		
}`
	eMap := map[string]interface{}{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{},
		"nested_fields":         false,
		"attributes_conns":      []string{},
		"string_indexed_fields": []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if rcv := daCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
			"string_indexed_fields": ["*req.string","*req.indexed","*req.fields"],
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],
			"nested_fields": false,
			"attributes_conns": ["*internal"],
		},
		
}`
	eMap = map[string]interface{}{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{"*req.prefix", "*req.indexed", "*req.fields"},
		"nested_fields":         false,
		"attributes_conns":      []string{"*internal"},
		"string_indexed_fields": []string{"*req.string", "*req.indexed", "*req.fields"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if rcv := daCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nRecived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
