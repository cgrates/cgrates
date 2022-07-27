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
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package efs

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	dbType       = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
	efsConfigDir string
	efsCfgPath   string
	efsCfg       *config.CGRConfig
	efsRpc       *birpc.Client
	exportPath   = []string{
		"/tmp/testKafkaLogger",
	}

	sTestsEfs = []func(t *testing.T){
		testCreateDirectory,
		testEfSInitCfg,
		testEfsInitDataDb,
		testEfsStartEngine,
		testEfSRPCConn,
		testEfsSKillEngine,
	}
)

func TestEfS(t *testing.T) {
	switch *dbType {
	case utils.MetaInternal:
		efsConfigDir = "efs_internal"
	case utils.MetaMongo:
		efsConfigDir = "efs_mongo"
	case utils.MetaMySQL:
		efsConfigDir = "efs_mysql"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsEfs {
		t.Run(efsConfigDir, stest)
	}
}

func testCreateDirectory(t *testing.T) {
	for _, dir := range exportPath {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal("Error removing folder: ", dir, err)
		}
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			t.Fatal("Error creating folder: ", dir, err)
		}
	}
}

func testEfSInitCfg(t *testing.T) {
	var err error
	efsCfgPath = path.Join("/usr/share/cgrates", "conf", "samples", "efs", efsConfigDir)
	efsCfg, err = config.NewCGRConfigFromPath(context.Background(), efsCfgPath)
	if err != nil {
		t.Error(err)
	}
}

func testEfsInitDataDb(t *testing.T) {
	if err := engine.InitDataDB(efsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testEfsStartEngine(t *testing.T) {
	fmt.Println(efsCfgPath)
	if _, err := engine.StopStartEngine(efsCfgPath, 100); err != nil {
		t.Fatal(err)
	}
}

func testEfSRPCConn(t *testing.T) {
	var err error
	efsRpc, err = jsonrpc.Dial(utils.TCP, efsCfg.ListenCfg().RPCJSONListen)
	if err != nil {

		t.Fatal(err)
	}
}

//Kill the engine when it is about to be finished
func testEfsSKillEngine(t *testing.T) {
	time.Sleep(7 * time.Second)
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
