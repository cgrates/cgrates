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
	"flag"
	"fmt"
	"io"
	"math/rand"
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
	dataDir   = flag.String("data_dir", "/usr/share/cgrates", "CGR data dir path here")
	waitRater = flag.Int("wait_rater", 500, "Number of milliseconds to wait for rater to start and cache")
	encoding  = flag.String("rpc", utils.MetaJSON, "what encoding would be used for rpc communication")
	dbType    = flag.String("dbtype", utils.MetaInternal, "The type of DataBase (Internal/Mongo/mySql)")
	err       error
)

func newRPCClient(cfg *config.ListenCfg) (c *birpc.Client, err error) {
	switch *encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("UNSUPPORTED_RPC")
	}
}

// TestEnvironment holds the setup parameters and configurations
// required for running integration tests.
type TestEnvironment struct {
	Name       string            // usually the name of the test
	ConfigPath string            // file path to the main configuration file
	ConfigJSON string            // contains the configuration JSON content if ConfigPath is missing
	TpPath     string            // specifies the path to the tariff plans
	TpFiles    map[string]string // maps CSV filenames to their content for tariff plan loading
	LogBuffer  io.Writer         // captures the log output of the test environment
	// Encoding   string         // specifies the data encoding type (e.g., JSON, GOB)
}

// Setup initializes the testing environment using the provided configuration. It loads the configuration
// from a specified path or creates a new one if the path is not provided. The method starts the engine,
// establishes an RPC client connection, and loads CSV data if provided. It returns an RPC client, the
// configuration, a shutdown function, and any error encountered.
func (env TestEnvironment) Setup(t *testing.T, engineDelay int,
) (client *birpc.Client, cfg *config.CGRConfig, shutdownFunc context.CancelFunc, err error) {

	switch {
	case env.ConfigPath != "":
		cfg, err = config.NewCGRConfigFromPath(env.ConfigPath)
	default:
		var clean func()
		cfg, env.ConfigPath, clean, err = initCfg(env.ConfigJSON)
		defer clean() // it is safe to defer clean func before error check
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	if err = flushDBs(cfg, true, true); err != nil {
		return nil, nil, nil, err
	}

	exec.Command("pkill", "cgr-engine").Run()
	time.Sleep(time.Duration(engineDelay) * time.Millisecond)

	cancel, err := startEngine(cfg, env.ConfigPath, engineDelay, env.LogBuffer)
	if err != nil {
		return nil, nil, nil, err
	}

	client, err = newRPCClient(cfg.ListenCfg())
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("could not connect to cgr-engine: %w", err)
	}

	var customTpPath string
	if len(env.TpFiles) != 0 {
		customTpPath = fmt.Sprintf("/tmp/testTPs/%s", env.Name)
	}

	if err := loadCSVs(client, env.TpPath, customTpPath, env.TpFiles); err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("failed to load csvs: %w", err)
	}

	return client, cfg, cancel, nil
}

// initCfg creates a new CGRConfig from the provided configuration content string. It generates a
// temporary directory and file path, writes the content to the configuration file, and returns the
// created CGRConfig, the file path, a cleanup function, and any error encountered.
func initCfg(cfgContent string) (cfg *config.CGRConfig, cfgPath string, cleanFunc func(), err error) {
	if cfgContent == utils.EmptyString {
		return nil, "", func() {}, errors.New("content should not be empty")
	}
	cfgPath = fmt.Sprintf("/tmp/config%d", rand.Int63n(10000))
	err = os.MkdirAll(cfgPath, 0755)
	if err != nil {
		return nil, "", func() {}, err
	}
	removeFunc := func() {
		os.RemoveAll(cfgPath)
	}
	filePath := filepath.Join(cfgPath, "cgrates.json")
	err = os.WriteFile(filePath, []byte(cfgContent), 0644)
	if err != nil {
		return nil, "", removeFunc, err
	}
	cfg, err = config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		return nil, "", removeFunc, err
	}

	return cfg, cfgPath, removeFunc, nil
}

// loadCSVs loads tariff plan data from CSV files into the service. It handles directory creation and file writing for custom
// paths, and loads data from the specified paths using the provided RPC client. Returns an error if any step fails.
func loadCSVs(client *birpc.Client, tpPath, customTpPath string, csvFiles map[string]string) (err error) {
	paths := make([]string, 0, 2)
	if customTpPath != "" {
		err = os.MkdirAll(customTpPath, 0755)
		if err != nil {
			return fmt.Errorf("could not create folder %s: %w", customTpPath, err)
		}
		defer func() {
			rmErr := os.RemoveAll(customTpPath)
			if rmErr != nil {
				err = errors.Join(
					err,
					fmt.Errorf("could not remove folder %s: %w", customTpPath, rmErr))
			}
		}()
		for fileName, content := range csvFiles {
			filePath := path.Join(customTpPath, fileName)
			err = os.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				return fmt.Errorf("could not write to file %s: %w", filePath, err)
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
		err = client.Call(context.Background(),
			utils.APIerSv1LoadTariffPlanFromFolder,
			args, &reply)
		if err != nil {
			return fmt.Errorf("%s call failed for path %s: %w", utils.APIerSv1LoadTariffPlanFromFolder, path, err)
		}
	}
	return nil
}

// flushDBs resets the databases specified in the configuration if the corresponding flags are true.
// Returns an error if flushing either of the databases fails.
func flushDBs(cfg *config.CGRConfig, flushDataDB, flushStorDB bool) error {
	if flushDataDB {
		if err := engine.InitDataDb(cfg); err != nil {
			return fmt.Errorf("failed to flush %s dataDB: %w", cfg.DataDbCfg().Type, err)
		}
	}
	if flushStorDB {
		if err := engine.InitStorDb(cfg); err != nil {
			return fmt.Errorf("failed to flush %s storDB: %w", cfg.StorDbCfg().Type, err)
		}
	}
	return nil
}

// startEngine starts the CGR engine process with the provided configuration. It writes engine logs to the provided logBuffer
// (if any) and waits for the engine to be ready. Returns a cancel function to stop the engine and any error encountered.
func startEngine(cfg *config.CGRConfig, cfgPath string, waitEngine int, logBuffer io.Writer) (context.CancelFunc, error) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	cancel := func() {
		exec.Command("pkill", "cgr-engine").Run()
	}
	engine := exec.Command(enginePath, "-config_path", cfgPath)
	if logBuffer != nil {
		engine.Stdout = logBuffer
		engine.Stderr = logBuffer
	}
	if err := engine.Start(); err != nil {
		return nil, err
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < 20; i++ {
		time.Sleep(fib())
		_, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("starting cgr-engine on port %s failed: %w", cfg.ListenCfg().RPCJSONListen, err)
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond)
	return cancel, nil
}
