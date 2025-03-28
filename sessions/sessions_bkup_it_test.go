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
	"fmt"
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
	sBkupCfgPath string
	sBkupCfgDIR  string
	sBkupCfg     *config.CGRConfig
	sBkupRPC     *birpc.Client
	dataDB       engine.DataDB
	updatedAt    time.Time

	SessionsBkupTests = []func(t *testing.T){
		testSessionSBkupInitCfg,
		testSessionSBkupResetDB,
		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupTPFromFolder,
		testSessionSBkupInitiate,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCheckRestored1,
		testSessionSBkupTerminate1,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCheckRestored2,
		testSessionSBkupTerminate2,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCheckRestored3,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCallBackup1,
		testSessionSBkupInitiate,
		testSessionSBkupCallBackup2,
		testSessionSBkupGetBackedupSessions,
		testSessionSBkupUpdateTerminate,
		testSessionSBkupCallBackup3,
		testSessionSBkupCheckUpdatedAt,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCheckUpdatedNotExpired,
		testSessionSBkupStopCgrEngine,

		testSessionSBkupStartEngine,
		testSessionSBkupApierRpcConn,
		testSessionSBkupCheckUpdatedExpired,
		testSessionSBkupStopCgrEngine,
	}
)

func TestSessionsBkup(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		sBkupCfgDIR = "sessions_backup_mysql"
	case utils.MetaMongo:
		sBkupCfgDIR = "sessions_backup_mongo"
	case utils.MetaPostgres:
		sBkupCfgDIR = "sessions_backup_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range SessionsBkupTests {
		t.Run(*utils.DBType, stest)
	}
}

func testSessionSBkupInitCfg(t *testing.T) {
	sBkupCfgPath = path.Join(*utils.DataDir, "conf", "samples", sBkupCfgDIR)
	if sBkupCfg, err = config.NewCGRConfigFromPath(sBkupCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSBkupResetDB(t *testing.T) {
	if *utils.DBType == utils.MetaInternal {
		if err := engine.PreInitDataDb(sBkupCfg); err != nil {
			t.Fatal(err)
		}
		if err := engine.PreInitStorDb(sBkupCfg); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := engine.InitDataDb(sBkupCfg); err != nil {
			t.Fatal(err)
		}
		if err := engine.InitStorDb(sBkupCfg); err != nil {
			t.Fatal(err)
		}
	}
}

// Start CGR Engine
func testSessionSBkupStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(sBkupCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionSBkupApierRpcConn(t *testing.T) {
	if sBkupRPC, err = newRPCClient(sBkupCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

// Load the tariff plan, creating accounts and their balances
func testSessionSBkupTPFromFolder(t *testing.T) {
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	var loadInst utils.LoadInstance
	if err := sBkupRPC.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
}

func testSessionSBkupInitiate(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	usage := time.Duration(90 * time.Second)
	argsInit := &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestBkupSessionExpiresAfter3s",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT1",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123450",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	var initRpl V1InitSessionReply
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated
	// Delay further 4 seconds to make the session unrestorable (session expiries in 4s in configs)
	time.Sleep(4 * time.Second)

	//check if the session was createad as active session
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	} else if aSessions[0].Usage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, aSessions[0].Usage)
	}

	usage = time.Duration(120 * time.Second)
	argsInit = &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestBkupSession2",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT2",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123452",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 2, 0, time.UTC),
				utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 12, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated

	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
	for _, extSess := range aSessions {
		if extSess.OriginID == "123452" {
			if extSess.Usage != usage {
				t.Errorf("OriginID: <123452>, Expecting : %+v, received: %+v", usage, extSess.Usage)
			}
		}
	}

	usage = time.Duration(150 * time.Second)
	argsInit = &V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestBkupSession1",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT1",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123451",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 3, 0, time.UTC),
				utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 13, 0, time.UTC),
				utils.Usage:        usage,
			},
		},
	}

	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		argsInit, &initRpl); err != nil {
		t.Error(err)
	}
	//compare the value
	if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, initRpl.MaxUsage)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Wait for the sessions to be populated

	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 3 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
	for _, extSess := range aSessions {
		if extSess.OriginID == "123451" {
			if extSess.Usage != usage {
				t.Errorf("OriginID: <123451>, Expecting : %+v, received: %+v", usage, extSess.Usage)
			}
		}
	}
}

