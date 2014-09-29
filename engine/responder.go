/*
Real-time Charging System for Telecom & ISP environments
Copyright (C) 2012-2014 ITsysCOM GmbH

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
	"net/rpc"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/cgrates/cgrates/balancer2go"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

type Responder struct {
	Bal      *balancer2go.Balancer
	ExitChan chan bool
	CdrSrv   *CDRS
}

/*
RPC method thet provides the external RPC interface for getting the rating information.
*/
func (rs *Responder) GetCost(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.GetCost")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetAccountKey(), func() (*CallCost, error) {
			return arg.GetCost()
		})
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) Debit(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.Debit")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetAccountKey(), func() (*CallCost, error) {
			return arg.Debit()
		})
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) MaxDebit(arg CallDescriptor, reply *CallCost) (err error) {
	if rs.Bal != nil {
		r, e := rs.getCallCost(&arg, "Responder.MaxDebit")
		*reply, err = *r, e
	} else {
		r, e := AccLock.GuardGetCost(arg.GetAccountKey(), func() (*CallCost, error) {
			return arg.MaxDebit()
		})
		if e != nil {
			return e
		} else if r != nil {
			*reply = *r
		}
	}
	return
}

func (rs *Responder) RefundIncrements(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.RefundIncrements")
	} else {
		r, e := AccLock.Guard(arg.GetAccountKey(), func() (float64, error) {
			return arg.RefundIncrements()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) GetMaxSessionTime(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.GetMaxSessionTime")
	} else {
		r, e := AccLock.Guard(arg.GetAccountKey(), func() (float64, error) {
			d, err := arg.GetMaxSessionDuration()
			return float64(d), err
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) GetDerivedChargers(attrs utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	// ToDo: Make it work with balancer if needed

	if dcsH, err := HandleGetDerivedChargers(accountingStorage, config.CgrConfig(), attrs); err != nil {
		return err
	} else if dcsH != nil {
		*dcs = dcsH
	}
	return nil
}

func (rs *Responder) ProcessCdr(cdr *utils.StoredCdr, reply *string) error {
	if rs.CdrSrv == nil {
		return errors.New("CdrServerNotRunning")
	}
	if err := rs.CdrSrv.ProcessCdr(cdr); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

func (rs *Responder) FlushCache(arg CallDescriptor, reply *float64) (err error) {
	if rs.Bal != nil {
		*reply, err = rs.callMethod(&arg, "Responder.FlushCache")
	} else {
		r, e := AccLock.Guard(arg.GetAccountKey(), func() (float64, error) {
			return 0, arg.FlushCache()
		})
		*reply, err = r, e
	}
	return
}

func (rs *Responder) Status(arg string, reply *map[string]interface{}) (err error) {
	memstats := new(runtime.MemStats)
	runtime.ReadMemStats(memstats)
	response := make(map[string]interface{})
	if rs.Bal != nil {
		response["Raters"] = rs.Bal.GetClientAddresses()
	}
	response["memstat"] = memstats.HeapAlloc / 1024
	response["footprint"] = memstats.Sys / 1024
	*reply = response
	return
}

func (rs *Responder) Shutdown(arg string, reply *string) (err error) {
	if rs.Bal != nil {
		rs.Bal.Shutdown("Responder.Shutdown")
	}
	dataStorage.(Storage).Close()
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
			Logger.Info("<Balancer> Waiting for raters to register...")
			time.Sleep(1 * time.Second) // wait one second and retry
		} else {
			reply = &CallCost{}
			reply, err = AccLock.GuardGetCost(key.GetAccountKey(), func() (*CallCost, error) {
				err = client.Call(method, *key, reply)
				return reply, err
			})
			if err != nil {
				Logger.Err(fmt.Sprintf("<Balancer> Got en error from rater: %v", err))
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
			reply, err = AccLock.Guard(key.GetAccountKey(), func() (float64, error) {
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
	RefundIncrements(CallDescriptor, *float64) error
	GetMaxSessionTime(CallDescriptor, *float64) error
	GetDerivedChargers(utils.AttrDerivedChargers, *utils.DerivedChargers) error
	ProcessCdr(*utils.StoredCdr, *string) error
}

type RPCClientConnector struct {
	Client *rpcclient.RpcClient
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

func (rcc *RPCClientConnector) RefundIncrements(cd CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.RefundIncrements", cd, resp)
}

func (rcc *RPCClientConnector) GetMaxSessionTime(cd CallDescriptor, resp *float64) error {
	return rcc.Client.Call("Responder.GetMaxSessionTime", cd, resp)
}

func (rcc *RPCClientConnector) GetDerivedChargers(attrs utils.AttrDerivedChargers, dcs *utils.DerivedChargers) error {
	return rcc.Client.Call("ApierV1.GetDerivedChargers", attrs, dcs)
}

func (rcc *RPCClientConnector) ProcessCdr(cdr *utils.StoredCdr, reply *string) error {
	return rcc.Client.Call("CDRSV1.ProcessCdr", cdr, reply)
}
