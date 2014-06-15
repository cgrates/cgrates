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

package mediator

import (
	"flag"
	"fmt"
	"net/http"
	"net/rpc"
	"net/url"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
README:

 Enable local tests by passing '-local' to the go test command
 It is expected that the data folder of CGRateS exists at path /usr/share/cgrates/data or passed via command arguments.
 Prior running the tests, create database and users by running:
  mysql -pyourrootpwd < /usr/share/cgrates/data/storage/mysql/create_db_with_users.sql
 What these tests do:
  * Flush tables in storDb to start clean.
  * Start engine with default configuration and give it some time to listen (here caching can slow down, hence the command argument parameter).
  * Connect rpc client depending on encoding defined in configuration.
  * Execute remote Apis and test their replies(follow prepaid1cent scenario so we can test load in dataDb also).
*/

var cfg *config.CGRConfig
var cgrRpc *rpc.Client
var cdrStor engine.CdrStorage
var httpClient *http.Client

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
var storDbType = flag.String("stordb_type", utils.MYSQL, "The type of the storDb database <mysql>")
var startDelay = flag.Int("delay_start", 300, "Number of miliseconds to it for rater to start and cache")
var cfgPath = path.Join(*dataDir, "conf", "samples", "mediator_test1.cfg")

func TestInitRatingDb(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cfg, err = config.NewCGRConfigFromFile(&cfgPath)
	if err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	ratingDb, err := engine.ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort, cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	if err := ratingDb.Flush(); err != nil {
		t.Fatal("Cannot reset dataDb", err)
	}
}

// Empty tables before using them
func TestInitStorDb(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	var err error
	if cdrStor, err = engine.ConfigureCdrStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = cdrStor.(*engine.MySQLStorage)
	}
	if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", *storDbType, engine.CREATE_CDRS_TABLES_SQL)); err != nil {
		t.Fatal("Error on mysql creation: ", err.Error())
		return // No point in going further
	}
	for _, tbl := range []string{utils.TBL_CDRS_PRIMARY, utils.TBL_CDRS_EXTRA} {
		if _, err := mysql.Db.Query(fmt.Sprintf("SELECT 1 from %s", tbl)); err != nil {
			t.Fatal(err.Error())
		}
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	engine := exec.Command(enginePath, "-config", cfgPath)
	if err := engine.Start(); err != nil {
		t.Fatal("Cannot start cgr-engine: ", err.Error())
	}
	time.Sleep(time.Duration(*startDelay) * time.Millisecond) // Give time to rater to fire up
	httpClient = new(http.Client)
}

// Connect rpc client
func TestRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	cgrRpc, err = rpc.Dial("tcp", cfg.RPCGOBListen) //ToDo: Fix with automatic config
	if err != nil {
		t.Fatal("Could not connect to CGR GOB-RPC Server: ", err.Error())
	}
}

func TestPostCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	cdrForm1 := url.Values{utils.TOR: []string{utils.VOICE}, utils.ACCID: []string{"dsafdsaf"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{"rated"}, utils.DIRECTION: []string{"*out"},
		utils.TENANT: []string{"cgrates.org"}, utils.CATEGORY: []string{"call"}, utils.ACCOUNT: []string{"2001"}, utils.SUBJECT: []string{"2001"},
		utils.DESTINATION: []string{"+4986517174963"},
		utils.ANSWER_TIME: []string{"2013-11-07T08:42:26Z"}, utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrForm2 := url.Values{utils.TOR: []string{utils.VOICE}, utils.ACCID: []string{"adsafdsaf"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{"rated"}, utils.DIRECTION: []string{"*out"},
		utils.TENANT: []string{"itsyscom.com"}, utils.CATEGORY: []string{"call"}, utils.ACCOUNT: []string{"1003"}, utils.SUBJECT: []string{"1003"}, utils.DESTINATION: []string{"+4986517174964"},
		utils.ANSWER_TIME: []string{"2013-11-07T08:42:26Z"}, utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	cdrFormData1 := url.Values{utils.TOR: []string{utils.DATA}, utils.ACCID: []string{"616350843"}, utils.CDRHOST: []string{"192.168.1.1"}, utils.REQTYPE: []string{"rated"},
		utils.DIRECTION: []string{"*out"}, utils.TENANT: []string{"cgrates.org"}, utils.CATEGORY: []string{"data"},
		utils.ACCOUNT: []string{"1010"}, utils.SUBJECT: []string{"1010"}, utils.ANSWER_TIME: []string{"2013-11-07T08:42:26Z"},
		utils.USAGE: []string{"10"}, "field_extr1": []string{"val_extr1"}, "fieldextr2": []string{"valextr2"}}
	for _, cdrForm := range []url.Values{cdrForm1, cdrForm2, cdrFormData1} {
		cdrForm.Set(utils.CDRSOURCE, engine.TEST_SQL)
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cgr", cfg.HTTPListen), cdrForm); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(100 * time.Millisecond) // Give time for CDRs to reach database
	if storedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, false, false, false); err != nil {
		t.Error(err)
	} else if len(storedCdrs) != 6 { // Make sure CDRs made it into StorDb
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(storedCdrs)))
	}
	if nonErrorCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, true, false, false); err != nil {
		t.Error(err)
	} else if len(nonErrorCdrs) != 0 {
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(nonErrorCdrs)))
	}
}

