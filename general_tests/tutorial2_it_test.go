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
	"net/rpc/jsonrpc"
	"path"
	"testing"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	tutCfgPath string
	tutCfg     *config.CGRConfig
	tutRpc     *rpc.Client
	tutCfgDir  string //run tests for specific configuration
	tutDelay   int
)

var sTutTests = []func(t *testing.T){
	testTutLoadConfig,
	testTutResetDB,
	testTutStartEngine,
	testTutRpcConn,
	testTutFromFolder,
	testTutGetCost,
	testTutStopEngine,
}

//Test start here
func TestTutorial2MySQL(t *testing.T) {
	tutCfgDir = "tutmysql2"
	for _, stest := range sTutTests {
		t.Run(tutCfgDir, stest)
	}
}

func TestTutorial2Mongo(t *testing.T) {
	tutCfgDir = "tutmongo2"
	for _, stest := range sTutTests {
		t.Run(tutCfgDir, stest)
	}
}

func testTutLoadConfig(t *testing.T) {
	var err error
	tutCfgPath = path.Join(*dataDir, "conf", "samples", tutCfgDir)
	if tutCfg, err = config.NewCGRConfigFromPath(tutCfgPath); err != nil {
		t.Error(err)
	}
	switch tutCfgDir {
	default:
		tutDelay = 2000
	}
}

func testTutResetDB(t *testing.T) {
	if err := engine.InitDataDb(tutCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(tutCfg); err != nil {
		t.Fatal(err)
	}
}

func testTutStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(tutCfgPath, tutDelay); err != nil {
		t.Fatal(err)
	}
}

func testTutRpcConn(t *testing.T) {
	var err error
	if tutRpc, err = jsonrpc.Dial("tcp", tutCfg.ListenCfg().RPCJSONListen); err != nil {
		t.Fatal("could not connect to rater: ", err.Error())
	}
}

func testTutFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{
		FolderPath: path.Join(*dataDir, "tariffplans", "tutorial2")}
	if err := tutRpc.Call(utils.ApierV1LoadTariffPlanFromFolder,
		attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}

func testTutGetCost(t *testing.T) {
	// Standard pricing for 1001->1002
	attrs := v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "45s",
	}
	var rply *engine.EventCost
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.45 { // FixMe: missing ConnectFee out of Cost
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T09:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 1.2 { // FixMe: missing ConnectFee out of Cost
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T21:00:00Z",
		Usage:       "45s",
	}
	/*
		// *any to 2001
		attrs = v1.AttrGetCost{
			Subject:     "1002",
			Destination: "2001",
			AnswerTime:  "*now",
			Usage:       "45s",
		}
		if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
			t.Error("Unexpected nil error received: ", err.Error())
		} else if *rply.Cost != 1.2 { // FixMe: missing ConnectFee out of Cost
			t.Errorf("Unexpected cost received: %f", *rply.Cost)
		}
	*/
	// *any to 2001 on NEW_YEAR
	attrs = v1.AttrGetCost{
		Subject:     "1002",
		Destination: "2001",
		AnswerTime:  "2020-01-01T21:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.45 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Fallback pricing from *any, Usage will be rounded to 60s
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "2019-03-11T21:00:00Z",
		Usage:       "45s",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.45 { // FixMe: missing ConnectFee out of Cost
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// Unauthorized destination
	attrs = v1.AttrGetCost{
		Subject:     "1001",
		Destination: "4003",
		AnswerTime:  "2019-03-11T09:00:00Z",
		Usage:       "1m",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: UNAUTHORIZED_DESTINATION" {
		t.Error("Unexpected nil error received: ", err)
	}
	// Data charging
	attrs = v1.AttrGetCost{
		Category:   "data",
		Subject:    "1001",
		AnswerTime: "*now",
		Usage:      "2048",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 2.0 { // FixMe: missing ConnectFee out of Cost
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging 1002
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1003",
		Destination: "1002",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.1 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging 10
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1001",
		Destination: "1003",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err != nil {
		t.Error("Unexpected nil error received: ", err.Error())
	} else if *rply.Cost != 0.2 {
		t.Errorf("Unexpected cost received: %f", *rply.Cost)
	}
	// SMS charging UNAUTHORIZED
	attrs = v1.AttrGetCost{
		Category:    "sms",
		Subject:     "1001",
		Destination: "2001",
		AnswerTime:  "*now",
		Usage:       "1",
	}
	if err := tutRpc.Call(utils.ApierV1GetCost, attrs, &rply); err == nil ||
		err.Error() != "SERVER_ERROR: UNAUTHORIZED_DESTINATION" {
		t.Error("Unexpected nil error received: ", err)
	}
}

func testTutStopEngine(t *testing.T) {
	if err := engine.KillEngine(tutDelay); err != nil {
		t.Error(err)
	}
}
