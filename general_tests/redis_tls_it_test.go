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
	"os/exec"
	"path"
	"testing"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
)

var (
	redisTLS          = flag.Bool("redisTLS", false, "Run tests with redis tls")
	redisTLSServer    *exec.Cmd
	redisTLSEngineCfg = path.Join(*utils.DataDir, "conf", "samples", "redisTLS")
	redisTLSCfg       *config.CGRConfig
	redisTLSRPC       *birpc.Client

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
	redisTLSRPC = engine.NewRPCClient(t, redisTLSCfg.ListenCfg())
}

func testRedisTLSSetGetAttribute(t *testing.T) {
	// status command to check if the engine starts
	var rply map[string]any
	if err := redisTLSRPC.Call(context.Background(), utils.CoreSv1Status, &utils.TenantWithAPIOpts{}, &rply); err != nil {
		t.Error(err)
	}
}

func testRedisTLSKillEngine(t *testing.T) {
	if err := engine.KillEngine(2000); err != nil {
		t.Error(err)
	}
	if err := exec.Command("pkill", "redis-server").Run(); err != nil {
		t.Error(err)
	}
}
