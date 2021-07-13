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
	"github.com/cgrates/cgrates/sessions"
)

func NewSMGenericV1(sS *sessions.SessionS) *SMGenericV1 {
	return &SMGenericV1{
		Ss: sS,
	}
}

// Exports RPC from SMGeneric
// DEPRECATED, use SessionSv1 instead
type SMGenericV1 struct {
	Ss *sessions.SessionS
}

// Returns MaxUsage (for calls in seconds), -1 for no limit
func (smgv1 *SMGenericV1) GetMaxUsage(ev map[string]interface{},
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1GetMaxUsage(nil, ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (smgv1 *SMGenericV1) InitiateSession(ev map[string]interface{},
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1InitiateSession(nil, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (smgv1 *SMGenericV1) UpdateSession(ev map[string]interface{},
	maxUsage *float64) error {
	return smgv1.Ss.BiRPCV1UpdateSession(nil, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (smgv1 *SMGenericV1) TerminateSession(ev map[string]interface{},
	reply *string) error {
	return smgv1.Ss.BiRPCV1TerminateSession(nil, ev, reply)
}

// Called on session end, should send the CDR to CDRS
func (smgv1 *SMGenericV1) ProcessCDR(ev map[string]interface{},
	reply *string) error {
	return smgv1.Ss.BiRPCV1ProcessCDR(nil, ev, reply)
}
