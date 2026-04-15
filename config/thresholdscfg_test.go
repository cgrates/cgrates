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

	"github.com/cgrates/cgrates/utils"
)

func TestThresholdSCfgloadFromJsonCfg(t *testing.T) {
	tests := []struct {
		name     string
		cfgJSON  *ThresholdSJsonCfg
		expected *ThresholdSCfg
		wantErr  string
	}{
		{
			name: "With values",
			cfgJSON: &ThresholdSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Store_interval:        utils.StringPointer("2"),
				String_indexed_fields: &[]string{"*req.prefix"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
				Opts: &ThresholdsOptsJson{
					ProfileIDs:           &[]string{},
					ProfileIgnoreFilters: utils.BoolPointer(false),
				},
			},
			expected: &ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       2,
				StringIndexedFields: &[]string{"*req.prefix"},
				PrefixIndexedFields: &[]string{"*req.index1"},
				SuffixIndexedFields: &[]string{"*req.index1"},
				ExistsIndexedFields: &[]string{"*req.index1"},
				EEsConns:            []string{},
				NestedFields:        true,
				Opts: &ThresholdsOpts{
					ProfileIDs:           []string{},
					ProfileIgnoreFilters: false,
				},
			},
		},
		{
			name: "Store_interval invalid value",
			cfgJSON: &ThresholdSJsonCfg{
				Enabled:               utils.BoolPointer(true),
				Indexed_selects:       utils.BoolPointer(true),
				Store_interval:        utils.StringPointer("2ss"),
				String_indexed_fields: &[]string{"*req.prefix"},
				Prefix_indexed_fields: &[]string{"*req.index1"},
				Suffix_indexed_fields: &[]string{"*req.index1"},
				ExistsIndexedFields:   &[]string{"*req.index1"},
				Nested_fields:         utils.BoolPointer(true),
				Opts: &ThresholdsOptsJson{
					ProfileIDs:           &[]string{},
					ProfileIgnoreFilters: utils.BoolPointer(false),
				},
			},
			expected: &ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				StringIndexedFields: nil,
				PrefixIndexedFields: &[]string{},
				SuffixIndexedFields: &[]string{},
				ExistsIndexedFields: &[]string{},
				EEsConns:            []string{},
				NestedFields:        false,
				Opts: &ThresholdsOpts{
					ProfileIDs:           []string{},
					ProfileIgnoreFilters: false,
				},
			},
			wantErr: `time: unknown unit "ss" in duration "2ss"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonCfg := NewDefaultCGRConfig()
			if err := jsonCfg.thresholdSCfg.loadFromJSONCfg(tt.cfgJSON); err != nil && tt.wantErr != err.Error() {
				t.Error(err)
			} else if !reflect.DeepEqual(tt.expected, jsonCfg.thresholdSCfg) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(tt.expected), utils.ToJSON(jsonCfg.thresholdSCfg))
			}
		})
	}
}

func TestThresholdsOptsLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		optsJSON *ThresholdsOptsJson
	}{
		{
			name:     "Nil case",
			optsJSON: nil,
		},
		{
			name:     "Empty",
			optsJSON: &ThresholdsOptsJson{},
		},
		{
			name: "With Values",
			optsJSON: &ThresholdsOptsJson{
				ProfileIDs:           &[]string{},
				ProfileIgnoreFilters: utils.BoolPointer(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.optsJSON
			jsnCfg := NewDefaultCGRConfig()
			jsnCfg.thresholdSCfg.Opts.loadFromJSONCfg(tt.optsJSON)

			if !reflect.DeepEqual(want, tt.optsJSON) {
				t.Errorf("Expected no changes, wanted %v but got %v", want, tt.optsJSON)
			}
		})
	}
}

func TestThresholdSCfgAsMapInterface(t *testing.T) {
	tests := []struct {
		name       string
		cfgJSONStr string
		eMap       map[string]any
	}{
		{
			name: "Empty",
			cfgJSONStr: `{
				"thresholds": {},
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             false,
				utils.StoreIntervalCfg:       "",
				utils.IndexedSelectsCfg:      true,
				utils.PrefixIndexedFieldsCfg: []string{},
				utils.SuffixIndexedFieldsCfg: []string{},
				utils.ExistsIndexedFieldsCfg: []string{},
				utils.EEsExporterIDsCfg:      []string{},
				utils.EEsConnsCfg:            []string{},
				utils.NestedFieldsCfg:        false,
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string{},
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
			},
		},
		{
			name: "With values",
			cfgJSONStr: `{
				"thresholds": {								
					"enabled": true,						
					"store_interval": "96h",					
					"indexed_selects": false,	
					"string_indexed_fields": ["*req.string"],
					"prefix_indexed_fields": ["*req.prefix","*req.indexed","*req.fields"],	
					"suffix_indexed_fields": ["*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"],		
					"exists_indexed_fields": ["*req.exists_indexed_field"],		
					"nested_fields": true,					
				},		
			}`,
			eMap: map[string]any{
				utils.EnabledCfg:             true,
				utils.StoreIntervalCfg:       "96h0m0s",
				utils.IndexedSelectsCfg:      false,
				utils.StringIndexedFieldsCfg: []string{"*req.string"},
				utils.PrefixIndexedFieldsCfg: []string{"*req.prefix", "*req.indexed", "*req.fields"},
				utils.SuffixIndexedFieldsCfg: []string{"*req.suffix_indexed_fields1", "*req.suffix_indexed_fields2"},
				utils.ExistsIndexedFieldsCfg: []string{"*req.exists_indexed_field"},
				utils.NestedFieldsCfg:        true,
				utils.EEsExporterIDsCfg:      []string{},
				utils.EEsConnsCfg:            []string{},
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string{},
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if cgrCfg, err := NewCGRConfigFromJSONStringWithDefaults(tt.cfgJSONStr); err != nil {
				t.Error(err)
			} else if rcv := cgrCfg.thresholdSCfg.AsMapInterface(); !reflect.DeepEqual(rcv, tt.eMap) {
				t.Errorf("Expextec %+v \n, recevied %+v", utils.ToJSON(tt.eMap), utils.ToJSON(rcv))
			}
		})
	}
}

