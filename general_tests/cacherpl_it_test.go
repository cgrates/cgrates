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
	"os/exec"
	"path"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dspEngine1Cfg     *config.CGRConfig
	dspEngine1CfgPath string
	dspEngine1RPC     *birpc.Client
	dspEngine2Cfg     *config.CGRConfig
	dspEngine2CfgPath string
	dspEngine2RPC     *birpc.Client
	engine1Cfg        *config.CGRConfig
	engine1CfgPath    string
	engine1RPC        *birpc.Client

	sTestsCacheRpl = []func(t *testing.T){
		testCacheRplInitCfg,
		testCacheRplInitDataDb,
		testCacheRplStartEngine,
		testCacheRplRpcConn,
		testCacheRplAddData,
		testCacheRplPing,
		testCacheRplCheckReplication,
		testCacheRplCheckLoadReplication,

		testCacheRplStopEngine,
	}

	sTestsCacheRplAA = []func(t *testing.T){
		testCacheRplAAInitCfg,
		testCacheRplInitDataDb,
		testCacheRplStartEngine,
		testCacheRplRpcConn,
		testCacheRplAAAddData,
		testCacheRplAACheckReplication,
		testCacheRplAACheckLoadReplication,

		testCacheRplStopEngine,
	}
)

func TestCacheReplications(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		for _, stest := range sTestsCacheRpl {
			t.Run("TestCacheReplications", stest)
		}
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

}

func TestCacheReplicationActiveActive(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		for _, stest := range sTestsCacheRplAA {
			t.Run("TestCacheReplicationActiveActive", stest)
		}
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
}

func testCacheRplInitCfg(t *testing.T) {
	var err error
	dspEngine1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_replicate", "dispatcher_engine")
	dspEngine1Cfg, err = config.NewCGRConfigFromPath(dspEngine1CfgPath)
	if err != nil {
		t.Error(err)
	}

	dspEngine2CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_replicate", "dispatcher_engine2")
	dspEngine2Cfg, err = config.NewCGRConfigFromPath(dspEngine2CfgPath)
	if err != nil {
		t.Error(err)
	}

	engine1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_replicate", "engine1")
	engine1Cfg, err = config.NewCGRConfigFromPath(engine1CfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCacheRplAAInitCfg(t *testing.T) {
	var err error
	dspEngine1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_rpl_active_active", "dispatcher_engine")
	dspEngine1Cfg, err = config.NewCGRConfigFromPath(dspEngine1CfgPath)
	if err != nil {
		t.Error(err)
	}

	dspEngine2CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_rpl_active_active", "dispatcher_engine2")
	dspEngine2Cfg, err = config.NewCGRConfigFromPath(dspEngine2CfgPath)
	if err != nil {
		t.Error(err)
	}

	engine1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "cache_rpl_active_active", "engine1")
	engine1Cfg, err = config.NewCGRConfigFromPath(engine1CfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testCacheRplInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(dspEngine1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(dspEngine2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testCacheRplStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dspEngine1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(dspEngine2CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(engine1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCacheRplRpcConn(t *testing.T) {
	dspEngine1RPC = engine.NewRPCClient(t, dspEngine1Cfg.ListenCfg())
	dspEngine2RPC = engine.NewRPCClient(t, dspEngine2Cfg.ListenCfg())
	engine1RPC = engine.NewRPCClient(t, engine1Cfg.ListenCfg())
}

func testCacheRplAddData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspEngine1CfgPath, "-path",
			path.Join(*utils.DataDir, "tariffplans", "cache_replications", "dispatcher_engine"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(2 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}

	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspEngine2CfgPath, "-path",
			path.Join(*utils.DataDir, "tariffplans", "cache_replications", "dispatcher_engine2"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(2 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}

	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DefaultCharger",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{utils.MetaNone},
			Weight:       20,
		},
	}
	var result string
	if err := engine1RPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testCacheRplAAAddData(t *testing.T) {
	wchan := make(chan struct{}, 1)
	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspEngine1CfgPath, "-path",
			path.Join(*utils.DataDir, "tariffplans", "cache_rpl_active_active", "dispatcher_engine"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(2 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}

	go func() {
		loaderPath, err := exec.LookPath("cgr-loader")
		if err != nil {
			t.Error(err)
		}
		loader := exec.Command(loaderPath, "-config_path", dspEngine2CfgPath, "-path",
			path.Join(*utils.DataDir, "tariffplans", "cache_rpl_active_active", "dispatcher_engine2"))

		if err := loader.Start(); err != nil {
			t.Error(err)
		}
		loader.Wait()
		wchan <- struct{}{}
	}()
	select {
	case <-wchan:
	case <-time.After(2 * time.Second):
		t.Errorf("cgr-loader failed: ")
	}

	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "DefaultCharger",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{utils.MetaNone},
			Weight:       20,
		},
	}
	var result string
	if err := engine1RPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testCacheRplPing(t *testing.T) {
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID: "testRoute123",
		},
	}
	if err := dspEngine1RPC.Call(context.Background(), utils.DispatcherSv1RemoteStatus, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "Engine1" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}

	var rpl string
	if err := dspEngine1RPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsRouteID: "testRoute123",
		},
	}, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.Pong {
		t.Errorf("Received: %s", rpl)
	}
}

func testCacheRplCheckReplication(t *testing.T) {
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.DispatcherSv1RemoteStatus, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "DispatcherEngine2" {
		t.Errorf("Received: %s", utils.ToJSON(reply))
	}
	var rcvKeys []string
	expKeys := []string{"testRoute123:*core", "testRoute123:*attributes"}
	argsAPI := utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherRoutes,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

	var rpl string
	if err := dspEngine2RPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsRouteID: "testRoute123",
		},
	}, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.Pong {
		t.Errorf("Received: %s", rpl)
	}
}

func testCacheRplAACheckReplication(t *testing.T) {
	var rcvKeys []string
	argsAPI := utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherRoutes,
		},
	}
	if err := dspEngine1RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var rpl string
	if err := dspEngine2RPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",

		APIOpts: map[string]any{
			utils.OptsRouteID: "testRouteFromDispatcher2",
		},
	}, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.Pong {
		t.Errorf("Received: %s", rpl)
	}

	if err := dspEngine1RPC.Call(context.Background(), utils.AttributeSv1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID: "testRouteFromDispatcher1",
		},
	}, &rpl); err != nil {
		t.Error(err)
	} else if rpl != utils.Pong {
		t.Errorf("Received: %s", rpl)
	}

	expKeys := []string{"testRouteFromDispatcher2:*attributes", "testRouteFromDispatcher1:*attributes"}

	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

	if err := dspEngine1RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

}

