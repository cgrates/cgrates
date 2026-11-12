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
	sessRefMonCfgPath string
	sessRefMonCfgDIR  string
	sessRefMonCfg     *config.CGRConfig
	sessRefMonRPC     *birpc.Client

	sTestSessRefMonIt = []func(t *testing.T){
		testSessRefMonItLoadConfig,
		testSessRefMonItResetDataDB,
		testSessRefMonItResetStorDb,
		testSessRefMonItStartEngine,
		testSessRefMonItRPCConn,
		testSessRefMonItLoadFromFolder,
		testSessRefMonItVoiceUnusedReservationRefunded,
		testSessRefMonItSMSUnusedReservationRefunded,
		testSessRefMonItStopCgrEngine,
	}
)

func TestSessionsRefundMonetaryIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sessRefMonCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		sessRefMonCfgDIR = "tutmysql_internal"
	case utils.MetaMongo:
		sessRefMonCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSessRefMonIt {
		t.Run(sessRefMonCfgDIR, stest)
	}
}

func testSessRefMonItLoadConfig(t *testing.T) {
	var err error
	sessRefMonCfgPath = path.Join(*utils.DataDir, "conf", "samples", sessRefMonCfgDIR)
	if sessRefMonCfg, err = config.NewCGRConfigFromPath(sessRefMonCfgPath); err != nil {
		t.Error(err)
	}
}

func testSessRefMonItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(sessRefMonCfg); err != nil {
		t.Fatal(err)
	}
}

func testSessRefMonItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sessRefMonCfg); err != nil {
		t.Fatal(err)
	}
}

func testSessRefMonItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sessRefMonCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSessRefMonItRPCConn(t *testing.T) {
	sessRefMonRPC = engine.NewRPCClient(t, sessRefMonCfg.ListenCfg())
}

func testSessRefMonItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSessRefMonItVoiceUnusedReservationRefunded(t *testing.T) {
	var reply string
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundMonVoice",
		BalanceType: utils.MetaMonetary,
		Value:       100,
		Balance: map[string]any{
			utils.ID: "monetary_only",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "refundMonVoice"}
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 100 {
		t.Fatalf("initial balance: expected 100, received %v", rcv)
	}

	var initRply sessions.V1InitSessionReply
	if err := sessRefMonRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "init_refundMonVoice",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundMonetary_refundMonVoice",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundMonVoice",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
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
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 97 {
		t.Fatalf("after init: expected 97, received %v", rcv)
	}

	var termRply string
	if err := sessRefMonRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "term_refundMonVoice",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "TestRefundMonetary_refundMonVoice",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundMonVoice",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
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
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 99 {
		t.Errorf("after terminate: expected 99, received %v", rcv)
	}
}

func testSessRefMonItSMSUnusedReservationRefunded(t *testing.T) {
	var reply string
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2SetBalance, utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "refundMonSMS",
		BalanceType: utils.MetaMonetary,
		Value:       100,
		Balance: map[string]any{
			utils.ID: "monetary_only",
		},
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("Received: %s", reply)
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "refundMonSMS"}
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 100 {
		t.Fatalf("initial balance: expected 100, received %v", rcv)
	}

	var initRply sessions.V1InitSessionReply
	if err := sessRefMonRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		&sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "init_refundMonSMS",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaSMS,
					utils.OriginID:     "TestRefundMonetary_refundMonSMS",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundMonSMS",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
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
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 97 {
		t.Fatalf("after init: expected 97, received %v", rcv)
	}

	var termRply string
	if err := sessRefMonRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&sessions.V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "term_refundMonSMS",
				Event: map[string]any{
					utils.Tenant:       "cgrates.org",
					utils.Category:     "generic",
					utils.ToR:          utils.MetaSMS,
					utils.OriginID:     "TestRefundMonetary_refundMonSMS",
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "refundMonSMS",
					utils.Subject:      "TEST",
					utils.Destination:  "TEST",
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
	if err := sessRefMonRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}
	if rcv := acnt.BalanceMap[utils.MetaMonetary].GetTotalValue(); rcv != 99 {
		t.Errorf("after terminate: expected 99, received %v", rcv)
	}
}

func testSessRefMonItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
