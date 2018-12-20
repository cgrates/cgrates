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
	tpResCfgPath   string
	tpResCfg       *config.CGRConfig
	tpResRPC       *rpc.Client
	tpResDataDir   = "/usr/share/cgrates"
	tpRes          *utils.TPResource
	tpResDelay     int
	tpResConfigDIR string //run tests for specific configuration
)

var sTestsTPResources = []func(t *testing.T){
	testTPResInitCfg,
	testTPResResetStorDb,
	testTPResStartEngine,
	testTPResRpcConn,
	testTPResGetTPResourceBeforeSet,
	testTPResSetTPResource,
	testTPResGetTPResourceAfterSet,
	testTPResUpdateTPResource,
	testTPResGetTPResourceAfterUpdate,
	testTPResRemTPResource,
	testTPResGetTPResourceAfterRemove,
	testTPResKillEngine,
}

//Test start here
func TestTPResITMySql(t *testing.T) {
	tpResConfigDIR = "tutmysql"
	for _, stest := range sTestsTPResources {
		t.Run(tpResConfigDIR, stest)
	}
}

func TestTPResITMongo(t *testing.T) {
	tpResConfigDIR = "tutmongo"
	for _, stest := range sTestsTPResources {
		t.Run(tpResConfigDIR, stest)
	}
}

func TestTPResITPG(t *testing.T) {
	tpResConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPResources {
		t.Run(tpResConfigDIR, stest)
	}
}

func testTPResInitCfg(t *testing.T) {
	var err error
	tpResCfgPath = path.Join(tpResDataDir, "conf", "samples", tpResConfigDIR)
	tpResCfg, err = config.NewCGRConfigFromFolder(tpResCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpResCfg.DataFolderPath = tpResDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpResCfg)
	switch tpResConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpResDelay = 2000
	default:
		tpResDelay = 1000
	}
}

// Wipe out the cdr database
func testTPResResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpResCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPResStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpResCfgPath, tpResDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPResRpcConn(t *testing.T) {
	var err error
	tpResRPC, err = jsonrpc.Dial("tcp", tpResCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPResGetTPResourceBeforeSet(t *testing.T) {
	var reply *utils.TPResource
	if err := tpResRPC.Call("ApierV1.GetTPResource", AttrGetTPResource{TPid: "TPR1", ID: "ResGroup1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPResSetTPResource(t *testing.T) {
	tpRes = &utils.TPResource{
		Tenant:    "cgrates.org",
		TPid:      "TPR1",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		UsageTTL:          "1s",
		Limit:             "7",
		AllocationMessage: "",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"ValOne", "ValTwo"},
	}
	var result string
	if err := tpResRPC.Call("ApierV1.SetTPResource", tpRes, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPResGetTPResourceAfterSet(t *testing.T) {
	var respond *utils.TPResource
	if err := tpResRPC.Call("ApierV1.GetTPResource", AttrGetTPResource{TPid: tpRes.TPid, ID: tpRes.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRes, respond) {
		t.Errorf("Expecting : %+v, received: %+v", tpRes, respond)
	}
}

func testTPResUpdateTPResource(t *testing.T) {
	var result string
	tpRes.FilterIDs = []string{"FLTR_1", "FLTR_STS1"}
	if err := tpResRPC.Call("ApierV1.SetTPResource", tpRes, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPResGetTPResourceAfterUpdate(t *testing.T) {
	var expectedTPR *utils.TPResource
	if err := tpResRPC.Call("ApierV1.GetTPResource", AttrGetTPResource{TPid: tpRes.TPid, ID: tpRes.ID}, &expectedTPR); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRes, expectedTPR) {
		t.Errorf("Expecting: %+v, received: %+v", tpRes, expectedTPR)
	}
}

func testTPResRemTPResource(t *testing.T) {
	var resp string
	if err := tpResRPC.Call("ApierV1.RemTPResource", AttrGetTPResource{TPid: tpRes.TPid, ID: tpRes.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPResGetTPResourceAfterRemove(t *testing.T) {
	var respond *utils.TPResource
	if err := tpResRPC.Call("ApierV1.GetTPResource", AttrGetTPStat{TPid: "TPS1", ID: "ResGroup1"}, &respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPResKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpResDelay); err != nil {
		t.Error(err)
	}
}
