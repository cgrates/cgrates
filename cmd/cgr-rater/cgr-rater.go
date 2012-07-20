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
	"github.com/cgrates/cgrates/timespans"
)

var (
	config       = flag.String("config", "/home/rif/Documents/prog/go/src/github.com/cgrates/cgrates/data/cgrates.config", "Configuration file location.")
	redis_server = "127.0.0.1:6379" //"redis address host:port"
	redis_db     = 10               // "redis database number"

	rater_standalone      = false            // "start standalone server (no balancer)"
	rater_balancer_server = "127.0.0.1:2000" // "balancer address host:port"
	rater_listen          = "127.0.0.1:1234" // "listening address host:port"
	rater_json            = false            // "use JSON for RPC encoding"

	balancer_standalone   = false            // "run standalone (run as a rater)")
	balancer_listen_rater = "127.0.0.1:2000" // "Rater server address (localhost:2000)"
	balancer_listen_api   = "127.0.0.1:2001" // "Json RPC server address (localhost:2001)"
	balancer_json         = false            // "use JSON for RPC encoding"

	scheduler_standalone = false // "run standalone (no other service)")
	scheduler_json       = false

	sm_standalone     = false            // "run standalone (run as a rater)")
	sm_api_server     = "127.0.0.1:2000" // "balancer address host:port"
	sm_freeswitchsrv  = "localhost:8021" // "freeswitch address host:port"
	sm_freeswitchpass = "ClueCon"        // freeswitch address host:port"
	sm_json           = false            // "use JSON for RPC encoding"
)

func readConfig() {
	flag.Parse()
	c, err := conf.ReadConfigFile(*config)
	if err != nil {
		timespans.Logger.Err(fmt.Sprintf("Could not open the configuration file: %v", err))
	}
	redis_server, err = c.GetString("global", "redis_server")
	if err != nil {
	}
	redis_db, err = c.GetInt("global", "redis_db")
	if err != nil {
	}
	rater_standalone, err = c.GetBool("rater", "standalone")
	if err != nil {
	}
	rater_balancer_server, err = c.GetString("rater", "balancer_server")
	if err != nil {
	}
	rater_listen, err = c.GetString("rater", "listen_api")
	if err != nil {
	}
	rater_json, err = c.GetBool("rater", "json")
	if err != nil {
	}
	balancer_standalone, err = c.GetBool("balancer", "standalone")
	if err != nil {
	}
	balancer_listen_rater, err = c.GetString("balancer", "listen_rater")
	if err != nil {
	}
	balancer_listen_api, err = c.GetString("balancer", "listen_api")
	if err != nil {
	}
	balancer_json, err = c.GetBool("balancer", "json")
	if err != nil {
	}
	scheduler_standalone, err = c.GetBool("scheduler", "standalone")
	if err != nil {
	}
	scheduler_json, err = c.GetBool("scheduler", "json")
	if err != nil {
	}
	sm_standalone, err = c.GetBool("session_manager", "standalone")
	if err != nil {
	}
	sm_api_server, err = c.GetString("session_manager", "api_server")
	if err != nil {
	}
	sm_freeswitchsrv, err = c.GetString("session_manager", "freeswitch_server")
	if err != nil {
	}
	sm_freeswitchpass, err = c.GetString("session_manager", "freeswitch_pass")
	if err != nil {
	}
	sm_json, err = c.GetBool("session_manager", "json")
	if err != nil {
	}
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
}

func main() {
	readConfig()
	resolveStandaloneConfilcts()
}
