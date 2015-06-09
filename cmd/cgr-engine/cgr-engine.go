/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/apier/v2"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

const (
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
	KAMAILIO = "kamailio"
	OSIPS    = "opensips"
)

var (
	cfgDir       = flag.String("config_dir", utils.CONFIG_DIR, "Configuration directory path.")
	version      = flag.Bool("version", false, "Prints the application version.")
	raterEnabled = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon .overwriting config")
	cdrsEnabled  = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	pidFile      = flag.String("pid", "", "Write pid file")
	cpuprofile   = flag.String("cpuprofile", "", "write cpu profile to file")
	singlecpu    = flag.Bool("singlecpu", false, "Run on single CPU core")
	bal          = balancer2go.NewBalancer()
	exitChan     = make(chan bool)
	server       = &engine.Server{}
	scribeServer history.Scribe
	cdrServer    *engine.CdrServer
	cdrStats     *engine.Stats
	sm           sessionmanager.SessionManager
	cfg          *config.CGRConfig
	err          error
)

func cacheData(ratingDb engine.RatingStorage, accountDb engine.AccountingStorage, doneChan chan struct{}) {
	if err := ratingDb.CacheRating(nil, nil, nil, nil, nil); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cache rating error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := accountDb.CacheAccounting(nil, nil, nil, nil); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cache accounting error: %s", err.Error()))
		exitChan <- true
		return
	}
	close(doneChan)
}

// Fires up a cdrc instance
func startCdrc(cdrsChan chan struct{}, cdrcCfgs map[string]*config.CdrcConfig, httpSkipTlsCheck bool, closeChan chan struct{}) {
	var cdrcCfg *config.CdrcConfig
	for _, cdrcCfg = range cdrcCfgs { // Take the first config out, does not matter which one
		break
	}
	if cdrcCfg.Cdrs == utils.INTERNAL {
		<-cdrsChan // Wait for CDRServer to come up before start processing
	}
	cdrc, err := cdrc.NewCdrc(cdrcCfgs, httpSkipTlsCheck, cdrServer, closeChan)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cdrc config parsing error: %s", err.Error()))
		exitChan <- true
		return
	}
	if err := cdrc.Run(); err != nil {
		engine.Logger.Crit(fmt.Sprintf("Cdrc run error: %s", err.Error()))
	}
	exitChan <- true // If run stopped, something is bad, stop the application
}

func startSmFreeSWITCH(responder *engine.Responder, cdrDb engine.CdrStorage, cacheChan chan struct{}) {
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	delay := utils.Fib()
	if cfg.SmFsConfig.Rater == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		raterConn = responder
	} else {
		var err error
		for i := 0; i < cfg.SmFsConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmFsConfig.Rater, cfg.SmFsConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to rater via RPC: %v", err))
			exitChan <- true
			return
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmFsConfig.Cdrs == cfg.SmFsConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmFsConfig.Cdrs == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		cdrsConn = responder
	} else if len(cfg.SmFsConfig.Cdrs) != 0 {
		delay = utils.Fib()
		for i := 0; i < cfg.SmFsConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmFsConfig.Cdrs, cfg.SmFsConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-FreeSWITCH> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm := sessionmanager.NewFSSessionManager(cfg.SmFsConfig, raterConn, cdrsConn)
	if err = sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", err))
	}
	exitChan <- true
}

func startSmKamailio(responder *engine.Responder, cdrDb engine.CdrStorage, cacheChan chan struct{}) {
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SmKamConfig.Rater == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		raterConn = responder
	} else {
		var err error
		delay := utils.Fib()
		for i := 0; i < cfg.SmKamConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmKamConfig.Rater, cfg.SmKamConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to rater: %v", err))
			exitChan <- true
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmKamConfig.Cdrs == cfg.SmKamConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmKamConfig.Cdrs == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		cdrsConn = responder
	} else if len(cfg.SmKamConfig.Cdrs) != 0 {
		delay := utils.Fib()
		for i := 0; i < cfg.SmKamConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmKamConfig.Cdrs, cfg.SmKamConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-Kamailio> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm, _ := sessionmanager.NewKamailioSessionManager(cfg.SmKamConfig, raterConn, cdrsConn)
	if err = sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", err))
	}
	exitChan <- true
}

