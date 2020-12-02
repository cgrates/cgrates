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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestsRouteSWithRateSV1 = []func(t *testing.T){
		testV1RouteSWithRateSLoadConfig,
		testV1RouteSWithRateSInitDataDb,
		testV1RouteSWithRateSResetStorDb,
		testV1RouteSWithRateSStartEngine,
		testV1RouteSWithRateSRpcConn,
		testV1RouteSWithRateSFromFolder,
		testV1RouteSWithRateSGetRoutes,
		testV1RouteSWithRateSAccountWithRateProfile,
		testV1RouteSWithRateSWithEmptyRateProfileIDs,
		testV1RouteSWithRateSStopEngine,
	}
)

// Test start here
func TestRouteSWithRateSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		routeSv1ConfDIR = "routes_with_rates_internal"
	case utils.MetaMySQL:
		routeSv1ConfDIR = "routes_with_rates_redis"
	case utils.MetaMongo:
		routeSv1ConfDIR = "routes_with_rates_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRouteSWithRateSV1 {
		t.Run(routeSv1ConfDIR, stest)
	}
}

func testV1RouteSWithRateSLoadConfig(t *testing.T) {
	var err error
	routeSv1CfgPath = path.Join(*dataDir, "conf", "samples", routeSv1ConfDIR)
	if routeSv1Cfg, err = config.NewCGRConfigFromPath(routeSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RouteSWithRateSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(routeSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RouteSWithRateSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(routeSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RouteSWithRateSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(routeSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RouteSWithRateSRpcConn(t *testing.T) {
	var err error
	routeSv1Rpc, err = newRPCClient(routeSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RouteSWithRateSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "routes_with_rates")}
	if err := routeSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1RouteSWithRateSGetRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testV1RouteSWithRateSGetRoutes",
				Event: map[string]interface{}{
					utils.Account:     "1003",
					utils.Subject:     "1003",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
					utils.Usage:       "1m20s",
					"EventName":       "RouteWithRateS",
				},
			},
		},
	}
	expRoutes := engine.SortedRoutes{
		ProfileID: "ROUTE_LC",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "route3",
				SortingData: map[string]interface{}{
					utils.Cost:               0.01333333333333334,
					utils.RateProfileMatched: "RT_SPECIAL_1002",
					utils.Weight:             15.0,
				},
			},
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:               0.01333333333333334,
					utils.RateProfileMatched: "RT_SPECIAL_1002",
					utils.Weight:             10.0,
				},
			},
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Cost:               0.4666666666666667,
					utils.RateProfileMatched: "RT_RETAIL1",
					utils.Weight:             20.0,
				},
			},
		},
	}
	var routesReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &routesReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expRoutes, routesReply) {
		t.Errorf("Expecting: %s, \n received: %s",
			utils.ToJSON(expRoutes), utils.ToJSON(routesReply))
	}
}

