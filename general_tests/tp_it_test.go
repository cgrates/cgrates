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
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpCfgPath  string
	tpCfg      *config.CGRConfig
	tpRPC      *rpc.Client
	tpLoadInst utils.LoadInstance // Share load information between tests

	sTestTp = []func(t *testing.T){
		testTpInitCfg,
		testTpResetDataDb,
		testTpResetStorDb,
		testTpStartEngine,
		testTpRpcConn,
		testTpLoadTariffPlanFromFolder,
		testTpBalanceCounter,
		testTpActionTriggers,
		testTpZeroCost,
		testTpZeroNegativeCost,
		testTpExecuteActionCgrRpc,
		testTpExecuteActionCgrRpcAcc,
		//testTpExecuteActionCgrRpcCdrStats,
		testTpCreateExecuteActionMatch,
		testTpSetRemoveActions,
		testTpRemoveActionsRefenced,
		testTpApierResetAccountActionTriggers,
		testTpStopCgrEngine,
	}
)

func TestTp(t *testing.T) {
	for _, stest := range sTestTp {
		t.Run("TestTp", stest)
	}
}
func testTpInitCfg(t *testing.T) {
	tpCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql")
	// Init config first
	var err error
	tpCfg, err = config.NewCGRConfigFromPath(tpCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpCfg)
}

