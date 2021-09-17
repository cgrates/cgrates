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

package apis

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/ltcache"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/config"
)

var (
	chcCfg         *config.CGRConfig
	chcRPC         *birpc.Client
	chcCfgPath     string
	cacheConfigDIR string

	sTestsCacheSV1 = []func(t *testing.T){
		testCacheSLoadConfig,
		testCacheSInitDataDb,
		testCacheSInitStorDb,
		testCacheSStartEngine,
		testCacheSRpcConn,
		testCacheSSetAttributeProfile,
		testCacheSHasItemAttributeProfile,
		testCacheSRemoveItemAttributeProfile,
		testCacheSLoadCache,
		testCacheSRemoveItemsAndReloadCache,
		testCacheSSetRateProfile,
		testCacheSSetMoreAttributeProfiles,
		testCacheGetStatusMoreIDs,
		testCacheSClearCache,

		testCacheSStopEngine,
	}
)

// Test start here
func TestCacheSv1IT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cacheConfigDIR = "tutmysql"
	case utils.MetaMongo:
		cacheConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCacheSV1 {
		t.Run(cacheConfigDIR, stest)
	}
}

func testCacheSLoadConfig(t *testing.T) {
	var err error
	chcCfgPath = path.Join(*dataDir, "conf", "samples", "precache", cacheConfigDIR)
	if chcCfg, err = config.NewCGRConfigFromPath(context.Background(), chcCfgPath); err != nil {
		t.Error(err)
	}
}

func testCacheSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(chcCfg); err != nil {
		t.Fatal(err)
	}
}

// Empty tables before using them
func testCacheSInitStorDb(t *testing.T) {
	if err := engine.InitStorDB(chcCfg); err != nil {
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
	chcRPC, err = newRPCClient(chcCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to RPC: ", err.Error())
	}
}

//Set an attribute profile and rate profile to test cache's apis
func testCacheSSetAttributeProfile(t *testing.T) {
	attrPrf := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    utils.CGRateSorg,
			ID:        "TEST_ATTRIBUTES_IT_TEST",
			FilterIDs: []string{"*string:~*req.Account:1002"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.AccountField,
					Type:  utils.MetaConstant,
					Value: "1002",
				},
				{
					Path:  "*tenant",
					Type:  utils.MetaConstant,
					Value: "cgrates.itsyscom",
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	var reply string
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(err)
	}
}

func testCacheSSetRateProfile(t *testing.T) {
	var reply string
	rtPrf := &utils.APIRateProfile{
		ID:        "DefaultRate",
		Tenant:    "cgrates.org",
		FilterIDs: []string{"*string:~*req.Subject:1001"},
		Weights:   ";10",
		Rates: map[string]*utils.APIRate{
			"RT_WEEK": {
				ID:              "RT_WEEK",
				Weights:         ";0",
				ActivationTimes: "* * * * *",
				IntervalRates: []*utils.APIIntervalRate{
					{
						IntervalStart: "0",
						RecurrentFee:  utils.Float64Pointer(0.12),
						Unit:          utils.Float64Pointer(float64(time.Minute)),
						Increment:     utils.Float64Pointer(float64(time.Minute)),
					},
				},
			},
		},
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetRateProfile,
		rtPrf, &reply); err != nil {
		t.Error(err)
	}

	var result *utils.RateProfile
	expRtPrf, err := rtPrf.AsRateProfile()
	if err != nil {
		t.Error(err)
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1GetRateProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "DefaultRate",
			},
		}, &result); err != nil {
	} else {
		expRtPrf.Compile()
		result.Compile()
		if !reflect.DeepEqual(expRtPrf, result) {
			t.Errorf("Expected %+v, received %+v", utils.ToJSON(expRtPrf), utils.ToJSON(result))
		}
	}
}

