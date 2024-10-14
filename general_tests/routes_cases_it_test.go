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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rtsCaseSv1CfgPath string
	rtsCaseSv1Cfg     *config.CGRConfig
	rtsCaseSv1Rpc     *birpc.Client
	rtsCasePrf        *v1.RouteWithAPIOpts
	rtsCaseSv1ConfDIR string //run tests for specific configuration

	sTestsRtsCaseSV1 = []func(t *testing.T){
		testV1RtsCaseLoadConfig,
		testV1RtsCaseInitDataDb,
		testV1RtsCaseResetStorDb,
		testV1RtsCaseStartEngine,
		testV1RtsCaseRpcConn,
		testV1RtsCaseFromFolder,
		testV1RtsCaseGetRoutesAfterLoading,
		testV1RtsCasesSortingRoutesWeightAccountValue,
		testV1RtsCasesSortingRoutesWeightAllRoutes,
		testV1RtsCasesSortingRoutesWeightNotMatchingValue,
		testV1RtsCasesSortingRoutesLowestCost,
		testV1RtsCasesSortingRoutesLowestCostDefaultUsage,
		testV1RtsCasesSortingRoutesLCSetStatsAndResForMatching,
		testV1RtsCasesSortingRoutesLowestCostStats,
		testV1RtsCasesSortingRoutesLowestCosMatchingAllRoutes,
		testV1RtsCasesSortingRoutesLowestCosMaxCost,
		testV1RtsCasesSortingRoutesLowestCosMaxCostNotMatch,
		testV1RtsCasesSortingRoutesProcessMetrics,
		testV1RtsCasesSortingRoutesQOS,
		testV1RtsCasesSortingRoutesQOSAllRoutes,
		testV1RtsCasesSortingRoutesQOSNotFound,
		testV1RtsCasesSortingRoutesAllocateResources,
		testV1RtsCasesSortingRoutesReasNotAllRoutes,
		testV1RtsCasesSortingRoutesReasAllRoutes,
		testV1RtsCasesRoutesProcessStatsForLoadRtsSorting,
		testV1RtsCasesRoutesLoadRtsSorting,
		testV1RtsCasesSortRoutesHigherCostV2V3,
		testV1RtsCasesSortRoutesHigherCostAllocateRes,
		testV1RtsCasesSortRoutesHigherCostV1V3,
		testV1RtsCasesSortRoutesHigherCostAllRoutes,
		testV1RtsCaseStopEngine,
	}
)

// Test start here
func TestRoutesCaseV1IT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		rtsCaseSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		rtsCaseSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		rtsCaseSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRtsCaseSV1 {
		t.Run(rtsCaseSv1ConfDIR, stest)
	}
}

