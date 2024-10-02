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
	"fmt"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// NewTrendS is the constructor for TrendS
func NewTrendS(dm *DataManager,
	connMgr *ConnManager,
	filterS *FilterS,
	cgrcfg *config.CGRConfig) *TrendS {
	return &TrendS{
		dm:          dm,
		connMgr:     connMgr,
		filterS:     filterS,
		cgrcfg:      cgrcfg,
		crn:         cron.New(),
		loopStopped: make(chan struct{}),
		crnTQsMux:   new(sync.RWMutex),
		crnTQs:      make(map[string]map[string]cron.EntryID),
	}
}

// TrendS is responsible of implementing the logic of TrendService
type TrendS struct {
	dm      *DataManager
	connMgr *ConnManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig

	crn *cron.Cron // cron reference

	crnTQsMux *sync.RWMutex                      // protects the crnTQs
	crnTQs    map[string]map[string]cron.EntryID // save the EntryIDs for TrendQueries so we can reschedule them when needed

	loopStopped chan struct{}
}

// computeTrend will query a stat and build the Trend for it
//
//	it is to be called by Cron service
func (tS *TrendS) computeTrend(tP *TrendProfile) {
	var floatMetrics map[string]float64
	if err := tS.connMgr.Call(context.Background(), tS.cgrcfg.TrendSCfg().StatSConns,
		utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tP.Tenant, ID: tP.StatID}},
		&floatMetrics); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> computing trend for with id: <%s:%s> stats <%s> error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, tP.StatID, err.Error()))
		return
	}
	trend, err := tS.dm.GetTrend(tP.Tenant, tP.ID, true, true, utils.NonTransactional)
	if err == utils.ErrNotFound {
		trend = &Trend{
			Tenant:   tP.Tenant,
			ID:       tP.ID,
			RunTimes: make([]time.Time, 0),
			Metrics:  make(map[time.Time]map[string]*MetricWithTrend),
			tMux:     new(sync.RWMutex),
		}
	} else if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying trend with id: <%s:%s> dm error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	if trend.tMux == nil {
		trend.tMux = new(sync.RWMutex)
	}
	trend.tMux.Lock()
	defer trend.tMux.Unlock()
	trend.cleanup(tP.TTL, tP.QueueLength)

	if len(trend.mTotals) == 0 { // indexes were not yet built
		trend.computeIndexes()
	}
	now := time.Now()
	var metrics []string
	if len(tP.Metrics) != 0 {
		metrics = tP.Metrics // read only
	}
	if len(metrics) == 0 { // unlimited metrics in trend
		for mID := range floatMetrics {
			metrics = append(metrics, mID)
		}
	}
	if len(metrics) == 0 {
		return // nothing to compute
	}
	trend.RunTimes = append(trend.RunTimes, now)
	if trend.Metrics == nil {
		trend.Metrics = make(map[time.Time]map[string]*MetricWithTrend)
	}
	trend.Metrics[now] = make(map[string]*MetricWithTrend)
	for _, mID := range metrics {
		mWt := &MetricWithTrend{ID: mID}
		var has bool
		if mWt.Value, has = floatMetrics[mID]; !has { // no stats computed for metric
			mWt.Value = -1.0
			mWt.TrendLabel = utils.NotAvailable
			continue
		}
		if mWt.TrendGrowth, err = trend.getTrendGrowth(mID, mWt.Value, tP.CorrelationType, tS.cgrcfg.GeneralCfg().RoundingDecimals); err != nil {
			mWt.TrendLabel = utils.NotAvailable
		} else {
			mWt.TrendLabel = trend.getTrendLabel(mWt.TrendGrowth, tP.Tolerance)
		}
		trend.Metrics[now][mWt.ID] = mWt
	}
	if err := tS.dm.SetTrend(trend); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> setting trend with id: <%s:%s> dm error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}

}

func (tS *TrendS) StartScheduling() {
	tS.crn.Start()
}

func (tS *TrendS) StopScheduling() {
	ctx := tS.crn.Stop()
	<-ctx.Done()
}

// scheduleTrendQueries will schedule/re-schedule specific trend queries
func (tS *TrendS) scheduleTrendQueries(ctx *context.Context, tnt string, tIDs []string) (scheduled int, err error) {
	var partial bool
	for _, tID := range tIDs {
		tS.crnTQsMux.RLock()
		if entryID, has := tS.crnTQs[tnt][tID]; has {
			tS.crn.Remove(entryID) // deschedule the query
		}
		tS.crnTQsMux.RUnlock()
		if tP, err := tS.dm.GetTrendProfile(tnt, tID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving TrendProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else if entryID, err := tS.crn.AddFunc(tP.Schedule,
			func() { tS.computeTrend(tP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling TrendProfile <%s:%s>, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else { // log the entry ID for debugging
			tS.crnTQsMux.Lock()
			tS.crnTQs[tP.Tenant] = make(map[string]cron.EntryID)
			tS.crnTQs[tP.Tenant][tP.ID] = entryID
			tS.crnTQsMux.Unlock()
		}
		scheduled += 1
	}
	if partial {
		return 0, utils.ErrPartiallyExecuted
	}
	return
}

// V1ScheduleQueries is the query for manually re-/scheduling Trend Queries
func (tS *TrendS) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleTrendQueries, scheduled *int) (err error) {
	if sched, errSched := tS.scheduleTrendQueries(ctx, args.Tenant, args.TrendIDs); errSched != nil {
		return errSched
	} else {
		*scheduled = sched
	}
	return
}

func (tS *TrendS) V1GetTrend(ctx *context.Context, arg *utils.ArgGetTrend, trend *Trend) (err error) {
	var tr *Trend
	tr, err = tS.dm.GetTrend(arg.Tenant, arg.ID, true, true, utils.NonTransactional)
	*trend = *tr
	return
}
