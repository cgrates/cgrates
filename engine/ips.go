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
	"github.com/cgrates/cgrates/guardian"
	"github.com/cgrates/cgrates/utils"
)

// IPProfile defines the configuration of an IPAllocations object.
type IPProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval
	Weight             float64
	TTL                time.Duration
	Stored             bool
	Pools              []*IPPool

	lockID string // reference ID of lock used when matching the IPProfile
}

// IPProfileWithAPIOpts wraps IPProfile with APIOpts.
type IPProfileWithAPIOpts struct {
	*IPProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (p *IPProfile) TenantID() string {
	return utils.ConcatenatedKey(p.Tenant, p.ID)
}

// Clone creates a deep copy of IPProfile for thread-safe use.
func (p *IPProfile) Clone() *IPProfile {
	if p == nil {
		return nil
	}
	pools := make([]*IPPool, 0, len(p.Pools))
	for _, pool := range p.Pools {
		pools = append(pools, pool.Clone())
	}
	return &IPProfile{
		Tenant:             p.Tenant,
		ID:                 p.ID,
		FilterIDs:          slices.Clone(p.FilterIDs),
		ActivationInterval: p.ActivationInterval.Clone(),
		Weight:             p.Weight,
		TTL:                p.TTL,
		Stored:             p.Stored,
		Pools:              pools,
		lockID:             p.lockID,
	}
}

// CacheClone returns a clone of IPProfile used by ltcache CacheCloner
func (p *IPProfile) CacheClone() any {
	return p.Clone()
}

// Lock acquires a guardian lock on the IPProfile and stores the lock ID.
// Uses given lockID or creates a new lock.
func (p *IPProfile) lock(lockID string) {
	if lockID == "" {
		lockID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipProfileLockKey(p.Tenant, p.ID))
	}
	p.lockID = lockID
}

// Unlock releases the lock on the IPProfile and clears the stored lock ID.
func (p *IPProfile) unlock() {
	if p.lockID == "" {
		return
	}

	// Store current lock ID before clearing to prevent race conditions.
	id := p.lockID
	p.lockID = ""
	guardian.Guardian.UnguardIDs(id)
}

// IPProfileLockKey returns the ID used to lock an IPProfile with guardian.
func ipProfileLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheIPProfiles, tnt, id)
}

// IPPool defines a pool of IP addresses within an IPProfile.
type IPPool struct {
	ID        string
	FilterIDs []string
	Type      string
	Range     string
	Strategy  string
	Message   string
	Weight    float64
	Blocker   bool
}

// Clone creates a deep copy of Pool for thread-safe use.
func (p *IPPool) Clone() *IPPool {
	if p == nil {
		return nil
	}
	return &IPPool{
		ID:        p.ID,
		FilterIDs: slices.Clone(p.FilterIDs),
		Type:      p.Type,
		Range:     p.Range,
		Strategy:  p.Strategy,
		Message:   p.Message,
		Weight:    p.Weight,
		Blocker:   p.Blocker,
	}
}

// PoolAllocation represents one allocation in the pool.
type PoolAllocation struct {
	PoolID  string     // pool ID within the IPProfile
	Address netip.Addr // computed IP address
	Time    time.Time  // when this allocation was created
}

// IsActive checks if the allocation is still active.
func (a *PoolAllocation) isActive(ttl time.Duration) bool {
	return time.Now().Before(a.Time.Add(ttl))
}

// Clone creates a deep copy of the PoolAllocation object.
func (a *PoolAllocation) Clone() *PoolAllocation {
	if a == nil {
		return nil
	}
	clone := *a
	return &clone
}

// AllocatedIP represents one IP allocated on a pool, together with the message.
type AllocatedIP struct {
	ProfileID string
	PoolID    string
	Message   string
	Address   netip.Addr
}

// AsNavigableMap implements engine.NavigableMapper.
func (ip *AllocatedIP) AsNavigableMap() map[string]*utils.DataNode {
	return map[string]*utils.DataNode{
		utils.ProfileID: utils.NewLeafNode(ip.ProfileID),
		utils.PoolID:    utils.NewLeafNode(ip.PoolID),
		utils.Message:   utils.NewLeafNode(ip.Message),
		utils.Address:   utils.NewLeafNode(ip.Address.String()),
	}
}