func testV1RtsCaseLoadConfig(t *testing.T) {
	var err error
	rtsCaseSv1CfgPath = path.Join(*utils.DataDir, "conf", "samples", rtsCaseSv1ConfDIR)
	if rtsCaseSv1Cfg, err = config.NewCGRConfigFromPath(rtsCaseSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RtsCaseInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rtsCaseSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RtsCaseResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rtsCaseSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RtsCaseStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rtsCaseSv1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RtsCaseRpcConn(t *testing.T) {
	rtsCaseSv1Rpc = engine.NewRPCClient(t, rtsCaseSv1Cfg.ListenCfg())
}

func testV1RtsCaseFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutroutes")}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1RtsCaseGetRoutesAfterLoading(t *testing.T) {
	// ROUTE_ACNT_1001
	expRt1 := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1001",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"*string:~*req.Account:1001"},
		Sorting:           "*weight",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:        "vendor1",
				FilterIDs: []string{"FLTR_DEST_1003"},
				Weight:    10,
			},
			{
				ID:        "vendor2",
				FilterIDs: []string{"*gte:~*accounts.1001.BalanceMap.*monetary[0].Value:10"},
				Weight:    20,
			},
			{
				ID:        "vendor3",
				FilterIDs: []string{"FLTR_DEST_1003", "*prefix:~*req.Account:10"},
				Weight:    40,
			},
			{
				ID:     "vendor4",
				Weight: 35,
			},
		},
	}
	var reply *engine.RouteProfile
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1001", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		if !reflect.DeepEqual(expRt1, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt1), utils.ToJSON(reply))
		}
	}

	// ROUTE_ACNT_1002
	expRt2 := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1002",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"*string:~*req.Account:1002"},
		Sorting:           "*lc",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:            "vendor1",
				FilterIDs:     []string{"*lte:~*resources.RES_GRP1.TotalUsage:5"},
				RatingPlanIDs: []string{"RP_VENDOR1"},
				Weight:        0,
			},
			{
				ID:            "vendor2",
				FilterIDs:     []string{"*gte:~*stats.STATS_VENDOR_2.*acd:1m"},
				RatingPlanIDs: []string{"RP_VENDOR2"},
				Weight:        0,
			},
			{
				ID:            "vendor3",
				RatingPlanIDs: []string{"RP_VENDOR2"},
				Weight:        10,
			},
			{
				ID:            "vendor4",
				FilterIDs:     []string{"*ai:~*req.AnswerTime:2013-06-01T00:00:00Z|2013-06-01T10:00:00Z"},
				RatingPlanIDs: []string{"RP_STANDARD"},
				Weight:        30,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1002", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		if !reflect.DeepEqual(expRt2, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt2), utils.ToJSON(reply))
		}
	}

	// ROUTE_ACNT_1003
	expRt3 := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1003",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"*string:~*req.Account:1003"},
		Sorting:           "*qos",
		SortingParameters: []string{"*acd", "*tcc"},
		Routes: []*engine.Route{
			{
				ID:      "vendor1",
				StatIDs: []string{"STATS_VENDOR_1"},
				Weight:  0,
			},
			{
				ID:        "vendor2",
				FilterIDs: []string{"*prefix:~*req.Destination:10"},
				StatIDs:   []string{"STATS_VENDOR_2"},
				Weight:    0,
			},
			{
				ID:        "vendor3",
				FilterIDs: []string{"*gte:~*stats.STATS_VENDOR_1.*tcc:6"},
				StatIDs:   []string{"STATS_VENDOR_1"},
				Weight:    20,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1003", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		sort.Strings(reply.SortingParameters)
		sort.Strings(expRt1.SortingParameters)
		if !reflect.DeepEqual(expRt3, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt3), utils.ToJSON(reply))
		}
	}

	// ROUTE_ACNT_1004
	expRt4 := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1004",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"*string:~*req.Account:1004"},
		Sorting:           "*reas",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:          "vendor1",
				ResourceIDs: []string{"RES_GRP1"},
				Weight:      0,
			},
			{
				ID:          "vendor2",
				ResourceIDs: []string{"RES_GRP2"},
				Weight:      0,
			},
			{
				ID:          "vendor3",
				FilterIDs:   []string{"*gte:~*resources.RES_GRP1.TotalUsage:9"},
				ResourceIDs: []string{"RES_GRP2"},
				Weight:      10,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1004", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		if !reflect.DeepEqual(expRt4, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt4), utils.ToJSON(reply))
		}
	}

	// ROUTE_ACNT_1005
	expRt5 := &engine.RouteProfile{
		ID:                "ROUTE_ACNT_1005",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"*string:~*req.Account:1005"},
		Sorting:           "*load",
		SortingParameters: []string{"vendor1:3", "*default:2"},
		Routes: []*engine.Route{
			{
				ID:      "vendor1",
				StatIDs: []string{"STATS_VENDOR_1:*sum#1"},
			},
			{
				ID:      "vendor2",
				StatIDs: []string{"STATS_VENDOR_2:*sum#1"},
				Weight:  10,
			},
			{
				ID:      "vendor3",
				StatIDs: []string{"STATS_VENDOR_2:*distinct#~*req.Usage"},
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1005", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		sort.Strings(reply.SortingParameters)
		sort.Strings(expRt5.SortingParameters)
		if !reflect.DeepEqual(expRt5, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt5), utils.ToJSON(reply))
		}
	}

	// ROUTE_STATS1
	expRt6 := &engine.RouteProfile{
		ID:                "ROUTE_HC1",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"Fltr_tcc"},
		Sorting:           "*hc",
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:            "route1",
				FilterIDs:     []string{"*gte:~*resources.RES_GRP2.Available:6"},
				RatingPlanIDs: []string{"RP_VENDOR2"},
				ResourceIDs:   []string{"RES_GRP2"},
				Weight:        20,
			},
			{
				ID:            "route2",
				FilterIDs:     []string{"*gte:~*resources.RES_GRP1.TotalUsage:9"},
				RatingPlanIDs: []string{"RP_VENDOR1"},
				ResourceIDs:   []string{"RES_GRP1"},
				Weight:        20,
			},
			{
				ID:            "route3",
				RatingPlanIDs: []string{"RP_VENDOR1"},
				ResourceIDs:   []string{"RES_GRP2"},
				Weight:        10,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_HC1", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
		if !reflect.DeepEqual(expRt6, reply) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRt6), utils.ToJSON(reply))
		}
	}
}

