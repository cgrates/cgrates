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
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var waitRater = flag.Int("wait_rater", 150, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

var daCfgPath string
var daCfg *config.CGRConfig
var smgRPC *rpc.Client

func TestSMGVoiceInitCfg(t *testing.T) {
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
func TestSMGVoiceResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGVoiceResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGVoiceStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGVoiceApierRpcConn(t *testing.T) {
	var err error
	smgRPC, err = jsonrpc.Dial("tcp", daCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGVoiceTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGVoiceMonetaryRefund(t *testing.T) {
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(90*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.700010
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123451",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1004",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 8.8
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceVoiceRefund(t *testing.T) {
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123452",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1003",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(90*time.Second) {
		t.Error("Received: ", maxUsage)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 120.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123452",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1003",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 150.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
}

func TestSMGVoiceMixedRefund(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	}
	//t.Logf("Initial monetary: %f", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	//t.Logf("Initial voice: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123453",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1002",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m30s",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(90*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	//var acnt *engine.Account
	//attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eVoiceVal := 90.0 * float64(time.Second)
	eMoneyVal := 8.7399
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected: %f, received: %f, expetced money: %f, recieved money : %f",
			eVoiceVal, acnt.BalanceMap[utils.VOICE].GetTotalValue(),
			eMoneyVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "123453",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1002",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "1m",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eVoiceVal = 90.0 * float64(time.Second)
	eMoneyVal = 8.79
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected voice: %f, received voice : %f, expected money: %f, received money: %f",
			eVoiceVal, acnt.BalanceMap[utils.VOICE].GetTotalValue(),
			eMoneyVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	//t.Logf("After monetary: %f", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	//t.Logf("After voice: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
}

func TestSMGVoiceLastUsed(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.790000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12350",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 7.39002
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12350",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "2m",
		utils.LastUsed:    "1m30s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 7.09005
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12350",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "2m",
		utils.LastUsed:    "2m30s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.590100
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12350",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "1m",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 7.59
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceLastUsedEnd(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 7.59000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234911",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234911",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "2m",
		utils.LastUsed:    "30s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.090030
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234911",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.LastUsed:    "0s",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 6.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceLastUsedNotFixed(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 6.59000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234922",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234922",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "2m",
		utils.LastUsed:    "13s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.123360
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "1234922",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1006",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.LastUsed:    "0s",
	}
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 5.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceSessionTTL(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 5.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12360",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1008",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	}
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(30 * time.Millisecond))
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	var aSessions []*ActiveSession
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT, utils.OriginID: "12360"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(120)*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 4.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12360",
		utils.Direction:   utils.OUT,
		utils.Account:     "1001",
		utils.Subject:     "1001",
		utils.Destination: "1008",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.Usage:       "2m",
		utils.LastUsed:    "30s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(10 * time.Millisecond))
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions", map[string]string{utils.RunID: utils.META_DEFAULT, utils.OriginID: "12360"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 4.090030
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	time.Sleep(100 * time.Millisecond)
	eAcntVal = 4.0565
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	time.Sleep(time.Duration(500 * time.Millisecond))
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{"1008"}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "2m30.05s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
		if cdrs[0].Cost != 1.5333 {
			t.Errorf("Unexpected CDR Cost received, cdr: %v %+v ", cdrs[0].Cost, cdrs[0])
		}
	}
}

func TestSMGVoiceSessionTTLWithRelocate(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account: "TestTTLWithRelocate", BalanceType: utils.VOICE,
		BalanceID:     utils.StringPointer("TestTTLWithRelocate"),
		Value:         utils.Float64Pointer(300 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero50ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant,
		Account: attrSetBalance.Account}
	eAcntVal := 300.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL_RELOCATE",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12361",
		utils.Direction:   utils.OUT,
		utils.Account:     "TestTTLWithRelocate",
		utils.Subject:     "TestTTLWithRelocate",
		utils.Destination: "1009",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	})
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	var aSessions []*ActiveSession
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: smgEv.GetStringIgnoreErrors(utils.OriginID)},
		&aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(120)*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 180.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv = engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:      smgEv[utils.EVENT_NAME],
		utils.ToR:             smgEv[utils.ToR],
		utils.InitialOriginID: smgEv[utils.OriginID],
		utils.OriginID:        "12362",
		utils.Direction:       smgEv[utils.Direction],
		utils.Account:         smgEv[utils.Account],
		utils.Subject:         smgEv[utils.Subject],
		utils.Destination:     smgEv[utils.Destination],
		utils.Category:        smgEv[utils.Category],
		utils.Tenant:          smgEv[utils.Tenant],
		utils.RequestType:     smgEv[utils.RequestType],
		utils.Usage:           "2m",
		utils.LastUsed:        "30s",
	})
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(20) * time.Millisecond)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: smgEv.GetStringIgnoreErrors(utils.OriginID)},
		&aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 150.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}

	time.Sleep(100 * time.Millisecond)
	eAcntVal = 149.95 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: smgEv.GetStringIgnoreErrors(utils.OriginID)},
		&aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	time.Sleep(5000 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT},
		DestinationPrefixes: []string{smgEv.GetStringIgnoreErrors(utils.Destination)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "2m30.05s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}

}

