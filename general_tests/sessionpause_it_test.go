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
	"encoding/json"
	"path"
	"reflect"
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
	sesPauseCfgDir  string
	sesPauseCfgPath string
	sesPauseCfg     *config.CGRConfig
	sesPauseRPC     *birpc.Client

	sesPauseTests = []func(t *testing.T){
		testSesPauseItLoadConfig,
		testSesPauseItResetDataDB,
		testSesPauseItResetStorDb,
		testSesPauseItStartEngine,
		testSesPauseItRPCConn,
		testSesPauseItLoadFromFolder,

		testSesPauseItAllPause,
		testSesPauseItInitPause,
		testSesPauseItInitUpdatePause,
		testSesPauseItUpdatePause,

		testSesPauseItStopCgrEngine,
	}
)

func TestSesPauseItSessions(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesPauseCfgDir = "tutinternal"
	case utils.MetaMySQL:
		sesPauseCfgDir = "tutmysql"
	case utils.MetaMongo:
		sesPauseCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sesPauseTests {
		t.Run(sesPauseCfgDir, stest)
	}
}

func testSesPauseItLoadConfig(t *testing.T) {
	var err error
	sesPauseCfgPath = path.Join(*utils.DataDir, "conf", "samples", sesPauseCfgDir)
	if sesPauseCfg, err = config.NewCGRConfigFromPath(sesPauseCfgPath); err != nil {
		t.Error(err)
	}
}

func testSesPauseItResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesPauseCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesPauseItResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesPauseCfg); err != nil {
		t.Fatal(err)
	}
}

func testSesPauseItStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesPauseCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesPauseItRPCConn(t *testing.T) {
	sesPauseRPC = engine.NewRPCClient(t, sesPauseCfg.ListenCfg())
}

func testSesPauseItLoadFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "tutorial")}
	if err := sesPauseRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func testSesPauseItInitSession(t *testing.T, cgrID string, chargeable bool, usage time.Duration) {
	args1 := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.CGRID:        cgrID,
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
			APIOpts: map[string]any{
				utils.OptsDebitInterval: "0s",
				utils.OptsChargeable:    chargeable,
			},
		},
	}
	var rply1 sessions.V1InitSessionReply
	if err := sesPauseRPC.Call(context.Background(), utils.SessionSv1InitiateSession,
		args1, &rply1); err != nil {
		t.Error(err)
		return
	} else if rply1.MaxUsage != nil && *rply1.MaxUsage != usage {
		t.Errorf("Unexpected MaxUsage: %v", rply1.MaxUsage)
	}
}

func testSesPauseItUpdateSession(t *testing.T, cgrID string, chargeable bool, usage time.Duration) {
	updtArgs := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.CGRID:        cgrID,
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
			APIOpts: map[string]any{
				utils.OptsChargeable: chargeable,
			},
		},
	}

	var updtRpl sessions.V1UpdateSessionReply
	if err := sesPauseRPC.Call(context.Background(), utils.SessionSv1UpdateSession, updtArgs, &updtRpl); err != nil {
		t.Error(err)
	}
	if updtRpl.MaxUsage == nil || *updtRpl.MaxUsage != usage {
		t.Errorf("Expecting : %+v, received: %+v", usage, updtRpl.MaxUsage)
	}
}

