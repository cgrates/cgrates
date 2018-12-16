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
)

var tEvs = []*utils.CGREvent{
	{ // hitting THD_ACNT_BALANCE_1
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType:     utils.AccountUpdate,
			utils.Account:       "1002",
			utils.AllowNegative: true,
			utils.Disabled:      false,
			utils.Units:         12.3}},
	{ // hitting THD_ACNT_BALANCE_1
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.EventType:  utils.BalanceUpdate,
			utils.Account:    "1002",
			utils.BalanceID:  utils.META_DEFAULT,
			utils.Units:      12.3,
			utils.ExpiryTime: time.Date(2009, 11, 10, 23, 00, 0, 0, time.UTC),
		}},
	{ // hitting THD_STATS_1
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "Stats1",
			utils.Account:   "1002",
			"ASR":           35.0,
			"ACD":           "2m45s",
			"TCC":           12.7,
			"TCD":           "12m15s",
			"ACC":           0.75,
			"PDD":           "2s",
		}},
	{ // hitting THD_STATS_1 and THD_STATS_2
		Tenant: "cgrates.org",
		ID:     "event4",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "STATS_HOURLY_DE",
			utils.Account:   "1002",
			"ASR":           35.0,
			"ACD":           "2m45s",
			"TCD":           "1h",
		}},
	{ // hitting THD_STATS_3
		Tenant: "cgrates.org",
		ID:     "event5",
		Event: map[string]interface{}{
			utils.EventType: utils.StatUpdate,
			utils.StatID:    "STATS_DAILY_DE",
			utils.Account:   "1002",
			"ACD":           "2m45s",
			"TCD":           "3h1s",
		}},
	{ // hitting THD_RES_1
		Tenant: "cgrates.org",
		ID:     "event6",
		Event: map[string]interface{}{
			utils.EventType:  utils.ResourceUpdate,
			utils.Account:    "1002",
			utils.ResourceID: "RES_GRP_1",
			utils.Usage:      10.0}},
	{ // hitting THD_RES_1
		Tenant: "cgrates.org",
		ID:     "event6",
		Event: map[string]interface{}{
			utils.EventType:  utils.ResourceUpdate,
			utils.Account:    "1002",
			utils.ResourceID: "RES_GRP_1",
			utils.Usage:      10.0}},
	{ // hitting THD_RES_1
		Tenant: "cgrates.org",
		ID:     "event6",
		Event: map[string]interface{}{
			utils.EventType:  utils.ResourceUpdate,
			utils.Account:    "1002",
			utils.ResourceID: "RES_GRP_1",
			utils.Usage:      10.0}},
	{ // hitting THD_CDRS_1
		Tenant: "cgrates.org",
		ID:     "cdrev1",
		Event: map[string]interface{}{
			utils.EventType:   utils.CDR,
			"field_extr1":     "val_extr1",
			"fieldextr2":      "valextr2",
			utils.CGRID:       utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC).String()),
			utils.RunID:       utils.MetaRaw,
			utils.OrderID:     123,
			utils.OriginHost:  "192.168.1.1",
			utils.Source:      utils.UNIT_TEST,
			utils.OriginID:    "dsafdsaf",
			utils.ToR:         utils.VOICE,
			utils.RequestType: utils.META_RATED,
			utils.Direction:   "*out",
			utils.Tenant:      "cgrates.org",
			utils.Category:    "call",
			utils.Account:     "1007",
			utils.Subject:     "1007",
			utils.Destination: "+4986517174963",
			utils.SetupTime:   time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
			utils.PDD:         time.Duration(0) * time.Second,
			utils.AnswerTime:  time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
			utils.Usage:       time.Duration(10) * time.Second,
			utils.SUPPLIER:    "SUPPL1",
			utils.COST:        -1.0}},
}

var sTestsThresholdSV1 = []func(t *testing.T){
	testV1TSLoadConfig,
	testV1TSInitDataDb,
	testV1TSResetStorDb,
	testV1TSStartEngine,
	testV1TSRpcConn,
	testV1TSFromFolder,
	testV1TSGetThresholds,
	testV1TSProcessEvent,
	testV1TSGetThresholdsAfterProcess,
	testV1TSGetThresholdsAfterRestart,
	testv1TSGetThresholdProfileIDs,
	testV1TSSetThresholdProfile,
	testV1TSUpdateThresholdProfile,
	testV1TSRemoveThresholdProfile,
	testV1TSMaxHits,
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
}

