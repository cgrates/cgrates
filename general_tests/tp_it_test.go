//go:build integration
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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	v2 "github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpCfgPath  string
	tpCfgDIR   string
	tpCfg      *config.CGRConfig
	tpRPC      *birpc.Client
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
	switch *utils.DBType {
	case utils.MetaInternal:
		tpCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		tpCfgDIR = "tutmysql"
	case utils.MetaMongo:
		tpCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestTp {
		t.Run(tpCfgDIR, stest)
	}
}
func testTpInitCfg(t *testing.T) {
	tpCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpCfgDIR)
	// Init config first
	var err error
	tpCfg, err = config.NewCGRConfigFromPath(tpCfgPath)
	if err != nil {
		t.Error(err)
	}
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
	tpRPC = engine.NewRPCClient(t, tpCfg.ListenCfg())
}

// Load the tariff plan, creating accounts and their balances
func testTpLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testtp")}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &tpLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testTpBalanceCounter(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Destination:   "+49",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(20 * time.Second),
		},
	}
	var cc engine.CallCost
	if err := tpRPC.Call(context.Background(), utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", cc.GetDuration())
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	} else if acnt.UnitCounters[utils.MetaMonetary][1].Counters[0].Value != 20.0 {
		t.Errorf("Calling APIerSv2.GetBalance received: %s", utils.ToIJSON(acnt))
	}
}

func testTpActionTriggers(t *testing.T) {
	var atrs engine.ActionTriggers
	if err := tpRPC.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err.Error())
	} else if len(atrs) != 4 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv1SetActionTrigger, v1.AttrSetActionTrigger{
		GroupID:  "TestATR",
		UniqueID: "Unique atr id",
		ActionTrigger: map[string]any{
			utils.BalanceID: utils.StringPointer("BID1"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling v1.SetActionTrigger got: %v", reply)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{}}, &atrs); err != nil {
		t.Error(err)
	} else if len(atrs) != 5 {
		t.Errorf("Calling v1.GetActionTriggers got: %v", atrs)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv1GetActionTriggers, &v1.AttrGetActionTriggers{GroupIDs: []string{"TestATR"}}, &atrs); err != nil {
		t.Error("Got error on APIerSv1.GetActionTriggers: ", err.Error())
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
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	}
	if acnt == nil {
		t.Errorf("Expecting acnt to not be nil")
		// acnt shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	balanceValueBefore := acnt.BalanceMap[utils.MetaMonetary][0].Value
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "free",
			Account:       "1012",
			Destination:   "+49",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(20 * time.Second),
		},
	}
	var cc engine.CallCost
	if err := tpRPC.Call(context.Background(), utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.Debit: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.MetaMonetary][0].Value != balanceValueBefore {
		t.Errorf("Calling APIerSv2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func testTpZeroNegativeCost(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "free",
			Account:       "1013",
			Destination:   "+4915",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(20 * time.Second),
		},
	}
	var cc engine.CallCost
	if err := tpRPC.Call(context.Background(), utils.ResponderDebit, cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.GetDuration() != 20*time.Second {
		t.Errorf("Calling Responder.MaxDebit got callcost: %v", utils.ToIJSON(cc))
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1013"}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	} else if acnt.BalanceMap[utils.MetaVoice][0].Value != 100.0 {
		t.Errorf("Calling APIerSv2.GetAccount received: %s", utils.ToIJSON(acnt))
	}
}

func testTpExecuteActionCgrRpc(t *testing.T) {
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, utils.AttrExecuteAction{ActionsId: "RPC"}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1013"}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	}
}

func testTpExecuteActionCgrRpcAcc(t *testing.T) {
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1016",
		ActionsId: "RPC_DEST",
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var dests []*engine.Destination
	attrs := &v2.AttrGetDestinations{DestinationIDs: []string{}}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetDestinations, attrs, &dests); err != nil {
		t.Error("Got error on APIerSv2.GetDestinations: ", err.Error())
	}
}

func testTpCreateExecuteActionMatch(t *testing.T) {
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv2SetActions, &utils.AttrSetActions{
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
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions got reply: %s", reply)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1015",
		ActionsId: "PAYMENT_2056bd2fe137082970f97102b64e42fd",
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ExecuteAction got reply: %s", reply)
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1015"}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error("Got error on APIerSv2.GetAccount: ", err.Error())
	}
	if len(acnt.BalanceMap) != 1 ||
		len(acnt.BalanceMap[utils.MetaMonetary]) != 1 ||
		acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != 21 {
		t.Error("error matching previous created balance: ", utils.ToIJSON(acnt.BalanceMap))
	}
}

func testTpSetRemoveActions(t *testing.T) {
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv2SetActions, &utils.AttrSetActions{
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
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions got reply: %s", reply)
	}
	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err != nil {
		t.Error("Got error on APIerSv2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling APIerSv2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2RemoveActions, v1.AttrRemoveActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.RemoveActions got reply: %s", reply)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{
		ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
	}, &actionsMap); err == nil {
		t.Error("no error on APIerSv2.GetActions: ", err)
	}
}

func testTpRemoveActionsRefenced(t *testing.T) {
	actionsMap := make(map[string]engine.Actions)
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetActions, &v2.AttrGetActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &actionsMap); err != nil {
		t.Error("Got error on APIerSv2.GetActions: ", err.Error())
	} else if len(actionsMap) != 1 {
		t.Errorf("Calling APIerSv2.GetActions got reply: %s", utils.ToIJSON(actionsMap))
	}
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv2RemoveActions, v1.AttrRemoveActions{
		ActionIDs: []string{"TOPUP_VOICE"},
	}, &reply); err != nil {
		t.Error("Error on APIerSv2.RemoveActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.RemoveActions got reply: %s", reply)
	}
	/*
		if err := tpRPC.Call(context.Background(),utils.APIerSv2GetActions, v2.AttrGetActions{
			ActionIDs: []string{"PAYMENT_2056bd2fe137082970f97102b64e42fd"},
		}, &actionsMap); err == nil {
			t.Error("no error on APIerSv2.GetActions: ", err)
		}
	*/
}

func testTpApierResetAccountActionTriggers(t *testing.T) {
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1005"}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.ActionTriggers[0].Executed != true {
		t.Skip("Skipping test since Executed is not yet true")
	}
	var reply string
	if err := tpRPC.Call(context.Background(), utils.APIerSv2ResetAccountActionTriggers, v1.AttrResetAccountActionTriggers{
		Tenant:   "cgrates.org",
		Account:  "1005",
		GroupID:  "STANDARD_TRIGGERS",
		Executed: true,
	}, &reply); err != nil {
		t.Error("Error on APIerSv2.ResetAccountActionTriggers: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.ResetAccountActionTriggers got reply: %s", reply)
	}
	if err := tpRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
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
