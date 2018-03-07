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
	splSv1CfgPath string
	splSv1Cfg     *config.CGRConfig
	splSv1Rpc     *rpc.Client
	splPrf        *engine.SupplierProfile
	splSv1ConfDIR string //run tests for specific configuration
	splsDelay     int
)

var sTestsSupplierSV1 = []func(t *testing.T){
	testV1SplSLoadConfig,
	testV1SplSInitDataDb,
	testV1SplSResetStorDb,
	testV1SplSStartEngine,
	testV1SplSRpcConn,
	testV1SplSFromFolder,
	testV1SplSGetWeightSuppliers,
	testV1SplSGetLeastCostSuppliers,
	testV1SplSGetSupplierWithoutFilter,
	testV1SplSSetSupplierProfiles,
	testV1SplSUpdateSupplierProfiles,
	testV1SplSRemSupplierProfiles,
	testV1SplSStopEngine,
}

// Test start here
func TestSuplSV1ITMySQL(t *testing.T) {
	splSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func TestSuplSV1ITMongo(t *testing.T) {
	splSv1ConfDIR = "tutmongo"
	time.Sleep(time.Duration(2 * time.Second)) // give time for engine to start
	for _, stest := range sTestsSupplierSV1 {
		t.Run(splSv1ConfDIR, stest)
	}
}

func testV1SplSLoadConfig(t *testing.T) {
	var err error
	splSv1CfgPath = path.Join(*dataDir, "conf", "samples", splSv1ConfDIR)
	if splSv1Cfg, err = config.NewCGRConfigFromFolder(splSv1CfgPath); err != nil {
		t.Error(err)
	}
	switch splSv1ConfDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		splsDelay = 4000
	default:
		splsDelay = 1000
	}
}

func testV1SplSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1SplSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(splSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(splSv1CfgPath, splsDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	var err error
	splSv1Rpc, err = jsonrpc.Dial("tcp", splSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := splSv1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1SplSGetWeightSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetWeightSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1007",
				utils.Destination: "+491511231234",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			&engine.SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
			},
			&engine.SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetLeastCostSuppliers(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_LEASTCOST_1",
		Sorting:   utils.MetaLeastCost,
		SortedSuppliers: []*engine.SortedSupplier{
			&engine.SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:         0.02,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       15.0,
				},
			},
			&engine.SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:         0.02,
					utils.RatingPlanID: "RP_SPECIAL_1002",
					utils.Weight:       10.0,
				},
			},
			&engine.SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:         0.46666,
					utils.RatingPlanID: "RP_RETAIL1",
					utils.Weight:       20.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSGetSupplierWithoutFilter(t *testing.T) {
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetSupplierWithoutFilter",
			Event: map[string]interface{}{
				utils.Account:     "1008",
				utils.Destination: "+49",
			},
		},
	}
	eSpls := engine.SortedSuppliers{
		ProfileID: "SPL_WEIGHT_2",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*engine.SortedSupplier{
			&engine.SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eSpls, suplsReply) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eSpls), utils.ToJSON(suplsReply))
	}
}

func testV1SplSSetSupplierProfiles(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &engine.SupplierProfile{
		Tenant:            "cgrates.org",
		ID:                "TEST_PROFILE1",
		FilterIDs:         []string{"FLTR_1"},
		Sorting:           "Sort1",
		SortingParameters: []string{"Param1", "Param2"},
		Suppliers: []*engine.Supplier{
			&engine.Supplier{
				ID:                 "SPL1",
				RatingPlanIDs:      []string{"RP1"},
				FilterIDs:          []string{"FLTR_1"},
				AccountIDs:         []string{"Acc"},
				ResourceIDs:        []string{"Res1", "ResGroup2"},
				StatIDs:            []string{"Stat1"},
				Weight:             20,
				Blocker:            false,
				SupplierParameters: "SortingParameter1",
			},
		},
		Weight: 10,
	}
	var result string
	if err := splSv1Rpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf, reply)
	}
}

func testV1SplSUpdateSupplierProfiles(t *testing.T) {
	splPrf.Suppliers = []*engine.Supplier{
		&engine.Supplier{
			ID:                 "SPL1",
			RatingPlanIDs:      []string{"RP1"},
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res1", "ResGroup2"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParameter1",
		},
		&engine.Supplier{
			ID:                 "SPL2",
			RatingPlanIDs:      []string{"RP2"},
			FilterIDs:          []string{"FLTR_2"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res2", "ResGroup2"},
			StatIDs:            []string{"Stat2"},
			Weight:             20,
			Blocker:            true,
			SupplierParameters: "SortingParameter2",
		},
	}
	reverseSuppliers := []*engine.Supplier{
		&engine.Supplier{
			ID:                 "SPL2",
			RatingPlanIDs:      []string{"RP2"},
			FilterIDs:          []string{"FLTR_2"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res2", "ResGroup2"},
			StatIDs:            []string{"Stat2"},
			Weight:             20,
			Blocker:            true,
			SupplierParameters: "SortingParameter2",
		},
		&engine.Supplier{
			ID:                 "SPL1",
			RatingPlanIDs:      []string{"RP1"},
			FilterIDs:          []string{"FLTR_1"},
			AccountIDs:         []string{"Acc"},
			ResourceIDs:        []string{"Res1", "ResGroup2"},
			StatIDs:            []string{"Stat1"},
			Weight:             20,
			Blocker:            false,
			SupplierParameters: "SortingParameter1",
		},
	}
	var result string
	if err := splSv1Rpc.Call("ApierV1.SetSupplierProfile", splPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf.Suppliers, reply.Suppliers) && !reflect.DeepEqual(reverseSuppliers, reply.Suppliers) {
		t.Errorf("Expecting: %+v, received: %+v", utils.ToJSON(splPrf), utils.ToJSON(reply))
	}
}

func testV1SplSRemSupplierProfiles(t *testing.T) {
	var resp string
	if err := splSv1Rpc.Call("ApierV1.RemoveSupplierProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