// Digest returns a string representation of the allocated IP for digest replies.
func (ip *AllocatedIP) Digest() string {
	return utils.ConcatenatedKey(
		ip.ProfileID,
		ip.PoolID,
		ip.Message,
		ip.Address.String(),
	)
}

// IPAllocations represents IP allocations with usage tracking and TTL management.
type IPAllocations struct {
	Tenant      string
	ID          string
	Allocations map[string]*PoolAllocation // map[allocID]*PoolAllocation
	TTLIndex    []string                   // allocIDs ordered by allocation time for TTL expiry

	prfl       *IPProfile
	poolRanges map[string]netip.Prefix          // parsed CIDR ranges by pool ID
	poolAllocs map[string]map[netip.Addr]string // IP to allocation ID mapping by pool (map[poolID]map[Addr]allocID)
	lockID     string
}

// IPAllocationsWithAPIOpts wraps IPAllocations with APIOpts.
type IPAllocationsWithAPIOpts struct {
	*IPAllocations
	APIOpts map[string]any
}

// ClearIPAllocationsArgs contains arguments for clearing IP allocations.
// If AllocationIDs is empty or nil, all allocations will be cleared.
type ClearIPAllocationsArgs struct {
	Tenant        string
	ID            string
	AllocationIDs []string
	APIOpts       map[string]any
}

// computeUnexported sets up unexported fields based on the provided profile.
// Safe to call multiple times with the same profile.
func (a *IPAllocations) computeUnexported(prfl *IPProfile) error {
	if prfl == nil {
		return nil // nothing to compute without a profile
	}
	if a.prfl == prfl {
		return nil // already computed for this profile
	}
	a.prfl = prfl
	a.poolAllocs = make(map[string]map[netip.Addr]string)
	for allocID, alloc := range a.Allocations {
		if _, hasPool := a.poolAllocs[alloc.PoolID]; !hasPool {
			a.poolAllocs[alloc.PoolID] = make(map[netip.Addr]string)
		}
		a.poolAllocs[alloc.PoolID][alloc.Address] = allocID
	}
	a.poolRanges = make(map[string]netip.Prefix)
	for _, poolCfg := range a.prfl.Pools {
		prefix, err := netip.ParsePrefix(poolCfg.Range)
		if err != nil {
			return err
		}
		a.poolRanges[poolCfg.ID] = prefix
	}
	return nil
}

// releaseAllocation releases the allocation for an ID.
func (a *IPAllocations) releaseAllocation(allocID string) error {
	alloc, has := a.Allocations[allocID] // Get the allocation first
	if !has {
		return fmt.Errorf("cannot find allocation record with id: %s", allocID)
	}
	if poolMap, hasPool := a.poolAllocs[alloc.PoolID]; hasPool {
		delete(poolMap, alloc.Address)
	}
	if a.prfl.TTL > 0 {
		for i, refID := range a.TTLIndex {
			if refID == allocID {
				a.TTLIndex = slices.Delete(a.TTLIndex, i, i+1)
				break
			}
		}
	}
	delete(a.Allocations, allocID)
	return nil
}

// clearAllocations clears specified IP allocations or all allocations if allocIDs is empty/nil.
// Either all specified IDs exist and get cleared, or none are cleared and an error is returned.
func (a *IPAllocations) clearAllocations(allocIDs []string) error {
	if len(allocIDs) == 0 {
		clear(a.Allocations)
		clear(a.poolAllocs)
		a.TTLIndex = a.TTLIndex[:0] // maintain capacity
		return nil
	}

	// Validate all IDs exist before clearing any.
	var notFound []string
	for _, allocID := range allocIDs {
		if _, has := a.Allocations[allocID]; !has {
			notFound = append(notFound, allocID)
		}
	}
	if len(notFound) > 0 {
		return fmt.Errorf("cannot find allocation records with ids: %v", notFound)
	}

	for _, allocID := range allocIDs {
		alloc := a.Allocations[allocID]
		if poolMap, hasPool := a.poolAllocs[alloc.PoolID]; hasPool {
			delete(poolMap, alloc.Address)
		}
		if a.prfl.TTL > 0 {
			for i, refID := range a.TTLIndex {
				if refID == allocID {
					a.TTLIndex = slices.Delete(a.TTLIndex, i, i+1)
					break
				}
			}
		}
		delete(a.Allocations, allocID)
	}

	return nil
}

