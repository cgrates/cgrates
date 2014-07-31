/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2013 ITsysCOM

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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	//"runtime"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/apier"
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
	OSIPS    = "opensips"
)

var (
	cfgPath         = flag.String("config", "/etc/cgrates/cgrates.cfg", "Configuration file location.")
	version         = flag.Bool("version", false, "Prints the application version.")
	raterEnabled    = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled    = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon .overwriting config")
	cdrsEnabled     = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	cdrcEnabled     = flag.Bool("cdrc", false, "Enforce starting of the cdrc service overwriting config")
	mediatorEnabled = flag.Bool("mediator", false, "Enforce starting of the mediator service overwriting config")
	pidFile         = flag.String("pid", "", "Write pid file")
	bal             = balancer2go.NewBalancer()
	exitChan        = make(chan bool)
	server          = &engine.Server{}
	scribeServer    history.Scribe
	cdrServer       *engine.CDRS
	cdrStats        *engine.Stats
	sm              sessionmanager.SessionManager
	medi            *engine.Mediator
	cfg             *config.CGRConfig
	err             error
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

func startMediator(responder *engine.Responder, loggerDb engine.LogStorage, cdrDb engine.CdrStorage, cacheChan, chanDone chan struct{}) {
	var connector engine.Connector
	if cfg.MediatorRater == utils.INTERNAL {
		<-cacheChan // Cache needs to come up before we are ready
		connector = responder
	} else {
		var client *rpcclient.RpcClient
		var err error

		for i := 0; i < cfg.MediatorReconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.MediatorRater, 0, cfg.MediatorReconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<Mediator> Could not connect to engine: %v", err))
			exitChan <- true
			return
		}
		connector = &engine.RPCClientConnector{Client: client}
	}
	var err error
	medi, err = engine.NewMediator(connector, loggerDb, cdrDb, cdrStats, cfg)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Mediator config parsing error: %v", err))
		exitChan <- true
		return
	}
	engine.Logger.Info("Registering Mediator RPC service.")
	server.RpcRegister(&apier.MediatorV1{Medi: medi})

	close(chanDone)
}

// Fires up a cdrc instance
func startCdrc(cdrsChan chan struct{}, cdrsAddress, cdrType, cdrInDir, cdrOutDir, cdrSourceId string, runDelay time.Duration, csvSep string, cdrFields map[string][]*utils.RSRField) {
	if cdrsAddress == utils.INTERNAL {
		<-cdrsChan // Wait for CDRServer to come up before start processing
	}
	cdrc, err := cdrc.NewCdrc(cdrsAddress, cdrType, cdrInDir, cdrOutDir, cdrSourceId, runDelay, csvSep, cdrFields, cdrServer)
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

func startSessionManager(responder *engine.Responder, loggerDb engine.LogStorage, cacheChan chan struct{}) {
	var raterConn engine.Connector
	var client *rpcclient.RpcClient
	if cfg.SMRater == utils.INTERNAL {
		<-cacheChan // Wait for the cache to init before start doing queries
		raterConn = responder
	} else {
		var err error
		for i := 0; i < cfg.SMReconnects; i++ {
			client, err = rpcclient.NewRpcClient("tcp", cfg.SMRater, 0, cfg.SMReconnects, utils.GOB)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to engine: %v", err))
			exitChan <- true
		}
		raterConn = &engine.RPCClientConnector{Client: client}
	}
	switch cfg.SMSwitchType {
	case FS:
		dp, _ := time.ParseDuration(fmt.Sprintf("%vs", cfg.SMDebitInterval))
		sm = sessionmanager.NewFSSessionManager(cfg, loggerDb, raterConn, dp)
	case OSIPS:
		var cdrsConn engine.Connector
		if cfg.OsipCDRS == cfg.SMRater {
			cdrsConn = raterConn
		} else if cfg.OsipCDRS == utils.INTERNAL {
			<-cacheChan // Wait for the cache to init before start doing queries
			cdrsConn = responder
		} else {
			for i := 0; i < cfg.OsipsReconnects; i++ {
				client, err = rpcclient.NewRpcClient("tcp", cfg.OsipCDRS, 0, cfg.SMReconnects, utils.GOB)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(time.Duration(i+1) * time.Second)
			}
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<SM-OpenSIPS> Could not connect to CDRS via RPC: %v", err))
				exitChan <- true
			}
			cdrsConn = &engine.RPCClientConnector{Client: client}
		}
		sm, _ = sessionmanager.NewOSipsSessionManager(cfg, raterConn, cdrsConn)
	default:
		engine.Logger.Err(fmt.Sprintf("<SessionManager> Unsupported session manger type: %s!", cfg.SMSwitchType))
	}
	if err = sm.Connect(); err != nil {
		engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", err))
	}
	exitChan <- true
}

