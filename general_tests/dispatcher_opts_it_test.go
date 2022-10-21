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
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	setterCfgPath   string
	cfg2CfgPath     string
	cfg1CfgPath     string
	setterCfg       *config.CGRConfig
	cfg2OptsCfg     *config.CGRConfig
	cfg1Cfg         *config.CGRConfig
	setterRPC       *rpc.Client
	cgr2RPC         *rpc.Client
	cgr1RPC         *rpc.Client
	cfg1ConfigDIR   string
	cfg2ConfigDIR   string
	setterConfigDIR string
	dpsOptsTest     = []func(t *testing.T){
		// FIRST PART OF THE TEST
		// Start engine with Dispatcher on engine 2012
		testDispatcherCgr1InitCfg,
		testDispatcherCgr1InitDataDb,
		testDispatcherCgr1StartEngine,
		testDispatcherCgr1RPCConn,

		// Sending Status requests in both engines, with *dispatchers:false
		testDispatcherCgr2InitCfg,
		testDispatcherCgr2StartEngine,
		testDispatcherCgr2RPCConn,

		testDispatcherCgr1CoreStatus, // *disaptchers:false
		testDispatcherCgr2CoreStatus, // *disaptchers:false

		testDispatcherGetItemBothEnginesFirstAttempt, // NOT FOUND

		testDispatcherCgr1StopEngine,
		testDispatcherCgr2StopEngine,

		// SECOND PART OF THE TEST
		// START HOST2 engine

		testDispatcherSetterInitCfg,
		testDispatcherSetterStartEngine,
		testDispatcherSetterRPCConn,

		testDispatcherCgr2StartEngine,
		testDispatcherCgr2RPCConn,

		testDispatcherSetterSetDispatcherProfile, // contains both hosts, HOST1 prio, host2 backup

		testDispatcherCgr2CoreStatusWithRouteID, // HOST2 matched because HOST1 is not started yet
		testDispatcherCgr2GetItemHOST2,

		// START HOST1 engine
		testDispatcherCgr1StartEngine,
		testDispatcherCgr1RPCConn,
		testDispatcherCgr1CoreStatusWithRouteIDSecondAttempt, // same HOST2 will be matched, due to routeID

		// clear cache in order to remove routeID
		testDisaptcherCacheClear,
		testDispatcherCgr1CoreStatusWithRouteIDButHost1, // due to clearing cache, HOST1 will be matched

		// verify cache of dispatchers, SetDispatcherProfile API should reload the dispatchers cache (instance, profile and route)
		testDispatcherCgr1CheckCacheAfterRouting,
		testDispatcherSetterSetDispatcherProfileOverwrite,
		testDispatcherCheckCacheAfterSetDispatcherDSP1,
		testDispatcherSetterSetAnotherProifle,          //DSP2
		testDispatcherCheckCacheAfterSetDispatcherDSP1, //we set DSP2, so for DSP1 nothing changed
		testDispatcherCheckCacheAfterSetDispatcherDSP2, //NOT_FOUND for every get, cause it was not used that profile before

		testDispatcherCgr1StopEngine,
		testDispatcherCgr2StopEngine,
	}
)

func TestDispatcherOpts(t *testing.T) {
	for _, test := range dpsOptsTest {
		t.Run("dispatcher-opts", test)
	}
}

