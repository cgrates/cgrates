//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func setUpConfigs(t testing.TB, opts string) (ng engine.TestEngine) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	content := fmt.Sprintf(`{
"logger": {
	"level": 7
},
"sessions": {
	"enabled": true,
	"conns": {
		"*thresholds": [{"connIDs": ["*localhost"]}]
	}
},
"actions": {
	"enabled": true,
},
"thresholds": {
	"enabled": true,
	"conns": {
		"*actions": [{"connIDs": ["*localhost"]}]
	},
},
"accounts": {
	"enabled": true,
	"conns": {
	  "*thresholds": [{"connIDs": ["*localhost"]}]
	},
	"opts": %s
  },
"admins": {
	"enabled": true
}}`, opts)

	ng = engine.TestEngine{
		ConfigJSON: content,
		DBCfg:      engine.InternalDBCfg,
		Encoding:   utils.MetaJSON,
		TpFiles: map[string]string{
			utils.ActionsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Schedule,TargetType,TargetIDs,ActionID,ActionFilterIDs,ActionTTL,ActionType,ActionOpts,ActionPath,ActionValue
cgrates.org,*log,,,,,,,,,,,,,`,
		},
	}
	return ng
}

func setThresholdProfile(t testing.TB, client *birpc.Client, thrPrfID string) (err error) {
	tPrfl := &utils.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &utils.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               thrPrfID,
			MaxHits:          -1,
			MinHits:          1,
			Weights:          utils.DynamicWeights{{Weight: 10}},
			ActionProfileIDs: []string{"*log"},
		},
	}
	var reply string
	if err = client.Call(context.Background(), utils.AdminSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		return
	}
	return
}

func setAccount(t testing.TB, client *birpc.Client, tenant, accID string, thrIDs []string) (err error) {
	acc := &utils.Account{
		Tenant:    tenant,
		ID:        accID,
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Balances: map[string]*utils.Balance{
			"AbstractBalance": {
				ID:      "AbstractBalance",
				Weights: utils.DynamicWeights{{Weight: 30}},
				Type:    utils.MetaAbstract,
				Units:   utils.NewDecimal(10, 0),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment:    utils.NewDecimal(1, 0),
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
			"ConcreteBalance": {
				ID: "ConcreteBalance",
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
						FixedFee:     utils.NewDecimal(0, 0),
						RecurrentFee: utils.NewDecimal(0, 0),
					},
				},
			},
		},
		ThresholdIDs: thrIDs,
	}
	var reply string
	if err = client.Call(context.Background(), utils.AdminSv1SetAccount, acc, &reply); err != nil {
		return
	}
	return
}

func getAccount(t testing.TB, client *birpc.Client, id, tenant string, opts map[string]any) (acc *utils.Account, err error) {
	var result utils.Account
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: tenant,
			ID:     id,
		},
		APIOpts: opts,
	}
	if err = client.Call(context.Background(), utils.AccountSv1GetAccount,
		args, &result); err != nil {
		return
	}
	return &result, nil
}

func expectedThresholdHits(t testing.TB, client *birpc.Client, id string, hits int) (err error) {
	var thd utils.Threshold
	if err = client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: id}}, &thd); err != nil {
		t.Fatal(err)
	} else if thd.Hits != hits {
		t.Errorf("expected %d hits, received %d", hits, thd.Hits)
	}
	return
}

