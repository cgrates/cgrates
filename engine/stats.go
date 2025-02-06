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
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// NewStatService initializes a StatService
func NewStatService(dm *DataManager, cgrcfg *config.CGRConfig,
	filterS *FilterS, connMgr *ConnManager) *StatS {
	return &StatS{
		dm:               dm,
		connMgr:          connMgr,
		fltrS:            filterS,
		cfg:              cgrcfg,
		storedStatQueues: make(utils.StringSet),
		loopStopped:      make(chan struct{}),
		stopBackup:       make(chan struct{}),
	}
}

// StatS builds stats for events
type StatS struct {
	dm               *DataManager
	connMgr          *ConnManager
	fltrS            *FilterS
	cfg              *config.CGRConfig
	loopStopped      chan struct{}
	stopBackup       chan struct{}
	storedStatQueues utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	ssqMux           sync.RWMutex    // protects storedStatQueues
}

// Reload stops the backupLoop and restarts it
func (sS *StatS) Reload(ctx *context.Context) {
	close(sS.stopBackup)
	<-sS.loopStopped // wait until the loop is done
	sS.stopBackup = make(chan struct{})
	go sS.runBackup(ctx)
}

// StartLoop starsS the gorutine with the backup loop
func (sS *StatS) StartLoop(ctx *context.Context) {
	go sS.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (sS *StatS) Shutdown(ctx *context.Context) {
	close(sS.stopBackup)
	sS.storeStats(ctx)
}

// runBackup will regularly store statQueues changed to dataDB
func (sS *StatS) runBackup(ctx *context.Context) {
	storeInterval := sS.cfg.StatSCfg().StoreInterval
	if storeInterval <= 0 {
		sS.loopStopped <- struct{}{}
		return
	}
	for {
		sS.storeStats(ctx)
		select {
		case <-sS.stopBackup:
			sS.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeStats represents one task of complete backup
func (sS *StatS) storeStats(ctx *context.Context) {
	var failedSqIDs []string
	for { // don't stop untill we store all dirty statQueues
		sS.ssqMux.Lock()
		sID := sS.storedStatQueues.GetOne()
		if sID != "" {
			sS.storedStatQueues.Remove(sID)
		}
		sS.ssqMux.Unlock()
		if sID == "" {
			break // no more keys, backup completed
		}
		sqIf, ok := Cache.Get(utils.CacheStatQueues, sID)
		if !ok || sqIf == nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed retrieving from cache stat queue with ID: %s",
					utils.StatService, sID))
			continue
		}
		s := sqIf.(*StatQueue)
		s.lock(utils.EmptyString)
		if err := sS.StoreStatQueue(ctx, s); err != nil {
			failedSqIDs = append(failedSqIDs, sID) // record failure so we can schedule it for next backup
		}
		s.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedSqIDs) != 0 { // there were errors on save, schedule the keys for next backup
		sS.ssqMux.Lock()
		sS.storedStatQueues.AddSlice(failedSqIDs)
		sS.ssqMux.Unlock()
	}
}

// StoreStatQueue stores the statQueue in DB and corrects dirty flag
func (sS *StatS) StoreStatQueue(ctx *context.Context, sq *StatQueue) (err error) {
	if sq.dirty == nil || !*sq.dirty {
		return
	}
	if err = sS.dm.SetStatQueue(ctx, sq); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed saving StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := sq.TenantID(); Cache.HasItem(utils.CacheStatQueues, tntID) { // only cache if previously there
		if err = Cache.Set(ctx, utils.CacheStatQueues, tntID, sq, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> failed caching StatQueue with ID: %s, error: %s",
					tntID, err.Error()))
			return
		}
	}
	*sq.dirty = false
	return
}

