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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	actsCdrStore CdrStorage
	actsCfgPath  string
	actsCfgDir   string
	actsCfg      *config.CGRConfig
	actsRPC      *rpc.Client
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
	testActionsKillEngine,
}

func TestActionsITRemoveSMCostRedis(t *testing.T) {
	actsCfgDir = "actions"
	for _, stest := range sTestsActions {
		t.Run("TestsActionsITRedis", stest)
	}
}

func TestActionsITRemoveSMCostMongo(t *testing.T) {
	actsCfgDir = "cdrsv2mongo"
	for _, stest := range sTestsActions {
		t.Run("TestsActionsITMongo", stest)
	}
}

func testActionsInitCfg(t *testing.T) {
	var err error
	actsCfgPath = path.Join(*dataDir, "conf", "samples", actsCfgDir)
	actsCfg, err = config.NewCGRConfigFromPath(actsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsInitCdrsStore(t *testing.T) {
	if actsCfgDir == "actions" {
		if actsCdrStore, err = NewMySQLStorage(actsCfg.StorDbCfg().StorDBHost,
			actsCfg.StorDbCfg().StorDBPort, actsCfg.StorDbCfg().StorDBName,
			actsCfg.StorDbCfg().StorDBUser, actsCfg.StorDbCfg().StorDBPass,
			actsCfg.StorDbCfg().StorDBMaxOpenConns, actsCfg.StorDbCfg().StorDBMaxIdleConns,
			actsCfg.StorDbCfg().StorDBConnMaxLifetime); err != nil {
			t.Fatal("Could not connect to mysql", err.Error())
		}

	} else if actsCfgDir == "cdrsv2mongo" {
		if actsCdrStore, err = NewMongoStorage(actsCfg.StorDbCfg().StorDBHost,
			actsCfg.StorDbCfg().StorDBPort, actsCfg.StorDbCfg().StorDBName,
			actsCfg.StorDbCfg().StorDBUser, actsCfg.StorDbCfg().StorDBPass,
			utils.StorDB, nil, actsCfg.CacheCfg(), false); err != nil {
			t.Fatal("Could not connect to mongo", err.Error())
		}
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
	if _, err := StopStartEngine(actsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testActionsRPCConn(t *testing.T) {
	var err error
	actsRPC, err = jsonrpc.Dial("tcp", actsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
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
	if rcv, err := actsCdrStore.GetSMCosts("", "", "", ""); err != nil {
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
				Identifier:      MetaRemoveSessionCosts,
				TimingTags:      utils.ASAP,
				ExtraParameters: "*string:~OriginID:13;*notstring:~OriginID:12",
				Weight:          20,
			},
		},
	}
	if err := actsRPC.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: attrsAA.ActionsId}
	if err := actsRPC.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}

	// READ
	if rcv, err := actsCdrStore.GetSMCosts("", "", "", ""); err != nil {
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
				Identifier:      MetaRemoveSessionCosts,
				TimingTags:      utils.ASAP,
				ExtraParameters: "",
				Weight:          20,
			},
		},
	}
	if err := actsRPC.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	attrsEA := &utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: attrsAA.ActionsId}
	if err := actsRPC.Call("ApierV1.ExecuteAction", attrsEA, &reply); err != nil {
		t.Error("Got error on ApierV1.ExecuteAction: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV1.ExecuteAction received: %s", reply)
	}

	// READ
	if _, err := actsCdrStore.GetSMCosts("", "", "", ""); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testActionsKillEngine(t *testing.T) {
	if err := KillEngine(100); err != nil {
		t.Error(err)
	}
}
