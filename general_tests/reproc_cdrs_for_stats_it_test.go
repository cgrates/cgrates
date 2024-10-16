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
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	rpcdrsCfgPath  string
	rpcdrsCfg      *config.CGRConfig
	rpcdrsRpc      *birpc.Client
	rpcdrsConfDIR1 string //run tests for specific configuration
	rpcdrsConfDIR2 string //run tests for specific configuration
	CGRID          = utils.GenUUID()

	sTestsRpcdrs = []func(t *testing.T){
		testRpcdrsLoadConfig1,
		testRpcdrsInitDataDb,
		testRpcdrsResetStorDb,
		testRpcdrsStartEngine,
		testRpcdrsRpcConn,
		testRpcdrsLoadTP,
		testRpcdrsSetBalance,
		testRpcdrsCheckInitialBalance,
		testRpcdrsProcessFirstCDR,
		testRpcdrsCheckAccountBalancesAfterFirstProcessCDR,
		testRpcdrsProcessSecondCDR,
		testRpcdrsCheckAccountBalancesAfterSecondProcessCDR,
		testRpcdrsGetCDRs,
		testRpcdrsStopEngine,
		testRpcdrsCreateDirectory,
		testRpcdrsNewEngineSameDB,
		testRpcdrsReprocessCDRs,
		testRpcdrsCheckAccountBalancesAfterSecondProcessCDR,
		testRpcdrsGetQueueStringMetrics,
		testRpcdrsStopEngine,
		testCsvVerifyExports,
		testRpcdrsRemoveDirectory,
	}
)

func TestReProcCDRs(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		rpcdrsConfDIR1 = "rerate_cdrs_mysql"
		rpcdrsConfDIR2 = "reprocess_cdrs_stats_ees_mysql"
	case utils.MetaMongo:
		rpcdrsConfDIR1 = "rerate_cdrs_mongo"
		rpcdrsConfDIR2 = "reprocess_cdrs_stats_ees_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRpcdrs {
		t.Run(rpcdrsConfDIR1, stest)
	}
}

func testRpcdrsLoadConfig1(t *testing.T) {
	var err error
	rpcdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", rpcdrsConfDIR1)
	if rpcdrsCfg, err = config.NewCGRConfigFromPath(rpcdrsCfgPath); err != nil {
		t.Error(err)
	}
}

func testRpcdrsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rpcdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRpcdrsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rpcdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRpcdrsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rpcdrsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testRpcdrsRpcConn(t *testing.T) {
	rpcdrsRpc = engine.NewRPCClient(t, rpcdrsCfg.ListenCfg())
}

func testRpcdrsLoadTP(t *testing.T) {
	engine.LoadCSVs(t, rpcdrsRpc, path.Join(*utils.DataDir, "tariffplans", "reratecdrs"), nil)
}

func testRpcdrsSetBalance(t *testing.T) {
	var reply string
	if err := rpcdrsRpc.Call(context.Background(), utils.APIerSv2SetBalance,
		utils.AttrSetBalance{
			Tenant:      "cgrates.org",
			Account:     "1001",
			Value:       float64(time.Minute),
			BalanceType: utils.MetaVoice,
			Balance: map[string]any{
				utils.ID: "voiceBalance1",
			},
		}, &reply); err != nil {
		t.Fatal(err)
	}
}

func testRpcdrsCheckInitialBalance(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaVoice: {
				{
					ID:    "voiceBalance1",
					Value: float64(time.Minute),
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rpcdrsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Fatalf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRpcdrsProcessFirstCDR(t *testing.T) {
	var reply string
	err := rpcdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags: []string{utils.MetaRALs},
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event1",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{
					utils.RunID:        "run_1",
					utils.CGRID:        CGRID,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "processCDR1",
					utils.OriginHost:   "OriginHost1",
					utils.RequestType:  utils.MetaPseudoPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2021, time.February, 2, 16, 14, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2021, time.February, 2, 16, 15, 0, 0, time.UTC),
					utils.Usage:        2 * time.Minute,
				},
			},
		}, &reply)
	if err != nil {
		t.Fatal(err)
	}

	var cdrs []*engine.CDR
	err = rpcdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_1"},
		}}, &cdrs)
	if err != nil {
		t.Fatal(err)
	}
	if cdrs[0].Usage != 2*time.Minute {
		t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	}
}

func testRpcdrsCheckAccountBalancesAfterFirstProcessCDR(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaVoice: {
				{
					ID:    "voiceBalance1",
					Value: 0,
				},
			},
			utils.MetaMonetary: {
				{
					ID:    utils.MetaDefault,
					Value: -0.6,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := rpcdrsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Fatal(err)
	}

	expAcnt.UpdateTime = acnt.UpdateTime
	expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
	expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
	acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
	if !reflect.DeepEqual(acnt, expAcnt) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
	}
}

