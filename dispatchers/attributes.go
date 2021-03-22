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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// AttributeSv1Ping interrogates AttributeS server responsible to process the event
func (dS *DispatcherService) AttributeSv1Ping(args *utils.CGREvent,
	reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.Tenant != utils.EmptyString {
		tnt = args.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaAttributes, utils.AttributeSv1Ping, args, reply)
}

// AttributeSv1GetAttributeForEvent is the dispatcher method for AttributeSv1.GetAttributeForEvent
func (dS *DispatcherService) AttributeSv1GetAttributeForEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1GetAttributeForEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaAttributes, utils.AttributeSv1GetAttributeForEvent, args, reply)
}

// AttributeSv1ProcessEvent .
func (dS *DispatcherService) AttributeSv1ProcessEvent(args *engine.AttrArgsProcessEvent,
	reply *engine.AttrSProcessEventReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.AttributeSv1ProcessEvent, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}

	}
	return dS.Dispatch(args.CGREvent, utils.MetaAttributes, utils.AttributeSv1ProcessEvent, args, reply)
}
