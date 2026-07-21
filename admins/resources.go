// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package admins

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// GetResourceProfile returns a resource configuration
func (adms *AdminS) V1GetResourceProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ResourceProfile) error {
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
func (adms *AdminS) V1GetResourceProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, rsPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	db, _, err := adms.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = db.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
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
func (admS *AdminS) V1GetResourceProfiles(ctx *context.Context, args *utils.ArgsItemIDs, rsPrfs *[]*utils.ResourceProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var rsPrfIDs []string
	if err = admS.V1GetResourceProfileIDs(ctx, args, &rsPrfIDs); err != nil {
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
func (admS *AdminS) V1GetResourceProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ResourceProfilesPrefix + tnt + utils.ConcatenatedKeySep
	db, _, err := admS.dm.DBConns().GetConn(utils.MetaResourceProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = db.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetResourceProfile adds a new resource configuration
func (adms *AdminS) V1SetResourceProfile(ctx *context.Context, arg *utils.ResourceProfileWithAPIOpts, reply *string) (err error) {
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
		utils.Logger.Info(fmt.Sprintf("<V1SetResourceProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
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
func (adms *AdminS) V1RemoveResourceProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
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
		utils.Logger.Info(fmt.Sprintf("<V1RemoveResourceProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
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
