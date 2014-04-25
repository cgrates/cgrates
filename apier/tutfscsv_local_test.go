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

var fscsvCfgPath string
var fscsvCfg *config.CGRConfig

func init() {
	fscsvCfgPath = path.Join(*dataDir, "tutorials", "fs_csv", "cgrates", "etc", "cgrates", "cgrates.cfg")
	fscsvCfg, _ = config.NewCGRConfigFromFile(&fscsvCfgPath)
}

// Remove here so they can be properly created by init script
func TestFsCsvRemoveDirs(t *testing.T) {
	if !*testLocal {
		return
	}
	for _, pathDir := range []string{cfg.CdreDir, cfg.CdrcCdrInDir, cfg.CdrcCdrOutDir, cfg.HistoryDir} {
		if err := os.RemoveAll(pathDir); err != nil {
			t.Fatal("Error removing folder: ", pathDir, err)
		}
	}
}

// Empty tables before using them
func TestFsCsvCreateTables(t *testing.T) {
	if !*testLocal {
		return
	}
	if *storDbType != utils.MYSQL {
		t.Fatal("Unsupported storDbType")
	}
	var mysql *engine.MySQLStorage
	if d, err := engine.NewMySQLStorage(fscsvCfg.StorDBHost, fscsvCfg.StorDBPort, fscsvCfg.StorDBName, fscsvCfg.StorDBUser, fscsvCfg.StorDBPass); err != nil {
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
	ratingDb, err := engine.ConfigureRatingStorage(fscsvCfg.RatingDBType, fscsvCfg.RatingDBHost, fscsvCfg.RatingDBPort, fscsvCfg.RatingDBName, fscsvCfg.RatingDBUser, fscsvCfg.RatingDBPass, fscsvCfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	accountDb, err := engine.ConfigureAccountingStorage(fscsvCfg.AccountDBType, fscsvCfg.AccountDBHost, fscsvCfg.AccountDBPort, fscsvCfg.AccountDBName,
		fscsvCfg.AccountDBUser, fscsvCfg.AccountDBPass, fscsvCfg.DBDataEncoding)
	if err != nil {
		t.Fatal("Cannot connect to dataDb", err)
	}
	for _, db := range []engine.Storage{ratingDb, accountDb} {
		if err := db.Flush(); err != nil {
			t.Fatal("Cannot reset dataDb", err)
		}
	}
}

func TestFsCsvStartFs(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "freeswitch").Run() // Just to make sure no freeswitch is running
	go func() {
		fs := exec.Command("sudo", "/usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch", "start")
		out, _ := fs.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsCsv: %s", out))
	}()
	time.Sleep(time.Duration(*waitFs) * time.Millisecond) // Give time to rater to fire up
}

// Finds cgr-engine executable and starts it with default configuration
func TestFsCsvStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	go func() {
		eng := exec.Command("sudo", "/usr/share/cgrates/tutorials/fs_json/cgrates/etc/init.d/cgrates", "start")
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
	rater, err = jsonrpc.Dial("tcp", fscsvCfg.RPCJSONListen)
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

// Make sure we start with fresh data
func TestFsCsvEmptyCache(t *testing.T) {
	if !*testLocal {
		return
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

func TestFsCsvGetAccount(t *testing.T) {
	if !*testLocal {
		return
	}
	var reply *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue() != 10.0 { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	}
}

func TestFsCsvCall1(t *testing.T) {
	if !*testLocal {
		return
	}
	tStart := time.Date(2014, 01, 15, 6, 0, 0, 0, time.UTC)
	tEnd := time.Date(2014, 01, 15, 6, 0, 35, 0, time.UTC)
	cd := engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		TimeStart:     tStart,
		TimeEnd:       tEnd,
		DurationIndex: 35,
	}
	var cc engine.CallCost
	// Make sure the cost is what we expect it is
	if err := rater.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetConnectFee() != 0.4 && cc.Cost != 0.6 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc)
	}
	// Make sure debit charges what cost returned
	if err := rater.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.MaxDebit: ", err.Error())
	} else if cc.GetConnectFee() != 0.4 && cc.Cost != 0.6 {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc)
	}
	// Make sure the account was debited correctly for the first loop index (ConnectFee included)
	var reply *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue() != 9.4 { // We expect 11.5 since we have added in the previous test 1.5
		t.Errorf("Calling ApierV1.GetAccount expected: 9.4, received: %f", reply.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	} else if len(reply.UnitCounters) != 1 ||
		utils.Round(reply.UnitCounters[0].Balances[0].Value, 2, utils.ROUNDING_MIDDLE) != 0.6 { // Make sure we correctly count usage
		t.Errorf("Received unexpected UnitCounters: %v", reply.UnitCounters)
	}
	cd = engine.CallDescriptor{
		Direction:     "*out",
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Account:       "1001",
		Destination:   "1002",
		TimeStart:     tStart,
		TimeEnd:       tEnd,
		DurationIndex: 35,
		LoopIndex:     1, // Should not charge ConnectFee
	}
	// Make sure debit charges what cost returned
	if err := rater.Call("Responder.MaxDebit", cd, &cc); err != nil {
		t.Error("Got error on Responder.MaxDebit: ", err.Error())
	} else if cc.GetConnectFee() != 0.4 && cc.Cost != 0.2 { // Does not contain connectFee, however connectFee should be correctly reported
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc)
	}
	// Make sure the account was debited correctly for the first loop index (ConnectFee included)
	var reply2 *engine.Account
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply2); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if utils.Round(reply2.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue(), 2, utils.ROUNDING_MIDDLE) != 9.20 {
		t.Errorf("Calling ApierV1.GetAccount expected: 9.2, received: %f", reply2.BalanceMap[engine.CREDIT+attrs.Direction].GetTotalValue())
	} else if len(reply2.UnitCounters) != 1 ||
		utils.Round(reply2.UnitCounters[0].Balances[0].Value, 2, utils.ROUNDING_MIDDLE) != 0.8 { // Make sure we correctly count usage
		t.Errorf("Received unexpected UnitCounters: %v", reply2.UnitCounters)
	}
}

// Simply kill the engine after we are done with tests within this file
func TestFsCsvStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	go func() {
		eng := exec.Command("/usr/share/cgrates/tutorials/fs_csv/cgrates/etc/init.d/cgrates", "stop")
		out, _ := eng.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsCsv: %s", out))
	}()
}

func TestFsCsvStopFs(t *testing.T) {
	if !*testLocal {
		return
	}
	go func() {
		fs := exec.Command("/usr/share/cgrates/tutorials/fs_csv/freeswitch/etc/init.d/freeswitch", "stop")
		out, _ := fs.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsCsv: %s", out))
	}()
}
