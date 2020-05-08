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
	splSv1CfgPath string
	splSv1Cfg     *config.CGRConfig
	splSv1Rpc     *rpc.Client
	splPrf        *RouteWithCache
	splSv1ConfDIR string //run tests for specific configuration

	sTestsRouteSV1 = []func(t *testing.T){
		testV1SplSLoadConfig,
		testV1SplSInitDataDb,
		testV1SplSResetStorDb,
		testV1SplSStartEngine,
		testV1SplSRpcConn,
		testV1SplSFromFolder,
		testV1SplSGetWeightRoutes,
		testV1SplSGetLeastCostRoutes,
		testV1SplSGetLeastCostRoutesWithoutUsage,
		testV1SplSGetLeastCostRoutesWithMaxCost,
		testV1SplSGetLeastCostRoutesWithMaxCost2,
		testV1SplSGetLeastCostRoutesWithMaxCostNotFound,
		testV1SplSGetHighestCostRoutes,
		testV1SplSGetLeastCostRoutesErr,
		testV1SplSPolulateStatsForQOS,
		testV1SplSGetQOSRoutes,
		testV1SplSGetQOSRoutes2,
		testV1SplSGetQOSRoutes3,
		testV1SplSGetQOSRoutesFiltred,
		testV1SplSGetQOSRoutesFiltred2,
		testV1SplSGetRouteWithoutFilter,
		testV1SplSSetRouteProfiles,
		testV1SplSGetRouteProfileIDs,
		testV1SplSUpdateRouteProfiles,
		testV1SplSRemRouteProfiles,
		testV1SplSGetRouteForEvent,
		// reset the database and load the TP again
		testV1SplSInitDataDb,
		testV1SplSFromFolder,
		testV1SplsOneRouteWithoutDestination,
		testV1SplRoutePing,
		testV1SplSStopEngine,
	}
)

