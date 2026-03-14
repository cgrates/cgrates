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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

func TestScheduledIDsDiffTrendSJsonCfg(t *testing.T) {
	v1 := &TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
		},
		ScheduledIDs: map[string][]string{"tenant1": {"id1"}},
	}
	v2 := &TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
		},
		ScheduledIDs: map[string][]string{"tenant1": {"id2"}},
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

func TestTrendSLoadFromJSONCfg_NilConfig(t *testing.T) {
	trendCfg := &TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"connection1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"threshold1"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"conn1"}}},
		},
		ScheduledIDs:   map[string][]string{"tenant1": {"id1"}},
		StoreInterval:  10 * time.Second,
		EEsExporterIDs: []string{"exporter1"},
	}
	err := trendCfg.loadFromJSONCfg(nil)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

}

func TestLoadFromJSONCfgStoreInterval(t *testing.T) {
	validInterval := "30"
	jsnCfgValid := &TrendSJsonCfg{
		Store_interval: &validInterval,
	}
	trendCfg := &TrendSCfg{}
	err := trendCfg.loadFromJSONCfg(jsnCfgValid)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if trendCfg.StoreInterval != 30*time.Nanosecond {
		t.Errorf("Expected StoreInterval to be 30ns, but got: %v", trendCfg.StoreInterval)
	}
	invalidInterval := "invalid_duration"
	jsnCfgInvalid := &TrendSJsonCfg{
		Store_interval: &invalidInterval,
	}
	trendCfg = &TrendSCfg{}
	err = trendCfg.loadFromJSONCfg(jsnCfgInvalid)
	if err == nil {
		t.Errorf("Expected an error, got none")
	}
}

func TestTrendSLoadFromJSONCfgEesConns(t *testing.T) {
	trendCfg := &TrendSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaEEs: {{ConnIDs: []string{"old_conn1", "old_conn2"}}},
		},
	}
	jsnCfg := &TrendSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaEEs: {{ConnIDs: []string{"new_conn1", "new_conn2"}}},
		},
	}
	err := trendCfg.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	// Loading replaces existing conn entries with new ones
	if len(trendCfg.Conns[utils.MetaEEs]) != 1 {
		t.Errorf("Expected EEs conns slice length to be 1, but got: %d", len(trendCfg.Conns[utils.MetaEEs]))
	}
	expectedFirst := []string{"new_conn1", "new_conn2"}
	if !reflect.DeepEqual(trendCfg.Conns[utils.MetaEEs][0].ConnIDs, expectedFirst) {
		t.Errorf("Expected first entry %v, but got: %v", expectedFirst, trendCfg.Conns[utils.MetaEEs][0].ConnIDs)
	}
}

func TestTrendSLoadFromJSONCfgEesExporterIDs(t *testing.T) {
	trendCfg := &TrendSCfg{
		EEsExporterIDs: []string{"old_exporter1", "old_exporter2"},
	}
	jsnCfg := &TrendSJsonCfg{
		Ees_exporter_ids: &[]string{"new_exporter1", "new_exporter2"},
	}
	err := trendCfg.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	expectedIDs := []string{"old_exporter1", "old_exporter2", "new_exporter1", "new_exporter2"}
	if len(trendCfg.EEsExporterIDs) != len(expectedIDs) {
		t.Errorf("Expected EEsExporterIDs length to be %d, but got: %d", len(expectedIDs), len(trendCfg.EEsExporterIDs))
	}
	for i, id := range expectedIDs {
		if trendCfg.EEsExporterIDs[i] != id {
			t.Errorf("Expected EEsExporterIDs[%d] to be %v, but got: %v", i, id, trendCfg.EEsExporterIDs[i])
		}
	}
}

