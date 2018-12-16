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
MERCHANTABILITY or FIdxCaTNESS FOR A PARTICULAR PURPOSE.  See the
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

var tFIdxCaRpc *rpc.Client

var sTestsFilterIndexesSV1Ca = []func(t *testing.T){
	testV1FIdxCaLoadConfig,
	testV1FIdxCaInitDataDb,
	testV1FIdxCaResetStorDb,
	testV1FIdxCaStartEngine,
	testV1FIdxCaRpcConn,

	testV1FIdxCaProcessEventWithNotFound,
	testV1FIdxCaSetThresholdProfile,
	testV1FIdxCaFromFolder,
	testV1FIdxCaGetThresholdFromTP,
	testV1FIdxCaUpdateThresholdProfile,
	testV1FIdxCaUpdateThresholdProfileFromTP,
	testV1FIdxCaRemoveThresholdProfile,

	testV1FIdxCaInitDataDb,
	testV1FIdxCaGetStatQueuesWithNotFound,
	testV1FIdxCaSetStatQueueProfile,
	testV1FIdxCaFromFolder,
	testV1FIdxCaGetStatQueuesFromTP,
	testV1FIdxCaUpdateStatQueueProfile,
	testV1FIdxCaUpdateStatQueueProfileFromTP,
	testV1FIdxCaRemoveStatQueueProfile,

	testV1FIdxCaInitDataDb,
	testV1FIdxCaProcessAttributeProfileEventWithNotFound,
	testV1FIdxCaSetAttributeProfile,
	testV1FIdxCaFromFolder,
	testV1FIdxCaGetAttributeProfileFromTP,
	testV1FIdxCaUpdateAttributeProfile,
	testV1FIdxCaUpdateAttributeProfileFromTP,
	testV1FIdxCaRemoveAttributeProfile,

	testV1FIdxCaInitDataDb,
	testV1FIdxCaGetResourceProfileWithNotFound,
	testV1FIdxCaSetResourceProfile,
	testV1FIdxCaFromFolder,
	testV1FIdxCaGetResourceProfileFromTP,
	testV1FIdxCaUpdateResourceProfile,
	testV1FIdxCaUpdateResourceProfileFromTP,
	testV1FIdxCaRemoveResourceProfile,
	testV1FIdxCaStopEngine,
}

// Test start here
func TestFIdxCaV1ITMySQL(t *testing.T) {
	tSv1ConfDIR = "tutmysql"
	for _, stest := range sTestsFilterIndexesSV1Ca {
		t.Run(tSv1ConfDIR, stest)
	}
}

func TestFIdxCaV1ITMongo(t *testing.T) {
	tSv1ConfDIR = "tutmongo"
	for _, stest := range sTestsFilterIndexesSV1Ca {
		t.Run(tSv1ConfDIR, stest)
	}
}

func testV1FIdxCaLoadConfig(t *testing.T) {
	var err error
	tSv1CfgPath = path.Join(*dataDir, "conf", "samples", tSv1ConfDIR)
	if tSv1Cfg, err = config.NewCGRConfigFromFolder(tSv1CfgPath); err != nil {
		t.Error(err)
	}
}

func testV1FIdxCaInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testV1FIdxCaResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tSv1Cfg); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxCaStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tSv1CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testV1FIdxCaRpcConn(t *testing.T) {
	var err error
	tFIdxCaRpc, err = jsonrpc.Dial("tcp", tSv1Cfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testV1FIdxCaFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := tFIdxCaRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

//ThresholdProfile
func testV1FIdxCaProcessEventWithNotFound(t *testing.T) {
	tEv := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType: utils.BalanceUpdate,
				utils.Account:   "1001"}}}
	var thIDs []string
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxCaSetThresholdProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1001"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.BalanceUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"TestFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MinHits:  1,
		MaxHits:  -1,
		MinSleep: time.Duration(5 * time.Minute),
		Blocker:  false,
		Weight:   20.0,
		Async:    true,
	}

	if err := tFIdxCaRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//matches TEST_PROFILE1
	tEv := &engine.ArgsProcessEvent{
		ThresholdIDs: []string{"TEST_PROFILE1"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType: utils.BalanceUpdate,
				utils.Account:   "1001"}}}
	var thIDs []string
	eIDs := []string{"TEST_PROFILE1"}
	//Testing ProcessEvent on set thresholdprofile using apier

	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting hits: %s, received: %s", eIDs, thIDs)
	}
}

