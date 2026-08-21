//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventRates(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
    	"*rates": [{"connIDs": ["*localhost"]}]
    },
},
"rates": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationStart,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP_SIMPLE,,;10,,,,RT_SIMPLE,*string:~*req.Destination:1002,"* * * * *",;10,false,0s,0,1,1m,1m`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
		// LogBuffer: new(bytes.Buffer),
	}

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

	client, _ := ng.Run(t)
	time.Sleep(500 * time.Millisecond)

	t.Run("noFlags", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noFlags",
				APIOpts: map[string]any{
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed without rates flag: %v", err)
		}
		if len(rply.RatesCost) > 0 {
			t.Errorf("RatesCost should be empty without *rates flag, got: %v", rply.RatesCost)
		}
	})

	t.Run("ratesFlag", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ratesFlag",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaUsage:    1 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed with *rates flag: %v", err)
		}
		if rply.RatesCost == nil {
			t.Fatal("RatesCost should not be nil with *rates flag")
		}
		cost, exists := rply.RatesCost[utils.MetaPrimary]
		if !exists {
			t.Fatalf("no RatesCost entry for *primary runID, got: %v", rply.RatesCost)
		}
		const wantCost = 1.0
		if cost != wantCost {
			t.Errorf("RatesCost[*primary] = %g, want %g", cost, wantCost)
		}
	})

	t.Run("ratesSecondInterval", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ratesSecondInterval",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaUsage:    2 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if rply.RatesCost == nil {
			t.Fatal("RatesCost should not be nil")
		}
		cost, exists := rply.RatesCost[utils.MetaPrimary]
		if !exists {
			t.Fatalf("no RatesCost entry for *primary runID, got: %v", rply.RatesCost)
		}
		const wantCost = 2.0
		if cost != wantCost {
			t.Errorf("RatesCost[*primary] = %g, want %g", cost, wantCost)
		}
	})

	t.Run("ratesDisabled", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "ratesDisabled",
				APIOpts: map[string]any{
					utils.MetaRates:    false,
					utils.MetaUsage:    2*time.Minute + 30*time.Second,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err != nil {
			t.Fatalf("ProcessEvent failed: %v", err)
		}
		if len(rply.RatesCost) > 0 {
			t.Errorf("RatesCost should be empty when *rates=false, got: %v", rply.RatesCost)
		}
	})

	t.Run("noMatchingRate", func(t *testing.T) {
		var rply V1ProcessEventReply

		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "noMatchingRate",
				APIOpts: map[string]any{
					utils.MetaRates:    true,
					utils.MetaUsage:    1 * time.Minute,
					utils.MetaOriginID: "OriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "9999",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply,
		); err == nil {
			t.Fatal("expected error, got none")
		}
	})
}

func TestSessionSv1ProcessEventInterimAccountDebit(t *testing.T) {

	ng := engine.TestEngine{
		ConfigJSON: `{
"logger": {"level": 7},
"sessions": {
    "enabled": true,
    "conns": {
        "*chargers": [{"connIDs": ["*localhost"]}],
        "*accounts": [{"connIDs": ["*localhost"]}]
    }
},
"chargers": {
    "enabled": true
},
"accounts": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"BAL1": {
						ID:      "BAL1",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(300 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	originID := "OriginID"

	t.Run("SessionStart", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "initiate1",
				APIOpts: map[string]any{
					utils.MetaSession:  true,
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaInitiate: true,
					utils.MetaUsage:    90 * time.Second,
					utils.MetaOriginID: originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(initiate): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountSUsage missing *default")
		}
		if usage != 90*time.Second {
			t.Errorf("AccountSUsage[*default] = %v, want 1m30s", usage)
		}
	})

	t.Run("SessionUpdate1", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "update1",
				APIOpts: map[string]any{
					utils.MetaSession:  true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUpdate:   true,
					utils.MetaUsage:    150 * time.Second,
					utils.MetaOriginID: originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(update): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountSUsage missing *default")
		}
		if usage != 150*time.Second {
			t.Errorf("AccountSUsage[*default] = %v, want 2m30s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else if want := utils.NewDecimalFromFloat64(float64(60 * time.Second)); acnt.Balances["BAL1"].Units.Compare(want) != 0 {
		t.Errorf("expected 60s remaining after interim debits, got: %+v", acnt.Balances["BAL1"].Units)
	}
}

func TestSessionSv1ProcessEventTerminate(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
        "*chargers": [{"connIDs": ["*localhost"]}],
        "*accounts": [{"connIDs": ["*localhost"]}]
    }
},
"chargers": {
    "enabled": true
},
"accounts": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"BAL1": {
						ID:      "BAL1",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(300 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	originID := "OriginIDTerminate"

	t.Run("SessionInitiate", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "initiate1",
				APIOpts: map[string]any{
					utils.MetaSession:  true,
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaInitiate: true,
					utils.MetaUsage:    90 * time.Second,
					utils.MetaOriginID: originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(initiate): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountSUsage missing *default")
		}
		if usage != 90*time.Second {
			t.Errorf("AccountSUsage[*default] = %v, want 1m30s", usage)
		}
	})

	t.Run("SessionTerminate", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "terminate1",
				APIOpts: map[string]any{
					utils.MetaSession:   true,
					utils.MetaAccounts:  true,
					utils.MetaDebit:     true,
					utils.MetaTerminate: true,
					utils.MetaUsage:     40 * time.Second,
					utils.MetaOriginID:  originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(terminate): %v", err)
		}
	})

	var aSessions []*ExternalSession
	if err := client.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err == nil || err.Error() != utils.NotFoundCaps {
		t.Errorf("expected session to be closed after terminate, got: %+v, err: %v", aSessions, err)
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else if want := utils.NewDecimalFromFloat64(float64(170 * time.Second)); acnt.Balances["BAL1"].Units.Compare(want) != 0 {
		t.Errorf("expected 170s remaining after terminate (300s - 90s initiate - 40s terminate), got: %+v", acnt.Balances["BAL1"].Units)
	}
}

func TestSessionSv1ProcessEventDebitTwoBalancesByWeight(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
        "*chargers": [{"connIDs": ["*localhost"]}],
        "*accounts": [{"connIDs": ["*localhost"]}]
    }
},
"chargers": {
    "enabled": true
},
"accounts": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)

	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				RunID:        utils.MetaDefault,
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"MonetaryBal": {
						ID:      "MonetaryBal",
						Type:    utils.MetaConcrete,
						Weights: utils.DynamicWeights{{Weight: 20}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(1, 1),
							},
						},
						Units: utils.NewDecimalFromFloat64(2.0),
					},
					"FreeMinutesBal": {
						ID:      "FreeMinutesBal",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(300 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("debit", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "twoBalancesEvent1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    60 * time.Second,
					utils.MetaOriginID: "OriginIDTwoBalances",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(debit): %v", err)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else {
		if want := utils.NewDecimalFromFloat64(0.00); acnt.Balances["MonetaryBal"].Units.Compare(want) != 0 {
			t.Errorf("expected MonetaryBal fully consumed to 0.00, got: %+v", acnt.Balances["MonetaryBal"].Units)
		}
		if want := utils.NewDecimalFromFloat64(float64(260 * time.Second)); acnt.Balances["FreeMinutesBal"].Units.Compare(want) != 0 {
			t.Errorf("expected FreeMinutesBal at 260s remaining, got: %+v", acnt.Balances["FreeMinutesBal"].Units)
		}
	}
}
