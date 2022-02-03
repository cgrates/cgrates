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
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	ratePrfCfgPath   string
	ratePrfCfg       *config.CGRConfig
	rateSRPC         *birpc.Client
	ratePrfConfigDIR string //run tests for specific configuration

	sTestsRatePrf = []func(t *testing.T){
		testRateSInitCfg,
		testRateSInitDataDb,
		testRateSResetStorDb,
		testRateSStartEngine,
		testRateSRPCConn,
		testGetRateProfileBeforeSet,
		testGetRateProfilesBeforeSet,
		testRateSetRateProfile,
		testRateGetRateProfileIDs,
		testRateGetRateProfiles,
		testRateGetRateCount,
		testGetRateProfileBeforeSet2,
		testRateSetRateProfile2,
		testRateGetRateIDs2,
		testRateGetRateProfiles2,
		testRateGetRateCount2,
		testRateRemoveRateProfile,
		testRateGetRateProfileIDs,
		testRateGetRateCount,
		testRateSetRateProfile3,
		testRateSetAttributeProfileBrokenReference,
		testRateRemoveRateProfileRates,
		testRateSetRateProfileRates,
		testRateSetRateProfilesWithPrefix,
		testRateSKillEngine,
	}
)

func TestRateSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		ratePrfConfigDIR = "rates_internal"
	case utils.MetaMongo:
		ratePrfConfigDIR = "rates_mongo"
	case utils.MetaMySQL:
		ratePrfConfigDIR = "rates_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRatePrf {
		t.Run(ratePrfConfigDIR, stest)
	}
}

func testRateSInitCfg(t *testing.T) {
	var err error
	ratePrfCfgPath = path.Join(*dataDir, "conf", "samples", ratePrfConfigDIR)
	ratePrfCfg, err = config.NewCGRConfigFromPath(context.Background(), ratePrfCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testRateSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(ratePrfCfg); err != nil {
		t.Fatal(err)
	}
}

func testRateSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(ratePrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testRateSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ratePrfCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRateSRPCConn(t *testing.T) {
	var err error
	rateSRPC, err = newRPCClient(ratePrfCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testGetRateProfileBeforeSet(t *testing.T) {
	var reply *utils.RateProfile
	var err error
	if err = rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testGetRateProfilesBeforeSet(t *testing.T) {
	var reply []*utils.RateProfile
	var args *utils.ArgsItemIDs
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		args, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
func testRateSetRateProfile(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_IT_TEST",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
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
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_IT_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: nil,
				Weight:    0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: []*utils.DynamicWeight{
					{
						FilterIDs: nil,
						Weight:    0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      nil,
						RecurrentFee:  nil,
						Unit:          nil,
						Increment:     nil,
					},
				},
			},
		},
	}
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result))
	}
}

func testRateGetRateProfileIDs(t *testing.T) {
	var reply []string
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_RATE_IT_TEST"}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testRateGetRateProfiles(t *testing.T) {
	var reply []*utils.RateProfile
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []*utils.RateProfile{
		{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_RATE_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: nil,
					Weight:    0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: []*utils.DynamicWeight{
						{
							FilterIDs: nil,
							Weight:    0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      nil,
							RecurrentFee:  nil,
							Unit:          nil,
							Increment:     nil,
						},
					},
				},
			},
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testRateGetRateCount(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfileCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Expected %+v \n, received %+v", 1, reply)
	}
}

func testGetRateProfileBeforeSet2(t *testing.T) {
	var reply *utils.TenantIDWithAPIOpts
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_SECOND",
			},
		}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testRateSetRateProfile2(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_IT_TEST_SECOND",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
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
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_IT_TEST_SECOND",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: nil,
				Weight:    0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: []*utils.DynamicWeight{
					{
						FilterIDs: nil,
						Weight:    0,
					},
				},
				ActivationTimes: "* * * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      nil,
						RecurrentFee:  nil,
						Unit:          nil,
						Increment:     nil,
					},
				},
			},
		},
	}
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_SECOND",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result))
	}
}

func testRateGetRateIDs2(t *testing.T) {
	var reply []string
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []string{"TEST_RATE_IT_TEST", "TEST_RATE_IT_TEST_SECOND"}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfileIDs,
		args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != len(expected) {
		t.Errorf("Expected %+v \n, received %+v", expected, reply)
	}
}

