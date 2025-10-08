/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package rankings

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

// RankingS implements the logic for processing and managing rankings based on stat metrics.
type RankingS struct {
	dm      *engine.DataManager
	connMgr *engine.ConnManager
	filterS *engine.FilterS
	cgrcfg  *config.CGRConfig

	crn            *cron.Cron                         // cron reference
	crnRQsMux      *sync.RWMutex                      // protects the crnTQs
	crnRQs         map[string]map[string]cron.EntryID // save the EntryIDs for rankingQueries so we can reschedule them when needed
	sRksMux        sync.RWMutex                       // protects storedRankings
	storedRankings utils.StringSet                    // keep a record of RankingS which need saving, map[rankingTenanrkID]bool
	storingStopped chan struct{}                      // signal back that the operations were stopped
	rankingStop    chan struct{}                      // signal to stop all operations
}

// NewRankingS creates a new RankingS service.
func NewRankingS(dm *engine.DataManager,
	connMgr *engine.ConnManager,
	filterS *engine.FilterS,
	cgrcfg *config.CGRConfig) *RankingS {
	return &RankingS{
		dm:             dm,
		connMgr:        connMgr,
		filterS:        filterS,
		cgrcfg:         cgrcfg,
		crn:            cron.New(),
		crnRQsMux:      new(sync.RWMutex),
		crnRQs:         make(map[string]map[string]cron.EntryID),
		storedRankings: make(utils.StringSet),
		storingStopped: make(chan struct{}),
		rankingStop:    make(chan struct{}),
	}
}

// computeRanking queries stats and builds the Ranking based on RankingProfile configuration.
// Called by Cron service at scheduled intervals.
func (r *RankingS) computeRanking(ctx *context.Context, rkP *utils.RankingProfile) {
	rk, err := r.dm.GetRanking(ctx, rkP.Tenant, rkP.ID, true, true, utils.NonTransactional)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying RankingProfile with ID: <%s:%s> dm error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	rk.Lock()
	defer rk.Unlock()
	if rk.Config() == nil {
		rk.SetConfig(rkP)
	}
	rk.LastUpdate = time.Now()
	rk.Metrics = make(map[string]map[string]float64) // reset previous values
	rk.SortedStatIDs = make([]string, 0)
	for _, statID := range rkP.StatIDs {
		var floatMetrics map[string]float64
		if err := r.connMgr.Call(context.Background(), r.cgrcfg.RankingSCfg().StatSConns,
			utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: rkP.Tenant, ID: statID}},
			&floatMetrics); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> computing Ranking with ID: <%s:%s> for Stats <%s> error: <%s>",
					utils.RankingS, rkP.Tenant, rkP.ID, statID, err.Error()))
			return
		}
		if len(rk.MetricIDs()) != 0 {
			for metricID := range floatMetrics {
				if _, has := rk.MetricIDs()[statID]; !has {
					delete(floatMetrics, metricID)
				}
			}
		}

		if len(floatMetrics) != 0 {
			rk.Metrics[statID] = make(map[string]float64)
		}
		for metricID, val := range floatMetrics {
			rk.Metrics[statID][metricID] = val
		}
	}
	if rk.SortedStatIDs, err = utils.RankingSortStats(rkP.Sorting,
		rkP.SortingParameters, rk.Metrics); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> sorting stats for Ranking with ID: <%s:%s> error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	if err = r.storeRanking(ctx, rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> storing Ranking with ID: <%s:%s> DM error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	if err := r.processThresholds(rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Ranking with id <%s:%s> error: <%s> with ThresholdS",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
	}
	if err := r.processEEs(rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with EEs",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
	}
}

