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
	testLdPrMatchAcCfgDir  string
	testLdPrMatchAcCfgPath string
	testLdPrMatchAcCfg     *config.CGRConfig
	testLdPrMatchAcRPC     *birpc.Client

	testLdPrMatchAcTests = []func(t *testing.T){
		testLdPrMatchAcLoadConfig,
		testLdPrMatchAcResetDataDB,
		testLdPrMatchAcResetStorDb,
		testLdPrMatchAcStartEngine,
		testLdPrMatchAcRPCConn,
		testLdPrMatchAcLoadTP,
		testLdPrMatchAcCDRSProcessEvent,
		testLdPrMatchAcStopCgrEngine,
	}
)

type TestRPC1 struct {
	Event *utils.CGREventWithEeIDs
}

var testRPC2 TestRPC1

func (rpc *TestRPC1) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]interface{}) (err error) {
	rpc.Event = cgrEv
	return nil
}

func TestLdPrMatchAcChange(t *testing.T) {
	birpc.RegisterName(utils.EeSv1, &testRPC2)
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
		testLdPrMatchAcCfgDir = "ld_process_match_rt_internal"
	case utils.MetaMySQL:
		testLdPrMatchAcCfgDir = "ld_process_match_rt_mysql"
	case utils.MetaMongo:
		testLdPrMatchAcCfgDir = "ld_process_match_rt_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testLdPrMatchAcTest := range testLdPrMatchAcTests {
		t.Run(testLdPrMatchAcCfgDir, testLdPrMatchAcTest)
	}
}

func testLdPrMatchAcLoadConfig(t *testing.T) {
	var err error
	testLdPrMatchAcCfgPath = path.Join(*dataDir, "conf", "samples", testLdPrMatchAcCfgDir)
	if testLdPrMatchAcCfg, err = config.NewCGRConfigFromPath(context.Background(), testLdPrMatchAcCfgPath); err != nil {
		t.Error(err)
	}
}

func testLdPrMatchAcResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(testLdPrMatchAcCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(testLdPrMatchAcCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testLdPrMatchAcCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcRPCConn(t *testing.T) {
	var err error
	testLdPrMatchAcRPC, err = newRPCClient(testLdPrMatchAcCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcLoadTP(t *testing.T) {
	caching := utils.MetaReload
	if testLdPrMatchAcCfg.DataDbCfg().Type == utils.Internal {
		caching = utils.MetaNone
	}
	var reply string
	if err := testLdPrMatchAcRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]interface{}{
				utils.MetaCache:       caching,
				utils.MetaStopOnError: true,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}
}

func testLdPrMatchAcCDRSProcessEvent(t *testing.T) {
	ev := &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "TestEv1",
		Event: map[string]interface{}{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestEv1",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]interface{}{
			utils.MetaUsage:      2 * time.Minute,
			utils.OptsRateS:      false,
			utils.OptsCDRsExport: true,
			utils.OptsAccountS:   true,
		},
	}
	var rply string
	if err := testLdPrMatchAcRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, ev, &rply); err != nil {
		t.Fatal(err)
	}
	expected := "OK"
	if !reflect.DeepEqual(utils.ToJSON(&expected), utils.ToJSON(&rply)) {
		t.Errorf("Expecting : %+v, received: %+v", utils.ToJSON(&expected), utils.ToJSON(&rply))
	}

	expected2 := &utils.CGREventWithEeIDs{
		EeIDs: nil,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "TestEv1",
			Event: map[string]interface{}{
				utils.MetaAccountSCost: map[string]interface{}{
					"Abstracts":  0,
					"Accounting": map[string]interface{}{},
					"Accounts": map[string]interface{}{
						"1001": map[string]interface{}{
							"Balances": map[string]interface{}{
								"VoiceBalance": map[string]interface{}{
									"AttributeIDs":   nil,
									"CostIncrements": nil,
									"FilterIDs":      nil,
									"ID":             "VoiceBalance",
									"Opts":           map[string]interface{}{},
									"RateProfileIDs": nil,
									"Type":           utils.MetaAbstract,
									"UnitFactors":    nil,
									"Units":          3600000000000,
									"Weights": []map[string]interface{}{
										{
											"FilterIDs": nil,
											"Weight":    10,
										},
									},
								},
							},
							"FilterIDs":    nil,
							"ID":           "1001",
							"Opts":         map[string]interface{}{},
							"Tenant":       "cgrates.org",
							"ThresholdIDs": nil,
							"Weights": []map[string]interface{}{{
								"FilterIDs": nil,
								"Weight":    0,
							}},
						},
					},
					"Charges":     nil,
					"Concretes":   nil,
					"Rates":       map[string]interface{}{},
					"Rating":      map[string]interface{}{},
					"UnitFactors": map[string]interface{}{},
				},
				"Account":     "1001",
				"Destination": "1002",
				"OriginID":    "TestEv1",
				"RequestType": "*prepaid",
				"Subject":     "1001",
				"ToR":         "*voice",
			},

			APIOpts: map[string]interface{}{
				utils.MetaUsage:      2 * time.Minute,
				utils.OptsCDRsExport: true,
				utils.OptsRateS:      false,
				utils.OptsAccountS:   true,
			},
		},
	}

	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(testRPC2.Event)) {
		t.Errorf("\nExpecting : %+v\n,received: %+v", utils.ToJSON(expected2), utils.ToJSON(testRPC2.Event))
	}

}

func testLdPrMatchAcStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
