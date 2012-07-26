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
	"database/sql"
	"flag"
	"fmt"
	"github.com/cgrates/cgrates/balancer"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
)

const (
	DISABLED = "disabled"
	INTERNAL = "internal"
)

var (
	config       = flag.String("config", "/home/rif/Documents/prog/go/src/github.com/cgrates/cgrates/conf/rater_standalone.config", "Configuration file location.")
	redis_server = "127.0.0.1:6379" // redis address host:port
	redis_db     = 10               // redis database number

	rater_enabled  = false            // start standalone server (no balancer)
	rater_balancer = DISABLED         // balancer address host:port
	rater_listen   = "127.0.0.1:1234" // listening address host:port
	rater_json     = false            // use JSON for RPC encoding

	balancer_enabled      = false
	balancer_listen_rater = "127.0.0.1:2000" // Rater server address	
	balancer_listen       = "127.0.0.1:2001" // Json RPC server address	
	balancer_json         = false            // use JSON for RPC encoding

	scheduler_enabled = false

	sm_enabled           = false
	sm_rater             = INTERNAL         // address where to access rater. Can be internal, direct rater address or the address of a balancer
	sm_freeswitch_server = "localhost:8021" // freeswitch address host:port
	sm_freeswitch_pass   = "ClueCon"        // reeswitch address host:port	
	sm_json              = false            // use JSON for RPC encoding

	mediator_enabled     = false
	mediator_cdr_file    = "Master.csv" // Freeswitch Master CSV CDR file.
	mediator_result_file = "out.csv"    // Generated file containing CDR and price info.
	mediator_host        = "localhost"  // The host to connect to. Values that start with / are for UNIX domain sockets.
	mediator_port        = "5432"       // The port to bind to.
	mediator_db          = "cgrates"    // The name of the database to connect to.
	mediator_user        = ""           // The user to sign in as.
	mediator_password    = ""           // The user's password.

	stats_enabled = false
	stats_listen  = "127.0.0.1:8000" // Web server address (for stat reports)

	bal      = balancer.NewBalancer()
	exitChan = make(chan bool)
)

func readConfig(configFn string) {
	c, err := conf.ReadConfigFile(configFn)
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("Could not open the configuration file: %v", err))
		return
	}
	redis_server, _ = c.GetString("global", "redis_server")
	redis_db, _ = c.GetInt("global", "redis_db")

	rater_enabled, _ = c.GetBool("rater", "enabled")
	rater_balancer, _ = c.GetString("rater", "balancer")
	rater_listen, _ = c.GetString("rater", "listen")
	rater_json, _ = c.GetBool("rater", "json")

	balancer_enabled, _ = c.GetBool("balancer", "enabled")
	balancer_listen_rater, _ = c.GetString("balancer", "listen_rater")
	balancer_listen, _ = c.GetString("balancer", "listen")
	balancer_json, _ = c.GetBool("balancer", "json")

	scheduler_enabled, _ = c.GetBool("scheduler", "enabled")

	sm_enabled, _ = c.GetBool("session_manager", "enabled")
	sm_rater, _ = c.GetString("session_manager", "rater")
	sm_freeswitch_server, _ = c.GetString("session_manager", "freeswitch_server")
	sm_freeswitch_pass, _ = c.GetString("session_manager", "freeswitch_pass")
	sm_json, _ = c.GetBool("session_manager", "json")

	mediator_enabled, _ = c.GetBool("mediator", "enabled")
	mediator_cdr_file, _ = c.GetString("mediator", "cdr_file")
	mediator_result_file, _ = c.GetString("mediator", "result_file")
	mediator_host, _ = c.GetString("mediator", "host")
	mediator_port, _ = c.GetString("mediator", "port")
	mediator_db, _ = c.GetString("mediator", "db")
	mediator_user, _ = c.GetString("mediator", "user")
	mediator_password, _ = c.GetString("mediator", "password")

	stats_enabled, _ = c.GetBool("stats_server", "enabled")
	stats_listen, _ = c.GetString("stats_server", "listen")
}

func listenToRPCRequests(rpcResponder interface{}, rpcAddress string, json bool) {
	l, err := net.Listen("tcp", rpcAddress)
	defer l.Close()

	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("could start the rpc server: %v", err))
		exitChan <- true
		return
	}

	timespans.Logger.Info(fmt.Sprintf("Listening for incomming RPC requests on %v", l.Addr()))
	rpc.Register(rpcResponder)

	for {
		conn, err := l.Accept()
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("accept error: %v", conn))
			continue
		}

		timespans.Logger.Info(fmt.Sprintf("connection started: %v", conn.RemoteAddr()))
		if json {
			// log.Print("json encoding")
			go jsonrpc.ServeConn(conn)
		} else {
			// log.Print("gob encoding")
			go rpc.ServeConn(conn)
		}
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

func startSessionManager(responder *timespans.Responder) {
	var connector sessionmanager.Connector
	if sm_rater == INTERNAL {
		connector = responder
	} else {
		var client *rpc.Client
		var err error
		if sm_json {
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
	sm := &sessionmanager.FSSessionManager{}
	sm.Connect(&sessionmanager.SessionDelegate{connector}, sm_freeswitch_server, sm_freeswitch_pass)
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
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	readConfig(*config)
	// some consitency checks
	go checkConfigSanity()

	getter, err := timespans.NewRedisStorage(redis_server, redis_db)
	if err != nil {
		timespans.Logger.Crit("Could not connect to redis, exiting!")
		exitChan <- true
	}
	defer getter.Close()
	timespans.SetStorageGetter(getter)

	if rater_enabled && rater_balancer != DISABLED && !balancer_enabled {
		go registerToBalancer()
		go stopRaterSingnalHandler()
	}
	responder := &timespans.Responder{ExitChan: exitChan}
	if rater_enabled && !balancer_enabled {
		go listenToRPCRequests(responder, rater_listen, rater_json)
	}
	if balancer_enabled {
		go stopBalancerSingnalHandler()
		go listenToRPCRequests(new(RaterServer), balancer_listen_rater, false)
		responder.Bal = bal
		go listenToRPCRequests(responder, balancer_listen, balancer_json)
		if rater_enabled {
			bal.AddClient("local", new(timespans.ResponderWorker))
		}
	}

	if stats_enabled {
		go listenToHttpRequests()
	}

	if scheduler_enabled {
		go func() {
			loadActionTimings(getter)
			go reloadSchedulerSingnalHandler(getter)
			s.loop()
		}()
	}

	if sm_enabled {
		go startSessionManager(responder)
	}

	if mediator_enabled {
		db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", mediator_host, mediator_port, mediator_db, mediator_user, mediator_password))
		//defer db.Close()
		if err != nil {
			timespans.Logger.Err(fmt.Sprintf("failed to open the database: %v", err))
		}
		var connector sessionmanager.Connector
		if sm_rater == INTERNAL {
			connector = responder
		} else {
			var client *rpc.Client
			var err error
			if sm_json {
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
		m := &Mediator{connector, db}
		_ = m
	}

	<-exitChan
}
