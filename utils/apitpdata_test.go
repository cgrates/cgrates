/*
Real-time Charging System for Telecom & ISP environments
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

import (
	"reflect"
	"testing"
	"time"
)

func TestTPDestinationAsExportSlice(t *testing.T) {
	tpDst := &TPDestination{
		TPid:          "TEST_TPID",
		DestinationId: "TEST_DEST",
		Prefixes:      []string{"49", "49176", "49151"},
	}
	expectedSlc := [][]string{
		[]string{"TEST_DEST", "49"},
		[]string{"TEST_DEST", "49176"},
		[]string{"TEST_DEST", "49151"},
	}
	if slc := tpDst.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRateAsExportSlice(t *testing.T) {
	tpRate := &TPRate{
		TPid:   "TEST_TPID",
		RateId: "TEST_RATEID",
		RateSlots: []*RateSlot{
			&RateSlot{
				ConnectFee:         0.100,
				Rate:               0.200,
				RateUnit:           "60",
				RateIncrement:      "60",
				GroupIntervalStart: "0"},
			&RateSlot{
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
	if slc := tpRate.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPDestinationRateAsExportSlice(t *testing.T) {
	tpDstRate := &TPDestinationRate{
		TPid:              "TEST_TPID",
		DestinationRateId: "TEST_DSTRATE",
		DestinationRates: []*DestinationRate{
			&DestinationRate{
				DestinationId:    "TEST_DEST1",
				RateId:           "TEST_RATE1",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
			&DestinationRate{
				DestinationId:    "TEST_DEST2",
				RateId:           "TEST_RATE2",
				RoundingMethod:   "*up",
				RoundingDecimals: 4},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_DSTRATE", "TEST_DEST1", "TEST_RATE1", "*up", "4"},
		[]string{"TEST_DSTRATE", "TEST_DEST2", "TEST_RATE2", "*up", "4"},
	}
	if slc := tpDstRate.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}

}

func TestApierTPTimingAsExportSlice(t *testing.T) {
	tpTiming := &ApierTPTiming{
		TPid:      "TEST_TPID",
		TimingId:  "TEST_TIMING",
		Years:     "*any",
		Months:    "*any",
		MonthDays: "*any",
		WeekDays:  "1;2;4",
		Time:      "00:00:01"}
	expectedSlc := [][]string{
		[]string{"TEST_TIMING", "*any", "*any", "*any", "1;2;4", "00:00:01"},
	}
	if slc := tpTiming.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRatingPlanAsExportSlice(t *testing.T) {
	tpRpln := &TPRatingPlan{
		TPid:         "TEST_TPID",
		RatingPlanId: "TEST_RPLAN",
		RatingPlanBindings: []*TPRatingPlanBinding{
			&TPRatingPlanBinding{
				DestinationRatesId: "TEST_DSTRATE1",
				TimingId:           "TEST_TIMING1",
				Weight:             10.0},
			&TPRatingPlanBinding{
				DestinationRatesId: "TEST_DSTRATE2",
				TimingId:           "TEST_TIMING2",
				Weight:             20.0},
		}}
	expectedSlc := [][]string{
		[]string{"TEST_RPLAN", "TEST_DSTRATE1", "TEST_TIMING1", "10"},
		[]string{"TEST_RPLAN", "TEST_DSTRATE2", "TEST_TIMING2", "20"},
	}
	if slc := tpRpln.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPRatingProfileAsExportSlice(t *testing.T) {
	tpRpf := &TPRatingProfile{
		TPid:      "TEST_TPID",
		LoadId:    "TEST_LOADID",
		Direction: OUT,
		Tenant:    "cgrates.org",
		Category:  "call",
		Subject:   "*any",
		RatingPlanActivations: []*TPRatingActivation{
			&TPRatingActivation{
				ActivationTime:   "2014-01-14T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN1",
				FallbackSubjects: "subj1;subj2"},
			&TPRatingActivation{
				ActivationTime:   "2014-01-15T00:00:00Z",
				RatingPlanId:     "TEST_RPLAN2",
				FallbackSubjects: "subj1;subj2"},
		},
	}
	expectedSlc := [][]string{
		[]string{OUT, "cgrates.org", "call", "*any", "2014-01-14T00:00:00Z", "TEST_RPLAN1", "subj1;subj2"},
		[]string{OUT, "cgrates.org", "call", "*any", "2014-01-15T00:00:00Z", "TEST_RPLAN2", "subj1;subj2"},
	}
	if slc := tpRpf.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionsAsExportSlice(t *testing.T) {
	tpActs := &TPActions{
		TPid:      "TEST_TPID",
		ActionsId: "TEST_ACTIONS",
		Actions: []*TPAction{
			&TPAction{
				Identifier:      "*topup_reset",
				BalanceType:     "*monetary",
				Direction:       OUT,
				Units:           5.0,
				ExpiryTime:      "*never",
				DestinationId:   "*any",
				RatingSubject:   "special1",
				Category:        "call",
				SharedGroup:     "GROUP1",
				BalanceWeight:   10.0,
				ExtraParameters: "",
				Weight:          10.0},
			&TPAction{
				Identifier:      "*http_post",
				BalanceType:     "",
				Direction:       "",
				Units:           0.0,
				ExpiryTime:      "",
				DestinationId:   "",
				RatingSubject:   "",
				Category:        "",
				SharedGroup:     "",
				BalanceWeight:   0.0,
				ExtraParameters: "http://localhost/&param1=value1",
				Weight:          20.0},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_ACTIONS", "*topup_reset", "", "*monetary", OUT, "call", "*any", "special1", "GROUP1", "*never", "5", "10", "10"},
		[]string{"TEST_ACTIONS", "*http_post", "http://localhost/&param1=value1", "", "", "", "", "", "", "", "0", "0", "20"},
	}
	if slc := tpActs.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

// SHARED_A,*any,*highest,
func TestTPSharedGroupsAsExportSlice(t *testing.T) {
	tpSGs := &TPSharedGroups{
		TPid:           "TEST_TPID",
		SharedGroupsId: "SHARED_GROUP_TEST",
		SharedGroups: []*TPSharedGroup{
			&TPSharedGroup{
				Account:       "*any",
				Strategy:      "*highest",
				RatingSubject: "special1"},
			&TPSharedGroup{
				Account:       "second",
				Strategy:      "*highest",
				RatingSubject: "special2"},
		},
	}
	expectedSlc := [][]string{
		[]string{"SHARED_GROUP_TEST", "*any", "*highest", "special1"},
		[]string{"SHARED_GROUP_TEST", "second", "*highest", "special2"},
	}
	if slc := tpSGs.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

//*in,cgrates.org,*any,EU_LANDLINE,LCR_STANDARD,*static,ivo;dan;rif,2012-01-01T00:00:00Z,10
func TestTPLcrRulesAsExportSlice(t *testing.T) {
	lcr := &TPLcrRules{
		TPid:       "TEST_TPID",
		LcrRulesId: "TEST_LCR",
		LcrRules: []*TPLcrRule{
			&TPLcrRule{
				Direction:     "*in",
				Tenant:        "cgrates.org",
				Customer:      "*any",
				DestinationId: "EU_LANDLINE",
				Category:      "LCR_STANDARD",
				Strategy:      "*static",
				Suppliers:     "ivo;dan;rif",
				ActivatinTime: "2012-01-01T00:00:00Z",
				Weight:        20.0},
			//*in,cgrates.org,*any,*any,LCR_STANDARD,*lowest_cost,,2012-01-01T00:00:00Z,20
			&TPLcrRule{
				Direction:     "*in",
				Tenant:        "cgrates.org",
				Customer:      "*any",
				DestinationId: "*any",
				Category:      "LCR_STANDARD",
				Strategy:      "*lowest_cost",
				Suppliers:     "",
				ActivatinTime: "2012-01-01T00:00:00Z",
				Weight:        10.0},
		},
	}
	expectedSlc := [][]string{
		[]string{"TEST_LCR", "*in", "cgrates.org", "*any", "EU_LANDLINE", "LCR_STANDARD", "*static", "ivo;dan;rif", "2012-01-01T00:00:00Z", "20"},
		[]string{"TEST_LCR", "*in", "cgrates.org", "*any", "*any", "LCR_STANDARD", "*lowest_cost", "", "2012-01-01T00:00:00Z", "10"},
	}
	if slc := lcr.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

//CDRST1,5,60m,ASR,2014-07-29T15:00:00Z;2014-07-29T16:00:00Z,*voice,87.139.12.167,FS_JSON,rated,*out,cgrates.org,call,dan,dan,49,5m;10m,default,rif,rif,0;2,STANDARD_TRIGGERS
func TestTPCdrStatsAsExportSlice(t *testing.T) {
	cdrStats := &TPCdrStats{
		TPid:       "TEST_TPID",
		CdrStatsId: "CDRST1",
		CdrStats: []*TPCdrStat{
			&TPCdrStat{
				QueueLength:       "5",
				TimeWindow:        "60m",
				Metrics:           "ASR;ACD",
				SetupInterval:     "2014-07-29T15:00:00Z;2014-07-29T16:00:00Z",
				TOR:               "*voice",
				CdrHost:           "87.139.12.167",
				CdrSource:         "FS_JSON",
				ReqType:           "rated",
				Direction:         "*out",
				Tenant:            "cgrates.org",
				Category:          "call",
				Account:           "dan",
				Subject:           "dan",
				DestinationPrefix: "49",
				UsageInterval:     "5m;10m",
				MediationRunIds:   "default",
				RatedAccount:      "rif",
				RatedSubject:      "rif",
				CostInterval:      "0;2",
				ActionTriggers:    "STANDARD_TRIGGERS"},
			&TPCdrStat{
				QueueLength:       "5",
				TimeWindow:        "60m",
				Metrics:           "ASR",
				SetupInterval:     "2014-07-29T15:00:00Z;2014-07-29T16:00:00Z",
				TOR:               "*voice",
				CdrHost:           "87.139.12.167",
				CdrSource:         "FS_JSON",
				ReqType:           "rated",
				Direction:         "*out",
				Tenant:            "cgrates.org",
				Category:          "call",
				Account:           "dan",
				Subject:           "dan",
				DestinationPrefix: "49",
				UsageInterval:     "5m;10m",
				MediationRunIds:   "default",
				RatedAccount:      "dan",
				RatedSubject:      "dan",
				CostInterval:      "0;2",
				ActionTriggers:    "STANDARD_TRIGGERS"},
		},
	}
	expectedSlc := [][]string{
		[]string{"CDRST1", "5", "60m", "ASR;ACD", "2014-07-29T15:00:00Z;2014-07-29T16:00:00Z", "*voice", "87.139.12.167", "FS_JSON", "rated", "*out", "cgrates.org", "call", "dan", "dan", "49", "5m;10m",
			"default", "rif", "rif", "0;2", "STANDARD_TRIGGERS"},
		[]string{"CDRST1", "5", "60m", "ASR", "2014-07-29T15:00:00Z;2014-07-29T16:00:00Z", "*voice", "87.139.12.167", "FS_JSON", "rated", "*out", "cgrates.org", "call", "dan", "dan", "49", "5m;10m",
			"default", "dan", "dan", "0;2", "STANDARD_TRIGGERS"},
	}
	if slc := cdrStats.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

//#Direction,Tenant,Category,Account,Subject,RunId,RunFilter,ReqTypeField,DirectionField,TenantField,CategoryField,AccountField,SubjectField,DestinationField,SetupTimeField,AnswerTimeField,UsageField
//*out,cgrates.org,call,1001,1001,derived_run1,,^rated,*default,*default,*default,*default,^1002,*default,*default,*default,*default
func TestTPDerivedChargersAsExportSlice(t *testing.T) {
	dcs := TPDerivedChargers{
		TPid:      "TEST_TPID",
		Loadid:    "TEST_LOADID",
		Direction: "*out",
		Tenant:    "cgrates.org",
		Category:  "call",
		Account:   "1001",
		Subject:   "1001",
		DerivedChargers: []*TPDerivedCharger{
			&TPDerivedCharger{
				RunId:            "derived_run1",
				RunFilters:       "",
				ReqTypeField:     "^rated",
				DirectionField:   "*default",
				TenantField:      "*default",
				CategoryField:    "*default",
				AccountField:     "*default",
				SubjectField:     "^1002",
				DestinationField: "*default",
				SetupTimeField:   "*default",
				AnswerTimeField:  "*default",
				UsageField:       "*default",
			},
			&TPDerivedCharger{
				RunId:            "derived_run2",
				RunFilters:       "",
				ReqTypeField:     "^rated",
				DirectionField:   "*default",
				TenantField:      "*default",
				CategoryField:    "*default",
				AccountField:     "^1002",
				SubjectField:     "*default",
				DestinationField: "*default",
				SetupTimeField:   "*default",
				AnswerTimeField:  "*default",
				UsageField:       "*default",
			},
		},
	}
	expectedSlc := [][]string{
		[]string{"*out", "cgrates.org", "call", "1001", "1001",
			"derived_run1", "", "^rated", "*default", "*default", "*default", "*default", "^1002", "*default", "*default", "*default", "*default"},
		[]string{"*out", "cgrates.org", "call", "1001", "1001",
			"derived_run2", "", "^rated", "*default", "*default", "*default", "^1002", "*default", "*default", "*default", "*default", "*default"},
	}
	if slc := dcs.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionTriggersAsExportSlice(t *testing.T) {
	ap := &TPActionPlan{
		TPid: "TEST_TPID",
		Id:   "PACKAGE_10",
		ActionPlan: []*TPActionTiming{
			&TPActionTiming{
				ActionsId: "TOPUP_RST_10",
				TimingId:  "ASAP",
				Weight:    10.0},
			&TPActionTiming{
				ActionsId: "TOPUP_RST_5",
				TimingId:  "ASAP",
				Weight:    20.0},
		},
	}
	expectedSlc := [][]string{
		[]string{"PACKAGE_10", "TOPUP_RST_10", "ASAP", "10"},
		[]string{"PACKAGE_10", "TOPUP_RST_5", "ASAP", "20"},
	}
	if slc := ap.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPActionPlanAsExportSlice(t *testing.T) {
	at := &TPActionTriggers{
		TPid:             "TEST_TPID",
		ActionTriggersId: "STANDARD_TRIGGERS",
		ActionTriggers: []*TPActionTrigger{
			&TPActionTrigger{
				BalanceType:           "*monetary",
				Direction:             "*out",
				ThresholdType:         "*min_balance",
				ThresholdValue:        2.0,
				Recurrent:             false,
				MinSleep:              time.Duration(0),
				DestinationId:         "",
				BalanceWeight:         0.0,
				BalanceExpirationDate: "*never",
				BalanceRatingSubject:  "special1",
				BalanceCategory:       "call",
				BalanceSharedGroup:    "SHARED_1",
				MinQueuedItems:        0,
				ActionsId:             "LOG_WARNING",
				Weight:                10},
			&TPActionTrigger{
				BalanceType:           "*monetary",
				Direction:             "*out",
				ThresholdType:         "*max_counter",
				ThresholdValue:        5.0,
				Recurrent:             false,
				MinSleep:              time.Duration(0),
				DestinationId:         "FS_USERS",
				BalanceWeight:         0.0,
				BalanceExpirationDate: "*never",
				BalanceRatingSubject:  "special1",
				BalanceCategory:       "call",
				BalanceSharedGroup:    "SHARED_1",
				MinQueuedItems:        0,
				ActionsId:             "LOG_WARNING",
				Weight:                10},
		},
	}
	expectedSlc := [][]string{
		[]string{"STANDARD_TRIGGERS", "*min_balance", "2", "false", "0", "*monetary", "*out", "call", "", "special1", "SHARED_1", "*never", "0", "0", "LOG_WARNING", "10"},
		[]string{"STANDARD_TRIGGERS", "*max_counter", "5", "false", "0", "*monetary", "*out", "call", "FS_USERS", "special1", "SHARED_1", "*never", "0", "0", "LOG_WARNING", "10"},
	}
	if slc := at.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}

func TestTPAccountActionsAsExportSlice(t *testing.T) {
	aa := &TPAccountActions{
		TPid:             "TEST_TPID",
		LoadId:           "TEST_LOADID",
		Tenant:           "cgrates.org",
		Account:          "1001",
		Direction:        "*out",
		ActionPlanId:     "PACKAGE_10_SHARED_A_5",
		ActionTriggersId: "STANDARD_TRIGGERS",
	}
	expectedSlc := [][]string{
		[]string{"cgrates.org", "1001", "*out", "PACKAGE_10_SHARED_A_5", "STANDARD_TRIGGERS"},
	}
	if slc := aa.AsExportSlice(); !reflect.DeepEqual(expectedSlc, slc) {
		t.Errorf("Expecting: %+v, received: %+v", expectedSlc, slc)
	}
}
