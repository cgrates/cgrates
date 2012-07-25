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

package timespans

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/balancer"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Responder struct {
	Bal      *balancer.Balancer
	ExitChan chan bool
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.GetCost")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetUserBalanceKey(), func() (*CallCost, error) {
			return arg.GetCost()
		})
		*reply, err = *r, e
	}
	return
}

func (rs *Responder) Debit(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.Debit")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetUserBalanceKey(), func() (*CallCost, error) {
			return arg.Debit()
		})
		*reply, err = *r, e
	}
	return
}

func (rs *Responder) DebitCents(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.DebitCents")
	} else {
		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return arg.DebitCents()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) DebitSMS(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.DebitSMS")
	} else {
		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return arg.DebitSMS()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) DebitSeconds(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.DebitSeconds")
	} else {
		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return 0, arg.DebitSeconds()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) GetMaxSessionTime(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.GetMaxSessionTime")
	} else {
		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return arg.GetMaxSessionTime()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) AddRecievedCallSeconds(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.AddRecievedCallSeconds")
	} else {

		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return 0, arg.AddRecievedCallSeconds()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) Status(arg string, reply *string) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	if rs.Bal != nil {

		*reply = "Connected raters:\n"
		for _, rater := range rs.Bal.GetClientAddresses() {
			log.Print(rater)
			*reply += fmt.Sprintf("%v\n", rater)
		}
		*reply += fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	} else {
		*reply = fmt.Sprintf("memstats before GC: %dKb footprint: %dKb", memstats.HeapAlloc/1024, memstats.Sys/1024)
	}
	return
}

func (rs *Responder) Shutdown(arg string, reply *string) (err error) {
	if rs.Bal != nil {
		rs.Bal.Shutdown()
	}
	storageGetter.Close()
	defer func() { rs.ExitChan <- true }()
	*reply = "Done!"
	return
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) getCallCost(key *CallDescriptor, method string) (reply *CallCost, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &CallCost{}
			reply, err = AccLock.GuardGetCost(key.GetUserBalanceKey(), func() (*CallCost, error) {
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
func (rs *Responder) callMethod(key *CallDescriptor, method string) (reply float64, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			log.Print("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply, err = AccLock.Guard(key.GetUserBalanceKey(), func() (float64, error) {
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

// Reflection worker type for not standalone balancer
type ResponderWorker struct{}

func (rw *ResponderWorker) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodName := strings.TrimLeft(serviceMethod, "Responder.")
	switch args.(type) {
	case CallDescriptor:
		cd := args.(CallDescriptor)
		switch reply.(type) {
		case *CallCost:
			rep := reply.(*CallCost)
			method := reflect.ValueOf(&cd).MethodByName(methodName)
			ret := method.Call([]reflect.Value{})
			*rep = *(ret[0].Interface().(*CallCost))
		case *float64:
			rep := reply.(*float64)
			method := reflect.ValueOf(&cd).MethodByName(methodName)
			ret := method.Call([]reflect.Value{})
			*rep = *(ret[0].Interface().(*float64))
		}
	case string:
		switch methodName {
		case "Status":
			*(reply.(*string)) = "Local!"
		case "Shutdown":
			*(reply.(*string)) = "Done!"
		}

	}
	return nil
}

func (rw *ResponderWorker) Close() error {
	return nil
}
