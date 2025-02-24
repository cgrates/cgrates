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

package engine

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestTrendProfileClone(t *testing.T) {

	original := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "ID",
		Schedule:        "Schedule",
		StatID:          "StatID",
		Metrics:         []string{"metric1", "metric2"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"thresh1", "thresh2"},
	}

	cloned := original.Clone()

	if cloned.Tenant != original.Tenant {
		t.Errorf("Expected Tenant %s, but got %s", original.Tenant, cloned.Tenant)
	}
	if cloned.ID != original.ID {
		t.Errorf("Expected ID %s, but got %s", original.ID, cloned.ID)
	}
	if cloned.Schedule != original.Schedule {
		t.Errorf("Expected Schedule %s, but got %s", original.Schedule, cloned.Schedule)
	}
	if cloned.StatID != original.StatID {
		t.Errorf("Expected StatID %s, but got %s", original.StatID, cloned.StatID)
	}
	if cloned.QueueLength != original.QueueLength {
		t.Errorf("Expected QueueLength %d, but got %d", original.QueueLength, cloned.QueueLength)
	}
	if cloned.TTL != original.TTL {
		t.Errorf("Expected TTL %v, but got %v", original.TTL, cloned.TTL)
	}
	if cloned.MinItems != original.MinItems {
		t.Errorf("Expected MinItems %d, but got %d", original.MinItems, cloned.MinItems)
	}
	if cloned.CorrelationType != original.CorrelationType {
		t.Errorf("Expected CorrelationType %s, but got %s", original.CorrelationType, cloned.CorrelationType)
	}
	if cloned.Tolerance != original.Tolerance {
		t.Errorf("Expected Tolerance %f, but got %f", original.Tolerance, cloned.Tolerance)
	}
	if cloned.Stored != original.Stored {
		t.Errorf("Expected Stored %v, but got %v", original.Stored, cloned.Stored)
	}

	if !reflect.DeepEqual(cloned.Metrics, original.Metrics) {
		t.Errorf("Expected Metrics %v, but got %v", original.Metrics, cloned.Metrics)
	}
	if !reflect.DeepEqual(cloned.ThresholdIDs, original.ThresholdIDs) {
		t.Errorf("Expected ThresholdIDs %v, but got %v", original.ThresholdIDs, cloned.ThresholdIDs)
	}

	if len(cloned.Metrics) > 0 && &cloned.Metrics[0] == &original.Metrics[0] {
		t.Errorf("Metrics slice was not deep copied")
	}
	if len(cloned.ThresholdIDs) > 0 && &cloned.ThresholdIDs[0] == &original.ThresholdIDs[0] {
		t.Errorf("ThresholdIDs slice was not deep copied")
	}
}

func TestTrendProfileTenantIDAndTrendProfileWithAPIOpts(t *testing.T) {

	tp := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "trend1",
		Schedule:        "*/5 * * * *",
		StatID:          "StatID",
		Metrics:         []string{"metric1", "metric2"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"thresh1", "thresh2"},
	}

	tenantID := tp.TenantID()

	expectedTenantID := "cgrates.org" + utils.ConcatenatedKeySep + "trend1"
	if tenantID != expectedTenantID {
		t.Errorf("Expected TenantID %s, but got %s", expectedTenantID, tenantID)
	}

	apiOpts := map[string]any{
		"option1": "value1",
		"option2": 42,
	}

	tpWithAPIOpts := &TrendProfileWithAPIOpts{
		TrendProfile: tp,
		APIOpts:      apiOpts,
	}

	if tpWithAPIOpts.Tenant != "cgrates.org" {
		t.Errorf("Expected Tenant %s, but got %s", "cgrates.org", tpWithAPIOpts.Tenant)
	}
	if tpWithAPIOpts.ID != "trend1" {
		t.Errorf("Expected ID %s, but got %s", "trend1", tpWithAPIOpts.ID)
	}

	expectedAPIOpts := map[string]any{
		"option1": "value1",
		"option2": 42,
	}
	if !reflect.DeepEqual(tpWithAPIOpts.APIOpts, expectedAPIOpts) {
		t.Errorf("Expected APIOpts %v, but got %v", expectedAPIOpts, tpWithAPIOpts.APIOpts)
	}

}

