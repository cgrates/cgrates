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
	"code.google.com/p/goconf/conf"
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
	config       = flag.String("config", "/etc/cgrates/cgrates.cfg", "Configuration file location.")
	version      = flag.Bool("version", false, "Prints the application version.")
	data_db_type = REDIS
	data_db_host = "localhost" // The host to connect to. Values that start with / are for UNIX domain sockets.
	data_db_port = ""          // The port to bind to.
	data_db_name = "10"        // The name of the database to connect to.
	data_db_user = ""          // The user to sign in as.
	data_db_pass = ""          // The user's password.
	log_db_type  = MONGO
	log_db_host  = "localhost" // The host to connect to. Values that start with / are for UNIX domain sockets.
	log_db_port  = ""          // The port to bind to.
	log_db_name  = "cgrates"   // The name of the database to connect to.
	log_db_user  = ""          // The user to sign in as.
	log_db_pass  = ""          // The user's password.

	rater_enabled      = false            // start standalone server (no balancer)
	rater_balancer     = DISABLED         // balancer address host:port
	rater_listen       = "127.0.0.1:1234" // listening address host:port
	rater_rpc_encoding = GOB              // use JSON for RPC encoding

	balancer_enabled      = false
	balancer_listen       = "127.0.0.1:2001" // Json RPC server address
	balancer_rpc_encoding = GOB              // use JSON for RPC encoding

	scheduler_enabled = false

	sm_enabled      = false
	sm_switch_type  = FS
	sm_rater        = INTERNAL // address where to access rater. Can be internal, direct rater address or the address of a balancer
	sm_debit_period = 10       // the period to be debited in advanced during a call (in seconds)
	sm_rpc_encoding = GOB      // use JSON for RPC encoding

	mediator_enabled        = false
	mediator_cdr_path       = ""       // Freeswitch Master CSV CDR path.
	mediator_cdr_out_path   = ""       // Freeswitch Master CSV CDR output path.
	mediator_rater          = INTERNAL // address where to access rater. Can be internal, direct rater address or the address of a balancer
	mediator_rpc_encoding   = GOB      // use JSON for RPC encoding
	mediator_skipdb         = false
	mediator_pseudo_prepaid = false

	freeswitch_server      = "localhost:8021" // freeswitch address host:port
	freeswitch_pass        = "ClueCon"        // reeswitch address host:port
	freeswitch_direction   = ""
	freeswitch_tor         = ""
	freeswitch_tenant      = ""
	freeswitch_subject     = ""
	freeswitch_account     = ""
	freeswitch_destination = ""
	freeswitch_time_start  = ""
	freeswitch_duration    = ""
	freeswitch_uuid        = ""
	freeswitch_reconnects = 5

	cfgParseErr error

	bal      = balancer2go.NewBalancer()
	exitChan = make(chan bool)
	sm       sessionmanager.SessionManager
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
	if mediator_rater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if mediator_rpc_encoding == JSON {
			client, err = jsonrpc.Dial("tcp", mediator_rater)
		} else {
			client, err = rpc.Dial("tcp", mediator_rater)
		}
		if err != nil {
			rater.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &rater.RPCClientConnector{Client: client}
	}
	if _, err := os.Stat(mediator_cdr_path); err != nil {
		rater.Logger.Crit(fmt.Sprintf("The input path for mediator does not exist: %v", mediator_cdr_path))
		exitChan <- true
	}
	if _, err := os.Stat(mediator_cdr_out_path); err != nil {
		rater.Logger.Crit(fmt.Sprintf("The output path for mediator does not exist: %v", mediator_cdr_out_path))
		exitChan <- true
	}
	// Check parsing errors
	if cfgParseErr != nil {
		rater.Logger.Crit(fmt.Sprintf("Errors on config parsing: <%v>", cfgParseErr))
		exitChan <- true
	}

	m, err := mediator.NewMediator(connector, loggerDb, mediator_skipdb, mediator_cdr_out_path, mediator_pseudo_prepaid, freeswitch_direction,
		freeswitch_tor, freeswitch_tenant, freeswitch_subject, freeswitch_account, freeswitch_destination,
		freeswitch_time_start, freeswitch_duration, freeswitch_uuid)
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Mediator config parsing error: %v", err))
		exitChan <- true
	}

	m.TrackCDRFiles(mediator_cdr_path)
}

