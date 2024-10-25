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
	"github.com/ericlagergren/decimal"
)

func TestRateSConfigloadFromJsonCfg(t *testing.T) {
	cfgJSON := &RateSJsonCfg{
		Enabled:                       utils.BoolPointer(true),
		Indexed_selects:               utils.BoolPointer(true),
		String_indexed_fields:         &[]string{"*req.index1"},
		Prefix_indexed_fields:         &[]string{"*req.index1"},
		Suffix_indexed_fields:         &[]string{"*req.index1"},
		Exists_indexed_fields:         &[]string{"*req.index1"},
		Notexists_indexed_fields:      &[]string{"*req.index1"},
		Nested_fields:                 utils.BoolPointer(true),
		Rate_indexed_selects:          utils.BoolPointer(true),
		Rate_string_indexed_fields:    &[]string{"*req.index1"},
		Rate_prefix_indexed_fields:    &[]string{"*req.index1"},
		Rate_suffix_indexed_fields:    &[]string{"*req.index1"},
		Rate_exists_indexed_fields:    &[]string{"*req.index1"},
		Rate_notexists_indexed_fields: &[]string{"*req.index1"},
		Rate_nested_fields:            utils.BoolPointer(true),
		Verbosity:                     utils.IntPointer(20),
	}
	expected := &RateSCfg{
		Enabled:                    true,
		IndexedSelects:             true,
		StringIndexedFields:        &[]string{"*req.index1"},
		PrefixIndexedFields:        &[]string{"*req.index1"},
		SuffixIndexedFields:        &[]string{"*req.index1"},
		ExistsIndexedFields:        &[]string{"*req.index1"},
		NotExistsIndexedFields:     &[]string{"*req.index1"},
		NestedFields:               true,
		RateIndexedSelects:         true,
		RateStringIndexedFields:    &[]string{"*req.index1"},
		RatePrefixIndexedFields:    &[]string{"*req.index1"},
		RateSuffixIndexedFields:    &[]string{"*req.index1"},
		RateExistsIndexedFields:    &[]string{"*req.index1"},
		RateNotExistsIndexedFields: &[]string{"*req.index1"},
		RateNestedFields:           true,
		Verbosity:                  20,
		Opts: &RatesOpts{
			ProfileIDs:           []*utils.DynamicStringSliceOpt{},
			StartTime:            []*utils.DynamicStringOpt{},
			Usage:                []*utils.DynamicDecimalBigOpt{},
			IntervalStart:        []*utils.DynamicDecimalBigOpt{},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	jsonCfg := NewDefaultCGRConfig()
	if err := jsonCfg.rateSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, jsonCfg.rateSCfg) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(jsonCfg.rateSCfg))
	}
	cfgJSON = nil
	if err := jsonCfg.rateSCfg.loadFromJSONCfg(cfgJSON); err != nil {
		t.Error(err)
	}
}

