// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
