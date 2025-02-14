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
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type RankingProfileWithAPIOpts struct {
	*RankingProfile
	APIOpts map[string]any
}

type RankingProfile struct {
	Tenant            string   // Tenant this profile belongs to
	ID                string   // Profile identification
	Schedule          string   // Cron schedule this profile should run at
	StatIDs           []string // List of stat instances to query
	MetricIDs         []string // Filter out only specific metrics in reply for sorting
	Sorting           string   // Sorting strategy. Possible values: <*asc|*desc>
	SortingParameters []string // Sorting parameters: depending on sorting type, list of metric ids for now with optional true or false in case of reverse logic is desired
	Stored            bool     // Offline storage activation for this profile
	ThresholdIDs      []string // List of threshold IDs to limit this Ranking to. *none to disable threshold processing for it.
}

func (sgp *RankingProfile) TenantID() string {
	return utils.ConcatenatedKey(sgp.Tenant, sgp.ID)
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
		cln.StatIDs = make([]string, len(rkP.StatIDs))
		copy(cln.StatIDs, rkP.StatIDs)
	}
	if rkP.MetricIDs != nil {
		cln.MetricIDs = make([]string, len(rkP.MetricIDs))
		copy(cln.MetricIDs, rkP.MetricIDs)
	}
	if rkP.SortingParameters != nil {

		cln.SortingParameters = make([]string, len(rkP.SortingParameters))
		copy(cln.SortingParameters, rkP.SortingParameters)
	}
	if rkP.ThresholdIDs != nil {
		cln.ThresholdIDs = make([]string, len(rkP.ThresholdIDs))
		copy(cln.ThresholdIDs, rkP.ThresholdIDs)
	}
	return
}

// rankingProfileLockKey returns the ID used to lock a RankingProfile with guardian
func rankingProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheRankingProfiles, tnt, id)
}
func NewRankingFromProfile(rkP *RankingProfile) (rk *Ranking) {
	rk = &Ranking{
		Tenant:  rkP.Tenant,
		ID:      rkP.ID,
		Sorting: rkP.Sorting,
		Metrics: make(map[string]map[string]float64),

		rkPrfl:    rkP,
		metricIDs: utils.NewStringSet(rkP.MetricIDs),
	}
	if rkP.SortingParameters != nil {
		rk.SortingParameters = make([]string, len(rkP.SortingParameters))
		copy(rk.SortingParameters, rkP.SortingParameters)
	}
	return
}

type RankingWithAPIOpts struct {
	*Ranking
	APIOpts map[string]any
}

// Ranking is one unit out of a profile
type Ranking struct {
	rMux sync.RWMutex

	Tenant            string
	ID                string
	LastUpdate        time.Time
	Metrics           map[string]map[string]float64 // map[statID]map[metricID]metricValue
	Sorting           string
	SortingParameters []string

	SortedStatIDs []string

	rkPrfl    *RankingProfile // store here the ranking profile so we can have it at hands further
	metricIDs utils.StringSet // convert the metricIDs here for faster matching

}

func (r *Ranking) TenantID() string {
	return utils.ConcatenatedKey(r.Tenant, r.ID)
}

// asRankingSummary converts the Ranking instance into a RankingSummary one
func (rk *Ranking) asRankingSummary() (rkSm *RankingSummary) {
	rkSm = &RankingSummary{
		Tenant:     rk.Tenant,
		ID:         rk.ID,
		LastUpdate: rk.LastUpdate,
	}
	rkSm.SortedStatIDs = make([]string, len(rk.SortedStatIDs))
	copy(rkSm.SortedStatIDs, rk.SortedStatIDs)
	return
}

type rankingSorter interface {
	sortStatIDs() []string // sortStatIDs returns the sorted list of statIDs
}

// rankingSortStats will return the list of sorted statIDs out of the sortingData map
func rankingSortStats(sortingType string, sortingParams []string,
	Metrics map[string]map[string]float64) (sortedStatIDs []string, err error) {
	var rnkSrtr rankingSorter
	if rnkSrtr, err = newRankingSorter(sortingType, sortingParams, Metrics); err != nil {
		return
	}
	return rnkSrtr.sortStatIDs(), nil
}

