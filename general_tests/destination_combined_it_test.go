//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package general_tests

import (
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Start test wth combines rating plan
// destination for 1002 cost 1CNT and *any cost 10CNT

var sTestsTutorials2 = []func(t *testing.T){
	testDestinationLoadConfig,
	testDestinationResetDB,
	testDestinationStartEngine,
	testDestinationRpcConn,
	testDestinationFromFolder,
	testDestinationGetCostFor1002,
	testDestinationGetCostFor1003,
	testTutorialStopEngine,
}

func TestDestinationCombines(t *testing.T) {
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
	for _, stest := range sTestsTutorials2 {
		t.Run(tutorialConfDIR, stest)
	}

}

func testDestinationLoadConfig(t *testing.T) {
	var err error
	tutorialCfgPath = path.Join(*utils.DataDir, "conf", "samples", tutorialConfDIR)
	if tutorialCfg, err = config.NewCGRConfigFromPath(tutorialCfgPath); err != nil {
		t.Error(err)
	}
	tutorialDelay = 2000

}

func testDestinationResetDB(t *testing.T) {
	if err := engine.InitDataDB(tutorialCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(tutorialCfg); err != nil {
		t.Fatal(err)
	}
}

func testDestinationStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutorialCfgPath, tutorialDelay); err != nil {
		t.Fatal(err)
	}
}

func testDestinationRpcConn(t *testing.T) {
	tutorialRpc = engine.NewRPCClient(t, tutorialCfg.ListenCfg())
}

func testDestinationFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tp_destination_with_any")}
	if err := tutorialRpc.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testDestinationGetCostFor1002(t *testing.T) {
	attrs := v1.AttrGetCost{
		Tenant:      "cgrates.org",
		Category:    "call",
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "1m",
	}
	var rply *engine.EventCost
	if err := tutorialRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected error received: ", err.Error())
	} else if *rply.Cost != 0.01 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testDestinationGetCostFor1003(t *testing.T) {
	attrs := v1.AttrGetCost{
		Tenant:      "cgrates.org",
		Category:    "call",
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "*now",
		Usage:       "1m",
	}
	var rply *engine.EventCost
	if err := tutorialRpc.Call(context.Background(), utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected error received: ", err.Error())
	} else if *rply.Cost != 0.3 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}
