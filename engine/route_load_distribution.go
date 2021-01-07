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

// NewLoadDistributionSorter .
func NewLoadDistributionSorter(rS *RouteService) *LoadDistributionSorter {
	return &LoadDistributionSorter{rS: rS,
		sorting: utils.MetaLoad}
}

// LoadDistributionSorter orders suppliers based on their Resource Usage
type LoadDistributionSorter struct {
	sorting string
	rS      *RouteService
}

// SortRoutes .
func (ws *LoadDistributionSorter) SortRoutes(prflID string,
	routes []*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting:      ws.sorting,
		SortedRoutes: make([]*SortedRoute, 0)}
	for _, route := range routes {
		// we should have at least 1 statID defined for counting CDR (a.k.a *sum:1)
		if len(route.StatIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty StatIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("StatIDs")
		}
		if srtSpl, pass, err := ws.rS.populateSortingData(suplEv, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			// Add the ratio in SortingData so we can used it later in SortLoadDistribution
			floatRatio, err := utils.IfaceAsFloat64(route.cacheRoute[utils.MetaRatio])
			if err != nil {
				utils.Logger.Warning(
					fmt.Sprintf("<%s> cannot convert ratio <%+v> to float64 supplier: <%s>",
						utils.RouteS, route.cacheRoute[utils.MetaRatio], route.ID))
			}
			srtSpl.SortingData[utils.Ratio] = floatRatio
			sortedRoutes.SortedRoutes = append(sortedRoutes.SortedRoutes, srtSpl)
		}
	}
	sortedRoutes.SortLoadDistribution()
	return
}
