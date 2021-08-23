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

//Test start here
func TestCostIT(t *testing.T) {
	switch *dbType {
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
	costCfgPath = path.Join(*dataDir, "conf", "samples", costConfigDIR)
	costCfg, err = config.NewCGRConfigFromPath(costCfgPath)
	if err != nil {
		t.Error(err)
	}
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
	if _, err := engine.StopStartEngine(costCfgPath, *waitRater); err != nil {
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
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := costRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testCostGetCost(t *testing.T) {
	attrs := AttrGetCost{Category: "call", Tenant: "cgrates.org",
		Subject: "1001", AnswerTime: "*now", Destination: "1002", Usage: "120000000000"} //120s ( 2m)
	var rply *engine.EventCost
	if err := costRPC.Call(utils.APIerSv1GetCost, &attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.700200 { // expect to get 0.7 (0.4 connect fee 0.2 first minute 0.1 each minute after)
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testCostKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
