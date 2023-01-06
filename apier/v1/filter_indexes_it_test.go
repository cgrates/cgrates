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

const (
	tenant = "cgrates.org"
)

var (
	tFIdxRpc *rpc.Client

	sTestsFilterIndexesSV1 = []func(t *testing.T){
		testV1FIdxLoadConfig,
		testV1FIdxdxInitDataDb,
		testV1FIdxResetStorDb,
		testV1FIdxStartEngine,
		testV1FIdxRpcConn,

		testSetProfilesWithFltrsAndOverwriteThemFIdx,
		testSetAndChangeFiltersOnProfiles,

		/* testV1FIdxSetThresholdProfile,
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

		testV1FIdxSetAttributeProfileIndexes,
		testV1FIdxComputeAttributeProfileIndexes,
		testV1FIdxSetSecondAttributeProfileIndexes,
		testV1FIdxSecondComputeAttributeProfileIndexes,
		testV1FIdxComputeWithAnotherContext,
		testV1FIdxRemoveAttributeProfile,
		// special case for multiple attributes and filters
		testV1FIdxSetMultipleAttributesMultipleFilters,
		// special case for multiple context and compute filters by different contexts of the same ID
		testV1FIdxdxInitDataDb,
		testV1FIdxResetStorDb,
		testV1FIdxClearCache,
		testV1FIdxSetAttributeProfileMultipleContextsAndComputes,

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

		// special cases with  multiple profiles and filters
		testV1FIdxdxInitDataDb,
		testV1FIdxResetStorDb,
		testV1FIdxClearCache,
		//testV1FIdxSetDispatcherComputeIDs,
		testV1FIdxSetResourceComputeIDs, */

		testV1FIdxStopEngine,
	}
)

