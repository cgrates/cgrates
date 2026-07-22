//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"flag"
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
	precacheCfgPath   string
	precacheCfg       *config.CGRConfig
	precacheRPC       *birpc.Client
	precacheConfigDIR string //run tests for specific configuration

	// use this flag to test the APIBan implementation for precache
	apiBan = flag.Bool("apiban", false, "used to control if we run the apiban tests")

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

// this tests may fail because of apiban limit( 5 requests per 2 minutes for an APIKey)
// if needed add more APIKeys
func TestPrecacheIT(t *testing.T) {
	switch *utils.DBType {
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
	if *apiBan {
		precacheConfigDIR += "_apiban"
	}
	for _, stest := range sTestsPrecache {
		t.Run(precacheConfigDIR, stest)
	}
}

func testPrecacheInitCfg(t *testing.T) {
	var err error
	precacheCfgPath = path.Join(*utils.DataDir, "conf", "samples", "precache", precacheConfigDIR)
	precacheCfg, err = config.NewCGRConfigFromPath(precacheCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testPrecacheResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(precacheCfg); err != nil {
		t.Fatal(err)
	}
}

func testPrecacheStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(precacheCfgPath, *utils.WaitRater); err != nil {
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
	if err := precacheRPC.Call(context.Background(), utils.CacheSv1GetItemIDs, args, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testPrecacheGetCacheStatsBeforeLoad(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	args := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: []string{},
	}
	dfltStats := engine.GetDefaultEmptyCacheStats()
	expectedStats := &dfltStats
	if err := precacheRPC.Call(context.Background(), utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "precache")}
	if err := precacheRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testPrecacheRestartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(precacheCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	precacheRPC, err = newRPCClient(precacheCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	time.Sleep(3*time.Second + 800*time.Millisecond) // let the *apiban cache to be populated
}

func testPrecacheGetCacheStatsAfterRestart(t *testing.T) {
	var reply *map[string]*ltcache.CacheStats
	args := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: []string{},
	}
	expectedStats := &map[string]*ltcache.CacheStats{
		utils.MetaDefault:             {},
		utils.CacheAccountActionPlans: {},
		utils.CacheActionPlans:        {Items: 4},
		utils.CacheActionTriggers:     {Items: 1},
		utils.CacheActions:            {Items: 9},
		utils.CacheAttributeFilterIndexes: {
			Items:  2,
			Groups: 2,
		},
		utils.CacheAttributeProfiles:       {Items: 1},
		utils.CacheChargerFilterIndexes:    {},
		utils.CacheChargerProfiles:         {},
		utils.CacheDispatcherFilterIndexes: {},
		utils.CacheDispatcherProfiles:      {Items: 6},
		utils.CacheDispatcherHosts:         {Items: 1},
		utils.CacheDispatcherRoutes:        {},
		utils.CacheDispatcherLoads:         {},
		utils.CacheDestinations:            {Items: 5},
		utils.CacheDispatchers:             {},
		utils.CacheEventResources:          {},
		utils.CacheEventIPs:                {},
		utils.CacheFilters:                 {Items: 15},
		utils.CacheRatingPlans:             {Items: 4},
		utils.CacheRatingProfiles:          {Items: 5},
		utils.CacheResourceFilterIndexes: {
			Items:  6,
			Groups: 1,
		},
		utils.CacheIPFilterIndexes:     {},
		utils.CacheResourceProfiles:    {Items: 3},
		utils.CacheResources:           {Items: 3},
		utils.CacheIPProfiles:          {},
		utils.CacheIPAllocations:       {},
		utils.CacheReverseDestinations: {Items: 7},
		utils.CacheRPCResponses:        {},
		utils.MetaSentryPeer:           {},
		utils.CacheSharedGroups:        {Items: 1},
		utils.CacheStatFilterIndexes: {
			Items:  2,
			Groups: 1,
		},
		utils.CacheStatQueueProfiles: {Items: 1},
		utils.CacheStatQueues:        {Items: 1},
		utils.CacheSTIR:              {},
		utils.CacheCapsEvents:        {},
		utils.CacheEventCharges:      {},
		utils.CacheRouteFilterIndexes: {
			Items:  6,
			Groups: 1,
		},
		utils.CacheRouteProfiles: {Items: 3},
		utils.CacheThresholdFilterIndexes: {
			Items:  12,
			Groups: 1,
		},
		utils.CacheThresholdProfiles:    {Items: 7},
		utils.CacheThresholds:           {Items: 7},
		utils.CacheTimings:              {},
		utils.CacheDiameterMessages:     {},
		utils.CacheClosedSessions:       {},
		utils.CacheLoadIDs:              {},
		utils.CacheRPCConnections:       {},
		utils.CacheCDRIDs:               {},
		utils.CacheRatingProfilesTmp:    {},
		utils.CacheUCH:                  {},
		utils.CacheReverseFilterIndexes: {},
		utils.MetaAPIBan:                {},
		utils.CacheReplicationHosts:     {},
		utils.CacheRadiusPackets:        {},
		utils.CacheRankings:             {},
		utils.CacheRankingProfiles:      {},
		utils.CacheTrends:               {},
		utils.CacheTrendProfiles:        {},
	}
	if *apiBan {
		(*expectedStats)[utils.MetaAPIBan] = &ltcache.CacheStats{Items: 254}
	}
	if err := precacheRPC.Call(context.Background(), utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
