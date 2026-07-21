//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
		testCGRLoaderFlushDBs,
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
	case utils.MetaRedis:
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
	if cgrloaderCfg, err = config.NewCGRConfigFromPath(context.Background(), cgrloaderCfgPath); err != nil {
		t.Fatal("Got config error: ", err.Error())
	}
}

func testCGRLoaderFlushDBs(t *testing.T) {
	if err := engine.InitDB(cgrloaderCfg); err != nil {
		t.Fatal(err)
	}
}

func testCGRLoaderStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(cgrloaderCfgPath, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

func testCGRLoaderRpcConn(t *testing.T) {
	cgrloaderRPC = engine.NewRPCClient(t, cgrloaderCfg.ListenCfg(), *utils.Encoding)
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
	if err := cgrloaderRPC.Call(context.Background(), utils.AdminSv1GetAttributeProfileIDs, &utils.ArgsItemIDs{Tenant: "cgrates.org"}, &result); err != nil {
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