// processThresholds sends the computed ranking to ThresholdS.
func (r *RankingS) processThresholds(rk *utils.Ranking) (err error) {
	if len(rk.SortedStatIDs) == 0 {
		return
	}
	if len(r.cgrcfg.RankingSCfg().ThresholdSConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.RankingUpdate,
	}
	var thIDs []string
	if len(rk.Config().ThresholdIDs) != 0 {
		if len(rk.Config().ThresholdIDs) == 1 &&
			rk.Config().ThresholdIDs[0] == utils.MetaNone {
			return
		}
		thIDs = make([]string, len(rk.Config().ThresholdIDs))
		copy(thIDs, rk.Config().ThresholdIDs)
	}
	opts[utils.OptsThresholdsProfileIDs] = thIDs
	sortedStatIDs := make([]string, len(rk.SortedStatIDs))
	copy(sortedStatIDs, rk.SortedStatIDs)
	ev := &utils.CGREvent{
		Tenant:  rk.Tenant,
		ID:      utils.GenUUID(),
		APIOpts: opts,
		Event: map[string]any{
			utils.RankingID:     rk.ID,
			utils.LastUpdate:    rk.LastUpdate,
			utils.SortedStatIDs: sortedStatIDs,
		},
	}
	var withErrs bool
	var rkIDs []string
	if err := r.connMgr.Call(context.TODO(), r.cgrcfg.RankingSCfg().ThresholdSConns,
		utils.ThresholdSv1ProcessEvent, ev, &rkIDs); err != nil &&
		(len(thIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %s processing event %+v with ThresholdS.", utils.RankingS, err.Error(), ev))
		withErrs = true
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEEs sends the computed ranking to EEs.
func (r *RankingS) processEEs(rk *utils.Ranking) (err error) {
	if len(rk.SortedStatIDs) == 0 {
		return
	}
	if len(r.cgrcfg.RankingSCfg().EEsConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.RankingUpdate,
	}
	sortedStatIDs := make([]string, len(rk.SortedStatIDs))
	copy(sortedStatIDs, rk.SortedStatIDs)
	ev := &utils.CGREvent{
		Tenant:  rk.Tenant,
		ID:      utils.GenUUID(),
		APIOpts: opts,
		Event: map[string]any{
			utils.RankingID:     rk.ID,
			utils.LastUpdate:    rk.LastUpdate,
			utils.SortedStatIDs: sortedStatIDs,
		},
	}
	var withErrs bool
	var reply map[string]map[string]any
	if err := r.connMgr.Call(context.TODO(), r.cgrcfg.RankingSCfg().EEsConns,
		utils.EeSv1ProcessEvent, ev, &reply); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> error: %q processing event %+v with EEs.", utils.RankingS, err.Error(), ev))
		withErrs = true
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// storeRanking stores or schedules the ranking for storage based on "store_interval".
func (r *RankingS) storeRanking(ctx *context.Context, rk *utils.Ranking) (err error) {
	if r.cgrcfg.RankingSCfg().StoreInterval == 0 {
		return
	}
	if r.cgrcfg.RankingSCfg().StoreInterval == -1 {
		return r.dm.SetRanking(ctx, rk)
	}
	// schedule the asynchronous save, relies for Ranking to be in cache
	r.sRksMux.Lock()
	r.storedRankings.Add(rk.Config().TenantID())
	r.sRksMux.Unlock()
	return
}

// storeRankings stores modified rankings from cache in dataDB
// Reschedules failed ranking IDs for next storage cycle.
// This function is safe for concurrent use.
func (r *RankingS) storeRankings(ctx *context.Context) {
	var failedRkIDs []string
	for {
		r.sRksMux.Lock()
		rkID := r.storedRankings.GetOne()
		if rkID != utils.EmptyString {
			r.storedRankings.Remove(rkID)
		}
		r.sRksMux.Unlock()
		if rkID == utils.EmptyString {
			break // no more keys, backup completed
		}
		rkIf, ok := engine.Cache.Get(utils.CacheRankings, rkID)
		if !ok || rkIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache Ranking with ID: %q",
					utils.RankingS, rkID))
			failedRkIDs = append(failedRkIDs, rkID) // record failure so we can schedule it for next backup
			continue
		}
		rk := rkIf.(*utils.Ranking)
		rk.RLock()
		if err := r.dm.SetRanking(ctx, rk); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed storing Trend with ID: %q, err: %q",
					utils.RankingS, rkID, err))
			failedRkIDs = append(failedRkIDs, rkID) // record failure so we can schedule it for next backup
		}
		rk.RUnlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRkIDs) != 0 { // there were errors on save, schedule the keys for next backup
		r.sRksMux.Lock()
		r.storedRankings.AddSlice(failedRkIDs)
		r.sRksMux.Unlock()
	}
}

