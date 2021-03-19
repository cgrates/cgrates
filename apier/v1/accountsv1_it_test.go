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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

var (
	acntSConfigDIR string //run tests for specific configuration
	acntSCfgPath   string
	acntSCfg       *config.CGRConfig
	acntSRPC       *rpc.Client
)

//Test start here
func TestAccountSv1IT(t *testing.T) {
	sTestsAccountS := []func(t *testing.T){
		testAccountSv1InitCfg,
		testAccountSv1InitDataDb,
		testAccountSv1ResetStorDb,
		testAccountSv1StartEngine,
		testAccountSv1RPCConn,
		testAccountSv1LoadFromFolder,
		//testAccountSv1AccountProfilesForEvent,
		//testAccountSv1MaxAbstracts,
		//testAccountSv1DebitAbstracts,
		//testAccountSv1SimpleDebit,
		//testAccountSv1DebitMultipleAcc,
		//testAccountSv1DebitMultipleAccLimited,
		testAccountSv1DebitWithAttributeSandRateS,
		//testAccountSv1DebitWithRateS,
		//testAccountSv1DebitWithRateS2,
		//testAccountSv1MaxConcretes,
		//testAccountSv1DebitConcretes,
		//testAccountSv1ActionSetBalance,
		//testAccountSv1ActionRemoveBalance,
		testAccountSv1KillEngine,
	}
	switch *dbType {
	case utils.MetaInternal:
		acntSConfigDIR = "accounts_internal"
	case utils.MetaMySQL:
		t.SkipNow()
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatalf("unknown Database type <%s>", *dbType)
	}
	for _, stest := range sTestsAccountS {
		t.Run(acntSConfigDIR, stest)
	}
}

