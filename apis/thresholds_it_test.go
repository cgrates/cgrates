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

package apis

import (
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	thCfgPath   string
	thCfg       *config.CGRConfig
	thRPC       *birpc.Client
	thConfigDIR string //run tests for specific configuration

	sTestsTh = []func(t *testing.T){
		testThresholdsInitCfg,
		testThresholdsInitDataDB,
		testThresholdsResetStorDB,
		testThresholdsStartEngine,
		testThresholdsRPCConn,
		testThresholdsGetThresholdBeforeSet,
		testThresholdsSetActionProfile,
		testThresholdsSetThresholdProfiles,
		testThresholdsGetThresholdAfterSet,
		testThresholdsGetThresholdIDs,
		testThresholdsGetThresholdProfileIDs,
		testThresholdsGetThresholdProfileCount,
		testThresholdsGetThresholdsForEvent,
		testThresholdsRemoveThresholdProfiles,
		testThresholdsGetThresholdsAfterRemove,

		// test if actions are executed properly when thresholds are hit
		testThresholdsSetActionProfileBeforeProcessEv,
		testThresholdsSetThresholdProfilesBeforeProcessEv,
		testThresholdsProcessEvent,
		testThresholdsGetThresholdsAfterFirstEvent,
		testThresholdsProcessEvent,
		testThresholdsGetThresholdsAfterSecondEvent,
		testThresholdsPing,
		testThresholdsKillEngine,
	}
)

func TestThresholdsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		thConfigDIR = "thresholds_internal"
	case utils.MetaMongo:
		thConfigDIR = "thresholds_mongo"
	case utils.MetaMySQL:
		thConfigDIR = "thresholds_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTh {
		t.Run(thConfigDIR, stest)
	}
}

func testThresholdsInitCfg(t *testing.T) {
	var err error
	thCfgPath = path.Join(*dataDir, "conf", "samples", thConfigDIR)
	thCfg, err = config.NewCGRConfigFromPath(thCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testThresholdsInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(thCfg); err != nil {
		t.Fatal(err)
	}
}

func testThresholdsResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(thCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testThresholdsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(thCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testThresholdsRPCConn(t *testing.T) {
	var err error
	thRPC, err = newRPCClient(thCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testThresholdsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testThresholdsPing(t *testing.T) {
	var reply string
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testThresholdsGetThresholdBeforeSet(t *testing.T) {
	var rplyTh engine.Threshold

	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &rplyTh); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}
func testThresholdsSetActionProfile(t *testing.T) {
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	var reply *string
	if err := thRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	}
}

func testThresholdsSetThresholdProfiles(t *testing.T) {
	thPrf1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_1",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          5,
			MinHits:          1,
			Weight:           10,
		},
	}

	var reply string
	if err := thRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
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
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          7,
			MinHits:          0,
			Weight:           20,
		},
	}

	if err := thRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testThresholdsGetThresholdAfterSet(t *testing.T) {
	var rplyTh engine.Threshold
	var rplyThPrf engine.ThresholdProfile
	expTh := engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_1",
	}
	expThPrf := engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_1",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          5,
		MinHits:          1,
		Weight:           10,
	}

	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THD_1",
			},
		}, &rplyTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyTh, expTh) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expTh), utils.ToJSON(rplyTh))
	}

	if err := thRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_1",
		}, &rplyThPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyThPrf, expThPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expThPrf), utils.ToJSON(rplyThPrf))
	}

	expTh = engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_2",
	}
	expThPrf = engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_2",
		FilterIDs:        []string{"*string:~*req.Account:1001"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          7,
		MinHits:          0,
		Weight:           20,
	}

	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THD_2",
			},
		}, &rplyTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyTh, expTh) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expTh), utils.ToJSON(rplyTh))
	}

	if err := thRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_2",
		}, &rplyThPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyThPrf, expThPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expThPrf), utils.ToJSON(rplyThPrf))
	}
}

func testThresholdsGetThresholdIDs(t *testing.T) {
	expIDs := []string{"THD_1", "THD_2"}
	var tIDs []string
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThresholdIDs,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &tIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(tIDs)
		if !reflect.DeepEqual(tIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, tIDs)
		}
	}
}

func testThresholdsGetThresholdProfileIDs(t *testing.T) {
	expIDs := []string{"THD_1", "THD_2"}
	var tIDs []string
	if err := thRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &tIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(tIDs)
		if !reflect.DeepEqual(tIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, tIDs)
		}
	}
}

func testThresholdsGetThresholdProfileCount(t *testing.T) {
	var reply int
	if err := thRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfileCount,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2, reply)
	}
}

func testThresholdsGetThresholdsForEvent(t *testing.T) {
	args := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_1", "THD_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "GetThresholdEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   0,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   0,
		},
	}

	var rplyThs engine.Thresholds
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyThs, expThs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expThs), utils.ToJSON(rplyThs))
	}
}

func testThresholdsRemoveThresholdProfiles(t *testing.T) {
	var reply string

	if err := thRPC.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THD_1",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	if err := thRPC.Call(context.Background(), utils.AdminSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THD_2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testThresholdsGetThresholdsAfterRemove(t *testing.T) {
	args := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_1", "THD_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "RemThresholdEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	var rplyThs engine.Thresholds
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testThresholdsSetActionProfileBeforeProcessEv(t *testing.T) {
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaResetThreshold,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaThresholds: {
					"THD_1": struct{}{},
					"THD_2": struct{}{},
				},
			},
		},
	}

	var reply *string
	if err := thRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	}
}

func testThresholdsSetThresholdProfilesBeforeProcessEv(t *testing.T) {
	thPrf1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_1",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          5,
			MinHits:          3,
			Weight:           10,
		},
	}

	var reply string
	if err := thRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
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
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          2,
			MinHits:          0,
			Weight:           20,
		},
	}

	if err := thRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testThresholdsProcessEvent(t *testing.T) {
	args := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_1", "THD_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdProcessEv",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	expIDs := []string{"THD_1", "THD_2"}
	var tIDs []string
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1ProcessEvent, args, &tIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(tIDs)
		if !reflect.DeepEqual(tIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, tIDs)
		}
	}
}

func testThresholdsGetThresholdsAfterFirstEvent(t *testing.T) {
	args := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_1", "THD_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   0,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   1,
		},
	}

	var rplyThs engine.Thresholds
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else {
		for idx, thd := range rplyThs {
			thd.Snooze = expThs[idx].Snooze
		}
		if !reflect.DeepEqual(rplyThs, expThs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThs), utils.ToJSON(rplyThs))
		}
	}
}

func testThresholdsGetThresholdsAfterSecondEvent(t *testing.T) {
	args := &engine.ThresholdsArgsProcessEvent{
		ThresholdIDs: []string{"THD_1", "THD_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "ThresholdEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expThs := engine.Thresholds{
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_2",
			Hits:   0,
		},
		&engine.Threshold{
			Tenant: "cgrates.org",
			ID:     "THD_1",
			Hits:   2,
		},
	}

	var rplyThs engine.Thresholds
	if err := thRPC.Call(context.Background(), utils.ThresholdSv1GetThresholdsForEvent,
		args, &rplyThs); err != nil {
		t.Error(err)
	} else {
		for idx, thd := range rplyThs {
			thd.Snooze = expThs[idx].Snooze
		}
		if !reflect.DeepEqual(rplyThs, expThs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expThs), utils.ToJSON(rplyThs))
		}
	}
}
