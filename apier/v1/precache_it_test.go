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
	"flag"
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
	precacheConfigDIR string //run tests for specific configuration

	// use this flag to test the APIBan implemnentation for precache
	apiBan = flag.Bool("apiban", true, "used to control if we run the apiban tests")

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
	if *apiBan {
		precacheConfigDIR += "_apiban"
	}
	for _, stest := range sTestsPrecache {
		t.Run(precacheConfigDIR, stest)
	}
}

func testPrecacheInitCfg(t *testing.T) {
	var err error
	precacheCfgPath = path.Join(*dataDir, "conf", "samples", "precache", precacheConfigDIR)
	precacheCfg, err = config.NewCGRConfigFromPath(precacheCfgPath)
	if err != nil {
		t.Error(err)
	}
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
	args := &utils.AttrCacheIDsWithAPIOpts{
		CacheIDs: []string{},
	}
	dfltStats := engine.GetDefaultEmptyCacheStats()
	expectedStats := &dfltStats
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v,\n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "precache")}
	if err := precacheRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
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
	time.Sleep(2 * time.Second) // let the *apiban cache to be populated
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
		utils.CacheAttributeProfiles:         {Items: 1},
		utils.CacheChargerFilterIndexes:      {},
		utils.CacheChargerProfiles:           {},
		utils.CacheDispatcherFilterIndexes:   {},
		utils.CacheDispatcherProfiles:        {Items: 6},
		utils.CacheDispatcherHosts:           {Items: 1},
		utils.CacheDispatcherRoutes:          {},
		utils.CacheDispatcherLoads:           {},
		utils.CacheDestinations:              {Items: 5},
		utils.CacheDispatchers:               {},
		utils.CacheEventResources:            {},
		utils.CacheFilters:                   {Items: 15},
		utils.CacheRateProfilesFilterIndexes: {},
		utils.CacheRateFilterIndexes:         {},
		utils.CacheRateProfiles:              {},
		utils.CacheRatingPlans:               {Items: 4},
		utils.CacheRatingProfiles:            {Items: 5},
		utils.CacheResourceFilterIndexes: {
			Items:  6,
			Groups: 1,
		},
		utils.CacheResourceProfiles:    {Items: 3},
		utils.CacheResources:           {Items: 3},
		utils.CacheReverseDestinations: {Items: 7},
		utils.CacheRPCResponses:        {},
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
			Items:  10,
			Groups: 1,
		},
		utils.CacheThresholdProfiles:            {Items: 7},
		utils.CacheThresholds:                   {Items: 7},
		utils.CacheTimings:                      {},
		utils.CacheDiameterMessages:             {},
		utils.CacheClosedSessions:               {},
		utils.CacheLoadIDs:                      {},
		utils.CacheRPCConnections:               {},
		utils.CacheCDRIDs:                       {},
		utils.CacheRatingProfilesTmp:            {},
		utils.CacheUCH:                          {},
		utils.CacheReverseFilterIndexes:         {},
		utils.CacheAccounts:                     {},
		utils.CacheVersions:                     {},
		utils.CacheTBLTPTimings:                 {},
		utils.CacheTBLTPDestinations:            {},
		utils.CacheTBLTPRates:                   {},
		utils.CacheTBLTPDestinationRates:        {},
		utils.CacheTBLTPRatingPlans:             {},
		utils.CacheTBLTPRatingProfiles:          {},
		utils.CacheTBLTPSharedGroups:            {},
		utils.CacheTBLTPActions:                 {},
		utils.CacheTBLTPActionPlans:             {},
		utils.CacheTBLTPActionTriggers:          {},
		utils.CacheTBLTPAccountActions:          {},
		utils.CacheTBLTPResources:               {},
		utils.CacheTBLTPStats:                   {},
		utils.CacheTBLTPThresholds:              {},
		utils.CacheTBLTPFilters:                 {},
		utils.CacheSessionCostsTBL:              {},
		utils.CacheCDRsTBL:                      {},
		utils.CacheTBLTPRoutes:                  {},
		utils.CacheTBLTPAttributes:              {},
		utils.CacheTBLTPChargers:                {},
		utils.CacheTBLTPDispatchers:             {},
		utils.CacheTBLTPDispatcherHosts:         {},
		utils.CacheTBLTPRateProfiles:            {},
		utils.MetaAPIBan:                        {},
		utils.CacheTBLTPAccountProfiles:         {},
		utils.CacheAccountProfiles:              {},
		utils.CacheAccountProfilesFilterIndexes: {},
		utils.CacheReplicationHosts:             {},
	}
	if *apiBan {
		(*expectedStats)[utils.MetaAPIBan] = &ltcache.CacheStats{Items: 254}
	}
	if err := precacheRPC.Call(utils.CacheSv1GetCacheStats, args, &reply); err != nil {
		t.Error(err.Error())
	} else if !reflect.DeepEqual(reply, expectedStats) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(expectedStats), utils.ToJSON(reply))
	}
}

func testPrecacheKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
