//go:build integration || flaky || call || performance

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

package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/birpc/jsonrpc"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/ltcache"
	"github.com/creack/pty"
)

func InitDataDB(cfg *config.CGRConfig) error {
	d, err := NewDataDBConn(cfg.DataDbCfg().Type,
		cfg.DataDbCfg().Host, cfg.DataDbCfg().Port,
		cfg.DataDbCfg().Name, cfg.DataDbCfg().User,
		cfg.DataDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.DataDbCfg().Opts, cfg.DataDbCfg().Items)
	if err != nil {
		return err
	}
	dm := NewDataManager(d, cfg.CacheCfg(), connMgr)

	if err := dm.DataDB().Flush(""); err != nil {
		return err
	}
	//	Write version before starting
	if err := OverwriteDBVersions(dm.dataDB); err != nil {
		return err
	}
	return nil
}

func InitStorDB(cfg *config.CGRConfig) error {
	storDB, err := NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts, cfg.StorDbCfg().Items)
	if err != nil {
		return err
	}

	dbPath := strings.Trim(cfg.StorDbCfg().Type, "*")
	if err := storDB.Flush(path.Join(cfg.DataFolderPath, "storage",
		dbPath)); err != nil {
		return err
	}

	if slices.Contains([]string{utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres},
		cfg.StorDbCfg().Type) {
		if err := SetDBVersions(storDB); err != nil {
			return err
		}
	}
	return nil
}

func InitConfigDB(cfg *config.CGRConfig) error {
	d, err := NewDataDBConn(cfg.ConfigDBCfg().Type,
		cfg.ConfigDBCfg().Host, cfg.ConfigDBCfg().Port,
		cfg.ConfigDBCfg().Name, cfg.ConfigDBCfg().User,
		cfg.ConfigDBCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.ConfigDBCfg().Opts, nil)
	if err != nil {
		return err
	}
	return d.Flush("")
}

// Return reference towards the command started so we can stop it if necessary
func StartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	enginePath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	engine := exec.Command(enginePath, "-config_path", cfgPath)
	if err := engine.Start(); err != nil {
		return nil, err
	}
	cfg, err := config.NewCGRConfigFromPath(context.Background(), cfgPath)
	if err != nil {
		return nil, err
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	var connected bool
	for i := 0; i < 200; i++ {
		time.Sleep(fib())
		if _, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
			utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
				err.Error(), cfg.ListenCfg().RPCJSONListen))
		} else {
			connected = true
			break
		}
	}
	if !connected {
		return nil, fmt.Errorf("engine did not open port <%s>", cfg.ListenCfg().RPCJSONListen)
	}
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsystems
	return engine, nil
}

// StartEngineWithContext return reference towards the command started so we can stop it if necessary
func StartEngineWithContext(ctx context.Context, cfgPath string, waitEngine int) (engine *exec.Cmd, err error) {
	engine = exec.CommandContext(ctx, "cgr-engine", "-config_path", cfgPath)
	if err = engine.Start(); err != nil {
		return nil, err
	}
	var cfg *config.CGRConfig
	if cfg, err = config.NewCGRConfigFromPath(context.Background(), cfgPath); err != nil {
		return
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < 200; i++ {
		time.Sleep(fib())
		if _, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err != nil {
			continue
		}
		time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsystems
		return
	}
	utils.Logger.Warning(fmt.Sprintf("Error <%s> when opening test connection to: <%s>",
		err.Error(), cfg.ListenCfg().RPCJSONListen))
	err = fmt.Errorf("engine did not open port <%s>", cfg.ListenCfg().RPCJSONListen)
	return
}

func KillEngine(waitEngine int) error {
	return KillProcName("cgr-engine", waitEngine)
}

func StopStartEngine(cfgPath string, waitEngine int) (*exec.Cmd, error) {
	KillEngine(waitEngine)
	return StartEngine(cfgPath, waitEngine)
}

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disableReverse bool,
	cacheConns, schedConns []string) error {
	loader, err := NewTpReader(dm.dataDB, NewFileCSVStorage(utils.CSVSep, tpPath), "",
		timezone, cacheConns, schedConns, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.WriteToDatabase(false, disableReverse); err != nil {
		return utils.NewErrServerError(err)
	}
	return nil
}

type PjsuaAccount struct {
	ID, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.ID, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
	}
	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return nil, err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf) // Free the content since otherwise pjsua will not start
	time.Sleep(waitDur)     // Give time to rater to fire up
	return fPty, nil
}

