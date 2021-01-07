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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sSv1CfgPath2      string
	sSv1Cfg2          *config.CGRConfig
	sSv1BiRpc2        *rpc2.Client
	sSApierRpc2       *rpc.Client
	disconnectEvChan2 = make(chan *utils.AttrDisconnectSession)
	sessionsConfDIR   string

	sessionsThresholdTests = []func(t *testing.T){
		testSessionSv1ItInitCfg,
		testSessionSv1ItResetDataDb,
		testSessionSv1ItResetStorDb,
		testSessionSv1ItStartEngine,
		testSessionSv1ItRpcConn,
		testSessionSv1ItTPFromFolder,
		testSessionSv1ItGetThreshold,
		testSessionSv1ItAuth,
		testSessionSv1ItInitiateSession,
		testSessionSv1ItTerminateSession,
		testSessionSv1ItAuthNotFoundThreshold,
		testSessionSv1ItInitNotFoundThreshold,
		testSessionSv1ItTerminateNotFoundThreshold,
		testSessionSv1ItAuthNotFoundThresholdAndStats,
		testSessionSv1ItInitNotFoundThresholdAndStats,
		testSessionSv1ItTerminateNotFoundThresholdAndStats,
		testSessionSv1ItStopCgrEngine,
	}
)

func TestSessionSITtests(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		sessionsConfDIR = "sessions_internal"
	case utils.MetaMySQL:
		sessionsConfDIR = "sessions_mysql"
	case utils.MetaMongo:
		sessionsConfDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sessionsThresholdTests {
		t.Run(sessionsConfDIR, stest)
	}
}

func handleDisconnectSession2(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan2 <- args
	*reply = utils.OK
	return nil
}

func testSessionSv1ItInitCfg(t *testing.T) {
	var err error
	sSv1CfgPath2 = path.Join(*dataDir, "conf", "samples", sessionsConfDIR)
	// Init config first
	sSv1Cfg2, err = config.NewCGRConfigFromPath(sSv1CfgPath2)
	if err != nil {
		t.Error(err)
	}
}

func testSessionSv1ItResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(sSv1Cfg2); err != nil {
		t.Fatal(err)
	}
}

func testSessionSv1ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sSv1Cfg2); err != nil {
		t.Fatal(err)
	}
}

func testSessionSv1ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sSv1CfgPath2, 100); err != nil {
		t.Fatal(err)
	}
}