func TestRatesCfgAsMapInterface(t *testing.T) {
	cfgJSONStr := `{
      "rates": {}
}`
	eMap := map[string]any{
		utils.EnabledCfg:                    false,
		utils.IndexedSelectsCfg:             true,
		utils.PrefixIndexedFieldsCfg:        []string{},
		utils.SuffixIndexedFieldsCfg:        []string{},
		utils.ExistsIndexedFieldsCfg:        []string{},
		utils.NotExistsIndexedFieldsCfg:     []string{},
		utils.NestedFieldsCfg:               false,
		utils.RateIndexedSelectsCfg:         true,
		utils.RatePrefixIndexedFieldsCfg:    []string{},
		utils.RateSuffixIndexedFieldsCfg:    []string{},
		utils.RateExistsIndexedFieldsCfg:    []string{},
		utils.RateNotExistsIndexedFieldsCfg: []string{},
		utils.RateNestedFieldsCfg:           false,
		utils.Verbosity:                     1000,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs:           []*utils.DynamicStringSliceOpt{},
			utils.MetaStartTime:            []*utils.DynamicStringOpt{},
			utils.MetaUsage:                []*utils.DynamicDecimalBigOpt{},
			utils.MetaIntervalStartCfg:     []*utils.DynamicDecimalBigOpt{},
			utils.MetaProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rateSCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRatesCfgAsMapInterface1(t *testing.T) {
	cfgJSONStr := `{
     "rates": {
	        "enabled": true,
	        "indexed_selects": false,				
	        "string_indexed_fields": ["*req.index1"],			
            "prefix_indexed_fields": ["*req.index1", "*req.index2"],			
            "suffix_indexed_fields": ["*req.index1"],
			"exists_indexed_fields": ["*req.index1", "*req.index2"],			
            "notexists_indexed_fields": ["*req.index1"],			
	        "nested_fields": true,					
	        "rate_indexed_selects": false,			
	        "rate_string_indexed_fields": ["*req.index1"],		
	        "rate_prefix_indexed_fields": ["*req.index1", "*req.index2"],		
         	"rate_suffix_indexed_fields": ["*req.index1", "*req.index2", "*req.index3"],	
			"rate_exists_indexed_fields": ["*req.index1", "*req.index2"],			
            "rate_notexists_indexed_fields": ["*req.index1"],	
	        "rate_nested_fields": true,			
     },
}`
	eMap := map[string]any{
		utils.EnabledCfg:                    true,
		utils.IndexedSelectsCfg:             false,
		utils.StringIndexedFieldsCfg:        []string{"*req.index1"},
		utils.PrefixIndexedFieldsCfg:        []string{"*req.index1", "*req.index2"},
		utils.SuffixIndexedFieldsCfg:        []string{"*req.index1"},
		utils.ExistsIndexedFieldsCfg:        []string{"*req.index1", "*req.index2"},
		utils.NotExistsIndexedFieldsCfg:     []string{"*req.index1"},
		utils.NestedFieldsCfg:               true,
		utils.RateIndexedSelectsCfg:         false,
		utils.RateStringIndexedFieldsCfg:    []string{"*req.index1"},
		utils.RatePrefixIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.RateSuffixIndexedFieldsCfg:    []string{"*req.index1", "*req.index2", "*req.index3"},
		utils.RateExistsIndexedFieldsCfg:    []string{"*req.index1", "*req.index2"},
		utils.RateNotExistsIndexedFieldsCfg: []string{"*req.index1"},
		utils.RateNestedFieldsCfg:           true,
		utils.Verbosity:                     1000,
		utils.OptsCfg: map[string]any{
			utils.MetaProfileIDs:           []*utils.DynamicStringSliceOpt{},
			utils.MetaStartTime:            []*utils.DynamicStringOpt{},
			utils.MetaUsage:                []*utils.DynamicDecimalBigOpt{},
			utils.MetaIntervalStartCfg:     []*utils.DynamicDecimalBigOpt{},
			utils.MetaProfileIgnoreFilters: []*utils.DynamicBoolOpt{},
		},
	}
	if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(cfgJSONStr); err != nil {
		t.Error(err)
	} else if rcv := cgrCfg.rateSCfg.AsMapInterface(""); !reflect.DeepEqual(rcv, eMap) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(eMap), utils.ToJSON(rcv))
	}
}

