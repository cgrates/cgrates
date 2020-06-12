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
	"sort"

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
func (rS *RateS) matchingRateProfileForEvent(cgrEv *utils.CGREvent) (rtPfl *engine.RateProfile, err error) {
	var rPfIDs utils.StringMap
	if rPfIDs, err = engine.MatchingItemIDsForEvent(
		cgrEv.Event,
		rS.cfg.RateSCfg().StringIndexedFields,
		rS.cfg.RateSCfg().PrefixIndexedFields,
		rS.dm, utils.CacheRateProfilesFilterIndexes,
		cgrEv.Tenant,
		rS.cfg.RouteSCfg().IndexedSelects,
		rS.cfg.RouteSCfg().NestedFields,
	); err != nil {
		return
	}
	var matchingRPfs []*engine.RateProfile
	evNm := utils.MapStorage{utils.MetaReq: cgrEv.Event}
	for rPfID := range rPfIDs {
		var rPf *engine.RateProfile
		if rPf, err = rS.dm.GetRateProfile(cgrEv.Tenant, rPfID, true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if rPf.ActivationInterval != nil && cgrEv.Time != nil &&
			!rPf.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		var pass bool
		if pass, err = rS.filterS.Pass(cgrEv.Tenant, rPf.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}

		matchingRPfs = append(matchingRPfs, rPf)
	}
	sort.Slice(matchingRPfs, func(i, j int) bool { return matchingRPfs[i].Weight > matchingRPfs[j].Weight })
	return
}

// V1CostForEvent will be called to calculate the cost for an event
func (rS *RateS) V1CostForEvent(cgrEv *utils.CGREventWithOpts, cC *utils.ChargedCost) (err error) {
	return
}
