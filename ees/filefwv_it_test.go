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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fwvConfigDir string
	fwvCfgPath   string
	fwvCfg       *config.CGRConfig
	fwvRpc       *rpc.Client

	sTestsFwv = []func(t *testing.T){
		testCreateDirectory,
		testFwvLoadConfig,
		testFwvResetDataDB,
		testFwvResetStorDb,
		testFwvStartEngine,
		testFwvRPCConn,
		testFwvExportEvent,
		testFwvVerifyExports,
		testStopCgrEngine,
		testCleanDirectory,
	}
)

func TestFwvExport(t *testing.T) {
	fwvConfigDir = "ees"
	for _, stest := range sTestsFwv {
		t.Run(fwvConfigDir, stest)
	}
}

func testFwvLoadConfig(t *testing.T) {
	var err error
	fwvCfgPath = path.Join(*dataDir, "conf", "samples", fwvConfigDir)
	if fwvCfg, err = config.NewCGRConfigFromPath(fwvCfgPath); err != nil {
		t.Error(err)
	}
}

func testFwvResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

func testFwvResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(fwvCfg); err != nil {
		t.Fatal(err)
	}
}

func testFwvStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(fwvCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testFwvRPCConn(t *testing.T) {
	var err error
	fwvRpc, err = newRPCClient(fwvCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testFwvExportEvent(t *testing.T) {
	event := &utils.CGREventWithEeIDs{
		EeIDs: []string{"FwvExporter"},
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "Event",
			Time:   utils.TimePointer(time.Now()),
			Event: map[string]interface{}{
				utils.OrderID:      1,
				utils.CGRID:        utils.Sha1("dsafdsaf", time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC).String()),
				utils.ToR:          utils.MetaVoice,
				utils.OriginID:     "dsafdsaf",
				utils.OriginHost:   "192.168.1.1",
				utils.RequestType:  utils.MetaRated,
				utils.Tenant:       "cgrates.org",
				utils.Category:     "call",
				utils.AccountField: "1001",
				utils.Subject:      "1001",
				utils.Destination:  "1002",
				utils.SetupTime:    time.Date(2013, 11, 7, 8, 42, 20, 0, time.UTC),
				utils.AnswerTime:   time.Date(2013, 11, 7, 8, 42, 26, 0, time.UTC),
				utils.Usage:        10 * time.Second,
				utils.RunID:        utils.MetaDefault,
				utils.Cost:         2.34567,
				"ExtraFields":      map[string]string{"field_extr1": "val_extr1", "fieldextr2": "valextr2"},
			},
		},
	}
	var reply map[string]utils.MapStorage
	if err := fwvRpc.Call(utils.EeSv1ProcessEvent, event, &reply); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
}

func testFwvVerifyExports(t *testing.T) {
	var files []string
	err := filepath.Walk("/tmp/testFWV/", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, utils.FWVSuffix) {
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
	eHdr := "10   VOI02062016520001                                                                                                         \n"
	eCnt := "201001        1001 cli            1002                    0211  071113084200100000      1op3dsafdsaf                        002.34567\n"
	eTrl := "90   VOI0000010000010s071113084200                                                                                             \n"
	if outContent1, err := os.ReadFile(files[0]); err != nil {
		t.Error(err)
	} else if len(eHdr+eTrl+eCnt) != len(outContent1) {
		t.Errorf("Expecting: <%+v>, received: <%+v>", len(eHdr+eTrl+eCnt), len(outContent1))
		t.Errorf("Expecting: \n <%+v> \n, received: \n<%+v>\n", eHdr+eTrl+eCnt, string(outContent1))
	}
}

func TestFileFwvInit(t *testing.T) {
	cgrCfg := config.NewDefaultCGRConfig()
	cgrCfg.EEsCfg().Exporters[0].ExportPath = "/tmp/TestInitFileFWV"
	if err := os.MkdirAll("/tmp/TestInitFileFWV", 0666); err != nil {
		t.Error(err)
	}
	dc, err := newEEMetrics(utils.FirstNonEmpty(
		"Local",
		utils.EmptyString,
	))
	if err != nil {
		t.Error(err)
	}
	fFwv := &FileFWVee{
		cgrCfg: cgrCfg,
		cfg:    cgrCfg.EEsCfg().Exporters[0],
		dc:     dc,
	}
	if err := fFwv.init(); err != nil {
		t.Error(err)
	}
	if err := os.RemoveAll("/tmp/TestInitFileCSV"); err != nil {
		t.Error(err)
	}
}
