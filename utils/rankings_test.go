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
package utils

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
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

	expectedMetricIDs := NewStringSet(profile.MetricIDs)
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

	rkSummary := rk.AsRankingSummary()

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
			sortingType:   MetaDesc,
			sortingParams: []string{"metric1"},
			expectedOrder: []string{"stat3", "stat1", "stat2"},
			expectError:   false,
		},
		{
			name:          "Sort Ascending by metric2",
			sortingType:   MetaAsc,
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
			sortedStatIDs, err := RankingSortStats(tt.sortingType, tt.sortingParams, metrics)
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
		{ID, []string{ID}, nil, "RP1"},
		{Tenant, []string{Tenant}, nil, "cgrates.org"},
		{Schedule, []string{Schedule}, nil, "@every 2s"},
		{StatIDs, []string{StatIDs + "[0]"}, nil, "Stat1"},
		{StatIDs, []string{StatIDs + "[1]"}, nil, "Stat2"},
		{MetricIDs, []string{MetricIDs + "[0]"}, nil, "*tcc"},
		{MetricIDs, []string{MetricIDs + "[1]"}, nil, "*acc"},
		{Sorting, []string{Sorting}, nil, "*asc"},
		{Stored, []string{Stored}, nil, false},
		{SortingParameters, []string{SortingParameters + "[0]"}, nil, "*acc"},
		{SortingParameters, []string{SortingParameters + "[1]"}, nil, "*pdd:false"},
		{ThresholdIDs, []string{ThresholdIDs + "[0]"}, nil, "Threshold1"},
		{ThresholdIDs, []string{ThresholdIDs + "[1]"}, nil, "Threshold2"},
		{"NonExistingField", []string{"Field1"}, ErrNotFound, nil},
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

func TestNewRankingSorter(t *testing.T) {
	Metrics := map[string]map[string]float64{
		"STATS1": {"*acc": 12.1, "*tcc": 24.2},
		"STATS2": {"*acc": 12.1, "*tcc": 24.3},
		"STATS3": {"*acc": 10.1, "*tcc": 25.3},
		"STATS4": {"*tcc": 26.3},
	}

	tests := []struct {
		sortingType      string
		sortingParams    []string
		expectErr        bool
		expectSorterType string
	}{
		{
			sortingType:      MetaAsc,
			sortingParams:    []string{"*acc"},
			expectErr:        false,
			expectSorterType: "RankingAscSorter",
		},
		{
			sortingType:      MetaDesc,
			sortingParams:    []string{"*tcc"},
			expectErr:        false,
			expectSorterType: "RankingDescSorter",
		},
		{
			sortingType:      "unsupported",
			sortingParams:    []string{"*tcc"},
			expectErr:        true,
			expectSorterType: "",
		},
	}

	for _, test := range tests {
		rkSorter, err := newRankingSorter(test.sortingType, test.sortingParams, Metrics)

		if test.expectErr {
			if err == nil {
				t.Errorf("Expected an error for sorting type %q, but got none", test.sortingType)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect an error for sorting type %q, but got: %v", test.sortingType, err)
			}
			switch test.sortingType {
			case MetaAsc:
				if _, ok := rkSorter.(*rankingAscSorter); !ok {
					t.Errorf("Expected sorter type 'rankingAscSorter', but got %T", rkSorter)
				}
			case MetaDesc:
				if _, ok := rkSorter.(*rankingDescSorter); !ok {
					t.Errorf("Expected sorter type 'rankingDescSorter', but got %T", rkSorter)
				}
			}
		}
	}
}

func TestRankingProfileSet(t *testing.T) {
	tests := []struct {
		name        string
		path        []string
		val         any
		expectedErr error
		expectedRP  RankingProfile
	}{
		{
			name:        "Set Tenant",
			path:        []string{Tenant},
			val:         "cgrates.org",
			expectedErr: nil,
			expectedRP:  RankingProfile{Tenant: "cgrates.org"},
		},
		{
			name:        "Set ID",
			path:        []string{ID},
			val:         "profile1",
			expectedErr: nil,
			expectedRP:  RankingProfile{ID: "profile1"},
		},
		{
			name:        "Set Schedule",
			path:        []string{Schedule},
			val:         "0 0 * * *",
			expectedErr: nil,
			expectedRP:  RankingProfile{Schedule: "0 0 * * *"},
		},
		{
			name:        "Set StatIDs",
			path:        []string{StatIDs},
			val:         []string{"stat1", "stat2"},
			expectedErr: nil,
			expectedRP:  RankingProfile{StatIDs: []string{"stat1", "stat2"}},
		},
		{
			name:        "Set MetricIDs",
			path:        []string{MetricIDs},
			val:         []string{"metric1", "metric2"},
			expectedErr: nil,
			expectedRP:  RankingProfile{MetricIDs: []string{"metric1", "metric2"}},
		},
		{
			name:        "Set Sorting",
			path:        []string{Sorting},
			val:         "asc",
			expectedErr: nil,
			expectedRP:  RankingProfile{Sorting: "asc"},
		},
		{
			name:        "Set SortingParameters",
			path:        []string{SortingParameters},
			val:         []string{"param1", "param2"},
			expectedErr: nil,
			expectedRP:  RankingProfile{SortingParameters: []string{"param1", "param2"}},
		},
		{
			name:        "Set Stored",
			path:        []string{Stored},
			val:         true,
			expectedErr: nil,
			expectedRP:  RankingProfile{Stored: true},
		},
		{
			name:        "Set ThresholdIDs",
			path:        []string{ThresholdIDs},
			val:         []string{"threshold1", "threshold2"},
			expectedErr: nil,
			expectedRP:  RankingProfile{ThresholdIDs: []string{"threshold1", "threshold2"}},
		},
		{
			name:        "Wrong path",
			path:        []string{"wrongpath"},
			val:         "value",
			expectedErr: ErrWrongPath,
			expectedRP:  RankingProfile{},
		},
		{
			name:        "Empty path",
			path:        []string{},
			val:         "value",
			expectedErr: ErrWrongPath,
			expectedRP:  RankingProfile{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rp := &RankingProfile{}
			err := rp.Set(tt.path, tt.val, false)

			if err != tt.expectedErr {
				t.Errorf("Test %s failed: expected error %v, got %v", tt.name, tt.expectedErr, err)
			}

			if rp.Tenant != tt.expectedRP.Tenant {
				t.Errorf("Test %s failed: expected Tenant %s, got %s", tt.name, tt.expectedRP.Tenant, rp.Tenant)
			}
			if rp.ID != tt.expectedRP.ID {
				t.Errorf("Test %s failed: expected ID %s, got %s", tt.name, tt.expectedRP.ID, rp.ID)
			}
			if rp.Schedule != tt.expectedRP.Schedule {
				t.Errorf("Test %s failed: expected Schedule %s, got %s", tt.name, tt.expectedRP.Schedule, rp.Schedule)
			}

			if !reflect.DeepEqual(rp.StatIDs, tt.expectedRP.StatIDs) {
				t.Errorf("Test %s failed: expected StatIDs %v, got %v", tt.name, tt.expectedRP.StatIDs, rp.StatIDs)
			}
			if !reflect.DeepEqual(rp.MetricIDs, tt.expectedRP.MetricIDs) {
				t.Errorf("Test %s failed: expected MetricIDs %v, got %v", tt.name, tt.expectedRP.MetricIDs, rp.MetricIDs)
			}
			if !reflect.DeepEqual(rp.SortingParameters, tt.expectedRP.SortingParameters) {
				t.Errorf("Test %s failed: expected SortingParameters %v, got %v", tt.name, tt.expectedRP.SortingParameters, rp.SortingParameters)
			}
			if !reflect.DeepEqual(rp.ThresholdIDs, tt.expectedRP.ThresholdIDs) {
				t.Errorf("Test %s failed: expected ThresholdIDs %v, got %v", tt.name, tt.expectedRP.ThresholdIDs, rp.ThresholdIDs)
			}
			if rp.Sorting != tt.expectedRP.Sorting {
				t.Errorf("Test %s failed: expected Sorting %s, got %s", tt.name, tt.expectedRP.Sorting, rp.Sorting)
			}
			if rp.Stored != tt.expectedRP.Stored {
				t.Errorf("Test %s failed: expected Stored %v, got %v", tt.name, tt.expectedRP.Stored, rp.Stored)
			}
		})
	}
}

func TestRankingProfileStringJson(t *testing.T) {
	tests := []struct {
		name         string
		rp           RankingProfile
		expectedJSON string
	}{
		{
			name: "Valid RankingProfile",
			rp: RankingProfile{
				Tenant:            "cgrates.org",
				ID:                "profile1",
				Schedule:          "0 0 * * *",
				StatIDs:           []string{"stat1", "stat2"},
				MetricIDs:         []string{"metric1", "metric2"},
				Sorting:           "asc",
				SortingParameters: []string{"param1", "param2"},
				Stored:            true,
				ThresholdIDs:      []string{"threshold1"},
			},
			expectedJSON: `{"Tenant":"cgrates.org","ID":"profile1","Schedule":"0 0 * * *","StatIDs":["stat1","stat2"],"MetricIDs":["metric1","metric2"],"Sorting":"asc","SortingParameters":["param1","param2"],"Stored":true,"ThresholdIDs":["threshold1"]}`,
		},
		{
			name: "Empty RankingProfile",
			rp: RankingProfile{
				Tenant:            "",
				ID:                "",
				Schedule:          "",
				StatIDs:           []string{},
				MetricIDs:         []string{},
				Sorting:           "",
				SortingParameters: []string{},
				Stored:            false,
				ThresholdIDs:      []string{},
			},
			expectedJSON: `{"Tenant":"","ID":"","Schedule":"","StatIDs":[],"MetricIDs":[],"Sorting":"","SortingParameters":[],"Stored":false,"ThresholdIDs":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rp.String()

			var resultMap map[string]any
			err := json.Unmarshal([]byte(result), &resultMap)
			if err != nil {
				t.Errorf("Error unmarshalling result: %v", err)
			}

			expectedMap := map[string]any{}
			err = json.Unmarshal([]byte(tt.expectedJSON), &expectedMap)
			if err != nil {
				t.Errorf("Error unmarshalling expected JSON: %v", err)
			}

			for key, value1 := range resultMap {
				if value2, exists := expectedMap[key]; exists {
					if value1Slice, ok1 := value1.([]any); ok1 {
						if value2Slice, ok2 := value2.([]any); ok2 {
							if len(value1Slice) != len(value2Slice) {
								t.Errorf("Test %s failed: slice length mismatch for key %s", tt.name, key)
							}
							for i, v1 := range value1Slice {
								if v1 != value2Slice[i] {
									t.Errorf("Test %s failed: slice mismatch for key %s at index %d", tt.name, key, i)
								}
							}
						}
					} else {
						if value1 != value2 {
							t.Errorf("Test %s failed: expected %v for key %s, got %v", tt.name, value2, key, value1)
						}
					}
				} else {
					t.Errorf("Test %s failed: key %s not found in expected result", tt.name, key)
				}
			}
		})
	}
}

func TestTpRankingProfileFieldAsString(t *testing.T) {
	tests := []struct {
		name      string
		profile   RankingProfile
		fldPath   []string
		expected  string
		expectErr bool
	}{
		{
			name: "Valid field path",
			profile: RankingProfile{
				Tenant:            "cgrates.org",
				ID:                "profile1",
				Schedule:          "0 0 * * *",
				StatIDs:           []string{"stat1", "stat2"},
				MetricIDs:         []string{"metric1", "metric2"},
				SortingParameters: []string{"param1", "param2"},
				ThresholdIDs:      []string{"threshold1", "threshold2"},
			},
			fldPath:   []string{"Tenant"},
			expected:  "cgrates.org",
			expectErr: false,
		},
		{
			name: "Invalid field path",
			profile: RankingProfile{
				Tenant:            "cgrates.org",
				ID:                "profile1",
				Schedule:          "0 0 * * *",
				StatIDs:           []string{"stat1", "stat2"},
				MetricIDs:         []string{"metric1", "metric2"},
				SortingParameters: []string{"param1", "param2"},
				ThresholdIDs:      []string{"threshold1", "threshold2"},
			},
			fldPath:   []string{"NonExistentField"},
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.profile.FieldAsString(tt.fldPath)

			if tt.expectErr && err == nil {
				t.Errorf("Expected an error for test %s, but got none", tt.name)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for test %s: %v", tt.name, err)
			}

			if result != tt.expected {
				t.Errorf("Test %s failed: expected %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}
