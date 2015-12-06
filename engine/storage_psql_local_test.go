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
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var psqlDb *PostgresStorage

func TestPSQLCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	var err error
	if psqlDb, err = NewPostgresStorage("localhost", "5432", cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass,
		cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	}
	for _, scriptName := range []string{utils.CREATE_CDRS_TABLES_SQL, utils.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := psqlDb.CreateTablesFromScript(path.Join(*dataDir, "storage", utils.POSTGRES, scriptName)); err != nil {
			t.Error("Error on psqlDb creation: ", err.Error())
			return // No point in going further
		}
	}
	if _, err := psqlDb.Db.Query(fmt.Sprintf("SELECT 1 from %s", utils.TBL_CDRS)); err != nil {
		t.Error(err.Error())
	}
}

func TestPSQLSetGetTPTiming(t *testing.T) {
	if !*testLocal {
		return
	}
	tm := &utils.ApierTPTiming{TPid: utils.TEST_SQL, TimingId: "ALWAYS", Time: "00:00:00"}
	mtm := APItoModelTiming(tm)
	mtms := []TpTiming{*mtm}
	if err := psqlDb.SetTpTimings(mtms); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mtms[0], tmgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mtms, tmgs)
	}
	// Update
	tm.Time = "00:00:01"
	if err := psqlDb.SetTpTimings(mtms); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mtms[0], tmgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mtms, tmgs)
	}
}

func TestPSQLSetGetTPDestination(t *testing.T) {
	if !*testLocal {
		return
	}
	dst := &utils.TPDestination{TPid: utils.TEST_SQL, DestinationId: utils.TEST_SQL, Prefixes: []string{"+49", "+49151", "+49176"}}
	dests := APItoModelDestination(dst)
	if err := psqlDb.SetTpDestinations(dests); err != nil {
		t.Error(err.Error())
	}
	storData, err := psqlDb.GetTpDestinations(utils.TEST_SQL, utils.TEST_SQL)
	dsts, err := TpDestinations(storData).GetDestinations()
	if err != nil {
		t.Error(err.Error())
	} else if len(dst.Prefixes) != len(dsts[utils.TEST_SQL].Prefixes) {
		t.Errorf("Expecting: %+v, received: %+v", dst, dsts[utils.TEST_SQL])
	}
}

func TestPSQLSetGetTPRates(t *testing.T) {
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
	expectedTPRate := &utils.TPRate{TPid: utils.TEST_SQL, RateId: RT_ID, RateSlots: rtSlots}
	mRates := APItoModelRate(expectedTPRate)
	if err := psqlDb.SetTpRates(mRates); err != nil {
		t.Error(err.Error())
	}
	rts, err := psqlDb.GetTpRates(utils.TEST_SQL, RT_ID)
	trts, err := TpRates(rts).GetRates()
	if err != nil {
		t.Error(err.Error())
	} else if len(expectedTPRate.RateSlots) != len(trts[RT_ID].RateSlots) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPRate, trts[RT_ID])
	}
}

func TestPSQLSetGetTPDestinationRates(t *testing.T) {
	if !*testLocal {
		return
	}
	DR_ID := "DR_1"
	dr := &utils.DestinationRate{DestinationId: "DST_1", RateId: "RT_1", RoundingMethod: "*up", RoundingDecimals: 4}
	eDrs := &utils.TPDestinationRate{TPid: utils.TEST_SQL, DestinationRateId: DR_ID, DestinationRates: []*utils.DestinationRate{dr}}
	mdrs := APItoModelDestinationRate(eDrs)
	if err := psqlDb.SetTpDestinationRates(mdrs); err != nil {
		t.Error(err.Error())
	}
	if drs, err := psqlDb.GetTpDestinationRates(utils.TEST_SQL, DR_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mdrs[0], drs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mdrs, drs)
	}
}

func TestPSQLSetGetTPRatingPlans(t *testing.T) {
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

	if err := psqlDb.SetTpRatingPlans(mrp); err != nil {
		t.Error(err.Error())
	}
	if drps, err := psqlDb.GetTpRatingPlans(utils.TEST_SQL, RP_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mrp[0], drps[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mrp, drps)
	}
}

