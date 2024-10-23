//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package general_tests

import (
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	ng1CfgPath, ng2CfgPath string
	ng1Cfg, ng2Cfg         *config.CGRConfig
	ng1RPC, ng2RPC         *birpc.Client
	ng1ConfDIR, ng2ConfDIR string //run tests for specific configuration
	rrDelay                int

	rrTests = []func(t *testing.T){
		testRPCExpLoadConfig,
		testRPCExpFlushDBs,

		testRPCExpStartEngine,
		testRPCExpRPCConn,
		testRPCExpSetThresholdProfilesBeforeProcessEv,
		testRPCExpProcessEventWithConfigOpts,
		testRPCExpGetThresholdsAfterFirstEvent,
		testRPCExpProcessEventWithAPIOpts,
		testRPCExpGetThresholdsAfterSecondEvent,
		testRPCExpProcessEventWithRPCAPIOpts,
		testRPCExpGetThresholdsAfterThirdEvent,
		testRPCExpStopEngine,
	}
)

func TestRPCExpIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		ng1ConfDIR = "rpcexp_opts_engine1_internal"
		ng2ConfDIR = "rpcexp_opts_engine2_internal"
	case utils.MetaMySQL:
		ng1ConfDIR = "rpcexp_opts_engine1_mysql"
		ng2ConfDIR = "rpcexp_opts_engine2_mysql"
	case utils.MetaMongo:
		ng1ConfDIR = "rpcexp_opts_engine1_mongo"
		ng2ConfDIR = "rpcexp_opts_engine2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range rrTests {
		t.Run(ng1ConfDIR, stest)
	}
}

func testRPCExpLoadConfig(t *testing.T) {
	var err error
	ng1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "rpcexp_multiple_engines", ng1ConfDIR)
	if ng1Cfg, err = config.NewCGRConfigFromPath(context.Background(), ng1CfgPath); err != nil {
		t.Error(err)
	}
	ng2CfgPath = path.Join(*utils.DataDir, "conf", "samples", "rpcexp_multiple_engines", ng2ConfDIR)
	if ng2Cfg, err = config.NewCGRConfigFromPath(context.Background(), ng2CfgPath); err != nil {
		t.Error(err)
	}
	rrDelay = 1000
}

func testRPCExpFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(ng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDB(ng2Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(ng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(ng2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testRPCExpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ng1CfgPath, rrDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(ng2CfgPath, rrDelay); err != nil {
		t.Fatal(err)
	}
}

func testRPCExpRPCConn(t *testing.T) {
	ng1RPC = engine.NewRPCClient(t, ng1Cfg.ListenCfg(), *utils.Encoding)
	ng2RPC = engine.NewRPCClient(t, ng2Cfg.ListenCfg(), *utils.Encoding)
}
func testRPCExpSetThresholdProfilesBeforeProcessEv(t *testing.T) {
	thPrf1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_1",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"*none)"},
			MaxHits:          5,
			MinHits:          3,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Async: true,
		},
	}

	var reply string
	if err := ng2RPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	thPrf2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_2",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"*none"},
			MaxHits:          3,
			MinHits:          2,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Async: true,
		},
	}

	if err := ng2RPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	thPrf3 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_3",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"*none"},
			MaxHits:          4,
			MinHits:          1,
			Weights: utils.DynamicWeights{
				{
					Weight: 15,
				},
			},
			Async: true,
		},
	}

	if err := ng2RPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testRPCExpProcessEventWithConfigOpts(t *testing.T) {
	args := utils.CGREventWithEeIDs{
		EeIDs: []string{"thProcessEv1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEv1",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{},
		},
	}
	var reply map[string]map[string]any
	if err := ng1RPC.Call(context.Background(), utils.EeSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)

}

func testRPCExpGetThresholdsAfterFirstEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThresholdEventTest1",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"THD_1", "THD_2", "THD_3"},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   1,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   0,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_3",
			Hits:   0,
		},
	}

	var rplyThs engine.Thresholds
	if err := ng2RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rplyThs, func(i, j int) bool {
			return rplyThs[i].ID < rplyThs[j].ID
		})
		for idx, thd := range rplyThs {
			thd.Snooze = expThs[idx].Snooze
		}
		if !reflect.DeepEqual(rplyThs, expThs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThs), utils.ToJSON(rplyThs))
		}
	}
}

func testRPCExpProcessEventWithAPIOpts(t *testing.T) {
	args := utils.CGREventWithEeIDs{
		EeIDs: []string{"thProcessEv1"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEv2",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsThresholdsProfileIDs: []string{"THD_2"},
			},
		},
	}
	var reply map[string]map[string]any
	if err := ng1RPC.Call(context.Background(), utils.EeSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
}

func testRPCExpGetThresholdsAfterSecondEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThresholdEventTest2",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"THD_1", "THD_2", "THD_3"},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   1,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   1,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_3",
			Hits:   0,
		},
	}

	var rplyThs engine.Thresholds
	if err := ng2RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rplyThs, func(i, j int) bool {
			return rplyThs[i].ID < rplyThs[j].ID
		})
		for idx, thd := range rplyThs {
			thd.Snooze = expThs[idx].Snooze
		}
		if !reflect.DeepEqual(rplyThs, expThs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThs), utils.ToJSON(rplyThs))
		}
	}
}

func testRPCExpProcessEventWithRPCAPIOpts(t *testing.T) {
	args := utils.CGREventWithEeIDs{
		EeIDs: []string{"thProcessEv2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEv3",
			Event: map[string]any{
				utils.AccountField: "1001",
			},
			APIOpts: map[string]any{
				utils.OptsThresholdsProfileIDs: []string{"THD_2"},
			},
		},
	}
	var reply map[string]map[string]any
	if err := ng1RPC.Call(context.Background(), utils.EeSv1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
}

func testRPCExpGetThresholdsAfterThirdEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "ThresholdEventTest3",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsThresholdsProfileIDs: []string{"THD_1", "THD_2", "THD_3"},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   1,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   1,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_3",
			Hits:   1,
		},
	}

	var rplyThs engine.Thresholds
	if err := ng2RPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else {
		sort.Slice(rplyThs, func(i, j int) bool {
			return rplyThs[i].ID < rplyThs[j].ID
		})
		for idx, thd := range rplyThs {
			thd.Snooze = expThs[idx].Snooze
		}
		if !reflect.DeepEqual(rplyThs, expThs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThs), utils.ToJSON(rplyThs))
		}
	}
}

func testRPCExpStopEngine(t *testing.T) {
	if err := engine.KillEngine(rrDelay); err != nil {
		t.Error(err)
	}
}