func TestAccountSThresholdSConfigs(t *testing.T) {
	opts := `{
		"*thresholdIDs": [
			{
				"tenant": "cgrates.org",
				"values": ["THR_ACC"]
			}
		],
		"*thresholds": [
         {
			"tenant": "cgrates.org",
             "value": "true"
         }
        ]
	}`
	ng := setUpConfigs(t, opts)
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates.org", "acc1", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}

	t.Run("V1GetAccount", func(t *testing.T) {
		var reply utils.Account
		args := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "acc1",
			},
		}
		if err = client.Call(context.Background(), utils.AccountSv1GetAccount,
			args, &reply); err != nil {
			t.Fatal(err)
		}
		expAcc := &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "acc1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance": {
					ID:      "AbstractBalance",
					Weights: utils.DynamicWeights{{Weight: 30}},
					Type:    utils.MetaAbstract,
					Units:   utils.NewDecimal(10, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(1, 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance": {
					ID: "ConcreteBalance",
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
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
			},
		}
		if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON(expAcc)) {
			t.Errorf("Expected %v, \nreceived %v", utils.ToJSON(expAcc), utils.ToJSON(reply))
		}
		expectedThresholdHits(t, client, "THR_ACC", 1)
	})

	t.Run("V1AccountsForEvent", func(t *testing.T) {
		var reply []*utils.Account
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAccountsProfileIDs: "acc1",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if len(reply) != 1 {
			t.Errorf("Expected 1 account, received %v", len(reply))
		}
		expectedThresholdHits(t, client, "THR_ACC", 2)
	})

	t.Run("V1MaxConcretes", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAccountsProfileIDs: "acc1",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1MaxConcretes, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Concretes.Compare(utils.NewDecimal(100, 0)) != 0 {
			t.Errorf("Expected 100, received %v", reply.Concretes)
		}
		expectedThresholdHits(t, client, "THR_ACC", 3)
	})

	t.Run("V1MaxAbstracts", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAccountsProfileIDs: "acc1",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1MaxAbstracts, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Abstracts.Compare(utils.NewDecimal(10, 0)) != 0 {
			t.Errorf("Expected 10, received %v", reply.Abstracts)
		}
		expectedThresholdHits(t, client, "THR_ACC", 4)
	})

	t.Run("V1DebitAbstracts", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaUsage:              3,
				utils.OptsAccountsProfileIDs: "acc1",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1DebitAbstracts, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Abstracts.Compare(utils.NewDecimal(3, 0)) != 0 {
			t.Errorf("Expected to use 3 units, used %v", reply.Abstracts)
		}
		expectedThresholdHits(t, client, "THR_ACC", 5)
	})

	t.Run("V1DebitConcretes", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaUsage: 20,
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1DebitConcretes, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Concretes.Compare(utils.NewDecimal(20, 0)) != 0 {
			t.Errorf("Expected to use 20 units, used %v", reply.Concretes)
		}
		expectedThresholdHits(t, client, "THR_ACC", 6)
	})

	t.Run("V1RefundCharges", func(t *testing.T) {
		acc, err := getAccount(t, client, "acc1", "cgrates.org", nil)
		if err != nil {
			t.Fatal(err)
		}
		var reply *string
		arg := &utils.APIEventCharges{
			EventCharges: &utils.EventCharges{
				Abstracts: utils.NewDecimal(3, 0),
				Concretes: utils.NewDecimal(20, 0),
				Charges: []*utils.ChargeEntry{
					{
						ChargingID:     "CHARGING1",
						CompressFactor: 1,
					},
					{
						ChargingID:     "CHARGING2",
						CompressFactor: 1,
					},
				},
				Accounting: map[string]*utils.AccountCharge{
					"CHARGING1": {
						AccountID: "acc1",
						BalanceID: "AbstractBalance",
						Units:     utils.NewDecimal(3, 0),
					},
					"CHARGING2": {
						AccountID: "acc1",
						BalanceID: "ConcreteBalance",
						Units:     utils.NewDecimal(20, 0),
					},
				},
				Accounts: map[string]*utils.Account{
					"acc1": acc,
				},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1RefundCharges, arg, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}

		accAfterRefund, err := getAccount(t, client, "acc1", "cgrates.org", nil)
		if err != nil {
			t.Fatal(err)
		}
		if accAfterRefund.Balances["AbstractBalance"].Units.Compare(utils.NewDecimal(10, 0)) != 0 {
			t.Errorf("Expected 10, received %v", accAfterRefund.Balances["AbstractBalance"].Units)
		}
		if accAfterRefund.Balances["ConcreteBalance"].Units.Compare(utils.NewDecimal(100, 0)) != 0 {
			t.Errorf("Expected 100, received %v", accAfterRefund.Balances["ConcreteBalance"].Units)
		}
		expectedThresholdHits(t, client, "THR_ACC", 9)
	})

	t.Run("V1ActionSetBalance", func(t *testing.T) {
		var reply *string
		args := &utils.ArgsActSetBalance{
			Tenant:    "cgrates.org",
			AccountID: "acc1",
			Diktats: []*utils.BalDiktat{
				{
					Path:  "*balance.AbstractBalance2.Units",
					Value: "30",
				},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1ActionSetBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}

		acc, err := getAccount(t, client, "acc1", "cgrates.org", nil)
		if err != nil {
			t.Fatal(err)
		}
		if acc.Balances["AbstractBalance2"].Units.Compare(utils.NewDecimal(30, 0)) != 0 {
			t.Errorf("Expected 30, received %v", acc.Balances["AbstractBalance2"].Units)
		}
		expectedThresholdHits(t, client, "THR_ACC", 11)
	})

	t.Run("V1ActionRemoveBalance", func(t *testing.T) {
		var reply *string
		args := &utils.ArgsActRemoveBalances{
			Tenant:     "cgrates.org",
			AccountID:  "acc1",
			BalanceIDs: []string{"AbstractBalance2"},
		}
		if err := client.Call(context.Background(), utils.AccountSv1ActionRemoveBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}

		acc, err := getAccount(t, client, "acc1", "cgrates.org", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(len(acc.Balances), 2) {
			t.Errorf("Expected 2 balances left, received %v", len(acc.Balances))
		}
		expectedThresholdHits(t, client, "THR_ACC", 13)
	})
}

func TestAccountSThresholdSAPIopts(t *testing.T) {
	ng := setUpConfigs(t, "{}")
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates.org", "acc1", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	t.Run("V1GetAccount", func(t *testing.T) {
		var reply utils.Account
		args := &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "acc1",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds: true,
			},
		}
		if err = client.Call(context.Background(), utils.AccountSv1GetAccount,
			args, &reply); err != nil {
			t.Fatal(err)
		}
		expAcc := &utils.Account{
			Tenant:    "cgrates.org",
			ID:        "acc1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Balances: map[string]*utils.Balance{
				"AbstractBalance": {
					ID:      "AbstractBalance",
					Weights: utils.DynamicWeights{{Weight: 30}},
					Type:    utils.MetaAbstract,
					Units:   utils.NewDecimal(10, 0),
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(1, 0),
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
				"ConcreteBalance": {
					ID: "ConcreteBalance",
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
							FixedFee:     utils.NewDecimal(0, 0),
							RecurrentFee: utils.NewDecimal(0, 0),
						},
					},
				},
			},
		}
		if !reflect.DeepEqual(utils.ToJSON(reply), utils.ToJSON(expAcc)) {
			t.Errorf("Expected %v, \nreceived %v", utils.ToJSON(expAcc), utils.ToJSON(reply))
		}
		expectedThresholdHits(t, client, "THR_ACC", 1)
	})

	t.Run("V1AccountsForEvent", func(t *testing.T) {
		var reply []*utils.Account
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if len(reply) != 1 {
			t.Errorf("Expected 1 account, received %v", len(reply))
		}
		expectedThresholdHits(t, client, "THR_ACC", 2)
	})

	t.Run("V1MaxConcretes", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1MaxConcretes, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Concretes.Compare(utils.NewDecimal(100, 0)) != 0 {
			t.Errorf("Expected 100, received %v", reply.Concretes)
		}
		expectedThresholdHits(t, client, "THR_ACC", 3)
	})

	t.Run("V1MaxAbstracts", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1MaxAbstracts, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Abstracts.Compare(utils.NewDecimal(10, 0)) != 0 {
			t.Errorf("Expected 10, received %s", reply.Abstracts)
		}
		expectedThresholdHits(t, client, "THR_ACC", 4)
	})

	t.Run("V1DebitAbstracts", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
				utils.MetaUsage:        3,
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1DebitAbstracts, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Abstracts.Compare(utils.NewDecimal(3, 0)) != 0 {
			t.Errorf("Expected to use 3 units, used %v", reply.Abstracts)
		}
		expectedThresholdHits(t, client, "THR_ACC", 5)
	})

	t.Run("V1DebitConcretes", func(t *testing.T) {
		var reply *utils.EventCharges
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
				utils.MetaUsage:        20,
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1DebitConcretes, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if reply.Concretes.Compare(utils.NewDecimal(20, 0)) != 0 {
			t.Errorf("Expected to use 20 units, used %v", reply.Concretes)
		}
		expectedThresholdHits(t, client, "THR_ACC", 6)
	})

	t.Run("V1RefundCharges", func(t *testing.T) {
		acc, err := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true})
		if err != nil {
			t.Fatal(err)
		}
		var reply *string
		arg := &utils.APIEventCharges{
			EventCharges: &utils.EventCharges{
				Abstracts: utils.NewDecimal(3, 0),
				Concretes: utils.NewDecimal(20, 0),
				Charges: []*utils.ChargeEntry{
					{
						ChargingID:     "CHARGING1",
						CompressFactor: 1,
					},
					{
						ChargingID:     "CHARGING2",
						CompressFactor: 1,
					},
				},
				Accounting: map[string]*utils.AccountCharge{
					"CHARGING1": {
						AccountID: "acc1",
						BalanceID: "AbstractBalance",
						Units:     utils.NewDecimal(3, 0),
					},
					"CHARGING2": {
						AccountID: "acc1",
						BalanceID: "ConcreteBalance",
						Units:     utils.NewDecimal(20, 0),
					},
				},
				Accounts: map[string]*utils.Account{
					"acc1": acc,
				},
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1RefundCharges, arg, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}
		expectedThresholdHits(t, client, "THR_ACC", 8)
	})

	t.Run("V1ActionSetBalance", func(t *testing.T) {
		var reply *string
		args := &utils.ArgsActSetBalance{
			Tenant:    "cgrates.org",
			AccountID: "acc1",
			Diktats: []*utils.BalDiktat{
				{
					Path:  "*balance.AbstractBalance2.Units",
					Value: "30",
				},
			},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1ActionSetBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}

		acc, err := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true})
		if err != nil {
			t.Fatal(err)
		}
		if acc.Balances["AbstractBalance2"].Units.Compare(utils.NewDecimal(30, 0)) != 0 {
			t.Errorf("Expected 30, received %v", acc.Balances["AbstractBalance2"].Units)
		}
		expectedThresholdHits(t, client, "THR_ACC", 10)
	})

	t.Run("V1ActionRemoveBalance", func(t *testing.T) {
		var reply *string
		args := &utils.ArgsActRemoveBalances{
			Tenant:     "cgrates.org",
			AccountID:  "acc1",
			BalanceIDs: []string{"AbstractBalance2"},
			APIOpts: map[string]any{
				utils.MetaThresholds:   true,
				utils.MetaThresholdIDs: []string{"THR_ACC"},
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1ActionRemoveBalance, args, &reply); err != nil {
			t.Fatal(err)
		}
		if *reply != utils.OK {
			t.Errorf("Expected ok, received %s", *reply)
		}
		acc, err := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(len(acc.Balances), 2) {
			t.Errorf("Expected 2 balances left, received %v", len(acc.Balances))
		}
		expectedThresholdHits(t, client, "THR_ACC", 12)
	})
}

