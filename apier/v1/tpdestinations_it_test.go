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
	tpDestinationCfgPath   string
	tpDestinationCfg       *config.CGRConfig
	tpDestinationRPC       *rpc.Client
	tpDestinationDataDir   = "/usr/share/cgrates"
	tpDestination          *utils.TPDestination
	tpDestinationDelay     int
	tpDestinationConfigDIR string //run tests for specific configuration
)

var sTestsTPDestinations = []func(t *testing.T){
	testTPDestinationsInitCfg,
	testTPDestinationsResetStorDb,
	testTPDestinationsStartEngine,
	testTPDestinationsRpcConn,
	testTPDestinationsGetTPDestinationBeforeSet,
	testTPDestinationsSetTPDestination,
	testTPDestinationsGetTPDestinationAfterSet,
	testTPDestinationsGetTPDestinationIds,
	testTPDestinationsUpdateTPDestination,
	testTPDestinationsGetTPDestinationAfterUpdate,
	testTPDestinationsRemTPDestination,
	testTPDestinationsGetTPDestinationAfterRemove,
	testTPDestinationsKillEngine,
}

//Test start here
func TestTPDestinationsITMySql(t *testing.T) {
	tpDestinationConfigDIR = "tutmysql"
	for _, stest := range sTestsTPDestinations {
		t.Run(tpDestinationConfigDIR, stest)
	}
}

func TestTPDestinationsITMongo(t *testing.T) {
	tpDestinationConfigDIR = "tutmongo"
	for _, stest := range sTestsTPDestinations {
		t.Run(tpDestinationConfigDIR, stest)
	}
}

func TestTPDestinationsITPG(t *testing.T) {
	tpDestinationConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPDestinations {
		t.Run(tpDestinationConfigDIR, stest)
	}
}

func testTPDestinationsInitCfg(t *testing.T) {
	var err error
	tpDestinationCfgPath = path.Join(tpDestinationDataDir, "conf", "samples", tpDestinationConfigDIR)
	tpDestinationCfg, err = config.NewCGRConfigFromFolder(tpDestinationCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpDestinationCfg.DataFolderPath = tpDestinationDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpDestinationCfg)
	switch tpDestinationConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpDestinationDelay = 2000
	default:
		tpDestinationDelay = 1000
	}
}

// Wipe out the cdr database
func testTPDestinationsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpDestinationCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPDestinationsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpDestinationCfgPath, tpDestinationDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPDestinationsRpcConn(t *testing.T) {
	var err error
	tpDestinationRPC, err = jsonrpc.Dial("tcp", tpDestinationCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPDestinationsGetTPDestinationBeforeSet(t *testing.T) {
	var reply *utils.TPDestination
	if err := tpDestinationRPC.Call("ApierV1.GetTPDestination",
		&AttrGetTPDestination{TPid: "TPD", ID: "GERMANY"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testTPDestinationsSetTPDestination(t *testing.T) {
	tpDestination = &utils.TPDestination{
		TPid:     "TPD",
		ID:       "GERMANY",
		Prefixes: []string{"+49", "+4915"},
	}
	var result string
	if err := tpDestinationRPC.Call("ApierV1.SetTPDestination", tpDestination, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDestinationsGetTPDestinationAfterSet(t *testing.T) {
	var reply *utils.TPDestination
	if err := tpDestinationRPC.Call("ApierV1.GetTPDestination",
		&AttrGetTPDestination{TPid: "TPD", ID: "GERMANY"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDestination.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpDestination.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpDestination.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpDestination.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpDestination.Prefixes), len(reply.Prefixes)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpDestination.Prefixes), len(reply.Prefixes))
	}

}

func testTPDestinationsGetTPDestinationIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"GERMANY"}
	if err := tpDestinationRPC.Call("ApierV1.GetTPDestinationIDs",
		&AttrGetTPDestinationIds{TPid: "TPD"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPDestinationsUpdateTPDestination(t *testing.T) {
	tpDestination.Prefixes = []string{"+49", "+4915", "+4916"}
	var result string
	if err := tpDestinationRPC.Call("ApierV1.SetTPDestination", tpDestination, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPDestinationsGetTPDestinationAfterUpdate(t *testing.T) {
	var reply *utils.TPDestination
	if err := tpDestinationRPC.Call("ApierV1.GetTPDestination",
		&AttrGetTPDestination{TPid: "TPD", ID: "GERMANY"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDestination.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpDestination.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpDestination.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpDestination.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpDestination.Prefixes), len(reply.Prefixes)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpDestination.Prefixes), len(reply.Prefixes))
	}

}

func testTPDestinationsRemTPDestination(t *testing.T) {
	var resp string
	if err := tpDestinationRPC.Call("ApierV1.RemTPDestination",
		&AttrGetTPDestination{TPid: "TPD", ID: "GERMANY"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPDestinationsGetTPDestinationAfterRemove(t *testing.T) {
	var reply *utils.TPDestination
	if err := tpDestinationRPC.Call("ApierV1.GetTPDestination",
		&AttrGetTPDestination{TPid: "TPD", ID: "GERMANY"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDestinationsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpDestinationDelay); err != nil {
		t.Error(err)
	}
}
