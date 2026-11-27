//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"context"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionTerminateRefund(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	ng := engine.TestEngine{
		ConfigJSON: `{
"admins": {
	"enabled": true
},
"accounts": {
	"enabled": true
},
"sessions": {
	"enabled": true,
	"conns": {
		"*accounts": [{"connIDs": ["*localhost"]}]
	},
}
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)
	var reply string
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant:    "cgrates.org",
				ID:        "1001",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				Balances: map[string]*utils.Balance{
					"dataBal": {
						Type:  utils.MetaAbstract,
						Units: utils.NewDecimal(1000000000, 0),
						CostIncrements: []*utils.CostIncrement{
							{
								FixedFee:  utils.NewDecimal(0, 0),
								Increment: utils.NewDecimal(1, 0),
							},
						},
					},
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	originID := "sessionRefund1"

	t.Run("AuthorizeEvent", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
				APIOpts: map[string]any{
					utils.MetaAccounts:  "true",
					utils.MetaAuthorize: "true",
					utils.MetaOriginID:  originID,
					utils.MetaUsage:     "1000",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.AccountsUsage[utils.MetaPrimary] != 1000 {
			t.Errorf("unexpected accounts usage: %v", rply.AccountsUsage)
		}
	})

	t.Run("StartEvent", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
				APIOpts: map[string]any{
					utils.MetaAccounts:     "true",
					utils.MetaDebit:        "true",
					utils.MetaInterimUsage: "50",
					utils.MetaOriginID:     originID,
					utils.MetaSession:      "true",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}
		if rply.AccountsUsage[utils.MetaPrimary] != 50 {
			t.Errorf("unexpected accounts usage: %v", rply.AccountsUsage)
		}

		var acnt utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
			&acnt); err != nil {
			t.Fatal(err)
		}
		dataBal, exist := acnt.Balances["dataBal"]
		if !exist {
			t.Fatal("balance doesnt exist")
		}
		if balUnits := dataBal.Units; balUnits.Compare(utils.NewDecimal(999999950, 0)) != 0 {
			t.Errorf("expected balance unitss remaning 999999950, received %v", balUnits)
		}
	})
	t.Skip("panics on *terminate: Accounting entry with empty BalanceID gives a nil balance in abstractConcretes")

	t.Run("TerminateEvent", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.GenUUID(),
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.Destination:  "1002",
				},
				APIOpts: map[string]any{
					utils.MetaAccounts:   "true",
					utils.MetaDebit:      "true",
					utils.MetaOriginID:   originID,
					utils.MetaSession:    "true",
					utils.MetaTerminate:  "true",
					utils.MetaTotalUsage: "44",
				},
			}, &rply); err != nil {
			t.Fatal(err)
		}

		var acnt utils.Account
		if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
			&acnt); err != nil {
			t.Fatal(err)
		}
		dataBal, exist := acnt.Balances["dataBal"]
		if !exist {
			t.Fatal("balance doesnt exist")
		}
		if balUnits := dataBal.Units; balUnits.Compare(utils.NewDecimal(999999956, 0)) != 0 {
			t.Errorf("expected balance units remaining 999999956, received %v", balUnits)
		}
	})
}
