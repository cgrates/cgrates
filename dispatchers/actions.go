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
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) ActionSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
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
		if err = dS.authorize(ctx, utils.ActionSv1Ping, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaActions, utils.ActionSv1Ping, args, reply)
}
func (dS *DispatcherService) ActionSv1ProcessEvent(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
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
		if err = dS.authorize(ctx, utils.ActionSv1ProcessEvent, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaActions, utils.ActionSv1ProcessEvent, args, reply)
}
func (dS *DispatcherService) ActionSv1ScheduleActions(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
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
		if err = dS.authorize(ctx, utils.ActionSv1ScheduleActions, tnt, utils.IfaceAsString(opts[utils.OptsAPIKey])); err != nil {
			return
		}
	}
	return dS.Dispatch(ctx, &utils.CGREvent{Tenant: tnt, Event: ev, APIOpts: opts}, utils.MetaActions, utils.ActionSv1ScheduleActions, args, reply)
}
