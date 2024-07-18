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
	"slices"
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
		testRouteSResetStorDB,
		testRoutesStartEngine,
		testRoutesRPCConn,

		// tests for AdminSv1 APIs
		testRoutesGetRouteProfileBeforeSet,
		testRoutesGetRouteProfileIDsBeforeSet,
		testRoutesGetRouteProfileCountBeforeSet,
		testRoutesGetRouteProfilesBeforeSet,
		testRoutesSetRouteProfiles,
		testRoutesGetRouteProfileAfterSet,
		testRoutesGetRouteProfileIDsAfterSet,
		testRoutesGetRouteProfileCountAfterSet,
		testRoutesGetRouteProfilesAfterSet,
		testRoutesRemoveRouteProfile,
		testRoutesGetRouteProfileAfterRemove,
		testRoutesGetRouteProfileIDsAfterRemove,
		testRoutesGetRouteProfileCountAfterRemove,
		testRoutesGetRouteProfilesAfterRemove,

		// RouteProfile blocker behaviour test
		testRoutesBlockerRemoveRouteProfiles,
		testRoutesBlockerSetRouteProfiles,
		testRoutesBlockerGetRouteProfilesForEvent,

		// Route blocker behaviour test
		testRoutesBlockerSetRouteProfile,
		testRoutesBlockerGetRoutes,

		testRoutesKillEngine,
	}
)

func TestRoutesIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		raConfigDIR = "routes_internal"
	case utils.MetaMongo:
		raConfigDIR = "routes_mongo"
	case utils.MetaMySQL:
		raConfigDIR = "routes_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRa {
		t.Run(raConfigDIR, stest)
	}
}

func testRouteSInitCfg(t *testing.T) {
	var err error
	raCfgPath = path.Join(*dataDir, "conf", "samples", raConfigDIR)
	raCfg, err = config.NewCGRConfigFromPath(context.Background(), raCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testRouteSInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(raCfg); err != nil {
		t.Fatal(err)
	}
}

func testRouteSResetStorDB(t *testing.T) {
	if err := engine.InitStorDB(raCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRoutesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(raCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRoutesRPCConn(t *testing.T) {
	var err error
	raRPC, err = newRPCClient(raCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRoutesGetRouteProfileBeforeSet(t *testing.T) {
	var replyRouteProfile engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ROUTE1",
			}}, &replyRouteProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRoutesGetRouteProfilesBeforeSet(t *testing.T) {
	var replyRouteProfiles *[]*engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfiles); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRoutesGetRouteProfileIDsBeforeSet(t *testing.T) {
	var replyRouteProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfileIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRoutesGetRouteProfileCountBeforeSet(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	} else if replyCount != 0 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRoutesSetRouteProfiles(t *testing.T) {
	routeProfiles := []*engine.RouteProfileWithAPIOpts{
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "TestA_ROUTE1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "TestA_ROUTE2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "TestA_ROUTE3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "TestB_ROUTE1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "TestB_ROUTE2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
				Weights: utils.DynamicWeights{
					{
						Weight: 25,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
	}

	var reply string
	for _, routeProfile := range routeProfiles {
		if err := raRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
			routeProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testRoutesGetRouteProfileAfterSet(t *testing.T) {
	expectedRouteProfile := engine.RouteProfile{
		ID:        "TestA_ROUTE1",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
		Weights: utils.DynamicWeights{
			{
				Weight: 30,
			},
		},
		Blockers: utils.DynamicBlockers{
			{
				Blocker: false,
			},
		},
		Sorting:           utils.MetaWeight,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID: "routeTest",
			},
		},
	}
	var replyRouteProfile engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ROUTE1",
			}}, &replyRouteProfile); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(replyRouteProfile, expectedRouteProfile) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>",
			utils.ToJSON(expectedRouteProfile), utils.ToJSON(replyRouteProfile))
	}
}

func testRoutesGetRouteProfileIDsAfterSet(t *testing.T) {
	expectedIDs := []string{"TestA_ROUTE1", "TestA_ROUTE2", "TestA_ROUTE3", "TestB_ROUTE1", "TestB_ROUTE2"}
	var replyRouteProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_ROUTE1", "TestA_ROUTE2", "TestA_ROUTE3"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_ROUTE1", "TestB_ROUTE2"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}
}

func testRoutesGetRouteProfileCountAfterSet(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 5 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 3 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRoutesGetRouteProfilesAfterSet(t *testing.T) {
	expectedRouteProfiles := []*engine.RouteProfile{
		{
			ID:        "TestA_ROUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestA_ROUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestA_ROUTE3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestB_ROUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestB_ROUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
	}
	var replyRouteProfiles []*engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyRouteProfiles, func(i, j int) bool {
			return replyRouteProfiles[i].ID < replyRouteProfiles[j].ID
		})
		if !reflect.DeepEqual(replyRouteProfiles, expectedRouteProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedRouteProfiles), utils.ToJSON(replyRouteProfiles))
		}
	}
}

