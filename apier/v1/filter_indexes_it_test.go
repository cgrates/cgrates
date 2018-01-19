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
MERCHANTABILITY or FIdxTNESS FOR A PARTICULAR PURPOSE.  See the
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
	tFIdxRpc *rpc.Client
	rdsITdb  *engine.RedisStorage
	mgoITdb  *engine.MongoStorage
	onStor   *engine.DataManager
	err      error
	indexes  map[string]utils.StringMap
)

var sTestsFilterIndexesSV1 = []func(t *testing.T){
	testFlush,
	testV1FIdxLoadConfig,
	testV1FIdxdxInitDataDb,
	testV1FIdxResetStorDb,
	testV1FIdxStartEngine,
	testV1FIdxRpcConn,

	testV1FIdxSetThresholdProfile,
	testV1FIdxComputeThresholdsIndexes,
	testV1FIdxSetSecondThresholdProfile,
	testV1FIdxSecondComputeThresholdsIndexes,
	testV1FIdxRemoveThresholdProfile,

	testV1FIdxSetStatQueueProfileIndexes,
	testV1FIdxComputeStatQueueProfileIndexes,
	testV1FIdxSetSecondStatQueueProfileIndexes,
	testV1FIdxSecondComputeStatQueueProfileIndexes,
	testV1FIdxRemoveStatQueueProfile,

	testV1FIdxSetResourceProfileIndexes,
	testV1FIdxComputeResourceProfileIndexes,
	testV1FIdxSetSecondResourceProfileIndexes,
	testV1FIdxSecondComputeResourceProfileIndexes,
	testV1FIdxRemoveResourceProfile,

	testV1FIdxSetSupplierProfileIndexes,
	testV1FIdxComputeSupplierProfileIndexes,
	testV1FIdxSetSecondSupplierProfileIndexes,
	testV1FIdxSecondComputeSupplierProfileIndexes,
	testV1FIdxRemoveSupplierProfile,

	testV1FIdxSetAttributeProfileIndexes,
	testV1FIdxComputeAttributeProfileIndexes,
	testV1FIdxSetSecondAttributeProfileIndexes,
	testV1FIdxSecondComputeAttributeProfileIndexes,
	testV1FIdxRemoveAttributeProfile,

	testV1FIdxStopEngine,
}

func TestFIdxV1ITMySQLConnect(t *testing.T) {
	cfg, _ := config.NewDefaultCGRConfig()
	rdsITdb, err = engine.NewRedisStorage(fmt.Sprintf("%s:%s", cfg.DataDbHost, cfg.DataDbPort), 10,
		cfg.DataDbPass, cfg.DBDataEncoding, utils.REDIS_MAX_CONNS, nil, 1)

	if err != nil {
		t.Fatal("Could not connect to Redis", err.Error())
	}
}

// Test start here
func TestFIdxV1ITMySQL(t *testing.T) {
	onStor = engine.NewDataManager(rdsITdb)
	tSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsFilterIndexesSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func TestFIdxV1ITMongoConnect(t *testing.T) {
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

func TestFIdxV1ITMongo(t *testing.T) {
	onStor = engine.NewDataManager(mgoITdb)
	tSv1ConfDIR = "tutmongo"
	time.Sleep(time.Duration(2 * time.Second)) // give time for engine to start
	for _, stest := range sTestsFilterIndexesSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxLoadConfig(t *testing.T) {
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

func testV1FIdxdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testFlush(t *testing.T) {
	onStor.DataDB().Flush("")
	if err := engine.SetDBVersions(onStor.DataDB()); err != nil {
		t.Error("Error  ", err.Error())
	}
}

func testV1FIdxStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, thdsDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxRpcConn(t *testing.T) {
	var err error
	tFIdxRpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

//ThresholdProfile
func testV1FIdxSetThresholdProfile(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.ThresholdProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "TestFilter",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}

	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    tenant,
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"TestFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Recurrent: true,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     true,
	}
	if err := tFIdxRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxComputeThresholdsIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: nil,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE1": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondThresholdProfile(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.ThresholdProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "TestFilter2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}

	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    tenant,
		ID:        "TEST_PROFILE2",
		FilterIDs: []string{"TestFilter2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Recurrent: true,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     true,
	}
	if err := tFIdxRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func testV1FIdxSecondComputeThresholdsIndexes(t *testing.T) {
	tenant := "cgrates.org"
	thid := []string{"TEST_PROFILE2"}
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &thid,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE2": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveThresholdProfile(t *testing.T) {
	var resp string
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: nil,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), engine.MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

//StatQueueProfile
func testV1FIdxSetStatQueueProfileIndexes(t *testing.T) {
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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "*sum",
				Parameters: "",
			},
			&utils.MetricWithParams{
				MetricID:   "*acd",
				Parameters: "",
			},
		},
		Thresholds: []string{"Val1", "Val2"},
		Blocker:    true,
		Stored:     true,
		Weight:     20,
		MinItems:   1,
	}
	if err := tFIdxRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxComputeStatQueueProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      nil,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE1": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondStatQueueProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.StatQueueProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE2",
		FilterIDs: []string{"FLTR_2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			&utils.MetricWithParams{
				MetricID:   "*sum",
				Parameters: "",
			},
			&utils.MetricWithParams{
				MetricID:   "*acd",
				Parameters: "",
			},
		},
		Thresholds: []string{"Val1", "Val2"},
		Blocker:    true,
		Stored:     true,
		Weight:     20,
		MinItems:   1,
	}
	if err := tFIdxRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeStatQueueProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	stid := []string{"TEST_PROFILE2"}
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &stid,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE2": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.StatQueueProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveStatQueueProfile(t *testing.T) {
	var resp string
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      nil,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

