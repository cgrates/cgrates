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
	dspOptsCfg       *config.CGRConfig
	dspOptsRPC       *birpc.Client
	dspOptsConfigDIR string
	dpsOptsTest      = []func(t *testing.T){
		// Start engine without Dispatcher on engine 4012
		testDispatcherOptsAdminInitCfg,
		testDispatcherOptsAdminInitDataDb,
		testDispatcherOptsAdminStartEngine,
		testDispatcherOptsAdminRPCConn,
		testDispatcherOptsAdminSetDispatcherProfile,
		testDispatcherOptsAdminStopEngine,

		// Start engine without Dispatcher on engine 2012 with profiles in database
		testDispatcherOptsDSPInitCfg,
		testDispatcherOptsDSPStartEngine,
		testDispatcherOptsDSPRPCConn,
		testDispatcherOptsCoreStatus,
		testDispatcherOptsDSPStopEngine,
	}
)

func TestDispatcherOpts(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts_admin"
	for _, test := range dpsOptsTest {
		t.Run(dspOptsConfigDIR, test)
	}
}

func testDispatcherOptsAdminInitCfg(t *testing.T) {
	var err error
	dspOptsCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	dspOptsCfg, err = config.NewCGRConfigFromPath(context.Background(), dspOptsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsAdminInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(dspOptsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsAdminStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dspOptsCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAdminRPCConn(t *testing.T) {
	var err error
	dspOptsRPC, err = newRPCClient(dspOptsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherOptsAdminSetDispatcherProfile(t *testing.T) {
	// Set DispatcherHost
	var replyStr string
	setDispatcherHost := &engine.DispatcherHostWithAPIOpts{
		DispatcherHost: &engine.DispatcherHost{
			Tenant: "cgrates.org",
			RemoteHost: &config.RemoteHost{
				ID:              "SELF_ENGINE",
				Address:         "*internal",
				ConnectAttempts: 1,
				Reconnects:      3,
				ConnectTimeout:  time.Minute,
				ReplyTimeout:    2 * time.Minute,
			},
		},
	}
	if err := dspOptsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherHost, setDispatcherHost, &replyStr); err != nil {
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
					ID:     "SELF_ENGINE",
					Weight: 5,
				},
			},
		},
	}
	if err := dspOptsRPC.Call(context.Background(), utils.AdminSv1SetDispatcherProfile, setDispatcherProfile, &replyStr); err != nil {
		t.Error("Unexpected error when calling AdminSv1.SetDispatcherProfile: ", err)
	} else if replyStr != utils.OK {
		t.Error("Unexpected reply returned", replyStr)
	}
}

func testDispatcherOptsAdminStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}

func testDispatcherOptsDSPInitCfg(t *testing.T) {
	dspOptsConfigDIR = "dispatcher_opts" //changed with the cfg with dispatcher on
	var err error
	dspOptsCfgPath = path.Join(*dataDir, "conf", "samples", dspOptsConfigDIR)
	dspOptsCfg, err = config.NewCGRConfigFromPath(context.Background(), dspOptsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Start CGR Engine woth Dispatcher enabled
func testDispatcherOptsDSPStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(dspOptsCfgPath, *waitRater); err != nil {
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
	var reply map[string]interface{}
	ev := utils.TenantWithAPIOpts{
		Tenant: "cgrates.org",
	}
	if err := dspOptsRPC.Call(context.Background(), utils.CoreSv1Status, &ev, &reply); err != nil {
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
