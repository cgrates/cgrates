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
	splSv1CfgPath string
	splSv1Cfg     *config.CGRConfig
	splSv1Rpc     *birpc.Client
	splPrf        *v1.RouteWithAPIOpts
	splSv1ConfDIR string //run tests for specific configuration

	sTestsSupplierSV1 = []func(t *testing.T){
		testV1SplSLoadConfig,
		testV1SplSInitDataDb,
		testV1SplSResetStorDb,
		testV1SplSStartEngine,
		testV1SplSRpcConn,
		testV1SplSFromFolder,
		testV1SplSSetRouteProfilesWithoutRatingPlanIDs,
		//tests for *reas sorting strategy
		testV1SplSAddNewRoutePrf,
		testV1SplSAddNewResPrf,
		testV1SplSPopulateResUsage,
		testV1SplSGetSortedRoutes,
		//tests for *reds sorting strategy
		testV1SplSAddNewRoutePrf2,
		testV1SplSGetSortedRoutes2,
		//tests for *load sorting strategy
		testV1SplSPopulateStats,
		testV1SplSGetSoredRoutesWithLoad,
		testV1SplSStopEngine,
	}
)

// Test start here
func TestRouteSV1IT(t *testing.T) {
	switch *utils.DBType {
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
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*utils.DataDir, "conf", "samples", splSv1ConfDIR)
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
	if _, err := engine.StopStartEngine(splSv1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	splSv1Rpc = engine.NewRPCClient(t, splSv1Cfg.ListenCfg())
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1SplSSetRouteProfilesWithoutRatingPlanIDs(t *testing.T) {
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &v1.RouteWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:  "cgrates.org",
			ID:      "TEST_PROFILE2",
			Sorting: utils.MetaLC,
			Routes: []*engine.Route{
				{
					ID:         "ROUTE1",
					AccountIDs: []string{"accc"},
					Weight:     20,
					Blocker:    false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.RouteProfile, reply)
	}
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1SplSGetLeastCostSuppliers",
		Event: map[string]any{
			utils.AccountField: "accc",
			utils.Subject:      "1003",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},
	}
	var suplsReply engine.SortedRoutesList
	if err := splSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &suplsReply); err == nil || err.Error() != utils.NewErrServerError(utils.ErrAccountNotFound).Error() {
		t.Error(err)
	}
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1RemoveRouteProfile, utils.TenantID{
		Tenant: splPrf.Tenant,
		ID:     splPrf.ID,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1SplSAddNewRoutePrf(t *testing.T) {
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ResourceTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a new Supplier Profile to test *reas and *reds sorting strategy
	splPrf = &v1.RouteWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "ROUTE_ResourceTest",
			Sorting:   utils.MetaReas,
			FilterIDs: []string{"*string:~*req.CustomField:ResourceTest"},
			Routes: []*engine.Route{
				//route1 will have ResourceUsage = 11
				{
					ID:          "route1",
					ResourceIDs: []string{"ResourceSupplier1", "Resource2Supplier1"},
					Weight:      20,
					Blocker:     false,
				},
				//route2 and route3 will have the same ResourceUsage = 7
				{
					ID:          "route2",
					ResourceIDs: []string{"ResourceSupplier2"},
					Weight:      20,
					Blocker:     false,
				},
				{
					ID:          "route3",
					ResourceIDs: []string{"ResourceSupplier3"},
					Weight:      35,
					Blocker:     false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ResourceTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.RouteProfile, reply)
	}
}

func testV1SplSAddNewResPrf(t *testing.T) {
	var result string
	//add ResourceSupplier1
	rPrf := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier1",
			FilterIDs: []string{"*string:~*req.Supplier:route1", "*string:~*req.ResID:ResourceSupplier1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, rPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add Resource2Supplier1
	rPrf2 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "Resource2Supplier1",
			FilterIDs: []string{"*string:~*req.Supplier:route1", "*string:~*req.ResID:Resource2Supplier1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       30,
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, rPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add ResourceSupplier2
	rPrf3 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier2",
			FilterIDs: []string{"*string:~*req.Supplier:route2", "*string:~*req.ResID:ResourceSupplier2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, rPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//add ResourceSupplier2
	rPrf4 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    "cgrates.org",
			ID:        "ResourceSupplier3",
			FilterIDs: []string{"*string:~*req.Supplier:route3", "*string:~*req.ResID:ResourceSupplier3"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			UsageTTL:     time.Minute,
			Limit:        10,
			Stored:       true,
			Weight:       20,
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, rPrf4, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1SplSPopulateResUsage(t *testing.T) {
	var reply string
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event1",
		Event: map[string]any{
			"Account":  "1002",
			"Supplier": "route1",
			"ResID":    "ResourceSupplier1",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RandomID",
			utils.OptsResourcesUnits:   4,
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg := "ResourceSupplier1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event2",
		Event: map[string]any{
			"Account":  "1002",
			"Supplier": "route1",
			"ResID":    "Resource2Supplier1",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RandomID2",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "Resource2Supplier1"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event3",
		Event: map[string]any{
			"Account":  "1002",
			"Supplier": "route2",
			"ResID":    "ResourceSupplier2",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RandomID3",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResourceSupplier2"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "Event4",
		Event: map[string]any{
			"Account":  "1002",
			"Supplier": "route3",
			"ResID":    "ResourceSupplier3",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "RandomID4",
			utils.OptsResourcesUnits:   7,
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		cgrEv, &reply); err != nil {
		t.Error(err)
	}
	eAllocationMsg = "ResourceSupplier3"
	if reply != eAllocationMsg {
		t.Errorf("Expecting: %+v, received: %+v", eAllocationMsg, reply)
	}

}

func testV1SplSGetSortedRoutes(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1SplSGetSortedRoutes",
		Event: map[string]any{
			"CustomField": "ResourceTest",
		},
	}
	expSupplierIDs := []string{"route3", "route2", "route1"}
	var suplsReply engine.SortedRoutesList
	if err := splSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply[0].Routes))
		for i, supl := range suplsReply[0].Routes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply[0].ProfileID != "ROUTE_ResourceTest" {
			t.Errorf("Expecting: ROUTE_ResourceTest, received: %s",
				suplsReply[0].ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSAddNewRoutePrf2(t *testing.T) {
	var reply *engine.RouteProfile
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ResourceDescendent"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//create a new Supplier Profile to test *reas and *reds sorting strategy
	splPrf = &v1.RouteWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			Tenant:    "cgrates.org",
			ID:        "ROUTE_ResourceDescendent",
			Sorting:   utils.MetaReds,
			FilterIDs: []string{"*string:~*req.CustomField:ResourceDescendent"},
			Routes: []*engine.Route{
				//route1 will have ResourceUsage = 11
				{
					ID:          "route1",
					ResourceIDs: []string{"ResourceSupplier1", "Resource2Supplier1"},
					Weight:      20,
					Blocker:     false,
				},
				//route2 and route3 will have the same ResourceUsage = 7
				{
					ID:          "route2",
					ResourceIDs: []string{"ResourceSupplier2"},
					Weight:      20,
					Blocker:     false,
				},
				{
					ID:          "route3",
					ResourceIDs: []string{"ResourceSupplier3"},
					Weight:      35,
					Blocker:     false,
				},
			},
			Weight: 10,
		},
	}
	var result string
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1SetRouteProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(context.Background(), utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ROUTE_ResourceDescendent"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.RouteProfile, reply)
	}
}

func testV1SplSGetSortedRoutes2(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1SplSGetSortedSuppliers2",
		Event: map[string]any{
			"CustomField": "ResourceDescendent",
		},
	}
	expSupplierIDs := []string{"route1", "route3", "route2"}
	var suplsReply engine.SortedRoutesList
	if err := splSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply[0].Routes))
		for i, supl := range suplsReply[0].Routes {
			rcvSupl[i] = supl.RouteID
		}
		if suplsReply[0].ProfileID != "ROUTE_ResourceDescendent" {
			t.Errorf("Expecting: ROUTE_ResourceDescendent, received: %s",
				suplsReply[0].ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSPopulateStats(t *testing.T) {
	// in this test we simulate some Stat Requests
	// so we can check the metrics in Suppliers for *load strategy
	var reply []string
	expected := []string{"Stat_Supplier1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier1",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_Supplier1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier1",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "LoadReq": "2",
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_Supplier1"}},
		&metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	expected = []string{"Stat_Supplier2"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier2",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_Supplier2"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event4",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier2",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_Supplier2"}},
		&metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	expected = []string{"Stat_Supplier3"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event5",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier3",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_Supplier3"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event6",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier3",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_Supplier3"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event7",
		Event: map[string]any{
			"LoadReq": 1,
			"StatID":  "Stat_Supplier3",
		},
	}
	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "LoadReq": "3",
	}

	if err := splSv1Rpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stat_Supplier3"}},
		&metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1SplSGetSoredRoutesWithLoad(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1SplSGetSoredSuppliersWithLoad",
		Event: map[string]any{
			"DistinctMatch": "LoadDistStrategy",
		},
	}

	expSuppliers := []*engine.SortedRoute{
		{
			RouteID:         "route2",
			RouteParameters: "",
			SortingData: map[string]any{
				"Load":   2.0,
				"Ratio":  7.0,
				"Weight": 20.0},
		},
		{
			RouteID:         "route3",
			RouteParameters: "",
			SortingData: map[string]any{
				"Load":   3.0,
				"Ratio":  5.0,
				"Weight": 35.0},
		},
		{
			RouteID:         "route1",
			RouteParameters: "",
			SortingData: map[string]any{
				"Load":   2.0,
				"Ratio":  2.0,
				"Weight": 10.0},
		},
	}

	var suplsReply engine.SortedRoutesList
	if err := splSv1Rpc.Call(context.Background(), utils.RouteSv1GetRoutes,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		if suplsReply[0].ProfileID != "ROUTE_LOAD_DIST" {
			t.Errorf("Expecting: ROUTE_LOAD_DIST, received: %s",
				suplsReply[0].ProfileID)
		}
		if !reflect.DeepEqual(suplsReply[0].Routes, expSuppliers) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				utils.ToJSON(expSuppliers), utils.ToJSON(suplsReply[0].Routes))
		}
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
