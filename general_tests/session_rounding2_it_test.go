//go:build flaky

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
	"encoding/json"
	"net/rpc"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

var (
	sesRnd2CfgPath string
	sesRnd2CfgDIR  string
	sesRnd2Cfg     *config.CGRConfig
	sesRnd2RPC     *rpc.Client

	sTestSesRnd2 = []func(t *testing.T){
		testSesRnd2RemoveFolders,
		testSesRnd2CreateFolders,

		testSesRnd2LoadConfig,
		testSesRnd2ResetDataDB,
		testSesRnd2ResetStorDb,
		testSesRnd2StartEngine,
		testSesRnd2RPCConn,
		testSesRnd2LoadTP,

		testSesRnd2PrepaidInit,
		testSesRnd2PrepaidUpdate1,
		testSesRnd2PrepaidUpdate2,
		testSesRnd2PrepaidTerminate,
		testSesRnd2PrepaidProcessCDR,

		testSesRnd2PostpaidProcessCDR,

		testSesRnd2CompareCosts,

		testSesRnd2StopCgrEngine,
		testSesRnd2RemoveFolders,
	}
)

func TestSesRnd2IT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		sesRnd2CfgDIR = "sessions_internal"
	case utils.MetaMySQL:
		sesRnd2CfgDIR = "sessions_mysql"
	case utils.MetaMongo:
		sesRnd2CfgDIR = "sessions_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestSesRnd2 {
		t.Run(sesRnd2CfgDIR, stest)
	}
}

func testSesRnd2LoadConfig(t *testing.T) {
	sesRnd2CfgPath = path.Join(*utils.DataDir, "conf", "samples", sesRnd2CfgDIR)
	var err error
	if sesRnd2Cfg, err = config.NewCGRConfigFromPath(sesRnd2CfgPath); err != nil {
		t.Error(err)
	}
}