// allocateIPOnPool allocates an IP from the specified pool or refreshes
// existing allocation. If dryRun is true, checks availability without
// allocating.
func (a *IPAllocations) allocateIPOnPool(allocID string, pool *IPPool,
	dryRun bool) (*AllocatedIP, error) {
	a.removeExpiredUnits()
	if poolAlloc, has := a.Allocations[allocID]; has && !dryRun {
		poolAlloc.Time = time.Now()
		if a.prfl.TTL > 0 {
			a.removeAllocFromTTLIndex(allocID)
		}
		a.TTLIndex = append(a.TTLIndex, allocID)
		return &AllocatedIP{
			ProfileID: a.ID,
			PoolID:    pool.ID,
			Message:   pool.Message,
			Address:   poolAlloc.Address,
		}, nil
	}
	poolRange := a.poolRanges[pool.ID]
	if !poolRange.IsSingleIP() {
		return nil, errors.New("only single IP Pools are supported for now")
	}
	addr := poolRange.Addr()
	if _, hasPool := a.poolAllocs[pool.ID]; hasPool {
		if alcID, inUse := a.poolAllocs[pool.ID][addr]; inUse {
			return nil, fmt.Errorf("allocation failed for pool %q, IP %q: %w (allocated to %q)",
				pool.ID, addr, utils.ErrIPAlreadyAllocated, alcID)
		}
	}
	allocIP := &AllocatedIP{
		ProfileID: a.ID,
		PoolID:    pool.ID,
		Message:   pool.Message,
		Address:   addr,
	}
	if dryRun {
		return allocIP, nil
	}
	a.Allocations[allocID] = &PoolAllocation{
		PoolID:  pool.ID,
		Address: addr,
		Time:    time.Now(),
	}
	if _, hasPool := a.poolAllocs[pool.ID]; !hasPool {
		a.poolAllocs[pool.ID] = make(map[netip.Addr]string)
	}
	a.poolAllocs[pool.ID][addr] = allocID
	return allocIP, nil
}

// removeExpiredUnits removes expired allocations.
// It stops at first active since TTLIndex is sorted by expiration.
func (a *IPAllocations) removeExpiredUnits() {
	expiredCount := 0
	for _, allocID := range a.TTLIndex {
		alloc, exists := a.Allocations[allocID]
		if exists && alloc.isActive(a.prfl.TTL) {
			break
		}
		if alloc != nil {
			if poolMap, hasPool := a.poolAllocs[alloc.PoolID]; hasPool {
				delete(poolMap, alloc.Address)
			}
		}
		delete(a.Allocations, allocID)
		expiredCount++
	}
	if expiredCount > 0 {
		a.TTLIndex = a.TTLIndex[expiredCount:]
	}
}

// removeAllocFromTTLIndex removes an allocationID from TTL index.
func (a *IPAllocations) removeAllocFromTTLIndex(allocID string) {
	for i, alID := range a.TTLIndex {
		if alID == allocID {
			a.TTLIndex = slices.Delete(a.TTLIndex, i, i+1)
			break
		}
	}
}

// lock acquires a guardian lock on the IPAllocations and stores the lock ID.
// Uses given lockID (assumes already acquired) or creates a new lock.
func (a *IPAllocations) lock(lockID string) {
	if lockID == "" {
		lockID = guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipAllocationsLockKey(a.Tenant, a.ID))
	}
	a.lockID = lockID
}

// unlock releases the lock on the IPAllocations and clears the stored lock ID.
func (a *IPAllocations) unlock() {
	if a.lockID == "" {
		return
	}

	// Store current lock ID before clearing to prevent race conditions.
	id := a.lockID
	a.lockID = ""
	guardian.Guardian.UnguardIDs(id)

	if a.prfl != nil {
		a.prfl.unlock()
	}
}

// config returns the IPAllocations' profile configuration.
func (a *IPAllocations) config() *IPProfile {
	return a.prfl
}

// TenantID returns the unique ID in a multi-tenant environment
func (a *IPAllocations) TenantID() string {
	return utils.ConcatenatedKey(a.Tenant, a.ID)
}

