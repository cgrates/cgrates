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
	"math/rand"
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
	stsV1CfgPath string
	stsV1Cfg     *config.CGRConfig
	stsV1Rpc     *rpc.Client
	statConfig   *engine.StatQueueProfileWithAPIOpts
	stsV1ConfDIR string //run tests for specific configuration

	evs = []*utils.CGREvent{
		{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        135 * time.Second,
				utils.Cost:         123.0}},
		{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        45 * time.Second}},
		{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.SetupTime:    time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        0}},
	}

	sTestsStatSV1 = []func(t *testing.T){
		testV1STSLoadConfig,
		testV1STSInitDataDb,
		testV1STSStartEngine,
		testV1STSRpcConn,
		testV1STSFromFolder,
		testV1STSGetStats,
		testV1STSProcessEvent,
		testV1STSGetStatsAfterRestart,
		testV1STSSetStatQueueProfile,
		testV1STSGetStatQueueProfileIDs,
		testV1STSUpdateStatQueueProfile,
		testV1STSRemoveStatQueueProfile,
		testV1STSStatsPing,
		testV1STSProcessMetricsWithFilter,
		testV1STSProcessStaticMetrics,
		testV1STSProcessStatWithThreshold,
		testV1STSV1GetQueueIDs,
		testV1STSV1GetStatQueuesForEventWithoutTenant,
		testV1STSV1StatSv1GetQueueStringMetricsWithoutTenant,
		testV1STSV1StatSv1ResetAction,
		testV1STSGetStatQueueProfileWithoutTenant,
		testV1STSRemStatQueueProfileWithoutTenant,
		testV1STSProcessCDRStat,
		testV1STSOverWriteStats,
		testV1STSProcessStatWithThreshold2,
		testV1STSSimulateAccountUpdate,
		testV1STSGetStatQueueWithoutExpired,
		testV1STSGetStatQueueWithoutStored,
		testV1STSStopEngine,
		testV1STSStartEngine,
		testV1STSRpcConn,
		testV1STSCheckMetricsAfterRestart,
		testV1STSStopEngine,
	}
)

func init() {
	rand.Seed(time.Now().UnixNano()) // used in benchmarks
}

//Test start here
func TestSTSV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		stsV1ConfDIR = "tutinternal"
		sTestsStatSV1 = sTestsStatSV1[:len(sTestsStatSV1)-4]
	case utils.MetaMySQL:
		stsV1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		stsV1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsStatSV1 {
		t.Run(stsV1ConfDIR, stest)
	}
}

