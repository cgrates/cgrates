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
	"github.com/cenkalti/rpc2"
)

// Publishes methods exported by SMGenericV1 as SMGenericV1 (so we can handle standard RPC methods via birpc socket)
func (smgv1 *SMGenericV1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		"SMGenericV1.GetMaxUsage":      smgv1.BiRPCV1GetMaxUsage,
		"SMGenericV1.InitiateSession":  smgv1.BiRPCV1InitiateSession,
		"SMGenericV1.UpdateSession":    smgv1.BiRPCV1UpdateSession,
		"SMGenericV1.TerminateSession": smgv1.BiRPCV1TerminateSession,
		"SMGenericV1.ProcessCDR":       smgv1.BiRPCV1ProcessCDR,
	}
}

/// Returns MaxUsage (for calls in seconds), -1 for no limit
func (smgv1 *SMGenericV1) BiRPCV1GetMaxUsage(clnt *rpc2.Client,
	ev map[string]interface{}, maxUsage *float64) (err error) {
	return smgv1.Ss.BiRPCV1GetMaxUsage(clnt, ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (smgv1 *SMGenericV1) BiRPCV1InitiateSession(clnt *rpc2.Client,
	ev map[string]interface{}, maxUsage *float64) (err error) {
	return smgv1.Ss.BiRPCV1InitiateSession(clnt, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (smgv1 *SMGenericV1) BiRPCV1UpdateSession(clnt *rpc2.Client,
	ev map[string]interface{}, maxUsage *float64) (err error) {
	return smgv1.Ss.BiRPCV1UpdateSession(clnt, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (smgv1 *SMGenericV1) BiRPCV1TerminateSession(clnt *rpc2.Client,
	ev map[string]interface{}, reply *string) (err error) {
	return smgv1.Ss.BiRPCV1TerminateSession(clnt, ev, reply)
}

// Called on session end, should send the CDR to CDRS
func (smgv1 *SMGenericV1) BiRPCV1ProcessCDR(clnt *rpc2.Client,
	ev map[string]interface{}, reply *string) (err error) {
	return smgv1.Ss.BiRPCV1ProcessCDR(clnt, ev, reply)
}