// CacheClone returns a clone of IPAllocations object used by ltcache CacheCloner.
func (a *IPAllocations) CacheClone() any {
	return a.Clone()
}

// Clone creates a deep clone of the IPAllocations object (lockID excluded).
func (a *IPAllocations) Clone() *IPAllocations {
	if a == nil {
		return nil
	}
	clone := &IPAllocations{
		Tenant:     a.Tenant,
		ID:         a.ID,
		TTLIndex:   slices.Clone(a.TTLIndex),
		prfl:       a.prfl.Clone(),
		poolRanges: maps.Clone(a.poolRanges),
	}
	if a.poolAllocs != nil {
		clone.poolAllocs = make(map[string]map[netip.Addr]string)
		for poolID, allocs := range a.poolAllocs {
			clone.poolAllocs[poolID] = maps.Clone(allocs)
		}
	}
	if a.Allocations != nil {
		clone.Allocations = make(map[string]*PoolAllocation, len(a.Allocations))
		for id, alloc := range a.Allocations {
			clone.Allocations[id] = alloc.Clone()
		}
	}
	return clone
}

// IPAllocationsLockKey builds the guardian key for locking IP allocations.
func ipAllocationsLockKey(tnt, id string) string {
	return utils.ConcatenatedKey(utils.CacheIPAllocations, tnt, id)
}

// IPService is the service handling IP allocations
type IPService struct {
	cfg          *config.CGRConfig
	dm           *DataManager // So we can load the data in cache and index it
	cm           *ConnManager
	fltrs        *FilterS
	storedIPsMux sync.RWMutex    // protects storedIPs
	storedIPs    utils.StringSet // keep a record of IP allocations which need saving, map[allocsID]bool
	stopBackup   chan struct{}   // control storing process
	loopStopped  chan struct{}
}

// NewIPService returns a new IPService.
func NewIPService(dm *DataManager, cfg *config.CGRConfig, fltrs *FilterS,
	cm *ConnManager) *IPService {
	return &IPService{dm: dm,
		storedIPs:   make(utils.StringSet),
		cfg:         cfg,
		cm:          cm,
		fltrs:       fltrs,
		loopStopped: make(chan struct{}),
		stopBackup:  make(chan struct{}),
	}
}

// Reload restarts the backup loop.
func (s *IPService) Reload() {
	close(s.stopBackup)
	<-s.loopStopped // wait until the loop is done
	s.stopBackup = make(chan struct{})
	go s.runBackup()
}

// StartLoop starts the gorutine with the backup loop
func (s *IPService) StartLoop() {
	go s.runBackup()
}

// Shutdown is called to shutdown the service
func (s *IPService) Shutdown() {
	close(s.stopBackup)
	s.storeIPAllocationsList()
}

// runBackup will regularly update IP allocations stored in dataDB.
func (s *IPService) runBackup() {
	storeInterval := s.cfg.IPsCfg().StoreInterval
	if storeInterval <= 0 {
		s.loopStopped <- struct{}{}
		return
	}
	for {
		s.storeIPAllocationsList()
		select {
		case <-s.stopBackup:
			s.loopStopped <- struct{}{}
			return
		case <-time.After(storeInterval):
		}
	}
}

// storeIPAllocationsList represents one task of complete backup
func (s *IPService) storeIPAllocationsList() {
	var failedAllocIDs []string
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
		allocIf, ok := Cache.Get(utils.CacheIPAllocations, allocsID)
		if !ok || allocIf == nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<%s> failed retrieving from cache IP allocations with ID %q", utils.IPs, allocsID))
			continue
		}
		allocs := allocIf.(*IPAllocations)
		allocs.lock(utils.EmptyString)
		if err := s.storeIPAllocations(allocs); err != nil {
			utils.Logger.Warning(fmt.Sprintf("<%s> %v", utils.IPs, err))
			failedAllocIDs = append(failedAllocIDs, allocsID) // record failure so we can schedule it for next backup
		}
		allocs.unlock()
		// randomize the CPU load and give up thread control
		runtime.Gosched()
	}
	if len(failedAllocIDs) != 0 { // there were errors on save, schedule the keys for next backup
		s.storedIPsMux.Lock()
		s.storedIPs.AddSlice(failedAllocIDs)
		s.storedIPsMux.Unlock()
	}
}