func TestThresholdSCfgClone(t *testing.T) {
	ban := &ThresholdSCfg{
		Enabled:             true,
		IndexedSelects:      true,
		StoreInterval:       2,
		EEsExporterIDs:      []string{"idTest"},
		StringIndexedFields: &[]string{"*req.index1"},
		PrefixIndexedFields: &[]string{"*req.index1"},
		SuffixIndexedFields: &[]string{"*req.index1"},
		NestedFields:        true,
		Opts: &ThresholdsOpts{
			ProfileIDs: []string{},
		},
	}
	rcv := ban.Clone()
	if !reflect.DeepEqual(ban, rcv) {
		t.Errorf("Expected: %+v\nReceived: %+v", utils.ToJSON(ban), utils.ToJSON(rcv))
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

func TestThresholdSCfgLoadFromJSONCfgSessionsConns(t *testing.T) {
	tests := []struct {
		name           string
		input          *ThresholdSJsonCfg
		expectedOutput []string
	}{
		{
			name: "Empty Sessions_conns",
			input: &ThresholdSJsonCfg{
				Sessions_conns: []string{},
			},
			expectedOutput: []string{},
		},
		{
			name: "Non-empty Sessions_conns without MetaInternal",
			input: &ThresholdSJsonCfg{
				Sessions_conns: []string{"conn1", "conn2"},
			},
			expectedOutput: []string{"conn1", "conn2"},
		},
		{
			name: "Non-empty Sessions_conns with MetaInternal",
			input: &ThresholdSJsonCfg{
				Sessions_conns: []string{"conn1", utils.MetaInternal},
			},
			expectedOutput: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
		},
		{
			name:           "Nil Sessions_conns",
			input:          nil,
			expectedOutput: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg ThresholdSCfg
			err := cfg.loadFromJSONCfg(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(cfg.SessionSConns) != len(tt.expectedOutput) {
				t.Fatalf("Expected %v, got %v", tt.expectedOutput, cfg.SessionSConns)
			}
			for i, conn := range cfg.SessionSConns {
				if conn != tt.expectedOutput[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expectedOutput[i], i, conn)
				}
			}
		})
	}
}

func TestThresholdSCfgLoadFromJSONCfgApiersConns(t *testing.T) {
	tests := []struct {
		name           string
		input          *ThresholdSJsonCfg
		expectedOutput []string
	}{
		{
			name: "Empty Apiers_conns",
			input: &ThresholdSJsonCfg{
				Apiers_conns: []string{},
			},
			expectedOutput: []string{},
		},
		{
			name: "Non-empty Apiers_conns without MetaInternal",
			input: &ThresholdSJsonCfg{
				Apiers_conns: []string{"conn1", "conn2"},
			},
			expectedOutput: []string{"conn1", "conn2"},
		},
		{
			name: "Non-empty Apiers_conns with MetaInternal",
			input: &ThresholdSJsonCfg{
				Apiers_conns: []string{"conn1", utils.MetaInternal},
			},
			expectedOutput: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg ThresholdSCfg
			err := cfg.loadFromJSONCfg(tt.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(cfg.ApierSConns) != len(tt.expectedOutput) {
				t.Fatalf("Expected %v, got %v", tt.expectedOutput, cfg.ApierSConns)
			}
			for i, conn := range cfg.ApierSConns {
				if conn != tt.expectedOutput[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expectedOutput[i], i, conn)
				}
			}
		})
	}
}

