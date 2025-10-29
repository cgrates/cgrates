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

package engine

import (
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

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

func TestTPGetTRatingPlanPgt(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	if _, err := storDB.GetTPRatingPlans("TP1", "", &utils.Paginator{Limit: utils.IntPointer(1), Offset: utils.IntPointer(1)}); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
	ratingPlans := []*utils.TPRatingPlan{
		{
			TPid: "TP1",
			ID:   "RP_1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "DR_FREESWITCH_USERS",
					TimingId:           "ALWAYS",
					Weight:             10,
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "Plan1",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{
					DestinationRatesId: "RateId",
					TimingId:           "TimingID",
					Weight:             12,
				},
			},
		},
		{
			TPid: "TP1",
			ID:   "RP_UP",
			RatingPlanBindings: []*utils.TPRatingPlanBinding{
				{DestinationRatesId: "DR_UP", TimingId: utils.ANY, Weight: 10},
			},
		},
	}
	if err := storDB.SetTPRatingPlans(ratingPlans); err != nil {
		t.Error(err)
	}

	if rPlans, err := storDB.GetTPRatingPlans("TP1", "", &utils.Paginator{Limit: utils.IntPointer(1), Offset: utils.IntPointer(1)}); err != nil {
		t.Error(err)
	} else if len(rPlans) != 1 {
		t.Errorf("Expected 1,Recived %v", len(rPlans))
	}
}

func TestIDBSetCDR(t *testing.T) {
	storDB := NewInternalDB([]string{utils.Account, utils.Subject, utils.RunID, utils.CGRID}, []string{utils.Destination}, false, config.CgrConfig().StorDbCfg().Items)
	cdr := &CDR{
		CGRID:       "Cdr1",
		OrderID:     123,
		ToR:         utils.VOICE,
		OriginID:    "OriginCDR1",
		OriginHost:  "192.168.1.1",
		Source:      "test",
		RequestType: utils.META_RATED,
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "+4986517174963",
		RunID:       utils.MetaDefault,
		Usage:       time.Duration(12000000000),
		ExtraFields: map[string]string{},
		Cost:        1.01,
	}
	if err := storDB.SetCDR(cdr, true); err != nil {
		t.Error(err)
	}

	if _, _, err := storDB.GetCDRs(&utils.CDRsFilter{
		CGRIDs: []string{"Cdr1"}, Accounts: []string{"1001"}, ToRs: []string{utils.VOICE}, Subjects: []string{"1001"}, RunIDs: []string{utils.MetaDefault}, DestinationPrefixes: []string{"+49"}}, true); err != nil {
		t.Error(err)
	}
}

