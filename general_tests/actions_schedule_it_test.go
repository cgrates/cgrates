//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestScheduleAddedAccount(t *testing.T) {
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
  "actions": {
    "enabled": true,
    "conns": {
      "*accounts": [
        {
          "ConnIDs": [
            "*localhost"
          ]
        }
      ]
    }
  }
}`,
		DBCfg:    engine.InternalDBCfg,
		Encoding: *utils.Encoding,
	}
	client, _ := ng.Run(t)
	ctx := context.Background()
	var reply string
	for _, id := range []string{"account1", "account2"} {
		if err := client.Call(ctx, utils.AdminSv1SetAccount, &utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     id,
				Balances: map[string]*utils.Balance{
					"balance": {
						ID:    "balance",
						Type:  utils.MetaAbstract,
						Units: utils.NewDecimal(0, 0),
					},
				},
			},
		}, &reply); err != nil {
			t.Fatal(err)
		}
	}
	if err := client.Call(ctx, utils.AdminSv1SetActionProfile, &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant:   "cgrates.org",
			ID:       "profile",
			Schedule: "@every 3s",
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: {"account1": {}},
			},
			Actions: []*utils.APAction{
				{
					ID:   "add",
					Type: utils.MetaAddBalance,
					Diktats: []*utils.APDiktat{
						{
							ID: "add",
							Opts: map[string]any{
								"*balancePath":  "*balance.balance.Units",
								"*balanceValue": 1,
							},
						},
					},
				},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}
	event := &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"profile"},
		},
	}
	if err := client.Call(ctx, utils.ActionSv1ScheduleActions, event, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(ctx, utils.AdminSv1SetActionProfile, &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "profile",
			Targets: map[string]utils.StringSet{
				utils.MetaAccounts: {"account2": {}},
			},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(ctx, utils.ActionSv1ScheduleActions, event, &reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(4 * time.Second)
	for _, id := range []string{"account1", "account2"} {
		var account utils.Account
		if err := client.Call(ctx, utils.AdminSv1GetAccount, &utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     id,
			},
		}, &account); err != nil {
			t.Fatal(err)
		}
		if got := account.Balances["balance"].Units; got.Big.Cmp(utils.NewDecimal(1, 0).Big) != 0 {
			t.Errorf("account %s: expected 1, received %s", id, got)
		}
	}
}
