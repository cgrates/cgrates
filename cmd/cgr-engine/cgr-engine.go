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
	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/loaders"
	"github.com/cgrates/cgrates/registrarc"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrEngineFlags    = flag.NewFlagSet(utils.CgrEngine, flag.ExitOnError)
	cfgPath           = cgrEngineFlags.String(utils.CfgPathCgr, utils.ConfigPath, "Configuration directory path")
	version           = cgrEngineFlags.Bool(utils.VersionCgr, false, "Print application version and exit")
	printConfig       = cgrEngineFlags.Bool(utils.PrintCfgCgr, false, "Print configuration object in JSON format")
	pidFile           = cgrEngineFlags.String(utils.PidCgr, utils.EmptyString, "Path to write the PID file")
	cpuProfDir        = cgrEngineFlags.String(utils.CpuProfDirCgr, utils.EmptyString, "Directory for CPU profiles")
	memProfDir        = cgrEngineFlags.String(utils.MemProfDirCgr, utils.EmptyString, "Directory for memory profiles")
	memProfInterval   = cgrEngineFlags.Duration(utils.MemProfIntervalCgr, 15*time.Second, "Interval between memory profile saves")
	memProfMaxFiles   = cgrEngineFlags.Int(utils.MemProfMaxFilesCgr, 1, "Number of memory profiles to keep (most recent)")
	memProfTimestamp  = cgrEngineFlags.Bool(utils.MemProfTimestampCgr, false, "Add timestamp to memory profile files")
	scheduledShutdown = cgrEngineFlags.Duration(utils.ScheduledShutdownCgr, 0, "Shutdown the engine after the specified duration")
	singleCPU         = cgrEngineFlags.Bool(utils.SingleCpuCgr, false, "Run on a single CPU core")
	syslogger         = cgrEngineFlags.String(utils.LoggerCfg, utils.EmptyString, "Logger type <*syslog|*stdout>")
	nodeID            = cgrEngineFlags.String(utils.NodeIDCfg, utils.EmptyString, "Node ID of the engine")
	logLevel          = cgrEngineFlags.Int(utils.LogLevelCfg, -1, "Log level (0=emergency to 7=debug)")
	preload           = cgrEngineFlags.String(utils.PreloadCgr, utils.EmptyString, "Loader IDs used to load data before engine starts")
	setVersions       = cgrEngineFlags.Bool(utils.SetVersionsCgr, false, "Overwrite database versions (equivalent to cgr-migrator -exec=*set_versions)")

	cfg *config.CGRConfig
)

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS, connMgr *engine.ConnManager, cfg *config.CGRConfig,
	dm *engine.DataManager) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, connMgr, dm)
}

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan birpc.ClientConnector,
	server *cores.Server, dm *engine.DataManager, shdChan *utils.SyncedChan,
	anz *services.AnalyzerService,
	cpS *engine.CapsStats) (*engine.CacheS, error) {
	chS := engine.NewCacheS(cfg, dm, cpS)
	go func() {
		if err := chS.Precache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			shdChan.CloseOnce()
		}
	}()

	srv, err := engine.NewService(v1.NewCacheSv1(chS))
	if err != nil {
		return nil, err
	}
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(srv)
	}
	internalCacheSChan <- anz.GetInternalCodec(srv, utils.CacheS)
	return chS, nil
}

func initGuardianSv1(internalGuardianSChan chan birpc.ClientConnector, server *cores.Server,
	anz *services.AnalyzerService) error {
	srv, err := engine.NewService(v1.NewGuardianSv1())
	if err != nil {
		return err
	}
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(srv)
	}
	internalGuardianSChan <- anz.GetInternalCodec(srv, utils.GuardianS)
	return nil
}

