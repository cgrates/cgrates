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

import "sort"

type AccountProfile struct {
	Tenant             string
	ID                 string
	FilterIDs          []string
	ActivationInterval *ActivationInterval
	Weight             float64

	Balances []*Balance
}

type Balance struct {
	ID        string
	FilterIDs []string
	Weight    float64
	Blocker   bool
	Type      string
	Opts      map[string]interface{}
	Value     float64
}

func (aP *AccountProfile) TenantID() string {
	return ConcatenatedKey(aP.Tenant, aP.ID)
}

// ActionProfiles is a sortable list of ActionProfiles
type AccountProfiles []*AccountProfile

// Sort is part of sort interface, sort based on Weight
func (aps AccountProfiles) Sort() {
	sort.Slice(aps, func(i, j int) bool { return aps[i].Weight > aps[j].Weight })
}

// AccountProfileWithOpts is used in API calls
type AccountProfileWithOpts struct {
	*AccountProfile
	Opts map[string]interface{}
}

type Account struct {
	Tenant string
	ID     string
}

func (ac *Account) TenantID() string {
	return ConcatenatedKey(ac.Tenant, ac.ID)
}
