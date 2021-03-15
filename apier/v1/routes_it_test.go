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
package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	routeSv1CfgPath string
	routeSv1Cfg     *config.CGRConfig
	routeSv1Rpc     *rpc.Client
	routePrf        *RouteWithCache
	routeSv1ConfDIR string //run tests for specific configuration

	sTestsRouteSV1 = []func(t *testing.T){
		testV1RouteLoadConfig,
		testV1RouteInitDataDb,
		testV1RouteResetStorDb,
		testV1RouteStartEngine,
		testV1RouteRpcConn,
		testV1RouteGetBeforeDataLoad,
		testV1RouteFromFolder,
		testV1RouteGetWeightRoutes,
		testV1RouteGetLeastCostRoutes,
		testV1RouteGetLeastCostRoutesWithoutUsage,
		testV1RouteGetLeastCostRoutesWithMaxCost,
		testV1RouteGetLeastCostRoutesWithMaxCost2,
		testV1RouteGetLeastCostRoutesWithMaxCostNotFound,
		testV1RouteGetHighestCostRoutes,
		testV1RouteGetLeastCostRoutesErr,
		testV1RoutePolulateStatsForQOS,
		testV1RouteGetQOSRoutes,
		testV1RouteGetQOSRoutes2,
		testV1RouteGetQOSRoutes3,
		testV1RouteGetQOSRoutesFiltred,
		testV1RouteGetQOSRoutesFiltred2,
		testV1RouteGetRouteWithoutFilter,
		testV1RouteSetRouteProfiles,
		testV1RouteGetRouteProfileIDs,
		testV1RouteUpdateRouteProfiles,
		testV1RouteRemRouteProfiles,
		testV1RouteGetRouteForEvent,
		testV1RouteSetRouteProfilesWithoutTenant,
		testV1RouteRemRouteProfilesWithoutTenant,
		// reset the database and load the TP again
		testV1RouteInitDataDb,
		testV1RouteFromFolder,
		testV1RoutesOneRouteWithoutDestination,
		testV1RouteRoutePing,
		testV1RouteMultipleRouteSameID,
		testV1RouteAccountWithRatingPlan,
		testV1RouteStopEngine,
	}
)

// Test start here
func TestRouteSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		routeSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		routeSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		routeSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRouteSV1 {
		t.Run(routeSv1ConfDIR, stest)
	}
}

