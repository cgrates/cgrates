//go:build integration
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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accCfgPath string
	accCfg     *config.CGRConfig
	accRpc     *birpc.Client
	accConfDIR string //run tests for specific configuration
	account    *engine.Account
	accDelay   int

	sTestsAcc = []func(t *testing.T){
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
)

// Test start here
func TestAccIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		accConfDIR = "tutinternal"
	case utils.MetaMySQL:
		accConfDIR = "tutmysql"
	case utils.MetaMongo:
		accConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsAcc {
		t.Run(accConfDIR, stest)
	}
}

func testV1AccLoadConfig(t *testing.T) {
	var err error
	accCfgPath = path.Join(*utils.DataDir, "conf", "samples", accConfDIR)
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
	accRpc = engine.NewRPCClient(t, accCfg.ListenCfg())
}

func testV1AccGetAccountBeforeSet(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := accRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(200 * time.Millisecond)
}

func testV1AccGetAccountAfterLoad(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"},
		&reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccount(t *testing.T) {
	var reply string
	if err := accRpc.Call(context.Background(), utils.APIerSv1RemoveAccount,
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "1001"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterDelete(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccSetAccount(t *testing.T) {
	var reply string
	if err := accRpc.Call(context.Background(), utils.APIerSv2SetAccount,
		&utils.AttrSetAccount{Tenant: "cgrates.org", Account: "testacc"}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterSet(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testacc"}, &reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccountSet(t *testing.T) {
	var reply string
	if err := accRpc.Call(context.Background(), utils.APIerSv1RemoveAccount,
		&utils.AttrRemoveAccount{Tenant: "cgrates.org", Account: "testacc"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountSetAfterDelete(t *testing.T) {
	var reply *engine.Account
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "testacc"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

/*
Need to investigate for redis why didn't return not found
func testV1AccRemAccountAfterDelete(t *testing.T) {
	var reply string
	if err := accRpc.Call(context.Background(),utils.APIerSv1RemoveAccount,
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
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1002"},
		&reply); err != nil {
		t.Error(err)
	} else if _, has := reply.BalanceMap[utils.MetaData]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(reply.BalanceMap[utils.MetaData]))
	} else if len(reply.BalanceMap[utils.MetaData]) != 1 {
		t.Error("Unexpected number of balances returned: ", len(reply.BalanceMap[utils.MetaData]))
	} else if reply.BalanceMap[utils.MetaData][0].ExpirationDate.After(timeAfter) &&
		reply.BalanceMap[utils.MetaData][0].ExpirationDate.Before(timeBefore) {
		t.Error("Unexpected expiration date returned: ", reply.BalanceMap[utils.MetaData][0].ExpirationDate)
	}

}

// Add test to check if AccountS send event to ThresholdS
func testV1AccSendToThreshold(t *testing.T) {
	var reply string

	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "DISABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaDisableAccount},
		{Identifier: utils.MetaLog},
	}}
	if err := accRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_AccDisableAndLog",
			FilterIDs: []string{"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:testAccThreshold"},
			MaxHits:   -1,
			MinSleep:  time.Second,
			Weight:    20.0,
			Async:     true,
			ActionIDs: []string{"DISABLE_LOG"},
		},
	}

	if err := accRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// Add an account
	attrs := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testAccThreshold",
		BalanceType: "*monetary",
		Value:       1.5,
		Balance: map[string]any{
			utils.ID: "testAccSetBalance",
		},
	}
	if err := accRpc.Call(context.Background(), utils.APIerSv1SetBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}

	// give time to threshold to made the change
	time.Sleep(100 * time.Millisecond)
	//verify the account
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testAccThreshold",
	}
	if err := accRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
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