func testCacheSHasItemAttributeProfile(t *testing.T) {
	var reply bool
	//it is not cached, so he cannot take it from cache
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1HasItem,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Unexpected reply result")
	}

	//also cannot take any itemIDs
	argsIds := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var result []string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}

	expectedAttr := &engine.APIAttributeProfile{
		Tenant:    utils.CGRateSorg,
		ID:        "TEST_ATTRIBUTES_IT_TEST",
		FilterIDs: []string{"*string:~*req.Account:1002"},
		Attributes: []*engine.ExternalAttribute{
			{
				Path:  utils.AccountField,
				Type:  utils.MetaConstant,
				Value: "1002",
			},
			{
				Path:  "*tenant",
				Type:  utils.MetaConstant,
				Value: "cgrates.itsyscom",
			},
		},
	}
	var resultAtr *engine.APIAttributeProfile
	if err := chcRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfile,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{
				Tenant: utils.CGRateSorg,
				ID:     "TEST_ATTRIBUTES_IT_TEST",
			},
		}, &resultAtr); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(resultAtr, expectedAttr) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expectedAttr), utils.ToJSON(result))
	}

	args = &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	//Getting an profile from db will set it in cache
	if err := chcRPC.Call(context.Background(), utils.CacheSv1HasItem,
		args, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Unexpected reply result")
	}

	//also can take from cache now
	expectedResult := []string{"cgrates.org:TEST_ATTRIBUTES_IT_TEST"}
	argsIds = &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %+v, received %+v", expectedResult, result)
	}
}

func testCacheSRemoveItemAttributeProfile(t *testing.T) {
	var reply string
	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1RemoveItem,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//as we removed it, we cannot take it from cache
	var result bool
	//it is not cached, so he cannot take it from cache
	argsHasItm := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1HasItem,
		argsHasItm, &result); err != nil {
		t.Error(err)
	} else if result {
		t.Errorf("Unexpected reply result")
	}

	//also cannot take any itemIDs
	argsIds := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	var resultIDs []string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &resultIDs); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, err)
	}
}

func testCacheSLoadCache(t *testing.T) {
	var reply string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1LoadCache,
		utils.NewAttrReloadCacheWithOpts(), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	var rcvStats map[string]*ltcache.CacheStats
	expstats := engine.GetDefaultEmptyCacheStats()
	expstats[utils.CacheAttributeProfiles].Items = 1
	expstats[utils.CacheAttributeFilterIndexes].Groups = 1
	expstats[utils.CacheAttributeFilterIndexes].Items = 1
	expstats[utils.CacheLoadIDs].Items = 27
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetCacheStats,
		new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvStats, expstats) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expstats), utils.ToJSON(rcvStats))
	}
}

func testCacheSRemoveItemsAndReloadCache(t *testing.T) {
	//as we loaded it, we can take the item
	var result []string
	expectedResult := []string{"cgrates.org:TEST_ATTRIBUTES_IT_TEST"}
	argsIds := &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %+v, received %+v", expectedResult, result)
	}

	//remove all items
	argsRemove := &utils.AttrReloadCacheWithAPIOpts{
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_IT_TEST"},
	}
	var replyRem string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1RemoveItems,
		argsRemove, &replyRem); err != nil {
		t.Error(err)
	} else if replyRem != utils.OK {
		t.Errorf("Unexpected reply return")
	}

	//as we removed the items, we cannot take it from cache
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected %+v, received %+v", utils.ErrNotFound, result)
	}

	//reload cache
	var reply string
	argsReload := &utils.AttrReloadCacheWithAPIOpts{
		AttributeProfileIDs: []string{"cgrates.org:TEST_ATTRIBUTES_IT_TEST"},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1LoadCache,
		argsReload, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	args := &utils.ArgsGetCacheItemWithAPIOpts{
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheAttributeProfiles,
			ItemID:  "cgrates.org:TEST_ATTRIBUTES_IT_TEST",
		},
	}
	var replyBool bool
	//Getting an profile from db will set it in cache
	if err := chcRPC.Call(context.Background(), utils.CacheSv1HasItem,
		args, &replyBool); err != nil {
		t.Error(err)
	} else if !replyBool {
		t.Errorf("Unexpected reply result")
	}

	//also can take from cache now
	expectedResult = []string{"cgrates.org:TEST_ATTRIBUTES_IT_TEST"}
	argsIds = &utils.ArgsGetCacheItemIDsWithAPIOpts{
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheAttributeProfiles,
		},
	}
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetItemIDs,
		argsIds, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedResult, result) {
		t.Errorf("Expected %+v, received %+v", expectedResult, result)
	}
}

