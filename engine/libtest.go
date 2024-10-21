//go:build integration || flaky || offline || kafka || call || race || performance

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
	"encoding/json"
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

func InitDataDb(cfg *config.CGRConfig) error {
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

func InitStorDb(cfg *config.CGRConfig) error {
	storDb, err := NewStorDBConn(cfg.StorDbCfg().Type,
		cfg.StorDbCfg().Host, cfg.StorDbCfg().Port,
		cfg.StorDbCfg().Name, cfg.StorDbCfg().User,
		cfg.StorDbCfg().Password, cfg.GeneralCfg().DBDataEncoding,
		cfg.StorDbCfg().StringIndexedFields, cfg.StorDbCfg().PrefixIndexedFields,
		cfg.StorDbCfg().Opts, cfg.StorDbCfg().Items)
	if err != nil {
		return err
	}
	dbPath := strings.Trim(cfg.StorDbCfg().Type, "*")
	if err := storDb.Flush(path.Join(cfg.DataFolderPath, "storage",
		dbPath)); err != nil {
		return err
	}
	if slices.Contains([]string{utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres},
		cfg.StorDbCfg().Type) {
		if err := SetDBVersions(storDb); err != nil {
			return err
		}
	}
	return nil
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
	cfg, err := config.NewCGRConfigFromPath(cfgPath)
	if err != nil {
		return nil, err
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	var connected bool
	for i := 0; i < 16; i++ {
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
	if cfg, err = config.NewCGRConfigFromPath(cfgPath); err != nil {
		return
	}
	fib := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < 16; i++ {
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

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disable_reverse bool,
	cacheConns, schedConns []string) error {
	csvStorage, err := NewFileCSVStorage(utils.CSVSep, tpPath)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	loader, err := NewTpReader(dm.dataDB, csvStorage, "",
		timezone, cacheConns, schedConns, false)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.LoadAll(); err != nil {
		return utils.NewErrServerError(err)
	}
	if err := loader.WriteToDatabase(false, disable_reverse); err != nil {
		return utils.NewErrServerError(err)
	}
	return nil
}

type PjsuaAccount struct {
	Id, Username, Password, Realm, Registrar string
}

// Returns file reference where we can write to control pjsua in terminal
func StartPjsuaListener(acnts []*PjsuaAccount, localPort, waitDur time.Duration) (*os.File, error) {
	cmdArgs := []string{fmt.Sprintf("--local-port=%d", localPort), "--null-audio", "--auto-answer=200", "--max-calls=32", "--app-log-level=0"}
	for idx, acnt := range acnts {
		if idx != 0 {
			cmdArgs = append(cmdArgs, "--next-account")
		}
		cmdArgs = append(cmdArgs, "--id="+acnt.Id, "--registrar="+acnt.Registrar, "--username="+acnt.Username, "--password="+acnt.Password, "--realm="+acnt.Realm)
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

func PjsuaCallUri(acnt *PjsuaAccount, dstUri, outboundUri string, callDur time.Duration, localPort int) error {
	cmdArgs := []string{"--null-audio", "--app-log-level=0", fmt.Sprintf("--local-port=%d", localPort), fmt.Sprintf("--duration=%d", int(callDur.Seconds())),
		"--outbound=" + outboundUri, "--id=" + acnt.Id, "--username=" + acnt.Username, "--password=" + acnt.Password, "--realm=" + acnt.Realm, dstUri}

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
		utils.MetaDefault:                  {},
		utils.CacheAccountActionPlans:      {},
		utils.CacheActionPlans:             {},
		utils.CacheActionTriggers:          {},
		utils.CacheActions:                 {},
		utils.CacheAttributeFilterIndexes:  {},
		utils.CacheAttributeProfiles:       {},
		utils.CacheChargerFilterIndexes:    {},
		utils.CacheChargerProfiles:         {},
		utils.CacheDispatcherFilterIndexes: {},
		utils.CacheDispatcherProfiles:      {},
		utils.CacheDispatcherHosts:         {},
		utils.CacheDispatcherRoutes:        {},
		utils.CacheDispatcherLoads:         {},
		utils.CacheDispatchers:             {},
		utils.CacheDestinations:            {},
		utils.CacheEventResources:          {},
		utils.CacheFilters:                 {},
		utils.CacheRatingPlans:             {},
		utils.CacheRatingProfiles:          {},
		utils.CacheResourceFilterIndexes:   {},
		utils.CacheResourceProfiles:        {},
		utils.CacheResources:               {},
		utils.CacheReverseDestinations:     {},
		utils.CacheRPCResponses:            {},
		utils.CacheSharedGroups:            {},
		utils.CacheStatFilterIndexes:       {},
		utils.CacheStatQueueProfiles:       {},
		utils.CacheStatQueues:              {},
		utils.CacheRankingProfiles:         {},
		utils.CacheSTIR:                    {},
		utils.CacheRouteFilterIndexes:      {},
		utils.CacheRouteProfiles:           {},
		utils.CacheThresholdFilterIndexes:  {},
		utils.CacheThresholdProfiles:       {},
		utils.CacheThresholds:              {},
		utils.CacheTimings:                 {},
		utils.CacheDiameterMessages:        {},
		utils.CacheClosedSessions:          {},
		utils.CacheLoadIDs:                 {},
		utils.CacheRPCConnections:          {},
		utils.CacheCDRIDs:                  {},
		utils.CacheRatingProfilesTmp:       {},
		utils.CacheUCH:                     {},
		utils.CacheEventCharges:            {},
		utils.CacheTrendProfiles:           {},
		utils.CacheTrends:                  {},
		utils.CacheReverseFilterIndexes:    {},
		utils.MetaAPIBan:                   {},
		utils.MetaSentryPeer:               {},
		utils.CacheCapsEvents:              {},
		utils.CacheReplicationHosts:        {},
		utils.CacheRadiusPackets:           {},
	}
}

// NewRPCClient creates and returns a new RPC client for cgr-engine.
func NewRPCClient(t testing.TB, cfg *config.ListenCfg) *birpc.Client {
	t.Helper()
	var err error
	var client *birpc.Client
	switch *utils.Encoding {
	case utils.MetaJSON:
		client, err = jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		client, err = birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		t.Fatalf("unsupported RPC encoding: %s", *utils.Encoding)
	}
	if err != nil {
		t.Fatalf("unable to connect to cgr-engine: %v", err)
	}
	return client

}

// TestEngine holds the setup parameters and configurations
// required for running integration tests.
type TestEngine struct {
	ConfigPath     string            // path to the main configuration file
	ConfigJSON     string            // JSON cfg content (standalone/overwrites static configs)
	DBCfg          DBCfg             // custom db settings for dynamic setup (overrides static config)
	LogBuffer      io.Writer         // captures log output of the test environment
	PreserveDataDB bool              // prevents automatic data_db flush when set
	PreserveStorDB bool              // prevents automatic stor_db flush when set
	TpPath         string            // path to the tariff plans
	TpFiles        map[string]string // CSV data for tariff plans: filename -> content

	// PreStartHook executes custom logic relying on CGRConfig
	// before starting cgr-engine.
	PreStartHook func(testing.TB, *config.CGRConfig)

	// TODO: add possibility to pass environment vars
}

// Run initializes a cgr-engine instance for testing. It calls t.Fatal on any setup failure.
func (ng TestEngine) Run(t testing.TB, extraFlags ...string) (*birpc.Client, *config.CGRConfig) {
	t.Helper()
	cfg := parseCfg(t, ng.ConfigPath, ng.ConfigJSON, ng.DBCfg)
	flushDBs(t, cfg, !ng.PreserveDataDB, !ng.PreserveStorDB)
	if ng.PreStartHook != nil {
		ng.PreStartHook(t, cfg)
	}
	startEngine(t, cfg, ng.LogBuffer)
	client := NewRPCClient(t, cfg.ListenCfg())
	LoadCSVs(t, client, ng.TpPath, ng.TpFiles)
	return client, cfg
}

// DBParams contains database connection parameters.
type DBParams struct {
	Type     *string `json:"db_type,omitempty"`
	Host     *string `json:"db_host,omitempty"`
	Port     *int    `json:"db_port,omitempty"`
	Name     *string `json:"db_name,omitempty"`
	User     *string `json:"db_user,omitempty"`
	Password *string `json:"db_password,omitempty"`
}

// DBCfg holds the configurations for data_db and/or stor_db.
type DBCfg struct {
	DataDB *DBParams `json:"data_db,omitempty"`
	StorDB *DBParams `json:"stor_db,omitempty"`
}

// parseCfg initializes and returns a CGRConfig. It handles both static and
// dynamic configs, including custom DB settings. For dynamic configs, it
// creates temporary configuration files in a new directory.
func parseCfg(t testing.TB, cfgPath, cfgJSON string, dbCfg DBCfg) (cfg *config.CGRConfig) {
	t.Helper()
	if cfgPath == "" && cfgJSON == "" {
		t.Fatal("missing config source")
	}

	// Defer CGRConfig constructor to avoid repetition.
	// cfg (named return) will be set by the deferred function.
	// cfgPath is guaranteed non-empty on successful return.
	defer func() {
		t.Helper()
		var err error
		cfg, err = config.NewCGRConfigFromPath(cfgPath)
		if err != nil {
			t.Fatalf("could not init config from path %s: %v", cfgPath, err)
		}
	}()

	hasCustomDBConfig := dbCfg.DataDB != nil || dbCfg.StorDB != nil
	if cfgPath != "" && cfgJSON == "" && !hasCustomDBConfig {
		// Config file already exists and is static; no need for
		// further processing.
		return
	}

	// Reaching this point means the configuration is at least partially dynamic.

	tmp := t.TempDir()
	if cfgPath != "" {
		// An existing configuration directory is specified. Since
		// configuration is not completely static, it's better to copy
		// its contents to the temporary directory instead.
		if err := os.CopyFS(tmp, os.DirFS(cfgPath)); err != nil {
			t.Fatal(err)
		}
	}
	cfgPath = tmp

	if hasCustomDBConfig {
		// Create a new JSON configuration file based on the DBConfigs object.
		b, err := json.Marshal(dbCfg)
		if err != nil {
			t.Fatal(err)
		}
		dbFilePath := filepath.Join(cfgPath, "zzz_dynamic_db.json")
		if err := os.WriteFile(dbFilePath, b, 0644); err != nil {
			t.Fatal(err)
		}
	}

	if cfgJSON != "" {
		// A JSON configuration string has been passed to the object.
		// It can be standalone or used to overwrite sections from an
		// existing configuration file. In case it's the latter, ensure
		// the file is processed towards the end.
		filePath := filepath.Join(cfgPath, "zzz_dynamic_cgrates.json")
		if err := os.WriteFile(filePath, []byte(cfgJSON), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return
}

func LoadCSVsWithCGRLoader(t testing.TB, cfgPath, tpPath string, logBuffer io.Writer, csvFiles map[string]string, extraFlags ...string) {
	t.Helper()

	if tpPath == "" && len(csvFiles) == 0 {
		return // nothing to load
	}

	paths := make([]string, 0, 2)
	var customTpPath string
	if len(csvFiles) != 0 {
		customTpPath = t.TempDir()
	}
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

	for _, path := range paths {
		flags := []string{"-config_path", cfgPath, "-path", path}
		flags = append(flags, extraFlags...)
		loader := exec.Command("cgr-loader", flags...)
		if logBuffer != nil {
			loader.Stdout = logBuffer
			loader.Stderr = logBuffer
		}
		if err := loader.Run(); err != nil {
			t.Fatal(err)
		}
	}
}

// LoadCSVs loads tariff plan data from CSV files into the service. It handles directory creation and file
// writing for custom paths, and loads data from the specified paths using the provided RPC client.
func LoadCSVs(t testing.TB, client *birpc.Client, tpPath string, csvFiles map[string]string) {
	t.Helper()

	if tpPath == "" && len(csvFiles) == 0 {
		return // nothing to load
	}

	paths := make([]string, 0, 2)
	var customTpPath string
	if len(csvFiles) != 0 {
		customTpPath = t.TempDir()
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	WaitForService(t, ctx, client, utils.APIerSv1)

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
func flushDBs(t testing.TB, cfg *config.CGRConfig, flushDataDB, flushStorDB bool) {
	t.Helper()
	if flushDataDB {
		if err := InitDataDb(cfg); err != nil {
			t.Fatalf("failed to flush %s dataDB: %v", cfg.DataDbCfg().Type, err)
		}
	}
	if flushStorDB {
		if err := InitStorDb(cfg); err != nil {
			t.Fatalf("failed to flush %s storDB: %v", cfg.StorDbCfg().Type, err)
		}
	}
}

// startEngine starts the CGR engine process with the provided configuration. It writes engine logs to the
// provided logBuffer (if any).
func startEngine(t testing.TB, cfg *config.CGRConfig, logBuffer io.Writer, extraFlags ...string) {
	t.Helper()
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"-config_path", cfg.ConfigPath}
	if logBuffer != nil {
		flags = append(flags, "-logger", utils.MetaStdLog)
	}
	if len(extraFlags) != 0 {
		flags = append(flags, extraFlags...)
	}
	engine := exec.Command(binPath, flags...)
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
	backoff := utils.FibDuration(time.Millisecond, 0)
	for i := 0; i < 16; i++ {
		time.Sleep(backoff())
		if _, err = jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("failed to start cgr-engine: %v", err)
	}
}

func WaitForService(t testing.TB, ctx *context.Context, client *birpc.Client, service string) {
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

// Default DB configurations. For Redis/MySQL, it's missing because
// it's the default.
var (
	InternalDBCfg = DBCfg{
		DataDB: &DBParams{
			Type: utils.StringPointer(utils.MetaInternal),
		},
		StorDB: &DBParams{
			Type: utils.StringPointer(utils.MetaInternal),
		},
	}
	MongoDBCfg = DBCfg{
		DataDB: &DBParams{
			Type: utils.StringPointer(utils.MetaMongo),
			Port: utils.IntPointer(27017),
			Name: utils.StringPointer("10"),
		},
		StorDB: &DBParams{
			Type:     utils.StringPointer(utils.MetaMongo),
			Port:     utils.IntPointer(27017),
			Name:     utils.StringPointer("cgrates"),
			Password: utils.StringPointer(""),
		},
	}
	PostgresDBCfg = DBCfg{
		StorDB: &DBParams{
			Type: utils.StringPointer(utils.MetaPostgres),
			Port: utils.IntPointer(5432),
		},
	}
)
