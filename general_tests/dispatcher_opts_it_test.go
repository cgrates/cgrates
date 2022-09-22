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
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	setterCfgPath string
	setterCfg     *config.CGRConfig
	setterRPC     *rpc.Client

	dspOptsCfgPath   string
	apierCfgPath     string
	dspOptsCfg       *config.CGRConfig
	apierCfg         *config.CGRConfig
	dspOptsRPC       *rpc.Client
	apierRPC         *rpc.Client
	dspOptsConfigDIR string
	dpsOptsTest      = []func(t *testing.T){
		testDispatcherOptsSetterInitCfg,
		testDispatcherOptsSetterInitDataDb,
		testDispatcherOptsSetterStartEngine,
		testDispatcherOptsSetterRPCConn,
		// Start engine without Dispatcher on engine 4012
		testDispatcherOptsAPIerInitCfg,
		testDispatcherOptsAPIerInitDataDb,
		testDispatcherOptsAPIerStartEngine,
		testDispatcherOptsAPIerRPCConn,
		testDispatcherOptsSetterSetDispatcherProfile,
		// testDispatcherOptsAPIerSetDispatcherProfile,

		// Start engine without Dispatcher on engine 2012 with profiles in database (*dispatchers:false)
		testDispatcherOptsDSPInitCfg,
		testDispatcherOptsDSPStartEngine,
		testDispatcherOptsDSPRPCConn,
		testDispatcherOptsCoreStatus, // self localhost(:2012) CoresV1Status

		testDispatcherOptsSetterSetDispatcherHost4012,
		// testDispatcherOptsAPIerSetDispatcherHost4012,
		testDispatcherOptsCoreStatusHost4012,

		testDispatcherOptsSetterSetDispatcherProfileDoubleHost,
		// testDispatcherOptsAPIerSetDispatcherProfileDoubleHost,
		testDispatcherOptsCoreStatusWithRouteID,

		testDispatcherOptsSetterSetDispatcherHostInexistent,
		// testDispatcherOptsAPIerSetDispatcherHostInexistent,
		testDispatcherOptsCoreStatusWithRouteID2,

		testDispatcherOptsCoreStatusWithoutRouteID,

		testDispatcherOptsDSPStopEngine,
		testDispatcherOptsAPIerStopEngine,
		// testDispatcherOptsSetterStopEngine,
	}
)

func TestDispatcherOpts(t *testing.T) {
	for _, test := range dpsOptsTest {
		t.Run(dspOptsConfigDIR, test)
	}
}

func testDispatcherOptsAPIerInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts_apier"
	var err error
	apierCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	apierCfg, err = config.NewCGRConfigFromPath(apierCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAPIerInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(apierCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsAPIerStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(apierCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAPIerRPCConn(t *testing.T) {
	var err error
	apierRPC, err = newRPCClient(apierCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAPIerSetDispatcherProfile(t *testing.T) {
	// Set DispatcherHost
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "SELF_ENGINE",
				Address:         "127.0.0.1:4012",
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
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
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
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "SELF_ENGINE",
					Weight: 5,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
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
	//SELF_ENGINE HOST
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else {
		/*
			t.Errorf("Received: %s", utils.ToJSON(reply))
		*/
	}
}

func testDispatcherOptsAPIerSetDispatcherHost4012(t *testing.T) {
	// Set DispatcherHost on 4012 host
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "HOST4012",
				Address:         "127.0.0.1:4012",
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
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
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
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST4012",
					Weight: 10,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherOptsCoreStatusHost4012(t *testing.T) {
	// status just for HOST4012
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else {
		/*
			t.Errorf("Received: %s", utils.ToJSON(reply))
		*/
	}
}

func testDispatcherOptsAPIerSetDispatcherProfileDoubleHost(t *testing.T) {
	// Set DispatcherProfile with both engines
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "SELF_ENGINE",
					Weight: 5,
				},
				{
					ID:     "HOST4012",
					Weight: 10,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	var replyStr string
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherOptsCoreStatusWithRouteID(t *testing.T) {
	// now it will dispatch in both engines
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else {
		/*
			t.Errorf("Received: %s", utils.ToJSON(reply))
		*/
	}
}

func testDispatcherOptsAPIerSetDispatcherHostInexistent(t *testing.T) {
	// Set DispatcherHost on 4012 host
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "INEXISTENT",
				Address:         "127.0.0.1:1223",
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
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherHost: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}

	// Set DispatcherProfile Different with an inexistent engine opened, but with a bigger weight(this should match now)
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Subsystems: []string{utils.MetaAny},
			Weight:     20,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "INEXISTENT",
					Weight: 10,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	if err := apierRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherOptsCoreStatusWithRouteID2(t *testing.T) {
	// because we have the routeID it will match DSP1 and last host matched, host4012
	// so again, both engines will match
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsRouteID: "account#dan.bogos",
		},
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else {
		/*
			t.Errorf("Received: %s", utils.ToJSON(reply))
		*/
	}
}

func testDispatcherOptsCoreStatusWithoutRouteID(t *testing.T) {
	// because we have the routeID it will match DSP1 and last host matched, host4012
	// so again, both engines will match
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}
	if err := dspOptsRPC.Call(utils.CoreSv1Status, &ev, &reply); err != nil {
		t.Error(err)
	} else {
		/*
			t.Errorf("Received: %s", utils.ToJSON(reply))
		*/
	}
}

func testDispatcherOptsDSPStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAPIerStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

// ----------------------------

func testDispatcherOptsSetterInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts_setter"
	var err error
	setterCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	setterCfg, err = config.NewCGRConfigFromPath(setterCfgPath)
	if err != nil {
		t.Error(err)
	}
}
func testDispatcherOptsSetterInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(setterCfg); err != nil {
		t.Fatal(err)
	}
}
func testDispatcherOptsSetterStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(setterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}
func testDispatcherOptsSetterRPCConn(t *testing.T) {
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
				ID:              "SELF_ENGINE",
				Address:         "127.0.0.1:4012",
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
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "SELF_ENGINE",
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
func testDispatcherOptsSetterSetDispatcherHost4012(t *testing.T) {
	// Set DispatcherHost on 4012 host
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "HOST4012",
				Address:         "127.0.0.1:4012",
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
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "HOST4012",
					Weight: 10,
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
func testDispatcherOptsSetterSetDispatcherProfileDoubleHost(t *testing.T) {
	// Set DispatcherProfile with both engines
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Subsystems: []string{utils.MetaAny},
			Weight:     10,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "SELF_ENGINE",
					Weight: 5,
				},
				{
					ID:     "HOST4012",
					Weight: 10,
				},
			},
		},
		APIOpts: map[string]interface{}{
			utils.OptsDispatchers: false,
		},
	}
	var replyStr string
	if err := setterRPC.Call(utils.APIerSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling APIerSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}
func testDispatcherOptsSetterSetDispatcherHostInexistent(t *testing.T) {
	// Set DispatcherHost on 4012 host
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "INEXISTENT",
				Address:         "127.0.0.1:1223",
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

	// Set DispatcherProfile Different with an inexistent engine opened, but with a bigger weight(this should match now)
	setDispatcherProfile := &engine.DispatcherProfileWithAPIOpts{
		DispatcherProfile: &engine.DispatcherProfile{
			Tenant:     "cgrates.org",
			ID:         "DSP1",
			Strategy:   "*weight",
			Subsystems: []string{utils.MetaAny},
			Weight:     20,
			Hosts: engine.DispatcherHostProfiles{
				{
					ID:     "INEXISTENT",
					Weight: 10,
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

func testDispatcherOptsSetterStopEngine(t *testing.T) {

}
