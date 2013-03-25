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
	"flag"
	"github.com/cgrates/cgrates/rater"
	"log"
	"net/rpc"
	//"net/rpc/jsonrpc"
	"net/rpc/jsonrpc"
	"runtime"
	"time"
)

var (
	balancer = flag.String("balancer", "localhost:2001", "balancer server address")
	runs     = flag.Int("runs", 10000, "stress cycle number")
	parallel = flag.Int("parallel", 0, "run n requests in parallel")
	json     = flag.Bool("json", false, "use JSON for RPC encoding")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	t1 := time.Date(2012, time.February, 02, 17, 30, 0, 0, time.UTC)
	t2 := time.Date(2012, time.February, 02, 18, 30, 0, 0, time.UTC)
	cd := rater.CallDescriptor{Direction: "OUT", TOR: "0", Tenant: "vdf", Subject: "rif", Destination: "0256", TimeStart: t1, TimeEnd: t2}
	result := rater.CallCost{}
	var client *rpc.Client
	var err error
	if *json {
		client, err = jsonrpc.Dial("tcp", *balancer)
	} else {
		client, err = rpc.Dial("tcp", *balancer)
	}
	if err != nil {
		log.Fatal("Could not connect to rater: ", err)
	}
	start := time.Now()
	if *parallel > 0 {
		// var divCall *rpc.Call
		var sem = make(chan int, *parallel)
		var finish = make(chan int)
		for i := 0; i < *runs; i++ {
			go func() {
				sem <- 1
				client.Call("Responder.GetCost", cd, &result)
				<-sem
				finish <- 1
				// divCall = client.Go("Responder.GetCost", cd, &result, nil)	
			}()
		}
		for i := 0; i < *runs; i++ {
			<-finish
		}
		// <-divCall.Done
	} else {
		for j := 0; j < *runs; j++ {
			client.Call("Responder.GetCost", cd, &result)
		}
	}
	duration := time.Since(start)
	log.Println(result)
	client.Close()
	log.Printf("Elapsed: %v resulted: %v req/s.", duration, float64(*runs)/duration.Seconds())
}
