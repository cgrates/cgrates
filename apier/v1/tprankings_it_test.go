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
	tpRankingCfgPath   string
	tpRankingCfg       *config.CGRConfig
	tpRankingRPC       *birpc.Client
	tpRanking          *utils.TPRankingProfile
	tpRankingDelay     int
	tpRankingConfigDIR string //run tests for specific configuration
)

var sTestsTPRankings = []func(t *testing.T){
	testTPRankingsInitCfg,
	testTPRankingsResetStorDb,
	testTPRankingsStartEngine,
	testTPRankingsRpcConn,
	testTPRankingsGetTPRankingBeforeSet,
	testTPRankingsSetTPRanking,
	testTPRankingsGetTPRankingAfterSet,
	testTPRankingsUpdateTPRanking,
	testTPRankingsGetTPRankingAfterUpdate,
	testTPRankingsRemoveTPRanking,
	testTPRankingsGetTPRankingAfterRemove,
	testTPRankingsKillEngine,
}

// Test start here
func TestTPRankingIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tpRankingConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRankingConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRankingConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpRankingConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRankings {
		t.Run(tpRankingConfigDIR, stest)
	}
}

func testTPRankingsInitCfg(t *testing.T) {
	var err error
	tpRankingCfgPath = path.Join(*utils.DataDir, "conf", "samples", tpRankingConfigDIR)
	tpRankingCfg, err = config.NewCGRConfigFromPath(tpRankingCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRankingDelay = 1000
}

// Wipe out the cdr database
func testTPRankingsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRankingCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRankingsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRankingCfgPath, tpRankingDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRankingsRpcConn(t *testing.T) {
	var err error
	tpRankingRPC, err = jsonrpc.Dial(utils.TCP, tpRankingCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRankingsGetTPRankingBeforeSet(t *testing.T) {
	var reply *utils.TPRankingProfile
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1GetTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Ranking1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRankingsSetTPRanking(t *testing.T) {
	tpRanking = &utils.TPRankingProfile{
		Tenant:       "cgrates.org",
		TPid:         "TPS1",
		ID:           "Ranking1",
		Schedule:     "1m",
		Sorting:      "*asc",
		StatIDs:      []string{"Stat1"},
		ThresholdIDs: []string{"ThreshValue", "ThreshValueTwo"},
	}
	sort.Strings(tpRanking.ThresholdIDs)
	var result string
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1SetTPRanking, tpRanking, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRankingsGetTPRankingAfterSet(t *testing.T) {
	var respond *utils.TPRankingProfile
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1GetTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Ranking1"}, &respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpRanking, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpRanking, respond)
	}
}

func testTPRankingsUpdateTPRanking(t *testing.T) {
	var result string

	tpRanking.MetricIDs = []string{"*sum", "*average"}
	sort.Strings(tpRanking.MetricIDs)
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1SetTPRanking, tpRanking, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRankingsGetTPRankingAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPRankingProfile
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1GetTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Ranking1"}, &expectedTPS); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPS.ThresholdIDs)
	sort.Strings(expectedTPS.MetricIDs)
	if !reflect.DeepEqual(tpRanking, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpRanking, expectedTPS)
	}
}

func testTPRankingsRemoveTPRanking(t *testing.T) {
	var resp string
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1RemoveTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Ranking1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRankingsGetTPRankingAfterRemove(t *testing.T) {
	var respond *utils.TPRankingProfile
	if err := tpRankingRPC.Call(context.Background(), utils.APIerSv1GetTPRanking,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Ranking1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRankingsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRankingDelay); err != nil {
		t.Error(err)
	}
}
