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
	filterS *FilterS, connMgr *ConnManager) (ss *StatService) {

	return &StatService{
		dm:               dm,
		connMgr:          connMgr,
		filterS:          filterS,
		cgrcfg:           cgrcfg,
		storedStatQueues: make(utils.StringSet),
		loopStoped:       make(chan struct{}),
		stopBackup:       make(chan struct{}),
	}
}

// StatService builds stats for events
type StatService struct {
	dm               *DataManager
	connMgr          *ConnManager
	filterS          *FilterS
	cgrcfg           *config.CGRConfig
	loopStoped       chan struct{}
	stopBackup       chan struct{}
	storedStatQueues utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	ssqMux           sync.RWMutex    // protects storedStatQueues
}

// Shutdown is called to shutdown the service
func (sS *StatService) Shutdown() {
	utils.Logger.Info("<StatS> service shutdown initialized")
	close(sS.stopBackup)
	sS.storeStats()
	utils.Logger.Info("<StatS> service shutdown complete")
}

// runBackup will regularly store resources changed to dataDB
func (sS *StatService) runBackup() {
	storeInterval := sS.cgrcfg.StatSCfg().StoreInterval
	if storeInterval <= 0 {
		sS.loopStoped <- struct{}{}
		return
	}
	for {
		sS.storeStats()
		select {
		case <-sS.stopBackup:
			sS.loopStoped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeResources represents one task of complete backup
func (sS *StatService) storeStats() {
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
		guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) (gRes interface{}, gErr error) {
			if sqIf, ok := Cache.Get(utils.CacheStatQueues, sID); !ok || sqIf == nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> failed retrieving from cache stat queue with ID: %s",
						utils.StatService, sID))
			} else if err := sS.StoreStatQueue(sqIf.(*StatQueue)); err != nil {
				failedSqIDs = append(failedSqIDs, sID) // record failure so we can schedule it for next backup
			}
			return
		}, sS.cgrcfg.GeneralCfg().LockingTimeout, utils.StatQueuePrefix+sID)
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
func (sS *StatService) StoreStatQueue(sq *StatQueue) (err error) {
	if sq.dirty == nil || !*sq.dirty {
		return
	}
	if err = sS.dm.SetStatQueue(sq, nil, 0, nil, 0, true); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed saving StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if err = sS.dm.CacheDataFromDB(context.TODO(), utils.StatQueuePrefix, []string{sq.TenantID()}, true); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed caching StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	*sq.dirty = false
	return
}

// matchingStatQueuesForEvent returns ordered list of matching resources which are active by the time of the call
func (sS *StatService) matchingStatQueuesForEvent(tnt string, statsIDs []string, actTime *time.Time, evNm utils.MapStorage) (sqs StatQueues, err error) {
	sqIDs := utils.NewStringSet(statsIDs)
	if len(sqIDs) == 0 {
		sqIDs, err = MatchingItemIDsForEvent(context.TODO(), evNm,
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
		sqPrfl, err := sS.dm.GetStatQueueProfile(tnt, sqID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if sqPrfl.ActivationInterval != nil && actTime != nil &&
			!sqPrfl.ActivationInterval.IsActiveAtTime(*actTime) { // not active
			continue
		}
		if pass, err := sS.filterS.Pass(context.TODO(), tnt, sqPrfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		var sq *StatQueue
		lkID := utils.StatQueuePrefix + utils.ConcatenatedKey(sqPrfl.Tenant, sqPrfl.ID)
		guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) (gRes interface{}, gErr error) {
			sq, err = sS.dm.GetStatQueue(sqPrfl.Tenant, sqPrfl.ID, true, true, "")
			return
		}, sS.cgrcfg.GeneralCfg().LockingTimeout, lkID)
		if err != nil {
			return nil, err
		}
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
		if s.sqPrfl.Blocker { // blocker will stop processing
			sqs = sqs[:i+1]
			break
		}
	}
	return
}

// StatsArgsProcessEvent the arguments for processing the event with stats
type StatsArgsProcessEvent struct {
	StatIDs []string
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
	var statsIDs []string
	if attr.StatIDs != nil {
		statsIDs = make([]string, len(attr.StatIDs))
		for i, id := range attr.StatIDs {
			statsIDs[i] = id
		}
	}
	return &StatsArgsProcessEvent{
		StatIDs:  statsIDs,
		CGREvent: attr.CGREvent.Clone(),
	}
}

func (sS *StatService) getStatQueue(tnt, id string) (sq *StatQueue, err error) {
	if sq, err = sS.dm.GetStatQueue(tnt, id, true, true, utils.EmptyString); err != nil {
		return
	}
	lkID := utils.StatQueuePrefix + sq.TenantID()
	var removed int
	guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) (gRes interface{}, gErr error) {
		removed, err = sq.remExpired()
		return
	}, sS.cgrcfg.GeneralCfg().LockingTimeout, lkID)
	if err != nil || removed == 0 {
		return
	}
	sS.storeStatQueue(sq)
	return
}

// storeStatQueue will store the sq if needed
func (sS *StatService) storeStatQueue(sq *StatQueue) {
	if sS.cgrcfg.StatSCfg().StoreInterval != 0 && sq.dirty != nil { // don't save
		*sq.dirty = true // mark it to be saved
		if sS.cgrcfg.StatSCfg().StoreInterval == -1 {
			sS.StoreStatQueue(sq)
		} else {
			sS.ssqMux.Lock()
			sS.storedStatQueues.Add(sq.TenantID())
			sS.ssqMux.Unlock()
		}
	}
}

