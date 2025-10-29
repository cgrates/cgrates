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

package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestTrendGetTrendGrowth(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-time.Second)
	t2 := now.Add(-2 * time.Second)
	t3 := now.Add(-3 * time.Second)
	trnd1 := &Trend{
		Tenant:   "cgrates.org",
		ID:       "TestTrendGetTrendLabel",
		RunTimes: []time.Time{t3, t2, t1},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t3: {utils.MetaTCD: {utils.MetaTCD, float64(41 * time.Second), -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 41.0, -1.0, utils.NotAvailable}},
			t2: {utils.MetaTCD: {utils.MetaTCD, float64(9 * time.Second), -78.048, utils.MetaNegative}, utils.MetaTCC: {utils.MetaTCC, 9.0, -78.048, utils.MetaNegative}},
			t1: {utils.MetaTCD: {utils.MetaTCD, float64(10 * time.Second), 11.11111, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 10.0, 11.11111, utils.MetaPositive}}},
	}
	trnd1.computeIndexes()
	if _, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.NotAvailable, 5); err != utils.ErrCorrelationUndefined {
		t.Error(err)
	}
	if growth, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.MetaLast, 5); err != nil || growth != 10.0 {
		t.Errorf("Expecting: <%f> got <%f>, err: %v", 10.0, growth, err)
	}
	if growth, err := trnd1.getTrendGrowth(utils.MetaTCD, float64(11*time.Second), utils.MetaAverage, 5); err != nil || growth != -45.0 {
		t.Errorf("Expecting: <%f> got <%f>, err: %v", -45.0, growth, err)
	}
}

func TestTrendGetTrendLabel(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-time.Second)
	t2 := now.Add(-2 * time.Second)
	t3 := now.Add(-3 * time.Second)
	trnd1 := &Trend{
		Tenant:   "cgrates.org",
		ID:       "TestTrendGetTrendLabel",
		RunTimes: []time.Time{t3, t2, t1},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t3: {utils.MetaTCD: {utils.MetaTCD, float64(41 * time.Second), -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 41.0, -1.0, utils.NotAvailable}},
			t2: {utils.MetaTCD: {utils.MetaTCD, float64(9 * time.Second), -78.048, utils.MetaNegative}, utils.MetaTCC: {utils.MetaTCC, 9.0, -78.048, utils.MetaNegative}},
			t1: {utils.MetaTCD: {utils.MetaTCD, float64(10 * time.Second), 11.11111, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 10.0, 11.11111, utils.MetaPositive}}},
	}
	trnd1.computeIndexes()
	expct := utils.MetaPositive
	if lbl := trnd1.getTrendLabel(11.0, 0.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	if lbl := trnd1.getTrendLabel(11.0, 10.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaConstant
	if lbl := trnd1.getTrendLabel(11.0, 11.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaNegative
	if lbl := trnd1.getTrendLabel(-9.0, 8.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
	expct = utils.MetaConstant
	if lbl := trnd1.getTrendLabel(-9.0, 10.0); lbl != expct {
		t.Errorf("Expecting: <%q> got <%q>", expct, lbl)
	}
}

func TestTrendCleanUp(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		qLength  int
		runtimes int
		want     bool
	}{
		{
			name:     "No Clean up",
			ttl:      2 * time.Minute,
			qLength:  10,
			want:     false,
			runtimes: 5,
		},
		{
			name:     "Clean up with TTL",
			ttl:      1 * time.Second,
			qLength:  10,
			want:     true,
			runtimes: 3,
		},
		{
			name:     "No clean up",
			ttl:      1 * time.Second,
			qLength:  10,
			want:     false,
			runtimes: 3,
		},
		{
			name:     "Clean up with queue length",
			ttl:      1 * time.Second,
			qLength:  2,
			want:     true,
			runtimes: 2,
		},
		{
			name:     "No clean up with negative ttl",
			ttl:      -1,
			qLength:  2,
			want:     false,
			runtimes: 2,
		},
	}
	now := time.Now()
	t1 := now.Add(-time.Minute)
	t2 := now.Add(-time.Second)
	t3 := now.Add(time.Second)
	t4 := now.Add(2 * time.Second)
	t5 := now.Add(time.Minute)
	trend := &Trend{
		Tenant: "cgrates.org",
		ID:     "TestTrendCleanUp",

		RunTimes: []time.Time{t1, t2, t3, t4, t5},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t1: {utils.MetaACC: {utils.MetaACC, 10.1, -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 10.1, -1.0, utils.NotAvailable}},
			t2: {utils.MetaACC: {utils.MetaACC, 15.1, 4.0, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 25.1, 15.1, utils.MetaPositive}},
			t3: {utils.MetaACC: {utils.MetaACC, 12.1, -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 34, -1.0, utils.NotAvailable}},
			t4: {utils.MetaACC: {utils.MetaACC, 19.1, 4.0, utils.MetaPositive}, utils.MetaTCC: {utils.MetaTCC, 48, 15.1, utils.MetaPositive}},
			t5: {utils.MetaACC: {utils.MetaACC, 117.1, -1.0, utils.NotAvailable}, utils.MetaTCC: {utils.MetaTCC, 56, -1.0, utils.NotAvailable}},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			altered := trend.cleanup(tt.ttl, tt.qLength)
			if altered != tt.want {
				t.Errorf("cleanup() = %v, want %v", altered, tt.want)
				return
			}
			if len(trend.RunTimes) != tt.runtimes {
				t.Errorf("runTimes length = %v, want %v", len(trend.RunTimes), tt.runtimes)
			}
		})

	}
}

