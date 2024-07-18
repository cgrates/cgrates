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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
)

var (
	tutSMGCfgPath string
	tutSMGCfgDIR  string
	tutSMGCfg     *config.CGRConfig
	tutSMGRpc     *birpc.Client

	sTestTutSMG = []func(t *testing.T){
		testTutSMGInitCfg,
		testTutSMGFlushDBs,

		testTutSMGStartEngine,
		testTutSMGRpcConn,
		testTutSMGLoadTariffPlanFromFolder,
		testTutSMGCacheStats,
		testTutSMGStopCgrEngine,
	}
)

func TestTutSMG(t *testing.T) {
	switch *dbType {
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
	tutSMGCfgPath = path.Join(*dataDir, "conf", "samples", tutSMGCfgDIR)
	// Init config first
	var err error
	tutSMGCfg, err = config.NewCGRConfigFromPath(context.Background(), tutSMGCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testTutSMGFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(tutSMGCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTutSMGStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutSMGCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTutSMGRpcConn(t *testing.T) {
	var err error
	tutSMGRpc, err = newRPCClient(tutSMGCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testTutSMGLoadTariffPlanFromFolder(t *testing.T) {
	caching := utils.MetaReload
	if tutSMGCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := tutSMGRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			// StopOnError: true,
			APIOpts: map[string]any{utils.MetaCache: caching},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
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
	expectedStats[utils.CacheAccountsFilterIndexes].Items = 1
	expectedStats[utils.CacheAccountsFilterIndexes].Groups = 1
	expectedStats[utils.CacheActionProfilesFilterIndexes].Items = 1
	expectedStats[utils.CacheActionProfilesFilterIndexes].Groups = 1
	expectedStats[utils.CacheActionProfiles].Items = 1
	expectedStats[utils.CacheRateFilterIndexes].Items = 2
	expectedStats[utils.CacheRateFilterIndexes].Groups = 2
	expectedStats[utils.CacheRateProfilesFilterIndexes].Items = 2
	expectedStats[utils.CacheRateProfilesFilterIndexes].Groups = 1
	expectedStats[utils.CacheRateProfiles].Items = 2
	expectedStats[utils.CacheActionProfiles].Items = 1
	expectedStats[utils.CacheResourceProfiles].Items = 1
	expectedStats[utils.CacheResources].Items = 1
	expectedStats[utils.CacheStatQueues].Items = 7
	expectedStats[utils.CacheStatQueueProfiles].Items = 7
	expectedStats[utils.CacheThresholds].Items = 1
	expectedStats[utils.CacheThresholdProfiles].Items = 1
	expectedStats[utils.CacheFilters].Items = 26
	expectedStats[utils.CacheRouteProfiles].Items = 12
	expectedStats[utils.CacheAttributeProfiles].Items = 8
	expectedStats[utils.MetaDefault].Items = 0
	expectedStats[utils.CacheLoadIDs].Items = 29
	expectedStats[utils.CacheChargerProfiles].Items = 3
	expectedStats[utils.CacheRPCConnections].Items = 1
	expectedStats[utils.CacheThresholdFilterIndexes].Items = 1
	expectedStats[utils.CacheThresholdFilterIndexes].Groups = 1
	expectedStats[utils.CacheStatFilterIndexes].Items = 7
	expectedStats[utils.CacheStatFilterIndexes].Groups = 1
	expectedStats[utils.CacheRouteFilterIndexes].Items = 16
	expectedStats[utils.CacheRouteFilterIndexes].Groups = 1
	expectedStats[utils.CacheResourceFilterIndexes].Items = 1
	expectedStats[utils.CacheResourceFilterIndexes].Groups = 1
	expectedStats[utils.CacheChargerFilterIndexes].Items = 1
	expectedStats[utils.CacheChargerFilterIndexes].Groups = 1
	expectedStats[utils.CacheRankingProfiles].Items = 0
	expectedStats[utils.CacheRankingProfiles].Groups = 0
	expectedStats[utils.CacheAttributeFilterIndexes].Items = 10
	expectedStats[utils.CacheAttributeFilterIndexes].Groups = 1
	expectedStats[utils.CacheReverseFilterIndexes].Items = 19
	expectedStats[utils.CacheReverseFilterIndexes].Groups = 16
	if err := tutSMGRpc.Call(context.Background(), utils.CacheSv1GetCacheStats, new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error("Got error on CacheSv1.GetCacheStats: ", err.Error())
	} else if !reflect.DeepEqual(expectedStats, rcvStats) {
		t.Errorf("expected: %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(rcvStats))
	}
}

func testTutSMGStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
