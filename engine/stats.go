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
	filterS *FilterS, connMgr *ConnManager) *StatService {
	return &StatService{
		dm:               dm,
		connMgr:          connMgr,
		filterS:          filterS,
		cgrcfg:           cgrcfg,
		storedStatQueues: make(utils.StringSet),
		loopStopped:      make(chan struct{}),
		stopBackup:       make(chan struct{}),
	}
}

// StatService builds stats for events
type StatService struct {
	dm               *DataManager
	connMgr          *ConnManager
	filterS          *FilterS
	cgrcfg           *config.CGRConfig
	loopStopped      chan struct{}
	stopBackup       chan struct{}
	storedStatQueues utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	ssqMux           sync.RWMutex    // protects storedStatQueues
}

// Reload stops the backupLoop and restarts it
func (sS *StatService) Reload(ctx *context.Context) {
	close(sS.stopBackup)
	<-sS.loopStopped // wait until the loop is done
	sS.stopBackup = make(chan struct{})
	go sS.runBackup(ctx)
}

// StartLoop starsS the gorutine with the backup loop
func (sS *StatService) StartLoop(ctx *context.Context) {
	go sS.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (sS *StatService) Shutdown(ctx *context.Context) {
	utils.Logger.Info("<StatS> service shutdown initialized")
	close(sS.stopBackup)
	sS.storeStats(ctx)
	utils.Logger.Info("<StatS> service shutdown complete")
}

// runBackup will regularly store statQueues changed to dataDB
func (sS *StatService) runBackup(ctx *context.Context) {
	storeInterval := sS.cgrcfg.StatSCfg().StoreInterval
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
func (sS *StatService) storeStats(ctx *context.Context) {
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
func (sS *StatService) StoreStatQueue(ctx *context.Context, sq *StatQueue) (err error) {
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
func (sS *StatService) matchingStatQueuesForEvent(ctx *context.Context, tnt string, statsIDs []string, evNm utils.MapStorage) (sqs StatQueues, err error) {
	sqIDs := utils.NewStringSet(statsIDs)
	if len(sqIDs) == 0 {
		sqIDs, err = MatchingItemIDsForEvent(ctx, evNm,
			sS.cgrcfg.StatSCfg().StringIndexedFields,
			sS.cgrcfg.StatSCfg().PrefixIndexedFields,
			sS.cgrcfg.StatSCfg().SuffixIndexedFields,
			sS.dm, utils.CacheStatFilterIndexes, tnt,
			sS.cgrcfg.StatSCfg().IndexedSelects,
			sS.cgrcfg.StatSCfg().NestedFields,
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
		var pass bool
		if pass, err = sS.filterS.Pass(ctx, tnt, sqPrfl.FilterIDs,
			evNm); err != nil {
			sqPrfl.unlock()
			sqs.unlock()
			return nil, err
		} else if !pass {
			sqPrfl.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			statQueueLockKey(sqPrfl.Tenant, sqPrfl.ID))
		var sq *StatQueue
		if sq, err = sS.dm.GetStatQueue(ctx, sqPrfl.Tenant, sqPrfl.ID, true, true, ""); err != nil {
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
		sq.sqPrfl = sqPrfl
		sqs = append(sqs, sq)
	}
	if len(sqs) == 0 {
		return nil, utils.ErrNotFound
	}
	// All good, convert from Map to Slice so we can sort
	sqs.Sort()
	for i, s := range sqs {
		if s.sqPrfl.Blocker && i != len(sqs)-1 { // blocker will stop processing and we are not at last index
			StatQueues(sqs[i+1:]).unlock()
			sqs = sqs[:i+1]
			break
		}
	}
	return
}

// StatsArgsProcessEvent the arguments for processing the event with stats
type StatsArgsProcessEvent struct {
	*utils.CGREvent
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *StatsArgsProcessEvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *StatsArgsProcessEvent) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *StatsArgsProcessEvent) Clone() *StatsArgsProcessEvent {
	return &StatsArgsProcessEvent{
		CGREvent: attr.CGREvent.Clone(),
	}
}

func (sS *StatService) getStatQueue(ctx *context.Context, tnt, id string) (sq *StatQueue, err error) {
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
func (sS *StatService) storeStatQueue(ctx *context.Context, sq *StatQueue) {
	if sS.cgrcfg.StatSCfg().StoreInterval != 0 && sq.dirty != nil { // don't save
		*sq.dirty = true // mark it to be saved
		if sS.cgrcfg.StatSCfg().StoreInterval == -1 {
			sS.StoreStatQueue(ctx, sq)
		} else {
			sS.ssqMux.Lock()
			sS.storedStatQueues.Add(sq.TenantID())
			sS.ssqMux.Unlock()
		}
	}
}

// processThresholds will pass the event for statQueue to ThresholdS
func (sS *StatService) processThresholds(ctx *context.Context, sQs StatQueues, opts map[string]interface{}) (err error) {
	if len(sS.cgrcfg.StatSCfg().ThresholdSConns) == 0 {
		return
	}
	if opts == nil {
		opts = make(map[string]interface{})
	}
	opts[utils.MetaEventType] = utils.StatUpdate
	var withErrs bool
	for _, sq := range sQs {
		if len(sq.sqPrfl.ThresholdIDs) != 0 {
			if len(sq.sqPrfl.ThresholdIDs) == 1 &&
				sq.sqPrfl.ThresholdIDs[0] == utils.MetaNone {
				continue
			}
			opts[utils.OptsThresholdsThresholdIDs] = sq.sqPrfl.ThresholdIDs
		}
		thEv := &ThresholdsArgsProcessEvent{
			CGREvent: &utils.CGREvent{
				Tenant: sq.Tenant,
				ID:     utils.GenUUID(),
				Event: map[string]interface{}{
					utils.EventType: utils.StatUpdate,
					utils.StatID:    sq.ID,
				},
				APIOpts: opts,
			},
		}
		for metricID, metric := range sq.SQMetrics {
			thEv.Event[metricID] = metric.GetValue(sS.cgrcfg.GeneralCfg().RoundingDecimals)
		}
		var tIDs []string
		if err := sS.connMgr.Call(ctx, sS.cgrcfg.StatSCfg().ThresholdSConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(opts[utils.OptsThresholdsThresholdIDs].([]string)) != 0 || err.Error() != utils.ErrNotFound.Error()) {
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

// processEvent processes a new event, dispatching to matching queues
// queues matching are also cached to speed up
func (sS *StatService) processEvent(ctx *context.Context, tnt string, args *StatsArgsProcessEvent) (statQueueIDs []string, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	var sqIDs []string
	if sqIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsStatsStatIDs); err != nil {
		return
	}
	matchSQs, err := sS.matchingStatQueuesForEvent(ctx, tnt, sqIDs, evNm)
	if err != nil {
		return nil, err
	}

	statQueueIDs = matchSQs.IDs()
	var withErrors bool
	for _, sq := range matchSQs {
		if err = sq.ProcessEvent(ctx, tnt, args.ID, sS.filterS, evNm); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Queue: %s, ignoring event: %s, error: %s",
					sq.TenantID(), utils.ConcatenatedKey(tnt, args.ID), err.Error()))
			withErrors = true
		}
		sS.storeStatQueue(ctx, sq)

	}
	if sS.processThresholds(ctx, matchSQs, args.APIOpts) != nil ||
		withErrors {
		err = utils.ErrPartiallyExecuted
	}
	matchSQs.unlock()
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatService) V1ProcessEvent(ctx *context.Context, args *StatsArgsProcessEvent, reply *[]string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	if ids, err = sS.processEvent(ctx, tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetStatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatService) V1GetStatQueuesForEvent(ctx *context.Context, args *StatsArgsProcessEvent, reply *[]string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var sqIDs []string
	if sqIDs, err = utils.OptAsStringSlice(args.APIOpts, utils.OptsStatsStatIDs); err != nil {
		return
	}
	var sQs StatQueues
	if sQs, err = sS.matchingStatQueuesForEvent(ctx, tnt, sqIDs, utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}); err != nil {
		return
	}

	*reply = sQs.IDs()
	sQs.unlock()
	return
}

// V1GetStatQueue returns a StatQueue object
func (sS *StatService) V1GetStatQueue(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *StatQueue) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
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
func (sS *StatService) V1GetQueueStringMetrics(ctx *context.Context, args *utils.TenantID, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
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
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue(sS.cgrcfg.GeneralCfg().RoundingDecimals)
	}
	*reply = metrics
	return
}

// V1GetQueueFloatMetrics returns the metrics as float64 values
func (sS *StatService) V1GetQueueFloatMetrics(ctx *context.Context, args *utils.TenantID, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
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
		metrics[metricID] = metric.GetFloat64Value(sS.cgrcfg.GeneralCfg().RoundingDecimals)
	}
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatService) V1GetQueueIDs(ctx *context.Context, tenant string, qIDs *[]string) (err error) {
	if tenant == utils.EmptyString {
		tenant = sS.cgrcfg.GeneralCfg().DefaultTenant
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
func (sS *StatService) V1ResetStatQueue(ctx *context.Context, tntID *utils.TenantID, rply *string) (err error) {
	if missing := utils.MissingStructFields(tntID, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
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
