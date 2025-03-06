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

package trends

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cron"
)

// TrendS implements the logic for processing and managing trends based on stat metrics.
type TrendS struct {
	dm      *engine.DataManager
	cfg     *config.CGRConfig
	fltrS   *engine.FilterS
	connMgr *engine.ConnManager

	crn *cron.Cron // cron refernce

	crnTQsMux *sync.RWMutex                      // protects the crnTQs
	crnTQs    map[string]map[string]cron.EntryID // save the EntryIDs for TrendQueries so we can reschedule them when needed

	storedTrends   utils.StringSet // keep a record of trends which need saving, map[trendTenantID]bool
	sTrndsMux      sync.RWMutex    // protects storedTrends
	storingStopped chan struct{}   // signal back that the operations were stopped

	loopStopped chan struct{}
	trendStop   chan struct{} // signal to stop all operations
}

// NewTrendService creates a new TrendS service.
func NewTrendService(dm *engine.DataManager,
	cgrcfg *config.CGRConfig, filterS *engine.FilterS, connMgr *engine.ConnManager) (tS *TrendS) {
	return &TrendS{
		dm:             dm,
		cfg:            cgrcfg,
		fltrS:          filterS,
		connMgr:        connMgr,
		loopStopped:    make(chan struct{}),
		crn:            cron.New(),
		crnTQs:         make(map[string]map[string]cron.EntryID),
		crnTQsMux:      new(sync.RWMutex),
		storedTrends:   make(utils.StringSet),
		storingStopped: make(chan struct{}),
		trendStop:      make(chan struct{}),
	}
}

// computeTrend queries a stat and builds the Trend for it based on the TrendProfile configuration.
// Called by Cron service at scheduled intervals.
func (tS *TrendS) computeTrend(ctx *context.Context, tP *utils.TrendProfile) {
	var floatMetrics map[string]float64
	if err := tS.connMgr.Call(context.Background(), tS.cfg.TrendSCfg().StatSConns,
		utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tP.Tenant, ID: tP.StatID}},
		&floatMetrics); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> computing trend with id: <%s:%s> for stats <%s> error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, tP.StatID, err.Error()))
		return
	}
	trnd, err := tS.dm.GetTrend(ctx, tP.Tenant, tP.ID, true, true, utils.NonTransactional)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying trend with id: <%s:%s> dm error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	trnd.Lock()
	defer trnd.Unlock()
	if trnd.Config() == nil {
		trnd.SetConfig(tP)
	}
	trnd.Compile(tP.TTL, tP.QueueLength)
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
	trnd.RunTimes = append(trnd.RunTimes, now)
	if trnd.Metrics == nil {
		trnd.Metrics = make(map[time.Time]map[string]*utils.MetricWithTrend)
	}
	trnd.Metrics[now] = make(map[string]*utils.MetricWithTrend)
	for _, mID := range metrics {
		mWt := &utils.MetricWithTrend{ID: mID}
		var has bool
		if mWt.Value, has = floatMetrics[mID]; !has { // no stats computed for metric
			mWt.Value = -1.0
			mWt.TrendLabel = utils.NotAvailable
			continue
		}
		if mWt.TrendGrowth, err = trnd.GetTrendGrowth(mID, mWt.Value, tP.CorrelationType,
			tS.cfg.GeneralCfg().RoundingDecimals); err != nil {
			mWt.TrendLabel = utils.NotAvailable
		} else {
			mWt.TrendLabel = utils.GetTrendLabel(mWt.TrendGrowth, tP.Tolerance)
		}
		trnd.Metrics[now][mWt.ID] = mWt
		trnd.IndexesAppendMetric(mWt, now)
	}
	if err = tS.storeTrend(ctx, trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> setting Trend with id: <%s:%s> DM error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	if err = tS.processThresholds(trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with ThresholdS",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
	}
	if err = tS.processEEs(trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with EEs",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
	}

}

