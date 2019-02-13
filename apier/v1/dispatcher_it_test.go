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
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dispatcherCfgPath   string
	dispatcherCfg       *config.CGRConfig
	dispatcherRPC       *rpc.Client
	dispatcherProfile   *engine.DispatcherProfile
	dispatcherConfigDIR string //run tests for specific configuration
)

var sTestsDispatcher = []func(t *testing.T){
	testDispatcherSInitCfg,
	testDispatcherSInitDataDb,
	testDispatcherSResetStorDb,
	testDispatcherSStartEngine,
	testDispatcherSRPCConn,
	testDispatcherSSetDispatcherProfile,
	testDispatcherSGetDispatcherProfileIDs,
	testDispatcherSUpdateDispatcherProfile,
	testDispatcherSRemDispatcherProfile,
	testDispatcherSKillEngine,
}

//Test start here
func TestDispatcherSITMySql(t *testing.T) {
	dispatcherConfigDIR = "tutmysql"
	for _, stest := range sTestsDispatcher {
		t.Run(dispatcherConfigDIR, stest)
	}
}

func TestDispatcherSITMongo(t *testing.T) {
	dispatcherConfigDIR = "tutmongo"
	for _, stest := range sTestsDispatcher {
		t.Run(dispatcherConfigDIR, stest)
	}
}

func testDispatcherSInitCfg(t *testing.T) {
	var err error
	dispatcherCfgPath = path.Join(*dataDir, "conf", "samples", dispatcherConfigDIR)
	dispatcherCfg, err = config.NewCGRConfigFromFolder(dispatcherCfgPath)
	if err != nil {
		t.Error(err)
	}
	dispatcherCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(dispatcherCfg)
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
	dispatcherRPC, err = jsonrpc.Dial("tcp", dispatcherCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testDispatcherSSetDispatcherProfile(t *testing.T) {
	var reply string
	if err := dispatcherRPC.Call(utils.ApierV1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	dispatcherProfile = &engine.DispatcherProfile{
		Tenant:    "cgrates.org",
		ID:        "Dsp1",
		FilterIDs: []string{"*string:Account:1001"},
		Strategy:  utils.MetaFirst,
		// Hosts:     []string{"192.168.56.203", "192.168.56.204"},
		Weight: 20,
	}

	if err := dispatcherRPC.Call(utils.ApierV1SetDispatcherProfile,
		dispatcherProfile,
		&reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, reply)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.ApierV1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile, dsp)
	}
}

func testDispatcherSGetDispatcherProfileIDs(t *testing.T) {
	var result []string
	expected := []string{"Dsp1"}
	if err := dispatcherRPC.Call("ApierV1.GetDispatcherProfileIDs",
		dispatcherProfile.Tenant, &result); err != nil {
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
	dispatcherProfile.Conns = []*engine.DispatcherConn{
		{ID: "HOST1", Weight: 20.0},
		{ID: "HOST2", Weight: 10.0},
	}
	if err := dispatcherRPC.Call(utils.ApierV1SetDispatcherProfile,
		dispatcherProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.ApierV1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(dispatcherProfile, dsp) {
		t.Errorf("Expecting : %+v, received: %+v", dispatcherProfile, dsp)
	}
}

func testDispatcherSRemDispatcherProfile(t *testing.T) {
	var result string
	if err := dispatcherRPC.Call(utils.ApierV1RemoveDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Errorf("Expecting : %+v, received: %+v", utils.OK, result)
	}

	var dsp *engine.DispatcherProfile
	if err := dispatcherRPC.Call(utils.ApierV1GetDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&dsp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := dispatcherRPC.Call(utils.ApierV1RemoveDispatcherProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Dsp1"},
		&result); err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Expected error: %v recived: %v", utils.ErrNotFound, err)
	}
}

func testDispatcherSKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
