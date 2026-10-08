//go:build flaky

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package apis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func TestCoreSProfilingFlags(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	cfgJSON := `{
"general": {
	"nodeID": "apis_cores_test"
},
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
			"internalDBRewriteInterval": "0s",
			"internalDBDumpInterval": "0s"
		}
    	}
	},

}
}`

	profDir := t.TempDir()

	// NOTE: This will be executed after the cgr-engine process is killed.
	t.Cleanup(func() {
		checkMemProfiles(t, profDir, 3) // max files=2 + final heap profile
	})

	ng := engine.TestEngine{
		ConfigJSON: cfgJSON,
		Encoding:   *utils.Encoding,
		DBCfg:      dbCfg,
	}
	client, _ := ng.Run(t,
		"-cpuProfDir", profDir,
		"-memProfDir", profDir, "-memProfInterval", "50ms", "-memProfMaxFiles", "2", "-memProfTimestamp")

	t.Run("err cpu prof already started", func(t *testing.T) {
		expectedErr := "start CPU profiling: already started"
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
			&utils.DirectoryArgs{
				DirPath: profDir,
			}, &reply); err == nil || err.Error() != expectedErr {
			t.Errorf("%s err=%v, want %v", utils.CoreSv1StartCPUProfiling, err, expectedErr)
		}
	})

	t.Run("err mem prof already started", func(t *testing.T) {
		expErr := "start memory profiling: already started"
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
			cores.MemoryProfilingParams{
				DirPath:      profDir,
				Interval:     100 * time.Millisecond,
				MaxFiles:     2,
				UseTimestamp: true,
			}, &reply); err == nil || err.Error() != expErr {
			t.Errorf("%s err=%v, want %v", utils.CoreSv1StartMemoryProfiling, err, expErr)
		}
	})

	// Test the sleep api here, instead of starting another engine. This
	// will allow the heap profiles to be generated as well.
	t.Run("sleep", func(t *testing.T) {
		args := &utils.DurationArgs{
			Duration: 150 * time.Millisecond,
		}
		before := time.Now()
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1Sleep,
			args, &reply); err != nil {
			t.Error(err)
		} else if reply != utils.OK {
			t.Errorf("%s - unexpected reply returned: %s", utils.CoreSv1Sleep, reply)
		}
		got := time.Since(before)
		want := args.Duration
		margin := 10 * time.Millisecond
		if diff := got - want; diff < 0 || diff > margin {
			t.Errorf("%s - slept for %s, wanted to sleep around %s (diff %v, margin %v)",
				utils.CoreSv1Sleep, got, want, diff, margin)
		}
	})

	t.Run("status", func(t *testing.T) {
		want := "apis_cores_test"
		var reply map[string]any
		if err := client.Call(context.Background(), utils.CoreSv1Status,
			&utils.TenantIDWithAPIOpts{}, &reply); err != nil {
			t.Fatal(err)
		} else if got := reply[utils.NodeID]; got != want {
			t.Errorf("%s nodeID=%v, want %v", utils.CoreSv1Status, got, want)
		}
	})

	t.Run("check cpu prof file", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
			new(utils.TenantIDWithAPIOpts), &reply); err != nil {
			t.Error(err)
		}
		cpuProfPath := filepath.Join(profDir, utils.CpuPathCgr)
		fi, err := os.Stat(cpuProfPath)
		if err != nil {
			t.Error(err)
		} else if size := fi.Size(); size < int64(300) {
			t.Errorf("Size of CPUProfile %v is lower that expected", size)
		}
	})

	// TODO: Expecting the same result even if not manually stopping the
	// memory profiling. Test it just to be sure.
	t.Run("stop mem profiling", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
			new(utils.TenantWithAPIOpts), &reply); err != nil {
			t.Errorf("%s unexpected err=%v", utils.CoreSv1StopMemoryProfiling, err)
		}
	})
}

