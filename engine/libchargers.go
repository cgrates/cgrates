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

package engine

import (
	"sort"

	"github.com/cgrates/cgrates/utils"
)

// ChargerProfile is the config for one Charger
type ChargerProfile struct {
	Tenant       string
	ID           string
	FilterIDs    []string
	Weights      utils.DynamicWeights
	Blockers     utils.DynamicBlockers
	RunID        string
	AttributeIDs []string // perform data aliasing based on these Attributes
	weight       float64
}

// ChargerProfileWithAPIOpts is used in replicatorV1 for dispatcher
type ChargerProfileWithAPIOpts struct {
	*ChargerProfile
	APIOpts map[string]any
}

func (cP *ChargerProfile) TenantID() string {
	return utils.ConcatenatedKey(cP.Tenant, cP.ID)
}

// ChargerProfiles is a sortable list of Charger profiles
type ChargerProfiles []*ChargerProfile

// Sort is part of sort interface, sort based on Weight
func (cps ChargerProfiles) Sort() {
	sort.Slice(cps, func(i, j int) bool { return cps[i].weight > cps[j].weight })
}

func (cp *ChargerProfile) Set(path []string, val any, newBranch bool) (err error) {
	if len(path) != 1 {
		return utils.ErrWrongPath
	}
	switch path[0] {
	default:
		return utils.ErrWrongPath
	case utils.Tenant:
		cp.Tenant = utils.IfaceAsString(val)
	case utils.ID:
		cp.ID = utils.IfaceAsString(val)
	case utils.FilterIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		cp.FilterIDs = append(cp.FilterIDs, valA...)
	case utils.RunID:
		cp.RunID = utils.IfaceAsString(val)
	case utils.AttributeIDs:
		var valA []string
		valA, err = utils.IfaceAsStringSlice(val)
		cp.AttributeIDs = append(cp.AttributeIDs, valA...)
	case utils.Weights:
		if val != utils.EmptyString {
			cp.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
		}
	case utils.Blockers:
		if val != utils.EmptyString {
			cp.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
		}
	}
	return
}

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

func (cp *ChargerProfile) String() string { return utils.ToJSON(cp) }
func (cp *ChargerProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = cp.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (cp *ChargerProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		fld, idx := utils.GetPathIndex(fldPath[0])
		if idx != nil {
			switch fld {
			case utils.AttributeIDs:
				if *idx < len(cp.AttributeIDs) {
					return cp.AttributeIDs[*idx], nil
				}
			case utils.FilterIDs:
				if *idx < len(cp.FilterIDs) {
					return cp.FilterIDs[*idx], nil
				}
			}
		}
		return nil, utils.ErrNotFound
	case utils.Tenant:
		return cp.Tenant, nil
	case utils.ID:
		return cp.ID, nil
	case utils.FilterIDs:
		return cp.FilterIDs, nil
	case utils.Weights:
		return cp.Weights, nil
	case utils.Blockers:
		return cp.Blockers, nil
	case utils.AttributeIDs:
		return cp.AttributeIDs, nil
	case utils.RunID:
		return cp.RunID, nil
	}
}