func testRoutesRemoveRouteProfile(t *testing.T) {
	var reply string
	if err := raRPC.Call(context.Background(), utils.AdminSv1RemoveRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_ROUTE2",
			}}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testRoutesGetRouteProfileAfterRemove(t *testing.T) {
	var replyRouteProfile engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "TestA_Route2",
			}}, &replyRouteProfile); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRoutesGetRouteProfileIDsAfterRemove(t *testing.T) {
	expectedIDs := []string{"TestA_ROUTE1", "TestA_ROUTE3", "TestB_ROUTE1", "TestB_ROUTE2"}
	var replyRouteProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}

	expectedIDs = []string{"TestA_ROUTE1", "TestA_ROUTE3"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}

	expectedIDs = []string{"TestB_ROUTE1", "TestB_ROUTE2"}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyRouteProfileIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(replyRouteProfileIDs)
		if !slices.Equal(replyRouteProfileIDs, expectedIDs) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expectedIDs, replyRouteProfileIDs)
		}
	}
}

func testRoutesGetRouteProfileCountAfterRemove(t *testing.T) {
	var replyCount int
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 4 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestA",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}

	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfilesCount,
		&utils.ArgsItemIDs{
			Tenant:      "cgrates.org",
			ItemsPrefix: "TestB",
		}, &replyCount); err != nil {
		t.Error(err)
	} else if replyCount != 2 {
		t.Errorf("expected <%+v>, \nreceived: <%+v>", 0, replyCount)
	}
}

func testRoutesGetRouteProfilesAfterRemove(t *testing.T) {
	expectedRouteProfiles := []*engine.RouteProfile{
		{
			ID:        "TestA_ROUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestA_ROUTE3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestB_ROUTE1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 5,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "TestB_ROUTE2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:AdminSAPIs"},
			Weights: utils.DynamicWeights{
				{
					Weight: 25,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
	}
	var replyRouteProfiles []*engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfiles,
		&utils.ArgsItemIDs{
			Tenant: "cgrates.org",
		}, &replyRouteProfiles); err != nil {
		t.Error(err)
	} else {
		sort.Slice(replyRouteProfiles, func(i, j int) bool {
			return replyRouteProfiles[i].ID < replyRouteProfiles[j].ID
		})
		if !reflect.DeepEqual(replyRouteProfiles, expectedRouteProfiles) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>",
				utils.ToJSON(expectedRouteProfiles), utils.ToJSON(replyRouteProfiles))
		}
	}
}

func testRoutesBlockerRemoveRouteProfiles(t *testing.T) {
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TestA_ROUTE1", "TestA_ROUTE3", "TestB_ROUTE1", "TestB_ROUTE2"}
	var routeProfileIDs []string
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs, args, &routeProfileIDs); err != nil {
		t.Fatal(err)
	} else {
		sort.Strings(routeProfileIDs)
		if !slices.Equal(routeProfileIDs, expected) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", expected, routeProfileIDs)
		}
	}
	var reply string
	for _, routeProfileID := range routeProfileIDs {
		argsRem := utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     routeProfileID,
			},
		}
		if err := raRPC.Call(context.Background(), utils.AdminSv1RemoveRouteProfile, argsRem, &reply); err != nil {
			t.Fatal(err)
		}
	}
	if err := raRPC.Call(context.Background(), utils.AdminSv1GetRouteProfileIDs, args, &routeProfileIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrNotFound, err)
	}
}

