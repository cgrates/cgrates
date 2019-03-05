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
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	rpcCfgPath string
	rpcCfg     *config.CGRConfig
	rpcRpc     *rpc.Client
	rpcConfDIR string //run tests for specific configuration
)

var sTestsRPCMethods = []func(t *testing.T){
	testRPCMethodsLoadConfig,
	testRPCMethodsInitDataDb,
	testRPCMethodsResetStorDb,
	testRPCMethodsStartEngine,
	testRPCMethodsRpcConn,
	testRPCMethodsFromFolder,
	testRPCMethodsAuthorizeSession,

	testRPCMethodsStopEngine,
}

// Test start here
func TestRPCMethods(t *testing.T) {
	rpcConfDIR = "rpc_methods"
	for _, stest := range sTestsRPCMethods {
		t.Run(rpcConfDIR, stest)
	}
}

func testRPCMethodsLoadConfig(t *testing.T) {
	var err error
	rpcCfgPath = path.Join(*dataDir, "conf", "samples", rpcConfDIR)
	if rpcCfg, err = config.NewCGRConfigFromFolder(rpcCfgPath); err != nil {
		t.Error(err)
	}
}

func testRPCMethodsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rpcCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRPCMethodsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rpcCfg); err != nil {
		t.Fatal(err)
	}
}

func testRPCMethodsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rpcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRPCMethodsRpcConn(t *testing.T) {
	var err error
	rpcRpc, err = jsonrpc.Dial("tcp", rpcCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testRPCMethodsFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := rpcRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRPCMethodsAuthorizeSession(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsAuthorizeSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsAuthorizeSession",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	//authorize the session
	var rplyFirst sessions.V1AuthorizeReply
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
		t.Fatal(err)
	}
	if *rplyFirst.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//delete the account
	var reply string
	if err := rpcRpc.Call("ApierV1.RemoveAccount",
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "1001"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify if the account was deleted
	var rcvAcc *engine.Account
	if err := rpcRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&rcvAcc); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//authorize again session (should take the response from cache)
	var rply sessions.V1AuthorizeReply
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(3 * time.Second)

	//authorize again session (this time we expect to receive an error)
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err == nil {
		t.Error("Unexpected error returned", err)
	}

}

func testRPCMethodsStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
