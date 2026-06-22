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
	"maps"
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
	resource *utils.Resource
	ttl      *time.Duration
	profile  *utils.ResourceProfile
	weight   float64
	lockID   string
}

// removeExpiredUnits removes units which are expired from the resource
func (r *matchedResource) removeExpiredUnits() {
	var firstActive int
	for _, rID := range r.resource.TTLIdx {
		if ru, has := r.resource.Usages[rID]; has && ru.IsActive(time.Now()) {
			break
		}
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	for _, rID := range r.resource.TTLIdx[:firstActive] {
		if _, has := r.resource.Usages[rID]; !has {
			continue
		}
		delete(r.resource.Usages, rID)
	}
	r.resource.TTLIdx = r.resource.TTLIdx[firstActive:]
}

// recordUsage records a new usage
func (r *matchedResource) recordUsage(ru *utils.ResourceUsage) error {
	if _, hasID := r.resource.Usages[ru.ID]; hasID {
		return fmt.Errorf("duplicate resource usage with id: %s", ru.TenantID())
	}
	if r.ttl != nil && *r.ttl != -1 {
		if *r.ttl == 0 {
			return nil // no recording for ttl of 0
		}
		ru = ru.Clone() // don't influence the initial ru
		ru.ExpiryTime = time.Now().Add(*r.ttl)
	}
	r.resource.Usages[ru.ID] = ru
	if !ru.ExpiryTime.IsZero() {
		r.resource.TTLIdx = append(r.resource.TTLIdx, ru.ID)
	}
	return nil
}

// clearUsage clears the usage for an ID
func (r *matchedResource) clearUsage(ruID string) error {
	ru, hasIt := r.resource.Usages[ruID]
	if !hasIt {
		return fmt.Errorf("cannot find usage record with id: %s", ruID)
	}
	if !ru.ExpiryTime.IsZero() {
		for i, ruIDIdx := range r.resource.TTLIdx {
			if ruIDIdx == ruID {
				r.resource.TTLIdx = append(r.resource.TTLIdx[:i], r.resource.TTLIdx[i+1:]...)
				break
			}
		}
	}
	delete(r.resource.Usages, ruID)
	return nil
}

type matchedResources []*matchedResource

// recordUsage will record the usage in all the resource limits, failing back on errors
func (rs matchedResources) recordUsage(ru *utils.ResourceUsage) error {
	var nonReservedIdx int // index of first resource not reserved
	var err error
	for _, r := range rs {
		if err = r.recordUsage(ru); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s>cannot record usage, err: %v", utils.ResourceS, err))
			break
		}
		nonReservedIdx++
	}
	if err != nil {
		for _, r := range rs[:nonReservedIdx] {
			if errClear := r.clearUsage(ru.ID); errClear != nil {
				utils.Logger.Warning(fmt.Sprintf("<%s> cannot clear usage, err: %v", utils.ResourceS, errClear))
			} // best effort
		}
	}
	return err
}

// clearUsage gives back the units to the pool
func (rs matchedResources) clearUsage(ruTntID string) error {
	var err error
	for _, r := range rs {
		if errClear := r.clearUsage(ruTntID); errClear != nil &&
			r.ttl != nil && *r.ttl != 0 { // we only consider not found error in case of ttl different than 0
			utils.Logger.Warning(fmt.Sprintf("<%s>, clear ruID: %s, err: %v", utils.ResourceS, ruTntID, errClear))
			err = errClear
		}
	}
	return err
}

