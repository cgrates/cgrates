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
	"net/rpc"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	psCdrsCfgPath string
	psCdrsCfg     *config.CGRConfig
	psCdrsRPC     *rpc.Client
	psCdrsConfDIR string //run tests for specific configuration
	psCdrsDelay   int
	psCdrsUUID    = utils.GenUUID()

	psCdrsTests = []func(t *testing.T){
		testPsCDRsLoadConfig,
		testPsCDRsInitDataDb,
		testPsCDRsResetStorDb,
		testPsCDRsStartEngine,
		testPsCDRsRPCConn,
		testPsCDRsLoadTP,
		testPsCDRsSetBalance,
		testPsCDRsProcessEventCDR1,
		testPsCDRsProcessEventCDR2,
		testPsCDRsGetAccount1,
		testPsCDRsRerateCDRs,
		testPsCDRsGetAccount2,
		testPsCDRsStopEngine,
	}
)

// Test start here
func TestPsCDRs(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		psCdrsConfDIR = "rerate_cdrs_mysql"
	case utils.MetaMongo:
		t.SkipNow()
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range psCdrsTests {
		t.Run(psCdrsConfDIR, stest)
	}
}

func testPsCDRsLoadConfig(t *testing.T) {
	var err error
	psCdrsCfgPath = path.Join(*dataDir, "conf", "samples", psCdrsConfDIR)
	if psCdrsCfg, err = config.NewCGRConfigFromPath(psCdrsCfgPath); err != nil {
		t.Error(err)
	}
	psCdrsDelay = 1000
}

func testPsCDRsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(psCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testPsCDRsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(psCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testPsCDRsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(psCdrsCfgPath, psCdrsDelay); err != nil {
		t.Fatal(err)
	}
}

func testPsCDRsRPCConn(t *testing.T) {
	var err error
	psCdrsRPC, err = newRPCClient(psCdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testPsCDRsLoadTP(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "reratecdrs")}
	if err := psCdrsRPC.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(200 * time.Millisecond)
}

func testPsCDRsStopEngine(t *testing.T) {
	if err := engine.KillEngine(psCdrsDelay); err != nil {
		t.Error(err)
	}
}

func testPsCDRsSetBalance(t *testing.T) {
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
	if err := psCdrsRPC.Call(utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testPsCDRsProcessEventCDR1(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]interface{}{
				utils.RunID:        "run_1",
				utils.CGRID:        psCdrsUUID,
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
	if err := psCdrsRPC.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := psCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_1"},
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	}
	// else {
	// 	fmt.Println("=================1", utils.ToJSON(cdrs))
	// }
}

func testPsCDRsProcessEventCDR2(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event2",
			Event: map[string]interface{}{
				utils.RunID:        "run_2",
				utils.CGRID:        psCdrsUUID,
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
	if err := psCdrsRPC.Call(utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	var cdrs []*engine.CDR
	if err := psCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			RunIDs: []string{"run_2"},
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[0].Cost)
	}
	// else {
	// 	fmt.Println("=================2", utils.ToJSON(cdrs))
	// }
}

func testPsCDRsGetAccount1(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := psCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	}
	// else {
	// 	fmt.Println("=================4", utils.ToJSON(acnt))
	// }
}

func testPsCDRsRerateCDRs(t *testing.T) {
	var reply string
	if err := psCdrsRPC.Call(utils.CDRsV1RateCDRs, &engine.ArgRateCDRs{
		Flags: []string{utils.MetaRerate, utils.MetaRALs},
		RPCCDRsFilter: utils.RPCCDRsFilter{
			CGRIDs: []string{psCdrsUUID},
		}}, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

	var cdrs []*engine.CDR
	if err := psCdrsRPC.Call(utils.CDRsV1GetCDRs, &utils.RPCCDRsFilterWithAPIOpts{
		RPCCDRsFilter: &utils.RPCCDRsFilter{
			CGRIDs:  []string{psCdrsUUID},
			OrderBy: utils.AnswerTime,
		}}, &cdrs); err != nil {
		t.Error(err)
	} else if cdrs[0].Cost != 0.6 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 0.6, cdrs[0].Cost)
	} else if cdrs[1].Cost != 1.2 {
		t.Errorf("expected cost to be <%+v>, received <%+v>", 1.2, cdrs[1].Cost)
	}
	// else {
	// 	fmt.Println("=================3", utils.ToJSON(cdrs))
	// }
}

func testPsCDRsGetAccount2(t *testing.T) {
	var acnt *engine.Account
	attrs := &utils.AttrGetAccount{Tenant: "cgrates.org", Account: "1001"}
	if err := psCdrsRPC.Call(utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	}
	// else {
	// 	fmt.Println("=================5", utils.ToJSON(acnt))
	// }
}
