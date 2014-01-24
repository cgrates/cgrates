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
	"net/rpc"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/cgrates/cgrates/apier/v1"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/cdrc"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

const (
	INTERNAL = "internal"
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
	MYSQL    = "mysql"
	MONGO    = "mongo"
	REDIS    = "redis"
	SAME     = "same"
	FS       = "freeswitch"
)

var (
	cfgPath         = flag.String("config", "/etc/cgrates/cgrates.cfg", "Configuration file location.")
	version         = flag.Bool("version", false, "Prints the application version.")
	raterEnabled    = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled    = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon overwriting config")
	cdrsEnabled     = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	cdrcEnabled     = flag.Bool("cdrc", false, "Enforce starting of the cdrc service overwriting config")
	mediatorEnabled = flag.Bool("mediator", false, "Enforce starting of the mediator service overwriting config")
	pidFile         = flag.String("pid", "", "Write pid file")
	bal             = balancer2go.NewBalancer()
	exitChan        = make(chan bool)
	server          = &engine.Server{}
	sm              sessionmanager.SessionManager
	medi            *mediator.Mediator
	cfg             *config.CGRConfig
	err             error
)

func startMediator(responder *engine.Responder, loggerDb engine.LogStorage, cdrDb engine.CdrStorage) {
	var connector engine.Connector
	if cfg.MediatorRater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error

		for i := 0; i < cfg.MediatorRaterReconnects; i++ {
			client, err = rpc.Dial("tcp", cfg.MediatorRater)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<Mediator> Could not connect to engine: %v", err))
			exitChan <- true
		}
		connector = &engine.RPCClientConnector{Client: client}
	}
	var err error
	medi, err = mediator.NewMediator(connector, loggerDb, cdrDb, cfg)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Mediator config parsing error: %v", err))
		exitChan <- true
	}
}

func startCdrc() {
	cdrc, err := cdrc.NewCdrc(cfg)
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

func startSessionManager(responder *engine.Responder, loggerDb engine.LogStorage) {
	var connector engine.Connector
	if cfg.SMRater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error

		for i := 0; i < cfg.SMRaterReconnects; i++ {
			client, err = rpc.Dial("tcp", cfg.SMRater)
			if err == nil { //Connected so no need to reiterate
				break
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("<SessionManager> Could not connect to engine: %v", err))
			exitChan <- true
		}
		connector = &engine.RPCClientConnector{Client: client}
	}
	switch cfg.SMSwitchType {
	case FS:
		dp, _ := time.ParseDuration(fmt.Sprintf("%vs", cfg.SMDebitInterval))
		sm = sessionmanager.NewFSSessionManager(loggerDb, connector, dp)
		errConn := sm.Connect(cfg)
		if errConn != nil {
			engine.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", errConn))
		}
	default:
		engine.Logger.Err(fmt.Sprintf("<SessionManager> Unsupported session manger type: %s!", cfg.SMSwitchType))
		exitChan <- true
	}
	exitChan <- true
}

func startCDRS(responder *engine.Responder, cdrDb engine.CdrStorage) {
	if cfg.CDRSMediator == INTERNAL {
		for i := 0; i < 3; i++ { // ToDo: If the right approach, make the reconnects configurable
			time.Sleep(time.Duration(i+1) * time.Second)
			if medi != nil { // Got our mediator, no need to wait any longer
				break
			}
		}
		if medi == nil {
			engine.Logger.Crit("<CDRS> Could not connect to mediator, exiting.")
			exitChan <- true
		}
	}
	cs := cdrs.New(cdrDb, medi, cfg)
	cs.RegisterHanlersToServer(server)
}

