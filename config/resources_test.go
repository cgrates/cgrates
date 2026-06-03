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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestResourceSCfgloadFromJsonCfgCase1(t *testing.T) {
	cfgJSON := &ResourceSJsonCfg{
		Enabled:         utils.BoolPointer(true),
		Indexed_selects: utils.BoolPointer(true),
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
		},
		Store_interval:           utils.StringPointer("2s"),
		String_indexed_fields:    &[]string{"*req.index1"},
		Prefix_indexed_fields:    &[]string{"*req.index1"},
		Suffix_indexed_fields:    &[]string{"*req.index1"},
		Exists_indexed_fields:    &[]string{"*req.index1"},
		Notexists_indexed_fields: &[]string{"*req.index1"},
		Nested_fields:            utils.BoolPointer(true),
	}
	expected := &ResourceSCfg{
		Enabled:        true,
		IndexedSelects: true,
		StoreInterval:  2 * time.Second,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
		},
		StringIndexedFields:    &[]string{"*req.index1"},
		PrefixIndexedFields:    &[]string{"*req.index1"},
		SuffixIndexedFields:    &[]string{"*req.index1"},
		ExistsIndexedFields:    &[]string{"*req.index1"},
		NotExistsIndexedFields: &[]string{"*req.index1"},
		NestedFields:           true,
		Opts: &ResourcesOpts{
			UsageID:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			UsageTTL: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			Units:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	cfg := NewDefaultCGRConfig()
	if err := cfg.resourceSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, cfg.resourceSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(cfg.resourceSCfg))
	}
	cfgJSON = nil
	if err := cfg.resourceSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}
}

func TestResourceSLoadFromJSONOpts(t *testing.T) {
	resOpts := &ResourcesOpts{
		UsageID: []*DynamicStringOpt{
			{
				value: utils.EmptyString,
			},
		},
		UsageTTL: []*DynamicDurationOpt{
			{
				value: 72 * time.Hour,
			},
		},
		Units: []*DynamicFloat64Opt{
			{
				value: 1,
			},
		},
	}

	resOptsJson := &ResourcesOptsJson{
		UsageID: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.org",
				Value:  "usg2",
			},
		},
		UsageTTL: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.org",
				Value:  "error",
			},
		},
		Units: []*DynamicInterfaceOpt{
			{
				Tenant: "cgrates.org",
				Value:  "2.5",
			},
		},
	}
	errExp := `time: invalid duration "error"`
	if err := resOpts.loadFromJSONCfg(resOptsJson); err == nil || err.Error() != errExp {
		t.Errorf("Expected %v \n but received \n %v", errExp, err.Error())
	}

	if err := resOpts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}
}

