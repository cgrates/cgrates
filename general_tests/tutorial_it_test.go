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
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tutorialCfgPath string
	tutorialCfg     *config.CGRConfig
	tutorialRpc     *rpc.Client
	tutorialConfDIR string //run tests for specific configuration
	tutorialDelay   int
	tutorialDataDir = "/usr/share/cgrates"
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
func TestTutorialMongo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping non-short test")
	}
	tutorialConfDIR = "mongo_atlas"
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
	tutorialCfgPath = path.Join(tutorialDataDir, "conf", "samples", tutorialConfDIR)
	fmt.Printf("Tutorial path %+v\n", tutorialCfgPath)
	if tutorialCfg, err = config.NewCGRConfigFromFolder(tutorialCfgPath); err != nil {
		t.Error(err)
	}
	fmt.Printf("Tutorial cfg %+v\n", tutorialCfg)
	switch tutorialConfDIR {
	case "mongo_atlas": // Mongo needs more time to reset db
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
	tutorialRpc, err = jsonrpc.Dial("tcp", tutorialCfg.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
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

}

func testTutorialStopEngine(t *testing.T) {
	if err := engine.KillEngine(tutorialDelay); err != nil {
		t.Error(err)
	}
}
