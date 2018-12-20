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
	tpRateDataDir   = "/usr/share/cgrates"
	tpRate          *utils.TPRate
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
	testTPRatesRemTPRate,
	testTPRatesGetTPRateAfterRemove,
	testTPRatesKillEngine,
}

//Test start here
func TestTPRatesITMySql(t *testing.T) {
	tpRateConfigDIR = "tutmysql"
	for _, stest := range sTestsTPRates {
		t.Run(tpRateConfigDIR, stest)
	}
}

func TestTPRatesITMongo(t *testing.T) {
	tpRateConfigDIR = "tutmongo"
	for _, stest := range sTestsTPRates {
		t.Run(tpRateConfigDIR, stest)
	}
}

func TestTPRatesITPG(t *testing.T) {
	tpRateConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPRates {
		t.Run(tpRateConfigDIR, stest)
	}
}

func testTPRatesInitCfg(t *testing.T) {
	var err error
	tpRateCfgPath = path.Join(tpRateDataDir, "conf", "samples", tpRateConfigDIR)
	tpRateCfg, err = config.NewCGRConfigFromFolder(tpRateCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpRateCfg.DataFolderPath = tpRateDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpRateCfg)
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
	tpRateRPC, err = jsonrpc.Dial("tcp", tpRateCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPRatesGetTPRateforeSet(t *testing.T) {
	var reply *utils.TPRate
	if err := tpRateRPC.Call("ApierV1.GetTPRate",
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatesSetTPRate(t *testing.T) {
	tpRate = &utils.TPRate{
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
	if err := tpRateRPC.Call("ApierV1.SetTPRate", tpRate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatesGetTPRateAfterSet(t *testing.T) {
	var reply *utils.TPRate
	if err := tpRateRPC.Call("ApierV1.GetTPRate", &AttrGetTPRate{TPid: "TPidTpRate", ID: tpRate.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpRate, reply)
	}
}

func testTPRatesGetTPRateIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"RT_FS_USERS"}
	if err := tpRateRPC.Call("ApierV1.GetTPRateIds", &AttrGetTPRateIds{TPid: "TPidTpRate"}, &result); err != nil {
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
	if err := tpRateRPC.Call("ApierV1.SetTPRate", tpRate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPRatesGetTPRateAfterUpdate(t *testing.T) {
	var reply *utils.TPRate
	if err := tpRateRPC.Call("ApierV1.GetTPRate",
		&AttrGetTPRate{TPid: "TPidTpRate", ID: tpRate.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpRate, reply)
	}

}

func testTPRatesRemTPRate(t *testing.T) {
	var resp string
	if err := tpRateRPC.Call("ApierV1.RemTPRate",
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPRatesGetTPRateAfterRemove(t *testing.T) {
	var reply *utils.TPRate
	if err := tpRateRPC.Call("ApierV1.GetTPRate",
		&AttrGetTPRate{TPid: "TPidTpRate", ID: "RT_FS_USERS"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPRatesKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpRateDelay); err != nil {
		t.Error(err)
	}
}
