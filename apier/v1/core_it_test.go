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

package v1

import (
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/engine"

	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
)

var (
	coreV1CfgPath string
	coreV1Cfg     *config.CGRConfig
	coreV1Rpc     *rpc.Client
	coreV1ConfDIR string //run tests for specific configuration
	argPath       string
	sTestCoreSv1  = []func(t *testing.T){
		testCoreSv1LoadCofig,
		testCoreSv1InitDataDB,
		testCoreSv1InitStorDB,
		testCoreSv1StartEngineByExecWithCPUProfiling,
		testCoreSv1RPCConn,
		testCoreSv1StartCPUProfilingErrorAlreadyStarted,
		testCoreSv1Sleep,
		testCoreSv1StopCPUProfiling,
		testCoreSv1KillEngine,
		testCoreSv1StartEngine,
		testCoreSv1RPCConn,
		testCoreSv1StopCPUProfilingBeforeStart,
		testCoreSv1StartCPUProfiling,
		testCoreSv1Sleep,
		testCoreSv1StopCPUProfiling,
		testCoreSv1KillEngine,
	}
)

func TestITCoreSv1(t *testing.T) {
	argPath = "/tmp"
	switch *dbType {
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
	coreV1CfgPath = path.Join(*dataDir, "conf", "samples", coreV1ConfDIR)
	var err error
	if coreV1Cfg, err = config.NewCGRConfigFromPath(coreV1CfgPath); err != nil {
		t.Error(err)
	}
}

func testCoreSv1InitDataDB(t *testing.T) {
	if err := engine.InitDataDb(coreV1Cfg); err != nil {
		t.Error(err)
	}
}

func testCoreSv1InitStorDB(t *testing.T) {
	if err := engine.InitStorDb(coreV1Cfg); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StartEngineByExecWithCPUProfiling(t *testing.T) {
	engine := exec.Command("cgr-engine", "-config_path", coreV1CfgPath, "-cpuprof_dir", "/tmp")
	if err := engine.Start(); err != nil {
		t.Error(err)
	}
	fib := utils.Fib()
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(time.Duration(fib()) * time.Millisecond)
		if _, err := jsonrpc.Dial(utils.TCP, coreV1Cfg.ListenCfg().RPCJSONListen); err != nil {
			t.Log(err)
		} else {
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
	expectedErr := "CPU profiling already started"
	if err := coreV1Rpc.Call(utils.CoreSv1StartCPUProfiling,
		argPath, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+v, received %+v", expectedErr, err)
	}
}

func testCoreSv1StartEngine(t *testing.T) {
	if _, err := engine.StartEngine(coreV1CfgPath, *waitRater); err != nil {
		t.Error(err)
	}
}

func testCoreSv1StopCPUProfilingBeforeStart(t *testing.T) {
	var reply string
	expectedErr := " cannot stop because CPUProfiling is not active"
	if err := coreV1Rpc.Call(utils.CoreSv1StopCPUProfiling,
		utils.EmptyString, &reply); err == nil || err.Error() != expectedErr {
		t.Errorf("Expected %+q, received %+q", expectedErr, err)
	}
}

func testCoreSv1StartCPUProfiling(t *testing.T) {
	var reply string
	if err := coreV1Rpc.Call(utils.CoreSv1StartCPUProfiling,
		argPath, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreSv1Sleep(t *testing.T) {
	args := &utils.DurationArgs{
		Duration: 500 * time.Millisecond,
	}
	var reply string
	if err := coreV1Rpc.Call(utils.CoreSv1Sleep,
		args, &reply); err != nil {
		t.Error(err)
	} else if reply != utils.OK {
		t.Errorf("Unexpected reply returned")
	}
}

func testCoreSv1StopCPUProfiling(t *testing.T) {
	var reply string
	if err := coreV1Rpc.Call(utils.CoreSv1StopCPUProfiling,
		utils.EmptyString, &reply); err != nil {
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

func testCoreSv1KillEngine(t *testing.T) {
	if err := engine.KillEngine(*waitRater); err != nil {
		t.Error(err)
	}
}
