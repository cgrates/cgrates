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
	"path"
	"reflect"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// subtests to be executed
var (
	doubleRemovePath    string
	doubleRemoveDIR     string
	doubleRemove        *config.CGRConfig
	doubleRemoveAccount = "refundAcc"
	doubleRemoveTenant  = "cgrates.org"

	doubleRemoveIT = []func(t *testing.T){
		testdoubleRemoveLoadConfig,
		testdoubleRemoveInitDataDb,
		testdoubleRemoveStartEngine,
		testdoubleRemoveRpcConn,
		testdoubleRemoveFromFolder,
		testdoubleRemoveStatQueueProfile,
		testdoubleRemoveKillEngine,
	}
)

//Test start here
func TestDoubleRemoveIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		doubleRemoveDIR = "tutinternal"
	case utils.MetaMySQL:
		doubleRemoveDIR = "tutmysql"
	case utils.MetaMongo:
		doubleRemoveDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range doubleRemoveIT {
		t.Run(doubleRemoveDIR, stest)
	}
}

func testdoubleRemoveLoadConfig(t *testing.T) {
	var err error
	doubleRemovePath = path.Join(*dataDir, "conf", "samples", doubleRemoveDIR)
	if doubleRemove, err = config.NewCGRConfigFromPath(doubleRemovePath); err != nil {
		t.Error(err)
	}
}

func testdoubleRemoveInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(doubleRemove); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(doubleRemovePath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testdoubleRemoveRpcConn(t *testing.T) {
	var err error
	sesRPC, err = newRPCClient(doubleRemove.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testdoubleRemoveFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := sesRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testdoubleRemoveStatQueueProfile(t *testing.T) {
	// check
	var reply *engine.StatQueueProfile
	if err := sesRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// set
	statConfig := &v1.StatQueueWithCache{
		StatQueueProfile: &engine.StatQueueProfile{
			Tenant: doubleRemoveTenant,
			ID:     "TEST_PROFILE1",
			ActivationInterval: &utils.ActivationInterval{
				ActivationTime: time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
				ExpiryTime:     time.Date(2020, 4, 18, 14, 25, 0, 0, time.UTC),
			},
			QueueLength: 10,
			TTL:         time.Duration(10) * time.Second,
			Metrics: []*engine.MetricWithFilters{
				&engine.MetricWithFilters{
					MetricID: "*sum",
				},
				&engine.MetricWithFilters{
					MetricID: "*acd",
				},
			},
			ThresholdIDs: []string{"Val1", "Val2"},
			Blocker:      true,
			Stored:       true,
			Weight:       20,
			MinItems:     1,
		},
	}
	var result string
	if err := sesRPC.Call(utils.APIerSv1SetStatQueueProfile, statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	time.Sleep(50 * time.Millisecond)
	//check
	if err := sesRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(statConfig.StatQueueProfile, reply) {
		t.Errorf("Expecting: %+v, received: %+v", statConfig.StatQueueProfile, reply)
	}

	//remove
	if err := sesRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := sesRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sesRPC.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &result); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
	// check
	if err := sesRPC.Call(utils.APIerSv1GetStatQueueProfile,
		&utils.TenantID{Tenant: doubleRemoveTenant, ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testdoubleRemoveKillEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
