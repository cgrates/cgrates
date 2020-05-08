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
func (APIerSv1 *APIerSv1) GetRouteProfile(arg utils.TenantID, reply *engine.RouteProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rp, err := APIerSv1.DataManager.GetRouteProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rp
	}
	return nil
}

// GetRouteProfileIDs returns list of routeProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetRouteProfileIDs(args utils.TenantArgWithPaginator, sppPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.RouteProfilePrefix + args.Tenant + ":"
	keys, err := APIerSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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

type RouteWithCache struct {
	*engine.RouteProfile
	Cache *string
}

//SetRouteProfile add a new Route configuration
func (APIerSv1 *APIerSv1) SetRouteProfile(args *RouteWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args.RouteProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetRouteProfile(args.RouteProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheRouteProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(GetCacheOpt(args.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveRouteProfile remove a specific Route configuration
func (APIerSv1 *APIerSv1) RemoveRouteProfile(args *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveRouteProfile(args.Tenant, args.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRouteProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRouteProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for SupplierProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheRouteProfiles,
		ItemID:  args.TenantID(),
	}
	if err := APIerSv1.CallCache(GetCacheOpt(args.Cache), argCache); err != nil {
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
func (rS *RouteSv1) GetRouteProfilesForEvent(args *utils.CGREventWithArgDispatcher,
	reply *[]*engine.RouteProfile) error {
	return rS.rS.V1GetRouteProfilesForEvent(args, reply)
}

func (rS *RouteSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