func startSmOpenSIPS(responder *engine.Responder, cdrDb engine.CdrStorage, cacheChan chan struct{}) {
	var raterConn, cdrsConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SmOsipsConfig.Rater == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		raterConn = responder
	} else {
		var err error
		delay := utils.Fib()
		for i := 0; i < cfg.SmOsipsConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmOsipsConfig.Rater, cfg.SmOsipsConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to rater: %v", err))
			exitChan <- true
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	if cfg.SmOsipsConfig.Cdrs == cfg.SmOsipsConfig.Rater {
		cdrsConn = raterConn
	} else if cfg.SmOsipsConfig.Cdrs == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		cdrsConn = responder
	} else if len(cfg.SmOsipsConfig.Cdrs) != 0 {
		delay := utils.Fib()
		for i := 0; i < cfg.SmOsipsConfig.Reconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SmOsipsConfig.Cdrs, cfg.SmOsipsConfig.Reconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SM-OpenSIPS> Could not connect to CDRS via RPC: %v", err))
			exitChan <- true
			return
		}
		cdrsConn = &engine.RPCClientConnector{Client: client}
	}
	sm, _ := sessionmanager.NewOSipsSessionManager(cfg.SmOsipsConfig, raterConn, cdrsConn)
	if err := sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SM-OpenSIPS> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(logDb engine.LogStorage, cdrDb engine.CdrStorage, responder *engine.Responder, responderReady, doneChan chan struct{}) {
	var err error
	var client *rpcclient.RpcClient
	// Rater connection init
	var raterConn engine.Connector
	if cfg.CDRSRater == utils.INTERNAL {
		<-responderReady // Wait for the cache to init before start doing queries
		raterConn = responder
	} else if len(cfg.CDRSRater) != 0 {
		delay := utils.Fib()
		for i := 0; i < cfg.CDRSReconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSRater, cfg.CDRSReconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(delay())
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to rater: %s", err.Error()))
			exitChan <- true
			return
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	// Stats connection init
	var statsConn engine.StatsInterface
	if cfg.CDRSStats == utils.INTERNAL {
		statsConn = cdrStats
	} else if len(cfg.CDRSStats) != 0 {
		if cfg.CDRSRater == cfg.CDRSStats {
			statsConn = &engine.ProxyStats{Client: client}
		} else {
			delay := utils.Fib()
			for i := 0; i < cfg.CDRSReconnects; i++ {
				client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSStats, cfg.CDRSReconnects, utils.GOB)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(delay())
			}
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to stats server: %s", err.Error()))
				exitChan <- true
				return
			}
			statsConn = &engine.ProxyStats{Client: client}
		}
	}

	cdrServer, _ = engine.NewCdrServer(cfg, cdrDb, raterConn, statsConn)
	engine.Logger.Info("Registering CDRS HTTP Handlers.")
	cdrServer.RegisterHanlersToServer(server)
	engine.Logger.Info("Registering CDRS RPC service.")
	cdrSrv := v1.CdrsV1{CdrSrv: cdrServer}
	server.RpcRegister(&cdrSrv)
	server.RpcRegister(&v2.CdrsV2{CdrsV1: cdrSrv})
	// Make the cdr servers available for internal communication
	responder.CdrSrv = cdrServer
	close(doneChan)
}

func startHistoryServer(chanDone chan struct{}) {
	if scribeServer, err = history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval); err != nil {
		engine.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not start, error: %s", err.Error()))
		exitChan <- true
		return
	}
	server.RpcRegisterName("Scribe", scribeServer)
	close(chanDone)
}

