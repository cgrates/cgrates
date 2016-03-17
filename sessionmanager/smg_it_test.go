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

func TestSMGInitCfg(t *testing.T) {
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
func TestSMGResetDataDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitDataDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func TestSMGResetStorDb(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.InitStorDb(daCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func TestSMGStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if _, err := engine.StopStartEngine(daCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestSMGApierRpcConn(t *testing.T) {
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
func TestSMGTPFromFolder(t *testing.T) {
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

func TestSMGMonetaryRefund(t *testing.T) {
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12345",
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
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
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
		utils.ACCID:       "12345",
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
	if err = smgRPC.Call("SMGenericV1.SessionEnd", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 8.8
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}

func TestSMGVoiceRefund(t *testing.T) {
	if !*testIntegration {
		return
	}
	smgEv := SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12345",
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
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 90 {
		t.Error("Bad max usage: ", maxUsage)
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
		utils.ACCID:       "12345",
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
	if err = smgRPC.Call("SMGenericV1.SessionEnd", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 150.0
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.VOICE].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.VOICE].GetTotalValue())
	}
}

func TestSMGMixedRefund(t *testing.T) {
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
		utils.ACCID:       "12345",
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
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
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
		utils.ACCID:       "12345",
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
	if err = smgRPC.Call("SMGenericV1.SessionEnd", smgEv, &rpl); err != nil || rpl != utils.OK {
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

func TestSMGLastUsed(t *testing.T) {
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
		utils.ACCID:       "12349",
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
	if err := smgRPC.Call("SMGenericV1.SessionStart", smgEv, &maxUsage); err != nil {
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
		utils.ACCID:       "12349",
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
	if err := smgRPC.Call("SMGenericV1.SessionUpdate", smgEv, &maxUsage); err != nil {
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
		utils.ACCID:       "12349",
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
	if err := smgRPC.Call("SMGenericV1.SessionUpdate", smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	if maxUsage != 120 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 6.5901
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
	smgEv = SMGenericEvent{
		utils.EVENT_NAME:  "TEST_EVENT",
		utils.TOR:         utils.VOICE,
		utils.ACCID:       "12349",
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
	if err = smgRPC.Call("SMGenericV1.SessionEnd", smgEv, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}
	eAcntVal = 7.59
	if err := smgRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MONETARY].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MONETARY].GetTotalValue())
	}
}
