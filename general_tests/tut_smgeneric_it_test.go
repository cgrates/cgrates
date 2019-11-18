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

package general_tests

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
	tutSMGCfgPath string
	tutSMGCfg     *config.CGRConfig
	tutSMGRpc     *rpc.Client
	smgLoadInst   utils.LoadInstance // Share load information between tests

	sTestTutSMG = []func(t *testing.T){
		TestTutSMGInitCfg,
		TestTutSMGResetDataDb,
		TestTutSMGResetStorDb,
		TestTutSMGStartEngine,
		TestTutSMGRpcConn,
		TestTutSMGLoadTariffPlanFromFolder,
		TestTutSMGCacheStats,
		TestTutSMGStopCgrEngine,
	}
)

func TestTutSMG(t *testing.T) {
	for _, stest := range sTestTutSMG {
		t.Run("TestTutSMG", stest)
	}
}

func TestTutSMGInitCfg(t *testing.T) {
	tutSMGCfgPath = path.Join(*dataDir, "conf", "samples", "smgeneric")
	// Init config first
	var err error
	tutSMGCfg, err = config.NewCGRConfigFromPath(tutSMGCfgPath)
	if err != nil {
		t.Error(err)
	}
	tutSMGCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tutSMGCfg)
}

// Remove data in both rating and accounting db
func TestTutSMGResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestTutSMGResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestTutSMGStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutSMGCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestTutSMGRpcConn(t *testing.T) {
	var err error
	tutSMGRpc, err = jsonrpc.Dial("tcp", tutSMGCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestTutSMGLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := tutSMGRpc.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &smgLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func TestTutSMGCacheStats(t *testing.T) {
	var reply string
	if err := tutSMGRpc.Call("CacheSv1.LoadCache", utils.AttrReloadCache{}, &reply); err != nil {
		t.Error(err)
	} else if reply != "OK" {
		t.Error(reply)
	}
	// expectedStats := &utils.CacheStats{Destinations: 5, ReverseDestinations: 7, RatingPlans: 4, RatingProfiles: 5,
	// 	Actions: 9, ActionPlans: 4, AccountActionPlans: 5, SharedGroups: 1, ResourceProfiles: 3,
	// 	Resources: 3, StatQueues: 1, StatQueueProfiles: 1, Thresholds: 7, ThresholdProfiles: 7, Filters: 15,
	// 	SupplierProfiles: 3, AttributeProfiles: 1}
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()
	expectedStats[utils.CacheDestinations].Items = 5
	expectedStats[utils.CacheReverseDestinations].Items = 7
	expectedStats[utils.CacheRatingPlans].Items = 4
	expectedStats[utils.CacheRatingProfiles].Items = 5
	expectedStats[utils.CacheActions].Items = 9
	expectedStats[utils.CacheActionPlans].Items = 4
	expectedStats[utils.CacheAccountActionPlans].Items = 5
	expectedStats[utils.CacheSharedGroups].Items = 1
	expectedStats[utils.CacheResourceProfiles].Items = 3
	expectedStats[utils.CacheResources].Items = 3
	expectedStats[utils.CacheStatQueues].Items = 1
	expectedStats[utils.CacheStatQueueProfiles].Items = 1
	expectedStats[utils.CacheThresholds].Items = 7
	expectedStats[utils.CacheThresholdProfiles].Items = 7
	expectedStats[utils.CacheFilters].Items = 15
	expectedStats[utils.CacheSupplierProfiles].Items = 3
	expectedStats[utils.CacheAttributeProfiles].Items = 1
	expectedStats[utils.MetaDefault].Items = 1
	expectedStats[utils.CacheActionTriggers].Items = 1
	expectedStats[utils.CacheLoadIDs].Items = 20
	expectedStats[utils.CacheChargerProfiles].Items = 1
	if err := tutSMGRpc.Call(utils.CacheSv1GetCacheStats, nil, &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling ApierV2.CacheSv1 expected: %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

func TestTutSMGStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
