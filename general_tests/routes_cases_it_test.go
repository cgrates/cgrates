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
	"net/rpc"
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rtsCaseSv1CfgPath string
	rtsCaseSv1Cfg     *config.CGRConfig
	rtsCaseSv1Rpc     *rpc.Client
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
		testV1RtsCaseStopEngine,
	}
)

// Test start here
func TestRoutesCaseV1IT(t *testing.T) {
	switch *dbType {
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
	rtsCaseSv1CfgPath = path.Join(*dataDir, "conf", "samples", rtsCaseSv1ConfDIR)
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
	if _, err := engine.StopStartEngine(rtsCaseSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RtsCaseRpcConn(t *testing.T) {
	var err error
	rtsCaseSv1Rpc, err = newRPCClient(rtsCaseSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RtsCaseFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutroutes")}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
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
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
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
		},
	}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
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
		SortingParameters: []string{"*acd"},
		Routes: []*engine.Route{
			{
				ID:      "vendor1",
				StatIDs: []string{"STATS_VENDOR_1"},
				Weight:  0,
			},
			{
				ID:      "vendor2",
				StatIDs: []string{"STATS_VENDOR_2"},
				Weight:  0,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_ACNT_1003", Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else {
		sort.Slice(reply.Routes, func(i, j int) bool {
			return reply.Routes[i].ID < reply.Routes[j].ID
		})
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
		},
	}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
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
		SortingParameters: []string{"vendor1:3", "vendor2:1", "*default:2"},
		Routes: []*engine.Route{
			{
				ID:      "vendor1",
				StatIDs: []string{"STATS_VENDOR_1:*sum#1"},
			},
			{
				ID:      "vendor2",
				StatIDs: []string{"STATS_VENDOR_2:*sum#1"},
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
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
		ID:                "ROUTE_STATS1",
		Tenant:            "cgrates.org",
		FilterIDs:         []string{"Fltr_tcc"},
		Sorting:           "*qos",
		SortingParameters: []string{"*tcc"},
		Routes: []*engine.Route{
			{
				ID:        "route1",
				FilterIDs: []string{"*string:~*req.Account:1001"},
				StatIDs:   []string{"STATS_TCC1"},
				Weight:    10,
			},
			{
				ID:            "route2",
				FilterIDs:     []string{"*string:~*req.Account:1002"},
				RatingPlanIDs: []string{"RP_VENDOR1"},
				StatIDs:       []string{"STATS_TCC1"},
				Weight:        10,
			},
			{
				ID:            "route3",
				RatingPlanIDs: []string{"RP_VENDOR1"},
				StatIDs:       []string{"STATS_TCC2"},
				Weight:        20,
			},
		},
	}
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{ID: "ROUTE_STATS1", Tenant: "cgrates.org"},
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
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			ID:     "WEIGHT_SORT",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
			},
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor4",
					SortingData: map[string]interface{}{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]interface{}{
						utils.Weight: 20.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCasesSortingRoutesWeightAllRoutes(t *testing.T) {
	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			ID:     "WEIGHT_SORT",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  "1003",
			},
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]interface{}{
						utils.Weight: 40.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]interface{}{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor2",
					SortingData: map[string]interface{}{
						utils.Weight: 20.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]interface{}{
						utils.Weight: 10.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(utils.RouteSv1GetRoutes,
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
		Balance: map[string]interface{}{
			utils.ID: utils.MetaDefault,
		},
	}
	var result string
	if err := rtsCaseSv1Rpc.Call(utils.APIerSv1SetBalance, attrBal,
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected result returned")
	}

	ev := &engine.ArgsGetRoutes{
		CGREvent: &utils.CGREvent{
			ID:     "WEIGHT_SORT",
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.Destination:  "1003",
			},
		},
	}
	expSrtdRoutes := &engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_ACNT_1001",
			Sorting:   "*weight",
			Routes: []*engine.SortedRoute{
				{
					RouteID: "vendor3",
					SortingData: map[string]interface{}{
						utils.Weight: 40.,
					},
				},
				{
					RouteID: "vendor4",
					SortingData: map[string]interface{}{
						utils.Weight: 35.,
					},
				},
				{
					RouteID: "vendor1",
					SortingData: map[string]interface{}{
						utils.Weight: 10.,
					},
				},
			},
		},
	}
	var reply *engine.SortedRoutesList
	if err := rtsCaseSv1Rpc.Call(utils.RouteSv1GetRoutes,
		ev, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expSrtdRoutes, reply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expSrtdRoutes), utils.ToJSON(reply))
	}
}

func testV1RtsCaseStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