func TestThresholdSCfgAsMapInterfaceSessionSConns(t *testing.T) {
	tests := []struct {
		name           string
		cfg            ThresholdSCfg
		expectedOutput map[string]any
	}{
		{
			name: "SessionSConns is nil",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       nil,
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              nil,
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
			},
		},
		{
			name: "SessionSConns has values",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{"conn1", "conn2", utils.MetaInternal},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              nil,
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{"conn1", "conn2", utils.MetaInternal},
			},
		},
		{
			name: "SessionSConns has MetaInternal and MetaSessionS",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              nil,
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{utils.MetaInternal},
			},
		},
		{
			name: "SessionSConns with conn and meta-s",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              nil,
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{"conn1", utils.MetaInternal},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.cfg.AsMapInterface()

			if !reflect.DeepEqual(utils.ToJSON(output), utils.ToJSON(tt.expectedOutput)) {
				t.Errorf("Expected %#v \n, received %#v", tt.expectedOutput, output)
			}
		})
	}
}

func TestThresholdSCfgLoadFromJSONCfgEesConns(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *ThresholdSJsonCfg
		expectedOutput []string
	}{
		{
			name: "Nil Ees_conn",
			cfg: &ThresholdSJsonCfg{
				Ees_conns: nil,
			},
			expectedOutput: nil,
		},
		{
			name: "Ees_conn with MetaInternal and MetaEEs",
			cfg: &ThresholdSJsonCfg{
				Ees_conns: utils.SliceStringPointer([]string{utils.MetaInternal, utils.MetaEEs}),
			},
			expectedOutput: []string{utils.MetaInternal, utils.MetaEEs},
		},
		{
			name: "Ees_conn empty",
			cfg: &ThresholdSJsonCfg{
				Ees_conns: utils.SliceStringPointer([]string{}),
			},
			expectedOutput: []string{},
		},
		{
			name: "Ees_conns with values",
			cfg: &ThresholdSJsonCfg{
				Ees_conns: utils.SliceStringPointer([]string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)}),
			},
			expectedOutput: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg ThresholdSCfg
			err := cfg.loadFromJSONCfg(tt.cfg)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(cfg.EEsConns) != len(tt.expectedOutput) {
				t.Fatalf("Expected %v, got %v", tt.expectedOutput, cfg.EEsConns)
			}
		})
	}
}

