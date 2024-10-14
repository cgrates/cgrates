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
	"path/filepath"
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
	rrErsCdrsUUID    = "38e4d9f4-577f-4260-a7a5-bae8cd5417de"
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
		testRerateCDRsERsGetCDRs1,
		testRerateCDRsERsExport,
		testRerateCDRsERsMoveFiles,
		testRerateCDRsERsGetCDRs2,
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
	folders := []string{"/tmp/ers/in", "/tmp/ees/mv", "/tmp/ers/out"}

	for _, folder := range folders {
		err := os.MkdirAll(folder, 0755)
		if err != nil {
			t.Fatalf("Failed to create folder %s: %v", folder, err)
		}
	}
}

func testRerateCDRsERsDeleteFolders(t *testing.T) {
	time.Sleep(5 * time.Second)
	folders := []string{"/tmp/ers/in", "/tmp/ees/mv", "/tmp/ers/out"}

	for _, folder := range folders {
		err := os.RemoveAll(folder)
		if err != nil {
			t.Fatalf("Failed to delete folder %s: %v", folder, err)
		}
	}
}

func testRerateCDRsERsLoadConfig(t *testing.T) {
	var err error
	rrErsCdrsCfgPath = path.Join(*utils.DataDir, "conf", "samples", "ers_rerate")
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
	rrErsCdrsRPC = engine.NewRPCClient(t, rrErsCdrsCfg.ListenCfg())
}

func testRerateCDRsERsLoadTP(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "reratecdrs")}
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
			t.Errorf("expected: <%+v>,\nreceived: \n<%+v>", utils.ToJSON(expAcnt), utils.ToJSON(acnt))
		}
	}
}

func testRerateCDRsERsProcessEventCDR1(t *testing.T) {
	argsEv := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaRALs, "*export:false"},
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

func testRerateCDRsERsGetCDRs1(t *testing.T) {
	rpsCdrFltr := &utils.RPCCDRsFilter{}

	var replies []*engine.ExternalCDR
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv2GetCDRs, rpsCdrFltr, &replies); err != nil {
		t.Error(err)
	}
	if len(replies) != 1 {
		t.Fatalf("Expected 1 reply, received \n<%+v>", utils.ToJSON(replies))
	}

	if reply, err := engine.NewCDRFromExternalCDR(replies[0], utils.EmptyString); err != nil {
		t.Error(err)
	} else if reply.Usage == 2*time.Minute {
		cdrEvent = reply.AsCGREvent()
		cdrEvent.Event[utils.Usage] = 1 * time.Minute
	} else {
		t.Errorf("Expected Usage <%+v>, Received CDR\n<%+v>", 2*time.Minute, utils.ToJSON(reply))
	}

}

func testRerateCDRsERsExport(t *testing.T) {
	cgrEv := &engine.CGREventWithEeIDs{
		CGREvent: cdrEvent,
	}
	exp := map[string]map[string]any{
		"CSVExporter": {},
	}
	var reply map[string]map[string]any
	if err := rrErsCdrsRPC.Call(context.Background(), utils.EeSv1ProcessEvent, cgrEv, &reply); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(reply, exp) {
		t.Errorf("Expected <%+v>, received \n<%+v>", exp, reply)
	}
}

func testRerateCDRsERsMoveFiles(t *testing.T) {
	time.Sleep(1 * time.Second)
	// Move all files from /tmp/ees/mv to /tmp/ers/in
	srcDir := "/tmp/ees/mv"
	destDir := "/tmp/ers/in"
	fileInfos, err := os.ReadDir(srcDir)
	if err != nil {
		t.Fatalf("Error reading source directory: %v", err)
	}

	for _, fileInfo := range fileInfos {
		srcPath := filepath.Join(srcDir, fileInfo.Name())
		destPath := filepath.Join(destDir, fileInfo.Name())
		if err := os.Rename(srcPath, destPath); err != nil {
			t.Fatalf("Error moving file: %v", err)
		}
	}
	time.Sleep(1 * time.Second)

}

func testRerateCDRsERsGetCDRs2(t *testing.T) {
	rpsCdrFltr := &utils.RPCCDRsFilter{}

	var replies []*engine.ExternalCDR
	if err := rrErsCdrsRPC.Call(context.Background(), utils.APIerSv2GetCDRs, rpsCdrFltr, &replies); err != nil {
		t.Error(err)
	}
	if len(replies) != 1 {
		t.Fatalf("Expected 1 reply, received \n<%+v>", utils.ToJSON(replies))
	}

	if reply, err := engine.NewCDRFromExternalCDR(replies[0], utils.EmptyString); err != nil {
		t.Error(err)
	} else if reply.Usage != 1*time.Minute {
		t.Errorf("Expected Usage <%+v>, Received CDR\n<%+v>", 1*time.Minute, utils.ToJSON(reply))
	}
}
