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

// GetResourceProfile returns a resource configuration
func (adms *AdminSv1) GetResourceProfile(ctx *context.Context, arg *utils.TenantID, reply *engine.ResourceProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if rcfg, err := adms.dm.GetResourceProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *rcfg
	}
	return nil
}

// GetResourceProfileIDs returns list of resourceProfile IDs registered for a tenant
func (adms *AdminSv1) GetResourceProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, rsPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
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
	*rsPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetResourceProfileIDsCount returns the total number of ResourceProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ResourceProfileIDs
func (admS *AdminSv1) GetResourceProfilesCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx,
		utils.ResourceProfilesPrefix+tnt+utils.ConcatenatedKeySep); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetResourceProfile adds a new resource configuration
func (adms *AdminSv1) SetResourceProfile(ctx *context.Context, arg *engine.ResourceProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.ResourceProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetResourceProfile(ctx, arg.ResourceProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err = adms.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheResourceProfiles: loadID,
			utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant, utils.CacheResourceProfiles,
		arg.TenantID(), &arg.FilterIDs, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//RemoveResourceProfile remove a specific resource configuration
func (adms *AdminSv1) RemoveResourceProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveResourceProfile(ctx, tnt, arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for ResourceProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheResourceProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheResourceProfiles and CacheResources and store it in database
	//make 1 insert for both ResourceProfile and Resources instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheResourceProfiles: loadID, utils.CacheResources: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewResourceSv1(rls *engine.ResourceService) *ResourceSv1 {
	return &ResourceSv1{rls: rls}
}

// Exports RPC from RLs
type ResourceSv1 struct {
	ping
	rls *engine.ResourceService
}

// GetResourcesForEvent returns Resources matching a specific event
func (rsv1 *ResourceSv1) GetResourcesForEvent(ctx *context.Context, args *utils.ArgRSv1ResourceUsage, reply *engine.Resources) error {
	return rsv1.rls.V1ResourcesForEvent(ctx, *args, reply)
}

// AuthorizeResources checks if there are limits imposed for event
func (rsv1 *ResourceSv1) AuthorizeResources(ctx *context.Context, args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AuthorizeResources(ctx, *args, reply)
}

// V1InitiateResourceUsage records usage for an event
func (rsv1 *ResourceSv1) AllocateResources(ctx *context.Context, args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1AllocateResources(ctx, *args, reply)
}

// V1TerminateResourceUsage releases usage for an event
func (rsv1 *ResourceSv1) ReleaseResources(ctx *context.Context, args *utils.ArgRSv1ResourceUsage, reply *string) error {
	return rsv1.rls.V1ReleaseResources(ctx, *args, reply)
}

// GetResource returns a resource configuration
func (rsv1 *ResourceSv1) GetResource(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.Resource) error {
	return rsv1.rls.V1GetResource(ctx, args, reply)
}

func (rsv1 *ResourceSv1) GetResourceWithConfig(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *engine.ResourceWithConfig) error {
	return rsv1.rls.V1GetResourceWithConfig(ctx, args, reply)
}
