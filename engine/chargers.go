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

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewChargerService(dm *DataManager, filterS *FilterS,
	cfg *config.CGRConfig, connMgr *ConnManager) (*ChargerService, error) {

	return &ChargerService{dm: dm, filterS: filterS,
		cfg: cfg, connMgr: connMgr}, nil
}

// ChargerService is performing charging
type ChargerService struct {
	dm      *DataManager
	filterS *FilterS
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

// ListenAndServe will initialize the service
func (cS *ChargerService) ListenAndServe(exitChan chan bool) (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> starting <%s> subsystem", utils.CoreS, utils.ChargerS))
	e := <-exitChan
	exitChan <- e
	return
}

// Shutdown is called to shutdown the service
func (cS *ChargerService) Shutdown() (err error) {
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown initialized", utils.ChargerS))
	utils.Logger.Info(fmt.Sprintf("<%s> shutdown complete", utils.ChargerS))
	return
}

// matchingChargingProfilesForEvent returns ordered list of matching chargers which are active by the time of the function call
func (cS *ChargerService) matchingChargerProfilesForEvent(cgrEv *utils.CGREventWithOpts) (cPs ChargerProfiles, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.Opts,
	}
	cpIDs, err := MatchingItemIDsForEvent(evNm,
		cS.cfg.ChargerSCfg().StringIndexedFields,
		cS.cfg.ChargerSCfg().PrefixIndexedFields,
		cS.dm, utils.CacheChargerFilterIndexes, cgrEv.Tenant,
		cS.cfg.ChargerSCfg().IndexedSelects,
		cS.cfg.ChargerSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	matchingCPs := make(map[string]*ChargerProfile)
	for cpID := range cpIDs {
		cP, err := cS.dm.GetChargerProfile(cgrEv.Tenant, cpID, true, true, utils.NonTransactional)
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
		if pass, err := cS.filterS.Pass(cgrEv.Tenant, cP.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		matchingCPs[cpID] = cP
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
	Opts               map[string]interface{}
}

func (cS *ChargerService) processEvent(cgrEv *utils.CGREventWithOpts) (rply []*ChrgSProcessEventReply, err error) {
	var cPs ChargerProfiles
	if cgrEv.Opts == nil {
		cgrEv.Opts = make(map[string]interface{})
	}
	cgrEv.Opts[utils.Subsys] = utils.MetaChargers
	if cPs, err = cS.matchingChargerProfilesForEvent(cgrEv); err != nil {
		return nil, err
	}
	rply = make([]*ChrgSProcessEventReply, len(cPs))
	for i, cP := range cPs {
		clonedEv := cgrEv.Clone()
		opts := MapEvent(cgrEv.Opts).Clone()
		clonedEv.Event[utils.RunID] = cP.RunID
		rply[i] = &ChrgSProcessEventReply{
			ChargerSProfile: cP.ID,
			CGREvent:        clonedEv.CGREvent,
			AlteredFields:   []string{utils.MetaReqRunID},
			Opts:            opts,
		}
		if len(cP.AttributeIDs) == 1 && cP.AttributeIDs[0] == utils.META_NONE {
			continue // AttributeS disabled
		}

		args := &AttrArgsProcessEvent{
			AttributeIDs: cP.AttributeIDs,
			Context: utils.StringPointer(utils.FirstNonEmpty(
				utils.IfaceAsString(opts[utils.OptsContext]),
				utils.MetaChargers)),
			ProcessRuns:   nil,
			CGREvent:      clonedEv.CGREvent,
			ArgDispatcher: clonedEv.ArgDispatcher,
			Opts:          opts,
		}
		var evReply AttrSProcessEventReply
		if err = cS.connMgr.Call(cS.cfg.ChargerSCfg().AttributeSConns, nil,
			utils.AttributeSv1ProcessEvent, args, &evReply); err != nil {
			return nil, err
		}
		rply[i].AttributeSProfiles = evReply.MatchedProfiles
		if len(evReply.AlteredFields) != 0 {
			rply[i].AlteredFields = append(rply[i].AlteredFields, evReply.AlteredFields...)
			rply[i].CGREvent = evReply.CGREvent
			rply[i].Opts = evReply.Opts
		}
	}
	return
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerService) V1ProcessEvent(args *utils.CGREventWithOpts,
	reply *[]*ChrgSProcessEventReply) (err error) {
	if args.CGREvent == nil ||
		args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	rply, err := cS.processEvent(args)
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
func (cS *ChargerService) V1GetChargersForEvent(args *utils.CGREventWithOpts,
	rply *ChargerProfiles) (err error) {
	cPs, err := cS.matchingChargerProfilesForEvent(args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*rply = cPs
	return
}