func TestIndexesAppendMetric(t *testing.T) {

	trend := &Trend{
		mLast:   make(map[string]time.Time),
		mCounts: make(map[string]int),
		mTotals: make(map[string]float64),
	}

	metric1 := &MetricWithTrend{ID: "metric1", Value: 5.0}
	metric2 := &MetricWithTrend{ID: "metric2", Value: 3.0}

	rTime1 := time.Now()
	rTime2 := rTime1.Add(10 * time.Minute)

	trend.indexesAppendMetric(metric1, rTime1)
	trend.indexesAppendMetric(metric2, rTime2)
	trend.indexesAppendMetric(metric1, rTime2)

	expectedMLast := map[string]time.Time{
		"metric1": rTime2,
		"metric2": rTime2,
	}
	if !reflect.DeepEqual(trend.mLast, expectedMLast) {
		t.Errorf("Expected mLast %v, but got %v", expectedMLast, trend.mLast)
	}

	expectedMCounts := map[string]int{
		"metric1": 2,
		"metric2": 1,
	}
	if !reflect.DeepEqual(trend.mCounts, expectedMCounts) {
		t.Errorf("Expected mCounts %v, but got %v", expectedMCounts, trend.mCounts)
	}

	expectedMTotals := map[string]float64{
		"metric1": 10.0,
		"metric2": 3.0,
	}
	if !reflect.DeepEqual(trend.mTotals, expectedMTotals) {
		t.Errorf("Expected mTotals %v, but got %v", expectedMTotals, trend.mTotals)
	}
}

func TestTrendTenantID(t *testing.T) {
	trend := &Trend{
		Tenant: "cgrates.org",
		ID:     "ID",
		RunTimes: []time.Time{
			time.Now(),
			time.Now().Add(-1 * time.Hour),
		},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			time.Now(): {
				"metric1": {ID: "metric1", Value: 1.5},
				"metric2": {ID: "metric2", Value: 2.0},
			},
			time.Now().Add(-1 * time.Hour): {
				"metric1": {ID: "metric1", Value: 1.0},
			},
		},
		CompressedMetrics: []byte{0x00, 0x01},
		mLast: map[string]time.Time{
			"metric1": time.Now(),
			"metric2": time.Now().Add(-1 * time.Hour),
		},
		mCounts: map[string]int{
			"metric1": 2,
			"metric2": 1,
		},
		mTotals: map[string]float64{
			"metric1": 2.5,
			"metric2": 2.0,
		},
		tPrfl: &TrendProfile{
			Tenant:          "cgrates.org",
			ID:              "trendProfileID",
			Schedule:        "0 * * * *",
			StatID:          "statID1",
			QueueLength:     10,
			TTL:             5 * time.Minute,
			MinItems:        1,
			CorrelationType: "average",
			Tolerance:       0.1,
			Stored:          true,
			ThresholdIDs:    []string{"threshold1", "threshold2"},
		},
	}

	tenantID := trend.TenantID()

	expectedTenantID := "cgrates.org:ID"
	if tenantID != expectedTenantID {
		t.Errorf("Expected TenantID %v, but got %v", expectedTenantID, tenantID)
	}

	if len(trend.RunTimes) != 2 {
		t.Errorf("Expected 2 run times, but got %d", len(trend.RunTimes))
	}

	if len(trend.Metrics) != 2 {
		t.Errorf("Expected 2 metrics time entries, but got %d", len(trend.Metrics))
	}

	if trend.tPrfl.QueueLength != 10 {
		t.Errorf("Expected QueueLength 10, but got %d", trend.tPrfl.QueueLength)
	}
}

