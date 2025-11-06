//go:build flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://.gnu.org/licenses/>
*/

package efs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/nats-io/nats.go"
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
		singleEvent := new(FailedExportersLog)
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
	case utils.MetaRedis:
		t.SkipNow()
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
	if err := engine.InitDB(efsCfg); err != nil {
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

// TestEFsReplayEvents tests the implementation of the EfSv1.ReplayEvents.
func TestEFsReplayEvents(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}
	failedDir := t.TempDir()
	content := fmt.Sprintf(`{
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal",
				"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
    	}
	},
},

"admins": {
	"enabled": true
},
"efs": {
	"enabled": true,
	"failed_posts_ttl": "3ms",
	"poster_attempts": 1
},
"ees": {
	"enabled": true,
	"exporters": [
		{
			"id": "nats_exporter",
			"type": "*natsJSONMap",
			"flags": ["*log"],
			"efs_conns": ["*localhost"],
			"export_path": "nats://localhost:4222",
			"attempts": 1,
			"failed_posts_dir": "%s",
			"synchronous": true,
			"opts": {
				"natsSubject": "processed_cdrs",
			},
			"fields":[
				{"tag": "TestField", "path": "*exp.TestField", "type": "*variable", "value": "~*req.TestField"},
			]
		}
	]
}
}`, failedDir)

	testEnv := engine.TestEngine{
		ConfigJSON: content,
		// LogBuffer:  &bytes.Buffer{},
		Encoding: *utils.Encoding,
	}
	// defer fmt.Println(testEnv.LogBuffer)
	client, _ := testEnv.Run(t)
	// helper to sort slices
	less := func(a, b string) bool { return a < b }
	// amount of events to export/replay
	count := 5
	t.Run("successful nats export", func(t *testing.T) {
		cmd := exec.Command("nats-server")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start nats-server: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		defer cmd.Process.Kill()
		nc, err := nats.Connect("nats://localhost:4222", nats.Timeout(time.Second), nats.DrainTimeout(time.Second))
		if err != nil {
			t.Fatalf("failed to connect to nats-server: %v", err)
		}
		defer nc.Drain()
		ch := make(chan *nats.Msg, count)
		sub, err := nc.ChanQueueSubscribe("processed_cdrs", "", ch)
		if err != nil {
			t.Fatalf("failed to subscribe to nats queue: %v", err)
		}
		var reply map[string]map[string]any
		for i := range count {
			if err := client.Call(context.Background(), utils.EeSv1ProcessEvent, &utils.CGREventWithEeIDs{
				CGREvent: &utils.CGREvent{
					Tenant: "cgrates.org",
					ID:     "test",
					Event: map[string]any{
						"TestField": i,
					},
				},
			}, &reply); err != nil {
				t.Errorf("EeSv1.ProcessEvent returned unexpected err: %v", err)
			}
		}
		time.Sleep(1 * time.Millisecond) // wait for the channel to receive the replayed exports
		want := make([]string, 0, count)
		for i := range count {
			want = append(want, fmt.Sprintf(`{"TestField":"%d"}`, i))
		}
		if err := sub.Unsubscribe(); err != nil {
			t.Errorf("failed to unsubscribe from nats subject: %v", err)
		}
		close(ch)
		got := make([]string, 0, count)
		for elem := range ch {
			got = append(got, string(elem.Data))
		}
		if diff := cmp.Diff(want, got, cmpopts.SortSlices(less)); diff != "" {
			t.Errorf("unexpected nats messages received over channel (-want +got): \n%s", diff)
		}
	})
	t.Run("replay failed nats export", func(t *testing.T) {
		t.Skip("skipping due to gob decoding err")
		var exportReply map[string]map[string]any
		for i := range count {
			err := client.Call(context.Background(), utils.EeSv1ProcessEvent,
				&utils.CGREventWithEeIDs{
					CGREvent: &utils.CGREvent{
						Tenant: "cgrates.org",
						ID:     "test",
						Event: map[string]any{
							"TestField": i,
						},
					},
				}, &exportReply)
			if err == nil || err.Error() != utils.ErrPartiallyExecuted.Error() {
				t.Errorf("EeSv1.ProcessEvent err = %v, want %v", err, utils.ErrPartiallyExecuted)
			}
		}
		time.Sleep(5 * time.Millisecond)
		replayFailedDir := t.TempDir()
		var replayReply string
		if err := client.Call(context.Background(), utils.EfSv1ReplayEvents, ReplayEventsParams{
			Tenant:     "cgrates.org",
			Provider:   utils.EEs,
			SourcePath: failedDir,
			FailedPath: replayFailedDir,
			Modules:    []string{"test", "EEs"},
		}, &replayReply); err != nil {
			t.Errorf("EfSv1.ReplayEvents returned unexpected err: %v", err)
		}
		cmd := exec.Command("nats-server")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start nats-server: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
		defer cmd.Process.Kill()
		nc, err := nats.Connect("nats://localhost:4222", nats.Timeout(time.Second), nats.DrainTimeout(time.Second))
		if err != nil {
			t.Fatalf("failed to connect to nats-server: %v", err)
		}
		defer nc.Drain()
		ch := make(chan *nats.Msg, count)
		sub, err := nc.ChanQueueSubscribe("processed_cdrs", "", ch)
		if err != nil {
			t.Fatalf("failed to subscribe to nats queue: %v", err)
		}
		if err := client.Call(context.Background(), utils.EfSv1ReplayEvents, ReplayEventsParams{
			Tenant:     "cgrates.org",
			Provider:   utils.EEs,
			SourcePath: replayFailedDir,
			FailedPath: utils.MetaNone,
			Modules:    []string{"test", "EEs"},
		}, &replayReply); err != nil {
			t.Errorf("EfSv1.ReplayEvents returned unexpected err: %v", err)
		}
		time.Sleep(time.Millisecond) // wait for the channel to receive the replayed exports
		want := make([]string, 0, count)
		for i := range count {
			want = append(want, fmt.Sprintf(`{"TestField":"%d"}`, i))
		}
		if err := sub.Unsubscribe(); err != nil {
			t.Errorf("failed to unsubscribe from nats subject: %v", err)
		}
		close(ch)
		got := make([]string, 0, count)
		for elem := range ch {
			got = append(got, string(elem.Data))
		}
		if diff := cmp.Diff(want, got, cmpopts.SortSlices(less)); diff != "" {
			t.Errorf("unexpected nats messages received over channel (-want +got): \n%s", diff)
		}
	})
}