func TestPSQLSetGetTPRatingProfiles(t *testing.T) {
	if !*testLocal {
		return
	}
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RP_1"}}
	rp := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}

	mrp := APItoModelRatingProfile(rp)
	if err := psqlDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := psqlDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mrp[0], rps[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mrp, rps)
	}
}

func TestPSQLSetGetTPSharedGroups(t *testing.T) {
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
	if err := psqlDb.SetTpSharedGroups(mSgs); err != nil {
		t.Error(err.Error())
	}
	if sgs, err := psqlDb.GetTpSharedGroups(utils.TEST_SQL, SG_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mSgs[0], sgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mSgs, sgs)
	}
}

func TestPSQLSetGetTPCdrStats(t *testing.T) {
	if !*testLocal {
		return
	}
	CS_ID := "CDRSTATS_1"
	setCS := &utils.TPCdrStats{
		TPid:       utils.TEST_SQL,
		CdrStatsId: CS_ID,
		CdrStats: []*utils.TPCdrStat{
			&utils.TPCdrStat{QueueLength: "10", TimeWindow: "10m", Metrics: "ASR", Tenants: "cgrates.org", Categories: "call"},
		},
	}
	mcs := APItoModelCdrStat(setCS)
	if err := psqlDb.SetTpCdrStats(mcs); err != nil {
		t.Error(err.Error())
	}
	if cs, err := psqlDb.GetTpCdrStats(utils.TEST_SQL, CS_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mcs[0], cs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mcs, cs)
	}
}

func TestPSQLSetGetTPDerivedChargers(t *testing.T) {
	if !*testLocal {
		return
	}
	dc := &utils.TPDerivedCharger{RunId: utils.DEFAULT_RUNID, ReqTypeField: "^" + utils.META_PREPAID, AccountField: "^rif", SubjectField: "^rif",
		UsageField: "cgr_duration", SupplierField: "^supplier1"}
	dcs := &utils.TPDerivedChargers{TPid: utils.TEST_SQL, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: []*utils.TPDerivedCharger{dc}}
	mdcs := APItoModelDerivedCharger(dcs)
	if err := psqlDb.SetTpDerivedChargers(mdcs); err != nil {
		t.Error(err.Error())
	}
	if rDCs, err := psqlDb.GetTpDerivedChargers(&mdcs[0]); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mdcs[0], rDCs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mdcs, rDCs)
	}
}

