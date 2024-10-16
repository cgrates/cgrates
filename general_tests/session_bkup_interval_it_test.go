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
	"fmt"
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
	sBkupCfgPath string
	sBkupCfgDIR  string
	sBkupCfg     *config.CGRConfig
	sBkupRPC     *birpc.Client
	dDB          engine.DataDB

	SessionsBkupIntrvlTests = []func(t *testing.T){
		testSessionSBkupIntrvlInitCfg,
		testSessionSBkupIntrvlResetDB,
		testSessionSBkupIntrvlStartEngine,
		testSessionSBkupIntrvlApierRpcConn,
		testSessionSBkupIntrvlLoadTP,
		testSessionSBkupIntrvlInitiate,
		testSessionSBkupIntrvlConcurrentAPIWithInterval,
		testSessionSBkupIntrvlGetBackedupSessions1,
		testSessionSBkupIntrvlStopCgrEngine,

		testSessionSBkupIntrvlStartEngine,
		testSessionSBkupIntrvlApierRpcConn,
		testSessionSBkupIntrvlGetActiveSessionsTerminate,
		testSessionSBkupIntrvlGetBackedupSessions2,
		testSessionSBkupIntrvlStopCgrEngine,

		testSessionSBkupIntrvlStartEngine,
		testSessionSBkupIntrvlApierRpcConn,
		testSessionSBkupIntrvlGetActiveSessions0,
		testSessionSBkupIntrvlStopCgrEngine,
	}
)

func TestSessionsBkupIntrvl(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		sBkupCfgDIR = "sessions_backup_interval_mysql"
	case utils.MetaMongo:
		sBkupCfgDIR = "sessions_backup_interval_mongo"
	case utils.MetaPostgres:
		sBkupCfgDIR = "sessions_backup_interval_postgres"
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range SessionsBkupIntrvlTests {
		t.Run(*utils.DBType, stest)
	}
}

