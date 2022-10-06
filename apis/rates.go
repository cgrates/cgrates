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
	"sort"
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/rates"

	"github.com/cgrates/cgrates/utils"
)

// GetRateProfile returns a Rate Profile based on tenant and id
func (admS *AdminSv1) GetRateProfile(ctx *context.Context, arg *utils.TenantIDWithAPIOpts, reply *utils.RateProfile) (err error) {
	if missing := utils.MissingStructFields(arg, []string{utils.ID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var rPrf *utils.RateProfile
	rPrf, err = admS.dm.GetRateProfile(ctx, tnt, arg.ID, true, true, utils.NonTransactional)
	if err != nil {
		if err.Error() != utils.ErrNotFound.Error() {
			err = utils.NewErrServerError(err)
		}
		return
	}
	rateIDs := make([]string, 0, len(rPrf.Rates))
	prefix := utils.IfaceAsString(arg.APIOpts[utils.ItemsPrefixOpt])
	for rateID := range rPrf.Rates {
		if strings.HasPrefix(rateID, prefix) {
			rateIDs = append(rateIDs, rateID)
		}
	}
	sort.Strings(rateIDs)
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(arg.APIOpts); err != nil {
		return
	}
	rateIDs, err = utils.Paginate(rateIDs, limit, offset, maxItems)
	if err != nil {
		return
	}
	paginatedRatePrf := &utils.RateProfile{
		Tenant:          rPrf.Tenant,
		ID:              rPrf.ID,
		FilterIDs:       rPrf.FilterIDs,
		Weights:         rPrf.Weights,
		MinCost:         rPrf.MinCost,
		MaxCost:         rPrf.MaxCost,
		MaxCostStrategy: rPrf.MaxCostStrategy,
	}
	paginatedRatePrf.Rates = make(map[string]*utils.Rate)
	for _, rateID := range rateIDs {
		paginatedRatePrf.Rates[rateID] = rPrf.Rates[rateID].Clone()
	}
	*reply = *paginatedRatePrf
	return
}

// GetRateProfile returns the rates of a profile based on their profile. Those rates will be returned back by matching a prefix.
func (admS *AdminSv1) GetRateProfileRates(ctx *context.Context, args *utils.ArgsSubItemIDs, reply *[]*utils.Rate) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ProfileID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	_, rates, err := admS.dm.GetRateProfileRates(ctx, args, false)
	if err != nil {
		return
	}
	if len(rates) == 0 {
		return utils.ErrNotFound
	}
	*reply = rates
	return
}

// GetRateProfileIDs returns a list of rate profile IDs registered for a tenant
func (admS *AdminSv1) GetRateProfileIDs(ctx *context.Context, args *utils.ArgsItemIDs, ratePrfIDs *[]string) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RateProfilePrefix + tnt + utils.ConcatenatedKeySep
	lenPrfx := len(prfx)
	prfx += args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
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
	*ratePrfIDs, err = utils.Paginate(retIDs, limit, offset, maxItems)
	return
}

// GetRateProfileRateIDs returns a list of rates from a specific RateProfile  registered for a tenant. RateIDs are returned back by matching a pattern given by ItemPrefix. If the ItemPrefix is not there, it will be returned all RateIDs.
func (admS *AdminSv1) GetRateProfileRateIDs(ctx *context.Context, args *utils.ArgsSubItemIDs, rateIDs *[]string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ProfileID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	var ids []string
	ids, _, err = admS.dm.GetRateProfileRates(ctx, args, true)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		return utils.ErrNotFound
	}
	var limit, offset, maxItems int
	if limit, offset, maxItems, err = utils.GetPaginateOpts(args.APIOpts); err != nil {
		return
	}
	*rateIDs, err = utils.Paginate(ids, limit, offset, maxItems)
	return
}

// GetRateProfiles returns a list of rate profiles registered for a tenant
func (admS *AdminSv1) GetRateProfiles(ctx *context.Context, args *utils.ArgsItemIDs, ratePrfs *[]*utils.RateProfile) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	var ratePrfIDs []string
	if err = admS.GetRateProfileIDs(ctx, args, &ratePrfIDs); err != nil {
		return
	}
	*ratePrfs = make([]*utils.RateProfile, 0, len(ratePrfIDs))
	for _, ratePrfID := range ratePrfIDs {
		var ratePrf *utils.RateProfile
		ratePrf, err = admS.dm.GetRateProfile(ctx, tnt, ratePrfID, true, true, utils.NonTransactional)
		if err != nil {
			return utils.APIErrorHandler(err)
		}
		*ratePrfs = append(*ratePrfs, ratePrf)
	}
	return
}

