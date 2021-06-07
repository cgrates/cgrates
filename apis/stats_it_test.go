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
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sqCfgPath   string
	sqCfg       *config.CGRConfig
	sqRPC       *birpc.Client
	sqConfigDIR string //run tests for specific configuration

	sTestsSq = []func(t *testing.T){
		testStatsInitCfg,
		testStatsInitDataDB,
		testStatsResetStorDB,
		testStatsStartEngine,
		testStatsRPCConn,
		testStatsGetStatQueueBeforeSet,
		testStatsSetStatQueueProfiles,
		// testStatsGetStatQueueAfterSet,
		// testStatsGetStatQueueIDs,
		// testStatsGetStatQueueProfileIDs,
		// testStatsGetStatQueueProfileCount,
		// testStatsGetStatQueuesForEvent,
		// testStatsRemoveStatQueueProfiles,
		// testStatsGetStatQueuesAfterRemove,
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
	sqCfg, err = config.NewCGRConfigFromPath(sqCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testStatsInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(sqCfg); err != nil {
		t.Fatal(err)
	}
}

func testStatsResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(sqCfg); err != nil {
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
	args := &engine.StatsArgsProcessEvent{
		StatIDs: []string{"SQ_1", "SQ_2"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "StatsEventTest",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}

	var rplySqs engine.StatQueues
	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueuesForEvent,
		args, &rplySqs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testStatsSetStatQueueProfiles(t *testing.T) {
	sqPrf1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:       "cgrates.org",
			ID:           "SQ_1",
			Weight:       10,
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
			Tenant:       "cgrates.org",
			ID:           "SQ_2",
			Weight:       20,
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

// func testStatsGetStatQueueAfterSet(t *testing.T) {
// 	var rplySq engine.StatQueue
// 	var rplySqPrf engine.StatQueueProfile
// 	expSq := engine.StatQueue{
// 		Tenant:    "cgrates.org",
// 		ID:        "SQ_1",
// 		SQMetrics: make(map[string]engine.StatMetric),
// 		SQItems:   []engine.SQItem{},
// 	}
// 	expSqPrf := engine.StatQueueProfile{
// 		Tenant:       "cgrates.org",
// 		ID:           "SQ_1",
// 		Weight:       10,
// 		ThresholdIDs: []string{utils.MetaNone},
// 	}

// 	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueue,
// 		&utils.TenantIDWithAPIOpts{
// 			TenantID: &utils.TenantID{
// 				Tenant: "cgrates.org",
// 				ID:     "SQ_1",
// 			},
// 		}, &rplySq); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySq, expSq) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSq), utils.ToJSON(rplySq))
// 	}

// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
// 		utils.TenantID{
// 			Tenant: "cgrates.org",
// 			ID:     "SQ_1",
// 		}, &rplySqPrf); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
// 	}

// 	expSq = engine.StatQueue{
// 		Tenant:    "cgrates.org",
// 		ID:        "SQ_2",
// 		SQMetrics: make(map[string]engine.StatMetric),
// 		SQItems:   []engine.SQItem{},
// 	}
// 	expSqPrf = engine.StatQueueProfile{
// 		Tenant:       "cgrates.org",
// 		ID:           "SQ_2",
// 		Weight:       20,
// 		ThresholdIDs: []string{utils.MetaNone},
// 	}

// 	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueue,
// 		&utils.TenantIDWithAPIOpts{
// 			TenantID: &utils.TenantID{
// 				Tenant: "cgrates.org",
// 				ID:     "SQ_2",
// 			},
// 		}, &rplySq); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySq, expSq) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSq), utils.ToJSON(rplySq))
// 	}

// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfile,
// 		&utils.TenantID{
// 			Tenant: "cgrates.org",
// 			ID:     "SQ_2",
// 		}, &rplySqPrf); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySqPrf, expSqPrf) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSqPrf), utils.ToJSON(rplySqPrf))
// 	}
// }

// func testStatsGetStatQueueIDs(t *testing.T) {
// 	expIDs := []string{"SQ_1", "SQ_2"}
// 	var sqIDs []string
// 	if err := sqRPC.Call(context.Background(), utils.StatSv1GetQueueIDs,
// 		&utils.TenantWithAPIOpts{
// 			Tenant: "cgrates.org",
// 		}, &sqIDs); err != nil {
// 		t.Error(err)
// 	} else {
// 		sort.Strings(sqIDs)
// 		if !reflect.DeepEqual(sqIDs, expIDs) {
// 			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, sqIDs)
// 		}
// 	}
// }

// func testStatsGetStatQueueProfileIDs(t *testing.T) {
// 	expIDs := []string{"SQ_1", "SQ_2"}
// 	var sqIDs []string
// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfileIDs,
// 		&utils.PaginatorWithTenant{
// 			Tenant:    "cgrates.org",
// 			Paginator: utils.Paginator{},
// 		}, &sqIDs); err != nil {
// 		t.Error(err)
// 	} else {
// 		sort.Strings(sqIDs)
// 		if !reflect.DeepEqual(sqIDs, expIDs) {
// 			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expIDs, sqIDs)
// 		}
// 	}
// }

// func testStatsGetStatQueueProfileCount(t *testing.T) {
// 	var reply int
// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1GetStatQueueProfileCount,
// 		&utils.TenantWithAPIOpts{
// 			Tenant: "cgrates.org",
// 		}, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != 2 {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>", 2, reply)
// 	}
// }

// func testStatsGetStatQueuesForEvent(t *testing.T) {
// 	args := &engine.StatsArgsProcessEvent{
// 		StatIDs: []string{"SQ_1", "SQ_2"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "StatsEventTest",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 			},
// 		},
// 	}
// 	expSqs := engine.StatQueues{
// 		&engine.StatQueue{
// 			Tenant:    "cgrates.org",
// 			ID:        "SQ_2",
// 			SQMetrics: make(map[string]engine.StatMetric),
// 		},
// 		&engine.StatQueue{
// 			Tenant:    "cgrates.org",
// 			ID:        "SQ_1",
// 			SQMetrics: make(map[string]engine.StatMetric),
// 		},
// 	}

// 	rplySqs := engine.StatQueues{
// 		{
// 			SQMetrics: map[string]engine.StatMetric{
// 				utils.MetaTCD: &engine.StatTCD{},
// 				utils.MetaASR: &engine.StatASR{},
// 				utils.MetaACD: &engine.StatACD{},
// 			},
// 		},
// 		{
// 			SQMetrics: map[string]engine.StatMetric{
// 				utils.MetaTCD: &engine.StatTCD{},
// 				utils.MetaASR: &engine.StatASR{},
// 				utils.MetaACD: &engine.StatACD{},
// 			},
// 		},
// 	}
// 	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueuesForEvent,
// 		args, &rplySqs); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySqs, expSqs) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSqs), utils.ToJSON(rplySqs))
// 	}
// }

// func testStatsRemoveStatQueueProfiles(t *testing.T) {
// 	var reply string

// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
// 		&utils.TenantIDWithAPIOpts{
// 			TenantID: &utils.TenantID{
// 				Tenant: "cgrates.org",
// 				ID:     "SQ_1",
// 			}}, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply returned:", reply)
// 	}

// 	if err := sqRPC.Call(context.Background(), utils.AdminSv1RemoveStatQueueProfile,
// 		&utils.TenantIDWithAPIOpts{
// 			TenantID: &utils.TenantID{
// 				Tenant: "cgrates.org",
// 				ID:     "SQ_2",
// 			}}, &reply); err != nil {
// 		t.Error(err)
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply returned:", reply)
// 	}
// }

// func testStatsGetStatQueuesAfterRemove(t *testing.T) {
// 	args := &engine.StatsArgsProcessEvent{
// 		StatIDs: []string{"SQ_1", "SQ_2"},
// 		CGREvent: &utils.CGREvent{
// 			Tenant: "cgrates.org",
// 			ID:     "StatsEventTest",
// 			Event: map[string]interface{}{
// 				utils.AccountField: "1001",
// 			},
// 		},
// 	}
// 	expSqs := engine.StatQueues{
// 		&engine.StatQueue{
// 			Tenant: "cgrates.org",
// 			ID:     "SQ_2",
// 		},
// 		&engine.StatQueue{
// 			Tenant: "cgrates.org",
// 			ID:     "SQ_1",
// 		},
// 	}

// 	var rplySqs engine.StatQueues
// 	if err := sqRPC.Call(context.Background(), utils.StatSv1GetStatQueuesForEvent,
// 		args, &rplySqs); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(rplySqs, expSqs) {
// 		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
// 			utils.ToJSON(expSqs), utils.ToJSON(rplySqs))
// 	}
// }
