//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"flag"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	itTestMongoAtalas = flag.Bool("mongo_atlas", false, "Run the test with mongo atalas connection")
	tutorialCfgPath   string
	tutorialCfg       *config.CGRConfig
	tutorialRpc       *birpc.Client
	tutorialConfDIR   string //run tests for specific configuration
	tutorialDelay     int

	sTestsTutorials = []func(t *testing.T){
		testTutorialLoadConfig,
		testTutorialResetDB,
		testTutorialStartEngine,
		testTutorialRpcConn,
		testTutorialFromFolder,
		testTutorialGetCost,
		testTutorialStopEngine,
	}
)

// Test start here
func TestTutorialMongoAtlas(t *testing.T) {
	if !*itTestMongoAtalas {
		return
	}
	tutorialConfDIR = "mongoatlas"
	for _, stest := range sTestsTutorials {
		t.Run(tutorialConfDIR, stest)
	}
}

func TestTutorial(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		tutorialConfDIR = "tutinternal"
	case utils.MetaMySQL:
		tutorialConfDIR = "tutmysql"
	case utils.MetaMongo:
		tutorialConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsTutorials {
		t.Run(tutorialConfDIR, stest)
	}
}

func testTutorialLoadConfig(t *testing.T) {
	var err error
	tutorialCfgPath = path.Join(*utils.DataDir, "conf", "samples", tutorialConfDIR)
	if tutorialCfg, err = config.NewCGRConfigFromPath(tutorialCfgPath); err != nil {
		t.Error(err)
	}
	switch tutorialConfDIR {
	case "mongoatlas": // Mongo needs more time to reset db
		tutorialDelay = 4000
	default:
		tutorialDelay = 2000
	}
}

func testTutorialResetDB(t *testing.T) {
	if err := engine.InitDataDB(tutorialCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(tutorialCfg); err != nil {
		t.Fatal(err)
	}
}

func testTutorialStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutorialCfgPath, tutorialDelay); err != nil {
		t.Fatal(err)
	}
}

func testTutorialRpcConn(t *testing.T) {
	tutorialRpc = engine.NewRPCClient(t, tutorialCfg.ListenCfg())
}

func testTutorialFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := tutorialRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testTutorialGetCost(t *testing.T) {
	attrs := v1.AttrGetCost{
		Tenant:      "cgrates.org",
		Category:    "call",
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "2m10s",
	}
	var rply *engine.EventCost
	if err := tutorialRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.716900 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testTutorialStopEngine(t *testing.T) {
	if err := engine.KillEngine(tutorialDelay); err != nil {
		t.Error(err)
	}
}