// processThresholds sends the computed trend to ThresholdS.
func (tS *TrendS) processThresholds(trnd *utils.Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.Config().MinItems {
		return
	}
	if len(tS.cfg.TrendSCfg().ThresholdSConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.TrendUpdate,
	}
	var thIDs []string
	if len(trnd.Config().ThresholdIDs) != 0 {
		if len(trnd.Config().ThresholdIDs) == 1 &&
			trnd.Config().ThresholdIDs[0] == utils.MetaNone {
			return
		}
		thIDs = make([]string, len(trnd.Config().ThresholdIDs))
		copy(thIDs, trnd.Config().ThresholdIDs)
	}
	opts[utils.OptsThresholdsProfileIDs] = thIDs
	ts := trnd.AsTrendSummary()
	trndEv := &utils.CGREvent{
		Tenant:  trnd.Tenant,
		ID:      utils.GenUUID(),
		APIOpts: opts,
		Event: map[string]any{
			utils.TrendID: trnd.ID,
			utils.Time:    ts.Time,
			utils.Metrics: ts.Metrics,
		},
	}
	var withErrs bool
	var tIDs []string
	if err := tS.connMgr.Call(context.TODO(), tS.cfg.TrendSCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent, trndEv, &tIDs); err != nil &&
		(len(thIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.", utils.TrendS, err.Error(), trndEv))
		withErrs = true
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEEs sends the computed trend to EEs.
func (tS *TrendS) processEEs(trnd *utils.Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.Config().MinItems {
		return
	}
	if len(tS.cfg.TrendSCfg().EEsConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.TrendUpdate,
	}
	ts := trnd.AsTrendSummary()
	trndEv := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant:  trnd.Tenant,
			ID:      utils.GenUUID(),
			APIOpts: opts,
			Event: map[string]any{
				utils.TrendID: trnd.ID,
				utils.Time:    ts.Time,
				utils.Metrics: ts.Metrics,
			},
		},
		EeIDs: tS.cfg.TrendSCfg().EEsExporterIDs,
	}
	var withErrs bool
	var reply map[string]map[string]any
	if err := tS.connMgr.Call(context.TODO(), tS.cfg.TrendSCfg().EEsConns,
		utils.EeSv1ProcessEvent, trndEv, &reply); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %q processing event %+v with EEs.", utils.TrendS, err.Error(), trndEv))
		withErrs = true
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// storeTrend stores or schedules the trend for storage based on "store_interval".
func (tS *TrendS) storeTrend(ctx *context.Context, trnd *utils.Trend) (err error) {
	if tS.cfg.TrendSCfg().StoreInterval == 0 {
		return
	}
	if tS.cfg.TrendSCfg().StoreInterval == -1 {
		return tS.dm.SetTrend(ctx, trnd)
	}

	// schedule the asynchronous save, relies for Trend to be in cache
	tS.sTrndsMux.Lock()
	tS.storedTrends.Add(trnd.TenantID())
	tS.sTrndsMux.Unlock()
	return
}

// storeTrends stores modified trends from cache in dataDB
// Reschedules failed trend IDs for next storage cycle.
// This function is safe for concurrent use.
func (tS *TrendS) storeTrends(ctx *context.Context) {
	var failedTrndIDs []string
	for {
		tS.sTrndsMux.Lock()
		trndID := tS.storedTrends.GetOne()
		if trndID != utils.EmptyString {
			tS.storedTrends.Remove(trndID)
		}
		tS.sTrndsMux.Unlock()
		if trndID == utils.EmptyString {
			break // no more keys, backup completed
		}
		trndIf, ok := engine.Cache.Get(utils.CacheTrends, trndID)
		if !ok || trndIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache Trend with ID: %q",
					utils.TrendS, trndID))
			failedTrndIDs = append(failedTrndIDs, trndID) // record failure so we can schedule it for next backup
			continue
		}
		trnd := trndIf.(*utils.Trend)
		trnd.Lock()
		if err := tS.dm.SetTrend(ctx, trnd); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed storing Trend with ID: %q, err: %q",
					utils.TrendS, trndID, err))
			failedTrndIDs = append(failedTrndIDs, trndID) // record failure so we can schedule it for next backup
		}
		trnd.Unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedTrndIDs) != 0 { // there were errors on save, schedule the keys for next backup
		tS.sTrndsMux.Lock()
		tS.storedTrends.AddSlice(failedTrndIDs)
		tS.sTrndsMux.Unlock()
	}
}

