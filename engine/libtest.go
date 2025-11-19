//go:build integration || flaky || call || performance || kafka

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
	"syscall"
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

func InitDB(cfg *config.CGRConfig) error {
	for _, dbConn := range cfg.DbCfg().DBConns {
		dataDB, err := NewDataDBConn(dbConn.Type,
			dbConn.Host, dbConn.Port,
			dbConn.Name, dbConn.User,
			dbConn.Password, cfg.GeneralCfg().DBDataEncoding,
			dbConn.StringIndexedFields, dbConn.PrefixIndexedFields,
			dbConn.Opts, cfg.DbCfg().Items)
		if err != nil {
			return err
		}
		defer dataDB.Close()
		var dbPath string
		if slices.Contains([]string{utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres},
			dbConn.Type) {
			dbPath = path.Join(cfg.DataFolderPath, "storage", strings.Trim(dbConn.Type, "*"))
		}
		if err := dataDB.Flush(dbPath); err != nil {
			return err
		}
		// Set versions before starting.
		if slices.Contains([]string{utils.MetaMongo, utils.MetaMySQL, utils.MetaPostgres},
			dbConn.Type) {
			if err := SetDBVersions(dataDB); err != nil {
				return err
			}
		} else {
			if err := OverwriteDBVersions(dataDB); err != nil {
				return err
			}
		}
	}
	return nil
}

