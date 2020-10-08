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
func (rS *RateS) ListenAndServe(exitChan chan bool, cfgRld chan struct{}) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.RateS))
	for {
		select {
		case e := <-exitChan: // global exit
			rS.Shutdown()
			exitChan <- e // put back for the others listening for shutdown request
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
func (rS *RateS) matchingRateProfileForEvent(args *ArgsCostForEvent, rPfIDs []string) (rtPfl *engine.RateProfile, err error) {
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
			args.CGREvent.Tenant,
			rS.cfg.RateSCfg().IndexedSelects,
			rS.cfg.RateSCfg().NestedFields,
		); err != nil {
			return
		}
		rPfIDs = rPfIDMp.AsSlice()
	}
	var sTime time.Time
	if sTime, err = args.StartTime(rS.cfg.GeneralCfg().DefaultTimezone); err != nil {
		return
	}
	for _, rPfID := range rPfIDs {
		var rPf *engine.RateProfile
		if rPf, err = rS.dm.GetRateProfile(args.CGREvent.Tenant, rPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if rPf.ActivationInterval != nil &&
			!rPf.ActivationInterval.IsActiveAtTime(sTime) { // not active
			continue
		}
		var pass bool
		if pass, err = rS.filterS.Pass(args.CGREvent.Tenant, rPf.FilterIDs, evNm); err != nil {
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

/*
// costForEvent computes the cost for an event based on a preselected rating profile
func (rS *RateS) rateProfileCostForEvent(rtPfl *engine.RateProfile, args *ArgsCostForEvent) (rts []*engine.RateSInterval, err error) {
	var rtIDs utils.StringSet
	if rtIDs, err = engine.MatchingItemIDsForEvent(
		args.CGRevent.Event,
		rS.cfg.RateSCfg().RateStringIndexedFields,
		rS.cfg.RateSCfg().RatePrefixIndexedFields,
		rS.dm,
		utils.CacheRateFilterIndexes,
		utils.ConcatenatedKey(args.CGRevent.Tenant, rtPfl.ID),
		rS.cfg.RateSCfg().RateIndexedSelects,
		rS.cfg.RateSCfg().RateNestedFields,
	); err != nil {
		return
	}
	aRates := make([]*engine.Rate, len(rtIDs))
	evNm := utils.MapStorage{utils.MetaReq: cgrEv.Event}
	for rtID := range rtIDs {
		rt := rtPfl.Rates[rtID] // pick the rate directly from map based on matched ID
		var pass bool
		if pass, err = rS.filterS.Pass(cgrEv.Tenant, rt.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		aRates = append(aRates, rt)
	}
	ordRts := orderRatesOnIntervals(aRates)
	return
}
*/

// ArgsCostForEvent arguments used for proccess event
type ArgsCostForEvent struct {
	RateProfileIDs []string
	Opts           map[string]interface{}
	*utils.CGREventWithOpts
}

// StartTime returns the event time used to check active rate profiles
func (args *ArgsCostForEvent) StartTime(tmz string) (sTime time.Time, err error) {
	if tIface, has := args.Opts[utils.OptsRatesStartTime]; has {
		return utils.IfaceAsTime(tIface, tmz)
	}
	if sTime, err = args.CGREvent.FieldAsTime(utils.AnswerTime, tmz); err != nil {
		if err != utils.ErrNotFound {
			return
		}
		// not found, try SetupTime
		if sTime, err = args.CGREvent.FieldAsTime(utils.SetupTime, tmz); err != nil &&
			err != utils.ErrNotFound {
			return
		}
	}
	if err == nil {
		return
	}
	if args.CGREvent.Time != nil {
		return *args.CGREvent.Time, nil
	}
	return time.Now(), nil
}

// V1CostForEvent will be called to calculate the cost for an event
func (rS *RateS) V1CostForEvent(args *ArgsCostForEvent, cC *utils.ChargedCost) (err error) {
	return
}