// allocateResource attempts allocating resources for a *ResourceUsage
// simulates on dryRun
// returns utils.ErrResourceUnavailable if allocation is not possible
func (rs matchedResources) allocateResource(ru *utils.ResourceUsage, dryRun bool) (allocMsg string, err error) {
	if len(rs) == 0 {
		return "", utils.ErrResourceUnavailable
	}
	for _, r := range rs {
		r.removeExpiredUnits()
		if _, hasID := r.resource.Usages[ru.ID]; hasID && !dryRun { // update
			_ = r.clearUsage(ru.ID) // can't fail: we just checked hasID
		}
		if allocMsg == "" &&
			(r.profile.Limit >= r.resource.TotalUsage()+ru.Units || r.profile.Limit == -1) {
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

// NewResourceService returns a new ResourceService
func NewResourceService(cfg *config.CGRConfig, dm *engine.DataManager,
	cache *engine.CacheS, filters *engine.FilterS, cm *engine.ConnManager) *ResourceS {
	return &ResourceS{
		cfg:             cfg,
		dm:              dm,
		cache:           cache,
		filters:         filters,
		cm:              cm,
		storedResources: make(utils.StringSet),
		stopBackup:      make(chan struct{}),
	}
}

// ResourceS is the service handling resources
type ResourceS struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	cache   *engine.CacheS
	filters *engine.FilterS
	cm      *engine.ConnManager

	storedMu        sync.Mutex
	storedResources utils.StringSet // resources that need saving

	stateMu    sync.Mutex // guards stopBackup
	stopBackup chan struct{}
	backupLoop sync.WaitGroup
}

// Reload restarts the backup loop. No-op after Shutdown.
func (s *ResourceS) Reload(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = make(chan struct{})
	s.StartLoop(ctx)
}

// StartLoop starts the goroutine with the backup loop
func (s *ResourceS) StartLoop(ctx *context.Context) {
	s.backupLoop.Add(1)
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *ResourceS) Shutdown(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = nil
	s.storeResources(ctx)
}

// backup will regularly store resources changed to DB
func (s *ResourceS) runBackup(ctx *context.Context) {
	defer s.backupLoop.Done()
	storeInterval := s.cfg.ResourceSCfg().StoreInterval
	if storeInterval <= 0 {
		return
	}
	for {
		s.storeResources(ctx)
		select {
		case <-s.stopBackup:
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeResources represents one task of complete backup
func (s *ResourceS) storeResources(ctx *context.Context) {
	var failedRIDs []string
	for { // don't stop until we store all pending resources
		s.storedMu.Lock()
		rID := s.storedResources.GetOne()
		if rID != "" {
			s.storedResources.Remove(rID)
		}
		s.storedMu.Unlock()
		if rID == "" {
			break // no more keys, backup completed
		}
		rIf, ok := s.cache.Get(utils.CacheResources, rID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> failed retrieving from cache resource with ID: %s", utils.ResourceS, rID))
			continue
		}
		r := rIf.(*utils.Resource)
		lkID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(r.Tenant, r.ID))
		if err := s.storeResource(ctx, r); err != nil {
			failedRIDs = append(failedRIDs, rID) // record failure so we can schedule it for next backup
		}
		guardian.Guardian.UnguardIDs(lkID)
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedMu.Lock()
		s.storedResources.AddSlice(failedRIDs)
		s.storedMu.Unlock()
	}
}

// storeResource stores the resource in DB and updates cache.
func (s *ResourceS) storeResource(ctx *context.Context, r *utils.Resource) error {
	if err := s.dm.SetResource(ctx, r); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<%s> failed saving Resource with ID: %s, error: %v",
				utils.ResourceS, r.ID, err))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := r.TenantID(); s.cache.HasItem(utils.CacheResources, tntID) { // only cache if previously there
		if err := s.cache.Set(ctx, utils.CacheResources, tntID, r, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> failed caching Resource with ID: %s, error: %v",
					utils.ResourceS, tntID, err))
			return err
		}
	}
	return nil
}

// storeMatchedResources will store the list of resources based on the StoreInterval
func (s *ResourceS) storeMatchedResources(ctx *context.Context, mtcRLs matchedResources) error {
	storeInterval := s.cfg.ResourceSCfg().StoreInterval
	if storeInterval == 0 {
		return nil
	}
	if storeInterval > 0 {
		s.storedMu.Lock()
		defer s.storedMu.Unlock()
	}
	for _, r := range mtcRLs {
		if !r.profile.Stored {
			continue
		}
		if storeInterval > 0 {
			s.storedResources.Add(r.resource.TenantID())
			continue
		}
		if err := s.storeResource(ctx, r.resource); err != nil {
			return err
		}
	}
	return nil
}

// processThresholds will pass the event for resource to ThresholdS
func (s *ResourceS) processThresholds(ctx *context.Context, tnt string, ev *utils.CGREvent, rs matchedResources) error {
	threshConns, err := engine.GetConnIDs(ctx, s.cfg.ResourceSCfg().Conns, utils.MetaThresholds, tnt, ev.AsDataProvider(), nil, s.filters)
	if err != nil {
		return err
	}
	if len(threshConns) == 0 {
		return nil
	}
	opts := ev.APIOpts
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
			Tenant: r.resource.Tenant,
			ID:     utils.GenUUID(),
			Event: map[string]any{
				utils.EventType:  utils.ResourceUpdate,
				utils.ResourceID: r.resource.ID,
				utils.Usage:      r.resource.TotalUsage(),
			},
			APIOpts: opts,
		}
		var tIDs []string
		if err := s.cm.Call(ctx, threshConns,
			utils.ThresholdSv1ProcessEvent, thEv, &tIDs); err != nil &&
			(len(r.profile.ThresholdIDs) != 0 || err.Error() != utils.ErrNotFound.Error()) {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> error: %v processing event %+v with %s",
					utils.ResourceS, err, thEv, utils.ThresholdS))
			withErrs = true
		}
	}
	if withErrs {
		return utils.ErrPartiallyExecuted
	}
	return nil
}

