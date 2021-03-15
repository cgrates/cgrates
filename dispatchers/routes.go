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

func (dS *DispatcherService) RouteSv1Ping(args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1Ping,
			args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1Ping, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRoutes(args *engine.ArgsGetRoutes, reply *engine.SortedRoutesList) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRoutes,
			args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaRoutes, utils.RouteSv1GetRoutes, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRoutesList(args *engine.ArgsGetRoutes, reply *[]string) (err error) {
	args.CGREvent.Tenant = utils.FirstNonEmpty(args.CGREvent.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRoutesList,
			args.CGREvent.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.CGREvent.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args.CGREvent, utils.MetaRoutes, utils.RouteSv1GetRoutesList, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRouteProfilesForEvent(args *utils.CGREvent, reply *[]*engine.RouteProfile) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRouteProfilesForEvent,
			args.Tenant,
			utils.IfaceAsString(args.Opts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1GetRouteProfilesForEvent, args, reply)
}