// asyncStoreRankings runs as a background process that periodically calls storeRankings.
func (r *RankingS) asyncStoreRankings(ctx *context.Context) {
	storeInterval := r.cgrcfg.RankingSCfg().StoreInterval
	if storeInterval <= 0 {
		close(r.storingStopped)
		return
	}
	for {
		r.storeRankings(ctx)
		select {
		case <-r.rankingStop:
			close(r.storingStopped)
			return
		case <-time.After(storeInterval): // continue to another storing loop
		}
	}
}

// StartRankingS activates the Cron service with scheduled ranking queries.
func (r *RankingS) StartRankingS(ctx *context.Context) (err error) {
	if err = r.scheduleAutomaticQueries(ctx); err != nil {
		return
	}
	r.crn.Start()
	go r.asyncStoreRankings(ctx)
	return
}

// StopRankingS gracefully shuts down Cron tasks and ranking operations.
func (r *RankingS) StopRankingS() {
	timeEnd := time.Now().Add(r.cgrcfg.CoreSCfg().ShutdownTimeout)

	ctx := r.crn.Stop()
	close(r.rankingStop)

	// Wait for cron
	select {
	case <-ctx.Done():
	case <-time.After(time.Until(timeEnd)):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for Cron to finish",
				utils.RankingS))
		return
	}
	// Wait for backup and other operations
	select {
	case <-r.storingStopped:
	case <-time.After(time.Until(timeEnd)):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for RankingS to finish",
				utils.RankingS))
		return
	}
}

// Reload restarts ranking services with updated configuration.
func (r *RankingS) Reload(ctx *context.Context) {
	crnCtx := r.crn.Stop()
	close(r.rankingStop)
	<-crnCtx.Done()
	<-r.storingStopped
	r.rankingStop = make(chan struct{})
	r.storingStopped = make(chan struct{})
	r.crn.Start()
	go r.asyncStoreRankings(ctx)
}

// scheduleAutomaticQueries schedules initial ranking queries based on configuration.
func (r *RankingS) scheduleAutomaticQueries(ctx *context.Context) error {
	schedData := make(map[string][]string)
	for k, v := range r.cgrcfg.RankingSCfg().ScheduledIDs {
		schedData[k] = v
	}
	var tnts []string
	if len(schedData) == 0 {
		tnts = make([]string, 0)
	}
	for tnt, rkIDs := range schedData {
		if len(rkIDs) == 0 {
			tnts = append(tnts, tnt)
		}
	}
	if tnts != nil {
		qrydData, err := r.dm.GetRankingProfileIDs(ctx, tnts)
		if err != nil {
			return err
		}
		for tnt, ids := range qrydData {
			schedData[tnt] = ids
		}
	}
	for tnt, rkIDs := range schedData {
		if _, err := r.scheduleRankingQueries(ctx, tnt, rkIDs); err != nil {
			return err
		}
	}
	return nil
}

// scheduleRankingQueries schedules or reschedules specific ranking queries.
// Safe for concurrent use.
func (r *RankingS) scheduleRankingQueries(ctx *context.Context,
	tnt string, rkIDs []string) (scheduled int, err error) {
	var partial bool
	r.crnRQsMux.Lock()
	if _, has := r.crnRQs[tnt]; !has {
		r.crnRQs[tnt] = make(map[string]cron.EntryID)
	}
	r.crnRQsMux.Unlock()
	for _, rkID := range rkIDs {
		r.crnRQsMux.RLock()
		if entryID, has := r.crnRQs[tnt][rkID]; has {
			r.crn.Remove(entryID) // deschedule the query
		}
		r.crnRQsMux.RUnlock()
		if rkP, err := r.dm.GetRankingProfile(ctx, tnt, rkID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving RankingProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.RankingS, tnt, rkID, err.Error()))
			partial = true
		} else if entryID, err := r.crn.AddFunc(rkP.Schedule,
			func() { r.computeRanking(ctx, rkP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling RankingProfile <%s:%s>, error: <%s>",
					utils.RankingS, tnt, rkID, err.Error()))
			partial = true
		} else { // log the entry ID for debugging
			r.crnRQsMux.Lock()
			r.crnRQs[rkP.Tenant][rkP.ID] = entryID
			r.crnRQsMux.Unlock()
			scheduled++
		}
	}
	if partial {
		return 0, utils.ErrPartiallyExecuted
	}
	return
}
