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

package services

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/cgrates/birpc"
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/registrarc"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewCGREngine(cfg *config.CGRConfig, cM *engine.ConnManager, shdWg *sync.WaitGroup, server *cores.Server, caps *engine.Caps) *CGREngine {
	return &CGREngine{
		cfg:        cfg,
		cM:         cM,
		caps:       caps,
		shdWg:      shdWg,
		srvManager: servmanager.NewServiceManager(shdWg, cM, cfg.GetReloadChan()),
		server:     server, // Rpc/http server
		srvDep: map[string]*sync.WaitGroup{
			utils.AnalyzerS:       new(sync.WaitGroup),
			utils.AdminS:          new(sync.WaitGroup),
			utils.AsteriskAgent:   new(sync.WaitGroup),
			utils.AttributeS:      new(sync.WaitGroup),
			utils.CDRServer:       new(sync.WaitGroup),
			utils.ChargerS:        new(sync.WaitGroup),
			utils.CoreS:           new(sync.WaitGroup),
			utils.DataDB:          new(sync.WaitGroup),
			utils.DiameterAgent:   new(sync.WaitGroup),
			utils.RegistrarC:      new(sync.WaitGroup),
			utils.DispatcherS:     new(sync.WaitGroup),
			utils.DNSAgent:        new(sync.WaitGroup),
			utils.EEs:             new(sync.WaitGroup),
			utils.ERs:             new(sync.WaitGroup),
			utils.FreeSWITCHAgent: new(sync.WaitGroup),
			utils.GlobalVarS:      new(sync.WaitGroup),
			utils.HTTPAgent:       new(sync.WaitGroup),
			utils.KamailioAgent:   new(sync.WaitGroup),
			utils.LoaderS:         new(sync.WaitGroup),
			utils.RadiusAgent:     new(sync.WaitGroup),
			utils.RateS:           new(sync.WaitGroup),
			utils.ResourceS:       new(sync.WaitGroup),
			utils.RouteS:          new(sync.WaitGroup),
			utils.SchedulerS:      new(sync.WaitGroup),
			utils.SessionS:        new(sync.WaitGroup),
			utils.SIPAgent:        new(sync.WaitGroup),
			utils.StatS:           new(sync.WaitGroup),
			utils.ThresholdS:      new(sync.WaitGroup),
			utils.ActionS:         new(sync.WaitGroup),
			utils.AccountS:        new(sync.WaitGroup),
			utils.TPeS:            new(sync.WaitGroup),
		},
		iFilterSCh: make(chan *engine.FilterS, 1),
	}
}

type CGREngine struct {
	cfg *config.CGRConfig

	srvManager *servmanager.ServiceManager
	srvDep     map[string]*sync.WaitGroup
	shdWg      *sync.WaitGroup
	cM         *engine.ConnManager
	server     *cores.Server

	caps       *engine.Caps
	memPrfStop chan struct{}
	cpuPrfF    io.Closer

	// services
	gvS    servmanager.Service
	dmS    *DataDBService
	anzS   *AnalyzerService
	coreS  *CoreService
	cacheS *CacheService
	ldrs   *LoaderService

	// chans (need to move this as services)
	iFilterSCh      chan *engine.FilterS
	iGuardianSCh    chan birpc.ClientConnector
	iConfigCh       chan birpc.ClientConnector
	iServeManagerCh chan birpc.ClientConnector

	iDispatcherSCh chan birpc.ClientConnector
}

func (cgr *CGREngine) GetServDeps() map[string]*sync.WaitGroup {
	return cgr.srvDep
}

func (cgr *CGREngine) AddService(service servmanager.Service, connName, apiPrefix string,
	iConnCh chan birpc.ClientConnector) {
	cgr.srvManager.AddServices(service)
	cgr.srvDep[service.ServiceName()] = new(sync.WaitGroup)
	cgr.cM.AddInternalConn(connName, apiPrefix, iConnCh)
}

