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
	"sort"
	"strings"
	"sync"
	"time"
)

// RankingProfile defines the configuration of the Ranking.
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

// RankingWithAPIOpts wraps Ranking with APIOpts.
type RankingProfileWithAPIOpts struct {
	*RankingProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (sgp *RankingProfile) TenantID() string {
	return ConcatenatedKey(sgp.Tenant, sgp.ID)
}

// Clone creates a deep copy of RankingProfile for thread-safe use.
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

// Set implements the profile interface, setting values in RankingProfile based on path.
func (rp *RankingProfile) Set(path []string, val any, _ bool) (err error) {
	if len(path) != 1 {
		return ErrWrongPath
	}

	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		rp.Tenant = IfaceAsString(val)
	case ID:
		rp.ID = IfaceAsString(val)
	case Schedule:
		rp.Schedule = IfaceAsString(val)
	case StatIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.StatIDs = append(rp.StatIDs, valA...)
	case MetricIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.MetricIDs = append(rp.MetricIDs, valA...)
	case Sorting:
		rp.Sorting = IfaceAsString(val)
	case SortingParameters:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.SortingParameters = append(rp.SortingParameters, valA...)
	case Stored:
		rp.Stored, err = IfaceAsBool(val)
	case ThresholdIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.ThresholdIDs = append(rp.ThresholdIDs, valA...)
	}
	return
}

// Merge implements the profile interface, merging values from another RankingProfile.
func (rp *RankingProfile) Merge(v2 any) {
	vi := v2.(*RankingProfile)
	if len(vi.Tenant) != 0 {
		rp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		rp.ID = vi.ID
	}
	if len(vi.Schedule) != 0 {
		rp.Schedule = vi.Schedule
	}
	rp.StatIDs = append(rp.StatIDs, vi.StatIDs...)
	rp.MetricIDs = append(rp.MetricIDs, vi.MetricIDs...)
	rp.SortingParameters = append(rp.SortingParameters, vi.SortingParameters...)
	rp.ThresholdIDs = append(rp.ThresholdIDs, vi.ThresholdIDs...)
	if len(vi.Sorting) != 0 {
		rp.Sorting = vi.Sorting
	}
	if vi.Stored {
		rp.Stored = vi.Stored
	}
}

// String implements the DataProvider interface, returning the RankingProfile in JSON format.
func (rp *RankingProfile) String() string { return ToJSON(rp) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (rp *RankingProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (rp *RankingProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case StatIDs:
				if *idx < len(rp.StatIDs) {
					return rp.StatIDs[*idx], nil
				}
			case MetricIDs:
				if *idx < len(rp.MetricIDs) {
					return rp.MetricIDs[*idx], nil
				}
			case SortingParameters:
				if *idx < len(rp.SortingParameters) {
					return rp.SortingParameters[*idx], nil
				}
			case ThresholdIDs:
				if *idx < len(rp.ThresholdIDs) {
					return rp.ThresholdIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return rp.Tenant, nil
	case ID:
		return rp.ID, nil
	case Schedule:
		return rp.Schedule, nil
	case Sorting:
		return rp.Sorting, nil
	case Stored:
		return rp.Stored, nil
	}
}

// RankingProfileLockKey returns the ID used to lock a RankingProfile with guardian.
func RankingProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheRankingProfiles, tnt, id)
}

// NewRankingFromProfile creates a new Ranking based on profile configuration.
func NewRankingFromProfile(rkP *RankingProfile) (rk *Ranking) {
	rk = &Ranking{
		Tenant:  rkP.Tenant,
		ID:      rkP.ID,
		Sorting: rkP.Sorting,
		Metrics: make(map[string]map[string]float64),

		rkPrfl:    rkP,
		metricIDs: NewStringSet(rkP.MetricIDs),
	}
	if rkP.SortingParameters != nil {
		rk.SortingParameters = make([]string, len(rkP.SortingParameters))
		copy(rk.SortingParameters, rkP.SortingParameters)
	}
	return
}

// Ranking represents a collection of metrics with ranked statistics.
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
	metricIDs StringSet       // convert the metricIDs here for faster matching

}

// RankingWithAPIOpts wraps Ranking with APIOpts.
type RankingWithAPIOpts struct {
	*Ranking
	APIOpts map[string]any
}

// RankingSummary holds the most recent ranking metrics.
type RankingSummary struct {
	Tenant        string
	ID            string
	LastUpdate    time.Time
	SortedStatIDs []string
}

// TenantID returns the concatenated tenant and ID.
func (r *Ranking) TenantID() string {
	return ConcatenatedKey(r.Tenant, r.ID)
}

// AsRankingSummary creates a summary with the most recent ranking data.
func (r *Ranking) AsRankingSummary() (rkSm *RankingSummary) {
	rkSm = &RankingSummary{
		Tenant:     r.Tenant,
		ID:         r.ID,
		LastUpdate: r.LastUpdate,
	}
	rkSm.SortedStatIDs = make([]string, len(r.SortedStatIDs))
	copy(rkSm.SortedStatIDs, r.SortedStatIDs)
	return
}

// Config returns the ranking's profile configuration.
func (r *Ranking) Config() *RankingProfile {
	return r.rkPrfl
}