func testSesRnd2ResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(sesRnd2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRnd2ResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(sesRnd2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testSesRnd2StartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(sesRnd2CfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testSesRnd2RPCConn(t *testing.T) {
	var err error
	if sesRnd2RPC, err = newRPCClient(sesRnd2Cfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testSesRnd2LoadTP(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestSesRnd2IT", fileName))
		if err != nil {
			return err
		}
		defer csvFile.Close()
		_, err = csvFile.WriteString(data)
		if err != nil {
			return err

		}
		return csvFile.Sync()
	}

	// Create and populate AccountActions.csv
	if err := writeFile(utils.AccountActionsCsv, `
#Tenant,Account,ActionPlanId,ActionTriggersId,AllowNegative,Disabled
cgrates.org,1001,AP_PACKAGE_10,,,
cgrates.org,1002,AP_PACKAGE_10,,,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate ActionPlans.csv
	if err := writeFile(utils.ActionPlansCsv, `
#Id,ActionsId,TimingId,Weight
AP_PACKAGE_10,ACT_TOPUP_RST_10,*asap,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Actions.csv
	if err := writeFile(utils.ActionsCsv, `
#ActionsId[0],Action[1],ExtraParameters[2],Filter[3],BalanceId[4],BalanceType[5],Categories[6],DestinationIds[7],RatingSubject[8],SharedGroup[9],ExpiryTime[10],TimingIds[11],Units[12],BalanceWeight[13],BalanceBlocker[14],BalanceDisabled[15],Weight[16]
ACT_TOPUP_RST_10,*topup_reset,,,test,*monetary,,*any,,,*unlimited,,10,10,false,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Chargers.csv
	if err := writeFile(utils.ChargersCsv, `
#Tenant,ID,FilterIDs,ActivationInterval,RunID,AttributeIDs,Weight
cgrates.org,DEFAULT,,,*default,*none,0
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Destinations.csv
	if err := writeFile(utils.DestinationsCsv, `
#Id,Prefix
DST_1001,1001
DST_1002,1002
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate DestinationRates.csv
	if err := writeFile(utils.DestinationRatesCsv, `
#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_1001_20CNT,DST_1001,RT_20CNT,*up,4,0,
DR_1002_20CNT,DST_1002,RT_20CNT,*up,4,0,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingPlans.csv
	if err := writeFile(utils.RatingPlansCsv, `
#Id,DestinationRatesId,TimingTag,Weight
RP_1001,DR_1002_20CNT,*any,10
RP_1002,DR_1001_20CNT,*any,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingProfiles.csv
	if err := writeFile(utils.RatingProfilesCsv, `
#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_1001,
cgrates.org,call,1002,2014-01-14T00:00:00Z,RP_1002,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Rates.csv
	if err := writeFile(utils.RatesCsv, `
#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_20CNT,0.4,0.2,60s,60s,0s
RT_20CNT,0,0.1,60s,1s,60s
`); err != nil {
		t.Fatal(err)
	}

	var reply string
	args := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestSesRnd2IT"}
	if err := sesRnd2RPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testSesRnd2PrepaidInit(t *testing.T) {
	argsInit := &sessions.V1InitSessionArgs{
		InitSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PrepaidInit",
			Event: map[string]any{
				utils.OriginID:         "session1",
				utils.Tenant:           "cgrates.org",
				utils.Category:         utils.CALL,
				utils.ToR:              utils.VOICE,
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "1001",
				utils.Subject:          "1001",
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:       time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.Usage:            time.Minute,
				utils.CGRDebitInterval: 0,
			},
		},
	}

	var replyInit sessions.V1InitSessionReply
	if err := sesRnd2RPC.Call(utils.SessionSv1InitiateSession,
		argsInit, &replyInit); err != nil {
		t.Error(err)
	}
}
func testSesRnd2PrepaidUpdate1(t *testing.T) {
	argsUpdate := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PrepaidUpdate1",
			Event: map[string]any{
				utils.OriginID:         "session1",
				utils.Tenant:           "cgrates.org",
				utils.Category:         utils.CALL,
				utils.ToR:              utils.VOICE,
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "1001",
				utils.Subject:          "1001",
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:       time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.Usage:            5 * time.Second,
				utils.CGRDebitInterval: 0,
			},
		},
	}

	var replyUpdate sessions.V1UpdateSessionReply
	if err := sesRnd2RPC.Call(utils.SessionSv1UpdateSession, argsUpdate, &replyUpdate); err != nil {
		t.Error(err)
	}
}
func testSesRnd2PrepaidUpdate2(t *testing.T) {
	argsUpdate := &sessions.V1UpdateSessionArgs{
		UpdateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PrepaidUpdate2",
			Event: map[string]any{
				utils.OriginID:         "session1",
				utils.Tenant:           "cgrates.org",
				utils.Category:         utils.CALL,
				utils.ToR:              utils.VOICE,
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "1001",
				utils.Subject:          "1001",
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:       time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.Usage:            2 * time.Second,
				utils.CGRDebitInterval: 0,
			},
		},
	}

	var replyUpdate sessions.V1UpdateSessionReply
	if err := sesRnd2RPC.Call(utils.SessionSv1UpdateSession, argsUpdate, &replyUpdate); err != nil {
		t.Error(err)
	}
}
func testSesRnd2PrepaidTerminate(t *testing.T) {
	argsTerminate := &sessions.V1TerminateSessionArgs{
		TerminateSession: true,
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PrepaidTerminate",
			Event: map[string]any{
				utils.OriginID:         "session1",
				utils.Tenant:           "cgrates.org",
				utils.Category:         utils.CALL,
				utils.ToR:              utils.VOICE,
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "1001",
				utils.Subject:          "1001",
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:       time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.CGRDebitInterval: 0,
			},
		},
	}
	var replyTerminate string
	if err := sesRnd2RPC.Call(utils.SessionSv1TerminateSession,
		argsTerminate, &replyTerminate); err != nil {
		t.Error(err)
	}
}

func testSesRnd2PrepaidProcessCDR(t *testing.T) {
	// content of args will be ignored, since SessionSv1.ProcessCDR will take the session data from cache
	argsProcessCDR := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PrepaidProcessCDR",
			Event: map[string]any{
				utils.OriginID:         "session1",
				utils.Tenant:           "cgrates.org",
				utils.Category:         utils.CALL,
				utils.ToR:              utils.VOICE,
				utils.RequestType:      utils.META_PREPAID,
				utils.Account:          "1001",
				utils.Subject:          "1001",
				utils.Destination:      "1002",
				utils.SetupTime:        time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:       time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.Usage:            0,
				utils.CGRDebitInterval: 0,
			},
		},
	}
	var replyProcessCDR string
	if err := sesRnd2RPC.Call(utils.SessionSv1ProcessCDR,
		argsProcessCDR, &replyProcessCDR); err != nil {
		t.Error(err)
	}
}
func testSesRnd2PostpaidProcessCDR(t *testing.T) {
	argsProcessCDR := &utils.CGREventWithArgDispatcher{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "testSesRnd2PostpaidProcessCDR",
			Event: map[string]any{
				utils.OriginID:    "session2",
				utils.Tenant:      "cgrates.org",
				utils.Category:    utils.CALL,
				utils.ToR:         utils.VOICE,
				utils.RequestType: utils.META_POSTPAID,
				utils.Account:     "1001",
				utils.Subject:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2023, time.February, 28, 8, 59, 50, 0, time.UTC),
				utils.AnswerTime:  time.Date(2023, time.February, 28, 9, 0, 0, 0, time.UTC),
				utils.Usage:       time.Minute + 7*time.Second,
			},
		},
	}
	var replyProcessCDR string
	if err := sesRnd2RPC.Call(utils.SessionSv1ProcessCDR,
		argsProcessCDR, &replyProcessCDR); err != nil {
		t.Error(err)
	}
}

func testSesRnd2CompareCosts(t *testing.T) {
	var cdrs []*engine.ExternalCDR
	req := utils.RPCCDRsFilter{OriginIDs: []string{"session1", "session2"}}
	if err := sesRnd2RPC.Call(utils.APIerSv2GetCDRs, req, &cdrs); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if len(cdrs) != 2 {
		t.Errorf("expected 2 cdrs, received <%d>", len(cdrs))
	} else {
		// Compare with the Cost within CostDetails, not the one from the main event for the
		// *postpaid session.
		var postpaidCD engine.EventCost
		err := json.Unmarshal([]byte(cdrs[1].CostDetails), &postpaidCD)
		if err != nil {
			t.Error(err)
		}
		var postpaidCost float64
		if postpaidCD.Cost != nil {
			postpaidCost = *postpaidCD.Cost
		}
		if cdrs[0].Cost != postpaidCost {
			t.Errorf("expected the costs to be equal, received: <%+v> and <%+v>",
				cdrs[0].Cost, postpaidCost)
		}
	}
}

func testSesRnd2StopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testSesRnd2CreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestSesRnd2IT", 0755); err != nil {
		t.Error(err)
	}
}
func testSesRnd2RemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestSesRnd2IT"); err != nil {
		t.Error(err)
	}
}
