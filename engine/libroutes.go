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
	"sort"
	"strings"

	"github.com/cgrates/cgrates/utils"
)

// SortedRoute represents one route in SortedRoutes
type SortedRoute struct {
	RouteID         string
	RouteParameters string
	SortingData     map[string]interface{} // store here extra info like cost or stats
}

// SortedRoutes is returned as part of GetRoutes call
type SortedRoutes struct {
	ProfileID string         // Profile matched
	Sorting   string         // Sorting algorithm
	Count     int            // number of routes returned
	Routes    []*SortedRoute // list of route IDs and SortingData data
}

// RouteIDs returns a list of route IDs
func (sRoutes *SortedRoutes) RouteIDs() (rIDs []string) {
	rIDs = make([]string, len(sRoutes.Routes))
	for i, sRoute := range sRoutes.Routes {
		rIDs[i] = sRoute.RouteID
	}
	return
}

// RoutesWithParams returns a list of routes IDs with Parameters
func (sRoutes *SortedRoutes) RoutesWithParams() (sPs []string) {
	sPs = make([]string, len(sRoutes.Routes))
	for i, spl := range sRoutes.Routes {
		sPs[i] = spl.RouteID
		if spl.RouteParameters != "" {
			sPs[i] += utils.InInFieldSep + spl.RouteParameters
		}
	}
	return
}

// SortWeight is part of sort interface, sort based on Weight
func (sRoutes *SortedRoutes) SortWeight() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
			return utils.BoolGenerator().RandomBool()
		}
		return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
	})
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortLeastCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].SortingData[utils.Cost].(float64) == sRoutes.Routes[j].SortingData[utils.Cost].(float64) {
			if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
		}
		return sRoutes.Routes[i].SortingData[utils.Cost].(float64) < sRoutes.Routes[j].SortingData[utils.Cost].(float64)
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortHighestCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].SortingData[utils.Cost].(float64) == sRoutes.Routes[j].SortingData[utils.Cost].(float64) {
			if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
		}
		return sRoutes.Routes[i].SortingData[utils.Cost].(float64) > sRoutes.Routes[j].SortingData[utils.Cost].(float64)
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sRoutes *SortedRoutes) SortQOS(params []string) {
	//sort routes
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			if sRoutes.Routes[i].SortingData[param].(float64) == sRoutes.Routes[j].SortingData[param].(float64) {
				continue
			}
			switch param {
			default:
				if sRoutes.Routes[i].SortingData[param].(float64) > sRoutes.Routes[j].SortingData[param].(float64) {
					return true
				}
				return false
			case utils.MetaPDD: //in case of pdd the smallest value if the best
				if sRoutes.Routes[i].SortingData[param].(float64) < sRoutes.Routes[j].SortingData[param].(float64) {
					return true
				}
				return false
			}

		}
		//in case that we have the same value for all params we sort base on weight
		if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
			return utils.BoolGenerator().RandomBool()
		}
		return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceAscendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].SortingData[utils.ResourceUsage].(float64) == sRoutes.Routes[j].SortingData[utils.ResourceUsage].(float64) {
			if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
		}
		return sRoutes.Routes[i].SortingData[utils.ResourceUsage].(float64) < sRoutes.Routes[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceDescendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].SortingData[utils.ResourceUsage].(float64) == sRoutes.Routes[j].SortingData[utils.ResourceUsage].(float64) {
			if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
		}
		return sRoutes.Routes[i].SortingData[utils.ResourceUsage].(float64) > sRoutes.Routes[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// SortLoadDistribution is part of sort interface,
// sort based on the following formula (float64(ratio + metricVal) / float64(ratio)) -1 with fallback on Weight
func (sRoutes *SortedRoutes) SortLoadDistribution() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		splIVal := ((sRoutes.Routes[i].SortingData[utils.Ratio].(float64)+sRoutes.Routes[i].SortingData[utils.Load].(float64))/sRoutes.Routes[i].SortingData[utils.Ratio].(float64) - 1.0)
		splJVal := ((sRoutes.Routes[j].SortingData[utils.Ratio].(float64)+sRoutes.Routes[j].SortingData[utils.Load].(float64))/sRoutes.Routes[j].SortingData[utils.Ratio].(float64) - 1.0)
		if splIVal == splJVal {
			if sRoutes.Routes[i].SortingData[utils.Weight].(float64) == sRoutes.Routes[j].SortingData[utils.Weight].(float64) {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].SortingData[utils.Weight].(float64) > sRoutes.Routes[j].SortingData[utils.Weight].(float64)
		}
		return splIVal < splJVal
	})
}