// processEvent processes a new event, dispatching to matching queues
// queues matching are also cached to speed up
func (sS *StatService) processEvent(tnt string, args *StatsArgsProcessEvent) (statQueueIDs []string, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	matchSQs, err := sS.matchingStatQueuesForEvent(tnt, args.StatIDs, args.Time, evNm)
	if err != nil {
		return nil, err
	}
	if len(matchSQs) == 0 {
		return nil, utils.ErrNotFound
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]interface{})
	}
	args.APIOpts[utils.MetaEventType] = utils.StatUpdate
	var stsIDs []string
	var withErrors bool
	for _, sq := range matchSQs {
		stsIDs = append(stsIDs, sq.ID)
		lkID := utils.StatQueuePrefix + sq.TenantID()
		guardian.Guardian.Guard(context.TODO(), func(_ *context.Context) (gRes interface{}, gErr error) {
			err = sq.ProcessEvent(tnt, args.ID, sS.filterS, evNm)
			return
		}, sS.cgrcfg.GeneralCfg().LockingTimeout, lkID)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatS> Queue: %s, ignoring event: %s, error: %s",
					sq.TenantID(), utils.ConcatenatedKey(tnt, args.ID), err.Error()))
			withErrors = true
		}
		sS.storeStatQueue(sq)
		if len(sS.cgrcfg.StatSCfg().ThresholdSConns) != 0 {
			var thIDs []string
			if len(sq.sqPrfl.ThresholdIDs) != 0 {
				if len(sq.sqPrfl.ThresholdIDs) == 1 && sq.sqPrfl.ThresholdIDs[0] == utils.MetaNone {
					continue
				}
				thIDs = sq.sqPrfl.ThresholdIDs
			}
			thEv := &ThresholdsArgsProcessEvent{
				ThresholdIDs: thIDs,
				CGREvent: &utils.CGREvent{
					Tenant: sq.Tenant,
					ID:     utils.GenUUID(),
					Event: map[string]interface{}{
						utils.EventType: utils.StatUpdate,
						utils.StatID:    sq.ID,
					},
					APIOpts: args.APIOpts,
				},
			}
			for metricID, metric := range sq.SQMetrics {
				thEv.Event[metricID] = metric.GetValue(sS.cgrcfg.GeneralCfg().RoundingDecimals)
			}
			var tIDs []string
			if err := sS.connMgr.Call(context.TODO(), sS.cgrcfg.StatSCfg().ThresholdSConns,
				utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<StatS> error: %s processing event %+v with ThresholdS.", err.Error(), thEv))
				withErrors = true
			}
		}
	}
	if len(stsIDs) != 0 {
		statQueueIDs = append(statQueueIDs, stsIDs...)
	} else {
		statQueueIDs = []string{}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatService) V1ProcessEvent(args *StatsArgsProcessEvent, reply *[]string) (err error) {
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
	if ids, err = sS.processEvent(tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetStatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatService) V1GetStatQueuesForEvent(args *StatsArgsProcessEvent, reply *[]string) (err error) {
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
	var sQs StatQueues
	if sQs, err = sS.matchingStatQueuesForEvent(tnt, args.StatIDs, args.Time, utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}); err != nil {
		return
	}
	ids := make([]string, len(sQs))
	for i, sq := range sQs {
		ids[i] = sq.ID
	}
	*reply = ids
	return
}

// V1GetStatQueue returns a StatQueue object
func (sS *StatService) V1GetStatQueue(args *utils.TenantIDWithAPIOpts, reply *StatQueue) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	sq, err := sS.getStatQueue(tnt, args.ID)
	if err != nil {
		return err
	}
	*reply = *sq
	return
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (sS *StatService) V1GetQueueStringMetrics(args *utils.TenantID, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	sq, err := sS.getStatQueue(tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	sq.RLock()
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue(sS.cgrcfg.GeneralCfg().RoundingDecimals)
	}
	sq.RUnlock()
	*reply = metrics
	return
}

// V1GetQueueFloatMetrics returns the metrics as float64 values
func (sS *StatService) V1GetQueueFloatMetrics(args *utils.TenantID, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	sq, err := sS.getStatQueue(tnt, args.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	sq.RLock()
	metrics := make(map[string]float64, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetFloat64Value(sS.cgrcfg.GeneralCfg().RoundingDecimals)
	}
	sq.RUnlock()
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatService) V1GetQueueIDs(tenant string, qIDs *[]string) (err error) {
	if tenant == utils.EmptyString {
		tenant = sS.cgrcfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueuePrefix + tenant + utils.ConcatenatedKeySep
	keys, err := sS.dm.DataDB().GetKeysForPrefix(context.TODO(), prfx)
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

// Reload stops the backupLoop and restarts it
func (sS *StatService) Reload() {
	close(sS.stopBackup)
	<-sS.loopStoped // wait until the loop is done
	sS.stopBackup = make(chan struct{})
	go sS.runBackup()
}

// StartLoop starsS the gorutine with the backup loop
func (sS *StatService) StartLoop() {
	go sS.runBackup()
}

// V1ResetStatQueue resets the stat queue
func (sS *StatService) V1ResetStatQueue(tntID *utils.TenantID, rply *string) (err error) {
	var sq *StatQueue
	if sq, err = sS.dm.GetStatQueue(tntID.Tenant, tntID.ID,
		true, true, utils.NonTransactional); err != nil {
		return
	}
	sq.Lock()
	defer sq.Unlock()
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
	sS.storeStatQueue(sq)
	*rply = utils.OK
	return
}
