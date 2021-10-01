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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SessionSv1Ping(args *utils.CGREvent, reply *string) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1Ping, args, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEvent(args *utils.CGREvent,
	reply *sessions.V1AuthorizeReply) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1AuthorizeEvent, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1AuthorizeEvent, args, reply)
}

func (dS *DispatcherService) SessionSv1AuthorizeEventWithDigest(args *utils.CGREvent,
	reply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1AuthorizeEventWithDigest, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1AuthorizeEventWithDigest, args, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSession(args *utils.CGREvent,
	reply *sessions.V1InitSessionReply) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1InitiateSession, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1InitiateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSessionWithDigest(args *utils.CGREvent,
	reply *sessions.V1InitReplyWithDigest) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1InitiateSessionWithDigest, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1InitiateSessionWithDigest, args, reply)
}

func (dS *DispatcherService) SessionSv1UpdateSession(args *utils.CGREvent,
	reply *sessions.V1UpdateSessionReply) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1UpdateSession, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1UpdateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1SyncSessions(args *utils.TenantWithAPIOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1SyncSessions, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1SyncSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1TerminateSession(args *utils.CGREvent,
	reply *string) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1TerminateSession, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1TerminateSession, args, reply)
}

func (dS *DispatcherService) SessionSv1ProcessCDR(args *utils.CGREvent,
	reply *string) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ProcessCDR, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1ProcessCDR, args, reply)
}

func (dS *DispatcherService) SessionSv1ProcessMessage(args *utils.CGREvent,
	reply *sessions.V1ProcessMessageReply) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ProcessMessage, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1ProcessMessage, args, reply)
}

func (dS *DispatcherService) SessionSv1ProcessEvent(args *utils.CGREvent,
	reply *sessions.V1ProcessEventReply) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ProcessEvent, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), args, utils.MetaSessionS, utils.SessionSv1ProcessEvent, args, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1GetActiveSessions,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1GetActiveSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1GetActiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1GetActiveSessionsCount,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1GetActiveSessionsCount, args, reply)
}

func (dS *DispatcherService) SessionSv1ForceDisconnect(args *utils.SessionFilter,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ForceDisconnect,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1ForceDisconnect, args, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessions(args *utils.SessionFilter,
	reply *[]*sessions.ExternalSession) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1GetPassiveSessions,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1GetPassiveSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1GetPassiveSessionsCount(args *utils.SessionFilter,
	reply *int) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1GetPassiveSessionsCount,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1GetPassiveSessionsCount, args, reply)
}

func (dS *DispatcherService) SessionSv1ReplicateSessions(args ArgsReplicateSessionsWithAPIOpts,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ReplicateSessions, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1ReplicateSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1SetPassiveSession(args *sessions.Session,
	reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1SetPassiveSession, tnt,
			utils.IfaceAsString(args.OptsStart[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.OptsStart,
	}, utils.MetaSessionS, utils.SessionSv1SetPassiveSession, args, reply)
}

func (dS *DispatcherService) SessionSv1ActivateSessions(args *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1ActivateSessions,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1ActivateSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1DeactivateSessions(args *utils.SessionIDsWithAPIOpts, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1DeactivateSessions,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1DeactivateSessions, args, reply)
}

func (dS *DispatcherService) SessionSv1STIRAuthenticate(args *sessions.V1STIRAuthenticateArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1STIRAuthenticate,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1STIRAuthenticate, args, reply)
}

func (dS *DispatcherService) SessionSv1STIRIdentity(args *sessions.V1STIRIdentityArgs, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.SessionSv1STIRIdentity,
			tnt, utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(context.TODO(), &utils.CGREvent{
		Tenant:  tnt,
		APIOpts: args.APIOpts,
	}, utils.MetaSessionS, utils.SessionSv1STIRIdentity, args, reply)
}
