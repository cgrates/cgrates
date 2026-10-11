// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package services

import (
	"errors"
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

// RunCGREngine loads configuration into cfg and runs the engine.
// cfg must already contain defaults and any extra sections.
// Extra services must use the loaded configuration in Start, not in their constructors.
func RunCGREngine(fs []string, cfg *config.CGRConfig, extraServices ...servmanager.Service) error {
	flags := newFlags()
	_ = flags.Parse(fs)

	vers, err := utils.GetCGRVersion()
	if err != nil {
		return err
	}
	if flags.config.version {
		fmt.Println(vers)
		return nil
	}
	if flags.process.pidFile != "" {
		pid := strconv.Itoa(os.Getpid())
		if err := os.WriteFile(flags.process.pidFile, []byte(pid), 0666); err != nil {
			return err
		}
		defer func() {
			if err := os.Remove(flags.process.pidFile); err != nil {
				utils.Logger.Warning("Could not remove pid file: " + err.Error())
			}
		}()
	}
	if flags.process.singleCPU {
		runtime.GOMAXPROCS(1)
	}

	if err := loadConfig(context.TODO(), cfg, flags); err != nil {
		return err
	}
	if flags.config.check {
		return nil
	}

	if cfg.LoggerCfg().Level >= 0 {
		switch cfg.LoggerCfg().Type {
		case utils.MetaSysLog:
			logger, err := utils.NewSysLogger(cfg.GeneralCfg().NodeID, cfg.LoggerCfg().Level)
			if err != nil {
				return err
			}
			utils.Logger = logger
		case utils.MetaStdLog, utils.MetaKafkaLog:
			// Kafka logging needs ConnManager, so use stdout during startup.
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

	shutdown := utils.NewSyncedChan()
	shdWg.Go(func() {
		handleSignals(shutdown, cfg)
	})

	if flags.process.scheduledShutdown != 0 {
		shdWg.Go(func() {
			select {
			case <-time.After(flags.process.scheduledShutdown):
				shutdown.CloseOnce()
			case <-shutdown.Done():
			}
		})
	}

	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, runtime.Version()))

	registry := servmanager.NewRegistry()
	cl := NewCommonListenerService(cfg)
	services := []servmanager.Service{
		NewGlobalVarS(cfg),
		NewCapService(cfg),
		cl,
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
	}
	srvManager := servmanager.NewServiceManager(shdWg, cfg, registry, append(services, extraServices...))

	srvManager.StartServices(shutdown)
	if err := RegisterServiceManagerV1(cfg, srvManager, registry, shutdown); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		shutdown.CloseOnce()
		return err
	}

	select {
	case <-shutdown.Done():
		return nil
	default:
	}
	cl.CLS().StartServer(cfg, shutdown)

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
	return nil
}

func loadConfig(ctx *context.Context, cfg *config.CGRConfig, flags *flags) error {
	if err := cfg.LoadFromPath(ctx, flags.config.path); err != nil {
		return fmt.Errorf("could not parse config: <%s>", err)
	}
	dbCfg := cfg.ConfigDBCfg()
	// Internal configDB is loaded only when dumping is enabled.
	if dbCfg.Type != utils.MetaInternal || dbCfg.Opts.InternalDBDumpInterval != 0 {
		d, err := engine.NewDBConn(dbCfg.Type,
			dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.User,
			dbCfg.Password, cfg.GeneralCfg().DBDataEncoding, dbCfg.Opts, nil)
		if err != nil {
			return fmt.Errorf("could not configure configDB: <%s>", err)
		}
		if err := cfg.LoadFromDB(ctx, d); err != nil {
			return fmt.Errorf("could not parse config from DB: <%s>", err)
		}
	}
	if flags.logger.nodeID != "" {
		cfg.GeneralCfg().NodeID = flags.logger.nodeID
	}
	if flags.logger.level != -1 {
		cfg.LoggerCfg().Level = flags.logger.level
	}
	if flags.logger.typ != "" {
		cfg.LoggerCfg().Type = flags.logger.typ
	}
	if utils.ConcurrentReqsLimit != 0 {
		cfg.CoreSCfg().Caps = utils.ConcurrentReqsLimit
	}
	if utils.ConcurrentReqsStrategy != "" {
		cfg.CoreSCfg().CapsStrategy = utils.ConcurrentReqsStrategy
	}
	return nil
}

func handleSignals(shutdown *utils.SyncedChan, cfg *config.CGRConfig) {
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shutdown.Done():
			return
		case <-shutdownSignal:
			shutdown.CloseOnce()
		case <-reloadSignal:
			// Do not block shutdown signals while reloading configuration.
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
