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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestStatSCfgloadFromJsonCfg(t *testing.T) {
	var statscfg, expected StatSCfg
	if err := statscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, statscfg)
	}
	if err := statscfg.loadFromJsonCfg(new(StatServJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, statscfg)
	}
	cfgJSONStr := `{
"stats": {									// Stat service (*new)
	"enabled": false,						// starts Stat service: <true|false>.
	"store_interval": "2s",					// dump cache regularly to dataDB, 0 - dump at start/shutdown: <""|$dur>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["index1", "index2"],			// query indexes based on these fields for faster processing
},	
}`
	expected = StatSCfg{
		StoreInterval:       time.Duration(time.Second * 2),
		ThresholdSConns:     []string{},
		PrefixIndexedFields: &[]string{"index1", "index2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStatSCfg, err := jsnCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = statscfg.loadFromJsonCfg(jsnStatSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, statscfg) {
		t.Errorf("Expected: %+v , received: %+v", expected, statscfg)
	}
}

func TestStatSCfgAsMapInterface(t *testing.T) {
	var statscfg, expected StatSCfg
	if err := statscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, statscfg)
	}
	if err := statscfg.loadFromJsonCfg(new(StatServJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, statscfg)
	}
	cfgJSONStr := `{
		"stats": {							
			"enabled": false,				
			"store_interval": "",			
			"store_uncompressed_limit": 0,	
			"thresholds_conns": [],			
			"indexed_selects":true,			
			"prefix_indexed_fields": [],	
			"nested_fields": false,	
		},	
		}`
	eMap := map[string]any{
		"enabled":                  false,
		"store_interval":           "",
		"store_uncompressed_limit": 0,
		"thresholds_conns":         []string{},
		"indexed_selects":          true,
		"prefix_indexed_fields":    []string{},
		"nested_fields":            false,
		"string_indexed_fields":    []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStatSCfg, err := jsnCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = statscfg.loadFromJsonCfg(jsnStatSCfg); err != nil {
		t.Error(err)
	} else if rcv := statscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}

	cfgJSONStr = `{
		"stats": {							
			"enabled": false,				
			"store_interval": "72h",			
			"store_uncompressed_limit": 0,	
			"thresholds_conns": ["*internal"],			
			"indexed_selects":true,			
			"prefix_indexed_fields": ["prefix_indexed_fields1","prefix_indexed_fields2"],	
			"nested_fields": false,	
		},	
		}`
	eMap = map[string]any{
		"enabled":                  false,
		"store_interval":           "72h0m0s",
		"store_uncompressed_limit": 0,
		"thresholds_conns":         []string{"*internal"},
		"indexed_selects":          true,
		"prefix_indexed_fields":    []string{"prefix_indexed_fields1", "prefix_indexed_fields2"},
		"nested_fields":            false,
		"string_indexed_fields":    []string{},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStatSCfg, err := jsnCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = statscfg.loadFromJsonCfg(jsnStatSCfg); err != nil {
		t.Error(err)
	} else if rcv := statscfg.AsMapInterface(); !reflect.DeepEqual(eMap, rcv) {
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestStatSCfgloadFromJsonCfg2(t *testing.T) {
	st := StatSCfg{}

	bl := true
	nm := 1
	dr := "1ns"
	slc := []string{"val1", "val2"}

	js := StatServJsonCfg{
		Enabled:                  &bl,
		Indexed_selects:          &bl,
		Store_interval:           &dr,
		Store_uncompressed_limit: &nm,
		Thresholds_conns:         &slc,
		String_indexed_fields:    &slc,
		Prefix_indexed_fields:    &slc,
		Nested_fields:            &bl,
	}

	exp := StatSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StoreInterval:          1 * time.Nanosecond,
		StoreUncompressedLimit: 1,
		ThresholdSConns:        slc,
		StringIndexedFields:    &slc,
		PrefixIndexedFields:    &slc,
		NestedFields:           true,
	}

	err := st.loadFromJsonCfg(&js)

	if err != nil {
		t.Errorf("didn't expect an error: %s", err)
	}

	if !reflect.DeepEqual(st, exp) {
		t.Errorf("expected %v, recived %v", exp, st)
	}

	t.Run("check error in parse duration with nanosecs", func(t *testing.T) {
		str := "test"

		js := StatServJsonCfg{
			Store_interval: &str,
		}

		err := st.loadFromJsonCfg(&js)
		exp := fmt.Errorf(`time: invalid duration "test"`)

		if err.Error() != exp.Error() {
			t.Errorf("recived %s, expected %s", err, exp)
		}
	})
}

func TestStatSCfgAsMapInterface2(t *testing.T) {

	st := StatSCfg{
		Enabled:                false,
		IndexedSelects:         false,
		StoreInterval:          1 * time.Second,
		StoreUncompressedLimit: 1,
		ThresholdSConns:        []string{"val1", "val2"},
		StringIndexedFields:    &[]string{"val1", "val2"},
		PrefixIndexedFields:    &[]string{"val1", "val2"},
		NestedFields:           true,
	}

	exp := map[string]any{
		utils.EnabledCfg:                st.Enabled,
		utils.IndexedSelectsCfg:         st.IndexedSelects,
		utils.StoreIntervalCfg:          "1s",
		utils.StoreUncompressedLimitCfg: st.StoreUncompressedLimit,
		utils.ThresholdSConnsCfg:        []string{"val1", "val2"},
		utils.StringIndexedFieldsCfg:    []string{"val1", "val2"},
		utils.PrefixIndexedFieldsCfg:    []string{"val1", "val2"},
		utils.NestedFieldsCfg:           st.NestedFields,
	}

	rcv := st.AsMapInterface()

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("recived %v, expected %v", rcv, exp)
	}
}
