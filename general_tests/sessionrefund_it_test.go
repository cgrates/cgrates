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
	srCfgPath string
	srCfg     *config.CGRConfig
	srrpc     *rpc.Client
	sraccount = "refundAcc"
	srtenant  = "cgrates.org"
)

func TestSrItLoadConfig(t *testing.T) {
	srCfgPath = path.Join(*dataDir, "conf", "samples", "tutmongo")
	if srCfg, err = config.NewCGRConfigFromPath(srCfgPath); err != nil {
		t.Error(err)
	}
}

func TestSrItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(srCfg); err != nil {
		t.Fatal(err)
	}
}

func TestSrItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(srCfg); err != nil {
		t.Fatal(err)
	}
}

func TestSrItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(srCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func TestSrItRPCConn(t *testing.T) {
	var err error
	srrpc, err = jsonrpc.Dial("tcp", srCfg.ListenCfg().RPCJSONListen)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSrItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := srrpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testAccountBalance(t *testing.T, sracc, srten, balType string, expected float64) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  srten,
		Account: sracc,
	}
	if err := srrpc.Call("ApierV2.GetAccount", attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[balType].GetTotalValue(); rply != expected {
		t.Errorf("Expecting: %v, received: %v",
			expected, rply)
	}
}

func TestSrItAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        srtenant,
		Account:       sraccount,
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestDynamicDebitBalance"),
		Value:         utils.Float64Pointer(5 * float64(time.Second)),
		RatingSubject: utils.StringPointer("*zero5ms"),
	}
	var reply string
	if err := srrpc.Call("ApierV2.SetBalance", attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddVoiceBalance", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.VOICE, 5*float64(time.Second)) })
}

func TestSrItInitSession(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItInitiateSession",
			Event: map[string]interface{}{
				utils.Tenant:      srtenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sraccount,
				utils.Subject:     "TEST",
				utils.Destination: "TEST",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       2 * time.Second,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := srrpc.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if *rply1.MaxUsage != 2*time.Second {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.VOICE, 3*float64(time.Second)) })
}

func TestSrItTerminateSession(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      srtenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sraccount,
				utils.Subject:     "TEST",
				utils.Destination: "TEST",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       0 * time.Second,
			},
		},
	}
	var rply string
	if err := srrpc.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := srrpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.VOICE, 5*float64(time.Second)) })
}

func TestSrItAddMonetaryBalance(t *testing.T) {
	sraccount += "2"
	attrs := &utils.AttrSetBalance{
		Tenant:      srtenant,
		Account:     sraccount,
		BalanceType: utils.MONETARY,
		BalanceID:   utils.StringPointer(utils.META_DEFAULT),
		Value:       utils.Float64Pointer(10.65),
	}
	var reply string
	if err := srrpc.Call("ApierV2.SetBalance", attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	t.Run("TestAddMonetaryBalance", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MONETARY, 10.65) })
}

func TestSrItInitSession2(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItInitiateSession1",
			Event: map[string]interface{}{
				utils.Tenant:      srtenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sraccount,
				utils.Subject:     "TEST",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := srrpc.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if *rply1.MaxUsage != 2*time.Minute {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
	t.Run("TestInitSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MONETARY, 10.3002) })
}

func TestSrItTerminateSession2(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: srtenant,
			ID:     "TestSrItUpdateSession",
			Event: map[string]interface{}{
				utils.Tenant:      srtenant,
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "TestRefund",
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     sraccount,
				utils.Subject:     "TEST",
				utils.Destination: "1001",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       0 * time.Second,
			},
		},
	}
	var rply string
	if err := srrpc.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	aSessions := make([]*sessions.ActiveSession, 0)
	if err := srrpc.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	t.Run("TestTerminateSession", func(t *testing.T) { testAccountBalance(t, sraccount, srtenant, utils.MONETARY, 10.65) })
}

func TestSrItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
