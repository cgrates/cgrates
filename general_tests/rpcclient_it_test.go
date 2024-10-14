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

package general_tests

import (
	"flag"
	"fmt"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var testRemoteRALs = flag.Bool("remote_rals", false, "Perform the tests in integration mode, not by default.") // This flag will be passed here via "go test -local" args

var ( // shared vars
	rpcITCfgPath1, rpcITCfgPath2   string
	rpcITCfgDIR1, rpcITCfgDIR2     string
	rpcITCfg1, rpcITCfg2           *config.CGRConfig
	rpcRAL1, rpcRAL2               *rpcclient.RPCClient
	rpcPoolFirst, rpcPoolBroadcast *rpcclient.RPCPool
	ral1, ral2                     *exec.Cmd
	node1                          = "node1"
	node2                          = "node2"
)

var ( // configuration opts
	RemoteRALsAddr1 = "192.168.244.137:2012"
	RemoteRALsAddr2 = "192.168.244.138:2012"
)

// subtests to be executed
var sTestRPCITLcl = []func(t *testing.T){
	testRPCITLclInitCfg,
	testRPCITLclStartSecondEngine,
	testRPCITLclRpcConnPoolFirst,
	testRPCITLclStatusSecondEngine,
	testRPCITLclStartFirstEngine,
	testRPCITLclStatusFirstInitial,
	testRPCITLclStatusFirstFailover,
	testRPCITLclStatusFirstFailback,
	testRPCITLclTDirectedRPC,
	testRPCITLclRpcConnPoolBcast,
	testRPCITLclBcastStatusInitial,
	testRPCITLclBcastStatusNoRals1,
	testRPCITLclBcastStatusBcastNoRals,
	testRPCITLclBcastStatusRALs2Up,
	testRPCITLclStatusBcastRALs1Up,
}

func TestRPCITLcl(t *testing.T) {
	switch *utils.DBType {
	case utils.MetaInternal:
		rpcITCfgDIR1 = "multiral1_internal"
		rpcITCfgDIR2 = "multiral2_internal"
	case utils.MetaMySQL:
		rpcITCfgDIR1 = "multiral1_mysql"
		rpcITCfgDIR2 = "multiral2_mysql"
	case utils.MetaMongo:
		rpcITCfgDIR1 = "multiral1_mongo"
		rpcITCfgDIR2 = "multiral2_mongo"
	case utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestRPCITLcl {
		t.Run(*utils.DBType, stest)
	}
}

func testRPCITLclInitCfg(t *testing.T) {
	var err error
	rpcITCfgPath1 = path.Join(*utils.DataDir, "conf", "samples", rpcITCfgDIR1)
	rpcITCfgPath2 = path.Join(*utils.DataDir, "conf", "samples", rpcITCfgDIR2)
	rpcITCfg1, err = config.NewCGRConfigFromPath(rpcITCfgPath1)
	if err != nil {
		t.Error(err)
	}
	rpcITCfg2, err = config.NewCGRConfigFromPath(rpcITCfgPath2)
	if err != nil {
		t.Error(err)
	}
	if err := engine.InitDataDb(rpcITCfg1); err != nil {
		t.Fatal(err)
	}
}

func testRPCITLclStartSecondEngine(t *testing.T) {
	var err error
	if ral2, err = engine.StopStartEngine(rpcITCfgPath2, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRPCITLclRpcConnPoolFirst(t *testing.T) {
	rpcPoolFirst = rpcclient.NewRPCPool(rpcclient.PoolFirst, 0)
	var err error
	rpcRAL1, err = rpcclient.NewRPCClient(context.Background(), utils.TCP, rpcITCfg1.ListenCfg().RPCJSONListen, false, "", "", "", 3, 1,
		0, utils.FibDuration, time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err == nil {
		t.Fatal("Should receive cannot connect error here")
	}
	rpcPoolFirst.AddClient(rpcRAL1)
	rpcRAL2, err = rpcclient.NewRPCClient(context.Background(), utils.TCP, rpcITCfg2.ListenCfg().RPCJSONListen, false, "", "", "", 3, 1,
		0, utils.FibDuration, time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL2)
}

// Connect rpc client to rater
func testRPCITLclStatusSecondEngine(t *testing.T) {
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node2 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

// Start first engine
func testRPCITLclStartFirstEngine(t *testing.T) {
	var err error
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
}

// Connect rpc client to rater
func testRPCITLclStatusFirstInitial(t *testing.T) {
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node2 {
		t.Fatalf("Should receive ralID different than second one, got: %s", status[utils.NodeID].(string))
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node1 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node1, status[utils.NodeID].(string))
	}
}

// Connect rpc client to rater
func testRPCITLclStatusFirstFailover(t *testing.T) {
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node1 {
		t.Fatalf("Should receive ralID different than first one, got: %s", status[utils.NodeID].(string))
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node2 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

func testRPCITLclStatusFirstFailback(t *testing.T) {
	var err error
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == node2 {
		t.Error("Should receive new ID")
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // Make sure second time we land on the same instance
		t.Error(err)
	} else if status[utils.NodeID].(string) != node1 {
		t.Errorf("Expecting:\n%s\nReceived:\n%s", node2, status[utils.NodeID].(string))
	}
}

// Make sure it executes on the first node supporting the command
func testRPCITLclTDirectedRPC(t *testing.T) {
	var sessions []*sessions.ExternalSession
	if err := rpcPoolFirst.Call(context.Background(), utils.SessionSv1GetActiveSessions, utils.SessionFilter{}, &sessions); err == nil || err.Error() != utils.ErrNotFound.Error() {
		t.Error(err)
	}
}

// func testRPCITLclTimeout(t *testing.T) {
// 	var status map[string]any
// 	if err := rpcPoolFirst.Call(context.Background(),utils.CoreSv1Status, "10s", &status); err == nil {
// 		t.Error("Expecting timeout")
// 	} else if err.Error() != rpcclient.ErrReplyTimeout.Error() {
// 		t.Error(err)
// 	}
// }

// Connect rpc client to rater
func testRPCITLclRpcConnPoolBcast(t *testing.T) {
	rpcPoolBroadcast = rpcclient.NewRPCPool(rpcclient.PoolBroadcast, 2*time.Second)
	rpcPoolBroadcast.AddClient(rpcRAL1)
	rpcPoolBroadcast.AddClient(rpcRAL2)
}

func testRPCITLclBcastStatusInitial(t *testing.T) {
	var status map[string]any
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func testRPCITLclBcastStatusNoRals1(t *testing.T) {
	if err := ral1.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	var status map[string]any
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func testRPCITLclBcastStatusBcastNoRals(t *testing.T) {
	if err := ral2.Process.Kill(); err != nil { // Kill the first RAL
		t.Error(err)
	}
	time.Sleep(time.Duration(*utils.WaitRater) * time.Millisecond)
	var status map[string]any
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err == nil {
		t.Error("Should get error")
	}
}

func testRPCITLclBcastStatusRALs2Up(t *testing.T) {
	var err error
	if ral2, err = engine.StartEngine(rpcITCfgPath2, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
}

func testRPCITLclStatusBcastRALs1Up(t *testing.T) {
	var err error
	if ral1, err = engine.StartEngine(rpcITCfgPath1, *utils.WaitRater); err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
	if err := rpcPoolBroadcast.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty InstanceID received")
	}
}

/*
func TestRPCITStatusBcastCmd(t *testing.T) {
	var stats utils.CacheStats
	if err := rpcRAL1.Call(context.Background(),utils.APIerSv2GetCacheStats, utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastRatingLoadID != utils.NotAvailable || stats.LastAccountingLoadID != utils.NotAvailable {
		t.Errorf("Received unexpected stats: %+v", stats)
	}
	var loadInst utils.LoadInstance
	attrs := &utils.AttrLoadTpFromFolder{FolderPath: path.Join(*utils.DataDir, "tariffplans", "oldtutorial")}
	if err := rpcRAL1.Call(context.Background(),utils.APIerSv2LoadTariffPlanFromFolder, attrs, &loadInst); err != nil {
		t.Error(err)
	} else if loadInst.RatingLoadID == "" || loadInst.AccountingLoadID == "" {
		t.Errorf("Empty loadId received, loadInstance: %+v", loadInst)
	}
	var reply string
	if err := rpcPoolBroadcast.Call(context.Background(),utils.APIerSv1ReloadCache, utils.AttrReloadCache{}, &reply); err != nil {
		t.Error("Got error on APIerSv1.ReloadCache: ", err.Error())
	} else if reply != utils.OK {
		t.Error("Calling APIerSv1.ReloadCache got reply: ", reply)
	}
	if err := rpcRAL1.Call(context.Background(),utils.APIerSv2GetCacheStats, utils.AttrCacheStats{}, &stats); err != nil {
		t.Error(err)
	} else if stats.LastRatingLoadID != loadInst.RatingLoadID {
		t.Errorf("Received unexpected stats:  %+v vs %+v", stats, loadInst)
	}
	if err := rpcRAL2.Call(context.Background(),utils.APIerSv2GetCacheStats, utils.AttrCacheStats{}, &stats); err != nil {
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
	rpcPoolFirst = rpcclient.NewRPCPool(rpcclient.PoolFirst, 0)
	rpcRALRmt, err := rpcclient.NewRPCClient(context.Background(), utils.TCP, RemoteRALsAddr1, false, "", "", "", 1, 1,
		0, utils.FibDuration, time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRALRmt)
	rpcRAL1, err = rpcclient.NewRPCClient(context.Background(), utils.TCP, RemoteRALsAddr2, false, "", "", "", 1, 1,
		0, utils.FibDuration, time.Second, 2*time.Second, rpcclient.JSONrpc, nil, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	rpcPoolFirst.AddClient(rpcRAL1)
}

func TestRPCITRmtStatusFirstInitial(t *testing.T) {
	if !*testRemoteRALs {
		return
	}
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil { // Make sure second time we land on the same instance
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
		time.Sleep(time.Second)
	}
	fmt.Println("\n\nExecuting query ...")
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node1 {
		t.Fatal("Did not failover")
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
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
		time.Sleep(time.Second)
	}
	fmt.Println("\n\nExecuting query ...")
	var status map[string]any
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
		t.Error(err)
	} else if status[utils.NodeID].(string) == "" {
		t.Error("Empty NodeID received")
	} else if status[utils.NodeID].(string) == node2 {
		t.Fatal("Did not do failback")
	}
	if err := rpcPoolFirst.Call(context.Background(), utils.CoreSv1Status, utils.TenantWithAPIOpts{}, &status); err != nil {
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