func TestPSQLSetGetTPActions(t *testing.T) {
	if !*testLocal {
		return
	}
	ACTS_ID := "PREPAID_10"
	acts := []*utils.TPAction{
		&utils.TPAction{Identifier: "*topup_reset", BalanceType: "*monetary", Directions: "*out", Units: 10, ExpiryTime: "*unlimited",
			DestinationIds: "*any", BalanceWeight: 10, Weight: 10}}
	tpActions := &utils.TPActions{TPid: utils.TEST_SQL, ActionsId: ACTS_ID, Actions: acts}
	mas := APItoModelAction(tpActions)
	if err := psqlDb.SetTpActions(mas); err != nil {
		t.Error(err.Error())
	}
	if rTpActs, err := psqlDb.GetTpActions(utils.TEST_SQL, ACTS_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(mas[0], rTpActs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", mas, rTpActs)
	}
}

func TestPSQLTPActionTimings(t *testing.T) {
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
	if err := psqlDb.SetTpActionPlans(maps); err != nil {
		t.Error(err.Error())
	}
	if rAP, err := psqlDb.GetTpActionPlans(utils.TEST_SQL, AP_ID); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(maps[0], rAP[0]) {
		t.Errorf("Expecting: %+v, received: %+v", maps, rAP)
	}
}

func TestPSQLSetGetTPActionTriggers(t *testing.T) {
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
	if err := psqlDb.SetTpActionTriggers(matrg); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if rcvMpAtrgs, err := psqlDb.GetTpActionTriggers(utils.TEST_SQL, utils.TEST_SQL); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !modelEqual(matrg[0], rcvMpAtrgs[0]) {
		t.Errorf("Expecting: %+v, received: %+v", matrg, rcvMpAtrgs)
	}
}

func TestPSQLSetGetTpAccountActions(t *testing.T) {
	if !*testLocal {
		return
	}
	aa := &utils.TPAccountActions{TPid: utils.TEST_SQL, Tenant: "cgrates.org", Account: "1001",
		ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	maa := APItoModelAccountAction(aa)
	if err := psqlDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := psqlDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if !modelEqual(*maa, aas[0]) {
		t.Errorf("Expecting: %+v, received: %+v", maa, aas)
	}
}

func TestPSQLGetTPIds(t *testing.T) {
	if !*testLocal {
		return
	}
	eTPIds := []string{utils.TEST_SQL}
	if tpIds, err := psqlDb.GetTpIds(); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(eTPIds, tpIds) {
		t.Errorf("Expecting: %+v, received: %+v", eTPIds, tpIds)
	}
}

func TestPSQLRemoveTPData(t *testing.T) {
	if !*testLocal {
		return
	}
	// Create Timings
	tm := &utils.ApierTPTiming{TPid: utils.TEST_SQL, TimingId: "ALWAYS", Time: "00:00:00"}
	tms := APItoModelTiming(tm)
	if err := psqlDb.SetTpTimings([]TpTiming{*tms}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Remove Timings
	if err := psqlDb.RemTpData(utils.TBL_TP_TIMINGS, utils.TEST_SQL, map[string]string{"tag": tm.TimingId}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tmgs) != 0 {
		t.Errorf("Timings should be empty, got instead: %+v", tmgs)
	}
	// Create RatingProfile
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1"}}
	rp := &utils.TPRatingProfile{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	mrp := APItoModelRatingProfile(rp)
	if err := psqlDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := psqlDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Remove RatingProfile
	if err := psqlDb.RemTpData(utils.TBL_TP_RATE_PROFILES, rp.TPid, map[string]string{"loadid": rp.LoadId, "direction": rp.Direction, "tenant": rp.Tenant, "category": rp.Category, "subject": rp.Subject}); err != nil {
		t.Error(err.Error())
	}
	if rps, err := psqlDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err)
	} else if len(rps) != 0 {
		t.Errorf("RatingProfiles different than 0: %+v", rps)
	}
	// Create AccountActions
	aa := &utils.TPAccountActions{TPid: utils.TEST_SQL, LoadId: utils.TEST_SQL, Tenant: "cgrates.org", Account: "1001",
		ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	maa := APItoModelAccountAction(aa)
	if err := psqlDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := psqlDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove AccountActions
	if err := psqlDb.RemTpData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.TPid, map[string]string{"loadid": aa.LoadId, "tenant": aa.Tenant, "account": aa.Account}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := psqlDb.GetTpAccountActions(maa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
	// Create again so we can test complete TP removal
	if err := psqlDb.SetTpTimings([]TpTiming{*tms}); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Create RatingProfile
	if err := psqlDb.SetTpRatingProfiles(mrp); err != nil {
		t.Error(err.Error())
	}
	if rps, err := psqlDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Create AccountActions
	if err := psqlDb.SetTpAccountActions([]TpAccountAction{*maa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := psqlDb.GetTpAccountActions(maa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove TariffPlan completely
	if err := psqlDb.RemTpData("", utils.TEST_SQL, nil); err != nil {
		t.Error(err.Error())
	}
	// Make sure we have removed it
	if tms, err := psqlDb.GetTpTimings(utils.TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tms) != 0 {
		t.Errorf("Non empty timings: %+v", tms)
	}
	if rpfs, err := psqlDb.GetTpRatingProfiles(&mrp[0]); err != nil {
		t.Error(err)
	} else if len(rpfs) != 0 {
		t.Errorf("Non empty rpfs: %+v", rpfs)
	}
	if aas, err := psqlDb.GetTpAccountActions(maa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
}

func TestPSQLSetCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa1", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:20Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "10s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2", utils.CDRSOURCE: utils.TEST_SQL}

	cgrCdr2 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa2", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_PREPAID, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:22Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "20", utils.PDD: "7s", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": utils.TEST_SQL}

	cgrCdr3 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa3", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "premium_call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "1001", utils.SETUP_TIME: "2013-11-07T08:42:24Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "60s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": utils.TEST_SQL}

	cgrCdr4 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa4", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_PSEUDOPREPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "+4986517174964", utils.SETUP_TIME: "2013-11-07T08:42:21Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "1m2s", utils.PDD: "4s", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": utils.TEST_SQL}

	cgrCdr5 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa5", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_POSTPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "+4986517174963", utils.SETUP_TIME: "2013-11-07T08:42:25Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "15s", utils.PDD: "7s", utils.SUPPLIER: "SUPPL1", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": utils.TEST_SQL}

	for _, cdr := range []*CgrCdr{cgrCdr1, cgrCdr2, cgrCdr3, cgrCdr4, cgrCdr5} {
		if err := psqlDb.SetCdr(cdr.AsStoredCdr("")); err != nil {
			t.Error(err.Error())
		}
	}
	strCdr1 := &CDR{TOR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CGRID = utils.Sha1(strCdr1.OriginID, strCdr1.SetupTime.String())
	strCdr2 := &CDR{TOR: utils.VOICE, OriginID: "bbb2", OriginHost: "192.168.1.2", Source: "UNKNOWN2", RequestType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, PDD: time.Duration(4) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CGRID = utils.Sha1(strCdr2.OriginID, strCdr2.SetupTime.String())
	strCdr3 := &CDR{TOR: utils.VOICE, OriginID: "bbb3", OriginHost: "192.168.1.1", Source: utils.TEST_SQL, RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1000", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(2) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr3.CGRID = utils.Sha1(strCdr3.OriginID, strCdr3.SetupTime.String())

	for _, cdr := range []*CDR{strCdr1, strCdr2, strCdr3} {
		if err := psqlDb.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestPSQLCallCost(t *testing.T) {
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
	if err := psqlDb.LogCallCost(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := psqlDb.GetCallCostLog(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
	// UPDATE test here
	cc.Category = "premium_call"
	if err := psqlDb.LogCallCost(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := psqlDb.GetCallCostLog(CGRID, utils.TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
}

func TestPSQLSetRatedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr1 := &CDR{TOR: utils.VOICE, OriginID: "bbb1", OriginHost: "192.168.1.1", Source: "UNKNOWN", RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CGRID = utils.Sha1(strCdr1.OriginID, strCdr1.SetupTime.String())
	strCdr2 := &CDR{TOR: utils.VOICE, OriginID: "bbb2", OriginHost: "192.168.1.2", Source: "UNKNOWN", RequestType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, PDD: time.Duration(7) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CGRID = utils.Sha1(strCdr2.OriginID, strCdr2.SetupTime.String())
	strCdr3 := &CDR{TOR: utils.VOICE, OriginID: "bbb3", OriginHost: "192.168.1.1", Source: utils.TEST_SQL, RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1002", Destination: "+4986517174964",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(2) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       "wholesale_run", Cost: 1.201}
	strCdr3.CGRID = utils.Sha1(strCdr3.OriginID, strCdr3.SetupTime.String())

	for _, cdr := range []*CDR{strCdr1, strCdr2, strCdr3} {
		if err := psqlDb.SetRatedCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestPSQLGetCDRs(t *testing.T) {
	if !*testLocal {
		return
	}
	var timeStart, timeEnd time.Time
	// All CDRs, no filter
	if storedCdrs, _, err := psqlDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count ALL
	if storedCdrs, count, err := psqlDb.GetCDRs(&utils.CDRsFilter{Count: true}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	} else if count != 8 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Limit 5
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Offset 5
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Offset with limit 2
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
	}
	// Filter on CGRIDs
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count on CGRIDS
	if _, count, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 2 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Filter on CGRIDs plus RequestType
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, RequestTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count on multiple filter
	if _, count, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, RequestTypes: []string{utils.META_PREPAID}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 1 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Filter on runId
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{RunIDs: []string{utils.DEFAULT_RUNID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on TOR
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{TORs: []string{utils.SMS}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple TOR
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{TORs: []string{utils.SMS, utils.VOICE}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on OriginHost
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple OriginHost
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{OriginHosts: []string{"192.168.1.1", "192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on Source
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Sources: []string{"UNKNOWN"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple Source
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Sources: []string{"UNKNOWN", "UNKNOWN2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on RequestType
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple RequestType
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on direction
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Directions: []string{"*out"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tenant
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple tenants
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on category
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple categories
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Categories: []string{"premium_call", "call"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on account
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple account
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Accounts: []string{"1001", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on subject
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple subject
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{Subjects: []string{"1000", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on destPrefix
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple DestinationPrefixes
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{DestinationPrefixes: []string{"1001", "+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 4 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreRated
	var OrderIDStart, OrderIDEnd int64 // Capture also orderIds for the next test
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{MaxCost: utils.Float64Pointer(0.0)}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	} else {
		for _, cdr := range storedCdrs {
			if cdr.OrderID < OrderIDStart {
				OrderIDStart = cdr.OrderID
			}
			if cdr.OrderID > OrderIDEnd {
				OrderIDEnd = cdr.OrderID
			}
		}
	}
	// Filter on OrderIDStart
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{OrderIDStart: OrderIDStart}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on OrderIDStart and OrderIDEnd
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{OrderIDStart: OrderIDStart, OrderIDEnd: OrderIDEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 4 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart
	timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on minPDD
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{MinPDD: utils.Float64Pointer(3)}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 7 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on maxPDD
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{MaxPDD: utils.Float64Pointer(3)}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on minPDD, maxPDD
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{MinPDD: utils.Float64Pointer(3), MaxPDD: utils.Float64Pointer(5)}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Combined filter
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{RequestTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreDerived
	if storedCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd, FilterOnRated: true}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 { // ToDo: Recheck this value
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

func TestPSQLRemCDRs(t *testing.T) {
	if !*testLocal {
		return
	}
	CGRIDB1 := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := psqlDb.RemCDRs([]string{CGRIDB1}); err != nil {
		t.Error(err.Error())
	}
	if storedCdrs, _, err := psqlDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 7 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
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
	if err := psqlDb.RemCDRs([]string{CGRIDA1, CGRIDA2, CGRIDA3, CGRIDA4, CGRIDA5,
		CGRIDB2, CGRIDB3}); err != nil {
		t.Error(err.Error())
	}
	if storedCdrs, _, err := psqlDb.GetCDRs(new(utils.CDRsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

// Make sure that what we get is what we set
func TestPSQLStoreRestoreCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr := &CDR{TOR: utils.VOICE, OriginID: "ccc1", OriginHost: "192.168.1.1", Source: "TEST_CDR", RequestType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, PDD: time.Duration(3) * time.Second, Supplier: "SUPPL1",
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		RunID:       utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr.CGRID = utils.Sha1(strCdr.OriginID, strCdr.SetupTime.String())
	if err := psqlDb.SetCdr(strCdr); err != nil {
		t.Error(err.Error())
	}
	if err := psqlDb.SetRatedCdr(strCdr); err != nil {
		t.Error(err.Error())
	}
	// Check RawCdr
	if rcvCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{strCdr.CGRID}}); err != nil {
		t.Error(err.Error())
	} else if len(rcvCdrs) != 1 {
		t.Errorf("Unexpected cdrs returned: %+v", rcvCdrs)
	} else {
		rcvCdr := rcvCdrs[0]
		if strCdr.CGRID != rcvCdr.CGRID ||
			strCdr.TOR != rcvCdr.TOR ||
			strCdr.OriginID != rcvCdr.OriginID ||
			strCdr.OriginHost != rcvCdr.OriginHost ||
			strCdr.RequestType != rcvCdr.RequestType ||
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
			t.Errorf("Expecting: %+v, received: %+v", strCdr, rcvCdrs[0])
		}
	}
	// Check RatedCdr
	if rcvCdrs, _, err := psqlDb.GetCDRs(&utils.CDRsFilter{CGRIDs: []string{strCdr.CGRID}, FilterOnRated: true}); err != nil {
		t.Error(err.Error())
	} else if len(rcvCdrs) != 1 {
		t.Errorf("Unexpected cdrs returned: %+v", rcvCdrs)
	} else {
		rcvCdr := rcvCdrs[0]
		if strCdr.CGRID != rcvCdr.CGRID ||
			strCdr.TOR != rcvCdr.TOR ||
			strCdr.OriginID != rcvCdr.OriginID ||
			strCdr.OriginHost != rcvCdr.OriginHost ||
			strCdr.RequestType != rcvCdr.RequestType ||
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
			t.Errorf("Expecting: %+v, received: %+v", strCdr, rcvCdrs[0])
		}
	}
}
