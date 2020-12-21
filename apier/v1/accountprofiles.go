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

	"github.com/cgrates/cgrates/accounts"

	"github.com/cgrates/cgrates/utils"
)

// GetAccountProfile returns an Account Profile
func (apierSv1 *APIerSv1) GetAccountProfile(arg *utils.TenantIDWithOpts, reply *utils.AccountProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	ap, err := apierSv1.DataManager.GetAccountProfile(tnt, arg.ID, true, true, utils.NonTransactional)
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
	prfx := utils.AccountProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
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
func (apierSv1 *APIerSv1) GetAccountProfileIDsCount(args *utils.TenantWithOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.AccountProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

type AccountProfileWithCache struct {
	*utils.AccountProfileWithOpts
	Cache *string
}

//SetAccountProfile add/update a new Account Profile
func (apierSv1 *APIerSv1) SetAccountProfile(ap *AccountProfileWithCache, reply *string) error {
	if missing := utils.MissingStructFields(ap.AccountProfile, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ap.Tenant == utils.EmptyString {
		ap.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}

	if err := apierSv1.DataManager.SetAccountProfile(ap.AccountProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAccountProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheAccountProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(ap.Cache, ap.Tenant, utils.CacheAccountProfiles,
		ap.TenantID(), &ap.FilterIDs, nil, ap.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveAccountProfile remove a specific Account Profile
func (apierSv1 *APIerSv1) RemoveAccountProfile(arg *utils.TenantIDWithCache, reply *string) error {
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
	if err := apierSv1.CallCache(arg.Cache, tnt, utils.CacheAccountProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// NewAccountSv1 initializes AccountSv1
func NewAccountSv1(aS *accounts.AccountS) *AccountSv1 {
	return &AccountSv1{aS: aS}
}

// AccountSv1 exports RPC from RLs
type AccountSv1 struct {
	aS *accounts.AccountS
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aSv1 *AccountSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(aSv1, serviceMethod, args, reply)
}

// Ping return pong if the service is active
func (aSv1 *AccountSv1) Ping(ign *utils.CGREventWithOpts, reply *string) error {
	*reply = utils.Pong
	return nil
}
