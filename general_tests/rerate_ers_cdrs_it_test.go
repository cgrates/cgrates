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
	"log"
	"os"
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
	rrErsCdrsCfgPath string
	rrErsCdrsCfg     *config.CGRConfig
	rrErsCdrsRPC     *birpc.Client
	rrErsCdrsDelay   int
	rrErsCdrsUUID    = utils.GenUUID()
	cdrEvent         *utils.CGREvent

	rrErsCdrsTests = []func(t *testing.T){
		testRerateCDRsERsCreateFolders,
		testRerateCDRsERsLoadConfig,
		testRerateCDRsERsInitDataDb,
		testRerateCDRsERsResetStorDb,
		testRerateCDRsERsStartEngine,
		testRerateCDRsERsRPCConn,
		testRerateCDRsERsLoadTP,
		testRerateCDRsERsSetBalance,
		testRerateCDRsERsGetAccountAfterBalanceSet,
		testRerateCDRsERsProcessEventCDR1,
		testRerateCDRsERsGetCDRs,
		testRerateCDRsERsExport,
		testRerateCDRsERsStopEngine,
		testRerateCDRsERsDeleteFolders,
	}
)

func TestReRateCDRsERs(t *testing.T) {
	t.Skip()
	for _, stest := range rrErsCdrsTests {
		t.Run("ers_rerate", stest)
	}
}

func testRerateCDRsERsCreateFolders(t *testing.T) {
	inPath := "/tmp/ers/in"
	outPath := "/tmp/ers/out"

	// Create the /tmp/ers/in folder
	err := os.MkdirAll(inPath, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create %s: %s", inPath, err)
	}

	// Create the /tmp/ers/out folder
	err = os.MkdirAll(outPath, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create %s: %s", outPath, err)
	}

	t.Log("Created folders successfully")
}

func testRerateCDRsERsDeleteFolders(t *testing.T) {
	time.Sleep(5 * time.Second)
	inPath := "/tmp/ers/in"
	outPath := "/tmp/ers/out"

	// Remove the /tmp/ers/in folder
	err := os.RemoveAll(inPath)
	if err != nil {
		t.Fatalf("Failed to delete %s: %s", inPath, err)
	}

	// Remove the /tmp/ers/out folder
	err = os.RemoveAll(outPath)
	if err != nil {
		t.Fatalf("Failed to delete %s: %s", outPath, err)
	}

	t.Log("Deleted folders successfully")
}

func testRerateCDRsERsLoadConfig(t *testing.T) {
	var err error
	rrErsCdrsCfgPath = path.Join(*dataDir, "conf", "samples", "ers_rerate")
	if rrErsCdrsCfg, err = config.NewCGRConfigFromPath(rrErsCdrsCfgPath); err != nil {
		t.Error(err)
	}
	rrErsCdrsDelay = 1000
}

func testRerateCDRsERsInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(rrErsCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsERsResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(rrErsCdrsCfg); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsERsStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(rrErsCdrsCfgPath, rrErsCdrsDelay); err != nil {
		t.Fatal(err)
	}
}

func testRerateCDRsERsRPCConn(t *testing.T) {
	var err error
	rrErsCdrsRPC, err = newRPCClient(rrErsCdrsCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal("Could not connect to rater: ", err.Error())
	}
}

func testRerateCDRsERsLoadTP(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "reratecdrs")}
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned", reply)
	}
	time.Sleep(200 * time.Millisecond)
}

func testRerateCDRsERsStopEngine(t *testing.T) {
	if err := engine.KillEngine(rrErsCdrsDelay); err != nil {
		t.Error(err)
	}
}

func testRerateCDRsERsSetBalance(t *testing.T) {
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
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv2SetBalance, attrSetBalance, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Received: %s", reply)
	}
}

func testRerateCDRsERsGetAccountAfterBalanceSet(t *testing.T) {
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
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv2GetAccount, attrs, &acnt); err != nil {
		t.Error(err)
	} else {
		expAcnt.UpdateTime = acnt.UpdateTime
		expAcnt.BalanceMap[utils.MetaVoice][0].Uuid = acnt.BalanceMap[utils.MetaVoice][0].Uuid
		if !reflect.DeepEqual(acnt, expAcnt) {
			t.Errorf("expected: <%+v>,\nreceived: <%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsERsProcessEventCDR1(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs},
		CGREvent: utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "event1",
			Event: map[string]any{
				utils.RunID:        "run_1",
				utils.CGRID:        rrErsCdrsUUID,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "processCDR1",
				utils.OriginHost:   "OriginHost1",
				utils.RequestType:  utils.MetaPseudoPrepaid,
				utils.Subject:      "1001",
				utils.AccountField: "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2023, time.October, 11, 16, 14, 50, 0, time.UTC),
				utils.AnswerTime:   time.Date(2023, time.October, 11, 16, 15, 0, 0, time.UTC),
				utils.Usage:        2 * time.Minute,
			},
		},
	}
	var reply string
	if err := rrErsCdrsRPC.Call(context.Background(), utils.CDRsV1ProcessEvent, argsEv, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}

}

func testRerateCDRsERsGetCDRs(t *testing.T) {
	attrs := &utils.RPCCDRsFilter{}

	var replies []*engine.ExternalCDR
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv2GetCDRs, attrs, &replies); err != nil {
		t.Error(err)
	}

	log.Printf("APIerSv2GetCDRsreply []*engine.ExternalCDR <%+v>", utils.ToJSON(replies))

	if len(replies) == 1 {
		if reply, err := engine.NewCDRFromExternalCDR(replies[0], utils.EmptyString); err != nil {
			t.Error(err)
		} else if reply != nil {
			cdrEvent = reply.AsCGREvent()
			log.Printf("\nreply <%+v>\n", utils.ToJSON(cdrEvent))
		}
	} else {
		t.Error("More than 1 reply")
	}

}

func testRerateCDRsERsExport(t *testing.T) {
	cgrEv := &engine.CGREventWithEeIDs{
		CGREvent: cdrEvent,
	}
	log.Printf("cgrEv <%+v>", utils.ToJSON(cgrEv))
	var reply map[string]map[string]any
	if err := rrErsCdrsRPC.Call(context.Background(), utils.EeSv1ProcessEvent, cgrEv, &reply); err != nil {
		t.Error(err)
	}

	log.Printf("EeSv1ProcessEvent reply <%+v>", reply)

}