func TestAccountSThresholdSAPIoptsANDConfigs(t *testing.T) {
	opts := `{
		"*thresholdIDs": [
			{
				"tenant": "cgrates.org",
				"values": ["THR_CONF"]
			}
		],
		"*thresholds": [
         {
			"tenant": "cgrates.org",
             "value": "true"
         }
        ]
	},`
	ng := setUpConfigs(t, opts)
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_CONF")
	if err != nil {
		t.Fatal(err)
	}
	err2 := setThresholdProfile(t, client, "THR_ACC")
	if err2 != nil {
		t.Fatal(err2)
	}

	t.Run("Expected to use the threshold profile that is on the account", func(t *testing.T) {
		err := setAccount(t, client, "cgrates.org", "acc1", []string{"THR_ACC"})
		if err != nil {
			t.Fatal(err)
		}
		_, errAcc := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true})
		if errAcc != nil {
			t.Fatal(errAcc)
		}
		expectedThresholdHits(t, client, "THR_ACC", 1)
		expectedThresholdHits(t, client, "THR_CONF", 0)
	})

	t.Run("Expected to use the threshold profile that is on configs", func(t *testing.T) {
		err := setAccount(t, client, "cgrates.org", "acc1", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, errAcc := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true})
		if errAcc != nil {
			t.Fatal(errAcc)
		}
		expectedThresholdHits(t, client, "THR_ACC", 1)
		expectedThresholdHits(t, client, "THR_CONF", 1)
	})

	t.Run("Expected the hits to stay uneffected when MetaThresholds is false", func(t *testing.T) {
		err := setAccount(t, client, "cgrates.org", "acc1", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, errAcc := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: false})
		if errAcc != nil {
			t.Fatal(errAcc)
		}
		expectedThresholdHits(t, client, "THR_ACC", 1)
		expectedThresholdHits(t, client, "THR_CONF", 1)
	})

	ng.Stop(t)
	ng = setUpConfigs(t, "{}")
	client, cfg = ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)
	err3 := setThresholdProfile(t, client, "THR_API")
	if err3 != nil {
		t.Fatal(err3)
	}
	t.Run("Expected to use the threshold profile that is on APIopts", func(t *testing.T) {
		err := setAccount(t, client, "cgrates.org", "acc1", nil)
		if err != nil {
			t.Fatal(err)
		}
		_, errAcc := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: true, utils.MetaThresholdIDs: []string{"THR_API"}})
		if errAcc != nil {
			t.Fatal(errAcc)
		}
		expectedThresholdHits(t, client, "THR_API", 1)
	})
}

