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
	smgBiRPCCfgPath  string
	smgBiRPCCfg      *config.CGRConfig
	smgBiRPC         *rpc2.Client
	disconnectEvChan = make(chan *utils.AttrDisconnectSession)
	err              error
)

func handleDisconnectSession(clnt *rpc2.Client,
	args *utils.AttrDisconnectSession, reply *string) error {
	disconnectEvChan <- args
	*reply = utils.OK
	return nil
}

func TestSMGBiRPCInitCfg(t *testing.T) {
	smgBiRPCCfgPath = path.Join(*dataDir, "conf", "samples", "smg_automatic_debits")
	// Init config first
	smgBiRPCCfg, err = config.NewCGRConfigFromFolder(smgBiRPCCfgPath)
	if err != nil {
		t.Error(err)
	}
	smgBiRPCCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(smgBiRPCCfg)
}

// Remove data in both rating and accounting db
func TestSMGBiRPCResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(smgBiRPCCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGBiRPCResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(smgBiRPCCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGBiRPCStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgBiRPCCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGBiRPCApierRpcConn(t *testing.T) {
	clntHandlers := map[string]interface{}{"SessionSv1.DisconnectSession": handleDisconnectSession}
	dummyClnt, err := utils.NewBiJSONrpcClient(smgBiRPCCfg.SessionSCfg().ListenBijson,
		clntHandlers)
	if err != nil { // First attempt is to make sure multiple clients are supported
		t.Fatal(err)
	}
	if smgBiRPC, err = utils.NewBiJSONrpcClient(smgBiRPCCfg.SessionSCfg().ListenBijson,
		clntHandlers); err != nil {
		t.Fatal(err)
	}
	if smgRPC, err = jsonrpc.Dial("tcp", smgBiRPCCfg.ListenCfg().RPCJSONListen); err != nil { // Connect also simple RPC so we can check accounts and such
		t.Fatal(err)
	}
	dummyClnt.Close() // close so we don't get EOF error when disconnecting server
}

// Load the tariff plan, creating accounts and their balances
func TestSMGBiRPCTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGBiRPCSessionAutomaticDisconnects(t *testing.T) {
	// Create a balance with 1 second inside and rating increments of 1ms (to be compatible with debit interval)
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:       "TestSMGBiRPCSessionAutomaticDisconnects",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestSMGBiRPCSessionAutomaticDisconnects"),
		Value:         utils.Float64Pointer(0.01 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant,
		Account: attrSetBalance.Account}
	eAcntVal := 0.01 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Direction:   utils.OUT,
		utils.Account:     attrSetBalance.Account,
		utils.Subject:     attrSetBalance.Account,
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      attrSetBalance.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       time.Duration(200 * time.Millisecond),
	})
	var maxUsage float64
	if err := smgBiRPC.Call(utils.SMGenericV1InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 0.01 {
		t.Error("Bad max usage: ", maxUsage)
	}
	// Make sure we are receiving a disconnect event
	select {
	case <-time.After(time.Duration(50 * time.Millisecond)):
		t.Error("Did not receive disconnect event")
	case disconnectEv := <-disconnectEvChan:
		if engine.NewMapEvent(disconnectEv.EventStart).GetStringIgnoreErrors(utils.OriginID) != smgEv[utils.OriginID] {
			t.Errorf("Unexpected event received: %+v", disconnectEv)
		}
		smgEv[utils.Usage] = disconnectEv.EventStart[utils.Usage]
	}
	var rpl string
	if err = smgBiRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(time.Duration(100 * time.Millisecond)) // Give time for  debits to occur
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != 0 {
		t.Errorf("Balance should be empty, have: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.ProcessCDR", smgEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		DestinationPrefixes: []string{smgEv.GetStringIgnoreErrors(utils.Destination)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
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

func TestSMGBiRPCSessionOriginatorTerminate(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:       "TestSMGBiRPCSessionOriginatorTerminate",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestSMGBiRPCSessionOriginatorTerminate"),
		Value:         utils.Float64Pointer(1 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant, Account: attrSetBalance.Account}
	eAcntVal := 1.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123452",
		utils.Direction:   utils.OUT,
		utils.Account:     attrSetBalance.Account,
		utils.Subject:     attrSetBalance.Account,
		utils.Destination: "1005",
		utils.Category:    "call",
		utils.Tenant:      attrSetBalance.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       time.Duration(200 * time.Millisecond),
	})
	var maxUsage float64
	if err := smgBiRPC.Call(utils.SMGenericV1InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 0.2 {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(10 * time.Millisecond)) // Give time for  debits to occur
	smgEv[utils.Usage] = "7ms"
	var rpl string
	if err = smgBiRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(time.Duration(50 * time.Millisecond)) // Give time for  debits to occur
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() > 0.995*float64(time.Second) { // FixMe: should be not 0.93?
		t.Errorf("Balance value: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.ProcessCDR", smgEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		DestinationPrefixes: []string{smgEv.GetStringIgnoreErrors(utils.Destination)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
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

func TestSMGBiRPCStopCgrEngine(t *testing.T) {
	if err := smgBiRPC.Close(); err != nil { // Close the connection so we don't get EOF warnings from client
		t.Error(err)
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
