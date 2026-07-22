// Copyright ITsysCOM GmbH
// SPDX-License-Identifier: AGPL-3.0-or-later

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
	routes map[string]*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sortedRoutes = &SortedRoutes{
		ProfileID: prflID,
		Sorting:   ws.sorting,
		Routes:    make([]*SortedRoute, 0),
	}
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
			srtSpl.sortingDataF64[utils.Ratio] = floatRatio
			sortedRoutes.Routes = append(sortedRoutes.Routes, srtSpl)
		}
	}
	sortedRoutes.SortLoadDistribution()
	return
}
