//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
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
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/routes"
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
		testSessVolDiscFlushDBs,

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

func TestSessVolDiscount(t *testing.T) {
	switch *utils.DBType {
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
	tSessVolDiscCfgPath = path.Join(*utils.DataDir, "conf", "samples", tSessVolDiscCfgDIR)
	var err error
	tSessVolDiscCfg, err = config.NewCGRConfigFromPath(context.Background(), tSessVolDiscCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessVolDiscFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(tSessVolDiscCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(tSessVolDiscCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessVolDiscStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSessVolDiscCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessVolDiscApierRpcConn(t *testing.T) {
	tSessVolDiscBiRPC = engine.NewRPCClient(t, tSessVolDiscCfg.ListenCfg(), *utils.Encoding)
}

func testSessVolDiscLoadersLoad(t *testing.T) {
	caching := utils.MetaReload
	if tSessVolDiscCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{utils.MetaCache: caching},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testSessVolDiscAuthorizeEventSortRoutes1Min30Sec(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: routes.SortedRoutesList{
			{
				ProfileID: "LC1",
				Sorting:   "*lc",
				Routes: []*routes.SortedRoute{
					{
						RouteID: "supplier1",
						SortingData: map[string]any{
							"AccountIDs": []any{"ACNT_VOL1"},
							"Cost":       nil, // returns from accounts null concretes, so the cost will be null,
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "supplier2",
						SortingData: map[string]any{
							"Cost":              float64(1.2),
							utils.RateProfileID: "RP_SUPPLIER2",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier4",
						SortingData: map[string]any{
							"Cost":              float64(1.365),
							utils.RateProfileID: "RP_SUPPLIER4",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier3",
						SortingData: map[string]any{
							"Cost":              float64(1.425),
							utils.RateProfileID: "RP_SUPPLIER3",
							"Weight":            float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                time.Minute + 30*time.Second,
			utils.MetaRoutes:               true,
			utils.MetaProfileIgnoreFilters: true,
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
		RouteProfiles: routes.SortedRoutesList{
			{
				ProfileID: "LC1",
				Sorting:   "*lc",
				Routes: []*routes.SortedRoute{
					{
						RouteID: "supplier1",
						SortingData: map[string]any{
							"AccountIDs": []any{"ACNT_VOL1"},
							"Cost":       float64(8.521666666666668), // returns from accounts null concretes, so the cost will be null,
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "supplier2",
						SortingData: map[string]any{
							"Cost":              float64(8.933333333333332),
							utils.RateProfileID: "RP_SUPPLIER2",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier4",
						SortingData: map[string]any{
							"Cost":              float64(10.16166666666667),
							utils.RateProfileID: "RP_SUPPLIER4",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier3",
						SortingData: map[string]any{
							"Cost":              float64(10.60833333333333),
							utils.RateProfileID: "RP_SUPPLIER3",
							"Weight":            float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                11*time.Minute + 10*time.Second,
			utils.MetaRoutes:               true,
			utils.MetaProfileIgnoreFilters: true,
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
		RouteProfiles: routes.SortedRoutesList{
			{
				ProfileID: "LC1",
				Sorting:   "*lc",
				Routes: []*routes.SortedRoute{
					{
						RouteID: "supplier2",
						SortingData: map[string]any{
							utils.RateProfileID: "RP_SUPPLIER2",
							"Cost":              float64(16), // returns from accounts null concretes, so the cost will be null,
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier1",
						SortingData: map[string]any{
							"Cost":       float64(17.09),
							"AccountIDs": []any{"ACNT_VOL1"},
							"Weight":     float64(0),
						},
					},
					{
						RouteID: "supplier4",
						SortingData: map[string]any{
							"Cost":              float64(18.2),
							utils.RateProfileID: "RP_SUPPLIER4",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier3",
						SortingData: map[string]any{
							"Cost":              float64(19),
							utils.RateProfileID: "RP_SUPPLIER3",
							"Weight":            float64(0),
						},
					},
				},
			},
		},
	}
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSessVolDiscAuthorizeEvent1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                20 * time.Minute,
			utils.MetaRoutes:               true,
			utils.MetaProfileIgnoreFilters: true,
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
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Destination:  "1002",
			utils.RouteID:      "supplier1",
		},
		APIOpts: map[string]any{
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
		Event: map[string]any{
			utils.AccountField: "DifferentAccount",
			utils.Destination:  "1002",
			utils.RouteID:      "supplier1",
		},
		APIOpts: map[string]any{
			utils.StartTime: time.Date(2020, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.MetaUsage: 15 * time.Minute,
			utils.MetaStore: false,
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
		Tenant: "cgrates.org",
		ID:     "ACNT_VOL1",
		Opts:   make(map[string]any),
		Balances: map[string]*utils.Balance{
			"ABS_VOLUME1": {
				ID: "ABS_VOLUME1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    30,
					},
				},
				Opts:  make(map[string]any),
				Type:  "*abstract",
				Units: utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)), // this should be -0
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
						FixedFee:  utils.NewDecimal(0, 0),
					},
				},
			},
			"ABS_VOLUME2": {
				ID: "ABS_VOLUME2",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    20,
					},
				},
				Opts:  make(map[string]any),
				Type:  "*abstract",
				Units: utils.SumDecimal(&utils.Decimal{Big: utils.NewDecimal(0, 0).Neg(utils.NewDecimal(1, 0).Big)}, utils.NewDecimal(1, 0)),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP_ABS_VOLUME2"},
			},
			"CNCRT_BALANCE1": {
				ID: "CNCRT_BALANCE1",
				Weights: utils.DynamicWeights{
					{
						FilterIDs: nil,
						Weight:    10,
					},
				},
				Opts: map[string]any{
					utils.MetaBalanceUnlimited: "true",
				},
				Type:  "*concrete",
				Units: utils.NewDecimal(9882400, 4),
				CostIncrements: []*utils.CostIncrement{
					{
						Increment: utils.NewDecimal(int64(time.Second), 0),
					},
				},
				RateProfileIDs: []string{"RP_SUPPLIER1"},
			},
		},
	}
	var result utils.Account
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1GetAccount,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "ACNT_VOL1",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedAcc) {
		t.Errorf("\nExpected %+v , \nreceived %+v", utils.ToJSON(expectedAcc), utils.ToJSON(result))
	}
}

func testSessVolDiscAuthorizeEventSortRoutes1Min30SecAfterDebiting(t *testing.T) {
	expected := &sessions.V1AuthorizeReply{
		RouteProfiles: routes.SortedRoutesList{
			{
				ProfileID: "LC1",
				Sorting:   "*lc",
				Routes: []*routes.SortedRoute{
					{
						RouteID: "supplier2",
						SortingData: map[string]any{
							"Cost":              float64(1.2),
							utils.RateProfileID: "RP_SUPPLIER2",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier4",
						SortingData: map[string]any{
							"Cost":              float64(1.365),
							utils.RateProfileID: "RP_SUPPLIER4",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier3",
						SortingData: map[string]any{
							"Cost":              float64(1.425),
							utils.RateProfileID: "RP_SUPPLIER3",
							"Weight":            float64(0),
						},
					},
					{
						RouteID: "supplier1",
						SortingData: map[string]any{
							"AccountIDs": []any{"ACNT_VOL1"},
							"Cost":       float64(1.455), // returns from accounts null concretes, so the cost will be null,
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
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Category:     "call",
			utils.ToR:          "*voice",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:                time.Minute + 30*time.Second,
			utils.MetaRoutes:               true,
			utils.MetaProfileIgnoreFilters: true,
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
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
