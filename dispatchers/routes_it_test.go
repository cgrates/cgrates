//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestsDspSup = []func(t *testing.T){
		testDspSupPingFailover,
		testDspSupGetSupFailover,
		testDspSupGetSupRoundRobin,

		testDspSupPing,
		testDspSupTestAuthKey,
		testDspSupTestAuthKey2,
		testDspSupGetSupplierForEvent,
	}
	nowTime = time.Now()
)

// Test start here
func TestDspSupplierS(t *testing.T) {
	var config1, config2, config3 string
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *utils.Encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspSup, "TestDspSupplierS", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspSupPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspSupPingFailover(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	ev := utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1Ping, &ev, &reply); err == nil {
		t.Errorf("Expected error but received %v and reply %v\n", err, reply)
	}
	allEngine.startEngine(t)
	allEngine2.startEngine(t)
}

func testDspSupGetSupFailover(t *testing.T) {
	var rpl engine.SortedRoutesList
	eRpl1 := engine.SortedRoutesList{{
		ProfileID: "ROUTE_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		Routes: []*engine.SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Weight: 10.0,
				},
			},
		},
	}}
	eRpl := engine.SortedRoutesList{{
		ProfileID: "ROUTE_ACNT_1002",
		Sorting:   utils.MetaLC,
		Routes: []*engine.SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID:         "route2",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.EventName:    "Event1",
			utils.AccountField: "1002",
			utils.Subject:      "1002",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl1, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl1), utils.ToJSON(rpl))
	}
	allEngine2.stopEngine(t)
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
	allEngine2.startEngine(t)
}

func testDspSupTestAuthKey(t *testing.T) {
	var rpl engine.SortedRoutesList
	args := &utils.CGREvent{
		ID:   utils.UUIDSha1Prefix(),
		Time: &nowTime,
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Subject:      "1002",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err == nil || err.Error() != utils.ErrUnauthorizedApi.Error() {
		t.Error(err)
	}
}

func testDspSupTestAuthKey2(t *testing.T) {
	var rpl engine.SortedRoutesList
	eRpl := engine.SortedRoutesList{{
		ProfileID: "ROUTE_ACNT_1002",
		Sorting:   utils.MetaLC,
		Routes: []*engine.SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID:         "route2",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Subject:      "1002",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
}

func testDspSupGetSupRoundRobin(t *testing.T) {
	var rpl engine.SortedRoutesList
	eRpl1 := engine.SortedRoutesList{{
		ProfileID: "ROUTE_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		Routes: []*engine.SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Weight: 10.0,
				},
			},
		},
	}}
	eRpl := engine.SortedRoutesList{{
		ProfileID: "ROUTE_ACNT_1002",
		Sorting:   utils.MetaLC,
		Routes: []*engine.SortedRoute{
			{
				RouteID:         "route1",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.3166,
					utils.RatingPlanID: "RP_1002_LOW",
					utils.Weight:       10.0,
				},
			},
			{
				RouteID:         "route2",
				RouteParameters: "",
				SortingData: map[string]any{
					utils.Cost:         0.6334,
					utils.RatingPlanID: "RP_1002",
					utils.Weight:       20.0,
				},
			},
		},
	}}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Time:   &nowTime,
		Event: map[string]any{
			utils.EventName:    "RoundRobin",
			utils.AccountField: "1002",
			utils.Subject:      "1002",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl1, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl1), utils.ToJSON(rpl))
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRoutes,
		args, &rpl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eRpl, rpl) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(eRpl), utils.ToJSON(rpl))
	}
}

func testDspSupGetSupplierForEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1SplSGetHighestCostSuppliers",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.Subject:      "1002",
			utils.Destination:  "1001",
			utils.SetupTime:    time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			utils.Usage:        "1m20s",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "sup12345",
		},
	}
	expected := engine.RouteProfile{
		Tenant:    "cgrates.org",
		ID:        "ROUTE_ACNT_1002",
		FilterIDs: []string{"FLTR_ACNT_1002"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 00, 00, 00, 00, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Routes: []*engine.Route{
			{
				ID:              "route1",
				FilterIDs:       nil,
				AccountIDs:      nil,
				RatingPlanIDs:   []string{"RP_1002_LOW"},
				ResourceIDs:     nil,
				StatIDs:         nil,
				Weight:          10,
				Blocker:         false,
				RouteParameters: "",
			},
			{
				ID:              "route2",
				FilterIDs:       nil,
				AccountIDs:      nil,
				RatingPlanIDs:   []string{"RP_1002"},
				ResourceIDs:     nil,
				StatIDs:         nil,
				Weight:          20,
				Blocker:         false,
				RouteParameters: "",
			},
		},
		Weight: 10,
	}
	if *utils.Encoding == utils.MetaGOB {
		expected.SortingParameters = nil // empty slices are nil in gob
	}
	var supProf []*engine.RouteProfile
	if err := dispEngine.RPC.Call(context.Background(), utils.RouteSv1GetRouteProfilesForEvent,
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
