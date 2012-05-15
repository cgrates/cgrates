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
	"log"
)

var (
	balancer       = flag.String("balancer", "127.0.0.1:2000", "balancer address host:port")
	freeswitchsrv  = flag.String("freeswitchsrv", "localhost:8021", "freeswitch address host:port")
	freeswitchpass = flag.String("freeswitchpass", "ClueCon", "freeswitch address host:port")
)

func main() {
	flag.Parse()
	log.Print("Mediator!")
}