func TestLibTrendSClone(t *testing.T) {
	original := &TrendProfile{
		Tenant:          "cgrates.org",
		ID:              "ID1",
		Schedule:        "0 * * * *",
		StatID:          "statID",
		Metrics:         []string{"metric1", "metric2"},
		TTL:             10 * time.Minute,
		QueueLength:     100,
		MinItems:        5,
		CorrelationType: "*average",
		Tolerance:       0.1,
		Stored:          true,
		ThresholdIDs:    []string{"threshold1", "threshold2"},
	}

	cloned := original.Clone()

	if cloned.Tenant != original.Tenant {
		t.Errorf("Expected Tenant %s, got %s", original.Tenant, cloned.Tenant)
	}
	if cloned.ID != original.ID {
		t.Errorf("Expected ID %s, got %s", original.ID, cloned.ID)
	}
	if cloned.Schedule != original.Schedule {
		t.Errorf("Expected Schedule %s, got %s", original.Schedule, cloned.Schedule)
	}
	if cloned.StatID != original.StatID {
		t.Errorf("Expected StatID %s, got %s", original.StatID, cloned.StatID)
	}
	if len(cloned.Metrics) != len(original.Metrics) {
		t.Errorf("Expected Metrics length %d, got %d", len(original.Metrics), len(cloned.Metrics))
	}
	for i, metric := range original.Metrics {
		if cloned.Metrics[i] != metric {
			t.Errorf("Expected Metrics[%d] %s, got %s", i, metric, cloned.Metrics[i])
		}
	}
	if cloned.TTL != original.TTL {
		t.Errorf("Expected TTL %v, got %v", original.TTL, cloned.TTL)
	}
	if cloned.QueueLength != original.QueueLength {
		t.Errorf("Expected QueueLength %d, got %d", original.QueueLength, cloned.QueueLength)
	}
	if cloned.MinItems != original.MinItems {
		t.Errorf("Expected MinItems %d, got %d", original.MinItems, cloned.MinItems)
	}
	if cloned.CorrelationType != original.CorrelationType {
		t.Errorf("Expected CorrelationType %s, got %s", original.CorrelationType, cloned.CorrelationType)
	}
	if cloned.Tolerance != original.Tolerance {
		t.Errorf("Expected Tolerance %f, got %f", original.Tolerance, cloned.Tolerance)
	}
	if cloned.Stored != original.Stored {
		t.Errorf("Expected Stored %v, got %v", original.Stored, cloned.Stored)
	}
	if len(cloned.ThresholdIDs) != len(original.ThresholdIDs) {
		t.Errorf("Expected ThresholdIDs length %d, got %d", len(original.ThresholdIDs), len(cloned.ThresholdIDs))
	}
	for i, thresholdID := range original.ThresholdIDs {
		if cloned.ThresholdIDs[i] != thresholdID {
			t.Errorf("Expected ThresholdIDs[%d] %s, got %s", i, thresholdID, cloned.ThresholdIDs[i])
		}
	}
}

