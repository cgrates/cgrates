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
package engine

import (
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	actsCdrStore CdrStorage
	actsCfgPath  string
	actsCfgDir   string
	actsCfg      *config.CGRConfig
	actsRPC      *birpc.Client
)

var sTestsActions = []func(t *testing.T){
	testActionsInitCfg,
	testActionsInitCdrsStore,
	testActionsInitDataDb,
	testActionsResetStorDb,
	testActionsStartEngine,
	testActionsRPCConn,
	testActionsSetSMCosts,
	testActionsExecuteRemoveSMCos1,
	testActionsExecuteRemoveSMCos2,
	testActionsUpdateBalance,
	testActionsKillEngine,
}

func TestActionsITRemoveSMCost(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		actsCfgDir = "actions_mysql"
	case utils.MetaMongo:
		actsCfgDir = "cdrsv2mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *utils.Encoding == utils.MetaGOB {
		actsCfgDir += "_gob"
	}

	for _, stest := range sTestsActions {
		t.Run(actsCfgDir, stest)
	}
}

func testActionsInitCfg(t *testing.T) {
	var err error
	actsCfgPath = path.Join(*utils.DataDir, "conf", "samples", actsCfgDir)
	actsCfg, err = config.NewCGRConfigFromPath(actsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsInitCdrsStore(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		actsCdrStore = NewInternalDB(nil, nil, true, actsCfg.DataDbCfg().Items)
	case utils.MetaMySQL:
		if actsCdrStore, err = NewMySQLStorage(actsCfg.StorDbCfg().Host,
			actsCfg.StorDbCfg().Port, actsCfg.StorDbCfg().Name,
			actsCfg.StorDbCfg().User, actsCfg.StorDbCfg().Password,
			100, 10, 0, "UTC", make(map[string]string)); err != nil {
			t.Fatal("Could not connect to mysql", err.Error())
		}
	case utils.MetaMongo:
		if actsCdrStore, err = NewMongoStorage("mongodb", actsCfg.StorDbCfg().Host,
			actsCfg.StorDbCfg().Port, actsCfg.StorDbCfg().Name,
			actsCfg.StorDbCfg().User, actsCfg.StorDbCfg().Password,
			actsCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second); err != nil {
			t.Fatal("Could not connect to mongo", err.Error())
		}
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
}

func testActionsInitDataDb(t *testing.T) {
	if err := InitDataDb(actsCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testActionsResetStorDb(t *testing.T) {
	if err := InitStorDb(actsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionsStartEngine(t *testing.T) {
	if _, err := StopStartEngine(actsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testActionsRPCConn(t *testing.T) {
	actsRPC = NewRPCClient(t, actsCfg.ListenCfg())
}

func testActionsSetSMCosts(t *testing.T) {
	var snd = []*SMCost{
		{
			CGRID:       "CGRID11",
			RunID:       "21",
			OriginHost:  "host32",
			OriginID:    "11",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID12",
			RunID:       "22",
			OriginHost:  "host32",
			OriginID:    "12",
			CostDetails: NewBareEventCost(),
		},
		{
			CGRID:       "CGRID13",
			RunID:       "23",
			OriginHost:  "host33",
			OriginID:    "13",
			CostDetails: NewBareEventCost(),
		},
	}
	for _, smc := range snd {
		if err := actsCdrStore.SetSMCost(smc); err != nil {
			t.Error(err)
		}
	}
	// READ
	if rcv, err := actsCdrStore.GetSMCosts(utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString); err != nil {
		t.Fatal(err)
	} else if len(rcv) != 3 {
		t.Errorf("Expected 3 results received %v ", len(rcv))
	}
}

func testActionsExecuteRemoveSMCos1(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "REMOVE_SMCOST1",
		Actions: []*utils.TPAction{
			{
				Identifier:      utils.MetaRemoveSessionCosts,
				ExtraParameters: "*string:~*sc.OriginID:13;*notstring:~*sc.OriginID:12",
				Weight:          20,
			},
		},
	}
	if err := actsRPC.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: attrsAA.ActionsId}
	if err := actsRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	// READ
	if rcv, err := actsCdrStore.GetSMCosts(utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString); err != nil {
		t.Error(err)
	} else if len(rcv) != 2 {
		t.Errorf("Expected 2 result received %v ", len(rcv))
	}
}

func testActionsExecuteRemoveSMCos2(t *testing.T) {
	var reply string
	attrsAA := &utils.AttrSetActions{
		ActionsId: "REMOVE_SMCOST2",
		Actions: []*utils.TPAction{
			{
				Identifier:      utils.MetaRemoveSessionCosts,
				ExtraParameters: "",
				Weight:          20,
			},
		},
	}
	if err := actsRPC.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: attrsAA.ActionsId}
	if err := actsRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}

	// READ
	if _, err := actsCdrStore.GetSMCosts(utils.EmptyString, utils.EmptyString, utils.EmptyString, utils.EmptyString); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testActionsUpdateBalance(t *testing.T) {
	var reply string
	attrsSetAccount := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "testAcc"}
	if err := actsRPC.Call(context.Background(), utils.APIerSv1SetAccount, attrsSetAccount, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
	topupAction := &utils.AttrSetActions{ActionsId: "ACT_TOPUP_RST", Actions: []*utils.TPAction{
		{Identifier: utils.MetaTopUp, BalanceId: "test", BalanceType: utils.MetaMonetary, Units: "5", ExpiryTime: utils.MetaUnlimited, Weight: 20.0},
	}}
	if err := actsRPC.Call(context.Background(), utils.APIerSv2SetActions, topupAction, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	changeBlockerAction := &utils.AttrSetActions{ActionsId: "ACT_BAL_UPDT", Actions: []*utils.TPAction{
		{Identifier: utils.MetaSetBalance, BalanceId: "test", BalanceBlocker: "true"},
	}}
	if err := actsRPC.Call(context.Background(), utils.APIerSv2SetActions, changeBlockerAction, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: topupAction.ActionsId}
	if err := actsRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	runtime.Gosched()
	attrsEA2 := &utils.AttrExecuteAction{Tenant: attrsSetAccount.Tenant, Account: attrsSetAccount.Account, ActionsId: changeBlockerAction.ActionsId}
	if err := actsRPC.Call(context.Background(), utils.APIerSv1ExecuteAction, attrsEA2, &reply); err != nil {
		t.Error("Got error on APIerSv1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.ExecuteAction received: %s", reply)
	}
	var acc Account
	attrs2 := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testAcc"}
	if err := actsRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs2, &acc); err != nil {
		t.Error("Got error on APIerSv1.GetAccount: ", err.Error())
	} else if acc.BalanceMap[utils.MetaMonetary][0].ID != "test" {
		t.Errorf("Expected test result received %v ", acc.BalanceMap[utils.MetaMonetary][0].ID)
	} else if acc.BalanceMap[utils.MetaMonetary][0].Blocker != true {
		t.Errorf("Expected true result received %v ", acc.BalanceMap[utils.MetaMonetary][0].Blocker)
	}
}

func testActionsKillEngine(t *testing.T) {
	if err := KillEngine(100); err != nil {
		t.Error(err)
	}
}
