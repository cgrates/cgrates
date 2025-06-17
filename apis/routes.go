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
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/routes"
	"github.com/cgrates/cgrates/utils"
)

// GetRouteProfile returns a Route configuration
func (adms *AdminSv1) GetRouteProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RouteProfile) error {
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
		*reply = *rp
	}
	return nil
}

// GetRouteProfileIDs returns list of routeProfile IDs registered for a tenant
func (adms *AdminSv1) GetRouteProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, routeProfileIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RouteProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	retIDs := make([]string, len(keys))
	for i, key := range keys {
		retIDs[i] = key[lenPrfx:]
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*routeProfileIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetRouteProfiles returns a list of route profiles registered for a tenant
func (admS *AdminSv1) GetRouteProfiles(ctx *context.Context, args *utils.ArgsItemIDs, rouPrfs *[]*utils.RouteProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var rouPrfIDs []string
	if err = admS.GetRouteProfileIDs(ctx, args, &rouPrfIDs); err != nil {
		return
	}
	*rouPrfs = make([]*utils.RouteProfile, 0, len(rouPrfIDs))
	for _, rouPrfID := range rouPrfIDs {
		var rouPrf *utils.RouteProfile
		rouPrf, err = admS.dm.GetRouteProfile(ctx, tnt, rouPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*rouPrfs = append(*rouPrfs, rouPrf)
	}
	return
}

// GetRouteProfilesCount sets in reply var the total number of RouteProfileIDs registered for the received tenant
// returns ErrNotFound in case of 0 RouteProfileIDs
func (adms *AdminSv1) GetRouteProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RouteProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	var keys []string
	if keys, err = adms.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetRouteProfile add a new Route configuration
func (adms *AdminSv1) SetRouteProfile(ctx *context.Context, args *utils.RouteProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.RouteProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if len(args.Routes) == 0 {
		return utils.NewErrMandatoryIeMissing("Routes")
	}
	if err := adms.dm.SetRouteProfile(ctx, args.RouteProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetRouteProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for SupplierProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheRouteProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRouteProfile remove a specific Route configuration
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
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveRouteProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for SupplierProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheRouteProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewRouteSv1 initializes the RouteSv1 object.
func NewRouteSv1(rpS *routes.RouteS) *RouteSv1 {
	return &RouteSv1{rpS: rpS}
}

// RouteSv1 represents the RPC object to register for routes v1 APIs.
type RouteSv1 struct {
	rpS *routes.RouteS
}

// V1GetRoutes returns the list of valid routes.
func (rpS *RouteSv1) V1GetRoutes(ctx *context.Context, args *utils.CGREvent, reply *routes.SortedRoutesList) (err error) {
	return rpS.rpS.V1GetRoutes(ctx, args, reply)
}

// V1GetRoutesList returns the list of valid routes.
func (rpS *RouteSv1) V1GetRoutesList(ctx *context.Context, args *utils.CGREvent, reply *[]string) (err error) {
	return rpS.rpS.V1GetRoutesList(ctx, args, reply)
}
