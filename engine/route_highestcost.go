// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func NewHighestCostSorter(rS *RouteService) *HightCostSorter {
	return &HightCostSorter{rS: rS,
		sorting: utils.MetaHC}
}

// HightCostSorter sorts routes based on their cost
type HightCostSorter struct {
	sorting string
	rS      *RouteService
}

func (hcs *HightCostSorter) SortRoutes(prflID string, routes map[string]*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: hcs.sorting,
		Routes:  make([]*SortedRoute, 0)}
	for _, route := range routes {
		if len(route.RatingPlanIDs) == 0 && len(route.AccountIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RatingPlanIDs or AccountIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("RatingPlanIDs or AccountIDs")
		}
		if srtSpl, pass, err := hcs.rS.populateSortingData(ev, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortHighestCost()
	return
}
