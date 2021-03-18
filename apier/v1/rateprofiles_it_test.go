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

package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	ratePrfCfgPath string
	ratePrfCfg     *config.CGRConfig
	ratePrfRpc     *rpc.Client
	ratePrfConfDIR string //run tests for specific configuration

	sTestsRatePrf = []func(t *testing.T){
		testV1RatePrfLoadConfig,
		testV1RatePrfInitDataDb,
		testV1RatePrfResetStorDb,
		testV1RatePrfStartEngine,
		testV1RatePrfRpcConn,
		testV1RatePrfNotFound,
		testV1RatePrfFromFolder,
		testV1RatePrfGetRateProfileIDs,
		testV1RatePrfGetRateProfileIDsCount,
		testV1RatePrfVerifyRateProfile,
		testV1RatePrfRemoveRateProfile,
		testV1RatePrfNotFound,
		testV1RatePrfSetRateProfileRates,
		testV1RatePrfRemoveRateProfileRates,
		testV1RatePing,
		testV1RateGetRemoveRateProfileWithoutTenant,
		testV1RatePrfRemoveRateProfileWithoutTenant,
		testV1RatePrfGetRateProfileRatesWithoutTenant,
		testV1RatePrfRemoveRateProfileRatesWithoutTenant,
		testV1RateCostForEventWithDefault,
		testV1RateCostForEventWithUsage,
		testV1RateCostForEventWithWrongUsage,
		testV1RateCostForEventWithStartTime,
		testV1RateCostForEventWithWrongStartTime,
		testV1RateCostForEventWithOpts,
		//testV1RateCostForEventSpecial,
		//testV1RateCostForEventThreeRates,
		testV1RatePrfStopEngine,
	}
)

//Test start here
func TestRatePrfIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		ratePrfConfDIR = "tutinternal"
	case utils.MetaMySQL:
		ratePrfConfDIR = "tutmysql"
	case utils.MetaMongo:
		ratePrfConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRatePrf {
		t.Run(ratePrfConfDIR, stest)
	}
}

func testV1RatePrfLoadConfig(t *testing.T) {
	var err error
	ratePrfCfgPath = path.Join(*dataDir, "conf", "samples", ratePrfConfDIR)
	if ratePrfCfg, err = config.NewCGRConfigFromPath(ratePrfCfgPath); err != nil {
		t.Error(err)
	}
}

func testV1RatePrfInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(ratePrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1RatePrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ratePrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1RatePrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ratePrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1RatePrfRpcConn(t *testing.T) {
	var err error
	ratePrfRpc, err = newRPCClient(ratePrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1RatePrfNotFound(t *testing.T) {
	var reply *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RatePrfFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutrates")}
	if err := ratePrfRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1RatePrfVerifyRateProfile(t *testing.T) {
	var reply *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MinCost:         utils.NewDecimal(1, 1),
		MaxCost:         utils.NewDecimal(6, 1),
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
						RecurrentFee:  utils.NewDecimal(12, 2),
						Unit:          minDecimal,
						Increment:     minDecimal,
					},
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(6, 2),
						Unit:          minDecimal,
						Increment:     secDecimal,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(rPrf, rPrf) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(rPrf), utils.ToJSON(rPrf))
	}
}

func testV1RatePrfRemoveRateProfile(t *testing.T) {
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}
}

