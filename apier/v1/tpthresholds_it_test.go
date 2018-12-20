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
	tpThresholdCfgPath   string
	tpThresholdCfg       *config.CGRConfig
	tpThresholdRPC       *rpc.Client
	tpThresholdDataDir   = "/usr/share/cgrates"
	tpThreshold          *utils.TPThreshold
	tpThresholdDelay     int
	tpThresholdConfigDIR string //run tests for specific configuration
)

var sTestsTPThreshold = []func(t *testing.T){
	testTPThreholdInitCfg,
	testTPThreholdResetStorDb,
	testTPThreholdStartEngine,
	testTPThreholdRpcConn,
	testTPThreholdGetTPThreholdBeforeSet,
	testTPThreholdSetTPThrehold,
	testTPThreholdGetTPThreholdAfterSet,
	testTPThreholdGetTPThreholdIds,
	testTPThreholdUpdateTPThrehold,
	testTPThreholdGetTPThreholdAfterUpdate,
	testTPThreholdRemTPThrehold,
	testTPThreholdGetTPThreholdAfterRemove,
	testTPThreholdKillEngine,
}

//Test start here
func TestTPThreholdITMySql(t *testing.T) {
	tpThresholdConfigDIR = "tutmysql"
	for _, stest := range sTestsTPThreshold {
		t.Run(tpThresholdConfigDIR, stest)
	}
}

func TestTPThreholdITMongo(t *testing.T) {
	tpThresholdConfigDIR = "tutmongo"
	for _, stest := range sTestsTPThreshold {
		t.Run(tpThresholdConfigDIR, stest)
	}
}

func TestTPThreholdITPG(t *testing.T) {
	tpThresholdConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPThreshold {
		t.Run(tpThresholdConfigDIR, stest)
	}
}

func testTPThreholdInitCfg(t *testing.T) {
	var err error
	tpThresholdCfgPath = path.Join(tpThresholdDataDir, "conf", "samples", tpThresholdConfigDIR)
	tpThresholdCfg, err = config.NewCGRConfigFromFolder(tpThresholdCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpThresholdCfg.DataFolderPath = tpThresholdDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpThresholdCfg)
	switch tpThresholdConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db
		tpThresholdDelay = 2000
	default:
		tpThresholdDelay = 1000
	}
}

// Wipe out the cdr database
func testTPThreholdResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpThresholdCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPThreholdStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpThresholdCfgPath, tpThresholdDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPThreholdRpcConn(t *testing.T) {
	var err error
	tpThresholdRPC, err = jsonrpc.Dial("tcp", tpThresholdCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPThreholdGetTPThreholdBeforeSet(t *testing.T) {
	var reply *utils.TPThreshold
	if err := tpThresholdRPC.Call("ApierV1.GetTPThreshold",
		AttrGetTPThreshold{TPid: "TH1", ID: "Threshold"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPThreholdSetTPThrehold(t *testing.T) {
	tpThreshold = &utils.TPThreshold{
		TPid:      "TH1",
		Tenant:    "cgrates.org",
		ID:        "Threhold",
		FilterIDs: []string{"FLTR_1", "FLTR_2"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		MinSleep:  "1s",
		Blocker:   true,
		Weight:    10,
		ActionIDs: []string{"Thresh1", "Thresh2"},
		Async:     true,
	}
	var result string
	if err := tpThresholdRPC.Call("ApierV1.SetTPThreshold", tpThreshold, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPThreholdGetTPThreholdAfterSet(t *testing.T) {
	var respond *utils.TPThreshold
	if err := tpThresholdRPC.Call("ApierV1.GetTPThreshold", &AttrGetTPThreshold{TPid: tpThreshold.TPid, ID: tpThreshold.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpThreshold, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpThreshold, respond)
	}
}

func testTPThreholdGetTPThreholdIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"Threhold"}
	if err := tpThresholdRPC.Call("ApierV1.GetTPThresholdIDs", &AttrGetTPThresholdIds{TPid: tpThreshold.TPid}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedTPID) {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedTPID)
	}
}

func testTPThreholdUpdateTPThrehold(t *testing.T) {
	var result string
	tpThreshold.FilterIDs = []string{"FLTR_1", "FLTR_2", "FLTR_3"}
	if err := tpThresholdRPC.Call("ApierV1.SetTPThreshold", tpThreshold, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPThreholdGetTPThreholdAfterUpdate(t *testing.T) {
	var respond *utils.TPThreshold
	if err := tpThresholdRPC.Call("ApierV1.GetTPThreshold", &AttrGetTPThreshold{TPid: tpThreshold.TPid, ID: tpThreshold.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpThreshold, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpThreshold, respond)
	}
}

func testTPThreholdRemTPThrehold(t *testing.T) {
	var resp string
	if err := tpThresholdRPC.Call("ApierV1.RemTPThreshold", &AttrGetTPThreshold{TPid: "TH1", ID: "Threhold"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPThreholdGetTPThreholdAfterRemove(t *testing.T) {
	var reply *utils.TPThreshold
	if err := tpThresholdRPC.Call("ApierV1.GetTPThreshold", AttrGetTPThreshold{TPid: "TH1", ID: "Threshold"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPThreholdKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpThresholdDelay); err != nil {
		t.Error(err)
	}
}
