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

func TestResourceSConfigloadFromJsonCfg(t *testing.T) {
	cfgJSON := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{utils.MetaInternal},
		Store_interval:        utils.StringPointer("2s"),
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &ResourceSConfig{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       time.Duration(2 * time.Second),
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	if cfgJson, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = cfgJson.resourceSCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfgJson.resourceSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfgJson.resourceSCfg))
	}
}

func TestResourceSConfigAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"resources": {},	
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             false,
		utils.StoreIntervalCfg:       utils.EmptyString,
		utils.ThresholdSConnsCfg:     []string{},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{},
		utils.SuffixIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:        false,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestResourceSConfigAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"resources": {								
			"enabled": true,						
			"store_interval": "7m",					
			"thresholds_conns": ["*internal"],					
			"indexed_selects":true,					
			"prefix_indexed_fields": ["*req.prefix_indexed_fields1","*req.prefix_indexed_fields2"],
            "suffix_indexed_fields": ["*req.prefix_indexed_fields1"],
			"nested_fields": true,					
		},	
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StoreIntervalCfg:       "7m0s",
		utils.ThresholdSConnsCfg:     []string{"*internal"},
		utils.IndexedSelectsCfg:      true,
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1", "*req.prefix_indexed_fields2"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1"},
		utils.NestedFieldsCfg:        true,
	}
	if cgrCfg, err := NewCGRConfigFromJsonStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