// matchingResourcesForEvent returns ordered list of matching resources which are active by the time of the call
func (s *ResourceS) matchingResourcesForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evUUID string, usageTTL *time.Duration) (rs matchedResources, unlock func(), err error) {

	unlockAll := func() {
		for _, r := range rs {
			guardian.Guardian.UnguardIDs(r.lockID)
		}
	}

	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	var itemIDs []string
	if x, ok := s.cache.Get(utils.CacheEventResources, evUUID); ok {
		if x == nil {
			return nil, nil, utils.ErrNotFound
		}
		itemIDs = x.([]string)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				// TODO: Consider using RemoveWithoutReplicate instead, as
				// partitions with Replicate=true call ReplicateRemove in
				// onEvict by default.
				if errCh := s.cache.Remove(ctx, utils.CacheEventResources, evUUID,
					true, utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()

	} else { // select the resourceIDs out of dataDB
		rIDs, err := engine.MatchingItemIDsForEvent(ctx, evNm,
			s.cfg.ResourceSCfg().StringIndexedFields,
			s.cfg.ResourceSCfg().PrefixIndexedFields,
			s.cfg.ResourceSCfg().SuffixIndexedFields,
			s.cfg.ResourceSCfg().ExistsIndexedFields,
			s.cfg.ResourceSCfg().NotExistsIndexedFields,
			s.dm, utils.CacheResourceFilterIndexes, tnt,
			s.cfg.ResourceSCfg().IndexedSelects,
			s.cfg.ResourceSCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := s.cache.Set(ctx, utils.CacheEventResources, evUUID, nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, nil, errCh
				}
			}
			return nil, nil, err
		}
		// Lock items in sorted order to prevent AB-BA deadlock.
		itemIDs = slices.Sorted(maps.Keys(rIDs))
	}

	rs = make(matchedResources, 0, len(itemIDs))
	for _, id := range itemIDs {
		lkID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.ResourceLockKey(tnt, id))

		rp, err := s.dm.GetResourceProfile(ctx, tnt, id,
			true, true, utils.NonTransactional)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			if err == utils.ErrNotFound {
				continue
			}
			unlockAll()
			return nil, nil, err
		}

		pass, err := s.filters.Pass(ctx, tnt, rp.FilterIDs, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}
		if !pass {
			guardian.Guardian.UnguardIDs(lkID)
			continue
		}

		res, err := s.dm.GetResource(ctx, rp.Tenant, rp.ID, true, true, "")
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}

		weight, err := engine.WeightFromDynamics(ctx, rp.Weights, s.filters, tnt, evNm)
		if err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			unlockAll()
			return nil, nil, err
		}

		r := &matchedResource{
			resource: res,
			profile:  rp,
			weight:   weight,
			lockID:   lkID,
		}
		if usageTTL != nil {
			if *usageTTL != 0 {
				r.ttl = usageTTL
			}
		} else if rp.UsageTTL >= 0 {
			r.ttl = utils.DurationPointer(rp.UsageTTL)
		}
		rs = append(rs, r)
	}

	if len(rs) == 0 {
		unlockAll()
		return nil, nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(rs, func(a, b *matchedResource) int {
		return cmp.Compare(b.weight, a.weight)
	})

	for i, r := range rs {
		if r.profile.Blocker && i != len(rs)-1 { // blocker will stop processing and we are not at last index
			for _, dropped := range rs[i+1:] {
				guardian.Guardian.UnguardIDs(dropped.lockID)
			}
			rs = rs[:i+1]
			break
		}
	}
	if errCh := s.cache.Set(ctx, utils.CacheEventResources, evUUID, itemIDs, nil, true, ""); errCh != nil {
		utils.Logger.Warning(fmt.Sprintf("<%s> failed caching event resources: %v", utils.ResourceS, errCh))
	}
	return rs, unlockAll, nil
}