func TestTComputeIndexes(t *testing.T) {
	runTime1 := time.Now()
	runTime2 := runTime1.Add(time.Minute)

	trend := &Trend{
		RunTimes: []time.Time{runTime1, runTime2},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			runTime1: {
				"metric1": &MetricWithTrend{ID: "metric1", Value: 10.0},
				"metric2": &MetricWithTrend{ID: "metric2", Value: 20.0},
			},
			runTime2: {
				"metric1": &MetricWithTrend{ID: "metric1", Value: 15.0},
			},
		},
	}

	trend.computeIndexes()

	if trend.mLast["metric1"] != runTime2 {
		t.Errorf("Expected last time for metric1 to be %v, got %v", runTime2, trend.mLast["metric1"])
	}

	if trend.mCounts["metric1"] != 2 {
		t.Errorf("Expected count for metric1 to be 2, got %d", trend.mCounts["metric1"])
	}

	if trend.mTotals["metric1"] != 25.0 {
		t.Errorf("Expected total for metric1 to be 25.0, got %f", trend.mTotals["metric1"])
	}

	if trend.mLast["metric2"] != runTime1 {
		t.Errorf("Expected last time for metric2 to be %v, got %v", runTime1, trend.mLast["metric2"])
	}

	if trend.mCounts["metric2"] != 1 {
		t.Errorf("Expected count for metric2 to be 1, got %d", trend.mCounts["metric2"])
	}

	if trend.mTotals["metric2"] != 20.0 {
		t.Errorf("Expected total for metric2 to be 20.0, got %f", trend.mTotals["metric2"])
	}
}

func TestGetTrendLabel(t *testing.T) {
	trend := &Trend{}

	tests := []struct {
		tGrowth   float64
		tolerance float64
		expected  string
	}{
		{1.0, 0.5, utils.MetaPositive},
		{-1.0, 0.5, utils.MetaNegative},
		{0.0, 0.5, utils.MetaConstant},
		{0.3, 0.5, utils.MetaConstant},
		{-0.3, 0.5, utils.MetaConstant},
		{0.6, 0.5, utils.MetaPositive},
		{-0.6, 0.5, utils.MetaNegative},
	}

	for _, test := range tests {
		result := trend.getTrendLabel(test.tGrowth, test.tolerance)
		if result != test.expected {
			t.Errorf("For tGrowth: %f and tolerance: %f, expected %s, got %s", test.tGrowth, test.tolerance, test.expected, result)
		}
	}
}

func TestGetTrendGrowth(t *testing.T) {

	trend := Trend{
		mLast:   map[string]time.Time{},
		Metrics: map[time.Time]map[string]*MetricWithTrend{},
		mTotals: map[string]float64{},
		mCounts: map[string]int{},
	}

	_, err := trend.getTrendGrowth("unknownID", 100, utils.MetaLast, 2)
	if !errors.Is(err, utils.ErrNotFound) {
		t.Errorf("Expected error ErrNotFound, got: %v", err)
	}

	now := time.Now()
	trend.mLast["metric1"] = now

	_, err = trend.getTrendGrowth("metric1", 100, utils.MetaLast, 2)
	if !errors.Is(err, utils.ErrNotFound) {
		t.Errorf("Expected error ErrNotFound, got: %v", err)
	}

	trend.Metrics = map[time.Time]map[string]*MetricWithTrend{
		now: {
			"metric1": {ID: "metric1", Value: 80},
		},
	}

	got, err := trend.getTrendGrowth("metric1", 100, utils.MetaLast, 2)
	if err != nil || got != 25.0 {
		t.Errorf("Mismatch for MetaLast correlation. Got: %v, expected: %v", got, 25.0)
	}

	trend.mTotals = map[string]float64{
		"metric1": 400,
	}
	trend.mCounts = map[string]int{
		"metric1": 4,
	}

	got, err = trend.getTrendGrowth("metric1", 120, utils.MetaAverage, 2)
	if err != nil || got != 20.0 {
		t.Errorf("Mismatch for MetaAverage correlation. Got: %v, expected: %v", got, 20.0)
	}

	_, err = trend.getTrendGrowth("metric1", 100, "invalidCorrelation", 2)
	if !errors.Is(err, utils.ErrCorrelationUndefined) {
		t.Errorf("Expected error ErrCorrelationUndefined, got: %v", err)
	}
}

