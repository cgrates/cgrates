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
	stsV1CfgPath string
	stsV1Cfg     *config.CGRConfig
	stsV1Rpc     *rpc.Client
	statConfig   *StatQueueWithCache
	stsV1ConfDIR string //run tests for specific configuration
)

var evs = []*utils.CGREvent{
	{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.Account:    "1001",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(135 * time.Second),
			utils.COST:       123.0}},
	{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(45 * time.Second)}},
	{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.Account:   "1002",
			utils.SetupTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:     0}},
}

func init() {
	rand.Seed(time.Now().UnixNano()) // used in benchmarks
}

var sTestsStatSV1 = []func(t *testing.T){
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
	testV1STSStopEngine,
}

//Test start here
func TestSTSV1ITMySQL(t *testing.T) {
	stsV1ConfDIR = "tutmysql"
	for _, stest := range sTestsStatSV1 {
		t.Run(stsV1ConfDIR, stest)
	}
}

func TestSTSV1ITMongo(t *testing.T) {
	stsV1ConfDIR = "tutmongo"
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
	stsV1Rpc, err = jsonrpc.Dial("tcp", stsV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1STSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := stsV1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1STSGetStats(t *testing.T) {
	var reply []string
	expectedIDs := []string{"Stats1"}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueIDs, &utils.TenantArg{"cgrates.org"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDs, reply) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, reply)
	}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaASR: utils.NOT_AVAILABLE,
		utils.MetaACD: utils.NOT_AVAILABLE,
		utils.MetaTCC: utils.NOT_AVAILABLE,
		utils.MetaTCD: utils.NOT_AVAILABLE,
		utils.MetaACC: utils.NOT_AVAILABLE,
		utils.MetaPDD: utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaSum, utils.Usage):     utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaAverage, utils.Usage): utils.NOT_AVAILABLE,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
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
				utils.Account:    "1001",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(135 * time.Second),
				utils.COST:       123.0,
				utils.PDD:        time.Duration(12 * time.Second)}}}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	//process with one event (should be N/A becaus MinItems is 2)
	expectedMetrics := map[string]string{
		utils.MetaASR: utils.NOT_AVAILABLE,
		utils.MetaACD: utils.NOT_AVAILABLE,
		utils.MetaTCC: utils.NOT_AVAILABLE,
		utils.MetaTCD: utils.NOT_AVAILABLE,
		utils.MetaACC: utils.NOT_AVAILABLE,
		utils.MetaPDD: utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaSum, utils.Usage):     utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaAverage, utils.Usage): utils.NOT_AVAILABLE,
	}
	var metrics map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &metrics); err != nil {
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
		utils.StatsJoin(utils.MetaSum, utils.Usage):     -1.0,
		utils.StatsJoin(utils.MetaAverage, utils.Usage): -1.0,
	}
	var floatMetrics map[string]float64
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueFloatMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &floatMetrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFloatMetrics, floatMetrics) {
		t.Errorf("expecting: %+v, received reply: %+v", expectedFloatMetrics, floatMetrics)
	}

	args2 := engine.StatsArgsProcessEvent{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.Account:    "1002",
				utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:      time.Duration(45 * time.Second),
				utils.Cost:       12.1}}}
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
				utils.Account:   "1002",
				utils.SetupTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				utils.Usage:     0,
				utils.Cost:      0}}}
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
		utils.MetaPDD: utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaSum, utils.Usage):     "180000000000",
		utils.StatsJoin(utils.MetaAverage, utils.Usage): "60000000000",
	}
	var metrics2 map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics, &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &metrics2); err != nil {
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
		utils.StatsJoin(utils.MetaSum, utils.Usage):     180000000000,
		utils.StatsJoin(utils.MetaAverage, utils.Usage): 60000000000,
	}
	var floatMetrics2 map[string]float64
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueFloatMetrics, &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &floatMetrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedFloatMetrics2, floatMetrics2) {
		t.Errorf("expecting: %+v, received reply: %+v", expectedFloatMetrics2, floatMetrics2)
	}

}

