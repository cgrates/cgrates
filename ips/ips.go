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

package ips

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

// ipProfile wraps IPProfile with lkID.
type ipProfile struct {
	IPProfile *utils.IPProfile
	lkID      string // holds the reference towards guardian lock key
}

// lock will lock the ipProfile using guardian and store the lock within p.lkID
// if lkID is passed as argument, the lock is considered as executed
func (p *ipProfile) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPProfileLockKey(p.IPProfile.Tenant, p.IPProfile.ID))
	}
	p.lkID = lkID
}

// unlock will unlock the ipProfile and clear p.lkID.
func (p *ipProfile) unlock() {
	if p.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(p.lkID)
	p.lkID = utils.EmptyString
}

// isLocked returns the locks status of this ipProfile
func (p *ipProfile) isLocked() bool {
	return p.lkID != utils.EmptyString
}

// ipAllocations represents ipAllocations in the system
// not thread safe, needs locking at process level
type ipAllocations struct {
	IPAllocations *utils.IPAllocations
	lkID          string         // ID of the lock used when matching the ipAllocations
	ttl           *time.Duration // time to leave for these ip allocations, picked up on each IPAllocations initialization out of config
	tUsage        *float64       // sum of all usages
	dirty         *bool          // the usages were modified, needs save, *bool so we only save if enabled in config
	cfg           *ipProfile     // for ordering purposes
}

// lock will lock the ipAllocations using guardian and store the lock within ipAllocations.lkID
// if lkID is passed as argument, the lock is considered as executed
func (a *ipAllocations) lock(lkID string) {
	if lkID == utils.EmptyString {
		lkID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPAllocationsLockKey(a.IPAllocations.Tenant, a.IPAllocations.ID))
	}
	a.lkID = lkID
}

// unlock will unlock the ipAllocations and clear ipAllocations.lkID
func (a *ipAllocations) unlock() {
	if a.lkID == utils.EmptyString {
		return
	}
	guardian.Guardian.UnguardIDs(a.lkID)
	a.lkID = utils.EmptyString
}

// isLocked returns the locks status of this ipAllocations object.
func (a *ipAllocations) isLocked() bool {
	return a.lkID != utils.EmptyString
}

// removeExpiredUnits removes units which are expired from the ipAllocations object.
func (a *ipAllocations) removeExpiredUnits() {
	var firstActive int
	for _, usageID := range a.IPAllocations.TTLIdx {
		if u, has := a.IPAllocations.Usages[usageID]; has && u.IsActive(time.Now()) {
			break
		}
		firstActive++
	}
	if firstActive == 0 {
		return
	}
	for _, uID := range a.IPAllocations.TTLIdx[:firstActive] {
		usage, has := a.IPAllocations.Usages[uID]
		if !has {
			continue
		}
		delete(a.IPAllocations.Usages, uID)
		if a.tUsage != nil { //  total usage was not yet calculated so we do not need to update it
			*a.tUsage -= usage.Units
			if *a.tUsage < 0 { // something went wrong
				utils.Logger.Warning(
					fmt.Sprintf("resetting total usage for IP allocations %q, usage smaller than 0: %f", a.IPAllocations.ID, *a.tUsage))
				a.tUsage = nil
			}
		}
	}
	a.IPAllocations.TTLIdx = a.IPAllocations.TTLIdx[firstActive:]
	a.tUsage = nil
}

// recordUsage records a new usage
func (a *ipAllocations) recordUsage(usage *utils.IPUsage) error {
	if _, has := a.IPAllocations.Usages[usage.ID]; has {
		return fmt.Errorf("duplicate ip usage with id: %s", usage.TenantID())
	}
	if a.ttl != nil && *a.ttl != -1 {
		if *a.ttl == 0 {
			return nil // no recording for ttl of 0
		}
		usage = usage.Clone() // don't influence the initial ru
		usage.ExpiryTime = time.Now().Add(*a.ttl)
	}
	a.IPAllocations.Usages[usage.ID] = usage
	if a.tUsage != nil {
		*a.tUsage += usage.Units
	}
	if !usage.ExpiryTime.IsZero() {
		a.IPAllocations.TTLIdx = append(a.IPAllocations.TTLIdx, usage.ID)
	}
	return nil
}

// clearUsage clears the usage for an ID
func (a *ipAllocations) clearUsage(usageID string) error {
	usage, has := a.IPAllocations.Usages[usageID]
	if !has {
		return fmt.Errorf("cannot find usage record with id: %s", usageID)
	}
	if !usage.ExpiryTime.IsZero() {
		for i, uIDIdx := range a.IPAllocations.TTLIdx {
			if uIDIdx == usageID {
				a.IPAllocations.TTLIdx = slices.Delete(a.IPAllocations.TTLIdx, i, i+1)
				break
			}
		}
	}
	if a.tUsage != nil {
		*a.tUsage -= usage.Units
	}
	delete(a.IPAllocations.Usages, usageID)
	return nil
}

