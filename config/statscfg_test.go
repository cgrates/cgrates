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

func TestStatSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &StatServJsonCfg{
		Enabled:                  utils.BoolPointer(true),
		Indexed_selects:          utils.BoolPointer(true),
		Store_interval:           utils.StringPointer("2"),
		Store_uncompressed_limit: utils.IntPointer(10),
		Thresholds_conns:         &[]string{utils.MetaInternal, "*conn1"},
		String_indexed_fields:    &[]string{"*req.string"},
		Prefix_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Suffix_indexed_fields:    &[]string{"*req.index1", "*req.index2"},
		Nested_fields:            utils.BoolPointer(true),
	}
	expected := &StatSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StoreInterval:          time.Duration(2),
		StoreUncompressedLimit: 10,
		ThresholdSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StringIndexedFields:    &[]string{"*req.string"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NestedFields:           true,
	}
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.statsCfg.loadFromJsonCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.statsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.statsCfg))
	}
}

func TestStatSCfgloadFromJsonCfgCase2(t *testing.T) {
	statscfgJson := &StatServJsonCfg{
		Store_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	if jsonCfg, err := NewDefaultCGRConfig(); err != nil {
		t.Error(err)
	} else if err = jsonCfg.statsCfg.loadFromJsonCfg(statscfgJson); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
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
			"thresholds_conns": ["*internal:*thresholds", "*conn1"],			
			"indexed_selects":false,			
            "string_indexed_fields": ["*req.string"],
			"prefix_indexed_fields": ["*req.prefix_indexed_fields1","*req.prefix_indexed_fields2"],
            "suffix_indexed_fields":["*req.suffix_indexed_fields"],
			"nested_fields": true,	
		},	
}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:                true,
		utils.StoreIntervalCfg:          "72h0m0s",
		utils.StoreUncompressedLimitCfg: 1,
		utils.ThresholdSConnsCfg:        []string{utils.MetaInternal, "*conn1"},
		utils.IndexedSelectsCfg:         false,
		utils.StringIndexedFieldsCfg:    []string{"*req.string"},
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
