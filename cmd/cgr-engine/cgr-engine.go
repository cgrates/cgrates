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

	"github.com/cgrates/cgrates/analyzers"
	v1 "github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/loaders"
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

func startCdrcs(internalCdrSChan, internalRaterChan, internalDispatcherSChan chan rpcclient.RpcClientConnection,
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
func startCdrc(internalCdrSChan, internalRaterChan chan rpcclient.RpcClientConnection, cdrcCfgs []*config.CdrcCfg, httpSkipTlsCheck bool,
	filterSChan chan *engine.FilterS, closeChan chan struct{}, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS
	var err error
	var cdrsConn rpcclient.RpcClientConnection
	cdrcCfg := cdrcCfgs[0]
	cdrsConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST, cfg.TlsCfg().ClientKey,
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
	internalStatSChan, internalResourceSChan, internalRalSChan chan rpcclient.RpcClientConnection, cfg *config.CGRConfig,
	dm *engine.DataManager, exitChan chan bool) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, internalStatSChan, internalResourceSChan, internalRalSChan, dm)
}

// loaderService will start and register APIs for LoaderService if enabled
func startLoaderS(internalLoaderSChan, cacheSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, dm *engine.DataManager, server *utils.Server,
	filterSChan chan *engine.FilterS, exitChan chan bool) {
	filterS := <-filterSChan
	filterSChan <- filterS

	ldrS := loaders.NewLoaderService(dm, cfg.LoaderCfg(),
		cfg.GeneralCfg().DefaultTimezone, exitChan, filterS, cacheSChan)
	if !ldrS.Enabled() {
		return
	}
	go ldrS.ListenAndServe(exitChan)
	ldrSv1 := v1.NewLoaderSv1(ldrS)
	server.RpcRegister(ldrSv1)
	internalLoaderSChan <- ldrSv1
}

// startDispatcherService fires up the DispatcherS
func startDispatcherService(internalDispatcherSChan, internalAttributeSChan chan rpcclient.RpcClientConnection,
	cfg *config.CGRConfig, cacheS *engine.CacheS, filterSChan chan *engine.FilterS,
	dm *engine.DataManager, server *utils.Server, exitChan chan bool) {
	utils.Logger.Info("Starting CGRateS Dispatcher service.")
	fltrS := <-filterSChan
	filterSChan <- fltrS
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherProfiles)
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherHosts)
	<-cacheS.GetPrecacheChannel(utils.CacheDispatcherFilterIndexes)

	var err error
	var attrSConn *rpcclient.RpcClientPool
	if len(cfg.DispatcherSCfg().AttributeSConns) != 0 { // AttributeS connection init
		attrSConn, err = engine.NewRPCPool(rpcclient.POOL_FIRST,
			cfg.TlsCfg().ClientKey,
			cfg.TlsCfg().ClientCerificate, cfg.TlsCfg().CaCertificate,
			cfg.GeneralCfg().ConnectAttempts, cfg.GeneralCfg().Reconnects,
			cfg.GeneralCfg().ConnectTimeout, cfg.GeneralCfg().ReplyTimeout,
			cfg.DispatcherSCfg().AttributeSConns, internalAttributeSChan, false)
		if err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.DispatcherS, utils.AttributeS, err.Error()))
			exitChan <- true
			return
		}
	}
	dspS, err := dispatchers.NewDispatcherService(dm, cfg, fltrS, attrSConn)
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.DispatcherS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := dspS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.DispatcherS, err.Error()))
		}
		dspS.Shutdown()
		exitChan <- true
		return
	}()

	// for the moment we dispable Apier through dispatcher
	// until we figured out a better sollution in case of gob server
	// server.SetDispatched()

	server.RpcRegister(v1.NewDispatcherSv1(dspS))

	server.RpcRegisterName(utils.ThresholdSv1,
		v1.NewDispatcherThresholdSv1(dspS))

	server.RpcRegisterName(utils.StatSv1,
		v1.NewDispatcherStatSv1(dspS))

	server.RpcRegisterName(utils.ResourceSv1,
		v1.NewDispatcherResourceSv1(dspS))

	server.RpcRegisterName(utils.SupplierSv1,
		v1.NewDispatcherSupplierSv1(dspS))

	server.RpcRegisterName(utils.AttributeSv1,
		v1.NewDispatcherAttributeSv1(dspS))

	server.RpcRegisterName(utils.SessionSv1,
		v1.NewDispatcherSessionSv1(dspS))

	server.RpcRegisterName(utils.ChargerSv1,
		v1.NewDispatcherChargerSv1(dspS))

	server.RpcRegisterName(utils.Responder,
		v1.NewDispatcherResponder(dspS))

	server.RpcRegisterName(utils.CacheSv1,
		v1.NewDispatcherCacheSv1(dspS))

	server.RpcRegisterName(utils.GuardianSv1,
		v1.NewDispatcherGuardianSv1(dspS))

	server.RpcRegisterName(utils.SchedulerSv1,
		v1.NewDispatcherSchedulerSv1(dspS))

	server.RpcRegisterName(utils.CDRsV1,
		v1.NewDispatcherSCDRsV1(dspS))

	server.RpcRegisterName(utils.ConfigSv1,
		v1.NewDispatcherConfigSv1(dspS))

	server.RpcRegisterName(utils.CoreSv1,
		v1.NewDispatcherCoreSv1(dspS))

	server.RpcRegisterName(utils.RALsV1,
		v1.NewDispatcherRALsV1(dspS))

	internalDispatcherSChan <- dspS
}