func TestResourceSCfgloadFromJsonCfgCase2(t *testing.T) {
	cfgJSON := &ResourceSJsonCfg{
		Store_interval: utils.StringPointer("2ss"),
	}
	expected := "time: unknown unit \"ss\" in duration \"2ss\""
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.resourceSCfg.loadFromJSONCfg(cfgJSON); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func TestResourceSCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
	"resources": {},	
}`
	eMap := map[string]any{
		utils.EnabledCfg:                false,
		utils.StoreIntervalCfg:          utils.EmptyString,
		utils.ConnsCfg:                  map[string][]*DynamicConns{},
		utils.IndexedSelectsCfg:         true,
		utils.PrefixIndexedFieldsCfg:    []string{},
		utils.SuffixIndexedFieldsCfg:    []string{},
		utils.ExistsIndexedFieldsCfg:    []string{},
		utils.NotExistsIndexedFieldsCfg: []string{},
		utils.NestedFieldsCfg:           false,
		utils.OptsCfg: map[string]any{
			utils.MetaUsageIDCfg:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			utils.MetaUsageTTLCfg: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			utils.MetaUnitsCfg:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestResourceSCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
		"resources": {								
			"enabled": true,						
			"storeInterval": "7m",					
			"conns": {"*thresholds": [{"connIDs": ["*internal", "*conn1"]}]},					
			"indexedSelects":true,		
            "stringIndexedFields": ["*req.index1"],
			"prefixIndexedFields": ["*req.prefixIndexedFields1","*req.prefixIndexedFields2"],
            "suffixIndexedFields": ["*req.prefixIndexedFields1"],
			"existsIndexedFields": ["*req.prefixIndexedFields1","*req.prefixIndexedFields2"],
            "notExistsIndexedFields": ["*req.prefixIndexedFields1"],
			"nestedFields": true,					
		},	
	}`
	eMap := map[string]any{
		utils.EnabledCfg:       true,
		utils.StoreIntervalCfg: "7m0s",
		utils.ConnsCfg: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{utils.MetaInternal, "*conn1"}}},
		},
		utils.IndexedSelectsCfg:         true,
		utils.StringIndexedFieldsCfg:    []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg:    []string{"*req.prefixIndexedFields1", "*req.prefixIndexedFields2"},
		utils.SuffixIndexedFieldsCfg:    []string{"*req.prefixIndexedFields1"},
		utils.ExistsIndexedFieldsCfg:    []string{"*req.prefixIndexedFields1", "*req.prefixIndexedFields2"},
		utils.NotExistsIndexedFieldsCfg: []string{"*req.prefixIndexedFields1"},
		utils.NestedFieldsCfg:           true,
		utils.OptsCfg: map[string]any{
			utils.MetaUsageIDCfg:  []*DynamicStringOpt{{value: ResourcesUsageIDDftOpt}},
			utils.MetaUsageTTLCfg: []*DynamicDurationOpt{{value: ResourcesUsageTTLDftOpt}},
			utils.MetaUnitsCfg:    []*DynamicFloat64Opt{{value: ResourcesUnitsDftOpt}},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestResourceSCfgClone(t *testing.T) {
	ban := &ResourceSCfg{
		Enabled:        true,
		IndexedSelects: true,
		StoreInterval:  2 * time.Second,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"}}},
		},
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
	if rcv.Conns[utils.MetaThresholds][0].ConnIDs[1] = ""; ban.Conns[utils.MetaThresholds][0].ConnIDs[1] != "*conn1" {
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

	v1 := &ResourceSCfg{
		Enabled:        false,
		IndexedSelects: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
		},
		StoreInterval:       1 * time.Second,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Opts: &ResourcesOpts{
			UsageID: []*DynamicStringOpt{
				{
					value: "usg1",
				},
			},
			UsageTTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			Units: []*DynamicFloat64Opt{
				{
					value: 1,
				},
			},
		},
	}

	v2 := &ResourceSCfg{
		Enabled:        true,
		IndexedSelects: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{"*birpc"}}},
		},
		StoreInterval:       2 * time.Second,
		StringIndexedFields: &[]string{"*req.index11"},
		PrefixIndexedFields: &[]string{"*req.index22"},
		SuffixIndexedFields: &[]string{"*req.index33"},
		NestedFields:        true,
		Opts: &ResourcesOpts{
			UsageID: []*DynamicStringOpt{
				{
					value: "usg2",
				},
			},
			UsageTTL: []*DynamicDurationOpt{
				{
					value: time.Minute,
				},
			},
			Units: []*DynamicFloat64Opt{
				{
					value: 2,
				},
			},
		},
	}

	expected := &ResourceSJsonCfg{
		Enabled:         utils.BoolPointer(true),
		Indexed_selects: utils.BoolPointer(true),
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{"*birpc"}}},
		},
		Store_interval:        utils.StringPointer("2s"),
		String_indexed_fields: &[]string{"*req.index11"},
		Prefix_indexed_fields: &[]string{"*req.index22"},
		Suffix_indexed_fields: &[]string{"*req.index33"},
		Nested_fields:         utils.BoolPointer(true),
		Opts: &ResourcesOptsJson{
			UsageID: []*DynamicInterfaceOpt{
				{
					Value: "usg2",
				},
			},
			UsageTTL: []*DynamicInterfaceOpt{
				{
					Value: time.Minute,
				},
			},
			Units: []*DynamicInterfaceOpt{
				{
					Value: float64(2),
				},
			},
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

func TestResourcesCloneSection(t *testing.T) {
	rsrCfg := &ResourceSCfg{
		Enabled:        false,
		IndexedSelects: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
		},
		StoreInterval:       1 * time.Second,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Opts: &ResourcesOpts{
			UsageID: []*DynamicStringOpt{
				{
					value: "usg1",
				},
			},
			UsageTTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			Units: []*DynamicFloat64Opt{
				{
					value: 1,
				},
			},
		},
	}

	exp := &ResourceSCfg{
		Enabled:        false,
		IndexedSelects: false,
		Conns: map[string][]*DynamicConns{
			utils.MetaThresholds: {{ConnIDs: []string{"*localhost"}}},
		},
		StoreInterval:       1 * time.Second,
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index2"},
		SuffixIndexedFields: &[]string{"*req.index3"},
		NestedFields:        false,
		Opts: &ResourcesOpts{
			UsageID: []*DynamicStringOpt{
				{
					value: "usg1",
				},
			},
			UsageTTL: []*DynamicDurationOpt{
				{
					value: time.Second,
				},
			},
			Units: []*DynamicFloat64Opt{
				{
					value: 1,
				},
			},
		},
	}

	rcv := rsrCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}
