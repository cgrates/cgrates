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
	"github.com/cgrates/cgrates/utils"
)

// ThresholdProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdProfileWithAPIOpts struct {
	*ThresholdProfile
	APIOpts map[string]interface{}
}

// ThresholdProfile the profile for threshold
type ThresholdProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Time when this limit becomes active and expires
	MaxHits            int
	MinHits            int
	MinSleep           time.Duration
	Blocker            bool    // blocker flag to stop processing on filters matched
	Weight             float64 // Weight to sort the thresholds
	ActionIDs          []string
	Async              bool
}

// TenantID returns the concatenated key beteen tenant and ID
func (tp *ThresholdProfile) TenantID() string {
	return utils.ConcatenatedKey(tp.Tenant, tp.ID)
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

	tPrfl *ThresholdProfile
	dirty *bool // needs save
}

// TenantID returns the concatenated key beteen tenant and ID
func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

// ProcessEvent processes an ThresholdEvent
// concurrentActions limits the number of simultaneous action sets executed
func (t *Threshold) ProcessEvent(args *ThresholdsArgsProcessEvent, dm *DataManager) (err error) {
	if t.Snooze.After(time.Now()) || // snoozed, not executing actions
		t.Hits < t.tPrfl.MinHits || // number of hits was not met, will not execute actions
		(t.tPrfl.MaxHits != -1 &&
			t.Hits > t.tPrfl.MaxHits) {
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

	// for _, actionSetID := range t.tPrfl.ActionIDs {
	// at := &ActionTiming{
	// 	Uuid:      utils.GenUUID(),
	// 	ActionsID: actionSetID,
	// 	ExtraData: args.CGREvent,
	// }
	// if tntAcnt != utils.EmptyString {
	// 	at.accountIDs = utils.NewStringMap(tntAcnt)
	// }
	// if t.tPrfl.Async {
	// 	go func() {
	// 		if errExec := at.Execute(nil, nil); errExec != nil {
	// 			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
	// 		}
	// 	}()
	// } else if errExec := at.Execute(nil, nil); errExec != nil {
	// 	utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
	// 	err = utils.ErrPartiallyExecuted
	// }
	// }
	return
}

// Thresholds is a sortable slice of Threshold
type Thresholds []*Threshold

// Sort sorts based on Weight
func (ts Thresholds) Sort() {
	sort.Slice(ts, func(i, j int) bool { return ts[i].tPrfl.Weight > ts[j].tPrfl.Weight })
}

// NewThresholdService the constructor for ThresoldS service
func NewThresholdService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS) (tS *ThresholdService) {
	return &ThresholdService{dm: dm,
		cgrcfg:      cgrcfg,
		filterS:     filterS,
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
	stopBackup  chan struct{}
	loopStoped  chan struct{}
	storedTdIDs utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	stMux       sync.RWMutex    // protects storedTdIDs
}

// Shutdown is called to shutdown the service
func (tS *ThresholdService) Shutdown() {
	utils.Logger.Info("<ThresholdS> shutdown initialized")
	close(tS.stopBackup)
	tS.storeThresholds()
	utils.Logger.Info("<ThresholdS> shutdown complete")
}

// backup will regularly store resources changed to dataDB
func (tS *ThresholdService) runBackup() {
	storeInterval := tS.cgrcfg.ThresholdSCfg().StoreInterval
	if storeInterval <= 0 {
		tS.loopStoped <- struct{}{}
		return
	}
	for {
		tS.storeThresholds()
		select {
		case <-tS.stopBackup:
			tS.loopStoped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeThresholds represents one task of complete backup
func (tS *ThresholdService) storeThresholds() {
	var failedTdIDs []string
	for { // don't stop until we store all dirty resources
		tS.stMux.Lock()
		tID := tS.storedTdIDs.GetOne()
		if tID != "" {
			tS.storedTdIDs.Remove(tID)
		}
		tS.stMux.Unlock()
		if tID == "" {
			break // no more keys, backup completed
		}
		if tIf, ok := Cache.Get(utils.CacheThresholds, tID); !ok || tIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed retrieving from cache treshold with ID: %s", tID))
		} else if err := tS.StoreThreshold(tIf.(*Threshold)); err != nil {
			failedTdIDs = append(failedTdIDs, tID) // record failure so we can schedule it for next backup
		}
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
func (tS *ThresholdService) StoreThreshold(t *Threshold) (err error) {
	if t.dirty == nil || !*t.dirty {
		return
	}
	if err = tS.dm.SetThreshold(t, 0, true); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ThresholdS> failed saving Threshold with tenant: %s and ID: %s, error: %s",
				t.Tenant, t.ID, err.Error()))
		return
	}
	*t.dirty = false
	return
}

// matchingThresholdsForEvent returns ordered list of matching thresholds which are active for an Event
func (tS *ThresholdService) matchingThresholdsForEvent(tnt string, args *ThresholdsArgsProcessEvent) (ts Thresholds, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	tIDs := utils.NewStringSet(args.ThresholdIDs)
	if len(tIDs) == 0 {
		tIDs, err = MatchingItemIDsForEvent(context.TODO(), evNm,
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
		tPrfl, err := tS.dm.GetThresholdProfile(tnt, tID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if tPrfl.ActivationInterval != nil && args.Time != nil &&
			!tPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			continue
		}
		if pass, err := tS.filterS.Pass(context.TODO(), tnt, tPrfl.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		t, err := tS.dm.GetThreshold(tPrfl.Tenant, tPrfl.ID, true, true, "")
		if err != nil {
			if err == utils.ErrNotFound { // corner case where the threshold was removed due to MaxHits
				continue
			}
			return nil, err
		}
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
		if t.tPrfl.Blocker { // blocker will stop processing
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
		thIDs = make([]string, len(attr.ThresholdIDs))
		for i, id := range attr.ThresholdIDs {
			thIDs[i] = id
		}
	}
	return &ThresholdsArgsProcessEvent{
		ThresholdIDs: thIDs,
		CGREvent:     attr.CGREvent.Clone(),
	}
}

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdService) processEvent(tnt string, args *ThresholdsArgsProcessEvent) (thresholdsIDs []string, err error) {
	matchTS, err := tS.matchingThresholdsForEvent(tnt, args)
	if err != nil {
		return nil, err
	}
	var withErrors bool
	var tIDs []string
	for _, t := range matchTS {
		tIDs = append(tIDs, t.ID)
		t.Hits++
		err = t.ProcessEvent(args, tS.dm)
		if err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> threshold: %s, ignoring event: %s, error: %s",
					t.TenantID(), utils.ConcatenatedKey(tnt, args.CGREvent.ID), err.Error()))
			withErrors = true
			continue
		}
		if t.dirty == nil || t.Hits == t.tPrfl.MaxHits { // one time threshold
			if err = tS.dm.RemoveThreshold(t.Tenant, t.ID, utils.NonTransactional); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> failed removing from database non-recurrent threshold: %s, error: %s",
						t.TenantID(), err.Error()))
				withErrors = true
			}
			//since we don't handle in DataManager caching we do a manual remove here
			if err = tS.dm.CacheDataFromDB(context.TODO(), utils.ThresholdPrefix, []string{t.TenantID()}, true); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> failed removing from cache non-recurrent threshold: %s, error: %s",
						t.TenantID(), err.Error()))
				withErrors = true
			}
			continue
		}
		t.Snooze = time.Now().Add(t.tPrfl.MinSleep)
		// recurrent threshold
		*t.dirty = true // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			tS.StoreThreshold(t)
		} else {
			tS.stMux.Lock()
			tS.storedTdIDs.Add(t.TenantID())
			tS.stMux.Unlock()
		}
	}
	if len(tIDs) == 0 {
		return nil, utils.ErrNotFound
	}
	thresholdsIDs = append(thresholdsIDs, tIDs...)
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements ThresholdService method for processing an Event
func (tS *ThresholdService) V1ProcessEvent(args *ThresholdsArgsProcessEvent, reply *[]string) (err error) {
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
	if ids, err = tS.processEvent(tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (tS *ThresholdService) V1GetThresholdsForEvent(args *ThresholdsArgsProcessEvent, reply *Thresholds) (err error) {
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
	if ts, err = tS.matchingThresholdsForEvent(tnt, args); err == nil {
		*reply = ts
	}
	return
}

// V1GetThresholdIDs returns list of thresholdIDs configured for a tenant
func (tS *ThresholdService) V1GetThresholdIDs(tenant string, tIDs *[]string) (err error) {
	if tenant == utils.EmptyString {
		tenant = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdPrefix + tenant + utils.ConcatenatedKeySep
	keys, err := tS.dm.DataDB().GetKeysForPrefix(context.TODO(), prfx)
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
func (tS *ThresholdService) V1GetThreshold(tntID *utils.TenantID, t *Threshold) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	if thd, err = tS.dm.GetThreshold(tnt, tntID.ID, true, true, ""); err != nil {
		return
	}
	*t = *thd
	return
}

// Reload stops the backupLoop and restarts it
func (tS *ThresholdService) Reload() {
	close(tS.stopBackup)
	<-tS.loopStoped // wait until the loop is done
	tS.stopBackup = make(chan struct{})
	go tS.runBackup()
}

// StartLoop starts the gorutine with the backup loop
func (tS *ThresholdService) StartLoop() {
	go tS.runBackup()
}

// V1ResetThreshold resets the threshold hits
func (tS *ThresholdService) V1ResetThreshold(tntID *utils.TenantID, rply *string) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	if thd, err = tS.dm.GetThreshold(tnt, tntID.ID, true, true, ""); err != nil {
		return
	}
	if thd.Hits != 0 {
		thd.Hits = 0
		thd.Snooze = time.Time{}
		thd.dirty = utils.BoolPointer(true) // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			if err = tS.StoreThreshold(thd); err != nil {
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
