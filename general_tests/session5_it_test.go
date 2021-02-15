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
	"encoding/json"
	"net/rpc"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	ses5CfgDir  string
	ses5CfgPath string
	ses5Cfg     *config.CGRConfig
	ses5RPC     *rpc.Client

	ses5Tests = []func(t *testing.T){
		testSes5ItLoadConfig,
		testSes5ItResetDataDB,
		testSes5ItResetStorDb,
		testSes5ItStartEngine,
		testSes5ItRPCConn,
		testSes5ItLoadFromFolder,

		testSes5ItAllPause,

		testSes5ItStopCgrEngine,
	}
)

func TestSes5ItSessions(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		ses5CfgDir = "tutinternal"
	case utils.MetaMySQL:
		ses5CfgDir = "tutmysql"
	case utils.MetaMongo:
		ses5CfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range ses5Tests {
		t.Run(ses5CfgDir, stest)
	}
}

func testSes5ItLoadConfig(t *testing.T) {
	ses5CfgPath = path.Join(*dataDir, "conf", "samples", ses5CfgDir)
	if ses5Cfg, err = config.NewCGRConfigFromPath(ses5CfgPath); err != nil {
		t.Error(err)
	}
}

func testSes5ItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(ses5Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes5ItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ses5Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSes5ItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ses5CfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSes5ItRPCConn(t *testing.T) {
	var err error
	ses5RPC, err = newRPCClient(ses5Cfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSes5ItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := ses5RPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSes5ItInitSession(t *testing.T, chargeable bool) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Category:     utils.Call,
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestDebitIterval",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        time.Second,
			},
			Opts: map[string]interface{}{
				utils.OptsDebitInterval: "0s",
				utils.OptsChargeable:    chargeable,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := ses5RPC.Call(utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage != nil && *rply1.MaxUsage != time.Second {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
}

func testSes5ItUpdateSession(t *testing.T, chargeable bool) {
	usage := 2 * time.Second
	updtArgs := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Category:     utils.Call,
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestDebitIterval",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        usage,
			},
			Opts: map[string]interface{}{
				utils.OptsChargeable: chargeable,
			},
		},
	}

	var updtRpl sessions.V1UpdateSessionReply
	if err := ses5RPC.Call(utils.SessionSv1UpdateSession, updtArgs, &updtRpl); err != nil {
		t.Error(err)
	}
	if updtRpl.MaxUsage == nil || *updtRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, updtRpl.MaxUsage)
	}
}

func testSes5ItTerminateSession(t *testing.T, chargeable bool) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]interface{}{
				utils.Category:     utils.Call,
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestDebitIterval",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1002",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        5 * time.Second,
			},
			Opts: map[string]interface{}{
				utils.OptsChargeable: chargeable,
			},
		},
	}
	var rply string
	if err := ses5RPC.Call(utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	if err := ses5RPC.Call(utils.SessionSv1ProcessCDR, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSes5ItProccesCDR",
		Event: map[string]interface{}{
			utils.Category:     utils.Call,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestDebitIterval",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1002",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        5 * time.Second,
		},
	}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Received reply: %s", rply)
	}
}

func testSes5ItAllPause(t *testing.T) {
	testSes5ItInitSession(t, false)
	testSes5ItUpdateSession(t, false)
	testSes5ItTerminateSession(t, false)
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}}
	if err := ses5RPC.Call(utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	exp := []*engine.ExternalCDR{{
		CGRID:       "1b676c7583ceb27ad7c991ded73b2417faa29a6a",
		RunID:       "*default",
		OrderID:     0,
		OriginHost:  "",
		Source:      "",
		OriginID:    "TestDebitIterval",
		ToR:         "*voice",
		RequestType: "*prepaid",
		Tenant:      "cgrates.org",
		Category:    "call",
		Account:     "1001",
		Subject:     "1001",
		Destination: "1002",
		SetupTime:   "2018-01-07T19:00:00+02:00",
		AnswerTime:  "2018-01-07T19:00:10+02:00",
		Usage:       "5s",
		ExtraFields: map[string]string{},
		CostSource:  "*sessions",
		Cost:        0,
		CostDetails: "",
		ExtraInfo:   "",
		PreRated:    false,
	}}
	if len(cdrs) == 0 {
		t.Fatal("No cdrs returned")
	}
	cdrs[0].OrderID = 0
	cdString := cdrs[0].CostDetails
	cdrs[0].CostDetails = ""
	if !reflect.DeepEqual(exp, cdrs) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(exp), utils.ToJSON(cdrs))
	}

	var cd engine.EventCost
	if err := json.Unmarshal([]byte(cdString), &cd); err != nil {
		t.Fatal(err)
	}
	evCost := engine.EventCost{
		CGRID:     "1b676c7583ceb27ad7c991ded73b2417faa29a6a",
		RunID:     "*default",
		StartTime: time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		Usage:     utils.DurationPointer(5 * time.Second),
		Cost:      utils.Float64Pointer(0),
		Charges: []*engine.ChargingInterval{{
			RatingID: utils.MetaPause,
			Increments: []*engine.ChargingIncrement{{
				Usage:          time.Second,
				Cost:           0,
				AccountingID:   utils.MetaPause,
				CompressFactor: 1,
			}},
			CompressFactor: 1,
		}, {
			RatingID: utils.MetaPause,
			Increments: []*engine.ChargingIncrement{{
				Usage:          2 * time.Second,
				Cost:           0,
				AccountingID:   utils.MetaPause,
				CompressFactor: 1,
			}},
			CompressFactor: 2,
		}},
		Rating: engine.Rating{
			utils.MetaPause: {
				ConnectFee:       0,
				RoundingMethod:   "*up",
				RoundingDecimals: 5,
				MaxCost:          0,
				MaxCostStrategy:  "",
				TimingID:         "",
				RatesID:          utils.MetaPause,
				RatingFiltersID:  utils.MetaPause,
			},
		},
		Accounting: engine.Accounting{
			utils.MetaPause: {
				AccountID:     "1001",
				BalanceUUID:   "",
				RatingID:      utils.MetaPause,
				Units:         0,
				ExtraChargeID: "",
			},
		},
		RatingFilters: engine.RatingFilters{
			utils.MetaPause: {
				utils.DestinationID:         "",
				utils.DestinationPrefixName: "",
				utils.RatingPlanID:          utils.MetaPause,
				utils.Subject:               "",
			},
		},
		Rates: engine.ChargedRates{
			utils.MetaPause: {{
				GroupIntervalStart: 0,
				Value:              0,
				RateIncrement:      1,
				RateUnit:           1,
			}},
		},
		Timings: engine.ChargedTimings{
			"": {
				StartTime: "00:00:00",
			},
		},
	}
	// the Timings are not relevant for this test
	for _, r := range cd.Rating {
		r.TimingID = ""
	}
	cd.Timings = evCost.Timings
	if !reflect.DeepEqual(evCost, cd) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(evCost), utils.ToJSON(cd))
	}

}

func testSes5ItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
