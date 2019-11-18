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
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	vrsCfgPath     string
	vrsCfg         *config.CGRConfig
	vrsRPC         *rpc.Client
	vrsDataDir     = "/usr/share/cgrates"
	vrsDelay       int
	vrsConfigDIR   string //run tests for specific configuration
	vrsStorageType string
)

var sTestsVrs = []func(t *testing.T){
	testVrsInitCfg,
	testVrsResetStorDb,
	testVrsStartEngine,
	testVrsRpcConn,
	testVrsDataDB,
	testVrsStorDB,
	testVrsSetDataDBVrs,
	testVrsSetStorDBVrs,
	testVrsKillEngine,
}

//Test start here
func TestVrsITMySql(t *testing.T) {
	vrsConfigDIR = "tutmysql"
	vrsStorageType = utils.REDIS
	for _, stest := range sTestsVrs {
		t.Run(vrsConfigDIR, stest)
	}
}

func TestVrsITMongo(t *testing.T) {
	vrsConfigDIR = "tutmongo"
	vrsStorageType = utils.MONGO
	for _, stest := range sTestsVrs {
		t.Run(vrsConfigDIR, stest)
	}
}

// func TestVrsITInternal(t *testing.T) {
// 	vrsConfigDIR = "tutinternal"
// 	vrsStorageType = utils.INTERNAL
// 	for _, stest := range sTestsVrs {
// 		t.Run(vrsConfigDIR, stest)
// 	}
// }

func testVrsInitCfg(t *testing.T) {
	var err error
	vrsCfgPath = path.Join(vrsDataDir, "conf", "samples", vrsConfigDIR)
	vrsCfg, err = config.NewCGRConfigFromPath(vrsCfgPath)
	if err != nil {
		t.Error(err)
	}
	vrsCfg.DataFolderPath = vrsDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	vrsDelay = 1000
}

// Wipe out the cdr database
func testVrsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(vrsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testVrsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(vrsCfgPath, vrsDelay); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testVrsRpcConn(t *testing.T) {
	var err error
	vrsRPC, err = jsonrpc.Dial("tcp", vrsCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testVrsDataDB(t *testing.T) {
	var result engine.Versions
	expectedVrs := engine.Versions{"ActionTriggers": 2,
		"Actions": 2, "RQF": 1, "ReverseDestinations": 1, "Attributes": 4, "RatingPlan": 1,
		"RatingProfile": 1, "User": 1, "Accounts": 3, "ActionPlans": 3, "Chargers": 1,
		"Destinations": 1, "LoadIDs": 1, "SharedGroups": 2, "Stats": 2, "Resource": 1,
		"Subscribers": 1, "Suppliers": 1, "Thresholds": 3, "Timing": 1}
	if err := vrsRPC.Call(utils.ApierV1GetDataDBVersions, "", &result); err != nil {
		t.Error(err)
	} else if expectedVrs.Compare(result, vrsStorageType, true) != "" {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedVrs)
	}
}

func testVrsStorDB(t *testing.T) {
	var result engine.Versions
	expectedVrs := engine.Versions{"TpDestinations": 1, "TpResource": 1, "TpThresholds": 1,
		"TpActions": 1, "TpDestinationRates": 1, "TpFilters": 1, "TpRates": 1, "CDRs": 2, "TpActionTriggers": 1, "TpRatingPlans": 1,
		"TpSharedGroups": 1, "TpSuppliers": 1, "SessionSCosts": 3, "TpRatingProfiles": 1, "TpStats": 1, "TpTiming": 1,
		"CostDetails": 2, "TpAccountActions": 1, "TpActionPlans": 1, "TpChargers": 1, "TpRatingProfile": 1,
		"TpRatingPlan": 1, "TpResources": 1}
	if err := vrsRPC.Call(utils.ApierV1GetStorDBVersions, "", &result); err != nil {
		t.Error(err)
	} else if expectedVrs.Compare(result, vrsStorageType, true) != "" {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedVrs)
	}
}

func testVrsSetDataDBVrs(t *testing.T) {
	var reply string
	args := SetVersionsArg{
		Versions: engine.Versions{
			"Attributes": 3,
		},
	}
	if err := vrsRPC.Call("ApierV1.SetDataDBVersions", args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	var result engine.Versions
	expectedVrs := engine.Versions{"ActionTriggers": 2,
		"Actions": 2, "RQF": 1, "ReverseDestinations": 1, "Attributes": 3, "RatingPlan": 1,
		"RatingProfile": 1, "User": 1, "Accounts": 3, "ActionPlans": 3, "Chargers": 1,
		"Destinations": 1, "LoadIDs": 1, "SharedGroups": 2, "Stats": 2, "Resource": 1,
		"Subscribers": 1, "Suppliers": 1, "Thresholds": 3, "Timing": 1}
	if err := vrsRPC.Call(utils.ApierV1GetDataDBVersions, "", &result); err != nil {
		t.Error(err)
	} else if expectedVrs.Compare(result, vrsStorageType, true) != "" {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedVrs)
	}

	args = SetVersionsArg{
		Versions: nil,
	}
	if err := vrsRPC.Call("ApierV1.SetDataDBVersions", args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}
}

func testVrsSetStorDBVrs(t *testing.T) {
	var reply string
	args := SetVersionsArg{
		Versions: engine.Versions{
			"TpResources": 2,
		},
	}
	if err := vrsRPC.Call("ApierV1.SetStorDBVersions", args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}

	var result engine.Versions
	expectedVrs := engine.Versions{"TpDestinations": 1, "TpResource": 1, "TpThresholds": 1,
		"TpActions": 1, "TpDestinationRates": 1, "TpFilters": 1, "TpRates": 1, "CDRs": 2, "TpActionTriggers": 1, "TpRatingPlans": 1,
		"TpSharedGroups": 1, "TpSuppliers": 1, "SessionSCosts": 3, "TpRatingProfiles": 1, "TpStats": 1, "TpTiming": 1,
		"CostDetails": 2, "TpAccountActions": 1, "TpActionPlans": 1, "TpChargers": 1, "TpRatingProfile": 1,
		"TpRatingPlan": 1, "TpResources": 2}
	if err := vrsRPC.Call(utils.ApierV1GetStorDBVersions, "", &result); err != nil {
		t.Error(err)
	} else if expectedVrs.Compare(result, vrsStorageType, true) != "" {
		t.Errorf("Expecting: %+v, received: %+v", result, expectedVrs)
	}

	args = SetVersionsArg{
		Versions: nil,
	}
	if err := vrsRPC.Call("ApierV1.SetStorDBVersions", args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Expecting: %+v, received: %+v", utils.OK, reply)
	}
}

func testVrsKillEngine(t *testing.T) {
	if err := engine.KillEngine(vrsDelay); err != nil {
		t.Error(err)
	}
}
