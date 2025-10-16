/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package apis

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/resources"
	"github.com/cgrates/cgrates/utils"
)

// GetResourceProfile returns a resource configuration
func (adms *AdminSv1) GetResourceProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ResourceProfile) error {
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
func (adms *AdminSv1) GetResourceProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, rsPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx); err != nil {
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
	*rsPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetResourceProfiles returns a list of resource profiles registered for a tenant
func (admS *AdminSv1) GetResourceProfiles(ctx *context.Context, args *utils.ArgsItemIDs, rsPrfs *[]*utils.ResourceProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var rsPrfIDs []string
	if err = admS.GetResourceProfileIDs(ctx, args, &rsPrfIDs); err != nil {
		return
	}
	*rsPrfs = make([]*utils.ResourceProfile, 0, len(rsPrfIDs))
	for _, rsPrfID := range rsPrfIDs {
		var rsPrf *utils.ResourceProfile
		rsPrf, err = admS.dm.GetResourceProfile(ctx, tnt, rsPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*rsPrfs = append(*rsPrfs, rsPrf)
	}
	return
}

// GetResourceProfilesCount returns the total number of ResourceProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ResourceProfileIDs
func (admS *AdminSv1) GetResourceProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetResourceProfile adds a new resource configuration
func (adms *AdminSv1) SetResourceProfile(ctx *context.Context, arg *utils.ResourceProfileWithAPIOpts, reply *string) (err error) {
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
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.SetResourceProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ResourceProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheResourceProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveResourceProfile remove a specific resource configuration
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
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<AdminSv1.RemoveResourceProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ResourceProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheResourceProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
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

// NewResourceSv1 initializes the ResourceSv1 object.
func NewResourceSv1(rsS *resources.ResourceS) *ResourceSv1 {
	return &ResourceSv1{rsS: rsS}
}

// ResourceSv1 represents the RPC object to register for resources v1 APIs.
type ResourceSv1 struct {
	rsS *resources.ResourceS
}

// V1GetResourcesForEvent returns active resource configs matching the event
func (rS *ResourceSv1) V1GetResourcesForEvent(ctx *context.Context, args *utils.CGREvent, reply *resources.Resources) (err error) {
	return rS.rsS.V1GetResourcesForEvent(ctx, args, reply)
}

// V1AuthorizeResources queries service to find if an Usage is allowed
func (rS *ResourceSv1) V1AuthorizeResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return rS.rsS.V1AuthorizeResources(ctx, args, reply)
}

// V1AllocateResources is called when a resource requires allocation
func (rS *ResourceSv1) V1AllocateResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return rS.rsS.V1AllocateResources(ctx, args, reply)
}

// V1ReleaseResources is called when we need to clear an allocation
func (rS *ResourceSv1) V1ReleaseResources(ctx *context.Context, args *utils.CGREvent, reply *string) (err error) {
	return rS.rsS.V1ReleaseResources(ctx, args, reply)
}

// V1GetResource returns a resource
func (rS *ResourceSv1) V1GetResource(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.Resource) error {
	return rS.rsS.V1GetResource(ctx, arg, reply)
}

// V1GetResource returns a resource configuration
func (rS *ResourceSv1) V1GetResourceWithConfig(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ResourceWithConfig) (err error) {
	return rS.rsS.V1GetResourceWithConfig(ctx, arg, reply)
}
