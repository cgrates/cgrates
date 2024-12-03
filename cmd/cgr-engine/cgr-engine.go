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
	"github.com/cgrates/rpcclient"
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
		utils.AccountS:        new(sync.WaitGroup),
		utils.ActionS:         new(sync.WaitGroup),
		utils.AdminS:          new(sync.WaitGroup),
		utils.AnalyzerS:       new(sync.WaitGroup),
		utils.AsteriskAgent:   new(sync.WaitGroup),
		utils.AttributeS:      new(sync.WaitGroup),
		utils.CDRServer:       new(sync.WaitGroup),
		utils.ChargerS:        new(sync.WaitGroup),
		utils.CoreS:           new(sync.WaitGroup),
		utils.DataDB:          new(sync.WaitGroup),
		utils.DiameterAgent:   new(sync.WaitGroup),
		utils.DispatcherS:     new(sync.WaitGroup),
		utils.DNSAgent:        new(sync.WaitGroup),
		utils.EEs:             new(sync.WaitGroup),
		utils.EFs:             new(sync.WaitGroup),
		utils.ERs:             new(sync.WaitGroup),
		utils.FreeSWITCHAgent: new(sync.WaitGroup),
		utils.GlobalVarS:      new(sync.WaitGroup),
		utils.HTTPAgent:       new(sync.WaitGroup),
		utils.KamailioAgent:   new(sync.WaitGroup),
		utils.LoaderS:         new(sync.WaitGroup),
		utils.RadiusAgent:     new(sync.WaitGroup),
		utils.RateS:           new(sync.WaitGroup),
		utils.RegistrarC:      new(sync.WaitGroup),
		utils.ResourceS:       new(sync.WaitGroup),
		utils.RouteS:          new(sync.WaitGroup),
		utils.SchedulerS:      new(sync.WaitGroup),
		utils.SessionS:        new(sync.WaitGroup),
		utils.SIPAgent:        new(sync.WaitGroup),
		utils.StatS:           new(sync.WaitGroup),
		utils.TrendS:          new(sync.WaitGroup),
		utils.StorDB:          new(sync.WaitGroup),
		utils.ThresholdS:      new(sync.WaitGroup),
		utils.TPeS:            new(sync.WaitGroup),
	}

	// init the channel here because we need to pass them to connManager
	iServeManagerCh := make(chan birpc.ClientConnector, 1)
	iConfigCh := make(chan birpc.ClientConnector, 1)
	iCoreSv1Ch := make(chan birpc.ClientConnector, 1)
	iCacheSCh := make(chan birpc.ClientConnector, 1)
	iGuardianSCh := make(chan birpc.ClientConnector, 1)
	iAnalyzerSCh := make(chan birpc.ClientConnector, 1)
	iCDRServerCh := make(chan birpc.ClientConnector, 1)
	iAttributeSCh := make(chan birpc.ClientConnector, 1)
	iDispatcherSCh := make(chan birpc.ClientConnector, 1)
	iSessionSCh := make(chan birpc.ClientConnector, 1)
	iChargerSCh := make(chan birpc.ClientConnector, 1)
	iThresholdSCh := make(chan birpc.ClientConnector, 1)
	iStatSCh := make(chan birpc.ClientConnector, 1)
	iTrendSCh := make(chan birpc.ClientConnector, 1)
	iRankingSCh := make(chan birpc.ClientConnector, 1)
	iResourceSCh := make(chan birpc.ClientConnector, 1)
	iRouteSCh := make(chan birpc.ClientConnector, 1)
	iAdminSCh := make(chan birpc.ClientConnector, 1)
	iLoaderSCh := make(chan birpc.ClientConnector, 1)
	iEEsCh := make(chan birpc.ClientConnector, 1)
	iRateSCh := make(chan birpc.ClientConnector, 1)
	iActionSCh := make(chan birpc.ClientConnector, 1)
	iAccountSCh := make(chan birpc.ClientConnector, 1)
	iTpeSCh := make(chan birpc.ClientConnector, 1)
	iEFsCh := make(chan birpc.ClientConnector, 1)
	iERsCh := make(chan birpc.ClientConnector, 1)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzerS), utils.AnalyzerSv1, iAnalyzerSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, iAdminSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, iAttributeSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, iCacheSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), utils.CDRsV1, iCDRServerCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), utils.ChargerSv1, iChargerSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian), utils.GuardianSv1, iGuardianSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders), utils.LoaderSv1, iLoaderSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, iResourceSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS), utils.SessionSv1, iSessionSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), utils.SessionSv1, iSessionSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, iStatSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings), utils.RankingSv1, iRankingSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends), utils.TrendSv1, iTrendSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes), utils.RouteSv1, iRouteSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, iThresholdSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager), utils.ServiceManagerV1, iServeManagerCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig), utils.ConfigSv1, iConfigCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore), utils.CoreSv1, iCoreSv1Ch)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), utils.EeSv1, iEEsCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRates), utils.RateSv1, iRateSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers), utils.DispatcherSv1, iDispatcherSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, iAccountSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), utils.ActionSv1, iActionSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTpes), utils.TPeSv1, iTpeSCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEFs), utils.EfSv1, iEFsCh)
	connMgr.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaERs), utils.ErSv1, iERsCh)

	clsCh := make(chan *commonlisteners.CommonListenerS, 1)
	anzCh := make(chan *services.AnalyzerService, 1)
	iFilterSCh := make(chan *engine.FilterS, 1)

	// ServiceIndexer will share service references to all services
	srvIdxr := servmanager.NewServiceIndexer()
	gvS := services.NewGlobalVarS(cfg, srvDep, srvIdxr)
	dmS := services.NewDataDBService(cfg, connMgr, *flags.SetVersions, srvDep, srvIdxr)
	sdbS := services.NewStorDBService(cfg, *flags.SetVersions, srvDep, srvIdxr)
	cls := services.NewCommonListenerService(cfg, caps, clsCh, srvDep, srvIdxr)
	anzS := services.NewAnalyzerService(cfg, clsCh, iFilterSCh, iAnalyzerSCh, anzCh, srvDep, srvIdxr)
	coreS := services.NewCoreService(cfg, caps, clsCh, iCoreSv1Ch, anzCh, cpuPrfF, shdWg, srvDep, srvIdxr)
	cacheS := services.NewCacheService(cfg, dmS, connMgr, clsCh, iCacheSCh, anzCh, coreS, srvDep, srvIdxr)
	dspS := services.NewDispatcherService(cfg, dmS, cacheS, iFilterSCh, clsCh, iDispatcherSCh, connMgr, anzCh, srvDep, srvIdxr)
	ldrs := services.NewLoaderService(cfg, dmS, iFilterSCh, clsCh, iLoaderSCh, connMgr, anzCh, srvDep, srvIdxr)
	efs := services.NewExportFailoverService(cfg, connMgr, iEFsCh, clsCh, srvDep, srvIdxr)
	adminS := services.NewAdminSv1Service(cfg, dmS, sdbS, iFilterSCh, clsCh, iAdminSCh, connMgr, anzCh, srvDep, srvIdxr)
	sessionS := services.NewSessionService(cfg, dmS, iFilterSCh, clsCh, iSessionSCh, connMgr, anzCh, srvDep, srvIdxr)
	attrS := services.NewAttributeService(cfg, dmS, cacheS, iFilterSCh, clsCh, iAttributeSCh, anzCh, dspS, srvDep, srvIdxr)
	chrgS := services.NewChargerService(cfg, dmS, cacheS, iFilterSCh, clsCh, iChargerSCh, connMgr, anzCh, srvDep, srvIdxr)
	routeS := services.NewRouteService(cfg, dmS, cacheS, iFilterSCh, clsCh, iRouteSCh, connMgr, anzCh, srvDep, srvIdxr)
	resourceS := services.NewResourceService(cfg, dmS, cacheS, iFilterSCh, clsCh, iResourceSCh, connMgr, anzCh, srvDep, srvIdxr)
	trendS := services.NewTrendService(cfg, dmS, cacheS, iFilterSCh, clsCh, iTrendSCh, connMgr, anzCh, srvDep, srvIdxr)
	rankingS := services.NewRankingService(cfg, dmS, cacheS, iFilterSCh, clsCh, iRankingSCh, connMgr, anzCh, srvDep, srvIdxr)
	thS := services.NewThresholdService(cfg, dmS, cacheS, iFilterSCh, connMgr, clsCh, iThresholdSCh, anzCh, srvDep, srvIdxr)
	stS := services.NewStatService(cfg, dmS, cacheS, iFilterSCh, clsCh, iStatSCh, connMgr, anzCh, srvDep, srvIdxr)
	erS := services.NewEventReaderService(cfg, iFilterSCh, connMgr, clsCh, iERsCh, anzCh, srvDep, srvIdxr)
	dnsAgent := services.NewDNSAgent(cfg, iFilterSCh, connMgr, srvDep, srvIdxr)
	fsAgent := services.NewFreeswitchAgent(cfg, connMgr, srvDep, srvIdxr)
	kamAgent := services.NewKamailioAgent(cfg, connMgr, srvDep, srvIdxr)
	janusAgent := services.NewJanusAgent(cfg, iFilterSCh, clsCh, connMgr, srvDep, srvIdxr)
	astAgent := services.NewAsteriskAgent(cfg, connMgr, srvDep, srvIdxr)
	radAgent := services.NewRadiusAgent(cfg, iFilterSCh, connMgr, srvDep, srvIdxr)
	diamAgent := services.NewDiameterAgent(cfg, iFilterSCh, connMgr, caps, srvDep, srvIdxr)
	httpAgent := services.NewHTTPAgent(cfg, iFilterSCh, clsCh, connMgr, srvDep, srvIdxr)
	sipAgent := services.NewSIPAgent(cfg, iFilterSCh, connMgr, srvDep, srvIdxr)
	eeS := services.NewEventExporterService(cfg, iFilterSCh, connMgr, clsCh, iEEsCh, anzCh, srvDep, srvIdxr)
	cdrS := services.NewCDRServer(cfg, dmS, sdbS, iFilterSCh, clsCh, iCDRServerCh, connMgr, anzCh, srvDep, srvIdxr)
	registrarcS := services.NewRegistrarCService(cfg, connMgr, srvDep, srvIdxr)
	rateS := services.NewRateService(cfg, cacheS, iFilterSCh, dmS, clsCh, iRateSCh, anzCh, srvDep, srvIdxr)
	actionS := services.NewActionService(cfg, dmS, cacheS, iFilterSCh, connMgr, clsCh, iActionSCh, anzCh, srvDep, srvIdxr)
	accS := services.NewAccountService(cfg, dmS, cacheS, iFilterSCh, connMgr, clsCh, iAccountSCh, anzCh, srvDep, srvIdxr)
	tpeS := services.NewTPeService(cfg, connMgr, dmS, clsCh, srvDep, srvIdxr)

	srvManager := servmanager.NewServiceManager(shdWg, connMgr, cfg, srvIdxr, []servmanager.Service{
		gvS,
		dmS,
		sdbS,
		cls,
		anzS,
		coreS,
		cacheS,
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
		anzCh <- anzS
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
	// Start FilterS
	go cgrStartFilterService(ctx, iFilterSCh, cacheS.GetCacheSChan(), connMgr, cfg, dmS)

	cgrInitServiceManagerV1(iServeManagerCh, srvManager, cfg, clsCh, anzS)
	cgrInitGuardianSv1(iGuardianSCh, cfg, clsCh, anzS)
	cgrInitConfigSv1(iConfigCh, cfg, clsCh, anzS)

	if *flags.Preload != utils.EmptyString {
		if err = cgrRunPreload(ctx, cfg, *flags.Preload, ldrs); err != nil {
			return
		}
	}

	// Serve rpc connections
	cgrStartRPC(ctx, cancel, cfg, clsCh, iDispatcherSCh)

	// TODO: find a better location for this if block
	if *flags.MemPrfDir != "" {
		if err := coreS.GetCoreS().StartMemoryProfiling(cores.MemoryProfilingParams{
			DirPath:      *flags.MemPrfDir,
			MaxFiles:     *flags.MemPrfMaxF,
			Interval:     *flags.MemPrfInterval,
			UseTimestamp: *flags.MemPrfTS,
		}); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
		}
	}

	<-ctx.Done()
	//<-stopChan
	return
}

