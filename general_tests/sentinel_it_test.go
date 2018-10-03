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
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	node1ConfigPath                    = path.Join(*dataDir, "sentinel", "node1.conf")
	node2ConfigPath                    = path.Join(*dataDir, "sentinel", "node2.conf")
	sentinelConfigPath                 = path.Join(*dataDir, "sentinel", "sentinel1.conf")
	engineConfigPath                   = path.Join(*dataDir, "conf", "samples", "tutsentinel")
	sentinelConfig                     *config.CGRConfig
	sentinelRPC                        *rpc.Client
	node1exec, node2exec, sentinelexec *exec.Cmd
	redisSentinel                      = flag.Bool("redis_sentinel", false, "Run tests with redis sentinel")
)

var sTestsRds = []func(t *testing.T){
	testRedisSentinelStartNodes,
	testRedisSentinelInitConfig,
	testRedisSentinelFlushDb,
	testRedisSentinelStartEngine,
	testRedisSentinelRPCCon,
	testRedisSentinelSetGetAttribute,
	testRedisSentinelShutDownNode1,
	testRedisSentinelGetAttrAfterFailover,
	testRedisSentinelKillEngine,
}

// Before running these tests make sure node1.conf, node2.conf, sentinel1.conf are the next
// Node1 will be master and start at port 16379
// Node2 will be slave of node1 and start at port 16380
// Sentinel will be started at port 16381 and will watch Node1
func TestRedisSentinel(t *testing.T) {
	if !*redisSentinel {
		return
	}
	for _, stest := range sTestsRds {
		t.Run("", stest)
	}
}

func testRedisSentinelStartNodes(t *testing.T) {
	node1exec = exec.Command("redis-server", node1ConfigPath)
	if err := node1exec.Start(); err != nil {
		t.Error(err)
	}
	node2exec = exec.Command("redis-server", node2ConfigPath)
	if err := node2exec.Start(); err != nil {
		t.Error(err)
	}
	sentinelexec = exec.Command("redis-sentinel", sentinelConfigPath)
	if err := sentinelexec.Start(); err != nil {
		t.Error(err)
	}
}

func testRedisSentinelInitConfig(t *testing.T) {
	var err error
	sentinelConfig, err = config.NewCGRConfigFromFolder(engineConfigPath)
	if err != nil {
		t.Error(err)
	}
	sentinelConfig.DataFolderPath = *dataDir // Share DataFolderPath through config towards StoreDb for Flush()
	config.SetCgrConfig(sentinelConfig)
}

func testRedisSentinelFlushDb(t *testing.T) {
	if err := engine.InitDataDb(sentinelConfig); err != nil {
		t.Fatal(err)
	}
}

func testRedisSentinelStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(engineConfigPath, 2000); err != nil {
		t.Fatal(err)
	}
}

func testRedisSentinelRPCCon(t *testing.T) {
	var err error
	sentinelRPC, err = jsonrpc.Dial("tcp", sentinelConfig.RPCJSONListen) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRedisSentinelSetGetAttribute(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:Account:1001"},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true),
				Append:     true,
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var result string
	if err := sentinelRPC.Call("ApierV1.SetAttributeProfile", alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := sentinelRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

// Here we kill node1 and sentinel will do failover and promote node2 to be master
func testRedisSentinelShutDownNode1(t *testing.T) {
	if err := node1exec.Process.Kill(); err != nil { // Kill the master
		t.Error(err)
	}
}

// After we kill node1 check the data if was replicated in node2
func testRedisSentinelGetAttrAfterFailover(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:Account:1001"},
		Attributes: []*engine.Attribute{
			&engine.Attribute{
				FieldName:  utils.Subject,
				Initial:    utils.ANY,
				Substitute: config.NewRSRParsersMustCompile("1001", true),
				Append:     true,
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := sentinelRPC.Call("ApierV1.GetAttributeProfile",
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Error(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testRedisSentinelKillEngine(t *testing.T) {
	if err := engine.KillEngine(2000); err != nil {
		t.Error(err)
	}
}
