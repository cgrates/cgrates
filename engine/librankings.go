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
	"sort"
	"sync"

	"github.com/cgrates/cgrates/utils"
)

type RankingProfileWithAPIOpts struct {
	*RankingProfile
	APIOpts map[string]any
}

type RankingProfile struct {
	Tenant            string
	ID                string
	Schedule          string
	StatIDs           []string
	MetricIDs         []string
	Sorting           string
	SortingParameters []string
	Stored            bool
	ThresholdIDs      []string
}

func (rkp *RankingProfile) TenantID() string {
	return utils.ConcatenatedKey(rkp.Tenant, rkp.ID)
}

// Clone will clone a RankingProfile
func (rkP *RankingProfile) Clone() (cln *RankingProfile) {
	cln = &RankingProfile{
		Tenant:   rkP.Tenant,
		ID:       rkP.ID,
		Schedule: rkP.Schedule,
		Sorting:  rkP.Sorting,
	}
	if rkP.StatIDs != nil {
		copy(cln.StatIDs, rkP.StatIDs)
	}
	if rkP.MetricIDs != nil {
		copy(cln.MetricIDs, rkP.MetricIDs)
	}
	if rkP.SortingParameters != nil {
		copy(cln.SortingParameters, rkP.SortingParameters)
	}
	if rkP.ThresholdIDs != nil {
		copy(cln.ThresholdIDs, rkP.ThresholdIDs)
	}
	return
}

// NewRankingFromProfile is a constructor for an empty ranking out of it's profile
func NewRankingFromProfile(rkP *RankingProfile) *Ranking {
	return &Ranking{
		Tenant:      rkP.Tenant,
		ID:          rkP.ID,
		StatMetrics: make(map[string]map[string]float64),

		rkPrfl:    rkP,
		metricIDs: utils.NewStringSet(rkP.MetricIDs),
	}
}

// Ranking is one unit out of a profile
type Ranking struct {
	rMux sync.RWMutex

	Tenant            string
	ID                string
	StatMetrics       map[string]map[string]float64 // map[statID]map[metricID]metricValue
	Sorting           string
	SortingParameters []string

	SortedStatIDs []string

	rkPrfl    *RankingProfile // store here the ranking profile so we can have it at hands further
	metricIDs utils.StringSet // convert the metricIDs here for faster matching

}

type rankingSorter interface {
	sortStatIDs() []string // sortStatIDs returns the sorted list of statIDs
}

// rankingSortStats will return the list of sorted statIDs out of the sortingData map
func rankingSortStats(sortingType string, sortingParams []string,
	statMetrics map[string]map[string]float64) (sortedStatIDs []string, err error) {
	var rnkSrtr rankingSorter
	if rnkSrtr, err = newRankingSorter(sortingType, sortingParams, statMetrics); err != nil {
		return
	}
	return rnkSrtr.sortStatIDs(), nil
}

// newRankingSorter is the constructor for various ranking sorters
//
//	returns error if the sortingType is not implemented
func newRankingSorter(sortingType string, sortingParams []string,
	statMetrics map[string]map[string]float64) (rkStr rankingSorter, err error) {
	switch sortingType {
	default:
		err = utils.ErrPrefixNotErrNotImplemented(sortingType)
		return
	case utils.MetaDesc:
		return newRankingDescSorter(sortingParams, statMetrics), nil
	}
	return
}

// newRankingDescSorter is a constructor for rankingDescSorter
func newRankingDescSorter(sortingParams []string,
	statMetrics map[string]map[string]float64) (rkDsrtr *rankingDescSorter) {
	rkDsrtr = &rankingDescSorter{
		sortingParams,
		statMetrics,
		make([]string, 0, len(statMetrics))}
	for statID := range rkDsrtr.statMetrics {
		rkDsrtr.statIDs = append(rkDsrtr.statIDs, statID)
	}
	return
}

// rankingDescSorter will sort data descendently for metrics in sortingParams or randomly if all equal
type rankingDescSorter struct {
	sortingParams []string
	statMetrics   map[string]map[string]float64

	statIDs []string // list of keys of the statMetrics
}

// sortStatIDs implements rankingSorter interface
func (rkDsrtr *rankingDescSorter) sortStatIDs() []string {
	if len(rkDsrtr.statIDs) == 0 {
		return rkDsrtr.statIDs
	}
	sort.Slice(rkDsrtr.statIDs, func(i, j int) bool {
		for _, metricID := range rkDsrtr.sortingParams {
			val1, hasMetric1 := rkDsrtr.statMetrics[rkDsrtr.statIDs[i]][metricID]
			if !hasMetric1 {
				return false
			}
			val2, hasMetric2 := rkDsrtr.statMetrics[rkDsrtr.statIDs[j]][metricID]
			if !hasMetric2 {
				return true
			}
			//in case we have the same value for the current metricID we skip to the next one
			if val1 == val2 {
				continue
			}
			if val1 > val2 {
				return true
			}
			return false
		}
		//in case that we have the same value for all params we return randomly
		return utils.BoolGenerator().RandomBool()
	})
	return rkDsrtr.statIDs
}
