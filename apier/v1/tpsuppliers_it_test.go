// +build offline_tp

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
	tpSplPrfCfgPath   string
	tpSplPrfCfg       *config.CGRConfig
	tpSplPrfRPC       *rpc.Client
	tpSplPrfDataDire  = "/usr/share/cgrates"
	tpSplPr           *utils.TPSupplierProfile
	tpSplPrfDelay     int
	tpSplPrfConfigDIR string //run tests for specific configuration
)

var sTestsTPSplPrf = []func(t *testing.T){
	testTPSplPrfInitCfg,
	testTPSplPrfResetStorDb,
	testTPSplPrfStartEngine,
	testTPSplPrfRPCConn,
	testTPSplPrfGetTPSplPrfBeforeSet,
	testTPSplPrfSetTPSplPrf,
	testTPSplPrfGetTPSplPrfAfterSet,
	testTPSplPrfGetTPSplPrfIDs,
	testTPSplPrfUpdateTPSplPrf,
	testTPSplPrfGetTPSplPrfAfterUpdate,
	testTPSplPrfRemTPSplPrf,
	testTPSplPrfGetTPSplPrfAfterRemove,
	testTPSplPrfKillEngine,
}

//Test start here
func TestTPSplPrfITMySql(t *testing.T) {
	tpSplPrfConfigDIR = "tutmysql"
	for _, stest := range sTestsTPSplPrf {
		t.Run(tpSplPrfConfigDIR, stest)
	}
}

func TestTPSplPrfITMongo(t *testing.T) {
	tpSplPrfConfigDIR = "tutmongo"
	for _, stest := range sTestsTPSplPrf {
		t.Run(tpSplPrfConfigDIR, stest)
	}
}

func testTPSplPrfInitCfg(t *testing.T) {
	var err error
	tpSplPrfCfgPath = path.Join(tpSplPrfDataDire, "conf", "samples", tpSplPrfConfigDIR)
	tpSplPrfCfg, err = config.NewCGRConfigFromFolder(tpSplPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpSplPrfCfg.DataFolderPath = tpSplPrfDataDire // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpSplPrfCfg)
	switch tpSplPrfConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpSplPrfDelay = 2000
	default:
		tpSplPrfDelay = 1000
	}
}

// Wipe out the cdr database
func testTPSplPrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpSplPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPSplPrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpSplPrfCfgPath, tpSplPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPSplPrfRPCConn(t *testing.T) {
	var err error
	tpSplPrfRPC, err = jsonrpc.Dial("tcp", tpSplPrfCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPSplPrfGetTPSplPrfBeforeSet(t *testing.T) {
	var reply *utils.TPSupplier
	if err := tpSplPrfRPC.Call("ApierV1.GetTPSupplierProfile",
		&AttrGetTPSupplierProfile{TPid: "TP1", ID: "SUPL_1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSplPrfSetTPSplPrf(t *testing.T) {
	tpSplPr = &utils.TPSupplierProfile{
		TPid:      "TP1",
		Tenant:    "cgrates.org",
		ID:        "SUPL_1",
		FilterIDs: []string{"FLTR_ACNT_dan", "FLTR_DST_DE"},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
		Sorting:           "*lowest_cost",
		SortingParameters: []string{},
		Suppliers: []*utils.TPSupplier{
			&utils.TPSupplier{
				ID:                 "supplier1",
				FilterIDs:          []string{"FLTR_1"},
				AccountIDs:         []string{"Acc1", "Acc2"},
				RatingPlanIDs:      []string{"RPL_1"},
				ResourceIDs:        []string{"ResGroup1"},
				StatIDs:            []string{"Stat1"},
				Weight:             10,
				Blocker:            false,
				SupplierParameters: "SortingParam1",
			},
		},
		Weight: 20,
	}
	var result string
	if err := tpSplPrfRPC.Call("ApierV1.SetTPSupplierProfile",
		tpSplPr, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSplPrfGetTPSplPrfAfterSet(t *testing.T) {
	var reply *utils.TPSupplierProfile
	if err := tpSplPrfRPC.Call("ApierV1.GetTPSupplierProfile",
		&AttrGetTPSupplierProfile{TPid: "TP1", ID: "SUPL_1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpSplPr, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpSplPr), utils.ToJSON(reply))
	}
}

func testTPSplPrfGetTPSplPrfIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"SUPL_1"}
	if err := tpSplPrfRPC.Call("ApierV1.GetTPSupplierProfileIDs",
		&AttrGetTPSupplierProfileIDs{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}

}

func testTPSplPrfUpdateTPSplPrf(t *testing.T) {
	tpSplPr.Suppliers = []*utils.TPSupplier{
		&utils.TPSupplier{
			ID:                 "supplier1",
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc1", "Acc2"},
			RatingPlanIDs:      []string{"RPL_1"},
			ResourceIDs:        []string{"ResGroup1"},
			StatIDs:            []string{"Stat1"},
			Weight:             10,
			Blocker:            true,
			SupplierParameters: "SortingParam1",
		},
		&utils.TPSupplier{
			ID:                 "supplier2",
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc3"},
			RatingPlanIDs:      []string{"RPL_1"},
			ResourceIDs:        []string{"ResGroup1"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParam2",
		},
	}
	var result string
	if err := tpSplPrfRPC.Call("ApierV1.SetTPSupplierProfile",
		tpSplPr, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPSplPrfGetTPSplPrfAfterUpdate(t *testing.T) {
	var reply *utils.TPSupplierProfile
	reverseSuppliers := []*utils.TPSupplier{
		&utils.TPSupplier{
			ID:                 "supplier2",
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc3"},
			RatingPlanIDs:      []string{"RPL_1"},
			ResourceIDs:        []string{"ResGroup1"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParam2",
		},
		&utils.TPSupplier{
			ID:                 "supplier1",
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc1", "Acc2"},
			RatingPlanIDs:      []string{"RPL_1"},
			ResourceIDs:        []string{"ResGroup1"},
			StatIDs:            []string{"Stat1"},
			Weight:             10,
			Blocker:            true,
			SupplierParameters: "SortingParam1",
		},
	}
	if err := tpSplPrfRPC.Call("ApierV1.GetTPSupplierProfile",
		&AttrGetTPSupplierProfile{TPid: "TP1", ID: "SUPL_1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpSplPr.Suppliers, reply.Suppliers) &&
		!reflect.DeepEqual(tpSplPr.Suppliers, reverseSuppliers) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(tpSplPr), utils.ToJSON(reply))
	}
}

func testTPSplPrfRemTPSplPrf(t *testing.T) {
	var resp string
	if err := tpSplPrfRPC.Call("ApierV1.RemTPSupplierProfile",
		&AttrRemTPSupplierProfile{TPid: "TP1", Tenant: "cgrates.org", ID: "SUPL_1"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPSplPrfGetTPSplPrfAfterRemove(t *testing.T) {
	var reply *utils.TPSupplierProfile
	if err := tpSplPrfRPC.Call("ApierV1.GetTPSupplierProfile",
		&AttrGetTPSupplierProfile{TPid: "TP1", ID: "SUPL_1"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPSplPrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpSplPrfDelay); err != nil {
		t.Error(err)
	}
}
