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
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"syscall"
	"time"

	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/services"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

var (
	cgrEngineFlags    = flag.NewFlagSet("cgr-engine", flag.ContinueOnError)
	cfgPath           = cgrEngineFlags.String("config_path", utils.CONFIG_PATH, "Configuration directory path.")
	version           = cgrEngineFlags.Bool("version", false, "Prints the application version.")
	checkConfig       = cgrEngineFlags.Bool("check_config", false, "Verify the config without starting the engine")
	pidFile           = cgrEngineFlags.String("pid", "", "Write pid file")
	httpPprofPath     = cgrEngineFlags.String("httprof_path", "", "http address used for program profiling")
	cpuProfDir        = cgrEngineFlags.String("cpuprof_dir", "", "write cpu profile to files")
	memProfDir        = cgrEngineFlags.String("memprof_dir", "", "write memory profile to file")
	memProfInterval   = cgrEngineFlags.Duration("memprof_interval", 5*time.Second, "Time betwen memory profile saves")
	memProfNrFiles    = cgrEngineFlags.Int("memprof_nrfiles", 1, "Number of memory profile to write")
	scheduledShutdown = cgrEngineFlags.String("scheduled_shutdown", "", "shutdown the engine after this duration")
	singlecpu         = cgrEngineFlags.Bool("singlecpu", false, "Run on single CPU core")
	syslogger         = cgrEngineFlags.String("logger", "", "logger <*syslog|*stdout>")
	nodeID            = cgrEngineFlags.String("node_id", "", "The node ID of the engine")
	logLevel          = cgrEngineFlags.Int("log_level", -1, "Log level (0-emergency to 7-debug)")

	cfg *config.CGRConfig
)

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS, connMgr *engine.ConnManager, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, connMgr, dm)
}

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan rpcclient.ClientConnector,
	server *utils.Server, dm *engine.DataManager, exitChan chan bool) (chS *engine.CacheS) {
	chS = engine.NewCacheS(cfg, dm)
	go func() {
		if err := chS.Precache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			exitChan <- true
		}
	}()

	chSv1 := v1.NewCacheSv1(chS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(chSv1)
	}
	internalCacheSChan <- chS
	return
}

func initGuardianSv1(internalGuardianSChan chan rpcclient.ClientConnector, server *utils.Server) {
	grdSv1 := v1.NewGuardianSv1()
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(grdSv1)
	}
	internalGuardianSChan <- grdSv1
}

func initCoreSv1(internalCoreSv1Chan chan rpcclient.ClientConnector, server *utils.Server) {
	cSv1 := v1.NewCoreSv1(engine.NewCoreService())
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cSv1)
	}
	internalCoreSv1Chan <- cSv1
}

func initServiceManagerV1(internalServiceManagerChan chan rpcclient.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *utils.Server) {
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	}
	internalServiceManagerChan <- srvMngr
}

