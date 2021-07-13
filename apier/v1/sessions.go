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
	"github.com/cgrates/cgrates/dispatchers"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func NewSessionSv1(sS *sessions.SessionS) *SessionSv1 {
	return &SessionSv1{
		sS: sS,
	}
}

// SessionSv1 exports RPC from SessionSv1
type SessionSv1 struct {
	sS *sessions.SessionS
}

func (ssv1 *SessionSv1) AuthorizeEvent(args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) error {
	return ssv1.sS.BiRPCv1AuthorizeEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) error {
	return ssv1.sS.BiRPCv1AuthorizeEventWithDigest(nil, args, rply)
}

func (ssv1 *SessionSv1) InitiateSession(args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) error {
	return ssv1.sS.BiRPCv1InitiateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) InitiateSessionWithDigest(args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) error {
	return ssv1.sS.BiRPCv1InitiateSessionWithDigest(nil, args, rply)
}

func (ssv1 *SessionSv1) UpdateSession(args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) error {
	return ssv1.sS.BiRPCv1UpdateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) SyncSessions(args *utils.TenantWithAPIOpts,
	rply *string) error {
	return ssv1.sS.BiRPCv1SyncSessions(nil, &utils.TenantWithAPIOpts{}, rply)
}

func (ssv1 *SessionSv1) TerminateSession(args *sessions.V1TerminateSessionArgs,
	rply *string) error {
	return ssv1.sS.BiRPCv1TerminateSession(nil, args, rply)
}

func (ssv1 *SessionSv1) ProcessCDR(cgrEv *utils.CGREvent, rply *string) error {
	return ssv1.sS.BiRPCv1ProcessCDR(nil, cgrEv, rply)
}

func (ssv1 *SessionSv1) ProcessMessage(args *sessions.V1ProcessMessageArgs,
	rply *sessions.V1ProcessMessageReply) error {
	return ssv1.sS.BiRPCv1ProcessMessage(nil, args, rply)
}

func (ssv1 *SessionSv1) ProcessEvent(args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) error {
	return ssv1.sS.BiRPCv1ProcessEvent(nil, args, rply)
}

func (ssv1 *SessionSv1) GetCost(args *sessions.V1ProcessEventArgs,
	rply *sessions.V1GetCostReply) error {
	return ssv1.sS.BiRPCv1GetCost(nil, args, rply)
}

func (ssv1 *SessionSv1) GetActiveSessions(args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.sS.BiRPCv1GetActiveSessions(nil, args, rply)
}

func (ssv1 *SessionSv1) GetActiveSessionsCount(args *utils.SessionFilter,
	rply *int) error {
	return ssv1.sS.BiRPCv1GetActiveSessionsCount(nil, args, rply)
}

func (ssv1 *SessionSv1) ForceDisconnect(args *utils.SessionFilter,
	rply *string) error {
	return ssv1.sS.BiRPCv1ForceDisconnect(nil, args, rply)
}

func (ssv1 *SessionSv1) GetPassiveSessions(args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.sS.BiRPCv1GetPassiveSessions(nil, args, rply)
}

func (ssv1 *SessionSv1) GetPassiveSessionsCount(args *utils.SessionFilter,
	rply *int) error {
	return ssv1.sS.BiRPCv1GetPassiveSessionsCount(nil, args, rply)
}

func (ssv1 *SessionSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}

func (ssv1 *SessionSv1) ReplicateSessions(args *dispatchers.ArgsReplicateSessionsWithAPIOpts, rply *string) error {
	return ssv1.sS.BiRPCv1ReplicateSessions(nil, args.ArgsReplicateSessions, rply)
}

func (ssv1 *SessionSv1) SetPassiveSession(args *sessions.Session,
	reply *string) error {
	return ssv1.sS.BiRPCv1SetPassiveSession(nil, args, reply)
}

// ActivateSessions is called to activate a list/all sessions
func (ssv1 *SessionSv1) ActivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return ssv1.sS.BiRPCv1ActivateSessions(nil, args, reply)
}

// DeactivateSessions is called to deactivate a list/all active sessios
func (ssv1 *SessionSv1) DeactivateSessions(args *utils.SessionIDsWithArgsDispatcher, reply *string) error {
	return ssv1.sS.BiRPCv1DeactivateSessions(nil, args, reply)
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (ssv1 *SessionSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(ssv1, serviceMethod, args, reply)
}

// ReAuthorize sends the RAR for filterd sessions
func (ssv1 *SessionSv1) ReAuthorize(args *utils.SessionFilter, reply *string) error {
	return ssv1.sS.BiRPCv1ReAuthorize(nil, args, reply)
}

// DisconnectPeer sends the DPR for the OriginHost and OriginRealm
func (ssv1 *SessionSv1) DisconnectPeer(args *utils.DPRArgs, reply *string) error {
	return ssv1.sS.BiRPCv1DisconnectPeer(nil, args, reply)
}

// STIRAuthenticate checks the identity using STIR/SHAKEN
func (ssv1 *SessionSv1) STIRAuthenticate(args *sessions.V1STIRAuthenticateArgs, reply *string) error {
	return ssv1.sS.BiRPCv1STIRAuthenticate(nil, args, reply)
}

// STIRIdentity creates the identity for STIR/SHAKEN
func (ssv1 *SessionSv1) STIRIdentity(args *sessions.V1STIRIdentityArgs, reply *string) error {
	return ssv1.sS.BiRPCv1STIRIdentity(nil, args, reply)
}
