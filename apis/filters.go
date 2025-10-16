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
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func validateFilterRules(rules []*engine.FilterRule) error {
	for _, rule := range rules {
		if !rule.IsValid() {
			return fmt.Errorf("there exists at least one filter rule that is not valid")
		}
	}
	return nil
}

// SetFilter add a new Filter
func (adms *AdminSv1) SetFilter(ctx *context.Context, arg *engine.FilterWithAPIOpts, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg.Filter, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if len(arg.Rules) == 0 {
		return utils.NewErrMandatoryIeMissing("Filter Rules")
	}
	if err = validateFilterRules(arg.Rules); err != nil {
		return utils.APIErrorHandler(err)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntID := arg.TenantID()
	argC := map[string][]string{utils.CacheFilters: {tntID}}
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
		utils.IfaceAsString(arg.APIOpts[utils.MetaCache]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, argC, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return
}

// GetFilter returns a Filter
func (adms *AdminSv1) GetFilter(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *engine.Filter) error {
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

// GetFilters returns a list of filters for a tenant
func (adms *AdminSv1) GetFilters(ctx *context.Context, args *utils.ArgsItemIDs, fltrs *[]*engine.Filter) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var fltrIDs []string
	if err = adms.GetFilterIDs(ctx, args, &fltrIDs); err != nil {
		return
	}
	*fltrs = make([]*engine.Filter, 0, len(fltrIDs))
	for _, fltrID := range fltrIDs {
		var fltr *engine.Filter
		if fltr, err = adms.dm.GetFilter(ctx, tnt, fltrID, true, true, utils.NonTransactional); err != nil {
			return utils.APIErrorHandler(err)
		}
		*fltrs = append(*fltrs, fltr)
	}
	return
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (adms *AdminSv1) GetFilterIDs(ctx *context.Context, args *utils.ArgsItemIDs, fltrIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	dataDB, _, err := adms.dm.DBConns().GetConn(utils.MetaFilters)
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
	*fltrIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// RemoveFilter  remove a specific filter
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
		utils.IfaceAsString(arg.APIOpts[utils.MetaCache]),
		adms.cfg.GeneralCfg().DefaultCaching,
		arg.Tenant, map[string][]string{utils.CacheFilters: {utils.ConcatenatedKey(tnt, arg.ID)}}, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// GetFiltersCount returns the total number of FilterIDs registered for a tenant
// returns ErrNotFound in case of 0 FilterIDs
func (admS *AdminSv1) GetFiltersCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	dataDB, _, err := admS.dm.DBConns().GetConn(utils.MetaFilters)
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

// FiltersMatch checks whether a set of filter IDs passes for the provided CGREvent
func (admS *AdminSv1) FiltersMatch(ctx *context.Context, args *engine.ArgsFiltersMatch, reply *bool) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	evDP := args.CGREvent.AsDataProvider()
	var pass bool
	if pass, err = admS.fltrS.Pass(ctx, tnt, args.FilterIDs, evDP); err != nil {
		return
	} else if pass {
		*reply = true
	}
	return
}
