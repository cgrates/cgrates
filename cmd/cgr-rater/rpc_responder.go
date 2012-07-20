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
	"fmt"
	"github.com/cgrates/cgrates/timespans"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"runtime"
	"time"
)

type RpcResponder struct{}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *RpcResponder) GetCost(arg timespans.CallDescriptor, replay *timespans.CallCost) (err error) {
	rs, err := GetCallCost(&arg, "Responder.GetCost")
	*replay = *rs
	return
}

func (r *RpcResponder) Debit(arg timespans.CallDescriptor, replay *timespans.CallCost) (err error) {
	rs, err := GetCallCost(&arg, "Responder.Debit")
	*replay = *rs
	return
}

func (r *RpcResponder) DebitBalance(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay, err = CallMethod(&arg, "Responder.DebitCents")
	return
}

func (r *RpcResponder) DebitSMS(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay, err = CallMethod(&arg, "Responder.DebitSMS")
	return
}

func (r *RpcResponder) DebitSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay, err = CallMethod(&arg, "Responder.DebitSeconds")
	return
}

func (r *RpcResponder) GetMaxSessionTime(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay, err = CallMethod(&arg, "Responder.GetMaxSessionTime")
	return
}

func (r *RpcResponder) AddVolumeDiscountSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay, err = CallMethod(&arg, "Responder.AddVolumeDiscountSeconds")
	return
}

func (r *RpcResponder) Status(arg timespans.CallDescriptor, replay *string) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	*replay = "Connected raters:\n"
	for _, rater := range bal.GetClientAddresses() {
		log.Print(rater)
		*replay += fmt.Sprintf("%v\n", rater)
	}
	*replay += fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	return
}

/*
Creates the json rpc server.
*/
func listenToRPCRequests() {
	l, err := net.Listen("tcp", *rpcAddress)
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening for incomming json RPC requests on ", l.Addr())

	responder := new(Responder)
	rpc.Register(responder)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %s", conn)
			continue
		}

		log.Printf("connection started: %v", conn.RemoteAddr())
		if *js {
			// log.Print("json encoding")
			go jsonrpc.ServeConn(conn)
		} else {
			// log.Print("gob encoding")
			go rpc.ServeConn(conn)
		}
	}
}

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
