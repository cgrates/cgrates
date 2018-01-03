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
	// "net/rpc"
	"fmt"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"
	// "log"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rdsITdb *engine.RedisStorage
	mgoITdb *engine.MongoStorage
	onStor  *engine.DataManager
	err     error
	indexes map[string]utils.StringMap
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
	//to add	testV1TSGetThresholdsAfterRestart,
	// testV1FIdxSetThresholdProfile,
	// testV1FIdxUpdateThresholdProfile,
	// testV1FIdxRemoveThresholdProfile,
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
	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}

	var result string
	if err := tSv1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    tenant,
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
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
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
		tenant, true), ""); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}

}

func testV1FIdxComputeThresholdsIndexes(t *testing.T) {
	tenant := "cgrates.org"
	emptySlice := []string{}
	var reply2 string
	if err := tSv1Rpc.Call(utils.ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
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
	expectedIDX := map[string]utils.StringMap{"Account:1001": {"TEST_PROFILE1": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE1": {"Account:1001": true}}
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
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}

	var result string
	if err := tSv1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    tenant,
		ID:        "TEST_PROFILE2",
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
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
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
		tenant, true), ""); err != nil {
		t.Error(err)
	}
	if indexes, err = onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false),
		nil); err != utils.ErrNotFound {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeThresholdsIndexes(t *testing.T) {
	tenant := "cgrates.org"
	thid := []string{"TEST_PROFILE2"}
	emptySlice := []string{}
	var reply2 string
	if err := tSv1Rpc.Call(utils.ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
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
	expectedIDX := map[string]utils.StringMap{"Account:1001": {"TEST_PROFILE2": true}}
	indexes, err := onStor.GetFilterIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, false), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
	expectedRevIDX := map[string]utils.StringMap{"TEST_PROFILE2": {"Account:1001": true}}
	indexes, err = onStor.GetFilterReverseIndexes(engine.GetDBIndexKey(utils.ThresholdProfilePrefix, tenant, true), nil)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedRevIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedRevIDX, utils.ToJSON(indexes))
	}
}

// 1.set threshold in datadb fara sa faca indexuri
// 2.fac querri la index sa fiu sigur ca is 0
// 3.compile indexes all
// 4.sa verific indexurile sa fie ok pt thresholdu setat de mine
// 5.set al doilea threshold
// 6.compute cu id
// 7.sa verific indexurile sa fie ok pt thresholdu setat de mine

// func testV1FIdxGetThresholdsAfterRestart(t *testing.T) {
// 	time.Sleep(time.Second)
// 	if _, err := engine.StopStartEngine(tSv1CfgPath, thdsDelay); err != nil {
// 		t.Fatal(err)
// 	}
// 	var err error
// 	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
// 	if err != nil {
// 		t.Fatal("Could not connect to rater: ", err.Error())
// 	}
// 	var td engine.Threshold
// 	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &td); err != nil {
// 		t.Error(err)
// 	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
// 		t.Errorf("received: %+v", td)
// 	}
// 	time.Sleep(time.Duration(1 * time.Second))
// }

/*
 testV1FIdxSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter",
		RequestFilters: []*engine.RequestFilter{
			&engine.RequestFilter{
				FieldName: "*string",
				Type:      "Account",
				Values:    []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
	}

	var result string
	if err := tSv1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
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
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
}
*/

// func testV1FIdxUpdateThresholdProfile(t *testing.T) {
// 	var result string
// 	filter = &engine.Filter{
// 		Tenant: "cgrates.org",
// 		ID:     "TestFilter2",
// 		RequestFilters: []*engine.RequestFilter{
// 			&engine.RequestFilter{
// 				FieldName: "*string",
// 				Type:      "Account",
// 				Values:    []string{"10", "20"},
// 			},
// 		},
// 		ActivationInterval: &utils.ActivationInterval{
// 			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
// 			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
// 		},
// 	}

// 	if err := tSv1Rpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}
// 	tPrfl.FilterIDs = []string{"TestFilter", "TestFilter2"}
// 	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
// 		t.Error(err)
// 	} else if result != utils.OK {
// 		t.Error("Unexpected reply returned", result)
// 	}
// 	time.Sleep(time.Duration(100 * time.Millisecond)) // mongo is async
// 	var reply *engine.ThresholdProfile
// 	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
// 		t.Error(err)
// 	} else if !reflect.DeepEqual(tPrfl, reply) {
// 		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
// 	}
// }

// func testV1FIdxRemoveThresholdProfile(t *testing.T) {
// 	var resp string
// 	if err := tSv1Rpc.Call("ApierV1.RemThresholdProfile",
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
// 		t.Error(err)
// 	} else if resp != utils.OK {
// 		t.Error("Unexpected reply returned", resp)
// 	}
// 	var sqp *engine.ThresholdProfile
// 	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
// 		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
// 		err.Error() != utils.ErrNotFound.Error() {
// 		t.Error(err)
// 	}
// }

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
