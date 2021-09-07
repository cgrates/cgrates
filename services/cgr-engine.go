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
	"log"
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

type CGREngine struct {
	cfg *config.CGRConfig

	srvManager *servmanager.ServiceManager
	srvDep     map[string]*sync.WaitGroup
	shdWg      sync.WaitGroup
	cM         *engine.ConnManager
	server     *cores.Server

	cS *cores.CoreService
}

func (cgr *CGREngine) AddService(service servmanager.Service, connName, apiPrefix string,
	iConnCh chan birpc.ClientConnector) {
	cgr.srvManager.AddServices(service)
	cgr.srvDep[service.ServiceName()] = new(sync.WaitGroup)
	cgr.cM.AddInternalConn(connName, apiPrefix, iConnCh)
}

func (cgr *CGREngine) InitConfigFromPath(path, nodeID string, lgLevel int) (err error) {
	// Init config
	if cgr.cfg, err = config.NewCGRConfigFromPath(path); err != nil {
		err = fmt.Errorf("could not parse config: <%s>", err)
		return
	}
	if cgr.cfg.ConfigDBCfg().Type != utils.MetaInternal {
		var d config.ConfigDB
		if d, err = engine.NewDataDBConn(cgr.cfg.ConfigDBCfg().Type,
			cgr.cfg.ConfigDBCfg().Host, cgr.cfg.ConfigDBCfg().Port,
			cgr.cfg.ConfigDBCfg().Name, cgr.cfg.ConfigDBCfg().User,
			cgr.cfg.ConfigDBCfg().Password, cgr.cfg.GeneralCfg().DBDataEncoding,
			cgr.cfg.ConfigDBCfg().Opts); err != nil { // Cannot configure getter database, show stopper
			err = fmt.Errorf("could not configure configDB: <%s>", err)
			return
		}
		if err = cgr.cfg.LoadFromDB(d); err != nil {
			err = fmt.Errorf("could not parse config from DB: <%s>", err)
			return
		}
	}
	if nodeID != utils.EmptyString {
		cgr.cfg.GeneralCfg().NodeID = nodeID
	}
	if lgLevel != -1 { // Modify the log level if provided by command arguments
		cgr.cfg.GeneralCfg().LogLevel = lgLevel
	}
	config.SetCgrConfig(cgr.cfg) // Share the config object
	return
}

