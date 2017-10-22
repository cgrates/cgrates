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
	tSv1CfgPath string
	tSv1Cfg     *config.CGRConfig
	tSv1Rpc     *rpc.Client
	tPrfl       *engine.ThresholdProfile
	tSv1ConfDIR string //run tests for specific configuration
	thdsDelay   int
)

var tEvs = []*engine.ThresholdEvent{
	&engine.ThresholdEvent{ // hitting THD_ACNT_BALANCE_1
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType:     utils.AccountUpdate,
			utils.ACCOUNT:       "1002",
			utils.AllowNegative: true,
			utils.Disabled:      false}},
	&engine.ThresholdEvent{ // hitting THD_ACNT_BALANCE_1
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.EventType:  utils.BalanceUpdate,
			utils.ACCOUNT:    "1002",
			utils.BalanceID:  utils.META_DEFAULT,
			utils.Units:      12.3,
			utils.ExpiryTime: "2009-11-10T23:00:00Z"}},
	&engine.ThresholdEvent{ // hitting THD_STATS_1
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "Stats1",
			utils.ACCOUNT:   "1002",
			"ASR":           35.0,
			"ACD":           "2m45s",
			"TCC":           12.7,
			"TCD":           "12m15s",
			"ACC":           0.75,
			"PDD":           "2s",
		}},
	&engine.ThresholdEvent{ // hitting THD_STATS_1 and THD_STATS_2
		Tenant: "cgrates.org",
		ID:     "event4",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "STATS_HOURLY_DE",
			utils.ACCOUNT:   "1002",
			"ASR":           35.0,
			"ACD":           "2m45s",
			"TCD":           "1h",
		}},
	&engine.ThresholdEvent{ // hitting THD_STATS_3
		Tenant: "cgrates.org",
		ID:     "event5",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "STATS_DAILY_DE",
			utils.ACCOUNT:   "1002",
			"ACD":           "2m45s",
			"TCD":           "3h1s",
		}},
	&engine.ThresholdEvent{ // hitting THD_RES_1
		Tenant: "cgrates.org",
		ID:     "event6",
		Event: map[string]interface{}{
			utils.EventType:  utils.ResourceUpdate,
			utils.ACCOUNT:    "1002",
			utils.ResourceID: "RES_GRP_1",
			utils.USAGE:      10.0}},
}

var sTestsThresholdSV1 = []func(t *testing.T){
	testV1TSLoadConfig,
	testV1TSInitDataDb,
	testV1TSStartEngine,
	testV1TSRpcConn,
	testV1TSFromFolder,
	testV1TSGetThresholds,
	testV1TSProcessEvent,
	testV1TSGetThresholdsAfterProcess,
	testV1TSGetThresholdsAfterRestart,
	testV1TSSetThresholdProfile,
	testV1TSUpdateThresholdProfile,
	testV1TSRemoveThresholdProfile,
	testV1TSStopEngine,
}

// Test start here
func TestTSV1ITMySQL(t *testing.T) {
	tSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsThresholdSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func TestTSV1ITMongo(t *testing.T) {
	tSv1ConfDIR = "tutmongo"
	time.Sleep(time.Duration(5 * time.Second)) // give time for engine to start
	for _, stest := range sTestsThresholdSV1 {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1TSLoadConfig(t *testing.T) {
	var err error
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	if tSv1Cfg, err = config.NewCGRConfigFromFolder(tSv1CfgPath); err != nil {
		t.Error(err)
	}
	switch tSv1ConfDIR {
	case "tutmongo": // Mongo needs more time to reset db, need to investigate
		thdsDelay = 4000
	default:
		thdsDelay = 1000
	}
}

func testV1TSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1TSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, thdsDelay); err != nil {
		t.Fatal(err)
	}
}

func testV1TSRpcConn(t *testing.T) {
	var err error
	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1TSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tSv1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
}

func testV1TSGetThresholds(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED", "THD_STATS_3"}
	if err := tSv1Rpc.Call("ThresholdSV1.GetThresholdIDs", "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: expectedIDs[0]}
	if err := tSv1Rpc.Call("ThresholdSV1.GetThreshold",
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testV1TSProcessEvent(t *testing.T) {
	var hits int
	eHits := 0
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[0], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
	eHits = 1
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[1], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
	eHits = 1
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[2], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
	eHits = 2
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[3], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
	eHits = 1
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[4], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
	eHits = 1
	if err := tSv1Rpc.Call("ThresholdSV1.ProcessEvent", tEvs[5], &hits); err != nil {
		t.Error(err)
	} else if hits != eHits {
		t.Errorf("Expecting hits: %d, received: %d", eHits, hits)
	}
}

func testV1TSGetThresholdsAfterProcess(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED"}
	if err := tSv1Rpc.Call("ThresholdSV1.GetThresholdIDs", "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) { // THD_STATS_3 is not reccurent, so it was removed
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call("ThresholdSV1.GetThreshold",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
}

func testV1TSGetThresholdsAfterRestart(t *testing.T) {
	time.Sleep(time.Second)
	if _, err := engine.StopStartEngine(tSv1CfgPath, thdsDelay); err != nil {
		t.Fatal(err)
	}
	var err error
	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call("ThresholdSV1.GetThreshold",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
	time.Sleep(time.Duration(1 * time.Second))
}

func testV1TSSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant: "cgrates.org",
		ID:     "TEST_PROFILE1",
		Filters: []*engine.RequestFilter{
			&engine.RequestFilter{
				Type:      "type",
				FieldName: "Name",
				Values:    []string{"FilterValue1", "FilterValue2"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC).Local(),
		},
		Recurrent: true,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1", "ACT_2"},
		Async:     true,
	}
	var result string
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
}

func testV1TSUpdateThresholdProfile(t *testing.T) {
	var result string
	tPrfl.Filters = []*engine.RequestFilter{
		&engine.RequestFilter{
			Type:      "type",
			FieldName: "Name",
			Values:    []string{"FilterValue1", "FilterValue2"},
		},
		&engine.RequestFilter{
			Type:      "*string",
			FieldName: "Accout",
			Values:    []string{"1001", "1002"},
		},
		&engine.RequestFilter{
			Type:      "*string_prefix",
			FieldName: "Destination",
			Values:    []string{"10", "20"},
		},
	}
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	time.Sleep(time.Duration(10 * time.Millisecond)) // mongo is async
	var reply *engine.ThresholdProfile
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
}

func testV1TSRemoveThresholdProfile(t *testing.T) {
	var resp string
	if err := tSv1Rpc.Call("ApierV1.RemThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1TSStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
