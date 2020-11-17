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
		case <-stopChan: // global exit
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

// Call implements rpcclient.ClientConnector interface for internal RPC
func (rS *RateS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(rS, serviceMethod, args, reply)
}

// matchingRateProfileForEvent returns the matched RateProfile for the given event
func (rS *RateS) matchingRateProfileForEvent(tnt string, rPfIDs []string, args *utils.ArgsCostForEvent) (rtPfl *engine.RateProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  args.CGREvent.Event,
		utils.MetaOpts: args.Opts,
	}
	if len(rPfIDs) == 0 {
		var rPfIDMp utils.StringSet
		if rPfIDMp, err = engine.MatchingItemIDsForEvent(
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
	for _, rPfID := range rPfIDs {
		var rPf *engine.RateProfile
		if rPf, err = rS.dm.GetRateProfile(tnt, rPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if rPf.ActivationInterval != nil && args.CGREvent.Time != nil &&
			!rPf.ActivationInterval.IsActiveAtTime(*args.CGREvent.Time) { // not active
			continue
		}
		var pass bool
		if pass, err = rS.filterS.Pass(tnt, rPf.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		if rtPfl == nil || rtPfl.Weight < rPf.Weight {
			rtPfl = rPf
		}
	}
	if rtPfl == nil {
		return nil, utils.ErrNotFound
	}

	return
}

// costForEvent computes the cost for an event based on a preselected rating profile
func (rS *RateS) rateProfileCostForEvent(rtPfl *engine.RateProfile, args *utils.ArgsCostForEvent) (rpCost *engine.RateProfileCost, err error) {
	var rtIDs utils.StringSet
	if rtIDs, err = engine.MatchingItemIDsForEvent(
		args.CGREventWithOpts.CGREvent.Event,
		rS.cfg.RateSCfg().RateStringIndexedFields,
		rS.cfg.RateSCfg().RatePrefixIndexedFields,
		rS.cfg.RateSCfg().RateSuffixIndexedFields,
		rS.dm,
		utils.CacheRateFilterIndexes,
		utils.ConcatenatedKey(args.CGREventWithOpts.CGREvent.Tenant, rtPfl.ID),
		rS.cfg.RateSCfg().RateIndexedSelects,
		rS.cfg.RateSCfg().RateNestedFields,
	); err != nil {
		return
	}
	aRates := make([]*engine.Rate, 0, len(rtIDs))
	evNm := utils.MapStorage{utils.MetaReq: args.CGREventWithOpts.CGREvent.Event}
	for rtID := range rtIDs {
		rt := rtPfl.Rates[rtID] // pick the rate directly from map based on matched ID
		var pass bool
		if pass, err = rS.filterS.Pass(args.CGREventWithOpts.CGREvent.Tenant, rt.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		aRates = append(aRates, rt)
	}
	var sTime time.Time
	if sTime, err = args.StartTime(rS.cfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	var usage time.Duration
	if usage, err = args.Usage(); err != nil {
		return
	}
	var ordRts []*orderedRate
	if ordRts, err = orderRatesOnIntervals(aRates, sTime, usage, true, 1000000); err != nil {
		return
	}
	rpCost = &engine.RateProfileCost{ID: rtPfl.ID}
	if rpCost.RateSIntervals, err = computeRateSIntervals(ordRts, 0, usage); err != nil {
		return nil, err
	}
	rpCost.Cost = engine.CostForIntervals(rpCost.RateSIntervals).Float64()
	return
}

// V1CostForEvent will be called to calculate the cost for an event
func (rS *RateS) V1CostForEvent(args *utils.ArgsCostForEvent, rpCost *engine.RateProfileCost) (err error) {
	rPfIDs := make([]string, len(args.RateProfileIDs))
	for i, rpID := range args.RateProfileIDs {
		rPfIDs[i] = rpID
	}
	var rtPrl *engine.RateProfile
	if rtPrl, err = rS.matchingRateProfileForEvent(args.CGREventWithOpts.Tenant, rPfIDs, args); err != nil {
		return utils.NewErrServerError(err)
	}
	if rcvCost, errCost := rS.rateProfileCostForEvent(rtPrl, args); errCost != nil {
		return utils.NewErrServerError(errCost)
	} else {
		*rpCost = *rcvCost
	}
	return
}
