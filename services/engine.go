// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

// flags holds all command line arguments. Flag descriptions are set in newFlags.
type flags struct {
	*flag.FlagSet
	config struct {
		path    string
		check   bool
		print   bool
		version bool
	}
	process struct {
		pidFile           string
		singleCPU         bool
		scheduledShutdown time.Duration
	}
	profiling struct {
		cpu struct {
			dir string
		}
		mem struct {
			dir      string
			interval time.Duration
			maxFiles int
			useTS    bool
		}
	}
	logger struct {
		level  int
		nodeID string
		typ    string // syslog|stdout|kafkaLog
	}
	data struct {
		preloadIDs  []string
		setVersions bool
	}
}

// newFlags creates and initializes a new flags instance with default values.
func newFlags() *flags {
	f := &flags{
		FlagSet: flag.NewFlagSet("cgr-engine", flag.ExitOnError),
	}

	f.StringVar(&f.config.path, utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path")
	f.BoolVar(&f.config.check, utils.CheckCfgCgr, false, "Verify the config without starting the engine")
	f.BoolVar(&f.config.print, utils.PrintCfgCgr, false, "Print configuration object in JSON format")
	f.BoolVar(&f.config.version, utils.VersionCgr, false, "Print application version and exit")

	f.StringVar(&f.process.pidFile, utils.PidCgr, "", "Path to write the PID file")
	f.BoolVar(&f.process.singleCPU, utils.SingleCpuCgr, false, "Run on a single CPU core")
	f.DurationVar(&f.process.scheduledShutdown, utils.ScheduledShutdownCgr, 0, "Shutdown the engine after the specified duration")

	f.StringVar(&f.profiling.cpu.dir, utils.CpuProfDirCgr, "", "Directory for CPU profiles")
	f.StringVar(&f.profiling.mem.dir, utils.MemProfDirCgr, "", "Directory for memory profiles")
	f.DurationVar(&f.profiling.mem.interval, utils.MemProfIntervalCgr, 15*time.Second, "Interval between memory profile saves")
	f.IntVar(&f.profiling.mem.maxFiles, utils.MemProfMaxFilesCgr, 1, "Number of memory profiles to keep (most recent)")
	f.BoolVar(&f.profiling.mem.useTS, utils.MemProfTimestampCgr, false, "Add timestamp to memory profile files")

	f.IntVar(&f.logger.level, utils.LogLevelCfg, -1, "Log level (0=emergency to 7=debug)")
	f.StringVar(&f.logger.nodeID, utils.NodeIDCfg, "", "Node ID of the engine")
	f.StringVar(&f.logger.typ, utils.LoggerCfg, "", "Logger type <*syslog|*stdout|*kafkaLog>")

	f.Func(utils.PreloadCgr, "Loader IDs used to load data before engine starts", func(val string) error {
		f.data.preloadIDs = strings.Split(val, utils.FieldsSep)
		return nil
	})
	f.BoolVar(&f.data.setVersions, utils.SetVersionsCgr, false, "Overwrite database versions")

	return f
}

// RunCGREngine initializes and runs the CGREngine with the provided command line arguments.
func RunCGREngine(fs []string) (err error) {
	flags := newFlags()
	flags.Parse(fs)

	var vers string
	if vers, err = utils.GetCGRVersion(); err != nil {
		return
	}
	if flags.config.version {
		fmt.Println(vers)
		return
	}
	if flags.process.pidFile != utils.EmptyString {
		if err = writePIDFile(flags.process.pidFile); err != nil {
			return
		}
		defer func() {
			if err := os.Remove(flags.process.pidFile); err != nil {
				utils.Logger.Warning("Could not remove pid file: " + err.Error())
			}
		}()
	}
	if flags.process.singleCPU {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	var cfg *config.CGRConfig
	if cfg, err = initConfigFromPath(context.TODO(), flags.config.path, flags.logger.nodeID,
		flags.logger.typ, flags.logger.level); err != nil || flags.config.check {
		return
	}

	if cfg.LoggerCfg().Level >= 0 {
		switch cfg.LoggerCfg().Type {
		case utils.MetaSysLog:
			utils.Logger, err = utils.NewSysLogger(cfg.GeneralCfg().NodeID, cfg.LoggerCfg().Level)
			if err != nil {
				return
			}
		case utils.MetaStdLog, utils.MetaKafkaLog:
			// If the logger is of type *kafka, use the *stdout logger until
			// LoggerService finishes startup.
			utils.Logger = utils.NewStdLogger(cfg.GeneralCfg().NodeID, cfg.LoggerCfg().Level)
		default:
			return fmt.Errorf("unsupported logger type: %q", cfg.LoggerCfg().Type)
		}
	}

	if flags.config.print {
		cfgJSON := utils.ToIJSON(cfg.AsMapInterface())
		utils.Logger.Info(fmt.Sprintf(
			"Configuration loaded from %q:\n%s", flags.config.path, cfgJSON))
	}

	shdWg := new(sync.WaitGroup)
	profiler := cores.NewProfiler(shdWg)
	if flags.profiling.cpu.dir != "" {
		cpuPath := filepath.Join(flags.profiling.cpu.dir, utils.CpuPathCgr)
		if err := profiler.StartCPUProfiling(cpuPath); err != nil {
			return err
		}
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CoreSCfg().ShutdownTimeout*10)
		defer cancel()
		if err := profiler.Shutdown(ctx); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
		utils.Logger.Info(fmt.Sprintf("<%s> CGRateS shutdown!", utils.ServiceManager))
	}()

	shdWg.Add(1)
	shutdown := utils.NewSyncedChan()
	go handleSignals(shutdown, cfg, shdWg)

	if flags.process.scheduledShutdown != 0 {
		shdWg.Add(1)
		go func() { // Schedule shutdown
			defer shdWg.Done()
			tm := time.NewTimer(flags.process.scheduledShutdown)
			select {
			case <-tm.C:
				shutdown.CloseOnce()
			case <-shutdown.Done():
				tm.Stop()
			}
		}()
	}

	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, runtime.Version()))

	registry := servmanager.NewRegistry()
	srvManager := servmanager.NewServiceManager(shdWg, cfg, registry, []servmanager.Service{
		NewGlobalVarS(cfg),
		NewCapService(cfg),
		NewCommonListenerService(cfg),
		NewAnalyzerService(cfg),
		NewConnManagerService(cfg),
		NewLoggerService(cfg, flags.logger.typ),
		NewDBService(cfg, flags.data.setVersions),
		NewConfigService(cfg),
		NewGuardianService(cfg),
		NewCoreService(cfg, profiler),
		NewCacheService(cfg),
		NewFilterService(cfg),
		NewLoaderService(cfg, flags.data.preloadIDs),
		NewExportFailoverService(cfg),
		NewAdminSv1Service(cfg),
		NewSessionService(cfg),
		NewAttributeService(cfg),
		NewChargerService(cfg),
		NewRouteService(cfg),
		NewResourceService(cfg),
		NewIPService(cfg),
		NewTrendService(cfg),
		NewRankingService(cfg),
		NewThresholdService(cfg),
		NewStatService(cfg),
		NewEventReaderService(cfg),
		NewDNSAgent(cfg),
		NewFreeswitchAgent(cfg),
		NewKamailioAgent(cfg),
		NewJanusAgent(cfg),
		NewAsteriskAgent(cfg),
		NewRadiusAgent(cfg),
		NewDiameterAgent(cfg),
		NewHTTPAgent(cfg),
		NewSIPAgent(cfg),
		NewEventExporterService(cfg),
		NewRegistrarCService(cfg),
		NewRateService(cfg),
		NewActionService(cfg),
		NewAccountService(cfg),
		NewTPeService(cfg),
		NewPrometheusAgent(cfg),
	})

	srvManager.StartServices(shutdown)
	RegisterServiceManagerV1(cfg, srvManager, registry, shutdown)

	// Serve rpc connections
	startRPC(cfg, registry, shutdown)

	if flags.profiling.mem.dir != "" {
		if err := profiler.StartMemoryProfiling(cores.MemoryProfilingParams{
			DirPath:      flags.profiling.mem.dir,
			MaxFiles:     flags.profiling.mem.maxFiles,
			Interval:     flags.profiling.mem.interval,
			UseTimestamp: flags.profiling.mem.useTS,
		}); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
	}

	<-shutdown.Done()
	return
}