func testV1RouteLoadConfig(t *testing.T) {
	var err error
	routeSv1CfgPath = path.Join(*dataDir, "conf", "samples", routeSv1ConfDIR)
	if routeSv1Cfg, err = config.NewCGRConfigFromPath(routeSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RouteInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(routeSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RouteResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(routeSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RouteStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(routeSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RouteRpcConn(t *testing.T) {
	var err error
	routeSv1Rpc, err = newRPCClient(routeSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RouteGetBeforeDataLoad(t *testing.T) {
	var suplsReply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "ROUTE_WEIGHT_1",
		}, &suplsReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RouteFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := routeSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1RouteGetWeightRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetWeightRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1007",
				utils.Destination:  "+491511231234",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}

	ev.CGREvent.Tenant = utils.EmptyString
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetLeastCostRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Subject:      "1003",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:        "1m20s",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2667,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetLeastCostRoutesWithoutUsage(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Subject:      "1003",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetLeastCostRoutesWithMaxCost(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: "0.30",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:        "1m20s",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetLeastCostRoutesWithMaxCostNotFound(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: "0.001",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:        "1m20s",
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RouteGetLeastCostRoutesWithMaxCost2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: utils.MetaEventCost,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Subject:      "SPECIAL_1002",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2014, 01, 14, 0, 0, 0, 0, time.UTC),
				utils.Usage:        "10m20s",
				utils.Category:     "call",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetHighestCostRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetHighestCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:        "1m20s",
				"DistinctMatch":    "*highest_cost",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_HIGHESTCOST_1",
		Sorting:   utils.MetaHC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2667,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteGetLeastCostRoutesErr(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		IgnoreErrors: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetHighestCostRoutes",
			Event: map[string]interface{}{
				utils.AccountField: "1000",
				utils.Destination:  "1001",
				utils.SetupTime:    "*now",
				"Subject":          "TEST",
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RoutePolulateStatsForQOS(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        11 * time.Second,
				utils.Cost:         10.0,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        11 * time.Second,
				utils.Cost:         10.5,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        5 * time.Second,
				utils.Cost:         12.5,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        6 * time.Second,
				utils.Cost:         17.5,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_3"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.AccountField: "1003",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        11 * time.Second,
				utils.Cost:         12.5,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Stat":           "Stat1_1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      11 * time.Second,
				utils.Cost:       12.5,
				utils.PDD:        12 * time.Second,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Stat":           "Stat1_1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      15 * time.Second,
				utils.Cost:       15.5,
				utils.PDD:        15 * time.Second,
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1RouteGetQOSRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos",
			},
		},
	}
	expRouteIDs := []string{"route1", "route3", "route2"}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "ROUTE_QOS_1" {
			t.Errorf("Expecting: ROUTE_QOS_1, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1RouteGetQOSRoutes2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos2",
			},
		},
	}
	expRouteIDs := []string{"route3", "route2", "route1"}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "ROUTE_QOS_2" {
			t.Errorf("Expecting: ROUTE_QOS_2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1RouteGetQOSRoutes3(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos3",
			},
		},
	}
	expRouteIDs := []string{"route1", "route3", "route2"}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "ROUTE_QOS_3" {
			t.Errorf("Expecting: ROUTE_QOS_3, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1RouteGetQOSRoutesFiltred(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos_filtred",
			},
		},
	}
	expRouteIDs := []string{"route1", "route3"}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "ROUTE_QOS_FILTRED" {
			t.Errorf("Expecting: ROUTE_QOS_FILTRED, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1RouteGetQOSRoutesFiltred2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch":    "*qos_filtred2",
				utils.AccountField: "1003",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:        "1m20s",
			},
		},
	}
	expRouteIDs := []string{"route3", "route2"}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "ROUTE_QOS_FILTRED2" {
			t.Errorf("Expecting: ROUTE_QOS_FILTRED2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1RouteGetRouteWithoutFilter(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RouteGetRouteWithoutFilter",
			Event: map[string]interface{}{
				utils.AccountField: "1008",
				utils.Destination:  "+49",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteSetRouteProfiles(t *testing.T) {
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"FLTR_NotFound"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{
				{
					ID:              "ROUTE1",
					RatingPlanIDs:   []string{"RP1"},
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc"},
					ResourceIDs:     []string{"Res1", "ResGroup2"},
					StatIDs:         []string{"Stat1"},
					Weight:          20,
					Blocker:         false,
					RouteParameters: "SortingParameter1",
				},
			},
			Weight: 10,
		},
	}

	var result string
	expErr := "SERVER_ERROR: broken reference to filter: FLTR_NotFound for item with ID: cgrates.org:TEST_PROFILE1"
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}

	var reply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	routePrf.FilterIDs = []string{"FLTR_1"}
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(routePrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", routePrf.RouteProfile, reply)
	}
}

func testV1RouteGetRouteProfileIDs(t *testing.T) {
	expected := []string{"ROUTE_HIGHESTCOST_1", "ROUTE_QOS_1", "ROUTE_QOS_2", "ROUTE_QOS_FILTRED", "ROUTE_QOS_FILTRED2",
		"ROUTE_ACNT_1001", "ROUTE_LEASTCOST_1", "ROUTE_WEIGHT_2", "ROUTE_WEIGHT_1", "ROUTE_QOS_3",
		"TEST_PROFILE1", "ROUTE_LOAD_DIST", "ROUTE_LCR"}
	var result []string
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfileIDs,
		&utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}

	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfileIDs,
		&utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1RouteUpdateRouteProfiles(t *testing.T) {
	routePrf.Routes = []*engine.Route{
		{
			ID:              "ROUTE1",
			RatingPlanIDs:   []string{"RP1"},
			FilterIDs:       []string{"FLTR_1"},
			AccountIDs:      []string{"Acc"},
			ResourceIDs:     []string{"Res1", "ResGroup2"},
			StatIDs:         []string{"Stat1"},
			Weight:          20,
			Blocker:         false,
			RouteParameters: "SortingParameter1",
		},
		{
			ID:              "ROUTE2",
			RatingPlanIDs:   []string{"RP2"},
			FilterIDs:       []string{"FLTR_2"},
			AccountIDs:      []string{"Acc"},
			ResourceIDs:     []string{"Res2", "ResGroup2"},
			StatIDs:         []string{"Stat2"},
			Weight:          20,
			Blocker:         true,
			RouteParameters: "SortingParameter2",
		},
	}
	reverseRoutes := []*engine.Route{
		{
			ID:              "ROUTE2",
			RatingPlanIDs:   []string{"RP2"},
			FilterIDs:       []string{"FLTR_2"},
			AccountIDs:      []string{"Acc"},
			ResourceIDs:     []string{"Res2", "ResGroup2"},
			StatIDs:         []string{"Stat2"},
			Weight:          20,
			Blocker:         true,
			RouteParameters: "SortingParameter2",
		},
		{
			ID:              "ROUTE1",
			RatingPlanIDs:   []string{"RP1"},
			FilterIDs:       []string{"FLTR_1"},
			AccountIDs:      []string{"Acc"},
			ResourceIDs:     []string{"Res1", "ResGroup2"},
			StatIDs:         []string{"Stat1"},
			Weight:          20,
			Blocker:         false,
			RouteParameters: "SortingParameter1",
		},
	}
	var result string
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(routePrf.Routes, reply.Routes) && !reflect.DeepEqual(reverseRoutes, reply.Routes) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(routePrf), utils.ToJSON(reply))
	}
}

func testV1RouteRemRouteProfiles(t *testing.T) {
	var resp string
	if err := routeSv1Rpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := routeSv1Rpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testV1RouteRoutePing(t *testing.T) {
	var resp string
	if err := routeSv1Rpc.Call(utils.RouteSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RouteGetRouteForEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1RouteGetHighestCostRoutes",
		Event: map[string]interface{}{
			utils.AccountField: "1000",
			utils.Destination:  "1001",
			utils.SetupTime:    "*now",
			utils.Subject:      "TEST",
		},
	}
	expected := engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_LCR",
		FilterIDs: []string{"FLTR_TEST"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 00, 00, 00, 00, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:              "route_1",
				FilterIDs:       nil,
				AccountIDs:      nil,
				RatingPlanIDs:   []string{"RP_TEST_1"},
				ResourceIDs:     nil,
				StatIDs:         nil,
				Weight:          10,
				Blocker:         false,
				RouteParameters: "",
			},
			{
				ID:              "route_2",
				FilterIDs:       nil,
				AccountIDs:      nil,
				RatingPlanIDs:   []string{"RP_TEST_2"},
				ResourceIDs:     nil,
				StatIDs:         nil,
				Weight:          0,
				Blocker:         false,
				RouteParameters: "",
			},
		},
		Weight: 50,
	}
	if *encoding == utils.MetaGOB { // in gob emtpty slice is encoded as nil
		expected.SortingParameters = nil
	}
	var supProf []*engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRouteProfilesForEvent,
		ev, &supProf); err != nil {
		t.Fatal(err)
	}
	sort.Slice(expected.Routes, func(i, j int) bool {
		return expected.Routes[i].Weight < expected.Routes[j].Weight
	})
	sort.Slice(supProf[0].Routes, func(i, j int) bool {
		return supProf[0].Routes[i].Weight < supProf[0].Routes[j].Weight
	})
	if !reflect.DeepEqual(expected, *supProf[0]) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(supProf))
	}

	supProf = nil
	ev.Tenant = utils.EmptyString
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRouteProfilesForEvent,
		ev, &supProf); err != nil {
		t.Fatal(err)
	}
	sort.Slice(expected.Routes, func(i, j int) bool {
		if expected.Routes[i].ID != expected.Routes[j].ID {
			return expected.Routes[i].ID < expected.Routes[j].ID
		}
		return expected.Routes[i].Weight < expected.Routes[j].Weight
	})
	sort.Slice(supProf[0].Routes, func(i, j int) bool {
		if supProf[0].Routes[i].ID != supProf[0].Routes[j].ID {
			return supProf[0].Routes[i].ID < supProf[0].Routes[j].ID
		}
		return supProf[0].Routes[i].Weight < supProf[0].Routes[j].Weight
	})
	if !reflect.DeepEqual(&expected, supProf[0]) {
		t.Errorf("Expected: %s \n,received: %s", utils.ToJSON(expected), utils.ToJSON(supProf[0]))
	}
}

