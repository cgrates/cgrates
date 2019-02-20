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

package general_tests

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrCfgPath string
	fltrCfg     *config.CGRConfig
	fltrRpc     *rpc.Client
	fltrConfDIR string //run tests for specific configuration
	fltrDelay   int
)

var sTestsFltr = []func(t *testing.T){
	testV1FltrLoadConfig,
	testV1FltrInitDataDb,
	testV1FltrResetStorDb,
	testV1FltrStartEngine,
	testV1FltrRpcConn,

	testV1FltrStopEngine,
}

// Test start here
func TestFltrIT(t *testing.T) {
	fltrConfDIR = "filters"
	for _, stest := range sTestsFltr {
		t.Run(fltrConfDIR, stest)
	}
}

func testV1FltrLoadConfig(t *testing.T) {
	var err error
	fltrCfgPath = path.Join(*dataDir, "conf", "samples", fltrConfDIR)
	if fltrCfg, err = config.NewCGRConfigFromFolder(fltrCfgPath); err != nil {
		t.Error(err)
	}
	fltrDelay = 1000
}

func testV1FltrInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrCfgPath, fltrDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrRpcConn(t *testing.T) {
	var err error
	fltrRpc, err = jsonrpc.Dial("tcp", fltrCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FltrLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := fltrRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1FltrStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
