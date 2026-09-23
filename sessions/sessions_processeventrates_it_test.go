//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions

import (
	"strings"
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

func TestSessionSv1ProcessEventDebitBlockedByBalanceBlocker(t *testing.T) {
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
					"BalHigh": {
						ID:       "BalHigh",
						Type:     utils.MetaAbstract,
						Weights:  utils.DynamicWeights{{Weight: 20}},
						Blockers: utils.DynamicBlockers{{Blocker: true}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(30 * time.Second)),
					},
					"BalLow": {
						ID:      "BalLow",
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
				ID:     "blockedEvent1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    60 * time.Second,
					utils.MetaOriginID: "OriginIDBlocked",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(debit): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountSUsage missing *default")
		}
		if usage != 30*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 30s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}
	if want := utils.NewDecimalFromFloat64(0); acnt.Balances["BalHigh"].Units.Compare(want) != 0 {
		t.Errorf("expected BalHigh consumed to 0, got: %+v", acnt.Balances["BalHigh"].Units)
	}
	if want := utils.NewDecimalFromFloat64(float64(300 * time.Second)); acnt.Balances["BalLow"].Units.Compare(want) != 0 {
		t.Errorf("expected BalLow 300s (blocked), got: %+v", acnt.Balances["BalLow"].Units)
	}
}

func TestSessionSv1ProcessEventMultipleUpdates(t *testing.T) {
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

	originID := "OriginIDMultiUpdate"

	t.Run("Initiate30s", func(t *testing.T) {
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
					utils.MetaUsage:    30 * time.Second,
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
	})

	t.Run("Update1Plus20s", func(t *testing.T) {
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
					utils.MetaUsage:    20 * time.Second,
					utils.MetaOriginID: originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(update1): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 20*time.Second {
			t.Errorf("AccountsUsage[*default] after update1 = %v, want 20s", usage)
		}
	})

	t.Run("Update2Plus15s", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "update2",
				APIOpts: map[string]any{
					utils.MetaSession:  true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUpdate:   true,
					utils.MetaUsage:    15 * time.Second,
					utils.MetaOriginID: originID,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(update2): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 15*time.Second {
			t.Errorf("AccountsUsage[*default] after update2 = %v, want 15s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else if want := utils.NewDecimalFromFloat64(float64(235 * time.Second)); acnt.Balances["BAL1"].Units.Compare(want) != 0 {
		t.Errorf("expected 235s remaining (300s - 30s initiate - 20s update1 - 15s update2), got: %+v", acnt.Balances["BAL1"].Units)
	}
}

func TestSessionSv1ProcessEventAuthorizeMaxUsage(t *testing.T) {
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
						Units: utils.NewDecimalFromFloat64(float64(50 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("authorizeRequestingMoreThanAvailable", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "authorizeOnly",
				APIOpts: map[string]any{
					utils.MetaChargers:  true,
					utils.MetaAccounts:  true,
					utils.MetaAuthorize: true,
					utils.MetaUsage:     100 * time.Second,
					utils.MetaOriginID:  "OriginIDAuthorizeOnly",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(authorize): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 50*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 50s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else if want := utils.NewDecimalFromFloat64(float64(50 * time.Second)); acnt.Balances["BAL1"].Units.Compare(want) != 0 {
		t.Errorf("balance should remain same after *authorize (no *debit flag), got: %+v, want 50s", acnt.Balances["BAL1"].Units)
	}
}

func TestSessionSv1ProcessEventFilteredBalances(t *testing.T) {
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
					"LocalBal": {
						ID:        "LocalBal",
						FilterIDs: []string{"*prefix:~*req.Destination:1"},
						Type:      utils.MetaConcrete,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(10, 2),
							},
						},
						Units: utils.NewDecimalFromFloat64(100.0),
					},
					"IntlBal": {
						ID:        "IntlBal",
						FilterIDs: []string{"*prefix:~*req.Destination:0044"},
						Type:      utils.MetaConcrete,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(50, 2),
							},
						},
						Units: utils.NewDecimalFromFloat64(100.0),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("LocalCall", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "localCall",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    10 * time.Second,
					utils.MetaOriginID: "OriginIDLocal",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(localCall): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 10*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 10s", usage)
		}
	})

	var acntAfterLocal utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acntAfterLocal); err != nil {
		t.Fatalf("AdminSv1GetAccount (after local): %v", err)
	} else {
		if want := utils.NewDecimalFromFloat64(99.0); acntAfterLocal.Balances["LocalBal"].Units.Compare(want) != 0 {
			t.Errorf("expected LocalBal 99.0 after local call (100 - 10s*0.10), got: %+v", acntAfterLocal.Balances["LocalBal"].Units)
		}
		if want := utils.NewDecimalFromFloat64(100.0); acntAfterLocal.Balances["IntlBal"].Units.Compare(want) != 0 {
			t.Errorf("expected IntlBal untouched at 100.0 after local call, got: %+v", acntAfterLocal.Balances["IntlBal"].Units)
		}
	}

	t.Run("InternationalCall", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "intlCall",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    10 * time.Second,
					utils.MetaOriginID: "OriginIDIntl",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "00447911123456",
					utils.AnswerTime:   "2018-01-07T17:05:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(intlCall): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 10*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 10s", usage)
		}
	})

	var acntFinal utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acntFinal); err != nil {
		t.Fatalf("AdminSv1GetAccount (final): %v", err)
	} else {
		if want := utils.NewDecimalFromFloat64(99.0); acntFinal.Balances["LocalBal"].Units.Compare(want) != 0 {
			t.Errorf("expected LocalBal still 99.0 after intl call, got: %+v", acntFinal.Balances["LocalBal"].Units)
		}
		if want := utils.NewDecimalFromFloat64(95.0); acntFinal.Balances["IntlBal"].Units.Compare(want) != 0 {
			t.Errorf("expected IntlBal 95.0 after intl call (100 - 10s*0.50), got: %+v", acntFinal.Balances["IntlBal"].Units)
		}
	}
}

