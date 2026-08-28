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

func TestSessionSv1ProcessEventSMSFallbackToMonetary(t *testing.T) {
	ng := engine.TestEngine{
		ConfigJSON: `{
"sessions": {
    "enabled": true,
    "conns": {
        "*accounts": [{"connIDs": ["*localhost"]}]
    }
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

	var setReply string
	if err := client.Call(context.Background(), utils.AdminSv1SetAccount,
		&utils.AccountWithAPIOpts{
			Account: &utils.Account{
				Tenant: "cgrates.org",
				ID:     "1001",
				Balances: map[string]*utils.Balance{
					"SMSFree": {
						ID:      "SMSFree",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 20}},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.ToR:*sms"},
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(3),
					},
					"MonetaryBal": {
						ID:      "MonetaryBal",
						Type:    utils.MetaConcrete,
						Weights: utils.DynamicWeights{{Weight: 10}},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.ToR:*sms"},
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(1, 1),
							},
						},
						Units: utils.NewDecimalFromFloat64(0.25),
					},
				},
			},
		}, &setReply); err != nil {
		t.Fatalf("AdminSv1SetAccount: %v", err)
	}

	t.Run("smsFree1", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent1",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID1",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(smsFree1): %v", err)
		}
	})

	var acnt1 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt1); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsFree1): %v", err)
	} else {
		if want := utils.NewDecimalFromFloat64(2); acnt1.Balances["SMSFree"].Units.Compare(want) != 0 {
			t.Errorf("expected SMSFree at 2, got: %+v", acnt1.Balances["SMSFree"].Units)
		}
		if want := utils.NewDecimalFromFloat64(0.25); acnt1.Balances["MonetaryBal"].Units.Compare(want) != 0 {
			t.Errorf("expected MonetaryBal untouched at 0.25, got: %+v", acnt1.Balances["MonetaryBal"].Units)
		}
	}

	t.Run("smsFree2", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent2",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID2",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(smsFree2): %v", err)
		}
	})

	var acnt2 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt2); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsFree2): %v", err)
	} else if want := utils.NewDecimalFromFloat64(1); acnt2.Balances["SMSFree"].Units.Compare(want) != 0 {
		t.Errorf("expected SMSFree at 1, got: %+v", acnt2.Balances["SMSFree"].Units)
	}

	t.Run("smsFree3", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent3",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID3",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(smsFree3): %v", err)
		}
	})

	var acnt3 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt3); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsFree3): %v", err)
	} else {
		if want := utils.NewDecimalFromFloat64(0); acnt3.Balances["SMSFree"].Units.Compare(want) != 0 {
			t.Errorf("expected SMSFree exhausted at 0, got: %+v", acnt3.Balances["SMSFree"].Units)
		}
		if want := utils.NewDecimalFromFloat64(0.25); acnt3.Balances["MonetaryBal"].Units.Compare(want) != 0 {
			t.Errorf("expected MonetaryBal untouched at 0.25, got: %+v", acnt3.Balances["MonetaryBal"].Units)
		}
	}

	t.Run("smsFallbackToMonetary1", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent4",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID4",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(smsFallbackToMonetary1): %v", err)
		}
	})

	var acnt4 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt4); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsFallbackToMonetary1): %v", err)
	} else if want := utils.NewDecimalFromFloat64(0.15); acnt4.Balances["MonetaryBal"].Units.Compare(want) != 0 {
		t.Errorf("expected MonetaryBal at 0.15, got: %+v", acnt4.Balances["MonetaryBal"].Units)
	}

	t.Run("smsFallbackToMonetary2", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent5",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    1,
					utils.MetaOriginID: "smsOriginID5",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent(smsFallbackToMonetary2): %v", err)
		}
	})

	var acnt5 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt5); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsFallbackToMonetary2): %v", err)
	} else if want := utils.NewDecimalFromFloat64(0.05); acnt5.Balances["MonetaryBal"].Units.Compare(want) != 0 {
		t.Errorf("expected MonetaryBal at 0.05, got: %+v", acnt5.Balances["MonetaryBal"].Units)
	}

	t.Run("smsInsufficientCredit", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "smsEvent6",
				APIOpts: map[string]any{
					utils.MetaAccounts:           true,
					utils.MetaDebit:              true,
					utils.MetaUsage:              1,
					utils.MetaOriginID:           "smsOriginID6",
					utils.MetaAccountsForceUsage: true,
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaSMS,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:00:00Z",
				},
			}, &rply)
		if err == nil {
			t.Fatalf("expected err due to insufficient credit, got nil (rply=%+v)", rply)
		}
	})

	var acnt6 utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt6); err != nil {
		t.Fatalf("AdminSv1GetAccount(after smsInsufficientCredit): %v", err)
	} else if want := utils.NewDecimalFromFloat64(0.05); acnt6.Balances["MonetaryBal"].Units.Compare(want) != 0 {
		t.Errorf("expected MonetaryBal unchanged at 0.05 after no debit, got: %+v", acnt6.Balances["MonetaryBal"].Units)
	}
}

func TestSessionSv1ProcessEventSMSData(t *testing.T) {
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
					"DATA1": {
						ID:      "DATA1",
						Type:    utils.MetaAbstract,
						Weights: utils.DynamicWeights{{Weight: 5}},
						CostIncrements: []*utils.CostIncrement{
							{
								FilterIDs:    []string{"*string:~*req.ToR:*data"},
								Increment:    utils.NewDecimal(1, 0),
								RecurrentFee: utils.NewDecimal(0, 0),
							},
						},
						Units: utils.NewDecimalFromFloat64(1000),
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

	t.Run("processEventData", func(t *testing.T) {
		var rply sessions.V1ProcessEventReply
		if err := client.Call(context.Background(), utils.SessionSv1ProcessEvent,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "dataEvent",
				APIOpts: map[string]any{
					utils.MetaAccounts: true,
					utils.MetaDebit:    true,
					utils.MetaUsage:    400,
					utils.MetaOriginID: "dataOriginID",
				},
				Event: map[string]any{
					utils.AccountField: "1001",
					utils.ToR:          utils.MetaData,
					utils.Destination:  "+40123456789",
					utils.AnswerTime:   "2024-01-07T17:01:00Z",
				},
			}, &rply); err != nil {
			t.Fatalf("ProcessEvent failed for Data event: %v", err)
		}
	})

	var acnt utils.Account
	if err := client.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "1001"}},
		&acnt); err != nil {
		t.Fatalf("AdminSv1GetAccount: %v", err)
	} else {
		if acnt.Balances["SMS1"].Units.Compare(utils.NewDecimalFromFloat64(9)) != 0 {
			t.Errorf("expected 9 *sms units remaining, got: %+v", acnt.Balances["SMS1"].Units)
		}
		if acnt.Balances["DATA1"].Units.Compare(utils.NewDecimalFromFloat64(600)) != 0 {
			t.Errorf("expected 600 *data units remaining, got: %+v", acnt.Balances["DATA1"].Units)
		}
	}
}