func testSessionSBkupIntrvlInitCfg(t *testing.T) {
	var err error
	sBkupCfgPath = path.Join(*utils.DataDir, "conf", "samples", sBkupCfgDIR)
	if sBkupCfg, err = config.NewCGRConfigFromPath(sBkupCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSBkupIntrvlResetDB(t *testing.T) {
	if err := engine.InitDataDb(sBkupCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(sBkupCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionSBkupIntrvlStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(sBkupCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSessionSBkupIntrvlApierRpcConn(t *testing.T) {
	sBkupRPC = engine.NewRPCClient(t, sBkupCfg.ListenCfg())
}

// Load the tariff plan, creating accounts and their balances
func testSessionSBkupIntrvlLoadTP(t *testing.T) {
	tpFiles := map[string]string{
		utils.AccountActionsCsv: `#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,`,
		utils.ActionPlansCsv: `#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10`,
		utils.ActionsCsv: `#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_RST_10,*topup_reset,,,test,*monetary,,*any,,,*unlimited,,100000000000000,10,false,false,10`,
		utils.ChargersCsv: `#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0`,
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY_20CNT,*any,RT_20CNT,*up,4,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0,1,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY_20CNT,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,*any,2014-01-14T00:00:00Z,RP_ANY,`,
	}

	engine.LoadCSVs(t, sBkupRPC, "", tpFiles)

	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
}

func testSessionSBkupIntrvlInitiate(t *testing.T) {
	var aSessions []*sessions.ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetPassiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}

	for i := 1; i <= 500; i++ {
		strI := fmt.Sprint(i)
		usage := time.Duration(i) * time.Second
		argsInit := &sessions.V1InitSessionArgs{
			InitSession: true,
			CGREvent: &utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "Test_" + strI,
				Event: map[string]any{
					utils.EventName:    "EVENT_" + strI,
					utils.Tenant:       "cgrates.org",
					utils.OriginID:     strI,
					utils.ToR:          utils.MetaVoice,
					utils.RequestType:  utils.MetaPrepaid,
					utils.AccountField: "1001",
					utils.Subject:      "1001",
					utils.Destination:  "1002",
					utils.Category:     "call",
					utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 0, 0, time.UTC),
					utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 10, 0, time.UTC),
					utils.Usage:        usage,
				},
			},
		}

		var initRpl sessions.V1InitSessionReply
		if err := sBkupRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
			argsInit, &initRpl); err != nil {
			t.Error(err)
		}
		if initRpl.MaxUsage == nil || *initRpl.MaxUsage != usage {
			t.Errorf("i <%v> Expecting : %+v, received: %+v", i, usage, initRpl.MaxUsage)
		}
	}

	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 500 {
		t.Errorf("Unexpected number of sessions received: <%+v>\n%+v", len(aSessions), utils.ToIJSON(aSessions))
	}
	for _, session := range aSessions {
		found := false
		for i := 1; i <= 500; i++ {
			strI := fmt.Sprint(i)
			if session.OriginID == strI && session.Usage == time.Duration(i)*time.Second && session.Source == "SessionS_EVENT_"+strI {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Session not found: %+v", session)
		}
	}
}

// try to reach concurrency by calling backup API multiple times while "backup_interval" runs in background
func testSessionSBkupIntrvlConcurrentAPIWithInterval(t *testing.T) {
	for i := 0; i <= 1000; i++ {
		var sessionsBackedup int
		if err := sBkupRPC.Call(context.Background(), utils.SessionSv1BackupActiveSessions,
			utils.EmptyString, &sessionsBackedup); err != nil {
			t.Fatal(err)
		} else if sessionsBackedup != 500 {
			t.Errorf("Expected 500 backedup sessions. Backed up: %+v", sessionsBackedup)
		}
	}
}

func testSessionSBkupIntrvlGetBackedupSessions1(t *testing.T) {
	var err error
	if *utils.DBType == utils.MetaMySQL || *utils.DBType == utils.MetaPostgres {
		dDB, err = engine.NewRedisStorage(
			fmt.Sprintf("%s:%s", sBkupCfg.DataDbCfg().Host, sBkupCfg.DataDbCfg().Port),
			10, sBkupCfg.DataDbCfg().User, sBkupCfg.DataDbCfg().Password, sBkupCfg.GeneralCfg().DBDataEncoding,
			10, 20, "", false, 5*time.Second, 0, 0, 0, 0, 150*time.Microsecond, 0, false, utils.EmptyString, utils.EmptyString, utils.EmptyString)
		if err != nil {
			t.Fatal("Could not connect to Redis", err.Error())
		}
	}
	if *utils.DBType == utils.MetaMongo {
		dDB, err = engine.NewMongoStorage("mongodb", sBkupCfg.DataDbCfg().Host,
			sBkupCfg.DataDbCfg().Port, sBkupCfg.DataDbCfg().Name,
			sBkupCfg.DataDbCfg().User, sBkupCfg.DataDbCfg().Password,
			sBkupCfg.GeneralCfg().DBDataEncoding,
			utils.DataDB, nil, 10*time.Second)
		if err != nil {
			t.Fatal(err)
		}
	}

	// wait for all sessions to be backedup, 2 intervals to make sure all sessions had time to be stored
	time.Sleep(1000 * time.Millisecond)
	storedSessions, err := dDB.GetSessionsBackupDrv(sBkupCfg.GeneralCfg().NodeID,
		sBkupCfg.GeneralCfg().DefaultTenant)
	if err != nil {
		t.Error(err)
	}

	if len(storedSessions) != 500 {
		t.Fatalf("Expected 500 sessions stored, received %v", len(storedSessions))
	}
	for _, oneSess := range storedSessions {
		var found bool
		for i := 1; i <= 500; i++ {
			strI := fmt.Sprint(i)
			if oneSess.ResourceID == strI && oneSess.SRuns[0].TotalUsage == time.Duration(i)*time.Second {
				found = true
				break
			}
		}
		if !found {
			for i := range oneSess.SRuns {
				t.Errorf("Session not found: <%+v>, SRun<%+v>", oneSess, oneSess.SRuns[i])
			}
		}
	}
}

func testSessionSBkupIntrvlGetActiveSessionsTerminate(t *testing.T) {
	var aSessions []*sessions.ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 500 {
		t.Errorf("Unexpected number of sessions received: <%+v>\n%+v", len(aSessions), utils.ToIJSON(aSessions))
	}
	for _, session := range aSessions {
		found := false
		for i := 1; i <= 500; i++ {
			strI := fmt.Sprint(i)
			if session.OriginID == strI && session.Usage == time.Duration(i)*time.Second && session.Source == "SessionS_EVENT_"+strI {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Session not found: %+v", session)
		}
	}

	var replyTerminate string
	for i := 1; i <= 500; i++ {
		strI := fmt.Sprint(i)
		if err := sBkupRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
			&sessions.V1TerminateSessionArgs{
				TerminateSession: true,
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "Test_" + strI,
					Event: map[string]any{
						utils.EventName:    "EVENT_" + strI,
						utils.Tenant:       "cgrates.org",
						utils.OriginID:     strI,
						utils.ToR:          utils.MetaVoice,
						utils.RequestType:  utils.MetaPrepaid,
						utils.AccountField: "1001",
						utils.Subject:      "1001",
						utils.Destination:  "1002",
						utils.Category:     "call",
						utils.SetupTime:    time.Date(2024, time.March, 7, 16, 60, 3, 0, time.UTC),
						utils.AnswerTime:   time.Date(2024, time.March, 7, 16, 60, 13, 0, time.UTC),
						utils.Usage:        time.Duration(2 * time.Second),
					},
				},
			}, &replyTerminate); err != nil {
			t.Fatal(err)
		} else if replyTerminate != utils.OK {
			t.Errorf("Expected reply <OK>, received <%+v>", replyTerminate)
		}
	}
	time.Sleep(1 * time.Second) // Wait for 2 500ms intervals so we are sure it removed all terminated sessions from dataDB
}
func testSessionSBkupIntrvlGetBackedupSessions2(t *testing.T) {
	storedSessions, err := dDB.GetSessionsBackupDrv(sBkupCfg.GeneralCfg().NodeID,
		sBkupCfg.GeneralCfg().DefaultTenant)
	if err != utils.ErrNoBackupFound {
		t.Error(err)
	}
	if len(storedSessions) != 0 { // Sessions terminated should instantly be removed from the backup
		t.Errorf("Expected 0 sessions in backup, received %v", len(storedSessions))
	}
}

func testSessionSBkupIntrvlGetActiveSessions0(t *testing.T) {
	var aSessions []*sessions.ExternalSession
	if err := sBkupRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions,
		new(utils.SessionFilter), &aSessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testSessionSBkupIntrvlStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
