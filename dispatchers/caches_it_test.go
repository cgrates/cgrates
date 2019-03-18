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
	testDspChcFlush,
}

//Test start here
func TestDspCacheSv1TMySQL(t *testing.T) {
	testDsp(t, sTestsDspChc, "TestDspCacheSv1", "all", "all2", "attributes", "dispatchers", "tutorial", "oldtutorial", "dispatchers")
}

func TestDspCacheSv1Mongo(t *testing.T) {
	testDsp(t, sTestsDspChc, "TestDspCacheSv1", "all", "all2", "attributes_mongo", "dispatchers_mongo", "tutorial", "oldtutorial", "dispatchers")
}

func testDspChcPing(t *testing.T) {
	var reply string
	if err := allEngine.RCP.Call(utils.CacheSv1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if dispEngine.RCP == nil {
		t.Fatal(dispEngine.RCP)
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1Ping, &CGREvWithApiKey{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
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
	expStats[utils.CacheAccountActionPlans].Items = 3
	expStats[utils.CacheActionPlans].Items = 1
	expStats[utils.CacheDestinations].Items = 4
	args := AttrCacheIDsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1GetCacheStats, args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := dispEngine.RCP.Call(utils.CacheSv1LoadCache, AttrReloadCacheWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	expStats[utils.CacheActions].Items = 2
	expStats[utils.CacheAttributeProfiles].Items = 9
	expStats[utils.CacheChargerProfiles].Items = 1
	expStats[utils.CacheFilters].Items = 7
	expStats[utils.CacheRatingPlans].Items = 5
	expStats[utils.CacheRatingProfiles].Items = 4
	expStats[utils.CacheResourceProfiles].Items = 1
	expStats[utils.CacheResources].Items = 1
	expStats[utils.CacheReverseDestinations].Items = 4
	expStats[utils.CacheStatQueueProfiles].Items = 2
	expStats[utils.CacheStatQueues].Items = 2
	expStats[utils.CacheSupplierProfiles].Items = 3
	expStats[utils.CacheThresholdProfiles].Items = 2
	expStats[utils.CacheThresholds].Items = 2

	if err := dispEngine.RCP.Call(utils.CacheSv1GetCacheStats, &args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testDspChcPrecacheStatus(t *testing.T) {
	var reply map[string]string
	expected := map[string]string{
		utils.CacheDestinations:            utils.MetaReady,
		utils.CacheReverseDestinations:     utils.MetaReady,
		utils.CacheRatingPlans:             utils.MetaReady,
		utils.CacheRatingProfiles:          utils.MetaReady,
		utils.CacheActions:                 utils.MetaReady,
		utils.CacheActionPlans:             utils.MetaReady,
		utils.CacheAccountActionPlans:      utils.MetaReady,
		utils.CacheActionTriggers:          utils.MetaReady,
		utils.CacheSharedGroups:            utils.MetaReady,
		utils.CacheResourceProfiles:        utils.MetaReady,
		utils.CacheResources:               utils.MetaReady,
		utils.CacheEventResources:          utils.MetaReady,
		utils.CacheTimings:                 utils.MetaReady,
		utils.CacheStatQueueProfiles:       utils.MetaReady,
		utils.CacheStatQueues:              utils.MetaReady,
		utils.CacheThresholdProfiles:       utils.MetaReady,
		utils.CacheThresholds:              utils.MetaReady,
		utils.CacheFilters:                 utils.MetaReady,
		utils.CacheSupplierProfiles:        utils.MetaReady,
		utils.CacheAttributeProfiles:       utils.MetaReady,
		utils.CacheChargerProfiles:         utils.MetaReady,
		utils.CacheDispatcherProfiles:      utils.MetaReady,
		utils.CacheDiameterMessages:        utils.MetaReady,
		utils.CacheAttributeFilterIndexes:  utils.MetaReady,
		utils.CacheResourceFilterIndexes:   utils.MetaReady,
		utils.CacheStatFilterIndexes:       utils.MetaReady,
		utils.CacheThresholdFilterIndexes:  utils.MetaReady,
		utils.CacheSupplierFilterIndexes:   utils.MetaReady,
		utils.CacheChargerFilterIndexes:    utils.MetaReady,
		utils.CacheDispatcherFilterIndexes: utils.MetaReady,
	}

	if err := dispEngine.RCP.Call(utils.CacheSv1PrecacheStatus, AttrCacheIDsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspChcGetItemIDs(t *testing.T) {
	var rcvKeys []string
	expKeys := []string{"cgrates.org:DEFAULT"}
	argsAPI := ArgsGetCacheItemIDsWithApiKey{
		ArgsGetCacheItemIDs: engine.ArgsGetCacheItemIDs{
			CacheID: utils.CacheChargerProfiles,
		},
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Fatalf("Got error on ApierV1.GetCacheStats: %s ", err.Error())
	}
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
}

func testDspChcHasItem(t *testing.T) {
	var reply bool
	expected := true
	argsAPI := ArgsGetCacheItemWithApiKey{
		ArgsGetCacheItem: engine.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testDspChcGetItemExpiryTime(t *testing.T) {
	var reply time.Time
	var expected time.Time
	argsAPI := ArgsGetCacheItemWithApiKey{
		ArgsGetCacheItem: engine.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1GetItemExpiryTime, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testDspChcReloadCache(t *testing.T) {
	reply := ""
	if err := dispEngine.RCP.Call(utils.CacheSv1ReloadCache, AttrReloadCacheWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testDspChcRemoveItem(t *testing.T) {
	var reply bool
	argsAPI := ArgsGetCacheItemWithApiKey{
		ArgsGetCacheItem: engine.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", true, reply)
	}
	var remReply string
	if err := dispEngine.RCP.Call(utils.CacheSv1RemoveItem, argsAPI, &remReply); err != nil {
		t.Error(err)
	} else if remReply != utils.OK {
		t.Errorf("Expected: %v , received:%v", utils.OK, remReply)
	}
	if err := dispEngine.RCP.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: %v , received:%v", false, reply)
	}
}

func testDspChcClear(t *testing.T) {
	reply := ""
	if err := dispEngine.RCP.Call(utils.CacheSv1Clear, AttrCacheIDsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := dispEngine.RCP.Call(utils.CacheSv1GetCacheStats, AttrCacheIDsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testDspChcFlush(t *testing.T) {
	reply := ""
	if err := dispEngine.RCP.Call(utils.CacheSv1FlushCache, AttrReloadCacheWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
		AttrReloadCache: utils.AttrReloadCache{
			FlushAll: true,
		},
	}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := dispEngine.RCP.Call(utils.CacheSv1GetCacheStats, AttrCacheIDsWithApiKey{
		DispatcherResource: DispatcherResource{
			APIKey: "chc12345",
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}
