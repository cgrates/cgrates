//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package ees

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	virtConfigDir string
	virtCfgPath   string
	virtCfg       *config.CGRConfig
	virtRpc       *birpc.Client

	sTestsVirt = []func(t *testing.T){
		testCreateDirectory,
		testVirtLoadConfig,
		testVirtResetDBs,
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
	virtCfgPath = path.Join(*utils.DataDir, "conf", "samples", virtConfigDir)
	if virtCfg, err = config.NewCGRConfigFromPath(context.Background(), virtCfgPath); err != nil {
		t.Error(err)
	}
}

func testVirtResetDBs(t *testing.T) {
	if err := engine.InitDB(virtCfg); err != nil {
		t.Fatal(err)
	}
}

func testVirtStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(virtCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testVirtRPCConn(t *testing.T) {
	virtRpc = engine.NewRPCClient(t, virtCfg.ListenCfg(), *utils.Encoding)
}

func testVirtExportSupplierEvent(t *testing.T) {
	supplierEvent := &utils.CGREventWithEeIDs{
		EeIDs: []string{"RouteExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "supplierEvent",
			Event: map[string]any{

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
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}

	var reply map[string]utils.MapStorage
	if err := virtRpc.Call(context.Background(), utils.EeSv1ProcessEvent, supplierEvent, &reply); err != nil {
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
			Event: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
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
			APIOpts: map[string]any{
				utils.MetaOriginID: utils.Sha1("dsafdsaf", time.Unix(1383813745, 0).UTC().String()),
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := virtRpc.Call(context.Background(), utils.EeSv1ProcessEvent, eventVoice, &reply); err != nil {
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
