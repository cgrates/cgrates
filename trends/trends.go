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

	crn            *cron.Cron                         // cron refernce
	crnTQsMux      *sync.RWMutex                      // protects the crnTQs
	crnTQs         map[string]map[string]cron.EntryID // save the EntryIDs for TrendQueries so we can reschedule them when needed
	sTrndsMux      sync.RWMutex                       // protects storedTrends
	storedTrends   utils.StringSet                    // keep a record of trends which need saving, map[trendTenantID]bool
	storingStopped chan struct{}                      // signal back that the operations were stopped
	trendStop      chan struct{}                      // signal to stop all operations
	loopStopped    chan struct{}
}

// NewTrendService creates a new TrendS service.
func NewTrendService(dm *engine.DataManager,
	cgrcfg *config.CGRConfig, filterS *engine.FilterS, connMgr *engine.ConnManager) *TrendS {
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
func (t *TrendS) computeTrend(ctx *context.Context, tP *utils.TrendProfile) {
	var floatMetrics map[string]float64
	if err := t.connMgr.Call(context.Background(), t.cfg.TrendSCfg().StatSConns,
		utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tP.Tenant, ID: tP.StatID}},
		&floatMetrics); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> computing trend with id: <%s:%s> for stats <%s> error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, tP.StatID, err.Error()))
		return
	}
	trnd, err := t.dm.GetTrend(ctx, tP.Tenant, tP.ID, true, true, utils.NonTransactional)
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
			t.cfg.GeneralCfg().RoundingDecimals); err != nil {
			mWt.TrendLabel = utils.NotAvailable
		} else {
			mWt.TrendLabel = utils.GetTrendLabel(mWt.TrendGrowth, tP.Tolerance)
		}
		trnd.Metrics[now][mWt.ID] = mWt
		trnd.IndexesAppendMetric(mWt, now)
	}
	if err = t.storeTrend(ctx, trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> setting Trend with id: <%s:%s> DM error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	if err = t.processThresholds(trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with ThresholdS",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
	}
	if err = t.processEEs(trnd); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with EEs",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
	}

}

