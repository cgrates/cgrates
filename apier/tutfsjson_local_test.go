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
	fsjsonCfg, _ = config.NewCGRConfig(&fsjsonCfgPath)
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
		fs := exec.Command("sudo", "/usr/share/cgrates/tutorials/fs_json/freeswitch/etc/init.d/freeswitch", "start")
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
		eng := exec.Command("sudo", "/usr/share/cgrates/tutorials/fs_json/cgrates/etc/init.d/cgrates", "start")
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
	time.Sleep(100 * time.Millisecond) // Give time for scheduler to execute topups
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 3, RatingPlans: 2, RatingProfiles: 2, Actions: 5}
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
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1002", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1003", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1004", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 10.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 10.0, received: %f", reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
	}
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1006", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err == nil {
		t.Error("Account not created and not returning error")
	}
	attrs = &AttrGetAccount{Tenant: "cgrates.org", Account: "1007", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 0.0 {
		t.Errorf("Calling ApierV1.GetBalance expected: 0, received: %f", reply.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
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
		TimeStart:   time.Now(),
		TimeEnd:     time.Now().Add(fsjsonCfg.SMMaxCallDuration),
	}
	var remainingDurationFloat float64
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(3)*time.Hour {
			t.Errorf("Expecting maxSessionTime around 3hs, received as: %v", remainingDuration)
		}
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1002",
		Account:     "1002",
		Destination: "1001",
		TimeStart:   time.Now(),
		TimeEnd:     time.Now().Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(3)*time.Hour {
			t.Errorf("Expecting maxSessionTime around 3hs, received as: %v", remainingDuration)
		}
	}
	cd = engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1006",
		Account:     "1006",
		Destination: "1001",
		TimeStart:   time.Now(),
		TimeEnd:     time.Now().Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(3)*time.Hour {
			t.Errorf("Expecting maxSessionTime around 3hs, received as: %v", remainingDuration)
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
		TimeStart:   time.Now(),
		TimeEnd:     time.Now().Add(fsjsonCfg.SMMaxCallDuration),
	}
	if err := rater.Call("Responder.GetMaxSessionTime", cd, &remainingDurationFloat); err != nil {
		t.Error(err)
	} else {
		remainingDuration := time.Duration(remainingDurationFloat)
		if remainingDuration < time.Duration(3)*time.Hour {
			t.Errorf("Expecting maxSessionTime around 3hs, received as: %v", remainingDuration)
		}
	}
}

func TestMaxDebit(t *testing.T) {
	cc := &engine.CallCost{}
	var acnt *engine.Account
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Tenant:      "cgrates.org",
		TOR:         "call",
		Subject:     "1001",
		Account:     "1001",
		Destination: "1002",
		TimeStart:   time.Now(),
		TimeEnd:     time.Now().Add(time.Duration(10) * time.Second),
	}
	if err := rater.Call("Responder.MaxDebit", cd, cc); err != nil {
		t.Error(err.Error())
	} else if cc.GetDuration() > time.Duration(1)*time.Minute {
		t.Errorf("Unexpected call duration received: %v", cc.GetDuration())
	}
	attrs := &AttrGetAccount{Tenant: "cgrates.org", Account: "1001", BalanceType: "*monetary", Direction: "*out"}
	if err := rater.Call("ApierV1.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV1.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue() != 9.6 {
		t.Errorf("Calling ApierV1.GetBalance expected: 9.6, received: %f", acnt.BalanceMap[attrs.BalanceType+attrs.Direction].GetTotalValue())
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