func PjsuaCallURI(acnt *PjsuaAccount, dstURI, outboundURI string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundURI, "--id=" + acnt.ID, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstURI}

	pjsuaPath, err := exec.LookPath("pjsua")
	if err != nil {
		return err
	}
	pjsua := exec.Command(pjsuaPath, cmdArgs...)
	fPty, err := pty.Start(pjsua)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	io.Copy(os.Stdout, buf)
	go func() {
		time.Sleep(callDur + 2*time.Second)
		fPty.Write([]byte("q\n")) // Destroy the listener
	}()
	return nil
}

func KillProcName(procName string, waitMs int) (err error) {
	if err = exec.Command("pkill", procName).Run(); err != nil {
		return
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return
}

func ForceKillProcName(procName string, waitMs int) error {
	if err := exec.Command("pkill", "-9", procName).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond)
	return nil
}

func CallScript(scriptPath string, subcommand string, waitMs int) error {
	if err := exec.Command(scriptPath, subcommand).Run(); err != nil {
		return err
	}
	time.Sleep(time.Duration(waitMs) * time.Millisecond) // Give time to rater to fire up
	return nil
}

func GetDefaultEmptyCacheStats() map[string]*ltcache.CacheStats {
	return map[string]*ltcache.CacheStats{
		utils.MetaDefault:                      {},
		utils.CacheAttributeFilterIndexes:      {},
		utils.CacheAttributeProfiles:           {},
		utils.CacheChargerFilterIndexes:        {},
		utils.CacheChargerProfiles:             {},
		utils.CacheDispatcherFilterIndexes:     {},
		utils.CacheDispatcherProfiles:          {},
		utils.CacheDispatcherHosts:             {},
		utils.CacheDispatcherRoutes:            {},
		utils.CacheDispatcherLoads:             {},
		utils.CacheDispatchers:                 {},
		utils.CacheEventResources:              {},
		utils.CacheFilters:                     {},
		utils.CacheResourceFilterIndexes:       {},
		utils.CacheResourceProfiles:            {},
		utils.CacheResources:                   {},
		utils.CacheRPCResponses:                {},
		utils.CacheStatFilterIndexes:           {},
		utils.CacheStatQueueProfiles:           {},
		utils.CacheStatQueues:                  {},
		utils.CacheSTIR:                        {},
		utils.CacheRankingProfiles:             {},
		utils.CacheRouteFilterIndexes:          {},
		utils.CacheRouteProfiles:               {},
		utils.CacheThresholdFilterIndexes:      {},
		utils.CacheThresholdProfiles:           {},
		utils.CacheThresholds:                  {},
		utils.CacheRateProfiles:                {},
		utils.CacheRateProfilesFilterIndexes:   {},
		utils.CacheRateFilterIndexes:           {},
		utils.CacheDiameterMessages:            {},
		utils.CacheClosedSessions:              {},
		utils.CacheLoadIDs:                     {},
		utils.CacheRPCConnections:              {},
		utils.CacheCDRIDs:                      {},
		utils.CacheUCH:                         {},
		utils.CacheEventCharges:                {},
		utils.CacheReverseFilterIndexes:        {},
		utils.MetaAPIBan:                       {},
		utils.MetaSentryPeer:                   {},
		utils.CacheCapsEvents:                  {},
		utils.CacheActionProfiles:              {},
		utils.CacheActionProfilesFilterIndexes: {},
		utils.CacheAccounts:                    {},
		utils.CacheAccountsFilterIndexes:       {},
		utils.CacheReplicationHosts:            {},
	}
}

