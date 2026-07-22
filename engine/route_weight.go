// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

func NewWeightSorter(rS *RouteService) *WeightSorter {
	return &WeightSorter{rS: rS,
		sorting: utils.MetaWeight}
}

// WeightSorter orders routes based on their weight, no cost involved
type WeightSorter struct {
	sorting string
	rS      *RouteService
}

func (ws *WeightSorter) SortRoutes(prflID string,
	routes map[string]*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: ws.sorting,
		Routes:  make([]*SortedRoute, 0)}
	for _, route := range routes {
		if srtRoute, pass, err := ws.rS.populateSortingData(suplEv, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtRoute != nil {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtRoute)
		}
	}
	sortedRoutes.SortWeight()
	return
}
