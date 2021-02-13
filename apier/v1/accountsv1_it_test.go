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
	acntSDataDir   = "/usr/share/cgrates"
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
		testAccountSv1AccountProfileForEvent,
		testAccountSv1MaxUsage,
		testAccountSv1DebitUsage,
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
	acntSCfgPath = path.Join(acntSDataDir, "conf", "samples", acntSConfigDIR)
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

func testAccountSv1AccountProfileForEvent(t *testing.T) {
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

func testAccountSv1MaxUsage(t *testing.T) {
	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1MaxUsage,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxUsage",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        "15m",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Usage == nil || *eEc.Usage != 800000000000.0 { // 500s from first monetary + 300s from last monetary
		t.Errorf("received usage: %v", *eEc.Usage)
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
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eAcnt, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(eAcnt), utils.ToJSON(reply))
	}
}

func testAccountSv1DebitUsage(t *testing.T) {
	var eEc *utils.ExtEventCharges
	if err := acntSRPC.Call(utils.AccountSv1DebitUsage,
		&utils.ArgsAccountsForEvent{CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testAccountSv1MaxUsage",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Usage:        "15m",
			}}}, &eEc); err != nil {
		t.Error(err)
	} else if eEc.Usage == nil || *eEc.Usage != 800000000000.0 { // 500s from first monetary + 300s from last monetary
		t.Fatalf("received usage: %v", *eEc.Usage)
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
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}}, &reply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(eAcnt, reply) {
		t.Errorf("Expecting : %+v \n received: %+v", utils.ToJSON(eAcnt), utils.ToJSON(reply))
	}
}