func InitConfigDB(cfg *config.CGRConfig) error {
	d, err := NewDataDBConn(cfg.ConfigDBCfg().Type,
		cfg.ConfigDBCfg().Host, cfg.ConfigDBCfg().Port,
		cfg.ConfigDBCfg().Name, cfg.ConfigDBCfg().User,
		cfg.ConfigDBCfg().Password, cfg.GeneralCfg().DBDataEncoding, nil, nil,
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
	for range 16 {
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

// Starts the engine from a string JSON config
func StartEngineFromString(cfgJSON string, waitEngine int, t testing.TB) (*exec.Cmd, error) {
	cfgPath := t.TempDir()
	// A JSON configuration string has been passed to the object.
	// It can be standalone or used to overwrite sections from an
	// existing configuration file. In case it's the latter, ensure
	// the file is processed towards the end.
	filePath := filepath.Join(cfgPath, "zzz_dynamic_cgrates.json")
	if err := os.WriteFile(filePath, []byte(cfgJSON), 0644); err != nil {
		return nil, err
	}
	var err error
	cfg, err := config.NewCGRConfigFromPath(context.TODO(), cfgPath)
	if err != nil {
		return nil, fmt.Errorf("could not init config from path %s: %v", cfgPath, err)
	}
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		return nil, err
	}
	flags := []string{"-config_path", cfg.ConfigPath}
	engine := exec.Command(binPath, flags...)
	if err := engine.Start(); err != nil {
		return nil, fmt.Errorf("cgr-engine command failed: %v", err)
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

func LoadTariffPlanFromFolder(tpPath, timezone string, dm *DataManager, disableReverse bool,
	cacheConns, schedConns []string) error {
	csvStorage, err := NewFileCSVStorage(utils.CSVSep, tpPath)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	dataDBs := make(map[string]DataDB, len(dm.DataDB()))
	for connID, dataDB := range dm.DataDB() {
		dataDBs[connID] = dataDB
	}
	dbcManager := NewDBConnManager(dataDBs, dm.cfg.DbCfg())
	loader, err := NewTpReader(dbcManager, csvStorage, "",
		timezone, cacheConns, schedConns)
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
		utils.CacheEventResources:              {},
		utils.CacheEventIPs:                    {},
		utils.CacheFilters:                     {},
		utils.CacheResourceFilterIndexes:       {},
		utils.CacheResourceProfiles:            {},
		utils.CacheResources:                   {},
		utils.CacheIPFilterIndexes:             {},
		utils.CacheIPProfiles:                  {},
		utils.CacheIPAllocations:               {},
		utils.CacheRPCResponses:                {},
		utils.CacheStatFilterIndexes:           {},
		utils.CacheStatQueueProfiles:           {},
		utils.CacheStatQueues:                  {},
		utils.CacheSTIR:                        {},
		utils.CacheRankingProfiles:             {},
		utils.CacheRankings:                    {},
		utils.CacheRouteFilterIndexes:          {},
		utils.CacheRouteProfiles:               {},
		utils.CacheThresholdFilterIndexes:      {},
		utils.CacheThresholdProfiles:           {},
		utils.CacheThresholds:                  {},
		utils.CacheTrendProfiles:               {},
		utils.CacheTrends:                      {},
		utils.CacheRateProfiles:                {},
		utils.CacheRateProfilesFilterIndexes:   {},
		utils.CacheRateFilterIndexes:           {},
		utils.CacheDiameterMessages:            {},
		utils.CacheRadiusPackets:               {},
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

// NewRPCClient creates and returns a new RPC client for cgr-engine.
func NewRPCClient(t testing.TB, cfg *config.ListenCfg, encoding string) *birpc.Client {
	t.Helper()
	var err error
	var client *birpc.Client
	switch encoding {
	case utils.MetaJSON:
		client, err = jsonrpc.Dial(utils.TCP, cfg.RPCJSONListen)
	case utils.MetaGOB:
		client, err = birpc.Dial(utils.TCP, cfg.RPCGOBListen)
	default:
		t.Fatalf("unsupported RPC encoding: %q", encoding)
	}
	if err != nil {
		t.Fatalf("unable to connect to cgr-engine: %v", err)
	}
	return client
}

// TestEngine holds the setup parameters and configurations
// required for running integration tests.
type TestEngine struct {
	ConfigPath       string            // path to the main configuration file
	ConfigJSON       string            // JSON cfg content (standalone/overwrites static configs)
	DBCfg            DBCfg             // custom db settings for dynamic setup (overrides static config)
	Encoding         string            // data encoding type (e.g. JSON, GOB)
	LogBuffer        io.Writer         // captures log output of the test environment
	PreserveDataDB   bool              // prevents automatic data_db flush when set
	TpPath           string            // path to the tariff plans
	TpFiles          map[string]string // CSV data for tariff plans: filename -> content
	GracefulShutdown bool              // shutdown the engine gracefuly, otherwise use process.Kill

	// PreStartHook executes custom logic relying on CGRConfig
	// before starting cgr-engine.
	PreStartHook func(testing.TB, *config.CGRConfig)

	// TODO: add possibility to pass environment vars
}

// Run initializes a cgr-engine instance for testing. It calls t.Fatal on any setup failure.
func (ng TestEngine) Run(t testing.TB, extraFlags ...string) (*birpc.Client, *config.CGRConfig) {
	t.Helper()
	cfg := parseCfg(t, ng.ConfigPath, ng.ConfigJSON, ng.DBCfg)
	FlushDBs(t, cfg, !ng.PreserveDataDB)
	if ng.TpPath != "" || len(ng.TpFiles) != 0 {
		if ng.TpPath == "" {
			ng.TpPath = t.TempDir()
		}
		setupLoader(t, ng.TpPath, cfg.ConfigPath)
	}
	if ng.PreStartHook != nil {
		ng.PreStartHook(t, cfg)
	}
	startEngine(t, cfg, ng.LogBuffer, ng.GracefulShutdown, extraFlags...)
	client := NewRPCClient(t, cfg.ListenCfg(), ng.Encoding)
	if ng.TpPath == "" {
		ng.TpPath = cfg.LoaderCfg()[0].TpInDir
	}
	// cfg gets edited in files but not in variable, get the cfg variable from files
	newCfg, err := config.NewCGRConfigFromPath(context.Background(), cfg.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if newCfg.LoaderCfg().Enabled() {
		WaitForServiceStart(t, client, utils.LoaderS, 200*time.Millisecond)
	}
	loadCSVs(t, ng.TpPath, ng.TpFiles)
	return client, newCfg
}

// Opts contains opts of database
type Opts struct {
	InternalDBDumpPath        *string `json:"internalDBDumpPath,omitempty"`
	InternalDBDumpInterval    *string `json:"internalDBDumpInterval,omitempty"`
	InternalDBRewriteInterval *string `json:"internalDBRewriteInterval,omitempty"`
}

// DBConn contains database connection parameters.
type DBConn struct {
	Type     *string `json:"db_type,omitempty"`
	Host     *string `json:"db_host,omitempty"`
	Port     *int    `json:"db_port,omitempty"`
	Name     *string `json:"db_name,omitempty"`
	User     *string `json:"db_user,omitempty"`
	Password *string `json:"db_password,omitempty"`
	Opts     Opts    `json:"opts,omitempty"`
}

// Item contains db item parameters
type Item struct {
	Limit  *int    `json:"limit,omitempty"`
	DbConn *string `json:"dbConn,omitempty"`
}

// DBParams contains database connection parameters.
type DBParams struct {
	DBConns map[string]DBConn `json:"db_conns,omitempty"`
	Items   map[string]Item   `json:"items,omitempty"`
}

// DBCfg holds the configurations for data_db and/or stor_db.
type DBCfg struct {
	DB *DBParams `json:"db,omitempty"`
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
		cfg, err = config.NewCGRConfigFromPath(context.TODO(), cfgPath)
		if err != nil {
			t.Fatalf("could not init config from path %s: %v", cfgPath, err)
		}
	}()

	hasCustomDBConfig := dbCfg.DB != nil
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

// setupLoader configures the *default loader to automatically load from the
// specified path.
func setupLoader(t testing.TB, tpPath, cfgPath string) {
	t.Helper()
	loadersJSON := fmt.Sprintf(`{
"loaders": [{
	"id": "*default",
	"enabled": true,
	"run_delay": "-1",
	"tp_in_dir": "%s",
	"tp_out_dir": "",
	"action": "*store",
	"opts": {
		"*stopOnError": true
	}
}]
}`, tpPath)
	filePath := filepath.Join(cfgPath, "zzz_dynamic_loader.json")
	if err := os.WriteFile(filePath, []byte(loadersJSON), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Remove(filepath.Join(cfgPath, "zzz_dynamic_loader.json")); err != nil {
			t.Error(err)
		}
	})
}

// loadCSVs loads tariff plan data from CSV files. The CSV files are created
// based on the csvFiles map, where the key represents the file name and the
// value the contains its contents. Assumes the data is loaded automatically
// (RunDelay != 0)
func loadCSVs(t testing.TB, tpPath string, csvFiles map[string]string) {
	t.Helper()
	if tpPath != "" {
		for fileName, content := range csvFiles {
			filePath := path.Join(tpPath, fileName)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("could not write to file %s: %v", filePath, err)
			}
		}
		t.Cleanup(func() {

		})
	}
}

// FlushDBs resets the databases specified in the configuration if the
// corresponding flags are true.
func FlushDBs(t testing.TB, cfg *config.CGRConfig, flushDataDB bool) {
	t.Helper()
	if flushDataDB {
		if err := InitDB(cfg); err != nil {
			t.Fatalf("failed to flush DataDB err: %v", err)
		}
	}
}

// startEngine starts the CGR engine process with the provided configuration
// and flags. It writes engine logs to the provided logBuffer (if any).
func startEngine(t testing.TB, cfg *config.CGRConfig, logBuffer io.Writer, gracefulShutdown bool, extraFlags ...string) {
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
		if gracefulShutdown {
			if err := engine.Process.Signal(syscall.SIGTERM); err != nil {
				t.Errorf("failed to kill cgr-engine process (%d): %v", engine.Process.Pid, err)
			}
			if err := engine.Wait(); err != nil {
				t.Errorf("cgr-engine process failed to exit cleanly: %v", err)
				t.Log("Logs: \n", logBuffer)
			}
		} else {
			if err := engine.Process.Kill(); err != nil {
				t.Errorf("failed to kill cgr-engine process (%d): %v", engine.Process.Pid, err)
			}
		}
	})
	backoff := utils.FibDuration(time.Millisecond, 0)
	for range 16 {
		time.Sleep(backoff())
		if _, err := jsonrpc.Dial(utils.TCP, cfg.ListenCfg().RPCJSONListen); err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("failed to start cgr-engine: %v", err)
	}
}

// serviceReceivers maps service names to their RPC receiver names.
// Services with empty receiver names don't implement a Ping method.
// Used for service availability testing (through pinging).
var serviceReceivers = map[string]string{
	utils.CacheS:          utils.CacheSv1,
	utils.ConfigS:         utils.ConfigSv1,
	utils.CoreS:           utils.CoreSv1,
	utils.AccountS:        utils.AccountSv1,
	utils.ActionS:         utils.ActionSv1,
	utils.AdminS:          utils.AdminSv1,
	utils.AnalyzerS:       utils.AnalyzerSv1,
	utils.AttributeS:      utils.AttributeSv1,
	utils.CDRServer:       utils.CDRsV1,
	utils.ChargerS:        utils.ChargerSv1,
	utils.EEs:             utils.EeSv1,
	utils.EFs:             utils.EfSv1,
	utils.ERs:             utils.ErSv1,
	utils.RateS:           utils.RateSv1,
	utils.ResourceS:       utils.ResourceSv1,
	utils.IPs:             utils.IPsV1,
	utils.RouteS:          utils.RouteSv1,
	utils.SessionS:        utils.SessionSv1,
	utils.StatS:           utils.StatSv1,
	utils.TPeS:            utils.TPeSv1,
	utils.ThresholdS:      utils.ThresholdSv1,
	utils.LoaderS:         utils.LoaderSv1,
	utils.TrendS:          utils.TrendSv1,
	utils.RankingS:        utils.RankingSv1,
	utils.CapS:            "",
	utils.CommonListenerS: "",
	utils.ConnManager:     "",
	utils.DB:              "",
	utils.FilterS:         "",
	utils.GlobalVarS:      "",
	utils.LoggerS:         "",
	utils.RegistrarC:      "",
	utils.AsteriskAgent:   "",
	utils.DiameterAgent:   "",
	utils.DNSAgent:        "",
	utils.FreeSWITCHAgent: "",
	utils.HTTPAgent:       "",
	utils.JanusAgent:      "",
	utils.KamailioAgent:   "",
	utils.RadiusAgent:     "",
	utils.SIPAgent:        "",
}

// WaitForServiceStart tries to ping the service until it receives a "Pong"
// reply or times out. Test will be marked as failed on timeout.
func WaitForServiceStart(t testing.TB, client *birpc.Client, service string, timeout time.Duration) {
	t.Helper()

	receiver := serviceReceivers[service]
	if receiver == "" {
		// Skip services that don't have a Ping method.
		return
	}
	method := receiver + ".Ping"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	backoff := utils.FibDuration(time.Millisecond, 0)
	var reply string
	for {
		err := client.Call(context.Background(), method, nil, &reply)
		if err == nil && reply == utils.Pong {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("service %q did not become available within %s", service, timeout)
		case <-time.After(backoff()):
			// continue to next iteration
		}
	}
}

// WaitForServiceShutdown tries to ping the service until it receives a "can't
// find service" error reply or times out. Test will be marked as failed on
// timeout.
func WaitForServiceShutdown(t testing.TB, client *birpc.Client, service string, timeout time.Duration) {
	t.Helper()

	receiver := serviceReceivers[service]
	if receiver == "" {
		// Skip services that don't have a Ping method.
		return
	}
	method := receiver + ".Ping"

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	backoff := utils.FibDuration(time.Millisecond, 0)
	var reply string
	for {
		err := client.Call(context.Background(), method, nil, &reply)
		if err != nil && strings.Contains(err.Error(), "can't find service") {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("service %q did not shut down within %s", service, timeout)
		case <-time.After(backoff()):
			// continue to next iteration
		}
	}
}

// Default DB configurations.
var (
	InternalDBCfg = DBCfg{
		DB: &DBParams{
			DBConns: map[string]DBConn{
				utils.MetaDefault: {
					Type: utils.StringPointer(utils.MetaInternal),
				},
			},
		},
	}
	MySQLDBCfg = DBCfg{
		DB: &DBParams{
			DBConns: map[string]DBConn{
				utils.MetaDefault: {
					Type:     utils.StringPointer(utils.MetaMySQL),
					Host:     utils.StringPointer("127.0.0.1"),
					Port:     utils.IntPointer(3306),
					Name:     utils.StringPointer(utils.CGRateSLwr),
					User:     utils.StringPointer(utils.CGRateSLwr),
					Password: utils.StringPointer("CGRateS.org"),
				},
			},
		},
	}
	RedisDBCfg = DBCfg{
		DB: &DBParams{
			DBConns: map[string]DBConn{
				utils.MetaDefault: {
					Type: utils.StringPointer(utils.MetaRedis),
					Host: utils.StringPointer("127.0.0.1"),
					Port: utils.IntPointer(6379),
					Name: utils.StringPointer("10"),
					User: utils.StringPointer(utils.CGRateSLwr),
				},
				utils.StorDB: {
					Type:     utils.StringPointer(utils.MetaMySQL),
					Host:     utils.StringPointer("127.0.0.1"),
					Port:     utils.IntPointer(3306),
					Name:     utils.StringPointer(utils.CGRateSLwr),
					User:     utils.StringPointer(utils.CGRateSLwr),
					Password: utils.StringPointer("CGRateS.org"),
				},
			},
			Items: map[string]Item{
				utils.MetaCDRs: {
					Limit:  utils.IntPointer(-1),
					DbConn: utils.StringPointer(utils.StorDB),
				},
			},
		},
	}
	MongoDBCfg = DBCfg{
		DB: &DBParams{
			DBConns: map[string]DBConn{
				utils.MetaDefault: {
					Type: utils.StringPointer(utils.MetaMongo),
					Host: utils.StringPointer("127.0.0.1"),
					Port: utils.IntPointer(27017),
					Name: utils.StringPointer("10"),
					User: utils.StringPointer(utils.CGRateSLwr),
				},
				utils.StorDB: {
					Type:     utils.StringPointer(utils.MetaMongo),
					Host:     utils.StringPointer("127.0.0.1"),
					Port:     utils.IntPointer(27017),
					Name:     utils.StringPointer(utils.CGRateSLwr),
					User:     utils.StringPointer(utils.CGRateSLwr),
					Password: utils.StringPointer(""),
				},
			},
			Items: map[string]Item{
				utils.MetaCDRs: {
					Limit:  utils.IntPointer(-1),
					DbConn: utils.StringPointer(utils.StorDB),
				},
			},
		},
	}
	PostgresDBCfg = DBCfg{
		DB: &DBParams{
			DBConns: map[string]DBConn{
				utils.MetaDefault: {
					Type:     utils.StringPointer(utils.MetaPostgres),
					Host:     utils.StringPointer("127.0.0.1"),
					Port:     utils.IntPointer(5432),
					Name:     utils.StringPointer(utils.CGRateSLwr),
					User:     utils.StringPointer(utils.CGRateSLwr),
					Password: utils.StringPointer("CGRateS.org"),
				},
			},
		},
	}
)

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
