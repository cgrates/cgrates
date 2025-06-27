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

package utils

import (
	"errors"
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"time"

	"github.com/cgrates/guardian"
)

// IPProfile defines the configuration of an IPAllocations object.
type IPProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Weights   DynamicWeights
	TTL       time.Duration
	Stored    bool
	Pools     []*IPPool

	lockID string // reference ID of lock used when matching the IPProfile
}

// IPProfileWithAPIOpts wraps IPProfile with APIOpts.
type IPProfileWithAPIOpts struct {
	*IPProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (p *IPProfile) TenantID() string {
	return ConcatenatedKey(p.Tenant, p.ID)
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
		Tenant:    p.Tenant,
		ID:        p.ID,
		FilterIDs: slices.Clone(p.FilterIDs),
		Weights:   p.Weights.Clone(),
		TTL:       p.TTL,
		Stored:    p.Stored,
		Pools:     pools,
		lockID:    p.lockID,
	}
}

// CacheClone returns a clone of IPProfile used by ltcache CacheCloner
func (p *IPProfile) CacheClone() any {
	return p.Clone()
}

func (p *IPProfile) Set(path []string, val any, newBranch bool) error {
	if len(path) != 1 && len(path) != 2 {
		return ErrWrongPath
	}
	var err error
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		p.Tenant = IfaceAsString(val)
	case ID:
		p.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		p.FilterIDs = append(p.FilterIDs, valA...)
	case Weights:
		if val != "" {
			p.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case TTL:
		p.TTL, err = IfaceAsDuration(val)
	case Stored:
		p.Stored, err = IfaceAsBool(val)
	case Pools:
		if len(path) != 2 {
			return ErrWrongPath
		}
		if val == "" {
			return nil
		}
		if len(p.Pools) == 0 || newBranch {
			p.Pools = append(p.Pools, new(IPPool))
		}
		pool := p.Pools[len(p.Pools)-1]
		return pool.Set(path[1:], val, newBranch)
	}
	return err
}

func (p *IPProfile) Merge(other any) {
	o := other.(*IPProfile)
	if len(o.Tenant) != 0 {
		p.Tenant = o.Tenant
	}
	if len(o.ID) != 0 {
		p.ID = o.ID
	}
	p.FilterIDs = append(p.FilterIDs, o.FilterIDs...)
	p.Weights = append(p.Weights, o.Weights...)
	if o.TTL != 0 {
		p.TTL = o.TTL
	}
	if o.Stored {
		p.Stored = o.Stored
	}
	for _, pool := range o.Pools {
		if idx := slices.IndexFunc(p.Pools, func(p *IPPool) bool {
			return p.ID == pool.ID
		}); idx != -1 {
			p.Pools[idx].Merge(pool)
			continue
		}
		p.Pools = append(p.Pools, pool)
	}
}

func (p *IPProfile) String() string { return ToJSON(p) }

func (p *IPProfile) FieldAsString(fldPath []string) (string, error) {
	val, err := p.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return IfaceAsString(val), nil
}

func (p *IPProfile) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case FilterIDs:
				if *idx < len(p.FilterIDs) {
					return p.FilterIDs[*idx], nil
				}
			case Pools:
				if *idx < len(p.Pools) {
					return p.Pools[*idx].FieldAsInterface(fldPath[1:])
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return p.Tenant, nil
	case ID:
		return p.ID, nil
	case FilterIDs:
		return p.FilterIDs, nil
	case Weights:
		return p.Weights, nil
	case TTL:
		return p.TTL, nil
	case Stored:
		return p.Stored, nil
	case Pools:
		return p.Pools, nil
	}
}

// Lock acquires a guardian lock on the IPProfile and stores the lock ID.
// Uses given lockID or creates a new lock.
func (p *IPProfile) Lock(lockID string) {
	if lockID == "" {
		lockID = guardian.Guardian.GuardIDs("",
			0, // TODO: find a way to pass timeout without importing config
			IPProfileLockKey(p.Tenant, p.ID))
	}
	p.lockID = lockID
}

// Unlock releases the lock on the IPProfile and clears the stored lock ID.
func (p *IPProfile) Unlock() {
	if p.lockID == "" {
		return
	}

	// Store current lock ID before clearing to prevent race conditions.
	id := p.lockID
	p.lockID = ""
	guardian.Guardian.UnguardIDs(id)
}

// IPProfileLockKey returns the ID used to lock an IPProfile with guardian.
func IPProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPProfiles, tnt, id)
}

// IPPool defines a pool of IP addresses within an IPProfile.
type IPPool struct {
	ID        string
	FilterIDs []string
	Type      string
	Range     string
	Strategy  string
	Message   string
	Weights   DynamicWeights
	Blockers  DynamicBlockers
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
		Weights:   p.Weights.Clone(),
		Blockers:  p.Blockers.Clone(),
	}
}

