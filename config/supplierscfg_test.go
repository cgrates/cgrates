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
	"strings"
	"testing"
)

func TestSupplierSCfgloadFromJsonCfg(t *testing.T) {
	var supscfg, expected SupplierSCfg
	if err := supscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, supscfg)
	}
	if err := supscfg.loadFromJsonCfg(new(SupplierSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(supscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, supscfg)
	}
	cfgJSONStr := `{
"suppliers": {								// Supplier service (*new)
	"enabled": false,						// starts SupplierS service: <true|false>.
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
	"attributes_conns": [],					// address where to reach the AttributeS <""|127.0.0.1:2013>
	"rals_conns": [
		{"address": "*internal"},			// address where to reach the RALs for cost/accounting  <*internal>
	],
	"resources_conns": [],					// address where to reach the Resource service, empty to disable functionality: <""|*internal|x.y.z.y:1234>
	"stats_conns": [],						// address where to reach the Stat service, empty to disable stats functionality: <""|*internal|x.y.z.y:1234>
},
}`
	expected = SupplierSCfg{
		PrefixIndexedFields: &[]string{"index1", "index2"},
		AttributeSConns:     []*HaPoolConfig{},
		RALsConns:           []*HaPoolConfig{{Address: "*internal"}},
		ResourceSConns:      []*HaPoolConfig{},
		StatSConns:          []*HaPoolConfig{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromReader(strings.NewReader(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnSupSCfg, err := jsnCfg.SupplierSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = supscfg.loadFromJsonCfg(jsnSupSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, supscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, supscfg)
	}
}
