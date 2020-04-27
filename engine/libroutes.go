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

	"github.com/cgrates/cgrates/config"
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
	ProfileID    string         // Profile matched
	Sorting      string         // Sorting algorithm
	Count        int            // number of suppliers returned
	SortedRoutes []*SortedRoute // list of supplier IDs and SortingData data
}

// RouteIDs returns a list of route IDs
func (sRoutes *SortedRoutes) RouteIDs() (rIDs []string) {
	rIDs = make([]string, len(sRoutes.SortedRoutes))
	for i, sRoute := range sRoutes.SortedRoutes {
		rIDs[i] = sRoute.RouteID
	}
	return
}

// RoutesWithParams returns a list of routes IDs with Parameters
func (sRoutes *SortedRoutes) RoutesWithParams() (sPs []string) {
	sPs = make([]string, len(sRoutes.SortedRoutes))
	for i, spl := range sRoutes.SortedRoutes {
		sPs[i] = spl.RouteID
		if spl.RouteParameters != "" {
			sPs[i] += utils.InInFieldSep + spl.RouteParameters
		}
	}
	return
}

// SortWeight is part of sort interface, sort based on Weight
func (sRoutes *SortedRoutes) SortWeight() {
	sort.Slice(sRoutes.SortedRoutes, func(i, j int) bool {
		return sRoutes.SortedRoutes[i].SortingData[utils.Weight].(float64) > sRoutes.SortedRoutes[j].SortingData[utils.Weight].(float64)
	})
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sSpls *SortedRoutes) SortLeastCost() {
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		if sSpls.SortedRoutes[i].SortingData[utils.Cost].(float64) == sSpls.SortedRoutes[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedRoutes[i].SortingData[utils.Cost].(float64) < sSpls.SortedRoutes[j].SortingData[utils.Cost].(float64)
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sSpls *SortedRoutes) SortHighestCost() {
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		if sSpls.SortedRoutes[i].SortingData[utils.Cost].(float64) == sSpls.SortedRoutes[j].SortingData[utils.Cost].(float64) {
			return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedRoutes[i].SortingData[utils.Cost].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Cost].(float64)
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sSpls *SortedRoutes) SortQOS(params []string) {
	//sort suppliers
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			if sSpls.SortedRoutes[i].SortingData[param].(float64) == sSpls.SortedRoutes[j].SortingData[param].(float64) {
				continue
			}
			switch param {
			default:
				if sSpls.SortedRoutes[i].SortingData[param].(float64) > sSpls.SortedRoutes[j].SortingData[param].(float64) {
					return true
				}
				return false
			case utils.MetaPDD: //in case of pdd the smallest value if the best
				if sSpls.SortedRoutes[i].SortingData[param].(float64) < sSpls.SortedRoutes[j].SortingData[param].(float64) {
					return true
				}
				return false
			}

		}
		//in case that we have the same value for all params we sort base on weight
		return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sSpls *SortedRoutes) SortResourceAscendent() {
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		if sSpls.SortedRoutes[i].SortingData[utils.ResourceUsage].(float64) == sSpls.SortedRoutes[j].SortingData[utils.ResourceUsage].(float64) {
			return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedRoutes[i].SortingData[utils.ResourceUsage].(float64) < sSpls.SortedRoutes[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sSpls *SortedRoutes) SortResourceDescendent() {
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		if sSpls.SortedRoutes[i].SortingData[utils.ResourceUsage].(float64) == sSpls.SortedRoutes[j].SortingData[utils.ResourceUsage].(float64) {
			return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
		}
		return sSpls.SortedRoutes[i].SortingData[utils.ResourceUsage].(float64) > sSpls.SortedRoutes[j].SortingData[utils.ResourceUsage].(float64)
	})
}

