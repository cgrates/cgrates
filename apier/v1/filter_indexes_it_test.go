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
	"net/rpc"
	"net/rpc/jsonrpc"
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
	tFIdxRpc   *rpc.Client
	emptySlice = []string{}
)

const (
	tenant = "cgrates.org"
)

var sTestsFilterIndexesSV1 = []func(t *testing.T){
	testV1FIdxLoadConfig,
	testV1FIdxdxInitDataDb,
	testV1FIdxResetStorDb,
	testV1FIdxStartEngine,
	testV1FIdxRpcConn,

	testV1FIdxSetThresholdProfile,
	testV1FIdxComputeThresholdsIndexes,
	testV1FIdxSetSecondThresholdProfile,
	testV1FIdxSecondComputeThresholdsIndexes,
	testV1FIdxThirdComputeThresholdsIndexes,
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
	testV1FIdxComputeWithAnotherContext,
	testV1FIdxRemoveAttributeProfile,

	testV1FIdxdxInitDataDb,
	testV1FIdxPopulateDatabase,
	testV1FIdxGetFilterIndexes1,
	testV1FIdxGetFilterIndexes2,
	testV1FIdxGetFilterIndexes3,
	testV1FIdxGetFilterIndexes4,

	testV1FIdxdxInitDataDb,
	testV1FIdxSetDispatcherProfile,
	testV1FIdxComputeDispatcherProfileIndexes,
	testV1FIdxSetDispatcherProfile2,
	testV1FIdxComputeDispatcherProfileIndexes2,

	testV1FIdxStopEngine,
}

