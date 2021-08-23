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
package v1

import (
	"net/rpc"
	"path"
	"reflect"
	"sort"
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
	splPrf        *SupplierWithCache
	splSv1ConfDIR string //run tests for specific configuration

	sTestsSupplierSV1 = []func(t *testing.T){
		testV1SplSLoadConfig,
		testV1SplSInitDataDb,
		testV1SplSResetStorDb,
		testV1SplSStartEngine,
		testV1SplSRpcConn,
		testV1SplSFromFolder,
		testV1SplSGetWeightSuppliers,
		testV1SplSGetLeastCostSuppliers,
		testV1SplSGetLeastCostSuppliersWithoutUsage,
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
		testV1SplSGetSupplierForEvent,
		// reset the database and load the TP again
		testV1SplSInitDataDb,
		testV1SplSFromFolder,
		// for the moment we decide to comment the tests
		// testV1SplsOneSupplierWithoutDestination,
		testV1SplMultipleSupplierSameID,
		testV1SplSupplierPing,
		testV1SplSStopEngine,
	}
)

// Test start here
func TestSuplSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		splSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		splSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		splSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*dataDir, "conf", "samples", splSv1ConfDIR)
	if splSv1Cfg, err = config.NewCGRConfigFromPath(splSv1CfgPath); err != nil {
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
	splSv1Rpc, err = newRPCClient(splSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1SplSGetWeightSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
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
		Count:     2,
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
		CGREvent: &utils.CGREvent{
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
		Sorting:   utils.MetaLC,
		Count:     3,
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
					utils.Cost:         1.2667,
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

func testV1SplSGetLeastCostSuppliersWithoutUsage(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1003",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLC,
		Count:     3,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.0102,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2,
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
		CGREvent: &utils.CGREvent{
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
		Sorting:   utils.MetaLC,
		Count:     2,
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
		CGREvent: &utils.CGREvent{
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
		CGREvent: &utils.CGREvent{
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
		Sorting:   utils.MetaLC,
		Count:     2,
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
		CGREvent: &utils.CGREvent{
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
		Sorting:   utils.MetaHC,
		Count:     3,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         1.2667,
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
		CGREvent: &utils.CGREvent{
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
	ev1 := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       10.0,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       10.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(5 * time.Second),
				utils.COST:       12.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(6 * time.Second),
				utils.COST:       17.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_3"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.Account:    "1003",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       12.5,
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Stat":           "Stat1_1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(11 * time.Second),
				utils.COST:       12.5,
				utils.PDD:        time.Duration(12 * time.Second),
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				"Stat":           "Stat1_1",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(15 * time.Second),
				utils.COST:       15.5,
				utils.PDD:        time.Duration(15 * time.Second),
			},
		},
	}
	if err := splSv1Rpc.Call(utils.StatSv1ProcessEvent, ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1SplSGetQOSSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos",
			},
		},
	}
	expSupplierIDs := []string{"supplier1", "supplier3", "supplier2"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_QOS_1" {
			t.Errorf("Expecting: SPL_QOS_1, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSSuppliers2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos2",
			},
		},
	}
	expSupplierIDs := []string{"supplier3", "supplier2", "supplier1"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_QOS_2" {
			t.Errorf("Expecting: SPL_QOS_2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSSuppliers3(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos3",
			},
		},
	}
	expSupplierIDs := []string{"supplier1", "supplier3", "supplier2"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_QOS_3" {
			t.Errorf("Expecting: SPL_QOS_3, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetQOSSuppliersFiltred(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetQOSSuppliers",
			Event: map[string]interface{}{
				"DistinctMatch": "*qos_filtred",
			},
		},
	}
	expSupplierIDs := []string{"supplier1", "supplier3"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_QOS_FILTRED" {
			t.Errorf("Expecting: SPL_QOS_FILTRED, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(suplsReply))
		}
	}
}

func testV1SplSGetQOSSuppliersFiltred2(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
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
	expSupplierIDs := []string{"supplier3", "supplier2"}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else {
		rcvSupl := make([]string, len(suplsReply.SortedSuppliers))
		for i, supl := range suplsReply.SortedSuppliers {
			rcvSupl[i] = supl.SupplierID
		}
		if suplsReply.ProfileID != "SPL_QOS_FILTRED2" {
			t.Errorf("Expecting: SPL_QOS_FILTRED2, received: %s",
				suplsReply.ProfileID)
		}
		if !reflect.DeepEqual(rcvSupl, expSupplierIDs) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				expSupplierIDs, utils.ToJSON(rcvSupl))
		}
	}
}

func testV1SplSGetSupplierWithoutFilter(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
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
		Count:     1,
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
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
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
		},
	}

	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
}