// newRankingSorter is the constructor for various ranking sorters
//
//	returns error if the sortingType is not implemented
func newRankingSorter(sortingType string, sortingParams []string,
	Metrics map[string]map[string]float64) (rkStr rankingSorter, err error) {
	switch sortingType {
	default:
		err = utils.ErrPrefixNotErrNotImplemented(sortingType)
		return
	case utils.MetaDesc:
		return newRankingDescSorter(sortingParams, Metrics), nil
	case utils.MetaAsc:
		return newRankingAscSorter(sortingParams, Metrics), nil
	}
}

// newRankingDescSorter is a constructor for rankingDescSorter
func newRankingDescSorter(sortingParams []string,
	Metrics map[string]map[string]float64) (rkDsrtr *rankingDescSorter) {
	clnSp := make([]string, len(sortingParams))
	sPReversed := make(utils.StringSet)
	for i, sP := range sortingParams { // clean the sortingParams, out of param:false or param:true definitions
		sPSlc := strings.Split(sP, utils.InInFieldSep)
		clnSp[i] = sPSlc[0]
		if len(sPSlc) > 1 && sPSlc[1] == utils.FalseStr {
			sPReversed.Add(sPSlc[0]) // param defined as param:false which should be added to reversing comparison
		}
	}
	rkDsrtr = &rankingDescSorter{
		clnSp,
		sPReversed,
		Metrics,
		make([]string, 0, len(Metrics))}
	for statID := range rkDsrtr.Metrics {
		rkDsrtr.statIDs = append(rkDsrtr.statIDs, statID)
	}
	return
}

// rankingDescSorter will sort data descendent for metrics in sortingParams or random if all equal
type rankingDescSorter struct {
	sMetricIDs []string
	sMetricRev utils.StringSet // list of exceptios for sortingParams, reverting the sorting logic
	Metrics    map[string]map[string]float64

	statIDs []string // list of keys of the Metrics
}

// sortStatIDs implements rankingSorter interface
func (rkDsrtr *rankingDescSorter) sortStatIDs() []string {
	if len(rkDsrtr.statIDs) == 0 {
		return rkDsrtr.statIDs
	}
	sort.Slice(rkDsrtr.statIDs, func(i, j int) bool {
		for _, metricID := range rkDsrtr.sMetricIDs {
			val1, hasMetric1 := rkDsrtr.Metrics[rkDsrtr.statIDs[i]][metricID]
			val2, hasMetric2 := rkDsrtr.Metrics[rkDsrtr.statIDs[j]][metricID]
			if !hasMetric1 && !hasMetric2 {
				continue
			}
			if !hasMetric1 {
				return false
			}
			if !hasMetric2 {
				return true
			}
			//in case we have the same value for the current metricID we skip to the next one
			if val1 == val2 {
				continue
			}
			ret := val1 > val2
			if rkDsrtr.sMetricRev.Has(metricID) {
				ret = !ret
			}
			return ret
		}
		//in case that we have the same value for all params we return randomly
		return utils.BoolGenerator().RandomBool()
	})
	return rkDsrtr.statIDs
}

// newRankingAscSorter is a constructor for rankingAscSorter
func newRankingAscSorter(sortingParams []string,
	Metrics map[string]map[string]float64) (rkASrtr *rankingAscSorter) {
	clnSp := make([]string, len(sortingParams))
	sPReversed := make(utils.StringSet)
	for i, sP := range sortingParams { // clean the sortingParams, out of param:false or param:true definitions
		sPSlc := strings.Split(sP, utils.InInFieldSep)
		clnSp[i] = sPSlc[0]
		if len(sPSlc) > 1 && sPSlc[1] == utils.FalseStr {
			sPReversed.Add(sPSlc[0]) // param defined as param:false which should be added to reversing comparison
		}
	}
	rkASrtr = &rankingAscSorter{
		clnSp,
		sPReversed,
		Metrics,
		make([]string, 0, len(Metrics))}
	for statID := range rkASrtr.Metrics {
		rkASrtr.statIDs = append(rkASrtr.statIDs, statID)
	}
	return
}

// rankingAscSorter will sort data ascendent for metrics in sortingParams or randomly if all equal
type rankingAscSorter struct {
	sMetricIDs []string
	sMetricRev utils.StringSet // list of exceptios for sortingParams, reverting the sorting logic
	Metrics    map[string]map[string]float64

	statIDs []string // list of keys of the Metrics
}

