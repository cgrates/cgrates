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
