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
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/cgrates/cgrates/cache"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewStatService initializes a StatService
func NewStatService(dm *DataManager, storeInterval time.Duration) (ss *StatService, err error) {
	return &StatService{dm: dm, storeInterval: storeInterval,
		stopBackup: make(chan struct{})}, nil
}

// StatService builds stats for events
type StatService struct {
	dm               *DataManager
	storeInterval    time.Duration
	stopBackup       chan struct{}
	storedStatQueues utils.StringMap // keep a record of stats which need saving, map[statsTenantID]bool
	ssqMux           sync.RWMutex    // protects storedStatQueues
}

// ListenAndServe loops keeps the service alive
func (sS *StatService) ListenAndServe(exitChan chan bool) error {
	go sS.runBackup() // start backup loop
	e := <-exitChan
	exitChan <- e // put back for the others listening for shutdown request
	return nil
}

// Shutdown is called to shutdown the service
func (sS *StatService) Shutdown() error {
	utils.Logger.Info("<StatS> service shutdown initialized")
	close(sS.stopBackup)
	sS.storeStats()
	utils.Logger.Info("<StatS> service shutdown complete")
	return nil
}

// runBackup will regularly store resources changed to dataDB
func (sS *StatService) runBackup() {
	if sS.storeInterval <= 0 {
		return
	}
	for {
		select {
		case <-sS.stopBackup:
			return
		default:
		}
		sS.storeStats()
		time.Sleep(sS.storeInterval)
	}
}

// storeResources represents one task of complete backup
func (sS *StatService) storeStats() {
	var failedSqIDs []string
	for { // don't stop untill we store all dirty statQueues
		sS.ssqMux.Lock()
		sID := sS.storedStatQueues.GetOne()
		if sID != "" {
			delete(sS.storedStatQueues, sID)
		}
		sS.ssqMux.Unlock()
		if sID == "" {
			break // no more keys, backup completed
		}
		if sqIf, ok := cache.Get(utils.StatQueuePrefix + sID); !ok || sqIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<StatS> failed retrieving from cache stat queue with ID: %s", sID))
		} else if err := sS.StoreStatQueue(sqIf.(*StatQueue)); err != nil {
			failedSqIDs = append(failedSqIDs, sID) // record failure so we can schedule it for next backup
		}
		// randomize the CPU load and give up thread control
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Nanosecond)
	}
	if len(failedSqIDs) != 0 { // there were errors on save, schedule the keys for next backup
		sS.ssqMux.Lock()
		for _, sqID := range failedSqIDs {
			sS.storedStatQueues[sqID] = true
		}
		sS.ssqMux.Unlock()
	}
}

