//go:build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// this tests db replication where both engines are active checks scheduled actions are filtered based on engine node id
func TestDBReplicationActiveActive(t *testing.T) {
	var dbcfg engine.DBCfg
	var dbcfg2 engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbcfg = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaInternal),
						Opts: engine.DBConnOpts{
							InternalDBDumpInterval:    utils.StringPointer("0s"),
							InternalDBRewriteInterval: utils.StringPointer("0s"),
						},
					},
				},
			},
		}
		dbcfg2 = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaInternal),
						Opts: engine.DBConnOpts{
							InternalDBDumpInterval:    utils.StringPointer("0s"),
							InternalDBRewriteInterval: utils.StringPointer("0s"),
						},
					},
				},
			},
		}
	case utils.MetaMySQL:
		dbcfg = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type:     utils.StringPointer(utils.MetaMySQL),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(3306),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
			},
		}
		dbcfg2 = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type:     utils.StringPointer(utils.MetaMySQL),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(3306),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
			},
		}
	case utils.Redis:
		dbcfg = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaRedis),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(6379),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
				},
			},
		}
		dbcfg2 = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaRedis),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(6379),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
				},
			},
		}
	case utils.MetaMongo:
		dbcfg = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaMongo),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(27017),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
					utils.StorDB: {
						Type:     utils.StringPointer(utils.MetaMongo),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(27017),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer(""),
					},
				},
				Items: map[string]engine.Item{
					utils.MetaCDRs: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
				},
			},
		}
		dbcfg2 = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type: utils.StringPointer(utils.MetaMongo),
						Host: utils.StringPointer("127.0.0.1"),
						Port: utils.IntPointer(27017),
						Name: utils.StringPointer("10"),
						User: utils.StringPointer(utils.CGRateSLwr),
					},
					utils.StorDB: {
						Type:     utils.StringPointer(utils.MetaMongo),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(27017),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer(""),
					},
				},
				Items: map[string]engine.Item{
					utils.MetaCDRs: {
						Limit:  utils.IntPointer(-1),
						DbConn: utils.StringPointer(utils.StorDB),
					},
				},
			},
		}
	case utils.MetaPostgres:
		dbcfg = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type:     utils.StringPointer(utils.MetaPostgres),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(5432),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
			},
		}
		dbcfg2 = engine.DBCfg{
			DB: &engine.DBParams{
				DBConns: map[string]engine.DBConn{
					utils.MetaDefault: {
						Type:     utils.StringPointer(utils.MetaPostgres),
						Host:     utils.StringPointer("127.0.0.1"),
						Port:     utils.IntPointer(5432),
						Name:     utils.StringPointer(utils.CGRateSLwr),
						User:     utils.StringPointer(utils.CGRateSLwr),
						Password: utils.StringPointer("CGRateS.org"),
					},
				},
			},
		}
	default:
		t.Fatal("unsupported dbtype value")
	}

	failedDirNG1 := t.TempDir()

	primaryCfg := `{
"general": {
	"nodeID": "ng1",
},
"logger": {
	"level": 7,
},
"listen": {
	"rpcJSON": ":4112",
	"rpcGOB": ":4113",
	"http": ":4180",
},
"rpcConns": {
	"rpl": {
		"conns": [{
				"address": "127.0.0.1:4123",
				"transport": "*gob"
		}]
	},
	"localhostCustom": {
		"conns": [{
			"address": "127.0.0.1:4112",
		}]
	}
},
"accounts": {
	"enabled": true,
},
  "actions": {
    "enabled": true,
    "conns": {
      "*accounts": [
        {
          "connIDs": [
            "*internal"
          ]
        }
      ],
    }
  },
"admins": {
	"enabled": true,
	"conns": {"*actions": [{"connIDs": ["*internal"]}]},
}
}`

	failedDirNG2 := t.TempDir()
	targetCfg := `{
"general": {
	"nodeID": "ng2",
},
"logger": {
	"level": 7,
},
"listen": {
	"rpcJSON": ":4122",
	"rpcGOB": ":4123",
	"http": ":4190",
},
"rpcConns": {
	"rpl": {
		"conns": [
			{
				"address": "127.0.0.1:4113",
				"transport": "*gob"
			}
		]
	},
	"localhostCustom": {
		"conns": [{
			"address": "127.0.0.1:4122",
		}]
	}
},
"accounts": {
	"enabled": true,
},
  "actions": {
    "enabled": true,
    "conns": {
      "*accounts": [
        {
          "connIDs": [
            "localhostCustom"
          ]
        }
      ],
    }
  },
"admins": {
	"enabled": true,
	"conns": {"*actions": [{"connIDs": ["localhostCustom"]}]},
}
}`
	db := dbcfg.DB.DBConns[utils.MetaDefault]
	db.ReplicationConns = utils.SliceStringPointer([]string{"rpl"})
	db.ReplicationFailedDir = utils.StringPointer(failedDirNG1)
	dbcfg.DB.DBConns[utils.MetaDefault] = db
	dbcfg.DB.Items = map[string]engine.Item{
		"*accountFilterIndexes":       {Replicate: utils.BoolPointer(true)},
		"*accounts":                   {Replicate: utils.BoolPointer(true)},
		"*actionProfileFilterIndexes": {Replicate: utils.BoolPointer(true)},
		"*actionProfiles":             {Replicate: utils.BoolPointer(true)},
		"*attributeFilterIndexes":     {Replicate: utils.BoolPointer(true)},
		"*attributeProfiles":          {Replicate: utils.BoolPointer(true)},
		"*cdrs":                       {Replicate: utils.BoolPointer(true)},
		"*chargerFilterIndexes":       {Replicate: utils.BoolPointer(true)},
		"*chargerProfiles":            {Replicate: utils.BoolPointer(true)},
		"*filters":                    {Replicate: utils.BoolPointer(true)},
		"*ipAllocations":              {Replicate: utils.BoolPointer(true)},
		"*ipFilterIndexes":            {Replicate: utils.BoolPointer(true)},
		"*ipProfiles":                 {Replicate: utils.BoolPointer(true)},
		"*loadIDs":                    {Replicate: utils.BoolPointer(true)},
		"*rankingProfiles":            {Replicate: utils.BoolPointer(true)},
		"*rankings":                   {Replicate: utils.BoolPointer(true)},
		"*rateFilterIndexes":          {Replicate: utils.BoolPointer(true)},
		"*rateProfileFilterIndexes":   {Replicate: utils.BoolPointer(true)},
		"*rateProfiles":               {Replicate: utils.BoolPointer(true)},
		"*resourceFilterIndexes":      {Replicate: utils.BoolPointer(true)},
		"*resourceProfiles":           {Replicate: utils.BoolPointer(true)},
		"*resources":                  {Replicate: utils.BoolPointer(true)},
		"*reverseFilterIndexes":       {Replicate: utils.BoolPointer(true)},
		"*routeFilterIndexes":         {Replicate: utils.BoolPointer(true)},
		"*routeProfiles":              {Replicate: utils.BoolPointer(true)},
		"*statFilterIndexes":          {Replicate: utils.BoolPointer(true)},
		"*statQueueProfiles":          {Replicate: utils.BoolPointer(true)},
		"*statQueues":                 {Replicate: utils.BoolPointer(true)},
		"*thresholdFilterIndexes":     {Replicate: utils.BoolPointer(true)},
		"*thresholdProfiles":          {Replicate: utils.BoolPointer(true)},
		"*thresholds":                 {Replicate: utils.BoolPointer(true)},
		"*trendProfiles":              {Replicate: utils.BoolPointer(true)},
		"*trends":                     {Replicate: utils.BoolPointer(true)},
		"*versions":                   {Replicate: utils.BoolPointer(true)},
	}
	ng1 := engine.TestEngine{
		DBCfg:      dbcfg,
		ConfigJSON: primaryCfg,
	}
	db2 := dbcfg.DB.DBConns[utils.MetaDefault]
	db2.ReplicationConns = utils.SliceStringPointer([]string{"rpl"})
	db2.ReplicationFailedDir = utils.StringPointer(failedDirNG2)
	dbcfg2.DB.DBConns[utils.MetaDefault] = db2
	dbcfg2.DB.Items = dbcfg.DB.Items
	ng2 := engine.TestEngine{
		DBCfg:      dbcfg2,
		ConfigJSON: targetCfg,
	}

	client1, _ := ng1.Run(t)
	time.Sleep(200 * time.Millisecond)

	client2, _ := ng2.Run(t)
	time.Sleep(200 * time.Millisecond)

	var acc1 utils.Account
	var acc2 utils.Account
	t.Run("Engine1SetData", func(t *testing.T) {
		if err := client1.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					ID: "1001",
				},
			}, new(string)); err != nil {
			t.Fatalf("AdminSv1SetAccount unexpected err: %v", err)
		}
		var reply string
		topupActionASAP := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				ID:        "TOPUP_ACTION_ASAP",
				FilterIDs: []string{"*string:~*cfg.general.nodeID:ng1"},
				Schedule:  utils.MetaASAP,
				Targets: map[string]utils.StringSet{
					utils.MetaAccounts: {"1001": struct{}{}},
				},
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						Weight: 21,
					},
				},
				Actions: []*utils.APAction{
					{
						ID:   "SET_BAL",
						Type: utils.MetaSetBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID1.Units",
									"*balanceValue": "5",
								},
							},
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID1.Type",
									"*balanceValue": "*monetary",
								},
							},
						},
					},
				},
			},
		}

		if err := client1.Call(context.Background(), utils.AdminSv1SetActionProfile, topupActionASAP, &reply); err != nil {
			t.Fatal(err)
		}
		topupActionRecurring1s := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				ID:        "TOPUP_ACTION",
				FilterIDs: []string{"*string:~*cfg.general.nodeID:ng1"},
				Schedule:  "@every 2s",
				Targets: map[string]utils.StringSet{
					utils.MetaAccounts: {"1001": struct{}{}},
				},
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						Weight: 20,
					},
				},
				Actions: []*utils.APAction{
					{
						ID:   "SET_BAL",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID1.Units",
									"*balanceValue": "23",
								},
							},
						},
					},
				},
			},
		}

		if err := client1.Call(context.Background(), utils.AdminSv1SetActionProfile, topupActionRecurring1s, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client1.Call(context.Background(), utils.ActionSv1ScheduleActions,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "EventExecuteActions",
				Event:  map[string]any{},
				APIOpts: map[string]any{
					utils.OptsActionsProfileIDs: []string{"TOPUP_ACTION_ASAP", "TOPUP_ACTION"},
				},
			}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond) // wait for ASAP execution
		if err := client1.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: "1001"},
			}, &acc1); err != nil {
			t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
		}
		// t.Run("GetSchedulers", func(t *testing.T) {
		// 	// engine 2 shouldnt have schdeduled actions since filtered by nodeID
		// 	var rply []*scheduler.ScheduledAction
		// 	if err := client1.Call(context.Background(), utils.AdminSv1GetScheduledActions,
		// 		scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
		// 		t.Error("Unexpected error: ", err)
		// 	} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
		// 		t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
		// 	}
		// 	if err := client2.Call(context.Background(), utils.AdminSv1GetScheduledActions,
		// 		scheduler.ArgsGetScheduledActions{}, &rply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		// 		t.Error("Unexpected error: ", err)
		// 	}
		// })
	})

	t.Run("Engine2GetData", func(t *testing.T) {
		var acnt utils.Account
		if err := client2.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: "1001"},
			}, &acnt); err != nil {
			t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
		}
		if !reflect.DeepEqual(acnt, acc1) {
			t.Errorf("expected <%#v>, \ngot <%#v>", acc1, acnt)
			t.Errorf("expected <%v>, \ngot <%v>", acc1.Balances["BalID1"].String(), acnt.Balances["BalID1"].String())
		} else if i := acnt.Balances["BalID1"].Units.Compare(utils.NewDecimalFromFloat64(5)); i != 0 {
			t.Errorf("expected 5 units, \ngot <%v>", acnt.Balances["BalID1"].String())
		}
	})

	t.Run("Engine2SetData", func(t *testing.T) {
		if err := client2.Call(context.Background(), utils.AdminSv1SetAccount,
			&utils.AccountWithAPIOpts{
				Account: &utils.Account{
					ID: "1002",
				},
			}, new(string)); err != nil {
			t.Fatalf("AdminSv1SetAccount unexpected err: %v", err)
		}
		var reply string
		topupActionASAP := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				ID:        "TOPUP_ACTION_ASAP2",
				FilterIDs: []string{"*string:~*cfg.general.nodeID:ng2"},
				Schedule:  utils.MetaASAP,
				Targets: map[string]utils.StringSet{
					utils.MetaAccounts: {"1002": struct{}{}},
				},
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						Weight: 21,
					},
				},
				Actions: []*utils.APAction{
					{
						ID:   "SET_BAL",
						Type: utils.MetaSetBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID2.Units",
									"*balanceValue": "5",
								},
							},
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID2.Type",
									"*balanceValue": "*monetary",
								},
							},
						},
					},
				},
			},
		}

		if err := client2.Call(context.Background(), utils.AdminSv1SetActionProfile, topupActionASAP, &reply); err != nil {
			t.Fatal(err)
		}
		topupActionRecurring1s := &utils.ActionProfileWithAPIOpts{
			ActionProfile: &utils.ActionProfile{
				ID:        "TOPUP_ACTION2",
				FilterIDs: []string{"*string:~*cfg.general.nodeID:ng2"},
				Schedule:  "@every 2s",
				Targets: map[string]utils.StringSet{
					utils.MetaAccounts: {"1002": struct{}{}},
				},
				Weights: utils.DynamicWeights{
					&utils.DynamicWeight{
						Weight: 20,
					},
				},
				Actions: []*utils.APAction{
					{
						ID:   "SET_BAL",
						Type: utils.MetaAddBalance,
						Diktats: []*utils.APDiktat{
							{
								ID: "SetBalUnits",
								Opts: map[string]any{
									"*balancePath":  "*balance.BalID2.Units",
									"*balanceValue": "23",
								},
							},
						},
					},
				},
			},
		}

		if err := client2.Call(context.Background(), utils.AdminSv1SetActionProfile, topupActionRecurring1s, &reply); err != nil {
			t.Fatal(err)
		}
		if err := client2.Call(context.Background(), utils.ActionSv1ScheduleActions,
			&utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "EventExecuteActions",
				Event:  map[string]any{},
				APIOpts: map[string]any{
					utils.OptsActionsProfileIDs: []string{"TOPUP_ACTION_ASAP2", "TOPUP_ACTION2"},
				},
			}, &reply); err != nil {
			t.Error(err)
		}
		time.Sleep(100 * time.Millisecond) // wait for ASAP execution
		if err := client2.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: "1002"},
			}, &acc2); err != nil {
			t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
		}

	})

	t.Run("Engine1GetData", func(t *testing.T) {
		var acnt utils.Account
		if err := client2.Call(context.Background(), utils.AdminSv1GetAccount,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{ID: "1002"},
			}, &acnt); err != nil {
			t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
		}
		if !reflect.DeepEqual(acnt, acc2) {
			t.Errorf("expected <%#v>, \ngot <%#v>", acc2, acnt)
			t.Errorf("expected <%v>, \ngot <%v>", acc2.Balances["BalID2"].String(), acnt.Balances["BalID2"].String())
		} else if i := acnt.Balances["BalID2"].Units.Compare(utils.NewDecimalFromFloat64(5)); i != 0 {
			t.Errorf("expected 5 units, \ngot <%v>", acnt.Balances["BalID2"].String())
		}
	})

	// t.Run("GetSchedulers", func(t *testing.T) {
	// 	var rply []*scheduler.ScheduledAction
	// 	if err := client1.Call(context.Background(), utils.AdminSv1GetScheduledActions,
	// 		scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
	// 		t.Error("Unexpected error: ", err)
	// 	} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
	// 		t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
	// 	}
	// 	if err := client2.Call(context.Background(), utils.AdminSv1GetScheduledActions,
	// 		scheduler.ArgsGetScheduledActions{}, &rply); err != nil {
	// 		t.Error("Unexpected error: ", err)
	// 	} else if len(rply) != 1 && rply[0].Accounts != 1 && rply[0].ActionPlanID != "APRecurring" && rply[0].ActionsID != "TOPUP_ACTION" {
	// 		t.Errorf("ScheduledActions: %+v", utils.ToJSON(rply))
	// 	}
	// })

	entries, err := os.ReadDir(failedDirNG1)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected failed dir to be empty, found %d entries", len(entries))
	}

	entries2, err := os.ReadDir(failedDirNG2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries2) != 0 {
		t.Fatal("expected 0 .gob files in failed dir, received", len(entries2))
	}

	time.Sleep(4 * time.Second)
	if err := client1.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "1001"},
		}, &acc1); err != nil {
		t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
	}
	if err := client1.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "1002"},
		}, &acc2); err != nil {
		t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
	}
	if i := acc1.Balances["BalID1"].Units.Compare(utils.NewDecimalFromFloat64(51)); i != 0 {
		t.Errorf("expected 51 units, \ngot <%v>", acc1.Balances["BalID1"].String())
	} else if i := acc2.Balances["BalID2"].Units.Compare(utils.NewDecimalFromFloat64(51)); i != 0 {
		t.Errorf("expected 51 units, \ngot <%v>", acc2.Balances["BalID2"].String())
	}
	if err := client2.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "1001"},
		}, &acc1); err != nil {
		t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
	}
	if err := client2.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{ID: "1002"},
		}, &acc2); err != nil {
		t.Fatalf("AdminSv1GetAccount unexpected err: %v", err)
	}
	if i := acc1.Balances["BalID1"].Units.Compare(utils.NewDecimalFromFloat64(51)); i != 0 {
		t.Errorf("expected 51 units, \ngot <%v>", acc1.Balances["BalID1"].String())
	} else if i := acc2.Balances["BalID2"].Units.Compare(utils.NewDecimalFromFloat64(51)); i != 0 {
		t.Errorf("expected 51 units, \ngot <%v>", acc2.Balances["BalID2"].String())
	}
}
