//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"

	"github.com/cgrates/cgrates/sessions"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"
)

var smgRplcRPC1, smgRplcRPC2 *birpc.Client

func TestSessionSRplcGracefulShutdown(t *testing.T) {
	var smgRplcCfgDIR1, smgRplcCfgDIR2 string
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

	ng1 := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "sessions_replication", smgRplcCfgDIR1),
		GracefulShutdown: true,
	}
	smgRplcRPC1, _ = ng1.Run(t)
	ng2 := engine.TestEngine{
		ConfigPath:       path.Join(*utils.DataDir, "conf", "samples", "sessions_replication", smgRplcCfgDIR2),
		PreserveDataDB:   true,
		PreserveStorDB:   true,
		GracefulShutdown: true,
	}
	smgRplcRPC2, _ = ng2.Run(t)

	for _, test := range []func(t *testing.T){
		testSessionSRplcApierGetActiveSessionsNotFound,
		testSessionSRplcApierSetChargerS,
		testSessionSRplcApierGetInitateSessions,
		testSessionSRplcApierGetActiveSessions,
		testSessionSRplcApierGetPassiveSessions,
		func(t *testing.T) { ng2.Stop(t) },
		testSessionSRplcApierGetPassiveSessionsAfterStop,
	} {
		t.Run(*utils.DBType, test)
	}
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
