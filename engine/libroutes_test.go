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
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortLeastCost()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibRoutesSortWeight(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.5,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.5,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestSortedRoutesDigest(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	exp := "route2:param2,route1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest2(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 30.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}
	exp := "route1:param1,route2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest3(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		Routes:    []*SortedRoute{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestLibRoutesSortHighestCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortHighestCost()
	eOrderedSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

// sort based on *acd and *tcd
func TestLibRoutesSortQOS(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the average value for route1 for *acd is 0.5 , *tcd  1.1
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the average value for route2 for *acd is 0.5 , *tcd 4.1
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:    0.1,
					utils.Weight:  15.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 4.1,
				},
				SortingData: map[string]interface{}{
					utils.Cost:    0.1,
					utils.Weight:  15.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 4.1,
				},
			},
			{
				//the average value for route3 for *acd is 0.4 , *tcd 5.1
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:    1.1,
					utils.Weight:  17.8,
					utils.MetaACD: 0.4,
					utils.MetaTCD: 5.1,
				},
				SortingData: map[string]interface{}{
					utils.Cost:    1.1,
					utils.Weight:  17.8,
					utils.MetaACD: 0.4,
					utils.MetaTCD: 5.1,
				},
			},
		},
	}

	//sort base on *acd and *tcd
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

// sort based on *acd and *tcd
func TestLibRoutesSortQOS2(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the average value for route1 for *acd is 0.5 , *tcd  1.1
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route1 for *acd is 0.5 , *tcd  1.1
				//route1 and route2 have the same value for *acd and *tcd
				//will be sorted based on weight
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight:  17.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  17.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.7,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.7,
					utils.MetaTCD: 1.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route3", "route2", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

// sort based on *pdd
func TestLibRoutesSortQOS3(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				//the worst value for route1 for *pdd is 0.7 , *tcd  1.1
				//route1 and route3 have the same value for *pdd
				//will be sorted based on weight
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight:  15.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route2 for *pdd is 1.2, *tcd  1.1
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight:  10.0,
					utils.MetaPDD: 1.2,
					utils.MetaTCD: 1.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 1.2,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for route3 for *pdd is 0.7, *tcd  10.1
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight:  10.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 10.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaPDD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route1", "route3", "route2"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS4(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route1", "route3", "route2"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS5(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaTCC, utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route3", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS6(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight:  15.0,
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight:  25.0,
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight:  20.0,
					utils.MetaACD: 0.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 0.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS7(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight:  20.0,
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: -1.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route3", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS8(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight:  20.0,
					utils.MetaACD: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 10.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route3", "route2", "route1"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortLoadDistribution(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Weight: 25.0,
					utils.Ratio:  4.0,
					utils.Load:   3.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.Ratio:  4.0,
					utils.Load:   3.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Weight: 15.0,
					utils.Ratio:  10.0,
					utils.Load:   5.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.Ratio:  10.0,
					utils.Load:   5.0,
				},
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Weight: 25.0,
					utils.Ratio:  1.0,
					utils.Load:   1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.Ratio:  1.0,
					utils.Load:   1.0,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	rcv := make([]string, len(sSpls.Routes))
	eIds := []string{"route2", "route1", "route3"}
	for i, spl := range sSpls.Routes {
		rcv[i] = spl.RouteID
	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesLCSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{RouteID: strconv.Itoa(i), SortingData: map[string]interface{}{
			utils.Cost:   0.1,
			utils.Weight: 10.0,
		}}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortLeastCost()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func TestLibRoutesHCSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{RouteID: strconv.Itoa(i), SortingData: map[string]interface{}{
			utils.Cost:   0.1,
			utils.Weight: 10.0,
		}}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortHighestCost()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func TestLibRoutesResAscSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{
			RouteID: strconv.Itoa(i),
			sortingDataF64: map[string]float64{
				utils.ResourceUsage: 5.0,
				utils.Weight:        10.0,
			},
			SortingData: map[string]interface{}{
				utils.ResourceUsage: 5.0,
				utils.Weight:        10.0,
			},
		}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortResourceAscendent()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func TestLibRoutesResDescSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{
			RouteID: strconv.Itoa(i),
			sortingDataF64: map[string]float64{
				utils.ResourceUsage: 5.0,
				utils.Weight:        10.0,
			},
			SortingData: map[string]interface{}{
				utils.ResourceUsage: 5.0,
				utils.Weight:        10.0,
			},
		}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortResourceDescendent()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}

}

func TestLibRoutesLoadDistSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{
			RouteID: strconv.Itoa(i),
			sortingDataF64: map[string]float64{
				utils.Ratio:  4.0,
				utils.Load:   3.0,
				utils.Weight: 10.0,
			},
			SortingData: map[string]interface{}{
				utils.Ratio:  4.0,
				utils.Load:   3.0,
				utils.Weight: 10.0,
			},
		}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortLoadDistribution()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func TestLibRoutesQOSSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{RouteID: strconv.Itoa(i), SortingData: map[string]interface{}{
			utils.Weight:  10.0,
			utils.MetaACD: -1.0,
		}}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortQOS([]string{utils.MetaACD})
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func TestLibRoutesSameWeight(t *testing.T) {
	sSpls := &SortedRoutes{}
	sortedSlice := &SortedRoutes{}
	for i := 0; i <= 10; i++ {
		route := &SortedRoute{RouteID: strconv.Itoa(i), SortingData: map[string]interface{}{
			utils.Weight: 10.0,
		}}
		sSpls.Routes = append(sSpls.Routes, route)
		sortedSlice.Routes = append(sortedSlice.Routes, route)
	}
	for i := 0; i < 3; i++ {
		sSpls.SortWeight()
		// we expect to receive this in a random order
		// the comparison logic is the following if the slice is the same as sorted slice we return error
		if reflect.DeepEqual(sortedSlice, sSpls) {
			t.Errorf("Expecting: %+v, \n received: %+v",
				sortedSlice, sSpls)
		}
	}
}

func BenchmarkRouteSortCost(b *testing.B) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		sSpls.SortLeastCost()
	}
}

