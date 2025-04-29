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
	"strconv"
	"strings"
	"time"
)

// ActionProfile represents the configuration of an Action profile.
type ActionProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Weights   DynamicWeights
	Blockers  DynamicBlockers
	Schedule  string
	Targets   map[string]StringSet

	Actions []*APAction
}

// ActionProfileWithAPIOpts wraps ActionProfile with APIOpts.
type ActionProfileWithAPIOpts struct {
	*ActionProfile
	APIOpts map[string]any
}

// TenantID returns the concatenated tenant and ID.
func (ap *ActionProfile) TenantID() string {
	return ConcatenatedKey(ap.Tenant, ap.ID)
}

// Set implements the profile interface, setting values in ActionProfile based on path.
func (ap *ActionProfile) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	case 0:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], Targets) &&
				path[0][7] == '[' && path[0][len(path[0])-1] == ']' {
				var valA []string
				valA, err = IfaceAsStringSlice(val)
				ap.Targets[path[0][8:len(path[0])-1]] = JoinStringSet(ap.Targets[path[0][8:len(path[0])-1]], NewStringSet(valA))
				return
			}
			return ErrWrongPath
		case Tenant:
			ap.Tenant = IfaceAsString(val)
		case ID:
			ap.ID = IfaceAsString(val)
		case Schedule:
			ap.Schedule = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.FilterIDs = append(ap.FilterIDs, valA...)
		case Weights:
			if val != EmptyString {
				ap.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				ap.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		}
		return
	case 2:
		if path[0] == Targets {
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			ap.Targets[path[1]] = JoinStringSet(ap.Targets[path[1]], NewStringSet(valA))
			return
		}
	default:
	}

	var acID string
	if path[0] == Actions {
		acID = path[1]
		path = path[1:]
	} else if strings.HasPrefix(path[0], Actions) &&
		path[0][7] == '[' && path[0][len(path[0])-1] == ']' {
		acID = path[0][8 : len(path[0])-1]
	}
	if acID == EmptyString {
		return ErrWrongPath
	}

	var ac *APAction
	for _, a := range ap.Actions {
		if a.ID == acID {
			ac = a
			break
		}
	}
	if ac == nil {
		ac = &APAction{ID: acID, Opts: make(map[string]any)}
		ap.Actions = append(ap.Actions, ac)
	}
	return ac.Set(path[1:], val, newBranch)
}

// Merge implements the profile interface, merging values from another ActionProfile.
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
		if k == EmptyString {
			continue
		}
		ap.Targets[k] = v
	}
}

// String implements the DataProvider interface, returning the ActionProfile in JSON format.
func (ap *ActionProfile) String() string { return ToJSON(ap) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (ap *ActionProfile) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = ap.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (ap *ActionProfile) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) == 1 {
		switch fldPath[0] {
		default:
			fld, idxStr := GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case Actions:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.Actions) {
						return ap.Actions[idx], nil
					}
				case FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(ap.FilterIDs) {
						return ap.FilterIDs[idx], nil
					}
				case Targets:
					if tr, has := ap.Targets[*idxStr]; has {
						return tr, nil
					}
				}
			}
			return nil, ErrNotFound
		case Tenant:
			return ap.Tenant, nil
		case ID:
			return ap.ID, nil
		case FilterIDs:
			return ap.FilterIDs, nil
		case Weights:
			return ap.Weights, nil
		case Blockers:
			return ap.Blockers, nil
		case Actions:
			return ap.Actions, nil
		case Schedule:
			return ap.Schedule, nil
		case Targets:
			return ap.Targets, nil
		}
	}
	if len(fldPath) == 0 {
		return nil, ErrNotFound
	}
	fld, idxStr := GetPathIndexString(fldPath[0])
	switch fld {
	default:
		return nil, ErrNotFound
	case Actions:
		if idxStr == nil {
			return nil, ErrNotFound
		}
		var idx int
		if idx, err = strconv.Atoi(*idxStr); err != nil {
			return
		}
		if idx >= len(ap.Actions) {
			return nil, ErrNotFound
		}
		return ap.Actions[idx].FieldAsInterface(fldPath[1:])
	case Targets:
		tr, has := ap.Targets[*idxStr]
		if !has {
			return nil, ErrNotFound
		}
		return tr.FieldAsInterface(fldPath[1:])
	}
}

// APAction defines action related information used within an ActionProfile.
type APAction struct {
	ID        string         // Action ID
	FilterIDs []string       // Action FilterIDs
	TTL       time.Duration  // Cancel Action if not executed within TTL
	Type      string         // Type of Action
	Opts      map[string]any // Extra options to pass depending on action type
	Diktats   []*APDiktat
}

