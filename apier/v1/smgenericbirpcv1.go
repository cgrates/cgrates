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
		"SMGenericV1.MaxUsage":         self.MaxUsage,
		"SMGenericV1.LCRSuppliers":     self.LCRSuppliers,
		"SMGenericV1.InitiateSession":  self.InitiateSession,
		"SMGenericV1.UpdateSession":    self.UpdateSession,
		"SMGenericV1.TerminateSession": self.TerminateSession,
		"SMGenericV1.ProcessCDR":       self.ProcessCDR,
	}
}

/// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericBiRpcV1) MaxUsage(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	maxUsageDur, err := self.sm.MaxUsage(ev, clnt)
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
func (self *SMGenericBiRpcV1) LCRSuppliers(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, suppliers *[]string) error {
	if supls, err := self.sm.LCRSuppliers(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*suppliers = supls
	}
	return nil
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericBiRpcV1) InitiateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.InitiateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericBiRpcV1) UpdateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.UpdateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should stop debit loop
func (self *SMGenericBiRpcV1) TerminateSession(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.TerminateSession(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}

// Called on individual Events (eg SMS)
func (self *SMGenericBiRpcV1) ChargeEvent(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, maxUsage *float64) error {
	if minMaxUsage, err := self.sm.ChargeEvent(ev, clnt); err != nil {
		return utils.NewErrServerError(err)
	} else {
		*maxUsage = minMaxUsage.Seconds()
	}
	return nil
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericBiRpcV1) ProcessCDR(clnt *rpc2.Client, ev sessionmanager.SMGenericEvent, reply *string) error {
	if err := self.sm.ProcessCDR(ev); err != nil {
		return utils.NewErrServerError(err)
	}
	*reply = utils.OK
	return nil
}
