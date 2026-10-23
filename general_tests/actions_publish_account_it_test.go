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

	// t.Skip("Related to  #2887, actions do not publish account summary")
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
