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
	tpAccPlansCfgPath   string
	tpAccPlansCfg       *config.CGRConfig
	tpAccPlansRPC       *rpc.Client
	tpAccPlansDataDir   = "/usr/share/cgrates"
	tpAccPlan           *utils.TPActionPlan
	tpAccPlansDelay     int
	tpAccPlansConfigDIR string //run tests for specific configuration

)

var sTestsTPAccPlans = []func(t *testing.T){
	testTPAccPlansInitCfg,
	testTPAccPlansResetStorDb,
	testTPAccPlansStartEngine,
	testTPAccPlansRpcConn,
	testTPAccPlansGetTPAccPlanBeforeSet,
	testTPAccPlansSetTPAccPlan,
	testTPAccPlansGetTPAccPlanAfterSet,
	testTPAccPlansGetTPAccPlanIds,
	testTPAccPlansUpdateTPAccPlan,
	testTPAccPlansGetTPAccPlanAfterUpdate,
	testTPAccPlansRemTPAccPlan,
	testTPAccPlansGetTPAccPlanAfterRemove,
	testTPAccPlansKillEngine,
}

//Test start here
func TestTPAccPlansITMySql(t *testing.T) {
	tpAccPlansConfigDIR = "tutmysql"
	for _, stest := range sTestsTPAccPlans {
		t.Run(tpAccPlansConfigDIR, stest)
	}
}

func TestTPAccPlansITMongo(t *testing.T) {
	tpAccPlansConfigDIR = "tutmongo"
	for _, stest := range sTestsTPAccPlans {
		t.Run(tpAccPlansConfigDIR, stest)
	}
}

func TestTPAccPlansITPG(t *testing.T) {
	tpAccPlansConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPAccPlans {
		t.Run(tpAccPlansConfigDIR, stest)
	}
}

func testTPAccPlansInitCfg(t *testing.T) {
	var err error
	tpAccPlansCfgPath = path.Join(tpAccPlansDataDir, "conf", "samples", tpAccPlansConfigDIR)
	tpAccPlansCfg, err = config.NewCGRConfigFromFolder(tpAccPlansCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpAccPlansCfg.DataFolderPath = tpAccPlansDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpAccPlansCfg)
	switch tpAccPlansConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpAccPlansDelay = 2000
	default:
		tpAccPlansDelay = 1000
	}
}

// Wipe out the cdr database
func testTPAccPlansResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpAccPlansCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPAccPlansStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpAccPlansCfgPath, tpAccPlansDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPAccPlansRpcConn(t *testing.T) {
	var err error
	tpAccPlansRPC, err = jsonrpc.Dial("tcp", tpAccPlansCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPAccPlansGetTPAccPlanBeforeSet(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpAccPlansRPC.Call("ApierV1.GetTPActionPlan",
		&AttrGetTPActionPlan{TPid: "TPAcc", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAccPlansSetTPAccPlan(t *testing.T) {
	tpAccPlan = &utils.TPActionPlan{
		TPid: "TPAcc",
		ID:   "ID",
		ActionPlan: []*utils.TPActionTiming{
			{
				ActionsId: "AccId",
				TimingId:  "TimingID",
				Weight:    10,
			},
			{
				ActionsId: "AccId2",
				TimingId:  "TimingID2",
				Weight:    11,
			},
		},
	}
	var result string
	if err := tpAccPlansRPC.Call("ApierV1.SetTPActionPlan", tpAccPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAccPlansGetTPAccPlanAfterSet(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpAccPlansRPC.Call("ApierV1.GetTPActionPlan",
		&AttrGetTPActionPlan{TPid: "TPAcc", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccPlan.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccPlan.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpAccPlan.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccPlan.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpAccPlan.ActionPlan), len(reply.ActionPlan)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpAccPlan.ActionPlan), len(reply.ActionPlan))
	}
}

func testTPAccPlansGetTPAccPlanIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID"}
	if err := tpAccPlansRPC.Call("ApierV1.GetTPActionPlanIds",
		&AttrGetTPActionPlanIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPAccPlansUpdateTPAccPlan(t *testing.T) {
	tpAccPlan.ActionPlan = []*utils.TPActionTiming{
		{
			ActionsId: "AccId",
			TimingId:  "TimingID",
			Weight:    10,
		},
		{
			ActionsId: "AccId2",
			TimingId:  "TimingID2",
			Weight:    11,
		},
		{
			ActionsId: "AccId3",
			TimingId:  "TimingID3",
			Weight:    12,
		},
	}
	var result string
	if err := tpAccPlansRPC.Call("ApierV1.SetTPActionPlan", tpAccPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPAccPlansGetTPAccPlanAfterUpdate(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpAccPlansRPC.Call("ApierV1.GetTPActionPlan",
		&AttrGetTPActionPlan{TPid: "TPAcc", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpAccPlan.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccPlan.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpAccPlan.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpAccPlan.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpAccPlan.ActionPlan), len(reply.ActionPlan)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpAccPlan.ActionPlan), len(reply.ActionPlan))
	}

}

func testTPAccPlansRemTPAccPlan(t *testing.T) {
	var resp string
	if err := tpAccPlansRPC.Call("ApierV1.RemTPActionPlan",
		&AttrGetTPActionPlan{TPid: "TPAcc", ID: "ID"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPAccPlansGetTPAccPlanAfterRemove(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpAccPlansRPC.Call("ApierV1.GetTPActionPlan",
		&AttrGetTPActionPlan{TPid: "TPAcc", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAccPlansKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpAccPlansDelay); err != nil {
		t.Error(err)
	}
}
