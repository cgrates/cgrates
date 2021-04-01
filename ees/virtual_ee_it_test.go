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

package ees

import (
	"net/rpc"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	virtConfigDir string
	virtCfgPath   string
	virtCfg       *config.CGRConfig
	virtRpc       *rpc.Client

	sTestsVirt = []func(t *testing.T){
		testCreateDirectory,
		testVirtLoadConfig,
		testVirtResetDataDB,
		testVirtResetStorDb,
		testVirtStartEngine,
		testVirtRPCConn,
		testVirtExportSupplierEvent,
		testVirtExportEvents,
		testVirtVerifyExports,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestVirtualExport(t *testing.T) {
	virtConfigDir = "ees"
	for _, stest := range sTestsVirt {
		t.Run(virtConfigDir, stest)
	}
}

func testVirtLoadConfig(t *testing.T) {
	var err error
	virtCfgPath = path.Join(*dataDir, "conf", "samples", virtConfigDir)
	if virtCfg, err = config.NewCGRConfigFromPath(virtCfgPath); err != nil {
		t.Error(err)
	}
}

func testVirtResetDataDB(t *testing.T) {
	if err := engine.InitDataDB(virtCfg); err != nil {
		t.Fatal(err)
	}
}

func testVirtResetStorDb(t *testing.T) {
	if err := engine.InitStorDB(virtCfg); err != nil {
		t.Fatal(err)
	}
}

func testVirtStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(virtCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testVirtRPCConn(t *testing.T) {
	var err error
	virtRpc, err = newRPCClient(virtCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testVirtExportSupplierEvent(t *testing.T) {
	supplierEvent := &utils.CGREventWithEeIDs{
		EeIDs: []string{"RouteExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "supplierEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        "SupplierRun",
				utils.Cost:         1.23,
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := virtRpc.Call(utils.EeSv1ProcessEvent, supplierEvent, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func testVirtExportEvents(t *testing.T) {
	eventVoice := &utils.CGREventWithEeIDs{
		EeIDs: []string{"CSVExporterFromVirt"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "voiceEvent",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Unix(1383813745, 0).UTC(),
				utils.AnswerTime:   time.Unix(1383813746, 0).UTC(),
				utils.Usage:        10 * time.Second,
				utils.RunID:        "SupplierRun",
				utils.Cost:         1.01,
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := virtRpc.Call(utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testVirtVerifyExports(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testCSVfromVirt/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.CSVSuffix) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if len(files) != 1 {
		t.Errorf("Expected %+v, received: %+v", 1, len(files))
	}
	eCnt := "dbafe9c8614c785a65aabd116dd3959c3c56f7f6,SupplierRun,dsafdsaf,cgrates.org,1001,1.01,CustomValue,1.23,SupplierRun\n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if eCnt != string(outContent1) {
		t.Errorf("Expecting: \n<%q>, \nreceived: \n<%q>", eCnt, string(outContent1))
	}
}
