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

	"github.com/cgrates/cgrates/utils"
)

func TestRankingDescSorterSortStatIDs(t *testing.T) {
	Metrics := map[string]map[string]float64{
		"STATS1": {"*acc": 12.1, "*tcc": 24.2},
		"STATS2": {"*acc": 12.1, "*tcc": 24.3},
		"STATS3": {"*acc": 10.1, "*tcc": 25.3},
		"STATS4": {"*tcc": 26.3},
	}
	sortMetrics := []string{"*acc", "*tcc"}
	rdscSrtr := newRankingDescSorter(sortMetrics, Metrics)
	eStatIDs := []string{"STATS2", "STATS1", "STATS3", "STATS4"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*acc:false", "*tcc"} // changed the order of checks, stats4 should come first
	rdscSrtr = newRankingDescSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS3", "STATS2", "STATS1", "STATS4"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc", "*acc:true"} // changed the order of checks, stats4 should come first
	rdscSrtr = newRankingDescSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS4", "STATS3", "STATS2", "STATS1"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc:false", "*acc"} // reversed *tcc which should consider ascendent instead of descendent
	rdscSrtr = newRankingDescSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS1", "STATS2", "STATS3", "STATS4"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
}

func TestRankingAscSorterSortStatIDs(t *testing.T) {
	Metrics := map[string]map[string]float64{
		"STATS1": {"*acc": 12.1, "*tcc": 24.2},
		"STATS2": {"*acc": 12.1, "*tcc": 24.3},
		"STATS3": {"*acc": 10.1, "*tcc": 25.3},
		"STATS4": {"*tcc": 26.3},
	}
	sortMetrics := []string{"*acc", "*tcc"}
	rtAscSrtr := newRankingAscSorter(sortMetrics, Metrics)
	eStatIDs := []string{"STATS3", "STATS1", "STATS2", "STATS4"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*acc:false", "*tcc"}
	rtAscSrtr = newRankingAscSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS1", "STATS2", "STATS3", "STATS4"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc", "*acc:true"}
	rtAscSrtr = newRankingAscSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS1", "STATS2", "STATS3", "STATS4"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc:false", "*acc"}
	rtAscSrtr = newRankingAscSorter(sortMetrics, Metrics)
	eStatIDs = []string{"STATS4", "STATS3", "STATS2", "STATS1"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
}

func TestTenantID(t *testing.T) {
	ranking := &Ranking{
		Tenant: "tenant1",
		ID:     "ranking1",
	}
	expectedTenantID := utils.ConcatenatedKey(ranking.Tenant, ranking.ID)
	actualTenantID := ranking.TenantID()
	if actualTenantID != expectedTenantID {
		t.Errorf("Expected TenantID %q, got %q", expectedTenantID, actualTenantID)
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
			sortingType:      utils.MetaAsc,
			sortingParams:    []string{"*acc"},
			expectErr:        false,
			expectSorterType: "RankingAscSorter",
		},
		{
			sortingType:      utils.MetaDesc,
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
			case utils.MetaAsc:
				if _, ok := rkSorter.(*rankingAscSorter); !ok {
					t.Errorf("Expected sorter type 'rankingAscSorter', but got %T", rkSorter)
				}
			case utils.MetaDesc:
				if _, ok := rkSorter.(*rankingDescSorter); !ok {
					t.Errorf("Expected sorter type 'rankingDescSorter', but got %T", rkSorter)
				}
			}
		}
	}
}

func TestRankingProfileClone(t *testing.T) {

	t.Run("Empty fields", func(t *testing.T) {
		original := &RankingProfile{}

		clone := original.Clone()

		if clone.Tenant != "" ||
			clone.ID != "" ||
			clone.Schedule != "" ||
			clone.StatIDs != nil ||
			clone.MetricIDs != nil ||
			clone.Sorting != "" ||
			clone.SortingParameters != nil ||
			clone.Stored != false ||
			clone.ThresholdIDs != nil {
			t.Errorf("Clone method did not create an empty RankingProfile for empty original")
		}
	})

	t.Run("Nil slices", func(t *testing.T) {
		original := &RankingProfile{
			Tenant:   "tenant",
			ID:       "profile_id",
			Schedule: "0 * * * *",
		}

		clone := original.Clone()

		if clone.StatIDs != nil ||
			clone.MetricIDs != nil ||
			clone.SortingParameters != nil ||
			clone.ThresholdIDs != nil {
			t.Errorf("Clone method did not handle nil slices correctly")
		}
	})
}

func TestNewRankingFromProfile(t *testing.T) {
	rkP := &RankingProfile{
		Tenant:            "tenant",
		ID:                "profile_id",
		Schedule:          "0 * * * *",
		StatIDs:           []string{"stat1", "stat2"},
		MetricIDs:         []string{"metricA", "metricB"},
		Sorting:           "asc",
		SortingParameters: []string{"metricA:true", "metricB:false"},
		Stored:            true,
		ThresholdIDs:      []string{"threshold1", "threshold2"},
	}

	expectedRk := &Ranking{
		Tenant:    rkP.Tenant,
		ID:        rkP.ID,
		Sorting:   rkP.Sorting,
		Metrics:   make(map[string]map[string]float64),
		rkPrfl:    rkP,
		metricIDs: utils.NewStringSet(rkP.MetricIDs),
	}

	rk := NewRankingFromProfile(rkP)

	if rk.Tenant != expectedRk.Tenant ||
		rk.ID != expectedRk.ID ||
		rk.Sorting != expectedRk.Sorting ||
		rk.rkPrfl != expectedRk.rkPrfl ||
		!reflect.DeepEqual(rk.metricIDs, expectedRk.metricIDs) {
		t.Errorf("NewRankingFromProfile returned unexpected Ranking object")
	}

}

func TestRankingSortStatss(t *testing.T) {
	metrics := map[string]map[string]float64{
		"STAT1": {"metric1": 10.1, "metric2": 5.2},
		"STAT2": {"metric1": 9.1, "metric2": 6.2},
		"STAT3": {"metric1": 11.1, "metric2": 4.2},
	}

	_, err := rankingSortStats("valid", []string{"metric1"}, metrics)
	if err != nil {
		expectedErr := "NOT_IMPLEMENTED:valid"
		if err.Error() != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	}

	_, err = rankingSortStats("invalid", []string{"metric1"}, metrics)
	if err == nil {
		t.Errorf("expected an error for invalid sorting type, but got nil")
	}
}

func TestRankingAscSorterEmptyStatIDs(t *testing.T) {
	rkASrtr := &rankingAscSorter{
		statIDs:    []string{},
		Metrics:    make(map[string]map[string]float64),
		sMetricIDs: []string{},
		sMetricRev: utils.StringSet{},
	}

	sortedStatIDs := rkASrtr.sortStatIDs()

	if len(sortedStatIDs) != 0 {
		t.Errorf("expected sortedStatIDs to be empty, got %v", sortedStatIDs)
	}
}

func TestRankingDescSorterEmptyStatIDs(t *testing.T) {
	rkDsrtr := &rankingDescSorter{
		statIDs:    []string{},
		Metrics:    make(map[string]map[string]float64),
		sMetricIDs: []string{},
		sMetricRev: utils.StringSet{},
	}

	sortedStatIDs := rkDsrtr.sortStatIDs()

	if len(sortedStatIDs) != 0 {
		t.Errorf("expected sortedStatIDs to be empty, got %v", sortedStatIDs)
	}
}