func (cgr *CGREngine) InitServices(ctx *context.Context, shtDwn context.CancelFunc, pprofPath string, cpuPrfFl io.Closer, memPrfDir string, memPrfStop chan struct{}) (err error) {
	iFilterSCh := make(chan *engine.FilterS, 1)
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
	iResourceSCh := make(chan birpc.ClientConnector, 1)
	iRouteSCh := make(chan birpc.ClientConnector, 1)
	iAdminSCh := make(chan birpc.ClientConnector, 1)
	iLoaderSCh := make(chan birpc.ClientConnector, 1)
	iEEsCh := make(chan birpc.ClientConnector, 1)
	iRateSCh := make(chan birpc.ClientConnector, 1)
	iActionSCh := make(chan birpc.ClientConnector, 1)
	iAccountSCh := make(chan birpc.ClientConnector, 1)

	cgr.srvDep = map[string]*sync.WaitGroup{
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
		utils.StorDB:          new(sync.WaitGroup),
		utils.ThresholdS:      new(sync.WaitGroup),
		utils.ActionS:         new(sync.WaitGroup),
		utils.AccountS:        new(sync.WaitGroup),
	}

	cncReqsLimit := cgr.cfg.CoreSCfg().Caps
	if utils.ConcurrentReqsLimit != 0 { // used as shared variable
		cncReqsLimit = utils.ConcurrentReqsLimit
	}
	cncReqsStrategy := cgr.cfg.CoreSCfg().CapsStrategy
	if len(utils.ConcurrentReqsStrategy) != 0 {
		cncReqsStrategy = utils.ConcurrentReqsStrategy
	}
	caps := engine.NewCaps(cncReqsLimit, cncReqsStrategy)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	cgr.cM = engine.NewConnManager(cgr.cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzer):       iAnalyzerSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAdminS):         iAdminSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes):     iAttributeSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):         iCacheSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs):           iCDRServerCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):       iChargerSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian):       iGuardianSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders):        iLoaderSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):      iResourceSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS):       iSessionSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):          iStatSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):         iRouteSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     iThresholdSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): iServeManagerCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         iConfigCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           iCoreSv1Ch,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs):            iEEsCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS):          iRateSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers):    iDispatcherSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAccounts):       iAccountSCh,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions):        iActionSCh,
		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS):  iSessionSCh,
	})
	gvService := NewGlobalVarS(cgr.cfg, cgr.srvDep)
	cgr.shdWg.Add(1)
	if err = gvService.Start(); err != nil {
		return
	}
	dmService := NewDataDBService(cgr.cfg, cgr.cM, cgr.srvDep)
	if dmService.ShouldRun() { // Some services can run without db, ie:  ERs
		cgr.shdWg.Add(1)
		if err = dmService.Start(); err != nil {
			return
		}
	}

	storDBService := NewStorDBService(cgr.cfg, cgr.srvDep)
	if storDBService.ShouldRun() { // Some services can run without db, ie:  ERs
		cgr.shdWg.Add(1)
		if err = storDBService.Start(); err != nil {
			return
		}
	}

	// Rpc/http server
	cgr.server = cores.NewServer(caps)
	if len(cgr.cfg.HTTPCfg().RegistrarSURL) != 0 {
		cgr.server.RegisterHTTPFunc(cgr.cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if cgr.cfg.ConfigSCfg().Enabled {
		cgr.server.RegisterHTTPFunc(cgr.cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	if pprofPath != utils.EmptyString {
		cgr.server.RegisterProfiler(pprofPath)
	}

	// init AnalyzerS
	anz := NewAnalyzerService(cgr.cfg, cgr.server, iFilterSCh, iAnalyzerSCh, cgr.srvDep, shtDwn)
	if anz.ShouldRun() {
		cgr.shdWg.Add(1)
		if err = anz.Start(); err != nil {
			return
		}
	}

	// init CoreSv1
	coreS := NewCoreService(cgr.cfg, caps, cgr.server, iCoreSv1Ch, anz, cpuPrfFl, memPrfDir, memPrfStop, &cgr.shdWg, cgr.srvDep, shtDwn)
	cgr.shdWg.Add(1)
	if err = coreS.Start(); err != nil {
		return
	}
	cgr.cS = coreS.GetCoreS()

	// init CacheS
	cacheS := cgrInitCacheS(ctx, shtDwn, iCacheSCh, cgr.server, cgr.cfg, dmService.GetDM(), anz, coreS.GetCoreS().CapsStats)
	engine.Cache = cacheS

	// init GuardianSv1
	cgrInitGuardianSv1(iGuardianSCh, cgr.server, anz)

	// Start ServiceManager
	cgr.srvManager = servmanager.NewServiceManager(cgr.cfg, &cgr.shdWg, cgr.cM)
	dspS := NewDispatcherService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server, iDispatcherSCh, cgr.cM, anz, cgr.srvDep)
	attrS := NewAttributeService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server, iAttributeSCh, anz, dspS, cgr.srvDep)
	dspH := NewRegistrarCService(cgr.cfg, cgr.server, cgr.cM, anz, cgr.srvDep)
	chrS := NewChargerService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server,
		iChargerSCh, cgr.cM, anz, cgr.srvDep)
	tS := NewThresholdService(cgr.cfg, dmService, cacheS, iFilterSCh,
		cgr.cM, cgr.server, iThresholdSCh, anz, cgr.srvDep)
	stS := NewStatService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server,
		iStatSCh, cgr.cM, anz, cgr.srvDep)
	reS := NewResourceService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server,
		iResourceSCh, cgr.cM, anz, cgr.srvDep)
	routeS := NewRouteService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.server,
		iRouteSCh, cgr.cM, anz, cgr.srvDep)

	admS := NewAdminSv1Service(cgr.cfg, dmService, storDBService, iFilterSCh, cgr.server,
		iAdminSCh, cgr.cM, anz, cgr.srvDep)

	cdrS := NewCDRServer(cgr.cfg, dmService, storDBService, iFilterSCh, cgr.server, iCDRServerCh,
		cgr.cM, anz, cgr.srvDep)

	smg := NewSessionService(cgr.cfg, dmService, cgr.server, iSessionSCh, cgr.cM, anz, cgr.srvDep, shtDwn)

	ldrs := NewLoaderService(cgr.cfg, dmService, iFilterSCh, cgr.server,
		iLoaderSCh, cgr.cM, anz, cgr.srvDep)

	cgr.srvManager.AddServices(gvService, attrS, chrS, tS, stS, reS, routeS,
		admS, cdrS, smg, coreS,
		NewEventReaderService(cgr.cfg, iFilterSCh, cgr.cM, cgr.srvDep, shtDwn),
		NewDNSAgent(cgr.cfg, iFilterSCh, cgr.cM, cgr.srvDep, shtDwn),
		NewFreeswitchAgent(cgr.cfg, cgr.cM, cgr.srvDep, shtDwn),
		NewKamailioAgent(cgr.cfg, cgr.cM, cgr.srvDep, shtDwn),
		NewAsteriskAgent(cgr.cfg, cgr.cM, cgr.srvDep, shtDwn),             // partial reload
		NewRadiusAgent(cgr.cfg, iFilterSCh, cgr.cM, cgr.srvDep, shtDwn),   // partial reload
		NewDiameterAgent(cgr.cfg, iFilterSCh, cgr.cM, cgr.srvDep, shtDwn), // partial reload
		NewHTTPAgent(cgr.cfg, iFilterSCh, cgr.server, cgr.cM, cgr.srvDep), // no reload
		ldrs, anz, dspS, dspH, dmService, storDBService,
		NewEventExporterService(cgr.cfg, iFilterSCh,
			cgr.cM, cgr.server, iEEsCh, anz, cgr.srvDep),
		NewRateService(cgr.cfg, cacheS, iFilterSCh, dmService,
			cgr.server, iRateSCh, anz, cgr.srvDep),
		NewSIPAgent(cgr.cfg, iFilterSCh, cgr.cM, cgr.srvDep, shtDwn),
		NewActionService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.cM, cgr.server, iActionSCh, anz, cgr.srvDep),
		NewAccountService(cgr.cfg, dmService, cacheS, iFilterSCh, cgr.cM, cgr.server, iAccountSCh, anz, cgr.srvDep),
	)
	return
}