func startCDRS(responder *engine.Responder, cdrDb engine.CdrStorage, mediChan, doneChan chan struct{}) {
	if cfg.CDRSMediator == utils.INTERNAL {
		<-mediChan // Deadlock if mediator not started
		if medi == nil {
			engine.Logger.Crit("<CDRS> Could not connect to mediator, exiting.")
			exitChan <- true
			return
		}
	}
	cdrServer = engine.NewCdrS(cdrDb, medi, cdrStats, cfg)
	cdrServer.RegisterHanlersToServer(server)
	engine.Logger.Info("Registering CDRS RPC service.")
	server.RpcRegister(&apier.CDRSV1{CdrSrv: cdrServer})
	responder.CdrSrv = cdrServer // Make the cdrserver available for internal communication
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
		for i := 0; i < 3; i++ { //ToDo: Make it globally configurable
			//engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Trying to connect, iteration: %d, time %s", i, time.Now()))
			if scribeServer, err = history.NewProxyScribe(cfg.HistoryServer); err == nil {
				break //Connected so no need to reiterate
			} else if i == 2 && err != nil {
				engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			time.Sleep(time.Duration(i) * time.Second)
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

func checkConfigSanity() error {
	if cfg.SMEnabled && cfg.RaterEnabled && cfg.RaterBalancer != "" {
		engine.Logger.Crit("The session manager must not be enabled on a worker engine (change [engine]/balancer to disabled)!")
		return errors.New("SessionManager on Worker")
	}
	if cfg.BalancerEnabled && cfg.RaterEnabled && cfg.RaterBalancer != "" {
		engine.Logger.Crit("The balancer is enabled so it cannot connect to another balancer (change rater/balancer to disabled)!")
		return errors.New("Improperly configured balancer")
	}
	if cfg.CDRSEnabled && cfg.CDRSMediator == utils.INTERNAL && !cfg.MediatorEnabled {
		engine.Logger.Crit("CDRS cannot connect to mediator, Mediator not enabled in configuration!")
		return errors.New("Internal Mediator required by CDRS")
	}
	if cfg.HistoryServerEnabled && cfg.HistoryServer == utils.INTERNAL && !cfg.HistoryServerEnabled {
		engine.Logger.Crit("The history agent is enabled and internal and history server is disabled!")
		return errors.New("Improperly configured history service")
	}
	return nil
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
	// runtime.GOMAXPROCS(runtime.NumCPU())   // For now it slows down computing due to CPU management, to be reviewed in future Go releases

	cfg, err = config.NewCGRConfigFromFile(cfgPath)
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
	if *cdrcEnabled {
		cfg.CdrcEnabled = *cdrcEnabled
	}
	if *mediatorEnabled {
		cfg.MediatorEnabled = *mediatorEnabled
	}

	// some consitency checks
	errCfg := checkConfigSanity()
	if errCfg != nil {
		engine.Logger.Crit(errCfg.Error())
		return
	}
	var ratingDb engine.RatingStorage
	var accountDb engine.AccountingStorage
	var logDb engine.LogStorage
	var loadDb engine.LoadStorage
	var cdrDb engine.CdrStorage
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

	if cfg.StorDBType == SAME {
		logDb = ratingDb.(engine.LogStorage)
	} else {
		logDb, err = engine.ConfigureLogStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort,
			cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding)
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

	engine.SetRoundingDecimals(cfg.RoundingDecimals)
	if cfg.SMDebitInterval > 0 {
		if dp, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.SMDebitInterval)); err == nil {
			engine.SetDebitPeriod(dp)
		}
	}

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

	responder := &engine.Responder{ExitChan: exitChan}
	apierRpc := &apier.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, LogDb: logDb, Config: cfg, Responder: responder}

	if cfg.RaterEnabled && !cfg.BalancerEnabled && cfg.RaterBalancer != utils.INTERNAL {
		engine.Logger.Info("Registering Rater service")
		server.RpcRegister(responder)
		server.RpcRegister(apierRpc)
	}

	if cfg.BalancerEnabled {
		engine.Logger.Info("Registering Balancer service.")
		go stopBalancerSignalHandler()
		stopHandled = true
		responder.Bal = bal
		server.RpcRegister(responder)
		server.RpcRegister(apierRpc)
		if cfg.RaterEnabled {
			engine.Logger.Info("<Balancer> Registering internal rater")
			bal.AddClient("local", new(engine.ResponderWorker))
		}
	}

	if !stopHandled {
		go generalSignalHandler()
	}

	if cfg.SchedulerEnabled {
		engine.Logger.Info("Starting CGRateS Scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, accountDb)
			apierRpc.Sched = sched
			sched.LoadActionTimings(accountDb)
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

	var medChan chan struct{}
	if cfg.MediatorEnabled {
		engine.Logger.Info("Starting CGRateS Mediator service.")
		medChan = make(chan struct{})
		go startMediator(responder, logDb, cdrDb, cacheChan, medChan)
	}

	if cfg.CDRStatsEnabled {
		cdrStats = &engine.Stats{}
		server.RpcRegister(cdrStats)
		server.RpcRegister(apier.CDRStatsV1{cdrStats}) // Public APIs
	}

	var cdrsChan chan struct{}
	if cfg.CDRSEnabled {
		engine.Logger.Info("Starting CGRateS CDRS service.")
		cdrsChan = make(chan struct{})
		httpWait = append(httpWait, cdrsChan)
		go startCDRS(responder, cdrDb, medChan, cdrsChan)
	}

	if cfg.SMEnabled {
		engine.Logger.Info("Starting CGRateS SessionManager service.")
		go startSessionManager(responder, logDb, cacheChan)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}
	var cdrcEnabled bool
	if cfg.CdrcEnabled { // Start default cdrc configured in csv here
		cdrcEnabled = true
		go startCdrc(cdrsChan, cfg.CdrcCdrs, cfg.CdrcCdrType, cfg.CdrcCdrInDir, cfg.CdrcCdrOutDir, cfg.CdrcSourceId, cfg.CdrcRunDelay, cfg.CdrcCsvSep, cfg.CdrcCdrFields)
	}
	if cfg.XmlCfgDocument != nil {
		for _, xmlCdrc := range cfg.XmlCfgDocument.GetCdrcCfgs("") {
			if !xmlCdrc.Enabled {
				continue
			}
			cdrcEnabled = true
			go startCdrc(cdrsChan, xmlCdrc.CdrsAddress, xmlCdrc.CdrType, xmlCdrc.CdrInDir, xmlCdrc.CdrOutDir,
				xmlCdrc.CdrSourceId, time.Duration(xmlCdrc.RunDelay), xmlCdrc.CsvSeparator, xmlCdrc.CdrRSRFields())
		}
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
