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
		if val == EmptyString {
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

// IPProfileLockKey returns the ID used to lock an IPProfile with guardian
func IPProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPProfiles, tnt, id)
}

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

// IPUsage represents a usage count.
type IPUsage struct {
	Tenant     string
	ID         string
	ExpiryTime time.Time
	Units      float64
}

// TenantID returns the concatenated key between tenant and ID.
func (u *IPUsage) TenantID() string {
	return ConcatenatedKey(u.Tenant, u.ID)
}

// IsActive checks ExpiryTime at some time
func (u *IPUsage) IsActive(atTime time.Time) bool {
	return u.ExpiryTime.IsZero() || u.ExpiryTime.Sub(atTime) > 0
}

// Clone duplicates the IPUsage
func (u *IPUsage) Clone() *IPUsage {
	if u == nil {
		return nil
	}
	clone := *u
	return &clone
}

// IPAllocations represents IP allocations with usage tracking and TTL management.
type IPAllocations struct {
	Tenant string
	ID     string
	Usages map[string]*IPUsage
	TTLIdx []string
}

// IPAllocationsWithAPIOpts wraps IPAllocations with APIOpts.
type IPAllocationsWithAPIOpts struct {
	*IPAllocations
	APIOpts map[string]any
}

// Clone clones IPAllocations object (lkID excluded)
func (a *IPAllocations) Clone() *IPAllocations {
	if a == nil {
		return nil
	}
	clone := &IPAllocations{
		Tenant: a.Tenant,
		ID:     a.ID,
		TTLIdx: slices.Clone(a.TTLIdx),
	}
	if a.Usages != nil {
		clone.Usages = make(map[string]*IPUsage, len(a.Usages))
		for key, usage := range a.Usages {
			clone.Usages[key] = usage.Clone()
		}
	}
	return clone
}

// CacheClone returns a clone of IPAllocations object used by ltcache CacheCloner.
func (a *IPAllocations) CacheClone() any {
	return a.Clone()
}

// TenantID returns the unique ID in a multi-tenant environment
func (a *IPAllocations) TenantID() string {
	return ConcatenatedKey(a.Tenant, a.ID)
}

// TotalUsage returns the sum of all usage units
// Exported to be used in FilterS
func (a *IPAllocations) TotalUsage() float64 {
	var tu float64
	for _, ru := range a.Usages {
		tu += ru.Units
	}
	return tu
}

// IPAllocationsLockKey returns the ID used to lock IP allocations with guardian
func IPAllocationsLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPAllocations, tnt, id)
}
