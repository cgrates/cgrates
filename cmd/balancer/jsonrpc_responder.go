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
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type Responder byte

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) GetCost(arg timespans.CallDescriptor, replay *timespans.CallCost) (err error) {
	*replay = *GetCost(&arg)
	return
}

func (r *Responder) DebitBalance(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.DebitCents")
	return
}

func (r *Responder) DebitSMS(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.DebitSMS")
	return
}

func (r *Responder) DebitSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.DebitSeconds")
	return
}

func (r *Responder) GetMaxSessionTime(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.GetMaxSessionTime")
	return
}

func (r *Responder) AddVolumeDiscountSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.AddVolumeDiscountSeconds")
	return
}

func (r *Responder) ResetVolumeDiscountSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.ResetVolumeDiscountSeconds")
	return
}

func (r *Responder) AddRecievedCallSeconds(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.AddRecievedCallSeconds")
	return
}

func (r *Responder) ResetUserBudget(arg timespans.CallDescriptor, replay *float64) (err error) {
	*replay = CallMethod(&arg, "Responder.ResetUserBudget")
	return
}

/*
Creates the json rpc server.
*/
func listenToJsonRPCRequests() {
	l, err := net.Listen("tcp", *jsonRpcAddress)
	defer l.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Print("Listening for incomming json RPC requests on ", l.Addr())

	responder := new(Responder)
	rpc.Register(responder)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %s", c)
			continue
		}

		log.Printf("connection started: %v", c.RemoteAddr())
		go jsonrpc.ServeConn(c)
	}
}