// GetRateProfilesCount returns the total number of RateProfileIDs registered for a tenant
// returns ErrNotFound in case of 0 RateProfileIDs
func (admS *AdminSv1) GetRateProfilesCount(ctx *context.Context, args *utils.ArgsItemIDs, reply *int) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = admS.cfg.GeneralCfg().DefaultTenant
	}
	prfx := utils.RateProfilePrefix + tnt + utils.ConcatenatedKeySep + args.ItemsPrefix
	var keys []string
	if keys, err = admS.dm.DataDB().GetKeysForPrefix(ctx, prfx); err != nil {
		return err
	}
	if len(keys) == 0 {
		return utils.ErrNotFound
	}
	*reply = len(keys)
	return
}

// GetRateProfileRatesCount count the rates from a specific RateProfile  registered for a tenant. The number of rates is returned back by matching a pattern given by ItemPrefix. If the ItemPrefix is not there, it will be counted all the rates.
func (admS *AdminSv1) GetRateProfileRatesCount(ctx *context.Context, args *utils.ArgsSubItemIDs, countIDs *int) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ProfileID}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}

	var ids []string
	ids, _, err = admS.dm.GetRateProfileRates(ctx, args, true)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		return utils.ErrNotFound
	}
	*countIDs = len(ids)
	return
}

// SetRateProfile add/update a new Rate Profile
func (admS *AdminSv1) SetRateProfile(ctx *context.Context, args *utils.APIRateProfile, reply *string) (err error) {
	if missing := utils.MissingStructFields(args, []string{utils.ID, utils.Rates}); len(missing) != 0 {
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	if args.Tenant == utils.EmptyString {
		args.Tenant = admS.cfg.GeneralCfg().DefaultTenant
	}
	// check if we want to overwrite our profile already existing in database
	var optOverwrite bool
	if _, has := args.APIOpts[utils.MetaRateSOverwrite]; has {
		optOverwrite, err = utils.IfaceAsBool(args.APIOpts[utils.MetaRateSOverwrite])
		if err != nil {
			return
		}
	}
	if err := admS.dm.SetRateProfile(ctx, args.RateProfile, optOverwrite, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), args.Tenant, utils.CacheRateProfiles,
		args.TenantID(), utils.EmptyString, &args.FilterIDs, args.APIOpts); err != nil {
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
	if err := admS.dm.RemoveRateProfileRates(ctx, tnt, args.ID, &args.RateIDs, true); err != nil {
		return utils.APIErrorHandler(err)
	}
	//generate a loadID for CacheRateProfiles and store it in database
	if err := admS.dm.SetLoadIDs(ctx, map[string]int64{utils.CacheRateProfiles: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := admS.CallCache(ctx, utils.IfaceAsString(args.APIOpts[utils.MetaCache]), tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, args.ID), utils.EmptyString, nil, args.APIOpts); err != nil {
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
	if err := admS.CallCache(ctx, utils.IfaceAsString(arg.APIOpts[utils.MetaCache]), tnt, utils.CacheRateProfiles,
		utils.ConcatenatedKey(tnt, arg.ID), utils.EmptyString, nil, arg.APIOpts); err != nil {
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

// RateProfilesForEvent returns the list of RateProfilesIDs that are matching the event
func (rSv1 *RateSv1) RateProfilesForEvent(ctx *context.Context, args *utils.CGREvent, rpIDs *[]string) (err error) {
	return rSv1.rS.V1RateProfilesForEvent(ctx, args, rpIDs)
}

// RateProfilesForEvent returns the list of rates that are matching the event from a specific profile
func (rSv1 *RateSv1) RateProfileRatesForEvent(ctx *context.Context, args *utils.CGREventWithRateProfile, rateIDs *[]string) (err error) {
	return rSv1.rS.V1RateProfileRatesForEvent(ctx, args, rateIDs)
}

// CostForEvent returs the costs for the event and all the rate profile information
func (rSv1 *RateSv1) CostForEvent(ctx *context.Context, args *utils.CGREvent, rpCost *utils.RateProfileCost) (err error) {
	return rSv1.rS.V1CostForEvent(ctx, args, rpCost)
}
