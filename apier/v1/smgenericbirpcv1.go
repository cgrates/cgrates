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
	"github.com/cenkalti/rpc2"
	"github.com/cgrates/cgrates/sessionmanager"
)

func NewSMGenericBiRpcV1(smRpc *SMGenericV1, sm *sessionmanager.GenericSessionManager) *SMGenericBiRpcV1 {
	return &SMGenericBiRpcV1{smRpc: smRpc, sm: sm}
}

type SMGenericBiRpcV1 struct {
	smRpc *SMGenericV1
	sm    *sessionmanager.GenericSessionManager
}

// Publishes methods exported by SMGenericBiRpcV1 as SMGenericV1 (so we can handle standard RPC methods via birpc socket)
func (self *SMGenericBiRpcV1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		"SMGenericV1.GetMaxUsage":   self.GetMaxUsage,
		"SMGenericV1.SessionStart":  self.SessionStart,
		"SMGenericV1.SessionUpdate": self.SessionUpdate,
		"SMGenericV1.SessionEnd":    self.SessionEnd,
		"SMGenericV1.ProcessCdr":    self.ProcessCdr,
	}
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (self *SMGenericBiRpcV1) GetMaxUsage(client *rpc2.Client, ev sessionmanager.GenericEvent, maxUsage *float64) error {
	return self.smRpc.GetMaxUsage(ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (self *SMGenericBiRpcV1) SessionStart(client *rpc2.Client, ev sessionmanager.GenericEvent, maxUsage *float64) error {
	return self.smRpc.SessionStart(ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (self *SMGenericBiRpcV1) SessionUpdate(client *rpc2.Client, ev sessionmanager.GenericEvent, maxUsage *float64) error {
	return self.smRpc.SessionUpdate(ev, maxUsage)
}

// Called on session end, should stop debit loop
func (self *SMGenericBiRpcV1) SessionEnd(client *rpc2.Client, ev sessionmanager.GenericEvent, reply *string) error {
	return self.smRpc.SessionEnd(ev, reply)
}

// Called on session end, should send the CDR to CDRS
func (self *SMGenericBiRpcV1) ProcessCdr(client *rpc2.Client, ev sessionmanager.GenericEvent, reply *string) error {
	return self.smRpc.ProcessCdr(ev, reply)
}
