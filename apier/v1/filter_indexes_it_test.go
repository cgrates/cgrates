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
MERCHANTABILITY or FIdxTNESS FOR A PARTICULAR PURPOSE.  See the
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

const (
	tenant = "cgrates.org"
)

var (
	tFIdxRpc   *rpc.Client
	emptySlice = []string{}

	sTestsFilterIndexesSV1 = []func(t *testing.T){
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

		testV1FIdxComputeIndexesMultipleProfilesAndFilters,

		testV1FIdxStopEngine,
	}
)

// Test start here
func TestFIdxV1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tSv1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tSv1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		tSv1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
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
	tFIdxRpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdWithCache{
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeThresholdsIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			ThresholdS: true,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1002"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}

	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdWithCache{
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeThresholdsIndexes(t *testing.T) {
	thid := []string{"TEST_PROFILE2"}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		utils.ArgsComputeFilterIndexIDs{
			Tenant:       tenant,
			ThresholdIDs: thid,
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1002:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxThirdComputeThresholdsIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:     tenant,
		ThresholdS: true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE1", "*string:~*req.Account:1002:TEST_PROFILE2"}
	sort.Strings(expectedIDX)
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:     tenant,
		ThresholdS: true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var sqp *engine.ThresholdProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &engine.StatQueueWithCache{
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
					MetricID: utils.MetaSum,
				},
				&engine.MetricWithFilters{
					MetricID: utils.MetaACD,
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant: tenant,
		StatS:  true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	statConfig = &engine.StatQueueWithCache{
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
					MetricID: utils.MetaACD,
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(
		utils.APIerSv1ComputeFilterIndexIDs, utils.ArgsComputeFilterIndexIDs{
			Tenant:  tenant,
			StatIDs: []string{"TEST_PROFILE2"},
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant: tenant,
		StatS:  true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: tenant, ID: "RCFG1"},
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeResourceProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:    tenant,
		ResourceS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:RCFG1"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: tenant, ID: "RCFG2"},
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil && err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeResourceProfileIndexes(t *testing.T) {
	rsid := []string{"RCFG2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		utils.ArgsComputeFilterIndexIDs{
			Tenant:      tenant,
			ResourceIDs: rsid,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:RCFG2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:    tenant,
		ResourceS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: tenant, ID: "RCFG1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveResourceProfile,
		&utils.TenantID{Tenant: tenant, ID: "RCFG2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: tenant, ID: "RCFG1"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetResourceProfile, &utils.TenantID{Tenant: tenant, ID: "RCFG2"},
		&reply2); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
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
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
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
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
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

	if err := tFIdxRpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeSupplierProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:    tenant,
		SupplierS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetSupplierProfile, splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.SupplierProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf.SupplierProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeSupplierProfileIndexes(t *testing.T) {
	spid := []string{"TEST_PROFILE2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		utils.ArgsComputeFilterIndexIDs{
			Tenant:      tenant,
			SupplierIDs: spid,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: utils.MetaString},
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, utils.ArgsComputeFilterIndexes{
		Tenant:    tenant,
		SupplierS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetSupplierProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers, Tenant: tenant, FilterType: utils.MetaString},
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err == nil ||
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
					FilterIDs: []string{"*string:~*req.FL1:In1"},
					Path:      "FL1",
					Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
				},
			},
			Weight: 20,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err != nil {
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

	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: tenant, FilterType: utils.MetaString,
		Context: utils.MetaSessionS}, &indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			Context:    utils.MetaSessionS,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:ApierTest"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Account,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
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
				FilterIDs: []string{"*string:~*req.FL1:In1"},
				Path:      "FL1",
				Value:     config.NewRSRParsersMustCompile("Al1", true, utils.INFIELD_SEP),
			}},
			Weight: 20,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err != nil {
		t.Error(err)
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
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		utils.ArgsComputeFilterIndexIDs{
			Tenant:       tenant,
			Context:      utils.MetaSessionS,
			AttributeIDs: []string{"ApierTest2"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:ApierTest2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			Context:    utils.META_ANY,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:~*req.Account:1001:ApierTest", "*string:~*req.Account:1001:ApierTest2"}
	revExpectedIDX := []string{"*string:~*req.Account:1001:ApierTest2", "*string:~*req.Account:1001:ApierTest"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			Context:    utils.MetaSessionS,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithCache{
		Tenant: tenant,
		ID:     "ApierTest"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithCache{
		Tenant: tenant,
		ID:     "ApierTest2"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithArgDispatcher{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
		Context:    utils.MetaSessionS}, &indexes); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxPopulateDatabase(t *testing.T) {
	var result string
	resPrf := ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: tenant,
			ID:     "ResProfile1",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"*string:~*req.Destination:1001",
				"*string:~*req.Destination:2001",
				"*string:~*req.Account:1002",
				"*prefix:~*req.Account:10",
				"*string:~*req.Destination:1001",
				"*prefix:~*req.Destination:20",
				"*string:~*req.Account:1002"},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	resPrf = ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: tenant,
			ID:     "ResProfile2",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"*string:~*req.Destination:1001",
				"*string:~*req.Destination:2001",
				"*string:~*req.Account:2002",
				"*prefix:~*req.Account:10",
				"*string:~*req.Destination:2001",
				"*prefix:~*req.Destination:20",
				"*string:~*req.Account:1002"},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, resPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	resPrf = ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: tenant,
			ID:     "ResProfile3",
			FilterIDs: []string{"*string:~*req.Account:3001",
				"*string:~*req.Destination:1001",
				"*string:~*req.Destination:2001",
				"*string:~*req.Account:1002",
				"*prefix:~*req.Account:10",
				"*prefix:~*req.Destination:1001",
				"*prefix:~*req.Destination:200",
				"*string:~*req.Account:1003"},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, resPrf, &result); err != nil {
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
		"*string:~*req.Account:3001:ResProfile3",
		"*string:~*req.Destination:1001:ResProfile1",
		"*string:~*req.Destination:1001:ResProfile2",
		"*string:~*req.Destination:1001:ResProfile3",
		"*string:~*req.Account:1002:ResProfile1",
		"*string:~*req.Account:1002:ResProfile2",
		"*string:~*req.Account:1002:ResProfile3",
		"*string:~*req.Account:1003:ResProfile3",
		"*prefix:~*req.Destination:20:ResProfile1",
		"*prefix:~*req.Destination:20:ResProfile2",
		"*string:~*req.Account:1001:ResProfile1",
		"*string:~*req.Account:1001:ResProfile2",
		"*string:~*req.Account:2002:ResProfile2",
		"*prefix:~*req.Destination:1001:ResProfile3",
		"*prefix:~*req.Destination:200:ResProfile3",
		"*string:~*req.Destination:2001:ResProfile1",
		"*string:~*req.Destination:2001:ResProfile2",
		"*string:~*req.Destination:2001:ResProfile3",
		"*prefix:~*req.Account:10:ResProfile1",
		"*prefix:~*req.Account:10:ResProfile2",
		"*prefix:~*req.Account:10:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
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
		"*string:~*req.Account:1003:ResProfile3",
		"*string:~*req.Account:3001:ResProfile3",
		"*string:~*req.Destination:1001:ResProfile1",
		"*string:~*req.Destination:1001:ResProfile2",
		"*string:~*req.Destination:1001:ResProfile3",
		"*string:~*req.Account:1002:ResProfile1",
		"*string:~*req.Account:1002:ResProfile2",
		"*string:~*req.Account:1002:ResProfile3",
		"*string:~*req.Account:1001:ResProfile1",
		"*string:~*req.Account:1001:ResProfile2",
		"*string:~*req.Destination:2001:ResProfile3",
		"*string:~*req.Destination:2001:ResProfile1",
		"*string:~*req.Destination:2001:ResProfile2",
		"*string:~*req.Account:2002:ResProfile2"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, reply)
	}
}

func testV1FIdxGetFilterIndexes3(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:     tenant,
		ItemType:   utils.MetaResources,
		FilterType: utils.MetaPrefix,
	}
	expectedIndexes := []string{
		"*prefix:~*req.Destination:20:ResProfile1",
		"*prefix:~*req.Destination:20:ResProfile2",
		"*prefix:~*req.Account:10:ResProfile1",
		"*prefix:~*req.Account:10:ResProfile2",
		"*prefix:~*req.Account:10:ResProfile3",
		"*prefix:~*req.Destination:200:ResProfile3",
		"*prefix:~*req.Destination:1001:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
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
		"*string:~*req.Account:1003:ResProfile3",
		"*string:~*req.Account:3001:ResProfile3",
		"*string:~*req.Account:1002:ResProfile1",
		"*string:~*req.Account:1002:ResProfile2",
		"*string:~*req.Account:1002:ResProfile3",
		"*string:~*req.Account:1001:ResProfile1",
		"*string:~*req.Account:1001:ResProfile2",
		"*string:~*req.Account:2002:ResProfile2"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
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
			FilterIDs:  []string{"*string:~*req.Account:1001", "*string:~*req.Subject:2012", "*prefix:~*req.RandomField:RandomValue"},
			Strategy:   utils.MetaFirst,
			Subsystems: []string{utils.MetaAttributes, utils.MetaSessionS},
			Weight:     20,
		},
	}

	if err := tFIdxRpc.Call(utils.APIerSv1SetDispatcherProfile,
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
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var idx []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
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
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
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
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}
	//remove the indexes for *sessions subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify if was removed
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg,
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
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(len(expectedIndexes), len(idx)) {
		t.Errorf("Expecting: %+v, received: %+v", len(expectedIndexes), len(idx))
	}
}

func testV1FIdxComputeDispatcherProfileIndexes(t *testing.T) {
	var result string
	//recompute indexes for dispatcherProfile for *sessions subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:      tenant,
			Context:     utils.MetaSessionS,
			DispatcherS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes := []string{
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
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

	if err := tFIdxRpc.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	//add a new dispatcherProfile with empty filterIDs
	//should create an index of type *none:*any:*any for *sessions subsystem
	dispatcherProfile2 := DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP_Test3",
			Subsystems: []string{utils.MetaSessionS},
			Weight:     20,
		},
	}

	if err := tFIdxRpc.Call(utils.APIerSv1SetDispatcherProfile,
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
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var idx []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
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
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}
	//remove the indexes for *sessions subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	//verify if indexes was removed for *sessions
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg,
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//remove the indexes for *attribute subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
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
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg,
		&idx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeDispatcherProfileIndexes2(t *testing.T) {
	var result string
	//recompute indexes for dispatcherProfile for *sessions subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:      tenant,
			Context:     utils.MetaSessionS,
			DispatcherS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes := []string{
		"*none:*any:*any:DSP_Test3",
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}

	//recompute indexes for dispatcherProfile for *attributes subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		utils.ArgsComputeFilterIndexes{
			Tenant:      tenant,
			Context:     utils.MetaAttributes,
			DispatcherS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes = []string{
		"*none:*any:*any:DSP_Test2",
		"*prefix:~*req.RandomField:RandomValue:DSP_Test1",
		"*string:~*req.Account:1001:DSP_Test1",
		"*string:~*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaAttributes}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}
}

func testV1FIdxComputeIndexesMultipleProfilesAndFilters(t *testing.T) {
	fltr := &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "fltr_for_attr",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Subject",
					Values:  []string{"1004", "6774", "22312"},
				},
				{
					Type:    utils.MetaString,
					Element: "~*opts.Subsystems",
					Values:  []string{"*attributes"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destinations",
					Values:  []string{"+0775", "+442"},
				},
			},
		},
	}
	fltr1 := &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "fltr_for_attr2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Usage",
					Values:  []string{"123s"},
				},
			},
		},
	}
	fltr2 := &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "fltr_for_attr3",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.AnswerTime",
					Values:  []string{"12", "33"},
				},
			},
		},
	}
	var reply string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter,
		fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter,
		fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter,
		fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply result", reply)
	}

	thPrf := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1001",
			FilterIDs: []string{"fltr_for_attr",
				"fltr_for_attr2", "fltr_for_attr3",
				"*string:~*req.Account:1001"},
			Weight:  10,
			MaxHits: -1,
			MinHits: 0,
		},
	}
	thPrf1 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
			FilterIDs: []string{"fltr_for_attr2", "fltr_for_attr3",
				"*string:~*req.Account:1001"},
			Weight:  20,
			MaxHits: 4,
			MinHits: 0,
		},
	}
	thPrf2 := &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1003",
			FilterIDs: []string{"fltr_for_attr",
				"*string:~*req.Account:1001"},
			Weight:  150,
			MaxHits: 4,
			MinHits: 2,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile,
		thPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile,
		thPrf1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile,
		thPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	// REMOVE
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		AttrRemFilterIndexes{Tenant: "cgrates.org", ItemType: utils.MetaThresholds},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	// COMPUTE AFTER REMOVE
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{Tenant: "cgrates.org", ThresholdS: true}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	// check indexes for all profiles
	var replyIdx []string
	expectedIDx := []string{"*string:~*req.Account:1001:THD_ACNT_1002",
		"*string:~*req.Account:1001:THD_ACNT_1003",
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*prefix:~*req.AnswerTime:12:THD_ACNT_1002",
		"*prefix:~*req.AnswerTime:33:THD_ACNT_1002",
		"*prefix:~*req.AnswerTime:12:THD_ACNT_1001",
		"*prefix:~*req.AnswerTime:33:THD_ACNT_1001",
		"*string:~*req.Usage:123s:THD_ACNT_1002",
		"*string:~*req.Usage:123s:THD_ACNT_1001",
		"*string:~*req.Subject:1004:THD_ACNT_1001",
		"*string:~*req.Subject:6774:THD_ACNT_1001",
		"*string:~*req.Subject:22312:THD_ACNT_1001",
		"*string:~*opts.Subsystems:*attributes:THD_ACNT_1001",
		"*prefix:~*req.Destinations:+0775:THD_ACNT_1001",
		"*prefix:~*req.Destinations:+442:THD_ACNT_1001",
		"*string:~*req.Subject:1004:THD_ACNT_1003",
		"*string:~*req.Subject:6774:THD_ACNT_1003",
		"*string:~*req.Subject:22312:THD_ACNT_1003",
		"*string:~*opts.Subsystems:*attributes:THD_ACNT_1003",
		"*prefix:~*req.Destinations:+0775:THD_ACNT_1003",
		"*prefix:~*req.Destinations:+442:THD_ACNT_1003"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: "cgrates.org", ItemType: utils.MetaThresholds}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	// REMOVE AND TRY TO COMPUTE BY IDs
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		AttrRemFilterIndexes{Tenant: "cgrates.org", ItemType: utils.MetaThresholds},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	// now we will ComputeFilterIndexes by IDs(2 of the 3 profiles)
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: "cgrates.org",
			ThresholdIDs: []string{"THD_ACNT_1001", "THD_ACNT_1002"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	expectedIDx = []string{
		"*string:~*req.Account:1001:THD_ACNT_1002",
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*prefix:~*req.AnswerTime:12:THD_ACNT_1002",
		"*prefix:~*req.AnswerTime:33:THD_ACNT_1002",
		"*prefix:~*req.AnswerTime:12:THD_ACNT_1001",
		"*prefix:~*req.AnswerTime:33:THD_ACNT_1001",
		"*string:~*req.Usage:123s:THD_ACNT_1002",
		"*string:~*req.Usage:123s:THD_ACNT_1001",
		"*string:~*req.Subject:1004:THD_ACNT_1001",
		"*string:~*req.Subject:6774:THD_ACNT_1001",
		"*string:~*req.Subject:22312:THD_ACNT_1001",
		"*string:~*opts.Subsystems:*attributes:THD_ACNT_1001",
		"*prefix:~*req.Destinations:+0775:THD_ACNT_1001",
		"*prefix:~*req.Destinations:+442:THD_ACNT_1001"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{Tenant: "cgrates.org", ItemType: utils.MetaThresholds}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expectedIDx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
		}
	}

	/*
		//now we will ComputeFilterIndexes of the remain profile
		if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
			&utils.ArgsComputeFilterIndexIDs{Tenant: "cgrates.org",
				ThresholdIDs: []string{"THD_ACNT_1003"}},
			&reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Error("Unexpected reply returned")
		}
		expectedIDx = []string{"*string:~*req.Account:1001:THD_ACNT_1002",
			"*string:~*req.Account:1001:THD_ACNT_1003",
			"*string:~*req.Account:1001:THD_ACNT_1001",
			"*prefix:~*req.AnswerTime:12:THD_ACNT_1002",
			"*prefix:~*req.AnswerTime:33:THD_ACNT_1002",
			"*prefix:~*req.AnswerTime:12:THD_ACNT_1001",
			"*prefix:~*req.AnswerTime:33:THD_ACNT_1001",
			"*string:~*req.Usage:123s:THD_ACNT_1002",
			"*string:~*req.Usage:123s:THD_ACNT_1001",
			"*string:~*req.Subject:1004:THD_ACNT_1001",
			"*string:~*req.Subject:6774:THD_ACNT_1001",
			"*string:~*req.Subject:22312:THD_ACNT_1001",
			"*string:~*opts.Subsystems:*attributes:THD_ACNT_1001",
			"*prefix:~*req.Destinations:+0775:THD_ACNT_1001",
			"*prefix:~*req.Destinations:+442:THD_ACNT_1001",
			"*string:~*req.Subject:1004:THD_ACNT_1003",
			"*string:~*req.Subject:6774:THD_ACNT_1003",
			"*string:~*req.Subject:22312:THD_ACNT_1003",
			"*string:~*opts.Subsystems:*attributes:THD_ACNT_1003",
			"*prefix:~*req.Destinations:+0775:THD_ACNT_1003",
			"*prefix:~*req.Destinations:+442:THD_ACNT_1003"}
		if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
			&AttrGetFilterIndexes{Tenant: "cgrates.org", ItemType: utils.MetaThresholds}, &replyIdx); err != nil {
			t.Error(err)
		} else {
			sort.Strings(expectedIDx)
			sort.Strings(replyIdx)
			if !reflect.DeepEqual(expectedIDx, replyIdx) {
				t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedIDx), utils.ToJSON(replyIdx))
			}
		}
	*/
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
