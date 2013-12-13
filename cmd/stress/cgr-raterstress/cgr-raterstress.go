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
	"log"
	"net/rpc"
	"os"
	"github.com/cgrates/cgrates/engine"
	//"net/rpc/jsonrpc"
	"net/rpc/jsonrpc"
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	balancer   = flag.String("balancer", "localhost:2012", "balancer server address")
	runs       = flag.Int("runs", 10000, "stress cycle number")
	parallel   = flag.Int("parallel", 0, "run n requests in parallel")
	json       = flag.Bool("json", false, "use JSON for RPC encoding")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	runtime.GOMAXPROCS(runtime.NumCPU())

	cd := engine.CallDescriptor{
		TimeStart:    time.Date(2013, time.December, 13, 22, 30, 0, 0, time.UTC),
		TimeEnd:      time.Date(2013, time.December, 13, 22, 31, 0, 0, time.UTC),
		CallDuration: 60 * time.Second,
		Direction:    "*out",
		TOR:          "call",
		Tenant:       "cgrates.org",
		Subject:      "1001",
		Destination:  "+49",
	}
	result := engine.CallCost{}
	var client *rpc.Client
	var err error
	if *json {
		client, err = jsonrpc.Dial("tcp", *balancer)
	} else {
		client, err = rpc.Dial("tcp", *balancer)
	}
	if err != nil {
		log.Fatal("Could not connect to engine: ", err)
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