func (cgr *CGREngine) InitServices(httpPrfPath string, cpuPrfFl io.Closer, memPrfDir string, memPrfStop chan struct{}) (err error) {
	if len(cgr.cfg.HTTPCfg().RegistrarSURL) != 0 {
		cgr.server.RegisterHTTPFunc(cgr.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if cgr.cfg.ConfigSCfg().Enabled {
		cgr.server.RegisterHTTPFunc(cgr.cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	if httpPrfPath != utils.EmptyString {
		cgr.server.RegisterProfiler(httpPrfPath)
	}

	// init the channel here because we need to pass them to connManager
	cgr.iServeManagerCh = make(chan birpc.ClientConnector, 1)
	cgr.iConfigCh = make(chan birpc.ClientConnector, 1)
	iCoreSv1Ch := make(chan birpc.ClientConnector, 1)
	iCacheSCh := make(chan birpc.ClientConnector, 1)
	cgr.iGuardianSCh = make(chan birpc.ClientConnector, 1)
	iAnalyzerSCh := make(chan birpc.ClientConnector, 1)
	iCDRServerCh := make(chan birpc.ClientConnector, 1)
	iAttributeSCh := make(chan birpc.ClientConnector, 1)
	cgr.iDispatcherSCh = make(chan birpc.ClientConnector, 1)
	iSessionSCh := make(chan birpc.ClientConnector, 1)
	iChargerSCh := make(chan birpc.ClientConnector, 1)
	iThresholdSCh := make(chan birpc.ClientConnector, 1)
	iStatSCh := make(chan birpc.ClientConnector, 1)
	iResourceSCh := make(chan birpc.ClientConnector, 1)
	iRouteSCh := make(chan birpc.ClientConnector, 1)
	iAdminSCh := make(chan birpc.ClientConnector, 1)
	iLoaderSCh := make(chan birpc.ClientConnector, 1)
	iEEsCh := make(chan birpc.ClientConnector, 1)
	iRateSCh := make(chan birpc.ClientConnector, 1)
	iActionSCh := make(chan birpc.ClientConnector, 1)
	iAccountSCh := make(chan birpc.ClientConnector, 1)
	iTpeSCh := make(chan birpc.ClientConnector, 1)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzer), utils.AnalyzerSv1, iAnalyzerSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS), utils.AdminSv1, iAdminSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes), utils.AttributeSv1, iAttributeSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches), utils.CacheSv1, iCacheSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs), utils.CDRsV1, iCDRServerCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers), utils.ChargerSv1, iChargerSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian), utils.GuardianSv1, cgr.iGuardianSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders), utils.LoaderSv1, iLoaderSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources), utils.ResourceSv1, iResourceSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS), utils.SessionSv1, iSessionSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS), utils.SessionSv1, iSessionSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats), utils.StatSv1, iStatSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes), utils.RouteSv1, iRouteSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds), utils.ThresholdSv1, iThresholdSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager), utils.ServiceManagerV1, cgr.iServeManagerCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig), utils.ConfigSv1, cgr.iConfigCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore), utils.CoreSv1, iCoreSv1Ch)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs), utils.EeSv1, iEEsCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS), utils.RateSv1, iRateSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers), utils.DispatcherSv1, cgr.iDispatcherSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts), utils.AccountSv1, iAccountSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions), utils.ActionSv1, iActionSCh)
	cgr.cM.AddInternalConn(utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTpes), utils.TPeSv1, iTpeSCh)

	cgr.gvS = NewGlobalVarS(cgr.cfg, cgr.srvDep)
	cgr.dmS = NewDataDBService(cgr.cfg, cgr.cM, cgr.srvDep)
	cgr.anzS = NewAnalyzerService(cgr.cfg, cgr.server,
		cgr.iFilterSCh, iAnalyzerSCh, cgr.srvDep) // init AnalyzerS

	cgr.coreS = NewCoreService(cgr.cfg, cgr.caps, cgr.server,
		iCoreSv1Ch, cgr.anzS, cpuPrfFl, memPrfDir, memPrfStop,
		cgr.shdWg, cgr.srvDep) // init CoreSv1
	cgr.memPrfStop = memPrfStop
	cgr.cpuPrfF = cpuPrfFl

	cgr.cacheS = NewCacheService(cgr.cfg, cgr.dmS, cgr.server,
		iCacheSCh, cgr.anzS, cgr.coreS,
		cgr.srvDep) // init CacheS

	dspS := NewDispatcherService(cgr.cfg, cgr.dmS, cgr.cacheS,
		cgr.iFilterSCh, cgr.server, cgr.iDispatcherSCh, cgr.cM,
		cgr.anzS, cgr.srvDep)

	cgr.ldrs = NewLoaderService(cgr.cfg, cgr.dmS, cgr.iFilterSCh, cgr.server,
		iLoaderSCh, cgr.cM, cgr.anzS, cgr.srvDep)

	cgr.srvManager.AddServices(cgr.gvS, cgr.coreS, cgr.cacheS,
		cgr.ldrs, cgr.anzS, dspS, cgr.dmS,
		NewAdminSv1Service(cgr.cfg, cgr.dmS, cgr.iFilterSCh, cgr.server,
			iAdminSCh, cgr.cM, cgr.anzS, cgr.srvDep),
		NewSessionService(cgr.cfg, cgr.dmS, cgr.iFilterSCh, cgr.server, iSessionSCh, cgr.cM, cgr.anzS, cgr.srvDep),
		NewAttributeService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.server, iAttributeSCh,
			cgr.anzS, dspS, cgr.srvDep),
		NewChargerService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.server,
			iChargerSCh, cgr.cM, cgr.anzS, cgr.srvDep),
		NewRouteService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.server,
			iRouteSCh, cgr.cM, cgr.anzS, cgr.srvDep),
		NewResourceService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.server,
			iResourceSCh, cgr.cM, cgr.anzS, cgr.srvDep),
		NewThresholdService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh,
			cgr.cM, cgr.server, iThresholdSCh, cgr.anzS, cgr.srvDep),
		NewStatService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.server,
			iStatSCh, cgr.cM, cgr.anzS, cgr.srvDep),

		NewEventReaderService(cgr.cfg, cgr.iFilterSCh, cgr.cM, cgr.srvDep),
		NewDNSAgent(cgr.cfg, cgr.iFilterSCh, cgr.cM, cgr.srvDep),
		NewFreeswitchAgent(cgr.cfg, cgr.cM, cgr.srvDep),
		NewKamailioAgent(cgr.cfg, cgr.cM, cgr.srvDep),
		NewAsteriskAgent(cgr.cfg, cgr.cM, cgr.srvDep),                         // partial reload
		NewRadiusAgent(cgr.cfg, cgr.iFilterSCh, cgr.cM, cgr.srvDep),           // partial reload
		NewDiameterAgent(cgr.cfg, cgr.iFilterSCh, cgr.cM, cgr.srvDep),         // partial reload
		NewHTTPAgent(cgr.cfg, cgr.iFilterSCh, cgr.server, cgr.cM, cgr.srvDep), // no reload
		NewSIPAgent(cgr.cfg, cgr.iFilterSCh, cgr.cM, cgr.srvDep),

		NewEventExporterService(cgr.cfg, cgr.iFilterSCh,
			cgr.cM, cgr.server, iEEsCh, cgr.anzS, cgr.srvDep),
		NewCDRServer(cgr.cfg, cgr.dmS, cgr.iFilterSCh, cgr.server, iCDRServerCh,
			cgr.cM, cgr.anzS, cgr.srvDep),

		NewRegistrarCService(cgr.cfg, cgr.server, cgr.cM, cgr.anzS, cgr.srvDep),

		NewRateService(cgr.cfg, cgr.cacheS, cgr.iFilterSCh, cgr.dmS,
			cgr.server, iRateSCh, cgr.anzS, cgr.srvDep),
		NewActionService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.cM, cgr.server, iActionSCh, cgr.anzS, cgr.srvDep),
		NewAccountService(cgr.cfg, cgr.dmS, cgr.cacheS, cgr.iFilterSCh, cgr.cM, cgr.server, iAccountSCh, cgr.anzS, cgr.srvDep),
		NewTPeService(cgr.cfg, cgr.cM, cgr.dmS, cgr.server, cgr.srvDep),
	)

	return
}

