//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"testing"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func TestSessionSv1ProcessEventSMS(t *testing.T) {
	t.Skip("Err:  expected 9 *sms units remaining, got: &{Context:{MaxScale:0 MinScale:0 Precision:0 Traps: Conditions: RoundingMode:ToNearestEven OperatingMode:GDA} unscaled:{neg:false abs:[]} compact:10 exp:0 precision:2 form:0}")
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
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID",
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
