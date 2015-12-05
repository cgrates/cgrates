/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mongoDb *MongoStorage

func TestMongoCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	var err error
	if mongoDb, err = NewMongoStorage("localhost", "27017", cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass); err != nil {
		t.Error("Error on opening database connection: ", err)
	}
}

func TestMongoSetGetTPTiming(t *testing.T) {
	if !*testLocal {
		return
	}
	tm := TpTiming{Tpid: utils.TEST_SQL, Tag: "ALWAYS", Time: "00:00:00"}
	if err := mongoDb.SetTpTimings([]TpTiming{tm}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.Tag); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(tm, tmgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tm, tmgs[0])
	}
	// Update
	tm.Time = "00:00:01"
	if err := mongoDb.SetTpTimings([]TpTiming{tm}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.Tag); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(tm, tmgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", tm, tmgs[0])
	}
}

func TestMongoSetGetTPDestination(t *testing.T) {
	if !*testLocal {
		return
	}
	dst := []TpDestination{
		TpDestination{Tpid: utils.TEST_SQL, Tag: utils.TEST_SQL, Prefix: "+49"},
		TpDestination{Tpid: utils.TEST_SQL, Tag: utils.TEST_SQL, Prefix: "+49151"},
		TpDestination{Tpid: utils.TEST_SQL, Tag: utils.TEST_SQL, Prefix: "+49176"},
	}
	if err := mongoDb.SetTpDestinations(dst); err != nil {
		t.Error(err.Error())
	}
	storData, err := mongoDb.GetTpDestinations(utils.TEST_SQL, utils.TEST_SQL)
	dsts, err := TpDestinations(storData).GetDestinations()
	expected := &Destination{Id: utils.TEST_SQL, Prefixes: []string{"+49", "+49151", "+49176"}}
	if err != nil {
		t.Error(err.Error())
	} else if !modelEqual(*expected, *dsts[utils.TEST_SQL]) {
		t.Errorf("Expecting: %+v, received: %+v", expected, dsts[utils.TEST_SQL])
	}
}

func TestMongoSetGetTPRates(t *testing.T) {
	if !*testLocal {
		return
	}
	RT_ID := "RT_1"
	rtSlots := []*utils.RateSlot{
		&utils.RateSlot{ConnectFee: 0.02, Rate: 0.01, RateUnit: "60s", RateIncrement: "60s", GroupIntervalStart: "0s"},
		&utils.RateSlot{ConnectFee: 0.00, Rate: 0.005, RateUnit: "60s", RateIncrement: "1s", GroupIntervalStart: "60s"},
	}
	for _, rs := range rtSlots {
		rs.SetDurations()
	}
	rates := &utils.TPRate{
		TPid:      utils.TEST_SQL,
		RateId:    RT_ID,
		RateSlots: rtSlots,
	}
	mRates := APItoModelRate(rates)
	if err := mongoDb.SetTpRates(mRates); err != nil {
		t.Error(err.Error())
	}
	if rts, err := mongoDb.GetTpRates(utils.TEST_SQL, RT_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mRates[0], rts[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mRates, rts)
	}
}

func TestMongoSetGetTPDestinationRates(t *testing.T) {
	if !*testLocal {
		return
	}
	DR_ID := "DR_1"
	dr := &utils.DestinationRate{DestinationId: "DST_1", RateId: "RT_1", RoundingMethod: "*up", RoundingDecimals: 4}

	eDrs := &utils.TPDestinationRate{TPid: utils.TEST_SQL, DestinationRateId: DR_ID, DestinationRates: []*utils.DestinationRate{dr}}
	mdrs := APItoModelDestinationRate(eDrs)
	if err := mongoDb.SetTpDestinationRates(mdrs); err != nil {
		t.Error(err.Error())
	}
	if drs, err := mongoDb.GetTpDestinationRates(utils.TEST_SQL, DR_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mdrs[0], drs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mdrs, drs)
	}
}

func TestMongoSetGetTPRatingPlans(t *testing.T) {
	if !*testLocal {
		return
	}
	RP_ID := "RP_1"
	rbBinding := &utils.TPRatingPlanBinding{DestinationRatesId: "DR_1", TimingId: "TM_1", Weight: 10.0}
	rp := &utils.TPRatingPlan{
		TPid:               utils.TEST_SQL,
		RatingPlanId:       RP_ID,
		RatingPlanBindings: []*utils.TPRatingPlanBinding{rbBinding},
	}
	mrp := APItoModelRatingPlan(rp)
	if err := mongoDb.SetTpRatingPlans(mrp); err != nil {
		t.Error(err.Error())
	}
	if drps, err := mongoDb.GetTpRatingPlans(utils.TEST_SQL, RP_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mrp[0], drps[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mrp, drps)
	}
}