func TestTrendSCfgAsMapInterface(t *testing.T) {
	storeInterval := 10 * time.Second
	storeUncompressedLimit := 500
	eesExporterIDs := []string{"exporter1", "exporter2"}
	scheduledIDs := map[string][]string{"tenant1": {"id1"}}

	conns := map[string][]*DynamicConns{
		utils.MetaStats:      {{ConnIDs: []string{"statConn1"}}},
		utils.MetaThresholds: {{ConnIDs: []string{"thresholdConn1"}}},
		utils.MetaEEs:        {{ConnIDs: []string{"eesConn1"}}},
	}
	trendCfg := TrendSCfg{
		Enabled:                true,
		Conns:                  conns,
		ScheduledIDs:           scheduledIDs,
		StoreInterval:          storeInterval,
		StoreUncompressedLimit: storeUncompressedLimit,
		EEsExporterIDs:         eesExporterIDs,
	}
	expectedMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.StoreIntervalCfg:          storeInterval.String(),
		utils.StoreUncompressedLimitCfg: storeUncompressedLimit,
		utils.ConnsCfg:                  stripConns(conns),
		utils.ScheduledIDsCfg:           scheduledIDs,
		utils.EEsExporterIDsCfg:         eesExporterIDs,
	}
	result := trendCfg.AsMapInterface().(map[string]any)
	if !reflect.DeepEqual(result, expectedMap) {
		t.Errorf("Expected: %+v, got: %+v", expectedMap, result)
	}
	trendCfg.StoreInterval = 0
	expectedMap[utils.StoreIntervalCfg] = utils.EmptyString
	result = trendCfg.AsMapInterface().(map[string]any)
	if result[utils.StoreIntervalCfg] != utils.EmptyString {
		t.Errorf("Expected StoreInterval to be '%s', but got: %v", utils.EmptyString, result[utils.StoreIntervalCfg])
	}
	if result[utils.StoreUncompressedLimitCfg] != storeUncompressedLimit {
		t.Errorf("Expected StoreUncompressedLimit to be %d, but got: %v", storeUncompressedLimit, result[utils.StoreUncompressedLimitCfg])
	}
}

func TestDiffTrendSJsonCfg(t *testing.T) {
	v1 := &TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
		},
		ScheduledIDs: map[string][]string{"tenant1": {"id1"}},
	}
	v2 := &TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
		},
		ScheduledIDs: map[string][]string{"tenant1": {"id1"}},
	}
	expected := &TrendSJsonCfg{
		Enabled:       nil,
		Conns:         nil,
		Scheduled_ids: nil,
	}

	d := diffTrendsJsonCfg(nil, v1, v2)

	if d.Enabled != expected.Enabled {
		t.Errorf("Enabled mismatch. Got: %v, expected: %v", d.Enabled, expected.Enabled)
	}
	if !reflect.DeepEqual(d.Conns, expected.Conns) {
		t.Errorf("Conns mismatch. Got: %v, expected: %v", d.Conns, expected.Conns)
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
	v2.Conns = map[string][]*DynamicConns{
		utils.MetaStats:      {{ConnIDs: []string{"conn2"}}},
		utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
	}
	expected = &TrendSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn2"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
		},
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if !reflect.DeepEqual(d.Conns, expected.Conns) {
		t.Errorf("Conns mismatch. Got: %v, expected: %v", d.Conns, expected.Conns)
	}

	v2.Conns = map[string][]*DynamicConns{
		utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
		utils.MetaThresholds: {{ConnIDs: []string{"thresh2"}}},
	}
	expected = &TrendSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh2"}}},
		},
	}

	d = diffTrendsJsonCfg(nil, v1, v2)

	if !reflect.DeepEqual(d.Conns, expected.Conns) {
		t.Errorf("Conns mismatch. Got: %v, expected: %v", d.Conns, expected.Conns)
	}

	v2.Conns = map[string][]*DynamicConns{
		utils.MetaStats:      {{ConnIDs: []string{"conn1"}}},
		utils.MetaThresholds: {{ConnIDs: []string{"thresh1"}}},
	}
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
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1", "conn2"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1", "thresh2"}}},
			utils.MetaEEs:        {{ConnIDs: []string{"eeconn1", "eeconn2"}}},
		},
		ScheduledIDs:           map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id3"}},
		StoreInterval:          30 * time.Second,
		StoreUncompressedLimit: 500,
		EEsExporterIDs:         []string{"exporter1", "exporter2"},
	}

	cloned := original.Clone()

	if cloned.Enabled != original.Enabled {
		t.Errorf("Enabled field mismatch: expected %v, got %v", original.Enabled, cloned.Enabled)
	}
	if !reflect.DeepEqual(cloned.Conns, original.Conns) {
		t.Errorf("Conns mismatch: expected %v, got %v", original.Conns, cloned.Conns)
	}
	if !reflect.DeepEqual(cloned.ScheduledIDs, original.ScheduledIDs) {
		t.Errorf("ScheduledIDs mismatch: expected %v, got %v", original.ScheduledIDs, cloned.ScheduledIDs)
	}
	if cloned.StoreInterval != original.StoreInterval {
		t.Errorf("StoreInterval mismatch: expected %v, got %v", original.StoreInterval, cloned.StoreInterval)
	}
	if cloned.StoreUncompressedLimit != original.StoreUncompressedLimit {
		t.Errorf("StoreUncompressedLimit mismatch: expected %v, got %v", original.StoreUncompressedLimit, cloned.StoreUncompressedLimit)
	}
	if !reflect.DeepEqual(cloned.EEsExporterIDs, original.EEsExporterIDs) {
		t.Errorf("EEsExporterIDs mismatch: expected %v, got %v", original.EEsExporterIDs, cloned.EEsExporterIDs)
	}

	cloned.Enabled = false
	cloned.Conns[utils.MetaStats][0].ConnIDs[0] = "modified_conn"
	cloned.Conns[utils.MetaThresholds][0].ConnIDs[0] = "modified_thresh"
	cloned.ScheduledIDs["tenant1"][0] = "modified_id"
	cloned.StoreInterval = 45 * time.Second
	cloned.StoreUncompressedLimit = 1000
	cloned.Conns[utils.MetaEEs][0].ConnIDs[0] = "modified_eeconn"
	cloned.EEsExporterIDs[0] = "modified_exporter"

	if cloned.Enabled == original.Enabled {
		t.Error("Modifying cloned.Enabled should not affect original.Enabled")
	}
	if reflect.DeepEqual(cloned.Conns, original.Conns) {
		t.Error("Modifying cloned.Conns should not affect original.Conns")
	}
	if reflect.DeepEqual(cloned.ScheduledIDs, original.ScheduledIDs) {
		t.Error("Modifying cloned.ScheduledIDs should not affect original.ScheduledIDs")
	}
	if cloned.StoreInterval == original.StoreInterval {
		t.Error("Modifying cloned.StoreInterval should not affect original.StoreInterval")
	}
	if cloned.StoreUncompressedLimit == original.StoreUncompressedLimit {
		t.Error("Modifying cloned.StoreUncompressedLimit should not affect original.StoreUncompressedLimit")
	}
	if reflect.DeepEqual(cloned.EEsExporterIDs, original.EEsExporterIDs) {
		t.Error("Modifying cloned.EEsExporterIDs should not affect original.EEsExporterIDs")
	}
}