func testCacheSSetMoreAttributeProfiles(t *testing.T) {
	attrPrf1 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_1",
			FilterIDs: []string{"*string:~*req.InitialField:InitialValue", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field1",
					Value: "Value1",
				},
			},
			Weight: 10,
		},
	}
	attrPrf2 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_2",
			FilterIDs: []string{"*string:~*req.Field1:Value1", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field2",
					Value: "Value2",
				},
			},
			Weight: 20,
		},
	}
	attrPrf3 := &engine.APIAttributeProfileWithAPIOpts{
		APIAttributeProfile: &engine.APIAttributeProfile{
			Tenant:    config.CgrConfig().GeneralCfg().DefaultTenant,
			ID:        "ATTR_3",
			FilterIDs: []string{"*string:~*req.Field2:Value2", "*ai:~*req.AnswerTime:2014-07-14T14:25:00Z", "*string:~*opts.*context:*sessions"},
			Attributes: []*engine.ExternalAttribute{
				{
					Path:  utils.MetaReq + utils.NestingSep + "Field3",
					Value: "Value3",
				},
			},
			Weight: 30,
		},
	}
	// Add attributeProfiles
	var result string
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf1, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf2, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := chcRPC.Call(context.Background(), utils.AdminSv1SetAttributeProfile,
		attrPrf3, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testCacheGetStatusMoreIDs(t *testing.T) {
	var reply string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1LoadCache,
		utils.NewAttrReloadCacheWithOpts(), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	var rcvStats map[string]*ltcache.CacheStats
	expstats := engine.GetDefaultEmptyCacheStats()
	expstats[utils.CacheAttributeProfiles].Items = 4
	expstats[utils.CacheAttributeFilterIndexes].Groups = 1
	expstats[utils.CacheAttributeFilterIndexes].Items = 5
	expstats[utils.CacheRateProfiles].Items = 1
	expstats[utils.CacheRateProfilesFilterIndexes].Groups = 1
	expstats[utils.CacheRateProfilesFilterIndexes].Items = 1
	expstats[utils.CacheRateFilterIndexes].Groups = 1
	expstats[utils.CacheRateFilterIndexes].Items = 1
	expstats[utils.CacheFilters].Items = 6
	expstats[utils.CacheRPCConnections].Items = 1
	expstats[utils.CacheLoadIDs].Items = 27
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetCacheStats,
		new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvStats, expstats) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expstats), utils.ToJSON(rcvStats))
	}
}

func testCacheSClearCache(t *testing.T) {
	var reply string
	if err := chcRPC.Call(context.Background(), utils.CacheSv1Clear,
		&utils.AttrCacheIDsWithAPIOpts{
			CacheIDs: nil,
		}, &reply); err != nil {
		t.Error(err)
	}

	//all cache cleared, empty items in cache
	var rcvStats map[string]*ltcache.CacheStats
	expStats := engine.GetDefaultEmptyCacheStats()
	if err := chcRPC.Call(context.Background(), utils.CacheSv1GetCacheStats,
		new(utils.AttrCacheIDsWithAPIOpts), &rcvStats); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rcvStats, expStats) {
		t.Errorf("Expected %+v \n, received %+v", utils.ToJSON(expStats), utils.ToJSON(rcvStats))
	}
}

func testCacheSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
