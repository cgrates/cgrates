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
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesNoneReqTypeCfgDir  string
	sesNoneReqTypeCfgPath string
	sesNoneReqTypeCfg     *config.CGRConfig
	sesNoneReqTypeRPC     *birpc.Client

	sesNoneReqTypeTests = []func(t *testing.T){
		testSesNoneReqTypeItLoadConfig,
		testSesNoneReqTypeItResetDataDB,
		testSesNoneReqTypeItResetStorDb,
		testSesNoneReqTypeItStartEngine,
		testSesNoneReqTypeItRPCConn,

		testSesNoneReqTypeItAddChargerS,
		testSesNoneReqTypeItInit,

		testSesNoneReqTypeItStopEngine,
	}
)

func TestSesNoneReqTypeItSessions(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesNoneReqTypeCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesNoneReqTypeCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesNoneReqTypeCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesNoneReqTypeTests {
		t.Run(sesNoneReqTypeCfgDir, stest)
	}
}

func testSesNoneReqTypeItLoadConfig(t *testing.T) {
	var err error
	sesNoneReqTypeCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesNoneReqTypeCfgDir)
	if sesNoneReqTypeCfg, err = config.NewCGRConfigFromPath(sesNoneReqTypeCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesNoneReqTypeItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesNoneReqTypeCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesNoneReqTypeItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesNoneReqTypeCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesNoneReqTypeItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesNoneReqTypeCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesNoneReqTypeItRPCConn(t *testing.T) {
	sesNoneReqTypeRPC = engine.NewRPCClient(t, sesNoneReqTypeCfg.ListenCfg())
}

func testSesNoneReqTypeItAddChargerS(t *testing.T) {
	//add a default charger
	chargerProfile := &v1.ChargerWithAPIOpts{
		ChargerProfile: &engine.ChargerProfile{
			Tenant:       "cgrates.org",
			ID:           "Default",
			RunID:        utils.MetaDefault,
			AttributeIDs: []string{"*none"},
			Weight:       20,
		},
	}
	var result string
	if err := sesNoneReqTypeRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, chargerProfile, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testSesNoneReqTypeItInit(t *testing.T) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.CGRID:        "cgrID",
				utils.Category:     utils.Call,
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestReqNone",
				utils.RequestType:  utils.MetaNone,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        10 * time.Second,
			},
			APIOpts: map[string]any{
				utils.OptsDebitInterval: "0s",
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sesNoneReqTypeRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage != nil && *rply1.MaxUsage != 10*time.Second {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
}

func testSesNoneReqTypeItStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
