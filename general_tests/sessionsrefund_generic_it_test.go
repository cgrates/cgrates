//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sessRefGenCfgPath string
	sessRefGenCfgDIR  string
	sessRefGenCfg     *config.CGRConfig
	sessRefGenRPC     *birpc.Client

	sTestSessRefGenIt = []func(t *testing.T){
		testSessRefGenItLoadConfig,
		testSessRefGenItResetDataDB,
		testSessRefGenItResetStorDb,
		testSessRefGenItStartEngine,
		testSessRefGenItRPCConn,
		testSessRefGenItLoadFromFolder,
		testSessRefGenItVoiceUnusedReservationRefunded,
		testSessRefGenItSMSWithFactorUnusedReservationRefunded,
		testSessRefGenItVoiceWithPairedMonetaryUnusedReservationRefunded,
		testSessRefGenItStopCgrEngine,
	}
)

func TestSessionsRefundGenericIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sessRefGenCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		sessRefGenCfgDIR = "tutmysql_internal"
	case utils.MetaMongo:
		sessRefGenCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSessRefGenIt {
		t.Run(sessRefGenCfgDIR, stest)
	}
}

func testSessRefGenItLoadConfig(t *testing.T) {
	var err error
	sessRefGenCfgPath = path.Join(*utils.DataDir, "conf", "samples", sessRefGenCfgDIR)
	if sessRefGenCfg, err = config.NewCGRConfigFromPath(sessRefGenCfgPath); err != nil {
		t.Error(err)
	}
}

func testSessRefGenItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(sessRefGenCfg); err != nil {
		t.Fatal(err)
	}
}

func testSessRefGenItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sessRefGenCfg); err != nil {
		t.Fatal(err)
	}
}

func testSessRefGenItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sessRefGenCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSessRefGenItRPCConn(t *testing.T) {
	sessRefGenRPC = engine.NewRPCClient(t, sessRefGenCfg.ListenCfg())
}

func testSessRefGenItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSessRefGenItVoiceUnusedReservationRefunded(t *testing.T) {
	var reply string
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundGenVoice",
		BalanceType: utils.MetaGeneric,
		Value:       100 * float64(time.Second),
		Balance: map[string]any{
			utils.ID:            "bundle_mix",
			utils.RatingSubject: "*zero5ms",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "refundGenVoice"}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice: balance before the session: %v", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 100*float64(time.Second); rcv != exp {
		t.Fatalf("initial balance: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}

	var initRply sessions.V1InitSessionReply
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "init_refundGenVoice",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundGeneric_refundGenVoice",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundGenVoice",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
					utils.SetupTime:    time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:        30 * time.Second,
				},
			},
		}, &initRply); err != nil {
		t.Fatal(err)
	} else if initRply.MaxUsage == nil || *initRply.MaxUsage != 30*time.Second {
		t.Errorf("Unexpected MaxUsage: %v", initRply.MaxUsage)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice: 30s reserved, balance after the reservation: %v", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 70*float64(time.Second); rcv != exp {
		t.Fatalf("after init: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}

	var termRply string
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "term_refundGenVoice",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundGeneric_refundGenVoice",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundGenVoice",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
					utils.SetupTime:    time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:        11 * time.Second,
				},
			},
		}, &termRply); err != nil {
		t.Fatal(err)
	} else if termRply != utils.OK {
		t.Errorf("Unexpected reply: %s", termRply)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice: 11s used out of 30s reserved, balance after terminate: %v (the unused 19s should be refunded)", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 89*float64(time.Second); rcv != exp {
		t.Errorf("after terminate: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}
}

func testSessRefGenItSMSWithFactorUnusedReservationRefunded(t *testing.T) {
	var reply string
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundGenSMS",
		BalanceType: utils.MetaGeneric,
		Value:       600 * float64(time.Second),
		Balance: map[string]any{
			utils.ID:            "bundle_mix",
			utils.RatingSubject: "*zero1ns",
			"Factors":           map[string]any{"sms": 60 * float64(time.Second)},
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "refundGenSMS"}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("sms: balance before the session: %v", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 600*float64(time.Second); rcv != exp {
		t.Fatalf("initial balance: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}

	var initRply sessions.V1InitSessionReply
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "init_refundGenSMS",
				Event: map[string]any{
					utils.Tenant:          "cgrates.org",
					utils.Category:        "call",
					utils.ToR:             utils.MetaSMS,
					utils.OriginID:        "TestRefundGeneric_refundGenSMS",
					utils.RequestType:     utils.MetaPrepaid,
					utils.AccountField:    "refundGenSMS",
					utils.Subject:         "TEST",
					utils.Destination:     "TEST",
					utils.SetupTime:       time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:      time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:           time.Duration(3),
					utils.BalanceFactorID: "sms",
				},
			},
		}, &initRply); err != nil {
		t.Fatal(err)
	} else if initRply.MaxUsage == nil || *initRply.MaxUsage != time.Duration(3) {
		t.Errorf("Unexpected MaxUsage: %v", initRply.MaxUsage)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("sms: 3 SMS reserved (1m0s each), balance after the reservation: %v", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 420*float64(time.Second); rcv != exp {
		t.Fatalf("after init: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}

	var termRply string
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "term_refundGenSMS",
				Event: map[string]any{
					utils.Tenant:          "cgrates.org",
					utils.Category:        "call",
					utils.ToR:             utils.MetaSMS,
					utils.OriginID:        "TestRefundGeneric_refundGenSMS",
					utils.RequestType:     utils.MetaPrepaid,
					utils.AccountField:    "refundGenSMS",
					utils.Subject:         "TEST",
					utils.Destination:     "TEST",
					utils.SetupTime:       time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:      time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:           time.Duration(1),
					utils.BalanceFactorID: "sms",
				},
			},
		}, &termRply); err != nil {
		t.Fatal(err)
	} else if termRply != utils.OK {
		t.Errorf("Unexpected reply: %s", termRply)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("sms: 1 SMS used out of 3 reserved, balance after terminate: %v (the 2 unused SMS should be refunded)", time.Duration(acnt.BalanceMap[utils.MetaGeneric].GetTotalValue()))
	if rcv, exp := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), 540*float64(time.Second); rcv != exp {
		t.Errorf("after terminate: expected %v, received %v", time.Duration(exp), time.Duration(rcv))
	}
}

func testSessRefGenItVoiceWithPairedMonetaryUnusedReservationRefunded(t *testing.T) {
	var reply string
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundGenVoicePaired",
		BalanceType: utils.MetaMonetary,
		Value:       100,
		Balance: map[string]any{
			utils.ID: "monetary_main",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundGenVoicePaired",
		BalanceType: utils.MetaGeneric,
		Value:       100,
		Balance: map[string]any{
			utils.ID:            "bundle_mix",
			utils.RatingSubject: "1001",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "refundGenVoicePaired"}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice+paired monetary: before the session: generic %v, monetary %v",
		acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	if rcv := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(); rcv != 100 {
		t.Fatalf("initial generic balance: expected 100, received %v", rcv)
	}

	var initRply sessions.V1InitSessionReply
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "init_refundGenVoicePaired",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundGeneric_refundGenVoicePaired",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundGenVoicePaired",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:        time.Duration(3),
				},
			},
		}, &initRply); err != nil {
		t.Fatal(err)
	} else if initRply.MaxUsage == nil || *initRply.MaxUsage != time.Duration(3) {
		t.Errorf("Unexpected MaxUsage: %v", initRply.MaxUsage)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice+paired monetary: 3 units reserved, balances after the reservation: generic %v, monetary %v (3 debited from each)",
		acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	if rcv := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(); rcv != 97 {
		t.Fatalf("after init: expected generic 97, received %v", rcv)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 97 {
		t.Fatalf("after init: expected monetary 97, received %v", rcv)
	}

	var termRply string
	if err := sessRefGenRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "term_refundGenVoicePaired",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundGeneric_refundGenVoicePaired",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundGenVoicePaired",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2018, time.January, 7, 17, 0, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2018, time.January, 7, 17, 0, 10, 0, time.UTC),
					utils.Usage:        time.Duration(1),
				},
			},
		}, &termRply); err != nil {
		t.Fatal(err)
	} else if termRply != utils.OK {
		t.Errorf("Unexpected reply: %s", termRply)
	}
	if err := sessRefGenRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	t.Logf("voice+paired monetary: 1 unit used out of 3 reserved, balances after terminate: generic %v, monetary %v (the unused 2 should be refunded to each)",
		acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(), acnt.BalanceMap[utils.MetaMonetary].GetTotalValue())
	if rcv := acnt.BalanceMap[utils.MetaGeneric].GetTotalValue(); rcv != 99 {
		t.Errorf("after terminate: expected generic 99, received %v", rcv)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 99 {
		t.Errorf("after terminate: expected monetary 99, received %v", rcv)
	}
}

func testSessRefGenItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
