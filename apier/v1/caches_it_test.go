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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	chcCfg     *config.CGRConfig
	chcRPC     *rpc.Client
	chcCfgPath string
	chcCfgDir  string
)

var sTestsCacheSV1 = []func(t *testing.T){
	testCacheSLoadConfig,
	testCacheSInitDataDb,
	testCacheSInitStorDb,
	testCacheSStartEngine,
	testCacheSRpcConn,
	testCacheSLoadTariffPlanFromFolder,
	testCacheSAfterLoadFromFolder,
	testCacheSFlush,
	testCacheSReload,
	testCacheSGetItemIDs,
	testCacheSHasItem,
	testCacheSGetItemExpiryTime,
	testCacheSReloadCache,
	testCacheSRemoveItem,
	testCacheSClear,
	testCacheSReload,
	testCacheSPrecacheStatus,
	testCacheSPing,
	testCacheSStopEngine,
}

// Test start here
func TestCacheSv1ITMySQL(t *testing.T) {
	chcCfgDir = "tutmysql"
	for _, stest := range sTestsCacheSV1 {
		t.Run(chcCfgDir, stest)
	}
}

func TestCacheSv1ITMongo(t *testing.T) {
	chcCfgDir = "tutmongo"
	for _, stest := range sTestsCacheSV1 {
		t.Run(chcCfgDir, stest)
	}
}

func testCacheSLoadConfig(t *testing.T) {
	var err error
	chcCfgPath = path.Join(*dataDir, "conf", "samples", "precache", chcCfgDir)
	if chcCfg, err = config.NewCGRConfigFromPath(chcCfgPath); err != nil {
		t.Error(err)
	}
}

func testCacheSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testCacheSInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func testCacheSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCacheSRpcConn(t *testing.T) {
	var err error
	chcRPC, err = jsonrpc.Dial("tcp", chcCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
}

func testCacheSLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := chcRPC.Call(utils.ApierV1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testCacheSAfterLoadFromFolder(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	expStats[utils.CacheAccountActionPlans].Items = 13
	expStats[utils.CacheActionPlans].Items = 7
	expStats[utils.CacheActions].Items = 6
	expStats[utils.CacheDestinations].Items = 3
	expStats[utils.CacheLoadIDs].Items = 14
	if err := chcRPC.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, \n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := chcRPC.Call(utils.CacheSv1LoadCache, utils.AttrReloadCache{}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	expStats[utils.CacheActionTriggers].Items = 1
	expStats[utils.CacheActions].Items = 13
	expStats[utils.CacheAttributeProfiles].Items = 1
	expStats[utils.CacheFilters].Items = 15
	expStats[utils.CacheRatingPlans].Items = 5
	expStats[utils.CacheRatingProfiles].Items = 5
	expStats[utils.CacheResourceProfiles].Items = 3
	expStats[utils.CacheResources].Items = 3
	expStats[utils.CacheReverseDestinations].Items = 5
	expStats[utils.CacheStatQueueProfiles].Items = 1
	expStats[utils.CacheStatQueues].Items = 1
	expStats[utils.CacheSupplierProfiles].Items = 2
	expStats[utils.CacheThresholdProfiles].Items = 1
	expStats[utils.CacheThresholds].Items = 1
	expStats[utils.CacheLoadIDs].Items = 20

	if err := chcRPC.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v,\n received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheSFlush(t *testing.T) {
	reply := ""
	if err := chcRPC.Call(utils.CacheSv1FlushCache, utils.AttrReloadCache{FlushAll: true}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := chcRPC.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheSReload(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	reply := ""
	// Simple test that command is executed without errors
	if err := chcRPC.Call(utils.CacheSv1LoadCache, utils.AttrReloadCache{}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	expStats[utils.CacheAccountActionPlans].Items = 13
	expStats[utils.CacheActionPlans].Items = 7
	expStats[utils.CacheActions].Items = 6
	expStats[utils.CacheDestinations].Items = 3
	expStats[utils.CacheActionTriggers].Items = 1
	expStats[utils.CacheActions].Items = 13
	expStats[utils.CacheAttributeProfiles].Items = 1
	expStats[utils.CacheFilters].Items = 15
	expStats[utils.CacheRatingPlans].Items = 5
	expStats[utils.CacheRatingProfiles].Items = 5
	expStats[utils.CacheResourceProfiles].Items = 3
	expStats[utils.CacheResources].Items = 3
	expStats[utils.CacheReverseDestinations].Items = 5
	expStats[utils.CacheStatQueueProfiles].Items = 1
	expStats[utils.CacheStatQueues].Items = 1
	expStats[utils.CacheSupplierProfiles].Items = 2
	expStats[utils.CacheThresholdProfiles].Items = 1
	expStats[utils.CacheThresholds].Items = 1
	expStats[utils.CacheLoadIDs].Items = 20

	if err := chcRPC.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheSGetItemIDs(t *testing.T) {
	var rcvKeys []string
	var expKeys []string
	argsAPI := utils.ArgsGetCacheItemIDs{
		CacheID:      utils.CacheThresholdProfiles,
		ItemIDPrefix: "NotExistent",
	}
	if err := chcRPC.Call(utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ", utils.ErrNotFound, err.Error(), rcvKeys)
	}

	expKeys = []string{"cgrates.org:Threshold1"}
	argsAPI = utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheThresholdProfiles,
	}
	if err := chcRPC.Call(utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Fatalf("Got error on ApierV1.GetCacheStats: %s ", err.Error())
	}
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
}

func testCacheSHasItem(t *testing.T) {
	var reply bool
	var expected bool
	argsAPI := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "NotExistent",
	}
	if err := chcRPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}

	expected = true
	argsAPI = utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "cgrates.org:Threshold1",
	}
	if err := chcRPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testCacheSGetItemExpiryTime(t *testing.T) {
	var reply time.Time
	var expected time.Time
	argsAPI := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "NotExistent",
	}
	if err := chcRPC.Call(utils.CacheSv1GetItemExpiryTime, argsAPI, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ", utils.ErrNotFound, err.Error(), reply)
	}

	// expected = true
	argsAPI = utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "cgrates.org:Threshold1",
	}
	if err := chcRPC.Call(utils.CacheSv1GetItemExpiryTime, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
}

func testCacheSReloadCache(t *testing.T) {
	reply := ""
	arc := new(utils.AttrReloadCache)
	if err := chcRPC.Call(utils.CacheSv1ReloadCache, arc, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}

func testCacheSRemoveItem(t *testing.T) {
	var reply bool
	argsAPI := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "cgrates.org:Threshold1",
	}
	if err := chcRPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", true, reply)
	}
	var remReply string
	if err := chcRPC.Call(utils.CacheSv1RemoveItem, argsAPI, &remReply); err != nil {
		t.Error(err)
	} else if remReply != utils.OK {
		t.Errorf("Expected: %v , received:%v", utils.OK, remReply)
	}
	if err := chcRPC.Call(utils.CacheSv1HasItem, argsAPI, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: %v , received:%v", false, reply)
	}
}

func testCacheSClear(t *testing.T) {
	reply := ""
	if err := chcRPC.Call(utils.CacheSv1Clear, nil, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := chcRPC.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheSPrecacheStatus(t *testing.T) {
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
	}

	if err := chcRPC.Call(utils.CacheSv1PrecacheStatus, nil, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", utils.ToJSON(expected), utils.ToJSON(reply))
	}
}

func testCacheSPing(t *testing.T) {
	var reply string
	expected := utils.Pong
	if err := chcRPC.Call(utils.CacheSv1Ping, nil, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expected, reply) {
		t.Errorf("Expected: %v , received:%v", utils.ToJSON(expected), utils.ToJSON(reply))
	}

}

func testCacheSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