// asyncStoreTrends runs as a background process that periodically calls storeTrends.
func (tS *TrendS) asyncStoreTrends(ctx *context.Context) {
	storeInterval := tS.cfg.TrendSCfg().StoreInterval
	if storeInterval <= 0 {
		close(tS.storingStopped)
		return
	}
	for {
		tS.storeTrends(ctx)
		select {
		case <-tS.trendStop:
			close(tS.storingStopped)
			return
		case <-time.After(storeInterval): // continue to another storing loop
		}
	}
}

// StartTrendS activates the Cron service with scheduled trend queries.
func (tS *TrendS) StartTrendS(ctx *context.Context) error {
	if err := tS.scheduleAutomaticQueries(ctx); err != nil {
		return err
	}
	tS.crn.Start()
	go tS.asyncStoreTrends(ctx)
	return nil
}

// StopTrendS gracefully shuts down Cron tasks and trend operations.
func (tS *TrendS) StopTrendS() {
	timeEnd := time.Now().Add(tS.cfg.CoreSCfg().ShutdownTimeout)

	crnctx := tS.crn.Stop()
	close(tS.trendStop)

	// Wait for cron
	select {
	case <-crnctx.Done():
	case <-time.After(time.Until(timeEnd)):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for Cron to finish",
				utils.TrendS))
		return
	}
	// Wait for backup and other operations
	select {
	case <-tS.storingStopped:
	case <-time.After(time.Until(timeEnd)):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for TrendS to finish",
				utils.TrendS))
		return
	}
}

// Reload restarts trend services with updated configuration.
func (tS *TrendS) Reload(ctx *context.Context) {
	crnctx := tS.crn.Stop()
	close(tS.trendStop)
	<-crnctx.Done()
	<-tS.storingStopped
	tS.trendStop = make(chan struct{})
	tS.storingStopped = make(chan struct{})
	tS.crn.Start()
	go tS.asyncStoreTrends(ctx)
}

// scheduleAutomaticQueries schedules initial trend queries based on configuration.
func (tS *TrendS) scheduleAutomaticQueries(ctx *context.Context) error {
	schedData := make(map[string][]string)
	for k, v := range tS.cfg.TrendSCfg().ScheduledIDs {
		schedData[k] = v
	}
	var tnts []string
	if len(schedData) == 0 {
		tnts = make([]string, 0)
	}
	for tnt, tIDs := range schedData {
		if len(tIDs) == 0 {
			tnts = append(tnts, tnt)
		}
	}
	if tnts != nil {
		qrydData, err := tS.dm.GetTrendProfileIDs(ctx, tnts)
		if err != nil {
			return err
		}
		for tnt, ids := range qrydData {
			schedData[tnt] = ids
		}
	}
	for tnt, tIDs := range schedData {
		if _, err := tS.scheduleTrendQueries(ctx, tnt, tIDs); err != nil {
			return err
		}
	}
	return nil
}

// scheduleTrendQueries schedules or reschedules specific trend queries
// Safe for concurrent use.
func (tS *TrendS) scheduleTrendQueries(ctx *context.Context, tnt string, tIDs []string) (scheduled int, err error) {
	var partial bool
	tS.crnTQsMux.Lock()
	if _, has := tS.crnTQs[tnt]; !has {
		tS.crnTQs[tnt] = make(map[string]cron.EntryID)
	}
	tS.crnTQsMux.Unlock()
	for _, tID := range tIDs {
		tS.crnTQsMux.RLock()
		if entryID, has := tS.crnTQs[tnt][tID]; has {
			tS.crn.Remove(entryID) // deschedule the query
		}
		tS.crnTQsMux.RUnlock()
		if tP, err := tS.dm.GetTrendProfile(ctx, tnt, tID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving TrendProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else if entryID, err := tS.crn.AddFunc(tP.Schedule,
			func() { tS.computeTrend(ctx, tP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling TrendProfile <%s:%s>, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else { // log the entry ID for debugging
			tS.crnTQsMux.Lock()
			tS.crnTQs[tP.Tenant][tP.ID] = entryID
			tS.crnTQsMux.Unlock()
			scheduled++
		}
	}
	if partial {
		return 0, utils.ErrPartiallyExecuted
	}
	return
}
