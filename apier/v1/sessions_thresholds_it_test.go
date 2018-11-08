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
)

func handleDisconnectSession2(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan2 <- args
	*reply = utils.OK
	return nil
}

func TestSessionSv1ItInitCfg(t *testing.T) {
	var err error
	sSv1CfgPath2 = path.Join(*dataDir, "conf", "samples", "sessions")
	// Init config first
	sSv1Cfg2, err = config.NewCGRConfigFromFolder(sSv1CfgPath2)
	if err != nil {
		t.Error(err)
	}
	sSv1Cfg2.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(sSv1Cfg2)
}

func TestSessionSv1ItResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(sSv1Cfg2); err != nil {
		t.Fatal(err)
	}
}

func TestSessionSv1ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sSv1Cfg2); err != nil {
		t.Fatal(err)
	}
}

func TestSessionSv1ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sSv1CfgPath2, 100); err != nil {
		t.Fatal(err)
	}
}

func TestSessionSv1ItRpcConn(t *testing.T) {
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
	if sSApierRpc2, err = jsonrpc.Dial("tcp", sSv1Cfg2.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

// Load the tariff plan, creating accounts and their balances
func TestSessionSv1ItTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	var loadInst utils.LoadInstance
	if err := sSApierRpc2.Call(utils.ApierV2LoadTariffPlanFromFolder,
		attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func TestSessionSv1ItGetThreshold(t *testing.T) {
	tPrfl := &engine.ThresholdProfile{
		Tenant:    "cgrates.org",
		ID:        "THD_ACNT_1001",
		FilterIDs: []string{"FLTR_ACCOUNT_1001"},
		ActivationInterval: &utils.ActivationInterval{
			ActivationTime: time.Date(2014, 7, 29, 15, 0, 0, 0, time.UTC),
		},
		MaxHits:   -1,
		MinSleep:  time.Duration(0),
		Blocker:   false,
		Weight:    10.0,
		ActionIDs: []string{"TOPUP_MONETARY_10"},
		Async:     false,
	}
	var reply *engine.ThresholdProfile
	if err := sSApierRpc2.Call("ApierV1.GetThresholdProfile",
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
					//Uuid:  "c9a2c620-5256-483a-a92d-c51e94bb7667",
					Value: 10,
					Directions: utils.StringMap{
						"*out": true},
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply2 *engine.Account
	if err := sSApierRpc2.Call("ApierV2.GetAccount",
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

func TestSessionSv1ItAuth(t *testing.T) {
	args := &sessions.V1AuthorizeArgs{
		AuthorizeResources: true,
		ProcessThresholds:  true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime: time.Date(2018,
					time.January, 7, 16, 60, 0, 0, time.UTC),
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
					//Uuid:  "c9a2c620-5256-483a-a92d-c51e94bb7667",
					Value: 20,
					Directions: utils.StringMap{
						"*out": true},
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply *engine.Account
	if err := sSApierRpc2.Call("ApierV2.GetAccount",
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

func TestSessionSv1ItInitiateSession(t *testing.T) {
	initUsage := 5 * time.Minute
	args := &sessions.V1InitSessionArgs{
		InitSession:       true,
		AllocateResources: true,
		ProcessThresholds: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime: time.Date(2018,
					time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime: time.Date(2018,
					time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage: initUsage,
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
					//Uuid:  "c9a2c620-5256-483a-a92d-c51e94bb7667",
					Value: 29.898000,
					Directions: utils.StringMap{
						"*out": true},
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply *engine.Account
	if err := sSApierRpc2.Call("ApierV2.GetAccount",
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

func TestSessionSv1ItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession:  true,
		ReleaseResources:  true,
		ProcessThresholds: true,
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItTerminateSession",
			Event: map[string]interface{}{
				utils.OriginID:    "TestSSv1It1",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       10 * time.Minute,
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
					//Uuid:  "c9a2c620-5256-483a-a92d-c51e94bb7667",
					Value: 39.796000,
					Directions: utils.StringMap{
						"*out": true},
					Weight: 10,
				},
			},
		},
	}
	// Uuid will be generated
	// so we will compare ID from Account and Value from BalanceMap
	var reply2 *engine.Account
	if err := sSApierRpc2.Call("ApierV2.GetAccount",
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

func TestSessionSv1ItStopCgrEngine(t *testing.T) {
	if err := sSv1BiRpc2.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
