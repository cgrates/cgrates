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
	sqSrv       *httptest.Server
	sqBody      []byte
	sqCfgPath   string
	sqCfg       *config.CGRConfig
	sqRPC       *birpc.Client
	sqConfigDIR string //run tests for specific configuration

	sTestsSq = []func(t *testing.T){
		testStatsInitCfg,
		testStatsInitDataDB,
		testStatsStartEngine,
		testStatsRPCConn,
		testStatsGetStatQueueBeforeSet,
		testStatsGetStatQueueProfilesBeforeSet,
		testStatsSetStatQueueProfiles,
		testStatsGetStatQueueAfterSet,
		testStatsGetStatQueueIDs,
		testStatsGetStatQueueProfileIDs,
		testStatsGetStatQueueProfiles1,
		testStatsGetStatQueueProfilesWithPrefix,
		testStatsGetStatQueueProfilesCount,
		testStatsRemoveStatQueueProfiles,
		testStatsGetStatQueuesAfterRemove,
		//testStatsProcessEventWithBlockersOnMetrics,
		// check if stats, thresholds and actions subsystems function properly together
		testStatsStartServer,
		testStatsSetActionProfileBeforeProcessEv,
		testStatsSetThresholdProfilesBeforeProcessEv,
		testStatsSetStatQueueProfileBeforeProcessEv,
		testStatsProcessEvent,
		testStatsGetStatQueuesAfterProcessEv,
		testStatsGetThresholdAfterProcessEvent,
		testStatsStopServer,
		testStatsPing,
		testStatsKillEngine,
	}
)

func TestStatsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sqConfigDIR = "stats_internal"
	case utils.MetaMongo:
		sqConfigDIR = "stats_mongo"
	case utils.MetaMySQL:
		sqConfigDIR = "stats_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSq {
		t.Run(sqConfigDIR, stest)
	}
}

func testStatsInitCfg(t *testing.T) {
	var err error
	sqCfgPath = path.Join(*dataDir, "conf", "samples", sqConfigDIR)
	sqCfg, err = config.NewCGRConfigFromPath(context.Background(), sqCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testStatsInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(sqCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testStatsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sqCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testStatsRPCConn(t *testing.T) {
	var err error
	sqRPC, err = newRPCClient(sqCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testStatsKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testStatsPing(t *testing.T) {
	var reply string
	if err := sqRPC.Call(context.Background(), utils.StatSv1Ping,
		new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testStatsGetStatQueueBeforeSet(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "StatsEventTest",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsStatsProfileIDs: []string{"SQ_1", "SQ_2"},
		},
	}

	var rplySqs engine.StatQueues
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueuesForEvent,
		args, &rplySqs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testStatsGetStatQueueProfilesBeforeSet(t *testing.T) {
	var reply []*engine.StatQueueProfile
	var args *utils.ArgsItemIDs
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testStatsSetStatQueueProfiles(t *testing.T) {
	sqPrf1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			QueueLength: 100,
			TTL:         time.Duration(1 * time.Minute),
			MinItems:    5,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACC,
				},
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaDDC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	var reply string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	sqPrf2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testStatsGetStatQueueAfterSet(t *testing.T) {
	var rplySqPrf engine.StatQueueProfile
	expStrMetrics := map[string]string{
		utils.MetaACC: utils.NotAvailable,
		utils.MetaACD: utils.NotAvailable,
		utils.MetaASR: utils.NotAvailable,
		utils.MetaDDC: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
	}
	expFloatMetrics := map[string]float64{
		utils.MetaACC: -1,
		utils.MetaACD: -1,
		utils.MetaASR: -1,
		utils.MetaDDC: -1,
		utils.MetaTCD: -1,
	}
	expSqPrf := engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_1",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		QueueLength: 100,
		TTL:         time.Duration(1 * time.Minute),
		MinItems:    5,
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaACC,
			},
			{
				MetricID: utils.MetaACD,
			},
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaDDC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	rplyStrMetrics := make(map[string]string)
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_1",
			},
		}, &rplyStrMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyStrMetrics, expStrMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expStrMetrics), utils.ToJSON(rplyStrMetrics))
	}

	rplyFloatMetrics := make(map[string]float64)
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_1",
			},
		}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFloatMetrics), utils.ToJSON(rplyFloatMetrics))
	}

	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_1",
		}, &rplySqPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
	}

	expStrMetrics = map[string]string{
		utils.MetaASR: utils.NotAvailable,
		utils.MetaPDD: utils.NotAvailable,
		utils.MetaTCC: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
	}
	expFloatMetrics = map[string]float64{
		utils.MetaASR: -1,
		utils.MetaPDD: -1,
		utils.MetaTCC: -1,
		utils.MetaTCD: -1,
	}
	expSqPrf = engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_2",
		Weights: utils.DynamicWeights{
			{
				Weight: 20,
			},
		},
		Metrics: []*engine.MetricWithFilters{
			{
				MetricID: utils.MetaASR,
			},
			{
				MetricID: utils.MetaTCD,
			},
			{
				MetricID: utils.MetaPDD,
			},
			{
				MetricID: utils.MetaTCC,
			},
			{
				MetricID: utils.MetaTCD,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}

	rplyStrMetrics = map[string]string{}
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
			},
		}, &rplyStrMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyStrMetrics, expStrMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expStrMetrics), utils.ToJSON(rplyStrMetrics))
	}

	rplyFloatMetrics = make(map[string]float64)
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
			},
		}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFloatMetrics), utils.ToJSON(rplyFloatMetrics))
	}

	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_2",
		}, &rplySqPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
	}
}

