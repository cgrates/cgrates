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

package sessions

import (
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	sessionsBiRPCCfgPath string
	sessionsBiRPCCfg     *config.CGRConfig
	sessionsBiRPC        *rpc2.Client
	disconnectEvChan     = make(chan *utils.AttrDisconnectSession, 1)
	err                  error
)

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func TestSessionsBiRPCInitCfg(t *testing.T) {
	sessionsBiRPCCfgPath = path.Join(*dataDir, "conf", "samples", "smg_automatic_debits")
	// Init config first
	sessionsBiRPCCfg, err = config.NewCGRConfigFromPath(sessionsBiRPCCfgPath)
	if err != nil {
		t.Error(err)
	}
	sessionsBiRPCCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(sessionsBiRPCCfg)
}

// Remove data in both rating and accounting db
func TestSessionsBiRPCResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(sessionsBiRPCCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSessionsBiRPCResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sessionsBiRPCCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSessionsBiRPCStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sessionsBiRPCCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSessionsBiRPCApierRpcConn(t *testing.T) {
	clntHandlers := map[string]interface{}{utils.SessionSv1DisconnectSession: handleDisconnectSession}
	dummyClnt, err := utils.NewBiJSONrpcClient(sessionsBiRPCCfg.SessionSCfg().ListenBijson,
		clntHandlers)
	if err != nil { // First attempt is to make sure multiple clients are supported
		t.Fatal(err)
	}
	if sessionsBiRPC, err = utils.NewBiJSONrpcClient(sessionsBiRPCCfg.SessionSCfg().ListenBijson,
		clntHandlers); err != nil {
		t.Fatal(err)
	}
	if sessionsRPC, err = jsonrpc.Dial("tcp", sessionsBiRPCCfg.ListenCfg().RPCJSONListen); err != nil { // Connect also simple RPC so we can check accounts and such
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

// Load the tariff plan, creating accounts and their balances
func TestSessionsBiRPCTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := sessionsRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSessionsBiRPCSessionAutomaticDisconnects(t *testing.T) {
	// Create a balance with 1 second inside and rating increments of 1ms (to be compatible with debit interval)
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:       "TestSessionsBiRPCSessionAutomaticDisconnects",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestSessionsBiRPCSessionAutomaticDisconnects"),
		Value:         utils.Float64Pointer(0.01 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := sessionsRPC.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant,
		Account: attrSetBalance.Account}
	eAcntVal := 0.01 * float64(time.Second)
	if err := sessionsRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}

	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsBiRPCSessionAutomaticDisconnects",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "123451",
				utils.Account:     attrSetBalance.Account,
				utils.Subject:     attrSetBalance.Account,
				utils.Destination: "1004",
				utils.Category:    "call",
				utils.Tenant:      attrSetBalance.Tenant,
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:       time.Duration(200 * time.Millisecond),
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsBiRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond) // give some time to allow the session to be created
	expMaxUsage := 3 * time.Hour      // MaxCallDuration from config
	if *initRpl.MaxUsage != expMaxUsage {
		t.Errorf("Expecting : %+v, received: %+v", expMaxUsage, *initRpl.MaxUsage)
	}
	// Make sure we are receiving a disconnect event
	select {
	case <-time.After(time.Duration(100 * time.Millisecond)):
		t.Error("Did not receive disconnect event")
	case disconnectEv := <-disconnectEvChan:
		if engine.NewMapEvent(disconnectEv.EventStart).GetStringIgnoreErrors(utils.OriginID) != initArgs.CGREvent.Event[utils.OriginID] {
			t.Errorf("Unexpected event received: %+v", disconnectEv)
		}
		initArgs.CGREvent.Event[utils.Usage] = disconnectEv.EventStart[utils.Usage]
	}
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsDataLastUsedData",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "123451",
				utils.Direction:   utils.META_OUT,
				utils.Account:     attrSetBalance.Account,
				utils.Subject:     attrSetBalance.Account,
				utils.Destination: "1004",
				utils.Category:    "call",
				utils.Tenant:      attrSetBalance.Tenant,
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:       initArgs.CGREvent.Event[utils.Usage],
			},
		},
	}

	var rpl string
	if err := sessionsBiRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(time.Duration(100 * time.Millisecond)) // Give time for  debits to occur
	if err := sessionsRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != 0 {
		t.Errorf("Balance should be empty, have: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := sessionsBiRPC.Call(utils.SessionSv1ProcessCDR, termArgs.CGREvent, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		DestinationPrefixes: []string{"1004"}}
	if err := sessionsRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "10ms" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		} else if cdrs[0].CostSource != utils.MetaSessionS {
			t.Errorf("Unexpected CDR CostSource received, cdr: %v %+v ", cdrs[0].CostSource, cdrs[0])
		}
	}
}

func TestSessionsBiRPCSessionOriginatorTerminate(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:       "TestSessionsBiRPCSessionOriginatorTerminate",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestSessionsBiRPCSessionOriginatorTerminate"),
		Value:         utils.Float64Pointer(1 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := sessionsRPC.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant, Account: attrSetBalance.Account}
	eAcntVal := 1.0 * float64(time.Second)
	if err := sessionsRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}

	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsBiRPCSessionOriginatorTerminate",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "123452",
				utils.Account:     attrSetBalance.Account,
				utils.Subject:     attrSetBalance.Account,
				utils.Destination: "1005",
				utils.Category:    "call",
				utils.Tenant:      attrSetBalance.Tenant,
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:       time.Duration(200 * time.Millisecond),
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsBiRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}

	expMaxUsage := 3 * time.Hour // MaxCallDuration from config
	if *initRpl.MaxUsage != expMaxUsage {
		t.Errorf("Expecting : %+v, received: %+v", expMaxUsage, *initRpl.MaxUsage)
	}

	time.Sleep(time.Duration(10 * time.Millisecond)) // Give time for  debits to occur

	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsBiRPCSessionOriginatorTerminate",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "123452",
				utils.Account:     attrSetBalance.Account,
				utils.Subject:     attrSetBalance.Account,
				utils.Destination: "1005",
				utils.Category:    "call",
				utils.Tenant:      attrSetBalance.Tenant,
				utils.RequestType: utils.META_PREPAID,
				utils.SetupTime:   time.Date(2016, time.January, 5, 18, 30, 59, 0, time.UTC),
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:       time.Duration(7 * time.Millisecond),
			},
		},
	}

	var rpl string
	if err := sessionsBiRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	time.Sleep(time.Duration(50 * time.Millisecond)) // Give time for  debits to occur
	if err := sessionsRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() > 0.995*float64(time.Second) { // FixMe: should be not 0.93?
		t.Errorf("Balance value: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}

	if err := sessionsRPC.Call(utils.SessionSv1ProcessCDR, termArgs.CGREvent, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(time.Duration(10) * time.Millisecond)

	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		DestinationPrefixes: []string{"1005"}}
	if err := sessionsRPC.Call(utils.ApierV2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "7ms" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		} else if cdrs[0].CostSource != utils.MetaSessionS {
			t.Errorf("Unexpected CDR CostSource received, cdr: %v %+v ", cdrs[0].CostSource, cdrs[0])
		}
	}
}

func TestSessionsBiRPCStopCgrEngine(t *testing.T) {
	if err := sessionsBiRPC.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