// matchingStatQueuesForEvent returns ordered list of matching statQueues which are active by the time of the call
func (sS *StatS) matchingStatQueuesForEvent(ctx *context.Context, tnt string, statsIDs []string, evNm utils.MapStorage, ignoreFilters bool) (sqs StatQueues, err error) {
	sqIDs := utils.NewStringSet(statsIDs)
	if len(sqIDs) == 0 {
		ignoreFilters = false
		sqIDs, err = MatchingItemIDsForEvent(ctx, evNm,
			sS.cfg.StatSCfg().StringIndexedFields,
			sS.cfg.StatSCfg().PrefixIndexedFields,
			sS.cfg.StatSCfg().SuffixIndexedFields,
			sS.cfg.StatSCfg().ExistsIndexedFields,
			sS.cfg.StatSCfg().NotExistsIndexedFields,
			sS.dm, utils.CacheStatFilterIndexes, tnt,
			sS.cfg.StatSCfg().IndexedSelects,
			sS.cfg.StatSCfg().NestedFields,
		)
		if err != nil {
			return
		}
	}
	sqs = make(StatQueues, 0, len(sqIDs))
	for sqID := range sqIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			statQueueProfileLockKey(tnt, sqID))
		var sqPrfl *StatQueueProfile
		if sqPrfl, err = sS.dm.GetStatQueueProfile(ctx, tnt, sqID, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			sqs.unlock()
			return
		}
		sqPrfl.lock(lkPrflID)
		if !ignoreFilters {
			var pass bool
			if pass, err = sS.fltrS.Pass(ctx, tnt, sqPrfl.FilterIDs,
				evNm); err != nil {
				sqPrfl.unlock()
				sqs.unlock()
				return nil, err
			} else if !pass {
				sqPrfl.unlock()
				continue
			}
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			statQueueLockKey(sqPrfl.Tenant, sqPrfl.ID))
		var sq *StatQueue
		if sq, err = sS.dm.GetStatQueue(ctx, sqPrfl.Tenant, sqPrfl.ID, true, true, utils.EmptyString); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			sqPrfl.unlock()
			sqs.unlock()
			return nil, err
		}
		sq.lock(lkID) // pass the lock into statQueue so we have it as reference
		if sqPrfl.Stored && sq.dirty == nil {
			sq.dirty = utils.BoolPointer(false)
		}
		if sqPrfl.TTL > 0 {
			sq.ttl = utils.DurationPointer(sqPrfl.TTL)
		}

		if sqPrfl.TTL == -1 && sqPrfl.QueueLength == -1 {
			sq.ttl = utils.DurationPointer(sqPrfl.TTL)
		}

		sq.sqPrfl = sqPrfl
		if sq.weight, err = WeightFromDynamics(ctx, sqPrfl.Weights,
			sS.fltrS, tnt, evNm); err != nil {
			return
		}
		sqs = append(sqs, sq)
	}
	if len(sqs) == 0 {
		return nil, utils.ErrNotFound
	}
	// All good, convert from Map to Slice so we can sort
	sqs.Sort()
	/*
		// verify the Blockers from the profiles
		for i, s := range sqs {
			// get the dynamic blocker from the profile and check if it pass trough its filters
			var blocker bool
			if blocker, err = BlockerFromDynamics(ctx, s.sqPrfl.Blockers, sS.fltrS, tnt, evNm); err != nil {
				return
			}
			if blocker && i != len(sqs)-1 { // blocker will stop processing and we are not at last index
				StatQueues(sqs[i+1:]).unlock()
				sqs = sqs[:i+1]
				break
			}
		}
	*/
	return
}

func (sS *StatS) getStatQueue(ctx *context.Context, tnt, id string) (sq *StatQueue, err error) {
	if sq, err = sS.dm.GetStatQueue(ctx, tnt, id, true, true, utils.EmptyString); err != nil {
		return
	}
	var removed int
	if removed, err = sq.remExpired(); err != nil || removed == 0 {
		return
	}
	sS.storeStatQueue(ctx, sq)
	return
}

// storeStatQueue will store the sq if needed
func (sS *StatS) storeStatQueue(ctx *context.Context, sq *StatQueue) {
	if sS.cfg.StatSCfg().StoreInterval != 0 && sq.dirty != nil { // don't save
		*sq.dirty = true // mark it to be saved
		if sS.cfg.StatSCfg().StoreInterval == -1 {
			sS.StoreStatQueue(ctx, sq)
		} else {
			sS.ssqMux.Lock()
			sS.storedStatQueues.Add(sq.TenantID())
			sS.ssqMux.Unlock()
		}
	}
}

