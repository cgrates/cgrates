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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tpAcctPrfCfgPath   string
	tpAcctPrfCfg       *config.CGRConfig
	tpAcctPrfRPC       *rpc.Client
	tpAcctPrf          *utils.TPAccountProfile
	tpAcctPrfDelay     int
	tpAcctPrfConfigDIR string //run tests for specific configuration
)

var sTestsTPAcctPrf = []func(t *testing.T){
	testTPAcctPrfInitCfg,
	testTPAcctPrfResetStorDb,
	testTPAcctPrfStartEngine,
	testTPAcctPrfRPCConn,
	testTPAcctPrfGetTPAcctPrfBeforeSet,
	testTPAcctPrfSetTPAcctPrf,
	testTPAcctPrfGetTPAcctPrfAfterSet,
	testTPAcctPrfGetTPAcctPrfIDs,
	testTPAcctPrfUpdateTPAcctBal,
	testTPAcctPrfGetTPAcctBalAfterUpdate,
	testTPAcctPrfRemTPAcctPrf,
	testTPAcctPrfGetTPAcctPrfAfterRemove,
	testTPAcctPrfKillEngine,
}

//Test start here
func TestTPAcctPrfIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		tpAcctPrfConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		tpAcctPrfConfigDIR = "tutmysql"
	case utils.MetaMongo:
		tpAcctPrfConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTPAcctPrf {
		t.Run(tpAcctPrfConfigDIR, stest)
	}
}

func testTPAcctPrfInitCfg(t *testing.T) {
	var err error
	tpAcctPrfCfgPath = path.Join(*dataDir, "conf", "samples", tpAcctPrfConfigDIR)
	tpAcctPrfCfg, err = config.NewCGRConfigFromPath(tpAcctPrfCfgPath)
	if err != nil {
		t.Error(err)
	}
	tpAcctPrfDelay = 1000
}

// Wipe out the cdr database
func testTPAcctPrfResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tpAcctPrfCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testTPAcctPrfStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tpAcctPrfCfgPath, tpAcctPrfDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testTPAcctPrfRPCConn(t *testing.T) {
	var err error
	tpAcctPrfRPC, err = jsonrpc.Dial(utils.TCP, tpAcctPrfCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testTPAcctPrfGetTPAcctPrfBeforeSet(t *testing.T) {
	var reply *utils.TPAccountProfile
	if err := tpAcctPrfRPC.Call(utils.APIerSv1GetTPAccountProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "1001"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAcctPrfSetTPAcctPrf(t *testing.T) {
	tpAcctPrf = &utils.TPAccountProfile{
		TPid:    "TP1",
		Tenant:  "cgrates.org",
		ID:      "1001",
		Weights: ";20",
		Balances: map[string]*utils.TPAccountBalance{
			"MonetaryBalance": {
				ID:      "MonetaryBalance",
				Weights: ";10",
				Type:    utils.MetaMonetary,
				Units:   14,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	var result string
	if err := tpAcctPrfRPC.Call(utils.APIerSv1SetTPAccountProfile, tpAcctPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAcctPrfGetTPAcctPrfAfterSet(t *testing.T) {
	var reply *utils.TPAccountProfile
	if err := tpAcctPrfRPC.Call(utils.APIerSv1GetTPAccountProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "1001"}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tpAcctPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(tpAcctPrf), utils.ToJSON(reply))
	}
}

func testTPAcctPrfGetTPAcctPrfIDs(t *testing.T) {
	var result []string
	expectedTPID := []string{"cgrates.org:1001"}
	if err := tpAcctPrfRPC.Call(utils.APIerSv1GetTPAccountProfileIDs,
		&AttrGetTPAccountProfileIDs{TPid: "TP1"}, &result); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedTPID, result) {
		t.Errorf("Expecting: %+v, received: %+v", expectedTPID, result)
	}
}

func testTPAcctPrfUpdateTPAcctBal(t *testing.T) {
	tpAcctPrf.Balances = map[string]*utils.TPAccountBalance{
		"MonetaryBalance2": {
			ID:      "MonetaryBalance2",
			Weights: ";12",
			Type:    utils.MetaMonetary,
			Units:   16,
		},
	}
	var result string
	if err := tpAcctPrfRPC.Call(utils.APIerSv1SetTPAccountProfile, tpAcctPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testTPAcctPrfGetTPAcctBalAfterUpdate(t *testing.T) {
	var reply *utils.TPAccountProfile
	revTPAcctPrf := &utils.TPAccountProfile{
		TPid:    "TP1",
		Tenant:  "cgrates.org",
		ID:      "1001",
		Weights: ";20",
		Balances: map[string]*utils.TPAccountBalance{
			"MonetaryBalance2": {
				ID:      "MonetaryBalance2",
				Weights: ";12",
				Type:    utils.MetaMonetary,
				Units:   16,
			},
		},
		ThresholdIDs: []string{utils.MetaNone},
	}
	if err := tpAcctPrfRPC.Call(utils.APIerSv1GetTPAccountProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "1001"}, &reply); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tpAcctPrf, reply) && !reflect.DeepEqual(revTPAcctPrf, reply) {
		t.Errorf("Expecting : %+v, \n received: %+v", utils.ToJSON(tpAcctPrf), utils.ToJSON(reply))
	}
}

func testTPAcctPrfRemTPAcctPrf(t *testing.T) {
	var resp string
	if err := tpAcctPrfRPC.Call(utils.APIerSv1RemoveTPAccountProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "1001"},
		&resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
}

func testTPAcctPrfGetTPAcctPrfAfterRemove(t *testing.T) {
	var reply *utils.TPAccountProfile
	if err := tpAcctPrfRPC.Call(utils.APIerSv1GetTPAccountProfile,
		&utils.TPTntID{TPid: "TP1", Tenant: "cgrates.org", ID: "1001"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testTPAcctPrfKillEngine(t *testing.T) {
	if err := engine.KillEngine(tpAcctPrfDelay); err != nil {
		t.Error(err)
	}
}
