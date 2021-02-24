/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package engine

import (
	"sort"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

// ActionProfile represents the configuration of a Action profile
type ActionProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval
	Weight             float64
	Schedule           string
	Targets            map[string]utils.StringSet

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
	Path  string            // Path to execute
	Value config.RSRParsers // Value to execute on path
}

// ActionProfileWithOpts is used in API calls
type ActionProfileWithOpts struct {
	*ActionProfile
	Opts map[string]interface{}
}
