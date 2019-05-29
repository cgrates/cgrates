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

func (dS *DispatcherService) SessionSv1Ping(args *utils.CGREventWithArgDispatcher, reply *string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1Ping,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1Ping, args, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEvent(args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReply) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1AuthorizeEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1AuthorizeEvent, args, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEventWithDigest(args *sessions.V1AuthorizeArgs,
	reply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1AuthorizeEventWithDigest,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1AuthorizeEventWithDigest, args, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSession(args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitSessionReply) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1InitiateSession,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1InitiateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSessionWithDigest(args *sessions.V1InitSessionArgs,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1InitiateSessionWithDigest,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1InitiateSessionWithDigest, args, reply)
}

func (dS *DispatcherService) SessionSv1UpdateSession(args *sessions.V1UpdateSessionArgs,
	reply *sessions.V1UpdateSessionReply) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1UpdateSession,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1UpdateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1SyncSessions(args *utils.TenantWithArgDispatcher,
	reply *sessions.V1UpdateSessionReply) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1SyncSessions, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaSessionS, routeID,
		utils.SessionSv1SyncSessions, &tnt, reply)
}

func (dS *DispatcherService) SessionSv1TerminateSession(args *sessions.V1TerminateSessionArgs,
	reply *string) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1TerminateSession,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1TerminateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1ProcessCDR(args *utils.CGREventWithArgDispatcher,
	reply *string) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1ProcessCDR,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1ProcessCDR, args, reply)
}

func (dS *DispatcherService) SessionSv1ProcessEvent(args *sessions.V1ProcessEventArgs,
	reply *sessions.V1ProcessEventReply) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1ProcessEvent,
			args.CGREvent.Tenant,
			args.APIKey, args.CGREvent.Time); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(args.CGREvent, utils.MetaSessionS, routeID,
		utils.SessionSv1ProcessEvent, args, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ActiveSession) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1GetActiveSessions,
			args.Tenant, args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1GetActiveSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1GetActiveSessionsCount,
			args.Tenant, args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1GetActiveSessionsCount, args, reply)
}

func (dS *DispatcherService) SessionSv1ForceDisconnect(args *utils.SessionFilter,
	reply *string) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1ForceDisconnect,
			args.Tenant, args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1ForceDisconnect, args, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ActiveSession) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1GetPassiveSessions,
			args.Tenant, args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1GetPassiveSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1GetPassiveSessionsCount,
			args.Tenant, args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1GetPassiveSessionsCount, args, reply)
}

func (dS *DispatcherService) SessionSv1ReplicateSessions(args ArgsReplicateSessionsWithApiKey,
	reply *string) (err error) {
	tnt := utils.FirstNonEmpty(args.TenantArg.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1ReplicateSessions, tnt,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: tnt}, utils.MetaSessionS, routeID,
		utils.SessionSv1ReplicateSessions, args.ArgsReplicateSessions, reply)
}

func (dS *DispatcherService) SessionSv1SetPassiveSession(args *sessions.Session,
	reply *string) (err error) {
	if dS.attrS != nil {
		if args.ArgDispatcher == nil {
			return utils.NewErrMandatoryIeMissing(utils.ArgDispatcherField)
		}
		if err = dS.authorize(utils.SessionSv1SetPassiveSession,
			args.Tenant,
			args.APIKey, utils.TimePointer(time.Now())); err != nil {
			return
		}
	}
	var routeID *string
	if args.ArgDispatcher != nil {
		routeID = args.ArgDispatcher.RouteID
	}
	return dS.Dispatch(&utils.CGREvent{Tenant: args.Tenant}, utils.MetaSessionS, routeID,
		utils.SessionSv1SetPassiveSession, args, reply)
}
