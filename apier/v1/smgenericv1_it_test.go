// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
package v1

import (
	"encoding/json"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var smgV1CfgPath string
var smgV1Cfg *config.CGRConfig
var smgV1Rpc *rpc.Client
var smgV1LoadInst utils.LoadInstance // Share load information between tests

func TestSMGV1InitCfg(t *testing.T) {
	smgV1CfgPath = path.Join(*dataDir, "conf", "samples", "smgeneric")
	// Init config first
	var err error
	smgV1Cfg, err = config.NewCGRConfigFromFolder(smgV1CfgPath)
	if err != nil {
		t.Error(err)
	}
	smgV1Cfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(smgV1Cfg)
}

// Remove data in both rating and accounting db
func TestSMGV1ResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(smgV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGV1ResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(smgV1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGV1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGV1RpcConn(t *testing.T) {
	var err error
	smgV1Rpc, err = jsonrpc.Dial("tcp", smgV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGV1LoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := smgV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestSMGV1CacheStats(t *testing.T) {
	var reply string
	if err := smgV1Rpc.Call("ApierV1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	var rcvStats *utils.CacheStats
	expectedStats := &utils.CacheStats{Destinations: 5, ReverseDestinations: 7, RatingPlans: 4, RatingProfiles: 10,
		Actions: 9, ActionPlans: 4, AccountActionPlans: 5, SharedGroups: 1, DerivedChargers: 1,
		Users: 3, Aliases: 1, ReverseAliases: 2, ResourceProfiles: 3, Resources: 3, StatQueues: 1,
		StatQueueProfiles: 1, Thresholds: 7, ThresholdProfiles: 7, Filters: 16, SupplierProfiles: 3, AttributeProfiles: 1}
	var args utils.AttrCacheStats
	if err := smgV1Rpc.Call("ApierV1.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV1.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

// Make sure account was debited properly
func TestSMGV1AccountsBefore(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := smgV1Rpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		jsn, _ := json.Marshal(reply)
		t.Errorf("Received: %s", jsn)
	}
}

// Make sure account was debited properly
func TestSMGV1GetMaxUsage(t *testing.T) {
	setupReq := map[string]interface{}{utils.RequestType: utils.META_PREPAID, utils.Tenant: "cgrates.org",
		utils.Account: "1003", utils.Destination: "1002", utils.SetupTime: "2015-11-10T15:20:00Z"}
	var maxTime float64
	if err := smgV1Rpc.Call("SMGenericV1.GetMaxUsage", setupReq, &maxTime); err != nil {
		t.Error(err)
	} else if maxTime != 2700 {
		t.Errorf("Calling ApierV1.GetMaxUsage got maxTime: %f", maxTime)
	}
}

func TestSMGV1StopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