// SetConfig sets the ranking's profile configuration.
func (r *Ranking) SetConfig(rp *RankingProfile) {
	r.rkPrfl = rp
}

// Lock locks the ranking mutex.
func (r *Ranking) Lock() {
	r.rMux.Lock()
}

// Unlock unlocks the ranking mutex.
func (r *Ranking) Unlock() {
	r.rMux.Unlock()
}

// RLock locks the ranking mutex for reading.
func (r *Ranking) RLock() {
	r.rMux.RLock()
}

// RUnlock unlocks the read lock on the ranking mutex.
func (r *Ranking) RUnlock() {
	r.rMux.RUnlock()
}

// MetricIDs returns the set of metric IDs for this ranking.
func (r *Ranking) MetricIDs() StringSet {
	return r.metricIDs
}

// rankingSorter defines interface for different ranking sorting strategies.
type rankingSorter interface {
	sortStatIDs() []string // sortStatIDs returns the sorted list of statIDs
}

// newRankingSorter is the constructor for various ranking sorters.
// Returns error if the sortingType is not implemented.
func newRankingSorter(sortingType string, sortingParams []string,
	Metrics map[string]map[string]float64) (rkStr rankingSorter, err error) {
	switch sortingType {
	default:
		err = ErrPrefixNotErrNotImplemented(sortingType)
		return
	case MetaDesc:
		return newRankingDescSorter(sortingParams, Metrics), nil
	case MetaAsc:
		return newRankingAscSorter(sortingParams, Metrics), nil
	}
}

// RankingSortStats sorts stat IDs based on their metrics according to the specified sorting strategy.
func RankingSortStats(sortingType string, sortingParams []string,
	Metrics map[string]map[string]float64) (sortedStatIDs []string, err error) {
	var rnkSrtr rankingSorter
	if rnkSrtr, err = newRankingSorter(sortingType, sortingParams, Metrics); err != nil {
		return
	}
	return rnkSrtr.sortStatIDs(), nil
}

// rankingDescSorter sorts data in descending order for metrics in sortingParams or randomly if all equal.
type rankingDescSorter struct {
	sMetricIDs []string
	sMetricRev StringSet // list of exceptios for sortingParams, reverting the sorting logic
	Metrics    map[string]map[string]float64

	statIDs []string // list of keys of the Metrics
}

// newRankingDescSorter is a constructor for rankingDescSorter
func newRankingDescSorter(sortingParams []string,
	Metrics map[string]map[string]float64) (rkDsrtr *rankingDescSorter) {
	clnSp := make([]string, len(sortingParams))
	sPReversed := make(StringSet)
	for i, sP := range sortingParams { // clean the sortingParams, out of param:false or param:true definitions
		sPSlc := strings.Split(sP, InInFieldSep)
		clnSp[i] = sPSlc[0]
		if len(sPSlc) > 1 && sPSlc[1] == FalseStr {
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

// sortStatIDs implements rankingSorter interface.
func (s *rankingDescSorter) sortStatIDs() []string {
	if len(s.statIDs) == 0 {
		return s.statIDs
	}
	sort.Slice(s.statIDs, func(i, j int) bool {
		for _, metricID := range s.sMetricIDs {
			val1, hasMetric1 := s.Metrics[s.statIDs[i]][metricID]
			val2, hasMetric2 := s.Metrics[s.statIDs[j]][metricID]
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
			if s.sMetricRev.Has(metricID) {
				ret = !ret
			}
			return ret
		}
		//in case that we have the same value for all params we return randomly
		return BoolGenerator().RandomBool()
	})
	return s.statIDs
}

// rankingAscSorter sorts data in ascending order for metrics in sortingParams or randomly if all equal.
type rankingAscSorter struct {
	sMetricIDs []string
	sMetricRev StringSet // list of exceptios for sortingParams, reverting the sorting logic
	Metrics    map[string]map[string]float64

	statIDs []string // list of keys of the Metrics
}

// newRankingAscSorter is a constructor for rankingAscSorter.
func newRankingAscSorter(sortingParams []string,
	Metrics map[string]map[string]float64) (rkASrtr *rankingAscSorter) {
	clnSp := make([]string, len(sortingParams))
	sPReversed := make(StringSet)
	for i, sP := range sortingParams { // clean the sortingParams, out of param:false or param:true definitions
		sPSlc := strings.Split(sP, InInFieldSep)
		clnSp[i] = sPSlc[0]
		if len(sPSlc) > 1 && sPSlc[1] == FalseStr {
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

// sortStatIDs implements rankingSorter interface.
func (s *rankingAscSorter) sortStatIDs() []string {
	if len(s.statIDs) == 0 {
		return s.statIDs
	}
	sort.Slice(s.statIDs, func(i, j int) bool {
		for _, metricID := range s.sMetricIDs {
			val1, hasMetric1 := s.Metrics[s.statIDs[i]][metricID]
			val2, hasMetric2 := s.Metrics[s.statIDs[j]][metricID]
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
			if s.sMetricRev.Has(metricID) {
				ret = !ret // reversed logic in case of metric:false in params
			}
			return ret
		}
		//in case that we have the same value for all params we return randomly
		return BoolGenerator().RandomBool()
	})
	return s.statIDs
}