// processThresholds will pass the event for statQueue to ThresholdS
func (sS *StatS) processThresholds(ctx *context.Context, sQs StatQueues, opts map[string]any) (err error) {
	if len(sS.cfg.StatSCfg().ThresholdSConns) == 0 {
		return
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	opts[utils.MetaEventType] = utils.StatUpdate
	var withErrs bool
	for _, sq := range sQs {
		if len(sq.sqPrfl.ThresholdIDs) == 1 &&
			sq.sqPrfl.ThresholdIDs[0] == utils.MetaNone {
			continue
		}
		opts[utils.OptsThresholdsProfileIDs] = sq.sqPrfl.ThresholdIDs
		thEv := &utils.CGREvent{
			Tenant: sq.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType: utils.StatUpdate,
				utils.StatID:    sq.ID,
			},
			APIOpts: opts,
		}
		for metricID, metric := range sq.SQMetrics {
			thEv.Event[metricID] = metric.GetValue()
		}

		var tIDs []string
		if err := sS.connMgr.Call(ctx, sS.cfg.StatSCfg().ThresholdSConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(sq.sqPrfl.ThresholdIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processThresholds will pass the event for statQueue to EEs
func (sS *StatS) processEEs(ctx *context.Context, sQs StatQueues, opts map[string]any) (err error) {
	if len(sS.cfg.StatSCfg().EEsConns) == 0 {
		return
	}
	var withErrs bool
	if opts == nil {
		opts = make(map[string]any)
	}
	for _, sq := range sQs {
		metrics := make(map[string]any)
		for metricID, metric := range sq.SQMetrics {
			metrics[metricID] = metric.GetValue()
		}
		cgrEv := &utils.CGREvent{
			Tenant: sq.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType: utils.StatUpdate,
				utils.StatID:    sq.ID,
				utils.Metrics:   metrics,
			},
			APIOpts: opts,
		}

		cgrEventWithID := &utils.CGREventWithEeIDs{
			CGREvent: cgrEv,
			EeIDs:    sS.cfg.StatSCfg().EEsExporterIDs,
		}
		var reply map[string]map[string]any
		if err := sS.connMgr.Call(ctx, sS.cfg.StatSCfg().EEsConns,
			utils.EeSv1ProcessEvent,
			&cgrEventWithID, &reply); err != nil &&
			err.Error() != utils.ErrNotFound.Error() {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> error: %s processing event %+v with EEs.", err.Error(), cgrEv))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// processEvent processes a new event, dispatching to matching queues.
// Queues matching are also cached to speed up
func (sS *StatS) processEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (statQueueIDs []string, err error) {
	evNm := args.AsDataProvider()
	var sqIDs []string
	if sqIDs, err = GetStringSliceOpts(ctx, tnt, evNm, sS.fltrS, sS.cfg.StatSCfg().Opts.ProfileIDs,
		config.StatsProfileIDsDftOpt, utils.OptsStatsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = GetBoolOpts(ctx, tnt, evNm, sS.fltrS, sS.cfg.StatSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	matchSQs, err := sS.matchingStatQueuesForEvent(ctx, tnt, sqIDs, evNm, ignFilters)
	if err != nil {
		return nil, err
	}
	defer matchSQs.unlock()

	statQueueIDs = matchSQs.IDs()
	var withErrors bool
	for idx, sq := range matchSQs {
		if err = sq.ProcessEvent(ctx, tnt, args.ID, sS.fltrS, evNm); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Queue: %s, ignoring event: %s, error: %s",
					sq.TenantID(), utils.ConcatenatedKey(tnt, args.ID), err.Error()))
			withErrors = true
		}
		sS.storeStatQueue(ctx, sq)
		// verify the Blockers from the profiles
		// get the dynamic blocker from the profile and check if it pass trough its filters
		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, sq.sqPrfl.Blockers, sS.fltrS, tnt, evNm); err != nil {
			return
		}
		if blocker && idx != len(matchSQs)-1 { // blocker will stop processing and we are not at last index
			break
		}

	}

	if sS.processThresholds(ctx, matchSQs, args.APIOpts) != nil || sS.processEEs(ctx, matchSQs, args.APIOpts) != nil || withErrors {
		err = utils.ErrPartiallyExecuted
	}

	var promIDs []string
	if promIDs, err = GetStringSliceOpts(ctx, tnt, evNm, sS.fltrS, sS.cfg.StatSCfg().Opts.PrometheusStatIDs,
		[]string{}, utils.OptsPrometheusStatIDs); err != nil {
		return
	}
	if len(promIDs) != 0 {
		if err = exportToPrometheus(matchSQs, utils.NewStringSet(promIDs)); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Failed to export the queues to Prometheus: error: %s",
					err.Error()))
			err = utils.ErrPartiallyExecuted
		}
	}
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	if ids, err = sS.processEvent(ctx, tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetStatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatS) V1GetStatQueuesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	var sqIDs []string
	dP := args.AsDataProvider()
	if sqIDs, err = GetStringSliceOpts(ctx, tnt, dP, sS.fltrS, sS.cfg.StatSCfg().Opts.ProfileIDs,
		config.StatsProfileIDsDftOpt, utils.OptsStatsProfileIDs); err != nil {
		return
	}
	evDp := args.AsDataProvider()
	var ignFilters bool
	if ignFilters, err = GetBoolOpts(ctx, tnt, evDp, sS.fltrS, sS.cfg.StatSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var sQs StatQueues
	if sQs, err = sS.matchingStatQueuesForEvent(ctx, tnt, sqIDs, evDp, ignFilters); err != nil {
		return
	}

	*reply = sQs.IDs()
	sQs.unlock()
	return
}

// V1GetStatQueue returns a StatQueue object
func (sS *StatS) V1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *StatQueue) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		statQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		return err
	}
	*reply = *sq
	return
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (sS *StatS) V1GetQueueStringMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		statQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	var rnd int
	if rnd, err = GetIntOpts(ctx, tnt, MapEvent{utils.Tenant: tnt, "*opts": map[string]any{}}, sS.fltrS,
		sS.cfg.StatSCfg().Opts.RoundingDecimals,
		utils.OptsRoundingDecimals); err != nil {
		return
	}
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue(rnd)
	}
	*reply = metrics
	return
}