func TestSessionSv1ProcessEventActivationIntervalBalance(t *testing.T) {
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

	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrActivationInterval",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaActivationInterval,
						Element: "~*req.AnswerTime",
						Values:  []string{"2027-01-01T00:00:00Z"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"ActIntervalBal": {
						ID:        "ActIntervalBal",
						FilterIDs: []string{"FltrActivationInterval"},
						Type:      utils.MetaAbstract,
						Weights:   utils.DynamicWeights{{Weight: 20}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(300 * time.Second)),
					},
					"FallbackBal": {
						ID:      "FallbackBal",
						Type:    utils.MetaConcrete,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimalFromFloat64(float64(time.Second)),
								RecurrentFee: utils.NewDecimal(10, 2),
							},
						},
						Units: utils.NewDecimalFromFloat64(100.0),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("BalanceDebit", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "skipActIntervalBal",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    10 * time.Second,
					utils.MetaOriginID: "OriginIDActIntervalBalSkip",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(debit): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 10*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 10s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}
	wantActIntervalBal := utils.NewDecimalFromFloat64(float64(300 * time.Second))
	if acnt.Balances["ActIntervalBal"].Units.Compare(wantActIntervalBal) != 0 {
		t.Errorf("expected ActIntervalBal not consumed (not yet active), got: %+v", acnt.Balances["ActIntervalBal"].Units)
	}
	wantFallBackBal := utils.NewDecimalFromFloat64(99.0)
	if acnt.Balances["FallbackBal"].Units.Compare(wantFallBackBal) != 0 {
		t.Errorf("expected FallbackBal 99.0 after debit (100 - 10s*0.10), got: %+v", acnt.Balances["FallbackBal"].Units)
	}

}

