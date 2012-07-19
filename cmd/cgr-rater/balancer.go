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
	"github.com/cgrates/cgrates/balancer"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"runtime"
	"time"
)

var (
	raterAddress   = flag.String("rateraddr", "127.0.0.1:2000", "Rater server address (localhost:2000)")
	rpcAddress     = flag.String("rpcaddr", "127.0.0.1:2001", "Json RPC server address (localhost:2001)")
	httpApiAddress = flag.String("httpapiaddr", "127.0.0.1:8000", "Http API server address (localhost:8000)")
	freeswitch     = flag.Bool("freeswitch", false, "connect to freeswitch server")
	freeswitchsrv  = flag.String("freeswitchsrv", "localhost:8021", "freeswitch address host:port")
	freeswitchpass = flag.String("freeswitchpass", "ClueCon", "freeswitch address host:port")
	js             = flag.Bool("json", false, "use JSON for RPC encoding")
	bal            *balancer.Balancer
	accLock        = timespans.NewAccountLock()
)

/*
The function that gets the information from the raters using balancer.
*/
func GetCallCost(key *timespans.CallDescriptor, method string) (reply *timespans.CallCost, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &timespans.CallCost{}
			reply, err = accLock.GuardGetCost(key.GetKey(), func() (*timespans.CallCost, error) {
				err = client.Call(method, *key, reply)
				return reply, err
			})
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
func CallMethod(key *timespans.CallDescriptor, method string) (reply float64, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply, err = accLock.Guard(key.GetKey(), func() (float64, error) {
				err = client.Call(method, *key, &reply)
				return reply, err
			})
			if err != nil {
				log.Printf("Got en error from rater: %v", err)
			}
		}
	}
	return
}

func maind() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	bal = balancer.NewBalancer()

	go stopSingnalHandler()
	go listenToRPCRaterRequests()
	go listenToRPCRequests()

	if *freeswitch {
		sm := &sessionmanager.FSSessionManager{}
		sm.Connect(sessionmanager.NewRPCBalancerSessionDelegate(bal), *freeswitchsrv, *freeswitchpass)
	}
	listenToHttpRequests()
}
