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
	"prefix_indexed_fields": ["*req.index1", "*req.index2"],			// query indexes based on these fields for faster processing
	},		
}`
	expected = ThresholdSCfg{
		StoreInterval:       time.Duration(time.Hour * 2),
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
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
	cfgJSONStr := `{
		"thresholds": {},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.StoreIntervalCfg:       "",
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.thresholdSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expextec %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestThresholdSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"thresholds": {								
			"enabled": true,						
			"store_interval": "96h",					
			"indexed_selects": false,	
			"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],	
            "suffix_indexed_fields": ["*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"],		
			"nested_fields": true,					
		},		
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StoreIntervalCfg:       "96h0m0s",
		utils.IndexedSelectsCfg:      false,
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"},
		utils.NestedFieldsCfg:        true,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.thresholdSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expextec %+v \n, recevied %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
