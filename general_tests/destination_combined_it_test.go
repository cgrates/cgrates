//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package general_tests

import (
	"net/rpc"
	"path"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// Start test wth combines rating plan
// destination for 1002 cost 1CNT and *any cost 10CNT
var (
	destCombCfgPath string
	destCombCfg     *config.CGRConfig
	destCombRPC     *rpc.Client
	destCombConfDIR string
	destCombDelay   int

	sTestsDestComb = []func(t *testing.T){
		testDestinationLoadConfig,
		testDestinationResetDB,
		testDestinationStartEngine,
		testDestinationRpcConn,
		testDestinationFromFolder,
		testDestinationGetCostFor1002,
		testDestinationGetCostFor1003,
		testDestinationStopEngine,
	}
)

func TestDestinationCombines(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		destCombConfDIR = "tutinternal"
	case utils.MetaMySQL:
		destCombConfDIR = "tutmysql"
	case utils.MetaMongo:
		destCombConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsDestComb {
		t.Run(destCombConfDIR, stest)
	}

}

func testDestinationLoadConfig(t *testing.T) {
	var err error
	destCombCfgPath = path.Join(*utils.DataDir, "conf", "samples", destCombConfDIR)
	if destCombCfg, err = config.NewCGRConfigFromPath(destCombCfgPath); err != nil {
		t.Error(err)
	}
}

func testDestinationResetDB(t *testing.T) {
	if err := engine.InitDataDb(destCombCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(destCombCfg); err != nil {
		t.Fatal(err)
	}
}

func testDestinationStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(destCombCfgPath, 2000); err != nil {
		t.Fatal(err)
	}
}

func testDestinationRpcConn(t *testing.T) {
	var err error
	destCombRPC, err = newRPCClient(destCombCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testDestinationFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tp_destination_with_any")}
	if err := destCombRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
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
	if err := destCombRPC.Call(utils.APIerSv1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
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
	if err := destCombRPC.Call(utils.APIerSv1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.3 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testDestinationStopEngine(t *testing.T) {
	if err := engine.KillEngine(2000); err != nil {
		t.Error(err)
	}
}
