// +build offline_tp

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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpAccActionsCfgPath   string
	tpAccActionsCfg       *config.CGRConfig
	tpAccActionsRPC       *rpc.Client
	tpAccActionsDataDir   = "/usr/share/cgrates"
	tpAccActions          *utils.TPAccountActions
	tpAccActionsDelay     int
	tpAccActionsConfigDIR string //run tests for specific configuration
	tpAccActionID         = "ID:cgrates.org:1001"
)

var sTestsTPAccActions = []func(t *testing.T){
	testTPAccActionsInitCfg,
	testTPAccActionsResetStorDb,
	testTPAccActionsStartEngine,
	testTPAccActionsRpcConn,
	testTPAccActionsGetTPAccActionBeforeSet,
	testTPAccActionsSetTPAccAction,
	testTPAccActionsGetTPAccActionAfterSet,
	testTPAccActionsGetTPAccountActionLoadIds,
	testTPAccActionsGetTPAccountActionIds,
	testTPAccActionsUpdateTPAccAction,
	testTPAccActionsGetTPAccActionAfterUpdate,
	testTPAccActionsRemTPAccAction,
	testTPAccActionsGetTPAccActionAfterRemove,
	testTPAccActionsKillEngine,
}

//Test start here
func TestTPAccActionsITMySql(t *testing.T) {
	tpAccActionsConfigDIR = "tutmysql"
	for _, stest := range sTestsTPAccActions {
		t.Run(tpAccActionsConfigDIR, stest)
	}
}

func TestTPAccActionsITMongo(t *testing.T) {
	tpAccActionsConfigDIR = "tutmongo"
	for _, stest := range sTestsTPAccActions {
		t.Run(tpAccActionsConfigDIR, stest)
	}
}

func TestTPAccActionsITPG(t *testing.T) {
	tpAccActionsConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPAccActions {
		t.Run(tpAccActionsConfigDIR, stest)
	}
}

func testTPAccActionsInitCfg(t *testing.T) {
	var err error
	tpAccActionsCfgPath = path.Join(tpAccActionsDataDir, "conf", "samples", tpAccActionsConfigDIR)
	tpAccActionsCfg, err = config.NewCGRConfigFromFolder(tpAccActionsCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpAccActionsCfg.DataFolderPath = tpAccActionsDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpAccActionsCfg)
	tpAccActionsDelay = 1000
}

// Wipe out the cdr database
func testTPAccActionsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpAccActionsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPAccActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpAccActionsCfgPath, tpAccActionsDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPAccActionsRpcConn(t *testing.T) {
	var err error
	tpAccActionsRPC, err = jsonrpc.Dial("tcp", tpAccActionsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPAccActionsGetTPAccActionBeforeSet(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActions",
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testTPAccActionsSetTPAccAction(t *testing.T) {
	tpAccActions = &utils.TPAccountActions{
		TPid:         "TPAcc",
		LoadId:       "ID",
		Tenant:       "cgrates.org",
		Account:      "1001",
		ActionPlanId: "PREPAID_10",
		//ActionTriggersId: "STANDARD_TRIGGERS", // ActionTriggersId is optional
		AllowNegative: true,
		Disabled:      false,
	}
	var result string
	if err := tpAccActionsRPC.Call("ApierV1.SetTPAccountActions", tpAccActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAccActionsGetTPAccActionAfterSet(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActions",
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccActions, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccActions, reply)
	}
}

func testTPAccActionsGetTPAccountActionLoadIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID"}
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActionLoadIds",
		&AttrGetTPAccountActionIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAccActionsGetTPAccountActionIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID:cgrates.org:1001"}
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActionIds",
		&AttrGetTPAccountActionIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAccActionsUpdateTPAccAction(t *testing.T) {
	tpAccActions.ActionPlanId = "PlanOne"
	var result string
	if err := tpAccActionsRPC.Call("ApierV1.SetTPAccountActions", tpAccActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPAccActionsGetTPAccActionAfterUpdate(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActions",
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccActions, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccActions, reply)
	}

}

func testTPAccActionsRemTPAccAction(t *testing.T) {
	var resp string
	if err := tpAccActionsRPC.Call("ApierV1.RemTPAccountActions",
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPAccActionsGetTPAccActionAfterRemove(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call("ApierV1.GetTPAccountActions",
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAccActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpDestinationDelay); err != nil {
		t.Error(err)
	}
}