func startHistoryAgent(scribeServer history.Scribe) {
	if cfg.HistoryServer != INTERNAL { // Connect in iteration since there are chances of concurrency here
		engine.Logger.Info("Starting History Agent.")
		for i := 0; i < 3; i++ { //ToDo: Make it globally configurable
			//engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Trying to connect, iteration: %d, time %s", i, time.Now()))
			if scribeServer, err = history.NewProxyScribe(cfg.HistoryServer); err == nil {
				break //Connected so no need to reiterate
			} else if i == 2 && err != nil {
				engine.Logger.Crit(fmt.Sprintf("<HistoryAgent> Could not connect to the server, error: %s", err.Error()))
				exitChan <- true
				return
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	engine.SetHistoryScribe(scribeServer)
	return
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
	if cfg.CDRSEnabled && cfg.CDRSMediator == INTERNAL && !cfg.MediatorEnabled {
		engine.Logger.Crit("CDRS cannot connect to mediator, Mediator not enabled in configuration!")
		return errors.New("Internal Mediator required by CDRS")
	}
	if cfg.HistoryServerEnabled && cfg.HistoryServer == INTERNAL && !cfg.HistoryServerEnabled {
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg, err = config.NewCGRConfig(cfgPath)
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
	ratingDb, err = engine.ConfigureRatingStorage(cfg.RatingDBType, cfg.RatingDBHost, cfg.RatingDBPort, cfg.RatingDBName, cfg.RatingDBUser, cfg.RatingDBPass, cfg.DBDataEncoding)
	if err != nil { // Cannot configure getter database, show stopper
		engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	defer ratingDb.Close()
	engine.SetRatingStorage(ratingDb)
	accountDb, err = engine.ConfigureAccountingStorage(cfg.AccountDBType, cfg.AccountDBHost, cfg.AccountDBPort, cfg.AccountDBName, cfg.AccountDBUser, cfg.AccountDBPass, cfg.DBDataEncoding)
	if err != nil { // Cannot configure getter database, show stopper
		engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	defer accountDb.Close()
	engine.SetAccountingStorage(accountDb)

	if cfg.StorDBType == SAME {
		logDb = ratingDb.(engine.LogStorage)
	} else {
		logDb, err = engine.ConfigureLogStorage(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass, cfg.DBDataEncoding)
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

	engine.SetRoundingMethodAndDecimals(cfg.RoundingMethod, cfg.RoundingDecimals)
	if cfg.SMDebitInterval > 0 {
		if dp, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.SMDebitInterval)); err == nil {
			engine.SetDebitPeriod(dp)
		}
	}

	stopHandled := false

	if cfg.RaterEnabled { // Cache rating if rater enabled
		if err := ratingDb.CacheRating(nil, nil, nil); err != nil {
			engine.Logger.Crit(fmt.Sprintf("Cache rating error: %s", err.Error()))
			return
		}
		if err := accountDb.CacheAccounting(nil); err != nil {
			engine.Logger.Crit(fmt.Sprintf("Cache accounting error: %s", err.Error()))
			return
		}
	}

	// Async starts here
	if cfg.RaterEnabled && cfg.RaterBalancer != "" && !cfg.BalancerEnabled {
		go registerToBalancer()
		go stopRaterSignalHandler()
		stopHandled = true
	}

	responder := &engine.Responder{ExitChan: exitChan}
	apier := &apier.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, Config: cfg}

	if cfg.RaterEnabled && !cfg.BalancerEnabled && cfg.RaterBalancer != INTERNAL {
		engine.Logger.Info("Registering CGRateS Rater service")
		server.RpcRegister(responder)
		server.RpcRegister(apier)
	}

	if cfg.BalancerEnabled {
		engine.Logger.Info("Registering CGRateS Balancer service")
		go stopBalancerSignalHandler()
		stopHandled = true
		responder.Bal = bal
		server.RpcRegister(responder)
		server.RpcRegister(apier)
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
			apier.Sched = sched
			sched.LoadActionTimings(accountDb)
			sched.Loop()
		}()
	}

	var scribeServer history.Scribe

	if cfg.HistoryServerEnabled {
		engine.Logger.Info("Registering CGRates History service")
		if scribeServer, err = history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval); err != nil {
			engine.Logger.Crit(fmt.Sprintf("<HistoryServer> Could not start, error: %s", err.Error()))
			exitChan <- true
			return
		}
		server.RpcRegisterName("Scribe", scribeServer)
	}
	go startHistoryAgent(scribeServer)

	go server.ServeGOB(cfg.RPCGOBListen)
	go server.ServeJSON(cfg.RPCJSONListen)

	go startHistoryAgent(scribeServer)

	if cfg.CDRSEnabled {
		engine.Logger.Info("Registering CGRateS CDR service")
		go startCDRS(responder, cdrDb)
	}

	go server.ServeHTTP(cfg.HTTPListen)

	if cfg.MediatorEnabled {
		engine.Logger.Info("Starting CGRateS Mediator service")
		go startMediator(responder, logDb, cdrDb)
	}

	if cfg.SMEnabled {
		engine.Logger.Info("Starting CGRateS SessionManager service")
		go startSessionManager(responder, logDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if cfg.CdrcEnabled {
		engine.Logger.Info("Starting CGRateS CDR client")
		go startCdrc()
	}

	<-exitChan
	if *pidFile != "" {
		if err := os.Remove(*pidFile); err != nil {
			engine.Logger.Warning("Could not remove pid file: " + err.Error())
		}
	}
	engine.Logger.Info("Stopped all components. CGRateS shutdown!")
}
