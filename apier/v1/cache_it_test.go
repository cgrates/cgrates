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
	testCacheLoadConfig,
	testCacheInitDataDb,
	testCacheInitStorDb,
	testCacheStartEngine,
	testCacheRpcConn,
	testCacheLoadTariffPlanFromFolder,
	testCacheAfterLoadFromFolder,
	testCacheFlush,
	testCacheReload,
	// testCacheReloadCache,
	testCacheGetCacheKeys,

	testCacheStopEngine,
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

func testCacheLoadConfig(t *testing.T) {
	var err error
	chcCfgPath = path.Join(*dataDir, "conf", "samples", chcCfgDir)
	if chcCfg, err = config.NewCGRConfigFromPath(chcCfgPath); err != nil {
		t.Error(err)
	}
}

func testCacheInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testCacheInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Start engine
func testCacheStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(chcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCacheRpcConn(t *testing.T) {
	var err error
	chcRPC, err = jsonrpc.Dial("tcp", chcCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
}

func testCacheLoadTariffPlanFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testtp")}
	if err := chcRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
}

func testCacheAfterLoadFromFolder(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	expStats[utils.CacheAccountActionPlans].Items = 13
	expStats[utils.CacheActionPlans].Items = 7
	expStats[utils.CacheActions].Items = 6
	expStats[utils.CacheDestinations].Items = 3
	if err := chcRPC.Call("CacheSv1.GetCacheStats", nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
	reply := ""
	// Simple test that command is executed without errors
	if err := chcRPC.Call("CacheSv1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
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

	if err := chcRPC.Call("CacheSv1.GetCacheStats", nil, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheFlush(t *testing.T) {
	reply := ""
	if err := chcRPC.Call("CacheSv1.FlushCache", utils.AttrReloadCache{FlushAll: true}, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := chcRPC.Call("CacheSv1.GetCacheStats", nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheReloadCache(t *testing.T) {
	reply := ""
	arc := new(utils.AttrReloadCache)
	if err := chcRPC.Call("CacheSv1.ReloadCache", arc, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != "OK" {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()

	if err := chcRPC.Call("CacheSv1.GetCacheStats", nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling CacheSv1.GetCacheStats expected: %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

func testCacheReload(t *testing.T) {
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	reply := ""
	// Simple test that command is executed without errors
	if err := chcRPC.Call("CacheSv1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
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

	if err := chcRPC.Call("CacheSv1.GetCacheStats", nil, &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expStats, rcvStats) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheGetCacheKeys(t *testing.T) {

	expKeys := utils.ArgsCache{
		ThresholdProfileIDs: &[]string{"cgrates.org:Threshold1"},
	}
	var rcvKeys utils.ArgsCache
	argsAPI := utils.ArgsCacheKeys{
		ArgsCache: utils.ArgsCache{
			ThresholdProfileIDs: &[]string{},
			ResourceIDs:         &[]string{"NotExistent"},
		},
	}
	if err := chcRPC.Call("CacheSv1.GetCacheKeys", argsAPI, &rcvKeys); err != nil {
		t.Fatalf("Got error on ApierV1.GetCacheStats: %s ", err.Error())
	}
	if !reflect.DeepEqual(*expKeys.ThresholdProfileIDs, *rcvKeys.ThresholdProfileIDs) {
		t.Errorf("Expected: %+v, received: %+v", expKeys.ThresholdProfileIDs, rcvKeys.ThresholdProfileIDs)
	}
}

func testCacheStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