func testV1STSLoadConfig(t *testing.T) {
	var err error
	stsV1CfgPath = path.Join(*dataDir, "conf", "samples", stsV1ConfDIR)
	if stsV1Cfg, err = config.NewCGRConfigFromPath(stsV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1STSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(stsV1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1STSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(stsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1STSRpcConn(t *testing.T) {
	var err error
	stsV1Rpc, err = newRPCClient(stsV1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1STSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := stsV1Rpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1STSGetStats(t *testing.T) {
	var reply []string
	expectedIDs := []string{"Stats1"}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueIDs,
		&utils.TenantWithAPIOpts{Tenant: "cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDs, reply) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, reply)
	}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaASR: utils.NotAvailable,
		utils.MetaACD: utils.NotAvailable,
		utils.MetaTCC: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
		utils.MetaACC: utils.NotAvailable,
		utils.MetaPDD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     utils.NotAvailable,
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}},
		&metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSV1StatSv1GetQueueStringMetricsWithoutTenant(t *testing.T) {
	var reply []string
	expectedIDs := []string{"CustomStatProfile", "Stats1", "StaticStatQueue", "StatWithThreshold"}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueIDs,
		&utils.TenantWithAPIOpts{}, &reply); err != nil {
		t.Error(err)
	} else {
		sort.Strings(reply)
		sort.Strings(expectedIDs)
		if !reflect.DeepEqual(expectedIDs, reply) {
			t.Errorf("expecting: %+v, received reply: %s", expectedIDs, reply)
		}
	}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaACD: "12s",
		utils.MetaTCD: "18s",
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.CustomValue: "10",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: expectedIDs[0]}},
		&metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", utils.ToJSON(expectedMetrics), utils.ToJSON(metrics))
	}
}
func testV1STSV1StatSv1ResetAction(t *testing.T) {
	var reply string
	if err := stsV1Rpc.Call(utils.APIerSv2SetActions, &utils.AttrSetActions{
		ActionsId: "ACT_RESET_STS",
		Actions:   []*utils.TPAction{{Identifier: utils.MetaResetStatQueue, ExtraParameters: "cgrates.org:CustomStatProfile"}},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	attrs := utils.AttrExecuteAction{Tenant: "cgrates.org", ActionsId: "ACT_RESET_STS"}
	if err := stsV1Rpc.Call(utils.APIerSv1ExecuteAction, attrs, &reply); err != nil {
		t.Error(err)
	}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaACD: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.CustomValue: utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "CustomStatProfile"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", utils.ToJSON(expectedMetrics), utils.ToJSON(metrics))
	}
}

func testV1STSProcessEvent(t *testing.T) {
	var reply []string
	expected := []string{"Stats1"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.AccountField: "1001",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        135 * time.Second,
				utils.Cost:         123.0,
				utils.PDD:          12 * time.Second,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	//process with one event (should be N/A becaus MinItems is 2)
	expectedMetrics := map[string]string{
		utils.MetaASR: utils.NotAvailable,
		utils.MetaACD: utils.NotAvailable,
		utils.MetaTCC: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
		utils.MetaACC: utils.NotAvailable,
		utils.MetaPDD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     utils.NotAvailable,
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: utils.NotAvailable,
	}
	var metrics map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	//process with one event (should be N/A becaus MinItems is 2)
	expectedFloatMetrics := map[string]float64{
		utils.MetaASR: -1.0,
		utils.MetaACD: -1.0,
		utils.MetaTCC: -1.0,
		utils.MetaTCD: -1.0,
		utils.MetaACC: -1.0,
		utils.MetaPDD: -1.0,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     -1.0,
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: -1.0,
	}
	var floatMetrics map[string]float64
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}}, &floatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFloatMetrics, floatMetrics) {
		t.Errorf("expecting: %+v, received reply: %+v", expectedFloatMetrics, floatMetrics)
	}

	args2 := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        45 * time.Second,
				utils.Cost:         12.1,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args2, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	args3 := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.AccountField: "1002",
				utils.SetupTime:    time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:        0,
				utils.Cost:         0,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args3, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	expectedMetrics2 := map[string]string{
		utils.MetaASR: "66.66667%",
		utils.MetaACD: "1m0s",
		utils.MetaACC: "45.03333",
		utils.MetaTCD: "3m0s",
		utils.MetaTCC: "135.1",
		utils.MetaPDD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     "180000000000",
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: "60000000000",
	}
	var metrics2 map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}}, &metrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics2, metrics2) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics2, metrics2)
	}

	expectedFloatMetrics2 := map[string]float64{
		utils.MetaASR: 66.66667,
		utils.MetaACD: 60,
		utils.MetaTCC: 135.1,
		utils.MetaTCD: 180,
		utils.MetaACC: 45.03333,
		utils.MetaPDD: -1.0,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     180000000000,
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: 60000000000,
	}
	var floatMetrics2 map[string]float64
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}}, &floatMetrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFloatMetrics2, floatMetrics2) {
		t.Errorf("expecting: %+v, received reply: %+v", expectedFloatMetrics2, floatMetrics2)
	}

	if err := stsV1Rpc.Call(utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "Stats1"}}, &floatMetrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFloatMetrics2, floatMetrics2) {
		t.Errorf("expecting: %+v, received reply: %+v", expectedFloatMetrics2, floatMetrics2)
	}

}