func testRpcdrsProcessSecondCDR(t *testing.T) {
	var reply string
	err := rpcdrsRpc.Call(context.Background(), utils.CDRsV1ProcessEvent,
		&engine.ArgV1ProcessEvent{
			Flags: []string{utils.MetaRALs},
			CGREvent: utils.CGREvent{
				Tenant: "cgrates.org",
				ID:     "event2",
				Time:   utils.TimePointer(time.Now()),
				Event: map[string]any{
					utils.RunID:        "run_2",
					utils.CGRID:        CGRID,
					utils.Tenant:       "cgrates.org",
					utils.Category:     "call",
					utils.ToR:          utils.MetaVoice,
					utils.OriginID:     "processCDR2",
					utils.OriginHost:   "OriginHost2",
					utils.RequestType:  utils.MetaPseudoPrepaid,
					utils.AccountField: "1001",
					utils.Destination:  "1002",
					utils.SetupTime:    time.Date(2021, time.February, 2, 15, 14, 50, 0, time.UTC),
					utils.AnswerTime:   time.Date(2021, time.February, 2, 15, 15, 0, 0, time.UTC),
					utils.Usage:        2 * time.Minute,
				},
			},
		}, &reply)
	if err != nil {
		t.Fatal(err)
	}
	var cdrs []*engine.CDR
	err = rpcdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_2"},
		}}, &cdrs)
	if err != nil {
		t.Fatal(err)
	}

	if cdrs[0].Usage != 2*time.Minute {
		t.Errorf("expected usage to be <%+v>, received <%+v>", 2*time.Minute, cdrs[0].Usage)
	} else if cdrs[0].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[0].Cost)
	}
}

func testRpcdrsCheckAccountBalancesAfterSecondProcessCDR(t *testing.T) {
	expAcnt := engine.Account{
		ID: "cgrates.org:1001",
		BalanceMap: map[string]engine.Balances{
			utils.MetaVoice: {
				{
					ID:    "voiceBalance1",
					Value: 0,
				},
			},
			utils.MetaMonetary: {
				{
					ID:    utils.MetaDefault,
					Value: -1.8,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	err := rpcdrsRpc.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt)
	if err != nil {
		t.Fatal(err)
	}

	expAcnt.UpdateTime = acnt.UpdateTime
	expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
	expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
	acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
	if !reflect.DeepEqual(acnt, expAcnt) {
		t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
	}
}

var cdrs []*engine.CDR

func testRpcdrsGetCDRs(t *testing.T) {
	err := rpcdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			OrderBy: utils.AnswerTime,
		}}, &cdrs)
	if err != nil {
		t.Fatal(err)
	}
}

func testRpcdrsStopEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testRpcdrsCreateDirectory(t *testing.T) {
	if err := os.RemoveAll("/tmp/testCSV"); err != nil {
		t.Fatal("Error removing folder: ", "/tmp/testCSV", err)
	}
	if err := os.MkdirAll("/tmp/testCSV", 0755); err != nil {
		t.Fatal("Error creating folder: ", "/tmp/testCSV", err)
	}
}

func testRpcdrsNewEngineSameDB(t *testing.T) {
	var err error
	rpcdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", rpcdrsConfDIR2)
	if rpcdrsCfg, err = config.NewCGRConfigFromPath(rpcdrsCfgPath); err != nil {
		t.Error(err)
	}

	tpFiles := map[string]string{
		utils.StatsCsv: `#Tenant[0],Id[1],FilterIDs[2],ActivationInterval[3],QueueLength[4],TTL[5],MinItems[6],Metrics[7],MetricFilterIDs[8],Stored[9],Blocker[10],Weight[11],ThresholdIDs[12]
cgrates.org,STAT_AGG,,2014-07-29T15:00:00Z,0,-1,0,*tcd;*tcc;*sum#1,,false,false,30,*none`,
	}

	if _, err := engine.StopStartEngine(rpcdrsCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}

	rpcdrsRpc = engine.NewRPCClient(t, rpcdrsCfg.ListenCfg())
	engine.LoadCSVs(t, rpcdrsRpc, "", tpFiles)

}

func testRpcdrsReprocessCDRs(t *testing.T) {
	var reply string
	if err := rpcdrsRpc.Call(context.Background(), utils.CDRsV1ReprocessCDRs, &engine.ArgRateCDRs{
		Flags: []string{"*stats"},
		RPCCDRsFilter: utils.RPCCDRsFilter{
			OrderBy: utils.AnswerTime,
		}}, &reply); err != nil {
		t.Fatal(err)
	} else if reply != utils.OK {
		t.Fatalf("expected <OK>, received <%v>", reply)
	}

	var cdrsRerated []*engine.CDR
	err := rpcdrsRpc.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			OrderBy: utils.AnswerTime,
		}}, &cdrsRerated)
	if err != nil {
		t.Fatal(err)
	}

	if utils.ToJSON(cdrs) != utils.ToJSON(cdrsRerated) {
		t.Errorf("expected <%v>, \nreceived\n<%v>", utils.ToJSON(cdrs), utils.ToJSON(cdrsRerated))
	}
}

func testRpcdrsGetQueueStringMetrics(t *testing.T) {
	expectedMetrics := map[string]string{
		"*sum#1": "2",
		"*tcc":   "1.8",
		"*tcd":   "4m0s",
	}
	var metrics map[string]string
	if err := rpcdrsRpc.Call(context.Background(), utils.StatSv1GetQueueStringMetrics,
		&utils.TenantIDWithAPIOpts{
			TenantID: &utils.TenantID{Tenant: "cgrates.org", ID: "STAT_AGG"}}, &metrics); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(expectedMetrics, metrics) {
		t.Errorf("expecting: %+v, received reply: %s", expectedMetrics, metrics)
	}
}

func testCsvVerifyExports(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testCSV/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "STAT_AGG,120000000000,1.2,1\nSTAT_AGG,240000000000,1.8,2\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eCnt) != len(string(outContent1)) {
		t.Errorf("Expecting: \n<%+v>, \nreceived: \n<%+v>", len(eCnt), len(string(outContent1)))
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}

func testRpcdrsRemoveDirectory(t *testing.T) {
	if err := os.RemoveAll("/tmp/testCSV"); err != nil {
		t.Fatal("Error removing folder: ", "/tmp/testCSV", err)
	}
}