func TestNewTrendFromProfile(t *testing.T) {
	profile := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "trendProfileID",
		Schedule:        "@every 1sec",
		StatID:          "statID1",
		QueueLength:     10,
		TTL:             5 * time.Minute,
		MinItems:        1,
		CorrelationType: "average",
		Tolerance:       0.1,
		Stored:          true,
		ThresholdIDs:    []string{"threshold1", "threshold2"},
	}

	trend := NewTrendFromProfile(profile)

	if trend.Tenant != profile.Tenant {
		t.Errorf("Expected Tenant %s, got %s", profile.Tenant, trend.Tenant)
	}
	if trend.ID != profile.ID {
		t.Errorf("Expected ID %s, got %s", profile.ID, trend.ID)
	}
	if trend.RunTimes == nil {
		t.Errorf("Expected RunTimes to be initialized, got nil")
	}
	if len(trend.RunTimes) != 0 {
		t.Errorf("Expected RunTimes to be empty, got length %d", len(trend.RunTimes))
	}
	if trend.Metrics == nil {
		t.Errorf("Expected Metrics to be initialized, got nil")
	}
	if len(trend.Metrics) != 0 {
		t.Errorf("Expected Metrics to be empty, got length %d", len(trend.Metrics))
	}
	if trend.tPrfl != profile {
		t.Errorf("Expected tPrfl to point to the original profile, got a different value")
	}
}

func TestTrendProfileFieldAsString(t *testing.T) {
	tests := []struct {
		name    string
		fldPath []string
		err     error
		val     any
	}{
		{utils.ID, []string{utils.ID}, nil, "Trend1"},
		{utils.Tenant, []string{utils.Tenant}, nil, "cgrates.org"},
		{utils.Schedule, []string{utils.Schedule}, nil, "@every 1m"},
		{utils.StatID, []string{utils.StatID}, nil, "Stat1"},
		{utils.Metrics, []string{utils.Metrics + "[0]"}, nil, "*acc"},
		{utils.Metrics, []string{utils.Metrics + "[1]"}, nil, "*tcd"},
		{utils.TTL, []string{utils.TTL}, nil, 10 * time.Minute},
		{utils.QueueLength, []string{utils.QueueLength}, nil, 100},
		{utils.MinItems, []string{utils.MinItems}, nil, 10},
		{utils.CorrelationType, []string{utils.CorrelationType}, nil, "*average"},
		{utils.Tolerance, []string{utils.Tolerance}, nil, 0.05},
		{utils.Stored, []string{utils.Stored}, nil, true},
		{utils.ThresholdIDs, []string{utils.ThresholdIDs + "[0]"}, nil, "Thresh1"},
		{utils.ThresholdIDs, []string{utils.ThresholdIDs + "[1]"}, nil, "Thresh2"},
		{"NonExistingField", []string{"Field1"}, utils.ErrNotFound, nil},
	}
	rp := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "Trend1",
		Schedule:        "@every 1m",
		StatID:          "Stat1",
		Metrics:         []string{"*acc", "*tcd"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "*average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"Thresh1", "Thresh2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, err := rp.FieldAsInterface(tc.fldPath)
			if tc.err != nil {
				if err == nil {
					t.Error("expect to receive an error")
				}
				if tc.err != err {
					t.Errorf("expected %v,received %v", tc.err, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}
			if val != tc.val {
				t.Errorf("expected %v,received %v", tc.val, val)
			}
		})
	}
}