func initServiceManagerV1(internalServiceManagerChan chan birpc.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *cores.Server,
	anz *services.AnalyzerService) error {
	srv, err := engine.NewService(v1.NewServiceManagerV1(srvMngr))
	if err != nil {
		return err
	}
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(srv)
	}
	internalServiceManagerChan <- anz.GetInternalCodec(srv, utils.ServiceManager)
	return nil
}

func initConfigSv1(internalConfigChan chan birpc.ClientConnector,
	server *cores.Server, anz *services.AnalyzerService) error {
	srv, err := engine.NewService(v1.NewConfigSv1(cfg))
	if err != nil {
		return err
	}
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(srv)
	}
	internalConfigChan <- anz.GetInternalCodec(srv, utils.ConfigSv1)
	return nil
}

func startRPC(server *cores.Server, internalRaterChan,
	internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalTrendSChan, internalSuplSChan,
	internalSMGChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalRALsv1Chan, internalCacheSChan,
	internalEEsChan, internalERsChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan) {
	if !cfg.DispatcherSCfg().Enabled {
		select { // Any of the rpc methods will unlock listening to rpc requests
		case resp := <-internalRaterChan:
			internalRaterChan <- resp
		case cdrs := <-internalCdrSChan:
			internalCdrSChan <- cdrs
		case smg := <-internalSMGChan:
			internalSMGChan <- smg
		case rls := <-internalRsChan:
			internalRsChan <- rls
		case statS := <-internalStatSChan:
			internalStatSChan <- statS
		case attrS := <-internalAttrSChan:
			internalAttrSChan <- attrS
		case chrgS := <-internalChargerSChan:
			internalChargerSChan <- chrgS
		case thS := <-internalThdSChan:
			internalThdSChan <- thS
		case trS := <-internalTrendSChan:
			internalTrendSChan <- trS
		case splS := <-internalSuplSChan:
			internalSuplSChan <- splS
		case analyzerS := <-internalAnalyzerSChan:
			internalAnalyzerSChan <- analyzerS
		case loaderS := <-internalLoaderSChan:
			internalLoaderSChan <- loaderS
		case ralS := <-internalRALsv1Chan:
			internalRALsv1Chan <- ralS
		case chS := <-internalCacheSChan: // added in order to start the RPC before precaching is done
			internalCacheSChan <- chS
		case eeS := <-internalEEsChan:
			internalEEsChan <- eeS
		case erS := <-internalERsChan:
			internalERsChan <- erS
		case <-shdChan.Done():
			return
		}
	} else {
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		case <-shdChan.Done():
			return
		}
	}

	go server.ServeJSON(cfg.ListenCfg().RPCJSONListen, shdChan)
	go server.ServeGOB(cfg.ListenCfg().RPCGOBListen, shdChan)
	go server.ServeHTTP(
		cfg.ListenCfg().HTTPListen,
		cfg.HTTPCfg().HTTPJsonRPCURL,
		cfg.HTTPCfg().HTTPWSURL,
		cfg.HTTPCfg().PrometheusURL,
		cfg.HTTPCfg().PprofPath,
		cfg.HTTPCfg().HTTPUseBasicAuth,
		cfg.HTTPCfg().HTTPAuthUsers,
		shdChan,
	)
	if (len(cfg.ListenCfg().RPCGOBTLSListen) != 0 ||
		len(cfg.ListenCfg().RPCJSONTLSListen) != 0 ||
		len(cfg.ListenCfg().HTTPTLSListen) != 0) &&
		(len(cfg.TLSCfg().ServerCerificate) == 0 ||
			len(cfg.TLSCfg().ServerKey) == 0) {
		utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		return
	}
	if cfg.ListenCfg().RPCGOBTLSListen != utils.EmptyString {
		go server.ServeGOBTLS(
			cfg.ListenCfg().RPCGOBTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			shdChan,
		)
	}
	if cfg.ListenCfg().RPCJSONTLSListen != utils.EmptyString {
		go server.ServeJSONTLS(
			cfg.ListenCfg().RPCJSONTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			shdChan,
		)
	}
	if cfg.ListenCfg().HTTPTLSListen != utils.EmptyString {
		go server.ServeHTTPTLS(
			cfg.ListenCfg().HTTPTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			cfg.HTTPCfg().HTTPJsonRPCURL,
			cfg.HTTPCfg().HTTPWSURL,
			cfg.HTTPCfg().PrometheusURL,
			cfg.HTTPCfg().PprofPath,
			cfg.HTTPCfg().HTTPUseBasicAuth,
			cfg.HTTPCfg().HTTPAuthUsers,
			shdChan,
		)
	}
}

