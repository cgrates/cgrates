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
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"

	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
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
		testSessVolDiscCheckRoutesAndRateProfiles,
		testSessVolDiscSetAccounts,
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

func testSessVolDiscCheckRoutesAndRateProfiles(t *testing.T) {
	expRp := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{},
		MinCost:   utils.NewDecimal(0, 0),
		MaxCost:   utils.NewDecimal(0, 0),
		Rates: map[string]*utils.Rate{
			"RT_10": {
				ID:        "RT_10",
				FilterIDs: []string{"*prefix:~*req.Destination:10"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(3, 2),
						Unit:          utils.NewDecimal(60000000000, 0),
						Increment:     utils.NewDecimal(1000000000, 0),
					},
				},
			},
			"RT_20": {
				ID:        "RT_20",
				FilterIDs: []string{"*prefix:~*req.Destination:20"},
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(1, 2),
						Unit:          utils.NewDecimal(60000000000, 0),
						Increment:     utils.NewDecimal(1000000000, 0),
					},
				},
			},
			"RT_DFLT": {
				ID: "RT_DFLT",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(92, 2),
						Unit:          utils.NewDecimal(60000000000, 0),
						Increment:     utils.NewDecimal(1000000000, 0),
					},
				},
			},
		},
	}
	var reply *utils.RateProfile
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "RP1",
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expRp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRp), utils.ToJSON(reply))
	}

	expRoutePrf := &engine.APIRouteProfile{
		Tenant:            "cgrates.org",
		ID:                "LC1",
		FilterIDs:         []string{},
		Sorting:           "*lc",
		SortingParameters: []string{},
		Routes: []*engine.ExternalRoute{
			{
				ID:             "supplier1",
				RateProfileIDs: []string{"RP1"},
			},
			{
				ID:             "supplier2",
				RateProfileIDs: []string{"RP2"},
			},
			{
				ID:             "supplier3",
				RateProfileIDs: []string{"RP3"},
			},
			{
				ID:             "supplier4",
				RateProfileIDs: []string{"RP4"},
			},
		},
	}
	var result *engine.APIRouteProfile
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1GetRouteProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: "cgrates.org",
				ID:     "LC1",
			},
		}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Slice(result.Routes, func(i, j int) bool {
			return result.Routes[i].ID < result.Routes[j].ID
		})
		if !reflect.DeepEqual(result, expRoutePrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRoutePrf), utils.ToJSON(result))
		}
	}
}

func testSessVolDiscSetAccounts(t *testing.T) {
	accPrf1 := &apis.APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACC1",
			FilterIDs: []string{"*string:~*req.Account:1"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance1": {
					ID:    "AbstractBalance1",
					Type:  utils.MetaAbstract,
					Units: float64(40 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	accPrf2 := &apis.APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACC2",
			FilterIDs: []string{"*string:~*req.Account:2"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance2": {
					ID:    "AbstractBalance2",
					Type:  utils.MetaAbstract,
					Units: float64(80 * time.Second),
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Minute)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	accPrf3 := &apis.APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACC3",
			FilterIDs: []string{"*string:~*req.Account:3"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance3": {
					ID:    "AbstractBalance3",
					Type:  utils.MetaAbstract,
					Units: float64(120 * time.Second),
					UnitFactors: []*utils.APIUnitFactor{
						{
							Factor: 100,
						},
					},
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Minute)),
							FixedFee:     utils.Float64Pointer(float64(0)),
							RecurrentFee: utils.Float64Pointer(float64(1)),
						},
					},
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	accPrf4 := &apis.APIAccountWithAPIOpts{
		APIAccount: &utils.APIAccount{
			Tenant:    "cgrates.org",
			ID:        "ACC3",
			FilterIDs: []string{"*string:~*req.Account:4"},
			Balances: map[string]*utils.APIBalance{
				"AbstractBalance4": {
					ID:    "AbstractBalance4",
					Type:  utils.MetaAbstract,
					Units: float64(30 * time.Second),
					UnitFactors: []*utils.APIUnitFactor{
						{
							Factor: 100,
						},
					},
					CostIncrements: []*utils.APICostIncrement{
						{
							Increment:    utils.Float64Pointer(float64(time.Second)),
							FixedFee:     utils.Float64Pointer(float64(0.01)),
							RecurrentFee: utils.Float64Pointer(float64(2)),
						},
					},
				},
			},
			Weights: ";10",
		},
		APIOpts: nil,
	}
	var reply string
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tSessVolDiscBiRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		accPrf4, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testSessVolDiscStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
