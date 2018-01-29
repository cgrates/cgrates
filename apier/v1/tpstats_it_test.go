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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
)

var (
	tpStatCfgPath   string
	tpStatCfg       *config.CGRConfig
	tpStatRPC       *rpc.Client
	tpStatDataDir   = "/usr/share/cgrates"
	tpStat          *utils.TPStats
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
	testTPStatsRemTPStat,
	testTPStatsGetTPStatAfterRemove,
	testTPStatsKillEngine,
}

//Test start here
func TestTPStatITMySql(t *testing.T) {
	tpStatConfigDIR = "tutmysql"
	for _, stest := range sTestsTPStats {
		t.Run(tpStatConfigDIR, stest)
	}
}

func TestTPStatITMongo(t *testing.T) {
	tpStatConfigDIR = "tutmongo"
	for _, stest := range sTestsTPStats {
		t.Run(tpStatConfigDIR, stest)
	}
}

func TestTPStatITPG(t *testing.T) {
	tpStatConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPStats {
		t.Run(tpStatConfigDIR, stest)
	}
}

func testTPStatsInitCfg(t *testing.T) {
	var err error
	tpStatCfgPath = path.Join(tpStatDataDir, "conf", "samples", tpStatConfigDIR)
	tpStatCfg, err = config.NewCGRConfigFromFolder(tpStatCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpStatCfg.DataFolderPath = tpStatDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpStatCfg)
	switch tpStatConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db
		tpStatDelay = 2000
	default:
		tpStatDelay = 1000
	}
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
	tpStatRPC, err = jsonrpc.Dial("tcp", tpStatCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPStatsGetTPStatBeforeSet(t *testing.T) {
	var reply *utils.TPStats
	if err := tpStatRPC.Call("ApierV1.GetTPStat",
		&AttrGetTPStat{TPid: "TPS1", ID: "Stat1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPStatsSetTPStat(t *testing.T) {
	tpStat = &utils.TPStats{
		Tenant:    "cgrates.org",
		TPid:      "TPS1",
		ID:        "Stat1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		TTL: "1",
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "*sum",
				Parameters: "Param1",
			},
		},
		Blocker:      false,
		Stored:       false,
		Weight:       20,
		MinItems:     1,
		ThresholdIDs: []string{"ThreshValue", "ThreshValueTwo"},
	}
	var result string
	if err := tpStatRPC.Call("ApierV1.SetTPStat", tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPStatsGetTPStatAfterSet(t *testing.T) {
	var respond *utils.TPStats
	if err := tpStatRPC.Call("ApierV1.GetTPStat",
		&AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpStat, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat, respond)
	}
}

func testTPStatsUpdateTPStat(t *testing.T) {
	var result string
	tpStat.Weight = 21
	tpStat.Metrics = []*utils.MetricWithParams{
		&utils.MetricWithParams{
			MetricID:   "*sum",
			Parameters: "Param1",
		},
		&utils.MetricWithParams{
			MetricID:   "*averege",
			Parameters: "Param1",
		},
	}
	if err := tpStatRPC.Call("ApierV1.SetTPStat", tpStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPStatsGetTPStatAfterUpdate(t *testing.T) {
	var expectedTPS *utils.TPStats
	if err := tpStatRPC.Call("ApierV1.GetTPStat",
		&AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &expectedTPS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpStat, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpStat, expectedTPS)
	}
}

func testTPStatsRemTPStat(t *testing.T) {
	var resp string
	if err := tpStatRPC.Call("ApierV1.RemTPStat",
		&AttrGetTPStat{TPid: tpStat.TPid, ID: tpStat.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPStatsGetTPStatAfterRemove(t *testing.T) {
	var respond *utils.TPStats
	if err := tpStatRPC.Call("ApierV1.GetTPStat",
		&AttrGetTPStat{TPid: "TPS1", ID: "Stat1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPStatsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpStatDelay); err != nil {
		t.Error(err)
	}
}
