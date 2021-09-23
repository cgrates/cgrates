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
package sessions

import (
	"fmt"
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	voiceCfgPath string
	voiceCfgDIR  string
	voiceCfg     *config.CGRConfig
	sessionsRPC  *rpc.Client

	sessionsVoiceTests = []func(t *testing.T){
		testSessionsVoiceInitCfg,
		testSessionsVoiceResetDataDb,
		testSessionsVoiceResetStorDb,
		testSessionsVoiceStartEngine,
		testSessionsVoiceApierRpcConn,
		testSessionsVoiceTPFromFolder,
		testSessionsVoiceMonetaryRefund,
		testSessionsVoiceVoiceRefund,
		testSessionsVoiceMixedRefund,
		testSessionsVoiceLastUsed,
		testSessionsVoiceLastUsedEnd,
		testSessionsVoiceLastUsedNotFixed,
		testSessionsVoiceSessionTTL,
		testSessionsVoiceSessionTTLWithRelocate,
		testSessionsVoiceRelocateWithOriginIDPrefix,
		testSessionsVoiceStopCgrEngine,
	}
)

func TestSessionsVoice(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		voiceCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		voiceCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		voiceCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sessionsVoiceTests {
		t.Run(voiceCfgDIR, stest)
	}
}

func testSessionsVoiceInitCfg(t *testing.T) {
	voiceCfgPath = path.Join(*dataDir, "conf", "samples", voiceCfgDIR)
	// Init config first
	var err error
	voiceCfg, err = config.NewCGRConfigFromPath(voiceCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessionsVoiceResetDataDb(t *testing.T) {
	if err := engine.InitDataDb(voiceCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSessionsVoiceResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(voiceCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionsVoiceStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(voiceCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionsVoiceApierRpcConn(t *testing.T) {
	var err error
	sessionsRPC, err = newRPCClient(voiceCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSessionsVoiceTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := sessionsRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionsVoiceMonetaryRefund(t *testing.T) {
	usage := time.Minute + 30*time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceMonetaryRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123451",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.700010
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage = time.Minute
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceMonetaryRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123451",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVal = 8.8
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsVoiceVoiceRefund(t *testing.T) {
	usage := time.Minute + 30*time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceVoiceRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123452",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 120.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage = time.Minute
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceVoiceRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123452",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVal = 150.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
}

func testSessionsVoiceMixedRefund(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	}

	usage := time.Minute + 30*time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceMixedRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123453",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	//var acnt *engine.Account
	//attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eVoiceVal := 90.0 * float64(time.Second)
	eMoneyVal := 8.7399
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected: %f, received: %f, expetced money: %f, recieved money : %f",
			eVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue(),
			eMoneyVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage = time.Minute
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceMixedRefund",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "123453",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eVoiceVal = 90.0 * float64(time.Second)
	eMoneyVal = 8.79
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eVoiceVal ||
		acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eMoneyVal {
		t.Errorf("Expected voice: %f, received voice : %f, expected money: %f, received money: %f",
			eVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue(),
			eMoneyVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
	//t.Logf("After monetary: %f", acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	//t.Logf("After voice: %f", acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
}

func testSessionsVoiceLastUsed(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 8.790000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVal = 7.39002
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	lastUsage := time.Minute + 30*time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EventName:    "Update1",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVal = 7.09005
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	lastUsage = 2*time.Minute + 30*time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVal = 6.590100
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage = time.Minute
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVal = 7.59
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsVoiceLastUsedEnd(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 7.59000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedEnd",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234911",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVal = 6.190020
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	lastUsage := 30 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedEnd",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234911",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVal = 6.090030
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedEnd",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234911",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.LastUsed:     "0s",
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVal = 6.590000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsVoiceLastUsedNotFixed(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 6.59000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedNotFixed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234922",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVal = 5.190020
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	//Update
	lastUsage := 12 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedNotFixed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234922",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVal = 5.123360
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedNotFixed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "1234922",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1006",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.LastUsed:     "0s",
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVal = 5.590000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsVoiceSessionTTL(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 5.590000
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceSessionTTL",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT_SESSION_TTL",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12360",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1008",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}

	time.Sleep(30 * time.Millisecond)
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	var aSessions []*ExternalSession
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12372-1"),
				"*string:~*req.no_field:10",
			},
		}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12360"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 120*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}

	eAcntVal = 4.190020
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	//Update
	lastUsage := 30 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceSessionTTL",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT_SESSION_TTL",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12360",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1008",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12360"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}

	eAcntVal = 4.090030
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
	time.Sleep(200 * time.Millisecond)
	eAcntVal = 4.0566 // rounding issue; old values : 4.0565 , 4.0566
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}

	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, DestinationPrefixes: []string{"1008"}}
	if err := sessionsRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "2m30.05s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
		if cdrs[0].Cost != 1.5334 {
			t.Errorf("Unexpected CDR Cost received, cdr: %v %+v ", cdrs[0].Cost, utils.ToJSON(cdrs[0]))
		}
	}
}

func testSessionsVoiceSessionTTLWithRelocate(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "TestTTLWithRelocate",
		BalanceType: utils.MetaVoice,
		Value:       300 * float64(time.Second),
		Balance: map[string]interface{}{
			utils.ID:            "TestTTLWithRelocate",
			utils.RatingSubject: "*zero50ms",
		},
	}
	var reply string
	if err := sessionsRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  attrSetBalance.Tenant,
		Account: attrSetBalance.Account,
	}
	eAcntVal := 300.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceSessionTTLWithRelocate",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT_SESSION_TTL_RELOCATE",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12361",
				utils.AccountField: "TestTTLWithRelocate",
				utils.Subject:      "TestTTLWithRelocate",
				utils.Destination:  "1009",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	time.Sleep(20 * time.Millisecond)
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	var aSessions []*ExternalSession
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12361"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 120*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 180.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	//Update
	lastUsage := 30 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceSessionTTLWithRelocate",
			Event: map[string]interface{}{
				utils.EventName:       "TEST_EVENT_SESSION_TTL_RELOCATE",
				utils.ToR:             utils.MetaVoice,
				utils.InitialOriginID: "12361", //take the initial originID from init
				utils.OriginID:        "12362",
				utils.AccountField:    "TestTTLWithRelocate",
				utils.Subject:         "TestTTLWithRelocate",
				utils.Destination:     "1009",
				utils.Category:        "call",
				utils.Tenant:          "cgrates.org",
				utils.RequestType:     utils.MetaPrepaid,
				utils.SetupTime:       time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:      time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:           usage,
				utils.LastUsed:        lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	time.Sleep(20 * time.Millisecond)
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12362"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 150.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	time.Sleep(200 * time.Millisecond) // should trigger the TTL from config
	eAcntVal = 149.95 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12362"),
			},
		}, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, utils.ToJSON(aSessions))
	}
	time.Sleep(100 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault}, DestinationPrefixes: []string{"1009"}}
	if err := sessionsRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "2m30.05s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}
}

func testSessionsVoiceRelocateWithOriginIDPrefix(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "TestRelocateWithOriginIDPrefix",
		BalanceType: utils.MetaVoice,
		Value:       300 * float64(time.Second),
		Balance: map[string]interface{}{
			utils.ID:            "TestRelocateWithOriginIDPrefix",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := sessionsRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  attrSetBalance.Tenant,
		Account: attrSetBalance.Account,
	}
	eAcntVal := 300.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage := 2 * time.Minute
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceRelocateWithOriginIDPrefix",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT_RELOCATE_ORIGPREFIX",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12371",
				utils.AccountField: attrSetBalance.Account,
				utils.Subject:      attrSetBalance.Account,
				utils.Destination:  "12371",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl *V1InitSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	time.Sleep(20 * time.Millisecond)
	var aSessions []*ExternalSession
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12371"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 120*time.Second {
		t.Errorf("Expecting 2m, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 180.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f", eAcntVal,
			acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	//Update
	lastUsage := 30 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceSessionTTLWithRelocate",
			Event: map[string]interface{}{
				utils.EventName:       "TEST_EVENT_RELOCATE_ORIGPREFIX",
				utils.ToR:             utils.MetaVoice,
				utils.InitialOriginID: "12371",
				utils.OriginID:        "12372-1",
				utils.AccountField:    attrSetBalance.Account,
				utils.Subject:         attrSetBalance.Account,
				utils.Destination:     "12371",
				utils.Category:        "call",
				utils.Tenant:          "cgrates.org",
				utils.RequestType:     utils.MetaPrepaid,
				utils.SetupTime:       time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:      time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:           usage,
				utils.LastUsed:        lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := sessionsRPC.Call(utils.SessionSv1UpdateSession,
		updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	time.Sleep(20 * time.Millisecond)
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12372-1"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting 2m30s, received usage: %v", aSessions[0].Usage)
	}
	eAcntVal = 150.0 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsedNotFixed",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT_RELOCATE_ORIGPREFIX",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12372-1",
				utils.AccountField: attrSetBalance.Account,
				utils.Subject:      attrSetBalance.Account,
				utils.Destination:  "12371",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}

	var rpl string
	if err := sessionsRPC.Call(utils.SessionSv1TerminateSession,
		termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	time.Sleep(10 * time.Millisecond)
	if err := sessionsRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.RunID, utils.MetaDefault),
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "12372-1"),
			},
		}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	eAcntVal = 240 * float64(time.Second)
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount,
		attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expecting: %f, received: %f",
			eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	if err := sessionsRPC.Call(utils.SessionSv1ProcessCDR, termArgs.CGREvent, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
	time.Sleep(100 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RunIDs: []string{utils.MetaDefault},
		DestinationPrefixes: []string{"12371"}}
	if err := sessionsRPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 1 {
		t.Error("Unexpected number of CDRs returned: ", len(cdrs))
	} else {
		if cdrs[0].Usage != "1m0s" {
			t.Errorf("Unexpected CDR Usage received, cdr: %v %+v ", cdrs[0].Usage, cdrs[0])
		}
	}
}

//This test was commented before
/*
func TestSMGDataDerivedChargingNoCredit(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1011"}
	eAcntVal := 50000.0
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	smgEv := map[string]interface{}{
		utils.EventName:  "TEST_EVENT",
		utils.ToR:         utils.MetaVoice,
		utils.OriginID:       "1234967",
		utils.AccountField:     "1011",
		utils.Subject:     "1011",
		utils.Destination: "+49",
		utils.Category:    "call",
		utils.Tenant:      "cgrates.org",
		utils.RequestType:     utils.MetaPrepaid,
		utils.SetupTime:  "2016-01-05 18:30:49",
		utils.AnswerTime: "2016-01-05 18:31:05",
		utils.Usage:       "100",
	}
	var maxUsage float64
	if err := sessionsRPC.Call(utils.SMGenericV2InitiateSession, smgEv, &maxUsage); err != nil {
		t.Error(err)
	}
	// the second derived charging run has no credit

	if maxUsage != 0 {
		t.Error("Bad max usage: ", maxUsage)
	}
	eAcntVal = 50000.0
	if err := sessionsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
}
*/

// ToDo: Add test for ChargeEvent with derived charging, one with debit possible and second not so we see refund and error.CreditInsufficient showing up.
func testSessionsVoiceStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
