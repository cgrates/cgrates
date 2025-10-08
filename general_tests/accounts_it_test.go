//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	accCfgPath string
	accCfg     *config.CGRConfig
	accRpc     *birpc.Client
	accConfDIR string //run tests for specific configuration
	accDelay   int

	sTestsAcc = []func(t *testing.T){
		testV1AccLoadConfig,
		testV1AccResetDBs,
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
		//testV1AccMonthly,
		//testV1AccSendToThreshold,
		testV1AccStopEngine,
	}
)

// Test start here
func TestAccIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		accConfDIR = "acc_generaltest_internal"
	case utils.MetaMySQL:
		accConfDIR = "acc_generaltest_mysql"
	case utils.MetaMongo:
		accConfDIR = "acc_generaltest_mongo"
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
	if accCfg, err = config.NewCGRConfigFromPath(context.Background(), accCfgPath); err != nil {
		t.Error(err)
	}
	accDelay = 1000
}

func testV1AccResetDBs(t *testing.T) {
	if err := engine.InitDataDB(accCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(accCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1AccStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accCfgPath, accDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1AccRpcConn(t *testing.T) {
	accRpc = engine.NewRPCClient(t, accCfg.ListenCfg(), *utils.Encoding)
}

func testV1AccGetAccountBeforeSet(t *testing.T) {
	var reply *utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1001",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccLoadTarrifPlans(t *testing.T) {
	caching := utils.MetaReload
	if accCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := accRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaStopOnError: true,
				utils.MetaCache:       caching,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
	time.Sleep(200 * time.Millisecond)
}

func testV1AccGetAccountAfterLoad(t *testing.T) {
	var reply *utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1001",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args,
		&reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccount(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1001",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1RemoveAccount,
		args,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterDelete(t *testing.T) {
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1001",
		},
	}
	var reply *utils.Account
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccSetAccount(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "testacc",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1SetAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountAfterSet(t *testing.T) {
	var reply *utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "testacc",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args, &reply); err != nil {
		t.Error(err)
	}
}

func testV1AccRemAccountSet(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "testacc",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1RemoveAccount,
		args,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1AccGetAccountSetAfterDelete(t *testing.T) {
	var reply *utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "testacc",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1AccMonthly(t *testing.T) {
	// add 10 seconds delay before and after
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "1002",
		},
	}
	var reply *utils.Account
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount,
		args,
		&reply); err != nil {
		t.Error(err)
	} else if _, has := reply.Balances[utils.MetaData]; !has {
		t.Error("Unexpected balance returned: ", utils.ToJSON(reply.Balances[utils.MetaData]))
	}

}

// Add test to check if AccountS send event to ThresholdS
func testV1AccSendToThreshold(t *testing.T) {
	var reply string

	// Add a disable and log action
	args := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			ID: "DISABLE_LOG",
			Actions: []*utils.APAction{
				{ID: utils.MetaDisableAccount},
				{ID: utils.MetaLog},
			},
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1SetActionProfile, args, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
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
			MaxHits:  -1,
			MinSleep: time.Second,
			Weights: utils.DynamicWeights{
				{
					Weight: 20.0,
				},
			},
			Async:            true,
			ActionProfileIDs: []string{"DISABLE_LOG"},
		},
	}

	if err := accRpc.Call(context.Background(), utils.AdminSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	// Add an account
	attrs := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "testAccThreshold",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.testAccSetBalance.*monetary",
				Value: "1.5",
			},
		},
	}
	if err := accRpc.Call(context.Background(), utils.AccountSv1ActionSetBalance, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", reply)
	}
	time.Sleep(10 * time.Millisecond)
	var acnt utils.Account
	attrAcc := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "testAccThreshold",
		},
	}
	if err := accRpc.Call(context.Background(), utils.AdminSv1GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	}
}

func testV1AccStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
