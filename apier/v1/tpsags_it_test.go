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
	tpSagCfgPath   string
	tpSagCfg       *config.CGRConfig
	tpSagRPC       *birpc.Client
	tpSag          *utils.TPSagsProfile
	tpSagDelay     int
	tpSagConfigDIR string //run tests for specific configuration
)

var sTestsTPSags = []func(t *testing.T){
	testTPSagsInitCfg,
	testTPSagsResetStorDb,
	testTPSagsStartEngine,
	testTPSagsRpcConn,
	testTPSagsGetTPSagBeforeSet,
	testTPSagsSetTPSag,
	testTPSagsGetTPSagAfterSet,
	testTPSagsUpdateTPSag,
	testTPSagsGetTPSagAfterUpdate,
	testTPSagsRemoveTPSag,
	testTPSagsGetTPSagAfterRemove,
	testTPSagsKillEngine,
}

// Test start here
func TestTPSagIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpSagConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpSagConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpSagConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpSagConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPSags {
		t.Run(tpSagConfigDIR, stest)
	}
}

func testTPSagsInitCfg(t *testing.T) {
	var err error
	tpSagCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpSagConfigDIR)
	tpSagCfg, err = config.NewCGRConfigFromPath(tpSagCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpSagDelay = 1000
}

// Wipe out the cdr database
func testTPSagsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpSagCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPSagsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpSagCfgPath, tpSagDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPSagsRpcConn(t *testing.T) {
	var err error
	tpSagRPC, err = jsonrpc.Dial(utils.TCP, tpSagCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPSagsGetTPSagBeforeSet(t *testing.T) {
	var reply *utils.TPSagsProfile
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1GetTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sag1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSagsSetTPSag(t *testing.T) {
	tpSag = &utils.TPSagsProfile{
		Tenant:        "cgrates.org",
		TPid:          "TPS1",
		ID:            "Sag1",
		QueryInterval: "1m",
		Sorting:       "*asc",
		StatIDs:       []string{"Stat1"},
		ThresholdIDs:  []string{"ThreshValue", "ThreshValueTwo"},
	}
	sort.Strings(tpSag.ThresholdIDs)
	var result string
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1SetTPSag, tpSag, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSagsGetTPSagAfterSet(t *testing.T) {
	var respond *utils.TPSagsProfile
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1GetTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sag1"}, &respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpSag, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpSag, respond)
	}
}

func testTPSagsUpdateTPSag(t *testing.T) {
	var result string

	tpSag.MetricIDs = []string{"*sum", "*average"}
	sort.Strings(tpSag.MetricIDs)
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1SetTPSag, tpSag, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSagsGetTPSagAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPSagsProfile
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1GetTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sag1"}, &expectedTPS); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPS.ThresholdIDs)
	sort.Strings(expectedTPS.MetricIDs)
	if !reflect.DeepEqual(tpSag, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpSag, expectedTPS)
	}
}

func testTPSagsRemoveTPSag(t *testing.T) {
	var resp string
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1RemoveTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sag1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPSagsGetTPSagAfterRemove(t *testing.T) {
	var respond *utils.TPSagsProfile
	if err := tpSagRPC.Call(context.Background(), utils.APIerSv1GetTPSag,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Sag1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSagsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpSagDelay); err != nil {
		t.Error(err)
	}
}
