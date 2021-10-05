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
	"errors"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	tSessVolDiscCfgPath string
	tSessVolDiscCfgDIR  string
	tSessVolDiscCfg     *config.CGRConfig
	tSessVolDiscBiRPC   *birpc.Client

	tSessVolDiscITtests = []func(t *testing.T){
		testSessVolDiscInitCfg,
		testSessVolDiscResetDataDb,
		testSessVolDiscResetStorDb,
		testSessVolDiscStartEngine,
		testSessVolDiscApierRpcConn,
		testSessVolDiscLoadersLoad,
		testSessVolDiscAuthorizeEventSortRoutes1Min30Sec,
		testSessVolDiscAuthorizeEventSortRoutes11Min10Sec,
		testSessVolDiscAuthorizeEventSortRoutes20Min,
		testSessVolDiscProcessCDRSupplier,
		testSessVolDiscProcessCDRCustomer,
		testSessVolDiscAccountAfterDebiting,
		testSessVolDiscAuthorizeEventSortRoutes1Min30SecAfterDebiting,
		testSessVolDiscStopCgrEngine,
	}
)

func newBiRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

func TestSessVolDiscount(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tSessVolDiscCfgDIR = "session_volume_discount_internal"
	case utils.MetaMySQL:
		tSessVolDiscCfgDIR = "session_volume_discount_mysql"
	case utils.MetaMongo:
		tSessVolDiscCfgDIR = "session_volume_discount_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range tSessVolDiscITtests {
		t.Run(tSessVolDiscCfgDIR, stest)
	}
}

// Init config firs
func testSessVolDiscInitCfg(t *testing.T) {
	tSessVolDiscCfgPath = path.Join(*dataDir, "conf", "samples", tSessVolDiscCfgDIR)
	var err error
	tSessVolDiscCfg, err = config.NewCGRConfigFromPath(context.Background(), tSessVolDiscCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessVolDiscResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(tSessVolDiscCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSessVolDiscResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(tSessVolDiscCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessVolDiscStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSessVolDiscCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessVolDiscApierRpcConn(t *testing.T) {
	var err error
	tSessVolDiscBiRPC, err = newBiRPCClient(tSessVolDiscCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testSessVolDiscLoadersLoad(t *testing.T) {
	var reply string
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.LoaderSv1Load,
		&loaders.ArgsProcessFolder{
			// StopOnError: true,
			Caching: utils.StringPointer(utils.MetaReload),
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testSessVolDiscAuthorizeEventSortRoutes1Min30Sec(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{
			{
				ProfileID: "RP1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID: "ROUTE1",
						SortingData: map[string]interface{}{
							"Cost":       float64(0.9), // 90second * 0.01/1sec
							"AccountIDs": []interface{}{"ACCOUNT1"},
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "ROUTE2",
						SortingData: map[string]interface{}{
							"Cost":         float64(4.5), // 90second * 0.05/1sec
							"RatingPlanID": "RP_ROUTE2",
							"Weight":       float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]interface{}{
			utils.AccountField: "dan.bogos",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:  time.Minute + 30*time.Second,
			utils.OptsRouteS: true,
		},
	}
	// authorize the session for 1m30s
	var rplyFirst *sessions.V1AuthorizeReply
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		args, &rplyFirst); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rplyFirst) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rplyFirst))
	}
}

func testSessVolDiscAuthorizeEventSortRoutes11Min10Sec(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{
			{
				ProfileID: "RP1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID: "ROUTE1",
						SortingData: map[string]interface{}{
							"Cost":       float64(7.4), // 600second * 0.01/1sec(from first balance) + 70second * 0.02/1sec(from second balance)
							"AccountIDs": []interface{}{"ACCOUNT1"},
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "ROUTE2",
						SortingData: map[string]interface{}{
							"Cost":         float64(33.5), // 670second * 0.05/1sec
							"RatingPlanID": "RP_ROUTE2",
							"Weight":       float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]interface{}{
			utils.AccountField: "dan.bogos",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:  11*time.Minute + 10*time.Second,
			utils.OptsRouteS: true,
		},
	}
	// authorize the session for 11m10s
	var rplyFirst *sessions.V1AuthorizeReply
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		args, &rplyFirst); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rplyFirst) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rplyFirst))
	}
}

func testSessVolDiscAuthorizeEventSortRoutes20Min(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{
			{
				ProfileID: "RP1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID: "ROUTE1",
						SortingData: map[string]interface{}{
							"Cost":       float64(42), // 1200total: 600second * 0.01/1sec(from first balance) + 300second * 0.02/1sec(from second balance) + 300sec * 0.1/1sec(third balance)
							"AccountIDs": []interface{}{"ACCOUNT1"},
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "ROUTE2",
						SortingData: map[string]interface{}{
							"Cost":         float64(60), // 1200second * 0.05/1sec
							"RatingPlanID": "RP_ROUTE2",
							"Weight":       float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]interface{}{
			utils.AccountField: "dan.bogos",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:  20 * time.Minute,
			utils.OptsRouteS: true,
		},
	}
	// authorize the session for 20m
	var rplyFirst *sessions.V1AuthorizeReply
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		args, &rplyFirst); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rplyFirst) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rplyFirst))
	}
}

