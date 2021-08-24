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
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package apis

import (
	"fmt"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	coreSCfgPath string
	coreSCfg     *config.CGRConfig
	coreSBiRpc   *birpc.Client
	coreSConfDIR string //run tests for specific configuration
	argPath      string
	sTestCoreIt  = []func(t *testing.T){
		testCoreItLoadCofig,
		testCoreItInitDataDB,
		testCoreItInitStorDB,

		//engine separate with cpu
		testCoreItStartEngineByExecWithCPUProfiling,
		testCoreItRPCConn,
		testCoreItStartCPUProfilingErrorAlreadyStarted,
		testCoreItSleep,
		testCoreItStopCPUProfiling,
		//status api
		testCoreItStatus,
		testCoreItKillEngine,

		//engine separate with memory
		testCoreItStartEngineByExecWIthMemProfiling,
		testCoreItRPCConn,
		testCoreItStartMemProfilingErrorAlreadyStarted,
		testCoreItSleep,
		testCoreItStopMemoryProfiling,
		testCoreItKillEngine,
		testCoreItCheckFinalMemProfiling,
		// test CPU and Memory just by APIs
		testCoreItStartEngine,
		testCoreItRPCConn,

		//CPUProfiles apis
		testCoreItStopCPUProfilingBeforeStart,
		testCoreItStartCPUProfiling,
		testCoreItSleep,
		testCoreItStopCPUProfiling,

		//MemoryProfiles apis
		testCoreItStopMemProfilingBeforeStart,
		testCoreItStartMemoryProfiling,
		testCoreItSleep,
		testCoreItStopMemoryProfiling,
		testCoreItKillEngine,
		testCoreItCheckFinalMemProfiling,
	}
)

func TestITCoreIt(t *testing.T) {
	argPath = "/tmp"
	switch *dbType {
	case utils.MetaInternal, utils.MetaMySQL, utils.MetaMongo:
		coreSConfDIR = "core_config"
	case utils.MetaPostgres:
		t.SkipNow()
	}
	for _, test := range sTestCoreIt {
		t.Run("CoreIt integration tests", test)
	}
}

func testCoreItLoadCofig(t *testing.T) {
	coreSCfgPath = path.Join(*dataDir, "conf", "samples", coreSConfDIR)
	var err error
	if coreSCfg, err = config.NewCGRConfigFromPath(coreSCfgPath); err != nil {
		t.Error(err)
	}
}

func testCoreItInitDataDB(t *testing.T) {
	if err := engine.InitDataDB(coreSCfg); err != nil {
		t.Error(err)
	}
}

func testCoreItInitStorDB(t *testing.T) {
	if err := engine.InitStorDB(coreSCfg); err != nil {
		t.Error(err)
	}
}

func testCoreItStartEngineByExecWithCPUProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreSCfgPath, "-cpuprof_dir", argPath)
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, coreSCfg.ListenCfg().RPCJSONListen); err != nil {
			t.Log(err)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", coreSCfg.ListenCfg().RPCJSONListen)
	}
}

func testCoreItRPCConn(t *testing.T) {
	var err error
	if coreSBiRpc, err = newRPCClient(coreSCfg.ListenCfg()); err != nil {
		t.Error(err)
	}
}

func testCoreItStartCPUProfilingErrorAlreadyStarted(t *testing.T) {
	var reply string
	dirPath := &utils.DirectoryArgs{
		DirPath: argPath,
	}
	expectedErr := "CPU profiling already started"
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		dirPath, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testCoreItStartEngine(t *testing.T) {
	if _, err := engine.StartEngine(coreSCfgPath, *waitRater); err != nil {
		t.Error(err)
	}
}

func testCoreItStopMemProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := " Memory Profiling is not started"
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
		new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreItStartEngineByExecWIthMemProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreSCfgPath,
		"-memprof_dir", argPath, "-memprof_interval", "100ms", "-memprof_nrfiles", "2")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, coreSCfg.ListenCfg().RPCJSONListen); err != nil {
			t.Log(err)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", coreSCfg.ListenCfg().RPCJSONListen)
	}
}

