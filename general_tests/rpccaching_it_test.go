//go:build integration
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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	rpcCfgPath string
	rpcCfg     *config.CGRConfig
	rpcRpc     *birpc.Client
	rpcConfDIR string //run tests for specific configuration

	sTestsRPCMethods = []func(t *testing.T){
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
		testRPCMethodsProcessCDR,
		testRPCMethodsProcessEvent,
		// reset the storDB and dataDB
		testRPCMethodsInitDataDb,
		testRPCMethodsResetStorDb,
		testRPCMethodsCdrsProcessCDR,
		testRPCMethodsCdrsStoreSessionCost,
		//reset the storDB and dataDB
		testRPCMethodsInitDataDb,
		testRPCMethodsResetStorDb,
		testRPCMethodsLoadData,
		testRPCMethodsResponderDebit,
		testRPCMethodsResponderMaxDebit,
		testRPCMethodsStopEngine,
	}
)

// Test start here
func TestRPCMethods(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		rpcConfDIR = "rpccaching_mysql"
	case utils.MetaMongo:
		rpcConfDIR = "rpccaching_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsRPCMethods {
		t.Run(rpcConfDIR, stest)
	}
}

func testRPCMethodsLoadConfig(t *testing.T) {
	var err error
	rpcCfgPath = path.Join(*utils.DataDir, "conf", "samples", rpcConfDIR)
	if rpcCfg, err = config.NewCGRConfigFromPath(rpcCfgPath); err != nil {
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
	if _, err := engine.StopStartEngine(rpcCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testRPCMethodsRpcConn(t *testing.T) {
	rpcRpc = engine.NewRPCClient(t, rpcCfg.ListenCfg())
}

func testRPCMethodsFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testRPCMethodsAddData(t *testing.T) {
	var resp string
	if err := rpcRpc.Call(context.Background(), utils.APIerSv1RemoveThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "THD_ACNT_1001"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply string
	// Add a disable and log action
	attrsAA := &utils.AttrSetActions{ActionsId: "DISABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaDisableAccount},
		{Identifier: utils.MetaLog},
	}}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}
	// Add an enable and log action
	attrsAA2 := &utils.AttrSetActions{ActionsId: "ENABLE_LOG", Actions: []*utils.TPAction{
		{Identifier: utils.MetaEnableAccount},
		{Identifier: utils.MetaLog},
	}}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2SetActions, attrsAA2, &reply); err != nil && err.Error() != utils.ErrExists.Error() {
		t.Error("Got error on APIerSv2.SetActions: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv2.SetActions received: %s", reply)
	}

	//Add a thresholdProfile to disable account
	tPrfl := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_AccDisableAndLog",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.DisableAction:DisableAction"},
			MaxHits:   -1,
			MinSleep:  time.Second,
			Weight:    30.0,
			ActionIDs: []string{"DISABLE_LOG"},
		},
	}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	//Add a thresholdProfile to enable account
	tPrfl2 := &engine.ThresholdProfileWithAPIOpts{
		ThresholdProfile: &engine.ThresholdProfile{
			Tenant:    "cgrates.org",
			ID:        "THD_AccEnableAndLog",
			FilterIDs: []string{"*string:~*req.Account:1001", "*string:~*req.EnableAction:EnableAction"},
			MaxHits:   -1,
			MinSleep:  time.Second,
			Weight:    30.0,
			ActionIDs: []string{"ENABLE_LOG"},
		},
	}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv1SetThresholdProfile, tPrfl2, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
}

func testRPCMethodsAuthorizeSession(t *testing.T) {
	authUsage := 5 * time.Minute
	args := &sessions.V1AuthorizeArgs{
		GetMaxUsage: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsAuthorizeSession",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testRPCMethodsAuthorizeSession",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.Usage:        authUsage,
			},
		},
	}
	//authorize the session
	var rplyFirst sessions.V1AuthorizeReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rplyFirst); err != nil {
		t.Fatal(err)
	}
	if rplyFirst.MaxUsage == nil || *rplyFirst.MaxUsage != authUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"DisableAction":    "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
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
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	//authorize again session (should take the response from cache)
	var rply sessions.V1AuthorizeReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	//authorize again session (this time we expect to receive an error)
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"EnableAction":     "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsInitSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsInitSession",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testRPCMethodsInitSession",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
			},
		},
	}
	var rplyFirst sessions.V1InitSessionReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		args, &rplyFirst); err != nil {
		t.Error(err)
	}
	if rplyFirst.MaxUsage == nil || *rplyFirst.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"DisableAction":    "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
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
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	var rply sessions.V1InitSessionReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	if err := rpcRpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		args, &rply); err == nil || !(err.Error() == "RALS_ERROR:ACCOUNT_DISABLED" ||
		err.Error() == utils.ErrExists.Error()) { // ErrExist -> initSession twice
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"EnableAction":     "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsUpdateSession(t *testing.T) {
	reqUsage := 5 * time.Minute
	args := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsUpdateSession",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testRPCMethodsUpdateSession",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        reqUsage,
			},
		},
	}
	var rplyFirst sessions.V1UpdateSessionReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1UpdateSession,
		args, &rplyFirst); err != nil {
		t.Error(err)
	}
	if rplyFirst.MaxUsage == nil || *rplyFirst.MaxUsage != reqUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"DisableAction":    "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
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
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	var rply sessions.V1UpdateSessionReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1UpdateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	if err := rpcRpc.Call(context.Background(), utils.SessionSv1UpdateSession,
		args, &rply); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"EnableAction":     "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsTerminateSession",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testRPCMethodsTerminateSession",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}
	var rply string
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}

	//replace event with empty
	args.CGREvent.Event = map[string]any{}

	if err := rpcRpc.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	if err := rpcRpc.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err == nil || err.Error() != "MANDATORY_IE_MISSING: [OriginID]" {
		t.Error(err)
	}

}

