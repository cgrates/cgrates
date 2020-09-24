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

func TestStatSCfgloadFromJsonCfg(t *testing.T) {
	var statscfg, expected StatSCfg
	if err := statscfg.loadFromJsonCfg(nil); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, statscfg)
	}
	if err := statscfg.loadFromJsonCfg(new(StatServJsonCfg)); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statscfg, expected) {
		t.Errorf("Expected: %+v ,recived: %+v", expected, statscfg)
	}
	cfgJSONStr := `{
"stats": {									// Stat service (*new)
	"enabled": false,						// starts Stat service: <true|false>.
	"store_interval": "2s",					// dump cache regularly to dataDB, 0 - dump at start/shutdown: <""|$dur>
	"thresholds_conns": [],					// address where to reach the thresholds service, empty to disable thresholds functionality: <""|*internal|x.y.z.y:1234>
	//"string_indexed_fields": [],			// query indexes based on these fields for faster processing
	"prefix_indexed_fields": ["*req.index1", "*req.index2"],			// query indexes based on these fields for faster processing
},	
}`
	expected = StatSCfg{
		StoreInterval:       time.Duration(time.Second * 2),
		ThresholdSConns:     []string{},
		PrefixIndexedFields: &[]string{"*req.index1", "*req.index2"},
	}
	if jsnCfg, err := NewCgrJsonCfgFromBytes([]byte(cfgJSONStr)); err != nil {
		t.Error(err)
	} else if jsnStatSCfg, err := jsnCfg.StatSJsonCfg(); err != nil {
		t.Error(err)
	} else if err = statscfg.loadFromJsonCfg(jsnStatSCfg); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, statscfg) {
		t.Errorf("Expected: %+v , recived: %+v", expected, statscfg)
	}
}

func TestStatSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"stats": {},	
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                false,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.StoreUncompressedLimitCfg: 0,
		utils.ThresholdSConnsCfg:        []string{},
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.NestedFieldsCfg:           false,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.statsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestStatSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"stats": {							
			"enabled": true,				
			"store_interval": "72h",			
			"store_uncompressed_limit": 1,	
			"thresholds_conns": ["*internal"],			
			"indexed_selects":false,			
			"prefix_indexed_fields": ["*req.prefix_indexed_fields1","*req.prefix_indexed_fields2"],
            "suffix_indexed_fields":["*req.suffix_indexed_fields"],
			"nested_fields": true,	
		},	
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                true,
		utils.StoreIntervalCfg:          "72h0m0s",
		utils.StoreUncompressedLimitCfg: 1,
		utils.ThresholdSConnsCfg:        []string{"*internal"},
		utils.IndexedSelectsCfg:         false,
		utils.PrefixIndexedFieldsCfg:    []string{"*req.prefix_indexed_fields1", "*req.prefix_indexed_fields2"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.suffix_indexed_fields"},
		utils.NestedFieldsCfg:           true,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.statsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}
