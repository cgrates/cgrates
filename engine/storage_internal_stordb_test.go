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
	"slices"
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
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

func TestIDBGetTpTableIds(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)

	thresholds := []*utils.TPThresholdProfile{
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threshold1",
			FilterIDs: []string{"FLTR_1", "FLTR_2"},
			MaxHits:   -1,
			MinSleep:  "1s",
			Blocker:   true,
			Weight:    10,
			ActionIDs: []string{"Thresh1"},
			Async:     true,
		},
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Threshold2",
			FilterIDs: []string{"FilterID1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2014-07-14T14:35:00Z",
				ExpiryTime:     "",
			},
			MaxHits:   12,
			MinHits:   10,
			MinSleep:  "1s",
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"WARN3"},
		},
	}

	if err := storDB.SetTPThresholds(thresholds); err != nil {
		t.Error(err)
	}
	expIds := []string{"cgrates.org:Threshold1", "cgrates.org:Threshold2"}

	if rcvIds, err := storDB.GetTpTableIds("TH1", utils.TBLTPThresholds, utils.TPDistinctIds{"id"}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if sort.Slice(rcvIds, func(i, j int) bool {
		return rcvIds[i] < rcvIds[j]
	}); !slices.Equal(expIds, rcvIds) {
		t.Errorf("Expected %v,Received %v", expIds, rcvIds)
	}

	rateProfiles := []*utils.TPRatingProfile{
		{
			TPid:     "TH1",
			LoadId:   "TEST_LOADID",
			Tenant:   "cgrates.org",
			Category: "call",
			Subject:  "1001",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2022-01-14T00:00:00Z",
					RatingPlanId:     "TEST_RPLAN1",
					FallbackSubjects: "subj1"},
			},
		},
		{
			TPid:     "TH1",
			LoadId:   "TEST_SQL",
			Tenant:   "cgrates.org",
			Category: "sms",
			Subject:  "1007",
			RatingPlanActivations: []*utils.TPRatingActivation{
				{
					ActivationTime:   "2023-07-29T15:00:00Z",
					RatingPlanId:     "PlanOne",
					FallbackSubjects: "FallBack",
				},
			},
		},
	}
	if err := storDB.SetTPRatingProfiles(rateProfiles); err != nil {
		t.Error(err)
	}

	erpPrf := []string{"TEST_LOADID", "TEST_SQL"}

	if rcvIds, err := storDB.GetTpTableIds("TH1", utils.TBLTPRateProfiles, utils.TPDistinctIds{"loadId"}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if sort.Slice(rcvIds, func(i, j int) bool {
		return rcvIds[i] < rcvIds[j]
	}); !slices.Equal(erpPrf, rcvIds) {
		t.Errorf("Expected %v,Received %v", erpPrf, rcvIds)
	}
	resources := []*utils.TPResourceProfile{
		{
			Tenant:             "cgrates.org",
			TPid:               "TH1",
			ID:                 "ResGroup1",
			FilterIDs:          []string{"FLTR_RES_GR_1"},
			ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2022-07-29T15:00:00Z"},
			Stored:             false,
			Blocker:            false,
			Weight:             10,
			Limit:              "2",
			ThresholdIDs:       []string{"TRes1"},
			AllocationMessage:  "asd",
		},
	}
	if err := storDB.SetTPResources(resources); err != nil {
		t.Error(err)
	}

	if resIds, err := storDB.GetTpTableIds("TH1", utils.TBLTPResources, utils.TPDistinctIds{}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if !slices.Equal(resIds, []string{utils.ConcatenatedKey(resources[0].Tenant, resources[0].ID)}) {
		t.Errorf("Expected : %v,Received: %v ", []string{utils.ConcatenatedKey(resources[0].Tenant, resources[0].ID)}, resIds)
	}

	stats := []*utils.TPStatProfile{
		{
			TPid:      "TH1",
			Tenant:    "cgrates.org",
			ID:        "Stats1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2021-07-29T15:00:00Z",
			},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*asr",
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20.0,
			Stored:       true,
			MinItems:     1,
		},
	}
	if err := storDB.SetTPStats(stats); err != nil {
		t.Error(err)
	}
	if ids, err := storDB.GetTpTableIds("TH1", utils.TBLTPStats, utils.TPDistinctIds{}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if !slices.Equal(ids, []string{utils.ConcatenatedKey(stats[0].Tenant, stats[0].ID)}) {
		t.Errorf("Expected %v, Received %v", []string{utils.ConcatenatedKey(stats[0].Tenant, stats[0].ID)}, utils.ToJSON(ids))
	}
	if err := storDB.RemTpData(utils.TBLTPThresholds, "TH1", map[string]string{"tag": utils.ConcatenatedKey(thresholds[0].Tenant, thresholds[0].ID)}); err != nil {
		t.Error(err)
	}
	if rcvIds, err := storDB.GetTpTableIds("TH1", utils.TBLTPThresholds, utils.TPDistinctIds{"id"}, nil, &utils.PaginatorWithSearch{}); err != nil {
		t.Error(err)
	} else if !slices.Equal(expIds[1:], rcvIds) {
		t.Errorf("Expected %v,Received %v", expIds[1:], rcvIds)
	}

	if err := storDB.RemTpData(utils.EmptyString, "TH1", nil); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPThresholds("TH1", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func TestIDBGetPaginator(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	destRates := []*utils.TPDestinationRate{
		{TPid: "TP1",
			ID: "P1",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "NAT",
					RateId:           "R4",
					Rate:             csvr.rates["R4"],
					RoundingMethod:   utils.ROUNDING_MIDDLE,
					RoundingDecimals: 4,
				},
			},
		}, {
			TPid: "TP1",
			ID:   "RT_DEFAULT",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "ALL",
					RateId:           "R2",
					Rate:             csvr.rates["R2"],
					RoundingMethod:   utils.ROUNDING_MIDDLE,
					RoundingDecimals: 4,
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "RT_STANDARD",
			DestinationRates: []*utils.DestinationRate{
				{
					DestinationId:    "GERMANY",
					RateId:           "R1",
					Rate:             csvr.rates["R1"],
					RoundingMethod:   utils.ROUNDING_MIDDLE,
					RoundingDecimals: 4,
				},
			},
		},
	}
	if err := storDB.SetTPDestinationRates(destRates); err != nil {
		t.Error(err)
	}
	if ids, err := storDB.GetTPDestinationRates("TP1", utils.EmptyString, &utils.Paginator{Limit: utils.IntPointer(1), Offset: utils.IntPointer(1)}); err != nil {
		t.Error(err)
	} else if len(ids) != 1 {
		t.Errorf("Expected 1,Received :%v", len(ids))
	}
}