// Scenario: We create two rating plans RP_MOBILE and RP_LOCAL
// RP_LOCAL contains destination for both mobile and local
// and RP_MOBILE contains destinations only for mobile
// Create a RouteProfile with *least_cost strategy with 2 routes
// route1 have attached RP_LOCAL and route2 have attach RP_MOBILE
func testV1RoutesOneRouteWithoutDestination(t *testing.T) {
	var reply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_DESTINATION"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "ROUTE_DESTINATION",
			FilterIDs: []string{"*string:~*req.Account:SpecialCase"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:            "local",
					RatingPlanIDs: []string{"RP_LOCAL"},
					Weight:        10,
				},
				{
					ID:            "mobile",
					RatingPlanIDs: []string{"RP_MOBILE"},
					FilterIDs:     []string{"*destinations:~*req.Destination:DST_MOBILE"},
					Weight:        10,
				},
			},
			Weight: 100,
		},
	}

	var result string
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1RoutesOneRouteWithoutDestination",
			Event: map[string]interface{}{
				utils.AccountField: "SpecialCase",
				utils.Destination:  "+24680",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "2m",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "ROUTE_DESTINATION",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "local",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0396,
					"RatingPlanID": "RP_LOCAL",
					utils.Weight:   10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteMultipleRouteSameID(t *testing.T) {
	var reply *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "MULTIPLE_ROUTES"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "MULTIPLE_ROUTES",
			FilterIDs: []string{"*string:~*req.Account:SpecialCase2"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:            "Route1",
					RatingPlanIDs: []string{"RP_LOCAL"},
					FilterIDs:     []string{"*string:~*req.Month:April"},
					Weight:        10,
				},
				{
					ID:            "Route1",
					RatingPlanIDs: []string{"RP_MOBILE"},
					FilterIDs:     []string{"*string:~*req.Month:May"},
					Weight:        10,
				},
			},
			Weight: 100,
		},
	}

	var result string
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tNow := time.Now()
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteMultipleRouteSameID",
			Event: map[string]interface{}{
				utils.AccountField: "SpecialCase2",
				utils.Destination:  "+135876",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "2m",
				"Month":            "April",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "MULTIPLE_ROUTES",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "Route1",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0396,
					"RatingPlanID": "RP_LOCAL",
					utils.Weight:   10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}

	ev = &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteMultipleRouteSameID",
			Event: map[string]interface{}{
				utils.AccountField: "SpecialCase2",
				utils.Destination:  "+135876",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "2m",
				"Month":            "May",
			},
		},
	}
	eSpls = engine.SortedRoutes{
		ProfileID: "MULTIPLE_ROUTES",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "Route1",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0204,
					"RatingPlanID": "RP_MOBILE",
					utils.Weight:   10.0,
				},
			},
		},
	}
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1RouteAccountWithRatingPlan(t *testing.T) {
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteWithAccAndRP",
			FilterIDs: []string{"*string:~*req.EventType:testV1RouteAccountWithRatingPlan"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:            "RouteWithAccAndRP",
					AccountIDs:    []string{"AccWithVoice"},
					RatingPlanIDs: []string{"RP_ANY2CNT_SEC"},
					Weight:        20,
				},
				{
					ID:            "RouteWithRP",
					RatingPlanIDs: []string{"RP_ANY1CNT_SEC"},
					Weight:        10,
				},
			},
			Weight: 100,
		},
	}

	var result string
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "AccWithVoice",
		BalanceType: utils.MetaVoice,
		Value:       30 * float64(time.Second),
		Balance: map[string]interface{}{
			utils.ID: "VoiceBalance",
		},
	}
	var reply string
	if err := routeSv1Rpc.Call(utils.APIerSv2SetBalance, &attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "AccWithVoice",
	}
	if err := routeSv1Rpc.Call(utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != 30*float64(time.Second) {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	// test for 30 seconds usage
	// we expect that the route with account to have cost 0
	tNow := time.Now()
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteAccountWithRatingPlan",
			Event: map[string]interface{}{
				utils.AccountField: "RandomAccount",
				utils.Destination:  "+135876",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "30s",
				"EventType":        "testV1RouteAccountWithRatingPlan",
			},
		},
	}
	eSpls := &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRP",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.0,
					"MaxUsage":         30000000000.0,
					utils.Weight:       20.0,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.3,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes = []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.,
					"MaxUsage":         30 * time.Second,
					utils.Weight:       20.,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.3,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.,
				},
			},
		}
	}
	var suplsReply *engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s \n received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}

	// test for 60 seconds usage
	// 30 seconds are covered by account and the remaining will be calculated
	ev = &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteAccountWithRatingPlan",
			Event: map[string]interface{}{
				utils.AccountField: "RandomAccount",
				utils.Destination:  "+135876",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "60s",
				"EventType":        "testV1RouteAccountWithRatingPlan",
			},
		},
	}
	eSpls = &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRP",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.6,
					"MaxUsage":         30000000000.0,
					"RatingPlanID":     "RP_ANY2CNT_SEC",
					utils.Weight:       20.0,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.6,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes = []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.6,
					"MaxUsage":         30 * time.Second,
					"RatingPlanID":     "RP_ANY2CNT_SEC",
					utils.Weight:       20.,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.6,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.,
				},
			},
		}
	}
	var routeRply *engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &routeRply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, routeRply) {
		t.Errorf("Expecting: %s \n received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(routeRply))
	}

	// test for 61 seconds usage
	// 30 seconds are covered by account and the remaining will be calculated
	ev = &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteAccountWithRatingPlan",
			Event: map[string]interface{}{
				utils.AccountField: "RandomAccount",
				utils.Destination:  "+135876",
				utils.SetupTime:    utils.MetaNow,
				utils.Usage:        "1m1s",
				"EventType":        "testV1RouteAccountWithRatingPlan",
			},
		},
	}
	eSpls = &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRP",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.61,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.0,
				},
			},
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.62,
					"MaxUsage":         30000000000.0,
					"RatingPlanID":     "RP_ANY2CNT_SEC",
					utils.Weight:       20.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes = []*engine.SortedRoute{
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:     0.61,
					"RatingPlanID": "RP_ANY1CNT_SEC",
					utils.Weight:   10.,
				},
			},
			{
				RouteID: "RouteWithAccAndRP",
				SortingData: map[string]interface{}{
					utils.AccountField: "AccWithVoice",
					utils.Cost:         0.62,
					"MaxUsage":         30 * time.Second,
					"RatingPlanID":     "RP_ANY2CNT_SEC",
					utils.Weight:       20.,
				},
			},
		}
	}
	var routeRply2 *engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &routeRply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, routeRply2) {
		t.Errorf("Expecting: %s \n received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(routeRply2))
	}

}

