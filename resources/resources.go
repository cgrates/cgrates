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

// resourceProfile represents the user configuration for the resource
type resourceProfile struct {
	ResourceProfile *utils.ResourceProfile
	lkID            string // holds the reference towards guardian lock key

}

// lock will lock the resourceProfile using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (rp *resourceProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.ResourceProfileLockKey(rp.ResourceProfile.Tenant, rp.ResourceProfile.ID))
	}
	rp.lkID = lkID
}

// unlock will unlock the resourceProfile and clear rp.lkID
func (rp *resourceProfile) unlock() {
	if rp.lkID == utils.EmptyString {
		return
	}
	tmp := rp.lkID
	rp.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this resourceProfile
func (rp *resourceProfile) isLocked() bool {
	return rp.lkID != utils.EmptyString
}

// resource represents a resource in the system
// not thread safe, needs locking at process level
type resource struct {
	Resource *utils.Resource
	lkID     string           // ID of the lock used when matching the resource
	ttl      *time.Duration   // time to leave for this resource, picked up on each Resource initialization out of config
	tUsage   *float64         // sum of all usages
	dirty    *bool            // the usages were modified, needs save, *bool so we only save if enabled in config
	rPrf     *resourceProfile // for ordering purposes
}

// lock will lock the resource using guardian and store the lock within r.lkID
// if lkID is passed as argument, the lock is considered as executed
func (r *resource) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(r.Resource.Tenant, r.Resource.ID))
	}
	r.lkID = lkID
}

// unlock will unlock the resource and clear r.lkID
func (r *resource) unlock() {
	if r.lkID == utils.EmptyString {
		return
	}
	tmp := r.lkID
	r.lkID = utils.EmptyString
	guardian.Guardian.UnguardIDs(tmp)
}

// isLocked returns the locks status of this resource
func (r *resource) isLocked() bool {
	return r.lkID != utils.EmptyString
}

// removeExpiredUnits removes units which are expired from the resource
func (r *resource) removeExpiredUnits() {
	var firstActive int
	for _, rID := range r.Resource.TTLIdx {
		if r, has := r.Resource.Usages[rID]; has && r.IsActive(time.Now()) {
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
		if r.tUsage != nil { //  total usage was not yet calculated so we do not need to update it
			*r.tUsage -= ru.Units
			if *r.tUsage < 0 { // something went wrong
				utils.Logger.Warning(
					fmt.Sprintf("resetting total usage for resourceID: %s, usage smaller than 0: %f", r.Resource.ID, *r.tUsage))
				r.tUsage = nil
			}
		}
	}
	r.Resource.TTLIdx = r.Resource.TTLIdx[firstActive:]
	r.tUsage = nil
}

// recordUsage records a new usage
func (r *resource) recordUsage(ru *utils.ResourceUsage) (err error) {
	if _, hasID := r.Resource.Usages[ru.ID]; hasID {
		return fmt.Errorf("duplicate resource usage with id: %s", ru.TenantID())
	}
	if r.ttl != nil && *r.ttl != -1 {
		if *r.ttl == 0 {
			return // no recording for ttl of 0
		}
		ru = ru.Clone() // don't influence the initial ru
		ru.ExpiryTime = time.Now().Add(*r.ttl)
	}
	r.Resource.Usages[ru.ID] = ru
	if r.tUsage != nil {
		*r.tUsage += ru.Units
	}
	if !ru.ExpiryTime.IsZero() {
		r.Resource.TTLIdx = append(r.Resource.TTLIdx, ru.ID)
	}
	return
}

// clearUsage clears the usage for an ID
func (r *resource) clearUsage(ruID string) (err error) {
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
	if r.tUsage != nil {
		*r.tUsage -= ru.Units
	}
	delete(r.Resource.Usages, ruID)
	return
}

// Resources is a collection of Resource objects.
type Resources []*resource

// unlock will unlock resources part of this slice
func (rs Resources) unlock() {
	for _, r := range rs {
		r.unlock()
		if r.rPrf != nil {
			r.rPrf.unlock()
		}
	}
}

// resIDsMp returns a map of resource IDs which is used for caching
func (rs Resources) resIDsMp() (mp utils.StringSet) {
	mp = make(utils.StringSet)
	for _, r := range rs {
		mp.Add(r.Resource.ID)
	}
	return mp
}

// recordUsage will record the usage in all the resource limits, failing back on errors
func (rs Resources) recordUsage(ru *utils.ResourceUsage) (err error) {
	var nonReservedIdx int // index of first resource not reserved
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
	return
}

// clearUsage gives back the units to the pool
func (rs Resources) clearUsage(ruTntID string) (err error) {
	for _, r := range rs {
		if errClear := r.clearUsage(ruTntID); errClear != nil &&
			r.ttl != nil && *r.ttl != 0 { // we only consider not found error in case of ttl different than 0
			utils.Logger.Warning(fmt.Sprintf("<%s>, clear ruID: %s, err: %s", utils.ResourceS, ruTntID, errClear.Error()))
			err = errClear
		}
	}
	return
}

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs Resources) allocateResource(ru *utils.ResourceUsage, dryRun bool) (alcMessage string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	// Simulate resource usage
	for _, r := range rs {
		r.removeExpiredUnits()
		if _, hasID := r.Resource.Usages[ru.ID]; hasID && !dryRun { // update
			r.clearUsage(ru.ID) // clearUsage returns error only when ru.ID does not exist in the Usages map
		}
		if r.rPrf == nil {
			err = fmt.Errorf("empty configuration for resourceID: %s", r.Resource.TenantID())
			return
		}
		if alcMessage == utils.EmptyString &&
			(r.rPrf.ResourceProfile.Limit >= r.Resource.TotalUsage()+ru.Units || r.rPrf.ResourceProfile.Limit == -1) {
			alcMessage = utils.FirstNonEmpty(r.rPrf.ResourceProfile.AllocationMessage, r.rPrf.ResourceProfile.ID)
		}
	}
	if alcMessage == "" {
		err = utils.ErrResourceUnavailable
		return
	}
	if dryRun {
		return
	}
	rs.recordUsage(ru) // recordUsage returns error only when ru.ID already exists in the Usages map
	return
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
		r := &resource{
			Resource: rIf.(*utils.Resource),

			// NOTE: dirty is hardcoded to true, otherwise resources would
			// never be stored.
			// Previously, dirty was part of the cached resource.
			dirty: utils.BoolPointer(true),
		}
		r.lock(utils.EmptyString)
		if err := rS.storeResource(ctx, r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		r.unlock()
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
func (rS *ResourceS) storeResource(ctx *context.Context, r *resource) (err error) {
	if r.dirty == nil || !*r.dirty {
		return
	}
	if err = rS.dm.SetResource(ctx, r.Resource); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<ResourceS> failed saving Resource with ID: %s, error: %s",
				r.Resource.ID, err.Error()))
		return
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := r.Resource.TenantID(); engine.Cache.HasItem(utils.CacheResources, tntID) { // only cache if previously there
		if err = engine.Cache.Set(ctx, utils.CacheResources, tntID, r.Resource, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<ResourceS> failed caching Resource with ID: %s, error: %s",
					tntID, err.Error()))
			return
		}
	}
	*r.dirty = false
	return
}

