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

import "github.com/cgrates/cgrates/utils"

func (dS *DispatcherService) ActionSv1Ping(args *utils.CGREvent, rpl *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ActionSv1Ping, args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaActions, utils.ActionSv1Ping, args, rpl)
}

func (dS *DispatcherService) ActionSv1ScheduleActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) (err error) {
	if args == nil {
		args = &utils.ArgActionSv1ScheduleActions{
			CGREvent: &utils.CGREvent{},
		}
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ActionSv1ScheduleActions, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaActions, utils.ActionSv1ScheduleActions, args, rpl)
}

func (dS *DispatcherService) ActionSv1ExecuteActions(args *utils.ArgActionSv1ScheduleActions, rpl *string) (err error) {
	if args == nil {
		args = &utils.ArgActionSv1ScheduleActions{
			CGREvent: &utils.CGREvent{},
		}
	}
	tnt := dS.cfg.GeneralCfg().DefaultTenant
	if args.CGREvent != nil && args.CGREvent.Tenant != utils.EmptyString {
		tnt = args.CGREvent.Tenant
	}
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.ActionSv1Ping, tnt,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaActions, utils.ActionSv1Ping, args, rpl)
}