func testV1RtsCasesSortingRoutesWeightAccountValue(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "WEIGHT_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Weight: 20.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesWeightAllRoutes(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "WEIGHT_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1003",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Weight: 40.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Weight: 20.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Weight: 10.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesWeightNotMatchingValue(t *testing.T) {
	//change account 1001 balance for not matching vendor2
	attrBal := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.MetaMonetary,
		Value:       5,
		Balance: map[string]any{
			utils.ID: utils.MetaDefault,
		},
	}
	var result string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.APIerSv1SetBalance, attrBal,
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result returned")
	}

	ev := &utils.CGREvent{
		ID:     "WEIGHT_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1003",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Weight: 40.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Weight: 10.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLowestCost(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.Usage:        "2m30s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1002",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       10.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Cost:         0.2505,
						utils.RatingPlanID: "RP_VENDOR1",
						utils.Weight:       0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	//gonna match one route because the totalUsage by ne-allocated resources is 0
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLowestCostDefaultUsage(t *testing.T) {
	// default usage given by routes is 1m
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1002",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Cost:         0.0498,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       10.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Cost:         0.1002,
						utils.RatingPlanID: "RP_VENDOR1",
						utils.Weight:       0.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Cost:         0.6,
						utils.RatingPlanID: "RP_STANDARD",
						utils.Weight:       30.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	//gonna match one route because the totalUsage by ne-allocated resources is 0
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLCSetStatsAndResForMatching(t *testing.T) {
	//not gonna match our vendor1 filter because 6 > 5
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   6,
		},
	}
	var reply string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "RES_GRP1" {
		t.Errorf("Unexpected reply returned: %s", reply)
	}

	//gonna match one stats for matching vendor 2 acd filter
	var result []string
	expected := []string{"STATS_VENDOR_2", "STATS_TCC1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Category:     "vendor2",
			utils.Usage:        "2m30s",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Cost:         1.0,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, result)
		}
	}
}

func testV1RtsCasesSortingRoutesLowestCostStats(t *testing.T) {
	//not gonna match vendor1 because of its TotalUsage by allocating resources
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "2m30s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1002",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       10.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       0.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Cost:         1.5,
						utils.RatingPlanID: "RP_STANDARD",
						utils.Weight:       30.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	//gonna match one route because the totalUsage by ne-allocated resources is 0
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLowestCosMatchingAllRoutes(t *testing.T) {
	// deallocate resources for matching vendor1
	evRes := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   4,
		},
	}
	var result string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		evRes, &result); err != nil {
		t.Error(err)
	} else if result != "RES_GRP1" {
		t.Errorf("Unexpected result returned: %s", result)
	}

	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "2m30s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1002",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       10.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       0.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Cost:         0.2505,
						utils.RatingPlanID: "RP_VENDOR1",
						utils.Weight:       0.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]any{
						utils.Cost:         1.5,
						utils.RatingPlanID: "RP_STANDARD",
						utils.Weight:       30.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLowestCosMaxCost(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "2m30s",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesMaxCost: 0.35,
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1002",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       10.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Cost:         0.1245,
						utils.RatingPlanID: "RP_VENDOR2",
						utils.Weight:       0.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Cost:         0.2505,
						utils.RatingPlanID: "RP_VENDOR1",
						utils.Weight:       0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesLowestCosMaxCostNotMatch(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Destination:  "1003",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "2m30s",
		},
		APIOpts: map[string]any{
			utils.OptsRoutesMaxCost: 0.05,
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func testV1RtsCasesSortingRoutesProcessMetrics(t *testing.T) {
	//we will process this stats 2 times
	//Vendor2
	expected := []string{"STATS_TCC1", "STATS_VENDOR_2"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Category:     "vendor2",
			utils.Usage:        "2m30s",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Cost:         1.0,
		},
	}
	var result []string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, result)
		}
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, result)
		}
	}

	//Vendor1
	expected = []string{"STATS_TCC1", "STATS_VENDOR_1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Category:     "vendor1",
			utils.Usage:        "2m30s",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Cost:         1.0,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, result)
		}
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, result)
		}
	}
}