func testSessionSBkupCheckRestored1(t *testing.T) {
	var aSessions []*ExternalSession
	usage1 := time.Duration(120 * time.Second)
	usage2 := time.Duration(150 * time.Second)
	//check if the sessions were restored
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 2 {
		t.Fatalf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
	for _, extSess := range aSessions {
		switch {
		case extSess.OriginID == "123451":
			if extSess.Usage != usage2 {
				t.Errorf("OriginID: <123451>, Expecting : %+v, received: %+v", usage2, extSess.Usage)
			}
		case extSess.OriginID == "123452":
			if extSess.Usage != usage1 {
				t.Errorf("OriginID: <123452>, Expecting : %+v, received: %+v", usage1, extSess.Usage)
			}
		default:
			t.Fatalf("Unexpected OriginID <%v> present in the session", extSess.OriginID)
		}
	}
}

func testSessionSBkupTerminate1(t *testing.T) {
	var replyTerminate string
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBkupSession1",
				Event: map[string]any{
					utils.EventName:    "TEST_EVENT1",
					utils.Tenant:       "cgrates.org",
					utils.OriginID:     "123451",
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1004",
					utils.Category:     "call",
					utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 3, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 13, 0, time.UTC),
				},
			},
		}, &replyTerminate); err != nil {
		t.Error(err)
	}
	if replyTerminate != utils.OK {
		t.Errorf("Expected reply <OK>, received <%+v>", replyTerminate)
	}
}

func testSessionSBkupCheckRestored2(t *testing.T) {
	var aSessions []*ExternalSession
	usage1 := time.Duration(120 * time.Second)
	//check if the sessions were restored
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Fatal(err)
	} else if len(aSessions) != 1 {
		t.Fatalf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
	if aSessions[0].OriginID == "123452" {
		if aSessions[0].Usage != usage1 {
			t.Errorf("OriginID: <123452>, Expecting : %+v, received: %+v", usage1, aSessions[0].Usage)
		}
	}
}

func testSessionSBkupTerminate2(t *testing.T) {
	var replyTerminate string
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		&V1TerminateSessionArgs{
			TerminateSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "TestBkupSession2",
				Event: map[string]any{
					utils.EventName:    "TEST_EVENT2",
					utils.Tenant:       "cgrates.org",
					utils.OriginID:     "123452",
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1004",
					utils.Category:     "call",
					utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 2, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 12, 0, time.UTC),
				},
			},
		}, &replyTerminate); err != nil {
		t.Error(err)
	}
	if replyTerminate != utils.OK {
		t.Errorf("Expected reply <OK>, received <%+v>", replyTerminate)
	}
}

func testSessionSBkupCheckRestored3(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if len(aSessions) != 0 {
		t.Errorf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
}

func testSessionSBkupCallBackup1(t *testing.T) {
	var sessionsBackedup int
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1BackupActiveSessions,
		utils.EmptyString, &sessionsBackedup); err != nil {
		t.Error(err)
	} else if sessionsBackedup != 0 {
		t.Errorf("Expected 0 backedup sessions. Backed up: %+v", sessionsBackedup)
	}
}

func testSessionSBkupCallBackup2(t *testing.T) {
	var sessionsBackedup int
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1BackupActiveSessions,
		utils.EmptyString, &sessionsBackedup); err != nil {
		t.Fatal(err)
	} else if sessionsBackedup != 3 {
		t.Errorf("Expected 3 backedup sessions. Backed up: %+v", sessionsBackedup)
	}
}