func TestTrendSCfgCloneSection(t *testing.T) {
	original := TrendSCfg{
		Enabled: true,
		Conns: map[string][]*DynamicConns{
			utils.MetaStats:      {{ConnIDs: []string{"conn1", "conn2"}}},
			utils.MetaThresholds: {{ConnIDs: []string{"thresh1", "thresh2"}}},
		},
		ScheduledIDs: map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id3"}},
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

	if !reflect.DeepEqual(cloned.Conns, original.Conns) {
		t.Errorf("Conns mismatch: expected %v, got %v", original.Conns, cloned.Conns)
	}

	if !reflect.DeepEqual(cloned.ScheduledIDs, original.ScheduledIDs) {
		t.Errorf("ScheduledIDs mismatch: expected %v, got %v", original.ScheduledIDs, cloned.ScheduledIDs)
	}

	cloned.Enabled = false
	cloned.Conns[utils.MetaStats][0].ConnIDs[0] = "modified_conn"
	cloned.Conns[utils.MetaThresholds][0].ConnIDs[0] = "modified_thresh"
	cloned.ScheduledIDs["tenant1"][0] = "modified_id"

	if cloned.Enabled == original.Enabled {
		t.Error("Modifying cloned.Enabled should not affect original.Enabled")
	}
	if reflect.DeepEqual(cloned.Conns, original.Conns) {
		t.Error("Modifying cloned.Conns should not affect original.Conns")
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

func TestDiffTrendsJsonCfgStoreInterval(t *testing.T) {
	v1 := &TrendSCfg{
		StoreInterval: 0,
	}
	v2 := &TrendSCfg{
		StoreInterval: 2 * time.Second,
	}
	expected := &TrendSJsonCfg{
		Store_interval: utils.StringPointer("2s"),
	}
	d := diffTrendsJsonCfg(nil, v1, v2)
	if !reflect.DeepEqual(d.Store_interval, expected.Store_interval) {
		t.Errorf("Store_interval mismatch. Got: %v, expected: %v", d.Store_interval, expected.Store_interval)
	}
	v2.StoreInterval = 0
	expected.Store_interval = nil
	d = diffTrendsJsonCfg(nil, v1, v2)
	if d.Store_interval != nil {
		t.Errorf("Expected Store_interval to be nil, but got: %v", d.Store_interval)
	}
}

func TestDiffTrendsJsonCfgEEsConns(t *testing.T) {
	v1 := &TrendSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaEEs: {{ConnIDs: []string{"conn1", "conn2"}}},
		},
	}
	v2 := &TrendSCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaEEs: {{ConnIDs: []string{"conn3", "conn4"}}},
		},
	}
	expected := &TrendSJsonCfg{
		Conns: map[string][]*DynamicConns{
			utils.MetaEEs: {{ConnIDs: []string{"conn3", "conn4"}}},
		},
	}
	d := diffTrendsJsonCfg(nil, v1, v2)
	if !reflect.DeepEqual(d.Conns, expected.Conns) {
		t.Errorf("Conns mismatch. Got: %v, expected: %v", d.Conns, expected.Conns)
	}
	v2.Conns = map[string][]*DynamicConns{
		utils.MetaEEs: {{ConnIDs: []string{"conn1", "conn2"}}},
	}
	expected.Conns = nil
	d = diffTrendsJsonCfg(nil, v1, v2)
	if d.Conns != nil {
		t.Errorf("Expected Conns to be nil, but got: %v", d.Conns)
	}
}

