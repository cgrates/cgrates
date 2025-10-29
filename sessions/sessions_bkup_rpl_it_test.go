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
	"fmt"
	"os/exec"
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
	sBRplcEng1CfgPath, sBRplcEng2CfgPath string
	sBRplcEng1CfgDIR, sBRplcEng2CfgDIR   string
	sBRplcEng1Cfg, sBRplcEng2Cfg         *config.CGRConfig
	sBRplcEng1RPC, sBRplcEng2RPC         *birpc.Client
	sBRplcEng1Eng, sBRplcEng2Eng         *exec.Cmd

	SessionsBkupRplTests = []func(t *testing.T){
		testSessionSBkupRplcInitCfg,
		testSessionSBkupRplcResetDB,
		testSessionSBkupRplcStartEngineBoth,
		testSessionSBkupRplcApierRpcConnBoth,
		testSessionSBkupRplcTPFromFolder,
		testSessionSBkupRplcInitiate,
		testSessionSBkupRplcGetSessions,
		testSessionSBkupRplcUpdate,
		testSessionSBkupRplcStartEngine2,
		testSessionSBkupRplcApierRpcConn2,
		testSessionSBkupRplcGetActvSessionsFromRestored,
		testSessionSBkupRplcTerminate,
		testSessionSBkupRplcStopCgrEngine,

		testSessionSBkupRplcStartEngine2,
		testSessionSBkupRplcApierRpcConn2,
		testSessionSBkupRplcGetNoActvSessionsFromRestored,
		testSessionSBkupRplcStopCgrEngine,
	}
)

func TestSessionSBkupRplc(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		sBRplcEng1CfgDIR = "sbkupreplcengine1_mysql"
		sBRplcEng2CfgDIR = "sbkupreplcengine2_mongo"
	case utils.MetaMongo:
		sBRplcEng1CfgDIR = "sbkupreplcengine1_mongo"
		sBRplcEng2CfgDIR = "sbkupreplcengine2_mysql"
	case utils.MetaPostgres:
		sBRplcEng1CfgDIR = "sbkupreplcengine1_postgres"
		sBRplcEng2CfgDIR = "sbkupreplcengine2_mongo"
	default:
		t.Fatal("Unknown Database type")
	}
	if *utils.Encoding == utils.MetaGOB {
		sBRplcEng1CfgDIR += "_gob"
		sBRplcEng2CfgDIR += "_gob"
	}
	for _, stest := range SessionsBkupRplTests {
		t.Run(*utils.DBType, stest)
	}
}

func testSessionSBkupRplcInitCfg(t *testing.T) {
	sBRplcEng1CfgPath = path.Join(*utils.DataDir, "conf", "samples", sBRplcEng1CfgDIR)
	if sBRplcEng1Cfg, err = config.NewCGRConfigFromPath(sBRplcEng1CfgPath); err != nil {
		t.Fatal(err)
	}
	sBRplcEng2CfgPath = path.Join(*utils.DataDir, "conf", "samples", sBRplcEng2CfgDIR)
	if sBRplcEng2Cfg, err = config.NewCGRConfigFromPath(sBRplcEng2CfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSBkupRplcResetDB(t *testing.T) {
	if err := engine.InitDataDB(sBRplcEng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(sBRplcEng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDB(sBRplcEng2Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(sBRplcEng2Cfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionSBkupRplcStartEngineBoth(t *testing.T) {
	if sBRplcEng2Eng, err = engine.StopStartEngine(sBRplcEng2CfgPath, *utils.WaitRater); err != nil { // Start engine2 before engine1
		t.Fatal(err)
	}
	if sBRplcEng1Eng, err = engine.StartEngine(sBRplcEng1CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionSBkupRplcApierRpcConnBoth(t *testing.T) {
	if sBRplcEng1RPC, err = newRPCClient(sBRplcEng1Cfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	if sBRplcEng2RPC, err = newRPCClient(sBRplcEng2Cfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSessionSBkupRplcTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := sBRplcEng1RPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
	attrs = &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := sBRplcEng2RPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionSBkupRplcInitiate(t *testing.T) {
	var aSessions []*ExternalSession
	//make sure we don't have active sessions on engine1 and passive on engine2
	if err := sBRplcEng1RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	usage := time.Minute + 30*time.Second
	argsInit := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSBkupRplcInitiate",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123451",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl V1InitSessionReply
	if err := sBRplcEng1RPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated

	//check if the session was createad as active session on engine1
	if err := sBRplcEng1RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	} else if aSessions[0].Usage != 90*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 90*time.Second, aSessions[0].Usage)
	}

	//check if the session was created as passive session on engine2
	var pSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123451"),
			},
		}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage != 90*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 90*time.Second, pSessions[0].Usage)
	}
}

func testSessionSBkupRplcGetSessions(t *testing.T) {
	time.Sleep(501 * time.Millisecond) // make sure active sessions are backed up from "backup_interval"
	if err := sBRplcEng1Eng.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process, error: %v", err.Error())
	}
	// make sure we have no active sessions on engine2
	var aSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	// make sure we have passive sessions on engine2
	var pSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		utils.SessionFilter{}, &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].Usage != 90*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 90*time.Second, pSessions[0].Usage)
	}
}

func testSessionSBkupRplcUpdate(t *testing.T) {
	//update the session on engine2 so the session should became active
	usage := time.Minute
	argsUpdate := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSBkupRplcUpdate",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123451",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}
	var updtRpl V1UpdateSessionReply
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1UpdateSession,
		argsUpdate, &updtRpl); err != nil {
		t.Error(err)
	}
	if updtRpl.MaxUsage == nil || *updtRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, updtRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated

	var aSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 150*time.Second, aSessions[0].Usage)
	}

	var pSessions []*ExternalSession
	// Make sure we don't have passive session on active host
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter),
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	time.Sleep(501 * time.Millisecond) // make sure active sessions are backed up from "backup_interval"
	if err := sBRplcEng2Eng.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process, error: %v", err.Error())
	}
}

func testSessionSBkupRplcStartEngine2(t *testing.T) {
	if sBRplcEng2Eng, err = engine.StopStartEngine(sBRplcEng2CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSessionSBkupRplcApierRpcConn2(t *testing.T) {
	if sBRplcEng2RPC, err = newRPCClient(sBRplcEng2Cfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testSessionSBkupRplcGetActvSessionsFromRestored(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	} else if aSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 150*time.Second, aSessions[0].Usage)
	}

	var pSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSessionSBkupRplcTerminate(t *testing.T) {
	args := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSBkupRplcTerminate",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123451",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        2*time.Minute + 30*time.Second,
			},
		},
	}
	var reply string
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1TerminateSession, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ExternalSession
	//check if the session was terminated on engine2
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(aSessions), utils.ToIJSON(aSessions))
	}
	// make sure we don't have passive session on engine2
	var pSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter),
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(pSessions)=%v , session : %+v", err, len(pSessions), utils.ToIJSON(pSessions))
	}
}

func testSessionSBkupRplcGetNoActvSessionsFromRestored(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{}, &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	var pSessions []*ExternalSession
	if err := sBRplcEng2RPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSessionSBkupRplcStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
