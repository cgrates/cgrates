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
	"fmt"
	"time"

	"github.com/ericlagergren/decimal"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewRateS instantiates the RateS
func NewRateS(cfg *config.CGRConfig, filterS *engine.FilterS, dm *engine.DataManager) *RateS {
	return &RateS{
		cfg:   cfg,
		fltrS: filterS,
		dm:    dm,
	}
}

// RateS calculates costs for events
type RateS struct {
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	dm    *engine.DataManager
}

// ListenAndServe keeps the service alive
func (rS *RateS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.RateS))
	for {
		select {
		case <-stopChan:
			return
		case rld := <-cfgRld: // configuration was reloaded
			cfgRld <- rld
		}
	}
}

// Shutdown is called to shutdown the service
func (rS *RateS) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.RateS))
	return
}

// matchingRateProfileForEvent returns the matched RateProfile for the given event
func (rS *RateS) matchingRateProfileForEvent(ctx *context.Context, tnt string, rPfIDs []string, args *utils.CGREvent, ignoreFilters bool) (rtPfl *utils.RateProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	if len(rPfIDs) == 0 {
		ignoreFilters = false
		var rPfIDMp utils.StringSet
		if rPfIDMp, err = engine.MatchingItemIDsForEvent(ctx,
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
		rPfIDs = rPfIDMp.AsSlice()
	}
	var rpWw *rpWithWeight
	for _, rPfID := range rPfIDs {
		var rPf *utils.RateProfile
		if rPf, err = rS.dm.GetRateProfile(ctx, tnt, rPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if !ignoreFilters {
			var pass bool
			if pass, err = rS.fltrS.Pass(ctx, tnt, rPf.FilterIDs, evNm); err != nil {
				return
			} else if !pass {
				continue
			}
		}
		var rPfWeight float64
		if rPfWeight, err = engine.WeightFromDynamics(ctx, rPf.Weights,
			rS.fltrS, tnt, evNm); err != nil {
			return
		}
		if rpWw == nil || rpWw.weight < rPfWeight {
			rpWw = &rpWithWeight{rPf, rPfWeight}
		}
	}
	if rpWw == nil {
		return nil, utils.ErrNotFound
	}

	return rpWw.RateProfile, nil
}

// rateProfileCostForEvent computes the rateProfileCost for an event based on a preselected rate profile
func (rS *RateS) rateProfileCostForEvent(ctx *context.Context, rtPfl *utils.RateProfile, args *utils.CGREvent, verbosity int) (rpCost *utils.RateProfileCost, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	rtIDs := utils.StringSet{}
	if rS.cfg.RateSCfg().RateIndexedSelects {
		if rtIDs, err = engine.MatchingItemIDsForEvent(ctx,
			evNm,
			rS.cfg.RateSCfg().RateStringIndexedFields,
			rS.cfg.RateSCfg().RatePrefixIndexedFields,
			rS.cfg.RateSCfg().RateSuffixIndexedFields,
			rS.cfg.RateSCfg().RateExistsIndexedFields,
			rS.cfg.RateSCfg().RateNotExistsIndexedFields,
			rS.dm,
			utils.CacheRateFilterIndexes,
			utils.ConcatenatedKey(args.Tenant, rtPfl.ID),
			rS.cfg.RateSCfg().RateIndexedSelects,
			rS.cfg.RateSCfg().RateNestedFields,
		); err != nil {
			return
		}
	} else {
		for id := range rtPfl.Rates {
			rtIDs.Add(id)
		}
	}
	aRates := make([]*utils.Rate, 0, len(rtIDs))
	for rtID := range rtIDs {
		rt := rtPfl.Rates[rtID] // pick the rate directly from map based on matched ID
		var pass bool
		if pass, err = rS.fltrS.Pass(ctx, args.Tenant, rt.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		aRates = append(aRates, rt)
	}
	// populate weights to be used on ordering
	wghts := make([]float64, len(aRates))
	for i, aRt := range aRates {
		if wghts[i], err = engine.WeightFromDynamics(ctx, aRt.Weights,
			rS.fltrS, args.Tenant, evNm); err != nil {
			return
		}
	}
	var sTime time.Time
	if sTime, err = engine.GetTimeOpts(ctx, args.Tenant, args, rS.fltrS, rS.cfg.RateSCfg().Opts.StartTime,
		rS.cfg.GeneralCfg().DefaultTimezone, config.RatesStartTimeDftOpt, utils.OptsRatesStartTime, utils.MetaStartTime); err != nil {
		return
	}
	var usage *decimal.Big
	if usage, err = engine.GetDecimalBigOpts(ctx, args.Tenant, args, rS.fltrS, rS.cfg.RateSCfg().Opts.Usage,
		config.RatesUsageDftOpt, utils.OptsRatesUsage, utils.MetaUsage); err != nil {
		return
	}
	var ordRts []*orderedRate
	if ordRts, err = orderRatesOnIntervals(aRates, wghts, sTime, usage, true, verbosity); err != nil {
		return
	}
	rpCost = &utils.RateProfileCost{
		ID:    rtPfl.ID,
		Rates: make(map[string]*utils.IntervalRate),
	}
	if rtPfl.MinCost != nil {
		rpCost.MinCost = rtPfl.MinCost
	}
	if rtPfl.MaxCost != nil {
		rpCost.MaxCost = rtPfl.MaxCost
	}
	var ivalStart *decimal.Big
	if ivalStart, err = engine.GetDecimalBigOpts(ctx, args.Tenant, args, rS.fltrS, rS.cfg.RateSCfg().Opts.IntervalStart,
		config.RatesIntervalStartDftOpt, utils.OptsRatesIntervalStart); err != nil {
		return
	}
	var costIntervals []*utils.RateSInterval
	if costIntervals, err = computeRateSIntervals(ordRts, ivalStart, usage, rpCost.Rates); err != nil {
		return nil, err
	}
	rpCost.CostIntervals = make([]*utils.RateSIntervalCost, len(costIntervals))
	finalCost := decimal.WithContext(utils.DecimalContext)
	for idx, costInt := range costIntervals {
		finalCost = utils.SumBig(finalCost, costInt.Cost(rpCost.Rates)) // sums the costs for all intervals
		rpCost.CostIntervals[idx] = costInt.AsRatesIntervalsCost()      //this does not contains IncrementStart and IntervalStart so we convert in RatesIntervalCosts
	}
	rpCost.Cost = &utils.Decimal{finalCost}
	return
}

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

// V1CostForEvent will be called to calculate the cost for an event
func (rS *RateS) V1CostForEvent(ctx *context.Context, args *utils.CGREvent, rpCost *utils.RateProfileCost) (err error) {
	var rPfIDs []string
	if rPfIDs, err = engine.GetStringSliceOpts(ctx, args.Tenant, args, rS.fltrS, rS.cfg.RateSCfg().Opts.ProfileIDs,
		config.RatesProfileIDsDftOpt, utils.OptsRatesProfileIDs); err != nil {
		return
	}
	var ignFilters bool
	if ignFilters, err = engine.GetBoolOpts(ctx, args.Tenant, args.AsDataProvider(), rS.fltrS, rS.cfg.RateSCfg().Opts.ProfileIgnoreFilters,
		config.RatesProfileIgnoreFiltersDftOpt, utils.MetaProfileIgnoreFilters); err != nil {
		return
	}
	var rtPrl *utils.RateProfile
	if rtPrl, err = rS.matchingRateProfileForEvent(ctx, args.Tenant, rPfIDs, args, ignFilters); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	var rcvCost *utils.RateProfileCost
	if rcvCost, err = rS.rateProfileCostForEvent(ctx, rtPrl, args, rS.cfg.RateSCfg().Verbosity); err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return
	}
	*rpCost = *rcvCost
	return
}
