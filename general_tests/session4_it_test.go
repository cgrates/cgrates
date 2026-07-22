//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
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
	ses4CfgDir  string
	ses4CfgPath string
	ses4Cfg     *config.CGRConfig
	ses4RPC     *birpc.Client

	ses4Tests = []func(t *testing.T){
		testSes4ItLoadConfig,
		testSes4ItResetDataDB,
		testSes4ItResetStorDb,
		testSes4ItStartEngine,
		testSes4ItRPCConn,
		testSes4ItLoadFromFolder,

		testSes4SetAccount,
		testSes4CDRsProcessCDR,

		testSes4ItStopCgrEngine,
	}
)

func TestSes4ItSessions(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		ses4CfgDir = "sessions_internal"
	case utils.MetaMySQL:
		ses4CfgDir = "sessions_mysql"
	case utils.MetaMongo:
		ses4CfgDir = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range ses4Tests {
		t.Run(ses4CfgDir, stest)
	}
}

func testSes4ItLoadConfig(t *testing.T) {
	var err error
	ses4CfgPath = path.Join(*utils.DataDir, "conf", "samples", ses4CfgDir)
	if ses4Cfg, err = config.NewCGRConfigFromPath(ses4CfgPath); err != nil {
		t.Error(err)
	}
}

func testSes4ItResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(ses4Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes4ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ses4Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes4ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ses4CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSes4ItRPCConn(t *testing.T) {
	ses4RPC = engine.NewRPCClient(t, ses4Cfg.ListenCfg())
}

func testSes4ItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "testit")}
	if err := ses4RPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSes4SetAccount(t *testing.T) {
	var reply string
	attrs := &utils.AttrSetAccount{Tenant: "cgrates.org", Account: "dan7"}
	if err := ses4RPC.Call(context.Background(), utils.APIerSv1SetAccount, attrs, &reply); err != nil {
		t.Error("Got error on APIerSv1.SetAccount: ", err.Error())
	} else if reply != utils.OK {
		t.Errorf("Calling APIerSv1.SetAccount received: %s", reply)
	}
}

func testSes4CDRsProcessCDR(t *testing.T) {
	// proccess twice the postpaid event that doesn't cost
	// this reproduce the issue #2123:
	// rerate a free postpaid event in the CDRServer
	// will make the BalanceInfo nil and result in a panic
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, utils.MetaStore, "*routes:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testV2CDRsProcessCDR1",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testV2CDRsProcessCDR",
				utils.RequestType:  utils.MetaPostpaid,
				utils.Category:     "free",
				utils.AccountField: "dan7",
				utils.Subject:      "RP_FREE",
				utils.Destination:  "0775692",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
			},
		},
	}

	// Process and store the given CDR.
	var reply string
	if err := ses4RPC.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	// Process the CDR again, after adding the *rerate flag.
	args.Flags = append(args.Flags, utils.MetaRerate)
	if err := ses4RPC.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testSes4ItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
