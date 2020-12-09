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
		testV1RateCostForEventSpecial,
		testV1RateCostForEventThreeRates,
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
	var reply *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}},
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
	var reply *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}}, &reply); err != nil {
		t.Fatal(err)
	}
	rPrf := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "RP1"}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}
}

func testV1RatePrfSetRateProfileRates(t *testing.T) {
	rPrf := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*wrong:inline"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
	}
	if err := rPrf.Compile(); err != nil {
		t.Fatal(err)
	}
	var reply string
	expErr := "SERVER_ERROR: broken reference to filter: *wrong:inline for item with ID: cgrates.org:RP1"
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&RateProfileWithCache{
			RateProfileWithOpts: &engine.RateProfileWithOpts{
				RateProfile: rPrf},
		}, &reply); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	rPrf.FilterIDs = []string{"*string:~*req.Subject:1001"}
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&RateProfileWithCache{
			RateProfileWithOpts: &engine.RateProfileWithOpts{
				RateProfile: rPrf},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfRates := &engine.RateProfile{
		Tenant: "cgrates.org",
		ID:     "RP1",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}

	rPrfRates.Rates["RT_WEEK"].FilterIDs = []string{"*wrong:inline"}
	expErr = "SERVER_ERROR: broken reference to filter: *wrong:inline for rate with ID: RT_WEEK"
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates,
		&RateProfileWithCache{
			RateProfileWithOpts: &engine.RateProfileWithOpts{
				RateProfile: rPrfRates},
		}, &reply); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	rPrfRates.Rates["RT_WEEK"].FilterIDs = nil

	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates,
		&RateProfileWithCache{
			RateProfileWithOpts: &engine.RateProfileWithOpts{
				RateProfile: rPrfRates},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "RP1",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	var rply *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "RP1"}}, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrfUpdated, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rPrfUpdated), utils.ToJSON(rply))
	}
}

func testV1RatePrfRemoveRateProfileRates(t *testing.T) {
	rPrf := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "SpecialRate",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_WEEKEND": {
				ID:              "RT_WEEKEND",
				Weight:          10,
				ActivationTimes: "* * * * 0,6",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile,
		&RateProfileWithCache{
			RateProfileWithOpts: &engine.RateProfileWithOpts{
				RateProfile: rPrf},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRates{
			Tenant:  "cgrates.org",
			ID:      "SpecialRate",
			RateIDs: []string{"RT_WEEKEND"},
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "SpecialRate",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates: map[string]*engine.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weight:          30,
				ActivationTimes: "* * 24 12 *",
				IntervalRates: []*engine.IntervalRate{
					{
						IntervalStart: 0,
						RecurrentFee:  0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
	}
	var rply *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "SpecialRate"}}, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrfUpdated, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rPrfUpdated), utils.ToJSON(rply))
	}

	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRates{
			Tenant: "cgrates.org",
			ID:     "SpecialRate",
		}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	rPrfUpdated2 := &engine.RateProfile{
		Tenant:           "cgrates.org",
		ID:               "SpecialRate",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		RoundingMethod:   "*up",
		RoundingDecimals: 4,
		MinCost:          0.1,
		MaxCost:          0.6,
		MaxCostStrategy:  "*free",
		Rates:            map[string]*engine.Rate{},
	}
	var rply2 *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "SpecialRate"}}, &rply2); err != nil {
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
	rateProfile := &RateProfileWithCache{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:               "RPWithoutTenant",
				FilterIDs:        []string{"*string:~*req.Subject:1001"},
				Weight:           0,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MinCost:          0.1,
				MaxCost:          0.6,
				MaxCostStrategy:  "*free",
				Rates: map[string]*engine.Rate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						Weight:          0,
						ActivationTimes: "* * * * 1-5",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 0,
								RecurrentFee:  0.12,
								Unit:          time.Minute,
								Increment:     time.Minute,
							},
						},
					},
				},
			},
		},
	}
	if *encoding == utils.MetaGOB {
		rateProfile.Rates["RT_WEEK"].FilterIDs = nil
	}
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfile, rateProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.RateProfile
	rateProfile.Tenant = "cgrates.org"
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "RPWithoutTenant"}},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, rateProfile.RateProfile) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rateProfile.RateProfile), utils.ToJSON(result))
	}
}

func testV1RatePrfRemoveRateProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfile,
		&utils.TenantIDWithCache{ID: "RPWithoutTenant"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "RPWithoutTenant"}},
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
		&utils.TenantWithOpts{},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected 1, received %+v", reply)
	}
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfileIDsCount,
		&utils.TenantWithOpts{Tenant: "cgrates.org"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected 1, received %+v", reply)
	}
}

