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
		dataDB, err := NewDBConn(dbConn.Type,
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
	d, err := NewDBConn(cfg.ConfigDBCfg().Type,
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
	go engine.Wait() // so pkill'd engines don't stay defunct
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
	go engine.Wait()                                         // so pkill'd engines don't stay defunct
	time.Sleep(time.Duration(waitEngine) * time.Millisecond) // wait for rater to register all subsystems
	return engine, nil
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
	dataDBs := make(map[string]DataDB, len(dm.DB()))
	for connID, dataDB := range dm.DB() {
		dataDBs[connID] = dataDB
	}
	dbcManager := NewDBConnManager(dataDBs, dm.cfg.DbCfg())
	loader, err := NewTpReader(dbcManager, csvStorage, "",
		timezone, cacheConns, schedConns, nil)
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
	Persist          bool              // keep generated config dir after test (logs its path)

	// PreStartHook executes custom logic relying on CGRConfig
	// before starting cgr-engine.
	PreStartHook func(testing.TB, *config.CGRConfig)

	cmd        *exec.Cmd
	cfg        *config.CGRConfig
	logBuf     *bytes.Buffer
	stopped    bool
	extraFlags []string
}

// Run sets up and starts a cgr-engine for testing. Engine logs
// are dumped via t.Log on failure.
func (ng *TestEngine) Run(t testing.TB, extraFlags ...string) (*birpc.Client, *config.CGRConfig) {
	t.Helper()
	ng.extraFlags = extraFlags
	ng.cfg = parseCfg(t, ng.ConfigPath, ng.ConfigJSON, ng.DBCfg, ng.Persist)
	FlushDBs(t, ng.cfg, !ng.PreserveDataDB)
	loadData := ng.TpPath != "" || len(ng.TpFiles) != 0
	if loadData {
		if ng.TpPath == "" {
			ng.TpPath = t.TempDir()
		}
		setupLoader(t, ng.TpPath, ng.cfg.ConfigPath)
	}
	if ng.PreStartHook != nil {
		ng.PreStartHook(t, ng.cfg)
	}

	ng.logBuf = new(bytes.Buffer)
	ng.start(t)

	logBuf := ng.logBuf
	t.Cleanup(func() {
		ng.stop(t)
		if t.Failed() && logBuf.Len() > 0 {
			t.Log(logBuf.String())
		}
	})

	client := NewRPCClient(t, ng.cfg.ListenCfg(), ng.Encoding)
	if ng.TpPath == "" {
		ng.TpPath = ng.cfg.LoaderCfg()[0].TpInDir
	}
	// cfg gets edited in files but not in variable, get the cfg variable from files
	newCfg, err := config.NewCGRConfigFromPath(context.Background(), ng.cfg.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if loadData && newCfg.LoaderCfg().Enabled() {
		WaitForServiceStart(t, client, utils.LoaderS, 200*time.Millisecond)
		for fileName, content := range ng.TpFiles {
			filePath := path.Join(ng.TpPath, fileName)
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("could not write to file %s: %v", filePath, err)
			}
		}
		var reply string
		if err := client.Call(context.Background(), utils.LoaderSv1Run,
			&struct{ LoaderID string }{}, &reply); err != nil {
			t.Fatal(err)
		}
	}
	return client, newCfg
}

// Stop shuts down the engine. No-op if already stopped or never started.
func (ng *TestEngine) Stop(t testing.TB) {
	t.Helper()
	ng.stop(t)
}

// Start re-starts the engine after a stop. Returns a fresh RPC client.
// Fatal if Run() was never called.
func (ng *TestEngine) Start(t testing.TB) *birpc.Client {
	t.Helper()
	if ng.cfg == nil {
		t.Fatal("Start() called before Run()")
	}
	ng.start(t)
	return NewRPCClient(t, ng.cfg.ListenCfg(), ng.Encoding)
}

func (ng *TestEngine) logWriter() io.Writer {
	if ng.LogBuffer != nil {
		return io.MultiWriter(ng.logBuf, ng.LogBuffer)
	}
	return ng.logBuf
}

func (ng *TestEngine) start(t testing.TB) {
	t.Helper()
	if ng.cmd != nil && !ng.stopped {
		ng.stop(t)
	}
	binPath, err := exec.LookPath("cgr-engine")
	if err != nil {
		t.Fatal(err)
	}
	flags := []string{"-config_path", ng.cfg.ConfigPath, "-logger", utils.MetaStdLog}
	flags = append(flags, ng.extraFlags...)
	ng.cmd = exec.Command(binPath, flags...)
	w := ng.logWriter()
	ng.cmd.Stdout = w
	ng.cmd.Stderr = w
	if err := ng.cmd.Start(); err != nil {
		t.Fatalf("cgr-engine command failed: %v", err)
	}
	ng.stopped = false
	backoff := utils.FibDuration(time.Millisecond, 0)
	var dialErr error
	for range 16 {
		time.Sleep(backoff())
		var conn *birpc.Client
		if conn, dialErr = jsonrpc.Dial(utils.TCP, ng.cfg.ListenCfg().RPCJSONListen); dialErr == nil {
			conn.Close()
			return
		}
	}
	t.Fatalf("engine did not open port <%s>: %v", ng.cfg.ListenCfg().RPCJSONListen, dialErr)
}

func (ng *TestEngine) stop(t testing.TB) {
	t.Helper()
	if ng.cmd == nil || ng.stopped {
		return
	}
	ng.stopped = true
	if ng.GracefulShutdown {
		if err := ng.cmd.Process.Signal(syscall.SIGTERM); err != nil {
			t.Errorf("failed to terminate cgr-engine process (%d): %v",
				ng.cmd.Process.Pid, err)
		}
		if err := ng.cmd.Wait(); err != nil {
			t.Errorf("cgr-engine process (%d) exited with error: %v",
				ng.cmd.Process.Pid, err)
		}
		return
	}
	if err := ng.cmd.Process.Kill(); err != nil {
		t.Errorf("failed to kill cgr-engine process (%d): %v",
			ng.cmd.Process.Pid, err)
	}
	_ = ng.cmd.Wait()
}

// DBConnOpts contains opts of database
type DBConnOpts struct {
	InternalDBDumpPath        *string `json:"internalDBDumpPath,omitempty"`
	InternalDBDumpInterval    *string `json:"internalDBDumpInterval,omitempty"`
	InternalDBRewriteInterval *string `json:"internalDBRewriteInterval,omitempty"`
}

// DBConn contains database connection parameters.
type DBConn struct {
	Type     *string    `json:"db_type,omitempty"`
	Host     *string    `json:"db_host,omitempty"`
	Port     *int       `json:"db_port,omitempty"`
	Name     *string    `json:"db_name,omitempty"`
	User     *string    `json:"db_user,omitempty"`
	Password *string    `json:"db_password,omitempty"`
	Opts     DBConnOpts `json:"opts"`
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

// parseCfg initializes and returns a CGRConfig. For dynamic configs, it
// creates configuration files in a temporary directory. If persist is true,
// the directory is kept after the test and its path is logged.
func parseCfg(t testing.TB, cfgPath, cfgJSON string, dbCfg DBCfg, persist bool) *config.CGRConfig {
	t.Helper()
	if cfgPath == "" && cfgJSON == "" {
		t.Fatal("missing config source")
	}

	if cfgPath != "" && cfgJSON == "" && dbCfg.DB == nil {
		cfg, err := config.NewCGRConfigFromPath(context.TODO(), cfgPath)
		if err != nil {
			t.Fatalf("could not init config from path %s: %v", cfgPath, err)
		}
		return cfg
	}

	var tmp string
	if persist {
		var err error
		tmp, err = os.MkdirTemp("", "cgr-test-*")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("config dir: %s", tmp)
	} else {
		tmp = t.TempDir()
	}
	if cfgPath != "" {
		if err := os.CopyFS(tmp, os.DirFS(cfgPath)); err != nil {
			t.Fatal(err)
		}
	}

	if dbCfg.DB != nil {
		b, err := json.Marshal(dbCfg)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "zzz_dynamic_db.json"), b, 0644); err != nil {
			t.Fatal(err)
		}
	}

	if cfgJSON != "" {
		filePath := filepath.Join(tmp, "zzz_dynamic_cgrates.json")
		if err := os.WriteFile(filePath, []byte(cfgJSON), 0644); err != nil {
			t.Fatal(err)
		}
	}

	cfg, err := config.NewCGRConfigFromPath(context.TODO(), tmp)
	if err != nil {
		t.Fatalf("could not init config from path %s: %v", tmp, err)
	}
	return cfg
}

// setupLoader configures the *default loader to load from the specified path.
func setupLoader(t testing.TB, tpPath, cfgPath string) {
	t.Helper()
	loadersJSON := fmt.Sprintf(`{
"loaders": [{
	"id": "*default",
	"enabled": true,
	"run_delay": "0",
	"tp_in_dir": "%s",
	"tp_out_dir": "",
	"action": "*store",
	"opts": {
		"*stopOnError": false
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
					Opts: DBConnOpts{
						InternalDBDumpInterval:    utils.StringPointer("0s"),
						InternalDBRewriteInterval: utils.StringPointer("0s"),
					},
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