func testV1FIdxCaGetThresholdFromTP(t *testing.T) {
	//matches THD_ACNT_BALANCE_1
	tEv := &engine.ArgsProcessEvent{
		ThresholdIDs: []string{"THD_ACNT_BALANCE_1"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType: utils.BalanceUpdate,
				utils.Account:   "1001",
				utils.BalanceID: utils.META_DEFAULT,
				utils.Units:     12.3}}}
	var thIDs []string
	eIDs := []string{"THD_ACNT_BALANCE_1"}
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent,
		tEv, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting hits: %s, received: %s", eIDs, thIDs)
	}
}

func testV1FIdxCaUpdateThresholdProfile(t *testing.T) {
	var result string
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1002"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.AccountUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tPrfl = &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"TestFilter2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		MaxHits:  -1,
		MinSleep: time.Duration(5 * time.Minute),
		Blocker:  false,
		Weight:   20.0,
		Async:    true,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//make sure doesn't match the thresholdprofile after update
	tEv := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType: utils.AccountUpdate,
				utils.Account:   "1001"}}}
	var thIDs []string
	eIDs := []string{}
	//Testing ProcessEvent on set thresholdprofile  after update making sure there are no hits
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//matches thresholdprofile after update
	tEv2 := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.EventType: utils.AccountUpdate,
				utils.Account:   "1002"}}}
	eIDs = []string{"TEST_PROFILE1"}
	//Testing ProcessEvent on set thresholdprofile after update
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv2, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting : %s, received: %s", eIDs, thIDs)
	}
}

func testV1FIdxCaUpdateThresholdProfileFromTP(t *testing.T) {
	var result string
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter3",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1003"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.BalanceUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	var reply *engine.ThresholdProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &reply); err != nil {
		t.Error(err)
	}
	if reply == nil {
		t.Errorf("Expecting reply to not be nil")
		// reply shoud not be nil so exit function
		// to avoid nil segmentation fault;
		// if this happens try to run this test manualy
		return
	}
	reply.FilterIDs = []string{"TestFilter3"}

	if err := tFIdxCaRpc.Call("ApierV1.SetThresholdProfile", reply, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tEv := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.Account:   "1002",
				utils.EventType: utils.BalanceUpdate}}}
	var thIDs []string
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	tEv2 := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]interface{}{
				utils.Account:   "1003",
				utils.EventType: utils.BalanceUpdate}}}
	eIDs := []string{"THD_ACNT_BALANCE_1"}
	//Testing ProcessEvent on set thresholdprofile using apier
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv2, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting : %s, received: %s", eIDs, thIDs)
	}
}

func testV1FIdxCaRemoveThresholdProfile(t *testing.T) {
	var resp string
	tEv := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event8",
			Event: map[string]interface{}{
				utils.Account:   "1002",
				utils.EventType: utils.AccountUpdate}}}
	var thIDs []string
	eIDs := []string{"TEST_PROFILE1"}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting : %s, received: %s", eIDs, thIDs)
	}

	tEv2 := &engine.ArgsProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event9",
			Event: map[string]interface{}{
				utils.Account:   "1003",
				utils.EventType: utils.BalanceUpdate}}}
	eIDs = []string{"THD_ACNT_BALANCE_1"}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv2, &thIDs); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(thIDs, eIDs) {
		t.Errorf("Expecting : %s, received: %s", eIDs, thIDs)
	}
	//Remove threshold profile that was set form api
	if err := tFIdxCaRpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//Remove threshold profile that was set form tariffplan
	if err := tFIdxCaRpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_BALANCE_1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv, &thIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.ThresholdSv1ProcessEvent, tEv2, &thIDs); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//StatQueue
