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

var (
	tFIdxHRpc *rpc.Client

	sTestsFilterIndexesSHealth = []func(t *testing.T){
		testV1FIdxHLoadConfig,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHStartEngine,
		testV1FIdxHRpcConn,

		testV1FIdxHLoadFromFolderTutorial2,
		testV1FIdxHAccountActionPlansHealth,
		testV1FIdxHReverseDestinationHealth,
		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,

		testV1FIdxHLoadFromFolderTutorial,
		testV1FIdxGetThresholdsIndexesHealth,
		testV1FIdxGetResourcesIndexesHealth,
		testV1FIdxGetStatsIndexesHealth,
		testV1FIdxGetSupplierIndexesHealth,
		testV1FIdxGetChargersIndexesHealth,
		testV1FIdxGetAttributesIndexesHealth,
		testV1FIdxCacheClear,

		testV1FIdxHdxInitDataDb,
		testV1FIdxHResetStorDb,
		testV1FIdxHLoadFromFolderDispatchers,
		testV1FIdxHGetDispatchersIndexesHealth,

		testV1FIdxHStopEngine,
	}
)

// Test start here
func TestFIdxHealthIT(t *testing.T) {
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
	for _, stest := range sTestsFilterIndexesSHealth {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxHLoadConfig(t *testing.T) {
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	var err error
	if tSv1Cfg, err = config.NewCGRConfigFromPath(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxHdxInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxHResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxHRpcConn(t *testing.T) {
	var err error
	tFIdxHRpc, err = newRPCClient(tSv1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxHLoadFromFolderTutorial2(t *testing.T) {
	var reply string
	if err := tFIdxHRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxHAccountActionPlansHealth(t *testing.T) {
	var reply engine.AccountActionPlanIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAccountActionPlansIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.AccountActionPlanIHReply{
		MissingAccountActionPlans: map[string][]string{},
		BrokenReferences:          map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxHReverseDestinationHealth(t *testing.T) {
	var reply engine.ReverseDestinationsIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetReverseDestinationsIndexHealth, engine.IndexHealthArgsWith2Ch{
		IndexCacheLimit:  -1,
		ObjectCacheLimit: -1,
	}, &reply); err != nil {
		t.Error(err)
	}
	exp := engine.ReverseDestinationsIHReply{
		MissingReverseDestinations: map[string][]string{},
		BrokenReferences:           map[string][]string{},
	}
	if !reflect.DeepEqual(exp, reply) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(exp), utils.ToJSON(reply))
	}
}

func testV1FIdxCacheClear(t *testing.T) {
	var reply string
	if err := tFIdxHRpc.Call(utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithArgDispatcher{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testV1FIdxHLoadFromFolderTutorial(t *testing.T) {
	var reply string
	if err := tFIdxHRpc.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: nil,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Reply: ", reply)
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxGetThresholdsIndexesHealth(t *testing.T) {
	// set another threshold profile different than the one from tariffplan
	tPrfl = &engine.ThresholdWithCache{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant: tenant,
			ID:     "TEST_PROFILE1",
			FilterIDs: []string{"*string:~*req.Account:1004",
				"*prefix:~*opts.Destination:+442",
				"*prefix:~*opts.Destination:+554"},
			MaxHits:   1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1", "ACT_2"},
			Async:     true,
		},
	}

	var rplyok string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetThresholdProfile, tPrfl, &rplyok); err != nil {
		t.Error(err)
	} else if rplyok != utils.OK {
		t.Error("Unexpected reply returned", rplyok)
	}

	// check all the indexes for thresholds
	expiIdx := []string{
		"*string:~*req.Account:1002:THD_ACNT_1002",
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*string:~*req.Account:1004:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+442:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+554:TEST_PROFILE1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expiIdx)
		if !reflect.DeepEqual(expiIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expiIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	args := &engine.IndexHealthArgsWith3Ch{}
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	var rply *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetThresholdsIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "THD_ACNT_1002",
		}, &rplyok); err != nil {
		t.Error(err)
	} else if rplyok != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	// check all the indexes for thresholds
	expiIdx = []string{
		"*string:~*req.Account:1001:THD_ACNT_1001",
		"*string:~*req.Account:1004:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+442:TEST_PROFILE1",
		"*prefix:~*opts.Destination:+554:TEST_PROFILE1",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaThresholds,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expiIdx)
		if !reflect.DeepEqual(expiIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expiIdx, result)
		}
	}
	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	expRPly = &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetThresholdsIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}
}

func testV1FIdxGetResourcesIndexesHealth(t *testing.T) {
	// set another resource profile different than the one from tariffplan
	var reply string
	rlsPrf := &ResourceWithCache{
		ResourceProfile: &engine.ResourceProfile{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
			FilterIDs: []string{"*string:~*req.Account:1001",
				"*prefix:~*opts.Destination:+334;+122"},
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
			},
			UsageTTL:          -1,
			Limit:             7,
			AllocationMessage: "",
			Stored:            true,
			Weight:            10,
			ThresholdIDs:      []string{utils.META_NONE},
		},
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1SetResourceProfile, rlsPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// check all the indexes for resources
	expIdx := []string{
		"*string:~*req.Account:1001:ResGroup2",
		"*prefix:~*opts.Destination:+334:ResGroup2",
		"*prefix:~*opts.Destination:+122:ResGroup2",
		"*string:~*req.Account:1001:ResGroup1",
		"*string:~*req.Account:1002:ResGroup1",
		"*string:~*req.Account:1003:ResGroup1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaResources,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rply *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetResourcesIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveResourceProfile,
		utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "ResGroup2",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetResourcesIndexesHealth,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rply))
	}
}

func testV1FIdxGetStatsIndexesHealth(t *testing.T) {
	// set another stats profile different than the one from tariffplan
	statConfig = &engine.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: "cgrates.org",
			ID:     "TEST_STATPROFILE_1",
			FilterIDs: []string{"*string:~*req.OriginID:RandomID",
				"*prefix:~*opts.Destination:+332;+234"},
			QueueLength: 10,
			TTL:         10 * time.Second,
			Metrics: []*engine.MetricWithFilters{
				{
					MetricID: utils.MetaTCD,
				},
			},
			ThresholdIDs: []string{"*none"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var rply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Error("Unexpected reply returned", rply)
	}

	// check all the indexes for statsQueue
	expIdx := []string{
		"*string:~*req.OriginID:RandomID:TEST_STATPROFILE_1",
		"*prefix:~*opts.Destination:+332:TEST_STATPROFILE_1",
		"*prefix:~*opts.Destination:+234:TEST_STATPROFILE_1",
		"*string:~*req.Account:1001:Stats2",
		"*string:~*req.Account:1002:Stats2",
		"*string:~*req.RunID:*default:Stats2",
		"*string:~*req.Destination:1001:Stats2",
		"*string:~*req.Destination:1002:Stats2",
		"*string:~*req.Destination:1003:Stats2",
		"*string:~*req.Account:1003:Stats2_1",
		"*string:~*req.RunID:*default:Stats2_1",
		"*string:~*req.Destination:1001:Stats2_1",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaStats,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rplyFl *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetStatsIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Stats2",
		}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetStatsIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetSupplierIndexesHealth(t *testing.T) {
	// set another routes profile different than the one from tariffplan
	rPrf := &SupplierWithCache{
		SupplierProfile: &engine.SupplierProfile{
			Tenant:            tenant,
			ID:                "TEST_PROFILE1",
			FilterIDs:         []string{"*prefix:~*req.Destination:+23331576354"},
			Sorting:           "Sort1",
			SortingParameters: []string{"Param1", "Param2"},
			Suppliers: []*engine.Supplier{{
				ID:            "SPL1",
				RatingPlanIDs: []string{"RP1"},
				FilterIDs:     []string{"FLTR_1"},
				Weight:        20,
				Blocker:       false,
			}},
			Weight: 10,
		},
	}
	var reply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetSupplierProfile, rPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// check all the indexes for routes
	expIdx := []string{
		"*prefix:~*req.Destination:+23331576354:TEST_PROFILE1",
		"*string:~*req.Account:1001:SPL_ACNT_1001",
		"*string:~*req.Account:1002:SPL_ACNT_1002",
		"*string:~*req.Account:1003:SPL_ACNT_1003",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaSuppliers,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rplyFl *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetSuppliersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveSupplierProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "TEST_PROFILE1",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("UNexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetSuppliersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetChargersIndexesHealth(t *testing.T) {
	// set another charger profile different than the one from tariffplan
	chargerProfile := &ChargerWithCache{
		ChargerProfile: &engine.ChargerProfile{
			Tenant: "cgrates.org",
			ID:     "Default",
			FilterIDs: []string{"*string:~*req.Destination:+1442",
				"*prefix:~*opts.Accounts:1002;1004"},
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var reply string
	if err := tFIdxHRpc.Call(utils.APIerSv1SetChargerProfile, chargerProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	// those 2 charger object (*none:*any:*any index) are from tutorial2 tariffplan, so on imternal we must delete them by api
	if tSv1Cfg.DataDbCfg().DataDbType == utils.INTERNAL {
		var result string
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "CRG_RESELLER1",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "CGR_DEFAULT",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
	}

	// check all the indexes for chargers
	expIdx := []string{
		"*string:~*req.Destination:+1442:Default",
		"*prefix:~*opts.Accounts:1002:Default",
		"*prefix:~*opts.Accounts:1004:Default",
		"*none:*any:*any:DEFAULT",
		"*none:*any:*any:Raw",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaChargers,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rplyFl *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetChargersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveChargerProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "Raw",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	if err := tFIdxHRpc.Call(utils.APIerSv1GetChargersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxGetAttributesIndexesHealth(t *testing.T) {
	// Attributes.csv from tutorial tariffplan got lots of profiles, so we will not set another attribute for this test
	// check all the indexes for attributes
	// simpleauth context
	expIdx := []string{
		"*string:~*req.Account:1001:ATTR_1001_SIMPLEAUTH",
		"*string:~*req.Account:1002:ATTR_1002_SIMPLEAUTH",
		"*string:~*req.Account:1003:ATTR_1003_SIMPLEAUTH",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  "simpleauth",
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// this attr object (*none:*any:*any index) must be deleted with api
	if tSv1Cfg.DataDbCfg().DataDbType == utils.INTERNAL {
		var result string
		if err := tFIdxHRpc.Call(utils.APIerSv1RemoveAttributeProfile,
			&utils.TenantIDWithCache{
				Tenant: "cgrates.org",
				ID:     "ATTR_CRG_SUPPLIER1",
			}, &result); err != nil {
			t.Error(err)
		} else if result != utils.OK {
			t.Errorf("Unexpected reply returned")
		}
	}

	// *sessions context
	expIdx = []string{
		"*string:~*req.Account:1001:ATTR_1001_SESSIONAUTH",
		"*string:~*req.Account:1002:ATTR_1002_SESSIONAUTH",
		"*string:~*req.Account:1003:ATTR_1003_SESSIONAUTH",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  utils.MetaSessionS,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// *any context tenant: cgrates.org
	expIdx = []string{
		"*string:~*req.SubscriberId:1006:ATTR_ACC_ALIAS",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// *any context tenant: cgrates.com
	expIdx = []string{
		"*string:~*req.SubscriberId:1006:ATTR_TNT_ALIAS",
		"*string:~*req.Account:1001:ATTR_TNT_1001",
		"*string:~*req.Account:testDiamInitWithSessionDisconnect:ATTR_TNT_DISC",
	}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		Tenant:   "cgrates.com",
		ItemType: utils.MetaAttributes,
		Context:  utils.META_ANY,
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rplyFl *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetAttributesIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxHLoadFromFolderDispatchers(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "dispatchers")}
	if err := tFIdxHRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testV1FIdxHGetDispatchersIndexesHealth(t *testing.T) {
	// *any context
	expIdx := []string{
		"*none:*any:*any:PING1",
		"*string:~*req.EventName:UnexistedHost:PING2",
		"*string:~*req.EventName:Event1:EVENT1",
		"*string:~*req.EventName:RoundRobin:EVENT2",
		"*string:~*req.EventName:Random:EVENT3",
		"*string:~*req.EventName:Broadcast:EVENT4",
		"*string:~*req.EventName:Internal:EVENT5",
		// "*string:~*opts.*method:DispatcherSv1.GetProfilesForEvent:EVENT6",
		// "*string:~*opts.EventType:LoadDispatcher:EVENT7",
	}
	var result []string
	if err := tFIdxHRpc.Call(utils.APIerSv1GetFilterIndexes, &AttrGetFilterIndexes{
		ItemType: utils.MetaDispatchers,
		Context:  utils.META_ANY,
		Tenant:   "cgrates.org",
	}, &result); err != nil {
		t.Error(err)
	} else {
		sort.Strings(result)
		sort.Strings(expIdx)
		if !reflect.DeepEqual(expIdx, result) {
			t.Errorf("Expecting: %+v, received: %+v", expIdx, result)
		}
	}

	// all indexes are set and points to their objects correctly
	expRPly := &engine.FilterIHReply{
		// MissingObjects: []string{},
		MissingIndexes: map[string][]string{},
		BrokenIndexes:  map[string][]string{},
		MissingFilters: map[string][]string{},
	}
	args := &engine.IndexHealthArgsWith3Ch{}
	var rplyFl *engine.FilterIHReply
	if err := tFIdxHRpc.Call(utils.APIerSv1GetDispatchersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}

	var reply string
	// removing a profile + their indexes
	if err := tFIdxHRpc.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{
			Tenant: "cgrates.org",
			ID:     "PING2",
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//as we removed the object, the index specified is removed too, so the health of the indexes is fine
	args = &engine.IndexHealthArgsWith3Ch{}
	if err := tFIdxHRpc.Call(utils.APIerSv1GetDispatchersIndexesHealth,
		args, &rplyFl); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rplyFl, expRPly) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRPly), utils.ToJSON(rplyFl))
	}
}

func testV1FIdxHStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
