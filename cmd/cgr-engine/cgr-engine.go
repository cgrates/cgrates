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
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cgrates/cgrates/cores"
	"github.com/cgrates/cgrates/dispatcherh"
	"github.com/cgrates/cgrates/loaders"

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
	preload           = cgrEngineFlags.String("preload", "", "LoaderIDs used to load the data before the engine starts")

	cfg *config.CGRConfig
)

// startFilterService fires up the FilterS
func startFilterService(filterSChan chan *engine.FilterS, cacheS *engine.CacheS, connMgr *engine.ConnManager, cfg *config.CGRConfig,
	dm *engine.DataManager) {
	<-cacheS.GetPrecacheChannel(utils.CacheFilters)
	filterSChan <- engine.NewFilterS(cfg, connMgr, dm)
}

// initCacheS inits the CacheS and starts precaching as well as populating internal channel for RPC conns
func initCacheS(internalCacheSChan chan rpcclient.ClientConnector,
	server *cores.Server, dm *engine.DataManager, shdChan *utils.SyncedChan,
	anz *services.AnalyzerService,
	cpS *engine.CapsStats) (chS *engine.CacheS) {
	chS = engine.NewCacheS(cfg, dm, cpS)
	go func() {
		if err := chS.Precache(); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> could not init, error: %s", utils.CacheS, err.Error()))
			shdChan.CloseOnce()
		}
	}()

	chSv1 := v1.NewCacheSv1(chS)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(chSv1)
	}
	var rpc rpcclient.ClientConnector = chS
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.CacheS)
	}
	internalCacheSChan <- rpc
	return
}

func initGuardianSv1(internalGuardianSChan chan rpcclient.ClientConnector, server *cores.Server,
	anz *services.AnalyzerService) {
	grdSv1 := v1.NewGuardianSv1()
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(grdSv1)
	}
	var rpc rpcclient.ClientConnector = grdSv1
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.GuardianS)
	}
	internalGuardianSChan <- rpc
}

func initServiceManagerV1(internalServiceManagerChan chan rpcclient.ClientConnector,
	srvMngr *servmanager.ServiceManager, server *cores.Server,
	anz *services.AnalyzerService) {
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(v1.NewServiceManagerV1(srvMngr))
	}
	var rpc rpcclient.ClientConnector = srvMngr
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ServiceManager)
	}
	internalServiceManagerChan <- rpc
}

func initConfigSv1(internalConfigChan chan rpcclient.ClientConnector,
	server *cores.Server, anz *services.AnalyzerService) {
	cfgSv1 := v1.NewConfigSv1(cfg)
	if !cfg.DispatcherSCfg().Enabled {
		server.RpcRegister(cfgSv1)
	}
	var rpc rpcclient.ClientConnector = cfgSv1
	if anz.IsRunning() {
		rpc = anz.GetAnalyzerS().NewAnalyzerConnector(rpc, utils.MetaInternal, utils.EmptyString, utils.ConfigSv1)
	}
	internalConfigChan <- rpc
}

