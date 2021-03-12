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

func NewResourceDescendentSorter(rS *RouteService) *ResourceDescendentSorter {
	return &ResourceDescendentSorter{rS: rS,
		sorting: utils.MetaReds}
}

// ResourceDescendentSorter orders suppliers based on their Resource Usage
type ResourceDescendentSorter struct {
	sorting string
	rS      *RouteService
}

func (ws *ResourceDescendentSorter) SortRoutes(prflID string,
	routes map[string]*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: ws.sorting,
		Routes:  make([]*SortedRoute, 0)}
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
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortResourceDescendent()
	return
}
