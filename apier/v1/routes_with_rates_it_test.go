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

func testV1RouteSWithRateSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
