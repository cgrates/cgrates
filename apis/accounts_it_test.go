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

package apis

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	accPrfCfgPath   string
	accPrfCfg       *config.CGRConfig
	accSRPC         *birpc.Client
	accPrfConfigDIR string //run tests for specific configuration

	sTestsAccPrf = []func(t *testing.T){
		testAccSInitCfg,
		testAccSInitDataDb,
		testAccSStartEngine,
		testAccSRPCConn,
		testGetAccProfileBeforeSet,
		testGetAccProfilesBeforeSet,
		testAccSetAccProfile,
		testAccGetAccIDs,
		testAccGetAccs,
		testAccGetAccIDsCount,
		testGetAccBeforeSet2,
		testAccSetAcc2,
		testAccGetAccIDs2,
		testAccGetAccs2,
		testAccGetAccIDsCount2,
		testAccRemoveAcc,
		testAccGetAccs3,
		testAccGetAccsWithPrefix,
		testAccGetAccountsForEvent,
		testAccMaxAbstracts,
		testAccDebitAbstracts,
		testAccMaxConcretes,
		testAccDebitConcretes,
		// RefundCharges test
		testAccRefundCharges,
		testAccActionSetRmvBalance,
		// Account with blocker debit
		testAccSInitDataDb,
		testAccSCacheClear,
		testAccDebitAbstractWithoutBlockers,
		testAccDebitAbstractWithBlockers,
		testAccDebitAbstractWithBlockersOnBalance,
		testAccSKillEngine,
	}
)

func TestAccSIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		accPrfConfigDIR = "tutinternal"
	case utils.MetaMongo:
		accPrfConfigDIR = "tutmongo"
	case utils.MetaRedis:
		accPrfConfigDIR = "tutredis"
	case utils.MetaMySQL:
		accPrfConfigDIR = "mysql_acc"
	case utils.MetaPostgres:
		accPrfConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAccPrf {
		t.Run(accPrfConfigDIR, stest)
	}
}

