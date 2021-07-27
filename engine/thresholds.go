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
	"sort"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// ThresholdProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdProfileWithAPIOpts struct {
	*ThresholdProfile
	APIOpts map[string]interface{}
}

// ThresholdProfile the profile for threshold
type ThresholdProfile struct {
	Tenant           string
	ID               string
	FilterIDs        []string
	MaxHits          int
	MinHits          int
	MinSleep         time.Duration
	Blocker          bool    // blocker flag to stop processing on filters matched
	Weight           float64 // Weight to sort the thresholds
	ActionProfileIDs []string
	Async            bool

	lkID string // holds the reference towards guardian lock key
}

// TenantID returns the concatenated key beteen tenant and ID
func (tp *ThresholdProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
}

// thresholdProfileLockKey returns the ID used to lock a ThresholdProfile with guardian
func thresholdProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheThresholdProfiles, tnt, id)
}

// lock will lock the ThresholdProfile using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (tp *ThresholdProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdProfileLockKey(tp.Tenant, tp.ID))
	}
	tp.lkID = lkID
}

// unlock will unlock the ThresholdProfile and clear rp.lkID
func (tp *ThresholdProfile) unlock() {
	if tp.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(tp.lkID)
	tp.lkID = utils.EmptyString
}

// isLocked returns the locks status of this ThresholdProfile
func (tp *ThresholdProfile) isLocked() bool {
	return tp.lkID != utils.EmptyString
}

// ThresholdWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdWithAPIOpts struct {
	*Threshold
	APIOpts map[string]interface{}
}

// Threshold is the unit matched by filters
type Threshold struct {
	Tenant string
	ID     string
	Hits   int       // number of hits for this threshold
	Snooze time.Time // prevent threshold to run too early

	lkID  string // ID of the lock used when matching the threshold
	tPrfl *ThresholdProfile
	dirty *bool // needs save
}

// TenantID returns the concatenated key beteen tenant and ID
func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

// thresholdLockKey returns the ID used to lock a threshold with guardian
func thresholdLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheThresholds, tnt, id)
}

// lock will lock the threshold using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (t *Threshold) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdLockKey(t.Tenant, t.ID))
	}
	t.lkID = lkID
}

// unlock will unlock the threshold and clear r.lkID
func (t *Threshold) unlock() {
	if t.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(t.lkID)
	t.lkID = utils.EmptyString
}

// isLocked returns the locks status of this threshold
func (t *Threshold) isLocked() bool {
	return t.lkID != utils.EmptyString
}

// processEventWithThreshold processes an ThresholdEvent
func processEventWithThreshold(ctx *context.Context, connMgr *ConnManager, actionsConns []string, args *utils.CGREvent, t *Threshold) (err error) {
	if t.Snooze.After(time.Now()) || // snoozed, not executing actions
		t.Hits < t.tPrfl.MinHits || // number of hits was not met, will not execute actions
		(t.tPrfl.MaxHits != -1 &&
			t.Hits > t.tPrfl.MaxHits) ||
		(len(t.tPrfl.ActionProfileIDs) == 1 &&
			t.tPrfl.ActionProfileIDs[0] == utils.MetaNone) {
		return
	}

	// var tntAcnt string
	// var acnt string
	// if utils.IfaceAsString(args.APIOpts[utils.MetaEventType]) == utils.AccountUpdate {
	// 	acnt, _ = args.FieldAsString(utils.ID)
	// } else {
	// 	acnt, _ = args.FieldAsString(utils.AccountField)
	// }
	// if acnt != utils.EmptyString {
	// 	tntAcnt = utils.ConcatenatedKey(args.Tenant, acnt)
	// }

	var reply string
	if !t.tPrfl.Async {
		return connMgr.Call(ctx, actionsConns, utils.ActionSv1ExecuteActions, &utils.ArgActionSv1ScheduleActions{
			CGREvent:         args,
			ActionProfileIDs: t.tPrfl.ActionProfileIDs,
		}, &reply)
	}
	go func() {
		if errExec := connMgr.Call(context.Background(), actionsConns, utils.ActionSv1ExecuteActions, &utils.ArgActionSv1ScheduleActions{
			CGREvent:         args,
			ActionProfileIDs: t.tPrfl.ActionProfileIDs,
		}, &reply); errExec != nil {
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions for threshold: %s, error: %s", t.TenantID(), errExec.Error()))
		}
	}()
	return
}

