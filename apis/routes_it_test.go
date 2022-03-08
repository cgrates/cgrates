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
	"path"
	"reflect"
	"sort"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	roCfgPath   string
	roCfg       *config.CGRConfig
	roRPC       *birpc.Client
	roConfigDIR string //run tests for specific configuration

	sTestsRo = []func(t *testing.T){
		testRouteSInitCfg,
		testRouteSInitDataDB,

		testRouteSStartEngine,
		testRouteSRPCConn,
		testRouteSGetRouteProfileBeforeSet,
		testRouteSGetRouteProfilesBeforeSet,
		testRouteSSetRoute,
		testRouteSSetRoute2,
		testRouteSSetRoute3,
		testFilterSGetRoutes,
		testFilterSGetRoutesWithPrefix,
		testRouteSKillEngine,
	}
)

func TestRouteSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMongo:
		roConfigDIR = "routes_mongo"
	case utils.MetaMySQL:
		roConfigDIR = "routes_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRo {
		t.Run(roConfigDIR, stest)
	}
}

func testRouteSInitCfg(t *testing.T) {
	var err error
	roCfgPath = path.Join(*dataDir, "conf", "samples", roConfigDIR)
	roCfg, err = config.NewCGRConfigFromPath(context.Background(), roCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testRouteSInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(roCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRouteSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(roCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRouteSRPCConn(t *testing.T) {
	var err error
	roRPC, err = newRPCClient(roCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRouteSGetRouteProfileBeforeSet(t *testing.T) {
	var reply *engine.Filter
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST__IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testRouteSGetRouteProfilesBeforeSet(t *testing.T) {
	var reply []*engine.RouteProfile
	args := &utils.ArgsItemIDs{}
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v \n, received %+v", utils.ErrNotFound, err)
	}
}

func testRouteSSetRoute(t *testing.T) {
	Prf := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:     "ROUTE_ACNT_1001",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}
	var reply string
	if err := roRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		Prf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expected := &engine.RouteProfile{
		ID:     "ROUTE_ACNT_1001",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	var result *engine.RouteProfile
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "ROUTE_ACNT_1001",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func testRouteSSetRoute2(t *testing.T) {
	Prf := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:     "PrefixROUTE_ACNT_1002",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}
	var reply string
	if err := roRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		Prf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expected := &engine.RouteProfile{
		ID:     "PrefixROUTE_ACNT_1002",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}
	var result *engine.RouteProfile
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "PrefixROUTE_ACNT_1002",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func testRouteSSetRoute3(t *testing.T) {
	Prf := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:     "PrefixROUTE_ACNT_1003",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}
	var reply string
	if err := roRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		Prf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expected := engine.RouteProfile{
		ID:     "PrefixROUTE_ACNT_1003",
		Tenant: "cgrates.org",
		Weights: utils.DynamicWeights{
			{
				Weight: 10,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "route1",
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
			},
		},
	}

	var result engine.RouteProfile
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantID{
			Tenant: utils.CGRateSorg,
			ID:     "PrefixROUTE_ACNT_1003",
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(result))
	}
}

func testFilterSGetRoutes(t *testing.T) {
	var reply []*engine.RouteProfile
	args := &utils.ArgsItemIDs{}
	expected := []*engine.RouteProfile{

		{
			ID:     "PrefixROUTE_ACNT_1002",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
		{
			ID:     "PrefixROUTE_ACNT_1003",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
		{
			ID:     "ROUTE_ACNT_1001",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(reply, func(i, j int) bool {
		return (reply)[i].ID < (reply)[j].ID
	})
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}
func testFilterSGetRoutesWithPrefix(t *testing.T) {
	var reply []*engine.RouteProfile
	args := &utils.ArgsItemIDs{
		ItemsPrefix: "PrefixROUTE",
	}
	expected := []*engine.RouteProfile{
		{
			ID:     "PrefixROUTE_ACNT_1002",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
		{
			ID:     "PrefixROUTE_ACNT_1003",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
			},
		},
	}
	if err := roRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(reply, func(i, j int) bool {
		return (reply)[i].ID < (reply)[j].ID
	})
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

//Kill the engine when it is about to be finished
func testRouteSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