func testV1RatePrfSetRateProfileRates(t *testing.T) {
	rPrf := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*wrong:inline"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
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
	}
	if err := rPrf.Compile(); err != nil {
		t.Fatal(err)
	}
	apiRPrf := &utils.APIRateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*wrong:inline"},
		Weights:         ";0",
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
				},
			},
		},
	}
	var reply string
	expErr := "SERVER_ERROR: broken reference to filter: *wrong:inline for item with ID: cgrates.org:RP1"
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&APIRateProfileWithCache{
			APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
				APIRateProfile: apiRPrf},
		}, &reply); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	apiRPrf.FilterIDs = []string{"*string:~*req.Subject:1001"}
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&APIRateProfileWithCache{
			APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
				APIRateProfile: apiRPrf},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	apiRPrfRates := &utils.APIRateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
					{
						IntervalStart: "1m",
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weights:         ";10",
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";30",
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
				},
			},
		},
	}

	apiRPrfRates.Rates["RT_WEEK"].FilterIDs = []string{"*wrong:inline"}
	expErr = "SERVER_ERROR: broken reference to filter: *wrong:inline for rate with ID: RT_WEEK"
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates,
		&APIRateProfileWithCache{
			APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
				APIRateProfile: apiRPrfRates},
		}, &reply); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	apiRPrfRates.Rates["RT_WEEK"].FilterIDs = nil

	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates,
		&APIRateProfileWithCache{
			APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
				APIRateProfile: apiRPrfRates},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "RP1",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
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
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
					},
				},
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	var rply *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}}, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrfUpdated, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rPrfUpdated), utils.ToJSON(rply))
	}
}

func testV1RatePrfRemoveRateProfileRates(t *testing.T) {
	apiRPrf := &utils.APIRateProfile{
		Tenant:          "cgrates.org",
		ID:              "SpecialRate",
		FilterIDs:       []string{"*string:~*req.Subject:1001"},
		Weights:         ";0",
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
					{
						IntervalStart: "1m",
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weights:         ";10",
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";30",
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
					},
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&APIRateProfileWithCache{
			APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
				APIRateProfile: apiRPrf},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRatesWithOpts{
			Tenant:  "cgrates.org",
			ID:      "SpecialRate",
			RateIDs: []string{"RT_WEEKEND"},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "SpecialRate",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
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
					{
						IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
					},
				},
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
					},
				},
			},
		},
	}
	var rply *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "SpecialRate"}}, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrfUpdated, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rPrfUpdated), utils.ToJSON(rply))
	}

	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRatesWithOpts{
			Tenant: "cgrates.org",
			ID:     "SpecialRate",
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated2 := &utils.RateProfile{
		Tenant:    "cgrates.org",
		ID:        "SpecialRate",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates:           map[string]*utils.Rate{},
	}
	var rply2 *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "SpecialRate"}}, &rply2); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrfUpdated2, rply2) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rPrfUpdated2), utils.ToJSON(rply2))
	}
}

func testV1RatePing(t *testing.T) {
	var resp string
	if err := ratePrfRpc.Call(utils.RateSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1RatePrfStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testV1RateGetRemoveRateProfileWithoutTenant(t *testing.T) {
	rateProfile := &utils.RateProfile{
		ID:        "RPWithoutTenant",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
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
	}
	if *encoding == utils.MetaGOB {
		rateProfile.Rates["RT_WEEK"].FilterIDs = nil
	}
	apiRPrf := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
			APIRateProfile: &utils.APIRateProfile{
				ID:              "RPWithoutTenant",
				FilterIDs:       []string{"*string:~*req.Subject:1001"},
				Weights:         ";0",
				MaxCostStrategy: "*free",
				Rates: map[string]*utils.APIRate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						Weights:         ";0",
						ActivationTimes: "* * * * 1-5",
						IntervalRates: []*utils.APIIntervalRate{
							{
								IntervalStart: "0",
							},
						},
					},
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile, apiRPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *utils.RateProfile
	rateProfile.Tenant = "cgrates.org"
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RPWithoutTenant"}},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, rateProfile) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rateProfile), utils.ToJSON(result))
	}
}

func testV1RatePrfRemoveRateProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RPWithoutTenant"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "RPWithoutTenant"}},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1RatePrfGetRateProfileIDs(t *testing.T) {
	var result []string
	expected := []string{"RP1"}
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfileIDs,
		&utils.PaginatorWithTenant{},
		&result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, result)
	}
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfileIDs,
		&utils.PaginatorWithTenant{Tenant: "cgrates.org"},
		&result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, result)
	}
}

