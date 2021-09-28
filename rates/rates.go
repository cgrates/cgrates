/*
Real-time Online/Offline Charging System (OerS) for Telecom & ISP environments
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
		cfg:     cfg,
		filterS: filterS,
		dm:      dm,
	}
}

// RateS calculates costs for events
type RateS struct {
	cfg     *config.CGRConfig
	filterS *engine.FilterS
	dm      *engine.DataManager
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
func (rS *RateS) matchingRateProfileForEvent(ctx *context.Context, tnt string, rPfIDs []string, args *utils.CGREvent) (rtPfl *utils.RateProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.Event,
		utils.MetaOpts: args.APIOpts,
	}
	if len(rPfIDs) == 0 {
		var rPfIDMp utils.StringSet
		if rPfIDMp, err = engine.MatchingItemIDsForEvent(ctx,
			evNm,
			rS.cfg.RateSCfg().StringIndexedFields,
			rS.cfg.RateSCfg().PrefixIndexedFields,
			rS.cfg.RateSCfg().SuffixIndexedFields,
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
		var pass bool
		if pass, err = rS.filterS.Pass(ctx, tnt, rPf.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		var rPfWeight float64
		if rPfWeight, err = engine.WeightFromDynamics(ctx, rPf.Weights,
			rS.filterS, tnt, evNm); err != nil {
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
		if pass, err = rS.filterS.Pass(ctx, args.Tenant, rt.FilterIDs, evNm); err != nil {
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
			rS.filterS, args.Tenant, evNm); err != nil {
			return
		}
	}
	var sTime time.Time
	if sTime, err = args.StartTime(rS.cfg.RateSCfg().Opts.StartTime, rS.cfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	var usage *decimal.Big
	if usage, err = args.OptsAsDecimal(rS.cfg.RateSCfg().Opts.Usage, utils.OptsRatesUsage, utils.MetaUsage); err != nil {
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
	if ivalStart, err = args.OptsAsDecimal(rS.cfg.RateSCfg().Opts.IntervalStart, utils.OptsRatesIntervalStart); err != nil {
		return
	}
	var costIntervals []*utils.RateSInterval
	if costIntervals, err = computeRateSIntervals(ordRts, ivalStart, usage, rpCost.Rates); err != nil {
		return nil, err
	}
	rpCost.CostIntervals = make([]*utils.RateSIntervalCost, len(costIntervals))
	finalCost := new(decimal.Big)
	for idx, costInt := range costIntervals {
		finalCost = utils.SumBig(finalCost, costInt.Cost(rpCost.Rates)) // sums the costs for all intervals
		rpCost.CostIntervals[idx] = costInt.AsRatesIntervalsCost()      //this does not contains IncrementStart and IntervalStart so we convert in RatesIntervalCosts
	}
	rpCost.Cost = &utils.Decimal{finalCost}
	return
}

// V1CostForEvent will be called to calculate the cost for an event
func (rS *RateS) V1CostForEvent(ctx *context.Context, args *utils.CGREvent, rpCost *utils.RateProfileCost) (err error) {
	var rPfIDs []string
	if rPfIDs, err = engine.FilterStringSliceCfgOpts(ctx, args.Tenant, args.AsDataProvider(), rS.filterS,
		rS.cfg.RateSCfg().Opts.RateProfileIDs); err != nil {
		return
	}
	if rPfIDs, err = args.OptsAsStringSlice(rPfIDs, utils.OptsRatesRateProfileIDs); err != nil {
		return
	}
	var rtPrl *utils.RateProfile
	if rtPrl, err = rS.matchingRateProfileForEvent(ctx, args.Tenant, rPfIDs, args); err != nil {
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
