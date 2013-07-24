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

package engine

import (
	"errors"
	"fmt"
	"github.com/cgrates/cgrates/balancer2go"
	"net/rpc"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Responder struct {
	Bal      *balancer2go.Balancer
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

func (rs *Responder) MaxDebit(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.MaxDebit")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetUserBalanceKey(), func() (*CallCost, error) {
			return arg.MaxDebit()
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

func (rs *Responder) FlushCache(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.FlushCache")
	} else {
		r, e := AccLock.Guard(arg.GetUserBalanceKey(), func() (float64, error) {
			return 0, arg.FlushCache()
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
		rs.Bal.Shutdown("Responder.Shutdown")
	}
	storageGetter.Close()
	defer func() { rs.ExitChan <- true }()
	*reply = "Done!"
	return
}

func (rs *Responder) GetMonetary(arg CallDescriptor, reply *CallCost) (err error) {
	err = rs.getBalance(&arg, CREDIT, reply)
	return err
}

func (rs *Responder) GetSMS(arg CallDescriptor, reply *CallCost) (err error) {
	err = rs.getBalance(&arg, SMS, reply)
	return err
}

func (rs *Responder) GetInternet(arg CallDescriptor, reply *CallCost) (err error) {
	err = rs.getBalance(&arg, TRAFFIC, reply)
	return err
}

func (rs *Responder) GetInternetTime(arg CallDescriptor, reply *CallCost) (err error) {
	err = rs.getBalance(&arg, TRAFFIC_TIME, reply)
	return err
}

func (rs *Responder) GetMinutes(arg CallDescriptor, reply *CallCost) (err error) {
	err = rs.getBalance(&arg, MINUTES, reply)
	return err
}

// Get balance
func (rs *Responder) getBalance(arg *CallDescriptor, balanceId string, reply *CallCost) (err error) {
	if rs.Bal != nil {
		return errors.New("No balancer supported for this command right now")
	}
	ubKey := arg.Direction + ":" + arg.Tenant + ":" + arg.Account
	userBalance, err := storageGetter.GetUserBalance(ubKey)
	if err != nil {
		return err
	}
	if balance, balExists := userBalance.BalanceMap[balanceId+arg.Direction]; !balExists {
		// No match, balanceId not found
		return errors.New("-BALANCE_NOT_FOUND")
	} else {
		reply.Tenant = arg.Tenant
		reply.Account = arg.Account
		reply.Direction = arg.Direction
		reply.Cost = balance.GetTotalValue()
	}
	return nil
}

/*
The function that gets the information from the raters using balancer.
*/
func (rs *Responder) getCallCost(key *CallDescriptor, method string) (reply *CallCost, err error) {
	err = errors.New("") //not nil value
	for err != nil {
		client := rs.Bal.Balance()
		if client == nil {
			Logger.Info("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &CallCost{}
			reply, err = AccLock.GuardGetCost(key.GetUserBalanceKey(), func() (*CallCost, error) {
				err = client.Call(method, *key, reply)
				return reply, err
			})
			if err != nil {
				Logger.Err(fmt.Sprintf("Got en error from rater: %v", err))
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
			Logger.Info("Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply, err = AccLock.Guard(key.GetUserBalanceKey(), func() (float64, error) {
				err = client.Call(method, *key, &reply)
				return reply, err
			})
			if err != nil {
				Logger.Info(fmt.Sprintf("Got en error from rater: %v", err))
			}
		}
	}
	return
}

/*
RPC method that receives a rater address, connects to it and ads the pair to the rater list for balancing
*/
func (rs *Responder) RegisterRater(clientAddress string, replay *int) error {
	Logger.Info(fmt.Sprintf("Started rater %v registration...", clientAddress))
	time.Sleep(2 * time.Second) // wait a second for Rater to start serving
	client, err := rpc.Dial("tcp", clientAddress)
	if err != nil {
		Logger.Err("Could not connect to client!")
		return err
	}
	rs.Bal.AddClient(clientAddress, client)
	Logger.Info(fmt.Sprintf("Rater %v registered succesfully.", clientAddress))
	return nil
}

/*
RPC method that recives a rater addres gets the connections and closes it and removes the pair from rater list.
*/
func (rs *Responder) UnRegisterRater(clientAddress string, replay *int) error {
	client, ok := rs.Bal.GetClient(clientAddress)
	if ok {
		client.Close()
		rs.Bal.RemoveClient(clientAddress)
		Logger.Info(fmt.Sprintf("Rater %v unregistered succesfully.", clientAddress))
	} else {
		Logger.Info(fmt.Sprintf("Server %v was not on my watch!", clientAddress))
	}
	return nil
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

type Connector interface {
	GetCost(CallDescriptor, *CallCost) error
	Debit(CallDescriptor, *CallCost) error
	MaxDebit(CallDescriptor, *CallCost) error
	DebitCents(CallDescriptor, *float64) error
	DebitSeconds(CallDescriptor, *float64) error
	GetMaxSessionTime(CallDescriptor, *float64) error
}

type RPCClientConnector struct {
	Client *rpc.Client
}

func (rcc *RPCClientConnector) GetCost(cd CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.GetCost", cd, cc)
}

func (rcc *RPCClientConnector) Debit(cd CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.Debit", cd, cc)
}

func (rcc *RPCClientConnector) MaxDebit(cd CallDescriptor, cc *CallCost) error {
	return rcc.Client.Call("Responder.MaxDebit", cd, cc)
}
func (rcc *RPCClientConnector) DebitCents(cd CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.DebitCents", cd, resp)
}
func (rcc *RPCClientConnector) DebitSeconds(cd CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.DebitSeconds", cd, resp)
}
func (rcc *RPCClientConnector) GetMaxSessionTime(cd CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.GetMaxSessionTime", cd, resp)
}