func testV1SplSGetSupplierProfileIDs(t *testing.T) {
	expected := []string{"SPL_HIGHESTCOST_1", "SPL_QOS_1", "SPL_QOS_2", "SPL_QOS_FILTRED", "SPL_QOS_FILTRED2",
		"SPL_ACNT_1001", "SPL_LEASTCOST_1", "SPL_WEIGHT_2", "SPL_WEIGHT_1", "SPL_QOS_3", "TEST_PROFILE1", "SPL_LCR"}
	var result []string
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfileIDs,
		&utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{"cgrates.org"}}, &result); err != nil {
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
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.Suppliers, reply.Suppliers) && !reflect.DeepEqual(reverseSuppliers, reply.Suppliers) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testV1SplSRemSupplierProfiles(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := splSv1Rpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testV1SplSupplierPing(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call(utils.SupplierSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1SplSGetSupplierForEvent(t *testing.T) {
	ev := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
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
	expected := engine.SupplierProfile{
		Tenant:    "cgrates.org",
		ID:        "SPL_LCR",
		FilterIDs: []string{"FLTR_TEST"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2017, 11, 27, 00, 00, 00, 00, time.UTC),
		},
		Sorting:           utils.MetaLC,
		SortingParameters: []string{},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:                 "supplier_1",
				FilterIDs:          nil,
				AccountIDs:         nil,
				RatingPlanIDs:      []string{"RP_TEST_1"},
				ResourceIDs:        nil,
				StatIDs:            nil,
				Weight:             10,
				Blocker:            false,
				SupplierParameters: "",
			},
			&engine.Supplier{
				ID:                 "supplier_2",
				FilterIDs:          nil,
				AccountIDs:         nil,
				RatingPlanIDs:      []string{"RP_TEST_2"},
				ResourceIDs:        nil,
				StatIDs:            nil,
				Weight:             0,
				Blocker:            false,
				SupplierParameters: "",
			},
		},
		Weight: 50,
	}
	if *encoding == utils.MetaGOB { // in gob emtpty slice is encoded as nil
		expected.SortingParameters = nil
	}
	var supProf []*engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSupplierProfilesForEvent,
		ev, &supProf); err != nil {
		t.Fatal(err)
	}
	sort.Slice(expected.Suppliers, func(i, j int) bool {
		return expected.Suppliers[i].Weight < expected.Suppliers[j].Weight
	})
	sort.Slice(supProf[0].Suppliers, func(i, j int) bool {
		return supProf[0].Suppliers[i].Weight < supProf[0].Suppliers[j].Weight
	})
	if !reflect.DeepEqual(expected, *supProf[0]) {
		t.Errorf("Expected: %s ,received: %s", utils.ToJSON(expected), utils.ToJSON(supProf))
	}
}

// Scenario: We create two rating plans RP_MOBILE and RP_LOCAL
// RP_LOCAL contains destination for both mobile and local
// and RP_MOBILE contains destinations only for mobile
// Create a SupplierProfile with *least_cost strategy with 2 suppliers
// supplier1 have attached RP_LOCAL and supplier2 have attach RP_MOBILE
func testV1SplsOneSupplierWithoutDestination(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SPL_DESTINATION"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "SPL_DESTINATION",
			FilterIDs: []string{"*string:~*req.Account:SpecialCase"},
			Sorting:   utils.MetaLC,
			Suppliers: []*engine.Supplier{
				{
					ID:            "local",
					RatingPlanIDs: []string{"RP_LOCAL"},
					Weight:        10,
				},
				{
					ID:            "mobile",
					RatingPlanIDs: []string{"RP_MOBILE"},
					FilterIDs:     []string{"*destinations:~*req.Destination:DST_MOBILE"},
					Weight:        10,
				},
			},
			Weight: 100,
		},
	}

	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplsOneSupplierWithoutDestination",
			Event: map[string]interface{}{
				utils.Account:     "SpecialCase",
				utils.Destination: "+24680",
				utils.SetupTime:   utils.MetaNow,
				utils.Usage:       "2m",
			},
		},
		IgnoreErrors: true,
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_DESTINATION",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "local",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0396,
					"RatingPlanID": "RP_LOCAL",
					utils.Weight:   10.0,
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

	//in case that we don't use ignore errors
	//we get an error for the second supplier
	ev.IgnoreErrors = false
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	}
}

func testV1SplMultipleSupplierSameID(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "MULTIPLE_ROUTES"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:    "cgrates.org",
			ID:        "MULTIPLE_ROUTES",
			FilterIDs: []string{"*string:~*req.Account:SpecialCase2"},
			Sorting:   utils.MetaLC,
			Suppliers: []*engine.Supplier{
				{
					ID:            "Route1",
					RatingPlanIDs: []string{"RP_LOCAL"},
					FilterIDs:     []string{"*string:~*req.Month:April"},
					Weight:        10,
				},
				{
					ID:            "Route1",
					RatingPlanIDs: []string{"RP_MOBILE"},
					FilterIDs:     []string{"*string:~*req.Month:May"},
					Weight:        10,
				},
			},
			Weight: 100,
		},
	}

	var result string
	if err := splSv1Rpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tNow := time.Now()
	ev := &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteMultipleRouteSameID",
			Event: map[string]interface{}{
				utils.Account:     "SpecialCase2",
				utils.Destination: "+135876",
				utils.SetupTime:   utils.MetaNow,
				utils.Usage:       "2m",
				"Month":           "April",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "MULTIPLE_ROUTES",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "Route1",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0396,
					"RatingPlanID": "RP_LOCAL",
					utils.Weight:   10.0,
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

	ev = &engine.ArgsGetSuppliers{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Time:   &tNow,
			ID:     "testV1RouteMultipleRouteSameID",
			Event: map[string]interface{}{
				utils.Account:     "SpecialCase2",
				utils.Destination: "+135876",
				utils.SetupTime:   utils.MetaNow,
				utils.Usage:       "2m",
				"Month":           "May",
			},
		},
	}
	eSpls = engine.SortedSuppliers{
		ProfileID: "MULTIPLE_ROUTES",
		Sorting:   utils.MetaLC,
		Count:     1,
		SortedSuppliers: []*engine.SortedSupplier{
			{
				SupplierID: "Route1",
				SortingData: map[string]interface{}{
					utils.Cost:     0.0204,
					"RatingPlanID": "RP_MOBILE",
					utils.Weight:   10.0,
				},
			},
		},
	}
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
