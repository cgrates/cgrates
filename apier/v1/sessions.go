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

func NewSessionSv1(sm *sessions.SMGeneric) *SessionSv1 {
	return &SessionSv1{SMG: sm}
}

// SessionSv1 exports RPC from SessionSv1
type SessionSv1 struct {
	SMG *sessions.SMGeneric
}

// Publishes BiJSONRPC methods exported by SessionSv1
func (ssv1 *SessionSv1) Handlers() map[string]interface{} {
	return map[string]interface{}{
		utils.SessionSv1AuthorizeEvent:             ssv1.BiRpcAuthorizeEvent,
		utils.SessionSv1AuthorizeEventWithDigest:   ssv1.BiRpcAuthorizeEventWithDigest,
		utils.SessionSv1InitiateSession:            ssv1.BiRpcInitiateSession,
		utils.SessionSv1InitiateSessionWithDigest:  ssv1.BiRpcInitiateSessionWithDigest,
		utils.SessionSv1UpdateSession:              ssv1.BiRpcUpdateSession,
		utils.SessionSv1SyncSessions:               ssv1.BiRpcSyncSessions,
		utils.SessionSv1TerminateSession:           ssv1.BiRpcTerminateSession,
		utils.SessionSv1ProcessCDR:                 ssv1.BiRpcProcessCDR,
		utils.SessionSv1ProcessEvent:               ssv1.BiRpcProcessEvent,
		utils.SessionSv1GetActiveSessions:          ssv1.BiRPCV1GetActiveSessions,
		utils.SessionSv1ForceDisconnect:            ssv1.BiRPCV1ForceDisconnect,
		utils.SessionSv1GetPassiveSessions:         ssv1.BiRPCV1GetPassiveSessions,
		utils.SessionSv1RegisterInternalBiJSONConn: ssv1.BiRPCv1RegisterInternalBiJSONConn,
		utils.SessionSv1Ping:                       ssv1.BiRPCPing,
	}
}

func (ssv1 *SessionSv1) AuthorizeEvent(args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) error {
	return ssv1.SMG.BiRPCv1AuthorizeEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) error {
	return ssv1.SMG.BiRPCv1AuthorizeEventWithDigest(nil, args, rply)
}

func (ssv1 *SessionSv1) InitiateSession(args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) error {
	return ssv1.SMG.BiRPCv1InitiateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) InitiateSessionWithDigest(args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) error {
	return ssv1.SMG.BiRPCv1InitiateSessionWithDigest(nil, args, rply)
}

func (ssv1 *SessionSv1) UpdateSession(args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) error {
	return ssv1.SMG.BiRPCv1UpdateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) SyncSessions(args *string,
	rply *string) error {
	return ssv1.SMG.BiRPCv1SyncSessions(nil, "", rply)
}

func (ssv1 *SessionSv1) TerminateSession(args *sessions.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.SMG.BiRPCv1TerminateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) ProcessCDR(cgrEv *utils.CGREvent, rply *string) error {
	return ssv1.SMG.BiRPCv1ProcessCDR(nil, cgrEv, rply)
}

func (ssv1 *SessionSv1) ProcessEvent(args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) error {
	return ssv1.SMG.BiRPCv1ProcessEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) GetActiveSessions(args map[string]string, rply *[]*sessions.ActiveSession) error {
	return ssv1.SMG.BiRPCV1GetActiveSessions(nil, args, rply)
}

func (ssv1 *SessionSv1) ForceDisconnect(args map[string]string, rply *string) error {
	return ssv1.SMG.BiRPCV1ForceDisconnect(nil, args, rply)
}

func (ssv1 *SessionSv1) GetPassiveSessions(args map[string]string, rply *[]*sessions.ActiveSession) error {
	return ssv1.SMG.BiRPCV1GetPassiveSessions(nil, args, rply)
}

func (ssv1 *SessionSv1) BiRpcAuthorizeEvent(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) error {
	return ssv1.SMG.BiRPCv1AuthorizeEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcAuthorizeEventWithDigest(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) error {
	return ssv1.SMG.BiRPCv1AuthorizeEventWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcInitiateSession(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) error {
	return ssv1.SMG.BiRPCv1InitiateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcInitiateSessionWithDigest(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) error {
	return ssv1.SMG.BiRPCv1InitiateSessionWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcUpdateSession(clnt *rpc2.Client, args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) error {
	return ssv1.SMG.BiRPCv1UpdateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcSyncSessions(clnt *rpc2.Client, args *string,
	rply *string) error {
	return ssv1.SMG.BiRPCv1SyncSessions(clnt, "", rply)
}

func (ssv1 *SessionSv1) BiRpcTerminateSession(clnt *rpc2.Client, args *sessions.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.SMG.BiRPCv1TerminateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRpcProcessCDR(clnt *rpc2.Client, cgrEv *utils.CGREvent, rply *string) error {
	return ssv1.SMG.BiRPCv1ProcessCDR(clnt, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRpcProcessEvent(clnt *rpc2.Client, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) error {
	return ssv1.SMG.BiRPCv1ProcessEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCV1GetActiveSessions(clnt *rpc2.Client, args map[string]string,
	rply *[]*sessions.ActiveSession) error {
	return ssv1.SMG.BiRPCV1GetActiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCV1ForceDisconnect(clnt *rpc2.Client, args map[string]string,
	rply *string) error {
	return ssv1.SMG.BiRPCV1ForceDisconnect(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCV1GetPassiveSessions(clnt *rpc2.Client, args map[string]string,
	rply *[]*sessions.ActiveSession) error {
	return ssv1.SMG.BiRPCV1GetPassiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1RegisterInternalBiJSONConn(clnt *rpc2.Client, args string,
	rply *string) error {
	return ssv1.SMG.BiRPCv1RegisterInternalBiJSONConn(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCPing(clnt *rpc2.Client, ign string, reply *string) error {
	return ssv1.Ping(ign, reply)
}

func (ssv1 *SessionSv1) Ping(ign string, reply *string) error {
	*reply = utils.Pong
	return nil
}
