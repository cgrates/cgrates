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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpRateCfgPath   string
	tpRateCfg       *config.CGRConfig
	tpRateRPC       *rpc.Client
	tpRate          *utils.TPRateRALs
	tpRateDelay     int
	tpRateConfigDIR string //run tests for specific configuration
)

var sTestsTPRates = []func(t *testing.T){
	testTPRatesInitCfg,
	testTPRatesResetStorDb,
	testTPRatesStartEngine,
	testTPRatesRpcConn,
	testTPRatesGetTPRateforeSet,
	testTPRatesSetTPRate,
	testTPRatesGetTPRateAfterSet,
	testTPRatesGetTPRateIds,
	testTPRatesUpdateTPRate,
	testTPRatesGetTPRateAfterUpdate,
	testTPRatesRemoveTPRate,
	testTPRatesGetTPRateAfterRemove,
	testTPRatesKillEngine,
}

//Test start here
func TestTPRatesIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpRateConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpRateConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpRateConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpRateConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPRates {
		t.Run(tpRateConfigDIR, stest)
	}
}

func testTPRatesInitCfg(t *testing.T) {
	var err error
	tpRateCfgPath = path.Join(*dataDir, "conf", "samples", tpRateConfigDIR)
	tpRateCfg, err = config.NewCGRConfigFromPath(tpRateCfgPath)
	if err != nil {
		t.Error(err)
	}
	switch tpRateConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpRateDelay = 2000
	default:
		tpRateDelay = 1000
	}
}

// Wipe out the cdr database
func testTPRatesResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpRateCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPRatesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpRateCfgPath, tpRateDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPRatesRpcConn(t *testing.T) {
	var err error
	tpRateRPC, err = jsonrpc.Dial(utils.TCP, tpRateCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatesGetTPRateforeSet(t *testing.T) {
	var reply *utils.TPRateRALs
	if err := tpRateRPC.Call(utils.APIerSv1GetTPRate,
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatesSetTPRate(t *testing.T) {
	tpRate = &utils.TPRateRALs{
		TPid: "TPidTpRate",
		ID:   "RT_FS_USERS",
		RateSlots: []*utils.RateSlot{
			&utils.RateSlot{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "6s",
				RateIncrement:      "6s",
				GroupIntervalStart: "0s",
			},
			&utils.RateSlot{
				ConnectFee:         12,
				Rate:               3,
				RateUnit:           "4s",
				RateIncrement:      "6s",
				GroupIntervalStart: "1s",
			},
		},
	}
	var result string
	if err := tpRateRPC.Call(utils.APIerSv1SetTPRate, tpRate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatesGetTPRateAfterSet(t *testing.T) {
	var reply *utils.TPRateRALs
	if err := tpRateRPC.Call(utils.APIerSv1GetTPRate, &AttrGetTPRate{TPid: "TPidTpRate", ID: tpRate.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpRate, reply)
	}
}

func testTPRatesGetTPRateIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"RT_FS_USERS"}
	if err := tpRateRPC.Call(utils.APIerSv1GetTPRateIds, &AttrGetTPRateIds{TPid: "TPidTpRate"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPRatesUpdateTPRate(t *testing.T) {
	var result string
	tpRate.RateSlots = []*utils.RateSlot{
		&utils.RateSlot{
			ConnectFee:         12,
			Rate:               3,
			RateUnit:           "6s",
			RateIncrement:      "6s",
			GroupIntervalStart: "0s",
		},
		&utils.RateSlot{
			ConnectFee:         12,
			Rate:               10,
			RateUnit:           "4s",
			RateIncrement:      "6s",
			GroupIntervalStart: "1s",
		},
		&utils.RateSlot{
			ConnectFee:         5,
			Rate:               10,
			RateUnit:           "4s",
			RateIncrement:      "6s",
			GroupIntervalStart: "3s",
		},
	}
	if err := tpRateRPC.Call(utils.APIerSv1SetTPRate, tpRate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatesGetTPRateAfterUpdate(t *testing.T) {
	var reply *utils.TPRateRALs
	if err := tpRateRPC.Call(utils.APIerSv1GetTPRate,
		&AttrGetTPRate{TPid: "TPidTpRate", ID: tpRate.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpRate, reply)
	}

}

func testTPRatesRemoveTPRate(t *testing.T) {
	var resp string
	if err := tpRateRPC.Call(utils.APIerSv1RemoveTPRate,
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatesGetTPRateAfterRemove(t *testing.T) {
	var reply *utils.TPRateRALs
	if err := tpRateRPC.Call(utils.APIerSv1GetTPRate,
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatesKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRateDelay); err != nil {
		t.Error(err)
	}
}