// Test start here
func TestFIdxV1IT(t *testing.T) {
	tSv1InternalRestart = false
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
	if *dbType == utils.MetaInternal && tSv1InternalRestart {
		testV1FIdxHStopEngine(t)
		testV1FIdxStartEngine(t)
		testV1FIdxRpcConn(t)
		return
	}
	tSv1InternalRestart = true
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1IndexClearCache(t *testing.T) {
	var reply string
	if err := tFIdxRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{}, &reply); err != nil {
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

// ThresholdProfile
func testV1FIdxSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
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
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    tenant,
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"TestFilter"},
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

	var indexes []string
	expectedIdx := []string{"*string:*req.Account:1001:TEST_PROFILE1"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(indexes, expectedIdx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIdx, indexes)
	}

	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1006", "1009"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	expectedIdx = []string{"*string:*req.Account:1006:TEST_PROFILE1",
		"*string:*req.Account:1009:TEST_PROFILE1"}
	sort.Strings(expectedIdx)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(indexes)
		if !reflect.DeepEqual(indexes, expectedIdx) {
			t.Errorf("Expecting: %+v, received: %+v", expectedIdx, indexes)
		}
	}

	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "TestFilter",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaThresholds, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

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
	filter = &engine.FilterWithAPIOpts{
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
	tPrfl = &engine.ThresholdProfileWithAPIOpts{
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
	expectedIDX := []string{"*string:*req.Account:1001:TEST_PROFILE1",
		"*string:*req.Account:1002:TEST_PROFILE2"}
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

// StatQueueProfile
func testV1FIdxSetStatQueueProfileIndexes(t *testing.T) {
	var reply *engine.StatQueueProfile
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
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
	statConfig = &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      tenant,
			ID:          "TEST_PROFILE1",
			FilterIDs:   []string{"FLTR_1"},
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

	//update the filter(element and values) for getting the indexes well
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Destinations,
				Type:    utils.MetaString,
				Values:  []string{"+122", "+5543"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	expectedIdx := []string{"*string:*req.Destinations:+122:TEST_PROFILE1",
		"*string:*req.Destinations:+5543:TEST_PROFILE1"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(indexes)
		sort.Strings(expectedIdx)
		if !reflect.DeepEqual(indexes, expectedIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIdx, indexes)
		}
	}

	//back to our initial filter
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaStats, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
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
	filter = &engine.FilterWithAPIOpts{
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
	statConfig = &engine.StatQueueProfileWithAPIOpts{
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

// ResourceProfile
func testV1FIdxSetResourceProfileIndexes(t *testing.T) {
	var reply *engine.ResourceProfile
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_RES_RCFG1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
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
	rlsConfig = &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            tenant,
			ID:                "RCFG1",
			FilterIDs:         []string{"FLTR_RES_RCFG1"},
			UsageTTL:          10 * time.Microsecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"Val1", "Val2"},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetResourceProfile, rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	//update the filter(element and values) for getting the indexes well
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_RES_RCFG1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
				Type:    utils.MetaString,
				Values:  []string{"20m", "45m", "10s"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	expectedIdx := []string{"*string:*req.Usage:20m:RCFG1",
		"*string:*req.Usage:45m:RCFG1",
		"*string:*req.Usage:10s:RCFG1"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(indexes)
		sort.Strings(expectedIdx)
		if !reflect.DeepEqual(indexes, expectedIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIdx, indexes)
		}
	}

	//back to our initial filter
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_RES_RCFG1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
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
	var reply *engine.ResourceProfile
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
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
	rlsConfig = &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            tenant,
			ID:                "RCFG2",
			FilterIDs:         []string{"FLTR_2"},
			UsageTTL:          10 * time.Microsecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"Val1", "Val2"},
		},
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

// RouteProfile
func testV1FIdxSetRouteProfileIndexes(t *testing.T) {
	var reply *engine.RouteProfile
	filter = &engine.FilterWithAPIOpts{
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
	rPrf := &RouteWithAPIOpts{
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

	//updating our filter(values and element) for getting the indexes well
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.CGRID,
				Type:    utils.MetaString,
				Values:  []string{"qweasdzxc", "iopjklbnm"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	expectedIdx := []string{"*string:*req.CGRID:qweasdzxc:TEST_PROFILE1",
		"*string:*req.CGRID:iopjklbnm:TEST_PROFILE1"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(indexes)
		sort.Strings(expectedIdx)
		if !reflect.DeepEqual(indexes, expectedIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIdx, indexes)
		}
	}

	//back to our initial filter
	filter = &engine.FilterWithAPIOpts{
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
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaRoutes, Tenant: tenant}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
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
	filter = &engine.FilterWithAPIOpts{
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
	rPrf := &RouteWithAPIOpts{
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

// AttributeProfile Indexes
func testV1FIdxSetAttributeProfileIndexes(t *testing.T) {
	var reply *engine.AttributeProfile
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    tenant,
			ID:        "ApierTest",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_1"},
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
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: tenant, ID: "ApierTest"}}, &reply); err != nil {
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

	//we will update the filter(element and values) for getting the indexes well
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaOpts + utils.NestingSep + utils.Subsystems,
				Type:    utils.MetaString,
				Values:  []string{utils.MetaChargers, utils.MetaThresholds, utils.MetaStats},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var indexes []string
	expectedIdx := []string{"*string:*opts.Subsystems:*chargers:ApierTest",
		"*string:*opts.Subsystems:*thresholds:ApierTest",
		"*string:*opts.Subsystems:*stats:ApierTest"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes, Tenant: tenant, FilterType: utils.MetaString,
		Context: utils.MetaSessionS},
		&indexes); err != nil {
		t.Error(err)
	} else {
		sort.Strings(indexes)
		sort.Strings(expectedIdx)
		if !reflect.DeepEqual(indexes, expectedIdx) {
			t.Errorf("Expected %+v \n, received %+v", expectedIdx, indexes)
		}
	}

	//back to our initial filter
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		ItemType: utils.MetaAttributes,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
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
	filter = &engine.FilterWithAPIOpts{
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
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	alsPrf = &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    tenant,
			ID:        "ApierTest2",
			Contexts:  []string{utils.MetaSessionS},
			FilterIDs: []string{"FLTR_2"},
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
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
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
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: tenant,
		ID:     "ApierTest"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveAttributeProfile, &utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
		Tenant: tenant,
		ID:     "ApierTest2"}}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
			Tenant: tenant, ID: "ApierTest2"}}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{
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

func testV1FIdxSetMultipleAttributesMultipleFilters(t *testing.T) {
	fltr1 := &engine.FilterWithAPIOpts{
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
	fltr2 := &engine.FilterWithAPIOpts{
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
	fltr := &engine.FilterWithAPIOpts{
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
	// First we will set a filter for usage
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

	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:   "cgrates.org",
			ID:       "TEST_ATTRIBUTES_IT_TEST",
			Contexts: []string{utils.MetaAny},
			FilterIDs: []string{"fltr_for_attr", "*string:~*opts.*context:*sessions",
				"fltr_for_attr2", "fltr_for_attr3"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("cgrates.itsyscom", utils.InfieldSep),
				},
			},
		},
	}
	attrPrf2 := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_ATTRIBUTES_new_fltr",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"fltr_for_attr2", "fltr_for_attr3", "*string:~*opts.*context:*chargers"},
			Attributes: []*engine.Attribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("1002", utils.InfieldSep),
				},
			},
		},
	}
	attrPrf3 := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_ATTRIBUTE3",
			Contexts:  []string{utils.MetaAny},
			FilterIDs: []string{"fltr_for_attr3", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Destinations",
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("1008", utils.InfieldSep),
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile,
		attrPrf2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile,
		attrPrf3, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes,
		&AttrRemFilterIndexes{
			Tenant:   "cgrates.org",
			Context:  utils.MetaAny,
			ItemType: utils.MetaAttributes,
		},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//not found for both cases
	var replyIdx []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			Tenant:   "cgrates.org",
			Context:  utils.MetaAny,
			ItemType: utils.MetaAttributes,
		}, &replyIdx); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// now we will ComputeFilterIndexes by IDs for *sessions context(but just only 1 profile, not both)
	var expIdx []string
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{
			Tenant:       "cgrates.org",
			Context:      utils.MetaAny,
			AttributeIDs: []string{"TEST_ATTRIBUTES_new_fltr", "TEST_ATTRIBUTE3"},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}

	//able to get indexes with context *sessions
	expIdx = []string{"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*chargers:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTE3",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr"}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			Tenant:   "cgrates.org",
			Context:  utils.MetaAny,
			ItemType: utils.MetaAttributes,
		}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}

	// compute for the last profile remain
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexIDs,
		&utils.ArgsComputeFilterIndexIDs{Tenant: "cgrates.org",
			Context:      utils.MetaAny,
			AttributeIDs: []string{"TEST_ATTRIBUTES_IT_TEST"},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned")
	}
	expIdx = []string{
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE3",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTES_new_fltr",
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTES_IT_TEST",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.*context:*chargers:TEST_ATTRIBUTES_new_fltr",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTE3",
		"*string:*opts.*context:*sessions:TEST_ATTRIBUTES_IT_TEST",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:1004:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:22312:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Subject:6774:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_IT_TEST",
		"*string:*req.Usage:123s:TEST_ATTRIBUTES_new_fltr",
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes,
		&AttrGetFilterIndexes{
			Tenant:   "cgrates.org",
			Context:  utils.MetaAny,
			ItemType: utils.MetaAttributes}, &replyIdx); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expIdx)
		sort.Strings(replyIdx)
		if !reflect.DeepEqual(expIdx, replyIdx) {
			t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expIdx), utils.ToJSON(replyIdx))
		}
	}
}

