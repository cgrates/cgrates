//go:build integration
// +build integration

// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	if err := engine.InitDataDB(redisTLSCfg); err != nil {
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
