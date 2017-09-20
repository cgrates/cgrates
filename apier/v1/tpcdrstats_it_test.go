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
	tpCdrStatCfgPath   string
	tpCdrStatCfg       *config.CGRConfig
	tpCdrStatRPC       *rpc.Client
	tpCdrStatDataDir   = "/usr/share/cgrates"
	tpCdrStats         *utils.TPCdrStats
	tpCdrStatDelay     int
	tpCdrStatConfigDIR string //run tests for specific configuration

)

var sTestsTPCdrStats = []func(t *testing.T){
	testTPCdrStatsInitCfg,
	testTPCdrStatsResetStorDb,
	testTPCdrStatsStartEngine,
	testTPCdrStatsRpcConn,
	testTPCdrStatsGetTPCdrStatsBeforeSet,
	testTPCdrStatsSetTPCdrStats,
	testTPCdrStatsGetTPCdrStatsAfterSet,
	testTPCdrStatsGetTPCdrStatsIds,
	testTPCdrStatsUpdateTPCdrStats,
	testTPCdrStatsGetTPCdrStatsAfterUpdate,
	testTPCdrStatsRemTPCdrStats,
	testTPCdrStatsGetTPCdrStatsAfterRemove,
	testTPCdrStatsKillEngine,
}

//Test start here
func TestTPCdrStatsITMySql(t *testing.T) {
	tpCdrStatConfigDIR = "tutmysql"
	for _, stest := range sTestsTPCdrStats {
		t.Run(tpCdrStatConfigDIR, stest)
	}
}

func TestTPCdrStatsITMongo(t *testing.T) {
	tpCdrStatConfigDIR = "tutmongo"
	for _, stest := range sTestsTPCdrStats {
		t.Run(tpCdrStatConfigDIR, stest)
	}
}

func TestTPCdrStatsITPG(t *testing.T) {
	tpCdrStatConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPCdrStats {
		t.Run(tpCdrStatConfigDIR, stest)
	}
}

func testTPCdrStatsInitCfg(t *testing.T) {
	var err error
	tpCdrStatCfgPath = path.Join(tpCdrStatDataDir, "conf", "samples", tpCdrStatConfigDIR)
	tpCdrStatCfg, err = config.NewCGRConfigFromFolder(tpCdrStatCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpCdrStatCfg.DataFolderPath = tpCdrStatDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpCdrStatCfg)
	switch tpCdrStatConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpCdrStatDelay = 2000
	default:
		tpCdrStatDelay = 1000
	}
}

