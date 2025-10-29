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

func TestSupplierSCfgloadFromJsonCfg(t *testing.T) {
	var supscfg, expected SupplierSCfg
	if err := supscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, supscfg)
	}
	if err := supscfg.loadFromJsonCfg(new(SupplierSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, supscfg)
	}
	cfgJSONStr := `{
"suppliers": {								// Supplier service (*new)
	"enabled": false,						// starts SupplierS service: <true|false>.
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
	"attributes_conns": [],					// address where to reach the AttributeS <""|127.0.0.1:2013>
	"resources_conns": [],					// address where to reach the Resource service, empty to disable functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the Stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
	"default_ratio":1,
},
}`
	expected = SupplierSCfg{
		PrefixIndexedFields: &[]string{"index1", "index2"},
		AttributeSConns:     []string{},
		ResourceSConns:      []string{},
		StatSConns:          []string{},
		DefaultRatio:        1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, supscfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, supscfg)
	}
}

func TestSupplierSCfgAsMapInterface(t *testing.T) {
	var supscfg SupplierSCfg
	cfgJSONStr := `{
	"suppliers": {
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
	eMap := map[string]any{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{},
		"string_indexed_fields": []string{},
		"nested_fields":         false,
		"attributes_conns":      []string{},
		"resources_conns":       []string{},
		"stats_conns":           []string{},
		"default_ratio":         1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if rcv := supscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"suppliers": {
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
	eMap = map[string]any{
		"enabled":               false,
		"indexed_selects":       true,
		"prefix_indexed_fields": []string{"prefix", "indexed", "fields"},
		"string_indexed_fields": []string{},
		"nested_fields":         false,
		"attributes_conns":      []string{"*internal"},
		"resources_conns":       []string{"*internal"},
		"stats_conns":           []string{"*internal"},
		"default_ratio":         1,
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if rcv := supscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestSupplierSCfgloadFromJsonCfg2(t *testing.T) {
	slc := []string{"val1", "val2"}

	spl := SupplierSCfg{
		Enabled:             false,
		IndexedSelects:      false,
		StringIndexedFields: &slc,
		PrefixIndexedFields: &slc,
		AttributeSConns:     []string{"val1", "val2"},
		ResourceSConns:      []string{"val1", "val2"},
		StatSConns:          []string{"val1", "val2"},
		RALsConns:           []string{"val1", "val2"},
		DefaultRatio:        1,
		NestedFields:        false,
	}

	bl := true
	nm := 2
	slc2 := []string{"val3", "val4"}

	js := SupplierSJsonCfg{
		Enabled:               &bl,
		Indexed_selects:       &bl,
		String_indexed_fields: &slc2,
		Prefix_indexed_fields: &slc2,
		Nested_fields:         &bl,
		Attributes_conns:      &slc2,
		Resources_conns:       &slc2,
		Stats_conns:           &slc2,
		Rals_conns:            &slc2,
		Default_ratio:         &nm,
	}

	exp := SupplierSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &slc2,
		PrefixIndexedFields: &slc2,
		AttributeSConns:     []string{"val3", "val4"},
		ResourceSConns:      []string{"val3", "val4"},
		StatSConns:          []string{"val3", "val4"},
		RALsConns:           []string{"val3", "val4"},
		DefaultRatio:        2,
		NestedFields:        true,
	}

	err := spl.loadFromJsonCfg(&js)

	if err != nil {
		t.Error("was not exoecting an error")
	}

	if !reflect.DeepEqual(spl, exp) {
		t.Errorf("expected %v, recived %v", exp, spl)
	}
}

func TestSupplierSCfgAsMapInterface2(t *testing.T) {
	slc := []string{"val1", "val2"}

	spl := SupplierSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StringIndexedFields: &slc,
		PrefixIndexedFields: &slc,
		AttributeSConns:     []string{"val1", "val2"},
		ResourceSConns:      []string{"val1", "val2"},
		StatSConns:          []string{"val1", "val2"},
		RALsConns:           []string{"val1", "val2"},
		DefaultRatio:        1,
		NestedFields:        false,
	}

	exp := map[string]any{
		utils.EnabledCfg:             spl.Enabled,
		utils.IndexedSelectsCfg:      spl.IndexedSelects,
		utils.StringIndexedFieldsCfg: []string{"val1", "val2"},
		utils.PrefixIndexedFieldsCfg: []string{"val1", "val2"},
		utils.AttributeSConnsCfg:     []string{"val1", "val2"},
		utils.ResourceSConnsCfg:      []string{"val1", "val2"},
		utils.StatSConnsCfg:          []string{"val1", "val2"},
		utils.DefaultRatioCfg:        spl.DefaultRatio,
		utils.NestedFieldsCfg:        spl.NestedFields,
	}

	rcv := spl.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %+v, recived %+v", exp, rcv)
	}
}
