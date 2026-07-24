// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"fmt"
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
		t.Errorf("Expected: %+v ,received: %+v", expected, thscfg)
	}
	if err := thscfg.loadFromJsonCfg(new(ThresholdSJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thscfg, expected) {
		t.Errorf("Expected: %+v ,received: %+v", expected, thscfg)
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
		t.Errorf("Expected: %+v , received: %+v", expected, thscfg)
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
	eMap := map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
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
	eMap = map[string]any{
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
		t.Errorf("\nExpected: %+v\nReceived: %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestThresholdCFGLoadFromJsonCFG(t *testing.T) {
	str := "test"

	to := ThresholdSCfg{}

	toj := ThresholdSJsonCfg{
		Store_interval: &str,
	}

	err := to.loadFromJsonCfg(&toj)
	exp := fmt.Errorf(`time: invalid duration "test"`)

	if err.Error() != exp.Error() {
		t.Fatalf("recived %s, expected %s", err, exp)
	}
}
