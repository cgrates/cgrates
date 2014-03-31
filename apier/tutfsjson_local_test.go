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

var waitFs = flag.Int("wait_fs", 500, "Number of miliseconds to wait for FreeSWITCH to start")

func init() {
	fsjsonCfgPath = path.Join(*dataDir, "tutorials", "fs_json", "cgrates", "etc", "cgrates", "cgrates.cfg")
	fsjsonCfg, _ = config.NewCGRConfigFromFile(&fsjsonCfgPath)
}

// Remove here so they can be properly created by init script
func TestFsJsonRemoveDirs(t *testing.T) {
	if !*testLocal {
		return
	}
	for _, pathDir := range []string{fsjsonCfg.CdreDir, fsjsonCfg.HistoryDir} {
		if err := os.RemoveAll(pathDir); err != nil {
			t.Fatal("Error removing folder: ", pathDir, err)
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

func TestFsJsonStartFs(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "freeswitch").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	go func() {
		fs := exec.Command("/usr/share/cgrates/tutorials/fs_json/freeswitch/etc/init.d/freeswitch", "start")
		out, _ := fs.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsJson: %s", out))
	}()
	time.Sleep(time.Duration(*waitFs) * time.Millisecond) // Give time to rater to fire up
}

// Finds cgr-engine executable and starts it with default configuration
func TestFsJsonStartEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	exec.Command("pkill", "cgr-engine").Run() // Just to make sure another one is not running, bit brutal maybe we can fine tune it
	go func() {
		eng := exec.Command("/usr/share/cgrates/tutorials/fs_json/cgrates/etc/init.d/cgrates", "start")
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
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tutorials", "fs_json", "cgrates", "tariffplans")}
	if err := rater.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.LoadTariffPlanFromFolder: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling ApierV1.LoadTariffPlanFromFolder got reply: ", reply)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 2, RatingProfiles: 2, Actions: 5, SharedGroups: 1, RatingAliases: 1, AccountAliases: 1}
	var args utils.AttrCacheStats
	if err := rater.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %v, received: %v", expectedStats, rcvStats)
	}
}

func TestFsJsonGetAccount1001(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 2 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blncLst := acnt.BalanceMap[attrs.BalanceType+attrs.Direction]
	for _, blnc := range blncLst {
		if len(blnc.SharedGroup) == 0 && blnc.Value != 5 {
			t.Errorf("Unexpected value for general balance: %f", blnc.Value)
		} else if blnc.SharedGroup == "SHARED_A" && blnc.Value != 5 {
			t.Errorf("Unexpected value for shared balance: %f", blnc.Value)
		}
	}
}

func TestFsJsonGetAccount1002(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1002", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 1 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blnc := acnt.BalanceMap[attrs.BalanceType+attrs.Direction][0]
	if blnc.Value != 10 {
		t.Errorf("Unexpected value for general balance: %f", blnc.Value)
	}
}

func TestFsJsonGetAccount1003(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1003", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 1 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blnc := acnt.BalanceMap[attrs.BalanceType+attrs.Direction][0]
	if blnc.Value != 10 {
		t.Errorf("Unexpected value for general balance: %f", blnc.Value)
	}
}

func TestFsJsonGetAccount1004(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1004", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 1 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blnc := acnt.BalanceMap[attrs.BalanceType+attrs.Direction][0]
	if blnc.Value != 10 {
		t.Errorf("Unexpected value for general balance: %f", blnc.Value)
	}
}

func TestFsJsonGetAccount1006(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1006", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err == nil {
		t.Error("Got no error when querying unexisting balance")
	}
}

func TestFsJsonGetAccount1007(t *testing.T) {
	if !*testLocal {
		return
	}
	var acnt *engine.Account
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1007", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 1 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blncLst := acnt.BalanceMap[attrs.BalanceType+attrs.Direction]
	for _, blnc := range blncLst {
		if len(blnc.SharedGroup) == 0 && blnc.Value != 0 { // General balance
			t.Errorf("Unexpected value for general balance: %f", blnc.Value)
		} else if blnc.SharedGroup == "SHARED_A" && blnc.Value != 0 {
			t.Errorf("Unexpected value for shared balance: %f", blnc.Value)
		}
	}
}