func startRpc(server *utils.Server, internalRaterChan,
	internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalSuplSChan,
	internalSMGChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalRALsv1Chan, internalCacheSChan chan rpcclient.ClientConnector,
	exitChan chan bool) {
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
		}
	} else {
		select {
		case dispatcherS := <-internalDispatcherSChan:
			internalDispatcherSChan <- dispatcherS
		}
	}

	go server.ServeJSON(cfg.ListenCfg().RPCJSONListen, exitChan)
	go server.ServeGOB(cfg.ListenCfg().RPCGOBListen, exitChan)
	go server.ServeHTTP(
		cfg.ListenCfg().HTTPListen,
		cfg.HTTPCfg().HTTPJsonRPCURL,
		cfg.HTTPCfg().HTTPWSURL,
		cfg.HTTPCfg().HTTPUseBasicAuth,
		cfg.HTTPCfg().HTTPAuthUsers,
		exitChan,
	)
	if cfg.ListenCfg().RPCGOBTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeGOBTLS(
				cfg.ListenCfg().RPCGOBTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
				exitChan,
			)
		}
	}
	if cfg.ListenCfg().RPCJSONTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeJSONTLS(
				cfg.ListenCfg().RPCJSONTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
				exitChan,
			)
		}
	}
	if cfg.ListenCfg().HTTPTLSListen != "" {
		if cfg.TlsCfg().ServerCerificate == "" || cfg.TlsCfg().ServerKey == "" {
			utils.Logger.Warning("WARNING: missing TLS certificate/key file!")
		} else {
			go server.ServeHTTPTLS(
				cfg.ListenCfg().HTTPTLSListen,
				cfg.TlsCfg().ServerCerificate,
				cfg.TlsCfg().ServerKey,
				cfg.TlsCfg().CaCertificate,
				cfg.TlsCfg().ServerPolicy,
				cfg.TlsCfg().ServerName,
				cfg.HTTPCfg().HTTPJsonRPCURL,
				cfg.HTTPCfg().HTTPWSURL,
				cfg.HTTPCfg().HTTPUseBasicAuth,
				cfg.HTTPCfg().HTTPAuthUsers,
				exitChan,
			)
		}
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

// initLogger will initialize syslog writter, needs to be called after config init
func initLogger(cfg *config.CGRConfig) error {
	sylogger := cfg.GeneralCfg().Logger
	if *syslogger != "" { // Modify the log level if provided by command arguments
		sylogger = *syslogger
	}
	err := utils.Newlogger(sylogger, cfg.GeneralCfg().NodeID)
	if err != nil {
		return err
	}
	return nil
}

func initConfigSv1(internalConfigChan chan rpcclient.ClientConnector,
	server *utils.Server) {
	cfgSv1 := v1.NewConfigSv1(cfg)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	internalConfigChan <- cfgSv1
}

func memProfFile(memProfPath string) bool {
	f, err := os.Create(memProfPath)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not create memory profile file: %s", err))
		return false
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<memProfile>could not write memory profile: %s", err))
		f.Close()
		return false
	}
	f.Close()
	return true
}

func memProfiling(memProfDir string, interval time.Duration, nrFiles int, exitChan chan bool) {
	for i := 1; ; i++ {
		time.Sleep(interval)
		memPath := path.Join(memProfDir, fmt.Sprintf("mem%v.prof", i))
		if !memProfFile(memPath) {
			exitChan <- true
		}
		if i%nrFiles == 0 {
			i = 0 // reset the counting
		}
	}
}

func cpuProfiling(cpuProfDir string, stopChan, doneChan chan struct{}, exitChan chan bool) {
	cpuPath := path.Join(cpuProfDir, "cpu.prof")
	f, err := os.Create(cpuPath)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<cpuProfiling>could not create cpu profile file: %s", err))
		exitChan <- true
		return
	}
	pprof.StartCPUProfile(f)
	<-stopChan
	pprof.StopCPUProfile()
	f.Close()
	doneChan <- struct{}{}
}

