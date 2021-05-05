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
func (rsv1 *ResourceSv1) GetResourcesForEvent(args *utils.ArgRSv1ResourceUsage, reply *engine.Resources) error {
	return rsv1.rls.V1ResourcesForEvent(*args, reply)
}

// AuthorizeResources checks if there are limits imposed for event
func (rsv1 *ResourceSv1) AuthorizeResources(args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AuthorizeResources(*args, reply)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSv1) AllocateResources(args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResources(*args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSv1) ReleaseResources(args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResources(*args, reply)
}

// GetResource returns a resource configuration
func (rsv1 *ResourceSv1) GetResource(args *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return rsv1.rls.V1GetResource(args, reply)
}

func (rsv1 *ResourceSv1) GetResourceWithConfig(args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) error {
	return rsv1.rls.V1GetResourceWithConfig(args, reply)
}

// GetResourceProfile returns a resource configuration
func (apierSv1 *APIerSv1) GetResourceProfile(arg *utils.TenantID, reply *engine.ResourceProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := utils.EmptyString
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if rcfg, err := apierSv1.DataManager.GetResourceProfile(tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rcfg
	}
	return nil
}

// GetResourceProfileIDs returns list of resourceProfile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetResourceProfileIDs(args *utils.PaginatorWithTenant, rsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
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
	*rsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

//SetResourceProfile adds a new resource configuration
func (apierSv1 *APIerSv1) SetResourceProfile(arg *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.ResourceProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err = apierSv1.DataManager.SetResourceProfile(arg.ResourceProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err = apierSv1.DataManager.SetLoadIDs(
		map[string]int64{utils.CacheResourceProfiles: loadID,
			utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	if err = apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheResourceProfiles,
		arg.TenantID(), &arg.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	var ttl *time.Duration
	if arg.UsageTTL > 0 {
		ttl = &arg.UsageTTL
	}
	// for non stored we do not save the metrics
	if err = apierSv1.DataManager.SetResource(&engine.Resource{
		Tenant: arg.Tenant,
		ID:     arg.ID,
		Usages: make(map[string]*engine.ResourceUsage),
	}, ttl, arg.Limit, !arg.Stored); err != nil {
		return
	}
	//handle caching for Resource
	if err = apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheResources,
		arg.TenantID(), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}

	*reply = utils.OK
	return nil
}

//RemoveResourceProfile remove a specific resource configuration
func (apierSv1 *APIerSv1) RemoveResourceProfile(arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveResourceProfile(tnt, arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheResourceProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.DataManager.RemoveResource(tnt, arg.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheResourceProfiles: loadID, utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Resource
	if err := apierSv1.CallCache(utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheResources,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func (rsv1 *ResourceSv1) Ping(ign *utils.CGREvent, reply *string) error {
	*reply = utils.Pong
	return nil
}
