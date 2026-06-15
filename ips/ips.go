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

package ips

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"net/netip"
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

// matchedIPAllocs holds IP allocations together with the profile and indexes
// computed while matching an event.
type matchedIPAllocs struct {
	allocs     *utils.IPAllocations
	profile    *utils.IPProfile
	poolRanges map[string]netip.Prefix          // parsed CIDR ranges by pool ID
	poolAllocs map[string]map[netip.Addr]string // allocated IP to allocation ID, by pool ID
	lockID     string
}

// newMatchedIPAllocs wraps allocs and builds the pool indexes from the profile.
func newMatchedIPAllocs(allocs *utils.IPAllocations, profile *utils.IPProfile) (*matchedIPAllocs, error) {
	m := &matchedIPAllocs{
		allocs:     allocs,
		profile:    profile,
		poolAllocs: make(map[string]map[netip.Addr]string),
	}
	for allocID, alloc := range allocs.Allocations {
		if _, has := m.poolAllocs[alloc.PoolID]; !has {
			m.poolAllocs[alloc.PoolID] = make(map[netip.Addr]string)
		}
		m.poolAllocs[alloc.PoolID][alloc.Address] = allocID
	}
	m.poolRanges = make(map[string]netip.Prefix, len(profile.Pools))
	for _, pool := range profile.Pools {
		prefix, err := netip.ParsePrefix(pool.Range)
		if err != nil {
			return nil, err
		}
		m.poolRanges[pool.ID] = prefix
	}
	return m, nil
}

// removeExpiredUnits removes expired allocations. It stops at the first active
// one since TTLIndex is ordered by expiration.
func (m *matchedIPAllocs) removeExpiredUnits() {
	expiredCount := 0
	for _, allocID := range m.allocs.TTLIndex {
		alloc, exists := m.allocs.Allocations[allocID]
		if exists && alloc.IsActive(m.profile.TTL) {
			break
		}
		if alloc != nil {
			if poolMap, hasPool := m.poolAllocs[alloc.PoolID]; hasPool {
				delete(poolMap, alloc.Address)
			}
		}
		delete(m.allocs.Allocations, allocID)
		expiredCount++
	}
	if expiredCount > 0 {
		m.allocs.TTLIndex = m.allocs.TTLIndex[expiredCount:]
	}
}

func (m *matchedIPAllocs) removeAllocFromTTLIndex(allocID string) {
	for i, alID := range m.allocs.TTLIndex {
		if alID == allocID {
			m.allocs.TTLIndex = slices.Delete(m.allocs.TTLIndex, i, i+1)
			break
		}
	}
}

func (m *matchedIPAllocs) releaseAllocation(allocID string) error {
	alloc, has := m.allocs.Allocations[allocID]
	if !has {
		return fmt.Errorf("cannot find allocation record with id: %s", allocID)
	}
	if poolMap, hasPool := m.poolAllocs[alloc.PoolID]; hasPool {
		delete(poolMap, alloc.Address)
	}
	if m.profile.TTL > 0 {
		m.removeAllocFromTTLIndex(allocID)
	}
	delete(m.allocs.Allocations, allocID)
	return nil
}