func testAccountSv1InitCfg(t *testing.T) {
	var err error
	acntSCfgPath = path.Join(*dataDir, "conf", "samples", acntSConfigDIR)
	acntSCfg, err = config.NewCGRConfigFromPath(acntSCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testAccountSv1InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(acntSCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAccountSv1ResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(acntSCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAccountSv1StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(acntSCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAccountSv1RPCConn(t *testing.T) {
	var err error
	acntSRPC, err = newRPCClient(acntSCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAccountSv1LoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutaccounts")}
	if err := acntSRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAccountSv1AccountProfilesForEvent(t *testing.T) {
	eAcnts := []*utils.AccountProfile{
		{
			Tenant:    "cgrates.org",
			ID:        "1001",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"GenericBalance1": &utils.Balance{
					ID: "GenericBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Type:  utils.MetaAbstract,
					Units: &utils.Decimal{decimal.New(int64(time.Hour), 0)},
					UnitFactors: []*utils.UnitFactor{
						&utils.UnitFactor{
							FilterIDs: []string{"*string:~*req.ToR:*data"},
							Factor:    &utils.Decimal{decimal.New(1024, 3)},
						},
					},
					CostIncrements: []*utils.CostIncrement{
						&utils.CostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*voice"},
							Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
						},
						&utils.CostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*data"},
							Increment:    &utils.Decimal{decimal.New(1024, 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
						},
					},
				},
				"MonetaryBalance1": &utils.Balance{
					ID: "MonetaryBalance1",
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
					Type:  utils.MetaConcrete,
					Units: &utils.Decimal{decimal.New(5, 0)},
					CostIncrements: []*utils.CostIncrement{
						&utils.CostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*voice"},
							Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
						},
						&utils.CostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*data"},
							Increment:    &utils.Decimal{decimal.New(1024, 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
						},
					},
				},
				"MonetaryBalance2": &utils.Balance{
					ID: "MonetaryBalance2",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					Type:  utils.MetaConcrete,
					Units: &utils.Decimal{decimal.New(3, 0)},
					CostIncrements: []*utils.CostIncrement{
						&utils.CostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*voice"},
							Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
							FixedFee:     &utils.Decimal{decimal.New(0, 0)},
							RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var acnts []*utils.AccountProfile
	if err := acntSRPC.Call(utils.AccountSv1AccountProfilesForEvent,
		&utils.ArgsAccountsForEvent{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testAccountSv1AccountProfileForEvent",
				Event: map[string]interface{}{
					utils.AccountField: "1001",
				}}}, &acnts); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eAcnts, acnts) {
		t.Errorf("Expecting : %s \n received: %s", utils.ToJSON(eAcnts), utils.ToJSON(acnts))
	}
}

func testAccountSv1MaxAbstracts(t *testing.T) {
	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1MaxAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxUsage",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        "15m",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 800000000000.0 { // 500s from first monetary + 300s from last monetary
		t.Errorf("received usage: %v", *eEc.Abstracts)
	}

	// Make sure we did not Debit anything from Account
	eAcnt := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"GenericBalance1": &utils.Balance{
				ID: "GenericBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(int64(time.Hour), 0)},
				UnitFactors: []*utils.UnitFactor{
					&utils.UnitFactor{
						FilterIDs: []string{"*string:~*req.ToR:*data"},
						Factor:    &utils.Decimal{decimal.New(1024, 3)},
					},
				},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1024, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
				},
			},
			"MonetaryBalance1": &utils.Balance{
				ID: "MonetaryBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(5, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1024, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
				},
			},
			"MonetaryBalance2": &utils.Balance{
				ID: "MonetaryBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(3, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	var reply *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eAcnt, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(eAcnt), utils.ToJSON(reply))
	}
}

func testAccountSv1DebitAbstracts(t *testing.T) {
	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxUsage",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        "15m",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 800000000000.0 { // 500s from first monetary + 300s from last monetary
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	// Make sure we debit the right units from Account
	eAcnt := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1001",
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"GenericBalance1": &utils.Balance{
				ID: "GenericBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(int64(3300*time.Second), 0)},
				UnitFactors: []*utils.UnitFactor{
					&utils.UnitFactor{
						FilterIDs: []string{"*string:~*req.ToR:*data"},
						Factor:    &utils.Decimal{decimal.New(1024, 3)},
					},
				},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1024, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
				},
			},
			"MonetaryBalance1": &utils.Balance{
				ID: "MonetaryBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1024, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 2)},
					},
				},
			},
			"MonetaryBalance2": &utils.Balance{
				ID: "MonetaryBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*voice"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	var reply *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eAcnt, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(eAcnt), utils.ToJSON(reply))
	}
}

func testAccountSv1SimpleDebit(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomAccount",
			FilterIDs: []string{"*string:~*req.Account:CustomAccount"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(0.1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var result string
	expErr := utils.ErrNotFound.Error()
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &result); err == nil || err.Error() != expErr {
		t.Errorf("Expected error: %v received: %v", expErr, err)
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1SimpleDebit",
			Event: map[string]interface{}{
				utils.AccountField: "CustomAccount",
				utils.Usage:        "10",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 10.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(99, 0)) != 0 {
		t.Errorf("Expecting : %+v, received: %s", decimal.New(99, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1DebitMultipleAcc(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomAccount",
			FilterIDs: []string{"*string:~*req.Account:CustomAccount"},
			Weights:   ";20",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(0.1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	accPrfAPI2 := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomAccount2",
			FilterIDs: []string{"*string:~*req.Account:CustomAccount"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   50,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(0.1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var convAcc2 *utils.AccountProfile
	if convAcc2, err = accPrfAPI2.AsAccountProfile(); err != nil {
		t.Fatal(err)
	}
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount2"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc2, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc2, reply2)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1SimpleDebit",
			Event: map[string]interface{}{
				utils.AccountField: "CustomAccount",
				utils.Usage:        "1400",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 1400.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(0, 0)) != 0 {
		t.Errorf("Expecting : %s, received: %s", decimal.New(0, 0), reply2.Balances["Balance1"].Units)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount2"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(10, 0)) != 0 {
		t.Errorf("Expecting : %s, received: %s", decimal.New(10, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1DebitMultipleAccLimited(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomAccount",
			FilterIDs: []string{"*string:~*req.Account:CustomAccount"},
			Weights:   ";20",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
					Opts: map[string]interface{}{
						utils.MetaBalanceLimit: 50.0,
					},
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(0.1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	accPrfAPI2 := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomAccount2",
			FilterIDs: []string{"*string:~*req.Account:CustomAccount"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   50,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(0.1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var convAcc2 *utils.AccountProfile
	if convAcc2, err = accPrfAPI2.AsAccountProfile(); err != nil {
		t.Fatal(err)
	}
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount2"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc2, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc2, reply2)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1SimpleDebit",
			Event: map[string]interface{}{
				utils.AccountField: "CustomAccount",
				utils.Usage:        "900",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 900.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(50, 0)) != 0 {
		t.Errorf("Expecting : %s, received: %s", decimal.New(50, 0), reply2.Balances["Balance1"].Units)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomAccount2"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(10, 0)) != 0 {
		t.Errorf("Expecting : %s, received: %s", decimal.New(10, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1DebitWithAttributeSandRateS(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "ACC_WITH_ATTRIBUTES",
			FilterIDs: []string{"*string:~*req.Account:ACC_WITH_ATTRIBUTES"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(-1),
						},
					},
					AttributeIDs: []string{"*constant:*req.CustomField:CustomValue"},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_ATTRIBUTES"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	//set a rate profile to be used in case of debit
	apiRPrf := &utils.APIRateProfile{
		Tenant:  "cgrates.org",
		ID:      "RP_Test",
		Weights: ";10",
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				Weights:         ";0",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.1),
						Increment:     utils.Float64Pointer(1),
						Unit:          utils.Float64Pointer(1),
					},
				},
			},
		},
	}

	if err := acntSRPC.Call(utils.APIerSv1SetRateProfile,
		&utils.APIRateProfileWithOpts{
			APIRateProfile: apiRPrf,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1DebitWithAttributeS",
			Event: map[string]interface{}{
				utils.AccountField: "ACC_WITH_ATTRIBUTES",
				utils.Usage:        "10",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 10.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_ATTRIBUTES"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(99, 0)) != 0 {
		t.Errorf("Expecting : %+v, received: %s", decimal.New(99, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1DebitWithRateS(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "ACC_WITH_RATES",
			FilterIDs: []string{"*string:~*req.Account:ACC_WITH_RATES"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(-1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_RATES"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	//set a rate profile to be used in case of debit
	apiRPrf := &utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_Test2",
		FilterIDs: []string{"*string:~*req.Account:ACC_WITH_RATES"},
		Weights:   ";20",
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				Weights:         ";0",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.5),
						Increment:     utils.Float64Pointer(2),
						Unit:          utils.Float64Pointer(2),
					},
				},
			},
		},
	}

	if err := acntSRPC.Call(utils.APIerSv1SetRateProfile,
		&utils.APIRateProfileWithOpts{
			APIRateProfile: apiRPrf,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1DebitWithAttributeS",
			Event: map[string]interface{}{
				utils.AccountField: "ACC_WITH_RATES",
				utils.Usage:        "20",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 20.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_RATES"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(95, 0)) != 0 {
		t.Errorf("Expecting : %+v, received: %s", decimal.New(95, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1DebitWithRateS2(t *testing.T) {
	accPrfAPI := &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "ACC_WITH_RATES2",
			FilterIDs: []string{"*string:~*req.Account:ACC_WITH_RATES2"},
			Weights:   ";10",
			Balances: map[string]*utils.APIBalance{
				"Balance1": &utils.APIBalance{
					ID:      "Balance1",
					Weights: ";10",
					Type:    utils.MetaAbstract,
					Units:   100,
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(1),
							RecurrentFee: utils.Float64Pointer(-1),
						},
					},
					RateProfileIDs: []string{"RP_Test22"},
				},
				"Balance2": &utils.APIBalance{
					ID:      "Balance2",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   100,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, accPrfAPI, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var err error
	var convAcc *utils.AccountProfile
	if convAcc, err = accPrfAPI.AsAccountProfile(); err != nil {
		t.Error(err)
	}
	var reply2 *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_RATES2"}}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(convAcc, reply2) {
		t.Errorf("Expecting : %+v, received: %+v", convAcc, reply2)
	}

	//set a rate profile to be used in case of debit
	apiRPrf := &utils.APIRateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP_Test22",
		FilterIDs: []string{"*string:~*req.Account:ACC_WITH_RATES2"},
		Weights:   ";20",
		Rates: map[string]*utils.APIRate{
			"RT_ALWAYS": {
				ID:              "RT_ALWAYS",
				Weights:         ";0",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.5),
						Increment:     utils.Float64Pointer(2),
						Unit:          utils.Float64Pointer(2),
					},
				},
			},
		},
	}

	if err := acntSRPC.Call(utils.APIerSv1SetRateProfile,
		&utils.APIRateProfileWithOpts{
			APIRateProfile: apiRPrf,
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitAbstracts,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1DebitWithAttributeS",
			Event: map[string]interface{}{
				utils.AccountField: "ACC_WITH_RATES2",
				utils.Usage:        "20",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Abstracts == nil || *eEc.Abstracts != 20.0 {
		t.Fatalf("received usage: %v", *eEc.Abstracts)
	}

	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "ACC_WITH_RATES2"}}, &reply2); err != nil {
		t.Error(err)
	} else if reply2.Balances["Balance1"].Units.Cmp(decimal.New(80, 0)) != 0 {
		t.Errorf("Expecting : %+v, received: %s", decimal.New(80, 0), reply2.Balances["Balance1"].Units)
	}
}

func testAccountSv1MaxConcretes(t *testing.T) {
	apiAccPrf = &utils.APIAccountProfileWithOpts{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    "cgrates.org",
			ID:        "1004",
			FilterIDs: []string{"*string:~*req.Account:1004"},
			Balances: map[string]*utils.APIBalance{
				"ConcreteBalance1": &utils.APIBalance{
					ID:      "ConcreteBalance1",
					Weights: ";20",
					Type:    utils.MetaConcrete,
					Units:   21,
					CostIncrements: []*utils.APICostIncrement{
						&utils.APICostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*data"},
							Increment:    utils.Float64Pointer(1),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(1),
						},
					},
				},
				"ConcreteBalance2": &utils.APIBalance{
					ID:      "ConcreteBalance2",
					Weights: ";10",
					Type:    utils.MetaConcrete,
					Units:   20,
					CostIncrements: []*utils.APICostIncrement{
						&utils.APICostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*data"},
							Increment:    utils.Float64Pointer(1),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(1),
						},
					},
				},
				"AbstractBalance1": &utils.APIBalance{
					ID:      "AbstractBalance1",
					Weights: ";5",
					Type:    utils.MetaAbstract,
					Units:   20,
					CostIncrements: []*utils.APICostIncrement{
						&utils.APICostIncrement{
							FilterIDs:    []string{"*string:~*req.ToR:*data"},
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(0),
							RecurrentFee: utils.Float64Pointer(1),
						},
					},
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	var reply string
	if err := acntSRPC.Call(utils.APIerSv1SetAccountProfile, apiAccPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	exp, err := apiAccPrf.AsAccountProfile()
	if err != nil {
		t.Error(err)
	}
	var result *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1004"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}

	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1MaxConcretes,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxConcretes",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.ToR:          utils.MetaData,
				utils.Usage:        "50ns",
			},
		}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Concretes == nil || *eEc.Concretes != 41 {
		t.Errorf("received usage: %v", *eEc.Concretes)
	}

	//make sure we did not Debit from our Account
	exp, err = apiAccPrf.AsAccountProfile()
	if err != nil {
		t.Error(err)
	}
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1004"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}
func testAccountSv1DebitConcretes(t *testing.T) {
	var eEc *utils.ExtEventCharges
	//Now we know the usage, we will debit it from account
	if err := acntSRPC.Call(utils.AccountSv1DebitConcretes,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxConcretes",
			Event: map[string]interface{}{
				utils.AccountField: "1004",
				utils.ToR:          utils.MetaData,
				utils.Usage:        "50ns",
			},
		}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Concretes == nil || *eEc.Concretes != 41 {
		t.Errorf("received usage: %v", *eEc.Concretes)
	}

	exp := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": &utils.Balance{
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
			"ConcreteBalance2": &utils.Balance{
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
			"AbstractBalance1": &utils.Balance{
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(20, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	var result *utils.AccountProfile
	//As we debit, our Account balances are changed now
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1004"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, result) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(exp), utils.ToJSON(result))
	}
}

func testAccountSv1ActionSetBalance(t *testing.T) {
	expectedSetBalance := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": &utils.Balance{
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"fltr1", "fltr2"},
						Weight:    20,
					},
					{
						FilterIDs: []string{"fltr1"},
						Weight:    30,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
			"ConcreteBalance2": &utils.Balance{
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
			"AbstractBalance1": &utils.Balance{
				ID: "AbstractBalance1",
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Type:  utils.MetaAbstract,
				Units: &utils.Decimal{decimal.New(120, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(int64(time.Second), 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{"TH_ID1"},
	}
	var reply string
	if err := acntSRPC.Call(utils.AccountSv1ActionSetBalance, &utils.ArgsActSetBalance{
		Tenant: "cgrates.org", AccountID: "1004",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*account.ThresholdIDs",
				Value: "TH_ID1",
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := acntSRPC.Call(utils.AccountSv1ActionSetBalance, &utils.ArgsActSetBalance{
		Tenant: "cgrates.org", AccountID: "1004",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.AbstractBalance1.Units",
				Value: "120",
			},
		},
		Reset: true,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := acntSRPC.Call(utils.AccountSv1ActionSetBalance, &utils.ArgsActSetBalance{
		Tenant: "cgrates.org", AccountID: "1004",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.ConcreteBalance1.Weights",
				Value: "fltr1&fltr2;20;fltr1;30",
			},
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	var result *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1004"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedSetBalance, result) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(expectedSetBalance), utils.ToJSON(result))
	}
}

func testAccountSv1ActionRemoveBalance(t *testing.T) {
	var reply string
	if err := acntSRPC.Call(utils.AccountSv1ActionRemoveBalance, &utils.ArgsActRemoveBalances{
		Tenant: "cgrates.org", AccountID: "1004",
		BalanceIDs: []string{"AbstractBalance1"},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	expectedSetBalance := &utils.AccountProfile{
		Tenant:    "cgrates.org",
		ID:        "1004",
		FilterIDs: []string{"*string:~*req.Account:1004"},
		Balances: map[string]*utils.Balance{
			"ConcreteBalance1": &utils.Balance{
				ID: "ConcreteBalance1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: []string{"fltr1", "fltr2"},
						Weight:    20,
					},
					{
						FilterIDs: []string{"fltr1"},
						Weight:    30,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
			"ConcreteBalance2": &utils.Balance{
				ID: "ConcreteBalance2",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Type:  utils.MetaConcrete,
				Units: &utils.Decimal{decimal.New(0, 0)},
				CostIncrements: []*utils.CostIncrement{
					&utils.CostIncrement{
						FilterIDs:    []string{"*string:~*req.ToR:*data"},
						Increment:    &utils.Decimal{decimal.New(1, 0)},
						FixedFee:     &utils.Decimal{decimal.New(0, 0)},
						RecurrentFee: &utils.Decimal{decimal.New(1, 0)},
					},
				},
			},
		},
		ThresholdIDs: []string{"TH_ID1"},
	}

	var result *utils.AccountProfile
	if err := acntSRPC.Call(utils.APIerSv1GetAccountProfile, &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1004"}}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedSetBalance, result) {
		t.Errorf("Expected %+v\n, received %+v", utils.ToJSON(expectedSetBalance), utils.ToJSON(result))
	}
}

func testAccountSv1KillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
