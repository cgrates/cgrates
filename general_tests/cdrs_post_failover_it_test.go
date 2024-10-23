//go:build flaky

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
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
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
		testCDRsPostFailoverFlushDBs,
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
	if cdrsPostFailCfg, err = config.NewCGRConfigFromPath(context.Background(), cdrsPostFailCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
	if err = os.RemoveAll(cdrsPostFailCfg.EFsCfg().FailedPostsDir); err != nil {
		t.Error(err)
	}
	if err = os.MkdirAll(cdrsPostFailCfg.EFsCfg().FailedPostsDir, 0755); err != nil {
		t.Error(err)
	}
}

func testCDRsPostFailoverFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(cdrsPostFailCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(cdrsPostFailCfg); err != nil {
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
	cdrsPostFailRpc = engine.NewRPCClient(t, cdrsPostFailCfg.ListenCfg(), *utils.Encoding)
}

func testCDRsPostFailoverLoadTariffPlanFromFolder(t *testing.T) {
	caching := utils.MetaReload
	if cdrsPostFailCfg.DataDbCfg().Type == utils.MetaInternal {
		caching = utils.MetaNone
	}
	var reply string
	if err := cdrsPostFailRpc.Call(context.Background(), utils.LoaderSv1Run,
		&loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaStopOnError: true,
				utils.MetaCache:       caching,
			},
		}, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error("Unexpected reply returned:", reply)
	}

	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond) // Give time for scheduler to execute topups
	var resp string
	if err := cdrsPostFailRpc.Call(context.Background(), utils.AdminSv1RemoveChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"}, &resp); err != nil {
		t.Error(err)
	} else if resp != utils.OK {
		t.Error("Unexpected reply returned", resp)
	}
	var replyChargerPrf *engine.ChargerProfile
	if err := cdrsPostFailRpc.Call(context.Background(), utils.AdminSv1GetChargerProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "SupplierCharges"},
		&replyChargerPrf); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func testCDRsPostFailoverProcessCDR(t *testing.T) {
	args := &utils.CGREvent{
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
		APIOpts: map[string]any{
			utils.OptsCDRsExport: true,
			utils.MetaAttributes: false,
			utils.MetaChargers:   false,
			utils.OptsCDRsStore:  false,
			utils.MetaThresholds: false,
			utils.MetaStats:      false,
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
	time.Sleep(3 * time.Second)
	filesInDir, _ := os.ReadDir(cdrsPostFailCfg.EFsCfg().FailedPostsDir)
	if len(filesInDir) == 0 {
		t.Fatalf("No files in directory: %s", cdrsPostFailCfg.EFsCfg().FailedPostsDir)
	}
	for _, file := range filesInDir { // First file in directory is the one we need, harder to find it's name out of config
		fileName := file.Name()
		filePath := path.Join(cdrsPostFailCfg.EFsCfg().FailedPostsDir, fileName)

		ev, err := efs.NewFailoverPosterFromFile(filePath, utils.EEs, &efs.EfS{})
		if err != nil {
			t.Errorf("<%s> for file <%s>", err, fileName)
			continue
		} else if len(ev.(*efs.FailedExportersEEs).Events) == 0 {
			t.Error("Expected at least one event")
			continue
		}
		if len(ev.(*efs.FailedExportersEEs).Events) != 3 {
			t.Errorf("Expected all the events to be saved in the same file, ony %v saved in this file.", len(ev.(*efs.FailedExportersEEs).Events))
		}
	}

}

func testCDRsPostFailoverKillEngine(t *testing.T) {
	if err := os.RemoveAll(cdrsPostFailCfg.EFsCfg().FailedPostsDir); err != nil {
		t.Error(err)
	}
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
