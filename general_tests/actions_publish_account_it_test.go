//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	actionsPublishAcntCfgPath string
	actionsPublishAcntCfg     *config.CGRConfig
	actionsPublishAcntRpc     *birpc.Client
	actionsPublishAcntConfDIR string
	actionsPublishAcntDelay   int

	sTestsActionsPublishAcnt = []func(t *testing.T){
		testActionsPublishAcntLoadConfig,
		testActionsPublishAcntInitDataDb,
		testActionsPublishAcntResetStorDb,
		testActionsPublishAcntStartEngine,
		testActionsPublishAcntRpcConn,
		testActionsPublishAcntSetActions,
		testActionsPublishAcntSetThresholdProfiles,
		testActionsPublishAcntTopup,
		testActionsPublishAcntTopupReset,
		testActionsPublishAcntStopEngine,
	}
)

func TestActionsPublishAccountIT(t *testing.T) {

	switch *utils.DBType {
	case utils.MetaInternal:
		actionsPublishAcntConfDIR = "tutinternal"
	case utils.MetaMySQL:
		actionsPublishAcntConfDIR = "tutmysql"
	case utils.MetaMongo:
		actionsPublishAcntConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsActionsPublishAcnt {
		t.Run(actionsPublishAcntConfDIR, stest)
	}
}

func testActionsPublishAcntLoadConfig(t *testing.T) {
	var err error
	actionsPublishAcntCfgPath = path.Join(*utils.DataDir, "conf", "samples", actionsPublishAcntConfDIR)
	if actionsPublishAcntCfg, err = config.NewCGRConfigFromPath(actionsPublishAcntCfgPath); err != nil {
		t.Error(err)
	}
	actionsPublishAcntDelay = 1000
}

func testActionsPublishAcntInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(actionsPublishAcntCfg); err != nil {
		t.Fatal(err)
	}
}

func testActionsPublishAcntResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(actionsPublishAcntCfg); err != nil {
		t.Fatal(err)
	}
}

func testActionsPublishAcntStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actionsPublishAcntCfgPath, actionsPublishAcntDelay); err != nil {
		t.Fatal(err)
	}
}

func testActionsPublishAcntRpcConn(t *testing.T) {
	actionsPublishAcntRpc = engine.NewRPCClient(t, actionsPublishAcntCfg.ListenCfg())
}

func testActionsPublishAcntSetActions(t *testing.T) {
	var reply string
	attrs := &v1.V1AttrSetActions{
		ActionsId: "ACT_LOG",
		Overwrite: true,
		Actions: []*v1.V1TPAction{
			{Identifier: utils.MetaLog, Weight: 10},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}

	attrs = &v1.V1AttrSetActions{
		ActionsId: "ACT_TOPUP",
		Overwrite: true,
		Actions: []*v1.V1TPAction{
			{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, Units: 10, Weight: 10},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}

	attrs = &v1.V1AttrSetActions{
		ActionsId: "ACT_TOPUP_RESET",
		Overwrite: true,
		Actions: []*v1.V1TPAction{
			{Identifier: utils.MetaTopUpReset, BalanceType: utils.MetaMonetary, Units: 10, Weight: 10},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}
}

func testActionsPublishAcntSetThresholdProfiles(t *testing.T) {
	var reply string
	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_PUBLISH_1001",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1001",
			},
			MaxHits:   -1,
			MinHits:   1,
			Async:     true,
			ActionIDs: []string{"ACT_LOG"},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thPrf, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}

	thPrf = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_PUBLISH_1002",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1002",
			},
			MaxHits:   -1,
			MinHits:   1,
			Async:     true,
			ActionIDs: []string{"ACT_LOG"},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thPrf, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}
}

func testActionsPublishAcntTopup(t *testing.T) {
	attrs := &utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1001",
		ActionsId: "ACT_TOPUP",
	}
	var reply string
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	var th *engine.Threshold
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_1001"},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, args, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1001 Hits 1 after *topup, received %d", th.Hits)
	}
}

func testActionsPublishAcntTopupReset(t *testing.T) {
	attrs := &utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1002",
		ActionsId: "ACT_TOPUP_RESET",
	}
	var reply string
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	var th *engine.Threshold
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_1002"},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, args, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1002 Hits 1 after *topup_reset, received %d", th.Hits)
	}
}

func testActionsPublishAcntStopEngine(t *testing.T) {
	if err := engine.KillEngine(actionsPublishAcntDelay); err != nil {
		t.Error(err)
	}
}

