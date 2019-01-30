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

package dispatchers

import (
	"sort"

	"github.com/cgrates/cgrates/utils"
)

// DispatcherProfile is the config for one Dispatcher
type DispatcherProfile struct {
	Tenant             string
	ID                 string
	Subsystems         []string
	FilterIDs          []string
	ActivationInterval *utils.ActivationInterval // Activation interval
	Strategy           string
	StrategyParams     map[string]interface{} // ie for distribution, set here the pool weights
	ConnIDs            []string               // dispatch to these connections
	Weight             float64
}

func (dP *DispatcherProfile) TenantID() string {
	return utils.ConcatenatedKey(dP.Tenant, dP.ID)
}

// DispatcherProfiles is a sortable list of Dispatcher profiles
type DispatcherProfiles []*DispatcherProfile

// Sort is part of sort interface, sort based on Weight
func (dps DispatcherProfiles) Sort() {
	sort.Slice(dps, func(i, j int) bool { return dps[i].Weight > dps[j].Weight })
}
