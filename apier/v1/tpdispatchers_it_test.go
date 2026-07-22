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
	tpDispatcherCfgPath   string
	tpDispatcherCfg       *config.CGRConfig
	tpDispatcherRPC       *birpc.Client
	tpDispatcher          *utils.TPDispatcherProfile
	tpDispatcherDelay     int
	tpDispatcherConfigDIR string //run tests for specific configuration
)

var sTestsTPDispatchers = []func(t *testing.T){
	testTPDispatcherInitCfg,
	testTPDispatcherResetStorDb,
	testTPDispatcherStartEngine,
	testTPDispatcherRpcConn,
	testTPDispatcherGetTPDispatcherBeforeSet,
	testTPDispatcherSetTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterSet,
	testTPDispatcherGetTPDispatcherIds,
	testTPDispatcherUpdateTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterUpdate,
	testTPDispatcherRemTPDispatcher,
	testTPDispatcherGetTPDispatcherAfterRemove,
	testTPDispatcherKillEngine,
}

// Test start here
func TestTPDispatcherIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpDispatcherConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpDispatcherConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpDispatcherConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPDispatchers {
		t.Run(tpDispatcherConfigDIR, stest)
	}
}

func testTPDispatcherInitCfg(t *testing.T) {
	var err error
	tpDispatcherCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpDispatcherConfigDIR)
	tpDispatcherCfg, err = config.NewCGRConfigFromPath(tpDispatcherCfgPath)
	if err != nil {
		t.Error(err)
	}
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
	tpDispatcherRPC, err = jsonrpc.Dial(utils.TCP, tpDispatcherCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPDispatcherGetTPDispatcherBeforeSet(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1GetTPDispatcherProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDispatcherSetTPDispatcher(t *testing.T) {
	tpDispatcher = &utils.TPDispatcherProfile{
		TPid:       "TP1",
		Tenant:     "cgrates.org",
		ID:         "Dsp1",
		FilterIDs:  []string{"*string:Account:1002"},
		Subsystems: []string{"testSys"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Strategy: utils.MetaFirst,
		Weight:   10,
	}

	var result string
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1SetTPDispatcherProfile, tpDispatcher, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDispatcherGetTPDispatcherAfterSet(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1GetTPDispatcherProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDispatcher, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpDispatcher), utils.ToJSON(reply))
	}
}

func testTPDispatcherGetTPDispatcherIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:Dsp1"}
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1GetTPDispatcherProfileIDs,
		&AttrGetTPDispatcherIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPDispatcherUpdateTPDispatcher(t *testing.T) {
	var result string
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1SetTPDispatcherProfile, tpDispatcher, &result); err != nil {
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
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1GetTPDispatcherProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDispatcher, reply) && !reflect.DeepEqual(revHosts, reply) {
		t.Errorf("Expecting : %+v \n and %+v\n, received: %+v", utils.ToJSON(tpDispatcher), utils.ToJSON(revHosts), utils.ToJSON(reply))
	}
}

func testTPDispatcherRemTPDispatcher(t *testing.T) {
	var resp string
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1RemoveTPDispatcherProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Dsp1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPDispatcherGetTPDispatcherAfterRemove(t *testing.T) {
	var reply *utils.TPDispatcherProfile
	if err := tpDispatcherRPC.Call(context.Background(), utils.APIerSv1GetTPDispatcherProfile,
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
