/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
		t.Errorf("Expected: %+v ,received: %+v", expected, daCfg)
	}
	if err := daCfg.loadFromJsonCfg(new(DispatcherSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, daCfg)
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
		t.Errorf("Expected: %+v,\nReceived: %+v", utils.ToJSON(expected), utils.ToJSON(daCfg))
	}
}

func TestDispatcherSCfgAsMapInterface(t *testing.T) {
	var daCfg, expected DispatcherSCfg
	if err := daCfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, daCfg)
	}
	if err := daCfg.loadFromJsonCfg(new(DispatcherSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(daCfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, daCfg)
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
	eMap := map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"dispatchers":{
			"enabled": false,
			"indexed_selects":true,
			"string_indexed_fields": ["string","indexed","fields"],
			"prefix_indexed_fields": ["prefix","indexed","fields"],
			"nested_fields": false,
			"attributes_conns": ["*internal"],
		},
		
}`
	eMap = map[string]any{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{"prefix", "indexed", "fields"},
		"nested_fields":         false,
		"attributes_conns":      []string{"*internal"},
		"string_indexed_fields": []string{"string", "indexed", "fields"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnDaCfg, err := jsnCfg.DispatcherSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = daCfg.loadFromJsonCfg(jsnDaCfg); err != nil {
		t.Error(err)
	} else if rcv := daCfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestDispatcherCfgloadFromJsonCfg(t *testing.T) {
	bl := false
	slc := []string{"val1", "val2"}

	d := DispatcherSCfg{}

	js := DispatcherSJsonCfg{
		Enabled:               &bl,
		Indexed_selects:       &bl,
		String_indexed_fields: &slc,
		Prefix_indexed_fields: &slc,
		Nested_fields:         &bl,
		Attributes_conns:      &slc,
	}

	exp := DispatcherSCfg{
		Enabled:             bl,
		IndexedSelects:      bl,
		StringIndexedFields: &slc,
		PrefixIndexedFields: &slc,
		AttributeSConns:     slc,
		NestedFields:        bl,
	}

	err := d.loadFromJsonCfg(&js)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(d, exp) {
		t.Errorf("received %v, expected %v", d, exp)
	}
}

func TestDispatcherSCfgAsMapInterface2(t *testing.T) {
	bl := false
	slc := []string{"val1", "val2"}

	dsp := DispatcherSCfg{
		Enabled:             bl,
		IndexedSelects:      bl,
		StringIndexedFields: &slc,
		PrefixIndexedFields: &slc,
		AttributeSConns:     slc,
		NestedFields:        bl,
	}

	exp := map[string]any{
		utils.EnabledCfg:             bl,
		utils.IndexedSelectsCfg:      bl,
		utils.StringIndexedFieldsCfg: slc,
		utils.PrefixIndexedFieldsCfg: slc,
		utils.AttributeSConnsCfg:     slc,
		utils.NestedFieldsCfg:        bl,
	}

	rcv := dsp.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("received %v, expected %v", rcv, exp)
	}
}