// IPAllocationsList is a collection of ipAllocations objects.
type IPAllocationsList []*ipAllocations

// unlock will unlock IP allocations in this slice
func (al IPAllocationsList) unlock() {
	for _, allocs := range al {
		allocs.unlock()
		if allocs.cfg != nil {
			allocs.cfg.unlock()
		}
	}
}

// ids returns a map of IP allocation IDs which is used for caching
func (al IPAllocationsList) ids() utils.StringSet {
	ids := make(utils.StringSet)
	for _, allocs := range al {
		ids.Add(allocs.IPAllocations.ID)
	}
	return ids
}

// IPService is the service handling IP allocations
type IPService struct {
	dm           *engine.DataManager // So we can load the data in cache and index it
	fltrs        *engine.FilterS
	storedIPsMux sync.RWMutex    // protects storedIPs
	storedIPs    utils.StringSet // keep a record of IP allocations which need saving, map[allocsID]bool
	cfg          *config.CGRConfig
	stopBackup   chan struct{} // control storing process
	loopStopped  chan struct{}
	cm           *engine.ConnManager
}

// NewIPService  returns a new IPService
func NewIPService(dm *engine.DataManager, cfg *config.CGRConfig,
	fltrs *engine.FilterS, cm *engine.ConnManager) *IPService {
	return &IPService{dm: dm,
		storedIPs:   make(utils.StringSet),
		cfg:         cfg,
		cm:          cm,
		fltrs:       fltrs,
		loopStopped: make(chan struct{}),
		stopBackup:  make(chan struct{}),
	}

}

// Reload stops the backupLoop and restarts it
func (s *IPService) Reload(ctx *context.Context) {
	close(s.stopBackup)
	<-s.loopStopped // wait until the loop is done
	s.stopBackup = make(chan struct{})
	go s.runBackup(ctx)
}

// StartLoop starts the gorutine with the backup loop
func (s *IPService) StartLoop(ctx *context.Context) {
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *IPService) Shutdown(ctx *context.Context) {
	close(s.stopBackup)
	s.storeIPAllocationsList(ctx)
}