func TestThresholdSCfgAsMapInterfaceEEsConns(t *testing.T) {
	tests := []struct {
		name           string
		cfg            ThresholdSCfg
		expectedOutput map[string]any
	}{
		{
			name: "EEsConns is Empty",
			cfg: ThresholdSCfg{
				EEsConns:       []string{},
				Enabled:        true,
				IndexedSelects: true,
				StoreInterval:  0,
				NestedFields:   true,
				SessionSConns:  nil,
				Opts:           &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.EEsConnsCfg: []string{},
			},
		},
		{
			name: "EEsConns is Nil",
			cfg: ThresholdSCfg{
				EEsConns:       nil,
				Enabled:        true,
				IndexedSelects: true,
				StoreInterval:  0,
				NestedFields:   true,
				SessionSConns:  nil,
				Opts:           &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
			},
		},
		{
			name: "EEsConns has values",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{"conn1", "conn2"},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
				EEsConns:            []string{"conn1", "conn2"},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{"conn1", "conn2"},
				utils.EEsConnsCfg:      []string{"conn1", "conn2"},
			},
		},
		{
			name: "EEsConns has MetaInternal and MetaSessionS",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
				EEsConns:            []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{utils.MetaInternal},
				utils.EEsConnsCfg:      []string{utils.MetaInternal},
			},
		},
		{
			name: "EEsConns has values and MetaInternal",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS)},
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
				EEsConns:            []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.SessionSConnsCfg: []string{"conn1", utils.MetaInternal},
				utils.EEsConnsCfg:      []string{"conn1", utils.MetaInternal},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.cfg.AsMapInterface()

			if !reflect.DeepEqual(utils.ToJSON(output), utils.ToJSON(tt.expectedOutput)) {
				t.Errorf("Expected %#v \n, received %#v", tt.expectedOutput, output)
			}
		})
	}
}
func TestThresholdSCfgAsMapInterfaceApierSConns(t *testing.T) {
	tests := []struct {
		name           string
		cfg            ThresholdSCfg
		expectedOutput map[string]any
	}{
		{
			name: "ApierSConns is nil",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       nil,
				ApierSConns:         nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
			},
		},
		{
			name: "ApierSConns is Empty",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       nil,
				ApierSConns:         []string{},
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.ApierSConnsCfg: []string{},
			},
		},
		{
			name: "ApierSConns has values",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				ApierSConns:         []string{"conn1", "conn2", utils.MetaInternal},
				SessionSConns:       nil,
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.ApierSConnsCfg: []string{"conn1", "conn2", utils.MetaInternal},
			},
		},
		{
			name: "ApierSConns has conn1 and meta-s",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       nil,
				ApierSConns:         []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)},
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.ApierSConnsCfg: []string{"conn1", utils.MetaInternal},
			},
		},
		{
			name: "ApierSConns has MetaInternal and MetaApier",
			cfg: ThresholdSCfg{
				Enabled:             true,
				IndexedSelects:      true,
				StoreInterval:       0,
				NestedFields:        true,
				SessionSConns:       nil,
				ApierSConns:         []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier)},
				StringIndexedFields: nil,
				PrefixIndexedFields: nil,
				SuffixIndexedFields: nil,
				Opts:                &ThresholdsOpts{},
			},
			expectedOutput: map[string]any{
				utils.EnabledCfg:        true,
				utils.EEsExporterIDsCfg: []string{},
				utils.IndexedSelectsCfg: true,
				utils.NestedFieldsCfg:   true,
				utils.StoreIntervalCfg:  "",
				utils.OptsCfg: map[string]any{
					utils.MetaProfileIDs:              []string(nil),
					utils.MetaProfileIgnoreFiltersCfg: false,
				},
				utils.ApierSConnsCfg: []string{utils.MetaInternal},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.cfg.AsMapInterface()

			if !reflect.DeepEqual(utils.ToJSON(output), utils.ToJSON(tt.expectedOutput)) {
				t.Errorf("Expected %#v \n, received %#v", tt.expectedOutput, output)
			}
		})
	}
}

func TestThresholdSCfgCloneSessionSConns(t *testing.T) {
	tests := []struct {
		name           string
		original       *ThresholdSCfg
		expectedCloned []string
	}{
		{
			name: "SessionSConns is nil",
			original: &ThresholdSCfg{
				SessionSConns: nil,
				Opts:          &ThresholdsOpts{},
			},
			expectedCloned: nil,
		},
		{
			name: "SessionSConns has values",
			original: &ThresholdSCfg{
				SessionSConns: []string{"conn1", "conn2"},
				Opts:          &ThresholdsOpts{},
			},
			expectedCloned: []string{"conn1", "conn2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned := tt.original.Clone()

			if tt.expectedCloned == nil {
				if cloned.SessionSConns != nil {
					t.Errorf("Expected nil, got %v", cloned.SessionSConns)
				}
			} else {
				if len(cloned.SessionSConns) != len(tt.expectedCloned) {
					t.Errorf("Expected %v, got %v", tt.expectedCloned, cloned.SessionSConns)
				}
				for i := range cloned.SessionSConns {
					if cloned.SessionSConns[i] != tt.expectedCloned[i] {
						t.Errorf("Expected %v, got %v", tt.expectedCloned, cloned.SessionSConns)
					}
				}
			}
		})
	}
}

func TestThresholdSCfgCloneApierSConns(t *testing.T) {
	tests := []struct {
		name           string
		original       *ThresholdSCfg
		expectedCloned []string
	}{
		{
			name: "ApierSConns is nil",
			original: &ThresholdSCfg{
				ApierSConns: nil,
				Opts:        &ThresholdsOpts{},
			},
			expectedCloned: nil,
		},
		{
			name: "ApierSConns has values",
			original: &ThresholdSCfg{
				ApierSConns: []string{"conn1", "conn2"},
				Opts:        &ThresholdsOpts{},
			},
			expectedCloned: []string{"conn1", "conn2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned := tt.original.Clone()

			if tt.expectedCloned == nil {
				if cloned.ApierSConns != nil {
					t.Errorf("Expected nil, got %v", cloned.ApierSConns)
				}
			} else {
				if len(cloned.ApierSConns) != len(tt.expectedCloned) {
					t.Errorf("Expected %v, got %v", tt.expectedCloned, cloned.ApierSConns)
				}
				for i := range cloned.ApierSConns {
					if cloned.ApierSConns[i] != tt.expectedCloned[i] {
						t.Errorf("Expected %v, got %v", tt.expectedCloned, cloned.ApierSConns)
					}
				}
			}
		})
	}
}
