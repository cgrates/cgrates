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
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// ThresholdProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ThresholdProfileWithAPIOpts struct {
	*ThresholdProfile
	APIOpts map[string]any
}

// ThresholdProfile the profile for threshold
type ThresholdProfile struct {
	Tenant           string
	ID               string
	FilterIDs        []string
	MaxHits          int
	MinHits          int
	MinSleep         time.Duration
	Blocker          bool                 // blocker flag to stop processing on filters matched
	Weights          utils.DynamicWeights // Weight to sort the thresholds
	ActionProfileIDs []string
	Async            bool

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
		Async:    tp.Async,
	}
	if tp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(tp.FilterIDs))
		copy(clone.FilterIDs, tp.FilterIDs)
	}
	if tp.ActionProfileIDs != nil {
		clone.ActionProfileIDs = make([]string, len(tp.ActionProfileIDs))
		copy(clone.ActionProfileIDs, tp.ActionProfileIDs)
	}
	if tp.Weights != nil {
		clone.Weights = tp.Weights.Clone()
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
		t.tPrfl.MaxHits != -1 &&
			t.Hits > t.tPrfl.MaxHits ||
		len(t.tPrfl.ActionProfileIDs) == 1 &&
			t.tPrfl.ActionProfileIDs[0] == utils.MetaNone {
		return
	}

	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}

	args.APIOpts[utils.OptsActionsProfileIDs] = t.tPrfl.ActionProfileIDs

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
		if err = connMgr.Call(ctx, actionsConns, utils.ActionSv1ExecuteActions, args, &reply); err != nil {
			return
		}
	} else {
		go func() {
			if errExec := connMgr.Call(context.Background(), actionsConns, utils.ActionSv1ExecuteActions,
				args, &reply); errExec != nil {
				utils.Logger.Warning(fmt.Sprintf("<ThresholdS> failed executing actions for threshold: %s, error: %s", t.TenantID(), errExec.Error()))
			}
		}()
	}
	t.Snooze = time.Now().Add(t.tPrfl.MinSleep)
	return
}

// unlockThresholds unlocks all locked Thresholds in the given slice.
func unlockThresholds(ts []*Threshold) {
	for _, t := range ts {
		t.unlock()
		if t.tPrfl != nil {
			t.tPrfl.unlock()
		}
	}
}

// NewThresholdService the constructor for ThresoldS service
func NewThresholdService(dm *DataManager, cgrcfg *config.CGRConfig, filterS *FilterS, connMgr *ConnManager) (tS *ThresholdS) {
	return &ThresholdS{
		dm:          dm,
		cfg:         cgrcfg,
		fltrS:       filterS,
		connMgr:     connMgr,
		stopBackup:  make(chan struct{}),
		loopStopped: make(chan struct{}),
		storedTdIDs: make(utils.StringSet),
	}
}

// ThresholdS manages Threshold execution and storing them to dataDB
type ThresholdS struct {
	dm          *DataManager
	cfg         *config.CGRConfig
	fltrS       *FilterS
	connMgr     *ConnManager
	stopBackup  chan struct{}
	loopStopped chan struct{}
	storedTdIDs utils.StringSet // keep a record of stats which need saving, map[statsTenantID]bool
	stMux       sync.RWMutex    // protects storedTdIDs
}

// Reload stops the backupLoop and restarts it
func (tS *ThresholdS) Reload(ctx *context.Context) {
	close(tS.stopBackup)
	<-tS.loopStopped // wait until the loop is done
	tS.stopBackup = make(chan struct{})
	go tS.runBackup(ctx)
}

