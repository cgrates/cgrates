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
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/utils"
)

func TestLibSuppliersSortCost(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier3",
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
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier1",
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
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				RouteParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				RouteParameters: "param3",
			},
			{
				RouteID: "supplier1",
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
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				RouteParameters: "param1",
			},
		},
	}
	exp := "supplier2:param2,supplier1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest2(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				RouteParameters: "param2",
			},
		},
	}
	exp := "supplier1:param1,supplier2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedRoutesDigest3(t *testing.T) {
	eSpls := SortedRoutes{
		ProfileID:    "SPL_WEIGHT_1",
		Sorting:      utils.MetaWeight,
		SortedRoutes: []*SortedRoute{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestLibRoutesSortHighestCost(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier3",
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
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "supplier3",
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

//sort based on *acd and *tcd
func TestLibRoutesSortQOS(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the average value for supplier2 for *acd is 0.5 , *tcd 4.1
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:    0.1,
					utils.Weight:  15.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 4.1,
				},
			},
			{
				//the average value for supplier3 for *acd is 0.4 , *tcd 5.1
				RouteID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *acd and *tcd
func TestLibRoutesSortQOS2(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				//supplier1 and supplier2 have the same value for *acd and *tcd
				//will be sorted based on weight
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  17.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			{

				RouteID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *pdd
func TestLibRoutesSortQOS3(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				//the worst value for supplier1 for *pdd is 0.7 , *tcd  1.1
				//supplier1 and supplier3 have the same value for *pdd
				//will be sorted based on weight
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier2 for *pdd is 1.2, *tcd  1.1
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 1.2,
					utils.MetaTCD: 1.1,
				},
			},
			{
				//the worst value for supplier3 for *pdd is 0.7, *tcd  10.1
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaPDD})
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS4(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS5(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
			},
			{
				RouteID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS6(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: 0.2,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 0.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS7(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: -1.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortQOS8(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 10.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibRoutesSortLoadDistribution(t *testing.T) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.Ratio:  4.0,
					utils.Load:   3.0,
				},
			},
			{
				RouteID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
					utils.Ratio:  10.0,
					utils.Load:   5.0,
				},
			},
			{
				RouteID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
					utils.Ratio:  1.0,
					utils.Load:   1.0,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	rcv := make([]string, len(sSpls.SortedRoutes))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedRoutes {
		rcv[i] = spl.RouteID
	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func BenchmarkRouteSortCost(b *testing.B) {
	sSpls := &SortedRoutes{
		SortedRoutes: []*SortedRoute{
			{
				RouteID: "route1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param1",
			},
			{
				RouteID: "route2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				RouteParameters: "param2",
			},
			{
				RouteID: "route3",
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
