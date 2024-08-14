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
package general_tests

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
		testLdPrMatchRtFlushDBs,

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

var testRPCrt1 TestRPC

func (rpc *TestRPC) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]any) (err error) {
	rpc.Event = cgrEv
	return nil
}

func TestLdPrMatchRtChange(t *testing.T) {
	birpc.RegisterName(utils.EeSv1, &testRPCrt1)
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

func testLdPrMatchRtFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(testLdPrMatchRtCfg); err != nil {
		t.Fatal(err)
	}
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
	testLdPrMatchRtRPC, err = engine.NewRPCClient(testLdPrMatchRtCfg.ListenCfg(), *encoding)
	if err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchRtLoadTP(t *testing.T) {
	var reply string
	if err := testLdPrMatchRtRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaStopOnError: true,
				utils.MetaCache:       utils.MetaReload,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLdPrMatchRtCDRSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEv1",
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestEv1",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.Usage:        time.Minute,
		},
		APIOpts: map[string]any{
			utils.MetaRates:      true,
			utils.OptsCDRsExport: true,
			utils.MetaAccounts:   false,
		},
	}
	var rply string
	if err := testLdPrMatchRtRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply); err != nil {
		t.Fatal(err)
	}
	expected := "OK"
	if expected != rply {
		t.Errorf("Expecting : %q, received: %q", expected, rply)
	}
	time.Sleep(50 * time.Millisecond)
	if testRPCrt1.Event == nil {
		t.Fatal("The rpc was not called")
	}
	costIntervalRatesID := testRPCrt1.Event.APIOpts[utils.MetaRateSCost].(map[string]any)["CostIntervals"].([]any)[0].(map[string]any)["Increments"].([]any)[0].(map[string]any)["RateID"]
	expected2 := &utils.CGREventWithEeIDs{
		EeIDs: nil,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]any{
				"Account":     "1001",
				"Destination": "1002",
				"OriginID":    "TestEv1",
				"RequestType": "*prepaid",
				"Subject":     "1001",
				"ToR":         "*voice",
				"Usage":       60000000000,
			},
			APIOpts: map[string]any{
				utils.MetaCost: 0.4,
				utils.MetaRateSCost: map[string]any{
					"Altered":  nil,
					utils.Cost: 0.4,
					"CostIntervals": []map[string]any{
						{
							"CompressFactor": 1,
							"Increments": []map[string]any{
								{
									"CompressFactor":    2,
									"RateID":            costIntervalRatesID,
									"RateIntervalIndex": 0,
									"Usage":             60000000000,
								},
							},
						},
					},
					"ID":              "RT_RETAIL1",
					"MaxCost":         0,
					"MaxCostStrategy": "",
					"MinCost":         0,
					"Rates": map[string]any{
						utils.IfaceAsString(costIntervalRatesID): map[string]any{
							"FixedFee":      0,
							"Increment":     30000000000,
							"IntervalStart": 0,
							"RecurrentFee":  0.4,
							"Unit":          60000000000,
						},
					},
				},
				utils.MetaRates:      true,
				utils.OptsCDRsExport: true,
				utils.MetaAccounts:   false,
			},
		},
	}
	delete(testRPCrt1.Event.APIOpts, utils.MetaCDRID)
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(testRPCrt1.Event)) {
		t.Errorf("\nExpecting : %+v \n,received: %+v", utils.ToJSON(expected2), utils.ToJSON(testRPCrt1.Event))
	}

}

func testLdPrMatchRtStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
