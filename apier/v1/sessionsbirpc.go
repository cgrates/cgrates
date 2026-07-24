// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package v1

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEvent(ctx *context.Context, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1AuthorizeEvent(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1AuthorizeEventWithDigest(ctx *context.Context, args *sessions.V1AuthorizeArgs,
	rply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1AuthorizeEventWithDigest(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSession(ctx *context.Context, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitSessionReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1InitiateSession(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1InitiateSessionWithDigest(ctx *context.Context, args *sessions.V1InitSessionArgs,
	rply *sessions.V1InitReplyWithDigest) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1InitiateSessionWithDigest(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1UpdateSession(ctx *context.Context, args *sessions.V1UpdateSessionArgs,
	rply *sessions.V1UpdateSessionReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1UpdateSession(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1SyncSessions(ctx *context.Context, args *string,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1SyncSessions(ctx.Client, "", rply)
}

func (ssv1 *SessionSv1) BiRPCv1TerminateSession(ctx *context.Context, args *sessions.V1TerminateSessionArgs,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1TerminateSession(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessCDR(ctx *context.Context, cgrEv *utils.CGREventWithArgDispatcher,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessCDR(ctx.Client, cgrEv, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessMessage(ctx *context.Context, args *sessions.V1ProcessMessageArgs,
	rply *sessions.V1ProcessMessageReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessMessage(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ProcessEvent(ctx *context.Context, args *sessions.V1ProcessEventArgs,
	rply *sessions.V1ProcessEventReply) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ProcessEvent(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessions(ctx *context.Context, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetActiveSessions(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetActiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	rply *int) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetActiveSessionsCount(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessions(ctx *context.Context, args *utils.SessionFilter,
	rply *[]*sessions.ExternalSession) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetPassiveSessions(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1GetPassiveSessionsCount(ctx *context.Context, args *utils.SessionFilter,
	rply *int) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1GetPassiveSessionsCount(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1ForceDisconnect(ctx *context.Context, args *utils.SessionFilter,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ForceDisconnect(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1RegisterInternalBiJSONConn(ctx *context.Context, args string,
	rply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1RegisterInternalBiJSONConn(ctx.Client, args, rply)
}

func (ssv1 *SessionSv1) BiRPCv1Ping(ctx *context.Context, ign *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ping(ign, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ReplicateSessions(ctx *context.Context,
	args sessions.ArgsReplicateSessions, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ReplicateSessions(ctx.Client, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1SetPassiveSession(ctx *context.Context,
	args *sessions.Session, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1SetPassiveSession(ctx.Client, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1ActivateSessions(ctx *context.Context,
	args []string, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1ActivateSessions(ctx.Client, args, reply)
}

func (ssv1 *SessionSv1) BiRPCv1DeactivateSessions(ctx *context.Context,
	args []string, reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1DeactivateSessions(ctx.Client, args, reply)
}

func (ssv1 *SessionSv1) BiRPCV1Sleep(ctx *context.Context, args *utils.DurationArgs,
	reply *string) (err error) {
	if err = utils.ConReqs.Allocate(); err != nil {
		return
	}
	defer utils.ConReqs.Deallocate()
	return ssv1.Ss.BiRPCv1Sleep(context.TODO(), args, reply)
}