func NewRPCClient(cfg *config.ListenCfg, encoding string) (*birpc.Client, error) {
	switch encoding {
	case utils.MetaJSON:
		return jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		return birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		return nil, errors.New("invalid encoding")
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
	Encoding   string            // specifies the data encoding type (e.g., JSON, GOB)
}

// Setup initializes the testing environment using the provided configuration. It loads the configuration
// from a specified path or creates a new one if the path is not provided. The method starts the engine,
// establishes an RPC client connection, and loads CSV data if provided. It returns an RPC client and the
// configuration.
func (env TestEnvironment) Setup(t *testing.T, ctx *context.Context, engineDelay int) (*birpc.Client, *config.CGRConfig) {
	t.Helper()

	var cfg *config.CGRConfig
	switch {
	case env.ConfigPath != "":
		var err error
		cfg, err = config.NewCGRConfigFromPath(ctx, env.ConfigPath)
		if err != nil {
			t.Fatalf("failed to init config from path %s: %v", env.ConfigPath, err)
		}
	default:
		cfg, env.ConfigPath = initCfg(t, ctx, env.ConfigJSON)
	}

	flushDBs(t, cfg, true, true)
	startEngine(t, ctx, cfg, env.ConfigPath, engineDelay, env.LogBuffer)

	client, err := NewRPCClient(cfg.ListenCfg(), env.Encoding)
	if err != nil {
		t.Fatalf("could not connect to cgr-engine: %v", err)
	}

	loadCSVs(t, env.TpPath, env.TpFiles)

	return client, cfg
}

// initCfg creates a new CGRConfig from the provided configuration content string.
// It generates a temporary directory and file path, writes the content to the configuration
// file, and returns the created CGRConfig and the configuration directory path.
func initCfg(t *testing.T, ctx *context.Context, cfgContent string) (cfg *config.CGRConfig, cfgPath string) {
	t.Helper()
	if cfgContent == utils.EmptyString {
		t.Fatal("ConfigJSON is required but empty")
	}
	cfgPath = t.TempDir()
	filePath := filepath.Join(cfgPath, "cgrates.json")
	if err := os.WriteFile(filePath, []byte(cfgContent), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.NewCGRConfigFromPath(ctx, cfgPath)
	if err != nil {
		t.Fatalf("failed to init config from path %s: %v", cfgPath, err)
	}
	return cfg, cfgPath
}

// loadCSVs loads tariff plan data from CSV files. The CSV files are created based on the csvFiles map, where
// the key represents the file name and the value the contains its contents. Assumes the data is loaded
// automatically (RunDelay != 0)
func loadCSVs(t *testing.T, tpPath string, csvFiles map[string]string) {

	t.Helper()
	if tpPath != "" {
		for fileName, content := range csvFiles {
			filePath := path.Join(tpPath, fileName)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("could not write to file %s: %v", filePath, err)
			}
		}
	}
}

// flushDBs resets the databases specified in the configuration if the corresponding flags are true.
func flushDBs(t *testing.T, cfg *config.CGRConfig, flushDataDB, flushStorDB bool) {
	t.Helper()
	if flushDataDB {
		if err := InitDataDB(cfg); err != nil {
			t.Fatalf("failed to flush %s dataDB: %v", cfg.DataDbCfg().Type, err)
		}
	}
	if flushStorDB {
		if err := InitStorDB(cfg); err != nil {
			t.Fatalf("failed to flush %s storDB: %v", cfg.StorDbCfg().Type, err)
		}
	}
}

// startEngine starts the CGR engine process with the provided configuration. It writes engine logs to the provided
// logBuffer (if any) and waits for the engine to be ready. If the passed context were to be cancelled, the engine
// would also shut down.
func startEngine(t *testing.T, ctx *context.Context, cfg *config.CGRConfig, cfgPath string, waitEngine int,
	logBuffer io.Writer) {
	t.Helper()
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal("could not find cgr-engine executable")
	}
	engine := exec.CommandContext(
		ctx,
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
