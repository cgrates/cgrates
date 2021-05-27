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
	"os/exec"
	"path"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	fltrUpdateCfgPath1, fltrUpdateCfgPath2 string
	fltrUpdateCfgDIR1, fltrUpdateCfgDIR2   string
	fltrUpdateCfg1, fltrUpdateCfg2         *config.CGRConfig
	fltrUpdateRPC1, fltrUpdateRPC2         *rpc.Client
	testEng1                               *exec.Cmd
	sTestsFilterUpdate                     = []func(t *testing.T){
		testFilterUpdateInitCfg,
		testFilterUpdateResetDB,
		testFilterUpdateStartEngine,
		testFilterUpdateRpcConn,
		testFilterUpdateSetFilterE1,
		testFilterUpdateSetAttrProfileE1,
		testFilterUpdateGetAttrProfileForEventFirstEvE1,
		testFilterUpdateGetAttrProfileForEventFirstEvE2,
		testFilterUpdateGetAttrProfileForEventSecondEvE1,
		testFilterUpdateGetAttrProfileForEventSecondEvE2,
		testFilterUpdateSetFilterAfterAttrE1,
		testFilterUpdateGetAttrProfileForEventFirstEvE1,
		testFilterUpdateGetAttrProfileForEventFirstEvE2,
		testFilterUpdateGetAttrProfileForEventSecondEvE1,
		testFilterUpdateGetAttrProfileForEventSecondEvE2,
		testFilterUpdateStopEngine,
	}
)

func TestFilterUpdateIT(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		t.SkipNow()
	case utils.MetaMySQL:
		fltrUpdateCfgDIR1 = "fltr_update_e1_mysql"
		fltrUpdateCfgDIR2 = "tutmysql"
	case utils.MetaMongo:
		fltrUpdateCfgDIR1 = "fltr_update_e1_mongo"
		fltrUpdateCfgDIR2 = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}

	for _, stest1 := range sTestsFilterUpdate {
		t.Run(*dbType, stest1)
	}
}

//Init Config
func testFilterUpdateInitCfg(t *testing.T) {
	fltrUpdateCfgPath1 = path.Join(*dataDir, "conf", "samples", "cache_replicate", fltrUpdateCfgDIR1)
	if fltrUpdateCfg1, err = config.NewCGRConfigFromPath(fltrUpdateCfgPath1); err != nil {
		t.Fatal(err)
	}
	fltrUpdateCfgPath2 = path.Join(*dataDir, "conf", "samples", fltrUpdateCfgDIR2)
	if fltrUpdateCfg2, err = config.NewCGRConfigFromPath(fltrUpdateCfgPath2); err != nil {
		t.Fatal(err)
	}
}

// Remove data in both rating and accounting db
func testFilterUpdateResetDB(t *testing.T) {
	if err := engine.InitDataDb(fltrUpdateCfg1); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDb(fltrUpdateCfg1); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testFilterUpdateStartEngine(t *testing.T) {
	if _, err = engine.StopStartEngine(fltrUpdateCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	if testEng1, err = engine.StartEngine(fltrUpdateCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}

}

// Connect rpc client to rater
func testFilterUpdateRpcConn(t *testing.T) {
	if fltrUpdateRPC1, err = newRPCClient(fltrUpdateCfg1.ListenCfg()); err != nil {
		t.Fatal(err)
	}
	if fltrUpdateRPC2, err = newRPCClient(fltrUpdateCfg2.ListenCfg()); err != nil {
		t.Fatal(err)
	}
}

func testFilterUpdateStopEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}

func testFilterUpdateSetFilterE1(t *testing.T) {

}

func testFilterUpdateSetAttrProfileE1(t *testing.T) {

}

func testFilterUpdateGetAttrProfileForEventFirstEvE1(t *testing.T) {

}

func testFilterUpdateGetAttrProfileForEventFirstEvE2(t *testing.T) {

}

func testFilterUpdateGetAttrProfileForEventSecondEvE1(t *testing.T) {

}

func testFilterUpdateGetAttrProfileForEventSecondEvE2(t *testing.T) {

}

func testFilterUpdateSetFilterAfterAttrE1(t *testing.T) {

}