func TestLibTrendsComputeIndexes(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(1 * time.Second)
	t3 := t2.Add(2 * time.Second)

	trend := &Trend{
		RunTimes: []time.Time{t3, t2, t1},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			t3: {
				"metric1": {ID: "metric1", Value: 41.0},
				"metric2": {ID: "metric2", Value: 10.0},
			},
			t2: {
				"metric1": {ID: "metric1", Value: 9.0},
			},
			t1: {
				"metric2": {ID: "metric2", Value: 5.0},
				"metric3": {ID: "metric3", Value: 2.0},
			},
		},
	}

	trend.computeIndexes()

	if count := trend.mCounts["metric1"]; count != 2 {
		t.Errorf("Expected metric1 count to be 2, got %d", count)
	}

	if total := trend.mTotals["metric1"]; total != 50.0 {
		t.Errorf("Expected metric1 total to be 50.0, got %f", total)
	}

	if count := trend.mCounts["metric2"]; count != 2 {
		t.Errorf("Expected metric2 count to be 2, got %d", count)
	}

	if total := trend.mTotals["metric2"]; total != 15.0 {
		t.Errorf("Expected metric2 total to be 15.0, got %f", total)
	}

	if count := trend.mCounts["metric3"]; count != 1 {
		t.Errorf("Expected metric3 count to be 1, got %d", count)
	}

	if total := trend.mTotals["metric3"]; total != 2.0 {
		t.Errorf("Expected metric3 total to be 2.0, got %f", total)
	}
}

func TestGetTrendGrowthNotFound(t *testing.T) {

	trend := &Trend{
		mLast:   map[string]time.Time{},
		Metrics: map[time.Time]map[string]*MetricWithTrend{},
	}

	tG, err := trend.getTrendGrowth("nonexistentID", 100, utils.MetaLast, 2)

	if tG != -1.0 {
		t.Errorf("Expected trend growth to be -1.0, but got: %v", tG)
	}

	if !errors.Is(err, utils.ErrNotFound) {
		t.Errorf("Expected error to be ErrNotFound, but got: %v", err)
	}
}

func TestGetTrendGrowthMetricNotFound(t *testing.T) {

	trend := &Trend{
		mLast: map[string]time.Time{
			"metricID": time.Now(),
		},
		Metrics: map[time.Time]map[string]*MetricWithTrend{
			time.Now(): {},
		},
	}

	tG, err := trend.getTrendGrowth("metricID", 150, utils.MetaLast, 2)

	if tG != -1.0 {
		t.Errorf("Expected trend growth to be -1.0, but got: %v", tG)
	}

	if !errors.Is(err, utils.ErrNotFound) {
		t.Errorf("Expected error to be ErrNotFound, but got: %v", err)
	}
}

func TestGetTrendLabel_DefaultCase(t *testing.T) {

	trend := &Trend{}

	tGrowth := 0.0
	tolerance := 1.0

	lbl := trend.getTrendLabel(tGrowth, tolerance)

	if lbl != utils.MetaConstant {
		t.Errorf("Expected label to be %s, but got: %s", utils.MetaConstant, lbl)
	}
}

func TestTrendCompile(t *testing.T) {

	trend := &Trend{
		mTotals: nil,
	}

	cleanTtl := time.Minute
	qLength := 10
	trend.Compile(cleanTtl, qLength)

	if trend.mTotals == nil {
		t.Error("Expected mTotals to be initialized, but it is nil")
	}
}

