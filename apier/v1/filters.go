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

type FilterWithCache struct {
	*engine.Filter
	Cache *string
	Opts  map[string]interface{}
}

//SetFilter add a new Filter
func (apierSv1 *APIerSv1) SetFilter(arg *FilterWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.Filter, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if arg.Tenant == utils.EmptyString {
		arg.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.SetFilter(arg.Filter, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	if err := apierSv1.CallCache(arg.Cache, arg.Tenant, utils.CacheFilters,
		arg.TenantID(), nil, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//GetFilter returns a Filter
func (apierSv1 *APIerSv1) GetFilter(arg *utils.TenantID, reply *engine.Filter) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if fltr, err := apierSv1.DataManager.GetFilter(tnt, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *fltr
	}
	return nil
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (apierSv1 *APIerSv1) GetFilterIDs(args *utils.PaginatorWithTenant, fltrIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.FilterPrefix + tnt + utils.ConcatenatedKeySep
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
	*fltrIDs = args.PaginateStringSlice(retIDs)
	return nil
}

//RemoveFilter  remove a specific filter
func (apierSv1 *APIerSv1) RemoveFilter(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveFilter(tnt, arg.ID, utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	if err := apierSv1.CallCache(arg.Cache, tnt, utils.CacheFilters,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
