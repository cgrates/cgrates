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
	"strconv"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// ActionProfile represents the configuration of a Action profile
type ActionProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Weights   utils.DynamicWeights
	Blockers  utils.DynamicBlockers
	Schedule  string
	Targets   map[string]utils.StringSet

	Actions []*APAction
	weight  float64
}

func (aP *ActionProfile) TenantID() string {
	return utils.ConcatenatedKey(aP.Tenant, aP.ID)
}

// ActionProfiles is a sortable list of ActionProfiles
type ActionProfiles []*ActionProfile

// Sort is part of sort interface, sort based on Weight
func (aps ActionProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].weight > aps[j].weight })
}

func (ap *ActionProfile) GetWeightFromDynamics(ctx *context.Context,
	fltrS *FilterS, tnt string, ev utils.DataProvider) (err error) {
	if ap.weight, err = WeightFromDynamics(ctx, ap.Weights, fltrS, tnt, ev); err != nil {
		return
	}
	return
}

// APAction defines action related information used within a ActionProfile
type APAction struct {
	ID        string         // Action ID
	FilterIDs []string       // Action FilterIDs
	TTL       time.Duration  // Cancel Action if not executed within TTL
	Type      string         // Type of Action
	Opts      map[string]any // Extra options to pass depending on action type
	Diktats   []*APDiktat
}

type APDiktat struct {
	Path  string // Path to execute
	Value string // Value to execute on Path

	valRSR config.RSRParsers
}

// RSRValues returns the Value as RSRParsers
func (dk *APDiktat) RSRValues() (config.RSRParsers, error) {
	if dk.valRSR == nil {
		return config.NewRSRParsers(dk.Value, utils.RSRSep)
	}
	return dk.valRSR, nil
}

// ActionProfileWithAPIOpts is used in API calls
type ActionProfileWithAPIOpts struct {
	*ActionProfile
	APIOpts map[string]any
}

func (aP *ActionProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	case 0:
		return utils.ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], utils.Targets) &&
				path[0][7] == '[' && path[0][len(path[0])-1] == ']' {
				var valA []string
				valA, err = utils.IfaceAsStringSlice(val)
				aP.Targets[path[0][8:len(path[0])-1]] = utils.JoinStringSet(aP.Targets[path[0][8:len(path[0])-1]], utils.NewStringSet(valA))
				return
			}
			return utils.ErrWrongPath
		case utils.Tenant:
			aP.Tenant = utils.IfaceAsString(val)
		case utils.ID:
			aP.ID = utils.IfaceAsString(val)
		case utils.Schedule:
			aP.Schedule = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			aP.FilterIDs = append(aP.FilterIDs, valA...)
		case utils.Weights:
			if val != utils.EmptyString {
				aP.Weights, err = utils.NewDynamicWeightsFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		case utils.Blockers:
			if val != utils.EmptyString {
				aP.Blockers, err = utils.NewDynamicBlockersFromString(utils.IfaceAsString(val), utils.InfieldSep, utils.ANDSep)
			}
		}
		return
	case 2:
		if path[0] == utils.Targets {
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			aP.Targets[path[1]] = utils.JoinStringSet(aP.Targets[path[1]], utils.NewStringSet(valA))
			return
		}
	default:
	}

	var acID string
	if path[0] == utils.Actions {
		acID = path[1]
		path = path[1:]
	} else if strings.HasPrefix(path[0], utils.Actions) &&
		path[0][7] == '[' && path[0][len(path[0])-1] == ']' {
		acID = path[0][8 : len(path[0])-1]
	}
	if acID == utils.EmptyString {
		return utils.ErrWrongPath
	}

	var ac *APAction
	for _, a := range aP.Actions {
		if a.ID == acID {
			ac = a
			break
		}
	}
	if ac == nil {
		ac = &APAction{ID: acID, Opts: make(map[string]any)}
		aP.Actions = append(aP.Actions, ac)
	}
	return ac.Set(path[1:], val, newBranch)
}

func (aP *APAction) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		if path[0] == utils.Opts {
			return utils.MapStorage(aP.Opts).Set(path[1:], val)
		}
		return utils.ErrWrongPath
	case 0:
		return utils.ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], utils.Opts) &&
				path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
				aP.Opts[path[0][5:len(path[0])-1]] = val
				return
			}
			return utils.ErrWrongPath
		case utils.ID:
			aP.ID = utils.IfaceAsString(val)
		case utils.Type:
			aP.Type = utils.IfaceAsString(val)
		case utils.FilterIDs:
			var valA []string
			valA, err = utils.IfaceAsStringSlice(val)
			aP.FilterIDs = append(aP.FilterIDs, valA...)
		case utils.TTL:
			aP.TTL, err = utils.IfaceAsDuration(val)
		case utils.Opts:
			aP.Opts, err = utils.NewMapFromCSV(utils.IfaceAsString(val))
		}
	case 2:
		switch path[0] {
		default:
			return utils.ErrWrongPath
		case utils.Opts:
			return utils.MapStorage(aP.Opts).Set(path[1:], val)
		case utils.Diktats:
			if len(aP.Diktats) == 0 || newBranch {
				aP.Diktats = append(aP.Diktats, new(APDiktat))
			}
			switch path[1] {
			case utils.Path:
				aP.Diktats[len(aP.Diktats)-1].Path = utils.IfaceAsString(val)
			case utils.Value:
				aP.Diktats[len(aP.Diktats)-1].Value = utils.IfaceAsString(val)
			}
		}
	}
	return
}

