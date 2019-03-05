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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	rpcCfgPath string
	rpcCfg     *config.CGRConfig
	rpcRpc     *rpc.Client
	rpcConfDIR string //run tests for specific configuration
)

var sTestsRPCMethods = []func(t *testing.T){
	testRPCMethodsLoadConfig,
	testRPCMethodsInitDataDb,
	testRPCMethodsResetStorDb,
	testRPCMethodsStartEngine,
	testRPCMethodsRpcConn,
	testRPCMethodsFromFolder,
	testRPCMethodsAddData,
	testRPCMethodsAuthorizeSession,
	testRPCMethodsInitSession,
	testRPCMethodsUpdateSession,
	testRPCMethodsTerminateSession,

	testRPCMethodsStopEngine,
}

// Test start here
func TestRPCMethods(t *testing.T) {
	rpcConfDIR = "rpccaching"
	for _, stest := range sTestsRPCMethods {
		t.Run(rpcConfDIR, stest)
	}
}

func testRPCMethodsLoadConfig(t *testing.T) {
	var err error
	rpcCfgPath = path.Join(*dataDir, "conf", "samples", rpcConfDIR)
	if rpcCfg, err = config.NewCGRConfigFromFolder(rpcCfgPath); err != nil {
		t.Error(err)
	}
}

func testRPCMethodsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rpcCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testRPCMethodsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rpcCfg); err != nil {
		t.Fatal(err)
	}
}

func testRPCMethodsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rpcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testRPCMethodsRpcConn(t *testing.T) {
	var err error
	rpcRpc, err = jsonrpc.Dial("tcp", rpcCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testRPCMethodsFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := rpcRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRPCMethodsAddData(t *testing.T) {
	var resp string
	if err := rpcRpc.Call("ApierV1.RemoveThresholdProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply string
	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "DISABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: engine.DISABLE_ACCOUNT},
		{Identifier: engine.LOG},
	}}
	if err := rpcRpc.Call("ApierV2.SetActions", attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	// Add an enable and log action
	attrsAA2 := &utils.AttrSetActions{ActionsId: "ENABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: engine.ENABLE_ACCOUNT},
		{Identifier: engine.LOG},
	}}
	if err := rpcRpc.Call("ApierV2.SetActions", attrsAA2, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on ApierV2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling ApierV2.SetActions received: %s", reply)
	}
	time.Sleep(10 * time.Millisecond)

	//Add a thresholdProfile to disable account
	tPrfl := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccDisableAndLog",
		FilterIDs: []string{"*string:Account:1001", "*string:DisableAction:DisableAction"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    30.0,
		ActionIDs: []string{"DISABLE_LOG"},
	}
	if err := rpcRpc.Call("ApierV1.SetThresholdProfile", tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//Add a thresholdProfile to enable account
	tPrfl2 := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_AccEnableAndLog",
		FilterIDs: []string{"*string:Account:1001", "*string:EnableAction:EnableAction"},
		MaxHits:   -1,
		MinSleep:  time.Duration(1 * time.Second),
		Weight:    30.0,
		ActionIDs: []string{"ENABLE_LOG"},
	}
	if err := rpcRpc.Call("ApierV1.SetThresholdProfile", tPrfl2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testRPCMethodsAuthorizeSession(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsAuthorizeSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsAuthorizeSession",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:       authUsage,
			},
		},
	}
	//authorize the session
	var rplyFirst sessions.V1AuthorizeReply
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
		t.Fatal(err)
	}
	if *rplyFirst.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]interface{}{
			utils.Account:   "1001",
			"DisableAction": "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccDisableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccDisableAndLog"}, ids)
	}

	//verify if account was disabled
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := rpcRpc.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	//authorize again session (should take the response from cache)
	var rply sessions.V1AuthorizeReply
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(1*time.Second + 500*time.Millisecond)

	//authorize again session (this time we expect to receive an error)
	if err := rpcRpc.Call(utils.SessionSv1AuthorizeEvent, args, &rply); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]interface{}{
			utils.Account:  "1001",
			"EnableAction": "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsInitSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsInitSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsInitSession",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       initUsage,
			},
		},
	}
	var rplyFirst sessions.V1InitSessionReply
	if err := rpcRpc.Call(utils.SessionSv1InitiateSession,
		args, &rplyFirst); err != nil {
		t.Error(err)
	}
	if *rplyFirst.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]interface{}{
			utils.Account:   "1001",
			"DisableAction": "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccDisableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccDisableAndLog"}, ids)
	}

	//verify if account was disabled
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := rpcRpc.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	var rply sessions.V1InitSessionReply
	if err := rpcRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(1*time.Second + 500*time.Millisecond)

	if err := rpcRpc.Call(utils.SessionSv1InitiateSession,
		args, &rply); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]interface{}{
			utils.Account:  "1001",
			"EnableAction": "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsUpdateSession",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       reqUsage,
			},
		},
	}
	var rplyFirst sessions.V1UpdateSessionReply
	if err := rpcRpc.Call(utils.SessionSv1UpdateSession,
		args, &rplyFirst); err != nil {
		t.Error(err)
	}
	if *rplyFirst.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]interface{}{
			utils.Account:   "1001",
			"DisableAction": "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccDisableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccDisableAndLog"}, ids)
	}

	//verify if account was disabled
	var acnt *engine.Account
	attrAcc := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1001",
	}
	if err := rpcRpc.Call("ApierV2.GetAccount", attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	var rply sessions.V1UpdateSessionReply
	if err := rpcRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(1*time.Second + 500*time.Millisecond)

	if err := rpcRpc.Call(utils.SessionSv1UpdateSession,
		args, &rply); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]interface{}{
			utils.Account:  "1001",
			"EnableAction": "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsTerminateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "testRPCMethodsTerminateSession",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Subject:     "ANY2CNT",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
			},
		},
	}
	var rply string
	if err := rpcRpc.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testRPCMethodsStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