func startRPC(server *cores.Server, internalRaterChan,
	internalCdrSChan, internalRsChan, internalStatSChan,
	internalAttrSChan, internalChargerSChan, internalThdSChan, internalSuplSChan,
	internalSMGChan, internalAnalyzerSChan, internalDispatcherSChan,
	internalLoaderSChan, internalRALsv1Chan, internalCacheSChan,
	internalEEsChan, internalRateSChan, internalActionSChan chan rpcclient.ClientConnector,
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
		case rateS := <-internalRateSChan:
			internalRateSChan <- rateS
		case actionS := <-internalActionSChan:
			internalActionSChan <- actionS
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
	if cfg.ListenCfg().RPCGOBTLSListen != "" {
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
	if cfg.ListenCfg().RPCJSONTLSListen != "" {
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
	if cfg.ListenCfg().HTTPTLSListen != "" {
		go server.ServeHTTPTLS(
			cfg.ListenCfg().HTTPTLSListen,
			cfg.TLSCfg().ServerCerificate,
			cfg.TLSCfg().ServerKey,
			cfg.TLSCfg().CaCertificate,
			cfg.TLSCfg().ServerPolicy,
			cfg.TLSCfg().ServerName,
			cfg.HTTPCfg().HTTPJsonRPCURL,
			cfg.HTTPCfg().HTTPWSURL,
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

func memProfiling(memProfDir string, interval time.Duration, nrFiles int, shdWg *sync.WaitGroup, shdChan *utils.SyncedChan) {
	tm := time.NewTimer(interval)
	for i := 1; ; i++ {
		select {
		case <-shdChan.Done():
			tm.Stop()
			shdWg.Done()
			return
		case <-tm.C:
		}
		time.Sleep(interval)
		memPath := path.Join(memProfDir, fmt.Sprintf("mem%v.prof", i))
		if !memProfFile(memPath) {
			shdChan.CloseOnce()
			shdWg.Done()
			return
		}
		if i%nrFiles == 0 {
			i = 0 // reset the counting
		}
		tm.Reset(interval)
	}
}

func startCPUProfiling(cpuProfDir string) (f io.WriteCloser, err error) {
	cpuPath := path.Join(cpuProfDir, "cpu.prof")
	if f, err = os.Create(cpuPath); err != nil {
		utils.Logger.Crit(fmt.Sprintf("<cpuProfiling>could not create cpu profile file: %s", err))
		return
	}
	pprof.StartCPUProfile(f)
	return
}

func stopCPUProfiling(f io.Closer) {
	pprof.StopCPUProfile()
	f.Close()
}

func singnalHandler(shdWg *sync.WaitGroup, shdChan *utils.SyncedChan) {
	shutdownSignal := make(chan os.Signal)
	reloadSignal := make(chan os.Signal)
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

func runPreload(loader *services.LoaderService, internalLoaderSChan chan rpcclient.ClientConnector,
	shdChan *utils.SyncedChan) {
	if !cfg.LoaderCfg().Enabled() {
		utils.Logger.Err(fmt.Sprintf("<%s> not enabled but required by preload mechanism", utils.LoaderS))
		shdChan.CloseOnce()
		return
	}

	ldrs := <-internalLoaderSChan
	internalLoaderSChan <- ldrs

	var reply string
	for _, loaderID := range strings.Split(*preload, utils.FIELDS_SEP) {
		if err := loader.GetLoaderS().V1Load(&loaders.ArgsProcessFolder{
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

	shdWg := new(sync.WaitGroup)
	shdChan := utils.NewSyncedChan()

	shdWg.Add(1)
	go singnalHandler(shdWg, shdChan)

	if *memProfDir != utils.EmptyString {
		shdWg.Add(1)
		go memProfiling(*memProfDir, *memProfInterval, *memProfNrFiles, shdWg, shdChan)
	}
	if *cpuProfDir != utils.EmptyString {
		f, err := startCPUProfiling(*cpuProfDir)
		if err != nil {
			return
		}
		defer stopCPUProfiling(f)
	}

	if *scheduledShutdown != utils.EmptyString {
		shutdownDur, err := utils.ParseDurationWithNanosecs(*scheduledShutdown)
		if err != nil {
			log.Fatal(err)
		}
		shdWg.Add(1)
		go func() { // Schedule shutdown
			tm := time.NewTimer(shutdownDur)
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
		return
	}
	if *checkConfig {
		if err := cfg.CheckConfigSanity(); err != nil {
			fmt.Println(err)
		}
		return
	}

	if *nodeID != utils.EmptyString {
		cfg.GeneralCfg().NodeID = *nodeID
	}

	config.SetCgrConfig(cfg) // Share the config object

	// init syslog
	if utils.Logger, err = utils.Newlogger(utils.FirstNonEmpty(*syslogger,
		cfg.GeneralCfg().Logger), cfg.GeneralCfg().NodeID); err != nil {
		log.Fatalf("Could not initialize syslog connection, err: <%s>", err.Error())
		return
	}
	lgLevel := cfg.GeneralCfg().LogLevel
	if *logLevel != -1 { // Modify the log level if provided by command arguments
		lgLevel = *logLevel
	}
	utils.Logger.SetLogLevel(lgLevel)
	// init the concurrentRequests
	cncReqsLimit := cfg.CoreSCfg().Caps
	if utils.ConcurrentReqsLimit != 0 { // used as shared variable
		cncReqsLimit = utils.ConcurrentReqsLimit
	}
	cncReqsStrategy := cfg.CoreSCfg().CapsStrategy
	if len(utils.ConcurrentReqsStrategy) != 0 {
		cncReqsStrategy = utils.ConcurrentReqsStrategy
	}
	caps := engine.NewCaps(cncReqsLimit, cncReqsStrategy)
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
	internalRouteSChan := make(chan rpcclient.ClientConnector, 1)
	internalSchedulerSChan := make(chan rpcclient.ClientConnector, 1)
	internalRALsChan := make(chan rpcclient.ClientConnector, 1)
	internalResponderChan := make(chan rpcclient.ClientConnector, 1)
	internalAPIerSv1Chan := make(chan rpcclient.ClientConnector, 1)
	internalAPIerSv2Chan := make(chan rpcclient.ClientConnector, 1)
	internalLoaderSChan := make(chan rpcclient.ClientConnector, 1)
	internalEEsChan := make(chan rpcclient.ClientConnector, 1)
	internalRateSChan := make(chan rpcclient.ClientConnector, 1)
	internalActionSChan := make(chan rpcclient.ClientConnector, 1)

	// initialize the connManager before creating the DMService
	// because we need to pass the connection to it
	connManager := engine.NewConnManager(cfg, map[string]chan rpcclient.ClientConnector{
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
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaStatS):          internalStatSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRoutes):         internalRouteSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaThresholds):     internalThresholdSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaServiceManager): internalServeManagerChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaConfig):         internalConfigChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaCore):           internalCoreSv1Chan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRALs):           internalRALsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaEEs):            internalEEsChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaRateS):          internalRateSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaActions):        internalActionSChan,
		utils.ConcatenatedKey(utils.MetaInternal, utils.MetaDispatchers):    internalDispatcherSChan,
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
		utils.DispatcherH:     new(sync.WaitGroup),
		utils.DispatcherS:     new(sync.WaitGroup),
		utils.DNSAgent:        new(sync.WaitGroup),
		utils.EventExporterS:  new(sync.WaitGroup),
		utils.ERs:             new(sync.WaitGroup),
		utils.FreeSWITCHAgent: new(sync.WaitGroup),
		utils.GlobalVarS:      new(sync.WaitGroup),
		utils.HTTPAgent:       new(sync.WaitGroup),
		utils.KamailioAgent:   new(sync.WaitGroup),
		utils.LoaderS:         new(sync.WaitGroup),
		utils.RadiusAgent:     new(sync.WaitGroup),
		utils.RALService:      new(sync.WaitGroup),
		utils.RateS:           new(sync.WaitGroup),
		utils.ResourceS:       new(sync.WaitGroup),
		utils.ResponderS:      new(sync.WaitGroup),
		utils.RouteS:          new(sync.WaitGroup),
		utils.SchedulerS:      new(sync.WaitGroup),
		utils.SessionS:        new(sync.WaitGroup),
		utils.SIPAgent:        new(sync.WaitGroup),
		utils.StatS:           new(sync.WaitGroup),
		utils.StorDB:          new(sync.WaitGroup),
		utils.ThresholdS:      new(sync.WaitGroup),
		utils.ActionS:         new(sync.WaitGroup),
	}
	gvService := services.NewGlobalVarS(cfg, srvDep)
	shdWg.Add(1)
	if err = gvService.Start(); err != nil {
		return
	}
	dmService := services.NewDataDBService(cfg, connManager, srvDep)
	if dmService.ShouldRun() { // Some services can run without db, ie:  ERs
		shdWg.Add(1)
		if err = dmService.Start(); err != nil {
			return
		}
	}

	storDBService := services.NewStorDBService(cfg, srvDep)
	if storDBService.ShouldRun() { // Some services can run without db, ie:  ERs
		shdWg.Add(1)
		if err = storDBService.Start(); err != nil {
			return
		}
	}

	// Rpc/http server
	server := cores.NewServer(caps)
	if len(cfg.HTTPCfg().DispatchersRegistrarURL) != 0 {
		server.RegisterHttpFunc(cfg.HTTPCfg().DispatchersRegistrarURL, dispatcherh.Registar)
	}
	if cfg.ConfigSCfg().Enabled {
		server.RegisterHttpFunc(cfg.ConfigSCfg().URL, config.HandlerConfigS)
	}
	if *httpPprofPath != "" {
		server.RegisterProfiler(*httpPprofPath)
	}

	// Define internal connections via channels
	filterSChan := make(chan *engine.FilterS, 1)

	// init AnalyzerS
	anz := services.NewAnalyzerService(cfg, server, filterSChan, shdChan, internalAnalyzerSChan, srvDep)
	if anz.ShouldRun() {
		shdWg.Add(1)
		if err := anz.Start(); err != nil {
			fmt.Println(err)
			return
		}
	}

	// init CoreSv1
	coreS := services.NewCoreService(cfg, caps, server, internalCoreSv1Chan, anz, srvDep)
	shdWg.Add(1)
	if err := coreS.Start(); err != nil {
		fmt.Println(err)
		return
	}

	// init CacheS
	cacheS := initCacheS(internalCacheSChan, server, dmService.GetDM(), shdChan, anz, coreS.GetCoreS().CapsStats)
	engine.SetCache(cacheS)

	// init GuardianSv1
	initGuardianSv1(internalGuardianSChan, server, anz)

	// Start ServiceManager
	srvManager := servmanager.NewServiceManager(cfg, shdChan, shdWg)
	attrS := services.NewAttributeService(cfg, dmService, cacheS, filterSChan, server, internalAttributeSChan, anz, srvDep)
	dspS := services.NewDispatcherService(cfg, dmService, cacheS, filterSChan, server, internalDispatcherSChan, connManager, anz, srvDep)
	dspH := services.NewDispatcherHostsService(cfg, server, connManager, anz, srvDep)
	chrS := services.NewChargerService(cfg, dmService, cacheS, filterSChan, server,
		internalChargerSChan, connManager, anz, srvDep)
	tS := services.NewThresholdService(cfg, dmService, cacheS, filterSChan, server, internalThresholdSChan, anz, srvDep)
	stS := services.NewStatService(cfg, dmService, cacheS, filterSChan, server,
		internalStatSChan, connManager, anz, srvDep)
	reS := services.NewResourceService(cfg, dmService, cacheS, filterSChan, server,
		internalResourceSChan, connManager, anz, srvDep)
	routeS := services.NewRouteService(cfg, dmService, cacheS, filterSChan, server,
		internalRouteSChan, connManager, anz, srvDep)

	schS := services.NewSchedulerService(cfg, dmService, cacheS, filterSChan,
		server, internalSchedulerSChan, connManager, anz, srvDep)

	rals := services.NewRalService(cfg, cacheS, server,
		internalRALsChan, internalResponderChan,
		shdChan, connManager, anz, srvDep)

	apiSv1 := services.NewAPIerSv1Service(cfg, dmService, storDBService, filterSChan, server, schS, rals.GetResponder(),
		internalAPIerSv1Chan, connManager, anz, srvDep)

	apiSv2 := services.NewAPIerSv2Service(apiSv1, cfg, server, internalAPIerSv2Chan, anz, srvDep)

	cdrS := services.NewCDRServer(cfg, dmService, storDBService, filterSChan, server, internalCDRServerChan,
		connManager, anz, srvDep)

	smg := services.NewSessionService(cfg, dmService, server, internalSessionSChan, shdChan, connManager, caps, anz, srvDep)

	ldrs := services.NewLoaderService(cfg, dmService, filterSChan, server,
		internalLoaderSChan, connManager, anz, srvDep)

	srvManager.AddServices(gvService, attrS, chrS, tS, stS, reS, routeS, schS, rals,
		apiSv1, apiSv2, cdrS, smg, coreS,
		services.NewEventReaderService(cfg, filterSChan, shdChan, connManager, srvDep),
		services.NewDNSAgent(cfg, filterSChan, shdChan, connManager, srvDep),
		services.NewFreeswitchAgent(cfg, shdChan, connManager, srvDep),
		services.NewKamailioAgent(cfg, shdChan, connManager, srvDep),
		services.NewAsteriskAgent(cfg, shdChan, connManager, srvDep),              // partial reload
		services.NewRadiusAgent(cfg, filterSChan, shdChan, connManager, srvDep),   // partial reload
		services.NewDiameterAgent(cfg, filterSChan, shdChan, connManager, srvDep), // partial reload
		services.NewHTTPAgent(cfg, filterSChan, server, connManager, srvDep),      // no reload
		ldrs, anz, dspS, dspH, dmService, storDBService,
		services.NewEventExporterService(cfg, filterSChan,
			connManager, server, internalEEsChan, anz, srvDep),
		services.NewRateService(cfg, cacheS, filterSChan, dmService,
			server, internalRateSChan, anz, srvDep),
		services.NewSIPAgent(cfg, filterSChan, shdChan, connManager, srvDep),
		services.NewActionService(cfg, dmService, cacheS, filterSChan, server, internalActionSChan, anz, srvDep),
	)
	srvManager.StartServices()
	// Start FilterS
	go startFilterService(filterSChan, cacheS, connManager,
		cfg, dmService.GetDM())

	initServiceManagerV1(internalServeManagerChan, srvManager, server, anz)

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
	engine.IntRPC.AddInternalRPCClient(utils.RouteSv1, internalRouteSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ThresholdSv1, internalThresholdSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ServiceManagerV1, internalServeManagerChan)
	engine.IntRPC.AddInternalRPCClient(utils.ConfigSv1, internalConfigChan)
	engine.IntRPC.AddInternalRPCClient(utils.CoreSv1, internalCoreSv1Chan)
	engine.IntRPC.AddInternalRPCClient(utils.RALsV1, internalRALsChan)
	engine.IntRPC.AddInternalRPCClient(utils.RateSv1, internalRateSChan)
	engine.IntRPC.AddInternalRPCClient(utils.ActionSv1, internalActionSChan)
	engine.IntRPC.AddInternalRPCClient(utils.EeSv1, internalEEsChan)
	engine.IntRPC.AddInternalRPCClient(utils.DispatcherSv1, internalDispatcherSChan)

	initConfigSv1(internalConfigChan, server, anz)

	if *preload != utils.EmptyString {
		runPreload(ldrs, internalLoaderSChan, shdChan)
	}

	// Serve rpc connections
	go startRPC(server, internalResponderChan, internalCDRServerChan,
		internalResourceSChan, internalStatSChan,
		internalAttributeSChan, internalChargerSChan, internalThresholdSChan,
		internalRouteSChan, internalSessionSChan, internalAnalyzerSChan,
		internalDispatcherSChan, internalLoaderSChan, internalRALsChan,
		internalCacheSChan, internalEEsChan, internalRateSChan, internalActionSChan, shdChan)

	<-shdChan.Done()
	shtdDone := make(chan struct{})
	go func() {
		shdWg.Wait()
		close(shtdDone)
	}()
	select {
	case <-shtdDone:
	case <-time.After(10 * time.Second): //cfg.CoreSCfg().ShutdownTimeout):
		utils.Logger.Err(fmt.Sprintf("<%s> Failed to shutdown all subsystems in the given time",
			utils.ServiceManager))
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
