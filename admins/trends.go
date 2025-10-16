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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/utils"
)

// V1GetTrendProfile returns a Trend profile
func (a *AdminS) V1GetTrendProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.TrendProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = a.cfg.GeneralCfg().DefaultTenant
	}
	sCfg, err := a.dm.GetTrendProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = *sCfg
	return
}

// V1GetTrendProfileIDs returns list of TrendProfile IDs registered for a tenant
func (a *AdminS) V1GetTrendProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, stsPrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = a.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.TrendProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := a.dm.DBConns().GetConn(utils.MetaTrendProfiles)
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
	*stsPrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// V1GetTrendProfiles returns a list of stats profiles registered for a tenant
func (a *AdminS) V1GetTrendProfiles(ctx *context.Context, args *utils.ArgsItemIDs, sqPrfs *[]*utils.TrendProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = a.cfg.GeneralCfg().DefaultTenant
	}
	var sqPrfIDs []string
	if err = a.V1GetTrendProfileIDs(ctx, args, &sqPrfIDs); err != nil {
		return
	}
	*sqPrfs = make([]*utils.TrendProfile, 0, len(sqPrfIDs))
	for _, sqPrfID := range sqPrfIDs {
		var sqPrf *utils.TrendProfile
		sqPrf, err = a.dm.GetTrendProfile(ctx, tnt, sqPrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*sqPrfs = append(*sqPrfs, sqPrf)
	}
	return
}

// V1GetTrendProfilesCount returns the total number of TrendProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 TrendProfileIDs
func (a *AdminS) V1GetTrendProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = a.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.TrendProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := a.dm.DBConns().GetConn(utils.MetaTrendProfiles)
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

// V1SetTrendProfile alters/creates a TrendProfile
func (a *AdminS) V1SetTrendProfile(ctx *context.Context, arg *utils.TrendProfileWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.TrendProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = a.cfg.GeneralCfg().DefaultTenant
	}
	if err = a.dm.SetTrendProfile(ctx, arg.TrendProfile); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// V1RemoveTrendProfile remove a specific stat configuration
func (a *AdminS) V1RemoveTrendProfile(ctx *context.Context, args *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = a.cfg.GeneralCfg().DefaultTenant
	}
	if err := a.dm.RemoveTrendProfile(ctx, tnt, args.ID); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
