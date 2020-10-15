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
	"net/rpc"
	"os/exec"
	"path"
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	redisTLS          = flag.Bool("redis_tls", false, "Run tests with redis tls")
	redisTLSServer    *exec.Cmd
	redisTLSEngineCfg = path.Join(*dataDir, "conf", "samples", "tutredis_tls")
	redisTLSCfg       *config.CGRConfig
	redisTLSRPC       *rpc.Client

	sTestsRedisTLS = []func(t *testing.T){
		testRedisTLSStartServer,
		testRedisTLSInitConfig,
		testRedisTLSFlushDb,
		testRedisTLSStartEngine,
		testRedisTLSRPCCon,
		testRedisTLSSetGetAttribute,
		testRedisTLSKillEngine,
	}
)

// Before running these tests first you need to make sure you build the redis server with TLS support
// https://redis.io/topics/encryption
func TestRedisTLS(t *testing.T) {
	if !*redisTLS {
		return
	}
	switch *dbType {
	case utils.MetaMySQL:
	case utils.MetaInternal,
		utils.MetaMongo,
		utils.MetaPostgres:
		t.SkipNow()
	default:
		t.Fatal("Unknown Database type")
	}
	for _, stest := range sTestsRedisTLS {
		t.Run("TestRedisTLS", stest)
	}
}

func testRedisTLSStartServer(t *testing.T) {
	// start the server with the server.crt server.key and ca.crt from /data/tls ( self sign certificate )
	args := []string{
		"--tls-port", "6400", "--port", "0", "--tls-cert-file", "/usr/share/cgrates/tls/server.crt",
		"--tls-key-file", "/usr/share/cgrates/tls/server.key", "--tls-ca-cert-file", "/usr/share/cgrates/tls/ca.crt",
	}
	redisTLSServer = exec.Command("redis-server", args...)
	if err := redisTLSServer.Start(); err != nil {
		t.Error(err)
	}
}

func testRedisTLSInitConfig(t *testing.T) {
	var err error
	redisTLSCfg, err = config.NewCGRConfigFromPath(redisTLSEngineCfg)
	if err != nil {
		t.Error(err)
	}
}

func testRedisTLSFlushDb(t *testing.T) {
	if err := engine.InitDataDb(redisTLSCfg); err != nil {
		t.Fatal(err)
	}
}

func testRedisTLSStartEngine(t *testing.T) {
	// for the engine we will use the client.crt client.key and ca.crt
	if _, err := engine.StopStartEngine(redisTLSEngineCfg, 2000); err != nil {
		t.Fatal(err)
	}
}

func testRedisTLSRPCCon(t *testing.T) {
	var err error
	redisTLSRPC, err = newRPCClient(redisTLSCfg.ListenCfg()) // We connect over JSON so we can also troubleshoot if needed
	if err != nil {
		t.Fatal(err)
	}
}

func testRedisTLSSetGetAttribute(t *testing.T) {
	alsPrf := &engine.AttributeProfile{
		Tenant:    "cgrates.org",
		ID:        "ApierTest",
		Contexts:  []string{utils.MetaSessionS, utils.MetaCDRs},
		FilterIDs: []string{"*string:~*req.Account:1001"},
		Attributes: []*engine.Attribute{
			{
				Path:  utils.MetaReq + utils.NestingSep + utils.Subject,
				Value: config.NewRSRParsersMustCompile("1001", utils.INFIELD_SEP),
			},
		},
		Weight: 20,
	}
	alsPrf.Compile()
	var result string
	if err := redisTLSRPC.Call(utils.APIerSv1SetAttributeProfile, alsPrf, &result); err != nil {
		t.Error(err)
	} else if result != utils.OK {
		t.Error("Unexpected reply returned", result)
	}
	var reply *engine.AttributeProfile
	if err := redisTLSRPC.Call(utils.APIerSv1GetAttributeProfile,
		&utils.TenantID{Tenant: "cgrates.org", ID: "ApierTest"}, &reply); err != nil {
		t.Fatal(err)
	}
	reply.Compile()
	if !reflect.DeepEqual(alsPrf, reply) {
		t.Errorf("Expecting : %+v, received: %+v", alsPrf, reply)
	}
}

func testRedisTLSKillEngine(t *testing.T) {
	if err := exec.Command("pkill", "redis-server").Run(); err != nil {
		t.Error(err)
	}

	if err := engine.KillEngine(2000); err != nil {
		t.Error(err)
	}
}
