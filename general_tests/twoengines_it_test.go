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
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
)

var (
	engineOneCfgPath string
	engineOneCfg     *config.CGRConfig
	engineOneRpc     *birpc.Client

	engineTwoCfgPath string
	engineTwoCfg     *config.CGRConfig
	engineTwoRpc     *birpc.Client
)

var sTestsTwoEnginesIT = []func(t *testing.T){
	testTwoEnginesInitConfig,
	testTwoEnginesInitDataDB,
	testTwoEnginesInitStorDB,
	testTwoEnginesStartEngine,
	testTwoEnginesRPC,
	testTwoEnginesCheckCacheBeforeSet,
	testTwoEnginesSetThreshold,
	testTwoEnginesCheckCacheAfterSet,
	testTwoEnginesUpdateThreshold,
	testTwoEnginesKillEngines,
}

func TestTwoEngines(t *testing.T) {
	for _, test := range sTestsTwoEnginesIT {
		t.Run("TestTwoEngines", test)
	}
}

func testTwoEnginesInitConfig(t *testing.T) {
	var err error
	engineOneCfgPath = path.Join(*utils.DataDir, "conf", "samples", "twoengines", "engine1")
	if engineOneCfg, err = config.NewCGRConfigFromPath(engineOneCfgPath); err != nil {
		t.Fatal(err)
	}
	engineTwoCfgPath = path.Join(*utils.DataDir, "conf", "samples", "twoengines", "engine2")
	if engineTwoCfg, err = config.NewCGRConfigFromPath(engineTwoCfgPath); err != nil {
		t.Fatal(err)
	}

}
func testTwoEnginesInitDataDB(t *testing.T) {
	if err := engine.InitDataDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
}
func testTwoEnginesInitStorDB(t *testing.T) {
	if err := engine.InitStorDb(engineOneCfg); err != nil {
		t.Fatal(err)
	}
}
func testTwoEnginesStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(engineOneCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engineTwoCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testTwoEnginesRPC(t *testing.T) {
	engineOneRpc = engine.NewRPCClient(t, engineOneCfg.ListenCfg())
	engineTwoRpc = engine.NewRPCClient(t, engineTwoCfg.ListenCfg())
}

func testTwoEnginesCheckCacheBeforeSet(t *testing.T) {
	var reply bool
	argHasItem := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "cgrates.org:THD_TwoEnginesTest",
	}
	if err := engineOneRpc.Call(context.Background(), utils.CacheSv1HasItem, argHasItem, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false , received: %v ", reply)
	}
	var rcvKeys []string
	argGetItemIDs := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheThresholdProfiles,
	}
	if err := engineOneRpc.Call(context.Background(), utils.CacheSv1GetItemIDs, argGetItemIDs, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ", utils.ErrNotFound, err.Error(), rcvKeys)
	}

	if err := engineTwoRpc.Call(context.Background(), utils.CacheSv1HasItem, argHasItem, &reply); err != nil {
		t.Error(err)
	} else if reply {
		t.Errorf("Expected: false , received: %v ", reply)
	}
	if err := engineTwoRpc.Call(context.Background(), utils.CacheSv1GetItemIDs, argGetItemIDs, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Fatalf("Expected error: %s received error: %s and reply: %v ", utils.ErrNotFound, err.Error(), rcvKeys)
	}
}

func testTwoEnginesSetThreshold(t *testing.T) {
	var reply *engine.ThresholdProfile
	// enforce caching with nil on engine2 so CacheSv1.ReloadCache load correctly the threshold
	if err := engineTwoRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_TwoEnginesTest"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	var result string
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_TwoEnginesTest",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	if err := engineOneRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := engineOneRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_TwoEnginesTest"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, reply)
	}
}

func testTwoEnginesCheckCacheAfterSet(t *testing.T) {
	var reply bool
	expected := true
	argHasItem := utils.ArgsGetCacheItem{
		CacheID: utils.CacheThresholdProfiles,
		ItemID:  "cgrates.org:THD_TwoEnginesTest",
	}
	if err := engineOneRpc.Call(context.Background(), utils.CacheSv1HasItem, argHasItem, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
	var rcvKeys []string
	expKeys := []string{"cgrates.org:THD_TwoEnginesTest"}
	argGetItemIDs := utils.ArgsGetCacheItemIDs{
		CacheID: utils.CacheThresholdProfiles,
	}
	if err := engineOneRpc.Call(context.Background(), utils.CacheSv1GetItemIDs, argGetItemIDs, &rcvKeys); err != nil {
		t.Fatalf("Got error on APIerSv1.GetCacheStats: %s ", err.Error())
	} else if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

	if err := engineTwoRpc.Call(context.Background(), utils.CacheSv1HasItem, argHasItem, &reply); err != nil {
		t.Error(err)
	} else if !reply {
		t.Errorf("Expected: %v , received:%v", expected, reply)
	}
	if err := engineTwoRpc.Call(context.Background(), utils.CacheSv1GetItemIDs, argGetItemIDs, &rcvKeys); err != nil {
		t.Fatalf("Got error on APIerSv1.GetCacheStats: %s ", err.Error())
	} else if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
	// after we verify the cache make sure it was set correctly there
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_TwoEnginesTest",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			MaxHits:   -1,
			MinSleep:  5 * time.Minute,
			Blocker:   false,
			Weight:    20.0,
			ActionIDs: []string{"ACT_1"},
			Async:     true,
		},
	}
	var rplTh *engine.ThresholdProfile
	if err := engineTwoRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_TwoEnginesTest"}, &rplTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rplTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rplTh)
	}
}

func testTwoEnginesUpdateThreshold(t *testing.T) {
	var rplTh *engine.ThresholdProfile
	var result string
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_TwoEnginesTest",
			FilterIDs: []string{"*string:~*req.Account:10"},
			MaxHits:   -1,
			MinSleep:  time.Minute,
			Blocker:   false,
			Weight:    50.0,
			ActionIDs: []string{"ACT_1.1"},
			Async:     true,
		},
		APIOpts: map[string]any{
			utils.CacheOpt: utils.MetaReload,
		},
	}
	if err := engineOneRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := engineOneRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_TwoEnginesTest"}, &rplTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rplTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rplTh)
	}
	if err := engineTwoRpc.Call(context.Background(), utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_TwoEnginesTest"}, &rplTh); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl.ThresholdProfile, rplTh) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl.ThresholdProfile, rplTh)
	}
}

func testTwoEnginesKillEngines(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