// chanStartServer will report when server is up, useful for internal requests
func startHistoryAgent(chanServerStarted chan struct{}) {
	if cfg.HistoryServer == utils.INTERNAL { // For internal requests, wait for server to come online before connecting
		engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Connecting internally to HistoryServer"))
		select {
		case <-time.After(1 * time.Minute):
			engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Timeout waiting for server to start."))
			exitChan <- true
			return
		case <-chanServerStarted:
		}
		//<-chanServerStarted // If server is not enabled, will have deadlock here
	} else { // Connect in iteration since there are chances of concurrency here
		delay := utils.Fib()
		for i := 0; i < 3; i++ { //ToDo: Make it globally configurable
			//engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Trying to connect, iteration: %d, time %s", i, time.Now()))
			if scribeServer, err = history.NewProxyScribe(cfg.HistoryServer); err == nil {
				break //Connected so no need to reiterate
			} else if i == 2 && err != nil {
				engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			time.Sleep(delay())
		}
	}
	engine.SetHistoryScribe(scribeServer) // scribeServer comes from global variable
	return
}

// Starts the rpc server, waiting for the necessary components to finish their tasks
func serveRpc(rpcWaitChans []chan struct{}) {
	for _, chn := range rpcWaitChans {
		<-chn
	}
	// Each of the serve blocks so need to start in their own goroutine
	go server.ServeJSON(cfg.RPCJSONListen)
	go server.ServeGOB(cfg.RPCGOBListen)
}

// Starts the http server, waiting for the necessary components to finish their tasks
func serveHttp(httpWaitChans []chan struct{}) {
	for _, chn := range httpWaitChans {
		<-chn
	}
	server.ServeHTTP(cfg.HTTPListen)
}

func writePid() {
	engine.Logger.Info(*pidFile)
	f, err := os.Create(*pidFile)
	if err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
	f.WriteString(strconv.Itoa(os.Getpid()))
	if err := f.Close(); err != nil {
		log.Fatal("Could not write pid file: ", err)
	}
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	if *pidFile != "" {
		writePid()
	}
	if !*singlecpu {
		runtime.GOMAXPROCS(runtime.NumCPU()) // For now it slows down computing due to CPU management, to be reviewed in future Go releases
	}
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	cfg, err = config.NewCGRConfigFromFolder(*cfgDir)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
		return
	}
	config.SetCgrConfig(cfg) // Share the config object
	if *raterEnabled {
		cfg.RaterEnabled = *raterEnabled
	}
	if *schedEnabled {
		cfg.SchedulerEnabled = *schedEnabled
	}
	if *cdrsEnabled {
		cfg.CDRSEnabled = *cdrsEnabled
	}
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var logDb engine.LogStorage
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
	if cfg.RaterEnabled || cfg.SchedulerEnabled { // Only connect to dataDb if required
		ratingDb, err = engine.ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort,
			cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
		if err != nil { // Cannot configure getter database, show stopper
			engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer ratingDb.Close()
		engine.SetRatingStorage(ratingDb)
		accountDb, err = engine.ConfigureAccountingStorage(cfg.AccountDBType, cfg.AccountDBHost, cfg.AccountDBPort,
			cfg.AccountDBName, cfg.AccountDBUser, cfg.AccountDBPass, cfg.DBDataEncoding)
		if err != nil { // Cannot configure getter database, show stopper
			engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
			return
		}
		defer accountDb.Close()
		engine.SetAccountingStorage(accountDb)
	}
	if cfg.RaterEnabled || cfg.CDRSEnabled || cfg.SchedulerEnabled { // Only connect to storDb if necessary
		if cfg.StorDBType == SAME {
			logDb = ratingDb.(engine.LogStorage)
		} else {
			logDb, err = engine.ConfigureLogStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
				cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding, cfg.StorDBMaxOpenConns, cfg.StorDBMaxIdleConns)
			if err != nil { // Cannot configure logger database, show stopper
				engine.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
				return
			}
		}
		defer logDb.Close()
		engine.SetStorageLogger(logDb)
		// loadDb,cdrDb and logDb are all mapped on the same stordb storage
		loadDb = logDb.(engine.LoadStorage)
		cdrDb = logDb.(engine.CdrStorage)
		engine.SetCdrStorage(cdrDb)
	}

	engine.SetRoundingDecimals(cfg.RoundingDecimals)
	stopHandled := false

	// Async starts here

	rpcWait := make([]chan struct{}, 0)  // Rpc server will start as soon as this list is consumed
	httpWait := make([]chan struct{}, 0) // Http server will start as soon as this list is consumed

	var cacheChan chan struct{}
	if cfg.RaterEnabled { // Cache rating if rater enabled
		cacheChan = make(chan struct{})
		rpcWait = append(rpcWait, cacheChan)
		go cacheData(ratingDb, accountDb, cacheChan)
	}

	if cfg.RaterEnabled && cfg.RaterBalancer != "" && !cfg.BalancerEnabled {
		go registerToBalancer()
		go stopRaterSignalHandler()
		stopHandled = true
	}

	if cfg.CDRStatsEnabled { // Init it here so we make it availabe to the Apier
		cdrStats = engine.NewStats(ratingDb)
		if cfg.CDRStatConfig != nil && len(cfg.CDRStatConfig.Metrics) != 0 {
			cdrStats.AddQueue(engine.NewCdrStatsFromCdrStatsCfg(cfg.CDRStatConfig), nil)
		}
		server.RpcRegister(cdrStats)
		server.RpcRegister(&v1.CDRStatsV1{CdrStats: cdrStats}) // Public APIs
	}

	responder := &engine.Responder{ExitChan: exitChan}
	apierRpcV1 := &v1.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, LogDb: logDb, Config: cfg, Responder: responder, CdrStatsSrv: cdrStats}
	apierRpcV2 := &v2.ApierV2{ApierV1: v1.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, LogDb: logDb, Config: cfg, Responder: responder, CdrStatsSrv: cdrStats}}

	if cfg.RaterEnabled && !cfg.BalancerEnabled && cfg.RaterBalancer != utils.INTERNAL {
		engine.Logger.Info("Registering Rater service")
		server.RpcRegister(responder)
		server.RpcRegister(apierRpcV1)
		server.RpcRegister(apierRpcV2)
	}

	if cfg.BalancerEnabled {
		engine.Logger.Info("Registering Balancer service.")
		go stopBalancerSignalHandler()
		stopHandled = true
		responder.Bal = bal
		server.RpcRegister(responder)
		server.RpcRegister(apierRpcV1)
		server.RpcRegister(apierRpcV2)
		if cfg.RaterEnabled {
			engine.Logger.Info("<Balancer> Registering internal rater")
			bal.AddClient("local", new(engine.ResponderWorker))
		}
	}

	if cfg.RaterCdrStats != "" {
		var statsConn engine.StatsInterface
		if cfg.CDRSStats == utils.INTERNAL {
			statsConn = cdrStats
		} else {
			delay := utils.Fib()
			var client *rpcclient.RpcClient
			for i := 0; i < cfg.CDRSReconnects; i++ {
				client, err = rpcclient.NewRpcClient("tcp", cfg.CDRSStats, cfg.CDRSReconnects, utils.GOB)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(delay())
			}
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<CDRS> Could not connect to stats server: %s", err.Error()))
				exitChan <- true
				return
			}
			statsConn = &engine.ProxyStats{Client: client}
		}
		responder.Stats = statsConn
	}

	if !stopHandled {
		go generalSignalHandler()
	}

	if cfg.SchedulerEnabled {
		engine.Logger.Info("Starting CGRateS Scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, accountDb)
			apierRpcV1.Sched = sched
			apierRpcV2.Sched = sched
			sched.LoadActionPlans(accountDb)
			sched.Loop()
		}()
	}

	var histServChan chan struct{} // Will be initialized only if the server starts
	if cfg.HistoryServerEnabled {
		histServChan = make(chan struct{})
		rpcWait = append(rpcWait, histServChan)
		go startHistoryServer(histServChan)
	}

	if cfg.HistoryAgentEnabled {
		engine.Logger.Info("Starting CGRateS History Agent.")
		go startHistoryAgent(histServChan)
	}

	var cdrsChan chan struct{}
	if cfg.CDRSEnabled {
		engine.Logger.Info("Starting CGRateS CDRS service.")
		cdrsChan = make(chan struct{})
		httpWait = append(httpWait, cdrsChan)
		go startCDRS(logDb, cdrDb, responder, cacheChan, cdrsChan)
	}

	if cfg.SmFsConfig.Enabled {
		engine.Logger.Info("Starting CGRateS SM-FreeSWITCH service.")
		go startSmFreeSWITCH(responder, cdrDb, cacheChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}
	if cfg.SmKamConfig.Enabled {
		engine.Logger.Info("Starting CGRateS SM-Kamailio service.")
		go startSmKamailio(responder, cdrDb, cacheChan)
	}
	if cfg.SmOsipsConfig.Enabled {
		engine.Logger.Info("Starting CGRateS SM-OpenSIPS service.")
		go startSmOpenSIPS(responder, cdrDb, cacheChan)
	}
	var cdrcEnabled bool
	for _, cdrcCfgs := range cfg.CdrcProfiles {
		var cdrcCfg *config.CdrcConfig
		for _, cdrcCfg = range cdrcCfgs { // Take a random config out since they should be the same
			break
		}
		if cdrcCfg.Enabled == false {
			continue // Ignore not enabled
		} else if !cdrcEnabled {
			cdrcEnabled = true // Mark that at least one cdrc service is active
		}
		go startCdrc(cdrsChan, cdrcCfgs, cfg.HttpSkipTlsVerify, cfg.ConfigReloads[utils.CDRC])
	}
	if cdrcEnabled {
		engine.Logger.Info("Starting CGRateS CDR client.")
	}

	// Start the servers
	go serveRpc(rpcWait)
	go serveHttp(httpWait)
	<-exitChan

	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			engine.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	engine.Logger.Info("Stopped all components. CGRateS shutdown!")
}
