//go:build integration
// +build integration

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package sessions

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tUsageCfgPath string
	tUsageCfgDIR  string
	tUsageCfg     *config.CGRConfig
	tUsageRPC     *birpc.Client

	sessionsTUsageTests = []func(t *testing.T){
		testSessionsTUsageInitCfg,
		testSessionsTUsageResetDataDb,
		testSessionsTUsageResetStorDb,
		testSessionsTUsageStartEngine,
		testSessionsTUsageApierRpcConn,
		testSessionsTUsageTPFromFolder,
		testSessionsTUsageHigherThanSessionTUsage,
		testSessionsTUsageTPFromFolder,
		testSessionsTUsageLowerThanSessionTUsage,
		testSessionsTUsageTPFromFolder,
		testSessionsTUsageSameWithSessionTUsage,
		testSessionsTUsageStopCgrEngine,
	}
)

func TestSessionsTUsage(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tUsageCfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		tUsageCfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		tUsageCfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sessionsTUsageTests {
		t.Run(tUsageCfgDIR, stest)
	}
}

func testSessionsTUsageInitCfg(t *testing.T) {
	tUsageCfgPath = path.Join(*utils.DataDir, "conf", "samples", tUsageCfgDIR)
	// Init config first
	var err error
	tUsageCfg, err = config.NewCGRConfigFromPath(tUsageCfgPath)
	if err != nil {
		t.Error(err)
	}
}

// Remove data in both rating and accounting db
func testSessionsTUsageResetDataDb(t *testing.T) {
	if err := engine.InitDataDB(tUsageCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testSessionsTUsageResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(tUsageCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionsTUsageStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tUsageCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionsTUsageApierRpcConn(t *testing.T) {
	var err error
	tUsageRPC, err = newRPCClient(tUsageCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSessionsTUsageTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionsTUsageHigherThanSessionTUsage(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 10.0
	eAcntVoiceVal := float64(210 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage := 5 * time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageMonetaryRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(205 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage := 5 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update1",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(200 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
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

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(195 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	ignoredUsage := 435 * time.Second
	totalUsage := 105 * time.Second
	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        ignoredUsage,
				utils.LastUsed:     lastUsage,
				utils.TotalUsage:   totalUsage,
			},
		},
	}

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != totalUsage-15*time.Second { // used 15 seconds up until this update. to get the max usage of this specific update it does totalUsage-15*time.Second
		t.Errorf("Expected: %+v, received: %+v", totalUsage-15*time.Second, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(totalUsage)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	usage = 110 * time.Second
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageVoiceRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(110*time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsTUsageLowerThanSessionTUsage(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 15.0
	eAcntVoiceVal := float64(210 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage := 5 * time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageMonetaryRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(205 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage := 5 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update1",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(200 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
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

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(195 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	ignoredUsage := 435 * time.Second
	totalUsage := 5 * time.Second
	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        ignoredUsage,
				utils.LastUsed:     lastUsage,
				utils.TotalUsage:   totalUsage,
			},
		},
	}

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != 0 { // event total usage 5 is smaller than session total usage 15, return 0 max usage
		t.Errorf("Expected: %+v, received: %+v", 0, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(25*time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	usage = 110 * time.Second
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageVoiceRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(110*time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsTUsageSameWithSessionTUsage(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	eAcntVal := 20.0
	eAcntVoiceVal := float64(210 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	usage := 5 * time.Second
	initArgs := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageMonetaryRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Error(err)
	}
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(205 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage := 5 * time.Second
	updateArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update1",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.Usage:        usage,
				utils.LastUsed:     lastUsage,
			},
		},
	}

	var updateRpl *V1UpdateSessionReply
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(200 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
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

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != usage {
		t.Errorf("Expected: %+v, received: %+v", usage, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(195 * time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}

	ignoredUsage := 435 * time.Second
	totalUsage := 15 * time.Second
	lastUsage = 5 * time.Second
	updateArgs = &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsVoiceLastUsed",
			Event: map[string]any{
				utils.EventName:    "Update2",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1003",
				utils.Category:     "call",
				utils.Tenant:       "cgrates.org",
				utils.RequestType:  utils.MetaPrepaid,
				utils.SetupTime:    time.Date(2016, time.January, 5, 18, 30, 49, 0, time.UTC),
				utils.AnswerTime:   time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
				utils.Usage:        ignoredUsage,
				utils.LastUsed:     lastUsage,
				utils.TotalUsage:   totalUsage,
			},
		},
	}

	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updateArgs, &updateRpl); err != nil {
		t.Error(err)
	}
	if updateRpl.MaxUsage == nil || *updateRpl.MaxUsage != 0 { // event total usage 15 is the same with the sessions total usage 15, return 0 max usage
		t.Errorf("Expected: %+v, received: %+v", 0, updateRpl.MaxUsage)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(15*time.Second) // no debiting was done
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	}
	usage = 110 * time.Second
	termArgs := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionsTUsageVoiceRefund",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "12350",
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
	if err := tUsageRPC.Call(context.Background(), utils.SessionSv1TerminateSession, termArgs, &rpl); err != nil || rpl != utils.OK {
		t.Error(err)
	}

	eAcntVoiceVal = float64(210*time.Second) - float64(110*time.Second)
	if err := tUsageRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if acnt.BalanceMap[utils.MetaVoice].GetTotalValue() != eAcntVoiceVal {
		t.Errorf("Expected: %f, received: %f", eAcntVoiceVal, acnt.BalanceMap[utils.MetaVoice].GetTotalValue())
	} else if acnt.BalanceMap[utils.MetaMonetary].GetTotalValue() != eAcntVal {
		t.Errorf("Expected: %f, received: %f", eAcntVal, acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	}
}

func testSessionsTUsageStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
