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
	"slices"
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
				Enabled:          utils.BoolPointer(true),
				Stats_conns:      &[]string{"conn1", "conn2"},
				Thresholds_conns: &[]string{"thresh1", "thresh2"},
				Scheduled_ids:    map[string][]string{"sched1": {"id1", "id2"}},
				Store_interval:   utils.StringPointer("1m30s"),
				Ees_conns:        &[]string{"ees1", "ees2"},
				Ees_exporter_ids: &[]string{"exporter1", "exporter2"},
			},
			expected: RankingSCfg{
				Enabled:         true,
				StatSConns:      []string{"conn1", "conn2"},
				ThresholdSConns: []string{"thresh1", "thresh2"},
				ScheduledIDs:    map[string][]string{"sched1": {"id1", "id2"}},
				StoreInterval:   time.Minute + 30*time.Second,
				EEsConns:        []string{"ees1", "ees2"},
				EEsExporterIDs:  []string{"exporter1", "exporter2"},
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

			if len(rankingCfg.StatSConns) != len(tt.expected.StatSConns) {
				t.Errorf("StatSConns length = %d, want %d", len(rankingCfg.StatSConns), len(tt.expected.StatSConns))
			}
			for i := range rankingCfg.StatSConns {
				if rankingCfg.StatSConns[i] != tt.expected.StatSConns[i] {
					t.Errorf("StatSConns[%d] = %v, want %v", i, rankingCfg.StatSConns[i], tt.expected.StatSConns[i])
				}
			}

			if len(rankingCfg.ThresholdSConns) != len(tt.expected.ThresholdSConns) {
				t.Errorf("ThresholdSConns length = %d, want %d", len(rankingCfg.ThresholdSConns), len(tt.expected.ThresholdSConns))
			}

			for i := range rankingCfg.ThresholdSConns {
				if rankingCfg.ThresholdSConns[i] != tt.expected.ThresholdSConns[i] {
					t.Errorf("ThresholdSConns[%d] = %v, want %v", i, rankingCfg.ThresholdSConns[i], tt.expected.ThresholdSConns[i])
				}
			}

			if !reflect.DeepEqual(rankingCfg.ScheduledIDs, tt.expected.ScheduledIDs) {
				t.Errorf("ScheduledIDs = %v, want %v", rankingCfg.ScheduledIDs, tt.expected.ScheduledIDs)
			}

			if rankingCfg.StoreInterval != tt.expected.StoreInterval {
				t.Errorf("StoreInterval = %v, want %v", rankingCfg.StoreInterval, tt.expected.StoreInterval)
			}

			if len(rankingCfg.EEsConns) != len(tt.expected.EEsConns) {
				t.Errorf("EEsConns length = %d, want %d", len(rankingCfg.EEsConns), len(tt.expected.EEsConns))
			}

			for i := range rankingCfg.EEsConns {
				if rankingCfg.EEsConns[i] != tt.expected.EEsConns[i] {
					t.Errorf("EEsConns[%d] = %v, want %v", i, rankingCfg.EEsConns[i], tt.expected.EEsConns[i])
				}
			}

			if len(rankingCfg.EEsExporterIDs) != len(tt.expected.EEsExporterIDs) {
				t.Errorf("EEsExporterIDs length = %d, want %d", len(rankingCfg.EEsExporterIDs), len(tt.expected.EEsExporterIDs))
			}

			for i := range rankingCfg.EEsExporterIDs {
				if rankingCfg.EEsExporterIDs[i] != tt.expected.EEsExporterIDs[i] {
					t.Errorf("EEsExporterIDs[%d] = %v, want %v", i, rankingCfg.EEsExporterIDs[i], tt.expected.EEsExporterIDs[i])
				}
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
			v1:   &RankingSCfg{Enabled: false, StatSConns: []string{"conn1"}},
			v2:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			expected: &RankingSJsonCfg{
				Enabled: utils.BoolPointer(true),
			},
		},
		{
			name: "statSConns diff",
			v1:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			v2:   &RankingSCfg{Enabled: true, StatSConns: []string{"conn2"}},
			expected: &RankingSJsonCfg{
				Stats_conns: utils.SliceStringPointer([]string{"conn2"}),
			},
		},
		{
			name: "thresholdSConns diff",
			v1:   &RankingSCfg{Enabled: true, ThresholdSConns: []string{"threshold1"}},
			v2:   &RankingSCfg{Enabled: true, ThresholdSConns: []string{"threshold2"}},
			expected: &RankingSJsonCfg{
				Thresholds_conns: utils.SliceStringPointer([]string{"threshold2"}),
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
			name: "eesConns diff",
			v1:   &RankingSCfg{EEsConns: []string{"ees1"}},
			v2:   &RankingSCfg{EEsConns: []string{"ees2"}},
			expected: &RankingSJsonCfg{
				Ees_conns: utils.SliceStringPointer([]string{"ees2"}),
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
			name:     "no diff",
			v1:       &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			v2:       &RankingSCfg{Enabled: true, StatSConns: []string{"conn1"}},
			expected: &RankingSJsonCfg{},
		},
		{
			name:     "no diff",
			v1:       &RankingSCfg{Enabled: false, StatSConns: []string{"conn1", "conn2"}},
			v2:       &RankingSCfg{Enabled: false, StatSConns: []string{"conn1", "conn2"}},
			expected: &RankingSJsonCfg{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := diffRankingsJsonCfg(nil, tt.v1, tt.v2)

			if (result.Enabled == nil && tt.expected.Enabled != nil) || (result.Enabled != nil && *result.Enabled != *tt.expected.Enabled) {
				t.Errorf("diffRankingsJsonCfg() Enabled = %v, want %v", *result.Enabled, *tt.expected.Enabled)
			}

			if (result.Stats_conns == nil && tt.expected.Stats_conns != nil) || (result.Stats_conns != nil && !reflect.DeepEqual(*result.Stats_conns, *tt.expected.Stats_conns)) {
				t.Errorf("diffRankingsJsonCfg() Stats_conns = %v, want %v", *result.Stats_conns, *tt.expected.Stats_conns)
			}

			if (result.Thresholds_conns == nil && tt.expected.Thresholds_conns != nil) || (result.Thresholds_conns != nil && !reflect.DeepEqual(*result.Thresholds_conns, *tt.expected.Thresholds_conns)) {
				t.Errorf("diffRankingsJsonCfg() Thresholds_conns = %v, want %v", *result.Thresholds_conns, *tt.expected.Thresholds_conns)
			}

			if (result.Store_interval == nil && tt.expected.Store_interval != nil) || (result.Store_interval != nil && *result.Store_interval != *tt.expected.Store_interval) {
				t.Errorf("diffRankingsJsonCfg() Store_interval = %v, want %v", *result.Store_interval, *tt.expected.Store_interval)
			}

			if (result.Ees_conns == nil && tt.expected.Ees_conns != nil) || (result.Ees_conns != nil && !reflect.DeepEqual(*result.Ees_conns, *tt.expected.Ees_conns)) {
				t.Errorf("diffRankingsJsonCfg() Ees_conns = %v, want %v", *result.Ees_conns, *tt.expected.Ees_conns)
			}

			if (result.Ees_exporter_ids == nil && tt.expected.Ees_exporter_ids != nil) || (result.Ees_exporter_ids != nil && !reflect.DeepEqual(*result.Ees_exporter_ids, *tt.expected.Ees_exporter_ids)) {
				t.Errorf("diffRankingsJsonCfg() Ees_exporter_ids = %v, want %v", *result.Ees_exporter_ids, *tt.expected.Ees_exporter_ids)
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
				Enabled:         true,
				StoreInterval:   20 * time.Second,
				StatSConns:      []string{"conn1", "conn2"},
				ThresholdSConns: []string{"thresh1", "thresh2"},
				ScheduledIDs:    map[string][]string{"sched1": {"id1", "id2"}},
				EEsConns:        []string{"ees1", "ees2"},
				EEsExporterIDs:  []string{"exporter1", "exporter2"},
			},
			expected: &RankingSCfg{
				Enabled:         true,
				StoreInterval:   20 * time.Second,
				StatSConns:      []string{"conn1", "conn2"},
				ThresholdSConns: []string{"thresh1", "thresh2"},
				ScheduledIDs:    map[string][]string{"sched1": {"id1", "id2"}},
				EEsConns:        []string{"ees1", "ees2"},
				EEsExporterIDs:  []string{"exporter1", "exporter2"},
			},
		},
		{
			name: "clone with zero and nil fields",
			input: &RankingSCfg{
				Enabled:         false,
				StoreInterval:   0,
				StatSConns:      nil,
				ThresholdSConns: nil,
				ScheduledIDs:    nil,
				EEsConns:        nil,
				EEsExporterIDs:  nil,
			},
			expected: &RankingSCfg{
				Enabled:         false,
				StoreInterval:   0,
				StatSConns:      nil,
				ThresholdSConns: nil,
				ScheduledIDs:    nil,
				EEsConns:        nil,
				EEsExporterIDs:  nil,
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
			if !slices.Equal(clone.StatSConns, tt.expected.StatSConns) {
				t.Errorf("Clone() StatSConns = %v, want %v", clone.StatSConns, tt.expected.StatSConns)
			}
			if !slices.Equal(clone.ThresholdSConns, tt.expected.ThresholdSConns) {
				t.Errorf("Clone() ThresholdSConns = %v, want %v", clone.ThresholdSConns, tt.expected.ThresholdSConns)
			}
			if !reflect.DeepEqual(clone.ScheduledIDs, tt.expected.ScheduledIDs) {
				t.Errorf("Clone() ScheduledIDs = %v, want %v", clone.ScheduledIDs, tt.expected.ScheduledIDs)
			}
			if !slices.Equal(clone.EEsConns, tt.expected.EEsConns) {
				t.Errorf("Clone() EEsConns = %v, want %v", clone.EEsConns, tt.expected.EEsConns)
			}
			if !slices.Equal(clone.EEsExporterIDs, tt.expected.EEsExporterIDs) {
				t.Errorf("Clone() EEsExporterIDs = %v, want %v", clone.EEsExporterIDs, tt.expected.EEsExporterIDs)
			}
			if tt.input.StatSConns != nil && clone.StatSConns != nil && &tt.input.StatSConns[0] == &clone.StatSConns[0] {
				t.Error("Clone() StatSConns has the same reference, expected a deep copy")
			}
			if tt.input.ThresholdSConns != nil && clone.ThresholdSConns != nil && &tt.input.ThresholdSConns[0] == &clone.ThresholdSConns[0] {
				t.Error("Clone() ThresholdSConns has the same reference, expected a deep copy")
			}
			if tt.input.EEsConns != nil && clone.EEsConns != nil && &tt.input.EEsConns[0] == &clone.EEsConns[0] {
				t.Error("Clone() EEsConns has the same reference, expected a deep copy")
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
				t.Errorf("CloneSection() = %v, want %v", clonedRankingSCfg, tt.expected)
			}
			if tt.input.StatSConns != nil && clonedRankingSCfg.StatSConns != nil && &tt.input.StatSConns[0] == &clonedRankingSCfg.StatSConns[0] {
				t.Error("CloneSection() StatSConns has the same reference, expected a deep copy")
			}
			if tt.input.ThresholdSConns != nil && clonedRankingSCfg.ThresholdSConns != nil && &tt.input.ThresholdSConns[0] == &clonedRankingSCfg.ThresholdSConns[0] {
				t.Error("CloneSection() ThresholdSConns has the same reference, expected a deep copy")
			}
			if tt.input.EEsConns != nil && clonedRankingSCfg.EEsConns != nil && &tt.input.EEsConns[0] == &clonedRankingSCfg.EEsConns[0] {
				t.Error("CloneSection() EEsConns has the same reference, expected a deep copy")
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
