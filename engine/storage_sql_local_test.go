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
	"testing"
)

/*
README:

 Enable these tests by passing '-local' to the go test command

 Only database supported for now is MySQL. Postgres could be easily extended if needed.

 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data.

 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_mysql_with_users.sql
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
		Direction: "*out", ActionTimingsId: "PREPAID_10", ActionTriggersId: "STANDARD_TRIGGERS"}
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
