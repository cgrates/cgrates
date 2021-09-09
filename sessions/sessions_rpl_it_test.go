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
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgRplcMasterCfgPath, smgRplcSlaveCfgPath string
	smgRplcMasterCfgDIR, smgRplcSlaveCfgDIR   string
	smgRplcMasterCfg, smgRplcSlaveCfg         *config.CGRConfig
	smgRplcMstrRPC, smgRplcSlvRPC             *rpc.Client

	SessionsRplTests = []func(t *testing.T){
		testSessionSRplInitCfg,
		testSessionSRplResetDB,
		/*
			testSessionSRplStartEngine,
			testSessionSRplApierRpcConn,
			testSessionSRplTPFromFolder,
			testSessionSRplInitiate,
			testSessionSRplUpdate,
			testSessionSRplTerminate,
			testSessionSRplManualReplicate,
			testSessionSRplActivateSessions,
			testSessionSRplStopCgrEngine,
		*/
	}
)

func TestSessionSRpl(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		smgRplcMasterCfgDIR = "smgreplcmaster_mysql"
		smgRplcSlaveCfgDIR = "smgreplcslave_mysql"
	case utils.MetaMongo:
		smgRplcMasterCfgDIR = "smgreplcmaster_mongo"
		smgRplcSlaveCfgDIR = "smgreplcslave_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	if *encoding == utils.MetaGOB {
		smgRplcMasterCfgDIR += "_gob"
		smgRplcSlaveCfgDIR += "_gob"
	}
	for _, stest := range SessionsRplTests {
		t.Run(*dbType, stest)
	}
}