func TestSMGVoiceRelocateWithOriginIDPrefix(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org",
		Account:       "TestRelocateWithOriginIDPrefix",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestRelocateWithOriginIDPrefix"),
		Value:         utils.Float64Pointer(300 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero1s")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant,
		Account: attrSetBalance.Account}
	eAcntVal := 300.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT_RELOCATE_ORIGPREFIX",
		utils.ToR:         utils.VOICE,
		utils.OriginID:    "12371",
		utils.Direction:   utils.OUT,
		utils.Account:     attrSetBalance.Account,
		utils.Subject:     attrSetBalance.Account,
		utils.Destination: "12371",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType: utils.META_PREPAID,
		utils.SetupTime:   "2016-01-05 18:30:49",
		utils.AnswerTime:  "2016-01-05 18:31:05",
		utils.Usage:       "2m",
	})
	var maxUsage time.Duration
	if err := smgRPC.Call(utils.SMGenericV2InitiateSession, smgEv,
		&maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(20) * time.Millisecond)
	var aSessions []*ActiveSession
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: smgEv.GetStringIgnoreErrors(utils.OriginID)},
		&aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(120)*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 180.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv = map[string]interface{}{
		utils.EVENT_NAME:      smgEv[utils.EVENT_NAME],
		utils.ToR:             smgEv[utils.ToR],
		utils.InitialOriginID: smgEv[utils.OriginID],
		utils.OriginID:        "12372-1",
		utils.Direction:       smgEv[utils.Direction],
		utils.Account:         smgEv[utils.Account],
		utils.Subject:         smgEv[utils.Subject],
		utils.Destination:     smgEv[utils.Destination],
		utils.Category:        smgEv[utils.Category],
		utils.Tenant:          smgEv[utils.Tenant],
		utils.RequestType:     smgEv[utils.RequestType],
		utils.Usage:           "2m",
		utils.LastUsed:        "30s",
	}
	if err := smgRPC.Call(utils.SMGenericV2UpdateSession,
		smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != time.Duration(120*time.Second) {
		t.Error("Bad max usage: ", maxUsage)
	}
	time.Sleep(time.Duration(20) * time.Millisecond)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: "12372-1"}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 150.0 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv = engine.NewMapEvent(map[string]interface{}{
		utils.EVENT_NAME:     smgEv[utils.EVENT_NAME],
		utils.ToR:            smgEv[utils.ToR],
		utils.OriginIDPrefix: "12372",
		utils.Direction:      smgEv[utils.Direction],
		utils.Account:        smgEv[utils.Account],
		utils.Subject:        smgEv[utils.Subject],
		utils.Destination:    smgEv[utils.Destination],
		utils.Category:       smgEv[utils.Category],
		utils.Tenant:         smgEv[utils.Tenant],
		utils.RequestType:    smgEv[utils.RequestType],
		utils.Usage:          "1m", // Total session usage
	})
	var rpl string
	if err := smgRPC.Call("SMGenericV1.TerminateSession",
		smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	if err := smgRPC.Call("SMGenericV1.GetActiveSessions",
		map[string]string{utils.RunID: utils.META_DEFAULT,
			utils.OriginID: "12372-1"}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	eAcntVal = 240 * float64(time.Second)
	if err := smgRPC.Call("ApierV2.GetAccount",
		attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
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
		if cdrs[0].Usage != "1m0s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}
}

/*
func TestSMGDataDerivedChargingNoCredit(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1011"}
	eAcntVal := 50000.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.OriginID:       "1234967",
		utils.Direction:   utils.OUT,
		utils.Account:     "1011",
		utils.Subject:     "1011",
		utils.Destination: "+49",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType:     utils.META_PREPAID,
		utils.SetupTime:  "2016-01-05 18:30:49",
		utils.AnswerTime: "2016-01-05 18:31:05",
		utils.Usage:       "100",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV2.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	// the second derived charging run has no credit

	if maxUsage != 0 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 50000.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
}
*/

// ToDo: Add test for ChargeEvent with derived charging, one with debit possible and second not so we see refund and error.CreditInsufficient showing up.

func TestSMGVoiceSessionStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
