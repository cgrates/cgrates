// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package routes

import (
	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/config"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

func NewWeightSorter(cfg *config.CGRConfig, fS *engine.FilterS) *WeightSorter {
	return &WeightSorter{cfg: cfg, fS: fS}
}

// WeightSorter orders routes based on their weight, no cost involved
type WeightSorter struct {
	cfg *config.CGRConfig
	fS  *engine.FilterS
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
		var pass bool
		if pass, err = routeLazyPass(ctx, route.lazyCheckRules, ev, srtRoute.SortingData,
			ws.cfg, ws.fS); err != nil {
			return
		} else if pass {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtRoute)
		}
	}
	sortedRoutes.SortWeight()
	return
}
