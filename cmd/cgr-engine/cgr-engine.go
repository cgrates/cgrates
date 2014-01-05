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
	"io"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
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
	DISABLED = "disabled"
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
	cfgPath      = flag.String("config", "/etc/cgrates/cgrates.cfg", "Configuration file location.")
	version      = flag.Bool("version", false, "Prints the application version.")
	raterEnabled = flag.Bool("rater", false, "Enforce starting of the rater daemon overwriting config")
	schedEnabled = flag.Bool("scheduler", false, "Enforce starting of the scheduler daemon overwriting config")
	cdrsEnabled  = flag.Bool("cdrs", false, "Enforce starting of the cdrs daemon overwriting config")
	cdrcEnabled  = flag.Bool("cdrc", false, "Enforce starting of the cdrc service overwriting config")
	pidFile      = flag.String("pid", "", "Write pid file")
	bal          = balancer2go.NewBalancer()
	exitChan     = make(chan bool)
	sm           sessionmanager.SessionManager
	medi         *mediator.Mediator
	cfg          *config.CGRConfig
	err          error
)

func listenToRPCRequests(rpcResponder interface{}, apier *apier.ApierV1, rpcAddress string, rpc_encoding string) {
	l, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("<Rater> Could not listen to %v: %v", rpcAddress, err))
		exitChan <- true
		return
	}
	defer l.Close()

	engine.Logger.Info(fmt.Sprintf("<Rater> Listening for incomming RPC requests on %v", l.Addr()))
	rpc.Register(rpcResponder)
	rpc.Register(apier)
	var serveFunc func(io.ReadWriteCloser)
	if rpc_encoding == JSON {
		serveFunc = jsonrpc.ServeConn
	} else {
		serveFunc = rpc.ServeConn
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			engine.Logger.Err(fmt.Sprintf("<Rater> Accept error: %v", conn))
			continue
		}

		engine.Logger.Info(fmt.Sprintf("<Rater> New incoming connection: %v", conn.RemoteAddr()))
		go serveFunc(conn)
	}
}

