//go:build flaky

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

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRPrf = []func(t *testing.T){
	testDspRPrfPing,
	testDspRPrfCostForEvent,
	testDspRPrfCostForEventWithoutFilters,
}

// Test start here
func TestDspRateSIT(t *testing.T) {
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
	testDsp(t, sTestsDspRPrf, "TestDspRateSIT", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspRPrfPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.RateSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(context.Background(), utils.RateSv1Ping, utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsAPIKey: "rPrf12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspRPrfCostForEvent(t *testing.T) {
	rPrf := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:        "DefaultRate",
			Tenant:    "cgrates.org",
			FilterIDs: []string{"*string:~*req.Subject:1001"},
			Weights: utils.DynamicWeights{
				{
					Weight: 0,
				},
			},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
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
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Minute), 0),
						},
					},
				},
			},
		},
	}
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.AdminSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK, received %+v", reply)
	}
	var rply *utils.RateProfile
	if err := allEngine.RPC.Call(context.Background(), utils.AdminSv1GetRateProfile, &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DefaultRate",
	}, &rply); err != nil {
		t.Error(err)
	}

	exp := &utils.RateProfileCost{
		ID:   "DefaultRate",
		Cost: utils.NewDecimal(12, 2),
		CostIntervals: []*utils.RateSIntervalCost{{
			Increments: []*utils.RateSIncrementCost{{
				Usage:          utils.NewDecimal(int64(time.Minute), 0),
				RateID:         "ec268a8",
				CompressFactor: 1,
			}},
			CompressFactor: 1,
		}},
		Rates: map[string]*utils.IntervalRate{"ec268a8": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimalFromFloat64(0.12),
			Unit:          utils.NewDecimal(60000000000, 0),
			Increment:     utils.NewDecimal(60000000000, 0),
		}},
	}

	var rpCost *utils.RateProfileCost
	if err := dispEngine.RPC.Call(context.Background(), utils.RateSv1CostForEvent, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DefaultRate",
		Event: map[string]any{
			utils.Subject: "1001",
		},

		APIOpts: map[string]any{
			utils.OptsAPIKey: "rPrf12345",
		}}, &rpCost); err != nil {
		t.Error(err)
	} else if !rpCost.Equals(exp) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rpCost))
	}
}

func testDspRPrfCostForEventWithoutFilters(t *testing.T) {
	rPrf := &utils.APIRateProfile{
		RateProfile: &utils.RateProfile{
			ID:     "ID_RP",
			Tenant: "cgrates.org",
			Weights: utils.DynamicWeights{
				{
					Weight: 10,
				},
			},
			Rates: map[string]*utils.Rate{
				"RT_WEEK": {
					ID: "RT_WEEK",
					Weights: utils.DynamicWeights{
						{
							Weight: 0,
						},
					},
					ActivationTimes: "* * * * *",
					IntervalRates: []*utils.IntervalRate{
						{
							IntervalStart: utils.NewDecimal(0, 0),
							RecurrentFee:  utils.NewDecimal(25, 2),
							Unit:          utils.NewDecimal(int64(time.Minute), 0),
							Increment:     utils.NewDecimal(int64(time.Second), 0),
						},
					},
				},
			},
		},
	}
	var reply string
	if err := allEngine.RPC.Call(context.Background(), utils.AdminSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK, received %+v", reply)
	}
	var rply *utils.RateProfile
	if err := allEngine.RPC.Call(context.Background(), utils.AdminSv1GetRateProfile, &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "ID_RP",
	}, &rply); err != nil {
		t.Error(err)
	}

	exp := &utils.RateProfileCost{
		ID:   "ID_RP",
		Cost: utils.NewDecimal(25, 2),
		CostIntervals: []*utils.RateSIntervalCost{{
			Increments: []*utils.RateSIncrementCost{{
				Usage:          utils.NewDecimal(int64(time.Minute), 0),
				RateID:         "ec268a8",
				CompressFactor: 60,
			}},
			CompressFactor: 1,
		}},
		Rates: map[string]*utils.IntervalRate{"ec268a8": {
			IntervalStart: utils.NewDecimal(0, 0),
			RecurrentFee:  utils.NewDecimalFromFloat64(0.25),
			Unit:          utils.NewDecimal(60000000000, 0),
			Increment:     utils.NewDecimal(1000000000, 0),
		}},
	}

	var rpCost *utils.RateProfileCost
	if err := dispEngine.RPC.Call(context.Background(), utils.RateSv1CostForEvent, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EVENT_RATE",
		Event: map[string]any{
			utils.Subject: "1002",
		},
		APIOpts: map[string]any{
			utils.OptsAPIKey: "rPrf12345",
		}}, &rpCost); err != nil {
		t.Error(err)
	} else if !rpCost.Equals(exp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rpCost))
	}
}
