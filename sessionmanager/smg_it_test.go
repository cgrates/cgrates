/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be u297seful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package sessionmanager

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

var testIntegration = flag.Bool("integration", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args
var waitRater = flag.Int("wait_rater", 150, "Number of miliseconds to wait for rater to start and cache")
var dataDir = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")

var daCfgPath string
var daCfg *config.CGRConfig
var smgRPC *rpc.Client
var err error

func TestSMGVoiceInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
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
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGVoiceResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGVoiceStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGVoiceApierRpcConn(t *testing.T) {
	if !*testIntegration {
		return
	}
	var err error
	smgRPC, err = jsonrpc.Dial("tcp", daCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func TestSMGVoiceTPFromFolder(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	var loadInst engine.LoadInstance
	if err := smgRPC.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func TestSMGVoiceMonetaryRefund(t *testing.T) {
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123451",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1004",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m30s",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 90 {
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
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123451",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1004",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
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
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123452",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1003",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m30s",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 90 {
		t.Error("Received: ", maxUsage)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 120.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123452",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1003",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 150.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
}

func TestSMGVoiceMixedRefund(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	}
	t.Logf("Initial monetary: %f", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	t.Logf("Initial voice: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123453",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1002",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m30s",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 90 {
		t.Error("Bad max usage: ", maxUsage)
	}
	//var acnt *engine.Account
	//attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eVoiceVal := 90.0
	eMoneyVal := 8.7399
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected: %f, received: %f, expetced money: %f, recieved money : %f", eVoiceVal, acnt.BalanceMap[utils.VOICE].GetTotalValue(), eMoneyVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "123453",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1002",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "1m",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eVoiceVal = 90.0
	eMoneyVal = 8.79
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected voice: %f, received voice : %f, expected money: %f, received money: %f", eVoiceVal, acnt.BalanceMap[utils.VOICE].GetTotalValue(), eMoneyVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	t.Logf("After monetary: %f", acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	t.Logf("After voice: %f", acnt.BalanceMap[utils.VOICE].GetTotalValue())
}

func TestSMGVoiceLastUsed(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.790000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12350",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "2m",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 7.39002
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12350",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "2m",
		utils.LastUsed:    "1m30s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 7.09005
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12350",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "2m",
		utils.LastUsed:    "2m30s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.590100
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12350",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "1m",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 7.59
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceLastUsedEnd(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 7.59000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234911",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "2m",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234911",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "2m",
		utils.LastUsed:    "30s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.090030
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234911",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.LastUsed:    "0s",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 6.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceLastUsedNotFixed(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 6.59000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234922",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "2m",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.190020
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234922",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "2m",
		utils.LastUsed:    "13s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 5.123360
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "1234922",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1006",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.LastUsed:    "0s",
	}
	var rpl string
	if err = smgRPC.Call("SMGenericV1.TerminateSession", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 5.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceSessionTTL(t *testing.T) {
	if !*testIntegration {
		return
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 5.590000
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12360",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1008",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "2m",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	var aSessions []*ActiveSession
	if err := smgRPC.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{RunID: utils.StringPointer(utils.META_DEFAULT), OriginID: utils.StringPointer("12360")}, &aSessions); err != nil {
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
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12360",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "1001",
		utils.SUBJECT:     "1001",
		utils.DESTINATION: "1008",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.USAGE:       "2m",
		utils.LastUsed:    "30s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	if err := smgRPC.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{RunID: utils.StringPointer(utils.META_DEFAULT), OriginID: utils.StringPointer("12360")}, &aSessions); err != nil {
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
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{"1008"}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "150.05" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
		if cdrs[0].Cost != 1.5333 {
			t.Errorf("Unexpected CDR Cost received, cdr: %v %+v ", cdrs[0].Cost, cdrs[0])
		}
	}
}

func TestSMGVoiceSessionTTLWithRelocate(t *testing.T) {
	if !*testIntegration {
		return
	}
	attrSetBalance := utils.AttrSetBalance{Tenant: "cgrates.org", Account: "TestTTLWithRelocate", BalanceType: utils.VOICE, BalanceID: utils.StringPointer("TestTTLWithRelocate"),
		Value: utils.Float64Pointer(300), RatingSubject: utils.StringPointer("*zero50ms")}
	var reply string
	if err := smgRPC.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: attrSetBalance.Tenant, Account: attrSetBalance.Account}
	eAcntVal := 300.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT_SESSION_TTL_RELOCATE",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12361",
		utils.DIRECTION:   utils.OUT,
		utils.ACCOUNT:     "TestTTLWithRelocate",
		utils.SUBJECT:     "TestTTLWithRelocate",
		utils.DESTINATION: "1009",
		utils.CATEGORY:    "call",
		utils.TENANT:      "cgrates.org",
		utils.REQTYPE:     utils.META_PREPAID,
		utils.SETUP_TIME:  "2016-01-05 18:30:49",
		utils.ANSWER_TIME: "2016-01-05 18:31:05",
		utils.USAGE:       "2m",
	}
	var maxUsage float64
	if err := smgRPC.Call("SMGenericV1.InitiateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 180.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	var aSessions []*ActiveSession
	if err := smgRPC.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{RunID: utils.StringPointer(utils.META_DEFAULT), OriginID: utils.StringPointer(smgEv.GetUUID())}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(120)*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:      smgEv[utils.EVENT_NAME],
		utils.TOR:             smgEv[utils.TOR],
		utils.InitialOriginID: smgEv[utils.ACCID],
		utils.ACCID:           "12362",
		utils.DIRECTION:       smgEv[utils.DIRECTION],
		utils.ACCOUNT:         smgEv[utils.ACCOUNT],
		utils.SUBJECT:         smgEv[utils.SUBJECT],
		utils.DESTINATION:     smgEv[utils.DESTINATION],
		utils.CATEGORY:        smgEv[utils.CATEGORY],
		utils.TENANT:          smgEv[utils.TENANT],
		utils.REQTYPE:         smgEv[utils.REQTYPE],
		utils.USAGE:           "2m",
		utils.LastUsed:        "30s",
	}
	if err := smgRPC.Call("SMGenericV1.UpdateSession", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 150.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{RunID: utils.StringPointer(utils.META_DEFAULT), OriginID: utils.StringPointer(smgEv.GetUUID())}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != time.Duration(150)*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	time.Sleep(100 * time.Millisecond)
	eAcntVal = 149.95
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
	if err := smgRPC.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{RunID: utils.StringPointer(utils.META_DEFAULT), OriginID: utils.StringPointer(smgEv.GetUUID())}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	}
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.META_DEFAULT}, DestinationPrefixes: []string{smgEv.GetDestination(utils.META_DEFAULT)}}
	if err := smgRPC.Call("ApierV2.GetCdrs", req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "150.05" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}

}
