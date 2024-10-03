//go:build integration || flaky || kafka

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
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

var (
	err error
)

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

// TestEngine holds the setup parameters and configurations
// required for running integration tests.
type TestEngine struct {
	ConfigPath     string            // path to the main configuration file
	ConfigJSON     string            // configuration JSON content (used if ConfigPath is empty)
	LogBuffer      io.Writer         // captures log output of the test environment
	PreserveDataDB bool              // prevents automatic data_db flush when set
	PreserveStorDB bool              // prevents automatic stor_db flush when set
	TpPath         string            // path to the tariff plans
	TpFiles        map[string]string // CSV data for tariff plans: filename -> content

	// PreStartHook executes custom logic relying on CGRConfig
	// before starting cgr-engine.
	PreStartHook func(*testing.T, *config.CGRConfig)
}

// Run initializes a cgr-engine instance for testing, loads tariff plans (if available) and returns
// an RPC client and the CGRConfig object. It calls t.Fatal on any setup failure.
func (env TestEngine) Run(t *testing.T) (*birpc.Client, *config.CGRConfig) {
	t.Helper()

	// Parse config files.
	var cfgPath string
	switch {
	case env.ConfigJSON != "":
		cfgPath = t.TempDir()
		filePath := filepath.Join(cfgPath, "cgrates.json")
		if err := os.WriteFile(filePath, []byte(env.ConfigJSON), 0644); err != nil {
			t.Fatal(err)
		}
	case env.ConfigPath != "":
		cfgPath = env.ConfigPath
	default:
		t.Fatal("missing config source")
	}
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("could not init config from path %s: %v", cfgPath, err)
	}

	flushDBs(t, cfg, !env.PreserveDataDB, !env.PreserveStorDB)

	if env.PreStartHook != nil {
		env.PreStartHook(t, cfg)
	}

	startEngine(t, cfg, env.LogBuffer)

	client, err := newRPCClient(cfg.ListenCfg())
	if err != nil {
		t.Fatalf("could not connect to cgr-engine: %v", err)
	}

	var customTpPath string
	if len(env.TpFiles) != 0 {
		customTpPath = t.TempDir()
	}
	loadCSVs(t, client, env.TpPath, customTpPath, env.TpFiles)

	return client, cfg
}

func waitForService(t *testing.T, ctx *context.Context, client *birpc.Client, service string) {
	t.Helper()
	method := service + ".Ping"
	backoff := utils.FibDuration(time.Millisecond, 0)
	var reply any
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("%s service did not become available: %v", service, ctx.Err())
		default:
			err := client.Call(context.Background(), method, nil, &reply)
			if err == nil && reply == utils.Pong {
				return
			}
			time.Sleep(backoff())
		}
	}
}

// loadCSVs loads tariff plan data from CSV files into the service. It handles directory creation and file
// writing for custom paths, and loads data from the specified paths using the provided RPC client.
func loadCSVs(t *testing.T, client *birpc.Client, tpPath, customTpPath string, csvFiles map[string]string) {
	t.Helper()
	paths := make([]string, 0, 2)
	if customTpPath != "" {
		for fileName, content := range csvFiles {
			filePath := path.Join(customTpPath, fileName)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("could not write to file %s: %v", filePath, err)
			}
		}
		paths = append(paths, customTpPath)
	}
	if tpPath != "" {
		paths = append(paths, tpPath)
	}
	if len(paths) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	waitForService(t, ctx, client, utils.APIerSv1)

	var reply string
	for _, path := range paths {
		args := &utils.AttrLoadTpFromFolder{FolderPath: path}
		err := client.Call(context.Background(), utils.APIerSv1LoadTariffPlanFromFolder, args, &reply)
		if err != nil {
			t.Fatalf("%s call failed for path %s: %v", utils.APIerSv1LoadTariffPlanFromFolder, path, err)
		}
	}
}

// flushDBs resets the databases specified in the configuration if the corresponding flags are true.
func flushDBs(t *testing.T, cfg *config.CGRConfig, flushDataDB, flushStorDB bool) {
	t.Helper()
	if flushDataDB {
		if err := engine.InitDataDb(cfg); err != nil {
			t.Fatalf("failed to flush %s dataDB: %v", cfg.DataDbCfg().Type, err)
		}
	}
	if flushStorDB {
		if err := engine.InitStorDb(cfg); err != nil {
			t.Fatalf("failed to flush %s storDB: %v", cfg.StorDbCfg().Type, err)
		}
	}
}

// startEngine starts the CGR engine process with the provided configuration. It writes engine logs to the
// provided logBuffer (if any).
func startEngine(t *testing.T, cfg *config.CGRConfig, logBuffer io.Writer) {
	t.Helper()
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal(err)
	}
	engine := exec.Command(
		binPath,
		"-config_path", cfg.ConfigPath,
		"-logger", utils.MetaStdLog,
	)
	if logBuffer != nil {
		engine.Stdout = logBuffer
		engine.Stderr = logBuffer
	}
	if err := engine.Start(); err != nil {
		t.Fatalf("cgr-engine command failed: %v", err)
	}
	t.Cleanup(func() {
		if err := engine.Process.Kill(); err != nil {
			t.Errorf("failed to kill cgr-engine process (%d): %v", engine.Process.Pid, err)
		}
	})
	fib := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < 16; i++ {
		time.Sleep(fib())
		if _, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("starting cgr-engine on port %s failed: %v", cfg.ListenCfg().RPCJSONListen, err)
	}
}
