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
	"errors"
	"math"
	"net/rpc"
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rrCdrsCfgPath string
	rrCdrsCfg     *config.CGRConfig
	rrCdrsRPC     *rpc.Client
	rrCdrsConfDIR string //run tests for specific configuration
	rrCdrsDelay   int
	rrCdrsUUID    = utils.GenUUID()

	rrCdrsTests = []func(t *testing.T){
		testRerateCDRsRemoveFolders,
		testRerateCDRsCreateFolders,
		testRerateCDRsLoadConfig,
		testRerateCDRsInitDataDb,
		testRerateCDRsResetStorDb,
		testRerateCDRsStartEngine,
		testRerateCDRsRPCConn,
		testRerateCDRsLoadTPs,

		testRerateCDRsSetBalance,
		testRerateCDRsGetAccountAfterBalanceSet,

		testRerateCDRsProcessEventCDR1,
		testRerateCDRsCheckCDRCostAfterProcessEvent1,
		testRerateCDRsGetAccountAfterProcessEvent1,

		testRerateCDRsProcessEventCDR2,
		testRerateCDRsCheckCDRCostAfterProcessEvent2,
		testRerateCDRsGetAccountAfterProcessEvent2,

		testRerateCDRsRerateCDRs,
		testRerateCDRsCheckCDRCostsAfterRerate,
		testRerateCDRsGetAccountAfterRerate,

		testRerateCDRsStopEngine,
		testRerateCDRsRemoveFolders,
	}
)

// Test start here
func TestRerateCDRs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		rrCdrsConfDIR = "rerate_cdrs_internal"
	case utils.MetaMySQL:
		rrCdrsConfDIR = "rerate_cdrs_mysql"
	case utils.MetaMongo:
		rrCdrsConfDIR = "rerate_cdrs_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range rrCdrsTests {
		t.Run(rrCdrsConfDIR, stest)
	}
}

func testRerateCDRsLoadConfig(t *testing.T) {
	var err error
	rrCdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", rrCdrsConfDIR)
	if rrCdrsCfg, err = config.NewCGRConfigFromPath(rrCdrsCfgPath); err != nil {
		t.Error(err)
	}
	rrCdrsDelay = 1000
}

func testRerateCDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rrCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rrCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rrCdrsCfgPath, rrCdrsDelay); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsRPCConn(t *testing.T) {
	var err error
	rrCdrsRPC, err = newRPCClient(rrCdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testRerateCDRsLoadTPs(t *testing.T) {
	writeFile := func(fileName, data string) error {
		csvFile, err := os.Create(path.Join("/tmp/TestRerateCDRs", fileName))
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

	// Create and populate DestinationRates.csv
	if err := writeFile(utils.DestinationRatesCsv, `
#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,20,0,
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate Rates.csv
	if err := writeFile(utils.RatesCsv, `
#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,0.6,60s,1s,0s
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingPlans.csv
	if err := writeFile(utils.RatingPlansCsv, `
#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10
`); err != nil {
		t.Fatal(err)
	}

	// Create and populate RatingProfiles.csv
	if err := writeFile(utils.RatingProfilesCsv, `
#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,2014-01-14T00:00:00Z,RP_ANY,	
`); err != nil {
		t.Fatal(err)
	}

	var loadInst string
	if err := rrCdrsRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: "/tmp/TestRerateCDRs"}, &loadInst); err != nil {
		t.Error(err)
	}
}

func testRerateCDRsStopEngine(t *testing.T) {
	if err := engine.KillEngine(rrCdrsDelay); err != nil {
		t.Error(err)
	}
}

func testRerateCDRsSetBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		Value:       float64(time.Minute),
		BalanceType: utils.VOICE,
		Balance: map[string]any{
			utils.ID: "1001",
		},
	}
	var reply string
	if err := rrCdrsRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testRerateCDRsGetAccountAfterBalanceSet(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.VOICE: {
				{
					ID:    "1001",
					Value: float64(time.Minute),
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.VOICE][0].Uuid = acnt.BalanceMap[utils.VOICE][0].Uuid
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsProcessEventCDR1(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.RunID:       "run_1",
				utils.CGRID:       rrCdrsUUID,
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "processCDR1",
				utils.OriginHost:  "OriginHost1",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
				utils.AnswerTime:  time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var reply string
	if err := rrCdrsRPC.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

}

func testRerateCDRsCheckCDRCostAfterProcessEvent1(t *testing.T) {
	var cdrs []*engine.CDR
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_1"},
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Usage != 2*time.Minute {
		t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	}
}

func testRerateCDRsGetAccountAfterProcessEvent1(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.VOICE: {
				{
					ID:    "1001",
					Value: 0,
				},
			},
			utils.MONETARY: {
				{
					ID:    utils.MetaDefault,
					Value: -0.6,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.VOICE][0].Uuid = acnt.BalanceMap[utils.VOICE][0].Uuid
		expAcnt.BalanceMap[utils.MONETARY][0].Uuid = acnt.BalanceMap[utils.MONETARY][0].Uuid
		acnt.BalanceMap[utils.MONETARY][0].Value = math.Round(acnt.BalanceMap[utils.MONETARY][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsProcessEventCDR2(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]any{
				utils.RunID:       "run_2",
				utils.CGRID:       rrCdrsUUID,
				utils.Tenant:      "cgrates.org",
				utils.Category:    "call",
				utils.ToR:         utils.VOICE,
				utils.OriginID:    "processCDR2",
				utils.OriginHost:  "OriginHost2",
				utils.RequestType: utils.META_PSEUDOPREPAID,
				utils.Account:     "1001",
				utils.Destination: "1002",
				utils.SetupTime:   time.Date(2021, time.February, 2, 15, 14, 50, 0, time.UTC),
				utils.AnswerTime:  time.Date(2021, time.February, 2, 15, 15, 0, 0, time.UTC),
				utils.Usage:       2 * time.Minute,
			},
		},
	}
	var reply string
	if err := rrCdrsRPC.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

}

func testRerateCDRsCheckCDRCostAfterProcessEvent2(t *testing.T) {
	var cdrs []*engine.CDR
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_2"},
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Usage != 2*time.Minute {
		t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
	} else if cdrs[0].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[0].Cost)
	}
}

func testRerateCDRsGetAccountAfterProcessEvent2(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.VOICE: {
				{
					ID:    "1001",
					Value: 0,
				},
			},
			utils.MONETARY: {
				{
					ID:    utils.MetaDefault,
					Value: -1.8,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.VOICE][0].Uuid = acnt.BalanceMap[utils.VOICE][0].Uuid
		expAcnt.BalanceMap[utils.MONETARY][0].Uuid = acnt.BalanceMap[utils.MONETARY][0].Uuid
		acnt.BalanceMap[utils.MONETARY][0].Value = math.Round(acnt.BalanceMap[utils.MONETARY][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsRerateCDRs(t *testing.T) {
	var reply string
	if err := rrCdrsRPC.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		Flags: []string{utils.MetaRerate},
		RPCCDRsFilter: utils.RPCCDRsFilter{
			OrderBy: utils.AnswerTime,
			CGRIDs:  []string{rrCdrsUUID},
		}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testRerateCDRsCheckCDRCostsAfterRerate(t *testing.T) {
	var cdrs []*engine.CDR
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithArgDispatcher{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			CGRIDs:  []string{rrCdrsUUID},
			OrderBy: utils.AnswerTime,
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	} else if cdrs[1].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[1].Cost)
	}
}

func testRerateCDRsGetAccountAfterRerate(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.VOICE: {
				{
					ID:    "1001",
					Value: 0,
				},
			},
			utils.MONETARY: {
				{
					ID:    utils.MetaDefault,
					Value: -1.8,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.VOICE][0].Uuid = acnt.BalanceMap[utils.VOICE][0].Uuid
		expAcnt.BalanceMap[utils.MONETARY][0].Uuid = acnt.BalanceMap[utils.MONETARY][0].Uuid
		acnt.BalanceMap[utils.MONETARY][0].Value = math.Round(acnt.BalanceMap[utils.MONETARY][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsCreateFolders(t *testing.T) {
	if err := os.MkdirAll("/tmp/TestRerateCDRs", 0755); err != nil {
		t.Error(err)
	}
}

func testRerateCDRsRemoveFolders(t *testing.T) {
	if err := os.RemoveAll("/tmp/TestRerateCDRs"); err != nil {
		t.Error(err)
	}
}

func TestRerateFailedCDR(t *testing.T) {
	jsonCfg := `{
"data_db": {
	"db_type": "*internal"
},
"stor_db": {
	"db_type": "*internal"
},
"rals": {
	"enabled": true
},
"cdrs": {
	"enabled": true,
	"rals_conns": ["*internal"]
},
"apiers": {
	"enabled": true
}
}`

	tpFiles := map[string]string{
		utils.DestinationRatesCsv: `#Id,DestinationId,RatesTag,RoundingMethod,RoundingDecimals,MaxCost,MaxCostStrategy
DR_ANY,*any,RT_ANY,*up,0,0,`,
		utils.RatesCsv: `#Id,ConnectFee,Rate,RateUnit,RateIncrement,GroupIntervalStart
RT_ANY,0,2,1s,1s,0s`,
		utils.RatingPlansCsv: `#Id,DestinationRatesId,TimingTag,Weight
RP_ANY,DR_ANY,*any,10`,
		utils.RatingProfilesCsv: `#Tenant,Category,Subject,ActivationTime,RatingPlanId,RatesFallbackSubject
cgrates.org,call,1001,,RP_ANY,`,
	}

	env := TestEnvironment{
		ConfigJSON: jsonCfg,
		TpFiles:    tpFiles,
	}
	client, _ := env.Setup(t, 0)

	balanceID := "test"
	processCDR := func(t *testing.T, from, to string, usage time.Duration, wantCost float64) {
		t.Helper()
		var reply string
		err := client.Call(utils.CDRsV1ProcessEvent,
			&engine.ArgV1ProcessEvent{
				Flags: []string{utils.MetaRALs},
				CGREvent: utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "event1",
					Event: map[string]any{
						utils.Tenant:      "cgrates.org",
						utils.Category:    "call",
						utils.ToR:         utils.VOICE,
						utils.OriginID:    "processCDR",
						utils.RequestType: utils.POSTPAID,
						utils.Account:     from,
						utils.Destination: to,
						utils.SetupTime:   time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
						utils.AnswerTime:  time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
						utils.Usage:       usage,
					},
				},
			}, &reply)
		if err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilter{
			OriginIDs: []string{"processCDR"},
		}, &cdrs); err != nil {
			t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
		}
		if cdrs[0].Cost != wantCost {
			t.Errorf("Cost=%v, want %v", cdrs[0].Cost, wantCost)
		}
		errMsg := "ACCOUNT_NOT_FOUND"
		if cdrs[0].Cost == -1 && cdrs[0].ExtraInfo != errMsg {
			t.Errorf("ExtraInfo err msg=%v, want %v", cdrs[0].ExtraInfo, errMsg)
		}
	}

	setAccount := func(t *testing.T, acc string, value float64) {
		t.Helper()
		var reply string
		if err := client.Call(utils.APIerSv2SetBalance,
			utils.AttrSetBalance{
				Tenant:      "cgrates.org",
				Account:     acc,
				Value:       value,
				BalanceType: utils.MONETARY,
				Balance: map[string]any{
					utils.ID: balanceID,
				},
			}, &reply); err != nil {
			t.Fatal(err)
		}

	}

	checkBalance := func(t *testing.T, acc string, want float64) {
		t.Helper()
		var acnt engine.Account
		err := client.Call(utils.APIerSv2GetAccount,
			&utils.AttrGetAccount{
				Tenant:  "cgrates.org",
				Account: acc,
			}, &acnt)
		if want == -1 {
			if err == nil || errors.Is(err, utils.ErrNotFound) {
				t.Fatalf("APIerSv2.GetAccount err=%v, want %v", err, utils.ErrNotFound)
			}
			return
		}
		if err != nil {
			t.Fatalf("APIerSv2GetAccount failed unexpectedly: %v", err)
		}
		got := acnt.BalanceMap[utils.MONETARY][0]
		if got == nil {
			t.Errorf("acnt.FindBalanceByID(%q) could not find balance", balanceID)
		} else if got.Value != want {
			t.Errorf("acnt.FindBalanceByID(%q) balance value=%v, want %v", balanceID, got.Value, want)
		}
	}

	rerateCDR := func(t *testing.T, cost float64) {
		t.Helper()
		var reply string
		if err := client.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
			Flags: []string{utils.MetaRerate},
			RPCCDRsFilter: utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR"},
			}}, &reply); err != nil {
			t.Fatal(err)
		}
		var cdrs []*engine.CDR
		if err := client.Call(utils.CDRsV1GetCDRs,
			&utils.RPCCDRsFilter{
				OriginIDs: []string{"processCDR"},
			}, &cdrs); err != nil {
			t.Errorf("CDRsV1GetCDRs failed unexpectedly: %v", err)
		}
		if cdrs[0].Cost != cost {
			t.Errorf("Cost=%v, want %v", cdrs[0].Cost, cost)
		}
	}

	checkBalance(t, "1001", -1)                      // account does not exist
	processCDR(t, "1001", "1002", 3*time.Second, -1) // processCDR fails due to non-existent acc
	setAccount(t, "1001", 10)                        // create acc with balance 10
	rerateCDR(t, 6)                                  // rerate the failed CDR
	checkBalance(t, "1001", 4)                       // check balance after rerate; expecting 4 (10-6)
}
