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
	"flag"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	itTestMongoAtalas = flag.Bool("mongo_atlas", false, "Run the test with mongo atalas connection")
	tutorialCfgPath   string
	tutorialCfg       *config.CGRConfig
	tutorialRpc       *rpc.Client
	tutorialConfDIR   string //run tests for specific configuration
	tutorialDelay     int
)

var sTestsTutorials = []func(t *testing.T){
	testTutorialLoadConfig,
	testTutorialResetDB,
	testTutorialStartEngine,
	testTutorialRpcConn,
	testTutorialFromFolder,
	testTutorialGetCost,
	testTutorialStopEngine,
}

//Test start here
func TestTutorialMongoAtlas(t *testing.T) {
	if !*itTestMongoAtalas {
		return
	}
	tutorialConfDIR = "mongoatlas"
	for _, stest := range sTestsTutorials {
		t.Run(tutorialConfDIR, stest)
	}
}

func TestTutorialMongo(t *testing.T) {
	tutorialConfDIR = "tutmongo"
	for _, stest := range sTestsTutorials {
		t.Run(tutorialConfDIR, stest)
	}
}

func TestTutorialMySQL(t *testing.T) {
	tutorialConfDIR = "tutmysql"
	for _, stest := range sTestsTutorials {
		t.Run(tutorialConfDIR, stest)
	}
}

func testTutorialLoadConfig(t *testing.T) {
	var err error
	tutorialCfgPath = path.Join(*dataDir, "conf", "samples", tutorialConfDIR)
	if tutorialCfg, err = config.NewCGRConfigFromFolder(tutorialCfgPath); err != nil {
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
	if err := engine.InitDataDb(tutorialCfg); err != nil {
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
	var err error
	tutorialRpc, err = jsonrpc.Dial("tcp", tutorialCfg.ListenCfg().RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testTutorialFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := tutorialRpc.Call("ApierV1.LoadTariffPlanFromFolder", attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
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
	if err := tutorialRpc.Call("ApierV1.GetCost", attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.316900 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
}

func testTutorialStopEngine(t *testing.T) {
	if err := engine.KillEngine(tutorialDelay); err != nil {
		t.Error(err)
	}
}
