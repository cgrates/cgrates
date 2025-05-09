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
package rates

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// V1RateProfilesForEvent will be called to list the RateProfilesIDs that are matching the event
func (rS *RateS) V1RateProfilesForEvent(ctx *context.Context, args *utils.CGREvent, rpIDs *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}
	evNm := args.AsDataProvider()
	var ratePrfs utils.StringSet
	if ratePrfs, err = engine.MatchingItemIDsForEvent(ctx,
		evNm,
		rS.cfg.RateSCfg().StringIndexedFields,
		rS.cfg.RateSCfg().PrefixIndexedFields,
		rS.cfg.RateSCfg().SuffixIndexedFields,
		rS.cfg.RateSCfg().ExistsIndexedFields,
		rS.cfg.RateSCfg().NotExistsIndexedFields,
		rS.dm,
		utils.CacheRateProfilesFilterIndexes,
		tnt,
		rS.cfg.RateSCfg().IndexedSelects,
		rS.cfg.RateSCfg().NestedFields,
	); err != nil {
		return
	}
	if len(ratePrfs) == 0 {
		return utils.ErrNotFound
	}
	var profilesMtched []string
	for _, ratePrf := range ratePrfs.AsSlice() {
		var rp *utils.RateProfile
		if rp, err = rS.dm.GetRateProfile(ctx, tnt, ratePrf, true, true, utils.NonTransactional); err != nil {
			return
		}
		var pass bool
		if pass, err = rS.fltrS.Pass(ctx, tnt, rp.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		profilesMtched = append(profilesMtched, ratePrf)
	}
	if len(profilesMtched) == 0 {
		return utils.ErrNotFound
	}
	*rpIDs = profilesMtched
	return
}

// RateProfilesForEvent returns the list of rates that are matching the event from a specific profile
func (rS *RateS) V1RateProfileRatesForEvent(ctx *context.Context, args *utils.CGREventWithRateProfile, rtIDs *[]string) (err error) {
	if args == nil {
		return utils.NewErrMandatoryIeMissing(utils.CGREventString)
	}
	if args.RateProfileID == utils.EmptyString {
		return utils.NewErrMandatoryIeMissing(utils.RateProfileID)
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = rS.cfg.GeneralCfg().DefaultTenant
	}
	evNM := args.AsDataProvider()
	var rateIDs utils.StringSet
	if rateIDs, err = engine.MatchingItemIDsForEvent(ctx,
		evNM,
		rS.cfg.RateSCfg().RateStringIndexedFields,
		rS.cfg.RateSCfg().RatePrefixIndexedFields,
		rS.cfg.RateSCfg().RateSuffixIndexedFields,
		rS.cfg.RateSCfg().RateExistsIndexedFields,
		rS.cfg.RateSCfg().RateNotExistsIndexedFields,
		rS.dm,
		utils.CacheRateFilterIndexes,
		utils.ConcatenatedKey(tnt, args.RateProfileID),
		rS.cfg.RateSCfg().RateIndexedSelects,
		rS.cfg.RateSCfg().RateNestedFields,
	); err != nil {
		return
	}
	if len(rateIDs) == 0 {
		return utils.ErrNotFound
	}
	var ratesMtched []string
	for _, rateID := range rateIDs.AsSlice() {
		var rp *utils.RateProfile
		if rp, err = rS.dm.GetRateProfile(ctx, tnt, args.RateProfileID, true, true, utils.NonTransactional); err != nil {
			return
		}
		rate := rp.Rates[rateID]
		var pass bool
		if pass, err = rS.fltrS.Pass(ctx, tnt, rate.FilterIDs, evNM); err != nil {
			return
		} else if !pass {
			continue
		}
		ratesMtched = append(ratesMtched, rateID)
	}
	if len(ratesMtched) == 0 {
		return utils.ErrNotFound
	}
	*rtIDs = ratesMtched
	return
}

// V1CostForEvent calculates the cost for an event using matching rate
// profiles. If a higher priority profile fails, it tries the next matching
// profile. This continues until a valid cost is found or all profiles are
// exhausted.
func (rS *RateS) V1CostForEvent(ctx *context.Context, args *utils.CGREvent, rpCost *utils.RateProfileCost) (err error) {
	var rPfIDs []string
	if rPfIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.RateSCfg().Opts.ProfileIDs,
		config.RatesProfileIDsDftOpt, utils.OptsRatesProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), nil, rS.fltrS, rS.cfg.RateSCfg().Opts.ProfileIgnoreFilters,
		utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	ignoredRPfIDs := utils.NewStringSet([]string{})
	var firstError error
	for i := range rS.cfg.RateSCfg().Verbosity {
		var rtPrl *utils.RateProfile
		if rtPrl, err = rS.matchingRateProfileForEvent(ctx, args.Tenant, rPfIDs, args, ignFilters, ignoredRPfIDs); err != nil {
			if err != utils.ErrNotFound {
				err = utils.NewErrServerError(err)
			} else if i != 0 { // no more fallback rating profiles, return the original error
				err = utils.NewErrServerError(firstError)
			}
			return
		}
		var rcvCost *utils.RateProfileCost
		if rcvCost, err = rS.rateProfileCostForEvent(ctx, rtPrl, args, rS.cfg.RateSCfg().Verbosity); err != nil {
			if err != utils.ErrNotFound {
				//err = utils.NewErrServerError(err)
				if i == 0 {
					firstError = err
				}
				ignoredRPfIDs.Add(rtPrl.ID)
				continue // no cost, go to the next matching RatingProfile
			}
			return
		}
		*rpCost = *rcvCost
		return
	}
	return
}
