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

	"github.com/cgrates/cgrates/rates"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// GetRateProfile returns an Rate Profile
func (apierSv1 *APIerSv1) GetRateProfile(arg *utils.TenantIDWithOpts, reply *engine.RateProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	rPrf, err := apierSv1.DataManager.GetRateProfile(tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rPrf
	return nil
}

// GetRateProfileIDs returns list of rate profile IDs registered for a tenant
func (apierSv1 *APIerSv1) GetRateProfileIDs(args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	prfx := utils.RateProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
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
	*attrPrfIDs = args.PaginateStringSlice(retIDs)
	return nil
}

// GetRateProfileIDsCount sets in reply var the total number of RateProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 RateProfileIDs
func (apierSv1 *APIerSv1) GetRateProfileIDsCount(args *utils.TenantWithOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.RateProfilePrefix + tnt + utils.CONCATENATED_KEY_SEP
	if keys, err = apierSv1.DataManager.DataDB().GetKeysForPrefix(prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

type RateProfileWithCache struct {
	*engine.RateProfileWithOpts
	Cache *string
}

//SetRateProfile add/update a new Rate Profile
func (apierSv1 *APIerSv1) SetRateProfile(rPrf *RateProfileWithCache, reply *string) error {
	if missing := utils.MissingStructFields(rPrf.RateProfile, []string{utils.ID, utils.Rates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rPrf.Tenant == utils.EmptyString {
		rPrf.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}

	if err := apierSv1.DataManager.SetRateProfile(rPrf.RateProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(rPrf.Cache, rPrf.Tenant, utils.CacheRateProfiles,
		rPrf.TenantID(), &rPrf.FilterIDs, nil, rPrf.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

//SetRateProfileRates add/update Rates from existing RateProfiles
func (apierSv1 *APIerSv1) SetRateProfileRates(rPrf *RateProfileWithCache, reply *string) (err error) {
	if missing := utils.MissingStructFields(rPrf.RateProfile, []string{utils.ID, utils.Rates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if rPrf.Tenant == utils.EmptyString {
		rPrf.Tenant = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err = apierSv1.DataManager.SetRateProfileRates(rPrf.RateProfile, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err = apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err = apierSv1.CallCache(rPrf.Cache, rPrf.Tenant, utils.CacheRateProfiles,
		rPrf.TenantID(), &rPrf.FilterIDs, nil, rPrf.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

type RemoveRPrfRates struct {
	Tenant  string
	ID      string
	RateIDs []string
	Cache   *string
	Opts    map[string]interface{}
}

func (apierSv1 *APIerSv1) RemoveRateProfileRates(args *RemoveRPrfRates, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveRateProfileRates(tnt, args.ID, args.RateIDs, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(args.Cache, tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, nil, args.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRateProfile remove a specific Rate Profile
func (apierSv1 *APIerSv1) RemoveRateProfile(arg *utils.TenantIDWithCache, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = apierSv1.Config.GeneralCfg().DefaultTenant
	}
	if err := apierSv1.DataManager.RemoveRateProfile(tnt, arg.ID,
		utils.NonTransactional, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := apierSv1.DataManager.SetLoadIDs(map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := apierSv1.CallCache(arg.Cache, tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, nil, arg.Opts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

func NewRateSv1(rateS *rates.RateS) *RateSv1 {
	return &RateSv1{rS: rateS}
}

// Exports RPC from RLs
type RateSv1 struct {
	rS *rates.RateS
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rSv1 *RateSv1) Call(serviceMethod string,
	args interface{}, reply interface{}) error {
	return utils.APIerRPCCall(rSv1, serviceMethod, args, reply)
}

func (rSv1 *RateSv1) CostForEvent(args *utils.ArgsCostForEvent, rpCost *engine.RateProfileCost) (err error) {
	return rSv1.rS.V1CostForEvent(args, rpCost)
}

func (rSv1 *RateSv1) Ping(ign *utils.CGREventWithOpts, reply *string) error {
	*reply = utils.Pong
	return nil
}
