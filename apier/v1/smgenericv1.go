/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

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
package v1

import (
	"reflect"
	"strings"

	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

func NewSMGenericV1(sm *sessionmanager.SMGeneric) *SMGenericV1 {
	return &SMGenericV1{sm: sm}
}

// Exports RPC from SMGeneric
type SMGenericV1 struct {
	sm *sessionmanager.SMGeneric
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericV1) MaxUsage(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	return self.sm.BiRPCV1MaxUsage(nil, ev, maxUsage)
}

// Returns list of suppliers which can be used for the request
func (self *SMGenericV1) LCRSuppliers(ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	return self.sm.BiRPCV1LCRSuppliers(nil, ev, suppliers)
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericV1) InitiateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	return self.sm.BiRPCV1InitiateSession(nil, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericV1) UpdateSession(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	return self.sm.BiRPCV1UpdateSession(nil, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (self *SMGenericV1) TerminateSession(ev sessionmanager.SMGenericEvent, reply *string) error {
	return self.sm.BiRPCV1TerminateSession(nil, ev, reply)
}

// Called on individual Events (eg SMS)
func (self *SMGenericV1) ChargeEvent(ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	return self.sm.BiRPCV1ChargeEvent(nil, ev, maxUsage)
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericV1) ProcessCDR(ev sessionmanager.SMGenericEvent, reply *string) error {
	return self.sm.BiRPCV1ProcessCDR(nil, ev, reply)
}

func (self *SMGenericV1) ActiveSessions(attrs map[string]string, reply *[]*sessionmanager.ActiveSession) error {
	return self.sm.BiRPCV1ActiveSessions(nil, attrs, reply)
}

func (self *SMGenericV1) ActiveSessionsCount(attrs map[string]string, reply *int) error {
	return self.sm.BiRPCV1ActiveSessionsCount(nil, attrs, reply)
}

func (self *SMGenericV1) PassiveSessions(attrs map[string]string, reply *[]*sessionmanager.ActiveSession) error {
	return self.sm.BiRPCV1PassiveSessions(nil, attrs, reply)
}

func (self *SMGenericV1) PassiveSessionsCount(attrs map[string]string, reply *int) error {
	return self.sm.BiRPCV1PassiveSessionsCount(nil, attrs, reply)
}

func (self *SMGenericV1) SetPassiveSessions(args sessionmanager.ArgsSetPassiveSessions, reply *string) error {
	return self.sm.BiRPCV1SetPassiveSessions(nil, args, reply)
}

// rpcclient.RpcClientConnection interface
func (self *SMGenericV1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	methodSplit := strings.Split(serviceMethod, ".")
	if len(methodSplit) != 2 {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	method := reflect.ValueOf(self).MethodByName(methodSplit[1])
	if !method.IsValid() {
		return rpcclient.ErrUnsupporteServiceMethod
	}
	params := []reflect.Value{reflect.ValueOf(args), reflect.ValueOf(reply)}
	ret := method.Call(params)
	if len(ret) != 1 {
		return utils.ErrServerError
	}
	if ret[0].Interface() == nil {
		return nil
	}
	err, ok := ret[0].Interface().(error)
	if !ok {
		return utils.ErrServerError
	}
	return err
}
