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
	"bytes"
	"encoding/gob"
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
	rsponder = &Responder{MaxComputedUsage: cfg.RalsCfg().RALsMaxComputedUsage}
}

// Test internal abilites of GetDerivedChargers
func TestResponderGetDerivedChargers(t *testing.T) {
	cfgedDC := &utils.DerivedChargers{DestinationIDs: utils.StringMap{}, Chargers: []*utils.DerivedCharger{{RunID: "responder1",
		RequestTypeField: utils.META_DEFAULT, DirectionField: "test", TenantField: "test",
		CategoryField: "test", AccountField: "test", SubjectField: "test", DestinationField: "test", SetupTimeField: "test", AnswerTimeField: "test", UsageField: "test"}}}
	attrs := &utils.AttrDerivedChargers{Tenant: "cgrates.org", Category: "call", Direction: "*out", Account: "responder_test", Subject: "responder_test"}
	if err := dm.DataDB().SetDerivedChargers(utils.DerivedChargersKey(utils.OUT, utils.ANY, utils.ANY, utils.ANY, utils.ANY), cfgedDC, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	dcs := &utils.DerivedChargers{}
	if err := rsponder.GetDerivedChargers(attrs, dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs, cfgedDC) {
		t.Errorf("Expecting: %v, received: %v ", cfgedDC, dcs)
	}
}

func TestResponderGetDerivedMaxSessionTime(t *testing.T) {
	testTenant := "vdf"
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf", OriginHost: "192.168.1.1", Source: "test",
		RequestType: utils.META_RATED, Tenant: testTenant, Category: "call", Account: "dan", Subject: "dan",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01}
	var maxSessionTime time.Duration
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != time.Duration(-1) {
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
	deTMobile := &Destination{Id: "DE_TMOBILE",
		Prefixes: []string{"+49151", "+49160", "+49170", "+49171", "+49175"}}
	if err := dm.DataDB().SetDestination(deTMobile, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.DataDB().SetReverseDestination(deTMobile, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	b10 := &Balance{Value: 10 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.NewStringMap("DE_TMOBILE")}
	b20 := &Balance{Value: 20 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.NewStringMap("DE_TMOBILE")}
	rifsAccount := &Account{ID: utils.ConcatenatedKey(testTenant, "rif"),
		BalanceMap: map[string]Balances{
			utils.VOICE: {b10}}}
	dansAccount := &Account{ID: utils.ConcatenatedKey(testTenant, "dan"),
		BalanceMap: map[string]Balances{utils.VOICE: {b20}}}
	if err := dm.DataDB().SetAccount(rifsAccount); err != nil {
		t.Error(err)
	}
	if err := dm.DataDB().SetAccount(dansAccount); err != nil {
		t.Error(err)
	}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan", "dan")
	charger1 := &utils.DerivedChargers{Chargers: []*utils.DerivedCharger{
		{RunID: "extra1", RequestTypeField: "^" + utils.META_PREPAID, DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^dan", SubjectField: "^dan", DestinationField: "^+49151708707", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		{RunID: "extra2", RequestTypeField: "*default", DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^ivo", SubjectField: "^ivo", DestinationField: "*default", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
		{RunID: "extra3", RequestTypeField: "^" + utils.META_PSEUDOPREPAID, DirectionField: "*default", TenantField: "*default", CategoryField: "*default",
			AccountField: "^rif", SubjectField: "^rif", DestinationField: "^+49151708707", SetupTimeField: "*default", AnswerTimeField: "*default", UsageField: "*default"},
	}}
	if err := dm.DataDB().SetDerivedChargers(keyCharger1, charger1, utils.NonTransactional); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	if rifStoredAcnt, err := dm.DataDB().GetAccount(utils.ConcatenatedKey(testTenant, "rif")); err != nil {
		t.Error(err)
		//} else if rifStoredAcnt.BalanceMap[utils.VOICE].Equal(rifsAccount.BalanceMap[utils.VOICE]) {
		//	t.Errorf("Expected: %+v, received: %+v", rifsAccount.BalanceMap[utils.VOICE][0], rifStoredAcnt.BalanceMap[utils.VOICE][0])
	} else if rifStoredAcnt.BalanceMap[utils.VOICE][0].GetValue() != rifsAccount.BalanceMap[utils.VOICE][0].GetValue() {
		t.Error("BalanceValue: ", rifStoredAcnt.BalanceMap[utils.VOICE][0].GetValue())
	}
	if danStoredAcnt, err := dm.DataDB().GetAccount(utils.ConcatenatedKey(testTenant, "dan")); err != nil {
		t.Error(err)
	} else if danStoredAcnt.BalanceMap[utils.VOICE][0].GetValue() != dansAccount.BalanceMap[utils.VOICE][0].GetValue() {
		t.Error("BalanceValue: ", danStoredAcnt.BalanceMap[utils.VOICE][0].GetValue())
	}
	dcs := &utils.DerivedChargers{}
	attrs := &utils.AttrDerivedChargers{Tenant: testTenant, Category: "call", Direction: "*out", Account: "dan", Subject: "dan"}
	if err := rsponder.GetDerivedChargers(attrs, dcs); err != nil {
		t.Error("Unexpected error", err.Error())
	} else if !reflect.DeepEqual(dcs.Chargers, charger1.Chargers) {
		t.Errorf("Expecting: %+v, received: %+v ", charger1, dcs)
	}
	if err := rsponder.GetDerivedMaxSessionTime(cdr, &maxSessionTime); err != nil {
		t.Error(err)
	} else if maxSessionTime != 1e+10 { // Smallest one, 10 seconds
		t.Error("Unexpected maxSessionTime received: ", maxSessionTime)
	}
}

func TestResponderGetSessionRuns(t *testing.T) {
	testTenant := "vdf"
	cdr := &CDR{CGRID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID: 123, ToR: utils.VOICE, OriginID: "dsafdsaf",
		OriginHost: "192.168.1.1", Source: "test", RequestType: utils.META_PREPAID,
		Tenant: testTenant, Category: "call", Account: "dan2", Subject: "dan2",
		Destination: "1002", SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:      utils.DEFAULT_RUNID, Usage: time.Duration(10) * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}, Cost: 1.01}
	keyCharger1 := utils.ConcatenatedKey("*out", testTenant, "call", "dan2", "dan2")
	dfDC := &utils.DerivedCharger{
		RunID: utils.DEFAULT_RUNID, RequestTypeField: utils.META_DEFAULT,
		DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: utils.META_DEFAULT, AccountField: utils.META_DEFAULT,
		SubjectField: utils.META_DEFAULT, DestinationField: utils.META_DEFAULT,
		SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT,
		AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
		SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT,
		CostField: utils.META_DEFAULT, PreRatedField: utils.META_DEFAULT}
	extra1DC := &utils.DerivedCharger{
		RunID: "extra1", RequestTypeField: "^" + utils.META_PREPAID,
		DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: "^0", AccountField: "^minitsboy", SubjectField: "^rif",
		DestinationField: "^0256", SetupTimeField: utils.META_DEFAULT,
		PDDField: utils.META_DEFAULT, AnswerTimeField: utils.META_DEFAULT,
		UsageField: utils.META_DEFAULT, SupplierField: utils.META_DEFAULT}
	extra2DC := &utils.DerivedCharger{
		RunID: "extra2", RequestTypeField: utils.META_DEFAULT,
		DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: utils.META_DEFAULT, AccountField: "^ivo", SubjectField: "^ivo",
		DestinationField: utils.META_DEFAULT, SetupTimeField: utils.META_DEFAULT,
		AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
		SupplierField: utils.META_DEFAULT}
	extra3DC := &utils.DerivedCharger{
		RunID: "extra3", RequestTypeField: "^" + utils.META_PSEUDOPREPAID,
		DirectionField: utils.META_DEFAULT, TenantField: utils.META_DEFAULT,
		CategoryField: "^0", AccountField: "^minu",
		SubjectField: "^rif", DestinationField: "^0256",
		SetupTimeField: utils.META_DEFAULT, PDDField: utils.META_DEFAULT,
		AnswerTimeField: utils.META_DEFAULT, UsageField: utils.META_DEFAULT,
		SupplierField: utils.META_DEFAULT, DisconnectCauseField: utils.META_DEFAULT}
	charger1 := &utils.DerivedChargers{Chargers: []*utils.DerivedCharger{extra1DC, extra2DC, extra3DC}}
	if err := dm.DataDB().SetDerivedChargers(keyCharger1, charger1,
		utils.NonTransactional); err != nil {
		t.Error("Error on setting DerivedChargers", err.Error())
	}
	sesRuns := make([]*SessionRun, 0)
	eSRuns := []*SessionRun{
		{RequestType: utils.META_PREPAID,
			DerivedCharger: extra1DC,
			CallDescriptor: &CallDescriptor{
				CgrID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
				RunID: "extra1", Direction: "*out", Category: "0",
				Tenant: "vdf", Subject: "rif", Account: "minitsboy",
				Destination: "0256", TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
				TimeEnd: time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE,
				ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}},
		{RequestType: utils.META_PREPAID,
			DerivedCharger: extra2DC,
			CallDescriptor: &CallDescriptor{
				CgrID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
				RunID: "extra2", Direction: "*out", Category: "call",
				Tenant: "vdf", Subject: "ivo", Account: "ivo", Destination: "1002",
				TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
				TimeEnd:   time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE,
				ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}},
		{RequestType: utils.META_PSEUDOPREPAID,
			DerivedCharger: extra3DC,
			CallDescriptor: &CallDescriptor{
				CgrID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
				RunID: "extra3", Direction: "*out", Category: "0",
				Tenant: "vdf", Subject: "rif", Account: "minu", Destination: "0256",
				TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
				TimeEnd:   time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE,
				ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}},
		{RequestType: utils.META_PREPAID,
			DerivedCharger: dfDC,
			CallDescriptor: &CallDescriptor{
				CgrID: utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
				RunID: "*default", Direction: "*out", Category: "call",
				Tenant: "vdf", Subject: "dan2", Account: "dan2", Destination: "1002",
				TimeStart: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
				TimeEnd:   time.Date(2013, 11, 7, 8, 42, 36, 0, time.UTC), TOR: utils.VOICE,
				ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"}}}}
	if err := rsponder.GetSessionRuns(cdr, &sesRuns); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSRuns, sesRuns) {
		t.Errorf("Expecting: %s, received: %s", utils.ToJSON(eSRuns), utils.ToJSON(sesRuns))
	}
}

func TestResponderGetLCR(t *testing.T) {
	rsponder.CdrStats = NewStats(dm, 0) // Load stats instance
	dstDe := &Destination{Id: "GERMANY", Prefixes: []string{"+49"}}
	if err := dm.DataDB().SetDestination(dstDe, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	if err := dm.DataDB().SetReverseDestination(dstDe, utils.NonTransactional); err != nil {
		t.Error(err)
	}
	rp1 := &RatingPlan{
		Id: "RP1",
		Timings: map[string]*RITiming{
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": {
				ConnectFee: 0,
				Rates: []*Rate{
					{
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
				{
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
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": {
				ConnectFee: 0,
				Rates: []*Rate{
					{
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
				{
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
			"30eab300": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
		Ratings: map[string]*RIRate{
			"b457f86d": {
				ConnectFee: 0,
				Rates: []*Rate{
					{
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
				{
					Timing: "30eab300",
					Rating: "b457f86d",
					Weight: 10,
				},
			},
		},
	}
	for _, rpf := range []*RatingPlan{rp1, rp2, rp3} {
		if err := dm.SetRatingPlan(rpf, utils.NonTransactional); err != nil {
			t.Error(err)
		}
	}
	danStatsId := "dan12_stats"
	var r int
	rsponder.CdrStats.Call("CDRStatsV1.AddQueue", &CdrStats{Id: danStatsId, Supplier: []string{"dan12"}, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}, &r)
	danRpfl := &RatingProfile{Id: "*out:tenant12:call:dan12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp1.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{danStatsId},
		}},
	}
	rifStatsId := "rif12_stats"
	rsponder.CdrStats.Call("CDRStatsV1.AddQueue", &CdrStats{Id: rifStatsId, Supplier: []string{"rif12"}, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}, &r)
	rifRpfl := &RatingProfile{Id: "*out:tenant12:call:rif12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp2.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{rifStatsId},
		}},
	}
	ivoStatsId := "ivo12_stats"
	rsponder.CdrStats.Call("CDRStatsV1.AddQueue", &CdrStats{Id: ivoStatsId, Supplier: []string{"ivo12"}, Metrics: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}}, &r)
	ivoRpfl := &RatingProfile{Id: "*out:tenant12:call:ivo12",
		RatingPlanActivations: RatingPlanActivations{&RatingPlanActivation{
			ActivationTime:  time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
			RatingPlanId:    rp3.Id,
			FallbackKeys:    []string{},
			CdrStatQueueIds: []string{ivoStatsId},
		}},
	}
	for _, rpfl := range []*RatingProfile{danRpfl, rifRpfl, ivoRpfl} {
		if err := dm.SetRatingProfile(rpfl, utils.NonTransactional); err != nil {
			t.Error(err)
		}
	}
	lcrStatic := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_static", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_STATIC, StrategyParams: "ivo12;dan12;rif12", Weight: 10.0}},
			},
		},
	}
	lcrLowestCost := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_least_cost", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0}},
			},
		},
	}
	lcrQosThreshold := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_qos_threshold", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0}},
			},
		},
	}
	lcrQos := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_qos", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_QOS, Weight: 10.0}},
			},
		},
	}
	lcrLoad := &LCR{Direction: utils.OUT, Tenant: "tenant12", Category: "call_load", Account: utils.ANY, Subject: utils.ANY,
		Activations: []*LCRActivation{
			{
				ActivationTime: time.Date(2015, 01, 01, 8, 0, 0, 0, time.UTC),
				Entries: []*LCREntry{
					{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOAD, StrategyParams: "ivo12:10;dan12:3", Weight: 10.0}},
			},
		},
	}
	for _, lcr := range []*LCR{lcrStatic, lcrLowestCost, lcrQosThreshold, lcrQos, lcrLoad} {
		if err := dm.SetLCR(lcr, utils.NonTransactional); err != nil {
			t.Error(err)
		}
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
			{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
		},
	}
	var lcr LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdStatic}, &lcr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eStLcr.Entry, lcr.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eStLcr.Entry, lcr.Entry)
	} else if !reflect.DeepEqual(eStLcr.SupplierCosts, lcr.SupplierCosts) {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", utils.ToJSON(eStLcr.SupplierCosts), utils.ToJSON(lcr.SupplierCosts))
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
			{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:rif12", Cost: 1.2, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:ivo12", Cost: 1.8, Duration: 60 * time.Second},
		},
	}
	var lcrLc LCRCost
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdLowestCost}, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.SupplierCosts, lcrLc.SupplierCosts)
	}
	bRif12 := &Balance{Value: 40 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.NewStringMap(dstDe.Id)}
	bIvo12 := &Balance{Value: 60 * float64(time.Second),
		Weight: 10, DestinationIDs: utils.NewStringMap(dstDe.Id)}
	rif12sAccount := &Account{
		ID: utils.ConcatenatedKey("tenant12", "rif12"),
		BalanceMap: map[string]Balances{
			utils.VOICE: {bRif12}}, AllowNegative: true}
	ivo12sAccount := &Account{
		ID: utils.ConcatenatedKey("tenant12", "ivo12"),
		BalanceMap: map[string]Balances{
			utils.VOICE: {bIvo12}}, AllowNegative: true}
	for _, acnt := range []*Account{rif12sAccount, ivo12sAccount} {
		if err := dm.DataDB().SetAccount(acnt); err != nil {
			t.Error(err)
		}
	}
	eLcLcr = &LCRCost{
		Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call", Strategy: LCR_STRATEGY_LOWEST, Weight: 10.0},
		SupplierCosts: []*LCRSupplierCost{
			{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:rif12", Cost: 0.4, Duration: 60 * time.Second},
			{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second},
		},
	}
	if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdLowestCost}, &lcrLc); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eLcLcr.Entry, lcrLc.Entry) {
		t.Errorf("Expecting: %+v, received: %+v", eLcLcr.Entry, lcrLc.Entry)

	} else if !reflect.DeepEqual(eLcLcr.SupplierCosts, lcrLc.SupplierCosts) {
		t.Errorf("Expecting: %s\n, received: %+v", utils.ToJSON(eLcLcr.SupplierCosts), utils.ToJSON(lcrLc.SupplierCosts))
	}
	/*
			// Test *qos_threshold strategy here,
			// FixMe with newly stats
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
				Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call",
					Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0},
				SupplierCosts: []*LCRSupplierCost{
					&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second,
						QOS: map[string]float64{TCD: -1, ACC: -1, TCC: -1, ASR: -1, ACD: -1, PDD: -1, DDC: -1}, qosSortParams: []string{"35", "4m"}},
					&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 0.4, Duration: 60 * time.Second,
						QOS: map[string]float64{TCD: -1, ACC: -1, TCC: -1, ASR: -1, ACD: -1, PDD: -1, DDC: -1}, qosSortParams: []string{"35", "4m"}},
					&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second,
						QOS: map[string]float64{TCD: -1, ACC: -1, TCC: -1, ASR: -1, ACD: -1, PDD: -1, DDC: -1}, qosSortParams: []string{"35", "4m"}},
				},
			}
			var lcrQT LCRCost
			if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQosThreshold}, &lcrQT); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
				t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)

			} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
				t.Errorf("Expecting: %+v, received: %+v", eQTLcr.SupplierCosts, lcrQT.SupplierCosts)
			}

			cdr := &CDR{AnswerTime: time.Now(), Usage: 3 * time.Minute, Cost: 1}
			rsponder.Stats.Call("CDRStatsV1.AppendCDR", cdr, &r)
			cdr = &CDR{AnswerTime: time.Now(), Usage: 5 * time.Minute, Cost: 2}
			rsponder.Stats.Call("CDRStatsV1.AppendCDR", cdr, &r)

			eQTLcr = &LCRCost{
				Entry: &LCREntry{DestinationId: utils.ANY, RPCategory: "call",
					Strategy: LCR_STRATEGY_QOS_THRESHOLD, StrategyParams: "35;;;;4m;;;;;;;;;", Weight: 10.0},
				SupplierCosts: []*LCRSupplierCost{
					&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second,
						QOS: map[string]float64{PDD: -1, TCD: -1, ACC: -1, TCC: -1, ASR: -1, ACD: -1, DDC: -1}, qosSortParams: []string{"35", "4m"}},
					&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second,
						QOS: map[string]float64{PDD: -1, ACD: 300, TCD: 300, ASR: 100, ACC: 2, TCC: 2, DDC: 2}, qosSortParams: []string{"35", "4m"}},
				},
			}
			if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQosThreshold}, &lcrQT); err != nil {
				t.Error(err)
			} else if !reflect.DeepEqual(eQTLcr.Entry, lcrQT.Entry) {
				t.Errorf("Expecting: %+v, received: %+v", eQTLcr.Entry, lcrQT.Entry)
			} else if !reflect.DeepEqual(eQTLcr.SupplierCosts, lcrQT.SupplierCosts) {
				t.Errorf("Expecting: %s, received: %s",
					utils.ToJSON(eQTLcr.SupplierCosts), utils.ToJSON(lcrQT.SupplierCosts))
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
				&LCRSupplierCost{Supplier: "*out:tenant12:call:ivo12", Cost: 0, Duration: 60 * time.Second,
					QOS: map[string]float64{ACD: -1, PDD: -1, TCD: -1, ASR: -1, ACC: -1, TCC: -1, DDC: -1}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
				&LCRSupplierCost{Supplier: "*out:tenant12:call:dan12", Cost: 0.6, Duration: 60 * time.Second,
					QOS: map[string]float64{ACD: 300, PDD: -1, TCD: 300, ASR: 100, ACC: 2, TCC: 2, DDC: 2}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
				&LCRSupplierCost{Supplier: "*out:tenant12:call:rif12", Cost: 0.4, Duration: 60 * time.Second,
					QOS: map[string]float64{ACD: 180, PDD: -1, TCD: 180, ASR: 100, ACC: 1, TCC: 1, DDC: 1}, qosSortParams: []string{ASR, PDD, ACD, TCD, ACC, TCC, DDC}},
			},
		}
		var lcrQ LCRCost
		if err := rsponder.GetLCR(&AttrGetLcr{CallDescriptor: cdQos}, &lcrQ); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(eQosLcr.Entry, lcrQ.Entry) {
			t.Errorf("Expecting: %+v, received: %+v", eQosLcr.Entry, lcrQ.Entry)

		} else if !reflect.DeepEqual(eQosLcr.SupplierCosts, lcrQ.SupplierCosts) {
			t.Errorf("Expecting: %+v, received: %+v", eQosLcr.SupplierCosts, lcrQ.SupplierCosts)
		}
	*/
}