// V1GetQueueFloatMetrics returns the metrics as float64 values
func (sS *StatS) V1GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		statQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]float64, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		val := metric.GetValue()
		metrics[metricID] = -1
		if val != utils.DecimalNaN {
			metrics[metricID], _ = val.Float64()
		}
	}
	*reply = metrics
	return
}

// V1GetQueueDecimalMetrics returns the metrics as decimal values
func (sS *StatS) V1GetQueueDecimalMetrics(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *map[string]*utils.Decimal) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		statQueueLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	sq, err := sS.getStatQueue(ctx, tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]*utils.Decimal, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetValue()
	}
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatS) V1GetQueueIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, qIDs *[]string) (err error) {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = sS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueuePrefix + tenant + utils.ConcatenatedKeySep
	keys, err := sS.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*qIDs = retIDs
	return
}

// V1ResetStatQueue resets the stat queue
func (sS *StatS) V1ResetStatQueue(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, rply *string) (err error) {
	if missing := utils.MissingStructFields(tntID, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cfg.GeneralCfg().DefaultTenant
	}
	// make sure statQueue is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		statQueueLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	var sq *StatQueue
	if sq, err = sS.dm.GetStatQueue(ctx, tnt, tntID.ID,
		true, true, utils.NonTransactional); err != nil {
		return
	}
	sq.SQItems = make([]SQItem, 0)
	metrics := sq.SQMetrics
	sq.SQMetrics = make(map[string]StatMetric)
	for id, m := range metrics {
		var metric StatMetric
		if metric, err = NewStatMetric(id,
			m.GetMinItems(), m.GetFilterIDs()); err != nil {
			return
		}
		sq.SQMetrics[id] = metric
	}
	sq.dirty = utils.BoolPointer(true)
	sS.storeStatQueue(ctx, sq)
	*rply = utils.OK
	return
}
