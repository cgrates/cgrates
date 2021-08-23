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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpStatCfgPath   string
	tpStatCfg       *config.CGRConfig
	tpStatRPC       *rpc.Client
	tpStat          *utils.TPStatProfile
	tpStatDelay     int
	tpStatConfigDIR string //run tests for specific configuration
)

var sTestsTPStats = []func(t *testing.T){
	testTPStatsInitCfg,
	testTPStatsResetStorDb,
	testTPStatsStartEngine,
	testTPStatsRpcConn,
	testTPStatsGetTPStatBeforeSet,
	testTPStatsSetTPStat,
	testTPStatsGetTPStatAfterSet,
	testTPStatsUpdateTPStat,
	testTPStatsGetTPStatAfterUpdate,
	testTPStatsRemoveTPStat,
	testTPStatsGetTPStatAfterRemove,
	testTPStatsKillEngine,
}

//Test start here
func TestTPStatIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpStatConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpStatConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpStatConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpStatConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPStats {
		t.Run(tpStatConfigDIR, stest)
	}
}

func testTPStatsInitCfg(t *testing.T) {
	var err error
	tpStatCfgPath = path.Join(*dataDir, "conf", "samples", tpStatConfigDIR)
	tpStatCfg, err = config.NewCGRConfigFromPath(tpStatCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpStatDelay = 1000
}

// Wipe out the cdr database
func testTPStatsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpStatCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPStatsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpStatCfgPath, tpStatDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPStatsRpcConn(t *testing.T) {
	var err error
	tpStatRPC, err = jsonrpc.Dial(utils.TCP, tpStatCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPStatsGetTPStatBeforeSet(t *testing.T) {
	var reply *utils.TPStatProfile
	if err := tpStatRPC.Call(utils.APIerSv1GetTPStat,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Stat1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPStatsSetTPStat(t *testing.T) {
	tpStat = &utils.TPStatProfile{
		Tenant:    "cgrates.org",
		TPid:      "TPS1",
		ID:        "Stat1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		TTL: "1",
		Metrics: []*utils.MetricWithFilters{
			&utils.MetricWithFilters{
				MetricID: "*sum",
			},
		},
		Blocker:      false,
		Stored:       false,
		Weight:       20,
		MinItems:     1,
		ThresholdIDs: []string{"ThreshValue", "ThreshValueTwo"},
	}
	sort.Strings(tpStat.ThresholdIDs)
	var result string
	if err := tpStatRPC.Call(utils.APIerSv1SetTPStat, tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPStatsGetTPStatAfterSet(t *testing.T) {
	var respond *utils.TPStatProfile
	if err := tpStatRPC.Call(utils.APIerSv1GetTPStat,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Stat1"}, &respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpStat, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat, respond)
	}
}

func testTPStatsUpdateTPStat(t *testing.T) {
	var result string
	tpStat.Weight = 21
	tpStat.Metrics = []*utils.MetricWithFilters{
		&utils.MetricWithFilters{
			MetricID: "*sum",
		},
		&utils.MetricWithFilters{
			MetricID: "*averege",
		},
	}
	sort.Slice(tpStat.Metrics, func(i, j int) bool {
		return strings.Compare(tpStat.Metrics[i].MetricID, tpStat.Metrics[j].MetricID) == -1
	})
	if err := tpStatRPC.Call(utils.APIerSv1SetTPStat, tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPStatsGetTPStatAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPStatProfile
	if err := tpStatRPC.Call(utils.APIerSv1GetTPStat,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Stat1"}, &expectedTPS); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPS.ThresholdIDs)
	sort.Slice(expectedTPS.Metrics, func(i, j int) bool {
		return strings.Compare(expectedTPS.Metrics[i].MetricID, expectedTPS.Metrics[j].MetricID) == -1
	})
	if !reflect.DeepEqual(tpStat, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat, expectedTPS)
	}
}

func testTPStatsRemoveTPStat(t *testing.T) {
	var resp string
	if err := tpStatRPC.Call(utils.APIerSv1RemoveTPStat,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Stat1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPStatsGetTPStatAfterRemove(t *testing.T) {
	var respond *utils.TPStatProfile
	if err := tpStatRPC.Call(utils.APIerSv1GetTPStat,
		&utils.TPTntID{TPid: "TPS1", Tenant: "cgrates.org", ID: "Stat1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPStatsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpStatDelay); err != nil {
		t.Error(err)
	}
}