// Test start here
func TestSuplSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		splSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		splSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		splSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRouteSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*dataDir, "conf", "samples", splSv1ConfDIR)
	if splSv1Cfg, err = config.NewCGRConfigFromPath(splSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1SplSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1SplSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(splSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	var err error
	splSv1Rpc, err = newRPCClient(splSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1SplSGetWeightRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetWeightRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1007",
				utils.Destination: "+491511231234",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2667,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostRoutesWithoutUsage(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostRoutesWithMaxCost(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: "0.30",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostRoutesWithMaxCostNotFound(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: "0.001",
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSGetLeastCostRoutesWithMaxCost2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		MaxCost: utils.MetaEventCost,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "SPECIAL_1002",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2014, 01, 14, 0, 0, 0, 0, time.UTC),
				utils.Usage:       "10m20s",
				utils.Category:    "call",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetHighestCostRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
				"DistinctMatch":   "*highest_cost",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_HIGHESTCOST_1",
		Sorting:   utils.MetaHC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2667,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostRoutesErr(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		IgnoreErrors: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1000",
				utils.Destination: "1001",
				utils.SetupTime:   "*now",
				"Subject":         "TEST",
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSPolulateStatsForQOS(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       10.0,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       10.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Account:    "1002",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(5 * time.Second),
				utils.COST:       12.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Account:    "1002",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(6 * time.Second),
				utils.COST:       17.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Account:    "1003",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       12.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       12.5,
				utils.PDD:        time.Duration(12 * time.Second),
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
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
				utils.Usage:      time.Duration(15 * time.Second),
				utils.COST:       15.5,
				utils.PDD:        time.Duration(15 * time.Second),
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1SplSGetQOSRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos",
			},
		},
	}
	expRouteIDs := []string{"supplier1", "supplier3", "supplier2"}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "SPL_QOS_1" {
			t.Errorf("Expecting: SPL_QOS_1, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSRoutes2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos2",
			},
		},
	}
	expRouteIDs := []string{"supplier3", "supplier2", "supplier1"}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "SPL_QOS_2" {
			t.Errorf("Expecting: SPL_QOS_2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSRoutes3(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos3",
			},
		},
	}
	expRouteIDs := []string{"supplier1", "supplier3", "supplier2"}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "SPL_QOS_3" {
			t.Errorf("Expecting: SPL_QOS_3, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSRoutesFiltred(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos_filtred",
			},
		},
	}
	expRouteIDs := []string{"supplier1", "supplier3"}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "SPL_QOS_FILTRED" {
			t.Errorf("Expecting: SPL_QOS_FILTRED, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSGetQOSRoutesFiltred2(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSRoutes",
			Event: map[string]interface{}{
				"DistinctMatch":   "*qos_filtred2",
				utils.Account:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	expRouteIDs := []string{"supplier3", "supplier2"}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedRoutes))
		for i, supl := range suplsReply.SortedRoutes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply.ProfileID != "SPL_QOS_FILTRED2" {
			t.Errorf("Expecting: SPL_QOS_FILTRED2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expRouteIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expRouteIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetRouteWithoutFilter(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetRouteWithoutFilter",
			Event: map[string]interface{}{
				utils.Account:     "1008",
				utils.Destination: "+49",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		Count:     1,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedRoutes
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSSetRouteProfiles(t *testing.T) {
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:            "cgrates.org",
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"FLTR_1"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{
				{
					ID:              "SPL1",
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
	if err := splSv1Rpc.Call(utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.RouteProfile, reply)
	}
}

func testV1SplSGetRouteProfileIDs(t *testing.T) {
	expected := []string{"SPL_HIGHESTCOST_1", "SPL_QOS_1", "SPL_QOS_2", "SPL_QOS_FILTRED", "SPL_QOS_FILTRED2",
		"SPL_ACNT_1001", "SPL_LEASTCOST_1", "SPL_WEIGHT_2", "SPL_WEIGHT_1", "SPL_QOS_3",
		"TEST_PROFILE1", "SPL_LOAD_DIST", "SPL_LCR"}
	var result []string
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfileIDs,
		&utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{"cgrates.org"}}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1SplSUpdateRouteProfiles(t *testing.T) {
	splPrf.Routes = []*engine.Route{
		{
			ID:              "SPL1",
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
			ID:              "SPL2",
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
			ID:              "SPL2",
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
			ID:              "SPL1",
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
	if err := splSv1Rpc.Call(utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.Routes, reply.Routes) && !reflect.DeepEqual(reverseRoutes, reply.Routes) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testV1SplSRemRouteProfiles(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testV1SplRoutePing(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call(utils.RouteSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1SplSGetRouteForEvent(t *testing.T) {
	ev := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostRoutes",
			Event: map[string]interface{}{
				utils.Account:     "1000",
				utils.Destination: "1001",
				utils.SetupTime:   "*now",
				"Subject":         "TEST",
			},
		},
	}
	expected := engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_LCR",
		FilterIDs: []string{"FLTR_TEST"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 00, 00, 00, 00, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			&engine.Route{
				ID:              "supplier_1",
				FilterIDs:       nil,
				AccountIDs:      nil,
				RatingPlanIDs:   []string{"RP_TEST_1"},
				ResourceIDs:     nil,
				StatIDs:         nil,
				Weight:          10,
				Blocker:         false,
				RouteParameters: "",
			},
			&engine.Route{
				ID:              "supplier_2",
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
	if err := splSv1Rpc.Call(utils.RouteSv1GetRouteProfilesForEvent,
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
}

// Scenario: We create two rating plans RP_MOBILE and RP_LOCAL
// RP_LOCAL contains destination for both mobile and local
// and RP_MOBILE contains destinations only for mobile
// Create a RouteProfile with *least_cost strategy with 2 suppliers
// supplier1 have attached RP_LOCAL and supplier2 have attach RP_MOBILE
func testV1SplsOneRouteWithoutDestination(t *testing.T) {
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_DESTINATION"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "SPL_DESTINATION",
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
	if err := splSv1Rpc.Call(utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplsOneRouteWithoutDestination",
			Event: map[string]interface{}{
				utils.Account:     "SpecialCase",
				utils.Destination: "+24680",
				utils.SetupTime:   utils.MetaNow,
				utils.Usage:       "2m",
			},
		},
	}
	eSpls := engine.SortedRoutes{
		ProfileID: "SPL_DESTINATION",
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
	if err := splSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
