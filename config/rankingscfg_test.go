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

func TestRankingSCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name     string
		jsonCfg  *RankingSJsonCfg
		expected RankingSCfg
		wantErr  bool
	}{
		{
			name: "successful load, enabled true with stats and thresholds",
			jsonCfg: &RankingSJsonCfg{
				Enabled: utils.BoolPointer(true),
				Conns: map[string][]*DynamicStringSliceOpt{
					utils.MetaStats:      {{Values: []string{"conn1", "conn2"}}},
					utils.MetaThresholds: {{Values: []string{"thresh1", "thresh2"}}},
					utils.MetaEEs:        {{Values: []string{"ees1", "ees2"}}},
				},
				Scheduled_ids:    map[string][]string{"sched1": {"id1", "id2"}},
				Store_interval:   utils.StringPointer("1m30s"),
				Ees_exporter_ids: &[]string{"exporter1", "exporter2"},
			},
			expected: RankingSCfg{
				Enabled: true,
				Conns: map[string][]*DynamicStringSliceOpt{
					utils.MetaStats:      {{Values: []string{"conn1", "conn2"}}},
					utils.MetaThresholds: {{Values: []string{"thresh1", "thresh2"}}},
					utils.MetaEEs:        {{Values: []string{"ees1", "ees2"}}},
				},
				ScheduledIDs:   map[string][]string{"sched1": {"id1", "id2"}},
				StoreInterval:  time.Minute + 30*time.Second,
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
			wantErr: false,
		},
		{
			name:     "nil jsonCfg",
			jsonCfg:  nil,
			expected: RankingSCfg{},
			wantErr:  false,
		},
		{
			name: "invalid Store_interval format",
			jsonCfg: &RankingSJsonCfg{
				Store_interval: utils.StringPointer("invalid_duration"),
			},
			expected: RankingSCfg{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rankingCfg RankingSCfg
			err := rankingCfg.loadFromJSONCfg(tt.jsonCfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("loadFromJSONCfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(rankingCfg.Conns, tt.expected.Conns) {
				t.Errorf("Conns = %v, want %v", utils.ToJSON(rankingCfg.Conns), utils.ToJSON(tt.expected.Conns))
			}

			if !reflect.DeepEqual(rankingCfg.ScheduledIDs, tt.expected.ScheduledIDs) {
				t.Errorf("ScheduledIDs = %v, want %v", rankingCfg.ScheduledIDs, tt.expected.ScheduledIDs)
			}

			if rankingCfg.StoreInterval != tt.expected.StoreInterval {
				t.Errorf("StoreInterval = %v, want %v", rankingCfg.StoreInterval, tt.expected.StoreInterval)
			}

			if !reflect.DeepEqual(rankingCfg.EEsExporterIDs, tt.expected.EEsExporterIDs) {
				t.Errorf("EEsExporterIDs = %v, want %v", rankingCfg.EEsExporterIDs, tt.expected.EEsExporterIDs)
			}

			if rankingCfg.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", rankingCfg.Enabled, tt.expected.Enabled)
			}
		})
	}
}

func TestDiffRankingsJsonCfg(t *testing.T) {
	tests := []struct {
		name     string
		v1       *RankingSCfg
		v2       *RankingSCfg
		expected *RankingSJsonCfg
	}{
		{
			name: "enabled diff",
			v1: &RankingSCfg{
				Enabled: false,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn1"}}}},
			},
			v2: &RankingSCfg{
				Enabled: true,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn1"}}}},
			},
			expected: &RankingSJsonCfg{
				Enabled: utils.BoolPointer(true),
			},
		},
		{
			name: "conns diff",
			v1: &RankingSCfg{
				Enabled: true,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn1"}}}},
			},
			v2: &RankingSCfg{
				Enabled: true,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn2"}}}},
			},
			expected: &RankingSJsonCfg{
				Conns: map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn2"}}}},
			},
		},
		{
			name: "storeInterval diff",
			v1:   &RankingSCfg{StoreInterval: 10 * time.Second},
			v2:   &RankingSCfg{StoreInterval: 20 * time.Second},
			expected: &RankingSJsonCfg{
				Store_interval: utils.StringPointer((20 * time.Second).String()),
			},
		},
		{
			name: "eesExporterIDs diff",
			v1:   &RankingSCfg{EEsExporterIDs: []string{"exporter1"}},
			v2:   &RankingSCfg{EEsExporterIDs: []string{"exporter2"}},
			expected: &RankingSJsonCfg{
				Ees_exporter_ids: utils.SliceStringPointer([]string{"exporter2"}),
			},
		},
		{
			name: "no diff",
			v1: &RankingSCfg{
				Enabled: true,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn1"}}}},
			},
			v2: &RankingSCfg{
				Enabled: true,
				Conns:   map[string][]*DynamicStringSliceOpt{utils.MetaStats: {{Values: []string{"conn1"}}}},
			},
			expected: &RankingSJsonCfg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diffRankingsJsonCfg(nil, tt.v1, tt.v2)

			if (result.Enabled == nil && tt.expected.Enabled != nil) || (result.Enabled != nil && *result.Enabled != *tt.expected.Enabled) {
				t.Errorf("diffRankingsJsonCfg() Enabled = %v, want %v", result.Enabled, tt.expected.Enabled)
			}

			if !reflect.DeepEqual(result.Conns, tt.expected.Conns) {
				t.Errorf("diffRankingsJsonCfg() Conns = %v, want %v", utils.ToJSON(result.Conns), utils.ToJSON(tt.expected.Conns))
			}

			if (result.Store_interval == nil && tt.expected.Store_interval != nil) || (result.Store_interval != nil && *result.Store_interval != *tt.expected.Store_interval) {
				t.Errorf("diffRankingsJsonCfg() Store_interval = %v, want %v", result.Store_interval, tt.expected.Store_interval)
			}

			if (result.Ees_exporter_ids == nil && tt.expected.Ees_exporter_ids != nil) || (result.Ees_exporter_ids != nil && !reflect.DeepEqual(*result.Ees_exporter_ids, *tt.expected.Ees_exporter_ids)) {
				t.Errorf("diffRankingsJsonCfg() Ees_exporter_ids = %v, want %v", result.Ees_exporter_ids, tt.expected.Ees_exporter_ids)
			}
		})
	}
}