func TestMaxCallDuration(t *testing.T) {
	if !*testLocal {
		return
	}
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(fsjsonCfg.SMMaxCallDuration),
	}
	var remainingDurationFloat float64
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(90)*time.Minute {
			t.Errorf("Expecting maxSessionTime around 1h30m, received as: %v", remainingDuration)
		}
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1001",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(45)*time.Minute {
			t.Errorf("Expecting maxSessionTime around 45m, received as: %v", remainingDuration)
		}
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1006",
		Account:     "1006",
		Destination: "1001",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(45)*time.Minute {
			t.Errorf("Expecting maxSessionTime around 45m, received as: %v", remainingDuration)
		}
	}
	// 1007 should use the 1001 balance when doing maxSessionTime
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1007",
		Account:     "1007",
		Destination: "1001",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(20)*time.Minute {
			t.Errorf("Expecting maxSessionTime around 20m, received as: %v", remainingDuration)
		}
	}
}

func TestMaxDebit1001(t *testing.T) {
	if !*testLocal {
		return
	}
	cc := &engine.CallCost{}
	var acnt *engine.Account
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(time.Duration(10) * time.Second),
	}
	if err := rater.Call("Responder.MaxDebit", cd, cc); err != nil {
		t.Error(err.Error())
	} else if cc.GetDuration() > time.Duration(1)*time.Minute {
		t.Errorf("Unexpected call duration received: %v", cc.GetDuration())
	}
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else {
		if len(acnt.BalanceMap["*monetary*out"]) != 2 {
			t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap["*monetary*out"]))
		}
		blncLst := acnt.BalanceMap["*monetary*out"]
		for _, blnc := range blncLst {
			if blnc.SharedGroup == "SHARED_A" && blnc.Value != 5 {
				t.Errorf("Unexpected value for shared balance: %f", blnc.Value)
			} else if len(blnc.SharedGroup) == 0 && blnc.Value != 4.4 {
				t.Errorf("Unexpected value for general balance: %f", blnc.Value)
			}
		}
	}
}

func TestMaxDebit1007(t *testing.T) {
	if !*testLocal {
		return
	}
	cc := &engine.CallCost{}
	var acnt *engine.Account
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1007",
		Account:     "1007",
		Destination: "1002",
		TimeStart:   time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC),
		TimeEnd:     time.Date(2014, 3, 27, 10, 42, 26, 0, time.UTC).Add(time.Duration(10) * time.Second),
	}
	if err := rater.Call("Responder.MaxDebit", cd, cc); err != nil {
		t.Error(err.Error())
	} else if cc.GetDuration() > time.Duration(1)*time.Minute {
		t.Errorf("Unexpected call duration received: %v", cc.GetDuration())
	}
	// Debit out of shared balance should reflect in the 1001 instead of 1007
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else {
		if len(acnt.BalanceMap["*monetary*out"]) != 2 {
			t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap["*monetary*out"]))
		}
		blncLst := acnt.BalanceMap["*monetary*out"]
		for _, blnc := range blncLst {
			if blnc.SharedGroup == "SHARED_A" && blnc.Value != 4 {
				t.Errorf("Unexpected value for shared balance: %f", blnc.Value)
			} else if len(blnc.SharedGroup) == 0 && blnc.Value != 4.4 {
				t.Errorf("Unexpected value for general balance: %f", blnc.Value)
			}
		}
	}
	// Make sure 1007 remains the same
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1007", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	}
	if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 0, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	if len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]) != 1 {
		t.Errorf("Unexpected number of balances found: %d", len(acnt.BalanceMap[attrs.BalanceType+attrs.Direction]))
	}
	blnc := acnt.BalanceMap[attrs.BalanceType+attrs.Direction][0]
	if len(blnc.SharedGroup) == 0 { // General balance
		t.Errorf("Unexpected general balance: %f", blnc.Value)
	} else if blnc.SharedGroup == "SHARED_A" && blnc.Value != 0 {
		t.Errorf("Unexpected value for shared balance: %f", blnc.Value)
	}
}

// Simply kill the engine after we are done with tests within this file
func TestFsJsonStopEngine(t *testing.T) {
	if !*testLocal {
		return
	}
	go func() {
		eng := exec.Command("/usr/share/cgrates/tutorials/fs_json/cgrates/etc/init.d/cgrates", "stop")
		out, _ := eng.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsJson: %s", out))
	}()
}

func TestFsJsonStopFs(t *testing.T) {
	if !*testLocal {
		return
	}
	go func() {
		fs := exec.Command("/usr/share/cgrates/tutorials/fs_json/freeswitch/etc/init.d/freeswitch", "stop")
		out, _ := fs.CombinedOutput()
		engine.Logger.Info(fmt.Sprintf("CgrEngine-TestFsJson: %s", out))
	}()
}