func (ap *ActionProfile) Merge(v2 any) {
	vi := v2.(*ActionProfile)
	if len(vi.Tenant) != 0 {
		ap.Tenant = vi.Tenant
	}
	if len(vi.ID) != 0 {
		ap.ID = vi.ID
	}
	ap.FilterIDs = append(ap.FilterIDs, vi.FilterIDs...)
	var equal bool
	for _, actionV2 := range vi.Actions {
		for _, action := range ap.Actions {
			if action.ID == actionV2.ID {
				action.Merge(actionV2)
				equal = true
				break
			}
		}
		if !equal {
			ap.Actions = append(ap.Actions, actionV2)
		}
		equal = false
	}

	ap.Weights = append(ap.Weights, vi.Weights...)
	ap.Blockers = append(ap.Blockers, vi.Blockers...)
	if len(vi.Schedule) != 0 {
		ap.Schedule = vi.Schedule
	}
	for k, v := range vi.Targets {
		if k == utils.EmptyString {
			continue
		}
		ap.Targets[k] = v
	}
}

func (apAct *APAction) Merge(v2 *APAction) {
	if len(v2.ID) != 0 {
		apAct.ID = v2.ID
	}
	if v2.TTL != 0 {
		apAct.TTL = v2.TTL
	}
	if len(v2.Type) != 0 {
		apAct.Type = v2.Type
	}
	for key, value := range v2.Opts {
		apAct.Opts[key] = value
	}
	apAct.FilterIDs = append(apAct.FilterIDs, v2.FilterIDs...)
	if len(apAct.Diktats) == 1 && apAct.Diktats[0].Path == utils.EmptyString {
		apAct.Diktats = apAct.Diktats[:0]
	}
	for _, diktat := range v2.Diktats {
		if diktat.Path != utils.EmptyString {
			apAct.Diktats = append(apAct.Diktats, diktat)
		}
	}
}

func (ap *ActionProfile) String() string { return utils.ToJSON(ap) }
func (ap *ActionProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = ap.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (ap *ActionProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := utils.GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case utils.Actions:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.Actions) {
						return ap.Actions[idx], nil
					}
				case utils.FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.FilterIDs) {
						return ap.FilterIDs[idx], nil
					}
				case utils.Targets:
					if tr, has := ap.Targets[*idxStr]; has {
						return tr, nil
					}
				}
			}
			return nil, utils.ErrNotFound
		case utils.Tenant:
			return ap.Tenant, nil
		case utils.ID:
			return ap.ID, nil
		case utils.FilterIDs:
			return ap.FilterIDs, nil
		case utils.Weights:
			return ap.Weights, nil
		case utils.Blockers:
			return ap.Blockers, nil
		case utils.Actions:
			return ap.Actions, nil
		case utils.Schedule:
			return ap.Schedule, nil
		case utils.Targets:
			return ap.Targets, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, utils.ErrNotFound
	}
	fld, idxStr := utils.GetPathIndexString(fldPath[0])
	switch fld {
	default:
		return nil, utils.ErrNotFound
	case utils.Actions:
		if idxStr == nil {
			return nil, utils.ErrNotFound
		}
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(ap.Actions) {
			return nil, utils.ErrNotFound
		}
		return ap.Actions[idx].FieldAsInterface(fldPath[1:])
	case utils.Targets:
		tr, has := ap.Targets[*idxStr]
		if !has {
			return nil, utils.ErrNotFound
		}
		return tr.FieldAsInterface(fldPath[1:])
	}
}

func (a *APAction) String() string { return utils.ToJSON(a) }
func (a *APAction) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = a.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (ap *APAction) FieldAsInterface(fldPath []string) (_ any, err error) {
	switch len(fldPath) {
	default:
		if fld, idxStr := utils.GetPathIndexString(fldPath[0]); fld == utils.Opts {
			path := fldPath[1:]
			if idxStr != nil {
				path = append([]string{*idxStr}, path...)
			}
			return utils.MapStorage(ap.Opts).FieldAsInterface(path)
		}
		fallthrough
	case 0:
		return nil, utils.ErrNotFound
	case 1:
		switch fldPath[0] {
		default:
			fld, idxStr := utils.GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case utils.FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.FilterIDs) {
						return ap.FilterIDs[idx], nil
					}
				case utils.Diktats:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.Diktats) {
						return ap.Diktats[idx], nil
					}
				case utils.Opts:
					return utils.MapStorage(ap.Opts).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, utils.ErrNotFound
		case utils.ID:
			return ap.ID, nil
		case utils.FilterIDs:
			return ap.FilterIDs, nil
		case utils.TTL:
			return ap.TTL, nil
		case utils.Diktats:
			return ap.Diktats, nil
		case utils.Type:
			return ap.Type, nil
		case utils.Opts:
			return ap.Opts, nil
		}
	case 2:
		fld, idxStr := utils.GetPathIndexString(fldPath[0])
		switch fld {
		default:
			return nil, utils.ErrNotFound
		case utils.Opts:
			path := fldPath[1:]
			if idxStr != nil {
				path = append([]string{*idxStr}, path...)
			}
			return utils.MapStorage(ap.Opts).FieldAsInterface(path)
		case utils.Diktats:
			if idxStr == nil {
				return nil, utils.ErrNotFound
			}
			var idx int
			if idx, err = strconv.Atoi(*idxStr); err != nil {
				return
			}
			if idx >= len(ap.Diktats) {
				return nil, utils.ErrNotFound
			}
			return ap.Diktats[idx].FieldAsInterface(fldPath[1:])
		}
	}
}

func (dk *APDiktat) String() string { return utils.ToJSON(dk) }
func (dk *APDiktat) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = dk.FieldAsInterface(fldPath); err != nil {
		return
	}
	return utils.IfaceAsString(val), nil
}
func (dk *APDiktat) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, utils.ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, utils.ErrNotFound
	case utils.Path:
		return dk.Path, nil
	case utils.Value:
		return dk.Value, nil
	}
}
