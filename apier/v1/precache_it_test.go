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

/* Need to investigate why add *default key Item1 in mongo
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
	precacheCfgPath   string
	precacheCfg       *config.CGRConfig
	precacheRPC       *rpc.Client
	precacheDataDir   = "/usr/share/cgrates"
	precacheDelay     int
	precacheConfigDIR string //run tests for specific configuration
)

var sTestsPrecache = []func(t *testing.T){
	testPrecacheInitCfg,
	testPrecacheResetDataDB,
	testPrecacheStartEngine,
	testPrecacheRpcConn,
	testPrecacheGetCacheStatsBeforeLoad,
	testPrecacheFromFolder,
	testPrecacheRestartEngine,
	testPrecacheGetItemIDs,
	testPrecacheGetCacheStatsAfterRestart,
	testPrecacheKillEngine,
}

func TestPrecacheITMySql(t *testing.T) {
	precacheConfigDIR = "tutmysql"
	for _, stest := range sTestsPrecache {
		t.Run(precacheConfigDIR, stest)
	}
}

func TestPrecacheITMongo(t *testing.T) {
	precacheConfigDIR = "tutmongo"
	for _, stest := range sTestsPrecache {
		t.Run(precacheConfigDIR, stest)
	}
}

func testPrecacheInitCfg(t *testing.T) {
	var err error
	precacheCfgPath = path.Join(precacheDataDir, "conf", "samples", precacheConfigDIR)
	precacheCfg, err = config.NewCGRConfigFromFolder(precacheCfgPath)
	if err != nil {
		t.Error(err)
	}
	precacheCfg.DataFolderPath = precacheDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(precacheCfg)
	switch precacheConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		precacheDelay = 2000
	default:
		precacheDelay = 1000
	}
}

func testPrecacheResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(precacheCfg); err != nil {
		t.Fatal(err)
	}
}

func testPrecacheStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(precacheCfgPath, precacheDelay); err != nil {
		t.Fatal(err)
	}
}

func testPrecacheRpcConn(t *testing.T) {
	var err error
	precacheRPC, err = jsonrpc.Dial("tcp", precacheCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testPrecacheGetItemIDs(t *testing.T) {
	args := &engine.ArgsGetCacheItemIDs{
		CacheID: "*default",
	}
	var reply *[]string
	if err := precacheRPC.Call(utils.CacheSv1GetItemIDs, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testPrecacheGetCacheStatsBeforeLoad(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	cacheIDs := []string{}
	expectedStats := &map[string]*ltcache.CacheStats{
		"*default": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"account_action_plans": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"action_plans": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"action_triggers": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"actions": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"aliases": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"attribute_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"attribute_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"attribute_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"cdr_stats": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"derived_chargers": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"destinations": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"event_resources": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"filters": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"lcr_rules": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"rating_plans": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"rating_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resource_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resource_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resource_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resources": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"reverse_aliases": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"reverse_destinations": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"shared_groups": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"stat_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"stat_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"statqueue_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"statqueues": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"supplier_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"supplier_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"supplier_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"threshold_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"threshold_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"threshold_profiles": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"thresholds": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"timings": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
	}
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, cacheIDs, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := precacheRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testPrecacheRestartEngine(t *testing.T) {
	time.Sleep(2 * time.Second)
	if _, err := engine.StopStartEngine(precacheCfgPath, precacheDelay); err != nil {
		t.Fatal(err)
	}
	var err error
	precacheRPC, err = jsonrpc.Dial("tcp", precacheCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPrecacheGetCacheStatsAfterRestart(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	cacheIDs := []string{}
	expectedStats := &map[string]*ltcache.CacheStats{
		"*default": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"account_action_plans": &ltcache.CacheStats{
			Items:  5,
			Groups: 0,
		},
		"action_plans": &ltcache.CacheStats{
			Items:  4,
			Groups: 0,
		},
		"action_triggers": &ltcache.CacheStats{
			Items:  4, // expected to have 4 items
			Groups: 0,
		},
		"actions": &ltcache.CacheStats{
			Items:  9, // expected to have 9 items
			Groups: 0,
		},
		"aliases": &ltcache.CacheStats{
			Items:  1,
			Groups: 0,
		},
		"attribute_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"attribute_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"attribute_profiles": &ltcache.CacheStats{
			Items:  1,
			Groups: 0,
		},
		"cdr_stats": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"derived_chargers": &ltcache.CacheStats{
			Items:  1, // expected to have 1 item
			Groups: 0,
		},
		"destinations": &ltcache.CacheStats{
			Items:  8,
			Groups: 0,
		},
		"event_resources": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"filters": &ltcache.CacheStats{
			Items:  16, // expected to have 16 items
			Groups: 0,
		},
		"lcr_rules": &ltcache.CacheStats{
			Items:  5, // expected to have 5 items
			Groups: 0,
		},
		"rating_plans": &ltcache.CacheStats{
			Items:  4, // expected to have 4 items
			Groups: 0,
		},
		"rating_profiles": &ltcache.CacheStats{
			Items:  10, // expected to have 10 items
			Groups: 0,
		},
		"resource_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resource_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"resource_profiles": &ltcache.CacheStats{
			Items:  4,
			Groups: 0,
		},
		"resources": &ltcache.CacheStats{
			Items:  3, //expected to have 3 items
			Groups: 0,
		},
		"reverse_aliases": &ltcache.CacheStats{
			Items:  2,
			Groups: 0,
		},
		"reverse_destinations": &ltcache.CacheStats{
			Items:  10,
			Groups: 0,
		},
		"shared_groups": &ltcache.CacheStats{
			Items:  1,
			Groups: 0,
		},
		"stat_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"stat_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"statqueue_profiles": &ltcache.CacheStats{
			Items:  2,
			Groups: 0,
		},
		"statqueues": &ltcache.CacheStats{
			Items:  1, // expected to have 1 item
			Groups: 0,
		},
		"supplier_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"supplier_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"supplier_profiles": &ltcache.CacheStats{
			Items:  3, // expected to have 3 items
			Groups: 0,
		},
		"threshold_filter_indexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"threshold_filter_revindexes": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
		"threshold_profiles": &ltcache.CacheStats{
			Items:  9,
			Groups: 0,
		},
		"thresholds": &ltcache.CacheStats{
			Items:  7, // expected to have 7 items
			Groups: 0,
		},
		"timings": &ltcache.CacheStats{
			Items:  0,
			Groups: 0,
		},
	}
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, cacheIDs, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheKillEngine(t *testing.T) {
	if err := engine.KillEngine(precacheDelay); err != nil {
		t.Error(err)
	}
}
*/