// storeIPAllocations stores the IP allocations in DB.
func (s *IPService) storeIPAllocations(allocs *IPAllocations) error {
	if err := s.dm.SetIPAllocations(allocs); err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<IPs> could not save IP allocations %q: %v", allocs.ID, err))
		return err
	}
	//since we no longer handle cache in DataManager do here a manual caching
	if tntID := allocs.TenantID(); Cache.HasItem(utils.CacheIPAllocations, tntID) { // only cache if previously there
		if err := Cache.Set(utils.CacheIPAllocations, tntID, allocs, nil,
			true, utils.NonTransactional); err != nil {
			utils.Logger.Warning(fmt.Sprintf(
				"<IPs> could not cache IP allocations %q: %v", tntID, err))
			return err
		}
	}
	return nil
}

// storeMatchedIPAllocations will store the list of IP allocations based on the StoreInterval.
func (s *IPService) storeMatchedIPAllocations(matched *IPAllocations) error {
	if s.cfg.IPsCfg().StoreInterval == 0 {
		return nil
	}
	if s.cfg.IPsCfg().StoreInterval > 0 {
		s.storedIPsMux.Lock()
		s.storedIPs.Add(matched.TenantID())
		s.storedIPsMux.Unlock()
		return nil
	}
	if err := s.storeIPAllocations(matched); err != nil {
		return err
	}
	return nil
}

// matchingIPAllocationsForEvent returns the IP allocation with the highest weight
// matching the event.
func (s *IPService) matchingIPAllocationsForEvent(tnt string,
	ev *utils.CGREvent, evUUID string) (allocs *IPAllocations, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  ev.Event,
		utils.MetaOpts: ev.APIOpts,
	}
	var itemIDs []string
	if x, ok := Cache.Get(utils.CacheEventIPs, evUUID); ok {
		// IPIDs cached as utils.StringSet{"resID":bool}
		if x == nil {
			return nil, utils.ErrNotFound
		}
		itemIDs = []string{x.(string)}
		defer func() { // make sure we uncache if we find errors
			if err != nil {
				// TODO: Consider using RemoveWithoutReplicate instead, as
				// partitions with Replicate=true call ReplicateRemove in
				// onEvict by default.
				if errCh := Cache.Remove(utils.CacheEventIPs, evUUID,
					true, utils.NonTransactional); errCh != nil {
					err = errCh
				}
			}
		}()
	} else { // select the IP allocation IDs out of dataDB
		matchedItemIDs, err := MatchingItemIDsForEvent(evNm,
			s.cfg.IPsCfg().StringIndexedFields,
			s.cfg.IPsCfg().PrefixIndexedFields,
			s.cfg.IPsCfg().SuffixIndexedFields,
			s.cfg.IPsCfg().ExistsIndexedFields,
			s.dm, utils.CacheIPFilterIndexes, tnt,
			s.cfg.IPsCfg().IndexedSelects,
			s.cfg.IPsCfg().NestedFields,
		)
		if err != nil {
			if err == utils.ErrNotFound {
				if errCh := Cache.Set(utils.CacheEventIPs, evUUID,
					nil, nil, true, ""); errCh != nil { // cache negative match
					return nil, errCh
				}
			}
			return nil, err
		}
		itemIDs = slices.Sorted(maps.Keys(matchedItemIDs))
	}
	var matchedPrfl *IPProfile
	var maxWeight float64
	for _, id := range itemIDs {
		lkPrflID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout,
			ipProfileLockKey(tnt, id))
		var prfl *IPProfile
		if prfl, err = s.dm.GetIPProfile(tnt, id, true, true, utils.NonTransactional); err != nil {
			guardian.Guardian.UnguardIDs(lkPrflID)
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		prfl.lock(lkPrflID)
		if prfl.ActivationInterval != nil && ev.Time != nil &&
			!prfl.ActivationInterval.IsActiveAtTime(*ev.Time) { // not active
			prfl.unlock()
			continue
		}
		var pass bool
		if pass, err = s.fltrs.Pass(tnt, prfl.FilterIDs, evNm); err != nil {
			prfl.unlock()
			return nil, err
		} else if !pass {
			prfl.unlock()
			continue
		}
		if matchedPrfl == nil || maxWeight < prfl.Weight {
			if matchedPrfl != nil {
				matchedPrfl.unlock()
			}
			matchedPrfl = prfl
			maxWeight = prfl.Weight
		} else {
			prfl.unlock()
		}
	}
	if matchedPrfl == nil {
		return nil, utils.ErrNotFound
	}
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		ipAllocationsLockKey(matchedPrfl.Tenant, matchedPrfl.ID))
	allocs, err = s.dm.GetIPAllocations(matchedPrfl.Tenant, matchedPrfl.ID, true, true, "", matchedPrfl)
	if err != nil {
		guardian.Guardian.UnguardIDs(lkID)
		matchedPrfl.unlock()
		return nil, err
	}
	allocs.lock(lkID)
	if err = Cache.Set(utils.CacheEventIPs, evUUID, allocs.ID, nil, true, ""); err != nil {
		allocs.unlock()
	}
	return allocs, nil
}

