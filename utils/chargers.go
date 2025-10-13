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

// ChargerProfile defines the configuration of a Charger.
type ChargerProfile struct {
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      DynamicWeights
	Blockers     DynamicBlockers
	RunID        string
	AttributeIDs []string // perform data aliasing based on these Attributes
}

// Clone method for ChargerProfile
func (cp *ChargerProfile) Clone() *ChargerProfile {
	if cp == nil {
		return nil
	}
	clone := &ChargerProfile{
		Tenant: cp.Tenant,
		ID:     cp.ID,
		RunID:  cp.RunID,
	}
	if cp.FilterIDs != nil {
		clone.FilterIDs = make([]string, len(cp.FilterIDs))
		copy(clone.FilterIDs, cp.FilterIDs)
	}
	if cp.AttributeIDs != nil {
		clone.AttributeIDs = make([]string, len(cp.AttributeIDs))
		copy(clone.AttributeIDs, cp.AttributeIDs)
	}
	if cp.Weights != nil {
		clone.Weights = cp.Weights.Clone()
	}
	if cp.Blockers != nil {
		clone.Blockers = cp.Blockers.Clone()
	}
	return clone
}

// CacheClone returns a clone of ChargerProfile used by ltcache CacheCloner
func (cp *ChargerProfile) CacheClone() any {
	return cp.Clone()
}

// TenantID returns the concatenated tenant and ID.
func (cp *ChargerProfile) TenantID() string {
	return ConcatenatedKey(cp.Tenant, cp.ID)
}

// Set implements the profile interface, setting values in ChargerProfile based on path.
func (cp *ChargerProfile) Set(path []string, val any, newBranch bool) (err error) {
	if len(path) != 1 {
		return ErrWrongPath
	}
	switch path[0] {
	default:
		return ErrWrongPath
	case Tenant:
		cp.Tenant = IfaceAsString(val)
	case ID:
		cp.ID = IfaceAsString(val)
	case FilterIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		cp.FilterIDs = append(cp.FilterIDs, valA...)
	case RunID:
		cp.RunID = IfaceAsString(val)
	case AttributeIDs:
		var valA []string
		valA, err = IfaceAsStringSlice(val)
		cp.AttributeIDs = append(cp.AttributeIDs, valA...)
	case Weights:
		if val != EmptyString {
			cp.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	case Blockers:
		if val != EmptyString {
			cp.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
		}
	}
	return
}

// Merge implements the profile interface, merging values from another ChargerProfile.
func (cp *ChargerProfile) Merge(v2 any) {
	vi := v2.(*ChargerProfile)
	if len(vi.Tenant) != 0 {
		cp.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		cp.ID = vi.ID
	}
	if len(vi.RunID) != 0 {
		cp.RunID = vi.RunID
	}
	cp.FilterIDs = append(cp.FilterIDs, vi.FilterIDs...)
	cp.AttributeIDs = append(cp.AttributeIDs, vi.AttributeIDs...)
	cp.Weights = append(cp.Weights, vi.Weights...)
	cp.Blockers = append(cp.Blockers, vi.Blockers...)
}

// String implements the DataProvider interface, returning the ChargerProfile in JSON format.
func (cp *ChargerProfile) String() string { return ToJSON(cp) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (cp *ChargerProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = cp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (cp *ChargerProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case AttributeIDs:
				if *idx < len(cp.AttributeIDs) {
					return cp.AttributeIDs[*idx], nil
				}
			case FilterIDs:
				if *idx < len(cp.FilterIDs) {
					return cp.FilterIDs[*idx], nil
				}
			}
		}
		return nil, ErrNotFound
	case Tenant:
		return cp.Tenant, nil
	case ID:
		return cp.ID, nil
	case FilterIDs:
		return cp.FilterIDs, nil
	case Weights:
		return cp.Weights, nil
	case Blockers:
		return cp.Blockers, nil
	case AttributeIDs:
		return cp.AttributeIDs, nil
	case RunID:
		return cp.RunID, nil
	}
}

// ChargerProfileWithAPIOpts wraps ChargerProfile with APIOpts.
type ChargerProfileWithAPIOpts struct {
	*ChargerProfile
	APIOpts map[string]any
}
