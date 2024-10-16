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

func TestRankingSCfgLoadFromJSONCfg(t *testing.T) {
	tests := []struct {
		name        string
		jsnCfg      *RankingsJsonCfg
		expectedCfg RankingSCfg
		expectErr   bool
	}{
		{
			name:        "nil input",
			jsnCfg:      nil,
			expectedCfg: RankingSCfg{},
			expectErr:   false,
		},
		{
			name: "enabled true, no stats conns",
			jsnCfg: &RankingsJsonCfg{
				Enabled:     utils.BoolPointer(true),
				Stats_conns: nil,
			},
			expectedCfg: RankingSCfg{
				Enabled: true,
			},
			expectErr: false,
		},
		{
			name: "enabled false with stats conns",
			jsnCfg: &RankingsJsonCfg{
				Enabled:     utils.BoolPointer(false),
				Stats_conns: &[]string{"conn1", utils.MetaInternal},
			},
			expectedCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
			},
			expectErr: false,
		},
		{
			name: "thresholds conns with meta internal",
			jsnCfg: &RankingsJsonCfg{
				Thresholds_conns: &[]string{"threshold1", utils.MetaInternal},
			},
			expectedCfg: RankingSCfg{
				ThresholdSConns: []string{"threshold1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
			},
			expectErr: false,
		},
		{
			name: "scheduled IDs",
			jsnCfg: &RankingsJsonCfg{
				Scheduled_ids: map[string][]string{
					"tenant1": {"id1", "id2"},
				},
			},
			expectedCfg: RankingSCfg{
				ScheduledIDs: map[string][]string{
					"tenant1": {"id1", "id2"},
				},
			},
			expectErr: false,
		},
		{
			name: "store interval valid in seconds",
			jsnCfg: &RankingsJsonCfg{
				Store_interval: utils.StringPointer("30s"),
			},
			expectedCfg: RankingSCfg{
				StoreInterval: 30 * time.Second,
			},
			expectErr: false,
		},
		{
			name: "ees conns with meta internal",
			jsnCfg: &RankingsJsonCfg{
				Ees_conns: &[]string{"ees1", utils.MetaInternal},
			},
			expectedCfg: RankingSCfg{
				EEsConns: []string{"ees1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
			},
			expectErr: false,
		},
		{
			name: "ees exporter IDs",
			jsnCfg: &RankingsJsonCfg{
				Ees_exporter_ids: &[]string{"exporter1", "exporter2"},
			},
			expectedCfg: RankingSCfg{
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sgsCfg RankingSCfg
			err := sgsCfg.loadFromJSONCfg(tt.jsnCfg)

			if (err != nil) != tt.expectErr {
				t.Errorf("loadFromJSONCfg() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if !reflect.DeepEqual(sgsCfg, tt.expectedCfg) {
				t.Errorf("loadFromJSONCfg() = %v, want %v", sgsCfg, tt.expectedCfg)
			}
		})
	}
}

func TestRankingSCfgLoadFromJSONCfgStoreInterval(t *testing.T) {
	tests := []struct {
		name          string
		storeInterval string
		expectError   bool
		expectedValue time.Duration
	}{
		{
			name:          "valid duration",
			storeInterval: "5s",
			expectError:   false,
			expectedValue: 5 * time.Second,
		},
		{
			name:          "valid duration with nanoseconds",
			storeInterval: "2.5s",
			expectError:   false,
			expectedValue: 2500 * time.Millisecond,
		},
		{
			name:          "invalid duration",
			storeInterval: "invalid",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsnCfg := &RankingsJsonCfg{
				Store_interval: &tt.storeInterval,
			}
			sgsCfg := &RankingSCfg{}

			err := sgsCfg.loadFromJSONCfg(jsnCfg)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if sgsCfg.StoreInterval != tt.expectedValue {
					t.Errorf("Expected StoreInterval = %v, got %v", tt.expectedValue, sgsCfg.StoreInterval)
				}
			}
		})
	}
}