func testV1STSGetStatsAfterRestart(t *testing.T) {
	time.Sleep(time.Second)
	if _, err := engine.StopStartEngine(stsV1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	stsV1Rpc, err = jsonrpc.Dial("tcp", stsV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
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
		utils.MetaPDD: utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaSum, utils.Usage):     "180000000000",
		utils.StatsJoin(utils.MetaAverage, utils.Usage): "60000000000",
	}
	var metrics2 map[string]string
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics, &utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &metrics2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics2, metrics2) {
		t.Errorf("After restat expecting: %+v, received reply: %s", expectedMetrics2, metrics2)
	}
}

func testV1STSSetStatQueueProfile(t *testing.T) {
	var reply *engine.StatQueueProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					FieldName: "~Account",
					Type:      "*string",
					Values:    []string{"1001"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := stsV1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				&engine.MetricWithFilters{
					MetricID: "*acd",
				},
				&engine.MetricWithFilters{
					MetricID: "*tcd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}

	if err := stsV1Rpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := stsV1Rpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV1STSGetStatQueueProfileIDs(t *testing.T) {
	expected := []string{"Stats1", "TEST_PROFILE1"}
	var result []string
	if err := stsV1Rpc.Call(utils.ApierV1GetStatQueueProfileIDs, utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: "cgrates.org"}}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1STSUpdateStatQueueProfile(t *testing.T) {
	var result string
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{
				{
					FieldName: "~Account",
					Type:      "*string",
					Values:    []string{"1001"},
				},
			},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	if err := stsV1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig.FilterIDs = []string{"FLTR_1", "FLTR_2"}
	if err := stsV1Rpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
}

func testV1STSRemoveStatQueueProfile(t *testing.T) {
	var resp string
	if err := stsV1Rpc.Call("ApierV1.RemoveStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.StatQueueProfile
	if err := stsV1Rpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := stsV1Rpc.Call("ApierV1.RemoveStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testV1STSProcessMetricsWithFilter(t *testing.T) {
	statConfig = &StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    "cgrates.org",
			ID:        "CustomStatProfile",
			FilterIDs: []string{"*string:~DistinctVal:RandomVal"}, //custom filter for event
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 100,
			TTL:         time.Duration(1) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				&engine.MetricWithFilters{
					MetricID:  "*acd",
					FilterIDs: []string{"*rsr::~Usage{*duration}(>10s)"},
				},
				&engine.MetricWithFilters{
					MetricID:  "*tcd",
					FilterIDs: []string{"*gt:~Usage:5s"},
				},
				&engine.MetricWithFilters{
					MetricID:  "*sum#CustomValue",
					FilterIDs: []string{"*exists:~CustomValue:", "*gte:~CustomValue:10.0"},
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
	if err := stsV1Rpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//verify it
	var reply *engine.StatQueueProfile
	if err := stsV1Rpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "CustomStatProfile"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(statConfig.StatQueueProfile), utils.ToJSON(reply))
	}
	//verify metrics
	expectedIDs := []string{"CustomStatProfile"}
	var metrics map[string]string
	expectedMetrics := map[string]string{
		utils.MetaACD: utils.NOT_AVAILABLE,
		utils.MetaTCD: utils.NOT_AVAILABLE,
		utils.StatsJoin(utils.MetaSum, "CustomValue"): utils.NOT_AVAILABLE,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &metrics); err != nil {
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
				utils.Usage:   time.Duration(6 * time.Second),
				"CustomValue": 7.0}}}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}
	//verify metrics after first process
	expectedMetrics = map[string]string{
		utils.MetaACD: utils.NOT_AVAILABLE,
		utils.MetaTCD: "6s",
		utils.StatsJoin(utils.MetaSum, "CustomValue"): utils.NOT_AVAILABLE,
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &metrics); err != nil {
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
				utils.Usage:   time.Duration(12 * time.Second),
				"CustomValue": 10.0}}}
	if err := stsV1Rpc.Call(utils.StatSv1ProcessEvent, &args, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply2, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply2)
	}

	expectedMetrics = map[string]string{
		utils.MetaACD: "12s",
		utils.MetaTCD: "18s",
		utils.StatsJoin(utils.MetaSum, "CustomValue"): "10",
	}
	if err := stsV1Rpc.Call(utils.StatSv1GetQueueStringMetrics,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &metrics); err != nil {
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
	stsV1Rpc, err = jsonrpc.Dial("tcp", stsV1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
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
			&utils.TenantID{Tenant: "cgrates.org", ID: "STATS_1"},
			&metrics); err != nil {
			b.Error(err)
		}
	}
}
