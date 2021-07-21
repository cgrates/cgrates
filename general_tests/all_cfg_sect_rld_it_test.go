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
	testSectCfgDir  string
	testSectCfgPath string
	testSectCfg     *config.CGRConfig
	testSectRPC     *rpc.Client

	testSectTests = []func(t *testing.T){
		testSectLoadConfig,
		testSectResetDataDB,
		testSectResetStorDb,
		testSectStartEngine,
		testSectRPCConn,
		testSectStopCgrEngine,
	}
)

func TestSectChange(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		testSectCfgDir = "tutinternal"
	case utils.MetaMySQL:
		testSectCfgDir = "tutmysql"
	case utils.MetaMongo:
		testSectCfgDir = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, testSectest := range testSectTests {
		t.Run(testSectCfgDir, testSectest)
	}
}

func testSectLoadConfig(t *testing.T) {
	testSectCfgPath = path.Join(*dataDir, "conf", "samples", testSectCfgDir)
	if testSectCfg, err = config.NewCGRConfigFromPath(testSectCfgPath); err != nil {
		t.Error(err)
	}
}

func testSectResetDataDB(t *testing.T) {
	if err := engine.InitDataDb(testSectCfg); err != nil {
		t.Fatal(err)
	}
}

func testSectResetStorDb(t *testing.T) {
	if err := engine.InitStorDb(testSectCfg); err != nil {
		t.Fatal(err)
	}
}

func testSectStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(testSectCfgPath, *waitRater); err != nil {
		t.Fatal(err)
	}
}

func testSectRPCConn(t *testing.T) {
	var err error
	testSectRPC, err = newRPCClient(testSectCfg.ListenCfg())
	if err != nil {
		t.Fatal(err)
	}
}

func testSectStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