func testAccSInitCfg(t *testing.T) {
	var err error
	accPrfCfgPath = path.Join(*utils.DataDir, "conf", "samples", accPrfConfigDIR)
	accPrfCfg, err = config.NewCGRConfigFromPath(context.Background(), accPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAccSInitDataDb(t *testing.T) {
	if err := engine.InitDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accPrfCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testAccSRPCConn(t *testing.T) {
	accSRPC = engine.NewRPCClient(t, accPrfCfg.ListenCfg(), *utils.Encoding)
}

func testGetAccProfileBeforeSet(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ACC_IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testGetAccProfilesBeforeSet(t *testing.T) {
	var reply *[]*utils.Account
	args := &utils.ArgsItemIDs{}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAccSetAccProfile(t *testing.T) {
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	var result utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "TEST_ACC_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}
}

func testAccGetAccIDs(t *testing.T) {
	var reply []string
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_ACC_IT_TEST"}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testAccGetAccs(t *testing.T) {
	var reply *[]*utils.Account
	args := &utils.ArgsItemIDs{}
	expected := &[]*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testAccGetAccIDsCount(t *testing.T) {
	var reply int
	args := &utils.ArgsItemIDs{
		Tenant:      utils.CGRateSorg,
		ItemsSearch: "",
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountsCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected %+v \n, received %+v", 1, reply)
	}
}

func testGetAccBeforeSet2(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ACC_IT_TEST_SECOND",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAccSetAcc2(t *testing.T) {
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST2",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	var result utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "TEST_ACC_IT_TEST2",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}
}

func testAccGetAccIDs2(t *testing.T) {
	var reply []string
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_ACC_IT_TEST", "TEST_ACC_IT_TEST2"}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testAccGetAccs2(t *testing.T) {
	var reply *[]*utils.Account
	args := &utils.ArgsItemIDs{}
	expected := &[]*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}

	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		args, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(*reply, func(i, j int) bool {
		return (*reply)[i].ID < (*reply)[j].ID
	})
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testAccGetAccIDsCount2(t *testing.T) {
	var reply int
	args := &utils.ArgsItemIDs{
		Tenant:      utils.CGRateSorg,
		ItemsSearch: "",
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountsCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("Expected %+v \n, received %+v", 2, reply)
	}
}

func testAccRemoveAcc(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "TEST_ACC_IT_TEST2",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1RemoveAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	//nothing to get from db
	var result *utils.RateProfile
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ACC_IT_TEST2",
			},
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v \n, received %+v", utils.ErrNotFound, err)
	}
}

func testAccGetAccs3(t *testing.T) {
	var reply *[]*utils.Account
	args := &utils.ArgsItemIDs{}
	expected := &[]*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}
func testAccGetAccsWithPrefix(t *testing.T) {
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "aTEST_ACC_IT_TEST2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "aTEST_ACC_IT_TEST2",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	var result utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "aTEST_ACC_IT_TEST2",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}

	var reply2 *[]*utils.Account
	args := &utils.ArgsItemIDs{
		ItemsSearch: "aTEST",
	}
	expected := &[]*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "aTEST_ACC_IT_TEST2",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccounts,
		args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expected), utils.ToJSON(reply2))
	}
}

func testAccGetAccountsForEvent(t *testing.T) {
	var reply []*utils.Account
	expected := []*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							FilterIDs: nil,
							Weight:    12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					FilterIDs: nil,
					Weight:    10,
				},
			},
		},
	}
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.Usage: 20 * time.Second,
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs: "TEST_ACC_IT_TEST",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1AccountsForEvent,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testAccMaxAbstracts(t *testing.T) {
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST4",
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(40*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(213, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 *utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:              "27s",
			utils.OptsAccountsProfileIDs: "TEST_ACC_IT_TEST4",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1MaxAbstracts,
		ev2, &reply3); err != nil {
		t.Error(err)
	}
	expected2 := &utils.EventCharges{
		Abstracts: utils.NewDecimal(int64(27*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "charge1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"charge1": {
				AccountID:    "TEST_ACC_IT_TEST4",
				BalanceID:    "AbstractBalance1",
				Units:        utils.NewDecimal(int64(27*time.Second), 0),
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     "Rating1",
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating: map[string]*utils.RateSInterval{
			"Rating1": {
				Increments: []*utils.RateSIncrement{
					{
						RateIntervalIndex: 0,
						RateID:            "rate1",
						CompressFactor:    1,
					}},
				CompressFactor: 1,
			}},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				FixedFee:     utils.NewDecimal(0, 0),
				RecurrentFee: utils.NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TEST_ACC_IT_TEST4": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST4",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimal(int64(13*time.Second), 0),
					},
					"ConcreteBalance2": {
						ID:        "ConcreteBalance2",
						FilterIDs: nil,
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type: "*concrete",
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimal(213, 0),
					},
				},
			},
		},
	}

	if !reply3.Equals(expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccDebitAbstracts(t *testing.T) {
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST5",
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(40*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(213, 0),
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:              "27s",
			utils.OptsAccountsProfileIDs: "TEST_ACC_IT_TEST5",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1DebitAbstracts,
		ev2, &reply3); err != nil {
		t.Error(err)
	}

	expected2 := utils.EventCharges{
		Abstracts: utils.NewDecimal(int64(27*time.Second), 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     "charge1",
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			"charge1": {
				AccountID:       "TEST_ACC_IT_TEST5",
				BalanceID:       "AbstractBalance1",
				Units:           utils.NewDecimal(int64(27*time.Second), 0),
				BalanceLimit:    utils.NewDecimal(0, 0),
				RatingID:        "rating1",
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating: map[string]*utils.RateSInterval{
			"rating1": {
				Increments: []*utils.RateSIncrement{{
					RateIntervalIndex: 0,
					RateID:            "rate1",
					CompressFactor:    1,
				}},
				CompressFactor: 1,
			},
		},
		Rates: map[string]*utils.IntervalRate{
			"rate1": {
				FixedFee:     utils.NewDecimal(0, 0),
				RecurrentFee: utils.NewDecimal(0, 0),
			},
		},
		Accounts: map[string]*utils.Account{
			"TEST_ACC_IT_TEST5": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST5",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimal(int64(13*time.Second), 0),
					},
					"ConcreteBalance2": {
						ID:        "ConcreteBalance2",
						FilterIDs: nil,
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  "*concrete",
						Units: utils.NewDecimal(213, 0),
					},
				},
			},
		},
	}

	if !expected2.Equals(&reply3) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccMaxConcretes(t *testing.T) {
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST6",
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(40*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(213, 0),
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.OptsRatesIntervalStart: "27s",
			utils.OptsAccountsProfileIDs: "TEST_ACC_IT_TEST6",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1MaxConcretes,
		ev2, &reply3); err != nil {
		t.Error(err)
	}

	var crgID string
	for _, val := range reply3.Charges {
		crgID = val.ChargingID
	}

	var accKEy, rtID string
	for key, val := range reply3.Accounting {
		accKEy = key
		rtID = val.RatingID
	}
	expRating := &utils.RateSInterval{
		IntervalStart: nil,
		Increments: []*utils.RateSIncrement{
			{
				RateIntervalIndex: 0,
				RateID:            utils.EmptyString,
				CompressFactor:    0,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range reply3.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	reply3.Rating = map[string]*utils.RateSInterval{}
	expected2 := utils.EventCharges{
		Concretes: utils.NewDecimal(213, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			accKEy: {
				AccountID:    "TEST_ACC_IT_TEST6",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(213, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     rtID,
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TEST_ACC_IT_TEST6": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST6",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimal(int64(40*time.Second), 0),
					},
					"ConcreteBalance2": {
						ID: "ConcreteBalance2",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  "*concrete",
						Units: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(reply3, expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccDebitConcretes(t *testing.T) {
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST7",
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance1": {
					ID: "AbstractBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(40*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance2": {
					ID: "ConcreteBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(213, 0),
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsUsage:      "27s",
			utils.OptsAccountsProfileIDs: "TEST_ACC_IT_TEST7",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1DebitConcretes,
		ev2, &reply3); err != nil {
		t.Error(err)
	}

	var crgID string
	for _, val := range reply3.Charges {
		crgID = val.ChargingID
	}

	var accKEy, rtID string
	for key, val := range reply3.Accounting {
		accKEy = key
		rtID = val.RatingID
	}
	expRating := &utils.RateSInterval{
		Increments: []*utils.RateSIncrement{
			{
				RateIntervalIndex: 0,
				CompressFactor:    0,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range reply3.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	reply3.Rating = map[string]*utils.RateSInterval{}
	expected2 := utils.EventCharges{
		Concretes: utils.NewDecimal(213, 0),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.AccountCharge{
			accKEy: {
				AccountID:    "TEST_ACC_IT_TEST7",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.NewDecimal(213, 0),
				BalanceLimit: utils.NewDecimal(0, 0),
				RatingID:     rtID,
			},
		},
		UnitFactors: map[string]*utils.UnitFactor{},
		Rating:      map[string]*utils.RateSInterval{},
		Rates:       map[string]*utils.IntervalRate{},
		Accounts: map[string]*utils.Account{
			"TEST_ACC_IT_TEST7": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST7",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.Balance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(int64(time.Second), 0),
								FixedFee:     utils.NewDecimal(0, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimal(int64(40*time.Second), 0),
					},
					"ConcreteBalance2": {
						ID: "ConcreteBalance2",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  "*concrete",
						Units: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(reply3, expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccRefundCharges(t *testing.T) {
	// we will set an account, we will debit it (with debitAbtracts), and after that we will call refundCharges to get back our units from the cost
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "AccountRefundCharges",
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			FilterIDs: []string{"*exists:~*opts.*acntUsage:", "*string:~*req.Destination:1004"},
			Balances: map[string]*utils.Balance{
				"AB": {
					ID: "AB",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(5*time.Minute), 0), // 300s
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(10, 0),
						},
					},
					Opts: map[string]any{
						utils.MetaBalanceLimit: -100.0,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(6, 1),
							RecurrentFee: utils.NewDecimal(1, 1),
						},
					},
				},
				"CB": {
					ID: "CB",
					Weights: utils.DynamicWeights{
						{
							Weight: 5,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(500, 1),
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(15, 0),
						},
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee: utils.NewDecimal(1, 1),
						},
					},
				},
			},
		},
	}
	var replyStr string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error(err)
	}

	var reply utils.EventCharges
	ev := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "tesRefundCHarges",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1004",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsUsage:      "3m27s",
			utils.OptsAccountsProfileIDs: "AccountRefundCharges",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1DebitAbstracts,
		ev, &reply); err != nil {
		t.Error(err)
	}
	// we will compare the costs of the eventCharges and the abstracts/concretes
	concreteCost := utils.NewDecimal(33, 1) // 3.3 were debited of concretes (3.3 intially and because of 15 unit factor of CB --> 3.3 * 15 = 49.5 units were taken from CB)

	abstractCost := utils.NewDecimal(int64(27*time.Second), 0) // 27s were debited of abstracts,
	if !reflect.DeepEqual(abstractCost, reply.Abstracts) {
		t.Errorf("Expected %v, received %v", abstractCost, reply.Abstracts)
	}
	if !reflect.DeepEqual(concreteCost, reply.Concretes) {
		t.Errorf("Expected %v, received %v", concreteCost, reply.Concretes)
	}

	// 50 - 49.5(3.3 * 15 uf) = 0.5   3m27s --> 207 seconds -->  2.7 debit + 0.6 fixedFee = 3.3 debited

	// we will get the Account after the debit was made
	var result *utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "AccountRefundCharges",
			},
		}, &result); err != nil {
		t.Error(err)
	} else {
		//now we will compare the units from both balances to see that the debit took the untis from account
		astractUnitsRemain := utils.NewDecimal(int64(30*time.Second), 0) // 5m - 27s because of uf --> 300s - 270s(27 * 10uf) = 30s
		if !reflect.DeepEqual(result.Balances["AB"].Units, astractUnitsRemain) {
			t.Errorf("Expected %v, received %v", astractUnitsRemain, result.Balances["AB"].Units)
		}
		concretesUnitsRemain := utils.NewDecimal(5, 1) // 50 - 49.5
		if !reflect.DeepEqual(result.Balances["CB"].Units, concretesUnitsRemain) {
			t.Errorf("Expected %v, received %v", concretesUnitsRemain, result.Balances["CB"].Units)
		}
	}

	// now as the account was debited, we will refund those charges
	args := &utils.APIEventCharges{
		Tenant:       "cgrates.org",
		EventCharges: &reply, // this reply was the eventCharges from the previous debit
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1RefundCharges,
		args, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != utils.OK {
		t.Error(err)
	}

	// now that the refund was called properly, we will check the account to see the changes
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "AccountRefundCharges",
			},
		}, &result); err != nil {
		t.Error(err)
	} else {
		if *utils.DBType == utils.MetaPostgres || accPrfConfigDIR == "mysql_acc" {
			acc.Account.Balances["CB"].Units = utils.NewDecimalFromFloat64(50)
		}
		if !reflect.DeepEqual(result, acc.Account) {
			t.Errorf("Expected %+v \n, received %+v", acc.Account, result)
		}
	}
}

func testAccActionSetRmvBalance(t *testing.T) {
	acc := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST8",
			Opts:   map[string]any{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]any{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	var reply3 string
	args2 := &utils.ArgsActSetBalance{
		AccountID: "TEST_ACC_IT_TEST8",
		Tenant:    "cgrates.org",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.AbstractBalance3.Units",
				Value: "10",
			},
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1ActionSetBalance,
		args2, &reply3); err != nil {
		t.Error(err)
	} else if reply3 != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, utils.OK)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST8",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"AbstractBalance3": {
				ID:    "AbstractBalance3",
				Type:  "*concrete",
				Units: utils.NewDecimal(10, 0),
				CostIncrements: []*utils.CostIncrement{
					{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    utils.NewDecimal(1000000000, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    utils.NewDecimal(1048576, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
					{
						FilterIDs:    []string{"*string:~*req.ToR:*sms"},
						Increment:    utils.NewDecimal(1, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						Weight: 12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}
	var result utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "TEST_ACC_IT_TEST8",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}

	var reply4 string
	args3 := &utils.ArgsActRemoveBalances{
		Tenant:     "",
		AccountID:  "TEST_ACC_IT_TEST8",
		BalanceIDs: []string{"AbstractBalance3"},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1ActionRemoveBalance,
		args3, &reply4); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(reply4, `OK`) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(`OK`), utils.ToJSON(reply4))
	}
	expectedAcc2 := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST8",
		Opts:   map[string]any{},
		Balances: map[string]*utils.Balance{
			"VoiceBalance": {
				ID:        "VoiceBalance",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    12,
					},
				},
				Type: "*abstract",
				Opts: map[string]any{
					"Destination": "10",
				},
				Units: utils.NewDecimal(0, 0),
			},
		},
		Weights: utils.DynamicWeights{
			{
				FilterIDs: nil,
				Weight:    10,
			},
		},
	}
	var result2 utils.Account
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "",
				ID:     "TEST_ACC_IT_TEST8",
			},
		}, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, expectedAcc2) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc2), utils.ToJSON(result2))
	}

}

