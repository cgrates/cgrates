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
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesCfgPath string
	sesCfg     *config.CGRConfig
	sesRPC     *rpc.Client
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
	for _, stest := range sTestSesIt {
		t.Run("TestSesIT", stest)
	}
}

// test for 0 balance with session terminate with 1s usage
func testSesItLoadConfig(t *testing.T) {
	sesCfgPath = path.Join(*dataDir, "conf", "samples", "tutmysql_internal")
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
	if _, err := engine.StopStartEngine(sesCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesItRPCConn(t *testing.T) {
	var err error
	sesRPC, err = jsonrpc.Dial("tcp", sesCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func testSesItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "testit")}
	if err := sesRPC.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testAccountBalance2(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := sesRPC.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func testSesItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        sesTenant,
		Account:       sesAccount,
		BalanceType:   utils.MONETARY,
		BalanceID:     utils.StringPointer("TestDynamicDebitBalance"),
		Value:         utils.Float64Pointer(0),
		RatingSubject: utils.StringPointer("*zero1s"),
	}
	var reply string
	if err := sesRPC.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MONETARY, 0) })
}

func testSesItInitSession(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: sesTenant,
			ID:     "TestSesItInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      sesTenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sesAccount,
				utils.Subject:     "TEST",
				utils.Destination: "TEST",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       5 * time.Second,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sesRPC.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if *rply1.MaxUsage != 0 {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MONETARY, 0) })
}

func testSesItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: sesTenant,
			ID:     "TestSesItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      sesTenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sesAccount,
				utils.Subject:     "TEST",
				utils.Destination: "TEST",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       1 * time.Second,
			},
		},
	}
	var rply string
	if err := sesRPC.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ExternalSession, 0)
	if err := sesRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) { testAccountBalance2(t, sesAccount, sesTenant, utils.MONETARY, 0) })
}

func testSesItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
