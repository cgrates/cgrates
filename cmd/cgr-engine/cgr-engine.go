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
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/apis"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
)

func main() {
	if err := runCGREngine(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

// runCGREngine configures the CGREngine object and runs it
func runCGREngine(fs []string) (err error) {
	flags := services.NewCGREngineFlags()
	flags.Parse(fs)
	var vers string
	if vers, err = utils.GetCGRVersion(); err != nil {
		return
	}
	if *flags.Version {
		fmt.Println(vers)
		return
	}
	if *flags.PidFile != utils.EmptyString {
		if err = services.CgrWritePid(*flags.PidFile); err != nil {
			return
		}
	}
	if *flags.SingleCPU {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	var cfg *config.CGRConfig
	if cfg, err = services.InitConfigFromPath(context.TODO(), *flags.CfgPath, *flags.NodeID,
		*flags.Logger, *flags.LogLevel); err != nil || *flags.CheckConfig {
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
	if *flags.CpuPrfDir != utils.EmptyString {
		cpuPath := filepath.Join(*flags.CpuPrfDir, utils.CpuPathCgr)
		if cpuPrfF, err = cores.StartCPUProfiling(cpuPath); err != nil {
			return
		}
	}

	shdWg := new(sync.WaitGroup)
	shdWg.Add(1)
	shutdown := utils.NewSyncedChan()
	go handleSignals(shutdown, cfg, shdWg)

	if *flags.ScheduledShutdown != utils.EmptyString {
		var shtDwDur time.Duration
		if shtDwDur, err = utils.ParseDurationWithNanosecs(*flags.ScheduledShutdown); err != nil {
			return
		}
		shdWg.Add(1)
		go func() { // Schedule shutdown
			defer shdWg.Done()
			tm := time.NewTimer(shtDwDur)
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
	gvS := services.NewGlobalVarS(cfg)
	caps := services.NewCapService(cfg)
	cls := services.NewCommonListenerService(cfg)
	anzS := services.NewAnalyzerService(cfg)
	cms := services.NewConnManagerService(cfg)
	lgs := services.NewLoggerService(cfg, *flags.Logger)
	dmS := services.NewDataDBService(cfg, *flags.SetVersions)
	sdbS := services.NewStorDBService(cfg, *flags.SetVersions)
	configS := services.NewConfigService(cfg)
	guardianS := services.NewGuardianService(cfg)
	coreS := services.NewCoreService(cfg, cpuPrfF, shdWg)
	cacheS := services.NewCacheService(cfg)
	fltrS := services.NewFilterService(cfg)
	dspS := services.NewDispatcherService(cfg)
	ldrs := services.NewLoaderService(cfg)
	efs := services.NewExportFailoverService(cfg)
	adminS := services.NewAdminSv1Service(cfg)
	sessionS := services.NewSessionService(cfg)
	attrS := services.NewAttributeService(cfg, dspS)
	chrgS := services.NewChargerService(cfg)
	routeS := services.NewRouteService(cfg)
	resourceS := services.NewResourceService(cfg)
	trendS := services.NewTrendService(cfg)
	rankingS := services.NewRankingService(cfg)
	thS := services.NewThresholdService(cfg)
	stS := services.NewStatService(cfg)
	erS := services.NewEventReaderService(cfg)
	dnsAgent := services.NewDNSAgent(cfg)
	fsAgent := services.NewFreeswitchAgent(cfg)
	kamAgent := services.NewKamailioAgent(cfg)
	janusAgent := services.NewJanusAgent(cfg)
	astAgent := services.NewAsteriskAgent(cfg)
	radAgent := services.NewRadiusAgent(cfg)
	diamAgent := services.NewDiameterAgent(cfg)
	httpAgent := services.NewHTTPAgent(cfg)
	sipAgent := services.NewSIPAgent(cfg)
	eeS := services.NewEventExporterService(cfg)
	cdrS := services.NewCDRServer(cfg)
	registrarcS := services.NewRegistrarCService(cfg)
	rateS := services.NewRateService(cfg)
	actionS := services.NewActionService(cfg)
	accS := services.NewAccountService(cfg)
	tpeS := services.NewTPeService(cfg)

	srvManager := servmanager.NewServiceManager(shdWg, cfg, registry, []servmanager.Service{
		gvS,
		caps,
		cls,
		anzS,
		cms,
		lgs,
		dmS,
		sdbS,
		configS,
		guardianS,
		coreS,
		cacheS,
		fltrS,
		dspS,
		ldrs,
		efs,
		adminS,
		sessionS,
		attrS,
		chrgS,
		routeS,
		resourceS,
		trendS,
		rankingS,
		thS,
		stS,
		erS,
		dnsAgent,
		fsAgent,
		kamAgent,
		janusAgent,
		astAgent,
		radAgent,
		diamAgent,
		httpAgent,
		sipAgent,
		eeS,
		cdrS,
		registrarcS,
		rateS,
		actionS,
		accS,
		tpeS,
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
		if *flags.PidFile != utils.EmptyString {
			if err := os.Remove(*flags.PidFile); err != nil {
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
	cgrInitServiceManagerV1(cfg, srvManager, registry)

	if *flags.Preload != utils.EmptyString {
		if err = cgrRunPreload(cfg, *flags.Preload, registry); err != nil {
			return
		}
	}

	// Serve rpc connections
	cgrStartRPC(cfg, registry, shutdown)

	// TODO: find a better location for this if block
	if *flags.MemPrfDir != "" {
		if err := coreS.CoreS().StartMemoryProfiling(cores.MemoryProfilingParams{
			DirPath:      *flags.MemPrfDir,
			MaxFiles:     *flags.MemPrfMaxF,
			Interval:     *flags.MemPrfInterval,
			UseTimestamp: *flags.MemPrfTS,
		}); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
	}

	<-shutdown.Done()
	return
}

// TODO: merge with LoaderService
func cgrRunPreload(cfg *config.CGRConfig, loaderIDs string,
	registry *servmanager.ServiceRegistry) (err error) {
	if !cfg.LoaderCfg().Enabled() {
		err = fmt.Errorf("<%s> not enabled but required by preload mechanism", utils.LoaderS)
		return
	}
	loader := registry.Lookup(utils.LoaderS).(*services.LoaderService)
	if utils.StructChanTimeout(loader.StateChan(utils.StateServiceUP), cfg.GeneralCfg().ConnectTimeout) {
		return utils.NewServiceStateTimeoutError(utils.PreloadCgr, utils.LoaderS, utils.StateServiceUP)
	}
	var reply string
	for _, loaderID := range strings.Split(loaderIDs, utils.FieldsSep) {
		if err = loader.GetLoaderS().V1Run(context.TODO(), &loaders.ArgsProcessFolder{
			APIOpts: map[string]any{
				utils.MetaForceLock:   true, // force lock will unlock the file in case is locked and return error
				utils.MetaStopOnError: true,
			},
			LoaderID: loaderID,
		}, &reply); err != nil {
			err = fmt.Errorf("<%s> preload failed on loadID <%s> , err: <%s>", utils.LoaderS, loaderID, err)
			return
		}
	}
	return
}

func cgrInitServiceManagerV1(cfg *config.CGRConfig, srvMngr *servmanager.ServiceManager,
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
	if !cfg.DispatcherSCfg().Enabled {
		cl.RpcRegister(srv)
	}
	cms.AddInternalConn(utils.ServiceManager, srv)
}

func cgrStartRPC(cfg *config.CGRConfig, registry *servmanager.ServiceRegistry, shutdown *utils.SyncedChan) {
	if cfg.DispatcherSCfg().Enabled { // wait only for dispatcher as cache is allways registered before this
		if utils.StructChanTimeout(
			registry.Lookup(utils.DispatcherS).StateChan(utils.StateServiceUP),
			cfg.GeneralCfg().ConnectTimeout) {
			return
		}
	}
	cl := registry.Lookup(utils.CommonListenerS).(*services.CommonListenerService).CLS()
	cl.StartServer(cfg, shutdown)
}

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
