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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var testRemoteRALs = flag.Bool("remote_rals", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args

var ( // shared vars
	rpcITCfgPath1, rpcITCfgPath2   string
	rpcITCfg1, rpcITCfg2           *config.CGRConfig
	rpcRAL1, rpcRAL2               *rpcclient.RpcClient
	rpcPoolFirst, rpcPoolBroadcast *rpcclient.RpcClientPool
	ral1, ral2                     *exec.Cmd
	err                            error
	node1                          = "node1"
	node2                          = "node2"
)

var ( // configuration opts
	RemoteRALsAddr1 = "192.168.244.137:2012"
	RemoteRALsAddr2 = "192.168.244.138:2012"
)

func TestRPCITLclInitCfg(t *testing.T) {
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

func TestRPCITLclStartSecondEngine(t *testing.T) {
	if ral2, err = engine.StopStartEngine(rpcITCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestRPCITLclRpcConnPoolFirst(t *testing.T) {
	rpcPoolFirst = rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST, 0)
	rpcRAL1, err = rpcclient.NewRpcClient("tcp", rpcITCfg1.ListenCfg().RPCJSONListen, false, "", "", "", 3, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil, false)
	if err == nil {
		t.Fatal("Should receive cannot connect error here")
	}
	rpcPoolFirst.AddClient(rpcRAL1)
	rpcRAL2, err = rpcclient.NewRpcClient("tcp", rpcITCfg2.ListenCfg().RPCJSONListen, false, "", "", "", 3, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL2)
}

// Connect rpc client to rater
func TestRPCITLclStatusSecondEngine(t *testing.T) {
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node2 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

// Start first engine
func TestRPCITLclStartFirstEngine(t *testing.T) {
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func TestRPCITLclStatusFirstInitial(t *testing.T) {
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node2 {
		t.Fatalf("Should receive ralID different than second one, got: %s", status[utils.NodeID].(string))
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node1 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node1, status[utils.NodeID].(string))
	}
}

// Connect rpc client to rater
func TestRPCITLclStatusFirstFailover(t *testing.T) {
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node1 {
		t.Fatalf("Should receive ralID different than first one, got: %s", status[utils.NodeID].(string))
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node2 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

func TestRPCITLclStatusFirstFailback(t *testing.T) {
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == node2 {
		t.Error("Should receive new ID")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node1 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

// Make sure it executes on the first node supporting the command
func TestRPCITLclTDirectedRPC(t *testing.T) {
	var sessions []*sessions.ActiveSession
	if err := rpcPoolFirst.Call("SMGenericV1.GetActiveSessions", map[string]string{}, &sessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

func TestRPCITLclTimeout(t *testing.T) {
	var status map[string]interface{}
	if err := rpcPoolFirst.Call("Responder.Status", "10s", &status); err == nil {
		t.Error("Expecting timeout")
	} else if err.Error() != rpcclient.ErrReplyTimeout.Error() {
		t.Error(err)
	}
}

// Connect rpc client to rater
func TestRPCITLclRpcConnPoolBcast(t *testing.T) {
	rpcPoolBroadcast = rpcclient.NewRpcClientPool(rpcclient.POOL_BROADCAST, time.Duration(2*time.Second))
	rpcPoolBroadcast.AddClient(rpcRAL1)
	rpcPoolBroadcast.AddClient(rpcRAL2)
}

func TestRPCITLclBcastStatusInitial(t *testing.T) {
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func TestRPCITLclBcastStatusNoRals1(t *testing.T) {
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func TestRPCITLclBcastStatusBcastNoRals(t *testing.T) {
	if err := ral2.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*waitRater) * time.Millisecond)
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err == nil {
		t.Error("Should get error")
	}
}

func TestRPCITLclBcastStatusRALs2Up(t *testing.T) {
	if ral2, err = engine.StartEngine(rpcITCfgPath2, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func TestRPCITLclStatusBcastRALs1Up(t *testing.T) {
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *waitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]interface{}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

/*
func TestRPCITStatusBcastCmd(t *testing.T) {
	var stats utils.CacheStats
	if err := rpcRAL1.Call("ApierV2.GetCacheStats", utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastRatingLoadID != utils.NOT_AVAILABLE || stats.LastAccountingLoadID != utils.NOT_AVAILABLE {
		t.Errorf("Received unexpected stats: %+v", stats)
	}
	var loadInst utils.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*dataDir, "tariffplans", "oldtutorial")}
	if err := rpcRAL1.Call("ApierV2.LoadTariffPlanFromFolder", attrs, &loadInst); err != nil {
		t.Error(err)
	} else if loadInst.RatingLoadID == "" || loadInst.AccountingLoadID == "" {
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
	} else if stats.LastRatingLoadID != loadInst.RatingLoadID {
		t.Errorf("Received unexpected stats:  %+v vs %+v", stats, loadInst)
	}
	if err := rpcRAL2.Call("ApierV2.GetCacheStats", utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastRatingLoadID != loadInst.RatingLoadID {
		t.Errorf("Received unexpected stats: %+v vs %+v", stats, loadInst)
	}
}
*/

// Special tests involving remote server (manually set)
// The server network will be manually disconnected without TCP close
// Run remote ones with: go test -tags=integration -run="TestRPCITRmt|TestRPCITStop" -remote_rals
func TestRPCITRmtRpcConnPool(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	rpcPoolFirst = rpcclient.NewRpcClientPool(rpcclient.POOL_FIRST, 0)
	rpcRALRmt, err := rpcclient.NewRpcClient("tcp", RemoteRALsAddr1, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRALRmt)
	rpcRAL1, err = rpcclient.NewRpcClient("tcp", RemoteRALsAddr2, false, "", "", "", 1, 1,
		time.Duration(1*time.Second), time.Duration(2*time.Second), rpcclient.JSON_RPC, nil, false)
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
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node1 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node1, status[utils.NodeID].(string))
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
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node1 {
		t.Fatal("Did not failover")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) != node2 {
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
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node2 {
		t.Fatal("Did not do failback")
	}
	if err := rpcPoolFirst.Call("Responder.Status", "", &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) != node1 {
		t.Fatal("Did not do failback")
	}
}

func TestRPCITStopCgrEngine(t *testing.T) {
	if err := engine.KillEngine(100); err != nil {
		t.Error(err)
	}
}