func testV1RouteSWithRateSAccountWithRateProfile(t *testing.T) {
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteWithAccAndRatePrf",
			FilterIDs: []string{"*string:~*req.EventType:testV1RouteSWithRateSAccountWithRateProfile"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:             "RouteWithAccAndRatePrf",
					AccountIDs:     []string{"AccWithVoice"},
					RateProfileIDs: []string{"RT_ANY2CNT_SEC"},
					Weight:         20,
				},
				{
					ID:             "RouteWithRP",
					RateProfileIDs: []string{"RT_ANY1CNT_SEC"},
					Weight:         10,
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
		BalanceType: utils.VOICE,
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
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != 30*float64(time.Second) {
		t.Errorf("Unexpected balance received : %+v", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}

	// test for 30 seconds usage
	// we expect that the route with account to have cost 0
	tNow := time.Now()
	ev := &engine.ArgsGetRoutes{
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Time:   &tNow,
				ID:     "testV1RouteAccountWithRatingPlan",
				Event: map[string]interface{}{
					utils.Account:     "RandomAccount",
					utils.Destination: "+135876",
					utils.SetupTime:   utils.MetaNow,
					utils.Usage:       "30s",
					"EventType":       "testV1RouteSWithRateSAccountWithRateProfile",
				},
			},
		},
	}
	eSpls := &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRatePrf",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRatePrf",
				SortingData: map[string]interface{}{
					utils.Account:     "AccWithVoice",
					utils.Cost:        0.0,
					utils.CapMaxUsage: 30000000000.0,
					utils.Weight:      20.0,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:               0.3,
					utils.RateProfileMatched: "RT_ANY1CNT_SEC",
					utils.Weight:             10.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes[0].SortingData[utils.CapMaxUsage] = 30 * time.Second
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Time:   &tNow,
				ID:     "testV1RouteAccountWithRatingPlan",
				Event: map[string]interface{}{
					utils.Account:     "RandomAccount",
					utils.Destination: "+135876",
					utils.SetupTime:   utils.MetaNow,
					utils.Usage:       "60s",
					"EventType":       "testV1RouteSWithRateSAccountWithRateProfile",
				},
			},
		},
	}
	eSpls = &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRatePrf",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithAccAndRatePrf",
				SortingData: map[string]interface{}{
					utils.Account:            "AccWithVoice",
					utils.Cost:               0.6,
					utils.CapMaxUsage:        30000000000.0,
					utils.RateProfileMatched: "RT_ANY2CNT_SEC",
					utils.Weight:             20.0,
				},
			},
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:               0.6,
					utils.RateProfileMatched: "RT_ANY1CNT_SEC",
					utils.Weight:             10.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes[0].SortingData[utils.CapMaxUsage] = 30 * time.Second
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				Time:   &tNow,
				ID:     "testV1RouteAccountWithRatingPlan",
				Event: map[string]interface{}{
					utils.Account:     "RandomAccount",
					utils.Destination: "+135876",
					utils.SetupTime:   utils.MetaNow,
					utils.Usage:       "1m1s",
					"EventType":       "testV1RouteSWithRateSAccountWithRateProfile",
				},
			},
		},
	}
	eSpls = &engine.SortedRoutes{
		ProfileID: "RouteWithAccAndRatePrf",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithRP",
				SortingData: map[string]interface{}{
					utils.Cost:               0.61,
					utils.RateProfileMatched: "RT_ANY1CNT_SEC",
					utils.Weight:             10.0,
				},
			},
			{
				RouteID: "RouteWithAccAndRatePrf",
				SortingData: map[string]interface{}{
					utils.Account:            "AccWithVoice",
					utils.Cost:               0.62,
					utils.CapMaxUsage:        30000000000.0,
					utils.RateProfileMatched: "RT_ANY2CNT_SEC",
					utils.Weight:             20.0,
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		eSpls.SortedRoutes[1].SortingData[utils.CapMaxUsage] = 30 * time.Second
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

func testV1RouteSWithRateSWithEmptyRateProfileIDs(t *testing.T) {
	routePrf = &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "RouteWithEmptyRatePRofileIDs",
			FilterIDs: []string{"*string:~*req.EventName:testV1RouteSWithRateSWithEmptyRateProfileIDs"},
			Sorting:   utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:             "Route1",
					RateProfileIDs: []string{"RT_ANY2CNT_SEC"},
					Weight:         20,
				},
				{
					ID:             "RouteWithEmptyRP",
					RateProfileIDs: []string{}, // we send empty RateProfileIDs and expected to match RT_DEFAULT
					Weight:         10,
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testV1RouteSWithRateSGetRoutes",
				Event: map[string]interface{}{
					utils.Account:     "1003",
					utils.Subject:     "1003",
					utils.Destination: "1002",
					utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
					utils.Usage:       "1m20s",
					"EventName":       "testV1RouteSWithRateSWithEmptyRateProfileIDs",
				},
			},
		},
	}

	expRoutes := engine.SortedRoutes{
		ProfileID: "RouteWithEmptyRatePRofileIDs",
		Sorting:   utils.MetaLC,
		Count:     2,
		SortedRoutes: []*engine.SortedRoute{
			{
				RouteID: "RouteWithEmptyRP",
				SortingData: map[string]interface{}{
					utils.Cost:               0.1333333333333334,
					utils.RateProfileMatched: "RT_DEFAULT",
					utils.Weight:             10.0,
				},
			},
			{
				RouteID: "Route1",
				SortingData: map[string]interface{}{
					utils.Cost:               1.6,
					utils.RateProfileMatched: "RT_ANY2CNT_SEC",
					utils.Weight:             20.0,
				},
			},
		},
	}
	var routesReply engine.SortedRoutes
	if err := routeSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &routesReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expRoutes, routesReply) {
		t.Errorf("Expecting: %s, \n received: %s",
			utils.ToJSON(expRoutes), utils.ToJSON(routesReply))
	}

}
func testV1RouteSWithRateSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