// Test start here
func TestFIdxV1ITMySQL(t *testing.T) {
	tSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsFilterIndexesSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func TestFIdxV1ITMongo(t *testing.T) {
	tSv1ConfDIR = "tutmongo"
	for _, stest := range sTestsFilterIndexesSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxLoadConfig(t *testing.T) {
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	var err error
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
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

func testV1FIdxStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxRpcConn(t *testing.T) {
	var err error
	tFIdxRpc, err = jsonrpc.Dial("tcp", tSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

//ThresholdProfile
func testV1FIdxSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"TestFilter"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeThresholdsIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			ThresholdIDs:  nil,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter2",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1002"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}

	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE2",
			FilterIDs: []string{"TestFilter2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeThresholdsIndexes(t *testing.T) {
	thid := []string{"TEST_PROFILE2"}
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			ThresholdIDs:  &thid,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1002:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxThirdComputeThresholdsIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  nil,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       &emptySlice,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE1", "*string:~Account:1002:TEST_PROFILE2"}
	sort.Strings(expectedIDX)
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveThresholdProfile(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  nil,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       &emptySlice,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
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
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//StatQueueProfile
func testV1FIdxSetStatQueueProfileIndexes(t *testing.T) {
	var reply *engine.StatQueueProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    tenant,
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
					MetricID: "*sum",
				},
				&engine.MetricWithFilters{
					MetricID: "*acd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       nil,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondStatQueueProfileIndexes(t *testing.T) {
	var reply *engine.StatQueueProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE2",
			FilterIDs: []string{"FLTR_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				&engine.MetricWithFilters{
					MetricID: "*sum",
				},
				&engine.MetricWithFilters{
					MetricID: "*acd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(
		utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &[]string{"TEST_PROFILE2"},
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveStatQueueProfile(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       nil,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//ResourceProfile
func testV1FIdxSetResourceProfileIndexes(t *testing.T) {
	var reply *engine.ResourceProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_RES_RCFG1",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: tenant, ID: "RCFG1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rlsConfig = &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    tenant,
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
			ThresholdIDs:      []string{"Val1", "Val2"},
		},
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
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeResourceProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   nil,
		StatIDs:       &emptySlice,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~Account:1001:RCFG1"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondResourceProfileIndexes(t *testing.T) {
	var reply *engine.StatQueueProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: tenant, ID: "RCFG2"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rlsConfig = &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:    tenant,
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
			ThresholdIDs:      []string{"Val1", "Val2"},
		},
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
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeResourceProfileIndexes(t *testing.T) {
	rsid := []string{"RCFG2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &rsid,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~Account:1001:RCFG2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveResourceProfile(t *testing.T) {
	var resp string
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   nil,
		StatIDs:       &emptySlice,
		SupplierIDs:   &emptySlice,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveResourceProfile",
		&utils.TenantID{Tenant: tenant, ID: "RCFG1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveResourceProfile",
		&utils.TenantID{Tenant: tenant, ID: "RCFG2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: tenant, ID: "RCFG1"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetResourceProfile", &utils.TenantID{Tenant: tenant, ID: "RCFG2"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//SupplierProfile
func testV1FIdxSetSupplierProfileIndexes(t *testing.T) {
	var reply *engine.SupplierProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					FieldName: "~Account",
					Type:      utils.MetaString,
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
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"FLTR_1"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Suppliers: []*engine.Supplier{{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_1"},
				AccountIDs:    []string{"Acc"},
				ResourceIDs:   []string{"Res1", "ResGroup2"},
				StatIDs:       []string{"Stat1"},
				Weight:        20,
				Blocker:       false,
			}},
			Weight: 10,
		},
	}

	if err := tFIdxRpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeSupplierProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       &emptySlice,
		SupplierIDs:   nil,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondSupplierProfileIndexes(t *testing.T) {
	var reply *engine.SupplierProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE2",
			FilterIDs:         []string{"FLTR_2"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Suppliers: []*engine.Supplier{{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_2"},
				AccountIDs:    []string{"Acc"},
				ResourceIDs:   []string{"Res1", "ResGroup2"},
				StatIDs:       []string{"Stat1"},
				Weight:        20,
				Blocker:       false,
			}},
			Weight: 10,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeSupplierProfileIndexes(t *testing.T) {
	spid := []string{"TEST_PROFILE2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &spid,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~Account:1001:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveSupplierProfile(t *testing.T) {
	var resp string
	var reply2 string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:        tenant,
		ThresholdIDs:  &emptySlice,
		AttributeIDs:  &emptySlice,
		ResourceIDs:   &emptySlice,
		StatIDs:       &emptySlice,
		SupplierIDs:   nil,
		ChargerIDs:    &emptySlice,
		DispatcherIDs: &emptySlice,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: engine.MetaString},
		&indexes); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//AttributeProfile
func testV1FIdxSetAttributeProfileIndexes(t *testing.T) {
	var reply *engine.AttributeProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{
		Tenant: tenant, ID: "ApierTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    tenant,
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_1"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{
				{
					FilterIDs: []string{"*string:FL1:In1"},
					FieldName: "FL1",
					Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: tenant, ID: "ApierTest"}, &reply); err != nil {
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

	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: tenant, FilterType: engine.MetaString,
		Context: utils.MetaSessionS}, &indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaSessionS,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  nil,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:ApierTest"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: engine.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondAttributeProfileIndexes(t *testing.T) {
	var reply *engine.AttributeProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				FieldName: "~Account",
				Type:      utils.MetaString,
				Values:    []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{
		Tenant: tenant, ID: "ApierTest2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &AttributeWithCache{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    tenant,
			ID:        "ApierTest2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_2"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			},
			Attributes: []*engine.Attribute{{
				FilterIDs: []string{"*string:FL1:In1"},
				FieldName: "FL1",
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			}},
			Weight: 20,
		},
	}
	if err := tFIdxRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{
		Tenant: tenant, ID: "ApierTest2"}, &reply); err != nil {
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
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: engine.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaSessionS,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &[]string{"ApierTest2"},
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:ApierTest2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: engine.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxComputeWithAnotherContext(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.META_ANY,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  nil,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~Account:1001:ApierTest", "*string:~Account:1001:ApierTest2"}
	revExpectedIDX := []string{"*string:~Account:1001:ApierTest", "*string:~Account:1001:ApierTest2"}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: engine.MetaString,
		Context:    utils.META_ANY}, &indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) && !reflect.DeepEqual(revExpectedIDX, indexes) {
		t.Errorf("Expecting: %+v or %+v, received: %+v",
			expectedIDX, revExpectedIDX, utils.ToJSON(indexes))
	}

}

func testV1FIdxRemoveAttributeProfile(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaSessionS,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  nil,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: &emptySlice,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveAttributeProfile", &utils.TenantIDWithCache{
		Tenant: tenant,
		ID:     "ApierTest"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call("ApierV1.RemoveAttributeProfile", &utils.TenantIDWithCache{
		Tenant: tenant,
		ID:     "ApierTest2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{
		Tenant: tenant, ID: "ApierTest2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{
		Tenant: tenant, ID: "ApierTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: engine.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxPopulateDatabase(t *testing.T) {
	var result string
	resPrf := &engine.ResourceProfile{
		Tenant: tenant,
		ID:     "ResProfile1",
		FilterIDs: []string{"*string:~Account:1001",
			"*string:~Destination:1001",
			"*string:~Destination:2001",
			"*string:~Account:1002",
			"*prefix:~Account:10",
			"*string:~Destination:1001",
			"*prefix:~Destination:20",
			"*string:~Account:1002"},
	}
	if err := tFIdxRpc.Call("ApierV1.SetResourceProfile", resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	resPrf = &engine.ResourceProfile{
		Tenant: tenant,
		ID:     "ResProfile2",
		FilterIDs: []string{"*string:~Account:1001",
			"*string:~Destination:1001",
			"*string:~Destination:2001",
			"*string:~Account:2002",
			"*prefix:~Account:10",
			"*string:~Destination:2001",
			"*prefix:~Destination:20",
			"*string:~Account:1002"},
	}
	if err := tFIdxRpc.Call("ApierV1.SetResourceProfile", resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	resPrf = &engine.ResourceProfile{
		Tenant: tenant,
		ID:     "ResProfile3",
		FilterIDs: []string{"*string:~Account:3001",
			"*string:~Destination:1001",
			"*string:~Destination:2001",
			"*string:~Account:1002",
			"*prefix:~Account:10",
			"*prefix:~Destination:1001",
			"*prefix:~Destination:200",
			"*string:~Account:1003"},
	}
	if err := tFIdxRpc.Call("ApierV1.SetResourceProfile", resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1FIdxGetFilterIndexes1(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:   tenant,
		ItemType: utils.MetaResources,
	}
	expectedIndexes := []string{
		"*string:~Account:3001:ResProfile3",
		"*string:~Destination:1001:ResProfile1",
		"*string:~Destination:1001:ResProfile2",
		"*string:~Destination:1001:ResProfile3",
		"*string:~Account:1002:ResProfile1",
		"*string:~Account:1002:ResProfile2",
		"*string:~Account:1002:ResProfile3",
		"*string:~Account:1003:ResProfile3",
		"*prefix:~Destination:20:ResProfile1",
		"*prefix:~Destination:20:ResProfile2",
		"*string:~Account:1001:ResProfile1",
		"*string:~Account:1001:ResProfile2",
		"*string:~Account:2002:ResProfile2",
		"*prefix:~Destination:1001:ResProfile3",
		"*prefix:~Destination:200:ResProfile3",
		"*string:~Destination:2001:ResProfile1",
		"*string:~Destination:2001:ResProfile2",
		"*string:~Destination:2001:ResProfile3",
		"*prefix:~Account:10:ResProfile1",
		"*prefix:~Account:10:ResProfile2",
		"*prefix:~Account:10:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, reply)
	}
}

func testV1FIdxGetFilterIndexes2(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:     tenant,
		ItemType:   utils.MetaResources,
		FilterType: utils.MetaString,
	}
	expectedIndexes := []string{
		"*string:~Account:1003:ResProfile3",
		"*string:~Account:3001:ResProfile3",
		"*string:~Destination:1001:ResProfile1",
		"*string:~Destination:1001:ResProfile2",
		"*string:~Destination:1001:ResProfile3",
		"*string:~Account:1002:ResProfile1",
		"*string:~Account:1002:ResProfile2",
		"*string:~Account:1002:ResProfile3",
		"*string:~Account:1001:ResProfile1",
		"*string:~Account:1001:ResProfile2",
		"*string:~Destination:2001:ResProfile3",
		"*string:~Destination:2001:ResProfile1",
		"*string:~Destination:2001:ResProfile2",
		"*string:~Account:2002:ResProfile2"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, reply)
	}
}

func testV1FIdxGetFilterIndexes3(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:     tenant,
		ItemType:   utils.MetaResources,
		FilterType: engine.MetaPrefix,
	}
	expectedIndexes := []string{
		"*prefix:~Destination:20:ResProfile1",
		"*prefix:~Destination:20:ResProfile2",
		"*prefix:~Account:10:ResProfile1",
		"*prefix:~Account:10:ResProfile2",
		"*prefix:~Account:10:ResProfile3",
		"*prefix:~Destination:200:ResProfile3",
		"*prefix:~Destination:1001:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, reply)
	}
}

func testV1FIdxGetFilterIndexes4(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:      tenant,
		ItemType:    utils.MetaResources,
		FilterType:  utils.MetaString,
		FilterField: "Account",
	}
	expectedIndexes := []string{
		"*string:~Account:1003:ResProfile3",
		"*string:~Account:3001:ResProfile3",
		"*string:~Account:1002:ResProfile1",
		"*string:~Account:1002:ResProfile2",
		"*string:~Account:1002:ResProfile3",
		"*string:~Account:1001:ResProfile1",
		"*string:~Account:1001:ResProfile2",
		"*string:~Account:2002:ResProfile2"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, reply)
	}
}

func testV1FIdxSetDispatcherProfile(t *testing.T) {
	var reply string
	//add a dispatcherProfile for 2 subsystems and verify if the index was created for both
	dispatcherProfile = &DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP_Test1",
			FilterIDs:  []string{"*string:~Account:1001", "*string:~Subject:2012", "*prefix:~RandomField:RandomValue"},
			Strategy:   utils.MetaFirst,
			Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
			Weight:     20,
		},
	}

	if err := tFIdxRpc.Call(utils.ApierV1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	//verify *string index for *attributes subsystem
	arg := &AttrGetFilterIndexes{
		Tenant:     tenant,
		Context:    utils.MetaAttributes,
		ItemType:   utils.MetaDispatchers,
		FilterType: utils.MetaString,
	}
	expectedIndexes := []string{
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var idx []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(len(expectedIndexes), len(idx)) {
		t.Errorf("Expecting: %+v, received: %+v", len(expectedIndexes), len(idx))
	}

	//verify *string index for *sessions subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:     tenant,
		Context:    utils.MetaSessionS,
		ItemType:   utils.MetaDispatchers,
		FilterType: utils.MetaString,
	}
	expectedIndexes = []string{
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(len(expectedIndexes), len(idx)) {
		t.Errorf("Expecting: %+v, received: %+v", len(expectedIndexes), len(idx))
	}

	//verify  indexes for *sessions subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}
	//remove the indexes for *sessions subsystem
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify if was removed
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg,
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//verify *string index for *attributes subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(len(expectedIndexes), len(idx)) {
		t.Errorf("Expecting: %+v, received: %+v", len(expectedIndexes), len(idx))
	}
}

func testV1FIdxComputeDispatcherProfileIndexes(t *testing.T) {
	var result string
	//recompute indexes for dispatcherProfile for *sessions subsystem
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaSessionS,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: nil,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes := []string{
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetDispatcherProfile2(t *testing.T) {
	var reply string
	//add a new dispatcherProfile with empty filterIDs
	//should create an index of type *none:*any:*any for *attributes subsystem
	dispatcherProfile = &DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP_Test2",
			Subsystems: []string{utils.MetaAttributes},
			Weight:     20,
		},
	}

	if err := tFIdxRpc.Call(utils.ApierV1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	//add a new dispatcherProfile with empty filterIDs
	//should create an index of type *none:*any:*any for *sessions subsystem
	dispatcherProfile2 := &engine.DispatcherProfile{
		Tenant:     "cgrates.org",
		ID:         "DSP_Test3",
		Subsystems: []string{utils.MetaSessionS},
		Weight:     20,
	}

	if err := tFIdxRpc.Call(utils.ApierV1SetDispatcherProfile,
		dispatcherProfile2,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	//verify indexes for *attributes subsystem
	arg := &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes := []string{
		"*none:*any:*any:DSP_Test2",
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var idx []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}

	//verify  indexes for *sessions subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{
		"*none:*any:*any:DSP_Test3",
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}
	//remove the indexes for *sessions subsystem
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify if indexes was removed for *sessions
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg,
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//remove the indexes for *attribute subsystem
	if err := tFIdxRpc.Call("ApierV1.RemoveFilterIndexes", &AttrRemFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaAttributes}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify indexes for *attributes subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", arg,
		&idx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeDispatcherProfileIndexes2(t *testing.T) {
	var result string
	//recompute indexes for dispatcherProfile for *sessions subsystem
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaSessionS,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: nil,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes := []string{
		"*none:*any:*any:DSP_Test3",
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var indexes []string
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}

	//recompute indexes for dispatcherProfile for *attributes subsystem
	if err := tFIdxRpc.Call(utils.ApierV1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:        tenant,
			Context:       utils.MetaAttributes,
			ThresholdIDs:  &emptySlice,
			AttributeIDs:  &emptySlice,
			ResourceIDs:   &emptySlice,
			StatIDs:       &emptySlice,
			SupplierIDs:   &emptySlice,
			ChargerIDs:    &emptySlice,
			DispatcherIDs: nil,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes = []string{
		"*none:*any:*any:DSP_Test2",
		"*prefix:~RandomField:RandomValue:DSP_Test1",
		"*string:~Account:1001:DSP_Test1",
		"*string:~Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call("ApierV1.GetFilterIndexes", &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaAttributes}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
