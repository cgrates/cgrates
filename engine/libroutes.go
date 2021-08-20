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
	SortingData     map[string]interface{} // store here extra info like cost or stats (can contain the data that we do not use to sort after)
	sortingDataF64  map[string]float64     // only the data we sort after
}

// SortedRoutes is returned as part of GetRoutes call
type SortedRoutes struct {
	ProfileID string         // Profile matched
	Sorting   string         // Sorting algorithm
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
		if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
			return utils.BoolGenerator().RandomBool()
		}
		return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
	})
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortLeastCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].sortingDataF64[utils.Cost] == sRoutes.Routes[j].sortingDataF64[utils.Cost] {
			if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
		}
		return sRoutes.Routes[i].sortingDataF64[utils.Cost] < sRoutes.Routes[j].sortingDataF64[utils.Cost]
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortHighestCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].sortingDataF64[utils.Cost] == sRoutes.Routes[j].sortingDataF64[utils.Cost] {
			if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
		}
		return sRoutes.Routes[i].sortingDataF64[utils.Cost] > sRoutes.Routes[j].sortingDataF64[utils.Cost]
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sRoutes *SortedRoutes) SortQOS(params []string) {
	//sort routes
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		for _, param := range params {
			param1 := sRoutes.Routes[i].sortingDataF64[param]
			param2 := sRoutes.Routes[j].sortingDataF64[param]

			//in case we have the same value for the current param we skip to the next one
			if param1 == param2 {
				continue
			}
			switch param {
			default:
				if param1 > param2 {
					return true
				}
				return false
			case utils.MetaPDD: //in case of pdd the smallest value if the best
				if param1 < param2 {
					return true
				}
				return false
			}
		}
		//in case that we have the same value for all params we sort base on weight
		if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
			return utils.BoolGenerator().RandomBool()
		}
		return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceAscendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].sortingDataF64[utils.ResourceUsage] == sRoutes.Routes[j].sortingDataF64[utils.ResourceUsage] {
			if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
		}
		return sRoutes.Routes[i].sortingDataF64[utils.ResourceUsage] < sRoutes.Routes[j].sortingDataF64[utils.ResourceUsage]
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceDescendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		if sRoutes.Routes[i].sortingDataF64[utils.ResourceUsage] == sRoutes.Routes[j].sortingDataF64[utils.ResourceUsage] {
			if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
		}
		return sRoutes.Routes[i].sortingDataF64[utils.ResourceUsage] > sRoutes.Routes[j].sortingDataF64[utils.ResourceUsage]
	})
}

// SortLoadDistribution is part of sort interface,
// sort based on the following formula float64(metricVal/ratio) with fallback on Weight
func (sRoutes *SortedRoutes) SortLoadDistribution() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		// ((ratio + metricVal) / (ratio)) -1 = ((ratio+metricVal)/ratio) - (ratio/ratio) = (ratio+metricVal-ratio)/ratio = metricVal/ratio
		splIVal := sRoutes.Routes[i].sortingDataF64[utils.Load] / sRoutes.Routes[i].sortingDataF64[utils.Ratio]
		splJVal := sRoutes.Routes[j].sortingDataF64[utils.Load] / sRoutes.Routes[j].sortingDataF64[utils.Ratio]
		if splIVal == splJVal {
			if sRoutes.Routes[i].sortingDataF64[utils.Weight] == sRoutes.Routes[j].sortingDataF64[utils.Weight] {
				return utils.BoolGenerator().RandomBool()
			}
			return sRoutes.Routes[i].sortingDataF64[utils.Weight] > sRoutes.Routes[j].sortingDataF64[utils.Weight]
		}
		return splIVal < splJVal
	})
}

// Digest returns list of routeIDs + parameters for easier outside access
// format route1:route1params,route2:route2params
func (sRoutes *SortedRoutes) Digest() string {
	return strings.Join(sRoutes.RoutesWithParams(), utils.FieldsSep)
}

func (ss *SortedRoute) AsNavigableMap() (nm *utils.DataNode) {
	nm = &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.RouteID:         utils.NewLeafNode(ss.RouteID),
			utils.RouteParameters: utils.NewLeafNode(ss.RouteParameters),
		},
	}
	sd := &utils.DataNode{Type: utils.NMMapType, Map: map[string]*utils.DataNode{}}
	for k, d := range ss.SortingData {
		sd.Map[k] = utils.NewLeafNode(d)
	}
	nm.Map[utils.SortingData] = sd
	return
}

func (sRoutes *SortedRoutes) AsNavigableMap() (nm *utils.DataNode) {
	nm = &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.ProfileID: utils.NewLeafNode(sRoutes.ProfileID),
			utils.Sorting:   utils.NewLeafNode(sRoutes.Sorting),
		},
	}
	sr := make([]*utils.DataNode, len(sRoutes.Routes))
	for i, ss := range sRoutes.Routes {
		sr[i] = ss.AsNavigableMap()
	}
	nm.Map[utils.CapRoutes] = &utils.DataNode{Type: utils.NMSliceType, Slice: sr}
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
	return sd.SortRoutes(prflID, suppls, suplEv, extraOpts)
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
func (sRs SortedRoutesList) AsNavigableMap() (nm *utils.DataNode) {
	nm = &utils.DataNode{Type: utils.NMSliceType, Slice: make([]*utils.DataNode, len(sRs))}
	for i, ss := range sRs {
		nm.Slice[i] = ss.AsNavigableMap()
	}
	return
}
