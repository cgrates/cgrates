//go:build offline
// +build offline

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpAccActionsCfgPath   string
	tpAccActionsCfg       *config.CGRConfig
	tpAccActionsRPC       *birpc.Client
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
	testTPAccActionsGetTPAccountActionsByLoadId,
	testTPAccActionsGetTPAccountActionLoadIds,
	testTPAccActionsGetTPAccountActionIds,
	testTPAccActionsUpdateTPAccAction,
	testTPAccActionsGetTPAccActionAfterUpdate,
	testTPAccActionsRemTPAccAction,
	testTPAccActionsGetTPAccActionAfterRemove,
	testTPAccActionsKillEngine,
}

// Test start here
func TestTPAccActionsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpAccActionsConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpAccActionsConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpAccActionsConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpAccActionsConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPAccActions {
		t.Run(tpAccActionsConfigDIR, stest)
	}
}

func testTPAccActionsInitCfg(t *testing.T) {
	var err error
	tpAccActionsCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpAccActionsConfigDIR)
	tpAccActionsCfg, err = config.NewCGRConfigFromPath(tpAccActionsCfgPath)
	if err != nil {
		t.Error(err)
	}
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
	tpAccActionsRPC, err = jsonrpc.Dial(utils.TCP, tpAccActionsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPAccActionsGetTPAccActionBeforeSet(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActions,
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
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
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1SetTPAccountActions, tpAccActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAccActionsGetTPAccActionAfterSet(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActions,
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccActions, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccActions, reply)
	}
}

func testTPAccActionsGetTPAccountActionsByLoadId(t *testing.T) {
	var reply *[]*utils.TPAccountActions
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActionsByLoadId,
		&utils.TPAccountActions{TPid: "TPAcc", LoadId: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccActions, (*reply)[0]) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccActions, (*reply)[0])
	}
}

func testTPAccActionsGetTPAccountActionLoadIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID"}
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActionLoadIds,
		&AttrGetTPAccountActionIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAccActionsGetTPAccountActionIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID:cgrates.org:1001"}
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActionIds,
		&AttrGetTPAccountActionIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAccActionsUpdateTPAccAction(t *testing.T) {
	tpAccActions.ActionPlanId = "PlanOne"
	var result string
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1SetTPAccountActions, tpAccActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPAccActionsGetTPAccActionAfterUpdate(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActions,
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccActions, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccActions, reply)
	}

}

func testTPAccActionsRemTPAccAction(t *testing.T) {
	var resp string
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1RemoveTPAccountActions,
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPAccActionsGetTPAccActionAfterRemove(t *testing.T) {
	var reply *utils.TPAccountActions
	if err := tpAccActionsRPC.Call(context.Background(), utils.APIerSv1GetTPAccountActions,
		&AttrGetTPAccountActions{TPid: "TPAcc", AccountActionsId: tpAccActionID}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAccActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpAccActionsDelay); err != nil {
		t.Error(err)
	}
}
