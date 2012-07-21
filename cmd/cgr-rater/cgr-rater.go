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
	"github.com/cgrates/cgrates/timespans"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

var (
	config       = flag.String("config", "/home/rif/Documents/prog/go/src/github.com/cgrates/cgrates/conf/rater.config", "Configuration file location.")
	redis_server = "127.0.0.1:6379" // redis address host:port
	redis_db     = 10               // redis database number

	rater_standalone      = false            // start standalone server (no balancer)
	rater_balancer_server = "127.0.0.1:2000" // balancer address host:port
	rater_listen          = "127.0.0.1:1234" // listening address host:port
	rater_json            = false            // use JSON for RPC encoding

	balancer_enabled      = false
	balancer_standalone   = false            // run standalone
	balancer_listen_rater = "127.0.0.1:2000" // Rater server address
	balancer_listen_api   = "127.0.0.1:2001" // Json RPC server address
	balancer_json         = false            // use JSON for RPC encoding

	scheduler_enabled    = false
	scheduler_standalone = false // run standalone (no other service)
	scheduler_json       = false

	sm_enabled           = false
	sm_standalone        = false            // run standalone 
	sm_api_server        = "127.0.0.1:2000" // balancer address host:port
	sm_freeswitch_server = "localhost:8021" // freeswitch address host:port
	sm_freeswitch_pass   = "ClueCon"        // reeswitch address host:port
	sm_json              = false            // use JSON for RPC encoding

	mediator_enabled     = false
	mediator_standalone  = false        // run standalone
	mediator_cdr_file    = "Master.csv" // Freeswitch Master CSV CDR file.
	mediator_result_file = "out.csv"    // Generated file containing CDR and price info.
	mediator_host        = "localhost"  // The host to connect to. Values that start with / are for UNIX domain sockets.
	mediator_port        = "5432"       // The port to bind to.
	mediator_db          = "cgrates"    // The name of the database to connect to.
	mediator_user        = ""           // The user to sign in as.
	mediator_password    = ""           // The user's password.

	bal      = balancer.NewBalancer()
	accLock  = timespans.NewAccountLock()
	exitChan = make(chan bool)
)

func readConfig(configFn string) {
	flag.Parse()
	c, err := conf.ReadConfigFile(configFn)
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("Could not open the configuration file: %v", err))
		return
	}
	redis_server, _ = c.GetString("global", "redis_server")
	redis_db, _ = c.GetInt("global", "redis_db")

	rater_standalone, _ = c.GetBool("rater", "standalone")
	rater_balancer_server, _ = c.GetString("rater", "balancer_server")
	rater_listen, _ = c.GetString("rater", "listen_api")
	rater_json, _ = c.GetBool("rater", "json")

	balancer_enabled, _ = c.GetBool("balancer", "enabled")
	balancer_standalone, _ = c.GetBool("balancer", "standalone")
	balancer_listen_rater, _ = c.GetString("balancer", "listen_rater")
	balancer_listen_api, _ = c.GetString("balancer", "listen_api")
	balancer_json, _ = c.GetBool("balancer", "json")

	scheduler_enabled, _ = c.GetBool("scheduler", "enabled")
	scheduler_standalone, _ = c.GetBool("scheduler", "standalone")
	scheduler_json, _ = c.GetBool("scheduler", "json")

	sm_enabled, _ = c.GetBool("session_manager", "enabled")
	sm_standalone, _ = c.GetBool("session_manager", "standalone")
	sm_api_server, _ = c.GetString("session_manager", "api_server")
	sm_freeswitch_server, _ = c.GetString("session_manager", "freeswitch_server")
	sm_freeswitch_pass, _ = c.GetString("session_manager", "freeswitch_pass")
	sm_json, _ = c.GetBool("session_manager", "json")

	mediator_enabled, _ = c.GetBool("mediator", "enabled")
	mediator_standalone, _ = c.GetBool("mediator", "standalone")
	mediator_cdr_file, _ = c.GetString("mediator", "cdr_file")
	mediator_result_file, _ = c.GetString("mediator", "result_file")
	mediator_host, _ = c.GetString("mediator", "host")
	mediator_port, _ = c.GetString("mediator", "port")
	mediator_db, _ = c.GetString("mediator", "db")
	mediator_user, _ = c.GetString("mediator", "user")
	mediator_password, _ = c.GetString("mediator", "password")
}

func resolveStandaloneConfilcts() {
	if balancer_standalone {
		rater_standalone = false
	}
	if scheduler_standalone {
		rater_standalone = false
		balancer_standalone = false
	}
	if sm_standalone {
		rater_standalone = false
		balancer_standalone = false
		scheduler_standalone = false
	}
	if mediator_standalone {
		rater_standalone = false
		balancer_standalone = false
		scheduler_standalone = false
		sm_standalone = false
	}
}

func listenToRPCRequests(responder interface{}, rpcAddress string, json bool) {
	l, err := net.Listen("tcp", rpcAddress)
	defer l.Close()

	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("could not connect to rpc server: %v", err))
	}

	timespans.Logger.Info(fmt.Sprintf("Listening for incomming RPC requests on %v", l.Addr()))
	rpc.Register(responder)

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

/*func listenToHttpRequests() {
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/getmem", memoryHandler)
	http.HandleFunc("/raters", ratersHandler)
	log.Print("The server is listening on ", *httpApiAddress)
	http.ListenAndServe(*httpApiAddress, nil)
}
*/
func main() {
	readConfig(*config)
	resolveStandaloneConfilcts()
	getter, err := timespans.NewRedisStorage(redis_server, redis_db)
	if err != nil {
		timespans.Logger.Crit("Could not connect to redis, exiting!")
		exitChan <- true
	}
	defer getter.Close()
	timespans.SetStorageGetter(getter)

	if !rater_standalone {
		go registerToBalancer(rater_balancer_server, rater_listen)
		go stopRaterSingnalHandler(rater_balancer_server, rater_listen, getter)
	}
	go listenToRPCRequests(&Responder{new(DirectResponder)}, rater_listen, false)
	<-exitChan
}
