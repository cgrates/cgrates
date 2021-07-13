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
	precacheConfigDIR string //run tests for specific configuration

	sTestsPrecache = []func(t *testing.T){
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
)

func TestPrecacheIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		precacheConfigDIR = "tutmysql"
	case utils.MetaMongo:
		precacheConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsPrecache {
		t.Run(precacheConfigDIR, stest)
	}
}

func testPrecacheInitCfg(t *testing.T) {
	var err error
	precacheCfgPath = path.Join(precacheDataDir, "conf", "samples", "precache", precacheConfigDIR)
	precacheCfg, err = config.NewCGRConfigFromPath(precacheCfgPath)
	if err != nil {
		t.Error(err)
	}
	precacheCfg.DataFolderPath = precacheDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(precacheCfg)
}

func testPrecacheResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(precacheCfg); err != nil {
		t.Fatal(err)
	}
}

func testPrecacheStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(precacheCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testPrecacheRpcConn(t *testing.T) {
	var err error
	precacheRPC, err = newRPCClient(precacheCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testPrecacheGetItemIDs(t *testing.T) {
	args := &utils.ArgsGetCacheItemIDs{
		CacheID: utils.MetaDefault,
	}
	var reply *[]string
	if err := precacheRPC.Call(utils.CacheSv1GetItemIDs, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testPrecacheGetCacheStatsBeforeLoad(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	args := &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: []string{},
	}
	dfltStats := engine.GetDefaultEmptyCacheStats()
	expectedStats := &dfltStats
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "precache")}
	if err := precacheRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testPrecacheRestartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(precacheCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	precacheRPC, err = newRPCClient(precacheCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPrecacheGetCacheStatsAfterRestart(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	args := &utils.AttrCacheIDsWithArgDispatcher{
		CacheIDs: []string{},
	}
	expectedStats := &map[string]*ltcache.CacheStats{
		utils.MetaDefault: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheAccountActionPlans: {
			Items:  5,
			Groups: 0,
		},
		utils.CacheActionPlans: {
			Items:  4,
			Groups: 0,
		},
		utils.CacheActionTriggers: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheActions: {
			Items:  9,
			Groups: 0,
		},
		utils.CacheAttributeFilterIndexes: {
			Items:  2,
			Groups: 0,
		},
		utils.CacheAttributeProfiles: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheChargerFilterIndexes: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheChargerProfiles: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheDispatcherFilterIndexes: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheDispatcherProfiles: {
			Items:  6,
			Groups: 0,
		},
		utils.CacheDispatcherHosts: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheDispatcherRoutes: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheDestinations: {
			Items:  5,
			Groups: 0,
		},
		utils.CacheEventResources: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheFilters: {
			Items:  15,
			Groups: 0,
		},
		utils.CacheRatingPlans: {
			Items:  4,
			Groups: 0,
		},
		utils.CacheRatingProfiles: {
			Items:  5,
			Groups: 0,
		},
		utils.CacheResourceFilterIndexes: {
			Items:  6,
			Groups: 0,
		},
		utils.CacheResourceProfiles: {
			Items:  3,
			Groups: 0,
		},
		utils.CacheResources: {
			Items:  3,
			Groups: 0,
		},
		utils.CacheReverseDestinations: {
			Items:  7,
			Groups: 0,
		},
		utils.CacheRPCResponses: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheSharedGroups: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheStatFilterIndexes: {
			Items:  2,
			Groups: 0,
		},
		utils.CacheStatQueueProfiles: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheStatQueues: {
			Items:  1,
			Groups: 0,
		},
		utils.CacheSupplierFilterIndexes: {
			Items:  6,
			Groups: 0,
		},
		utils.CacheSupplierProfiles: {
			Items:  3,
			Groups: 0,
		},
		utils.CacheThresholdFilterIndexes: {
			Items:  10,
			Groups: 0,
		},
		utils.CacheThresholdProfiles: {
			Items:  7,
			Groups: 0,
		},
		utils.CacheThresholds: {
			Items:  7,
			Groups: 0,
		},
		utils.CacheTimings: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheDiameterMessages: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheClosedSessions: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheLoadIDs: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheRPCConnections: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheCDRIDs: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheRatingProfilesTmp: {
			Items:  0,
			Groups: 0,
		},
		utils.CacheReplicationHosts: {},
	}
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
