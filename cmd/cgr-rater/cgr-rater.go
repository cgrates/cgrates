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
	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/mediator"
	"github.com/cgrates/cgrates/rater"
	"github.com/cgrates/cgrates/scheduler"
	"github.com/cgrates/cgrates/sessionmanager"
	"io"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	DISABLED = "disabled"
	INTERNAL = "internal"
	JSON     = "json"
	GOB      = "gob"
	POSTGRES = "postgres"
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
	cfg      *CGRConfig
	err	error
)

func listenToRPCRequests(rpcResponder interface{}, rpcAddress string, rpc_encoding string) {
	l, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("<Rater> Could not listen to %v: %v", rpcAddress, err))
		exitChan <- true
		return
	}
	defer l.Close()

	rater.Logger.Info(fmt.Sprintf("<Rater> Listening for incomming RPC requests on %v", l.Addr()))
	rpc.Register(rpcResponder)
	var serveFunc func(io.ReadWriteCloser)
	if rpc_encoding == JSON {
		serveFunc = jsonrpc.ServeConn
	} else {
		serveFunc = rpc.ServeConn
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			rater.Logger.Err(fmt.Sprintf("<Rater> Accept error: %v", conn))
			continue
		}

		rater.Logger.Info(fmt.Sprintf("<Rater> New incoming connection: %v", conn.RemoteAddr()))
		go serveFunc(conn)
	}
}

func startMediator(responder *rater.Responder, loggerDb rater.DataStorage) {
	var connector rater.Connector
	if cfg.mediator_rater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if cfg.mediator_rpc_encoding == JSON {
			client, err = jsonrpc.Dial("tcp", cfg.mediator_rater)
		} else {
			client, err = rpc.Dial("tcp", cfg.mediator_rater)
		}
		if err != nil {
			rater.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &rater.RPCClientConnector{Client: client}
	}
	if _, err := os.Stat(cfg.mediator_cdr_path); err != nil {
		rater.Logger.Crit(fmt.Sprintf("The input path for mediator does not exist: %v", cfg.mediator_cdr_path))
		exitChan <- true
	}
	if _, err := os.Stat(cfg.mediator_cdr_out_path); err != nil {
		rater.Logger.Crit(fmt.Sprintf("The output path for mediator does not exist: %v", cfg.mediator_cdr_out_path))
		exitChan <- true
	}
	// ToDo: Why is here 
	// Check parsing errors
	//if cfgParseErr != nil {
	//	rater.Logger.Crit(fmt.Sprintf("Errors on config parsing: <%v>", cfgParseErr))
	//	exitChan <- true
	//}

	m, err := mediator.NewMediator(connector, loggerDb, cfg.mediator_skipdb, cfg.mediator_cdr_out_path, cfg.mediator_pseudo_prepaid, 
		cfg.freeswitch_direction, cfg.freeswitch_tor, cfg.freeswitch_tenant, cfg.freeswitch_subject, cfg.freeswitch_account,
		 cfg.freeswitch_destination, cfg.freeswitch_time_start, cfg.freeswitch_duration, cfg.freeswitch_uuid)
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Mediator config parsing error: %v", err))
		exitChan <- true
	}

	m.TrackCDRFiles(cfg.mediator_cdr_path)
}

func startSessionManager(responder *rater.Responder, loggerDb rater.DataStorage) {
	var connector rater.Connector
	if cfg.sm_rater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if cfg.sm_rpc_encoding == JSON {
			client, err = jsonrpc.Dial("tcp", cfg.sm_rater)
		} else {
			client, err = rpc.Dial("tcp", cfg.sm_rater)
		}
		if err != nil {
			rater.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &rater.RPCClientConnector{Client: client}
	}
	switch cfg.sm_switch_type {
	case FS:
		dp, _ := time.ParseDuration(fmt.Sprintf("%vs", cfg.sm_debit_period))
		sm = sessionmanager.NewFSSessionManager(loggerDb, connector, dp)
		errConn := sm.Connect(cfg.freeswitch_server, cfg.freeswitch_pass, cfg.freeswitch_reconnects)
		if errConn != nil {
			rater.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", errConn))
		}
	default:
		rater.Logger.Err(fmt.Sprintf("<SessionManager> Unsupported session manger type: %s!", cfg.sm_switch_type))
		exitChan <- true
	}
	exitChan <- true
}

