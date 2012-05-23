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
	"github.com/rif/cgrates/sessionmanager"
	"github.com/rif/cgrates/timespans"
	"log"
)

var (
	balancer       = flag.String("balancer", "127.0.0.1:2000", "balancer address host:port")
	freeswitchsrv  = flag.String("freeswitchsrv", "localhost:8021", "freeswitch address host:port")
	freeswitchpass = flag.String("freeswitchpass", "ClueCon", "freeswitch address host:port")
	redissrv       = flag.String("redissrv", "127.0.0.1:6379", "redis address host:port")
	redisdb        = flag.Int("redisdb", 10, "redis database number")
)

func main() {
	flag.Parse()
	sm := &sessionmanager.FSSessionManager{}
	getter, err := timespans.NewRedisStorage(*redissrv, *redisdb)
	defer getter.Close()
	if err != nil {
		log.Fatalf("Cannot open storage: %v", err)
	}
	sm.Connect(sessionmanager.NewDirectSessionDelegate(getter), *freeswitchsrv, *freeswitchpass)
	waitChan := make(<-chan byte)
	log.Print("CGRateS is listening!")
	<-waitChan
}
