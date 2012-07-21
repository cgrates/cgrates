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
	"flag"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
)

func mainc() {
	flag.Parse()
	sm := &sessionmanager.FSSessionManager{}
	getter, err := timespans.NewRedisStorage(redis_server, redis_db)
	defer getter.Close()
	if err != nil {
		log.Fatalf("Cannot open storage: %v", err)
	}
	if sm_standalone {
		sm.Connect(sessionmanager.NewDirectSessionDelegate(getter), sm_freeswitch_server, sm_freeswitch_pass)
	} else {
		var client *rpc.Client
		if sm_json {
			client, err = jsonrpc.Dial("tcp", sm_api_server)
		} else {
			client, err = rpc.Dial("tcp", sm_api_server)
		}
		if err != nil {
			log.Fatalf("could not connect to balancer: %v", err)
		}
		sm.Connect(sessionmanager.NewRPCClientSessionDelegate(client), sm_freeswitch_server, sm_freeswitch_pass)
	}
	waitChan := make(<-chan byte)
	log.Print("CGRateS is listening!")
	<-waitChan
}
