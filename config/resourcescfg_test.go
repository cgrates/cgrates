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

func TestResourceSConfigloadFromJsonCfg(t *testing.T) {

	tests := []struct {
		name     string
		cfgJSON  *ResourceSJsonCfg
		expected *ResourceSConfig
		wantErr  string
	}{
		{
			name: "With values",
			cfgJSON: &ResourceSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
				Store_interval:        utils.StringPointer("2s"),
				String_indexed_fields: &[]string{"*req.index1"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
			},
			expected: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2 * time.Second,
				ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
				StringIndexedFields: &[]string{"*req.index1"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				ExistsIndexedFields: &[]string{"*req.index1"},
				NestedFields:        true,
				Opts: &ResourcesOpts{
					Units: 1,
				},
			},
		},
		{
			name: "Correct input for opts",
			cfgJSON: &ResourceSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
				Store_interval:        utils.StringPointer("2s"),
				String_indexed_fields: &[]string{"*req.index1"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
				Opts: &ResourcesOptsJson{
					UsageTTL: utils.StringPointer("1000"),
				},
			},
			expected: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2 * time.Second,
				ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
				StringIndexedFields: &[]string{"*req.index1"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				ExistsIndexedFields: &[]string{"*req.index1"},
				NestedFields:        true,
				Opts: &ResourcesOpts{
					Units:    1,
					UsageTTL: utils.DurationPointer(1000),
				},
			},
		},
		{
			name: "Invalid input for opts",
			cfgJSON: &ResourceSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Thresholds_conns:      &[]string{utils.MetaInternal, "*conn1"},
				Store_interval:        utils.StringPointer("2s"),
				String_indexed_fields: &[]string{"*req.index1"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
				Opts: &ResourcesOptsJson{
					UsageTTL: utils.StringPointer("test"),
				},
			},
			expected: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2 * time.Second,
				ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
				StringIndexedFields: &[]string{"*req.index1"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				ExistsIndexedFields: &[]string{"*req.index1"},
				NestedFields:        true,
				Opts: &ResourcesOpts{
					Units: 1,
				},
			},
			wantErr: `time: invalid duration "test"`,
		},
		{
			name: "Invalid input for StoreInterval",
			cfgJSON: &ResourceSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Thresholds_conns:      &[]string{},
				Store_interval:        utils.StringPointer("2ss"),
				String_indexed_fields: &[]string{"*req.index1"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
				Opts: &ResourcesOptsJson{
					UsageTTL: utils.StringPointer("test"),
				},
			},
			expected: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				ThresholdSConns:     []string{},
				PrefixIndexedFields: &[]string{},
				SuffixIndexedFields: &[]string{},
				ExistsIndexedFields: &[]string{},
				Opts: &ResourcesOpts{
					Units: 1,
				},
			},
			wantErr: `time: unknown unit "ss" in duration "2ss"`,
		},
		{
			name:     "Nil Case",
			cfgJSON:  nil,
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultCGRConfig()
			if err := cfg.resourceSCfg.loadFromJSONCfg(tt.cfgJSON); err != nil && tt.wantErr != err.Error() {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.expected, cfg.resourceSCfg) && tt.cfgJSON != nil {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(cfg.resourceSCfg))
			}
		})
	}
}

func TestResourcesOptsLoadFromJsonCfg(t *testing.T) {

	tests := []struct {
		name              string
		resourcesOptsJson *ResourcesOptsJson
		expected          *ResourcesOpts
		wantErr           string
	}{
		{
			name: "With values",
			resourcesOptsJson: &ResourcesOptsJson{
				UsageID: utils.StringPointer(utils.EmptyString),
				Units:   utils.Float64Pointer(1),
			},
			expected: &ResourcesOpts{
				UsageID:  utils.EmptyString,
				UsageTTL: utils.DurationPointer(1000),
				Units:    1,
			},
		},
		{
			name: "With nil values",
			resourcesOptsJson: &ResourcesOptsJson{
				UsageID: nil,
				Units:   nil,
			},
			expected: &ResourcesOpts{
				UsageID: "",
				Units:   0,
			},
		},
		{
			name: "Invalid input for UsageTTL",
			resourcesOptsJson: &ResourcesOptsJson{
				UsageID:  utils.StringPointer(utils.EmptyString),
				UsageTTL: utils.StringPointer("test"),
				Units:    utils.Float64Pointer(1),
			},
			expected: &ResourcesOpts{
				UsageID:  utils.EmptyString,
				UsageTTL: utils.DurationPointer(1000),
				Units:    1,
			},
			wantErr: `time: invalid duration "test"`,
		},
		{
			name:              "Nil case",
			resourcesOptsJson: nil,
			expected:          nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewDefaultCGRConfig()
			if err := cfg.resourceSCfg.Opts.loadFromJSONCfg(tt.resourcesOptsJson); err != nil && tt.wantErr != err.Error() {
				t.Error(err)
			}
		})
	}
}

