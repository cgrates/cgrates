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
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestStorageInternalStorDBSetNilFields(t *testing.T) {
	iDB := &InternalDB{}
	tests := []struct {
		name string
		rcv  error
		exp  string
	}{
		{
			name: "SetTPTimings",
			rcv:  iDB.SetTPTimings([]*utils.ApierTPTiming{}),
			exp:  "",
		},
		{
			name: "SetTPDestinations",
			rcv:  iDB.SetTPDestinations([]*utils.TPDestination{}),
			exp:  "",
		},
		{
			name: "SetTPRates",
			rcv:  iDB.SetTPRates([]*utils.TPRate{}),
			exp:  "",
		},
		{
			name: "SetTPDestinationRates",
			rcv:  iDB.SetTPDestinationRates([]*utils.TPDestinationRate{}),
			exp:  "",
		},
		{
			name: "SetTPRatingPlans",
			rcv:  iDB.SetTPRatingPlans([]*utils.TPRatingPlan{}),
			exp:  "",
		},
		{
			name: "SetTPRatingProfiles",
			rcv:  iDB.SetTPRatingProfiles([]*utils.TPRatingProfile{}),
			exp:  "",
		},
		{
			name: "SetTPSharedGroups",
			rcv:  iDB.SetTPSharedGroups([]*utils.TPSharedGroups{}),
			exp:  "",
		},
		{
			name: "SetTPActions",
			rcv:  iDB.SetTPActions([]*utils.TPActions{}),
			exp:  "",
		},
		{
			name: "SetTPActionPlans",
			rcv:  iDB.SetTPActionPlans([]*utils.TPActionPlan{}),
			exp:  "",
		},
		{
			name: "SetTPActionTriggers",
			rcv:  iDB.SetTPActionTriggers([]*utils.TPActionTriggers{}),
			exp:  "",
		},
		{
			name: "SetTPAccountActions",
			rcv:  iDB.SetTPAccountActions([]*utils.TPAccountActions{}),
			exp:  "",
		},
		{
			name: "SetTPResources",
			rcv:  iDB.SetTPResources([]*utils.TPResourceProfile{}),
			exp:  "",
		},
		{
			name: "SetTPStats",
			rcv:  iDB.SetTPStats([]*utils.TPStatProfile{}),
			exp:  "",
		},
		{
			name: "SetTPFilters",
			rcv:  iDB.SetTPFilters([]*utils.TPFilterProfile{}),
			exp:  "",
		},
		{
			name: "SetTPSuppliers",
			rcv:  iDB.SetTPSuppliers([]*utils.TPSupplierProfile{}),
			exp:  "",
		},
		{
			name: "SetTPAttributes",
			rcv:  iDB.SetTPAttributes([]*utils.TPAttributeProfile{}),
			exp:  "",
		},
		{
			name: "SetTPChargers",
			rcv:  iDB.SetTPChargers([]*utils.TPChargerProfile{}),
			exp:  "",
		},
		{
			name: "SetTPDispatcherProfiles",
			rcv:  iDB.SetTPDispatcherProfiles([]*utils.TPDispatcherProfile{}),
			exp:  "",
		},
		{
			name: "SetTPDispatcherHosts",
			rcv:  iDB.SetTPDispatcherHosts([]*utils.TPDispatcherHost{}),
			exp:  "",
		},
		{
			name: "SetTPThresholds",
			rcv:  iDB.SetTPThresholds([]*utils.TPThresholdProfile{}),
			exp:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rcv != nil {
				if tt.rcv.Error() != tt.exp {
					t.Error(tt.rcv)
				}
			}
		})
	}
}

func TestStorageInternalStorDBNil(t *testing.T) {
	iDB := &InternalDB{}
	smCost := SMCost{}

	err := iDB.SetSMCost(&smCost)
	if err != nil {
		t.Error(err)
	}
}
