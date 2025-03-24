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

	"github.com/cgrates/birpc/context"

	"github.com/cgrates/cgrates/utils"
)

// SortedRoute represents one route in SortedRoutes
type SortedRoute struct {
	RouteID            string
	RouteParameters    string
	SortingData        map[string]any            // store here extra info like cost or stats (can contain the data that we do not use to sort after)
	sortingDataDecimal map[string]*utils.Decimal // only the data we sort after
}

// getSortedData returns the data from map or a 0( in case it was not populated)
func (s *SortedRoute) getSortedData(field string) *utils.Decimal {
	if v, has := s.sortingDataDecimal[field]; has {
		return v
	}
	return utils.NewDecimalFromFloat64(0)
}

// SortedRoutes represents all viable routes inside one routing profile
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

func (sRoutes *SortedRoutes) compareWeight(i, j int) bool {
	cmp := sRoutes.Routes[i].getSortedData(utils.Weight).Compare(sRoutes.Routes[j].getSortedData(utils.Weight))
	if cmp == 0 {
		return utils.BoolGenerator().RandomBool()
	}
	return cmp > 0
}

// SortWeight is part of sort interface, sort based on Weight
func (sRoutes *SortedRoutes) SortWeight() {
	sort.Slice(sRoutes.Routes, sRoutes.compareWeight)
}

// SortLeastCost is part of sort interface,
// sort ascendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortLeastCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(utils.Cost).Compare(sRoutes.Routes[j].getSortedData(utils.Cost))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
	})
}

// SortHighestCost is part of sort interface,
// sort descendent based on Cost with fallback on Weight
func (sRoutes *SortedRoutes) SortHighestCost() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(utils.Cost).Compare(sRoutes.Routes[j].getSortedData(utils.Cost))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp > 0
	})
}

// SortQOS is part of sort interface,
// sort based on Stats
func (sRoutes *SortedRoutes) SortQOS(params []string) {
	//sort routes
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		for _, param := range params {
			//in case we have the same value for the current param we skip to the next one
			cmp := sRoutes.Routes[i].getSortedData(param).Compare(sRoutes.Routes[j].getSortedData(param))
			if cmp == 0 {
				continue
			}
			if param == utils.MetaPDD {
				return cmp < 0
			}
			return cmp > 0
		}
		//in case that we have the same value for all params we sort base on weight
		return sRoutes.compareWeight(i, j)
	})
}

// SortResourceAscendent is part of sort interface,
// sort ascendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceAscendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(utils.ResourceUsage).Compare(sRoutes.Routes[j].getSortedData(utils.ResourceUsage))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
	})
}

// SortResourceDescendent is part of sort interface,
// sort descendent based on ResourceUsage with fallback on Weight
func (sRoutes *SortedRoutes) SortResourceDescendent() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		cmp := sRoutes.Routes[i].getSortedData(utils.ResourceUsage).Compare(sRoutes.Routes[j].getSortedData(utils.ResourceUsage))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp > 0
	})
}

// SortLoadDistribution is part of sort interface,
// sort based on the following formula float64(metricVal/ratio) with fallback on Weight
func (sRoutes *SortedRoutes) SortLoadDistribution() {
	sort.Slice(sRoutes.Routes, func(i, j int) bool {
		// ((ratio + metricVal) / (ratio)) -1 = ((ratio+metricVal)/ratio) - (ratio/ratio) = (ratio+metricVal-ratio)/ratio = metricVal/ratio
		cmp := utils.DivideDecimal(sRoutes.Routes[i].getSortedData(utils.Load), sRoutes.Routes[i].getSortedData(utils.Ratio)).Compare(
			utils.DivideDecimal(sRoutes.Routes[j].getSortedData(utils.Load), sRoutes.Routes[j].getSortedData(utils.Ratio)))
		if cmp == 0 {
			return sRoutes.compareWeight(i, j)
		}
		return cmp < 0
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
	SortRoutes(*context.Context, string, map[string]*RouteWithWeight, *utils.CGREvent, *optsGetRoutes) (*SortedRoutes, error)
}

// RouteSortDispatcher will initialize strategies
// and dispatch requests to them
type RouteSortDispatcher map[string]RoutesSorter

func (ssd RouteSortDispatcher) SortRoutes(ctx *context.Context, prflID, strategy string,
	suppls map[string]*RouteWithWeight, suplEv *utils.CGREvent, extraOpts *optsGetRoutes) (_ *SortedRoutes, err error) {
	sd, has := ssd[strategy]
	if !has {
		err = fmt.Errorf("unsupported sorting strategy: %s", strategy)
		return
	}
	return sd.SortRoutes(ctx, prflID, suppls, suplEv, extraOpts)
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
	routeIDs := make(utils.StringSet)
	for _, sR := range sRs {
		for _, spl := range sR.Routes {
			route := spl.RouteID
			if spl.RouteParameters != utils.EmptyString {
				route += utils.InInFieldSep + spl.RouteParameters
			}
			if !routeIDs.Has(route) {
				routeIDs.Add(route)
				sPs = append(sPs, route)
			}
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

// routeLazyPass filters the route based on
func routeLazyPass(ctx *context.Context, filters []*FilterRule, ev *utils.CGREvent, data utils.MapStorage,
	resConns, statConns, acntConns, trdConns, rnkConns []string) (pass bool, err error) {
	if len(filters) == 0 {
		return true, nil
	}

	dynDP := newDynamicDP(ctx, resConns, statConns, acntConns, //construct the DP and pass it to filterS
		trdConns, rnkConns, ev.Tenant, utils.MapStorage{
			utils.MetaReq:  ev.Event,
			utils.MetaOpts: ev.APIOpts,
			utils.MetaVars: data,
		})

	for _, rule := range filters { // verify the rules remaining from PartialPass
		if pass, err = rule.Pass(ctx, dynDP); err != nil || !pass {
			return
		}
	}
	return true, nil
}

// RouteWithWeight attaches static weight to Route
type RouteWithWeight struct {
	*Route
	Weight         float64
	blocker        bool
	lazyCheckRules []*FilterRule
}
