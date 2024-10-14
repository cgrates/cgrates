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
	"os/exec"
	"path"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

var (
	smgRplcCfgPath1, smgRplcCfgPath2 string
	smgRplcCfgDIR1, smgRplcCfgDIR2   string
	smgRplCfg1, smgRplCfg2           *config.CGRConfig
	smgRplcRPC1, smgRplcRPC2         *birpc.Client
	testEngine1, testEngine2         *exec.Cmd
	sTestsSession1                   = []func(t *testing.T){
		testSessionSRplcInitCfg,
		testSessionSRplcResetDB,
		testSessionSRplcStartEngine,
		testSessionSRplcApierRpcConn,
		testSessionSRplcApierGetActiveSessionsNotFound,
		testSessionSRplcApierSetChargerS,
		testSessionSRplcApierGetInitateSessions,
		testSessionSRplcApierGetActiveSessions,
		testSessionSRplcApierGetPassiveSessions,
		testSessionSRplcApierStopSession2,
		testSessionSRplcApierGetPassiveSessionsAfterStop,
		testSessionSRplcStopCgrEngine,
	}
)

func TestSessionSRplcGracefulShutdown(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		smgRplcCfgDIR1 = "rplcTestGracefulShutdown1_internal"
		smgRplcCfgDIR2 = "rplcTestGracefulShutdown2_internal"
	case utils.MetaMySQL:
		smgRplcCfgDIR1 = "rplcTestGracefulShutdown1_mysql"
		smgRplcCfgDIR2 = "rplcTestGracefulShutdown2_mysql"
	case utils.MetaMongo:
		smgRplcCfgDIR1 = "rplcTestGracefulShutdown1_mongo"
		smgRplcCfgDIR2 = "rplcTestGracefulShutdown2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest1 := range sTestsSession1 {
		t.Run(*utils.DBType, stest1)
	}
}

// Init Config
func testSessionSRplcInitCfg(t *testing.T) {
	var err error
	smgRplcCfgPath1 = path.Join(*utils.DataDir, "conf", "samples", "sessions_replication", smgRplcCfgDIR1)
	if smgRplCfg1, err = config.NewCGRConfigFromPath(smgRplcCfgPath1); err != nil {
		t.Fatal(err)
	}
	smgRplcCfgPath2 = path.Join(*utils.DataDir, "conf", "samples", "sessions_replication", smgRplcCfgDIR2)
	if smgRplCfg2, err = config.NewCGRConfigFromPath(smgRplcCfgPath2); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSessionSRplcResetDB(t *testing.T) {
	if err := engine.InitDataDb(smgRplCfg1); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(smgRplCfg1); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSessionSRplcStartEngine(t *testing.T) {
	var err error
	if _, err = engine.StopStartEngine(smgRplcCfgPath1, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	if testEngine1, err = engine.StartEngine(smgRplcCfgPath2, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}

}

// Connect rpc client to rater
func testSessionSRplcApierRpcConn(t *testing.T) {
	smgRplcRPC1 = engine.NewRPCClient(t, smgRplCfg1.ListenCfg())
	smgRplcRPC2 = engine.NewRPCClient(t, smgRplCfg2.ListenCfg())
}

func testSessionSRplcApierGetActiveSessionsNotFound(t *testing.T) {
	aSessions1 := make([]*sessions.ExternalSession, 0)
	expected := "NOT_FOUND"
	if err := smgRplcRPC1.Call(context.Background(), utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions1); err == nil || err.Error() != expected {
		t.Error(err)
	}
	aSessions2 := make([]*sessions.ExternalSession, 0)
	if err := smgRplcRPC2.Call(context.Background(), utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions2); err == nil || err.Error() != expected {
		t.Error(err)
	}
}

func testSessionSRplcApierSetChargerS(t *testing.T) {
	chargerProfile1 := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result1 string
	if err := smgRplcRPC1.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile1, &result1); err != nil {
		t.Error(err)
	} else if result1 != utils.OK {
		t.Error("Unexpected reply returned", result1)
	}

	chargerProfile2 := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result2 string
	if err := smgRplcRPC2.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile2, &result2); err != nil {
		t.Error(err)
	} else if result2 != utils.OK {
		t.Error("Unexpected reply returned", result2)
	}
}

func testSessionSRplcApierGetInitateSessions(t *testing.T) {
	args := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestSSv1ItInitiateSession",
			Event: map[string]any{
				utils.Tenant:      "cgrates.org",
				utils.RequestType: utils.MetaNone,
				utils.CGRID:       "testSessionRplCGRID",
				utils.OriginID:    "testSessionRplORIGINID",
			},
		},
	}
	var rply sessions.V1InitSessionReply
	if err := smgRplcRPC2.Call(context.Background(), utils.SessionSv1InitiateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
}

