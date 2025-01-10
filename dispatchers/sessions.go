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

// do not modify this code because it's generated
package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SessionSv1ActivateSessions(ctx *context.Context, args *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ActivateSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1AuthorizeEvent(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1AuthorizeReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1AuthorizeEvent, args, reply)
}
func (dS *DispatcherService) SessionSv1AuthorizeEventWithDigest(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1AuthorizeEventWithDigest, args, reply)
}
func (dS *DispatcherService) SessionSv1DeactivateSessions(ctx *context.Context, args *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1DeactivateSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1DisconnectPeer(ctx *context.Context, args *utils.DPRArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	ev := make(map[string]any)
	opts := make(map[string]any)
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1DisconnectPeer, args, reply)
}
func (dS *DispatcherService) SessionSv1ForceDisconnect(ctx *context.Context, args *utils.SessionFilter, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ForceDisconnect, args, reply)
}
func (dS *DispatcherService) SessionSv1GetActiveSessions(ctx *context.Context, args *utils.SessionFilter, reply *[]*sessions.ExternalSession) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1GetActiveSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1GetActiveSessionsCount(ctx *context.Context, args *utils.SessionFilter, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1GetActiveSessionsCount, args, reply)
}
func (dS *DispatcherService) SessionSv1GetPassiveSessions(ctx *context.Context, args *utils.SessionFilter, reply *[]*sessions.ExternalSession) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1GetPassiveSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1GetPassiveSessionsCount(ctx *context.Context, args *utils.SessionFilter, reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1GetPassiveSessionsCount, args, reply)
}
func (dS *DispatcherService) SessionSv1InitiateSession(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1InitSessionReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1InitiateSession, args, reply)
}
func (dS *DispatcherService) SessionSv1InitiateSessionWithDigest(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1InitReplyWithDigest) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1InitiateSessionWithDigest, args, reply)
}
func (dS *DispatcherService) SessionSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1Ping, args, reply)
}
func (dS *DispatcherService) SessionSv1ProcessCDR(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ProcessCDR, args, reply)
}
func (dS *DispatcherService) SessionSv1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1ProcessEventReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ProcessEvent, args, reply)
}
func (dS *DispatcherService) SessionSv1ProcessMessage(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1ProcessMessageReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ProcessMessage, args, reply)
}
func (dS *DispatcherService) SessionSv1ReAuthorize(ctx *context.Context, args *utils.SessionFilter, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ReAuthorize, args, reply)
}
func (dS *DispatcherService) SessionSv1RegisterInternalBiJSONConn(ctx *context.Context, args string, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	ev := make(map[string]any)
	opts := make(map[string]any)
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1RegisterInternalBiJSONConn, args, reply)
}
func (dS *DispatcherService) SessionSv1ReplicateSessions(ctx *context.Context, args sessions.ArgsReplicateSessions, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := args.APIOpts
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1ReplicateSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1STIRAuthenticate(ctx *context.Context, args *sessions.V1STIRAuthenticateArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1STIRAuthenticate, args, reply)
}
func (dS *DispatcherService) SessionSv1STIRIdentity(ctx *context.Context, args *sessions.V1STIRIdentityArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1STIRIdentity, args, reply)
}
func (dS *DispatcherService) SessionSv1SetPassiveSession(ctx *context.Context, args *sessions.Session, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.OriginCGREvent.Tenant) != 0 {
		tnt = args.OriginCGREvent.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1SetPassiveSession, args, reply)
}
func (dS *DispatcherService) SessionSv1SyncSessions(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1SyncSessions, args, reply)
}
func (dS *DispatcherService) SessionSv1TerminateSession(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1TerminateSession, args, reply)
}
func (dS *DispatcherService) SessionSv1UpdateSession(ctx *context.Context, args *utils.CGREvent, reply *sessions.V1UpdateSessionReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]any)
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]any)
	if args != nil {
		opts = args.APIOpts
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaSessionS, utils.SessionSv1UpdateSession, args, reply)
}
