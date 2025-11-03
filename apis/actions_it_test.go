//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package apis

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"slices"
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
	actSrv       *httptest.Server
	actBody      []byte
	actCfgPath   string
	actCfg       *config.CGRConfig
	actRPC       *birpc.Client
	actConfigDIR string //run tests for specific configuration

	sTestsAct = []func(t *testing.T){
		testActionsInitCfg,
		testActionsInitDataDB,

		testActionsStartEngine,
		testActionsRPCConn,
		testActionsGetActionProfileBeforeSet,
		testActionsGetActionProfilesBeforeSet,
		testActionsGetActionProfileIDsBeforeSet,
		testActionsGetActionProfilesCountBeforeSet,
		testActionsSetActionProfile,
		testActionsGetActionProfileAfterSet,
		testActionsGetActionProfilesAfterSet,
		testActionsGetActionProfileIDsAfterSet,
		testActionsGetActionProfilesCountAfterSet,
		testActionsRemoveActionProfile,
		testActionsGetActionProfileAfterRemove,
		testActionsGetActionProfilesAfterRemove,
		testActionsSetActionProfilesWithPrefix,
		testActionsPing,

		// execute http_post
		testActionsStartServer,
		testActionsSetActionProfileBeforeExecuteHTTPPost,
		testActionsExecuteActionsHTTPPost,
		testActionsStopServer,
		testActionsRemoveActionProfile,
		testActionsGetActionProfileAfterRemove,

		// execute reset_statqueue
		testActionsSetStatQueueProfileBeforeExecuteResetSQ,
		testActionsStatProcessEvent,
		testActionsSetActionProfileBeforeExecuteResetSQ,
		testActionsGetStatQueuesBeforeReset,
		testActionsScheduleActionsResetSQ,
		testActionsGetStatQueuesBeforeReset,
		testActionsSleep,
		testActionsGetStatQueueAfterReset,
		testActionsRemoveActionProfile,
		testActionsGetActionProfileAfterRemove,

		// ActionProfile blocker behaviour test
		testActionsRemoveActionProfiles,
		testActionsSetActionProfiles,
		testActionsExecuteActions,

		testActionsKillEngine,
	}
)

func TestActionsIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		actConfigDIR = "apis_actions_internal"
	case utils.MetaMongo:
		actConfigDIR = "apis_actions_mongo"
	case utils.MetaRedis:
		t.SkipNow()
	case utils.MetaMySQL:
		actConfigDIR = "apis_actions_mysql"
	case utils.MetaPostgres:
		actConfigDIR = "apis_actions_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAct {
		t.Run(actConfigDIR, stest)
	}
}

func testActionsInitCfg(t *testing.T) {
	var err error
	actCfgPath = path.Join(*utils.DataDir, "conf", "samples", actConfigDIR)
	actCfg, err = config.NewCGRConfigFromPath(context.Background(), actCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsInitDataDB(t *testing.T) {
	if err := engine.InitDB(actCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testActionsRPCConn(t *testing.T) {
	actRPC = engine.NewRPCClient(t, actCfg.ListenCfg(), *utils.Encoding)
}

// Kill the engine when it is about to be finished
func testActionsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testActionsPing(t *testing.T) {
	var reply string
	if err := actRPC.Call(context.Background(), utils.StatSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testActionsGetActionProfileBeforeSet(t *testing.T) {
	var rplyAct utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfilesBeforeSet(t *testing.T) {
	var rplyAct *[]*utils.ActionProfile
	var args *utils.ArgsItemIDs
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfiles,
		args, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfileIDsBeforeSet(t *testing.T) {
	var rplyActIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rplyActIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfilesCountBeforeSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfilesCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if rplyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, rplyCount)
	}
}

func testActionsSetActionProfile(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
				{
					ID: "actID",
				},
			},
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyActPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActPrf, *actPrf.ActionProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", actPrf.ActionProfile, rplyActPrf)
	}
}

func testActionsGetActionProfileAfterSet(t *testing.T) {
	expAct := utils.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "actPrfID",
		Actions: []*utils.APAction{
			{
				ID: "actID",
			},
		},
	}

	var rplyAct utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyAct, expAct) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expAct), utils.ToJSON(rplyAct))
	}
}

func testActionsGetActionProfileIDsAfterSet(t *testing.T) {
	expActIDs := []string{"actPrfID"}

	var rplyActIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &rplyActIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActIDs, expActIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActIDs, rplyActIDs)
	}
}

func testActionsGetActionProfilesAfterSet(t *testing.T) {
	expActs := []*utils.ActionProfile{
		{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
				{
					ID: "actID",
				},
			},
		},
	}
	var args *utils.ArgsItemIDs
	var rplyActs []*utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfiles,
		args, &rplyActs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActs, expActs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActs, rplyActs)
	}
}

func testActionsGetActionProfilesCountAfterSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfilesCount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyCount); err != nil {
		t.Error(err)
	} else if rplyCount != 1 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 1, rplyCount)
	}
}

func testActionsRemoveActionProfile(t *testing.T) {
	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testActionsGetActionProfileAfterRemove(t *testing.T) {
	var rplyAct utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfilesAfterRemove(t *testing.T) {
	var args *utils.ArgsItemIDs
	var rplyActs []*utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfiles,
		args, &rplyActs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsSetActionProfilesWithPrefix(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "aactPrfID",
			Actions: []*utils.APAction{
				{
					ID: "aactID",
				},
			},
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "aactPrfID",
			}}, &rplyActPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActPrf, *actPrf.ActionProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", actPrf.ActionProfile, rplyActPrf)
	}

	expActs := []*utils.ActionProfile{
		{
			Tenant: "cgrates.org",
			ID:     "aactPrfID",
			Actions: []*utils.APAction{
				{
					ID: "aactID",
				},
			},
		},
	}
	args := &utils.ArgsItemIDs{
		ItemsPrefix: "aa",
	}
	var rplyActs []*utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfiles,
		args, &rplyActs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActs, expActs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActs, rplyActs)
	}
}