func TestIDBVersions(t *testing.T) {
	dataDB := NewInternalDB(nil, nil, true, config.CgrConfig().DataDbCfg().Items)
	if _, err := dataDB.GetVersions(utils.Accounts); err != utils.ErrNotFound {
		t.Error(err)
	}
	vrs := Versions{
		utils.Accounts:       3,
		utils.Actions:        2,
		utils.ActionTriggers: 2,
		utils.ActionPlans:    2,
		utils.SharedGroups:   2,
		utils.CostDetails:    1,
	}
	if err := dataDB.SetVersions(vrs, false); err != nil {
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	delete(vrs, utils.SharedGroups)
	if err := dataDB.SetVersions(vrs, true); err != nil { // overwrite
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	eAcnts := Versions{utils.Accounts: vrs[utils.Accounts]}
	if rcv, err := dataDB.GetVersions(utils.Accounts); err != nil { //query one element
		t.Error(err)
	} else if !reflect.DeepEqual(eAcnts, rcv) {
		t.Errorf("Expecting: %v, received: %v", eAcnts, rcv)
	}
	if _, err := dataDB.GetVersions("Not Avaible"); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
	eAcnts[utils.Accounts] = 2
	vrs[utils.Accounts] = eAcnts[utils.Accounts]
	if err := dataDB.SetVersions(eAcnts, false); err != nil { // change one element
		t.Error(err)
	}
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = dataDB.RemoveVersions(eAcnts); err != nil { // remove one element
		t.Error(err)
	}
	delete(vrs, utils.Accounts)
	if rcv, err := dataDB.GetVersions(""); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(vrs, rcv) {
		t.Errorf("Expecting: %v, received: %v", vrs, rcv)
	}
	if err = dataDB.RemoveVersions(nil); err != nil { // remove one element
		t.Error(err)
	}
	if _, err := dataDB.GetVersions(""); err != utils.ErrNotFound { //query non-existent
		t.Error(err)
	}
}

func TestIDBGetCDR(t *testing.T) {
	storDB := NewInternalDB([]string{"Account", utils.RunID, utils.Source, utils.ToR, "Subject", "OriginHost", "ExtraHeader1", "ExtraHeader2"}, []string{"Destination", "Header2"}, false, config.CgrConfig().StorDbCfg().Items)

	cdrS := []*CDR{
		{
			CGRID:       "CGR1",
			RunID:       utils.MetaRaw,
			OrderID:     time.Now().UnixNano(),
			OriginHost:  "127.0.0.1",
			Source:      "testSetCDRs",
			OriginID:    "testevent1",
			ToR:         "*voice",
			RequestType: utils.META_PREPAID,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1004",
			Subject:     "1004",
			Destination: "1007",
			SetupTime:   time.Date(2015, 12, 12, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2015, 12, 12, 14, 52, 20, 0, time.UTC),
			Usage:       35 * time.Second,
			ExtraFields: map[string]string{"ExtraHeader1": "ExtraVal1", "Header2": "Val2", "ExtraHeader2": "Val"},
			Cost:        12,
		},
		{
			ToR:         utils.VOICE,
			OriginID:    "testDspCDRsProcessExternalCDR",
			OriginHost:  "127.0.0.1",
			Source:      utils.UNIT_TEST,
			RequestType: utils.META_RATED,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1001",
			SetupTime:   time.Date(2021, 12, 12, 14, 52, 0, 0, time.UTC),
			AnswerTime:  time.Date(2021, 12, 12, 14, 52, 19, 0, time.UTC),
			Usage:       19 * time.Second,
			Cost:        18.7,
		}}

	for _, cdr := range cdrS {
		if err := storDB.SetCDR(cdr, false); err != nil {
			t.Error(err)
		}
	}

	if cdrs, _, err := storDB.GetCDRs(&utils.CDRsFilter{RunIDs: []string{utils.MetaRaw}, Subjects: []string{"1004"}, MaxUsage: "50000000000", OrderBy: utils.Usage}, false); err != nil {
		t.Error(err)
	} else if len(cdrs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %d", len(cdrs))
	}

	if cdrs, _, err := storDB.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1003"}, RequestTypes: []string{"*rated"}, MinUsage: "1000000000", OrderBy: utils.COST}, false); err != nil {
		t.Error(err)
	} else if len(cdrs) != 1 {
		t.Errorf("Unexpected number of CDRs returned: %d", len(cdrs))
	}
}

func TestIDBRemoveSMC(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
	// WRITE
	var snd = []*SMCost{
		{
			CGRID:       "CGRID1",
			RunID:       "11",
			OriginHost:  "host22",
			OriginID:    "O1",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID2",
			RunID:       "12",
			OriginHost:  "host22",
			OriginID:    "O2",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID3",
			RunID:       "13",
			OriginHost:  "host23",
			OriginID:    "O3",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := storDB.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "host22", ""); err != nil {
		t.Fatal(err)
	} else if len(rcv) != 2 {
		t.Errorf("Expected 2 results received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCosts(&utils.SMCostFilter{
		RunIDs:      []string{"12", "11"},
		NotRunIDs:   []string{"13"},
		OriginHosts: []string{"host22"},
	}); err != nil {
		t.Error(err)
	}
	// READ
	if rcv, err := storDB.GetSMCosts("", "", "", ""); err != nil {
		t.Error(err)
	} else if len(rcv) != 1 {
		t.Errorf("Expected 1 result received %v ", len(rcv))
	}
	// REMOVE
	if err := storDB.RemoveSMCost(snd[2]); err != nil {
		t.Error(err)
	}
	// READ
	if _, err := storDB.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func TestIDBSharedGroups(t *testing.T) {
	storDB := NewInternalDB(nil, nil, false, config.CgrConfig().StorDbCfg().Items)
	sharedGroups := []*utils.TPSharedGroups{
		{
			TPid: "TPS1",
			ID:   "Group1",
			SharedGroups: []*utils.TPSharedGroup{
				{
					Account:       "AccOne",
					Strategy:      "StrategyOne",
					RatingSubject: "SubOne",
				},
				{
					Account:       "AccTwo",
					Strategy:      "StrategyTwo",
					RatingSubject: "SubTwo",
				},
			},
		},
	}
	if err := storDB.SetTPSharedGroups(sharedGroups); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPSharedGroups("TPS1", "Group1"); err != nil {
		t.Error(err)
	}
	resources := []*utils.TPResourceProfile{
		{
			Tenant:            "cgrates.org",
			TPid:              "TPS1",
			ID:                "ResGroup1",
			FilterIDs:         []string{"FLTR_RES_GR_1"},
			Stored:            false,
			Blocker:           false,
			Weight:            10,
			Limit:             "2",
			ThresholdIDs:      []string{"TRes1"},
			AllocationMessage: "asd",
		}}
	if err := storDB.SetTPResources(resources); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPResources("TPS1", "cgrates.org", "ResGroup1"); err != nil {
		t.Error(err)
	}

	stats := []*utils.TPStatProfile{

		{
			TPid:        "TPS1",
			Tenant:      "cgrates.org",
			ID:          "Stat1",
			FilterIDs:   []string{"*string:Account:1002"},
			QueueLength: 100,
			TTL:         "1s",
			Metrics: []*utils.MetricWithFilters{
				{
					MetricID: "*tcc",
				},
				{
					MetricID: "*average#Usage",
				},
			},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     2,
			ThresholdIDs: []string{"Th1"},
		},
	}
	if err := storDB.SetTPStats(stats); err != nil {
		t.Error(err)
	}
	if _, err := storDB.GetTPStats("TPS1", "cgrates.org", "Stat1"); err != nil {
		t.Error(err)
	}

	actionPlans := []*utils.TPActionPlan{
		{TPid: "TPS1",
			ID: "PCK_10",
			ActionPlan: []*utils.TPActionTiming{
				{
					ActionsId: "TOPUP_RST_10",
					TimingId:  "ASAP",
					Weight:    10.0},
				{
					ActionsId: "TOPUP_RST_5",
					TimingId:  "ASAP",
					Weight:    20.0},
			},
		},
	}
	if err := storDB.SetTPActionPlans(actionPlans); err != nil {
		t.Error(err)
	}

	if _, err := storDB.GetTPActionPlans("TPS1", "PCK_22"); err == nil || err != utils.ErrNotFound {
		t.Error(err)
	}
}