func TestMongoSetGetTPRatingProfiles(t *testing.T) {
	if !*testLocal {
		return
	}
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RP_1"}}
	rp := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	mrp := APItoModelRatingProfile(rp)
	if err := mongoDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mongoDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mrp[0], rps[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mrp, rps)
	}

}

func TestMongoSetGetTPSharedGroups(t *testing.T) {
	if !*testLocal {
		return
	}
	SG_ID := "SG_1"
	tpSgs := &utils.TPSharedGroups{
		TPid:           utils.TEST_SQL,
		SharedGroupsId: SG_ID,
		SharedGroups: []*utils.TPSharedGroup{
			&utils.TPSharedGroup{Account: "dan", Strategy: "*lowest_first", RatingSubject: "lowest_rates"},
		},
	}
	mSgs := APItoModelSharedGroup(tpSgs)
	if err := mongoDb.SetTpSharedGroups(mSgs); err != nil {
		t.Error(err.Error())
	}
	if sgs, err := mongoDb.GetTpSharedGroups(utils.TEST_SQL, SG_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mSgs[0], sgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mSgs, sgs)
	}
}

func TestMongoSetGetTPcdrstats(t *testing.T) {
	if !*testLocal {
		return
	}
	CS_ID := "cdrsTATS_1"
	setCS := &utils.TPCdrStats{
		TPid:       utils.TEST_SQL,
		CdrStatsId: CS_ID,
		CdrStats: []*utils.TPCdrStat{
			&utils.TPCdrStat{QueueLength: "10", TimeWindow: "10m", Metrics: "ASR", Tenants: "cgrates.org", Categories: "call"},
		},
	}
	mcs := APItoModelCdrStat(setCS)
	if err := mongoDb.SetTpCdrStats(mcs); err != nil {
		t.Error(err.Error())
	}
	if cs, err := mongoDb.GetTpCdrStats(utils.TEST_SQL, CS_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mcs[0], cs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mcs, cs)
	}
}

func TestMongoSetGetTPDerivedChargers(t *testing.T) {
	if !*testLocal {
		return
	}
	dc := &utils.TPDerivedCharger{RunId: utils.DEFAULT_RUNID, ReqTypeField: "^" + utils.META_PREPAID, AccountField: "^rif", SubjectField: "^rif",
		UsageField: "cgr_duration", SupplierField: "^supplier1"}
	dcs := &utils.TPDerivedChargers{TPid: utils.TEST_SQL, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: []*utils.TPDerivedCharger{dc}}

	mdcs := APItoModelDerivedCharger(dcs)
	if err := mongoDb.SetTpDerivedChargers(mdcs); err != nil {
		t.Error(err.Error())
	}
	if rDCs, err := mongoDb.GetTpDerivedChargers(&mdcs[0]); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mdcs[0], rDCs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mdcs, rDCs)
	}
}

func TestMongoSetGetTPActions(t *testing.T) {
	if !*testLocal {
		return
	}
	ACTS_ID := "PREPAID_10"
	acts := []*utils.TPAction{
		&utils.TPAction{Identifier: "*topup_reset", BalanceType: "*monetary", Directions: "*out", Units: 10, ExpiryTime: "*unlimited",
			DestinationIds: "*any", BalanceWeight: 10, Weight: 10}}
	tpActions := &utils.TPActions{TPid: utils.TEST_SQL, ActionsId: ACTS_ID, Actions: acts}
	mas := APItoModelAction(tpActions)
	if err := mongoDb.SetTpActions(mas); err != nil {
		t.Error(err.Error())
	}
	if rTpActs, err := mongoDb.GetTpActions(utils.TEST_SQL, ACTS_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mas[0], rTpActs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mas, rTpActs)
	}
}