func startSessionManager(responder *rater.Responder, loggerDb rater.DataStorage) {
	var connector rater.Connector
	if sm_rater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if sm_rpc_encoding == JSON {
			client, err = jsonrpc.Dial("tcp", sm_rater)
		} else {
			client, err = rpc.Dial("tcp", sm_rater)
		}
		if err != nil {
			rater.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &rater.RPCClientConnector{Client: client}
	}
	switch sm_switch_type {
	case FS:
		dp, _ := time.ParseDuration(fmt.Sprintf("%vs", sm_debit_period))
		sm = sessionmanager.NewFSSessionManager(loggerDb, connector, dp)
		errConn := sm.Connect(freeswitch_server, freeswitch_pass, freeswitch_reconnects)
		if errConn != nil {
			rater.Logger.Err(fmt.Sprintf("<SessionManager> error: %s!", errConn))
		}
	default:
		rater.Logger.Err(fmt.Sprintf("<SessionManager> Unsupported session manger type: %s!", sm_switch_type))
		exitChan <- true
	}
	exitChan <-true
}

func checkConfigSanity() {
	if sm_enabled && rater_enabled && rater_balancer != DISABLED {
		rater.Logger.Crit("The session manager must not be enabled on a worker rater (change [rater]/balancer to disabled)!")
		exitChan <- true
	}
	if balancer_enabled && rater_enabled && rater_balancer != DISABLED {
		rater.Logger.Crit("The balancer is enabled so it cannot connect to anatoher balancer (change [rater]/balancer to disabled)!")
		exitChan <- true
	}

	// check if the session manager or mediator is connectting via loopback
	// if they are using the same encoding as the rater/balancer
	// this scenariou should be used for debug puropses only (it is racy anyway)
	// and it might be forbidden in the future
	if strings.Contains(sm_rater, "localhost") || strings.Contains(sm_rater, "127.0.0.1") {
		if balancer_enabled {
			if balancer_rpc_encoding != sm_rpc_encoding {
				rater.Logger.Crit("If you are connecting the session manager via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if rater_enabled {
			if rater_rpc_encoding != sm_rpc_encoding {
				rater.Logger.Crit("If you are connecting the session manager via the loopback to the arter use the same type of rpc encoding!")
				exitChan <- true
			}
		}
	}
	if strings.Contains(mediator_rater, "localhost") || strings.Contains(mediator_rater, "127.0.0.1") {
		if balancer_enabled {
			if balancer_rpc_encoding != mediator_rpc_encoding {
				rater.Logger.Crit("If you are connecting the mediator via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if rater_enabled {
			if rater_rpc_encoding != mediator_rpc_encoding {
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
	c, err := conf.ReadConfigFile(*config)
	if err != nil {
		rater.Logger.Err(fmt.Sprintf("Could not open the configuration file: %v", err))
		return
	}
	readConfig(c)
	// some consitency checks
	go checkConfigSanity()

	var getter, loggerDb rater.DataStorage
	getter, err = configureDatabase(data_db_type, data_db_host, data_db_port, data_db_name, data_db_user, data_db_pass)

	if err == nil {
		defer getter.Close()
		rater.SetDataStorage(getter)
	}

	if log_db_type == SAME {
		loggerDb = getter
	} else {
		loggerDb, err = configureDatabase(log_db_type, log_db_host, log_db_port, log_db_name, log_db_user, log_db_pass)
	}
	if err == nil {
		defer loggerDb.Close()
		rater.SetStorageLogger(loggerDb)
	}

	if sm_debit_period > 0 {
		if dp, err := time.ParseDuration(fmt.Sprintf("%vs", sm_debit_period)); err == nil {
			rater.SetDebitPeriod(dp)
		}
	}

	if rater_enabled && rater_balancer != DISABLED && !balancer_enabled {
		go registerToBalancer()
		go stopRaterSingnalHandler()
	}
	responder := &rater.Responder{ExitChan: exitChan}
	if rater_enabled && !balancer_enabled && rater_listen != INTERNAL {
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS Rater on %s.", rater_listen))
		go listenToRPCRequests(responder, rater_listen, rater_rpc_encoding)
	}
	if balancer_enabled {
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS Balancer on %s.", balancer_listen))
		go stopBalancerSingnalHandler()
		responder.Bal = bal
		go listenToRPCRequests(responder, balancer_listen, balancer_rpc_encoding)
		if rater_enabled {
			rater.Logger.Info("Starting internal rater.")
			bal.AddClient("local", new(rater.ResponderWorker))
		}
	}

	if scheduler_enabled {
		rater.Logger.Info("Starting CGRateS Scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, getter)
			sched.LoadActionTimings(getter)
			sched.Loop()
		}()
	}

	if sm_enabled {
		rater.Logger.Info("Starting CGRateS SessionManager.")
		go startSessionManager(responder, loggerDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if mediator_enabled {
		rater.Logger.Info("Starting CGRateS Mediator.")
		go startMediator(responder, loggerDb)
	}
	<-exitChan
	rater.Logger.Info("Stopped all components. CGRateS shutdown!")
}
