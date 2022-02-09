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
	"net/rpc"
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
		testRerateCDRsLoadConfig,
		testRerateCDRsInitDataDb,
		testRerateCDRsResetStorDb,
		testRerateCDRsStartEngine,
		testRerateCDRsRPCConn,
		testRerateCDRsLoadTP,
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
	}
)

// Test start here
func TestRerateCDRs(t *testing.T) {
	switch *dbType {
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
	rrCdrsCfgPath = path.Join(*dataDir, "conf", "samples", rrCdrsConfDIR)
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

func testRerateCDRsLoadTP(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "reratecdrs")}
	if err := rrCdrsRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(200 * time.Millisecond)
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
		BalanceType: utils.MetaVoice,
		Balance: map[string]interface{}{
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
			utils.MetaVoice: {
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
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
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
			Event: map[string]interface{}{
				utils.RunID:        "run_1",
				utils.CGRID:        rrCdrsUUID,
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
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
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
			utils.MetaVoice: {
				{
					ID:    "1001",
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
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
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
			Event: map[string]interface{}{
				utils.RunID:        "run_2",
				utils.CGRID:        rrCdrsUUID,
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
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
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
			utils.MetaVoice: {
				{
					ID:    "1001",
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
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsRerateCDRs(t *testing.T) {
	var reply string
	if err := rrCdrsRPC.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		Flags: []string{utils.MetaRALs},
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
	if err := rrCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
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
			utils.MetaVoice: {
				{
					ID:    "1001",
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
	if err := rrCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		expAcnt.BalanceMap[utils.MetaMonetary][0].Uuid = acnt.BalanceMap[utils.MetaMonetary][0].Uuid
		acnt.BalanceMap[utils.MetaMonetary][0].Value = math.Round(acnt.BalanceMap[utils.MetaMonetary][0].Value*10) / 10
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>, \nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}
