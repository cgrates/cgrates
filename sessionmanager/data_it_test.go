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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestSMGDataInitCfg(t *testing.T) {
	daCfgPath = path.Join(*dataDir, "conf", "samples", "smg")
	// Init config first
	var err error
	daCfg, err = config.NewCGRConfigFromFolder(daCfgPath)
	if err != nil {
		t.Error(err)
	}
	daCfg.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(daCfg)
}

// Remove data in both rating and accounting db
func TestSMGDataResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGDataResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGDataStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGDataApierRpcConn(t *testing.T) {
	var err error
	smgRPC, err = jsonrpc.Dial("tcp", daCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGDataTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst utils.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGDataLastUsedData(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 102400.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	tStart, _ := utils.ParseDate("2016-01-05T18:31:05Z")
	cd := engine.CallDescriptor{
		Direction:   "*out",
		Category:    "data",
		Tenant:      "cgrates.org",
		Subject:     "1001",
		Account:     "1001",
		Destination: utils.DATA,
		TimeStart:   tStart,
		TimeEnd:     tStart.Add(time.Duration(1024)),
	}
	var cc engine.CallCost
	// Make sure the cost is what we expect to be for 1MB of data
	if err := smgRPC.Call("Responder.GetCost", cd, &cc); err != nil {
		t.Error("Got error on Responder.GetCost: ", err.Error())
	} else if cc.Cost != 1024 {
		t.Errorf("Calling Responder.GetCost got callcost: %v", cc.Cost)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123491",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:59",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "5120", // 5MB
	}
	var maxUsage int64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 5120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 97280.0 // 100 -5
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123491",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:59",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "5120",
		utils.LastUsed:    "4096",
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 5120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 93184.0 // 100-9
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123491",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:59",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.LastUsed:    "0",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 98304.0 //100-4
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
}

func TestSMGDataLastUsedMultipleUpdates(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestSMGDataLastUsedMultipleData"}
	eAcntVal := 102400.0
	attrSetBalance := utils.AttrSetBalance{
		Tenant: acntAttrs.Tenant, Account: acntAttrs.Account,
		BalanceType: utils.DATA,
		BalanceID:   utils.StringPointer("TestSMGDataLastUsedMultipleData"),
		Value:       utils.Float64Pointer(eAcntVal)}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123492",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      acntAttrs.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:50",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "6144", // 6 MB
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 6144 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 96256 // 100-6
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	aSessions := make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		aSessions[0].Usage != time.Duration(6144) {
		t.Errorf("wrong active sessions: %f", aSessions[0].Usage.Seconds())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123492",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      acntAttrs.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:50",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "8192", // 8 MB
		utils.LastUsed:    "7168",
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 8192 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 87040.000000 // 15MB used
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		aSessions[0].Usage != time.Duration(15360) {
		t.Errorf("wrong active sessions: %f", aSessions[0].Usage)
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123492",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      acntAttrs.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:50",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1024", // 8 MB
		utils.LastUsed:    "5120", // 5 MB
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 1024 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 87040.000000 // the amount is not modified and there will be 1024 extra left in SMG
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		aSessions[0].Usage != time.Duration(13312) { // 14MB in used, 2MB extra reserved
		t.Errorf("wrong active sessions: %+v", aSessions[0].Usage)
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123492",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      acntAttrs.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:50",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1024",
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 1024 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 87040.000000
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		aSessions[0].Usage != time.Duration(14336) { // 14MB in use
		t.Errorf("wrong active sessions: %f", aSessions[0].Usage)
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123492",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      acntAttrs.Tenant,
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:50",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.LastUsed:    "0", // refund 1024 (extra used) + 1024 (extra reserved)
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 89088.000000
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRPC.Call("SMGenericV1.ProcessCDR", smgEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(time.Duration(20) * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		Accounts: []string{acntAttrs.Account}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "13312" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}
}

func TestSMGDataTTLExpired(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestSMGDataTTLExpired"}
	eAcntVal := 102400.0
	attrSetBalance := utils.AttrSetBalance{
		Tenant: acntAttrs.Tenant, Account: acntAttrs.Account,
		BalanceType: utils.DATA,
		BalanceID:   utils.StringPointer("TestSMGDataTTLExpired"),
		Value:       utils.Float64Pointer(eAcntVal)}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:      "TEST_EVENT",
		utils.TOR:             utils.DATA,
		utils.OriginID:        "TestSMGDataTTLExpired",
		utils.Direction:       utils.OUT,
		utils.Account:         acntAttrs.Account,
		utils.Subject:         acntAttrs.Account,
		utils.Destination:     utils.DATA,
		utils.Category:        "data",
		utils.Tenant:          "cgrates.org",
		utils.RequestType:     utils.META_PREPAID,
		utils.SetupTime:       "2016-01-05 18:30:52",
		utils.AnswerTime:      "2016-01-05 18:31:05",
		utils.Usage:           "1024",
		utils.SessionTTLUsage: "2048", // will be charged on TTL
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 1024 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 101376.000000
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	time.Sleep(70 * time.Millisecond)
	eAcntVal = 99328.000000
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
}

func TestSMGDataTTLExpMultiUpdates(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestSMGDataTTLExpMultiUpdates"}
	eAcntVal := 102400.0
	attrSetBalance := utils.AttrSetBalance{
		Tenant: acntAttrs.Tenant, Account: acntAttrs.Account,
		BalanceType: utils.DATA,
		BalanceID:   utils.StringPointer("TestSMGDataTTLExpMultiUpdates"),
		Value:       utils.Float64Pointer(eAcntVal)}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "4096", // 3MB
	}
	var maxUsage int64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 4096 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 98304.000000 //96MB
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	aSessions := make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		int64(aSessions[0].Usage) != 4096 {
		t.Errorf("wrong active sessions: %d", int64(aSessions[0].Usage))
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:         "TEST_EVENT",
		utils.TOR:                utils.DATA,
		utils.OriginID:           "123495",
		utils.Direction:          utils.OUT,
		utils.Account:            acntAttrs.Account,
		utils.Subject:            acntAttrs.Account,
		utils.Destination:        utils.DATA,
		utils.Category:           "data",
		utils.Tenant:             "cgrates.org",
		utils.RequestType:        utils.META_PREPAID,
		utils.SetupTime:          "2016-01-05 18:30:53",
		utils.AnswerTime:         "2016-01-05 18:31:05",
		utils.LastUsed:           "1024",
		utils.Usage:              "4096",
		utils.SessionTTLUsage:    "2048", // will be charged on TTL
		utils.SessionTTLLastUsed: "1024"} // will force last usage on timeout
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 4096 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 97280.000000 // 20480
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	time.Sleep(60 * time.Millisecond) // TTL will kick in
	eAcntVal = 98304.000000           // 1MB is returned
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
}

func TestSMGDataMultipleDataNoUsage(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestSMGDataTTLExpMultiUpdates"}
	eAcntVal := 102400.0
	attrSetBalance := utils.AttrSetBalance{
		Tenant: acntAttrs.Tenant, Account: acntAttrs.Account,
		BalanceType: utils.DATA,
		BalanceID:   utils.StringPointer("TestSMGDataTTLExpMultiUpdates"),
		Value:       utils.Float64Pointer(eAcntVal)}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2048",
	}
	var maxUsage int64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 2048 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 100352.000000 // 1054720
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	aSessions := make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		int64(aSessions[0].Usage) != 2048 {
		t.Errorf("wrong active sessions usage: %d", int64(aSessions[0].Usage))
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.SessionTTL:  "1h", // cancel timeout since usage 0 will not update it
		utils.Usage:       "1024",
		utils.LastUsed:    "1024",
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 1024 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 100352.000000 // 1054720
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	aSessions = make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		int64(aSessions[0].Usage) != 2048 {
		t.Errorf("wrong active sessions usage: %d", int64(aSessions[0].Usage))
	}

	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.SessionTTL:  "1h", // cancel timeout since usage 0 will not update it
		utils.Usage:       "0",
		utils.LastUsed:    "0",
	}
	if err := smgRPC.Call("SMGenericV2.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 0 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 100352.000000 // 1054720
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	aSessions = make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		int64(aSessions[0].Usage) != 1024 {
		t.Errorf("wrong active sessions usage: %d", int64(aSessions[0].Usage))
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.LastUsed:    "0",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 101376.000000 // refunded last 1MB reserved and unused
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.DATA].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.DATA].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
}

// TestSMGDataTTLUsageProtection makes sure that original TTL (50ms)
// limits the additional debit without overloading memory
func TestSMGDataTTLUsageProtection(t *testing.T) {
	var acnt *engine.Account
	acntAttrs := &utils.AttrGetAccount{Tenant: "cgrates.org",
		Account: "TestSMGDataTTLUsageProtection"}
	eAcntVal := 102400.0
	attrSetBalance := utils.AttrSetBalance{
		Tenant: acntAttrs.Tenant, Account: acntAttrs.Account,
		BalanceType: utils.DATA,
		BalanceID:   utils.StringPointer("TestSMGDataTTLUsageProtection"),
		Value:       utils.Float64Pointer(eAcntVal)}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if totalVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); totalVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", totalVal)
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.DATA,
		utils.OriginID:    "123495",
		utils.Direction:   utils.OUT,
		utils.Account:     acntAttrs.Account,
		utils.Subject:     acntAttrs.Account,
		utils.Destination: utils.DATA,
		utils.Category:    "data",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:53",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2048",
	}
	var maxUsage int64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 2048 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 100352.000000 // 1054720
	if err := smgRPC.Call("ApierV2.GetAccount", acntAttrs, &acnt); err != nil {
		t.Error(err)
	} else if dataVal := acnt.BalanceMap[utils.DATA].GetTotalValue(); dataVal != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, dataVal)
	}
	aSessions := make([]*ActiveSession, 0)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 ||
		int64(aSessions[0].Usage) != 2048 {
		t.Errorf("wrong active sessions usage: %d", int64(aSessions[0].Usage))
	}
	time.Sleep(60 * time.Millisecond)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
}
