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

func TestResourceSConfigloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index1"},
		Prefix_indexed_fields: &[]string{"*req.index1"},
		Suffix_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:         utils.BoolPointer(true),
	}
	expected := &ResourceSConfig{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2 * time.Second,
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
	}
	cfg := NewDefaultCGRConfig()
	if err = cfg.resourceSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.resourceSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.resourceSCfg))
	}
}

func TestResourceSConfigloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &ResourceSJsonCfg{
		Store_interval: utils.StringPointer("2ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"2ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err = jsonCfg.resourceSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
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
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestResourceSConfigAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"resources": {								
			"enabled": true,						
			"store_interval": "7m",					
			"thresholds_conns": ["*internal:*thresholds", "*conn1"],					
			"indexed_selects":true,		
            "string_indexed_fields": ["*req.index1"],
			"prefix_indexed_fields": ["*req.prefix_indexed_fields1","*req.prefix_indexed_fields2"],
            "suffix_indexed_fields": ["*req.prefix_indexed_fields1"],
			"nested_fields": true,					
		},	
	}`
	eMap := map[string]interface{}{
		utils.EnabledCfg:             true,
		utils.StoreIntervalCfg:       "7m0s",
		utils.ThresholdSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
		utils.IndexedSelectsCfg:      true,
		utils.StringIndexedFieldsCfg: []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1", "*req.prefix_indexed_fields2"},
		utils.SuffixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1"},
		utils.NestedFieldsCfg:        true,
		utils.OptsCfg:                map[string]interface{}{},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestResourceSConfigClone(t *testing.T) {
	ban := &ResourceSConfig{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2 * time.Second,
		ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		Opts:                &ResourcesOpts{},
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

func TestDiffResourceSJsonCfg(t *testing.T) {
	var d *ResourceSJsonCfg

	v1 := &ResourceSConfig{
		Enabled:             false,
		IndexedSelects:      false,
		ThresholdSConns:     []string{"*localhost"},
		StoreInterval:       1 * time.Second,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Opts: &ResourcesOpts{
			UsageID:  "usg1",
			UsageTTL: utils.DurationPointer(time.Second),
			Units:    1,
		},
	}

	v2 := &ResourceSConfig{
		Enabled:             true,
		IndexedSelects:      true,
		ThresholdSConns:     []string{"*birpc"},
		StoreInterval:       2 * time.Second,
		StringIndexedFields: &[]string{"*req.index11"},
		PrefixIndexedFields: &[]string{"*req.index22"},
		SuffixIndexedFields: &[]string{"*req.index33"},
		NestedFields:        true,
		Opts: &ResourcesOpts{
			UsageID:  "usg2",
			UsageTTL: utils.DurationPointer(time.Minute),
			Units:    2,
		},
	}

	expected := &ResourceSJsonCfg{
		Enabled:               utils.BoolPointer(true),
		Indexed_selects:       utils.BoolPointer(true),
		Thresholds_conns:      &[]string{"*birpc"},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &ResourcesOptsJson{
			UsageID:  utils.StringPointer("usg2"),
			UsageTTL: utils.StringPointer("1m0s"),
			Units:    utils.Float64Pointer(2),
		},
	}

	rcv := diffResourceSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &ResourceSJsonCfg{
		Opts: &ResourcesOptsJson{},
	}
	rcv = diffResourceSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}
