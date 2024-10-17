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
)

func TestRankingDescSorterSortStatIDs(t *testing.T) {
	statMetrics := map[string]map[string]float64{
		"STATS1": {"*acc": 12.1, "*tcc": 24.2},
		"STATS2": {"*acc": 12.1, "*tcc": 24.3},
		"STATS3": {"*acc": 10.1, "*tcc": 25.3},
		"STATS4": {"*tcc": 26.3},
	}
	sortMetrics := []string{"*acc", "*tcc"}
	rdscSrtr := newRankingDescSorter(sortMetrics, statMetrics)
	eStatIDs := []string{"STATS2", "STATS1", "STATS3", "STATS4"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc", "*acc"} // changed the order of checks, stats4 should come first
	rdscSrtr = newRankingDescSorter(sortMetrics, statMetrics)
	eStatIDs = []string{"STATS4", "STATS3", "STATS2", "STATS1"}
	if statIDs := rdscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
}

func TestRankingAscSorterSortStatIDs(t *testing.T) {
	statMetrics := map[string]map[string]float64{
		"STATS1": {"*acc": 12.1, "*tcc": 24.2},
		"STATS2": {"*acc": 12.1, "*tcc": 24.3},
		"STATS3": {"*acc": 10.1, "*tcc": 25.3},
		"STATS4": {"*tcc": 26.3},
	}
	sortMetrics := []string{"*acc", "*tcc"}
	rtAscSrtr := newRankingAscSorter(sortMetrics, statMetrics)
	eStatIDs := []string{"STATS3", "STATS1", "STATS2", "STATS4"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
	sortMetrics = []string{"*tcc", "*acc"}
	rtAscSrtr = newRankingAscSorter(sortMetrics, statMetrics)
	eStatIDs = []string{"STATS1", "STATS2", "STATS3", "STATS4"}
	if statIDs := rtAscSrtr.sortStatIDs(); !reflect.DeepEqual(eStatIDs, statIDs) {
		t.Errorf("Expecting: %v, received %v", eStatIDs, statIDs)
	}
}
