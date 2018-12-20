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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpFilterCfgPath   string
	tpFilterCfg       *config.CGRConfig
	tpFilterRPC       *rpc.Client
	tpFilterDataDir   = "/usr/share/cgrates"
	tpFilter          *utils.TPFilterProfile
	tpFilterDelay     int
	tpFilterConfigDIR string //run tests for specific configuration
)

var sTestsTPFilters = []func(t *testing.T){
	testTPFilterInitCfg,
	testTPFilterResetStorDb,
	testTPFilterStartEngine,
	testTPFilterRpcConn,
	ttestTPFilterGetTPFilterBeforeSet,
	testTPFilterSetTPFilter,
	testTPFilterGetTPFilterAfterSet,
	testTPFilterGetFilterIds,
	testTPFilterUpdateTPFilter,
	testTPFilterGetTPFilterAfterUpdate,
	testTPFilterRemTPFilter,
	testTPFilterGetTPFilterAfterRemove,
	testTPFilterKillEngine,
}

//Test start here
func TestTPFilterITMySql(t *testing.T) {
	tpFilterConfigDIR = "tutmysql"
	for _, stest := range sTestsTPFilters {
		t.Run(tpFilterConfigDIR, stest)
	}
}

func TestTPFilterITMongo(t *testing.T) {
	tpFilterConfigDIR = "tutmongo"
	for _, stest := range sTestsTPFilters {
		t.Run(tpFilterConfigDIR, stest)
	}
}

func TestTPFilterITPG(t *testing.T) {
	tpFilterConfigDIR = "tutpostgres"
	for _, stest := range sTestsTPFilters {
		t.Run(tpFilterConfigDIR, stest)
	}
}

func testTPFilterInitCfg(t *testing.T) {
	var err error
	tpFilterCfgPath = path.Join(tpFilterDataDir, "conf", "samples", tpFilterConfigDIR)
	tpFilterCfg, err = config.NewCGRConfigFromFolder(tpFilterCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpFilterCfg.DataFolderPath = tpFilterDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(tpFilterCfg)
	switch tpFilterConfigDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		tpFilterDelay = 2000
	default:
		tpFilterDelay = 1000
	}
}

// Wipe out the cdr database
func testTPFilterResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpFilterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPFilterStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpFilterCfgPath, tpFilterDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPFilterRpcConn(t *testing.T) {
	var err error
	tpFilterRPC, err = jsonrpc.Dial("tcp", tpFilterCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func ttestTPFilterGetTPFilterBeforeSet(t *testing.T) {
	var reply *utils.TPFilterProfile
	if err := tpFilterRPC.Call("ApierV1.GetTPFilterProfile",
		&AttrGetTPFilterProfile{TPid: "TP1", ID: "Filter"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPFilterSetTPFilter(t *testing.T) {
	tpFilter = &utils.TPFilterProfile{
		TPid:   "TP1",
		Tenant: "cgrates.org",
		ID:     "Filter",
		Filters: []*utils.TPFilter{
			&utils.TPFilter{
				Type:      "*string",
				FieldName: "Account",
				Values:    []string{"1001", "1002"},
			},
		},
		ActivationInterval: &utils.TPActivationInterval{
			ActivationTime: "2014-07-29T15:00:00Z",
			ExpiryTime:     "",
		},
	}

	var result string
	if err := tpFilterRPC.Call("ApierV1.SetTPFilterProfile", tpFilter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPFilterGetTPFilterAfterSet(t *testing.T) {
	var reply *utils.TPFilterProfile
	if err := tpFilterRPC.Call("ApierV1.GetTPFilterProfile",
		&AttrGetTPFilterProfile{TPid: "TP1", ID: "Filter"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpFilter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpFilter, reply)
	}
}

func testTPFilterGetFilterIds(t *testing.T) {
	var result []string
	expectedTPID := []string{"Filter"}
	if err := tpFilterRPC.Call("ApierV1.GetTPFilterProfileIds",
		&AttrGetTPFilterProfileIds{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPFilterUpdateTPFilter(t *testing.T) {
	tpFilter.Filters = []*utils.TPFilter{
		&utils.TPFilter{
			Type:      "*string",
			FieldName: "Account",
			Values:    []string{"1001", "1002"},
		},
		&utils.TPFilter{
			Type:      engine.MetaPrefix,
			FieldName: "Destination",
			Values:    []string{"10", "20"},
		},
	}
	var result string
	if err := tpFilterRPC.Call("ApierV1.SetTPFilterProfile", tpFilter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPFilterGetTPFilterAfterUpdate(t *testing.T) {
	var reply *utils.TPFilterProfile
	if err := tpFilterRPC.Call("ApierV1.GetTPFilterProfile",
		&AttrGetTPFilterProfile{TPid: "TP1", ID: "Filter"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tpFilter, reply) {
		t.Errorf("Expecting : %+v, received: %+v", tpFilter, reply)
	}
}

func testTPFilterRemTPFilter(t *testing.T) {
	var resp string
	if err := tpFilterRPC.Call("ApierV1.RemTPFilterProfile",
		&AttrRemTPFilterProfile{TPid: "TP1", Tenant: "cgrates.org", ID: "Filter"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPFilterGetTPFilterAfterRemove(t *testing.T) {
	var reply *utils.TPFilterProfile
	if err := tpFilterRPC.Call("ApierV1.GetTPFilterProfile",
		&AttrGetTPFilterProfile{TPid: "TP1", ID: "Filter"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPFilterKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpFilterDelay); err != nil {
		t.Error(err)
	}
}
