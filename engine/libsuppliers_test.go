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

func TestLibSuppliersSortQOS(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.05,
					utils.MetaTCD: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.05,
					utils.MetaTCD: 10.0,
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

func TestLibSuppliersSortQOS2(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD, utils.MetaTCD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
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

func TestLibSuppliersSortQOS3(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: 1.2,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: -1.0,
				},
				SupplierParameters: "param2",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersSortQOS4(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaTCC, utils.MetaASR, utils.MetaACD, utils.MetaTCD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 20.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaTCD: 10.0,
					utils.MetaASR: 1.2,
					utils.MetaTCC: 10.1,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaTCD: 15.0,
					utils.MetaASR: -1.0,
					utils.MetaTCC: 10.1,
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

func TestLibSuppliersSortQOS5(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaPDD: 0.5,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaPDD: 0.6,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaPDD: 0.2,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaPDD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
					utils.MetaPDD: 0.2,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaPDD: 0.5,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
					utils.MetaPDD: 0.6,
				},
				SupplierParameters: "param2",
			},
		},
	}
	if !reflect.DeepEqual(eOrderedSpls, sSpls) {
		t.Errorf("Expecting: %s, received: %s",
			utils.ToJSON(eOrderedSpls), utils.ToJSON(sSpls))
	}
}

func TestLibSuppliersSortQOS6(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: 0.2,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},

			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 0.1,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
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

func TestLibSuppliersSortQOS7(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
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

func TestLibSuppliersSortQOS8(t *testing.T) {
	sSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
				},
				SupplierParameters: "param1",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param3",
			},
		},
	}
	sSpls.SortQOS([]string{utils.MetaACD})
	eOrderedSpls := &SortedSuppliers{
		SortedSuppliers: []*SortedSupplier{
			&SortedSupplier{
				SupplierID: "supplier3",
				globalStats: map[string]float64{
					utils.MetaACD: 10.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 20.0,
				},
				SupplierParameters: "param3",
			},
			&SortedSupplier{
				SupplierID: "supplier2",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 25.0,
				},
				SupplierParameters: "param2",
			},

			&SortedSupplier{
				SupplierID: "supplier1",
				globalStats: map[string]float64{
					utils.MetaACD: -1.0,
				},
				SortingData: map[string]interface{}{
					utils.Weight: 15.0,
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