func (p *IPPool) Set(path []string, val any, _ bool) error {
	if len(path) != 1 {
		return ErrWrongPath
	}
	var err error
	switch path[0] {
	default:
		return ErrWrongPath
	case ID:
		p.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		p.FilterIDs = append(p.FilterIDs, valA...)
	case Type:
		p.Type = IfaceAsString(val)
	case Range:
		p.Range = IfaceAsString(val)
	case Strategy:
		p.Strategy = IfaceAsString(val)
	case Message:
		p.Message = IfaceAsString(val)
	case Weights:
		if val != "" {
			p.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case Blockers:
		if val != "" {
			p.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	}
	return err
}

func (p *IPPool) Merge(other any) {
	o := other.(*IPPool)

	// NOTE: Merge gets called when the IDs are equal, so this is a no-op.
	// Kept for consistency with other components.
	if len(o.ID) != 0 {
		p.ID = o.ID
	}

	p.FilterIDs = append(p.FilterIDs, o.FilterIDs...)
	if o.Type != "" {
		p.Type = o.Type
	}
	if o.Range != "" {
		p.Range = o.Range
	}
	if o.Strategy != "" {
		p.Strategy = o.Strategy
	}
	if o.Message != "" {
		p.Message = o.Message
	}
	p.Weights = append(p.Weights, o.Weights...)
	p.Blockers = append(p.Blockers, o.Blockers...)
}

func (p *IPPool) String() string { return ToJSON(p) }

func (p *IPPool) FieldAsString(fldPath []string) (string, error) {
	val, err := p.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return IfaceAsString(val), nil
}

func (p *IPPool) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) == 0 {
		return p, nil
	}
	if len(fldPath) > 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case FilterIDs:
				if *idx < len(p.FilterIDs) {
					return p.FilterIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case ID:
		return p.ID, nil
	case FilterIDs:
		return p.FilterIDs, nil
	case Type:
		return p.Type, nil
	case Range:
		return p.Range, nil
	case Strategy:
		return p.Strategy, nil
	case Message:
		return p.Message, nil
	case Weights:
		return p.Weights, nil
	case Blockers:
		return p.Blockers, nil
	}
}

// PoolAllocation represents one allocation in the pool.
type PoolAllocation struct {
	PoolID  string     // pool ID within the IPProfile
	Address netip.Addr // computed IP address
	Time    time.Time  // when this allocation was created
}

// IsActive checks if the allocation is still active.
func (a *PoolAllocation) IsActive(ttl time.Duration) bool {
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
	PoolID  string
	Message string
	Address netip.Addr
}

// IPAllocations represents IP allocations with usage tracking and TTL management.
type IPAllocations struct {
	Tenant      string
	ID          string
	Allocations map[string]*PoolAllocation // map[allocID]*PoolAllocation
	TTLIndex    []string                   // allocIDs ordered by allocation time for TTL expiry

	prfl       *IPProfile                       // cached profile configuration
	poolRanges map[string]netip.Prefix          // parsed CIDR ranges by pool ID
	poolAllocs map[string]map[netip.Addr]string // IP to allocation ID mapping by pool (map[poolID]map[Addr]allocID)
	lockID     string                           // guardian lock reference
}

// IPAllocationsWithAPIOpts wraps IPAllocations with APIOpts.
type IPAllocationsWithAPIOpts struct {
	*IPAllocations
	APIOpts map[string]any
}

// ComputeUnexported populates lookup maps and profile reference from exported fields.
// Must be called after retrieving from DB.
func (a *IPAllocations) ComputeUnexported(prfl *IPProfile) error {
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

// ReleaseAllocation releases the allocation for an ID.
func (a *IPAllocations) ReleaseAllocation(allocID string) error {
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

// AllocateIPOnPool allocates an IP from the specified pool or refreshes
// existing allocation. If dryRun is true, checks availability without
// allocating.
func (a *IPAllocations) AllocateIPOnPool(allocID string, pool *IPPool,
	dryRun bool) (*AllocatedIP, error) {
	a.removeExpiredUnits()
	if poolAlloc, has := a.Allocations[allocID]; has && !dryRun {
		poolAlloc.Time = time.Now()
		if a.prfl.TTL > 0 {
			a.removeAllocFromTTLIndex(allocID)
		}
		a.TTLIndex = append(a.TTLIndex, allocID)
		return &AllocatedIP{
			PoolID:  pool.ID,
			Message: pool.Message,
			Address: poolAlloc.Address,
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
				pool.ID, addr, ErrIPAlreadyAllocated, alcID)
		}
	}
	allocIP := &AllocatedIP{
		PoolID:  pool.ID,
		Message: pool.Message,
		Address: addr,
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
		if exists && alloc.IsActive(a.prfl.TTL) {
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

// Lock acquires a guardian lock on the IPAllocations and stores the lock ID.
// Uses given lockID (assumes already acquired) or creates a new lock.
func (a *IPAllocations) Lock(lockID string) {
	if lockID == "" {
		lockID = guardian.Guardian.GuardIDs("",
			0, // TODO: find a way to pass timeout without importing config
			IPAllocationsLockKey(a.Tenant, a.ID))
	}
	a.lockID = lockID
}

// Unlock releases the lock on the IPAllocations and clears the stored lock ID.
func (a *IPAllocations) Unlock() {
	if a.lockID == "" {
		return
	}

	// Store current lock ID before clearing to prevent race conditions.
	id := a.lockID
	a.lockID = ""
	guardian.Guardian.UnguardIDs(id)
}

// Config returns the IPAllocations' profile configuration.
func (a *IPAllocations) Config() *IPProfile {
	return a.prfl
}

// TenantID returns the unique ID in a multi-tenant environment
func (a *IPAllocations) TenantID() string {
	return ConcatenatedKey(a.Tenant, a.ID)
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
func IPAllocationsLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPAllocations, tnt, id)
}
