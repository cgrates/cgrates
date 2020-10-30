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
	time.Sleep(500 * time.Millisecond)
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
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
						Value:         0.06,
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
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
						Value:         0.06,
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
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
						Value:         0.06,
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
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
						Value:         0.06,
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
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
		ConnectFee:       0.1,
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
	rateProfile := &engine.RateProfile{
		ID:               "RPWithoutTenant",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
				},
			},
		},
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
	} else if !reflect.DeepEqual(result, rateProfile) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(rateProfile), utils.ToJSON(result))
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
	rPrf := &engine.RateProfile{
		ID:               "SpecialRate",
		FilterIDs:        []string{"*string:~*req.Subject:1001"},
		Weight:           0,
		ConnectFee:       0.1,
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
						Value:         0.12,
						Unit:          time.Minute,
						Increment:     time.Minute,
					},
					{
						IntervalStart: time.Minute,
						Value:         0.06,
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
						Value:         0.06,
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
						Value:         0.06,
						Unit:          time.Minute,
						Increment:     time.Second,
					},
				},
			},
		},
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
	} else if !reflect.DeepEqual(rPrf, rply) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(rPrf), utils.ToJSON(rply))
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