func testV1FIdxSetAttributeProfileMultipleContextsAndComputes(t *testing.T) {
	// set multiple filters for usage
	fltr1 := &engine.FilterWithAPIOpts{
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
	fltr2 := &engine.FilterWithAPIOpts{
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
	fltr := &engine.FilterWithAPIOpts{
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
	var reply string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, fltr, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, fltr1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, fltr2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	// set an attributeProfile with multiple contexts
	attrPrf := &engine.AttributeProfileWithAPIOpts{
		AttributeProfile: &engine.AttributeProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_ATTRIBUTE_CONTEXTS",
			Contexts:  []string{utils.MetaChargers, utils.MetaSessionS, utils.MetaThresholds},
			FilterIDs: []string{"fltr_for_attr3", "fltr_for_attr", "fltr_for_attr2"},
			Attributes: []*engine.Attribute{
				{
					Path:  "*req.Usage",
					Type:  utils.MetaConstant,
					Value: config.NewRSRParsersMustCompile("10m", utils.InfieldSep),
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}

	// check indexes for our attributeProfile
	expIdx := []string{
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:6774:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:1004:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:22312:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Usage:123s:TEST_ATTRIBUTE_CONTEXTS",
	}
	sort.Strings(expIdx)
	var result []string
	// same expecteded indexes for *sessions, *chargers and *thresholds
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaChargers,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaThresholds,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}

	// remove indexes for all contexts
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaAttributes,
	}, &reply); err != nil {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaChargers,
		ItemType: utils.MetaAttributes,
	}, &reply); err != nil {
		t.Error(err)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1RemoveFilterIndexes, &AttrRemFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaThresholds,
		ItemType: utils.MetaAttributes,
	}, &reply); err != nil {
		t.Error(err)
	}

	// compute indexes by with different contexts and check them
	// firstly for *sessions context
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:     "cgrates.org",
			Context:    utils.MetaSessionS,
			AttributeS: true,
		}, &reply); err != nil {
		t.Error(err)
	}

	// check the indexes after computing with *sessions context, the rest will not be computed

	expIdx = []string{
		"*prefix:*req.Destinations:+0775:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:6774:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.AnswerTime:33:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.AnswerTime:12:TEST_ATTRIBUTE_CONTEXTS",
		"*prefix:*req.Destinations:+442:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:1004:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Subject:22312:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*opts.Subsystems:*attributes:TEST_ATTRIBUTE_CONTEXTS",
		"*string:*req.Usage:123s:TEST_ATTRIBUTE_CONTEXTS",
	}
	sort.Strings(expIdx)
	// same expected indexes for *sessions, *chargers and *thresholds
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}

	// as for *sessions was computed, for *chargers and *thresaholds should not be computed, so NOT FOUND will be returned
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaChargers,
		ItemType: utils.MetaAttributes,
	}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaThresholds,
		ItemType: utils.MetaAttributes,
	}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	// noe we will compute for *chargers, and the remain context for compute indexes will remain *thresholds
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:     "cgrates.org",
			Context:    utils.MetaChargers,
			AttributeS: true,
		}, &reply); err != nil {
		t.Error(err)
	}

	// check for *sesssions and for *chargers, and for *threshold will be NOT FOUND
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaSessionS,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaChargers,
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}

	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaThresholds,
		ItemType: utils.MetaAttributes,
	}, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	// compute with the remain context *thresholds, so in the end, all indexes will be computed for all contexts
	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:     "cgrates.org",
			Context:    utils.MetaThresholds,
			AttributeS: true,
		}, &reply); err != nil {
		t.Error(err)
	}

	// check again all the indexes for all contexts
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaSessionS, // *sesssions
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaChargers, // *chargers
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		Context:  utils.MetaThresholds, // *thresholds
		ItemType: utils.MetaAttributes,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		if !reflect.DeepEqual(result, expIdx) {
			t.Errorf("Expected %+v received %+v", utils.ToJSON(expIdx), utils.ToJSON(result))
		}
	}
}

