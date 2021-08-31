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

package apis

import (
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetRouteProfile returns a Route configuration
func (adms *AdminSv1) GetRouteProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.APIRouteProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if rp, err := adms.dm.GetRouteProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *engine.NewAPIRouteProfile(rp)
	}
	return nil
}

// GetRouteProfileIDs returns list of routeProfile IDs registered for a tenant
func (adms *AdminSv1) GetRouteProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, sppPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RouteProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := adms.dm.DataDB().GetKeysForPrefix(ctx, prfx)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[len(prfx):]
	}
	*sppPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetRouteProfileCount sets in reply var the total number of RouteProfileIDs registered for the received tenant
// returns ErrNotFound in case of 0 RouteProfileIDs
func (adms *AdminSv1) GetRouteProfileCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx,
		utils.RouteProfilePrefix+tnt+utils.ConcatenatedKeySep); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return nil
}

//SetRouteProfile add a new Route configuration
func (adms *AdminSv1) SetRouteProfile(ctx *context.Context, args *engine.APIRouteProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.APIRouteProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	rp, err := args.AsRouteProfile()
	if err != nil {
		return err
	}
	if err := adms.dm.SetRouteProfile(ctx, rp, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), args.Tenant, utils.CacheRouteProfiles,
		rp.TenantID(), &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveRouteProfile remove a specific Route configuration
func (adms *AdminSv1) RemoveRouteProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveRouteProfile(ctx, tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheRouteProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewRouteSv1(rS *engine.RouteService) *RouteSv1 {
	return &RouteSv1{rS: rS}
}

// RouteSv1 exports RPC from RouteS
type RouteSv1 struct {
	ping
	rS *engine.RouteService
}

// GetRoutes returns sorted list of routes for Event
func (rS *RouteSv1) GetRoutes(ctx *context.Context, args *engine.ArgsGetRoutes, reply *engine.SortedRoutesList) error {
	return rS.rS.V1GetRoutes(ctx, args, reply)
}

// GetRouteProfilesForEvent returns a list of route profiles that match for Event
func (rS *RouteSv1) GetRouteProfilesForEvent(ctx *context.Context, args *utils.CGREvent, reply *[]*engine.RouteProfile) error {
	return rS.rS.V1GetRouteProfilesForEvent(ctx, args, reply)
}

// GetRoutesList returns sorted list of routes for Event as a string slice
func (rS *RouteSv1) GetRoutesList(ctx *context.Context, args *engine.ArgsGetRoutes, reply *[]string) error {
	return rS.rS.V1GetRoutesList(ctx, args, reply)
}
