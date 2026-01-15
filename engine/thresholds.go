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

package engine

import (
	"fmt"
	"maps"
	"runtime"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// ThresholdProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdProfileWithAPIOpts struct {
	*ThresholdProfile
	APIOpts map[string]any
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
	EeIDs              []string

	lkID string // holds the reference towards guardian lock key
}

// Clone clones *ThresholdProfile (lkID excluded)
func (tp *ThresholdProfile) Clone() *ThresholdProfile {
	if tp == nil {
		return nil
	}
	clone := &ThresholdProfile{
		Tenant:   tp.Tenant,
		ID:       tp.ID,
		MaxHits:  tp.MaxHits,
		MinHits:  tp.MinHits,
		MinSleep: tp.MinSleep,
		Blocker:  tp.Blocker,
		Weight:   tp.Weight,
		Async:    tp.Async,
	}
	if tp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(tp.FilterIDs))
		copy(clone.FilterIDs, tp.FilterIDs)
	}
	if tp.ActionIDs != nil {
		clone.ActionIDs = make([]string, len(tp.ActionIDs))
		copy(clone.ActionIDs, tp.ActionIDs)
	}
	if tp.ActivationInterval != nil {
		clone.ActivationInterval = tp.ActivationInterval.Clone()
	}
	if tp.EeIDs != nil {
		clone.EeIDs = make([]string, len(tp.EeIDs))
		copy(clone.EeIDs, tp.EeIDs)
	}
	return clone
}

// CacheClone returns a clone of ThresholdProfile used by ltcache CacheCloner
func (tp *ThresholdProfile) CacheClone() any {
	return tp.Clone()
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
	tmp := tp.lkID
	tp.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this ThresholdProfile
func (tp *ThresholdProfile) isLocked() bool {
	return tp.lkID != utils.EmptyString
}

// ThresholdWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdWithAPIOpts struct {
	*Threshold
	APIOpts map[string]any
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

// Clone clones *Threshold (lkID excluded)
func (t *Threshold) Clone() *Threshold {
	if t == nil {
		return nil
	}
	clone := &Threshold{
		Tenant: t.Tenant,
		ID:     t.ID,
		Hits:   t.Hits,
		Snooze: t.Snooze,
	}
	if t.tPrfl != nil {
		clone.tPrfl = t.tPrfl.Clone()
	}
	if t.dirty != nil {
		clone.dirty = new(bool)
		*clone.dirty = *t.dirty
	}
	return clone
}

// CacheClone returns a clone of Threshold used by ltcache CacheCloner
func (t *Threshold) CacheClone() any {
	return t.Clone()
}

// TenantID returns the concatenated key beteen tenant and ID
func (t *Threshold) TenantID() string {
	return utils.ConcatenatedKey(t.Tenant, t.ID)
}

type ThresholdConfig struct {
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval
	MaxHits            int
	MinHits            int
	MinSleep           time.Duration
	Blocker            bool
	Weight             float64
	ActionIDs          []string
	Async              bool
	EeIDs              []string
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
	tmp := t.lkID
	t.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this threshold
func (t *Threshold) isLocked() bool {
	return t.lkID != utils.EmptyString
}

// processEEs processes to the EEs for this threshold
func (t *ThresholdService) processEEs(opts map[string]any, th *Threshold) (err error) {
	var targetEeIDs []string
	if len(th.tPrfl.EeIDs) > 0 {
		targetEeIDs = th.tPrfl.EeIDs
		if isNone := slices.Contains(th.tPrfl.EeIDs, utils.MetaNone); isNone {
			targetEeIDs = []string{}
		}
	} else {
		targetEeIDs = t.cgrcfg.ThresholdSCfg().EEsExporterIDs
	}
	if len(targetEeIDs) > 0 {
		if len(t.cgrcfg.ThresholdSCfg().EEsConns) == 0 {
			return utils.NewErrNotConnected(utils.EEs)
		}
	} else {
		return nil // no EEs to process
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	sortedFilterIDs := make([]string, len(th.tPrfl.FilterIDs))
	copy(sortedFilterIDs, th.tPrfl.FilterIDs)
	slices.Sort(sortedFilterIDs)
	opts[utils.MetaEventType] = utils.ThresholdHit
	cgrEv := &utils.CGREvent{
		Tenant: th.Tenant,
		ID:     utils.GenUUID(),
		Time:   utils.TimePointer(time.Now()),
		Event: map[string]any{
			utils.EventType: utils.ThresholdHit,
			utils.ID:        th.ID,
			utils.Hits:      th.Hits,
			utils.Snooze:    th.Snooze,
			utils.ThresholdConfig: ThresholdConfig{
				FilterIDs:          sortedFilterIDs,
				ActivationInterval: th.tPrfl.ActivationInterval,
				MaxHits:            th.tPrfl.MaxHits,
				MinHits:            th.tPrfl.MinHits,
				MinSleep:           th.tPrfl.MinSleep,
				Blocker:            th.tPrfl.Blocker,
				Weight:             th.tPrfl.Weight,
				ActionIDs:          th.tPrfl.ActionIDs,
				Async:              th.tPrfl.Async,
				EeIDs:              th.tPrfl.EeIDs,
			},
		},
		APIOpts: opts,
	}
	cgrEventWithID := &CGREventWithEeIDs{
		CGREvent: cgrEv,
		EeIDs:    targetEeIDs,
	}
	var reply map[string]map[string]any
	if th.tPrfl.Async {
		go func() {
			if err := t.connMgr.Call(context.TODO(), t.cgrcfg.ThresholdSCfg().EEsConns,
				utils.EeSv1ProcessEvent,
				cgrEventWithID, &reply); err != nil &&
				err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdS> error: %s processing event %+v with EEs.", err.Error(), cgrEv))
			}
		}()
	} else if errExec := t.connMgr.Call(context.TODO(), t.cgrcfg.ThresholdSCfg().EEsConns,
		utils.EeSv1ProcessEvent,
		cgrEventWithID, &reply); errExec != nil &&
		errExec.Error() != utils.ErrNotFound.Error() {
		utils.Logger.Warning(
			fmt.Sprintf("<ThresholdS> error: %s processing event %+v with EEs.", errExec.Error(), cgrEv))
		err = utils.ErrPartiallyExecuted
	}
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
func NewThresholdService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS, conn *ConnManager) *ThresholdService {
	return &ThresholdService{
		dm:          dm,
		cgrcfg:      cgrcfg,
		filterS:     filterS,
		connMgr:     conn,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}),
		storedTdIDs: make(utils.StringSet),
		biJClnts:    make(map[birpc.ClientConnector]string),
		biJIDs:      make(map[string]*biJClient),
	}
}

