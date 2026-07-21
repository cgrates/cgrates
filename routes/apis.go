// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
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
	attrConns, err := engine.GetConnIDs(ctx, rpS.cfg.RouteSCfg().Conns, utils.MetaAttributes, tnt, args.AsDataProvider(), nil, rpS.fltrS)
	if err != nil {
		return err
	}
	if len(attrConns) != 0 {
		args.APIOpts[utils.MetaSubsys] = utils.MetaRoutes
		var context string
		if context, err = engine.GetStringOpts(ctx, tnt, args.AsDataProvider(), nil, rpS.fltrS, rpS.cfg.RouteSCfg().Opts.Context,
			utils.OptsContext); err != nil {
			return
		}
		args.APIOpts[utils.OptsContext] = context
		var rplyEv attributes.ProcessEventReply
		if err := rpS.connMgr.Call(ctx, attrConns,
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