// storeMatchedResources will store the list of resources based on the StoreInterval
func (rS *ResourceS) storeMatchedResources(ctx *context.Context, mtcRLs Resources) (err error) {
	if rS.cfg.ResourceSCfg().StoreInterval == 0 {
		return
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
			if err = rS.storeResource(ctx, r); err != nil {
				return
			}
		}

	}
	return
}

// processThresholds will pass the event for resource to ThresholdS
func (rS *ResourceS) processThresholds(ctx *context.Context, rs Resources, opts map[string]any) (err error) {
	if len(rS.cfg.ResourceSCfg().ThresholdSConns) == 0 {
		return
	}
	if opts == nil {
		opts = make(map[string]any)
	}
	opts[utils.MetaEventType] = utils.ResourceUpdate

	var withErrs bool
	for _, r := range rs {
		if len(r.rPrf.ResourceProfile.ThresholdIDs) == 1 &&
			r.rPrf.ResourceProfile.ThresholdIDs[0] == utils.MetaNone {
			continue
		}
		opts[utils.OptsThresholdsProfileIDs] = r.rPrf.ResourceProfile.ThresholdIDs

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
		if err := rS.connMgr.Call(ctx, rS.cfg.ResourceSCfg().ThresholdSConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(r.rPrf.ResourceProfile.ThresholdIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %s processing event %+v with %s.",
					utils.ResourceS, err.Error(), thEv, utils.ThresholdS))
			withErrs = true
		}
	}
	if withErrs {
		err = utils.ErrPartiallyExecuted
	}
	return
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (rS *ResourceS) matchingResourcesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (rs Resources, err error) {
	var rIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := engine.Cache.Get(utils.CacheEventResources, evUUID); ok { // The ResourceIDs were cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
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
					return nil, errCh
				}
			}
			return
		}
	}
	rs = make(Resources, 0, len(rIDs))
	weights := make(map[string]float64) // stores sorting weights by resource ID
	for resName := range rIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.ResourceProfileLockKey(tnt, resName))
		var rp *utils.ResourceProfile
		if rp, err = rS.dm.GetResourceProfile(ctx, tnt, resName,
			true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			rs.unlock()
			return
		}
		rPrf := &resourceProfile{
			ResourceProfile: rp,
		}
		rPrf.lock(lkPrflID)
		var pass bool
		if pass, err = rS.fltrS.Pass(ctx, tnt, rPrf.ResourceProfile.FilterIDs,
			evNm); err != nil {
			rPrf.unlock()
			rs.unlock()
			return nil, err
		} else if !pass {
			rPrf.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(rPrf.ResourceProfile.Tenant, rPrf.ResourceProfile.ID))
		var res *utils.Resource
		if res, err = rS.dm.GetResource(ctx, rPrf.ResourceProfile.Tenant, rPrf.ResourceProfile.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			rPrf.unlock()
			rs.unlock()
			return nil, err
		}
		r := &resource{
			Resource: res,
		}
		r.lock(lkID) // pass the lock into resource so we have it as reference
		if rPrf.ResourceProfile.Stored && r.dirty == nil {
			r.dirty = utils.BoolPointer(false)
		}
		if usageTTL != nil {
			if *usageTTL != 0 {
				r.ttl = usageTTL
			}
		} else if rPrf.ResourceProfile.UsageTTL >= 0 {
			r.ttl = utils.DurationPointer(rPrf.ResourceProfile.UsageTTL)
		}
		r.rPrf = rPrf
		weight, err := engine.WeightFromDynamics(ctx, rPrf.ResourceProfile.Weights, rS.fltrS, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[r.Resource.ID] = weight
		rs = append(rs, r)
	}

	if len(rs) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(rs, func(a, b *resource) int {
		return cmp.Compare(weights[b.Resource.ID], weights[a.Resource.ID])
	})

	for i, r := range rs {
		if r.rPrf.ResourceProfile.Blocker && i != len(rs)-1 { // blocker will stop processing and we are not at last index
			Resources(rs[i+1:]).unlock()
			rs = rs[:i+1]
			break
		}
	}
	if err = engine.Cache.Set(ctx, utils.CacheEventResources, evUUID, rs.resIDsMp(), nil, true, ""); err != nil {
		rs.unlock()
	}
	return
}
