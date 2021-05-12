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
}

//SetFilter add a new Filter
func (APIerSv1 *APIerSv1) SetFilter(arg *FilterWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg.Filter, []string{"Tenant", "ID"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.SetFilter(arg.Filter); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheFilters,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//GetFilter returns a Filter
func (APIerSv1 *APIerSv1) GetFilter(arg utils.TenantID, reply *engine.Filter) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if fltr, err := engine.GetFilter(APIerSv1.DataManager, arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	} else {
		*reply = *fltr
	}
	return nil
}

// GetFilterIDs returns list of Filter IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetFilterIDs(args utils.TenantArgWithPaginator, fltrIDs *[]string) error {
	if missing := utils.MissingStructFields(&args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.FilterPrefix + args.Tenant + ":"
	keys, err := APIerSv1.DataManager.DataDB().GetKeysForPrefix(prfx)
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
func (APIerSv1 *APIerSv1) RemoveFilter(arg utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(&arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveFilter(arg.Tenant, arg.ID, utils.NonTransactional); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheFilters and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheFilters: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	//handle caching for Filter
	argCache := utils.ArgsGetCacheItem{
		CacheID: utils.CacheFilters,
		ItemID:  arg.TenantID(),
	}
	if err := APIerSv1.CallCache(arg.Tenant, GetCacheOpt(arg.Cache), argCache); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
