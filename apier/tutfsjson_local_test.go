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
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var fsjsonCfgPath string
var fsjsonCfg *config.CGRConfig

func init() {
	fsjsonCfgPath = path.Join(*dataDir, "tutorials", "fs_json", "cgrates", "etc", "cgrates", "cgrates.cfg")
	fsjsonCfg, _ = config.NewCGRConfig(&fsjsonCfgPath)
}

func TestFsJsonCreateDirs(t *testing.T) {
	if !*testLocal {
		return
	}
	for _, pathDir := range []string{cfg.CdreDir, cfg.HistoryDir} {
		if err := os.RemoveAll(pathDir); err != nil {
			t.Fatal("Error removing folder: ", pathDir, err)
		}
		if err := os.MkdirAll(pathDir, 0755); err != nil {
			t.Fatal("Error creating folder: ", pathDir, err)
		}
	}
}

// Empty tables before using them
func TestFsJsonCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(fsjsonCfg.StorDBHost, fsjsonCfg.StorDBPort, fsjsonCfg.StorDBName, fsjsonCfg.StorDBUser, fsjsonCfg.StorDBPass); err != nil {
		t.Fatal("Error on opening database connection: ", err)
	} else {
		mysql = d.(*engine.MySQLStorage)
	}
	for _, scriptName := range []string{engine.CREATE_CDRS_TABLES_SQL, engine.CREATE_COSTDETAILS_TABLES_SQL, engine.CREATE_MEDIATOR_TABLES_SQL} {
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

func TestFsJsonInitDataDb(t *testing.T) {
	if !*testLocal {
		return
	}
	ratingDb, err := engine.ConfigureRatingStorage(fsjsonCfg.RatingDBType, fsjsonCfg.RatingDBHost, fsjsonCfg.RatingDBPort, fsjsonCfg.RatingDBName, fsjsonCfg.RatingDBUser, fsjsonCfg.RatingDBPass, fsjsonCfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	accountDb, err := engine.ConfigureAccountingStorage(fsjsonCfg.AccountDBType, fsjsonCfg.AccountDBHost, fsjsonCfg.AccountDBPort, fsjsonCfg.AccountDBName,
		fsjsonCfg.AccountDBUser, fsjsonCfg.AccountDBPass, fsjsonCfg.DBDataEncoding)
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
func TestFsJsonStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("Cannot find cgr-engine executable")
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	go func() {
		eng := exec.Command(enginePath, "-config", fsjsonCfgPath)
		out, _ := eng.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsJson: %s", out))
	}()
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time to rater to fire up
}

// Connect rpc client to rater
func TestFsJsonRpcConn(t *testing.T) {
	if !*testLocal {
		return
	}
	var err error
	rater, err = jsonrpc.Dial("tcp", fsjsonCfg.RPCJSONListen)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Make sure we start with fresh data
func TestFsJsonEmptyCache(t *testing.T) {
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

func TestFsJsonLoadTariffPlans(t *testing.T) {
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
	time.Sleep(100 * time.Millisecond) // Give time for scheduler to execute topups
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 1, RatingProfiles: 1, Actions: 2}
	var args utils.AttrCacheStats
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestFsJsonGetAccount(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
}

// Simply kill the engine after we are done with tests within this file
func TestFsJsonStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run()
}
