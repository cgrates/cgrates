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

package routes

import (
	"fmt"

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func populateResourcesForRoutes(ctx *context.Context, cfg *config.CGRConfig,
	connMgr *engine.ConnManager, routes map[string]*RouteWithWeight,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes []*SortedRoute, err error) {
	if len(cfg.RouteSCfg().ResourceSConns) == 0 {
		return nil, utils.NewErrMandatoryIeMissing("connIDs")
	}
	sortedRoutes = make([]*SortedRoute, 0, len(routes))
	for _, route := range routes {
		if len(route.ResourceIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty ResourceIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("ResourceIDs")
		}
		srtRoute := &SortedRoute{
			RouteID: route.ID,
			SortingData: map[string]any{
				utils.Weight: route.Weight,
			},
			sortingDataDecimal: map[string]*utils.Decimal{
				utils.Weight: utils.NewDecimalFromFloat64(route.Weight),
			},
			RouteParameters: route.RouteParameters,
		}
		if route.blocker {
			srtRoute.SortingData[utils.Blocker] = true
		}
		var tUsage float64
		for _, resID := range route.ResourceIDs {
			var res engine.Resource
			if err = connMgr.Call(ctx, cfg.RouteSCfg().ResourceSConns, utils.ResourceSv1GetResource,
				&utils.TenantIDWithAPIOpts{TenantID: &utils.TenantID{Tenant: ev.Tenant, ID: resID}},
				&res); err != nil && err.Error() != utils.ErrNotFound.Error() {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> error: %s getting resource for ID : %s", utils.RouteS, err.Error(), resID))
				err = nil
				continue
			}
			tUsage += res.TotalUsage()
		}
		srtRoute.SortingData[utils.ResourceUsage] = tUsage
		srtRoute.sortingDataDecimal[utils.ResourceUsage] = utils.NewDecimalFromFloat64(tUsage)
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			cfg.FilterSCfg().ResourceSConns,
			cfg.FilterSCfg().StatSConns,
			cfg.FilterSCfg().AccountSConns,
			cfg.FilterSCfg().TrendSConns,
			cfg.FilterSCfg().RankingSConns); err != nil {
			return
		} else if pass {
			sortedRoutes = append(sortedRoutes, srtRoute)
		}
	}
	return
}

func NewResourceAscendetSorter(cfg *config.CGRConfig, connMgr *engine.ConnManager) *ResourceAscendentSorter {
	return &ResourceAscendentSorter{cfg: cfg, connMgr: connMgr}
}

// ResourceAscendentSorter orders ascendent routes based on their Resource Usage
type ResourceAscendentSorter struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
}

func (ws *ResourceAscendentSorter) SortRoutes(ctx *context.Context, prflID string,
	routes map[string]*RouteWithWeight, ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateResourcesForRoutes(ctx, ws.cfg, ws.connMgr, routes, ev, extraOpts); err != nil {
		return
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaReas,
		Routes:    sRoutes,
	}
	sortedRoutes.SortResourceAscendent()
	return
}

func NewResourceDescendentSorter(cfg *config.CGRConfig, connMgr *engine.ConnManager) *ResourceDescendentSorter {
	return &ResourceDescendentSorter{cfg: cfg, connMgr: connMgr}
}

// ResourceDescendentSorter orders suppliers based on their Resource Usage
type ResourceDescendentSorter struct {
	cfg     *config.CGRConfig
	connMgr *engine.ConnManager
}

func (ws *ResourceDescendentSorter) SortRoutes(ctx *context.Context, prflID string,
	routes map[string]*RouteWithWeight, ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	var sRoutes []*SortedRoute
	if sRoutes, err = populateResourcesForRoutes(ctx, ws.cfg, ws.connMgr, routes, ev, extraOpts); err != nil {
		return
	}
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaReds,
		Routes:    sRoutes,
	}
	sortedRoutes.SortResourceDescendent()
	return
}
