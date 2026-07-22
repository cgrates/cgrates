// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package dispatchers

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func (dS *DispatcherService) RouteSv1Ping(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	if args == nil {
		args = new(utils.CGREvent)
	}
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1Ping,
			args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1Ping, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *engine.SortedRoutesList) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRoutes,
			args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1GetRoutes, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRoutesList,
			args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1GetRoutesList, args, reply)
}

func (dS *DispatcherService) RouteSv1GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*engine.RouteProfile) (err error) {
	args.Tenant = utils.FirstNonEmpty(args.Tenant, dS.cfg.GeneralCfg().DefaultTenant)
	if len(dS.cfg.DispatcherSCfg().AttributeSConns) != 0 {
		if err = dS.authorize(utils.RouteSv1GetRouteProfilesForEvent,
			args.Tenant,
			utils.IfaceAsString(args.APIOpts[utils.OptsAPIKey]), args.Time); err != nil {
			return
		}
	}
	return dS.Dispatch(args, utils.MetaRoutes, utils.RouteSv1GetRouteProfilesForEvent, args, reply)
}