func testV1FIdxPopulateDatabase(t *testing.T) {
	var result string
	resPrf := engine.ResourceProfileWithAPIOpts{
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
	resPrf = engine.ResourceProfileWithAPIOpts{
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
	resPrf = engine.ResourceProfileWithAPIOpts{
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
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Subject,
					Type:    utils.MetaString,
					Values:  []string{"2012"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "RandomField",
					Type:    utils.MetaPrefix,
					Values:  []string{"RandomValue"},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//add a dispatcherProfile for 2 subsystems and verify if the index was created for both
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP_Test1",
			FilterIDs:  []string{"FLTR_1"},
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

	//update the filter(element and values) for getting the indexes well
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001", "1234"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Usage,
					Type:    utils.MetaString,
					Values:  []string{"15m", "1s"},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//get new indexes for *attributes context
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Account:1234:DSP_Test1",
		"*string:*req.Usage:15m:DSP_Test1",
		"*string:*req.Usage:1s:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant: tenant, Context: utils.MetaAttributes, ItemType: utils.MetaDispatchers,
	}, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}
	//get new indexes for *sessions context
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Account:1234:DSP_Test1",
		"*string:*req.Usage:15m:DSP_Test1",
		"*string:*req.Usage:1s:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant: tenant, Context: utils.MetaSessionS, ItemType: utils.MetaDispatchers,
	}, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(expectedIndexes, idx) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, idx)
	}

	//back to our initial filter
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Subject,
					Type:    utils.MetaString,
					Values:  []string{"2012"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "RandomField",
					Type:    utils.MetaPrefix,
					Values:  []string{"RandomValue"},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
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
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, AttrGetFilterIndexes{
		Tenant: tenant, Context: utils.MetaSessionS, ItemType: utils.MetaDispatchers,
	}, &indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
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
	dispatcherProfile = &DispatcherWithAPIOpts{
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
	dispatcherProfile2 := DispatcherWithAPIOpts{
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

func testV1FIdxClearCache(t *testing.T) {
	var reply string
	if err := tFIdxRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
}

func testV1FIdxSetDispatcherComputeIDs(t *testing.T) {
	var reply string
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_1",
			Rules: []*engine.FilterRule{
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
					Type:    utils.MetaString,
					Values:  []string{"1001"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.Subject,
					Type:    utils.MetaString,
					Values:  []string{"2012"},
				},
				{
					Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + "RandomField",
					Type:    utils.MetaPrefix,
					Values:  []string{"RandomValue"},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//add a dispatcherProfile for 2 subsystems and verify if the index was created for both
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP_Test1",
			FilterIDs:  []string{"FLTR_1"},
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
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, AttrGetFilterIndexes{
		Tenant: tenant, Context: utils.MetaSessionS, ItemType: utils.MetaDispatchers,
	}, &indexes); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	//verify *string index for *attributes subsystem
	arg := &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes := []string{
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
	}
	var idx []string
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &idx); err != nil {
		t.Error(err)
	} else if sort.Strings(idx); !reflect.DeepEqual(len(expectedIndexes), len(idx)) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(idx))
	}

	if err := tFIdxRpc.Call(utils.APIerSv1ComputeFilterIndexes,
		&utils.ArgsComputeFilterIndexes{
			Tenant:      tenant,
			Context:     utils.MetaSessionS,
			DispatcherS: true,
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Error: %+v", reply)
	}
	expectedIndexes = []string{
		"*prefix:*req.RandomField:RandomValue:DSP_Test1",
		"*string:*req.Account:1001:DSP_Test1",
		"*string:*req.Subject:2012:DSP_Test1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Tenant:   tenant,
		Context:  utils.MetaSessionS}, &indexes); err != nil {
		t.Error(err)
	} else if sort.Strings(indexes); !reflect.DeepEqual(expectedIndexes, indexes) {
		t.Errorf("Expecting: %+v, received: %+v", expectedIndexes, utils.ToJSON(indexes))
	}

	//verify *string index for *attributes subsystem after computing for *session subsystem
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
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(idx))
	}

	//add a new dispatcherProfile with empty filterIDs
	//should create an index of type *none:*any:*any for *attributes subsystem
	dispatcherProfile = &DispatcherWithAPIOpts{
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
	dispatcherProfile2 := DispatcherWithAPIOpts{
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
	arg = &AttrGetFilterIndexes{
		Tenant:   tenant,
		Context:  utils.MetaAttributes,
		ItemType: utils.MetaDispatchers,
	}
	expectedIndexes = []string{
		"*none:*any:*any:DSP_Test2",
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
}

func testV1FIdxSetResourceComputeIDs(t *testing.T) {
	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_RES_RCFG1",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	rlsConfig = &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            tenant,
			ID:                "RCFG1",
			FilterIDs:         []string{"FLTR_RES_RCFG1"},
			UsageTTL:          10 * time.Microsecond,
			Limit:             10,
			AllocationMessage: "MessageAllocation",
			Blocker:           true,
			Stored:            true,
			Weight:            20,
			ThresholdIDs:      []string{"Val1", "Val2"},
		},
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
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}

	filter = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: tenant,
			ID:     "FLTR_2",
			Rules: []*engine.FilterRule{{
				Element: utils.DynamicDataPrefix + utils.MetaReq + utils.NestingSep + utils.AccountField,
				Type:    utils.MetaString,
				Values:  []string{"1001"},
			}},
		},
	}

	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	rlsConfig = &engine.ResourceProfileWithAPIOpts{
		ResourceProfile: &engine.ResourceProfile{
			Tenant:            tenant,
			ID:                "RCFG2",
			FilterIDs:         []string{"FLTR_2"},
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

	expectedIDX = []string{
		"*string:*req.Account:1001:RCFG1",
		"*string:*req.Account:1001:RCFG2",
	}
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources, Tenant: tenant, FilterType: utils.MetaString},
		&indexes); err != nil {
		t.Error(err)
	}
	sort.Strings(indexes)
	if !reflect.DeepEqual(expectedIDX, indexes) {
		t.Errorf("Expecting: %+v, received: %+v",
			expectedIDX, utils.ToJSON(indexes))
	}
}