func checkConfigSanity() {
	if cfg.sm_enabled && cfg.rater_enabled && cfg.rater_balancer != DISABLED {
		rater.Logger.Crit("The session manager must not be enabled on a worker rater (change [rater]/balancer to disabled)!")
		exitChan <- true
	}
	if cfg.balancer_enabled && cfg.rater_enabled && cfg.rater_balancer != DISABLED {
		rater.Logger.Crit("The balancer is enabled so it cannot connect to anatoher balancer (change [rater]/balancer to disabled)!")
		exitChan <- true
	}

	// check if the session manager or mediator is connectting via loopback
	// if they are using the same encoding as the rater/balancer
	// this scenariou should be used for debug puropses only (it is racy anyway)
	// and it might be forbidden in the future
	if strings.Contains(cfg.sm_rater, "localhost") || strings.Contains(cfg.sm_rater, "127.0.0.1") {
		if cfg.balancer_enabled {
			if cfg.balancer_rpc_encoding != cfg.sm_rpc_encoding {
				rater.Logger.Crit("If you are connecting the session manager via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if cfg.rater_enabled {
			if cfg.rater_rpc_encoding != cfg.sm_rpc_encoding {
				rater.Logger.Crit("If you are connecting the session manager via the loopback to the arter use the same type of rpc encoding!")
				exitChan <- true
			}
		}
	}
	if strings.Contains(cfg.mediator_rater, "localhost") || strings.Contains(cfg.mediator_rater, "127.0.0.1") {
		if cfg.balancer_enabled {
			if cfg.balancer_rpc_encoding != cfg.mediator_rpc_encoding {
				rater.Logger.Crit("If you are connecting the mediator via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if cfg.rater_enabled {
			if cfg.rater_rpc_encoding != cfg.mediator_rpc_encoding {
				rater.Logger.Crit("If you are connecting the mediator via the loopback to the arter use the same type of rpc encoding!")
				exitChan <- true
			}
		}
	}
}

func configureDatabase(db_type, host, port, name, user, pass string) (getter rater.DataStorage, err error) {
	switch db_type {
	case REDIS:
		db_nb, err := strconv.Atoi(name)
		if err != nil {
			rater.Logger.Crit("Redis db name must be an integer!")
			exitChan <- true
		}
		if port != "" {
			host += ":" + port
		}
		getter, err = rater.NewRedisStorage(host, db_nb, pass)
	case MONGO:
		getter, err = rater.NewMongoStorage(host, port, name, user, pass)
	case POSTGRES:
		getter, err = rater.NewPostgresStorage(host, port, name, user, pass)
	default:
		err = errors.New("unknown db")
		rater.Logger.Crit("Unknown db type, exiting!")
		exitChan <- true
	}

	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Could not connect to db: %v, exiting!", err))
		exitChan <- true
	}
	return
}

func main() {
	flag.Parse()
	if *version {
		fmt.Println("CGRateS " + rater.VERSION)
		return
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg, err = NewCGRConfig( cfgPath )
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Could not parse config: %s exiting!", err))
		return
	}
	// some consitency checks
	go checkConfigSanity()

	var getter, loggerDb rater.DataStorage
	getter, err = configureDatabase(cfg.data_db_type, cfg.data_db_host, cfg.data_db_port, cfg.data_db_name, cfg.data_db_user, cfg.data_db_pass)

	if err == nil {
		defer getter.Close()
		rater.SetDataStorage(getter)
	}

	if cfg.log_db_type == SAME {
		loggerDb = getter
	} else {
		loggerDb, err = configureDatabase(cfg.log_db_type, cfg.log_db_host, cfg.log_db_port, cfg.log_db_name, cfg.log_db_user, cfg.log_db_pass)
	}
	if err == nil {
		defer loggerDb.Close()
		rater.SetStorageLogger(loggerDb)
	}

	if cfg.sm_debit_period > 0 {
		if dp, err := time.ParseDuration(fmt.Sprintf("%vs", cfg.sm_debit_period)); err == nil {
			rater.SetDebitPeriod(dp)
		}
	}

	if cfg.rater_enabled && cfg.rater_balancer != DISABLED && !cfg.balancer_enabled {
		go registerToBalancer()
		go stopRaterSingnalHandler()
	}
	responder := &rater.Responder{ExitChan: exitChan}
	if cfg.rater_enabled && !cfg.balancer_enabled && cfg.rater_listen != INTERNAL {
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS Rater on %s.", cfg.rater_listen))
		go listenToRPCRequests(responder, cfg.rater_listen, cfg.rater_rpc_encoding)
	}
	if cfg.balancer_enabled {
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS Balancer on %s.", cfg.balancer_listen))
		go stopBalancerSingnalHandler()
		responder.Bal = bal
		go listenToRPCRequests(responder, cfg.balancer_listen, cfg.balancer_rpc_encoding)
		if cfg.rater_enabled {
			rater.Logger.Info("Starting internal rater.")
			bal.AddClient("local", new(rater.ResponderWorker))
		}
	}

	if cfg.scheduler_enabled {
		rater.Logger.Info("Starting CGRateS Scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, getter)
			sched.LoadActionTimings(getter)
			sched.Loop()
		}()
	}

	if cfg.sm_enabled {
		rater.Logger.Info("Starting CGRateS SessionManager.")
		go startSessionManager(responder, loggerDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if cfg.mediator_enabled {
		rater.Logger.Info("Starting CGRateS Mediator.")
		go startMediator(responder, loggerDb)
	}
	<-exitChan
	rater.Logger.Info("Stopped all components. CGRateS shutdown!")
}