func TestSessionSv1ProcessEventBlockerBalanceStopsDebit(t *testing.T) {
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
		// LogBuffer: new(bytes.Buffer),
	}

	client, _ := ng.Run(t)

	// t.Cleanup(func() {
	// 	if ng.LogBuffer != nil {
	// 		fmt.Println(ng.LogBuffer)
	// 	}
	// })

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
					"BalBlocker": {
						ID:      "BalBlocker",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						Blockers: utils.DynamicBlockers{
							{Blocker: true},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(5 * time.Second)),
					},
					"BalFallback": {
						ID:      "BalFallback",
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

	t.Run("DebitWithBlockedBalance", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "debitBlocker",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    20 * time.Second,
					utils.MetaOriginID: "OriginIDBlocker",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)

		if err == nil {
			usage, ok := rply.AccountsUsage[utils.MetaDefault]
			if !ok {
				t.Fatal("AccountsUsage missing *default")
			}
			if usage != 5*time.Second {
				t.Errorf("expected debit stopped at 5s due to Blocker, got %v debited", usage)
			}
		} else {
			t.Logf("err): %v", err)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	if want := utils.NewDecimalFromFloat64(0); acnt.Balances["BalBlocker"].Units.Compare(want) != 0 {
		t.Errorf("expected BalBlocker fully consumed (0s), got: %+v", acnt.Balances["BalBlocker"].Units)
	}

	if want := utils.NewDecimalFromFloat64(float64(300 * time.Second)); acnt.Balances["BalFallback"].Units.Compare(want) != 0 {
		t.Errorf("BalFallback  used despite Blocker, expected 300s, got: %+v",
			acnt.Balances["BalFallback"].Units)
	}
}

func TestSessionSv1ProcessEventBlockerBalanceWeights(t *testing.T) {
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
					"BalFirst": {
						ID:      "BalFirst",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(5 * time.Second)),
					},
					"BalSecondBlocked": {
						ID:      "BalSecondBlocked",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						Blockers: utils.DynamicBlockers{
							{Blocker: true},
						},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(5 * time.Second)),
					},
					"BalThirdFallback": {
						ID:      "BalThirdFallback",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 1}},
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

	t.Run("DebitWithThreeBalancesStopsAtBlocker", func(t *testing.T) {
		var rply V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "debitMidBlocker",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    20 * time.Second,
					utils.MetaOriginID: "OriginIDMidBlocker",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply)
		if err != nil {
			t.Fatalf("ProcessEvent: %v", err)
		}

		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 10*time.Second {
			t.Errorf("expected total debit stopped at 10s (5s+5s with the two usable balances), got %v", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	if want := utils.NewDecimalFromFloat64(0); acnt.Balances["BalFirst"].Units.Compare(want) != 0 {
		t.Errorf("expected BalFirst fully consumed (0s) due to weight order, got: %+v",
			acnt.Balances["BalFirst"].Units)
	}

	if want := utils.NewDecimalFromFloat64(0); acnt.Balances["BalSecondBlocked"].Units.Compare(want) != 0 {
		t.Errorf("expected BalSecondBlocked fully consumed (0s), got: %+v",
			acnt.Balances["BalSecondBlocked"].Units)
	}

	if want := utils.NewDecimalFromFloat64(float64(300 * time.Second)); acnt.Balances["BalThirdFallback"].Units.Compare(want) != 0 {
		t.Errorf("BalThirdFallback was used despite Blocker on BalSecondBlocked, expected at 300s, got: %+v",
			acnt.Balances["BalThirdFallback"].Units)
	}
}

func TestSessionSv1ProcessEventVoiceSMSData(t *testing.T) {
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

	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrVoiceToR",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.ToR",
						Values:  []string{utils.MetaVoice},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(voice): %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrSMSToR",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.ToR",
						Values:  []string{utils.MetaSMS},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(sms): %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrDataToR",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.ToR",
						Values:  []string{utils.MetaData},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(data): %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"VoiceBal": {
						ID:        "VoiceBal",
						FilterIDs: []string{"FltrVoiceToR"},
						Type:      utils.MetaAbstract,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(300 * time.Second)),
					},
					"SMSBal": {
						ID:        "SMSBal",
						FilterIDs: []string{"FltrSMSToR"},
						Type:      utils.MetaConcrete,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(5, 2),
							},
						},
						Units: utils.NewDecimalFromFloat64(50.0),
					},
					"DataBal": {
						ID:        "DataBal",
						FilterIDs: []string{"FltrDataToR"},
						Type:      utils.MetaConcrete,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(1, 2),
							},
						},
						Units: utils.NewDecimalFromFloat64(500.0),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("VoiceCallDebitsOnlyVoiceBal", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "voiceCall1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    10 * time.Second,
					utils.MetaOriginID: "OriginIDVoice1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaVoice,
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(voice): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 10*time.Second {
			t.Errorf("AccountsUsage[*default] = %v, want 10s (VoiceBal)", usage)
		}
	})

	t.Run("SMSDebitsOnlySMSBal", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsMsg1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "OriginIDSMS1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:05:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(sms): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 1 {
			t.Errorf("AccountsUsage[*default] = %v, want 1 (SMSBal)", usage)
		}
	})

	t.Run("DataDebitsOnlyDataBal", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dataSession1",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    100,
					utils.MetaOriginID: "OriginIDData1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaData,
					utils.Destination:  "1002",
					utils.AnswerTime:   "2018-01-07T17:10:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(data): %v", err)
		}
		usage, ok := rply.AccountsUsage[utils.MetaDefault]
		if !ok {
			t.Fatal("AccountsUsage missing *default")
		}
		if usage != 100 {
			t.Errorf("AccountsUsage[*default] = %v, want 100 (DataBal)", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	wantVoiceBal := utils.NewDecimalFromFloat64(float64(290 * time.Second))
	if acnt.Balances["VoiceBal"].Units.Compare(wantVoiceBal) != 0 {
		t.Errorf("VoiceBal = %+v, want 290s", acnt.Balances["VoiceBal"].Units)
	}

	wantSMSBal := utils.NewDecimalFromFloat64(49.95)
	if acnt.Balances["SMSBal"].Units.Compare(wantSMSBal) != 0 {
		t.Errorf("SMSBal = %+v, want 49.95", acnt.Balances["SMSBal"].Units)
	}

	wantDataBal := utils.NewDecimalFromFloat64(499.0)
	if acnt.Balances["DataBal"].Units.Compare(wantDataBal) != 0 {
		t.Errorf("DataBal = %+v, want 499", acnt.Balances["DataBal"].Units)
	}
}

func TestSessionSv1ProcessEventAccountsForceUsage(t *testing.T) {
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

	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrVoiceToR",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.ToR",
						Values:  []string{utils.MetaVoice},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(voice): %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"VoiceBal": {
						ID:        "VoiceBal",
						FilterIDs: []string{"FltrVoiceToR"},
						Type:      utils.MetaAbstract,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(2 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	var rply V1ProcessEventReply
	err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ForceUsageFlag",
			APIOpts: map[string]any{
				utils.MetaChargers:           true,
				utils.MetaAccounts:           true,
				utils.MetaDebit:              true,
				utils.MetaUsage:              5 * time.Minute,
				utils.OptsAccountsForceUsage: true,
				utils.MetaOriginID:           "OriginIDForceUsage",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.ToR:          utils.MetaVoice,
				utils.Destination:  "1002",
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
		}, &rply)

	if err == nil {
		t.Fatal("expected an err got nil")
	}
	if err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("ProcessEvent error = %q, want %q", err.Error(), utils.ErrPartiallyExecuted.Error())
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	wantVoiceBal := utils.NewDecimalFromFloat64(float64(2 * time.Second))
	if acnt.Balances["VoiceBal"].Units.Compare(wantVoiceBal) != 0 {
		t.Errorf("VoiceBal = %+v, want 2s", acnt.Balances["VoiceBal"].Units)
	}
}

func TestSessionSv1ProcessEventChargerRuns(t *testing.T) {
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
				ID:           "ChargerRun1",
				RunID:        "run1",
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile(run1): %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "ChargerRunB",
				RunID:        "run2",
				AttributeIDs: []string{utils.MetaNone},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetChargerProfile(run2): %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrRun1",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*opts.*runID",
						Values:  []string{"run1"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(run1): %v", err)
	}
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrRun2",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*opts.*runID",
						Values:  []string{"run2"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter(run2): %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"Balance1": {
						ID:        "Balance1",
						FilterIDs: []string{"FltrRun1"},
						Type:      utils.MetaAbstract,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(100 * time.Second)),
					},
					"Balance2": {
						ID:        "Balance2",
						FilterIDs: []string{"FltrRun2"},
						Type:      utils.MetaAbstract,
						Weights:   utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(100 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	var rply V1ProcessEventReply
	if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "MultiRunEvent",
			APIOpts: map[string]any{
				utils.MetaChargers: true,
				utils.MetaAccounts: true,
				utils.MetaDebit:    true,
				utils.MetaUsage:    10 * time.Second,
				utils.MetaOriginID: "OriginIDMultiRun",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
		}, &rply); err != nil {
		t.Fatalf("ProcessEvent: %v", err)
	}

	usageA, ok := rply.AccountsUsage["run1"]
	if !ok {
		t.Fatal("AccountsUsage missing run1")
	}
	if usageA != 10*time.Second {
		t.Errorf("AccountsUsage[run1] = %v, want 10s", usageA)
	}
	usageB, ok := rply.AccountsUsage["run2"]
	if !ok {
		t.Fatal("AccountsUsage missing run2")
	}
	if usageB != 10*time.Second {
		t.Errorf("AccountsUsage[run2] = %v, want 10s", usageB)
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	wantBalance1 := utils.NewDecimalFromFloat64(float64(90 * time.Second))
	if acnt.Balances["Balance1"].Units.Compare(wantBalance1) != 0 {
		t.Errorf("Balance1 = %+v, want 90s", acnt.Balances["Balance1"].Units)
	}
	wantBalance2 := utils.NewDecimalFromFloat64(float64(90 * time.Second))
	if acnt.Balances["Balance2"].Units.Compare(wantBalance2) != 0 {
		t.Errorf("Balance2 = %+v, want 90s", acnt.Balances["Balance2"].Units)
	}
}

func TestSessionSv1ProcessEventChargerFilterIDsNoMatch(t *testing.T) {
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
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrSubjectNoMatch",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Subject",
						Values:  []string{"noMatch"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter: %v", err)
	}

	if err := client.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		&utils.ChargerProfileWithAPIOpts{
			ChargerProfile: &utils.ChargerProfile{
				Tenant:       "cgrates.org",
				ID:           "DEFAULT",
				FilterIDs:    []string{"FltrSubjectNoMatch"},
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
					"AbstractBal": {
						ID:      "AbstractBal",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(60 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	var rply V1ProcessEventReply
	err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EvNoMatch",
			APIOpts: map[string]any{
				utils.MetaChargers: true,
				utils.MetaAccounts: true,
				utils.MetaDebit:    true,
				utils.MetaUsage:    30 * time.Second,
				utils.MetaOriginID: "originNoMatch",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1009",
				utils.Destination:  "1002",
				utils.ToR:          utils.MetaVoice,
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
		}, &rply)

	if err == nil {
		t.Fatal("expected CHARGERS_ERROR:NOT_FOUND, got nil")
	}
	if !strings.Contains(err.Error(), "CHARGERS_ERROR:NOT_FOUND") {
		t.Errorf("ProcessEvent error = %q, want CHARGERS_ERROR:NOT_FOUND", err.Error())
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	wantUnits := utils.NewDecimalFromFloat64(float64(60 * time.Second))
	if acnt.Balances["AbstractBal"].Units.Compare(wantUnits) != 0 {
		t.Errorf("AbstractBal = %+v, want 60s", acnt.Balances["AbstractBal"].Units)
	}
}
func TestSessionSv1ProcessEventAccountFilterIDsNoMatch(t *testing.T) {
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
	if err := client.Call(context.Background(), utils.AdminSv1SetFilter,
		&engine.FilterWithAPIOpts{
			Filter: &engine.Filter{
				Tenant: "cgrates.org",
				ID:     "FltrSubjectNoMatch",
				Rules: []*engine.FilterRule{
					{
						Type:    utils.MetaString,
						Element: "~*req.Subject",
						Values:  []string{"noMatch"},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetFilter: %v", err)
	}

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
				Tenant:    "cgrates.org",
				ID:        "1001",
				FilterIDs: []string{"FltrSubjectNoMatch"},
				Balances: map[string]*utils.Balance{
					"AbstractBal": {
						ID:      "AbstractBal",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(float64(60 * time.Second)),
					},
				},
			},
		}, &reply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	var rply V1ProcessEventReply
	err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EvAccountNoMatch",
			APIOpts: map[string]any{
				utils.MetaChargers: true,
				utils.MetaAccounts: true,
				utils.MetaDebit:    true,
				utils.MetaUsage:    30 * time.Second,
				utils.MetaOriginID: "originAccountNoMatch",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Subject:      "1009",
				utils.Destination:  "1002",
				utils.ToR:          utils.MetaVoice,
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
		}, &rply)

	if err == nil {
		t.Fatal("expected PARTIALLY_EXECUTED, got nil")
	}
	if err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("ProcessEvent error = %q, want %q", err.Error(), utils.ErrPartiallyExecuted.Error())
	}
	if len(rply.AccountsUsage) != 0 {
		t.Errorf("expected empty AccountsUsage, got %+v", rply.AccountsUsage)
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	wantUnits := utils.NewDecimalFromFloat64(float64(60 * time.Second))
	if acnt.Balances["AbstractBal"].Units.Compare(wantUnits) != 0 {
		t.Errorf("AbstractBal = %+v, want 60s", acnt.Balances["AbstractBal"].Units)
	}
}
