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
	"maps"
	"slices"
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

// Clone clones ActionProfile
func (ap *ActionProfile) Clone() *ActionProfile {
	if ap == nil {
		return nil
	}
	cloned := &ActionProfile{
		Tenant:   ap.Tenant,
		ID:       ap.ID,
		Schedule: ap.Schedule,
	}
	if ap.FilterIDs != nil {
		cloned.FilterIDs = make([]string, len(ap.FilterIDs))
		copy(cloned.FilterIDs, ap.FilterIDs)
	}
	if ap.Weights != nil {
		cloned.Weights = ap.Weights.Clone()
	}
	if ap.Blockers != nil {
		cloned.Blockers = ap.Blockers.Clone()
	}
	if ap.Targets != nil {
		cloned.Targets = make(map[string]StringSet)
		for k, v := range ap.Targets {
			cloned.Targets[k] = v.Clone()
		}
	}
	if ap.Actions != nil {
		cloned.Actions = make([]*APAction, len(ap.Actions))
		for i, action := range ap.Actions {
			if action != nil {
				cloned.Actions[i] = action.Clone()
			}
		}
	}
	return cloned
}

// CacheClone returns a clone of ActionProfile used by ltcache CacheCloner
func (ap *ActionProfile) CacheClone() any {
	return ap.Clone()
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

// AsMapStringInterface converts ActionProfile struct to map[string]any
func (ap *ActionProfile) AsMapStringInterface() map[string]any {
	if ap == nil {
		return nil
	}
	return map[string]any{
		Tenant:    ap.Tenant,
		ID:        ap.ID,
		FilterIDs: ap.FilterIDs,
		Weights:   ap.Weights,
		Blockers:  ap.Blockers,
		Schedule:  ap.Schedule,
		Targets:   ap.Targets,
		Actions:   ap.Actions,
	}
}

// MapStringInterfaceToActionProfile converts map[string]any to ActionProfile struct
func MapStringInterfaceToActionProfile(m map[string]any) (ap *ActionProfile, err error) {
	ap = &ActionProfile{}
	if v, ok := m[Tenant].(string); ok {
		ap.Tenant = v
	}
	if v, ok := m[ID].(string); ok {
		ap.ID = v
	}
	ap.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	ap.Weights = InterfaceToDynamicWeights(m[Weights])
	ap.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	if v, ok := m[Schedule].(string); ok {
		ap.Schedule = v
	}
	ap.Targets = InterfaceToMapStringStringSet(m[Targets])
	if ap.Actions, err = InterfaceToAPActions(m[Actions]); err != nil {
		return nil, err
	}
	return ap, nil
}

// InterfaceToAPActions converts any to []*APAction
func InterfaceToAPActions(m any) ([]*APAction, error) {
	v, ok := m.([]any)
	if !ok {
		return nil, nil
	}

	actions := make([]*APAction, 0, len(v))
	for _, actionAny := range v {
		if actionMap, ok := actionAny.(map[string]any); ok {
			action, err := MapToAPAction(actionMap)
			if err != nil {
				return nil, err
			}
			actions = append(actions, action)
		}
	}
	return actions, nil
}

// MapToAPAction converts map[string]any to APAction struct
func MapToAPAction(m map[string]any) (*APAction, error) {
	action := &APAction{}
	if v, ok := m[ID].(string); ok {
		action.ID = v
	}
	action.FilterIDs = InterfaceToStringSlice(m[FilterIDs])
	if v, ok := m[TTL].(string); ok {
		if dur, err := time.ParseDuration(v); err != nil {
			return nil, err
		} else {
			action.TTL = dur
		}
	} else if v, ok := m[TTL].(float64); ok { // for -1 cases
		action.TTL = time.Duration(v)
	}
	if v, ok := m[Type].(string); ok {
		action.Type = v
	}
	if v, ok := m[Opts].(map[string]any); ok {
		action.Opts = v
	}
	action.Weights = InterfaceToDynamicWeights(m[Weights])
	action.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	if v, ok := m[Diktats].([]any); ok {
		action.Diktats = make([]*APDiktat, 0, len(v))
		for _, diktatAny := range v {
			if diktatMap, ok := diktatAny.(map[string]any); ok {
				diktat, err := MapToAPDiktat(diktatMap)
				if err != nil {
					return nil, err
				}
				action.Diktats = append(action.Diktats, diktat)
			}
		}
	}
	return action, nil
}

// MapToAPDiktat converts map[string]any to APDiktat struct
func MapToAPDiktat(m map[string]any) (*APDiktat, error) {
	diktat := &APDiktat{}
	if v, ok := m[ID].(string); ok {
		diktat.ID = v
	}
	diktat.FilterIDs = InterfaceToStringSlice(m[FilterIDs])

	if v, ok := m[Opts].(map[string]any); ok {
		diktat.Opts = v
	}
	diktat.Weights = InterfaceToDynamicWeights(m[Weights])
	diktat.Blockers = InterfaceToDynamicBlockers(m[Blockers])
	return diktat, nil
}

// APAction defines action related information used within an ActionProfile.
type APAction struct {
	ID        string         // Action ID
	FilterIDs []string       // Action FilterIDs
	TTL       time.Duration  // Cancel Action if not executed within TTL
	Type      string         // Type of Action
	Opts      map[string]any // Extra options to pass depending on action type
	Weights   DynamicWeights
	Blockers  DynamicBlockers
	Diktats   []*APDiktat
}

// Clone clones APAction
func (a *APAction) Clone() *APAction {
	if a == nil {
		return nil
	}
	cloned := &APAction{
		ID:   a.ID,
		TTL:  a.TTL,
		Type: a.Type,
	}
	if a.FilterIDs != nil {
		cloned.FilterIDs = make([]string, len(a.FilterIDs))
		copy(cloned.FilterIDs, a.FilterIDs)
	}
	if a.Opts != nil {
		cloned.Opts = make(map[string]any, len(a.Opts))
		maps.Copy(cloned.Opts, a.Opts)
	}
	if a.Weights != nil {
		cloned.Weights = a.Weights.Clone()
	}
	if a.Blockers != nil {
		cloned.Blockers = a.Blockers.Clone()
	}
	if a.Diktats != nil {
		cloned.Diktats = make([]*APDiktat, len(a.Diktats))
		for i, diktat := range a.Diktats {
			if diktat != nil {
				cloned.Diktats[i] = diktat.Clone()
			}
		}
	}
	return cloned
}

// Set implements the profile interface, setting values in APAction based on path.
func (a *APAction) Set(path []string, val any, newBranch bool) (err error) {
	switch len(path) {
	default:
		if path[0] == Opts {
			return MapStorage(a.Opts).Set(path[1:], val)
		}
		if path[0] == Diktats && path[1] == Opts {
			if len(a.Diktats) == 0 || newBranch {
				a.Diktats = append(a.Diktats, new(APDiktat))
			}
			return MapStorage(a.Diktats[len(a.Diktats)-1].Opts).Set(path[2:], val)
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
		case Weights:
			if val != EmptyString {
				a.Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
		case Blockers:
			if val != EmptyString {
				a.Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
			}
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
			default:
				if strings.HasPrefix(path[1], Opts) &&
					path[1][4] == '[' && path[1][len(path[1])-1] == ']' {
					a.Opts[path[1][5:len(path[1])-1]] = val
					return
				}
				return ErrWrongPath
			case ID:
				if dID := IfaceAsString(val); dID == EmptyString {
					return ErrWrongPath
				} else {
					a.Diktats[len(a.Diktats)-1].ID = dID
				}
			case FilterIDs:
				var valA []string
				valA, err = IfaceAsStringSlice(val)
				a.Diktats[len(a.Diktats)-1].FilterIDs = append(a.Diktats[len(a.Diktats)-1].
					FilterIDs, valA...)
			case Opts:
				a.Diktats[len(a.Diktats)-1].Opts, err = NewMapFromCSV(IfaceAsString(val))
			case Weights:
				if val != EmptyString {
					a.Diktats[len(a.Diktats)-1].Weights, err = NewDynamicWeightsFromString(IfaceAsString(val), InfieldSep, ANDSep)
				}
			case Blockers:
				if val != EmptyString {
					a.Diktats[len(a.Diktats)-1].Blockers, err = NewDynamicBlockersFromString(IfaceAsString(val), InfieldSep, ANDSep)
				}
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
	a.FilterIDs = append(a.FilterIDs, v2.FilterIDs...)
	if v2.TTL != 0 {
		a.TTL = v2.TTL
	}
	if len(v2.Type) != 0 {
		a.Type = v2.Type
	}
	maps.Copy(a.Opts, v2.Opts)
	if v2.Blockers != nil {
		a.Blockers = append(a.Blockers, v2.Blockers...)
	}
	if v2.Weights != nil {
		a.Weights = append(a.Weights, v2.Weights...)
	}
	if len(a.Diktats) == 1 && len(a.Diktats[0].Opts) == 0 {
		a.Diktats = a.Diktats[:0]
	}
	for _, diktatV2 := range v2.Diktats {
		if idx := slices.IndexFunc(a.Diktats, func(a *APDiktat) bool {
			return a.ID == diktatV2.ID
		}); idx != -1 {
			a.Diktats[idx].Merge(diktatV2)
			continue
		}
		a.Diktats = append(a.Diktats, diktatV2)
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
		case Weights:
			return a.Weights, nil
		case Blockers:
			return a.Blockers, nil
		}
	case 2:
		if fld, _ := GetPathIndexString(fldPath[0]); fld != Opts {
			return nil, ErrNotFound
		}
		return MapStorage(a.Opts).FieldAsInterface(fldPath[1:])
	case 3:
		if fld, idxStr := GetPathIndexString(fldPath[0]); fld != Diktats {
			return nil, ErrNotFound
		} else {
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
	ID        string         // Diktat ID
	FilterIDs []string       // Diktat FilterIDs
	Opts      map[string]any // Diktat options to pass
	Weights   DynamicWeights
	Blockers  DynamicBlockers

	valRSR RSRParsers
}

// Clone clones APAction
func (d *APDiktat) Clone() *APDiktat {
	if d == nil {
		return nil
	}
	cloned := &APDiktat{
		ID: d.ID,
	}
	if d.FilterIDs != nil {
		cloned.FilterIDs = make([]string, len(d.FilterIDs))
		copy(cloned.FilterIDs, d.FilterIDs)
	}
	if d.Opts != nil {
		cloned.Opts = make(map[string]any, len(d.Opts))
		maps.Copy(cloned.Opts, d.Opts)
	}
	if d.Weights != nil {
		cloned.Weights = d.Weights.Clone()
	}
	if d.Blockers != nil {
		cloned.Blockers = d.Blockers.Clone()
	}
	if d.valRSR != nil {
		cloned.valRSR = d.valRSR.Clone()
	}
	return cloned
}

// Merge combines the values from another APDiktat into this one.
func (a *APDiktat) Merge(v2 *APDiktat) {
	a.FilterIDs = append(a.FilterIDs, v2.FilterIDs...)
	maps.Copy(a.Opts, v2.Opts)
	a.Blockers = append(a.Blockers, v2.Blockers...)
	a.Weights = append(a.Weights, v2.Weights...)
}

// RSRValues returns the Value as RSRParsers or creates new ones if not initialized.
func (dk *APDiktat) RSRValues() (RSRParsers, error) {
	if dk.valRSR == nil {
		if _, has := dk.Opts[MetaBalanceValue]; has {
			return NewRSRParsers(IfaceAsString(dk.Opts[MetaBalanceValue]), RSRSep)
		}
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
	switch len(fldPath) {
	default:
		if fld, idxStr := GetPathIndexString(fldPath[0]); fld == Opts {
			path := fldPath[1:]
			if idxStr != nil {
				path = append([]string{*idxStr}, path...)
			}
			return MapStorage(dk.Opts).FieldAsInterface(path)
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
					if idx < len(dk.FilterIDs) {
						return dk.FilterIDs[idx], nil
					}
				case Opts:
					return MapStorage(dk.Opts).FieldAsInterface([]string{*idxStr})
				}
			}
			return nil, ErrNotFound
		case ID:
			return dk.ID, nil
		case FilterIDs:
			return dk.FilterIDs, nil
		case Opts:
			return dk.Opts, nil
		case Weights:
			return dk.Weights, nil
		case Blockers:
			return dk.Blockers, nil
		}
	case 2:
		if fld, _ := GetPathIndexString(fldPath[0]); fld != Opts {
			return nil, ErrNotFound
		}
		return MapStorage(dk.Opts).FieldAsInterface(fldPath[1:])
	}
}
