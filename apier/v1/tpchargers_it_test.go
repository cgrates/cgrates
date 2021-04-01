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
	tpChrgsCfgPath   string
	tpChrgsCfg       *config.CGRConfig
	tpChrgsRPC       *rpc.Client
	tpChrgs          *utils.TPChargerProfile
	tpChrgsDelay     int
	tpChrgsConfigDIR string //run tests for specific configuration
)

var sTestsTPChrgs = []func(t *testing.T){
	testTPChrgsInitCfg,
	testTPChrgsResetStorDb,
	testTPChrgsStartEngine,
	testTPChrgsRPCConn,
	testTPChrgsGetTPChrgsBeforeSet,
	testTPChrgsSetTPChrgs,
	testTPChrgsGetTPChrgsAfterSet,
	testTPChrgsGetTPChrgsIDs,
	testTPChrgsUpdateTPChrgs,
	testTPChrgsGetTPChrgsAfterUpdate,
	testTPChrgsRemTPChrgs,
	testTPChrgsGetTPChrgsAfterRemove,
	testTPChrgsKillEngine,
}

//Test start here
func TestTPChrgsIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpChrgsConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpChrgsConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpChrgsConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPChrgs {
		t.Run(tpChrgsConfigDIR, stest)
	}
}

func testTPChrgsInitCfg(t *testing.T) {
	var err error
	tpChrgsCfgPath = path.Join(*dataDir, "conf", "samples", tpChrgsConfigDIR)
	tpChrgsCfg, err = config.NewCGRConfigFromPath(tpChrgsCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpChrgsDelay = 1000

}

// Wipe out the cdr database
func testTPChrgsResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(tpChrgsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPChrgsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpChrgsCfgPath, tpChrgsDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPChrgsRPCConn(t *testing.T) {
	var err error
	tpChrgsRPC, err = jsonrpc.Dial(utils.TCP, tpChrgsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPChrgsGetTPChrgsBeforeSet(t *testing.T) {
	var reply *utils.TPChargerProfile
	if err := tpChrgsRPC.Call(utils.APIerSv1GetTPCharger,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Chrgs"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPChrgsSetTPChrgs(t *testing.T) {
	tpChrgs = &utils.TPChargerProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Chrgs",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		RunID:        utils.MetaDefault,
		AttributeIDs: []string{"Attr1", "Attr2"},
		Weight:       20,
	}
	sort.Strings(tpChrgs.FilterIDs)
	sort.Strings(tpChrgs.AttributeIDs)
	var result string
	if err := tpChrgsRPC.Call(utils.APIerSv1SetTPCharger, tpChrgs, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPChrgsGetTPChrgsAfterSet(t *testing.T) {
	var reply *utils.TPChargerProfile
	if err := tpChrgsRPC.Call(utils.APIerSv1GetTPCharger,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Chrgs"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Strings(reply.AttributeIDs)
	if !reflect.DeepEqual(tpChrgs, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpChrgs), utils.ToJSON(reply))
	}
}

func testTPChrgsGetTPChrgsIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:Chrgs"}
	if err := tpChrgsRPC.Call(utils.APIerSv1GetTPChargerIDs,
		&AttrGetTPAttributeProfileIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(expectedTPID), utils.ToJSON(result))
	}
}

func testTPChrgsUpdateTPChrgs(t *testing.T) {
	tpChrgs.AttributeIDs = []string{"Attr1", "Attr2", "Attr3"}
	var result string
	if err := tpChrgsRPC.Call(utils.APIerSv1SetTPCharger, tpChrgs, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPChrgsGetTPChrgsAfterUpdate(t *testing.T) {
	var reply *utils.TPChargerProfile
	if err := tpChrgsRPC.Call(utils.APIerSv1GetTPCharger,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Chrgs"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Strings(reply.AttributeIDs)
	if !reflect.DeepEqual(tpChrgs, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpChrgs), utils.ToJSON(reply))
	}
}

func testTPChrgsRemTPChrgs(t *testing.T) {
	var resp string
	if err := tpChrgsRPC.Call(utils.APIerSv1RemoveTPCharger,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Chrgs"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPChrgsGetTPChrgsAfterRemove(t *testing.T) {
	var reply *utils.TPChargerProfile
	if err := tpChrgsRPC.Call(utils.APIerSv1GetTPCharger,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Chrgs"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPChrgsKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpChrgsDelay); err != nil {
		t.Error(err)
	}
}