func testActionsStartServer(t *testing.T) {
	actSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		actBody, err = io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}

		r.Body.Close()
	}))
}

func testActionsStopServer(t *testing.T) {
	actSrv.Close()
}

func testActionsSetActionProfileBeforeExecuteHTTPPost(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
				{
					ID:   "actID",
					Type: utils.MetaHTTPPost,
					Diktats: []*utils.APDiktat{
						{
							ID: "HttpPost",
							Opts: map[string]any{
								"*url": actSrv.URL,
							},
						},
					},
					TTL: time.Duration(time.Minute),
				},
			},
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf utils.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyActPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActPrf, *actPrf.ActionProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", actPrf.ActionProfile, rplyActPrf)
	}
}

func testActionsExecuteActionsHTTPPost(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	expBody := `{"*opts":{"*actProfileIDs":["actPrfID"]},"*req":{"Account":"1001"}}`
	var reply string
	if err := actRPC.Call(context.Background(), utils.ActionSv1ExecuteActions,
		ev, &reply); err != nil {
		t.Error(err)
	}

	if string(actBody) != expBody {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expBody, string(actBody))
	}
}

func testActionsSetStatQueueProfileBeforeExecuteResetSQ(t *testing.T) {
	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_ID",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			QueueLength: 100,
			TTL:         time.Duration(1 * time.Minute),
			MinItems:    0,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
			Stored:       true,
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	expSqPrf := engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_ID",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength: 100,
		TTL:         time.Duration(1 * time.Minute),
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
		Stored:       true,
	}

	var rplySqPrf engine.StatQueueProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_ID",
		}, &rplySqPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
	}
}

func testActionsStatProcessEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "StatsEventTest",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsStatsProfileIDs: []string{"SQ_ID"},
			utils.MetaUsage:           30 * time.Second,
		},
	}
	expected := []string{"SQ_ID"}
	var reply []string
	if err := actRPC.Call(context.Background(), utils.StatSv1ProcessEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, reply)
	}
}

func testActionsSetActionProfileBeforeExecuteResetSQ(t *testing.T) {
	actPrf := &utils.ActionProfileWithAPIOpts{
		ActionProfile: &utils.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*utils.APAction{
				{
					ID:   "actID",
					Type: utils.MetaResetStatQueue,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaStats: {
					"SQ_ID": struct{}{},
				},
			},
			Schedule: "@every 1s",
		},
	}

	var reply *string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	}
}

func testActionsGetStatQueuesBeforeReset(t *testing.T) {
	expFloatMetrics := map[string]float64{
		utils.MetaTCD: 30000000000,
	}

	rplyFloatMetrics := make(map[string]float64)
	if err := actRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_ID",
			},
		}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFloatMetrics), utils.ToJSON(rplyFloatMetrics))
	}
}

func testActionsScheduleActionsResetSQ(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]any{
			utils.OptsActionsProfileIDs: []string{"actPrfID"},
		},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.ActionSv1ScheduleActions,
		ev, &reply); err != nil {
		t.Error(err)
	}
}

func testActionsGetStatQueueAfterReset(t *testing.T) {
	expFloatMetrics := map[string]float64{
		utils.MetaTCD: -1,
	}

	rplyFloatMetrics := make(map[string]float64)
	if err := actRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_ID",
			},
		}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFloatMetrics), utils.ToJSON(rplyFloatMetrics))
	}
}

func testActionsSleep(t *testing.T) {
	time.Sleep(time.Second)
}

func testActionsRemoveActionProfiles(t *testing.T) {
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"aactPrfID"}
	var actionProfileIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs, args, &actionProfileIDs); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(actionProfileIDs)
		if !slices.Equal(actionProfileIDs, expected) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, actionProfileIDs)
		}
	}
	var reply string
	for _, actionProfileID := range actionProfileIDs {
		argsRem := utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     actionProfileID,
			},
		}
		if err := actRPC.Call(context.Background(), utils.AdminSv1RemoveActionProfile, argsRem, &reply); err != nil {
			t.Fatal(err)
		}
	}
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs, args, &actionProfileIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsSetActionProfiles(t *testing.T) {
	actionProfiles := []*utils.ActionProfileWithAPIOpts{
		{
			ActionProfile: &utils.ActionProfile{
				ID:        "ACTION_TEST_1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:ActionProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "action1",
						Type: utils.MetaLog,
					},
				},
			},
		},
		{
			ActionProfile: &utils.ActionProfile{
				ID:        "ACTION_TEST_2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:ActionProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "action2",
						Type: utils.MetaLog,
					},
				},
			},
		},
		{
			ActionProfile: &utils.ActionProfile{
				ID:        "ACTION_TEST_3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:ActionProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "action3",
						Type: utils.MetaLog,
					},
				},
			},
		},
		{
			ActionProfile: &utils.ActionProfile{
				ID:        "ACTION_TEST_4",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:ActionProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Schedule: utils.MetaASAP,
				Actions: []*utils.APAction{
					{
						ID:   "action4",
						Type: utils.MetaLog,
					},
				},
			},
		},
	}

	var reply string
	for _, actionProfile := range actionProfiles {
		if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
			actionProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testActionsExecuteActions(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventExecuteActions",
		Event: map[string]any{
			"TestCase": "ActionProfileBlockerBehaviour",
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := actRPC.Call(context.Background(), utils.ActionSv1ExecuteActions, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}
