//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	costCfgPath   string
	costCfg       *config.CGRConfig
	costRPC       *rpc.Client
	costConfigDIR string //run tests for specific configuration
	costDataDir   = "/usr/share/cgrates"

	sTestsCost = []func(t *testing.T){
		testCostInitCfg,
		testCostInitDataDb,
		testCostResetStorDb,
		testCostStartEngine,
		testCostRPCConn,
		testCostLoadFromFolder,
		testCostGetCost,
		testCostKillEngine,
	}
)

// Test start here
func TestCostIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		costConfigDIR = "tutinternal"
	case utils.MetaMySQL:
		costConfigDIR = "tutmysql"
	case utils.MetaMongo:
		costConfigDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCost {
		t.Run(costConfigDIR, stest)
	}
}

func testCostInitCfg(t *testing.T) {
	var err error
	costCfgPath = path.Join(costDataDir, "conf", "samples", costConfigDIR)
	costCfg, err = config.NewCGRConfigFromPath(costCfgPath)
	if err != nil {
		t.Error(err)
	}
	costCfg.DataFolderPath = costDataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(costCfg)
}

func testCostInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(costCfg); err != nil {
		t.Fatal(err)
	}
}

// Wipe out the cdr database
func testCostResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(costCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testCostStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(costCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCostRPCConn(t *testing.T) {
	var err error
	costRPC, err = newRPCClient(costCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testCostLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := costRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testCostGetCost(t *testing.T) {
	attrs := AttrGetCost{Category: "call", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: "*now", Destination: "1002", Usage: "120000000000"} //120s ( 2m)
	var rply *engine.EventCost
	if err := costRPC.Call(utils.APIerSv1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.700200 { // expect to get 0.7 (0.4 connect fee 0.2 first minute 0.1 each minute after)
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testCostKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
