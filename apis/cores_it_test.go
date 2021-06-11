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

package apis

import (
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	coreItCfgPath string
	coreItDirPath string
	coreItCfg     *config.CGRConfig
	coreItBiRPC   *birpc.Client
	coreItTests   = []func(t *testing.T){
		testCoreItLoadConfig,
		testCoreItInitDataDb,
		testCoreItInitStorDb,
		testCoreItStartEngine,
		testCoreItRpcConn,
		testCoreItStatus,
		testCoreItKillEngine,
	}
)

func TestCoreItTests(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		coreItDirPath = "all2"
	case utils.MetaMongo:
		coreItDirPath = "all2_mongo"
	case utils.MetaMySQL:
		coreItDirPath = "all2_mysql"
	default:
		t.Fatalf("Unsupported database")
	}
	for _, test := range coreItTests {
		t.Run("Running integration tests", test)
	}
}

func testCoreItLoadConfig(t *testing.T) {
	var err error
	coreItCfgPath = path.Join(*dataDir, "conf", "samples", "dispatchers", coreItDirPath)
	if coreItCfg, err = config.NewCGRConfigFromPath(coreItCfgPath); err != nil {
		t.Fatal(err)
	}
}

func testCoreItInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(coreItCfg); err != nil {
		t.Fatal(err)
	}
}

func testCoreItInitStorDb(t *testing.T) {
	if err := engine.InitStorDB(coreItCfg); err != nil {
		t.Fatal(err)
	}
}

func testCoreItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(coreItCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testCoreItRpcConn(t *testing.T) {
	var err error
	if coreItBiRPC, err = newRPCClient(coreItCfg.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testCoreItStatus(t *testing.T) {
	args := &utils.TenantIDWithAPIOpts{}
	var reply map[string]interface{}
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1Status,
		args, &reply); err != nil {
		t.Fatal(err)
	} else if reply[utils.NodeID] != "ALL2" {
		t.Errorf("Expected ALL2 but received %v", reply[utils.NodeID])
	}
}

func testCoreItKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
