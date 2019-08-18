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
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortLeastCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersSortWeight(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortWeight()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight: 10.5,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight: 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestSortedSuppliersDigest(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 10.0,
				},
				SupplierParameters: "param1",
			},
		},
	}
	exp := "supplier2:param2,supplier1:param1"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest2(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID: "SPL_WEIGHT_1",
		Sorting:   utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					"Weight": 30.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					"Weight": 20.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	exp := "supplier1:param1,supplier2:param2"
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestSortedSuppliersDigest3(t *testing.T) {
	eSpls := SortedSuppliers{
		ProfileID:       "SPL_WEIGHT_1",
		Sorting:         utils.MetaWeight,
		SortedSuppliers: []*SortedSupplier{},
	}
	exp := ""
	rcv := eSpls.Digest()
	if !reflect.DeepEqual(rcv, exp) {
		t.Errorf("Expecting: %+v, received: %+v", exp, rcv)
	}
}

func TestLibSuppliersSortHighestCost(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortHighestCost()
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:   0.2,
					utils.Weight: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:   0.1,
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Cost:   0.05,
					utils.Weight: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

//sort based on *acd and *tcd
func TestLibSuppliersSortQOS(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Cost:    0.5,
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			&SortedSupplier{
				//the average value for supplier2 for *acd is 0.5 , *tcd 4.1
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Cost:    0.1,
					utils.Weight:  15.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 4.1,
				},
			},
			&SortedSupplier{
				//the average value for supplier3 for *acd is 0.4 , *tcd 5.1
				SupplierID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *acd and *tcd
func TestLibSuppliersSortQOS2(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the average value for supplier1 for *acd is 0.5 , *tcd  1.1
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			&SortedSupplier{
				//the worst value for supplier1 for *acd is 0.5 , *tcd  1.1
				//supplier1 and supplier2 have the same value for *acd and *tcd
				//will be sorted based on weight
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  17.0,
					utils.MetaACD: 0.5,
					utils.MetaTCD: 1.1,
				},
			},
			&SortedSupplier{

				SupplierID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

//sort based on *pdd
func TestLibSuppliersSortQOS3(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				//the worst value for supplier1 for *pdd is 0.7 , *tcd  1.1
				//supplier1 and supplier3 have the same value for *pdd
				//will be sorted based on weight
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 1.1,
				},
			},
			&SortedSupplier{
				//the worst value for supplier2 for *pdd is 1.2, *tcd  1.1
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 1.2,
					utils.MetaTCD: 1.1,
				},
			},
			&SortedSupplier{
				//the worst value for supplier3 for *pdd is 0.7, *tcd  10.1
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  10.0,
					utils.MetaPDD: 0.7,
					utils.MetaTCD: 10.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaPDD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS4(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier1", "supplier3", "supplier2"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS5(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
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
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS6(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: 0.2,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: 0.2,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 0.1,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS7(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: -1.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier3", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortQOS8(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:  15.0,
					utils.MetaACD: -1.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:  25.0,
					utils.MetaACD: -1.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:  20.0,
					utils.MetaACD: 10.0,
				},
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier3", "supplier2", "supplier1"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID

	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}

func TestLibSuppliersSortLoadDistribution(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				SortingData: map[string]interface{}{
					utils.Weight:    25.0,
					utils.Ratio:     4.0,
					utils.LoadValue: 3.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				SortingData: map[string]interface{}{
					utils.Weight:    15.0,
					utils.Ratio:     10.0,
					utils.LoadValue: 5.0,
				},
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				SortingData: map[string]interface{}{
					utils.Weight:    25.0,
					utils.Ratio:     1.0,
					utils.LoadValue: 1.0,
				},
			},
		},
	}
	sSpls.SortLoadDistribution()
	rcv := make([]string, len(sSpls.SortedSuppliers))
	eIds := []string{"supplier2", "supplier1", "supplier3"}
	for i, spl := range sSpls.SortedSuppliers {
		rcv[i] = spl.SupplierID
	}
	if !reflect.DeepEqual(eIds, rcv) {
		t.Errorf("Expecting: %+v, \n received: %+v",
			eIds, rcv)
	}
}