func singnalHandler(exitChan chan bool) {
	shutdownSignal := make(chan os.Signal)
	reloadSignal := make(chan os.Signal)
	signal.Notify(shutdownSignal, os.Interrupt,
		syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	for {
		select {
		case <-shutdownSignal:
			exitChan <- true
		case <-reloadSignal:
			//  do it in it's own gorutine in order to not block the signal handler with the reload functionality
			go func() {
				var reply string
				if err := config.CgrConfig().V1ReloadConfigFromPath(
					&config.ConfigReloadWithArgDispatcher{
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

func main() {
	if err := cgrEngineFlags.Parse(os.Args[1:]); err != nil {
		return
	}
	vers, err := utils.GetCGRVersion()
	if err != nil {
		fmt.Println(err)
		return
	}
	goVers := runtime.Version()
	if *version {
		fmt.Println(vers)
		return
	}
	if *pidFile != utils.EmptyString {
		writePid()
	}
	if *singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	exitChan := make(chan bool)
	go singnalHandler(exitChan)

	if *memProfDir != utils.EmptyString {
		go memProfiling(*memProfDir, *memProfInterval, *memProfNrFiles, exitChan)
	}
	cpuProfChanStop := make(chan struct{})
	cpuProfChanDone := make(chan struct{})
	if *cpuProfDir != utils.EmptyString {
		go cpuProfiling(*cpuProfDir, cpuProfChanStop, cpuProfChanDone, exitChan)
	}

	if *scheduledShutdown != utils.EmptyString {
		shutdownDur, err := utils.ParseDurationWithNanosecs(*scheduledShutdown)
		if err != nil {
			log.Fatal(err)
		}
		go func() { // Schedule shutdown
			time.Sleep(shutdownDur)
			exitChan <- true
			return
		}()
	}
	// Init config
	cfg, err = config.NewCGRConfigFromPath(*cfgPath)
	if err != nil {
		log.Fatalf("Could not parse config: <%s>", err.Error())
		return
	}
	if *nodeID != utils.EmptyString {
		cfg.GeneralCfg().NodeID = *nodeID
	}
	if *checkConfig {
		return
	}
	config.SetCgrConfig(cfg) // Share the config object

	// init syslog
	if err = initLogger(cfg); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err.Error())
		return
	}
	lgLevel := cfg.GeneralCfg().LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)
	// init the concurrentRequests
	utils.ConReqs = utils.NewConReqs(cfg.GeneralCfg().ConcurrentRequests, cfg.GeneralCfg().ConcurrentStrategy)
	utils.Logger.Info(fmt.Sprintf("<CoreS> starting version <%s><%s>", vers, goVers))
	cfg.LazySanityCheck()

	// init the channel here because we need to pass them to connManager
	internalServeManagerChan := make(chan rpcclient.ClientConnector, 1)
	internalConfigChan := make(chan rpcclient.ClientConnector, 1)
	internalCoreSv1Chan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalGuardianSChan := make(chan rpcclient.ClientConnector, 1)
	internalAnalyzerSChan := make(chan rpcclient.ClientConnector, 1)
	internalCDRServerChan := make(chan rpcclient.ClientConnector, 1)
	internalAttributeSChan := make(chan rpcclient.ClientConnector, 1)
	internalDispatcherSChan := make(chan rpcclient.ClientConnector, 1)
	internalSessionSChan := make(chan rpcclient.ClientConnector, 1)
	internalChargerSChan := make(chan rpcclient.ClientConnector, 1)
	internalThresholdSChan := make(chan rpcclient.ClientConnector, 1)
	internalStatSChan := make(chan rpcclient.ClientConnector, 1)
	internalResourceSChan := make(chan rpcclient.ClientConnector, 1)
	internalSupplierSChan := make(chan rpcclient.ClientConnector, 1)
	internalSchedulerSChan := make(chan rpcclient.ClientConnector, 1)
	internalRALsChan := make(chan rpcclient.ClientConnector, 1)
	internalResponderChan := make(chan rpcclient.ClientConnector, 1)
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	internalAPIerSv2Chan := make(chan rpcclient.ClientConnector, 1)
	internalLoaderSChan := make(chan rpcclient.ClientConnector, 1)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	connManager := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAnalyzer):       internalAnalyzerSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaApier):          internalAPIerSv1Chan,
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
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS):          internalStatSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSuppliers):      internalSupplierSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     internalThresholdSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): internalServeManagerChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         internalConfigChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           internalCoreSv1Chan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):           internalRALsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers):    internalDispatcherSChan,
	})

	dmService := services.NewDataDBService(cfg, connManager)
	storDBService := services.NewStorDBService(cfg)
	if dmService.ShouldRun() { // Some services can run without db, ie:  ERs
		if err = dmService.Start(); err != nil {
			return
		}
	}
	// Done initing DBs
	engine.SetRoundingDecimals(cfg.GeneralCfg().RoundingDecimals)
	engine.SetFailedPostCacheTTL(cfg.GeneralCfg().FailedPostsTTL)

	// Rpc/http server
	server := utils.NewServer()

	if *httpPprofPath != "" {
		go server.RegisterProfiler(*httpPprofPath)
	}
	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)

	// init CacheS
	cacheS := initCacheS(internalCacheSChan, server, dmService.GetDM(), exitChan)

	// init GuardianSv1
	initGuardianSv1(internalGuardianSChan, server)

	// init CoreSv1
	initCoreSv1(internalCoreSv1Chan, server)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, exitChan)
	attrS := services.NewAttributeService(cfg, dmService, cacheS, filterSChan, server, internalAttributeSChan)
	dspS := services.NewDispatcherService(cfg, dmService, cacheS, filterSChan, server, internalDispatcherSChan, connManager)
	chrS := services.NewChargerService(cfg, dmService, cacheS, filterSChan, server,
		internalChargerSChan, connManager)
	tS := services.NewThresholdService(cfg, dmService, cacheS, filterSChan, server, internalThresholdSChan)
	stS := services.NewStatService(cfg, dmService, cacheS, filterSChan, server,
		internalStatSChan, connManager)
	reS := services.NewResourceService(cfg, dmService, cacheS, filterSChan, server,
		internalResourceSChan, connManager)
	supS := services.NewSupplierService(cfg, dmService, cacheS, filterSChan, server,
		internalSupplierSChan, connManager)

	schS := services.NewSchedulerService(cfg, dmService, cacheS, filterSChan,
		server, internalSchedulerSChan, connManager)

	rals := services.NewRalService(cfg, cacheS, server,
		internalRALsChan, internalResponderChan,
		exitChan, connManager)

	apiSv1 := services.NewAPIerSv1Service(cfg, dmService, storDBService, filterSChan, server, schS, rals.GetResponderService(),
		internalAPIerSv1Chan, connManager)

	apiSv2 := services.NewAPIerSv2Service(apiSv1, cfg, server, internalAPIerSv2Chan)

	cdrS := services.NewCDRServer(cfg, dmService, storDBService, filterSChan, server, internalCDRServerChan,
		connManager)

	smg := services.NewSessionService(cfg, dmService, server, internalSessionSChan, exitChan, connManager)

	ldrs := services.NewLoaderService(cfg, dmService, filterSChan, server, exitChan,
		internalLoaderSChan, connManager)
	anz := services.NewAnalyzerService(cfg, server, exitChan, internalAnalyzerSChan)

	srvManager.AddServices(attrS, chrS, tS, stS, reS, supS, schS, rals,
		rals.GetResponder(), apiSv1, apiSv2, cdrS, smg,
		services.NewEventReaderService(cfg, filterSChan, exitChan, connManager),
		services.NewDNSAgent(cfg, filterSChan, exitChan, connManager),
		services.NewFreeswitchAgent(cfg, exitChan, connManager),
		services.NewKamailioAgent(cfg, exitChan, connManager),
		services.NewAsteriskAgent(cfg, exitChan, connManager),              // partial reload
		services.NewRadiusAgent(cfg, filterSChan, exitChan, connManager),   // partial reload
		services.NewDiameterAgent(cfg, filterSChan, exitChan, connManager), // partial reload
		services.NewHTTPAgent(cfg, filterSChan, server, connManager),       // no reload
		ldrs, anz, dspS, dmService, storDBService,
	)
	srvManager.StartServices()
	// Start FilterS
	go startFilterService(filterSChan, cacheS, connManager,
		cfg, dmService.GetDM(), exitChan)

	initServiceManagerV1(internalServeManagerChan, srvManager, server)

	// init internalRPCSet because we can have double connections in rpc_conns and one of it could be *internal
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
	engine.IntRPC.AddInternalRPCClient(utils.SupplierSv1, internalSupplierSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ThresholdSv1, internalThresholdSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ServiceManagerV1, internalServeManagerChan)
	engine.IntRPC.AddInternalRPCClient(utils.ConfigSv1, internalConfigChan)
	engine.IntRPC.AddInternalRPCClient(utils.CoreSv1, internalCoreSv1Chan)
	engine.IntRPC.AddInternalRPCClient(utils.RALsV1, internalRALsChan)

	initConfigSv1(internalConfigChan, server)

	// Serve rpc connections
	go startRpc(server, internalResponderChan, internalCDRServerChan,
		internalResourceSChan, internalStatSChan,
		internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
		internalSupplierSChan, internalSessionSChan, internalAnalyzerSChan,
		internalDispatcherSChan, internalLoaderSChan, internalRALsChan, internalCacheSChan, exitChan)
	<-exitChan

	if *cpuProfDir != "" { // wait to end cpuProfiling
		cpuProfChanStop <- struct{}{}
		<-cpuProfChanDone
	}
	if *memProfDir != "" { // write last memory profiling
		memProfFile(path.Join(*memProfDir, "mem_final.prof"))
	}
	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			utils.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	utils.Logger.Info("<CoreS> stopped all components. CGRateS shutdown!")
}