func testSetProfilesWithFltrsAndOverwriteThemFIdx(t *testing.T) {
	// FLTR_Charger, FLTR_Charger2  will be changed
	filter1 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Bank",
					Values:  []string{"BoA", "CEC"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Customer",
					Values:  []string{"11", "22"},
				},
			},
		},
	}
	filter2 := &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Account",
					Values:  []string{"1001"},
				},
			},
		},
	}
	var result string
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	stat1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats1",
			FilterIDs: []string{
				"FLTR_Charger",
				"FLTR_Charger2",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	stat2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			FilterIDs: []string{
				"FLTR_Charger",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	var reply string
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, stat1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, stat2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	tPrfl1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "TEST_PROFILE1",
			FilterIDs: []string{"FLTR_Charger"},
			MaxHits:   1,
			MinSleep:  time.Duration(5 * time.Minute),
			Blocker:   false,
			Weight:    10.0,
			Async:     true,
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	arg := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		// Stats1
		"*string:*req.Bank:BoA:Stats1",
		"*string:*req.Bank:CEC:Stats1",
		"*prefix:*req.Customer:11:Stats1",
		"*prefix:*req.Customer:22:Stats1",
		"*string:*req.Account:1001:Stats1",

		// Stats2
		"*string:*req.Bank:BoA:Stats2",
		"*string:*req.Bank:CEC:Stats2",
		"*prefix:*req.Customer:11:Stats2",
		"*prefix:*req.Customer:22:Stats2",
	}
	sort.Strings(expectedIndexes)
	var replyIDx []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}

	// FLTR_Charger, FLTR_Charger12312 and FLTR_Charger4564 will be changed
	filter1 = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.RatingPlan",
					Values:  []string{"RP1"},
				},
				{
					Type:    utils.MetaPrefix,
					Element: "~*req.Subject",
					Values:  []string{"1001", "1002"},
				},
			},
		},
	}
	filter2 = &engine.FilterWithAPIOpts{
		Filter: &engine.Filter{
			Tenant: "cgrates.org",
			ID:     "FLTR_Charger2",
			Rules: []*engine.FilterRule{
				{
					Type:    utils.MetaString,
					Element: "~*req.Destination",
					Values:  []string{"randomID"},
				},
			},
		},
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetFilter, filter2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	arg = &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes = []string{
		// Stats1
		"*string:*req.RatingPlan:RP1:Stats1",
		"*prefix:*req.Subject:1001:Stats1",
		"*prefix:*req.Subject:1002:Stats1",
		"*string:*req.Destination:randomID:Stats1",

		// Stats2
		"*string:*req.RatingPlan:RP1:Stats2",
		"*prefix:*req.Subject:1001:Stats2",
		"*prefix:*req.Subject:1002:Stats2",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}
}

