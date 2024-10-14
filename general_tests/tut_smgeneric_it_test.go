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

package general_tests

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	tutSMGCfgPath string
	tutSMGCfgDIR  string
	tutSMGCfg     *config.CGRConfig
	tutSMGRpc     *birpc.Client
	smgLoadInst   utils.LoadInstance // Share load information between tests

	sTestTutSMG = []func(t *testing.T){
		testTutSMGInitCfg,
		testTutSMGResetDataDb,
		testTutSMGResetStorDb,
		testTutSMGStartEngine,
		testTutSMGRpcConn,
		testTutSMGLoadTariffPlanFromFolder,
		testTutSMGCacheStats,
		testTutSMGStopCgrEngine,
	}
)

func TestTutSMG(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		tutSMGCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	//mongo and sql tutmongo tutmysql
	for _, stest := range sTestTutSMG {
		t.Run(tutSMGCfgDIR, stest)
	}
}

func testTutSMGInitCfg(t *testing.T) {
	tutSMGCfgPath = path.Join(*utils.DataDir, "conf", "samples", tutSMGCfgDIR)
	// Init config first
	var err error
	tutSMGCfg, err = config.NewCGRConfigFromPath(tutSMGCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testTutSMGResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testTutSMGResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTutSMGStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutSMGCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTutSMGRpcConn(t *testing.T) {
	tutSMGRpc = engine.NewRPCClient(t, tutSMGCfg.ListenCfg())
}

// Load the tariff plan, creating accounts and their balances
func testTutSMGLoadTariffPlanFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := tutSMGRpc.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &smgLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

// Check loaded stats
func testTutSMGCacheStats(t *testing.T) {
	var reply string
	if err := tutSMGRpc.Call(context.Background(), utils.CacheSv1LoadCache, utils.NewAttrReloadCacheWithOpts(), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
	// expectedStats := &utils.CacheStats{Destinations: 5, ReverseDestinations: 7, RatingPlans: 4, RatingProfiles: 5,
	// 	Actions: 9, ActionPlans: 4, AccountActionPlans: 5, SharedGroups: 1, ResourceProfiles: 3,
	// 	Resources: 3, StatQueues: 1, StatQueueProfiles: 1, Thresholds: 7, ThresholdProfiles: 7, Filters: 15,
	// 	SupplierProfiles: 3, AttributeProfiles: 2}
	var rcvStats map[string]*ltcache.CacheStats
	expectedStats := engine.GetDefaultEmptyCacheStats()
	expectedStats[utils.CacheDestinations].Items = 5
	expectedStats[utils.CacheReverseDestinations].Items = 7
	expectedStats[utils.CacheRatingPlans].Items = 4
	expectedStats[utils.CacheRatingProfiles].Items = 5
	expectedStats[utils.CacheActions].Items = 10
	expectedStats[utils.CacheActionPlans].Items = 4
	expectedStats[utils.CacheSharedGroups].Items = 1
	expectedStats[utils.CacheResourceProfiles].Items = 3
	expectedStats[utils.CacheResources].Items = 3
	expectedStats[utils.CacheStatQueues].Items = 1
	expectedStats[utils.CacheStatQueueProfiles].Items = 1
	expectedStats[utils.CacheThresholds].Items = 7
	expectedStats[utils.CacheThresholdProfiles].Items = 7
	expectedStats[utils.CacheFilters].Items = 15
	expectedStats[utils.CacheRouteProfiles].Items = 3
	expectedStats[utils.CacheAttributeProfiles].Items = 3
	expectedStats[utils.MetaDefault].Items = 1
	expectedStats[utils.CacheActionTriggers].Items = 1
	expectedStats[utils.CacheLoadIDs].Items = 34
	expectedStats[utils.CacheChargerProfiles].Items = 1
	expectedStats[utils.CacheRPCConnections].Items = 2
	expectedStats[utils.CacheTimings].Items = 14
	expectedStats[utils.CacheThresholdFilterIndexes].Items = 12
	expectedStats[utils.CacheThresholdFilterIndexes].Groups = 1
	expectedStats[utils.CacheStatFilterIndexes].Items = 2
	expectedStats[utils.CacheStatFilterIndexes].Groups = 1
	expectedStats[utils.CacheRouteFilterIndexes].Items = 6
	expectedStats[utils.CacheRouteFilterIndexes].Groups = 1
	expectedStats[utils.CacheResourceFilterIndexes].Items = 6
	expectedStats[utils.CacheResourceFilterIndexes].Groups = 1
	expectedStats[utils.CacheChargerFilterIndexes].Items = 1
	expectedStats[utils.CacheChargerFilterIndexes].Groups = 1
	expectedStats[utils.CacheAttributeFilterIndexes].Items = 5
	expectedStats[utils.CacheAttributeFilterIndexes].Groups = 3
	expectedStats[utils.CacheReverseFilterIndexes].Items = 15
	expectedStats[utils.CacheReverseFilterIndexes].Groups = 13
	if err := tutSMGRpc.Call(context.Background(), utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("Calling APIerSv2.CacheSv1 expected: %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

func testTutSMGStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
