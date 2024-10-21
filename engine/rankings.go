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

// NewRankingS is the constructor for RankingS
func NewRankingS(dm *DataManager,
	connMgr *ConnManager,
	filterS *FilterS,
	cgrcfg *config.CGRConfig) *RankingS {
	return &RankingS{
		dm:             dm,
		connMgr:        connMgr,
		filterS:        filterS,
		cgrcfg:         cgrcfg,
		crn:            cron.New(),
		crnTQsMux:      new(sync.RWMutex),
		crnTQs:         make(map[string]map[string]cron.EntryID),
		storedRankings: make(utils.StringSet),
		storingStopped: make(chan struct{}),
		rankingStop:    make(chan struct{}),
	}
}

// RankingS is responsible of implementing the logic of RankingService
type RankingS struct {
	dm      *DataManager
	connMgr *ConnManager
	filterS *FilterS
	cgrcfg  *config.CGRConfig

	crn *cron.Cron // cron reference

	crnTQsMux *sync.RWMutex                      // protects the crnTQs
	crnTQs    map[string]map[string]cron.EntryID // save the EntryIDs for rankingQueries so we can reschedule them when needed

	storedRankings utils.StringSet // keep a record of RankingS which need saving, map[rankingTenanrkID]bool
	sRksMux        sync.RWMutex    // protects storedRankings
	storingStopped chan struct{}   // signal back that the operations were stopped

	rankingStop chan struct{} // signal to stop all operations

}

// computeRanking will query the stats and build the Ranking for them
//
//	it is to be called by Cron service
func (rkS *RankingS) computeRanking(rkP *RankingProfile) {
	rk, err := rkS.dm.GetRanking(rkP.Tenant, rkP.ID, true, true, utils.NonTransactional)
	if err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> querying RankingProfile with ID: <%s:%s> dm error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	rk.rMux.Lock()
	defer rk.rMux.Unlock()
	if rk.rkPrfl == nil {
		rk.rkPrfl = rkP
	}
	rk.LastUpdate = time.Now()
	rk.Metrics = make(map[string]map[string]float64) // reset previous values
	rk.SortedStatIDs = make([]string, 0)
	for _, statID := range rkP.StatIDs {
		var floatMetrics map[string]float64
		if err := rkS.connMgr.Call(context.Background(), rkS.cgrcfg.RankingSCfg().StatSConns,
			utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: rkP.Tenant, ID: statID}},
			&floatMetrics); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> computing Ranking with ID: <%s:%s> for Stats <%s> error: <%s>",
					utils.RankingS, rkP.Tenant, rkP.ID, statID, err.Error()))
			return
		}
		if len(rk.metricIDs) != 0 {
			for metricID := range floatMetrics {
				if _, has := rk.metricIDs[statID]; !has {
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

	if rk.SortedStatIDs, err = rankingSortStats(rkP.Sorting,
		rkP.SortingParameters, rk.Metrics); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> sorting stats for Ranking with ID: <%s:%s> error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	if err = rkS.storeRanking(rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> storing Ranking with ID: <%s:%s> DM error: <%s>",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
		return
	}
	if err := rkS.processThresholds(rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Ranking with id <%s:%s> error: <%s> with ThresholdS",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
	}
	if err := rkS.processEEs(rk); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> Trend with id <%s:%s> error: <%s> with EEs",
				utils.RankingS, rkP.Tenant, rkP.ID, err.Error()))
	}
}

