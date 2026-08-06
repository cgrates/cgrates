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

func TestSessionSv1ProcessEventSMS(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
        "*accounts": [{"connIDs": ["*localhost"]}]
    },
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

	// Account 1001 with a single *sms abstract balance: 10 units
	var setReply string
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"SMS1": {
						ID:      "SMS1",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.ToR:*sms"},
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(10),
					},
				},
			},
		}, &setReply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("processEventSMS", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID",
					utils.MetaUR:       true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed for SMS event: %v", err)
		}
		usageRecord, has := rply.UsageRecords[utils.MetaPrimary]
		if !has {
			t.Fatalf("missing %s UsageRecord", utils.MetaPrimary)
		}
		usage, err := utils.IfaceAsDuration(usageRecord.APIOpts[utils.MetaUsage])
		if err != nil {
			t.Fatal(err)
		}
		if usage != time.Nanosecond {
			t.Errorf("unexpected UsageRecord usage: %s", usage)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else if acnt.Balances["SMS1"].Units.Compare(utils.NewDecimalFromFloat64(9)) != 0 {
		t.Errorf("expected 9 *sms units remaining, got: %+v", acnt.Balances["SMS1"].Units)
	}
}
