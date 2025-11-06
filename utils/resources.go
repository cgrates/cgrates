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
	"time"
)

// ResourceProfile represents the user configuration for the resource
type ResourceProfile struct {
	Tenant            string
	ID                string // identifier of this resource
	FilterIDs         []string
	UsageTTL          time.Duration // auto-expire the usage after this duration
	Limit             float64       // limixt value
	AllocationMessage string        // message returned by the winning resource on allocation
	Blocker           bool          // blocker flag to stop processing on filters matched
	Stored            bool
	Weights           DynamicWeights // Weight to sort the resources
	ThresholdIDs      []string       // Thresholds to check after changing Limit
}

// Clone clones *ResourceProfile (lkID excluded)
func (rp *ResourceProfile) Clone() *ResourceProfile {
	if rp == nil {
		return nil
	}
	clone := &ResourceProfile{

		Tenant:            rp.Tenant,
		ID:                rp.ID,
		UsageTTL:          rp.UsageTTL,
		Limit:             rp.Limit,
		AllocationMessage: rp.AllocationMessage,
		Blocker:           rp.Blocker,
		Stored:            rp.Stored,
	}
	if rp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(rp.FilterIDs))
		copy(clone.FilterIDs, rp.FilterIDs)
	}
	if rp.ThresholdIDs != nil {
		clone.ThresholdIDs = make([]string, len(rp.ThresholdIDs))
		copy(clone.ThresholdIDs, rp.ThresholdIDs)
	}
	if rp.Weights != nil {
		clone.Weights = rp.Weights.Clone()
	}
	return clone
}

// CacheClone returns a clone of ResourceProfile used by ltcache CacheCloner
func (rp *ResourceProfile) CacheClone() any {
	return rp.Clone()
}

// ResourceProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ResourceProfileWithAPIOpts struct {
	*ResourceProfile
	APIOpts map[string]any
}

// TenantID returns unique identifier of the ResourceProfile in a multi-tenant environment
func (rp *ResourceProfile) TenantID() string {
	return ConcatenatedKey(rp.Tenant, rp.ID)
}

// ResourceUsage represents an usage counted
type ResourceUsage struct {
	Tenant     string
	ID         string // Unique identifier of this ResourceUsage, Eg: FreeSWITCH UUID
	ExpiryTime time.Time
	Units      float64 // Number of units used
}

// TenantID returns the concatenated key between tenant and ID
func (ru *ResourceUsage) TenantID() string {
	return ConcatenatedKey(ru.Tenant, ru.ID)
}

// isActive checks ExpiryTime at some time
func (ru *ResourceUsage) IsActive(atTime time.Time) bool {
	return ru.ExpiryTime.IsZero() || ru.ExpiryTime.Sub(atTime) > 0
}

// Clone duplicates ru
func (ru *ResourceUsage) Clone() (cln *ResourceUsage) {
	cln = new(ResourceUsage)
	*cln = *ru
	return
}

// Resource represents a resource in the system
// not thread safe, needs locking at process level
type Resource struct {
	Tenant string
	ID     string
	Usages map[string]*ResourceUsage
	TTLIdx []string // holds ordered list of ResourceIDs based on their TTL, empty if feature is disabled
}

// Clone clones *Resource (lkID excluded)
func (r *Resource) Clone() *Resource {
	if r == nil {
		return nil
	}
	clone := &Resource{
		Tenant: r.Tenant,
		ID:     r.ID,
	}
	if r.Usages != nil {
		clone.Usages = make(map[string]*ResourceUsage, len(r.Usages))
		for key, usage := range r.Usages {
			clone.Usages[key] = usage.Clone()
		}
	}
	if r.TTLIdx != nil {
		clone.TTLIdx = make([]string, len(r.TTLIdx))
		copy(clone.TTLIdx, r.TTLIdx)
	}
	return clone
}

// CacheClone returns a clone of Resource used by ltcache CacheCloner
func (r *Resource) CacheClone() any {
	return r.Clone()
}

// ResourceWithAPIOpts is used in replicatorV1 for dispatcher
type ResourceWithAPIOpts struct {
	*Resource
	APIOpts map[string]any
}

// TenantID returns the unique ID in a multi-tenant environment
func (r *Resource) TenantID() string {
	return ConcatenatedKey(r.Tenant, r.ID)
}

// TotalUsage returns the sum of all usage units
// Exported to be used in FilterS
func (r *Resource) TotalUsage() float64 {
	var tu float64
	for _, ru := range r.Usages {
		tu += ru.Units
	}
	return tu
}

// AsMapStringInterface converts Resource struct to map[string]any
func (rp *Resource) AsMapStringInterface() map[string]any {
	if rp == nil {
		return nil
	}
	return map[string]any{
		Tenant: rp.Tenant,
		ID:     rp.ID,
		Usages: rp.Usages,
		TTLIdx: rp.TTLIdx,
	}
}

// MapStringInterfaceToResource converts map[string]any to Resource struct
func MapStringInterfaceToResource(m map[string]any) *Resource {
	rp := &Resource{}
	if v, ok := m[Tenant].(string); ok {
		rp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		rp.ID = v
	}
	rp.Usages = InterfaceToMapStringResourceUsage(m[Usages])
	rp.TTLIdx = InterfaceToStringSlice(m[TTLIdx])
	return rp
}

// InterfaceToMapStringResourceUsage converts any to map[string]*ResourceUsage
func InterfaceToMapStringResourceUsage(v any) map[string]*ResourceUsage {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case map[string]*ResourceUsage:
		return val
	case map[string]any:
		result := make(map[string]*ResourceUsage)
		for k, v := range val {
			if balMap, ok := v.(map[string]any); ok {
				result[k] = MapStringInterfaceToResourceUsage(balMap)
			} else if bal, ok := v.(*ResourceUsage); ok {
				result[k] = bal
			}
		}
		return result
	}
	return nil
}

