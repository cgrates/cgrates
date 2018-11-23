/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or56
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
		[]string{"TEST_DEST", "49"},
		[]string{"TEST_DEST", "49176"},
		[]string{"TEST_DEST", "49151"},
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
	eTPDestinations := []*utils.TPDestination{&utils.TPDestination{TPid: "TEST_TPID", ID: "TEST_DEST",
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
			&utils.RateSlot{
				ConnectFee:         0.100,
				Rate:               0.200,
				RateUnit:           "60",
				RateIncrement:      "60",
				GroupIntervalStart: "0"},
			&utils.RateSlot{
				ConnectFee:         0.0,
				Rate:               0.1,
				RateUnit:           "1",
				RateIncrement:      "60",
				GroupIntervalStart: "60"},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_RATEID", "0.1", "0.2", "60", "60", "0"},
		[]string{"TEST_RATEID", "0", "0.1", "1", "60", "60"},
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
			&utils.DestinationRate{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
			&utils.DestinationRate{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_DSTRATE", "TEST_DEST1", "TEST_RATE1", "*up", "4", "0", ""},
		[]string{"TEST_DSTRATE", "TEST_DEST2", "TEST_RATE2", "*up", "4", "0", ""},
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
		[]string{"TEST_TIMING", "*any", "*any", "*any", "1;2;4", "00:00:01"},
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

func TestAPItoModelStats(t *testing.T) {
	tpS := &utils.TPStats{
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
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "*tcc",
				Parameters: "",
			},
			&utils.MetricWithParams{
				MetricID:   "*sum:Value",
				Parameters: "Value",
			},
			&utils.MetricWithParams{
				MetricID:   "*average:Value",
				Parameters: "Value",
			},
			&utils.MetricWithParams{
				MetricID:   "*sum:Cost",
				Parameters: "Cost",
			},
			&utils.MetricWithParams{
				MetricID:   "*average:Usage",
				Parameters: "Usage",
			},
		},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     2,
		ThresholdIDs: []string{"Th1", "Th2", "Th3", "Th4"},
	}
	rcv := APItoModelStats(tpS)
	eRcv := []*TpStats{
		&TpStats{
			Tpid:               "TPS1",
			Tenant:             "cgrates.org",
			ID:                 "Stat1",
			FilterIDs:          "*string:Account:1002",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			Metrics:            "*tcc;*sum:Value;*average:Value;*sum:Cost;*average:Usage",
			Parameters:         "Value;Cost;Usage",
			ThresholdIDs:       "Th1;Th2;Th3;Th4",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
	}
	if !reflect.DeepEqual(eRcv[0].Tenant, rcv[0].Tenant) {
		t.Errorf("Expecting: %+v, received: %+v", eRcv[0].Tenant, rcv[0].Tenant)
	} else if !reflect.DeepEqual(eRcv[0].ID, rcv[0].ID) {
		t.Errorf("Expecting: %+v, received: %+v", eRcv[0].ID, rcv[0].ID)
	} else if !reflect.DeepEqual(eRcv[0].FilterIDs, rcv[0].FilterIDs) {
		t.Errorf("Expecting: %+v, received: %+v", eRcv[0].FilterIDs, rcv[0].FilterIDs)
	} else if !reflect.DeepEqual(eRcv[0].ActivationInterval, rcv[0].ActivationInterval) {
		t.Errorf("Expecting: %+v, received: %+v", eRcv[0].ActivationInterval, rcv[0].ActivationInterval)
	} else if !reflect.DeepEqual(eRcv[0].QueueLength, rcv[0].QueueLength) {
		t.Errorf("Expecting: %+v, received: %+v", eRcv[0].QueueLength, rcv[0].QueueLength)
	} else if !reflect.DeepEqual(len(eRcv[0].Metrics), len(rcv[0].Metrics)) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv[0].Metrics), len(rcv[0].Metrics))
	} else if !reflect.DeepEqual(len(eRcv[0].Parameters), len(rcv[0].Parameters)) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv[0].Parameters), len(rcv[0].Parameters))
	} else if !reflect.DeepEqual(len(eRcv[0].ThresholdIDs), len(rcv[0].ThresholdIDs)) {
		t.Errorf("Expecting: %+v, received: %+v", len(eRcv[0].ThresholdIDs), len(rcv[0].ThresholdIDs))
	}
}

