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
	"runtime"
	"slices"
	"strings"
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
		dm:             dm,
		connMgr:        connMgr,
		filterS:        filterS,
		cgrcfg:         cgrcfg,
		crn:            cron.New(),
		crnTQsMux:      new(sync.RWMutex),
		crnTQs:         make(map[string]map[string]cron.EntryID),
		storedTrends:   make(utils.StringSet),
		storingStopped: make(chan struct{}),
		trendStop:      make(chan struct{}),
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

	storedTrends   utils.StringSet // keep a record of trends which need saving, map[trendTenantID]bool
	sTrndsMux      sync.RWMutex    // protects storedTrends
	storingStopped chan struct{}   // signal back that the operations were stopped

	trendStop chan struct{} // signal to stop all operations

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
				"<%s> computing trend with id: <%s:%s> for stats <%s> error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, tP.StatID, err.Error()))
		return
	}
	trnd, err := tS.dm.GetTrend(tP.Tenant, tP.ID, true, true, utils.NonTransactional)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying trend with id: <%s:%s> dm error: <%s>",
				utils.TrendS, tP.Tenant, tP.ID, err.Error()))
		return
	}
	trnd.tMux.Lock()
	defer trnd.tMux.Unlock()
	if trnd.tPrfl == nil {
		trnd.tPrfl = tP
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
		trnd.Metrics = make(map[time.Time]map[string]*MetricWithTrend)
	}
	trnd.Metrics[now] = make(map[string]*MetricWithTrend)
	for _, mID := range metrics {
		mWt := &MetricWithTrend{ID: mID}
		var has bool
		if mWt.Value, has = floatMetrics[mID]; !has { // no stats computed for metric
			mWt.Value = -1.0
			mWt.TrendLabel = utils.NotAvailable
			continue
		}
		if mWt.TrendGrowth, err = trnd.getTrendGrowth(mID, mWt.Value, tP.CorrelationType,
			tS.cgrcfg.GeneralCfg().RoundingDecimals); err != nil {
			mWt.TrendLabel = utils.NotAvailable
		} else {
			mWt.TrendLabel = trnd.getTrendLabel(mWt.TrendGrowth, tP.Tolerance)
		}
		trnd.Metrics[now][mWt.ID] = mWt
		trnd.indexesAppendMetric(mWt, now)
	}
	if err = tS.storeTrend(trnd); err != nil {
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

// processThresholds will pass the Trend event to ThresholdS
func (tS *TrendS) processThresholds(trnd *Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.tPrfl.MinItems {
		return
	}
	if len(tS.cgrcfg.TrendSCfg().ThresholdSConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.TrendUpdate,
	}
	var thIDs []string
	if len(trnd.tPrfl.ThresholdIDs) != 0 {
		if len(trnd.tPrfl.ThresholdIDs) == 1 &&
			trnd.tPrfl.ThresholdIDs[0] == utils.MetaNone {
			return
		}
		thIDs = make([]string, len(trnd.tPrfl.ThresholdIDs))
		copy(thIDs, trnd.tPrfl.ThresholdIDs)
	}
	opts[utils.OptsThresholdsProfileIDs] = thIDs
	ts := trnd.asTrendSummary()
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
	if err := tS.connMgr.Call(context.TODO(), tS.cgrcfg.TrendSCfg().ThresholdSConns,
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

// processEEs will pass the Trend event to EEs
func (tS *TrendS) processEEs(trnd *Trend) (err error) {
	if len(trnd.RunTimes) == 0 ||
		len(trnd.RunTimes) < trnd.tPrfl.MinItems {
		return
	}
	if len(tS.cgrcfg.TrendSCfg().EEsConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.TrendUpdate,
	}
	ts := trnd.asTrendSummary()
	trndEv := &CGREventWithEeIDs{
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
		EeIDs: tS.cgrcfg.TrendSCfg().EEsExporterIDs,
	}
	var withErrs bool
	var reply map[string]map[string]any
	if err := tS.connMgr.Call(context.TODO(), tS.cgrcfg.TrendSCfg().EEsConns,
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

// storeTrend will store or schedule the trend based on settings
func (tS *TrendS) storeTrend(trnd *Trend) (err error) {
	if tS.cgrcfg.TrendSCfg().StoreInterval == 0 {
		return
	}
	if tS.cgrcfg.TrendSCfg().StoreInterval == -1 {
		return tS.dm.SetTrend(trnd)
	}

	// schedule the asynchronous save, relies for Trend to be in cache
	tS.sTrndsMux.Lock()
	tS.storedTrends.Add(trnd.TenantID())
	tS.sTrndsMux.Unlock()
	return
}

// storeTrends will do one round for saving modified trends
//
//		from cache to dataDB
//	 designed to run asynchronously
func (tS *TrendS) storeTrends() {
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
		trndIf, ok := Cache.Get(utils.CacheTrends, trndID)
		if !ok || trndIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache Trend with ID: %q",
					utils.TrendS, trndID))
			failedTrndIDs = append(failedTrndIDs, trndID) // record failure so we can schedule it for next backup
			continue
		}
		trnd := trndIf.(*Trend)
		trnd.tMux.RLock()
		if err := tS.dm.SetTrend(trnd); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed storing Trend with ID: %q, err: %q",
					utils.TrendS, trndID, err))
			failedTrndIDs = append(failedTrndIDs, trndID) // record failure so we can schedule it for next backup
		}
		trnd.tMux.RUnlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedTrndIDs) != 0 { // there were errors on save, schedule the keys for next backup
		tS.sTrndsMux.Lock()
		tS.storedTrends.AddSlice(failedTrndIDs)
		tS.sTrndsMux.Unlock()
	}
}

