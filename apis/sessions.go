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

package apis

import (
	"github.com/cgrates/birpc/context"
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
	ping
	sS *sessions.SessionS
}

func (ssv1 *SessionSv1) AuthorizeEvent(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1AuthorizeReply) error {
	return ssv1.sS.BiRPCv1AuthorizeEvent(ctx, args, rply)
}

func (ssv1 *SessionSv1) AuthorizeEventWithDigest(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1AuthorizeReplyWithDigest) error {
	return ssv1.sS.BiRPCv1AuthorizeEventWithDigest(ctx, args, rply)
}

func (ssv1 *SessionSv1) InitiateSession(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1InitSessionReply) error {
	return ssv1.sS.BiRPCv1InitiateSession(ctx, args, rply)
}

func (ssv1 *SessionSv1) InitiateSessionWithDigest(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1InitReplyWithDigest) error {
	return ssv1.sS.BiRPCv1InitiateSessionWithDigest(ctx, args, rply)
}

func (ssv1 *SessionSv1) UpdateSession(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1UpdateSessionReply) error {
	return ssv1.sS.BiRPCv1UpdateSession(ctx, args, rply)
}

func (ssv1 *SessionSv1) SyncSessions(ctx *context.Context, args *utils.TenantWithAPIOpts,
	rply *string) error {
	return ssv1.sS.BiRPCv1SyncSessions(ctx, &utils.TenantWithAPIOpts{}, rply)
}

func (ssv1 *SessionSv1) TerminateSession(ctx *context.Context, args *utils.CGREvent,
	rply *string) error {
	return ssv1.sS.BiRPCv1TerminateSession(ctx, args, rply)
}

func (ssv1 *SessionSv1) ProcessCDR(ctx *context.Context, cgrEv *utils.CGREvent, rply *string) error {
	return ssv1.sS.BiRPCv1ProcessCDR(ctx, cgrEv, rply)
}

func (ssv1 *SessionSv1) ProcessMessage(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1ProcessMessageReply) error {
	return ssv1.sS.BiRPCv1ProcessMessage(ctx, args, rply)
}

func (ssv1 *SessionSv1) ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	rply *sessions.V1ProcessEventReply) error {
	return ssv1.sS.BiRPCv1ProcessEvent(ctx, args, rply)
}

func (ssv1 *SessionSv1) GetActiveSessions(ctx *context.Context, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.sS.BiRPCv1GetActiveSessions(ctx, args, rply)
}

func (ssv1 *SessionSv1) GetActiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	rply *int) error {
	return ssv1.sS.BiRPCv1GetActiveSessionsCount(ctx, args, rply)
}

func (ssv1 *SessionSv1) ForceDisconnect(ctx *context.Context, args *utils.SessionFilter,
	rply *string) error {
	return ssv1.sS.BiRPCv1ForceDisconnect(ctx, args, rply)
}

func (ssv1 *SessionSv1) GetPassiveSessions(ctx *context.Context, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) error {
	return ssv1.sS.BiRPCv1GetPassiveSessions(ctx, args, rply)
}

func (ssv1 *SessionSv1) GetPassiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	rply *int) error {
	return ssv1.sS.BiRPCv1GetPassiveSessionsCount(ctx, args, rply)
}

func (ssv1 *SessionSv1) ReplicateSessions(ctx *context.Context, args *dispatchers.ArgsReplicateSessionsWithAPIOpts, rply *string) error {
	return ssv1.sS.BiRPCv1ReplicateSessions(ctx, args.ArgsReplicateSessions, rply)
}

func (ssv1 *SessionSv1) SetPassiveSession(ctx *context.Context, args *sessions.Session,
	reply *string) error {
	return ssv1.sS.BiRPCv1SetPassiveSession(ctx, args, reply)
}

// ActivateSessions is called to activate a list/all sessions
func (ssv1 *SessionSv1) ActivateSessions(ctx *context.Context, args *utils.SessionIDsWithAPIOpts, reply *string) error {
	return ssv1.sS.BiRPCv1ActivateSessions(ctx, args, reply)
}

// DeactivateSessions is called to deactivate a list/all active sessios
func (ssv1 *SessionSv1) DeactivateSessions(ctx *context.Context, args *utils.SessionIDsWithAPIOpts, reply *string) error {
	return ssv1.sS.BiRPCv1DeactivateSessions(ctx, args, reply)
}

// ReAuthorize sends the RAR for filterd sessions
func (ssv1 *SessionSv1) ReAuthorize(ctx *context.Context, args *utils.SessionFilter, reply *string) error {
	return ssv1.sS.BiRPCv1ReAuthorize(ctx, args, reply)
}

// DisconnectPeer sends the DPR for the OriginHost and OriginRealm
func (ssv1 *SessionSv1) DisconnectPeer(ctx *context.Context, args *utils.DPRArgs, reply *string) error {
	return ssv1.sS.BiRPCv1DisconnectPeer(ctx, args, reply)
}

// STIRAuthenticate checks the identity using STIR/SHAKEN
func (ssv1 *SessionSv1) STIRAuthenticate(ctx *context.Context, args *sessions.V1STIRAuthenticateArgs, reply *string) error {
	return ssv1.sS.BiRPCv1STIRAuthenticate(ctx, args, reply)
}

// STIRIdentity creates the identity for STIR/SHAKEN
func (ssv1 *SessionSv1) STIRIdentity(ctx *context.Context, args *sessions.V1STIRIdentityArgs, reply *string) error {
	return ssv1.sS.BiRPCv1STIRIdentity(ctx, args, reply)
}

func (ssv1 *SessionSv1) RegisterInternalBiJSONConn(ctx *context.Context, args string, rply *string) (err error) {
	return ssv1.sS.BiRPCv1RegisterInternalBiJSONConn(ctx, args, rply)
}
