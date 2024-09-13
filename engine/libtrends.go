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
	"sync"
	"time"

	"github.com/cgrates/cgrates/utils"
)

type TrendProfile struct {
	Tenant       string
	ID           string
	Schedule     string // Cron expression scheduling gathering of the metrics
	StatID       string
	Metrics      []*MetricWithSettings
	QueueLength  int
	TTL          time.Duration
	TrendType    string // *last, *average
	ThresholdIDs []string
}

// Clone will clone the TrendProfile so it can be used by scheduler safely
func (tP *TrendProfile) Clone() (clnTp *TrendProfile) {
	clnTp = &TrendProfile{
		Tenant:      tP.Tenant,
		ID:          tP.ID,
		Schedule:    tP.Schedule,
		StatID:      tP.StatID,
		QueueLength: tP.QueueLength,
		TTL:         tP.TTL,
		TrendType:   tP.TrendType,
	}
	if tP.Metrics != nil {
		clnTp.Metrics = make([]*MetricWithSettings, len(tP.Metrics))
		for i, m := range tP.Metrics {
			clnTp.Metrics[i] = &MetricWithSettings{MetricID: m.MetricID,
				TrendSwingMargin: m.TrendSwingMargin}
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

// MetricWithSettings adds specific settings to the Metric
type MetricWithSettings struct {
	MetricID         string
	TrendSwingMargin float64 // allow this margin for *neutral trend
}

type TrendProfileWithAPIOpts struct {
	*TrendProfile
	APIOpts map[string]any
}

type TrendProfilesAPI struct {
	Tenant string
	TpIDs  []string
}

func (srp *TrendProfile) TenantID() string {
	return utils.ConcatenatedKey(srp.Tenant, srp.ID)
}

type TrendWithAPIOpts struct {
	*Trend
	APIOpts map[string]any
}

// Trend is the unit matched by filters
type Trend struct {
	sync.RWMutex

	Tenant   string
	ID       string
	RunTimes []time.Time
	Metrics  map[time.Time]map[string]*MetricWithTrend

	// indexes help faster processing
	mLast   map[string]time.Time // last time a metric was present
	mCounts map[string]int       // number of times a metric is present in Metrics
	mTotals map[string]float64   // cached sum, used for average calculations
}

// computeIndexes should be called after each retrieval from DB
func (t *Trend) computeIndexes() {
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

// getTrendLabel identifies the trend label for the instant value of the metric
//
//	*positive, *negative, *constant, N/A
func (t *Trend) getTrendLabel(mID string, mVal float64, swingMargin float64) (lbl string) {
	var prevVal *float64
	if _, has := t.mLast[mID]; has {
		prevVal = &t.Metrics[t.mLast[mID]][mID].Value
	}
	if prevVal == nil {
		return utils.NotAvailable
	}
	diffVal := mVal - *prevVal
	switch {
	case diffVal > 0:
		lbl = utils.MetaPositive
	case diffVal < 0:
		lbl = utils.MetaNegative
	default:
		lbl = utils.MetaConstant
	}
	if math.Abs(diffVal*100/(*prevVal)) <= swingMargin { // percentage value of diff is lower than threshold
		lbl = utils.MetaConstant
	}
	return
}

// MetricWithTrend represents one read from StatS
type MetricWithTrend struct {
	ID    string  // Metric ID
	Value float64 // Metric Value
	Trend string  // *positive, *negative, *constant, N/A
}

func (tr *Trend) TenantID() string {
	return utils.ConcatenatedKey(tr.Tenant, tr.ID)
}
