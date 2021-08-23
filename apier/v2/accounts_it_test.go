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

package v2

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accConfigDIR string //run tests for specific configuration
	accCfgPath   string
	accCfg       *config.CGRConfig
	accRPC       *rpc.Client

	sTestsAcc = []func(t *testing.T){
		testAccountsInitCfg,
		testAccountsInitDataDb,
		testAccountsResetStorDb,
		testAccountsStartEngine,
		testAccountsRPCConn,

		testApierSetActions,
		testAccountsSetActPlans,
		testAccountsSet1,
		testAccountsGetActionPlan1,
		testAccountsSet2,
		testAccountsGetAccountActionPlan,
		testAccountsGetActionPlan2,

		testAccountsKillEngine,
	}
)

//Test start here
func TestAccountsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		accConfigDIR = "tutmysql"
	case utils.MetaMongo:
		accConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsAcc {
		t.Run(accConfigDIR, stest)
	}
}

func testAccountsInitCfg(t *testing.T) {
	var err error
	accCfgPath = path.Join(*dataDir, "conf", "samples", accConfigDIR)
	accCfg, err = config.NewCGRConfigFromPath(accCfgPath)
	if err != nil {
		t.Error(err)
	}
	accCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(accCfg)
}

func testAccountsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAccountsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(accCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccountsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAccountsRPCConn(t *testing.T) {
	var err error
	accRPC, err = newRPCClient(accCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testApierSetActions(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv1SetActions, &v1.V1AttrSetActions{
		ActionsId: "TestAccountAction",
		Actions: []*v1.V1TPAction{{
			Identifier:  utils.MetaTopUpReset,
			BalanceType: utils.MetaMonetary,
			Units:       75.0,
			ExpiryTime:  utils.MetaUnlimited,
			Weight:      20.0,
		}},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActions received: %s", reply)
	}
}

func testAccountsSetActPlans(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv1SetActionPlan, &v1.AttrSetActionPlan{
		Id: "TestAccountAP1",
		ActionPlan: []*v1.AttrActionPlan{{
			ActionsId: "TestAccountAction",
			MonthDays: "1",
			Time:      "00:00:00",
			Weight:    20.0,
		}},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}
	if err := accRPC.Call(utils.APIerSv1SetActionPlan, &v1.AttrSetActionPlan{
		Id: "TestAccountAP2",
		ActionPlan: []*v1.AttrActionPlan{{
			ActionsId: "TestAccountAction",
			MonthDays: "2",
			Time:      "00:00:00",
			Weight:    20.0,
		}},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}
	if err := accRPC.Call(utils.APIerSv1SetActionPlan, &v1.AttrSetActionPlan{
		Id: "TestAccountAP3",
		ActionPlan: []*v1.AttrActionPlan{{
			ActionsId: "TestAccountAction",
			MonthDays: "2",
			Time:      "00:00:00",
			Weight:    20.0,
		}},
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetActionPlan: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetActionPlan received: %s", reply)
	}
}

func testAccountsSet1(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv2SetAccount, AttrSetAccount{
		Tenant:               "cgrates.org",
		Account:              "dan",
		ReloadScheduler:      true,
		ActionPlanIDs:        []string{"TestAccountAP1", "TestAccountAP3"},
		ActionPlansOverwrite: true,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetAccount received: %s", reply)
	}
	if err := accRPC.Call(utils.APIerSv2SetAccount, AttrSetAccount{
		Tenant:               "cgrates.org",
		Account:              "dan2",
		ReloadScheduler:      true,
		ActionPlanIDs:        []string{"TestAccountAP1"},
		ActionPlansOverwrite: true,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetAccount received: %s", reply)
	}
}

func testAccountsGetActionPlan1(t *testing.T) {
	var aps []*engine.ActionPlan
	accIDsStrMp := utils.StringMap{
		"cgrates.org:dan":  true,
		"cgrates.org:dan2": true,
	}
	if err := accRPC.Call(utils.APIerSv1GetActionPlan,
		v1.AttrGetActionPlan{ID: "TestAccountAP1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "TestAccountAP1" {
		t.Errorf("Expected: %v,\n received: %v", "TestAccountAP1", aps[0].Id)
	} else if !reflect.DeepEqual(aps[0].AccountIDs, accIDsStrMp) {
		t.Errorf("Expected: %v,\n received: %v", accIDsStrMp, aps[0].AccountIDs)
	}
}

func testAccountsSet2(t *testing.T) {
	var reply string
	if err := accRPC.Call(utils.APIerSv2SetAccount, AttrSetAccount{
		Tenant:               "cgrates.org",
		Account:              "dan",
		ReloadScheduler:      true,
		ActionPlanIDs:        []string{"TestAccountAP2"},
		ActionPlansOverwrite: true,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetAccount received: %s", reply)
	}
}

func testAccountsGetAccountActionPlan(t *testing.T) {
	var reply []*v1.AccountActionTiming
	if err := accRPC.Call(utils.APIerSv1GetAccountActionPlan, utils.TenantAccount{
		Tenant:  "cgrates.org",
		Account: "dan",
	}, &reply); err != nil {
		t.Error("Got error on APIerSv1.GetAccountActionPlan: ", err.Error())
	} else if len(reply) != 1 {
		t.Error("Unexpected action plan received: ", utils.ToJSON(reply))
	} else if reply[0].ActionPlanId != "TestAccountAP2" {
		t.Errorf("Unexpected ActionoveAccountPlanId received")
	}
}

func testAccountsGetActionPlan2(t *testing.T) {
	var aps []*engine.ActionPlan
	accIDsStrMp := utils.StringMap{
		"cgrates.org:dan2": true,
	}
	if err := accRPC.Call(utils.APIerSv1GetActionPlan,
		v1.AttrGetActionPlan{ID: "TestAccountAP1"}, &aps); err != nil {
		t.Error(err)
	} else if len(aps) != 1 {
		t.Errorf("Expected: %v,\n received: %v", 1, len(aps))
	} else if aps[0].Id != "TestAccountAP1" {
		t.Errorf("Expected: %v,\n received: %v", "TestAccountAP1", aps[0].Id)
	} else if !reflect.DeepEqual(aps[0].AccountIDs, accIDsStrMp) {
		t.Errorf("Expected: %v,\n received: %v", accIDsStrMp, aps[0].AccountIDs)
	}
}

func testAccountsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