func TestAccountSThresholdSConfigsThresholdsFalseWithThresholdIDs(t *testing.T) {
	opts := `{
		"*thresholdIDs": [
			{
				"tenant": "cgrates.org",
				"values": ["THR_ACC"]
			}
		],
		"*thresholds": [
         {
			"tenant": "cgrates.org",
             "value": "false"
         }
        ]
	}`
	ng := setUpConfigs(t, opts)
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates.org", "acc1", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	var result utils.Account
	tnID := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "acc1",
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1GetAccount,
		tnID, &result); err != nil {
		t.Fatal(err)
	}
	expectedThresholdHits(t, client, "THR_ACC", 0)

	var replyAcc []*utils.Account
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs: "acc1",
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &replyAcc); err != nil {
		t.Fatal(err)
	}

	var replyEv *utils.EventCharges
	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs: "acc1",
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1MaxConcretes, ev, &replyEv); err != nil {
		t.Fatal(err)
	}

	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs: "acc1",
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1MaxAbstracts, ev, &replyEv); err != nil {
		t.Fatal(err)
	}

	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:              3,
			utils.OptsAccountsProfileIDs: "acc1",
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1DebitAbstracts, ev, &replyEv); err != nil {
		t.Fatal(err)
	}

	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.MetaUsage: 20,
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1DebitConcretes, ev, &replyEv); err != nil {
		t.Fatal(err)
	}

	acc, err := getAccount(t, client, "acc1", "cgrates.org", nil)
	if err != nil {
		t.Fatal(err)
	}

	var reply *string
	arg := &utils.APIEventCharges{
		EventCharges: &utils.EventCharges{
			Abstracts: utils.NewDecimal(3, 0),
			Concretes: utils.NewDecimal(20, 0),
			Charges: []*utils.ChargeEntry{
				{
					ChargingID:     "CHARGING1",
					CompressFactor: 1,
				},
				{
					ChargingID:     "CHARGING2",
					CompressFactor: 1,
				},
			},
			Accounting: map[string]*utils.AccountCharge{
				"CHARGING1": {
					AccountID: "acc1",
					BalanceID: "AbstractBalance",
					Units:     utils.NewDecimal(3, 0),
				},
				"CHARGING2": {
					AccountID: "acc1",
					BalanceID: "ConcreteBalance",
					Units:     utils.NewDecimal(20, 0),
				},
			},
			Accounts: map[string]*utils.Account{
				"acc1": acc,
			},
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1RefundCharges, arg, &reply); err != nil {
		t.Fatal(err)
	}

	args := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "acc1",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.AbstractBalance2.Units",
				Value: "30",
			},
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1ActionSetBalance, args, &reply); err != nil {
		t.Fatal(err)
	}

	rmvBal := &utils.ArgsActRemoveBalances{
		Tenant:     "cgrates.org",
		AccountID:  "acc1",
		BalanceIDs: []string{"AbstractBalance2"},
	}
	if err := client.Call(context.Background(), utils.AccountSv1ActionRemoveBalance, rmvBal, &reply); err != nil {
		t.Fatal(err)
	}
	expectedThresholdHits(t, client, "THR_ACC", 0)
}

