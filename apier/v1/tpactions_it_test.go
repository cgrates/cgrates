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
	tpActionCfgPath   string
	tpActionCfg       *config.CGRConfig
	tpActionRPC       *rpc.Client
	tpActionDataDir   = "/usr/share/cgrates"
	tpActions         *utils.TPActions
	tpActionDelay     int
	tpActionConfigDIR string //run tests for specific configuration

)

var sTestsTPActions = []func(t *testing.T){
	testTPActionsInitCfg,
	testTPActionsResetStorDb,
	testTPActionsStartEngine,
	testTPActionsRpcConn,
	testTPActionsGetTPActionBeforeSet,
	testTPActionsSetTPAction,
	testTPActionsGetTPActionAfterSet,
	testTPActionsGetTPActionIds,
	testTPActionsUpdateTPAction,
	testTPActionsGetTPActionAfterUpdate,
	testTPActionsRemTPAction,
	testTPActionsGetTPActionAfterRemove,
	testTPActionsKillEngine,
}

//Test start here
func TestTPActionsITMySql(t *testing.T) {
	tpActionConfigDIR = "tutmysql"
	for _, stest := range sTestsTPActions {
		t.Run(tpActionConfigDIR, stest)
	}
}

func TestTPActionsITMongo(t *testing.T) {
	tpActionConfigDIR = "tutmongo"
	for _, stest := range sTestsTPActions {
		t.Run(tpActionConfigDIR, stest)
	}
}

func TestTPActionsITPG(t *testing.T) {
	tpActionConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPActions {
		t.Run(tpActionConfigDIR, stest)
	}
}

func testTPActionsInitCfg(t *testing.T) {
	var err error
	tpActionCfgPath = path.Join(tpActionDataDir, "conf", "samples", tpActionConfigDIR)
	tpActionCfg, err = config.NewCGRConfigFromFolder(tpActionCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpActionCfg.DataFolderPath = tpActionDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpActionCfg)
	switch tpActionConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpActionDelay = 2000
	default:
		tpActionDelay = 1000
	}
}

// Wipe out the cdr database
func testTPActionsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpActionCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpActionCfgPath, tpActionDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPActionsRpcConn(t *testing.T) {
	var err error
	tpActionRPC, err = jsonrpc.Dial("tcp", tpActionCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPActionsGetTPActionBeforeSet(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpActionRPC.Call("ApierV1.GetTPActions",
		&AttrGetTPActions{TPid: "TPAcc", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPActionsSetTPAction(t *testing.T) {
	tpActions = &utils.TPActions{
		TPid: "TPAcc",
		ID:   "ID",
		Actions: []*utils.TPAction{
			&utils.TPAction{
				Identifier:      "*topup_reset",
				BalanceId:       "BalID",
				BalanceUuid:     "BalUuid",
				BalanceType:     "*data",
				Directions:      "*out",
				Units:           "10",
				ExpiryTime:      "*unlimited",
				Filter:          "",
				TimingTags:      "2014-01-14T00:00:00Z",
				DestinationIds:  "DST_1002",
				RatingSubject:   "SPECIAL_1002",
				Categories:      "",
				SharedGroups:    "SHARED_A",
				BalanceWeight:   "10",
				ExtraParameters: "",
				BalanceBlocker:  "false",
				BalanceDisabled: "false",
				Weight:          10,
			},
			&utils.TPAction{
				Identifier:      "*log",
				BalanceId:       "BalID",
				BalanceUuid:     "BalUuid",
				BalanceType:     "*monetary",
				Directions:      "*out",
				Units:           "120",
				ExpiryTime:      "*unlimited",
				Filter:          "",
				TimingTags:      "2014-01-14T00:00:00Z",
				DestinationIds:  "*any",
				RatingSubject:   "SPECIAL_1002",
				Categories:      "",
				SharedGroups:    "SHARED_A",
				BalanceWeight:   "11",
				ExtraParameters: "",
				BalanceBlocker:  "false",
				BalanceDisabled: "false",
				Weight:          11,
			},
		},
	}
	var result string
	if err := tpActionRPC.Call("ApierV1.SetTPActions", tpActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPActionsGetTPActionAfterSet(t *testing.T) {
	var reply *utils.TPActions
	if err := tpActionRPC.Call("ApierV1.GetTPActions",
		&AttrGetTPActions{TPid: "TPAcc", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpActions.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpActions.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpActions.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpActions.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpActions.Actions), len(reply.Actions)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpActions.Actions), len(reply.Actions))
	}
}

func testTPActionsGetTPActionIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID"}
	if err := tpActionRPC.Call("ApierV1.GetTPActionIds",
		&AttrGetTPActionIds{TPid: "TPAcc"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPActionsUpdateTPAction(t *testing.T) {
	tpActions.Actions = []*utils.TPAction{
		&utils.TPAction{
			Identifier:      "*topup_reset",
			BalanceId:       "BalID",
			BalanceUuid:     "BalUuid",
			BalanceType:     "*data",
			Directions:      "*out",
			Units:           "10",
			ExpiryTime:      "*unlimited",
			Filter:          "",
			TimingTags:      "2014-01-14T00:00:00Z",
			DestinationIds:  "DST_1002",
			RatingSubject:   "SPECIAL_1002",
			Categories:      "",
			SharedGroups:    "SHARED_A",
			BalanceWeight:   "10",
			ExtraParameters: "",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          10,
		},
		&utils.TPAction{
			Identifier:      "*log",
			BalanceId:       "BalID",
			BalanceUuid:     "BalUuid",
			BalanceType:     "*monetary",
			Directions:      "*out",
			Units:           "120",
			ExpiryTime:      "*unlimited",
			Filter:          "",
			TimingTags:      "2014-01-14T00:00:00Z",
			DestinationIds:  "*any",
			RatingSubject:   "SPECIAL_1002",
			Categories:      "",
			SharedGroups:    "SHARED_A",
			BalanceWeight:   "11",
			ExtraParameters: "",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          11,
		},
		&utils.TPAction{
			Identifier:      "*topup",
			BalanceId:       "BalID",
			BalanceUuid:     "BalUuid",
			BalanceType:     "*voice",
			Directions:      "*out",
			Units:           "102400",
			ExpiryTime:      "*unlimited",
			Filter:          "",
			TimingTags:      "2014-01-14T00:00:00Z",
			DestinationIds:  "*any",
			RatingSubject:   "SPECIAL_1002",
			Categories:      "",
			SharedGroups:    "SHARED_A",
			BalanceWeight:   "20",
			ExtraParameters: "",
			BalanceBlocker:  "false",
			BalanceDisabled: "false",
			Weight:          11,
		},
	}
	var result string
	if err := tpActionRPC.Call("ApierV1.SetTPActions", tpActions, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPActionsGetTPActionAfterUpdate(t *testing.T) {
	var reply *utils.TPActions
	if err := tpActionRPC.Call("ApierV1.GetTPActions",
		&AttrGetTPActions{TPid: "TPAcc", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpActions.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpActions.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpActions.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpActions.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpActions.Actions), len(reply.Actions)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpActions.Actions), len(reply.Actions))
	}

}

func testTPActionsRemTPAction(t *testing.T) {
	var resp string
	if err := tpActionRPC.Call("ApierV1.RemTPActions",
		&AttrGetTPActions{TPid: "TPAcc", ID: "ID"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPActionsGetTPActionAfterRemove(t *testing.T) {
	var reply *utils.TPActionPlan
	if err := tpActionRPC.Call("ApierV1.GetTPActions",
		&AttrGetTPActions{TPid: "TPAcc", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpActionDelay); err != nil {
		t.Error(err)
	}
}