func testV1RatePrfGetRateProfileIDsCount(t *testing.T) {
	var reply int
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfileIDsCount,
		&utils.TenantWithAPIOpts{},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected 1, received %+v", reply)
	}
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfileIDsCount,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected 1, received %+v", reply)
	}
}

func testV1RatePrfGetRateProfileRatesWithoutTenant(t *testing.T) {
	rPrf := &utils.RateProfile{
		ID:        "SpecialRate",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
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
			},
			"RT_WEEKEND": {
				ID: "RT_WEEKEND",
				Weights: utils.DynamicWeights{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "* * * * 0,6",
			},
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: utils.DynamicWeights{
					{
						Weight: 30,
					},
				},
				ActivationTimes: "* * 24 12 *",
			},
		},
	}
	apiRPrf := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
			APIRateProfile: &utils.APIRateProfile{
				ID:              "SpecialRate",
				FilterIDs:       []string{"*string:~*req.Subject:1001"},
				Weights:         ";0",
				MaxCostStrategy: "*free",
				Rates: map[string]*utils.APIRate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						Weights:         ";0",
						ActivationTimes: "* * * * 1-5",
					},
					"RT_WEEKEND": {
						ID:              "RT_WEEKEND",
						Weights:         ";10",
						ActivationTimes: "* * * * 0,6",
					},
					"RT_CHRISTMAS": {
						ID:              "RT_CHRISTMAS",
						Weights:         ";30",
						ActivationTimes: "* * 24 12 *",
					},
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		rPrf.Rates["RT_WEEK"].FilterIDs = nil
		rPrf.Rates["RT_WEEKEND"].FilterIDs = nil
		rPrf.Rates["RT_CHRISTMAS"].FilterIDs = nil
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates, apiRPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	rPrf.Tenant = "cgrates.org"
	var rply *utils.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "SpecialRate"}},
		&rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrf, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(rPrf), utils.ToJSON(rply))
	}
}

