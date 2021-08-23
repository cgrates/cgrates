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
	dispatcherProfile   *DispatcherWithCache
	dispatcherHost      *DispatcherHostWithCache
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

		testDispatcherSSetDispatcherHost,
		testDispatcherSGetDispatcherHostIDs,
		testDispatcherSUpdateDispatcherHost,
		testDispatcherSGetDispatcherHostCache,
		testDispatcherSRemDispatcherHost,

		testDispatcherSKillEngine,
	}
)

//Test start here
func TestDispatcherSIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		dispatcherConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		dispatcherConfigDIR = "tutmysql"
	case utils.MetaMongo:
		dispatcherConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsDispatcher {
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
	if err := engine.InitDataDb(dispatcherCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testDispatcherSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(dispatcherCfg); err != nil {
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
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherProfile = &DispatcherWithCache{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:    "cgrates.org",
			ID:        "Dsp1",
			FilterIDs: []string{"*string:~*req.Account:1001"},
			Strategy:  utils.MetaFirst,
			Weight:    20,
		},
	}

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
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfileIDs,
		utils.TenantArgWithPaginator{}, &result); err == nil {
		t.Errorf("Expected: %s , received: %v", utils.NewErrMandatoryIeMissing(utils.Tenant).Error(), err)
	}
	expected := []string{"Dsp1"}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherProfileIDs,
		utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: dispatcherProfile.Tenant}}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testDispatcherSUpdateDispatcherProfile(t *testing.T) {
	var result string
	dispatcherProfile.Strategy = utils.MetaWeight
	dispatcherProfile.Subsystems = []string{utils.MetaAttributes, utils.MetaSessionS, utils.MetaRALs, utils.MetaCDRs}
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
	if err := dispatcherRPC.Call(utils.CacheSv1GetCacheStats, &utils.AttrCacheIDsWithArgDispatcher{}, &rcvStats); err != nil {
		t.Error(err)
	} else if rcvStats[utils.CacheDispatcherProfiles].Items != 1 {
		t.Errorf("Expecting: 1 DispatcherProfiles, received: %+v", rcvStats[utils.CacheDispatcherProfiles])
	}
}

func testDispatcherSRemDispatcherProfile(t *testing.T) {
	var result string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherProfile,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp1"},
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "Dsp1"},
		&result); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testDispatcherSSetDispatcherHost(t *testing.T) {
	var reply string
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHost,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherHost = &DispatcherHostWithCache{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			ID:     "DspHst1",
			Conns: []*config.RemoteHost{
				&config.RemoteHost{
					Address: "*internal",
				},
				&config.RemoteHost{
					Address:   ":2012",
					Transport: utils.MetaJSON,
					TLS:       true,
				},
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
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHostIDs,
		utils.TenantArgWithPaginator{}, &result); err == nil {
		t.Errorf("Expected: %s , received: %v", utils.NewErrMandatoryIeMissing(utils.Tenant), err)
	}
	expected := []string{"DspHst1"}
	if err := dispatcherRPC.Call(utils.APIerSv1GetDispatcherHostIDs,
		utils.TenantArgWithPaginator{TenantArg: utils.TenantArg{Tenant: dispatcherHost.Tenant}}, &result); err != nil {
		t.Error(err)
	} else if len(result) != len(expected) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testDispatcherSUpdateDispatcherHost(t *testing.T) {
	var result string
	dispatcherHost.Conns = append(dispatcherHost.Conns, &config.RemoteHost{
		Address:   ":4012",
		Transport: utils.MetaGOB,
		TLS:       false,
	})
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
	var rcvStats map[string]*ltcache.CacheStats
	if err := dispatcherRPC.Call(utils.CacheSv1GetCacheStats, &utils.AttrCacheIDsWithArgDispatcher{}, &rcvStats); err != nil {
		t.Error(err)
	} else if rcvStats[utils.CacheDispatcherHosts].Items != 0 {
		t.Errorf("Expecting: 0 DispatcherProfiles, received: %+v", rcvStats[utils.CacheDispatcherProfiles])
	}
}

func testDispatcherSRemDispatcherHost(t *testing.T) {
	var result string
	if err := dispatcherRPC.Call(utils.APIerSv1RemoveDispatcherHost,
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "DspHst1"},
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
		&utils.TenantIDWithCache{Tenant: "cgrates.org", ID: "DspHst1"},
		&result); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testDispatcherSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