func TestRateSCfgClone(t *testing.T) {
	sa := &RateSCfg{
		Enabled:                 true,
		IndexedSelects:          true,
		StringIndexedFields:     &[]string{"*req.index1"},
		PrefixIndexedFields:     &[]string{"*req.index1"},
		SuffixIndexedFields:     &[]string{"*req.index1"},
		NestedFields:            true,
		RateIndexedSelects:      true,
		RateStringIndexedFields: &[]string{"*req.index1"},
		RatePrefixIndexedFields: &[]string{"*req.index1"},
		RateSuffixIndexedFields: &[]string{"*req.index1"},
		RateNestedFields:        true,
		Verbosity:               20,
		Opts:                    &RatesOpts{},
	}
	rcv := sa.Clone()
	if !reflect.DeepEqual(sa, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(sa), utils.ToJSON(rcv))
	}
	(*rcv.StringIndexedFields)[0] = ""
	if (*sa.StringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	(*rcv.PrefixIndexedFields)[0] = ""
	if (*sa.PrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	(*rcv.SuffixIndexedFields)[0] = ""
	if (*sa.SuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}

	(*rcv.RateStringIndexedFields)[0] = ""
	if (*sa.RateStringIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	(*rcv.RatePrefixIndexedFields)[0] = ""
	if (*sa.RatePrefixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
	(*rcv.RateSuffixIndexedFields)[0] = ""
	if (*sa.RateSuffixIndexedFields)[0] != "*req.index1" {
		t.Errorf("Expected clone to not modify the cloned")
	}
}

func TestDiffRateSJsonCfg(t *testing.T) {
	var d *RateSJsonCfg

	v1 := &RateSCfg{
		Enabled:                 false,
		IndexedSelects:          false,
		StringIndexedFields:     &[]string{"*req.index1"},
		PrefixIndexedFields:     &[]string{"*req.index2"},
		SuffixIndexedFields:     &[]string{"*req.index3"},
		NestedFields:            false,
		RateIndexedSelects:      false,
		RateStringIndexedFields: &[]string{"*req.rateIndex1"},
		RatePrefixIndexedFields: &[]string{"*req.rateIndex2"},
		RateSuffixIndexedFields: &[]string{"*req.rateIndex3"},
		RateNestedFields:        false,
		Verbosity:               2,
		Opts: &RatesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.org",
					Value:  []string{"RP1"},
				},
			},
			StartTime: []*utils.DynamicStringOpt{
				{
					Tenant: "cgrates.org",
					Value:  "",
				},
			},
			Usage: []*utils.DynamicDecimalBigOpt{
				{
					Tenant: "cgrates.org",
					Value:  decimal.WithContext(utils.DecimalContext).SetUint64(2),
				},
			},
			IntervalStart: []*utils.DynamicDecimalBigOpt{
				{
					Tenant: "cgrates.org",
					Value:  decimal.WithContext(utils.DecimalContext).SetUint64(2),
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.org",
					Value:  false,
				},
			},
		},
	}

	v2 := &RateSCfg{
		Enabled:                 true,
		IndexedSelects:          true,
		StringIndexedFields:     &[]string{"*req.index11"},
		PrefixIndexedFields:     &[]string{"*req.index22"},
		SuffixIndexedFields:     &[]string{"*req.index33"},
		NestedFields:            true,
		RateIndexedSelects:      true,
		RateStringIndexedFields: &[]string{"*req.rateIndex11"},
		RatePrefixIndexedFields: &[]string{"*req.rateIndex22"},
		RateSuffixIndexedFields: &[]string{"*req.rateIndex33"},
		RateNestedFields:        true,
		Verbosity:               3,
		Opts: &RatesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Value:  []string{"RP2"},
				},
			},
			StartTime: []*utils.DynamicStringOpt{
				{
					Tenant: "cgrates.net",
					Value:  utils.MetaNow,
				},
			},
			Usage: []*utils.DynamicDecimalBigOpt{
				{
					Tenant: "cgrates.net",
					Value:  decimal.WithContext(utils.DecimalContext).SetUint64(3),
				},
			},
			IntervalStart: []*utils.DynamicDecimalBigOpt{
				{
					Tenant: "cgrates.net",
					Value:  decimal.WithContext(utils.DecimalContext).SetUint64(3),
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
		},
	}

	expected := &RateSJsonCfg{
		Enabled:                    utils.BoolPointer(true),
		Indexed_selects:            utils.BoolPointer(true),
		String_indexed_fields:      &[]string{"*req.index11"},
		Prefix_indexed_fields:      &[]string{"*req.index22"},
		Suffix_indexed_fields:      &[]string{"*req.index33"},
		Nested_fields:              utils.BoolPointer(true),
		Rate_indexed_selects:       utils.BoolPointer(true),
		Rate_string_indexed_fields: &[]string{"*req.rateIndex11"},
		Rate_prefix_indexed_fields: &[]string{"*req.rateIndex22"},
		Rate_suffix_indexed_fields: &[]string{"*req.rateIndex33"},
		Rate_nested_fields:         utils.BoolPointer(true),
		Verbosity:                  utils.IntPointer(3),
		Opts: &RatesOptsJson{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Tenant: "cgrates.net",
					Value:  []string{"RP2"},
				},
			},
			StartTime: []*utils.DynamicStringOpt{
				{
					Tenant: "cgrates.net",
					Value:  utils.MetaNow,
				},
			},
			Usage: []*utils.DynamicStringOpt{
				{
					Tenant: "cgrates.net",
					Value:  "3",
				},
			},
			IntervalStart: []*utils.DynamicStringOpt{
				{
					Tenant: "cgrates.net",
					Value:  "3",
				},
			},
			ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
				{
					Tenant: "cgrates.net",
					Value:  true,
				},
			},
		},
	}

	rcv := diffRateSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}

	v1 = v2
	expected = &RateSJsonCfg{
		Opts: &RatesOptsJson{},
	}
	rcv = diffRateSJsonCfg(d, v1, v2)
	if !reflect.DeepEqual(rcv, expected) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestRateSCloneSection(t *testing.T) {
	rateSCfg := &RateSCfg{
		Enabled:                 false,
		IndexedSelects:          false,
		StringIndexedFields:     &[]string{"*req.index1"},
		PrefixIndexedFields:     &[]string{"*req.index2"},
		SuffixIndexedFields:     &[]string{"*req.index3"},
		NestedFields:            false,
		RateIndexedSelects:      false,
		RateStringIndexedFields: &[]string{"*req.rateIndex1"},
		RatePrefixIndexedFields: &[]string{"*req.rateIndex2"},
		RateSuffixIndexedFields: &[]string{"*req.rateIndex3"},
		RateNestedFields:        false,
		Verbosity:               2,
		Opts: &RatesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"RP1"},
				},
			},
		},
	}

	exp := &RateSCfg{
		Enabled:                 false,
		IndexedSelects:          false,
		StringIndexedFields:     &[]string{"*req.index1"},
		PrefixIndexedFields:     &[]string{"*req.index2"},
		SuffixIndexedFields:     &[]string{"*req.index3"},
		NestedFields:            false,
		RateIndexedSelects:      false,
		RateStringIndexedFields: &[]string{"*req.rateIndex1"},
		RatePrefixIndexedFields: &[]string{"*req.rateIndex2"},
		RateSuffixIndexedFields: &[]string{"*req.rateIndex3"},
		RateNestedFields:        false,
		Verbosity:               2,
		Opts: &RatesOpts{
			ProfileIDs: []*utils.DynamicStringSliceOpt{
				{
					Value: []string{"RP1"},
				},
			},
		},
	}

	rcv := rateSCfg.CloneSection()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expected %v \n but received \n %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestRatesOptsLoadFromJSON(t *testing.T) {
	rateOpts := &RatesOpts{
		ProfileIDs: []*utils.DynamicStringSliceOpt{
			{
				Value: []string{},
			},
		},
		StartTime: []*utils.DynamicStringOpt{
			{
				Value: utils.MetaNow,
			},
		},
		Usage: []*utils.DynamicDecimalBigOpt{
			{
				Value: nil,
			},
		},
		IntervalStart: []*utils.DynamicDecimalBigOpt{
			{
				Value: nil,
			},
		},
		ProfileIgnoreFilters: []*utils.DynamicBoolOpt{
			{
				Value: false,
			},
		},
	}

	//first check for empty cfg
	if err := rateOpts.loadFromJSONCfg(nil); err != nil {
		t.Error(err)
	}

	jsnCfg := &RatesOptsJson{
		ProfileIDs: []*utils.DynamicStringSliceOpt{
			{
				Tenant: "cgrates.org",
				Value:  []string{"RP2"},
			},
		},
		Usage: []*utils.DynamicStringOpt{
			{
				Tenant: "cgrates.org",
				Value:  "error",
			},
		},
	}
	errExpect := "can't convert <error> to decimal"
	if err := rateOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but  received \n %v", errExpect, err.Error())
	}

	jsnCfg = &RatesOptsJson{
		IntervalStart: []*utils.DynamicStringOpt{
			{
				Tenant: "cgrates.org",
				Value:  "error",
			},
		},
	}
	if err := rateOpts.loadFromJSONCfg(jsnCfg); err == nil || err.Error() != errExpect {
		t.Errorf("Expected %v \n but  received \n %v", errExpect, err.Error())
	}

}
