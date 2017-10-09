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
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"
)

var (
	filterProfileCfgPath   string
	filterProfileCfg       *config.CGRConfig
	filterProfileRPC       *rpc.Client
	filterProfileDataDir   = "/usr/share/cgrates"
	filterProfile          *engine.FilterProfile
	filterProfileDelay     int
	filterProfileConfigDIR string //run tests for specific configuration
)

var sTestsFilterProfile = []func(t *testing.T){
	testFilterProfileInitCfg,
	testFilterProfileResetDataDB,
	testFilterProfileStartEngine,
	testFilterProfileRpcConn,
	testFilterProfileGetFilterProfileBeforeSet,
	testFilterProfileSetFilterProfile,
	testFilterProfileGetFilterProfileAfterSet,
	testFilterProfileUpdateFilterProfile,
	testFilterProfileGetFilterProfileAfterUpdate,
	testFilterProfileRemFilterProfile,
	testFilterProfileGetFilterProfileAfterRemove,
	testFilterProfileKillEngine,
}

//Test start here
func TestFilterProfileITMySql(t *testing.T) {
	filterProfileConfigDIR = "tutmysql"
	for _, stest := range sTestsFilterProfile {
		t.Run(filterProfileConfigDIR, stest)
	}
}

func TestFilterProfileITMongo(t *testing.T) {
	filterProfileConfigDIR = "tutmongo"
	for _, stest := range sTestsFilterProfile {
		t.Run(filterProfileConfigDIR, stest)
	}
}

func TestFilterProfileITPG(t *testing.T) {
	filterProfileConfigDIR = "tutpostgres"
	for _, stest := range sTestsFilterProfile {
		t.Run(filterProfileConfigDIR, stest)
	}
}

func testFilterProfileInitCfg(t *testing.T) {
	var err error
	filterProfileCfgPath = path.Join(filterProfileDataDir, "conf", "samples", filterProfileConfigDIR)
	filterProfileCfg, err = config.NewCGRConfigFromFolder(filterProfileCfgPath)
	if err != nil {
		t.Error(err)
	}
	filterProfileCfg.DataFolderPath = filterProfileDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(filterProfileCfg)
	switch filterProfileConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		filterProfileDelay = 2000
	default:
		filterProfileDelay = 1000
	}
}

// Wipe out the cdr database
func testFilterProfileResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(filterProfileCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFilterProfileStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(filterProfileCfgPath, filterProfileDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testFilterProfileRpcConn(t *testing.T) {
	var err error
	filterProfileRPC, err = jsonrpc.Dial("tcp", filterProfileCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testFilterProfileGetFilterProfileBeforeSet(t *testing.T) {
	var reply *engine.FilterProfile
	if err := filterProfileRPC.Call("ApierV1.GetFilterProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterProfileSetFilterProfile(t *testing.T) {
	filterProfile = &engine.FilterProfile{
		Tenant:            "cgrates.org",
		ID:                "Filter1",
		FilterType:        "*string_prefix",
		FilterFieldName:   "Account",
		FilterFieldValues: []string{"10", "20"},
	}

	var result string
	if err := filterProfileRPC.Call("ApierV1.SetFilterProfile", filterProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterProfileGetFilterProfileAfterSet(t *testing.T) {
	var reply *engine.FilterProfile
	if err := filterProfileRPC.Call("ApierV1.GetFilterProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filterProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filterProfile, reply)
	}
}

func testFilterProfileUpdateFilterProfile(t *testing.T) {
	filterProfile = &engine.FilterProfile{
		Tenant:            "cgrates.org",
		ID:                "Filter1",
		FilterType:        "*string_prefix",
		FilterFieldName:   "Destination",
		FilterFieldValues: []string{"1001", "1002"},
	}
	var result string
	if err := filterProfileRPC.Call("ApierV1.SetFilterProfile", filterProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testFilterProfileGetFilterProfileAfterUpdate(t *testing.T) {
	var reply *engine.FilterProfile
	if err := filterProfileRPC.Call("ApierV1.GetFilterProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(filterProfile, reply) {
		t.Errorf("Expecting : %+v, received: %+v", filterProfile, reply)
	}
}

func testFilterProfileRemFilterProfile(t *testing.T) {
	var resp string
	if err := filterProfileRPC.Call("ApierV1.RemFilterProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testFilterProfileGetFilterProfileAfterRemove(t *testing.T) {
	var reply *engine.FilterProfile
	if err := filterProfileRPC.Call("ApierV1.GetFilterProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "Filter1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testFilterProfileKillEngine(t *testing.T) {
	if err := engine.KillEngine(filterProfileDelay); err != nil {
		t.Error(err)
	}
}
