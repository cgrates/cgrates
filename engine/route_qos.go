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
	"github.com/cgrates/birpc/context"
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

func (qos *QOSRouteSorter) SortRoutes(ctx *context.Context, prflID string, routes map[string]*Route,
	ev *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{ProfileID: prflID,
		Sorting: qos.sorting,
		Routes:  make([]*SortedRoute, 0)}
	for _, route := range routes {
		if srtSpl, pass, err := qos.rS.populateSortingData(ctx, ev, route, extraOpts); err != nil {
			return nil, err
		} else if pass && srtSpl != nil {
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortQOS(extraOpts.sortingParameters)
	return
}
