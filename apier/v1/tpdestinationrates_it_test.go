// +build integration

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
	tpDstRateCfgPath   string
	tpDstRateCfg       *config.CGRConfig
	tpDstRateRPC       *rpc.Client
	tpDstRateDataDir   = "/usr/share/cgrates"
	tpDstRate          *utils.TPDestinationRate
	tpDstRateDelay     int
	tpDstRateConfigDIR string //run tests for specific configuration
)

var sTestsTPDstRates = []func(t *testing.T){
	testTPDstRateInitCfg,
	testTPDstRateResetStorDb,
	testTPDstRateStartEngine,
	testTPDstRateRpcConn,
	testTPDstRateGetTPDstRateBeforeSet,
	testTPDstRateSetTPDstRate,
	testTPDstRateGetTPDstRateAfterSet,
	testTPDstRateGetTPDstRateIds,
	testTPDstRateUpdateTPDstRate,
	testTPDstRateGetTPDstRateAfterUpdate,
	testTPDstRateRemTPDstRate,
	testTPDstRateGetTPDstRateAfterRemove,
	testTPDstRateKillEngine,
}

//Test start here
func TestTPDstRateITMySql(t *testing.T) {
	tpDstRateConfigDIR = "tutmysql"
	for _, stest := range sTestsTPDstRates {
		t.Run(tpDstRateConfigDIR, stest)
	}
}

func TestTPDstRateITMongo(t *testing.T) {
	tpDstRateConfigDIR = "tutmongo"
	for _, stest := range sTestsTPDstRates {
		t.Run(tpDstRateConfigDIR, stest)
	}
}

func TestTPDstRateITPG(t *testing.T) {
	tpDstRateConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPDstRates {
		t.Run(tpDstRateConfigDIR, stest)
	}
}

func testTPDstRateInitCfg(t *testing.T) {
	var err error
	tpDstRateCfgPath = path.Join(tpDstRateDataDir, "conf", "samples", tpDstRateConfigDIR)
	tpDstRateCfg, err = config.NewCGRConfigFromPath(tpDstRateCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpDstRateCfg.DataFolderPath = tpDstRateDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpDstRateCfg)
	tpDstRateDelay = 1000
}

// Wipe out the cdr database
func testTPDstRateResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpDstRateCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPDstRateStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpDstRateCfgPath, tpDstRateDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPDstRateRpcConn(t *testing.T) {
	var err error
	tpDstRateRPC, err = jsonrpc.Dial("tcp", tpDstRateCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPDstRateGetTPDstRateBeforeSet(t *testing.T) {
	var reply *utils.TPDestinationRate
	if err := tpDstRateRPC.Call("ApierV1.GetTPDestinationRate",
		&AttrGetTPDestinationRate{TPid: "TP1", ID: "DR_FREESWITCH_USERS"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testTPDstRateSetTPDstRate(t *testing.T) {
	tpDstRate = &utils.TPDestinationRate{
		TPid: "TP1",
		ID:   "DR_FREESWITCH_USERS",
		DestinationRates: []*utils.DestinationRate{
			&utils.DestinationRate{
				DestinationId:    "FS_USERS",
				RateId:           "RT_FS_USERS",
				RoundingMethod:   "*up",
				RoundingDecimals: 2},
		},
	}
	var result string
	if err := tpDstRateRPC.Call("ApierV1.SetTPDestinationRate", tpDstRate, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPDstRateGetTPDstRateAfterSet(t *testing.T) {
	var reply *utils.TPDestinationRate
	if err := tpDstRateRPC.Call("ApierV1.GetTPDestinationRate",
		&AttrGetTPDestinationRate{TPid: "TP1", ID: "DR_FREESWITCH_USERS"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDstRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpDstRate, reply)
	}
}

func testTPDstRateGetTPDstRateIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"DR_FREESWITCH_USERS"}
	if err := tpDstRateRPC.Call("ApierV1.GetTPDestinationRateIds",
		&AttrTPDestinationRateIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPDstRateUpdateTPDstRate(t *testing.T) {

}

func testTPDstRateGetTPDstRateAfterUpdate(t *testing.T) {
	var reply *utils.TPDestinationRate
	if err := tpDstRateRPC.Call("ApierV1.GetTPDestinationRate",
		&AttrGetTPDestinationRate{TPid: "TP1", ID: "DR_FREESWITCH_USERS"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpDstRate, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpDstRate, reply)
	}
}

func testTPDstRateRemTPDstRate(t *testing.T) {
	var resp string
	if err := tpDstRateRPC.Call("ApierV1.RemoveTPDestinationRate",
		&AttrGetTPDestinationRate{TPid: "TP1", ID: "DR_FREESWITCH_USERS"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

}

func testTPDstRateGetTPDstRateAfterRemove(t *testing.T) {
	var reply *utils.TPDestinationRate
	if err := tpDstRateRPC.Call("ApierV1.GetTPDestinationRate",
		&AttrGetTPDestinationRate{TPid: "TP1", ID: "DR_FREESWITCH_USERS"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPDstRateKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpDstRateDelay); err != nil {
		t.Error(err)
	}
}
