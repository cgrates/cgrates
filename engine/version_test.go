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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestVersionCompare(t *testing.T) {
	x := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	y := Versions{utils.Accounts: 1, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	z := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 1,
		utils.SharedGroups: 2, utils.CostDetails: 2}
	q := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 1, utils.CostDetails: 2}
	c := Versions{utils.CostDetails: 1}
	a := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2,
		utils.SessionSCosts: 1}
	b := Versions{utils.Accounts: 2, utils.Actions: 2,
		utils.ActionTriggers: 2, utils.ActionPlans: 2,
		utils.SharedGroups: 2, utils.CostDetails: 2,
		utils.SessionSCosts: 2}
	message1 := y.Compare(x, utils.MetaMongo, true)
	if message1 != "cgr-migrator -exec=*accounts" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*accounts", message1)
	}
	message2 := z.Compare(x, utils.MetaMongo, true)
	if message2 != "cgr-migrator -exec=*action_plans" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*action_plans", message2)
	}
	message3 := q.Compare(x, utils.MetaMongo, true)
	if message3 != "cgr-migrator -exec=*shared_groups" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*shared_groups", message3)
	}
	message4 := c.Compare(x, utils.MetaMongo, false)
	if message4 != "cgr-migrator -exec=*cost_details" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*cost_details", message4)
	}
	message5 := a.Compare(b, utils.MetaMySQL, false)
	if message5 != "cgr-migrator -exec=*sessions_costs" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*sessions_costs", message5)
	}
	message6 := a.Compare(b, utils.MetaPostgres, false)
	if message6 != "cgr-migrator -exec=*sessions_costs" {
		t.Errorf("Error failed to compare to current version expected: %s received: %s", "cgr-migrator -exec=*sessions_costs", message6)
	}

}

func TestVarsionCheckVersions(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)

	err := CheckVersions(data)

	if err != nil {
		t.Error(err)
	}

	vrs := Versions{"test": 1}
	data.SetVersions(vrs, true)

	err = CheckVersions(data)

	if err != nil {
		if err.Error()[0] != 'M' {
			t.Error(err)
		}
	}

	vrs2 := Versions{
		utils.StatS:               3,
		utils.Accounts:            3,
		utils.Actions:             2,
		utils.ActionTriggers:      2,
		utils.ActionPlans:         3,
		utils.SharedGroups:        2,
		utils.Thresholds:          3,
		utils.Suppliers:           1,
		utils.Attributes:          5,
		utils.Timing:              1,
		utils.RQF:                 4,
		utils.Resource:            1,
		utils.Subscribers:         1,
		utils.Destinations:        1,
		utils.ReverseDestinations: 1,
		utils.RatingPlan:          1,
		utils.RatingProfile:       1,
		utils.Chargers:            1,
		utils.Dispatchers:         1,
		utils.LoadIDsVrs:          1,
		utils.CostDetails:         2,
		utils.SessionSCosts:       3,
		utils.CDRs:                2,
		utils.TpRatingPlans:       1,
		utils.TpFilters:           1,
		utils.TpDestinationRates:  1,
		utils.TpActionTriggers:    1,
		utils.TpAccountActionsV:   1,
		utils.TpActionPlans:       1,
		utils.TpActions:           1,
		utils.TpThresholds:        1,
		utils.TpSuppliers:         1,
		utils.TpStats:             1,
		utils.TpSharedGroups:      1,
		utils.TpRatingProfiles:    1,
		utils.TpResources:         1,
		utils.TpRates:             1,
		utils.TpTiming:            1,
		utils.TpResource:          1,
		utils.TpDestinations:      1,
		utils.TpRatingPlan:        1,
		utils.TpRatingProfile:     1,
		utils.TpChargers:          1,
		utils.TpDispatchers:       1,
	}
	data.SetVersions(vrs2, true)

	err = CheckVersions(data)

	if err != nil {
		t.Error(err)
	}
}

func TestVersionSetDBVersions(t *testing.T) {
	defaultCfg, _ := config.NewDefaultCGRConfig()
	data := NewInternalDB(nil, nil, true, defaultCfg.DataDbCfg().Items)

	err := SetDBVersions(data)

	if err != nil {
		t.Error(err)
	}
}

func TestVersionCurrentDBVersions(t *testing.T) {
	type args struct {
		storType string
		isDataDB bool
	}
	tests := []struct {
		name string
		args args
		exp  Versions
	}{
		{
			name: "CurrentDataDBVersions",
			args: args{
				storType: utils.MetaMongo,
				isDataDB: true,
			},
			exp: CurrentDataDBVersions(),
		},
		{
			name: "CurrentStorDBVersions",
			args: args{
				storType: utils.MetaMongo,
				isDataDB: false,
			},
			exp: CurrentStorDBVersions(),
		},
		{
			name: "CurrentStorDBVersions",
			args: args{
				storType: utils.MetaPostgres,
				isDataDB: false,
			},
			exp: CurrentStorDBVersions(),
		},
		{
			name: "CurrentDataDBVersions",
			args: args{
				storType: utils.MetaRedis,
				isDataDB: false,
			},
			exp: CurrentDataDBVersions(),
		},
		{
			name: "default",
			args: args{
				storType: str,
				isDataDB: false,
			},
			exp: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := CurrentDBVersions(tt.args.storType, tt.args.isDataDB)

			if !reflect.DeepEqual(rcv, tt.exp) {
				t.Errorf("expected %s, received %s", utils.ToJSON(tt.exp), utils.ToJSON(rcv))
			}
		})
	}
}

func TestVersionCompare2(t *testing.T) {
	vers := Versions{}
	current := Versions{}

	rcv := vers.Compare(current, utils.MetaInternal, false)

	if rcv != "" {
		t.Error(rcv)
	}

	rcv = vers.Compare(current, utils.MetaRedis, false)

	if rcv != "" {
		t.Error(rcv)
	}
}

func TestSetRoundingDecimals(t *testing.T) {
	initialRoundingDecimals := globalRoundingDecimals
	testCases := []struct {
		name        string
		roundingDec int
		expected    int
	}{
		{name: "Positive rounding decimals", roundingDec: 2., expected: 2},
		{name: "Zero rounding decimals", roundingDec: 0, expected: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetRoundingDecimals(tc.roundingDec)
			if globalRoundingDecimals != tc.expected {
				t.Errorf("Expected rounding decimals to be %d, got %d", tc.expected, globalRoundingDecimals)
			}
			globalRoundingDecimals = initialRoundingDecimals
		})
	}
}

func TestCallDescriptor_AddRatingInfo(t *testing.T) {
	initialRatingInfos := []*RatingInfo{}
	cd := &CallDescriptor{RatingInfos: initialRatingInfos}
	ratingInfo1 := &RatingInfo{}
	ratingInfo2 := &RatingInfo{}
	cd.AddRatingInfo(ratingInfo1, ratingInfo2)
	expectedLength := len(initialRatingInfos) + 2
	if len(cd.RatingInfos) != expectedLength {
		t.Errorf("Expected RatingInfos length to be %d, got %d", expectedLength, len(cd.RatingInfos))
	}
	if cd.RatingInfos[0] != ratingInfo1 || cd.RatingInfos[1] != ratingInfo2 {
		t.Errorf("Added RatingInfo objects don't match expectations")
	}
}