// startAnalyzerService fires up the AnalyzerS
func startAnalyzerService(internalAnalyzerSChan chan rpcclient.RpcClientConnection,
	server *utils.Server, exitChan chan bool) {
	var err error
	aS, err := analyzers.NewAnalyzerService()
	if err != nil {
		utils.Logger.Crit(fmt.Sprintf("<%s> Could not init, error: %s", utils.AnalyzerS, err.Error()))
		exitChan <- true
		return
	}
	go func() {
		if err := aS.ListenAndServe(exitChan); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Error: %s listening for packets", utils.AnalyzerS, err.Error()))
		}
		aS.Shutdown()
		exitChan <- true
		return
	}()
	aSv1 := v1.NewAnalyzerSv1(aS)
	server.RpcRegister(aSv1)
	internalAnalyzerSChan <- aSv1
}

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan rpcclient.RpcClientConnection,
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

func initGuardianSv1(internalGuardianSChan chan rpcclient.RpcClientConnection, server *utils.Server) {
	grdSv1 := v1.NewGuardianSv1()
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(grdSv1)
	}
	internalGuardianSChan <- grdSv1
}

func initCoreSv1(internalCoreSv1Chan chan rpcclient.RpcClientConnection, server *utils.Server) {
	cSv1 := v1.NewCoreSv1(engine.NewCoreService())
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cSv1)
	}
	internalCoreSv1Chan <- cSv1
}

