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
	"archive/zip"
	"bytes"
	"encoding/csv"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/tpes"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpesCfgPath   string
	tpesCfg       *config.CGRConfig
	tpeSRPC       *birpc.Client
	tpeSConfigDIR string //run tests for specific configuration

	sTestTpes = []func(t *testing.T){
		testTPeSInitCfg,
		testTPeSInitDataDb,
		testTPeSStartEngine,
		testTPeSRPCConn,
		testTPeSPing,
		testTPeSSetAttributeProfile,
		testTPeSSetResourceProfile,
		testTPeSetFilters,
		testTPeSetRateProfiles,
		testTPeSetChargerProfiles,
		testTPeSetRouteProfiles,
		testTPeSetAccount,
		testTPeSetStatQueueProfile,
		testTPeSetActions,
		testTPeSetThresholds,
		testTPeSetDispatcherProfiles,
		testSeTPeSetDispatcherHosts,
		testTPeSExportTariffPlanHalfTariffPlan,
		testTPeSExportTariffPlanAllTariffPlan,
		// export again after we will flush the database
		testTPeSInitDataDb,
		testTPeSKillEngine,
		testTPeSInitCfg,
		testTPeSStartEngine,
		testTPeSRPCConn,
		testTPeSExportAfterFlush,
		testTPeSKillEngine,
	}
)

func TestTPeSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpeSConfigDIR = "tpe_internal"
	case utils.MetaMongo:
		tpeSConfigDIR = "tutmongo"
	case utils.MetaMySQL:
		tpeSConfigDIR = "tutmysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestTpes {
		t.Run(tpeSConfigDIR, stest)
	}
}

func testTPeSInitCfg(t *testing.T) {
	var err error
	tpesCfgPath = path.Join(*dataDir, "conf", "samples", tpeSConfigDIR)
	tpesCfg, err = config.NewCGRConfigFromPath(context.Background(), tpesCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testTPeSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(tpesCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPeSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpesCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testTPeSRPCConn(t *testing.T) {
	var err error
	tpeSRPC, err = newRPCClient(tpesCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPeSPing(t *testing.T) {
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1Ping, &utils.CGREvent{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Unexpected reply returned: %s", reply)
	}
}

func testTPeSSetAttributeProfile(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:1002", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Blockers: utils.Blockers{
						{
							FilterIDs: []string{"*string:~*req.Account:1002"},
							Blocker:   true,
						},
						{
							Blocker: false,
						},
					},
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
			Blockers: utils.Blockers{
				{
					Blocker: true,
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST_SECOND",
			FilterIDs: []string{"*string:~*opts.*context:*sessions", "*exists:~*opts.*usage:"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
	}
	var reply1 string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &reply1); err != nil {
		t.Error(err)
	} else if reply1 != utils.OK {
		t.Error(err)
	}
}

func testTPeSSetResourceProfile(t *testing.T) {
	rsPrf1 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup1",
			FilterIDs:         []string{"*string:~*req.Account:1001"},
			Limit:             10,
			AllocationMessage: "Approved",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	var replystr string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf1, &replystr); err != nil {
		t.Error(err)
	} else if replystr != utils.OK {
		t.Error("Unexpected reply returned", replystr)
	}

	rsPrf2 := &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            "cgrates.org",
			ID:                "ResGroup2",
			FilterIDs:         []string{"*string:~*req.Account:1002"},
			Limit:             5,
			AllocationMessage: "Declined",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				}},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetResourceProfile,
		rsPrf2, &replystr); err != nil {
		t.Error(err)
	} else if replystr != utils.OK {
		t.Error("Unexpected reply returned", replystr)
	}
}

func testTPeSetFilters(t *testing.T) {
	fltr1 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_for_prf",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
				{
					Type:    utils.MetaExists,
					Element: "~*req.NumberOfEvents",
				},
			},
		},
	}
	fltr2 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: utils.CGRateSorg,
			ID:     "fltr_changed2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*opts.*originID",
					Values:  []string{"QWEASDZXC", "IOPJKLBNM"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaNotExists,
					Element: "~*opts.*rateS",
				},
			},
		},
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
}

