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
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgRplcMasterCfgPath, smgRplcSlaveCfgPath string
	smgRplcMasterCfg, smgRplcSlaveCfg         *config.CGRConfig
	smgRplcMstrRPC, smgRplcSlvRPC             *rpc.Client
	masterEngine                              *exec.Cmd
	sTestsSession                             = []func(t *testing.T){
		testSessionSRplInitCfg,
		testSessionSRplResetDB,
		testSessionSRplStartEngine,
		testSessionSRplApierRpcConn,
		testSessionSRplTPFromFolder,
		testSessionSRplAddVoiceBalance,
		testSessionSRplInitiate,
		testSessionSRplActivateSlave,
		testSessionSRplCheckAccount,
		testSessionSRplTerminate,
		testSessionSRplStopCgrEngine,
	}
)

func TestSessionSRpl(t *testing.T) {
	for _, stest := range sTestsSession {
		t.Run("TestSessionSRpl", stest)
	}
}

//topup
func testSessionSRplAddVoiceBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:        "cgrates.org",
		Account:       "1005",
		BalanceType:   utils.VOICE,
		BalanceID:     utils.StringPointer("TestDynamicDebitBalance"),
		Value:         utils.Float64Pointer(5 * float64(time.Second)), //value -> 20ms for future
		RatingSubject: utils.StringPointer("*zero5ms"),
	}
	var reply string
	if err := smgRplcMstrRPC.Call(utils.ApierV2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}
	//get balance
	if err := smgRplcMstrRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != float64(5*time.Second) {
		t.Errorf("Expecting: %v, received: %v",
			float64(5*time.Second), rply)
	}
}

//Init Config
func testSessionSRplInitCfg(t *testing.T) {
	smgRplcMasterCfgPath = path.Join(*dataDir, "conf", "samples", "sessions_replication", "smgreplcmaster")
	if smgRplcMasterCfg, err = config.NewCGRConfigFromPath(smgRplcMasterCfgPath); err != nil {
		t.Fatal(err)
	}
	smgRplcMasterCfg.DataFolderPath = *dataDir
	config.SetCgrConfig(smgRplcMasterCfg)
	smgRplcSlaveCfgPath = path.Join(*dataDir, "conf", "samples", "sessions_replication", "smgreplcslave")
	if smgRplcSlaveCfg, err = config.NewCGRConfigFromPath(smgRplcSlaveCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSRplResetDB(t *testing.T) {
	if err := engine.InitDataDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionSRplStartEngine(t *testing.T) {
	if _, err = engine.StopStartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if masterEngine, err = engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}

}

// Connect rpc client to rater
func testSessionSRplApierRpcConn(t *testing.T) {
	if smgRplcMstrRPC, err = jsonrpc.Dial("tcp", smgRplcMasterCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = jsonrpc.Dial("tcp", smgRplcSlaveCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSessionSRplTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := smgRplcMstrRPC.Call(utils.ApierV2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionSRplInitiate(t *testing.T) {
	var aSessions []*sessions.ExternalSession
	//make sure we don't have active sessions on master and passive on slave
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions,
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	argsInit := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSRplInitiate",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.Tenant:      "cgrates.org",
				utils.OriginID:    "123451",
				utils.ToR:         utils.VOICE,
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1005",
				utils.Subject:     "1005",
				utils.Destination: "1004",
				utils.Category:    "call",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       0,
			},
		},
	}

	var initRpl sessions.V1InitSessionReply
	if err := smgRplcMstrRPC.Call(utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	eMaxUsage := 3 * time.Hour // MaxCallDuration from config
	if *initRpl.MaxUsage != eMaxUsage {
		t.Errorf("Expecting : %+v, received: %+v", eMaxUsage, initRpl.MaxUsage)
	}

	//check active session
	time.Sleep(10 * time.Millisecond)
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
		// a tolerance of +/- 5ms is acceptable
	} else if aSessions[0].Usage < 5*time.Millisecond || aSessions[0].Usage > 15*time.Millisecond {
		t.Errorf("Expecting : ~%+v, received: %+v", 10*time.Millisecond, aSessions[0].Usage) //here
	}
	//check passive session
	var autoDebit1, autoDebit2 time.Time

	var pSessions []*sessions.ExternalSession
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~%s:%s", utils.OriginID, "123451"),
			},
		}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage < 5*time.Millisecond || pSessions[0].Usage > 15*time.Millisecond {
		t.Errorf("Expecting : %+v, received: %+v", 10*time.Millisecond, pSessions[0].Usage)
	} else if autoDebit1 = pSessions[0].NextAutoDebit; autoDebit1.IsZero() {
		t.Errorf("unexpected NextAutoDebit: %s", utils.ToIJSON(aSessions[0]))
	}

	//check active session (II)
	time.Sleep(10 * time.Millisecond)
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
		// a tolerance of +/- 5ms is acceptable
	} else if aSessions[0].Usage < 15*time.Millisecond || aSessions[0].Usage > 25*time.Millisecond {
		t.Errorf("Expecting : ~%+v, received: %+v", 20*time.Millisecond, aSessions[0].Usage) //here
	}

	//check passive session (II)
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~%s:%s", utils.OriginID, "123451"),
			},
		}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage < 15*time.Millisecond || pSessions[0].Usage > 25*time.Millisecond {
		t.Errorf("Expecting : %+v, received: %+v", 20*time.Millisecond, pSessions[0].Usage)
	} else if autoDebit2 = pSessions[0].NextAutoDebit; autoDebit2.IsZero() {
		t.Errorf("unexpected NextAutoDebit: %s", utils.ToIJSON(aSessions[0]))
	} else if autoDebit1 == autoDebit2 {
		t.Error("Expecting NextAutoDebit to be different from the previous one")
	}

	//get balance
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}
	if err := smgRplcMstrRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
		// a tolerance of +/- 5ms is acceptable
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply < float64(5*time.Second-25*time.Millisecond) || rply > float64(5*time.Second-15*time.Millisecond) {
		t.Errorf("Expecting: ~%v, received: %v", float64(5*time.Second-20*time.Millisecond), rply)
	}
}