func TestResourceSConfigAsMapInterface(t *testing.T) {
	tests := []struct {
		name       string
		cfgJSONStr string
		eMap       map[string]any
	}{
		{
			name: "Empty",
			cfgJSONStr: `{
				"resources": {},
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             false,
				utils.StoreIntervalCfg:       utils.EmptyString,
				utils.ThresholdSConnsCfg:     []string{},
				utils.IndexedSelectsCfg:      true,
				utils.PrefixIndexedFieldsCfg: []string{},
				utils.SuffixIndexedFieldsCfg: []string{},
				utils.ExistsIndexedFieldsCfg: []string{},
				utils.NestedFieldsCfg:        false,
				utils.OptsCfg: map[string]any{
					utils.MetaUnitsCfg:   1.,
					utils.MetaUsageIDCfg: "",
				},
			},
		},
		{
			name: "With Values",
			cfgJSONStr: `{
				"resources": {								
					"enabled": true,						
					"store_interval": "7m",					
					"thresholds_conns": ["*internal:*thresholds", "*conn1"],					
					"indexed_selects":true,		
					"string_indexed_fields": ["*req.index1"],
					"prefix_indexed_fields": ["*req.prefix_indexed_fields1","*req.prefix_indexed_fields2"],
					"suffix_indexed_fields": ["*req.prefix_indexed_fields1"],
					"exists_indexed_fields": ["*req.exists_indexed_field"],
					"nested_fields": true,	
					"opts":{
						"*usageTTL":"1"

					}		
				},	
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             true,
				utils.StoreIntervalCfg:       "7m0s",
				utils.ThresholdSConnsCfg:     []string{utils.MetaInternal, "*conn1"},
				utils.IndexedSelectsCfg:      true,
				utils.StringIndexedFieldsCfg: []string{"*req.index1"},
				utils.PrefixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1", "*req.prefix_indexed_fields2"},
				utils.SuffixIndexedFieldsCfg: []string{"*req.prefix_indexed_fields1"},
				utils.ExistsIndexedFieldsCfg: []string{"*req.exists_indexed_field"},
				utils.NestedFieldsCfg:        true,
				utils.OptsCfg: map[string]any{
					utils.MetaUnitsCfg:    1.,
					utils.MetaUsageIDCfg:  "",
					utils.MetaUsageTTLCfg: 1 * time.Nanosecond,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(tt.cfgJSONStr); err != nil {
				t.Error(err)
			} else if rcv := cgrCfg.resourceSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, tt.eMap) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.eMap), utils.ToJSON(rcv))
			}
		})
	}

}

func TestResourceSConfigClone(t *testing.T) {
	tests := []struct {
		name            string
		resourceSConfig *ResourceSConfig
	}{
		{
			name: "Complete ResourceSConfig",
			resourceSConfig: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2 * time.Second,
				ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
				StringIndexedFields: &[]string{"*req.index1"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				NestedFields:        true,
				Opts: &ResourcesOpts{
					UsageTTL: utils.DurationPointer(1 * time.Second),
				},
			},
		},
		{
			name: "Nil Opts",
			resourceSConfig: &ResourceSConfig{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2 * time.Second,
				ThresholdSConns:     []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), "*conn1"},
				StringIndexedFields: &[]string{"*req.index1"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				NestedFields:        true,
				Opts:                nil,
			},
		},
		{
			name:            "Nil ResourceSConfig",
			resourceSConfig: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := tt.resourceSConfig.Clone()
			if !reflect.DeepEqual(tt.resourceSConfig, rcv) {
				t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(tt.resourceSConfig), utils.ToJSON(rcv))
			}

			if tt.resourceSConfig != nil && rcv != nil {
				if rcv.ThresholdSConns[1] = ""; tt.resourceSConfig.ThresholdSConns[1] != "*conn1" {
					t.Errorf("Expected clone to not modify the cloned")
				}
				if (*rcv.StringIndexedFields)[0] = ""; (*tt.resourceSConfig.StringIndexedFields)[0] != "*req.index1" {
					t.Errorf("Expected clone to not modify the cloned")
				}
				if (*rcv.PrefixIndexedFields)[0] = ""; (*tt.resourceSConfig.PrefixIndexedFields)[0] != "*req.index1" {
					t.Errorf("Expected clone to not modify the cloned")
				}
				if (*rcv.SuffixIndexedFields)[0] = ""; (*tt.resourceSConfig.SuffixIndexedFields)[0] != "*req.index1" {
					t.Errorf("Expected clone to not modify the cloned")
				}
			}
		})
	}
}
