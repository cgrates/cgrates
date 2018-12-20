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
	tpDerivedChargersCfgPath   string
	tpDerivedChargersCfg       *config.CGRConfig
	tpDerivedChargersRPC       *rpc.Client
	tpDerivedChargersDataDir   = "/usr/share/cgrates"
	tpDerivedChargers          *utils.TPDerivedChargers
	tpDerivedChargersDelay     int
	tpDerivedChargersConfigDIR string //run tests for specific configuration
	tpDerivedChargersID        = "LoadID:*out:cgrates.org:call:1001:1001"
)

var sTestsTPDerivedChargers = []func(t *testing.T){
	testTPDerivedChargersInitCfg,
	testTPDerivedChargersResetStorDb,
	testTPDerivedChargersStartEngine,
	testTPDerivedChargersRpcConn,
	testTPDerivedChargersGetTPDerivedChargersBeforeSet,
	testTPDerivedChargersSetTPDerivedChargers,
	testTPDerivedChargersGetTPDerivedChargersAfterSet,
	testTPDerivedChargersGetTPDerivedChargerIds,
	testTPDerivedChargersUpdateTPDerivedChargers,
	testTPDerivedChargersGetTPDerivedChargersAfterUpdate,
	testTPDerivedChargersRemTPDerivedChargers,
	testTPDerivedChargersGetTPDerivedChargersAfterRemove,
	testTPDerivedChargersKillEngine,
}

//Test start here
func TestTPDerivedChargersITMySql(t *testing.T) {
	tpDerivedChargersConfigDIR = "tutmysql"
	for _, stest := range sTestsTPDerivedChargers {
		t.Run(tpDerivedChargersConfigDIR, stest)
	}
}

func TestTPDerivedChargersITMongo(t *testing.T) {
	tpDerivedChargersConfigDIR = "tutmongo"
	for _, stest := range sTestsTPDerivedChargers {
		t.Run(tpDerivedChargersConfigDIR, stest)
	}
}

func TestTPDerivedChargersITPG(t *testing.T) {
	tpDerivedChargersConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPDerivedChargers {
		t.Run(tpDerivedChargersConfigDIR, stest)
	}
}

func testTPDerivedChargersInitCfg(t *testing.T) {
	var err error
	tpDerivedChargersCfgPath = path.Join(tpDerivedChargersDataDir, "conf", "samples", tpDerivedChargersConfigDIR)
	tpDerivedChargersCfg, err = config.NewCGRConfigFromFolder(tpDerivedChargersCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpDerivedChargersCfg.DataFolderPath = tpDerivedChargersDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpDerivedChargersCfg)
	switch tpDerivedChargersConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpDerivedChargersDelay = 2000
	default:
		tpDerivedChargersDelay = 1000
	}
}

// Wipe out the cdr database
func testTPDerivedChargersResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpDerivedChargersCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPDerivedChargersStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpDerivedChargersCfgPath, tpDerivedChargersDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPDerivedChargersRpcConn(t *testing.T) {
	var err error
	tpDerivedChargersRPC, err = jsonrpc.Dial("tcp", tpDerivedChargersCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPDerivedChargersGetTPDerivedChargersBeforeSet(t *testing.T) {
	var reply *utils.TPDerivedChargers
	if err := tpDerivedChargersRPC.Call("ApierV1.GetTPDerivedChargers",
		&AttrGetTPDerivedChargers{TPid: "TPD", DerivedChargersId: tpDerivedChargersID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testTPDerivedChargersSetTPDerivedChargers(t *testing.T) {
	tpDerivedChargers = &utils.TPDerivedChargers{
		TPid:           "TPD",
		LoadId:         "LoadID",
		Direction:      "*out",
		Tenant:         "cgrates.org",
		Category:       "call",
		Account:        "1001",
		Subject:        "1001",
		DestinationIds: "",
		DerivedChargers: []*utils.TPDerivedCharger{
			&utils.TPDerivedCharger{
				RunId:                "derived_run1",
				RunFilters:           "",
				ReqTypeField:         "^*rated",
				DirectionField:       "*default",
				TenantField:          "*default",
				CategoryField:        "*default",
				AccountField:         "*default",
				SubjectField:         "^1002",
				DestinationField:     "*default",
				SetupTimeField:       "*default",
				PddField:             "*default",
				AnswerTimeField:      "*default",
				UsageField:           "*default",
				SupplierField:        "*default",
				DisconnectCauseField: "*default",
				CostField:            "*default",
				RatedField:           "*default",
			},
		},
	}
	var result string
	if err := tpDerivedChargersRPC.Call("ApierV1.SetTPDerivedChargers", tpDerivedChargers, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDerivedChargersGetTPDerivedChargersAfterSet(t *testing.T) {
	var reply *utils.TPDerivedChargers
	if err := tpDerivedChargersRPC.Call("ApierV1.GetTPDerivedChargers",
		&AttrGetTPDerivedChargers{TPid: "TPD", DerivedChargersId: tpDerivedChargersID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDerivedChargers.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpDerivedChargers.LoadId, reply.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.LoadId, reply.LoadId)
	} else if !reflect.DeepEqual(tpDerivedChargers.Direction, reply.Direction) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Direction, reply.Direction)
	} else if !reflect.DeepEqual(tpDerivedChargers.Tenant, reply.Tenant) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Tenant, reply.Tenant)
	} else if !reflect.DeepEqual(tpDerivedChargers.Category, reply.Category) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Category, reply.Category)
	} else if !reflect.DeepEqual(tpDerivedChargers.Account, reply.Account) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Account, reply.Account)
	} else if !reflect.DeepEqual(tpDerivedChargers.Subject, reply.Subject) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Subject, reply.Subject)
	} else if !reflect.DeepEqual(tpDerivedChargers.DestinationIds, reply.DestinationIds) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.DestinationIds, reply.DestinationIds)
	} else if !reflect.DeepEqual(len(tpDerivedChargers.DerivedChargers), len(reply.DerivedChargers)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpDerivedChargers.DerivedChargers), len(reply.DerivedChargers))
	}

}