func TestActionsPublishAccountThresholds(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaPostgres, utils.MetaMongo, utils.MetaMySQL:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true,
	"thresholds_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"thresholds": {
	"enabled": true,
	"indexed_selects": false,
	"store_interval": "-1"
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("SetActions", func(t *testing.T) {
		var reply string
		for _, attrs := range []*v1.V1AttrSetActions{
			{
				ActionsId: "ACT_LOG",
				Overwrite: true,
				Actions:   []*v1.V1TPAction{{Identifier: utils.MetaLog, Weight: 10}},
			},
			{
				ActionsId: "ACT_TOPUP",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_TOPUP", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_SET_BALANCE",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaSetBalance, BalanceType: utils.MetaMonetary, BalanceId: "BAL_SET", Units: 20, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_TOPUP_REM",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_REM", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_REMOVE_BALANCE",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaRemoveBalance, BalanceType: utils.MetaMonetary, BalanceId: "BAL_REM", Weight: 10},
				},
			},
			{
				ActionsId: "ACT_TOPUP_DEBIT",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_DEBIT", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_DEBIT_RESET",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaDebitReset, BalanceType: utils.MetaMonetary, BalanceId: "BAL_DEBIT", Units: 3, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_TOPUP_PUB",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_PUB", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_PUBLISH_ACCOUNT",
				Overwrite: true,
				Actions:   []*v1.V1TPAction{{Identifier: utils.MetaPublishAccount, Weight: 10}},
			},
		} {
			if err := client.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
				t.Fatal(err)
			} else if reply != utils.OK {
				t.Fatalf("unexpected reply: %s", reply)
			}
		}
	})

	t.Run("SetThresholdProfiles", func(t *testing.T) {
		var reply string
		for _, thPrf := range []*engine.ThresholdProfileWithAPIOpts{
			{ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_TOPUP",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1001",
					"*gte:~*asm.BalanceSummaries.BAL_TOPUP.Value:10",
				},
				MaxHits: -1, MinHits: 1, Async: true, ActionIDs: []string{"ACT_LOG"},
			}},
			{ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_SET_BALANCE",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1002",
					"*gte:~*asm.BalanceSummaries.BAL_SET.Value:20",
				},
				MaxHits: -1, MinHits: 1, Async: true, ActionIDs: []string{"ACT_LOG"},
			}},
			{ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_REMOVE_BALANCE",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1003",
					"*notexists:~*asm.BalanceSummaries.BAL_REM.Value:",
				},
				MaxHits: -1, MinHits: 1, Async: true, ActionIDs: []string{"ACT_LOG"},
			}},
			{ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_DEBIT_RESET",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1004",
					"*lt:~*asm.BalanceSummaries.BAL_DEBIT.Value:10",
				},
				MaxHits: -1, MinHits: 1, Async: true, ActionIDs: []string{"ACT_LOG"},
			}},
			{ThresholdProfile: &engine.ThresholdProfile{
				Tenant: "cgrates.org",
				ID:     "THD_PUBLISH_ACCOUNT",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1005",
					"*gte:~*asm.BalanceSummaries.BAL_PUB.Value:10",
				},
				MaxHits: -1, MinHits: 1, Async: true, ActionIDs: []string{"ACT_LOG"},
			}},
		} {
			if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile, thPrf, &reply); err != nil {
				t.Fatal(err)
			} else if reply != utils.OK {
				t.Fatalf("unexpected reply: %s", reply)
			}
		}
	})

	t.Run("Topup", func(t *testing.T) {
		var reply string
		var th *engine.Threshold
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_TOPUP"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 1 {
			t.Errorf("expected THD_TOPUP Hits 1 after ACT_TOPUP, received %d", th.Hits)
		}
	})

	t.Run("SetBalance", func(t *testing.T) {
		var reply string
		var th *engine.Threshold
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1002", ActionsId: "ACT_SET_BALANCE"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_SET_BALANCE"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 1 {
			t.Errorf("expected THD_SET_BALANCE Hits 1 after ACT_SET_BALANCE, received %d", th.Hits)
		}
	})

	t.Run("RemoveBalance", func(t *testing.T) {
		var reply string
		var th *engine.Threshold
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1003", ActionsId: "ACT_TOPUP_REM"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_REMOVE_BALANCE"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 0 {
			t.Errorf("expected THD_REMOVE_BALANCE Hits 0 after ACT_TOPUP_REM, received %d", th.Hits)
		}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1003", ActionsId: "ACT_REMOVE_BALANCE"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_REMOVE_BALANCE"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 1 {
			t.Errorf("expected THD_REMOVE_BALANCE Hits 1 after ACT_REMOVE_BALANCE, received %d", th.Hits)
		}
	})

	t.Run("DebitReset", func(t *testing.T) {
		var reply string
		var th *engine.Threshold
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1004", ActionsId: "ACT_TOPUP_DEBIT"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_DEBIT_RESET"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 0 {
			t.Errorf("expected THD_DEBIT_RESET Hits 0 after ACT_TOPUP_DEBIT, received %d", th.Hits)
		}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1004", ActionsId: "ACT_DEBIT_RESET"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_DEBIT_RESET"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 1 {
			t.Errorf("expected THD_DEBIT_RESET Hits 1 after ACT_DEBIT_RESET, received %d", th.Hits)
		}
	})

	t.Run("PublishAccountNoDoublePublish", func(t *testing.T) {
		var reply string
		var th *engine.Threshold
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1005", ActionsId: "ACT_TOPUP_PUB"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_ACCOUNT"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 1 {
			t.Errorf("expected THD_PUBLISH_ACCOUNT Hits 1 after ACT_TOPUP_PUB, received %d", th.Hits)
		}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1005", ActionsId: "ACT_PUBLISH_ACCOUNT"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_ACCOUNT"}}, &th); err != nil {
			t.Fatal(err)
		} else if th.Hits != 2 {
			t.Errorf("expected THD_PUBLISH_ACCOUNT Hits 2 after ACT_PUBLISH_ACCOUNT, received %d", th.Hits)
		}
	})
}

