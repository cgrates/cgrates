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

func populateCostForRoutes(ctx *context.Context, cfg *config.CGRConfig,
	connMgr *ConnManager, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes []*SortedRoute, err error) {
	if len(cfg.RouteSCfg().RateSConns) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("connIDs")
	}
	sortedRoutes = make([]*SortedRoute, 0, len(routes))
	for _, route := range routes {
		if len(route.RateProfileIDs) == 0 && len(route.AccountIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RateProfileIDs or AccountIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("RateProfileIDs or AccountIDs")
		}
		srtRoute := &SortedRoute{
			RouteID: route.ID,
			SortingData: map[string]interface{}{
				utils.Weight: route.Weight,
			},
			sortingDataF64: map[string]float64{
				utils.Weight: route.Weight,
			},
			RouteParameters: route.RouteParameters,
		}
		var rpCost utils.RateProfileCost
		if err = connMgr.Call(ctx, cfg.RouteSCfg().RateSConns,
			utils.RateSv1CostForEvent,
			&utils.ArgsCostForEvent{
				RateProfileIDs: route.RateProfileIDs,
				CGREvent:       ev,
			}, &rpCost); err != nil {
			if extraOpts.ignoreErrors {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> ignoring route with ID: %s, err: %s",
						utils.RouteS, route.ID, err.Error()))
				continue
			}
			return
		}
		cost, _ := rpCost.Cost.Float64()
		srtRoute.sortingDataF64[utils.Cost] = cost
		srtRoute.SortingData[utils.Cost] = cost
		srtRoute.SortingData[utils.RatingPlanID] = rpCost.ID
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			cfg.FilterSCfg().ResourceSConns,
			cfg.FilterSCfg().StatSConns,
			cfg.FilterSCfg().AccountSConns); err != nil {
			return
		} else if pass {
			sortedRoutes = append(sortedRoutes, srtRoute)
		}
	}
	return
}

func NewHighestCostSorter(cfg *config.CGRConfig, connMgr *ConnManager) *HightCostSorter {
	return &HightCostSorter{cfg: cfg, connMgr: connMgr}
}

// HightCostSorter sorts routes based on their cost
type HightCostSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

func (hcs *HightCostSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateCostForRoutes(ctx, hcs.cfg, hcs.connMgr, routes, ev, extraOpts); err != nil {
		return
	}

	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaHC,
		Routes:    sRoutes,
	}
	sortedRoutes.SortHighestCost()
	return
}

func NewLeastCostSorter(cfg *config.CGRConfig, connMgr *ConnManager) *LeastCostSorter {
	return &LeastCostSorter{cfg: cfg, connMgr: connMgr}
}

// LeastCostSorter sorts routes based on their cost
type LeastCostSorter struct {
	cfg     *config.CGRConfig
	connMgr *ConnManager
}

func (lcs *LeastCostSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateCostForRoutes(ctx, lcs.cfg, lcs.connMgr, routes, ev, extraOpts); err != nil {
		return
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaLC,
		Routes:    sRoutes,
	}
	sortedRoutes.SortLeastCost()
	return
}
