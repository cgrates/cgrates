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

// IPProfile defines the configuration of the IP.
type IPProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Weights   DynamicWeights
	TTL       time.Duration
	Stored    bool
	Pools     []*Pool
}

// TenantID returns the concatenated tenant and ID.
func (ip *IPProfile) TenantID() string {
	return ConcatenatedKey(ip.Tenant, ip.ID)
}

// Clone creates a deep copy of IPProfile for thread-safe use.
func (ip *IPProfile) Clone() *IPProfile {
	if ip == nil {
		return nil
	}
	pools := make([]*Pool, 0, len(ip.Pools))
	for _, pool := range ip.Pools {
		pools = append(pools, pool.Clone())
	}
	return &IPProfile{
		Tenant:    ip.Tenant,
		ID:        ip.ID,
		FilterIDs: slices.Clone(ip.FilterIDs),
		Weights:   ip.Weights.Clone(),
		TTL:       ip.TTL,
		Stored:    ip.Stored,
		Pools:     pools,
	}
}

// CacheClone returns a clone of IPProfile used by ltcache CacheCloner
func (ip *IPProfile) CacheClone() any {
	return ip.Clone()
}

func (ip *IPProfile) Set(path []string, val any, newBranch bool) error {
	if len(path) != 1 && len(path) != 2 {
		return ErrWrongPath
	}
	var err error
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		ip.Tenant = IfaceAsString(val)
	case ID:
		ip.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		ip.FilterIDs = append(ip.FilterIDs, valA...)
	case Weights:
		if val != "" {
			ip.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case TTL:
		ip.TTL, err = IfaceAsDuration(val)
	case Stored:
		ip.Stored, err = IfaceAsBool(val)
	case Pools:
		if len(path) != 2 {
			return ErrWrongPath
		}
		if val == EmptyString {
			return nil
		}
		if len(ip.Pools) == 0 || newBranch {
			ip.Pools = append(ip.Pools, new(Pool))
		}
		pool := ip.Pools[len(ip.Pools)-1]
		return pool.Set(path[1:], val, newBranch)
	}
	return err
}

func (ip *IPProfile) Merge(other any) {
	o := other.(*IPProfile)
	if len(o.Tenant) != 0 {
		ip.Tenant = o.Tenant
	}
	if len(o.ID) != 0 {
		ip.ID = o.ID
	}
	ip.FilterIDs = append(ip.FilterIDs, o.FilterIDs...)
	ip.Weights = append(ip.Weights, o.Weights...)
	if o.TTL != 0 {
		ip.TTL = o.TTL
	}
	if o.Stored {
		ip.Stored = o.Stored
	}
	for _, pool := range o.Pools {
		if idx := slices.IndexFunc(ip.Pools, func(p *Pool) bool {
			return p.ID == pool.ID
		}); idx != -1 {
			ip.Pools[idx].Merge(pool)
			continue
		}
		ip.Pools = append(ip.Pools, pool)
	}
}

func (ip *IPProfile) String() string { return ToJSON(ip) }

func (ip *IPProfile) FieldAsString(fldPath []string) (string, error) {
	val, err := ip.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return IfaceAsString(val), nil
}

func (ip *IPProfile) FieldAsInterface(fldPath []string) (any, error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case FilterIDs:
				if *idx < len(ip.FilterIDs) {
					return ip.FilterIDs[*idx], nil
				}
			case Pools:
				if *idx < len(ip.Pools) {
					return ip.Pools[*idx].FieldAsInterface(fldPath[1:])
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return ip.Tenant, nil
	case ID:
		return ip.ID, nil
	case FilterIDs:
		return ip.FilterIDs, nil
	case Weights:
		return ip.Weights, nil
	case TTL:
		return ip.TTL, nil
	case Stored:
		return ip.Stored, nil
	case Pools:
		return ip.Pools, nil
	}
}

// IPProfileLockKey returns the ID used to lock a resourceProfile with guardian
func IPProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPProfiles, tnt, id)
}

// IPProfileWithAPIOpts wraps IPProfile with APIOpts.
type IPProfileWithAPIOpts struct {
	*IPProfile
	APIOpts map[string]any
}

type Pool struct {
	ID        string
	FilterIDs []string
	Type      string
	Range     string
	Strategy  string
	Message   string
	Weights   DynamicWeights
	Blockers  DynamicBlockers
}

// Clone creates a deep copy of IPProfile for thread-safe use.
func (p *Pool) Clone() *Pool {
	if p == nil {
		return nil
	}
	return &Pool{
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

func (p *Pool) Set(path []string, val any, _ bool) error {
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

func (p *Pool) Merge(v2 any) {
	vi := v2.(*Pool)

	// NOTE: Merge gets called when the IDs are equal, so this is a no-op.
	// Kept for consistency with other components.
	if len(vi.ID) != 0 {
		p.ID = vi.ID
	}

	p.FilterIDs = append(p.FilterIDs, vi.FilterIDs...)
	if vi.Type != "" {
		p.Type = vi.Type
	}
	if vi.Range != "" {
		p.Range = vi.Range
	}
	if vi.Strategy != "" {
		p.Strategy = vi.Strategy
	}
	if vi.Message != "" {
		p.Message = vi.Message
	}
	p.Weights = append(p.Weights, vi.Weights...)
	p.Blockers = append(p.Blockers, vi.Blockers...)
}

func (p *Pool) String() string { return ToJSON(p) }

func (p *Pool) FieldAsString(fldPath []string) (string, error) {
	val, err := p.FieldAsInterface(fldPath)
	if err != nil {
		return "", err
	}
	return IfaceAsString(val), nil
}

func (p *Pool) FieldAsInterface(fldPath []string) (any, error) {
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

// IPUsage represents an usage counted.
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

// isActive checks ExpiryTime at some time
func (u *IPUsage) IsActive(atTime time.Time) bool {
	return u.ExpiryTime.IsZero() || u.ExpiryTime.Sub(atTime) > 0
}

// Clone duplicates ru
func (u *IPUsage) Clone() *IPUsage {
	if u == nil {
		return nil
	}
	clone := *u
	return &clone
}

// IP represents ...
type IP struct {
	Tenant string
	ID     string
	Usages map[string]*IPUsage
	TTLIdx []string
}

// Clone clones *IP (lkID excluded)
func (ip *IP) Clone() *IP {
	if ip == nil {
		return nil
	}
	clone := &IP{
		Tenant: ip.Tenant,
		ID:     ip.ID,
		TTLIdx: slices.Clone(ip.TTLIdx),
	}
	if ip.Usages != nil {
		clone.Usages = make(map[string]*IPUsage, len(ip.Usages))
		for key, usage := range ip.Usages {
			clone.Usages[key] = usage.Clone()
		}
	}
	return clone
}

// CacheClone returns a clone of IP used by ltcache CacheCloner
func (ip *IP) CacheClone() any {
	return ip.Clone()
}

// IPWithAPIOpts wraps IP with APIOpts.
type IPWithAPIOpts struct {
	*IP
	APIOpts map[string]any
}

// TenantID returns the unique ID in a multi-tenant environment
func (ip *IP) TenantID() string {
	return ConcatenatedKey(ip.Tenant, ip.ID)
}

// TotalUsage returns the sum of all usage units
// Exported to be used in FilterS
func (ip *IP) TotalUsage() float64 {
	var tu float64
	for _, ru := range ip.Usages {
		tu += ru.Units
	}
	return tu
}

// IPLockKey returns the ID used to lock a resource with guardian
func IPLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPs, tnt, id)
}
