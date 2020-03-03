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
	"net/rpc"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	expCfgDir  string
	expCfgPath string
	expCfg     *config.CGRConfig
	expRpc     *rpc.Client

	sTestsExp = []func(t *testing.T){
		testExpLoadConfig,
		testExpResetDataDB,
		testExpResetStorDb,
		testExpStartEngine,
		testExpRPCConn,
		testExpLoadTPFromFolder,
		testExpAttribute,
		testExpStopCgrEngine,
	}
)

func TestExport(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		expCfgDir = "tutinternal"
	case utils.MetaMySQL:
		expCfgDir = "tutmysql"
	case utils.MetaMongo:
		expCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest := range sTestsExp {
		t.Run(expCfgDir, stest)
	}
}

func testExpLoadConfig(t *testing.T) {
	expCfgPath = path.Join(*dataDir, "conf", "samples", expCfgDir)
	if expCfg, err = config.NewCGRConfigFromPath(expCfgPath); err != nil {
		t.Error(err)
	}
}

func testExpResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(expCfg); err != nil {
		t.Fatal(err)
	}
}

func testExpResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(expCfg); err != nil {
		t.Fatal(err)
	}
}

func testExpStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(expCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testExpRPCConn(t *testing.T) {
	var err error
	expRpc, err = newRPCClient(expCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testExpLoadTPFromFolder(t *testing.T) {
	var reply string
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := expRpc.Call(utils.APIerSv1LoadTariffPlanFromFolder, attrs, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpAttribute(t *testing.T) {
	var reply string
	arg := &utils.ArgExportToFolder{
		Path:  "/tmp",
		Items: []string{utils.MetaAttributes},
	}
	if err := expRpc.Call(utils.APIerSv1ExportToFolder, arg, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Error(reply)
	}
}

func testExpStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
