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

	"github.com/cgrates/cgrates/utils"
)

// GetAccountProfile returns an Account Profile
func (apierSv1 *APIerSv1) GetAccountProfile(arg *utils.TenantIDWithAPIOpts, reply *utils.AccountProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := apierSv1.DataManager.GetAccountProfile(tnt, arg.ID)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *ap
	return nil
}

// GetAccountProfileIDs returns list of action profile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetAccountProfileIDs(args *utils.PaginatorWithTenant, actPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.AccountProfilePrefix + tnt + utils.ConcatenatedKeySep
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
	*actPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetAccountProfileIDsCount sets in reply var the total number of AccountProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 AccountProfileIDs
func (apierSv1 *APIerSv1) GetAccountProfileIDsCount(args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.AccountProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

//SetAccountProfile add/update a new Account Profile
func (apierSv1 *APIerSv1) SetAccountProfile(extAp *utils.APIAccountProfileWithOpts, reply *string) error {
	if missing := utils.MissingStructFields(extAp.APIAccountProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if extAp.Tenant == utils.EmptyString {
		extAp.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := extAp.AsAccountProfile()
	if err != nil {
		return err
	}
	if err := apierSv1.DataManager.SetAccountProfile(ap, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccountProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAccountProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveAccountProfile remove a specific Account Profile
func (apierSv1 *APIerSv1) RemoveAccountProfile(arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveAccountProfile(tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccountProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAccountProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}