// MapStringInterfaceToResourceUsage converts map[string]any to *ResourceUsage
func MapStringInterfaceToResourceUsage(m map[string]any) *ResourceUsage {
	resUsage := &ResourceUsage{}
	if v, ok := m[Tenant].(string); ok {
		resUsage.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		resUsage.ID = v
	}
	if v, ok := m[ExpiryTime].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			resUsage.ExpiryTime = t
		}
	}
	if v, ok := m[Units].(float64); ok {
		resUsage.Units = v
	}
	return resUsage
}

// Available returns the available number of units
// Exported method to be used by filterS
func (r *ResourceWithConfig) Available() float64 {
	return r.Config.Limit - r.TotalUsage()
}

type ResourceWithConfig struct {
	*Resource
	Config *ResourceProfile
}

func (rp *ResourceProfile) Set(path []string, val any, _ bool) (err error) {
	if len(path) != 1 {
		return ErrWrongPath
	}
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		rp.Tenant = IfaceAsString(val)
	case ID:
		rp.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.FilterIDs = append(rp.FilterIDs, valA...)
	case UsageTTL:
		rp.UsageTTL, err = IfaceAsDuration(val)
	case Limit:
		if val != EmptyString {
			rp.Limit, err = IfaceAsFloat64(val)
		}
	case AllocationMessage:
		rp.AllocationMessage = IfaceAsString(val)
	case Blocker:
		rp.Blocker, err = IfaceAsBool(val)
	case Stored:
		rp.Stored, err = IfaceAsBool(val)
	case Weights:
		if val != EmptyString {
			rp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case ThresholdIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		rp.ThresholdIDs = append(rp.ThresholdIDs, valA...)
	}
	return
}

func (rp *ResourceProfile) Merge(v2 any) {
	vi := v2.(*ResourceProfile)
	if len(vi.Tenant) != 0 {
		rp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		rp.ID = vi.ID
	}
	rp.FilterIDs = append(rp.FilterIDs, vi.FilterIDs...)
	rp.ThresholdIDs = append(rp.ThresholdIDs, vi.ThresholdIDs...)
	if len(vi.AllocationMessage) != 0 {
		rp.AllocationMessage = vi.AllocationMessage
	}
	if vi.UsageTTL != 0 {
		rp.UsageTTL = vi.UsageTTL
	}
	if vi.Limit != 0 {
		rp.Limit = vi.Limit
	}
	if vi.Blocker {
		rp.Blocker = vi.Blocker
	}
	if vi.Stored {
		rp.Stored = vi.Stored
	}
	rp.Weights = append(rp.Weights, vi.Weights...)
}

func (rp *ResourceProfile) String() string { return ToJSON(rp) }
func (rp *ResourceProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = rp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}
func (rp *ResourceProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case ThresholdIDs:
				if *idx < len(rp.ThresholdIDs) {
					return rp.ThresholdIDs[*idx], nil
				}
			case FilterIDs:
				if *idx < len(rp.FilterIDs) {
					return rp.FilterIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return rp.Tenant, nil
	case ID:
		return rp.ID, nil
	case FilterIDs:
		return rp.FilterIDs, nil
	case UsageTTL:
		return rp.UsageTTL, nil
	case Limit:
		return rp.Limit, nil
	case AllocationMessage:
		return rp.AllocationMessage, nil
	case Blocker:
		return rp.Blocker, nil
	case Stored:
		return rp.Stored, nil
	case Weights:
		return rp.Weights, nil
	case ThresholdIDs:
		return rp.ThresholdIDs, nil
	}
}

// ResourceProfileLockKey returns the ID used to lock a resourceProfile with guardian
func ResourceProfileLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheResourceProfiles, tnt, id)
}

// ResourceLockKey returns the ID used to lock a resource with guardian
func ResourceLockKey(tnt, id string) string {
	return ConcatenatedKey(CacheResources, tnt, id)
}

// AsMapStringInterface converts ResourceProfile struct to map[string]any
func (rp *ResourceProfile) AsMapStringInterface() map[string]any {
	if rp == nil {
		return nil
	}
	return map[string]any{
		Tenant:            rp.Tenant,
		ID:                rp.ID,
		FilterIDs:         rp.FilterIDs,
		UsageTTL:          rp.FilterIDs,
		Limit:             rp.Limit,
		AllocationMessage: rp.AllocationMessage,
		Blocker:           rp.Blocker,
		Stored:            rp.Stored,
		Weights:           rp.Weights,
		ThresholdIDs:      rp.ThresholdIDs,
	}
}

// MapStringInterfaceToResourceProfile converts map[string]any to ResourceProfile struct
func MapStringInterfaceToResourceProfile(m map[string]any) (rp *ResourceProfile, err error) {
	rp = &ResourceProfile{}
	if v, ok := m[Tenant].(string); ok {
		rp.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		rp.ID = v
	}
	rp.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	if v, ok := m[UsageTTL].(string); ok {
		if dur, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			rp.UsageTTL = dur
		}
	} else if v, ok := m[UsageTTL].(float64); ok { // for -1 cases
		rp.UsageTTL = time.Duration(v)
	}
	if v, ok := m[Limit].(float64); ok {
		rp.Limit = v
	}
	if v, ok := m[AllocationMessage].(string); ok {
		rp.AllocationMessage = v
	}
	if v, ok := m[Blocker].(bool); ok {
		rp.Blocker = v
	}
	if v, ok := m[Stored].(bool); ok {
		rp.Stored = v
	}
	rp.Weights = InterfaceToDynamicWeights(m[Weights])
	rp.ThresholdIDs = InterfaceToStringSlice(m[ThresholdIDs])
	return
}