func testRateGetRateProfiles2(t *testing.T) {
	var reply []*utils.RateProfile
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []*utils.RateProfile{
		{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_RATE_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: nil,
					Weight:    0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: []*utils.DynamicWeight{
						{
							FilterIDs: nil,
							Weight:    0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      nil,
							RecurrentFee:  nil,
							Unit:          nil,
							Increment:     nil,
						},
					},
				},
			},
		},
		{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_RATE_IT_TEST_SECOND",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: nil,
					Weight:    0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: []*utils.DynamicWeight{
						{
							FilterIDs: nil,
							Weight:    0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      nil,
							RecurrentFee:  nil,
							Unit:          nil,
							Increment:     nil,
						},
					},
				},
			},
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	sort.Slice(reply, func(i, j int) bool {
		return (reply)[i].ID < (reply)[j].ID
	})
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testRateGetRateCount2(t *testing.T) {
	var reply int
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			Tenant: utils.CGRateSorg,
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfileCount,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != 2 {
		t.Errorf("Expected %+v \n, received %+v", 2, reply)
	}
}

func testRateRemoveRateProfile(t *testing.T) {
	var reply string
	args := &utils.TenantIDWithAPIOpts{
		TenantID: &utils.TenantID{
			ID:     "TEST_RATE_IT_TEST_SECOND",
			Tenant: utils.CGRateSorg,
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1RemoveRateProfile,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	//nothing to get from db
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_SECOND",
			},
		}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v \n, received %+v", utils.ErrNotFound, err)
	}
}

func testRateGetRateProfilesAfterRemove(t *testing.T) {
	var reply []*utils.RateProfile
	args := &utils.ArgsItemIDs{
		Tenant: "cgrates.org",
	}
	expected := []*utils.RateProfile{
		{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_RATE_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: nil,
					Weight:    0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: []*utils.DynamicWeight{
						{
							FilterIDs: nil,
							Weight:    0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      nil,
							RecurrentFee:  nil,
							Unit:          nil,
							Increment:     nil,
						},
					},
				},
			},
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		args, &reply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testRateSetAttributeProfileBrokenReference(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_IT_TEST_SECOND",
		FilterIDs:       []string{"invalid_filter_format", "*string:~*opts.*context:*sessions"},
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
	expectedErr := "SERVER_ERROR: broken reference to filter: <invalid_filter_format> for item with ID: cgrates.org:TEST_RATE_IT_TEST_SECOND"
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v \n, received %+v", expectedErr, err)
	}
}

