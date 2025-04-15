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

package routes

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1GetRoutes returns the list of valid routes.
func (rpS *RouteS) V1GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *SortedRoutesList) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cfg.GeneralCfg().DefaultTenant
	}
	if args.APIOpts == nil {
		args.APIOpts = make(map[string]any)
	}
	if len(rpS.cfg.RouteSCfg().AttributeSConns) != 0 {
		args.APIOpts[utils.MetaSubsys] = utils.MetaRoutes
		var context string
		if context, err = engine.GetStringOpts(ctx, tnt, args.AsDataProvider(), nil, rpS.fltrS, rpS.cfg.RouteSCfg().Opts.Context,
			utils.OptsContext); err != nil {
			return
		}
		args.APIOpts[utils.OptsContext] = context
		var rplyEv engine.AttrSProcessEventReply
		if err := rpS.connMgr.Call(ctx, rpS.cfg.RouteSCfg().AttributeSConns,
			utils.AttributeSv1ProcessEvent, args, &rplyEv); err == nil && len(rplyEv.AlteredFields) != 0 {
			args = rplyEv.CGREvent
			args.APIOpts = rplyEv.CGREvent.APIOpts
		} else if err = utils.CastRPCErr(err); err != utils.ErrNotFound {
			return utils.NewErrRouteS(err)
		}
	}
	var sSps SortedRoutesList
	if sSps, err = rpS.sortedRoutesForEvent(ctx, tnt, args); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*reply = sSps
	return
}

// V1GetRouteProfilesForEvent returns the list of valid route profiles.
func (rpS *RouteS) V1GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*utils.RouteProfile) (_ error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	} else if args.Event == nil {
		return utils.NewErrMandatoryIeMissing(utils.Event)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rpS.cfg.GeneralCfg().DefaultTenant
	}
	sPs, err := rpS.matchingRouteProfilesForEvent(ctx, tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = sPs
	return
}

// V1GetRoutesList returns the list of valid routes.
func (rpS *RouteS) V1GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	sR := new(SortedRoutesList)
	if err = rpS.V1GetRoutes(ctx, args, sR); err != nil {
		return
	}
	*reply = sR.RoutesWithParams()
	return
}
