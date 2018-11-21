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

package general_tests

import (
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

var tutSMGCfgPath string
var tutSMGCfg *config.CGRConfig
var tutSMGRpc *rpc.Client
var smgLoadInst utils.LoadInstance // Share load information between tests

func TestTutSMGInitCfg(t *testing.T) {
	tutSMGCfgPath = path.Join(*dataDir, "conf", "samples", "smgeneric")
	// Init config first
	var err error
	tutSMGCfg, err = config.NewCGRConfigFromFolder(tutSMGCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutSMGCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutSMGCfg)
}

// Remove data in both rating and accounting db
func TestTutSMGResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutSMGResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutSMGStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutSMGCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutSMGRpcConn(t *testing.T) {
	var err error
	tutSMGRpc, err = jsonrpc.Dial("tcp", tutSMGCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutSMGLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := tutSMGRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &smgLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutSMGCacheStats(t *testing.T) {
	var reply string
	if err := tutSMGRpc.Call("ApierV1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
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
	if err := tutSMGRpc.Call("ApierV2.GetCacheStats", args, &rcvStats); err != nil {
		t.Error("Got error on ApierV2.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV2.GetCacheStats expected: %+v, received: %+v", expectedStats, rcvStats)
	}
}

/*
// Make sure account was debited properly
func TestTutSMGAccountsBefore(t *testing.T) {
	var reply *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		jsn, _ := json.Marshal(reply)
		t.Errorf("Calling ApierV2.GetBalance received: %s", jsn)
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1003"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1004"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 10.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1007"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if reply.BalanceMap[utils.MONETARY].GetTotalValue() != 0.0 { // Make sure we debitted
		t.Errorf("Calling ApierV1.GetBalance received: %f", reply.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	attrs = &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tutSMGRpc.Call("ApierV2.GetAccount", attrs, &reply); err == nil || !strings.HasSuffix(err.Error(), "does not exist") {
		t.Error("Got error on ApierV2.GetAccount: %v", err)
	}
}

*/

func TestTutSMGStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