// processThresholds will pass the Ranking event to ThresholdS
func (rkS *RankingS) processThresholds(rk *Ranking) (err error) {
	if len(rk.SortedStatIDs) == 0 {
		return
	}
	if len(rkS.cgrcfg.RankingSCfg().ThresholdSConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.RankingUpdate,
	}
	var thIDs []string
	if len(rk.rkPrfl.ThresholdIDs) != 0 {
		if len(rk.rkPrfl.ThresholdIDs) == 1 &&
			rk.rkPrfl.ThresholdIDs[0] == utils.MetaNone {
			return
		}
		thIDs = make([]string, len(rk.rkPrfl.ThresholdIDs))
		copy(thIDs, rk.rkPrfl.ThresholdIDs)
	}
	opts[utils.OptsThresholdsProfileIDs] = thIDs
	ev := &utils.CGREvent{
		Tenant:  rk.Tenant,
		ID:      utils.GenUUID(),
		APIOpts: opts,
		Event: map[string]any{
			utils.RankingID:     rk.ID,
			utils.LastUpdate:    rk.LastUpdate,
			utils.SortedStatIDs: copy([]string{}, rk.SortedStatIDs),
		},
	}
	var withErrs bool
	var rkIDs []string
	if err := rkS.connMgr.Call(context.TODO(), rkS.cgrcfg.RankingSCfg().ThresholdSConns,
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

// processEEs will pass the Ranking event to EEs
func (rkS *RankingS) processEEs(rk *Ranking) (err error) {
	if len(rk.SortedStatIDs) == 0 {
		return
	}
	if len(rkS.cgrcfg.RankingSCfg().EEsConns) == 0 {
		return
	}
	opts := map[string]any{
		utils.MetaEventType: utils.RankingUpdate,
	}
	ev := &utils.CGREvent{
		Tenant:  rk.Tenant,
		ID:      utils.GenUUID(),
		APIOpts: opts,
		Event: map[string]any{
			utils.RankingID:     rk.ID,
			utils.LastUpdate:    rk.LastUpdate,
			utils.SortedStatIDs: copy([]string{}, rk.SortedStatIDs),
		},
	}
	var withErrs bool
	var reply map[string]map[string]any
	if err := rkS.connMgr.Call(context.TODO(), rkS.cgrcfg.RankingSCfg().EEsConns,
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

// storeTrend will store or schedule the trend based on settings
func (rkS *RankingS) storeRanking(rk *Ranking) (err error) {
	if rkS.cgrcfg.RankingSCfg().StoreInterval == 0 {
		return
	}
	if rkS.cgrcfg.RankingSCfg().StoreInterval == -1 {
		return rkS.dm.SetRanking(rk)
	}
	// schedule the asynchronous save, relies for Ranking to be in cache
	rkS.sRksMux.Lock()
	rkS.storedRankings.Add(rk.rkPrfl.TenantID())
	rkS.sRksMux.Unlock()
	return
}

// storeRankings will do one round for saving modified Rankings
//
//		from cache to dataDB
//	 designed to run asynchronously
func (rkS *RankingS) storeRankings() {
	var failedRkIDs []string
	for {
		rkS.sRksMux.Lock()
		rkID := rkS.storedRankings.GetOne()
		if rkID != utils.EmptyString {
			rkS.storedRankings.Remove(rkID)
		}
		rkS.sRksMux.Unlock()
		if rkID == utils.EmptyString {
			break // no more keys, backup completed
		}
		rkIf, ok := Cache.Get(utils.CacheRankings, rkID)
		if !ok || rkIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache Ranking with ID: %q",
					utils.RankingS, rkID))
			failedRkIDs = append(failedRkIDs, rkID) // record failure so we can schedule it for next backup
			continue
		}
		rk := rkIf.(*Ranking)
		rk.rMux.RLock()
		if err := rkS.dm.SetRanking(rk); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed storing Trend with ID: %q, err: %q",
					utils.RankingS, rkID, err))
			failedRkIDs = append(failedRkIDs, rkID) // record failure so we can schedule it for next backup
		}
		rk.rMux.RUnlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRkIDs) != 0 { // there were errors on save, schedule the keys for next backup
		rkS.sRksMux.Lock()
		rkS.storedRankings.AddSlice(failedRkIDs)
		rkS.sRksMux.Unlock()
	}
}

// asyncStoreRankings runs as a backround process, calling storeRankings based on storeInterval
func (rkS *RankingS) asyncStoreRankings() {
	storeInterval := rkS.cgrcfg.RankingSCfg().StoreInterval
	if storeInterval <= 0 {
		close(rkS.storingStopped)
		return
	}
	for {
		rkS.storeRankings()
		select {
		case <-rkS.rankingStop:
			close(rkS.storingStopped)
			return
		case <-time.After(storeInterval): // continue to another storing loop
		}
	}
}

// StartRankings will activates the Cron, together with all scheduled Ranking queries
func (rkS *RankingS) StartRankingS() (err error) {
	if err = rkS.scheduleAutomaticQueries(); err != nil {
		return
	}
	rkS.crn.Start()
	go rkS.asyncStoreRankings()
	return
}

// StopCron will shutdown the Cron tasks
func (rkS *RankingS) StopRankingS() {
	timeEnd := time.Now().Add(rkS.cgrcfg.CoreSCfg().ShutdownTimeout)

	ctx := rkS.crn.Stop()
	close(rkS.rankingStop)

	// Wait for cron
	select {
	case <-ctx.Done():
	case <-time.After(timeEnd.Sub(time.Now())):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for Cron to finish",
				utils.RankingS))
		return
	}
	// Wait for backup and other operations
	select {
	case <-rkS.storingStopped:
	case <-time.After(timeEnd.Sub(time.Now())):
		utils.Logger.Warning(
			fmt.Sprintf(
				"<%s> timeout waiting for RankingS to finish",
				utils.RankingS))
		return
	}
}

func (rkS *RankingS) Reload() {
	ctx := rkS.crn.Stop()
	close(rkS.rankingStop)
	<-ctx.Done()
	<-rkS.storingStopped
	rkS.rankingStop = make(chan struct{})
	rkS.storingStopped = make(chan struct{})
	rkS.crn.Start()
	go rkS.asyncStoreRankings()
}

// scheduleAutomaticQueries will schedule the queries at start/reload based on configured
func (rkS *RankingS) scheduleAutomaticQueries() error {
	schedData := make(map[string][]string)
	for k, v := range rkS.cgrcfg.RankingSCfg().ScheduledIDs {
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
		qrydData, err := rkS.dm.GetTrendProfileIDs(tnts)
		if err != nil {
			return err
		}
		for tnt, ids := range qrydData {
			schedData[tnt] = ids
		}
	}
	for tnt, rkIDs := range schedData {
		if _, err := rkS.scheduleRankingQueries(context.TODO(), tnt, rkIDs); err != nil {
			return err
		}
	}
	return nil
}

