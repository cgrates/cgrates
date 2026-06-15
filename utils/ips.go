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

package utils

import (
	"net/netip"
	"slices"
	"time"
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
	if len(fldPath) == 0 {
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
	ProfileID string
	PoolID    string
	Message   string
	Address   netip.Addr
}

// AsNavigableMap implements engine.NavigableMapper.
func (ip *AllocatedIP) AsNavigableMap() map[string]*DataNode {
	return map[string]*DataNode{
		ProfileID: NewLeafNode(ip.ProfileID),
		PoolID:    NewLeafNode(ip.PoolID),
		Message:   NewLeafNode(ip.Message),
		Address:   NewLeafNode(ip.Address.String()),
	}
}

// IPAllocations represents IP allocations with usage tracking and TTL management.
type IPAllocations struct {
	Tenant      string
	ID          string
	Allocations map[string]*PoolAllocation // map[allocID]*PoolAllocation
	TTLIndex    []string                   // allocIDs ordered by allocation time for TTL expiry
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

// AsMapStringInterface converts IPProfile struct to map[string]any
func (p *IPProfile) AsMapStringInterface() map[string]any {
	if p == nil {
		return nil
	}
	return map[string]any{
		Tenant:    p.Tenant,
		ID:        p.ID,
		FilterIDs: p.FilterIDs,
		Weights:   p.Weights,
		TTL:       p.TTL,
		Stored:    p.Stored,
		Pools:     p.Pools,
	}
}

// MapStringInterfaceToIPProfile converts map[string]any to IPProfile struct
func MapStringInterfaceToIPProfile(m map[string]any) (*IPProfile, error) {
	ipp := &IPProfile{}

	if v, ok := m[Tenant].(string); ok {
		ipp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		ipp.ID = v
	}
	ipp.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	ipp.Weights = InterfaceToDynamicWeights(m[Weights])
	if v, ok := m[TTL].(string); ok {
		if dur, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			ipp.TTL = dur
		}
	} else if v, ok := m[TTL].(float64); ok { // for -1 cases
		ipp.TTL = time.Duration(v)
	}
	if v, ok := m[Stored].(bool); ok {
		ipp.Stored = v
	}
	ipp.Pools = InterfaceToPools(m[Pools])
	return ipp, nil
}

// InterfaceToPools converts any to []*IPPool
func InterfaceToPools(v any) []*IPPool {
	if v == nil {
		return nil
	}
	if pools, ok := v.([]any); ok {
		ipPools := make([]*IPPool, 0, len(pools))
		for _, p := range pools {
			pm, ok := p.(map[string]any)
			if !ok {
				break
			}
			pool := &IPPool{}
			if v, ok := pm[ID].(string); ok {
				pool.ID = v
			}

			pool.FilterIDs = InterfaceToStringSlice(pm[FilterIDs])
			if v, ok := pm[Type].(string); ok {
				pool.Type = v
			}
			if v, ok := pm[Range].(string); ok {
				pool.Range = v
			}
			if v, ok := pm[Strategy].(string); ok {
				pool.Strategy = v
			}
			if v, ok := pm[Message].(string); ok {
				pool.Message = v
			}
			pool.Weights = InterfaceToDynamicWeights(pm[Weights])
			pool.Blockers = InterfaceToDynamicBlockers(pm[Blockers])
			ipPools = append(ipPools, pool)
		}
		return ipPools
	}
	return nil
}

// TenantID returns the unique ID in a multi-tenant environment
func (a *IPAllocations) TenantID() string {
	return ConcatenatedKey(a.Tenant, a.ID)
}

// CacheClone returns a clone of IPAllocations object used by ltcache CacheCloner.
func (a *IPAllocations) CacheClone() any {
	return a.Clone()
}

// Clone creates a deep clone of the IPAllocations object.
func (a *IPAllocations) Clone() *IPAllocations {
	if a == nil {
		return nil
	}
	clone := &IPAllocations{
		Tenant:   a.Tenant,
		ID:       a.ID,
		TTLIndex: slices.Clone(a.TTLIndex),
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

// AsMapStringInterface converts IPAllocations struct to map[string]any
func (p *IPAllocations) AsMapStringInterface() map[string]any {
	if p == nil {
		return nil
	}
	return map[string]any{
		Tenant:      p.Tenant,
		ID:          p.ID,
		Allocations: p.Allocations,
		TTLIndex:    p.TTLIndex,
	}
}

// MapStringInterfaceToIPAllocations converts map[string]any to IPAllocations struct
func MapStringInterfaceToIPAllocations(m map[string]any) *IPAllocations {
	ipa := &IPAllocations{}

	if v, ok := m[Tenant].(string); ok {
		ipa.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		ipa.ID = v
	}
	ipa.Allocations = InterfaceToAllocations(m[Allocations])
	ipa.TTLIndex = InterfaceToStringSlice(m[TTLIndex])
	return ipa
}

// InterfaceToAllocations converts any to map[string]*PoolAllocation
func InterfaceToAllocations(v any) map[string]*PoolAllocation {
	if v == nil {
		return nil
	}
	if allocs, ok := v.(map[string]any); ok {
		ipAllocs := make(map[string]*PoolAllocation)
		for allocID, val := range allocs {
			allocMap, ok := val.(map[string]any)
			if !ok {
				break
			}
			allocation := &PoolAllocation{}
			if v, ok := allocMap[PoolID].(string); ok {
				allocation.PoolID = v
			}
			if v, ok := allocMap[Address].(string); ok {
				if addr, err := netip.ParseAddr(v); err == nil {
					allocation.Address = addr
				}
			}
			if v, ok := allocMap[Time].(string); ok {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					allocation.Time = t
				}
			}
			ipAllocs[allocID] = allocation
		}
		return ipAllocs
	}
	return nil
}
