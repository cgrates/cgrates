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
	setterCfgPath    string
	dspOptsCfgPath   string
	apierCfgPath     string
	setterCfg        *config.CGRConfig
	dspOptsCfg       *config.CGRConfig
	apierCfg         *config.CGRConfig
	setterRPC        *rpc.Client
	dspOptsRPC       *rpc.Client
	apierRPC         *rpc.Client
	dspOptsConfigDIR string
	dpsOptsTest      = []func(t *testing.T){
		// FIRST APRT OF THE TEST
		// Start engine without Dispatcher on engine 4012
		testDispatcherOptsAdminInitCfg,
		testDispatcherOptsAdminInitDataDb,
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

		testDispatcherSetterInitCfg,
		testDispatcherSetterStartEngine,
		testDispatcherSetterRPCConn,

		testDispatcherOptsAdminStartEngine,
		testDispatcherOptsAdminRPCConn,

		testDispatcherOptsSetterSetDispatcherProfile, // contains both hosts, HOST1 prio, host2 backup

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
		testDispatcherSetterSetDispatcherProfileOverwrite,
		testDispatcherCheckCacheAfterSetDispatcherDSP1,
		/* testDispatcherSetterSetAnotherProifle,          //DSP2
		testDispatcherCheckCacheAfterSetDispatcherDSP1, //we set DSP2, so for DSP1 nothing changed
		testDispatcherCheckCacheAfterSetDispatcherDSP2, */ //NOT_FOUND for every get, cause it was not used that profile before

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
	dspOptsConfigDIR = "dispatcher_opts_apier"
	var err error
	apierCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAdminInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsAdminStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(apierCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAdminRPCConn(t *testing.T) {
	var err error
	apierRPC, err = newRPCClient(apierCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsDSPInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts" //changed with the cfg with dispatcher on
	var err error
	dspOptsCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	dspOptsCfg, err = config.NewCGRConfigFromPath(dspOptsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsDSPStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(dspOptsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsDSPRPCConn(t *testing.T) {
	var err error
	dspOptsRPC, err = newRPCClient(dspOptsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsCoreStatus(t *testing.T) {
	// HOST1 host matched
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminCoreStatus(t *testing.T) {
	// HOST2 host matched because it was called from engine with port :4012 -> host2
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID:     "account#dan.bogos",
			utils.OptsDispatchers: false,
		},
	}
	if err := apierRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
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
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	if err := dspOptsRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatcher_profiles
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := dspOptsRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// get for *dispatchers
	argsCache = &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := dspOptsRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherSetterInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts_setter"
	var err error
	setterCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
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

func testDispatcherOptsSetterSetDispatcherProfile(t *testing.T) {
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
			utils.OptsDispatchers: false,
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
			utils.OptsDispatchers: false,
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
			utils.OptsDispatchers: false,
		},
	}
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherAdminCoreStatusWithRouteID(t *testing.T) {
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := apierRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST2" {
		t.Errorf("Expected HOST2, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminGetItemHOST2(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{}
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherProfiles,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
			utils.OptsDispatchers: false,
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// reply here is an interface type(singleResultDispatcher), it exists
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDisaptcherCacheClear(t *testing.T) {
	var reply string
	if err := apierRPC.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	if err := dspOptsRPC.Call(utils.CacheSv1Clear, &utils.AttrCacheIDsWithAPIOpts{
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testDispatcherAdminCoreStatusWithRouteIDButHost1(t *testing.T) {
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := apierRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else if reply[utils.NodeID] != "HOST1" {
		t.Errorf("Expected HOST1, received %v", reply[utils.NodeID])
	}
}

func testDispatcherAdminCheckCacheAfterRouting(t *testing.T) {
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err != nil {
		t.Error(err)
	}
}

func testDispatcherSetterSetDispatcherProfileOverwrite(t *testing.T) {
	// as the cache was cleard, now that previously the HOST1 was matched, setting the profile wiht only HOST2 will remove the
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
}

func testDispatcherCheckCacheAfterSetDispatcherDSP1(t *testing.T) {
	// get for *dispatcher_routes
	argsCache := &utils.ArgsGetCacheItemWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.MetaDispatchers: false,
			"adi":                 "nu",
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatcherRoutes,
			ItemID:  "account#dan.bogos:*core",
		},
	}
	var reply interface{} // Should receive NOT_FOUND, as CallCache that was called in API will remove the DispatcherRoute
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}

	/* // get for *dispatcher_profiles
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
	// as the DSP1 profile was overwritten, only HOST2 in profile will be contained
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
		},
		ArgsGetCacheItem: utils.ArgsGetCacheItem{
			CacheID: utils.CacheDispatchers,
			ItemID:  "cgrates.org:DSP1",
		},
	}
	// DispatcherInstance should also be removed, so it will be NOT_FOUND
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	} */
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
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
	if err := apierRPC.Call(utils.CacheSv1GetItemWithRemote, argsCache,
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Unexpected error returned: %v", err)
	}
}

func testDispatcherOptsDSPStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAdminStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