// writePIDFile creates a file at the specified path containing the current process ID.
func writePIDFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create pid file: %s", err)
	}
	if _, err := f.WriteString(strconv.Itoa(os.Getpid())); err != nil {
		f.Close()
		return fmt.Errorf("failed to write to pid file: %s", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close pid file: %s", err)
	}
	return nil
}

// initConfigFromPath loads and initializes the CGR configuration from the specified path.
func initConfigFromPath(ctx *context.Context, path, nodeID, logType string, logLevel int) (cfg *config.CGRConfig, err error) {
	if cfg, err = config.NewCGRConfigFromPath(ctx, path); err != nil {
		err = fmt.Errorf("could not parse config: <%s>", err)
		return
	}
	if !(cfg.ConfigDBCfg().Type == utils.MetaInternal &&
		cfg.ConfigDBCfg().Opts.InternalDBDumpInterval == 0) { // ignore if offline internal is disabled
		var d config.ConfigDB
		if d, err = engine.NewDBConn(cfg.ConfigDBCfg().Type,
			cfg.ConfigDBCfg().Host, cfg.ConfigDBCfg().Port,
			cfg.ConfigDBCfg().Name, cfg.ConfigDBCfg().User,
			cfg.ConfigDBCfg().Password, cfg.GeneralCfg().DBDataEncoding,
			cfg.ConfigDBCfg().Opts, nil); err != nil { // Cannot configure getter database, show stopper
			err = fmt.Errorf("could not configure configDB: <%s>", err)
			return
		}
		if err = cfg.LoadFromDB(ctx, d); err != nil {
			err = fmt.Errorf("could not parse config from DB: <%s>", err)
			return
		}
	}
	if nodeID != utils.EmptyString {
		cfg.GeneralCfg().NodeID = nodeID
	}
	if logLevel != -1 { // Modify the log level if provided by command arguments
		cfg.LoggerCfg().Level = logLevel
	}
	if logType != utils.EmptyString {
		cfg.LoggerCfg().Type = logType
	}
	if utils.ConcurrentReqsLimit != 0 { // used as shared variable
		cfg.CoreSCfg().Caps = utils.ConcurrentReqsLimit
	}
	if len(utils.ConcurrentReqsStrategy) != 0 {
		cfg.CoreSCfg().CapsStrategy = utils.ConcurrentReqsStrategy
	}
	return
}

// handleSignals manages system signals for graceful shutdown and configuration reload.
func handleSignals(shutdown *utils.SyncedChan, cfg *config.CGRConfig, shdWg *sync.WaitGroup) {
	defer shdWg.Done()
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shutdown.Done():
			return
		case <-shutdownSignal:
			shutdown.CloseOnce()
		case <-reloadSignal:
			//  do it in its own goroutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := cfg.V1ReloadConfig(context.TODO(),
					new(config.ReloadArgs), &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}

// startRPC initializes and starts the RPC server.
func startRPC(cfg *config.CGRConfig, registry *servmanager.Registry, shutdown *utils.SyncedChan) {
	cl := registry.Lookup(utils.CommonListenerS).(*CommonListenerService).CLS()
	cl.StartServer(cfg, shutdown)
}
