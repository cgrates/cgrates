/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012  Radu Ioan Fericean

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
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/balancer"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
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
	SAME     = "same"
	FS       = "freeswitch"
)

var (
	config              = flag.String("config", "rater_standalone.config", "Configuration file location.")
	redis_server        = "127.0.0.1:6379" // redis address host:port
	redis_db            = 10               // redis database number
	redis_pass          = ""
	logging_db_type     = MONGO
	logging_db_host     = "localhost" // The host to connect to. Values that start with / are for UNIX domain sockets.
	logging_db_port     = ""          // The port to bind to.
	logging_db_name     = "cgrates"   // The name of the database to connect to.
	logging_db_user     = ""          // The user to sign in as.
	logging_db_password = ""          // The user's password.

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

	mediator_enabled      = false
	mediator_cdr_file     = "Master.csv" // Freeswitch Master CSV CDR file.
	mediator_result_file  = "out.csv"    // Generated file containing CDR and price info.
	mediator_rater        = INTERNAL     // address where to access rater. Can be internal, direct rater address or the address of a balancer	
	mediator_rpc_encoding = GOB          // use JSON for RPC encoding
	mediator_skipdb       = false

	stats_enabled = false
	stats_listen  = "127.0.0.1:8000" // Web server address (for stat reports)

	freeswitch_server = "localhost:8021" // freeswitch address host:port
	freeswitch_pass   = "ClueCon"        // reeswitch address host:port	

	bal      = balancer.NewBalancer()
	exitChan = make(chan bool)
)

// this function will reset to zero values the variables that are not present
func readConfig(c *conf.ConfigFile) {
	redis_server, _ = c.GetString("global", "redis_server")
	redis_db, _ = c.GetInt("global", "redis_db")
	redis_pass, _ = c.GetString("global", "redis_pass")
	logging_db_type, _ = c.GetString("global", "logdb_type")
	logging_db_host, _ = c.GetString("global", "logdb_host")
	logging_db_port, _ = c.GetString("global", "logdb_port")
	logging_db_name, _ = c.GetString("global", "logdb_name")
	logging_db_user, _ = c.GetString("global", "logdb_user")
	logging_db_password, _ = c.GetString("global", "logdb_passwd")

	rater_enabled, _ = c.GetBool("rater", "enabled")
	rater_balancer, _ = c.GetString("rater", "balancer")
	rater_listen, _ = c.GetString("rater", "listen")
	rater_rpc_encoding, _ = c.GetString("rater", "rpc_encoding")

	balancer_enabled, _ = c.GetBool("balancer", "enabled")
	balancer_listen, _ = c.GetString("balancer", "listen")
	balancer_rpc_encoding, _ = c.GetString("balancer", "rpc_encoding")

	scheduler_enabled, _ = c.GetBool("scheduler", "enabled")

	sm_enabled, _ = c.GetBool("session_manager", "enabled")
	sm_switch_type, _ = c.GetString("session_manager", "switch_type")
	sm_rater, _ = c.GetString("session_manager", "rater")
	sm_debit_period, _ = c.GetInt("session_manager", "debit_period")
	sm_rpc_encoding, _ = c.GetString("session_manager", "rpc_encoding")

	mediator_enabled, _ = c.GetBool("mediator", "enabled")
	mediator_cdr_file, _ = c.GetString("mediator", "cdr_file")
	mediator_result_file, _ = c.GetString("mediator", "result_file")
	mediator_rater, _ = c.GetString("mediator", "rater")
	mediator_rpc_encoding, _ = c.GetString("mediator", "rpc_encoding")
	mediator_skipdb, _ = c.GetBool("mediator", "skipdb")

	stats_enabled, _ = c.GetBool("stats_server", "enabled")
	stats_listen, _ = c.GetString("stats_server", "listen")

	freeswitch_server, _ = c.GetString("freeswitch", "server")
	freeswitch_pass, _ = c.GetString("freeswitch", "pass")

}

func listenToRPCRequests(rpcResponder interface{}, rpcAddress string, rpc_encoding string) {
	l, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("could not connect to %v: %v", rpcAddress, err))
		exitChan <- true
		return
	}
	defer l.Close()

	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("could start the rpc server: %v", err))
		exitChan <- true
		return
	}

	timespans.Logger.Info(fmt.Sprintf("Listening for incomming RPC requests on %v", l.Addr()))
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
			timespans.Logger.Err(fmt.Sprintf("accept error: %v", conn))
			continue
		}

		timespans.Logger.Info(fmt.Sprintf("connection started: %v", conn.RemoteAddr()))
		go serveFunc(conn)
	}
}

func listenToHttpRequests() {
	http.Handle("/static/", http.FileServer(http.Dir("")))
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/getmem", memoryHandler)
	http.HandleFunc("/raters", ratersHandler)
	timespans.Logger.Info(fmt.Sprintf("The server is listening on %s", stats_listen))
	http.ListenAndServe(stats_listen, nil)
}

func startMediator(responder *timespans.Responder, loggerDb timespans.StorageGetter) {
	var connector sessionmanager.Connector
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
			timespans.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &sessionmanager.RPCClientConnector{client}
	}
	m := &Mediator{connector, loggerDb, mediator_skipdb}
	m.parseCSV()
}