func (cgr *CGREngine) Start(ctx *context.Context, shtDw context.CancelFunc, flags *CGREngineFlags) (err error) {
	var vers string
	goVers := runtime.Version()
	if vers, err = utils.GetCGRVersion(); err != nil {
		return
	}
	if *flags.Version {
		fmt.Println(vers)
		return
	}
	if *flags.PidFile != utils.EmptyString {
		cgrWritePid(*flags.PidFile)
	}
	if *flags.Singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}
	cgr.shdWg.Add(1)
	go cgrSingnalHandler(ctx, shtDw, cgr.cfg, &cgr.shdWg)

	var stopMemProf chan struct{}
	if *flags.MemPrfDir != utils.EmptyString {
		cgr.shdWg.Add(1)
		stopMemProf = make(chan struct{})
		go cores.MemProfiling(*flags.MemPrfDir, *flags.MemPrfInterval, *flags.MemPrfNoF, &cgr.shdWg, stopMemProf, shtDw)
		defer func() { //here
			if cgr.cS == nil {
				close(stopMemProf)
			}
		}()
	}

	var cpuProfileFile io.Closer
	if *flags.CpuPrfDir != utils.EmptyString {
		cpuProfileFile, err = cores.StartCPUProfiling(path.Join(*flags.CpuPrfDir, utils.CpuPathCgr))
		if err != nil {
			return
		}
		defer func() { //here
			if cgr.cS == nil {
				pprof.StopCPUProfile()
				cpuProfileFile.Close()
			}
		}()
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

	// Init config
	if err = cgr.InitConfigFromPath(*flags.CfgPath, *flags.NodeID, *flags.LogLevel); err != nil {
		return
	}
	if *flags.CheckConfig {
		return
	}

	// init syslog
	if utils.Logger, err = utils.Newlogger(utils.FirstNonEmpty(*flags.SysLogger,
		cgr.cfg.GeneralCfg().Logger), cgr.cfg.GeneralCfg().NodeID); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err)
		return
	}
	utils.Logger.SetLogLevel(cgr.cfg.GeneralCfg().LogLevel)
	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, goVers))
	cgr.cfg.LazySanityCheck()

	return
}