func testCoreItStopMemoryProfiling(t *testing.T) {
	var reply string
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StopMemoryProfiling,
		new(utils.TenantWithAPIOpts), &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}

	//mem_prof1, mem_prof2
	for i := 1; i <= 2; i++ {
		file, err := os.Open(path.Join(argPath, fmt.Sprintf("mem%v.prof", i)))
		if err != nil {
			t.Error(err)
		}
		defer file.Close()

		//compare the size
		size, err := file.Stat()
		if err != nil {
			t.Error(err)
		} else if size.Size() < int64(415) {
			t.Errorf("Size of MemoryProfile %v is lower that expected", size.Size())
		}
		//after we checked that CPUProfile was made successfully, can delete it
		if err := os.Remove(path.Join(argPath, fmt.Sprintf("mem%v.prof", i))); err != nil {
			t.Error(err)
		}
	}
}

func testCoreItCheckFinalMemProfiling(t *testing.T) {
	// as the engine was killed, mem_final.prof was created and we must check it
	file, err := os.Open(path.Join(argPath, fmt.Sprintf(utils.MemProfFileCgr)))
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	//compare the size
	size, err := file.Stat()
	if err != nil {
		t.Error(err)
	} else if size.Size() < int64(415) {
		t.Errorf("Size of MemoryProfile %v is lower that expected", size.Size())
	}
	//after we checked that CPUProfile was made successfully, can delete it
	if err := os.Remove(path.Join(argPath, fmt.Sprintf(utils.MemProfFileCgr))); err != nil {
		t.Error(err)
	}
}

func testCoreItStartMemProfilingErrorAlreadyStarted(t *testing.T) {
	var reply string
	args := &utils.MemoryPrf{
		DirPath:  argPath,
		Interval: 100 * time.Millisecond,
		NrFiles:  2,
	}
	expErr := "Memory Profiling already started"
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
		args, &reply); err == nil || err.Error() != expErr {
		t.Errorf("Expected %+v, received %+v", expErr, err)
	}
}

func testCoreItStopCPUProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := " cannot stop because CPUProfiling is not active"
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		new(utils.TenantWithAPIOpts), &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreItStartCPUProfiling(t *testing.T) {
	var reply string
	dirPath := &utils.DirectoryArgs{
		DirPath: argPath,
	}
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		dirPath, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreItSleep(t *testing.T) {
	args := &utils.DurationArgs{
		Duration: 600 * time.Millisecond,
	}
	var reply string
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1Sleep,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreItStopCPUProfiling(t *testing.T) {
	var reply string
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		new(utils.TenantIDWithAPIOpts), &reply); err != nil {
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
	} else if size.Size() < int64(415) {
		t.Errorf("Size of CPUProfile %v is lower that expected", size.Size())
	}
	//after we checked that CPUProfile was made successfully, can delete it
	if err := os.Remove(path.Join(argPath, utils.CpuPathCgr)); err != nil {
		t.Error(err)
	}
}

func testCoreItStartMemoryProfiling(t *testing.T) {
	var reply string
	args := &utils.MemoryPrf{
		DirPath:  argPath,
		Interval: 100 * time.Millisecond,
		NrFiles:  2,
	}
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1StartMemoryProfiling,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreItStatus(t *testing.T) {
	args := &utils.TenantIDWithAPIOpts{}
	var reply map[string]interface{}
	if err := coreSBiRpc.Call(context.Background(), utils.CoreSv1Status,
		args, &reply); err != nil {
		t.Fatal(err)
	} else if reply[utils.NodeID] != "Cores_apis_test" {
		t.Errorf("Expected ALL2 but received %v", reply[utils.NodeID])
	}
}

func testCoreItKillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
	time.Sleep(500 * time.Millisecond)
}
