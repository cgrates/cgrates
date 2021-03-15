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

		testV1FIdxSetRouteProfileIndexes,
		testV1FIdxComputeRouteProfileIndexes,
		testV1FIdxSetSecondRouteProfileIndexes,
		testV1FIdxSecondComputeRouteProfileIndexes,
		testV1FIdxRemoveRouteProfile,

		testV1FIdxdxInitDataDb,
		testV1FISetAccountProfileIndexes,
		testV1FIComputeAccountProfileIndexes,
		testV1FISetSecondFilterForAccountProfile,
		testV1FIComputeIDsAccountProfileIndexes,
		testV1FIRemoveAccountProfile,

		testV1FIdxdxInitDataDb,
		testV1FISetActionProfileIndexes,
		testV1FIComputeActionProfileIndexes,
		testVF1SetSecondActionProfile,
		testVF1ComputeIDsActionProfileIndexes,
		testV1FIRemoveActionProfile,
		testV1FIdxdxInitDataDb,

		testV1FISetRateProfileRatesIndexes,
		testV1FIComputeRateProfileRatesIndexes,
		testV1FISetSecondRateProfileRate,
		testVF1ComputeIDsRateProfileRateIndexes,
		testVF1RemoveRateProfileRates,
		testV1FIdxdxInitDataDb,

		testV1FISetRateProfileIndexes,
		testV1FIComputeRateProfileIndexes,
		testV1FISetSecondRateProfile,
		testV1FIComputeIDsRateProfileIndexes,
		testVF1RemoveRateProfile,
		testV1FIdxdxInitDataDb,

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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			MinSleep:  5 * time.Minute,
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
		t.Error(indexes)
	}
}

func testV1FIdxComputeThresholdsIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			ThresholdS: true,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE1"}
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			MinSleep:  5 * time.Minute,
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeThresholdsIndexes(t *testing.T) {
	thid := []string{"TEST_PROFILE2"}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:       tenant,
			ThresholdIDs: thid,
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1002:TEST_PROFILE2"}
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

func testV1FIdxThirdComputeThresholdsIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
		Tenant:     tenant,
		ThresholdS: true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE1", "*string:*req.Account:1002:TEST_PROFILE2"}
	sort.Strings(expectedIDX)
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaSum,
				},
				{
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
		Tenant: tenant,
		StatS:  true,
	}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE1"}
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: "*sum",
				},
				{
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeStatQueueProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(
		utils.APIerSv1ComputeFilterIndexIDs, &utils.ArgsComputeFilterIndexIDs{
			Tenant:  tenant,
			StatIDs: []string{"TEST_PROFILE2"},
		}, &result); err != nil {
		t.Error(err)
	}
	if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE2"}
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			UsageTTL:          10 * time.Microsecond,
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeResourceProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
		Tenant:    tenant,
		ResourceS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:*req.Account:1001:RCFG1"}
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
			UsageTTL:          10 * time.Microsecond,
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeResourceProfileIndexes(t *testing.T) {
	rsid := []string{"RCFG2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:      tenant,
			ResourceIDs: rsid,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:*req.Account:1001:RCFG2"}
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
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//RouteProfile
func testV1FIdxSetRouteProfileIndexes(t *testing.T) {
	var reply *engine.RouteProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
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
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rPrf := &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"FLTR_1"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{{
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

	if err := tFIdxRpc.Call(utils.APIerSv1SetRouteProfile, rPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", rPrf.RouteProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeRouteProfileIndexes(t *testing.T) {
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
		Tenant: tenant,
		RouteS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE1"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxSetSecondRouteProfileIndexes(t *testing.T) {
	var reply *engine.RouteProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	rPrf := &RouteWithCache{
		RouteProfile: &engine.RouteProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE2",
			FilterIDs:         []string{"FLTR_2"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Routes: []*engine.Route{{
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
	if err := tFIdxRpc.Call(utils.APIerSv1SetRouteProfile, rPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rPrf.RouteProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", rPrf.RouteProfile, reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeRouteProfileIndexes(t *testing.T) {
	spid := []string{"TEST_PROFILE2"}
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:   tenant,
			RouteIDs: spid,
		}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE2"}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testV1FIdxRemoveRouteProfile(t *testing.T) {
	var resp string
	var reply2 string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes, &utils.ArgsComputeFilterIndexes{
		Tenant: tenant,
		RouteS: true,
	}, &reply2); err != nil {
		t.Error(err)
	}
	if reply2 != utils.OK {
		t.Errorf("Error: %+v", reply2)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE1"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetRouteProfile,
		&utils.TenantID{Tenant: tenant, ID: "TEST_PROFILE2"}, &reply2); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//AccountProfile
func testV1FISetAccountProfileIndexes(t *testing.T) {
	var reply *utils.AccountProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "ACCPRF_FLTR",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//there is not an accPrf in database, so we will get NOT_FOUND
	if err := tFIdxRpc.Call(utils.APIerSv1GetAccountProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACC_PRF"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//set in db an accPrf then we will get it without errors
	accPrf := &APIAccountProfileWithCache{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    tenant,
			ID:        "ACC_PRF",
			FilterIDs: []string{"*prefix:~*req.Destination:123", "ACCPRF_FLTR"},
			Balances: map[string]*utils.APIBalance{
				"ConcreteBalance": {
					ID:    "ConcreteBalance",
					Type:  utils.MetaConcrete,
					Units: 200,
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAccountProfile, accPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	newAccPrf, err := accPrf.AsAccountProfile()
	if err != nil {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAccountProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACC_PRF"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, newAccPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(newAccPrf), utils.ToJSON(reply))
	}

	var indexes []string
	expectedIDx := []string{"*string:*req.Account:1001:ACC_PRF", "*string:*req.Account:1002:ACC_PRF", "*prefix:*req.Destination:123:ACC_PRF"}
	//trying to get indexes,
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIComputeAccountProfileIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	//nothing to get from db, as we removed them
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:   tenant,
			AccountS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.Account:1001:ACC_PRF", "*string:*req.Account:1002:ACC_PRF", "*prefix:*req.Destination:123:ACC_PRF"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FISetSecondFilterForAccountProfile(t *testing.T) {
	//new filter
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "ACCPRF_FLTR2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.CGRID",
					Values:  []string{"Dan1"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//we will overwrite this AccPrf with our new filter
	accPrf := &APIAccountProfileWithCache{
		APIAccountProfile: &utils.APIAccountProfile{
			Tenant:    tenant,
			ID:        "ACC_PRF",
			FilterIDs: []string{"*prefix:~*req.Destination:123", "ACCPRF_FLTR", "ACCPRF_FLTR2"},
			Balances: map[string]*utils.APIBalance{
				"ConcreteBalance": {
					ID:    "ConcreteBalance",
					Type:  utils.MetaConcrete,
					Units: 200,
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAccountProfile, accPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	newAccPrf, err := accPrf.AsAccountProfile()
	if err != nil {
		t.Error(err)
	}
	var reply *utils.AccountProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetAccountProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACC_PRF"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, newAccPrf) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(newAccPrf), utils.ToJSON(reply))
	}

	var indexes []string
	expectedIDx := []string{"*string:*req.Account:1001:ACC_PRF", "*string:*req.Account:1002:ACC_PRF",
		"*prefix:*req.Destination:123:ACC_PRF", "*string:*req.CGRID:Dan1:ACC_PRF"}
	//trying to get indexes, should be indexes for both filters:"ACCPRF_FLTR" and "ACCPRF_FLTR2"
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIComputeIDsAccountProfileIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	//nothing to get from db, as we removed them,
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:            tenant,
			AccountProfileIDs: []string{"ACC_PRF"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.Account:1001:ACC_PRF", "*string:*req.Account:1002:ACC_PRF",
		"*prefix:*req.Destination:123:ACC_PRF", "*string:*req.CGRID:Dan1:ACC_PRF"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIRemoveAccountProfile(t *testing.T) {
	//removing accPrf from db will delete the indexes from dB
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveAccountProfile,
		&utils.TenantIDWithCache{Tenant: tenant, ID: "ACC_PRF"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	var reply *utils.AccountProfile
	//there is not an accPrf in database, so we will get NOT_FOUND
	if err := tFIdxRpc.Call(utils.APIerSv1GetAccountProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACC_PRF"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var indexes []string
	//there are no indexes in db, as we removed actprf from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaAccountProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//ActionProfile
func testV1FISetActionProfileIndexes(t *testing.T) {
	//set a new filter in db
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "ACTION_FLTR",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.ToR",
					Values:  []string{"*sms", "*data", "~*req.Voice"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//there is not an actPrf in db, so we will get NOT_FOUND
	var reply *engine.ActionProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACT_PRF"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//set an actPrf in db, so we will get it without any problems
	actPrf := &ActionProfileWithCache{
		ActionProfileWithOpts: &engine.ActionProfileWithOpts{
			ActionProfile: &engine.ActionProfile{
				Tenant:    tenant,
				ID:        "ACT_PRF",
				FilterIDs: []string{"*prefix:~*req.Account:1001|1002", "ACTION_FLTR"},
				Schedule:  "* * * * *",
				Actions: []*engine.APAction{
					{
						ID:        "TOPUP",
						FilterIDs: []string{},
						Type:      utils.MetaLog,
						Diktats: []*engine.APDiktat{{
							Path:  "~*balance.TestBalance.Value",
							Value: "10",
						}},
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetActionProfile, actPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//get it from db and compare
	if err := tFIdxRpc.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACT_PRF"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, actPrf.ActionProfile) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(actPrf.ActionProfile), utils.ToJSON(reply))
	}

	//get indexes to verify if these are indexed well
	var indexes []string
	expectedIDx := []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF",
		"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only for that inline actPRf filter (with Type *prefix)
	expectedIDx = []string{"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant, FilterType: utils.MetaPrefix},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with Field ToR
	expectedIDx = []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant, FilterField: "*req.ToR"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get the indexes only with Value 1001
	expectedIDx = []string{"*prefix:*req.Account:1001:ACT_PRF"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant, FilterValue: "1001"},
		&indexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(indexes, expectedIDx) {
		t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
	}
}

func testV1FIComputeActionProfileIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//nothing to get from db
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:  tenant,
			ActionS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF",
		"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testVF1SetSecondActionProfile(t *testing.T) {
	//second filter in db
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "ACTION_FLTR2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.OriginID",
					Values:  []string{"Dan1"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//set the second actPrf in db with our filter
	actPrf := &ActionProfileWithCache{
		ActionProfileWithOpts: &engine.ActionProfileWithOpts{
			ActionProfile: &engine.ActionProfile{
				Tenant:    tenant,
				ID:        "ACT_PRF2",
				FilterIDs: []string{"ACTION_FLTR2"},
				Actions: []*engine.APAction{
					{
						ID:        "TORESET",
						FilterIDs: []string{},
						Type:      utils.MetaLog,
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetActionProfile, actPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//get it from db and compare
	var reply *engine.ActionProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACT_PRF2"}},
		&reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, actPrf.ActionProfile) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(actPrf.ActionProfile), utils.ToJSON(reply))
	}

	//get indexes to verify if these are indexed well
	var indexes []string
	expectedIDx := []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF",
		"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF", "*string:*req.OriginID:Dan1:ACT_PRF2"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testVF1ComputeIDsActionProfileIndexes(t *testing.T) {
	//remove indexes from db for both actPrf
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	//nothing to get from db, as we removed them,
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//firstly, compute indexes for "ACT_PRF"
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:           tenant,
			ActionProfileIDs: []string{"ACT_PRF"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF",
		"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//secondly, compute indexes for "ACT_PRF2"
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:           tenant,
			ActionProfileIDs: []string{"ACT_PRF2"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx = []string{"*string:*req.ToR:*sms:ACT_PRF", "*string:*req.ToR:*data:ACT_PRF",
		"*prefix:*req.Account:1001:ACT_PRF", "*prefix:*req.Account:1002:ACT_PRF", "*string:*req.OriginID:Dan1:ACT_PRF2"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIRemoveActionProfile(t *testing.T) {
	//we will remove actionProfiles 1 by one(ACT_PRF) first
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveActionProfile,
		&utils.TenantIDWithCache{Tenant: tenant, ID: "ACT_PRF"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	var reply *engine.ActionProfile
	//there is not an actPrf in database, so we will get NOT_FOUND
	if err := tFIdxRpc.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACT_PRF"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//we will remove actionProfiles 1 by one(ACT_PRF2) second
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveActionProfile,
		&utils.TenantIDWithCache{Tenant: tenant, ID: "ACT_PRF2"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//there is not an actPrf in database, so we will get NOT_FOUND
	if err := tFIdxRpc.Call(utils.APIerSv1GetActionProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ACT_PRF2"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//bcs both profiles are removed, there are not any indexes in db
	var indexes []string
	//there are no indexes in db, as we removed actprf from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaActionProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//RateProfileRate Indexes
func testV1FISetRateProfileRatesIndexes(t *testing.T) {
	//set a filter for our rates
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "RATE_FLTR1",
			Rules: []*engine.FilterRule{{
				Type:    utils.MetaString,
				Element: "~*req.Destination",
				Values:  []string{"234"},
			}},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//there are not any rates in db
	var reply *engine.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RRATE_PRF"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//set in db a ratePrf with double populated rates with our filter
	ratePrfRates := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &engine.APIRateProfileWithOpts{
			APIRateProfile: &engine.APIRateProfile{
				Tenant:          "cgrates.org",
				ID:              "RP1",
				FilterIDs:       []string{"*string:~*req.Usage:10m"},
				MaxCostStrategy: "*free",
				Rates: map[string]*engine.APIRate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						FilterIDs:       []string{"RATE_FLTR1", "*suffix:~*req.Account:1009"},
						ActivationTimes: "* * * * 1-5",
					},
					"RT_MONTH": {
						ID:              "RT_MONTH",
						FilterIDs:       []string{"RATE_FLTR1"},
						ActivationTimes: "* * * * *",
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetRateProfile, ratePrfRates, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	expRtPrf, err := ratePrfRates.AsRateProfile()
	if err != nil {
		t.Error(err)
	}

	//get it from db and compare
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP1"}},
		&reply); err != nil {
		t.Error(err)
	} else {
		expRtPrf.Compile()
		reply.Compile()
		if !reflect.DeepEqual(reply, expRtPrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRtPrf), utils.ToJSON(reply))
		}
	}

	//get indexes to verify if these are indexed well
	var indexes []string
	expectedIDx := []string{"*suffix:*req.Account:1009:RT_WEEK", "*string:*req.Destination:234:RT_WEEK",
		"*string:*req.Destination:234:RT_MONTH"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with Type *string
	expectedIDx = []string{"*string:*req.Destination:234:RT_WEEK", "*string:*req.Destination:234:RT_MONTH"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates, Tenant: tenant,
			FilterType: utils.MetaString, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with Field Destination
	expectedIDx = []string{"*string:*req.Destination:234:RT_WEEK", "*string:*req.Destination:234:RT_MONTH"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates, Tenant: tenant,
			FilterField: "*req.Destination", Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with 1009 Destination
	expectedIDx = []string{"*suffix:*req.Account:1009:RT_WEEK"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates, Tenant: tenant,
			FilterValue: "1009", Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIComputeRateProfileRatesIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//nothing to get from db
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant: tenant,
			RateS:  true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.Destination:234:RT_WEEK",
		"*string:*req.Destination:234:RT_MONTH", "*suffix:*req.Account:1009:RT_WEEK"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FISetSecondRateProfileRate(t *testing.T) {
	//second filter for a new rate in the same rate profile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "RTPRF_FLTR3",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Usage",
					Values:  []string{"10m", "40m", "~*opts.Usage"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//append a new rate in the same rate profile
	ratePrfRates := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &engine.APIRateProfileWithOpts{
			APIRateProfile: &engine.APIRateProfile{
				Tenant:          "cgrates.org",
				ID:              "RP1",
				FilterIDs:       []string{"*string:~*req.Usage:10m"},
				MaxCostStrategy: "*free",
				Rates: map[string]*engine.APIRate{
					"RT_YEAR": {
						ID:              "RT_YEAR",
						FilterIDs:       []string{"RTPRF_FLTR3"},
						ActivationTimes: "* * * * *",
					},
				},
			},
		},
	}
	expRatePrf := utils.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Usage:10m"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				FilterIDs:       []string{"RATE_FLTR1", "*suffix:~*req.Account:1009"},
				ActivationTimes: "* * * * 1-5",
			},
			"RT_MONTH": {
				ID:              "RT_MONTH",
				FilterIDs:       []string{"RATE_FLTR1"},
				ActivationTimes: "* * * * *",
			},
			"RT_YEAR": {
				ID:              "RT_YEAR",
				FilterIDs:       []string{"RTPRF_FLTR3"},
				ActivationTimes: "* * * * *",
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetRateProfileRates, ratePrfRates, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("unexpected reply returned")
	}

	//get it from db and compare
	var reply utils.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP1"}},
		&reply); err != nil {
		t.Error(err)
	} else {
		expRatePrf.Compile()
		reply.Compile()
		if !reflect.DeepEqual(reply, expRatePrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRatePrf), utils.ToJSON(reply))
		}
	}

	var indexes []string
	expectedIDx := []string{"*string:*req.Destination:234:RT_WEEK",
		"*string:*req.Destination:234:RT_MONTH", "*suffix:*req.Account:1009:RT_WEEK",
		"*string:*req.Usage:10m:RT_YEAR", "*string:*req.Usage:40m:RT_YEAR"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testVF1ComputeIDsRateProfileRateIndexes(t *testing.T) {
	//remove indexes
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var indexes []string
	//nothing to get from db, as we removed them,
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//compute indexes for all three rates by ids
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:         tenant,
			RateProfileIDs: []string{"RP1"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.Destination:234:RT_WEEK",
		"*string:*req.Destination:234:RT_MONTH", "*suffix:*req.Account:1009:RT_WEEK",
		"*string:*req.Usage:10m:RT_YEAR", "*string:*req.Usage:40m:RT_YEAR"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testVF1RemoveRateProfileRates(t *testing.T) {
	//removing rates from db will delete the indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRates{ID: "RP1",
			Tenant: tenant, RateIDs: []string{"RT_WEEK", "RT_YEAR"}},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	expRatePrf := utils.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP1",
		FilterIDs:       []string{"*string:~*req.Usage:10m"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_MONTH": {
				ID:              "RT_MONTH",
				FilterIDs:       []string{"RATE_FLTR1"},
				ActivationTimes: "* * * * *",
			},
		},
	}

	var reply utils.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP1"}},
		&reply); err != nil {
		t.Error(err)
	} else {
		expRatePrf.Compile()
		reply.Compile()
		if !reflect.DeepEqual(reply, expRatePrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRatePrf), utils.ToJSON(reply))
		}
	}

	//compute the indexes only for the left rate
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:         tenant,
			RateProfileIDs: []string{"RP1"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.Destination:234:RT_MONTH"}
	//as we compute them, next we will try to get them again from db
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//no we will remove the left rate and the profile
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRateProfileRates,
		&RemoveRPrfRates{ID: "RP1",
			Tenant: tenant, RateIDs: []string{"RT_MONTH"}},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//no indexes in db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfileRates,
			Tenant: tenant, Context: "RP1"},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//RateProfile Indexes
func testV1FISetRateProfileIndexes(t *testing.T) {
	//set a filter for our rates
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "RATEFLTR_FLTR1",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.OriginID",
					Values:  []string{"~*opts.Account", "ID1"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Destination",
					Values:  []string{"~*opts.Account", "123"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//there are not any rates in db
	var reply *engine.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//set in db a ratePrf with with our filterS
	ratePrfRates := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &engine.APIRateProfileWithOpts{
			APIRateProfile: &engine.APIRateProfile{
				Tenant:          "cgrates.org",
				ID:              "RP2",
				FilterIDs:       []string{"*string:~*req.Usage:10m", "RATEFLTR_FLTR1"},
				MaxCostStrategy: "*free",
				Rates: map[string]*engine.APIRate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						FilterIDs:       []string{"*suffix:~*req.Account:1009"},
						ActivationTimes: "* * * * 1-5",
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetRateProfile, ratePrfRates, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expRtPrf, err := ratePrfRates.AsRateProfile()
	if err != nil {
		t.Error(err)
	}

	//get it from db and compare
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP2"}},
		&reply); err != nil {
		t.Error(err)
	} else {
		expRtPrf.Compile()
		reply.Compile()
		if !reflect.DeepEqual(reply, expRtPrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRtPrf), utils.ToJSON(reply))
		}
	}

	//get indexes to verify if these are indexed well
	var indexes []string
	expectedIDx := []string{"*string:*req.OriginID:ID1:RP2", "*prefix:*req.Destination:123:RP2",
		"*string:*req.Usage:10m:RP2"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with Type *string
	expectedIDx = []string{"*string:*req.OriginID:ID1:RP2",
		"*string:*req.Usage:10m:RP2"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles, Tenant: tenant,
			FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with Field OriginID
	expectedIDx = []string{"*string:*req.OriginID:ID1:RP2"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles, Tenant: tenant,
			FilterField: "*req.OriginID"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}

	//get indexes only with 10m
	expectedIDx = []string{"*string:*req.Usage:10m:RP2"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles, Tenant: tenant,
			FilterValue: "10m"},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIComputeRateProfileIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//nothing to get from db
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant: tenant,
			RateS:  true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.OriginID:ID1:RP2", "*prefix:*req.Destination:123:RP2",
		"*string:*req.Usage:10m:RP2"}
	//as we compute them, next we will try to get them again from db
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FISetSecondRateProfile(t *testing.T) {
	//second filter for a new rate profile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "RTPRF_FLTR6",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.ToR",
					Values:  []string{"*sms", "~*opts.Usage"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//another rate profile
	ratePrfRates := &APIRateProfileWithCache{
		APIRateProfileWithOpts: &engine.APIRateProfileWithOpts{
			APIRateProfile: &engine.APIRateProfile{
				Tenant:          "cgrates.org",
				ID:              "RP3",
				FilterIDs:       []string{"RTPRF_FLTR6"},
				MaxCostStrategy: "*free",
				Rates: map[string]*engine.APIRate{
					"RT_WEEK": {
						ID:              "RT_WEEK",
						FilterIDs:       []string{"*suffix:~*req.Account:1019"},
						ActivationTimes: "* * * * 1-5",
					},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetRateProfile, ratePrfRates, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("unexpected reply returned")
	}
	expRatePrf := utils.RateProfile{
		Tenant:          "cgrates.org",
		ID:              "RP3",
		FilterIDs:       []string{"RTPRF_FLTR6"},
		MaxCostStrategy: "*free",
		Rates: map[string]*utils.Rate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				FilterIDs:       []string{"*suffix:~*req.Account:1019"},
				ActivationTimes: "* * * * 1-5",
			},
		},
	}
	//get it from db and compare
	var reply utils.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP3"}},
		&reply); err != nil {
		t.Error(err)
	} else {
		expRatePrf.Compile()
		reply.Compile()
		if !reflect.DeepEqual(reply, expRatePrf) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expRatePrf), utils.ToJSON(reply))
		}
	}

	var indexes []string
	expectedIDx := []string{"*string:*req.OriginID:ID1:RP2", "*prefix:*req.Destination:123:RP2",
		"*string:*req.Usage:10m:RP2", "*string:*req.ToR:*sms:RP3"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testV1FIComputeIDsRateProfileIndexes(t *testing.T) {
	//remove indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//nothing to get from db
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	//compute them, to put indexes again in db for the right subsystem
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:         tenant,
			RateProfileIDs: []string{"RP3", "RP2"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIDx := []string{"*string:*req.OriginID:ID1:RP2", "*prefix:*req.Destination:123:RP2",
		"*string:*req.Usage:10m:RP2", "*string:*req.ToR:*sms:RP3"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{ItemType: utils.MetaRateProfiles,
			Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expectedIDx)
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIDx) {
			t.Errorf("Expected %+v, received %+v", expectedIDx, indexes)
		}
	}
}

func testVF1RemoveRateProfile(t *testing.T) {
	//removing rate profile from db will delete the indexes from db
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRateProfile,
		&utils.TenantIDWithCache{ID: "RP2",
			Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1RemoveRateProfile,
		&utils.TenantIDWithCache{ID: "RP3",
			Tenant: tenant},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected result returned", result)
	}

	//nothing to get from db
	var reply utils.RateProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP2"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1GetRateProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "RP3"}},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//as we removed our profiles, the indexes are removed as well
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRateProfiles, Tenant: tenant},
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//AttributeProfile Indexes
func testV1FIdxSetAttributeProfileIndexes(t *testing.T) {
	var reply *engine.AttributeProfile
	filter = &FilterWithCache{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.AttributeProfileWithOpts{
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
					Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err != nil {
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
		Context: utils.MetaSessionS}, &indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			Context:    utils.MetaSessionS,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1001:ApierTest"}
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
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
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
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.AttributeProfileWithOpts{
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
				Value:     config.NewRSRParsersMustCompile("Al1", utils.InfieldSep),
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
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{
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
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{
			ItemType: utils.MetaAttributes,
			Tenant:   tenant,
			Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			ItemType:   utils.MetaAttributes,
			Tenant:     tenant,
			FilterType: utils.MetaString,
			Context:    utils.MetaSessionS}, &indexes); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxSecondComputeAttributeProfileIndexes(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:       tenant,
			Context:      utils.MetaSessionS,
			AttributeIDs: []string{"ApierTest2"},
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIDX := []string{"*string:*req.Account:1001:ApierTest2"}
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
		&utils.ArgsComputeFilterIndexes{
			Tenant:     tenant,
			Context:    utils.MetaAny,
			AttributeS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	var indexes []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType:   utils.MetaAttributes,
		Tenant:     tenant,
		FilterType: utils.MetaString,
		Context:    utils.MetaAny}, &indexes); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
		t.Error(indexes)
	}
}

func testV1FIdxRemoveAttributeProfile(t *testing.T) {
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
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
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithOpts{TenantID: &utils.TenantID{
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
		"*string:*req.Account:3001:ResProfile3",
		"*string:*req.Destination:1001:ResProfile1",
		"*string:*req.Destination:1001:ResProfile2",
		"*string:*req.Destination:1001:ResProfile3",
		"*string:*req.Account:1002:ResProfile1",
		"*string:*req.Account:1002:ResProfile2",
		"*string:*req.Account:1002:ResProfile3",
		"*string:*req.Account:1003:ResProfile3",
		"*prefix:*req.Destination:20:ResProfile1",
		"*prefix:*req.Destination:20:ResProfile2",
		"*string:*req.Account:1001:ResProfile1",
		"*string:*req.Account:1001:ResProfile2",
		"*string:*req.Account:2002:ResProfile2",
		"*prefix:*req.Destination:1001:ResProfile3",
		"*prefix:*req.Destination:200:ResProfile3",
		"*string:*req.Destination:2001:ResProfile1",
		"*string:*req.Destination:2001:ResProfile2",
		"*string:*req.Destination:2001:ResProfile3",
		"*prefix:*req.Account:10:ResProfile1",
		"*prefix:*req.Account:10:ResProfile2",
		"*prefix:*req.Account:10:ResProfile3"}
	sort.Strings(expectedIndexes)
	var reply []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &reply); err != nil {
		t.Error(err)
	} else if sort.Strings(reply); !reflect.DeepEqual(expectedIndexes, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(reply))
	}
}

func testV1FIdxGetFilterIndexes2(t *testing.T) {
	arg := &AttrGetFilterIndexes{
		Tenant:     tenant,
		ItemType:   utils.MetaResources,
		FilterType: utils.MetaString,
	}
	expectedIndexes := []string{
		"*string:*req.Account:1003:ResProfile3",
		"*string:*req.Account:3001:ResProfile3",
		"*string:*req.Destination:1001:ResProfile1",
		"*string:*req.Destination:1001:ResProfile2",
		"*string:*req.Destination:1001:ResProfile3",
		"*string:*req.Account:1002:ResProfile1",
		"*string:*req.Account:1002:ResProfile2",
		"*string:*req.Account:1002:ResProfile3",
		"*string:*req.Account:1001:ResProfile1",
		"*string:*req.Account:1001:ResProfile2",
		"*string:*req.Destination:2001:ResProfile3",
		"*string:*req.Destination:2001:ResProfile1",
		"*string:*req.Destination:2001:ResProfile2",
		"*string:*req.Account:2002:ResProfile2"}
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
		"*prefix:*req.Destination:20:ResProfile1",
		"*prefix:*req.Destination:20:ResProfile2",
		"*prefix:*req.Account:10:ResProfile1",
		"*prefix:*req.Account:10:ResProfile2",
		"*prefix:*req.Account:10:ResProfile3",
		"*prefix:*req.Destination:200:ResProfile3",
		"*prefix:*req.Destination:1001:ResProfile3"}
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
		"*string:*req.Account:1003:ResProfile3",
		"*string:*req.Account:3001:ResProfile3",
		"*string:*req.Account:1002:ResProfile1",
		"*string:*req.Account:1002:ResProfile2",
		"*string:*req.Account:1002:ResProfile3",
		"*string:*req.Account:1001:ResProfile1",
		"*string:*req.Account:1001:ResProfile2",
		"*string:*req.Account:2002:ResProfile2"}
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
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		&indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//verify *string index for *attributes subsystem
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		&utils.ArgsComputeFilterIndexes{
			Tenant:      tenant,
			Context:     utils.MetaSessionS,
			DispatcherS: true,
		}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Error: %+v", result)
	}
	expectedIndexes := []string{
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		&utils.ArgsComputeFilterIndexes{
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
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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
		&utils.ArgsComputeFilterIndexes{
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
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
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

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