func testV1FIdxCaGetStatQueuesWithNotFound(t *testing.T) {
	var reply *[]string
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1001",
		},
	}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxCaSetStatQueueProfile(t *testing.T) {
	tenant := "cgrates.org"
	filter = &engine.Filter{
		Tenant: tenant,
		ID:     "FLTR_1",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1001"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.AccountUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string

	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			{
				MetricID:   "*sum",
				Parameters: "Val",
			},
		},
		ThresholdIDs: []string{"Val1", "Val2"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     1,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1001",
		},
	}
	var reply []string
	expected := []string{"TEST_PROFILE1"}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent,
		tEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1FIdxCaGetStatQueuesFromTP(t *testing.T) {
	var reply []string
	expected := []string{"Stats1"}
	ev2 := utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event2",
		Event: map[string]interface{}{
			utils.Account:    "1002",
			utils.AnswerTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:      time.Duration(45 * time.Second)}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, &ev2, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	ev3 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event3",
		Event: map[string]interface{}{
			utils.Account:   "1002",
			utils.SetupTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
			utils.Usage:     0}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, &ev3, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}

	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1001",
			"Val":           7,
		}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, &tEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	tEv2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1001",
			"Val":           8,
		}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, &tEv2, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1FIdxCaUpdateStatQueueProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_2",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1003"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.BalanceUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	statConfig = &engine.StatQueueProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		FilterIDs: []string{"FLTR_2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		QueueLength: 10,
		TTL:         time.Duration(10) * time.Second,
		Metrics: []*utils.MetricWithParams{
			{
				MetricID:   "*sum",
				Parameters: "",
			},
		},
		ThresholdIDs: []string{"Val1", "Val2"},
		Blocker:      true,
		Stored:       true,
		Weight:       20,
		MinItems:     1,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetStatQueueProfile", statConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply []string
	expected := []string{"TEST_PROFILE1"}
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.BalanceUpdate,
			utils.Account:   "1003",
		}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
}

func testV1FIdxCaUpdateStatQueueProfileFromTP(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_3",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1003"},
			},
			{
				FieldName: utils.EventType,
				Type:      "*string",
				Values:    []string{utils.AccountUpdate},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply engine.StatQueueProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &reply); err != nil {
		t.Error(err)
	}
	reply.FilterIDs = []string{"FLTR_3"}
	reply.ActivationInterval = &utils.ActivationInterval{ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}
	if err := tFIdxCaRpc.Call("ApierV1.SetStatQueueProfile",
		reply, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1003",
		}}
	var ids []string
	expected := []string{"Stats1"}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent,
		tEv, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, ids)
	}
}

func testV1FIdxCaRemoveStatQueueProfile(t *testing.T) {
	var reply []string
	expected := []string{"TEST_PROFILE1"}
	tEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.BalanceUpdate,
			utils.Account:   "1003",
		}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	expected = []string{"Stats1"}
	tEv2 := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "event1",
		Event: map[string]interface{}{
			utils.EventType: utils.AccountUpdate,
			utils.Account:   "1003",
		}}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv2, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, expected) {
		t.Errorf("Expecting: %+v, received: %+v", expected, reply)
	}
	var result string
	//Remove threshold profile that was set form api
	if err := tFIdxCaRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var sqp *engine.StatQueueProfile
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//Remove threshold profile that was set form tariffplan
	if err := tFIdxCaRpc.Call("ApierV1.RemStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetStatQueueProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stats1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.StatSv1ProcessEvent, tEv2, &reply); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

//AttributeProfile
func testV1FIdxCaProcessAttributeProfileEventWithNotFound(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "3009",
			utils.Destination: "+492511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxCaSetAttributeProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1009"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"+491511231234"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"TestFilter"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.Account,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     false,
			},
			{
				FieldName:  utils.Subject,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//matches TEST_PROFILE1
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1009",
			utils.Destination: "+491511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent,
		ev, &rplyEv); err != nil {
		t.Error(err)
	}
}

func testV1FIdxCaGetAttributeProfileFromTP(t *testing.T) {
	//matches ATTR_1
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "1007",
			utils.Destination: "+491511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err != nil {
		t.Error(err)
	}
}

func testV1FIdxCaUpdateAttributeProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter2",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"2009"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"+492511231234"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "TEST_PROFILE1",
		Contexts:  []string{utils.MetaSessionS},
		FilterIDs: []string{"TestFilter2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 35, 0, 0, time.UTC),
		},
		Attributes: []*engine.Attribute{
			{
				FieldName:  utils.Account,
				Initial:    utils.META_ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     false,
			},
			{
				FieldName:  utils.Subject,
				Initial:    "*any",
				Substitute: config.NewRSRParsersMustCompile("1001", true, utils.INFIELD_SEP),
				Append:     true,
			},
		},
		Weight: 20,
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//matches TEST_PROFILE1
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "2009",
			utils.Destination: "+492511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err != nil {
		t.Error(err)
	}
}

func testV1FIdxCaUpdateAttributeProfileFromTP(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "TestFilter3",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"3009"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"+492511231234"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply engine.AttributeProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetAttributeProfile", &utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}, &reply); err != nil {
		t.Error(err)
	}
	reply.FilterIDs = []string{"TestFilter3"}
	if err := tFIdxCaRpc.Call("ApierV1.SetAttributeProfile", reply, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	//matches TEST_PROFILE1
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "3009",
			utils.Destination: "+492511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err != nil {
		t.Error(err)
	}
}

func testV1FIdxCaRemoveAttributeProfile(t *testing.T) {
	var resp string
	ev := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "3009",
			utils.Destination: "+492511231234",
		},
	}
	var rplyEv engine.AttrSProcessEventReply
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err != nil {
		t.Error(err)
	}

	ev2 := &utils.CGREvent{
		Tenant:  "cgrates.org",
		ID:      "testAttributeSProcessEvent",
		Context: utils.StringPointer(utils.MetaSessionS),
		Event: map[string]interface{}{
			utils.Account:     "2009",
			utils.Destination: "+492511231234",
		},
	}
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev2, &rplyEv); err != nil {
		t.Error(err)
	}
	//Remove threshold profile that was set form api
	if err := tFIdxCaRpc.Call("ApierV1.RemoveAttributeProfile", &ArgRemoveAttrProfile{Tenant: "cgrates.org",
		ID: "TEST_PROFILE1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.AttributeProfile
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "TEST_PROFILE1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//Remove threshold profile that was set form tariffplan
	if err := tFIdxCaRpc.Call("ApierV1.RemoveAttributeProfile", &ArgRemoveAttrProfile{Tenant: "cgrates.org",
		ID: "ATTR_1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	//Test the remove
	if err := tFIdxCaRpc.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ATTR_1"}, &sqp); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev, &rplyEv); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.AttributeSv1ProcessEvent, ev2, &rplyEv); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

// ResourceProfile
func testV1FIdxCaGetResourceProfileWithNotFound(t *testing.T) {
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &reply); err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}
func testV1FIdxCaSetResourceProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_RES_RCFG1",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1001"},
			},
			{
				FieldName: utils.Subject,
				Type:      "*string",
				Values:    []string{"1002"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"1001"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	rlsConfig = &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RCFG1",
		FilterIDs: []string{"FLTR_RES_RCFG1"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(0) * time.Microsecond,
		AllocationMessage: "Approved",
		Limit:             10,
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"Val1", "Val2"},
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetResourceProfile", rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1002",
				utils.Destination: "1001"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &result); err != nil {
		t.Error(err)
	} else if result != "Approved" {
		t.Error("Unexpected reply returned", result)
	}

	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &result); err != nil {
		t.Error(err)
	} else if result != "Approved" {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1FIdxCaGetResourceProfileFromTP(t *testing.T) {
	var reply string
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e63",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1001",
				utils.Subject:     "1002",
				utils.Destination: "1001"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AllocateResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Error("Unexpected reply returned", reply)
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources,
		argsRU, &reply); err != nil {
		t.Error(err)
	} else if reply != "Approved" {
		t.Error("Unexpected reply returned", reply)
	}

	argsReU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources,
		argsReU, &reply); err != nil {
		t.Error(err)
	} else if reply != "ResGroup1" {
		t.Error("Unexpected reply returned", reply)
	}
}