// asyncStoreTrends runs as a backround process, calling storeTrends based on storeInterval
func (tS *TrendS) asyncStoreTrends() {
	storeInterval := tS.cgrcfg.TrendSCfg().StoreInterval
	if storeInterval <= 0 {
		close(tS.storingStopped)
		return
	}
	for {
		tS.storeTrends()
		select {
		case <-tS.trendStop:
			close(tS.storingStopped)
			return
		case <-time.After(storeInterval): // continue to another storing loop
		}
	}
}

// StartCron will activates the Cron, together with all scheduled Trend queries
func (tS *TrendS) StartTrendS() error {
	if err := tS.scheduleAutomaticQueries(); err != nil {
		return err
	}
	tS.crn.Start()
	go tS.asyncStoreTrends()
	return nil
}

// StopCron will shutdown the Cron tasks
func (tS *TrendS) StopTrendS() {
	timeEnd := time.Now().Add(tS.cgrcfg.CoreSCfg().ShutdownTimeout)

	ctx := tS.crn.Stop()
	close(tS.trendStop)

	// Wait for cron
	select {
	case <-ctx.Done():
	case <-time.After(timeEnd.Sub(time.Now())):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for Cron to finish",
				utils.TrendS))
		return
	}
	// Wait for backup and other operations
	select {
	case <-tS.storingStopped:
	case <-time.After(timeEnd.Sub(time.Now())):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for TrendS to finish",
				utils.TrendS))
		return
	}
}

func (tS *TrendS) Reload() {
	ctx := tS.crn.Stop()
	close(tS.trendStop)
	<-ctx.Done()
	<-tS.storingStopped
	tS.trendStop = make(chan struct{})
	tS.storingStopped = make(chan struct{})
	tS.crn.Start()
	go tS.asyncStoreTrends()
}

// scheduleAutomaticQueries will schedule the queries at start/reload based on configured
func (tS *TrendS) scheduleAutomaticQueries() error {
	schedData := make(map[string][]string)
	for k, v := range tS.cgrcfg.TrendSCfg().ScheduledIDs {
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
		qrydData, err := tS.dm.GetTrendProfileIDs(tnts)
		if err != nil {
			return err
		}
		for tnt, ids := range qrydData {
			schedData[tnt] = ids
		}
	}
	for tnt, tIDs := range schedData {
		if _, err := tS.scheduleTrendQueries(context.TODO(), tnt, tIDs); err != nil {
			return err
		}
	}
	return nil
}

