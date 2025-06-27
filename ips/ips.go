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
	"errors"
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
		allocIf, ok := engine.Cache.Get(utils.CacheIPAllocations, allocsID)
		if !ok || allocIf == nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> failed retrieving from cache IP allocations with ID %q", utils.IPs, allocsID))
			continue
		}
		allocs := allocIf.(*utils.IPAllocations)
		allocs.Lock(utils.EmptyString)
		if err := s.storeIPAllocations(ctx, allocs); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.IPs, err))
			failedRIDs = append(failedRIDs, allocsID) // record failure so we can schedule it for next backup
		}
		allocs.Unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedIPsMux.Lock()
		s.storedIPs.AddSlice(failedRIDs)
		s.storedIPsMux.Unlock()
	}
}

// storeIPAllocations stores the IP allocations in DB.
func (s *IPService) storeIPAllocations(ctx *context.Context, allocs *utils.IPAllocations) error {
	if err := s.dm.SetIPAllocations(ctx, allocs); err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<IPs> could not save IP allocations %q: %v", allocs.ID, err))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := allocs.TenantID(); engine.Cache.HasItem(utils.CacheIPAllocations, tntID) { // only cache if previously there
		if err := engine.Cache.Set(ctx, utils.CacheIPAllocations, tntID, allocs, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<IPs> could not cache IP allocations %q: %v", tntID, err))
			return err
		}
	}
	return nil
}

// storeMatchedIPAllocations will store the list of IP allocations based on the StoreInterval
func (s *IPService) storeMatchedIPAllocations(ctx *context.Context, matched *utils.IPAllocations) error {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return nil
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedIPsMux.Lock()
		s.storedIPs.Add(matched.TenantID())
		s.storedIPsMux.Unlock()
		return nil
	}
	if err := s.storeIPAllocations(ctx, matched); err != nil {
		return err
	}
	return nil
}

// matchingIPAllocationsForEvent returns the IP allocation with the highest weight
// matching the event.
func (s *IPService) matchingIPAllocationsForEvent(ctx *context.Context, tnt string,
	ev *utils.CGREvent, evUUID string) (allocs *utils.IPAllocations, err error) {
	var itemIDs utils.StringSet
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	if x, ok := engine.Cache.Get(utils.CacheEventIPs, evUUID); ok {
		// IPIDs cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		itemIDs = x.(utils.StringSet)
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
		itemIDs, err = engine.MatchingItemIDsForEvent(ctx, evNm,
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
				if errCh := engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID,
					nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, errCh
				}
			}
			return nil, err
		}
	}
	var matchedPrfl *utils.IPProfile
	var maxWeight float64
	for id := range itemIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			utils.IPProfileLockKey(tnt, id))
		var prfl *utils.IPProfile
		if prfl, err = s.dm.GetIPProfile(ctx, tnt, id, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		prfl.Lock(lkPrflID)
		var pass bool
		if pass, err = s.fltrs.Pass(ctx, tnt, prfl.FilterIDs, evNm); err != nil {
			prfl.Unlock()
			return nil, err
		} else if !pass {
			prfl.Unlock()
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, prfl.Weights, s.fltrs, tnt, evNm); err != nil {
			prfl.Unlock()
			return nil, err
		}
		if matchedPrfl == nil || maxWeight < weight {
			if matchedPrfl != nil {
				matchedPrfl.Unlock()
			}
			matchedPrfl = prfl
			maxWeight = weight
		} else {
			prfl.Unlock()
		}
	}
	if matchedPrfl == nil {
		return nil, utils.ErrNotFound
	}
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		utils.IPAllocationsLockKey(matchedPrfl.Tenant, matchedPrfl.ID))
	allocs, err = s.dm.GetIPAllocations(ctx, matchedPrfl.Tenant, matchedPrfl.ID, true, true, "")
	if err != nil {
		guardian.Guardian.UnguardIDs(lkID)
		matchedPrfl.Unlock()
		return nil, err
	}
	allocs.Lock(lkID)

	// Clone profile to avoid modifying cached version during pool sorting.
	prfl := matchedPrfl.Clone()
	if err = sortPools(ctx, prfl, s.fltrs, evNm); err != nil {
		allocs.Unlock()
		return nil, err
	}
	if err = allocs.ComputeUnexported(prfl); err != nil {
		allocs.Unlock()
		return nil, err
	}

	if err = engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID,

		// TODO: check if we still should rely on caching previously matched
		// allocations for an allocationID. Currently setting a StringSet to
		// maintain previous functionality, but this doesn't seem right.
		utils.StringSet{allocs.ID: struct{}{}},

		nil, true, ""); err != nil {
		allocs.Unlock()
	}
	return allocs, nil
}

// allocateFromPools attempts IP allocation across all pools in priority order.
// Continues to next pool only if current pool returns ErrIPAlreadyAllocated.
// Returns first successful allocation or the last allocation error.
func (s *IPService) allocateFromPools(allocs *utils.IPAllocations, allocID string,
	dryRun bool) (*utils.AllocatedIP, error) {
	var err error
	for _, pool := range allocs.Config().Pools {
		var result *utils.AllocatedIP
		if result, err = allocs.AllocateIPOnPool(allocID, pool, dryRun); err == nil {
			return result, nil
		}
		if !errors.Is(err, utils.ErrIPAlreadyAllocated) {
			return nil, err
		}
	}
	return nil, err
}

// sortPools orders pools by weight (highest first) and truncates at first blocking pool.
func sortPools(ctx *context.Context, prfl *utils.IPProfile, fltrs *engine.FilterS,
	ev utils.DataProvider) error {
	weights := make(map[string]float64) // stores sorting weights by pool ID
	for _, pool := range prfl.Pools {
		weight, err := engine.WeightFromDynamics(ctx, pool.Weights, fltrs, prfl.Tenant, ev)
		if err != nil {
			return err
		}
		weights[pool.ID] = weight
	}

	// Sort by weight (higher values first).
	slices.SortFunc(prfl.Pools, func(a, b *utils.IPPool) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	for i, pool := range prfl.Pools {
		block, err := engine.BlockerFromDynamics(ctx, pool.Blockers, fltrs, prfl.Tenant, ev)
		if err != nil {
			return err
		}
		if block {
			prfl.Pools = prfl.Pools[0 : i+1]
			break
		}
	}
	return nil
}