func TestTrendCleanUp(t *testing.T) {
	tm := time.Now().Add(-19 * time.Second)
	tm2 := tm.Add(15 * time.Second)
	tm3 := time.Now().Add(1)
	tm4 := time.Now().Add(-5 * time.Second)
	tm5 := time.Now().Add(-3 * time.Second)
	tr := &Trend{
		Tenant: "cgrates.org",
		ID:     "TREND1",
		RunTimes: []time.Time{
			tm,
			tm2,
			tm3,
			tm4,
			tm5,
		},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			tm: {
				utils.MetaTCC: {ID: utils.MetaTCC, Value: 13, TrendGrowth: -1, TrendLabel: utils.NotAvailable},
				utils.MetaACC: {ID: utils.MetaACC, Value: 13, TrendGrowth: -1, TrendLabel: utils.NotAvailable},
			},
			tm2: {
				utils.MetaTCC: {ID: utils.MetaTCC, Value: 30, TrendGrowth: 120, TrendLabel: utils.MetaPositive},
				utils.MetaACC: {ID: utils.MetaACC, Value: 15, TrendGrowth: 4, TrendLabel: utils.MetaPositive},
			},
			tm3: {
				utils.MetaTCC: {ID: utils.MetaTCC, Value: 30, TrendGrowth: 120, TrendLabel: utils.MetaPositive},
				utils.MetaACC: {ID: utils.MetaACC, Value: 15, TrendGrowth: 4, TrendLabel: utils.MetaPositive},
			},
			tm4: {
				utils.MetaTCC: {ID: utils.MetaTCC, Value: 30, TrendGrowth: 120, TrendLabel: utils.MetaPositive},
				utils.MetaACC: {ID: utils.MetaACC, Value: 15, TrendGrowth: 4, TrendLabel: utils.MetaPositive},
			},
			tm5: {
				utils.MetaTCC: {ID: utils.MetaTCC, Value: 30, TrendGrowth: 120, TrendLabel: utils.MetaPositive},
				utils.MetaACC: {ID: utils.MetaACC, Value: 15, TrendGrowth: 4, TrendLabel: utils.MetaPositive},
			},
		},
	}

	tests := []struct {
		name               string
		ttl                time.Duration
		tr                 *Trend
		qLength            int
		altered            bool
		runtimesmetriclens int
	}{
		{"TTLlgThan0", 10 * time.Second, tr, 3, true, 3},
		{"QueueLenLg0", -1, tr, 2, true, 2},
		{"TLLExpiredAll", 2 * time.Second, tr, 0, true, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			altered := tc.tr.cleanup(tc.ttl, tc.qLength)
			if tc.altered {
				if !altered {
					t.Errorf("expected trend to be altered")
				}
				if len(tc.tr.RunTimes) != tc.runtimesmetriclens || len(tc.tr.Metrics) != tc.runtimesmetriclens {
					t.Errorf("expected len to be %d,got %d metrics,%d runtimes", tc.runtimesmetriclens, len(tc.tr.Metrics), len(tc.tr.RunTimes))
				}

				return
			}
			if altered {
				t.Error("expected trend to not be altered")
			}
		})
	}
}