func TestRouteIDsGetIDs(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
		},
	}
	expected := []string{"route1", "route2", "route3"}
	sort.Strings(expected)
	rcv := sSpls.RouteIDs()
	sort.Strings(rcv)
	if !reflect.DeepEqual(expected, rcv) {
		t.Errorf("Expected %+v, received %+v", expected, rcv)
	}
}

func TestSortHighestCost(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 11.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 11.0,
				},
			},
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
			},
		},
	}
	sSpls.SortHighestCost()
	ex := sSpls
	if !reflect.DeepEqual(ex, sSpls) {
		t.Errorf("Expected %+v, received %+v", ex, sSpls)
	}
}

func TestSortResourceAscendentDescendent(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "route2",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 10.0,
					utils.Weight:        10.0,
				},
				SortingData: map[string]interface{}{
					utils.ResourceUsage: 10.0,
					utils.Weight:        10.0,
				},
			},
			{
				RouteID: "route1",
				sortingDataF64: map[string]float64{
					utils.ResourceUsage: 10.0,
					utils.Weight:        11.0,
				},
				SortingData: map[string]interface{}{
					utils.ResourceUsage: 10.0,
					utils.Weight:        11.0,
				},
			},
		},
	}

	//SortingResourceAscendent/Descendent while ResourceUsages are equal
	expSRts := sSpls
	sSpls.SortResourceAscendent()
	if !reflect.DeepEqual(expSRts, sSpls) {
		t.Errorf("Expected %+v, received %+v", expSRts, sSpls)
	}

	sSpls.SortResourceDescendent()
	if !reflect.DeepEqual(expSRts, sSpls) {
		t.Errorf("Expected %+v, received %+v", expSRts, sSpls)
	}

	//SortingResourceAscendent/Descendent while ResourceUsages are not equal
	sSpls.Routes[0].SortingData[utils.ResourceUsage] = 11.0
	sSpls.Routes[0].sortingDataF64[utils.ResourceUsage] = 11.0
	expSRts = sSpls
	sSpls.SortResourceAscendent()
	if !reflect.DeepEqual(expSRts, sSpls) {
		t.Errorf("Expected %+v, received %+v", expSRts, sSpls)
	}

	sSpls.SortResourceDescendent()
	if !reflect.DeepEqual(expSRts, sSpls) {
		t.Errorf("Expected %+v, received %+v", expSRts, sSpls)
	}
}

