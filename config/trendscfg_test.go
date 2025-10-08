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

func TestTrendSLoadFromJSONCfg_NilConfig(t *testing.T) {
	trendCfg := &TrendSCfg{
		Enabled:         true,
		StatSConns:      []string{"connection1"},
		ThresholdSConns: []string{"threshold1"},
		ScheduledIDs:    map[string][]string{"tenant1": {"id1"}},
		StoreInterval:   10 * time.Second,
		EEsConns:        []string{"conn1"},
		EEsExporterIDs:  []string{"exporter1"},
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
		EEsConns: []string{"old_conn1", "old_conn2"},
	}
	jsnCfg := &TrendSJsonCfg{
		Ees_conns: &[]string{"new_conn1", "new_conn2"},
	}
	err := trendCfg.loadFromJSONCfg(jsnCfg)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	expectedConns := []string{"new_conn1", "new_conn2"}
	if len(trendCfg.EEsConns) != len(expectedConns) {
		t.Errorf("Expected EEsConns length to be %d, but got: %d", len(expectedConns), len(trendCfg.EEsConns))
	}
	for i, conn := range expectedConns {
		if trendCfg.EEsConns[i] != conn {
			t.Errorf("Expected EEsConns[%d] to be %v, but got: %v", i, conn, trendCfg.EEsConns[i])
		}
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
	statSConns := []string{"statConn1"}
	thresholdSConns := []string{"thresholdConn1"}
	scheduledIDs := map[string][]string{"tenant1": {"id1"}}
	eesConns := []string{"eesConn1"}

	trendCfg := TrendSCfg{
		Enabled:                true,
		StatSConns:             statSConns,
		ThresholdSConns:        thresholdSConns,
		ScheduledIDs:           scheduledIDs,
		StoreInterval:          storeInterval,
		StoreUncompressedLimit: storeUncompressedLimit,
		EEsConns:               eesConns,
		EEsExporterIDs:         eesExporterIDs,
	}
	expectedMap := map[string]any{
		utils.EnabledCfg:                true,
		utils.StoreIntervalCfg:          storeInterval.String(),
		utils.StoreUncompressedLimitCfg: storeUncompressedLimit,
		utils.StatSConnsCfg:             stripInternalConns(statSConns),
		utils.ThresholdSConnsCfg:        stripInternalConns(thresholdSConns),
		utils.ScheduledIDsCfg:           scheduledIDs,
		utils.EEsConnsCfg:               stripInternalConns(eesConns),
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
		Enabled:                true,
		StatSConns:             []string{"conn1", "conn2"},
		ThresholdSConns:        []string{"thresh1", "thresh2"},
		ScheduledIDs:           map[string][]string{"tenant1": {"id1", "id2"}, "tenant2": {"id3"}},
		StoreInterval:          30 * time.Second,
		StoreUncompressedLimit: 500,
		EEsConns:               []string{"eeconn1", "eeconn2"},
		EEsExporterIDs:         []string{"exporter1", "exporter2"},
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
	if cloned.StoreInterval != original.StoreInterval {
		t.Errorf("StoreInterval mismatch: expected %v, got %v", original.StoreInterval, cloned.StoreInterval)
	}
	if cloned.StoreUncompressedLimit != original.StoreUncompressedLimit {
		t.Errorf("StoreUncompressedLimit mismatch: expected %v, got %v", original.StoreUncompressedLimit, cloned.StoreUncompressedLimit)
	}
	if !reflect.DeepEqual(cloned.EEsConns, original.EEsConns) {
		t.Errorf("EEsConns mismatch: expected %v, got %v", original.EEsConns, cloned.EEsConns)
	}
	if !reflect.DeepEqual(cloned.EEsExporterIDs, original.EEsExporterIDs) {
		t.Errorf("EEsExporterIDs mismatch: expected %v, got %v", original.EEsExporterIDs, cloned.EEsExporterIDs)
	}

	cloned.Enabled = false
	cloned.StatSConns[0] = "modified_conn"
	cloned.ThresholdSConns[0] = "modified_thresh"
	cloned.ScheduledIDs["tenant1"][0] = "modified_id"
	cloned.StoreInterval = 45 * time.Second
	cloned.StoreUncompressedLimit = 1000
	cloned.EEsConns[0] = "modified_eeconn"
	cloned.EEsExporterIDs[0] = "modified_exporter"

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
	if cloned.StoreInterval == original.StoreInterval {
		t.Error("Modifying cloned.StoreInterval should not affect original.StoreInterval")
	}
	if cloned.StoreUncompressedLimit == original.StoreUncompressedLimit {
		t.Error("Modifying cloned.StoreUncompressedLimit should not affect original.StoreUncompressedLimit")
	}
	if reflect.DeepEqual(cloned.EEsConns, original.EEsConns) {
		t.Error("Modifying cloned.EEsConns should not affect original.EEsConns")
	}
	if reflect.DeepEqual(cloned.EEsExporterIDs, original.EEsExporterIDs) {
		t.Error("Modifying cloned.EEsExporterIDs should not affect original.EEsExporterIDs")
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
		EEsConns: []string{"conn1", "conn2"},
	}
	v2 := &TrendSCfg{
		EEsConns: []string{"conn3", "conn4"},
	}
	expected := &TrendSJsonCfg{
		Ees_conns: utils.SliceStringPointer(([]string{"conn3", "conn4"})),
	}
	d := diffTrendsJsonCfg(nil, v1, v2)
	if !reflect.DeepEqual(d.Ees_conns, expected.Ees_conns) {
		t.Errorf("Ees_conns mismatch. Got: %v, expected: %v", d.Ees_conns, expected.Ees_conns)
	}
	v2.EEsConns = []string{"conn1", "conn2"}
	expected.Ees_conns = nil
	d = diffTrendsJsonCfg(nil, v1, v2)
	if d.Ees_conns != nil {
		t.Errorf("Expected Ees_conns to be nil, but got: %v", d.Ees_conns)
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