func testSessionSRplcApierGetActiveSessions(t *testing.T) {
	expected := []*sessions.ExternalSession{
		{
			CGRID:         "testSessionRplCGRID",
			RunID:         "*default",
			ToR:           "",
			OriginID:      "testSessionRplORIGINID",
			OriginHost:    "",
			Source:        "SessionS_",
			RequestType:   utils.MetaNone,
			Tenant:        "cgrates.org",
			Category:      "",
			Account:       "",
			Subject:       "",
			Destination:   "",
			SetupTime:     time.Time{},
			AnswerTime:    time.Time{},
			Usage:         0,
			ExtraFields:   map[string]string{},
			NodeID:        "MasterReplication",
			LoopIndex:     0,
			DurationIndex: 0,
			MaxRate:       0,
			MaxRateUnit:   0,
			MaxCostSoFar:  0,
			DebitInterval: 0,
			NextAutoDebit: time.Time{},
		},
	}
	aSessions2 := make([]*sessions.ExternalSession, 0)
	if err := smgRplcRPC2.Call(context.Background(), utils.SessionSv1GetActiveSessions, &utils.SessionFilter{}, &aSessions2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&aSessions2, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(&aSessions2), utils.ToJSON(&expected))

	}
}

func testSessionSRplcApierGetPassiveSessions(t *testing.T) {
	expected := []*sessions.ExternalSession{
		{
			CGRID:         "testSessionRplCGRID",
			RunID:         "*default",
			ToR:           "",
			OriginID:      "testSessionRplORIGINID",
			OriginHost:    "",
			Source:        "SessionS_",
			RequestType:   utils.MetaNone,
			Tenant:        "cgrates.org",
			Category:      "",
			Account:       "",
			Subject:       "",
			Destination:   "",
			SetupTime:     time.Time{},
			AnswerTime:    time.Time{},
			Usage:         0,
			ExtraFields:   map[string]string{},
			NodeID:        "MasterReplication",
			LoopIndex:     0,
			DurationIndex: 0,
			MaxRate:       0,
			MaxRateUnit:   0,
			MaxCostSoFar:  0,
			DebitInterval: 0,
			NextAutoDebit: time.Time{},
		},
	}
	aSessions2 := make([]*sessions.ExternalSession, 0)
	if err := smgRplcRPC1.Call(context.Background(), utils.SessionSv1GetPassiveSessions, &utils.SessionFilter{}, &aSessions2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&aSessions2, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(&aSessions2), utils.ToJSON(&expected))

	}
}

func testSessionSRplcApierStopSession2(t *testing.T) {
	err := testEngine1.Process.Signal(syscall.SIGTERM)
	if err != nil {
		t.Error(err)
	}
	err = testEngine1.Wait()
	if err != nil {
		t.Error(err)
	}
}

func testSessionSRplcApierGetPassiveSessionsAfterStop(t *testing.T) {
	expected := []*sessions.ExternalSession{
		{
			CGRID:         "testSessionRplCGRID",
			RunID:         "*default",
			ToR:           "",
			OriginID:      "testSessionRplORIGINID",
			OriginHost:    "",
			Source:        "SessionS_",
			RequestType:   utils.MetaNone,
			Tenant:        "cgrates.org",
			Category:      "",
			Account:       "",
			Subject:       "",
			Destination:   "",
			SetupTime:     time.Time{},
			AnswerTime:    time.Time{},
			Usage:         0,
			ExtraFields:   map[string]string{},
			NodeID:        "MasterReplication",
			LoopIndex:     0,
			DurationIndex: 0,
			MaxRate:       0,
			MaxRateUnit:   0,
			MaxCostSoFar:  0,
			DebitInterval: 0,
			NextAutoDebit: time.Time{},
		},
	}
	aSessions2 := make([]*sessions.ExternalSession, 0)
	if err := smgRplcRPC1.Call(context.Background(), utils.SessionSv1GetPassiveSessions, &utils.SessionFilter{}, &aSessions2); err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(&aSessions2, &expected) {
		t.Errorf("\nExpected <%+v>, \nReceived <%+v>", utils.ToJSON(&aSessions2), utils.ToJSON(&expected))

	}
}

func testSessionSRplcStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