func TestCoreSProfilingAPI(t *testing.T) {
	var dbCfg engine.DBCfg
	switch *utils.DBType {
	case utils.MetaInternal:
		dbCfg = engine.InternalDBCfg
	case utils.MetaMySQL, utils.MetaRedis, utils.MetaMongo, utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("unsupported dbtype value")
	}

	profDir := t.TempDir()

	// NOTE: This will be executed after the cgr-engine process is killed.
	t.Cleanup(func() {
		checkMemProfiles(t, profDir, 3) // max files=2 + final heap profile
	})

	ng := engine.TestEngine{
		ConfigJSON: `{
"db": {
	"dbConns": {
		"*default": {
			"dbType": "*internal",
			"opts":{
			"internalDBRewriteInterval": "0s",
			"internalDBDumpInterval": "0s"
		}
    	}
	}
},}`,
		Encoding: *utils.Encoding,
		DBCfg:    dbCfg,
	}
	client, _ := ng.Run(t)

	t.Run("err cpu prof stop before start", func(t *testing.T) {
		expectedErr := "stop CPU profiling: not started yet"
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
			new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
			t.Errorf("%s err=%v, want %v", utils.CoreSv1StopCPUProfiling, err, expectedErr)
		}
	})

	t.Run("start cpu profiling", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
			&utils.DirectoryArgs{
				DirPath: profDir,
			}, &reply); err != nil {
			t.Errorf("%s unexpected err=%v", utils.CoreSv1StartCPUProfiling, err)
		}
	})

	t.Run("check cpu prof file", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
			new(utils.TenantIDWithAPIOpts), &reply); err != nil {
			t.Error(err)
		}
		cpuProfPath := filepath.Join(profDir, utils.CpuPathCgr)
		fi, err := os.Stat(cpuProfPath)
		if err != nil {
			t.Error(err)
		} else if size := fi.Size(); size < int64(300) {
			t.Errorf("Size of CPUProfile %v is lower that expected", size)
		}
	})

	t.Run("err mem prof stop before start", func(t *testing.T) {
		var reply string
		expectedErr := "stop memory profiling: not started yet"
		if err := client.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
			new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
			t.Errorf("%s err=%v, want %v", utils.CoreSv1StopMemoryProfiling, err, expectedErr)
		}
	})

	t.Run("start mem profiling", func(t *testing.T) {
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
			cores.MemoryProfilingParams{
				DirPath:      profDir,
				Interval:     50 * time.Millisecond,
				MaxFiles:     2,
				UseTimestamp: true,
			}, &reply); err != nil {
			t.Errorf("%s unexpected err=%v", utils.CoreSv1StartMemoryProfiling, err)
		}
	})

	// TODO: Expecting the same result even if not manually stopping the
	// memory profiling. Test it just to be sure.
	t.Run("stop mem profiling", func(t *testing.T) {
		time.Sleep(200 * time.Millisecond) // wait for the heap profiles to be generated
		var reply string
		if err := client.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
			new(utils.TenantWithAPIOpts), &reply); err != nil {
			t.Errorf("%s unexpected err=%v", utils.CoreSv1StopMemoryProfiling, err)
		}
	})
}

func checkMemProfiles(t *testing.T, memDirPath string, wantCount int) {
	t.Helper()
	entries, err := os.ReadDir(memDirPath)
	if err != nil {
		t.Fatal(err)
	}
	hasFinal := false
	memFileCount := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "mem_") || !strings.HasSuffix(name, ".prof") {
			continue
		}
		memFileCount++
		path := filepath.Join(memDirPath, name)
		fi, err := entry.Info()
		if err != nil {
			t.Fatalf("failed to retrieve FileInfo from %q: %v", path, err)
		}
		if fi.Size() == 0 {
			t.Errorf("memory profile file %q is empty", path)
		}
		if name == utils.MemProfFinalFile {
			hasFinal = true
		}
	}

	if wantCount != 0 && !hasFinal {
		t.Error("final mem file is missing")
	}
	if memFileCount != wantCount {
		t.Errorf("memory file count = %d, want %d (including final mem profile)", memFileCount, wantCount)
	}
}