func TestResponderGobSMCost(t *testing.T) {
	cc := &CallCost{
		Direction:   "*out",
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		TOR:         "*data",
		Cost:        0,
		Timespans: TimeSpans{&TimeSpan{
			TimeStart: time.Date(2016, 1, 5, 12, 30, 10, 0, time.UTC),
			TimeEnd:   time.Date(2016, 1, 5, 12, 55, 46, 0, time.UTC),
			Cost:      0,
			RateInterval: &RateInterval{
				Timing: nil,
				Rating: &RIRate{
					ConnectFee:       0,
					RoundingMethod:   "",
					RoundingDecimals: 0,
					MaxCost:          0,
					MaxCostStrategy:  "",
					Rates: RateGroups{&Rate{
						GroupIntervalStart: 0,
						Value:              0,
						RateIncrement:      1 * time.Second,
						RateUnit:           1 * time.Second,
					},
					},
				},
				Weight: 0,
			},
			DurationIndex: 0,
			Increments: Increments{&Increment{
				Duration: 1 * time.Second,
				Cost:     0,
				BalanceInfo: &DebitInfo{
					Unit: &UnitInfo{
						UUID:          "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
						ID:            "",
						Value:         100864,
						DestinationID: "data",
						Consumed:      1,
						TOR:           "*data",
						RateInterval:  nil,
					},
					Monetary:  nil,
					AccountID: "cgrates.org:1001",
				},
				CompressFactor: 1536,
			},
			},
			RoundIncrement: nil,
			MatchedSubject: "fa0aa280-2b76-4b5b-bb06-174f84b8c321",
			MatchedPrefix:  "data",
			MatchedDestId:  "*any",
			RatingPlanId:   "*none",
			CompressFactor: 1,
		},
		},
		RatedUsage: 1536,
	}
	attr := AttrCDRSStoreSMCost{
		Cost: &SMCost{
			CGRID:       "b783a8bcaa356570436983cd8a0e6de4993f9ba6",
			RunID:       utils.META_DEFAULT,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMR",
			Usage:       1536,
			CostDetails: NewEventCostFromCallCost(cc, "b783a8bcaa356570436983cd8a0e6de4993f9ba6", utils.META_DEFAULT),
		},
		CheckDuplicate: false,
	}

	var network bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&network) // Will write to network.
	dec := gob.NewDecoder(&network) // Will read from network.
	err := enc.Encode(attr)
	if err != nil {
		t.Error("encode error: ", err)
	}

	// Decode (receive) and print the values.
	var q AttrCDRSStoreSMCost
	err = dec.Decode(&q)
	if err != nil {
		t.Error("decode error: ", err)
	}
	if !reflect.DeepEqual(attr, q) {
		t.Error("wrong transmission")
	}
}