// allocateFromPools attempts IP allocation across all pools in priority order.
// Continues to next pool only if current pool returns ErrIPAlreadyAllocated.
// Returns first successful allocation or the last allocation error.
func (s *IPService) allocateFromPools(allocs *IPAllocations, allocID string,
	poolIDs []string, dryRun bool) (*AllocatedIP, error) {
	var err error
	for _, poolID := range poolIDs {
		pool := findPoolByID(allocs.config().Pools, poolID)
		if pool == nil {
			return nil, fmt.Errorf("pool %q: %w", poolID, utils.ErrNotFound)
		}
		var result *AllocatedIP
		if result, err = allocs.allocateIPOnPool(allocID, pool, dryRun); err == nil {
			return result, nil
		}
		if !errors.Is(err, utils.ErrIPAlreadyAllocated) {
			return nil, err
		}
	}
	return nil, err
}

func findPoolByID(pools []*IPPool, id string) *IPPool {
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
func filterAndSortPools(tenant string, pools []*IPPool,
	fltrs *FilterS, ev utils.DataProvider) ([]string, error) {
	var filteredPools []*IPPool
	for _, pool := range pools {
		pass, err := fltrs.Pass(tenant, pool.FilterIDs, ev)
		if err != nil {
			return nil, err
		}
		if !pass {
			continue
		}
		filteredPools = append(filteredPools, pool)
	}
	if len(filteredPools) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(filteredPools, func(a, b *IPPool) int {
		return cmp.Compare(b.Weight, a.Weight)
	})

	var poolIDs []string
	for _, pool := range filteredPools {
		poolIDs = append(poolIDs, pool.ID)
		if pool.Blocker {
			break
		}
	}
	return poolIDs, nil
}

// V1GetIPAllocationForEvent returns the IPAllocations object matching the event.
func (s *IPService) V1GetIPAllocationForEvent(ctx *context.Context, args *utils.CGREvent, reply *IPAllocations) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	allocID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.AllocationID, utils.OptsIPsAllocationID)
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1GetIPAllocationForEvent, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*IPAllocations)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var allocs *IPAllocations
	if allocs, err = s.matchingIPAllocationsForEvent(tnt, args, allocID); err != nil {
		return err
	}
	defer allocs.unlock()
	*reply = *allocs
	return
}

// V1AuthorizeIP checks if it's able to allocate an IP address for the given event.
func (s *IPService) V1AuthorizeIP(ctx *context.Context, args *utils.CGREvent, reply *AllocatedIP) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	allocID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.AllocationID, utils.OptsIPsAllocationID)
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AuthorizeIP, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*AllocatedIP)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var allocs *IPAllocations
	if allocs, err = s.matchingIPAllocationsForEvent(tnt, args, allocID); err != nil {
		return err
	}
	defer allocs.unlock()

	var poolIDs []string
	if poolIDs, err = filterAndSortPools(tnt, allocs.config().Pools, s.fltrs,
		args.AsDataProvider()); err != nil {
		return err
	}

	var allocIP *AllocatedIP
	if allocIP, err = s.allocateFromPools(allocs, allocID, poolIDs, true); err != nil {
		if errors.Is(err, utils.ErrIPAlreadyAllocated) {
			return utils.ErrIPUnauthorized
		}
		return err
	}

	*reply = *allocIP
	return nil
}