func testStatsGetStatQueueIDs(t *testing.T) {
	expIDs := []string{"SQ_1", "SQ_2"}
	var sqIDs []string
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueIDs,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &sqIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(sqIDs)
		if !reflect.DeepEqual(sqIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, sqIDs)
		}
	}
}

func testStatsGetStatQueueProfileIDs(t *testing.T) {
	expIDs := []string{"SQ_1", "SQ_2"}
	var sqIDs []string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &sqIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(sqIDs)
		if !reflect.DeepEqual(sqIDs, expIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, sqIDs)
		}
	}
}

func testStatsGetStatQueueProfiles1(t *testing.T) {
	var reply []*engine.StatQueueProfile
	var args *utils.ArgsItemIDs
	exp := []*engine.StatQueueProfile{
		{
			Tenant: "cgrates.org",
			ID:     "SQ_1",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			QueueLength: 100,
			TTL:         time.Duration(1 * time.Minute),
			MinItems:    5,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACC,
				},
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaDDC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
		{
			Tenant: "cgrates.org",
			ID:     "SQ_2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(reply, func(i int, j int) bool {
		return (reply)[i].ID < (reply)[j].ID
	})
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
	}
}

func testStatsGetStatQueueProfilesWithPrefix(t *testing.T) {
	var reply []*engine.StatQueueProfile
	args := &utils.ArgsItemIDs{
		ItemsPrefix: "SQ_2",
	}
	exp := []*engine.StatQueueProfile{
		{
			Tenant: "cgrates.org",
			ID:     "SQ_2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(reply, exp) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", exp, reply)
	}
}
func testStatsGetStatQueueProfilesCount(t *testing.T) {
	var reply int
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfilesCount,
		&utils.TenantWithAPIOpts{
			Tenant: "cgrates.org",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2, reply)
	}
}

func testStatsRemoveStatQueueProfiles(t *testing.T) {
	var reply string

	if err := sqRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_1",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	if err := sqRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testStatsGetStatQueuesAfterRemove(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "StatsEventTest",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsStatsProfileIDs: []string{"SQ_1", "SQ_2"},
		},
	}

	var reply []string
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueuesForEvent,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testStatsProcessEventWithBlockersOnMetrics(t *testing.T) {
	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "SQ_WithBlockers",
			FilterIDs:   []string{"*string:~*req.StatsMetrics:*exists"},
			QueueLength: 100,
			TTL:         time.Duration(1 * time.Minute),
			MinItems:    0,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaASR,
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"*prefix:~*req.CallerID:4433"},
							Blocker:   true,
						},
					},
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaACD,
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"*prefix:~*req.CallerID:44112"},
							Blocker:   true,
						},
					},
				},
				{
					MetricID: utils.MetaDDC,
				},
			},
			ThresholdIDs: []string{"*none"},
		},
	}

	var reply string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "SQ_WithBlockers_Event",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
			"StatsMetrics":     "*exists",
			"CallerID":         "443321",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:       30 * time.Second,
			utils.MetaCost:        102.1,
			utils.MetaDestination: "332214",
			utils.MetaStartTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.MetaPDD:         5 * time.Second,
		},
	}
	expected := []string{"SQ_WithBlockers"}
	var replyStats []string
	if err := sqRPC.Call(context.Background(), utils.StatSv1ProcessEvent,
		args, &replyStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyStats, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, replyStats)
	}

	/*
		expFloat := map[string]float64{
			utils.MetaACD: 46400000000.,
			utils.MetaASR: 60.,
		}
		var rplyDec map[string]float64
		if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
			&utils.TenantIDWithAPIOpts{
				TenantID: &utils.TenantID{
					Tenant: "cgrates.org",
					ID:     "SQ_WithBlockers",
				},
			}, &rplyDec); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(rplyDec, expFloat) {
			t.Errorf("Expected %v, received %v", utils.ToJSON(expFloat), utils.ToJSON(rplyDec))
		}
	*/
}