func testV1RatePrfRemoveRateProfileRatesWithoutTenant(t *testing.T) {
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRatesWithOpts{ID: "SpecialRate"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1RateCostForEventWithDefault(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rate1 := &utils.Rate{
		ID: "RATE1",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(12, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(6, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rPrf := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &utils.APIRateProfileWithOpts{
			APIRateProfile: &utils.APIRateProfile{
				ID:        "DefaultRate",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Weights:   ";10",
				Rates: map[string]*utils.APIRate{
					"RATE1": &utils.APIRate{
						ID:              "RATE1",
						Weights:         ";0",
						ActivationTimes: "* * * * *",
						IntervalRates: []*utils.APIIntervalRate{
							{
								IntervalStart: "0",
								RecurrentFee:  utils.Float64Pointer(0.12),
								Unit:          utils.Float64Pointer(60000000000),
								Increment:     utils.Float64Pointer(60000000000),
							},
							{
								IntervalStart: "1m",
								RecurrentFee:  utils.Float64Pointer(0.06),
								Unit:          utils.Float64Pointer(60000000000),
								Increment:     utils.Float64Pointer(1000000000),
							},
						},
					},
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
		},
	}
	exp := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.12,
		RateSIntervals: []*utils.RateSInterval{{
			IntervalStart: utils.NewDecimal(0, 0),
			Increments: []*utils.RateSIncrement{{
				IncrementStart:    utils.NewDecimal(0, 0),
				Usage:             utils.NewDecimal(int64(time.Minute), 0),
				Rate:              rate1,
				IntervalRateIndex: 0,
				CompressFactor:    1,
			}},
			CompressFactor: 1,
		}},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func testV1RateCostForEventWithUsage(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "2m10s",
			},
		},
	}
	rate1 := &utils.Rate{
		ID: "RATE1",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(12, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(6, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	exp := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.19,
		RateSIntervals: []*utils.RateSInterval{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				Increments: []*utils.RateSIncrement{
					{
						IncrementStart:    utils.NewDecimal(0, 0),
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
						Usage:             utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    70,
					},
				},
				CompressFactor: 1,
			},
		},
	}

	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	argsRt2 := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "4h10m15s",
			},
		},
	}
	exp2 := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 15.075,
		RateSIntervals: []*utils.RateSInterval{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				Increments: []*utils.RateSIncrement{
					{
						IncrementStart:    utils.NewDecimal(0, 0),
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
						Usage:             utils.NewDecimal(int64(4*time.Hour+9*time.Minute+15*time.Second), 0),
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    14955,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt2, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp2, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp2), utils.ToJSON(rply))
	}
}

func testV1RateCostForEventWithWrongUsage(t *testing.T) {
	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "wrongUsage",
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: can't convert <wrongUsage> to decimal" {
		t.Errorf("Expected %+v \n, received %+v", "SERVER_ERROR: time: invalid duration \"wrongUsage\"", err)
	}
}

func testV1RateCostForEventWithStartTime(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rate1 := &utils.Rate{
		ID: "RATE1",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(12, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(6, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			},
		},
	}
	exp := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.12,
		RateSIntervals: []*utils.RateSInterval{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				Increments: []*utils.RateSIncrement{
					{
						IncrementStart:    utils.NewDecimal(0, 0),
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	argsRt2 := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC).String(),
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt2, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}
}

func testV1RateCostForEventWithWrongStartTime(t *testing.T) {
	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: "wrongTime",
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: Unsupported time format" {
		t.Errorf("Expected %+v \n, received %+v", "SERVER_ERROR: Unsupported time format", err)
	}
}

func testV1RateCostForEventWithOpts(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	var rply *utils.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.OptsRatesUsage:     "2m10s",
			},
		},
	}
	rate1 := &utils.Rate{
		ID: "RATE1",
		Weights: utils.DynamicWeights{
			{
				Weight: 0,
			},
		},
		ActivationTimes: "* * * * *",
		IntervalRates: []*utils.IntervalRate{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				RecurrentFee:  utils.NewDecimal(12, 2),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: utils.NewDecimal(int64(time.Minute), 0),
				RecurrentFee:  utils.NewDecimal(6, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	exp := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.19,
		RateSIntervals: []*utils.RateSInterval{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				Increments: []*utils.RateSIncrement{
					{
						IncrementStart:    utils.NewDecimal(0, 0),
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
						Usage:             utils.NewDecimal(int64(time.Minute+10*time.Second), 0),
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    70,
					},
				},
				CompressFactor: 1,
			},
		},
	}

	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))
	}

	argsRt2 := &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.OptsRatesUsage:     "4h10m15s",
			},
		},
	}
	exp2 := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 15.075,
		RateSIntervals: []*utils.RateSInterval{
			{
				IntervalStart: utils.NewDecimal(0, 0),
				Increments: []*utils.RateSIncrement{
					{
						IncrementStart:    utils.NewDecimal(0, 0),
						Usage:             utils.NewDecimal(int64(time.Minute), 0),
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						IncrementStart:    utils.NewDecimal(int64(time.Minute), 0),
						Usage:             utils.NewDecimal(int64(4*time.Hour+9*time.Minute+15*time.Second), 0),
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    14955,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt2, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp2, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp2), utils.ToJSON(rply))
	}
}