func (cgr *CGREngine) StartServices(ctx *context.Context, shtDw context.CancelFunc, preload string) (err error) {
	defer func() {
		if err != nil {
			cgr.srvManager.ShutdownServices()
		}
	}()
	cgr.shdWg.Add(1)
	if err = cgr.gvS.Start(ctx, shtDw); err != nil {
		cgr.shdWg.Done()
		return
	}
	if cgr.dmS.ShouldRun() { // Some services can run without db, ie:  ERs
		cgr.shdWg.Add(1)
		if err = cgr.dmS.Start(ctx, shtDw); err != nil {
			cgr.shdWg.Done()
			return
		}
	}

	if cgr.anzS.ShouldRun() {
		cgr.shdWg.Add(1)
		if err = cgr.anzS.Start(ctx, shtDw); err != nil {
			cgr.shdWg.Done()
			return
		}
	}

	cgr.shdWg.Add(1)
	if err = cgr.coreS.Start(ctx, shtDw); err != nil {
		cgr.shdWg.Done()
		return
	}
	cgr.shdWg.Add(1)
	if err = cgr.cacheS.Start(ctx, shtDw); err != nil {
		cgr.shdWg.Done()
		return
	}
	cgr.srvManager.StartServices(ctx, shtDw)
	// Start FilterS
	go cgrStartFilterService(ctx, cgr.iFilterSCh, cgr.cacheS.GetCacheSChan(), cgr.cM,
		cgr.cfg, cgr.dmS)

	cgrInitServiceManagerV1(cgr.iServeManagerCh, cgr.srvManager, cgr.cfg, cgr.server, cgr.anzS)
	cgrInitGuardianSv1(cgr.iGuardianSCh, cgr.cfg, cgr.server, cgr.anzS) // init GuardianSv1
	cgrInitConfigSv1(cgr.iConfigCh, cgr.cfg, cgr.server, cgr.anzS)

	if preload != utils.EmptyString {
		if err = cgrRunPreload(ctx, cgr.cfg, preload, cgr.ldrs); err != nil {
			return
		}
	}

	// Serve rpc connections
	cgrStartRPC(ctx, shtDw, cgr.cfg, cgr.server, cgr.iDispatcherSCh)

	return
}