// sortStatIDs implements rankingSorter interface
func (rkASrtr *rankingAscSorter) sortStatIDs() []string {
	if len(rkASrtr.statIDs) == 0 {
		return rkASrtr.statIDs
	}
	sort.Slice(rkASrtr.statIDs, func(i, j int) bool {
		for _, metricID := range rkASrtr.sMetricIDs {
			val1, hasMetric1 := rkASrtr.Metrics[rkASrtr.statIDs[i]][metricID]
			val2, hasMetric2 := rkASrtr.Metrics[rkASrtr.statIDs[j]][metricID]
			if !hasMetric1 && !hasMetric2 {
				continue
			}
			if !hasMetric1 {
				return false
			}
			if !hasMetric2 {
				return true
			}
			//in case we have the same value for the current metricID we skip to the next one
			if val1 == val2 {
				continue
			}
			ret := val2 > val1
			if rkASrtr.sMetricRev.Has(metricID) {
				ret = !ret // reversed logic in case of metric:false in params
			}
			return ret
		}
		//in case that we have the same value for all params we return randomly
		return utils.BoolGenerator().RandomBool()
	})
	return rkASrtr.statIDs
}

// RankingSummary is the event sent to TrendS and EEs
type RankingSummary struct {
	Tenant        string
	ID            string
	LastUpdate    time.Time
	SortedStatIDs []string
}

func (tp *RankingProfile) Set(path []string, val any, _ bool) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}

	switch path[0] {
	default:
		return utils.ErrWrongPath
	case utils.Tenant:
		tp.Tenant = utils.IfaceAsString(val)
	case utils.ID:
		tp.ID = utils.IfaceAsString(val)
	case utils.Schedule:
		tp.Schedule = utils.IfaceAsString(val)
	case utils.StatIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.StatIDs = append(tp.StatIDs, valA...)
	case utils.MetricIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.MetricIDs = append(tp.MetricIDs, valA...)
	case utils.Sorting:
		tp.Sorting = utils.IfaceAsString(val)
	case utils.SortingParameters:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.SortingParameters = append(tp.SortingParameters, valA...)
	case utils.Stored:
		tp.Stored, err = utils.IfaceAsBool(val)
	case utils.ThresholdIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.ThresholdIDs = append(tp.ThresholdIDs, valA...)
	}
	return
}

func (tp *RankingProfile) Merge(v2 any) {
	vi := v2.(*RankingProfile)
	if len(vi.Tenant) != 0 {
		tp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		tp.ID = vi.ID
	}
	if len(vi.Schedule) != 0 {
		tp.Schedule = vi.Schedule
	}
	tp.StatIDs = append(tp.StatIDs, vi.StatIDs...)
	tp.MetricIDs = append(tp.MetricIDs, vi.MetricIDs...)
	tp.SortingParameters = append(tp.SortingParameters, vi.SortingParameters...)
	tp.ThresholdIDs = append(tp.ThresholdIDs, vi.ThresholdIDs...)
	if len(vi.Sorting) != 0 {
		tp.Sorting = vi.Sorting
	}
	if vi.Stored {
		tp.Stored = vi.Stored
	}
}

func (tp *RankingProfile) String() string { return utils.ToJSON(tp) }
func (tp *RankingProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = tp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (tp *RankingProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case utils.StatIDs:
				if *idx < len(tp.StatIDs) {
					return tp.StatIDs[*idx], nil
				}
			case utils.MetricIDs:
				if *idx < len(tp.MetricIDs) {
					return tp.MetricIDs[*idx], nil
				}
			case utils.SortingParameters:
				if *idx < len(tp.SortingParameters) {
					return tp.SortingParameters[*idx], nil
				}
			case utils.ThresholdIDs:
				if *idx < len(tp.ThresholdIDs) {
					return tp.ThresholdIDs[*idx], nil
				}
			}
		}
		return nil, utils.ErrNotFound
	case utils.Tenant:
		return tp.Tenant, nil
	case utils.ID:
		return tp.ID, nil
	case utils.Schedule:
		return tp.Schedule, nil
	case utils.Sorting:
		return tp.Sorting, nil
	case utils.Stored:
		return tp.Stored, nil
	}
}