// Wipe out the cdr database
func testTPCdrStatsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpCdrStatCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPCdrStatsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpCdrStatCfgPath, tpCdrStatDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPCdrStatsRpcConn(t *testing.T) {
	var err error
	tpCdrStatRPC, err = jsonrpc.Dial("tcp", tpCdrStatCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPCdrStatsGetTPCdrStatsBeforeSet(t *testing.T) {
	var reply *utils.TPCdrStats
	if err := tpCdrStatRPC.Call("ApierV1.GetTPCdrStats", &AttrGetTPCdrStats{TPid: "TPCdr", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPCdrStatsSetTPCdrStats(t *testing.T) {
	tpCdrStats = &utils.TPCdrStats{
		TPid: "TPCdr",
		ID:   "ID",
		CdrStats: []*utils.TPCdrStat{
			&utils.TPCdrStat{
				QueueLength:      "10",
				TimeWindow:       "0",
				SaveInterval:     "10s",
				Metrics:          "ASR",
				SetupInterval:    "",
				TORs:             "",
				CdrHosts:         "",
				CdrSources:       "",
				ReqTypes:         "",
				Directions:       "",
				Tenants:          "cgrates.org",
				Categories:       "",
				Accounts:         "",
				Subjects:         "1001",
				DestinationIds:   "1003",
				PddInterval:      "",
				UsageInterval:    "",
				Suppliers:        "suppl1",
				DisconnectCauses: "",
				MediationRunIds:  "*default",
				RatedAccounts:    "",
				RatedSubjects:    "",
				CostInterval:     "",
				ActionTriggers:   "CDRST1_WARN",
			},
			&utils.TPCdrStat{
				QueueLength:      "10",
				TimeWindow:       "0",
				SaveInterval:     "10s",
				Metrics:          "ACC",
				SetupInterval:    "",
				TORs:             "",
				CdrHosts:         "",
				CdrSources:       "",
				ReqTypes:         "",
				Directions:       "",
				Tenants:          "cgrates.org",
				Categories:       "",
				Accounts:         "",
				Subjects:         "1002",
				DestinationIds:   "1003",
				PddInterval:      "",
				UsageInterval:    "",
				Suppliers:        "suppl1",
				DisconnectCauses: "",
				MediationRunIds:  "*default",
				RatedAccounts:    "",
				RatedSubjects:    "",
				CostInterval:     "",
				ActionTriggers:   "CDRST1_WARN",
			},
		},
	}

	var result string
	if err := tpCdrStatRPC.Call("ApierV1.SetTPCdrStats", tpCdrStats, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPCdrStatsGetTPCdrStatsAfterSet(t *testing.T) {
	var reply *utils.TPCdrStats
	if err := tpCdrStatRPC.Call("ApierV1.GetTPCdrStats", &AttrGetTPCdrStats{TPid: "TPCdr", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpCdrStats.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpCdrStats.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpCdrStats.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpCdrStats.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpCdrStats.CdrStats), len(reply.CdrStats)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpCdrStats.CdrStats), len(reply.CdrStats))
	}
}

func testTPCdrStatsGetTPCdrStatsIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"ID"}
	if err := tpCdrStatRPC.Call("ApierV1.GetTPCdrStatsIds", &AttrGetTPCdrStatIds{TPid: "TPCdr"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPCdrStatsUpdateTPCdrStats(t *testing.T) {
	tpCdrStats.CdrStats = []*utils.TPCdrStat{
		&utils.TPCdrStat{
			QueueLength:      "10",
			TimeWindow:       "0",
			SaveInterval:     "10s",
			Metrics:          "ASR",
			SetupInterval:    "",
			TORs:             "",
			CdrHosts:         "",
			CdrSources:       "",
			ReqTypes:         "",
			Directions:       "",
			Tenants:          "cgrates.org",
			Categories:       "",
			Accounts:         "",
			Subjects:         "1001",
			DestinationIds:   "1003",
			PddInterval:      "",
			UsageInterval:    "",
			Suppliers:        "suppl1",
			DisconnectCauses: "",
			MediationRunIds:  "*default",
			RatedAccounts:    "",
			RatedSubjects:    "",
			CostInterval:     "",
			ActionTriggers:   "CDRST1_WARN",
		},
		&utils.TPCdrStat{
			QueueLength:      "10",
			TimeWindow:       "0",
			SaveInterval:     "10s",
			Metrics:          "ACC",
			SetupInterval:    "",
			TORs:             "",
			CdrHosts:         "",
			CdrSources:       "",
			ReqTypes:         "",
			Directions:       "",
			Tenants:          "cgrates.org",
			Categories:       "",
			Accounts:         "",
			Subjects:         "1002",
			DestinationIds:   "1003",
			PddInterval:      "",
			UsageInterval:    "",
			Suppliers:        "suppl1",
			DisconnectCauses: "",
			MediationRunIds:  "*default",
			RatedAccounts:    "",
			RatedSubjects:    "",
			CostInterval:     "",
			ActionTriggers:   "CDRST1_WARN",
		},
		&utils.TPCdrStat{
			QueueLength:      "10",
			TimeWindow:       "0",
			SaveInterval:     "10s",
			Metrics:          "ACC",
			SetupInterval:    "",
			TORs:             "",
			CdrHosts:         "",
			CdrSources:       "",
			ReqTypes:         "",
			Directions:       "",
			Tenants:          "cgrates.org",
			Categories:       "",
			Accounts:         "",
			Subjects:         "1002",
			DestinationIds:   "1003",
			PddInterval:      "",
			UsageInterval:    "",
			Suppliers:        "suppl1",
			DisconnectCauses: "",
			MediationRunIds:  "*default",
			RatedAccounts:    "",
			RatedSubjects:    "",
			CostInterval:     "",
			ActionTriggers:   "CDRST1001_WARN",
		},
	}
	var result string
	if err := tpCdrStatRPC.Call("ApierV1.SetTPCdrStats", tpCdrStats, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

}

func testTPCdrStatsGetTPCdrStatsAfterUpdate(t *testing.T) {
	var reply *utils.TPCdrStats
	if err := tpCdrStatRPC.Call("ApierV1.GetTPCdrStats", &AttrGetTPCdrStats{TPid: "TPCdr", ID: "ID"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpCdrStats.TPid, reply.TPid) {
		t.Errorf("Expecting : %+v, received: %+v", tpCdrStats.TPid, reply.TPid)
	} else if !reflect.DeepEqual(tpCdrStats.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", tpCdrStats.ID, reply.ID)
	} else if !reflect.DeepEqual(len(tpCdrStats.CdrStats), len(reply.CdrStats)) {
		t.Errorf("Expecting : %+v, received: %+v", len(tpCdrStats.CdrStats), len(reply.CdrStats))
	}

}

func testTPCdrStatsRemTPCdrStats(t *testing.T) {
	var resp string
	if err := tpCdrStatRPC.Call("ApierV1.RemTPCdrStats", &AttrGetTPCdrStats{TPid: "TPCdr", ID: "ID"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPCdrStatsGetTPCdrStatsAfterRemove(t *testing.T) {
	var reply *utils.TPCdrStats
	if err := tpCdrStatRPC.Call("ApierV1.GetTPCdrStats", &AttrGetTPCdrStats{TPid: "TPCdr", ID: "ID"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPCdrStatsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpCdrStatDelay); err != nil {
		t.Error(err)
	}
}
