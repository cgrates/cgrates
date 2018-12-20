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
	tpRatingPlanCfgPath   string
	tpRatingPlanCfg       *config.CGRConfig
	tpRatingPlanRPC       *rpc.Client
	tpRatingPlanDataDir   = "/usr/share/cgrates"
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
	testTPRatingPlansRemTPRatingPlan,
	testTPRatingPlansGetTPRatingPlanAfterRemove,
	testTPRatingPlansKillEngine,
}

//Test start here
func TestTPRatingPlansITMySql(t *testing.T) {
	tpRatingPlanConfigDIR = "tutmysql"
	for _, stest := range sTestsTPRatingPlans {
		t.Run(tpRatingPlanConfigDIR, stest)
	}
}

func TestTPRatingPlansITMongo(t *testing.T) {
	tpRatingPlanConfigDIR = "tutmongo"
	for _, stest := range sTestsTPRatingPlans {
		t.Run(tpRatingPlanConfigDIR, stest)
	}
}

func TestTPRatingPlansITPG(t *testing.T) {
	tpRatingPlanConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPRatingPlans {
		t.Run(tpRatingPlanConfigDIR, stest)
	}
}

func testTPRatingPlansInitCfg(t *testing.T) {
	var err error
	tpRatingPlanCfgPath = path.Join(tpRatingPlanDataDir, "conf", "samples", tpRatingPlanConfigDIR)
	tpRatingPlanCfg, err = config.NewCGRConfigFromFolder(tpRatingPlanCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRatingPlanCfg.DataFolderPath = tpRatingPlanDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpRatingPlanCfg)
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
	tpRatingPlanRPC, err = jsonrpc.Dial("tcp", tpRatingPlanCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatingPlansGetTPRatingPlanBeforeSet(t *testing.T) {
	var reply *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call("ApierV1.GetTPRatingPlan",
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingPlansSetTPRatingPlan(t *testing.T) {
	tpRatingPlan = &utils.TPRatingPlan{
		TPid: "TPRP1",
		ID:   "Plan1",
		RatingPlanBindings: []*utils.TPRatingPlanBinding{
			&utils.TPRatingPlanBinding{
				DestinationRatesId: "RateId",
				TimingId:           "TimingID",
				Weight:             12,
			},
			&utils.TPRatingPlanBinding{
				DestinationRatesId: "DR_FREESWITCH_USERS",
				TimingId:           "ALWAYS",
				Weight:             10,
			},
		},
	}
	var result string
	if err := tpRatingPlanRPC.Call("ApierV1.SetTPRatingPlan", tpRatingPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterSet(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call("ApierV1.GetTPRatingPlan",
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
	if err := tpRatingPlanRPC.Call("ApierV1.GetTPRatingPlanIds",
		&AttrGetTPRatingPlanIds{TPid: tpRatingPlan.TPid}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expecting: %+v, received: %+v", expected, result)
	}
}

func testTPRatingPlansUpdateTPRatingPlan(t *testing.T) {
	tpRatingPlan.RatingPlanBindings = []*utils.TPRatingPlanBinding{
		&utils.TPRatingPlanBinding{
			DestinationRatesId: "RateId",
			TimingId:           "TimingID",
			Weight:             12,
		},
		&utils.TPRatingPlanBinding{
			DestinationRatesId: "DR_FREESWITCH_USERS",
			TimingId:           "ALWAYS",
			Weight:             10,
		},
		&utils.TPRatingPlanBinding{
			DestinationRatesId: "RateID2",
			TimingId:           "ALWAYS",
			Weight:             11,
		},
	}
	var result string
	if err := tpRatingPlanRPC.Call("ApierV1.SetTPRatingPlan", tpRatingPlan, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterUpdate(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call("ApierV1.GetTPRatingPlan",
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

func testTPRatingPlansRemTPRatingPlan(t *testing.T) {
	var resp string
	if err := tpRatingPlanRPC.Call("ApierV1.RemTPRatingPlan",
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatingPlansGetTPRatingPlanAfterRemove(t *testing.T) {
	var respond *utils.TPRatingPlan
	if err := tpRatingPlanRPC.Call("ApierV1.GetTPRatingPlan",
		&AttrGetTPRatingPlan{TPid: "TPRP1", ID: "Plan1"}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatingPlansKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRatingPlanDelay); err != nil {
		t.Error(err)
	}
}