func testSessVolDiscProcessCDRSupplier(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItProcessCDR",
		Event: map[string]interface{}{
			utils.AccountField: "dan.bogos",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{
			// utils.OptsAttributeS: true,
			// utils.OptsChargerS: true,
			// utils.OptsAccountS: true,
			utils.StartTime: time.Date(2020, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.MetaUsage: 15 * time.Minute,
		},
	}

	var rply string
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testSessVolDiscProcessCDRCustomer(t *testing.T) {
	args := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestSSv1ItProcessCDR",
		Event: map[string]interface{}{
			utils.AccountField: "DIFFERENT_ACCOUNT1",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{
			// utils.OptsAttributeS: true,
			// utils.OptsChargerS: true,
			// utils.OptsAccountS: true,
			utils.StartTime: time.Date(2020, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.MetaUsage: 15 * time.Minute,
		},
	}

	var rply string
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testSessVolDiscAccountAfterDebiting(t *testing.T) {
	expectedAcc := utils.Account{
		Tenant:    "cgrates.org",
		ID:        "ACCOUNT1",
		FilterIDs: []string{},
		Balances: map[string]*utils.Balance{
			"ABS_BALANCE1": {
				ID: "ABS_BALANCE1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    30,
					},
				},
				Type:           "*abstract",
				Units:          &utils.Decimal{utils.SumDecimalAsBig(&utils.Decimal{utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0))}, // this should be -0
				RateProfileIDs: []string{"RP_ABS_BALANCE1"},
			},
			"ABS_BALANCE2": {
				ID: "ABS_BALANCE2",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    20,
					},
				},
				Type:           "*abstract",
				Units:          utils.NewDecimal(0, 0),
				RateProfileIDs: []string{"RP_ABS_BALANCE2"},
			},
			"CNCRT_BALANCE1": {
				ID: "CNCRT_BALANCE1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Opts: map[string]interface{}{
					utils.MetaBalanceUnlimited: "true",
				},
				Type:           "*concrete",
				Units:          utils.NewDecimal(98800, 2),
				RateProfileIDs: []string{"RP_CNCRT_BALANCE1"},
			},
		},
		ThresholdIDs: []string{},
	}
	var result utils.Account
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ACCOUNT1",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}
}

func testSessVolDiscAuthorizeEventSortRoutes1Min30SecAfterDebiting(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: engine.SortedRoutesList{
			{
				ProfileID: "RP1",
				Sorting:   "*lc",
				Routes: []*engine.SortedRoute{
					{
						RouteID: "ROUTE2",
						SortingData: map[string]interface{}{
							"Cost":         float64(4.5), // 90second * 0.05/1sec
							"RatingPlanID": "RP_ROUTE2",
							"Weight":       float64(0),
						},
					},
					{
						RouteID: "ROUTE1",
						SortingData: map[string]interface{}{
							"Cost":       float64(9), // 90second * 0.1/1sec (the last balance with remain units)
							"AccountIDs": []interface{}{"ACCOUNT1"},
							"Weight":     float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]interface{}{
			utils.AccountField: "dan.bogos",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:  time.Minute + 30*time.Second,
			utils.OptsRouteS: true,
		},
	}
	// authorize the session for 1m30s
	var rplyFirst *sessions.V1AuthorizeReply
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent,
		args, &rplyFirst); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, rplyFirst) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(rplyFirst))
	}
}

func testSessVolDiscStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
