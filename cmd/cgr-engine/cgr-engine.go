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
	"github.com/cgrates/cgrates/apier"
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/cdrs"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/history"
	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
	"time"
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
	cfgPath  = flag.String("config", "/etc/cgrates/cgrates.cfg", "Configuration file location.")
	version  = flag.Bool("version", false, "Prints the application version.")
	bal      = balancer2go.NewBalancer()
	exitChan = make(chan bool)
	sm       sessionmanager.SessionManager
	medi     *mediator.Mediator
	cfg      *config.CGRConfig
	err      error
)

func listenToRPCRequests(rpcResponder interface{}, apier *apier.Apier, rpcAddress string, rpc_encoding string, getter engine.DataStorage, loggerDb engine.DataStorage) {
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

func startMediator(responder *engine.Responder, loggerDb engine.DataStorage) {
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
	medi, err = mediator.NewMediator(connector, loggerDb, cfg)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Mediator config parsing error: %v", err))
		exitChan <- true
	}

	if cfg.MediatorCDRType == utils.FSCDR_FILE_CSV { //Mediator as standalone service for file CDRs
		medi.TrackCDRFiles()
	}
}

func startSessionManager(responder *engine.Responder, loggerDb engine.DataStorage) {
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

func startCDRS(responder *engine.Responder, loggerDb engine.DataStorage) {
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
	cs := cdrs.New(loggerDb, medi, cfg)
	cs.StartCapturingCDRs()
	exitChan <- true
}

func checkConfigSanity() error {
	if cfg.SMEnabled && cfg.RaterEnabled && cfg.RaterBalancer != DISABLED {
		engine.Logger.Crit("The session manager must not be enabled on a worker engine (change [engine]/balancer to disabled)!")
		return errors.New("SessionManager on Worker")
	}
	if cfg.BalancerEnabled && cfg.RaterEnabled && cfg.RaterBalancer != DISABLED {
		engine.Logger.Crit("The balancer is enabled so it cannot connect to anatoher balancer (change [engine]/balancer to disabled)!")
		return errors.New("Improperly configured balancer")
	}

	return nil
}

func startHistoryScribe() (err error) {
	var scribe history.Scribe
	flag.Parse()
	if "*masterAddr" != "" {
		scribe, err = history.NewProxyScribe("*masterAddr")
	} else {
		scribe, err = history.NewFileScribe("*dataFile")
	}
	rpc.RegisterName("Scribe", scribe)
	rpc.HandleHTTP()
	_, e := net.Listen("tcp", ":1234")
	if e != nil {
		return err
	}
	//http.Serve(l, nil)
	return nil
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + utils.VERSION)
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg, err = config.NewCGRConfig(cfgPath)
	if err != nil {
		engine.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
		return
	}
	// some consitency checks
	errCfg := checkConfigSanity()
	if errCfg != nil {
		engine.Logger.Crit(errCfg.Error())
		return
	}

	var getter, loggerDb engine.DataStorage
	getter, err = engine.ConfigureDatabase(cfg.DataDBType, cfg.DataDBHost, cfg.DataDBPort, cfg.DataDBName, cfg.DataDBUser, cfg.DataDBPass)
	if err != nil { // Cannot configure getter database, show stopper
		engine.Logger.Crit(fmt.Sprintf("Could not configure dataDb: %s exiting!", err))
		return
	}
	defer getter.Close()
	engine.SetDataStorage(getter)
	if cfg.StorDBType == SAME {
		loggerDb = getter
	} else {
		loggerDb, err = engine.ConfigureDatabase(cfg.StorDBType, cfg.StorDBHost, cfg.StorDBPort, cfg.StorDBName, cfg.StorDBUser, cfg.StorDBPass)
		if err != nil { // Cannot configure logger database, show stopper
			engine.Logger.Crit(fmt.Sprintf("Could not configure logger database: %s exiting!", err))
			return
		}
	}
	defer loggerDb.Close()
	engine.SetStorageLogger(loggerDb)
	engine.SetRoundingMethodAndDecimals(cfg.RoundingMethod, cfg.RoundingDecimals)

	if cfg.SMDebitInterval > 0 {
		if dp, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.SMDebitInterval)); err == nil {
			engine.SetDebitPeriod(dp)
		}
	}

	// Async starts here
	if cfg.RaterEnabled && cfg.RaterBalancer != DISABLED && !cfg.BalancerEnabled {
		go registerToBalancer()
		go stopRaterSingnalHandler()
	}
	responder := &engine.Responder{ExitChan: exitChan}
	apier := &apier.Apier{StorDb: loggerDb, DataDb: getter}
	if cfg.RaterEnabled && !cfg.BalancerEnabled && cfg.RaterListen != INTERNAL {
		engine.Logger.Info(fmt.Sprintf("Starting CGRateS Rater on %s.", cfg.RaterListen))
		go listenToRPCRequests(responder, apier, cfg.RaterListen, cfg.RPCEncoding, getter, loggerDb)
	}
	if cfg.BalancerEnabled {
		engine.Logger.Info(fmt.Sprintf("Starting CGRateS Balancer on %s.", cfg.BalancerListen))
		go stopBalancerSingnalHandler()
		responder.Bal = bal
		go listenToRPCRequests(responder, apier, cfg.BalancerListen, cfg.RPCEncoding, getter, loggerDb)
		if cfg.RaterEnabled {
			engine.Logger.Info("Starting internal engine.")
			bal.AddClient("local", new(engine.ResponderWorker))
		}
	}

	if cfg.SchedulerEnabled {
		engine.Logger.Info("Starting CGRateS Scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, getter)
			apier.Sched = sched
			sched.LoadActionTimings(getter)
			sched.Loop()
		}()
	}

	if cfg.SMEnabled {
		engine.Logger.Info("Starting CGRateS SessionManager.")
		go startSessionManager(responder, loggerDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if cfg.MediatorEnabled {
		engine.Logger.Info("Starting CGRateS Mediator.")
		go startMediator(responder, loggerDb)
	}

	if cfg.CDRSListen != "" {
		engine.Logger.Info("Starting CGRateS CDR Server.")
		go startCDRS(responder, loggerDb)
	}
	<-exitChan
	engine.Logger.Info("Stopped all components. CGRateS shutdown!")
}