func testSessionSRplInitCfg(t *testing.T) {
	smgRplcMasterCfgPath = path.Join(*dataDir, "conf", "samples", smgRplcMasterCfgDIR)
	if smgRplcMasterCfg, err = config.NewCGRConfigFromPath(smgRplcMasterCfgPath); err != nil {
		t.Fatal(err)
	}
	smgRplcSlaveCfgPath = path.Join(*dataDir, "conf", "samples", smgRplcSlaveCfgDIR)
	if smgRplcSlaveCfg, err = config.NewCGRConfigFromPath(smgRplcSlaveCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSRplResetDB(t *testing.T) {
	if err := engine.InitDataDB(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(smgRplcMasterCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionSRplStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil { // Start slave before master
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionSRplApierRpcConn(t *testing.T) {
	if smgRplcMstrRPC, err = newRPCClient(smgRplcMasterCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = newRPCClient(smgRplcSlaveCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

/*

// Load the tariff plan, creating accounts and their balances
func testSessionSRplTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := smgRplcMstrRPC.Call(utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionSRplInitiate(t *testing.T) {
	var aSessions []*ExternalSession
	//make sure we don't have active sessions on master and passive on slave
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	usage := time.Minute + 30*time.Second
	argsInit := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSRplInitiate",
			Event: map[string]interface{}{
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
	if err := smgRplcMstrRPC.Call(utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated

	//check if the session was createad as active session on master
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
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

	//check if the session was created as passive session on slave
	var pSessions []*ExternalSession
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions,
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

func testSessionSRplUpdate(t *testing.T) {
	//update the session on slave so the session should became active
	usage := time.Minute
	argsUpdate := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSRplUpdate",
			Event: map[string]interface{}{
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
	if err := smgRplcSlvRPC.Call(utils.SessionSv1UpdateSession,
		argsUpdate, &updtRpl); err != nil {
		t.Error(err)
	}
	if updtRpl.MaxUsage == nil || *updtRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, updtRpl.MaxUsage)
	}

	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ExternalSession
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions,
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
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter),
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	// Master should not longer have activeSession
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(aSessions), utils.ToJSON(aSessions))
	}

	cgrID := GetSetCGRID(engine.NewMapEvent(argsUpdate.Event))
	// Make sure session was replicated
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &pSessions); err != nil {
		t.Error(err)
	} else if len(pSessions) != 1 {
		t.Errorf("PassiveSessions: %+v", pSessions)
	} else if pSessions[0].CGRID != cgrID {
		t.Errorf("PassiveSession: %+v", pSessions[0])
	} else if pSessions[0].Usage != 150*time.Second {
		t.Errorf("Expecting : %+v, received: %+v", 150*time.Second, pSessions[0].Usage)
	}
}

func testSessionSRplTerminate(t *testing.T) {
	args := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSessionSRplTerminate",
			Event: map[string]interface{}{
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
	if err := smgRplcMstrRPC.Call(utils.SessionSv1TerminateSession, args, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	var aSessions []*ExternalSession
	//check if the session was terminated on master
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions,
		utils.SessionFilter{
			Filters: []string{
				fmt.Sprintf("*string:~*req.%s:%s", utils.OriginID, "123451"),
			},
		}, &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(aSessions), utils.ToIJSON(aSessions))
	}
	//check if the session was terminated on slave
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(aSessions)=%v , session : %+v", err, len(aSessions), utils.ToIJSON(aSessions))
	}
	// check to don't have passive session on master and slave
	var pSessions []*ExternalSession
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter),
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(pSessions)=%v , session : %+v", err, len(pSessions), utils.ToIJSON(pSessions))
	}
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter),
		&pSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Errorf("Error: %v with len(pSessions)=%v , session : %+v", err, len(pSessions), utils.ToIJSON(pSessions))
	}
}

func testSessionSRplManualReplicate(t *testing.T) {
	masterProc, err := engine.StopStartEngine(smgRplcMasterCfgPath, *waitRater)
	if err != nil { // Kill both and start Master
		t.Fatal(err)
	}
	if smgRplcMstrRPC, err = newRPCClient(smgRplcMasterCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	// create two sessions
	argsInit1 := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth",
			Event: map[string]interface{}{
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
				utils.Usage:        time.Minute + 30*time.Second,
			},
		},
	}

	argsInit2 := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItAuth2",
			Event: map[string]interface{}{
				utils.EventName:    "TEST_EVENT",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123481",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1005",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        time.Minute + 30*time.Second,
			},
		},
	}

	for _, args := range []*V1InitSessionArgs{argsInit1, argsInit2} {
		var initRpl *V1InitSessionReply
		if err := smgRplcMstrRPC.Call(utils.SessionSv1InitiateSession, args, &initRpl); err != nil {
			t.Error(err)
		}
		if initRpl.MaxUsage == nil || *initRpl.MaxUsage != 90*time.Second {
			t.Error("Bad max usage: ", initRpl.MaxUsage)
		}
	}
	//verify if the sessions was created on master and are active
	var aSessions []*ExternalSession
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToJSON(aSessions))
	} else if aSessions[0].Usage != 90*time.Second && aSessions[1].Usage != 90*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	// Start slave, should not have any active session at beginning
	slave, err := engine.StartEngine(smgRplcSlaveCfgPath, *waitRater)
	if err != nil {
		t.Fatal(err)
	}
	if err := slave.Process.Kill(); err != nil { // restart the slave
		t.Error(err)
	}
	if _, err := engine.StartEngine(smgRplcSlaveCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if smgRplcSlvRPC, err = newRPCClient(smgRplcSlaveCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	// when we start slave after master we expect to don't have sessions
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err == nil ||
		err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	argsRepl := ArgsReplicateSessions{
		ConnIDs: []string{"rplConn"},
	}
	//replicate manually the session from master to slave
	var repply string
	if err := smgRplcMstrRPC.Call(utils.SessionSv1ReplicateSessions, &argsRepl, &repply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 90*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
	// kill master
	if err := masterProc.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process, error: %v", err.Error())
	}
	var status map[string]interface{}
	if err := smgRplcMstrRPC.Call(utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err == nil { // master should not longer be reachable
		t.Error(err, status)
	}
	if err := smgRplcSlvRPC.Call(utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // slave should be still operational
		t.Error(err)
	}
	// start master
	if _, err := engine.StartEngine(smgRplcMasterCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
	if smgRplcMstrRPC, err = newRPCClient(smgRplcMasterCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	// Master should have no session active/passive
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	// recover passive sessions from slave
	argsRepl = ArgsReplicateSessions{
		Passive: true,
		ConnIDs: []string{"rplConn"},
	}
	if err := smgRplcSlvRPC.Call(utils.SessionSv1ReplicateSessions, &argsRepl, &repply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond) // Wait for the sessions to be populated
	// Master should have no session active/passive
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err, aSessions)
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", aSessions)
	} else if aSessions[0].Usage != 90*time.Second {
		t.Errorf("Received usage: %v", aSessions[0].Usage)
	}
}

func testSessionSRplActivateSessions(t *testing.T) {
	var aSessions []*ExternalSession
	var reply string
	// Activate first session (with ID: ede927f8e42318a8db02c0f74adc2d9e16770339)
	args := &utils.SessionIDsWithAPIOpts{
		IDs: []string{"ede927f8e42318a8db02c0f74adc2d9e16770339"},
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1ActivateSessions, args, &reply); err != nil {
		t.Error(err)
	}
	// Check the sessions on master engine (at this point should have one active and one passive session)
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Expecting: 1 session, received: %+v sessions", len(aSessions))
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Expecting: 1 session, received: %+v sessions", len(aSessions))
	}
	// Check the sessions on slave engine (at this point should have one active and one passive session)
	// if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
	// 	t.Error(err)
	// } else if len(aSessions) != 1 {
	// 	t.Errorf("Expecting: 1 session, received: %+v sessions", len(aSessions)) //received 2
	// }
	// if err := smgRplcSlvRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
	// 	t.Error(err) //received not found
	// } else if len(aSessions) != 1 {
	// 	t.Errorf("Expecting: 1 session, received: %+v sessions", len(aSessions))
	// }
	//activate the second session (with ID: 3b0417028f8cefc0e02ddbd37a6dda6fbef4f5e0)
	args = &utils.SessionIDsWithAPIOpts{
		IDs: []string{"3b0417028f8cefc0e02ddbd37a6dda6fbef4f5e0"},
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1ActivateSessions, args, &reply); err != nil {
		t.Error(err)
	}
	//Check the sessions on master engine (2 active, 0 passive)
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Expecting: 2 session, received: %+v sessions", len(aSessions))
	}
	if err := smgRplcMstrRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	//check the number of passive sessions on slave engine
	if err := smgRplcSlvRPC.Call(utils.SessionSv1GetPassiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Expecting: 2 session, received: %+v sessions", len(aSessions))
	}
}

func testSessionSRplStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
*/