func testSessionSv1ItRpcConn(t *testing.T) {
	dummyClnt, err := utils.NewBiJSONrpcClient(sSv1Cfg2.SessionSCfg().ListenBijson,
		nil)
	if err != nil {
		t.Fatal(err)
	}
	clntHandlers := map[string]interface{}{
		utils.SessionSv1DisconnectSession: handleDisconnectSession2,
	}
	if sSv1BiRpc2, err = utils.NewBiJSONrpcClient(sSv1Cfg2.SessionSCfg().ListenBijson,
		clntHandlers); err != nil {
		t.Fatal(err)
	}
	if sSApierRpc2, err = newRPCClient(sSv1Cfg2.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

// Load the tariff plan, creating accounts and their balances
func testSessionSv1ItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := sSApierRpc2.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSessionSv1ItGetThreshold(t *testing.T) {
	tPrfl := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinSleep:  0,
		Blocker:   false,
		Weight:    10.0,
		ActionIDs: []string{"TOPUP_MONETARY_10"},
		Async:     false,
	}
	var reply *engine.ThresholdProfile
	if err := sSApierRpc2.Call(utils.APIerSv1GetThresholdProfile,
		&utils.TenantID{Tenant: "cgrates.org",
			ID: "THD_ACNT_1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(tPrfl, reply) {
		t.Errorf("Expecting: %+v, received: %+v",
			utils.ToJSON(tPrfl), utils.ToJSON(reply))
	}
	// Verify account before authorization
	expectedAccount := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MONETARY: []*engine.Balance{
				{
					Value:  10,
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply2 *engine.Account
	if err := sSApierRpc2.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org",
			Account: "1001"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccount.ID, reply2.ID) {
		t.Errorf("Expecting: %s, received: %s",
			expectedAccount.ID, reply2.ID)
	} else if !reflect.DeepEqual(
		expectedAccount.BalanceMap[utils.MONETARY][0].Value,
		reply2.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %f, received: %f",
			expectedAccount.BalanceMap[utils.MONETARY][0].Value,
			reply2.BalanceMap[utils.MONETARY][0].Value)
	}
}

func testSessionSv1ItAuth(t *testing.T) {
	args := &sessions.V1AuthorizeArgs{
		AuthorizeResources: true,
		ProcessThresholds:  true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItAuth",
				Event: map[string]interface{}{
					utils.OriginID:     "TestSSv1It1",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1AuthorizeEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if *rply.ResourceAllocation == "" {
		t.Errorf("Unexpected ResourceAllocation: %s", *rply.ResourceAllocation)
	}
	if !reflect.DeepEqual(*rply.ThresholdIDs, []string{"THD_ACNT_1001"}) {
		t.Errorf("Unexpected ThresholdIDs: %v", *rply.ThresholdIDs)
	}
	// Hit threshold and execute action (topup with 10 units)
	expectedAccount := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MONETARY: []*engine.Balance{
				{
					Value:  20,
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply *engine.Account
	if err := sSApierRpc2.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org",
			Account: "1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccount.ID, reply.ID) {
		t.Errorf("Expecting: %s, received: %s",
			expectedAccount.ID, reply.ID)
	} else if !reflect.DeepEqual(
		expectedAccount.BalanceMap[utils.MONETARY][0].Value,
		reply.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %f, received: %f",
			expectedAccount.BalanceMap[utils.MONETARY][0].Value,
			reply.BalanceMap[utils.MONETARY][0].Value)
	}
}

func testSessionSv1ItInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItInitiateSession",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.VOICE,
					utils.OriginID:     "TestSSv1It1",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime: time.Date(2018,
						time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage: initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(*rply.ThresholdIDs, []string{"THD_ACNT_1001"}) {
		t.Errorf("Unexpected ThresholdIDs: %v", *rply.ThresholdIDs)
	}
	expectedAccount := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MONETARY: []*engine.Balance{
				{
					Value:  29.898000,
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply *engine.Account
	if err := sSApierRpc2.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org",
			Account: "1001"}, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccount.ID, reply.ID) {
		t.Errorf("Expecting: %s, received: %s",
			expectedAccount.ID, reply.ID)
	} else if !reflect.DeepEqual(
		expectedAccount.BalanceMap[utils.MONETARY][0].Value,
		reply.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %f, received: %f",
			expectedAccount.BalanceMap[utils.MONETARY][0].Value,
			reply.BalanceMap[utils.MONETARY][0].Value)
	}
}

func testSessionSv1ItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSSv1ItTerminateSession",
				Event: map[string]interface{}{
					utils.OriginID:     "TestSSv1It1",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        10 * time.Minute,
				},
			},
		},
	}
	var rply string
	if err := sSv1BiRpc2.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	expectedAccount := &engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MONETARY: []*engine.Balance{
				{
					Value:  39.796000,
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply2 *engine.Account
	if err := sSApierRpc2.Call(utils.APIerSv2GetAccount,
		&utils.AttrGetAccount{Tenant: "cgrates.org",
			Account: "1001"}, &reply2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedAccount.ID, reply2.ID) {
		t.Errorf("Expecting: %s, received: %s",
			expectedAccount.ID, reply2.ID)
	} else if !reflect.DeepEqual(
		expectedAccount.BalanceMap[utils.MONETARY][0].Value,
		reply2.BalanceMap[utils.MONETARY][0].Value) {
		t.Errorf("Expecting: %f, received: %f",
			expectedAccount.BalanceMap[utils.MONETARY][0].Value,
			reply2.BalanceMap[utils.MONETARY][0].Value)
	}
}

func testSessionSv1ItAuthNotFoundThreshold(t *testing.T) {
	args := &sessions.V1AuthorizeArgs{
		ProcessStats:      true,
		GetMaxUsage:       true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSesssonSv1ItNotFoundThreshold",
				Event: map[string]interface{}{
					utils.OriginID:     "TestSesssonSv1ItNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1AuthorizeEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.ThresholdIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.ThresholdIDs)
	}
	if rply.StatQueueIDs != nil && len(*rply.StatQueueIDs) != 1 && (*rply.StatQueueIDs)[0] != "Stat_2" {
		t.Errorf("Unexpected StatQueueIDs: %+v", rply.StatQueueIDs)
	}
}

func testSessionSv1ItInitNotFoundThreshold(t *testing.T) {
	initUsage := 1024
	args := &sessions.V1InitSessionArgs{
		ProcessStats:      true,
		InitSession:       true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSessionSv1ItInitNotFoundThreshold",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.DATA,
					utils.OriginID:     "TestSessionSv1ItInitNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Subject:      "RP_ANY2CNT",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime: time.Date(2018,
						time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage: initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != 1024 {
		t.Errorf("Expecting: %+v, received: %+v",
			1024, rply.MaxUsage)
	}
	if rply.ThresholdIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.ThresholdIDs)
	}
	if rply.StatQueueIDs != nil && len(*rply.StatQueueIDs) != 1 && (*rply.StatQueueIDs)[0] != "Stat_2" {
		t.Errorf("Unexpected StatQueueIDs: %+v", rply.StatQueueIDs)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc2.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testSessionSv1ItTerminateNotFoundThreshold(t *testing.T) {
	initUsage := 1024
	args := &sessions.V1TerminateSessionArgs{
		ProcessStats:      true,
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSessionSv1ItTerminateNotFoundThreshold",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.DATA,
					utils.OriginID:     "TestSessionSv1ItInitNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Subject:      "RP_ANY2CNT",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime: time.Date(2018,
						time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage: initUsage,
				},
			},
		},
	}
	var rply string
	if err := sSv1BiRpc2.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc2.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSessionSv1ItAuthNotFoundThresholdAndStats(t *testing.T) {
	var resp string
	if err := sSApierRpc2.Call(utils.APIerSv1RemoveStatQueueProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "Stat_2"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}

	args := &sessions.V1AuthorizeArgs{
		ProcessStats:      true,
		GetMaxUsage:       true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSesssonSv1ItNotFoundThreshold",
				Event: map[string]interface{}{
					utils.OriginID:     "TestSesssonSv1ItNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
				},
			},
		},
	}
	var rply sessions.V1AuthorizeReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1AuthorizeEvent,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.ThresholdIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.ThresholdIDs)
	}
	if rply.StatQueueIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.StatQueueIDs)
	}
}