func testV1TSInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1TSResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1TSStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1TSRpcConn(t *testing.T) {
	var err error
	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1TSFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := tSv1Rpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testV1TSGetThresholds(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED", "THD_STATS_3", "THD_CDRS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdIDs, "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) {
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: expectedIDs[0]}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: expectedIDs[0]}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd, td) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
}

func testV1TSProcessEvent(t *testing.T) {
	var ids []string
	eIDs := []string{}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[0], &ids); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	eIDs = []string{"THD_ACNT_BALANCE_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[1], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[2], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_2", "THD_STATS_1"}
	eIDs2 := []string{"THD_STATS_1", "THD_STATS_2"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[3], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) && !reflect.DeepEqual(ids, eIDs2) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_STATS_3"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[4], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_RES_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[5], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[6], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[7], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	eIDs = []string{"THD_CDRS_1"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, tEvs[8], &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
}

func testV1TSGetThresholdsAfterProcess(t *testing.T) {
	var tIDs []string
	expectedIDs := []string{"THD_RES_1", "THD_STATS_2", "THD_STATS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED"}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThresholdIDs, "cgrates.org", &tIDs); err != nil {
		t.Error(err)
	} else if len(expectedIDs) != len(tIDs) { // THD_STATS_3 is not reccurent, so it was removed
		t.Errorf("expecting: %+v, received reply: %s", expectedIDs, tIDs)
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
}

func testV1TSGetThresholdsAfterRestart(t *testing.T) {
	time.Sleep(time.Second)
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	var err error
	tSv1Rpc, err = jsonrpc.Dial("tcp", tSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
	var td engine.Threshold
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &td); err != nil {
		t.Error(err)
	} else if td.Snooze.IsZero() { // make sure Snooze time was reset during execution
		t.Errorf("received: %+v", td)
	}
}

func testv1TSGetThresholdProfileIDs(t *testing.T) {
	expected := []string{"THD_STATS_1", "THD_STATS_2", "THD_STATS_3", "THD_RES_1", "THD_CDRS_1", "THD_ACNT_BALANCE_1", "THD_ACNT_EXPIRED"}
	var result []string
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfileIDs", "cgrates.org", &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testV1TSSetThresholdProfile(t *testing.T) {
	var reply *engine.ThresholdProfile
	var result string
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_Test",
		FilterIDs: []string{"*string:Account:1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
			ExpiryTime:     time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinSleep:  time.Duration(5 * time.Minute),
		Blocker:   false,
		Weight:    20.0,
		ActionIDs: []string{"ACT_1"},
		Async:     true,
	}
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
}

func testV1TSUpdateThresholdProfile(t *testing.T) {
	var result string
	tPrfl.FilterIDs = []string{"*string:Account:1001", "*prefix:DST:10"}
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.ThresholdProfile
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v", tPrfl, reply)
	}
}

func testV1TSRemoveThresholdProfile(t *testing.T) {
	var resp string
	if err := tSv1Rpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_Test"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Recived %s and the error:%+v", utils.ToJSON(sqp), err)
	}
}

func testV1TSMaxHits(t *testing.T) {
	var reply string
	// check if exist
	if err := tSv1Rpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:  "cgrates.org",
		ID:      "TH3",
		MaxHits: 3,
	}
	//set
	if err := tSv1Rpc.Call("ApierV1.SetThresholdProfile", tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}

	var ids []string
	eIDs := []string{"TH3"}
	thEvent := &utils.CGREvent{ // hitting TH3
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.Account: "1002",
		},
	}
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after first process ( hits : 1)
	var td engine.Threshold
	eTd := engine.Threshold{Tenant: "cgrates.org", ID: "TH3", Hits: 1}
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after second process ( hits : 2)
	eTd.Hits = 2
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}, &td); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(eTd.Hits, td.Hits) {
		t.Errorf("expecting: %+v, received: %+v", eTd, td)
	}
	//process event
	if err := tSv1Rpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, eIDs) {
		t.Errorf("Expecting ids: %s, received: %s", eIDs, ids)
	}
	//check threshold after third process (reached the maximum hits and should be removed)
	if err := tSv1Rpc.Call(utils.ThresholdSv1GetThreshold,
		&utils.TenantID{Tenant: "cgrates.org", ID: "TH3"}, &td); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

}

func testV1TSStopEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