// Thresholds is a sortable slice of Threshold
type Thresholds []*Threshold

// Sort sorts based on Weight
func (ts Thresholds) Sort() {
	sort.Slice(ts, func(i, j int) bool { return ts[i].tPrfl.Weight > ts[j].tPrfl.Weight })
}

// unlock will unlock thresholds part of this slice
func (ts Thresholds) unlock() {
	for _, t := range ts {
		t.unlock()
		if t.tPrfl != nil {
			t.tPrfl.unlock()
		}
	}
}

// NewThresholdService the constructor for ThresoldS service
func NewThresholdService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS, connMgr *ConnManager) (tS *ThresholdService) {
	return &ThresholdService{
		dm:          dm,
		cgrcfg:      cgrcfg,
		filterS:     filterS,
		connMgr:     connMgr,
		stopBackup:  make(chan struct{}),
		loopStoped:  make(chan struct{}),
		storedTdIDs: make(utils.StringSet),
	}
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm          *DataManager
	cgrcfg      *config.CGRConfig
	filterS     *FilterS
	connMgr     *ConnManager
	stopBackup  chan struct{}
	loopStoped  chan struct{}
	storedTdIDs utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	stMux       sync.RWMutex    // protects storedTdIDs
}

// Reload stops the backupLoop and restarts it
func (tS *ThresholdService) Reload(ctx *context.Context) {
	close(tS.stopBackup)
	<-tS.loopStoped // wait until the loop is done
	tS.stopBackup = make(chan struct{})
	go tS.runBackup(ctx)
}

