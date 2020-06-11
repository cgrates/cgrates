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

// GetRateProfile returns an Rate Profile
func (APIerSv1 *APIerSv1) GetRateProfile(arg *utils.TenantIDWithArgDispatcher, reply *engine.RateProfile) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rPrf, err := APIerSv1.DataManager.GetRateProfile(arg.Tenant, arg.ID, true, true, utils.NonTransactional); err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	} else {
		*reply = *rPrf
	}
	return nil
}

// GetRateProfileIDs returns list of rate profile IDs registered for a tenant
func (APIerSv1 *APIerSv1) GetRateProfileIDs(args *utils.TenantArgWithPaginator, attrPrfIDs *[]string) error {
	if missing := utils.MissingStructFields(args, []string{utils.Tenant}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	prfx := utils.RateProfilePrefix + args.Tenant + ":"
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
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetRateProfileIDsCount sets in reply var the total number of RateProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 RateProfileIDs
func (APIerSv1 *APIerSv1) GetRateProfileIDsCount(args *utils.TenantArg, reply *int) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.Tenant}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	var keys []string
	prfx := utils.RateProfilePrefix + args.Tenant + ":"
	if keys, err = APIerSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

type RateProfileWithCache struct {
	*engine.RateProfileWithArgDispatcher
	Cache *string
}

//SetRateProfile add/update a new Rate Profile
func (APIerSv1 *APIerSv1) SetRateProfile(rPrf *RateProfileWithCache, reply *string) error {
	if missing := utils.MissingStructFields(rPrf.RateProfile, []string{"Tenant", "ID", "Rates"}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}

	if err := APIerSv1.DataManager.SetRateProfile(rPrf.RateProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	args := utils.ArgsGetCacheItem{
		CacheID: utils.CacheRateProfiles,
		ItemID:  rPrf.TenantID(),
	}
	if err := APIerSv1.CallCache(GetCacheOpt(rPrf.Cache), args); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRateProfile remove a specific Rate Profile
func (APIerSv1 *APIerSv1) RemoveRateProfile(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{"Tenant", "ID"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if err := APIerSv1.DataManager.RemoveRateProfile(arg.Tenant, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := APIerSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	args := utils.ArgsGetCacheItem{
		CacheID: utils.CacheRateProfiles,
		ItemID:  utils.ConcatenatedKey(arg.Tenant, arg.ID),
	}
	if err := APIerSv1.CallCache(GetCacheOpt(arg.Cache), args); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