func testRPCMethodsProcessCDR(t *testing.T) {
	args := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testRPCMethodsProcessCDR",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "testRPCMethodsProcessCDR",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "ANY2CNT",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        10 * time.Minute,
		},
	}
	var rply string
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	time.Sleep(100 * time.Millisecond)
	//verify the CDR
	var cdrs []*engine.CDR
	argsCDR := &utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
	//change originID so CGRID be different
	args.Event[utils.OriginID] = "testRPCMethodsProcessCDR2"
	// we should get response from cache
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	time.Sleep(100 * time.Millisecond)
	//verify the CDR
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	//change originID so CGRID be different
	args.Event[utils.OriginID] = "testRPCMethodsProcessCDR3"
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessCDR,
		args, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	time.Sleep(100 * time.Millisecond)
	//verify the CDR
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testRPCMethodsProcessEvent(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1ProcessMessageArgs{
		Debit: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsProcessEvent",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testRPCMethodsProcessEvent",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        initUsage,
			},
		},
	}
	var rplyFirst sessions.V1ProcessMessageReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rplyFirst); err != nil {
		t.Error(err)
	} else if rplyFirst.MaxUsage == nil || *rplyFirst.MaxUsage != initUsage {
		t.Errorf("Unexpected MaxUsage: %v", rplyFirst.MaxUsage)
	}

	//disable the account
	var ids []string
	thEvent := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "DisableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"DisableAction":    "DisableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
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
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrAcc, &acnt); err != nil {
		t.Error(err)
	} else if acnt.Disabled != true {
		t.Errorf("Expecting: true, received: %v", acnt.Disabled)
	}

	//get response from cache
	var rply sessions.V1ProcessMessageReply
	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(rply, rplyFirst) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(rplyFirst), utils.ToJSON(rply))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)

	if err := rpcRpc.Call(context.Background(), utils.SessionSv1ProcessMessage,
		args, &rplyFirst); err == nil || err.Error() != "RALS_ERROR:ACCOUNT_DISABLED" {
		t.Error("Unexpected error returned", err)
	}

	//enable the account
	thEvent = &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "EnableAccount",
		Event: map[string]any{
			utils.AccountField: "1001",
			"EnableAction":     "EnableAction",
		},
	}
	//process event
	if err := rpcRpc.Call(context.Background(), utils.ThresholdSv1ProcessEvent, thEvent, &ids); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ids, []string{"THD_AccEnableAndLog"}) {
		t.Errorf("Expecting ids: %s, received: %s", []string{"THD_AccEnableAndLog"}, ids)
	}
}

func testRPCMethodsCdrsProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testRPCMethodsCdrsProcessCDR",
			Event: map[string]any{
				utils.Tenant:       "cgrates.org",
				utils.ToR:          utils.MetaVoice,
				utils.OriginHost:   "host",
				utils.OriginID:     "testRPCMethodsCdrsProcessCDR",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Minute,
			},
		},
	}

	var reply string
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(150 * time.Millisecond) // Give time for CDR to be rated
	//verify the CDR
	var cdrs []*engine.CDR
	argsCDR := utils.RPCCDRsFilterWithAPIOpts{RPCCDRsFilter: &utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}}}
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
	//change originID so CGRID be different
	args.Event[utils.OriginID] = "testRPCMethodsProcessCDR2"
	// we should get response from cache
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(100 * time.Millisecond)
	//verify the CDR
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)
	//change originID so CGRID be different
	args.Event[utils.OriginID] = "testRPCMethodsProcessCDR4"
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(150 * time.Millisecond) // Give time for CDR to be rated
	//verify the CDR
	if err := rpcRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &argsCDR, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	}
}

