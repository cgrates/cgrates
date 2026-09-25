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

func TestSessionSv1ProcessEventDebitAbstractCSV(t *testing.T) {
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
		TpFiles: map[string]string{
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
cgrates.org,DEFAULT,,,,*default,*none`,
			utils.AccountsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,,,,Balance1,,;10,,*abstract,300s,,,;1s;0;0,,,`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)

	t.Run("debit", func(t *testing.T) {
		var rply V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "csvBasicDebit",
				APIOpts: map[string]any{
					utils.MetaChargers: true,
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    30 * time.Second,
					utils.MetaOriginID: "OriginIDCSVBasic",
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

	if want := utils.NewDecimalFromFloat64(float64(270 * time.Second)); acnt.Balances["Balance1"].Units.Compare(want) != 0 {
		t.Errorf("expected Balance1 debited to 270s remaining, got: %+v", acnt.Balances["Balance1"].Units)
	}
}

func TestSessionSv1ProcessEventDebitConcreteCSV(t *testing.T) {
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
    "enabled": true,
    "conns": {
        "*rates": [{"connIDs": ["*localhost"]}]
    }
},
"rates": {
    "enabled": true
},
"admins": {
    "enabled": true
}
}`,
		TpFiles: map[string]string{
			utils.ChargersCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,RunID,AttributeIDs
cgrates.org,DEFAULT,,,,*default,*none`,
			utils.AccountsCsv: `#Tenant,ID,FilterIDs,Weights,Blockers,Opts,BalanceID,BalanceFilterIDs,BalanceWeights,BalanceBlockers,BalanceType,BalanceUnits,BalanceUnitFactors,BalanceOpts,BalanceCostIncrements,BalanceAttributeIDs,BalanceRateProfileIDs,ThresholdIDs
cgrates.org,1001,,,,,Balance1,,;10,,*concrete,10,,,,,RP1,`,
			utils.RatesCsv: `#Tenant,ID,FilterIDs,Weights,MinCost,MaxCost,MaxCostStrategy,RateID,RateFilterIDs,RateActivationTimes,RateWeights,RateBlocker,RateIntervalStart,RateFixedFee,RateRecurrentFee,RateUnit,RateIncrement
cgrates.org,RP1,,;0,,,,RT1,,* * * * *,;0,,0s,0,0.05,1s,1s`,
		},
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}

	client, _ := ng.Run(t)

	var rply V1ProcessEventReply
	if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
		&utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "csvConcreteDebit",
			APIOpts: map[string]any{
				utils.MetaChargers: true,
				utils.MetaAccounts: true,
				utils.MetaDebit:    true,
				utils.MetaUsage:    30 * time.Second,
				utils.MetaOriginID: "OriginIDCSVConcrete",
			},
			Event: map[string]any{
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   "2018-01-07T17:00:00Z",
			},
		}, &rply); err != nil {
		t.Fatalf("ProcessEvent(debit): %v", err)
	}

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	}

	if want := utils.NewDecimalFromFloat64(8.5); acnt.Balances["Balance1"].Units.Compare(want) != 0 {
		t.Errorf("expected Balance1 debited, got: %+v", acnt.Balances["Balance1"].Units)
	}
}
