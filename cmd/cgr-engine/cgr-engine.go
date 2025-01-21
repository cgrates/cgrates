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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func main() {
	if err := runCGREngine(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// flags holds all command line arguments. Flag descriptions are set in newFlags.
type flags struct {
	*flag.FlagSet
	config struct {
		path    string
		check   bool
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

// runCGREngine initializes and runs the CGREngine with the provided command line arguments.
func runCGREngine(fs []string) (err error) {
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

	var cpuPrfF *os.File
	if flags.profiling.cpu.dir != utils.EmptyString {
		cpuPath := filepath.Join(flags.profiling.cpu.dir, utils.CpuPathCgr)
		if cpuPrfF, err = cores.StartCPUProfiling(cpuPath); err != nil {
			return
		}
	}

	shdWg := new(sync.WaitGroup)
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

	// ServiceIndexer will share service references to all services
	registry := servmanager.NewServiceRegistry()
	coreS := services.NewCoreService(cfg, cpuPrfF, shdWg)

	srvManager := servmanager.NewServiceManager(shdWg, cfg, registry, []servmanager.Service{
		services.NewGlobalVarS(cfg),
		services.NewCapService(cfg),
		services.NewCommonListenerService(cfg),
		services.NewAnalyzerService(cfg),
		services.NewConnManagerService(cfg),
		services.NewLoggerService(cfg, flags.logger.typ),
		services.NewDataDBService(cfg, flags.data.setVersions),
		services.NewStorDBService(cfg, flags.data.setVersions),
		services.NewConfigService(cfg),
		services.NewGuardianService(cfg),
		coreS,
		services.NewCacheService(cfg),
		services.NewFilterService(cfg),
		services.NewLoaderService(cfg, flags.data.preloadIDs),
		services.NewExportFailoverService(cfg),
		services.NewAdminSv1Service(cfg),
		services.NewSessionService(cfg),
		services.NewAttributeService(cfg),
		services.NewChargerService(cfg),
		services.NewRouteService(cfg),
		services.NewResourceService(cfg),
		services.NewTrendService(cfg),
		services.NewRankingService(cfg),
		services.NewThresholdService(cfg),
		services.NewStatService(cfg),
		services.NewEventReaderService(cfg),
		services.NewDNSAgent(cfg),
		services.NewFreeswitchAgent(cfg),
		services.NewKamailioAgent(cfg),
		services.NewJanusAgent(cfg),
		services.NewAsteriskAgent(cfg),
		services.NewRadiusAgent(cfg),
		services.NewDiameterAgent(cfg),
		services.NewHTTPAgent(cfg),
		services.NewSIPAgent(cfg),
		services.NewEventExporterService(cfg),
		services.NewCDRServer(cfg),
		services.NewRegistrarCService(cfg),
		services.NewRateService(cfg),
		services.NewActionService(cfg),
		services.NewAccountService(cfg),
		services.NewTPeService(cfg),
	})

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.CoreSCfg().ShutdownTimeout*10)
		go func() {
			shdWg.Wait()
			cancel()
		}()
		<-ctx.Done()
		if ctx.Err() != context.Canceled {
			utils.Logger.Err(fmt.Sprintf("<%s> failed to shut down all services in the given time",
				utils.ServiceManager))
		}
		if flags.process.pidFile != utils.EmptyString {
			if err := os.Remove(flags.process.pidFile); err != nil {
				utils.Logger.Warning("Could not remove pid file: " + err.Error())
			}
		}
		if cpuPrfF != nil && coreS == nil {
			pprof.StopCPUProfile()
			if err := cpuPrfF.Close(); err != nil {
				utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
			}
		}

		// TODO: check if there's any need to manually stop memory profiling.
		// It should be stopped automatically during CoreS service shutdown.

		utils.Logger.Info(fmt.Sprintf("<%s> stopped all services. CGRateS shutdown!", utils.ServiceManager))
	}()

	srvManager.StartServices(shutdown)
	initServiceManagerV1(cfg, srvManager, registry)

	// Serve rpc connections
	startRPC(cfg, registry, shutdown)

	// TODO: find a better location for this if block
	if flags.profiling.mem.dir != "" {
		if err := coreS.CoreS().StartMemoryProfiling(cores.MemoryProfilingParams{
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
	if cfg.ConfigDBCfg().Type != utils.MetaInternal {
		var d config.ConfigDB
		if d, err = engine.NewDataDBConn(cfg.ConfigDBCfg().Type,
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
	config.SetCgrConfig(cfg) // Share the config object
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

// initServiceManagerV1 registers the ServiceManager methods.
func initServiceManagerV1(cfg *config.CGRConfig, srvMngr *servmanager.ServiceManager,
	registry *servmanager.ServiceRegistry) {
	srvDeps, err := services.WaitForServicesToReachState(utils.StateServiceUP,
		[]string{
			utils.CommonListenerS,
			utils.ConnManager,
		},
		registry, cfg.GeneralCfg().ConnectTimeout)
	if err != nil {
		return
	}
	cl := srvDeps[utils.CommonListenerS].(*services.CommonListenerService).CLS()
	cms := srvDeps[utils.ConnManager].(*services.ConnManagerService)
	srv, _ := birpc.NewService(apis.NewServiceManagerV1(srvMngr), utils.EmptyString, false)
	cl.RpcRegister(srv)
	cms.AddInternalConn(utils.ServiceManager, srv)
}

// startRPC initializes and starts the RPC server.
func startRPC(cfg *config.CGRConfig, registry *servmanager.ServiceRegistry, shutdown *utils.SyncedChan) {
	cl := registry.Lookup(utils.CommonListenerS).(*services.CommonListenerService).CLS()
	cl.StartServer(cfg, shutdown)
}
