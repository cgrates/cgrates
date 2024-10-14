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
	"path"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrCfgPath         string
	fltrCfg             *config.CGRConfig
	fltrRpc             *birpc.Client
	fltrConfDIR         string //run tests for specific configuration
	fltrDelay           int
	fltrInternalRestart bool // to reset db on internal restart engine

	sTestsFltr = []func(t *testing.T){
		testV1FltrLoadConfig,
		testV1FltrInitDataDb,
		testV1FltrResetStorDb,
		testV1FltrStartEngine,
		testV1FltrRpcConn,
		testV1FltrLoadTarrifPlans,
		testV1FltrAddStats,
		testV1FltrPopulateThreshold,
		testV1FltrGetThresholdForEvent,
		testV1FltrGetThresholdForEvent2,
		testV1FltrPopulateResources,
		testV1FltrPopulateResourcesAvailableUnits,
		testV1FltrAccounts,
		testV1FltrAccountsExistsDynamicaly,
		testV1FltrAttributesPrefix,
		testV1FltrInitDataDb,
		testV1FltrChargerSuffix,
		testV1FltrPopulateTimings,
		testV1FltrStopEngine,
	}
)

// Test start here
func TestFltrIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		fltrConfDIR = "filters_internal"
	case utils.MetaMySQL:
		fltrConfDIR = "filters_mysql"
	case utils.MetaMongo:
		fltrConfDIR = "filters_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsFltr {
		t.Run(fltrConfDIR, stest)
	}
}

func testV1FltrLoadConfig(t *testing.T) {
	var err error
	fltrCfgPath = path.Join(*utils.DataDir, "conf", "samples", fltrConfDIR)
	if *utils.Encoding == utils.MetaGOB {
		cdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", fltrConfDIR+"_gob")
	}
	if fltrCfg, err = config.NewCGRConfigFromPath(fltrCfgPath); err != nil {
		t.Error(err)
	}
	fltrDelay = 1000
}

func testV1FltrInitDataDb(t *testing.T) {
	if *utils.DBType == utils.MetaInternal && fltrInternalRestart {
		testV1FltrStopEngine(t)
		testV1FltrStartEngine(t)
		testV1FltrRpcConn(t)
		return
	}
	fltrInternalRestart = true
	if err := engine.InitDataDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fltrCfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fltrCfgPath, fltrDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1FltrRpcConn(t *testing.T) {
	fltrRpc = engine.NewRPCClient(t, fltrCfg.ListenCfg())
}

func testV1FltrLoadTarrifPlans(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FltrAddStats(t *testing.T) {
	var reply []string
	expected := []string{"Stat_1"}
	ev1 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         10.0,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         10.5,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        5 * time.Second,
			utils.Cost:         12.5,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_2"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "1002",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        6 * time.Second,
			utils.Cost:         17.5,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_3"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]any{
			utils.AccountField: "1003",
			utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:        11 * time.Second,
			utils.Cost:         12.5,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]any{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      11 * time.Second,
			utils.Cost:       12.5,
			utils.PDD:        12 * time.Second,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	expected = []string{"Stat_1_1"}
	ev1 = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]any{
			"Stat":           "Stat1_1",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      15 * time.Second,
			utils.Cost:       15.5,
			utils.PDD:        15 * time.Second,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, &ev1, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1FltrPopulateThreshold(t *testing.T) {
	//Add a filter of type *stats and check if acd metric is minim 10 ( greater than 10)
	//we expect that acd from Stat_1 to be 11 so the filter should pass (11 > 10)
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Stats1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*stats.Stat_1.*acd",
					Values:  []string{"10.0"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaLog},
	}}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &result); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", result)
	}

	//Add a threshold with filter from above and an inline filter for Account 1010
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Stats1",
			FilterIDs: []string{"FLTR_TH_Stats1", "*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}
}

func testV1FltrGetThresholdForEvent(t *testing.T) {
	// check the event
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010"},
	}
	var ids []string
	eIDs := []string{"TH_Stats1"}
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
}

