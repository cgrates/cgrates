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

package dispatchers

import (
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sTestsDspCDRs = []func(t *testing.T){
		testDspCDRsPing,
		testDspCDRsProcessEvent,
		testDspCDRsCountCDR,
		testDspCDRsGetCDR,
		testDspCDRsGetCDRWithoutTenant,
		testDspCDRsProcessCDR,
		testDspCDRsGetCDR2,
		testDspCDRsProcessExternalCDR,
		testDspCDRsGetCDR3,
		testDspCDRsV2ProcessEvent,
		// testDspCDRsV2StoreSessionCost,
	}

	sTestsDspCDRsWithoutAuth = []func(t *testing.T){
		testDspCDRsPingNoAuth,
		testDspCDRsProcessEventNoAuth,
		testDspCDRsCountCDRNoAuth,
		testDspCDRsGetCDRNoAuth,
		testDspCDRsGetCDRNoAuthWithoutTenant,
		testDspCDRsProcessCDRNoAuth,
		testDspCDRsGetCDR2NoAuth,
		testDspCDRsProcessExternalCDRNoAuth,
		testDspCDRsGetCDR3NoAuth,
		testDspCDRsV2ProcessEventNoAuth,
		// testDspCDRsV2StoreSessionCostNoAuth,
	}
)

//Test start here
func TestDspCDRsIT(t *testing.T) {
	var config1, config2, config3 string
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		config1 = "all_mysql"
		config2 = "all2_mysql"
		config3 = "dispatchers_mysql"
	case utils.MetaMongo:
		config1 = "all_mongo"
		config2 = "all2_mongo"
		config3 = "dispatchers_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	dispDIR := "dispatchers"
	if *encoding == utils.MetaGOB {
		dispDIR += "_gob"
	}
	testDsp(t, sTestsDspCDRs, "TestDspCDRs", config1, config2, config3, "tutorial", "oldtutorial", dispDIR)
}

func TestDspCDRsITMySQLWithoutAuth(t *testing.T) {
	if *dbType != utils.MetaMySQL {
		t.SkipNow()
	}
	if *encoding == utils.MetaGOB {
		testDsp(t, sTestsDspCDRsWithoutAuth, "TestDspCDRsWithoutAuth", "all_mysql", "all2_mysql", "dispatchers_no_attributes", "tutorial", "oldtutorial", "dispatchers_gob")
	} else {
		testDsp(t, sTestsDspCDRsWithoutAuth, "TestDspCDRsWithoutAuth", "all_mysql", "all2_mysql", "dispatchers_no_attributes", "tutorial", "oldtutorial", "dispatchers")
	}
}

func testDspCDRsPing(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.CDRsV1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsProcessEvent(t *testing.T) {
	var reply string
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:     "testDspCDRsProcessEvent",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testDspCDRsProcessEvent",
				utils.RequestType:  utils.MetaRated,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},

			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "cdrs12345",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsCountCDR(t *testing.T) {
	var reply int64
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRsCount, args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Received: %+v", reply)
	}
}

