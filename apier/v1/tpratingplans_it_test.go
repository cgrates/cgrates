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
	tpRatingPlanCfgPath   string
	tpRatingPlanCfg       *config.CGRConfig
	tpRatingPlanRPC       *birpc.Client
	tpRatingPlan          *utils.TPRatingPlan
	tpRatingPlanDelay     int
	tpRatingPlanConfigDIR string //run tests for specific configuration
)

var sTestsTPRatingPlans = []func(t *testing.T){
	testTPRatingPlansInitCfg,
	testTPRatingPlansResetStorDb,
	testTPRatingPlansStartEngine,
	testTPRatingPlansRpcConn,
	testTPRatingPlansGetTPRatingPlanBeforeSet,
	testTPRatingPlansSetTPRatingPlan,
	testTPRatingPlansGetTPRatingPlanAfterSet,
	testTPRatingPlansGetTPRatingPlanIds,
	testTPRatingPlansUpdateTPRatingPlan,
	testTPRatingPlansGetTPRatingPlanAfterUpdate,
	testTPRatingPlansRemoveTPRatingPlan,
	testTPRatingPlansGetTPRatingPlanAfterRemove,
	testTPRatingPlansKillEngine,
}

// Test start here
func TestTPRatingPlansIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpRatingPlanConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRatingPlanConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRatingPlanConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpRatingPlanConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRatingPlans {
		t.Run(tpRatingPlanConfigDIR, stest)
	}
}

func testTPRatingPlansInitCfg(t *testing.T) {
	var err error
	tpRatingPlanCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpRatingPlanConfigDIR)
	tpRatingPlanCfg, err = config.NewCGRConfigFromPath(tpRatingPlanCfgPath)
	if err != nil {
		t.Error(err)
	}
	switch tpRatingPlanConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpRatingPlanDelay = 2000
	default:
		tpRatingPlanDelay = 1000
	}
}

// Wipe out the cdr database
func testTPRatingPlansResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRatingPlanCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRatingPlansStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRatingPlanCfgPath, tpRatingPlanDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRatingPlansRpcConn(t *testing.T) {
	var err error
	tpRatingPlanRPC, err = jsonrpc.Dial(utils.TCP, tpRatingPlanCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatingPlansGetTPRatingPlanBeforeSet(t *testing.T) {
	var reply *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1GetTPRatingPlan,
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingPlansSetTPRatingPlan(t *testing.T) {
	tpRatingPlan = &utils.TPRatingPlan{
		TPid: "TPRP1",
		ID:   "Plan1",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			{
				DestinationRatesId: "RateId",
				TimingId:           "TimingID",
				Weight:             12,
			},
			{
				DestinationRatesId: "DR_FREESWITCH_USERS",
				TimingId:           "ALWAYS",
				Weight:             10,
			},
		},
	}
	var result string
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1SetTPRatingPlan, tpRatingPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterSet(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1GetTPRatingPlan,
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingPlan.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingPlan.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingPlan.ID, respond.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingPlan, respond)
	} else if !reflect.DeepEqual(len(tpRatingPlan.RatingPlanBindings), len(respond.RatingPlanBindings)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpRatingPlan.RatingPlanBindings), len(respond.RatingPlanBindings))
	}
}

func testTPRatingPlansGetTPRatingPlanIds(t *testing.T) {
	var result []string
	expected := []string{"Plan1"}
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1GetTPRatingPlanIds,
		&AttrGetTPRatingPlanIds{TPid: tpRatingPlan.TPid}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting: %+v, received: %+v", expected, result)
	}
}

func testTPRatingPlansUpdateTPRatingPlan(t *testing.T) {
	tpRatingPlan.RatingPlanBindings = []*utils.TPRatingPlanBinding{
		{
			DestinationRatesId: "RateId",
			TimingId:           "TimingID",
			Weight:             12,
		},
		{
			DestinationRatesId: "DR_FREESWITCH_USERS",
			TimingId:           "ALWAYS",
			Weight:             10,
		},
		{
			DestinationRatesId: "RateID2",
			TimingId:           "ALWAYS",
			Weight:             11,
		},
	}
	var result string
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1SetTPRatingPlan, tpRatingPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterUpdate(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1GetTPRatingPlan,
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRatingPlan.TPid, respond.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingPlan.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpRatingPlan.ID, respond.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpRatingPlan, respond)
	} else if !reflect.DeepEqual(len(tpRatingPlan.RatingPlanBindings), len(respond.RatingPlanBindings)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpRatingPlan.RatingPlanBindings), len(respond.RatingPlanBindings))
	}
}

func testTPRatingPlansRemoveTPRatingPlan(t *testing.T) {
	var resp string
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1RemoveTPRatingPlan,
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterRemove(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call(context.Background(), utils.APIerSv1GetTPRatingPlan,
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingPlansKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRatingPlanDelay); err != nil {
		t.Error(err)
	}
}