func testRPCMethodsCdrsStoreSessionCost(t *testing.T) {
	cc := &engine.CallCost{
		Category:    "generic",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: "data",
		ToR:         "*data",
		Cost:        0,
	}
	args := &engine.ArgsV2CDRSStoreSMCost{
		CheckDuplicate: true,
		Cost: &engine.V2SMCost{
			CGRID:       "testRPCMethodsCdrsStoreSessionCost",
			RunID:       utils.MetaDefault,
			OriginHost:  "",
			OriginID:    "testdatagrp_grp1",
			CostSource:  "SMR",
			Usage:       1536,
			CostDetails: engine.NewEventCostFromCallCost(cc, "testRPCMethodsCdrsStoreSessionCost", utils.MetaDefault),
		},
	}

	var reply string
	if err := rpcRpc.Call(context.Background(), utils.CDRsV2StoreSessionCost, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	time.Sleep(150 * time.Millisecond)

	//change originID so CGRID be different
	args.Cost.CGRID = "testRPCMethodsCdrsStoreSessionCost"
	// we should get response from cache
	if err := rpcRpc.Call(context.Background(), utils.CDRsV2StoreSessionCost, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)
	//change originID so CGRID be different
	args.Cost.CGRID = "testRPCMethodsCdrsStoreSessionCost"
	if err := rpcRpc.Call(context.Background(), utils.CDRsV2StoreSessionCost, args,
		&reply); err == nil || err.Error() != "SERVER_ERROR: EXISTS" {
		t.Error("Unexpected error: ", err.Error())
	}
}

// Load the tariff plan, creating accounts and their balances
func testRPCMethodsLoadData(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testtp")}
	if err := rpcRpc.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &tpLoadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testRPCMethodsResponderDebit(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			CgrID:         "testRPCMethodsResponderDebit",
			Category:      "call",
			Tenant:        "cgrates.org",
			Subject:       "1001",
			Destination:   "+49",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(15 * time.Second),
		},
	}
	var cc engine.CallCost
	//cache the response
	if err := rpcRpc.Call(context.Background(), utils.ResponderDebit, cd, &cc); err != nil {
		t.Error(err)
	} else if cc.GetDuration() != 15*time.Second {
		t.Errorf("Expecting: %+v, \n received: %+v",
			15*time.Second, cc.GetDuration())
	} else if cc.Cost != 15 {
		t.Errorf("Expecting: %+v, \n received: %+v",
			15, cc.Cost)
	}
	cd2 := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			CgrID: "testRPCMethodsResponderDebit",
		},
	}
	var ccCache engine.CallCost
	//cache the response
	if err := rpcRpc.Call(context.Background(), utils.ResponderDebit, cd2, &ccCache); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ccCache, cc) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(cc), utils.ToJSON(ccCache))
	}
	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)
	if err := rpcRpc.Call(context.Background(), utils.ResponderDebit, cd2, &cc); err == nil || err.Error() != "ACCOUNT_NOT_FOUND" {
		t.Error("Unexpected error returned", err)
	}
}

func testRPCMethodsResponderMaxDebit(t *testing.T) {
	tStart := time.Date(2016, 3, 31, 0, 0, 0, 0, time.UTC)
	cd := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			CgrID:         "testRPCMethodsResponderMaxDebit",
			Category:      "call",
			Tenant:        "cgrates.org",
			Account:       "1001",
			Subject:       "free",
			Destination:   "+49",
			DurationIndex: 0,
			TimeStart:     tStart,
			TimeEnd:       tStart.Add(15 * time.Second),
		},
	}
	var cc engine.CallCost
	//cache the response
	if err := rpcRpc.Call(context.Background(), utils.ResponderMaxDebit, cd, &cc); err != nil {
		t.Error(err)
	} else if cc.GetDuration() != 15*time.Second {
		t.Errorf("Expecting: %+v, \n received: %+v",
			15*time.Second, cc.GetDuration())
	} else if cc.Cost != 0 {
		t.Errorf("Expecting: %+v, \n received: %+v",
			0, cc.Cost)
	}
	cd2 := &engine.CallDescriptorWithAPIOpts{
		CallDescriptor: &engine.CallDescriptor{
			CgrID: "testRPCMethodsResponderMaxDebit",
		},
	}
	var ccCache engine.CallCost
	//cache the response
	if err := rpcRpc.Call(context.Background(), utils.ResponderMaxDebit, cd2, &ccCache); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(ccCache, cc) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			utils.ToJSON(cc), utils.ToJSON(ccCache))
	}
	//give time to CGRateS to delete the response from cache
	time.Sleep(time.Second)
	if err := rpcRpc.Call(context.Background(), utils.ResponderMaxDebit, cd2, &cc); err == nil || err.Error() != "ACCOUNT_NOT_FOUND" {
		t.Error("Unexpected error returned", err)
	}
}

func testRPCMethodsStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