func initServiceManagerV1(internalServiceManagerChan chan rpcclient.RpcClientConnection,
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
	internalLoaderSChan, internalRALsv1Chan, internalCacheSChan chan rpcclient.RpcClientConnection,
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

func initConfigSv1(internalConfigChan chan rpcclient.RpcClientConnection,
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
	vers := utils.GetCGRVersion()
	if *version {
		fmt.Println(vers)
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

	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	var dm *engine.DataManager
	if needsDB := cfg.RalsCfg().Enabled || cfg.SchedulerCfg().Enabled || cfg.ChargerSCfg().Enabled ||
		cfg.AttributeSCfg().Enabled || cfg.ResourceSCfg().Enabled || cfg.StatSCfg().Enabled ||
		cfg.ThresholdSCfg().Enabled || cfg.SupplierSCfg().Enabled || cfg.DispatcherSCfg().Enabled; needsDB ||
		cfg.SessionSCfg().Enabled { // Some services can run without db, ie:  CDRC
		dm, err = engine.ConfigureDataStorage(cfg.DataDbCfg().DataDbType,
			cfg.DataDbCfg().DataDbHost, cfg.DataDbCfg().DataDbPort,
			cfg.DataDbCfg().DataDbName, cfg.DataDbCfg().DataDbUser,
			cfg.DataDbCfg().DataDbPass, cfg.GeneralCfg().DBDataEncoding,
			cfg.CacheCfg(), cfg.DataDbCfg().DataDbSentinelName)
		if needsDB && err != nil { // Cannot configure getter database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		} else if cfg.SessionSCfg().Enabled && err != nil {
			utils.Logger.Warning(fmt.Sprintf("Could not configure dataDb: %s.Some SessionS APIs will not work", err))
		} else {
			defer dm.DataDB().Close()
			engine.SetDataStorage(dm)
			if err := engine.CheckVersions(dm.DataDB()); err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}
	if cfg.RalsCfg().Enabled || cfg.CdrsCfg().Enabled {
		storDb, err := engine.ConfigureStorStorage(cfg.StorDbCfg().StorDBType,
			cfg.StorDbCfg().StorDBHost, cfg.StorDbCfg().StorDBPort,
			cfg.StorDbCfg().StorDBName, cfg.StorDbCfg().StorDBUser,
			cfg.StorDbCfg().StorDBPass, cfg.GeneralCfg().DBDataEncoding,
			cfg.StorDbCfg().StorDBMaxOpenConns, cfg.StorDbCfg().StorDBMaxIdleConns,
			cfg.StorDbCfg().StorDBConnMaxLifetime, cfg.StorDbCfg().StorDBStringIndexedFields,
			cfg.StorDbCfg().StorDBPrefixIndexedFields)
		if err != nil { // Cannot configure logger database, show stopper
			utils.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
			return
		}
		defer storDb.Close()
		// loadDb,cdrDb and storDb are all mapped on the same stordb storage
		loadDb = storDb.(engine.LoadStorage)
		cdrDb = storDb.(engine.CdrStorage)
		engine.SetCdrStorage(cdrDb)
		if err := engine.CheckVersions(storDb); err != nil {
			fmt.Println(err.Error())
			return
		}
	}
	// Done initing DBs
	engine.SetRoundingDecimals(cfg.GeneralCfg().RoundingDecimals)
	engine.SetRpSubjectPrefixMatching(cfg.RalsCfg().RpSubjectPrefixMatching)

	// Rpc/http server
	server := utils.NewServer()

	if *httpPprofPath != "" {
		go server.RegisterProfiler(*httpPprofPath)
	}
	// Async starts here, will follow cgrates.json start order

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)
	internalDispatcherSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalAnalyzerSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalLoaderSChan := make(chan rpcclient.RpcClientConnection, 1)

	internalServeManagerChan := make(chan rpcclient.RpcClientConnection, 1)
	internalConfigChan := make(chan rpcclient.RpcClientConnection, 1)
	internalCoreSv1Chan := make(chan rpcclient.RpcClientConnection, 1)
	internalCacheSChan := make(chan rpcclient.RpcClientConnection, 1)
	internalGuardianSChan := make(chan rpcclient.RpcClientConnection, 1)

	// init CacheS
	cacheS := initCacheS(internalCacheSChan, server, dm, exitChan)

	// init GuardianSv1
	initGuardianSv1(internalGuardianSChan, server)

	// init CoreSv1
	initCoreSv1(internalCoreSv1Chan, server)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, dm, cdrDb,
		loadDb, filterSChan, server, internalDispatcherSChan, exitChan)

	attrS := services.NewAttributeService(cfg, dm, cacheS, filterSChan, server)
	chrS := services.NewChargerService(cfg, dm, cacheS, filterSChan, server,
		attrS.GetIntenternalChan(), internalDispatcherSChan)
	tS := services.NewThresholdService(cfg, dm, cacheS, filterSChan, server)
	stS := services.NewStatService(cfg, dm, cacheS, filterSChan, server,
		tS.GetIntenternalChan(), internalDispatcherSChan)
	reS := services.NewResourceService(cfg, dm, cacheS, filterSChan, server,
		tS.GetIntenternalChan(), internalDispatcherSChan)
	supS := services.NewSupplierService(cfg, dm, cacheS, filterSChan, server,
		attrS.GetIntenternalChan(), stS.GetIntenternalChan(),
		reS.GetIntenternalChan(), internalDispatcherSChan)
	schS := services.NewSchedulerService(cfg, dm, cacheS, server, internalDispatcherSChan)
	rals := services.NewRalService(cfg, dm, cdrDb, loadDb, cacheS, filterSChan, server,
		tS.GetIntenternalChan(), stS.GetIntenternalChan(), internalCacheSChan,
		schS.GetIntenternalChan(), attrS.GetIntenternalChan(), internalDispatcherSChan,
		schS, exitChan)
	cdrS := services.NewCDRServer(cfg, dm, cdrDb, filterSChan, server,
		chrS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(),
		attrS.GetIntenternalChan(), tS.GetIntenternalChan(),
		stS.GetIntenternalChan(), internalDispatcherSChan)
	schS.SetCdrsConns(cdrS.GetIntenternalChan())

	smg := services.NewSessionService(cfg, dm, server, cdrS.GetIntenternalChan(),
		rals.GetResponder().GetIntenternalChan(), reS.GetIntenternalChan(),
		tS.GetIntenternalChan(), stS.GetIntenternalChan(), supS.GetIntenternalChan(),
		attrS.GetIntenternalChan(), cdrS.GetIntenternalChan(), internalDispatcherSChan, exitChan)

	srvManager.AddServices(attrS, chrS, tS, stS, reS, supS, schS, rals,
		rals.GetResponder(), rals.GetAPIv1(), rals.GetAPIv2(), cdrS, smg,
		services.NewEventReaderService(cfg, filterSChan, smg.GetIntenternalChan(), internalDispatcherSChan, exitChan),
		services.NewDNSAgent(cfg, filterSChan, smg.GetIntenternalChan(), internalDispatcherSChan, exitChan),
		services.NewFreeswitchAgent(cfg, smg.GetIntenternalChan(), internalDispatcherSChan, exitChan),
		services.NewKamailioAgent(cfg, smg.GetIntenternalChan(), internalDispatcherSChan, exitChan),
	)
	/*
		services.NewAsteriskAgent(), // partial reload
		services.NewRadiusAgent(),   // partial reload
		services.NewDiameterAgent(), // partial reload
		services.NewHTTPAgent(),     // no reload
	*/

	srvManager.StartServices()

	// Start FilterS
	go startFilterService(filterSChan, cacheS, stS.GetIntenternalChan(),
		reS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(),
		cfg, dm, exitChan)

	initServiceManagerV1(internalServeManagerChan, srvManager, server)

	// init internalRPCSet
	engine.IntRPC = engine.NewRPCClientSet()
	if cfg.DispatcherSCfg().Enabled {
		engine.IntRPC.AddInternalRPCClient(utils.AnalyzerSv1, internalAnalyzerSChan)
		engine.IntRPC.AddInternalRPCClient(utils.ApierV1, rals.GetAPIv1().GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ApierV2, rals.GetAPIv2().GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.AttributeSv1, attrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.CacheSv1, internalCacheSChan)
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV1, cdrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.CDRsV2, cdrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.ChargerSv1, chrS.GetIntenternalChan())
		engine.IntRPC.AddInternalRPCClient(utils.GuardianSv1, internalGuardianSChan)
		engine.IntRPC.AddInternalRPCClient(utils.LoaderSv1, internalLoaderSChan)
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
	go startCdrcs(cdrS.GetIntenternalChan(), rals.GetResponder().GetIntenternalChan(), internalDispatcherSChan, filterSChan, exitChan)

	if cfg.DispatcherSCfg().Enabled {
		go startDispatcherService(internalDispatcherSChan,
			attrS.GetIntenternalChan(), cfg, cacheS, filterSChan,
			dm, server, exitChan)
	}

	if cfg.AnalyzerSCfg().Enabled {
		go startAnalyzerService(internalAnalyzerSChan, server, exitChan)
	}

	go startLoaderS(internalLoaderSChan, internalCacheSChan, cfg, dm, server, filterSChan, exitChan)

	// Serve rpc connections
	go startRpc(server, rals.GetResponder().GetIntenternalChan(), cdrS.GetIntenternalChan(),
		reS.GetIntenternalChan(), stS.GetIntenternalChan(),
		attrS.GetIntenternalChan(), chrS.GetIntenternalChan(), tS.GetIntenternalChan(),
		supS.GetIntenternalChan(), smg.GetIntenternalChan(), internalAnalyzerSChan,
		internalDispatcherSChan, internalLoaderSChan, rals.GetIntenternalChan(), internalCacheSChan, exitChan)
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
