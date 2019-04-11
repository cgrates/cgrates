// +build integration

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
	tpDispatcherCfgPath   string
	tpDispatcherCfg       *config.CGRConfig
	tpDispatcherRPC       *rpc.Client
	tpDispatcherDataDir   = "/usr/share/cgrates"
	tpDispatcher          *utils.TPDispatcherProfile
	tpDispatcherDelay     int
	tpDispatcherConfigDIR string //run tests for specific configuration
)

var sTestsTPDispatchers = []func(t *testing.T){
	testTPDispatcherInitCfg,
	testTPDispatcherResetStorDb,
	testTPDispatcherStartEngine,
	testTPDispatcherRpcConn,
	ttestTPDispatcherGetTPDispatcherBeforeSet,
	testTPDispatcherSetTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterSet,
	testTPDispatcherGetFilterIds,
	testTPDispatcherUpdateTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterUpdate,
	testTPDispatcherRemTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterRemove,
	testTPDispatcherKillEngine,
}

//Test start here
func TestTPDispatcherITMySql(t *testing.T) {
	tpDispatcherConfigDIR = "tutmysql"
	for _, stest := range sTestsTPDispatchers {
		t.Run(tpDispatcherConfigDIR, stest)
	}
}

func TestTPDispatcherITMongo(t *testing.T) {
	tpDispatcherConfigDIR = "tutmongo"
	for _, stest := range sTestsTPDispatchers {
		t.Run(tpDispatcherConfigDIR, stest)
	}
}

func testTPDispatcherInitCfg(t *testing.T) {
	var err error
	tpDispatcherCfgPath = path.Join(tpDispatcherDataDir, "conf", "samples", tpDispatcherConfigDIR)
	tpDispatcherCfg, err = config.NewCGRConfigFromPath(tpDispatcherCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpDispatcherCfg.DataFolderPath = tpDispatcherDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpDispatcherCfg)
	tpDispatcherDelay = 1000

}

// Wipe out the cdr database
func testTPDispatcherResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpDispatcherCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPDispatcherStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpDispatcherCfgPath, tpDispatcherDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPDispatcherRpcConn(t *testing.T) {
	var err error
	tpDispatcherRPC, err = jsonrpc.Dial("tcp", tpDispatcherCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func ttestTPDispatcherGetTPDispatcherBeforeSet(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call("ApierV1.GetTPDispatcherProfile",
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDispatcherSetTPDispatcher(t *testing.T) {
	tpDispatcher = &utils.TPDispatcherProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:Account:1002"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Strategy: utils.MetaFirst,
		Weight:   10,
	}

	var result string
	if err := tpDispatcherRPC.Call("ApierV1.SetTPDispatcherProfile", tpDispatcher, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDispatcherGetTPDispatcherAfterSet(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call("ApierV1.GetTPDispatcherProfile",
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDispatcher, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpDispatcher, reply)
	}
}

func testTPDispatcherGetFilterIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"Dsp1"}
	if err := tpDispatcherRPC.Call("ApierV1.GetTPDispatcherProfileIDs",
		&AttrGetTPDispatcherIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPDispatcherUpdateTPDispatcher(t *testing.T) {
	var result string
	if err := tpDispatcherRPC.Call("ApierV1.SetTPDispatcherProfile", tpDispatcher, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDispatcherGetTPDispatcherAfterUpdate(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	revHosts := &utils.TPDispatcherProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:Account:1002"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Strategy: utils.MetaFirst,
		Weight:   10,
	}
	if err := tpDispatcherRPC.Call("ApierV1.GetTPDispatcherProfile",
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDispatcher, reply) && !reflect.DeepEqual(revHosts, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpDispatcher), utils.ToJSON(reply))
	}
}

func testTPDispatcherRemTPDispatcher(t *testing.T) {
	var resp string
	if err := tpDispatcherRPC.Call("ApierV1.RemoveTPDispatcherProfile",
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPDispatcherGetTPDispatcherAfterRemove(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call("ApierV1.GetTPDispatcherProfile",
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDispatcherKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpDispatcherDelay); err != nil {
		t.Error(err)
	}
}