// processThresholds sends the computed trend to ThresholdS.
func (t *TrendS) processThresholds(trnd *utils.Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.Config().MinItems {
		return
	}
	if len(t.cfg.TrendSCfg().ThresholdSConns) == 0 {
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
	if err := t.connMgr.Call(context.TODO(), t.cfg.TrendSCfg().ThresholdSConns,
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
func (t *TrendS) processEEs(trnd *utils.Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.Config().MinItems {
		return
	}
	if len(t.cfg.TrendSCfg().EEsConns) == 0 {
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
		EeIDs: t.cfg.TrendSCfg().EEsExporterIDs,
	}
	var withErrs bool
	var reply map[string]map[string]any
	if err := t.connMgr.Call(context.TODO(), t.cfg.TrendSCfg().EEsConns,
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
func (t *TrendS) storeTrend(ctx *context.Context, trnd *utils.Trend) (err error) {
	if t.cfg.TrendSCfg().StoreInterval == 0 {
		return
	}
	if t.cfg.TrendSCfg().StoreInterval == -1 {
		return t.dm.SetTrend(ctx, trnd)
	}

	// schedule the asynchronous save, relies for Trend to be in cache
	t.sTrndsMux.Lock()
	t.storedTrends.Add(trnd.TenantID())
	t.sTrndsMux.Unlock()
	return
}

// storeTrends stores modified trends from cache in dataDB
// Reschedules failed trend IDs for next storage cycle.
// This function is safe for concurrent use.
func (t *TrendS) storeTrends(ctx *context.Context) {
	var failedTrndIDs []string
	for {
		t.sTrndsMux.Lock()
		trndID := t.storedTrends.GetOne()
		if trndID != utils.EmptyString {
			t.storedTrends.Remove(trndID)
		}
		t.sTrndsMux.Unlock()
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
		if err := t.dm.SetTrend(ctx, trnd); err != nil {
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
		t.sTrndsMux.Lock()
		t.storedTrends.AddSlice(failedTrndIDs)
		t.sTrndsMux.Unlock()
	}
}

// asyncStoreTrends runs as a background process that periodically calls storeTrends.
func (t *TrendS) asyncStoreTrends(ctx *context.Context) {
	storeInterval := t.cfg.TrendSCfg().StoreInterval
	if storeInterval <= 0 {
		close(t.storingStopped)
		return
	}
	for {
		t.storeTrends(ctx)
		select {
		case <-t.trendStop:
			close(t.storingStopped)
			return
		case <-time.After(storeInterval): // continue to another storing loop
		}
	}
}

// StartTrendS activates the Cron service with scheduled trend queries.
func (t *TrendS) StartTrendS(ctx *context.Context) error {
	if err := t.scheduleAutomaticQueries(ctx); err != nil {
		return err
	}
	t.crn.Start()
	go t.asyncStoreTrends(ctx)
	return nil
}

// StopTrendS gracefully shuts down Cron tasks and trend operations.
func (t *TrendS) StopTrendS() {
	timeEnd := time.Now().Add(t.cfg.CoreSCfg().ShutdownTimeout)

	crnctx := t.crn.Stop()
	close(t.trendStop)

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
	case <-t.storingStopped:
	case <-time.After(time.Until(timeEnd)):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for TrendS to finish",
				utils.TrendS))
		return
	}
}

// Reload restarts trend services with updated configuration.
func (t *TrendS) Reload(ctx *context.Context) {
	crnctx := t.crn.Stop()
	close(t.trendStop)
	<-crnctx.Done()
	<-t.storingStopped
	t.trendStop = make(chan struct{})
	t.storingStopped = make(chan struct{})
	t.crn.Start()
	go t.asyncStoreTrends(ctx)
}

// scheduleAutomaticQueries schedules initial trend queries based on configuration.
func (t *TrendS) scheduleAutomaticQueries(ctx *context.Context) error {
	schedData := make(map[string][]string)
	for k, v := range t.cfg.TrendSCfg().ScheduledIDs {
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
		qrydData, err := t.dm.GetTrendProfileIDs(ctx, tnts)
		if err != nil {
			return err
		}
		for tnt, ids := range qrydData {
			schedData[tnt] = ids
		}
	}
	for tnt, tIDs := range schedData {
		if _, err := t.scheduleTrendQueries(ctx, tnt, tIDs); err != nil {
			return err
		}
	}
	return nil
}

// scheduleTrendQueries schedules or reschedules specific trend queries
// Safe for concurrent use.
func (t *TrendS) scheduleTrendQueries(ctx *context.Context, tnt string, tIDs []string) (scheduled int, err error) {
	var partial bool
	t.crnTQsMux.Lock()
	if _, has := t.crnTQs[tnt]; !has {
		t.crnTQs[tnt] = make(map[string]cron.EntryID)
	}
	t.crnTQsMux.Unlock()
	for _, tID := range tIDs {
		t.crnTQsMux.RLock()
		if entryID, has := t.crnTQs[tnt][tID]; has {
			t.crn.Remove(entryID) // deschedule the query
		}
		t.crnTQsMux.RUnlock()
		if tP, err := t.dm.GetTrendProfile(ctx, tnt, tID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving TrendProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else if entryID, err := t.crn.AddFunc(tP.Schedule,
			func() { t.computeTrend(ctx, tP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling TrendProfile <%s:%s>, error: <%s>",
					utils.TrendS, tnt, tID, err.Error()))
			partial = true
		} else { // log the entry ID for debugging
			t.crnTQsMux.Lock()
			t.crnTQs[tP.Tenant][tP.ID] = entryID
			t.crnTQsMux.Unlock()
			scheduled++
		}
	}
	if partial {
		return 0, utils.ErrPartiallyExecuted
	}
	return
}