func TestSortLoadDistribution(t *testing.T) {
	sSpls := &SortedRoutes{
		Routes: []*SortedRoute{
			{
				RouteID: "ROUTE1",
				sortingDataF64: map[string]float64{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 15.5,
				},
				SortingData: map[string]interface{}{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 15.5,
				},
			},
			{
				RouteID: "ROUTE2",
				sortingDataF64: map[string]float64{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 14.5,
				},
				SortingData: map[string]interface{}{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 14.5,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	expSSPls := sSpls
	if !reflect.DeepEqual(sSpls, expSSPls) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expSSPls), utils.ToJSON(sSpls))
	}
}

func TestSortedRouteAsNavigableMap(t *testing.T) {
	sSpls := &SortedRoute{
		RouteID:         "ROUTE1",
		RouteParameters: "SORTING_PARAMETER",
		sortingDataF64: map[string]float64{
			utils.Ratio:  6.0,
			utils.Load:   10.0,
			utils.Weight: 15.5,
		},
		SortingData: map[string]interface{}{
			utils.Ratio:  6.0,
			utils.Load:   10.0,
			utils.Weight: 15.5,
		},
	}
	expNavMap := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.RouteID:         utils.NewLeafNode("ROUTE1"),
			utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER"),
			utils.SortingData: {
				Type: utils.NMMapType,
				Map: map[string]*utils.DataNode{
					utils.Ratio:  utils.NewLeafNode(6.0),
					utils.Load:   utils.NewLeafNode(10.0),
					utils.Weight: utils.NewLeafNode(15.5),
				},
			},
		},
	}
	if rcv := sSpls.AsNavigableMap(); !reflect.DeepEqual(rcv, expNavMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expNavMap), utils.ToJSON(rcv))
	}
}

func TestSortedRoutesAsNavigableMap(t *testing.T) {
	sSpls := &SortedRoutes{
		ProfileID: "TEST_ID1",
		Sorting:   utils.MetaWeight,
		Routes: []*SortedRoute{
			{
				RouteID:         "ROUTE1",
				RouteParameters: "SORTING_PARAMETER",
				sortingDataF64: map[string]float64{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 15.5,
				},
				SortingData: map[string]interface{}{
					utils.Ratio:  6.0,
					utils.Load:   10.0,
					utils.Weight: 15.5,
				},
			},
			{
				RouteID:         "ROUTE2",
				RouteParameters: "SORTING_PARAMETER_SECOND",
				sortingDataF64: map[string]float64{
					utils.Ratio: 7.0,
					utils.Load:  10.0,
				},
				SortingData: map[string]interface{}{
					utils.Ratio: 7.0,
					utils.Load:  10.0,
				},
			},
		},
	}

	expNavMap := &utils.DataNode{
		Type: utils.NMMapType,
		Map: map[string]*utils.DataNode{
			utils.ProfileID: utils.NewLeafNode("TEST_ID1"),
			utils.Sorting:   utils.NewLeafNode(utils.MetaWeight),
			utils.CapRoutes: {
				Type: utils.NMSliceType,
				Slice: []*utils.DataNode{
					{
						Type: utils.NMMapType,
						Map: map[string]*utils.DataNode{
							utils.RouteID:         utils.NewLeafNode("ROUTE1"),
							utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER"),
							utils.SortingData: {
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.Ratio:  utils.NewLeafNode(6.0),
									utils.Load:   utils.NewLeafNode(10.0),
									utils.Weight: utils.NewLeafNode(15.5),
								},
							},
						},
					},
					{
						Type: utils.NMMapType,
						Map: map[string]*utils.DataNode{
							utils.RouteID:         utils.NewLeafNode("ROUTE2"),
							utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER_SECOND"),
							utils.SortingData: {
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.Ratio: utils.NewLeafNode(7.0),
									utils.Load:  utils.NewLeafNode(10.0),
								},
							},
						},
					},
				},
			},
		},
	}

	if rcv := sSpls.AsNavigableMap(); !reflect.DeepEqual(rcv, expNavMap) {
		t.Errorf("Expected %+v, received %+v", utils.ToJSON(expNavMap), utils.ToJSON(rcv))
	}
}

