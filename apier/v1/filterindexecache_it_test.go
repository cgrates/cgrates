// +build integration2

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FIdxCaTNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/
package v1

import (
	"fmt"
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
	tFIdxCaRpc *rpc.Client
)

var sTestsFilterIndexesSV1Ca = []func(t *testing.T){
	testFlush,
	testV1FIdxCaLoadConfig,
	testV1FIdxCadxInitDataDb,
	testV1FIdxCaResetStorDb,
	testV1FIdxCaStartEngine,
	testV1FIdxCaRpcConn,

	testV1FIdxCaGetThresholdsWithNotFound,
	testV1FIdxCaSetThresholdProfile,
	testV1FIdxCaFromFolder,
	testV1FIdxCaGetThresholds,
	testV1FIdxCaUpdateThresholdProfile,
	testV1FIdxCaRemoveThresholdProfile,

	// To be implemented after threshold one works
	/*
		testFlush,
		testV1FIdxCaGetStatQueuesWithNotFound,
		testV1FIdxCaSetStatQueueProfile,
		testV1FIdxCaGetStatQueues,
		testV1FIdxCaFromFolder,
		testV1FIdxCaUpdateStatQueueProfile,
		testV1FIdxCaRemoveStatQueueProfile,
	*/
}

func TestFIdxCaV1ITMySQLConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err = engine.NewRedisStorage(fmt.Sprintf("%s:%s", cfg.DataDbHost, cfg.DataDbPort), 10,
		cfg.DataDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)

	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
}

// Test start here
func TestFIdxCaV1ITMySQL(t *testing.T) {
	onStor = engine.NewDataManager(rdsITdb)
	tSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsFilterIndexesSV1Ca {
		t.Run(tSv1ConfDIR, stest)
	}
}

func TestFIdxCaV1ITMongoConnect(t *testing.T) {
	cdrsMongoCfgPath := path.Join(*dataDir, "conf", "samples", "tutmongo")
	mgoITCfg, err := config.NewCGRConfigFromFolder(cdrsMongoCfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if mgoITdb, err = engine.NewMongoStorage(mgoITCfg.DataDbHost, mgoITCfg.DataDbPort,
		mgoITCfg.DataDbName, mgoITCfg.DataDbUser, mgoITCfg.DataDbPass,
		utils.DataDB, nil, mgoITCfg.CacheCfg(), mgoITCfg.LoadHistorySize); err != nil {
		t.Fatal(err)
	}
}

func TestFIdxCaV1ITMongo(t *testing.T) {
	onStor = engine.NewDataManager(mgoITdb)
	tSv1ConfDIR = "tutmongo"
	time.Sleep(time.Duration(2 * time.Second)) // give time for engine to start
	for _, stest := range sTestsFilterIndexesSV1Ca {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxCaLoadConfig(t *testing.T) {
	var err error
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	if tSv1Cfg, err = config.NewCGRConfigFromFolder(tSv1CfgPath); err != nil {
		t.Error(err)
	}
	switch tSv1ConfDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		thdsDelay = 4000
	default:
		thdsDelay = 1000
	}
}

func testV1FIdxCadxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxCaResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxCaStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, thdsDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxCaRpcConn(t *testing.T) {
	var err error
	tFIdxCaRpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxCaFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tFIdxCaRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

//ThresholdProfile
func testV1FIdxCaGetThresholdsWithNotFound(t *testing.T) {
	var tIDs []string
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1GetThresholdIDs, "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	}
	_, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}
func testV1FIdxCaSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}

	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"TestFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Recurrent: true,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     true,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
	expectedIDX := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDX, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, matchedindexes)
	}
	fldNameVal2 := map[string]string{"TEST_PROFILE1": ""}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1001": true}}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, indexes)
	}
}

func testV1FIdxCaGetThresholds(t *testing.T) {
	// var resp string
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED", "THD_STATS_3", "THD_CDRS_1"}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1GetThresholdIDs, "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: expectedIDs[0]}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
	idx := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(idx, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", idx, utils.ToJSON(matchedindexes))
	}
	idx2 := map[string]utils.StringMap{"THD_ACNT_BALANCE_1": {"Account:1001": true, "Account:1002": true, "EventType:BalanceUpdate": true}}
	fldNameVal2 := map[string]string{"THD_ACNT_BALANCE_1": ""}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(idx2, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", idx2, utils.ToJSON(indexes))
	}
}

