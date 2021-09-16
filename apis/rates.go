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
	"github.com/cgrates/cgrates/rates"

	"github.com/cgrates/cgrates/utils"
)

// GetRateProfile returns a Rate Profile based on tenant and id
func (admS *AdminSv1) GetRateProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	rPrf, err := admS.dm.GetRateProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = *rPrf
	return nil
}

// GetRateProfileIDs returns a list of rate profile IDs registered for a tenant
func (admS *AdminSv1) GetRateProfileIDs(ctx *context.Context, args *utils.PaginatorWithTenant, attrPrfIDs *[]string) error {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RateProfilePrefix + tnt + utils.ConcatenatedKeySep
	keys, err := admS.dm.DataDB().GetKeysForPrefix(ctx, prfx)
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

// GetRateProfileCount returns the total number of RateProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 RateProfileIDs
func (admS *AdminSv1) GetRateProfileCount(ctx *context.Context, args *utils.TenantWithAPIOpts, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var keys []string
	prfx := utils.RateProfilePrefix + tnt + utils.ConcatenatedKeySep
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// SetRateProfile add/update a new Rate Profile
func (admS *AdminSv1) SetRateProfile(ctx *context.Context, ext *utils.APIRateProfile, reply *string) error {
	if missing := utils.MissingStructFields(ext, []string{utils.ID, utils.Rates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ext.Tenant == utils.EmptyString {
		ext.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	rPrf, err := ext.AsRateProfile()
	if err != nil {
		return err
	}
	if err := admS.dm.SetRateProfile(ctx, rPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(ext.APIOpts[utils.CacheOpt]), rPrf.Tenant, utils.CacheRateProfiles,
		rPrf.TenantID(), &rPrf.FilterIDs, ext.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// SetRateProfileRates add/update Rates from existing RateProfiles
func (admS *AdminSv1) SetRateProfileRates(ctx *context.Context, ext *utils.APIRateProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(ext, []string{utils.ID, utils.Rates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if ext.Tenant == utils.EmptyString {
		ext.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	rPrf, err := ext.AsRateProfile()
	if err != nil {
		return err
	}
	if err = admS.dm.SetRateProfileRates(ctx, rPrf, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err = admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err = admS.CallCache(ctx, utils.IfaceAsString(ext.APIOpts[utils.CacheOpt]), rPrf.Tenant, utils.CacheRateProfiles,
		rPrf.TenantID(), &rPrf.FilterIDs, ext.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRateProfileRates removes the rates from the Rate Profile
func (admS *AdminSv1) RemoveRateProfileRates(ctx *context.Context, args *utils.RemoveRPrfRates, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveRateProfileRates(ctx, tnt, args.ID, args.RateIDs, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]), tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, args.ID), nil, args.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return nil
}

// RemoveRateProfile remove a specific Rate Profile specified by tenant and id
func (admS *AdminSv1) RemoveRateProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *string) error {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	if err := admS.dm.RemoveRateProfile(ctx, tnt, arg.ID,
		true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheAttributeProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), nil, arg.APIOpts); err != nil {
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
	ping
	rS *rates.RateS
}

// CostForEvent returs the costs for the event and all the rate profile information
func (rSv1 *RateSv1) CostForEvent(ctx *context.Context, args *utils.CGREvent, rpCost *utils.RateProfileCost) (err error) {
	return rSv1.rS.V1CostForEvent(ctx, args, rpCost)
}
