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
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var mysqlDb *MySQLStorage

func TestMySQLCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	if d, err := NewMySQLStorage(cgrConfig.StorDBHost, cgrConfig.StorDBPort, cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass,
		cgrConfig.StorDBMaxOpenConns, cgrConfig.StorDBMaxIdleConns); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		mysqlDb = d.(*MySQLStorage)
	}
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysqlDb.CreateTablesFromScript(path.Join(*dataDir, "storage", utils.MYSQL, scriptName)); err != nil {
			t.Error("Error on mysqlDb creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysqlDb.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMySQLSetGetTPTiming(t *testing.T) {
	if !*testLocal {
		return
	}
	tm := &utils.ApierTPTiming{TPid: TEST_SQL, TimingId: "ALWAYS", Time: "00:00:00"}
	if err := mysqlDb.SetTPTiming(tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(tm, tmgs[tm.TimingId]) {
		t.Errorf("Expecting: %+v, received: %+v", tm, tmgs[tm.TimingId])
	}
	// Update
	tm.Time = "00:00:01"
	if err := mysqlDb.SetTPTiming(tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(tm, tmgs[tm.TimingId]) {
		t.Errorf("Expecting: %+v, received: %+v", tm, tmgs[tm.TimingId])
	}
}

func TestMySQLSetGetTPDestination(t *testing.T) {
	if !*testLocal {
		return
	}
	dst := &Destination{Id: TEST_SQL, Prefixes: []string{"+49", "+49151", "+49176"}}
	if err := mysqlDb.SetTPDestination(TEST_SQL, dst); err != nil {
		t.Error(err.Error())
	}
	if dsts, err := mysqlDb.GetTpDestinations(TEST_SQL, TEST_SQL); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(dst, dsts[TEST_SQL]) {
		t.Errorf("Expecting: %+v, received: %+v", dst, dsts[TEST_SQL])
	}
}

func TestMySQLSetGetTPRates(t *testing.T) {
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
	mpRates := map[string][]*utils.RateSlot{RT_ID: rtSlots}
	expectedTPRate := &utils.TPRate{TPid: TEST_SQL, RateId: RT_ID, RateSlots: rtSlots}
	if err := mysqlDb.SetTPRates(TEST_SQL, mpRates); err != nil {
		t.Error(err.Error())
	}
	if rts, err := mysqlDb.GetTpRates(TEST_SQL, RT_ID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(expectedTPRate, rts[RT_ID]) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPRate, rts[RT_ID])
	}
}

func TestMySQLSetGetTPDestinationRates(t *testing.T) {
	if !*testLocal {
		return
	}
	DR_ID := "DR_1"
	dr := &utils.DestinationRate{DestinationId: "DST_1", RateId: "RT_1", RoundingMethod: "*up", RoundingDecimals: 4}
	drs := map[string][]*utils.DestinationRate{DR_ID: []*utils.DestinationRate{dr}}
	eDrs := &utils.TPDestinationRate{TPid: TEST_SQL, DestinationRateId: DR_ID, DestinationRates: []*utils.DestinationRate{dr}}
	if err := mysqlDb.SetTPDestinationRates(TEST_SQL, drs); err != nil {
		t.Error(err.Error())
	}
	if drs, err := mysqlDb.GetTpDestinationRates(TEST_SQL, DR_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(eDrs, drs[DR_ID]) {
		fmt.Printf("Received: %+v\n", drs[DR_ID].DestinationRates[0])
		t.Errorf("Expecting: %+v, received: %+v", eDrs, drs[DR_ID])
	}
}

func TestMySQLSetGetTPRatingPlans(t *testing.T) {
	if !*testLocal {
		return
	}
	RP_ID := "RP_1"
	rbBinding := &utils.TPRatingPlanBinding{DestinationRatesId: "DR_1", TimingId: "TM_1", Weight: 10.0}
	drts := map[string][]*utils.TPRatingPlanBinding{RP_ID: []*utils.TPRatingPlanBinding{rbBinding}}
	if err := mysqlDb.SetTPRatingPlans(TEST_SQL, drts); err != nil {
		t.Error(err.Error())
	}
	if drps, err := mysqlDb.GetTpRatingPlans(TEST_SQL, RP_ID, nil); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(drts, drps) {
		t.Errorf("Expecting: %+v, received: %+v", drts, drps)
	}
}

func TestMySQLSetGetTPRatingProfiles(t *testing.T) {
	if !*testLocal {
		return
	}
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RP_1"}}
	rp := &utils.TPRatingProfile{TPid: TEST_SQL, LoadId: TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	setRps := map[string]*utils.TPRatingProfile{rp.KeyId(): rp}
	if err := mysqlDb.SetTPRatingProfiles(TEST_SQL, setRps); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(setRps, rps) {
		t.Errorf("Expecting: %+v, received: %+v", setRps, rps)
	}

}

func TestMySQLSetGetTPSharedGroups(t *testing.T) {
	if !*testLocal {
		return
	}
	SG_ID := "SG_1"
	setSgs := map[string][]*utils.TPSharedGroup{SG_ID: []*utils.TPSharedGroup{&utils.TPSharedGroup{Account: "dan", Strategy: "*lowest_first", RatingSubject: "lowest_rates"}}}
	if err := mysqlDb.SetTPSharedGroups(TEST_SQL, setSgs); err != nil {
		t.Error(err.Error())
	}
	if sgs, err := mysqlDb.GetTpSharedGroups(TEST_SQL, SG_ID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(setSgs, sgs) {
		t.Errorf("Expecting: %+v, received: %+v", setSgs, sgs)
	}
}

func TestMySQLSetGetTPCdrStats(t *testing.T) {
	if !*testLocal {
		return
	}
	CS_ID := "CDRSTATS_1"
	setCS := map[string][]*utils.TPCdrStat{CS_ID: []*utils.TPCdrStat{
		&utils.TPCdrStat{QueueLength: "10", TimeWindow: "10m", Metrics: "ASR", Tenant: "cgrates.org", Category: "call"},
	}}
	if err := mysqlDb.SetTPCdrStats(TEST_SQL, setCS); err != nil {
		t.Error(err.Error())
	}
	if cs, err := mysqlDb.GetTpCdrStats(TEST_SQL, CS_ID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(setCS, cs) {
		t.Errorf("Expecting: %+v, received: %+v", setCS, cs)
	}
}

func TestMySQLSetGetTPDerivedChargers(t *testing.T) {
	if !*testLocal {
		return
	}
	dc := &utils.TPDerivedCharger{RunId: utils.DEFAULT_RUNID, ReqTypeField: "^" + utils.META_PREPAID, AccountField: "^rif", SubjectField: "^rif", UsageField: "cgr_duration"}
	dcs := &utils.TPDerivedChargers{TPid: TEST_SQL, Direction: utils.OUT, Tenant: "cgrates.org", Category: "call", Account: "dan", Subject: "dan", DerivedChargers: []*utils.TPDerivedCharger{dc}}
	DCS_ID := dcs.GetDerivedChargesId()
	setDCs := map[string][]*utils.TPDerivedCharger{DCS_ID: []*utils.TPDerivedCharger{dc}}
	if err := mysqlDb.SetTPDerivedChargers(TEST_SQL, setDCs); err != nil {
		t.Error(err.Error())
	}
	if rDCs, err := mysqlDb.GetTpDerivedChargers(dcs); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(dcs, rDCs[DCS_ID]) {
		t.Errorf("Expecting: %+v, received: %+v", dcs, rDCs[DCS_ID])
	}
}

func TestMySQLSetGetTPActions(t *testing.T) {
	if !*testLocal {
		return
	}
	ACTS_ID := "PREPAID_10"
	acts := []*utils.TPAction{
		&utils.TPAction{Identifier: "*topup_reset", BalanceType: "*monetary", Direction: "*out", Units: 10, ExpiryTime: "*unlimited",
			DestinationId: "*any", BalanceWeight: 10, Weight: 10}}
	tpActions := &utils.TPActions{TPid: TEST_SQL, ActionsId: ACTS_ID, Actions: acts}
	if err := mysqlDb.SetTPActions(TEST_SQL, map[string][]*utils.TPAction{ACTS_ID: acts}); err != nil {
		t.Error(err.Error())
	}
	if rTpActs, err := mysqlDb.GetTPActions(TEST_SQL, ACTS_ID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(tpActions, rTpActs) {
		t.Errorf("Expecting: %+v, received: %+v", tpActions, rTpActs)
	}
}

func TestMySQLTPActionTimings(t *testing.T) {
	if !*testLocal {
		return
	}
	AP_ID := "AP_1"
	ap := map[string][]*utils.TPActionTiming{AP_ID: []*utils.TPActionTiming{&utils.TPActionTiming{ActionsId: "ACTS_1", TimingId: "TM_1", Weight: 10.0}}}
	if err := mysqlDb.SetTPActionTimings(TEST_SQL, ap); err != nil {
		t.Error(err.Error())
	}
	if rAP, err := mysqlDb.GetTPActionTimings(TEST_SQL, AP_ID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(ap, rAP) {
		t.Errorf("Expecting: %+v, received: %+v", ap, rAP)
	}
}

func TestMySQLSetGetTPActionTriggers(t *testing.T) {
	if !*testLocal {
		return
	}
	atrg := &utils.TPActionTrigger{
		Id:                   "MY_FIRST_ATGR",
		BalanceType:          "*monetary",
		BalanceDirection:     "*out",
		ThresholdType:        "*min_balance",
		ThresholdValue:       2.0,
		Recurrent:            true,
		BalanceDestinationId: "*any",
		Weight:               10.0,
		ActionsId:            "LOG_BALANCE",
	}
	mpAtrgs := map[string][]*utils.TPActionTrigger{TEST_SQL: []*utils.TPActionTrigger{atrg}}
	if err := mysqlDb.SetTPActionTriggers(TEST_SQL, mpAtrgs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if rcvMpAtrgs, err := mysqlDb.GetTpActionTriggers(TEST_SQL, TEST_SQL); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(mpAtrgs, rcvMpAtrgs) {
		t.Errorf("Expecting: %v, received: %v", mpAtrgs, rcvMpAtrgs)
	}
}

func TestMySQLSetGetTpAccountActions(t *testing.T) {
	if !*testLocal {
		return
	}
	aa := &utils.TPAccountActions{TPid: TEST_SQL, Tenant: "cgrates.org", Account: "1001",
		Direction: "*out", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	if err := mysqlDb.SetTPAccountActions(aa.TPid, map[string]*utils.TPAccountActions{aa.KeyId(): aa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(aa, aas[aa.KeyId()]) {
		t.Errorf("Expecting: %+v, received: %+v", aa, aas[aa.KeyId()])
	}
}

func TestMySQLGetTPIds(t *testing.T) {
	if !*testLocal {
		return
	}
	eTPIds := []string{TEST_SQL}
	if tpIds, err := mysqlDb.GetTPIds(); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(eTPIds, tpIds) {
		t.Errorf("Expecting: %+v, received: %+v", eTPIds, tpIds)
	}
}

func TestMySQLRemoveTPData(t *testing.T) {
	if !*testLocal {
		return
	}
	// Create Timings
	tm := &utils.ApierTPTiming{TPid: TEST_SQL, TimingId: "ALWAYS", Time: "00:00:00"}
	if err := mysqlDb.SetTPTiming(tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Remove Timings
	if err := mysqlDb.RemTPData(utils.TBL_TP_TIMINGS, TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tmgs) != 0 {
		t.Errorf("Timings should be empty, got instead: %+v", tmgs)
	}
	// Create RatingProfile
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1"}}
	rp := &utils.TPRatingProfile{TPid: TEST_SQL, LoadId: TEST_SQL, Tenant: "cgrates.org", Category: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	if err := mysqlDb.SetTPRatingProfiles(TEST_SQL, map[string]*utils.TPRatingProfile{rp.KeyId(): rp}); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Remove RatingProfile
	if err := mysqlDb.RemTPData(utils.TBL_TP_RATE_PROFILES, rp.TPid, rp.LoadId, rp.Direction, rp.Tenant, rp.Category, rp.Subject); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err)
	} else if len(rps) != 0 {
		t.Errorf("RatingProfiles different than 0: %+v", rps)
	}
	// Create AccountActions
	aa := &utils.TPAccountActions{TPid: TEST_SQL, LoadId: TEST_SQL, Tenant: "cgrates.org", Account: "1001",
		Direction: "*out", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	if err := mysqlDb.SetTPAccountActions(aa.TPid, map[string]*utils.TPAccountActions{aa.KeyId(): aa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove AccountActions
	if err := mysqlDb.RemTPData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.TPid, aa.LoadId, aa.Direction, aa.Tenant, aa.Account); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
	// Create again so we can test complete TP removal
	if err := mysqlDb.SetTPTiming(tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Create RatingProfile
	if err := mysqlDb.SetTPRatingProfiles(TEST_SQL, map[string]*utils.TPRatingProfile{rp.KeyId(): rp}); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Create AccountActions
	if err := mysqlDb.SetTPAccountActions(aa.TPid, map[string]*utils.TPAccountActions{aa.KeyId(): aa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove TariffPlan completely
	if err := mysqlDb.RemTPData("", TEST_SQL); err != nil {
		t.Error(err.Error())
	}
	// Make sure we have removed it
	if tms, err := mysqlDb.GetTpTimings(TEST_SQL, tm.TimingId); err != nil {
		t.Error(err)
	} else if len(tms) != 0 {
		t.Errorf("Non empty timings: %+v", tms)
	}
	if rpfs, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err)
	} else if len(rpfs) != 0 {
		t.Errorf("Non empty rpfs: %+v", rpfs)
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err)
	} else if len(aas) != 0 {
		t.Errorf("Non empty account actions: %+v", aas)
	}
}

func TestMySQLSetCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa1", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:20Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "10s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", utils.CDRSOURCE: TEST_SQL}
	cgrCdr2 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa2", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_PREPAID, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:22Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "20", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr3 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa3", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: utils.META_RATED, utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "premium_call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "1001", utils.SETUP_TIME: "2013-11-07T08:42:24Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "60s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr4 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa4", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_PSEUDOPREPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "+4986517174964", utils.SETUP_TIME: "2013-11-07T08:42:21Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "1m2s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr5 := &CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa5", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: utils.META_POSTPAID, utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "+4986517174963", utils.SETUP_TIME: "2013-11-07T08:42:25Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "15s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	for _, cdr := range []*CgrCdr{cgrCdr1, cgrCdr2, cgrCdr3, cgrCdr4, cgrCdr5} {
		if err := mysqlDb.SetCdr(cdr.AsStoredCdr()); err != nil {
			t.Error(err.Error())
		}
	}
	strCdr1 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CgrId = utils.Sha1(strCdr1.AccId, strCdr1.SetupTime.String())
	strCdr2 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: "UNKNOWN2", ReqType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CgrId = utils.Sha1(strCdr2.AccId, strCdr2.SetupTime.String())
	strCdr3 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1000", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr3.CgrId = utils.Sha1(strCdr3.AccId, strCdr3.SetupTime.String())

	for _, cdr := range []*StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysqlDb.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMySQLSetRatedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr1 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CgrId = utils.Sha1(strCdr1.AccId, strCdr1.SetupTime.String())
	strCdr2 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: "UNKNOWN", ReqType: utils.META_PREPAID,
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CgrId = utils.Sha1(strCdr2.AccId, strCdr2.SetupTime.String())
	strCdr3 := &StoredCdr{TOR: utils.VOICE, AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: utils.META_RATED,
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1002", Destination: "+4986517174964",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: "wholesale_run", Cost: 1.201}
	strCdr3.CgrId = utils.Sha1(strCdr3.AccId, strCdr3.SetupTime.String())

	for _, cdr := range []*StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysqlDb.SetRatedCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestMySQLCallCost(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrId := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
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
	if err := mysqlDb.LogCallCost(cgrId, TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mysqlDb.GetCallCostLog(cgrId, TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
	// UPDATE test here
	cc.Category = "premium_call"
	if err := mysqlDb.LogCallCost(cgrId, TEST_SQL, utils.DEFAULT_RUNID, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mysqlDb.GetCallCostLog(cgrId, TEST_SQL, utils.DEFAULT_RUNID); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
}

func TestMySQLGetStoredCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var timeStart, timeEnd time.Time
	// All CDRs, no filter
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(new(utils.CdrsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count ALL
	if storedCdrs, count, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Count: true}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	} else if count != 8 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Limit 5
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Offset 5
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(5), Offset: utils.IntPointer(0)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Offset with limit 2
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Paginator: utils.Paginator{Limit: utils.IntPointer(2), Offset: utils.IntPointer(5)}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", len(storedCdrs))
	}
	// Filter on cgrids
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count on CGRIDS
	if _, count, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 2 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Filter on cgrids plus reqType
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Count on multiple filter
	if _, count, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CgrIds: []string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())}, ReqTypes: []string{utils.META_PREPAID}, Count: true}); err != nil {
		t.Error(err.Error())
	} else if count != 1 {
		t.Error("Unexpected count of StoredCdrs returned: ", count)
	}
	// Filter on runId
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{RunIds: []string{utils.DEFAULT_RUNID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on TOR
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Tors: []string{utils.SMS}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple TOR
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Tors: []string{utils.SMS, utils.VOICE}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrHost
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CdrHosts: []string{"192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple cdrHost
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CdrHosts: []string{"192.168.1.1", "192.168.1.2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrSource
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CdrSources: []string{"UNKNOWN"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple cdrSource
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CdrSources: []string{"UNKNOWN", "UNKNOWN2"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on reqType
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{ReqTypes: []string{utils.META_PREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple reqType
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{ReqTypes: []string{utils.META_PREPAID, utils.META_PSEUDOPREPAID}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on direction
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Directions: []string{"*out"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tenant
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Tenants: []string{"itsyscom.com"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple tenants
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Tenants: []string{"itsyscom.com", "cgrates.org"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on category
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Categories: []string{"premium_call"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple categories
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Categories: []string{"premium_call", "call"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on account
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Accounts: []string{"1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple account
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Accounts: []string{"1001", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on subject
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Subjects: []string{"1000"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple subject
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{Subjects: []string{"1000", "1002"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on destPrefix
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{DestPrefixes: []string{"+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple destPrefixes
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{DestPrefixes: []string{"1001", "+498651"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 4 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ratedAccount
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{RatedAccounts: []string{"8001"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ratedSubject
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{RatedSubjects: []string{"91001"}}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreRated
	var orderIdStart, orderIdEnd int64 // Capture also orderIds for the next test
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{CostEnd: utils.Float64Pointer(0.0)}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	} else {
		for _, cdr := range storedCdrs {
			if cdr.OrderId < orderIdStart {
				orderIdStart = cdr.OrderId
			}
			if cdr.OrderId > orderIdEnd {
				orderIdEnd = cdr.OrderId
			}
		}
	}
	// Filter on orderIdStart
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{OrderIdStart: orderIdStart}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on orderIdStart and orderIdEnd
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{OrderIdStart: orderIdStart, OrderIdEnd: orderIdEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 4 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart
	timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{AnswerTimeStart: &timeStart}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Combined filter
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{ReqTypes: []string{utils.META_RATED}, AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreDerived
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(&utils.CdrsFilter{AnswerTimeStart: &timeStart, AnswerTimeEnd: &timeEnd, IgnoreDerived: true}); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 { // ToDo: Recheck this value
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

func TestMySQLRemStoredCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrIdB1 := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := mysqlDb.RemStoredCdrs([]string{cgrIdB1}); err != nil {
		t.Error(err.Error())
	}
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(new(utils.CdrsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 7 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	tm, _ := utils.ParseTimeDetectLayout("2013-11-08T08:42:20Z")
	cgrIdA1 := utils.Sha1("aaa1", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-08T08:42:22Z")
	cgrIdA2 := utils.Sha1("aaa2", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:24Z")
	cgrIdA3 := utils.Sha1("aaa3", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:21Z")
	cgrIdA4 := utils.Sha1("aaa4", tm.String())
	tm, _ = utils.ParseTimeDetectLayout("2013-11-07T08:42:25Z")
	cgrIdA5 := utils.Sha1("aaa5", tm.String())
	cgrIdB2 := utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	cgrIdB3 := utils.Sha1("bbb3", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := mysqlDb.RemStoredCdrs([]string{cgrIdA1, cgrIdA2, cgrIdA3, cgrIdA4, cgrIdA5,
		cgrIdB2, cgrIdB3}); err != nil {
		t.Error(err.Error())
	}
	if storedCdrs, _, err := mysqlDb.GetStoredCdrs(new(utils.CdrsFilter)); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}
