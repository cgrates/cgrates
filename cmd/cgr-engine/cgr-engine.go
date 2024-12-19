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
	"github.com/cgrates/cgrates/commonlisteners"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/efs"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/guardian"
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

	// Init config
	ctx, cancel := context.WithCancel(context.Background())
	var cfg *config.CGRConfig
	if cfg, err = services.InitConfigFromPath(ctx, *flags.CfgPath, *flags.NodeID, *flags.LogLevel); err != nil || *flags.CheckConfig {
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
	go handleSignals(ctx, cancel, cfg, shdWg)

	if *flags.ScheduledShutdown != utils.EmptyString {
		var shtDwDur time.Duration
		if shtDwDur, err = utils.ParseDurationWithNanosecs(*flags.ScheduledShutdown); err != nil {
			return
		}
		shdWg.Add(1)
		go func() { // Schedule shutdown
			tm := time.NewTimer(shtDwDur)
			select {
			case <-tm.C:
				cancel()
			case <-ctx.Done():
				tm.Stop()
			}
			shdWg.Done()
		}()
	}

	connMgr := engine.NewConnManager(cfg)
	// init syslog
	if utils.Logger, err = engine.NewLogger(ctx,
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
	iConfigCh := make(chan birpc.ClientConnector, 1)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig), utils.ConfigSv1, iConfigCh)
	iGuardianSCh := make(chan birpc.ClientConnector, 1)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian), utils.GuardianSv1, iGuardianSCh)

	// ServiceIndexer will share service references to all services
	srvIdxr := servmanager.NewServiceIndexer()
	gvS := services.NewGlobalVarS(cfg, srvIdxr)
	dmS := services.NewDataDBService(cfg, connMgr, *flags.SetVersions, srvDep, srvIdxr)
	sdbS := services.NewStorDBService(cfg, *flags.SetVersions, srvIdxr)
	cls := services.NewCommonListenerService(cfg, caps, srvIdxr)
	anzS := services.NewAnalyzerService(cfg, srvIdxr)
	coreS := services.NewCoreService(cfg, caps, cpuPrfF, shdWg, srvIdxr)
	cacheS := services.NewCacheService(cfg, connMgr, srvIdxr)
	fltrS := services.NewFilterService(cfg, connMgr, srvIdxr)
	dspS := services.NewDispatcherService(cfg, connMgr, srvIdxr)
	ldrs := services.NewLoaderService(cfg, connMgr, srvIdxr)
	efs := services.NewExportFailoverService(cfg, connMgr, srvIdxr)
	adminS := services.NewAdminSv1Service(cfg, connMgr, srvIdxr)
	sessionS := services.NewSessionService(cfg, connMgr, srvIdxr)
	attrS := services.NewAttributeService(cfg, dspS, srvIdxr)
	chrgS := services.NewChargerService(cfg, connMgr, srvIdxr)
	routeS := services.NewRouteService(cfg, connMgr, srvIdxr)
	resourceS := services.NewResourceService(cfg, connMgr, srvDep, srvIdxr)
	trendS := services.NewTrendService(cfg, connMgr, srvDep, srvIdxr)
	rankingS := services.NewRankingService(cfg, connMgr, srvDep, srvIdxr)
	thS := services.NewThresholdService(cfg, connMgr, srvDep, srvIdxr)
	stS := services.NewStatService(cfg, connMgr, srvDep, srvIdxr)
	erS := services.NewEventReaderService(cfg, connMgr, srvIdxr)
	dnsAgent := services.NewDNSAgent(cfg, connMgr, srvIdxr)
	fsAgent := services.NewFreeswitchAgent(cfg, connMgr, srvIdxr)
	kamAgent := services.NewKamailioAgent(cfg, connMgr, srvIdxr)
	janusAgent := services.NewJanusAgent(cfg, connMgr, srvIdxr)
	astAgent := services.NewAsteriskAgent(cfg, connMgr, srvIdxr)
	radAgent := services.NewRadiusAgent(cfg, connMgr, srvIdxr)
	diamAgent := services.NewDiameterAgent(cfg, connMgr, caps, srvIdxr)
	httpAgent := services.NewHTTPAgent(cfg, connMgr, srvIdxr)
	sipAgent := services.NewSIPAgent(cfg, connMgr, srvIdxr)
	eeS := services.NewEventExporterService(cfg, connMgr, srvIdxr)
	cdrS := services.NewCDRServer(cfg, connMgr, srvIdxr)
	registrarcS := services.NewRegistrarCService(cfg, connMgr, srvIdxr)
	rateS := services.NewRateService(cfg, srvIdxr)
	actionS := services.NewActionService(cfg, connMgr, srvIdxr)
	accS := services.NewAccountService(cfg, connMgr, srvIdxr)
	tpeS := services.NewTPeService(cfg, connMgr, srvIdxr)

	srvManager := servmanager.NewServiceManager(shdWg, connMgr, cfg, srvIdxr,
		[]servmanager.Service{
			gvS,
			dmS,
			sdbS,
			cls,
			anzS,
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
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown all subsystems in the given time",
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

		utils.Logger.Info("<CoreS> stopped all components. CGRateS shutdown!")
	}()

	shdWg.Add(1)
	if err = gvS.Start(ctx, cancel); err != nil {
		shdWg.Done()
		srvManager.ShutdownServices()
		return
	}
	if cls.ShouldRun() {
		shdWg.Add(1)
		if err = cls.Start(ctx, cancel); err != nil {
			shdWg.Done()
			srvManager.ShutdownServices()
			return
		}
	}
	if efs.ShouldRun() { // efs checking first because of loggers
		shdWg.Add(1)
		if err = efs.Start(ctx, cancel); err != nil {
			shdWg.Done()
			srvManager.ShutdownServices()
			return
		}
	}
	if dmS.ShouldRun() { // Some services can run without db, ie:  ERs
		shdWg.Add(1)
		if err = dmS.Start(ctx, cancel); err != nil {
			shdWg.Done()
			srvManager.ShutdownServices()
			return
		}
	}
	if sdbS.ShouldRun() {
		shdWg.Add(1)
		if err = sdbS.Start(ctx, cancel); err != nil {
			shdWg.Done()
			srvManager.ShutdownServices()
			return
		}
	}

	if anzS.ShouldRun() {
		shdWg.Add(1)
		if err = anzS.Start(ctx, cancel); err != nil {
			shdWg.Done()
			srvManager.ShutdownServices()
			return
		}
	} else {
		close(anzS.StateChan(utils.StateServiceUP))
	}

	shdWg.Add(1)
	if err = coreS.Start(ctx, cancel); err != nil {
		shdWg.Done()
		srvManager.ShutdownServices()
		return
	}
	shdWg.Add(1)
	if err = cacheS.Start(ctx, cancel); err != nil {
		shdWg.Done()
		srvManager.ShutdownServices()
		return
	}
	srvManager.StartServices(ctx, cancel)

	cgrInitServiceManagerV1(iServeManagerCh, srvManager, cfg, cls.CLS(), anzS)
	cgrInitGuardianSv1(iGuardianSCh, cfg, cls.CLS(), anzS)
	cgrInitConfigSv1(iConfigCh, cfg, cls.CLS(), anzS)

	if *flags.Preload != utils.EmptyString {
		if err = cgrRunPreload(ctx, cfg, *flags.Preload, srvIdxr); err != nil {
			return
		}
	}

	// Serve rpc connections
	cgrStartRPC(ctx, cancel, cfg, srvIdxr)

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

	<-ctx.Done()
	return
}