func TestActionsPublishAccountStats(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaPostgres, utils.MetaMongo, utils.MetaMySQL:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true,
	"stats_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	t.Run("SetActions", func(t *testing.T) {
		var reply string
		for _, attrs := range []*v1.V1AttrSetActions{
			{
				ActionsId: "ACT_TOPUP",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_TOPUP", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_SET_BALANCE",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaSetBalance, BalanceType: utils.MetaMonetary, BalanceId: "BAL_SET", Units: 20, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_TOPUP_DEBIT",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_DEBIT", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_DEBIT_RESET",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaDebitReset, BalanceType: utils.MetaMonetary, BalanceId: "BAL_DEBIT", Units: 3, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_TOPUP_PUB",
				Overwrite: true,
				Actions: []*v1.V1TPAction{
					{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_PUB", Units: 10, Weight: 10},
				},
			},
			{
				ActionsId: "ACT_PUBLISH_ACCOUNT",
				Overwrite: true,
				Actions:   []*v1.V1TPAction{{Identifier: utils.MetaPublishAccount, Weight: 10}},
			},
		} {
			if err := client.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
				t.Fatal(err)
			} else if reply != utils.OK {
				t.Fatalf("unexpected reply: %s", reply)
			}
		}
	})

	t.Run("SetStatQueueProfiles", func(t *testing.T) {
		var reply string
		for _, sqPrf := range []*engine.StatQueueProfileWithAPIOpts{
			{StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "STS_TOPUP",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1001",
				},
				QueueLength: 100,
				TTL:         -1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum#~*asm.BalanceSummaries.BAL_TOPUP.Value"}},
			}},
			{StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "STS_SET_BALANCE",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1002",
				},
				QueueLength: 100,
				TTL:         -1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum#~*asm.BalanceSummaries.BAL_SET.Value"}},
			}},
			{StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "STS_DEBIT_RESET",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1004",
				},
				QueueLength: 100,
				TTL:         -1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum#~*asm.BalanceSummaries.BAL_DEBIT.Value"}},
			}},
			{StatQueueProfile: &engine.StatQueueProfile{
				Tenant: "cgrates.org",
				ID:     "STS_PUBLISH_ACCOUNT",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1005",
				},
				QueueLength: 100,
				TTL:         -1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum#~*asm.BalanceSummaries.BAL_PUB.Value"}},
			}},
		} {
			if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, sqPrf, &reply); err != nil {
				t.Fatal(err)
			} else if reply != utils.OK {
				t.Fatalf("unexpected reply: %s", reply)
			}
		}
	})

	t.Run("Topup", func(t *testing.T) {
		var reply string
		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1001", ActionsId: "ACT_TOPUP"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_TOPUP"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_TOPUP.Value"]; got != 10 {
			t.Errorf("expected STS_TOPUP sum 10 after *topup, received %v (%v)", got, metrics)
		}
	})

	t.Run("SetBalance", func(t *testing.T) {
		var reply string
		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1002", ActionsId: "ACT_SET_BALANCE"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_SET_BALANCE"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_SET.Value"]; got != 20 {
			t.Errorf("expected STS_SET_BALANCE sum 20 after *set_balance, received %v (%v)", got, metrics)
		}
	})

	t.Run("DebitReset", func(t *testing.T) {
		var reply string
		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1004", ActionsId: "ACT_TOPUP_DEBIT"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_DEBIT_RESET"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_DEBIT.Value"]; got != 10 {
			t.Fatalf("expected STS_DEBIT_RESET sum 10 after *topup, received %v (%v)", got, metrics)
		}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1004", ActionsId: "ACT_DEBIT_RESET"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_DEBIT_RESET"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_DEBIT.Value"]; got == 10 {
			t.Errorf("expected STS_DEBIT_RESET sum to change after *debit_reset, received %v (%v)", got, metrics)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_DEBIT.Value"]; got != 7 {
			t.Errorf("expected STS_DEBIT_RESET sum 7 after *debit_reset (10 + (-3)), received %v (%v)", got, metrics)
		}
	})

	t.Run("PublishAccountNoDoublePublish", func(t *testing.T) {
		var reply string
		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1005", ActionsId: "ACT_TOPUP_PUB"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_PUBLISH_ACCOUNT"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_PUB.Value"]; got != 10 {
			t.Fatalf("expected STS_PUBLISH_ACCOUNT sum 10 after *topup, received %v (%v)", got, metrics)
		}
		if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
			&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1005", ActionsId: "ACT_PUBLISH_ACCOUNT"}, &reply); err != nil {
			t.Fatal(err)
		}
		time.Sleep(500 * time.Millisecond)
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_PUBLISH_ACCOUNT"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_PUB.Value"]; got != 20 {
			t.Errorf("expected STS_PUBLISH_ACCOUNT sum 20 after *publish_account, received %v (%v)", got, metrics)
		}
	})
}

