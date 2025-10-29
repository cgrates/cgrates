/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func TestModelHelperCsvLoad(t *testing.T) {
	l, err := csvLoad(TpDestination{}, []string{"TEST_DEST", "+492"})
	tpd, ok := l.(TpDestination)
	if err != nil || !ok || tpd.Tag != "TEST_DEST" || tpd.Prefix != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestModelHelperCsvDump(t *testing.T) {
	tpd := TpDestination{
		Tag:    "TEST_DEST",
		Prefix: "+492"}
	csv, err := csvDump(tpd)
	if err != nil || csv[0] != "TEST_DEST" || csv[1] != "+492" {
		t.Errorf("model load failed: %+v", tpd)
	}
}

func TestTPDestinationAsExportSlice(t *testing.T) {
	tpDst := &utils.TPDestination{
		TPid:     "TEST_TPID",
		ID:       "TEST_DEST",
		Prefixes: []string{"49", "49176", "49151"},
	}
	expectedSlc := [][]string{
		{"TEST_DEST", "49"},
		{"TEST_DEST", "49176"},
		{"TEST_DEST", "49151"},
	}
	mdst := APItoModelDestination(tpDst)
	var slc [][]string
	for _, md := range mdst {
		lc, err := csvDump(md)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTpDestinationsAsTPDestinations(t *testing.T) {
	tpd1 := TpDestination{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+491"}
	tpd2 := TpDestination{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+492"}
	tpd3 := TpDestination{Tpid: "TEST_TPID", Tag: "TEST_DEST", Prefix: "+493"}
	eTPDestinations := []*utils.TPDestination{{TPid: "TEST_TPID", ID: "TEST_DEST",
		Prefixes: []string{"+491", "+492", "+493"}}}
	if tpDst := TpDestinations([]TpDestination{tpd1, tpd2, tpd3}).AsTPDestinations(); !reflect.DeepEqual(eTPDestinations, tpDst) {
		t.Errorf("Expecting: %+v, received: %+v", eTPDestinations, tpDst)
	}

}

func TestTPRateAsExportSlice(t *testing.T) {
	tpRate := &utils.TPRate{
		TPid: "TEST_TPID",
		ID:   "TEST_RATEID",
		RateSlots: []*utils.RateSlot{
			{
				ConnectFee:         0.100,
				Rate:               0.200,
				RateUnit:           "60",
				RateIncrement:      "60",
				GroupIntervalStart: "0"},
			{
				ConnectFee:         0.0,
				Rate:               0.1,
				RateUnit:           "1",
				RateIncrement:      "60",
				GroupIntervalStart: "60"},
		},
	}
	expectedSlc := [][]string{
		{"TEST_RATEID", "0.1", "0.2", "60", "60", "0"},
		{"TEST_RATEID", "0", "0.1", "1", "60", "60"},
	}

	ms := APItoModelRate(tpRate)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc[0], slc[0])
	}
}

func TestTPDestinationRateAsExportSlice(t *testing.T) {
	tpDstRate := &utils.TPDestinationRate{
		TPid: "TEST_TPID",
		ID:   "TEST_DSTRATE",
		DestinationRates: []*utils.DestinationRate{
			{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
			{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}
	expectedSlc := [][]string{
		{"TEST_DSTRATE", "TEST_DEST1", "TEST_RATE1", "*up", "4", "0", ""},
		{"TEST_DSTRATE", "TEST_DEST2", "TEST_RATE2", "*up", "4", "0", ""},
	}
	ms := APItoModelDestinationRate(tpDstRate)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}

}

func TestApierTPTimingAsExportSlice(t *testing.T) {
	tpTiming := &utils.ApierTPTiming{
		TPid:      "TEST_TPID",
		ID:        "TEST_TIMING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;4",
		Time:      "00:00:01"}
	expectedSlc := [][]string{
		{"TEST_TIMING", "*any", "*any", "*any", "1;2;4", "00:00:01"},
	}
	ms := APItoModelTiming(tpTiming)
	var slc [][]string

	lc, err := csvDump(ms)
	if err != nil {
		t.Error("Error dumping to csv: ", err)
	}
	slc = append(slc, lc)

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRatingPlanAsExportSlice(t *testing.T) {
	tpRpln := &utils.TPRatingPlan{
		TPid: "TEST_TPID",
		ID:   "TEST_RPLAN",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "TEST_DSTRATE1",
				TimingId:           "TEST_TIMING1",
				Weight:             10.0},
			{
				DestinationRatesId: "TEST_DSTRATE2",
				TimingId:           "TEST_TIMING2",
				Weight:             20.0},
		}}
	expectedSlc := [][]string{
		{"TEST_RPLAN", "TEST_DSTRATE1", "TEST_TIMING1", "10"},
		{"TEST_RPLAN", "TEST_DSTRATE2", "TEST_TIMING2", "20"},
	}

	ms := APItoModelRatingPlan(tpRpln)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRatingProfileAsExportSlice(t *testing.T) {
	tpRpf := &utils.TPRatingProfile{
		TPid:     "TEST_TPID",
		LoadId:   "TEST_LOADID",
		Tenant:   "cgrates.org",
		Category: "call",
		Subject:  "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
			{
				ActivationTime:   "2014-01-15T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN2",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	expectedSlc := [][]string{
		{"cgrates.org", "call", "*any", "2014-01-14T00:00:00Z", "TEST_RPLAN1", "subj1;subj2"},
		{"cgrates.org", "call", "*any", "2014-01-15T00:00:00Z", "TEST_RPLAN2", "subj1;subj2"},
	}

	ms := APItoModelRatingProfile(tpRpf)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionsAsExportSlice(t *testing.T) {
	tpActs := &utils.TPActions{
		TPid: "TEST_TPID",
		ID:   "TEST_ACTIONS",
		Actions: []*utils.TPAction{
			{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Units:           "5.0",
				ExpiryTime:      "*never",
				DestinationIds:  "*any",
				RatingSubject:   "special1",
				Categories:      "call",
				SharedGroups:    "GROUP1",
				BalanceWeight:   "10.0",
				ExtraParameters: "",
				Weight:          10.0},
			{
				Identifier:      "*http_post",
				BalanceType:     "",
				Units:           "0.0",
				ExpiryTime:      "",
				DestinationIds:  "",
				RatingSubject:   "",
				Categories:      "",
				SharedGroups:    "",
				BalanceWeight:   "0.0",
				ExtraParameters: "http://localhost/&param1=value1",
				Weight:          20.0},
		},
	}
	expectedSlc := [][]string{
		{"TEST_ACTIONS", "*topup_reset", "", "", "", "*monetary", "call", "*any", "special1", "GROUP1", "*never", "", "5.0", "10.0", "", "", "10"},
		{"TEST_ACTIONS", "*http_post", "http://localhost/&param1=value1", "", "", "", "", "", "", "", "", "", "0.0", "0.0", "", "", "20"},
	}

	ms := APItoModelAction(tpActs)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}

	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: \n%+v, received: \n%+v", expectedSlc, slc)
	}
}

// SHARED_A,*any,*highest,
func TestTPSharedGroupsAsExportSlice(t *testing.T) {
	tpSGs := &utils.TPSharedGroups{
		TPid: "TEST_TPID",
		ID:   "SHARED_GROUP_TEST",
		SharedGroups: []*utils.TPSharedGroup{
			{
				Account:       "*any",
				Strategy:      "*highest",
				RatingSubject: "special1"},
			{
				Account:       "second",
				Strategy:      "*highest",
				RatingSubject: "special2"},
		},
	}
	expectedSlc := [][]string{
		{"SHARED_GROUP_TEST", "*any", "*highest", "special1"},
		{"SHARED_GROUP_TEST", "second", "*highest", "special2"},
	}

	ms := APItoModelSharedGroup(tpSGs)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionTriggersAsExportSlice(t *testing.T) {
	ap := &utils.TPActionPlan{
		TPid: "TEST_TPID",
		ID:   "PACKAGE_10",
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
	}
	expectedSlc := [][]string{
		{"PACKAGE_10", "TOPUP_RST_10", "ASAP", "10"},
		{"PACKAGE_10", "TOPUP_RST_5", "ASAP", "20"},
	}
	ms := APItoModelActionPlan(ap)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionPlanAsExportSlice(t *testing.T) {
	at := &utils.TPActionTriggers{
		TPid: "TEST_TPID",
		ID:   "STANDARD_TRIGGERS",
		ActionTriggers: []*utils.TPActionTrigger{
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "1",
				ThresholdType:         "*min_balance",
				ThresholdValue:        2.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b1",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
			{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "2",
				ThresholdType:         "*max_event_counter",
				ThresholdValue:        5.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b2",
				BalanceType:           "*monetary",
				BalanceDestinationIds: "FS_USERS",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				ActionsId:             "LOG_WARNING",
				Weight:                10},
		},
	}
	expectedSlc := [][]string{
		{"STANDARD_TRIGGERS", "1", "*min_balance", "2", "false", "0", "", "", "b1", "*monetary", "call", "", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "LOG_WARNING", "10"},
		{"STANDARD_TRIGGERS", "2", "*max_event_counter", "5", "false", "0", "", "", "b2", "*monetary", "call", "FS_USERS", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "LOG_WARNING", "10"},
	}
	ms := APItoModelActionTrigger(at)
	var slc [][]string
	for _, m := range ms {
		lc, err := csvDump(m)
		if err != nil {
			t.Error("Error dumping to csv: ", err)
		}
		slc = append(slc, lc)
	}
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPAccountActionsAsExportSlice(t *testing.T) {
	aa := &utils.TPAccountActions{
		TPid:             "TEST_TPID",
		LoadId:           "TEST_LOADID",
		Tenant:           "cgrates.org",
		Account:          "1001",
		ActionPlanId:     "PACKAGE_10_SHARED_A_5",
		ActionTriggersId: "STANDARD_TRIGGERS",
	}
	expectedSlc := [][]string{
		{"cgrates.org", "1001", "PACKAGE_10_SHARED_A_5", "STANDARD_TRIGGERS", "false", "false"},
	}
	ms := APItoModelAccountAction(aa)
	var slc [][]string
	lc, err := csvDump(*ms)
	if err != nil {
		t.Error("Error dumping to csv: ", err)
	}
	slc = append(slc, lc)
	if !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTpResourcesAsTpResources(t *testing.T) {
	tps := []*TpResource{
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup1",
			FilterIDs:          "FLTR_RES_GR1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			Stored:             false,
			Blocker:            false,
			Weight:             10.0,
			Limit:              "45",
			ThresholdIDs:       "WARN_RES1;WARN_RES1"},
		{
			Tpid:         "TEST_TPID",
			ID:           "ResGroup1",
			Tenant:       "cgrates.org",
			FilterIDs:    "FLTR_RES_GR1",
			ThresholdIDs: "WARN3"},
		{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "ResGroup2",
			FilterIDs:          "FLTR_RES_GR2",
			ActivationInterval: "2014-07-29T15:00:00Z",
			Stored:             false,
			Blocker:            false,
			Weight:             10.0,
			Limit:              "20"},
	}
	eTPs := []*utils.TPResourceProfile{
		{
			TPid:      tps[0].Tpid,
			Tenant:    tps[0].Tenant,
			ID:        tps[0].ID,
			FilterIDs: []string{"FLTR_RES_GR1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			Stored:       tps[0].Stored,
			Blocker:      tps[0].Blocker,
			Weight:       tps[0].Weight,
			Limit:        tps[0].Limit,
			ThresholdIDs: []string{"WARN_RES1", "WARN3"},
		},
		{
			TPid:      tps[2].Tpid,
			Tenant:    tps[2].Tenant,
			ID:        tps[2].ID,
			FilterIDs: []string{"FLTR_RES_GR2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[2].ActivationInterval,
			},
			Stored:  tps[2].Stored,
			Blocker: tps[2].Blocker,
			Weight:  tps[2].Weight,
			Limit:   tps[2].Limit,
		},
	}
	rcvTPs := TpResources(tps).AsTPResources()
	if len(rcvTPs) != len(eTPs) {
		t.Errorf("Expecting: %+v Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:             "cgrates.org",
		TPid:               testTPID,
		ID:                 "ResGroup1",
		FilterIDs:          []string{"FLTR_RES_GR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		Stored:             false,
		Blocker:            false,
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       []string{"TRes1"},
		AllocationMessage:  "asd",
	}
	eRL := &ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                tpRL.ID,
		Stored:            tpRL.Stored,
		Blocker:           tpRL.Blocker,
		Weight:            tpRL.Weight,
		FilterIDs:         []string{"FLTR_RES_GR_1"},
		ThresholdIDs:      []string{"TRes1"},
		AllocationMessage: tpRL.AllocationMessage,
		Limit:             2,
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eRL.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}
	if rl, err := APItoResource(tpRL, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRL, rl) {
		t.Errorf("Expecting: %+v, received: %+v", eRL, rl)
	}
}

func TestAPItoModelResource(t *testing.T) {
	tpRL := &utils.TPResourceProfile{
		Tenant:             "cgrates.org",
		TPid:               testTPID,
		ID:                 "ResGroup1",
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		Weight:             10,
		Limit:              "2",
		ThresholdIDs:       []string{"TRes1"},
		AllocationMessage:  "test",
	}
	expModel := &TpResource{
		Tpid:               testTPID,
		Tenant:             "cgrates.org",
		ID:                 "ResGroup1",
		ActivationInterval: "2014-07-29T15:00:00Z",
		Weight:             10.0,
		Limit:              "2",
		ThresholdIDs:       "TRes1",
		AllocationMessage:  "test",
	}
	rcv := APItoModelResource(tpRL)
	if len(rcv) != 1 {
		t.Errorf("Expecting: 1, received: %+v", len(rcv))
	} else if !reflect.DeepEqual(rcv[0], expModel) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expModel), utils.ToJSON(rcv[0]))
	}
}

func TestTPStatsAsTPStats(t *testing.T) {
	tps := TpStats{
		&TpStat{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*asr;*acc;*tcc;*acd;*tcd;*pdd",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&TpStat{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs:       "THRESH3",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&TpStat{
			Tpid:               "TEST_TPID",
			Tenant:             "itsyscom.com",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*sum#BalanceValue;*average#BalanceValue;*tcc",
			ThresholdIDs:       "THRESH4",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
	}
	rcvTPs := tps.AsTPStats()
	if len(rcvTPs) != 2 {
		t.Errorf("Expecting: 2, received: %+v", len(rcvTPs))
	}
	for _, rcvTP := range rcvTPs {
		if rcvTP.Tenant == "cgrates.org" {
			if len(rcvTP.Metrics) != 8 {
				t.Errorf("Expecting: 8, received: %+v", len(rcvTP.Metrics))
			}
		} else {
			if len(rcvTP.Metrics) != 3 {
				t.Errorf("Expecting: 3, received: %+v", len(rcvTP.Metrics))
			}
		}
	}
}

func TestAPItoTPStats(t *testing.T) {
	tps := &utils.TPStatProfile{
		TPid:               testTPID,
		ID:                 "Stats1",
		FilterIDs:          []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		QueueLength:        100,
		TTL:                "1s",
		Metrics: []*utils.MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blocker:      false,
		Weight:       20.0,
	}
	eTPs := &StatQueueProfile{ID: tps.ID,
		QueueLength: tps.QueueLength,
		Metrics: []*MetricWithFilters{
			{
				MetricID: "*sum#BalanceValue",
			},
			{
				MetricID: "*average#BalanceValue",
			},
			{
				MetricID: "*tcc",
			},
		},
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		FilterIDs:    []string{"FLTR_1"},
		Stored:       tps.Stored,
		Blocker:      tps.Blocker,
		Weight:       20.0,
		MinItems:     tps.MinItems,
	}
	if eTPs.TTL, err = utils.ParseDurationWithNanosecs(tps.TTL); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eTPs.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}

	if st, err := APItoStats(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestAPItoModelStats(t *testing.T) {
	tpS := &utils.TPStatProfile{
		TPid:      "TPS1",
		Tenant:    "cgrates.org",
		ID:        "Stat1",
		FilterIDs: []string{"*string:Account:1002"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
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
	}
	rcv := APItoModelStats(tpS)
	eRcv := TpStats{
		&TpStat{
			Tpid:               "TPS1",
			Tenant:             "cgrates.org",
			ID:                 "Stat1",
			FilterIDs:          "*string:Account:1002",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			MetricIDs:          "*tcc",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
			ThresholdIDs:       "Th1",
		},
		&TpStat{
			Tpid:      "TPS1",
			Tenant:    "cgrates.org",
			ID:        "Stat1",
			MetricIDs: "*average#Usage",
		},
	}
	if len(rcv) != len(eRcv) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv), len(rcv))
	} else if !reflect.DeepEqual(eRcv, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(eRcv), utils.ToJSON(rcv))
	}
}

func TestTPThresholdsAsTPThreshold(t *testing.T) {
	tps := []*TpThreshold{
		{
			Tpid:               "TEST_TPID",
			ID:                 "Threhold",
			FilterIDs:          "FilterID1;FilterID2;FilterID1;FilterID2;FilterID2",
			ActivationInterval: "2014-07-29T15:00:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
	}
	eTPs := []*utils.TPThresholdProfile{
		{
			TPid:      tps[0].Tpid,
			ID:        tps[0].ID,
			FilterIDs: []string{"FilterID1", "FilterID2"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			MinSleep:  tps[0].MinSleep,
			MaxHits:   tps[0].MaxHits,
			MinHits:   tps[0].MinHits,
			Blocker:   tps[0].Blocker,
			Weight:    tps[0].Weight,
			ActionIDs: []string{"WARN3"},
		},
		{
			TPid:      tps[0].Tpid,
			ID:        tps[0].ID,
			FilterIDs: []string{"FilterID2", "FilterID1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			MinSleep:  tps[0].MinSleep,
			MaxHits:   tps[0].MaxHits,
			MinHits:   tps[0].MinHits,
			Blocker:   tps[0].Blocker,
			Weight:    tps[0].Weight,
			ActionIDs: []string{"WARN3"},
		},
	}
	rcvTPs := TpThresholds(tps).AsTPThreshold()
	if !reflect.DeepEqual(eTPs[0], rcvTPs[0]) && !reflect.DeepEqual(eTPs[1], rcvTPs[0]) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoModelTPThreshold(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
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
	}
	models := TpThresholds{
		&TpThreshold{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FilterID1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold2(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1", "FLTR_2"},
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
	}
	models := TpThresholds{
		&TpThreshold{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		&TpThreshold{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			FilterIDs: "FLTR_2",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold3(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3", "LOG"},
	}
	models := TpThresholds{
		&TpThreshold{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		&TpThreshold{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			ActionIDs: "LOG",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold4(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"WARN3", "LOG"},
	}
	models := TpThresholds{
		&TpThreshold{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "TH_1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			MaxHits:            12,
			MinHits:            10,
			MinSleep:           "1s",
			Blocker:            false,
			Weight:             20.0,
			ActionIDs:          "WARN3",
		},
		&TpThreshold{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "TH_1",
			ActionIDs: "LOG",
		},
	}
	rcv := APItoModelTPThreshold(th)
	if !reflect.DeepEqual(models, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(models), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPThreshold5(t *testing.T) {
	th := &utils.TPThresholdProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "TH_1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		MaxHits:   12,
		MinHits:   10,
		MinSleep:  "1s",
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{},
	}
	rcv := APItoModelTPThreshold(th)
	if rcv != nil {
		t.Errorf("Expecting : nil, received: %+v", utils.ToJSON(rcv))
	}
}

func TestAPItoTPThreshold(t *testing.T) {
	tps := &utils.TPThresholdProfile{
		TPid:               testTPID,
		ID:                 "TH1",
		FilterIDs:          []string{"FilterID1", "FilterID2"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		MaxHits:            12,
		MinHits:            10,
		MinSleep:           "1s",
		Blocker:            false,
		Weight:             20.0,
		ActionIDs:          []string{"WARN3"},
	}

	eTPs := &ThresholdProfile{
		ID:        tps.ID,
		MaxHits:   tps.MaxHits,
		Blocker:   tps.Blocker,
		MinHits:   tps.MinHits,
		Weight:    tps.Weight,
		FilterIDs: tps.FilterIDs,
		ActionIDs: []string{"WARN3"},
	}
	if eTPs.MinSleep, err = utils.ParseDurationWithNanosecs(tps.MinSleep); err != nil {
		t.Errorf("Got error: %+v", err)
	}
	at, _ := utils.ParseTimeDetectLayout("2014-07-29T15:00:00Z", "UTC")
	eTPs.ActivationInterval = &utils.ActivationInterval{ActivationTime: at}
	if st, err := APItoThresholdProfile(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestTPFilterAsTPFilter(t *testing.T) {
	tps := []*TpFilter{
		{
			Tpid:    "TEST_TPID",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}

	rcvTPs := TpFilterS(tps).AsTPFilter()
	if !(reflect.DeepEqual(eTPs, rcvTPs) || reflect.DeepEqual(eTPs[0], rcvTPs[0])) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestTPFilterAsTPFilter2(t *testing.T) {
	tps := []*TpFilter{
		{
			Tpid:    "TEST_TPID",
			Tenant:  "cgrates.org",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1001;1002",
		},
		{
			Tpid:    "TEST_TPID",
			Tenant:  "anotherTenant",
			ID:      "Filter1",
			Type:    utils.MetaPrefix,
			Element: "Account",
			Values:  "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
		{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				{
					Type:    utils.MetaPrefix,
					Element: "Account",
					Values:  []string{"1010"},
				},
			},
		},
	}

	rcvTPs := TpFilterS(tps).AsTPFilter()
	if len(eTPs) != len(rcvTPs) {
		t.Errorf("Expecting: %+v ,Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoTPFilter(t *testing.T) {
	tps := &utils.TPFilterProfile{
		TPid:   testTPID,
		Tenant: "cgrates.org",
		ID:     "Filter1",
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}

	eTPs := &Filter{
		Tenant: "cgrates.org",
		ID:     tps.ID,
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	if st, err := APItoFilter(tps, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTPs, st) {
		t.Errorf("Expecting: %+v, received: %+v", eTPs, st)
	}
}

func TestFilterToTPFilter(t *testing.T) {
	filter := &Filter{
		Tenant: "cgrates.org",
		ID:     "Fltr1",
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 1, 14, 0, 0, 0, 0, time.UTC),
		},
		Rules: []*FilterRule{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	tpfilter := &utils.TPFilterProfile{
		ID:     "Fltr1",
		Tenant: "cgrates.org",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-01-14T00:00:00Z",
			ExpiryTime:     "2014-01-14T00:00:00Z",
		},
		Filters: []*utils.TPFilter{
			{
				Element: "Account",
				Type:    utils.MetaString,
				Values:  []string{"1001", "1002"},
			},
		},
	}
	eTPFilter := FilterToTPFilter(filter)
	if !reflect.DeepEqual(tpfilter, eTPFilter) {
		t.Errorf("Expecting: %+v, received: %+v", tpfilter, eTPFilter)
	}
}

func TestAPItoAttributeProfile(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}
	expected := &AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	if rcv, err := APItoAttributeProfile(tpAlsPrf, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPAttribute(t *testing.T) {
	tpAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + "FL1",
				Value: "Al1",
			},
		},
		Weight: 20,
	}
	expected := TPAttributes{
		&TPAttribute{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPAttribute(tpAlsPrf)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestModelAsTPAttribute(t *testing.T) {
	models := TPAttributes{
		&TPAttribute{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "ALS1",
			Contexts:           "con1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Path:               utils.MetaReq + utils.NestingSep + "FL1",
			Value:              "Al1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	expected2 := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		Contexts:  []string{"con1"},
		FilterIDs: []string{"FLTR_DST_DE", "FLTR_ACNT_dan"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Attributes: []*utils.TPAttribute{
			{
				FilterIDs: []string{},
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
			},
		},
		Weight: 20,
	}
	rcv := models.AsTPAttributes()
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoChargerProfile(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}

	expected := &ChargerProfile{
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		RunID:        "*rated",
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	if rcv, err := APItoChargerProfile(tpCPP, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs and AttributeIDs are equal
func TestAPItoModelTPCharger(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_DST_DE",
			AttributeIDs:       "ATTR2",
			ActivationInterval: "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is smaller than AttributeIDs
func TestAPItoModelTPCharger2(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1", "ATTR2"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			AttributeIDs:       "ATTR2",
			ActivationInterval: "",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// Number of FilterIDs is greater than AttributeIDs
func TestAPItoModelTPCharger3(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
		&TPCharger{
			Tpid:      "TP1",
			Tenant:    "cgrates.org",
			ID:        "Charger1",
			FilterIDs: "FLTR_DST_DE",
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(AttributeIDs) is 0
func TestAPItoModelTPCharger4(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Weight: 20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan",
			RunID:              "*rated",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// len(FilterIDs) is 0
func TestAPItoModelTPCharger5(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "Charger1",
		RunID:  "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

// both len(AttributeIDs) and len(FilterIDs) are 0
func TestAPItoModelTPCharger6(t *testing.T) {
	tpCharger := &utils.TPChargerProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "Charger1",
		RunID:  "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		Weight: 20,
	}
	expected := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			RunID:              "*rated",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	rcv := APItoModelTPCharger(tpCharger)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestModelAsTPChargers(t *testing.T) {
	models := TPChargers{
		&TPCharger{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Charger1",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			RunID:              "*rated",
			AttributeIDs:       "ATTR1",
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
		},
	}
	expected := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	expected2 := &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Charger1",
		FilterIDs: []string{"FLTR_DST_DE", "FLTR_ACNT_dan"},
		RunID:     "*rated",
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		AttributeIDs: []string{"ATTR1"},
		Weight:       20,
	}
	rcv := models.AsTPChargers()
	if !reflect.DeepEqual(expected, rcv[0]) && !reflect.DeepEqual(expected2, rcv[0]) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv[0]))
	}
}

func TestAPItoDispatcherProfile(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:       "TP1",
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203", "*ratio:2"},
				Blocker:   false,
			},
		},
	}

	expected := &DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		StrategyParams: map[string]any{},
		Weight:         20,
		Hosts: DispatcherHostProfiles{
			&DispatcherHostProfile{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    map[string]any{"0": "192.168.54.203", utils.MetaRatio: "2"},
				Blocker:   false,
			},
		},
	}
	if rcv, err := APItoDispatcherProfile(tpDPP, "UTC"); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestAPItoModelTPDispatcher(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:       "TP1",
		Tenant:     "cgrates.org",
		ID:         "Dsp",
		Subsystems: []string{"*any"},
		FilterIDs:  []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		Strategy:   utils.MetaFirst,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-14T14:35:00Z",
			ExpiryTime:     "",
		},
		StrategyParams: []any{},
		Weight:         20,
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				ID:        "C1",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.203"},
				Blocker:   false,
			},
			{
				ID:        "C2",
				FilterIDs: []string{},
				Weight:    10,
				Params:    []any{"192.168.54.204"},
				Blocker:   false,
			},
		},
	}
	expected := TPDispatcherProfiles{
		&TPDispatcherProfile{
			Tpid:               "TP1",
			Tenant:             "cgrates.org",
			ID:                 "Dsp",
			Subsystems:         "*any",
			FilterIDs:          "FLTR_ACNT_dan;FLTR_DST_DE",
			Strategy:           utils.MetaFirst,
			ActivationInterval: "2014-07-14T14:35:00Z",
			Weight:             20,
			ConnID:             "C1",
			ConnWeight:         10,
			ConnBlocker:        false,
			ConnParameters:     "192.168.54.203",
		},
		&TPDispatcherProfile{
			Tpid:           "TP1",
			Tenant:         "cgrates.org",
			ID:             "Dsp",
			ConnID:         "C2",
			ConnWeight:     10,
			ConnBlocker:    false,
			ConnParameters: "192.168.54.204",
		},
	}
	rcv := APItoModelTPDispatcherProfile(tpDPP)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(expected), utils.ToJSON(rcv))
	}
}

func TestTPSuppliersAsTPSupplierProfiles(t *testing.T) {
	mdl := TpSuppliers{
		&TpSupplier{
			PK:                    1,
			Tpid:                  "TP",
			Tenant:                "cgrates.org",
			ID:                    "SupplierPrf",
			FilterIDs:             "FltrSupplier",
			ActivationInterval:    "2017-11-27T00:00:00Z",
			Sorting:               "*weight",
			SortingParameters:     "srtPrm1;srtPrm2",
			SupplierID:            "supplier1",
			SupplierFilterIDs:     "",
			SupplierAccountIDs:    "",
			SupplierRatingplanIDs: "",
			SupplierResourceIDs:   "",
			SupplierStatIDs:       "",
			SupplierWeight:        10.0,
			SupplierBlocker:       false,
			SupplierParameters:    "",
			Weight:                10.0,
			CreatedAt:             time.Time{},
		},
		&TpSupplier{
			PK:                    2,
			Tpid:                  "TP",
			Tenant:                "cgrates.org",
			ID:                    "SupplierPrf",
			FilterIDs:             "",
			ActivationInterval:    "",
			Sorting:               "",
			SortingParameters:     "",
			SupplierID:            "supplier2",
			SupplierFilterIDs:     "",
			SupplierAccountIDs:    "",
			SupplierRatingplanIDs: "",
			SupplierResourceIDs:   "",
			SupplierStatIDs:       "",
			SupplierWeight:        20.0,
			SupplierBlocker:       false,
			SupplierParameters:    "",
			Weight:                0,
			CreatedAt:             time.Time{},
		},
	}
	expPrf := []*utils.TPSupplierProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "SupplierPrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1", "srtPrm2"},
			FilterIDs:         []string{"FltrSupplier"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "",
			},
			Suppliers: []*utils.TPSupplier{
				{
					ID:     "supplier1",
					Weight: 10.0,
				},
				{
					ID:     "supplier2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcv := mdl.AsTPSuppliers()
	sort.Slice(rcv[0].Suppliers, func(i, j int) bool {
		return strings.Compare(rcv[0].Suppliers[i].ID, rcv[0].Suppliers[j].ID) < 0
	})
	sort.Strings(rcv[0].SortingParameters)
	if !reflect.DeepEqual(rcv, expPrf) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrf), utils.ToJSON(rcv))
	}

	mdlReverse := TpSuppliers{
		&TpSupplier{
			PK:                    2,
			Tpid:                  "TP",
			Tenant:                "cgrates.org",
			ID:                    "SupplierPrf",
			FilterIDs:             "",
			ActivationInterval:    "",
			Sorting:               "",
			SortingParameters:     "",
			SupplierID:            "supplier2",
			SupplierFilterIDs:     "",
			SupplierAccountIDs:    "",
			SupplierRatingplanIDs: "",
			SupplierResourceIDs:   "",
			SupplierStatIDs:       "",
			SupplierWeight:        20.0,
			SupplierBlocker:       false,
			SupplierParameters:    "",
			Weight:                0,
			CreatedAt:             time.Time{},
		},
		&TpSupplier{
			PK:                    1,
			Tpid:                  "TP",
			Tenant:                "cgrates.org",
			ID:                    "SupplierPrf",
			FilterIDs:             "FltrSupplier",
			ActivationInterval:    "2017-11-27T00:00:00Z",
			Sorting:               "*weight",
			SortingParameters:     "srtPrm1;srtPrm2",
			SupplierID:            "supplier1",
			SupplierFilterIDs:     "",
			SupplierAccountIDs:    "",
			SupplierRatingplanIDs: "",
			SupplierResourceIDs:   "",
			SupplierStatIDs:       "",
			SupplierWeight:        10.0,
			SupplierBlocker:       false,
			SupplierParameters:    "",
			Weight:                10.0,
			CreatedAt:             time.Time{},
		},
	}
	expPrfRev := []*utils.TPSupplierProfile{
		{
			TPid:              "TP",
			Tenant:            "cgrates.org",
			ID:                "SupplierPrf",
			Sorting:           "*weight",
			SortingParameters: []string{"srtPrm1", "srtPrm2"},
			FilterIDs:         []string{"FltrSupplier"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: "2017-11-27T00:00:00Z",
				ExpiryTime:     "",
			},
			Suppliers: []*utils.TPSupplier{
				{
					ID:     "supplier1",
					Weight: 10.0,
				},
				{
					ID:     "supplier2",
					Weight: 20.0,
				},
			},
			Weight: 10,
		},
	}
	rcvRev := mdlReverse.AsTPSuppliers()
	sort.Slice(rcvRev[0].Suppliers, func(i, j int) bool {
		return strings.Compare(rcvRev[0].Suppliers[i].ID, rcvRev[0].Suppliers[j].ID) < 0
	})
	sort.Strings(rcvRev[0].SortingParameters)
	if !reflect.DeepEqual(rcvRev, expPrfRev) {
		t.Errorf("Expecting: %+v,\nReceived: %+v", utils.ToJSON(expPrfRev), utils.ToJSON(rcvRev))
	}
}

func TestModelHelpersAsMapDestinations(t *testing.T) {
	tp := TpDestination{
		Id:        1,
		Tpid:      "test",
		Tag:       "test",
		Prefix:    "test",
		CreatedAt: time.Now(),
	}
	tps := TpDestinations{tp}

	d := &Destination{
		Id:       "test",
		Prefixes: []string{"test"},
	}
	exp := map[string]*Destination{"test": d}

	rcv, err := tps.AsMapDestinations()
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelDestination(t *testing.T) {
	d := &utils.TPDestination{
		TPid: "test",
		ID:   "test",
	}

	tp := TpDestination{
		Tpid: "test",
		Tag:  "test",
	}
	exp := TpDestinations{tp}

	rcv := APItoModelDestination(d)

	if !reflect.DeepEqual(exp, rcv) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelTimings(t *testing.T) {
	ts := []*utils.ApierTPTiming{{
		TPid:      str,
		ID:        str,
		Years:     str,
		Months:    str,
		MonthDays: str,
		WeekDays:  str,
		Time:      str,
	}}

	tp := TpTiming{
		Id:        0,
		Tpid:      str,
		Tag:       str,
		Years:     str,
		Months:    str,
		MonthDays: str,
		WeekDays:  str,
		Time:      str,
	}
	exp := TpTimings{tp}

	rcv := APItoModelTimings(ts)

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelRate(t *testing.T) {
	r := &utils.TPRate{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelRate(r)
	exp := TpRates{{
		Tpid: r.TPid,
		Tag:  r.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelRates(t *testing.T) {
	rs := []*utils.TPRate{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelRates(rs)
	exp := APItoModelRate(rs[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersMapTPDestinationRates(t *testing.T) {
	s := []*utils.TPDestinationRate{{
		TPid: str,
		ID:   str,
	}, {
		TPid: str,
		ID:   str,
	}}

	_, err := MapTPDestinationRates(s)

	if err != nil {
		if err.Error() != "Non unique ID test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelDestinationRate(t *testing.T) {
	d := &utils.TPDestinationRate{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelDestinationRate(d)
	exp := TpDestinationRates{{
		Tpid: d.TPid,
		Tag:  d.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelDestinationRates(t *testing.T) {
	drs := []*utils.TPDestinationRate{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelDestinationRates(drs)
	exp := APItoModelDestinationRate(drs[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelRatingPlan(t *testing.T) {
	rp := &utils.TPRatingPlan{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelRatingPlan(rp)
	exp := TpRatingPlans{{
		Tpid: rp.TPid,
		Tag:  rp.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelRatingPlans(t *testing.T) {
	rps := []*utils.TPRatingPlan{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelRatingPlans(rps)
	exp := APItoModelRatingPlan(rps[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersMapTPRatingProfiles(t *testing.T) {
	s := []*utils.TPRatingProfile{{
		TPid:   str,
		LoadId: str,
	}, {
		TPid:   str,
		LoadId: str,
	}}

	_, err := MapTPRatingProfiles(s)

	if err != nil {
		if err.Error() != "Non unique id test:*out:::" {
			t.Error(err)
		}
	}
}

func TestModelHelpersMapTPRates(t *testing.T) {
	s := []*utils.TPRate{{
		TPid: str,
		ID:   str,
	}, {
		TPid: str,
		ID:   str,
	}}

	_, err := MapTPRates(s)

	if err != nil {
		if err.Error() != "Non unique ID test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersMapTPTimings(t *testing.T) {
	tps := []*utils.ApierTPTiming{{
		TPid: str,
		ID:   str,
	}, {
		TPid: str,
		ID:   str,
	}}

	_, err := MapTPTimings(tps)

	if err != nil {
		if err.Error() != "duplicate timing tag: test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelRatingProfile(t *testing.T) {
	rp := &utils.TPRatingProfile{
		TPid:     str,
		LoadId:   str,
		Tenant:   str,
		Category: str,
		Subject:  str,
	}

	rcv := APItoModelRatingProfile(rp)
	exp := TpRatingProfiles{{
		Tpid:     rp.TPid,
		Loadid:   rp.LoadId,
		Tenant:   rp.Tenant,
		Category: rp.Category,
		Subject:  rp.Subject,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelRatingProfiles(t *testing.T) {
	rps := []*utils.TPRatingProfile{{
		TPid:     str,
		LoadId:   str,
		Tenant:   str,
		Category: str,
		Subject:  str,
	}}

	rcv := APItoModelRatingProfiles(rps)
	exp := APItoModelRatingProfile(rps[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersMapTPSharedGroup(t *testing.T) {
	s := []*utils.TPSharedGroups{{
		TPid: str,
		ID:   str,
		SharedGroups: []*utils.TPSharedGroup{{
			Account:       str,
			Strategy:      str,
			RatingSubject: str,
		}},
	},
		{
			TPid: str,
			ID:   str,
			SharedGroups: []*utils.TPSharedGroup{{
				Account:       str,
				Strategy:      str,
				RatingSubject: str,
			}},
		}}

	rcv := MapTPSharedGroup(s)
	exp := map[string][]*utils.TPSharedGroup{str: {{
		Account:       str,
		Strategy:      str,
		RatingSubject: str,
	}, {
		Account:       str,
		Strategy:      str,
		RatingSubject: str,
	}}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelSharedGroup(t *testing.T) {
	sgs := &utils.TPSharedGroups{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelSharedGroup(sgs)
	exp := TpSharedGroups{{
		Tpid: sgs.TPid,
		Tag:  sgs.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelSharedGroups(t *testing.T) {
	sgs := []*utils.TPSharedGroups{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelSharedGroups(sgs)
	exp := APItoModelSharedGroup(sgs[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelAction(t *testing.T) {
	as := &utils.TPActions{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelAction(as)
	exp := TpActions{{
		Tpid: as.TPid,
		Tag:  as.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelActions(t *testing.T) {
	as := []*utils.TPActions{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelActions(as)
	exp := APItoModelAction(as[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelActionPlan(t *testing.T) {
	a := &utils.TPActionPlan{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelActionPlan(a)
	exp := TpActionPlans{{
		Tpid: a.TPid,
		Tag:  a.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelActionPlans(t *testing.T) {
	aps := []*utils.TPActionPlan{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelActionPlans(aps)
	exp := APItoModelActionPlan(aps[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelActionTrigger(t *testing.T) {
	ats := &utils.TPActionTriggers{
		TPid: str,
		ID:   str,
	}

	rcv := APItoModelActionTrigger(ats)
	exp := TpActionTriggers{{
		Tpid: ats.TPid,
		Tag:  ats.ID,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelActionTriggers(t *testing.T) {
	ts := []*utils.TPActionTriggers{{
		TPid: str,
		ID:   str,
	}}

	rcv := APItoModelActionTriggers(ts)
	exp := APItoModelActionTrigger(ts[0])

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersMapTPAccountActions(t *testing.T) {
	s := []*utils.TPAccountActions{{
		TPid:   str,
		LoadId: str,
	}, {
		TPid:   str,
		LoadId: str,
	}}

	_, err := MapTPAccountActions(s)

	if err != nil {
		if err.Error() != "Non unique ID :" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelAccountActions(t *testing.T) {
	aas := []*utils.TPAccountActions{{
		TPid:   str,
		LoadId: str,
	}}

	rcv := APItoModelAccountActions(aas)
	exp := TpAccountActions{*APItoModelAccountAction(aas[0])}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", exp, rcv)
	}
}

func TestModelHelpersAPItoModelResourceNil(t *testing.T) {
	var rl *utils.TPResourceProfile

	rcv := APItoModelResource(rl)

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestModelHelpersAPItoModelTPFilter(t *testing.T) {
	th := &utils.TPFilterProfile{
		TPid:   str,
		Tenant: str,
		ID:     str,
		Filters: []*utils.TPFilter{{
			Type:    str,
			Element: str,
			Values:  []string{"val1", "val2"},
		}},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	rcv := APItoModelTPFilter(th)
	exp := TpFilterS{{
		Tpid:               str,
		Tenant:             str,
		ID:                 str,
		Type:               str,
		Element:            str,
		Values:             "val1;val2",
		ActivationInterval: "test;test",
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	th2 := &utils.TPFilterProfile{}
	rcv = APItoModelTPFilter(th2)

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestModelHelpersAsMapTPSharedGroups(t *testing.T) {
	tps := TpSharedGroups{{
		Id:   1,
		Tpid: str,
	},
		{
			Id:   1,
			Tpid: str,
		}}

	rcv, err := tps.AsMapTPSharedGroups()
	if err != nil {
		t.Error(err)
	}
	exp := map[string]*utils.TPSharedGroups{"": {
		TPid:         str,
		ID:           "",
		SharedGroups: []*utils.TPSharedGroup{{}, {}},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoModelResource(t *testing.T) {
	slc := []string{"val1", "val2"}
	rl := &utils.TPResourceProfile{
		TPid:      str,
		Tenant:    str,
		ID:        str,
		FilterIDs: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		UsageTTL:          str,
		Limit:             str,
		AllocationMessage: str,
		Blocker:           bl,
		Stored:            bl,
		Weight:            fl,
		ThresholdIDs:      slc,
	}

	rcv := APItoModelResource(rl)
	exp := TpResources{{
		Tpid:               str,
		Tenant:             str,
		ID:                 str,
		FilterIDs:          "val1",
		ActivationInterval: "test;test",
		UsageTTL:           str,
		Limit:              str,
		AllocationMessage:  str,
		Blocker:            bl,
		Stored:             bl,
		Weight:             fl,
		ThresholdIDs:       "val1;val2",
	}, {
		FilterIDs: "val2",
		Blocker:   bl,
		ID:        str,
		Stored:    bl,
		Tenant:    str,
		Tpid:      str,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoModelStats(t *testing.T) {
	slc := []string{"val1", "val2"}
	st := &utils.TPStatProfile{
		TPid:      str,
		Tenant:    str,
		ID:        str,
		FilterIDs: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		QueueLength: nm,
		TTL:         str,
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: slc,
				MetricID:  str,
			}, {
				FilterIDs: slc,
				MetricID:  str,
			},
		},
		Blocker:      bl,
		Stored:       bl,
		Weight:       fl,
		MinItems:     nm,
		ThresholdIDs: slc,
	}

	rcv := APItoModelStats(st)
	exp := TpStats{{
		Tpid:               str,
		Tenant:             str,
		ID:                 str,
		FilterIDs:          "val1;val2",
		ActivationInterval: "test;test",
		QueueLength:        nm,
		TTL:                str,
		MinItems:           nm,
		MetricIDs:          str,
		MetricFilterIDs:    "val1;val2",
		Stored:             bl,
		Blocker:            bl,
		Weight:             fl,
		ThresholdIDs:       "val1;val2",
	}, {
		ID:              str,
		MetricFilterIDs: "val1;val2",
		MetricIDs:       str,
		Tenant:          str,
		Tpid:            str,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoModelTPSuppliers(t *testing.T) {
	slc := []string{"val1", "val2"}
	st := &utils.TPSupplierProfile{
		TPid:      str,
		Tenant:    str,
		ID:        str,
		FilterIDs: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		Sorting:           str,
		SortingParameters: slc,
		Suppliers: []*utils.TPSupplier{
			{
				ID:                 str,
				FilterIDs:          slc,
				AccountIDs:         slc,
				RatingPlanIDs:      slc,
				ResourceIDs:        slc,
				StatIDs:            slc,
				Weight:             fl,
				Blocker:            bl,
				SupplierParameters: str,
			},
		},
		Weight: fl,
	}

	rcv := APItoModelTPSuppliers(st)
	exp := TpSuppliers{{
		Tpid:                  str,
		Tenant:                str,
		ID:                    str,
		FilterIDs:             "val1;val2",
		ActivationInterval:    "test;test",
		Sorting:               str,
		SortingParameters:     "val1;val2",
		SupplierID:            str,
		SupplierFilterIDs:     "val1;val2",
		SupplierAccountIDs:    "val1;val2",
		SupplierRatingplanIDs: "val1;val2",
		SupplierResourceIDs:   "val1;val2",
		SupplierStatIDs:       "val1;val2",
		SupplierWeight:        fl,
		SupplierBlocker:       bl,
		SupplierParameters:    str,
		Weight:                fl,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	st2 := &utils.TPSupplierProfile{
		Suppliers: []*utils.TPSupplier{},
	}
	rcv = APItoModelTPSuppliers(st2)

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestModelHepersAPItoModelTPDispatcherHost(t *testing.T) {
	rcv := APItoModelTPDispatcherHost(nil)

	if rcv != nil {
		t.Error(err)
	}

	tpDPH := &utils.TPDispatcherHost{
		TPid:   str,
		Tenant: str,
		ID:     str,
		Conns: []*utils.TPDispatcherHostConn{{
			Address:   str,
			Transport: str,
			TLS:       bl,
		}},
	}

	rcv = APItoModelTPDispatcherHost(tpDPH)
	exp := TPDispatcherHosts{{
		Tpid:      tpDPH.TPid,
		Tenant:    tpDPH.Tenant,
		ID:        tpDPH.ID,
		Address:   tpDPH.Conns[0].Address,
		Transport: tpDPH.Conns[0].Transport,
		TLS:       tpDPH.Conns[0].TLS,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoDispatcherHost(t *testing.T) {
	rcv := APItoDispatcherHost(nil)

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestModelHelpersAPItoModelTPAttribute(t *testing.T) {
	slc := []string{"val1", "val2"}
	th := &utils.TPAttributeProfile{
		Attributes: []*utils.TPAttribute{},
	}
	rcv := APItoModelTPAttribute(th)

	if rcv != nil {
		t.Error(rcv)
	}

	th2 := &utils.TPAttributeProfile{
		TPid:      str,
		Tenant:    str,
		ID:        str,
		FilterIDs: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		Contexts: slc,
		Attributes: []*utils.TPAttribute{{
			FilterIDs: slc,
			Path:      str,
			Type:      str,
			Value:     str,
		}},
		Blocker: bl,
		Weight:  fl,
	}

	rcv = APItoModelTPAttribute(th2)
	exp := TPAttributes{{
		Tpid:               str,
		Tenant:             str,
		ID:                 str,
		Contexts:           "val1;val2",
		FilterIDs:          "val1;val2",
		ActivationInterval: "test;test",
		AttributeFilterIDs: "val1;val2",
		Path:               str,
		Type:               str,
		Value:              str,
		Blocker:            bl,
		Weight:             fl,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersmodelEqual(t *testing.T) {
	type Test struct {
		Fl  float64 `index:"0"`
		Nm  int     `index:"1"`
		Bl  bool    `index:"2"`
		Str string  `index:"3"`
	}
	tst := Test{
		Fl:  fl,
		Nm:  nm,
		Bl:  bl,
		Str: str,
	}
	tst2 := Test{
		Fl: 2.5,
	}
	tst3 := Test{
		Fl: fl,
		Nm: 5,
	}
	tst4 := Test{
		Fl: fl,
		Nm: nm,
		Bl: false,
	}
	tst5 := Test{
		Fl:  fl,
		Nm:  nm,
		Bl:  bl,
		Str: "test1",
	}
	type args struct {
		this  any
		other any
	}
	tests := []struct {
		name string
		args args
		exp  bool
	}{
		{
			name: "true return",
			args: args{tst, tst},
			exp:  true,
		},
		{
			name: "false return",
			args: args{nm, str},
			exp:  false,
		},
		{
			name: "false return",
			args: args{tst, tst2},
			exp:  false,
		},
		{
			name: "false return",
			args: args{tst, tst3},
			exp:  false,
		},
		{
			name: "false return",
			args: args{tst, tst4},
			exp:  false,
		},
		{
			name: "false return",
			args: args{tst, tst5},
			exp:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rcv := modelEqual(tt.args.this, tt.args.other)

			if rcv != tt.exp {
				t.Errorf("received %v, expected %v", rcv, tt.exp)
			}
		})
	}
}

func TestModelHelpersparamsToString(t *testing.T) {
	rcv := paramsToString([]any{str, str})

	if rcv != "test;test" {
		t.Error(rcv)
	}
}

func TestModelHelpersAPItoModelTPDispatcherProfile(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		TPid:       str,
		Tenant:     str,
		ID:         str,
		Subsystems: slc,
		FilterIDs:  slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		Strategy:       str,
		StrategyParams: []any{str, str},
		Weight:         fl,
		Hosts:          []*utils.TPDispatcherHostProfile{},
	}

	rcv := APItoModelTPDispatcherProfile(tpDPP)
	exp := TPDispatcherProfiles{{
		Tpid:               str,
		Tenant:             str,
		ID:                 str,
		Subsystems:         str,
		FilterIDs:          str,
		ActivationInterval: "test;test",
		Strategy:           str,
		StrategyParameters: "test;test",
		ConnID:             "",
		ConnFilterIDs:      "",
		ConnWeight:         0,
		ConnBlocker:        false,
		ConnParameters:     "",
		Weight:             fl,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	rcv = APItoModelTPDispatcherProfile(nil)

	if rcv != nil {
		t.Error(rcv)
	}
}

func TestModelHelpersAsTPDispatcherHosts(t *testing.T) {
	tps := TPDispatcherHosts{{
		Tpid:      str,
		Tenant:    str,
		ID:        str,
		Address:   utils.EmptyString,
		Transport: str,
		TLS:       bl,
	}, {
		Tpid:      str,
		Tenant:    str,
		ID:        str,
		Address:   str,
		Transport: utils.EmptyString,
		TLS:       bl,
	}}

	rcv := tps.AsTPDispatcherHosts()
	exp := []*utils.TPDispatcherHost{{
		TPid:   str,
		Tenant: str,
		ID:     str,
		Conns: []*utils.TPDispatcherHostConn{{
			Address:   str,
			Transport: "*json",
			TLS:       true,
		}},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoResource(t *testing.T) {
	errStr := "test`"
	tpRL := &utils.TPResourceProfile{
		TPid:      str,
		Tenant:    str,
		ID:        str,
		FilterIDs: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		UsageTTL:          errStr,
		Limit:             str,
		AllocationMessage: str,
		Blocker:           bl,
		Stored:            bl,
		Weight:            fl,
		ThresholdIDs:      slc,
	}

	_, err := APItoResource(tpRL, "")

	if err != nil {
		if err.Error() != errors.New("time: invalid duration "+`"`+errStr+`"`).Error() {
			t.Error(err)
		}
	}

	tpRL.UsageTTL = utils.EmptyString

	_, err = APItoResource(tpRL, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}

	tpRL.ActivationInterval = nil
	_, err = APItoResource(tpRL, "")

	if err != nil {
		if err.Error() != `strconv.ParseFloat: parsing "test": invalid syntax` {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoStats(t *testing.T) {
	errStr := "test`"
	tpST := &utils.TPStatProfile{
		TTL: errStr,
	}

	_, err := APItoStats(tpST, "")

	if err != nil {
		if err.Error() != errors.New("time: invalid duration "+`"`+errStr+`"`).Error() {
			t.Error(err)
		}
	}

	tpST.TTL = utils.EmptyString

	_, err = APItoStats(tpST, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoThresholdProfile(t *testing.T) {
	errStr := "test`"
	tpST := &utils.TPThresholdProfile{
		MinSleep: errStr,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err := APItoThresholdProfile(tpST, "")
	if err != nil {
		if err.Error() != errors.New("time: invalid duration "+`"`+errStr+`"`).Error() {
			t.Error(err)
		}
	}

	tpST.MinSleep = ""
	_, err = APItoThresholdProfile(tpST, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoDispatcherProfile(t *testing.T) {
	tpDPP := &utils.TPDispatcherProfile{
		StrategyParams: []any{str},
		Hosts: []*utils.TPDispatcherHostProfile{
			{
				Params: []any{""},
			},
		},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err := APItoDispatcherProfile(tpDPP, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoChargerProfile(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err := APItoChargerProfile(tpCPP, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelTPCharger(t *testing.T) {
	tpCPP := &utils.TPChargerProfile{
		FilterIDs: []string{},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	rcv := APItoModelTPCharger(tpCPP)
	exp := TPChargers{{
		ActivationInterval: "test;test",
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}

	tpCPP2 := &utils.TPChargerProfile{
		FilterIDs: []string{"test"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	rcv = APItoModelTPCharger(tpCPP2)
	exp = TPChargers{{
		FilterIDs:          str,
		ActivationInterval: "test;test",
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoAttributeProfile(t *testing.T) {
	errStr := "test`"
	tpAttr := &utils.TPAttributeProfile{
		Attributes: []*utils.TPAttribute{
			{
				Value: errStr,
			},
		},
	}

	_, err := APItoAttributeProfile(tpAttr, "")

	if err != nil {
		if err.Error() != "Unclosed unspilit syntax" {
			t.Error(err)
		}
	}

	tpAttr2 := &utils.TPAttributeProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err = APItoAttributeProfile(tpAttr2, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoSupplierProfile(t *testing.T) {
	tpSPP := &utils.TPSupplierProfile{
		SortingParameters: slc,
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err := APItoSupplierProfile(tpSPP, str)

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoFilter(t *testing.T) {
	tpTH := &utils.TPFilterProfile{
		Filters: []*utils.TPFilter{
			{
				Type:   utils.MetaRSR,
				Values: []string{"test)"},
			},
		},
	}

	_, err := APItoFilter(tpTH, "")

	if err != nil {
		if err.Error() != "invalid RSRFilter start rule in string: <test)>" {
			t.Error(err)
		}
	}

	tpTH2 := &utils.TPFilterProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err = APItoFilter(tpTH2, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelTPThreshold(t *testing.T) {
	th := &utils.TPThresholdProfile{
		ActionIDs: []string{str},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	rcv := APItoModelTPThreshold(th)
	exp := TpThresholds{{
		ActivationInterval: "test;test",
		ActionIDs:          str,
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAPItoStatsErr(t *testing.T) {
	tpST := &utils.TPStatProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}

	_, err := APItoStats(tpST, "test")

	if err != nil {
		if err.Error() != "unknown time zone test" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAPItoModelResourceErrors(t *testing.T) {
	rl := &utils.TPResourceProfile{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		ThresholdIDs: []string{str, str},
	}

	rcv := APItoModelResource(rl)
	exp := TpResources{{
		ThresholdIDs:       "test;test",
		ActivationInterval: "test;test",
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsMapRates(t *testing.T) {
	tps := TpRates{{
		Id:                 1,
		Tpid:               str,
		Tag:                str,
		ConnectFee:         fl,
		Rate:               fl,
		RateUnit:           str,
		RateIncrement:      str,
		GroupIntervalStart: str,
	}}

	_, err := tps.AsMapRates()

	if err != nil {
		if err.Error() != `time: invalid duration "test"` {
			t.Error(err)
		}
	}
}

func TestModelHelpersAsTPRates(t *testing.T) {
	tps := TpRates{{
		Id:                 1,
		Tpid:               str,
		Tag:                str,
		ConnectFee:         fl,
		Rate:               fl,
		RateUnit:           str,
		RateIncrement:      str,
		GroupIntervalStart: str,
	}}

	_, err := tps.AsTPRates()

	if err != nil {
		if err.Error() != `time: invalid duration "test"` {
			t.Error(err)
		}
	}
}

func TestModelHelperscsvDump(t *testing.T) {
	type Test struct {
		Fl float64 `index:"a"`
	}
	tst := Test{1.2}

	_, err := csvDump(&tst)

	if err != nil {
		if err.Error() != "invalid Test.Fl index a" {
			t.Error(err)
		}
	}
}

func TestModelHelpersAsTPResources(t *testing.T) {
	tps := TpResources{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPResources()
	exp := []*utils.TPResourceProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPThreshold(t *testing.T) {
	tps := TpThresholds{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPThreshold()
	exp := []*utils.TPThresholdProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPFilter(t *testing.T) {
	tps := TpFilterS{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPFilter()
	exp := []*utils.TPFilterProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPSuppliers(t *testing.T) {
	tps := TpSuppliers{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPSuppliers()
	exp := []*utils.TPSupplierProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		SortingParameters: []string{},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPAttributes(t *testing.T) {
	tps := TPAttributes{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPAttributes()
	exp := []*utils.TPAttributeProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPChargers(t *testing.T) {
	tps := TPChargers{{
		ActivationInterval: "test;test",
	}}

	rcv := tps.AsTPChargers()
	exp := []*utils.TPChargerProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPDispatcherProfiles(t *testing.T) {
	tps := TPDispatcherProfiles{{
		ActivationInterval: "test;test",
		StrategyParameters: str,
	}}

	rcv := tps.AsTPDispatcherProfiles()
	exp := []*utils.TPDispatcherProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		StrategyParams: []any{str},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelpersAsTPStats(t *testing.T) {
	tps := TpStats{{
		ActivationInterval: "test;test",
		MetricIDs:          str,
		MetricFilterIDs:    str,
	}}

	rcv := tps.AsTPStats()
	exp := []*utils.TPStatProfile{{
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: str,
			ExpiryTime:     str,
		},
		Metrics: []*utils.MetricWithFilters{
			{
				FilterIDs: []string{str},
				MetricID:  str,
			},
		},
	}}

	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("expected %v, received %v", utils.ToJSON(exp), utils.ToJSON(rcv))
	}
}

func TestModelHelperscsvLoadErrors(t *testing.T) {
	type Test struct {
		Fl float64 `index:"a"`
	}
	type Test2 struct {
		Field string `index:"0" re:"//w+"`
	}
	tst := Test{Fl: 1.2}
	tst2 := Test2{}
	type args struct {
		s      any
		values []string
	}
	tests := []struct {
		name string
		args args
		err  string
	}{
		{
			name: "index tag error",
			args: args{tst, []string{"test"}},
			err:  "invalid Test.Fl index a",
		},
		{
			name: "regex tag error",
			args: args{tst2, []string{"123"}},
			err:  "invalid Test2.Field value 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := csvLoad(tt.args.s, tt.args.values)

			if err != nil {
				if err.Error() != tt.err {
					t.Error(err)
				}
			}
		})
	}
}
