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
package engine

import (
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var rsponder *Responder

func init() {
	cfg, _ := config.NewDefaultCGRConfig()
	config.SetCgrConfig(cfg)
}

// Test internal abilites of GetDerivedChargers
func TestResponderGetDerivedChargers(t *testing.T) {

	cfgedDC := utils.DerivedChargers{&utils.DerivedCharger{RunId: "responder1", ReqTypeField: "test", DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}
	rsponder = &Responder{}
	attrs := utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "responder_test", Subject: "responder_test"}
	if err := accountingStorage.SetDerivedChargers(utils.DerivedChargersKey(utils.OUT, utils.ANY, utils.ANY, utils.ANY, utils.ANY), cfgedDC); err != nil {
		t.Error(err)
	}
	if err := accountingStorage.CacheAccounting(nil, []string{}, []string{}, []string{}); err != nil {
		t.Error(err)
	}
	var dcs utils.DerivedChargers
	if err := rsponder.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		t.Errorf("Expecting: %v, received: %v ", cfgedDC, dcs)
	}
}

func TestGetDerivedMaxSessionTime(t *testing.T) {
	testTenant := "vdf"
	cdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_RATED, Direction: "*out", Tenant: testTenant, Category: "call", Account: "dan", Subject: "dan",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		MediationRunId: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan"}
	var maxSessionTime float64
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != -1 {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
	deTMobile := &Destination{Id: "DE_TMOBILE", Prefixes: []string{"+49151", "+49160", "+49170", "+49171", "+49175"}}
	if err := dataStorage.SetDestination(deTMobile); err != nil {
		t.Error(err)
	}
	b10 := &Balance{Value: 10, Weight: 10, DestinationId: "DE_TMOBILE"}
	b20 := &Balance{Value: 20, Weight: 10, DestinationId: "DE_TMOBILE"}
	rifsAccount := &Account{Id: utils.ConcatenatedKey(utils.OUT, testTenant, "rif"), BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b10}}}
	dansAccount := &Account{Id: utils.ConcatenatedKey(utils.OUT, testTenant, "dan"), BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{b20}}}
	if err := accountingStorage.SetAccount(rifsAccount); err != nil {
		t.Error(err)
	}
	if err := accountingStorage.SetAccount(dansAccount); err != nil {
		t.Error(err)
	}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan", "dan")
	charger1 := utils.DerivedChargers{
		&utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^" + utils.META_PREPAID, DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^dan", SubjectField: "^dan", DestinationField: "^+49151708707", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra2", ReqTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^ivo", SubjectField: "^ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		&utils.DerivedCharger{RunId: "extra3", ReqTypeField: "^" + utils.META_PSEUDOPREPAID, DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^rif", SubjectField: "^rif", DestinationField: "^+49151708707", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}
	if err := accountingStorage.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	dataStorage.CacheRating(nil, nil, nil, nil, nil)
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
	if rifStoredAcnt, err := accountingStorage.GetAccount(utils.ConcatenatedKey(utils.OUT, testTenant, "rif")); err != nil {
		t.Error(err)
		//} else if rifStoredAcnt.BalanceMap[MINUTES+OUTBOUND].Equal(rifsAccount.BalanceMap[MINUTES+OUTBOUND]) {
		//	t.Errorf("Expected: %+v, received: %+v", rifsAccount.BalanceMap[MINUTES+OUTBOUND][0], rifStoredAcnt.BalanceMap[MINUTES+OUTBOUND][0])
	} else if rifStoredAcnt.BalanceMap[MINUTES+OUTBOUND][0].Value != rifsAccount.BalanceMap[MINUTES+OUTBOUND][0].Value {
		t.Error("BalanceValue: ", rifStoredAcnt.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	if danStoredAcnt, err := accountingStorage.GetAccount(utils.ConcatenatedKey(utils.OUT, testTenant, "dan")); err != nil {
		t.Error(err)
	} else if danStoredAcnt.BalanceMap[MINUTES+OUTBOUND][0].Value != dansAccount.BalanceMap[MINUTES+OUTBOUND][0].Value {
		t.Error("BalanceValue: ", danStoredAcnt.BalanceMap[MINUTES+OUTBOUND][0].Value)
	}
	var dcs utils.DerivedChargers
	attrs := utils.AttrDerivedChargers{Tenant: testTenant, Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	if err := rsponder.GetDerivedChargers(attrs, &dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, charger1) {
		t.Errorf("Expecting: %+v, received: %+v ", charger1, dcs)
	}
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != 1e+10 { // Smallest one, 10 seconds
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestGetSessionRuns(t *testing.T) {
	testTenant := "vdf"
	cdr := StoredCdr{CgrId: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()), OrderId: 123, TOR: utils.VOICE, AccId: "dsafdsaf",
		CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.META_PREPAID, Direction: "*out", Tenant: testTenant, Category: "call", Account: "dan2", Subject: "dan2",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), Supplier: "suppl1",
		MediationRunId: utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost: 1.01, RatedAccount: "dan", RatedSubject: "dan"}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan2", "dan2")
	dfDC := &utils.DerivedCharger{RunId: utils.DEFAULT_RUNID, ReqTypeField: utils.META_DEFAULT, DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: utils.META_DEFAULT, AccountField: utils.META_DEFAULT, SubjectField: utils.META_DEFAULT, DestinationField: utils.META_DEFAULT,
		SetupTimeField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT, SupplierField: utils.META_DEFAULT}
	extra1DC := &utils.DerivedCharger{RunId: "extra1", ReqTypeField: "^" + utils.META_PREPAID, DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: "^0", AccountField: "^minitsboy", SubjectField: "^rif", DestinationField: "^0256",
		SetupTimeField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT, SupplierField: utils.META_DEFAULT}
	extra2DC := &utils.DerivedCharger{RunId: "extra2", ReqTypeField: utils.META_DEFAULT, DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: utils.META_DEFAULT, AccountField: "^ivo", SubjectField: "^ivo", DestinationField: utils.META_DEFAULT,
		SetupTimeField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT, SupplierField: utils.META_DEFAULT}
	extra3DC := &utils.DerivedCharger{RunId: "extra3", ReqTypeField: "^" + utils.META_PSEUDOPREPAID, DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: "^0", AccountField: "^minu", SubjectField: "^rif", DestinationField: "^0256",
		SetupTimeField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT, SupplierField: utils.META_DEFAULT}
	charger1 := utils.DerivedChargers{extra1DC, extra2DC, extra3DC}
	if err := accountingStorage.SetDerivedChargers(keyCharger1, charger1); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	accountingStorage.CacheAccounting(nil, nil, nil, nil)
	sesRuns := make([]*SessionRun, 0)
	eSRuns := []*SessionRun{
		&SessionRun{DerivedCharger: extra1DC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "0", Tenant: "vdf", Subject: "rif", Account: "minitsboy", Destination: "0256", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}},
		&SessionRun{DerivedCharger: extra2DC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "call", Tenant: "vdf", Subject: "ivo", Account: "ivo", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}},
		&SessionRun{DerivedCharger: dfDC,
			CallDescriptor: &CallDescriptor{Direction: "*out", Category: "call", Tenant: "vdf", Subject: "dan2", Account: "dan2", Destination: "1002", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC)}}}
	if err := rsponder.GetSessionRuns(cdr, &sesRuns); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSRuns, sesRuns) {
		t.Errorf("Received: %+v", sesRuns)
	}
}