func writePid() {
	utils.Logger.Info(*pidFile)
	f, err := os.Create(*pidFile)
	if err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
	f.WriteString(strconv.Itoa(os.Getpid()))
	if err := f.Close(); err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
}

func singnalHandler(shdWg *sync.WaitGroup, shdChan *utils.SyncedChan) {
	shutdownSignal := make(chan os.Signal, 1)
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shdChan.Done():
			shdWg.Done()
			return
		case <-shutdownSignal:
			shdChan.CloseOnce()
			shdWg.Done()
			return
		case <-reloadSignal:
			//  do it in it's own gorutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := config.CgrConfig().V1ReloadConfig(
					context.TODO(),
					&config.ReloadArgs{
						Section: utils.EmptyString,
						Path:    config.CgrConfig().ConfigPath, // use the same path
					}, &reply); err != nil {
					utils.Logger.Warning(
						fmt.Sprintf("Error reloading configuration: <%s>", err))
				}
			}()
		}
	}
}

func runPreload(loader *services.LoaderService, internalLoaderSChan chan birpc.ClientConnector,
	shdChan *utils.SyncedChan) {
	if !cfg.LoaderCfg().Enabled() {
		utils.Logger.Err(fmt.Sprintf("<%s> not enabled but required by preload mechanism", utils.LoaderS))
		shdChan.CloseOnce()
		return
	}

	ldrs := <-internalLoaderSChan
	internalLoaderSChan <- ldrs

	var reply string
	for _, loaderID := range strings.Split(*preload, utils.FieldsSep) {
		if err := loader.GetLoaderS().V1Load(context.TODO(),
			&loaders.ArgsProcessFolder{
				ForceLock:   true, // force lock will unlock the file in case is locked and return error
				LoaderID:    loaderID,
				StopOnError: true,
			}, &reply); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> preload failed on loadID <%s> , err: <%s>", utils.LoaderS, loaderID, err.Error()))
			shdChan.CloseOnce()
			return
		}
	}
}

