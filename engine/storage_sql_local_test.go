/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"path"
	"reflect"
	"testing"
	"time"
)

/*
README:

 Enable these tests by passing '-local' to the go test command

 Only database supported for now is mysqlDb. Postgres could be easily extended if needed.

 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data.

 Prior running the tests, create database and users by running:
  mysqlDb -pyourrootpwd < /usr/share/cgrates/storage/mysqlDb/create_mysqlDb_with_users.sql
*/

var mysqlDb *MySQLStorage

func TestCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	if d, err := NewMySQLStorage(cgrConfig.StorDBHost, cgrConfig.StorDBPort, cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		mysqlDb = d.(*MySQLStorage)
	}
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysqlDb.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", scriptName)); err != nil {
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

func TestRemoveData(t *testing.T) {
	if !*testLocal {
		return
	}
	// Create Timings
	tm := &utils.TPTiming{Id: "ALWAYS", StartTime: "00:00:00"}
	if err := mysqlDb.SetTPTiming(TEST_SQL, tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Remove Timings
	if err := mysqlDb.RemTPData(utils.TBL_TP_TIMINGS, TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysqlDb.GetTpTimings(TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) != 0 {
		t.Error("Did not remove TPTiming")
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
	if err := mysqlDb.RemTPData(utils.TBL_TP_RATE_PROFILES, rp.TPid, rp.LoadId, rp.Tenant, rp.Category, rp.Direction, rp.Subject); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysqlDb.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if len(rps) != 0 {
		t.Error("Did not remove TPRatingProfile")
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
	if err := mysqlDb.RemTPData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.TPid, aa.LoadId, aa.Tenant, aa.Account, aa.Direction); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysqlDb.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if len(aas) != 0 {
		t.Error("Did not remove TPAccountActions")
	}
}

func TestSetCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := &utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa1", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: "rated", utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:20Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "10s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", utils.CDRSOURCE: TEST_SQL}
	cgrCdr2 := &utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa2", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: "prepaid", utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "1002", utils.SETUP_TIME: "2013-11-08T08:42:22Z",
		utils.ANSWER_TIME: "2013-11-08T08:42:26Z", utils.USAGE: "20", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr3 := &utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa3", utils.CDRHOST: "192.168.1.1", utils.REQTYPE: "rated", utils.DIRECTION: "*out", utils.TENANT: "cgrates.org",
		utils.CATEGORY: "premium_call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "1001", utils.SETUP_TIME: "2013-11-07T08:42:24Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "60s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr4 := &utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa4", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: "pseudoprepaid", utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1001", utils.SUBJECT: "1001", utils.DESTINATION: "+4986517174964", utils.SETUP_TIME: "2013-11-07T08:42:21Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "1m2s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	cgrCdr5 := &utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaa5", utils.CDRHOST: "192.168.1.2", utils.REQTYPE: "postpaid", utils.DIRECTION: "*out", utils.TENANT: "itsyscom.com",
		utils.CATEGORY: "call", utils.ACCOUNT: "1002", utils.SUBJECT: "1002", utils.DESTINATION: "+4986517174963", utils.SETUP_TIME: "2013-11-07T08:42:25Z",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "15s", "field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}

	for _, cdr := range []*utils.CgrCdr{cgrCdr1, cgrCdr2, cgrCdr3, cgrCdr4, cgrCdr5} {
		if err := mysqlDb.SetCdr(cdr.AsStoredCdr()); err != nil {
			t.Error(err.Error())
		}
	}
	strCdr1 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CgrId = utils.Sha1(strCdr1.AccId, strCdr1.SetupTime.String())
	strCdr2 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: "UNKNOWN2", ReqType: "prepaid",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CgrId = utils.Sha1(strCdr2.AccId, strCdr2.SetupTime.String())
	strCdr3 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: "rated",
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1000", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr3.CgrId = utils.Sha1(strCdr3.AccId, strCdr3.SetupTime.String())

	for _, cdr := range []*utils.StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysqlDb.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestSetRatedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr1 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr1.CgrId = utils.Sha1(strCdr1.AccId, strCdr1.SetupTime.String())
	strCdr2 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: "UNKNOWN", ReqType: "prepaid",
		Direction: "*out", Tenant: "cgrates.org", Category: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr2.CgrId = utils.Sha1(strCdr2.AccId, strCdr2.SetupTime.String())
	strCdr3 := &utils.StoredCdr{TOR: utils.VOICE, AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: "rated",
		Direction: "*out", Tenant: "itsyscom.com", Category: "call", Account: "1002", Subject: "1002", Destination: "+4986517174964",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: "wholesale_run", Cost: 1.201}
	strCdr3.CgrId = utils.Sha1(strCdr3.AccId, strCdr3.SetupTime.String())

	for _, cdr := range []*utils.StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysqlDb.SetRatedCdr(cdr, ""); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestGetStoredCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var timeStart, timeEnd time.Time
	// All CDRs, no filter
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cgrids
	if storedCdrs, err := mysqlDb.GetStoredCdrs([]string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cgrids plus reqType
	if storedCdrs, err := mysqlDb.GetStoredCdrs([]string{utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String()),
		utils.Sha1("bbb2", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())},
		nil, nil, nil, nil, []string{"prepaid"}, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on runId
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, []string{utils.DEFAULT_RUNID},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on TOR
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, []string{utils.SMS},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple TOR
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, []string{utils.SMS, utils.VOICE},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrHost
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, []string{"192.168.1.2"},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple cdrHost
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, []string{"192.168.1.1", "192.168.1.2"}, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrSource
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, []string{"UNKNOWN"},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple cdrSource
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, []string{"UNKNOWN", "UNKNOWN2"},
		nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on reqType
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, []string{"prepaid"},
		nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple reqType
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, []string{"prepaid", "pseudoprepaid"}, nil, nil, nil, nil, nil, nil, nil, nil,
		0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on direction
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, []string{"*out"}, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tenant
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, []string{"itsyscom.com"}, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple tenants
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, []string{"itsyscom.com", "cgrates.org"}, nil, nil, nil, nil, nil, nil,
		0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tor
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, []string{"premium_call"},
		nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple tor
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, []string{"premium_call", "call"},
		nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on account
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"1002"},
		nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple account
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"1001", "1002"},
		nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on subject
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"1000"},
		nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple subject
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"1000", "1002"},
		nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on destPrefix
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, []string{"+498651"}, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on multiple destPrefixes
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, []string{"1001", "+498651"}, nil, nil,
		0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 4 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreErr
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, true, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreRated
	var orderIdStart, orderIdEnd int64 // Capture also orderIds for the next test
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, true, false); err != nil {
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
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, orderIdStart, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on orderIdStart and orderIdEnd
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, orderIdStart, orderIdEnd+1, timeStart, timeEnd,
		false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart
	timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Combined filter
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, []string{"rated"}, nil, nil, nil, nil, nil,
		nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreDerived
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, true); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 { // ToDo: Recheck this value
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

func TestCallCost(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrId := utils.Sha1("bbb1", "123")
	cc := &CallCost{
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
	if err := mysqlDb.LogCallCost(cgrId, TEST_SQL, TEST_SQL, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mysqlDb.GetCallCostLog(cgrId, TEST_SQL, TEST_SQL); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
}

func TestRemStoredCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var timeStart, timeEnd time.Time
	cgrIdB1 := utils.Sha1("bbb1", time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC).String())
	if err := mysqlDb.RemStoredCdrs([]string{cgrIdB1}); err != nil {
		t.Error(err.Error())
	}
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
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
	if storedCdrs, err := mysqlDb.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, timeStart, timeEnd, false, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 0 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

func TestSetGetTPActionTriggers(t *testing.T) {
	if !*testLocal {
		return
	}
	atrg := &utils.TPActionTrigger{
		BalanceType:    "*monetary",
		Direction:      "*out",
		ThresholdType:  "*min_balance",
		ThresholdValue: 2.0,
		Recurrent:      true,
		DestinationId:  "*any",
		Weight:         10.0,
		ActionsId:      "LOG_BALANCE",
	}
	mpAtrgs := map[string][]*utils.TPActionTrigger{TEST_SQL: []*utils.TPActionTrigger{atrg}}
	if err := mysqlDb.SetTPActionTriggers(TEST_SQL+"1", mpAtrgs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if rcvMpAtrgs, err := mysqlDb.GetTpActionTriggers(TEST_SQL+"1", TEST_SQL); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if !reflect.DeepEqual(mpAtrgs, rcvMpAtrgs) {
		t.Errorf("Expecting: %v, received: %v", mpAtrgs, rcvMpAtrgs)
	}
}
