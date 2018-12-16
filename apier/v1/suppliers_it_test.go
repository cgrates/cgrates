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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	splSv1CfgPath string
	splSv1Cfg     *config.CGRConfig
	splSv1Rpc     *rpc.Client
	splPrf        *engine.SupplierProfile
	splSv1ConfDIR string //run tests for specific configuration
)

var sTestsSupplierSV1 = []func(t *testing.T){
	testV1SplSLoadConfig,
	testV1SplSInitDataDb,
	testV1SplSResetStorDb,
	testV1SplSStartEngine,
	testV1SplSRpcConn,
	testV1SplSFromFolder,
	testV1SplSGetWeightSuppliers,
	testV1SplSGetLeastCostSuppliers,
	testV1SplSGetLeastCostSuppliersWithMaxCost,
	testV1SplSGetLeastCostSuppliersWithMaxCost2,
	testV1SplSGetLeastCostSuppliersWithMaxCostNotFound,
	testV1SplSGetHighestCostSuppliers,
	testV1SplSGetLeastCostSuppliersErr,
	testV1SplSPolulateStatsForQOS,
	testV1SplSGetQOSSuppliers,
	testV1SplSGetQOSSuppliers2,
	testV1SplSGetQOSSuppliers3,
	testV1SplSGetQOSSuppliersFiltred,
	testV1SplSGetQOSSuppliersFiltred2,
	testV1SplSGetSupplierWithoutFilter,
	testV1SplSSetSupplierProfiles,
	testV1SplSGetSupplierProfileIDs,
	testV1SplSUpdateSupplierProfiles,
	testV1SplSRemSupplierProfiles,
	testV1SplSupplierPing,
	testV1SplSStopEngine,
}