func testCacheRplAACheckLoadReplication(t *testing.T) {
	var rcvKeys []string
	argsAPI := utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherLoads,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := dspEngine1RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var wgDisp1 sync.WaitGroup
	var wgDisp2 sync.WaitGroup
	for i := 0; i < 10; i++ {
		wgDisp1.Add(1)
		wgDisp2.Add(1)
		go func() {
			var rpl []*engine.ChrgSProcessEventReply
			if err := dspEngine1RPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "testCacheRplAACheckLoadReplication",
				Event: map[string]any{
					utils.AccountField: "1007",
					utils.Destination:  "+491511231234",
					"EventName":        "TestLoad",
				},

				APIOpts: map[string]any{
					utils.OptsRouteID: "testRouteFromDispatcher1",
				},
			}, &rpl); err != nil {
				t.Error(err)
			} else if rpl[0].ChargerSProfile != "DefaultCharger" {
				t.Errorf("Received: %+v", utils.ToJSON(rpl))
			}
			wgDisp1.Done()
		}()
		go func() {
			var rpl []*engine.ChrgSProcessEventReply
			if err := dspEngine2RPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, &utils.CGREvent{

				Tenant: "cgrates.org",
				ID:     "testCacheRplAACheckLoadReplication",
				Event: map[string]any{
					utils.AccountField: "1007",
					utils.Destination:  "+491511231234",
					"EventName":        "TestLoad",
				},

				APIOpts: map[string]any{
					utils.OptsRouteID: "testRouteFromDispatcher2",
				},
			}, &rpl); err != nil {
				t.Error(err)
			} else if rpl[0].ChargerSProfile != "DefaultCharger" {
				t.Errorf("Received: %+v", utils.ToJSON(rpl))
			}
			wgDisp2.Done()
		}()
	}
	wgDisp1.Wait()
	wgDisp2.Wait()
	expKeys := []string{"testRouteFromDispatcher1:*attributes",
		"testRouteFromDispatcher1:*chargers", "testRouteFromDispatcher2:*attributes",
		"testRouteFromDispatcher2:*chargers"}
	argsAPI = utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherRoutes,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
	if err := dspEngine1RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

	expKeys = []string{"cgrates.org:Engine2"}
	argsAPI = utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherLoads,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
	if err := dspEngine1RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}
}

func testCacheRplCheckLoadReplication(t *testing.T) {
	var rcvKeys []string
	argsAPI := utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherLoads,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var rpl []*engine.ChrgSProcessEventReply
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			if err := dspEngine1RPC.Call(context.Background(), utils.ChargerSv1ProcessEvent, &utils.CGREvent{

				Tenant: "cgrates.org",
				ID:     "testCacheRplCheckLoadReplication",
				Event: map[string]any{
					utils.AccountField: "1007",
					utils.Destination:  "+491511231234",
					"EventName":        "TestLoad",
				},

				APIOpts: map[string]any{
					utils.OptsRouteID: "testRoute123",
				},
			}, &rpl); err != nil {
				t.Error(err)
			} else if rpl[0].ChargerSProfile != "DefaultCharger" {
				t.Errorf("Received: %+v", utils.ToJSON(rpl))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	expKeys := []string{"testRoute123:*core", "testRoute123:*attributes", "testRoute123:*chargers"}
	argsAPI = utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherRoutes,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

	expKeys = []string{"cgrates.org:Engine2"}
	argsAPI = utils.ArgsGetCacheItemIDsWithAPIOpts{
		Tenant: "cgrates.org",
		ArgsGetCacheItemIDs: utils.ArgsGetCacheItemIDs{
			CacheID: utils.CacheDispatcherLoads,
		},
	}
	if err := dspEngine2RPC.Call(context.Background(), utils.CacheSv1GetItemIDs, argsAPI, &rcvKeys); err != nil {
		t.Error(err.Error())
	}
	sort.Strings(rcvKeys)
	sort.Strings(expKeys)
	if !reflect.DeepEqual(expKeys, rcvKeys) {
		t.Errorf("Expected: %+v, received: %+v", expKeys, rcvKeys)
	}

}

func testCacheRplStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