//ResourceProfile
func testV1FIdxSetResourceProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.ResourceProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_RES_RCFG1",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rlsConfig = &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RCFG1",
		FilterIDs: []string{"FLTR_RES_RCFG1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(10) * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		Thresholds:        []string{"Val1", "Val2"},
	}
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxComputeResourceProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  nil,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"RCFG1": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"RCFG1": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondResourceProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.StatQueueProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "RCFG2"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rlsConfig = &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RCFG2",
		FilterIDs: []string{"FLTR_2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(10) * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		Thresholds:        []string{"Val1", "Val2"},
	}
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeResourceProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	rsid := []string{"RCFG2"}
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &rsid,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"RCFG2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"RCFG2": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveResourceProfile(t *testing.T) {
	var resp string
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  nil,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemResourceProfile",
		&utils.TenantID{Tenant: tenant, ID: "RCFG1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemResourceProfile",
		&utils.TenantID{Tenant: tenant, ID: "RCFG2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "RCFG2"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, false), engine.MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ResourceProfilesPrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

//SupplierProfile
func testV1FIdxSetSupplierProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.SupplierProfile
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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &engine.SupplierProfile{
		Tenant:        "cgrates.org",
		ID:            "TEST_PROFILE1",
		FilterIDs:     []string{"FLTR_1"},
		Sorting:       "Sort1",
		SortingParams: []string{"Param1", "Param2"},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_1"},
				AccountIDs:    []string{"Acc"},
				ResourceIDs:   []string{"Res1", "ResGroup2"},
				StatIDs:       []string{"Stat1"},
				Weight:        20,
			},
		},
		Blocker: false,
		Weight:  10,
	}
	if err := tFIdxRpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxComputeSupplierProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  nil,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE1": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondSupplierProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.SupplierProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &engine.SupplierProfile{
		Tenant:        "cgrates.org",
		ID:            "TEST_PROFILE2",
		FilterIDs:     []string{"FLTR_2"},
		Sorting:       "Sort1",
		SortingParams: []string{"Param1", "Param2"},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_2"},
				AccountIDs:    []string{"Acc"},
				ResourceIDs:   []string{"Res1", "ResGroup2"},
				StatIDs:       []string{"Stat1"},
				Weight:        20,
			},
		},
		Blocker: false,
		Weight:  10,
	}
	if err := tFIdxRpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf, reply)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeSupplierProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	spid := []string{"TEST_PROFILE2"}
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &spid,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"TEST_PROFILE2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE2": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveSupplierProfile(t *testing.T) {
	var resp string
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &emptySlice,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  nil,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, false), engine.MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.SupplierProfilePrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

//AttributeProfile
func testV1FIdxSetAttributeProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.ExternalAttributeProfile
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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.ExternalAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{"*rating"},
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	if err := tFIdxRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf.FilterIDs, reply.FilterIDs) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.FilterIDs, reply.FilterIDs)
	} else if !reflect.DeepEqual(alsPrf.ActivationInterval, reply.ActivationInterval) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ActivationInterval, reply.ActivationInterval)
	} else if !reflect.DeepEqual(len(alsPrf.Attributes), len(reply.Attributes)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.Attributes), utils.ToJSON(reply.Attributes))
	} else if !reflect.DeepEqual(alsPrf.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ID, reply.ID)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxComputeAttributeProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: nil,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"ApierTest": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"ApierTest": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondAttributeProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	var reply *engine.ExternalAttributeProfile
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_2",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "Account",
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.ExternalAttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest2",
		Contexts:  []string{"*rating"},
		FilterIDs: []string{"FLTR_2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  "FL1",
				Initial:    "In1",
				Substitute: "Al1",
				Append:     true,
			},
		},
		Weight: 20,
	}
	if err := tFIdxRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf.FilterIDs, reply.FilterIDs) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.FilterIDs, reply.FilterIDs)
	} else if !reflect.DeepEqual(alsPrf.ActivationInterval, reply.ActivationInterval) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ActivationInterval, reply.ActivationInterval)
	} else if !reflect.DeepEqual(len(alsPrf.Attributes), len(reply.Attributes)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf.Attributes), utils.ToJSON(reply.Attributes))
	} else if !reflect.DeepEqual(alsPrf.ID, reply.ID) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf.ID, reply.ID)
	}
	if err = onStor.RemoveFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
		tenant, false)); err != nil {
		t.Error(err)
	}
	if err := onStor.RemoveFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix,
		tenant, true)); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, false), engine.MetaString,
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeAttributeProfileIndexes(t *testing.T) {
	tenant := "cgrates.org"
	apid := []string{"ApierTest2"}
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: &apid,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := map[string]utils.StringMap{"*string:Account:1001": {"ApierTest2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, false), engine.MetaString, nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"ApierTest2": {"*string:Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveAttributeProfile(t *testing.T) {
	var resp string
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:       "cgrates.org",
		ThresholdIDs: &emptySlice,
		AttributeIDs: nil,
		ResourceIDs:  &emptySlice,
		StatIDs:      &emptySlice,
		SupplierIDs:  &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemAttributeProfile",
		&ArgRemoveAttrProfile{Tenant: "cgrates.org", ID: "ApierTest", Contexts: []string{"*rating"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemAttributeProfile",
		&ArgRemoveAttrProfile{Tenant: "cgrates.org", ID: "ApierTest2", Contexts: []string{"*rating"}}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.ExternalAttributeProfile
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: tenant, ID: "ApierTest2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: tenant, ID: "ApierTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if _, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, false), engine.MetaString,
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
	if _, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.AttributeProfilePrefix, tenant, true),
		nil); err != nil && err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