// Set implements the profile interface, setting values in APAction based on path.
func (a *APAction) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		if path[0] == Opts {
			return MapStorage(a.Opts).Set(path[1:], val)
		}
		return ErrWrongPath
	case 0:
		return ErrWrongPath
	case 1:
		switch path[0] {
		default:
			if strings.HasPrefix(path[0], Opts) &&
				path[0][4] == '[' && path[0][len(path[0])-1] == ']' {
				a.Opts[path[0][5:len(path[0])-1]] = val
				return
			}
			return ErrWrongPath
		case ID:
			a.ID = IfaceAsString(val)
		case Type:
			a.Type = IfaceAsString(val)
		case FilterIDs:
			var valA []string
			valA, err = IfaceAsStringSlice(val)
			a.FilterIDs = append(a.FilterIDs, valA...)
		case TTL:
			a.TTL, err = IfaceAsDuration(val)
		case Opts:
			a.Opts, err = NewMapFromCSV(IfaceAsString(val))
		}
	case 2:
		switch path[0] {
		default:
			return ErrWrongPath
		case Opts:
			return MapStorage(a.Opts).Set(path[1:], val)
		case Diktats:
			if len(a.Diktats) == 0 || newBranch {
				a.Diktats = append(a.Diktats, new(APDiktat))
			}
			switch path[1] {
			case Path:
				a.Diktats[len(a.Diktats)-1].Path = IfaceAsString(val)
			case Value:
				a.Diktats[len(a.Diktats)-1].Value = IfaceAsString(val)
			}
		}
	}
	return
}

// Merge combines the values from another APAction into this one.
func (a *APAction) Merge(v2 *APAction) {
	if len(v2.ID) != 0 {
		a.ID = v2.ID
	}
	if v2.TTL != 0 {
		a.TTL = v2.TTL
	}
	if len(v2.Type) != 0 {
		a.Type = v2.Type
	}
	for key, value := range v2.Opts {
		a.Opts[key] = value
	}
	a.FilterIDs = append(a.FilterIDs, v2.FilterIDs...)
	if len(a.Diktats) == 1 && a.Diktats[0].Path == EmptyString {
		a.Diktats = a.Diktats[:0]
	}
	for _, diktat := range v2.Diktats {
		if diktat.Path != EmptyString {
			a.Diktats = append(a.Diktats, diktat)
		}
	}
}

// String implements the DataProvider interface, returning the APAction in JSON format.
func (a *APAction) String() string { return ToJSON(a) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (a *APAction) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = a.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (a *APAction) FieldAsInterface(fldPath []string) (_ any, err error) {
	switch len(fldPath) {
	default:
		if fld, idxStr := GetPathIndexString(fldPath[0]); fld == Opts {
			path := fldPath[1:]
			if idxStr != nil {
				path = append([]string{*idxStr}, path...)
			}
			return MapStorage(a.Opts).FieldAsInterface(path)
		}
		fallthrough
	case 0:
		return nil, ErrNotFound
	case 1:
		switch fldPath[0] {
		default:
			fld, idxStr := GetPathIndexString(fldPath[0])
			if idxStr != nil {
				switch fld {
				case FilterIDs:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(a.FilterIDs) {
						return a.FilterIDs[idx], nil
					}
				case Diktats:
					var idx int
					if idx, err = strconv.Atoi(*idxStr); err != nil {
						return
					}
					if idx < len(a.Diktats) {
						return a.Diktats[idx], nil
					}
				case Opts:
					return MapStorage(a.Opts).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, ErrNotFound
		case ID:
			return a.ID, nil
		case FilterIDs:
			return a.FilterIDs, nil
		case TTL:
			return a.TTL, nil
		case Diktats:
			return a.Diktats, nil
		case Type:
			return a.Type, nil
		case Opts:
			return a.Opts, nil
		}
	case 2:
		fld, idxStr := GetPathIndexString(fldPath[0])
		switch fld {
		default:
			return nil, ErrNotFound
		case Opts:
			path := fldPath[1:]
			if idxStr != nil {
				path = append([]string{*idxStr}, path...)
			}
			return MapStorage(a.Opts).FieldAsInterface(path)
		case Diktats:
			if idxStr == nil {
				return nil, ErrNotFound
			}
			var idx int
			if idx, err = strconv.Atoi(*idxStr); err != nil {
				return
			}
			if idx >= len(a.Diktats) {
				return nil, ErrNotFound
			}
			return a.Diktats[idx].FieldAsInterface(fldPath[1:])
		}
	}
}

// APDiktat defines a path and value operation to be executed by an action.
type APDiktat struct {
	Path  string // Path to execute
	Value string // Value to execute on Path

	valRSR RSRParsers
}

// RSRValues returns the Value as RSRParsers or creates new ones if not initialized.
func (dk *APDiktat) RSRValues() (RSRParsers, error) {
	if dk.valRSR == nil {
		return NewRSRParsers(dk.Value, RSRSep)
	}
	return dk.valRSR, nil
}

// String implements the DataProvider interface, returning the APDiktat in JSON format.
func (dk *APDiktat) String() string { return ToJSON(dk) }

// FieldAsString implements the DataProvider interface, retrieving field value as string.
func (dk *APDiktat) FieldAsString(fldPath []string) (_ string, err error) {
	var val any
	if val, err = dk.FieldAsInterface(fldPath); err != nil {
		return
	}
	return IfaceAsString(val), nil
}

// FieldAsInterface implements the DataProvider interface, retrieving field value as interface.
func (dk *APDiktat) FieldAsInterface(fldPath []string) (_ any, err error) {
	if len(fldPath) != 1 {
		return nil, ErrNotFound
	}
	switch fldPath[0] {
	default:
		return nil, ErrNotFound
	case Path:
		return dk.Path, nil
	case Value:
		return dk.Value, nil
	}
}
