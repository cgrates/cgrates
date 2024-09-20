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

func NewTrendService(dm *DataManager,
	cgrcfg *config.CGRConfig, filterS *FilterS, connMgr *ConnManager) (tS *TrendS) {
	return &TrendS{
		dm:          dm,
		cfg:         cgrcfg,
		fltrS:       filterS,
		connMgr:     connMgr,
		loopStopped: make(chan struct{}),
		crnTQs:      make(map[string]map[string]cron.EntryID),
		crnTQsMux:   new(sync.RWMutex),
	}
}

// TrendS is responsible of implementing the logic of TrendService
type TrendS struct {
	dm      *DataManager
	cfg     *config.CGRConfig
	fltrS   *FilterS
	connMgr *ConnManager

	crn *cron.Cron // cron refernce

	crnTQsMux *sync.RWMutex                      // protects the crnTQs
	crnTQs    map[string]map[string]cron.EntryID // save the EntryIDs for TrendQueries so we can reschedule them when needed

	loopStopped chan struct{}
}

// computeTrend will query a stat and build the Trend for it
//
// it is to be called by Cron service
func (tS *TrendS) computeTrend(tP *TrendProfile) {
	var floatMetrics map[string]float64
	if err := tS.connMgr.Call(context.Background(), tS.cfg.TrendSCfg().StatSConns,
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
		}
	} else if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying trend with id: <%s:%s> dm error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	trend.tMux.Lock()
	defer trend.tMux.Unlock()

	now := time.Now()
	var metrics []string
	if len(tP.Metrics) != 0 {
		metrics = tP.Metrics
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
	trend.Metrics[now] = make(map[string]*MetricWithTrend)
	for _, mID := range metrics {
		mWt := &MetricWithTrend{ID: mID}
		var has bool
		if mWt.Value, has = floatMetrics[mID]; !has { // no stats computed for metric
			mWt.Value = -1.0
			mWt.TrendLabel = utils.NotAvailable
			continue
		}
		if mWt.TrendGrowth, err = trend.getTrendGrowth(mID, mWt.Value, tP.CorrelationType, tS.cfg.GeneralCfg().RoundingDecimals); err != nil {
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

// scheduleTrendQueries will schedule/re-schedule specific trend queries
func (tS *TrendS) scheduleTrendQueries(ctx *context.Context, tnt string, tIds []string) (complete bool) {
	complete = true
	for _, tID := range tIds {
		tS.crnTQsMux.RLock()
		if entryID, has := tS.crnTQs[tnt][tID]; has {
			tS.crn.Remove(entryID)
		}
		tS.crnTQsMux.RUnlock()
		if tP, err := tS.dm.GetTrendProfile(ctx, tnt, tID); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving TrendProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			complete = false
		} else if entryID, err := tS.crn.AddFunc(tP.Schedule,
			func() { tS.computeTrend(tP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling TrendProfile <%s:%s>, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			complete = false
		} else {
			tS.crnTQsMux.Lock()
			tS.crnTQs[tP.Tenant][tP.ID] = entryID
			tS.crnTQsMux.Unlock()
		}
	}
	return
}
