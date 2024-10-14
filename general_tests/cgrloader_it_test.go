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
	"flag"
	"os/exec"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	loaderGoogleSheet = flag.Bool("google_sheet", false, "Run the test with google sheet")
	cgrloaderCfgPath  string
	cgrloaderCfg      *config.CGRConfig
	cgrloaderRPC      *birpc.Client
	cgrloaderConfDIR  string //run tests for specific configuration

	sTestsCGRLoaders = []func(t *testing.T){
		testCGRLoaderInitConfig,
		testCGRLoaderInitDataDb,
		testCGRLoaderInitCdrDb,
		testCGRLoaderStartEngine,
		testCGRLoaderRpcConn,
		testCGRLoaderLoadData,
		testCGRLoaderGetData,
		testCGRLoaderKillEngine,
	}
)

// Test start here
func TestCGRLoader(t *testing.T) {
	if !*loaderGoogleSheet {
		t.SkipNow()
		return
	}
	switch *utils.DBType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		cgrloaderConfDIR = "loader_mysql"
	case utils.MetaMongo:
		cgrloaderConfDIR = "loader_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsCGRLoaders {
		t.Run(cgrloaderConfDIR, stest)
	}
}

func testCGRLoaderInitConfig(t *testing.T) {
	var err error
	cgrloaderCfgPath = path.Join(*utils.DataDir, "conf", "samples", cgrloaderConfDIR)
	if cgrloaderCfg, err = config.NewCGRConfigFromPath(cgrloaderCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testCGRLoaderInitDataDb(t *testing.T) {
	if err := engine.InitDataDb(cgrloaderCfg); err != nil {
		t.Fatal(err)
	}
}

func testCGRLoaderInitCdrDb(t *testing.T) {
	if err := engine.InitStorDb(cgrloaderCfg); err != nil {
		t.Fatal(err)
	}
}

func testCGRLoaderStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cgrloaderCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCGRLoaderRpcConn(t *testing.T) {
	cgrloaderRPC = engine.NewRPCClient(t, cgrloaderCfg.ListenCfg())
}

func testCGRLoaderLoadData(t *testing.T) {
	loaderPath, err := exec.LookPath("cgr-loader")
	if err != nil {
		t.Fatal(err)
	}
	loader := exec.Command(loaderPath, "-config_path", cgrloaderCfgPath, "-path", "*gapi:1pRFnsFBlKeGnD3wysZ1CXxylZI7r_Zh5iZI99ViOyPM")
	if err := loader.Start(); err != nil {
		t.Fatal(err)
	}
	if err := loader.Wait(); err != nil {
		t.Error(err)
	}
}

func testCGRLoaderGetData(t *testing.T) {
	expected := []string{"ATTR_1001_SIMPLEAUTH", "ATTR_1002_SIMPLEAUTH", "ATTR_1003_SIMPLEAUTH",
		"ATTR_1001_SESSIONAUTH", "ATTR_1002_SESSIONAUTH", "ATTR_1003_SESSIONAUTH",
		"ATTR_ACC_ALIAS"}
	var result []string
	if err := cgrloaderRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfileIDs, &utils.PaginatorWithTenant{Tenant: "cgrates.org"}, &result); err != nil {
		t.Error(err)
	} else if len(expected) != len(result) {
		t.Errorf("Expecting : %+v, received: %+v", expected, result)
	}
}

func testCGRLoaderKillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
}
