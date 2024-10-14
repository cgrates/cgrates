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
	sesCfgPath string
	sesCfgDIR  string
	sesCfg     *config.CGRConfig
	sesRPC     *birpc.Client
	sesAccount = "refundAcc"
	sesTenant  = "cgrates.org"

	sTestSesIt = []func(t *testing.T){
		testSesItLoadConfig,
		testSesItResetDataDB,
		testSesItResetStorDb,
		testSesItStartEngine,
		testSesItRPCConn,
		testSesItLoadFromFolder,
		testSesItAddVoiceBalance,
		testSesItInitSession,
		testSesItTerminateSession,
		testSesItStopCgrEngine,
	}
)

func TestSesIt(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesCfgDIR = "tutinternal"
	case utils.MetaMySQL:
		sesCfgDIR = "tutmysql_internal"
	case utils.MetaMongo:
		sesCfgDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSesIt {
		t.Run(sesCfgDIR, stest)
	}
}

// test for 0 balance with session terminate with 1s usage
func testSesItLoadConfig(t *testing.T) {
	var err error
	sesCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesCfgDIR)
	if sesCfg, err = config.NewCGRConfigFromPath(sesCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesItRPCConn(t *testing.T) {
	sesRPC = engine.NewRPCClient(t, sesCfg.ListenCfg())
}

func testSesItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := sesRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testAccountBalance2(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := sesRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testSesItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      sesTenant,
		Account:     sesAccount,
		BalanceType: utils.MetaMonetary,
		Value:       0,
		Balance: map[string]any{
			utils.ID:            "TestDynamicDebitBalance",
			utils.RatingSubject: "*zero1s",
		},
	}
	var reply string
	if err := sesRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MetaMonetary, 0) })
}

func testSesItInitSession(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: sesTenant,
			ID:     "TestSesItInitiateSession",
			Event: map[string]any{
				utils.Tenant:       sesTenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sesAccount,
				utils.Subject:      "TEST",
				utils.Destination:  "TEST",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        5 * time.Second,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sesRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage != nil && *rply1.MaxUsage != 0 {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MetaMonetary, 0) })
}

func testSesItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: sesTenant,
			ID:     "TestSesItUpdateSession",
			Event: map[string]any{
				utils.Tenant:       sesTenant,
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestRefund",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: sesAccount,
				utils.Subject:      "TEST",
				utils.Destination:  "TEST",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        time.Second,
			},
		},
	}
	var rply string
	if err := sesRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sesRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MetaMonetary, 0) })
}

func testSesItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