func testV1STSGetStatsAfterRestart(t *testing.T) {
	// in case of internal we skip this test
	if stsV1ConfDIR == "tutinternal" {
		t.SkipNow()
	}
	// time.Sleep(time.Second)
	if _, err := engine.StopStartEngine(stsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	stsV1Rpc, err = newRPCClient(stsV1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}

	//get stats metrics after restart
	expectedMetrics2 := map[string]string{
		utils.MetaASR: "66.66667%",
		utils.MetaACD: "1m0s",
		utils.MetaACC: "45.03333",
		utils.MetaTCD: "3m0s",
		utils.MetaTCC: "135.1",
		utils.MetaPDD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage:     "180000000000",
		utils.MetaAverage + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage: "60000000000",
	}
	var metrics2 map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}}, &metrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics2, metrics2) {
		t.Errorf("After restat expecting: %+v, received reply: %s", expectedMetrics2, metrics2)
	}
}

func testV1STSSetStatQueueProfile(t *testing.T) {
	var result string
	var reply *engine.StatQueueProfile
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"*wrong:inline"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}

	expErr := "SERVER_ERROR: broken reference to filter: *wrong:inline for item with ID: cgrates.org:TEST_PROFILE1"
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatal(err)
	}
	statConfig.FilterIDs = []string{"FLTR_1"}
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV1STSGetStatQueueProfileIDs(t *testing.T) {
	expected := []string{"Stats1", "TEST_PROFILE1"}
	var result []string
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1STSUpdateStatQueueProfile(t *testing.T) {
	var result string
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig.FilterIDs = []string{"FLTR_1", "FLTR_2"}
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV1STSRemoveStatQueueProfile(t *testing.T) {
	var resp string
	if err := stsV1Rpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testV1STSProcessMetricsWithFilter(t *testing.T) {
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomStatProfile",
			FilterIDs: []string{"*string:~*req.DistinctVal:RandomVal"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID:  utils.MetaACD,
					FilterIDs: []string{"*gt:~*req.Usage:10s"},
				},
				{
					MetricID:  utils.MetaTCD,
					FilterIDs: []string{"*gt:~*req.Usage:5s"},
				},
				{
					MetricID:  utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "CustomValue",
					FilterIDs: []string{"*exists:~*req.CustomValue:", "*gte:~*req.CustomValue:10.0"},
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	//set the custom statProfile
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "CustomStatProfile"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
	//verify metrics
	expectedIDs := []string{"CustomStatProfile"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaACD: utils.NotAvailable,
		utils.MetaTCD: utils.NotAvailable,
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "CustomValue": utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	//process event
	var reply2 []string
	expected := []string{"CustomStatProfile"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"DistinctVal": "RandomVal",
				utils.Usage:   6 * time.Second,
				"CustomValue": 7.0,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	expectedMetrics = map[string]string{
		utils.MetaACD: utils.NotAvailable,
		utils.MetaTCD: "6s",
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "CustomValue": utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	//second process
	args = engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				"DistinctVal": "RandomVal",
				utils.Usage:   12 * time.Second,
				"CustomValue": 10.0,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	expectedMetrics = map[string]string{
		utils.MetaACD: "12s",
		utils.MetaTCD: "18s",
		utils.MetaSum + utils.HashtagSep + utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "CustomValue": "10",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSProcessStaticMetrics(t *testing.T) {
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "StaticStatQueue",
			FilterIDs: []string{"*string:~*req.StaticMetrics:StaticMetrics"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "1",
				},
				{
					MetricID: utils.MetaAverage + utils.HashtagSep + "2",
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	//set the custom statProfile
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "StaticStatQueue"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
	//verify metrics
	expectedIDs := []string{"StaticStatQueue"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + "1":     utils.NotAvailable,
		utils.MetaAverage + utils.HashtagSep + "2": utils.NotAvailable,
	}
	//process event
	var reply2 []string
	expected := []string{"StaticStatQueue"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"StaticMetrics": "StaticMetrics",
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "1":     "1",
		utils.MetaAverage + utils.HashtagSep + "2": "2",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	//second process
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "1":     "2",
		utils.MetaAverage + utils.HashtagSep + "2": "2",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}

	if err := stsV1Rpc.Call(utils.StatSv1ResetStatQueue,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("expecting: %+v, received reply: %s", utils.OK, result)
	}
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "1":     utils.NotAvailable,
		utils.MetaAverage + utils.HashtagSep + "2": utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSStatsPing(t *testing.T) {
	var resp string
	if err := stsV1Rpc.Call(utils.StatSv1Ping, new(utils.CGREvent), &resp); err != nil {
		t.Error(err)
	} else if resp != utils.Pong {
		t.Error("Unexpected reply returned", resp)
	}
}

func testV1STSProcessStatWithThreshold(t *testing.T) {
	stTh := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "StatWithThreshold",
			FilterIDs: []string{"*string:~*req.CustomEvent:CustomEvent"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "2",
				},
			},
			Stored:   true,
			Weight:   20,
			MinItems: 1,
		},
	}
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, stTh, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	thSts := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_Stat",
			FilterIDs: []string{"*string:~*req.EventType:StatUpdate",
				"*string:~*req.StatID:StatWithThreshold", "*exists:~*req.*tcd:", "*gte:~*req.*tcd:1s"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Weight:    20.0,
			ActionIDs: []string{"LOG_WARNING"},
			Async:     true,
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetThresholdProfile, thSts, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//process event
	var reply2 []string
	expected := []string{"StatWithThreshold"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"CustomEvent": "CustomEvent",
				utils.Usage:   45 * time.Second,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_Stat", Hits: 1}
	if err := stsV1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Stat"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testV1STSProcessCDRStat(t *testing.T) {
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "StatForCDR",
			FilterIDs: []string{"*string:~*req.OriginID:dsafdsaf"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "~*req.CostDetails.Usage",
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       50,
			MinItems:     1,
		},
	}
	//set the custom statProfile
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "StatForCDR"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
	//verify metrics
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + "1": utils.NotAvailable,
	}
	//process event
	cc := &engine.CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        1.01,
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "AccountFromAccountSummary",
			BalanceSummaries: []*engine.BalanceSummary{
				{
					UUID:  "f9be602747f4",
					ID:    "monetary",
					Type:  utils.MetaMonetary,
					Value: 0.5,
				},
				{
					UUID:  "2e02510ab90a",
					ID:    "voice",
					Type:  utils.MetaVoice,
					Value: 10,
				},
			},
		},
	}

	cdr := &engine.CDR{
		CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
		OrderID:     123,
		ToR:         utils.MetaVoice,
		OriginID:    "dsafdsaf",
		OriginHost:  "192.168.1.1",
		Source:      utils.UnitTest,
		RequestType: utils.MetaRated,
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1002",
		Subject:     "1001",
		Destination: "+4986517174963",
		SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
		AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
		RunID:       utils.MetaDefault,
		Usage:       10 * time.Second,
		ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		Cost:        1.01,
		CostDetails: engine.NewEventCostFromCallCost(cc, "TestCDRTestCDRAsMapStringIface2", utils.MetaDefault),
	}
	cdr.CostDetails.Compute()
	cdr.CostDetails.Usage = utils.DurationPointer(10 * time.Second)

	var reply2 []string
	expected := []string{"StatForCDR"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: cdr.AsCGREvent(),
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*req.CostDetails.Usage": "10000000000",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "StatForCDR"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	//second process
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*req.CostDetails.Usage": "20000000000",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "StatForCDR"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSOverWriteStats(t *testing.T) {
	initStat := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "InitStat",
			FilterIDs: []string{"*string:~*req.OriginID:dsafdsaf"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "1",
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	//set the custom statProfile
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, initStat, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "InitStat"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(initStat.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(initStat.StatQueueProfile), utils.ToJSON(reply))
	}

	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + "1": utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "InitStat"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	// set the new profile with other metric and make sure the statQueue is updated
	initStat2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "InitStat",
			FilterIDs: []string{"*string:~*req.OriginID:dsafdsaf"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "~*req.Test",
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, initStat2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var metrics2 map[string]string
	expectedMetrics2 := map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*req.Test": utils.NotAvailable,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "InitStat"}}, &metrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics2, metrics2) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics2, metrics2)
	}
}

func testV1STSProcessStatWithThreshold2(t *testing.T) {
	stTh := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "StatWithThreshold2",
			FilterIDs: []string{"*string:~*req.CustomEvent2:CustomEvent2"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "2",
				},
			},
			Stored:   true,
			Weight:   20,
			MinItems: 1,
		},
	}
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, stTh, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	thSts := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_Stat2",
			FilterIDs: []string{"*string:~*req.EventType:StatUpdate",
				"*string:~*req.StatID:StatWithThreshold2", "*exists:~*req.*sum#2:", "*gt:~*req.*sum#2:1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Weight:    20.0,
			ActionIDs: []string{"LOG_WARNING"},
			Async:     true,
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetThresholdProfile, thSts, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//process event
	var reply2 []string
	expected := []string{"StatWithThreshold2"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				"CustomEvent2": "CustomEvent2",
				utils.Usage:    45 * time.Second,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "THD_Stat2", Hits: 1}
	if err := stsV1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "THD_Stat"}}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testV1STSStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// Run benchmarks with: <go test -tags=integration -run=TestSTSV1 -bench=.>
// BenchmarkStatSV1SetEvent         	    5000	    263437 ns/op
func BenchmarkSTSV1SetEvent(b *testing.B) {
	if _, err := engine.StopStartEngine(stsV1CfgPath, 1000); err != nil {
		b.Fatal(err)
	}
	b.StopTimer()
	var err error
	stsV1Rpc, err = newRPCClient(stsV1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		b.Fatal("Could not connect to rater: ", err.Error())
	}
	var reply string
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &evs[rand.Intn(len(evs))],
			&reply); err != nil {
			b.Error(err)
		} else if reply != utils.OK {
			b.Errorf("received reply: %s", reply)
		}
	}
}

// BenchmarkStatSV1GetQueueStringMetrics 	   20000	     94607 ns/op
func BenchmarkSTSV1GetQueueStringMetrics(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var metrics map[string]string
		if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
			&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STATS_1"}},
			&metrics); err != nil {
			b.Error(err)
		}
	}
}

