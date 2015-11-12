/*
Rating system designed to be used in VoIP Carriers World
Copyright (C) 2012-2015 ITsysCOM

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
	"time"

	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/sessionmanager"
	"github.com/cgrates/cgrates/utils"
)

func NewSMGenericBiRpcV1(sm *sessionmanager.SMGeneric) *SMGenericBiRpcV1 {
	return &SMGenericBiRpcV1{sm: sm}
}

type SMGenericBiRpcV1 struct {
	sm *sessionmanager.SMGeneric
}

// Publishes methods exported by SMGenericBiRpcV1 as SMGenericV1 (so we can handle standard RPC methods via birpc socket)
func (self *SMGenericBiRpcV1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		"SMGenericV1.GetMaxUsage":     self.GetMaxUsage,
		"SMGenericV1.GetLcrSuppliers": self.GetLcrSuppliers,
		"SMGenericV1.SessionStart":    self.SessionStart,
		"SMGenericV1.SessionUpdate":   self.SessionUpdate,
		"SMGenericV1.SessionEnd":      self.SessionEnd,
		"SMGenericV1.ProcessCdr":      self.ProcessCdr,
	}
}

/// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericBiRpcV1) GetMaxUsage(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.GetMaxUsage(ev, clnt)
	if err != nil {
		return utils.NewErrServerError(err)
	}
	if maxUsageDur == time.Duration(-1) {
		*maxUsage = -1.0
	} else {
		*maxUsage = maxUsageDur.Seconds()
	}
	return nil
}

/// Returns list of suppliers which can be used for the request
func (self *SMGenericBiRpcV1) GetLcrSuppliers(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.GetLcrSuppliers(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericBiRpcV1) SessionStart(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if err := self.sm.SessionStart(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	return self.GetMaxUsage(clnt, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericBiRpcV1) SessionUpdate(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if err := self.sm.SessionUpdate(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	return self.GetMaxUsage(clnt, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (self *SMGenericBiRpcV1) SessionEnd(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.SessionEnd(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericBiRpcV1) ProcessCdr(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCdr(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
