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

package resources

import (
	"cmp"
	"fmt"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/guardian"
)

// matchedResource holds a resource together with state set during matching.
type matchedResource struct {
	Resource   *utils.Resource
	ttl        *time.Duration
	totalUsage *float64
	dirty      *bool // the usages were modified, needs save, *bool so we only save if enabled in config
	profile    *utils.ResourceProfile
}

// removeExpiredUnits removes units which are expired from the resource
func (r *matchedResource) removeExpiredUnits() {
	var firstActive int
	for _, rID := range r.Resource.TTLIdx {
		if ru, has := r.Resource.Usages[rID]; has && ru.IsActive(time.Now()) {
			break
		}
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	for _, rID := range r.Resource.TTLIdx[:firstActive] {
		ru, has := r.Resource.Usages[rID]
		if !has {
			continue
		}
		delete(r.Resource.Usages, rID)
		if r.totalUsage != nil { //  total usage was not yet calculated so we do not need to update it
			*r.totalUsage -= ru.Units
			if *r.totalUsage < 0 { // something went wrong
				utils.Logger.Warning(
					fmt.Sprintf("resetting total usage for resourceID: %s, usage smaller than 0: %f", r.Resource.ID, *r.totalUsage))
				r.totalUsage = nil
			}
		}
	}
	r.Resource.TTLIdx = r.Resource.TTLIdx[firstActive:]
	r.totalUsage = nil
}

// recordUsage records a new usage
func (r *matchedResource) recordUsage(ru *utils.ResourceUsage) error {
	if _, hasID := r.Resource.Usages[ru.ID]; hasID {
		return fmt.Errorf("duplicate resource usage with id: %s", ru.TenantID())
	}
	if r.ttl != nil && *r.ttl != -1 {
		if *r.ttl == 0 {
			return nil // no recording for ttl of 0
		}
		ru = ru.Clone() // don't influence the initial ru
		ru.ExpiryTime = time.Now().Add(*r.ttl)
	}
	r.Resource.Usages[ru.ID] = ru
	if r.totalUsage != nil {
		*r.totalUsage += ru.Units
	}
	if !ru.ExpiryTime.IsZero() {
		r.Resource.TTLIdx = append(r.Resource.TTLIdx, ru.ID)
	}
	return nil
}

// clearUsage clears the usage for an ID
func (r *matchedResource) clearUsage(ruID string) error {
	ru, hasIt := r.Resource.Usages[ruID]
	if !hasIt {
		return fmt.Errorf("cannot find usage record with id: %s", ruID)
	}
	if !ru.ExpiryTime.IsZero() {
		for i, ruIDIdx := range r.Resource.TTLIdx {
			if ruIDIdx == ruID {
				r.Resource.TTLIdx = append(r.Resource.TTLIdx[:i], r.Resource.TTLIdx[i+1:]...)
				break
			}
		}
	}
	if r.totalUsage != nil {
		*r.totalUsage -= ru.Units
	}
	delete(r.Resource.Usages, ruID)
	return nil
}

// Resources is a collection of matchedResource objects.
type Resources []*matchedResource

// resIDsMp returns a map of resource IDs which is used for caching
func (rs Resources) resIDsMp() utils.StringSet {
	mp := make(utils.StringSet)
	for _, r := range rs {
		mp.Add(r.Resource.ID)
	}
	return mp
}

// recordUsage will record the usage in all the resource limits, failing back on errors
func (rs Resources) recordUsage(ru *utils.ResourceUsage) error {
	var nonReservedIdx int // index of first resource not reserved
	var err error
	for _, r := range rs {
		if err = r.recordUsage(ru); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s>cannot record usage, err: %s", utils.ResourceS, err.Error()))
			break
		}
		nonReservedIdx++
	}
	if err != nil {
		for _, r := range rs[:nonReservedIdx] {
			if errClear := r.clearUsage(ru.ID); errClear != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> cannot clear usage, err: %s", utils.ResourceS, errClear.Error()))
			} // best effort
		}
	}
	return err
}

// clearUsage gives back the units to the pool
func (rs Resources) clearUsage(ruTntID string) error {
	var err error
	for _, r := range rs {
		if errClear := r.clearUsage(ruTntID); errClear != nil &&
			r.ttl != nil && *r.ttl != 0 { // we only consider not found error in case of ttl different than 0
			utils.Logger.Warning(fmt.Sprintf("<%s>, clear ruID: %s, err: %s", utils.ResourceS, ruTntID, errClear.Error()))
			err = errClear
		}
	}
	return err
}

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) allocateResource(ru *utils.ResourceUsage, dryRun bool) (allocMsg string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	for _, r := range rs {
		r.removeExpiredUnits()
		if _, hasID := r.Resource.Usages[ru.ID]; hasID && !dryRun { // update
			_ = r.clearUsage(ru.ID) // can't fail: we just checked hasID
		}
		if allocMsg == "" &&
			(r.profile.Limit >= r.Resource.TotalUsage()+ru.Units || r.profile.Limit == -1) {
			allocMsg = utils.FirstNonEmpty(r.profile.AllocationMessage, r.profile.ID)
		}
	}
	if allocMsg == "" {
		return "", utils.ErrResourceUnavailable
	}
	if dryRun {
		return allocMsg, nil
	}
	_ = rs.recordUsage(ru) // can't error: dup check already cleared above
	return allocMsg, nil
}

