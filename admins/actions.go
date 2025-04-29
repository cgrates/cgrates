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

package admins

import (
	"fmt"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// GetActionProfile returns an Action Profile
func (admS *AdminS) V1GetActionProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.ActionProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	ap, err := admS.dm.GetActionProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}

// GetActionProfileIDs returns list of action profile IDs registered for a tenant
func (admS *AdminS) V1GetActionProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, actPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
	*actPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetActionProfiles returns a list of action profiles registered for a tenant
func (admS *AdminS) V1GetActionProfiles(ctx *context.Context, args *utils.ArgsItemIDs, actPrfs *[]*utils.ActionProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var actPrfIDs []string
	if err = admS.V1GetActionProfileIDs(ctx, args, &actPrfIDs); err != nil {
		return
	}
	*actPrfs = make([]*utils.ActionProfile, 0, len(actPrfIDs))
	for _, actPrfID := range actPrfIDs {
		var ap *utils.ActionProfile
		ap, err = admS.dm.GetActionProfile(ctx, tnt, actPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*actPrfs = append(*actPrfs, ap)
	}
	return
}

// GetActionProfilesCount sets in reply var the total number of ActionProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 ActionProfileIDs
func (admS *AdminS) V1GetActionProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.ActionProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
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

// SetActionProfile add/update a new Action Profile
func (admS *AdminS) V1SetActionProfile(ctx *context.Context, ap *utils.ActionProfileWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(ap.ActionProfile, []string{utils.ID, utils.Actions}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ap.Tenant == utils.EmptyString {
		ap.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}

	if err := admS.dm.SetActionProfile(ctx, ap.ActionProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1SetActionProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(ap.APIOpts[utils.MetaCache]), ap.Tenant, utils.CacheActionProfiles,
		ap.TenantID(), utils.EmptyString, &ap.FilterIDs, ap.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveActionProfile remove a specific Action Profile
func (admS *AdminS) V1RemoveActionProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveActionProfile(ctx, tnt, arg.ID,
		true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheActionProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheActionProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	// delay if needed before cache call
	if admS.cfg.GeneralCfg().CachingDelay != 0 {
		utils.Logger.Info(fmt.Sprintf("<V1RemoveActionProfile> Delaying cache call for %v", admS.cfg.GeneralCfg().CachingDelay))
		time.Sleep(admS.cfg.GeneralCfg().CachingDelay)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheActionProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