func startMediator(responder *engine.Responder, loggerDb engine.LogStorage, cdrDb engine.CdrStorage) {
	var connector engine.Connector
	if cfg.MediatorRater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if cfg.RPCEncoding == JSON {
			for i := 0; i < cfg.MediatorRaterReconnects; i++ {
				client, err = jsonrpc.Dial("tcp", cfg.MediatorRater)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(time.Duration(i/2) * time.Second)
			}
		} else {
			for i := 0; i < cfg.MediatorRaterReconnects; i++ {
				client, err = rpc.Dial("tcp", cfg.MediatorRater)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(time.Duration(i/2) * time.Second)
			}
		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("Could not connect to engine: %v", err))
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
		if cfg.RPCEncoding == JSON {
			// We attempt to reconnect more times
			for i := 0; i < cfg.SMRaterReconnects; i++ {
				client, err = jsonrpc.Dial("tcp", cfg.SMRater)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(time.Duration(i/2) * time.Second)
			}
		} else {
			for i := 0; i < cfg.SMRaterReconnects; i++ {
				client, err = rpc.Dial("tcp", cfg.SMRater)
				if err == nil { //Connected so no need to reiterate
					break
				}
				time.Sleep(time.Duration(i/2) * time.Second)
			}

		}
		if err != nil {
			engine.Logger.Crit(fmt.Sprintf("Could not connect to engine: %v", err))
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
			time.Sleep(time.Duration(i/2) * time.Second)
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
	cs.StartCapturingCDRs()
	exitChan <- true
}

func startHistoryScribe() {
	var scribeServer history.Scribe

	if cfg.HistoryServerEnabled {
		if scribeServer, err = history.NewFileScribe(cfg.HistoryDir, cfg.HistorySaveInterval); err != nil {
			engine.Logger.Crit(err.Error())
			exitChan <- true
			return
		}
	}

	if cfg.HistoryServerEnabled {
		if cfg.HistoryListen != INTERNAL {
			rpc.RegisterName("Scribe", scribeServer)
			var serveFunc func(io.ReadWriteCloser)
			if cfg.RPCEncoding == JSON {
				serveFunc = jsonrpc.ServeConn
			} else {
				serveFunc = rpc.ServeConn
			}
			l, err := net.Listen("tcp", cfg.HistoryListen)
			if err != nil {
				engine.Logger.Crit(fmt.Sprintf("<History> Could not listen to %v: %v", cfg.HistoryListen, err))
				exitChan <- true
				return
			}
			defer l.Close()
			for {
				conn, err := l.Accept()
				if err != nil {
					engine.Logger.Err(fmt.Sprintf("<History> Accept error: %v", conn))
					continue
				}

				engine.Logger.Info(fmt.Sprintf("<History> New incoming connection: %v", conn.RemoteAddr()))
				go serveFunc(conn)
			}
		}
	}

	var scribeAgent history.Scribe

	if cfg.HistoryAgentEnabled {
		if cfg.HistoryServer != INTERNAL { // Connect in iteration since there are chances of concurrency here
			for i := 0; i < 3; i++ { //ToDo: Make it globally configurable
				if scribeAgent, err = history.NewProxyScribe(cfg.HistoryServer, cfg.RPCEncoding); err == nil {
					break //Connected so no need to reiterate
				} else if i == 2 && err != nil {
					engine.Logger.Crit(err.Error())
					exitChan <- true
					return
				}
				time.Sleep(time.Duration(i/2) * time.Second)
			}
		} else {
			scribeAgent = scribeServer
		}
	}
	if scribeAgent != nil {
		engine.SetHistoryScribe(scribeAgent)
	} else {
		engine.SetHistoryScribe(scribeServer) // if it is nil so be it
	}

	return
}

func checkConfigSanity() error {
	if cfg.SMEnabled && cfg.RaterEnabled && cfg.RaterBalancer != DISABLED {
		engine.Logger.Crit("The session manager must not be enabled on a worker engine (change [engine]/balancer to disabled)!")
		return errors.New("SessionManager on Worker")
	}
	if cfg.BalancerEnabled && cfg.RaterEnabled && cfg.RaterBalancer != DISABLED {
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
	if cfg.RaterEnabled {
		if err := ratingDb.CacheRating(nil, nil, nil); err != nil {
			engine.Logger.Crit(fmt.Sprintf("Cache rating error: %v", err))
			return
		}
	}
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
	// Async starts here
	if cfg.RaterEnabled && cfg.RaterBalancer != DISABLED && !cfg.BalancerEnabled {
		go registerToBalancer()
		go stopRaterSignalHandler()
		stopHandled = true
	}
	responder := &engine.Responder{ExitChan: exitChan}
	apier := &apier.ApierV1{StorDb: loadDb, RatingDb: ratingDb, AccountDb: accountDb, CdrDb: cdrDb, Config: cfg}
	if cfg.RaterEnabled && !cfg.BalancerEnabled && cfg.RaterListen != INTERNAL {
		engine.Logger.Info(fmt.Sprintf("Starting CGRateS Rater on %s.", cfg.RaterListen))
		go listenToRPCRequests(responder, apier, cfg.RaterListen, cfg.RPCEncoding)
	}
	if cfg.BalancerEnabled {
		engine.Logger.Info(fmt.Sprintf("Starting CGRateS Balancer on %s.", cfg.BalancerListen))
		go stopBalancerSignalHandler()
		stopHandled = true
		responder.Bal = bal
		go listenToRPCRequests(responder, apier, cfg.BalancerListen, cfg.RPCEncoding)
		if cfg.RaterEnabled {
			engine.Logger.Info("Starting internal engine.")
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

	if cfg.SMEnabled {
		engine.Logger.Info("Starting CGRateS SessionManager.")
		go startSessionManager(responder, logDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if cfg.MediatorEnabled {
		engine.Logger.Info("Starting CGRateS Mediator.")
		go startMediator(responder, logDb, cdrDb)
	}

	if cfg.CDRSEnabled {
		engine.Logger.Info("Starting CGRateS CDR Server.")
		go startCDRS(responder, cdrDb)
	}

	if cfg.HistoryServerEnabled || cfg.HistoryAgentEnabled {
		engine.Logger.Info("Starting History Service.")
		go startHistoryScribe()
	}
	if cfg.CdrcEnabled {
		engine.Logger.Info("Starting CGRateS CDR Client.")
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