// NewResourceService  returns a new ResourceService
func NewResourceService(dm *engine.DataManager, cgrcfg *config.CGRConfig,
	filterS *engine.FilterS, connMgr *engine.ConnManager) *ResourceS {
	return &ResourceS{dm: dm,
		storedResources: make(utils.StringSet),
		cfg:             cgrcfg,
		fltrS:           filterS,
		loopStopped:     make(chan struct{}),
		stopBackup:      make(chan struct{}),
		connMgr:         connMgr,
	}

}

// ResourceS is the service handling resources
type ResourceS struct {
	dm              *engine.DataManager // So we can load the data in cache and index it
	fltrS           *engine.FilterS
	storedResources utils.StringSet // keep a record of resources which need saving, map[resID]bool
	srMux           sync.RWMutex    // protects storedResources
	cfg             *config.CGRConfig
	stopBackup      chan struct{} // control storing process
	loopStopped     chan struct{}
	connMgr         *engine.ConnManager
}

// Reload stops the backupLoop and restarts it
func (rS *ResourceS) Reload(ctx *context.Context) {
	close(rS.stopBackup)
	<-rS.loopStopped // wait until the loop is done
	rS.stopBackup = make(chan struct{})
	go rS.runBackup(ctx)
}

// StartLoop starts the gorutine with the backup loop
func (rS *ResourceS) StartLoop(ctx *context.Context) {
	go rS.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (rS *ResourceS) Shutdown(ctx *context.Context) {
	close(rS.stopBackup)
	rS.storeResources(ctx)
}

// backup will regularly store resources changed to dataDB
func (rS *ResourceS) runBackup(ctx *context.Context) {
	storeInterval := rS.cfg.ResourceSCfg().StoreInterval
	if storeInterval <= 0 {
		rS.loopStopped <- struct{}{}
		return
	}
	for {
		rS.storeResources(ctx)
		select {
		case <-rS.stopBackup:
			rS.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeResources represents one task of complete backup
func (rS *ResourceS) storeResources(ctx *context.Context) {
	var failedRIDs []string
	for { // don't stop until we store all dirty resources
		rS.srMux.Lock()
		rID := rS.storedResources.GetOne()
		if rID != "" {
			rS.storedResources.Remove(rID)
		}
		rS.srMux.Unlock()
		if rID == "" {
			break // no more keys, backup completed
		}
		rIf, ok := engine.Cache.Get(utils.CacheResources, rID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache resource with ID: %s", utils.ResourceS, rID))
			continue
		}
		r := &matchedResource{
			Resource: rIf.(*utils.Resource),
			// NOTE: dirty is hardcoded to true, otherwise resources would
			// never be stored.
			// Previously, dirty was part of the cached resource.
			dirty: utils.BoolPointer(true),
		}
		lkID := guardian.Guardian.GuardIDs("",
			rS.cfg.GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(r.Resource.Tenant, r.Resource.ID))
		if err := rS.storeResource(ctx, r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		guardian.Guardian.UnguardIDs(lkID)
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		rS.srMux.Lock()
		rS.storedResources.AddSlice(failedRIDs)
		rS.srMux.Unlock()
	}
}

// StoreResource stores the resource in DB and corrects dirty flag
func (rS *ResourceS) storeResource(ctx *context.Context, r *matchedResource) error {
	if r.dirty == nil || !*r.dirty {
		return nil
	}
	if err := rS.dm.SetResource(ctx, r.Resource); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed saving Resource with ID: %s, error: %s",
				r.Resource.ID, err.Error()))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := r.Resource.TenantID(); engine.Cache.HasItem(utils.CacheResources, tntID) { // only cache if previously there
		if err := engine.Cache.Set(ctx, utils.CacheResources, tntID, r.Resource, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ResourceS> failed caching Resource with ID: %s, error: %s",
					tntID, err.Error()))
			return err
		}
	}
	*r.dirty = false
	return nil
}

// storeMatchedResources will store the list of resources based on the StoreInterval
func (rS *ResourceS) storeMatchedResources(ctx *context.Context, mtcRLs Resources) error {
	if rS.cfg.ResourceSCfg().StoreInterval == 0 {
		return nil
	}
	if rS.cfg.ResourceSCfg().StoreInterval > 0 {
		rS.srMux.Lock()
		defer rS.srMux.Unlock()
	}
	for _, r := range mtcRLs {
		if r.dirty != nil {
			*r.dirty = true // mark it to be saved
			if rS.cfg.ResourceSCfg().StoreInterval > 0 {
				rS.storedResources.Add(r.Resource.TenantID())
				continue
			}
			if err := rS.storeResource(ctx, r); err != nil {
				return err
			}
		}

	}
	return nil
}

// processThresholds will pass the event for resource to ThresholdS
func (rS *ResourceS) processThresholds(ctx *context.Context, rs Resources, opts map[string]any) error {
	threshConns, err := engine.GetConnIDs(ctx, rS.cfg.ResourceSCfg().Conns[utils.MetaThresholds], utils.MetaAny, utils.MapStorage{}, rS.fltrS)
	if err != nil {
		return err
	}
	if len(threshConns) == 0 {
		return nil
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	opts[utils.MetaEventType] = utils.ResourceUpdate

	var withErrs bool
	for _, r := range rs {
		if len(r.profile.ThresholdIDs) == 1 &&
			r.profile.ThresholdIDs[0] == utils.MetaNone {
			continue
		}
		opts[utils.OptsThresholdsProfileIDs] = r.profile.ThresholdIDs

		thEv := &utils.CGREvent{
			Tenant: r.Resource.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType:  utils.ResourceUpdate,
				utils.ResourceID: r.Resource.ID,
				utils.Usage:      r.Resource.TotalUsage(),
			},
			APIOpts: opts,
		}
		var tIDs []string
		if err := rS.connMgr.Call(ctx, threshConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(r.profile.ThresholdIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with %s.",
					utils.ResourceS, err.Error(), thEv, utils.ThresholdS))
			withErrs = true
		}
	}
	if withErrs {
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceS) matchingResourcesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (rs Resources, unlock func(), err error) {

	var lockIDs []string
	unlockAll := func() {
		for _, lkID := range lockIDs {
			guardian.Guardian.UnguardIDs(lkID)
		}
	}

	var rIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := engine.Cache.Get(utils.CacheEventResources, evUUID); ok { // The ResourceIDs were cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, nil, utils.ErrNotFound
		}
		rIDs = x.(utils.StringSet)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				// TODO: Consider using RemoveWithoutReplicate instead, as
				// partitions with Replicate=true call ReplicateRemove in
				// onEvict by default.
				if errCh := engine.Cache.Remove(ctx, utils.CacheEventResources, evUUID,
					true, utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()

	} else { // select the resourceIDs out of dataDB
		rIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
			rS.cfg.ResourceSCfg().StringIndexedFields,
			rS.cfg.ResourceSCfg().PrefixIndexedFields,
			rS.cfg.ResourceSCfg().SuffixIndexedFields,
			rS.cfg.ResourceSCfg().ExistsIndexedFields,
			rS.cfg.ResourceSCfg().NotExistsIndexedFields,
			rS.dm, utils.CacheResourceFilterIndexes, tnt,
			rS.cfg.ResourceSCfg().IndexedSelects,
			rS.cfg.ResourceSCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := engine.Cache.Set(ctx, utils.CacheEventResources, evUUID, nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, nil, errCh
				}
			}
			return nil, nil, err
		}
	}
	rs = make(Resources, 0, len(rIDs))
	weights := make(map[string]float64) // stores sorting weights by resource ID
	for resName := range rIDs {
		lkID := guardian.Guardian.GuardIDs("",
			rS.cfg.GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(tnt, resName))

		rp, err := rS.dm.GetResourceProfile(ctx, tnt, resName,
			true, true, utils.NonTransactional)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			if err == utils.ErrNotFound {
				continue
			}
			unlockAll()
			return nil, nil, err
		}

		pass, err := rS.fltrS.Pass(ctx, tnt, rp.FilterIDs, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}
		if !pass {
			guardian.Guardian.UnguardIDs(lkID)
			continue
		}

		res, err := rS.dm.GetResource(ctx, rp.Tenant, rp.ID, true, true, "")
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}

		weight, err := engine.WeightFromDynamics(ctx, rp.Weights, rS.fltrS, tnt, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}

		lockIDs = append(lockIDs, lkID)

		r := &matchedResource{
			Resource: res,
			profile:  rp,
		}
		if rp.Stored && r.dirty == nil {
			r.dirty = utils.BoolPointer(false)
		}
		if usageTTL != nil {
			if *usageTTL != 0 {
				r.ttl = usageTTL
			}
		} else if rp.UsageTTL >= 0 {
			r.ttl = utils.DurationPointer(rp.UsageTTL)
		}
		weights[r.Resource.ID] = weight
		rs = append(rs, r)
	}

	if len(rs) == 0 {
		unlockAll()
		return nil, nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(rs, func(a, b *matchedResource) int {
		return cmp.Compare(weights[b.Resource.ID], weights[a.Resource.ID])
	})

	for i, r := range rs {
		if r.profile.Blocker && i != len(rs)-1 { // blocker will stop processing and we are not at last index
			for _, lkID := range lockIDs[i+1:] {
				guardian.Guardian.UnguardIDs(lkID)
			}
			lockIDs = lockIDs[:i+1]
			rs = rs[:i+1]
			break
		}
	}
	if errCh := engine.Cache.Set(ctx, utils.CacheEventResources, evUUID, rs.resIDsMp(), nil, true, ""); errCh != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> failed caching event resources: %s", utils.ResourceS, errCh.Error()))
	}
	return rs, unlockAll, nil
}