func TestTPRatingPlanAsExportSlice(t *testing.T) {
	tpRpln := &utils.TPRatingPlan{
		TPid: "TEST_TPID",
		ID:   "TEST_RPLAN",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			&utils.TPRatingPlanBinding{
				DestinationRatesId: "TEST_DSTRATE1",
				TimingId:           "TEST_TIMING1",
				Weight:             10.0},
			&utils.TPRatingPlanBinding{
				DestinationRatesId: "TEST_DSTRATE2",
				TimingId:           "TEST_TIMING2",
				Weight:             20.0},
		}}
	expectedSlc := [][]string{
		[]string{"TEST_RPLAN", "TEST_DSTRATE1", "TEST_TIMING1", "10"},
		[]string{"TEST_RPLAN", "TEST_DSTRATE2", "TEST_TIMING2", "20"},
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
		TPid:      "TEST_TPID",
		LoadId:    "TEST_LOADID",
		Direction: utils.OUT,
		Tenant:    "cgrates.org",
		Category:  "call",
		Subject:   "*any",
		RatingPlanActivations: []*utils.TPRatingActivation{
			&utils.TPRatingActivation{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
			&utils.TPRatingActivation{
				ActivationTime:   "2014-01-15T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN2",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	expectedSlc := [][]string{
		[]string{utils.OUT, "cgrates.org", "call", "*any", "2014-01-14T00:00:00Z", "TEST_RPLAN1", "subj1;subj2", ""},
		[]string{utils.OUT, "cgrates.org", "call", "*any", "2014-01-15T00:00:00Z", "TEST_RPLAN2", "subj1;subj2", ""},
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
			&utils.TPAction{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Directions:      utils.OUT,
				Units:           "5.0",
				ExpiryTime:      "*never",
				DestinationIds:  "*any",
				RatingSubject:   "special1",
				Categories:      "call",
				SharedGroups:    "GROUP1",
				BalanceWeight:   "10.0",
				ExtraParameters: "",
				Weight:          10.0},
			&utils.TPAction{
				Identifier:      "*http_post",
				BalanceType:     "",
				Directions:      "",
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
		[]string{"TEST_ACTIONS", "*topup_reset", "", "", "", "*monetary", utils.OUT, "call", "*any", "special1", "GROUP1", "*never", "", "5.0", "10.0", "", "", "10"},
		[]string{"TEST_ACTIONS", "*http_post", "http://localhost/&param1=value1", "", "", "", "", "", "", "", "", "", "", "0.0", "0.0", "", "", "20"},
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
			&utils.TPSharedGroup{
				Account:       "*any",
				Strategy:      "*highest",
				RatingSubject: "special1"},
			&utils.TPSharedGroup{
				Account:       "second",
				Strategy:      "*highest",
				RatingSubject: "special2"},
		},
	}
	expectedSlc := [][]string{
		[]string{"SHARED_GROUP_TEST", "*any", "*highest", "special1"},
		[]string{"SHARED_GROUP_TEST", "second", "*highest", "special2"},
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

//#Direction,Tenant,Category,Account,Subject,RunId,RunFilter,ReqTypeField,DirectionField,TenantField,CategoryField,AccountField,SubjectField,DestinationField,SetupTimeField,AnswerTimeField,UsageField
//*out,cgrates.org,call,1001,1001,derived_run1,,^rated,*default,*default,*default,*default,^1002,*default,*default,*default,*default
func TestTPDerivedChargersAsExportSlice(t *testing.T) {
	dcs := &utils.TPDerivedChargers{
		TPid:      "TEST_TPID",
		LoadId:    "TEST_LOADID",
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "1001",
		Subject:   "1001",
		DerivedChargers: []*utils.TPDerivedCharger{
			&utils.TPDerivedCharger{
				RunId:                "derived_run1",
				RunFilters:           "",
				ReqTypeField:         "^rated",
				DirectionField:       utils.META_DEFAULT,
				TenantField:          utils.META_DEFAULT,
				CategoryField:        utils.META_DEFAULT,
				AccountField:         utils.META_DEFAULT,
				SubjectField:         "^1002",
				DestinationField:     utils.META_DEFAULT,
				SetupTimeField:       utils.META_DEFAULT,
				PddField:             utils.META_DEFAULT,
				AnswerTimeField:      utils.META_DEFAULT,
				UsageField:           utils.META_DEFAULT,
				SupplierField:        utils.META_DEFAULT,
				DisconnectCauseField: utils.META_DEFAULT,
				CostField:            utils.META_DEFAULT,
				RatedField:           utils.META_DEFAULT,
			},
			&utils.TPDerivedCharger{
				RunId:                "derived_run2",
				RunFilters:           "",
				ReqTypeField:         "^rated",
				DirectionField:       utils.META_DEFAULT,
				TenantField:          utils.META_DEFAULT,
				CategoryField:        utils.META_DEFAULT,
				AccountField:         "^1002",
				SubjectField:         utils.META_DEFAULT,
				DestinationField:     utils.META_DEFAULT,
				SetupTimeField:       utils.META_DEFAULT,
				PddField:             utils.META_DEFAULT,
				AnswerTimeField:      utils.META_DEFAULT,
				UsageField:           utils.META_DEFAULT,
				SupplierField:        utils.META_DEFAULT,
				DisconnectCauseField: utils.META_DEFAULT,
				RatedField:           utils.META_DEFAULT,
				CostField:            utils.META_DEFAULT,
			},
		},
	}
	expectedSlc := [][]string{
		[]string{"*out", "cgrates.org", "call", "1001", "1001", "",
			"derived_run1", "", "^rated", utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, "^1002", utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT},
		[]string{"*out", "cgrates.org", "call", "1001", "1001", "",
			"derived_run2", "", "^rated", utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, "^1002", utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT, utils.META_DEFAULT},
	}
	ms := APItoModelDerivedCharger(dcs)
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
			&utils.TPActionTiming{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
			&utils.TPActionTiming{
				ActionsId: "TOPUP_RST_5",
				TimingId:  "ASAP",
				Weight:    20.0},
		},
	}
	expectedSlc := [][]string{
		[]string{"PACKAGE_10", "TOPUP_RST_10", "ASAP", "10"},
		[]string{"PACKAGE_10", "TOPUP_RST_5", "ASAP", "20"},
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
			&utils.TPActionTrigger{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "1",
				ThresholdType:         "*min_balance",
				ThresholdValue:        2.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b1",
				BalanceType:           "*monetary",
				BalanceDirections:     "*out",
				BalanceDestinationIds: "",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				MinQueuedItems:        0,
				ActionsId:             "LOG_WARNING",
				Weight:                10},
			&utils.TPActionTrigger{
				Id:                    "STANDARD_TRIGGERS",
				UniqueID:              "2",
				ThresholdType:         "*max_event_counter",
				ThresholdValue:        5.0,
				Recurrent:             false,
				MinSleep:              "0",
				BalanceId:             "b2",
				BalanceType:           "*monetary",
				BalanceDirections:     "*out",
				BalanceDestinationIds: "FS_USERS",
				BalanceWeight:         "0.0",
				BalanceExpirationDate: "*never",
				BalanceTimingTags:     "T1",
				BalanceRatingSubject:  "special1",
				BalanceCategories:     "call",
				BalanceSharedGroups:   "SHARED_1",
				BalanceBlocker:        "false",
				BalanceDisabled:       "false",
				MinQueuedItems:        0,
				ActionsId:             "LOG_WARNING",
				Weight:                10},
		},
	}
	expectedSlc := [][]string{
		[]string{"STANDARD_TRIGGERS", "1", "*min_balance", "2", "false", "0", "", "", "b1", "*monetary", "*out", "call", "", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "0", "LOG_WARNING", "10"},
		[]string{"STANDARD_TRIGGERS", "2", "*max_event_counter", "5", "false", "0", "", "", "b2", "*monetary", "*out", "call", "FS_USERS", "special1", "SHARED_1", "*never", "T1", "0.0", "false", "false", "0", "LOG_WARNING", "10"},
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
		[]string{"cgrates.org", "1001", "PACKAGE_10_SHARED_A_5", "STANDARD_TRIGGERS", "false", "false"},
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
		&TpResource{
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
		&TpResource{
			Tpid:         "TEST_TPID",
			ID:           "ResGroup1",
			Tenant:       "cgrates.org",
			FilterIDs:    "FLTR_RES_GR1",
			ThresholdIDs: "WARN3"},
		&TpResource{
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
	eTPs := []*utils.TPResource{
		&utils.TPResource{
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
		&utils.TPResource{
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
	tpRL := &utils.TPResource{
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

func TestTPStatsAsTPStats(t *testing.T) {
	tps := []*TpStats{
		&TpStats{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			Metrics:            "*asr;*acc;*tcc;*acd;*tcd;*pdd",
			Parameters:         "",
			ThresholdIDs:       "THRESH1;THRESH2",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&TpStats{
			Tpid:               "TEST_TPID",
			Tenant:             "cgrates.org",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			Metrics:            "*sum;*average;*tcc",
			Parameters:         "BalanceValue",
			ThresholdIDs:       "THRESH3",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
		&TpStats{
			Tpid:               "TEST_TPID",
			Tenant:             "itsyscom.com",
			ID:                 "Stats1",
			FilterIDs:          "FLTR_1",
			ActivationInterval: "2014-07-29T15:00:00Z",
			QueueLength:        100,
			TTL:                "1s",
			MinItems:           2,
			Metrics:            "*sum;*average;*tcc",
			Parameters:         "BalanceValue",
			ThresholdIDs:       "THRESH4",
			Stored:             true,
			Blocker:            true,
			Weight:             20.0,
		},
	}
	eTPs := []*utils.TPStats{
		&utils.TPStats{
			TPid:      tps[0].Tpid,
			Tenant:    tps[0].Tenant,
			ID:        tps[0].ID,
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			QueueLength: tps[0].QueueLength,
			TTL:         tps[0].TTL,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{MetricID: "*asr", Parameters: ""},
				&utils.MetricWithParams{MetricID: "*acc", Parameters: ""},
				&utils.MetricWithParams{MetricID: "*acd", Parameters: ""},
				&utils.MetricWithParams{MetricID: "*tcd", Parameters: ""},
				&utils.MetricWithParams{MetricID: "*pdd", Parameters: ""},
				&utils.MetricWithParams{MetricID: "*sum:BalanceValue", Parameters: "BalanceValue"},
				&utils.MetricWithParams{MetricID: "*average:BalanceValue", Parameters: "BalanceValue"},
				&utils.MetricWithParams{MetricID: "*tcc:BalanceValue", Parameters: "BalanceValue"},
			},
			MinItems:     tps[0].MinItems,
			ThresholdIDs: []string{"THRESH1", "THRESH2", "THRESH3"},
			Stored:       tps[0].Stored,
			Blocker:      tps[0].Blocker,
			Weight:       tps[0].Weight,
		},
		&utils.TPStats{
			TPid:      tps[0].Tpid,
			ID:        tps[0].ID,
			Tenant:    tps[2].Tenant,
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.TPActivationInterval{
				ActivationTime: tps[0].ActivationInterval,
			},
			QueueLength: tps[0].QueueLength,
			TTL:         tps[0].TTL,
			Metrics: []*utils.MetricWithParams{
				&utils.MetricWithParams{MetricID: "*sum:BalanceValue", Parameters: "BalanceValue"},
				&utils.MetricWithParams{MetricID: "*average:BalanceValue", Parameters: "BalanceValue"},
				&utils.MetricWithParams{MetricID: "*tcc:BalanceValue", Parameters: "BalanceValue"},
			},
			MinItems:     tps[0].MinItems,
			ThresholdIDs: []string{"THRESH4"},
			Stored:       tps[0].Stored,
			Blocker:      tps[0].Blocker,
			Weight:       tps[0].Weight,
		},
	}
	rcvTPs := TpStatsS(tps).AsTPStats()
	if !(reflect.DeepEqual(eTPs[1].TPid, rcvTPs[1].TPid) && reflect.DeepEqual(eTPs[0].TPid, rcvTPs[0].TPid)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[0].TPid), utils.ToIJSON(rcvTPs[0].TPid))
	} else if !(reflect.DeepEqual(eTPs[1].ID, rcvTPs[1].ID) && reflect.DeepEqual(eTPs[0].ID, rcvTPs[0].ID)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[0].ID), utils.ToIJSON(rcvTPs[0].ID))
	} else if !(reflect.DeepEqual(eTPs[1].FilterIDs, rcvTPs[1].FilterIDs) && reflect.DeepEqual(eTPs[0].FilterIDs, rcvTPs[0].FilterIDs)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[0].FilterIDs), utils.ToIJSON(rcvTPs[0].FilterIDs))
	} else if len(utils.ToIJSON(eTPs[0].Metrics)) != len(utils.ToIJSON(rcvTPs[0].Metrics)) &&
		len(utils.ToIJSON(eTPs[1].Metrics)) != len(utils.ToIJSON(rcvTPs[1].Metrics)) &&
		len(utils.ToIJSON(eTPs[1].Metrics)) != len(utils.ToIJSON(rcvTPs[0].Metrics)) &&
		len(utils.ToIJSON(eTPs[0].Metrics)) != len(utils.ToIJSON(rcvTPs[1].Metrics)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[0].Metrics), utils.ToIJSON(rcvTPs[0].Metrics))
	}
	if !(reflect.DeepEqual(eTPs[1].TPid, rcvTPs[1].TPid) && reflect.DeepEqual(eTPs[0].TPid, rcvTPs[0].TPid)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[1].TPid), utils.ToIJSON(rcvTPs[1].TPid))
	} else if !(reflect.DeepEqual(eTPs[1].ID, rcvTPs[1].ID) && reflect.DeepEqual(eTPs[0].ID, rcvTPs[0].ID)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[1].ID), utils.ToIJSON(rcvTPs[1].ID))
	} else if !(reflect.DeepEqual(eTPs[1].FilterIDs, rcvTPs[1].FilterIDs) && reflect.DeepEqual(eTPs[0].FilterIDs, rcvTPs[0].FilterIDs)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[1].FilterIDs), utils.ToIJSON(rcvTPs[1].FilterIDs))
	} else if len(utils.ToIJSON(eTPs[0].Metrics)) != len(utils.ToIJSON(rcvTPs[0].Metrics)) &&
		len(utils.ToIJSON(eTPs[1].Metrics)) != len(utils.ToIJSON(rcvTPs[1].Metrics)) &&
		len(utils.ToIJSON(eTPs[1].Metrics)) != len(utils.ToIJSON(rcvTPs[0].Metrics)) &&
		len(utils.ToIJSON(eTPs[0].Metrics)) != len(utils.ToIJSON(rcvTPs[1].Metrics)) {
		t.Errorf("\nExpecting:\n%+v\nReceived:\n%+v", utils.ToIJSON(eTPs[1].Metrics), utils.ToIJSON(rcvTPs[1].Metrics))
	}
}

