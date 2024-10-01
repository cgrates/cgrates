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

func TestLoadFromJSONCfgs(t *testing.T) {
	tests := []struct {
		name     string
		input    *TrendsJsonCfg
		expected *TrendSCfg
	}{
		{
			name: "Test with MetaInternal in Stats_conns",
			input: &TrendsJsonCfg{
				Enabled:          utils.BoolPointer(true),
				Stats_conns:      &[]string{"conn1", utils.MetaInternal},
				Thresholds_conns: &[]string{"thresh1"},
				Scheduled_ids:    map[string][]string{"tenant1": {"id1"}},
			},
			expected: &TrendSCfg{
				Enabled:         true,
				StatSConns:      []string{"conn1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats)},
				ThresholdSConns: []string{"thresh1"},
				ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
			},
		},
		{
			name: "Test with MetaInternal in Thresholds_conns",
			input: &TrendsJsonCfg{
				Enabled:          utils.BoolPointer(true),
				Stats_conns:      &[]string{"conn1"},
				Thresholds_conns: &[]string{"thresh1", utils.MetaInternal},
				Scheduled_ids:    map[string][]string{"tenant1": {"id1"}},
			},
			expected: &TrendSCfg{
				Enabled:         true,
				StatSConns:      []string{"conn1"},
				ThresholdSConns: []string{"thresh1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
				ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
			},
		},
		{
			name: "Test without any connections",
			input: &TrendsJsonCfg{
				Enabled:          utils.BoolPointer(true),
				Stats_conns:      nil,
				Thresholds_conns: nil,
				Scheduled_ids:    nil,
			},
			expected: &TrendSCfg{
				Enabled:         true,
				StatSConns:      nil,
				ThresholdSConns: nil,
				ScheduledIDs:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trendCfg TrendSCfg

			err := trendCfg.loadFromJSONCfg(tt.input)

			if err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
			if trendCfg.Enabled != tt.expected.Enabled {
				t.Errorf("Expected Enabled to be %v, but got: %v", tt.expected.Enabled, trendCfg.Enabled)
			}
			if len(trendCfg.StatSConns) != len(tt.expected.StatSConns) {
				t.Errorf("Expected StatSConns length to be %v, but got: %v", len(tt.expected.StatSConns), len(trendCfg.StatSConns))
			} else {
				for i := range trendCfg.StatSConns {
					if trendCfg.StatSConns[i] != tt.expected.StatSConns[i] {
						t.Errorf("Expected StatSConns[%d] to be %v, but got: %v", i, tt.expected.StatSConns[i], trendCfg.StatSConns[i])
					}
				}
			}
			if len(trendCfg.ThresholdSConns) != len(tt.expected.ThresholdSConns) {
				t.Errorf("Expected ThresholdSConns length to be %v, but got: %v", len(tt.expected.ThresholdSConns), len(trendCfg.ThresholdSConns))
			} else {
				for i := range trendCfg.ThresholdSConns {
					if trendCfg.ThresholdSConns[i] != tt.expected.ThresholdSConns[i] {
						t.Errorf("Expected ThresholdSConns[%d] to be %v, but got: %v", i, tt.expected.ThresholdSConns[i], trendCfg.ThresholdSConns[i])
					}
				}
			}
			if len(trendCfg.ScheduledIDs) != len(tt.expected.ScheduledIDs) {
				t.Errorf("Expected ScheduledIDs length to be %v, but got: %v", len(tt.expected.ScheduledIDs), len(trendCfg.ScheduledIDs))
			} else {
				for key, ids := range tt.expected.ScheduledIDs {
					if len(trendCfg.ScheduledIDs[key]) != len(ids) {
						t.Errorf("Expected ScheduledIDs[%s] length to be %v, but got: %v", key, len(ids), len(trendCfg.ScheduledIDs[key]))
					} else {
						for i := range ids {
							if trendCfg.ScheduledIDs[key][i] != ids[i] {
								t.Errorf("Expected ScheduledIDs[%s][%d] to be %v, but got: %v", key, i, ids[i], trendCfg.ScheduledIDs[key][i])
							}
						}
					}
				}
			}
		})
	}
}

func TestAsMapInterface(t *testing.T) {
	trendCfg := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), "conn1"},
		ThresholdSConns: []string{"thresh1", utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds)},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
	}

	expected := map[string]any{
		utils.EnabledCfg:         true,
		utils.StatSConnsCfg:      []string{utils.MetaInternal, "conn1"},
		utils.ThresholdSConnsCfg: []string{"thresh1", utils.MetaInternal},
		utils.ScheduledIDsCfg:    map[string][]string{"tenant1": {"id1"}},
	}

	got := trendCfg.AsMapInterface()

	if got[utils.EnabledCfg] != expected[utils.EnabledCfg] {
		t.Errorf("Expected Enabled to be %v, but got: %v", expected[utils.EnabledCfg], got[utils.EnabledCfg])
	}

	if len(got[utils.StatSConnsCfg].([]string)) != len(expected[utils.StatSConnsCfg].([]string)) {
		t.Errorf("Expected StatSConns length to be %v, but got: %v", len(expected[utils.StatSConnsCfg].([]string)), len(got[utils.StatSConnsCfg].([]string)))
	} else {
		for i, conn := range expected[utils.StatSConnsCfg].([]string) {
			if got[utils.StatSConnsCfg].([]string)[i] != conn {
				t.Errorf("Expected StatSConns[%d] to be %v, but got: %v", i, conn, got[utils.StatSConnsCfg].([]string)[i])
			}
		}
	}

	if len(got[utils.ThresholdSConnsCfg].([]string)) != len(expected[utils.ThresholdSConnsCfg].([]string)) {
		t.Errorf("Expected ThresholdSConns length to be %v, but got: %v", len(expected[utils.ThresholdSConnsCfg].([]string)), len(got[utils.ThresholdSConnsCfg].([]string)))
	} else {
		for i, thresh := range expected[utils.ThresholdSConnsCfg].([]string) {
			if got[utils.ThresholdSConnsCfg].([]string)[i] != thresh {
				t.Errorf("Expected ThresholdSConns[%d] to be %v, but got: %v", i, thresh, got[utils.ThresholdSConnsCfg].([]string)[i])
			}
		}
	}

	if len(got[utils.ScheduledIDsCfg].(map[string][]string)) != len(expected[utils.ScheduledIDsCfg].(map[string][]string)) {
		t.Errorf("Expected ScheduledIDs length to be %v, but got: %v", len(expected[utils.ScheduledIDsCfg].(map[string][]string)), len(got[utils.ScheduledIDsCfg].(map[string][]string)))
	} else {
		for key, ids := range expected[utils.ScheduledIDsCfg].(map[string][]string) {
			if got[utils.ScheduledIDsCfg].(map[string][]string)[key][0] != ids[0] {
				t.Errorf("Expected ScheduledIDs[%s] to be %v, but got: %v", key, ids, got[utils.ScheduledIDsCfg].(map[string][]string)[key])
			}
		}
	}
}

