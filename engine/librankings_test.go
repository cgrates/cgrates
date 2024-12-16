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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"
)

func TestRankingProfileTenantID(t *testing.T) {
	tests := []struct {
		name     string
		profile  RankingProfile
		expected string
	}{
		{
			name:     "Tenant and ID",
			profile:  RankingProfile{Tenant: "cgrates.org", ID: "1"},
			expected: "cgrates.org:1",
		},
		{
			name:     "Empty tenant",
			profile:  RankingProfile{ID: "2"},
			expected: ":2",
		},
		{
			name:     "Empty ID",
			profile:  RankingProfile{Tenant: "cgrates.org"},
			expected: "cgrates.org:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.profile.TenantID()
			if got != tc.expected {
				t.Errorf("TenantID() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestRankingProfileClone(t *testing.T) {
	original := &RankingProfile{
		Tenant:   "cgrates.org",
		ID:       "ID1",
		Schedule: "@every 1sec",
		Sorting:  "asc",
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
	if cloned.Sorting != original.Sorting {
		t.Errorf("Expected Sorting %s, got %s", original.Sorting, cloned.Sorting)
	}

	if cloned == original {
		t.Error("Clone should return a new instance, but it returned the same reference")
	}
}

func TestNewRankingFromProfile(t *testing.T) {
	profile := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "ID1",
		Sorting:           "asc",
		MetricIDs:         []string{"metric1", "metric2"},
		SortingParameters: []string{"param1", "param2"},
	}

	ranking := NewRankingFromProfile(profile)

	if ranking.Tenant != profile.Tenant {
		t.Errorf("Expected Tenant %s, got %s", profile.Tenant, ranking.Tenant)
	}
	if ranking.ID != profile.ID {
		t.Errorf("Expected ID %s, got %s", profile.ID, ranking.ID)
	}
	if ranking.Sorting != profile.Sorting {
		t.Errorf("Expected Sorting %s, got %s", profile.Sorting, ranking.Sorting)
	}

	if ranking.Metrics == nil {
		t.Error("Expected Metrics map to be initialized, but it is nil")
	}

	expectedMetricIDs := utils.NewStringSet(profile.MetricIDs)
	if !ranking.metricIDs.Equals(expectedMetricIDs) {
		t.Errorf("Expected metricIDs %v, got %v", expectedMetricIDs, ranking.metricIDs)
	}

	if len(ranking.SortingParameters) != len(profile.SortingParameters) {
		t.Errorf("Expected SortingParameters length %d, got %d", len(profile.SortingParameters), len(ranking.SortingParameters))
	}
	for i, param := range profile.SortingParameters {
		if ranking.SortingParameters[i] != param {
			t.Errorf("Expected SortingParameters[%d] %s, got %s", i, param, ranking.SortingParameters[i])
		}
	}

	if ranking.rkPrfl != profile {
		t.Error("Expected rkPrfl to reference the original profile")
	}
}

func TestRankingTenantID(t *testing.T) {
	r := &Ranking{
		Tenant: "cgrates.org",
		ID:     "1",
	}
	expectedTenantID := "cgrates.org:1"
	actualTenantID := r.TenantID()
	if actualTenantID != expectedTenantID {
		t.Errorf("Expected tenant ID %s, got %s", expectedTenantID, actualTenantID)
	}
}

func TestRanking_asRankingSummary(t *testing.T) {
	rk := &Ranking{
		Tenant:        "cgrates.org",
		ID:            "ID1",
		LastUpdate:    time.Now(),
		SortedStatIDs: []string{"stat1", "stat2", "stat3"},
	}

	rkSummary := rk.asRankingSummary()

	if rkSummary.Tenant != rk.Tenant {
		t.Errorf("Expected Tenant %s, but got %s", rk.Tenant, rkSummary.Tenant)
	}
	if rkSummary.ID != rk.ID {
		t.Errorf("Expected ID %s, but got %s", rk.ID, rkSummary.ID)
	}
	if !rkSummary.LastUpdate.Equal(rk.LastUpdate) {
		t.Errorf("Expected LastUpdate %v, but got %v", rk.LastUpdate, rkSummary.LastUpdate)
	}
	if !reflect.DeepEqual(rkSummary.SortedStatIDs, rk.SortedStatIDs) {
		t.Errorf("Expected SortedStatIDs %v, but got %v", rk.SortedStatIDs, rkSummary.SortedStatIDs)
	}

	if &rkSummary.SortedStatIDs == &rk.SortedStatIDs {
		t.Errorf("Expected SortedStatIDs slice to be copied, not referenced")
	}
}

func TestRankingSortStats(t *testing.T) {
	metrics := map[string]map[string]float64{
		"stat1": {
			"metric1": 5.0,
			"metric2": 10.0,
		},
		"stat2": {
			"metric1": 3.0,
			"metric2": 12.0,
		},
		"stat3": {
			"metric1": 7.0,
			"metric2": 8.0,
		},
	}

	tests := []struct {
		name          string
		sortingType   string
		sortingParams []string
		expectedOrder []string
		expectError   bool
	}{
		{
			name:          "Sort Descending by metric1",
			sortingType:   utils.MetaDesc,
			sortingParams: []string{"metric1"},
			expectedOrder: []string{"stat3", "stat1", "stat2"},
			expectError:   false,
		},
		{
			name:          "Sort Ascending by metric2",
			sortingType:   utils.MetaAsc,
			sortingParams: []string{"metric2"},
			expectedOrder: []string{"stat3", "stat1", "stat2"},
			expectError:   false,
		},
		{
			name:          "Unsupported sorting type",
			sortingType:   "unsupported",
			sortingParams: []string{"metric1"},
			expectedOrder: nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sortedStatIDs, err := rankingSortStats(tt.sortingType, tt.sortingParams, metrics)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for i, id := range sortedStatIDs {
				if id != tt.expectedOrder[i] {
					t.Errorf("expected sorted statID %v at index %d, got %v", tt.expectedOrder[i], i, id)
				}
			}
		})
	}
}

func TestRankingMixedOrder(t *testing.T) {
	statmetrics := map[string]map[string]float64{
		"Stat1": {"*acc": 13},
		"Stat6": {"*acc": 10, "*pdd": 700, "*tcc": 121},
		"Stat2": {"*acc": 14},
		"Stat5": {"*acc": 10, "*pdd": 700, "*tcc": 120},
		"Stat3": {"*acc": 12.1, "*pdd": 900},
		"Stat7": {"*acc": 10, "*pdd": 600, "*tcc": 123},
		"Stat4": {"*acc": 12.1, "*pdd": 1000},
	}

	testCases := []struct {
		name       string
		sortMetric []string
		sorter     string
		statIDs    []string
		expErr     error
	}{
		{
			name:       "TestSortStatsAsc",
			sortMetric: []string{"*acc", "*pdd:false", "*tcc"},
			sorter:     "*asc",
			statIDs:    []string{"Stat5", "Stat6", "Stat7", "Stat4", "Stat3", "Stat1", "Stat2"},
		},
		{
			name:       "TestSortStatsDesc",
			sortMetric: []string{"*tcc", "*pdd:false", "*acc"},
			sorter:     "*desc",
			statIDs:    []string{"Stat7", "Stat6", "Stat5", "Stat3", "Stat4", "Stat2", "Stat1"},
		},
		{
			name:       "TestSortStatsDesc",
			sortMetric: []string{"*acc", "*tcc", "*pdd:false"},
			sorter:     "*desc",
			statIDs:    []string{"Stat2", "Stat1", "Stat3", "Stat4", "Stat7", "Stat6", "Stat5"},
		},
		{
			name:       "TestSortStatsAsc",
			sortMetric: []string{"*tcc", "*pdd:false", "*acc"},
			sorter:     "*asc",
			statIDs:    []string{"Stat5", "Stat6", "Stat7", "Stat4", "Stat3", "Stat1", "Stat2"},
		},
		{
			name:       "TestSortStatsDesc",
			sortMetric: []string{"*pdd:false", "*acc", "*tcc"},
			sorter:     "*desc",
			statIDs:    []string{"Stat7", "Stat6", "Stat5", "Stat3", "Stat4", "Stat2", "Stat1"},
		},
		{
			name:       "TestSortStatsAsc",
			sortMetric: []string{"*tcc", "*acc", "*pdd:false"},
			sorter:     "*asc",
			statIDs:    []string{"Stat5", "Stat6", "Stat7", "Stat4", "Stat3", "Stat1", "Stat2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rs, err := newRankingSorter(tc.sorter, tc.sortMetric, statmetrics)
			if tc.expErr != nil {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tc.expErr.Error() != err.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if resStatIDs := rs.sortStatIDs(); !reflect.DeepEqual(resStatIDs, tc.statIDs) {
				t.Errorf("Expecting: %v, received %v", tc.statIDs, resStatIDs)
			}
		})
	}

}

func TestRankingProfileFieldAsString(t *testing.T) {
	tests := []struct {
		name    string
		fldPath []string
		err     error
		val     any
	}{
		{utils.ID, []string{utils.ID}, nil, "RP1"},
		{utils.Tenant, []string{utils.Tenant}, nil, "cgrates.org"},
		{utils.Schedule, []string{utils.Schedule}, nil, "@every 2s"},
		{utils.StatIDs, []string{utils.StatIDs + "[0]"}, nil, "Stat1"},
		{utils.StatIDs, []string{utils.StatIDs + "[1]"}, nil, "Stat2"},
		{utils.MetricIDs, []string{utils.MetricIDs + "[0]"}, nil, "*tcc"},
		{utils.MetricIDs, []string{utils.MetricIDs + "[1]"}, nil, "*acc"},
		{utils.Sorting, []string{utils.Sorting}, nil, "*asc"},
		{utils.Stored, []string{utils.Stored}, nil, false},
		{utils.SortingParameters, []string{utils.SortingParameters + "[0]"}, nil, "*acc"},
		{utils.SortingParameters, []string{utils.SortingParameters + "[1]"}, nil, "*pdd:false"},
		{utils.ThresholdIDs, []string{utils.ThresholdIDs + "[0]"}, nil, "Threshold1"},
		{utils.ThresholdIDs, []string{utils.ThresholdIDs + "[1]"}, nil, "Threshold2"},
		{"NonExistingField", []string{"Field1"}, utils.ErrNotFound, nil},
	}
	rp := &RankingProfile{
		Tenant:            "cgrates.org",
		ID:                "RP1",
		Schedule:          "@every 2s",
		StatIDs:           []string{"Stat1", "Stat2"},
		MetricIDs:         []string{"*tcc", "*acc", "*pdd"},
		Sorting:           "*asc",
		SortingParameters: []string{"*acc", "*pdd:false"},
		ThresholdIDs:      []string{"Threshold1", "Threshold2"},
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