func testRoutesBlockerSetRouteProfiles(t *testing.T) {
	routeProfiles := []*engine.RouteProfileWithAPIOpts{
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "ROUTE_TEST_1",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: false,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "ROUTE_TEST_2",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				Blockers: utils.DynamicBlockers{
					{
						Blocker: true,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "ROUTE_TEST_3",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 20,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
		{
			RouteProfile: &engine.RouteProfile{
				ID:        "ROUTE_TEST_4",
				Tenant:    "cgrates.org",
				FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
				Weights: utils.DynamicWeights{
					{
						Weight: 5,
					},
				},
				Sorting:           utils.MetaWeight,
				SortingParameters: []string{},
				Routes: []*engine.Route{
					{
						ID: "routeTest",
					},
				},
			},
		},
	}

	var reply string
	for _, routeProfile := range routeProfiles {
		if err := raRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
			routeProfile, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error(err)
		}
	}
}

func testRoutesBlockerGetRouteProfilesForEvent(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventGetRouteProfiles",
		Event: map[string]any{
			"TestCase": "RouteProfileBlockerBehaviour",
		},
		APIOpts: map[string]any{},
	}
	expected := []*engine.RouteProfile{
		{
			ID:        "ROUTE_TEST_1",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "ROUTE_TEST_3",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
		{
			ID:        "ROUTE_TEST_2",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:RouteProfileBlockerBehaviour"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: true,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "routeTest",
				},
			},
		},
	}
	var reply []*engine.RouteProfile
	if err := raRPC.Call(context.Background(), utils.RouteSv1GetRouteProfilesForEvent, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testRoutesBlockerSetRouteProfile(t *testing.T) {
	routeProfile := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:        "ROUTE_BLOCKER_TEST",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.TestCase:RouteBlockerBehaviour"},
			Weights: utils.DynamicWeights{
				{
					Weight: 30,
				},
			},
			Blockers: utils.DynamicBlockers{
				{
					Blocker: false,
				},
			},
			Sorting:           utils.MetaWeight,
			SortingParameters: []string{},
			Routes: []*engine.Route{
				{
					ID: "route1",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
				},
				{
					ID: "route2",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					Blockers: utils.DynamicBlockers{
						{
							Blocker: true,
						},
					},
				},
				{
					ID: "route3",
					Weights: utils.DynamicWeights{
						{
							Weight: 40,
						},
					},
				},
				{
					ID: "route4",
					Weights: utils.DynamicWeights{
						{
							Weight: 35,
						},
					},
				},
			},
		},
	}

	var reply string
	if err := raRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		routeProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testRoutesBlockerGetRoutes(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EventGetRoutes",
		Event: map[string]any{
			"TestCase": "RouteBlockerBehaviour",
		},
		APIOpts: map[string]any{},
	}
	expected := engine.SortedRoutesList{
		{
			ProfileID: "ROUTE_BLOCKER_TEST",
			Sorting:   utils.MetaWeight,
			Routes: []*engine.SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: utils.EmptyString,
					SortingData: map[string]any{
						utils.Weight: 40.,
					},
				},
				{
					RouteID:         "route4",
					RouteParameters: utils.EmptyString,
					SortingData: map[string]any{
						utils.Weight: 35.,
					},
				},
				{
					RouteID:         "route2",
					RouteParameters: utils.EmptyString,
					SortingData: map[string]any{
						utils.Weight:  20.,
						utils.Blocker: true,
					},
				},
			},
		},
	}

	var reply engine.SortedRoutesList
	if err := raRPC.Call(context.Background(), utils.RouteSv1GetRoutes, args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

// Kill the engine when it is about to be finished
func testRoutesKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
