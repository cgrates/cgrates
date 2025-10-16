//go:build integration
// +build integration

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
		testLdPrMatchAcFlushDBs,

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

func (rpc *TestRPC1) ProcessEvent(ctx *context.Context, cgrEv *utils.CGREventWithEeIDs, rply *map[string]map[string]any) (err error) {
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

	switch *utils.DBType {
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
	testLdPrMatchAcCfgPath = path.Join(*utils.DataDir, "conf", "samples", testLdPrMatchAcCfgDir)
	if testLdPrMatchAcCfg, err = config.NewCGRConfigFromPath(context.Background(), testLdPrMatchAcCfgPath); err != nil {
		t.Error(err)
	}
}

func testLdPrMatchAcFlushDBs(t *testing.T) {
	if err := engine.InitDB(testLdPrMatchAcCfg); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testLdPrMatchAcCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testLdPrMatchAcRPCConn(t *testing.T) {
	testLdPrMatchAcRPC = engine.NewRPCClient(t, testLdPrMatchAcCfg.ListenCfg(), *utils.Encoding)
}

func testLdPrMatchAcLoadTP(t *testing.T) {
	caching := utils.MetaReload
	if testLdPrMatchAcCfg.DbCfg().DBConns[utils.MetaDefault].Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := testLdPrMatchAcRPC.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
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
		Event: map[string]any{
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestEv1",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
		},
		APIOpts: map[string]any{
			utils.MetaUsage:      2 * time.Minute,
			utils.MetaRates:      false,
			utils.OptsCDRsExport: true,
			utils.MetaAccounts:   true,
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
			Event: map[string]any{

				"Account":     "1001",
				"Destination": "1002",
				"OriginID":    "TestEv1",
				"RequestType": "*prepaid",
				"Subject":     "1001",
				"ToR":         "*voice",
			},
			APIOpts: map[string]any{
				utils.MetaAccountSCost: map[string]any{
					"Abstracts":  0,
					"Accounting": map[string]any{},
					"Accounts": map[string]any{
						"1001": map[string]any{
							"Balances": map[string]any{
								"VoiceBalance": map[string]any{
									"AttributeIDs":   nil,
									"Blockers":       nil,
									"CostIncrements": nil,
									"FilterIDs":      nil,
									"ID":             "VoiceBalance",
									"Opts":           map[string]any{},
									"RateProfileIDs": nil,
									"Type":           utils.MetaAbstract,
									"UnitFactors":    nil,
									"Units":          3600000000000,
									"Weights": []map[string]any{
										{
											"FilterIDs": nil,
											"Weight":    10,
										},
									},
								},
							},
							"Blockers":     nil,
							"FilterIDs":    nil,
							"ID":           "1001",
							"Opts":         map[string]any{},
							"Tenant":       "cgrates.org",
							"ThresholdIDs": nil,
							"Weights":      nil,
						},
					},
					"Charges":     nil,
					"Concretes":   nil,
					"Rates":       map[string]any{},
					"Rating":      map[string]any{},
					"UnitFactors": map[string]any{},
				},
				utils.MetaUsage:      2 * time.Minute,
				utils.OptsCDRsExport: true,
				utils.MetaRates:      false,
				utils.MetaAccounts:   true,
			},
		},
	}
	delete(testRPC2.Event.APIOpts, utils.MetaCDRID)
	if !reflect.DeepEqual(utils.ToJSON(expected2), utils.ToJSON(testRPC2.Event)) {
		t.Errorf("\nExpecting : %+v\n,received: %+v", utils.ToJSON(expected2), utils.ToJSON(testRPC2.Event))
	}

}

func testLdPrMatchAcStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
