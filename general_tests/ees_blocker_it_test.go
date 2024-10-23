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

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	eeSBlockerFiles   = []string{"/tmp/CSVFile1", "/tmp/CSVFile2"}
	eesBlockerCfgPath string
	eesBlockerCfg     *config.CGRConfig
	eesBlockerRPC     *birpc.Client
	eesBlockerConfDIR string //run tests for specific configuration

	eesBlockerTests = []func(t *testing.T){
		testEEsBlockerCreateFiles,
		testEEsBlockerLoadConfig,
		testEEsBlockerFlushDBs,
		testEEsBlockerStartEngine,
		testEEsBlockerRpcConn,
		testEEsBlockerExportEvent,
		testEEsBlockerVerifyExport,
		testEEsBlockerStopEngine,
		testEEsBlockerDeleteFiles,
	}
)

// Test start here
func TestEEsBlocker(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		eesBlockerConfDIR = "ees_blocker_internal"
	case utils.MetaMySQL:
		eesBlockerConfDIR = "ees_blocker_mysql"
	case utils.MetaMongo:
		eesBlockerConfDIR = "ees_blocker_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range eesBlockerTests {
		t.Run(eesBlockerConfDIR, stest)
	}
}

func testEEsBlockerCreateFiles(t *testing.T) {
	for _, dir := range eeSBlockerFiles {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testEEsBlockerDeleteFiles(t *testing.T) {
	for _, dir := range eeSBlockerFiles {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
	}
}

func testEEsBlockerLoadConfig(t *testing.T) {
	var err error
	eesBlockerCfgPath = path.Join(*utils.DataDir, "conf", "samples", eesBlockerConfDIR)
	if eesBlockerCfg, err = config.NewCGRConfigFromPath(context.Background(), eesBlockerCfgPath); err != nil {
		t.Error(err)
	}
}

func testEEsBlockerFlushDBs(t *testing.T) {
	if err := engine.InitDataDB(eesBlockerCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(eesBlockerCfg); err != nil {
		t.Fatal(err)
	}
}

func testEEsBlockerStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(eesBlockerCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testEEsBlockerRpcConn(t *testing.T) {
	eesBlockerRPC = engine.NewRPCClient(t, eesBlockerCfg.ListenCfg(), *utils.Encoding)
}

func testEEsBlockerExportEvent(t *testing.T) {
	exportedEvent := &utils.CGREventWithEeIDs{
		CGREvent: &utils.CGREvent{
			Tenant: "cgrates.org",
			ID:     "EEsProcessEvent",
			Event: map[string]any{
				"TestCase": "EEsBlockerBehaviour",
			},
			APIOpts: map[string]any{},
		},
	}

	var reply map[string]map[string]any
	if err := eesBlockerRPC.Call(context.Background(), utils.EeSv1ProcessEvent, exportedEvent, &reply); err != nil {
		t.Error(err)
	}
}

func testEEsBlockerVerifyExport(t *testing.T) {
	for i, dir := range eeSBlockerFiles {
		if files, err := os.ReadDir(dir); err != nil {
			t.Fatal(err)
		} else if i == 0 && len(files) != 1 {
			t.Errorf("expected to find only 1 file, received <%d>", len(files))
		} else if i == 1 && len(files) != 0 {
			t.Errorf("expected to find 0 files, received <%d>", len(files))
		}
	}
}

func testEEsBlockerStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