/*
func testV1RateCostForEventSpecial(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rtChristmas := &engine.Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
		IntervalRates: []*engine.IntervalRate{{
			IntervalStart: 0,
			RecurrentFee:  utils.NewDecimal(6, 2),
			Unit:          minDecimal,
			Increment:     secDecimal,
		}},
	}
	rPrf := &engine.RateProfileWithOpts{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:        "RateChristmas",
				FilterIDs: []string{"*string:~*req.Subject:1002"},
				Weight:    50,
				Rates: map[string]*engine.Rate{
					"RATE1":          rate1,
					"RATE_CHRISTMAS": rtChristmas,
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2020, 12, 23, 23, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "25h12m15s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1002",
				},
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "RateChristmas",
		Cost: 93.725,
		RateSIntervals: []*engine.RateSInterval{
			{
				UsageStart: 0,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        0,
						Usage:             time.Minute,
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						UsageStart:        1 * time.Minute,
						Usage:             59 * time.Minute,
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    3540,
					},
				},
				CompressFactor: 1,
			},
			{
				UsageStart: time.Hour,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        time.Hour,
						Usage:             24 * time.Hour,
						Rate:              rtChristmas,
						IntervalRateIndex: 0,
						CompressFactor:    86400,
					},
				},
				CompressFactor: 1,
			},
			{
				UsageStart: 25 * time.Hour,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        25 * time.Hour,
						Usage:             735 * time.Second,
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    735,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))

	}
}

func testV1RateCostForEventThreeRates(t *testing.T) {
	minDecimal, err := utils.NewDecimalFromUsage("1m")
	if err != nil {
		t.Error(err)
	}
	secDecimal, err := utils.NewDecimalFromUsage("1s")
	if err != nil {
		t.Error(err)
	}
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(2, 1),
				Unit:          minDecimal,
				Increment:     minDecimal,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  utils.NewDecimal(1, 1),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}

	rtNewYear1 := &engine.Rate{
		ID:              "NEW_YEAR1",
		ActivationTimes: "* 12-23 31 12 *",
		Weight:          20,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(8, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rtNewYear2 := &engine.Rate{
		ID:              "NEW_YEAR2",
		ActivationTimes: "* 0-12 1 1 *",
		Weight:          30,
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  utils.NewDecimal(5, 2),
				Unit:          minDecimal,
				Increment:     secDecimal,
			},
		},
	}
	rPrf := &engine.RateProfileWithOpts{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:        "RateNewYear",
				FilterIDs: []string{"*string:~*req.Subject:1003"},
				Weight:    50,
				Rates: map[string]*engine.Rate{
					"RATE1":     rate1,
					"NEW_YEAR1": rtNewYear1,
					"NEW_YEAR2": rtNewYear2,
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2020, 12, 31, 10, 0, 0, 0, time.UTC),
				utils.OptsRatesUsage:     "35h12m15s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1003",
				},
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "RateNewYear",
		Cost: 157.925,
		RateSIntervals: []*engine.RateSInterval{
			{
				UsageStart: 0,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        0,
						Usage:             time.Minute,
						Rate:              rate1,
						IntervalRateIndex: 0,
						CompressFactor:    1,
					},
					{
						UsageStart:        1 * time.Minute,
						Usage:             119 * time.Minute,
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    7140,
					},
				},
				CompressFactor: 1,
			},
			{
				UsageStart: 2 * time.Hour,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        2 * time.Hour,
						Usage:             12 * time.Hour,
						Rate:              rtNewYear1,
						IntervalRateIndex: 0,
						CompressFactor:    43200,
					},
				},
				CompressFactor: 1,
			},
			{
				UsageStart: 14 * time.Hour,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        14 * time.Hour,
						Usage:             46800 * time.Second,
						Rate:              rtNewYear2,
						IntervalRateIndex: 0,
						CompressFactor:    46800,
					},
				},
				CompressFactor: 1,
			},
			{
				UsageStart: 27 * time.Hour,
				Increments: []*engine.RateSIncrement{
					{
						UsageStart:        27 * time.Hour,
						Usage:             29535 * time.Second,
						Rate:              rate1,
						IntervalRateIndex: 1,
						CompressFactor:    29535,
					},
				},
				CompressFactor: 1,
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(exp, rply) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rply))

	}
}
*/
