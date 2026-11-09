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
		testActionsPublishAcntNoRepublish,
		testActionsPublishAcntSameSet,
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
func testActionsPublishAcntNoRepublish(t *testing.T) {
	var reply string
	attrs := &v1.V1AttrSetActions{
		ActionsId: "ACT_RESET_TRIGGERS",
		Overwrite: true,
		Actions: []*v1.V1TPAction{
			{Identifier: utils.MetaResetTriggers, Weight: 10},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_PUBLISH_1003",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1003",
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

	var th *engine.Threshold
	thArgs := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_1003"},
	}

	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1003", ActionsId: "ACT_TOPUP"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, thArgs, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1003 Hits 1 after ACT_TOPUP, received %d", th.Hits)
	}

	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1003", ActionsId: "ACT_LOG"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, thArgs, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1003 Hits 1 after ACT_LOG, received %d", th.Hits)
	}

	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1003", ActionsId: "ACT_RESET_TRIGGERS"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, thArgs, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1003 Hits 1 after ACT_RESET_TRIGGERS, received %d", th.Hits)
	}
}

func testActionsPublishAcntSameSet(t *testing.T) {
	var reply string

	attrs := &v1.V1AttrSetActions{
		ActionsId: "ACT_TOPUP_AND_RESET",
		Overwrite: true,
		Actions: []*v1.V1TPAction{
			{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, Units: 10, Weight: 20},
			{Identifier: utils.MetaResetTriggers, Weight: 10},
		},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("unexpected reply: %s", reply)
	}

	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_PUBLISH_1004",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1004",
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

	execAttrs := &utils.AttrExecuteAction{
		Tenant:    "cgrates.org",
		Account:   "1004",
		ActionsId: "ACT_TOPUP_AND_RESET",
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.APIerSv1ExecuteAction, execAttrs, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	var th *engine.Threshold
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_PUBLISH_1004"},
	}
	if err := actionsPublishAcntRpc.Call(context.Background(), utils.ThresholdSv1GetThreshold, args, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected THD_PUBLISH_1004 Hits 1 for [*topup,*reset_triggers] in one set, received %d", th.Hits)
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
				ID:     "STS_TOPUP_INITIAL",
				FilterIDs: []string{
					"*string:~*opts.*eventType:AccountUpdate",
					"*string:~*asm.AccountID:1001",
				},
				QueueLength: 100,
				TTL:         -1,
				Metrics:     []*engine.MetricWithFilters{{MetricID: "*sum#~*asm.BalanceSummaries.BAL_TOPUP.Initial"}},
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

	t.Run("TopUpInitial", func(t *testing.T) {
		var metrics map[string]float64
		if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_TOPUP_INITIAL"}}, &metrics); err != nil {
			t.Fatal(err)
		} else if got := metrics["*sum#~*asm.BalanceSummaries.BAL_TOPUP.Initial"]; got != 10 {
			t.Errorf("expected STS_TOPUP_INITIAL sum 10 after *topup (0 = Initial not filled), received %v (%v)", got, metrics)
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

func TestActionsPublishAccountBonusTriggerStats(t *testing.T) {
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

	var acntBefore *engine.Account
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1009"}, &acntBefore); err != nil {
		t.Fatal(err)
	}
	for _, b := range acntBefore.BalanceMap[utils.MetaMonetary] {
		if b.ID == "BAL_BONUS" {
			t.Fatalf("trigger fired at creation, BAL_BONUS %v", b.Value)
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
			QueueLength: 1,
			TTL:         -1,
			Metrics: []*engine.MetricWithFilters{
				{MetricID: "*sum#~*asm.BalanceSummaries.BAL_BONUS.Value"},
				{MetricID: "*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"},
			},
		}}, &reply); err != nil {
		t.Fatal(err)
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

	const bonusMetric = "*sum#~*asm.BalanceSummaries.BAL_BONUS.Value"
	if got := metrics[bonusMetric]; got != 10 {
		t.Errorf("expected the last published summary to have BAL_BONUS 10, received %v (%v)", got, metrics)
	}

	const mainMetric = "*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"
	if got := metrics[mainMetric]; got != 70 {
		t.Errorf("expected the last published summary to have BAL_MAIN 70, received %v (%v)", got, metrics)
	}
}

func TestActionsPublishAccountBalanceAutoRecharge(t *testing.T) {
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
    "store_interval": "-1"
},
"apiers": {
    "enabled": true,
    "scheduler_conns": ["*internal"]
}
}`

	ng := engine.TestEngine{
		ConfigJSON: content,
		// LogBuffer:  os.Stderr,
	}
	client, _ := ng.Run(t)

	var reply string
	for _, attrs := range []*v1.V1AttrSetActions{
		{
			ActionsId: "ACT_AUTO_RECHARGE",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 20, Weight: 10},
			},
		},
		{
			ActionsId: "ACT_DEBIT_18",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 18, Weight: 10},
			},
		},
		{
			ActionsId: "ACT_DEBIT_10",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 10, Weight: 10},
			},
		},
		{
			ActionsId: "ACT_DEBIT_9",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaDebit, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 9, Weight: 10},
			},
		},
	} {
		if err := client.Call(context.Background(), utils.APIerSv1SetActions, attrs, &reply); err != nil {
			t.Fatal(err)
		} else if reply != utils.OK {
			t.Fatalf("unexpected reply: %s", reply)
		}
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetThresholdProfile,
		&engine.ThresholdProfileWithAPIOpts{ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_LOW_BALANCE",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1010",
				"*lt:~*asm.BalanceSummaries.BAL_MAIN.Value:5",
			},
			MaxHits:   -1,
			MinHits:   1,
			Weight:    10,
			Async:     true,
			ActionIDs: []string{"ACT_AUTO_RECHARGE"},
		}}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		&utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1010",
			BalanceType: utils.MetaMonetary,
			Value:       20,
			Balance: map[string]any{
				utils.ID: "BAL_MAIN",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	var acnt *engine.Account
	var th *engine.Threshold

	if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_LOW_BALANCE"}}, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 0 {
		t.Fatalf("threshold hit while the balance is 20, Hits: %d", th.Hits)
	}

	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1010", ActionsId: "ACT_DEBIT_18"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	acnt, th = nil, nil
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}, &acnt); err != nil {
		t.Fatal(err)
	} else if len(acnt.BalanceMap[utils.MetaMonetary]) != 1 {
		t.Fatalf("expected 1 monetary balance, received %s", utils.ToJSON(acnt.BalanceMap))
	} else if val := acnt.BalanceMap[utils.MetaMonetary][0].Value; val != 22 {
		t.Errorf("after the low balance call expected BAL_MAIN 22 (2 + 20 recharge), received %v", val)
	}
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_LOW_BALANCE"}}, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("expected 1 hit after the balance dropped under 5, received %d", th.Hits)
	}

	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1010", ActionsId: "ACT_DEBIT_10"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	acnt, th = nil, nil
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}, &acnt); err != nil {
		t.Fatal(err)
	} else if val := acnt.BalanceMap[utils.MetaMonetary][0].Value; val != 12 {
		t.Errorf("after the normal call expected BAL_MAIN 12, received %v", val)
	}
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_LOW_BALANCE"}}, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 1 {
		t.Errorf("threshold fired with the balance above 5, Hits: %d", th.Hits)
	}

	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1010", ActionsId: "ACT_DEBIT_9"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	acnt, th = nil, nil
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1010"}, &acnt); err != nil {
		t.Fatal(err)
	} else if val := acnt.BalanceMap[utils.MetaMonetary][0].Value; val != 23 {
		t.Errorf("after the second low balance call expected BAL_MAIN 23, received %v", val)
	}
	if err := client.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_LOW_BALANCE"}}, &th); err != nil {
		t.Fatal(err)
	} else if th.Hits != 2 {
		t.Errorf("expected 2 hits in total, received %d", th.Hits)
	}
}

func TestActionsPublishAccountTriggerBalanceWithBonus(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaPostgres, utils.MetaMongo, utils.MetaMySQL:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	content := `{
"general": {"log_level": 7},
"data_db": {"db_type": "*internal"},
"stor_db": {"db_type": "*internal"},
"rals": {"enabled": true, "stats_conns": ["*internal"]},
"schedulers": {"enabled": true},
"stats": {"enabled": true, "store_interval": "-1"},
"apiers": {"enabled": true, "scheduler_conns": ["*internal"]}
}`

	ng := engine.TestEngine{ConfigJSON: content}
	client, _ := ng.Run(t)

	var reply string
	for _, attrs := range []*v1.V1AttrSetActions{
		{
			ActionsId: "ACT_BONUS_10",
			Overwrite: true,
			Actions: []*v1.V1TPAction{
				{Identifier: utils.MetaTopUp, BalanceType: utils.MetaMonetary, BalanceId: "BAL_MAIN", Units: 10, Weight: 10},
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
		}
	}

	if err := client.Call(context.Background(), utils.APIerSv2SetBalance,
		&utils.AttrSetBalance{
			Tenant: "cgrates.org", Account: "1011",
			BalanceType: utils.MetaMonetary, Value: 10,
			Balance: map[string]any{utils.ID: "BAL_MAIN"},
		}, &reply); err != nil {
		t.Fatal(err)
	}
	if err := client.Call(context.Background(), utils.APIerSv1SetAccountActionTriggers,
		&v1.AttrSetAccountActionTriggers{
			Tenant: "cgrates.org", Account: "1011",
			AttrSetActionTrigger: engine.AttrSetActionTrigger{
				GroupID: "TRG_BONUS",
				ActionTrigger: map[string]any{
					utils.ThresholdType:  "*max_balance",
					utils.ThresholdValue: 50,
					utils.BalanceType:    utils.MetaMonetary,
					utils.BalanceID:      "BAL_MAIN",
					utils.ActionsID:      "ACT_BONUS_10",
				},
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1SetStatQueueProfile,
		&engine.StatQueueProfileWithAPIOpts{StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "STS_SAME_BAL",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.AccountID:1011",
			},
			QueueLength: 1,
			TTL:         -1,
			Metrics: []*engine.MetricWithFilters{
				{MetricID: "*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"},
			},
		}}, &reply); err != nil {
		t.Fatal(err)
	}

	if err := client.Call(context.Background(), utils.APIerSv1ExecuteAction,
		&utils.AttrExecuteAction{Tenant: "cgrates.org", Account: "1011", ActionsId: "ACT_TOPUP_60"}, &reply); err != nil {
		t.Fatal(err)
	}
	time.Sleep(500 * time.Millisecond)

	var acnt *engine.Account
	if err := client.Call(context.Background(), utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1011"}, &acnt); err != nil {
		t.Fatal(err)
	}
	if bal := acnt.BalanceMap[utils.MetaMonetary]; len(bal) != 1 || bal[0].Value != 80 {
		t.Fatalf("expected BAL_MAIN 80 after the trigger, received %v", utils.ToJSON(bal))
	}

	var metrics map[string]float64
	if err := client.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STS_SAME_BAL"}}, &metrics); err != nil {
		t.Fatal(err)
	}
	const metric = "*sum#~*asm.BalanceSummaries.BAL_MAIN.Value"
	if got := metrics[metric]; got != 80 {
		t.Errorf("expected the last published summary to have BAL_MAIN 80, received %v", got)
	}
}
