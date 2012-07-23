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
	"runtime"
	"time"
)

type Responder struct {
	rpc bool
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (r *Responder) GetCost(arg timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	if r.rpc {
		r, e := GetCallCost(&arg, "Responder.GetCost")
		*reply, err = *r, e
	} else {
		r, e := timespans.AccLock.GuardGetCost(arg.GetUserBalanceKey(), func() (*timespans.CallCost, error) {
			return (&arg).GetCost()
		})
		*reply, err = *r, e
	}
	return
}

func (r *Responder) Debit(arg timespans.CallDescriptor, reply *timespans.CallCost) (err error) {
	if r.rpc {
		r, e := GetCallCost(&arg, "Responder.Debit")
		*reply, err = *r, e
	} else {
		r, e := timespans.AccLock.GuardGetCost(arg.GetUserBalanceKey(), func() (*timespans.CallCost, error) {
			return (&arg).Debit()
		})
		*reply, err = *r, e
	}
	return
}

func (r *Responder) DebitCents(arg timespans.CallDescriptor, reply *float64) (err error) {
	if r.rpc {
		*reply, err = CallMethod(&arg, "Responder.DebitCents")
	} else {
		r, e := timespans.AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return (&arg).DebitCents()
		})
		*reply, err = r, e
	}
	return
}

func (r *Responder) DebitSMS(arg timespans.CallDescriptor, reply *float64) (err error) {
	if r.rpc {
		*reply, err = CallMethod(&arg, "Responder.DebitSMS")
	} else {
		r, e := timespans.AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return (&arg).DebitSMS()
		})
		*reply, err = r, e
	}
	return
}

func (r *Responder) DebitSeconds(arg timespans.CallDescriptor, reply *float64) (err error) {
	if r.rpc {
		*reply, err = CallMethod(&arg, "Responder.DebitSeconds")
	} else {
		r, e := timespans.AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return 0, (&arg).DebitSeconds()
		})
		*reply, err = r, e
	}
	return
}

func (r *Responder) GetMaxSessionTime(arg timespans.CallDescriptor, reply *float64) (err error) {
	if r.rpc {
		*reply, err = CallMethod(&arg, "Responder.GetMaxSessionTime")
	} else {
		r, e := timespans.AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return (&arg).GetMaxSessionTime()
		})
		*reply, err = r, e
	}
	return
}

func (r *Responder) AddRecievedCallSeconds(arg timespans.CallDescriptor, reply *float64) (err error) {
	if r.rpc {
		*reply, err = CallMethod(&arg, "Responder.AddRecievedCallSeconds")
	} else {

		r, e := timespans.AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return 0, (&arg).AddRecievedCallSeconds()
		})
		*reply, err = r, e
	}
	return
}

func (r *Responder) Status(arg timespans.CallDescriptor, reply *string) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	if r.rpc {

		*reply = "Connected raters:\n"
		for _, rater := range bal.GetClientAddresses() {
			log.Print(rater)
			*reply += fmt.Sprintf("%v\n", rater)
		}
		*reply += fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	} else {
		*reply = fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	}
	return
}

func (r *Responder) Shutdown(arg string, reply *string) (err error) {
	bal.Shutdown()
	getter.Close()
	defer func() { exitChan <- true }()
	*reply = "Done!"
	return
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
			reply, err = accLock.GuardGetCost(key.GetUserBalanceKey(), func() (*timespans.CallCost, error) {
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
			reply, err = accLock.Guard(key.GetUserBalanceKey(), func() (float64, error) {
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