func startSessionManager(responder *timespans.Responder, loggerDb timespans.StorageGetter) {
	var connector sessionmanager.Connector
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
			timespans.Logger.Crit(fmt.Sprintf("Could not connect to rater: %v", err))
			exitChan <- true
		}
		connector = &sessionmanager.RPCClientConnector{client}
	}
	switch sm_switch_type {
	case FS:
		sm := sessionmanager.NewFSSessionManager(loggerDb)
		sm.Connect(&sessionmanager.SessionDelegate{connector, time.Duration(sm_debit_period) * time.Second}, freeswitch_server, freeswitch_pass)
	default:
		timespans.Logger.Err(fmt.Sprintf("Cannot start session manger of type: %s!", sm_switch_type))
	}
}

func checkConfigSanity() {
	if sm_enabled && rater_enabled && rater_balancer != DISABLED {
		timespans.Logger.Crit("The session manager must not be enabled on a worker rater (change [rater]/balancer to disabled)!")
		exitChan <- true
	}
	if balancer_enabled && rater_enabled && rater_balancer != DISABLED {
		timespans.Logger.Crit("The balancer is enabled so it cannot connect to anatoher balancer (change [rater]/balancer to disabled)!")
		exitChan <- true
	}

	// check if the session manager or mediator is connectting via loopback 
	// if they are using the same encoding as the rater/balancer
	// this scenariou should be used for debug puropses only (it is racy anyway)
	// and it might be forbidden in the future
	if strings.Contains(sm_rater, "localhost") || strings.Contains(sm_rater, "127.0.0.1") {
		if balancer_enabled {
			if balancer_rpc_encoding != sm_rpc_encoding {
				timespans.Logger.Crit("If you are connecting the session manager via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if rater_enabled {
			if rater_rpc_encoding != sm_rpc_encoding {
				timespans.Logger.Crit("If you are connecting the session manager via the loopback to the arter use the same type of rpc encoding!")
				exitChan <- true
			}
		}
	}
	if strings.Contains(mediator_rater, "localhost") || strings.Contains(mediator_rater, "127.0.0.1") {
		if balancer_enabled {
			if balancer_rpc_encoding != mediator_rpc_encoding {
				timespans.Logger.Crit("If you are connecting the mediator via the loopback to the balancer use the same type of rpc encoding!")
				exitChan <- true
			}
		}
		if rater_enabled {
			if rater_rpc_encoding != mediator_rpc_encoding {
				timespans.Logger.Crit("If you are connecting the mediator via the loopback to the arter use the same type of rpc encoding!")
				exitChan <- true
			}
		}
	}
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	c, err := conf.ReadConfigFile(*config)
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("Could not open the configuration file: %v", err))
		return
	}
	readConfig(c)
	// some consitency checks
	go checkConfigSanity()

	getter, err := timespans.NewRedisStorage(redis_server, redis_db, redis_pass)
	//getter, err := timespans.NewMongoStorage("localhost", "cgrates")
	if err != nil {
		timespans.Logger.Crit("Could not connect to redis, exiting!")
		exitChan <- true
	}
	defer getter.Close()
	timespans.SetStorageGetter(getter)

	if sm_debit_period > 0 {
		timespans.SetDebitPeriod(time.Duration(sm_debit_period))
	}

	var loggerDb timespans.StorageGetter
	switch logging_db_type {
	case POSTGRES:
		loggerDb, err = timespans.NewPostgresStorage(logging_db_host, logging_db_port, logging_db_name, logging_db_user, logging_db_password)
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("Could not connect to logger database: %v", err))
		}
	case MONGO:
		loggerDb, err = timespans.NewMongoStorage(logging_db_host, logging_db_port, logging_db_name, logging_db_user, logging_db_password)
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("Could not connect to logger database: %v", err))
		}
	case SAME:
		loggerDb = getter
	default:
		timespans.Logger.Crit("Could not open logging database")
		exitChan <- true
	}

	timespans.SetStorageLogger(loggerDb)

	if rater_enabled && rater_balancer != DISABLED && !balancer_enabled {
		go registerToBalancer()
		go stopRaterSingnalHandler()
	}
	responder := &timespans.Responder{ExitChan: exitChan}
	if rater_enabled && !balancer_enabled && rater_listen != INTERNAL {
		timespans.Logger.Info(fmt.Sprintf("Starting CGRateS rater on %s.", rater_listen))
		go listenToRPCRequests(responder, rater_listen, rater_rpc_encoding)
	}
	if balancer_enabled {
		timespans.Logger.Info(fmt.Sprintf("Starting CGRateS balancer on %s.", balancer_listen))
		go stopBalancerSingnalHandler()
		responder.Bal = bal
		go listenToRPCRequests(responder, balancer_listen, balancer_rpc_encoding)
		if rater_enabled {
			timespans.Logger.Info("Starting internal rater.")
			bal.AddClient("local", new(timespans.ResponderWorker))
		}
	}

	if stats_enabled {
		timespans.Logger.Info(fmt.Sprintf("Starting CGRateS stats server on %v.", stats_listen))
		go listenToHttpRequests()
	}

	if scheduler_enabled {
		timespans.Logger.Info("Starting CGRateS scheduler.")
		go func() {
			loadActionTimings(getter)
			go reloadSchedulerSingnalHandler(getter)
			sched.loop()
		}()
	}

	if sm_enabled {
		timespans.Logger.Info("Starting CGRateS session manager.")
		go startSessionManager(responder, loggerDb)
	}

	if mediator_enabled {
		timespans.Logger.Info("Starting CGRateS mediator.")
		go startMediator(responder, loggerDb)
	}

	<-exitChan
}
