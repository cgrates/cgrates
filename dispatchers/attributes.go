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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) AttributeSv1GetAttributeForEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.APIAttributeProfile) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.AttributeSv1GetAttributeForEvent, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAttributes, utils.AttributeSv1GetAttributeForEvent, args, reply)
}
func (dS *DispatcherService) AttributeSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.AttributeSv1Ping, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAttributes, utils.AttributeSv1Ping, args, reply)
}
func (dS *DispatcherService) AttributeSv1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *engine.AttrSProcessEventReply) (err error) {
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args != nil && len(args.Tenant) != 0 {
		tnt = args.Tenant
	}
	ev := make(map[string]interface{})
	if args != nil {
		ev = args.Event
	}
	opts := make(map[string]interface{})
	if args != nil {
		opts = args.APIOpts
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(ctx, utils.AttributeSv1ProcessEvent, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaAttributes, utils.AttributeSv1ProcessEvent, args, reply)
}
