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

func (hcs *HightCostSorter) SortRoutes(prflID string, routes []*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting:      hcs.sorting,
		SortedRoutes: make([]*SortedRoute, 0)}
	for _, route := range routes {
		if len(route.RatingPlanIDs) == 0 && len(route.AccountIDs) == 0 && len(route.RateProfileIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RatingPlanIDs or AccountIDs or RateProfileIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("RatingPlanIDs or AccountIDs or RateProfileIDs")
		}
		if srtSpl, pass, err := hcs.rS.populateSortingData(ev, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.SortedRoutes = append(sortedRoutes.SortedRoutes, srtSpl)
		}
	}
	sortedRoutes.SortHighestCost()
	return
}