func TestUncompressAndSortRunTimes(t *testing.T) {

	trend := &Trend{
		CompressedMetrics: []byte(`{
            "2024-10-24T10:00:00Z": {
                "metric1": {
                    "ID": "metric1",
                    "Value": 75.0,
                    "TrendGrowth": 5.0,
                    "TrendLabel": "*positive"
                }
            },
            "2024-10-24T11:00:00Z": {
                "metric2": {
                    "ID": "metric2",
                    "Value": 65.0,
                    "TrendGrowth": -3.0,
                    "TrendLabel": "*negative"
                }
            },
            "2024-10-24T12:00:00Z": {
                "metric3": {
                    "ID": "metric3",
                    "Value": 65.0,
                    "TrendGrowth": 0.0,
                    "TrendLabel": "*constant"
                }
            }
        }`),
		Metrics: make(map[time.Time]map[string]*MetricWithTrend),
	}

	marshaler := &JSONMarshaler{}

	err := trend.uncompress(marshaler)
	if err != nil {
		t.Fatalf("Expected no error, got <%+v>", err)
	}

	if len(trend.RunTimes) == 0 {
		t.Fatalf("Expected RunTimes to be populated, got empty slice")
	}

	if len(trend.Metrics) == 0 {
		t.Fatalf("Expected Metrics to be populated, got empty map")
	}

	for i := 1; i < len(trend.RunTimes); i++ {
		if trend.RunTimes[i].Before(trend.RunTimes[i-1]) {
			t.Errorf("RunTimes are not sorted: %v at index %d is before %v at index %d", trend.RunTimes[i], i, trend.RunTimes[i-1], i-1)
		}
	}

	if metric, exists := trend.Metrics[trend.RunTimes[0]]["metric1"]; exists {
		if metric.ID != "metric1" || metric.Value != 75.0 || metric.TrendGrowth != 5.0 || metric.TrendLabel != "*positive" {
			t.Errorf("Metric1 data does not match expected values: %+v", metric)
		}
	} else {
		t.Errorf("Expected metric1 not found in Metrics")
	}

	if metric, exists := trend.Metrics[trend.RunTimes[1]]["metric2"]; exists {
		if metric.ID != "metric2" || metric.Value != 65.0 || metric.TrendGrowth != -3.0 || metric.TrendLabel != "*negative" {
			t.Errorf("Metric2 data does not match expected values: %+v", metric)
		}
	} else {
		t.Errorf("Expected metric2 not found in Metrics")
	}

	if metric, exists := trend.Metrics[trend.RunTimes[2]]["metric3"]; exists {
		if metric.ID != "metric3" || metric.Value != 65.0 || metric.TrendGrowth != 0.0 || metric.TrendLabel != "*constant" {
			t.Errorf("Metric3 data does not match expected values: %+v", metric)
		}
	} else {
		t.Errorf("Expected metric3 not found in Metrics")
	}
}

func TestTrendCompressSuccess(t *testing.T) {
	marshaler := &JSONMarshaler{}

	metrics := map[time.Time]map[string]*MetricWithTrend{
		time.Now(): {
			"metric1": {ID: "metric1", Value: 1.0, TrendGrowth: 0.1, TrendLabel: "*positive"},
		},
	}

	trend := &Trend{
		Metrics:  metrics,
		RunTimes: []time.Time{time.Now()},
	}

	tr, err := trend.compress(marshaler)
	if err != nil {
		t.Errorf("Expected no error, got <%+v>", err)
	}

	if tr.CompressedMetrics == nil {
		t.Errorf("Expected CompressedMetrics to be populated, got: %+v", tr.CompressedMetrics)
	}

	if tr.Metrics != nil {
		t.Errorf("Expected Metrics to be nil after compression, got: %+v", tr.Metrics)
	}
	if tr.RunTimes != nil {
		t.Errorf("Expected RunTimes to be nil after compression, got: %+v", tr.RunTimes)
	}
}

func TestTrendUncompressNil(t *testing.T) {
	marshaler := &JSONMarshaler{}

	var trend *Trend
	err := trend.uncompress(marshaler)
	if err != nil {
		t.Errorf("Expected no error when Trend is nil, got <%+v>", err)
	}

	trend = &Trend{CompressedMetrics: nil}
	err = trend.uncompress(marshaler)
	if err != nil {
		t.Errorf("Expected no error when CompressedMetrics is nil, got <%+v>", err)
	}
}

func TestTrendUncompressSuccess(t *testing.T) {
	marshaler := &JSONMarshaler{}

	metrics := map[time.Time]map[string]*MetricWithTrend{
		time.Now(): {
			"metric1": {ID: "metric1", Value: 1.0, TrendGrowth: 0.1, TrendLabel: "*positive"},
		},
	}

	compressedMetrics, err := marshaler.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal metrics: %v", err)
	}

	trend := &Trend{
		CompressedMetrics: compressedMetrics,
	}

	err = trend.uncompress(marshaler)
	if err != nil {
		t.Errorf("Expected no error, got <%+v>", err)
	}

	if trend.Metrics == nil {
		t.Errorf("Expected Metrics to be populated, got: %+v", trend.Metrics)
	}

	if len(trend.Metrics) != len(metrics) {
		t.Errorf("Expected Metrics length to be %d, got: %d", len(metrics), len(trend.Metrics))
	}
}
