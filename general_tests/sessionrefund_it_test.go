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
	srCfgPath string
	srCfgDIR  string
	srCfg     *config.CGRConfig
	srrpc     *birpc.Client
	sraccount = "refundAcc"
	srtenant  = "cgrates.org"

	sTestSrIt = []func(t *testing.T){
		testSrItLoadConfig,
		testSrItResetDataDB,
		testSrItResetStorDb,
		testSrItStartEngine,
		testSrItRPCConn,
		testSrItLoadFromFolder,
		testSrItAddVoiceBalance,
		testSrItInitSession,
		testSrItTerminateSession,
		testSrItAddMonetaryBalance,
		testSrItInitSession2,
		testSrItTerminateSession2,
		testSrItStopCgrEngine,
	}
)

func TestSrIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		srCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		srCfgDIR = "tutmysql_internal"
	case utils.MetaMongo:
		srCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSrIt {
		t.Run(srCfgDIR, stest)
	}
}

func testSrItLoadConfig(t *testing.T) {
	var err error
	srCfgPath = path.Join(*utils.DataDir, "conf", "samples", srCfgDIR)
	if srCfg, err = config.NewCGRConfigFromPath(srCfgPath); err != nil {
		t.Error(err)
	}
}

func testSrItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(srCfg); err != nil {
		t.Fatal(err)
	}
}

func testSrItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(srCfg); err != nil {
		t.Fatal(err)
	}
}

func testSrItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(srCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSrItRPCConn(t *testing.T) {
	srrpc = engine.NewRPCClient(t, srCfg.ListenCfg())
}

func testSrItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := srrpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAccountBalance(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := srrpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testSrItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      srtenant,
		Account:     sraccount,
		BalanceType: utils.MetaVoice,
		Value:       5 * float64(time.Second),
		Balance: map[string]any{
			utils.ID:            "TestDynamicDebitBalance",
			utils.RatingSubject: "*zero5ms",
		},
	}
	var reply string
	if err := srrpc.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) {
		testAccountBalance(t, sraccount, srtenant, utils.MetaVoice, 5*float64(time.Second))
	})
}

func testSrItInitSession(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,

		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItInitiateSession",
			Event: map[string]any{
				utils.Tenant:       srtenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sraccount,
				utils.Subject:      "TEST",
				utils.Destination:  "TEST",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        2 * time.Second,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := srrpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage == nil || *rply1.MaxUsage != 2*time.Second {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) {
		testAccountBalance(t, sraccount, srtenant, utils.MetaVoice, 3*float64(time.Second))
	})
}

func testSrItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItUpdateSession",
			Event: map[string]any{
				utils.Tenant:       srtenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sraccount,
				utils.Subject:      "TEST",
				utils.Destination:  "TEST",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        0,
			},
		},
	}
	var rply string
	if err := srrpc.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := srrpc.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) {
		testAccountBalance(t, sraccount, srtenant, utils.MetaVoice, 5*float64(time.Second))
	})
}

func testSrItAddMonetaryBalance(t *testing.T) {
	sraccount += "2"
	attrs := &utils.AttrSetBalance{
		Tenant:      srtenant,
		Account:     sraccount,
		BalanceType: utils.MetaMonetary,
		Value:       10.65,
		Balance: map[string]any{
			utils.ID: utils.MetaDefault,
		},
	}
	var reply string
	if err := srrpc.Call(context.Background(), utils.APIerSv2SetBalance, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddMonetaryBalance", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MetaMonetary, 10.65) })
}

func testSrItInitSession2(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,

		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItInitiateSession1",
			Event: map[string]any{
				utils.Tenant:       srtenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sraccount,
				utils.Subject:      "TEST",
				utils.Destination:  "1001",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := srrpc.Call(context.Background(), utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage == nil || *rply1.MaxUsage != 2*time.Minute {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MetaMonetary, 10.3002) })
}

func testSrItTerminateSession2(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,

		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItUpdateSession",
			Event: map[string]any{
				utils.Tenant:       srtenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sraccount,
				utils.Subject:      "TEST",
				utils.Destination:  "1001",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        0,
			},
		},
	}
	var rply string
	if err := srrpc.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := srrpc.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MetaMonetary, 10.65) })
}

func testSrItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
