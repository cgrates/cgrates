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
	"fmt"
	"net/rpc"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Empty tables before using them
func TestFsCsvCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = d.(*engine.MySQLStorage)
	}
	for _, scriptName := range []string{engine.CREATE_CDRS_TABLES_SQL, engine.CREATE_COSTDETAILS_TABLES_SQL, engine.CREATE_MEDIATOR_TABLES_SQL, engine.CREATE_TARIFFPLAN_TABLES_SQL} {
		if err := mysql.CreateTablesFromScript(path.Join(*dataDir, "storage", *storDbType, scriptName)); err != nil {
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

func TestFsCsvInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	ratingDb, err := engine.ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort, cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	accountDb, err := engine.ConfigureAccountingStorage(cfg.AccountDBType, cfg.AccountDBHost, cfg.AccountDBPort, cfg.AccountDBName,
		cfg.AccountDBUser, cfg.AccountDBPass, cfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	for _, db := range []engine.Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			t.Fatal("Cannot reset dataDb", err)
		}
	}
}

// Finds cgr-engine executable and starts it with default configuration
func TestFsCsvStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	go func() {
		eng := exec.Command(enginePath, "-rater", "-scheduler", "-cdrs", "-mediator", "-config", path.Join(*dataDir, "conf", "cgrates.cfg"))
		out, _ := eng.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsCsv: %s", out))
	}()
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestFsCsvRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	rater, err = rpc.Dial("tcp", "127.0.0.1:2013") // ToDo: Fix here with config loaded from file
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Make sure we start with fresh data
func TestFsCsvEmptyCache(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	arc := new(utils.ApiReloadCache)
	// Simple test that command is executed without errors
	if err := rater.Call("ApierV1.ReloadCache", arc, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 0, RatingPlans: 0, RatingProfiles: 0, Actions: 0}
	var args utils.AttrCacheStats
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestFsCsvLoadTariffPlans(t *testing.T) {
	if !*testLocal {
		return
	}
	reply := ""
	// Simple test that command is executed without errors
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tutorials", "fs_csv", "cgrates", "tariffplans")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 1, RatingProfiles: 1, Actions: 1}
	var args utils.AttrCacheStats
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestFsCsvCall1(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart := time.Date(2014, 01, 15, 6, 0, 0, 0, time.UTC)
	tEnd := time.Date(2014, 01, 15, 6, 0, 35, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:    "*out",
		TOR:          "call",
		Tenant:       "cgrates.org",
		Subject:      "1001",
		Account:      "1001",
		Destination:  "1002",
		TimeStart:    tStart,
		TimeEnd:      tEnd,
		CallDuration: 35,
	}
	var cc engine.CallCost
	// Simple test that command is executed without errors
	if err := rater.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetConnectFee() != 0.4 && cc.Cost != 0.6 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
}

// Simply kill the engine after we are done with tests within this file
func TestFsCsvStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run()
}