func testV1FIdxCaUpdateThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}

	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"TestFilter2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Recurrent: true,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     true,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
	expectedIDX := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1002")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDX, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, matchedindexes)
	}
	fldNameVal2 := map[string]string{"TEST_PROFILE1": ""}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1002": true}}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxCaRemoveThresholdProfile(t *testing.T) {
	var resp string

	expectedIDX := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDX, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, matchedindexes)
	}

	expectedIDX = utils.StringMap{"THD_ACNT_BALANCE_1": true}
	matchedindexes, err = onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1002")
	if err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedIDX, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, matchedindexes)
	}

	if err := tFIdxCaRpc.Call("ApierV1.RemThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxCaRpc.Call("ApierV1.RemThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	matchedindexes, err = onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1002")
	if err != nil && err != utils.ErrNotFound {
		t.Error(err)
	} else if matchedindexes != nil {
		t.Errorf("Expecting: %+v, received: %+v", nil, matchedindexes)
	}

	matchedindexes, err = onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", false),
		"Account", "1002")
	if err != nil && err != utils.ErrNotFound {
		t.Error(err)
	} else if matchedindexes != nil {
		t.Errorf("Expecting: %+v, received: %+v", nil, matchedindexes)
	}
	fldNameVals2 := map[string]string{"THD_ACNT_BALANCE_1": "", "TEST_PROFILE1": ""}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, "cgrates.org", true),
		fldNameVals2); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

/*
//StatQueue
func testV1FIdxCaGetStatQueuesWithNotFound(t *testing.T) {
	var reply *engine.StatQueueProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxCaSetStatQueueProfile(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.StatQueueProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_1",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "MetricValue",
				Parameters: "",
			},
			&utils.MetricWithParams{
				MetricID:   "MetricValueTwo",
				Parameters: "",
			},
		},
		Thresholds: []string{"Val1", "Val2"},
		Blocker:    true,
		Stored:     true,
		Weight:     20,
		MinItems:   1,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig, reply)
	}
	expectedIDX := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(matchedindexes))
	}
	fldNameVal2 := map[string]string{"TEST_PROFILE1": ""}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1001": true}}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxCaGetStatQueues(t *testing.T) {
	tenant := "cgrates.org"
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_1",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var td engine.StatQueueProfile
	eTd := engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
		QueueLength: 10,
		TTL:         10000000000,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "MetricValue",
				Parameters: ""},
			&utils.MetricWithParams{MetricID: "MetricValueTwo",
				Parameters: ""}},
		Thresholds: []string{"Val1", "Val2"},
		Blocker:    true,
		Stored:     true,
		Weight:     20,
		MinItems:   1,
	}

	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", utils.ToJSON(eTd), utils.ToJSON(td))
	}
	idx := utils.StringMap{"TEST_PROFILE1": true}
	matchedindexes, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(idx, matchedindexes) {
		t.Errorf("Expecting: %+v, received: %+v", idx, utils.ToJSON(matchedindexes))
	}

	idx2 := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1001": true}}
	fldNameVal2 := map[string]string{"TEST_PROFILE1": ""}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(idx2, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", idx2, utils.ToJSON(indexes))
	}
}

func testV1FIdxCaUpdateStatQueueProfile(t *testing.T) {

	tenant := "cgrates.org"
	var reply *engine.StatQueueProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_1",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "MetricValue",
				Parameters: "",
			},
			&utils.MetricWithParams{
				MetricID:   "MetricValueTwo",
				Parameters: "",
			},
		},
		Thresholds: []string{"Val1", "Val2"},
		Blocker:    true,
		Stored:     true,
		Weight:     20,
		MinItems:   1,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig, reply)
	}
	// idx := utils.StringMap{"TEST_PROFILE1": true}
	// matchedindexes
	_, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != nil {
		t.Error(err)
	}
	fldNameVal := map[string]string{"Account": "1002"}
	expectedIDX := map[string]utils.StringMap{"Account:1002": {"Stats1": true, "TEST_PROFILE1": true}}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		fldNameVal); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}

	fldNameVal2 := map[string]string{"TEST_PROFILE1": ""}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1001": true, "Account:1002": true}}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", true),
		fldNameVal2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxCaRemoveStatQueueProfile(t *testing.T) {
	var resp string
	fldNameVal := map[string]string{"Account": "1001"}
	expectedIDX := map[string]utils.StringMap{"Account:1001": {"Stats1": true, "TEST_PROFILE1": true}}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		fldNameVal); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}

	fldNameVal3 := map[string]string{"Account": "1002"}
	expectedIDX3 := map[string]utils.StringMap{"Account:1002": {"Stats1": true, "TEST_PROFILE1": true}}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		fldNameVal3); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX3, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX3, utils.ToJSON(indexes))
	}

	if err := tFIdxCaRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxCaRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// idx := utils.StringMap{"TEST_PROFILE1": true}
	// matchedindexes
	_, err := onStor.MatchFilterIndex(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		"Account", "1001")
	if err != utils.ErrNotFound {
		t.Error(err)
	}
	fldNameVal1 := map[string]string{"Account": "1001"}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		fldNameVal1); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if len(indexes) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(indexes))
	}
	fldNameVals := map[string]string{"Account": "1002"}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", false),
		fldNameVals); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if len(indexes) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(indexes))
	}
	fldNameVals2 := map[string]string{"THD_ACNT_BALANCE_1": "", "TEST_PROFILE1": ""}
	if indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, "cgrates.org", true),
		fldNameVals2); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if len(indexes) != 0 {
		t.Errorf("Expecting: %+v, received: %+v", 0, len(indexes))
	}
}
*/