func cgrRunPreload(ctx *context.Context, cfg *config.CGRConfig, loaderIDs string,
	sIdxr *servmanager.ServiceIndexer) (err error) {
	if !cfg.LoaderCfg().Enabled() {
		err = fmt.Errorf("<%s> not enabled but required by preload mechanism", utils.LoaderS)
		return
	}
	loader := sIdxr.GetService(utils.LoaderS).(*services.LoaderService)
	select {
	case <-loader.StateChan(utils.StateServiceUP):
	case <-ctx.Done():
		return
	}

	var reply string
	for _, loaderID := range strings.Split(loaderIDs, utils.FieldsSep) {
		if err = loader.GetLoaderS().V1Run(ctx, &loaders.ArgsProcessFolder{
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

func cgrInitGuardianSv1(iGuardianSCh chan birpc.ClientConnector, cfg *config.CGRConfig,
	cl *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	srv, _ := engine.NewServiceWithName(guardian.Guardian, utils.GuardianS, true)
	if !cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cl.RpcRegister(s)
		}
	}
	iGuardianSCh <- anz.GetInternalCodec(srv, utils.GuardianS)
}

func cgrInitServiceManagerV1(iServMngrCh chan birpc.ClientConnector,
	srvMngr *servmanager.ServiceManager, cfg *config.CGRConfig,
	cl *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	srv, _ := birpc.NewService(apis.NewServiceManagerV1(srvMngr), utils.EmptyString, false)
	if !cfg.DispatcherSCfg().Enabled {
		cl.RpcRegister(srv)
	}
	iServMngrCh <- anz.GetInternalCodec(srv, utils.ServiceManager)
}

func cgrInitConfigSv1(iConfigCh chan birpc.ClientConnector,
	cfg *config.CGRConfig, cl *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	srv, _ := engine.NewServiceWithName(cfg, utils.ConfigS, true)
	// srv, _ := birpc.NewService(apis.NewConfigSv1(cfg), "", false)
	if !cfg.DispatcherSCfg().Enabled {
		for _, s := range srv {
			cl.RpcRegister(s)
		}
	}
	iConfigCh <- anz.GetInternalCodec(srv, utils.ConfigSv1)
}

func cgrStartRPC(ctx *context.Context, shtdwnEngine context.CancelFunc,
	cfg *config.CGRConfig, sIdxr *servmanager.ServiceIndexer) {
	if cfg.DispatcherSCfg().Enabled { // wait only for dispatcher as cache is allways registered before this
		select {
		case <-sIdxr.GetService(utils.DispatcherS).StateChan(utils.StateServiceUP):
		case <-ctx.Done():
			return
		}
	}
	cl := sIdxr.GetService(utils.CommonListenerS).(*services.CommonListenerService).CLS()
	cl.StartServer(ctx, shtdwnEngine, cfg)
}

func handleSignals(ctx *context.Context, shutdown context.CancelFunc,
	cfg *config.CGRConfig, shdWg *sync.WaitGroup) {
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			shdWg.Done()
			return
		case <-shutdownSignal:
			shutdown()
			shdWg.Done()
			return
		case <-reloadSignal:
			//  do it in its own goroutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := cfg.V1ReloadConfig(ctx,
					new(config.ReloadArgs), &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}