func TestAPItoTPStats(t *testing.T) {
	tps := &utils.TPStats{
		TPid:               testTPID,
		ID:                 "Stats1",
		FilterIDs:          []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{ActivationTime: "2014-07-29T15:00:00Z"},
		QueueLength:        100,
		TTL:                "1s",
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{MetricID: "*asr", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acd", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acc", Parameters: ""},
		},
		MinItems:     1,
		ThresholdIDs: []string{"THRESH1", "THRESH2"},
		Stored:       false,
		Blocker:      false,
		Weight:       20.0,
	}

	eTPs := &StatQueueProfile{ID: tps.ID,
		QueueLength: tps.QueueLength,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{MetricID: "*asr", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acd", Parameters: ""},
			&utils.MetricWithParams{MetricID: "*acc", Parameters: ""},
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

func TestTPThresholdsAsTPThreshold(t *testing.T) {
	tps := []*TpThreshold{
		&TpThreshold{
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
	eTPs := []*utils.TPThreshold{
		&utils.TPThreshold{
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
		&utils.TPThreshold{
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
	rcvTPs := TpThresholdS(tps).AsTPThreshold()
	if !reflect.DeepEqual(eTPs[0], rcvTPs[0]) && !reflect.DeepEqual(eTPs[1], rcvTPs[0]) {
		t.Errorf("Expecting: %+v , Received: %+v", utils.ToIJSON(eTPs), utils.ToIJSON(rcvTPs))
	}
}

func TestAPItoModelTPThreshold(t *testing.T) {
	th := &utils.TPThreshold{
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
	models := TpThresholdS{
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
	th := &utils.TPThreshold{
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
	models := TpThresholdS{
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
	th := &utils.TPThreshold{
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
	models := TpThresholdS{
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
	th := &utils.TPThreshold{
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
	models := TpThresholdS{
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
	th := &utils.TPThreshold{
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
	tps := &utils.TPThreshold{
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
		&TpFilter{
			Tpid:              "TEST_TPID",
			ID:                "Filter1",
			FilterType:        MetaPrefix,
			FilterFieldName:   "Account",
			FilterFieldValues: "1001;1002",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		&utils.TPFilterProfile{
			TPid: tps[0].Tpid,
			ID:   tps[0].ID,
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					Type:      MetaPrefix,
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
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
		&TpFilter{
			Tpid:              "TEST_TPID",
			Tenant:            "cgrates.org",
			ID:                "Filter1",
			FilterType:        MetaPrefix,
			FilterFieldName:   "Account",
			FilterFieldValues: "1001;1002",
		},
		&TpFilter{
			Tpid:              "TEST_TPID",
			Tenant:            "anotherTenant",
			ID:                "Filter1",
			FilterType:        MetaPrefix,
			FilterFieldName:   "Account",
			FilterFieldValues: "1010",
		},
	}
	eTPs := []*utils.TPFilterProfile{
		&utils.TPFilterProfile{
			TPid:   tps[0].Tpid,
			Tenant: "cgrates.org",
			ID:     tps[0].ID,
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					Type:      MetaPrefix,
					FieldName: "Account",
					Values:    []string{"1001", "1002"},
				},
			},
		},
		&utils.TPFilterProfile{
			TPid:   tps[1].Tpid,
			Tenant: "anotherTenant",
			ID:     tps[1].ID,
			Filters: []*utils.TPFilter{
				&utils.TPFilter{
					Type:      MetaPrefix,
					FieldName: "Account",
					Values:    []string{"1010"},
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
			&utils.TPFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001", "1002"},
			},
		},
	}

	eTPs := &Filter{
		Tenant: "cgrates.org",
		ID:     tps.ID,
		Rules: []*FilterRule{
			&FilterRule{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001", "1002"},
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
			&FilterRule{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001", "1002"},
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
			&utils.TPFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001", "1002"},
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
			&utils.TPAttribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	mapSubstitutes := make(map[string]map[interface{}]*Attribute)
	mapSubstitutes["FL1"] = make(map[interface{}]*Attribute)
	mapSubstitutes["FL1"]["In1"] = &Attribute{
		FieldName:  "FL1",
		Initial:    "In1",
		Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
		Append:     true,
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
			&Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				Append:     true,
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
			&utils.TPAttribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
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
			FieldName:          "FL1",
			Initial:            "In1",
			Substitute:         "Al1",
			Append:             true,
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
			FieldName:          "FL1",
			Initial:            "In1",
			Substitute:         "Al1",
			Append:             true,
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
			&utils.TPAttribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
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
			&utils.TPAttribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
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

//Number of FilterIDs and AttributeIDs are equal
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

//Number of FilterIDs is smaller than AttributeIDs
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

//Number of FilterIDs is greater than AttributeIDs
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

//len(AttributeIDs) is 0
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

//len(FilterIDs) is 0
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

//both len(AttributeIDs) and len(FilterIDs) are 0
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