func testV1STSGetStatQueueProfileWithoutTenant(t *testing.T) {
	statConfig := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			ID:        "TEST_PROFILE10",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig.Tenant = "cgrates.org"
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{ID: "TEST_PROFILE10"},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, statConfig.StatQueueProfile) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV1STSRemStatQueueProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := stsV1Rpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "TEST_PROFILE10"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{ID: "TEST_PROFILE10"},
		&result); err == nil || utils.ErrNotFound.Error() != err.Error() {
		t.Error(err)
	}
}

func testV1STSV1GetQueueIDs(t *testing.T) {
	expected := []string{"StatWithThreshold", "Stats1", "StaticStatQueue", "CustomStatProfile"}
	sort.Strings(expected)
	var qIDs []string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueIDs,
		&utils.TenantWithAPIOpts{},
		&qIDs); err != nil {
		t.Error(err)
	} else {
		sort.Strings(qIDs)
		if !reflect.DeepEqual(qIDs, expected) {
			t.Errorf("Expected %+v \n ,received %+v", expected, qIDs)
		}
		if err := stsV1Rpc.Call(utils.StatSv1GetQueueIDs,
			&utils.TenantWithAPIOpts{Tenant: "cgrates.org"},
			&qIDs); err != nil {
			t.Error(err)
		} else {
			sort.Strings(qIDs)
			if !reflect.DeepEqual(qIDs, expected) {
				t.Errorf("Expected %+v \n ,received %+v", expected, qIDs)
			}
		}
	}
}