func testDspCDRsGetCDR(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "9ee4c71fcd67eef5fb25a4bb3f190487de3073f5" {
		t.Errorf("Expected: 9ee4c71fcd67eef5fb25a4bb3f190487de3073f5 , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsGetCDRWithoutTenant(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "9ee4c71fcd67eef5fb25a4bb3f190487de3073f5" {
		t.Errorf("Expected: 9ee4c71fcd67eef5fb25a4bb3f190487de3073f5 , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsProcessCDR(t *testing.T) {
	var reply string
	args := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			Tenant:      "cgrates.org",
			OriginID:    "testDspCDRsProcessCDR",
			OriginHost:  "192.168.1.1",
			Source:      "testDspCDRsProcessCDR",
			RequestType: utils.MetaRated,
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       2 * time.Minute,
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsGetCDR2(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts:  []string{"1001"},
			RunIDs:    []string{utils.MetaDefault},
			OriginIDs: []string{"testDspCDRsProcessCDR"},
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "f08dfd32930b6bea326bb8ec4e38ab03d781c0bf" {
		t.Errorf("Expected: f08dfd32930b6bea326bb8ec4e38ab03d781c0bf , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsProcessExternalCDR(t *testing.T) {
	var reply string
	args := &engine.ExternalCDRWithAPIOpts{
		ExternalCDR: &engine.ExternalCDR{
			ToR:         utils.MetaVoice,
			OriginID:    "testDspCDRsProcessExternalCDR",
			OriginHost:  "127.0.0.1",
			Source:      utils.UnitTest,
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1001",
			SetupTime:   "2014-08-04T13:00:00Z",
			AnswerTime:  "2014-08-04T13:00:07Z",
			Usage:       "1s",
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsGetCDR3(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts:  []string{"1003"},
			RunIDs:    []string{utils.MetaDefault},
			OriginIDs: []string{"testDspCDRsProcessExternalCDR"},
		},
		Tenant: "cgrates.org",
		APIOpts: map[string]interface{}{
			utils.OptsAPIKey: "cdrs12345",
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "8ae63781b39f3265d014d2ba6a70437172fba46d" {
		t.Errorf("Expected: 8ae63781b39f3265d014d2ba6a70437172fba46d , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsV2ProcessEvent(t *testing.T) {
	var reply []*utils.EventWithFlags
	args := &engine.ArgV1ProcessEvent{
		// Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:     "testDspCDRsV2ProcessEvent",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testDspCDRsV2ProcessEvent",
				utils.RequestType:  utils.MetaRated,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
			APIOpts: map[string]interface{}{
				utils.OptsAPIKey: "cdrsv212345",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV2ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 2 {
		for _, procEv := range reply {
			if procEv.Event[utils.RequestType] == utils.MetaRated && procEv.Event[utils.Cost] != 0.6 {
				t.Errorf("Expected: %+v , received: %v", 0.6, procEv.Event[utils.Cost])
			}
		}
	}
}

// func testDspCDRsV2StoreSessionCost(t *testing.T) {
// 	var reply string
// 	cc := &engine.CallCost{
// 		Category:    "generic",
// 		Tenant:      "cgrates.org",
// 		Subject:     "1001",
// 		Account:     "1001",
// 		Destination: "data",
// 		ToR:         "*data",
// 		Cost:        0,
// 	}
// 	args := &engine.ArgsV2CDRSStoreSMCost{
// 		CheckDuplicate: true,
// 		Cost: &engine.V2SMCost{
// 			CGRID:       "testDspCDRsV2StoreSessionCost",
// 			RunID:       utils.MetaDefault,
// 			OriginHost:  "",
// 			OriginID:    "testdatagrp_grp1",
// 			CostSource:  "SMR",
// 			Usage:       1536,
// 			CostDetails: engine.NewEventCostFromCallCost(cc, "testDspCDRsV2StoreSessionCost", utils.MetaDefault),
// 		},
// 		APIOpts: map[string]interface{}{
// 			utils.OptsAPIKey: "cdrsv212345",
// 		},
// 	}

// 	if err := dispEngine.RPC.Call(utils.CDRsV2StoreSessionCost, args, &reply); err != nil {
// 		t.Error("Unexpected error: ", err.Error())
// 	} else if reply != utils.OK {
// 		t.Error("Unexpected reply received: ", reply)
// 	}
// 	time.Sleep(150 * time.Millisecond)
// 	if err := dispEngine.RPC.Call(utils.CDRsV2StoreSessionCost, args,
// 		&reply); err == nil || err.Error() != "SERVER_ERROR: EXISTS" {
// 		t.Error("Unexpected error: ", err)
// 	}
// }

func testDspCDRsPingNoAuth(t *testing.T) {
	var reply string
	if err := allEngine.RPC.Call(utils.CDRsV1Ping, new(utils.CGREvent), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1Ping, &utils.CGREvent{
		Tenant: "cgrates.org",
	}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.Pong {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsProcessEventNoAuth(t *testing.T) {
	var reply string
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:     "testDspCDRsProcessEvent",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testDspCDRsProcessEvent",
				utils.RequestType:  utils.MetaRated,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsCountCDRNoAuth(t *testing.T) {
	var reply int64
	args := &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		Tenant: "cgrates.org",
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRsCount, args, &reply); err != nil {
		t.Error(err)
	} else if reply != 1 {
		t.Errorf("Received: %+v", reply)
	}
}

func testDspCDRsGetCDRNoAuth(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
		Tenant: "cgrates.org",
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "9ee4c71fcd67eef5fb25a4bb3f190487de3073f5" {
		t.Errorf("Expected: 9ee4c71fcd67eef5fb25a4bb3f190487de3073f5 , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsGetCDRNoAuthWithoutTenant(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts: []string{"1001"},
			RunIDs:   []string{utils.MetaDefault},
		},
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "9ee4c71fcd67eef5fb25a4bb3f190487de3073f5" {
		t.Errorf("Expected: 9ee4c71fcd67eef5fb25a4bb3f190487de3073f5 , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsProcessCDRNoAuth(t *testing.T) {
	var reply string
	args := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			Tenant:      "cgrates.org",
			OriginID:    "testDspCDRsProcessCDR",
			OriginHost:  "192.168.1.1",
			Source:      "testDspCDRsProcessCDR",
			RequestType: utils.MetaRated,
			Account:     "1001",
			Subject:     "1001",
			Destination: "1002",
			AnswerTime:  time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
			Usage:       2 * time.Minute,
		},
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsGetCDR2NoAuth(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts:  []string{"1001"},
			RunIDs:    []string{utils.MetaDefault},
			OriginIDs: []string{"testDspCDRsProcessCDR"},
		},
		Tenant: "cgrates.org",
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "f08dfd32930b6bea326bb8ec4e38ab03d781c0bf" {
		t.Errorf("Expected: f08dfd32930b6bea326bb8ec4e38ab03d781c0bf , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsProcessExternalCDRNoAuth(t *testing.T) {
	var reply string
	args := &engine.ExternalCDRWithAPIOpts{
		ExternalCDR: &engine.ExternalCDR{
			ToR:         utils.MetaVoice,
			OriginID:    "testDspCDRsProcessExternalCDR",
			OriginHost:  "127.0.0.1",
			Source:      utils.UnitTest,
			RequestType: utils.MetaRated,
			Tenant:      "cgrates.org",
			Category:    "call",
			Account:     "1003",
			Subject:     "1003",
			Destination: "1001",
			SetupTime:   "2014-08-04T13:00:00Z",
			AnswerTime:  "2014-08-04T13:00:07Z",
			Usage:       "1s",
			ExtraFields: map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
		},
	}
	if err := dispEngine.RPC.Call(utils.CDRsV1ProcessExternalCDR, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testDspCDRsGetCDR3NoAuth(t *testing.T) {
	var reply []*engine.CDR
	args := utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			Accounts:  []string{"1003"},
			RunIDs:    []string{utils.MetaDefault},
			OriginIDs: []string{"testDspCDRsProcessExternalCDR"},
		},
		Tenant: "cgrates.org",
	}

	if err := dispEngine.RPC.Call(utils.CDRsV1GetCDRs, &args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 1 {
		t.Errorf("Received: %+v", reply)
	} else if reply[0].CGRID != "8ae63781b39f3265d014d2ba6a70437172fba46d" {
		t.Errorf("Expected: 8ae63781b39f3265d014d2ba6a70437172fba46d , received:%v", reply[0].CGRID)
	}
}

func testDspCDRsV2ProcessEventNoAuth(t *testing.T) {
	var reply []*utils.EventWithFlags
	args := &engine.ArgV1ProcessEvent{
		// Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.OriginID:     "testDspCDRsV2ProcessEventNoAuth",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testDspCDRsV2ProcessEventNoAuth",
				utils.RequestType:  utils.MetaRated,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}
	if err := dispEngine.RPC.Call(utils.CDRsV2ProcessEvent, args, &reply); err != nil {
		t.Error(err)
	} else if len(reply) != 2 {
		for _, procEv := range reply {
			if procEv.Event[utils.RequestType] == utils.MetaRated && procEv.Event[utils.Cost] != 0.6 {
				t.Errorf("Expected: %+v , received: %v", 0.6, procEv.Event[utils.Cost])
			}
		}
	}
}

func TestDspCDRsV1PingError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.CGREvent{}
	var reply *string
	result := dspSrv.CDRsV1Ping(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1PingNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.CGREvent{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CDRsV1Ping(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1PingNilError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	var reply *string
	result := dspSrv.CDRsV1Ping(nil, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1GetCDRsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.RPCCDRsFilterWithAPIOpts{}
	var reply *[]*engine.CDR
	result := dspSrv.CDRsV1GetCDRs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1GetCDRsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.RPCCDRsFilterWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *[]*engine.CDR
	result := dspSrv.CDRsV1GetCDRs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1GetCDRsCountError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &utils.RPCCDRsFilterWithAPIOpts{}
	var reply *int64
	result := dspSrv.CDRsV1GetCDRsCount(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1GetCDRsCountNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &utils.RPCCDRsFilterWithAPIOpts{
		Tenant: "tenant",
	}
	var reply *int64
	result := dspSrv.CDRsV1GetCDRsCount(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1RateCDRsError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ArgRateCDRs{}
	var reply *string
	result := dspSrv.CDRsV1RateCDRs(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1RateCDRsNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ArgRateCDRs{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CDRsV1RateCDRs(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessExternalCDRError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ExternalCDRWithAPIOpts{
		ExternalCDR: &engine.ExternalCDR{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessExternalCDR(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessExternalCDRNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ExternalCDRWithAPIOpts{
		ExternalCDR: &engine.ExternalCDR{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessExternalCDR(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessCDRError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessCDR(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1ProcessCDRNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.CDRWithAPIOpts{
		CDR: &engine.CDR{
			Tenant: "tenant",
		},
	}
	var reply *string
	result := dspSrv.CDRsV1ProcessCDR(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV2ProcessEventError(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ArgV1ProcessEvent{
		Flags: nil,
		CGREvent: utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]*utils.EventWithFlags
	result := dspSrv.CDRsV2ProcessEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV2ProcessEventNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ArgV1ProcessEvent{
		Flags: nil,
		CGREvent: utils.CGREvent{
			Tenant: "tenant",
		},
	}
	var reply *[]*utils.EventWithFlags
	result := dspSrv.CDRsV2ProcessEvent(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV2ProcessEventErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ArgV1ProcessEvent{
		Flags:    nil,
		CGREvent: utils.CGREvent{},
	}
	var reply *[]*utils.EventWithFlags
	result := dspSrv.CDRsV2ProcessEvent(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1StoreSessionCostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.AttrCDRSStoreSMCost{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CDRsV1StoreSessionCost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV1StoreSessionCostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.AttrCDRSStoreSMCost{}
	var reply *string
	result := dspSrv.CDRsV1StoreSessionCost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV2StoreSessionCostNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	CGREvent := &engine.ArgsV2CDRSStoreSMCost{
		Tenant: "tenant",
	}
	var reply *string
	result := dspSrv.CDRsV2StoreSessionCost(CGREvent, reply)
	expected := "DISPATCHER_ERROR:NOT_FOUND"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}

func TestDspCDRsV2StoreSessionCostErrorNil(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	dspSrv := NewDispatcherService(nil, cgrCfg, nil, nil)
	cgrCfg.DispatcherSCfg().AttributeSConns = []string{"test"}
	CGREvent := &engine.ArgsV2CDRSStoreSMCost{}
	var reply *string
	result := dspSrv.CDRsV2StoreSessionCost(CGREvent, reply)
	expected := "MANDATORY_IE_MISSING: [ApiKey]"
	if result == nil || result.Error() != expected {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", expected, result)
	}
}
