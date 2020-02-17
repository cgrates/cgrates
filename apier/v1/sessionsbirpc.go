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
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

// Bidirectional JSON methods following
func (ssv1 *SessionSv1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1GetActiveSessions:       ssv1.BiRPCv1GetActiveSessions,
		utils.SessionSv1GetActiveSessionsCount:  ssv1.BiRPCv1GetActiveSessionsCount,
		utils.SessionSv1GetPassiveSessions:      ssv1.BiRPCv1GetPassiveSessions,
		utils.SessionSv1GetPassiveSessionsCount: ssv1.BiRPCv1GetPassiveSessionsCount,

		utils.SessionSv1AuthorizeEvent:            ssv1.BiRPCv1AuthorizeEvent,
		utils.SessionSv1AuthorizeEventWithDigest:  ssv1.BiRPCv1AuthorizeEventWithDigest,
		utils.SessionSv1InitiateSession:           ssv1.BiRPCv1InitiateSession,
		utils.SessionSv1InitiateSessionWithDigest: ssv1.BiRPCv1InitiateSessionWithDigest,
		utils.SessionSv1UpdateSession:             ssv1.BiRPCv1UpdateSession,
		utils.SessionSv1SyncSessions:              ssv1.BiRPCv1SyncSessions,
		utils.SessionSv1TerminateSession:          ssv1.BiRPCv1TerminateSession,
		utils.SessionSv1ProcessCDR:                ssv1.BiRPCv1ProcessCDR,
		utils.SessionSv1ProcessMessage:            ssv1.BiRPCv1ProcessMessage,
		utils.SessionSv1ProcessEvent:              ssv1.BiRPCv1ProcessEvent,

		utils.SessionSv1ForceDisconnect:            ssv1.BiRPCv1ForceDisconnect,
		utils.SessionSv1RegisterInternalBiJSONConn: ssv1.BiRPCv1RegisterInternalBiJSONConn,
		utils.SessionSv1Ping:                       ssv1.BiRPCPing,

		utils.SessionSv1ReplicateSessions:  ssv1.BiRPCv1ReplicateSessions,
		utils.SessionSv1SetPassiveSession:  ssv1.BiRPCv1SetPassiveSession,
		utils.SessionSv1ActivateSessions:   ssv1.BiRPCv1ActivateSessions,
		utils.SessionSv1DeactivateSessions: ssv1.BiRPCv1DeactivateSessions,
	}
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEvent(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) error {
	return ssv1.Ss.BiRPCv1AuthorizeEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEventWithDigest(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) error {
	return ssv1.Ss.BiRPCv1AuthorizeEventWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSession(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) error {
	return ssv1.Ss.BiRPCv1InitiateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSessionWithDigest(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) error {
	return ssv1.Ss.BiRPCv1InitiateSessionWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1UpdateSession(clnt *rpc2.Client, args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) error {
	return ssv1.Ss.BiRPCv1UpdateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1SyncSessions(clnt *rpc2.Client, args *utils.TenantWithArgDispatcher,
	rply *string) error {
	return ssv1.Ss.BiRPCv1SyncSessions(clnt, &utils.TenantWithArgDispatcher{}, rply)
}

func (ssv1 *SessionSv1) BiRPCv1TerminateSession(clnt *rpc2.Client, args *sessions.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.Ss.BiRPCv1TerminateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessCDR(clnt *rpc2.Client, cgrEv *utils.CGREventWithArgDispatcher, rply *string) error {
	return ssv1.Ss.BiRPCv1ProcessCDR(clnt, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessMessage(clnt *rpc2.Client, args *sessions.V1ProcessMessageArgs,
	rply *sessions.V1ProcessMessageReply) error {
	return ssv1.Ss.BiRPCv1ProcessMessage(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessEvent(clnt *rpc2.Client, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) error {
	return ssv1.Ss.BiRPCv1ProcessEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessions(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.Ss.BiRPCv1GetActiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessionsCount(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *int) error {
	return ssv1.Ss.BiRPCv1GetActiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessions(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.Ss.BiRPCv1GetPassiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessionsCount(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *int) error {
	return ssv1.Ss.BiRPCv1GetPassiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ForceDisconnect(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *string) error {
	return ssv1.Ss.BiRPCv1ForceDisconnect(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1RegisterInternalBiJSONConn(clnt *rpc2.Client, args string,
	rply *string) error {
	return ssv1.Ss.BiRPCv1RegisterInternalBiJSONConn(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCPing(clnt *rpc2.Client, ign *utils.CGREventWithArgDispatcher, reply *string) error {
	return ssv1.Ping(ign, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ReplicateSessions(clnt *rpc2.Client,
	args sessions.ArgsReplicateSessions, reply *string) error {
	return ssv1.BiRPCv1ReplicateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1SetPassiveSession(clnt *rpc2.Client,
	args *sessions.Session, reply *string) error {
	return ssv1.Ss.BiRPCv1SetPassiveSession(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ActivateSessions(clnt *rpc2.Client,
	args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return ssv1.Ss.BiRPCv1ActivateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1DeactivateSessions(clnt *rpc2.Client,
	args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return ssv1.Ss.BiRPCv1DeactivateSessions(clnt, args, reply)
}