func TestMongoTPActionTimings(t *testing.T) {
	if !*testLocal {
		return
	}
	AP_ID := "AP_1"
	ap := &utils.TPActionPlan{
		TPid:         utils.TEST_SQL,
		ActionPlanId: AP_ID,
		ActionPlan:   []*utils.TPActionTiming{&utils.TPActionTiming{ActionsId: "ACTS_1", TimingId: "TM_1", Weight: 10.0}},
	}
	maps := APItoModelActionPlan(ap)
	if err := mongoDb.SetTpActionPlans(maps); err != nil {
		t.Error(err.Error())
	}
	if rAP, err := mongoDb.GetTpActionPlans(utils.TEST_SQL, AP_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(maps[0], rAP[0]) {
		t.Errorf("Expecting: %+v, received: %+v", maps, rAP)
	}
}

func TestMongoSetGetTPActionTriggers(t *testing.T) {
	if !*testLocal {
		return
	}
	atrg := &utils.TPActionTrigger{
		Id:                    "MY_FIRST_ATGR",
		BalanceType:           "*monetary",
		BalanceDirections:     "*out",
		ThresholdType:         "*min_balance",
		ThresholdValue:        2.0,
		Recurrent:             true,
		BalanceDestinationIds: "*any",
		Weight:                10.0,
		ActionsId:             "LOG_BALANCE",
	}
	atrgs := &utils.TPActionTriggers{
		TPid:             utils.TEST_SQL,
		ActionTriggersId: utils.TEST_SQL,
		ActionTriggers:   []*utils.TPActionTrigger{atrg},
	}
	matrg := APItoModelActionTrigger(atrgs)
	if err := mongoDb.SetTpActionTriggers(matrg); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if rcvMpAtrgs, err := mongoDb.GetTpActionTriggers(utils.TEST_SQL, utils.TEST_SQL); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !modelEqual(matrg[0], rcvMpAtrgs[0]) {
		t.Errorf("Expecting: %v, received: %v", matrg, rcvMpAtrgs)
	}
}

func TestMongoSetGetTpAccountActions(t *testing.T) {
	if !*testLocal {
		return
	}
	aa := &utils.TPAccountActions{TPid: utils.TEST_SQL, Tenant: "cgrates.org", Account: "1001", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	maa := APItoModelAccountAction(aa)
	if err := mongoDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mongoDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(*maa, aas[0]) {
		t.Errorf("Expecting: %+v, received: %+v", maa, aas)
	}
}

func TestMongoGetTPIds(t *testing.T) {
	if !*testLocal {
		return
	}
	eTPIds := []string{utils.TEST_SQL}
	if tpIds, err := mongoDb.GetTpIds(); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(eTPIds, tpIds) {
		t.Errorf("Expecting: %+v, received: %+v", eTPIds, tpIds)
	}
}

func TestMongoRemoveTPData(t *testing.T) {
	if !*testLocal {
		return
	}
	// Create Timings
	tm := &utils.ApierTPTiming{TPid: utils.TEST_SQL, TimingId: "ALWAYS", Time: "00:00:00"}
	tms := APItoModelTiming(tm)
	if err := mongoDb.SetTpTimings([]TpTiming{*tms}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Remove Timings
	if err := mongoDb.RemTpData(utils.TBL_TP_TIMINGS, utils.TEST_SQL, map[string]string{"tag": tm.TimingId}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tmgs) != 0 {
		t.Errorf("Timings should be empty, got instead: %+v", tmgs)
	}
	// Create RatingProfile
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1"}}
	rp := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	mrp := APItoModelRatingProfile(rp)
	if err := mongoDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mongoDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Remove RatingProfile
	if err := mongoDb.RemTpData(utils.TBL_TP_RATE_PROFILES, rp.TPid, map[string]string{"loadid": rp.LoadId, "direction": rp.Direction, "tenant": rp.Tenant, "category": rp.Category, "subject": rp.Subject}); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mongoDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err)
	} else if len(rps) != 0 {
		t.Errorf("RatingProfiles different than 0: %+v", rps)
	}
	// Create AccountActions
	aa := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Account: "1001", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	maa := APItoModelAccountAction(aa)
	if err := mongoDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mongoDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove AccountActions
	if err := mongoDb.RemTpData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.TPid, map[string]string{"loadid": aa.LoadId, "tenant": aa.Tenant, "account": aa.Account}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mongoDb.GetTpAccountActions(maa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
	// Create again so we can test complete TP removal
	if err := mongoDb.SetTpTimings([]TpTiming{*tms}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Create RatingProfile
	if err := mongoDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mongoDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Create AccountActions
	if err := mongoDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mongoDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove TariffPlan completely
	if err := mongoDb.RemTpData("", utils.TEST_SQL, nil); err != nil {
		t.Error(err.Error())
	}
	// Make sure we have removed it
	if tms, err := mongoDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tms) != 0 {
		t.Errorf("Non empty timings: %+v", tms)
	}
	if rpfs, err := mongoDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err)
	} else if len(rpfs) != 0 {
		t.Errorf("Non empty rpfs: %+v", rpfs)
	}
	if aas, err := mongoDb.GetTpAccountActions(maa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
}

func TestMongoSetCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa1", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:20Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "10s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", utils.CDRSOURCE: utils.TEST_SQL}

	cgrCdr2 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa2", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_PREPAID, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:22Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "20", utils.PDD: "7s", utils.SUPPLIER: "SUPPL1",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "Source": utils.TEST_SQL}

	cgrCdr3 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa3", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "premium_call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "1001", utils.SETUP_TIME: "2013-11-07T08:42:24Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "60s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "Source": utils.TEST_SQL}

	cgrCdr4 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa4", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_PSEUDOPREPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "+4986517174964", utils.SETUP_TIME: "2013-11-07T08:42:21Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "1m2s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "Source": utils.TEST_SQL}

	cgrCdr5 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa5", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_POSTPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "+4986517174963", utils.SETUP_TIME: "2013-11-07T08:42:25Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "15s", utils.PDD: "7s", utils.SUPPLIER: "SUPPL1",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "Source": utils.TEST_SQL}

	for _, cdr := range []*CgrCdr{cgrCdr1, cgrCdr2, cgrCdr3, cgrCdr4, cgrCdr5} {
		if err := mongoDb.SetCdr(cdr.AsStoredCdr("")); err != nil {
			t.Error(err.Error())
		}
	}
	strCdr1 := &CDR{TOR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "UNKNOWN", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CGRID = utils.Sha1(strCdr1.OriginID, strCdr1.SetupTime.String())
	strCdr2 := &CDR{TOR: utils.VOICE, OriginID: "bbb2", OriginHost: "192.168.1.2", Source: "UNKNOWN2", ReqType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, PDD: time.Duration(4) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CGRID = utils.Sha1(strCdr2.OriginID, strCdr2.SetupTime.String())
	strCdr3 := &CDR{TOR: utils.VOICE, OriginID: "bbb3", OriginHost: "192.168.1.1", Source: utils.TEST_SQL, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1000", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(2) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr3.CGRID = utils.Sha1(strCdr3.OriginID, strCdr3.SetupTime.String())

	for _, cdr := range []*CDR{strCdr1, strCdr2, strCdr3} {
		if err := mongoDb.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMongoSetRatedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr1 := &CDR{TOR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "UNKNOWN", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CGRID = utils.Sha1(strCdr1.OriginID, strCdr1.SetupTime.String())
	strCdr2 := &CDR{TOR: utils.VOICE, OriginID: "bbb2", OriginHost: "192.168.1.2", Source: "UNKNOWN", ReqType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, PDD: time.Duration(7) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CGRID = utils.Sha1(strCdr2.OriginID, strCdr2.SetupTime.String())
	strCdr3 := &CDR{TOR: utils.VOICE, OriginID: "bbb3", OriginHost: "192.168.1.1", Source: utils.TEST_SQL, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1002", Destination: "+4986517174964",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(2) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       "wholesale_run", Cost: 1.201}
	strCdr3.CGRID = utils.Sha1(strCdr3.OriginID, strCdr3.SetupTime.String())

	for _, cdr := range []*CDR{strCdr1, strCdr2, strCdr3} {
		if err := mongoDb.SetRatedCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMongoCallCost(t *testing.T) {
	if !*testLocal {
		return
	}
	CGRID := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	cc := &CallCost{
		Direction:   "*out",
		Category:    "call",
		Tenant:      "cgrates.org",
		Subject:     "91001",
		Account:     "8001",
		Destination: "1002",
		TOR:         utils.VOICE,
		Timespans: []*TimeSpan{
			&TimeSpan{
				TimeStart: time.Date(2013, 9, 10, 13, 40, 0, 0, time.UTC),
				TimeEnd:   time.Date(2013, 9, 10, 13, 41, 0, 0, time.UTC),
			},
			&TimeSpan{
				TimeStart: time.Date(2013, 9, 10, 13, 41, 0, 0, time.UTC),
				TimeEnd:   time.Date(2013, 9, 10, 13, 41, 30, 0, time.UTC),
			},
		},
	}
	if err := mongoDb.LogCallCost(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mongoDb.GetCallCostLog(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if cc.Cost != ccRcv.Cost {
		t.Errorf("Expecting call cost:\n%+v,\nreceived:\n%+v", cc.Timespans[0], ccRcv.Timespans[0])
	}
	// UPDATE test here
	cc.Category = "premium_call"
	if err := mongoDb.LogCallCost(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mongoDb.GetCallCostLog(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if cc.Cost != ccRcv.Cost {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
}

func TestMongoGetcdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var timeStart, timeEnd time.Time
	// All cdrs, no filter
	if cdrs, _, err := mongoDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 20 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Count ALL
	if cdrs, count, err := mongoDb.GetCDRs(&utils.CDRsFilter{Count: true}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 0 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	} else if count != 20 {
		t.Error("Unexpected count of cdrs returned: ", count)
	}
	// Limit 5
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 5 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Offset 5
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 5 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Offset with limit 2
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on CGRIDs
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Count on CGRIDS
	if _, count, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 3 {
		t.Error("Unexpected count of cdrs returned: ", count)
	}
	// Filter on CGRIDs plus reqType
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Count on multiple filter
	if _, count, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 1 {
		t.Error("Unexpected count of cdrs returned: ", count)
	}
	// Filter on runId
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{RunIDs: []string{utils.DEFAULT_RUNID}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 14 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on TOR
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{TORs: []string{utils.SMS}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 0 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple TOR
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{TORs: []string{utils.SMS, utils.VOICE}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 15 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on OriginHost
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 3 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple OriginHost
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"192.168.1.1", "192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 15 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on Source
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Sources: []string{"UNKNOWN"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple Source
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Sources: []string{"UNKNOWN", "UNKNOWN2"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on reqType
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 5 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple reqType
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 6 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on direction
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Directions: []string{"*out"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 15 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on tenant
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 4 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple tenants
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 15 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on category
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple categories
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call", "call"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 15 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on account
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 6 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple account
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1001", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 13 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on subject
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple subject
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 6 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on destPrefix
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 4 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on multiple DestinationPrefixes
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1001", "+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 5 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on ignoreRated
	var OrderIDStart, OrderIDEnd int64 // Capture also OrderIDs for the next test
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{MaxCost: utils.Float64Pointer(0.0)}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 7 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	} else {
		for _, cdr := range cdrs {
			if cdr.OrderID < OrderIDStart {
				OrderIDStart = cdr.OrderID
			}
			if cdr.OrderID > OrderIDEnd {
				OrderIDEnd = cdr.OrderID
			}
		}
	}
	// Filter on OrderIDStart
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{OrderIDStart: OrderIDStart}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 20 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on OrderIDStart and OrderIDEnd
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{OrderIDStart: OrderIDStart, OrderIDEnd: OrderIDEnd}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 20 { // TODO: find mongo equivalent
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on timeStart
	timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 6 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on minPDD
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{MinPDD: utils.Float64Pointer(float64(3 * time.Second))}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 7 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on maxPDD
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{MaxPDD: utils.Float64Pointer(float64(3 * time.Second))}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 13 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on minPDD, maxPDD
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{MinPDD: utils.Float64Pointer(float64(3 * time.Second)), MaxPDD: utils.Float64Pointer(float64(5 * time.Second))}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 4 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Combined filter
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{ReqTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	// Filter on ignoreDerived
	if cdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd, FilterOnRated: true}); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 2 { // ToDo: Recheck this value
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
}

func TestMongoRemCDRs(t *testing.T) {
	if !*testLocal {
		return
	}
	CGRIDB1 := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := mongoDb.RemCDRs([]string{CGRIDB1}); err != nil {
		t.Error(err.Error())
	}
	if cdrs, _, err := mongoDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 20 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
	tm, _ := utils.ParseTimeDetectLayout("2013-11-08T08:42:20Z", "")
	CGRIDA1 := utils.Sha1("aaa1", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-08T08:42:22Z", "")
	CGRIDA2 := utils.Sha1("aaa2", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:24Z", "")
	CGRIDA3 := utils.Sha1("aaa3", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:21Z", "")
	CGRIDA4 := utils.Sha1("aaa4", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:25Z", "")
	CGRIDA5 := utils.Sha1("aaa5", tm.String())
	CGRIDB2 := utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	CGRIDB3 := utils.Sha1("bbb3", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := mongoDb.RemCDRs([]string{CGRIDA1, CGRIDA2, CGRIDA3, CGRIDA4, CGRIDA5,
		CGRIDB2, CGRIDB3}); err != nil {
		t.Error(err.Error())
	}
	if cdrs, _, err := mongoDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(cdrs) != 20 {
		t.Error("Unexpected number of cdrs returned: ", len(cdrs))
	}
}

// Make sure that what we get is what we set
func TestMongoStoreRestoreCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr := &CDR{TOR: utils.VOICE, OriginID: "ccc1", OriginHost: "192.168.1.1", Source: "TEST_CDR", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr.CGRID = utils.Sha1(strCdr.OriginID, strCdr.SetupTime.String())
	if err := mongoDb.SetCdr(strCdr); err != nil {
		t.Error(err.Error())
	}
	if err := mongoDb.SetRatedCdr(strCdr); err != nil {
		t.Error(err.Error())
	}
	// Check RawCdr
	if rcvcdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{strCdr.CGRID}}); err != nil {
		t.Error(err.Error())
	} else if len(rcvcdrs) != 1 {
		t.Errorf("Unexpected cdrs returned: %+v", rcvcdrs)
	} else {
		rcvCdr := rcvcdrs[0]
		if strCdr.CGRID != rcvCdr.CGRID ||
			strCdr.TOR != rcvCdr.TOR ||
			strCdr.OriginID != rcvCdr.OriginID ||
			strCdr.OriginHost != rcvCdr.OriginHost ||
			strCdr.ReqType != rcvCdr.ReqType ||
			strCdr.Direction != rcvCdr.Direction ||
			strCdr.Tenant != rcvCdr.Tenant ||
			strCdr.Category != rcvCdr.Category ||
			strCdr.Account != rcvCdr.Account ||
			strCdr.Subject != rcvCdr.Subject ||
			strCdr.Destination != rcvCdr.Destination ||
			!strCdr.SetupTime.Equal(rcvCdr.SetupTime) ||
			!strCdr.AnswerTime.Equal(rcvCdr.AnswerTime) ||
			strCdr.Usage != rcvCdr.Usage ||
			strCdr.PDD != rcvCdr.PDD ||
			strCdr.Supplier != rcvCdr.Supplier ||
			strCdr.DisconnectCause != rcvCdr.DisconnectCause ||
			!reflect.DeepEqual(strCdr.ExtraFields, rcvCdr.ExtraFields) {
			t.Errorf("Expecting: %+v, received: %+v", strCdr, rcvcdrs[0])
		}
	}
	// Check RatedCdr
	if rcvcdrs, _, err := mongoDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{strCdr.CGRID}, FilterOnRated: true}); err != nil {
		t.Error(err.Error())
	} else if len(rcvcdrs) != 1 {
		t.Errorf("Unexpected cdrs returned: %+v", rcvcdrs)
	} else {
		rcvCdr := rcvcdrs[0]
		if strCdr.CGRID != rcvCdr.CGRID ||
			strCdr.TOR != rcvCdr.TOR ||
			strCdr.OriginID != rcvCdr.OriginID ||
			strCdr.OriginHost != rcvCdr.OriginHost ||
			strCdr.ReqType != rcvCdr.ReqType ||
			strCdr.Direction != rcvCdr.Direction ||
			strCdr.Tenant != rcvCdr.Tenant ||
			strCdr.Category != rcvCdr.Category ||
			strCdr.Account != rcvCdr.Account ||
			strCdr.Subject != rcvCdr.Subject ||
			strCdr.Destination != rcvCdr.Destination ||
			//!strCdr.SetupTime.Equal(rcvCdr.SetupTime) || // FixMe
			//!strCdr.AnswerTime.Equal(rcvCdr.AnswerTime) || // FixMe
			strCdr.Usage != rcvCdr.Usage ||
			strCdr.PDD != rcvCdr.PDD ||
			strCdr.Supplier != rcvCdr.Supplier ||
			strCdr.DisconnectCause != rcvCdr.DisconnectCause ||
			strCdr.Cost != rcvCdr.Cost ||
			!reflect.DeepEqual(strCdr.ExtraFields, rcvCdr.ExtraFields) {
			t.Errorf("Expecting: %+v, received: %+v", strCdr, rcvcdrs[0])
		}
	}
}
