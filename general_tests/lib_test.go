//go:build integration || flaky

/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/
package general_tests

import (
	"errors"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func newRPCClient(cfg *config.ListenCfg) (c *rpc.Client, err error) {
	switch *utils.Encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return rpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

// TestEnvironment holds the setup parameters and configurations
// required for running integration tests.
type TestEnvironment struct {
	ConfigPath     string            // file path to the main configuration file
	ConfigJSON     string            // contains the configuration JSON content if ConfigPath is missing
	TpPath         string            // specifies the path to the tariff plans
	TpFiles        map[string]string // maps CSV filenames to their content for tariff plan loading
	PreserveDataDB bool              // prevents automatic data_db flush when set
	PreserveStorDB bool              // prevents automatic stor_db flush when set
	LogBuffer      io.Writer         // captures the log output of the test environment
}

// Setup initializes the testing environment using the provided configuration. It loads the configuration
// from a specified path or creates a new one if the path is not provided. The method starts the engine,
// establishes an RPC client connection, and loads CSV data if provided. It returns an RPC client and the
// configuration.
func (env TestEnvironment) Setup(t *testing.T, engineDelay int) (*rpc.Client, *config.CGRConfig) {
	t.Helper()

	var cfg *config.CGRConfig
	switch {
	case env.ConfigPath != "":
		var err error
		cfg, err = config.NewCGRConfigFromPath(env.ConfigPath)
		if err != nil {
			t.Fatalf("failed to init config from path %s: %v", env.ConfigPath, err)
		}
	default:
		cfg, env.ConfigPath = initCfg(t, env.ConfigJSON)
	}

	flushDBs(t, cfg, !env.PreserveDataDB, !env.PreserveStorDB)
	startEngine(t, cfg, env.ConfigPath, engineDelay, env.LogBuffer)

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

// initCfg creates a new CGRConfig from the provided configuration content string.
// It generates a temporary directory and file path, writes the content to the configuration
// file, and returns the created CGRConfig and the configuration directory path.
func initCfg(t *testing.T, cfgContent string) (cfg *config.CGRConfig, cfgPath string) {
	t.Helper()
	if cfgContent == utils.EmptyString {
		t.Fatal("ConfigJSON is required but empty")
	}
	cfgPath = t.TempDir()
	filePath := filepath.Join(cfgPath, "cgrates.json")
	if err := os.WriteFile(filePath, []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("failed to init config from path %s: %v", cfgPath, err)
	}
	return cfg, cfgPath
}

// loadCSVs loads tariff plan data from CSV files into the service. It handles directory creation and file
// writing for custom paths, and loads data from the specified paths using the provided RPC client.
func loadCSVs(t *testing.T, client *rpc.Client, tpPath, customTpPath string, csvFiles map[string]string) {
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

	if tpPath != utils.EmptyString {
		paths = append(paths, tpPath)
	}

	var reply string
	for _, path := range paths {
		args := &utils.AttrLoadTpFromFolder{FolderPath: path}
		err := client.Call(utils.APIerSv1LoadTariffPlanFromFolder, args, &reply)
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
			t.Fatalf("failed to flush %s dataDB: %v", cfg.DataDbCfg().DataDbType, err)
		}
	}
	if flushStorDB {
		if err := engine.InitStorDb(cfg); err != nil {
			t.Fatalf("failed to flush %s storDB: %v", cfg.StorDbCfg().Type, err)
		}
	}
}

// startEngine starts the CGR engine process with the provided configuration. It writes engine logs to the
// provided logBuffer (if any) and waits for the engine to be ready.
func startEngine(t *testing.T, cfg *config.CGRConfig, cfgPath string, waitEngine int, logBuffer io.Writer) {
	t.Helper()
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("could not find cgr-engine executable")
	}
	engine := exec.Command(
		binPath,
		"-config_path", cfgPath,
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
			t.Logf("failed to kill cgr-engine process (%d): %v", engine.Process.Pid, err)
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
	time.Sleep(time.Duration(waitEngine) * time.Millisecond)
}
