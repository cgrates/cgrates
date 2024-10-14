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
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

/*
 * Documentation:
 * This code should work on redis 5 or later:
 *		`redis-cli --cluster create 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 127.0.0.1:7006 --cluster-replicas 1`
 * For redis 4 or before you need to create create the cluster manualy:
 *     	- install ruby
 * 	   	- install redis gem: `gem install redis`
 *     	- download the `redis-trib.rb` from the source code
 *     	- start the 6 nodes with the command `redis-server node1.conf`
 *     	- configure the cluster with the following command:
 *         	`./redis-trib.rb create --replicas 1 127.0.0.1:7001 127.0.0.1:7002 127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 127.0.0.1:7006`
 *
 * To run the tests you need to specify the `redisCluster` flag and have the redis stopped:
 *    	`go test github.com/cgrates/cgrates/general_tests -tags=integration -dbtype=*mysql -run=TestRedisCluster -redisCluster  -v`
 *
 * The configuration of the cluster is the following:
 *		- node1 127.0.0.1:7001 master
 *		- node2 127.0.0.1:7002 master
 *		- node3 127.0.0.1:7003 master
 * 		- node4 127.0.0.1:7004 replica
 *		- node5 127.0.0.1:7005 replica
 * 		- node6 127.0.0.1:7006 replica
 * The replicas do not allways select the same master
 */

var (
	clsrConfig *config.CGRConfig
	clsrRPC    *birpc.Client

	clsrNodeCfgPath   = path.Join(*utils.DataDir, "redisCluster", "node%v.conf")
	clsrEngineCfgPath = path.Join(*utils.DataDir, "conf", "samples", "redisCluster")
	clsrNodes         = make(map[string]*exec.Cmd)
	clsrOutput        = make(map[string]*bytes.Buffer) // in order to debug if something is not working
	clsrNoNodes       = 6                              // this is the minimum number of nodes for a cluster with 1 replica for each master
	clsrRedisFlag     = flag.Bool("redisCluster", false, "Run tests for redis cluster")
	clsrTests         = []func(t *testing.T){
		testClsrPrepare,
		testClsrStartNodes,
		testClsrCreateCluster,
		testClsrInitConfig,
		testClsrFlushDb,
		testClsrStartEngine,
		testClsrRPCConection,
		testClsrSetGetAttribute,
		testClsrStopMaster,
		testClsrSetGetAttribute2,
		testClsrReStartMaster,
		testClsrGetAttribute,
		testClsrStopNodes,
		testClsrKillEngine,
		testClsrDeleteFolder,
		// testClsrPrintOutput,
	}

	clsrRedisCliArgs = []string{
		"--cluster", "create",
		"127.0.0.1:7001",
		"127.0.0.1:7002",
		"127.0.0.1:7003",
		"127.0.0.1:7004",
		"127.0.0.1:7005",
		"127.0.0.1:7006",
		"--cluster-replicas", "1",
	}
)

const (
	clsrRedisCmd    = "redis-server"
	clsrRedisCliCmd = "redis-cli"
	clsrDir         = "/tmp/cluster/"
)

func TestRedisCluster(t *testing.T) {
	if !*clsrRedisFlag {
		t.SkipNow()
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
	for _, stest := range clsrTests {
		t.Run("TestRedisCluster", stest)
	}
}

func testClsrPrepare(t *testing.T) {
	if err := os.MkdirAll(clsrDir, 0755); err != nil {
		t.Fatalf("Error creating folder<%s>:%s", clsrDir, err)
	}
}

func testClsrStartNodes(t *testing.T) {
	for i := 1; i <= clsrNoNodes; i++ {
		path := fmt.Sprintf(clsrNodeCfgPath, i)
		clsrNodes[path] = exec.Command(clsrRedisCmd, path)
		clsrOutput[path] = bytes.NewBuffer(nil)
		clsrNodes[path].Stdout = clsrOutput[path]
		if err := clsrNodes[path].Start(); err != nil {
			t.Fatalf("Could not start node %v because %s", i, err)
		}
	}
}

func testClsrCreateCluster(t *testing.T) {
	cmd := exec.Command(clsrRedisCliCmd, clsrRedisCliArgs...)
	cmd.Stdin = bytes.NewBuffer([]byte("yes\n"))
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	if err := cmd.Run(); err != nil {
		t.Errorf("Could not create the cluster because %s", err)
		t.Logf("The output was:\n %s", stdOut.String()) // print the output to debug the error
	}
	time.Sleep(200 * time.Millisecond)
}

func testClsrInitConfig(t *testing.T) {
	var err error
	clsrConfig, err = config.NewCGRConfigFromPath(clsrEngineCfgPath)
	if err != nil {
		t.Error(err)
	}
	clsrConfig.DataFolderPath = *utils.DataDir // Share DataFolderPath through config towards StoreDb for Flush()
}

func testClsrFlushDb(t *testing.T) {
	if err := engine.InitDataDb(clsrConfig); err != nil {
		t.Fatal(err)
	}
}

func testClsrStartEngine(t *testing.T) {
	if _, err := engine.StopStartEngine(clsrEngineCfgPath, 200); err != nil {
		t.Fatal(err)
	}
}

func testClsrRPCConection(t *testing.T) {
	clsrRPC = engine.NewRPCClient(t, clsrConfig.ListenCfg())
}

func testClsrSetGetAttribute(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ClsrTest",
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
	if err := clsrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := clsrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ClsrTest"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testClsrStopMaster(t *testing.T) {
	path := fmt.Sprintf(clsrNodeCfgPath, 3)
	if err := clsrNodes[path].Process.Kill(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)
}

func testClsrSetGetAttribute2(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ClsrTest",
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
	if err := clsrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ClsrTest"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
	// add another attribute
	alsPrf.ID += "2"
	var result string
	if err := clsrRPC.Call(context.Background(), utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
}

func testClsrReStartMaster(t *testing.T) {
	path := fmt.Sprintf(clsrNodeCfgPath, 3)
	clsrNodes[path] = exec.Command(clsrRedisCmd, path)
	clsrOutput[path] = bytes.NewBuffer(nil)
	clsrNodes[path].Stdout = clsrOutput[path]
	if err := clsrNodes[path].Start(); err != nil {
		t.Fatalf("Could not start node %v because %s", 3, err)
	}
	time.Sleep(200 * time.Millisecond)
}

func testClsrGetAttribute(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ClsrTest2",
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
	if err := clsrRPC.Call(context.Background(), utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ClsrTest2"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testClsrStopNodes(t *testing.T) {
	for path, node := range clsrNodes {
		if err := node.Process.Kill(); err != nil {
			t.Fatalf("Could not stop node with path <%s> because %s", path, err)
		}
	}
}

func testClsrPrintOutput(t *testing.T) {
	for path, node := range clsrOutput {
		t.Logf("The output of the node <%s> is:\n%s", path, node.String())
		t.Logf("==========================================================")
	}
}

func testClsrKillEngine(t *testing.T) {
	if err := engine.KillEngine(200); err != nil {
		t.Error(err)
	}
}

func testClsrDeleteFolder(t *testing.T) {
	if err := os.RemoveAll(clsrDir); err != nil {
		t.Fatalf("Error removing folder<%s>: %s", clsrDir, err)
	}
}
