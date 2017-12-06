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
	alsPrfCfgPath   string
	alsPrfCfg       *config.CGRConfig
	alsPrfRPC       *rpc.Client
	alsPrfDataDir   = "/usr/share/cgrates"
	alsPrf          *engine.ExternalAliasProfile
	alsPrfDelay     int
	alsPrfConfigDIR string //run tests for specific configuration
)

var sTestsAlsPrf = []func(t *testing.T){
	testAlsPrfInitCfg,
	testAlsPrfInitDataDb,
	testAlsPrfResetStorDb,
	testAlsPrfStartEngine,
	testAlsPrfRPCConn,
	testAlsPrfGetAlsPrfBeforeSet,
	testAlsPrfSetAlsPrf,
	testAlsPrfUpdateAlsPrf,
	testAlsPrfRemAlsPrf,
	testAlsPrfKillEngine,
}

//Test start here
func TestAlsPrfITMySql(t *testing.T) {
	alsPrfConfigDIR = "tutmysql"
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func TestAlsPrfITMongo(t *testing.T) {
	alsPrfConfigDIR = "tutmongo"
	for _, stest := range sTestsAlsPrf {
		t.Run(alsPrfConfigDIR, stest)
	}
}

func testAlsPrfInitCfg(t *testing.T) {
	var err error
	alsPrfCfgPath = path.Join(alsPrfDataDir, "conf", "samples", alsPrfConfigDIR)
	alsPrfCfg, err = config.NewCGRConfigFromFolder(alsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	alsPrfCfg.DataFolderPath = alsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(alsPrfCfg)
	switch alsPrfConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		alsPrfDelay = 2000
	default:
		alsPrfDelay = 1000
	}
}

func testAlsPrfInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testAlsPrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(alsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testAlsPrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(alsPrfCfgPath, alsPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testAlsPrfRPCConn(t *testing.T) {
	var err error
	alsPrfRPC, err = jsonrpc.Dial("tcp", alsPrfCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testAlsPrfGetAlsPrfBeforeSet(t *testing.T) {
	var reply *engine.ExternalAliasProfile
	if err := alsPrfRPC.Call("ApierV1.GetAliasProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAlsPrfSetAlsPrf(t *testing.T) {
	alsPrf = &engine.ExternalAliasProfile{
		Tenant:    "cgrates.org",
		ID:        "ALS1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC).Local(),
		},
		Aliases: []*engine.AliasEntry{
			&engine.AliasEntry{
				FieldName: "FL1",
				Initial:   "In1",
				Alias:     "Al1",
			},
		},
		Weight: 20,
	}
	var result string
	if err := alsPrfRPC.Call("ApierV1.SetAliasProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ExternalAliasProfile
	if err := alsPrfRPC.Call("ApierV1.GetAliasProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf), utils.ToJSON(reply))
	}
}

func testAlsPrfUpdateAlsPrf(t *testing.T) {
	alsPrf.Aliases = []*engine.AliasEntry{
		&engine.AliasEntry{
			FieldName: "FL1",
			Initial:   "In1",
			Alias:     "Al1",
		},
		&engine.AliasEntry{
			FieldName: "FL2",
			Initial:   "In2",
			Alias:     "Al2",
		},
	}
	var result string
	if err := alsPrfRPC.Call("ApierV1.SetAliasProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ExternalAliasProfile
	if err := alsPrfRPC.Call("ApierV1.GetAliasProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(alsPrf), utils.ToJSON(reply))
	}
}

func testAlsPrfRemAlsPrf(t *testing.T) {
	var resp string
	if err := alsPrfRPC.Call("ApierV1.RemAliasProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.ExternalAliasProfile
	if err := alsPrfRPC.Call("ApierV1.GetAliasProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ALS1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testAlsPrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(alsPrfDelay); err != nil {
		t.Error(err)
	}
}
