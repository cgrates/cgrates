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
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/routes"
	"github.com/cgrates/cgrates/utils"
)

var (
	RtStatsSv1CfgPath string
	RtStatsSv1Cfg     *config.CGRConfig
	RtStatsSv1BiRpc   *birpc.Client
	RtStatsSv1ConfDIR string //

	sTestsRtStatsSV1 = []func(t *testing.T){
		testV1RtStatsLoadConfig,
		testV1RtStatsFlushDBs,
		testV1RtStatsStartEngine,
		testV1RtStatsRpcConn,
		testV1RtStatsFromFolder,
		testV1RtStatsProcessStatsValid,
		testV1RtStatsProcessStatsNotAnswered,
		testV1RtStatsGetMetrics,
		testV1RtStatsGetRoutesQOSStrategy,
		testV1RtStatsGetRoutesLowestCostStrategy,
		testV1RtStatsStopEngine,
	}
)

// Test start here
func TestRtStatsCaseV1IT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		RtStatsSv1ConfDIR = "routes_cases_internal"
	case utils.MetaMySQL:
		RtStatsSv1ConfDIR = "routes_cases_mysql"
	case utils.MetaMongo:
		RtStatsSv1ConfDIR = "routes_cases_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRtStatsSV1 {
		t.Run(RtStatsSv1ConfDIR, stest)
	}
}

func testV1RtStatsLoadConfig(t *testing.T) {
	var err error
	RtStatsSv1CfgPath = path.Join(*utils.DataDir, "conf", "samples", RtStatsSv1ConfDIR)
	if RtStatsSv1Cfg, err = config.NewCGRConfigFromPath(context.Background(), RtStatsSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RtStatsFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(RtStatsSv1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(RtStatsSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RtStatsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(RtStatsSv1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RtStatsRpcConn(t *testing.T) {
	RtStatsSv1BiRpc = engine.NewRPCClient(t, RtStatsSv1Cfg.ListenCfg(), *utils.Encoding)
}

func testV1RtStatsFromFolder(t *testing.T) {
	caching := utils.MetaReload
	if RtStatsSv1Cfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			// StopOnError: true,
			APIOpts: map[string]any{utils.MetaCache: caching},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testV1RtStatsProcessStatsValid(t *testing.T) {
	var reply []string
	expected := []string{"STATS_TCC1", "STATS_TOP1", "STATS_TOP2", "STATS_TOP3"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Destination:  "1021",
			utils.Category:     "call",
		},
		APIOpts: map[string]any{
			utils.MetaCost:      1.8,
			utils.MetaStartTime: "2022-04-01T05:00:00Z",
			utils.MetaUsage:     "1m20s",
		},
	}
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
}

func testV1RtStatsProcessStatsNotAnswered(t *testing.T) {
	// not answered means that our event does not have AnsweredTime or *startTime
	var reply []string
	expected := []string{"STATS_TCC1", "STATS_TOP1", "STATS_TOP2"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Destination:  "1021",
		},
		APIOpts: map[string]any{
			utils.MetaCost:  1.8,
			utils.MetaUsage: "26s",
		},
	}
	// we will process this two times
	// 1
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
	// 2
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}

	// process again some stats two times
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Category:     "call",
		},
		APIOpts: map[string]any{
			utils.MetaCost:      1.8,
			utils.MetaUsage:     "50s",
			utils.MetaStartTime: "2022-04-01T05:00:00Z",
		},
	}
	expected = []string{"STATS_TCC1", "STATS_TOP1", "STATS_TOP3"}
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
}

func testV1RtStatsGetMetrics(t *testing.T) {
	expDecimals := map[string]float64{
		utils.MetaACD: 46400000000.,
		utils.MetaASR: 60.,
	}
	var rplyDec map[string]float64
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "STATS_TOP1",
			},
		}, &rplyDec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyDec, expDecimals) {
		t.Errorf("Expected %v, received %v", utils.ToJSON(expDecimals), utils.ToJSON(rplyDec))
	}

	expDecimals = map[string]float64{
		utils.MetaACD: float64(44 * time.Second),
		utils.MetaASR: 33.33333333333333,
	}
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "STATS_TOP2",
			},
		}, &rplyDec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expDecimals, rplyDec) {
		t.Errorf("Expected %v, received %v", utils.ToJSON(expDecimals), utils.ToJSON(rplyDec))
	}

	expDecimals = map[string]float64{
		utils.MetaACD: float64(time.Minute),
		utils.MetaASR: 100,
	}
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "STATS_TOP3",
			},
		}, &rplyDec); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyDec, expDecimals) {
		t.Errorf("Expected %v, received %v", utils.ToJSON(expDecimals), utils.ToJSON(rplyDec))
	}
}

func testV1RtStatsGetRoutesQOSStrategy(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "10015",
			utils.Destination:  "+33426654",
		},
	}
	expSrtdRoutes := &routes.SortedRoutesList{
		{
			ProfileID: "ROUTE_QOS_STATS",
			Sorting:   "*qos",
			Routes: []*routes.SortedRoute{
				{
					RouteID: "route1",
					SortingData: map[string]any{
						utils.MetaACD: 46400000000.,
						utils.MetaASR: 60.,
						utils.Weight:  20.,
					},
				},
				{
					RouteID: "route2",
					SortingData: map[string]any{
						utils.MetaACD: 44000000000.,
						utils.MetaASR: 33.33333333333333,
						utils.Weight:  50.,
					},
				},
			},
		},
	}
	var reply *routes.SortedRoutesList
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtStatsGetRoutesLowestCostStrategy(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "10015",
			utils.Destination:  "+2273676400",
		},
	}
	expSrtdRoutes := &routes.SortedRoutesList{
		{
			ProfileID: "ROUTE_LCR",
			Sorting:   "*lc",
			Routes: []*routes.SortedRoute{
				{
					RouteID: "route3",
					SortingData: map[string]any{
						utils.Cost:          0.05,
						utils.RateProfileID: "RP_VENDOR2",
						utils.Weight:        10.,
					},
				},
				{
					RouteID: "route1",
					SortingData: map[string]any{
						utils.Cost:          0.1,
						utils.RateProfileID: "RP_VENDOR1",
						utils.Weight:        20.,
					},
				},
				{
					RouteID: "route2",
					SortingData: map[string]any{
						utils.Cost:          0.6,
						utils.RateProfileID: "RP_STANDARD",
						utils.Weight:        15.,
					},
				},
			},
		},
	}
	var reply *routes.SortedRoutesList
	if err := RtStatsSv1BiRpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtStatsStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
