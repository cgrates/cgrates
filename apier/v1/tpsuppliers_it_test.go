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
	tpLCRCfgPath   string
	tpLCRCfg       *config.CGRConfig
	tpLCRRPC       *rpc.Client
	tpLCRDataDir   = "/usr/share/cgrates"
	tpLCR          *utils.TPLCR
	tpLCRDelay     int
	tpLCRConfigDIR string //run tests for specific configuration
)

var sTestsTPLCRs = []func(t *testing.T){
	testTPLCRInitCfg,
	testTPLCRResetStorDb,
	testTPLCRStartEngine,
	testtpLCRRPCConn,
	testTPLCRGetTPLCRBeforeSet,
	testTPLCRSetTPLCR,
	testTPLCRGetTPLCRAfterSet,
	testTPLCRGetTPLCRIds,
	testTPLCRUpdateTPLCR,
	testTPLCRGetTPLCRAfterUpdate,
	testTPLCRRemTPLCR,
	testTPLCRGetTPLCRAfterRemove,
	testTPLCRKillEngine,
}

//Test start here
func TestTPLCRITMySql(t *testing.T) {
	tpLCRConfigDIR = "tutmysql"
	for _, stest := range sTestsTPLCRs {
		t.Run(tpLCRConfigDIR, stest)
	}
}

func TestTPLCRITMongo(t *testing.T) {
	tpLCRConfigDIR = "tutmongo"
	for _, stest := range sTestsTPLCRs {
		t.Run(tpLCRConfigDIR, stest)
	}
}

func TestTPLCRITPG(t *testing.T) {
	tpLCRConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPLCRs {
		t.Run(tpLCRConfigDIR, stest)
	}
}

func testTPLCRInitCfg(t *testing.T) {
	var err error
	tpLCRCfgPath = path.Join(tpLCRDataDir, "conf", "samples", tpLCRConfigDIR)
	tpLCRCfg, err = config.NewCGRConfigFromFolder(tpLCRCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpLCRCfg.DataFolderPath = tpLCRDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpLCRCfg)
	switch tpLCRConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpLCRDelay = 2000
	default:
		tpLCRDelay = 1000
	}
}

// Wipe out the cdr database
func testTPLCRResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpLCRCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPLCRStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpLCRCfgPath, tpLCRDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testtpLCRRPCConn(t *testing.T) {
	var err error
	tpLCRRPC, err = jsonrpc.Dial("tcp", tpLCRCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPLCRGetTPLCRBeforeSet(t *testing.T) {
	var reply *utils.TPLCR
	if err := tpLCRRPC.Call("ApierV1.GetTPLCR", &AttrGetTPLCR{TPid: "TP1", ID: "LCR_1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPLCRSetTPLCR(t *testing.T) {
	tpLCR = &utils.TPLCR{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "LCR_1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Strategy:       "*lowest_cost",
		StrategyParams: []string{},
		SupplierID:     "supplier1",
		RatingPlanIDs:  []string{"RPL_1"},
		StatIDs:        []string{},
		Weight:         20,
	}
	var result string
	if err := tpLCRRPC.Call("ApierV1.SetTPLCR", tpLCR, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPLCRGetTPLCRAfterSet(t *testing.T) {
	var reply *utils.TPLCR
	if err := tpLCRRPC.Call("ApierV1.GetTPLCR", &AttrGetTPLCR{TPid: "TP1", ID: "LCR_1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpLCR, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpLCR, reply)
	}

}

func testTPLCRGetTPLCRIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"LCR_1"}
	if err := tpLCRRPC.Call("ApierV1.GetTPLCRIds", &AttrGetTPLCRIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPLCRUpdateTPLCR(t *testing.T) {
	tpLCR.StatIDs = []string{"STS_1", "STS_2"}
	var result string
	if err := tpLCRRPC.Call("ApierV1.SetTPLCR", tpLCR, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPLCRGetTPLCRAfterUpdate(t *testing.T) {
	var reply *utils.TPLCR
	if err := tpLCRRPC.Call("ApierV1.GetTPLCR", &AttrGetTPLCR{TPid: "TP1", ID: "LCR_1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpLCR, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpLCR, reply)
	}
}

func testTPLCRRemTPLCR(t *testing.T) {
	var resp string
	if err := tpLCRRPC.Call("ApierV1.RemTPLCR", &AttrRemTPLCR{TPid: "TP1", Tenant: "cgrates.org", ID: "LCR_1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPLCRGetTPLCRAfterRemove(t *testing.T) {
	var reply *utils.TPLCR
	if err := tpLCRRPC.Call("ApierV1.GetTPLCR", &AttrGetTPLCR{TPid: "TP1", ID: "LCR_1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPLCRKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpLCRDelay); err != nil {
		t.Error(err)
	}
}
