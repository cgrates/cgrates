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

package dispatchers

import (
	"time"

	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SessionSv1Ping(args *CGREvWithApiKey, reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1Ping, args.CGREvent, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEvent(args *AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReply) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1AuthorizeEvent,
			args.V1AuthorizeArgs.CGREvent.Tenant,
			args.APIKey, args.V1AuthorizeArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1AuthorizeArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1AuthorizeEvent, args.V1AuthorizeArgs, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEventWithDigest(args *AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1AuthorizeEventWithDigest,
			args.V1AuthorizeArgs.CGREvent.Tenant,
			args.APIKey, args.V1AuthorizeArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1AuthorizeArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1AuthorizeEventWithDigest, args.V1AuthorizeArgs, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSession(args *InitArgsWithApiKey,
	reply *sessions.V1InitSessionReply) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1InitiateSession,
			args.V1InitSessionArgs.CGREvent.Tenant,
			args.APIKey, args.V1InitSessionArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1InitSessionArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1InitiateSession, args.V1InitSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSessionWithDigest(args *InitArgsWithApiKey,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1InitiateSessionWithDigest,
			args.V1InitSessionArgs.CGREvent.Tenant,
			args.APIKey, args.V1InitSessionArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1InitSessionArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1InitiateSessionWithDigest, args.V1InitSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1UpdateSession(args *UpdateSessionWithApiKey,
	reply *sessions.V1UpdateSessionReply) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1UpdateSession,
			args.V1UpdateSessionArgs.CGREvent.Tenant,
			args.APIKey, args.V1UpdateSessionArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1UpdateSessionArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1UpdateSession, args.V1UpdateSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1SyncSessions(args *TntWithApiKey,
	reply *sessions.V1UpdateSessionReply) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1SyncSessions,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1SyncSessions, &args.TenantArg.Tenant, reply)
}

func (dS *DispatcherService) SessionSv1TerminateSession(args *TerminateSessionWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1TerminateSession,
			args.V1TerminateSessionArgs.CGREvent.Tenant,
			args.APIKey, args.V1TerminateSessionArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.V1TerminateSessionArgs.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1TerminateSession, args.V1TerminateSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1ProcessCDR(args *CGREvWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1ProcessCDR,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1ProcessCDR, args.CGREvent, reply)
}

func (dS *DispatcherService) SessionSv1ProcessEvent(args *ProcessEventWithApiKey,
	reply *sessions.V1ProcessEventReply) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1ProcessEvent,
			args.V1ProcessEventArgs.CGREvent.Tenant,
			args.APIKey, args.V1ProcessEventArgs.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(&args.CGREvent, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1ProcessEvent, args.V1ProcessEventArgs, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessions(args *FilterSessionWithApiKey,
	reply *[]*sessions.ActiveSession) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1GetActiveSessions,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1GetActiveSessions, args.Filters, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessionsCount(args *FilterSessionWithApiKey,
	reply *int) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1GetActiveSessionsCount,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1GetActiveSessionsCount, args.Filters, reply)
}

func (dS *DispatcherService) SessionSv1ForceDisconnect(args *FilterSessionWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1ForceDisconnect,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1ForceDisconnect, args.Filters, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessions(args *FilterSessionWithApiKey,
	reply *[]*sessions.ActiveSession) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1GetPassiveSessions,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1GetPassiveSessions, args.Filters, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessionsCount(args *FilterSessionWithApiKey,
	reply *int) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1GetPassiveSessionsCount,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1GetPassiveSessionsCount, args.Filters, reply)
}

func (dS *DispatcherService) SessionSv1ReplicateSessions(args *ArgsReplicateSessionsWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1ReplicateSessions,
			args.TenantArg.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.TenantArg.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1ReplicateSessions, args.ArgsReplicateSessions, reply)
}

func (dS *DispatcherService) SessionSv1SetPassiveSession(args *SessionWithApiKey,
	reply *string) (err error) {
	if dS.attrS != nil {
		if err = dS.authorize(utils.SessionSv1SetPassiveSession,
			args.Session.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Session.Tenant}, utils.MetaSessionS, args.RouteID,
		utils.SessionSv1SetPassiveSession, args.Session, reply)
}