// Remove data in both rating and accounting db
func testTpResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testTpResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpCfgPath, 1000); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTpRpcConn(t *testing.T) {
	var err error
	tpRPC, err = jsonrpc.Dial("tcp", tpCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testTpLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := tpRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &tpLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testTpBalanceCounter(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "1001",
		Destination:   "+49",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call(utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.UnitCounters[utils.MONETARY][1].Counters[0].Value != 20.0 {
		t.Errorf("Calling ApierV2.GetBalance received: %s", utils.ToIJSON(acnt))
	}
}

func testTpActionTriggers(t *testing.T) {
	var atrs engine.ActionTriggers
	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error("Got error on ApierV1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 4 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	var reply string
	if err := tpRPC.Call("ApierV1.SetActionTrigger", v1.AttrSetActionTrigger{
		GroupID:   "TestATR",
		UniqueID:  "Unique atr id",
		BalanceID: utils.StringPointer("BID1"),
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling v1.SetActionTrigger got: %v", reply)
	}
	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error(err)
	} else if len(atrs) != 5 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	if err := tpRPC.Call("ApierV1.GetActionTriggers", v1.AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err != nil {
		t.Error("Got error on ApierV1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 1 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	if atrs == nil {
		t.Errorf("Expecting atrs to not be nil")
		// atrs shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	if atrs[0].ID != "TestATR" ||
		atrs[0].UniqueID != "Unique atr id" ||
		*atrs[0].Balance.ID != "BID1" {
		t.Error("Wrong action trigger set: ", utils.ToIJSON(atrs[0]))
	}
}

func testTpZeroCost(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1012"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
	if acnt == nil {
		t.Errorf("Expecting acnt to not be nil")
		// acnt shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	balanceValueBefore := acnt.BalanceMap[utils.MONETARY][0].Value
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "free",
		Account:       "1012",
		Destination:   "+49",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call(utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.MONETARY][0].Value != balanceValueBefore {
		t.Errorf("Calling ApierV2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func testTpZeroNegativeCost(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := engine.CallDescriptor{
		Category:      "call",
		Tenant:        "cgrates.org",
		Subject:       "free",
		Account:       "1013",
		Destination:   "+4915",
		DurationIndex: 0,
		TimeStart:     tStart,
		TimeEnd:       tStart.Add(time.Duration(20) * time.Second),
	}
	var cc engine.CallCost
	if err := tpRPC.Call(utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1013"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.VOICE][0].Value != 100.0 {
		t.Errorf("Calling ApierV2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func testTpExecuteActionCgrRpc(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{ActionsId: "RPC"}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1013"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
}

func testTpExecuteActionCgrRpcAcc(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1016",
		ActionsId: "RPC_DEST",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var dests []*engine.Destination
	attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
	if err := tpRPC.Call("ApierV2.GetDestinations", attrs, &dests); err != nil {
		t.Error("Got error on ApierV2.GetDestinations: ", err.Error())
	}
}

// Deprecated
// func //(t *testing.T) {
// 	var reply string
// 	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
// 		ActionsId: "RPC_CDRSTATS",
// 	}, &reply); err != nil {
// 		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Errorf("Calling ExecuteAction got reply: %s", reply)
// 	}
// 	var queue engine.CDRStatsQueue
// 	time.Sleep(20 * time.Millisecond)
// 	if err := tpRPC.Call("CDRStatsV1.GetQueue", "qtest", &queue); err != nil {
// 		t.Error("Got error on CDRStatsV1.GetQueue: ", err.Error())
// 	}
// }

func testTpCreateExecuteActionMatch(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.SetActions", utils.AttrSetActions{
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
		Actions: []*utils.TPAction{
			{
				BalanceType:   "*monetary",
				Identifier:    "*topup",
				RatingSubject: "",
				Units:         "10.500000",
				Weight:        10,
			},
		},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.ExecuteAction", utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1015"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error("Got error on ApierV2.GetAccount: ", err.Error())
	}
	if len(acnt.BalanceMap) != 1 ||
		len(acnt.BalanceMap[utils.MONETARY]) != 1 ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != 21 {
		t.Error("error matching previous created balance: ", utils.ToIJSON(acnt.BalanceMap))
	}
}

func testTpSetRemoveActions(t *testing.T) {
	var reply string
	if err := tpRPC.Call("ApierV2.SetActions", utils.AttrSetActions{
		ActionsId: "TO_BE_DELETED",
		Actions: []*utils.TPAction{
			{
				BalanceType:   "*monetary",
				Identifier:    "*topup",
				RatingSubject: "",
				Units:         "10.500000",
				Weight:        10,
			},
		},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions got reply: %s", reply)
	}
	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err != nil {
		t.Error("Got error on ApierV2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling ApierV2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	if err := tpRPC.Call("ApierV2.RemoveActions", v1.AttrRemoveActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &reply); err != nil {
		t.Error("Got error on ApierV2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.RemoveActions got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err == nil {
		t.Error("no error on ApierV2.GetActions: ", err)
	}
}

func testTpRemoveActionsRefenced(t *testing.T) {
	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &actionsMap); err != nil {
		t.Error("Got error on ApierV2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling ApierV2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	var reply string
	if err := tpRPC.Call("ApierV2.RemoveActions", v1.AttrRemoveActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &reply); err != nil {
		t.Error("Error on ApierV2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.RemoveActions got reply: %s", reply)
	}
	/*
		if err := tpRPC.Call("ApierV2.GetActions", v2.AttrGetActions{
			ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
		}, &actionsMap); err == nil {
			t.Error("no error on ApierV2.GetActions: ", err)
		}
	*/
}

func testTpApierResetAccountActionTriggers(t *testing.T) {
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ActionTriggers[0].Executed != true {
		t.Skip("Skipping test since Executed is not yet true")
	}
	var reply string
	if err := tpRPC.Call("ApierV2.ResetAccountActionTriggers", v1.AttrResetAccountActionTriggers{
		Tenant:   "cgrates.org",
		Account:  "1005",
		GroupID:  "STANDARD_TRIGGERS",
		Executed: true,
	}, &reply); err != nil {
		t.Error("Error on ApierV2.ResetAccountActionTriggers: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.ResetAccountActionTriggers got reply: %s", reply)
	}
	if err := tpRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ActionTriggers[0].Executed == false {
		t.Errorf("wrong action trigger executed flag: %s", utils.ToIJSON(acnt.ActionTriggers))
	}
}

func testTpStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