func testRateSetRateProfile3(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_IT_TEST_THIRD",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
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
			"RT_MONTH": {
				ID:              "RT_MONTH",
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
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_IT_TEST_THIRD",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: nil,
				Weight:    0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: []*utils.DynamicWeight{
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
			"RT_MONTH": {
				ID: "RT_MONTH",
				Weights: []*utils.DynamicWeight{
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
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_THIRD",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result))
	}
}

func testRateRemoveRateProfileRates(t *testing.T) {
	var reply string
	args := &utils.RemoveRPrfRates{
		ID:      "TEST_RATE_IT_TEST_THIRD",
		Tenant:  utils.CGRateSorg,
		RateIDs: []string{"RT_WEEK"},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1RemoveRateProfileRates,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected %+v \n, received %+v", utils.OK, reply)
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_IT_TEST_THIRD",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				FilterIDs: nil,
				Weight:    0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_MONTH": {
				ID: "RT_MONTH",
				Weights: []*utils.DynamicWeight{
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
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_THIRD",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result))
	}

	/*
		// as we removed our RT_WEEK, we will add it back to our profile
		argsRate := &utils.APIRateProfile{
			Tenant:          utils.CGRateSorg,
			ID:              "TEST_RATE_IT_TEST_THIRD",
			FilterIDs:       []string{"*string:~*req.Account:dan"},
			Weights:         ";0",
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.APIRate{
				// RT_WEEK wich we added back we added back
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
		if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfileRates,
			argsRate, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("UNexpected reply returned: %v", reply)
		}
	*/
}

func testRateSetRateProfileRates(t *testing.T) {
	argsRate := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "TEST_RATE_IT_TEST_THIRD",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
		Weights:         ";0",
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.APIRate{
			// RT_MONTH rate will be updated
			"RT_MONTH": {
				ID:              "RT_MONTH",
				Weights:         "*exists:~*req.Destination:;25",
				ActivationTimes: "* 10 * * 1-5",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						FixedFee:      utils.Float64Pointer(0.4),
						Increment:     utils.Float64Pointer(float64(time.Second)),
						RecurrentFee:  utils.Float64Pointer(float64(time.Second)),
						Unit:          utils.Float64Pointer(1),
					},
				},
			},
			// RT_CHRISTMAS will be added as a new rate into our profile
			"RT_CHRISTMAS": {
				ID:              "RT_CHRISTMAS",
				Weights:         ";10",
				ActivationTimes: "1 * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.2),
						Unit:          utils.Float64Pointer(float64(time.Minute)),
						Increment:     utils.Float64Pointer(0),
					},
				},
			},
		},
	}
	var result *string
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfileRates,
		argsRate, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, utils.StringPointer("OK")) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON("OK"), utils.ToJSON(result))
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_RATE_IT_TEST_THIRD",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			// RT_WEEK that remains the same
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: []*utils.DynamicWeight{
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
			// RT_MONTH that was updated
			"RT_MONTH": {
				ID: "RT_MONTH",
				Weights: []*utils.DynamicWeight{
					{
						FilterIDs: []string{"*exists:~*req.Destination:"},
						Weight:    25,
					},
				},
				ActivationTimes: "* 10 * * 1-5",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						FixedFee:      utils.NewDecimal(4, 1),
						Increment:     utils.NewDecimal(int64(time.Second), 0),
						RecurrentFee:  utils.NewDecimal(int64(time.Second), 0),
						Unit:          utils.NewDecimal(1, 0),
					},
				},
			},
			// RT_YEAR that was uadded as a new rate
			"RT_CHRISTMAS": {
				ID: "RT_CHRISTMAS",
				Weights: []*utils.DynamicWeight{
					{
						Weight: 10,
					},
				},
				ActivationTimes: "1 * * * *",
				IntervalRates: []*utils.IntervalRate{
					{
						IntervalStart: utils.NewDecimal(0, 0),
						Increment:     utils.NewDecimal(0, 0),
						RecurrentFee:  utils.NewDecimal(2, 1),
						Unit:          utils.NewDecimal(int64(time.Minute), 0),
					},
				},
			},
		},
	}
	var result2 *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_RATE_IT_TEST_THIRD",
			},
		}, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result2))
	}
}

func testRateSetRateProfilesWithPrefix(t *testing.T) {
	ratePrf := &utils.APIRateProfile{
		Tenant:          utils.CGRateSorg,
		ID:              "PrefixTEST_RATE_IT_TEST",
		FilterIDs:       []string{"*string:~*req.Account:dan"},
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
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		ratePrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	expectedRate := &utils.RateProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "PrefixTEST_RATE_IT_TEST",
		FilterIDs: []string{"*string:~*req.Account:dan"},
		Weights: []*utils.DynamicWeight{
			{
				Weight: 0,
			},
		},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID: "RT_WEEK",
				Weights: []*utils.DynamicWeight{
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
	var result *utils.RateProfile
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "PrefixTEST_RATE_IT_TEST",
			},
		}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, expectedRate) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedRate), utils.ToJSON(result))
	}

	var reply2 []*utils.RateProfile
	args := &utils.ArgsItemIDs{
		Tenant:      "cgrates.org",
		ItemsPrefix: "PrefixTEST",
	}
	expected := []*utils.RateProfile{
		{
			Tenant:    utils.CGRateSorg,
			ID:        "PrefixTEST_RATE_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:dan"},
			Weights: []*utils.DynamicWeight{
				{
					FilterIDs: nil,
					Weight:    0,
				},
			},
			MaxCostStrategy: "*free",
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: []*utils.DynamicWeight{
						{
							FilterIDs: nil,
							Weight:    0,
						},
					},
					ActivationTimes: "* * * * 1-5",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							FixedFee:      nil,
							RecurrentFee:  nil,
							Unit:          nil,
							Increment:     nil,
						},
					},
				},
			},
		},
	}
	if err := rateSRPC.Call(context.Background(), utils.AdminSv1GetRateProfiles,
		args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expected), utils.ToJSON(reply2))
	}
}

//Kill the engine when it is about to be finished
func testRateSKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
