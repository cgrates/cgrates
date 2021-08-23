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
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	ses2CfgDir  string
	ses2CfgPath string
	ses2Cfg     *config.CGRConfig
	ses2RPC     *rpc.Client

	ses2Tests = []func(t *testing.T){
		testSes2ItLoadConfig,
		testSes2ItResetDataDB,
		testSes2ItResetStorDb,
		testSes2ItStartEngine,
		testSes2ItRPCConn,
		testSes2ItLoadFromFolder,
		testSes2ItInitSession,
		testSes2ItAsActiveSessions,
		testSes2ItStopCgrEngine,
	}
)

func TestSes2It(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		ses2CfgDir = "tutinternal"
	case utils.MetaMySQL:
		ses2CfgDir = "tutmysql"
	case utils.MetaMongo:
		ses2CfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range ses2Tests {
		t.Run(ses2CfgDir, stest)
	}
}

func testSes2ItLoadConfig(t *testing.T) {
	ses2CfgPath = path.Join(*dataDir, "conf", "samples", ses2CfgDir)
	if ses2Cfg, err = config.NewCGRConfigFromPath(ses2CfgPath); err != nil {
		t.Error(err)
	}
}

func testSes2ItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(ses2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ses2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ses2CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSes2ItRPCConn(t *testing.T) {
	var err error
	ses2RPC, err = newRPCClient(ses2Cfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSes2ItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := ses2RPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testSes2ItInitSession(t *testing.T) {
	// Set balance
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		BalanceType: utils.VOICE,
		Value:       float64(time.Hour),
		Balance: map[string]interface{}{
			utils.ID: "TestDynamicDebitBalance",
		},
	}
	var reply string
	if err := ses2RPC.Call(utils.APIerSv2SetBalance,
		attrSetBalance, &reply); err != nil {
		t.Fatal(err)
	}

	// Init session
	initArgs := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     utils.UUIDSha1Prefix(),
			Event: map[string]interface{}{
				utils.EVENT_NAME:  "TEST_EVENT",
				utils.OriginID:    utils.UUIDSha1Prefix(),
				utils.ToR:         utils.VOICE,
				utils.Category:    "call",
				utils.Tenant:      "cgrates.org",
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.RequestType: utils.META_PREPAID,
				utils.AnswerTime:  time.Date(2016, time.January, 5, 18, 31, 05, 0, time.UTC),
			},
		},
	}
	var initRpl *sessions.V1InitSessionReply
	if err := ses2RPC.Call(utils.SessionSv1InitiateSession,
		initArgs, &initRpl); err != nil {
		t.Fatal(err)
	}

}

func testSes2ItAsActiveSessions(t *testing.T) {
	var count int
	if err := ses2RPC.Call(utils.SessionSv1GetActiveSessionsCount, utils.SessionFilter{
		Filters: []string{"*string:~*req.Account:1001"},
	}, &count); err != nil {
		t.Fatal(err)
	} else if count != 2 { // 2 chargers
		t.Errorf("Expeced 2 session received %v session(s)", count)
	}
	if err := ses2RPC.Call(utils.SessionSv1GetActiveSessionsCount, utils.SessionFilter{
		Filters: []string{"*string:~*req.Account:1002"},
	}, &count); err != nil {
		t.Fatal(err)
	} else if count != 0 {
		t.Errorf("Expeced 0 session received %v session(s)", count)
	}
}

func testSes2ItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
