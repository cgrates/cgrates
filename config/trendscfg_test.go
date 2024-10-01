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
	"errors"
	"reflect"
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestTrendSLoadFromJSONCfg(t *testing.T) {
	trendCfg := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"connection1"},
		ThresholdSConns: []string{"threshold1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
	}

	err := trendCfg.loadFromJSONCfg(nil)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	if trendCfg.Enabled != true {
		t.Errorf("Expected Enabled to be true, but got: %v", trendCfg.Enabled)
	}
	if len(trendCfg.StatSConns) != 1 || trendCfg.StatSConns[0] != "connection1" {
		t.Errorf("Expected StatSConns to be unchanged, but got: %v", trendCfg.StatSConns)
	}
	if len(trendCfg.ThresholdSConns) != 1 || trendCfg.ThresholdSConns[0] != "threshold1" {
		t.Errorf("Expected ThresholdSConns to be unchanged, but got: %v", trendCfg.ThresholdSConns)
	}
	if len(trendCfg.ScheduledIDs) != 1 || trendCfg.ScheduledIDs["tenant1"][0] != "id1" {
		t.Errorf("Expected ScheduledIDs to be unchanged, but got: %v", trendCfg.ScheduledIDs)
	}
}

func TestScheduledIDsDiffTrendSJsonCfg(t *testing.T) {
	v1 := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1"},
		ThresholdSConns: []string{"thresh1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
	}
	v2 := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1"},
		ThresholdSConns: []string{"thresh1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id2"}},
	}
	expected := &TrendSJsonCfg{
		Scheduled_ids: map[string][]string{"tenant1": {"id2"}},
	}

	var d *TrendSJsonCfg
	got := diffTrendsJsonCfg(d, v1, v2)

	if (got.Scheduled_ids == nil && len(expected.Scheduled_ids) != 0) || !reflect.DeepEqual(got.Scheduled_ids, expected.Scheduled_ids) {
		t.Errorf("Scheduled_ids mismatch. Got: %v, expected: %v", got.Scheduled_ids, expected.Scheduled_ids)
	}
}

func TestDiffTrendSJsonCfg(t *testing.T) {
	v1 := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1"},
		ThresholdSConns: []string{"thresh1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
	}
	v2 := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1"},
		ThresholdSConns: []string{"thresh1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
	}
	expected := &TrendSJsonCfg{
		Enabled:          nil,
		Stats_conns:      nil,
		Thresholds_conns: nil,
		Scheduled_ids:    nil,
	}

	d := diffTrendsJsonCfg(nil, v1, v2)

	if d.Enabled != expected.Enabled {
		t.Errorf("Enabled mismatch. Got: %v, expected: %v", d.Enabled, expected.Enabled)
	}
	if !reflect.DeepEqual(d.Stats_conns, expected.Stats_conns) {
		t.Errorf("Stats_conns mismatch. Got: %v, expected: %v", d.Stats_conns, expected.Stats_conns)
	}
	if !reflect.DeepEqual(d.Thresholds_conns, expected.Thresholds_conns) {
		t.Errorf("Thresholds_conns mismatch. Got: %v, expected: %v", d.Thresholds_conns, expected.Thresholds_conns)
	}

	v2.Enabled = false
	expected = &TrendSJsonCfg{
		Enabled: utils.BoolPointer(false),
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if d.Enabled == nil || *d.Enabled != *expected.Enabled {
		t.Errorf("Enabled mismatch. Got: %v, expected: %v", d.Enabled, expected.Enabled)
	}

	v2.Enabled = true
	v2.StatSConns = []string{"conn2"}
	expected = &TrendSJsonCfg{
		Stats_conns: utils.SliceStringPointer([]string{"conn2"}),
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if !reflect.DeepEqual(d.Stats_conns, expected.Stats_conns) {
		t.Errorf("Stats_conns mismatch. Got: %v, expected: %v", d.Stats_conns, expected.Stats_conns)
	}

	v2.StatSConns = []string{"conn1"}
	v2.ThresholdSConns = []string{"thresh2"}
	expected = &TrendSJsonCfg{
		Thresholds_conns: utils.SliceStringPointer([]string{"thresh2"}),
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if !reflect.DeepEqual(d.Thresholds_conns, expected.Thresholds_conns) {
		t.Errorf("Thresholds_conns mismatch. Got: %v, expected: %v", d.Thresholds_conns, expected.Thresholds_conns)
	}

	v2.ThresholdSConns = []string{"thresh1"}
	v2.ScheduledIDs = map[string][]string{"tenant1": {"id2"}}
	expected = &TrendSJsonCfg{
		Scheduled_ids: map[string][]string{"tenant1": {"id2"}},
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if (d.Scheduled_ids == nil && len(expected.Scheduled_ids) != 0) ||
		(d.Scheduled_ids != nil && !reflect.DeepEqual(d.Scheduled_ids, expected.Scheduled_ids)) {
		t.Errorf("Scheduled_ids mismatch. Got: %v, expected: %v", d.Scheduled_ids, expected.Scheduled_ids)
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

	if cloned == original {
		t.Error("Clone should return a different instance, but got the same")
	}

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

func TestTrendSCfgCloneSection(t *testing.T) {
	original := TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"conn1", "conn2"},
		ThresholdSConns: []string{"thresh1", "thresh2"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id3"}},
	}

	clonedSection := original.CloneSection()

	if clonedSection == nil {
		t.Fatal("CloneSection returned nil")
	}

	if _, ok := clonedSection.(*TrendSCfg); !ok {
		t.Fatalf("CloneSection should return a Section interface, but got %T", clonedSection)
	}

	cloned := clonedSection.(*TrendSCfg)

	if cloned == &original {
		t.Error("CloneSection should return a different instance, but got the same")
	}

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

func TestTrendSCfgLoadError(t *testing.T) {
	ctx := context.Background()
	mockDB := &mockDb{
		GetSectionF: func(ctx *context.Context, section string, val any) error {
			return errors.New("mock error")
		},
	}
	trendSCfg := &TrendSCfg{}
	err := trendSCfg.Load(ctx, mockDB, nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	expectedErr := "mock error"
	if err.Error() != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