// scheduleTrendQueries will schedule/re-schedule specific trend queries
func (rkS *RankingS) scheduleRankingQueries(_ *context.Context,
	tnt string, rkIDs []string) (scheduled int, err error) {
	var partial bool
	rkS.crnTQsMux.Lock()
	if _, has := rkS.crnTQs[tnt]; !has {
		rkS.crnTQs[tnt] = make(map[string]cron.EntryID)
	}
	rkS.crnTQsMux.Unlock()
	for _, rkID := range rkIDs {
		rkS.crnTQsMux.RLock()
		if entryID, has := rkS.crnTQs[tnt][rkID]; has {
			rkS.crn.Remove(entryID) // deschedule the query
		}
		rkS.crnTQsMux.RUnlock()
		if rkP, err := rkS.dm.GetRankingProfile(tnt, rkID, true, true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> failed retrieving RankingProfile with id: <%s:%s> for scheduling, error: <%s>",
					utils.RankingS, tnt, rkID, err.Error()))
			partial = true
		} else if entryID, err := rkS.crn.AddFunc(utils.EmptyString,
			func() { rkS.computeRanking(rkP.Clone()) }); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf(
					"<%s> scheduling RankingProfile <%s:%s>, error: <%s>",
					utils.RankingS, tnt, rkID, err.Error()))
			partial = true
		} else { // log the entry ID for debugging
			rkS.crnTQsMux.Lock()
			rkS.crnTQs[rkP.Tenant][rkP.ID] = entryID
			rkS.crnTQsMux.Unlock()
		}
		scheduled += 1
	}
	if partial {
		return 0, utils.ErrPartiallyExecuted
	}
	return
}

// V1ScheduleQueries is the query for manually re-/scheduling Ranking Queries
func (rkS *RankingS) V1ScheduleQueries(ctx *context.Context, args *utils.ArgScheduleRankingQueries, scheduled *int) (err error) {
	if sched, errSched := rkS.scheduleRankingQueries(ctx, args.Tenant, args.RankingIDs); errSched != nil {
		return errSched
	} else {
		*scheduled = sched
	}
	return
}

// V1GetRanking is the API to return the Ranking instance
func (rkS *RankingS) V1GetRanking(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, retRanking *Ranking) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var rk *Ranking
	if rk, err = rkS.dm.GetRanking(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	rk.rMux.RLock()
	defer rk.rMux.RUnlock()
	retRanking.Tenant = rk.Tenant // avoid vet complaining for mutex copying
	retRanking.ID = rk.ID
	retRanking.Metrics = make(map[string]map[string]float64)
	for statID, metrics := range rk.Metrics {
		retRanking.Metrics[statID] = make(map[string]float64)
		for metricID, val := range metrics {
			retRanking.Metrics[statID][metricID] = val
		}
	}
	retRanking.Sorting = rk.Sorting
	copy(retRanking.SortingParameters, rk.SortingParameters)
	copy(retRanking.SortedStatIDs, rk.SortedStatIDs)
	return
}

// V1GetSchedule returns the active schedule for Raking queries
func (rkS *RankingS) V1GetSchedule(ctx *context.Context, args *utils.ArgScheduledRankings, schedRankings *[]utils.ScheduledRanking) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rkS.cgrcfg.GeneralCfg().DefaultTenant
	}
	rkS.crnTQsMux.RLock()
	defer rkS.crnTQsMux.RUnlock()
	trendIDsMp, has := rkS.crnTQs[tnt]
	if !has {
		return utils.ErrNotFound
	}
	var scheduledRankings []utils.ScheduledRanking
	var entryIds map[string]cron.EntryID
	if len(args.RankingIDPrefixes) == 0 {
		entryIds = trendIDsMp
	} else {
		entryIds = make(map[string]cron.EntryID)
		for _, rkID := range args.RankingIDPrefixes {
			for key, entryID := range trendIDsMp {
				if strings.HasPrefix(key, rkID) {
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
		entry = rkS.crn.Entry(entryID)
		if entry.ID == 0 {
			continue
		}
		scheduledRankings = append(scheduledRankings,
			utils.ScheduledRanking{
				RankingID: id,
				Next:      entry.Next,
				Previous:  entry.Prev,
			})
	}
	slices.SortFunc(scheduledRankings, func(a, b utils.ScheduledRanking) int {
		return a.Next.Compare(b.Next)
	})
	*schedRankings = scheduledRankings
	return nil
}

// V1GetRankingSummary returns a summary of ascending/descending stat of the last updated ranking
func (rS *RankingS) V1GetRankingSummary(ctx *context.Context, arg utils.TenantIDWithAPIOpts, reply *RankingSummary) (err error) {
	var rnk *Ranking
	if rnk, err = rS.dm.GetRanking(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return
	}
	rnk.rMux.RLock()
	rnkS := rnk.asRankingSummary()
	rnk.rMux.RUnlock()
	*reply = *rnkS
	return
}
