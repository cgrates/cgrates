/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can Storagetribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITH*out ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package general_tests

import (
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var rpcITCfgPath1, rpcITCfgPath2 string
var rpcITCfg1, rpcITCfg2 *config.CGRConfig
var rpcRAL1, rpcRAL2 *rpcclient.RpcClient
var rpcPoolFirst *rpcclient.RpcClientPool
var ral1, ral2 *exec.Cmd
var err error
var ral1ID, ral2ID string

func TestRPCITInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	rpcITCfgPath1 = path.Join(*dataDir, "conf", "samples", "multiral1")
	rpcITCfgPath2 = path.Join(*dataDir, "conf", "samples", "multiral2")
	// Init config first
	rpcITCfg1, err = config.NewCGRConfigFromFolder(rpcITCfgPath1)
	if err != nil {
		t.Error(err)
	}
	rpcITCfg2, err = config.NewCGRConfigFromFolder(rpcITCfgPath2)
	if err != nil {
		t.Error(err)
	}
}

func TestRPCITStartEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral1, err = engine.StopStartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	if ral2, err = engine.StartEngine(rpcITCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestRPCITRpcConnPool(t *testing.T) {
	if !*testIntegration {
		return
	}
	rpcPoolFirst = rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST)
	rpcRAL1, err = rpcclient.NewRpcClient("tcp", rpcITCfg1.RPCJSONListen, 3, 1, rpcclient.JSON_RPC, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL1)
	rpcRAL2, err = rpcclient.NewRpcClient("tcp", rpcITCfg2.RPCJSONListen, 3, 1, rpcclient.JSON_RPC, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL2)
}

// Connect rpc client to rater
func TestRPCITStatusFirstInitial(t *testing.T) {
	if !*testIntegration {
		return
	}
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else {
		ral1ID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ral1ID {
		t.Errorf("Expecting: %s, received: %s", ral1ID, status[utils.InstanceID].(string))
	}
}

// Connect rpc client to rater
func TestRPCITStatusFirstFailover(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else {
		ral1ID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ral1ID {
		t.Errorf("Expecting: %s, received: %s", ral1ID, status[utils.InstanceID].(string))
	} else {
		ral2ID = status[utils.InstanceID].(string)
	}
}

func TestRPCITStatusFirstFailback(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == ral2ID {
		t.Error("Should receive new ID")
	} else {
		ral1ID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ral1ID {
		t.Errorf("Expecting: %s, received: %s", ral1ID, status[utils.InstanceID].(string))
	}
}
