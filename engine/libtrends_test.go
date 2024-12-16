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