func TestTrendSCfgClone(t *testing.T) {
	original := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1", "conn2"},
		ThresholdSConns: []string{"thresh1", "thresh2"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id3"}},
	}

	cloned := original.Clone()

	if cloned.Enabled != original.Enabled {
		t.Errorf("Enabled field mismatch: expected %v, got %v", original.Enabled, cloned.Enabled)
	}

	if !reflect.DeepEqual(cloned.StatSConns, original.StatSConns) {
		t.Errorf("StatSConns mismatch: expected %v, got %v", original.StatSConns, cloned.StatSConns)
	}
	if !reflect.DeepEqual(cloned.ThresholdSConns, original.ThresholdSConns) {
		t.Errorf("ThresholdSConns mismatch: expected %v, got %v", original.ThresholdSConns, cloned.ThresholdSConns)
	}

	if !reflect.DeepEqual(cloned.ScheduledIDs, original.ScheduledIDs) {
		t.Errorf("ScheduledIDs mismatch: expected %v, got %v", original.ScheduledIDs, cloned.ScheduledIDs)
	}

	cloned.Enabled = false
	cloned.StatSConns[0] = "modified_conn"
	cloned.ThresholdSConns[0] = "modified_thresh"
	cloned.ScheduledIDs["tenant1"][0] = "modified_id"

	if cloned.Enabled == original.Enabled {
		t.Error("Modifying cloned.Enabled should not affect original.Enabled")
	}
	if reflect.DeepEqual(cloned.StatSConns, original.StatSConns) {
		t.Error("Modifying cloned.StatSConns should not affect original.StatSConns")
	}
	if reflect.DeepEqual(cloned.ThresholdSConns, original.ThresholdSConns) {
		t.Error("Modifying cloned.ThresholdSConns should not affect original.ThresholdSConns")
	}
	if reflect.DeepEqual(cloned.ScheduledIDs, original.ScheduledIDs) {
		t.Error("Modifying cloned.ScheduledIDs should not affect original.ScheduledIDs")
	}
}
