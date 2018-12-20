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
	tpTimingCfgPath   string
	tpTimingCfg       *config.CGRConfig
	tpTimingRPC       *rpc.Client
	tpTimingDataDir   = "/usr/share/cgrates"
	tpTiming          *utils.ApierTPTiming
	tpTimingDelay     int
	tpTimingConfigDIR string //run tests for specific configuration
)

var sTestsTPTiming = []func(t *testing.T){
	testTPTimingsInitCfg,
	testTPTimingsResetStorDb,
	testTPTimingsStartEngine,
	testTPTimingsRpcConn,
	testTPTimingsGetTPTimingBeforeSet,
	testTPTimingsSetTPTiming,
	testTPTimingsGetTPTimingAfterSet,
	testTPTimingsGetTPTimingIds,
	testTPTimingsUpdateTPTiming,
	testTPTimingsGetTPTimingAfterUpdate,
	testTPTimingsRemTPTiming,
	testTPTimingsGetTPTimingAfterRemove,
	testTPTimingsKillEngine,
}

//Test start here
func TestTPTimingITMySql(t *testing.T) {
	tpTimingConfigDIR = "tutmysql"
	for _, stest := range sTestsTPTiming {
		t.Run(tpTimingConfigDIR, stest)
	}
}

func TestTPTimingITMongo(t *testing.T) {
	tpTimingConfigDIR = "tutmongo"
	for _, stest := range sTestsTPTiming {
		t.Run(tpTimingConfigDIR, stest)
	}
}

func TestTPTimingITPG(t *testing.T) {
	tpTimingConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPTiming {
		t.Run(tpTimingConfigDIR, stest)
	}
}

func testTPTimingsInitCfg(t *testing.T) {
	var err error
	tpTimingCfgPath = path.Join(tpTimingDataDir, "conf", "samples", tpTimingConfigDIR)
	tpTimingCfg, err = config.NewCGRConfigFromFolder(tpTimingCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpTimingCfg.DataFolderPath = tpTimingDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpTimingCfg)
	switch tpTimingConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db
		tpTimingDelay = 2000
	default:
		tpTimingDelay = 1000
	}
}

// Wipe out the cdr database
func testTPTimingsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpTimingCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPTimingsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpTimingCfgPath, tpTimingDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPTimingsRpcConn(t *testing.T) {
	var err error
	tpTimingRPC, err = jsonrpc.Dial("tcp", tpTimingCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPTimingsGetTPTimingBeforeSet(t *testing.T) {
	var reply *utils.ApierTPTiming
	if err := tpTimingRPC.Call("ApierV1.GetTPTiming", AttrGetTPTiming{TPid: "TPT1", ID: "Timining"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPTimingsSetTPTiming(t *testing.T) {
	tpTiming = &utils.ApierTPTiming{
		TPid:      "TPT1",
		ID:        "Timing",
		Years:     "2017",
		Months:    "05",
		MonthDays: "01",
		WeekDays:  "1",
		Time:      "15:00:00Z",
	}
	var result string
	if err := tpTimingRPC.Call("ApierV1.SetTPTiming", tpTiming, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPTimingsGetTPTimingAfterSet(t *testing.T) {
	var respond *utils.ApierTPTiming
	if err := tpTimingRPC.Call("ApierV1.GetTPTiming", &AttrGetTPTiming{TPid: tpTiming.TPid, ID: tpTiming.ID}, &respond); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpTiming, respond) {
		t.Errorf("Expecting: %+v, received: %+v", tpTiming, respond)
	}
}

func testTPTimingsGetTPTimingIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"Timing"}
	if err := tpTimingRPC.Call("ApierV1.GetTPTimingIds", &AttrGetTPTimingIds{TPid: tpTiming.TPid}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedTPID) {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedTPID)
	}
}

func testTPTimingsUpdateTPTiming(t *testing.T) {
	var result string
	tpTiming.Years = "2015"
	if err := tpTimingRPC.Call("ApierV1.SetTPTiming", tpTiming, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPTimingsGetTPTimingAfterUpdate(t *testing.T) {
	var expectedTPS *utils.ApierTPTiming
	if err := tpTimingRPC.Call("ApierV1.GetTPTiming", &AttrGetTPTiming{TPid: tpTiming.TPid, ID: tpTiming.ID}, &expectedTPS); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpTiming, expectedTPS) {
		t.Errorf("Expecting: %+v, received: %+v", tpTiming, expectedTPS)
	}
}

func testTPTimingsRemTPTiming(t *testing.T) {
	var resp string
	if err := tpTimingRPC.Call("ApierV1.RemTPTiming", &AttrGetTPTiming{TPid: tpTiming.TPid, ID: tpTiming.ID}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPTimingsGetTPTimingAfterRemove(t *testing.T) {
	var reply *utils.ApierTPTiming
	if err := tpTimingRPC.Call("ApierV1.GetTPTiming", AttrGetTPTiming{TPid: "TPT1", ID: "Timining"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPTimingsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpTimingDelay); err != nil {
		t.Error(err)
	}
}