func testSessionSBkupGetBackedupSessions(t *testing.T) {

	if *utils.DBType == utils.MetaMySQL || *utils.DBType == utils.MetaPostgres {
		dataDB, err = engine.NewRedisStorage(
			fmt.Sprintf("%s:%s", sBkupCfg.DataDbCfg().Host, sBkupCfg.DataDbCfg().Port),
			10, sBkupCfg.DataDbCfg().User, sBkupCfg.DataDbCfg().Password, sBkupCfg.GeneralCfg().DBDataEncoding,
			10, 20, "", false, 5*time.Second, 0, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
	}
	if *utils.DBType == utils.MetaMongo {
		dataDB, err = engine.NewMongoStorage("mongodb", sBkupCfg.DataDbCfg().Host,
			sBkupCfg.DataDbCfg().Port, sBkupCfg.DataDbCfg().Name,
			sBkupCfg.DataDbCfg().User, sBkupCfg.DataDbCfg().Password,
			sBkupCfg.GeneralCfg().DBDataEncoding,
			utils.StorDB, nil, 10*time.Second)
		if err != nil {
			t.Fatal(err)
		}
	}

	var getBackSess []*Session
	storedSessions, err := dataDB.GetSessionsBackupDrv(sBkupCfg.GeneralCfg().NodeID,
		sBkupCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		t.Error(err)
	}
	for _, storSess := range storedSessions {
		sess := newSessionFromStoredSession(storSess)
		getBackSess = append(getBackSess, sess)
	}

	if len(getBackSess) != 3 { // even though one of them expired, we are not restoring them currently so in db there should be 3 sessions stored
		t.Fatalf("Expected 3 sessions stored, received %v", len(getBackSess))
	}
	for _, oneSess := range getBackSess {
		switch {
		case oneSess.ResourceID == "123450":
			if oneSess.totalUsage() != time.Duration(90*time.Second) {
				t.Errorf("Expected <%v>, received <%v>", time.Duration(90*time.Second), oneSess.totalUsage())
			}
		case oneSess.ResourceID == "123451":
			if oneSess.totalUsage() != time.Duration(150*time.Second) {
				t.Errorf("Expected <%v>, received <%v>", time.Duration(150*time.Second), oneSess.totalUsage())
			}
		case oneSess.ResourceID == "123452":
			if oneSess.totalUsage() != time.Duration(120*time.Second) {
				t.Errorf("Expected <%v>, received <%v>", time.Duration(120*time.Second), oneSess.totalUsage())
			}
			updatedAt = oneSess.UpdatedAt
		}
	}

}

func testSessionSBkupUpdateTerminate(t *testing.T) {
	updatedUsage := time.Duration(200 * time.Second)
	var upRply V1UpdateSessionReply
	upArgs := &V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestBkupSession2",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT2",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123452",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 2, 0, time.UTC),
				utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 12, 0, time.UTC),
				utils.Usage:        updatedUsage,
			},
		},
	}
	if err = sBkupRPC.Call(context.Background(), utils.SessionSv1UpdateSession, upArgs, &upRply); err != nil {
		t.Error(err)
	} else if *upRply.MaxUsage != updatedUsage {
		t.Errorf("Expected <%+v>, Received <%+v>", updatedUsage, *upRply.MaxUsage)
	}

	args := &V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestBkupSession1",
			Event: map[string]any{
				utils.EventName:    "TEST_EVENT1",
				utils.Tenant:       "cgrates.org",
				utils.OriginID:     "123451",
				utils.ToR:          utils.MetaVoice,
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1004",
				utils.Category:     "call",
				utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 3, 0, time.UTC),
				utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 13, 0, time.UTC),
				utils.Usage:        time.Duration(160 * time.Second),
			},
		},
	}
	var rply string
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
}

func testSessionSBkupCallBackup3(t *testing.T) {
	var sessionsBackedup int
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1BackupActiveSessions,
		utils.EmptyString, &sessionsBackedup); err != nil {
		t.Fatal(err)
	} else if sessionsBackedup != 2 { // should've backed up 2 since we terminated one of them
		t.Errorf("Expected 2 backedup sessions. Backed up: %+v", sessionsBackedup)
	}
}

func testSessionSBkupCheckUpdatedAt(t *testing.T) {
	var getBackSess []*Session
	storedSessions, err := dataDB.GetSessionsBackupDrv(sBkupCfg.GeneralCfg().NodeID,
		sBkupCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		t.Error(err)
	}
	for _, storSess := range storedSessions {
		sess := newSessionFromStoredSession(storSess)
		getBackSess = append(getBackSess, sess)
	}

	if len(getBackSess) != 2 { // even though one of them expired, we are not restoring them currently so in db there should be 2 sessions stored
		t.Fatalf("Expected 2 sessions stored, received %v", len(getBackSess))
	}
	for _, oneSess := range getBackSess {
		switch {
		case oneSess.ResourceID == "123450":
			if oneSess.totalUsage() != time.Duration(90*time.Second) {
				t.Errorf("Expected <%v>, received <%v>", time.Duration(90*time.Second), oneSess.totalUsage())
			}
		case oneSess.ResourceID == "123452":
			if oneSess.totalUsage() != time.Duration(320*time.Second) { // usage should be updated to 320 seconds
				t.Errorf("Expected <%v>, received <%v>", time.Duration(320*time.Second), oneSess.totalUsage())
			}
			if oneSess.UpdatedAt == updatedAt {
				t.Errorf("Expected UpdatedAt field to be changed on update. Received the same time as before <%+v>", oneSess.UpdatedAt)
			}
		}
	}
}

func testSessionSBkupCheckUpdatedNotExpired(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Fatalf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
	if aSessions[0].OriginID != "123452" {
		t.Errorf("Expected backed up session source 123452, received <%+v>", aSessions[0].OriginID)
	}
	time.Sleep(4 * time.Second) // Wait for updated session to expire
}

func testSessionSBkupCheckUpdatedExpired(t *testing.T) {
	var aSessions []*ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	} else if len(aSessions) != 0 {
		t.Fatalf("Unexpected number of sessions received: %+v", utils.ToIJSON(aSessions))
	}
}

func testSessionSBkupStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
