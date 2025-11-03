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
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package apis

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
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
	"node_id": "apis_cores_test"
},
"db": {
	"db_conns": {
		"*default": {
			"db_type": "*internal"
    	}
	},
	"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
	}
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
		"-cpuprof_dir", profDir,
		"-memprof_dir", profDir, "-memprof_interval", "50ms", "-memprof_maxfiles", "2", "-memprof_timestamp")

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
		time.Sleep(10 * time.Millisecond) // wait for the final mem prof file to be written
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
	"db_conns": {
		"*default": {
			"db_type": "*internal"
    	}
	},
	"opts":{
		"internalDBRewriteInterval": "0s",
		"internalDBDumpInterval": "0s"
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
		time.Sleep(10 * time.Millisecond) // wait for the final mem prof file to be written
	})
}

func checkMemProfiles(t *testing.T, memDirPath string, wantCount int) {
	t.Helper()
	hasFinal := false
	memFileCount := 0
	_ = filepath.WalkDir(memDirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Logf("failed to access path %s: %v", path, err)
			return nil // skip paths that cause an error
		}
		defer func() {
		}()
		switch {
		case d.IsDir():
			// Memory profiles should be directly under 'memDirPath', skip all directories (excluding 'memDirPath')
			// and their contents.
			if path == memDirPath {
				return nil
			}
			return filepath.SkipDir
		case !strings.HasPrefix(d.Name(), "mem_") || !strings.HasSuffix(d.Name(), ".prof"):
			return nil // skip files that don't have 'mem_*.prof' format
		case d.Name() == utils.MemProfFinalFile:
			hasFinal = true
			fallthrough // test should be the same as for a normal mem file
		default: // files with format 'mem_*.prof'
			fi, err := d.Info()
			if err != nil {
				t.Errorf("failed to retrieve FileInfo from %q: %v", path, err)
			}
			if fi.Size() == 0 {
				t.Errorf("memory profile file %q is empty", path)
			}
			if d.Name() != utils.MemProfFinalFile {
				// Check that date within file name is from within this minute.
				layout := "20060102150405"
				timestamp := strings.TrimPrefix(d.Name(), "mem_")
				timestamp = strings.TrimSuffix(timestamp, ".prof")
				date, extra, has := strings.Cut(timestamp, "_")
				if !has {
					t.Errorf("expected timestamp to have '<date>_<milliseconds>' format, got: %s", timestamp)
				}
				parsedTime, err := time.ParseInLocation(layout, date, time.Local)
				if err != nil {
					t.Errorf("time.Parse(%q,%q) returned unexpected err: %v", layout, date, err)
				}

				// Convert 'extra' to microseconds and add to the parsed time.
				microSCount, err := strconv.Atoi(extra)
				if err != nil {
					t.Errorf("strconv.Atoi(%q) returned unexpected err: %v", extra, err)
				}
				parsedTime.Add(time.Duration(microSCount) * time.Microsecond)

				now := time.Now()
				oneMinuteEarlier := now.Add(-time.Minute)
				if parsedTime.Before(oneMinuteEarlier) || parsedTime.After(now) {
					t.Errorf("file name (%s) timestamp not from within last minute", d.Name())
				}
			}
			memFileCount++
		}
		return nil
	})

	if wantCount != 0 && !hasFinal {
		t.Error("final mem file is missing")
	}
	if memFileCount != wantCount {
		t.Errorf("memory file count = %d, want %d (including final mem profile)", memFileCount, wantCount)
	}
}
