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

// GetThresholdProfile returns a Threshold Profile
func (adms *AdminS) V1GetThresholdProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.ThresholdProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	th, err := adms.dm.GetThresholdProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *th
	return
}

// GetThresholdProfileIDs returns list of thresholdProfile IDs registered for a tenant
func (adms *AdminS) V1GetThresholdProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, thPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaThresholdProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
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
	*thPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetThresholdProfiles returns a list of threshold profiles registered for a tenant
func (admS *AdminS) V1GetThresholdProfiles(ctx *context.Context, args *utils.ArgsItemIDs, thdPrfs *[]*engine.ThresholdProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var thdPrfIDs []string
	if err = admS.V1GetThresholdProfileIDs(ctx, args, &thdPrfIDs); err != nil {
		return
	}
	*thdPrfs = make([]*engine.ThresholdProfile, 0, len(thdPrfIDs))
	for _, thdPrfID := range thdPrfIDs {
		var thdPrf *engine.ThresholdProfile
		thdPrf, err = admS.dm.GetThresholdProfile(ctx, tnt, thdPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*thdPrfs = append(*thdPrfs, thdPrf)
	}
	return
}

// GetThresholdProfilesCount sets in reply var the total number of ThresholdProfileIDs registered for the received tenant
// returns ErrNotFound in case of 0 ThresholdProfileIDs
func (adms *AdminS) V1GetThresholdProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ThresholdProfilePrefix + tnt + utils.ConcatenatedKeySep
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaThresholdProfiles)
	if err != nil {
		return err
	}
	var keys []string
	if keys, err = dataDB.GetKeysForPrefix(ctx, prfx, args.ItemsSearch); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetThresholdProfile alters/creates a ThresholdProfile
func (adms *AdminS) V1SetThresholdProfile(ctx *context.Context, args *engine.ThresholdProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args.ThresholdProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.SetThresholdProfile(ctx, args.ThresholdProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetThresholdProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ThresholdProfile and Threshold
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheThresholdProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveThresholdProfile removes a specific Threshold Profile
func (adms *AdminS) V1RemoveThresholdProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveThresholdProfile(ctx, tnt, args.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if adms.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1RemoveThresholdProfile> Delaying cache call for %v", adms.cfg.GeneralCfg().CachingDelay))
		time.Sleep(adms.cfg.GeneralCfg().CachingDelay)
	}
	//handle caching for ThresholdProfile
	if err := adms.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheThresholdProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheThresholdProfiles and CacheThresholds and store it in database
	//make 1 insert for both ThresholdProfile and Threshold instead of 2
	loadID := time.Now().UnixNano()
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheThresholdProfiles: loadID, utils.CacheThresholds: loadID}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
