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
	dispatcherCfgPath   string
	dispatcherCfg       *config.CGRConfig
	dispatcherRPC       *rpc.Client
	dispatcherProfile   *DispatcherWithAPIOpts
	dispatcherHost      *engine.DispatcherHostWithAPIOpts
	dispatcherConfigDIR string //run tests for specific configuration

	sTestsDispatcher = []func(t *testing.T){

		testDispatcherSInitCfg,
		testDispatcherSInitDataDb,
		testDispatcherSResetStorDb,
		testDispatcherSStartEngine,
		testDispatcherSRPCConn,

		testDispatcherSSetDispatcherProfile,
		testDispatcherSGetDispatcherProfileIDs,
		testDispatcherSUpdateDispatcherProfile,
		testDispatcherSGetDispatcherProfileCache,
		testDispatcherSRemDispatcherProfile,
		testDispatcherSSetDispatcherProfileWithoutTenant,
		testDispatcherSRemDispatcherProfileWithoutTenant,

		testDispatcherSSetDispatcherHost,
		testDispatcherSGetDispatcherHostIDs,
		testDispatcherSUpdateDispatcherHost,
		testDispatcherSGetDispatcherHostCache,
		testDispatcherSRemDispatcherHost,
		testDispatcherSSetDispatcherHostWithoutTenant,
		testDispatcherSRemDispatcherHostWithoutTenant,

		testDispatcherSKillEngine,

		//cache test
		testDispatcherSInitCfg,
		testDispatcherSInitDataDb,
		testDispatcherSResetStorDb,
		testDispatcherSStartEngine,
		testDispatcherSRPCConn,
		testDispatcherSCacheTestGetNotFound,
		testDispatcherSCacheTestSet,
		testDispatcherSCacheTestGetNotFound,
		testDispatcherSCacheReload,
		testDispatcherSCacheTestGetFound,
		testDispatcherSKillEngine,
	}
)

//Test start here
func TestDispatcherSIT(t *testing.T) {
	sTestsDispatcherCacheSV1 := sTestsDispatcher
	switch *dbType {
	case utils.MetaInternal:
		dispatcherConfigDIR = "tutinternal"
		sTestsDispatcherCacheSV1 = sTestsDispatcherCacheSV1[:len(sTestsDispatcherCacheSV1)-10]
	case utils.MetaMySQL:
		dispatcherConfigDIR = "tutmysql"
	case utils.MetaMongo:
		dispatcherConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDispatcherCacheSV1 {
		t.Run(dispatcherConfigDIR, stest)
	}
}

func testDispatcherSInitCfg(t *testing.T) {
	var err error
	dispatcherCfgPath = path.Join(*dataDir, "conf", "samples", dispatcherConfigDIR)
	dispatcherCfg, err = config.NewCGRConfigFromPath(dispatcherCfgPath)
	if err != nil {
		t.Error(err)
	}
	dispatcherCfg.DataFolderPath = *dataDir
}

func testDispatcherSInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(dispatcherCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDispatcherSResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(dispatcherCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testDispatcherSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dispatcherCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testDispatcherSRPCConn(t *testing.T) {
	var err error
	dispatcherRPC, err = newRPCClient(dispatcherCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherSSetDispatcherProfile(t *testing.T) {
	var reply string
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*wrong:inline"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
		},
	}

	expErr := "SERVER_ERROR: broken reference to filter: *wrong:inline for item with ID: cgrates.org:Dsp1"
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err == nil || err.Error() != expErr {
		t.Fatalf("Expected error: %q, received: %v", expErr, err)
	}

	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherProfile.FilterIDs = []string{"*string:~*req.Account:1001"}

	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile.DispatcherProfile, dsp)
	}
}

func testDispatcherSGetDispatcherProfileIDs(t *testing.T) {
	var result []string
	expected := []string{"Dsp1"}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfileIDs,
		&utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfileIDs,
		&utils.PaginatorWithTenant{Tenant: dispatcherProfile.Tenant}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testDispatcherSUpdateDispatcherProfile(t *testing.T) {
	var result string
	dispatcherProfile.Strategy = utils.MetaWeight
	dispatcherProfile.Subsystems = []string{utils.MetaAttributes, utils.MetaSessionS, utils.MetaCDRs}
	dispatcherProfile.ActivationInterval = &utils.ActivationInterval{
		ActivationTime: time.Date(2019, 3, 1, 0, 0, 0, 0, time.UTC),
		ExpiryTime:     time.Date(2019, 4, 1, 0, 0, 0, 0, time.UTC),
	}
	dispatcherProfile.Hosts = engine.DispatcherHostProfiles{
		&engine.DispatcherHostProfile{ID: "HOST1", Weight: 20.0},
		&engine.DispatcherHostProfile{ID: "HOST2", Weight: 10.0},
	}
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile.DispatcherProfile, dsp)
	}
}

