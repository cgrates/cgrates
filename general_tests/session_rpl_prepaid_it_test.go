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
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesRplPrePaidCfgPath string
	sesRplPrePaidCfgDIR  string
	sesRplPrePaidCfg     *config.CGRConfig
	sesRplPrePaidRPC     *birpc.Client

	sesRplPrePaidTests = []func(t *testing.T){
		testSeSRplPrepaidInitCfg,
		testSeSRplPrepaidResetDB,
		testSeSRplPrepaidStartEngine,
		testSeSRplPrepaidApierRpcConn,
		testSeSRplPrepaidTPFromFolder,
		testSeSRplPrepaidActivateSessions,
		testSeSRplPrepaidStopCgrEngine,
	}
)

func TestSeSRplPrepaid(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaMySQL:
		sesRplPrePaidCfgDIR = "tutmysql"
	case utils.MetaMongo:
		sesRplPrePaidCfgDIR = "tutmongo"
	case utils.MetaInternal, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesRplPrePaidTests {
		t.Run(*utils.DBType, stest)
	}
}

func testSeSRplPrepaidInitCfg(t *testing.T) {
	var err error
	sesRplPrePaidCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesRplPrePaidCfgDIR)
	if sesRplPrePaidCfg, err = config.NewCGRConfigFromPath(sesRplPrePaidCfgPath); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testSeSRplPrepaidResetDB(t *testing.T) {
	if err := engine.InitDataDb(sesRplPrePaidCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(sesRplPrePaidCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testSeSRplPrepaidStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesRplPrePaidCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testSeSRplPrepaidApierRpcConn(t *testing.T) {
	sesRplPrePaidRPC = engine.NewRPCClient(t, sesRplPrePaidCfg.ListenCfg())
}

// Load the tariff plan, creating accounts and their balances
func testSeSRplPrepaidTPFromFolder(t *testing.T) {
	var result string
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.APIerSv1SetChargerProfile, &engine.ChargerProfile{
		Tenant:       "cgrates.org",
		ID:           "Default",
		RunID:        utils.MetaRaw,
		AttributeIDs: []string{utils.MetaNone},
		Weight:       20,
	}, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1005",
		BalanceType: utils.MetaVoice,
		Value:       5 * float64(time.Hour), //value -> 20ms for future
		Balance: map[string]any{
			utils.ID:            "TestRplDebitBalance",
			utils.RatingSubject: "*zero5ms",
		},
	}
	var reply string
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
	var acnt *engine.Account
	//get balance
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply != float64(5*time.Hour) {
		t.Errorf("Expecting: %v, received: %v",
			float64(5*time.Second), rply)
	}
}

func testSeSRplPrepaidActivateSessions(t *testing.T) {
	var reply string
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.SessionSv1SetPassiveSession, sessions.Session{
		CGRID:         "ede927f8e42318a8db02c0f74adc2d9e16770339",
		Tenant:        "cgrates.org",
		ResourceID:    "testSeSRplPrepaidActivateSessions",
		DebitInterval: time.Second,
		EventStart: engine.NewMapEvent(map[string]any{
			utils.CGRID:        "ede927f8e42318a8db02c0f74adc2d9e16770339",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "testSeSRplPrepaidActivateSessions",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1005",
			utils.Destination:  "1004",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		}),
		SRuns: []*sessions.SRun{{
			Event: engine.NewMapEvent(map[string]any{
				utils.RunID:        utils.MetaRaw,
				utils.CGRID:        "ede927f8e42318a8db02c0f74adc2d9e16770339",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "testSeSRplPrepaidActivateSessions",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1005",
				utils.Destination:  "1004",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			}),
			CD: &engine.CallDescriptor{
				Category:      "call",
				Tenant:        "cgrates.org",
				Subject:       "1005",
				Account:       "1005",
				Destination:   "1004",
				TimeStart:     time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				TimeEnd:       time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				LoopIndex:     1,
				ToR:           utils.MetaVoice,
				CgrID:         "ede927f8e42318a8db02c0f74adc2d9e16770339",
				RunID:         utils.MetaRaw,
				DurationIndex: time.Second,
			},
			EventCost: &engine.EventCost{
				CGRID:     "ede927f8e42318a8db02c0f74adc2d9e16770339",
				RunID:     utils.MetaRaw,
				StartTime: time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				Charges: []*engine.ChargingInterval{{
					RatingID: "7edb98a",
					Increments: []*engine.ChargingIncrement{{
						Usage:          time.Second,
						AccountingID:   "5d1f19d",
						CompressFactor: 1,
					}},
					CompressFactor: 1,
				}},
				AccountSummary: &engine.AccountSummary{
					Tenant: "cgrates.org",
					ID:     "1005",
					BalanceSummaries: engine.BalanceSummaries{{
						UUID:    "dc88acfb-18c7-4658-9d36-056b74b84b57",
						ID:      "TestRplDebitBalance",
						Type:    utils.MetaVoice,
						Initial: 18000000000000,
						Value:   17999000000000,
					}},
				},
				Rating: engine.Rating{"7edb98a": {
					RatesID:         "5dad4b7",
					RatingFiltersID: "d24bb65",
				}},
				Accounting: engine.Accounting{"5d1f19d": {
					AccountID:     "cgrates.org:1005",
					BalanceUUID:   "dc88acfb-18c7-4658-9d36-056b74b84b57",
					Units:         float64(time.Second),
					ExtraChargeID: utils.MetaNone,
				}},
				RatingFilters: engine.RatingFilters{"d24bb65": {
					"DestinationID":     utils.MetaAny,
					"DestinationPrefix": "1004",
					"RatingPlanID":      utils.MetaNone,
					"Subject":           "dc88acfb-18c7-4658-9d36-056b74b84b57",
				}},
				Rates: engine.ChargedRates{"5dad4b7": {{
					RateIncrement: time.Second,
					RateUnit:      time.Second,
				}}},
				Timings: engine.ChargedTimings{},
			},
			LastUsage:  time.Second,
			TotalUsage: time.Second,
		}},
		OptsStart: map[string]any{
			utils.DebitInterval: "5ms",
		},
		Chargeable: true,
	}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	var aSessions []*sessions.ExternalSession
	// Activate first session (with ID: ede927f8e42318a8db02c0f74adc2d9e16770339)
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.SessionSv1ActivateSessions, &utils.SessionIDsWithArgsDispatcher{}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply %s", reply)
	}
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.SessionSv1GetActiveSessions, new(utils.SessionFilter), &aSessions); err != nil {
		t.Error(err)
	} else if len(aSessions) != 1 {
		t.Errorf("Expecting: 1 session, received: %+v sessions", len(aSessions))
	}
	time.Sleep(10 * time.Millisecond)
	var acnt *engine.Account
	if err := sesRplPrePaidRPC.Call(context.Background(), utils.APIerSv2GetAccount, &utils.AttrGetAccount{
		Tenant:  "cgrates.org",
		Account: "1005",
	}, &acnt); err != nil {
		t.Error(err)
	} else if rply := acnt.BalanceMap[utils.MetaVoice].GetTotalValue(); rply == float64(5*time.Hour) {
		t.Errorf("Expecting account to be debited") // no debit happend
	}
}

func testSeSRplPrepaidStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(1000); err != nil {
		t.Error(err)
	}
}
