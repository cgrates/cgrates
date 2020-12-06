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

package actions

import (
	"fmt"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// NewActionS instantiates the ActionS
func NewActionS(cfg *config.CGRConfig, fltrS *engine.FilterS, dm *engine.DataManager) *ActionS {
	return &ActionS{
		cfg:   cfg,
		fltrS: fltrS,
		dm:    dm,
	}
}

// ActionS manages exection of Actions
type ActionS struct {
	cfg   *config.CGRConfig
	fltrS *engine.FilterS
	dm    *engine.DataManager
}

// ListenAndServe keeps the service alive
func (aS *ActionS) ListenAndServe(stopChan, cfgRld chan struct{}) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s>",
		utils.CoreS, utils.ActionS))
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
func (aS *ActionS) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown <%s>", utils.CoreS, utils.ActionS))
	return
}

// Call implements rpcclient.ClientConnector interface for internal RPC
func (aS *ActionS) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return utils.RPCCall(aS, serviceMethod, args, reply)
}

// matchingActionProfilesForEvent returns the matched ActionProfiles for the given event
func (aS *ActionS) matchingActionProfilesForEvent(tnt string, aPrflIDs []string,
	cgrEv *utils.CGREventWithOpts) (actPrfls engine.ActionProfiles, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.CGREvent.Event,
		utils.MetaOpts: cgrEv.Opts,
	}
	if len(aPrflIDs) == 0 {
		var aPfIDMp utils.StringSet
		if aPfIDMp, err = engine.MatchingItemIDsForEvent(
			evNm,
			aS.cfg.ActionSCfg().StringIndexedFields,
			aS.cfg.ActionSCfg().PrefixIndexedFields,
			aS.cfg.ActionSCfg().SuffixIndexedFields,
			aS.dm,
			utils.CacheActionProfilesFilterIndexes,
			tnt,
			aS.cfg.ActionSCfg().IndexedSelects,
			aS.cfg.ActionSCfg().NestedFields,
		); err != nil {
			return
		}
		aPrflIDs = aPfIDMp.AsSlice()
	}
	for _, aPfID := range aPrflIDs {
		var aPf *engine.ActionProfile
		if aPf, err = aS.dm.GetActionProfile(tnt, aPfID,
			true, true, utils.NonTransactional); err != nil {
			if err == utils.ErrNotFound {
				err = nil
				continue
			}
			return
		}
		if aPf.ActivationInterval != nil && cgrEv.Time != nil &&
			!aPf.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		var pass bool
		if pass, err = aS.fltrS.Pass(tnt, aPf.FilterIDs, evNm); err != nil {
			return
		} else if !pass {
			continue
		}
		actPrfls = append(actPrfls, aPf)
	}
	if len(actPrfls) == 0 {
		return nil, utils.ErrNotFound
	}
	actPrfls.Sort()
	return
}
