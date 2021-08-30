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

func NewWeightSorter(cfg *config.CGRConfig) *WeightSorter {
	return &WeightSorter{cfg: cfg}
}

// WeightSorter orders routes based on their weight, no cost involved
type WeightSorter struct {
	cfg *config.CGRConfig
}

func (ws *WeightSorter) SortRoutes(ctx *context.Context, prflID string,
	routes map[string]*RouteWithWeight, ev *utils.CGREvent, _ *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   utils.MetaWeight,
		Routes:    make([]*SortedRoute, 0, len(routes)),
	}
	for _, route := range routes {
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
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			ws.cfg.FilterSCfg().ResourceSConns,
			ws.cfg.FilterSCfg().StatSConns,
			ws.cfg.FilterSCfg().AccountSConns); err != nil {
			return
		} else if pass {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtRoute)
		}
	}
	sortedRoutes.SortWeight()
	return
}
