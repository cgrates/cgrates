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
		ExistsIndexedFields:      &[]string{"*req.index1", "*req.index2"},
		Nested_fields:            utils.BoolPointer(true),
		Ees_conns:                &[]string{utils.MetaInternal, "*conn1"},
		Ees_exporter_ids:         &[]string{"exporterID"},
	}
	expected := &StatSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StoreInterval:          2,
		StoreUncompressedLimit: 10,
		ThresholdSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StringIndexedFields:    &[]string{"*req.string"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		ExistsIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NestedFields:           true,
		Opts: &StatsOpts{
			ProfileIDs: []string{},
		},
		EEsConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), "*conn1"},
		EEsExporterIDs: []string{"exporterID"},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.statsCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.statsCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.statsCfg))
	}
	jsonCfg.statsCfg.Opts.loadFromJSONCfg(nil)
	if reflect.DeepEqual(nil, jsonCfg.statsCfg.Opts) {

		t.Error("expected nil")
	}
}

func TestStatSCfgloadFromJsonCfgCase2(t *testing.T) {
	statscfgJSON := &StatServJsonCfg{
		Store_interval: utils.StringPointer("1ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"1ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.statsCfg.loadFromJSONCfg(statscfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestStatSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
		"stats": {},	
}`
	eMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.StoreUncompressedLimitCfg: 0,
		utils.ThresholdSConnsCfg:        []string{},
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.ExistsIndexedFieldsCfg:    []string{},
		utils.NestedFieldsCfg:           false,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs:              []string{},
			utils.MetaProfileIgnoreFiltersCfg: false,
		},
		utils.EEsConnsCfg:       []string{},
		utils.EEsExporterIDsCfg: []string{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
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
            "exists_indexed_fields":["*req.exists_indexed_field"],
			"nested_fields": true,
			"ees_conns": ["*internal:*ees", "*conn1"],
			"ees_exporter_ids":["exporterID"],
		},	
}`
	eMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.StoreIntervalCfg:          "72h0m0s",
		utils.StoreUncompressedLimitCfg: 1,
		utils.ThresholdSConnsCfg:        []string{utils.MetaInternal, "*conn1"},
		utils.IndexedSelectsCfg:         false,
		utils.StringIndexedFieldsCfg:    []string{"*req.string"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.prefix_indexed_fields1", "*req.prefix_indexed_fields2"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.suffix_indexed_fields"},
		utils.ExistsIndexedFieldsCfg:    []string{"*req.exists_indexed_field"},
		utils.NestedFieldsCfg:           true,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs:              []string{},
			utils.MetaProfileIgnoreFiltersCfg: false,
		},
		utils.EEsConnsCfg:       []string{utils.MetaInternal, "*conn1"},
		utils.EEsExporterIDsCfg: []string{"exporterID"},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.statsCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}
func TestStatSCfgClone(t *testing.T) {
	ban := &StatSCfg{
		Enabled:                true,
		IndexedSelects:         true,
		StoreInterval:          2,
		StoreUncompressedLimit: 10,
		ThresholdSConns:        []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		SuffixIndexedFields:    &[]string{"*req.index1", "*req.index2"},
		NestedFields:           true,
		Opts: &StatsOpts{
			ProfileIDs: []string{},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
	}
	if rcv.ThresholdSConns[1] = ""; ban.ThresholdSConns[1] != "*conn1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.StringIndexedFields)[0] = ""; (*ban.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.PrefixIndexedFields)[0] = ""; (*ban.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	if (*rcv.SuffixIndexedFields)[0] = ""; (*ban.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}
