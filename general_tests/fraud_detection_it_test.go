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
	"os"
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
	fraudCfgPath string
	fraudCfg     *config.CGRConfig
	fraudRPC     *birpc.Client
	fraudDelay   int
	fraudConfDIR string

	sTestsFraud = []func(t *testing.T){
		testFraudRemoveFolders,
		testFraudCreateFolders,

		testFraudLoadConfig,
		testFraudInitDataDb,
		testFraudInitStorDb,
		testFraudStartEngine,
		testFraudRPCConn,
		testFraudLoadTarriffPlans,

		testFraudAuthorizeandProcess1,
		testFraudAuthorizeandProcess2,
		testFraudAuthorizeandProcess3,
		testFraudFinalAuthorize,

		testFraudStopEngine,
		testFraudRemoveFolders,
	}
)

func TestFraudIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		fraudConfDIR = "fraud_internal"
	case utils.MetaMySQL:
		fraudConfDIR = "fraud_mysql"
	case utils.MetaMongo:
		fraudConfDIR = "fraud_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsFraud {
		t.Run(fraudConfDIR, stest)
	}
}

func testFraudCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestFraudIT", 0755); err != nil {
		t.Error(err)
	}
}

func testFraudRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestFraudIT"); err != nil {
		t.Error(err)
	}
}

func testFraudLoadConfig(t *testing.T) {
	var err error
	fraudCfgPath = path.Join(*utils.DataDir, "conf", "samples", fraudConfDIR)
	if fraudCfg, err = config.NewCGRConfigFromPath(fraudCfgPath); err != nil {
		t.Error(err)
	}
	fraudDelay = 1000
}

func testFraudInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(fraudCfg); err != nil {
		t.Fatal(err)
	}
}

func testFraudInitStorDb(t *testing.T) {
	if err := engine.InitStorDb(fraudCfg); err != nil {
		t.Fatal(err)
	}
}

func testFraudStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fraudCfgPath, fraudDelay); err != nil {
		t.Fatal(err)
	}
}

func testFraudRPCConn(t *testing.T) {
	fraudRPC = engine.NewRPCClient(t, fraudCfg.ListenCfg())
}

func testFraudLoadTarriffPlans(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestFraudIT", fileName))
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
ACT_FRD_LOG,*log,,,,,,,,,,,,,,,0
ACT_FRD_STOP,*disable_account,,,,,,,,,,,,,,,10
ACT_TOPUP_RST_10,*topup_reset,,,test,*monetary,,*any,,,*unlimited,,10,10,false,false,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Attributes.csv
	if err := writeFile(utils.AttributesCsv, `
#Tenant,ID,Context,FilterIDs,ActivationInterval,AttributeFilterIDs,Path,Type,Value,Blocker,Weight
cgrates.org,ATTR_FRD,*cdrs,,,,*opts.*account,*variable,~*req.Account,false,0
`); err != nil {
		t.Fatal(err)
	}

	// cgrates.org,Raw,,,*raw,*constant:*req.RequestType:*none,0
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

	// Create and populate Stats.csv
	if err := writeFile(utils.StatsCsv, `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,STATS_FRD,,,-1,24h,0,*tcc,,true,false,0,THD_FRD
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Thresholds.csv
	if err := writeFile(utils.ThresholdsCsv, `
#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],MaxHits[4],MinHits[5],MinSleep[6],Blocker[7],Weight[8],ActionIDs[9],Async[10]
cgrates.org,THD_FRD,*gte:~*req.*tcc:2,,-1,1,0,false,0,ACT_FRD_STOP;ACT_FRD_LOG,true
`); err != nil {
		t.Fatal(err)
	}

	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestFraudIT"}
	if err := fraudRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(500 * time.Millisecond)
}

func testFraudAuthorizeandProcess1(t *testing.T) {
	originID := utils.GenUUID()
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		false, false, false, cgrEv, utils.Paginator{}, false, "")

	var rply sessions.V1AuthorizeReply
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Error(err)
	}
	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        5 * time.Minute,
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		cgrEv, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}
}

func testFraudAuthorizeandProcess2(t *testing.T) {
	originID := utils.GenUUID()
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		false, false, false, cgrEv, utils.Paginator{}, false, "")

	var rply sessions.V1AuthorizeReply
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Error(err)
	}
	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        2 * time.Minute,
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		cgrEv, &reply); err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ErrPartiallyExecuted, err)
	}
}

func testFraudAuthorizeandProcess3(t *testing.T) {
	originID := utils.GenUUID()
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		false, false, false, cgrEv, utils.Paginator{}, false, "")

	var rply sessions.V1AuthorizeReply
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args, &rply); err != nil {
		t.Error(err)
	}
	cgrEv = &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     originID,
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
			utils.Usage:        3 * time.Minute,
		},
		APIOpts: map[string]any{},
	}
	var reply string
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1ProcessCDR,
		cgrEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received reply: %s", reply)
	}
}

func testFraudFinalAuthorize(t *testing.T) {
	cgrEv := &utils.CGREvent{
		Tenant: "cgrates.org",
		Event: map[string]any{
			utils.Tenant:       "cgrates.org",
			utils.Category:     "call",
			utils.ToR:          utils.MetaVoice,
			utils.OriginID:     utils.GenUUID(),
			utils.RequestType:  utils.MetaPseudoPrepaid,
			utils.AccountField: "1001",
			utils.Subject:      "1001",
			utils.Destination:  "1002",
			utils.SetupTime:    time.Date(2022, time.January, 7, 16, 60, 0, 0, time.UTC),
			utils.AnswerTime:   time.Date(2022, time.January, 7, 16, 60, 10, 0, time.UTC),
		},
		APIOpts: map[string]any{},
	}
	args := sessions.NewV1AuthorizeArgs(false, []string{},
		false, []string{}, false, []string{}, false, true,
		false, false, false, cgrEv, utils.Paginator{}, false, "")

	expErr := `RALS_ERROR:ACCOUNT_DISABLED`
	var rply sessions.V1AuthorizeReply
	if err := fraudRPC.Call(context.Background(), utils.SessionSv1AuthorizeEvent, args,
		&rply); err == nil || err.Error() != expErr {
		t.Error(err)
	}
}

func testFraudStopEngine(t *testing.T) {
	if err := engine.KillEngine(fraudDelay); err != nil {
		t.Error(err)
	}
}
