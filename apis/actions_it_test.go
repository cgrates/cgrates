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
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
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
		testActionsResetStorDB,
		testActionsStartEngine,
		testActionsRPCConn,
		testActionsGetActionProfileBeforeSet,
		testActionsGetActionProfileIDsBeforeSet,
		testActionsGetActionProfileCountBeforeSet,
		testActionsSetActionProfile,
		testActionsGetActionProfileAfterSet,
		testActionsGetActionProfileIDsAfterSet,
		testActionsGetActionProfileCountAfterSet,
		testActionsRemoveActionProfile,
		testActionsGetActionProfileAfterRemove,
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

		testActionsKillEngine,
	}
)

func TestActionsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		actConfigDIR = "apis_actions_internal"
	case utils.MetaMongo:
		actConfigDIR = "apis_actions_mongo"
	case utils.MetaMySQL:
		actConfigDIR = "apis_actions_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsAct {
		t.Run(actConfigDIR, stest)
	}
}

func testActionsInitCfg(t *testing.T) {
	var err error
	actCfgPath = path.Join(*dataDir, "conf", "samples", actConfigDIR)
	actCfg, err = config.NewCGRConfigFromPath(context.Background(), actCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testActionsInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(actCfg); err != nil {
		t.Fatal(err)
	}
}

func testActionsResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(actCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testActionsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(actCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testActionsRPCConn(t *testing.T) {
	var err error
	actRPC, err = newRPCClient(actCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
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
	var rplyAct engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfileIDsBeforeSet(t *testing.T) {
	var rplyActIDs []string
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileIDs,
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rplyActIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testActionsGetActionProfileCountBeforeSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileCount,
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

	var reply string
	if err := actRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf engine.ActionProfile
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
	expAct := engine.ActionProfile{
		Tenant: "cgrates.org",
		ID:     "actPrfID",
		Actions: []*engine.APAction{
			{
				ID: "actID",
			},
		},
	}

	var rplyAct engine.ActionProfile
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
		&utils.PaginatorWithTenant{
			Tenant:    "cgrates.org",
			Paginator: utils.Paginator{},
		}, &rplyActIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyActIDs, expActIDs) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expActIDs, rplyActIDs)
	}
}

func testActionsGetActionProfileCountAfterSet(t *testing.T) {
	var rplyCount int
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfileCount,
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
	var rplyAct engine.ActionProfile
	if err := actRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "actPrfID",
			}}, &rplyAct); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
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
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaHTTPPost,
					Diktats: []*engine.APDiktat{
						{
							Path: actSrv.URL,
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

	var rplyActPrf engine.ActionProfile
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
	args := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	expBody := `{"*opts":null,"*req":{"Account":"1001"}}`
	var reply string
	if err := actRPC.Call(context.Background(), utils.ActionSv1ExecuteActions,
		args, &reply); err != nil {
		t.Error(err)
	}

	if string(actBody) != expBody {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expBody, string(actBody))
	}
}

func testActionsSetStatQueueProfileBeforeExecuteResetSQ(t *testing.T) {
	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "SQ_ID",
			Weight:      10,
			QueueLength: 100,
			TTL:         time.Duration(1 * time.Minute),
			MinItems:    0,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
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
		Tenant:      "cgrates.org",
		ID:          "SQ_ID",
		Weight:      10,
		QueueLength: 100,
		TTL:         time.Duration(1 * time.Minute),
		MinItems:    0,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
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
	args := &engine.StatsArgsProcessEvent{
		StatIDs: []string{"SQ_ID"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "StatsEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Usage:        30 * time.Second,
			},
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
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "actPrfID",
			Actions: []*engine.APAction{
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
	args := &utils.ArgActionSv1ScheduleActions{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EventExecuteActions",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
		ActionProfileIDs: []string{"actPrfID"},
	}

	var reply string
	if err := actRPC.Call(context.Background(), utils.ActionSv1ScheduleActions,
		args, &reply); err != nil {
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
