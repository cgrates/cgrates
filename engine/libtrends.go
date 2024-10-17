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
	"math"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// A TrendProfile represents the settings of a Trend
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

// Clone will clone the TrendProfile so it can be used by scheduler safely
func (tP *TrendProfile) Clone() (clnTp *TrendProfile) {
	clnTp = &TrendProfile{
		Tenant:          tP.Tenant,
		ID:              tP.ID,
		Schedule:        tP.Schedule,
		StatID:          tP.StatID,
		QueueLength:     tP.QueueLength,
		TTL:             tP.TTL,
		MinItems:        tP.MinItems,
		CorrelationType: tP.CorrelationType,
		Tolerance:       tP.Tolerance,
		Stored:          tP.Stored,
	}
	if tP.Metrics != nil {
		clnTp.Metrics = make([]string, len(tP.Metrics))
		for i, mID := range tP.Metrics {
			clnTp.Metrics[i] = mID
		}
	}
	if tP.ThresholdIDs != nil {
		clnTp.ThresholdIDs = make([]string, len(tP.ThresholdIDs))
		for i, tID := range tP.ThresholdIDs {
			clnTp.ThresholdIDs[i] = tID
		}
	}
	return
}

type TrendProfileWithAPIOpts struct {
	*TrendProfile
	APIOpts map[string]any
}

func (srp *TrendProfile) TenantID() string {
	return utils.ConcatenatedKey(srp.Tenant, srp.ID)
}

type TrendWithAPIOpts struct {
	*Trend
	APIOpts map[string]any
}

// NewTrendFromProfile is a constructor for an empty trend out of it's profile
func NewTrendFromProfile(tP *TrendProfile) *Trend {
	return &Trend{
		Tenant:   tP.Tenant,
		ID:       tP.ID,
		RunTimes: make([]time.Time, 0),
		Metrics:  make(map[time.Time]map[string]*MetricWithTrend),

		tPrfl: tP,
	}
}

// Trend is the unit matched by filters
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

func (t *Trend) Clone() (tC *Trend) {
	return
}

// AsTrendSummary transforms the trend into TrendSummary
func (t *Trend) asTrendSummary() (ts *TrendSummary) {
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

func (t *Trend) compress(ms Marshaler) (err error) {
	if config.CgrConfig().TrendSCfg().StoreUncompressedLimit > len(t.RunTimes) {
		return
	}
	t.CompressedMetrics, err = ms.Marshal(t.Metrics)
	if err != nil {
		return
	}
	t.Metrics = nil
	t.RunTimes = nil
	return nil
}

func (t *Trend) uncompress(ms Marshaler) (err error) {
	if t == nil || t.CompressedMetrics == nil {
		return
	}
	err = ms.Unmarshal(t.CompressedMetrics, &t.Metrics)
	if err != nil {
		return
	}
	t.CompressedMetrics = []byte{}
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

// Compile is used to initialize or cleanup the Trend
//
//	thread safe since it should be used close to source
func (t *Trend) Compile(cleanTtl time.Duration, qLength int) {
	t.cleanup(cleanTtl, qLength)
	if len(t.mTotals) == 0 { // indexes were not yet built
		t.computeIndexes()
	}
}

// cleanup will clean stale data out of
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

// computeIndexes should be called after each retrieval from DB
func (t *Trend) computeIndexes() {
	t.mLast = make(map[string]time.Time)
	t.mCounts = make(map[string]int)
	t.mTotals = make(map[string]float64)
	for _, runTime := range t.RunTimes {
		for _, mWt := range t.Metrics[runTime] {
			t.indexesAppendMetric(mWt, runTime)

		}
	}
}

// indexesAppendMetric appends a single metric to indexes
func (t *Trend) indexesAppendMetric(mWt *MetricWithTrend, rTime time.Time) {
	t.mLast[mWt.ID] = rTime
	t.mCounts[mWt.ID] += 1
	t.mTotals[mWt.ID] += mWt.Value
}

// getTrendGrowth returns the percentage growth for a specific metric
//
// @correlation parameter will define whether the comparison is against last or average value
// errors in case of previous
func (t *Trend) getTrendGrowth(mID string, mVal float64, correlation string, roundDec int) (tG float64, err error) {
	var prevVal float64
	if _, has := t.mLast[mID]; !has {
		return -1.0, utils.ErrNotFound
	}
	if _, has := t.Metrics[t.mLast[mID]][mID]; !has {
		return -1.0, utils.ErrNotFound
	}

	switch correlation {
	case utils.MetaLast:
		prevVal = t.Metrics[t.mLast[mID]][mID].Value
	case utils.MetaAverage:
		prevVal = t.mTotals[mID] / float64(t.mCounts[mID])
	default:
		return -1.0, utils.ErrCorrelationUndefined
	}

	diffVal := mVal - prevVal
	return utils.Round(diffVal*100/prevVal, roundDec, utils.MetaRoundingMiddle), nil
}

// getTrendLabel identifies the trend label for the instant value of the metric
//
//	*positive, *negative, *constant, N/A
func (t *Trend) getTrendLabel(tGrowth float64, tolerance float64) (lbl string) {
	switch {
	case tGrowth > 0:
		lbl = utils.MetaPositive
	case tGrowth < 0:
		lbl = utils.MetaNegative
	default:
		lbl = utils.MetaConstant
	}
	if math.Abs(tGrowth) <= tolerance { // percentage value of diff is lower than threshold
		lbl = utils.MetaConstant
	}
	return
}

// MetricWithTrend represents one read from StatS
type MetricWithTrend struct {
	ID          string  // Metric ID
	Value       float64 // Metric Value
	TrendGrowth float64 // Difference between last and previous
	TrendLabel  string  // *positive, *negative, *constant, N/A
}

func (tr *Trend) TenantID() string {
	return utils.ConcatenatedKey(tr.Tenant, tr.ID)
}

// TrendSummary represents the last trend computed
type TrendSummary struct {
	Tenant  string
	ID      string
	Time    time.Time
	Metrics map[string]*MetricWithTrend
}
