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

package engine

import (
	"fmt"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewChargerService(dm *DataManager, filterS *FilterS,
	cfg *config.CGRConfig, connMgr *ConnManager) *ChargerService {
	return &ChargerService{dm: dm, filterS: filterS,
		cfg: cfg, connMgr: connMgr}
}

// ChargerService is performing charging
type ChargerService struct {
	dm      *DataManager
	filterS *FilterS
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

// Shutdown is called to shutdown the service
func (cS *ChargerService) Shutdown() {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.ChargerS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.ChargerS))
}

// matchingChargingProfilesForEvent returns ordered list of matching chargers which are active by the time of the function call
func (cS *ChargerService) matchingChargerProfilesForEvent(tnt string, cgrEv *utils.CGREvent) (cPs ChargerProfiles, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	cpIDs, err := MatchingItemIDsForEvent(context.TODO(), evNm,
		cS.cfg.ChargerSCfg().StringIndexedFields,
		cS.cfg.ChargerSCfg().PrefixIndexedFields,
		cS.cfg.ChargerSCfg().SuffixIndexedFields,
		cS.dm, utils.CacheChargerFilterIndexes, tnt,
		cS.cfg.ChargerSCfg().IndexedSelects,
		cS.cfg.ChargerSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	matchingCPs := make(map[string]*ChargerProfile)
	for cpID := range cpIDs {
		cP, err := cS.dm.GetChargerProfile(tnt, cpID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if cP.ActivationInterval != nil && cgrEv.Time != nil &&
			!cP.ActivationInterval.IsActiveAtTime(*cgrEv.Time) { // not active
			continue
		}
		if pass, err := cS.filterS.Pass(context.TODO(), tnt, cP.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingCPs[cpID] = cP
	}
	if len(matchingCPs) == 0 {
		return nil, utils.ErrNotFound
	}
	cPs = make(ChargerProfiles, len(matchingCPs))
	i := 0
	for _, cP := range matchingCPs {
		cPs[i] = cP
		i++
	}
	cPs.Sort()
	return
}

// ChrgSProcessEventReply is the reply to processEvent
type ChrgSProcessEventReply struct {
	ChargerSProfile    string
	AttributeSProfiles []string
	AlteredFields      []string
	CGREvent           *utils.CGREvent
}

func (cS *ChargerService) processEvent(tnt string, cgrEv *utils.CGREvent) (rply []*ChrgSProcessEventReply, err error) {
	var cPs ChargerProfiles
	var processRuns *int
	if val, has := cgrEv.APIOpts[utils.OptsAttributesProcessRuns]; has {
		if v, err := utils.IfaceAsTInt64(val); err == nil {
			processRuns = utils.IntPointer(int(v))
		}
	}
	if cPs, err = cS.matchingChargerProfilesForEvent(tnt, cgrEv); err != nil {
		return nil, err
	}
	rply = make([]*ChrgSProcessEventReply, len(cPs))
	for i, cP := range cPs {
		clonedEv := cgrEv.Clone()
		clonedEv.Tenant = tnt
		clonedEv.Event[utils.RunID] = cP.RunID
		clonedEv.APIOpts[utils.Subsys] = utils.MetaChargers
		rply[i] = &ChrgSProcessEventReply{
			ChargerSProfile: cP.ID,
			CGREvent:        clonedEv,
			AlteredFields:   []string{utils.MetaReqRunID},
		}
		if len(cP.AttributeIDs) == 1 && cP.AttributeIDs[0] == utils.MetaNone {
			continue // AttributeS disabled
		}

		args := &AttrArgsProcessEvent{
			AttributeIDs: cP.AttributeIDs,
			Context: utils.StringPointer(utils.FirstNonEmpty(
				utils.IfaceAsString(clonedEv.APIOpts[utils.OptsContext]),
				utils.MetaChargers)),
			ProcessRuns: processRuns,
			CGREvent:    clonedEv,
		}
		var evReply AttrSProcessEventReply
		if err = cS.connMgr.Call(context.TODO(), cS.cfg.ChargerSCfg().AttributeSConns,
			utils.AttributeSv1ProcessEvent, args, &evReply); err != nil {
			if err.Error() != utils.ErrNotFound.Error() {
				return nil, err
			}
			err = nil
		}
		rply[i].AttributeSProfiles = evReply.MatchedProfiles
		if len(evReply.AlteredFields) != 0 {
			rply[i].AlteredFields = append(rply[i].AlteredFields, evReply.AlteredFields...)
			rply[i].CGREvent = evReply.CGREvent
		}
	}
	return
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerService) V1ProcessEvent(args *utils.CGREvent,
	reply *[]*ChrgSProcessEventReply) (err error) {
	if args == nil ||
		args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	rply, err := cS.processEvent(tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*reply = rply
	return
}

// V1GetChargersForEvent exposes the list of ordered matching ChargingProfiles for an event
func (cS *ChargerService) V1GetChargersForEvent(args *utils.CGREvent,
	rply *ChargerProfiles) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	cPs, err := cS.matchingChargerProfilesForEvent(tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*rply = cPs
	return
}