func TestSortedRoutesListRouteIDs(t *testing.T) {
	sr := SortedRoutesList{
		{
			Routes: []*SortedRoute{
				{
					RouteID: "sr1id1",
				},
				{
					RouteID: "sr1id2",
				},
			},
		},
		{
			Routes: []*SortedRoute{
				{
					RouteID: "sr2id1",
				},
				{
					RouteID: "sr2id2",
				},
			},
		},
	}

	val := sr.RouteIDs()
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	exp := []string{"sr1id1", "sr1id2", "sr2id1", "sr2id2"}
	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", exp, val)
	}
}

func TestSortedRoutesListRoutesWithParams(t *testing.T) {
	sRs := SortedRoutesList{
		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route1",
					RouteParameters: "params1",
				},
				{
					RouteID:         "route2",
					RouteParameters: "params2",
				},
			},
		},
		{
			Routes: []*SortedRoute{
				{
					RouteID:         "route3",
					RouteParameters: "params3",
				},
				{
					RouteID:         "route4",
					RouteParameters: "params4",
				},
			},
		},
	}
	val := sRs.RoutesWithParams()
	sort.Slice(val, func(i, j int) bool {
		return val[i] < val[j]
	})
	exp := []string{"route1:params1", "route2:params2", "route3:params3", "route4:params4"}

	if !reflect.DeepEqual(val, exp) {
		t.Errorf("expected %v ,received %v", val, exp)
	}

}

