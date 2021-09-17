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

package apis

import (
	"path"
	"reflect"
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
		testAccSResetStorDb,
		testAccSStartEngine,
		testAccSRPCConn,
		testGetAccProfileBeforeSet,
		testAccSetAccProfile,
		testAccGetAccIDs,
		testAccGetAccIDsCount,
		testGetAccBeforeSet2,
		testAccSetAcc2,
		testAccGetAccIDs2,
		testAccGetAccIDsCount2,
		testAccRemoveAcc,
		testAccGetAccountsForEvent,
		testAccMaxAbstracts,
		testAccDebitAbstracts,
		testAccMaxConcretes,
		testAccDebitConcretes,
		testAccActionSetRmvBalance,
		testAccSKillEngine,
	}
)

func TestAccSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		accPrfConfigDIR = "tutinternal"
	case utils.MetaMongo:
		accPrfConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		accPrfConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAccPrf {
		t.Run(accPrfConfigDIR, stest)
	}
}

func testAccSInitCfg(t *testing.T) {
	var err error
	accPrfCfgPath = path.Join(*dataDir, "conf", "samples", accPrfConfigDIR)
	accPrfCfg, err = config.NewCGRConfigFromPath(context.Background(), accPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAccSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testAccSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(accPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(accPrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testAccSRPCConn(t *testing.T) {
	var err error
	accSRPC, err = newRPCClient(accPrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
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

func testAccSetAccProfile(t *testing.T) {
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": "10",
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST",
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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
	args := &utils.PaginatorWithTenant{
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

func testAccGetAccIDsCount(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountCount,
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
	accPrf := &APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST2",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": "10",
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST2",
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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
	args := &utils.PaginatorWithTenant{
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

func testAccGetAccIDsCount2(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AdminSv1GetAccountCount,
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

func testAccGetAccountsForEvent(t *testing.T) {
	var reply []*utils.Account
	expected := []*utils.Account{
		{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST",
			Opts:   map[string]interface{}{},
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
					Opts: map[string]interface{}{
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
		Event: map[string]interface{}{
			utils.Usage: 20 * time.Second,
		},
		APIOpts: map[string]interface{}{
			utils.OptsAccountsAccountIDs: "TEST_ACC_IT_TEST",
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
	accPrf := APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "TEST_ACC_IT_TEST4",
			Weights:   ";0",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";25",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(0),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";20",
					Type:    utils.MetaConcrete,
					Units:   213,
					/*
						CostIncrements: []*utils.APICostIncrement{
							{
								Increment:    utils.Float64Pointer(float64(time.Second)),
								FixedFee:     utils.Float64Pointer(0),
								RecurrentFee: utils.Float64Pointer(0),
							},
						},

					*/
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.ExtEventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:              "27s",
			utils.OptsAccountsAccountIDs: "TEST_ACC_IT_TEST4",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1MaxAbstracts,
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
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IncrementStart:    nil,
				IntervalRateIndex: 0,
				RateID:            "",
				CompressFactor:    0,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range reply3.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	reply3.Rating = map[string]*utils.ExtRateSInterval{}
	expected2 := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(27000000000),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:       "TEST_ACC_IT_TEST4",
				BalanceID:       "AbstractBalance1",
				Units:           utils.Float64Pointer(27000000000),
				BalanceLimit:    utils.Float64Pointer(0),
				UnitFactorID:    "",
				RatingID:        rtID,
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TEST_ACC_IT_TEST4": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST4",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.ExtCostIncrement{
							{
								Increment:    utils.Float64Pointer(1000000000),
								FixedFee:     utils.Float64Pointer(0),
								RecurrentFee: utils.Float64Pointer(0),
							},
						},
						Units: utils.Float64Pointer(13000000000),
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
						Units: utils.Float64Pointer(213),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(reply3, expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccDebitAbstracts(t *testing.T) {
	accPrf := APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "TEST_ACC_IT_TEST5",
			Weights:   ";0",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";25",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(0),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";20",
					Type:    utils.MetaConcrete,
					Units:   213,
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.ExtEventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:              "27s",
			utils.OptsAccountsAccountIDs: "TEST_ACC_IT_TEST5",
		},
	}
	if err := accSRPC.Call(context.Background(), utils.AccountSv1DebitAbstracts,
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
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IncrementStart:    nil,
				IntervalRateIndex: 0,
				RateID:            "",
				CompressFactor:    0,
				Usage:             nil,
			},
		},
		CompressFactor: 1,
	}
	for _, val := range reply3.Rating {
		if !reflect.DeepEqual(val, expRating) {
			t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expRating, val)
		}
	}
	reply3.Rating = map[string]*utils.ExtRateSInterval{}
	expected2 := utils.ExtEventCharges{
		Abstracts: utils.Float64Pointer(27000000000),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:       "TEST_ACC_IT_TEST5",
				BalanceID:       "AbstractBalance1",
				Units:           utils.Float64Pointer(27000000000),
				BalanceLimit:    utils.Float64Pointer(0),
				UnitFactorID:    "",
				RatingID:        rtID,
				JoinedChargeIDs: nil,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TEST_ACC_IT_TEST5": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST5",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.ExtCostIncrement{
							{
								Increment:    utils.Float64Pointer(1000000000),
								FixedFee:     utils.Float64Pointer(0),
								RecurrentFee: utils.Float64Pointer(0),
							},
						},
						Units: utils.Float64Pointer(13000000000),
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
						Units: utils.Float64Pointer(213),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(reply3, expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccMaxConcretes(t *testing.T) {
	accPrf := APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "TEST_ACC_IT_TEST6",
			Weights:   ";0",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";25",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(0),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";20",
					Type:    utils.MetaConcrete,
					Units:   213,
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.ExtEventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesIntervalStart: "27s",
			utils.OptsAccountsAccountIDs: "TEST_ACC_IT_TEST6",
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
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IntervalRateIndex: 0,
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
	reply3.Rating = map[string]*utils.ExtRateSInterval{}
	expected2 := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(213),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:    "TEST_ACC_IT_TEST6",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.Float64Pointer(213),
				BalanceLimit: utils.Float64Pointer(0),
				UnitFactorID: "",
				RatingID:     rtID,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TEST_ACC_IT_TEST6": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST6",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.ExtCostIncrement{
							{
								Increment:    utils.Float64Pointer(1000000000),
								FixedFee:     utils.Float64Pointer(0),
								RecurrentFee: utils.Float64Pointer(0),
							},
						},
						Units: utils.Float64Pointer(40000000000),
					},
					"ConcreteBalance2": {
						ID: "ConcreteBalance2",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  "*concrete",
						Units: utils.Float64Pointer(0),
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
	accPrf := APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "TEST_ACC_IT_TEST7",
			Weights:   ";0",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:      "AbstractBalance1",
					Weights: ";25",
					Type:    utils.MetaAbstract,
					Units:   float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(0),
						},
					},
				},
				"ConcreteBalance2": {
					ID:      "ConcreteBalance2",
					Weights: ";20",
					Type:    utils.MetaConcrete,
					Units:   213,
				},
			},
		},
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	var reply3 utils.ExtEventCharges
	ev2 := &utils.CGREvent{
		Tenant: utils.CGRateSorg,
		ID:     "testIDEvent",
		Event: map[string]interface{}{
			utils.AccountField: "1004",
		},
		APIOpts: map[string]interface{}{
			utils.OptsRatesUsage:         "27s",
			utils.OptsAccountsAccountIDs: "TEST_ACC_IT_TEST7",
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
	expRating := &utils.ExtRateSInterval{
		IntervalStart: nil,
		Increments: []*utils.ExtRateSIncrement{
			{
				IntervalRateIndex: 0,
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
	reply3.Rating = map[string]*utils.ExtRateSInterval{}
	expected2 := utils.ExtEventCharges{
		Concretes: utils.Float64Pointer(213),
		Charges: []*utils.ChargeEntry{
			{
				ChargingID:     crgID,
				CompressFactor: 1,
			},
		},
		Accounting: map[string]*utils.ExtAccountCharge{
			accKEy: &utils.ExtAccountCharge{
				AccountID:    "TEST_ACC_IT_TEST7",
				BalanceID:    "ConcreteBalance2",
				Units:        utils.Float64Pointer(213),
				BalanceLimit: utils.Float64Pointer(0),
				UnitFactorID: "",
				RatingID:     rtID,
			},
		},
		UnitFactors: map[string]*utils.ExtUnitFactor{},
		Rating:      map[string]*utils.ExtRateSInterval{},
		Rates:       map[string]*utils.ExtIntervalRate{},
		Accounts: map[string]*utils.ExtAccount{
			"TEST_ACC_IT_TEST7": {
				Tenant:    "cgrates.org",
				ID:        "TEST_ACC_IT_TEST7",
				FilterIDs: []string{"*string:~*req.Account:1004"},
				Weights: utils.DynamicWeights{
					{
						Weight: 0,
					},
				},
				Balances: map[string]*utils.ExtBalance{
					"AbstractBalance1": {
						ID: "AbstractBalance1",
						Weights: utils.DynamicWeights{
							{
								Weight: 25,
							},
						},
						Type: "*abstract",
						CostIncrements: []*utils.ExtCostIncrement{
							{
								Increment:    utils.Float64Pointer(1000000000),
								FixedFee:     utils.Float64Pointer(0),
								RecurrentFee: utils.Float64Pointer(0),
							},
						},
						Units: utils.Float64Pointer(40000000000),
					},
					"ConcreteBalance2": {
						ID: "ConcreteBalance2",
						Weights: utils.DynamicWeights{
							{
								Weight: 20,
							},
						},
						Type:  "*concrete",
						Units: utils.Float64Pointer(0),
					},
				},
			},
		},
	}

	if !reflect.DeepEqual(reply3, expected2) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected2), utils.ToJSON(reply3))
	}
}

func testAccActionSetRmvBalance(t *testing.T) {
	accPrf := APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant: "cgrates.org",
			ID:     "TEST_ACC_IT_TEST8",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.APIBalance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights:   ";12",
					Type:      "*abstract",
					Opts: map[string]interface{}{
						"Destination": "10",
					},
					Units: 0,
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	var reply string
	if err := accSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf, &reply); err != nil {
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
	}
	if !reflect.DeepEqual(reply3, `OK`) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(`OK`), utils.ToJSON(reply3))
	}

	expectedAcc := utils.Account{
		Tenant: "cgrates.org",
		ID:     "TEST_ACC_IT_TEST8",
		Opts:   map[string]interface{}{},
		Balances: map[string]*utils.Balance{
			"AbstractBalance3": {
				ID:          "AbstractBalance3",
				FilterIDs:   nil,
				Weights:     nil,
				Type:        "*concrete",
				Units:       utils.NewDecimal(10, 0),
				UnitFactors: nil,
				Opts:        nil,
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
				Opts: map[string]interface{}{
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
		Opts:   map[string]interface{}{},
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
				Opts: map[string]interface{}{
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

//Kill the engine when it is about to be finished
func testAccSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
