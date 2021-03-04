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

package dispatchers

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var sTestsDspChc = []func(t *testing.T){
	testDspChcPing,
	testDspChcLoadAfterFolder,
	testDspChcPrecacheStatus,
	testDspChcGetItemIDs,
	testDspChcHasItem,
	testDspChcGetItemExpiryTime,
	testDspChcReloadCache,
	testDspChcRemoveItem,
	testDspChcClear,
}

//Test start here
func TestDspCacheSv1(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspChc, "TestDspCacheSv1", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func testDspChcPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.CacheSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RPC == nil {
		t.Fatal(dispEngine.RPC)
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1Ping, &utils.CGREvent{

		Tenant: "cgrates.org",

		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspChcLoadAfterFolder(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	expStats[utils.CacheActionPlans].Items = 1
	expStats[utils.CacheActions].Items = 1
	expStats[utils.CacheDestinations].Items = 4
	expStats[utils.CacheLoadIDs].Items = 17
	expStats[utils.CacheRPCConnections].Items = 2
	args := utils.AttrCacheIDsWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := dispEngine.RPC.Call(utils.CacheSv1LoadCache, utils.AttrReloadCacheWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant:    "cgrates.org",
		ArgsCache: utils.NewAttrReloadCacheWithOpts().ArgsCache,
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	expStats[utils.CacheActions].Items = 2
	expStats[utils.CacheAttributeProfiles].Items = 11
	expStats[utils.CacheChargerProfiles].Items = 2
	expStats[utils.CacheFilters].Items = 7
	expStats[utils.CacheRatingPlans].Items = 6
	expStats[utils.CacheRatingProfiles].Items = 7
	expStats[utils.CacheResourceProfiles].Items = 1
	expStats[utils.CacheResources].Items = 1
	expStats[utils.CacheReverseDestinations].Items = 4
	expStats[utils.CacheStatQueueProfiles].Items = 2
	expStats[utils.CacheStatQueues].Items = 2
	expStats[utils.CacheRouteProfiles].Items = 3
	expStats[utils.CacheThresholdProfiles].Items = 2
	expStats[utils.CacheThresholds].Items = 2
	expStats[utils.CacheLoadIDs].Items = 33
	expStats[utils.CacheTimings].Items = 10
	expStats[utils.CacheThresholdFilterIndexes].Items = 2
	expStats[utils.CacheThresholdFilterIndexes].Groups = 1
	expStats[utils.CacheStatFilterIndexes].Items = 7
	expStats[utils.CacheStatFilterIndexes].Groups = 1
	expStats[utils.CacheRouteFilterIndexes].Items = 3
	expStats[utils.CacheRouteFilterIndexes].Groups = 1
	expStats[utils.CacheResourceFilterIndexes].Items = 3
	expStats[utils.CacheResourceFilterIndexes].Groups = 1
	expStats[utils.CacheChargerFilterIndexes].Items = 1
	expStats[utils.CacheChargerFilterIndexes].Groups = 1
	expStats[utils.CacheAttributeFilterIndexes].Items = 11
	expStats[utils.CacheAttributeFilterIndexes].Groups = 4
	expStats[utils.CacheReverseFilterIndexes].Items = 8
	expStats[utils.CacheReverseFilterIndexes].Groups = 6
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, &args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testDspChcPrecacheStatus(t *testing.T) {
	var reply map[string]string
	expected := map[string]string{
		utils.CacheDestinations:                 utils.MetaReady,
		utils.CacheReverseDestinations:          utils.MetaReady,
		utils.CacheRatingPlans:                  utils.MetaReady,
		utils.CacheRatingProfiles:               utils.MetaReady,
		utils.CacheActions:                      utils.MetaReady,
		utils.CacheActionPlans:                  utils.MetaReady,
		utils.CacheAccountActionPlans:           utils.MetaReady,
		utils.CacheActionTriggers:               utils.MetaReady,
		utils.CacheSharedGroups:                 utils.MetaReady,
		utils.CacheResourceProfiles:             utils.MetaReady,
		utils.CacheResources:                    utils.MetaReady,
		utils.CacheTimings:                      utils.MetaReady,
		utils.CacheStatQueueProfiles:            utils.MetaReady,
		utils.CacheStatQueues:                   utils.MetaReady,
		utils.CacheThresholdProfiles:            utils.MetaReady,
		utils.CacheThresholds:                   utils.MetaReady,
		utils.CacheFilters:                      utils.MetaReady,
		utils.CacheRouteProfiles:                utils.MetaReady,
		utils.CacheAttributeProfiles:            utils.MetaReady,
		utils.CacheChargerProfiles:              utils.MetaReady,
		utils.CacheDispatcherProfiles:           utils.MetaReady,
		utils.CacheDispatcherHosts:              utils.MetaReady,
		utils.CacheDiameterMessages:             utils.MetaReady,
		utils.CacheAttributeFilterIndexes:       utils.MetaReady,
		utils.CacheResourceFilterIndexes:        utils.MetaReady,
		utils.CacheStatFilterIndexes:            utils.MetaReady,
		utils.CacheThresholdFilterIndexes:       utils.MetaReady,
		utils.CacheRouteFilterIndexes:           utils.MetaReady,
		utils.CacheChargerFilterIndexes:         utils.MetaReady,
		utils.CacheDispatcherFilterIndexes:      utils.MetaReady,
		utils.CacheRateProfilesFilterIndexes:    utils.MetaReady,
		utils.CacheRateFilterIndexes:            utils.MetaReady,
		utils.CacheRateProfiles:                 utils.MetaReady,
		utils.CacheLoadIDs:                      utils.MetaReady,
		utils.CacheCDRIDs:                       utils.MetaReady,
		utils.CacheClosedSessions:               utils.MetaReady,
		utils.CacheDispatcherRoutes:             utils.MetaReady,
		utils.CacheEventResources:               utils.MetaReady,
		utils.CacheRPCConnections:               utils.MetaReady,
		utils.CacheRPCResponses:                 utils.MetaReady,
		utils.CacheRatingProfilesTmp:            utils.MetaReady,
		utils.CacheUCH:                          utils.MetaReady,
		utils.CacheSTIR:                         utils.MetaReady,
		utils.CacheDispatcherLoads:              utils.MetaReady,
		utils.CacheDispatchers:                  utils.MetaReady,
		utils.CacheEventCharges:                 utils.MetaReady,
		utils.CacheReverseFilterIndexes:         utils.MetaReady,
		utils.CacheCapsEvents:                   utils.MetaReady,
		utils.CacheActionProfiles:               utils.MetaReady,
		utils.CacheActionProfilesFilterIndexes:  utils.MetaReady,
		utils.CacheAccountProfilesFilterIndexes: utils.MetaReady,
		utils.CacheAccountProfiles:              utils.MetaReady,

		utils.CacheAccounts:              utils.MetaReady,
		utils.CacheVersions:              utils.MetaReady,
		utils.CacheTBLTPTimings:          utils.MetaReady,
		utils.CacheTBLTPDestinations:     utils.MetaReady,
		utils.CacheTBLTPRates:            utils.MetaReady,
		utils.CacheTBLTPDestinationRates: utils.MetaReady,
		utils.CacheTBLTPRatingPlans:      utils.MetaReady,
		utils.CacheTBLTPRatingProfiles:   utils.MetaReady,
		utils.CacheTBLTPSharedGroups:     utils.MetaReady,
		utils.CacheTBLTPActions:          utils.MetaReady,
		utils.CacheTBLTPActionPlans:      utils.MetaReady,
		utils.CacheTBLTPActionTriggers:   utils.MetaReady,
		utils.CacheTBLTPAccountActions:   utils.MetaReady,
		utils.CacheTBLTPResources:        utils.MetaReady,
		utils.CacheTBLTPStats:            utils.MetaReady,
		utils.CacheTBLTPThresholds:       utils.MetaReady,
		utils.CacheTBLTPFilters:          utils.MetaReady,
		utils.CacheSessionCostsTBL:       utils.MetaReady,
		utils.CacheCDRsTBL:               utils.MetaReady,
		utils.CacheTBLTPRoutes:           utils.MetaReady,
		utils.CacheTBLTPAttributes:       utils.MetaReady,
		utils.CacheTBLTPChargers:         utils.MetaReady,
		utils.CacheTBLTPDispatchers:      utils.MetaReady,
		utils.CacheTBLTPDispatcherHosts:  utils.MetaReady,
		utils.CacheTBLTPRateProfiles:     utils.MetaReady,
		utils.MetaAPIBan:                 utils.MetaReady,
		utils.CacheTBLTPActionProfiles:   utils.MetaReady,
		utils.CacheTBLTPAccountProfiles:  utils.MetaReady,
		utils.CacheReplicationHosts:      utils.MetaReady,
	}

	if err := dispEngine.RPC.Call(utils.CacheSv1PrecacheStatus, utils.AttrCacheIDsWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , \n received:%v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspChcGetItemIDs(t *testing.T) {
	var rcvKeys []string
	expKeys := []string{"cgrates.org:DEFAULT", "cgrates.org:Raw"}
	argsAPI := utils.ArgsGetCacheItemIDsWithOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheChargerProfiles,
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Fatalf("Got error on APIerSv1.GetCacheStats: %s ", err.Error())
	}
	sort.Strings(rcvKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
}

func testDspChcHasItem(t *testing.T) {
	var reply bool
	expected := true
	argsAPI := utils.ArgsGetCacheItemWithOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testDspChcGetItemExpiryTime(t *testing.T) {
	var reply time.Time
	var expected time.Time
	argsAPI := utils.ArgsGetCacheItemWithOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1GetItemExpiryTime, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testDspChcReloadCache(t *testing.T) {
	reply := ""
	if err := dispEngine.RPC.Call(utils.CacheSv1ReloadCache, &utils.AttrReloadCacheWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testDspChcRemoveItem(t *testing.T) {
	var reply bool
	argsAPI := utils.ArgsGetCacheItemWithOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", true, reply)
	}
	var remReply string
	if err := dispEngine.RPC.Call(utils.CacheSv1RemoveItem, argsAPI, &remReply); err != nil {
		t.Error(err)
	} else if remReply != utils.OK {
		t.Errorf("Expected: %v , received:%v", utils.OK, remReply)
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: %v , received:%v", false, reply)
	}
}

func testDspChcClear(t *testing.T) {
	reply := ""
	if err := dispEngine.RPC.Call(utils.CacheSv1Clear, utils.AttrCacheIDsWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, utils.AttrCacheIDsWithOpts{
		Opts: map[string]interface{}{
			utils.OptsAPIKey: "chc12345",
		},
		Tenant: "cgrates.org",
	}, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}
