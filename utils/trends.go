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
	"math"
	"slices"
	"sync"
	"time"
)

// TrendProfile defines the configuration of the Trend.
type TrendProfile struct {
	Tenant          string
	ID              string
	Schedule        string // Cron expression scheduling gathering of the metrics
	StatID          string
	Metrics         []string
	TTL             time.Duration
	QueueLength     int
	MinItems        int     // minimum number of items for building Trends
	CorrelationType string  // *last, *average
	Tolerance       float64 // allow this deviation margin for *constant trend
	Stored          bool    // store the Trend in dataDB
	ThresholdIDs    []string
}

// TrendProfileWithAPIOpts wraps TrendProfile with APIOpts.
type TrendProfileWithAPIOpts struct {
	*TrendProfile
	APIOpts map[string]any
}

// Clone creates a deep copy of TrendProfile for thread-safe use.
func (tp *TrendProfile) Clone() (clnTp *TrendProfile) {
	clnTp = &TrendProfile{
		Tenant:          tp.Tenant,
		ID:              tp.ID,
		Schedule:        tp.Schedule,
		StatID:          tp.StatID,
		QueueLength:     tp.QueueLength,
		TTL:             tp.TTL,
		MinItems:        tp.MinItems,
		CorrelationType: tp.CorrelationType,
		Tolerance:       tp.Tolerance,
		Stored:          tp.Stored,
	}
	if tp.Metrics != nil {
		clnTp.Metrics = make([]string, len(tp.Metrics))
		for i, mID := range tp.Metrics {
			clnTp.Metrics[i] = mID
		}
	}
	if tp.ThresholdIDs != nil {
		clnTp.ThresholdIDs = make([]string, len(tp.ThresholdIDs))
		for i, tID := range tp.ThresholdIDs {
			clnTp.ThresholdIDs[i] = tID
		}
	}
	return
}

// TenantID returns the concatenated tenant and ID.
func (tp *TrendProfile) TenantID() string {
	return ConcatenatedKey(tp.Tenant, tp.ID)
}

// Set implements the profile interface, setting values in TrendProfile based on path.
func (tp *TrendProfile) Set(path []string, val any, _ bool) (err error) {
	if len(path) != 1 {
		return ErrWrongPath
	}

	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		tp.Tenant = IfaceAsString(val)
	case ID:
		tp.ID = IfaceAsString(val)
	case Schedule:
		tp.Schedule = IfaceAsString(val)
	case StatID:
		tp.StatID = IfaceAsString(val)
	case Metrics:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.Metrics = append(tp.Metrics, valA...)
	case TTL:
		tp.TTL, err = IfaceAsDuration(val)
	case QueueLength:
		tp.QueueLength, err = IfaceAsInt(val)
	case MinItems:
		tp.MinItems, err = IfaceAsInt(val)
	case CorrelationType:
		tp.CorrelationType = IfaceAsString(val)
	case Tolerance:
		tp.Tolerance, err = IfaceAsFloat64(val)
	case Stored:
		tp.Stored, err = IfaceAsBool(val)
	case ThresholdIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		tp.ThresholdIDs = append(tp.ThresholdIDs, valA...)
	}
	return
}

// Merge implements the profile interface, merging values from another TrendProfile.
func (tp *TrendProfile) Merge(v2 any) {
	vi := v2.(*TrendProfile)
	if len(vi.Tenant) != 0 {
		tp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		tp.ID = vi.ID
	}
	if len(vi.Schedule) != 0 {
		tp.Schedule = vi.Schedule
	}
	if len(vi.StatID) != 0 {
		tp.StatID = vi.StatID
	}
	tp.Metrics = append(tp.Metrics, vi.Metrics...)
	tp.ThresholdIDs = append(tp.ThresholdIDs, vi.ThresholdIDs...)
	if vi.Stored {
		tp.Stored = vi.Stored
	}
	if vi.TTL != 0 {
		tp.TTL = vi.TTL
	}
	if vi.QueueLength != 0 {
		tp.QueueLength = vi.QueueLength
	}
	if vi.MinItems != 0 {
		tp.MinItems = vi.MinItems
	}
	if len(vi.CorrelationType) != 0 {
		tp.CorrelationType = vi.CorrelationType
	}
	if vi.Tolerance != 0 {
		tp.Tolerance = vi.Tolerance
	}
}

