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
package admins

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetStatQueueProfile returns a StatQueue profile
func (adms *AdminS) V1GetStatQueueProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.StatQueueProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	sCfg, err := adms.dm.GetStatQueueProfile(ctx, tnt, arg.ID,
		true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// GetStatQueueProfileIDs returns list of statQueueProfile IDs registered for a tenant
func (adms *AdminS) V1GetStatQueueProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, stsPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep
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
	*stsPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetStatQueueProfiles returns a list of stats profiles registered for a tenant
func (admS *AdminS) V1GetStatQueueProfiles(ctx *context.Context, args *utils.ArgsItemIDs, sqPrfs *[]*engine.StatQueueProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var sqPrfIDs []string
	if err = admS.V1GetStatQueueProfileIDs(ctx, args, &sqPrfIDs); err != nil {
		return
	}
	*sqPrfs = make([]*engine.StatQueueProfile, 0, len(sqPrfIDs))
	for _, sqPrfID := range sqPrfIDs {
		var sqPrf *engine.StatQueueProfile
		sqPrf, err = admS.dm.GetStatQueueProfile(ctx, tnt, sqPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*sqPrfs = append(*sqPrfs, sqPrf)
	}
	return
}

// GetStatQueueProfilesCount returns the total number of StatQueueProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 StatQueueProfileIDs
func (admS *AdminS) V1GetStatQueueProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.StatQueueProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetStatQueueProfile alters/creates a StatQueueProfile
func (adms *AdminS) V1SetStatQueueProfile(ctx *context.Context, arg *engine.StatQueueProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.StatQueueProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err = adms.dm.SetStatQueueProfile(ctx, arg.StatQueueProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err = adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetStatQueueProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for StatQueueProfile
	if err = adms.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), arg.Tenant, utils.CacheStatQueueProfiles,
		arg.TenantID(), utils.EmptyString, &arg.FilterIDs, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveStatQueueProfile remove a specific stat configuration
func (adms *AdminS) V1RemoveStatQueueProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveStatQueueProfile(ctx, tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1RemoveStatQueueProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for StatQueueProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheStatQueueProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheStatQueueProfiles and CacheStatQueues and store it in database
	//make 1 insert for both StatQueueProfile and StatQueue instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheStatQueueProfiles: loadID, utils.CacheStatQueues: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