func testTPDerivedChargersGetTPDerivedChargerIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"LoadID:*out:cgrates.org:call:1001:1001"}
	if err := tpDerivedChargersRPC.Call("ApierV1.GetTPDerivedChargerIds",
		&AttrGetTPDerivedChargeIds{TPid: "TPD"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPDerivedChargersUpdateTPDerivedChargers(t *testing.T) {
	tpDerivedChargers.DerivedChargers = []*utils.TPDerivedCharger{
		&utils.TPDerivedCharger{
			RunId:                "derived_run1",
			RunFilters:           "",
			ReqTypeField:         "^*rated",
			DirectionField:       "*default",
			TenantField:          "*default",
			CategoryField:        "*default",
			AccountField:         "*default",
			SubjectField:         "^1002",
			DestinationField:     "*default",
			SetupTimeField:       "*default",
			PddField:             "*default",
			AnswerTimeField:      "*default",
			UsageField:           "*default",
			SupplierField:        "*default",
			DisconnectCauseField: "*default",
			CostField:            "*default",
			RatedField:           "*default",
		},
		&utils.TPDerivedCharger{
			RunId:                "derived_run2",
			RunFilters:           "",
			ReqTypeField:         "^*rated",
			DirectionField:       "*default",
			TenantField:          "*default",
			CategoryField:        "*default",
			AccountField:         "*default",
			SubjectField:         "^1003",
			DestinationField:     "*default",
			SetupTimeField:       "*default",
			PddField:             "*default",
			AnswerTimeField:      "*default",
			UsageField:           "*default",
			SupplierField:        "*default",
			DisconnectCauseField: "*default",
			CostField:            "*default",
			RatedField:           "*default",
		},
	}
	var result string
	if err := tpDerivedChargersRPC.Call("ApierV1.SetTPDerivedChargers", tpDerivedChargers, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPDerivedChargersGetTPDerivedChargersAfterUpdate(t *testing.T) {
	var reply *utils.TPDerivedChargers
	if err := tpDerivedChargersRPC.Call("ApierV1.GetTPDerivedChargers",
		&AttrGetTPDerivedChargers{TPid: "TPD", DerivedChargersId: tpDerivedChargersID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDerivedChargers.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpDerivedChargers.LoadId, reply.LoadId) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.LoadId, reply.LoadId)
	} else if !reflect.DeepEqual(tpDerivedChargers.Direction, reply.Direction) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Direction, reply.Direction)
	} else if !reflect.DeepEqual(tpDerivedChargers.Tenant, reply.Tenant) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Tenant, reply.Tenant)
	} else if !reflect.DeepEqual(tpDerivedChargers.Category, reply.Category) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Category, reply.Category)
	} else if !reflect.DeepEqual(tpDerivedChargers.Account, reply.Account) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Account, reply.Account)
	} else if !reflect.DeepEqual(tpDerivedChargers.Subject, reply.Subject) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.Subject, reply.Subject)
	} else if !reflect.DeepEqual(tpDerivedChargers.DestinationIds, reply.DestinationIds) {
		t.Errorf("Expecting : %+v, received: %+v", tpDerivedChargers.DestinationIds, reply.DestinationIds)
	} else if !reflect.DeepEqual(len(tpDerivedChargers.DerivedChargers), len(reply.DerivedChargers)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpDerivedChargers.DerivedChargers), len(reply.DerivedChargers))
	}

}

func testTPDerivedChargersRemTPDerivedChargers(t *testing.T) {
	var resp string
	if err := tpDerivedChargersRPC.Call("ApierV1.RemTPDerivedChargers",
		&AttrGetTPDerivedChargers{TPid: "TPD", DerivedChargersId: tpDerivedChargersID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPDerivedChargersGetTPDerivedChargersAfterRemove(t *testing.T) {
	var reply *utils.TPDerivedChargers
	if err := tpDerivedChargersRPC.Call("ApierV1.GetTPDerivedChargers",
		&AttrGetTPDerivedChargers{TPid: "TPD", DerivedChargersId: tpDerivedChargersID}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDerivedChargersKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpDerivedChargersDelay); err != nil {
		t.Error(err)
	}
}