func testStatsStartServer(t *testing.T) {
	sqSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		sqBody, err = io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}

		r.Body.Close()
	}))
}

func testStatsStopServer(t *testing.T) {
	sqSrv.Close()
}

func testStatsSetActionProfileBeforeProcessEv(t *testing.T) {
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
							Path: sqSrv.URL,
						},
					},
					TTL: time.Duration(time.Minute),
				},
			},
		},
	}

	var reply string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyActPrf engine.ActionProfile
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetActionProfile,
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

func testStatsSetThresholdProfilesBeforeProcessEv(t *testing.T) {
	thPrf := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_ID",
			FilterIDs:        []string{"*string:~*req.EventType:StatUpdate"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          2,
			MinHits:          0,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
	}

	var reply string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	var rplyTh engine.Threshold
	var rplyThPrf engine.ThresholdProfile
	expTh := engine.Threshold{
		Tenant: "cgrates.org",
		ID:     "THD_ID",
	}
	expThPrf := engine.ThresholdProfile{
		Tenant:           "cgrates.org",
		ID:               "THD_ID",
		FilterIDs:        []string{"*string:~*req.EventType:StatUpdate"},
		ActionProfileIDs: []string{"actPrfID"},
		MaxHits:          2,
		MinHits:          0,
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
	}

	if err := sqRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "THD_ID",
			},
		}, &rplyTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyTh, expTh) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expTh), utils.ToJSON(rplyTh))
	}

	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetThresholdProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "THD_ID",
		}, &rplyThPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyThPrf, expThPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expThPrf), utils.ToJSON(rplyThPrf))
	}

}

func testStatsSetStatQueueProfileBeforeProcessEv(t *testing.T) {
	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_3",
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
			ThresholdIDs: []string{"THD_ID"},
		},
	}

	var reply string
	if err := sqRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	expSqPrf := engine.StatQueueProfile{
		Tenant: "cgrates.org",
		ID:     "SQ_3",
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
		ThresholdIDs: []string{"THD_ID"},
	}

	var rplySqPrf engine.StatQueueProfile
	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
		utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "SQ_3",
		}, &rplySqPrf); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
	}
}

func testStatsProcessEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "StatsEventTest",
		Event: map[string]interface{}{
			utils.AccountField: "1001",
		},
		APIOpts: map[string]interface{}{
			utils.OptsStatsProfileIDs: []string{"SQ_3"},
			utils.MetaUsage:           30 * time.Second,
		},
	}
	expected := []string{"SQ_3"}
	expBody := `{"*opts":{"*actProfileIDs":["actPrfID"],"*eventType":"StatUpdate","*statsProfileIDs":["SQ_3"],"*thdProfileIDs":["THD_ID"],"*usage":30000000000},"*req":{"*tcd":30000000000,"EventType":"StatUpdate","StatID":"SQ_3"}}`
	var reply []string
	if err := sqRPC.Call(context.Background(), utils.StatSv1ProcessEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, reply)
	}

	if expBody != string(sqBody) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", expBody, string(sqBody))
	}
}

func testStatsGetStatQueuesAfterProcessEv(t *testing.T) {
	expFloatMetrics := map[string]float64{
		utils.MetaTCD: 30000000000,
	}

	rplyFloatMetrics := make(map[string]float64)
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "SQ_3",
			},
		}, &rplyFloatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFloatMetrics, expFloatMetrics) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expFloatMetrics), utils.ToJSON(rplyFloatMetrics))
	}
}

func testStatsGetThresholdAfterProcessEvent(t *testing.T) {
	args := &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "THD_ID",
	}
	var reply *engine.Threshold
	if err := sqRPC.Call(context.Background(), utils.ThresholdSv1GetThreshold,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply.Hits != 1 {
		t.Errorf("\nexpected: <%+v>, \nreceived: <%+v>", 1, reply.Hits)
	}
}
