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

package dispatcher

import (
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) SessionSv1Ping(ign string, rpl *string) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	return dS.sessionS.Call(utils.SessionSv1Ping, ign, rpl)
}

func (dS *DispatcherService) SessionSv1AuthorizeEventWithDigest(args *AuthorizeArgsWithApiKey,
	reply *sessions.V1AuthorizeReplyWithDigest) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.V1AuthorizeArgs.CGREvent.Tenant,
		utils.SessionSv1AuthorizeEventWithDigest, args.V1AuthorizeArgs.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1AuthorizeEventWithDigest, args.V1AuthorizeArgs, reply)
}

func (dS *DispatcherService) SessionSv1InitiateSessionWithDigest(args *InitArgsWithApiKey,
	reply *sessions.V1InitSessionReply) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.V1InitSessionArgs.CGREvent.Tenant,
		utils.SessionSv1InitiateSessionWithDigest, args.V1InitSessionArgs.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1InitiateSessionWithDigest, args.V1InitSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1ProcessCDR(args *CGREvWithApiKey,
	reply *string) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.CGREvent.Tenant,
		utils.SessionSv1ProcessCDR, args.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1ProcessCDR, args.CGREvent, reply)
}

func (dS *DispatcherService) SessionSv1ProcessEvent(args *ProcessEventWithApiKey,
	reply *sessions.V1ProcessEventReply) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.V1ProcessEventArgs.CGREvent.Tenant,
		utils.SessionSv1ProcessEvent, args.V1ProcessEventArgs.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1ProcessEvent, args.V1ProcessEventArgs, reply)
}

func (dS *DispatcherService) SessionSv1TerminateSession(args *TerminateSessionWithApiKey,
	reply *string) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.V1TerminateSessionArgs.CGREvent.Tenant,
		utils.SessionSv1TerminateSession, args.V1TerminateSessionArgs.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1TerminateSession, args.V1TerminateSessionArgs, reply)
}

func (dS *DispatcherService) SessionSv1UpdateSession(args *UpdateSessionWithApiKey,
	reply *sessions.V1UpdateSessionReply) (err error) {
	if dS.sessionS == nil {
		return utils.NewErrNotConnected(utils.SessionS)
	}
	if err = dS.authorizeMethod(args.APIKey, args.V1UpdateSessionArgs.CGREvent.Tenant,
		utils.SessionSv1UpdateSession, args.V1UpdateSessionArgs.CGREvent.Time); err != nil {
		return
	}
	return dS.sessionS.Call(utils.SessionSv1UpdateSession, args.V1UpdateSessionArgs, reply)
}
