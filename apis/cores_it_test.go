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
)

var (
	coreItCfgPath string
	coreItDirPath string
	argPath       string
	coreItCfg     *config.CGRConfig
	coreItBiRPC   *birpc.Client
	coreItTests   = []func(t *testing.T){
		testCoreItLoadConfig,
		testCoreItInitDataDb,
		testCoreItInitStorDb,
		testCoreItStartEngineByExecWithCPUProfiling,
		testCoreItRpcConn,
		testCoreItStartCPUProfilingErrorAlreadyStarted,
		testCoreItSleep,
		testCoreItStopCPUProfiling,
		testCoreItKillEngine,
		testCoreItStartEngine,
		testCoreItRpcConn,
		testCoreItStopCPUProfilingBeforeStart,
		testCoreItStartCPUProfiling,
		testCoreItSleep,
		testCoreItStopCPUProfiling,
		testCoreItStatus,
		testCoreItKillEngine,
	}
)

func TestCoreItTests(t *testing.T) {
	argPath = "/tmp/cpu.prof"
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

func testCoreItStartEngineByExecWithCPUProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreItCfgPath, "-cpuprof_dir", "/tmp")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, coreItCfg.ListenCfg().RPCJSONListen); err != nil {
			t.Log(err)
		} else {
			connected = true
			break
		}
	}
	if !connected {
		t.Errorf("engine did not open port <%s>", coreItCfg.ListenCfg().RPCJSONListen)
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

func testCoreItStartCPUProfilingErrorAlreadyStarted(t *testing.T) {
	var reply string
	expectedErr := "CPU profiling already started"
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		argPath, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testCoreItSleep(t *testing.T) {
	args := &utils.DurationArgs{
		Duration: 500 * time.Millisecond,
	}
	var reply string
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1Sleep,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreItStopCPUProfiling(t *testing.T) {
	var reply string
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		utils.EmptyString, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
	file, err := os.Open(argPath)
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
	if err := os.Remove(argPath); err != nil {
		t.Error(err)
	}
}

func testCoreItStopCPUProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := " cannot stop because CPUProfiling is not active"
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1StopCPUProfiling,
		utils.EmptyString, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreItStartCPUProfiling(t *testing.T) {
	var reply string
	if err := coreItBiRPC.Call(context.Background(), utils.CoreSv1StartCPUProfiling,
		argPath, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
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
