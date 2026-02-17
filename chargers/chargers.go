/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>
*/

package chargers

import (
	"cmp"
	"slices"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/attributes"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

// ChargerS is performing charging
type ChargerS struct {
	dm      *engine.DataManager
	fltrS   *engine.FilterS
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
}

func NewChargerService(dm *engine.DataManager, filterS *engine.FilterS,
	cfg *config.CGRConfig, connMgr *engine.ConnManager) *ChargerS {
	return &ChargerS{dm: dm, fltrS: filterS,
		cfg: cfg, connMgr: connMgr}
}

// matchingChargingProfilesForEvent returns ordered list of matching chargers which are active by the time of the function call
func (cS *ChargerS) matchingChargerProfilesForEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent) (cPs []*utils.ChargerProfile, err error) {
	evNm := utils.MapStorage{
		utils.MetaReq:  cgrEv.Event,
		utils.MetaOpts: cgrEv.APIOpts,
	}
	cpIDs, err := engine.MatchingItemIDsForEvent(ctx, evNm,
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

	weights := make(map[string]float64) // stores sorting weights by profile ID
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
		weight, err := engine.WeightFromDynamics(ctx, cP.Weights, cS.fltrS, tnt, evNm)
		if err != nil {
			return nil, err
		}
		weights[cP.ID] = weight
		cPs = append(cPs, cP)
	}
	if len(cPs) == 0 {
		return nil, utils.ErrNotFound
	}

	// Sort by weight (higher values first).
	slices.SortFunc(cPs, func(a, b *utils.ChargerProfile) int {
		return cmp.Compare(weights[b.ID], weights[a.ID])
	})

	for i, cp := range cPs {
		var blocker bool
		if blocker, err = engine.BlockerFromDynamics(ctx, cp.Blockers, cS.fltrS, tnt, evNm); err != nil {
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
	AlteredFields   []*attributes.FieldsAltered
	CGREvent        *utils.CGREvent
}

// so we can compare from outside if event changed with AttributeS
var ChargerSDefaultAlteredFields = []string{utils.MetaOptsRunID,
	utils.MetaOpts + utils.NestingSep + utils.MetaChargeID,
	utils.MetaOpts + utils.NestingSep + utils.MetaSubsys}

func (cS *ChargerS) processEvent(ctx *context.Context, tnt string, cgrEv *utils.CGREvent) (rply []*ChrgSProcessEventReply, err error) {
	cPs, err := cS.matchingChargerProfilesForEvent(ctx, tnt, cgrEv)
	if err != nil {
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
			AlteredFields: []*attributes.FieldsAltered{
				{
					MatchedProfileID: utils.MetaDefault,
					Fields:           slices.Clone(ChargerSDefaultAlteredFields),
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
		var evReply attributes.AttrSProcessEventReply
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