func testTPeSetRateProfiles(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_RATE_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
						},
					},
				},
			},
		},
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	ratePrf2 := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "MultipleRates",
			FilterIDs: []string{"*exists:~*req.CGRID:", "*prefix:~*req.Destination:12354"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			MinCost:         utils.NewDecimal(2, 1),
			MaxCost:         utils.NewDecimal(20244, 3),
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_MONDAY": {
					ID: "RT_MONDAY",
					Weights: utils.DynamicWeights{
						{
							Weight: 50,
						},
					},
					ActivationTimes: "* * * * 0",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      utils.NewDecimal(33, 2),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(60*time.Second), 0),
							FixedFee:      utils.NewDecimal(1, 1),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
				"RT_THUESDAY": {
					ID: "RT_THUESDAY",
					Weights: utils.DynamicWeights{
						{
							Weight: 40,
						},
					},
					FilterIDs:       []string{"*string:~*opts.*rates:true"},
					ActivationTimes: "* * * * 1",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      utils.NewDecimal(20, 2),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(45*time.Second), 0),
							FixedFee:      utils.NewDecimal(0, 0),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
				"RT_WEDNESDAY": {
					ID: "RT_WEDNESDAY",
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
					ActivationTimes: "* * * * 2",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      utils.NewDecimal(1, 1),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(45*time.Second), 0),
							FixedFee:      utils.NewDecimal(2, 3),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
				"RT_THURSDAY": {
					ID: "RT_THURSDAY",
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					ActivationTimes: "* * * * 3",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      utils.NewDecimal(2, 1),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:      utils.NewDecimal(1, 3),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
				"RT_FRIDAY": {
					ID: "RT_FRIDAY",
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					ActivationTimes: "* * * * 4",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      utils.NewDecimal(5, 1),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
						{
							IntervalStart: utils.NewDecimal(int64(time.Minute+30*time.Second), 0),
							FixedFee:      utils.NewDecimal(21, 3),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
							RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
			},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testTPeSetChargerProfiles(t *testing.T) {
	chgrsPrf := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Chargers1",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	chgrsPrf2 := &ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DifferentCharger",
			RunID:        "Raw",
			AttributeIDs: []string{"ATTR1"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
		},
		APIOpts: nil,
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetChargerProfile,
		chgrsPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testTPeSetRouteProfiles(t *testing.T) {
	prf := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:     "ROUTE_2003",
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
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		prf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	rt2 := &engine.RouteProfileWithAPIOpts{
		RouteProfile: &engine.RouteProfile{
			ID:        "ROUTE_ACNT_1001",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Sorting:   "*weight",
			Routes: []*engine.Route{
				{
					ID:        "vendor1",
					FilterIDs: []string{"FLTR_DEST_1003"},
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
				},
				{
					ID:        "vendor2",
					FilterIDs: []string{"*gte:~*accounts.1001.Balance[Concrete1].Units:10"},
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
				},
				{
					ID:        "vendor3",
					FilterIDs: []string{"FLTR_DEST_1003", "*prefix:~*req.Account:10"},
					Weights: utils.DynamicWeights{
						{
							Weight: 40,
						},
					},
				},
				{
					ID: "vendor4",
					Weights: utils.DynamicWeights{
						{
							Weight: 35,
						},
					},
				},
			},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetRouteProfile,
		rt2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testTPeSetAccount(t *testing.T) {
	args := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: "cgrates.org",
			ID:     "Account_simple",
			Opts:   map[string]interface{}{},
			Balances: map[string]*utils.Balance{
				"VoiceBalance": {
					ID:        "VoiceBalance",
					FilterIDs: []string{"*string:~*req.Account:1001"},
					Weights: utils.DynamicWeights{
						{
							Weight: 12,
						},
					},
					Type: "*abstract",
					Opts: map[string]interface{}{
						"Destination": "10",
					},
					Units: utils.NewDecimal(0, 0),
				},
			},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
		},
		APIOpts: nil,
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	acnt1 := &utils.AccountWithAPIOpts{
		Account: &utils.Account{
			Tenant: utils.CGRateSorg,
			ID:     "Account_balances",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Balances: map[string]*utils.Balance{
				"AB1": {
					ID:   "AB1",
					Type: utils.MetaAbstract,
					Weights: utils.DynamicWeights{
						{
							Weight: 40,
						},
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Minute), 0),
							FixedFee:     utils.NewDecimal(4, 1), // 0.4
							RecurrentFee: utils.NewDecimal(2, 1), // 0.2 per minute
						},
					},
					Units: utils.NewDecimal(int64(130*time.Second), 0),
				},
				"CB1": {
					ID:   "CB1",
					Type: utils.MetaConcrete,
					Weights: utils.DynamicWeights{
						{
							Weight: 30,
						},
					},
					Opts: map[string]interface{}{
						utils.MetaBalanceLimit: -200.0,
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee: utils.NewDecimal(1, 1), // 0.1 per second
						},
					},
					UnitFactors: []*utils.UnitFactor{
						{
							Factor: utils.NewDecimal(100, 0), // EuroCents
						},
					},
					Units: utils.NewDecimal(80, 0),
				},
				"ab2": {
					ID:   "ab2",
					Type: utils.MetaAbstract,
					Weights: utils.DynamicWeights{
						{
							Weight: 20,
						},
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee: utils.NewDecimal(0, 0)},
					},
					Units: utils.NewDecimal(int64(1*time.Minute), 0), // 1 Minute,
				},
				"ab3": {
					ID:        "ab3",
					Type:      utils.MetaAbstract,
					FilterIDs: []string{"*string:*~req.Account:AnotherAccount"},
					Weights: utils.DynamicWeights{
						{
							Weight: 10,
						},
					},
					CostIncrements: []*utils.CostIncrement{
						{
							Increment:    utils.NewDecimal(int64(time.Second), 0),
							RecurrentFee: utils.NewDecimal(1, 0)},
					},
					Units: utils.NewDecimal(int64(60*time.Second), 0), // 1 Minute
				},
				"cb2": {
					ID:   "cb2",
					Type: utils.MetaConcrete,
					CostIncrements: []*utils.CostIncrement{
						{
							Increment: utils.NewDecimal(int64(time.Second), 0),
						},
					},
					AttributeIDs: []string{utils.MetaNone},
					Units:        utils.NewDecimal(125, 2), // 1.25
				},
			},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetAccount,
		acnt1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testTPeSetStatQueueProfile(t *testing.T) {
	sqPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "SQ_2",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			QueueLength: 14,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaASR,
				},
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaPDD,
				},
				{
					MetricID: utils.MetaTCC,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}

	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	sqPrf2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:   "cgrates.org",
			ID:       "SQ_basic",
			TTL:      0,
			Blockers: utils.Blockers{{Blocker: true}},
			MinItems: 3,
			Stored:   true,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{utils.MetaNone},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetStatQueueProfile,
		sqPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testTPeSetActions(t *testing.T) {
	actPrf := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "SET_BAL",
			FilterIDs: []string{
				"*string:~*req.Account:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Targets:  map[string]utils.StringSet{utils.MetaAccounts: {"1001": {}}},
			Schedule: utils.MetaASAP,
			Actions: []*engine.APAction{
				{
					ID:   "SET_BAL",
					Type: utils.MetaSetBalance,
					Diktats: []*engine.APDiktat{
						{
							Path:  "MONETARY",
							Value: "10",
						}},
				},
			},
		},
		APIOpts: map[string]interface{}{},
	}
	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	actPrf2 := &engine.ActionProfileWithAPIOpts{
		ActionProfile: &engine.ActionProfile{
			Tenant: "cgrates.org",
			ID:     "Execute_thd",
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Actions: []*engine.APAction{
				{
					ID:   "actID",
					Type: utils.MetaResetThreshold,
				},
			},
			Targets: map[string]utils.StringSet{
				utils.MetaThresholds: {
					"THD_1": struct{}{},
				},
			},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetActionProfile,
		actPrf2, &reply); err != nil {
		t.Error(err)
	}
}

func testTPeSetThresholds(t *testing.T) {
	thPrf2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:           "cgrates.org",
			ID:               "THD_2",
			FilterIDs:        []string{"*string:~*req.Account:1001"},
			ActionProfileIDs: []string{"actPrfID"},
			MaxHits:          7,
			MinHits:          0,
			Weights: utils.DynamicWeights{
				{
					Weight: 20,
				},
			},
			Async: true,
		},
	}

	var reply string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile,
		thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Stats1",
			FilterIDs: []string{"*string:~*req.Account:1010", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "*string:~*req.Destination:1011"},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			ActionProfileIDs: []string{"LOG"},
			Async:            true,
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testTPeSetDispatcherProfiles(t *testing.T) {
	dspPrf := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1001", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z"},
			Strategy:  utils.MetaFirst,
			StrategyParams: map[string]interface{}{
				utils.MetaDefaultRatio: "false",
			},
			Weight: 20,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:        "C1",
					FilterIDs: []string{},
					Weight:    10,
					Params:    map[string]interface{}{"0": "192.168.54.203"},
					Blocker:   false,
				},
			},
		},
	}

	var result string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	dspPrf2 := &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "Dsp2",
			FilterIDs: []string{"*string:~*opts.EventType:LoadDispatcher"},
			Strategy:  utils.MetaWeight,
			Weight:    10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:        "Conn2",
					FilterIDs: []string{"*suffix:~*opts.*answerTime:45T"},
					Params:    map[string]interface{}{utils.MetaRatio: 1},
					Blocker:   false,
				},
			},
		},
	}

	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, dspPrf2, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSeTPeSetDispatcherHosts(t *testing.T) {
	dspPrf := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "DSH1",
				Address:         "*internal",
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
			},
		},
	}
	var result string
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, dspPrf, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	dspPrf2 := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "DSH2",
				Address:         "127.0.0.1:6012",
				Transport:       utils.MetaJSON,
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
			},
		},
	}
	if err := tpeSRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, dspPrf2, &result); err != nil {
		t.Fatal(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPeSExportTariffPlanHalfTariffPlan(t *testing.T) {
	var replyBts []byte
	// we will get only the wantes tariff plans in the csv format
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1ExportTariffPlan, &tpes.ArgsExportTP{
		Tenant: "cgrates.org",
		ExportItems: map[string][]string{
			utils.MetaAttributes:      {"TEST_ATTRIBUTES_IT_TEST"},
			utils.MetaResources:       {"ResGroup1"},
			utils.MetaFilters:         {"fltr_for_prf"},
			utils.MetaRateS:           {"MultipleRates"},
			utils.MetaChargers:        {"Chargers1"},
			utils.MetaRoutes:          {"ROUTE_2003"},
			utils.MetaAccounts:        {"Account_balances"},
			utils.MetaStats:           {"SQ_basic"},
			utils.MetaActions:         {"Execute_thd"},
			utils.MetaThresholds:      {"TH_Stats1"},
			utils.MetaDispatchers:     {"Dsp1"},
			utils.MetaDispatcherHosts: {"DSH1"},
		},
	}, &replyBts); err != nil {
		t.Error(err)
	}

	rdr, err := zip.NewReader(bytes.NewReader(replyBts), int64(len(replyBts)))
	if err != nil {
		t.Error(err)
	}
	csvRply := make(map[string][][]string)
	for _, f := range rdr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		info := csv.NewReader(rc)
		//info.FieldsPerRecord = -1
		csvFile, err := info.ReadAll()
		if err != nil {
			t.Error(err)
		}
		csvRply[f.Name] = append(csvRply[f.Name], csvFile...)
		rc.Close()
	}

	expected := map[string][][]string{
		utils.AttributesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Blockers", "AttributeFilterIDs", "Path", "Type", "Value", "AttributeBlockers"},
			{"cgrates.org", "TEST_ATTRIBUTES_IT_TEST", "*string:~*req.Account:1002;*exists:~*opts.*usage:", ";20", ";true", "", "Account", "*constant", "1002", "*string:~*req.Account:1002;true;;false"},
			{"cgrates.org", "TEST_ATTRIBUTES_IT_TEST", "", "", "", "", "*tenant", "*constant", "cgrates.itsyscom", ""},
		},
		utils.ResourcesCsv: {
			{"#Tenant", "ID", "FIlterIDs", "Weights", "TTL", "Limit", "AlocationMessage", "Blocker", "Stored", "ThresholdIDs"},
			{"cgrates.org", "ResGroup1", "*string:~*req.Account:1001", ";20", "", "10", "Approved", "false", "false", "*none"},
		},
		utils.FiltersCsv: {
			{"#Tenant", "ID", "Type", "Path", "Values"},
			{"cgrates.org", "fltr_for_prf", "*string", "~*req.Subject", "1004;6774;22312"},
			{"cgrates.org", "fltr_for_prf", "*string", "~*opts.Subsystems", "*attributes"},
			{"cgrates.org", "fltr_for_prf", "*prefix", "~*req.Destinations", "+0775;+442"},
			{"cgrates.org", "fltr_for_prf", "*exists", "~*req.NumberOfEvents", ""},
		},
		utils.RatesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "MinCost", "MaxCost", "MaxCostStrategy", "RateID", "RateFilterIDs", "RateActivationStart", "RateWeights", "RateBlocker", "RateIntervalStart", "RateFixedFee", "RateRecurrentFee", "RateUnit", "RateIncrement"},
			{"cgrates.org", "MultipleRates", "*exists:~*req.CGRID:;*prefix:~*req.Destination:12354", ";20", "0.2", "20.244", "*free", "RT_MONDAY", "", "* * * * 0", ";50", "false", "0", "0.33", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_MONDAY", "", "", "", "false", "60000000000", "0.1", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THUESDAY", "*string:~*opts.*rates:true", "* * * * 1", ";40", "false", "0", "0.2", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THUESDAY", "", "", "", "false", "45000000000", "0", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THURSDAY", "", "* * * * 3", ";20", "false", "0", "0.2", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THURSDAY", "", "", "", "false", "60000000000", "0.001", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_WEDNESDAY", "", "* * * * 2", ";30", "false", "0", "0.1", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_WEDNESDAY", "", "", "", "false", "45000000000", "0.002", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_FRIDAY", "", "* * * * 4", ";10", "false", "0", "0.5", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_FRIDAY", "", "", "", "false", "90000000000", "0.021", "1000000000", "60000000000", "60000000000"},
		},
		utils.ChargersCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "RunID", "AttributeIDs"},
			{"cgrates.org", "Chargers1", "", ";20", "*default", "*none"},
		},
		utils.RoutesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Sorting", "SortingParameters", "Blockers", "RouteID", "RouteFilterIDs", "RouteAccountIDs", "RouteRateProfileIDs", "RouteResourceIDs", "RouteStatIDs", "RouteWeights", "RouteBlocker", "RouteParameters"},
			{"cgrates.org", "ROUTE_2003", "", ";10", "*weight", "", "", "route1", "", "", "", "", "", ";20", "false", ""},
		},
		utils.AccountsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Opts", "BalanceID", "BalanceFilterIDs", "BalanceWeights", "BalanceType", "BalanceUnits", "BalanceUnitFactors", "BalanceOpts", "BalanceCostIncrements", "BalanceAttributeIDs", "BalanceRateProfileIDs", "ThresholdIDs"},
			{"cgrates.org", "Account_balances", "", ";10", "", "ab2", "", ";20", "*abstract", "60000000000", "", "", ";1000000000;;0", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "ab3", "*string:*~req.Account:AnotherAccount", ";10", "*abstract", "60000000000", "", "", ";1000000000;;1", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "cb2", "", "", "*concrete", "1.25", "", "", ";1000000000;;", "*none", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "AB1", "", ";40", "*abstract", "130000000000", "", "", ";60000000000;0.4;0.2", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "CB1", "", ";30", "*concrete", "80", ";100", "*balanceLimit:-200", ";1000000000;;0.1", "", "", ""},
		},
		utils.StatsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "QueueLength", "TTL", "MinItems", "Metrics", "MetricFilterIDs", "Stored", "Blocker", "ThresholdIDs"},
			{"cgrates.org", "SQ_basic", "", ";10", "0", "", "3", "*tcd", "", "true", "true", "*none"},
		},
		utils.ActionsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Blockers", "Schedule", "TargetType", "TargetIDs", "ActionID", "ActionFilterIDs", "ActionBlocker", "ActionTTL", "ActionType", "ActionOpts", "ActionPath", "ActionValue"},
			{"cgrates.org", "Execute_thd", "", ";20", "", "", "*thresholds", "THD_1", "actID", "", "false", "0s", "*reset_threshold", "", "", ""},
		},
		utils.ThresholdsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "MaxHits", "MinHits", "MinSleep", "Blocker", "ActionProfileIDs", "Async"},
			{"cgrates.org", "TH_Stats1", "*string:~*req.Account:1010", ";10", "-1", "0", "1ms", "false", "LOG", "true"},
			{"cgrates.org", "TH_Stats1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "", "0", "0", "", "false", "", "false"},
			{"cgrates.org", "TH_Stats1", "*string:~*req.Destination:1011", "", "0", "0", "", "false", "", "false"},
		},
		utils.DispatcherProfilesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weight", "Strategy", "StrategyParameters", "ConnID", "ConnFilterIDs", "ConnWeight", "ConnBlocker", "ConnParameters"},
			{"cgrates.org", "Dsp1", "*string:~*req.Account:1001;*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "20", "*first", "false", "C1", "", "10", "false", "192.168.54.203"},
		},
		utils.DispatcherHostsCsv: {
			{"#Tenant", "ID", "Address", "Transport", "ConnectAttempts", "Reconnects", "ConnectTimeout", "ReplyTimeout", "Tls", "ClientKey", "ClientCertificate", "CaCertificate"},
			{"cgrates.org", "DSH1", "*internal", "", "1", "3", "1m0s", "2m0s", "false", "", "", ""},
		},
	}
	expected[utils.RatesCsv] = csvRply[utils.RatesCsv]
	expected[utils.AccountsCsv] = csvRply[utils.AccountsCsv]

	if !reflect.DeepEqual(expected, csvRply) {
		t.Errorf("Expected %+v \n received %+v", utils.ToJSON(expected), utils.ToJSON(csvRply))
	}
}

