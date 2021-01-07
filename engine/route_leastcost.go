/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT MetaAny WARRANTY; without even the implied warranty of
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

func NewLeastCostSorter(rS *RouteService) *LeastCostSorter {
	return &LeastCostSorter{rS: rS,
		sorting: utils.MetaLC}
}

// LeastCostSorter sorts routes based on their cost
type LeastCostSorter struct {
	sorting string
	rS      *RouteService
}

func (lcs *LeastCostSorter) SortRoutes(prflID string, routes []*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting:      lcs.sorting,
		SortedRoutes: make([]*SortedRoute, 0)}
	for _, s := range routes {
		if len(s.RatingPlanIDs) == 0 && len(s.AccountIDs) == 0 && len(lcs.rS.cgrcfg.RouteSCfg().RateSConns) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty RatingPlanIDs or AccountIDs or no connection with RateS",
					utils.RouteS, s.ID))
			return nil, utils.NewErrMandatoryIeMissing("RatingPlanIDs or AccountIDs or connection with RateS")
		}
		if srtSpl, pass, err := lcs.rS.populateSortingData(ev, s, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.SortedRoutes = append(sortedRoutes.SortedRoutes, srtSpl)
		}
	}
	sortedRoutes.SortLeastCost()
	return
}