// Directly inject CDRs into storDb
func TestInjectCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	cgrCdr1 := utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "aaaaadsafdsaf", "cdrsource": engine.TEST_SQL, utils.CDRHOST: "192.168.1.1", utils.REQTYPE: "rated", utils.DIRECTION: "*out",
		utils.TENANT: "cgrates.org", utils.CATEGORY: "call", utils.ACCOUNT: "dan", utils.SUBJECT: "dan", utils.DESTINATION: "+4986517174963",
		utils.ANSWER_TIME: "2013-11-07T08:42:26Z", utils.USAGE: "10"}
	cgrCdr2 := utils.CgrCdr{utils.TOR: utils.VOICE, utils.ACCID: "baaaadsafdsaf", "cdrsource": engine.TEST_SQL, utils.CDRHOST: "192.168.1.1", utils.REQTYPE: "rated", utils.DIRECTION: "*out",
		utils.TENANT: "cgrates.org", utils.CATEGORY: "call", utils.ACCOUNT: "dan", utils.SUBJECT: "dan", utils.DESTINATION: "+4986517173964",
		utils.ANSWER_TIME: "2013-11-07T09:42:26Z", utils.USAGE: "20"}
	for _, cdr := range []utils.CgrCdr{cgrCdr1, cgrCdr2} {
		if err := cdrStor.SetCdr(cdr.AsStoredCdr()); err != nil {
			t.Error(err)
		}
	}
	if storedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, false, false, false); err != nil {
		t.Error(err)
	} else if len(storedCdrs) != 8 { // Make sure CDRs made it into StorDb
		t.Error(fmt.Sprintf("Unexpected number of CDRs stored: %d", len(storedCdrs)))
	}
	if nonRatedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, true, true, false); err != nil {
		t.Error(err)
	} else if len(nonRatedCdrs) != 2 { // Just two of them should be non-rated
		t.Error(fmt.Sprintf("Unexpected number of CDRs non-rated: %d", len(nonRatedCdrs)))
	}
}

// Test here LoadTariffPlanFromFolder
func TestLoadTariffPlanFromFolder(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "prepaid1centpsec")}
	if err := cgrRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
}

func TestRateCdrs(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply string
	if err := cgrRpc.Call("MediatorV1.RateCdrs", utils.AttrRateCdrs{}, &reply); err != nil {
		t.Error(err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply: %s", reply)
	}
	if nonRatedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, true, true, false); err != nil {
		t.Error(err)
	} else if len(nonRatedCdrs) != 0 { // All CDRs should be rated
		t.Error(fmt.Sprintf("Unexpected number of CDRs non-rated: %d", len(nonRatedCdrs)))
	}
	if errRatedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, false, true, false); err != nil {
		t.Error(err)
	} else if len(errRatedCdrs) != 8 { // The first 2 with errors should be still there before rerating
		t.Error(fmt.Sprintf("Unexpected number of CDRs with errors: %d", len(errRatedCdrs)))
	}
	if err := cgrRpc.Call("MediatorV1.RateCdrs", utils.AttrRateCdrs{RerateErrors: true}, &reply); err != nil {
		t.Error(err.Error())
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply: %s", reply)
	}
	if errRatedCdrs, err := cdrStor.GetStoredCdrs(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, 0, 0, time.Time{}, time.Time{}, false, true, false); err != nil {
		t.Error(err)
	} else if len(errRatedCdrs) != 4 {
		t.Error(fmt.Sprintf("Unexpected number of CDRs with errors: %d", len(errRatedCdrs)))
	}
}

/*
func TestMediatePseudoprepaid(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003", Direction: "*out"}
	if err := cgrRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue() != 11 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	}
	voiceCdr := &utils.StoredCdr{TOR: utils.VOICE, AccId: "dsafdsaf", CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.PSEUDOPREPAID, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "call", Account: "1003", Subject: "1003", Destination: "+4986517174963",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(5) * time.Second}
	dataCdr := &utils.StoredCdr{TOR: utils.DATA, AccId: "6163508432", CdrHost: "192.168.1.1", CdrSource: "test", ReqType: utils.PSEUDOPREPAID, Direction: utils.OUT,
		Tenant: "cgrates.org", Category: "data", Account: "1003", Subject: "1003",
		SetupTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC), AnswerTime: time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		Usage: time.Duration(10) * time.Second}
	for _, cdrForm := range []url.Values{voiceCdr.AsHttpForm(), dataCdr.AsHttpForm()} {
		cdrForm.Set(utils.CDRSOURCE, engine.TEST_SQL)
		if _, err := httpClient.PostForm(fmt.Sprintf("http://%s/cgr", cfg.HTTPListen), cdrForm); err != nil {
			t.Error(err.Error())
		}
	}
	time.Sleep(time.Duration(*startDelay) * time.Millisecond) // Give time for debits to happen
	expectBalance := 5.998
	if err := cgrRpc.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue() != expectBalance { // 5 from voice, 0.002 from DATA
		t.Errorf("Calling ApierV1.GetBalance expected: %f, received: %f", expectBalance, reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	}
}
*/

// Simply kill the engine after we are done with tests within this file
func TestStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run()
}