func TestActionsPublishAccountTriggerStats(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaPostgres, utils.MetaMongo, utils.MetaMySQL:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"general": {
	"log_level": 7
},
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true,
	"stats_conns": ["*internal"]
},
"schedulers": {
	"enabled": true
},
"stats": {
	"enabled": true,
	"store_interval": "-1"
},
"apiers": {
	"enabled": true,
	"scheduler_conns": ["*internal"]
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
	}
	client, _ := ng.Run(t)

	var reply string
	for _, attrs := range []*v1.V1AttrSetActions{
		{
			ActionsId: "ACT_BONUS",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_BONUS", Units: 10, Weight: 10},
			},
		},
		{
			ActionsId: "ACT_TOPUP_60",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 60, Weight: 10},
			},
		},
	} {
		if err := client.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
			t.Fatal(err)
		} else if reply != utils.OK {
			t.Fatalf("unexpected reply: %s", reply)
		}
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile,
		&engine.StatQueueProfileWithAPIOpts{StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "STS_ACTION_TRIGGER",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1009",
			},
			QueueLength: 100,
			TTL:         -1,
			Metrics: []*engine.MetricWithFilters{
				{MetricID: "*sum#~*asm.BalanceSummaries.BAL_BONUS.Value"},
				{MetricID: "*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"},
			},
		}}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		&utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1009",
			BalanceType: utils.MetaMonetary,
			Value:       10,
			Balance: map[string]any{
				utils.ID: "BAL_MAIN",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetAccountActionTriggers,
		&v1.AttrSetAccountActionTriggers{
			Tenant:  "cgrates.org",
			Account: "1009",
			AttrSetActionTrigger: engine.AttrSetActionTrigger{
				GroupID: "TRG_MAIN_HIGH",
				ActionTrigger: map[string]any{
					utils.ThresholdType:  "*max_balance",
					utils.ThresholdValue: 50,
					utils.BalanceType:    utils.MetaMonetary,
					utils.BalanceID:      "BAL_MAIN",
					utils.ActionsID:      "ACT_BONUS",
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	var before *engine.Account
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1009"}, &before); err != nil {
		t.Fatal(err)
	}
	for _, b := range before.BalanceMap[utils.MetaMonetary] {
		if b.ID == "BAL_BONUS" {
			t.Fatalf("trigger fired at creation, BAL_BONUS %v", b.Value)
		}
	}

	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1009", ActionsId: "ACT_TOPUP_60"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	var acnt *engine.Account
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1009"}, &acnt); err != nil {
		t.Fatal(err)
	}
	var main, bonus float64
	for _, b := range acnt.BalanceMap[utils.MetaMonetary] {
		switch b.ID {
		case "BAL_MAIN":
			main = b.Value
		case "BAL_BONUS":
			bonus = b.Value
		}
	}
	if main != 70 || bonus != 10 {
		t.Fatalf("expected BAL_MAIN 70 and BAL_BONUS 10 after the trigger fired, received %v and %v", main, bonus)
	}

	var metrics map[string]float64
	if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_ACTION_TRIGGER"}}, &metrics); err != nil {
		t.Fatal(err)
	}
	got := metrics["*sum#~*asm.BalanceSummaries.BAL_BONUS.Value"]
	if got != 10 {
		t.Errorf("expected one published summary with BAL_BONUS (sum 10; 0 = never published, 20 = published twice), received %v (%v)", got, metrics)
	}
	if got := metrics["*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"]; got != 70 {
		t.Errorf("expected one published summary with BAL_MAIN 70, received %v (%v)", got, metrics)
	}
}