func testV1FltrGetThresholdForEvent2(t *testing.T) {
	//Add a filter of type *stats and check if acd metric is maximum 10 ( lower than 10)
	//we expect that acd from Stat_1 to be 11 so the filter should not pass (11 > 10)
	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Stats1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*lt",
					Element: "~*stats.Stat_1.*acd",
					Values:  []string{"10.0"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//update the threshold with new filter
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Stats1",
			FilterIDs: []string{"FLTR_TH_Stats1", "*string:~*req.Account:1010"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010"},
	}
	var ids []string
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrPopulateResources(t *testing.T) {
	//create a resourceProfile
	rlsConfig := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "ResTest",
		UsageTTL:          time.Minute,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, &engine.ResourceProfileWithAPIOpts{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ResourceProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}

	// Allocate 3 units for resource ResTest
	argsRU := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "3001",
			"Destination": "3002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
			utils.OptsResourcesUnits:   3,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	}

	//we allocate 3 units to resource and add a filter for Usages > 2
	//should match (3>2)
	filter := engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Resource",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*resources.ResTest.TotalUsage",
					Values:  []string{"2.0"},
				},
			},
		},
	}

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_ResTest",
			FilterIDs: []string{"FLTR_TH_Resource", "*string:~*req.Account:2020"},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weight:    10.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}

	// check the event
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "2020"},
	}

	var ids []string
	eIDs := []string{"TH_ResTest"}
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}

	//change the filter
	//we allocate 3 units to resource and add a filter for Usages < 2
	//should fail (3<2)
	filter.Filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_Resource",
		Rules: []*engine.FilterRule{
			{
				Type:    "*lt",
				Element: "~*resources.ResTest.TotalUsage",
				Values:  []string{"2.0"},
			},
		},
	}

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//Overwrite the threshold
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//expect NotFound error because filter doesn't match
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrPopulateResourcesAvailableUnits(t *testing.T) {
	//create a resourceProfile
	rlsConfig := &engine.ResourceProfile{
		Tenant:            "cgrates.org",
		ID:                "RES_TEST",
		UsageTTL:          time.Minute,
		Limit:             23,
		AllocationMessage: "Test_Available",
		Stored:            true,
		Weight:            25,
		ThresholdIDs:      []string{utils.MetaNone},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetResourceProfile, &engine.ResourceProfileWithAPIOpts{ResourceProfile: rlsConfig}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ResourceProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetResourceProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: rlsConfig.ID}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, rlsConfig) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(rlsConfig), utils.ToJSON(reply))
	}

	//Allocate 9 units for resource ResTest
	argsRU := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     utils.UUIDSha1Prefix(),
		Event: map[string]any{
			"Account":     "3001",
			"Destination": "3002",
		},
		APIOpts: map[string]any{
			utils.OptsResourcesUsageID: "651a8db2-4f67-4cf8-b622-169e8a482e21",
			utils.OptsResourcesUnits:   9,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.ResourceSv1AllocateResources, argsRU, &result); err != nil {
		t.Error(err)
	} else if result != "Test_Available" {
		t.Error("Unexpected reply returned", result)
	}

	//as we allocate 9 units, there should be available 14 more
	//our filter should match for *gt or *gte
	filter := engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_ST_Resource1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*resources.RES_TEST.Available",
					Values:  []string{"13.0"},
				},
				{
					Type:    "*gte",
					Element: "~*resources.RES_TEST.Available",
					Values:  []string{"14.0"},
				},
			},
		},
	}

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//set a statQueueProfile with that filter
	statsPrf := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "STATS_RES_TEST12",
			FilterIDs: []string{"FLTR_ST_Resource1", "*string:~*req.Account:1001"},
			Weight:    50,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetStatQueueProfile, statsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var replyStats *engine.StatQueueProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile, &utils.TenantID{Tenant: "cgrates.org",
		ID: "STATS_RES_TEST12"}, &replyStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statsPrf.StatQueueProfile, replyStats) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(statsPrf.StatQueueProfile), utils.ToJSON(replyStats))
	}

	//here will check the event
	statsEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event_nr2",
		Event: map[string]any{
			utils.AccountField: "1001",
			utils.Usage:        "1",
		},
	}
	var ids []string
	expectedIDs := []string{"STATS_RES_TEST12", "Stat_1"}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, statsEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDs, ids) {
		t.Errorf("Expected %+v, received %+v", expectedIDs, ids)
	}

	//set another filter that will not match
	filter = engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_ST_Resource1",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*resources.RES_TEST.Available",
					Values:  []string{"17.0"},
				},
			},
		},
	}

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//overwrite the StatQueueProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetStatQueueProfile, &utils.TenantID{Tenant: "cgrates.org",
		ID: "STATS_RES_TEST12"}, &replyStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statsPrf.StatQueueProfile, replyStats) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(statsPrf.StatQueueProfile), utils.ToJSON(replyStats))
	}

	//This filter won't match
	expectedIDs = []string{"Stat_1"}
	if err := fltrRpc.Call(context.Background(), utils.StatSv1ProcessEvent, statsEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDs, ids) {
		t.Errorf("Expected %+v, received %+v", expectedIDs, ids)
	}
}