func testV1RtsCasesSortingRoutesQOS(t *testing.T) {
	//not gonna match vendor3 because *tcc is not bigger that 6
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "50s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1003",
			Sorting:   "*qos",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.MetaACC:           1.,
						utils.MetaACD:           150. * 1e9,
						"*sum#1":                3.,
						"*distinct#~*req.Usage": 1.,
						utils.MetaTCC:           3.,
						utils.MetaTCD:           450. * 1e9,
						utils.Weight:            0.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.MetaACC: 1.,
						utils.MetaACD: 150. * 1e9,
						"*sum#1":      2.,
						utils.MetaTCC: 2.,
						utils.MetaTCD: 300. * 1e9,
						utils.Weight:  0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesQOSAllRoutes(t *testing.T) {
	// process *tcc metric for matching vendor3
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Category:     "vendor1",
			utils.Usage:        "2m30s",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Cost:         10.0,
		},
	}
	var result []string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &result); err != nil {
		t.Error(err)
	}

	// match all 3 routes
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1003",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "50s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1003",
			Sorting:   "*qos",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.MetaACC: 4.,
						utils.MetaACD: 150. * 1e9,
						"*sum#1":      3.,
						utils.MetaTCC: 12.,
						utils.MetaTCD: 450. * 1e9,
						utils.Weight:  20.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.MetaACC: 4.,
						utils.MetaACD: 150. * 1e9,
						"*sum#1":      3.,
						utils.MetaTCC: 12.,
						utils.MetaTCD: 450. * 1e9,
						utils.Weight:  0.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.MetaACC:           1.,
						utils.MetaACD:           150. * 1e9,
						"*sum#1":                3.,
						"*distinct#~*req.Usage": 1.,
						utils.MetaTCC:           3.,
						utils.MetaTCD:           450. * 1e9,
						utils.Weight:            0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesQOSNotFound(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1008",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.AnswerTime:   "2013-06-01T05:00:00Z",
			utils.Usage:        "50s",
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func testV1RtsCasesSortingRoutesAllocateResources(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   6,
		},
	}
	var reply string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "RES_GRP1" {
		t.Errorf("Unexpected reply returned: %s", reply)
	}

	ev = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1004",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e31",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		ev, &reply); err != nil {
		t.Error(err)
	} else if reply != "RES_GRP2" {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
}

func testV1RtsCasesSortingRoutesReasNotAllRoutes(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Destination:  "1007",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1004",
			Sorting:   "*reas",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.ResourceUsage: 6.0,
						utils.Weight:        0.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.ResourceUsage: 7.0,
						utils.Weight:        0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesReasAllRoutes(t *testing.T) {
	evRs := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   9,
		},
	}
	var replyStr string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		evRs, &replyStr); err != nil {
		t.Error(err)
	} else if replyStr != "RES_GRP1" {
		t.Errorf("Unexpected reply returned: %s", replyStr)
	}
	//allocate more resources for matching
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1004",
			utils.Destination:  "1007",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1004",
			Sorting:   "*reas",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.ResourceUsage: 7.0,
						utils.Weight:        10.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.ResourceUsage: 7.0,
						utils.Weight:        0.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.ResourceUsage: 9.0,
						utils.Weight:        0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesRoutesProcessStatsForLoadRtsSorting(t *testing.T) {
	// "STATS_VENDOR_1"
	var reply []string
	expected := []string{"STATS_VENDOR_1", "STATS_TCC1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			//utils.AccountField: "1004",
			utils.Category:   "vendor1",
			utils.Usage:      "1m20s",
			utils.AnswerTime: "2013-06-01T05:00:00Z",
			utils.Cost:       1.8,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
	// different usage for *distinct metric
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			//utils.AccountField: "1004",
			utils.Category:   "vendor1",
			utils.Usage:      "20s",
			utils.AnswerTime: "2013-06-01T05:00:00Z",
			utils.Cost:       1.8,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}

	// "STATS_VENDOR_2"
	expected = []string{"STATS_VENDOR_2", "STATS_TCC1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			//utils.AccountField: "1004",
			utils.Category:   "vendor2",
			utils.Usage:      "30s",
			utils.AnswerTime: "2013-06-01T05:00:00Z",
			utils.Cost:       0.77,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected)
		sort.Strings(reply)
		if !reflect.DeepEqual(reply, expected) {
			t.Errorf("Expecting: %+v, received: %+v", expected, reply)
		}
	}
}

