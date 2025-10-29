/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package v1

import (
	"github.com/cgrates/birpc"
	"github.com/cgrates/cgrates/utils"
)

// Publishes methods exported by SMGenericV1 as SMGenericV1 (so we can handle standard RPC methods via birpc socket)
func (smgv1 *SMGenericV1) Handlers() map[string]any {
	return map[string]any{
		"SMGenericV1.GetMaxUsage":      smgv1.BiRPCV1GetMaxUsage,
		"SMGenericV1.InitiateSession":  smgv1.BiRPCV1InitiateSession,
		"SMGenericV1.UpdateSession":    smgv1.BiRPCV1UpdateSession,
		"SMGenericV1.TerminateSession": smgv1.BiRPCV1TerminateSession,
		"SMGenericV1.ProcessCDR":       smgv1.BiRPCV1ProcessCDR,
	}
}

// / Returns MaxUsage (for calls in seconds), -1 for no limit
func (smgv1 *SMGenericV1) BiRPCV1GetMaxUsage(clnt birpc.ClientConnector,
	ev map[string]any, maxUsage *float64) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return smgv1.Ss.BiRPCV1GetMaxUsage(clnt, ev, maxUsage)
}

// Called on session start, returns the maximum number of seconds the session can last
func (smgv1 *SMGenericV1) BiRPCV1InitiateSession(clnt birpc.ClientConnector,
	ev map[string]any, maxUsage *float64) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return smgv1.Ss.BiRPCV1InitiateSession(clnt, ev, maxUsage)
}

// Interim updates, returns remaining duration from the rater
func (smgv1 *SMGenericV1) BiRPCV1UpdateSession(clnt birpc.ClientConnector,
	ev map[string]any, maxUsage *float64) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return smgv1.Ss.BiRPCV1UpdateSession(clnt, ev, maxUsage)
}

// Called on session end, should stop debit loop
func (smgv1 *SMGenericV1) BiRPCV1TerminateSession(clnt birpc.ClientConnector,
	ev map[string]any, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return smgv1.Ss.BiRPCV1TerminateSession(clnt, ev, reply)
}

// Called on session end, should send the CDR to CDRS
func (smgv1 *SMGenericV1) BiRPCV1ProcessCDR(clnt birpc.ClientConnector,
	ev map[string]any, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return smgv1.Ss.BiRPCV1ProcessCDR(clnt, ev, reply)
}
