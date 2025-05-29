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
	Tenant      string
	ID          string
	FilterIDs   []string
	TTL         time.Duration
	Type        string
	AddressPool string
	Allocation  string
	Stored      bool
	Weights     DynamicWeights
}

// Clone creates a deep copy of IPProfile for thread-safe use.
func (ip *IPProfile) Clone() *IPProfile {
	if ip == nil {
		return nil
	}
	return &IPProfile{
		Tenant:      ip.Tenant,
		ID:          ip.ID,
		FilterIDs:   slices.Clone(ip.FilterIDs),
		TTL:         ip.TTL,
		Type:        ip.Type,
		AddressPool: ip.AddressPool,
		Allocation:  ip.Allocation,
		Stored:      ip.Stored,
		Weights:     ip.Weights.Clone(),
	}
}

// CacheClone returns a clone of IPProfile used by ltcache CacheCloner
func (ip *IPProfile) CacheClone() any {
	return ip.Clone()
}

// IPProfileWithAPIOpts wraps IPProfile with APIOpts.
type IPProfileWithAPIOpts struct {
	*IPProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (ip *IPProfile) TenantID() string {
	return ConcatenatedKey(ip.Tenant, ip.ID)
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

func (ip *IPProfile) Set(path []string, val any, _ bool) error {
	if len(path) != 1 {
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
	case TTL:
		ip.TTL, err = IfaceAsDuration(val)
	case Type:
		ip.Type = IfaceAsString(val)
	case AddressPool:
		ip.AddressPool = IfaceAsString(val)
	case Allocation:
		ip.Allocation = IfaceAsString(val)
	case Stored:
		ip.Stored, err = IfaceAsBool(val)
	case Weights:
		if val != "" {
			ip.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	}
	return err
}

func (ip *IPProfile) Merge(v2 any) {
	vi := v2.(*IPProfile)
	if len(vi.Tenant) != 0 {
		ip.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		ip.ID = vi.ID
	}
	ip.FilterIDs = append(ip.FilterIDs, vi.FilterIDs...)
	if len(vi.Allocation) != 0 {
		ip.Allocation = vi.Allocation
	}
	if vi.TTL != 0 {
		ip.TTL = vi.TTL
	}
	if vi.Type != "" {
		ip.Type = vi.Type
	}
	if vi.AddressPool != "" {
		ip.AddressPool = vi.AddressPool
	}
	if vi.Stored {
		ip.Stored = vi.Stored
	}
	ip.Weights = append(ip.Weights, vi.Weights...)
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
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return ip.Tenant, nil
	case ID:
		return ip.ID, nil
	case FilterIDs:
		return ip.FilterIDs, nil
	case TTL:
		return ip.TTL, nil
	case Type:
		return ip.Type, nil
	case AddressPool:
		return ip.AddressPool, nil
	case Allocation:
		return ip.Allocation, nil
	case Stored:
		return ip.Stored, nil
	case Weights:
		return ip.Weights, nil
	}
}

// IPProfileLockKey returns the ID used to lock a resourceProfile with guardian
func IPProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPProfiles, tnt, id)
}

// IPLockKey returns the ID used to lock a resource with guardian
func IPLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheIPs, tnt, id)
}
