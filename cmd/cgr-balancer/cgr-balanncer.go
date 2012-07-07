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
	"errors"
	"flag"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"github.com/cgrates/cgrates/balancer"
	"log"
	"runtime"
	"time"
	"sync"
)

var (
	raterAddress    = flag.String("rateraddr", "127.0.0.1:2000", "Rater server address (localhost:2000)")
	jsonRpcAddress  = flag.String("jsonrpcaddr", "127.0.0.1:2001", "Json RPC server address (localhost:2001)")
	httpApiAddress  = flag.String("httpapiaddr", "127.0.0.1:8000", "Http API server address (localhost:2002)")
	freeswitchsrv   = flag.String("freeswitchsrv", "localhost:8021", "freeswitch address host:port")
	freeswitchpass  = flag.String("freeswitchpass", "ClueCon", "freeswitch address host:port")
	bal             *balancer.Balancer
	balancerRWMutex sync.RWMutex
)

/*
The function that gets the information from the raters using balancer.
*/
func GetCallCost(key *timespans.CallDescriptor, method string) (reply *timespans.CallCost) {
	// balancerRWMutex.RLock()
	// defer balancerRWMutex.RUnlock()
	err := errors.New("") //not nil value
	for err != nil {
		client := bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &timespans.CallCost{}
			err = client.Call(method, *key, reply)
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}
	}
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func CallMethod(key *timespans.CallDescriptor, method string) (reply float64) {
	// balancerRWMutex.Lock()
	// defer balancerRWMutex.Unlock()
	err := errors.New("") //not nil value
	for err != nil {
		client := bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			err = client.Call(method, *key, &reply)
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}
	}
	return
}

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	bal = balancer.NewBalancer()

	go StopSingnalHandler()
	go listenToRPCRaterRequests()
	go listenToJsonRPCRequests()

	sm := &sessionmanager.FSSessionManager{}
	sm.Connect(sessionmanager.NewRPCBalancerSessionDelegate(bal), *freeswitchsrv, *freeswitchpass)

	listenToHttpRequests()
}
