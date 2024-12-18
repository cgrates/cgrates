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
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/utils"
)

func NewChargerService(dm *DataManager, filterS *FilterS,
	cfg *config.CGRConfig, connMgr *ConnManager) *ChargerS {
	return &ChargerS{dm: dm, fltrS: filterS,
		cfg: cfg, connMgr: connMgr}
}

// ChargerS is performing charging
type ChargerS struct {
	dm      *DataManager
	fltrS   *FilterS
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

// matchingChargingProfilesForEvent returns ordered list of matching chargers which are active by the time of the function call
func (cS *ChargerS) matchingChargerProfilesForEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent) (cPs ChargerProfiles, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	cpIDs, err := MatchingItemIDsForEvent(ctx, evNm,
		cS.cfg.ChargerSCfg().StringIndexedFields,
		cS.cfg.ChargerSCfg().PrefixIndexedFields,
		cS.cfg.ChargerSCfg().SuffixIndexedFields,
		cS.cfg.ChargerSCfg().ExistsIndexedFields,
		cS.cfg.ChargerSCfg().NotExistsIndexedFields,
		cS.dm, utils.CacheChargerFilterIndexes, tnt,
		cS.cfg.ChargerSCfg().IndexedSelects,
		cS.cfg.ChargerSCfg().NestedFields,
	)
	if err != nil {
		return nil, err
	}
	for cpID := range cpIDs {
		cP, err := cS.dm.GetChargerProfile(ctx, tnt, cpID, true, true, utils.NonTransactional)
		if err != nil {
			if err == utils.ErrNotFound {
				continue
			}
			return nil, err
		}
		if pass, err := cS.fltrS.Pass(ctx, tnt, cP.FilterIDs,
			evNm); err != nil {
			return nil, err
		} else if !pass {
			continue
		}
		if cP.weight, err = WeightFromDynamics(ctx, cP.Weights, cS.fltrS, tnt, evNm); err != nil {
			return nil, err
		}
		cPs = append(cPs, cP)
	}
	if len(cPs) == 0 {
		return nil, utils.ErrNotFound
	}
	cPs.Sort()
	for i, cp := range cPs {
		var blocker bool
		if blocker, err = BlockerFromDynamics(ctx, cp.Blockers, cS.fltrS, tnt, evNm); err != nil {
			return
		}
		if blocker {
			cPs = cPs[0 : i+1]
			break
		}
	}
	return
}

// ChrgSProcessEventReply is the reply to processEvent
type ChrgSProcessEventReply struct {
	ChargerSProfile string
	AlteredFields   []*FieldsAltered
	CGREvent        *utils.CGREvent
}

func (cS *ChargerS) processEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent) (rply []*ChrgSProcessEventReply, err error) {
	var cPs ChargerProfiles
	if cPs, err = cS.matchingChargerProfilesForEvent(ctx, tnt, cgrEv); err != nil {
		return nil, err
	}

	rply = make([]*ChrgSProcessEventReply, len(cPs))
	for i, cP := range cPs {
		clonedEv := cgrEv.Clone()
		clonedEv.Tenant = tnt
		clonedEv.APIOpts[utils.MetaRunID] = cP.RunID
		clonedEv.APIOpts[utils.MetaSubsys] = utils.MetaChargers
		clonedEv.APIOpts[utils.MetaChargeID] = utils.Sha1(utils.IfaceAsString(clonedEv.APIOpts[utils.MetaOriginID]), cP.RunID)
		rply[i] = &ChrgSProcessEventReply{
			ChargerSProfile: cP.ID,
			CGREvent:        clonedEv,
			AlteredFields: []*FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           []string{utils.MetaOptsRunID, utils.MetaOpts + utils.NestingSep + utils.MetaChargeID, utils.MetaOpts + utils.NestingSep + utils.MetaSubsys},
				},
			},
		}
		if len(cP.AttributeIDs) == 1 && cP.AttributeIDs[0] == utils.MetaNone {
			continue // AttributeS disabled
		}
		clonedEv.APIOpts[utils.OptsContext] = utils.FirstNonEmpty(
			utils.IfaceAsString(clonedEv.APIOpts[utils.OptsContext]),
			utils.MetaChargers)
		clonedEv.APIOpts[utils.OptsAttributesProfileIDs] = cP.AttributeIDs
		var evReply AttrSProcessEventReply
		if err = cS.connMgr.Call(ctx, cS.cfg.ChargerSCfg().AttributeSConns,
			utils.AttributeSv1ProcessEvent, clonedEv, &evReply); err != nil {
			if err.Error() != utils.ErrNotFound.Error() {
				return nil, err
			}
			err = nil
		}
		if evReply.AlteredFields != nil {
			rply[i].AlteredFields = append(rply[i].AlteredFields, evReply.AlteredFields...)
			rply[i].CGREvent = evReply.CGREvent
		}
	}
	return
}

// V1ProcessEvent will process the event received via API and return list of events forked
func (cS *ChargerS) V1ProcessEvent(ctx *context.Context, args *utils.CGREvent,
	reply *[]*ChrgSProcessEventReply) (err error) {
	if args == nil ||
		args.Event == nil {
		return utils.NewErrMandatoryIeMissing("Event")
	}
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	rply, err := cS.processEvent(ctx, tnt, args)
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
func (cS *ChargerS) V1GetChargersForEvent(ctx *context.Context, args *utils.CGREvent,
	rply *ChargerProfiles) (err error) {
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = cS.cfg.GeneralCfg().DefaultTenant
	}
	cPs, err := cS.matchingChargerProfilesForEvent(ctx, tnt, args)
	if err != nil {
		if err != utils.ErrNotFound {
			err = utils.NewErrServerError(err)
		}
		return err
	}
	*rply = cPs
	return
}