func testAccSCacheClear(t *testing.T) {
	var reply string
	if err := accSRPC.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			CacheIDs: nil,
		}, &reply); err != nil {
		t.Error(err)
	}
}

func testAccDebitAbstractWithoutBlockers(t *testing.T) {
	acc1 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "AccountBlocker1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			FilterIDs: []string{"*string:~*req.Blockers:*exists"},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Balances: map[string]*utils.Balance{
				"AB1": {
					ID: "AB1",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(70*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							FixedFee:     utils.NewDecimal(6, 1),
							RecurrentFee: utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
				"CB1": {
					ID: "CB1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(10, 0),
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(999, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(2, 0),
							RecurrentFee: utils.NewDecimal(1, 0),
						},
					},
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	acc2 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "AccountBlocker2",
			Weights: utils.DynamicWeights{
				{
					Weight: 15,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					FilterIDs: []string{"*string:~*req.Destination:1002"},
					Blocker:   true,
				},
			},
			FilterIDs: []string{"*string:~*req.Blockers:*exists"},
			Balances: map[string]*utils.Balance{
				"AB_ForBlocker": {
					ID: "AB_ForBLocker",
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(5*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(3*time.Second), 0),
							RecurrentFee: utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
				"CB_WIthBlocker": {
					ID: "CB_WIthBlocker",
					Weights: utils.DynamicWeights{
						{
							Weight: 28,
						},
					},
					Blockers: utils.DynamicBlockers{
						{
							FilterIDs: []string{"*string:~*req.BlockerAbstract:yes"},
							Blocker:   true,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(5, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(1, 0),
							RecurrentFee: utils.NewDecimal(1, 0),
						},
					},
				},
				"AB2": {
					ID: "AB2",
					Weights: utils.DynamicWeights{
						{
							Weight: 25,
						},
					},
					Type:  utils.MetaAbstract,
					Units: utils.NewDecimal(int64(20*time.Second), 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee: utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
				"CB2": {
					ID: "CB2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaConcrete,
					Units: utils.NewDecimal(100, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(1, 0),
							RecurrentFee: utils.NewDecimal(1, 0),
						},
					},
				},
			},
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acc2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	// We will try to use de maxAbstratct to see the cost for from both accounts matched, without blocker for now. In order to match blocker, Destination must be 1002
	var replyEv *utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			"Blockers":         "*exists",
			utils.AccountField: "101223",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 90 * time.Second,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1MaxAbstracts,
		ev2, &replyEv); err != nil {
		t.Error(err)
	} else {
		// AccountBlocker1 should not be debited!!!
		if _, has := replyEv.Accounts["AccountBlocker1"]; has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker1")
		}
		if _, has := replyEv.Accounts["AccountBlocker2"]; !has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker2")
		}
	}
}

func testAccDebitAbstractWithBlockers(t *testing.T) {
	// we will try to use de maxAbstratct to see the cost for from one account matched, with blocker for now (it will be AccountBlocker2)
	var replyEv *utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			"Blockers":         "*exists",
			utils.AccountField: "101223",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 90 * time.Second,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1MaxAbstracts,
		ev2, &replyEv); err != nil {
		t.Error(err)
	} else {
		// the debit worked properly, it was debited from both accounts, but the next accoutns from the balance named "CB_WIthBlocker" should not be debited
		if _, has := replyEv.Accounts["AccountBlocker1"]; !has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker1")
		}
		if _, has := replyEv.Accounts["AccountBlocker2"]; !has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker2")
		}
	}
}

func testAccDebitAbstractWithBlockersOnBalance(t *testing.T) {
	// AB_WithBLocker balance from AccountBLocker2 account will be the first balance to be charged, but it has blocker when BlockerAbstract field exist, this will stop deibting the other balances from AccountBLocker2
	var replyEv *utils.EventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]any{
			"Blockers":         "*exists",
			utils.AccountField: "101223",
			"BlockerAbstract":  "yes",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 90 * time.Second,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1DebitAbstracts,
		ev2, &replyEv); err != nil {
		t.Error(err)
	} else {
		// the debit worked properly, it was debited from both accounts
		if _, has := replyEv.Accounts["AccountBlocker1"]; !has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker1")
		}
		if _, has := replyEv.Accounts["AccountBlocker2"]; !has {
			t.Errorf("The accounts %v was not debited", "AccountBlocker2")
		} else {
			// CB_WithBlocker balance has a blocker, so the next two balances will be skipped
			for _, acc := range replyEv.Accounting {
				if acc.BalanceID == "AB2" || acc.BalanceID == "CB2" {
					t.Errorf("The balance <%s> from <AccountBlocker2> should not be debited", acc.BalanceID)
				}
			}
		}
	}
}

// Kill the engine when it is about to be finished
func testAccSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
