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
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	ng1CfgPath, ng2CfgPath string
	ng1Cfg, ng2Cfg         *config.CGRConfig
	ng1RPC, ng2RPC         *birpc.Client
	ng1ConfDIR, ng2ConfDIR string //run tests for specific configuration
	rrDelay                int
	ng1UUID                = utils.GenUUID()
	ng2UUID                = utils.GenUUID()

	rrTests = []func(t *testing.T){
		testRerateExpLoadConfig,
		testRerateExpInitDataDb,
		testRerateExpResetStorDb,
		testRerateExpStartEngine,
		testRerateExpRPCConn,
		testRerateExpLoadTP,
		testRerateExpSetBalance,
		testRerateExpGetAccountAfterBalanceSet,
		testRerateExpProcessEventCDR1, // *
		testRerateExpCheckCDRCostAfterProcessEvent1,
		testRerateExpGetAccountAfterProcessEvent1,
		testRerateExpProcessEventCDR2, // *
		testRerateExpCheckCDRCostAfterProcessEvent2,
		testRerateExpGetAccountAfterProcessEvent2,
		testRerateExpProcessEventCDR3, // **
		testRerateExpCheckCDRCostAfterProcessEvent3,
		testRerateExpGetAccountAfterProcessEvent3,
		testRerateExpRerateCDRs, // ***
		testRerateExpCheckCDRCostsAfterRerate,
		testRerateExpGetAccountAfterRerate,
		testRerateExpStopEngine,
	}

	// *	-CDRs are processed and debited on the first engine, then exported and stored on second engine's storDB;
	// **	-Mimics the storage of external unprocessed cdrs (second engine's storDB);
	// ***	-Rerates all the existing cdrs (processed or not), ordering them by AnswerTime.
)

func TestRerateExpIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		ng1ConfDIR = "rerate_exp_engine1_internal"
		ng2ConfDIR = "rerate_exp_engine2_internal"
	case utils.MetaMySQL:
		ng1ConfDIR = "rerate_exp_engine1_mysql"
		ng2ConfDIR = "rerate_exp_engine2_mysql"
	case utils.MetaMongo:
		ng1ConfDIR = "rerate_exp_engine1_mongo"
		ng2ConfDIR = "rerate_exp_engine2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range rrTests {
		t.Run(ng1ConfDIR, stest)
	}
}

func testRerateExpLoadConfig(t *testing.T) {
	var err error
	ng1CfgPath = path.Join(*utils.DataDir, "conf", "samples", "rerate_exp_multiple_engines", ng1ConfDIR)
	if ng1Cfg, err = config.NewCGRConfigFromPath(ng1CfgPath); err != nil {
		t.Error(err)
	}
	ng2CfgPath = path.Join(*utils.DataDir, "conf", "samples", "rerate_exp_multiple_engines", ng2ConfDIR)
	if ng2Cfg, err = config.NewCGRConfigFromPath(ng2CfgPath); err != nil {
		t.Error(err)
	}
	rrDelay = 1000
}

func testRerateExpInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(ng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitDataDb(ng2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateExpResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(ng1Cfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(ng2Cfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateExpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(ng1CfgPath, rrDelay); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.StartEngine(ng2CfgPath, rrDelay); err != nil {
		t.Fatal(err)
	}
}

func testRerateExpRPCConn(t *testing.T) {
	ng1RPC = engine.NewRPCClient(t, ng1Cfg.ListenCfg())
	ng2RPC = engine.NewRPCClient(t, ng2Cfg.ListenCfg())
}

func testRerateExpLoadTP(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "reratecdrs")}
	if err := ng1RPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(200 * time.Millisecond)
}

func testRerateExpStopEngine(t *testing.T) {
	if err := engine.KillEngine(rrDelay); err != nil {
		t.Error(err)
	}
}

func testRerateExpSetBalance(t *testing.T) {
	attrSetBalance := utils.AttrSetBalance{
		Tenant:      "cgrates.org",
		Account:     "1001",
		Value:       float64(time.Minute),
		BalanceType: utils.MetaVoice,
		Balance: map[string]any{
			utils.ID: "1001",
		},
	}
	var reply string
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testRerateExpGetAccountAfterBalanceSet(t *testing.T) {
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
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateExpProcessEventCDR1(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, utils.MetaExport},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.RunID:        "run_1",
				utils.CGRID:        ng1UUID,
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
	if err := ng1RPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
}

func testRerateExpCheckCDRCostAfterProcessEvent1(t *testing.T) {
	var cdrs []*engine.CDR
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
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

func testRerateExpGetAccountAfterProcessEvent1(t *testing.T) {
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
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
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

func testRerateExpProcessEventCDR2(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, utils.MetaExport},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]any{
				utils.RunID:        "run_2",
				utils.CGRID:        ng1UUID,
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
	if err := ng1RPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(50 * time.Millisecond)
}

func testRerateExpCheckCDRCostAfterProcessEvent2(t *testing.T) {
	var cdrs []*engine.CDR
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
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

func testRerateExpGetAccountAfterProcessEvent2(t *testing.T) {
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
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
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

func testRerateExpProcessEventCDR3(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{"*rals:false"},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event3",
			Event: map[string]any{
				utils.RunID:        "run_3",
				utils.CGRID:        ng2UUID,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "processCDR3",
				utils.OriginHost:   "OriginHost3",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2021, time.February, 2, 17, 14, 50, 0, time.UTC),
				utils.AnswerTime:   time.Date(2021, time.February, 2, 17, 15, 0, 0, time.UTC),
				utils.Usage:        1 * time.Minute,
			},
		},
	}
	var reply string
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testRerateExpCheckCDRCostAfterProcessEvent3(t *testing.T) {
	var cdrs []*engine.CDR
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_2"},
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Usage != 2*time.Minute {
		t.Errorf("expected usage to be <%+v>, received <%+v>", 1*time.Minute, cdrs[0].Usage)
	} else if cdrs[0].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	}
}

func testRerateExpGetAccountAfterProcessEvent3(t *testing.T) {
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
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
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

func testRerateExpRerateCDRs(t *testing.T) {
	var reply string
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		Flags: []string{utils.MetaRerate},
		RPCCDRsFilter: utils.RPCCDRsFilter{
			OrderBy: utils.AnswerTime,
			CGRIDs:  []string{ng1UUID, ng2UUID},
		}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testRerateExpCheckCDRCostsAfterRerate(t *testing.T) {
	var cdrs []*engine.CDR
	if err := ng2RPC.Call(context.Background(), utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			CGRIDs:  []string{ng1UUID, ng2UUID},
			OrderBy: utils.AnswerTime,
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	} else if cdrs[1].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[1].Cost)
	} else if cdrs[2].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[2].Cost)
	}
}

func testRerateExpGetAccountAfterRerate(t *testing.T) {
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
					Value: -2.4,
				},
			},
		},
	}
	var acnt engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := ng1RPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
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