// StartLoop starts the gorutine with the backup loop
func (tS *ThresholdService) StartLoop(ctx *context.Context) {
	go tS.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (tS *ThresholdService) Shutdown(ctx *context.Context) {
	utils.Logger.Info("<ThresholdS> shutdown initialized")
	close(tS.stopBackup)
	tS.storeThresholds(ctx)
	utils.Logger.Info("<ThresholdS> shutdown complete")
}

// backup will regularly store thresholds changed to dataDB
func (tS *ThresholdService) runBackup(ctx *context.Context) {
	storeInterval := tS.cgrcfg.ThresholdSCfg().StoreInterval
	if storeInterval <= 0 {
		tS.loopStoped <- struct{}{}
		return
	}
	for {
		tS.storeThresholds(ctx)
		select {
		case <-tS.stopBackup:
			tS.loopStoped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeThresholds represents one task of complete backup
func (tS *ThresholdService) storeThresholds(ctx *context.Context) {
	var failedTdIDs []string
	for { // don't stop until we store all dirty thresholds
		tS.stMux.Lock()
		tID := tS.storedTdIDs.GetOne()
		if tID != "" {
			tS.storedTdIDs.Remove(tID)
		}
		tS.stMux.Unlock()
		if tID == "" {
			break // no more keys, backup completed
		}
		tIf, ok := Cache.Get(utils.CacheThresholds, tID)
		if !ok || tIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed retrieving from cache treshold with ID: %s", tID))
			continue
		}
		t := tIf.(*Threshold)
		t.lock(utils.EmptyString)
		if err := tS.StoreThreshold(ctx, t); err != nil {
			failedTdIDs = append(failedTdIDs, tID) // record failure so we can schedule it for next backup
		}
		t.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedTdIDs) != 0 { // there were errors on save, schedule the keys for next backup
		tS.stMux.Lock()
		tS.storedTdIDs.AddSlice(failedTdIDs)
		tS.stMux.Unlock()
	}
}

// StoreThreshold stores the threshold in DB and corrects dirty flag
func (tS *ThresholdService) StoreThreshold(ctx *context.Context, t *Threshold) (err error) {
	if t.dirty == nil || !*t.dirty {
		return
	}
	if err = tS.dm.SetThreshold(ctx, t); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ThresholdS> failed saving Threshold with tenant: %s and ID: %s, error: %s",
				t.Tenant, t.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := t.TenantID(); Cache.HasItem(utils.CacheThresholds, tntID) { // only cache if previously there
		if err = Cache.Set(ctx, utils.CacheThresholds, tntID, t, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> failed caching Threshold with ID: %s, error: %s",
					t.TenantID(), err.Error()))
			return
		}
	}
	*t.dirty = false
	return
}

// matchingThresholdsForEvent returns ordered list of matching thresholds which are active for an Event
func (tS *ThresholdService) matchingThresholdsForEvent(ctx *context.Context, tnt string, args *ThresholdsArgsProcessEvent) (ts Thresholds, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	tIDs := utils.NewStringSet(args.ThresholdIDs)
	if len(tIDs) == 0 {
		tIDs, err = MatchingItemIDsForEvent(ctx, evNm,
			tS.cgrcfg.ThresholdSCfg().StringIndexedFields,
			tS.cgrcfg.ThresholdSCfg().PrefixIndexedFields,
			tS.cgrcfg.ThresholdSCfg().SuffixIndexedFields,
			tS.dm, utils.CacheThresholdFilterIndexes, tnt,
			tS.cgrcfg.ThresholdSCfg().IndexedSelects,
			tS.cgrcfg.ThresholdSCfg().NestedFields,
		)
		if err != nil {
			return nil, err
		}
	}
	ts = make(Thresholds, 0, len(tIDs))
	for tID := range tIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdProfileLockKey(tnt, tID))
		var tPrfl *ThresholdProfile
		if tPrfl, err = tS.dm.GetThresholdProfile(ctx, tnt, tID, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			ts.unlock()
			return nil, err
		}
		tPrfl.lock(lkPrflID)
		var pass bool
		if pass, err = tS.filterS.Pass(ctx, tnt, tPrfl.FilterIDs,
			evNm); err != nil {
			tPrfl.unlock()
			ts.unlock()
			return nil, err
		} else if !pass {
			tPrfl.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdLockKey(tPrfl.Tenant, tPrfl.ID))
		var t *Threshold
		if t, err = tS.dm.GetThreshold(ctx, tPrfl.Tenant, tPrfl.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			tPrfl.unlock()
			if err == utils.ErrNotFound { // corner case where the threshold was removed due to MaxHits
				err = nil
				continue
			}
			ts.unlock()
			return nil, err
		}
		t.lock(lkID)
		if t.dirty == nil || tPrfl.MaxHits == -1 || t.Hits < tPrfl.MaxHits {
			t.dirty = utils.BoolPointer(false)
		}
		t.tPrfl = tPrfl
		ts = append(ts, t)
	}
	// All good, convert from Map to Slice so we can sort
	if len(ts) == 0 {
		return nil, utils.ErrNotFound
	}
	ts.Sort()
	for i, t := range ts {
		if t.tPrfl.Blocker && i != len(ts)-1 { // blocker will stop processing and we are not at last index
			Thresholds(ts[i+1:]).unlock()
			ts = ts[:i+1]
			break
		}
	}
	return
}

// ThresholdsArgsProcessEvent are the arguments to proccess the event with thresholds
type ThresholdsArgsProcessEvent struct {
	ThresholdIDs []string
	*utils.CGREvent
	clnb bool //rpcclonable
}

// SetCloneable sets if the args should be clonned on internal connections
func (attr *ThresholdsArgsProcessEvent) SetCloneable(rpcCloneable bool) {
	attr.clnb = rpcCloneable
}

// RPCClone implements rpcclient.RPCCloner interface
func (attr *ThresholdsArgsProcessEvent) RPCClone() (interface{}, error) {
	if !attr.clnb {
		return attr, nil
	}
	return attr.Clone(), nil
}