// clearAllocations clears the given IDs, or every allocation when allocIDs is
// empty. Either all the given IDs exist and are cleared, or none are.
func (m *matchedIPAllocs) clearAllocations(allocIDs []string) error {
	if len(allocIDs) == 0 {
		clear(m.allocs.Allocations)
		clear(m.poolAllocs)
		m.allocs.TTLIndex = m.allocs.TTLIndex[:0]
		return nil
	}
	var notFound []string
	for _, allocID := range allocIDs {
		if _, has := m.allocs.Allocations[allocID]; !has {
			notFound = append(notFound, allocID)
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("cannot find allocation records with ids: %v", notFound)
	}
	for _, allocID := range allocIDs {
		alloc := m.allocs.Allocations[allocID]
		if poolMap, hasPool := m.poolAllocs[alloc.PoolID]; hasPool {
			delete(poolMap, alloc.Address)
		}
		if m.profile.TTL > 0 {
			m.removeAllocFromTTLIndex(allocID)
		}
		delete(m.allocs.Allocations, allocID)
	}
	return nil
}

// allocateIPOnPool allocates an IP from the pool or refreshes an existing
// allocation. If dryRun is true, it checks availability without allocating.
func (m *matchedIPAllocs) allocateIPOnPool(allocID string, pool *utils.IPPool,
	dryRun bool) (*utils.AllocatedIP, error) {
	m.removeExpiredUnits()
	if poolAlloc, has := m.allocs.Allocations[allocID]; has && !dryRun {
		poolAlloc.Time = time.Now()
		if m.profile.TTL > 0 {
			m.removeAllocFromTTLIndex(allocID)
			m.allocs.TTLIndex = append(m.allocs.TTLIndex, allocID)
		}
		return &utils.AllocatedIP{
			ProfileID: m.allocs.ID,
			PoolID:    pool.ID,
			Message:   pool.Message,
			Address:   poolAlloc.Address,
		}, nil
	}
	poolRange := m.poolRanges[pool.ID]
	if !poolRange.IsSingleIP() {
		return nil, errors.New("only single IP Pools are supported for now")
	}
	addr := poolRange.Addr()
	if _, hasPool := m.poolAllocs[pool.ID]; hasPool {
		if alcID, inUse := m.poolAllocs[pool.ID][addr]; inUse {
			return nil, fmt.Errorf("allocation failed for pool %q, IP %q: %w (allocated to %q)",
				pool.ID, addr, utils.ErrIPAlreadyAllocated, alcID)
		}
	}
	allocIP := &utils.AllocatedIP{
		ProfileID: m.allocs.ID,
		PoolID:    pool.ID,
		Message:   pool.Message,
		Address:   addr,
	}
	if dryRun {
		return allocIP, nil
	}
	m.allocs.Allocations[allocID] = &utils.PoolAllocation{
		PoolID:  pool.ID,
		Address: addr,
		Time:    time.Now(),
	}
	if _, hasPool := m.poolAllocs[pool.ID]; !hasPool {
		m.poolAllocs[pool.ID] = make(map[netip.Addr]string)
	}
	m.poolAllocs[pool.ID][addr] = allocID
	if m.profile.TTL > 0 {
		m.allocs.TTLIndex = append(m.allocs.TTLIndex, allocID)
	}
	return allocIP, nil
}

// allocateFromPools attempts IP allocation across all pools in priority order.
// It continues to the next pool only when the current one returns
// ErrIPAlreadyAllocated, returning the first success or the last error.
func (m *matchedIPAllocs) allocateFromPools(allocID string, poolIDs []string,
	dryRun bool) (*utils.AllocatedIP, error) {
	var err error
	for _, poolID := range poolIDs {
		pool := findPoolByID(m.profile.Pools, poolID)
		if pool == nil {
			return nil, fmt.Errorf("pool %q: %w", poolID, utils.ErrNotFound)
		}
		var result *utils.AllocatedIP
		if result, err = m.allocateIPOnPool(allocID, pool, dryRun); err == nil {
			return result, nil
		}
		if !errors.Is(err, utils.ErrIPAlreadyAllocated) {
			return nil, err
		}
	}
	return nil, err
}

// IPs is the service handling IP allocations
type IPs struct {
	cfg     *config.CGRConfig
	dm      *engine.DataManager
	filters *engine.FilterS
	cm      *engine.ConnManager

	storedMu  sync.Mutex
	storedIPs utils.StringSet // IP allocations that need saving

	stateMu    sync.Mutex // guards stopBackup
	stopBackup chan struct{}
	backupLoop sync.WaitGroup
}

// NewIPService returns a new IPs service
func NewIPService(cfg *config.CGRConfig, dm *engine.DataManager,
	filters *engine.FilterS, cm *engine.ConnManager) *IPs {
	return &IPs{
		cfg:        cfg,
		dm:         dm,
		filters:    filters,
		cm:         cm,
		storedIPs:  make(utils.StringSet),
		stopBackup: make(chan struct{}),
	}
}

// Reload restarts the backup loop. No-op after Shutdown.
func (s *IPs) Reload(ctx *context.Context) {
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
func (s *IPs) StartLoop(ctx *context.Context) {
	s.backupLoop.Add(1)
	go s.runBackup(ctx)
}

// Shutdown is called to shutdown the service
func (s *IPs) Shutdown(ctx *context.Context) {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.stopBackup == nil {
		return
	}
	close(s.stopBackup)
	s.backupLoop.Wait()
	s.stopBackup = nil
	s.storeIPAllocationsList(ctx)
}

// backup will regularly store IP allocations changed to DB
func (s *IPs) runBackup(ctx *context.Context) {
	defer s.backupLoop.Done()
	storeInterval := s.cfg.IPsCfg().StoreInterval
	if storeInterval <= 0 {
		return
	}
	for {
		s.storeIPAllocationsList(ctx)
		select {
		case <-s.stopBackup:
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeIPAllocationsList represents one task of complete backup
func (s *IPs) storeIPAllocationsList(ctx *context.Context) {
	var failedRIDs []string
	for { // don't stop until we store all dirty IP allocations
		s.storedMu.Lock()
		allocsID := s.storedIPs.GetOne()
		if allocsID != "" {
			s.storedIPs.Remove(allocsID)
		}
		s.storedMu.Unlock()
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
		lkID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.IPAllocationsLockKey(allocs.Tenant, allocs.ID))
		if err := s.storeIPAllocations(ctx, allocs); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.IPs, err))
			failedRIDs = append(failedRIDs, allocsID) // record failure so we can schedule it for next backup
		}
		guardian.Guardian.UnguardIDs(lkID)
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedRIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedMu.Lock()
		s.storedIPs.AddSlice(failedRIDs)
		s.storedMu.Unlock()
	}
}

// storeIPAllocations stores the IP allocations in DB.
func (s *IPs) storeIPAllocations(ctx *context.Context, allocs *utils.IPAllocations) error {
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
func (s *IPs) storeMatchedIPAllocations(ctx *context.Context, matched *utils.IPAllocations) error {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return nil
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedMu.Lock()
		s.storedIPs.Add(matched.TenantID())
		s.storedMu.Unlock()
		return nil
	}
	if err := s.storeIPAllocations(ctx, matched); err != nil {
		return err
	}
	return nil
}

// matchingIPAllocationsForEvent returns the IP allocation with the highest
// weight matching the event. Callers must release the lock via unlock.
func (s *IPs) matchingIPAllocationsForEvent(ctx *context.Context, tnt string,
	ev *utils.CGREvent, evUUID string) (matched *matchedIPAllocs, unlock func(), err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	var itemIDs []string
	if x, ok := engine.Cache.Get(utils.CacheEventIPs, evUUID); ok {
		if x == nil {
			return nil, nil, utils.ErrNotFound
		}
		itemIDs = []string{x.(string)}
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
	} else { // select the IP allocation IDs out of DB
		matchedItemIDs, err := engine.MatchingItemIDsForEvent(ctx, evNm,
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
					return nil, nil, errCh
				}
			}
			return nil, nil, err
		}
		itemIDs = slices.Sorted(maps.Keys(matchedItemIDs))
	}
	var matchedPrfl *utils.IPProfile
	var matchedLockID string
	var maxWeight float64
	for _, id := range itemIDs {
		lkID := guardian.Guardian.GuardIDs("",
			s.cfg.GeneralCfg().LockingTimeout,
			utils.IPAllocationsLockKey(tnt, id))
		var prfl *utils.IPProfile
		if prfl, err = s.dm.GetIPProfile(ctx, tnt, id, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			if err == utils.ErrNotFound {
				continue
			}
			if matchedPrfl != nil {
				guardian.Guardian.UnguardIDs(matchedLockID)
			}
			return nil, nil, err
		}
		var pass bool
		if pass, err = s.filters.Pass(ctx, tnt, prfl.FilterIDs, evNm); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			if matchedPrfl != nil {
				guardian.Guardian.UnguardIDs(matchedLockID)
			}
			return nil, nil, err
		} else if !pass {
			guardian.Guardian.UnguardIDs(lkID)
			continue
		}
		var weight float64
		if weight, err = engine.WeightFromDynamics(ctx, prfl.Weights, s.filters, tnt, evNm); err != nil {
			guardian.Guardian.UnguardIDs(lkID)
			if matchedPrfl != nil {
				guardian.Guardian.UnguardIDs(matchedLockID)
			}
			return nil, nil, err
		}
		if matchedPrfl == nil || maxWeight < weight {
			if matchedPrfl != nil {
				guardian.Guardian.UnguardIDs(matchedLockID)
			}
			matchedPrfl = prfl
			matchedLockID = lkID
			maxWeight = weight
		} else {
			guardian.Guardian.UnguardIDs(lkID)
		}
	}
	if matchedPrfl == nil {
		return nil, nil, utils.ErrNotFound
	}
	allocs, err := s.dm.GetIPAllocations(ctx, matchedPrfl.Tenant, matchedPrfl.ID, true, true, "")
	if err != nil {
		guardian.Guardian.UnguardIDs(matchedLockID)
		return nil, nil, err
	}
	if matched, err = newMatchedIPAllocs(allocs, matchedPrfl); err != nil {
		guardian.Guardian.UnguardIDs(matchedLockID)
		return nil, nil, err
	}
	matched.lockID = matchedLockID
	if err = engine.Cache.Set(ctx, utils.CacheEventIPs, evUUID, allocs.ID, nil, true, ""); err != nil {
		guardian.Guardian.UnguardIDs(matchedLockID)
		return nil, nil, err
	}
	return matched, func() { guardian.Guardian.UnguardIDs(matched.lockID) }, nil
}

