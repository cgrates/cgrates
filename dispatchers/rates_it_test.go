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

package dispatchers

import (
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

var sTestsDspRPrf = []func(t *testing.T){
	testDspRPrfPing,
	testDspRPrfCostForEvent,
	testDspRPrfCostForEventWithoutFilters,
}

//Test start here
func TestDspRateSIT(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
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
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspRPrf, "TestDspRateSIT", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspRPrfPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.RateSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.RateSv1Ping, utils.CGREvent{
		Tenant: "cgrates.org",
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "rPrf12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspRPrfCostForEvent(t *testing.T) {
	rPrf := &engine.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weight:    10,
		Rates: map[string]*engine.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.12),
						Unit:          utils.Float64Pointer(float64(time.Minute)),
						Increment:     utils.Float64Pointer(float64(time.Minute)),
					},
				},
			},
		},
	}
	var reply string
	if err := allEngine.RPC.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK, received %+v", reply)
	}
	var rply *engine.RateProfile
	if err := allEngine.RPC.Call(utils.APIerSv1GetRateProfile, &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "DefaultRate",
	}, &rply); err != nil {
		t.Error(err)
	}
	rtWeek, err := rPrf.Rates["RT_WEEK"].AsRate()
	if err != nil {
		t.Fatal(err)
	}
	exp := &engine.RateProfileCost{
		ID:   "DefaultRate",
		Cost: 0.12,
		RateSIntervals: []*engine.RateSInterval{{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{{
				UsageStart:        0,
				Usage:             time.Minute,
				Rate:              rtWeek,
				IntervalRateIndex: 0,
				CompressFactor:    1,
			}},
			CompressFactor: 1,
		}},
	}

	var rpCost *engine.RateProfileCost
	if err := dispEngine.RPC.Call(utils.RateSv1CostForEvent, &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "DefaultRate",
			Event: map[string]interface{}{
				utils.Subject: "1001",
			},

			Opts: map[string]interface{}{
				utils.OptsAPIKey: "rPrf12345",
			}}}, &rpCost); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpCost, exp) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(exp), utils.ToJSON(rpCost))
	}
}

func testDspRPrfCostForEventWithoutFilters(t *testing.T) {
	rPrf := &engine.APIRateProfile{
		ID:     "ID_RP",
		Tenant: "cgrates.org",
		Weight: 10,
		Rates: map[string]*engine.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weight:          0,
				ActivationTimes: "* * * * *",
				IntervalRates: []*engine.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.25),
						Unit:          utils.Float64Pointer(float64(time.Minute)),
						Increment:     utils.Float64Pointer(float64(time.Second)),
					},
				},
			},
		},
	}
	var reply string
	if err := allEngine.RPC.Call(utils.APIerSv1SetRateProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expected OK, received %+v", reply)
	}
	var rply *engine.RateProfile
	if err := allEngine.RPC.Call(utils.APIerSv1GetRateProfile, &utils.TenantID{
		Tenant: "cgrates.org",
		ID:     "ID_RP",
	}, &rply); err != nil {
		t.Error(err)
	}
	rtWeek, err := rPrf.Rates["RT_WEEK"].AsRate()
	if err != nil {
		t.Fatal(err)
	}
	exp := &engine.RateProfileCost{
		ID:   "ID_RP",
		Cost: 0.25,
		RateSIntervals: []*engine.RateSInterval{{
			UsageStart: 0,
			Increments: []*engine.RateSIncrement{{
				UsageStart:        0,
				Usage:             time.Minute,
				Rate:              rtWeek,
				IntervalRateIndex: 0,
				CompressFactor:    60,
			}},
			CompressFactor: 1,
		}},
	}

	var rpCost *engine.RateProfileCost
	if err := dispEngine.RPC.Call(utils.RateSv1CostForEvent, &utils.ArgsCostForEvent{
		CGREvent: &utils.CGREvent{

			Tenant: "cgrates.org",
			ID:     "EVENT_RATE",
			Event: map[string]interface{}{
				utils.Subject: "1002",
			},

			Opts: map[string]interface{}{
				utils.OptsAPIKey: "rPrf12345",
			}}}, &rpCost); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rpCost, exp) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(exp), utils.ToJSON(rpCost))
	}
}