// ThresholdService manages Threshold execution and storing them to dataDB
type ThresholdService struct {
	dm          *DataManager
	cgrcfg      *config.CGRConfig
	filterS     *FilterS
	stopBackup  chan struct{}
	loopStopped chan struct{}
	storedTdIDs utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	stMux       sync.RWMutex    // protects storedTdIDs
	connMgr     *ConnManager
	biJMux      sync.RWMutex                     // mux protecting BI-JSON connections
	biJClnts    map[birpc.ClientConnector]string // index BiJSONConnection so we can sync them later
	biJIDs      map[string]*biJClient            // identifiers of bidirectional JSON conns, used to call RPC based on connIDs
}

// Reload stops the backupLoop and restarts it
func (tS *ThresholdService) Reload() {
	close(tS.stopBackup)
	<-tS.loopStopped // wait until the loop is done
	tS.stopBackup = make(chan struct{})
	go tS.runBackup()
}

// StartLoop starts the gorutine with the backup loop
func (tS *ThresholdService) StartLoop() {
	go tS.runBackup()
}

// Shutdown is called to shutdown the service
func (tS *ThresholdService) Shutdown() {
	utils.Logger.Info("<ThresholdS> shutdown initialized")
	close(tS.stopBackup)
	tS.storeThresholds()
	utils.Logger.Info("<ThresholdS> shutdown complete")
}

// OnBiJSONConnect handles new client connections.
func (tS *ThresholdService) OnBiJSONConnect(c birpc.ClientConnector) {
	nodeID := utils.UUIDSha1Prefix() // connection identifier, should be later updated as login procedure
	tS.biJMux.Lock()
	tS.biJClnts[c] = nodeID
	tS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: 2.0,
	}
	tS.biJMux.Unlock()
}

// OnBiJSONDisconnect handles client disconnects
func (tS *ThresholdService) OnBiJSONDisconnect(c birpc.ClientConnector) {
	tS.biJMux.Lock()
	if nodeID, has := tS.biJClnts[c]; has {
		delete(tS.biJClnts, c)
		delete(tS.biJIDs, nodeID)
	}
	tS.biJMux.Unlock()
}

// RegisterIntBiJConn is called on internal BiJ connection towards ThresholdS
func (tS *ThresholdService) RegisterIntBiJConn(c birpc.ClientConnector, nodeID string) {
	if nodeID == utils.EmptyString {
		nodeID = tS.cgrcfg.GeneralCfg().NodeID
	}
	tS.biJMux.Lock()
	tS.biJClnts[c] = nodeID
	tS.biJIDs[nodeID] = &biJClient{
		conn:  c,
		proto: 2.0,
	}
	tS.biJMux.Unlock()
}