// V1AllocateIP allocates an IP address for the given event.
func (s *IPService) V1AllocateIP(ctx *context.Context, args *utils.CGREvent, reply *AllocatedIP) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	allocID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.AllocationID, utils.OptsIPsAllocationID)
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1AllocateIP, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*AllocatedIP)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var allocs *IPAllocations
	if allocs, err = s.matchingIPAllocationsForEvent(tnt, args, allocID); err != nil {
		return err
	}
	defer allocs.unlock()

	var poolIDs []string
	if poolIDs, err = filterAndSortPools(tnt, allocs.config().Pools, s.fltrs,
		args.AsDataProvider()); err != nil {
		return err
	}

	var allocIP *AllocatedIP
	if allocIP, err = s.allocateFromPools(allocs, allocID, poolIDs, false); err != nil {
		return err
	}

	// index it for storing
	if err = s.storeMatchedIPAllocations(allocs); err != nil {
		return err
	}
	*reply = *allocIP
	return nil
}

// V1ReleaseIP releases an allocated IP address for the given event.
func (s *IPService) V1ReleaseIP(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Event}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	allocID := utils.GetStringOpts(args, s.cfg.IPsCfg().Opts.AllocationID, utils.OptsIPsAllocationID)
	if allocID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.AllocationID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// RPC caching
	if config.CgrConfig().CacheCfg().Partitions[utils.CacheRPCResponses].Limit != 0 {
		cacheKey := utils.ConcatenatedKey(utils.IPsV1ReleaseIP, utils.ConcatenatedKey(tnt, args.ID))
		refID := guardian.Guardian.GuardIDs("",
			config.CgrConfig().GeneralCfg().LockingTimeout, cacheKey) // RPC caching needs to be atomic
		defer guardian.Guardian.UnguardIDs(refID)
		if itm, has := Cache.Get(utils.CacheRPCResponses, cacheKey); has {
			cachedResp := itm.(*utils.CachedRPCResponse)
			if cachedResp.Error == nil {
				*reply = *cachedResp.Result.(*string)
			}
			return cachedResp.Error
		}
		defer Cache.Set(utils.CacheRPCResponses, cacheKey,
			&utils.CachedRPCResponse{Result: reply, Error: err},
			nil, true, utils.NonTransactional)
	}
	// end of RPC caching

	var allocs *IPAllocations
	if allocs, err = s.matchingIPAllocationsForEvent(tnt, args, allocID); err != nil {
		return err
	}
	defer allocs.unlock()

	if err = allocs.releaseAllocation(allocID); err != nil {
		utils.Logger.Warning(fmt.Sprintf(
			"<%s> failed to remove allocation from IPAllocations with ID %q: %v", utils.IPs, allocs.TenantID(), err))
	}

	// Handle storing
	if err = s.storeMatchedIPAllocations(allocs); err != nil {
		return err
	}

	*reply = utils.OK
	return nil
}

// V1GetIPAllocations returns all IP allocations for a tenantID.
func (s *IPService) V1GetIPAllocations(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *IPAllocations) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	// make sure resource is locked at process level
	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		ipAllocationsLockKey(tnt, arg.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	ip, err := s.dm.GetIPAllocations(tnt, arg.ID, true, true, utils.NonTransactional, nil)
	if err != nil {
		return err
	}
	*reply = *ip
	return nil
}

// V1ClearIPAllocations clears IP allocations from an IPAllocations object.
// If args.AllocationIDs is empty or nil, all allocations will be cleared.
func (s *IPService) V1ClearIPAllocations(ctx *context.Context, args *ClearIPAllocationsArgs, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = s.cfg.GeneralCfg().DefaultTenant
	}

	lkID := guardian.Guardian.GuardIDs(utils.EmptyString,
		config.CgrConfig().GeneralCfg().LockingTimeout,
		ipAllocationsLockKey(tnt, args.ID))
	defer guardian.Guardian.UnguardIDs(lkID)

	allocs, err := s.dm.GetIPAllocations(tnt, args.ID, true, true, utils.NonTransactional, nil)
	if err != nil {
		return err
	}
	if err := allocs.clearAllocations(args.AllocationIDs); err != nil {
		return err
	}
	if err := s.storeIPAllocations(allocs); err != nil {
		return err
	}

	*reply = utils.OK
	return nil
}
