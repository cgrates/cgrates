// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"fmt"

	"github.com/cgrates/cgrates/utils"
)

func NewLeastCostSorter(rS *RouteService) *LeastCostSorter {
	return &LeastCostSorter{rS: rS,
		sorting: utils.MetaLC}
}

// LeastCostSorter sorts routes based on their cost
type LeastCostSorter struct {
	sorting string
	rS      *RouteService
}

func (lcs *LeastCostSorter) SortRoutes(prflID string, routes map[string]*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: lcs.sorting,
		Routes:  make([]*SortedRoute, 0)}
	for _, s := range routes {
		if len(s.RatingPlanIDs) == 0 && len(s.AccountIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RatingPlanIDs or AccountIDs",
					utils.RouteS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("RatingPlanIDs or AccountIDs")
		}
		if srtSpl, pass, err := lcs.rS.populateSortingData(ev, s, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortLeastCost()
	return
}