func findPoolByID(pools []*utils.IPPool, id string) *utils.IPPool {
	for _, pool := range pools {
		if pool.ID == id {
			return pool
		}
	}
	return nil
}

// filterAndSortPools filters pools by their FilterIDs, sorts by weight
// (highest first), and truncates at the first blocking pool.
// TODO: check whether pre-allocating filteredPools & poolIDs improves
// performance or wastes memory when filtering is aggressive.
func filterAndSortPools(ctx *context.Context, tenant string, pools []*utils.IPPool,
	filters *engine.FilterS, ev utils.DataProvider) ([]string, error) {
	var filteredPools []*utils.IPPool
	weights := make(map[string]float64) // stores sorting weights by pool ID
	for _, pool := range pools {
		pass, err := filters.Pass(ctx, tenant, pool.FilterIDs, ev)
		if err != nil {
			return nil, err
		}
		if !pass {
			continue
		}
		weight, err := engine.WeightFromDynamics(ctx, pool.Weights, filters, tenant, ev)
		if err != nil {
			return nil, err
		}
		weights[pool.ID] = weight
		filteredPools = append(filteredPools, pool)
	}
	if len(filteredPools) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(filteredPools, func(a, b *utils.IPPool) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	var poolIDs []string
	for _, pool := range filteredPools {
		block, err := engine.BlockerFromDynamics(ctx, pool.Blockers, filters, tenant, ev)
		if err != nil {
			return nil, err
		}
		poolIDs = append(poolIDs, pool.ID)
		if block {
			break
		}
	}
	return poolIDs, nil
}
