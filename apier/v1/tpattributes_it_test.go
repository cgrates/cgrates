//go:build offline
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
	"strings"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpAlsPrfCfgPath   string
	tpAlsPrfCfg       *config.CGRConfig
	tpAlsPrfRPC       *rpc.Client
	tpAlsPrfDataDir   = "/usr/share/cgrates"
	tpAlsPrf          *utils.TPAttributeProfile
	tpAlsPrfDelay     int
	tpAlsPrfConfigDIR string //run tests for specific configuration
)

var sTestsTPAlsPrf = []func(t *testing.T){
	testTPAlsPrfInitCfg,
	testTPAlsPrfResetStorDb,
	testTPAlsPrfStartEngine,
	testTPAlsPrfRPCConn,
	testTPAlsPrfGetTPAlsPrfBeforeSet,
	testTPAlsPrfSetTPAlsPrf,
	testTPAlsPrfGetTPAlsPrfAfterSet,
	testTPAlsPrfGetTPAlsPrfIDs,
	testTPAlsPrfUpdateTPAlsPrf,
	testTPAlsPrfGetTPAlsPrfAfterUpdate,
	testTPAlsPrfRemTPAlsPrf,
	testTPAlsPrfGetTPAlsPrfAfterRemove,
	testTPAlsPrfKillEngine,
}

//Test start here
func TestTPAlsPrfIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpAlsPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpAlsPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpAlsPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPAlsPrf {
		t.Run(tpAlsPrfConfigDIR, stest)
	}
}

func testTPAlsPrfInitCfg(t *testing.T) {
	var err error
	tpAlsPrfCfgPath = path.Join(tpAlsPrfDataDir, "conf", "samples", tpAlsPrfConfigDIR)
	tpAlsPrfCfg, err = config.NewCGRConfigFromPath(tpAlsPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpAlsPrfCfg.DataFolderPath = tpAlsPrfDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpAlsPrfCfg)
	tpAlsPrfDelay = 1000
}

// Wipe out the cdr database
func testTPAlsPrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpAlsPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPAlsPrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpAlsPrfCfgPath, tpAlsPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPAlsPrfRPCConn(t *testing.T) {
	var err error
	tpAlsPrfRPC, err = jsonrpc.Dial(utils.TCP, tpAlsPrfCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPAlsPrfGetTPAlsPrfBeforeSet(t *testing.T) {
	var reply *utils.TPAttributeProfile
	if err := tpAlsPrfRPC.Call(utils.APIerSv1GetTPAttributeProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAlsPrfSetTPAlsPrf(t *testing.T) {
	tpAlsPrf = &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Attr1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Contexts: []string{"con1"},
		Attributes: []*utils.TPAttribute{
			&utils.TPAttribute{
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
				FilterIDs: []string{},
			},
		},
		Weight: 20,
	}
	sort.Strings(tpAlsPrf.FilterIDs)
	var result string
	if err := tpAlsPrfRPC.Call(utils.APIerSv1SetTPAttributeProfile, tpAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAlsPrfGetTPAlsPrfAfterSet(t *testing.T) {
	var reply *utils.TPAttributeProfile
	if err := tpAlsPrfRPC.Call(utils.APIerSv1GetTPAttributeProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	if !reflect.DeepEqual(tpAlsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpAlsPrf), utils.ToJSON(reply))
	}
}

func testTPAlsPrfGetTPAlsPrfIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:Attr1"}
	if err := tpAlsPrfRPC.Call(utils.APIerSv1GetTPAttributeProfileIds,
		&AttrGetTPAttributeProfileIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAlsPrfUpdateTPAlsPrf(t *testing.T) {
	tpAlsPrf.Attributes = []*utils.TPAttribute{
		&utils.TPAttribute{
			Path:      utils.MetaReq + utils.NestingSep + "FL1",
			Value:     "Al1",
			FilterIDs: []string{},
		},
		&utils.TPAttribute{
			Path:      utils.MetaReq + utils.NestingSep + "FL2",
			Value:     "Al2",
			FilterIDs: []string{},
		},
	}
	var result string
	if err := tpAlsPrfRPC.Call(utils.APIerSv1SetTPAttributeProfile, tpAlsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAlsPrfGetTPAlsPrfAfterUpdate(t *testing.T) {
	var reply *utils.TPAttributeProfile
	revTPAlsPrf := &utils.TPAttributeProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "Attr1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Contexts: []string{"con1"},
		Attributes: []*utils.TPAttribute{
			&utils.TPAttribute{
				Path:      utils.MetaReq + utils.NestingSep + "FL2",
				Value:     "Al2",
				FilterIDs: []string{},
			},
			&utils.TPAttribute{
				Path:      utils.MetaReq + utils.NestingSep + "FL1",
				Value:     "Al1",
				FilterIDs: []string{},
			},
		},
		Weight: 20,
	}
	sort.Strings(revTPAlsPrf.FilterIDs)
	sort.Slice(revTPAlsPrf.Attributes, func(i, j int) bool {
		return strings.Compare(revTPAlsPrf.Attributes[i].Path, revTPAlsPrf.Attributes[j].Path) == -1
	})
	if err := tpAlsPrfRPC.Call(utils.APIerSv1GetTPAttributeProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"}, &reply); err != nil {
		t.Fatal(err)
	}
	sort.Strings(reply.FilterIDs)
	sort.Slice(reply.Attributes, func(i, j int) bool {
		return strings.Compare(reply.Attributes[i].Path, reply.Attributes[j].Path) == -1
	})
	if !reflect.DeepEqual(tpAlsPrf, reply) && !reflect.DeepEqual(revTPAlsPrf, reply) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(tpAlsPrf), utils.ToJSON(reply))
	}
}

func testTPAlsPrfRemTPAlsPrf(t *testing.T) {
	var resp string
	if err := tpAlsPrfRPC.Call(utils.APIerSv1RemoveTPAttributeProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPAlsPrfGetTPAlsPrfAfterRemove(t *testing.T) {
	var reply *utils.TPAttributeProfile
	if err := tpAlsPrfRPC.Call(utils.APIerSv1GetTPAttributeProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "Attr1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAlsPrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpAlsPrfDelay); err != nil {
		t.Error(err)
	}
}
