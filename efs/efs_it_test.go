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
	"bytes"
	"encoding/gob"
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
		testEfsResetDBs,
		testEfsStartEngine,
		testEfSRPCConn,
		//testEfsProcessEvent,
		testEfsSKillEngine,
	}
)

func TestDecodeExportEvents(t *testing.T) {
	dirPath := "/var/spool/cgrates/failed_posts"
	filesInDir, err := os.ReadDir(dirPath)
	if err != nil {
		t.Error(err)
	}
	for _, file := range filesInDir {
		content, err := os.ReadFile(path.Join(dirPath, file.Name()))
		if err != nil {
			t.Error(err)
		}
		dec := gob.NewDecoder(bytes.NewBuffer(content))
		gob.Register(new(utils.CGREvent))
		singleEvent := new(FailedExportersLogg)
		if err := dec.Decode(&singleEvent); err != nil {
			t.Error(err)
		} else {
			strContent, err := utils.ToUnescapedJSON(singleEvent)
			if err != nil {
				t.Error(err)
			}
			fmt.Printf("singleEvent: %v \n", string(strContent))
		}
	}

}

func TestEfS(t *testing.T) {
	switch *utils.DBType {
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

func testEfsResetDBs(t *testing.T) {
	if err := engine.InitDataDB(efsCfg); err != nil {
		t.Fatal(err)
	}
	if err := engine.InitStorDB(efsCfg); err != nil {
		t.Fatal(err)
	}
}

// Start CGR Engine
func testEfsStartEngine(t *testing.T) {
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

func testEfsProcessEvent(t *testing.T) {
	args := &utils.ArgsFailedPosts{
		Tenant: "cgrates.org",
		Path:   "localhost:9092",
		Event: &utils.CGREvent{
			Tenant: "cgrates.org",
			Event: map[string]any{
				utils.AccountField: "1002",
				utils.Destination:  "1003",
			},
		},
		FailedDir: "/var/spool/cgrates/failed_posts",
		Module:    utils.Kafka,
		APIOpts: map[string]any{
			utils.Level:          efsCfg.LoggerCfg().Level,
			utils.Format:         "TutorialTopic",
			utils.Conn:           "localhost:9092",
			utils.FailedPostsDir: "/var/spool/cgrates/failed_posts",
			utils.Attempts:       efsCfg.LoggerCfg().Opts.KafkaAttempts,
		},
	}
	var reply string
	if err := efsRpc.Call(context.Background(), utils.EfSv1ProcessEvent,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

// Kill the engine when it is about to be finished
func testEfsSKillEngine(t *testing.T) {
	time.Sleep(7 * time.Second)
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
