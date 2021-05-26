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

//SetFilter add a new Filter
func (adms *AdminSv1) SetFilter(ctx *context.Context, arg *engine.FilterWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.Filter, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntID := arg.TenantID()
	argC := map[string][]string{utils.FilterIDs: {tntID}}
	if fltr, err := adms.dm.GetFilter(ctx, arg.Filter.Tenant, arg.Filter.ID, true, false, utils.NonTransactional); err != nil {
		if err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
	} else if argC, err = composeCacheArgsForFilter(adms.dm, ctx, fltr, fltr.Tenant, tntID, argC); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.dm.SetFilter(ctx, arg.Filter, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	if argC, err = composeCacheArgsForFilter(adms.dm, ctx, arg.Filter, arg.Filter.Tenant, tntID, argC); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := adms.dm.SetLoadIDs(ctx,
		map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	if err := callCacheForFilter(adms.connMgr, adms.cfg.AdminSCfg().CachesConns, ctx,
		utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, argC, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//GetFilter returns a Filter
func (adms *AdminSv1) GetFilter(ctx *context.Context, arg *utils.TenantID, reply *engine.Filter) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if fltr, err := adms.dm.GetFilter(ctx, tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *fltr
	}
	return nil
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (adms *AdminSv1) GetFilterIDs(ctx *context.Context, args *utils.PaginatorWithTenant, fltrIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
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
	*fltrIDs = args.PaginateStringSlice(retIDs)
	return nil
}

//RemoveFilter  remove a specific filter
func (adms *AdminSv1) RemoveFilter(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	if err := adms.dm.RemoveFilter(ctx, tnt, arg.ID, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := adms.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	if err := callCacheForFilter(adms.connMgr, adms.cfg.AdminSCfg().CachesConns, ctx,
		utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, map[string][]string{utils.FilterIDs: {utils.ConcatenatedKey(tnt, arg.ID)}}, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
