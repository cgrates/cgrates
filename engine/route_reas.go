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

func NewResourceAscendetSorter(rS *RouteService) *ResourceAscendentSorter {
	return &ResourceAscendentSorter{rS: rS,
		sorting: utils.MetaReas}
}

// ResourceAscendentSorter orders ascendent routes based on their Resource Usage
type ResourceAscendentSorter struct {
	sorting string
	rS      *RouteService
}

func (ws *ResourceAscendentSorter) SortRoutes(prflID string,
	routes []*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting:      ws.sorting,
		SortedRoutes: make([]*SortedRoute, 0)}
	for _, route := range routes {
		if len(route.ResourceIDs) == 0 {
			utils.Logger.Warning(
				fmt.Sprintf("<%s> supplier: <%s> - empty ResourceIDs",
					utils.RouteS, route.ID))
			return nil, utils.NewErrMandatoryIeMissing("ResourceIDs")
		}
		if srtSpl, pass, err := ws.rS.populateSortingData(suplEv, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.SortedRoutes = append(sortedRoutes.SortedRoutes, srtSpl)
		}
	}
	sortedRoutes.SortResourceAscendent()
	return
}
