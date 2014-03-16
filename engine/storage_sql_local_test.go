/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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

 Only database supported for now is MySQL. Postgres could be easily extended if needed.

 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data.

 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/storage/mysql/create_mysql_with_users.sql
*/

var mysql *MySQLStorage

func TestCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrConfig, _ := config.NewDefaultCGRConfig()
	if d, err := NewMySQLStorage(cgrConfig.StorDBHost, cgrConfig.StorDBPort, cgrConfig.StorDBName, cgrConfig.StorDBUser, cgrConfig.StorDBPass); err != nil {
		t.Error("Error on opening database connection: ", err)
		return
	} else {
		mysql = d.(*MySQLStorage)
	}
	for _, scriptName := range []string{CREATE_CDRS_TABLES_SQL, CREATE_COSTDETAILS_TABLES_SQL, CREATE_MEDIATOR_TABLES_SQL, CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", scriptName)); err != nil {
			t.Error("Error on mysql creation: ", err.Error())
			return // No point in going further
		}
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysql.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
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
	if err := mysql.SetTPTiming(TEST_SQL, tm); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysql.GetTpTimings(TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) == 0 {
		t.Error("Could not store TPTiming")
	}
	// Remove Timings
	if err := mysql.RemTPData(utils.TBL_TP_TIMINGS, TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	}
	if tmgs, err := mysql.GetTpTimings(TEST_SQL, tm.Id); err != nil {
		t.Error(err.Error())
	} else if len(tmgs) != 0 {
		t.Error("Did not remove TPTiming")
	}
	// Create RatingProfile
	ras := []*utils.TPRatingActivation{&utils.TPRatingActivation{ActivationTime: "2012-01-01T00:00:00Z", RatingPlanId: "RETAIL1"}}
	rp := &utils.TPRatingProfile{TPid: TEST_SQL, LoadId: TEST_SQL, Tenant: "cgrates.org", TOR: "call", Direction: "*out", Subject: "*any", RatingPlanActivations: ras}
	if err := mysql.SetTPRatingProfiles(TEST_SQL, map[string]*utils.TPRatingProfile{rp.KeyId(): rp}); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysql.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if len(rps) == 0 {
		t.Error("Could not store TPRatingProfile")
	}
	// Remove RatingProfile
	if err := mysql.RemTPData(utils.TBL_TP_RATE_PROFILES, rp.TPid, rp.LoadId, rp.Tenant, rp.TOR, rp.Direction, rp.Subject); err != nil {
		t.Error(err.Error())
	}
	if rps, err := mysql.GetTpRatingProfiles(rp); err != nil {
		t.Error(err.Error())
	} else if len(rps) != 0 {
		t.Error("Did not remove TPRatingProfile")
	}

	// Create AccountActions
	aa := &utils.TPAccountActions{TPid: TEST_SQL, LoadId: TEST_SQL, Tenant: "cgrates.org", Account: "1001",
		Direction: "*out", ActionPlanId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
	if err := mysql.SetTPAccountActions(aa.TPid, map[string]*utils.TPAccountActions{aa.KeyId(): aa}); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysql.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if len(aas) == 0 {
		t.Error("Could not create TPAccountActions")
	}
	// Remove AccountActions
	if err := mysql.RemTPData(utils.TBL_TP_ACCOUNT_ACTIONS, aa.TPid, aa.LoadId, aa.Tenant, aa.Account, aa.Direction); err != nil {
		t.Error(err.Error())
	}
	if aas, err := mysql.GetTpAccountActions(aa); err != nil {
		t.Error(err.Error())
	} else if len(aas) != 0 {
		t.Error("Did not remove TPAccountActions")
	}
}

func TestSetCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := &utils.CgrCdr{"accid": "aaa1", "cdrhost": "192.168.1.1", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-08T08:42:20Z", "answer_time": "2013-11-08T08:42:26Z", "duration": "10s",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}
	cgrCdr2 := &utils.CgrCdr{"accid": "aaa2", "cdrhost": "192.168.1.1", "reqtype": "prepaid", "direction": "*out", "tenant": "cgrates.org", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "1002", "setup_time": "2013-11-08T08:42:22Z", "answer_time": "2013-11-08T08:42:26Z", "duration": "20",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}
	cgrCdr3 := &utils.CgrCdr{"accid": "aaa3", "cdrhost": "192.168.1.1", "reqtype": "rated", "direction": "*out", "tenant": "cgrates.org", "tor": "premium_call",
		"account": "1002", "subject": "1002", "destination": "1001", "setup_time": "2013-11-07T08:42:24Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "60s",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}
	cgrCdr4 := &utils.CgrCdr{"accid": "aaa4", "cdrhost": "192.168.1.2", "reqtype": "pseudoprepaid", "direction": "*out", "tenant": "itsyscom.com", "tor": "call",
		"account": "1001", "subject": "1001", "destination": "+4986517174964", "setup_time": "2013-11-07T08:42:21Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "1m2s",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}
	cgrCdr5 := &utils.CgrCdr{"accid": "aaa5", "cdrhost": "192.168.1.2", "reqtype": "postpaid", "direction": "*out", "tenant": "itsyscom.com", "tor": "call",
		"account": "1002", "subject": "1002", "destination": "+4986517174963", "setup_time": "2013-11-07T08:42:25Z", "answer_time": "2013-11-07T08:42:26Z", "duration": "15s",
		"field_extr1": "val_extr1", "fieldextr2": "valextr2", "cdrsource": TEST_SQL}
	for _, cdr := range []*utils.CgrCdr{cgrCdr1, cgrCdr2, cgrCdr3, cgrCdr4, cgrCdr5} {
		if err := mysql.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
	strCdr1 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb1"), AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr2 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb2"), AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: TEST_SQL, ReqType: "prepaid",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr3 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb3"), AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: "rated",
		Direction: "*out", Tenant: "itsyscom.com", TOR: "call", Account: "1002", Subject: "1000", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}

	for _, cdr := range []*utils.StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysql.SetCdr(cdr); err != nil {
			t.Error(err.Error())
		}
	}
}

func TestSetRatedCdr(t *testing.T) {
	if !*testLocal {
		return
	}
	strCdr1 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb1"), AccId: "bbb1", CdrHost: "192.168.1.1", CdrSource: "UNKNOWN", ReqType: "rated",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 1.201}
	strCdr2 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb2"), AccId: "bbb2", CdrHost: "192.168.1.2", CdrSource: TEST_SQL, ReqType: "prepaid",
		Direction: "*out", Tenant: "cgrates.org", TOR: "call", Account: "1001", Subject: "1001", Destination: "1002",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(12) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: utils.DEFAULT_RUNID, Cost: 0.201}
	strCdr3 := &utils.StoredCdr{CgrId: utils.FSCgrId("bbb3"), AccId: "bbb3", CdrHost: "192.168.1.1", CdrSource: TEST_SQL, ReqType: "rated",
		Direction: "*out", Tenant: "itsyscom.com", TOR: "call", Account: "1002", Subject: "1002", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 12, 7, 8, 42, 24, 0, time.UTC), AnswerTime: time.Date(2013, 12, 7, 8, 42, 26, 0, time.UTC),
		Duration: time.Duration(10) * time.Second, ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		MediationRunId: "wholesale_run", Cost: 1.201}

	for _, cdr := range []*utils.StoredCdr{strCdr1, strCdr2, strCdr3} {
		if err := mysql.SetRatedCdr(cdr, ""); err != nil {
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
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on runId
	if storedCdrs, err := mysql.GetStoredCdrs(utils.DEFAULT_RUNID, "", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrHost
	if storedCdrs, err := mysql.GetStoredCdrs("", "192.168.1.2", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on cdrSource
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "UNKNOWN", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on reqType
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "prepaid", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on direction
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "*out", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tenant
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "itsyscom.com", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on tor
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "premium_call", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on account
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "1002", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 3 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on subject
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "1000", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreErr
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "", "", timeStart, timeEnd, true, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 8 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on ignoreRated
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, true); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart
	timeStart = time.Date(2013, 11, 8, 8, 0, 0, 0, time.UTC)
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 5 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Filter on timeStart and timeEnd
	timeEnd = time.Date(2013, 12, 1, 8, 0, 0, 0, time.UTC)
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 2 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
	// Combined filter
	if storedCdrs, err := mysql.GetStoredCdrs("", "", "", "rated", "", "", "", "", "", "", timeStart, timeEnd, false, false); err != nil {
		t.Error(err.Error())
	} else if len(storedCdrs) != 1 {
		t.Error("Unexpected number of StoredCdrs returned: ", storedCdrs)
	}
}

func TestCallCost(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrId := utils.FSCgrId("bbb1")
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
	if err := mysql.LogCallCost("bbb1", TEST_SQL, TEST_SQL, cc); err != nil {
		t.Error(err.Error())
	}
	if ccRcv, err := mysql.GetCallCostLog(cgrId, TEST_SQL, TEST_SQL); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(cc, ccRcv) {
		t.Errorf("Expecting call cost: %v, received: %v", cc, ccRcv)
	}
}