func main() {
	cgrEngineFlags.Parse(os.Args[1:])
	vers, err := utils.GetCGRVersion()
	if err != nil {
		log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
	}
	goVers := runtime.Version()
	if *version {
		fmt.Println(vers)
		return
	}
	if *pidFile != utils.EmptyString {
		writePid()
	}
	if *singleCPU {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	shdWg := new(sync.WaitGroup)
	shdChan := utils.NewSyncedChan()

	shdWg.Add(1)
	go singnalHandler(shdWg, shdChan)

	var cS *cores.CoreService
	var cpuProf *os.File
	if *cpuProfDir != utils.EmptyString {
		cpuPath := filepath.Join(*cpuProfDir, utils.CpuPathCgr)
		cpuProf, err = cores.StartCPUProfiling(cpuPath)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if cS != nil {
				// Use CoreService's StopCPUProfiling method if it has been initialized.
				if err := cS.StopCPUProfiling(); err != nil {
					log.Print(err)
				}
				return
			}
			pprof.StopCPUProfile()
			if err := cpuProf.Close(); err != nil {
				log.Print(err)
			}
		}()
	}

	if *scheduledShutdown != 0 {
		shdWg.Add(1)
		go func() { // Schedule shutdown
			tm := time.NewTimer(*scheduledShutdown)
			select {
			case <-tm.C:
				shdChan.CloseOnce()
			case <-shdChan.Done():
				tm.Stop()
			}
			shdWg.Done()
		}()
	}

	// Init config
	cfg, err = config.NewCGRConfigFromPath(*cfgPath)
	if err != nil {
		log.Fatalf("Could not parse config: <%s>", err.Error())
	}

	if *nodeID != utils.EmptyString {
		cfg.GeneralCfg().NodeID = *nodeID
	}

	config.SetCgrConfig(cfg) // Share the config object

	// init syslog
	if utils.Logger, err = utils.Newlogger(utils.FirstNonEmpty(*syslogger,
		cfg.GeneralCfg().Logger), cfg.GeneralCfg().NodeID); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err.Error())
	}
	lgLevel := cfg.GeneralCfg().LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)

	if *printConfig {
		cfgJSON := utils.ToIJSON(cfg.AsMapInterface(cfg.GeneralCfg().RSRSep))
		utils.Logger.Info(fmt.Sprintf("Configuration loaded from %q:\n%s", *cfgPath, cfgJSON))
	}

	// init the concurrentRequests
	caps := engine.NewCaps(cfg.CoreSCfg().Caps, cfg.CoreSCfg().CapsStrategy)
	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, goVers))

	// init the channel here because we need to pass them to connManager
	internalServeManagerChan := make(chan birpc.ClientConnector, 1)
	internalConfigChan := make(chan birpc.ClientConnector, 1)
	internalCoreSv1Chan := make(chan birpc.ClientConnector, 1)
	internalCacheSChan := make(chan birpc.ClientConnector, 1)
	internalGuardianSChan := make(chan birpc.ClientConnector, 1)
	internalAnalyzerSChan := make(chan birpc.ClientConnector, 1)
	internalCDRServerChan := make(chan birpc.ClientConnector, 1)
	internalAttributeSChan := make(chan birpc.ClientConnector, 1)
	internalDispatcherSChan := make(chan birpc.ClientConnector, 1)
	internalSessionSChan := make(chan birpc.ClientConnector, 1)
	internalChargerSChan := make(chan birpc.ClientConnector, 1)
	internalThresholdSChan := make(chan birpc.ClientConnector, 1)
	internalStatSChan := make(chan birpc.ClientConnector, 1)
	internalTrendSChan := make(chan birpc.ClientConnector, 1)
	internalRankingSChan := make(chan birpc.ClientConnector, 1)
	internalResourceSChan := make(chan birpc.ClientConnector, 1)
	internalRouteSChan := make(chan birpc.ClientConnector, 1)
	internalSchedulerSChan := make(chan birpc.ClientConnector, 1)
	internalRALsChan := make(chan birpc.ClientConnector, 1)
	internalResponderChan := make(chan birpc.ClientConnector, 1)
	internalAPIerSv1Chan := make(chan birpc.ClientConnector, 1)
	internalAPIerSv2Chan := make(chan birpc.ClientConnector, 1)
	internalLoaderSChan := make(chan birpc.ClientConnector, 1)
	internalEEsChan := make(chan birpc.ClientConnector, 1)
	internalERsChan := make(chan birpc.ClientConnector, 1)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	connManager := engine.NewConnManager(cfg, map[string]chan birpc.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzer):       internalAnalyzerSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier):          internalAPIerSv2Chan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes):     internalAttributeSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):         internalCacheSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCDRs):           internalCDRServerChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers):       internalChargerSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaGuardian):       internalGuardianSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaLoaders):        internalLoaderSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources):      internalResourceSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResponder):      internalResponderChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaScheduler):      internalSchedulerSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS):       internalSessionSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStats):          internalStatSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):         internalRouteSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaTrends):         internalTrendSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRankings):       internalRankingSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     internalThresholdSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): internalServeManagerChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         internalConfigChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           internalCoreSv1Chan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):           internalRALsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs):            internalEEsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaERs):            internalERsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers):    internalDispatcherSChan,

		utils.ConcatenatedKey(rpcclient.BiRPCInternal, utils.MetaSessionS): internalSessionSChan,
	})
	srvDep := map[string]*sync.WaitGroup{
		utils.AnalyzerS:       new(sync.WaitGroup),
		utils.APIerSv1:        new(sync.WaitGroup),
		utils.APIerSv2:        new(sync.WaitGroup),
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
		utils.RALService:      new(sync.WaitGroup),
		utils.ResourceS:       new(sync.WaitGroup),
		utils.ResponderS:      new(sync.WaitGroup),
		utils.RouteS:          new(sync.WaitGroup),
		utils.SchedulerS:      new(sync.WaitGroup),
		utils.SessionS:        new(sync.WaitGroup),
		utils.SIPAgent:        new(sync.WaitGroup),
		utils.StatS:           new(sync.WaitGroup),
		utils.TrendS:          new(sync.WaitGroup),
		utils.RankingS:        new(sync.WaitGroup),
		utils.StorDB:          new(sync.WaitGroup),
		utils.ThresholdS:      new(sync.WaitGroup),
		utils.AccountS:        new(sync.WaitGroup),
	}
	gvService := services.NewGlobalVarS(cfg, srvDep)
	shdWg.Add(1)
	if err = gvService.Start(); err != nil {
		log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
	}
	dmService := services.NewDataDBService(cfg, connManager, *setVersions, srvDep)
	if dmService.ShouldRun() { // Some services can run without db, ie:  ERs
		shdWg.Add(1)
		if err = dmService.Start(); err != nil {
			log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
		}
	}

	storDBService := services.NewStorDBService(cfg, *setVersions, srvDep)
	if storDBService.ShouldRun() { // Some services can run without db, ie:  ERs
		shdWg.Add(1)
		if err = storDBService.Start(); err != nil {
			log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
		}
	}

	// Rpc/http server
	server := cores.NewServer(caps)
	if len(cfg.HTTPCfg().RegistrarSURL) != 0 {
		server.RegisterHttpFunc(cfg.HTTPCfg().RegistrarSURL, registrarc.Registrar)
	}
	if cfg.ConfigSCfg().Enabled {
		server.RegisterHttpFunc(cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)

	// init AnalyzerS
	anz := services.NewAnalyzerService(cfg, server, filterSChan, shdChan, internalAnalyzerSChan, srvDep)
	if anz.ShouldRun() {
		shdWg.Add(1)
		if err := anz.Start(); err != nil {
			log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
		}
	}

	// init CoreSv1

	coreS := services.NewCoreService(cfg, caps, server, internalCoreSv1Chan, anz, cpuProf, shdWg, shdChan, srvDep)
	shdWg.Add(1)
	if err := coreS.Start(); err != nil {
		log.Fatalf("<%s> error received: <%s>, exiting!", utils.InitS, err.Error())
	}
	cS = coreS.GetCoreS()

	// init CacheS
	cacheS, err := initCacheS(internalCacheSChan, server, dmService.GetDM(), shdChan, anz, coreS.GetCoreS().CapsStats)
	if err != nil {
		log.Fatal(err)
	}
	engine.Cache = cacheS

	// init GuardianSv1
	err = initGuardianSv1(internalGuardianSChan, server, anz)
	if err != nil {
		log.Fatal(err)
	}

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, shdChan, shdWg, connManager)
	attrS := services.NewAttributeService(cfg, dmService, cacheS, filterSChan, server, internalAttributeSChan, anz, srvDep)
	dspS := services.NewDispatcherService(cfg, dmService, cacheS, filterSChan, server, internalDispatcherSChan, connManager, anz, srvDep)
	dspH := services.NewRegistrarCService(cfg, server, connManager, anz, srvDep)
	chrS := services.NewChargerService(cfg, dmService, cacheS, filterSChan, server,
		internalChargerSChan, connManager, anz, srvDep)
	tS := services.NewThresholdService(cfg, dmService, cacheS, filterSChan, server, internalThresholdSChan, anz, srvDep)
	stS := services.NewStatService(cfg, dmService, cacheS, filterSChan, server,
		internalStatSChan, connManager, anz, srvDep)
	trS := services.NewTrendService(cfg, dmService, cacheS, filterSChan, server,
		internalTrendSChan, connManager, anz, srvDep)
	sgS := services.NewRankingService(cfg, dmService, cacheS, filterSChan, server,
		internalRankingSChan, connManager, anz, srvDep)
	reS := services.NewResourceService(cfg, dmService, cacheS, filterSChan, server,
		internalResourceSChan, connManager, anz, srvDep)
	routeS := services.NewRouteService(cfg, dmService, cacheS, filterSChan, server,
		internalRouteSChan, connManager, anz, srvDep)

	schS := services.NewSchedulerService(cfg, dmService, cacheS, filterSChan,
		server, internalSchedulerSChan, connManager, anz, srvDep)

	rals := services.NewRalService(cfg, cacheS, server,
		internalRALsChan, internalResponderChan,
		shdChan, connManager, anz, srvDep, filterSChan)

	apiSv1 := services.NewAPIerSv1Service(cfg, dmService, storDBService, filterSChan, server, schS, rals.GetResponder(),
		internalAPIerSv1Chan, connManager, anz, srvDep)

	apiSv2 := services.NewAPIerSv2Service(apiSv1, cfg, server, internalAPIerSv2Chan, anz, srvDep)

	cdrS := services.NewCDRServer(cfg, dmService, storDBService, filterSChan, server, internalCDRServerChan,
		connManager, anz, srvDep)

	smg := services.NewSessionService(cfg, dmService, server, internalSessionSChan, shdChan, connManager, anz, srvDep)

	ldrs := services.NewLoaderService(cfg, dmService, filterSChan, server,
		internalLoaderSChan, connManager, anz, srvDep)

	srvManager.AddServices(gvService, attrS, chrS, tS, stS, trS, sgS, reS, routeS, schS, rals,
		apiSv1, apiSv2, cdrS, smg, coreS,
		services.NewDNSAgent(cfg, filterSChan, shdChan, connManager, srvDep),
		services.NewFreeswitchAgent(cfg, shdChan, connManager, srvDep),
		services.NewKamailioAgent(cfg, shdChan, connManager, srvDep),
		services.NewAsteriskAgent(cfg, shdChan, connManager, srvDep),                    // partial reload
		services.NewRadiusAgent(cfg, filterSChan, shdChan, connManager, srvDep),         // partial reload
		services.NewDiameterAgent(cfg, filterSChan, shdChan, connManager, caps, srvDep), // partial reload
		services.NewHTTPAgent(cfg, filterSChan, server, connManager, srvDep),            // no reload
		ldrs, anz, dspS, dspH, dmService, storDBService,
		services.NewEventExporterService(cfg, filterSChan,
			connManager, server, internalEEsChan, anz, srvDep),
		services.NewEventReaderService(cfg, filterSChan,
			shdChan, connManager, server, internalERsChan, anz, srvDep),
		services.NewSIPAgent(cfg, filterSChan, shdChan, connManager, srvDep),
		services.NewJanusAgent(cfg, filterSChan, server, connManager, srvDep),
	)
	srvManager.StartServices()
	// Start FilterS
	go startFilterService(filterSChan, cacheS, connManager,
		cfg, dmService.GetDM())

	err = initServiceManagerV1(internalServeManagerChan, srvManager, server, anz)
	if err != nil {
		log.Fatal(err)
	}

	// init internalRPCSet to share internal connections among the engine
	engine.IntRPC = engine.NewRPCClientSet()
	engine.IntRPC.AddInternalRPCClient(utils.AnalyzerSv1, internalAnalyzerSChan)
	engine.IntRPC.AddInternalRPCClient(utils.APIerSv1, internalAPIerSv1Chan)
	engine.IntRPC.AddInternalRPCClient(utils.APIerSv2, internalAPIerSv2Chan)
	engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, internalAttributeSChan)
	engine.IntRPC.AddInternalRPCClient(utils.CacheSv1, internalCacheSChan)
	engine.IntRPC.AddInternalRPCClient(utils.CDRsV1, internalCDRServerChan)
	engine.IntRPC.AddInternalRPCClient(utils.CDRsV2, internalCDRServerChan)
	engine.IntRPC.AddInternalRPCClient(utils.ChargerSv1, internalChargerSChan)
	engine.IntRPC.AddInternalRPCClient(utils.GuardianSv1, internalGuardianSChan)
	engine.IntRPC.AddInternalRPCClient(utils.LoaderSv1, internalLoaderSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ResourceSv1, internalResourceSChan)
	engine.IntRPC.AddInternalRPCClient(utils.Responder, internalResponderChan)
	engine.IntRPC.AddInternalRPCClient(utils.SchedulerSv1, internalSchedulerSChan)
	engine.IntRPC.AddInternalRPCClient(utils.SessionSv1, internalSessionSChan)
	engine.IntRPC.AddInternalRPCClient(utils.StatSv1, internalStatSChan)
	engine.IntRPC.AddInternalRPCClient(utils.TrendSv1, internalTrendSChan)
	engine.IntRPC.AddInternalRPCClient(utils.RouteSv1, internalRouteSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ThresholdSv1, internalThresholdSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ServiceManagerV1, internalServeManagerChan)
	engine.IntRPC.AddInternalRPCClient(utils.ConfigSv1, internalConfigChan)
	engine.IntRPC.AddInternalRPCClient(utils.CoreSv1, internalCoreSv1Chan)
	engine.IntRPC.AddInternalRPCClient(utils.RALsV1, internalRALsChan)
	engine.IntRPC.AddInternalRPCClient(utils.EeSv1, internalEEsChan)
	engine.IntRPC.AddInternalRPCClient(utils.ErSv1, internalERsChan)
	engine.IntRPC.AddInternalRPCClient(utils.DispatcherSv1, internalDispatcherSChan)

	err = initConfigSv1(internalConfigChan, server, anz)
	if err != nil {
		log.Fatal(err)
	}

	if *preload != utils.EmptyString {
		runPreload(ldrs, internalLoaderSChan, shdChan)
	}

	// Serve rpc connections
	go startRPC(server, internalResponderChan, internalCDRServerChan,
		internalResourceSChan, internalStatSChan,
		internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
		internalTrendSChan, internalRouteSChan, internalSessionSChan, internalAnalyzerSChan,
		internalDispatcherSChan, internalLoaderSChan, internalRALsChan,
		internalCacheSChan, internalEEsChan, internalERsChan, shdChan)

	if *memProfDir != utils.EmptyString {
		if err := cS.StartMemoryProfiling(cores.MemoryProfilingParams{
			DirPath:      *memProfDir,
			Interval:     *memProfInterval,
			MaxFiles:     *memProfMaxFiles,
			UseTimestamp: *memProfTimestamp,
		}); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> %v", utils.CoreS, err))
			return
		}
		defer cS.StopMemoryProfiling() // safe to ignore error (irrelevant)
	}

	<-shdChan.Done()
	shtdDone := make(chan struct{})
	go func() {
		shdWg.Wait()
		close(shtdDone)
	}()
	select {
	case <-shtdDone:
	case <-time.After(cfg.CoreSCfg().ShutdownTimeout):
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown all subsystems in the given time",
			utils.ServiceManager))
	}

	if *pidFile != utils.EmptyString {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("<CoreS> stopped all components. CGRateS shutdown!")
}