// String implements the DataProvider interface, returning the TrendProfile in JSON format.
func (tp *TrendProfile) String() string { return ToJSON(tp) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (tp *TrendProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = tp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (tp *TrendProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case Metrics:
				if *idx < len(tp.Metrics) {
					return tp.Metrics[*idx], nil
				}
			case ThresholdIDs:
				if *idx < len(tp.ThresholdIDs) {
					return tp.ThresholdIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return tp.Tenant, nil
	case ID:
		return tp.ID, nil
	case Schedule:
		return tp.Schedule, nil
	case StatID:
		return tp.StatID, nil
	case TTL:
		return tp.TTL, nil
	case QueueLength:
		return tp.QueueLength, nil
	case MinItems:
		return tp.MinItems, nil
	case CorrelationType:
		return tp.CorrelationType, nil
	case Tolerance:
		return tp.Tolerance, nil
	case Stored:
		return tp.Stored, nil
	}
}

// Trend represents a collection of metrics with trend analysis.
type Trend struct {
	tMux sync.RWMutex

	Tenant            string
	ID                string
	RunTimes          []time.Time
	Metrics           map[time.Time]map[string]*MetricWithTrend
	CompressedMetrics []byte // if populated, Metrics and RunTimes will be emty

	// indexes help faster processing
	mLast   map[string]time.Time // last time a metric was present
	mCounts map[string]int       // number of times a metric is present in Metrics
	mTotals map[string]float64   // cached sum, used for average calculations

	tPrfl *TrendProfile // store here the trend profile so we can have it at hands further
}

// TrendWithAPIOpts wraps Trend with APIOpts.
type TrendWithAPIOpts struct {
	*Trend
	APIOpts map[string]any
}

// TrendSummary holds the most recent trend metrics.
type TrendSummary struct {
	Tenant  string
	ID      string
	Time    time.Time
	Metrics map[string]*MetricWithTrend
}

// MetricWithTrend contains a metric value with its calculated trend info.
type MetricWithTrend struct {
	ID          string  // Metric ID
	Value       float64 // Metric Value
	TrendGrowth float64 // Difference between last and previous
	TrendLabel  string  // *positive, *negative, *constant, N/A
}

// NewTrendFromProfile creates an empty trend based on profile configuration.
func NewTrendFromProfile(tP *TrendProfile) *Trend {
	return &Trend{
		Tenant:   tP.Tenant,
		ID:       tP.ID,
		RunTimes: make([]time.Time, 0),
		Metrics:  make(map[time.Time]map[string]*MetricWithTrend),

		tPrfl: tP,
	}
}

func (t *Trend) TenantID() string {
	return ConcatenatedKey(t.Tenant, t.ID)
}

// Config returns the trend's profile configuration.
func (t *Trend) Config() *TrendProfile {
	return t.tPrfl
}

// SetConfig sets the trend's profile configuration.
func (t *Trend) SetConfig(tp *TrendProfile) {
	t.tPrfl = tp
}

// AsTrendSummary creates a summary with the most recent trend data.
func (t *Trend) AsTrendSummary() (ts *TrendSummary) {
	ts = &TrendSummary{
		Tenant:  t.Tenant,
		ID:      t.ID,
		Metrics: make(map[string]*MetricWithTrend),
	}
	if len(t.RunTimes) != 0 {
		ts.Time = t.RunTimes[len(t.RunTimes)-1]
		for mID, mWt := range t.Metrics[ts.Time] {
			ts.Metrics[mID] = &MetricWithTrend{
				ID:          mWt.ID,
				Value:       mWt.Value,
				TrendGrowth: mWt.TrendGrowth,
				TrendLabel:  mWt.TrendLabel,
			}
		}
	}
	return
}

// Compress creates a compressed version of the trend.
func (t *Trend) Compress(ms Marshaler, limit int) (tr *Trend, err error) {
	if limit > len(t.RunTimes) {
		return
	}
	tr = &Trend{
		Tenant: t.Tenant,
		ID:     t.ID,
	}
	tr.CompressedMetrics, err = ms.Marshal(tr.Metrics)
	if err != nil {
		return
	}
	return tr, nil
}

// Uncompress expands a compressed trend.
func (t *Trend) Uncompress(ms Marshaler) (err error) {
	if t == nil || t.CompressedMetrics == nil {
		return
	}

	err = ms.Unmarshal(t.CompressedMetrics, &t.Metrics)
	if err != nil {
		return
	}
	t.CompressedMetrics = nil
	t.RunTimes = make([]time.Time, len(t.Metrics))
	i := 0
	for key := range t.Metrics {
		t.RunTimes[i] = key
		i++
	}
	slices.SortFunc(t.RunTimes, func(a, b time.Time) int {
		return a.Compare(b)
	})
	return
}

// Compile initializes or cleans up the Trend.
// Safe for concurrent use.
func (t *Trend) Compile(cleanTtl time.Duration, qLength int) {
	t.cleanup(cleanTtl, qLength)
	if len(t.mTotals) == 0 { // indexes were not yet built
		t.computeIndexes()
	}
}

// cleanup removes stale data based on TTL and queue length limits.
func (t *Trend) cleanup(ttl time.Duration, qLength int) (altered bool) {
	if ttl >= 0 {
		expTime := time.Now().Add(-ttl)
		var expIdx *int
		for i, rT := range t.RunTimes {
			if rT.After(expTime) {
				continue
			}
			expIdx = &i
			delete(t.Metrics, rT)
		}
		if expIdx != nil {
			if len(t.RunTimes)-1 == *expIdx {
				t.RunTimes = make([]time.Time, 0)
			} else {
				t.RunTimes = t.RunTimes[*expIdx+1:]
			}
			altered = true
		}
	}

	diffLen := len(t.RunTimes) - qLength
	if qLength > 0 && diffLen > 0 {
		var rmTms []time.Time
		rmTms, t.RunTimes = t.RunTimes[:diffLen], t.RunTimes[diffLen:]
		for _, rmTm := range rmTms {
			delete(t.Metrics, rmTm)
		}
		altered = true
	}
	if altered {
		t.computeIndexes()
	}
	return
}

// computeIndexes rebuilds internal indexes after DB retrieval.
func (t *Trend) computeIndexes() {
	t.mLast = make(map[string]time.Time)
	t.mCounts = make(map[string]int)
	t.mTotals = make(map[string]float64)
	for _, runTime := range t.RunTimes {
		for _, mWt := range t.Metrics[runTime] {
			t.IndexesAppendMetric(mWt, runTime)
		}
	}
}

// IndexesAppendMetric adds a single metric to internal indexes.
func (t *Trend) IndexesAppendMetric(mWt *MetricWithTrend, rTime time.Time) {
	t.mLast[mWt.ID] = rTime
	t.mCounts[mWt.ID]++
	t.mTotals[mWt.ID] += mWt.Value
}

// GetTrendGrowth calculates percentage growth for a metric compared to previous values.
func (t *Trend) GetTrendGrowth(mID string, mVal float64, correlation string, roundDec int) (tG float64, err error) {
	var prevVal float64
	if _, has := t.mLast[mID]; !has {
		return -1.0, ErrNotFound
	}
	if _, has := t.Metrics[t.mLast[mID]][mID]; !has {
		return -1.0, ErrNotFound
	}
	switch correlation {
	case MetaLast:
		prevVal = t.Metrics[t.mLast[mID]][mID].Value
	case MetaAverage:
		prevVal = t.mTotals[mID] / float64(t.mCounts[mID])
	default:
		return -1.0, ErrCorrelationUndefined
	}

	diffVal := mVal - prevVal
	return Round(diffVal*100/prevVal, roundDec, MetaRoundingMiddle), nil
}

// Lock locks the trend mutex.
func (t *Trend) Lock() {
	t.tMux.Lock()
}

// Unlock unlocks the trend mutex.
func (t *Trend) Unlock() {
	t.tMux.Unlock()
}

// RLock locks the trend mutex for reading.
func (t *Trend) RLock() {
	t.tMux.RLock()
}

// RUnlock unlocks the read lock on the trend mutex.
func (t *Trend) RUnlock() {
	t.tMux.RUnlock()
}

// GetTrendLabel determines trend direction based on growth percentage and tolerance.
// Returns "*positive", "*negative", "*constant", or "N/A" based on the growth value.
func GetTrendLabel(tGrowth float64, tolerance float64) (lbl string) {
	switch {
	case tGrowth > 0:
		lbl = MetaPositive
	case tGrowth < 0:
		lbl = MetaNegative
	default:
		lbl = MetaConstant
	}
	if math.Abs(tGrowth) <= tolerance { // percentage value of diff is lower than threshold
		lbl = MetaConstant
	}
	return
}
