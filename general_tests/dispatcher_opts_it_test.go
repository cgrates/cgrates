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
)

var (
	dspOptsCfgPath   string
	adminsCfgPath    string
	dspOptsCfg       *config.CGRConfig
	adminsCfg        *config.CGRConfig
	dspOptsRPC       *birpc.Client
	adminsRPC        *birpc.Client
	dspOptsConfigDIR string
	dpsOptsTest      = []func(t *testing.T){
		// FIRST APRT OF THE TEST
		// Start engine without Dispatcher on engine 4012
		testDispatcherOptsAdminInitCfg,
		testDispatcherOptsAdminFlushDBs,
		testDispatcherOptsAdminStartEngine,
		testDispatcherOptsAdminRPCConn,

		// Sending Status requests in both engines, with *dispatchers:false
		testDispatcherOptsDSPInitCfg,
		testDispatcherOptsDSPStartEngine,
		testDispatcherOptsDSPRPCConn,

		testDispatcherOptsCoreStatus,  // *disaptchers:false
		testDispatcherAdminCoreStatus, // *disaptchers:false

		testDispatcherGetItemBothEnginesFirstAttempt, // NOT FOUND

		testDispatcherOptsDSPStopEngine,
		testDispatcherOptsAdminStopEngine,

		// SECOND PART OF THE TEST
		// START HOST2 engine
		testDispatcherOptsAdminStartEngine,
		testDispatcherOptsAdminRPCConn,

		testDispatcherOptsAdminSetDispatcherProfile, // contains both hosts, HOST1 prio, host2 backup

		testDispatcherAdminCoreStatusWithRouteID, // HOST2 matched because HOST1 is not started yet
		testDispatcherAdminGetItemHOST2,

		// START HOST1 engine
		testDispatcherOptsDSPStartEngine,
		testDispatcherOptsDSPRPCConn,
		testDispatcherAdminCoreStatusWithRouteID, // same HOST2 will be matched, due to routeID

		// clear cache in order to remove routeID
		testDisaptcherCacheClear,
		testDispatcherAdminCoreStatusWithRouteIDButHost1, // due to clearing cache, HOST1 will be matched

		// verify cache of dispatchers, SetDispatcherProfile API should reload the dispatchers cache (instance, profile and route)
		testDispatcherAdminCheckCacheAfterRouting,
		testDispatcherSetDispatcherProfileOverwrite,
		testDispatcherCheckCacheAfterSetDispatcherDSP1,
		testDispatcherSetAnotherProifle,                //DSP2
		testDispatcherCheckCacheAfterSetDispatcherDSP1, //we set DSP2, so for DSP1 nothing changed
		testDispatcherCheckCacheAfterSetDispatcherDSP2, //NOT_FOUND for every get, cause it was not used that profile before

		testDispatcherOptsDSPStopEngine,
		testDispatcherOptsAdminStopEngine,
	}
)

func TestDispatcherOpts(t *testing.T) {
	for _, test := range dpsOptsTest {
		t.Run(dspOptsConfigDIR, test)
	}
}

func testDispatcherOptsAdminInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts_admin"
	var err error
	adminsCfgPath = path.Join(*utils.DataDir, "conf", "samples", dspOptsConfigDIR)
	adminsCfg, err = config.NewCGRConfigFromPath(context.Background(), adminsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAdminFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(adminsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(adminsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsAdminStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(adminsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAdminRPCConn(t *testing.T) {
	adminsRPC = engine.NewRPCClient(t, adminsCfg.ListenCfg(), *utils.Encoding)
}

func testDispatcherOptsDSPInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts" //changed with the cfg with dispatcher on
	var err error
	dspOptsCfgPath = path.Join(*utils.DataDir, "conf", "samples", dspOptsConfigDIR)
	dspOptsCfg, err = config.NewCGRConfigFromPath(context.Background(), dspOptsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsDSPStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(dspOptsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsDSPRPCConn(t *testing.T) {
	dspOptsRPC = engine.NewRPCClient(t, dspOptsCfg.ListenCfg(), *utils.Encoding)
}

func testDispatcherOptsCoreStatus(t *testing.T) {
	// HOST1 host matched
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := dspOptsRPC.Call(context.Background(), utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminCoreStatus(t *testing.T) {
	// HOST2 host matched because it was called from engine with port :4012 -> host2
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID:     "account#dan.bogos",
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherGetItemBothEnginesFirstAttempt(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply any
	if err := dspOptsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := dspOptsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := dspOptsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherOptsAdminSetDispatcherProfile(t *testing.T) {
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherHost: ", err)
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherHost: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Set DispatcherProfile
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:   "cgrates.org",
			ID:       "DSP1",
			Strategy: "*weight",
			Weight:   10,
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherAdminCoreStatusWithRouteID(t *testing.T) {
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminGetItemHOST2(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply any
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]any{
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]any{
			utils.FilterIDs: nil,
			utils.Hosts: []any{
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST1",
					utils.Params:    nil,
					utils.Weight:    10.,
				},
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:             "DSP1",
			utils.Strategy:       "*weight",
			utils.StrategyParams: nil,
			utils.Tenant:         "cgrates.org",
			utils.Weight:         10.,
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", expected, reply)
		}
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// reply here is an interface type(singleResultDispatcher), it exists
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDisaptcherCacheClear(t *testing.T) {
	var reply string
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testDispatcherAdminCoreStatusWithRouteIDButHost1(t *testing.T) {
	var reply map[string]any
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminCheckCacheAfterRouting(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply any
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]any{
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]any{
			utils.FilterIDs: nil,
			utils.Hosts: []any{
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST1",
					utils.Params:    nil,
					utils.Weight:    10.,
				},
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:             "DSP1",
			utils.Strategy:       "*weight",
			utils.StrategyParams: nil,
			utils.Tenant:         "cgrates.org",
			utils.Weight:         10.,
		}
		if !reflect.DeepEqual(expected, reply) {
			t.Errorf("Expected %+v, \n received %+v", expected, reply)
		}
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// reply here is an interface type(singleResultDispatcher), it exists
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDispatcherSetDispatcherProfileOverwrite(t *testing.T) {
	// as the cache was cleard, now that previously the HOST1 was matched, setting the profile wiht only HOST2 will remove the
	// DispatcherRoutes, DispatcherProfile and the DispatcherInstance
	var replyStr string
	// Set DispatcherProfile
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:   "cgrates.org",
			ID:       "DSP1",
			Strategy: "*weight",
			Weight:   10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST2",
					Weight: 5,
				},
			},
		},
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherCheckCacheAfterSetDispatcherDSP1(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply any // Should receive NOT_FOUND, as CallCache that was called in API will remove the DispatcherRoute
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// as the DSP1 profile was overwritten, only HOST2 in profile will be contained
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	} else {
		expected := map[string]any{
			utils.FilterIDs: nil,
			utils.Hosts: []any{
				map[string]any{
					utils.Blocker:   false,
					utils.FilterIDs: nil,
					utils.ID:        "HOST2",
					utils.Params:    nil,
					utils.Weight:    5.,
				},
			},
			utils.ID:             "DSP1",
			utils.Strategy:       "*weight",
			utils.StrategyParams: nil,
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// DispatcherInstance should also be removed, so it will be NOT_FOUND
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func testDispatcherSetAnotherProifle(t *testing.T) {
	var replyStr string
	// Set DispatcherProfile DSP2 with the existing hosts
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:   "cgrates.org",
			ID:       "DSP2",
			Strategy: "*weight",
			Weight:   20,
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
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
	}
	if err := adminsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherCheckCacheAfterSetDispatcherDSP2(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply any
	// NOT_FOUND
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP2",
		},
	}
	// NOT_FOUND
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]any{
			utils.MetaDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP2",
		},
	}
	// NOT_FOUND
	if err := adminsRPC.Call(context.Background(), utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func testDispatcherOptsDSPStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAdminStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