func TestSortedRoutesListAsNavigableMap(t *testing.T) {
	sRs := SortedRoutesList{
		&SortedRoutes{
			ProfileID: "TEST_ID1",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID:         "ROUTE1",
					RouteParameters: "SORTING_PARAMETER",
					sortingDataF64: map[string]float64{
						utils.Ratio:  6.0,
						utils.Load:   10.0,
						utils.Weight: 15.5,
					},
					SortingData: map[string]interface{}{
						utils.Ratio:  6.0,
						utils.Load:   10.0,
						utils.Weight: 15.5,
					},
				},
				{
					RouteID:         "ROUTE2",
					RouteParameters: "SORTING_PARAMETER_SECOND",
					sortingDataF64: map[string]float64{
						utils.Ratio: 7.0,
						utils.Load:  10.0,
					},
					SortingData: map[string]interface{}{
						utils.Ratio: 7.0,
						utils.Load:  10.0,
					},
				},
			},
		},
		&SortedRoutes{
			ProfileID: "TEST_ID2",
			Sorting:   utils.MetaWeight,
			Routes: []*SortedRoute{
				{
					RouteID:         "ROUTE1",
					RouteParameters: "SORTING_PARAMETER",
					sortingDataF64: map[string]float64{
						utils.Ratio:  6.0,
						utils.Load:   10.0,
						utils.Weight: 15.5,
					},
					SortingData: map[string]interface{}{
						utils.Ratio:  6.0,
						utils.Load:   10.0,
						utils.Weight: 15.5,
					},
				},
				{
					RouteID:         "ROUTE2",
					RouteParameters: "SORTING_PARAMETER_SECOND",
					sortingDataF64: map[string]float64{
						utils.Ratio: 7.0,
						utils.Load:  10.0,
					},
					SortingData: map[string]interface{}{
						utils.Ratio: 7.0,
						utils.Load:  10.0,
					},
				},
			},
		},
	}
	expNavMap := &utils.DataNode{
		Type: utils.NMSliceType,
		Slice: []*utils.DataNode{
			{
				Type: utils.NMMapType,
				Map: map[string]*utils.DataNode{
					utils.ProfileID: utils.NewLeafNode("TEST_ID1"),
					utils.Sorting:   utils.NewLeafNode(utils.MetaWeight),
					utils.CapRoutes: {
						Type: utils.NMSliceType,
						Slice: []*utils.DataNode{
							{
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.RouteID:         utils.NewLeafNode("ROUTE1"),
									utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER"),
									utils.SortingData: {
										Type: utils.NMMapType,
										Map: map[string]*utils.DataNode{
											utils.Ratio:  utils.NewLeafNode(6.0),
											utils.Load:   utils.NewLeafNode(10.0),
											utils.Weight: utils.NewLeafNode(15.5),
										},
									},
								},
							},
							{
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.RouteID:         utils.NewLeafNode("ROUTE2"),
									utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER_SECOND"),
									utils.SortingData: {
										Type: utils.NMMapType,
										Map: map[string]*utils.DataNode{
											utils.Ratio: utils.NewLeafNode(7.0),
											utils.Load:  utils.NewLeafNode(10.0),
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Type: utils.NMMapType,
				Map: map[string]*utils.DataNode{
					utils.ProfileID: utils.NewLeafNode("TEST_ID2"),
					utils.Sorting:   utils.NewLeafNode(utils.MetaWeight),
					utils.CapRoutes: {
						Type: utils.NMSliceType,
						Slice: []*utils.DataNode{
							{
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.RouteID:         utils.NewLeafNode("ROUTE1"),
									utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER"),
									utils.SortingData: {
										Type: utils.NMMapType,
										Map: map[string]*utils.DataNode{
											utils.Ratio:  utils.NewLeafNode(6.0),
											utils.Load:   utils.NewLeafNode(10.0),
											utils.Weight: utils.NewLeafNode(15.5),
										},
									},
								},
							},
							{
								Type: utils.NMMapType,
								Map: map[string]*utils.DataNode{
									utils.RouteID:         utils.NewLeafNode("ROUTE2"),
									utils.RouteParameters: utils.NewLeafNode("SORTING_PARAMETER_SECOND"),
									utils.SortingData: {
										Type: utils.NMMapType,
										Map: map[string]*utils.DataNode{
											utils.Ratio: utils.NewLeafNode(7.0),
											utils.Load:  utils.NewLeafNode(10.0),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if val := sRs.AsNavigableMap(); !reflect.DeepEqual(val, expNavMap) {
		t.Errorf("expected %v ,received %v", utils.ToJSON(expNavMap), utils.ToJSON(val))
	}

}

func TestSortedRoutesListDigest(t *testing.T) {

	sRs := &SortedRoutesList{
		{
			ProfileID: "TEST_ID1",
			Routes: []*SortedRoute{
				{
					RouteID:         "ROUTE_ID1",
					RouteParameters: "PARAM_1",
				},
				{
					RouteID:         "ROUTE_ID2",
					RouteParameters: "PARAM_2",
				},
			},
		},
		{
			ProfileID: "TEST_ID2",
			Routes: []*SortedRoute{
				{
					RouteID:         "ROUTE_ID1",
					RouteParameters: "PARAM_1",
				},
				{
					RouteID:         "ROUTE_ID2",
					RouteParameters: "PARAM_2",
				},
			},
		},
	}

	exp := "ROUTE_ID1:PARAM_1,ROUTE_ID2:PARAM_2,ROUTE_ID1:PARAM_1,ROUTE_ID2:PARAM_2"

	if val := sRs.Digest(); val != exp {
		t.Errorf("received %v", val)
	}
}

func TestRouteSortDispatcher(t *testing.T) {
	ssd := RouteSortDispatcher{}
	strategy := "strategy"
	if _, err := ssd.SortRoutes("prfID", strategy, map[string]*Route{}, &utils.CGREvent{}, &optsGetRoutes{}); err == nil || err.Error() != fmt.Sprintf("unsupported sorting strategy: %s", strategy) {
		t.Error(err)
	}
}
