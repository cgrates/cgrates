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

func NewResourceSv1(rls *engine.ResourceService) *ResourceSv1 {
	return &ResourceSv1{rls: rls}
}

// Exports RPC from RLs
type ResourceSv1 struct {
	rls *engine.ResourceService
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rsv1 *ResourceSv1) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rsv1, serviceMethod, args, reply)
}

// GetResourcesForEvent returns Resources matching a specific event
func (rsv1 *ResourceSv1) GetResourcesForEvent(args utils.ArgRSv1ResourceUsage, reply *engine.Resources) error {
	return rsv1.rls.V1ResourcesForEvent(args, reply)
}

// AuthorizeResources checks if there are limits imposed for event
func (rsv1 *ResourceSv1) AuthorizeResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AuthorizeResources(args, reply)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSv1) AllocateResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResource(args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSv1) ReleaseResources(args utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResource(args, reply)
}

// GetResource returns a resource configuration
func (rsv1 *ResourceSv1) GetResource(args *utils.TenantID, reply *engine.Resource) error {
	return rsv1.rls.V1GetResource(args, reply)
}

// GetResourceProfile returns a resource configuration
func (APIerSv1 *APIerSv1) GetResourceProfile(arg utils.TenantID, reply *engine.ResourceProfile) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rcfg, err := APIerSv1.DataManager.GetResourceProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rcfg
	}
	return nil
}

// GetResourceProfileIDs returns list of resourceProfile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetResourceProfileIDs(args utils.TenantArgWithPaginator, rsPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.ResourceProfilesPrefix + args.Tenant + ":"
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
	*rsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

type ResourceWithCache struct {
	*engine.ResourceProfile
	Cache *string
}

//SetResourceProfile adds a new resource configuration
func (APIerSv1 *APIerSv1) SetResourceProfile(arg *ResourceWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.ResourceProfile, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetResourceProfile(arg.ResourceProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(
		map[string]int64{utils.CacheResourceProfiles: loadID,
			utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheResourceProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	//add the resource only if it's not present
	if has, err := APIerSv1.DataManager.HasData(utils.ResourcesPrefix, arg.ID, arg.Tenant); err != nil {
		return err
	} else if !has {
		if err := APIerSv1.DataManager.SetResource(
			&engine.Resource{Tenant: arg.Tenant,
				ID:     arg.ID,
				Usages: make(map[string]*engine.ResourceUsage)}); err != nil {
			return utils.APIErrorHandler(err)
		}
		//handle caching for Resource
		argCache = utils.ArgsGetCacheItem{
			CacheID: utils.CacheResources,
			ItemID:  arg.TenantID(),
		}
		if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
			return utils.APIErrorHandler(err)
		}
	}

	*reply = utils.OK
	return nil
}

//RemoveResourceProfile remove a specific resource configuration
func (APIerSv1 *APIerSv1) RemoveResourceProfile(arg utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveResourceProfile(arg.Tenant, arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheResourceProfiles,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := APIerSv1.DataManager.RemoveResource(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheResourceProfiles: loadID, utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Resource
	argCache = utils.ArgsGetCacheItem{
		CacheID: utils.CacheResources,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (rsv1 *ResourceSv1) Ping(ign *utils.CGREventWithArgDispatcher, reply *string) error {
	*reply = utils.Pong
	return nil
}