// scheduleTrendQueries will schedule/re-schedule specific trend queries
func (tS *TrendS) scheduleTrendQueries(_ *context.Context, tnt string, tIDs []string) (scheduled int, err error) {
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

// V1GetTrend is the API to return the trend Metrics
// The number of runTimes can be filtered based on indexes and times provided as arguments
//
//	in this way being possible to work with paginators
func (tS *TrendS) V1GetTrend(ctx *context.Context, arg *utils.ArgGetTrend, retTrend *Trend) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var trnd *Trend
	if trnd, err = tS.dm.GetTrend(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	retTrend.Tenant = trnd.Tenant // avoid vet complaining for mutex copying
	retTrend.ID = trnd.ID
	startIdx := arg.RunIndexStart
	if startIdx > len(trnd.RunTimes) {
		startIdx = len(trnd.RunTimes)
	}
	endIdx := arg.RunIndexEnd
	if endIdx > len(trnd.RunTimes) ||
		endIdx < startIdx ||
		endIdx == 0 {
		endIdx = len(trnd.RunTimes)
	}
	runTimes := trnd.RunTimes[startIdx:endIdx]
	if len(runTimes) == 0 {
		return utils.ErrNotFound
	}
	var tStart, tEnd time.Time
	if arg.RunTimeStart == utils.EmptyString {
		tStart = runTimes[0]
	} else if tStart, err = utils.ParseTimeDetectLayout(arg.RunTimeStart, tS.cgrcfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	if arg.RunTimeEnd == utils.EmptyString {
		tEnd = runTimes[len(runTimes)-1].Add(time.Duration(1))
	} else if tEnd, err = utils.ParseTimeDetectLayout(arg.RunTimeEnd, tS.cgrcfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	retTrend.RunTimes = make([]time.Time, 0, len(runTimes))
	for _, runTime := range runTimes {
		if !runTime.Before(tStart) && runTime.Before(tEnd) {
			retTrend.RunTimes = append(retTrend.RunTimes, runTime)
		}
	}
	if len(retTrend.RunTimes) == 0 { // filtered out all
		return utils.ErrNotFound
	}
	retTrend.Metrics = make(map[time.Time]map[string]*MetricWithTrend)
	for _, runTime := range retTrend.RunTimes {
		retTrend.Metrics[runTime] = trnd.Metrics[runTime]
	}
	return
}

func (tS *TrendS) V1GetScheduledTrends(ctx *context.Context, args *utils.ArgScheduledTrends, schedTrends *[]utils.ScheduledTrend) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	tS.crnTQsMux.RLock()
	defer tS.crnTQsMux.RUnlock()
	trendIDsMp, has := tS.crnTQs[tnt]
	if !has {
		return utils.ErrNotFound
	}
	var scheduledTrends []utils.ScheduledTrend
	var entryIds map[string]cron.EntryID
	if len(args.TrendIDPrefixes) == 0 {
		entryIds = trendIDsMp
	} else {
		entryIds = make(map[string]cron.EntryID)
		for _, tID := range args.TrendIDPrefixes {
			for key, entryID := range trendIDsMp {
				if strings.HasPrefix(key, tID) {
					entryIds[key] = entryID
				}
			}
		}
	}
	if len(entryIds) == 0 {
		return utils.ErrNotFound
	}
	var entry cron.Entry
	for id, entryID := range entryIds {
		entry = tS.crn.Entry(entryID)
		if entry.ID == 0 {
			continue
		}
		scheduledTrends = append(scheduledTrends, utils.ScheduledTrend{
			TrendID:  id,
			Next:     entry.Next,
			Previous: entry.Prev,
		})
	}
	slices.SortFunc(scheduledTrends, func(a, b utils.ScheduledTrend) int {
		return a.Next.Compare(b.Next)
	})
	*schedTrends = scheduledTrends
	return nil
}

func (tS *TrendS) V1GetTrendSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *TrendSummary) (err error) {
	var trnd *Trend
	if trnd, err = tS.dm.GetTrend(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	trnd.tMux.RLock()
	trndS := trnd.asTrendSummary()
	trnd.tMux.RUnlock()
	*reply = *trndS
	return
}