func cgrRunPreload(ctx *context.Context, cfg *config.CGRConfig, loaderIDs string,
	loader *services.LoaderService) (err error) {
	if !cfg.LoaderCfg().Enabled() {
		err = fmt.Errorf("<%s> not enabled but required by preload mechanism", utils.LoaderS)
		return
	}
	ch := loader.GetRPCChan()
	select {
	case ldrs := <-ch:
		ch <- ldrs
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

// cgrStartFilterService fires up the FilterS
func cgrStartFilterService(ctx *context.Context, iFilterSCh chan *engine.FilterS,
	cacheSCh chan *engine.CacheS, connMgr *engine.ConnManager,
	cfg *config.CGRConfig, db *services.DataDBService) {
	var cacheS *engine.CacheS
	select {
	case cacheS = <-cacheSCh:
		cacheSCh <- cacheS
	case <-ctx.Done():
		return
	}
	dm, err := db.WaitForDM(ctx)
	if err != nil {
		return
	}
	select {
	case <-cacheS.GetPrecacheChannel(utils.CacheFilters):
		iFilterSCh <- engine.NewFilterS(cfg, connMgr, dm)
	case <-ctx.Done():
	}
}

func cgrInitGuardianSv1(iGuardianSCh chan birpc.ClientConnector, cfg *config.CGRConfig,
	clSChan chan *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	cl := <-clSChan
	clSChan <- cl
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
	clSChan chan *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	cl := <-clSChan
	clSChan <- cl
	srv, _ := birpc.NewService(apis.NewServiceManagerV1(srvMngr), utils.EmptyString, false)
	if !cfg.DispatcherSCfg().Enabled {
		cl.RpcRegister(srv)
	}
	iServMngrCh <- anz.GetInternalCodec(srv, utils.ServiceManager)
}

func cgrInitConfigSv1(iConfigCh chan birpc.ClientConnector,
	cfg *config.CGRConfig, clSChan chan *commonlisteners.CommonListenerS, anz *services.AnalyzerService) {
	cl := <-clSChan
	clSChan <- cl
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
	cfg *config.CGRConfig, clSChan chan *commonlisteners.CommonListenerS, internalDispatcherSChan chan birpc.ClientConnector) {
	cl := <-clSChan
	clSChan <- cl
	if cfg.DispatcherSCfg().Enabled { // wait only for dispatcher as cache is allways registered before this
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		case <-ctx.Done():
			return
		}
	}
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
			//  do it in it's own goroutine in order to not block the signal handler with the reload functionality
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
