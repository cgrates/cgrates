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
	"fmt"
	"github.com/rif/cgrates/sessionmanager"
	"github.com/rif/cgrates/timespans"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"runtime"
)

var (
	balancer   = flag.String("balancer", "127.0.0.1:2000", "balancer address host:port")
	listen     = flag.String("listen", "127.0.0.1:1234", "listening address host:port")
	standalone = flag.Bool("standalone", false, "start standalone server (no balancer), and use json for rpc encoding")
	storage    Responder
)

type Responder struct {
	sg timespans.StorageGetter
}

func NewStorage(nsg timespans.StorageGetter) *Responder {
	return &Responder{sg: nsg}
}

/*
RPC method providing the rating information from the storage.
*/
func (s *Responder) GetCost(cd timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	r, e := descriptor.GetCost()
	*reply, err = *r, e
	return err
}

func (s *Responder) DebitCents(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	r, e := descriptor.DebitCents()
	*reply, err = r, e
	return err
}

func (s *Responder) DebitSMS(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	r, e := descriptor.DebitSMS()
	*reply, err = r, e
	return err
}

func (s *Responder) DebitSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	e := descriptor.DebitSeconds()
	*reply, err = 0.0, e
	return err
}

func (s *Responder) GetMaxSessionTime(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	r, e := descriptor.GetMaxSessionTime()
	*reply, err = r, e
	return err
}

func (s *Responder) AddVolumeDiscountSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	e := descriptor.AddVolumeDiscountSeconds()
	*reply, err = 0, e
	return err
}

func (s *Responder) ResetVolumeDiscountSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	e := descriptor.ResetVolumeDiscountSeconds()
	*reply, err = 0, e
	return err
}

func (s *Responder) AddRecievedCallSeconds(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	e := descriptor.AddRecievedCallSeconds()
	*reply, err = 0, e
	return err
}

func (s *Responder) ResetUserBudget(cd timespans.CallDescriptor, reply *float64) (err error) {
	descriptor := &cd
	descriptor.SetStorageGetter(s.sg)
	e := descriptor.ResetUserBudget()
	*reply, err = 0, e
	return err
}

func (r *Responder) Status(arg timespans.CallDescriptor, replay *string) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	*replay = fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	return
}

/*
RPC method that triggers rater shutdown in case of balancer exit.
*/
func (s *Responder) Shutdown(args string, reply *string) (err error) {
	s.sg.Close()
	defer os.Exit(0)
	*reply = "Done!"
	return nil
}

func main() {
	flag.Parse()
	//getter, err := timespans.NewKyotoStorage("storage.kch")
	getter, err := timespans.NewRedisStorage("tcp:127.0.0.1:6379", 10)
	defer getter.Close()
	if err != nil {
		log.Printf("Cannot open storage: %v", err)
		os.Exit(1)
	}
	if standalone {
		sm = &sessionmanager.SessionManager{}
		sm.Connect("localhost:8021", "ClueCon")
		sm.SetSessionDelegate(new(DirectSessionDelegate))
		sm.StartEventLoop()
	} else {
		go RegisterToServer(balancer, listen)
		go StopSingnalHandler(balancer, listen, getter)
	}
	/*rpc.Register(NewStorage(getter))
	rpc.HandleHTTP()
	addr, err1 := net.ResolveTCPAddr("tcp", *listen)
	l, err2 := net.ListenTCP("tcp", addr)
	if err1 != nil || err2 != nil {
		log.Print("cannot create listener for specified address ", *listen)
		os.Exit(1)
	}
	rpc.Accept(l)*/

	log.Print("Starting Server...")
	l, err := net.Listen("tcp", *listen)
	defer l.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("listening on: ", l.Addr())
	rpc.Register(NewStorage(getter))
	for {
		log.Print("waiting for connections ...")
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %s", conn)
			continue
		}
		log.Printf("connection started: %v", conn.RemoteAddr())
		if *standalone {
			log.Print("json encoding")
			go jsonrpc.ServeConn(conn)
		} else {
			log.Print("gob encoding")
			go rpc.ServeConn(conn)
		}

	}
}