func testV1FIdxCaUpdateResourceProfile(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_RES_RCFG2",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"2002"},
			},
			{
				FieldName: utils.Subject,
				Type:      "*string",
				Values:    []string{"2001"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"2002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	rlsConfig = &engine.ResourceProfile{
		Tenant:    "cgrates.org",
		ID:        "RCFG1",
		FilterIDs: []string{"FLTR_RES_RCFG2"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
		UsageTTL:          time.Duration(10) * time.Microsecond,
		Limit:             10,
		AllocationMessage: "MessageAllocation",
		Blocker:           true,
		Stored:            true,
		Weight:            20,
		ThresholdIDs:      []string{"Val1", "Val2"},
	}
	if err := tFIdxCaRpc.Call("ApierV1.SetResourceProfile",
		rlsConfig, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	argsReU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "2002",
				utils.Subject:     "2001",
				utils.Destination: "2002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources,
		argsReU, &result); err != nil {
		t.Error(err)
	} else if result != "MessageAllocation" {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1FIdxCaUpdateResourceProfileFromTP(t *testing.T) {
	filter = &engine.Filter{
		Tenant: "cgrates.org",
		ID:     "FLTR_RES_RCFG3",
		Rules: []*engine.FilterRule{
			{
				FieldName: utils.Account,
				Type:      "*string",
				Values:    []string{"1002"},
			},
			{
				FieldName: utils.Subject,
				Type:      "*string",
				Values:    []string{"1001"},
			},
			{
				FieldName: utils.Destination,
				Type:      "*string",
				Values:    []string{"1002"},
			},
		},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC),
		},
	}
	var result string
	if err := tFIdxCaRpc.Call("ApierV1.SetFilter", filter, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply engine.ResourceProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}, &reply); err != nil {
		t.Error(err)
	}
	reply.FilterIDs = []string{"FLTR_RES_RCFG3"}
	reply.ActivationInterval = &utils.ActivationInterval{ActivationTime: time.Date(2014, 7, 14, 14, 25, 0, 0, time.UTC)}

	if err := tFIdxCaRpc.Call("ApierV1.SetResourceProfile", &reply, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	argsReU := utils.ArgRSv1ResourceUsage{
		UsageID: "651a8db2-4f67-4cf8-b622-169e8a482e65",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources, argsReU, &result); err != nil {
		t.Error(err)
	} else if result != "ResGroup1" {
		t.Error("Unexpected reply returned", result)
	}
}

func testV1FIdxCaRemoveResourceProfile(t *testing.T) {
	var resp string
	argsReU := utils.ArgRSv1ResourceUsage{
		UsageID: "653a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "2002",
				utils.Subject:     "2001",
				utils.Destination: "2002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AllocateResources, argsReU, &resp); err != nil {
		t.Error(err)
	} else if resp != "MessageAllocation" {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources, argsReU, &resp); err != nil {
		t.Error(err)
	} else if resp != "MessageAllocation" {
		t.Error("Unexpected reply returned", resp)
	}
	argsRU := utils.ArgRSv1ResourceUsage{
		UsageID: "654a8db2-4f67-4cf8-b622-169e8a482e61",
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Account:     "1002",
				utils.Subject:     "1001",
				utils.Destination: "1002"},
		},
		Units: 6,
	}
	if err := tFIdxCaRpc.Call(utils.ResourceSv1AuthorizeResources, argsRU, &resp); err != nil {
		t.Error(err)
	} else if resp != "ResGroup1" {
		t.Error("Unexpected reply returned", resp)
	}

	if err := tFIdxCaRpc.Call("ApierV1.RemoveResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	if err := tFIdxCaRpc.Call("ApierV1.RemoveResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var sqp *engine.ThresholdProfile
	if err := tFIdxCaRpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "RCFG1"}, &sqp); err == nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := tFIdxCaRpc.Call("ApierV1.GetResourceProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ResGroup1"}, &sqp); err == nil &&
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testV1FIdxCaStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