func TestRankingSCfgAsMapInterface(t *testing.T) {
	tests := []struct {
		name        string
		rankingSCfg RankingSCfg
		expectedMap map[string]any
	}{
		{
			name: "enabled true, no stat conns",
			rankingSCfg: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        true,
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{},
			},
		},
		{
			name: "enabled false, stat conns with MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        false,
				utils.StatSConnsCfg:     []string{"conn1", utils.MetaInternal},
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{},
			},
		},
		{
			name: "stat conns without MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        false,
				utils.StatSConnsCfg:     []string{"conn1", "conn2"},
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{},
			},
		},
		{
			name: "threshold conns with MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:         false,
				ThresholdSConns: []string{"threshold1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:         false,
				utils.ThresholdSConnsCfg: []string{"threshold1", utils.MetaInternal},
				utils.StoreIntervalCfg:   utils.EmptyString,
				utils.EEsExporterIDsCfg:  []string{},
			},
		},
		{
			name: "scheduled IDs",
			rankingSCfg: RankingSCfg{
				Enabled:      false,
				ScheduledIDs: map[string][]string{"schedule1": {"id1", "id2"}},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        false,
				utils.ScheduledIDsCfg:   map[string][]string{"schedule1": {"id1", "id2"}},
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{},
			},
		},
		{
			name: "ees conns with MetaInternal",
			rankingSCfg: RankingSCfg{
				Enabled:  false,
				EEsConns: []string{"ees1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs)},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        false,
				utils.EEsConnsCfg:       []string{"ees1", utils.MetaInternal},
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{},
			},
		},
		{
			name: "ees exporter IDs",
			rankingSCfg: RankingSCfg{
				Enabled:        false,
				EEsExporterIDs: []string{"exp1", "exp2"},
			},
			expectedMap: map[string]any{
				utils.EnabledCfg:        false,
				utils.StoreIntervalCfg:  utils.EmptyString,
				utils.EEsExporterIDsCfg: []string{"exp1", "exp2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rankingSCfg.AsMapInterface()
			if !reflect.DeepEqual(result, tt.expectedMap) {
				t.Errorf("AsMapInterface() = %v, want %v", result, tt.expectedMap)
			}
		})
	}
}

func TestRankingSCfgClone(t *testing.T) {
	tests := []struct {
		name          string
		originalCfg   RankingSCfg
		expectedClone RankingSCfg
	}{
		{
			name: "enabled true, no stat conns",
			originalCfg: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
			expectedClone: RankingSCfg{
				Enabled:    true,
				StatSConns: nil,
			},
		},
		{
			name: "enabled false, stat conns present",
			originalCfg: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
			expectedClone: RankingSCfg{
				Enabled:    false,
				StatSConns: []string{"conn1", "conn2"},
			},
		},
		{
			name: "threshold conns present",
			originalCfg: RankingSCfg{
				ThresholdSConns: []string{"threshold1", "threshold2"},
			},
			expectedClone: RankingSCfg{
				ThresholdSConns: []string{"threshold1", "threshold2"},
			},
		},
		{
			name: "scheduled IDs present",
			originalCfg: RankingSCfg{
				ScheduledIDs: map[string][]string{
					"schedule1": {"id1", "id2"},
				},
			},
			expectedClone: RankingSCfg{
				ScheduledIDs: map[string][]string{
					"schedule1": {"id1", "id2"},
				},
			},
		},
		{
			name: "EEs conns present",
			originalCfg: RankingSCfg{
				EEsConns: []string{"ees1", "ees2"},
			},
			expectedClone: RankingSCfg{
				EEsConns: []string{"ees1", "ees2"},
			},
		},
		{
			name: "EEs exporter IDs present",
			originalCfg: RankingSCfg{
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
			expectedClone: RankingSCfg{
				EEsExporterIDs: []string{"exporter1", "exporter2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clone := tt.originalCfg.Clone()

			if !reflect.DeepEqual(clone, &tt.expectedClone) {
				t.Errorf("Clone() = %v, want %v", clone, &tt.expectedClone)
			}

			if clone.StatSConns != nil && tt.originalCfg.StatSConns != nil && &clone.StatSConns[0] == &tt.originalCfg.StatSConns[0] {
				t.Errorf("StatSConns points to the same slice, expected a deep copy")
			}

			if clone.ThresholdSConns != nil && tt.originalCfg.ThresholdSConns != nil && &clone.ThresholdSConns[0] == &tt.originalCfg.ThresholdSConns[0] {
				t.Errorf("ThresholdSConns points to the same slice, expected a deep copy")
			}

			if clone.ScheduledIDs != nil && tt.originalCfg.ScheduledIDs != nil {
				for key := range clone.ScheduledIDs {
					if &clone.ScheduledIDs[key][0] == &tt.originalCfg.ScheduledIDs[key][0] {
						t.Errorf("ScheduledIDs points to the same slice for key %v, expected a deep copy", key)
					}
				}
			}

			if clone.EEsConns != nil && tt.originalCfg.EEsConns != nil && &clone.EEsConns[0] == &tt.originalCfg.EEsConns[0] {
				t.Errorf("EEsConns points to the same slice, expected a deep copy")
			}

			if clone.EEsExporterIDs != nil && tt.originalCfg.EEsExporterIDs != nil && &clone.EEsExporterIDs[0] == &tt.originalCfg.EEsExporterIDs[0] {
				t.Errorf("EEsExporterIDs points to the same slice, expected a deep copy")
			}
		})
	}
}