func testV1RtsCasesRoutesLoadRtsSorting(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1005",
			//utils.Destination:  "1007",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1005",
			Sorting:   "*load",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]any{
						utils.Load:   2.,
						utils.Ratio:  2.,
						utils.Weight: 0.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]any{
						utils.Load:   4.,
						utils.Ratio:  2.,
						utils.Weight: 10.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]any{
						utils.Load:   7.,
						utils.Ratio:  3.,
						utils.Weight: 0.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortRoutesHigherCostV2V3(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1008",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.Usage:        "3m25s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_HC1",
			Sorting:   "*hc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "route2",
					SortingData: map[string]any{
						utils.Cost:          0.34235,
						utils.RatingPlanID:  "RP_VENDOR1",
						utils.ResourceUsage: 9.,
						utils.Weight:        20.,
					},
				},
				{
					RouteID: "route3",
					SortingData: map[string]any{
						utils.Cost:          0.34235,
						utils.RatingPlanID:  "RP_VENDOR1",
						utils.ResourceUsage: 7.,
						utils.Weight:        10.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortRoutesHigherCostAllocateRes(t *testing.T) {
	// to match route 1, RES_GRP2 must have *gte available 6 resources
	// first we have to remove them
	var result string
	evRs := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1004",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e31",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1ReleaseResources,
		evRs, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result returned: %s", result)
	}

	evRs = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1ReleaseResources,
		evRs, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result returned: %s", result)
	}

	evRs = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1004",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e31",
			utils.OptsResourcesUnits:   1,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		evRs, &result); err != nil {
		t.Error(err)
	} else if result != "RES_GRP2" {
		t.Errorf("Unexpected result returned: %s", result)
	}

	// also, to not match route2, totalUsage of RES_GRP1 must be lower than 9
	evRs = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   4,
		},
	}
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		evRs, &result); err != nil {
		t.Error(err)
	} else if result != "RES_GRP1" {
		t.Errorf("Unexpected result returned: %s", result)
	}
}

func testV1RtsCasesSortRoutesHigherCostV1V3(t *testing.T) {
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1008",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.Usage:        "3m25s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_HC1",
			Sorting:   "*hc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "route3",
					SortingData: map[string]any{
						utils.Cost:          0.34235,
						utils.RatingPlanID:  "RP_VENDOR1",
						utils.ResourceUsage: 1.,
						utils.Weight:        10.,
					},
				},
				{
					RouteID: "route1",
					SortingData: map[string]any{
						utils.Cost:          0.17015,
						utils.RatingPlanID:  "RP_VENDOR2",
						utils.ResourceUsage: 1.,
						utils.Weight:        20.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortRoutesHigherCostAllRoutes(t *testing.T) {
	//allocate for matching all routes
	evRs := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account": "1002"},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e51",
			utils.OptsResourcesUnits:   9,
		},
	}
	var result string
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		evRs, &result); err != nil {
		t.Error(err)
	} else if result != "RES_GRP1" {
		t.Errorf("Unexpected result returned: %s", result)
	}
	ev := &utils.CGREvent{
		ID:     "LC_SORT",
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.AccountField: "1008",
			utils.Destination:  "1007",
			utils.SetupTime:    "2013-06-01T00:00:00Z",
			utils.Usage:        "3m25s",
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_HC1",
			Sorting:   "*hc",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "route2",
					SortingData: map[string]any{
						utils.Cost:          0.34235,
						utils.RatingPlanID:  "RP_VENDOR1",
						utils.ResourceUsage: 9.,
						utils.Weight:        20.,
					},
				},
				{
					RouteID: "route3",
					SortingData: map[string]any{
						utils.Cost:          0.34235,
						utils.RatingPlanID:  "RP_VENDOR1",
						utils.ResourceUsage: 1.,
						utils.Weight:        10.,
					},
				},
				{
					RouteID: "route1",
					SortingData: map[string]any{
						utils.Cost:          0.17015,
						utils.RatingPlanID:  "RP_VENDOR2",
						utils.ResourceUsage: 1.,
						utils.Weight:        20.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expSrtdRoutes) {
		t.Errorf("Expecting: %+v \n, received: %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCaseStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