func testSesPauseItTerminateSession(t *testing.T, cgrID string, chargeable bool, usage time.Duration) {
	args := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.CGRID:        cgrID,
				utils.Category:     utils.Call,
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "TestDebitIterval",
				utils.RequestType:  utils.MetaPrepaid,
				utils.AccountField: "1001",
				utils.Subject:      "1002",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
				utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
				utils.Usage:        usage,
			},
			APIOpts: map[string]any{
				utils.OptsChargeable: chargeable,
			},
		},
	}
	var rply string
	if err := sesPauseRPC.Call(context.Background(), utils.SessionSv1TerminateSession,
		args, &rply); err != nil {
		t.Error(err)
	}
	if rply != utils.OK {
		t.Errorf("Unexpected reply: %s", rply)
	}
	if err := sesPauseRPC.Call(context.Background(), utils.SessionSv1ProcessCDR, &utils.CGREvent{
		Tenant: "cgrates.org",
		ID:     "testSesPauseItProccesCDR",
		Event: map[string]any{
			utils.CGRID:        cgrID,
			utils.Category:     utils.Call,
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     "TestDebitIterval",
			utils.RequestType:  utils.MetaPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1002",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2018, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        usage,
		},
	}, &rply); err != nil {
		t.Error(err)
	} else if rply != utils.OK {
		t.Errorf("Received reply: %s", rply)
	}
}

func testSesPauseItAllPause(t *testing.T) {
	cgrID := "1b676c7583ceb27ad7c991ded73b2417faa29a6a"
	testSesPauseItInitSession(t, cgrID, false, time.Second)
	testSesPauseItUpdateSession(t, cgrID, false, 2*time.Second)
	testSesPauseItTerminateSession(t, cgrID, false, 5*time.Second)
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}, CGRIDs: []string{cgrID}}
	if err := sesPauseRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	exp := []*engine.ExternalCDR{{
		CGRID:       cgrID,
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
		SetupTime:   "2018-01-07 17:00:00 +0000 UTC",
		AnswerTime:  "2018-01-07 17:00:10 +0000 UTC",
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
	var err error
	var val time.Time
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].AnswerTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].AnswerTime = val.UTC().String()
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].SetupTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].SetupTime = val.UTC().String()
	if !reflect.DeepEqual(exp, cdrs) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(exp), utils.ToJSON(cdrs))
	}

	var cd engine.EventCost
	if err := json.Unmarshal([]byte(cdString), &cd); err != nil {
		t.Fatal(err)
	}
	evCost := engine.EventCost{
		CGRID:     cgrID,
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
func testSesPauseItInitPause(t *testing.T) {
	cgrID := "1b676c7583ceb27ad7c991ded73b2417faa29a6b"
	testSesPauseItInitSession(t, cgrID, false, time.Second)
	testSesPauseItUpdateSession(t, cgrID, true, 2*time.Second)
	testSesPauseItTerminateSession(t, cgrID, true, 5*time.Second)
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}, CGRIDs: []string{cgrID}}
	if err := sesPauseRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	exp := []*engine.ExternalCDR{{
		CGRID:       cgrID,
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
		SetupTime:   "2018-01-07 17:00:00 +0000 UTC",
		AnswerTime:  "2018-01-07 17:00:10 +0000 UTC",
		Usage:       "5s",
		ExtraFields: map[string]string{},
		CostSource:  "*sessions",
		Cost:        0.2,
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
	var err error
	var val time.Time
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].AnswerTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].AnswerTime = val.UTC().String()
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].SetupTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].SetupTime = val.UTC().String()
	if !reflect.DeepEqual(exp, cdrs) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(exp), utils.ToJSON(cdrs))
	}

	var cd engine.EventCost
	if err := json.Unmarshal([]byte(cdString), &cd); err != nil {
		t.Fatal(err)
	}
	evCost := engine.EventCost{
		CGRID:     cgrID,
		RunID:     "*default",
		StartTime: time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		Usage:     utils.DurationPointer(61 * time.Second),
		Cost:      utils.Float64Pointer(0.2),
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
			RatingID: "6b4139e",
			Increments: []*engine.ChargingIncrement{{
				Usage:          60 * time.Second,
				Cost:           0.2,
				AccountingID:   "8305f9f",
				CompressFactor: 1,
			}},
			CompressFactor: 1,
		}},
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: engine.BalanceSummaries{{
				UUID:    "219cabeb-16cb-446c-837c-8cc74230eecf",
				ID:      "test",
				Type:    "*monetary",
				Initial: 10,
				Value:   9.8,
				Weight:  10,
			}},
		},
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
			"6b4139e": {
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          0,
				MaxCostStrategy:  "",
				TimingID:         "",
				RatesID:          "3d3179b",
				RatingFiltersID:  "00eaefb",
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
			"8305f9f": {
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "219cabeb-16cb-446c-837c-8cc74230eecf",
				RatingID:      "",
				Units:         0.2,
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
			"00eaefb": {
				utils.DestinationID:         "DST_1002",
				utils.DestinationPrefixName: "1002",
				utils.RatingPlanID:          "RP_1001",
				utils.Subject:               "*out:cgrates.org:call:1001",
			},
		},
		Rates: engine.ChargedRates{
			utils.MetaPause: {{
				GroupIntervalStart: 0,
				Value:              0,
				RateIncrement:      1,
				RateUnit:           1,
			}},
			"3d3179b": {{
				GroupIntervalStart: 0,
				Value:              0.2,
				RateIncrement:      60 * time.Second,
				RateUnit:           60 * time.Second,
			}, {
				GroupIntervalStart: 60 * time.Second,
				Value:              0.1,
				RateIncrement:      time.Second,
				RateUnit:           60 * time.Second,
			}},
		},
		Timings: engine.ChargedTimings{
			"": {
				StartTime: "00:00:00",
			},
		},
	}
	// we already tested that the keys are populated corectly
	// sync the curent keys to better compare
	cd.AccountSummary.BalanceSummaries[0].UUID = evCost.AccountSummary.BalanceSummaries[0].UUID
	for k, ac := range cd.Accounting {
		if k != utils.MetaPause {
			ac.BalanceUUID = evCost.AccountSummary.BalanceSummaries[0].UUID
		}
	}
	cd.SyncKeys(&evCost)
	if !reflect.DeepEqual(evCost, cd) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(evCost), utils.ToJSON(cd))
	}
}