func testV1STSV1GetStatQueuesForEventWithoutTenant(t *testing.T) {
	var reply []string
	estats := []string{"Stats1"}
	if err := stsV1Rpc.Call(utils.StatSv1GetStatQueuesForEvent,
		&engine.StatsArgsProcessEvent{
			CGREvent: &utils.CGREvent{
				ID: "GetStats",
				Event: map[string]interface{}{
					utils.AccountField: "1002",
					utils.AnswerTime:   time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
					utils.Usage:        45 * time.Second,
					utils.RunID:        utils.MetaDefault,
					utils.Cost:         10.0,
					utils.Destination:  "1001",
				},
				APIOpts: map[string]interface{}{
					utils.OptsAPIKey: "stat12345",
				},
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(estats, reply) {
		t.Errorf("expecting: %+v, received reply: %v", estats, reply)
	}
}

func testV1STSSimulateAccountUpdate(t *testing.T) {
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "StatForAccountUpdate",
			FilterIDs: []string{
				"*string:~*opts.*eventType:AccountUpdate",
				"*string:~*asm.ID:testV1STSSimulateAccountUpdate",
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum + utils.HashtagSep + "~*asm.BalanceSummaries.HolidayBalance.Value",
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       50,
			MinItems:     1,
		},
	}
	//set the custom statProfile
	var result string
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "StatForAccountUpdate"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
	//verify metrics
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*asm.BalanceSummaries.HolidayBalance.Value": utils.NotAvailable,
	}

	var reply2 []string
	expected := []string{"StatForAccountUpdate"}

	attrSetBalance := &utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "testV1STSSimulateAccountUpdate",
		BalanceType: "*monetary",
		Value:       1.5,
		Balance: map[string]interface{}{
			utils.ID: "HolidayBalance",
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetBalance, attrSetBalance, &result); err != nil {
		t.Error("Got error on APIerSv1.SetBalance: ", err.Error())
	} else if result != utils.OK {
		t.Errorf("Calling APIerSv1.SetBalance received: %s", result)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "testV1STSSimulateAccountUpdate",
	}
	if err := stsV1Rpc.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	}

	acntUpdateEv := &engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{ // hitting TH_ACNT_UPDATE_EV
			Tenant: "cgrates.org",
			ID:     "SIMULATE_ACNT_UPDATE_EV",
			Event:  acnt.AsAccountSummary().AsMapInterface(),
			APIOpts: map[string]interface{}{
				utils.MetaEventType: utils.AccountUpdate,
			},
		},
	}

	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &acntUpdateEv, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*asm.BalanceSummaries.HolidayBalance.Value": "1.5",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "StatForAccountUpdate"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	//second process
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &acntUpdateEv, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	expectedMetrics = map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*asm.BalanceSummaries.HolidayBalance.Value": "3",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "StatForAccountUpdate"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSGetStatQueueWithoutExpired(t *testing.T) {
	var result string
	var reply *engine.StatQueueProfile
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "Sq1Nanao",
			FilterIDs:   []string{"*string:~*req.StatQ:Sq1Nanao"},
			QueueLength: 10,
			TTL:         1,
			Metrics: []*engine.MetricWithFilters{{
				MetricID: utils.MetaTCD,
			}},
			Blocker:  true,
			Stored:   true,
			Weight:   200,
			MinItems: 1,
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sq1Nanao"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}

	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaTCD: utils.NotAvailable,
	}
	//process event
	var reply2 []string
	expected := []string{"Sq1Nanao"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1012",
			Event: map[string]interface{}{
				"StatQ":     "Sq1Nanao",
				utils.Usage: 10,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Sq1Nanao"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSGetStatQueueWithoutStored(t *testing.T) {
	var result string
	var reply *engine.StatQueueProfile
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "Sq1NotStored",
			FilterIDs:   []string{"*string:~*req.StatQ:Sq1NotStored"},
			QueueLength: 10,
			TTL:         time.Second,
			Metrics: []*engine.MetricWithFilters{{
				MetricID: utils.MetaTCD,
			}},
			Blocker:  true,
			Weight:   200,
			MinItems: 1,
		},
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Sq1NotStored"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}

	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaTCD: "10s",
	}
	//process event
	var reply2 []string
	expected := []string{"Sq1NotStored"}
	args := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Sq1NotStored",
			Event: map[string]interface{}{
				"StatQ":     "Sq1NotStored",
				utils.Usage: 10 * time.Second,
			},
		},
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Sq1NotStored"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	if err := stsV1Rpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Sq1NotStored"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testV1STSCheckMetricsAfterRestart(t *testing.T) {
	var metrics map[string]string

	expectedMetrics := map[string]string{
		utils.MetaSum + utils.HashtagSep + "~*asm.BalanceSummaries.HolidayBalance.Value": "3",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "StatForAccountUpdate"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
	expectedMetrics = map[string]string{
		utils.MetaTCD: utils.NotAvailable,
	}
	metrics = map[string]string{}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Sq1NotStored"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}