// StoreStatQueue stores the statQueue in DB and corrects dirty flag
func (sS *StatService) StoreStatQueue(sq *StatQueue) (err error) {
	if sq.dirty == nil || !*sq.dirty {
		return
	}
	if err = sS.dm.SetStatQueue(sq); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<StatS> failed saving StatQueue with ID: %s, error: %s",
				sq.TenantID(), err.Error()))
		return
	}
	*sq.dirty = false
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (sS *StatService) matchingStatQueuesForEvent(ev *StatEvent) (sqs StatQueues, err error) {
	if ev.Tenant == "" {
		return nil, errors.New("missing Tenant information")
	}
	matchingSQs := make(map[string]*StatQueue)
	sqIDs, err := matchingItemIDsForEvent(ev.Fields, sS.dm.DataDB(), utils.StatQueuesStringIndex+ev.Tenant)
	if err != nil {
		return nil, err
	}
	lockIDs := utils.PrefixSliceItems(sqIDs.Slice(), utils.StatQueuesStringIndex)
	guardian.Guardian.GuardIDs(config.CgrConfig().LockingTimeout, lockIDs...)
	defer guardian.Guardian.UnguardIDs(lockIDs...)
	for sqID := range sqIDs {
		sqPrfl, err := sS.dm.DataDB().GetStatQueueProfile(ev.Tenant, sqID, false, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if sqPrfl.ActivationInterval != nil &&
			!sqPrfl.ActivationInterval.IsActiveAtTime(time.Now()) { // not active
			continue
		}
		passAllFilters := true
		for _, fltr := range sqPrfl.Filters {
			if pass, err := fltr.Pass(ev.Fields, "", sS); err != nil {
				return nil, err
			} else if !pass {
				passAllFilters = false
				continue
			}
		}
		if !passAllFilters {
			continue
		}
		s, err := sS.dm.GetStatQueue(sqPrfl.Tenant, sqPrfl.ID, false, "")
		if err != nil {
			return nil, err
		}
		if sqPrfl.Stored && s.dirty == nil {
			s.dirty = utils.BoolPointer(false)
		}
		if sqPrfl.TTL >= 0 {
			s.ttl = utils.DurationPointer(sqPrfl.TTL)
		}
		s.sqPrfl = sqPrfl
		matchingSQs[sqPrfl.ID] = s
	}
	// All good, convert from Map to Slice so we can sort
	sqs = make(StatQueues, len(matchingSQs))
	i := 0
	for _, s := range matchingSQs {
		sqs[i] = s
		i++
	}
	sqs.Sort()
	for i, s := range sqs {
		if s.sqPrfl.Blocker { // blocker will stop processing
			sqs = sqs[:i+1]
			break
		}
	}
	return
}

// Call implements rpcclient.RpcClientConnection interface for internal RPC
// here for cases when passing StatsService as rpccclient.RpcClientConnection (ie. in ResourceS)
func (ss *StatService) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(ss).MethodByName(methodSplit[0][len(methodSplit[0])-2:] + methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}

// processEvent processes a new event, dispatching to matching queues
// queues matching are also cached to speed up
func (sS *StatService) processEvent(ev *StatEvent) (err error) {
	if missing := utils.MissingStructFields(ev, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	matchSQs, err := sS.matchingStatQueuesForEvent(ev)
	if err != nil {
		return err
	}
	if len(matchSQs) == 0 {
		return utils.ErrNotFound
	}
	var withErrors bool
	for _, sq := range matchSQs {
		if err = sq.ProcessEvent(ev); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<StatService> Queue: %s, ignoring event: %s, error: %s",
					sq.TenantID(), ev.TenantID(), err.Error()))
			withErrors = true
		}
		if sq.sqPrfl.Blocker {
			break
		}
	}
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements StatV1 method for processing an Event
func (sS *StatService) V1ProcessEvent(ev *StatEvent, reply *string) (err error) {
	if err = sS.processEvent(ev); err == nil {
		*reply = utils.OK
	}
	return
}

// V1StatQueuesForEvent implements StatV1 method for processing an Event
func (sS *StatService) V1GetStatQueuesForEvent(ev *StatEvent, reply *StatQueues) (err error) {
	var sQs StatQueues
	if sQs, err = sS.matchingStatQueuesForEvent(ev); err == nil {
		*reply = sQs
	}
	return
}

// V1GetQueueStringMetrics returns the metrics of a Queue as string values
func (sS *StatService) V1GetQueueStringMetrics(args *utils.TenantID, reply *map[string]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sq, err := sS.dm.GetStatQueue(args.Tenant, args.ID, false, "")
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]string, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetStringValue("")
	}
	*reply = metrics
	return
}

// V1GetFloatMetrics returns the metrics as float64 values
func (sS *StatService) V1GetQueueFloatMetrics(args *utils.TenantID, reply *map[string]float64) (err error) {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	sq, err := sS.dm.GetStatQueue(args.Tenant, args.ID, false, "")
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	metrics := make(map[string]float64, len(sq.SQMetrics))
	for metricID, metric := range sq.SQMetrics {
		metrics[metricID] = metric.GetFloat64Value()
	}
	*reply = metrics
	return
}

// V1GetQueueIDs returns list of queueIDs registered for a tenant
func (sS *StatService) V1GetQueueIDs(tenant string, qIDs *[]string) (err error) {
	prfx := utils.StatQueuePrefix + tenant + ":"
	keys, err := sS.dm.DataDB().GetKeysForPrefix(prfx)
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
