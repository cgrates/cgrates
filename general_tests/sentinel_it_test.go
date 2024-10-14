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
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	node1ConfigPath     = path.Join(*utils.DataDir, "redisSentinel", "node1.conf")
	node2ConfigPath     = path.Join(*utils.DataDir, "redisSentinel", "node2.conf")
	sentinel1ConfigPath = path.Join(*utils.DataDir, "redisSentinel", "sentinel1.conf")
	sentinel2ConfigPath = path.Join(*utils.DataDir, "redisSentinel", "sentinel2.conf")
	engineConfigPath    = path.Join(*utils.DataDir, "conf", "samples", "redisSentinel")
	sentinelConfig      *config.CGRConfig
	sentinelRPC         *birpc.Client
	node1Exec           *exec.Cmd
	node2Exec           *exec.Cmd
	stlExec1            *exec.Cmd
	stlExec2            *exec.Cmd
	redisSentinel       = flag.Bool("redisSentinel", false, "Run tests with redis sentinel")

	sTestsRds = []func(t *testing.T){
		testRedisSentinelStartNodes,
		testRedisSentinelInitConfig,
		testRedisSentinelFlushDb,
		testRedisSentinelStartEngine,
		testRedisSentinelRPCCon,
		testRedisSentinelSetGetAttribute,
		testRedisSentinelInsertion,
		testRedisSentinelGetAttrAfterFailover,
		testRedisSentinelKillEngine,
	}
)

// Before running these tests make sure node1.conf, node2.conf, sentinel1.conf are the next
// Node1 will be master and start at port 16379
// Node2 will be slave of node1 and start at port 16380
// Sentinel1 will be started at port 16381 and will watch Node1
// Sentinel2 will be started at port 16382 and will watch Node1
// Also make sure that redis process is stopped
func TestRedisSentinel(t *testing.T) {
	if !*redisSentinel {
		return
	}
	switch *utils.DBType {
	case utils.MetaMySQL:
	case utils.MetaInternal,
		utils.MetaMongo,
		utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRds {
		t.Run("TestRedisSentinel", stest)
	}
}

func testRedisSentinelStartNodes(t *testing.T) {
	if err := os.MkdirAll("/tmp/sentinel/", 0755); err != nil {
		t.Fatal("Error creating folder: /tmp/sentinel/ ", err)
	}

	node1Exec = exec.Command("redis-server", node1ConfigPath)
	if err := node1Exec.Start(); err != nil {
		t.Error(err)
	}
	node2Exec = exec.Command("redis-server", node2ConfigPath)
	if err := node2Exec.Start(); err != nil {
		t.Error(err)
	}
	stlExec1 = exec.Command("redis-sentinel", sentinel1ConfigPath)
	if err := stlExec1.Start(); err != nil {
		t.Error(err)
	}
	stlExec2 = exec.Command("redis-sentinel", sentinel2ConfigPath)
	if err := stlExec2.Start(); err != nil {
		t.Error(err)
	}
}

func testRedisSentinelInitConfig(t *testing.T) {
	var err error
	sentinelConfig, err = config.NewCGRConfigFromPath(engineConfigPath)
	if err != nil {
		t.Error(err)
	}
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
	sentinelRPC = engine.NewRPCClient(t, sentinelConfig.ListenCfg())
}

func testRedisSentinelSetGetAttribute(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var result string
	if err := sentinelRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := sentinelRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testRedisSentinelInsertion(t *testing.T) {
	var nrFails1, nrFails2 int
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:~*reqAccount:1001"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	orgiginID := alsPrf.ID + "_"
	index := 0
	var result string
	addFunc := func(t *testing.T, nrFail *int) {
		alsPrf.ID = orgiginID + strconv.Itoa(index)
		if err := sentinelRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
			if err.Error() == "SERVER_ERROR: EOF" {
				*nrFail = *nrFail + 1
			} else {
				t.Error(err)
			}
		}
		index++
	}
	forFunc1 := func(t *testing.T) {
		for i := 0; i < 25; i++ {
			t.Run("add", func(t *testing.T) {
				t.Parallel()
				addFunc(t, &nrFails1)
			})
			if i == 5 {
				t.Run("stop1", func(t *testing.T) {
					t.Parallel()
					if err := node1Exec.Process.Kill(); err != nil {
						t.Error(err)
					}
					if err := stlExec1.Process.Kill(); err != nil {
						t.Error(err)
					}
				})
			}
			if i == 10 {
				t.Run("stop2", func(t *testing.T) {
					t.Parallel()
					if err := node2Exec.Process.Kill(); err != nil {
						t.Error(err)
					}
					if err := stlExec2.Process.Kill(); err != nil {
						t.Error(err)
					}
				})
			}
			t.Run("add2", func(t *testing.T) {
				t.Parallel()
				addFunc(t, &nrFails1)
			})
		}
	}
	forFunc2 := func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run("add", func(t *testing.T) {
				t.Parallel()
				addFunc(t, &nrFails2)
			})
			t.Run("add2", func(t *testing.T) {
				t.Parallel()
				addFunc(t, &nrFails2)
			})
		}
	}
	t.Run("for1", forFunc1)
	if nrFails1 == 0 {
		t.Error("Fail tests in case of failover")
	}
	node1Exec = exec.Command("redis-server", node1ConfigPath)
	if err := node1Exec.Start(); err != nil {
		t.Error(err)
	}
	node2Exec = exec.Command("redis-server", node2ConfigPath)
	if err := node2Exec.Start(); err != nil {
		t.Error(err)
	}
	t.Run("for2", forFunc2)
	if nrFails2 > 19 {
		t.Errorf("Fail tests in case of failback ")
	}
}

// After we kill node1 check the data if was replicated in node2
func testRedisSentinelGetAttrAfterFailover(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1001", utils.InfieldSep),
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var reply *engine.AttributeProfile
	if err := sentinelRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testRedisSentinelKillEngine(t *testing.T) {
	if err := exec.Command("pkill", "redis-server").Run(); err != nil {
		t.Error(err)
	}
	if err := exec.Command("pkill", "redis-sentinel").Run(); err != nil {
		t.Error(err)
	}
	if err := exec.Command("pkill", "redis-ser").Run(); err != nil {
		t.Error(err)
	}
	if err := exec.Command("pkill", "redis-sen").Run(); err != nil {
		t.Error(err)
	}

	if err := engine.KillEngine(2000); err != nil {
		t.Error(err)
	}
}