// Clone creates a clone of the object
func (attr *ThresholdsArgsProcessEvent) Clone() *ThresholdsArgsProcessEvent {
	var thIDs []string
	if attr.ThresholdIDs != nil {
		thIDs = utils.CloneStringSlice(attr.ThresholdIDs)
	}
	return &ThresholdsArgsProcessEvent{
		ThresholdIDs: thIDs,
		CGREvent:     attr.CGREvent.Clone(),
	}
}

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdService) processEvent(ctx *context.Context, tnt string, args *ThresholdsArgsProcessEvent) (thresholdsIDs []string, err error) {
	var matchTs Thresholds
	if matchTs, err = tS.matchingThresholdsForEvent(ctx, tnt, args); err != nil {
		return nil, err
	}
	var withErrors bool
	thresholdsIDs = make([]string, 0, len(matchTs))
	for _, t := range matchTs {
		thresholdsIDs = append(thresholdsIDs, t.ID)
		t.Hits++
		if err = processEventWithThreshold(ctx, tS.connMgr,
			tS.cgrcfg.ThresholdSCfg().ActionSConns, args.CGREvent, t); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> threshold: %s, ignoring event: %s, error: %s",
					t.TenantID(), utils.ConcatenatedKey(tnt, args.CGREvent.ID), err.Error()))
			withErrors = true
			continue
		}
		if t.dirty == nil || t.Hits == t.tPrfl.MaxHits { // one time threshold
			if err = tS.dm.RemoveThreshold(ctx, t.Tenant, t.ID); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> failed removing from database non-recurrent threshold: %s, error: %s",
						t.TenantID(), err.Error()))
				withErrors = true
			}
			//since we don't handle in DataManager caching we do a manual remove here
			if tntID := t.TenantID(); Cache.HasItem(utils.CacheThresholds, tntID) { // only cache if previously there
				if err = Cache.Set(ctx, utils.CacheThresholds, tntID, nil, nil,
					true, utils.NonTransactional); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("<ThresholdService> failed removing from cache non-recurrent threshold: %s, error: %s",
							t.TenantID(), err.Error()))
					withErrors = true
				}
			}
			continue
		}
		t.Snooze = time.Now().Add(t.tPrfl.MinSleep)
		// recurrent threshold
		*t.dirty = true // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			tS.StoreThreshold(ctx, t)
		} else {
			tS.stMux.Lock()
			tS.storedTdIDs.Add(t.TenantID())
			tS.stMux.Unlock()
		}
	}
	matchTs.unlock()
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements ThresholdService method for processing an Event
func (tS *ThresholdService) V1ProcessEvent(ctx *context.Context, args *ThresholdsArgsProcessEvent, reply *[]string) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	if ids, err = tS.processEvent(ctx, tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (tS *ThresholdService) V1GetThresholdsForEvent(ctx *context.Context, args *ThresholdsArgsProcessEvent, reply *Thresholds) (err error) {
	if args.CGREvent == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.CGREvent.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var ts Thresholds
	if ts, err = tS.matchingThresholdsForEvent(ctx, tnt, args); err == nil {
		*reply = ts
		ts.unlock()
	}
	return
}

// V1GetThresholdIDs returns list of thresholdIDs configured for a tenant
func (tS *ThresholdService) V1GetThresholdIDs(ctx *context.Context, tenant string, tIDs *[]string) (err error) {
	if tenant == utils.EmptyString {
		tenant = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdPrefix + tenant + utils.ConcatenatedKeySep
	keys, err := tS.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*tIDs = retIDs
	return
}

// V1GetThreshold retrieves a Threshold
func (tS *ThresholdService) V1GetThreshold(ctx *context.Context, tntID *utils.TenantID, t *Threshold) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	// make sure threshold is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		thresholdLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	if thd, err = tS.dm.GetThreshold(ctx, tnt, tntID.ID, true, true, ""); err != nil {
		return
	}
	*t = *thd
	return
}

// V1ResetThreshold resets the threshold hits
func (tS *ThresholdService) V1ResetThreshold(ctx *context.Context, tntID *utils.TenantID, rply *string) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	// make sure threshold is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		thresholdLockKey(tnt, tntID.ID))
	defer guardian.Guardian.UnguardIDs(lkID)
	if thd, err = tS.dm.GetThreshold(ctx, tnt, tntID.ID, true, true, ""); err != nil {
		return
	}
	if thd.Hits != 0 {
		thd.Hits = 0
		thd.Snooze = time.Time{}
		thd.dirty = utils.BoolPointer(true) // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			if err = tS.StoreThreshold(ctx, thd); err != nil {
				return
			}
		} else {
			tS.stMux.Lock()
			tS.storedTdIDs.Add(thd.TenantID())
			tS.stMux.Unlock()
		}
	}
	*rply = utils.OK
	return
}
