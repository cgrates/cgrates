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

	cfgParseErr error

	bal      = balancer2go.NewBalancer()
	exitChan = make(chan bool)
	sm       sessionmanager.SessionManager
)

// this function will reset to zero values the variables that are not present
func readConfig(c *conf.ConfigFile) {
	data_db_type, _ = c.GetString("global", "datadb_type")
	data_db_host, _ = c.GetString("global", "datadb_host")
	data_db_port, _ = c.GetString("global", "datadb_port")
	data_db_name, _ = c.GetString("global", "datadb_name")
	data_db_user, _ = c.GetString("global", "datadb_user")
	data_db_pass, _ = c.GetString("global", "datadb_passwd")
	log_db_type, _ = c.GetString("global", "logdb_type")
	log_db_host, _ = c.GetString("global", "logdb_host")
	log_db_port, _ = c.GetString("global", "logdb_port")
	log_db_name, _ = c.GetString("global", "logdb_name")
	log_db_user, _ = c.GetString("global", "logdb_user")
	log_db_pass, _ = c.GetString("global", "logdb_passwd")

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
	mediator_cdr_path, _ = c.GetString("mediator", "cdr_path")
	mediator_cdr_out_path, _ = c.GetString("mediator", "cdr_out_path")
	mediator_rater, _ = c.GetString("mediator", "rater")
	mediator_rpc_encoding, _ = c.GetString("mediator", "rpc_encoding")
	mediator_skipdb, _ = c.GetBool("mediator", "skipdb")
	mediator_pseudo_prepaid, _ = c.GetBool("mediator", "pseudo_prepaid")

	freeswitch_server, _ = c.GetString("freeswitch", "server")
	freeswitch_pass, _ = c.GetString("freeswitch", "pass")
	freeswitch_tor, _ = c.GetString("freeswitch", "tor_index")
	freeswitch_tenant, _ = c.GetString("freeswitch", "tenant_index")
	freeswitch_direction, _ = c.GetString("freeswitch", "direction_index")
	freeswitch_subject, _ = c.GetString("freeswitch", "subject_index")
	freeswitch_account, _ = c.GetString("freeswitch", "account_index")
	freeswitch_destination, _ = c.GetString("freeswitch", "destination_index")
	freeswitch_time_start, _ = c.GetString("freeswitch", "time_start_index")
	freeswitch_duration, _ = c.GetString("freeswitch", "duration_index")
	freeswitch_uuid, _ = c.GetString("freeswitch", "uuid_index")
}

func listenToRPCRequests(rpcResponder interface{}, rpcAddress string, rpc_encoding string) {
	l, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		rater.Logger.Crit(fmt.Sprintf("Could not listen to %v: %v", rpcAddress, err))
		exitChan <- true
		return
	}
	defer l.Close()

	rater.Logger.Info(fmt.Sprintf("Listening for incomming RPC requests on %v", l.Addr()))
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
			rater.Logger.Err(fmt.Sprintf("accept error: %v", conn))
			continue
		}

		rater.Logger.Info(fmt.Sprintf("connection started: %v", conn.RemoteAddr()))
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
		sm.Connect(freeswitch_server, freeswitch_pass)
	default:
		rater.Logger.Err(fmt.Sprintf("Cannot start session manger of type: %s!", sm_switch_type))
		exitChan <- true
	}
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
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS rater on %s.", rater_listen))
		go listenToRPCRequests(responder, rater_listen, rater_rpc_encoding)
	}
	if balancer_enabled {
		rater.Logger.Info(fmt.Sprintf("Starting CGRateS balancer on %s.", balancer_listen))
		go stopBalancerSingnalHandler()
		responder.Bal = bal
		go listenToRPCRequests(responder, balancer_listen, balancer_rpc_encoding)
		if rater_enabled {
			rater.Logger.Info("Starting internal rater.")
			bal.AddClient("local", new(rater.ResponderWorker))
		}
	}

	if scheduler_enabled {
		rater.Logger.Info("Starting CGRateS scheduler.")
		go func() {
			sched := scheduler.NewScheduler()
			go reloadSchedulerSingnalHandler(sched, getter)
			sched.LoadActionTimings(getter)
			sched.Loop()
		}()
	}

	if sm_enabled {
		rater.Logger.Info("Starting CGRateS session manager.")
		go startSessionManager(responder, loggerDb)
		// close all sessions on shutdown
		go shutdownSessionmanagerSingnalHandler()
	}

	if mediator_enabled {
		rater.Logger.Info("Starting CGRateS mediator.")
		go startMediator(responder, loggerDb)
	}
	<-exitChan
}