// biJClient contains info we need to reach back a bidirectional json client
type biJClient struct {
	conn  birpc.ClientConnector // connection towards BiJ client
	proto float64               // client protocol version
}

// backup will regularly store thresholds changed to dataDB
func (tS *ThresholdService) runBackup() {
	storeInterval := tS.cgrcfg.ThresholdSCfg().StoreInterval
	if storeInterval <= 0 {
		tS.loopStopped <- struct{}{}
		return
	}
	for {
		tS.storeThresholds()
		select {
		case <-tS.stopBackup:
			tS.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeThresholds represents one task of complete backup
func (tS *ThresholdService) storeThresholds() {
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
			utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed retrieving from cache threshold with ID: %s", tID))
			continue
		}
		t := tIf.(*Threshold)
		t.lock(utils.EmptyString)
		if err := tS.StoreThreshold(t); err != nil {
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
func (tS *ThresholdService) StoreThreshold(t *Threshold) (err error) {
	if t.dirty == nil || !*t.dirty {
		return
	}
	if err = tS.dm.SetThreshold(t); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ThresholdS> failed saving Threshold with tenant: %s and ID: %s, error: %s",
				t.Tenant, t.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := t.TenantID(); Cache.HasItem(utils.CacheThresholds, tntID) { // only cache if previously there
		if err = Cache.Set(utils.CacheThresholds, tntID, t, nil,
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
func (tS *ThresholdService) matchingThresholdsForEvent(tnt string, args *utils.CGREvent) (ts Thresholds, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	var thdIDs []string
	if thdIDs, err = utils.GetStringSliceOpts(args, tS.cgrcfg.ThresholdSCfg().Opts.ProfileIDs,
		utils.OptsThresholdsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = utils.GetBoolOpts(args, tS.cgrcfg.ThresholdSCfg().Opts.ProfileIgnoreFilters,
		utils.OptsThresholdsProfileIgnoreFilters); err != nil {
		return
	}
	tIDs := utils.NewStringSet(thdIDs)
	if len(tIDs) == 0 {
		ignFilters = false
		tIDs, err = MatchingItemIDsForEvent(evNm,
			tS.cgrcfg.ThresholdSCfg().StringIndexedFields,
			tS.cgrcfg.ThresholdSCfg().PrefixIndexedFields,
			tS.cgrcfg.ThresholdSCfg().SuffixIndexedFields,
			tS.cgrcfg.ThresholdSCfg().ExistsIndexedFields,
			tS.dm, utils.CacheThresholdFilterIndexes, tnt,
			tS.cgrcfg.ThresholdSCfg().IndexedSelects,
			tS.cgrcfg.ThresholdSCfg().NestedFields,
		)
		if err != nil {
			return nil, err
		}
	}

	// Lock items in sorted order to prevent AB-BA deadlock.
	itemIDs := slices.Sorted(maps.Keys(tIDs))

	ts = make(Thresholds, 0, len(itemIDs))
	for _, id := range itemIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdProfileLockKey(tnt, id))
		var tPrfl *ThresholdProfile
		if tPrfl, err = tS.dm.GetThresholdProfile(tnt, id, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			ts.unlock()
			return nil, err
		}
		tPrfl.lock(lkPrflID)
		if tPrfl.ActivationInterval != nil && args.Time != nil &&
			!tPrfl.ActivationInterval.IsActiveAtTime(*args.Time) { // not active
			tPrfl.unlock()
			continue
		}
		if !ignFilters {
			var pass bool
			if pass, err = tS.filterS.Pass(tnt, tPrfl.FilterIDs,
				evNm); err != nil {
				tPrfl.unlock()
				ts.unlock()
				return nil, err
			} else if !pass {
				tPrfl.unlock()
				continue
			}
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			thresholdLockKey(tPrfl.Tenant, tPrfl.ID))
		var t *Threshold
		if t, err = tS.dm.GetThreshold(tPrfl.Tenant, tPrfl.ID, true, true, ""); err != nil {
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

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdService) processEvent(tnt string, args *utils.CGREvent) (thresholdsIDs []string, err error) {
	var matchTs Thresholds
	if matchTs, err = tS.matchingThresholdsForEvent(tnt, args); err != nil {
		return nil, err
	}
	var withErrors bool
	thresholdsIDs = make([]string, 0, len(matchTs))
	for _, t := range matchTs {
		if t.tPrfl.MaxHits != -1 && t.Hits >= t.tPrfl.MaxHits {
			continue // MaxHits will disable the threshold
		}
		thresholdsIDs = append(thresholdsIDs, t.ID)
		t.Hits++
		if time.Now().After(t.Snooze) && // snoozed, not executing actions
			t.Hits >= t.tPrfl.MinHits && // number of hits was not met, will not execute actions
			(t.tPrfl.MaxHits == -1 ||
				t.Hits <= t.tPrfl.MaxHits) {
			var tntAcnt string
			var acnt string
			if utils.IfaceAsString(args.APIOpts[utils.MetaEventType]) == utils.AccountUpdate {
				acnt, _ = args.FieldAsString(utils.ID)
			} else {
				acnt, _ = args.FieldAsString(utils.AccountField)
			}
			if _, has := args.APIOpts[utils.MetaAccountID]; has {
				acnt, _ = args.OptAsString(utils.MetaAccountID)
			}
			if acnt != utils.EmptyString {
				tntAcnt = utils.ConcatenatedKey(args.Tenant, acnt)
			}
			for _, actionSetID := range t.tPrfl.ActionIDs {
				at := &ActionTiming{
					Uuid:      utils.GenUUID(),
					ActionsID: actionSetID,
					ExtraData: args,
				}
				if tntAcnt != utils.EmptyString {
					at.accountIDs = utils.NewStringMap(tntAcnt)
				}
				if t.tPrfl.Async {
					go func(setID string) {
						if errExec := at.Execute(tS.filterS, utils.ThresholdS); errExec != nil {
							utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", setID, errExec.Error()))
						}
					}(actionSetID)
				} else if errExec := at.Execute(tS.filterS, utils.ThresholdS); errExec != nil {
					utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions: %s, error: %s", actionSetID, errExec.Error()))
					withErrors = true
				}
			}
			t.Snooze = time.Now().Add(t.tPrfl.MinSleep)
			if err = tS.processEEs(args.APIOpts, t); err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<ThresholdService> received error: %s when processing with EEs.", err.Error()))
				withErrors = true
			}
		}
		*t.dirty = true // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			tS.StoreThreshold(t)
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
func (tS *ThresholdService) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
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
func (tS *ThresholdService) V1GetThresholdsForEvent(ctx *context.Context, args *utils.CGREvent, reply *Thresholds) (err error) {
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
		tnt = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	var ts Thresholds
	if ts, err = tS.matchingThresholdsForEvent(tnt, args); err == nil {
		*reply = ts
		ts.unlock()
	}
	return
}

// V1GetThresholdIDs returns list of thresholdIDs configured for a tenant
func (tS *ThresholdService) V1GetThresholdIDs(ctx *context.Context, tenant string, tIDs *[]string) (err error) {
	// var rply string // unfinished remove later
	// if err := ctx.Client.Call(ctx, utils.SessionSv1Ping, &utils.CGREvent{}, &rply); err != nil {
	// }
	if tenant == utils.EmptyString {
		tenant = tS.cgrcfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdPrefix + tenant + utils.ConcatenatedKeySep
	keys, err := tS.dm.DataDB().GetKeysForPrefix(prfx, utils.EmptyString)
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
	if thd, err = tS.dm.GetThreshold(tnt, tntID.ID, true, true, ""); err != nil {
		return
	}
	*t = *thd
	return
}

// V1ResetThreshold resets the threshold hits
// If the threshold does not exist (e.g., removed after MaxHits), it attempts to recreate it
// based on its ThresholdProfile.
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
	if thd, err = tS.dm.GetThreshold(tnt, tntID.ID, true, true, ""); err != nil {
		utils.Logger.Warning(fmt.Sprintf("threshold with ID %s not found", tntID.ID))
		return err
	}
	if thd.Hits != 0 {
		thd.Hits = 0
		thd.Snooze = time.Time{}
		thd.dirty = utils.BoolPointer(true) // mark it to be saved
		if tS.cgrcfg.ThresholdSCfg().StoreInterval == -1 {
			if err = tS.StoreThreshold(thd); err != nil {
				return err
			}
		} else {
			tS.stMux.Lock()
			tS.storedTdIDs.Add(thd.TenantID())
			tS.stMux.Unlock()
		}
	}
	*rply = utils.OK
	return nil
}

// BiRPCv1RegisterInternalBiJSONConn will register the internal BiRPC connection towards ThresholdS
func (tS *ThresholdService) BiRPCv1RegisterInternalBiJSONConn(ctx *context.Context,
	connID string, reply *string) error {
	tS.RegisterIntBiJConn(ctx.Client, connID)
	*reply = utils.OK
	return nil
}
