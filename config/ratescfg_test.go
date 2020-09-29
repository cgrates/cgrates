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

func TestRateSConfigloadFromJsonCfg(t *testing.T) {
	cfgJSON := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Indexed_selects:            utils.BoolPointer(true),
		Prefix_indexed_fields:      &[]string{"*req.index1"},
		Suffix_indexed_fields:      &[]string{"*req.index1"},
		Nested_fields:              utils.BoolPointer(true),
		Rate_indexed_selects:       utils.BoolPointer(true),
		Rate_prefix_indexed_fields: &[]string{"*req.index1"},
		Rate_suffix_indexed_fields: &[]string{"*req.index1"},
		Rate_nested_fields:         utils.BoolPointer(true),
	}
	expected := &RateSCfg{
		Enabled:                 true,
		IndexedSelects:          true,
		PrefixIndexedFields:     &[]string{"*req.index1"},
		SuffixIndexedFields:     &[]string{"*req.index1"},
		NestedFields:            true,
		RateIndexedSelects:      true,
		RatePrefixIndexedFields: &[]string{"*req.index1"},
		RateSuffixIndexedFields: &[]string{"*req.index1"},
		RateNestedFields:        true,
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.rateSCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.rateSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rateSCfg))
	}
}

func TestRatesCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "rates": {}
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 false,
		utils.IndexedSelectsCfg:          true,
		utils.PrefixIndexedFieldsCfg:     []string{},
		utils.SuffixIndexedFieldsCfg:     []string{},
		utils.NestedFieldsCfg:            false,
		utils.RateIndexedSelectsCfg:      true,
		utils.RatePrefixIndexedFieldsCfg: []string{},
		utils.RateSuffixIndexedFieldsCfg: []string{},
		utils.RateNestedFieldsCfg:        false,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rateSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}

func TestRatesCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
     "rates": {
	        "enabled": true,
	        "indexed_selects": false,				
	         //"string_indexed_fields": [],			
            "prefix_indexed_fields": ["*req.index1", "*req.index2"],			
            "suffix_indexed_fields": ["*req.index1"],			
	        "nested_fields": true,					
	        "rate_indexed_selects": false,			
	        //"rate_string_indexed_fields": [],		
	         "rate_prefix_indexed_fields": ["*req.index1", "*req.index2"],		
         	"rate_suffix_indexed_fields": ["*req.index1", "*req.index2", "*req.index3"],		
	        "rate_nested_fields": true,			
     },
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                 true,
		utils.IndexedSelectsCfg:          false,
		utils.PrefixIndexedFieldsCfg:     []string{"*req.index1", "*req.index2"},
		utils.SuffixIndexedFieldsCfg:     []string{"*req.index1"},
		utils.NestedFieldsCfg:            true,
		utils.RateIndexedSelectsCfg:      false,
		utils.RatePrefixIndexedFieldsCfg: []string{"*req.index1", "*req.index2"},
		utils.RateSuffixIndexedFieldsCfg: []string{"*req.index1", "*req.index2", "*req.index3"},
		utils.RateNestedFieldsCfg:        true,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rateSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", eMap, rcv)
	}
}