// Test start here
func TestSuplSV1ITMySQL(t *testing.T) {
	splSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func TestSuplSV1ITMongo(t *testing.T) {
	splSv1ConfDIR = "tutmongo"
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*dataDir, "conf", "samples", splSv1ConfDIR)
	if splSv1Cfg, err = config.NewCGRConfigFromFolder(splSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1SplSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1SplSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(splSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	var err error
	splSv1Rpc, err = jsonrpc.Dial("tcp", splSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1SplSGetWeightSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetWeightSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1007",
				utils.Destination: "+491511231234",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         0.46666,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostSuppliersWithMaxCost(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		MaxCost: "0.30",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostSuppliersWithMaxCostNotFound(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		MaxCost: "0.001",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSGetLeastCostSuppliersWithMaxCost2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		MaxCost: utils.MetaEventCost,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "SPECIAL_1002",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2014, 01, 14, 0, 0, 0, 0, time.UTC),
				utils.Usage:       "10m20s",
				utils.Category:    "call",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.1054,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetHighestCostSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
				"DistinctMatch":   "*highest_cost",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_HIGHESTCOST_1",
		Sorting:   utils.MetaHighestCost,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         0.46666,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0136,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostSuppliersErr(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		IgnoreErrors: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetHighestCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1000",
				utils.Destination: "1001",
				utils.SetupTime:   "*now",
				"Subject":         "TEST",
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSPolulateStatsForQOS(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.Account:    "1001",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       10.0,
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1001",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       10.5,
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(5 * time.Second),
			utils.COST:       12.5,
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(6 * time.Second),
			utils.COST:       17.5,
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_3"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.Account:    "1003",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       12.5,
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(11 * time.Second),
			utils.COST:       12.5,
			utils.PDD:        time.Duration(12 * time.Second),
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(15 * time.Second),
			utils.COST:       15.5,
			utils.PDD:        time.Duration(15 * time.Second),
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1SplSGetQOSSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_QOS_1",
		Sorting:   utils.MetaQOS,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"*acd:Stat_1":   11.0,
					"*acd:Stat_1_1": 13.0,
					"*asr:Stat_1":   100.0,
					"*pdd:Stat_1_1": 13.5,
					"*tcd:Stat_1":   22.0,
					"*tcd:Stat_1_1": 26.0,
					utils.Weight:    10.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"*acd:Stat_3": 11.0,
					"*asr:Stat_3": 100.0,
					"*tcd:Stat_3": 11.0,
					utils.Weight:  35.0,
				},
			},

			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"*acd:Stat_2": 5.5,
					"*asr:Stat_2": 100.0,
					"*tcd:Stat_2": 11.0,
					utils.Weight:  20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetQOSSuppliers2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos2",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_QOS_2",
		Sorting:   utils.MetaQOS,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"*acd:Stat_3": 11.0,
					"*asr:Stat_3": 100.0,
					"*tcd:Stat_3": 11.0,
					utils.Weight:  35.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"*acd:Stat_2": 5.5,
					"*asr:Stat_2": 100.0,
					"*tcd:Stat_2": 11.0,
					utils.Weight:  20.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"*acd:Stat_1":   11.0,
					"*acd:Stat_1_1": 13.0,
					"*asr:Stat_1":   100.0,
					"*pdd:Stat_1_1": 13.5,
					"*tcd:Stat_1":   22.0,
					"*tcd:Stat_1_1": 26.0,
					utils.Weight:    10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetQOSSuppliers3(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos3",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_QOS_3",
		Sorting:   utils.MetaQOS,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"*acd:Stat_1":   11.0,
					"*acd:Stat_1_1": 13.0,
					"*asr:Stat_1":   100.0,
					"*pdd:Stat_1_1": 13.5,
					"*tcd:Stat_1":   22.0,
					"*tcd:Stat_1_1": 26.0,
					utils.Weight:    10.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"*acd:Stat_3": 11.0,
					"*asr:Stat_3": 100.0,
					"*tcd:Stat_3": 11.0,
					utils.Weight:  35.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"*acd:Stat_2": 5.5,
					"*asr:Stat_2": 100.0,
					"*tcd:Stat_2": 11.0,
					utils.Weight:  20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetQOSSuppliersFiltred(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos_filtred",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_QOS_FILTRED",
		Sorting:   utils.MetaQOS,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"*acd:Stat_1":   11.0,
					"*acd:Stat_1_1": 13.0,
					"*asr:Stat_1":   100.0,
					"*pdd:Stat_1_1": 13.5,
					"*tcd:Stat_1":   22.0,
					"*tcd:Stat_1_1": 26.0,
					utils.Weight:    10.0,
				},
			},
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"*acd:Stat_3": 11.0,
					"*asr:Stat_3": 100.0,
					"*tcd:Stat_3": 11.0,
					utils.Weight:  35.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetQOSSuppliersFiltred2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch":   "*qos_filtred2",
				utils.Account:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_QOS_FILTRED2",
		Sorting:   utils.MetaQOS,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					"*acd:Stat_3": 11.0,
					"*asr:Stat_3": 100.0,
					"*tcd:Stat_3": 11.0,
					utils.Weight:  35.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"*acd:Stat_2":      5.5,
					"*asr:Stat_2":      100.0,
					"*tcd:Stat_2":      11.0,
					utils.Cost:         0.46666,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetSupplierWithoutFilter(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetSupplierWithoutFilter",
			Event: map[string]interface{}{
				utils.Account:     "1008",
				utils.Destination: "+49",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSSetSupplierProfiles(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &engine.SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "TEST_PROFILE1",
		FilterIDs:         []string{"FLTR_1"},
		Sorting:           "Sort1",
		SortingParameters: []string{"Param1", "Param2"},
		Suppliers: []*engine.Supplier{
			{
				ID:                 "SPL1",
				RatingPlanIDs:      []string{"RP1"},
				FilterIDs:          []string{"FLTR_1"},
				AccountIDs:         []string{"Acc"},
				ResourceIDs:        []string{"Res1", "ResGroup2"},
				StatIDs:            []string{"Stat1"},
				Weight:             20,
				Blocker:            false,
				SupplierParameters: "SortingParameter1",
			},
		},
		Weight: 10,
	}
	var result string
	if err := splSv1Rpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf, reply)
	}
}

func testV1SplSGetSupplierProfileIDs(t *testing.T) {
	expected := []string{"SPL_HIGHESTCOST_1", "SPL_QOS_1", "SPL_QOS_2", "SPL_QOS_FILTRED", "SPL_QOS_FILTRED2",
		"SPL_ACNT_1001", "SPL_LEASTCOST_1", "SPL_WEIGHT_2", "SPL_WEIGHT_1", "SPL_QOS_3", "TEST_PROFILE1", "SPL_LCR"}
	var result []string
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfileIDs", "cgrates.org", &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1SplSUpdateSupplierProfiles(t *testing.T) {
	splPrf.Suppliers = []*engine.Supplier{
		{
			ID:                 "SPL1",
			RatingPlanIDs:      []string{"RP1"},
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res1", "ResGroup2"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParameter1",
		},
		{
			ID:                 "SPL2",
			RatingPlanIDs:      []string{"RP2"},
			FilterIDs:          []string{"FLTR_2"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res2", "ResGroup2"},
			StatIDs:            []string{"Stat2"},
			Weight:             20,
			Blocker:            true,
			SupplierParameters: "SortingParameter2",
		},
	}
	reverseSuppliers := []*engine.Supplier{
		{
			ID:                 "SPL2",
			RatingPlanIDs:      []string{"RP2"},
			FilterIDs:          []string{"FLTR_2"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res2", "ResGroup2"},
			StatIDs:            []string{"Stat2"},
			Weight:             20,
			Blocker:            true,
			SupplierParameters: "SortingParameter2",
		},
		{
			ID:                 "SPL1",
			RatingPlanIDs:      []string{"RP1"},
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res1", "ResGroup2"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParameter1",
		},
	}
	var result string
	if err := splSv1Rpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.Suppliers, reply.Suppliers) && !reflect.DeepEqual(reverseSuppliers, reply.Suppliers) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testV1SplSRemSupplierProfiles(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call("ApierV1.RemoveSupplierProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSupplierPing(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call(utils.SupplierSv1Ping, "", &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