func testDispatcherSGetDispatcherProfileCache(t *testing.T) {
	if dispatcherConfigDIR == "tutinternal" {
		t.SkipNow()
	}
	var rcvStats map[string]*ltcache.CacheStats
	if err := dispatcherRPC.Call(utils.CacheSv1GetCacheStats, &utils.AttrCacheIDsWithAPIOpts{}, &rcvStats); err != nil {
		t.Error(err)
	} else if rcvStats[utils.CacheDispatcherProfiles].Items != 1 {
		t.Errorf("Expecting: 1 DispatcherProfiles, received: %+v", rcvStats[utils.CacheDispatcherProfiles])
	}
}

func testDispatcherSRemDispatcherProfile(t *testing.T) {
	var result string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"}},
		&result); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testDispatcherSSetDispatcherHost(t *testing.T) {
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherHost = &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:      "DspHst1",
				Address: "*internal",
			},
		},
	}

	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherHost,
		dispatcherHost,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	var dsp *engine.DispatcherHost
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherHost.DispatcherHost, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherHost.DispatcherHost, dsp)
	}
}

func testDispatcherSGetDispatcherHostIDs(t *testing.T) {
	var result []string
	expected := []string{"DspHst1"}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHostIDs,
		&utils.PaginatorWithTenant{}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHostIDs,
		&utils.PaginatorWithTenant{Tenant: dispatcherHost.Tenant}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testDispatcherSUpdateDispatcherHost(t *testing.T) {
	var result string
	dispatcherHost.RemoteHost = &config.RemoteHost{
		ID:        "DspHst1",
		Address:   ":4012",
		Transport: utils.MetaGOB,
		TLS:       false,
	}
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherHost,
		dispatcherHost, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherHost
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherHost.DispatcherHost, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherHost.DispatcherHost, dsp)
	}
}

func testDispatcherSGetDispatcherHostCache(t *testing.T) {
	if dispatcherConfigDIR == "tutinternal" {
		t.SkipNow()
	}
	var rcvStats map[string]*ltcache.CacheStats
	if err := dispatcherRPC.Call(utils.CacheSv1GetCacheStats, &utils.AttrCacheIDsWithAPIOpts{}, &rcvStats); err != nil {
		t.Error(err)
	} else if rcvStats[utils.CacheDispatcherHosts].Items != 0 {
		t.Errorf("Expecting: 0 DispatcherProfiles, received: %+v", rcvStats[utils.CacheDispatcherHosts])
	}
}

func testDispatcherSRemDispatcherHost(t *testing.T) {
	var result string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherHost,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"}},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherHost
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"},
		&dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherHost,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "DspHst1"}},
		&result); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v received: %v", utils.ErrNotFound, err)
	}
}

func testDispatcherSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherSSetDispatcherProfileWithoutTenant(t *testing.T) {
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
		},
	}
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherProfile, dispatcherProfile, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	dispatcherProfile.DispatcherProfile.Tenant = "cgrates.org"
	var result *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{ID: "Dsp1"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile.DispatcherProfile, result) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(dispatcherProfile.DispatcherProfile), utils.ToJSON(result))
	}
}

func testDispatcherSRemDispatcherProfileWithoutTenant(t *testing.T) {
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "Dsp1"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{ID: "Dsp1"},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherSSetDispatcherHostWithoutTenant(t *testing.T) {
	dispatcherHost = &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			RemoteHost: &config.RemoteHost{
				ID:      "DspHst7",
				Address: "*internal",
			},
		},
	}
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherHost, dispatcherHost, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	dispatcherHost.DispatcherHost.Tenant = "cgrates.org"
	var result *engine.DispatcherHost
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{ID: "DspHst7"},
		&result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(result, dispatcherHost.DispatcherHost) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(dispatcherHost.DispatcherHost), utils.ToJSON(result))
	}
}

func testDispatcherSRemDispatcherHostWithoutTenant(t *testing.T) {
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherHost,
		&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{ID: "DspHst7"}},
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	var result *engine.DispatcherHost
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{ID: "DspHst7"},
		&result); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherSCacheTestGetNotFound(t *testing.T) {
	var suplsReply *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "DISPATCHER_CACHE",
		}, &suplsReply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testDispatcherSCacheTestGetFound(t *testing.T) {
	var suplsReply *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{
			Tenant: "cgrates.org",
			ID:     "DISPATCHER_CACHE",
		}, &suplsReply); err != nil {
		t.Error(err)
	}
}

func testDispatcherSCacheTestSet(t *testing.T) {
	var reply string
	dispatcherProfile = &DispatcherWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant: "cgrates.org",
			ID:     "DISPATCHER_CACHE",
		},
		APIOpts: map[string]interface{}{
			utils.CacheOpt: utils.MetaNone,
		},
	}
	if err := dispatcherRPC.Call(utils.APIerSv1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

}

func testDispatcherSCacheReload(t *testing.T) {
	cache := &utils.AttrReloadCacheWithAPIOpts{
		ArgsCache: map[string][]string{
			utils.DispatcherProfileIDs: {"cgrates.org:DISPATCHER_CACHE"},
		},
	}
	var reply string
	if err := dispatcherRPC.Call(utils.CacheSv1ReloadCache, cache, &reply); err != nil {
		t.Error("Got error on CacheSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling CacheSv1.ReloadCache got reply: ", reply)
	}
}