func (cgr *CGREngine) Init(ctx *context.Context, shtDw context.CancelFunc, flags *CGREngineFlags, vers string) (err error) {
	cgr.shdWg.Add(1)
	go cgrSingnalHandler(ctx, shtDw, cgr.cfg, cgr.shdWg)

	var memPrfStop chan struct{}
	if *flags.MemPrfDir != utils.EmptyString {
		cgr.shdWg.Add(1)
		memPrfStop = make(chan struct{})
		go cores.MemProfiling(*flags.MemPrfDir, *flags.MemPrfInterval, *flags.MemPrfNoF, cgr.shdWg, memPrfStop, shtDw)
	}

	var cpuPrfF io.Closer
	if *flags.CpuPrfDir != utils.EmptyString {
		if cpuPrfF, err = cores.StartCPUProfiling(
			path.Join(*flags.CpuPrfDir, utils.CpuPathCgr)); err != nil {
			return
		}
	}

	if *flags.ScheduledShutDown != utils.EmptyString {
		var shtDwDur time.Duration
		if shtDwDur, err = utils.ParseDurationWithNanosecs(*flags.ScheduledShutDown); err != nil {
			return
		}
		cgr.shdWg.Add(1)
		go func() { // Schedule shutdown
			tm := time.NewTimer(shtDwDur)
			select {
			case <-tm.C:
				shtDw()
			case <-ctx.Done():
				tm.Stop()
			}
			cgr.shdWg.Done()
		}()
	}

	// init syslog
	if utils.Logger, err = utils.NewLogger(
		utils.FirstNonEmpty(*flags.SysLogger, cgr.cfg.LoggerCfg().Type),
		cgr.cfg.GeneralCfg().DefaultTenant,
		cgr.cfg.GeneralCfg().NodeID,
		cgr.cfg.LoggerCfg().Level,
		cgr.cfg.LoggerCfg().Opts.Attempts,
		cgr.cfg.LoggerCfg().Opts.KafkaConn,
		cgr.cfg.LoggerCfg().Opts.KafkaTopic,
		cgr.cfg.LoggerCfg().Opts.FailedPostsDir); err != nil {
		return fmt.Errorf("Could not initialize syslog connection, err: <%s>", err)
	}
	utils.SetFailedPostCacheTTL(cgr.cfg.GeneralCfg().FailedPostsTTL) // init failedPosts to posts loggers/exporters in case of failing
	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, runtime.Version()))
	cgr.cfg.LazySanityCheck()

	return cgr.InitServices(*flags.HttpPrfPath, cpuPrfF, *flags.MemPrfDir, memPrfStop)
}

func (cgr *CGREngine) Stop(memPrfDir, pidFile string) {
	if cgr.memPrfStop != nil && cgr.coreS == nil {
		close(cgr.memPrfStop)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cgr.cfg.CoreSCfg().ShutdownTimeout*10)
	go func() {
		cgr.shdWg.Wait()
		cancel()
	}()
	<-ctx.Done()
	if ctx.Err() != context.Canceled {
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown all subsystems in the given time",
			utils.ServiceManager))
	}

	if memPrfDir != utils.EmptyString { // write last memory profiling
		cores.MemProfFile(path.Join(memPrfDir, utils.MemProfFileCgr))
	}
	if pidFile != utils.EmptyString {
		if err := os.Remove(pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	if cgr.cpuPrfF != nil && cgr.coreS == nil {
		pprof.StopCPUProfile()
		cgr.cpuPrfF.Close()
	}
	utils.Logger.Info("<CoreS> stopped all components. CGRateS shutdown!")
}