func testDispatcherCgr1InitCfg(t *testing.T) {
	cfg1ConfigDIR = "dispatcher_opts_host1"
	var err error
	cfg1CfgPath = path.Join(*dataDir, "conf", "samples", cfg1ConfigDIR)
	cfg1Cfg, err = config.NewCGRConfigFromPath(cfg1CfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherCgr1InitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cfg1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherCgr1StartEngine(t *testing.T) {
	if _, err := engine.StartEngine(cfg1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherCgr1RPCConn(t *testing.T) {
	var err error
	cgr1RPC, err = newRPCClient(cfg1Cfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherCgr2InitCfg(t *testing.T) {
	cfg2ConfigDIR = "dispatcher_opts_host2" //changed with the cfg with dispatcher on
	var err error
	cfg2CfgPath = path.Join(*dataDir, "conf", "samples", cfg2ConfigDIR)
	cfg2OptsCfg, err = config.NewCGRConfigFromPath(cfg2CfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherCgr2StartEngine(t *testing.T) {
	if _, err := engine.StartEngine(cfg2CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherCgr2RPCConn(t *testing.T) {
	var err error
	cgr2RPC, err = newRPCClient(cfg2OptsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherCgr1CoreStatus(t *testing.T) {
	// HOST1 host matched :2012
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID:     "account#dan.bogos",
			utils.MetaDispatchers: false,
		},
	}
	if err := cgr1RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherCgr2CoreStatus(t *testing.T) {
	// HOST2 host matched because it was called from engine with port :4012 -> host2
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID:     "account#dan.bogos",
			utils.MetaDispatchers: false,
		},
	}
	if err := cgr2RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherGetItemBothEnginesFirstAttempt(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherSetterInitCfg(t *testing.T) {
	setterConfigDIR = "dispatcher_opts_setter"
	var err error
	setterCfgPath = path.Join(*dataDir, "conf", "samples", setterConfigDIR)
	setterCfg, err = config.NewCGRConfigFromPath(setterCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherSetterStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(setterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherSetterRPCConn(t *testing.T) {
	var err error
	setterRPC, err = newRPCClient(setterCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherSetterSetDispatcherProfile(t *testing.T) {
	// Set DispatcherHost
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "HOST1",
				Address:         "127.0.0.1:2012", // CGR1
				Transport:       "*json",
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
			},
		},
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherHost: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	setDispatcherHost = &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "HOST2",
				Address:         "127.0.0.1:4012", // CGR2
				Transport:       "*json",
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
			},
		},
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherHost: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Set DispatcherProfile
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Weight:     10,
			Subsystems: []string{utils.MetaAny},
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST1",
					Weight: 10,
				},
				{
					ID:     "HOST2",
					Weight: 5,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherCgr2CoreStatusWithRouteID(t *testing.T) {
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	// even if HOST1 is prio, this engine was not staretd yet, so HOST2 matched
	if err := cgr2RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherCgr1CoreStatusWithRouteIDSecondAttempt(t *testing.T) {
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	// same HOST2 will be matched, due to routeID
	if err := cgr1RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherCgr2GetItemHOST2(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]interface{}{
			utils.Tenant:    "cgrates.org",
			utils.ProfileID: "DSP1",
			"HostID":        "HOST2",
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]interface{}{
			utils.FilterIDs: nil,
			"Hosts": []interface{}{
				map[string]interface{}{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST1",
					utils.Params:    nil,
					utils.Weight:    10.,
				},
				map[string]interface{}{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ActivationIntervalString: nil,
			utils.ID:                       "DSP1",
			utils.Strategy:                 "*weight",
			utils.Subsystems:               []interface{}{"*any"},
			"StrategyParams":               nil,
			utils.Tenant:                   "cgrates.org",
			utils.Weight:                   10.,
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// reply here is an interface type(singleResultDispatcher), it exists
	if err := cgr2RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDisaptcherCacheClear(t *testing.T) {
	var reply string
	if err := cgr1RPC.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	if err := cgr2RPC.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testDispatcherCgr1CoreStatusWithRouteIDButHost1(t *testing.T) {
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	// as the cache was cleared, HOST1 will match due to his high prio, and it will be set as *dispatcher_routes as HOST1
	if err := cgr1RPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherCgr1CheckCacheAfterRouting(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]interface{}{
			utils.Tenant:    "cgrates.org",
			utils.ProfileID: "DSP1",
			"HostID":        "HOST1",
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]interface{}{
			utils.ActivationIntervalString: nil,
			utils.FilterIDs:                nil,
			"Hosts": []interface{}{
				map[string]interface{}{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST1",
					utils.Params:    nil,
					utils.Weight:    10.,
				},
				map[string]interface{}{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:         "DSP1",
			utils.Strategy:   "*weight",
			utils.Subsystems: []interface{}{"*any"},
			"StrategyParams": nil,
			utils.Tenant:     "cgrates.org",
			utils.Weight:     10.,
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// reply here is an interface type(singleResultDispatcher), it exists
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDispatcherSetterSetDispatcherProfileOverwrite(t *testing.T) {
	// as the cache was cleared, and previously the HOST1 matched, setting the profile with only HOST2 will remove the
	// DispatcherRoutes, DispatcherProfile and the DispatcherInstance
	var replyStr string
	// Set DispatcherProfile
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Weight:     10,
			Subsystems: []string{utils.MetaAny},
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST2",
					Weight: 5,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	time.Sleep(100 * time.Millisecond)
}

func testDispatcherCheckCacheAfterSetDispatcherDSP1(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
			"adi3":                "nu",
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{} // Should receive NOT_FOUND, as CallCache that was called in API will remove the DispatcherRoute
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
			"adi2":                "nu",
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// as the DSP1 profile was overwritten, only HOST2 in profile will be contained
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]interface{}{
			utils.FilterIDs: nil,
			"Hosts": []interface{}{
				map[string]interface{}{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			"ActivationInterval": nil,
			"Subsystems":         []interface{}{"*any"},
			utils.ID:             "DSP1",
			utils.Strategy:       "*weight",
			"StrategyParams":     nil,
			utils.Tenant:         "cgrates.org",
			utils.Weight:         10.,
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", utils.ToJSON(expected), utils.ToJSON(reply))
		}
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
			"adi1":                "nu",
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// DispatcherInstance should also be removed, so it will be NOT_FOUND
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v and reply: %v", err, reply)
	}
}

func testDispatcherSetterSetAnotherProifle(t *testing.T) {
	var replyStr string
	// Set DispatcherProfile DSP2 with the existing hosts
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP2",
			Strategy:   "*weight",
			Weight:     20,
			Subsystems: []string{utils.MetaAny},
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST1",
					Weight: 50,
				},
				{
					ID:     "HOST2",
					Weight: 125,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
	time.Sleep(100 * time.Millisecond)
}

func testDispatcherCheckCacheAfterSetDispatcherDSP2(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	// NOT_FOUND
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP2",
		},
	}
	// NOT_FOUND
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP2",
		},
	}
	// NOT_FOUND
	if err := cgr1RPC.Call(utils.CacheSv1GetItem, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func testDispatcherCgr1StopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherCgr2StopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