func TestDiffTrendsJsonCfgEEsExporterIDs(t *testing.T) {
	v1 := &TrendSCfg{
		EEsExporterIDs: []string{"exporter1", "exporter2"},
	}
	v2 := &TrendSCfg{
		EEsExporterIDs: []string{"exporter3", "exporter4"},
	}
	expected := &TrendSJsonCfg{
		Ees_exporter_ids: &[]string{"exporter3", "exporter4"},
	}
	d := diffTrendsJsonCfg(nil, v1, v2)
	if !reflect.DeepEqual(d.Ees_exporter_ids, expected.Ees_exporter_ids) {
		t.Errorf("Ees_exporter_ids mismatch. Got: %v, expected: %v", d.Ees_exporter_ids, expected.Ees_exporter_ids)
	}
	v2.EEsExporterIDs = []string{"exporter1", "exporter2"}
	expected.Ees_exporter_ids = nil
	d = diffTrendsJsonCfg(nil, v1, v2)
	if d.Ees_exporter_ids != nil {
		t.Errorf("Expected Ees_exporter_ids to be nil, but got: %v", d.Ees_exporter_ids)
	}
}

func TestTrendSCfgLoadFromJSONCfgStoreUncompressedLimit(t *testing.T) {
	tests := []struct {
		name          string
		initialLimit  int
		jsonLimit     *int
		expectedLimit int
	}{
		{
			name:          "StoreUncompressedLimit is set",
			initialLimit:  0,
			jsonLimit:     utils.IntPointer(1000),
			expectedLimit: 1000,
		},
		{
			name:          "StoreUncompressedLimit is nil",
			initialLimit:  500,
			jsonLimit:     nil,
			expectedLimit: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trendSCfg := &TrendSCfg{
				StoreUncompressedLimit: tt.initialLimit,
			}

			jsnCfg := &TrendSJsonCfg{
				Store_uncompressed_limit: tt.jsonLimit,
			}

			err := trendSCfg.loadFromJSONCfg(jsnCfg)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if trendSCfg.StoreUncompressedLimit != tt.expectedLimit {
				t.Errorf("Expected StoreUncompressedLimit to be %d, got %d", tt.expectedLimit, trendSCfg.StoreUncompressedLimit)
			}
		})
	}
}

func TestDiffTrendsJsonCfgStoreUncompressedLimit(t *testing.T) {
	v1 := &TrendSCfg{
		StoreUncompressedLimit: 500,
	}
	v2 := &TrendSCfg{
		StoreUncompressedLimit: 1000,
	}
	diff := diffTrendsJsonCfg(nil, v1, v2)
	if diff.Store_uncompressed_limit == nil {
		t.Error("Expected Store_uncompressed_limit to be set, but got nil")
	} else if *diff.Store_uncompressed_limit != v2.StoreUncompressedLimit {
		t.Errorf("Store_uncompressed_limit mismatch: expected %v, got %v", v2.StoreUncompressedLimit, *diff.Store_uncompressed_limit)
	}
	v2.StoreUncompressedLimit = 500
	diff = diffTrendsJsonCfg(nil, v1, v2)
	if diff.Store_uncompressed_limit != nil {
		t.Errorf("Expected Store_uncompressed_limit to be nil when limits are the same, but got %v", *diff.Store_uncompressed_limit)
	}
}
