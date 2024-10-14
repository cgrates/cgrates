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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesRoutesCfgDir  string
	sesRoutesCfgPath string
	sesRoutesCfg     *config.CGRConfig
	sesRoutesRPC     *birpc.Client

	sesRoutesTests = []func(t *testing.T){
		testSesRoutesItLoadConfig,
		testSesRoutesItResetDataDB,
		testSesRoutesItResetStorDb,
		testSesRoutesItStartEngine,
		testSesRoutesItRPCConn,
		testSesRoutesItLoadFromFolder,

		testSesRoutesAuthorizeEvent,
		testSesRoutesProcessMessage,
		testSesRoutesProcessEvent,

		testSesRoutesItStopCgrEngine,
	}
)

func TestSesRoutesItSessions(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesRoutesCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesRoutesCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesRoutesCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesRoutesTests {
		t.Run(sesRoutesCfgDir, stest)
	}
}

func testSesRoutesItLoadConfig(t *testing.T) {
	var err error
	sesRoutesCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesRoutesCfgDir)
	if sesRoutesCfg, err = config.NewCGRConfigFromPath(sesRoutesCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesRoutesItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesRoutesCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRoutesItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesRoutesCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRoutesItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesRoutesCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesRoutesItRPCConn(t *testing.T) {
	sesRoutesRPC = engine.NewRPCClient(t, sesRoutesCfg.ListenCfg())
}

func testSesRoutesItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := sesRoutesRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSesRoutesAuthorizeEvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Source:       "testV4CDRsProcessCDR",
			utils.OriginID:     "testV4CDRsProcessCDR",
			utils.OriginHost:   "192.168.1.1",
			utils.RequestType:  utils.MetaPostpaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1003",
			utils.Subject:      "1003",
			utils.Destination:  "1002",
			utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:    time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:        time.Minute,
		},
		APIOpts: map[string]any{utils.OptsRoutesProfileCount: 1},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "")

	var rply sessions.V1AuthorizeReply
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{{
			ProfileID: "ROUTE_LEASTCOST_1",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					RouteID:         "route1",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				}, {
					RouteID:         "route2",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         1.2,
						"RatingPlanID": "RP_RETAIL1",
						"Weight":       20.,
					},
				},
			},
		},
		}}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "2")

	rply = sessions.V1AuthorizeReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{{
			ProfileID: "ROUTE_LEASTCOST_1",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					RouteID:         "route1",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				},
			},
		},
		}}
	// now we will set the maxCOst to be 1 in order to match route3 and route1
	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "1")

	rply = sessions.V1AuthorizeReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, true, cgrEv, utils.Paginator{}, false, "")

	rply = sessions.V1AuthorizeReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesRoutesProcessMessage(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Source:       "testV4CDRsProcessCDR",
			utils.OriginID:     "testV4CDRsProcessCDR",
			utils.OriginHost:   "192.168.1.1",
			utils.RequestType:  utils.MetaPostpaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1003",
			utils.Subject:      "1003",
			utils.Destination:  "1002",
			utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:    time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:        time.Minute,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfileCount: 1,
		},
	}
	args := sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "")

	var rply sessions.V1ProcessMessageReply
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1ProcessMessageReply{
		RouteProfiles: engine.SortedRoutesList{{
			ProfileID: "ROUTE_LEASTCOST_1",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					RouteID:         "route1",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				}, {
					RouteID:         "route2",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         1.2,
						"RatingPlanID": "RP_RETAIL1",
						"Weight":       20.,
					},
				},
			},
		},
		}}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "2")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1ProcessMessageReply{
		RouteProfiles: engine.SortedRoutesList{{
			ProfileID: "ROUTE_LEASTCOST_1",
			Sorting:   "*lc",
			Routes: []*engine.SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       15.,
					},
				}, {
					RouteID:         "route1",
					RouteParameters: "",
					SortingData: map[string]any{
						"Cost":         0.0102,
						"RatingPlanID": "RP_SPECIAL_1002",
						"Weight":       10.,
					},
				},
			},
		},
		}}
	// now we will set the maxCOst to be 1 in order to match route3 and route1
	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, false, cgrEv, utils.Paginator{}, false, "1")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.NewV1ProcessMessageArgs(false, []string{},
		false, []string{}, false, []string{}, false, false,
		true, false, true, cgrEv, utils.Paginator{}, false, "")

	rply = sessions.V1ProcessMessageReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesRoutesProcessEvent(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Source:       "testV4CDRsProcessCDR",
			utils.OriginID:     "testV4CDRsProcessCDR",
			utils.OriginHost:   "192.168.1.1",
			utils.RequestType:  utils.MetaPostpaid,
			utils.Category:     utils.Call,
			utils.AccountField: "1003",
			utils.Subject:      "1003",
			utils.Destination:  "1002",
			utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			utils.SetupTime:    time.Date(2018, 8, 24, 16, 00, 00, 0, time.UTC),
			utils.Usage:        time.Minute,
		},
		APIOpts: map[string]any{
			utils.OptsRoutesProfileCount: 1,
		},
	}
	args := sessions.V1ProcessEventArgs{
		Flags:     []string{"*routes"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	var rply sessions.V1ProcessEventReply
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessEvent, args, &rply); err != nil {
		t.Fatal(err)
	}
	expected := sessions.V1ProcessEventReply{
		RouteProfiles: map[string]engine.SortedRoutesList{
			utils.MetaRaw: {{
				ProfileID: "ROUTE_LEASTCOST_1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID:         "route3",
						RouteParameters: "",
						SortingData: map[string]any{
							"Cost":         0.0102,
							"RatingPlanID": "RP_SPECIAL_1002",
							"Weight":       15.,
						},
					}, {
						RouteID:         "route1",
						RouteParameters: "",
						SortingData: map[string]any{
							"Cost":         0.0102,
							"RatingPlanID": "RP_SPECIAL_1002",
							"Weight":       10.,
						},
					}, {
						RouteID:         "route2",
						RouteParameters: "",
						SortingData: map[string]any{
							"Cost":         1.2,
							"RatingPlanID": "RP_RETAIL1",
							"Weight":       20.,
						},
					},
				},
			},
			},
		}}
	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*routes"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	rply = sessions.V1ProcessEventReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	expected = sessions.V1ProcessEventReply{
		RouteProfiles: map[string]engine.SortedRoutesList{
			utils.MetaRaw: {{
				ProfileID: "ROUTE_LEASTCOST_1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID:         "route3",
						RouteParameters: "",
						SortingData: map[string]any{
							"Cost":         0.0102,
							"RatingPlanID": "RP_SPECIAL_1002",
							"Weight":       15.,
						},
					}, {
						RouteID:         "route1",
						RouteParameters: "",
						SortingData: map[string]any{
							"Cost":         0.0102,
							"RatingPlanID": "RP_SPECIAL_1002",
							"Weight":       10.,
						},
					},
				},
			},
			},
		}}
	// now we will set the routes max cost to be 1 in case of matching just route3 and route 1
	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*routes:*maxcost:1"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}
	args.CGREvent.APIOpts[utils.OptsRoutesMaxCost] = "1"
	rply = sessions.V1ProcessEventReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}

	args = sessions.V1ProcessEventArgs{
		Flags:     []string{"*routes:*event_cost"},
		CGREvent:  cgrEv,
		Paginator: utils.Paginator{},
	}

	rply = sessions.V1ProcessEventReply{}
	if err := sesRoutesRPC.Call(context.Background(), utils.SessionSv1ProcessEvent,
		args, &rply); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(rply, expected) {
		t.Errorf("Expected: %s, received: %s", utils.ToJSON(expected), utils.ToJSON(rply))
	}
}

func testSesRoutesItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
