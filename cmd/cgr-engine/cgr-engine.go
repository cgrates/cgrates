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
	"github.com/cgrates/cgrates/cdrc"
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

func startCdrcs(internalCdrSChan, internalRaterChan, internalDispatcherSChan chan rpcclient.ClientConnector,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	cdrcInitialized := false           // Control whether the cdrc was already initialized (so we don't reload in that case)
	var cdrcChildrenChan chan struct{} // Will use it to communicate with the children of one fork
	intCdrSChan := internalCdrSChan
	if cfg.DispatcherSCfg().Enabled {
		intCdrSChan = internalDispatcherSChan
	}
	for {
		select {
		case <-exitChan: // Stop forking CDRCs
			break
		case <-cfg.ConfigReloads[utils.CDRC]: // Consume the load request and wait for a new one
			if cdrcInitialized {
				utils.Logger.Info("<CDRC> Configuration reload")
				close(cdrcChildrenChan) // Stop all the children of the previous run
			}
			cdrcChildrenChan = make(chan struct{})
		}
		// Start CDRCs
		for _, cdrcCfgs := range cfg.CdrcProfiles {
			var enabledCfgs []*config.CdrcCfg
			for _, cdrcCfg := range cdrcCfgs { // Take a random config out since they should be the same
				if cdrcCfg.Enabled {
					enabledCfgs = append(enabledCfgs, cdrcCfg)
				}
			}
			if len(enabledCfgs) != 0 {
				go startCdrc(intCdrSChan, internalRaterChan, enabledCfgs,
					cfg.GeneralCfg().HttpSkipTlsVerify, filterSChan,
					cdrcChildrenChan, exitChan)
			} else {
				utils.Logger.Info("<CDRC> No enabled CDRC clients")
			}
		}
		cdrcInitialized = true // Initialized
	}
}

