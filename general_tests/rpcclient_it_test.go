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
	"flag"
	"fmt"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var rpcITCfgPath1, rpcITCfgPath2 string
var rpcITCfg1, rpcITCfg2 *config.CGRConfig
var rpcRAL1, rpcRAL2 *rpcclient.RpcClient
var rpcPoolFirst, rpcPoolBroadcast *rpcclient.RpcClientPool
var ral1, ral2 *exec.Cmd
var err error
var ral1ID, ral2ID, ralRmtID string

var testRemoteRALs = flag.Bool("remote_rals", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args

func TestRPCITInitCfg(t *testing.T) {
	if !*testIntegration {
		return
	}
	rpcITCfgPath1 = path.Join(*dataDir, "conf", "samples", "multiral1")
	rpcITCfgPath2 = path.Join(*dataDir, "conf", "samples", "multiral2")
	rpcITCfg1, err = config.NewCGRConfigFromFolder(rpcITCfgPath1)
	if err != nil {
		t.Error(err)
	}
	rpcITCfg2, err = config.NewCGRConfigFromFolder(rpcITCfgPath2)
	if err != nil {
		t.Error(err)
	}
	if err := engine.InitDataDb(rpcITCfg1); err != nil {
		t.Fatal(err)
	}
}

func TestRPCITStartSecondEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral2, err = engine.StopStartEngine(rpcITCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestRPCITRpcConnPoolFirst(t *testing.T) {
	if !*testIntegration {
		return
	}
	rpcPoolFirst = rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST, 0)
	rpcRAL1, err = rpcclient.NewRpcClient("tcp", rpcITCfg1.RPCJSONListen, 3, 1, time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil)
	if err == nil {
		t.Fatal("Should receive cannot connect error here")
	}
	rpcPoolFirst.AddClient(rpcRAL1)
	rpcRAL2, err = rpcclient.NewRpcClient("tcp", rpcITCfg2.RPCJSONListen, 3, 1, time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL2)
}

// Connect rpc client to rater
func TestRPCITStatusSecondEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else {
		ral2ID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ral2ID {
		t.Errorf("Expecting: %s, received: %s", ral2ID, status[utils.InstanceID].(string))
	}
}

// Start first engine
func TestRPCITStartFirstEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
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
	} else if status[utils.InstanceID].(string) == ral2ID {
		t.Fatalf("Should receive ralID different than second one, got: %s", status[utils.InstanceID].(string))
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
	} else if status[utils.InstanceID].(string) == ral1ID {
		t.Fatalf("Should receive ralID different than first one, got: %s", status[utils.InstanceID].(string))
	} else {
		ral2ID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ral2ID {
		t.Errorf("Expecting: %s, received: %s", ral2ID, status[utils.InstanceID].(string))
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

// Make sure it executes on the first node supporting the command
func TestRPCITDirectedRPC(t *testing.T) {
	if !*testIntegration {
		return
	}
	var sessions []*sessionmanager.ActiveSession
	if err := rpcPoolFirst.Call("SMGenericV1.ActiveSessions", utils.AttrSMGGetActiveSessions{}, &sessions); err != nil {
		t.Error(err) // {"id":2,"result":null,"error":"rpc: can't find service SMGenericV1.ActiveSessions"}
	} else if len(sessions) != 0 {
		t.Errorf("Received sessions: %+v", sessions)
	}
}

// Special tests involving remote server (manually set)
// The server network will be manually disconnected without TCP close
func TestRPCITRmtRpcConnPool(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	rpcPoolFirst = rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST, 0)
	rpcRALRmt, err := rpcclient.NewRpcClient("tcp", "172.16.254.83:2012", 1, 1, time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRALRmt)
	rpcRAL1, err = rpcclient.NewRpcClient("tcp", rpcITCfg1.RPCJSONListen, 1, 1, time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL1)
}

func TestRPCITRmtStatusFirstInitial(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else if status[utils.InstanceID].(string) == ral1ID {
		t.Fatal("Should receive ralID different than first one")
	} else {
		ralRmtID = status[utils.InstanceID].(string)
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.InstanceID].(string) != ralRmtID {
		t.Errorf("Expecting: %s, received: %s", ralRmtID, status[utils.InstanceID].(string))
	}
}

func TestRPCITRmtStatusFirstFailover(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	fmt.Println("Ready for doing failover")
	remaining := 5
	for i := 0; i < remaining; i++ {
		fmt.Printf("\n\t%d", remaining-i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n\nExecuting query ...")
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else if status[utils.InstanceID].(string) != ral1ID {
		t.Fatal("Did not do failover")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else if status[utils.InstanceID].(string) != ral1ID {
		t.Fatal("Did not do failover")
	}
}

func TestRPCITRmtStatusFirstFailback(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	fmt.Println("Ready for doing failback")
	remaining := 10
	for i := 0; i < remaining; i++ {
		fmt.Printf("\n\t%d", remaining-i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println("\n\nExecuting query ...")
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else if status[utils.InstanceID].(string) != ralRmtID {
		t.Fatal("Did not do failback")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	} else if status[utils.InstanceID].(string) != ralRmtID {
		t.Fatal("Did not do failback")
	}
}

// Connect rpc client to rater
func TestRPCITRpcConnPoolBcast(t *testing.T) {
	if !*testIntegration {
		return
	}
	rpcPoolBroadcast = rpcclient.NewRpcClientPool(rpcclient.POOL_BROADCAST, time.Duration(2*time.Second))
	rpcPoolBroadcast.AddClient(rpcRAL1)
	rpcPoolBroadcast.AddClient(rpcRAL2)
}

func TestRPCITBcastStatusInitial(t *testing.T) {
	if !*testIntegration {
		return
	}
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

func TestRPCITBcastStatusNoRals1(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

func TestRPCITBcastStatusBcastNoRals(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := ral2.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err == nil {
		t.Error("Should get error")
	}
}

func TestRPCITBcastStatusRALs2Up(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral2, err = engine.StartEngine(rpcITCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

func TestRPCITStatusBcastRALs1Up(t *testing.T) {
	if !*testIntegration {
		return
	}
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.InstanceID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

func TestRPCITStatusBcastCmd(t *testing.T) {
	if !*testIntegration {
		return
	}
	var stats utils.CacheStats
	if err := rpcRAL1.Call("ApierV2.GetCacheStats", utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastLoadId != utils.NOT_AVAILABLE {
		t.Errorf("Received unexpected stats: %+v", stats)
	}
	var loadInst utils.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "tutorial")}
	if err := rpcRAL1.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	} else if loadInst.LoadId == "" {
		t.Errorf("Empty loadId received, loadInstance: %+v", loadInst)
	}
	var reply string
	if err := rpcPoolBroadcast.Call("ApierV1.ReloadCache", utils.AttrReloadCache{}, &reply); err != nil {
		t.Error("Got error on ApierV1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling ApierV1.ReloadCache got reply: ", reply)
	}
	if err := rpcRAL1.Call("ApierV2.GetCacheStats", utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastLoadId != loadInst.LoadId {
		t.Errorf("Received unexpected stats: %+v", stats)
	}
	if err := rpcRAL2.Call("ApierV2.GetCacheStats", utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastLoadId != loadInst.LoadId {
		t.Errorf("Received unexpected stats: %+v", stats)
	}
}

func TestRPCITStopCgrEngine(t *testing.T) {
	if !*testIntegration {
		return
	}
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