func TestAccountSThresholdSAPIoptsThresholdsFalseWithThresholdIDs(t *testing.T) {
	ng := setUpConfigs(t, "{}")
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates.org", "acc1", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	_, errGet := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: false})
	if errGet != nil {
		t.Fatal(errGet)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	var replyAcc []*utils.Account
	var replyEV *utils.EventCharges
	var replyStr *string

	ev.APIOpts = map[string]any{
		utils.MetaThresholds:   false,
		utils.MetaThresholdIDs: []string{"THR_ACC"},
	}
	if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &replyAcc); err != nil {
		t.Fatal(err)
	}

	ev.APIOpts = map[string]any{
		utils.MetaThresholds:   false,
		utils.MetaThresholdIDs: []string{"THR_ACC"},
	}
	if err := client.Call(context.Background(), utils.AccountSv1MaxConcretes, ev, &replyEV); err != nil {
		t.Fatal(err)
	}

	ev.APIOpts = map[string]any{
		utils.MetaThresholds:   false,
		utils.MetaThresholdIDs: []string{"THR_ACC"},
	}
	if err := client.Call(context.Background(), utils.AccountSv1MaxAbstracts, ev, &replyEV); err != nil {
		t.Fatal(err)
	}

	ev.APIOpts = map[string]any{
		utils.MetaThresholds:   false,
		utils.MetaThresholdIDs: []string{"THR_ACC"},
		utils.MetaUsage:        3,
	}
	if err := client.Call(context.Background(), utils.AccountSv1DebitAbstracts, ev, &replyEV); err != nil {
		t.Fatal(err)
	}

	ev.APIOpts = map[string]any{
		utils.MetaThresholds:   false,
		utils.MetaThresholdIDs: []string{"THR_ACC"},
		utils.MetaUsage:        20,
	}
	if err := client.Call(context.Background(), utils.AccountSv1DebitConcretes, ev, &replyEV); err != nil {
		t.Fatal(err)
	}

	acc, err := getAccount(t, client, "acc1", "cgrates.org", map[string]any{utils.MetaThresholds: false})
	if err != nil {
		t.Fatal(err)
	}
	arg := &utils.APIEventCharges{
		EventCharges: &utils.EventCharges{
			Abstracts: utils.NewDecimal(3, 0),
			Concretes: utils.NewDecimal(20, 0),
			Charges: []*utils.ChargeEntry{
				{
					ChargingID:     "CHARGING1",
					CompressFactor: 1,
				},
				{
					ChargingID:     "CHARGING2",
					CompressFactor: 1,
				},
			},
			Accounting: map[string]*utils.AccountCharge{
				"CHARGING1": {
					AccountID: "acc1",
					BalanceID: "AbstractBalance",
					Units:     utils.NewDecimal(3, 0),
				},
				"CHARGING2": {
					AccountID: "acc1",
					BalanceID: "ConcreteBalance",
					Units:     utils.NewDecimal(20, 0),
				},
			},
			Accounts: map[string]*utils.Account{
				"acc1": acc,
			},
		},
		APIOpts: map[string]any{
			utils.MetaThresholds:   false,
			utils.MetaThresholdIDs: []string{"THR_ACC"},
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1RefundCharges, arg, &replyStr); err != nil {
		t.Fatal(err)
	}

	args := &utils.ArgsActSetBalance{
		Tenant:    "cgrates.org",
		AccountID: "acc1",
		Diktats: []*utils.BalDiktat{
			{
				Path:  "*balance.AbstractBalance2.Units",
				Value: "30",
			},
		},
		APIOpts: map[string]any{
			utils.MetaThresholds:   false,
			utils.MetaThresholdIDs: []string{"THR_ACC"},
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1ActionSetBalance, args, &replyStr); err != nil {
		t.Fatal(err)
	}

	rmvBal := &utils.ArgsActRemoveBalances{
		Tenant:     "cgrates.org",
		AccountID:  "acc1",
		BalanceIDs: []string{"AbstractBalance2"},
		APIOpts: map[string]any{
			utils.MetaThresholds:   false,
			utils.MetaThresholdIDs: []string{"THR_ACC"},
		},
	}
	if err := client.Call(context.Background(), utils.AccountSv1ActionRemoveBalance, rmvBal, &replyStr); err != nil {
		t.Fatal(err)
	}
	expectedThresholdHits(t, client, "THR_ACC", 0)
}

func TestAccountSThresholdSWithSpecificAccountThresholdIDs(t *testing.T) {
	opts := `{
		"*thresholds": [
         {
			"tenant": "cgrates.org",
             "value": "true"
         }
        ]
	}`
	ng := setUpConfigs(t, opts)
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates.org", "acc1", []string{"THR_ACC"})
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	_, errGet := getAccount(t, client, "acc1", "cgrates.org", nil)
	if errGet != nil {
		t.Fatal(errGet)
	}
	t.Run("V1AccountsForEvent", func(t *testing.T) {
		var reply []*utils.Account
		ev := &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAccountsProfileIDs: "acc1",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &reply); err != nil {
			t.Fatal(err)
		}
		if len(reply) != 1 {
			t.Errorf("Expected 1 account, received %v", len(reply))
		}
		expectedThresholdHits(t, client, "THR_ACC", 2)
	})
}

func TestAccountSThresholdSNilThresholIDsWithTenant(t *testing.T) {
	opts := `{
		"*thresholds": [
         {
			"tenant": "cgrates.org",
             "value": "true"
         }
        ]
	}`
	ng := setUpConfigs(t, opts)
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)

	err := setThresholdProfile(t, client, "THR_ACC")
	if err != nil {
		t.Fatal(err)
	}
	errAcc := setAccount(t, client, "cgrates", "acc2", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	_, errGet := getAccount(t, client, "acc2", "cgrates", nil)
	if errGet != nil {
		t.Fatal(errGet)
	}

	t.Run("V1AccountsForEvent", func(t *testing.T) {
		var reply []*utils.Account
		ev := &utils.CGREvent{
			Tenant: "cgrates",
			ID:     "ev1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsAccountsProfileIDs: "acc2",
			},
		}
		if err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &reply); err != nil {
			t.Fatal(err)
		}
		expectedThresholdHits(t, client, "THR_ACC", 0)
	})
}