func testSetAndChangeFiltersOnProfiles(t *testing.T) {
	stat1 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant:      "cgrates.org",
			ID:          "Stats1",
			FilterIDs:   []string{},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}

	stat2 := &engine.StatQueueProfileWithAPIOpts{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "Stats2",
			FilterIDs: []string{
				"FLTR_Charger2",
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaACD,
				},
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Weight:       20,
			MinItems:     1,
		},
	}
	var reply string
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, stat1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	if err := tFIdxRpc.Call(utils.APIerSv1SetStatQueueProfile, stat2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	tPrfl1 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: "cgrates.org",
			ID:     "TEST_PROFILE1",
			FilterIDs: []string{"FLTR_Charger",
				"FLTR_Charger2"},
			MaxHits:  1,
			MinSleep: time.Duration(5 * time.Minute),
			Blocker:  false,
			Weight:   10.0,
			Async:    true,
		},
	}

	if err := tFIdxRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl1, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	arg := &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaStats,
	}
	expectedIndexes := []string{
		"*none:*any:*any:Stats1",
		"*string:*req.Destination:randomID:Stats2",
	}
	sort.Strings(expectedIndexes)
	var replyIDx []string
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}

	arg = &AttrGetFilterIndexes{
		Tenant:   "cgrates.org",
		ItemType: utils.MetaThresholds,
	}
	expectedIndexes = []string{
		"*string:*req.Destination:randomID:TEST_PROFILE1",
		"*string:*req.RatingPlan:RP1:TEST_PROFILE1",
		"*prefix:*req.Subject:1001:TEST_PROFILE1",
		"*prefix:*req.Subject:1002:TEST_PROFILE1",
	}
	sort.Strings(expectedIndexes)
	if err := tFIdxRpc.Call(utils.APIerSv1GetFilterIndexes, arg, &replyIDx); err != nil {
		t.Error(err)
	} else if sort.Strings(replyIDx); !reflect.DeepEqual(expectedIndexes, replyIDx) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedIndexes), utils.ToJSON(replyIDx))
	}
}

func testV1FIdxStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
