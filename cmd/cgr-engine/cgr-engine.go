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
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
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
	if cfg, err = services.InitConfigFromPath(context.TODO(), *flags.CfgPath, *flags.NodeID, *flags.LogLevel); err != nil || *flags.CheckConfig {
		return
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
	shutdown := make(chan struct{})
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
				close(shutdown)
			case <-shutdown:
				tm.Stop()
			}
		}()
	}

	connMgr := engine.NewConnManager(cfg)
	// init syslog
	if utils.Logger, err = engine.NewLogger(context.TODO(),
		utils.FirstNonEmpty(*flags.Logger, cfg.LoggerCfg().Type),
		cfg.GeneralCfg().DefaultTenant,
		cfg.GeneralCfg().NodeID,
		connMgr, cfg); err != nil {
		return fmt.Errorf("Could not initialize syslog connection, err: <%s>", err)
	}
	efs.SetFailedPostCacheTTL(cfg.EFsCfg().FailedPostsTTL) // init failedPosts to posts loggers/exporters in case of failing
	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, runtime.Version()))

	caps := engine.NewCaps(cfg.CoreSCfg().Caps, cfg.CoreSCfg().CapsStrategy)
	srvDep := map[string]*sync.WaitGroup{
		utils.DataDB: new(sync.WaitGroup),
	}

	iServeManagerCh := make(chan birpc.ClientConnector, 1)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager), utils.ServiceManagerV1, iServeManagerCh)

	// ServiceIndexer will share service references to all services
	registry := servmanager.NewServiceRegistry()
	gvS := services.NewGlobalVarS(cfg, registry)
	dmS := services.NewDataDBService(cfg, connMgr, *flags.SetVersions, srvDep, registry)
	sdbS := services.NewStorDBService(cfg, *flags.SetVersions, registry)
	cls := services.NewCommonListenerService(cfg, caps, registry)
	anzS := services.NewAnalyzerService(cfg, registry)
	configS := services.NewConfigService(cfg, registry)
	guardianS := services.NewGuardianService(cfg, registry)
	coreS := services.NewCoreService(cfg, caps, cpuPrfF, shdWg, registry)
	cacheS := services.NewCacheService(cfg, connMgr, registry)
	fltrS := services.NewFilterService(cfg, connMgr, registry)
	dspS := services.NewDispatcherService(cfg, connMgr, registry)
	ldrs := services.NewLoaderService(cfg, connMgr, registry)
	efs := services.NewExportFailoverService(cfg, connMgr, registry)
	adminS := services.NewAdminSv1Service(cfg, connMgr, registry)
	sessionS := services.NewSessionService(cfg, connMgr, registry)
	attrS := services.NewAttributeService(cfg, dspS, registry)
	chrgS := services.NewChargerService(cfg, connMgr, registry)
	routeS := services.NewRouteService(cfg, connMgr, registry)
	resourceS := services.NewResourceService(cfg, connMgr, srvDep, registry)
	trendS := services.NewTrendService(cfg, connMgr, srvDep, registry)
	rankingS := services.NewRankingService(cfg, connMgr, srvDep, registry)
	thS := services.NewThresholdService(cfg, connMgr, srvDep, registry)
	stS := services.NewStatService(cfg, connMgr, srvDep, registry)
	erS := services.NewEventReaderService(cfg, connMgr, registry)
	dnsAgent := services.NewDNSAgent(cfg, connMgr, registry)
	fsAgent := services.NewFreeswitchAgent(cfg, connMgr, registry)
	kamAgent := services.NewKamailioAgent(cfg, connMgr, registry)
	janusAgent := services.NewJanusAgent(cfg, connMgr, registry)
	astAgent := services.NewAsteriskAgent(cfg, connMgr, registry)
	radAgent := services.NewRadiusAgent(cfg, connMgr, registry)
	diamAgent := services.NewDiameterAgent(cfg, connMgr, caps, registry)
	httpAgent := services.NewHTTPAgent(cfg, connMgr, registry)
	sipAgent := services.NewSIPAgent(cfg, connMgr, registry)
	eeS := services.NewEventExporterService(cfg, connMgr, registry)
	cdrS := services.NewCDRServer(cfg, connMgr, registry)
	registrarcS := services.NewRegistrarCService(cfg, connMgr, registry)
	rateS := services.NewRateService(cfg, registry)
	actionS := services.NewActionService(cfg, connMgr, registry)
	accS := services.NewAccountService(cfg, connMgr, registry)
	tpeS := services.NewTPeService(cfg, connMgr, registry)

	srvManager := servmanager.NewServiceManager(shdWg, connMgr, cfg, registry, []servmanager.Service{
		gvS,
		dmS,
		sdbS,
		cls,
		anzS,
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
	cgrInitServiceManagerV1(iServeManagerCh, cfg, srvManager, registry)

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

	<-shutdown
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

func cgrInitServiceManagerV1(iServMngrCh chan birpc.ClientConnector, cfg *config.CGRConfig,
	srvMngr *servmanager.ServiceManager, registry *servmanager.ServiceRegistry) {
	cls := registry.Lookup(utils.CommonListenerS).(*services.CommonListenerService)
	if utils.StructChanTimeout(cls.StateChan(utils.StateServiceUP), cfg.GeneralCfg().ConnectTimeout) {
		return
	}
	cl := cls.CLS()
	anz := registry.Lookup(utils.AnalyzerS).(*services.AnalyzerService)
	if utils.StructChanTimeout(anz.StateChan(utils.StateServiceUP), cfg.GeneralCfg().ConnectTimeout) {
		return
	}
	srv, _ := birpc.NewService(apis.NewServiceManagerV1(srvMngr), utils.EmptyString, false)
	if !cfg.DispatcherSCfg().Enabled {
		cl.RpcRegister(srv)
	}
	iServMngrCh <- anz.GetInternalCodec(srv, utils.ServiceManager)
}

func cgrStartRPC(cfg *config.CGRConfig, registry *servmanager.ServiceRegistry, shutdown chan struct{}) {
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

func handleSignals(stopChan chan struct{}, cfg *config.CGRConfig, shdWg *sync.WaitGroup) {
	defer shdWg.Done()
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-stopChan:
			return
		case <-shutdownSignal:
			close(stopChan)
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
