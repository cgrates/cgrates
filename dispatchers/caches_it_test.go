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
	testDspChcPingEmptyCGREventWIthArgDispatcher,
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
	if err := dispEngine.RPC.Call(utils.CacheSv1Ping, &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspChcPingEmptyCGREventWIthArgDispatcher(t *testing.T) {
	var reply string
	expected := "MANDATORY_IE_MISSING: [APIKey]"
	if err := dispEngine.RPC.Call(utils.CacheSv1Ping,
		&utils.CGREventWithArgDispatcher{}, &reply); err == nil || err.Error() != expected {
		t.Errorf("Expected %+v, received %+v", expected, err)
	}
}

func testDspChcLoadAfterFolder(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	expStats[utils.CacheActions].Items = 1
	expStats[utils.CacheDestinations].Items = 4
	expStats[utils.CacheLoadIDs].Items = 17
	expStats[utils.CacheRPCConnections].Items = 2
	args := utils.AttrCacheIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := dispEngine.RPC.Call(utils.CacheSv1LoadCache, utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	expStats[utils.CacheAccountActionPlans].Items = 3
	expStats[utils.CacheActionPlans].Items = 1
	expStats[utils.CacheActionPlans].Items = 1
	expStats[utils.CacheActions].Items = 2
	expStats[utils.CacheAttributeProfiles].Items = 10
	expStats[utils.CacheChargerProfiles].Items = 2
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
	expStats[utils.CacheLoadIDs].Items = 20
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, &args, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
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
		utils.CacheDispatcherHosts:         utils.MetaReady,
		utils.CacheDiameterMessages:        utils.MetaReady,
		utils.CacheAttributeFilterIndexes:  utils.MetaReady,
		utils.CacheResourceFilterIndexes:   utils.MetaReady,
		utils.CacheStatFilterIndexes:       utils.MetaReady,
		utils.CacheThresholdFilterIndexes:  utils.MetaReady,
		utils.CacheSupplierFilterIndexes:   utils.MetaReady,
		utils.CacheChargerFilterIndexes:    utils.MetaReady,
		utils.CacheDispatcherFilterIndexes: utils.MetaReady,
		utils.CacheLoadIDs:                 utils.MetaReady,
		utils.CacheCDRIDs:                  utils.MetaReady,
		utils.CacheClosedSessions:          utils.MetaReady,
		utils.CacheDispatcherRoutes:        utils.MetaReady,
		utils.CacheEventResources:          utils.MetaReady,
		utils.CacheRPCConnections:          utils.MetaReady,
		utils.CacheRPCResponses:            utils.MetaReady,
		utils.CacheRatingProfilesTmp:       utils.MetaReady,
	}

	if err := dispEngine.RPC.Call(utils.CacheSv1PrecacheStatus, utils.AttrCacheIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , \n received:%v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testDspChcGetItemIDs(t *testing.T) {
	var rcvKeys []string
	expKeys := []string{"cgrates.org:DEFAULT", "cgrates.org:Raw"}
	argsAPI := utils.ArgsGetCacheItemIDsWithArgDispatcher{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheChargerProfiles,
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
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
	argsAPI := utils.ArgsGetCacheItemWithArgDispatcher{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
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
	argsAPI := utils.ArgsGetCacheItemWithArgDispatcher{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}
	if err := dispEngine.RPC.Call(utils.CacheSv1GetItemExpiryTime, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testDspChcReloadCache(t *testing.T) {
	reply := ""
	if err := dispEngine.RPC.Call(utils.CacheSv1ReloadCache, utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
	}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testDspChcRemoveItem(t *testing.T) {
	var reply bool
	argsAPI := utils.ArgsGetCacheItemWithArgDispatcher{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheChargerProfiles,
			ItemID:  "cgrates.org:DEFAULT",
		},
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
		},
		TenantArg: utils.TenantArg{
			Tenant: "cgrates.org",
		},
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
	if err := dispEngine.RPC.Call(utils.CacheSv1Clear, utils.AttrCacheIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
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
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, utils.AttrCacheIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
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
	if err := dispEngine.RPC.Call(utils.CacheSv1FlushCache, utils.AttrReloadCacheWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
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
	if err := dispEngine.RPC.Call(utils.CacheSv1GetCacheStats, utils.AttrCacheIDsWithArgDispatcher{
		ArgDispatcher: &utils.ArgDispatcher{
			APIKey: utils.StringPointer("chc12345"),
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
