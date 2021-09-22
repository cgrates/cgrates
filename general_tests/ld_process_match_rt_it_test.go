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
package apis

import (
	"net"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/utils"
)

var (
	testLdPrMatchRtCfgDir  string
	testLdPrMatchRtCfgPath string
	testLdPrMatchRtCfg     *config.CGRConfig
	testLdPrMatchRtRPC     *birpc.Client

	testLdPrMatchRtTests = []func(t *testing.T){
		testLdPrMatchRtLoadConfig,
		testLdPrMatchRtResetDataDB,
		testLdPrMatchRtResetStorDb,
		testLdPrMatchRtStartEngine,
		testLdPrMatchRtRPCConn,
		testLdPrMatchRtLoadTP,
		testLdPrMatchRtCDRSProcessEvent,

		testLdPrMatchRtStopCgrEngine,
	}
)

type TestRPC struct {
	Event *utils.CGREventWithEeIDs
}

var testRPC1 TestRPC

func (rpc *TestRPC) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]interface{}) (err error) {
	rpc.Event = cgrEv
	return nil
}

func TestLdPrMatchRtChange(t *testing.T) {
	birpc.RegisterName(utils.EeSv1, &testRPC1)
	l, err := net.Listen(utils.TCP, ":22012")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go jsonrpc.ServeConn(c)
		}
	}()

	switch *dbType {
	case utils.MetaInternal:
		testLdPrMatchRtCfgDir = "ld_process_match_rt_internal"
	case utils.MetaMySQL:
		testLdPrMatchRtCfgDir = "ld_process_match_rt_mysql"
	case utils.MetaMongo:
		testLdPrMatchRtCfgDir = "ld_process_match_rt_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testLdPrMatchRtTest := range testLdPrMatchRtTests {
		t.Run(testLdPrMatchRtCfgDir, testLdPrMatchRtTest)
	}
}

func testLdPrMatchRtLoadConfig(t *testing.T) {
	var err error
	testLdPrMatchRtCfgPath = path.Join(*dataDir, "conf", "samples", testLdPrMatchRtCfgDir)
	if testLdPrMatchRtCfg, err = config.NewCGRConfigFromPath(context.Background(), testLdPrMatchRtCfgPath); err != nil {
		t.Error(err)
	}
}

func testLdPrMatchRtResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(testLdPrMatchRtCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchRtResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(testLdPrMatchRtCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchRtStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testLdPrMatchRtCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchRtRPCConn(t *testing.T) {
	var err error
	testLdPrMatchRtRPC, err = newRPCClient(testLdPrMatchRtCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchRtLoadTP(t *testing.T) {
	var reply string
	if err := testLdPrMatchRtRPC.Call(context.Background(), utils.LoaderSv1Load,
		&loaders.ArgsProcessFolder{
			StopOnError: true,
			Caching:     utils.StringPointer(utils.MetaReload),
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLdPrMatchRtCDRSProcessEvent(t *testing.T) {
	ev := &engine.ArgV1ProcessEvent{
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]interface{}{
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestEv1",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.Usage:        time.Minute,
			},
			APIOpts: map[string]interface{}{
				utils.OptsRateS:      true,
				utils.OptsCDRsExport: true,
			},
		},
	}
	var rply string
	if err := testLdPrMatchRtRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply); err != nil {
		t.Fatal(err)
	}
	expected := "OK"
	if !reflect.DeepEqual(utils.ToJSON(&expected), utils.ToJSON(&rply)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(&expected), utils.ToJSON(&rply))
	}
	// expected2 := utils.CGREvent{
	// 	Tenant: "cgrates.org",
	// 	ID:     "TestEv1",
	// 	Event: map[string]interface{}{
	// 		"Altered":         nil,
	// 		utils.Cost:        0.4,
	// 		"CostIntervals":   []map[string]interface{}{},
	// 		"ID":              "RT_RETAIL1",
	// 		"MaxCost":         0,
	// 		"MaxCostStrategy": "",
	// 		"MinCost":         0,
	// 		"Rates":           map[string]interface{}{},
	// 	},
	// }
	// if !reflect.DeepEqual(utils.ToJSON(expected2.Event), utils.ToJSON(testRPC1.Event.Event["*rateSCost"])) {
	// 	t.Errorf("\nExpecting : %+v \n,received: %+v", utils.ToJSON(expected2.Event), utils.ToJSON(testRPC1.Event.Event["*rateSCost"]))
	// }

}

func testLdPrMatchRtStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