func TestGetLCR(t *testing.T) {
	rsponder.Stats = NewStats(dataStorage) // Load stats instance
	dstDe := &Destination{Id: "GERMANY", Prefixes: []string{"+49"}}
	if err := dataStorage.SetDestination(dstDe); err != nil {
		t.Error(err)
	}
	rp1 := &RatingPlan{
		Id: "RP1",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.01,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			dstDe.Id: []*RPRate{
				&RPRate{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	rp2 := &RatingPlan{
		Id: "RP2",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.02,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	rp3 := &RatingPlan{
		Id: "RP3",
		Timings: map[string]*RITiming{
			"30eab300": &RITiming{
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": &RIRate{
				ConnectFee: 0,
				Rates: []*Rate{
					&Rate{
						GroupIntervalStart: 0,
						Value:              0.03,
						RateIncrement:      time.Second,
						RateUnit:           time.Second,
					},
				},
				RoundingMethod:   utils.ROUNDING_MIDDLE,
				RoundingDecimals: 4,
			},
		},
		DestinationRates: map[string]RPRateList{
			"GERMANY": []*RPRate{
				&RPRate{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	for _, rpf := range []*RatingPlan{rp1, rp2, rp3} {
		if err := dataStorage.SetRatingPlan(rpf); err != nil {
			t.Error(err)
		}
	}
	danStatsId := "dan12_stats"
	rsponder.Stats.AddQueue(&CdrStats{Id: danStatsId, Supplier: []string{"dan12"}, Metrics: []string{ASR, ACD, ACC, TCC}}, nil)
	danRpfl := &RatingProfile{Id: "*out:tenant12:call:dan12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp1.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{danStatsId},
		}},
	}
	rifStatsId := "rif12_stats"
	rsponder.Stats.AddQueue(&CdrStats{Id: rifStatsId, Supplier: []string{"rif12"}, Metrics: []string{ASR, ACD, ACC, TCC}}, nil)
	rifRpfl := &RatingProfile{Id: "*out:tenant12:call:rif12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp2.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{rifStatsId},
		}},
	}
	ivoStatsId := "ivo12_stats"
	rsponder.Stats.AddQueue(&CdrStats{Id: ivoStatsId, Supplier: []string{"ivo12"}, Metrics: []string{ASR, ACD, ACC, TCC}}, nil)
	ivoRpfl := &RatingProfile{Id: "*out:tenant12:call:ivo12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp3.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{ivoStatsId},
		}},
	}
	for _, rpfl := range []*RatingProfile{danRpfl, rifRpfl, ivoRpfl} {
		if err := dataStorage.SetRatingProfile(rpfl); err != nil {
			t.Error(err)
		}
	}
	lcrStatic := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_static", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0}},
			},
		},
	}
	lcrLowestCost := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_least_cost", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0}},
			},
		},
	}
	lcrQosThreshold := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_qos_threshold", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;4m;;;;;", Weight: 10.0}},
			},
		},
	}
	lcrQos := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_qos", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			&LCRActivation{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					&LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS, Weight: 10.0}},
			},
		},
	}
	for _, lcr := range []*LCR{lcrStatic, lcrLowestCost, lcrQosThreshold, lcrQos} {
		if err := dataStorage.SetLCR(lcr); err != nil {
			t.Error(err)
		}
	}
	if err := dataStorage.CacheRating([]string{DESTINATION_PREFIX + dstDe.Id},
		[]string{RATING_PLAN_PREFIX + rp1.Id, RATING_PLAN_PREFIX + rp2.Id},
		[]string{RATING_PROFILE_PREFIX + danRpfl.Id, RATING_PROFILE_PREFIX + rifRpfl.Id},
		[]string{},
		[]string{LCR_PREFIX + lcrStatic.GetId(), LCR_PREFIX + lcrLowestCost.GetId()}); err != nil {
		t.Error(err)
	}
	cdStatic := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "tenant12",
		Direction:   utils.OUT,
		Category:    "call_static",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eStLcr := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	var lcr LCRCost
	if err := rsponder.GetLCR(cdStatic, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.SupplierCosts[2], lcr.SupplierCosts[2])
	}
	// Test *least_cost strategy here
	cdLowestCost := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "tenant12",
		Direction:   utils.OUT,
		Category:    "call_least_cost",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eLcLcr := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
		},
	}
	var lcrLc LCRCost
	if err := rsponder.GetLCR(cdLowestCost, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.SupplierCosts[1], lcrLc.SupplierCosts[1])
	}
	bRif12 := &Balance{Value: 40, Weight: 10, DestinationId: dstDe.Id}
	bIvo12 := &Balance{Value: 60, Weight: 10, DestinationId: dstDe.Id}
	rif12sAccount := &Account{Id: utils.ConcatenatedKey(utils.OUT, "tenant12", "rif12"), BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{bRif12}}, AllowNegative: true}
	ivo12sAccount := &Account{Id: utils.ConcatenatedKey(utils.OUT, "tenant12", "ivo12"), BalanceMap: map[string]BalanceChain{MINUTES + OUTBOUND: BalanceChain{bIvo12}}, AllowNegative: true}
	for _, acnt := range []*Account{rif12sAccount, ivo12sAccount} {
		if err := accountingStorage.SetAccount(acnt); err != nil {
			t.Error(err)
		}
	}
	eLcLcr = &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 0.4, Duration: 60 * time.Second},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
		},
	}
	if err := rsponder.GetLCR(cdLowestCost, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.SupplierCosts, lcrLc.SupplierCosts)
	}

	// Test *qos_threshold strategy here
	cdQosThreshold := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "tenant12",
		Direction:   utils.OUT,
		Category:    "call_qos_threshold",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eQTLcr := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;4m;;;;;", Weight: 10.0},
	}
	var lcrQT LCRCost
	if err := rsponder.GetLCR(cdQosThreshold, &lcrQT); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)

	} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.SupplierCosts, lcrQT.SupplierCosts)
	}
	cdr := &StoredCdr{Supplier: "rif12", AnswerTime: time.Now(), Usage: 3 * time.Minute, Cost: 1}
	rsponder.Stats.AppendCDR(cdr, nil)
	cdr = &StoredCdr{Supplier: "dan12", AnswerTime: time.Now(), Usage: 5 * time.Minute, Cost: 2}
	rsponder.Stats.AppendCDR(cdr, nil)
	eQTLcr = &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;4m;;;;;", Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second,
				QOS: map[string]float64{ACD: 300, ASR: 100, ACC: 2, TCC: 2}, qosSortParams: []string{"35", "4m"}},
		},
	}
	if err := rsponder.GetLCR(cdQosThreshold, &lcrQT); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)

	} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eQTLcr.SupplierCosts[0], lcrQT.SupplierCosts[0])
	}

	// Test *qos strategy here
	cdQos := &CallDescriptor{
		TimeStart:   time.Date(2015, 04, 06, 17, 40, 0, 0, time.UTC),
		TimeEnd:     time.Date(2015, 04, 06, 17, 41, 0, 0, time.UTC),
		Tenant:      "tenant12",
		Direction:   utils.OUT,
		Category:    "call_qos",
		Destination: "+4986517174963",
		Account:     "dan",
		Subject:     "dan",
	}
	eQosLcr := &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second, QOS: map[string]float64{ACD: 300, ASR: 100, ACC: 2, TCC: 2}, qosSortParams: []string{ASR, ACD, ACC, TCC}},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 0.4, Duration: 60 * time.Second, QOS: map[string]float64{ACD: 180, ASR: 100, ACC: 1, TCC: 1}, qosSortParams: []string{ASR, ACD, ACC, TCC}},
			&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second, QOS: map[string]float64{ACD: 0, ASR: 0, ACC: 0, TCC: 0}, qosSortParams: []string{ASR, ACD, ACC, TCC}},
		},
	}
	var lcrQ LCRCost
	if err := rsponder.GetLCR(cdQos, &lcrQ); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eQosLcr.Entry, lcrQ.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eQosLcr.Entry, lcrQ.Entry)

	} else if !reflect.DeepEqual(eQosLcr.SupplierCosts, lcrQ.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eQosLcr.SupplierCosts[1], lcrQ.SupplierCosts[1])
		for _, sc := range lcrQ.SupplierCosts {
			log.Printf("%+v", sc)
		}
	}
}
