// +build offline

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
	"sort"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpResCfgPath   string
	tpResCfg       *config.CGRConfig
	tpResRPC       *rpc.Client
	tpRes          *utils.TPResourceProfile
	tpResDelay     int
	tpResConfigDIR string //run tests for specific configuration
)

var sTestsTPResources = []func(t *testing.T){
	testTPResInitCfg,
	testTPResResetStorDb,
	testTPResStartEngine,
	testTPResRpcConn,
	testTPResGetTPResourceBeforeSet,
	testTPResSetTPResource,
	testTPResGetTPResourceAfterSet,
	testTPResUpdateTPResource,
	testTPResGetTPResourceAfterUpdate,
	testTPResRemoveTPResource,
	testTPResGetTPResourceAfterRemove,
	testTPResKillEngine,
}

//Test start here
func TestTPResIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpResConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpResConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpResConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		tpResConfigDIR = "tutpostgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPResources {
		t.Run(tpResConfigDIR, stest)
	}
}

func testTPResInitCfg(t *testing.T) {
	var err error
	tpResCfgPath = path.Join(*dataDir, "conf", "samples", tpResConfigDIR)
	tpResCfg, err = config.NewCGRConfigFromPath(tpResCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpResDelay = 1000
}

// Wipe out the cdr database
func testTPResResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(tpResCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPResStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpResCfgPath, tpResDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPResRpcConn(t *testing.T) {
	var err error
	tpResRPC, err = jsonrpc.Dial(utils.TCP, tpResCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPResGetTPResourceBeforeSet(t *testing.T) {
	var reply *utils.TPResourceProfile
	if err := tpResRPC.Call(utils.APIerSv1GetTPResource,
		&utils.TPTntID{TPid: "TPR1", Tenant: "cgrates.org", ID: "ResGroup1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPResSetTPResource(t *testing.T) {
	tpRes = &utils.TPResourceProfile{
		Tenant:    "cgrates.org",
		TPid:      "TPR1",
		ID:        "ResGroup1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		UsageTTL:          "1s",
		Limit:             "7",
		AllocationMessage: "",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"ValOne", "ValTwo"},
	}
	sort.Strings(tpRes.ThresholdIDs)
	var result string
	if err := tpResRPC.Call(utils.APIerSv1SetTPResource, tpRes, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPResGetTPResourceAfterSet(t *testing.T) {
	var respond *utils.TPResourceProfile
	if err := tpResRPC.Call(utils.APIerSv1GetTPResource, &utils.TPTntID{TPid: "TPR1", Tenant: "cgrates.org", ID: "ResGroup1"},
		&respond); err != nil {
		t.Fatal(err)
	}
	sort.Strings(respond.ThresholdIDs)
	if !reflect.DeepEqual(tpRes, respond) {
		t.Errorf("Expecting : %+v, received: %+v", tpRes, respond)
	}
}

func testTPResUpdateTPResource(t *testing.T) {
	var result string
	tpRes.FilterIDs = []string{"FLTR_1", "FLTR_STS1"}
	sort.Strings(tpRes.FilterIDs)
	if err := tpResRPC.Call(utils.APIerSv1SetTPResource, tpRes, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPResGetTPResourceAfterUpdate(t *testing.T) {
	var expectedTPR *utils.TPResourceProfile
	if err := tpResRPC.Call(utils.APIerSv1GetTPResource, &utils.TPTntID{TPid: "TPR1", Tenant: "cgrates.org", ID: "ResGroup1"},
		&expectedTPR); err != nil {
		t.Fatal(err)
	}
	sort.Strings(expectedTPR.FilterIDs)
	sort.Strings(expectedTPR.ThresholdIDs)
	if !reflect.DeepEqual(tpRes, expectedTPR) {
		t.Errorf("Expecting: %+v, received: %+v", tpRes, expectedTPR)
	}
}

func testTPResRemoveTPResource(t *testing.T) {
	var resp string
	if err := tpResRPC.Call(utils.APIerSv1RemoveTPResource, &utils.TPTntID{TPid: "TPR1", Tenant: "cgrates.org", ID: "ResGroup1"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPResGetTPResourceAfterRemove(t *testing.T) {
	var respond *utils.TPResourceProfile
	if err := tpResRPC.Call(utils.APIerSv1GetTPResource, &utils.TPTntID{TPid: "TPR1", Tenant: "cgrates.org", ID: "ResGroup1"},
		&respond); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPResKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpResDelay); err != nil {
		t.Error(err)
	}
}
