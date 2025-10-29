//go:build integration
// +build integration

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

package v1

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	coreV1CfgPath string
	coreV1Cfg     *config.CGRConfig
	coreV1Rpc     *birpc.Client
	coreV1ConfDIR string //run tests for specific configuration
	argPath       string
	memProfNr     int
	sTestCoreSv1  = []func(t *testing.T){
		testCoreSv1LoadCofig,
		testCoreSv1InitDataDB,
		testCoreSv1InitStorDB,

		//engine separate with cpu
		testCoreSv1StartEngineByExecWithCPUProfiling,
		testCoreSv1RPCConn,
		testCoreSv1StartCPUProfilingErrorAlreadyStarted,
		testCoreSv1Sleep,
		testCoreSv1StopCPUProfiling,
		testCoreSv1KillEngine,

		//engine separate with memory
		testCoreSv1StartEngineByExecWIthMemProfiling,
		testCoreSv1RPCConn,
		testCoreSv1StartMemProfilingErrorAlreadyStarted,
		testCoreSv1Sleep,
		testCoreSv1StopMemoryProfiling,
		testCoreSv1KillEngine,

		// test CPU and Memory just by APIs
		testCoreSv1StartEngine,
		testCoreSv1RPCConn,

		//CPUProfiles apis
		testCoreSv1StopCPUProfilingBeforeStart,
		testCoreSv1StartCPUProfiling,
		testCoreSv1Sleep,
		testCoreSv1StopCPUProfiling,

		//MemoryProfiles apis
		testCoreSv1StopMemProfilingBeforeStart,
		testCoreSv1StartMemoryProfiling,
		testCoreSv1Sleep,
		testCoreSv1StopMemoryProfiling,
		testCoreSv1KillEngine,
	}
)

func TestITCoreSv1(t *testing.T) {
	argPath = t.TempDir()
	switch *utils.DBType {
	case utils.MetaInternal:
		coreV1ConfDIR = "tutinternal"
	case utils.MetaMySQL:
		coreV1ConfDIR = "tutmysql"
	case utils.MetaMongo:
		coreV1ConfDIR = "tutmongo"
	case utils.MetaPostgres:
		t.SkipNow()
	}
	for _, test := range sTestCoreSv1 {
		t.Run("CoreSv1 integration tests", test)
	}
}

func testCoreSv1LoadCofig(t *testing.T) {
	coreV1CfgPath = path.Join(*utils.DataDir, "conf", "samples", coreV1ConfDIR)
	var err error
	if coreV1Cfg, err = config.NewCGRConfigFromPath(coreV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testCoreSv1InitDataDB(t *testing.T) {
	if err := engine.InitDataDB(coreV1Cfg); err != nil {
		t.Error(err)
	}
}

func testCoreSv1InitStorDB(t *testing.T) {
	if err := engine.InitStorDb(coreV1Cfg); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StartEngineByExecWithCPUProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreV1CfgPath, "-cpuprof_dir", argPath)
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	var connected bool
	for i := 0; i < 16; i++ {
		time.Sleep(fib())
		if _, err := jsonrpc.Dial(utils.TCP, coreV1Cfg.ListenCfg().RPCJSONListen); err == nil {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", coreV1Cfg.ListenCfg().RPCJSONListen)
	}
}

func testCoreSv1RPCConn(t *testing.T) {
	var err error
	if coreV1Rpc, err = newRPCClient(coreV1Cfg.ListenCfg()); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StartCPUProfilingErrorAlreadyStarted(t *testing.T) {
	var reply string
	dirPath := &utils.DirectoryArgs{
		DirPath: argPath,
	}
	expectedErr := "start CPU profiling: already started"
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		dirPath, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testCoreSv1StartEngine(t *testing.T) {
	if _, err := engine.StartEngine(coreV1CfgPath, *utils.WaitRater); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StopMemProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := "stop memory profiling: not started yet"
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
		new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreSv1StartEngineByExecWIthMemProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreV1CfgPath,
		"-memprof_dir", argPath, "-memprof_interval", "100ms", "-memprof_maxfiles", "2", "-memprof_timestamp")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	var connected bool
	for i := 0; i < 16; i++ {
		time.Sleep(fib())
		if _, err := jsonrpc.Dial(utils.TCP, coreV1Cfg.ListenCfg().RPCJSONListen); err == nil {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", coreV1Cfg.ListenCfg().RPCJSONListen)
	}
	memProfNr = 3
}

func testCoreSv1StopMemoryProfiling(t *testing.T) {
	var reply string
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
		new(utils.TenantWithAPIOpts), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	time.Sleep(10 * time.Millisecond)
}

func testCoreSv1StartMemProfilingErrorAlreadyStarted(t *testing.T) {
	var reply string
	args := &cores.MemoryProfilingParams{
		DirPath:      argPath,
		Interval:     100 * time.Millisecond,
		MaxFiles:     2,
		UseTimestamp: true,
	}
	expErr := "start memory profiling: already started"
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
		args, &reply); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func testCoreSv1StopCPUProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := "stop CPU profiling: not started yet"
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreSv1StartCPUProfiling(t *testing.T) {
	var reply string
	dirPath := &utils.DirectoryArgs{
		DirPath: argPath,
	}
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		dirPath, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreSv1Sleep(t *testing.T) {
	args := &utils.DurationArgs{
		Duration: 600 * time.Millisecond,
	}
	var reply string
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1Sleep,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreSv1StopCPUProfiling(t *testing.T) {
	var reply string
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		new(utils.TenantWithAPIOpts), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	file, err := os.Open(path.Join(argPath, utils.CpuPathCgr))
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	//compare the size
	size, err := file.Stat()
	if err != nil {
		t.Error(err)
	} else if size.Size() < int64(300) {
		t.Errorf("Size of CPUProfile %v is lower that expected", size.Size())
	}
	//after we checked that CPUProfile was made successfully, can delete it
	if err := os.Remove(path.Join(argPath, utils.CpuPathCgr)); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StartMemoryProfiling(t *testing.T) {
	var reply string
	args := cores.MemoryProfilingParams{
		DirPath:      argPath,
		Interval:     100 * time.Millisecond,
		MaxFiles:     2,
		UseTimestamp: true,
	}
	if err := coreV1Rpc.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	memProfNr = 5
}

func testCoreSv1KillEngine(t *testing.T) {
	if err := engine.KillEngine(*utils.WaitRater); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	checkMemProfiles(t, argPath, memProfNr)
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

	if err := os.Remove(filepath.Join(memDirPath, utils.MemProfFinalFile)); err != nil &&
		!errors.Is(err, fs.ErrNotExist) {
		t.Error(err)
	}
	memProfNr = 0
}