// backup will regularly store IP allocations changed to dataDB
func (s *IPService) runBackup(ctx *context.Context) {
	storeInterval := s.cfg.IPsCfg().StoreInterval
	if storeInterval <= 0 {
		s.loopStopped <- struct{}{}
		return
	}
	for {
		s.storeIPAllocationsList(ctx)
		select {
		case <-s.stopBackup:
			s.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeIPAllocationsList represents one task of complete backup
func (s *IPService) storeIPAllocationsList(ctx *context.Context) {
	var failedRIDs []string
	for { // don't stop until we store all dirty IP allocations
		s.storedIPsMux.Lock()
		allocsID := s.storedIPs.GetOne()
		if allocsID != "" {
			s.storedIPs.Remove(allocsID)
		}
		s.storedIPsMux.Unlock()
		if allocsID == "" {
			break // no more keys, backup completed
		}
		rIf, ok := engine.Cache.Get(utils.CacheIPAllocations, allocsID)
		if !ok || rIf == nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> failed retrieving from cache IP allocations with ID %q", utils.IPs, allocsID))
			continue
		}
		allocs := &ipAllocations{
			IPAllocations: rIf.(*utils.IPAllocations),

			// NOTE: dirty is hardcoded to true, otherwise IP allocations would
			// never be stored.
			// Previously, dirty was part of the cached resource.
			dirty: utils.BoolPointer(true),
		}
		allocs.lock(utils.EmptyString)
		if err := s.storeIPAllocations(ctx, allocs); err != nil {
			failedRIDs = append(failedRIDs, allocsID) // record failure so we can schedule it for next backup
		}
		allocs.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedIPsMux.Lock()
		s.storedIPs.AddSlice(failedRIDs)
		s.storedIPsMux.Unlock()
	}
}

// storeIPAllocations stores the IP allocations in DB and corrects dirty flag.
func (s *IPService) storeIPAllocations(ctx *context.Context, allocs *ipAllocations) error {
	if allocs.dirty == nil || !*allocs.dirty {
		return nil
	}
	if err := s.dm.SetIPAllocations(ctx, allocs.IPAllocations); err != nil {
		utils.Logger.Warning(
			fmt.Sprintf("<IPs> failed saving IP allocations %q: %v",
				allocs.IPAllocations.ID, err))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := allocs.IPAllocations.TenantID(); engine.Cache.HasItem(utils.CacheIPAllocations, tntID) { // only cache if previously there
		if err := engine.Cache.Set(ctx, utils.CacheIPAllocations, tntID, allocs.IPAllocations, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(
				fmt.Sprintf("<IPs> failed caching IP allocations %q: %v", tntID, err))
			return err
		}
	}
	*allocs.dirty = false
	return nil
}

// storeMatchedIPAllocations will store the list of IP allocations based on the StoreInterval
func (s *IPService) storeMatchedIPAllocations(ctx *context.Context, matched IPAllocationsList) (err error) {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedIPsMux.Lock()
		defer s.storedIPsMux.Unlock()
	}
	for _, allocs := range matched {
		if allocs.dirty != nil {
			*allocs.dirty = true // mark it to be saved
			if s.cfg.IPsCfg().StoreInterval > 0 {
				s.storedIPs.Add(allocs.IPAllocations.TenantID())
				continue
			}
			if err = s.storeIPAllocations(ctx, allocs); err != nil {
				return
			}
		}

	}
	return
}

// matchingIPAllocationsForEvent returns ordered list of matching IP allocations which are active by the time of the call
func (s *IPService) matchingIPAllocationsForEvent(ctx *context.Context, tnt string, ev *utils.CGREvent,
	evUUID string, ttl *time.Duration) (al IPAllocationsList, err error) {
	var rIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := engine.Cache.Get(utils.CacheEventIPs, evUUID); ok { // The IPIDs were cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		rIDs = x.(utils.StringSet)
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				// TODO: Consider using RemoveWithoutReplicate instead, as
				// partitions with Replicate=true call ReplicateRemove in
				// onEvict by default.
				if errCh := engine.Cache.Remove(ctx, utils.CacheEventIPs, evUUID,
					true, utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()
	} else { // select the IP allocation IDs out of dataDB
		rIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
			s.cfg.IPsCfg().StringIndexedFields,
			s.cfg.IPsCfg().PrefixIndexedFields,
			s.cfg.IPsCfg().SuffixIndexedFields,
			s.cfg.IPsCfg().ExistsIndexedFields,
			s.cfg.IPsCfg().NotExistsIndexedFields,
			s.dm, utils.CacheIPFilterIndexes, tnt,
			s.cfg.IPsCfg().IndexedSelects,
			s.cfg.IPsCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID, nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, errCh
				}
			}
			return
		}
	}
	al = make(IPAllocationsList, 0, len(rIDs))
	weights := make(map[string]float64) // stores sorting weights by IP allocation ID
	for resName := range rIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPProfileLockKey(tnt, resName))
		var rp *utils.IPProfile
		if rp, err = s.dm.GetIPProfile(ctx, tnt, resName,
			true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			al.unlock()
			return
		}
		rPrf := &ipProfile{
			IPProfile: rp,
		}
		rPrf.lock(lkPrflID)
		var pass bool
		if pass, err = s.fltrs.Pass(ctx, tnt, rPrf.IPProfile.FilterIDs,
			evNm); err != nil {
			rPrf.unlock()
			al.unlock()
			return nil, err
		} else if !pass {
			rPrf.unlock()
			continue
		}
		lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPAllocationsLockKey(rPrf.IPProfile.Tenant, rPrf.IPProfile.ID))
		var res *utils.IPAllocations
		if res, err = s.dm.GetIPAllocations(ctx, rPrf.IPProfile.Tenant, rPrf.IPProfile.ID, true, true, ""); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			rPrf.unlock()
			al.unlock()
			return nil, err
		}
		allocs := &ipAllocations{
			IPAllocations: res,
		}
		allocs.lock(lkID) // pass the lock into IP allocations so we have it as reference
		if rPrf.IPProfile.Stored && allocs.dirty == nil {
			allocs.dirty = utils.BoolPointer(false)
		}
		if ttl != nil {
			if *ttl != 0 {
				allocs.ttl = ttl
			}
		} else if rPrf.IPProfile.TTL >= 0 {
			allocs.ttl = utils.DurationPointer(rPrf.IPProfile.TTL)
		}
		allocs.cfg = rPrf
		weight, err := engine.WeightFromDynamics(ctx, rPrf.IPProfile.Weights, s.fltrs, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[allocs.IPAllocations.ID] = weight
		al = append(al, allocs)
	}

	if len(al) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(al, func(a, b *ipAllocations) int {
		return cmp.Compare(weights[b.IPAllocations.ID], weights[a.IPAllocations.ID])
	})

	if err = engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID, al.ids(), nil, true, ""); err != nil {
		al.unlock()
	}
	return
}