func testSesPauseItInitUpdatePause(t *testing.T) {
	cgrID := "1b676c7583ceb27ad7c991ded73b2417faa29a6c"
	testSesPauseItInitSession(t, cgrID, false, time.Second)
	testSesPauseItUpdateSession(t, cgrID, false, 2*time.Second)
	testSesPauseItTerminateSession(t, cgrID, true, 5*time.Second)
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}, CGRIDs: []string{cgrID}}
	if err := sesPauseRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	exp := []*engine.ExternalCDR{{
		CGRID:       cgrID,
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
		SetupTime:   "2018-01-07 17:00:00 +0000 UTC",
		AnswerTime:  "2018-01-07 17:00:10 +0000 UTC",
		Usage:       "5s",
		ExtraFields: map[string]string{},
		CostSource:  "*sessions",
		Cost:        0.2,
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
	var err error
	var val time.Time
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].AnswerTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].AnswerTime = val.UTC().String()
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].SetupTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].SetupTime = val.UTC().String()
	if !reflect.DeepEqual(exp, cdrs) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(exp), utils.ToJSON(cdrs))
	}

	var cd engine.EventCost
	if err := json.Unmarshal([]byte(cdString), &cd); err != nil {
		t.Fatal(err)
	}
	evCost := engine.EventCost{
		CGRID:     cgrID,
		RunID:     "*default",
		StartTime: time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		Usage:     utils.DurationPointer(63 * time.Second),
		Cost:      utils.Float64Pointer(0.2),
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
			CompressFactor: 1,
		}, {
			RatingID: "6b4139e",
			Increments: []*engine.ChargingIncrement{{
				Usage:          60 * time.Second,
				Cost:           0.2,
				AccountingID:   "8305f9f",
				CompressFactor: 1,
			}},
			CompressFactor: 1,
		}},
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: engine.BalanceSummaries{{
				UUID:    "219cabeb-16cb-446c-837c-8cc74230eecf",
				ID:      "test",
				Type:    "*monetary",
				Initial: 9.8,
				Value:   9.6,
				Weight:  10,
			}},
		},
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
			"6b4139e": {
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          0,
				MaxCostStrategy:  "",
				TimingID:         "",
				RatesID:          "3d3179b",
				RatingFiltersID:  "00eaefb",
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
			"8305f9f": {
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "219cabeb-16cb-446c-837c-8cc74230eecf",
				RatingID:      "",
				Units:         0.2,
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
			"00eaefb": {
				utils.DestinationID:         "DST_1002",
				utils.DestinationPrefixName: "1002",
				utils.RatingPlanID:          "RP_1001",
				utils.Subject:               "*out:cgrates.org:call:1001",
			},
		},
		Rates: engine.ChargedRates{
			utils.MetaPause: {{
				GroupIntervalStart: 0,
				Value:              0,
				RateIncrement:      1,
				RateUnit:           1,
			}},
			"3d3179b": {{
				GroupIntervalStart: 0,
				Value:              0.2,
				RateIncrement:      60 * time.Second,
				RateUnit:           60 * time.Second,
			}, {
				GroupIntervalStart: 60 * time.Second,
				Value:              0.1,
				RateIncrement:      time.Second,
				RateUnit:           60 * time.Second,
			}},
		},
		Timings: engine.ChargedTimings{
			"": {
				StartTime: "00:00:00",
			},
		},
	}
	// we already tested that the keys are populated corectly
	// sync the curent keys to better compare
	cd.AccountSummary.BalanceSummaries[0].UUID = evCost.AccountSummary.BalanceSummaries[0].UUID
	for k, ac := range cd.Accounting {
		if k != utils.MetaPause {
			ac.BalanceUUID = evCost.AccountSummary.BalanceSummaries[0].UUID
		}
	}
	cd.SyncKeys(&evCost)
	if !reflect.DeepEqual(evCost, cd) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(evCost), utils.ToJSON(cd))
	}
}

