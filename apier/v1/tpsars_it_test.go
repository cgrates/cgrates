//go:build offline
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
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpSarCfgPath   string
	tpSarCfg       *config.CGRConfig
	tpSarRPC       *birpc.Client
	tpSar          *utils.TPSarsProfile
	tpSarDelay     int
	tpSarConfigDIR string //run tests for specific configuration
)

var sTestsTPSars = []func(t *testing.T){
	testTPSarsInitCfg,
	testTPSarsResetStorDb,
	testTPSarsStartEngine,
	testTPSarsRpcConn,
	testTPSarsGetTPSarBeforeSet,
	testTPSarsSetTPSar,
	testTPSarsGetTPSarAfterSet,
	testTPSarsUpdateTPSar,
	testTPSarsGetTPSarAfterUpdate,
	testTPSarsRemoveTPSar,
	testTPSarsGetTPSarAfterRemove,
	testTPSarsKillEngine,
}

// Test start here
func TestTPSarIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpSarConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpSarConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpSarConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpSarConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPSars {
		t.Run(tpSarConfigDIR, stest)
	}
}

func testTPSarsInitCfg(t *testing.T) {
	var err error
	tpSarCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpSarConfigDIR)
	tpSarCfg, err = config.NewCGRConfigFromPath(tpSarCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpSarDelay = 1000
}

// Wipe out the cdr database
func testTPSarsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpSarCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPSarsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpSarCfgPath, tpSarDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPSarsRpcConn(t *testing.T) {
	var err error
	tpSarRPC, err = jsonrpc.Dial(utils.TCP, tpSarCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPSarsGetTPSarBeforeSet(t *testing.T) {
	var reply *utils.TPSarsProfile
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1GetTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sar1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSarsSetTPSar(t *testing.T) {
	tpSar = &utils.TPSarsProfile{
		Tenant:        "cgrates.org",
		TPid:          "TPS1",
		ID:            "Sar1",
		QueryInterval: "1m",
		ThresholdIDs:  []string{"ThreshValue", "ThreshValueTwo"},
	}
	sort.Strings(tpSar.ThresholdIDs)
	var result string
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1SetTPSar, tpSar, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSarsGetTPSarAfterSet(t *testing.T) {
	var respond *utils.TPSarsProfile
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1GetTPSar,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sar1"}, &respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpSar, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpSar, respond)
	}
}

func testTPSarsUpdateTPSar(t *testing.T) {
	var result string
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1SetTPSar, tpSar, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSarsGetTPSarAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPSarsProfile
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1GetTPSar,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sar1"}, &expectedTPS); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPS.ThresholdIDs)
	if !reflect.DeepEqual(tpSar, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpSar, expectedTPS)
	}
}

func testTPSarsRemoveTPSar(t *testing.T) {
	var resp string
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1RemoveTPSar,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sar1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPSarsGetTPSarAfterRemove(t *testing.T) {
	var respond *utils.TPSarsProfile
	if err := tpSarRPC.Call(context.Background(), utils.APIerSv1GetTPSar,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sar1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSarsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpSarDelay); err != nil {
		t.Error(err)
	}
}