func testV1FltrAccounts(t *testing.T) {
	var resp string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	// Add a filter with fieldName taken value from account 1001
	// and check if *monetary balance is minim 9 ( greater than 9)
	// we expect that the balance to be 10 so the filter should pass (10 > 9)
	filter := engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TH_Accounts",
			Rules: []*engine.FilterRule{
				{
					Type:    "*gt",
					Element: "~*accounts.1001.BalanceMap.*monetary[0].Value",
					Values:  []string{"9"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	// Add a log action
	attrsAA := &utils.AttrSetActions{ActionsId: "LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaLog},
	}}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &result); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", result)
	}
	//Add a threshold with filter from above and an inline filter for Account 1010
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_Account",
			FilterIDs: []string{"FLTR_TH_Accounts", "*string:~*req.Account:1001"},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weight:    90.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1001"},
	}
	var ids []string
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"TH_Account"}) {
		t.Error("Unexpected reply returned", ids)
	}

	// update the filter
	// Add a filter with fieldName taken value from account 1001
	// and check if *monetary balance is is minim 11 ( greater than 11)
	// we expect that the balance to be 10 so the filter should not pass (10 > 11)
	filter.Filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_TH_Accounts",
		Rules: []*engine.FilterRule{
			{
				Type:    "*gt",
				Element: "~*accounts.1001.BalanceMap.*monetary[0].Value",
				Values:  []string{"11"},
			},
		},
	}

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrAccountsExistsDynamicaly(t *testing.T) {
	var resp string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "TH_Account"}}, &resp); err != nil {
		if err.Error() != utils.ErrNotFound.Error() { // no error if the threshold is already removed
			t.Error(err)
		}
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	var result string
	// Add a log action
	attrsAA := &utils.AttrSetActions{ActionsId: "LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaLog},
	}}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &result); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	}
	//Add a threshold with filter from above and an inline filter for Account 1010
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TH_AccountDinamic",
			FilterIDs: []string{"*exists:~*accounts.<~*req.Account>:"},
			MaxHits:   -1,
			MinSleep:  time.Millisecond,
			Weight:    90.0,
			ActionIDs: []string{"LOG"},
			Async:     true,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var rcvTh *engine.ThresholdProfile
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tPrfl.Tenant, ID: tPrfl.ID}, &rcvTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rcvTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rcvTh)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1001"},
	}
	var ids []string
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"TH_AccountDinamic"}) {
		t.Error("Unexpected reply returned", ids)
	}

	tEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]any{
			utils.AccountField: "non"},
	}
	ids = nil
	if err := fltrRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, tEv, &ids); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FltrChargerSuffix(t *testing.T) {
	var reply string
	if err := fltrRpc.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "IntraCharger",
			FilterIDs:    []string{"*suffix:~*req.Subject:intra"},
			RunID:        "Intra",
			AttributeIDs: []string{"*constant:*req.Subject:intraState"},
			Weight:       20,
		},
	}
	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	chargerProfile2 := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "InterCharger",
			FilterIDs:    []string{"*suffix:~*req.Subject:inter"},
			RunID:        "Inter",
			AttributeIDs: []string{"*constant:*req.Subject:interState"},
			Weight:       20,
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	processedEv := []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "IntraCharger",
			AttributeSProfiles: []string{"*constant:*req.Subject:intraState"},
			AlteredFields:      []string{utils.MetaReqRunID, "*req.Subject"},

			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					utils.AccountField: "1010",
					utils.Subject:      "intraState",
					utils.RunID:        "Intra",
					utils.Destination:  "999",
				},
				APIOpts: map[string]any{
					utils.MetaSubsys:               utils.MetaChargers,
					utils.OptsAttributesProfileIDs: []any{"*constant:*req.Subject:intraState"},
				},
			},
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Subject:      "Something_intra",
			utils.Destination:  "999",
		},
	}
	var result2 []*engine.ChrgSProcessEventReply
	if err := fltrRpc.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, processedEv) {
		t.Errorf("Expecting : %s, \n received: %s", utils.ToJSON(processedEv), utils.ToJSON(result2))
	}

	processedEv = []*engine.ChrgSProcessEventReply{
		{
			ChargerSProfile:    "InterCharger",
			AttributeSProfiles: []string{"*constant:*req.Subject:interState"},
			AlteredFields:      []string{utils.MetaReqRunID, "*req.Subject"},

			CGREvent: &utils.CGREvent{ // matching Charger1
				Tenant: "cgrates.org",
				ID:     "event1",
				Event: map[string]any{
					utils.AccountField: "1010",
					utils.Subject:      "interState",
					utils.RunID:        "Inter",
					utils.Destination:  "999",
				},
				APIOpts: map[string]any{
					utils.MetaSubsys:               utils.MetaChargers,
					utils.OptsAttributesProfileIDs: []any{"*constant:*req.Subject:interState"},
				},
			},
		},
	}
	cgrEv = &utils.CGREvent{

		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]any{
			utils.AccountField: "1010",
			utils.Subject:      "Something_inter",
			utils.Destination:  "999",
		},
	}
	if err := fltrRpc.Call(context.Background(), utils.ChargerSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, processedEv) {
		t.Errorf("Expecting : %s, \n received: %s", utils.ToJSON(processedEv), utils.ToJSON(result2))
	}
}

