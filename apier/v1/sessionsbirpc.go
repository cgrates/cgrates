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

	"github.com/cgrates/birpc"
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

		utils.SessionSv1Sleep: ssv1.BiRPCV1Sleep, // Sleep method is used to test the concurrent requests mechanism
	}
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEvent(clnt birpc.ClientConnector, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1AuthorizeEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEventWithDigest(clnt birpc.ClientConnector, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1AuthorizeEventWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSession(clnt birpc.ClientConnector, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1InitiateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSessionWithDigest(clnt birpc.ClientConnector, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1InitiateSessionWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1UpdateSession(clnt birpc.ClientConnector, args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1UpdateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1SyncSessions(clnt birpc.ClientConnector, args *string,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1SyncSessions(clnt, "", rply)
}

func (ssv1 *SessionSv1) BiRPCv1TerminateSession(clnt birpc.ClientConnector, args *sessions.V1TerminateSessionArgs,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1TerminateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessCDR(clnt birpc.ClientConnector, cgrEv *utils.CGREventWithArgDispatcher,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessCDR(clnt, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessMessage(clnt birpc.ClientConnector, args *sessions.V1ProcessMessageArgs,
	rply *sessions.V1ProcessMessageReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessMessage(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessEvent(clnt birpc.ClientConnector, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessions(clnt birpc.ClientConnector, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetActiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessionsCount(clnt birpc.ClientConnector, args *utils.SessionFilter,
	rply *int) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetActiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessions(clnt birpc.ClientConnector, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetPassiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessionsCount(clnt birpc.ClientConnector, args *utils.SessionFilter,
	rply *int) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetPassiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ForceDisconnect(clnt birpc.ClientConnector, args *utils.SessionFilter,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ForceDisconnect(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1RegisterInternalBiJSONConn(clnt birpc.ClientConnector, args string,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1RegisterInternalBiJSONConn(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCPing(clnt birpc.ClientConnector, ign *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ping(ign, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ReplicateSessions(clnt birpc.ClientConnector,
	args sessions.ArgsReplicateSessions, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.BiRPCv1ReplicateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1SetPassiveSession(clnt birpc.ClientConnector,
	args *sessions.Session, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1SetPassiveSession(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ActivateSessions(clnt birpc.ClientConnector,
	args []string, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ActivateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1DeactivateSessions(clnt birpc.ClientConnector,
	args []string, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1DeactivateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCV1Sleep(clnt birpc.ClientConnector, arg *DurationArgs,
	reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	time.Sleep(arg.DurationTime)
	*reply = utils.OK
	return nil
}
