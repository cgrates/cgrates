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
	"github.com/cgrates/cgrates/ees"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	cdrsPostFailCfgPath string
	cdrsPostFailCfg     *config.CGRConfig
	cdrsPostFailRpc     *birpc.Client
	cdrsPostFailConfDIR string // run the tests for specific configuration

	// subtests to be executed for each confDIR
	sTestsCDRsPostFailIT = []func(t *testing.T){
		testCDRsPostFailoverInitConfig,
		testCDRsPostFailoverInitDataDb,
		testCDRsPostFailoverInitCdrDb,
		testCDRsPostFailoverStartEngine,
		testCDRsPostFailoverRpcConn,
		testCDRsPostFailoverLoadTariffPlanFromFolder,
		testCDRsPostFailoverProcessCDR,
		testCDRsPostFailoverToFile,
		testCDRsPostFailoverKillEngine,
	}
)

// Tests starting here
func TestCDRsPostFailoverIT(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		cdrsPostFailConfDIR = "cdrsv_failover_internal"
	case utils.MetaMySQL:
		cdrsPostFailConfDIR = "cdrsv_failover_mysql"
	case utils.MetaMongo:
		cdrsPostFailConfDIR = "cdrsv_failover_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsCDRsPostFailIT {
		t.Run(cdrsPostFailConfDIR, stest)
	}
}

func testCDRsPostFailoverInitConfig(t *testing.T) {
	var err error
	cdrsPostFailCfgPath = path.Join(*utils.DataDir, "conf", "samples", cdrsPostFailConfDIR)
	if cdrsPostFailCfg, err = config.NewCGRConfigFromPath(cdrsPostFailCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	if err = os.RemoveAll(cdrsPostFailCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Error(err)
	}
	if err = os.MkdirAll(cdrsPostFailCfg.GeneralCfg().FailedPostsDir, 0755); err != nil {
		t.Error(err)
	}
}

func testCDRsPostFailoverInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cdrsPostFailCfg); err != nil {
		t.Fatal(err)
	}
}

// InitDb so we can rely on count
func testCDRsPostFailoverInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cdrsPostFailCfg); err != nil {
		t.Fatal(err)
	}
}

func testCDRsPostFailoverStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cdrsPostFailCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testCDRsPostFailoverRpcConn(t *testing.T) {
	cdrsPostFailRpc = engine.NewRPCClient(t, cdrsPostFailCfg.ListenCfg())
}

func testCDRsPostFailoverLoadTariffPlanFromFolder(t *testing.T) {
	var loadInst utils.LoadInstance
	if err := cdrsPostFailRpc.Call(context.Background(), utils.APIerSv2LoadTariffPlanFromFolder,
		&utils.AttrLoadTpFromFolder{FolderPath: path.Join(
			*utils.DataDir, "tariffplans", "testit")}, &loadInst); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
	var resp string
	if err := cdrsPostFailRpc.Call(context.Background(), utils.APIerSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var reply *engine.ChargerProfile
	if err := cdrsPostFailRpc.Call(context.Background(), utils.APIerSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"},
		&reply); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCDRsPostFailoverProcessCDR(t *testing.T) {
	args := &engine.ArgV1ProcessEvent{
		Flags: []string{utils.MetaExport, "*attributes:false", "*rals:false", "*chargers:false",
			"*store:false", "*thresholds:false", "*stats:false"}, // only export the CDR
		CGREvent: utils.CGREvent{
			ID:     "1",
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.OriginID:     "testCDRsPostFailoverProcessCDR",
				utils.OriginHost:   "192.168.1.1",
				utils.Source:       "testCDRsPostFailoverProcessCDR",
				utils.RequestType:  utils.MetaRated,
				utils.Category:     "call",
				utils.AccountField: "testCDRsPostFailoverProcessCDR",
				utils.Subject:      "ANY2CNT",
				utils.Destination:  "+4986517174963",
				utils.AnswerTime:   time.Date(2018, 8, 24, 16, 00, 26, 0, time.UTC),
				utils.Usage:        time.Minute,
				"field_extr1":      "val_extr1",
				"fieldextr2":       "valextr2",
			},
		},
	}

	var reply string
	if err := cdrsPostFailRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	args.ID = "2"
	args.Event[utils.OriginID] = "2"
	if err := cdrsPostFailRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
	args.ID = "3"
	args.Event[utils.OriginID] = "3"
	if err := cdrsPostFailRpc.Call(context.Background(), utils.CDRsV1ProcessEvent, args, &reply); err != nil {
		t.Error("Unexpected error: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Unexpected reply received: ", reply)
	}
}

func testCDRsPostFailoverToFile(t *testing.T) {
	time.Sleep(2 * time.Second)
	filesInDir, _ := os.ReadDir(cdrsPostFailCfg.GeneralCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsPostFailCfg.GeneralCfg().FailedPostsDir)
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsPostFailCfg.GeneralCfg().FailedPostsDir, fileName)

		ev, err := ees.NewExportEventsFromFile(filePath)
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		if ev.Type != utils.MetaS3jsonMap {
			t.Errorf("Expected event to use %q received: %q", utils.MetaS3jsonMap, ev.Type)
		}
		if len(ev.Events) != 3 {
			t.Errorf("Expected all the events to be saved in the same file, ony %v saved in this file.", len(ev.Events))
		}
	}
}

func testCDRsPostFailoverKillEngine(t *testing.T) {
	if err := os.RemoveAll(cdrsPostFailCfg.GeneralCfg().FailedPostsDir); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
