// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

package engine

import (
	"github.com/cgrates/cgrates/utils"
)

func NewQOSRouteSorter(rS *RouteService) *QOSRouteSorter {
	return &QOSRouteSorter{rS: rS,
		sorting: utils.MetaQOS}
}

// QOSSorter sorts route based on stats
type QOSRouteSorter struct {
	sorting string
	rS      *RouteService
}

func (qos *QOSRouteSorter) SortRoutes(prflID string, routes map[string]*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: qos.sorting,
		Routes:  make([]*SortedRoute, 0)}
	for _, route := range routes {
		if srtSpl, pass, err := qos.rS.populateSortingData(ev, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortQOS(extraOpts.sortingParameters)
	return
}
