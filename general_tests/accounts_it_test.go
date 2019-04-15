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

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accCfgPath string
	accCfg     *config.CGRConfig
	accRpc     *rpc.Client
	accConfDIR string //run tests for specific configuration
	account    *engine.Account
	accDelay   int
)

var sTestsAcc = []func(t *testing.T){
	testV1AccLoadConfig,
	testV1AccInitDataDb,
	testAccResetStorDb,
	testV1AccStartEngine,
	testV1AccRpcConn,
	testV1AccGetAccountBeforeSet,
	testV1AccLoadTarrifPlans,
	testV1AccGetAccountAfterLoad,
	testV1AccRemAccount,
	testV1AccGetAccountAfterDelete,
	testV1AccSetAccount,
	testV1AccGetAccountAfterSet,
	testV1AccRemAccountSet,
	testV1AccGetAccountSetAfterDelete,
	//testV1AccRemAccountAfterDelete,
	testV1AccMonthly,
	testV1AccSendToThreshold,
	testV1AccStopEngine,
}

// Test start here
func TestAccITMySQL(t *testing.T) {
	accConfDIR = "tutmysql"
	for _, stest := range sTestsAcc {
		t.Run(accConfDIR, stest)
	}
}

func TestAccITMongo(t *testing.T) {
	accConfDIR = "tutmongo"
	for _, stest := range sTestsAcc {
		t.Run(accConfDIR, stest)
	}
}

func testV1AccLoadConfig(t *testing.T) {
	var err error
	accCfgPath = path.Join(*dataDir, "conf", "samples", accConfDIR)
	if accCfg, err = config.NewCGRConfigFromPath(accCfgPath); err != nil {
		t.Error(err)
	}
	accDelay = 1000
}

func testV1AccInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1AccStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accCfgPath, accDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1AccRpcConn(t *testing.T) {
	var err error
	accRpc, err = jsonrpc.Dial("tcp", accCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1AccGetAccountBeforeSet(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := accRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1AccGetAccountAfterLoad(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(utils.ApierV2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccount(t *testing.T) {
	var reply string
	if err := accRpc.Call("ApierV1.RemoveAccount",
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "1001"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterDelete(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccSetAccount(t *testing.T) {
	var reply string
	if err := accRpc.Call("ApierV2.SetAccount",
		&utils.AttrSetAccount{Tenant: "cgrates.org", Account: "testacc"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterSet(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testacc"}, &reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccountSet(t *testing.T) {
	var reply string
	if err := accRpc.Call("ApierV1.RemoveAccount",
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "testacc"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountSetAfterDelete(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testacc"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

/*
Need to investigate for redis why didn't return not found
func testV1AccRemAccountAfterDelete(t *testing.T) {
	var reply string
	if err := accRpc.Call("ApierV1.RemoveAccount",
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "testacc"},
		&reply); err == nil || err.Error() != utils.NewErrServerError(utils.ErrNotFound).Error() {
		t.Error(err)
	}
}
*/

func testV1AccMonthly(t *testing.T) {
	// add 10 seconds delay before and after
	timeAfter := time.Now().Add(10*time.Second).AddDate(0, 1, 0)
	timeBefore := time.Now().Add(-10*time.Second).AddDate(0, 1, 0)
	var reply *engine.Account
	if err := accRpc.Call("ApierV2.GetAccount",
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"},
		&reply); err != nil {
		t.Error(err)
	} else if _, has := reply.BalanceMap[utils.DATA]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(reply.BalanceMap[utils.DATA]))
	} else if len(reply.BalanceMap[utils.DATA]) != 1 {
		t.Error("Unexpected number of balances returned: ", len(reply.BalanceMap[utils.DATA]))
	} else if reply.BalanceMap[utils.DATA][0].ExpirationDate.After(timeAfter) &&
		reply.BalanceMap[utils.DATA][0].ExpirationDate.Before(timeBefore) {
		t.Error("Unexpected expiration date returned: ", reply.BalanceMap[utils.DATA][0].ExpirationDate)
	}

}

//Add test to check if AccountS send event to ThresholdS
func testV1AccSendToThreshold(t *testing.T) {
	var reply string

	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "DISABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: engine.DISABLE_ACCOUNT},
		{Identifier: engine.LOG},
	}}
	if err := accRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	time.Sleep(10 * time.Millisecond)

	tPrfl := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccDisableAndLog",
		FilterIDs: []string{"*string:~Account:testAccThreshold"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    20.0,
		Async:     true,
		ActionIDs: []string{"DISABLE_LOG"},
	}

	if err := accRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// Add an account
	attrs := &v1.AttrAddBalance{Tenant: "cgrates.org", Account: "testAccThreshold",
		BalanceId:   utils.StringPointer("testAccSetBalance"),
		BalanceType: "*monetary", Value: 1.5}
	if err := accRpc.Call(utils.ApierV1SetBalance, attrs, &reply); err != nil {
		t.Error("Got error on ApierV1.SetBalance: ", err.Error())
	} else if reply != "OK" {
		t.Errorf("Calling ApierV1.SetBalance received: %s", reply)
	}

	// give time to threshold to made the change
	time.Sleep(10 * time.Millisecond)
	//verify the account
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testAccThreshold",
	}
	if err := accRpc.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}
}

func testV1AccStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