// SortLoadDistribution is part of sort interface,
// sort based on the following formula (float64(ratio + metricVal) / float64(ratio)) -1 with fallback on Weight
func (sSpls *SortedRoutes) SortLoadDistribution() {
	sort.Slice(sSpls.SortedRoutes, func(i, j int) bool {
		splIVal := ((sSpls.SortedRoutes[i].SortingData[utils.Ratio].(float64)+sSpls.SortedRoutes[i].SortingData[utils.Load].(float64))/sSpls.SortedRoutes[i].SortingData[utils.Ratio].(float64) - 1.0)
		splJVal := ((sSpls.SortedRoutes[j].SortingData[utils.Ratio].(float64)+sSpls.SortedRoutes[j].SortingData[utils.Load].(float64))/sSpls.SortedRoutes[j].SortingData[utils.Ratio].(float64) - 1.0)
		if splIVal == splJVal {
			return sSpls.SortedRoutes[i].SortingData[utils.Weight].(float64) > sSpls.SortedRoutes[j].SortingData[utils.Weight].(float64)
		}
		return splIVal < splJVal
	})
}

// Digest returns list of supplierIDs + parameters for easier outside access
// format suppl1:suppl1params,suppl2:suppl2params
func (sSpls *SortedRoutes) Digest() string {
	return strings.Join(sSpls.RoutesWithParams(), utils.FIELDS_SEP)
}

func (sSpls *SortedRoutes) AsNavigableMap() (nm *config.NavigableMap) {
	mp := map[string]interface{}{
		"ProfileID": sSpls.ProfileID,
		"Sorting":   sSpls.Sorting,
		"Count":     sSpls.Count,
	}
	sm := make([]map[string]interface{}, len(sSpls.SortedRoutes))
	for i, ss := range sSpls.SortedRoutes {
		sm[i] = map[string]interface{}{
			"SupplierID":         ss.RouteID,
			"SupplierParameters": ss.RouteParameters,
			"SortingData":        ss.SortingData,
		}
	}
	mp["SortedSuppliers"] = sm
	return config.NewNavigableMap(mp)
}

type SupplierWithParams struct {
	SupplierName   string
	SupplierParams string
}

// SuppliersSorter is the interface which needs to be implemented by supplier sorters
type RoutesSorter interface {
	SortRoutes(string, []*Route, *utils.CGREvent, *optsGetSuppliers) (*SortedRoutes, error)
}

// NewRouteSortDispatcher constructs RouteSortDispatcher
func NewRouteSortDispatcher(lcrS *RouteService) (rsd RouteSortDispatcher, err error) {
	rsd = make(map[string]RoutesSorter)
	rsd[utils.MetaWeight] = NewWeightSorter(lcrS)
	rsd[utils.MetaLC] = NewLeastCostSorter(lcrS)
	rsd[utils.MetaHC] = NewHighestCostSorter(lcrS)
	rsd[utils.MetaQOS] = NewQOSSupplierSorter(lcrS)
	rsd[utils.MetaReas] = NewResourceAscendetSorter(lcrS)
	rsd[utils.MetaReds] = NewResourceDescendentSorter(lcrS)
	rsd[utils.MetaLoad] = NewLoadDistributionSorter(lcrS)
	return
}

// RouteSortDispatcher will initialize strategies
// and dispatch requests to them
type RouteSortDispatcher map[string]RoutesSorter

func (ssd RouteSortDispatcher) SortSuppliers(prflID, strategy string,
	suppls []*Route, suplEv *utils.CGREvent, extraOpts *optsGetSuppliers) (sortedRoutes *SortedRoutes, err error) {
	sd, has := ssd[strategy]
	if !has {
		return nil, fmt.Errorf("unsupported sorting strategy: %s", strategy)
	}
	if sortedRoutes, err = sd.SortRoutes(prflID, suppls, suplEv, extraOpts); err != nil {
		return
	}
	if len(sortedRoutes.SortedRoutes) == 0 {
		return nil, utils.ErrNotFound
	}
	return
}
