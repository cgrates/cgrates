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
	"github.com/rif/cgrates/timespans"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

var (
	balancer = flag.String("balancer", "localhost:2001", "balancer server address")
	runs     = flag.Int("runs", 10000, "stress cycle number")
	parallel = flag.Bool("parallel", false, "run requests in parallel")
)

func main() {
	flag.Parse()
	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := timespans.CallDescriptor{CstmId: "vdf", Subject: "rif", DestinationPrefix: "0256", TimeStart: t1, TimeEnd: t2}
	result := timespans.CallCost{}
	client, err := jsonrpc.Dial("tcp", *balancer)
	if err != nil {
		log.Fatalf("could not connect to balancer: %v", err)
	}
	if *parallel {
		var divCall *rpc.Call
		for i := 0; i < *runs; i++ {
			divCall = client.Go("Responder.GetCost", cd, &result, nil)
		}
		<-divCall.Done
	} else {
		for j := 0; j < *runs; j++ {
			client.Call("Responder.GetCost", cd, &result)
		}
	}
	log.Println(result)
	client.Close()
}
