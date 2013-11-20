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

package apier

import (
	"flag"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/cgrates/engine"
	"os/exec"
	"path"
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
	"reflect"
	"testing"
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
var rater *rpc.Client

var testLocal = flag.Bool("local", false, "Perform the tests only on local test environment, not by default.") // This flag will be passed here via "go test -local" args
var dataDir = flag.String("data_dir", "/usr/share/cgrates/data", "CGR data dir path here")
var waitRater = flag.Int("wait_rater", 200, "Number of miliseconds to wait for rater to start and cache")

func TestCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	cfg, _ = config.NewDefaultCGRConfig()
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = d.(*engine.MySQLStorage)
	}
	for _, scriptName := range []string{engine.CREATE_CDRS_TABLES_SQL, engine.CREATE_COSTDETAILS_TABLES_SQL, engine.CREATE_MEDIATOR_TABLES_SQL, engine.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", "mysql", scriptName)); err != nil {
			t.Fatal("Error on mysql creation: ", err.Error())
			return // No point in going further
		}
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
	engine := exec.Command(enginePath, "-rater", "-config", path.Join(*dataDir,"conf","cgrates.cfg"))
	if err := engine.Start(); err != nil {
		t.Fatal("Cannot start cgr-engine: ", err.Error())
	}
	time.Sleep(time.Duration(*waitRater)*time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	if cfg.RPCEncoding == utils.JSON {
		rater, err = jsonrpc.Dial("tcp", cfg.MediatorRater)
	} else {
		rater, err = rpc.Dial("tcp", cfg.MediatorRater)
	}
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Test here TPTiming APIs
func TestApierTPTiming(t *testing.T) {
	if !*testLocal {
		return
	}
	// ALWAYS,*any,*any,*any,*any,00:00:00
	tmAlways := &utils.ApierTPTiming{TPid: engine.TEST_SQL,
					TimingId: "ALWAYS",
					Years: "*any",
					Months: "*any",
					MonthDays: "*any",
					WeekDays: "*any",
					Time: "00:00:00",
					}
	tmAlways2 := new(utils.ApierTPTiming)
	*tmAlways2 = *tmAlways
	tmAlways2.TimingId = "ALWAYS2"
	tmAsap := &utils.ApierTPTiming{TPid: engine.TEST_SQL,
					TimingId: "ASAP",
					Years: "*any",
					Months: "*any",
					MonthDays: "*any",
					WeekDays: "*any",
					Time: "*asap",
					}
	reply := ""
	for _, tm := range []*utils.ApierTPTiming{tmAlways, tmAsap, tmAlways2} {
		if err := rater.Call("ApierV1.SetTPTiming", tm, &reply); err!=nil { 
			t.Error("Got error on ApierV1.SetTPTiming: ", err.Error())
		} else if reply != "OK" {
			t.Error("Unexpected reply received when calling ApierV1.SetTPTiming: ", reply)
		}
	}
	// Check second set
	if err := rater.Call("ApierV1.SetTPTiming", tmAlways, &reply); err!=nil { 
		t.Error("Got error on second ApierV1.SetTPTiming: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.SetTPTiming got reply: ", reply)
	}
	// Check missing params
	if err := rater.Call("ApierV1.SetTPTiming", new(utils.ApierTPTiming), &reply); err==nil {
		t.Error("Calling ApierV1.SetTPTiming, expected error, received: ", reply)
	} else if err.Error() != "MANDATORY_IE_MISSING:[TPid TimingId Years Months MonthDays WeekDays Time]" { 
		t.Error("Calling ApierV1.SetTPTiming got unexpected error: ", err.Error())
	}
	// Test get
	var rplyTmAlways2 *utils.ApierTPTiming
	if err := rater.Call("ApierV1.GetTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.TimingId}, &rplyTmAlways2); err!=nil { 
		t.Error("Calling ApierV1.GetTPTiming, got error: ", err.Error())
	} else if !reflect.DeepEqual(tmAlways2, rplyTmAlways2)  {
		t.Errorf("Calling ApierV1.GetTPTiming expected: %v, received: %v", tmAlways, rplyTmAlways2)
	}
	// Test remove
	if err := rater.Call("ApierV1.RemTPTiming", AttrGetTPTiming{tmAlways2.TPid, tmAlways2.TimingId}, &reply); err!=nil { 
		t.Error("Calling ApierV1.RemTPTiming, got error: ", err.Error())
	} else if reply != "OK"  {
		t.Error("Calling ApierV1.RemTPTiming received: ", reply)
	}
	// Test getIds
	var rplyTmIds []string
	expectedTmIds := []string{"ALWAYS", "ASAP"}
	if err := rater.Call("ApierV1.GetTPTimingIds", AttrGetTPTimingIds{tmAlways.TPid}, &rplyTmIds); err!=nil { 
		t.Error("Calling ApierV1.GetTPTimingIds, got error: ", err.Error())
	} else if !reflect.DeepEqual(expectedTmIds, rplyTmIds)  {
		t.Errorf("Calling ApierV1.GetTPTimingIds expected: %v, received: %v", expectedTmIds, rplyTmIds)
	}
}