func testSessionSv1ItInitNotFoundThresholdAndStats(t *testing.T) {
	initUsage := 1024
	args := &sessions.V1InitSessionArgs{
		ProcessStats:      true,
		InitSession:       true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSessionSv1ItInitNotFoundThreshold",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.DATA,
					utils.OriginID:     "TestSessionSv1ItInitNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Subject:      "RP_ANY2CNT",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime: time.Date(2018,
						time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage: initUsage,
				},
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := sSv1BiRpc2.Call(utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply.MaxUsage == nil || *rply.MaxUsage != 1024 {
		t.Errorf("Expecting: %+v, received: %+v",
			1024, rply.MaxUsage)
	}
	if rply.ThresholdIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.ThresholdIDs)
	}
	if rply.StatQueueIDs != nil {
		t.Errorf("Expecting: nil, received: %s",
			rply.StatQueueIDs)
	}

	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc2.Call(utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("wrong active sessions: %s \n , and len(aSessions) %+v", utils.ToJSON(aSessions), len(aSessions))
	}
}

func testSessionSv1ItTerminateNotFoundThresholdAndStats(t *testing.T) {
	initUsage := 1024
	args := &sessions.V1TerminateSessionArgs{
		ProcessStats:      true,
		TerminateSession:  true,
		ProcessThresholds: true,
		CGREventWithOpts: &utils.CGREventWithOpts{
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestSessionSv1ItTerminateNotFoundThreshold",
				Event: map[string]interface{}{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.DATA,
					utils.OriginID:     "TestSessionSv1ItInitNotFoundThreshold",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1002",
					utils.Subject:      "RP_ANY2CNT",
					utils.Destination:  "1001",
					utils.SetupTime: time.Date(2018,
						time.January, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime: time.Date(2018,
						time.January, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage: initUsage,
				},
			},
		},
	}
	var rply string
	if err := sSv1BiRpc2.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sSv1BiRpc2.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSessionSv1ItStopCgrEngine(t *testing.T) {
	if err := sSv1BiRpc2.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
