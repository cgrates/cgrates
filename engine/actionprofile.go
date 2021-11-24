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
	"strings"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// ActionProfile represents the configuration of a Action profile
type ActionProfile struct {
	Tenant    string
	ID        string
	FilterIDs []string
	Weight    float64
	Schedule  string
	Targets   map[string]utils.StringSet

	Actions []*APAction
}

func (aP *ActionProfile) TenantID() string {
	return utils.ConcatenatedKey(aP.Tenant, aP.ID)
}

// ActionProfiles is a sortable list of ActionProfiles
type ActionProfiles []*ActionProfile

// Sort is part of sort interface, sort based on Weight
func (aps ActionProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// APAction defines action related information used within a ActionProfile
type APAction struct {
	ID        string                 // Action ID
	FilterIDs []string               // Action FilterIDs
	Blocker   bool                   // Blocker will stop further actions running in the chain
	TTL       time.Duration          // Cancel Action if not executed within TTL
	Type      string                 // Type of Action
	Opts      map[string]interface{} // Extra options to pass depending on action type
	Diktats   []*APDiktat
}

type APDiktat struct {
	Path  string // Path to execute
	Value string // Value to execute on Path

	valRSR config.RSRParsers
}

// RSRValues returns the Value as RSRParsers
func (dk *APDiktat) RSRValues(sep string) (v config.RSRParsers, err error) {
	if dk.valRSR == nil {
		dk.valRSR, err = config.NewRSRParsers(dk.Value, sep)
		if err != nil {
			return
		}
	}
	return dk.valRSR, nil
}

// ActionProfileWithAPIOpts is used in API calls

type ActionProfileWithAPIOpts struct {
	*ActionProfile
	APIOpts map[string]interface{}
}

func (aP *ActionProfile) Set(path []string, val interface{}, newBranch bool, _ string) (err error) {
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
		case utils.Weight:
			if val != utils.EmptyString {
				aP.Weight, err = utils.IfaceAsFloat64(val)
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
	if strings.HasPrefix(path[0], utils.Actions) &&
		path[0][5] == '[' && path[0][len(path[0])-1] == ']' {
		acID = path[0][6 : len(path[0])-1]
	} else if path[0] == utils.Actions {
		acID = path[1]
		path = path[1:]
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
		ac = &APAction{ID: acID, Opts: make(map[string]interface{})}
		aP.Actions = append(aP.Actions, ac)
	}

	return ac.Set(path[1:], val, newBranch)
}

func (aP *APAction) Set(path []string, val interface{}, newBranch bool) (err error) {
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
		case utils.Blocker:
			aP.Blocker, err = utils.IfaceAsBool(val)
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
