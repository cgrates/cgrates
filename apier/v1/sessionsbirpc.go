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
		utils.SessionSv1GetCost:                   ssv1.BiRPCv1GetCost,

		utils.SessionSv1ForceDisconnect:            ssv1.BiRPCv1ForceDisconnect,
		utils.SessionSv1RegisterInternalBiJSONConn: ssv1.BiRPCv1RegisterInternalBiJSONConn,
		utils.SessionSv1Ping:                       ssv1.BiRPCPing,

		utils.SessionSv1ReplicateSessions:  ssv1.BiRPCv1ReplicateSessions,
		utils.SessionSv1SetPassiveSession:  ssv1.BiRPCv1SetPassiveSession,
		utils.SessionSv1ActivateSessions:   ssv1.BiRPCv1ActivateSessions,
		utils.SessionSv1DeactivateSessions: ssv1.BiRPCv1DeactivateSessions,

		utils.SessionSv1ReAuthorize:    ssv1.BiRPCV1ReAuthorize,
		utils.SessionSv1DisconnectPeer: ssv1.BiRPCV1DisconnectPeer,

		utils.SessionSv1STIRAuthenticate: ssv1.BiRPCV1STIRAuthenticate,
		utils.SessionSv1STIRIdentity:     ssv1.BiRPCV1STIRIdentity,

		utils.SessionSv1Sleep: ssv1.BiRPCV1Sleep, // Sleep method is used to test the concurrent requests mechanism
	}
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEvent(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1AuthorizeEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEventWithDigest(clnt *rpc2.Client, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1AuthorizeEventWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSession(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1InitiateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSessionWithDigest(clnt *rpc2.Client, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1InitiateSessionWithDigest(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1UpdateSession(clnt *rpc2.Client, args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1UpdateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1SyncSessions(clnt *rpc2.Client, args *utils.TenantWithAPIOpts,
	rply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1SyncSessions(clnt, &utils.TenantWithAPIOpts{}, rply)
}

func (ssv1 *SessionSv1) BiRPCv1TerminateSession(clnt *rpc2.Client, args *sessions.V1TerminateSessionArgs,
	rply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1TerminateSession(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessCDR(clnt *rpc2.Client, cgrEv *utils.CGREvent,
	rply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ProcessCDR(clnt, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessMessage(clnt *rpc2.Client, args *sessions.V1ProcessMessageArgs,
	rply *sessions.V1ProcessMessageReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ProcessMessage(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessEvent(clnt *rpc2.Client, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ProcessEvent(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetCost(clnt *rpc2.Client, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1GetCostReply) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1GetCost(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessions(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1GetActiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessionsCount(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *int) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1GetActiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessions(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1GetPassiveSessions(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessionsCount(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *int) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1GetPassiveSessionsCount(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ForceDisconnect(clnt *rpc2.Client, args *utils.SessionFilter,
	rply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ForceDisconnect(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1RegisterInternalBiJSONConn(clnt *rpc2.Client, args string,
	rply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1RegisterInternalBiJSONConn(clnt, args, rply)
}

func (ssv1 *SessionSv1) BiRPCPing(clnt *rpc2.Client, ign *utils.CGREvent,
	reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.Ping(ign, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ReplicateSessions(clnt *rpc2.Client,
	args sessions.ArgsReplicateSessions, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.BiRPCv1ReplicateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1SetPassiveSession(clnt *rpc2.Client,
	args *sessions.Session, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1SetPassiveSession(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ActivateSessions(clnt *rpc2.Client,
	args *utils.SessionIDsWithArgsDispatcher, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ActivateSessions(clnt, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1DeactivateSessions(clnt *rpc2.Client,
	args *utils.SessionIDsWithArgsDispatcher, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1DeactivateSessions(clnt, args, reply)
}

// BiRPCV1ReAuthorize sends the RAR for filterd sessions
func (ssv1 *SessionSv1) BiRPCV1ReAuthorize(clnt *rpc2.Client,
	args *utils.SessionFilter, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1ReAuthorize(clnt, args, reply)
}

// BiRPCV1DisconnectPeer sends the DPR for the OriginHost and OriginRealm
func (ssv1 *SessionSv1) BiRPCV1DisconnectPeer(clnt *rpc2.Client,
	args *utils.DPRArgs, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1DisconnectPeer(clnt, args, reply)
}

// BiRPCV1STIRAuthenticate checks the identity using STIR/SHAKEN
func (ssv1 *SessionSv1) BiRPCV1STIRAuthenticate(clnt *rpc2.Client,
	args *sessions.V1STIRAuthenticateArgs, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1STIRAuthenticate(clnt, args, reply)
}

// BiRPCV1STIRIdentity creates the identity for STIR/SHAKEN
func (ssv1 *SessionSv1) BiRPCV1STIRIdentity(clnt *rpc2.Client,
	args *sessions.V1STIRIdentityArgs, reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	return ssv1.sS.BiRPCv1STIRIdentity(nil, args, reply)
}

func (ssv1 *SessionSv1) BiRPCV1Sleep(clnt *rpc2.Client, arg *utils.DurationArgs,
	reply *string) (err error) {
	if ssv1.caps.IsLimited() {
		if err = ssv1.caps.Allocate(); err != nil {
			return
		}
		defer ssv1.caps.Deallocate()
	}
	time.Sleep(arg.Duration)
	*reply = utils.OK
	return nil
}