func testSesPauseItUpdatePause(t *testing.T) {
	cgrID := "1b676c7583ceb27ad7c991ded73b2417faa29a6d"
	testSesPauseItInitSession(t, cgrID, true, 30*time.Second)      // debit 60s( the ratre interval is 60s)
	testSesPauseItUpdateSession(t, cgrID, false, 10*time.Second)   // have a pause of 10s
	testSesPauseItUpdateSession(t, cgrID, true, 40*time.Second)    // debit the 430s left from previsos debit and 10s extra
	testSesPauseItTerminateSession(t, cgrID, true, 65*time.Second) // end call with less than 80s and refund 10s
	time.Sleep(20 * time.Millisecond)
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{RequestTypes: []string{utils.MetaPrepaid}, CGRIDs: []string{cgrID}}
	if err := sesPauseRPC.Call(context.Background(), utils.APIerSv2GetCDRs, &req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}

	exp := []*engine.ExternalCDR{{
		CGRID:       cgrID,
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
		SetupTime:   "2018-01-07 17:00:00 +0000 UTC",
		AnswerTime:  "2018-01-07 17:00:10 +0000 UTC",
		Usage:       "1m5s",
		ExtraFields: map[string]string{},
		CostSource:  "*sessions",
		Cost:        0.6,
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
	var err error
	var val time.Time
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].AnswerTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].AnswerTime = val.UTC().String()
	if val, err = utils.ParseTimeDetectLayout(cdrs[0].SetupTime, ""); err != nil {
		t.Fatal(err)
	}
	cdrs[0].SetupTime = val.UTC().String()
	if !reflect.DeepEqual(exp, cdrs) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(exp), utils.ToJSON(cdrs))
	}

	var cd engine.EventCost
	if err := json.Unmarshal([]byte(cdString), &cd); err != nil {
		t.Fatal(err)
	}
	evCost := engine.EventCost{
		CGRID:     cgrID,
		RunID:     "*default",
		StartTime: time.Date(2018, time.January, 7, 16, 60, 10, 0, time.UTC),
		Usage:     utils.DurationPointer(70 * time.Second),
		Cost:      utils.Float64Pointer(0.6),
		Charges: []*engine.ChargingInterval{
			{
				RatingID: "6b4139e",
				Increments: []*engine.ChargingIncrement{
					{
						Usage:          0,
						Cost:           0.4,
						AccountingID:   "f2137e2",
						CompressFactor: 1,
					},
					{
						Usage:          60000000000,
						Cost:           0.2,
						AccountingID:   "8305f9f",
						CompressFactor: 1,
					}},
				CompressFactor: 1,
			},
			{
				RatingID: "*pause",
				Increments: []*engine.ChargingIncrement{{
					Usage:          10000000000,
					Cost:           0,
					AccountingID:   "*pause",
					CompressFactor: 1,
				}},
				CompressFactor: 1,
			},
		},
		AccountSummary: &engine.AccountSummary{
			Tenant: "cgrates.org",
			ID:     "1001",
			BalanceSummaries: engine.BalanceSummaries{{
				UUID:    "219cabeb-16cb-446c-837c-8cc74230eecf",
				ID:      "test",
				Type:    "*monetary",
				Initial: 9.6,
				Value:   9.,
				Weight:  10,
			}},
		},
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
			"6b4139e": {
				ConnectFee:       0.4,
				RoundingMethod:   "*up",
				RoundingDecimals: 4,
				MaxCost:          0,
				MaxCostStrategy:  "",
				TimingID:         "",
				RatesID:          "3d3179b",
				RatingFiltersID:  "00eaefb",
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
			"8305f9f": {
				AccountID:     "cgrates.org:1001",
				BalanceUUID:   "219cabeb-16cb-446c-837c-8cc74230eecf",
				RatingID:      "",
				Units:         0.2,
				ExtraChargeID: "",
			},
			"f2137e2": {
				AccountID:   "cgrates.org:1001",
				BalanceUUID: "219cabeb-16cb-446c-837c-8cc74230eecf",
				RatingID:    "",
				Units:       0.4,
			},
		},
		RatingFilters: engine.RatingFilters{
			utils.MetaPause: {
				utils.DestinationID:         "",
				utils.DestinationPrefixName: "",
				utils.RatingPlanID:          utils.MetaPause,
				utils.Subject:               "",
			},
			"00eaefb": {
				utils.DestinationID:         "DST_1002",
				utils.DestinationPrefixName: "1002",
				utils.RatingPlanID:          "RP_1001",
				utils.Subject:               "*out:cgrates.org:call:1001",
			},
		},
		Rates: engine.ChargedRates{
			utils.MetaPause: {{
				GroupIntervalStart: 0,
				Value:              0,
				RateIncrement:      1,
				RateUnit:           1,
			}},
			"3d3179b": {{
				GroupIntervalStart: 0,
				Value:              0.2,
				RateIncrement:      60 * time.Second,
				RateUnit:           60 * time.Second,
			}, {
				GroupIntervalStart: 60 * time.Second,
				Value:              0.1,
				RateIncrement:      time.Second,
				RateUnit:           60 * time.Second,
			}},
		},
		Timings: engine.ChargedTimings{
			"": {
				Years:     utils.Years{},
				Months:    utils.Months{},
				MonthDays: utils.MonthDays{},
				WeekDays:  utils.WeekDays{},
				StartTime: "00:00:00",
			},
		},
	}
	// we already tested that the keys are populated corectly
	// sync the curent keys to better compare
	cd.AccountSummary.BalanceSummaries[0].UUID = evCost.AccountSummary.BalanceSummaries[0].UUID
	for k, ac := range cd.Accounting {
		if k != utils.MetaPause {
			ac.BalanceUUID = evCost.AccountSummary.BalanceSummaries[0].UUID
		}
	}
	cd.SyncKeys(&evCost)
	if !reflect.DeepEqual(evCost, cd) {
		t.Errorf("Expected %s \n received: %s", utils.ToJSON(evCost), utils.ToJSON(cd))
	}
}

func testSesPauseItStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
