// +build offline

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
	tpSharedGroupCfgPath   string
	tpSharedGroupCfg       *config.CGRConfig
	tpSharedGroupRPC       *rpc.Client
	tpSharedGroups         *utils.TPSharedGroups
	tpSharedGroupDelay     int
	tpSharedGroupConfigDIR string //run tests for specific configuration
)

var sTestsTPSharedGroups = []func(t *testing.T){
	testTPSharedGroupsInitCfg,
	testTPSharedGroupsResetStorDb,
	testTPSharedGroupsStartEngine,
	testTPSharedGroupsRpcConn,
	testTPSharedGroupsBeforeSet,
	testTPSharedGroupsSetSharedGroups,
	testTPSharedGroupsAfterSet,
	testTPSharedGroupsGetTPSharedGroupIds,
	testTPSharedGroupsUpdateTPShareGroups,
	testTpSharedGroupsGetTPSharedGroupsAfterUpdate,
	testTPSharedGroupsRemoveTPSharedGroups,
	testTPSharedGroupsGetTPSharedGroupsAfterRemove,
	testTPSharedGroupsKillEngine,
}

//Test start here
func TestTPSharedGroupsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpSharedGroupConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpSharedGroupConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpSharedGroupConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpSharedGroupConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPSharedGroups {
		t.Run(tpSharedGroupConfigDIR, stest)
	}
}

func testTPSharedGroupsInitCfg(t *testing.T) {
	var err error
	tpSharedGroupCfgPath = path.Join(*dataDir, "conf", "samples", tpSharedGroupConfigDIR)
	tpSharedGroupCfg, err = config.NewCGRConfigFromPath(tpSharedGroupCfgPath)
	if err != nil {
		t.Error(err)
	}
	switch tpSharedGroupConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db
		tpSharedGroupDelay = 2000
	default:
		tpSharedGroupDelay = 1000
	}
}

// Wipe out the cdr database
func testTPSharedGroupsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpSharedGroupCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPSharedGroupsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpSharedGroupCfgPath, tpSharedGroupDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPSharedGroupsRpcConn(t *testing.T) {
	var err error
	tpSharedGroupRPC, err = jsonrpc.Dial(utils.TCP, tpSharedGroupCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPSharedGroupsBeforeSet(t *testing.T) {
	var reply *utils.TPSharedGroups
	if err := tpSharedGroupRPC.Call(utils.APIerSv1GetTPSharedGroups, &AttrGetTPSharedGroups{TPid: "TPS1", ID: "Group1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSharedGroupsSetSharedGroups(t *testing.T) {
	tpSharedGroups = &utils.TPSharedGroups{
		TPid: "TPS1",
		ID:   "Group1",
		SharedGroups: []*utils.TPSharedGroup{
			&utils.TPSharedGroup{
				Account:       "AccOne",
				Strategy:      "StrategyOne",
				RatingSubject: "SubOne",
			},
			&utils.TPSharedGroup{
				Account:       "AccTow",
				Strategy:      "StrategyTwo",
				RatingSubject: "SubTwo",
			},
		},
	}
	var result string
	if err := tpSharedGroupRPC.Call(utils.APIerSv1SetTPSharedGroups, &tpSharedGroups, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSharedGroupsAfterSet(t *testing.T) {
	var respond *utils.TPSharedGroups
	if err := tpSharedGroupRPC.Call(utils.APIerSv1GetTPSharedGroups, &AttrGetTPSharedGroups{TPid: tpSharedGroups.TPid, ID: tpSharedGroups.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpSharedGroups.TPid, respond.TPid) {
		t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups.TPid, respond.TPid)
	} else if !reflect.DeepEqual(tpSharedGroups.ID, respond.ID) {
		t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups.ID, respond.ID)
	} else if !reflect.DeepEqual(len(tpSharedGroups.SharedGroups), len(respond.SharedGroups)) {
		t.Errorf("Expecting: %+v, received: %+v", len(tpSharedGroups.SharedGroups), len(respond.SharedGroups))
	}
}

func testTPSharedGroupsGetTPSharedGroupIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"Group1"}
	if err := tpSharedGroupRPC.Call(utils.APIerSv1GetTPSharedGroupIds,
		&AttrGetTPSharedGroupIds{tpSharedGroups.TPid, utils.PaginatorWithSearch{}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedTPID) {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedTPID)
	}
}

func testTPSharedGroupsUpdateTPShareGroups(t *testing.T) {
	var result string
	tpSharedGroups.SharedGroups = []*utils.TPSharedGroup{
		&utils.TPSharedGroup{
			Account:       "AccOne",
			Strategy:      "StrategyOne",
			RatingSubject: "SubOne",
		},
		&utils.TPSharedGroup{
			Account:       "AccTow",
			Strategy:      "StrategyTwo",
			RatingSubject: "SubTwo",
		},
		&utils.TPSharedGroup{
			Account:       "AccPlus",
			Strategy:      "StrategyPlus",
			RatingSubject: "SubPlus",
		},
	}
	if err := tpSharedGroupRPC.Call(utils.APIerSv1SetTPSharedGroups, tpSharedGroups, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTpSharedGroupsGetTPSharedGroupsAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPSharedGroups
	if err := tpSharedGroupRPC.Call(utils.APIerSv1GetTPSharedGroups, &AttrGetTPSharedGroups{TPid: tpSharedGroups.TPid, ID: tpSharedGroups.ID}, &expectedTPS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpSharedGroups.TPid, expectedTPS.TPid) {
		t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups.TPid, expectedTPS.TPid)
	} else if !reflect.DeepEqual(tpSharedGroups.ID, expectedTPS.ID) {
		t.Errorf("Expecting: %+v, received: %+v", tpSharedGroups.ID, expectedTPS.ID)
	} else if !reflect.DeepEqual(len(tpSharedGroups.SharedGroups), len(expectedTPS.SharedGroups)) {
		t.Errorf("Expecting: %+v, received: %+v", len(tpSharedGroups.SharedGroups), len(expectedTPS.SharedGroups))
	}
}

func testTPSharedGroupsRemoveTPSharedGroups(t *testing.T) {
	var resp string
	if err := tpSharedGroupRPC.Call(utils.APIerSv1RemoveTPSharedGroups, &AttrGetTPSharedGroups{TPid: tpSharedGroups.TPid, ID: tpSharedGroups.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPSharedGroupsGetTPSharedGroupsAfterRemove(t *testing.T) {
	var reply *utils.TPSharedGroups
	if err := tpSharedGroupRPC.Call(utils.APIerSv1GetTPSharedGroups, &AttrGetTPSharedGroups{TPid: "TPS1", ID: "Group1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSharedGroupsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpSharedGroupDelay); err != nil {
		t.Error(err)
	}
}
