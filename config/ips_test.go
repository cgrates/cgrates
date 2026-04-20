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

func TestIPsCfgLoadFromJsonCfg(t *testing.T) {

	tests := []struct {
		name        string
		jsonCfg     *IPsJsonCfg
		expected    *IPsCfg
		expectedErr string
	}{
		{
			name: "With values",
			jsonCfg: &IPsJsonCfg{
				Enabled:             utils.BoolPointer(true),
				IndexedSelects:      utils.BoolPointer(false),
				StoreInterval:       utils.StringPointer("1s"),
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        utils.BoolPointer(false),
				Opts: &IPsOptsJson{
					AllocationID: utils.StringPointer(""),
					TTL:          utils.StringPointer("72h"),
				},
			},
			expected: &IPsCfg{
				Enabled:             true,
				IndexedSelects:      false,
				StoreInterval:       1 * time.Second,
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        false,
				Opts: &IPsOpts{
					AllocationID: "",
					TTL:          utils.DurationPointer(72 * time.Hour),
				},
			},
		},
		{
			name: "Invalid value for StoreInterval",
			jsonCfg: &IPsJsonCfg{
				Enabled:             utils.BoolPointer(true),
				IndexedSelects:      utils.BoolPointer(false),
				StoreInterval:       utils.StringPointer("1ss"),
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        utils.BoolPointer(false),
				Opts: &IPsOptsJson{
					AllocationID: utils.StringPointer(""),
					TTL:          utils.StringPointer("72h"),
				},
			},
			expected: &IPsCfg{
				Enabled:             true,
				IndexedSelects:      false,
				StoreInterval:       0,
				PrefixIndexedFields: utils.SliceStringPointer([]string{}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{}),
				NestedFields:        false,
				Opts: &IPsOpts{
					AllocationID: "",
					TTL:          utils.DurationPointer(72 * time.Hour),
				},
			},
			expectedErr: `time: unknown unit "ss" in duration "1ss"`,
		},
		{
			name: "Invalid value for Opts",
			jsonCfg: &IPsJsonCfg{
				Enabled:             utils.BoolPointer(true),
				IndexedSelects:      utils.BoolPointer(false),
				StoreInterval:       utils.StringPointer("1s"),
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        utils.BoolPointer(false),
				Opts: &IPsOptsJson{
					AllocationID: utils.StringPointer(""),
					TTL:          utils.StringPointer("err"),
				},
			},
			expected: &IPsCfg{
				Enabled:             true,
				IndexedSelects:      false,
				StoreInterval:       1 * time.Second,
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        false,
				Opts: &IPsOpts{
					AllocationID: "",
					TTL:          utils.DurationPointer(72 * time.Hour),
				},
			},
			expectedErr: `time: invalid duration "err"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := NewDefaultCGRConfig()

			if err := jsnCfg.ipsCfg.loadFromJSONCfg(tt.jsonCfg); err != nil && err.Error() != tt.expectedErr {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.expected, jsnCfg.ipsCfg) {
				t.Errorf("Expected %+v, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsnCfg.ipsCfg))
			}
		})
	}
}

func TestIPsOptsLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name        string
		optsJs      *IPsOptsJson
		expectedErr string
	}{
		{
			name: "With values",
			optsJs: &IPsOptsJson{
				AllocationID: utils.StringPointer("id1"),
				TTL:          utils.StringPointer("72h"),
			},
		},
		{
			name: "Empty fields",
			optsJs: &IPsOptsJson{
				AllocationID: utils.StringPointer(""),
				TTL:          utils.StringPointer(""),
			},
		},
		{
			name: "Invalid value for TTL",
			optsJs: &IPsOptsJson{
				AllocationID: utils.StringPointer("id1"),
				TTL:          utils.StringPointer("err"),
			},
			expectedErr: `time: invalid duration "err"`,
		},
		{
			name:   "Nil case",
			optsJs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.optsJs
			jsnCfg := NewDefaultCGRConfig()
			if err := jsnCfg.ipsCfg.Opts.loadFromJSONCfg(tt.optsJs); err != nil && err.Error() != tt.expectedErr {
				t.Error(err)
			}

			if !reflect.DeepEqual(want, tt.optsJs) {
				t.Errorf("Expected no changes, wanted %v but got %v", want, tt.optsJs)
			}
		})
	}
}

func TestIPsCfgAsMapInterface(t *testing.T) {
	tests := []struct {
		name       string
		cfgJSONStr string
		eMap       map[string]any
	}{
		{
			name: "Complete struct",
			cfgJSONStr: `{
				"ips": {
					"enabled": false,		
					"store_interval": "0s",		
					"indexed_selects": true,	
					"string_indexed_fields": ["*req.index"],	
					"prefix_indexed_fields": ["*req.index"],	
					"suffix_indexed_fields": ["*req.index"],	
					"exists_indexed_fields": ["*req.index"],	
					"nested_fields": false,		
					"opts": {
						"*allocationID": "",
						"*ttl": "72h"
					}
				}
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             false,
				utils.StoreIntervalCfg:       "0s",
				utils.IndexedSelectsCfg:      true,
				utils.StringIndexedFieldsCfg: []string{"*req.index"},
				utils.PrefixIndexedFieldsCfg: []string{"*req.index"},
				utils.SuffixIndexedFieldsCfg: []string{"*req.index"},
				utils.ExistsIndexedFieldsCfg: []string{"*req.index"},
				utils.NestedFieldsCfg:        false,
				utils.OptsCfg: map[string]any{
					utils.MetaAllocationIDCfg: "",
					utils.MetaTTLCfg:          259200000000000,
				},
			},
		},
		{
			name: "Empty struct",
			cfgJSONStr: `{
				"ips": {}
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             false,
				utils.StoreIntervalCfg:       "0s",
				utils.IndexedSelectsCfg:      true,
				utils.StringIndexedFieldsCfg: nil,
				utils.PrefixIndexedFieldsCfg: []string{},
				utils.SuffixIndexedFieldsCfg: []string{},
				utils.ExistsIndexedFieldsCfg: []string{},
				utils.NestedFieldsCfg:        false,
				utils.OptsCfg: map[string]any{
					utils.MetaAllocationIDCfg: "",
					utils.MetaTTLCfg:          259200000000000,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(tt.cfgJSONStr); err != nil {
				t.Error(err)
			} else if rcv := cgrCfg.ipsCfg.AsMapInterface(); !reflect.DeepEqual(utils.ToJSON(tt.eMap), utils.ToJSON(rcv)) {
				t.Errorf("Expected: %+v\n Received: %+v", tt.eMap, rcv)
			}
		})
	}
}

func TestIPsCfgClone(t *testing.T) {
	tests := []struct {
		name   string
		ipsCfg *IPsCfg
	}{
		{
			name: "With Values",
			ipsCfg: &IPsCfg{
				Enabled:             false,
				IndexedSelects:      true,
				StoreInterval:       1 * time.Millisecond,
				StringIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				PrefixIndexedFields: utils.SliceStringPointer([]string{"*req.index1", "*req.index2"}),
				SuffixIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				ExistsIndexedFields: utils.SliceStringPointer([]string{"*req.index1"}),
				NestedFields:        false,
				Opts: &IPsOpts{
					AllocationID: "",
					TTL:          utils.DurationPointer(72 * time.Hour),
				},
			},
		},
		{
			name: "With Nil Values",
			ipsCfg: &IPsCfg{
				Enabled:             false,
				IndexedSelects:      true,
				StoreInterval:       1 * time.Millisecond,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				ExistsIndexedFields: nil,
				NestedFields:        false,
				Opts: &IPsOpts{
					AllocationID: "",
					TTL:          utils.DurationPointer(72 * time.Hour),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ipsCfg.Clone()

			if !reflect.DeepEqual(result, tt.ipsCfg) {
				t.Errorf("Clone() = %v, want %v", result, tt.ipsCfg)
			}

			if result != nil && result == tt.ipsCfg {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}

func TestIPsOptsClone(t *testing.T) {
	tests := []struct {
		name    string
		ipsOpts *IPsOpts
	}{
		{
			name: "Complete IPsOpts",
			ipsOpts: &IPsOpts{
				AllocationID: "testID",
				TTL:          utils.DurationPointer(72 * time.Hour),
			},
		},
		{
			name: "Nil TTL",
			ipsOpts: &IPsOpts{
				AllocationID: "testID",
				TTL:          nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ipsOpts.Clone()

			if !reflect.DeepEqual(result, tt.ipsOpts) {
				t.Errorf("Clone() = %v, want %v", result, tt.ipsOpts)
			}

			if result != nil && result == tt.ipsOpts {
				t.Errorf("Clone returned the same instance, expected a new instance")
			}
		})
	}
}