func testV1RatePrfGetRateProfileRatesWithoutTenant(t *testing.T) {
	rPrf := &RateProfileWithCache{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:               "SpecialRate",
				FilterIDs:        []string{"*string:~*req.Subject:1001"},
				Weight:           0,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MinCost:          0.1,
				MaxCost:          0.6,
				MaxCostStrategy:  "*free",
				Rates: map[string]*engine.Rate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						Weight:          0,
						ActivationTimes: "* * * * 1-5",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 0,
								RecurrentFee:  0.12,
								Unit:          time.Minute,
								Increment:     time.Minute,
							},
							{
								IntervalStart: time.Minute,
								RecurrentFee:  0.06,
								Unit:          time.Minute,
								Increment:     time.Second,
							},
						},
					},
					"RT_WEEKEND": {
						ID:              "RT_WEEKEND",
						Weight:          10,
						ActivationTimes: "* * * * 0,6",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 0,
								RecurrentFee:  0.06,
								Unit:          time.Minute,
								Increment:     time.Second,
							},
						},
					},
					"RT_CHRISTMAS": {
						ID:              "RT_CHRISTMAS",
						Weight:          30,
						ActivationTimes: "* * 24 12 *",
						IntervalRates: []*engine.IntervalRate{
							{
								IntervalStart: 0,
								RecurrentFee:  0.06,
								Unit:          time.Minute,
								Increment:     time.Second,
							},
						},
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
	if err := ratePrfRpc.Call(utils.APIerSv1SetRateProfileRates, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	rPrf.Tenant = "cgrates.org"
	var rply *engine.RateProfile
	if err := ratePrfRpc.Call(utils.APIerSv1GetRateProfile,
		utils.TenantIDWithOpts{TenantID: &utils.TenantID{ID: "SpecialRate"}},
		&rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rPrf.RateProfile, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(rPrf.RateProfile), utils.ToJSON(rply))
	}
}

func testV1RatePrfRemoveRateProfileRatesWithoutTenant(t *testing.T) {
	var reply string
	if err := ratePrfRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRates{ID: "SpecialRate"},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1RateCostForEventWithDefault(t *testing.T) {
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.12,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.06,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	rPrf := &RateProfileWithCache{
		RateProfileWithOpts: &engine.RateProfileWithOpts{
			RateProfile: &engine.RateProfile{
				ID:        "DefaultRate",
				FilterIDs: []string{"*string:~*req.Subject:1001"},
				Weight:    10,
				Rates: map[string]*engine.Rate{
					"RATE1": rate1,
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
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.12,
		RateSIntervals: []*engine.RateSInterval{{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{{
				UsageStart:        0,
				Usage:             time.Minute,
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
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "2m10s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.12,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.06,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.19,
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
						UsageStart:        time.Minute,
						Usage:             time.Minute + 10*time.Second,
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "4h10m15s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	exp2 := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 15.075,
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
						UsageStart:        time.Minute,
						Usage:             4*time.Hour + 9*time.Minute + 15*time.Second,
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
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesUsage: "wrongUsage",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: time: invalid duration \"wrongUsage\"" {
		t.Errorf("Expected %+v \n, received %+v", "SERVER_ERROR: time: invalid duration \"wrongUsage\"", err)
	}
}

func testV1RateCostForEventWithStartTime(t *testing.T) {
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.12,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.06,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}

	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.12,
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC).String(),
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
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
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: "wrongTime",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	if err := ratePrfRpc.Call(utils.RateSv1CostForEvent, &argsRt, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: Unsupported time format" {
		t.Errorf("Expected %+v \n, received %+v", "SERVER_ERROR: Unsupported time format", err)
	}
}

func testV1RateCostForEventWithOpts(t *testing.T) {
	var rply *engine.RateProfileCost
	argsRt := &utils.ArgsCostForEvent{
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.OptsRatesUsage:     "2m10s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.12,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.06,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	exp := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.19,
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
						UsageStart:        time.Minute,
						Usage:             time.Minute + 10*time.Second,
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
		CGREventWithOpts: &utils.CGREventWithOpts{
			Opts: map[string]interface{}{
				utils.OptsRatesStartTime: time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.OptsRatesUsage:     "4h10m15s",
			},
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     utils.UUIDSha1Prefix(),
				Event: map[string]interface{}{
					utils.Subject: "1001",
				},
			},
		},
	}
	exp2 := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 15.075,
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
						UsageStart:        time.Minute,
						Usage:             4*time.Hour + 9*time.Minute + 15*time.Second,
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

func testV1RateCostForEventSpecial(t *testing.T) {
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.10,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	rtChristmas := &engine.Rate{
		ID:              "RT_CHRISTMAS",
		Weight:          30,
		ActivationTimes: "* * 24 12 *",
		IntervalRates: []*engine.IntervalRate{{
			IntervalStart: 0,
			RecurrentFee:  0.06,
			Unit:          time.Minute,
			Increment:     time.Second,
		}},
	}
	rPrf := &RateProfileWithCache{
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
	rate1 := &engine.Rate{
		ID:              "RATE1",
		Weight:          0,
		ActivationTimes: "* * * * *",
		IntervalRates: []*engine.IntervalRate{
			{
				IntervalStart: 0,
				RecurrentFee:  0.20,
				Unit:          time.Minute,
				Increment:     time.Minute,
			},
			{
				IntervalStart: time.Minute,
				RecurrentFee:  0.10,
				Unit:          time.Minute,
				Increment:     time.Second,
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
				RecurrentFee:  0.08,
				Unit:          time.Minute,
				Increment:     time.Second,
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
				RecurrentFee:  0.05,
				Unit:          time.Minute,
				Increment:     time.Second,
			},
		},
	}
	rPrf := &RateProfileWithCache{
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
