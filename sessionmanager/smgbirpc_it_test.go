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

package sessionmanager

import (
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cenk/rpc2"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgBiRPCCfgPath  string
	smgBiRPCCfg      *config.CGRConfig
	smgBiRPC         *rpc2.Client
	disconnectEvChan = make(chan *utils.AttrDisconnectSession)
)

func handleDisconnectSession(clnt *rpc2.Client, args *utils.AttrDisconnectSession, reply *string) error {
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
	clntHandlers := map[string]interface{}{"SMGClientV1.DisconnectSession": handleDisconnectSession}
	if _, err = utils.NewBiJSONrpcClient(smgBiRPCCfg.SmGenericConfig.ListenBijson, clntHandlers); err != nil { // First attempt is to make sure multiple clients are supported
		t.Fatal(err)
	}
	if smgBiRPC, err = utils.NewBiJSONrpcClient(smgBiRPCCfg.SmGenericConfig.ListenBijson, clntHandlers); err != nil {
		t.Fatal(err)
	}
	if smgRPC, err = jsonrpc.Dial("tcp", smgBiRPCCfg.RPCJSONListen); err != nil { // Connect also simple RPC so we can check accounts and such
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGBiRPCTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGBiRPCSessionAutomaticDisconnects(t *testing.T) {
	// Create a balance with 1 second inside and rating increments of 1ms (to be compatible with debit interval)
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org", Account: "TestSMGBiRPCSessionAutomaticDisconnects", BalanceType: utils.VOICE, BalanceID: utils.StringPointer("TestSMGBiRPCSessionAutomaticDisconnects"),
		Value: utils.Float64Pointer(0.01), RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant, Account: attrSetBalance.Account}
	eAcntVal := 0.01
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123451",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     attrSetBalance.Account,
		utils.SUBJECT:     attrSetBalance.Account,
		utils.DESTINATION: "1004",
		utils.CATEGORY:    "call",
		utils.TENANT:      attrSetBalance.Tenant,
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
	}
	var maxUsage float64
	if err := smgBiRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != -1 {
		t.Error("Bad max usage: ", maxUsage)
	}
	// Make sure we are receiving a disconnect event
	select {
	case <-time.After(time.Duration(50 * time.Millisecond)):
		t.Error("Did not receive disconnect event")
	case disconnectEv := <-disconnectEvChan:
		if SMGenericEvent(disconnectEv.EventStart).GetOriginID(utils.META_DEFAULT) != smgEv[utils.ACCID] {
			t.Errorf("Unexpected event received: %+v", disconnectEv)
		}
		smgEv[utils.USAGE] = disconnectEv.EventStart[utils.USAGE]
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
	time.Sleep(time.Duration(10) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{smgEv.GetDestination(utils.META_DEFAULT)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "0.01" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		} else if cdrs[0].CostSource != utils.SESSION_MANAGER_SOURCE {
			t.Errorf("Unexpected CDR CostSource received, cdr: %v %+v ", cdrs[0].CostSource, cdrs[0])
		}
	}
}

func TestSMGBiRPCSessionOriginatorTerminate(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org", Account: "TestSMGBiRPCSessionOriginatorTerminate", BalanceType: utils.VOICE, BalanceID: utils.StringPointer("TestSMGBiRPCSessionOriginatorTerminate"),
		Value: utils.Float64Pointer(1), RatingSubject: utils.StringPointer("*zero1ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrGetAcnt := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant, Account: attrSetBalance.Account}
	eAcntVal := 1.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123452",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     attrSetBalance.Account,
		utils.SUBJECT:     attrSetBalance.Account,
		utils.DESTINATION: "1005",
		utils.CATEGORY:    "call",
		utils.TENANT:      attrSetBalance.Tenant,
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
	}
	var maxUsage float64
	if err := smgBiRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != -1 {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(10 * time.Millisecond)) // Give time for  debits to occur
	smgEv[utils.USAGE] = "7ms"
	var rpl string
	if err = smgBiRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(time.Duration(50 * time.Millisecond)) // Give time for  debits to occur
	if err := smgRPC.Call("ApierV2.GetAccount", attrGetAcnt, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() > 0.995 { // FixMe: should be not 0.93?
		t.Errorf("Balance value: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.ProcessCDR", smgEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{smgEv.GetDestination(utils.META_DEFAULT)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "0.007" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		} else if cdrs[0].CostSource != utils.SESSION_MANAGER_SOURCE {
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