// StartLoop starts the gorutine with the backup loop
func (tS *ThresholdS) StartLoop(ctx *context.Context) {
	go tS.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (tS *ThresholdS) Shutdown(ctx *context.Context) {
	close(tS.stopBackup)
	tS.storeThresholds(ctx)
}

// backup will regularly store thresholds changed to dataDB
func (tS *ThresholdS) runBackup(ctx *context.Context) {
	storeInterval := tS.cfg.ThresholdSCfg().StoreInterval
	if storeInterval <= 0 {
		tS.loopStopped <- struct{}{}
		return
	}
	for {
		tS.storeThresholds(ctx)
		select {
		case <-tS.stopBackup:
			tS.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeThresholds represents one task of complete backup
func (tS *ThresholdS) storeThresholds(ctx *context.Context) {
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
func (tS *ThresholdS) StoreThreshold(ctx *context.Context, t *Threshold) (err error) {
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
func (tS *ThresholdS) matchingThresholdsForEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (ts []*Threshold, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	var thIDs []string
	if thIDs, err = GetStringSliceOpts(ctx, tnt, evNm, nil, tS.fltrS, tS.cfg.ThresholdSCfg().Opts.ProfileIDs,
		config.ThresholdsProfileIDsDftOpt, utils.OptsThresholdsProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = GetBoolOpts(ctx, tnt, evNm, nil, tS.fltrS, tS.cfg.ThresholdSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}

	tIDs := utils.NewStringSet(thIDs)
	if len(tIDs) == 0 {
		ignFilters = false
		tIDs, err = MatchingItemIDsForEvent(ctx, evNm,
			tS.cfg.ThresholdSCfg().StringIndexedFields,
			tS.cfg.ThresholdSCfg().PrefixIndexedFields,
			tS.cfg.ThresholdSCfg().SuffixIndexedFields,
			tS.cfg.ThresholdSCfg().ExistsIndexedFields,
			tS.cfg.ThresholdSCfg().NotExistsIndexedFields,
			tS.dm, utils.CacheThresholdFilterIndexes, tnt,
			tS.cfg.ThresholdSCfg().IndexedSelects,
			tS.cfg.ThresholdSCfg().NestedFields,
		)
		if err != nil {
			return nil, err
		}
	}
	ts = make([]*Threshold, 0, len(tIDs))
	weights := make(map[string]float64) // stores sorting weights by tID
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
			unlockThresholds(ts)
			return nil, err
		}
		tPrfl.lock(lkPrflID)
		if !ignFilters {
			var pass bool
			if pass, err = tS.fltrS.Pass(ctx, tnt, tPrfl.FilterIDs,
				evNm); err != nil {
				tPrfl.unlock()
				unlockThresholds(ts)
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
		if t, err = tS.dm.GetThreshold(ctx, tPrfl.Tenant, tPrfl.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			tPrfl.unlock()
			if err == utils.ErrNotFound { // corner case where the threshold was removed due to MaxHits
				err = nil
				continue
			}
			unlockThresholds(ts)
			return nil, err
		}
		t.lock(lkID)
		if t.dirty == nil || tPrfl.MaxHits == -1 || t.Hits < tPrfl.MaxHits {
			t.dirty = utils.BoolPointer(false)
		}

		t.tPrfl = tPrfl
		weight, err := WeightFromDynamics(ctx, tPrfl.Weights,
			tS.fltrS, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[t.ID] = weight
		ts = append(ts, t)
	}
	if len(ts) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(ts, func(a, b *Threshold) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	for i, t := range ts {
		if t.tPrfl.Blocker && i != len(ts)-1 { // blocker will stop processing and we are not at last index
			unlockThresholds(ts[i+1:])
			ts = ts[:i+1]
			break
		}
	}
	return
}

// processEvent processes a new event, dispatching to matching thresholds
func (tS *ThresholdS) processEvent(ctx *context.Context, tnt string, args *utils.CGREvent) (thresholdsIDs []string, err error) {
	matchTs, err := tS.matchingThresholdsForEvent(ctx, tnt, args)
	if err != nil {
		return nil, err
	}
	var withErrors bool
	thresholdsIDs = make([]string, 0, len(matchTs))
	for _, t := range matchTs {
		thresholdsIDs = append(thresholdsIDs, t.ID)
		t.Hits++
		if err = processEventWithThreshold(ctx, tS.connMgr,
			tS.cfg.ThresholdSCfg().ActionSConns, args, t); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ThresholdService> threshold: %s, ignoring event: %s, error: %s",
					t.TenantID(), utils.ConcatenatedKey(tnt, args.ID), err.Error()))
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

		// recurrent threshold
		*t.dirty = true // mark it to be saved
		if tS.cfg.ThresholdSCfg().StoreInterval == -1 {
			tS.StoreThreshold(ctx, t)
		} else {
			tS.stMux.Lock()
			tS.storedTdIDs.Add(t.TenantID())
			tS.stMux.Unlock()
		}
	}
	unlockThresholds(matchTs)
	if withErrors {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// V1ProcessEvent implements ThresholdService method for processing an Event
func (tS *ThresholdS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
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
		tnt = tS.cfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	if ids, err = tS.processEvent(ctx, tnt, args); err != nil {
		return
	}
	*reply = ids
	return
}

// V1GetThresholdsForEvent queries thresholds matching an Event
func (tS *ThresholdS) V1GetThresholdsForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*Threshold) (err error) {
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
		tnt = tS.cfg.GeneralCfg().DefaultTenant
	}
	var ts []*Threshold
	if ts, err = tS.matchingThresholdsForEvent(ctx, tnt, args); err == nil {
		*reply = ts
		unlockThresholds(ts)
	}
	return
}

// V1GetThresholdIDs returns list of thresholdIDs configured for a tenant
func (tS *ThresholdS) V1GetThresholdIDs(ctx *context.Context, args *utils.TenantWithAPIOpts, tIDs *[]string) (err error) {
	tenant := args.Tenant
	if tenant == utils.EmptyString {
		tenant = tS.cfg.GeneralCfg().DefaultTenant
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
func (tS *ThresholdS) V1GetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, t *Threshold) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cfg.GeneralCfg().DefaultTenant
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
func (tS *ThresholdS) V1ResetThreshold(ctx *context.Context, tntID *utils.TenantIDWithAPIOpts, rply *string) (err error) {
	var thd *Threshold
	tnt := tntID.Tenant
	if tnt == utils.EmptyString {
		tnt = tS.cfg.GeneralCfg().DefaultTenant
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
		if tS.cfg.ThresholdSCfg().StoreInterval == -1 {
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

func (tp *ThresholdProfile) Set(path []string, val any, _ bool) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}

	switch path[0] {
	default:
		return utils.ErrWrongPath
	case utils.Tenant:
		tp.Tenant = utils.IfaceAsString(val)
	case utils.ID:
		tp.ID = utils.IfaceAsString(val)
	case utils.Blocker:
		tp.Blocker, err = utils.IfaceAsBool(val)
	case utils.Weights:
		if val != utils.EmptyString {
			tp.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
		}
	case utils.FilterIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.FilterIDs = append(tp.FilterIDs, valA...)
	case utils.MaxHits:
		if val != utils.EmptyString {
			tp.MaxHits, err = utils.IfaceAsInt(val)
		}
	case utils.MinHits:
		if val != utils.EmptyString {
			tp.MinHits, err = utils.IfaceAsInt(val)
		}
	case utils.MinSleep:
		tp.MinSleep, err = utils.IfaceAsDuration(val)
	case utils.ActionProfileIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		tp.ActionProfileIDs = append(tp.ActionProfileIDs, valA...)
	case utils.Async:
		tp.Async, err = utils.IfaceAsBool(val)
	}
	return
}

func (tp *ThresholdProfile) Merge(v2 any) {
	vi := v2.(*ThresholdProfile)
	if len(vi.Tenant) != 0 {
		tp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		tp.ID = vi.ID
	}
	tp.FilterIDs = append(tp.FilterIDs, vi.FilterIDs...)
	tp.ActionProfileIDs = append(tp.ActionProfileIDs, vi.ActionProfileIDs...)
	if vi.Blocker {
		tp.Blocker = vi.Blocker
	}
	if vi.Async {
		tp.Async = vi.Async
	}
	tp.Weights = append(tp.Weights, vi.Weights...)
	if vi.MaxHits != 0 {
		tp.MaxHits = vi.MaxHits
	}
	if vi.MinHits != 0 {
		tp.MinHits = vi.MinHits
	}
	if vi.MinSleep != 0 {
		tp.MinSleep = vi.MinSleep
	}
}

func (tp *ThresholdProfile) String() string { return utils.ToJSON(tp) }
func (tp *ThresholdProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = tp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (tp *ThresholdProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case utils.ActionProfileIDs:
				if *idx < len(tp.ActionProfileIDs) {
					return tp.ActionProfileIDs[*idx], nil
				}
			case utils.FilterIDs:
				if *idx < len(tp.FilterIDs) {
					return tp.FilterIDs[*idx], nil
				}
			}
		}
		return nil, utils.ErrNotFound
	case utils.Tenant:
		return tp.Tenant, nil
	case utils.ID:
		return tp.ID, nil
	case utils.FilterIDs:
		return tp.FilterIDs, nil
	case utils.Weights:
		return tp.Weights, nil
	case utils.ActionProfileIDs:
		return tp.ActionProfileIDs, nil
	case utils.MaxHits:
		return tp.MaxHits, nil
	case utils.MinHits:
		return tp.MinHits, nil
	case utils.MinSleep:
		return tp.MinSleep, nil
	case utils.Blocker:
		return tp.Blocker, nil
	case utils.Async:
		return tp.Async, nil
	}
}
