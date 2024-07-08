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
	tpTrendCfgPath   string
	tpTrendCfg       *config.CGRConfig
	tpTrendRPC       *birpc.Client
	tpTrend          *utils.TPTrendsProfile
	tpTrendDelay     int
	tpTrendConfigDIR string //run tests for specific configuration
)

var sTestsTPTrends = []func(t *testing.T){
	testTPTrendsInitCfg,
	testTPTrendsResetStorDb,
	testTPTrendsStartEngine,
	testTPTrendsRpcConn,
	testTPTrendsGetTPTrendBeforeSet,
	testTPTrendsSetTPTrend,
	testTPTrendsGetTPTrendAfterSet,
	testTPTrendsUpdateTPTrend,
	testTPTrendsGetTPTrendAfterUpdate,
	testTPTrendsRemoveTPTrend,
	testTPTrendsGetTPTrendAfterRemove,
	testTPTrendsKillEngine,
}

// Test start here
func TestTPTrendIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpTrendConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpTrendConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpTrendConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpTrendConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPTrends {
		t.Run(tpTrendConfigDIR, stest)
	}
}

func testTPTrendsInitCfg(t *testing.T) {
	var err error
	tpTrendCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpTrendConfigDIR)
	tpTrendCfg, err = config.NewCGRConfigFromPath(tpTrendCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpTrendDelay = 1000
}

// Wipe out the cdr database
func testTPTrendsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpTrendCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPTrendsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpTrendCfgPath, tpTrendDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client
func testTPTrendsRpcConn(t *testing.T) {
	var err error
	tpTrendRPC, err = jsonrpc.Dial(utils.TCP, tpTrendCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPTrendsGetTPTrendBeforeSet(t *testing.T) {
	var reply *utils.TPTrendsProfile
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1GetTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Trend1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPTrendsSetTPTrend(t *testing.T) {
	tpTrend = &utils.TPTrendsProfile{
		Tenant:        "cgrates.org",
		TPid:          "TPS1",
		ID:            "Trend1",
		QueryInterval: "1m",
		ThresholdIDs:  []string{"ThreshValue", "ThreshValueTwo"},
	}
	sort.Strings(tpTrend.ThresholdIDs)
	var result string
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1SetTPTrend, tpTrend, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPTrendsGetTPTrendAfterSet(t *testing.T) {
	var respond *utils.TPTrendsProfile
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1GetTPTrend,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Trend1"}, &respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpTrend, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpTrend, respond)
	}
}

func testTPTrendsUpdateTPTrend(t *testing.T) {
	var result string
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1SetTPTrend, tpTrend, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPTrendsGetTPTrendAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPTrendsProfile
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1GetTPTrend,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Trend1"}, &expectedTPS); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPS.ThresholdIDs)
	if !reflect.DeepEqual(tpTrend, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpTrend, expectedTPS)
	}
}

func testTPTrendsRemoveTPTrend(t *testing.T) {
	var resp string
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1RemoveTPTrend,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Trend1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPTrendsGetTPTrendAfterRemove(t *testing.T) {
	var respond *utils.TPTrendsProfile
	if err := tpTrendRPC.Call(context.Background(), utils.APIerSv1GetTPTrend,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Trend1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPTrendsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpTrendDelay); err != nil {
		t.Error(err)
	}
}