func TestRankingSCfgClone_CloneSection(t *testing.T) {
	tests := []struct {
		name     string
		input    *RankingSCfg
		expected *RankingSCfg
	}{
		{
			name: "clone with non-nil fields",
			input: &RankingSCfg{
				Enabled:       true,
				StoreInterval: 20 * time.Second,
				Conns: map[string][]*DynamicStringSliceOpt{
					utils.MetaStats:      {{Values: []string{"conn1", "conn2"}}},
					utils.MetaThresholds: {{Values: []string{"thresh1", "thresh2"}}},
					utils.MetaEEs:        {{Values: []string{"ees1", "ees2"}}},
				},
				ScheduledIDs:   map[string][]string{"sched1": {"id1", "id2"}},
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
			expected: &RankingSCfg{
				Enabled:       true,
				StoreInterval: 20 * time.Second,
				Conns: map[string][]*DynamicStringSliceOpt{
					utils.MetaStats:      {{Values: []string{"conn1", "conn2"}}},
					utils.MetaThresholds: {{Values: []string{"thresh1", "thresh2"}}},
					utils.MetaEEs:        {{Values: []string{"ees1", "ees2"}}},
				},
				ScheduledIDs:   map[string][]string{"sched1": {"id1", "id2"}},
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
		},
		{
			name: "clone with zero and nil fields",
			input: &RankingSCfg{
				Enabled:        false,
				StoreInterval:  0,
				Conns:          nil,
				ScheduledIDs:   nil,
				EEsExporterIDs: nil,
			},
			expected: &RankingSCfg{
				Enabled:        false,
				StoreInterval:  0,
				Conns:          nil,
				ScheduledIDs:   nil,
				EEsExporterIDs: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_Clone", func(t *testing.T) {
			clone := tt.input.Clone()
			if clone.Enabled != tt.expected.Enabled {
				t.Errorf("Clone() Enabled = %v, want %v", clone.Enabled, tt.expected.Enabled)
			}
			if clone.StoreInterval != tt.expected.StoreInterval {
				t.Errorf("Clone() StoreInterval = %v, want %v", clone.StoreInterval, tt.expected.StoreInterval)
			}
			if !reflect.DeepEqual(clone.Conns, tt.expected.Conns) {
				t.Errorf("Clone() Conns = %v, want %v", utils.ToJSON(clone.Conns), utils.ToJSON(tt.expected.Conns))
			}
			if !reflect.DeepEqual(clone.ScheduledIDs, tt.expected.ScheduledIDs) {
				t.Errorf("Clone() ScheduledIDs = %v, want %v", clone.ScheduledIDs, tt.expected.ScheduledIDs)
			}
			if !reflect.DeepEqual(clone.EEsExporterIDs, tt.expected.EEsExporterIDs) {
				t.Errorf("Clone() EEsExporterIDs = %v, want %v", clone.EEsExporterIDs, tt.expected.EEsExporterIDs)
			}
			if tt.input.EEsExporterIDs != nil && clone.EEsExporterIDs != nil && &tt.input.EEsExporterIDs[0] == &clone.EEsExporterIDs[0] {
				t.Error("Clone() EEsExporterIDs has the same reference, expected a deep copy")
			}
		})

		t.Run(tt.name+"_CloneSection", func(t *testing.T) {
			clonedSection := tt.input.CloneSection()

			clonedRankingSCfg, ok := clonedSection.(*RankingSCfg)
			if !ok {
				t.Errorf("CloneSection() returned wrong type, got %T, want *RankingSCfg", clonedSection)
			}
			if !reflect.DeepEqual(clonedRankingSCfg, tt.expected) {
				t.Errorf("CloneSection() = %v, want %v", utils.ToJSON(clonedRankingSCfg), utils.ToJSON(tt.expected))
			}
			if tt.input.EEsExporterIDs != nil && clonedRankingSCfg.EEsExporterIDs != nil && &tt.input.EEsExporterIDs[0] == &clonedRankingSCfg.EEsExporterIDs[0] {
				t.Error("CloneSection() EEsExporterIDs has the same reference, expected a deep copy")
			}
		})
	}
}

func TestRankingSCfgAsMapInterfaceStoreInterval(t *testing.T) {
	rnk := &RankingSCfg{
		StoreInterval: 30 * time.Second,
	}
	result := rnk.AsMapInterface()
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Expected result to be of type map[string]any, got %T", result)
	}
	expectedValue := "30s"
	if val, ok := resultMap[utils.StoreIntervalCfg]; !ok || val != expectedValue {
		t.Errorf("Expected StoreInterval to be %v, got %v", expectedValue, val)
	}
}