func TestAccountSThresholdSNilThresholIdsNilTenant(t *testing.T) {
	opts := `{
		"*thresholds": [
         {
             "value": "true"
         }
        ]
	}`

	buf := &bytes.Buffer{}
	ng := setUpConfigs(t, opts)
	ng.LogBuffer = buf
	client, cfg := ng.Run(t)
	engine.WaitForServiceStart(t, client, utils.SessionS, cfg.GeneralCfg().ConnectTimeout)
	errTh := setThresholdProfile(t, client, "THR_ACC")
	if errTh != nil {
		t.Fatal(errTh)
	}
	errAcc := setAccount(t, client, "cgrates", "acc2", nil)
	if errAcc != nil {
		t.Fatal(errAcc)
	}
	var reply []*utils.Account
	ev := &utils.CGREvent{
		Tenant: "cgrates",
		ID:     "ev1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsAccountsProfileIDs: "acc2",
		},
	}

	err := client.Call(context.Background(), utils.AccountSv1AccountsForEvent, ev, &reply)
	if err == nil || err.Error() != "PARTIALLY_EXECUTED" {
		t.Error(err)
	}

	expLog := "<AccountS> error: NOT_FOUND processing account:"
	waitForLog(t, buf, expLog, 2*time.Second)
	if got := strings.Count(buf.String(), expLog); got != 1 {
		t.Fatalf("expected 1 LOG record, got %d", got)
	}
	buf = &bytes.Buffer{}
	expectedThresholdHits(t, client, "THR_ACC", 0)
}