// Fires up a cdrc instance
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.ClientConnector, cdrcCfgs []*config.CdrcCfg, httpSkipTlsCheck bool,
	filterSChan chan *engine.FilterS, closeChan chan struct{}, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var err error
	var cdrsConn rpcclient.ClientConnector
	cdrcCfg := cdrcCfgs[0]
	cdrsConn, err = engine.NewRPCPool(rpcclient.PoolFirst, cfg.TlsCfg().ClientKey,
		cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
		cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
		cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
		cdrcCfg.CdrsConns, internalCdrSChan, false)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<CDRC> Could not connect to CDRS via RPC: %s", err.Error()))
		exitChan <- true
		return
	}

	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrsConn, closeChan,
		cfg.GeneralCfg().DefaultTimezone, filterS)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("Cdrc config parsing error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := cdrc.Run(); err != nil {
		utils.Logger.Crit(fmt.Sprintf("Cdrc run error: %s", err.Error()))
		exitChan <- true // If run stopped, something is bad, stop the application
		return
	}
}

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS,
	internalStatSChan, internalResourceSChan, internalRalSChan chan rpcclient.ClientConnector, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, internalStatSChan, internalResourceSChan, internalRalSChan, dm)
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

	go server.ServeJSON(cfg.ListenCfg().RPCJSONListen)
	go server.ServeGOB(cfg.ListenCfg().RPCGOBListen)
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
				if err := config.CgrConfig().V1ReloadConfig(
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
	if *version {
		if vers, err := utils.GetCGRVersion(); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(vers)
		}
		return
	}
	if *pidFile != "" {
		writePid()
	}
	if *singlecpu {
		runtime.GOMAXPROCS(1) // Having multiple cpus may slow down computing due to CPU management, to be reviewed in future Go releases
	}

	exitChan := make(chan bool)
	go singnalHandler(exitChan)

	if *memProfDir != "" {
		go memProfiling(*memProfDir, *memProfInterval, *memProfNrFiles, exitChan)
	}
	cpuProfChanStop := make(chan struct{})
	cpuProfChanDone := make(chan struct{})
	if *cpuProfDir != "" {
		go cpuProfiling(*cpuProfDir, cpuProfChanStop, cpuProfChanDone, exitChan)
	}

	if *scheduledShutdown != "" {
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
	var err error
	// Init config
	cfg, err = config.NewCGRConfigFromPath(*cfgPath)
	if err != nil {
		log.Fatalf("Could not parse config: <%s>", err.Error())
		return
	}
	if *nodeID != "" {
		cfg.GeneralCfg().NodeID = *nodeID
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

	cfg.LazySanityCheck()

	dmService := services.NewDataDBService(cfg)
	storDBService := services.NewStorDBService(cfg)
	if dmService.ShouldRun() { // Some services can run without db, ie:  CDRC
		if err = dmService.Start(); err != nil {
			return
		}
	}
	if storDBService.ShouldRun() {
		if err = storDBService.Start(); err != nil {
			return
		}
	}
	// Done initing DBs
	engine.SetRoundingDecimals(cfg.GeneralCfg().RoundingDecimals)

	// Rpc/http server
	server := utils.NewServer()

	if *httpPprofPath != "" {
		go server.RegisterProfiler(*httpPprofPath)
	}
	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)

	internalServeManagerChan := make(chan rpcclient.ClientConnector, 1)
	internalConfigChan := make(chan rpcclient.ClientConnector, 1)
	internalCoreSv1Chan := make(chan rpcclient.ClientConnector, 1)
	internalCacheSChan := make(chan rpcclient.ClientConnector, 1)
	internalGuardianSChan := make(chan rpcclient.ClientConnector, 1)

	internalCDRServerChan := make(chan rpcclient.RpcClientConnection, 1)   // needed to avod cyclic dependency
	internalAttributeSChan := make(chan rpcclient.RpcClientConnection, 1)  // needed to avod cyclic dependency
	internalDispatcherSChan := make(chan rpcclient.RpcClientConnection, 1) // needed to avod cyclic dependency
	internalSessionSChan := make(chan rpcclient.RpcClientConnection, 1)    // needed to avod cyclic dependency
	internalChargerSChan := make(chan rpcclient.RpcClientConnection, 1)    // needed to avod cyclic dependency
	internalThresholdSChan := make(chan rpcclient.RpcClientConnection, 1)  // needed to avod cyclic dependency
	internalStatSChan := make(chan rpcclient.RpcClientConnection, 1)       // needed to avod cyclic dependency
	internalResourceSChan := make(chan rpcclient.RpcClientConnection, 1)   // needed to avod cyclic dependency

	// init CacheS
	cacheS := initCacheS(internalCacheSChan, server, dmService.GetDM(), exitChan)

	// init GuardianSv1
	initGuardianSv1(internalGuardianSChan, server)

	// init CoreSv1
	initCoreSv1(internalCoreSv1Chan, server)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, exitChan)
	connManager := services.NewConnManagerService(cfg, map[string]chan rpcclient.RpcClientConnection{
		//utils.AnalyzerSv1:  anz.GetIntenternalChan(),
		//utils.ApierV1:      rals.GetAPIv1().GetIntenternalChan(),
		//utils.ApierV2:      rals.GetAPIv2().GetIntenternalChan(),
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaAttributes): internalAttributeSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCaches):     internalCacheSChan,
		//utils.CDRsV1:       cdrS.GetIntenternalChan(),
		//utils.CDRsV2:       cdrS.GetIntenternalChan(),
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaChargers): internalChargerSChan,
		utils.GuardianSv1: internalGuardianSChan,
		//utils.LoaderSv1:    ldrs.GetIntenternalChan(),
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaResources): internalResourceSChan,
		//utils.Responder:    rals.GetResponder().GetIntenternalChan(),
		//utils.SchedulerSv1: schS.GetIntenternalChan(),
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaSessionS): internalSessionSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS):    internalStatSChan,
		//utils.SupplierSv1:      supS.GetIntenternalChan(),
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     internalThresholdSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): internalServeManagerChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         internalConfigChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           internalCoreSv1Chan,
		//utils.RALsV1:           rals.GetIntenternalChan(),
	})

	attrS := services.NewAttributeService(cfg, dmService, cacheS, filterSChan, server, internalAttributeSChan)
	dspS := services.NewDispatcherService(cfg, dmService, cacheS, filterSChan, server, internalAttributeSChan, internalDispatcherSChan)
	chrS := services.NewChargerService(cfg, dmService, cacheS, filterSChan, server,
		internalChargerSChan, connManager.GetConnMgr())
	tS := services.NewThresholdService(cfg, dmService, cacheS, filterSChan, server, internalThresholdSChan)
	stS := services.NewStatService(cfg, dmService, cacheS, filterSChan, server,
		internalStatSChan, connManager.GetConnMgr())
	reS := services.NewResourceService(cfg, dmService, cacheS, filterSChan, server,
		internalResourceSChan, connManager.GetConnMgr())
	supS := services.NewSupplierService(cfg, dmService, cacheS, filterSChan, server,
		attrS.GetIntenternalChan(), stS.GetIntenternalChan(),
		reS.GetIntenternalChan(), dspS.GetIntenternalChan())
	schS := services.NewSchedulerService(cfg, dmService, cacheS, filterSChan, server, internalCDRServerChan, dspS.GetIntenternalChan())
	rals := services.NewRalService(cfg, dmService, storDBService, cacheS, filterSChan, server,
		tS.GetIntenternalChan(), stS.GetIntenternalChan(), internalCacheSChan,
		schS.GetIntenternalChan(), attrS.GetIntenternalChan(), dspS.GetIntenternalChan(),
		schS, exitChan)
	cdrS := services.NewCDRServer(cfg, dmService, storDBService, filterSChan, server, internalCDRServerChan,
		chrS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(),
		attrS.GetIntenternalChan(), tS.GetIntenternalChan(),
		stS.GetIntenternalChan(), dspS.GetIntenternalChan())

	smg := services.NewSessionService(cfg, dmService, server, chrS.GetIntenternalChan(),
		rals.GetResponder().GetIntenternalChan(), reS.GetIntenternalChan(),
		tS.GetIntenternalChan(), stS.GetIntenternalChan(), supS.GetIntenternalChan(),
		attrS.GetIntenternalChan(), cdrS.GetIntenternalChan(), dspS.GetIntenternalChan(), internalSessionSChan, exitChan)
	ldrs := services.NewLoaderService(cfg, dmService, filterSChan, server, internalCacheSChan, dspS.GetIntenternalChan(), exitChan)
	anz := services.NewAnalyzerService(cfg, server, exitChan)

	srvManager.AddServices(connManager, attrS, chrS, tS, stS, reS, supS, schS, rals,
		rals.GetResponder(), rals.GetAPIv1(), rals.GetAPIv2(), cdrS, smg,
		services.NewEventReaderService(cfg, filterSChan, exitChan, connManager.GetConnMgr()),
		services.NewDNSAgent(cfg, filterSChan, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan),
		services.NewFreeswitchAgent(cfg, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan),
		services.NewKamailioAgent(cfg, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan),
		services.NewAsteriskAgent(cfg, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan),              // partial reload
		services.NewRadiusAgent(cfg, filterSChan, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan),   // partial reload
		services.NewDiameterAgent(cfg, filterSChan, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), exitChan), // partial reload
		services.NewHTTPAgent(cfg, filterSChan, smg.GetIntenternalChan(), dspS.GetIntenternalChan(), server),       // no reload
		ldrs, anz, dspS, dmService, storDBService,
	)
	srvManager.StartServices()
	// Start FilterS
	go startFilterService(filterSChan, cacheS, stS.GetIntenternalChan(),
		reS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(),
		cfg, dmService.GetDM(), exitChan)

	initServiceManagerV1(internalServeManagerChan, srvManager, server)

	// init internalRPCSet
	engine.IntRPC = engine.NewRPCClientSet()
	if cfg.DispatcherSCfg().Enabled {
		engine.IntRPC.AddInternalRPCClient(utils.AnalyzerSv1, anz.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ApierV1, rals.GetAPIv1().GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ApierV2, rals.GetAPIv2().GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, attrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.CacheSv1, internalCacheSChan)
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV1, cdrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV2, cdrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ChargerSv1, chrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.GuardianSv1, internalGuardianSChan)
		engine.IntRPC.AddInternalRPCClient(utils.LoaderSv1, ldrs.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ResourceSv1, reS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.Responder, rals.GetResponder().GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.SchedulerSv1, schS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.SessionSv1, smg.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.StatSv1, stS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.SupplierSv1, supS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ThresholdSv1, tS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ServiceManagerV1, internalServeManagerChan)
		engine.IntRPC.AddInternalRPCClient(utils.ConfigSv1, internalConfigChan)
		engine.IntRPC.AddInternalRPCClient(utils.CoreSv1, internalCoreSv1Chan)
		engine.IntRPC.AddInternalRPCClient(utils.RALsV1, rals.GetIntenternalChan())
	}

	initConfigSv1(internalConfigChan, server)

	// Start CDRC components if necessary
	go startCdrcs(cdrS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(), dspS.GetIntenternalChan(), filterSChan, exitChan)

	// Serve rpc connections
	go startRpc(server, rals.GetResponder().GetIntenternalChan(), cdrS.GetIntenternalChan(),
		reS.GetIntenternalChan(), stS.GetIntenternalChan(),
		attrS.GetIntenternalChan(), chrS.GetIntenternalChan(), tS.GetIntenternalChan(),
		supS.GetIntenternalChan(), smg.GetIntenternalChan(), anz.GetIntenternalChan(),
		dspS.GetIntenternalChan(), ldrs.GetIntenternalChan(), rals.GetIntenternalChan(), internalCacheSChan, exitChan)
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
	utils.Logger.Info("Stopped all components. CGRateS shutdown!")
}
