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

package v1

import (
	"time"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetRouteProfile returns a Route configuration
func (apierSv1 *APIerSv1) GetRouteProfile(arg *utils.TenantID, reply *engine.RouteProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if rp, err := apierSv1.DataManager.GetRouteProfile(tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rp
	}
	return nil
}

// GetRouteProfileIDs returns list of routeProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetRouteProfileIDs(args *utils.PaginatorWithTenant, sppPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.RouteProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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

type RouteWithOpts struct {
	*engine.RouteProfile
	Opts map[string]interface{}
}

//SetRouteProfile add a new Route configuration
func (apierSv1 *APIerSv1) SetRouteProfile(args *RouteWithOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.RouteProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetRouteProfile(args.RouteProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.Opts[utils.CacheOpt]), args.Tenant, utils.CacheRouteProfiles,
		args.TenantID(), &args.FilterIDs, nil, args.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveRouteProfile remove a specific Route configuration
func (apierSv1 *APIerSv1) RemoveRouteProfile(args *utils.TenantIDWithOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveRouteProfile(tnt, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(args.Opts[utils.CacheOpt]), tnt, utils.CacheRouteProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewRouteSv1(rS *engine.RouteService) *RouteSv1 {
	return &RouteSv1{rS: rS}
}

// Exports RPC from RouteS
type RouteSv1 struct {
	rS *engine.RouteService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rS *RouteSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rS, serviceMethod, args, reply)
}

// GetRoutes returns sorted list of routes for Event
func (rS *RouteSv1) GetRoutes(args *engine.ArgsGetRoutes,
	reply *engine.SortedRoutes) error {
	return rS.rS.V1GetRoutes(args, reply)
}

// GetRouteProfilesForEvent returns a list of route profiles that match for Event
func (rS *RouteSv1) GetRouteProfilesForEvent(args *utils.CGREvent,
	reply *[]*engine.RouteProfile) error {
	return rS.rS.V1GetRouteProfilesForEvent(args, reply)
}

func (rS *RouteSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
