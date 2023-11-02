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
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"testing"

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

// setupTest prepares the testing environment. It takes optional file paths for
// existing configuration files or tariff plans and a content string for generating a new configuration
// if no path is provided. It also takes a map of CSV filenames to content strings for loading data.
// Returns an RPC client to interact with the engine, the configuration, a shutdown function to close
// the engine, and an error if any step of the initialization fails.
//
// If cfgPath is provided, it loads configuration from the specified file; otherwise, it creates a new
// configuration with the content provided. If tpPath is given, it uses the path for CSV loading; if it's
// empty but csvFiles is not, it creates a temporary directory with CSV files for loading into the service.
func setupTest(t *testing.T, testName, cfgPath, tpPath, content string, csvFiles map[string]string,
) (client *birpc.Client, cfg *config.CGRConfig, shutdownFunc func(), err error) {
	switch {
	case cfgPath != "":
		cfg, err = config.NewCGRConfigFromPath(cfgPath)
	default:
		var clean func()
		cfg, cfgPath, clean, err = initTestCfg(content)
		defer clean() // it is safe to defer clean func before error check
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	if err = flushDBs(cfg, true, true); err != nil {
		return nil, nil, nil, err
	}

	_, err = engine.StopStartEngine(cfgPath, *waitRater)
	if err != nil {
		return nil, nil, nil, err
	}

	shutdownFunc = func() {
		err := engine.KillEngine(*waitRater)
		if err != nil {
			t.Log(err)
		}
	}

	client, err = newRPCClient(cfg.ListenCfg())
	if err != nil {
		shutdownFunc()
		return nil, nil, nil, fmt.Errorf("could not connect to cgr-engine: %w", err)
	}

	var customTpPath string
	if len(csvFiles) != 0 {
		customTpPath = fmt.Sprintf("/tmp/testTPs/%s", testName)
	}

	if err := loadCSVs(client, tpPath, customTpPath, csvFiles); err != nil {
		shutdownFunc()
		return nil, nil, nil, fmt.Errorf("failed to load csvs: %w", err)
	}

	return client, cfg, shutdownFunc, nil
}

// initTestCfg creates a new CGRConfig from the provided configuration content string.
// It generates a temporary file path, writes the content to a configuration file,
// and returns the created CGRConfig, the path to the configuration file,
// a cleanup function to remove the temporary configuration file,
// and an error if the content is empty or an issue occurs during file creation or
// configuration initialization.
func initTestCfg(cfgContent string) (cfg *config.CGRConfig, cfgPath string, cleanFunc func(), err error) {
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

// loadCSVs loads tariff plan data from specified CSV files by calling the 'APIerSv1.LoadTariffPlanFromFolder' method using
// the client parameter.
// It handles the creation of a custom temporary path if provided and ensures the data from the given CSV files
// is written and loaded as well. If no custom path is provided, it will load CSVs from the tpPath if it is not empty.
// Returns an error if directory creation, file writing, or data loading fails.
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