// Digest returns list of routeIDs + parameters for easier outside access
// format route1:route1params,route2:route2params
func (sRoutes *SortedRoutes) Digest() string {
	return strings.Join(sRoutes.RoutesWithParams(), utils.FieldsSep)
}

func (ss *SortedRoute) AsNavigableMap() (nm utils.NavigableMap2) {
	nm = utils.NavigableMap2{
		utils.RouteID:         utils.NewNMData(ss.RouteID),
		utils.RouteParameters: utils.NewNMData(ss.RouteParameters),
	}
	sd := utils.NavigableMap2{}
	for k, d := range ss.SortingData {
		sd[k] = utils.NewNMData(d)
	}
	nm[utils.SortingData] = sd
	return
}

func (sRoutes *SortedRoutes) AsNavigableMap() (nm utils.NavigableMap2) {
	nm = utils.NavigableMap2{
		utils.ProfileID: utils.NewNMData(sRoutes.ProfileID),
		utils.Sorting:   utils.NewNMData(sRoutes.Sorting),
		utils.Count:     utils.NewNMData(sRoutes.Count),
	}
	sr := make(utils.NMSlice, len(sRoutes.Routes))
	for i, ss := range sRoutes.Routes {
		sr[i] = ss.AsNavigableMap()
	}
	nm[utils.SortedRoutes] = &sr
	return
}

// RoutesSorter is the interface which needs to be implemented by routes sorters
type RoutesSorter interface {
	SortRoutes(string, map[string]*Route, *utils.CGREvent, *optsGetRoutes) (*SortedRoutes, error)
}

// NewRouteSortDispatcher constructs RouteSortDispatcher
func NewRouteSortDispatcher(lcrS *RouteService) (rsd RouteSortDispatcher) {
	rsd = make(map[string]RoutesSorter)
	rsd[utils.MetaWeight] = NewWeightSorter(lcrS)
	rsd[utils.MetaLC] = NewLeastCostSorter(lcrS)
	rsd[utils.MetaHC] = NewHighestCostSorter(lcrS)
	rsd[utils.MetaQOS] = NewQOSRouteSorter(lcrS)
	rsd[utils.MetaReas] = NewResourceAscendetSorter(lcrS)
	rsd[utils.MetaReds] = NewResourceDescendentSorter(lcrS)
	rsd[utils.MetaLoad] = NewLoadDistributionSorter(lcrS)
	return
}

// RouteSortDispatcher will initialize strategies
// and dispatch requests to them
type RouteSortDispatcher map[string]RoutesSorter

func (ssd RouteSortDispatcher) SortRoutes(prflID, strategy string,
	suppls map[string]*Route, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (sortedRoutes *SortedRoutes, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	if sortedRoutes, err = sd.SortRoutes(prflID, suppls, suplEv, extraOpts); err != nil {
		return
	}
	if len(sortedRoutes.Routes) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}

// SortedRoutesList represents the list of matched routes grouped based of profile
type SortedRoutesList []*SortedRoutes

// RouteIDs returns a list of route IDs
func (sRs SortedRoutesList) RouteIDs() (rIDs []string) {
	for _, sR := range sRs {
		for _, r := range sR.Routes {
			rIDs = append(rIDs, r.RouteID)
		}
	}
	return
}

// RoutesWithParams returns a list of routes IDs with Parameters
func (sRs SortedRoutesList) RoutesWithParams() (sPs []string) {
	for _, sR := range sRs {
		for _, spl := range sR.Routes {
			route := spl.RouteID
			if spl.RouteParameters != "" {
				route += utils.InInFieldSep + spl.RouteParameters
			}
			sPs = append(sPs, route)
		}
	}
	return
}

// Digest returns list of routeIDs + parameters for easier outside access
// format route1:route1params,route2:route2params
func (sRs SortedRoutesList) Digest() string {
	return strings.Join(sRs.RoutesWithParams(), utils.FieldsSep)
}

// AsNavigableMap returns the SortedRoutesSet as NMInterface object
func (sRs SortedRoutesList) AsNavigableMap() (nm utils.NMSlice) {
	nm = make(utils.NMSlice, len(sRs))
	for i, ss := range sRs {
		nm[i] = ss.AsNavigableMap()
	}
	return
}