func testSessionSRplStopMasterEngine(t *testing.T) {
	//stop the master engine
	if err := masterEngine.Process.Kill(); err != nil {
		t.Error(err)
	}
	//check if the active session is on slave now
	var aSessions []*sessions.ExternalSession
	time.Sleep(10 * time.Millisecond)
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
		// a tolerance of +/- 5ms is acceptable
	} else if aSessions[0].Usage < 15*time.Millisecond || aSessions[0].Usage > 25*time.Millisecond {
		t.Errorf("Expecting : ~%+v, received: %+v", 20*time.Millisecond, aSessions[0].Usage) //here
	}
}

func testSessionSRplActivateSlave(t *testing.T) {
	//stop the master engine
	if err := masterEngine.Process.Kill(); err != nil {
		t.Error(err)
	}
	// activate sessions on slave
	var rplActivate string
	if err := smgRplcSlvRPC.Call(utils.SessionSv1ActivateSessions, nil, &rplActivate); err != nil {
		t.Error(err)
	}
	time.Sleep(5 * time.Millisecond)
	//check if the active session is on slave now
	var aSessions []*sessions.ExternalSession
	var autoDebit1, autoDebit2 time.Time
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
		// a tolerance of +/- 5ms is acceptable
	} else if aSessions[0].Usage < 20*time.Millisecond || aSessions[0].Usage > 30*time.Millisecond {
		t.Errorf("Expecting : ~%+v, received: %+v", 25*time.Millisecond, aSessions[0].Usage) //here
	} else if autoDebit1 = aSessions[0].NextAutoDebit; autoDebit1.IsZero() {
		t.Errorf("unexpected NextAutoDebit: %s", utils.ToIJSON(aSessions[0]))
	}
	time.Sleep(10 * time.Millisecond)
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions, nil, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
		// a tolerance of +/- 5ms is acceptable
	} else if aSessions[0].Usage < 30*time.Millisecond || aSessions[0].Usage > 45*time.Millisecond {
		t.Errorf("Expecting : ~%+v, received: %+v", 40*time.Millisecond, aSessions[0].Usage) //here
	} else if autoDebit2 = aSessions[0].NextAutoDebit; autoDebit2.IsZero() {
		t.Errorf("unexpected NextAutoDebit: %s", utils.ToIJSON(aSessions[0]))
	} else if autoDebit1 == autoDebit2 {
		t.Error("Expecting NextAutoDebit to be different from the previous one")
	}
}

func testSessionSRplCheckAccount(t *testing.T) {
	//check de account and make sure the session debit works correctly
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}

	if err := smgRplcSlvRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
		// a tolerance of +/- 5ms is acceptable
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply < float64(5*time.Second-45*time.Millisecond) || rply > float64(5*time.Second-30*time.Millisecond) {
		t.Errorf("Expecting: ~%v, received: %v", float64(5*time.Second-40*time.Millisecond), rply)
	}
}

func testSessionSRplTerminate(t *testing.T) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSRplTerminate",
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.Tenant:      "cgrates.org",
				utils.OriginID:    "123451",
				utils.ToR:         utils.VOICE,
				utils.RequestType: utils.META_PREPAID,
				utils.Account:     "1005",
				utils.Subject:     "1005",
				utils.Destination: "1004",
				utils.Category:    "call",
				utils.SetupTime:   time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:  time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:       2 * time.Second,
			},
		},
	}
	var reply string
	if err := smgRplcSlvRPC.Call(utils.SessionSv1TerminateSession, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*sessions.ExternalSession

	//check if the session was terminated on slave
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions,
		nil, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(aSessions), utils.ToIJSON(aSessions))
	}
	// check to don't have passive session on slave
	var pSessions []*sessions.ExternalSession
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, nil,
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(pSessions)=%v , session : %+v", err, len(pSessions), utils.ToIJSON(pSessions))
	}

	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}

	if err := smgRplcSlvRPC.Call(utils.ApierV2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
		// a tolerance of +/- 5ms is acceptable
	} else if rply := acnt.BalanceMap[utils.VOICE].GetTotalValue(); rply != float64(3*time.Second) {
		t.Errorf("Expecting: ~%v, received: %v", 3*time.Second, rply)
	}
}

func testSessionSRplStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