func TestTrendProfileString(t *testing.T) {
	profile := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "Trend1",
		Schedule:        "@every 1m",
		StatID:          "Stat1",
		Metrics:         []string{"*acc", "*tcd"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "*average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"Thresh1", "Thresh2"},
	}

	jsonStr := profile.String()
	if jsonStr == "" {
		t.Errorf("Expected non-empty JSON string representation of TrendProfile")
	}

	expectedFields := []string{
		`"cgrates.org"`, `"Trend1"`, `"@every 1m"`, `"Stat1"`,
		`"*acc"`, `"*tcd"`, `"*average"`, `true`, `"Thresh1"`, `"Thresh2"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON output to contain field: %s", field)
		}
	}
}

func TestTrendProfileFieldAssString(t *testing.T) {
	tests := []struct {
		name     string
		fldPath  []string
		expected string
		hasError bool
	}{
		{
			name:     "Valid Tenant Field",
			fldPath:  []string{"Tenant"},
			expected: "cgrates.org",
			hasError: false,
		},
		{
			name:     "Valid ID Field",
			fldPath:  []string{"ID"},
			expected: "Trend1",
			hasError: false,
		},
		{
			name:     "Valid Schedule Field",
			fldPath:  []string{"Schedule"},
			expected: "@every 1m",
			hasError: false,
		},
		{
			name:     "Invalid Field Path",
			fldPath:  []string{"NonExistentField"},
			expected: "",
			hasError: true,
		},
	}

	tp := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "Trend1",
		Schedule:        "@every 1m",
		StatID:          "Stat1",
		Metrics:         []string{"*acc", "*tcd"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "*average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"Thresh1", "Thresh2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tp.FieldAsString(tt.fldPath)

			if tt.hasError && err == nil {
				t.Errorf("Expected error for fldPath %v, but got none", tt.fldPath)
			} else if !tt.hasError && err != nil {
				t.Errorf("Expected no error for fldPath %v, but got: %v", tt.fldPath, err)
			}

			if result != tt.expected {
				t.Errorf("For fldPath %v, expected %v, but got %v", tt.fldPath, tt.expected, result)
			}
		})
	}
}

func TestTrendProfileSet(t *testing.T) {
	tp := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "Trend1",
		Schedule:        "@every 1m",
		StatID:          "Stat1",
		Metrics:         []string{"*acc", "*tcd"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        10,
		CorrelationType: "*average",
		Tolerance:       0.05,
		Stored:          true,
		ThresholdIDs:    []string{"Thresh1", "Thresh2"},
	}

	tests := []struct {
		name     string
		path     []string
		val      any
		expected any
		hasError bool
	}{
		{
			name:     "Set Tenant",
			path:     []string{utils.Tenant},
			val:      "newTenant",
			expected: "newTenant",
			hasError: false,
		},
		{
			name:     "Set ID",
			path:     []string{utils.ID},
			val:      "newID",
			expected: "newID",
			hasError: false,
		},
		{
			name:     "Set Schedule",
			path:     []string{utils.Schedule},
			val:      "@every 2m",
			expected: "@every 2m",
			hasError: false,
		},
		{
			name:     "Set StatID",
			path:     []string{utils.StatID},
			val:      "newStatID",
			expected: "newStatID",
			hasError: false,
		},
		{
			name:     "Set Metrics",
			path:     []string{utils.Metrics},
			val:      []string{"*newMetric"},
			expected: []string{"*acc", "*tcd", "*newMetric"},
			hasError: false,
		},
		{
			name:     "Set TTL",
			path:     []string{utils.TTL},
			val:      "15m",
			expected: 15 * time.Minute,
			hasError: false,
		},
		{
			name:     "Set QueueLength",
			path:     []string{utils.QueueLength},
			val:      50,
			expected: 50,
			hasError: false,
		},
		{
			name:     "Set MinItems",
			path:     []string{utils.MinItems},
			val:      20,
			expected: 20,
			hasError: false,
		},
		{
			name:     "Set CorrelationType",
			path:     []string{utils.CorrelationType},
			val:      "*sum",
			expected: "*sum",
			hasError: false,
		},
		{
			name:     "Set Tolerance",
			path:     []string{utils.Tolerance},
			val:      0.1,
			expected: 0.1,
			hasError: false,
		},
		{
			name:     "Set Stored",
			path:     []string{utils.Stored},
			val:      false,
			expected: false,
			hasError: false,
		},
		{
			name:     "Set ThresholdIDs",
			path:     []string{utils.ThresholdIDs},
			val:      []string{"Thresh3", "Thresh4"},
			expected: []string{"Thresh1", "Thresh2", "Thresh3", "Thresh4"},
			hasError: false,
		},
		{
			name:     "Set Invalid Path",
			path:     []string{"InvalidPath"},
			val:      "invalid",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tp.Set(tt.path, tt.val, false)

			if tt.hasError && err == nil {
				t.Errorf("Expected error for path %v, but got none", tt.path)
			} else if !tt.hasError && err != nil {
				t.Errorf("Expected no error for path %v, but got: %v", tt.path, err)
			}

			switch tt.path[0] {
			case utils.Tenant:
				if tp.Tenant != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Tenant)
				}
			case utils.ID:
				if tp.ID != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.ID)
				}
			case utils.Schedule:
				if tp.Schedule != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Schedule)
				}
			case utils.StatID:
				if tp.StatID != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.StatID)
				}
			case utils.Metrics:
				if len(tp.Metrics) != len(tt.expected.([]string)) {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Metrics)
				} else {
					for i := range tp.Metrics {
						if tp.Metrics[i] != tt.expected.([]string)[i] {
							t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Metrics)
							break
						}
					}
				}
			case utils.TTL:
				if tp.TTL != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.TTL)
				}
			case utils.QueueLength:
				if tp.QueueLength != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.QueueLength)
				}
			case utils.MinItems:
				if tp.MinItems != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.MinItems)
				}
			case utils.CorrelationType:
				if tp.CorrelationType != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.CorrelationType)
				}
			case utils.Tolerance:
				if tp.Tolerance != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Tolerance)
				}
			case utils.Stored:
				if tp.Stored != tt.expected {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.Stored)
				}
			case utils.ThresholdIDs:
				if len(tp.ThresholdIDs) != len(tt.expected.([]string)) {
					t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.ThresholdIDs)
				} else {
					for i := range tp.ThresholdIDs {
						if tp.ThresholdIDs[i] != tt.expected.([]string)[i] {
							t.Errorf("For path %v, expected %v, but got %v", tt.path, tt.expected, tp.ThresholdIDs)
							break
						}
					}
				}
			}
		})
	}
}

func TestTrendProfileMergeV2(t *testing.T) {
	tp1 := &TrendProfile{
		Tenant:          "tenant1",
		ID:              "id1",
		Schedule:        "schedule1",
		StatID:          "stat1",
		Metrics:         []string{"metric1", "metric2"},
		ThresholdIDs:    []string{"threshold1"},
		Stored:          true,
		TTL:             10,
		QueueLength:     100,
		MinItems:        5,
		CorrelationType: "type1",
		Tolerance:       0.5,
	}

	tp2 := &TrendProfile{
		Tenant:          "tenant2",
		ID:              "id2",
		Schedule:        "schedule2",
		StatID:          "stat2",
		Metrics:         []string{"metric3", "metric4"},
		ThresholdIDs:    []string{"threshold2"},
		Stored:          true,
		TTL:             20,
		QueueLength:     200,
		MinItems:        10,
		CorrelationType: "type2",
		Tolerance:       1.5,
	}

	tp1.Merge(tp2)

	if tp1.Tenant != "tenant2" {
		t.Errorf("Expected Tenant to be 'tenant2', but got: %s", tp1.Tenant)
	}

	if tp1.ID != "id2" {
		t.Errorf("Expected ID to be 'id2', but got: %s", tp1.ID)
	}

	if tp1.Schedule != "schedule2" {
		t.Errorf("Expected Schedule to be 'schedule2', but got: %s", tp1.Schedule)
	}

	if tp1.StatID != "stat2" {
		t.Errorf("Expected StatID to be 'stat2', but got: %s", tp1.StatID)
	}

	expectedMetrics := []string{"metric1", "metric2", "metric3", "metric4"}
	if len(tp1.Metrics) != len(expectedMetrics) {
		t.Errorf("Expected Metrics to be %v, but got: %v", expectedMetrics, tp1.Metrics)
	}

	expectedThresholdIDs := []string{"threshold1", "threshold2"}
	if len(tp1.ThresholdIDs) != len(expectedThresholdIDs) {
		t.Errorf("Expected ThresholdIDs to be %v, but got: %v", expectedThresholdIDs, tp1.ThresholdIDs)
	}

	if tp1.Stored != true {
		t.Errorf("Expected Stored to be 'true', but got: %v", tp1.Stored)
	}

	if tp1.TTL != 20 {
		t.Errorf("Expected TTL to be 20, but got: %d", tp1.TTL)
	}

	if tp1.QueueLength != 200 {
		t.Errorf("Expected QueueLength to be 200, but got: %d", tp1.QueueLength)
	}

	if tp1.MinItems != 10 {
		t.Errorf("Expected MinItems to be 10, but got: %d", tp1.MinItems)
	}

	if tp1.CorrelationType != "type2" {
		t.Errorf("Expected CorrelationType to be 'type2', but got: %s", tp1.CorrelationType)
	}

	if tp1.Tolerance != 1.5 {
		t.Errorf("Expected Tolerance to be 1.5, but got: %f", tp1.Tolerance)
	}
}