func testV1RouteStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testV1RouteSetRouteProfilesWithoutTenant(t *testing.T) {
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "TEST_PROFILE10",
			FilterIDs:         []string{"FLTR_1"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{
				{
					ID:              "ROUTE1",
					RatingPlanIDs:   []string{"RP1"},
					FilterIDs:       []string{"FLTR_1"},
					AccountIDs:      []string{"Acc"},
					ResourceIDs:     []string{"Res1", "ResGroup2"},
					StatIDs:         []string{"Stat1"},
					Weight:          20,
					Blocker:         false,
					RouteParameters: "SortingParameter1",
				},
			},
			Weight: 10,
		},
	}
	var reply string
	if err := routeSv1Rpc.Call(utils.APIerSv1SetRouteProfile, routePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	routePrf.Tenant = "cgrates.org"
	var result *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "TEST_PROFILE10"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, routePrf.RouteProfile) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(routePrf.RouteProfile), utils.ToJSON(result))
	}
}

func testV1RouteRemRouteProfilesWithoutTenant(t *testing.T) {
	var reply string
	if err := routeSv1Rpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "TEST_PROFILE10"}},
		&reply); err != nil {
		t.Error(err)
	}
	var result *engine.RouteProfile
	if err := routeSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "TEST_PROFILE10"},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