func testV1FltrAttributesPrefix(t *testing.T) {
	chargerProfile := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.new",
			ID:        "ATTR_1001",
			FilterIDs: []string{"*prefix:~*req.CustomField:2007|+2007", "*prefix:~*req.CustomField2:2007|+2007", "FLTR_1"},
			Contexts:  []string{"prefix"},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{},
					Path:      utils.MetaReq + utils.NestingSep + "CustomField",
					Type:      utils.MetaConstant,
					Value:     config.NewRSRParsersMustCompile("2007", utils.InfieldSep),
				},
			},
			Weight: 20.0,
		},
	}
	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetAttributeProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	processedEv := &engine.AttrSProcessEventReply{
		AlteredFields:   []string{"*req.CustomField"},
		MatchedProfiles: []string{"cgrates.new:ATTR_1001"},

		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.new",
			ID:     "event1",
			Event: map[string]any{
				"CustomField":     "2007",
				"CustomField2":    "+2007",
				utils.Destination: "+1207",
			},
			APIOpts: map[string]any{
				utils.OptsContext: "prefix",
			},
		},
	}
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.new",
		ID:     "event1",
		Event: map[string]any{
			"CustomField":     "+2007",
			"CustomField2":    "+2007",
			utils.Destination: "+1207",
		},
		APIOpts: map[string]any{
			utils.OptsContext: "prefix",
		},
	}
	var result2 *engine.AttrSProcessEventReply
	if err := fltrRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent, cgrEv, &result2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result2, processedEv) {
		t.Errorf("Expecting : %s, \n received: %s", utils.ToJSON(processedEv), utils.ToJSON(result2))
	}

}

func testV1FltrPopulateTimings(t *testing.T) {
	timing := &utils.TPTimingWithAPIOpts{
		TPTiming: &utils.TPTiming{
			ID:        "TM_MORNING",
			WeekDays:  utils.WeekDays{1, 2, 3, 4, 5},
			StartTime: "08:00:00",
			EndTime:   "09:00:00",
		},
	}

	var reply string

	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetTiming, timing, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	filter := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_TM_1",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaTimings,
					Element: "~*req.AnswerTime",
					Values:  []string{"TM_MORNING"},
				},
			},
		},
	}

	var result string
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "FltrTest",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"FLTR_TM_1"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + utils.AnswerTime,
					Value: config.NewRSRParsersMustCompile("2021-04-29T10:45:00Z", utils.InfieldSep),
				},
			},
			Weight: 10,
		},
	}
	attrPrf.Compile()
	if err := fltrRpc.Call(context.Background(), utils.APIerSv1SetAttributeProfile, attrPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testV1FltrPopulateTimings",
		Event: map[string]any{
			utils.AnswerTime: "2021-04-29T08:35:00Z",
		},
		APIOpts: map[string]any{
			utils.OptsContext:              utils.MetaAny,
			utils.OptsAttributesProfileIDs: []string{"FltrTest"},
		},
	}
	eRply := &engine.AttrSProcessEventReply{
		MatchedProfiles: []string{"cgrates.org:FltrTest"},
		AlteredFields:   []string{utils.MetaReq + utils.NestingSep + utils.AnswerTime},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1FltrPopulateTimings",
			Event: map[string]any{
				utils.AnswerTime: "2021-04-29T10:45:00Z",
			},
			APIOpts: map[string]any{
				utils.OptsAttributesProfileIDs: []any{"FltrTest"},
				utils.OptsContext:              utils.MetaAny,
			},
		},
	}

	var rplyEv1 engine.AttrSProcessEventReply
	if err := fltrRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv1); err != nil {
		t.Error(err)
	} else {
		sort.Strings(eRply.AlteredFields)
		sort.Strings(rplyEv1.AlteredFields)
		if !reflect.DeepEqual(eRply, &rplyEv1) {
			t.Errorf("\nexpected: %s, \nreceived: %s",
				utils.ToJSON(eRply), utils.ToJSON(rplyEv1))
		}
	}

	ev.Event[utils.AnswerTime] = "2021-04-29T13:35:00Z"

	var rplyEv2 engine.AttrSProcessEventReply
	if err := fltrRpc.Call(context.Background(), utils.AttributeSv1ProcessEvent,
		ev, &rplyEv2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %+v,%+v", err, utils.ErrNotFound)
	}
}

func testV1FltrStopEngine(t *testing.T) {
	if err := engine.KillEngine(accDelay); err != nil {
		t.Error(err)
	}
}
