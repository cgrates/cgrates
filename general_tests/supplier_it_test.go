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

package general_tests

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
)

var sTestsSupplierSV1 = []func(t *testing.T){
	testV1SplSLoadConfig,
	testV1SplSInitDataDb,
	testV1SplSResetStorDb,
	testV1SplSStartEngine,
	testV1SplSRpcConn,
	testV1SplSFromFolder,
	testV1SplSSetSupplierProfilesWithoutRatingPlanIDs,
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
	if _, err := engine.StopStartEngine(splSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1SplSRpcConn(t *testing.T) {
	var err error
	splSv1Rpc, err = jsonrpc.Dial("tcp", splSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1SplSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := splSv1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1SplSSetSupplierProfilesWithoutRatingPlanIDs(t *testing.T) {
	var reply *engine.SupplierProfile
	if err := splSv1Rpc.Call("ApierV1.GetSupplierProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	splPrf = &engine.SupplierProfile{
		Tenant:  "cgrates.org",
		ID:      "TEST_PROFILE2",
		Sorting: utils.MetaLeastCost,
		Suppliers: []*engine.Supplier{
			{
				ID:         "SPL1",
				FilterIDs:  []string{"FLTR_1"},
				AccountIDs: []string{"accc"},
				Weight:     20,
				Blocker:    false,
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
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE2"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(splPrf, reply) {
		t.Errorf("Expecting: %+v, received: %+v", splPrf, reply)
	}
	ev := &engine.ArgsGetSuppliers{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testV1SplSGetLeastCostSuppliers",
			Event: map[string]interface{}{
				utils.Account:     "accc",
				utils.Subject:     "1003",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2017, 12, 1, 14, 25, 0, 0, time.UTC),
				utils.Usage:       "1m20s",
			},
		},
	}
	var suplsReply engine.SortedSuppliers
	if err := splSv1Rpc.Call(utils.SupplierSv1GetSuppliers,
		ev, &suplsReply); err.Error() != utils.NewErrServerError(utils.NewErrMandatoryIeMissing("RatingPlanIDs")).Error() {
		t.Errorf("Expected error MANDATORY_IE_MISSING: [RatingPlanIDs] recieved:%v\n", err)
	}
}

func testV1SplSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