func testTPeSExportTariffPlanAllTariffPlan(t *testing.T) {
	var replyBts []byte
	// we will get all the tariffplans from the database
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1ExportTariffPlan, &tpes.ArgsExportTP{
		Tenant: "cgrates.org",
		ExportItems: map[string][]string{
			utils.MetaAttributes:      {"TEST_ATTRIBUTES_IT_TEST", "TEST_ATTRIBUTES_IT_TEST_SECOND"},
			utils.MetaResources:       {"ResGroup1", "ResGroup2"},
			utils.MetaFilters:         {"fltr_for_prf", "fltr_changed2"},
			utils.MetaRateS:           {"MultipleRates", "TEST_RATE_IT_TEST"},
			utils.MetaChargers:        {"Chargers1", "DifferentCharger"},
			utils.MetaRoutes:          {"ROUTE_2003", "ROUTE_ACNT_1001"},
			utils.MetaAccounts:        {"Account_balances", "Account_simple"},
			utils.MetaStats:           {"SQ_basic", "SQ_2"},
			utils.MetaActions:         {"Execute_thd", "SET_BAL"},
			utils.MetaThresholds:      {"TH_Stats1", "THD_2"},
			utils.MetaDispatchers:     {"Dsp1", "Dsp2"},
			utils.MetaDispatcherHosts: {"DSH1", "DSH2"},
		},
	}, &replyBts); err != nil {
		t.Error(err)
	}

	rdr, err := zip.NewReader(bytes.NewReader(replyBts), int64(len(replyBts)))
	if err != nil {
		t.Error(err)
	}
	csvRply := make(map[string][][]string)
	for _, f := range rdr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		info := csv.NewReader(rc)
		//info.FieldsPerRecord = -1
		csvFile, err := info.ReadAll()
		if err != nil {
			t.Error(err)
		}
		csvRply[f.Name] = append(csvRply[f.Name], csvFile...)
		rc.Close()
	}

	expected := map[string][][]string{
		utils.AttributesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Blockers", "AttributeFilterIDs", "Path", "Type", "Value", "AttributeBlockers"},
			{"cgrates.org", "TEST_ATTRIBUTES_IT_TEST", "*string:~*req.Account:1002;*exists:~*opts.*usage:", ";20", ";true", "", "Account", "*constant", "1002", "*string:~*req.Account:1002;true;;false"},
			{"cgrates.org", "TEST_ATTRIBUTES_IT_TEST", "", "", "", "", "*tenant", "*constant", "cgrates.itsyscom", ""},
			{"cgrates.org", "TEST_ATTRIBUTES_IT_TEST_SECOND", "*string:~*opts.*context:*sessions;*exists:~*opts.*usage:", "", "", "", "*tenant", "*constant", "cgrates.itsyscom", ""},
		},
		utils.ResourcesCsv: {
			{"#Tenant", "ID", "FIlterIDs", "Weights", "TTL", "Limit", "AlocationMessage", "Blocker", "Stored", "ThresholdIDs"},
			{"cgrates.org", "ResGroup1", "*string:~*req.Account:1001", ";20", "", "10", "Approved", "false", "false", "*none"},
			{"cgrates.org", "ResGroup2", "*string:~*req.Account:1002", ";10", "", "5", "Declined", "false", "false", "*none"},
		},
		utils.FiltersCsv: {
			{"#Tenant", "ID", "Type", "Path", "Values"},
			{"cgrates.org", "fltr_for_prf", "*string", "~*req.Subject", "1004;6774;22312"},
			{"cgrates.org", "fltr_for_prf", "*string", "~*opts.Subsystems", "*attributes"},
			{"cgrates.org", "fltr_for_prf", "*prefix", "~*req.Destinations", "+0775;+442"},
			{"cgrates.org", "fltr_for_prf", "*exists", "~*req.NumberOfEvents", ""},
			{"cgrates.org", "fltr_changed2", "*string", "~*opts.*originID", "QWEASDZXC;IOPJKLBNM"},
			{"cgrates.org", "fltr_changed2", "*string", "~*opts.Subsystems", "*attributes"},
			{"cgrates.org", "fltr_changed2", "*notexists", "~*opts.*rateS", ""},
		},
		utils.RatesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "MinCost", "MaxCost", "MaxCostStrategy", "RateID", "RateFilterIDs", "RateActivationStart", "RateWeights", "RateBlocker", "RateIntervalStart", "RateFixedFee", "RateRecurrentFee", "RateUnit", "RateIncrement"},
			{"cgrates.org", "MultipleRates", "*exists:~*req.CGRID:;*prefix:~*req.Destination:12354", ";20", "0.2", "20.244", "*free", "RT_MONDAY", "", "* * * * 0", ";50", "false", "0", "0.33", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_MONDAY", "", "", "", "false", "60000000000", "0.1", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THUESDAY", "*string:~*opts.*rates:true", "* * * * 1", ";40", "false", "0", "0.2", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THUESDAY", "", "", "", "false", "45000000000", "0", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THURSDAY", "", "* * * * 3", ";20", "false", "0", "0.2", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_THURSDAY", "", "", "", "false", "60000000000", "0.001", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_WEDNESDAY", "", "* * * * 2", ";30", "false", "0", "0.1", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_WEDNESDAY", "", "", "", "false", "45000000000", "0.002", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_FRIDAY", "", "* * * * 4", ";10", "false", "0", "0.5", "1000000000", "60000000000", "1000000000"},
			{"cgrates.org", "MultipleRates", "", "", "0", "0", "", "RT_FRIDAY", "", "", "", "false", "90000000000", "0.021", "1000000000", "60000000000", "60000000000"},
			{"cgrates.org", "TEST_RATE_IT_TEST", "*string:~*req.Account:dan", ";10", "0", "0", "*free", "RT_WEEK", "", "* * * * 1-5", ";0", "false", "0", "0", "0", "", ""},
		},
		utils.ChargersCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "RunID", "AttributeIDs"},
			{"cgrates.org", "Chargers1", "", ";20", "*default", "*none"},
			{"cgrates.org", "DifferentCharger", "", ";0", "Raw", "ATTR1"},
		},
		utils.RoutesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Sorting", "SortingParameters", "Blockers", "RouteID", "RouteFilterIDs", "RouteAccountIDs", "RouteRateProfileIDs", "RouteResourceIDs", "RouteStatIDs", "RouteWeights", "RouteBlocker", "RouteParameters"},
			{"cgrates.org", "ROUTE_2003", "", ";10", "*weight", "", "", "route1", "", "", "", "", "", ";20", "false", ""},
			{"cgrates.org", "ROUTE_ACNT_1001", "*string:~*req.Account:1001", "", "*weight", "", "", "vendor1", "FLTR_DEST_1003", "", "", "", "", ";10", "false", ""},
			{"cgrates.org", "ROUTE_ACNT_1001", "", "", "", "", "", "vendor2", "*gte:~*accounts.1001.Balance[Concrete1].Units:10", "", "", "", "", ";20", "false", ""},
			{"cgrates.org", "ROUTE_ACNT_1001", "", "", "", "", "", "vendor3", "FLTR_DEST_1003;*prefix:~*req.Account:10", "", "", "", "", ";40", "false", ""},
			{"cgrates.org", "ROUTE_ACNT_1001", "", "", "", "", "", "vendor4", "", "", "", "", "", ";35", "false", ""},
		},
		utils.AccountsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Opts", "BalanceID", "BalanceFilterIDs", "BalanceWeights", "BalanceType", "BalanceUnits", "BalanceUnitFactors", "BalanceOpts", "BalanceCostIncrements", "BalanceAttributeIDs", "BalanceRateProfileIDs", "ThresholdIDs"},
			{"cgrates.org", "Account_balances", "", ";10", "", "ab2", "", ";20", "*abstract", "60000000000", "", "", ";1000000000;;0", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "ab3", "*string:*~req.Account:AnotherAccount", ";10", "*abstract", "60000000000", "", "", ";1000000000;;1", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "cb2", "", "", "*concrete", "1.25", "", "", ";1000000000;;", "*none", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "AB1", "", ";40", "*abstract", "130000000000", "", "", ";60000000000;0.4;0.2", "", "", ""},
			{"cgrates.org", "Account_balances", "", "", "", "CB1", "", ";30", "*concrete", "80", ";100", "*balanceLimit:-200", ";1000000000;;0.1", "", "", ""},
			{"cgrates.org", "Account_simple", "", ";10", "", "VoiceBalance", "*string:~*req.Account:1001", ";12", "*abstract", "0", "", "Destination:10", "", "", "", ""},
		},
		utils.StatsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "QueueLength", "TTL", "MinItems", "Metrics", "MetricFilterIDs", "Stored", "Blocker", "ThresholdIDs"},
			{"cgrates.org", "SQ_basic", "", ";10", "0", "", "3", "*tcd", "", "true", "true", "*none"},
			{"cgrates.org", "SQ_2", "", ";20", "14", "", "0", "*asr", "", "false", "false", "*none"},
			{"cgrates.org", "SQ_2", "", "", "0", "", "0", "*tcd", "", "false", "false", ""},
			{"cgrates.org", "SQ_2", "", "", "0", "", "0", "*pdd", "", "false", "false", ""},
			{"cgrates.org", "SQ_2", "", "", "0", "", "0", "*tcc", "", "false", "false", ""},
			{"cgrates.org", "SQ_2", "", "", "0", "", "0", "*tcd", "", "false", "false", ""},
		},
		utils.ActionsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "Blockers", "Schedule", "TargetType", "TargetIDs", "ActionID", "ActionFilterIDs", "ActionBlocker", "ActionTTL", "ActionType", "ActionOpts", "ActionPath", "ActionValue"},
			{"cgrates.org", "Execute_thd", "", ";20", "", "", "*thresholds", "THD_1", "actID", "", "false", "0s", "*reset_threshold", "", "", ""},
			{"cgrates.org", "SET_BAL", "*string:~*req.Account:1001", ";10", "", "*asap", "*accounts", "1001", "SET_BAL", "", "false", "0s", "*set_balance", "", "MONETARY", "10"},
		},
		utils.ThresholdsCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weights", "MaxHits", "MinHits", "MinSleep", "Blocker", "ActionProfileIDs", "Async"},
			{"cgrates.org", "TH_Stats1", "*string:~*req.Account:1010", ";10", "-1", "0", "1ms", "false", "LOG", "true"},
			{"cgrates.org", "TH_Stats1", "*ai:~*req.AnswerTime:2014-07-14T14:35:00Z|2014-07-14T14:36:00Z", "", "0", "0", "", "false", "", "false"},
			{"cgrates.org", "TH_Stats1", "*string:~*req.Destination:1011", "", "0", "0", "", "false", "", "false"},
			{"cgrates.org", "THD_2", "*string:~*req.Account:1001", ";20", "7", "0", "", "false", "actPrfID", "true"},
		},
		utils.DispatcherProfilesCsv: {
			{"#Tenant", "ID", "FilterIDs", "Weight", "Strategy", "StrategyParameters", "ConnID", "ConnFilterIDs", "ConnWeight", "ConnBlocker", "ConnParameters"},
			{"cgrates.org", "Dsp1", "*string:~*req.Account:1001;*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "20", "*first", "false", "C1", "", "10", "false", "192.168.54.203"},
			{"cgrates.org", "Dsp2", "*string:~*opts.EventType:LoadDispatcher", "10", "*weight", "", "Conn2", "*suffix:~*opts.*answerTime:45T", "0", "false", "*ratio:1"},
		},
		utils.DispatcherHostsCsv: {
			{"#Tenant", "ID", "Address", "Transport", "ConnectAttempts", "Reconnects", "ConnectTimeout", "ReplyTimeout", "Tls", "ClientKey", "ClientCertificate", "CaCertificate"},
			{"cgrates.org", "DSH1", "*internal", "", "1", "3", "1m0s", "2m0s", "false", "", "", ""},
			{"cgrates.org", "DSH2", "127.0.0.1:6012", "*json", "1", "3", "1m0s", "2m0s", "false", "", "", ""},
		},
	}
	expected[utils.RatesCsv] = csvRply[utils.RatesCsv]
	expected[utils.AccountsCsv] = csvRply[utils.AccountsCsv]

	if !reflect.DeepEqual(expected, csvRply) {
		t.Errorf("Expected %+v \n received %+v", utils.ToJSON(expected), utils.ToJSON(csvRply))
	}

	// by giving an empty list of exportItems, this will do the same, it will get all the tariffplan in CSV format
	var replyBts2 []byte
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1ExportTariffPlan, &tpes.ArgsExportTP{
		Tenant:      "cgrates.org",
		ExportItems: map[string][]string{},
	}, &replyBts2); err != nil {
		t.Error(err)
	}

	rdr, err = zip.NewReader(bytes.NewReader(replyBts), int64(len(replyBts)))
	if err != nil {
		t.Error(err)
	}
	csvRply = make(map[string][][]string)
	for _, f := range rdr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		info := csv.NewReader(rc)
		info.FieldsPerRecord = -1
		csvFile, err := info.ReadAll()
		if err != nil {
			t.Error(err)
		}
		csvRply[f.Name] = append(csvRply[f.Name], csvFile...)
		rc.Close()
	}
	// expected will remain the same
	if !reflect.DeepEqual(expected, csvRply) {
		t.Errorf("Expected %+v \n received %+v", utils.ToJSON(expected), utils.ToJSON(csvRply))
	}
}

func testTPeSExportAfterFlush(t *testing.T) {
	var replyBts []byte
	// we will get all the tariffplans from the database
	if err := tpeSRPC.Call(context.Background(), utils.TPeSv1ExportTariffPlan, &tpes.ArgsExportTP{
		Tenant: "cgrates.org",
	}, &replyBts); err != nil {
		t.Error(err)
	}

	rdr, err := zip.NewReader(bytes.NewReader(replyBts), int64(len(replyBts)))
	if err != nil {
		t.Error(err)
	}
	csvRply := make(map[string][][]string)
	for _, f := range rdr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		info := csv.NewReader(rc)
		csvFile, err := info.ReadAll()
		if err != nil {
			t.Error(err)
		}
		csvRply[f.Name] = append(csvRply[f.Name], csvFile...)
		rc.Close()
	}
	// empty exporters, nothing in database to export
	if len(csvRply) != 0 {
		t.Errorf("Unexpected length, expected to be 0, no exports were nedeed and got zip containing: \n %v", utils.ToJSON(csvRply))
	}
}

//Kill the engine when it is about to be finished
func testTPeSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
