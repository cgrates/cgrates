// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	sortedRoutes.SortResourceAscendent()
	return
}
